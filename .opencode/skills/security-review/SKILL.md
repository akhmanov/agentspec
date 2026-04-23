---
name: security-review
description: Use when work touches trust boundaries, auth, secrets, config, or privileged operations and you need a practical security review before implementation or merge.
---

# Security Review

## Purpose

Review changes adversarially so security-sensitive work does not rely on optimism.

## Focus

1. Identify assets, trust boundaries, and who controls each input.
2. Check authn, authz, secret handling, config exposure, and privileged actions.
3. Look for injection surfaces, unsafe file or shell access, and missing validation.
4. Check logging, auditability, rotation, revocation, and failure behavior.
5. State blast radius and rollback strategy plainly when the change can lock out users, leak data, or widen access.

## Guardrails

- Security review is findings-first, not reassurance-first.
- Treat implicit trust as a bug until proven otherwise.
- Prefer explicit deny/default-safe behavior at boundaries.
