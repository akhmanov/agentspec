# Plan: Transition Workspace Lifecycle to OpenSpec While Preserving Beads

- Bead: `wspace-q66`
- Status: Approved
- Revision: 2
- Supersedes: 1
- Spec: `docs/specs/wspace-q66-transition-workspace-lifecycle-to-openspec-preserve-beads.md`
- Topology: `single-repo`
- Execution repo: `aw`
- Repo entry: isolated `wspace-q66` worktree

## Revision Reason

Approved plan revision 1 got the migration substantially in place, but pre-integration review found two remaining execution gaps:

1. The new bead-aware propose flow does not operationalize the repo instruction that newly created or ready-picked beads must be claimed before use.
2. The repo still lacks an explicit checked-in regeneration path for the generated OpenSpec/OpenCode surface and the repo-local supplements layered on top of it.

These are execution and handoff defects, not spec defects. This revision narrows the remaining work to the smallest repair needed before returning to `/review` and `/ship`.

## Plan Validation

- The latest spec revision is approved.
- The spec declares `Execution Path: planned`.
- The spec remains executable as written.
- The repair stays inside the approved scope:
  - OpenSpec remains the primary artifact workflow.
  - `beads` remains the task-state and continuity layer.
  - `docs/specs/` and `docs/plans/` remain legacy history for new work.
- The current gaps do not require revising the chosen direction, topology, or trust boundaries in the spec.

## Execution Topology

This remains `single-repo` work in `aw`.

- The repair is bounded to repo-local workflow and handoff surfaces.
- No additional repo, worktree family, or product-runtime boundary is introduced.
- The owning surfaces for this repair are:
  - `.opencode/commands/`
  - `.opencode/skills/`
  - `.opencode/instructions/INSTRUCTIONS.md`
  - `openspec/OPERATING_MODEL.md`
  - `openspec/config.yaml`
  - one checked-in regeneration artifact such as a script, doc, or both
  - verification evidence for the repaired workflow

## Dependency Graph

Execution dependency order:

1. Slice A repairs bead-claim semantics in the propose path.
2. Slice B adds an explicit checked-in regeneration and update path for the OpenSpec/OpenCode surface.
3. Slice C re-verifies the runtime workflow and closes the handoff gaps before `/review`.

Default execution should stay sequential. The repair is small, but both slices affect the same repo-local workflow contract and should not diverge.

## Slice A: Bead Claim Semantics Repair

Goal:
Make the new propose flow consistent with the repo's bead ownership and continuity rules.

Scope:

- Update `.opencode/commands/opsx-propose.md` so it distinguishes these cases explicitly:
  - if a bead is discovered via `bd ready`, claim it before proceeding
  - if a bead is created via `bd q`, capture the id and claim it before proceeding
  - if a bead id is provided explicitly by the user, do not silently change ownership; instead respect the supplied context and stop if ownership or intent is unclear
- Update `.opencode/skills/openspec-propose/SKILL.md` to match the command behavior.
- Keep the existing bead-to-change mapping rule unchanged.
- Preserve the current rule that OpenSpec must not replace bead identity or status.

Outputs:

- One coherent propose surface whose command text, skill text, and repo instructions agree on claim behavior.

Verification checkpoint:

- The propose command and propose skill both encode claim behavior for ready-picked and newly created beads.
- The repo instructions no longer contradict the propose path.
- The repair does not introduce silent claim behavior for explicitly user-provided beads.
- The bead mapping step still uses `openspec.change` and bead-prefixed change ids.

Exit criteria:

- A contributor can follow the propose instructions without leaving a newly selected or newly created bead unclaimed by accident.

## Slice B: Checked-In Regeneration Path

Goal:
Make the OpenSpec/OpenCode surface reproducible and maintainable without hidden session knowledge or personal global setup.

Scope:

- Add one explicit checked-in regeneration path for the OpenSpec/OpenCode surface.
- The path may be implemented as a small script, a small doc plus script, or another equally explicit checked-in mechanism, but it must be runnable from the repo without relying on undocumented memory.
- The regeneration path must cover:
  - re-running the upstream OpenSpec generation for OpenCode
  - preserving or reapplying repo-local supplements and overrides required by this repo contract
  - identifying which surfaces are generated, which are patched, and which are repo-local supplements
- At minimum, the checked-in path must account for the current repo-local supplement set:
  - `opsx-continue`
  - `openspec-continue-change`
  - bead-aware patches to the generated `opsx-explore`, `opsx-propose`, `opsx-apply`, and `opsx-archive` command and skill surfaces

Outputs:

- One checked-in regeneration and update path that future contributors can run and audit.
- One explicit repo-local description of which workflow artifacts are generated versus maintained as supplements or patches.

Verification checkpoint:

- The regeneration path is checked into the repo and discoverable from repo-local docs or workflow instructions.
- The path does not depend on personal global OpenSpec configuration or undocumented manual memory.
- Running the checked-in path reproduces the expected `opsx-*` and `openspec-*` surface for this repo contract.
- The repaired surface still matches the `OpenSpec + Beads` operating model after regeneration.

Exit criteria:

- Future updates to the OpenSpec/OpenCode surface have a supported repo-local maintenance path instead of relying on session-specific knowledge.

## Slice C: Reverification And Review Readiness

Goal:
Close the repair with fresh evidence that both review findings are resolved in the live repo surface.

Scope:

- Re-run the targeted workflow checks for the repaired propose flow.
- Re-run the targeted workflow checks for the checked-in regeneration path.
- Reconfirm the runtime OpenSpec surface still works with the mapped change for `wspace-q66`.
- Reconfirm that the repo did not accidentally reintroduce retired lifecycle commands or hidden dependencies.

Outputs:

- Fresh evidence that the two review findings are closed.
- A revised handoff state ready for another `/review` pass.

Verification checkpoint:

- The repaired propose path, skill, and instructions are consistent.
- The regeneration path is checked in and executable from the repo.
- `openspec validate <change-name> --type change` still succeeds for the mapped `wspace-q66` change.
- `openspec status --change <change-name> --json` still reflects a coherent mapped change.
- `go test ./...` still passes as a repo guardrail.
- No retired lifecycle shell is reintroduced in `.opencode/opencode.json` or `.opencode/commands/`.

Exit criteria:

- The change is ready to return to `/review` with fresh evidence focused on the repaired gaps.

## Cross-Slice Constraints

- Do not reopen the product decision or reintroduce the retired wspace lifecycle shell.
- Do not weaken the authority split between OpenSpec artifacts and `beads` state.
- Do not add a third canonical memory or task-state layer.
- Do not reintroduce hidden routing or workflow dependencies into the new workflow contract.
- Do not rely on undocumented session knowledge for surface maintenance.
- Do not silently change ownership of explicitly user-provided beads.

## Verification Strategy

Verification for this revision should focus on workflow integrity and maintainability.

Verification layers for this revision:

- prompt and skill consistency checks for the repaired bead-claim semantics
- execution of the checked-in regeneration path from the repo
- OpenSpec validation for the mapped `wspace-q66` change
- OpenSpec status inspection for the mapped `wspace-q66` change
- repo-level guardrail test run with `go test ./...`
- targeted diff or directory inspection showing that retired lifecycle commands were not reintroduced

## Expected Integration Shape For `/ship`

Repo mapping:

- `aw`: repaired bead-aware propose flow, explicit regeneration path, OpenSpec/OpenCode workflow surface, retained domain guidance, mapped `wspace-q66` change, and fresh verification evidence

Expected merged-state shape:

- `docs/specs/wspace-q66-transition-workspace-lifecycle-to-openspec-preserve-beads.md` remains the approved canonical spec.
- `docs/plans/wspace-q66-transition-workspace-lifecycle-to-openspec-preserve-beads.md` becomes approved revision 2 and supersedes revision 1.
- The propose command and propose skill operationalize bead claim semantics in a way that matches repo instructions.
- A checked-in regeneration path exists for the OpenSpec/OpenCode surface and its repo-local supplements.
- `.opencode/opencode.json` still exposes only the new OpenSpec-driven surface.
- `openspec/OPERATING_MODEL.md` and related repo docs still describe `OpenSpec + Beads`, not a reintroduced lifecycle shell.
- Fresh verification evidence demonstrates the repaired workflow is coherent and maintainable.

Closure conditions:

- The work returns to `/review` and clears the prior findings.
- The branch remains within the original spec scope and is ready for `/ship` once the repaired revision is reviewed cleanly.

## Bounce Conditions

Return to `/spec` before `/build` if either of these becomes true during execution:

- Closing the claim gap requires changing the approved ownership semantics for explicitly user-provided beads.
- A checked-in regeneration path cannot be made reproducible without introducing a new overlay system, new hidden runtime dependency, or another material architectural layer not covered by the approved spec.

## Next Transition

Wait for approval.

If approved, proceed to `/build` using this revision as the execution source.
