---
name: openspec-apply-change
description: Implement tasks from a bead-mapped OpenSpec change.
license: MIT
compatibility: Requires openspec CLI.
metadata:
  author: openspec
  version: "1.0"
  generatedBy: "1.3.1"
---

Use `beads` as the continuity source and OpenSpec as the artifact source.

1. Select the bead first.
2. Resolve the mapped change from bead metadata `openspec.change`.
3. Run `openspec status --change "<change-name>" --json` and `openspec instructions apply --change "<change-name>" --json`.
4. If apply is blocked, send the user to `/opsx-continue <change-name>`.
5. Read all context files, implement tasks, and update task checkboxes immediately.

Guardrails:

- Use `/opsx-continue` when a mapped change exists but still needs proposal, specs, design, or tasks artifacts.
- If bead metadata is missing or conflicts with the requested change, stop and ask.
- Treat `tasks.md` as an execution checklist, not the task manager.
