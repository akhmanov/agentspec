---
name: openspec-propose
description: Propose a bead-mapped OpenSpec change and generate apply-ready artifacts.
license: MIT
compatibility: Requires openspec CLI.
metadata:
  author: openspec
  version: "1.0"
  generatedBy: "1.3.1"
---

Use this repo's `OpenSpec + Beads` contract.

1. Recover or create the bead first with `bd list --status=in_progress` or `bd ready`.
   - Ask before creating a new bead with `bd q`.
   - If the bead came from `bd ready`, claim it with `bd update <bead-id> --claim` before proceeding.
   - If the bead was created with `bd q`, capture the id and claim it before proceeding.
   - If the user provided a bead id explicitly, do not silently change ownership; stop and ask if ownership or intent is unclear.
2. Derive the change name from the bead id. Change names MUST be bead-prefixed.
3. Read or set bead metadata `openspec.change=<change-name>`.
4. Create the change with `openspec new change "<change-name>"` if it does not exist.
5. Use `openspec status --change "<change-name>" --json` and `openspec instructions <artifact> --change "<change-name>" --json` to create artifacts until apply-ready.
6. Preserve the repo contract in generated artifacts:
   - include bead traceability when the schema asks for it
   - treat `tasks.md` as an execution checklist, not the task manager

Guardrails:

- Never create or continue an unprefixed change in this repo.
- Never let OpenSpec replace bead identity or status.
- If bead metadata and change directory conflict, stop and ask.
