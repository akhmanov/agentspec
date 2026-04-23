---
name: api-contracts
description: Use when changing HTTP or service contracts and you need to preserve validation, error shape, compatibility, and clear transport-to-domain boundaries.
---

# API Contracts

## Purpose

Keep request and response contracts explicit, validated, and safe to evolve.

## Focus

1. Identify the contract surface: route, RPC, webhook, internal service call, or CLI boundary.
2. Validate inputs and outputs at the transport edge.
3. Keep transport DTOs separate from domain models when their lifecycles differ.
4. Check auth, authz, error shape, pagination/filter semantics, and backward compatibility.
5. Call out consumer impact, migration needs, and failure behavior when contracts change.

## Guardrails

- Do not smuggle domain behavior into request parsing.
- Do not break existing consumers silently.
- Prefer additive evolution over incompatible rewrites unless the migration path is explicit.
