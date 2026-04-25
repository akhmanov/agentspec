---
description: Archive a completed change in the experimental workflow
---

Archive a completed change in the experimental workflow.

**Input**: Optionally specify a change name after `/opsx-archive` (e.g., `/opsx-archive add-auth`). If omitted, check if it can be inferred from conversation context. If vague or ambiguous, prompt for available changes.

**Steps**

1. **If no change name provided, prompt for selection**

   Run `openspec list --json` to get available changes, then ask the user to select.

   Show only active changes (not already archived).
   Include the schema used for each change if available.

   **IMPORTANT**: Do NOT guess or auto-select a change. Always let the user choose.

2. **Check artifact completion status**

   Run `openspec status --change "<name>" --json` to check artifact completion.

   Parse the JSON to understand:
   - `schemaName`: The workflow being used
   - `artifacts`: List of artifacts with their status (`done` or other)

   **If any artifacts are not `done`:**
   - Display a warning listing incomplete artifacts
   - Ask the user whether to continue anyway
   - Proceed only if the user confirms

3. **Check task completion status**

   Read the tasks file (typically `tasks.md`) to check for incomplete tasks.

   Count tasks marked with `- [ ]` (incomplete) vs `- [x]` (complete).

   **If incomplete tasks are found:**
   - Display a warning showing the count of incomplete tasks
   - Ask the user whether to continue anyway
   - Proceed only if the user confirms

   **If no tasks file exists:** Proceed without a task-related warning.

4. **Choose archive mode**

   Check whether delta specs exist at `openspec/changes/<name>/specs/`.

   - If delta specs exist, use the normal archive flow so OpenSpec updates main specs:
     ```bash
     openspec archive "<name>"
     ```
   - If no delta specs exist, archive with spec syncing skipped:
     ```bash
     openspec archive "<name>" --skip-specs
     ```
   - If delta specs exist but the user explicitly wants to skip spec updates, confirm that choice and use:
     ```bash
     openspec archive "<name>" --skip-specs
     ```

   Do not use ad hoc file moves. Let `openspec archive` handle validation and archive placement.

5. **Perform the archive**

   Run the chosen `openspec archive` command.

6. **Display summary**

   Show archive completion summary including:
   - Change name
   - Schema that was used
   - Archive location
   - Spec sync status (synced / skipped)
   - Note about any warnings (incomplete artifacts/tasks)

**Output On Success**

```
## Archive Complete

**Change:** <change-name>
**Schema:** <schema-name>
**Archived to:** openspec/changes/archive/YYYY-MM-DD-<name>/
**Specs:** ✓ Synced to main specs

All artifacts complete. All tasks complete.
```

**Output On Success With Warnings**

```
## Archive Complete (with warnings)

**Change:** <change-name>
**Schema:** <schema-name>
**Archived to:** openspec/changes/archive/YYYY-MM-DD-<name>/
**Specs:** Sync skipped

**Warnings:**
- Archived with 2 incomplete artifacts
- Archived with 3 incomplete tasks
- Spec syncing was skipped

Review the archive if this was not intentional.
```

**Guardrails**
- Always prompt for change selection if not provided.
- Use artifact graph (`openspec status --json`) for completion checking.
- Do not block archive on warnings; inform and confirm.
- Show a clear summary of what happened.
- Use the real `openspec archive` command instead of manually moving directories.
- If delta specs exist, default to syncing specs through `openspec archive`.
- Use `--skip-specs` only when there are no delta specs or the user explicitly requests skipping spec updates.
