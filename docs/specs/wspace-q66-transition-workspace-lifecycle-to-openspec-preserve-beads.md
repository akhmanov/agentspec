# Spec: Transition Workspace Lifecycle to OpenSpec While Preserving Beads

- Bead: `wspace-q66`
- Status: Approved
- Revision: 1
- Supersedes: none
- Topology: `single-repo`
- Target repo: `aw`
- Execution Path: `planned`

## Chosen Direction

Adopt OpenSpec as the primary repo-local change and specification workflow for `aw`, and retire the current custom wspace lifecycle shell in `.opencode`.

This slice will:

- replace the current lifecycle command and skill surface (`/explore`, `/debate`, `/spec`, `/plan`, `/build`, `/test`, `/review`, `/ship`) with the OpenSpec workflow surface for OpenCode
- make `openspec/changes/` and `openspec/specs/` the canonical artifact source for future workflow changes
- keep `beads` as the source of truth for task identity, status, and continuity
- define one explicit mapping between a bead and an OpenSpec change so the repo does not drift into split state
- retire `docs/specs/` and `docs/plans/` from forward-looking canonical use while preserving existing files as legacy history unless execution proves targeted migration is needed
- retain repo-local supplemental guidance only where it adds value without recreating the retired lifecycle under another name

## Why This Direction

This direction is worth doing because the repo owner has made an explicit decision to standardize on OpenSpec instead of continuing to evolve the current bespoke `.opencode` lifecycle.

The right way to honor that decision is a real cutover, not a half-migration. OpenSpec should become the primary workflow model for new work, while `beads` stays in place as the continuity and task-state layer that OpenSpec does not natively provide.

This preserves one important repo-specific capability while still collapsing the workflow shell onto the chosen ecosystem standard.

## Remaining Assumptions

- Full transition means OpenSpec becomes the primary repo-local workflow, not an optional sidecar.
- `beads` remains the source of truth for task identity and status; OpenSpec does not replace `bd`.
- Each OpenSpec change can be mapped to exactly one bead through naming, metadata, or another explicit repo-local rule chosen during planning.
- Existing historical files in `docs/specs/` and `docs/plans/` may remain as frozen legacy records instead of being fully converted.
- The currently active bead `wspace-yvu` will either finish under the current workflow or be bridged explicitly during migration; there must be no silent mid-flight split-brain period.
- Repo-local non-lifecycle skills may survive only if they remain compatible with OpenSpec and do not reintroduce the retired workspace lifecycle by another path.

## Scope

1. Define the target OpenSpec layout and repo-local configuration for this repository.
2. Replace the current `.opencode` lifecycle command surface with OpenSpec-generated command and skill artifacts for OpenCode.
3. Update repo instructions so OpenSpec is the primary workflow UX and `beads` is explicitly retained as the task-state layer.
4. Define one canonical bead-to-OpenSpec mapping model for future work.
5. Reframe `docs/specs/` and `docs/plans/` as legacy workflow history rather than the canonical source for new work.
6. Define cutover rules for active and in-flight work, including how `wspace-yvu` is handled.
7. Verify the resulting workflow end to end inside this repository.

## Not Doing

- No change to the `aw` product contract or CLI behavior as an end-user feature of the tool itself.
- No attempt to make OpenSpec natively understand `beads`; their integration remains repo-local policy.
- No full historical backfill of every legacy spec and plan into OpenSpec unless execution proves a targeted migration is necessary.
- No preservation of the current wspace phase model as a second equal workflow after cutover.
- No multi-repo rollout beyond this repository in this slice.

## Success Criteria

- OpenSpec commands and skills are the primary repo-local workflow entrypoints in OpenCode for new work.
- The current `.opencode` lifecycle commands and workspace-lifecycle skills are removed or clearly retired from primary use.
- `beads` remains documented and operational as the task and continuity layer.
- A contributor can identify one unambiguous mapping between a bead and an OpenSpec change.
- New canonical workflow artifacts are created under `openspec/`, not newly added to `docs/specs/` or `docs/plans/`.
- The repo makes the status of legacy artifacts explicit so contributors can tell which source of truth applies.
- Verification shows a contributor can start, progress, and finish a repo-local change using the new workflow without relying on the retired wspace command set.

## Architecture Impact

- This is a repo-process and control-plane migration, not a product-runtime change.
- `.opencode` shifts from a custom lifecycle shell to an OpenSpec-installed command and skill surface plus any retained repo-local supplemental guidance.
- `openspec/` becomes the canonical home for ongoing change and specification artifacts.
- `bd` remains a separate source of truth for issue state, so the repo must define a clear interface between bead identity and OpenSpec change identity.
- Current workspace instructions that treat `docs/specs/` and `docs/plans/` as canonical will be intentionally superseded by the new model.

## Execution Path

`planned`

This work should use a planned path because it is a repo-wide process migration with several coupled surfaces: command UX, instructions, artifact authority, legacy history policy, bead integration, and cutover verification.

The implementation order matters. The migration should minimize the time spent in a dual-source-of-truth state and must define explicit bridge rules before any broad cutover.

## Blast Radius

- `.opencode/opencode.json`
- `.opencode/commands/`
- `.opencode/skills/`
- `.opencode/agents/`
- `.opencode/instructions/INSTRUCTIONS.md`
- `openspec/`
- `docs/specs/`
- `docs/plans/`
- repo-local onboarding and workflow documentation that still describes the retired lifecycle
- bead metadata, notes, or linking conventions needed to preserve continuity
- verification evidence for the new workflow contract

## Failure Modes To Guard Against

- Two equally canonical artifact sources remain active after cutover.
- OpenSpec change identity drifts from bead identity or continuity state.
- Contributors use the new OpenSpec surface while repo instructions still direct them to retired wspace phases.
- Active work is cut over midstream without an explicit bridge and becomes unclear or stranded.
- OpenSpec-generated artifacts overwrite or obscure repo-local guidance that still matters for safe work.
- Retained legacy commands or skills conflict with OpenSpec-generated ones and leave the UX ambiguous.
- The migration gets interpreted as productizing OpenSpec into `aw` instead of changing this repository's local workflow.
- Verification proves only file generation, not real workflow usability.

## Boundary Concerns To Preserve During Execution

- `aw` remains a declarative workspace sync tool, not a workflow runtime.
- `beads` remains the continuity and task-state system unless a future spec explicitly removes it.
- OpenSpec becomes the repo-local change and spec workflow, not the product contract of `aw` itself.
- New work must have one authoritative artifact model after cutover.
- Repo-local supplemental skills may exist, but they must not recreate the retired lifecycle under another name.

## Backward-Compatibility Concerns To Preserve During Execution

- Existing historical docs remain available and understandable after cutover even if they are no longer canonical for new work.
- Active beads are not stranded without a documented migration or completion path.
- Contributors should not need to infer whether the old or new workflow applies to a given change.
- OpenSpec integration should rely on repo-local checked-in configuration where needed rather than undocumented personal global state.

## Next Transition

Wait for approval.

If approved, proceed to `/plan` because the `Execution Path` is `planned`.
