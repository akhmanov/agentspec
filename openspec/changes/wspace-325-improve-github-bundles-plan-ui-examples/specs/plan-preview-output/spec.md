## ADDED Requirements

### Requirement: Plan output provides a grouped default preview

The CLI SHALL print `agentspec plan` output as a grouped default preview instead of an unstructured stream of change lines.

#### Scenario: Default plan groups managed changes and conflicts
- **WHEN** a plan contains creates, updates, deletes, or conflicts
- **THEN** the CLI prints those items grouped by kind
- **AND** each group is compact enough for routine review

#### Scenario: Empty plan remains explicit
- **WHEN** a plan has no managed changes and no conflicts
- **THEN** the CLI prints `No managed changes.`

### Requirement: Verbose plan output expands the same change set

The CLI SHALL support `agentspec plan --verbose` as an expanded view of the same computed plan.

#### Scenario: Verbose mode adds detail without changing the plan result
- **WHEN** a user reruns the same plan command with `--verbose`
- **THEN** the CLI reports the same changes and conflicts as the default mode
- **AND** it includes additional detail such as grouped explanations or conflict context
