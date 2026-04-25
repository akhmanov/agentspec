# Architecture

`agentspec` is a small Go CLI that turns one declarative `agentspec.yaml` file into a target-specific workspace surface.

This document defines the repo-level architecture contract for v1. It exists to keep the code simple, boundary-safe, and easy to extend without turning a small CLI into a framework.

## Goals

- Keep the core product loop small: load config, resolve sources, render target outputs, sync safely.
- Keep target-neutral logic separate from target-specific rendering.
- Prefer explicit code over hidden magic.
- Favor simple package boundaries over abstract "clean architecture" theater.
- Make ownership and safety rules obvious in code and tests.

## Non-Goals

- No generic plugin system.
- No repository or service abstraction unless a real boundary requires it.
- No runtime orchestration, task lifecycle, or package-management behavior.
- No speculative remote-provider framework before local selectors are working.

## Package Direction

The expected package shape for the first slice is:

- repository root
  - CLI transport only
  - process entrypoint
  - command wiring and execution-context parsing
- `internal/config`
  - parse and validate `agentspec.yaml`
- `internal/resolve`
  - resolve `inline` and `path`
  - later `http`, `github`, `gitlab`
- `internal/model`
  - normalized resolved resource shapes shared by adapters and sync
- `internal/adapter`
  - target selection and target-specific rendering dispatch
- `internal/adapter/opencode`
  - OpenCode rendering only
- `internal/adapter/claudecode`
  - Claude Code rendering only
- `internal/sync`
  - desired-state apply, marker updates, and orphan cleanup
- `internal/state`
  - only if ownership-safe cleanup later proves to need durable tracking

Dependencies must point inward.

- the repository root package may depend on `internal/config`, `internal/resolve`, `internal/model`, adapter packages, and `internal/sync`.
- `internal/config` must not depend on adapter or sync packages.
- `internal/resolve` must not know target file layout and should resolve local selectors relative to the loaded config file.
- adapters must not parse raw config or resolve selectors.
- `internal/sync` may operate on rendered desired outputs, but must not absorb target-specific resolve logic.

## Core Flow

The intended execution flow is:

1. CLI loads args and selects a command.
2. Config layer loads and validates `agentspec.yaml`.
3. Resolve layer resolves configured sources.
4. Model layer holds normalized resolved resources.
5. Target adapter renders desired outputs.
6. Sync layer applies owned file changes and marker-scoped instruction updates.

Business rules should live in the smallest package that owns them.

## CLI Rules

- Use `github.com/urfave/cli/v3`.
- Keep CLI handlers thin.
- Parse flags, call package logic, print user-facing output, return errors.
- Do not place validation, resolution, rendering, or sync policy directly in CLI command handlers.
- Exit the process once, in `main`.

## Error Handling

- Return errors upward instead of logging and returning the same error.
- Add short context at package boundaries.
- Avoid panic for expected failures.
- Keep user-facing errors clear about resource kind, resource id, path, or selector when relevant.

## DI And Boundaries

Use dependency injection only at real edges.

Good candidates:

- filesystem access when it improves testing or boundary clarity
- target adapter selection
- future HTTP client injection for remote selectors

Avoid interface-first design for internal code that has only one concrete implementation and no real seam yet.

Small interfaces defined near the consumer are preferred over shared "god" interfaces.

## Testing Strategy

- Follow TDD for production code changes.
- Prefer table-driven tests for validation and resolution cases.
- Use fixture-based tests for sync and marker behavior.
- Favor real filesystem tests with temp directories over mocks when practical.
- Test ownership boundaries explicitly.

Every new behavior should be justified by a failing test first.

## Code Style

The repo follows standard Go formatting plus selected rules from the Uber Go Style Guide:

- keep functions and packages small
- no `init()` magic
- no mutable globals for behavior
- one process exit in `main`
- explicit errors with concise wrapping
- table-driven tests where they improve clarity
- composition over embedding when public API shape matters

Also for this repo:

- prefer simple names over layered jargon
- prefer one clear flow over extra indirection
- add helpers only when reuse or readability is real
- keep comments rare and useful

## Decision Rule

When two approaches are both correct, choose the one that:

1. introduces fewer abstractions
2. keeps ownership boundaries clearer
3. makes tests easier to write and understand
4. keeps future remote-selector work possible without designing it now
