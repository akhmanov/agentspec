# OpenSpec + Beads Operating Model

This repository uses `OpenSpec + Beads`.

`OpenSpec` is the primary repo-local workflow for change artifacts.
`beads` is the primary repo-local workflow for task state and continuity.

## Authority Matrix

| Concern | System of record |
| --- | --- |
| Change intent and why | `openspec/changes/<change>/proposal.md` |
| Behavioral requirements | `openspec/specs/` and approved deltas in `openspec/changes/<change>/specs/` |
| Technical design | `openspec/changes/<change>/design.md` |
| Execution checklist | `openspec/changes/<change>/tasks.md` |
| Task identity | `beads` |
| Task status | `beads` |
| Assignee / priority / dependencies | `beads` |
| Operational continuity and coordination | `beads` |
| Legacy workflow history | `docs/specs/` and `docs/plans/` |

`tasks.md` is an execution checklist, not the authoritative task manager.

No third canonical memory layer is allowed.

## OpenCode Workflow Surface

The repo targets the OpenSpec-generated OpenCode workflow surface.

The generated command set uses the reproducible core profile, plus one repo-local supplement required to complete blocked changes:

- `opsx-explore`
- `opsx-propose`
- `opsx-continue`
- `opsx-apply`
- `opsx-archive`

Regenerate or update this surface with `./openspec/regenerate-opencode-surface.sh`. That path refreshes the upstream OpenSpec instruction files, removes the retired lifecycle shell, and reapplies this repo's maintained patches and supplements from `openspec/opencode-overrides/`.

The repo still requires validation before review and integration, but that validation should be driven by explicit CLI usage such as `openspec validate <change-name> --type change` plus repo-local review policy instead of a hidden dependency on per-user global OpenSpec workflow configuration.

## Bead To Change Mapping

Each OpenSpec change maps to exactly one bead.

The mapping contract is:

1. OpenSpec change directory names are prefixed with the bead id.
2. The bead stores the mapped change id in metadata under `openspec.change`.
3. Generated OpenSpec artifacts should surface the bead id in visible content or metadata when templates allow it.

Example:

- bead: `wspace-q66`
- change id: `wspace-q66-transition-workspace-lifecycle-to-openspec`

This rule must stay deterministic and auditable from repo files plus bead metadata.

## Active-Work Cutover Rule

Active pre-cutover beads stay on the legacy workflow unless they are explicitly bridged.

The default cutover rule is:

1. Finish active pre-cutover work under the workflow it already started with.
2. Start new work under OpenSpec only after the OpenCode workflow surface flips.
3. Do not migrate an in-flight bead into OpenSpec silently.
4. If an active bead must move, record the bridge explicitly in the bead and in the mapped OpenSpec change.

Current examples include `wspace-yvu` and `wspace-jrq`.

## Retained And Retired Repo-Local Assets

Retained after cutover:

- domain and cross-cutting skills that complement OpenSpec rather than replace it
- `backend-architecture`
- `api-contracts`
- `security-review`
- `infra-devops`
- `gitops-review`
- process skills from the baseline that remain compatible with OpenSpec

Retired after cutover:

- the custom lifecycle command shell in `.opencode/opencode.json`
- lifecycle command files in `.opencode/commands/` for `/explore`, `/debate`, `/spec`, `/plan`, `/build`, `/test`, `/review`, `/ship`
- lifecycle-specific repo-local skills that recreate the retired shell
- lifecycle debate-agent scaffolding that only served the retired workflow

## Legacy History Policy

`docs/specs/` and `docs/plans/` remain in the repository as legacy history.

They are not the canonical source for new work after the OpenSpec cutover.
