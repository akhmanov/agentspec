package resolve

import (
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/akhmanov/agentspec/internal/config"
	"github.com/akhmanov/agentspec/internal/model"
)

func TestNewRemoteLoaderIgnoresGitHubBaseURLEnvVars(t *testing.T) {
	t.Setenv("AGENTSPEC_GITHUB_RAW_BASE_URL", "https://example.invalid/raw")

	loader := newRemoteLoader(nil)

	if loader.githubRawBaseURL != "https://raw.githubusercontent.com" {
		t.Fatalf("got github raw base url %q", loader.githubRawBaseURL)
	}
}

func TestResolveLoadsHTTPDocumentsAndSkill(t *testing.T) {
	server := newRemoteTestServer(t, map[string]string{
		"/http/sections/core.md":    "Core rules\n",
		"/http/commands/explore.md": "Explore\n",
		"/http/agents/reviewer.md":  "Review\n",
		"/http/skills/debug.md":     "---\nname: debug\ndescription: Debug skill\n---\n\n# Debug\n",
	}, nil)
	defer server.Close()

	loader := newRemoteLoader(server.Client())

	cfg := &config.Config{
		Sections:     map[string]config.Source{"core": {HTTP: server.URL + "/http/sections/core.md"}},
		Commands:     map[string]config.Source{"explore": {HTTP: server.URL + "/http/commands/explore.md"}},
		Agents:       map[string]config.Source{"reviewer": {HTTP: server.URL + "/http/agents/reviewer.md"}},
		Skills:       map[string]config.Source{"debug": {HTTP: server.URL + "/http/skills/debug.md"}},
		SectionOrder: []string{"core"},
		CommandOrder: []string{"explore"},
		AgentOrder:   []string{"reviewer"},
		SkillOrder:   []string{"debug"},
	}

	res, err := resolveWithLoader(t.TempDir(), cfg, loader)
	if err != nil {
		t.Fatalf("resolve remote config: %v", err)
	}

	if got := res.Sections[0].Body; got != "Core rules\n" {
		t.Fatalf("got section %q, want %q", got, "Core rules\n")
	}
	if got := res.Commands[0].Body; got != "Explore\n" {
		t.Fatalf("got command %q, want %q", got, "Explore\n")
	}
	if got := res.Agents[0].Body; got != "Review\n" {
		t.Fatalf("got agent %q, want %q", got, "Review\n")
	}
	if got := res.Skills[0].Files[0].Path; got != "SKILL.md" {
		t.Fatalf("got skill path %q, want %q", got, "SKILL.md")
	}
}

func TestResolveReportsHTTPFetchFailure(t *testing.T) {
	server := newRemoteTestServer(t, map[string]string{}, nil)
	defer server.Close()

	loader := newRemoteLoader(server.Client())

	cfg := &config.Config{
		Sections:     map[string]config.Source{"core": {HTTP: server.URL + "/http/missing.md"}},
		Commands:     map[string]config.Source{},
		Agents:       map[string]config.Source{},
		Skills:       map[string]config.Source{},
		SectionOrder: []string{"core"},
	}

	_, err := resolveWithLoader(t.TempDir(), cfg, loader)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "unexpected status") {
		t.Fatalf("got error %q, want fetch failure", err)
	}
}

func TestResolveLoadsGitHubFileResourcesAndSingleFileSkill(t *testing.T) {
	server := newRemoteTestServer(t, map[string]string{
		"/raw/owner/repo/main/sections/core.md":    "Core rules\n",
		"/raw/owner/repo/main/commands/explore.md": "Explore\n",
		"/raw/owner/repo/main/agents/reviewer.md":  "Review\n",
		"/raw/owner/repo/main/skills/debug.md":     "---\nname: debug\ndescription: Debug skill\n---\n\n# Debug\n",
	}, nil)
	defer server.Close()

	loader := newRemoteLoader(server.Client())
	loader.githubRawBaseURL = server.URL + "/raw"

	cfg := &config.Config{
		Sections:     map[string]config.Source{"core": {GitHub: &config.GitSource{Repo: "owner/repo", Ref: "main", Path: "sections/core.md"}}},
		Commands:     map[string]config.Source{"explore": {GitHub: &config.GitSource{Repo: "owner/repo", Ref: "main", Path: "commands/explore.md"}}},
		Agents:       map[string]config.Source{"reviewer": {GitHub: &config.GitSource{Repo: "owner/repo", Ref: "main", Path: "agents/reviewer.md"}}},
		Skills:       map[string]config.Source{"debug": {GitHub: &config.GitSource{Repo: "owner/repo", Ref: "main", Path: "skills/debug.md"}}},
		SectionOrder: []string{"core"},
		CommandOrder: []string{"explore"},
		AgentOrder:   []string{"reviewer"},
		SkillOrder:   []string{"debug"},
	}

	res, err := resolveWithLoader(t.TempDir(), cfg, loader)
	if err != nil {
		t.Fatalf("resolve github config: %v", err)
	}

	if got := res.Commands[0].Body; got != "Explore\n" {
		t.Fatalf("got command %q, want %q", got, "Explore\n")
	}
	if got := res.Skills[0].Files[0].Path; got != "SKILL.md" {
		t.Fatalf("got skill path %q, want %q", got, "SKILL.md")
	}
}

func TestResolveLoadsGitHubDirectorySkillBundle(t *testing.T) {
	logPath := installFakeGit(t, "bundle")

	var rawRequests int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/raw/owner/repo/main/skills/debug":
			atomic.AddInt32(&rawRequests, 1)
			http.NotFound(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	loader := newRemoteLoader(server.Client())
	loader.githubRawBaseURL = server.URL + "/raw"

	cfg := &config.Config{
		Sections:   map[string]config.Source{},
		Commands:   map[string]config.Source{},
		Agents:     map[string]config.Source{},
		Skills:     map[string]config.Source{"debug": {GitHub: &config.GitSource{Repo: "owner/repo", Ref: "main", Path: "skills/debug"}}},
		SkillOrder: []string{"debug"},
	}

	res, err := resolveWithLoader(t.TempDir(), cfg, loader)
	if err != nil {
		t.Fatalf("resolve github bundle: %v", err)
	}

	if len(res.Skills) != 1 || len(res.Skills[0].Files) != 2 {
		t.Fatalf("got skills %#v", res.Skills)
	}
	if got := atomic.LoadInt32(&rawRequests); got != 1 {
		t.Fatalf("got %d raw requests, want 1", got)
	}
	if got := res.Skills[0].Files[1].Path; got != filepath.Join("notes", "guide.md") {
		t.Fatalf("got bundle path %q, want %q", got, filepath.Join("notes", "guide.md"))
	}

	logBody, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read fake git log: %v", err)
	}
	if !strings.Contains(string(logBody), "--depth 1") {
		t.Fatalf("git log %q missing shallow fetch/clone flag", string(logBody))
	}
	if !strings.Contains(string(logBody), "--filter=blob:none") {
		t.Fatalf("git log %q missing filtered fetch flag", string(logBody))
	}
	if !strings.Contains(string(logBody), "lfs=1") {
		t.Fatalf("git log %q missing GIT_LFS_SKIP_SMUDGE=1", string(logBody))
	}
	if !strings.Contains(string(logBody), "cmd=ls-tree") {
		t.Fatalf("git log %q missing path inspection", string(logBody))
	}
	if !strings.Contains(string(logBody), "cmd=cat-file") {
		t.Fatalf("git log %q missing blob reads", string(logBody))
	}
	if strings.Contains(string(logBody), "cmd=checkout") {
		t.Fatalf("git log %q should not materialize worktree checkout", string(logBody))
	}
}

func TestResolveLoadsGitHubDirectorySkillBundleWithDottedPath(t *testing.T) {
	logPath := installFakeGit(t, "bundle")
	t.Setenv("AGENTSPEC_GIT_BUNDLE_DIR", "skills/debug.v1")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/raw/owner/repo/main/skills/debug.v1":
			http.NotFound(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	loader := newRemoteLoader(server.Client())
	loader.githubRawBaseURL = server.URL + "/raw"

	cfg := &config.Config{
		Sections:   map[string]config.Source{},
		Commands:   map[string]config.Source{},
		Agents:     map[string]config.Source{},
		Skills:     map[string]config.Source{"debug": {GitHub: &config.GitSource{Repo: "owner/repo", Ref: "main", Path: "skills/debug.v1"}}},
		SkillOrder: []string{"debug"},
	}

	res, err := resolveWithLoader(t.TempDir(), cfg, loader)
	if err != nil {
		t.Fatalf("resolve github dotted bundle: %v", err)
	}
	if len(res.Skills) != 1 || len(res.Skills[0].Files) != 2 {
		t.Fatalf("got skills %#v", res.Skills)
	}
	if _, err := os.Stat(logPath); err != nil {
		t.Fatalf("expected git fallback log, stat err = %v", err)
	}
}

func TestResolveLoadsGitHubDirectorySkillBundleWithMarkdownPath(t *testing.T) {
	logPath := installFakeGit(t, "bundle")
	t.Setenv("AGENTSPEC_GIT_BUNDLE_DIR", "skills/debug.md")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/raw/owner/repo/main/skills/debug.md":
			http.NotFound(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	loader := newRemoteLoader(server.Client())
	loader.githubRawBaseURL = server.URL + "/raw"

	cfg := &config.Config{
		Sections:   map[string]config.Source{},
		Commands:   map[string]config.Source{},
		Agents:     map[string]config.Source{},
		Skills:     map[string]config.Source{"debug": {GitHub: &config.GitSource{Repo: "owner/repo", Ref: "main", Path: "skills/debug.md"}}},
		SkillOrder: []string{"debug"},
	}

	res, err := resolveWithLoader(t.TempDir(), cfg, loader)
	if err != nil {
		t.Fatalf("resolve github markdown bundle: %v", err)
	}
	if len(res.Skills) != 1 || len(res.Skills[0].Files) != 2 {
		t.Fatalf("got skills %#v", res.Skills)
	}
	logBody, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read fake git log: %v", err)
	}
	if !strings.Contains(string(logBody), "cmd=ls-tree") {
		t.Fatalf("git log %q missing path inspection", string(logBody))
	}
}

func TestResolveMissingGitHubSingleFileSkillDoesNotFallBackToGitBundle(t *testing.T) {
	logPath := installFakeGit(t, "file")

	server := newRemoteTestServer(t, nil, nil)
	defer server.Close()

	loader := newRemoteLoader(server.Client())
	loader.githubRawBaseURL = server.URL + "/raw"

	cfg := &config.Config{
		Sections:   map[string]config.Source{},
		Commands:   map[string]config.Source{},
		Agents:     map[string]config.Source{},
		Skills:     map[string]config.Source{"debug": {GitHub: &config.GitSource{Repo: "owner/repo", Ref: "main", Path: "skills/debug.md"}}},
		SkillOrder: []string{"debug"},
	}

	_, err := resolveWithLoader(t.TempDir(), cfg, loader)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "unexpected status 404 Not Found") {
		t.Fatalf("got error %q, want raw 404", err)
	}
	logBody, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read fake git log: %v", err)
	}
	if !strings.Contains(string(logBody), "cmd=ls-tree") {
		t.Fatalf("git log %q missing path inspection", string(logBody))
	}
	if strings.Contains(string(logBody), "cmd=cat-file") {
		t.Fatalf("git log %q should not read bundle blobs for file path", string(logBody))
	}
}

func TestResolveRejectsGitHubBundleWithoutRootSkillFile(t *testing.T) {
	installFakeGit(t, "missing-root")

	server := newRemoteTestServer(t, nil, nil)
	defer server.Close()

	loader := newRemoteLoader(server.Client())
	loader.githubRawBaseURL = server.URL + "/raw"

	cfg := &config.Config{
		Sections:   map[string]config.Source{},
		Commands:   map[string]config.Source{},
		Agents:     map[string]config.Source{},
		Skills:     map[string]config.Source{"debug": {GitHub: &config.GitSource{Repo: "owner/repo", Ref: "main", Path: "skills/debug"}}},
		SkillOrder: []string{"debug"},
	}

	_, err := resolveWithLoader(t.TempDir(), cfg, loader)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "missing root SKILL.md") {
		t.Fatalf("got error %q, want missing root SKILL.md", err)
	}
}

func TestValidateSkillFilesRejectsUnsafeBundlePath(t *testing.T) {
	err := validateSkillFiles([]model.File{
		{Path: "SKILL.md", Body: "---\nname: debug\ndescription: Debug skill\n---\n\n# Debug\n"},
		{Path: filepath.Join("..", "escape.txt"), Body: "nope\n"},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "unsafe bundle path") {
		t.Fatalf("got error %q, want unsafe bundle path", err)
	}
}

func newRemoteTestServer(t *testing.T, text map[string]string, blobs map[string][]byte) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if body, ok := blobs[r.URL.Path]; ok {
			_, _ = w.Write(body)
			return
		}
		if body, ok := text[r.URL.Path]; ok {
			_, _ = w.Write([]byte(body))
			return
		}
		http.NotFound(w, r)
	}))
}

func installFakeGit(t *testing.T, scenario string) string {
	t.Helper()

	binDir := t.TempDir()
	logPath := filepath.Join(binDir, "git.log")
	gitPath := filepath.Join(binDir, "git")
	script := `#!/bin/sh
set -eu

bundle_dir() {
	printf '%s' "${AGENTSPEC_GIT_BUNDLE_DIR:-skills/debug}"
}

requested_path() {
	last=""
	for arg in "$@"; do
		last="$arg"
	done
	last=${last#':(literal)'}
	printf '%s' "$last"
}

write_valid_bundle() {
	bundle_dir=$(bundle_dir)
	mkdir -p "$1/$bundle_dir/notes"
	cat > "$1/$bundle_dir/SKILL.md" <<'EOF'
---
name: debug
description: Debug skill
---

# Debug
EOF
	cat > "$1/$bundle_dir/notes/guide.md" <<'EOF'
Guide
EOF
}

write_missing_root_bundle() {
	bundle_dir=$(bundle_dir)
	mkdir -p "$1/$bundle_dir/notes"
	cat > "$1/$bundle_dir/notes/guide.md" <<'EOF'
Guide
EOF
}

print_entry() {
	printf '%s %s %s\t%s\0' "$1" "$2" "$3" "$4"
}

ls_tree_single() {
	path="$1"
	case "${AGENTSPEC_GIT_SCENARIO:-bundle}" in
		bundle|missing-root)
			if [ "$path" = "$(bundle_dir)" ]; then
				print_entry 040000 tree tree-obj "$path"
			fi
			;;
		file)
			file_path="${AGENTSPEC_GIT_FILE_PATH:-skills/debug.md}"
			if [ "$path" = "$file_path" ]; then
				print_entry 100644 blob file-obj "$path"
			fi
			;;
		esac
}

ls_tree_recursive() {
	path="$1"
	bundle_dir=$(bundle_dir)
	if [ "$path" != "$bundle_dir" ]; then
		return
	fi

	case "${AGENTSPEC_GIT_SCENARIO:-bundle}" in
		bundle)
			print_entry 100644 blob skill-obj "$bundle_dir/SKILL.md"
			print_entry 100644 blob guide-obj "$bundle_dir/notes/guide.md"
			;;
		missing-root)
			print_entry 100644 blob guide-obj "$bundle_dir/notes/guide.md"
			;;
		esac
}

blob_body() {
	case "$1" in
		skill-obj|file-obj)
			cat <<'EOF'
---
name: debug
description: Debug skill
---

# Debug
EOF
			;;
		guide-obj)
			cat <<'EOF'
Guide
EOF
			;;
		esac
}

write_bundle() {
	case "${AGENTSPEC_GIT_SCENARIO:-bundle}" in
		bundle)
			write_valid_bundle "$1"
			;;
		missing-root)
			write_missing_root_bundle "$1"
			;;
		esac
}

require_fetch_flags() {
	case " $* " in
		*" --depth 1 "*) ;;
		*)
			echo "missing --depth 1" >&2
			exit 91
			;;
	esac
	case "${GIT_LFS_SKIP_SMUDGE:-}" in
		1) ;;
		*)
			echo "missing GIT_LFS_SKIP_SMUDGE=1" >&2
			exit 92
			;;
	esac
	case " $* " in
		*" --filter=blob:none "*) ;;
		*)
			echo "missing --filter=blob:none" >&2
			exit 93
			;;
	esac
}

repo_dir="$PWD"
if [ "${1:-}" = "-C" ]; then
	repo_dir="$2"
	shift 2
fi

cmd="${1:-}"
if [ -n "$cmd" ]; then
	shift
fi

printf 'dir=%s lfs=%s cmd=%s args=%s\n' "$repo_dir" "${GIT_LFS_SKIP_SMUDGE:-}" "$cmd" "$*" >> "$AGENTSPEC_GIT_LOG"

if [ "${AGENTSPEC_GIT_SCENARIO:-}" = "sleep" ]; then
	sleep "${AGENTSPEC_GIT_SLEEP:-1}"
fi

case "$cmd" in
	init)
		target=""
		for arg in "$@"; do
			target="$arg"
		done
		mkdir -p "$target"
		;;
	remote)
		;;
	fetch)
		require_fetch_flags "$*"
		;;
	ls-tree)
		path=$(requested_path "$@")
		recursive=0
		for arg in "$@"; do
			if [ "$arg" = "-r" ]; then
				recursive=1
			fi
		done
		if [ "$recursive" = "1" ]; then
			ls_tree_recursive "$path"
		else
			ls_tree_single "$path"
		fi
		;;
	cat-file)
		blob_body "${2:-}"
		;;
	checkout)
		write_bundle "$repo_dir"
		;;
	esac
`
	if err := os.WriteFile(gitPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake git: %v", err)
	}

	t.Setenv("AGENTSPEC_GIT_LOG", logPath)
	t.Setenv("AGENTSPEC_GIT_SCENARIO", scenario)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	return logPath
}

func TestNewRemoteLoaderSetsDefaults(t *testing.T) {
	loader := newRemoteLoader(nil)

	if loader.client == nil {
		t.Fatal("expected default client, got nil")
	}

	if loader.maxBytes <= 0 {
		t.Fatalf("got max bytes %d, want positive", loader.maxBytes)
	}

	client, ok := loader.client.(*http.Client)
	if !ok {
		t.Fatalf("got client %#v, want *http.Client", loader.client)
	}

	if client.Timeout != defaultHTTPTimeout {
		t.Fatalf("got timeout %v, want %v", client.Timeout, defaultHTTPTimeout)
	}
}

func TestGitCommandDefaultTimeoutExceedsHTTPTimeout(t *testing.T) {
	if defaultGitTimeout <= defaultHTTPTimeout {
		t.Fatalf("got git timeout %v, want greater than http timeout %v", defaultGitTimeout, defaultHTTPTimeout)
	}
}

func TestRemoteLoaderFetchHTTPReadsBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("remote\n"))
	}))
	defer server.Close()

	loader := newRemoteLoader(server.Client())
	body, err := loader.fetchHTTP(server.URL)
	if err != nil {
		t.Fatalf("fetch http: %v", err)
	}

	if string(body) != "remote\n" {
		t.Fatalf("got body %q, want %q", string(body), "remote\n")
	}
}

func TestRemoteLoaderFetchHTTPRejectsOversizedBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("abcdef"))
	}))
	defer server.Close()

	loader := newRemoteLoader(server.Client())
	loader.maxBytes = 4

	_, err := loader.fetchHTTP(server.URL)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "too large") {
		t.Fatalf("got error %q, want size limit", err)
	}
}

func TestRunGitCommandTimesOut(t *testing.T) {
	installFakeGit(t, "sleep")
	t.Setenv("AGENTSPEC_GIT_TIMEOUT", "20ms")
	t.Setenv("AGENTSPEC_GIT_SLEEP", "1")

	start := time.Now()
	err := runGitCommand("", nil, "fetch", "--depth", "1", "origin", "main")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("got error %q, want timeout", err)
	}
	if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
		t.Fatalf("git timeout took %v, want under %v", elapsed, 500*time.Millisecond)
	}
}

func TestResolveUsesConfigDirectoryForRelativeLocalPaths(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "config", "agentspec.yaml")

	if err := os.MkdirAll(filepath.Join(root, "config", "resources"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "config", "resources", "explore.md"), []byte("Explore\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "config", "resources", "debug.md"), []byte("---\nname: debug\ndescription: Debug skill\n---\n\n# Debug\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	raw := []byte("sections: {}\ncommands:\n  explore:\n    path: ./resources/explore.md\nagents: {}\nskills:\n  debug:\n    path: ./resources/debug.md\n")
	if err := os.WriteFile(configPath, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	res, err := Resolve(root, cfg)
	if err != nil {
		t.Fatalf("resolve config: %v", err)
	}

	if got := res.Commands[0].Body; got != "Explore\n" {
		t.Fatalf("got command %q, want %q", got, "Explore\n")
	}
	if got := res.Skills[0].Files[0].Body; got != "---\nname: debug\ndescription: Debug skill\n---\n\n# Debug\n" {
		t.Fatalf("got skill body %q", got)
	}
}

func TestResolvePreservesAbsoluteLocalPaths(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "config", "agentspec.yaml")
	absCommand := filepath.Join(root, "shared", "explore.md")

	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(absCommand), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(absCommand, []byte("Explore\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	raw := []byte("sections: {}\ncommands:\n  explore:\n    path: " + absCommand + "\nagents: {}\nskills: {}\n")
	if err := os.WriteFile(configPath, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	res, err := Resolve(root, cfg)
	if err != nil {
		t.Fatalf("resolve config: %v", err)
	}

	if got := res.Commands[0].Body; got != "Explore\n" {
		t.Fatalf("got command %q, want %q", got, "Explore\n")
	}
}

func TestResolveLoadsInlinePathAndSingleFileSkill(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".agentspec", "commands"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".agentspec", "skills"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, ".agentspec", "commands", "explore.md"), []byte("Explore\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".agentspec", "skills", "debug.md"), []byte("---\nname: debug\ndescription: Debug skill\n---\n\n# Debug\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	raw := []byte("sections:\n  core:\n    inline: |\n      Core rules\ncommands:\n  explore:\n    path: ./.agentspec/commands/explore.md\nagents: {}\nskills:\n  debug:\n    path: ./.agentspec/skills/debug.md\n")
	path := filepath.Join(dir, "agentspec.yaml")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	res, err := Resolve(dir, cfg)
	if err != nil {
		t.Fatalf("resolve config: %v", err)
	}

	if len(res.Sections) != 1 || res.Sections[0].Body != "Core rules\n" {
		t.Fatalf("got sections %#v", res.Sections)
	}

	if len(res.Commands) != 1 || res.Commands[0].Body != "Explore\n" {
		t.Fatalf("got commands %#v", res.Commands)
	}

	if len(res.Skills) != 1 || len(res.Skills[0].Files) != 1 {
		t.Fatalf("got skills %#v", res.Skills)
	}

	if res.Skills[0].Files[0].Path != "SKILL.md" {
		t.Fatalf("got skill path %q, want %q", res.Skills[0].Files[0].Path, "SKILL.md")
	}

	if res.Skills[0].Files[0].Body != "---\nname: debug\ndescription: Debug skill\n---\n\n# Debug\n" {
		t.Fatalf("got skill body %q, want %q", res.Skills[0].Files[0].Body, "---\nname: debug\ndescription: Debug skill\n---\n\n# Debug\n")
	}
}

func TestResolveLoadsDirectorySkillBundle(t *testing.T) {
	dir := t.TempDir()
	skill := filepath.Join(dir, ".agentspec", "skills", "debug")
	if err := os.MkdirAll(filepath.Join(skill, "notes"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(skill, "SKILL.md"), []byte("---\nname: debug\ndescription: Debug skill\n---\n\n# Debug\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skill, "notes", "guide.md"), []byte("Guide\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	raw := []byte("sections: {}\ncommands: {}\nagents: {}\nskills:\n  debug:\n    path: ./.agentspec/skills/debug\n")
	path := filepath.Join(dir, "agentspec.yaml")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	res, err := Resolve(dir, cfg)
	if err != nil {
		t.Fatalf("resolve config: %v", err)
	}

	if len(res.Skills) != 1 || len(res.Skills[0].Files) != 2 {
		t.Fatalf("got skills %#v", res.Skills)
	}

	if res.Skills[0].Files[0].Path != "SKILL.md" {
		t.Fatalf("got first skill file %q, want %q", res.Skills[0].Files[0].Path, "SKILL.md")
	}

	if res.Skills[0].Files[1].Path != filepath.Join("notes", "guide.md") {
		t.Fatalf("got second skill file %q, want %q", res.Skills[0].Files[1].Path, filepath.Join("notes", "guide.md"))
	}
}

func TestResolveRejectsDirectorySkillBundleWithSymlinkEntry(t *testing.T) {
	dir := t.TempDir()
	skill := filepath.Join(dir, ".agentspec", "skills", "debug")
	target := filepath.Join(dir, "guide.md")
	if err := os.MkdirAll(filepath.Join(skill, "notes"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skill, "SKILL.md"), []byte("---\nname: debug\ndescription: Debug skill\n---\n\n# Debug\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("Guide\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(skill, "notes", "guide.md")); err != nil {
		t.Fatal(err)
	}

	raw := []byte("sections: {}\ncommands: {}\nagents: {}\nskills:\n  debug:\n    path: ./.agentspec/skills/debug\n")
	path := filepath.Join(dir, "agentspec.yaml")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	_, err = Resolve(dir, cfg)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "symlink bundle path") {
		t.Fatalf("got error %q, want symlink bundle path", err)
	}
}

func TestResolveRejectsDirectorySkillBundleWithNonRegularEntry(t *testing.T) {
	dir := t.TempDir()
	skill := filepath.Join(dir, ".agentspec", "skills", "debug")
	if err := os.MkdirAll(filepath.Join(skill, "notes"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skill, "SKILL.md"), []byte("---\nname: debug\ndescription: Debug skill\n---\n\n# Debug\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("unix", filepath.Join(skill, "notes", "guide.sock"))
	if err != nil {
		t.Skipf("unix sockets unavailable: %v", err)
	}
	defer listener.Close()

	raw := []byte("sections: {}\ncommands: {}\nagents: {}\nskills:\n  debug:\n    path: ./.agentspec/skills/debug\n")
	path := filepath.Join(dir, "agentspec.yaml")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	_, err = Resolve(dir, cfg)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "non-regular bundle path") {
		t.Fatalf("got error %q, want non-regular bundle path", err)
	}
}

func TestResolveRejectsDirectorySkillWithoutRootSkillFile(t *testing.T) {
	dir := t.TempDir()
	skill := filepath.Join(dir, ".agentspec", "skills", "debug")
	if err := os.MkdirAll(skill, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skill, "guide.md"), []byte("Guide\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	raw := []byte("sections: {}\ncommands: {}\nagents: {}\nskills:\n  debug:\n    path: ./.agentspec/skills/debug\n")
	path := filepath.Join(dir, "agentspec.yaml")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	_, err = Resolve(dir, cfg)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "missing root SKILL.md") {
		t.Fatalf("got error %q, want missing root skill file", err)
	}
}

func TestResolveRejectsInvalidSingleFileSkill(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".agentspec", "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".agentspec", "skills", "debug.md"), []byte("not a skill\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	raw := []byte("sections: {}\ncommands: {}\nagents: {}\nskills:\n  debug:\n    path: ./.agentspec/skills/debug.md\n")
	path := filepath.Join(dir, "agentspec.yaml")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	_, err = Resolve(dir, cfg)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "invalid SKILL.md") {
		t.Fatalf("got error %q, want invalid skill content", err)
	}
}

func TestResolveRejectsInvalidBundleRootSkill(t *testing.T) {
	dir := t.TempDir()
	skill := filepath.Join(dir, ".agentspec", "skills", "debug")
	if err := os.MkdirAll(skill, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skill, "SKILL.md"), []byte("---\nname: debug\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	raw := []byte("sections: {}\ncommands: {}\nagents: {}\nskills:\n  debug:\n    path: ./.agentspec/skills/debug\n")
	path := filepath.Join(dir, "agentspec.yaml")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	_, err = Resolve(dir, cfg)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "invalid SKILL.md") {
		t.Fatalf("got error %q, want invalid skill content", err)
	}
}

func TestResolveAcceptsCRLFSingleFileSkill(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".agentspec", "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	body := "---\r\nname: debug\r\ndescription: Debug skill\r\n---\r\n\r\n# Debug\r\n"
	if err := os.WriteFile(filepath.Join(dir, ".agentspec", "skills", "debug.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	raw := []byte("sections: {}\ncommands: {}\nagents: {}\nskills:\n  debug:\n    path: ./.agentspec/skills/debug.md\n")
	path := filepath.Join(dir, "agentspec.yaml")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	res, err := Resolve(dir, cfg)
	if err != nil {
		t.Fatalf("resolve config: %v", err)
	}
	if len(res.Skills) != 1 {
		t.Fatalf("got skills %#v", res.Skills)
	}
}

func TestResolveAcceptsCRLFBundleRootSkill(t *testing.T) {
	dir := t.TempDir()
	skill := filepath.Join(dir, ".agentspec", "skills", "debug")
	if err := os.MkdirAll(skill, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "---\r\nname: debug\r\ndescription: Debug skill\r\n---\r\n\r\n# Debug\r\n"
	if err := os.WriteFile(filepath.Join(skill, "SKILL.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	raw := []byte("sections: {}\ncommands: {}\nagents: {}\nskills:\n  debug:\n    path: ./.agentspec/skills/debug\n")
	path := filepath.Join(dir, "agentspec.yaml")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	res, err := Resolve(dir, cfg)
	if err != nil {
		t.Fatalf("resolve config: %v", err)
	}
	if len(res.Skills) != 1 {
		t.Fatalf("got skills %#v", res.Skills)
	}
}
