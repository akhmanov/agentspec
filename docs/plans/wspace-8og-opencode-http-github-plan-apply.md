# Plan: OpenCode HTTP and GitHub With Plan and Apply

- Bead: `wspace-8og`
- Status: Approved
- Revision: 1
- Supersedes: none
- Spec: `docs/specs/wspace-8og-opencode-http-github-plan-apply.md`
- Topology: `single-repo`
- Execution repo: `aw`
- Repo entry: workspace root

## Plan Validation

- The latest spec revision is approved.
- The spec declares `Execution Path: planned`.
- The spec is executable as written for one repo without reopening product scope.
- Execution stays inside the approved scope: OpenCode only, `http` and `github` added on top of existing `inline` and `path`, with `plan` and `apply` replacing `sync`.

Planning assumptions that stay within the approved spec:

- `plan` is read-only and computes its output from the live workspace plus current source resolution.
- `apply` recomputes desired state instead of consuming a saved plan artifact.
- Remote fetch support in this slice is limited to HTTPS and unauthenticated public endpoints unless execution proves a broader auth requirement, in which case the workflow returns to `/spec`.

## Execution Topology

This is `single-repo` work in `aw`.

- `aw` owns the canonical spec, canonical plan, CLI surface, config validation, remote resolution, OpenCode adapter, sync logic, and tests.
- No external repo edits, cross-repo handoffs, or Claude-target changes are required.
- Execution is bounded to:
- `docs/specs/`
- `docs/plans/`
- `cmd/aw/`
- `internal/config/`
- `internal/resolve/`
- `internal/model/`
- `internal/adapter/opencode/`
- `internal/sync/`
- Go tests and fixtures

## Dependency Graph

Execution dependency order:

1. Slice A establishes one shared plan/apply contract so the new CLI surface has a safe backend.
2. Slice B extends validation and remote transport seams without leaking target-specific logic.
3. Slice C implements `http` and `github` resolution against those seams.
4. Slice D wires the CLI end to end, removes `sync` from the public surface, and closes verification.

Optional parallelism:

- After Slice A stabilizes the plan/report contract, parts of Slice B and Slice D test scaffolding can be drafted in parallel.
- After Slice B stabilizes the remote transport seam, HTTP and GitHub resolution tests in Slice C can proceed in parallel with CLI output formatting in Slice D.
- Default execution should stay mostly sequential because the remote resolver and CLI flow both depend on the final shared diff contract.

## Slice A: Shared Plan and Apply Contract

Goal:
Create one source of truth for managed diffs, ownership conflicts, and apply decisions so `plan` and `apply` cannot drift apart.

Scope:

- Replace the CLI command shell from `sync` to `plan` and `apply` while keeping handlers transport-thin.
- Introduce or refine model types for planned operations and conflicts.
- Refactor `internal/sync` so diff computation and write execution are separate steps owned by the same package.
- Keep the OpenCode adapter output contract stable.
- Preserve the existing state and ownership rules while exposing them through a read-only planning entrypoint.

Outputs:

- A shared sync-layer contract that can both preview and apply managed changes.
- A CLI surface ready to call `plan` without side effects.

Verification checkpoint:

- `go test ./...` passes.
- CLI tests or command wiring tests prove `aw` exposes `init`, `plan`, and `apply`.
- A package-level test proves the planning path does not write files, markers, or state in a temp workspace.

Exit criteria:

- The repo has one authoritative diff path for both preview and apply behavior.

## Slice B: Config Validation and Remote Transport Edge

Goal:
Extend selector validation and network access at the boundary without mixing those concerns into adapters or CLI code.

Scope:

- Allow `http` and `github` selectors in config while preserving exactly-one-selector enforcement.
- Enforce HTTPS-only `http` inputs.
- Validate GitHub selector shape with required `repo`, `ref`, and `path` fields.
- Reject unsafe or ambiguous selector paths before fetch.
- Introduce the smallest transport seam needed for remote fetches so tests can run against local servers instead of live network calls.
- Bound remote fetch behavior with explicit timeout, size, and error-wrapping rules.

Outputs:

- Stable validation rules for the new selectors.
- A resolver-owned remote fetch seam suitable for deterministic tests.

Verification checkpoint:

- `go test ./...` passes.
- Config tests cover valid and invalid `http` and `github` selector shapes.
- Tests prove non-HTTPS URLs, missing GitHub fields, and unsafe paths fail with resource context.

Exit criteria:

- Remote bytes can be requested through a bounded resolver edge without leaking target layout.

## Slice C: HTTP and GitHub Resolution

Goal:
Resolve remote resources into the same normalized model already used by the OpenCode adapter.

Scope:

- Implement `http` single-file resolution for `sections`, `commands`, `agents`, and single-file `skills`.
- Implement GitHub file resolution for `sections`, `commands`, `agents`, and single-file `skills`.
- Implement GitHub directory resolution for bundled `skills` by fetching a ref snapshot and extracting only the selected subtree.
- Preserve existing skill validation rules, including required root `SKILL.md` and frontmatter checks.
- Keep remote resolution target-neutral and independent of `.opencode` or `AGENTS.md` knowledge.

Outputs:

- Resolved remote resources that are indistinguishable from local ones to downstream packages.
- Deterministic error handling for fetch failures, missing files, missing bundle roots, and invalid skill content.

Verification checkpoint:

- `go test ./...` passes.
- Resolver tests cover HTTP success and failure paths using local test servers.
- Resolver tests cover GitHub file and directory paths, including bundle extraction, missing root `SKILL.md`, and path traversal defenses.

Exit criteria:

- `internal/adapter/opencode` can consume remote-backed resources without any code changes specific to remote selectors.

## Slice D: CLI Integration and Closeout Verification

Goal:
Ship the user-visible `plan` and `apply` workflow with evidence that preview and apply stay aligned.

Scope:

- Wire `aw plan --opencode` to load config, resolve resources, render desired state, compute managed diff, and print the result without writes.
- Wire `aw apply --opencode` to reuse the same diff logic and then execute writes.
- Remove `sync` from the public CLI surface and replace its tests with `plan` and `apply` coverage.
- Preserve idempotence, orphan cleanup, and foreign-content protection.
- Add end-to-end style tests around `plan` then `apply` on temp workspaces with remote fixtures.

Outputs:

- Final CLI surface: `init`, `plan`, `apply`.
- End-to-end evidence that `plan` previews what `apply` performs.

Verification checkpoint:

- `go test ./...` passes.
- `go run ./cmd/aw --help` shows `init`, `plan`, and `apply`.
- A bounded smoke path demonstrates:
- `plan` makes no filesystem changes
- `apply` performs the previewed managed updates
- a second `plan` after `apply` reports no managed changes
- ownership conflicts surface in preview and still block writes

Exit criteria:

- The change set is ready for `/review` with fresh evidence.

## Cross-Slice Constraints

- `plan` must never write files, markers, or state.
- `plan` and `apply` must share one sync-layer diff and ownership policy source of truth.
- Resolver code must remain target-neutral.
- Adapter code must not fetch remote content or compute diffs.
- Remote fetch support in this slice is HTTPS-only and unauthenticated.
- No cache, lockfile, or persisted download store may be introduced in this plan.
- No `gitlab` or Claude-target work may leak into execution.

## Verification Strategy

Verification layers for this plan:

- unit and package tests for config validation, remote resolution, plan diff behavior, and apply behavior
- CLI-oriented tests for `aw plan --opencode` and `aw apply --opencode`
- resolver tests backed by local HTTP test servers instead of live network dependency
- fixture-based ownership and marker-preservation tests
- one bounded smoke path in a temp workspace covering remote and local resources together

The closeout loop should prove this slice only: OpenCode plus `inline`, `path`, `http`, and `github`, with `plan` and `apply` as the CLI contract.

## Expected Integration Shape For `/ship`

Repo mapping:

- `aw`: canonical spec, canonical plan, CLI implementation, config validation, remote resolution, shared plan/apply sync logic, and tests

Expected merged-state shape:

- `docs/specs/wspace-8og-opencode-http-github-plan-apply.md` is approved and reflected by implementation.
- `docs/plans/wspace-8og-opencode-http-github-plan-apply.md` is approved and reflected by implementation.
- `cmd/aw` exposes `init`, `plan`, and `apply`.
- `internal/resolve` supports `inline`, `path`, `http`, and `github` within the approved scope.
- `internal/sync` provides shared preview and apply logic.
- The repo test suite protects remote resolution, ownership boundaries, preview/apply consistency, and idempotence.

Closure conditions:

- The `aw` branch reaches merged state under repo-local review and handoff policy.
- `go test ./...` passes on the final change set.
- A smoke path confirms `plan` is read-only, `apply` matches previewed changes, and unchanged workspaces plan cleanly after apply.
- No hidden follow-up work for `gitlab` or Claude is required to call this slice complete.

## Bounce Conditions

Return to `/spec` before `/build` if any of these become true during execution:

- Supporting GitHub directory-backed skills safely requires schema-visible selector changes.
- The repo needs authenticated or private GitHub access to make the slice usable.
- Sharing one diff engine between preview and apply requires a broader ownership or state model than the approved spec allows.
- Removing `sync` reveals a concrete compatibility requirement that the approved spec did not capture.
- Remote fetch support requires cache or persistence behavior that widens the product boundary beyond this slice.

## Next Transition

Wait for approval.

If approved, proceed to `/build` using this plan and the approved spec as the execution source.
