## Context

`agentspec` already has a small and healthy implementation core: the CLI verbs are stable, the test suite is green, and the repo has clear product and architecture documents. The remaining publication work is mostly about the public surface. Today the module path is still `agentspec`, the installable entrypoint lives under `cmd/agentspec`, there is no root `README.md`, the CLI does not expose `-v` or `--version`, the help text does not explain defaults and supported targets well enough, `init` writes an empty unannotated scaffold, and `AGENTSPEC_GIT_TIMEOUT` exists as a real operator knob without user-facing documentation.

The change needs to improve first-run experience without expanding the product boundary. `agentspec` remains a declarative workspace sync tool with the same `init`, `plan`, and `apply` model; the work here is to make that model installable, discoverable, and self-describing from the repository root.

## Goals / Non-Goals

**Goals:**
- Make the published module directly installable with `go install github.com/akhmanov/agentspec@latest`.
- Add a canonical root README that explains installation, version verification, quickstart, supported targets, ownership semantics, and advanced environment knobs.
- Expose a stable version surface through `-v` and `--version`.
- Improve help text, invalid-target errors, and starter-config output so the CLI is understandable without reading internal docs.
- Keep duplicate `--target` values from producing duplicate work or duplicate output groups.

**Non-Goals:**
- Add new workflow commands such as `check`, `doctor`, or release-management commands.
- Introduce package-manager-specific installation guidance beyond the canonical `go install` path.
- Change the core plan/apply ownership semantics or target-native file layout.
- Add new runtime behavior unrelated to install, help, versioning, or onboarding.

## Decisions

### 1. Use the repository root as the public CLI entrypoint

The published module path will become `github.com/akhmanov/agentspec`, and the repository root will host the installable `main` package.

This is the simplest way to support the desired install command: `go install github.com/akhmanov/agentspec@latest`. It also makes the root README, root build commands, and public module path line up cleanly.

Alternatives considered:
- Keep `cmd/agentspec` as the canonical install target. Rejected because it exposes repo internals in the public install path.
- Support both a root entrypoint and `cmd/agentspec`. Rejected because it adds redundant surface area for a single-binary CLI without solving a real user problem.

### 2. Keep the command wiring minimal during the move

The change should relocate the existing CLI entrypoint rather than use the publication work as an excuse to add a new abstraction layer. Existing command construction can stay in the `main` package unless a tiny extraction is required to keep tests readable.

Alternatives considered:
- Extract a new reusable CLI application package. Rejected as unnecessary abstraction for a small CLI.

### 3. Expose version metadata through flags only

The root command will support `-v` and `--version`. Both will print a single-line `agentspec <version>` string and exit successfully without requiring a subcommand. The version value will default to `dev` and remain build-injectable for tagged releases.

Alternatives considered:
- Add a `version` subcommand. Rejected because it widens the surface without improving the core install-and-verify flow.
- Hardcode a version string in source. Rejected because publication needs a release-injected value.

### 4. Prefer explicit help text over new discovery commands

The CLI help output should directly explain execution-context defaults, supported targets, repeatability, and verbose behavior. Invalid-target errors should also list supported values. This keeps discovery close to the existing interface instead of inventing new commands for basic explanation.

Alternatives considered:
- Add new commands purely for discovery, such as a target-listing command. Rejected because help and errors are sufficient for this scope.

### 5. Deduplicate repeated `--target` values while preserving first-seen order

Repeated target values should be normalized once before plan/apply execution. This makes the CLI more forgiving in shell usage and avoids duplicate output groups without changing multi-target semantics.

Alternatives considered:
- Treat duplicate targets as an error. Rejected because it makes the interface stricter without user benefit.
- Preserve duplicates verbatim. Rejected because it produces confusing repeated output and repeated work.

### 6. Make `init` write a commented scaffold, not an opinionated example

`agentspec init` should stay minimal but explain what the four top-level resource maps represent. Comments give first-time users guidance without turning `init` into a generator with repo-specific defaults.

Alternatives considered:
- Keep the current empty skeleton. Rejected because it is valid but not instructive.
- Generate a pre-filled example config. Rejected because it pushes opinionated content into the starter file.

### 7. Document `AGENTSPEC_GIT_TIMEOUT` as an advanced README knob, not a CLI flag

`AGENTSPEC_GIT_TIMEOUT` affects only internal git subprocesses used for GitHub-backed skill bundle resolution. It is a real operator control, but it is too narrow to deserve a first-class CLI flag. The right place for it is an advanced environment-variable section in the root README.

Alternatives considered:
- Promote it to a normal CLI flag. Rejected because it overexposes an implementation detail.
- Leave it undocumented. Rejected because publication should not leave a real operational knob half-hidden.

### 8. Keep the committed examples as smoke-validation assets

The committed `example/local-smoke` and `example/github-smoke` workspaces remain valuable, but they should support validation rather than carry the burden of first-run onboarding. The root README becomes the primary user path; the examples stay as explicit smoke flows.

## Risks / Trade-offs

- [Contributor paths change from `cmd/agentspec`] -> Update tests, example READMEs, and any repo-local invocation guidance in the same change so build and smoke flows stay coherent.
- [Root install contract can drift from the module layout] -> Keep the root package installable and verify it with build and smoke checks as part of the change.
- [Commented starter output changes exact test expectations] -> Update tests to assert the new annotated starter content deliberately rather than treating the previous empty scaffold as canonical.
- [Version output can drift from release metadata] -> Keep a `dev` fallback for local builds and document the build-injected release path.

## Migration Plan

1. Update the module path and public CLI entrypoint so the root package becomes installable.
2. Adjust CLI tests, example READMEs, and smoke assumptions to follow the root entrypoint.
3. Add version/help/init/target-polish behavior and the root README.
4. Verify the final public surface with tests, build, and binary smoke checks.

## Open Questions

No open design questions remain for this slice.
