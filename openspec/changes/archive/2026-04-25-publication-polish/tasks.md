## 1. Root Install Surface

- [x] 1.1 Update the module path to `github.com/akhmanov/agentspec`.
- [x] 1.2 Move the installable CLI entrypoint from `cmd/agentspec` to the repository root while keeping the existing command behavior intact.
- [x] 1.3 Update CLI tests and any repo-local invocation paths that currently assume `cmd/agentspec`.

## 2. CLI Polish

- [x] 2.1 Add build-injected version metadata with a `dev` fallback and expose it through `-v` and `--version`.
- [x] 2.2 Expand root and `plan` help text to explain execution-context fallbacks, supported targets, repeatability, and verbose output.
- [x] 2.3 Improve invalid-target errors and deduplicate repeated `--target` values while preserving first-seen order.
- [x] 2.4 Change `agentspec init` to write an annotated starter config.
- [x] 2.5 Add or update tests for version flags, help text, duplicate-target normalization, invalid-target errors, and annotated starter output.

## 3. Repository Onboarding

- [x] 3.1 Add a root `README.md` with product summary, canonical `go install` command, version verification, quickstart, supported targets, ownership model, and reference links.
- [x] 3.2 Reposition committed examples as smoke-validation flows and align their documentation with the root entrypoint.
- [x] 3.3 Document `AGENTSPEC_GIT_TIMEOUT` as an advanced environment variable in the root README.

## 4. Validation

- [x] 4.1 Run focused CLI tests covering the new install/help/version/onboarding behavior.
- [x] 4.2 Run `go test ./...`, `go test -race ./...`, `go vet ./...`, and `go build ./...`.
- [x] 4.3 Smoke test the built binary for `--version`, `--help`, `init`, and `plan/apply` on at least the local smoke example.
- [x] 4.4 Smoke test at least one `claude-code` target path so the documented supported-target list is backed by a real CLI run.
