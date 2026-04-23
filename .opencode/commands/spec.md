# Workspace Spec Mode

Create or update the canonical spec for: $ARGUMENTS

Reference: `.opencode/skills/workspace-spec/SKILL.md`.

1. Start from an approved `/explore` outcome or an explicit developer direction.
2. Restate the chosen direction and surface any remaining assumptions.
3. Decide whether this is `repo_less`, `single-repo`, `multi-repo`, or still unknown.
4. Write explicit scope, not-doing, success criteria, architecture impact, and why this direction was chosen.
5. Record the canonical artifact header fields `Status`, `Revision`, and `Supersedes`. New or revised specs stay `Status: Draft` until explicitly approved, then update the header to `Approved`. Record `Execution Path` as either `direct-build` or `planned` and justify the choice.
6. Call out blast radius, failure modes, and boundary concerns that execution must preserve.
7. Save the canonical spec to `docs/specs/<bead-id>-<slug>.md`.
8. Stop and wait for explicit approval. After approval, go to `/plan` only when `Execution Path` is `planned`; otherwise go straight to `/build`.
