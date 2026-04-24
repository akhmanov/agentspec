---
description: Propose a bead-mapped OpenSpec change and generate apply-ready artifacts
---

Propose a bead-mapped OpenSpec change and create the artifacts needed for implementation.

When ready to implement, run `/opsx-apply`.

---

**Input**: The argument after `/opsx-propose` is either:

- a bead-prefixed change name like `wspace-q66-transition-workspace-lifecycle-to-openspec`
- a bead id plus a short description
- a plain description of what the user wants to build

**Steps**

1. **Recover or create the bead first**

   `beads` is the continuity layer in this repo.

   - If the user already gave a bead id, use it.
      - Do not silently change ownership of an explicitly provided bead.
      - If ownership, assignee, or intent is unclear, stop and ask before claiming it.
   - Otherwise check active and ready work first:
     ```bash
     bd list --status=in_progress
     bd ready
     ```
      - If you select a bead from `bd ready`, claim it before proceeding:
        ```bash
        bd update <bead-id> --claim
        ```
   - If no suitable bead exists, ask whether to create one with `bd q`.
      - If creating a new bead, capture the id and claim it before proceeding:
        ```bash
        BEAD_ID=$(bd q "<title>")
        bd update "$BEAD_ID" --claim
        ```

   Do not create an OpenSpec change without a bead.

2. **Derive the change name from the bead**

   The change name MUST be prefixed with the bead id.

   Example:

   - bead `wspace-q66`
   - change `wspace-q66-transition-workspace-lifecycle-to-openspec`

   If the user provided only a description, derive a slug and prefix it with the bead id.

3. **Check or set the bead mapping**

   - Inspect bead metadata for `openspec.change`.
   - If it exists and matches the derived change name, reuse it.
   - If it exists and conflicts, stop and ask before proceeding.
   - If it is missing, set it:
     ```bash
     bd update <bead-id> --set-metadata openspec.change=<change-name>
     ```

4. **Create the change directory if needed**

   ```bash
   openspec new change "<change-name>"
   ```

5. **Get the artifact build order**

   ```bash
   openspec status --change "<change-name>" --json
   ```

6. **Create artifacts in sequence until apply-ready**

   For each artifact with `status: "ready"`:

   - Get instructions:
     ```bash
     openspec instructions <artifact-id> --change "<change-name>" --json
     ```
   - Read dependency artifacts for context.
   - Apply the returned `context`, `rules`, and `template`.
   - Preserve the repo contract:
     - include the bead traceability section requested by the schema
     - treat `tasks.md` as an execution checklist, not the task manager

   Continue until every artifact named in `applyRequires` is `done`.

7. **Show final status**

   ```bash
   openspec status --change "<change-name>"
   ```

**Guardrails**

- Never create or continue an unprefixed change in this repo.
- Never let OpenSpec replace bead identity or status.
- If the bead metadata and change directory disagree, stop and ask.
- If the change already exists, ask whether to continue it instead of creating a duplicate.
