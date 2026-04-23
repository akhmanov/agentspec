# Workspace Debate Mode

Run a bounded advisory debate for: $ARGUMENTS

Reference: `.opencode/skills/workspace-debate/SKILL.md`.

1. Restate the contested decision and the current lifecycle context if it is known.
2. Identify 1-3 grounded candidate options from the provided context before dispatching role agents.
3. Run the mandatory core panel: `user-advocate`, `architect`, and `skeptic`.
4. Add `security` or `operator` only when the question needs a specific missing lens.
5. Keep the debate bounded to at most 2 rounds and in-memory round state.
6. Return an advisory result with `Recommended decision`, `Alternatives considered`, `Strongest objection`, `Why objection did not win`, `Residual risks`, `Confidence`, `Escalation`, and `Next`.
7. Stay read-only and do not silently move the lifecycle forward.
