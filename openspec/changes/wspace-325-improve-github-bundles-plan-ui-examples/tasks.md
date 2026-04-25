## Traceability

- Bead: `wspace-325`
- Change: `wspace-325-improve-github-bundles-plan-ui-examples`
- Task state authority: `beads`

Tasks in this file are an execution checklist. Ownership, status, assignee, priority, and dependencies live in `beads`.

## 1. GitHub Bundle Resolver

- [x] 1.1 Add resolver tests that prove directory-backed GitHub skills can fall back from raw-file lookup to a git-based bundle read without relying on archive download size.
- [x] 1.2 Add the shallow-clone GitHub bundle fallback with `--depth 1` and `GIT_LFS_SKIP_SMUDGE=1`.
- [x] 1.3 Preserve existing root-`SKILL.md`, frontmatter, and safe-path validation for cloned bundles.

## 2. Plan Preview Output

- [x] 2.1 Add CLI tests for grouped default plan output, grouped conflicts, verbose output, and the unchanged empty-plan message.
- [x] 2.2 Add `--verbose` to `agentspec plan` and implement grouped rendering without changing preview/apply diff computation.

## 3. Repository Examples And Docs

- [x] 3.1 Add `example/local-smoke` with committed local resources and a README that documents deterministic smoke commands.
- [x] 3.2 Add `example/github-smoke` with pinned live public GitHub sources and a README that documents the live smoke commands and drift expectations.
- [x] 3.3 Update repo docs to describe the GitHub bundle behavior and the short versus verbose `plan` modes.

## 4. Verification

- [x] 4.1 Run targeted resolver and CLI tests, then run `go test ./...`.
- [x] 4.2 Validate the OpenSpec change and manually exercise both example workspaces.
