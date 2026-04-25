## Why

`agentspec` already behaves like a reliable small CLI, but the repository and command surface still expose too much internal structure to first-time users. Before publication, the project needs a canonical root install path, a root README, version flags, and a few CLI ergonomics so users can install, verify, and try the tool without learning repo internals.

## What Changes

- Publish a root install surface so users can install the CLI with `go install github.com/akhmanov/agentspec@latest`.
- Add a root README with installation, install verification, quickstart, supported targets, ownership semantics, and advanced operator notes.
- Extend the CLI with `-v` / `--version`, more descriptive help text, and clearer target-selection errors.
- Improve first-run ergonomics by writing an annotated starter config from `agentspec init`.
- Normalize repeated `--target` values while preserving first-seen order.
- Reposition the committed example workspaces as smoke or validation flows rather than the primary onboarding path.

## Capabilities

### New Capabilities
- `repo-publication-surface`: define the canonical root install path and the user-facing repository onboarding contract.

### Modified Capabilities
- `cli-context-and-target-surfaces`: extend the CLI surface with version flags, clearer help and target errors, annotated starter config output, and duplicate-target normalization.

## Impact

- Root module path and installable CLI entrypoint.
- CLI command wiring, help text, target parsing, and starter-config output.
- CLI and smoke tests that currently assume `cmd/agentspec` as the entrypoint path.
- Root repository documentation and example README positioning.
