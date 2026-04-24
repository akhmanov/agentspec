## Traceability

- Bead: `wspace-q66`
- Change: `wspace-q66-transition-workspace-lifecycle-to-openspec`
- Task state authority: `beads`

Tasks in this file are an execution checklist. Ownership, status, assignee, priority, and dependencies live in `beads`.

## 1. Contract And Foundation

- [x] 1.1 Define the `OpenSpec + Beads` authority model and cutover rules in repo-local docs.
- [x] 1.2 Initialize OpenSpec for OpenCode and fork a project-local schema for the repo contract.

## 2. Surface Flip

- [x] 2.1 Replace the custom `.opencode` lifecycle surface with the `opsx-*` workflow surface.
- [x] 2.2 Keep only the retained domain skills and add the repo-local `opsx-continue` supplement needed for blocked apply flows.

## 3. Legacy And Mapping Proof

- [x] 3.1 Mark `docs/specs/` and `docs/plans/` as legacy history for new work.
- [x] 3.2 Record the bead-to-change mapping and active-bead cutover notes.

## 4. Verification

- [x] 4.1 Verify the customized schema resolves and validates.
- [x] 4.2 Verify the mapped OpenSpec change and Go test suite run successfully in the worktree.
