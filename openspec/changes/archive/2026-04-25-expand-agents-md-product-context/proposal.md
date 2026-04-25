## Why

`AGENTS.md` currently tells an agent what `agentspec` is called and which high-level boundaries matter, but it does not yet explain the product purpose or the mental model behind the resource types and sync flow. That makes first-contact reasoning weaker than it should be: an agent can see the names `sections`, `commands`, `agents`, `skills`, `plan`, and `apply`, but not why those concepts exist or how they affect the workspace.

## What Changes

- Expand `AGENTS.md` from a label-only briefing into a compact product explainer.
- Explain why `agentspec` exists and which workspace-maintenance problem it solves.
- Define the four top-level resource types in terms of their workspace effect, not only their schema name.
- Explain `plan --opencode` and `apply --opencode` as safe desired-state sync operations with ownership boundaries.
- Keep `AGENTS.md` concise enough for agent onboarding and avoid turning it into a duplicate of `SPEC.md` or `ARCH.md`.

## Capabilities

### New Capabilities
- `repo-agent-briefing`: Defines the required product context and sync-model guidance that the repo briefing in `AGENTS.md` must provide to an agent.

### Modified Capabilities

## Impact

- Affects `AGENTS.md` content and the repo's agent-onboarding surface.
- Aligns briefing text more closely with the product semantics already described in `SPEC.md`, `ARCH.md`, `internal/adapter/opencode`, and `internal/sync`.
- Does not change the CLI surface or runtime behavior of `agentspec`.
