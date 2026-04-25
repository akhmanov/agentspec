## ADDED Requirements

### Requirement: Repository includes a deterministic local smoke example

The repository SHALL include a committed local example workspace that demonstrates supported `agentspec plan` and `agentspec apply` behavior without requiring network access.

#### Scenario: Contributor exercises the local smoke example
- **WHEN** a contributor opens `example/local-smoke`
- **THEN** the example contains all files needed to run the documented local smoke commands
- **AND** those commands exercise local resource handling without live network dependencies

### Requirement: Repository includes a live GitHub smoke example

The repository SHALL include a committed GitHub-backed example workspace that references pinned public repositories.

#### Scenario: Contributor exercises the GitHub smoke example
- **WHEN** a contributor opens `example/github-smoke`
- **THEN** the example contains a documented `agentspec.yaml` that references live public GitHub sources using pinned refs
- **AND** the example demonstrates at least one GitHub-backed file resource and one directory-backed skill resource

### Requirement: Example docs describe stability expectations

The repository SHALL document the different stability characteristics of the local and live smoke examples.

#### Scenario: Contributor reads example documentation
- **WHEN** a contributor reads the example README files
- **THEN** the local example is described as deterministic
- **AND** the GitHub example is described as depending on public upstream availability
