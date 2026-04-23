---
name: workspace-planning
description: Use when an approved wspace spec with `Execution Path: planned` must be turned into an execution plan with dependency order, repo topology, verification checkpoints, and integration shape.
---

# Workspace Planning

## Purpose

Turn a planned-path spec into an executable plan.

## Flow

1. Read the latest approved spec revision first.
2. Confirm that the spec declares `Execution Path: planned`. If not, return to `/build`.
3. Validate that the spec is executable. If not, return to `/spec`.
4. Define execution topology: `repo_less`, `single-repo`, or `multi-repo`.
5. Build the dependency graph and bounded execution slices.
6. Define verification checkpoints for each slice.
7. Define the expected integration shape for `/ship`.
8. Save the canonical plan to `docs/plans/<bead-id>-<slug>.md` with explicit `Status`, `Revision`, and `Supersedes` headers. New or revised plans stay `Status: Draft` until explicitly approved, then update the header to `Approved`.
9. Stop before implementation and wait for approval.

## Guardrails

- Do not re-solve the product problem here.
- Do not create a plan when the spec already chose `direct-build`.
- Favor the smallest executable plan that matches the approved spec.
- Treat boundary leaks as design risks.
- Keep repo-local verify and handoff policy out of the workspace plan.
