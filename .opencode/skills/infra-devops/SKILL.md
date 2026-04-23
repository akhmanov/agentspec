---
name: infra-devops
description: Use when changing infrastructure, runtime environments, build or delivery mechanics, and you need to preserve operator workflows, rollout safety, and ownership boundaries.
---

# Infra DevOps

## Purpose

Keep infrastructure and delivery changes operable, reversible, and aligned with the workspace repo split.

## Focus

1. Identify whether the change belongs to infra bootstrap, runtime operations, platform build/publish, CI/CD, or workload delivery.
2. Prefer stable repo entrypoints such as checked-in task runners over ad-hoc shell sequences.
3. Check rollout order, rollback path, secret handling, and environment-specific blast radius.
4. Keep ownership boundaries clear across infra, platform, cloud, and gitops repos.
5. Call out observability gaps, manual steps, and operator traps before the change ships.

## Guardrails

- Operational convenience does not justify hidden state or one-off commands.
- Prefer explicit runbooks and repeatable entrypoints.
- Treat missing rollback and missing verification as release blockers.
