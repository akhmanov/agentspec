---
description: Archive a completed bead-mapped OpenSpec change
---

Archive a completed bead-mapped OpenSpec change.

**Input**: Optionally specify a bead id or a bead-prefixed change name after `/opsx-archive`.

**Steps**

1. **Select the bead first**

   - If the input is a bead id, use it.
   - If the input is a change name, infer the bead id from the prefix.
   - Otherwise recover active work with `bd list --status=in_progress` and ask the user to choose.

2. **Resolve the mapped change**

   - Read `openspec.change` from bead metadata.
   - If it is missing, stop and ask before archiving.
   - If it conflicts with the requested change name, stop and ask.

3. **Check completion state**

   - Run `openspec status --change "<change-name>" --json`.
   - Warn if any artifacts are incomplete.
   - Read `tasks.md` if present and warn if any checkboxes remain incomplete.
   - Ask the user for confirmation before archiving with warnings.

4. **Archive through the CLI**

   Use the CLI instead of hand-moving directories:

   ```bash
   openspec archive "<change-name>"
   ```

5. **Report the result**

   Summarize:

   - bead id
   - change name
   - archive location
   - any warnings the user accepted
   - reminder that bead closure still follows repo policy in `beads`

**Guardrails**

- Never archive a change that is not mapped back to a bead.
- Prefer `openspec archive` over manual filesystem moves.
- Do not reference `openspec-sync-specs`; this repo uses the generated core workflow set.
- Keep `beads` as the task-state system even after archive.
