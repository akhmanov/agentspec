---
name: gitops-review
description: Use when reviewing desired-state, promotion, or rollout changes and you need to reason about deployment safety, drift, and merged-state correctness.
---

# GitOps Review

## Purpose

Review GitOps changes as deployment intent, not just YAML edits.

## Focus

1. Identify the desired-state unit being changed: workload config, image tag, release plan, policy, or shared runtime artifact.
2. Check ownership boundaries between build/publish, promotion, and deploy truth.
3. Review rollout sequencing, migration ordering, readiness checks, and rollback behavior.
4. Look for drift risk, accidental fan-out, and cluster-wide blast radius.
5. State the expected merged-state outcome and the operator evidence needed to trust it.

## Guardrails

- Do not treat GitOps as passive config when it encodes release behavior.
- Small diff does not mean small blast radius.
- Prefer explicit promotion intent over implicit tag movement.
