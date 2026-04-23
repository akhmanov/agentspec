# Workspace Ship Mode

Prepare the integration path for: $ARGUMENTS

Reference: `.opencode/skills/workspace-ship/SKILL.md`.

1. Treat `/ship` as the integration phase after `/review` for the latest approved artifact revision.
2. Confirm the change is ready for integration and that `/test` has fresh evidence for the latest approved spec and plan revision in scope.
3. Describe the expected integration shape: repo set, expected PR/MR units, and merged-state closure.
4. In v1, do not pretend to own deploy or release automation.
5. Close the bead only after merged completion, not merely after local implementation.
