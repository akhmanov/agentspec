# Spec: Rename aw Product Surface to agentspec

- Bead: `wspace-jrq`
- Status: Approved
- Revision: 2
- Supersedes: Revision 1
- Topology: `single-repo`
- Target repo: `aw`
- Execution Path: `direct-build`

## Chosen Direction

Rename the product surface from `aw` to `agentspec` as a hard cut.

This revision chooses one steady-state product identity and removes all legacy rename support rather than carrying migration logic during development.

This slice will:

- rename the user-facing CLI from `aw` to `agentspec`
- rename the default config file from `aw.yaml` to `agentspec.yaml`
- rename the Go module and internal import paths away from `aw`
- update canonical product docs, architecture docs, tests, fixtures, and examples to the new product name
- use only `agentspec` ownership artifacts in steady state, including state paths, state owner identifiers, and managed section markers
- remove legacy `aw` adoption and migration logic from the product surface
- remove runtime GitHub base URL override env vars from the supported product surface
- keep target-specific output paths such as `.opencode/` and `.agents/` unchanged unless they are directly tied to the product namespace

This is a hard-cut product rename. It is separate from the repo-local OpenSpec workflow migration tracked in `wspace-q66`.

## Why This Direction

This direction is worth doing because the project is still in active development and does not need compatibility scaffolding for an older shipped contract.

Carrying legacy `aw` support now would increase surface area, complicate the sync layer, and make the rename harder to reason about without delivering real user value. A hard cut keeps the product honest: one binary name, one config name, one ownership namespace, and one supported runtime contract.

Removing the runtime GitHub base URL override env vars follows the same principle. They are not required for the current development-stage product and would widen the product surface without being part of the core rename value.

## Remaining Assumptions

- `agentspec` is the chosen long-term product name, not a temporary repo alias.
- Development-stage simplicity is more valuable than preserving adoption or migration paths for legacy `aw` workspaces.
- No shipped workflow depends on `aw.yaml`, `.aw/state`, `aw` markers, or `AW_*` / `AGENTSPEC_*` GitHub override env vars.
- The current approved plan revision for `wspace-jrq` is no longer the execution source because this spec revision removes the migration slice and chooses `direct-build`.

## Scope

1. Rename the CLI transport, help text, command invocation examples, and user-facing docs from `aw` to `agentspec`.
2. Rename the default config filename from `aw.yaml` to `agentspec.yaml` across implementation, tests, docs, and examples.
3. Rename the Go module path and internal imports away from `aw`.
4. Rename repo-local architecture and product documentation so they describe `agentspec` consistently.
5. Rename steady-state test fixtures and expectations that currently hard-code `aw`, `aw.yaml`, `.aw/state`, or `aw` markers.
6. Remove legacy `aw` migration and adoption logic from the CLI, resolver surface, sync layer, and tests.
7. Remove runtime GitHub base URL override env vars from the supported product surface.
8. Preserve target-specific output layout that is independent of the product name, such as `.opencode/commands`, `.opencode/agents`, and `.agents/skills`.

## Not Doing

- No backward compatibility for `aw` CLI invocation.
- No fallback from `agentspec.yaml` to `aw.yaml`.
- No adoption or migration of legacy `.aw/state` artifacts.
- No adoption or migration of legacy `aw:section:*` markers.
- No `-config`, `-chdir`, or any other new execution flags in this slice.
- No Terraform-style global CLI behavior in this slice.
- No OpenSpec workflow migration work from `wspace-q66` in this slice.
- No new target support such as Claude Code.
- No new selector support such as GitLab.
- No change to target-owned output locations that are not product-branded.

## Success Criteria

- The shipped CLI name is `agentspec`, not `aw`.
- The default config file written and read by the product is `agentspec.yaml`.
- Canonical product and architecture docs describe `agentspec` consistently.
- The Go module and internal imports no longer use `aw` as the module path.
- Managed section markers and persisted state written by the product use only the `agentspec` namespace.
- The runtime product surface no longer supports legacy `aw` migration behavior.
- The runtime product surface no longer supports GitHub base URL override env vars.
- Tests cover the new steady-state contract without depending on legacy migration behavior.

## Architecture Impact

- `cmd/aw` and any remaining `aw` command wiring should be removed in favor of one repository-root `agentspec` transport surface.
- `internal/config`, `internal/resolve`, `internal/adapter/opencode`, and `internal/sync` remain the main affected layers because the rename crosses transport, config discovery, ownership persistence, and marker semantics.
- The sync layer should become simpler than Revision 1 by removing legacy ownership loading, cleanup, and marker upgrade paths.
- Resolve should keep testability through internal seams and test fixtures, but runtime env-based GitHub endpoint overrides are no longer part of the product contract.
- Target-neutral resolve logic should remain focused on config and sources; it should not absorb target file layout or legacy migration policy.

## Execution Path

`direct-build`

This work now fits a direct-build path because the revised direction is a simplification, not a multi-stage migration.

The remaining work is a bounded single-repo cleanup: remove legacy support, remove now-unwanted runtime overrides, align docs and tests, and verify the one steady-state contract. A new planned execution graph would add ceremony without reducing meaningful risk.

## Blast Radius

- `go.mod`
- `cmd/**`
- `internal/config/**`
- `internal/resolve/**`
- `internal/adapter/opencode/**`
- `internal/sync/**`
- `SPEC.md`
- `ARCH.md`
- tests and fixtures
- examples and sample config files

## Failure Modes To Guard Against

- The binary is renamed, but the tool still reads or writes `aw.yaml`.
- The rename updates docs but leaves the Go module or internal imports inconsistent.
- Legacy `.aw/state` or `aw` markers remain partially supported and silently widen the surface back to dual-mode behavior.
- Runtime GitHub override env vars remain documented or wired even though the product no longer intends to support them.
- Removing legacy code accidentally changes target-owned output paths such as `.opencode/` or `.agents/`.
- Product rename work gets entangled with the separate OpenSpec workflow migration and leaves both directions partially applied.
- The slice grows new CLI flags or execution behavior under the cover of the rename.

## Boundary Concerns To Preserve During Execution

- This slice changes the product contract of the tool itself; it does not change the fact that the tool is a declarative workspace sync tool rather than a workflow runtime.
- Adapter code must not become the home for product rename side effects that belong in CLI, config, resolve, or sync.
- Resolve code must not absorb target file layout concerns.
- Sync code must stay ownership-safe for the current `agentspec` namespace without carrying legacy adoption policy.
- Repo-local workflow branding such as OpenSpec remains separate from the product rename.

## Backward-Compatibility Concerns To Preserve During Execution

- There are intentionally no backward-compatibility guarantees in this revision.
- Execution should remove legacy behavior cleanly rather than leaving partial fallback paths behind.
- Contributors should not need to guess whether `aw` and `agentspec` are both supported names after the rename lands.

## Next Transition

Wait for approval.

If approved, proceed directly to `/build` because the `Execution Path` is `direct-build`.
