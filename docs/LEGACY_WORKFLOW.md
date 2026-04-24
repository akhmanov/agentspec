# Legacy Workflow

This document exists for active beads that started before the `OpenSpec + Beads` cutover.

Use this path only for pre-cutover work that has not been bridged into OpenSpec yet.

## When To Use It

- the bead already started on the old `docs/specs/` + `docs/plans/` workflow
- no explicit bridge to an OpenSpec change has been recorded
- the operating model says to finish the work where it started

## How To Work Legacy Beads

1. Recover the active bead with `bd list --status=in_progress`.
2. Read the existing canonical spec in `docs/specs/<bead-id>-<slug>.md`.
3. If a planned execution artifact exists, read `docs/plans/<bead-id>-<slug>.md`.
4. Make progress directly against those legacy artifacts and the bead state.
5. Do not silently create or attach an OpenSpec change unless the bridge is recorded explicitly.

## Bridge Rule

If a legacy bead must move to OpenSpec:

- record the bridge in the bead metadata and notes
- create the mapped OpenSpec change with the bead id prefix
- note the bridge in the OpenSpec change artifacts

Until that happens, the bead stays on the legacy path.
