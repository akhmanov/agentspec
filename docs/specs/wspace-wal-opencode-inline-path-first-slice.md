# Spec: OpenCode Inline and Path First Slice

- Bead: `wspace-wal`
- Status: Approved
- Revision: 1
- Supersedes: none
- Topology: `single-repo`
- Target repo: `aw`
- Execution Path: `planned`

## Chosen Direction

Build the first implementation slice of `aw` as a single-repo Go CLI focused on OpenCode only.

This slice will:

- add a repo-level `ARCH.md` before feature implementation
- initialize the Go project and CLI transport with `urfave/cli`
- implement `aw init` to write a starter `aw.yaml`
- implement `aw sync --opencode`
- support only the local source selectors `inline` and `path`
- support those selectors across `sections`, `commands`, `agents`, and `skills`
- preserve the v1 ownership rules from `SPEC.md` for managed files and managed instruction-file regions
- defer all network-backed source selectors and the Claude Code adapter

The implementation should follow a light clean-architecture shape rather than a ceremonial one: thin CLI transport, target-neutral config and resolve layers, target-specific adapter code, and a sync layer that applies desired state safely.

## Why This Direction

This direction is worth doing because it validates the core product loop without expanding the trust boundary too early.

`inline` and `path` are enough to prove that `aw.yaml` can drive repeatable workspace materialization, while OpenCode-only delivery keeps the first adapter concrete and testable. Deferring `http`, `github`, `gitlab`, and `--claude-code` removes the largest sources of ambiguity for v1-first execution: remote fetching, authentication, rate limits, provider-specific failure handling, and partial-target behavior.

Adding `ARCH.md` first is also worth doing because the repo currently has no implementation. It gives the project one durable contract for package boundaries, dependency direction, error-handling rules, and the level of abstraction expected from modern Go code.

## Remaining Assumptions

- `SPEC.md` remains the product-level source of truth for schema and ownership semantics; this spec only narrows the first execution slice.
- The first slice should cover all four resource kinds, but only for `inline` and `path` selectors.
- OpenCode instruction-file placement stays adapter-owned and must not leak into `aw.yaml`, even if the concrete target file choice is finalized during implementation.
- The repo can adopt a new Go module and its initial dependency set now without needing backward-compatibility shims for older code, because no `aw` implementation exists yet.
- `ARCH.md` is a repo-level architecture contract, not a second spec or a replacement for package-level documentation.

## Scope

1. Initialize the `aw` Go module and repository layout needed for the first CLI implementation.
2. Add `ARCH.md` before feature code, documenting project goals, package boundaries, dependency direction, style expectations, and the specific rules adopted from the Uber Go style guide.
3. Add `cmd/aw` as the CLI entrypoint using `urfave/cli` with a thin transport layer.
4. Implement `aw init` to write a starter `aw.yaml` with the four top-level maps from `SPEC.md` and no aggressive workspace discovery.
5. Implement config loading and validation for the first slice of `aw.yaml`, including `sections`, `commands`, `agents`, and `skills` with exactly one selector per resource and only `inline` or `path` allowed.
6. Implement local source resolution for `inline` and `path`.
7. Implement the normalized internal model needed to pass resolved resources into target adapters without leaking raw config shape into sync logic.
8. Implement the OpenCode adapter for the first slice so it can render managed instruction-file sections for `sections`, `.opencode/commands/<id>.md` for `commands`, `.opencode/agents/<id>.md` for `agents`, and `.agents/skills/<id>/...` for `skills`, including file and directory semantics for `path`.
9. Implement sync/apply behavior for create, update, idempotent re-run, and orphan cleanup of `aw`-owned outputs.
10. Add tests for config validation, source resolution, adapter rendering, sync ownership behavior, and the main CLI flows for the first slice.

## Not Doing

- No `http`, `github`, or `gitlab` support in this slice.
- No `aw sync --claude-code` in this slice.
- No packs, `aw check`, `aw doctor`, or migration features.
- No aggressive adoption or normalization of pre-existing workspaces.
- No target filtering in schema.
- No multi-file inline resources.
- No generic remote-provider abstraction introduced before the first local slice is working.
- No persistence or state database unless orphan ownership cannot be implemented safely with deterministic file inspection.
- No framework-heavy clean architecture with empty interfaces, service shells, or repository abstractions that do not yet protect a real boundary.

## Success Criteria

- `aw init` writes a readable starter `aw.yaml` matching the first-slice schema.
- `aw sync --opencode` can materialize `sections`, `commands`, `agents`, and `skills` from `inline` and `path` sources.
- `path` supports file resolution for `sections`, `commands`, and `agents`.
- `path` supports file and directory resolution for `skills`.
- Single-file `skills` materialize as `.agents/skills/<id>/SKILL.md`.
- Directory `skills` materialize as `.agents/skills/<id>/...` and require `SKILL.md` at the bundle root.
- Re-running `aw sync --opencode` without config changes is idempotent.
- `aw` updates or deletes only `aw`-owned files and only modifies managed section blocks inside instruction files.
- Sync removes orphaned managed outputs when resources are removed from `aw.yaml`.
- Validation errors are explicit about resource id, resource kind, and invalid selector shape.
- Tests cover the ownership boundary strongly enough to catch accidental rewrites of foreign files or foreign instruction-file regions.

## Architecture Impact

- Introduces the initial Go module, dependency set, and CLI transport for the repo.
- Establishes `ARCH.md` as the repo-level architecture contract before implementation spreads across packages.
- Sets the package direction for the first slice as `cmd/aw` for CLI transport, `internal/config` for parsing and validation, `internal/resolve` for selector resolution, `internal/model` for normalized resource forms, `internal/adapter/opencode` for target rendering, `internal/sync` for desired-state application and orphan cleanup, and `internal/state` only if deterministic ownership tracking proves insufficient.
- Keeps business rules out of CLI command handlers and out of adapter code that should only render target-specific outputs.
- Uses dependency injection only at real edges such as filesystem access, target adapter selection, and future network clients, instead of pre-emptive abstractions.
- Aligns implementation style with modern Go expectations from the Uber guide: one process exit in `main`, no `init()` magic, explicit errors, minimal globals, and table-driven tests.

## Execution Path

`planned`

This work should use a planned path because it is the first implementation in the repo and the order matters.

`ARCH.md`, package layout, config rules, path semantics, adapter rendering, ownership-safe sync, and tests should be sequenced deliberately. The first slice also includes multiple behavioral edges that need explicit checkpoints: file versus directory skills, instruction-file markers, orphan cleanup, and idempotent re-sync behavior.

## Blast Radius

- `.beads/`
- `ARCH.md`
- `go.mod`
- `go.sum`
- `cmd/aw/**`
- `internal/**`
- `docs/specs/wspace-wal-opencode-inline-path-first-slice.md`
- test fixtures and Go tests

No external repos, no remote providers, and no non-OpenCode targets should change in this slice.

## Failure Modes To Guard Against

- `aw sync --opencode` rewrites foreign files or foreign instruction-file content outside `aw` markers.
- A pre-existing file under an `aw`-target path is treated as owned just because the path matches.
- Resolver code grows target-specific knowledge about `.opencode`, instruction files, or marker format.
- Adapter code reparses config or re-implements selector resolution instead of consuming normalized resolved resources.
- `path` handling allows incorrect file-versus-directory behavior for `skills`.
- Orphan cleanup deletes foreign files or entire directories when only managed files should be removed.
- The initial architecture grows too many interfaces and layers before the first behavior is proven.
- `aw init` starts discovering or migrating existing workspace state, violating the v1 constraint.
- Future network support is anticipated so aggressively that the first local slice becomes harder to read and test.

## Boundary Concerns To Preserve During Execution

- `aw` is a declarative sync tool, not a workflow runtime, package manager, bootstrap engine, or generic merge tool.
- `aw.yaml` stays target-neutral at the schema level; target behavior belongs in adapters.
- Resolver logic must not know about target file layout.
- Adapter logic must not own selector parsing or source fetching.
- CLI code must remain transport-only and should not absorb config, resolve, or sync business logic.
- Ownership is the safety boundary: only `aw`-owned files and explicit `aw`-managed instruction blocks may be changed.
- `ARCH.md` should reinforce simple boundaries and DI at edges, not justify abstraction-heavy code.

## Backward-Compatibility Concerns To Preserve During Execution

- The `aw.yaml` shape for `inline` and `path` must remain compatible with adding `http`, `github`, and `gitlab` later without breaking the local-source configuration model.
- OpenCode output paths and marker conventions introduced in this slice should remain stable enough that later source selectors do not require a migration of already materialized local-source outputs.
- Adding `--claude-code` later must not force different semantics for already working OpenCode resources.
- `ARCH.md` should describe stable project rules rather than temporary implementation accidents.

## Initial Verification Intent

The planning phase should define verification for at least these checkpoints:

- `aw init` writes the expected starter config
- config validation rejects invalid selector combinations and invalid skill path shapes
- `inline` and `path` resolution preserve expected content and file bundle behavior
- OpenCode rendering produces the expected file layout for all four resource kinds
- sync is idempotent on repeated runs
- orphan cleanup removes only managed outputs
- managed instruction-file updates preserve foreign content outside markers

## Next Transition

Wait for approval.

If approved, proceed to `/plan` because the `Execution Path` is `planned`.
