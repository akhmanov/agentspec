## ADDED Requirements

### Requirement: AGENTS briefing explains product purpose

The repository briefing in `AGENTS.md` SHALL explain why `agentspec` exists, which workspace-maintenance problem it solves, and that `agentspec.yaml` is the single source of truth for the managed workspace surface.

#### Scenario: Briefing explains the problem being solved
- **WHEN** an agent reads the product introduction in `AGENTS.md`
- **THEN** it can identify that `agentspec` exists to reduce repeated manual workspace setup and prevent workspace drift across repositories

#### Scenario: Briefing explains the source of truth
- **WHEN** an agent reads the product purpose section in `AGENTS.md`
- **THEN** it sees that `agentspec.yaml` defines the desired managed workspace surface

### Requirement: AGENTS briefing defines resource types by workspace effect

`AGENTS.md` SHALL define `sections`, `commands`, `agents`, and `skills` in terms of what each resource materializes into for the current target workspace, not only by repeating the YAML key names.

#### Scenario: Briefing explains section materialization
- **WHEN** an agent reads the resource model section in `AGENTS.md`
- **THEN** it can understand that `sections` are managed instruction fragments inserted into `AGENTS.md`

#### Scenario: Briefing explains file-based resources
- **WHEN** an agent reads the resource model section in `AGENTS.md`
- **THEN** it can understand that `commands` and `agents` render to target documents and that `skills` are directory-based bundles rather than single files

### Requirement: AGENTS briefing explains sync semantics and ownership boundaries

`AGENTS.md` SHALL describe `agentspec plan --opencode` and `agentspec apply --opencode` as managed desired-state sync operations with explicit ownership boundaries.

#### Scenario: Briefing explains plan behavior
- **WHEN** an agent reads the sync model in `AGENTS.md`
- **THEN** it sees that `plan --opencode` previews managed create, update, delete, and conflict results without writing workspace files

#### Scenario: Briefing explains apply safety
- **WHEN** an agent reads the sync model in `AGENTS.md`
- **THEN** it sees that `apply --opencode` materializes managed state while refusing to silently overwrite foreign content
