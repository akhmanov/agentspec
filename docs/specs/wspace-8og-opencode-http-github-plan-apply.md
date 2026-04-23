# Spec: OpenCode HTTP and GitHub With Plan and Apply

- Bead: `wspace-8og`
- Status: Approved
- Revision: 1
- Supersedes: `docs/specs/wspace-wal-opencode-inline-path-first-slice.md`
- Topology: `single-repo`
- Target repo: `aw`
- Execution Path: `planned`

## Chosen Direction

Extend the current OpenCode implementation instead of branching into Claude support.

This slice will:

- keep OpenCode as the only target
- keep existing `inline` and `path` support working
- add `http` and `github` selectors
- replace the public `aw sync --opencode` workflow with `aw plan --opencode` and `aw apply --opencode`
- define `aw plan` as a read-only preview of managed creates, updates, deletes, and ownership conflicts
- define `aw apply` as the only command that writes files, updates managed instruction blocks, removes owned orphans, and updates state

The product stays a declarative sync tool. This spec changes the source and CLI surface, not the ownership model.

## Why This Direction

This direction is worth doing because it improves the product in the two places that matter next: reusable sources and safer execution.

`http` and `github` make `aw.yaml` useful outside one local machine by allowing shared commands, sections, agents, and skills to be referenced without copying them into each repo. `plan` and `apply` make the sync behavior easier to trust because a user can inspect the exact managed changes before any file is written.

Deferring Claude support is also the right call here. The adapter, instruction-file placement, and partial-support policy for Claude are a separate design surface. Mixing that work into remote selectors and CLI contract changes would expand the blast radius without making this slice better.

## Remaining Assumptions

- OpenCode remains the only target in scope, so `--opencode` stays explicit on `plan` and `apply`.
- The existing `sync` command surface can be removed rather than preserved behind a compatibility alias because there is no shipped compatibility requirement yet.
- `plan` is preview-only and does not persist a plan artifact for later replay.
- `apply` recomputes desired state from `aw.yaml` and current sources instead of consuming output from a prior `plan` run.
- `http` means HTTPS only and only for single-file resources, matching the product-level spec.
- `github` should support file-backed resources for `sections`, `commands`, and `agents`, plus file-or-directory semantics for `skills`.

## Scope

1. Update the CLI surface from `init` plus `sync` to `init` plus `plan` plus `apply`.
2. Implement a read-only planning flow for OpenCode that loads config, resolves resources, renders desired state, compares against the workspace, and reports managed creates, updates, deletes, and ownership conflicts.
3. Keep the existing apply flow as the only writer, while aligning it with the new `apply` command contract.
4. Extend config validation to allow `http` and `github` selectors in addition to `inline` and `path`.
5. Implement `http` resolution for single-file resources across `sections`, `commands`, `agents`, and `skills`.
6. Implement `github` resolution for file-backed resources across `sections`, `commands`, and `agents`.
7. Implement `github` file and directory resolution for `skills`, including bundled-skill import rules with required root `SKILL.md`.
8. Preserve target-neutral resolve output so adapters still consume normalized resources instead of raw selector config.
9. Add tests for the new selector validation rules, remote resolution paths, plan preview behavior, and CLI flows.
10. Keep ownership-safe apply and orphan cleanup behavior intact while reusing the same decision rules between plan and apply.

## Not Doing

- No `--claude-code` target work.
- No `gitlab` selector support in this slice.
- No plan file written to disk.
- No offline cache, lockfile, or download store for remote content in this slice.
- No support for multi-file `http` resources.
- No pack support, `aw check`, or `aw doctor`.
- No change to the schema shape for sections, commands, agents, or skills beyond enabling the approved selectors.
- No ownership expansion beyond current `aw`-owned files, managed blocks, and state rules.
- No generic provider abstraction that tries to solve GitHub, GitLab, HTTP, and future sources all at once.

## Success Criteria

- `aw plan --opencode` runs without writing files or state and reports the managed operations implied by the current config.
- `aw apply --opencode` writes the same managed changes that `plan` previews, subject to current workspace state and ownership checks.
- `inline` and `path` keep working for all currently supported resource kinds.
- `http` works for single-file `sections`, `commands`, `agents`, and `skills`.
- `github` works for file-backed `sections`, `commands`, and `agents`.
- `github` works for single-file and directory-backed `skills` and enforces root `SKILL.md` validation for bundles.
- Non-HTTPS `http` inputs are rejected clearly.
- Invalid or unsupported GitHub selector shapes are rejected clearly with resource kind and id context.
- Re-running `aw plan --opencode` after `aw apply --opencode` on an unchanged workspace reports no managed changes.
- Ownership conflicts are surfaced in `plan` and still block `apply`.
- The test suite covers remote selector validation, remote fetch failures, skill bundle handling, plan preview behavior, and the updated CLI surface.

## Architecture Impact

- `cmd/aw` changes from a write-only `sync` transport to a transport that exposes separate read-only plan and write-only apply commands.
- `internal/config` expands selector validation to admit `http` and `github` while keeping exactly-one-selector enforcement.
- `internal/resolve` grows remote fetch logic for `http` and `github` but must remain target-neutral.
- `internal/model` may need a small plan/report shape so CLI output can describe intended operations without teaching `cmd/aw` how to diff files itself.
- `internal/sync` should own the managed diff and ownership decision rules so `plan` and `apply` do not diverge.
- `internal/adapter/opencode` should continue to render only desired outputs and should not absorb remote-fetch or diff logic.

## Execution Path

`planned`

This work should use a planned path because the implementation order matters.

The CLI contract changes, remote selectors extend validation and resolution, and the new plan preview must share ownership logic with apply instead of growing as a separate approximation. Those pieces should be sequenced deliberately so the repo does not end up with duplicated diff logic, accidental writes during planning, or source-specific behavior leaking into the wrong layer.

## Blast Radius

- `.beads/`
- `docs/specs/wspace-8og-opencode-http-github-plan-apply.md`
- future `docs/plans/` artifact for this bead after approval
- `cmd/aw/**`
- `internal/config/**`
- `internal/resolve/**`
- `internal/model/**`
- `internal/adapter/opencode/**`
- `internal/sync/**`
- Go tests and remote-resolution fixtures

## Failure Modes To Guard Against

- `aw plan` writes files, updates state, or mutates managed instruction files.
- `aw plan` and `aw apply` use different ownership or diff logic and therefore disagree about what would change.
- `http` accepts insecure URLs or silently rewrites them.
- `github` resolution permits unsafe paths, ambiguous selector shapes, or bundle traversal outside the intended repo subtree.
- Remote skill bundles skip validation of the root `SKILL.md`.
- Apply overwrites a foreign file because the remote selector resolved successfully.
- Adapter code learns about remote selector parsing or fetch details.
- CLI handlers accumulate business logic for diffing, ownership checks, or fetch policy.

## Boundary Concerns To Preserve During Execution

- `aw` remains a declarative sync tool, not a workflow runtime, remote cache manager, or package installer.
- Resolver code must stay ignorant of `.opencode`, `AGENTS.md`, and marker placement.
- Adapter code must stay ignorant of raw selector parsing and remote transport details.
- Planning and applying must share one ownership policy source of truth.
- Only `aw`-owned files and explicit managed instruction blocks may be written or deleted.
- Remote selectors expand the source boundary, not the write boundary.

## Backward-Compatibility Concerns To Preserve During Execution

- Existing `inline` and `path` configs must keep working unchanged.
- Existing OpenCode output paths and marker formats should stay stable.
- The removal of `sync` is an intentional CLI contract change for this repo state and should not be hidden behind extra compatibility code unless a concrete user need appears.
- Adding `gitlab` or Claude support later should not require redesigning the resolved resource model introduced here.

## Next Transition

Wait for approval.

If approved, proceed to `/plan` because the `Execution Path` is `planned`.
