# Wspace OpenCode Instructions

`wspace` is the workspace kernel for `assistant-wi` plus an opinionated OpenCode lifecycle shell.

## Boundaries

- Use lifecycle slash commands as the primary workspace UX.
- Treat `bd`, `mise run projects:*`, `mise run work:start`, and `mise run validate:finish` as the canonical backend primitives.
- Do not move product logic, repo-local verify pipelines, PR flow, or handoff policy into `wspace`.
- After entering a target repo or managed worktree, follow the nearest repo-local `AGENTS.md`.
- Use `.opencode/skills/baseline.toml` as the curated workspace skill baseline. Imported baseline skills are installed project-local under `.agents/skills/`, while local baseline skills stay under `.opencode/skills/`.
- `docs/specs/<bead-id>-<slug>.md` is the canonical commitment artifact.
- `docs/plans/<bead-id>-<slug>.md` is the canonical execution artifact when the spec chooses a planned execution path.
- Canonical spec and plan headers record `Bead`, `Status`, `Revision`, and `Supersedes`; execution should follow the latest approved revision in scope.
- New or revised artifacts stay `Status: Draft` until explicitly approved, then update the header to `Approved`.
- Advisory memory, when present, stays in `memory-local/` and does not replace `.beads/` or canonical docs.

## Operating Model

1. Recover active work with `bd list --status=in_progress` before looking for new work.
2. If the user did not provide a bead id, use `bd ready` or create a new bead with `bd q` and claim it.
3. `/explore` is the optional side-loop for broad research, alternatives, and worth-doing checks before commitment, or when an in-flight change must reopen the problem.
4. `/debate` is the optional advisory-only, read-only side-loop for contested decisions that can be pressure-tested through internal subagent debate before asking the human to break the tie.
5. `/spec` is mandatory for substantial work. It records the chosen direction, scope, success criteria, canonical revision headers, and `Execution Path`.
6. `/plan` only turns the latest approved planned-path spec revision into an execution graph. If the spec chooses `direct-build`, go straight to `/build`.
7. Classify execution as `repo_less`, `single-repo`, or `multi-repo`.
8. Stay in `wspace/` for shaping, planning, and workspace-only docs changes.
9. For repo-bound build work, pick the repo from `projects/manifest.toml` and enter it with `mise run work:start <repo> <task>`.
10. `/test` verifies the correct layer and sends the workflow back into repair if checks fail.
11. `/ship` in v1 is an integration contract. It aims at merged state and workspace closure, but does not own deploy or release automation.
12. Use `mise run validate:finish <repo-path>` only as a local git guardrail.
13. If the `wspace_memory_wakeup` tool is available and you are starting substantive work in `wspace`, call it once near the start to recover advisory continuity context before searching old sessions manually.

## Commands

- `/explore`: investigate a task before formal lifecycle commitment, or reopen it when assumptions break.
- `/debate`: run an optional advisory-only, read-only internal debate before escalating a contested decision.
- `/spec`: shape a request into a draft spec with explicit revision headers.
- `/spec`: commit a request into a canonical spec with an execution path, then wait for explicit approval before execution.
- `/plan`: create an execution plan from the latest approved planned-path spec revision.
- `/build`: execute approved work from either the latest approved direct-build spec revision or the latest approved plan revision.
- `/test`: verify at the right layer and debug failures.
- `/review`: perform findings-first pre-integration review.
- `/ship`: prepare the route to merged state and close the workspace only after merged completion.
