## Traceability

- Bead: `wspace-325`
- Change: `wspace-325-improve-github-bundles-plan-ui-examples`

## Why

`agentspec` already handles GitHub-backed single-file resources, but directory-backed skills still fall over on some popular public repos because bundle resolution downloads a full archive behind a fixed response-size cap. The same workflow also exposes a flat `plan` preview and has no committed smoke examples, which makes behavior harder to inspect, compare, and trust.

## What Changes

- Improve GitHub directory-backed skill resolution so valid bundles from large public repos do not fail solely because unrelated repository contents make archive downloads impractical.
- Add a grouped default `plan` preview plus an expanded `--verbose` mode without changing preview/apply semantics.
- Add committed `example/local-smoke` and `example/github-smoke` workspaces to exercise local and live GitHub-backed flows.
- Leave legacy `docs/specs/` and `docs/plans/` history in place; new work remains canonically described by bead-mapped OpenSpec artifacts.

## Capabilities

### New Capabilities
- `github-skill-bundles`: Defines how `agentspec` resolves valid directory-backed GitHub skills from public repositories while preserving existing bundle validation and safety checks.
- `plan-preview-output`: Defines the default and verbose preview behavior for `agentspec plan`.
- `repository-smoke-examples`: Defines the committed local and live GitHub example workspaces that demonstrate supported flows.

### Modified Capabilities

## Impact

- `internal/resolve/`
- `cmd/agentspec/`
- `SPEC.md`
- `example/`
- `openspec/changes/wspace-325-improve-github-bundles-plan-ui-examples/`
