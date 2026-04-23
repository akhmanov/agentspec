# Workspace Test Mode

Verify the current lifecycle change for: $ARGUMENTS

Reference: `.opencode/skills/workspace-test/SKILL.md`.

1. Verify at the right layer: workspace, owning repo, or both.
2. Run the checks that prove the current slice works.
3. If checks fail, localize the failure and bounce the workflow back into repair before attempting `/ship`.
4. Use `mise run validate:finish <repo-path>` only as a workspace-side git guardrail, never as proof of full correctness.
5. Do not claim completion without fresh evidence.
