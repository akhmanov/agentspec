# GitHub Smoke Example

This example is a live smoke: it resolves pinned public GitHub refs, but it still depends on upstream availability.

It also requires a local `git` binary in `PATH` because the directory-backed skill bundle is resolved through a shallow Git fetch.

Pinned sources in `agentspec.yaml`:

- `humanizer-readme` is a GitHub-backed file resource from `blader/humanizer`
- `dev-browser` is a directory-backed skill bundle from `SawyerHood/dev-browser`

Run the smoke from a temp copy or a fresh checkout of the repository so `apply` does not mutate the committed fixture in place. This is a live validation flow, not the primary installation path. From this example directory:

- `cd ../.. && cp -R . /tmp/agentspec-github-smoke-repo && cd /tmp/agentspec-github-smoke-repo/example/github-smoke`
- `go run ../.. plan --target opencode`
- `go run ../.. plan --verbose --target opencode`
- `go run ../.. apply --target opencode`

Upstream drift expectations:

- the refs are pinned, so successful fetches should stay stable for those commits
- the smoke can still fail if GitHub is unavailable or an upstream repository becomes unreachable
- if those pinned sources stop resolving, update the example and its README together
