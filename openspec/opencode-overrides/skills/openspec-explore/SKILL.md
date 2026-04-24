---
name: openspec-explore
description: Explore ideas without implementing, using beads for continuity and OpenSpec for artifact context.
license: MIT
compatibility: Requires openspec CLI.
metadata:
  author: openspec
  version: "1.0"
  generatedBy: "1.3.1"
---

Explore mode is for thinking, not implementation.

Start from `beads`, not from OpenSpec:

```bash
bd list --status=in_progress
bd ready
```

If a bead already has `openspec.change` metadata, use that mapped change for OpenSpec artifact context.

Use `openspec list --json` only as a secondary view of the artifact layer.

Guardrails:

- Do not implement application code.
- Do not create an unprefixed OpenSpec change.
- Do not let OpenSpec replace bead identity or status.
- If an active pre-cutover bead is still on the legacy path, use `docs/LEGACY_WORKFLOW.md` for process context.
