---
name: workspace-ship
description: Use when a wspace task is ready for integration and you need to define the route to merged completion, keep deploy out of scope, and close the workspace only after merged state.
---

# Workspace Ship

## Purpose

Own the integration contract for the workspace lifecycle.

## Flow

1. Confirm that the latest approved artifact revision still reflects the intended integration path and that `/review` and `/test` are already satisfied with fresh evidence.
2. Describe the integration shape: repos involved, expected PR/MR units, and merged-state closure conditions.
3. If review comments or failing checks send the work backward, bounce to `/build` or `/test` instead of pretending shipping is still in progress.
4. Close the bead only after merged completion.
5. Keep deploy, release, and rollout automation out of v1 scope.

## Guardrails

- `/ship` is an integration contract in v1, not a full provider automation engine.
- For multi-repo work, track merged completion across the whole repo set before closing the workspace.
