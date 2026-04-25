## ADDED Requirements

### Requirement: Directory-backed GitHub skills remain usable on large public repositories

The system SHALL resolve a valid directory-backed GitHub skill declared with `repo`, `ref`, and a directory `path` even when unrelated repository contents would make whole-archive download impractical.

#### Scenario: Valid skill bundle in a large public repository
- **WHEN** a user declares a GitHub skill whose `path` points at a valid bundle directory in a public repository that also contains unrelated large assets
- **THEN** `agentspec plan --opencode` resolves and previews the selected skill bundle
- **AND** `agentspec apply --opencode` materializes only the declared skill bundle under `.agents/skills/<id>/...`

### Requirement: GitHub bundle imports preserve skill validation and path safety

The system SHALL apply the same root `SKILL.md`, frontmatter, and bundle-path safety checks to directory-backed GitHub skills as it does to local bundled skills.

#### Scenario: Missing root skill file in a GitHub bundle
- **WHEN** the declared GitHub directory does not contain `SKILL.md` at the bundle root
- **THEN** resolution fails with a contextual error
- **AND** no managed files are previewed or written for that skill

#### Scenario: Unsafe path encountered in a GitHub bundle
- **WHEN** the resolved GitHub bundle contains an unsafe relative path
- **THEN** resolution fails with a bundle-path safety error
- **AND** no managed files are previewed or written for that skill
