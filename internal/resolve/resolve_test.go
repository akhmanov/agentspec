package resolve

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"aw/internal/config"
)

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
	archive := buildGitHubArchive(t, map[string]string{
		"repo-main/skills/debug/SKILL.md":       "---\nname: debug\ndescription: Debug skill\n---\n\n# Debug\n",
		"repo-main/skills/debug/notes/guide.md": "Guide\n",
	})
	server := newRemoteTestServer(t, nil, map[string][]byte{
		"/archive/owner/repo/tar.gz/main": archive,
	})
	defer server.Close()

	loader := newRemoteLoader(server.Client())
	loader.githubRawBaseURL = server.URL + "/raw"
	loader.githubArchiveBaseURL = server.URL + "/archive"

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
	if got := res.Skills[0].Files[1].Path; got != filepath.Join("notes", "guide.md") {
		t.Fatalf("got bundle path %q, want %q", got, filepath.Join("notes", "guide.md"))
	}
}

func TestResolveRejectsGitHubBundleWithoutRootSkillFile(t *testing.T) {
	archive := buildGitHubArchive(t, map[string]string{
		"repo-main/skills/debug/notes/guide.md": "Guide\n",
	})
	server := newRemoteTestServer(t, nil, map[string][]byte{
		"/archive/owner/repo/tar.gz/main": archive,
	})
	defer server.Close()

	loader := newRemoteLoader(server.Client())
	loader.githubRawBaseURL = server.URL + "/raw"
	loader.githubArchiveBaseURL = server.URL + "/archive"

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

func TestResolveRejectsGitHubBundleTraversal(t *testing.T) {
	archive := buildGitHubArchive(t, map[string]string{
		"repo-main/skills/debug/SKILL.md":      "---\nname: debug\ndescription: Debug skill\n---\n\n# Debug\n",
		"repo-main/skills/debug/../escape.txt": "nope\n",
	})
	server := newRemoteTestServer(t, nil, map[string][]byte{
		"/archive/owner/repo/tar.gz/main": archive,
	})
	defer server.Close()

	loader := newRemoteLoader(server.Client())
	loader.githubRawBaseURL = server.URL + "/raw"
	loader.githubArchiveBaseURL = server.URL + "/archive"

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

func TestResolveLoadsInlinePathAndSingleFileSkill(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".aw", "commands"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".aw", "skills"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, ".aw", "commands", "explore.md"), []byte("Explore\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".aw", "skills", "debug.md"), []byte("---\nname: debug\ndescription: Debug skill\n---\n\n# Debug\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	raw := []byte("sections:\n  core:\n    inline: |\n      Core rules\ncommands:\n  explore:\n    path: ./.aw/commands/explore.md\nagents: {}\nskills:\n  debug:\n    path: ./.aw/skills/debug.md\n")
	path := filepath.Join(dir, "aw.yaml")
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
	skill := filepath.Join(dir, ".aw", "skills", "debug")
	if err := os.MkdirAll(filepath.Join(skill, "notes"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(skill, "SKILL.md"), []byte("---\nname: debug\ndescription: Debug skill\n---\n\n# Debug\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skill, "notes", "guide.md"), []byte("Guide\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	raw := []byte("sections: {}\ncommands: {}\nagents: {}\nskills:\n  debug:\n    path: ./.aw/skills/debug\n")
	path := filepath.Join(dir, "aw.yaml")
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

func TestResolveRejectsDirectorySkillWithoutRootSkillFile(t *testing.T) {
	dir := t.TempDir()
	skill := filepath.Join(dir, ".aw", "skills", "debug")
	if err := os.MkdirAll(skill, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skill, "guide.md"), []byte("Guide\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	raw := []byte("sections: {}\ncommands: {}\nagents: {}\nskills:\n  debug:\n    path: ./.aw/skills/debug\n")
	path := filepath.Join(dir, "aw.yaml")
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
	if err := os.MkdirAll(filepath.Join(dir, ".aw", "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".aw", "skills", "debug.md"), []byte("not a skill\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	raw := []byte("sections: {}\ncommands: {}\nagents: {}\nskills:\n  debug:\n    path: ./.aw/skills/debug.md\n")
	path := filepath.Join(dir, "aw.yaml")
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
	skill := filepath.Join(dir, ".aw", "skills", "debug")
	if err := os.MkdirAll(skill, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skill, "SKILL.md"), []byte("---\nname: debug\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	raw := []byte("sections: {}\ncommands: {}\nagents: {}\nskills:\n  debug:\n    path: ./.aw/skills/debug\n")
	path := filepath.Join(dir, "aw.yaml")
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
	if err := os.MkdirAll(filepath.Join(dir, ".aw", "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	body := "---\r\nname: debug\r\ndescription: Debug skill\r\n---\r\n\r\n# Debug\r\n"
	if err := os.WriteFile(filepath.Join(dir, ".aw", "skills", "debug.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	raw := []byte("sections: {}\ncommands: {}\nagents: {}\nskills:\n  debug:\n    path: ./.aw/skills/debug.md\n")
	path := filepath.Join(dir, "aw.yaml")
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
	skill := filepath.Join(dir, ".aw", "skills", "debug")
	if err := os.MkdirAll(skill, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "---\r\nname: debug\r\ndescription: Debug skill\r\n---\r\n\r\n# Debug\r\n"
	if err := os.WriteFile(filepath.Join(skill, "SKILL.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	raw := []byte("sections: {}\ncommands: {}\nagents: {}\nskills:\n  debug:\n    path: ./.aw/skills/debug\n")
	path := filepath.Join(dir, "aw.yaml")
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
