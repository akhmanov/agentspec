---
name: openspec-continue-change
description: Continue an existing bead-mapped OpenSpec change by creating the next missing artifacts.
license: MIT
compatibility: Requires openspec CLI.
metadata:
  author: openspec
  version: "1.0"
  generatedBy: "1.3.1+repo"
---

Use this repo's `OpenSpec + Beads` contract.

1. Select the bead first.
   - If multiple active beads exist, ask the user to choose which bead to continue.
2. Resolve the mapped change from bead metadata `openspec.change`.
3. Run `openspec status --change "<change-name>" --json`.
4. For each artifact with `status: "ready"`, run `openspec instructions <artifact-id> --change "<change-name>" --json` and create it using the returned template, context, and rules.
5. Re-run status until the artifacts required for apply are complete.

Guardrails:

- Do not create a new change here; only continue an existing mapped one.
- Do not let OpenSpec replace bead identity or status.
- If bead metadata and change directory conflict, stop and ask.
