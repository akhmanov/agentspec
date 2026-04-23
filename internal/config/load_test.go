package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLoadParsesInlineAndPathResources(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "aw.yaml")

	raw := []byte("sections:\n  core:\n    inline: |\n      Core rules\ncommands:\n  explore:\n    path: ./.aw/commands/explore.md\nagents: {}\nskills:\n  debug:\n    path: ./.aw/skills/debug\n")
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

	if got := cfg.Commands["explore"].Path; got != "./.aw/commands/explore.md" {
		t.Fatalf("got command path %q, want %q", got, "./.aw/commands/explore.md")
	}

	if got := cfg.Skills["debug"].Path; got != "./.aw/skills/debug" {
		t.Fatalf("got skill path %q, want %q", got, "./.aw/skills/debug")
	}
}

func TestLoadRejectsMultipleSelectors(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "aw.yaml")

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
	path := filepath.Join(dir, "aw.yaml")

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
	path := filepath.Join(dir, "aw.yaml")

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
	path := filepath.Join(dir, "aw.yaml")

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
	path := filepath.Join(dir, "aw.yaml")

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
	path := filepath.Join(dir, "aw.yaml")

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
	path := filepath.Join(dir, "aw.yaml")

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
	path := filepath.Join(dir, "aw.yaml")

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
