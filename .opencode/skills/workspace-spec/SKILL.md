---
name: workspace-spec
description: Use when a wspace direction is chosen and you need to commit to scope, success criteria, execution path, and architectural boundaries in a canonical spec before implementation.
---

# Workspace Spec

## Purpose

Commit to the chosen solution before execution.

## Flow

1. Start from an approved `/explore` outcome or an explicit developer direction.
2. Restate the chosen solution and surface any remaining assumptions.
3. Classify the topology as `repo_less`, `single-repo`, `multi-repo`, or `unknown`.
4. Record why this direction is worth doing, plus scope, not-doing, success criteria, blast radius, and architectural impact.
5. Record the canonical artifact header fields `Status`, `Revision`, and `Supersedes`. New or revised specs stay `Status: Draft` until explicitly approved, then update the header to `Approved`. Choose `Execution Path` as `direct-build` or `planned` and justify it.
6. Record the boundary, failure-mode, and backward-compatibility concerns that execution must preserve.
7. Save the canonical spec to `docs/specs/<bead-id>-<slug>.md`.
8. Wait for approval before `/plan` or `/build`, depending on the chosen execution path.

## Guardrails

- `/spec` is mandatory for substantial work even when `/explore` is skipped.
- `/spec` is the commitment artifact, not a broad research phase.
- If the direction is still fuzzy, contested, or not worth doing, return to `/explore` instead of forcing a spec.
