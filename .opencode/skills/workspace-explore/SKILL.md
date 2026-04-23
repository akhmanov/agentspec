---
name: workspace-explore
description: Use when a wspace task needs broad exploration before formal lifecycle commitment, or when an in-flight change needs to reopen the problem instead of blindly executing.
---

# Workspace Explore

## Purpose

Explore the problem widely before committing the workspace lifecycle to a direction.

## Flow

1. Restate the problem, question, or uncertainty and whether this is pre-task exploration or an in-flight side-loop.
2. Explore the relevant code, docs, history, and repo topology broadly enough to understand the whole picture.
3. Clarify the user need, success criteria, constraints, and trust boundaries that matter.
4. Assess whether the task is worth doing now, whether there is a simpler framing, and whether the current request would make the system worse.
5. Compare 2-3 viable directions, including defer or do-nothing when appropriate, and recommend one.
6. End with an explicit next transition: stay exploratory, create or reuse a bead and move to `/spec`, or bounce back to `/spec` or `/plan` if execution assumptions broke.

## Guardrails

- `/explore` is an optional side-loop, not a mandatory first stage of the lifecycle.
- Do not create a canonical artifact by default. Use durable output only when the developer asks for it or when the result must survive the session.
- Read-only exploration may start before a bead exists. Create or reuse a bead before durable docs, notes, or repo changes.
- Do not let `/explore` quietly turn into `/spec` or `/plan`; commitment and execution design happen in those modes.
