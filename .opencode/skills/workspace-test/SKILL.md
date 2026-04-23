---
name: workspace-test
description: Use when verifying a wspace change and you need to test at the correct layer, gather fresh evidence, and return the lifecycle to repair when checks fail.
---

# Workspace Test

## Purpose

Verify the current slice or change at the right layer and feed failures back into repair.

## Flow

1. Identify the verification layer: workspace, owning repo, or both.
2. Run the commands that prove the current slice works.
3. If checks fail, localize the failure and send the lifecycle back into repair before `/ship`.
4. Use `mise run validate:finish <repo-path>` only as a workspace-side git guardrail.
5. Report what is actually verified and what still remains.

## Guardrails

- `/test` is a verify + debug gate, not a greenwashing phase.
- Never claim completion without fresh evidence.
