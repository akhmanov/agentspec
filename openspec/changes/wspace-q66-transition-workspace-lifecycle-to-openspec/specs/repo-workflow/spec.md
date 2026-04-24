## ADDED Requirements

### Requirement: Repo-local workflow uses OpenSpec for artifacts and beads for task state

The repository SHALL use OpenSpec as the canonical source for new change artifacts while keeping `beads` as the canonical source for task identity, status, assignee, priority, dependencies, and continuity.

#### Scenario: Starting new work from the new workflow surface

- **WHEN** a contributor starts a new repo-local change after the cutover
- **THEN** they use a bead-mapped OpenSpec change name prefixed with the bead id
- **AND** the bead stores the mapped change id in metadata under `openspec.change`

### Requirement: Active pre-cutover beads remain on an explicit legacy path until bridged

The repository SHALL provide an explicit legacy workflow path for active pre-cutover beads until an OpenSpec bridge is recorded.

#### Scenario: Continuing an active legacy bead

- **WHEN** an active bead predates the OpenSpec cutover and has no recorded OpenSpec bridge
- **THEN** contributors use the documented legacy workflow path
- **AND** the bead is not silently migrated into OpenSpec

### Requirement: OpenCode exposes a coherent bead-aware OpenSpec command surface

The repository SHALL expose an OpenCode command surface that is coherent with the `OpenSpec + Beads` operating model.

#### Scenario: Continuing a mapped change that is not apply-ready

- **WHEN** a contributor attempts to apply a mapped change whose required artifacts are still missing
- **THEN** the workflow points them to a valid continuation path for creating the missing artifacts
- **AND** it does not direct them to a missing command or skill
