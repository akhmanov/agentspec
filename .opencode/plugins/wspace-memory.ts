import { type Plugin, type PluginInput, tool } from "@opencode-ai/plugin"

type OpenCodePluginContext = {
  project: PluginInput["project"]
  directory: string
  worktree: string
  client: PluginInput["client"]
}

type ToolExecuteAfterInput = {
  tool: string
  sessionID: string
  callID: string
  args: Record<string, unknown>
}

type ToolExecuteAfterOutput = {
  title: string
  output: string
  metadata: Record<string, unknown>
}

type SessionCreatedEvent = {
  event: {
    sessionID: string
  }
}

type SessionCompactedEvent = {
  event: {
    sessionID: string
    summary?: string
  }
}

type FileEditedEvent = {
  event: {
    sessionID: string
    path: string
  }
}

type OpenCodeEvent = {
  type: string
  sessionID?: string
  summary?: string
  path?: string
  [key: string]: unknown
}

declare const Bun: {
  spawn(command: string[], options: {
    cwd: string
    stdin?: string
    stdout: "pipe"
    stderr: "pipe"
  }): {
    stdout: ReadableStream<Uint8Array>
    stderr: ReadableStream<Uint8Array>
    exited: Promise<number>
  }
}

const MAX_TEXT = 2000

function pluginRoot(ctx: OpenCodePluginContext): string {
  return ctx.worktree || ctx.directory
}

function memoryScriptPath(ctx: OpenCodePluginContext): string {
  return `${pluginRoot(ctx)}/scripts/memory.py`
}

async function log(ctx: OpenCodePluginContext, level: "debug" | "info" | "warn" | "error", message: string, extra: Record<string, unknown> = {}) {
  if (!ctx.client?.app?.log) return
  await ctx.client.app.log({
    body: {
      service: "wspace-memory",
      level,
      message,
      extra,
    },
  })
}

async function runMemory(ctx: OpenCodePluginContext, args: string[], payload?: Record<string, unknown>): Promise<{ ok: boolean; stdout: string; stderr: string }> {
  const proc = Bun.spawn(["python3", memoryScriptPath(ctx), ...args], {
    cwd: pluginRoot(ctx),
    stdin: payload ? JSON.stringify(payload) : undefined,
    stdout: "pipe",
    stderr: "pipe",
  })

  const [stdout, stderr, exitCode] = await Promise.all([
    new Response(proc.stdout).text(),
    new Response(proc.stderr).text(),
    proc.exited,
  ])

  return {
    ok: exitCode === 0,
    stdout: stdout.trim(),
    stderr: stderr.trim(),
  }
}

async function ensureMemoryReady(ctx: OpenCodePluginContext): Promise<boolean> {
  const result = await runMemory(ctx, ["status"])
  return result.ok && result.stdout.includes("state=initialized")
}

function basePayload(ctx: OpenCodePluginContext, sessionID: string) {
  return {
    session_id: sessionID,
    project_id: ctx.project?.id || "wspace",
    repo_id: "wspace",
    cwd: ctx.directory,
    worktree_path: pluginRoot(ctx),
  }
}

function truncate(value: string | undefined): string {
  const text = value || ""
  return text.length > MAX_TEXT ? text.slice(0, MAX_TEXT) : text
}

function safeToolPayload(input: ToolExecuteAfterInput, output: ToolExecuteAfterOutput) {
  return {
    tool: input.tool,
    arg_keys: Object.keys(input.args || {}).sort(),
    title: output.title,
    output_length: truncate(output.output).length,
    metadata_keys: Object.keys(output.metadata || {}).sort(),
  }
}

export const WspaceMemoryPlugin: Plugin = async (ctx) => {
  return {
    "tool.execute.after": async (input: ToolExecuteAfterInput, output: ToolExecuteAfterOutput) => {
      if (!(await ensureMemoryReady(ctx))) return
      const result = await runMemory(ctx, ["capture"], {
        kind: "observation",
        observation_id: `tool:${input.sessionID}:${input.callID}`,
        ...basePayload(ctx, input.sessionID),
        event_kind: `tool.${input.tool}`,
        created_at: new Date().toISOString(),
        payload: safeToolPayload(input, output),
      })
      if (!result.ok) {
        await log(ctx, "warn", "tool observation capture failed", { stderr: result.stderr, tool: input.tool })
      }
    },

    "experimental.session.compacting": async (_input: Record<string, unknown>, output: { context: string[] }) => {
      if (!(await ensureMemoryReady(ctx))) return
      const result = await runMemory(ctx, ["wakeup", "--project", ctx.project?.id || "wspace", "--limit", "2"])
      if (!result.ok || !result.stdout || result.stdout === "No memory context available.") return
      output.context.push(result.stdout)
    },

    event: async ({ event }: { event: OpenCodeEvent }) => {
      if (!(await ensureMemoryReady(ctx))) return

      switch (event.type) {
        case "session.created": {
          const result = await runMemory(ctx, ["capture"], {
            kind: "session_start",
            ...basePayload(ctx, String(event.sessionID || "")),
            started_at: new Date().toISOString(),
            last_active_at: new Date().toISOString(),
          })
          if (!result.ok) {
            await log(ctx, "warn", "session capture failed", { stderr: result.stderr })
          }
          break
        }

        case "session.compacted": {
          if (!event.summary) break
          const result = await runMemory(ctx, ["capture"], {
            kind: "summary",
            summary_id: `summary:${event.sessionID}:${Date.now()}`,
            ...basePayload(ctx, String(event.sessionID || "")),
            summary_kind: "compaction",
            summary_length: truncate(event.summary).length,
            summary_text: "Compaction summary withheld for privacy.",
            created_at: new Date().toISOString(),
          })
          if (!result.ok) {
            await log(ctx, "warn", "summary capture failed", { stderr: result.stderr })
          }
          break
        }

        case "file.edited": {
          const result = await runMemory(ctx, ["capture"], {
            kind: "file_edit",
            touch_id: `file:${event.sessionID}:${Date.now()}:${event.path}`,
            ...basePayload(ctx, String(event.sessionID || "")),
            file_path: event.path,
            created_at: new Date().toISOString(),
          })
          if (!result.ok) {
            await log(ctx, "warn", "file capture failed", { stderr: result.stderr, path: event.path })
          }
          break
        }
      }
    },

    tool: {
      wspace_memory_wakeup: tool({
        description: "Retrieve advisory continuity context from the local wspace memory store",
        args: {
          limit: tool.schema.number().int().min(1).max(5).optional().describe("How many recent sessions to include"),
        },
        async execute(args, context): Promise<string> {
          if (!(await ensureMemoryReady(ctx))) {
            return "Wspace memory is unavailable."
          }

          const directory = typeof context.directory === "string" ? context.directory : pluginRoot(ctx)
          const project = ctx.project?.id || "wspace"
          const limit = typeof args.limit === "number" ? args.limit : 3
          const result = await runMemory({ ...ctx, directory }, ["wakeup", "--project", project, "--limit", String(limit)])
          if (!result.ok) {
            await log(ctx, "warn", "wakeup tool failed", { stderr: result.stderr })
            return "Wspace memory is unavailable."
          }
          return result.stdout || "No memory context available."
        },
      }),
    },
  }
}

export default WspaceMemoryPlugin
