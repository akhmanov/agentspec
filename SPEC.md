# AgentSpec Specification

- Status: Draft
- Revision: 2
- Supersedes: Revision 1

## Summary

`agentspec` is a small Go CLI that turns a single declarative `agentspec.yaml` file into an agent workspace surface for a supported target.

In the current v1 slice, `agentspec` is a preview-and-apply sync tool only. It does not own workflow execution, task tracking, project bootstrap, package installation, or runtime orchestration.

## Problem

Setting up an agent workspace across many projects currently requires repeating the same manual work:

- installing or copying skills
- adding commands
- adding agents
- editing instruction files such as `AGENTS.md` or `CLAUDE.md`
- keeping all of those pieces in sync over time

This process is slow, error-prone, and hard to reproduce. Small changes to workspace conventions require touching several files by hand. Existing workspaces drift because there is no single source of truth.

## Goal

Provide one declarative file, `agentspec.yaml`, plus a small CLI that can materialize a workspace surface safely and repeatably.

Core user flow:

1. Create `agentspec.yaml` with `agentspec init`.
2. Describe the desired workspace resources in `agentspec.yaml`.
3. Run `agentspec plan --opencode` to preview managed changes.
4. Run `agentspec apply --opencode` to materialize those changes.
5. Get a predictable, target-specific workspace surface with only `agentspec`-owned parts updated.

## Non-Goals

The following are explicitly out of scope for v1:

- owning or defining a lifecycle framework
- embedding or wrapping `beads`
- adding `agentspec task`
- requiring or generating `mise`
- using `npx skills`
- depending on `skills.sh`
- acting as a project bootstrap tool
- acting as a generic merge engine for arbitrary existing target files
- auto-migrating an arbitrary existing workspace into the `agentspec` model
- supporting packs as part of the initial v1 scope
- supporting multi-file inline resources

## Design Principles

- Simplicity over abstraction.
- Readability over perfect uniformity.
- Conservative ownership.
- Target adapters render; schema stays mostly target-neutral.
- Hidden magic is a bug.
- Only `agentspec`-owned files and `agentspec` markers may be updated or deleted.

## Product Shape

`agentspec` is a declarative resource sync tool.

It owns four top-level resource types in v1:

- `sections`
- `commands`
- `agents`
- `skills`

The first three are single-document resources. `skills` are the only special-case resource type because the ecosystem already treats them as directory-based resources centered around `SKILL.md`.

## CLI Surface

### `agentspec init`

Creates a starter `agentspec.yaml`.

Constraints:

- no aggressive discovery
- no migration of an existing workspace
- no side effects beyond writing the config file

### `agentspec plan`

Previews the managed workspace changes for a supported target.

Initial target flags:

- `agentspec plan --opencode`

Constraints:

- target-specific rendering only
- no writes to workspace files or state
- no silent rewriting of foreign files
- surface ownership conflicts clearly

### `agentspec apply`

Materializes the desired workspace state for a supported target.

Initial target flags:

- `agentspec apply --opencode`

Constraints:

- target-specific rendering only
- no silent rewriting of foreign files
- delete only `agentspec`-owned orphaned files
- update only `agentspec` markers inside instruction files
- recompute current desired state instead of consuming a saved plan artifact

### Deferred CLI Surface

These are deferred from v1 and are not required by this spec:

- `agentspec check`
- `agentspec doctor`
- any pack-specific commands

## `agentspec.yaml` Schema

### Top-Level Shape

```yaml
sections: {}
commands: {}
agents: {}
skills: {}
```

Each map key is a user-defined resource id.

Example:

```yaml
sections:
  workspace-core:
    inline: |
      Core workspace rules...

commands:
  explore:
    path: ./.agentspec/commands/explore.md

agents:
  reality-checker:
    github:
      repo: msitarzewski/agency-agents
      ref: main
      path: testing/testing-reality-checker.md

skills:
  frontend-design:
    github:
      repo: anthropics/skills
      ref: main
      path: skills/frontend-design
```

## Source Selectors

Every resource entry uses exactly one source selector.

Supported selectors in v1:

- `inline`
- `path`
- `http`
- `github`

Deferred selectors:

- `gitlab`

### `inline`

Inline content embedded directly in `agentspec.yaml`.

```yaml
sections:
  workspace-core:
    inline: |
      Core workspace rules...
```

Rules:

- single-file only in v1
- plain text or markdown payload

### `path`

Filesystem path relative to the project root or absolute path.

For `sections`, `commands`, and `agents`, `path` points to one file.

For `skills`, `path` may point to either:

- a file
- a directory

### `http`

HTTPS URL to one markdown file.

Rules:

- HTTPS only
- single-file only
- for `skills`, `http` must point directly to a `SKILL.md`-compatible document

### `github`

GitHub-backed source.

```yaml
skills:
  frontend-design:
    github:
      repo: anthropics/skills
      ref: main
      path: skills/frontend-design
```

Fields:

- `repo`: `owner/repo`
- `ref`: pinned branch, tag, or commit-ish
- `path`: path inside the repo

`gitlab` is deferred from the current v1 slice and is not part of the supported selector set for the current CLI.

## Resource Semantics

### Sections

Purpose:

- universal instruction chunks
- rendered by a target adapter into that target's primary instruction surface

Shape:

- single markdown document only
- source selectors: `inline`, `path`, `http`, `github`

Example:

```yaml
sections:
  workspace-core:
    inline: |
      Core workspace rules...

  review-rules:
    path: ./.agentspec/sections/review-rules.md
```

Important behavior:

- sections are universal in v1
- there is no target filtering in schema
- section order is preserved from YAML order

### Commands

Purpose:

- command documents materialized into target-specific command surfaces

Shape:

- single markdown document only
- source selectors: `inline`, `path`, `http`, `github`

Example:

```yaml
commands:
  explore:
    inline: |
      Explore the task before formal commitment...

  review:
    path: ./.agentspec/commands/review.md
```

Important behavior:

- commands are universal resources
- adapters decide how they are rendered for a target

### Agents

Purpose:

- agent prompt documents materialized into target-specific agent surfaces

Shape:

- single markdown document only
- source selectors: `inline`, `path`, `http`, `github`

Example:

```yaml
agents:
  reality-checker:
    github:
      repo: msitarzewski/agency-agents
      ref: main
      path: testing/testing-reality-checker.md
```

Important behavior:

- agents are universal resources
- adapters decide how they are rendered for a target

### Skills

Purpose:

- import or define reusable `SKILL.md`-based skills

Shape:

- only resource type with file-or-directory semantics

#### Skill Rules in v1

- `inline` = single-file only
- `http` = single-file only
- `path`, `github`:
  - if resolved path is a file, save only that file
  - if resolved path is a directory, import the bundle

Examples:

```yaml
skills:
  local-review:
    inline: |
      ---
      name: local-review
      description: Review local changes
      ---
      ...

  release-check:
    http: https://example.com/skills/release-check/SKILL.md

  systematic-debugging:
    path: ./.agentspec/skills/systematic-debugging

  frontend-design:
    github:
      repo: anthropics/skills
      ref: main
      path: skills/frontend-design
```

Materialization rules:

- single-file skill -> write as `.agents/skills/<id>/SKILL.md`
- directory skill -> write full bundle under `.agents/skills/<id>/...`

Validation rules:

- single-file skills must parse as valid skill content
- bundled skills must contain `SKILL.md` at bundle root

## Target Adapters

Schema stays mostly target-neutral. Rendering is target-specific.

### OpenCode Adapter

OpenCode is the only supported target in the current v1 slice.

Responsibilities:

- materialize `sections` into managed sections inside the primary instruction file
- materialize `commands` into `.opencode/commands/<id>.md`
- materialize `agents` into `.opencode/agents/<id>.md`
- materialize `skills` into `.agents/skills/<id>/...`

### Deferred Targets

Additional targets such as Claude Code are deferred from the current v1 slice.

When future target work is added, it should:

- materialize only clearly supported resource kinds
- emit clear warnings for unsupported resource kinds
- avoid silent no-op behavior
- avoid fake parity with OpenCode

## Instruction File Materialization

Adapters own instruction-file placement.

`agentspec.yaml` does not mention `AGENTS.md`, `CLAUDE.md`, or any other instruction filename directly.

Instead:

- `sections` are universal
- each target decides which instruction file to use
- each section becomes its own managed block

### Marker Format

Simple `agentspec` markers:

```md
<!-- agentspec:section:start workspace-core -->
...content...
<!-- agentspec:section:end workspace-core -->
```

Rules:

- one managed block per section
- preserve YAML order
- update only content inside `agentspec` markers
- never rewrite content outside `agentspec` markers

## Ownership Model

`agentspec` may only modify things it owns.

### `agentspec`-Owned Files

Examples:

- materialized `.opencode/commands/*.md`
- materialized `.opencode/agents/*.md`
- materialized `.agents/skills/<id>/...`
- persisted `.agentspec/state/<target>.json`

Ownership policy:

- create and update only `agentspec`-owned files
- delete only orphaned `agentspec`-owned files
- do not claim pre-existing foreign files just because a path matches

### `agentspec`-Managed Regions

In instruction files, ownership is marker-scoped.

Policy:

- only content inside `agentspec` markers is managed
- everything outside those markers is foreign content

## Orphan Deletion Policy

If a resource is removed from `agentspec.yaml`:

- delete the corresponding orphaned `agentspec`-owned file(s)
- remove the corresponding orphaned managed section block

Do not delete:

- foreign files
- foreign regions in instruction files

## Existing Workspace Adoption

v1 uses conservative adoption.

Meaning:

- `agentspec` works alongside an existing workspace
- it does not try to auto-normalize or import arbitrary existing files into schema
- it only manages what it explicitly owns

This avoids accidental takeover of hand-maintained workspaces.

## Plan And Apply Semantics

`agentspec plan` should conceptually perform these steps:

1. Load and validate `agentspec.yaml`.
2. Resolve each resource from its selected source.
3. Normalize resources into internal resolved forms.
4. Ask the target adapter to build desired output.
5. Compare desired output against the current workspace.
6. Report managed creates, updates, deletes, and ownership conflicts without writing files or state.

`agentspec apply` should conceptually perform these steps:

1. Load and validate `agentspec.yaml`.
2. Resolve each resource from its selected source.
3. Normalize resources into internal resolved forms.
4. Ask the target adapter to build desired output.
5. Apply file updates for `agentspec`-owned files.
6. Apply section updates inside `agentspec` markers.
7. Remove orphaned `agentspec`-owned outputs.
8. Persist owned state if needed for safe future apply and prune behavior.

`agentspec apply` recomputes desired state from current config and current sources instead of consuming output from a previous `agentspec plan` run.

## Internal Architecture

Expected layers:

- `config`
  - parse and validate `agentspec.yaml`
- `resolve`
  - inline/path/http/github
- `model`
  - normalized resolved resources
- `adapter`
  - target-specific rendering
- `sync`
  - apply desired state and orphan cleanup
- `state`
  - track `agentspec`-owned outputs if needed
- `cmd/agentspec`
  - CLI transport

Key boundary:

- fetch/resolve logic must not know about `.opencode` or `CLAUDE.md`
- adapters must not own source parsing

## Success Criteria

v1 is successful when:

- a developer can describe a workspace using one `agentspec.yaml`
- `agentspec plan --opencode` can safely preview managed changes without writing files or state
- `agentspec apply --opencode` can safely materialize the agreed resources
- only `agentspec`-owned files are updated or deleted
- only `agentspec` markers are modified in instruction files
- the resulting schema stays readable without hidden modes or a mini DSL

## Deferred Questions

These are intentionally deferred and not required for v1:

- packs
- `agentspec check`
- `agentspec doctor`
- richer metadata for commands or agents outside markdown
- target filters on sections
- one big managed instruction region instead of per-section blocks
- multi-file inline skills
- multi-file commands or agents
- direct support for additional source kinds such as `gitlab`
- additional targets such as Claude Code

## Why This Direction

This direction is worth doing because it captures the repeatable value in a workspace setup without taking ownership of runtime or workflow policy.

It keeps the schema small, makes target-specific behavior explicit, and gives `agentspec` a narrow, high-confidence responsibility: turn declarative workspace resources into a safe, reproducible target surface.
