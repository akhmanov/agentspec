# cli-context-and-target-surfaces Specification

## Purpose
TBD - created by archiving change polish-cli-context-and-target-surfaces. Update Purpose after archive.
## Requirements
### Requirement: CLI commands support explicit execution context
`agentspec` SHALL support explicit workspace and config selection through long-form execution-context inputs.

The effective workspace root SHALL be determined by this precedence order:

1. `--root`
2. `AGENTSPEC_ROOT`
3. the current working directory

The effective config path SHALL be determined by this precedence order:

1. `--config`
2. `AGENTSPEC_CONFIG`
3. `<effective-root>/agentspec.yaml`

Relative `--config` values SHALL resolve relative to the effective root when `--root` is provided, otherwise relative to the current working directory.

#### Scenario: Flags override environment and defaults
- **WHEN** the user runs `agentspec --root /work/repo --config config/dev/agentspec.yaml plan --target opencode`
- **THEN** `agentspec` treats `/work/repo` as the workspace root
- **AND** it loads the config from `/work/repo/config/dev/agentspec.yaml`

#### Scenario: Environment fills in execution context when flags are absent
- **WHEN** `AGENTSPEC_ROOT=/work/repo` and `AGENTSPEC_CONFIG=ops/agentspec.yaml` are set and the user runs `agentspec plan --target opencode`
- **THEN** `agentspec` treats `/work/repo` as the workspace root
- **AND** it loads the config from `/work/repo/ops/agentspec.yaml`

#### Scenario: Init uses the effective config path
- **WHEN** the user runs `agentspec --root /work/repo init`
- **THEN** `agentspec` writes the starter config to `/work/repo/agentspec.yaml`

### Requirement: Local selector paths are resolved relative to the loaded config file
For local selectors in `agentspec.yaml`, relative `path:` values SHALL resolve from the directory that contains the loaded config file.

#### Scenario: Relative path selector uses config directory
- **WHEN** `agentspec` loads `/work/repo/config/agentspec.yaml`
- **AND** that config contains `path: ./resources/review.md`
- **THEN** `agentspec` reads `/work/repo/config/resources/review.md`

#### Scenario: Absolute path selector is preserved
- **WHEN** a resource uses an absolute `path:` selector
- **THEN** `agentspec` reads that exact absolute path instead of rebasing it to the config directory

### Requirement: Sync commands use explicit repeatable target selection
`agentspec plan` and `agentspec apply` SHALL require at least one `--target` flag. `--target` SHALL be repeatable, SHALL preserve user-specified order, and SHALL accept supported target identifiers including `opencode` and `claude-code`.

Each target SHALL keep separate planning and apply state. Multi-target apply SHALL run sequentially in target order and SHALL stop on the first target error without rolling back earlier successful targets.

#### Scenario: Plan groups multiple targets distinctly
- **WHEN** the user runs `agentspec plan --target opencode --target claude-code`
- **THEN** the output identifies which planned changes belong to `opencode`
- **AND** the output identifies which planned changes belong to `claude-code`

#### Scenario: Apply stops on the first failing target
- **WHEN** the user runs `agentspec apply --target opencode --target claude-code`
- **AND** `opencode` applies successfully
- **AND** `claude-code` encounters a target-specific conflict
- **THEN** `agentspec` reports the `claude-code` error
- **AND** it does not roll back the successful `opencode` apply

#### Scenario: Missing target is rejected
- **WHEN** the user runs `agentspec plan` or `agentspec apply` without any `--target`
- **THEN** `agentspec` fails with an error that the command requires at least one target

### Requirement: OpenCode uses a fully target-native artifact surface
When syncing for the `opencode` target, `agentspec` SHALL materialize sections into `AGENTS.md`, commands into `.opencode/commands/<id>.md`, agents into `.opencode/agents/<id>.md`, and skills into `.opencode/skills/<id>/...`.

#### Scenario: OpenCode plan reflects target-native paths
- **WHEN** the user runs `agentspec plan --target opencode` with sections, commands, agents, and skills in the desired config
- **THEN** the reported OpenCode changes refer only to `AGENTS.md`, `.opencode/commands/...`, `.opencode/agents/...`, and `.opencode/skills/...`

#### Scenario: Prior managed OpenCode skills transition from legacy path
- **WHEN** the existing `opencode` state tracks managed skill files under `.agents/skills/...`
- **AND** the user runs `agentspec plan --target opencode`
- **THEN** the plan reports deletion of the prior managed `.agents/skills/...` outputs
- **AND** it reports creation of the new `.opencode/skills/...` outputs instead of rejecting the prior state as foreign

### Requirement: Claude Code uses a fully target-native artifact surface
When syncing for the `claude-code` target, `agentspec` SHALL materialize sections into `CLAUDE.md`, commands into `.claude/commands/<id>.md`, agents into `.claude/agents/<id>.md`, and skills into `.claude/skills/<id>/...`.

#### Scenario: Claude Code plan reflects target-native paths
- **WHEN** the user runs `agentspec plan --target claude-code` with sections, commands, agents, and skills in the desired config
- **THEN** the reported Claude Code changes refer only to `CLAUDE.md`, `.claude/commands/...`, `.claude/agents/...`, and `.claude/skills/...`

#### Scenario: Claude Code state stays separate from OpenCode
- **WHEN** the user runs `agentspec apply --target opencode --target claude-code`
- **THEN** `agentspec` writes separate managed state files for the two targets
- **AND** a later `plan --target claude-code` does not depend on the `opencode` state file to compute Claude Code ownership

