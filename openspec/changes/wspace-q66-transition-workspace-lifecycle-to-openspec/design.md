## Traceability

- Bead: `wspace-q66`
- Change: `wspace-q66-transition-workspace-lifecycle-to-openspec`

## Context

The repo previously exposed a custom `.opencode` command shell and workspace-specific skills. The approved migration moves the artifact workflow to OpenSpec while retaining `beads` for operational continuity.

## Goals / Non-Goals

**Goals:**
- make OpenSpec the primary repo-local artifact workflow
- keep `beads` as the task-state system
- replace the old `.opencode` lifecycle surface with a coherent `opsx-*` surface
- keep legacy pre-cutover beads workable without silent migration

**Non-Goals:**
- changing the `aw` product contract
- adding a third canonical memory layer
- preserving the old lifecycle shell under new names

## Decisions

- Use a project-local OpenSpec schema `aw-openspec-beads` so artifact templates carry bead traceability and make `tasks.md` explicitly non-authoritative for task state.
- Keep the generated OpenSpec core surface and add one repo-local `opsx-continue` supplement because blocked apply flows require artifact continuation and the generated core profile does not provide it.
- Keep `openspec.change` in bead metadata as the canonical mapping from bead to OpenSpec change.
- Surface a legacy-work doc and explicit bead notes for active pre-cutover work instead of silently bridging those beads.

## Risks / Trade-offs

- Generated OpenSpec surface plus repo-local supplements can drift if future updates overwrite local expectations. Mitigation: keep the repo-local contract documented in `openspec/OPERATING_MODEL.md` and validate the live command surface after changes.
- Legacy and new workflow surfaces can become ambiguous. Mitigation: keep legacy usage documented only in `docs/LEGACY_WORKFLOW.md` and remove the retired lifecycle command shell.

## Migration Plan

- establish the operating model and mapping contract
- initialize OpenSpec and customize schema/templates
- flip the OpenCode-visible workflow surface
- mark legacy docs as history and record active-bead cutover notes

## Open Questions

- Whether this repo will eventually want generated verify support or continue relying on explicit `openspec validate <change> --type change` plus repo-local review flow.
