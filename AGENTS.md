# AGENTS

## Product

`agentspec` is a small Go CLI that materializes a managed agent workspace surface from `agentspec.yaml`.

## Why It Exists

Maintaining agent workspace files by hand across repositories means repeatedly copying commands, agents, skills, and shared instruction text, then keeping them aligned over time. `agentspec` exists to stop that drift by treating `agentspec.yaml` as the single source of truth for the managed workspace surface.

## Product Boundary

`agentspec` is a declarative workspace sync tool. It is not a workflow runtime, task tracker, bootstrap tool, or package manager.

## Sync Flow

1. Run `agentspec init` to create `agentspec.yaml`.
2. Describe the desired workspace resources in `agentspec.yaml`.
3. Run `agentspec plan --target opencode` to preview managed changes.
4. Run `agentspec apply --target opencode` to materialize the workspace surface.

## Resource Model

The top-level resource types describe what `agentspec` materializes into the workspace:

- `sections`: managed instruction fragments inserted into `AGENTS.md`
- `commands`: target command documents such as `.opencode/commands/<id>.md`
- `agents`: target agent documents such as `.opencode/agents/<id>.md`
- `skills`: target-native skill bundles such as `.opencode/skills/<id>/...` or `.claude/skills/<id>/...`

These are workspace resources, not runtime objects. `agentspec` materializes their files; it does not execute them.

## Sync Model

- `agentspec plan --target <target>` previews the managed create, update, delete, and conflict set without writing workspace files.
- `agentspec apply --target <target>` materializes the desired managed state, updates managed instruction sections, prunes orphaned `agentspec`-owned outputs, and refuses to silently overwrite foreign content.

Only the `agentspec`-owned surface is updated. Foreign content is left alone unless it is already inside an `agentspec`-managed boundary.

## Architecture Boundaries

Keep the package direction clear:

- `cmd/agentspec` wires the CLI only.
- `internal/config` loads and validates config.
- `internal/resolve` resolves sources.
- `internal/model` holds normalized resource shapes.
- `internal/adapter` selects target-specific renderers.
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
