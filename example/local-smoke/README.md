# Local Smoke Example

This example is deterministic: every source file lives in this repository, and the smoke commands do not require network access.

Run the smoke from a temp copy or a fresh checkout of the repository so `apply` does not mutate the committed fixture in place. This is a repository validation flow, not the primary installation path. From this example directory:

- `cd ../.. && cp -R . /tmp/agentspec-local-smoke-repo && cd /tmp/agentspec-local-smoke-repo/example/local-smoke`
- `go run ../.. plan --target opencode`
- `go run ../.. plan --verbose --target opencode`
- `go run ../.. apply --target opencode`
- `go run ../.. plan --target opencode`

Expected behavior:

- the first `plan` reports three creates for `AGENTS.md#workspace-core`, `.opencode/commands/explore.md`, and `.opencode/skills/local-audit/SKILL.md`
- `apply` writes those managed files
- the final `plan` prints `No managed changes.`
