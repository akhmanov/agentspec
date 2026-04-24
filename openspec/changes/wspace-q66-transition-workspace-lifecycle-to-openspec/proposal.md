## Traceability

- Bead: `wspace-q66`
- Change: `wspace-q66-transition-workspace-lifecycle-to-openspec`

## Why

The repository is replacing a bespoke `.opencode` lifecycle shell with `OpenSpec + Beads`. The goal is to standardize change artifacts on OpenSpec while preserving `beads` as the continuity and task-state layer.

## What Changes

- Add a checked-in OpenSpec project foundation for this repository.
- Replace the old custom OpenCode lifecycle surface with `opsx-*` commands and OpenSpec skills.
- Keep `beads` authoritative for task identity, status, assignee, priority, dependencies, and continuity.
- Add an explicit bead-to-change mapping contract and a surfaced legacy path for pre-cutover beads.
- Reframe `docs/specs/` and `docs/plans/` as legacy workflow history for new work.

## Capabilities

### New Capabilities
- `repo-workflow`: Defines the repo-local `OpenSpec + Beads` operating model, the bead-mapped OpenCode workflow surface, and the legacy-work cutover behavior.

### Modified Capabilities

## Impact

- `.opencode/`
- `openspec/`
- `docs/specs/`
- `docs/plans/`
- bead metadata and notes
