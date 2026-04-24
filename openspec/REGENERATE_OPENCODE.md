# Regenerate OpenCode Surface

The OpenCode workflow surface in `.opencode/` is not pure upstream output.

This repo maintains three layers:

1. Upstream OpenSpec-generated base for OpenCode.
2. Repo-maintained patches to the generated bead-aware workflow files.
3. Repo-local supplements that upstream core does not generate.

Use the checked-in regeneration path instead of relying on session memory:

```bash
./openspec/regenerate-opencode-surface.sh
```

## What The Script Does

1. Runs pinned upstream OpenSpec update for instruction files.
2. Removes the retired workspace lifecycle shell artifacts and scrubs unexpected future `opsx-*` and `openspec-*` drift.
3. Reapplies repo-maintained files from `openspec/opencode-overrides/`.

## Maintained Repo Layer

Repo-maintained patched files:

- `.opencode/opencode.json`
- `.opencode/instructions/INSTRUCTIONS.md`
- `.opencode/commands/opsx-explore.md`
- `.opencode/commands/opsx-propose.md`
- `.opencode/commands/opsx-apply.md`
- `.opencode/commands/opsx-archive.md`
- `.opencode/skills/openspec-explore/SKILL.md`
- `.opencode/skills/openspec-propose/SKILL.md`
- `.opencode/skills/openspec-apply-change/SKILL.md`
- `.opencode/skills/openspec-archive-change/SKILL.md`

Repo-local supplements:

- `.opencode/commands/opsx-continue.md`
- `.opencode/skills/openspec-continue-change/SKILL.md`

The source of truth for those patched and supplemental files lives under `openspec/opencode-overrides/`.

## After Regeneration

Run the normal workflow verification for the change you are working on:

```bash
openspec validate "<change-name>" --type change
go test ./...
```

If upstream OpenSpec changes the generated core surface in a way that conflicts with this repo contract, reconcile the diff in `openspec/opencode-overrides/` instead of patching `.opencode/` ad hoc.
