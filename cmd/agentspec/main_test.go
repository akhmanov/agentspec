package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

var runCommandMu sync.Mutex

func TestCommandIncludesInitPlanAndApply(t *testing.T) {
	cmd := newCommand()

	if cmd.Name != "agentspec" {
		t.Fatalf("got name %q, want %q", cmd.Name, "agentspec")
	}

	want := map[string]bool{
		"init":  false,
		"plan":  false,
		"apply": false,
	}
	if len(cmd.Commands) != len(want) {
		t.Fatalf("got %d commands, want %d", len(cmd.Commands), len(want))
	}

	for _, sub := range cmd.Commands {
		seen, ok := want[sub.Name]
		if !ok {
			t.Fatalf("unexpected command %q", sub.Name)
		}
		if seen {
			t.Fatalf("duplicate command %q", sub.Name)
		}
		want[sub.Name] = true
	}

	for name, seen := range want {
		if !seen {
			t.Fatalf("missing command %q", name)
		}
	}
}

func TestInitWritesStarterConfig(t *testing.T) {
	dir := t.TempDir()
	_, err := runCommand(t, dir, []string{"agentspec", "init"})
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

func TestInitWritesStarterConfigToExplicitRoot(t *testing.T) {
	workRoot := t.TempDir()
	runDir := t.TempDir()

	_, err := runCommand(t, runDir, []string{"agentspec", "--root", workRoot, "init"})
	if err != nil {
		t.Fatalf("run init with root: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(workRoot, "agentspec.yaml"))
	if err != nil {
		t.Fatalf("read rooted agentspec.yaml: %v", err)
	}

	want := "sections: {}\ncommands: {}\nagents: {}\nskills: {}\n"
	if string(raw) != want {
		t.Fatalf("got config %q, want %q", string(raw), want)
	}
}

func TestPlanUsesExplicitRootAndRelativeConfig(t *testing.T) {
	workRoot := t.TempDir()
	runDir := t.TempDir()

	writeWorkspaceFile(t, filepath.Join(workRoot, "config", "dev", "agentspec.yaml"), strings.Join([]string{
		"sections:",
		"  core:",
		"    inline: |",
		"      Core rules",
		"commands:",
		"  explore:",
		"    inline: |",
		"      Explore",
		"agents: {}",
		"skills: {}",
		"",
	}, "\n"))

	out, err := runCommand(t, runDir, []string{"agentspec", "--root", workRoot, "--config", filepath.Join("config", "dev", "agentspec.yaml"), "plan", "--target", "opencode"})
	if err != nil {
		t.Fatalf("run plan with explicit context: %v", err)
	}

	if out != compactGroup("create", commandPath("explore"), "AGENTS.md#core") {
		t.Fatalf("got plan output %q", out)
	}
}

func TestPlanUsesEnvironmentFallbacksForRootAndConfig(t *testing.T) {
	workRoot := t.TempDir()
	runDir := t.TempDir()

	writeWorkspaceFile(t, filepath.Join(workRoot, "ops", "agentspec.yaml"), strings.Join([]string{
		"sections:",
		"  core:",
		"    inline: |",
		"      Core rules",
		"commands:",
		"  explore:",
		"    inline: |",
		"      Explore",
		"agents: {}",
		"skills: {}",
		"",
	}, "\n"))

	t.Setenv("AGENTSPEC_ROOT", workRoot)
	t.Setenv("AGENTSPEC_CONFIG", filepath.Join("ops", "agentspec.yaml"))

	out, err := runCommand(t, runDir, []string{"agentspec", "plan", "--target", "opencode"})
	if err != nil {
		t.Fatalf("run plan with env context: %v", err)
	}

	if out != compactGroup("create", commandPath("explore"), "AGENTS.md#core") {
		t.Fatalf("got plan output %q", out)
	}
}

func TestPlanRequiresAtLeastOneTarget(t *testing.T) {
	dir := t.TempDir()
	writeWorkspaceConfig(t, dir, "sections: {}\ncommands: {}\nagents: {}\nskills: {}\n")

	_, err := runCommand(t, dir, []string{"agentspec", "plan"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "target flag") {
		t.Fatalf("got error %q, want target flag context", err)
	}
}

func TestPlanGroupsMultipleTargets(t *testing.T) {
	dir := t.TempDir()
	writeWorkspaceConfig(t, dir, "sections:\n  core:\n    inline: |\n      Core rules\ncommands:\n  explore:\n    inline: |\n      Explore\nagents: {}\nskills: {}\n")

	out, err := runCommand(t, dir, []string{"agentspec", "plan", "--target", "opencode", "--target", "claude-code"})
	if err != nil {
		t.Fatalf("run multi-target plan: %v", err)
	}

	want := strings.Join([]string{
		"opencode:",
		strings.TrimSuffix(compactGroup("create", filepath.Join(".opencode", "commands", "explore.md"), "AGENTS.md#core"), "\n"),
		"",
		"claude-code:",
		strings.TrimSuffix(compactGroup("create", filepath.Join(".claude", "commands", "explore.md"), "CLAUDE.md#core"), "\n"),
		"",
	}, "\n")
	if out != want {
		t.Fatalf("got grouped output %q, want %q", out, want)
	}
}

func TestApplySupportsMultipleTargetsWithSeparateState(t *testing.T) {
	dir := t.TempDir()
	writeWorkspaceConfig(t, dir, "sections:\n  core:\n    inline: |\n      Core rules\ncommands:\n  explore:\n    inline: |\n      Explore\nagents: {}\nskills: {}\n")

	if _, err := runCommand(t, dir, []string{"agentspec", "apply", "--target", "opencode", "--target", "claude-code"}); err != nil {
		t.Fatalf("run multi-target apply: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".agentspec", "state", "opencode.json")); err != nil {
		t.Fatalf("stat opencode state: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".agentspec", "state", "claude-code.json")); err != nil {
		t.Fatalf("stat claude-code state: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".opencode", "commands", "explore.md")); err != nil {
		t.Fatalf("stat opencode command: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".claude", "commands", "explore.md")); err != nil {
		t.Fatalf("stat claude command: %v", err)
	}
}

func TestApplyStopsOnFirstFailingTargetWithoutRollback(t *testing.T) {
	dir := t.TempDir()
	writeWorkspaceConfig(t, dir, "sections: {}\ncommands:\n  explore:\n    inline: |\n      Explore\nagents: {}\nskills: {}\n")
	writeWorkspaceFile(t, filepath.Join(dir, ".claude", "commands", "explore.md"), "foreign\n")

	_, err := runCommand(t, dir, []string{"agentspec", "apply", "--target", "opencode", "--target", "claude-code"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "foreign file") {
		t.Fatalf("got error %q, want foreign file conflict", err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".opencode", "commands", "explore.md")); err != nil {
		t.Fatalf("stat preserved successful opencode output: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, ".claude", "commands", "explore.md"))
	if err != nil {
		t.Fatalf("read preserved claude file: %v", err)
	}
	if string(raw) != "foreign\n" {
		t.Fatalf("got claude file %q, want %q", string(raw), "foreign\n")
	}
}

func TestFreshWorkspaceSmokePath(t *testing.T) {
	dir := t.TempDir()

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

	out, err := runCommand(t, dir, []string{"agentspec", "plan", "--target", "opencode"})
	if err != nil {
		t.Fatalf("run plan: %v", err)
	}
	if out != compactGroup("create", commandPath("explore"), "AGENTS.md#core") {
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

	_, err = runCommand(t, dir, []string{"agentspec", "apply", "--target", "opencode"})
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

	out, err = runCommand(t, dir, []string{"agentspec", "plan", "--target", "opencode"})
	if err != nil {
		t.Fatalf("run second plan: %v", err)
	}
	if out != "No managed changes.\n" {
		t.Fatalf("got second plan output %q", out)
	}
}

func TestPlanOpencodeGroupsDefaultOutput(t *testing.T) {
	dir := managedWorkspace(t)

	writeWorkspaceFile(t, filepath.Join(dir, ".agentspec", "commands", "build.md"), "Build\n")
	writeWorkspaceFile(t, filepath.Join(dir, ".agentspec", "commands", "explore.md"), "Explore more\n")
	writeWorkspaceConfig(t, dir, "sections: {}\ncommands:\n  build:\n    path: ./.agentspec/commands/build.md\n  explore:\n    path: ./.agentspec/commands/explore.md\nagents: {}\nskills: {}\n")

	out, err := runCommand(t, dir, []string{"agentspec", "plan", "--target", "opencode"})
	if err != nil {
		t.Fatalf("run plan: %v", err)
	}

	want := strings.Join([]string{
		strings.TrimSuffix(compactGroup("create", commandPath("build")), "\n"),
		strings.TrimSuffix(compactGroup("update", commandPath("explore")), "\n"),
		"delete (1): AGENTS.md#core",
		"",
	}, "\n")
	if out != want {
		t.Fatalf("got grouped output %q, want %q", out, want)
	}
}

func TestPlanOpencodeGroupsConflicts(t *testing.T) {
	dir := t.TempDir()

	writeWorkspaceFile(t, filepath.Join(dir, ".agentspec", "commands", "explore.md"), "Explore\n")
	writeWorkspaceConfig(t, dir, "sections: {}\ncommands:\n  explore:\n    path: ./.agentspec/commands/explore.md\nagents: {}\nskills: {}\n")
	writeWorkspaceFile(t, filepath.Join(dir, ".opencode", "commands", "explore.md"), "foreign\n")

	out, err := runCommand(t, dir, []string{"agentspec", "plan", "--target", "opencode"})
	if err != nil {
		t.Fatalf("run plan: %v", err)
	}

	if out != compactGroup("conflict", commandPath("explore")) {
		t.Fatalf("got conflict output %q", out)
	}
}

func TestPlanOpencodeVerboseShowsConflictReasons(t *testing.T) {
	dir := t.TempDir()

	writeWorkspaceFile(t, filepath.Join(dir, ".agentspec", "commands", "explore.md"), "Explore\n")
	writeWorkspaceConfig(t, dir, "sections: {}\ncommands:\n  explore:\n    path: ./.agentspec/commands/explore.md\nagents: {}\nskills: {}\n")
	writeWorkspaceFile(t, filepath.Join(dir, ".opencode", "commands", "explore.md"), "foreign\n")

	out, err := runCommand(t, dir, []string{"agentspec", "plan", "--verbose", "--target", "opencode"})
	if err != nil {
		t.Fatalf("run verbose plan: %v", err)
	}

	want := strings.Join([]string{
		"conflict (1):",
		"  - " + commandPath("explore") + ": refusing to overwrite foreign file",
		"",
	}, "\n")
	if out != want {
		t.Fatalf("got verbose conflict output %q, want %q", out, want)
	}
}

func TestPlanOpencodeVerboseShowsExpandedOutput(t *testing.T) {
	dir := managedWorkspace(t)

	writeWorkspaceFile(t, filepath.Join(dir, ".agentspec", "commands", "build.md"), "Build\n")
	writeWorkspaceFile(t, filepath.Join(dir, ".agentspec", "commands", "explore.md"), "Explore more\n")
	writeWorkspaceConfig(t, dir, "sections: {}\ncommands:\n  build:\n    path: ./.agentspec/commands/build.md\n  explore:\n    path: ./.agentspec/commands/explore.md\nagents: {}\nskills: {}\n")

	out, err := runCommand(t, dir, []string{"agentspec", "plan", "--verbose", "--target", "opencode"})
	if err != nil {
		t.Fatalf("run verbose plan: %v", err)
	}

	want := strings.Join([]string{
		"create (1):",
		"  - " + commandPath("build"),
		"update (1):",
		"  - " + commandPath("explore"),
		"delete (1):",
		"  - AGENTS.md#core",
		"",
	}, "\n")
	if out != want {
		t.Fatalf("got verbose output %q, want %q", out, want)
	}
}

func TestCommittedLocalSmokeExampleRunsDeterministically(t *testing.T) {
	dir := copyExampleWorkspace(t, "local-smoke")

	readme, err := os.ReadFile(filepath.Join(dir, "README.md"))
	if err != nil {
		t.Fatalf("read local smoke README: %v", err)
	}
	if !strings.Contains(string(readme), "go run ../../cmd/agentspec plan --target opencode") {
		t.Fatalf("expected plan smoke command in README, got %q", string(readme))
	}
	if !strings.Contains(string(readme), "go run ../../cmd/agentspec apply --target opencode") {
		t.Fatalf("expected apply smoke command in README, got %q", string(readme))
	}

	out, err := runCommand(t, dir, []string{"agentspec", "plan", "--target", "opencode"})
	if err != nil {
		t.Fatalf("run plan: %v", err)
	}

	want := compactGroup(
		"create",
		commandPath("explore"),
		filepath.Join(".opencode", "skills", "local-audit", "SKILL.md"),
		"AGENTS.md#workspace-core",
	)
	if out != want {
		t.Fatalf("got plan output %q, want %q", out, want)
	}

	if _, err := runCommand(t, dir, []string{"agentspec", "apply", "--target", "opencode"}); err != nil {
		t.Fatalf("run apply: %v", err)
	}

	commandBody, err := os.ReadFile(filepath.Join(dir, commandPath("explore")))
	if err != nil {
		t.Fatalf("read command output: %v", err)
	}
	if string(commandBody) != "Explore the local smoke example.\n" {
		t.Fatalf("got command output %q", string(commandBody))
	}

	skillBody, err := os.ReadFile(filepath.Join(dir, ".opencode", "skills", "local-audit", "SKILL.md"))
	if err != nil {
		t.Fatalf("read skill output: %v", err)
	}
	if !strings.Contains(string(skillBody), "name: local-audit") {
		t.Fatalf("expected local skill frontmatter, got %q", string(skillBody))
	}

	agentsBody, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	if !strings.Contains(string(agentsBody), "Local smoke rules") {
		t.Fatalf("expected section content in AGENTS.md, got %q", string(agentsBody))
	}

	out, err = runCommand(t, dir, []string{"agentspec", "plan", "--target", "opencode"})
	if err != nil {
		t.Fatalf("run second plan: %v", err)
	}
	if out != "No managed changes.\n" {
		t.Fatalf("got second plan output %q", out)
	}
}

func managedWorkspace(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	writeWorkspaceFile(t, filepath.Join(dir, ".agentspec", "commands", "explore.md"), "Explore\n")
	writeWorkspaceConfig(t, dir, "sections:\n  core:\n    inline: |\n      Core rules\ncommands:\n  explore:\n    path: ./.agentspec/commands/explore.md\nagents: {}\nskills: {}\n")

	if _, err := runCommand(t, dir, []string{"agentspec", "apply", "--target", "opencode"}); err != nil {
		t.Fatalf("apply baseline workspace: %v", err)
	}

	return dir
}

func writeWorkspaceConfig(t *testing.T, dir, body string) {
	t.Helper()
	writeWorkspaceFile(t, filepath.Join(dir, "agentspec.yaml"), body)
}

func writeWorkspaceFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func copyExampleWorkspace(t *testing.T, name string) string {
	t.Helper()

	src := filepath.Join(repoRoot(t), "example", name)
	dst := t.TempDir()
	copyDir(t, src, dst)
	return dst
}

func repoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func copyDir(t *testing.T, src, dst string) {
	t.Helper()

	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}

		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode().Perm())
		}

		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, raw, info.Mode().Perm())
	})
	if err != nil {
		t.Fatalf("copy %q to %q: %v", src, dst, err)
	}
}

func runCommand(t *testing.T, dir string, args []string) (string, error) {
	t.Helper()
	runCommandMu.Lock()
	defer runCommandMu.Unlock()

	type readResult struct {
		raw []byte
		err error
	}

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

	readCh := make(chan readResult, 1)
	go func() {
		raw, err := io.ReadAll(r)
		readCh <- readResult{raw: raw, err: err}
	}()

	runErr := newCommand().Run(context.Background(), args)
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	result := <-readCh
	if result.err != nil {
		t.Fatal(result.err)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}

	return string(result.raw), runErr
}

func commandPath(name string) string {
	return filepath.Join(".opencode", "commands", name+".md")
}

func compactGroup(kind string, paths ...string) string {
	return fmt.Sprintf("%s (%d): %s\n", kind, len(paths), strings.Join(paths, ", "))
}
