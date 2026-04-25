## 1. Add execution-context plumbing

- [x] 1.1 Add global `--root` and `--config` handling with `AGENTSPEC_ROOT` and `AGENTSPEC_CONFIG` fallbacks, and share the effective execution context across `init`, `plan`, and `apply`
- [x] 1.2 Update CLI tests to cover flag and environment precedence, default config discovery from the effective root, and relative `--config` resolution

## 2. Make local selectors config-relative

- [x] 2.1 Change config and resolve plumbing so relative local `path:` selectors resolve from the loaded config file directory instead of the process working directory
- [x] 2.2 Add config and resolve tests for config-relative local paths and absolute-path preservation

## 3. Replace boolean targets with repeatable `--target`

- [x] 3.1 Replace `--opencode` parsing with repeatable `--target` validation and target dispatch for `plan` and `apply`
- [x] 3.2 Update plan and apply output and CLI tests so multi-target runs are grouped per target, reject missing targets, and preserve sequential apply behavior

## 4. Make target adapters fully target-native

- [x] 4.1 Update the OpenCode adapter and sync path validation so OpenCode skills render to `.opencode/skills/<id>/...` alongside the existing `.opencode/commands/...` and `.opencode/agents/...` surfaces
- [x] 4.2 Add transition-aware OpenCode state handling and tests so prior managed `.agents/skills/...` outputs can move safely to `.opencode/skills/...`
- [x] 4.3 Add the Claude Code adapter and tests for `CLAUDE.md`, `.claude/commands/<id>.md`, `.claude/agents/<id>.md`, `.claude/skills/<id>/...`, and per-target state isolation

## 5. Align product docs and verification

- [x] 5.1 Update `SPEC.md`, `ARCH.md`, examples, and command help text to describe `--root`, `--config`, repeatable `--target`, and the new target-native artifact paths
- [x] 5.2 Run OpenSpec validation plus focused and full Go test coverage for CLI, resolve, adapter, and sync behavior, and fix any contract drift that appears during verification
