# agentspec

`agentspec` is a small Go CLI that materializes a managed agent workspace surface from one declarative `agentspec.yaml` file.

It exists to stop workspace drift: instead of hand-copying commands, agents, skills, and shared instruction sections across repositories, you declare the desired surface once and let `agentspec` plan and apply the managed outputs.

## Install

```bash
go install github.com/akhmanov/agentspec@latest
```

Verify the install:

```bash
agentspec --version
```

## Quickstart

Create a starter config:

```bash
agentspec init
```

Minimal example:

```yaml
sections:
  workspace-core:
    inline: |
      Workspace rules

commands:
  explore:
    inline: |
      Explore the repository carefully.

agents: {}
skills: {}
```

Preview and apply the managed workspace surface:

```bash
agentspec plan --target opencode
agentspec apply --target opencode
```

## Supported Targets

- `opencode`
- `claude-code`

## Ownership Model

- `agentspec plan --target <target>` previews managed creates, updates, deletes, and conflicts without writing workspace files.
- `agentspec apply --target <target>` materializes only the managed surface for that target.
- Foreign content is not silently overwritten. Ownership conflicts are surfaced instead.

## Examples

- `example/local-smoke`: deterministic smoke validation using only repository-local resources
- `example/github-smoke`: live smoke validation for pinned GitHub-backed sources

These examples are validation flows for the repository. The canonical user path is the root install and quickstart flow above.

## Advanced Environment

- `AGENTSPEC_ROOT`: default workspace root when `--root` is not provided
- `AGENTSPEC_CONFIG`: default config path when `--config` is not provided
- `AGENTSPEC_GIT_TIMEOUT`: overrides the timeout for internal git operations used when resolving GitHub-backed skill bundles; use Go duration values such as `60s`

## Reference

- `SPEC.md`
- `ARCH.md`
