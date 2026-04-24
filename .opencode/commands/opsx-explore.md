---
description: Explore ideas and requirements without implementing, using beads for continuity and OpenSpec for artifact context
---

Enter explore mode. Think deeply, investigate the repo, and clarify requirements without implementing.

**IMPORTANT**: Explore mode is for thinking, not application code changes.

---

## Continuity First

Start from `beads`, not from OpenSpec.

At the start, recover context with:

```bash
bd list --status=in_progress
bd ready
```

If a bead already has `openspec.change` metadata, use that mapped change for OpenSpec artifact context.

Use `openspec list --json` only as a secondary view of the artifact layer, not as the primary continuity source.

## What Explore Mode Does Here

- ask clarifying questions
- investigate the codebase
- compare options
- surface risks and unknowns
- read mapped OpenSpec artifacts when they exist
- offer to formalize a bead-mapped change with `/opsx-propose`

## Guardrails

- Do not implement application code.
- Do not create an unprefixed OpenSpec change.
- Do not let OpenSpec replace bead identity or status.
- If an active pre-cutover bead is still on the legacy path, use `docs/LEGACY_WORKFLOW.md` for its process context.
