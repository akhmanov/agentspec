# Workspace Review Mode

Review the requested changes: $ARGUMENTS

Reference: `.opencode/skills/workspace-review/SKILL.md`.

1. Treat this as the pre-integration review before `/ship`.
2. Report findings first.
3. Prioritize boundary leaks, auth/authz issues, validation gaps, failure modes, and missing verification.
4. Distinguish workspace-kernel concerns from owning-repo concerns, and state whether findings require revising the current artifact revision or bouncing back to `/spec`.
5. Keep summaries brief and secondary.
