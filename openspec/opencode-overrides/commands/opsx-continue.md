---
description: Create the next missing artifacts for an existing bead-mapped OpenSpec change
---

Create the next missing artifacts for an existing bead-mapped OpenSpec change.

**Input**: Optionally specify a bead id or a bead-prefixed change name.

**Steps**

1. Select the bead first.
    - Recover active work with `bd list --status=in_progress` if no bead was provided.
    - If multiple active beads exist, ask the user which bead to continue before proceeding.
    - If the input is a change name, infer the bead id from the prefix.

2. Resolve the mapped change from bead metadata `openspec.change`.
   - If it is missing, stop and direct the user to `/opsx-propose`.

3. Inspect artifact status.
   ```bash
   openspec status --change "<change-name>" --json
   ```

4. For each artifact with `status: "ready"`:
   - fetch instructions with `openspec instructions <artifact-id> --change "<change-name>" --json`
   - read completed dependency artifacts for context
   - create the artifact using the returned template, context, and rules
   - preserve bead traceability and the `OpenSpec + Beads` contract

5. Re-run `openspec status --change "<change-name>" --json` after each artifact.
   - Stop when all artifacts required for apply are `done`.

6. Show final status and tell the user whether `/opsx-apply` is now ready.

**Guardrails**

- Do not create a new change here; only continue an existing mapped one.
- Do not let OpenSpec replace bead identity or status.
- If the bead metadata and change directory disagree, stop and ask.
