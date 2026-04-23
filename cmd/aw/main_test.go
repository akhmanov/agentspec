package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCommandIncludesInitPlanAndApply(t *testing.T) {
	cmd := newCommand()

	if cmd.Name != "aw" {
		t.Fatalf("got name %q, want %q", cmd.Name, "aw")
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

	err = newCommand().Run(context.Background(), []string{"aw", "init"})
	if err != nil {
		t.Fatalf("run init: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(dir, "aw.yaml"))
	if err != nil {
		t.Fatalf("read aw.yaml: %v", err)
	}

	want := "sections: {}\ncommands: {}\nagents: {}\nskills: {}\n"
	if string(raw) != want {
		t.Fatalf("got config %q, want %q", string(raw), want)
	}
}

func TestPlanDoesNotWriteAndApplyMaterializesOpenCodeOutputs(t *testing.T) {
	dir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join(dir, ".aw", "commands"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".aw", "commands", "explore.md"), []byte("Explore\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	raw := []byte("sections:\n  core:\n    inline: |\n      Core rules\ncommands:\n  explore:\n    path: ./.aw/commands/explore.md\nagents: {}\nskills: {}\n")
	if err := os.WriteFile(filepath.Join(dir, "aw.yaml"), raw, 0o644); err != nil {
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

	out, err := runCommand(t, dir, []string{"aw", "plan", "--opencode"})
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

	if _, err := os.Stat(filepath.Join(dir, ".aw", "state", "opencode.json")); !os.IsNotExist(err) {
		t.Fatalf("expected no state after plan, got err %v", err)
	}

	_, err = runCommand(t, dir, []string{"aw", "apply", "--opencode"})
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
	if string(agents) != "<!-- aw:section:start core -->\nCore rules\n<!-- aw:section:end core -->\n" {
		t.Fatalf("got AGENTS.md %q", string(agents))
	}

	out, err = runCommand(t, dir, []string{"aw", "plan", "--opencode"})
	if err != nil {
		t.Fatalf("run second plan: %v", err)
	}
	if !strings.Contains(out, "No managed changes.") {
		t.Fatalf("got second plan output %q", out)
	}
}

func TestPlanReportsConflictsWithoutWriting(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".opencode", "commands"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".aw", "commands"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".opencode", "commands", "explore.md"), []byte("foreign\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".aw", "commands", "explore.md"), []byte("Explore\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	raw := []byte("sections: {}\ncommands:\n  explore:\n    path: ./.aw/commands/explore.md\nagents: {}\nskills: {}\n")
	if err := os.WriteFile(filepath.Join(dir, "aw.yaml"), raw, 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := runCommand(t, dir, []string{"aw", "plan", "--opencode"})
	if err != nil {
		t.Fatalf("run plan: %v", err)
	}
	if !strings.Contains(out, "conflict .opencode/commands/explore.md: refusing to overwrite foreign file") {
		t.Fatalf("got plan output %q", out)
	}

	if _, err := os.Stat(filepath.Join(dir, ".aw", "state", "opencode.json")); !os.IsNotExist(err) {
		t.Fatalf("expected no state after conflict plan, got err %v", err)
	}
}

func TestPlanAndApplyWorkWithGitHubRemoteFixtures(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".aw", "commands"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".aw", "commands", "local.md"), []byte("Local\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	archive := buildGitHubArchive(t, map[string]string{
		"repo-main/skills/debug/SKILL.md":       "---\nname: debug\ndescription: Debug skill\n---\n\n# Debug\n",
		"repo-main/skills/debug/notes/guide.md": "Guide\n",
	})
	server := newGitHubFixtureServer(map[string]string{
		"/raw/owner/repo/main/commands/remote.md": "Remote\n",
	}, map[string][]byte{
		"/archive/owner/repo/tar.gz/main": archive,
	})
	defer server.Close()

	t.Setenv("AW_GITHUB_RAW_BASE_URL", server.URL+"/raw")
	t.Setenv("AW_GITHUB_ARCHIVE_BASE_URL", server.URL+"/archive")

	raw := []byte("sections:\n  core:\n    inline: |\n      Core rules\ncommands:\n  local:\n    path: ./.aw/commands/local.md\n  remote:\n    github:\n      repo: owner/repo\n      ref: main\n      path: commands/remote.md\nagents: {}\nskills:\n  debug:\n    github:\n      repo: owner/repo\n      ref: main\n      path: skills/debug\n")
	if err := os.WriteFile(filepath.Join(dir, "aw.yaml"), raw, 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := runCommand(t, dir, []string{"aw", "plan", "--opencode"})
	if err != nil {
		t.Fatalf("run plan: %v", err)
	}
	if !strings.Contains(out, "create .opencode/commands/remote.md") {
		t.Fatalf("got plan output %q", out)
	}
	if !strings.Contains(out, "create .agents/skills/debug/SKILL.md") {
		t.Fatalf("got plan output %q", out)
	}

	_, err = runCommand(t, dir, []string{"aw", "apply", "--opencode"})
	if err != nil {
		t.Fatalf("run apply: %v", err)
	}

	cmd, err := os.ReadFile(filepath.Join(dir, ".opencode", "commands", "remote.md"))
	if err != nil {
		t.Fatalf("read remote command output: %v", err)
	}
	if string(cmd) != "Remote\n" {
		t.Fatalf("got remote command output %q", string(cmd))
	}

	skill, err := os.ReadFile(filepath.Join(dir, ".agents", "skills", "debug", "notes", "guide.md"))
	if err != nil {
		t.Fatalf("read remote skill bundle output: %v", err)
	}
	if string(skill) != "Guide\n" {
		t.Fatalf("got remote skill output %q", string(skill))
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

func newGitHubFixtureServer(text map[string]string, blobs map[string][]byte) *httptest.Server {
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

func buildGitHubArchive(t *testing.T, files map[string]string) []byte {
	t.Helper()

	var gz bytes.Buffer
	zw := gzip.NewWriter(&gz)
	tw := tar.NewWriter(zw)
	for name, body := range files {
		hdr := &tar.Header{Name: name, Mode: 0o644, Size: int64(len(body))}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatalf("write tar header: %v", err)
		}
		if _, err := tw.Write([]byte(body)); err != nil {
			t.Fatalf("write tar body: %v", err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("close tar: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close gzip: %v", err)
	}

	return gz.Bytes()
}
