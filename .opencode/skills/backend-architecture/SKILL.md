---
name: backend-architecture
description: Use when reviewing or designing backend work in the workspace stack and you need to preserve transport, usecase, domain, adapter, storage, and config boundaries.
---

# Backend Architecture

## Purpose

Keep backend changes simple, boundary-safe, and aligned with the workspace service architecture.

## Focus

1. Identify which layer owns the change: transport, usecase, domain, adapter, storage, or config.
2. Push business logic out of handlers, controllers, CLI wrappers, and infra shells.
3. Check that domain rules stay cohesive and that dependencies still point inward.
4. Treat env/config access, persistence details, and provider SDK usage as boundary edges, not domain concerns.
5. Call out blast radius, failure modes, and backward-compatibility risks before implementation or review approval.

## Guardrails

- Prefer one clear flow over extra abstractions.
- Treat layer leaks as architecture bugs, not style issues.
- Be suspicious of cross-layer helpers that hide ownership.
