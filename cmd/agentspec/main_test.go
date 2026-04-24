package main

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCommandIncludesInitPlanAndApply(t *testing.T) {
	cmd := newCommand()

	if cmd.Name != "agentspec" {
		t.Fatalf("got name %q, want %q", cmd.Name, "agentspec")
	}

	if len(cmd.Commands) != 3 {
		t.Fatalf("got %d commands, want %d", len(cmd.Commands), 3)
	}

	if cmd.Commands[0].Name != "init" {
		t.Fatalf("got first command %q, want %q", cmd.Commands[0].Name, "init")
	}

	if cmd.Commands[1].Name != "plan" {
		t.Fatalf("got second command %q, want %q", cmd.Commands[1].Name, "plan")
	}

	if cmd.Commands[2].Name != "apply" {
		t.Fatalf("got third command %q, want %q", cmd.Commands[2].Name, "apply")
	}
}

func TestInitWritesStarterConfig(t *testing.T) {
	dir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatal(err)
		}
	}()

	err = newCommand().Run(context.Background(), []string{"agentspec", "init"})
	if err != nil {
		t.Fatalf("run init: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(dir, "agentspec.yaml"))
	if err != nil {
		t.Fatalf("read agentspec.yaml: %v", err)
	}

	want := "sections: {}\ncommands: {}\nagents: {}\nskills: {}\n"
	if string(raw) != want {
		t.Fatalf("got config %q, want %q", string(raw), want)
	}
}

func TestFreshWorkspaceSmokePath(t *testing.T) {
	dir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join(dir, ".agentspec", "commands"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".agentspec", "commands", "explore.md"), []byte("Explore\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	raw := []byte("sections:\n  core:\n    inline: |\n      Core rules\ncommands:\n  explore:\n    path: ./.agentspec/commands/explore.md\nagents: {}\nskills: {}\n")
	if err := os.WriteFile(filepath.Join(dir, "agentspec.yaml"), raw, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatal(err)
		}
	}()

	out, err := runCommand(t, dir, []string{"agentspec", "plan", "--opencode"})
	if err != nil {
		t.Fatalf("run plan: %v", err)
	}
	if !strings.Contains(out, "create .opencode/commands/explore.md") {
		t.Fatalf("got plan output %q", out)
	}
	if !strings.Contains(out, "create AGENTS.md#core") {
		t.Fatalf("got plan output %q", out)
	}
	if _, err := os.Stat(filepath.Join(dir, ".opencode", "commands", "explore.md")); !os.IsNotExist(err) {
		t.Fatalf("expected no command output after plan, got err %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "AGENTS.md")); !os.IsNotExist(err) {
		t.Fatalf("expected no AGENTS.md after plan, got err %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".agentspec", "state", "opencode.json")); !os.IsNotExist(err) {
		t.Fatalf("expected no state after plan, got err %v", err)
	}

	_, err = runCommand(t, dir, []string{"agentspec", "apply", "--opencode"})
	if err != nil {
		t.Fatalf("run apply: %v", err)
	}

	cmd, err := os.ReadFile(filepath.Join(dir, ".opencode", "commands", "explore.md"))
	if err != nil {
		t.Fatalf("read command output: %v", err)
	}
	if string(cmd) != "Explore\n" {
		t.Fatalf("got command output %q, want %q", string(cmd), "Explore\n")
	}

	agents, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	text := string(agents)
	if !strings.Contains(text, "<!-- agentspec:section:start core -->") {
		t.Fatalf("missing agentspec section start marker in %q", text)
	}
	if !strings.Contains(text, "<!-- agentspec:section:end core -->") {
		t.Fatalf("missing agentspec section end marker in %q", text)
	}
	if _, err := os.Stat(filepath.Join(dir, ".agentspec", "state", "opencode.json")); err != nil {
		t.Fatalf("stat agentspec state file: %v", err)
	}

	out, err = runCommand(t, dir, []string{"agentspec", "plan", "--opencode"})
	if err != nil {
		t.Fatalf("run second plan: %v", err)
	}
	if !strings.Contains(out, "No managed changes.") {
		t.Fatalf("got second plan output %q", out)
	}
}

func runCommand(t *testing.T, dir string, args []string) (string, error) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatal(err)
		}
	}()

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	defer func() {
		os.Stdout = old
	}()

	runErr := newCommand().Run(context.Background(), args)
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}

	return string(raw), runErr
}
