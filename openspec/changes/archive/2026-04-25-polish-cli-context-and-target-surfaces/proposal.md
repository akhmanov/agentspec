## Why

`agentspec` currently assumes too much execution context: it reads `agentspec.yaml` from the current directory, derives workspace behavior from `cwd`, and selects the only supported target through `--opencode`. That makes the CLI awkward to use from wrappers, CI, or nested directories, and it leaves the target model too narrow and inconsistent for adding `claude-code` cleanly.

## What Changes

- Add explicit global execution-context controls for workspace root and config path.
- Add environment-variable fallbacks for root and config discovery.
- Change local config path resolution so `path:` selectors resolve relative to the loaded config file instead of the process working directory.
- **BREAKING** Replace `--opencode` with repeatable `--target` selection for `plan` and `apply`.
- Add `claude-code` as a supported target surface alongside `opencode`.
- **BREAKING** Move OpenCode skill artifacts from `.agents/skills/<id>/...` to `.opencode/skills/<id>/...` so the OpenCode adapter owns one fully target-native surface.
- Keep interpolation or template expansion inside `agentspec.yaml` out of scope for this change.

## Capabilities

### New Capabilities
- `cli-context-and-target-surfaces`: Defines the execution-context contract, target-selection behavior, and target-native artifact layout rules for `agentspec` workspace sync commands.

### Modified Capabilities

## Impact

- Affects `cmd/agentspec`, config loading, selector resolution, target adapters, sync state handling, CLI help text, and plan/apply output.
- Adds a new target-specific workspace surface for Claude Code and changes the canonical OpenCode skill output location.
- Requires updates to `SPEC.md`, `ARCH.md`, examples, and tests so the documented product contract matches the new CLI and artifact behavior.
