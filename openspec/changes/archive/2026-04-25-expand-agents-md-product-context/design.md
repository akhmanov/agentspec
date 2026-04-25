## Context

`AGENTS.md` is the first repo-level briefing an agent sees, so it needs to explain not only the product label but also the mental model the agent should use when reasoning about the repository. The current file already captures product boundaries, architecture boundaries, lifecycle, and retained skills, but it still assumes the reader already understands why `agentspec` exists, what the four resource kinds mean in practice, and what `plan` and `apply` actually do.

The authoritative product semantics already exist elsewhere:

- `SPEC.md` explains the problem statement, goals, non-goals, and CLI contract.
- `ARCH.md` explains the package boundaries and desired-state flow.
- `internal/adapter/opencode` shows how each resource kind materializes into the workspace.
- `internal/sync` shows that `plan` and `apply` are ownership-aware sync operations, not generic workspace rewrites.

The design problem is therefore not missing product truth; it is missing compression. `AGENTS.md` needs a compact way to surface that truth to an agent quickly.

## Goals / Non-Goals

**Goals:**

- Make `AGENTS.md` explain why `agentspec` exists and which drift problem it solves.
- Define `sections`, `commands`, `agents`, and `skills` in terms of their workspace effect.
- Explain `plan --opencode` and `apply --opencode` as safe desired-state sync operations with ownership boundaries.
- Keep the file scannable for first-contact agent use.

**Non-Goals:**

- Turn `AGENTS.md` into a full user manual for `agentspec`.
- Duplicate large sections of `SPEC.md` or `ARCH.md`.
- Introduce new product behavior, workflow behavior, or target support.
- Document implementation details that are only useful when editing Go code.

## Decisions

### Decision: Keep `AGENTS.md` as a compact explainer, not a deep reference

`AGENTS.md` should stay short enough to function as a repo briefing. The expansion should add meaning, not bulk.

Alternatives considered:

- Deep mini-reference in `AGENTS.md`
  Rejected because it would compete with `SPEC.md` and `ARCH.md` and become harder for agents to scan.
- Keep the current label-only structure
  Rejected because it leaves too much product understanding implicit.

### Decision: Explain the product in flow order

The document should explain the product in this order:

1. why `agentspec` exists
2. what it manages
3. how sync works
4. what it is not

This order matches how a new agent forms a mental model: purpose first, then surface area, then behavior, then boundaries.

Alternatives considered:

- Resource-first ordering
  Rejected because schema labels without purpose are harder to interpret correctly.
- Architecture-first ordering
  Rejected because package boundaries matter less than product intent during initial onboarding.

### Decision: Define resources by workspace effect

The resource model should be described in terms of where each resource ends up in the workspace:

- `sections` -> managed instruction blocks in `AGENTS.md`
- `commands` -> target command documents such as `.opencode/commands/<id>.md`
- `agents` -> target agent documents such as `.opencode/agents/<id>.md`
- `skills` -> directory-based skill bundles such as `.agents/skills/<id>/...`

This is more useful than repeating the YAML keys alone because it tells an agent what each resource actually means in the rendered workspace.

### Decision: Explain `plan` and `apply` as sync semantics, not just CLI verbs

The briefing should make explicit that:

- `plan --opencode` previews the managed create/update/delete/conflict set without writing files
- `apply --opencode` materializes the desired managed state, updates managed sections, prunes orphaned owned outputs, and refuses to silently overwrite foreign content

Alternatives considered:

- Only list the commands in the core flow
  Rejected because it does not communicate the ownership-safety contract.

### Decision: Prefer references over duplication

`AGENTS.md` should summarize the product model and point to `SPEC.md` and `ARCH.md` for full detail. This reduces the risk of the briefing drifting into an inconsistent second specification.

## Risks / Trade-offs

- [Risk] `AGENTS.md` becomes too verbose for agent onboarding
  -> Mitigation: keep each explanatory section compact and avoid examples unless they clarify a concept faster than prose.

- [Risk] The briefing oversimplifies product behavior and drifts from code
  -> Mitigation: anchor the wording to `SPEC.md`, `ARCH.md`, `internal/adapter/opencode`, and `internal/sync`.

- [Risk] The document starts duplicating canonical specs
  -> Mitigation: describe only the mental model and boundaries; leave full product detail in the canonical docs.

## Migration Plan

1. Expand `AGENTS.md` with a short "Why It Exists" section.
2. Rewrite the resource model section so each resource type is explained by workspace effect.
3. Add a sync-model section that explains `plan` and `apply` as ownership-aware sync operations.
4. Keep the existing architecture, repo rules, lifecycle, and skill sections intact unless wording changes are needed for consistency.

Rollback is a simple revert of the `AGENTS.md` content change.

## Open Questions

- None at the moment. The main design choice, compact explainer versus deep reference, has already been resolved in favor of the compact explainer.
