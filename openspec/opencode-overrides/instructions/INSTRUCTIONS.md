# OpenSpec + Beads Instructions

This repository uses `OpenSpec + Beads`.

## Boundaries

- OpenSpec is the primary repo-local workflow for change artifacts.
- `beads` is the primary repo-local workflow for task identity, status, assignee, priority, dependencies, and continuity.
- `aw` remains a declarative workspace sync tool, not a workflow runtime.
- `docs/specs/` and `docs/plans/` remain readable legacy history, not the canonical source for new work.
- Use `.opencode/skills/baseline.toml` as the curated repo skill baseline. Generated OpenSpec skills live beside retained repo-local domain skills.

## Operating Model

1. Recover active work with `bd list --status=in_progress` before starting new work.
2. If the user did not provide a bead id, use `bd ready` or create a new bead with `bd q` and claim it.
3. Each OpenSpec change maps to exactly one bead.
4. OpenSpec change names must be prefixed with the bead id.
5. Bead metadata should record the mapped change id under `openspec.change`.
6. Use the generated OpenSpec command surface for new work:
   - `/opsx-explore`
   - `/opsx-propose`
   - `/opsx-continue`
   - `/opsx-apply`
   - `/opsx-archive`
7. Run explicit validation such as `openspec validate <change-name> --type change` before review or integration.
8. Finish active pre-cutover work under its existing workflow unless a bridge is recorded explicitly in both the bead and the mapped OpenSpec change.
9. Use `docs/LEGACY_WORKFLOW.md` for any active bead that still lives on the pre-cutover workflow.
10. Regenerate or update the OpenCode workflow surface with `./openspec/regenerate-opencode-surface.sh`; see `openspec/REGENERATE_OPENCODE.md`.

## Retained Guidance

- Keep using retained repo-local domain skills such as `backend-architecture`, `api-contracts`, `security-review`, `infra-devops`, and `gitops-review` when they apply.
- Do not recreate the retired `/explore` `/debate` `/spec` `/plan` `/build` `/test` `/review` `/ship` lifecycle under different names.
