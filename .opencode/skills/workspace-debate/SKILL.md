---
name: workspace-debate
description: Use when a workspace question is contested and the agent should run a bounded advisory debate across internal subagents before escalating to the human.
---

# Workspace Debate

## Purpose

Resolve a contested workspace decision through a bounded internal debate before asking the human to break the tie.

## Flow

1. Restate the question, current lifecycle context, and trust boundaries that matter.
2. Identify 1-3 candidate options grounded in the available evidence.
3. Always run the core panel:
- `debate-user-advocate`
- `debate-architect`
- `debate-skeptic`
4. Add `debate-security` or `debate-operator` only when a concrete missing lens appears.
5. Round 1: collect positions, arguments, risks, confidence, and what would change each role's mind.
6. Summarize the strongest objections and unresolved points.
7. Round 2 only if needed. Pass a structured in-memory round summary instead of a raw transcript.
8. Return an advisory result with:
- `Recommended decision`
- `Alternatives considered`
- `Strongest objection`
- `Why objection did not win`
- `Residual risks`
- `Confidence`
- `Escalation`
- `Next`

## Guardrails

- `/debate` is an optional side-loop, not a mandatory lifecycle phase.
- `/debate` is advisory-only and read-only in v1.
- Keep the mandatory core panel of `user-advocate`, `architect`, and `skeptic`.
- Default to 3 roles. Expand only when a specific missing lens appears, and never exceed 5 roles in v1.
- Never run more than 2 rounds.
- Keep round state in memory only. Do not create repo-backed transcripts or hidden persistent state.
- Do not silently move into `/spec`, `/plan`, `/build`, or `/ship`.
- If requirements are missing or a blocker remains on security, ownership, destructive operations, or lifecycle semantics, return `Escalation: required`.
- Preserve dissent. Do not flatten disagreement into false consensus.
