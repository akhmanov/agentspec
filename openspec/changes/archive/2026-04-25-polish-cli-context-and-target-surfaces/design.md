## Context

`agentspec` currently hardcodes too much runtime context into the CLI transport. `cmd/agentspec` reads `agentspec.yaml` from the current directory, derives the workspace root from `cwd`, and selects the only supported target through `--opencode`. At the same time, the OpenCode adapter already exposes an inconsistent target surface: `commands` and `agents` render into `.opencode/...`, while `skills` still render into `.agents/skills/...`.

This change crosses multiple architecture boundaries:

- CLI transport must parse explicit execution context and target selection.
- Config and resolve layers must stop assuming that local selectors are rooted at the process working directory.
- Adapter selection must scale from one hardcoded target to multiple target-specific renderers.
- Sync and state handling must preserve ownership safety while OpenCode skill outputs move to a new canonical path.

The design therefore needs to keep responsibilities separated: transport chooses context and targets, resolve interprets selectors relative to the loaded config, adapters own target-specific paths, and sync continues to own planning, apply, and per-target state safety.

## Goals / Non-Goals

**Goals:**

- Add explicit execution-context controls for workspace root and config path.
- Introduce repeatable `--target` selection for `plan` and `apply`.
- Make local `path:` selectors resolve relative to the loaded config file.
- Support both `opencode` and `claude-code` through target-native adapters.
- Move OpenCode skills to `.opencode/skills/...` without abandoning ownership-safe cleanup for already managed outputs.
- Keep plan/apply behavior simple, explicit, and aligned with the current desired-state sync model.

**Non-Goals:**

- Add interpolation or template expansion inside `agentspec.yaml`.
- Add install-path overrides such as `--skills-dir` or `--state-path`.
- Add short flags in this slice.
- Introduce a universal `.agents` target or any cross-target shared artifact surface.
- Promise atomic multi-target apply or cross-target rollback.

## Decisions

### Decision: Represent execution context explicitly instead of relying on process cwd

The CLI should compute one effective execution context before dispatching a command:

- `root = --root | AGENTSPEC_ROOT | cwd`
- `config = --config | AGENTSPEC_CONFIG | <root>/agentspec.yaml`

Relative `--config` values should resolve against `root` when `--root` is provided, otherwise against `cwd`. This keeps `root` as the workspace boundary without mutating process-global working-directory state.

This choice keeps ownership semantics explicit. The workspace root is a product concept, not just a shell convenience.

Alternatives considered:

- Rely on `cwd` only
  Rejected because it keeps CI, wrapper, and nested-directory usage awkward.
- Model the feature as implicit `chdir`
  Rejected because it hides the workspace-boundary concept behind shell-like behavior and makes future path rules harder to reason about.

### Decision: Resolve local selectors relative to the loaded config file

Local `path:` selectors should resolve relative to the directory that contains the loaded config file, not relative to the workspace root or process working directory.

This makes the config file self-contained: moving `agentspec.yaml` into a subdirectory keeps nearby local references coherent. It also matches common config-file behavior in other CLIs without leaking target knowledge into the resolve layer.

Alternatives considered:

- Keep selectors root-relative
  Rejected because it couples selector meaning to invocation context instead of config location.
- Keep selectors cwd-relative
  Rejected because it makes behavior brittle and dependent on how the command is launched.

### Decision: Replace target booleans with repeatable `--target`

`plan` and `apply` should accept one or more `--target` flags instead of target-specific booleans. The CLI should preserve the order provided by the user and process targets sequentially.

This keeps the transport surface scalable as additional targets appear and avoids hardcoding mutual-exclusion logic around one flag per target.

For apply behavior, sequential execution with no rollback is the right level of complexity for this product. `agentspec` is a declarative sync tool, not an orchestration runtime.

Alternatives considered:

- Keep `--opencode` and add `--claude-code`
  Rejected because the surface grows linearly and becomes harder to validate and document.
- Implicitly apply all known targets
  Rejected because it hides target choice and increases blast radius.

### Decision: Keep artifact paths fully target-native

Each adapter should own one canonical surface:

- OpenCode: `AGENTS.md`, `.opencode/commands/...`, `.opencode/agents/...`, `.opencode/skills/...`
- Claude Code: `CLAUDE.md`, `.claude/commands/...`, `.claude/agents/...`, `.claude/skills/...`

The OpenCode adapter should stop using `.agents/skills/...`. If a future shared `.agents` surface is worth supporting, it should be added explicitly as its own product concept rather than staying as a hidden exception inside one adapter.

Alternatives considered:

- Keep the current hybrid OpenCode layout
  Rejected because it breaks the target-native mental model and makes multi-target behavior harder to explain.
- Add configurable install roots now
  Rejected because that turns stable adapter contracts into per-run behavior and widens the CLI surface prematurely.

### Decision: Migrate OpenCode skill ownership conservatively through state-aware transition handling

The OpenCode skill path change touches persisted ownership state, so the change needs a bounded migration rule. Existing state that tracks `.agents/skills/...` for the `opencode` target should continue to load during the transition, allowing `plan` to show deletes for old managed outputs and creates for new `.opencode/skills/...` outputs. New state written after apply should contain only the new target-native paths.

This preserves conservative ownership: already managed files can transition cleanly, while foreign files are still protected.

Alternatives considered:

- Reject old OpenCode state as invalid immediately
  Rejected because it strands previously managed outputs and breaks ownership-safe cleanup for persisted data.
- Support both old and new OpenCode skill paths indefinitely
  Rejected because it keeps the product in a permanent dual-contract state.

### Decision: Report and store results per target

Planning and apply should stay target-scoped internally. Each target keeps its own desired output, plan result, and state file such as `.agentspec/state/opencode.json` or `.agentspec/state/claude-code.json`.

User-facing plan output should make target boundaries visible instead of flattening changes across multiple targets into one ambiguous list.

Alternatives considered:

- Merge all targets into one combined plan
  Rejected because it hides which adapter owns each change.
- Add transactional coordination across targets
  Rejected because it would push `agentspec` toward workflow-runtime behavior.

## Risks / Trade-offs

- [Risk] Replacing `--opencode` with `--target` is a breaking CLI change.
  -> Mitigation: update canonical docs, examples, and tests together so the new contract is unambiguous.

- [Risk] Config-relative selector resolution changes behavior for repositories that assumed root-relative `path:` semantics.
  -> Mitigation: document the rule clearly and cover it with focused config/resolve tests.

- [Risk] Multi-target apply can leave earlier targets applied when a later target fails.
  -> Mitigation: keep the behavior explicit, sequential, and per-target; do not claim rollback semantics the product does not provide.

- [Risk] OpenCode skill-path migration could accidentally weaken ownership validation.
  -> Mitigation: limit dual-path support to loading prior managed state, continue validating hashes, and emit only the new path on write.

- [Risk] Claude Code support could drift into fake parity if a resource type is not actually target-supported.
  -> Mitigation: keep the adapter contract explicit and verify the target-native surface through tests and docs before claiming support.

## Migration Plan

1. Add an explicit execution-context layer in the CLI that computes effective root and config path once and shares it with `init`, `plan`, and `apply`.
2. Update config loading and local path resolution so selector interpretation uses the loaded config directory rather than `cwd`.
3. Replace boolean target selection with repeatable `--target` parsing and a small target registry that dispatches to target-specific adapters.
4. Add the Claude Code adapter and update the OpenCode adapter so all artifacts, including skills, render to target-native paths.
5. Teach OpenCode state loading to recognize previously managed `.agents/skills/...` entries during migration while writing only `.opencode/skills/...` on new apply.
6. Update sync output, docs, examples, and tests to the new CLI and artifact contract.

Rollback is a code and spec revert. Because the on-disk OpenCode skill location changes intentionally, rolling back after a new apply may require a follow-up apply under the previous contract to rematerialize old paths.

## Open Questions

- None at the moment. The major user-facing choices for execution context, target selection, target-native paths, and out-of-scope interpolation have already been resolved for this change.
