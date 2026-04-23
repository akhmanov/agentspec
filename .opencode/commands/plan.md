# Workspace Plan Mode

Create the execution plan for: $ARGUMENTS

Reference: `.opencode/skills/workspace-planning/SKILL.md`.

1. Read the latest approved spec revision first. If no approved spec exists, stop and direct the workflow to `/spec`.
2. Confirm that the spec declares `Execution Path: planned`. If it declares `direct-build`, stop and direct the workflow to `/build`.
3. Validate that the spec is executable. If it is ambiguous or missing execution structure, bounce back to `/spec`.
4. Build the execution topology: `repo_less`, `single-repo`, or `multi-repo`.
5. Produce dependency-aware execution slices with verification checkpoints.
6. Define the expected integration shape for `/ship`, including repo mapping and merged-state closure.
7. Save the canonical plan to `docs/plans/<bead-id>-<slug>.md` with explicit `Status`, `Revision`, and `Supersedes` header fields. New or revised plans stay `Status: Draft` until explicitly approved, then update the header to `Approved`.
8. Stop before implementation and wait for explicit approval.
