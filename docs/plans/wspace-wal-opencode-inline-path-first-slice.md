# Plan: OpenCode Inline and Path First Slice

- Bead: `wspace-wal`
- Status: Approved
- Revision: 1
- Supersedes: none
- Spec: `docs/specs/wspace-wal-opencode-inline-path-first-slice.md`
- Topology: `single-repo`
- Execution repo: `aw`
- Repo entry: workspace root

## Plan Validation

- The spec is approved.
- The spec declares `Execution Path: planned`.
- The spec remains executable.
- This plan stays inside the approved first slice: OpenCode only, `inline` and `path` only, with `ARCH.md` landing before feature implementation.

The planning work does not reopen product direction. It turns the approved scope into an execution sequence for one repo: `aw`.

## Execution Topology

This is `single-repo` work in `aw`.

- `aw` owns the canonical spec, canonical plan, architecture contract, Go module, CLI, adapters, sync logic, and tests.
- No project clone, external worktree, or cross-repo handoff is required for this slice.
- Execution is bounded to the repo root and these paths:
- `ARCH.md`
- `go.mod`
- `go.sum`
- `cmd/aw/`
- `internal/`
- `docs/specs/`
- `docs/plans/`
- test fixtures and Go tests

## Dependency Graph

Execution dependency order:

1. Slice A establishes the repo contract and transport foundation.
2. Slice B adds config parsing, validation, selector resolution, and normalized model flow.
3. Slice C adds the OpenCode adapter plus ownership-safe sync behavior.
4. Slice D adds verification breadth, CLI flow coverage, and closeout evidence.

Optional parallelism:

- After Slice A stabilizes package layout and CLI entrypoints, parts of Slice B test fixtures can be drafted in parallel.
- After Slice B stabilizes resolved-model shape, portions of Slice D adapter fixtures can be drafted in parallel with Slice C.
- Default execution should stay mostly sequential because adapter and sync behavior depend on the final config, resolve, and model contracts.

## Slice A: Architecture Contract And CLI Foundation

Goal:
Establish the repo architecture, Go module, and thin CLI transport without mixing in business rules.

Scope:

- Add `ARCH.md` as the repo-level architecture contract.
- Initialize `go.mod` and first dependencies.
- Add `cmd/aw` with a minimal `urfave/cli` app and the command shell for `init` and `sync`.
- Define the initial package layout:
  - `cmd/aw`
  - `internal/config`
  - `internal/resolve`
  - `internal/model`
  - `internal/adapter/opencode`
  - `internal/sync`
- Keep command handlers transport-thin and return errors upward.
- Decide where reusable filesystem helpers belong, if any, without inventing a generic infra layer.

Outputs:

- The repo has one stable architecture document.
- The Go module builds.
- The CLI starts, exposes the intended command surface, and is ready to wire real behavior.

Verification checkpoint:

- `go test ./...` passes with the initial module and any foundation tests present.
- `go run ./cmd/aw --help` shows the expected top-level command surface.
- `ARCH.md` matches the approved spec boundaries: thin CLI, target-neutral resolve/config, adapter-owned rendering, sync-owned apply behavior.

Exit criteria:

- Feature implementation can proceed without re-deciding project structure.

## Slice B: Config, Validation, Local Resolution, And Normalized Model

Goal:
Turn `aw.yaml` into validated, target-neutral resolved resources for the approved local selectors.

Scope:

- Define config types for `sections`, `commands`, `agents`, and `skills`.
- Enforce exactly one selector per resource.
- Accept only `inline` and `path` in this slice.
- Implement `aw init` starter-config generation.
- Implement config loading from `aw.yaml`.
- Implement validation errors that identify resource kind, resource id, and selector problems clearly.
- Implement `inline` resolution.
- Implement `path` resolution with resource-sensitive rules:
  - file-only for `sections`, `commands`, and `agents`
  - file-or-directory for `skills`
- Normalize resolved resources into a model that the adapter can consume without reparsing raw config.

Outputs:

- `aw init` can write the starter config.
- `aw.yaml` can be parsed and validated for the first slice.
- Resolved resources flow into one normalized model independent of OpenCode path layout.

Verification checkpoint:

- `go test ./...` passes for config, validation, and resolution packages.
- Tests cover invalid selector combinations, unsupported selectors, missing path targets, and skill file-versus-directory behavior.
- `go run ./cmd/aw init` writes the expected starter `aw.yaml` in a temp fixture or equivalent CLI test path.

Exit criteria:

- The adapter can be written against a stable resolved model rather than raw YAML structures.

## Slice C: OpenCode Rendering And Ownership-Safe Sync

Goal:
Materialize the approved resources for OpenCode while preserving the ownership boundary from the spec.

Scope:

- Implement the OpenCode adapter for:
  - managed instruction-file sections
  - `.opencode/commands/<id>.md`
  - `.opencode/agents/<id>.md`
  - `.agents/skills/<id>/...`
- Finalize the OpenCode instruction-file placement used for managed section blocks.
- Implement marker-based section updates that preserve foreign content outside `aw` blocks.
- Implement desired-state sync/apply for create and update.
- Implement orphan cleanup for managed outputs only.
- Add deterministic ownership rules for outputs so a matching path alone does not imply ownership.
- Introduce `internal/state` only if orphan handling cannot be implemented safely without it.

Outputs:

- `aw sync --opencode` can render the first-slice resources.
- Managed file updates are idempotent.
- Orphan cleanup is bounded to `aw`-owned outputs and managed instruction blocks.

Verification checkpoint:

- `go test ./...` passes for adapter and sync packages.
- Tests prove that foreign files are preserved when paths collide.
- Tests prove that content outside `aw` markers is preserved.
- Tests prove single-file versus directory `skills` materialize to the correct `.agents/skills/<id>/...` layout.
- One fixture-based CLI test or integration-style package test proves repeated `aw sync --opencode` runs are idempotent.

Exit criteria:

- The repo has one safe end-to-end materialization path for OpenCode local sources.

## Slice D: CLI Flow Coverage And Closeout Verification

Goal:
Gather fresh evidence that the first slice works end-to-end and is ready for review without hidden assumptions.

Scope:

- Add or finish end-to-end oriented tests around `aw init` and `aw sync --opencode`.
- Run the narrow package tests from earlier slices plus the full repo test suite.
- Run one bounded manual smoke path against a temp fixture workspace with:
  - `inline` and `path` examples
  - at least one managed section
  - at least one command
  - at least one agent
  - both a single-file skill and a directory skill
- Confirm idempotent re-sync and orphan cleanup behavior in the smoke path.
- Capture enough execution notes for `/review` and `/ship` to understand what was verified.

Outputs:

- Fresh automated evidence for config, resolution, adapter rendering, sync safety, and CLI flows.
- One manual proof that the approved first slice behaves correctly in a realistic temp workspace.

Verification checkpoint:

- `go test ./...` passes.
- The manual smoke path shows expected OpenCode outputs and preserved foreign content.
- Re-running sync produces no unintended diffs.
- Removing a managed resource removes only the corresponding managed output.

Exit criteria:

- The change set is ready for `/review` with current evidence instead of inference.

## Cross-Slice Constraints

- Do not add `http`, `github`, `gitlab`, or `--claude-code` in this plan.
- Do not let config or resolve code learn OpenCode-specific output paths.
- Do not let the OpenCode adapter parse selectors or raw YAML.
- Do not let CLI handlers accumulate business logic that belongs in config, resolve, adapter, or sync layers.
- Do not treat pre-existing files as owned solely because they live at an `aw` target path.
- Do not add a state database unless deterministic ownership-safe cleanup cannot be achieved otherwise.
- Do not use `ARCH.md` as a place to hide speculative abstractions for remote providers.

## Verification Strategy

Each slice should verify the narrowest stable contract first and keep earlier checks alive as later slices land.

Verification layers for this plan:

- unit-level and package-level Go tests for config, validation, resolve, adapter, and sync behavior
- CLI-oriented tests for `aw init` and `aw sync --opencode`
- fixture-based ownership and marker-preservation tests
- one manual smoke path against a temp workspace

The closeout loop should not claim remote-source readiness, Claude Code support, or long-term migration guarantees. Verification is only for the approved local OpenCode first slice.

## Expected Integration Shape For `/ship`

Repo mapping:

- `aw`: canonical spec, canonical plan, `ARCH.md`, Go module, CLI implementation, adapter/sync logic, and tests

Expected merged-state shape:

- `aw` contains the approved spec, this plan, and the implementation for the first OpenCode local-source slice.
- `ARCH.md` documents the project architecture and coding rules used by the implementation.
- `aw init` writes the starter config.
- `aw sync --opencode` materializes `sections`, `commands`, `agents`, and `skills` from `inline` and `path` sources.
- The repo test suite protects the ownership boundary and idempotent sync behavior.

Closure conditions:

- The `aw` branch reaches merged state under the repo-local review and handoff policy.
- `go test ./...` passes on the final change set.
- A manual smoke path confirms correct OpenCode materialization, preserved foreign content, idempotent re-sync, and bounded orphan cleanup.
- No hidden follow-up work for remote selectors or Claude Code support is required to call this first slice complete.

## Bounce Conditions

Return to `/spec` before `/build` if any of these become true during execution:

- OpenCode section materialization requires schema-visible target configuration instead of adapter-owned placement.
- Ownership-safe orphan cleanup cannot be achieved without a broader state or adoption model than the approved scope allows.
- Supporting `skills` path semantics safely requires broader schema or packaging rules than the approved first slice defines.
- The project needs `http`, `github`, `gitlab`, or Claude Code support to make the first slice useful enough to validate.
- The architecture needs a broader layer model or cross-package contract than the approved light clean-architecture shape.

## Next Transition

Wait for approval.

If approved, proceed to `/build` using this plan and the approved spec as the execution source.
