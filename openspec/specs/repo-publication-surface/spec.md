# repo-publication-surface Specification

## Purpose
TBD - created by archiving change publication-polish. Update Purpose after archive.
## Requirements
### Requirement: Repository root is directly installable
The published module SHALL use the canonical path `github.com/akhmanov/agentspec`, and the repository root SHALL expose the installable `agentspec` CLI so users can run `go install github.com/akhmanov/agentspec@latest` without referencing an internal subdirectory.

#### Scenario: Root go install resolves the CLI
- **WHEN** a user runs `go install github.com/akhmanov/agentspec@latest`
- **THEN** the Go tool resolves the repository root as an installable command package
- **AND** installs a binary named `agentspec`

### Requirement: Root README onboards first-time users
A `README.md` file SHALL exist at the repository root. It SHALL explain what `agentspec` does, show the canonical `go install github.com/akhmanov/agentspec@latest` command, show how to verify the install with `agentspec --version`, provide a minimal quickstart using `init`, `plan`, and `apply`, list supported targets, and explain the managed ownership model.

#### Scenario: README shows canonical install and verify commands
- **WHEN** a user opens the repository root
- **THEN** `README.md` shows `go install github.com/akhmanov/agentspec@latest`
- **AND** `README.md` shows `agentspec --version` as the install verification step

#### Scenario: README shows minimal quickstart
- **WHEN** a user reads the getting-started section in `README.md`
- **THEN** it includes a minimal `agentspec.yaml` example
- **AND** it shows `agentspec init`, `agentspec plan --target <target>`, and `agentspec apply --target <target>`

#### Scenario: README explains ownership and supported targets
- **WHEN** a user reads the reference section in `README.md`
- **THEN** it lists `opencode` and `claude-code` as supported targets
- **AND** it explains that `plan` previews managed changes and `apply` only materializes managed outputs while refusing to silently overwrite foreign content

### Requirement: README separates onboarding from validation examples
The root README SHALL present the root install and quickstart flow as the primary onboarding path and SHALL describe committed example workspaces as smoke or validation examples rather than as the required installation path.

#### Scenario: README positions smoke examples as validation flows
- **WHEN** a user reads the examples section in `README.md`
- **THEN** it links to the committed smoke examples
- **AND** it describes them as validation or smoke flows instead of the canonical install path

### Requirement: README documents advanced GitHub bundle timeout control
The root README SHALL document `AGENTSPEC_GIT_TIMEOUT` as an advanced environment variable that overrides the timeout for internal git operations used to resolve GitHub-backed skill bundles. The documented value format SHALL use Go duration strings.

#### Scenario: README documents git timeout example
- **WHEN** a user reads the advanced environment section in `README.md`
- **THEN** it explains that `AGENTSPEC_GIT_TIMEOUT` affects internal git-based GitHub bundle resolution
- **AND** it includes an example duration such as `60s`

