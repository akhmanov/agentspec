---
description: Implement tasks from a bead-mapped OpenSpec change
---

Implement tasks from a bead-mapped OpenSpec change.

**Input**: Optionally specify a bead id or a bead-prefixed change name. If omitted, recover the active bead first.

**Steps**

1. **Select the bead first**

   `beads` is the primary continuity layer.

   - If the input is a bead id, use it.
   - If the input is a change name, infer the bead id from the prefix.
   - Otherwise recover active work:
     ```bash
     bd list --status=in_progress
     ```
   - If multiple active beads exist, ask the user to choose.

2. **Resolve the mapped change**

   - Read `openspec.change` from the bead metadata.
   - If it is missing, ask the user whether to create or map a change first with `/opsx-propose`.
   - If it conflicts with the provided change name, stop and ask.

   Always announce the bead id and the mapped change name.

3. **Check status to understand the schema**

   ```bash
   openspec status --change "<change-name>" --json
   ```

4. **Get apply instructions**

   ```bash
   openspec instructions apply --change "<change-name>" --json
   ```

   Handle states explicitly:

   - If `state: "blocked"`, show what artifacts are missing and direct the user to `/opsx-continue <change-name>`.
   - If `state: "all_done"`, show that all tasks are complete and suggest `/opsx-archive <change-name>`.
   - Otherwise continue.

5. **Read context files**

   Read every file path listed under `contextFiles`.

6. **Implement tasks**

   For each pending task:

   - make the required changes
   - keep changes minimal and focused
   - mark the task checkbox complete immediately
   - continue until done or blocked

7. **Show completion or pause status**

   Include:

   - bead id
   - change name
   - task progress
   - reminder that bead state still lives in `beads`

**Guardrails**

- Treat the bead as the continuity source and the OpenSpec change as the artifact source.
- Always read context files before starting.
- If tasks are unclear or artifacts are missing, stop and ask instead of guessing.
- Use `/opsx-continue` when a mapped change exists but still needs proposal, specs, design, or tasks artifacts.
