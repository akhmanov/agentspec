## Traceability

- Bead: `wspace-325`
- Change: `wspace-325-improve-github-bundles-plan-ui-examples`

## Context

`agentspec` currently resolves GitHub-backed files by fetching raw file content, then resolves directory-backed skills by downloading a GitHub archive and extracting the requested subtree. That works for small repos but fails on at least some popular skill repositories because the archive path is bounded by a fixed HTTP response-size cap. The CLI also prints plan output as flat lines with no mode distinction, and the repo has no committed example workspaces for either deterministic local flows or live public GitHub flows.

This change stays inside the repo's current product boundary. `beads` remains the authority for task identity and status, OpenSpec remains the authority for new change artifacts, and no new selector types, cache layer, registry service, or workflow runtime are introduced.

## Goals / Non-Goals

**Goals:**
- keep GitHub single-file resource handling stable
- make valid directory-backed GitHub skills usable from larger public repositories
- add a compact default `plan` preview and an expanded `--verbose` mode
- add committed local and live GitHub smoke examples
- preserve the OpenSpec and beads authority split

**Non-Goals:**
- adding new selectors such as `gitlab`
- introducing a marketplace, registry, cache, or persisted download store
- replacing deterministic automated tests with live-network-only verification
- changing the ownership, state, or apply semantics in `internal/sync`

## Decisions

- Keep the existing GitHub raw-file path for `sections`, `commands`, `agents`, and single-file `skills`. The large-repo problem is concentrated in directory-backed skill bundles, so the change should stay narrowly scoped to that branch of resolution.
- Replace the GitHub archive fallback for directory-backed skills with a resolver-owned git transport that performs a shallow checkout of the requested ref and reads only the selected subtree. The checkout will disable Git LFS smudging so unrelated large media tracked by LFS does not get pulled into the workspace while resolving text-only skills.
- Introduce a small resolver seam around the git fallback so tests can verify fallback behavior and transport options without depending on live GitHub or shelling out in every unit test. The resolved bundle output should still flow through the same root-`SKILL.md` and frontmatter validation used today.
- Keep diff computation and ownership behavior in `internal/sync` unchanged. The `plan` improvement is a CLI rendering change: default output becomes grouped and compact, while `--verbose` prints the same change set with more detail instead of recomputing a different plan.
- Add two committed example workspaces under `example/`: one fully local and deterministic, and one using pinned live public GitHub repos. The live example is for demonstration and manual smoke coverage, not the sole correctness signal.
- Keep companion docs in `docs/plans/` explicitly marked as non-canonical so the repo contract remains intact: OpenSpec artifacts are authoritative for this new work, and `docs/plans/` remains readable history plus optional companion notes.

## Risks / Trade-offs

- [Git dependency in resolver fallback] -> Mitigation: keep the git path limited to directory-backed GitHub skills, use shallow clone, and cover it with seam-based tests.
- [Live GitHub example drift] -> Mitigation: pin refs, document that the example depends on public upstream availability, and keep deterministic package tests separate.
- [Verbose mode turns into a second UX surface] -> Mitigation: keep one underlying plan model and limit `--verbose` to grouped detail rather than full content dumps.
- [Shallow clone is still heavier than raw-file fetch] -> Mitigation: keep raw-file resolution as the first path and only clone when the requested GitHub skill path is not a single file.

## Migration Plan

- add proposal/spec/design/tasks artifacts for the bead-mapped change
- implement and test the GitHub bundle resolver fallback
- implement and test grouped `plan` rendering with `--verbose`
- add committed example workspaces and accompanying docs
- validate the change artifacts and run the Go test suite plus smoke commands

## Open Questions

- No blocking design questions remain for this slice. A future follow-up may add `plan --json`, but that is intentionally left out of this change.
