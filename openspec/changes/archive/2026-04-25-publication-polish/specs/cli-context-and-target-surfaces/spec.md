## ADDED Requirements

### Requirement: CLI exposes version flags
The `agentspec` root command SHALL accept `--version` and `-v`. Each flag invocation SHALL print a single-line version string in the format `agentspec <version>` to standard output and SHALL exit successfully without requiring a subcommand. When no explicit build version is injected, the reported version SHALL be `dev`.

#### Scenario: Long version flag prints the version
- **WHEN** a user runs `agentspec --version`
- **THEN** the command prints `agentspec <version>` to standard output
- **AND** exits successfully without running a subcommand

#### Scenario: Short version flag prints the same version
- **WHEN** a user runs `agentspec -v`
- **THEN** the command prints the same `agentspec <version>` string as `--version`
- **AND** exits successfully without running a subcommand

### Requirement: CLI help explains execution-context and target-selection behavior
The root help output SHALL describe how `--root` and `--config` resolve through `AGENTSPEC_ROOT`, `AGENTSPEC_CONFIG`, and their default paths. The `plan` help output SHALL describe `--target` as repeatable, list supported target identifiers `opencode` and `claude-code`, and describe the behavior of `--verbose`. Errors for invalid target identifiers SHALL list the supported target values.

#### Scenario: Root help explains execution-context fallbacks
- **WHEN** a user runs `agentspec --help`
- **THEN** the help output mentions `AGENTSPEC_ROOT`
- **AND** the help output mentions `AGENTSPEC_CONFIG`
- **AND** the help output explains the default config path `<root>/agentspec.yaml`

#### Scenario: Plan help explains target and verbose flags
- **WHEN** a user runs `agentspec plan --help`
- **THEN** the help output states that `--target` is repeatable
- **AND** the help output lists `opencode` and `claude-code` as supported values
- **AND** the help output explains that `--verbose` expands the computed change set with per-path output and conflict reasons

#### Scenario: Invalid target error lists supported values
- **WHEN** a user runs `agentspec plan --target nope`
- **THEN** the command fails with an error identifying `nope` as unsupported
- **AND** the error lists `opencode` and `claude-code` as supported values

### Requirement: Init writes an annotated starter config
`agentspec init` SHALL write a valid starter `agentspec.yaml` that keeps the four top-level resource maps and adds comments that explain what `sections`, `commands`, `agents`, and `skills` materialize into.

#### Scenario: Starter config includes explanatory comments
- **WHEN** a user runs `agentspec init`
- **THEN** `agentspec.yaml` contains the `sections`, `commands`, `agents`, and `skills` top-level keys
- **AND** the file includes explanatory comments for each top-level resource type

## MODIFIED Requirements

### Requirement: Sync commands use explicit repeatable target selection
`agentspec plan` and `agentspec apply` SHALL require at least one `--target` flag. `--target` SHALL be repeatable, repeated target values SHALL be deduplicated while preserving first-occurrence order, and supported target identifiers SHALL include `opencode` and `claude-code`.

Each distinct target SHALL keep separate planning and apply state. Multi-target apply SHALL run sequentially in target order and SHALL stop on the first target error without rolling back earlier successful targets.

#### Scenario: Plan groups multiple distinct targets
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

#### Scenario: Duplicate targets are normalized once
- **WHEN** the user runs `agentspec plan --target opencode --target opencode --target claude-code`
- **THEN** `agentspec` plans for `opencode` first and `claude-code` second
- **AND** the output contains one group for `opencode`
- **AND** the output contains one group for `claude-code`
