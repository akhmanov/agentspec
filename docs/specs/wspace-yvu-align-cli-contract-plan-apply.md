# Spec: Align Canonical CLI Contract to Plan and Apply

- Bead: `wspace-yvu`
- Status: Approved
- Revision: 1
- Supersedes: none
- Topology: `single-repo`
- Target repo: `aw`
- Execution Path: `direct-build`

## Chosen Direction

Update the canonical product contract so `aw` publicly exposes `init`, `plan`, and `apply`, with `plan` as the read-only preview step and `apply` as the only writing command.

This slice exists to realign the repo's top-level product description with the already approved and implemented OpenCode workflow instead of reviving `sync` as a second public contract.

This slice will:

- replace `aw sync --opencode` language in canonical docs with `aw plan --opencode` and `aw apply --opencode`
- define `plan` as preview-only and `apply` as the command that writes managed changes
- remove or reframe `sync`-based examples so the repo no longer documents a stale public CLI
- keep the current implementation direction where `sync` is not preserved as a compatibility alias
- defer `example/` smoke-workspace work until after the canonical contract is aligned

## Why This Direction

This direction is worth doing because the repo currently has two conflicting stories: the root `SPEC.md` still describes `sync`, while the approved implementation slice and shipped CLI use `plan` and `apply`.

That mismatch makes it hard to answer the basic product question of "what command should a user run?" The right next step is to collapse the repo back to one clear public contract before adding broader smoke coverage or new examples.

## Remaining Assumptions

- The current `init` plus `plan` plus `apply` CLI is the intended v1 contract to standardize, not a temporary implementation detour.
- `plan` remains a live preview computed from current config, current sources, and current workspace state; it does not write a reusable plan artifact.
- `apply` remains the only command that writes files, updates managed instruction sections, prunes owned orphans, and persists state.
- If canonical docs still mention unsupported target parity, this slice may reframe that language as deferred rather than current behavior.
- The reproducible `example/` workspace should be handled in a follow-up slice once the canonical CLI contract is no longer ambiguous.

## Scope

1. Update root product documentation to describe `init`, `plan`, and `apply` instead of `sync`.
2. Align command examples and core user-flow text with the preview-then-apply model.
3. Remove or rewrite stale wording that implies `sync` is still the public entrypoint.
4. Tighten canonical wording around `plan` being read-only and `apply` being the sole writer.
5. Reconcile any product-level wording that would otherwise imply unsupported command or target behavior.

## Not Doing

- No new CLI implementation work.
- No reintroduction of `sync` as an alias or compatibility shim.
- No `example/` smoke workspace in this slice.
- No new target implementation such as Claude Code support.
- No new selector support such as GitLab.
- No ownership-model, state-model, or adapter-behavior redesign.

## Success Criteria

- `SPEC.md` no longer instructs users to run `aw sync`.
- Canonical docs clearly describe `aw plan --opencode` as preview-only.
- Canonical docs clearly describe `aw apply --opencode` as the only writing command.
- The repo no longer presents `sync` and `plan/apply` as competing public workflows.
- Product-level wording does not promise unsupported parity that the current implementation does not provide.
- Verification confirms the current CLI still exposes `init`, `plan`, and `apply` and the Go test suite remains green.

## Architecture Impact

- No intended architecture change to the Go implementation.
- The main impact is restoring one canonical contract across product docs, approved workspace specs, and the shipped CLI.
- This slice reduces product confusion without widening the product boundary or changing internal ownership rules.

## Execution Path

`direct-build`

This work should use a direct-build path because it is a narrow, single-repo alignment slice.

The direction is already chosen, the mismatch is already identified, and execution is mostly bounded to canonical docs plus lightweight verification. There is no need for a separate execution plan before editing and verifying the docs.

## Blast Radius

- `docs/specs/wspace-yvu-align-cli-contract-plan-apply.md`
- `SPEC.md`
- any nearby canonical docs that still describe `sync` as the public CLI, if needed for consistency
- verification commands and their outputs

## Failure Modes To Guard Against

- Root docs still mention `sync` after the slice is complete.
- Docs imply that `plan` writes files or state.
- Docs imply that `apply` is just a replay of a saved plan artifact when it is actually a fresh recomputation.
- Docs keep promising unsupported command or target behavior and therefore remain misleading.
- Execution drifts into `example/` design or other product-scope expansion before the core contract is aligned.

## Boundary Concerns To Preserve During Execution

- `aw` remains a declarative workspace sync tool, not a workflow runtime.
- The preview/apply split remains a transport and safety contract, not an invitation to add persisted plan files or orchestration features.
- Ownership boundaries stay unchanged: only `aw`-owned files and managed instruction blocks may be written or deleted.
- Repo-local approved specs for implementation slices remain valid; this slice aligns the product-level contract above them.

## Backward-Compatibility Concerns To Preserve During Execution

- Do not add compatibility code or a documented `sync` alias unless a concrete external need appears.
- Keep the current `plan` and `apply` semantics stable while updating docs around them.
- Avoid wording that would force future Claude or GitLab work to preserve promises the implementation has never actually shipped.

## Next Transition

Wait for approval.

If approved, proceed to `/build` because the `Execution Path` is `direct-build`.
