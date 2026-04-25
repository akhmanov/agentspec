# AGENTS

## Product

`agentspec` is a small Go CLI that materializes agent workspace resources from `agentspec.yaml`.

## Product Boundary

`agentspec` is a declarative workspace sync tool. It is not a workflow runtime, task tracker, bootstrap tool, or package manager.

## Core User Flow

1. Run `agentspec init` to create `agentspec.yaml`.
2. Describe the desired workspace resources in `agentspec.yaml`.
3. Run `agentspec plan --opencode` to preview managed changes.
4. Run `agentspec apply --opencode` to materialize the workspace surface.

## Resource Model

The top-level resource types are:

- `sections`
- `commands`
- `agents`
- `skills`

## Architecture Boundaries

Keep the package direction clear:

- `cmd/agentspec` wires the CLI only.
- `internal/config` loads and validates config.
- `internal/resolve` resolves sources.
- `internal/model` holds normalized resource shapes.
- `internal/adapter/opencode` renders target-specific output.
- `internal/sync` previews and applies owned workspace changes.

Do not move target-specific rendering into resolve logic or workflow behavior into the CLI.

## Repo Rules

- Prefer minimal changes.
- Keep target-neutral logic separate from target-specific rendering.
- Do not add workflow-runtime behavior.
- Validate before claiming completion.

## OpenSpec Lifecycle

Use the standard OpenSpec flow:

1. propose
2. specs
3. design
4. tasks
5. apply
6. validate
7. archive

## Task Tracking

In this repo, OpenSpec `tasks.md` is the task-state source of truth.

## Key References

- `SPEC.md`
- `ARCH.md`

## Retained Local Skills

Retain and use these repo-local skills when they apply:

- `backend-architecture`
- `api-contracts`
- `security-review`
- `infra-devops`
- `gitops-review`
