---
name: openspec-archive-change
description: Archive a completed bead-mapped OpenSpec change.
license: MIT
compatibility: Requires openspec CLI.
metadata:
  author: openspec
  version: "1.0"
  generatedBy: "1.3.1"
---

Archive through the repo's `OpenSpec + Beads` contract.

1. Select the bead first.
2. Resolve the mapped change from bead metadata `openspec.change`.
3. Check artifact and task completion with `openspec status --change "<change-name>" --json` and the change `tasks.md`.
4. Warn and confirm before archiving with incomplete work.
5. Archive with `openspec archive "<change-name>"`.

Guardrails:

- Never archive a change that is not mapped back to a bead.
- Do not reference `openspec-sync-specs`; this repo uses the generated core workflow set.
- Keep `beads` as the task-state system after archive.
