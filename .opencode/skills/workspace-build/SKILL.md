---
name: workspace-build
description: Use when executing approved wspace work from either a direct-build spec or an approved plan while preserving hidden workspace routing and repo-local ownership boundaries.
---

# Workspace Build

## Purpose

Execute the approved work without exposing the kernel mechanics as user-facing workflow.

## Flow

1. Confirm the bead and the latest approved spec revision.
2. If the spec declares `Execution Path: planned`, confirm the latest approved plan revision too; otherwise use the latest approved spec revision as the execution source.
3. Execute one slice at a time.
4. For `repo_less` work, stay in `wspace/`.
5. For repo-bound work, pick the next repo from the plan or spec topology and enter it with `mise run work:start <repo> <task>` before changing files.
6. After repo entry, follow the nearest repo-local `AGENTS.md`.
7. If implementation uncovers broken exploration, spec, or planning assumptions, bounce back to `/explore`, `/spec`, or `/plan` instead of silently widening scope.

## Guardrails

- `wspace` owns routing, not repo-specific implementation policy.
- A `direct-build` spec is still a formal workflow artifact, not an informal shortcut.
- Do not turn `/build` into a hidden substitute for `/test`, `/review`, or `/ship`.
