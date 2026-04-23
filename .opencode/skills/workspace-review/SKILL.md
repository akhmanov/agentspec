---
name: workspace-review
description: Use when reviewing a wspace lifecycle change before integration and you need a findings-first review that checks boundaries, verification gaps, and failure modes before summaries.
---

# Workspace Review

## Purpose

Run the pre-integration review with architecture and workflow boundaries in mind.

## Review Order

1. Findings first.
2. Check for boundary leaks between `wspace` and owning repos.
3. Check auth/authz, validation, failure modes, and backward-compatibility risk.
4. Call out missing verification, integration gaps, missing repo-local handoff, and whether findings require revising the current artifact revision or bouncing back to `/spec`.
5. Keep summaries brief and secondary.

## Guardrails

- Do not approve changes just because workspace-side checks pass.
- Distinguish workspace shell concerns from repo product concerns.
- `/review` happens before `/ship`, not instead of it.
