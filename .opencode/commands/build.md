# Workspace Build Mode

Execute the approved lifecycle work for: $ARGUMENTS

Reference: `.opencode/skills/workspace-build/SKILL.md`.

1. Confirm that the latest approved spec revision exists.
2. If the spec uses `Execution Path: planned`, confirm that the latest approved plan revision exists too; otherwise use the latest approved spec revision as the execution source.
3. Execute one bounded slice at a time.
4. For `repo_less` work, stay in `wspace/`.
5. For repo-bound work, use hidden workspace routing to choose the next repo and enter it through `mise run work:start <repo> <task>` before changing files.
6. Follow the nearest repo-local `AGENTS.md` after entering an owning repo.
7. Do not silently expand scope beyond the approved spec, and do not treat `direct-build` as license to skip verification or review.
