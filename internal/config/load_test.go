package config

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func TestLoadParsesInlineAndPathResources(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "agentspec.yaml")

	raw := []byte("sections:\n  core:\n    inline: |\n      Core rules\ncommands:\n  explore:\n    path: ./.agentspec/commands/explore.md\nagents: {}\nskills:\n  debug:\n    path: ./.agentspec/skills/debug\n")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if got := cfg.Sections["core"].Inline; got != "Core rules\n" {
		t.Fatalf("got inline %q, want %q", got, "Core rules\n")
	}

	if got := cfg.Commands["explore"].Path; got != "./.agentspec/commands/explore.md" {
		t.Fatalf("got command path %q, want %q", got, "./.agentspec/commands/explore.md")
	}

	if got := cfg.Skills["debug"].Path; got != "./.agentspec/skills/debug" {
		t.Fatalf("got skill path %q, want %q", got, "./.agentspec/skills/debug")
	}
}

func TestLoadRejectsMultipleSelectors(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "agentspec.yaml")

	raw := []byte("sections:\n  core:\n    inline: hi\n    path: ./core.md\ncommands: {}\nagents: {}\nskills: {}\n")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "sections.core") {
		t.Fatalf("got error %q, want resource id", err)
	}
}

func TestLoadRejectsUnsupportedGitLabSelector(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "agentspec.yaml")

	raw := []byte("sections:\n  core:\n    gitlab:\n      repo: group/repo\n      ref: main\n      path: sections/core.md\ncommands: {}\nagents: {}\nskills: {}\n")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "gitlab") {
		t.Fatalf("got error %q, want selector name", err)
	}
}

func TestLoadAcceptsHTTPAndGitHubSelectors(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "agentspec.yaml")

	raw := []byte("sections:\n  core:\n    http: https://example.com/core.md\ncommands:\n  explore:\n    github:\n      repo: owner/repo\n      ref: main\n      path: commands/explore.md\nagents: {}\nskills: {}\n")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if got := cfg.Sections["core"].HTTP; got != "https://example.com/core.md" {
		t.Fatalf("got http %q, want %q", got, "https://example.com/core.md")
	}

	if cfg.Commands["explore"].GitHub == nil {
		t.Fatal("expected github selector, got nil")
	}
}

func TestLoadRejectsNonHTTPSHTTPSelector(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "agentspec.yaml")

	raw := []byte("sections:\n  core:\n    http: http://example.com/core.md\ncommands: {}\nagents: {}\nskills: {}\n")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "https") {
		t.Fatalf("got error %q, want https validation", err)
	}
}

func TestLoadRejectsIncompleteGitHubSelector(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "agentspec.yaml")

	raw := []byte("sections: {}\ncommands:\n  explore:\n    github:\n      repo: owner/repo\n      ref: main\nagents: {}\nskills: {}\n")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "github") || !strings.Contains(err.Error(), "path") {
		t.Fatalf("got error %q, want github path validation", err)
	}
}

func TestLoadRejectsUnsafeGitHubSelectorPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "agentspec.yaml")

	raw := []byte("sections: {}\ncommands:\n  explore:\n    github:\n      repo: owner/repo\n      ref: main\n      path: ../commands/explore.md\nagents: {}\nskills: {}\n")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "unsafe") {
		t.Fatalf("got error %q, want unsafe path validation", err)
	}
}

func TestLoadPreservesSectionOrder(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "agentspec.yaml")

	raw := []byte("sections:\n  first:\n    inline: one\n  second:\n    inline: two\ncommands: {}\nagents: {}\nskills: {}\n")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	want := []string{"first", "second"}
	if !reflect.DeepEqual(cfg.SectionOrder, want) {
		t.Fatalf("got section order %v, want %v", cfg.SectionOrder, want)
	}
}

func TestLoadRejectsUnsafeResourceID(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "agentspec.yaml")

	raw := []byte("sections: {}\ncommands:\n  ../explore:\n    inline: hi\nagents: {}\nskills: {}\n")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "../explore") {
		t.Fatalf("got error %q, want unsafe id context", err)
	}
	if !strings.Contains(err.Error(), "invalid resource id") {
		t.Fatalf("got error %q, want invalid resource id", err)
	}
}

func TestLoadParsesCommittedGitHubSmokeExample(t *testing.T) {
	root := repoRoot(t)
	readmePath := filepath.Join(root, "example", "github-smoke", "README.md")
	readme, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("read github smoke README: %v", err)
	}
	if !strings.Contains(string(readme), "upstream") {
		t.Fatalf("expected upstream note in README, got %q", string(readme))
	}

	cfg, err := Load(filepath.Join(root, "example", "github-smoke", "agentspec.yaml"))
	if err != nil {
		t.Fatalf("load github smoke config: %v", err)
	}

	section := cfg.Sections["humanizer-readme"]
	if section.GitHub == nil {
		t.Fatal("expected GitHub-backed file resource for humanizer-readme")
	}
	if section.GitHub.Repo == "" {
		t.Fatal("expected section GitHub repo")
	}
	if !regexp.MustCompile(`^[0-9a-f]{40}$`).MatchString(section.GitHub.Ref) {
		t.Fatalf("got section ref %q, want pinned commit sha", section.GitHub.Ref)
	}
	if got := section.GitHub.Path; got != "README.md" {
		t.Fatalf("got section path %q, want %q", got, "README.md")
	}

	skill := cfg.Skills["dev-browser"]
	if skill.GitHub == nil {
		t.Fatal("expected GitHub-backed skill bundle for dev-browser")
	}
	if skill.GitHub.Repo == "" {
		t.Fatal("expected skill GitHub repo")
	}
	if !regexp.MustCompile(`^[0-9a-f]{40}$`).MatchString(skill.GitHub.Ref) {
		t.Fatalf("got skill ref %q, want pinned commit sha", skill.GitHub.Ref)
	}
	if got := skill.GitHub.Path; got != "skills/dev-browser" {
		t.Fatalf("got skill path %q, want %q", got, "skills/dev-browser")
	}
	if strings.HasSuffix(skill.GitHub.Path, "SKILL.md") {
		t.Fatalf("expected directory-backed skill path, got %q", skill.GitHub.Path)
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
