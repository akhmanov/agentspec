## 1. Expand product context

- [x] 1.1 Add a concise "Why It Exists" section to `AGENTS.md` that explains the workspace-drift problem and the role of `agentspec.yaml` as the source of truth
- [x] 1.2 Update the product and boundary wording in `AGENTS.md` so it reads as a compact explainer rather than only a label list

## 2. Clarify the resource and sync model

- [x] 2.1 Expand the resource model section in `AGENTS.md` so `sections`, `commands`, `agents`, and `skills` are defined by their workspace effect
- [x] 2.2 Add a sync-model section to `AGENTS.md` that explains `agentspec plan --opencode` and `agentspec apply --opencode` as ownership-aware desired-state sync operations

## 3. Verify consistency and brevity

- [x] 3.1 Review the updated `AGENTS.md` against `SPEC.md`, `ARCH.md`, `internal/adapter/opencode`, and `internal/sync` to keep the wording accurate
- [x] 3.2 Trim or rewrite any duplicated detail so `AGENTS.md` stays a compact repo briefing instead of becoming a second full reference
