package sync

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agentspec/internal/model"
)

func TestPreviewReportsManagedCreatesWithoutWriting(t *testing.T) {
	dir := t.TempDir()
	des := &model.Desired{
		Files: []model.Output{{
			Path: filepath.Join(".opencode", "commands", "explore.md"),
			Body: "Explore\n",
		}},
		Sections: []model.Section{{
			Path: "AGENTS.md",
			ID:   "core",
			Body: "Core rules\n",
		}},
	}

	plan, err := Preview(dir, "opencode", des)
	if err != nil {
		t.Fatalf("preview desired state: %v", err)
	}

	if len(plan.Changes) != 2 {
		t.Fatalf("got %d changes, want %d", len(plan.Changes), 2)
	}

	if plan.Changes[0].Kind != Create || plan.Changes[0].Path != filepath.Join(".opencode", "commands", "explore.md") {
		t.Fatalf("got first change %#v", plan.Changes[0])
	}

	if plan.Changes[1].Kind != Create || plan.Changes[1].Path != "AGENTS.md#core" {
		t.Fatalf("got second change %#v", plan.Changes[1])
	}

	if len(plan.Conflicts) != 0 {
		t.Fatalf("got conflicts %#v, want none", plan.Conflicts)
	}

	if _, err := os.Stat(filepath.Join(dir, ".opencode", "commands", "explore.md")); !os.IsNotExist(err) {
		t.Fatalf("expected no managed file writes, got err %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "AGENTS.md")); !os.IsNotExist(err) {
		t.Fatalf("expected no section writes, got err %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".agentspec", "state", "opencode.json")); !os.IsNotExist(err) {
		t.Fatalf("expected no state writes, got err %v", err)
	}
}

func TestPreviewReportsOwnershipConflictWithoutError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".opencode", "commands")
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "explore.md"), []byte("foreign\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	des := &model.Desired{
		Files: []model.Output{{
			Path: filepath.Join(".opencode", "commands", "explore.md"),
			Body: "Explore\n",
		}},
	}

	plan, err := Preview(dir, "opencode", des)
	if err != nil {
		t.Fatalf("preview desired state: %v", err)
	}

	if len(plan.Conflicts) != 1 {
		t.Fatalf("got %d conflicts, want %d", len(plan.Conflicts), 1)
	}

	if plan.Conflicts[0].Path != filepath.Join(".opencode", "commands", "explore.md") {
		t.Fatalf("got conflict %#v", plan.Conflicts[0])
	}
}

func TestPreviewReportsConflictForForeignSectionMarkers(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("<!-- foreign:section:start core -->\nCore rules\n<!-- foreign:section:end core -->\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	des := &model.Desired{
		Sections: []model.Section{{
			Path: "AGENTS.md",
			ID:   "core",
			Body: "Core rules\n",
		}},
	}

	plan, err := Preview(dir, "opencode", des)
	if err != nil {
		t.Fatalf("preview desired state: %v", err)
	}
	if len(plan.Changes) != 0 {
		t.Fatalf("got changes %#v, want none", plan.Changes)
	}
	if len(plan.Conflicts) != 1 {
		t.Fatalf("got conflicts %#v", plan.Conflicts)
	}
	if plan.Conflicts[0].Path != "AGENTS.md#core" {
		t.Fatalf("got conflict %#v", plan.Conflicts[0])
	}
	if !strings.Contains(plan.Conflicts[0].Reason, "foreign section file") {
		t.Fatalf("got conflict %#v", plan.Conflicts[0])
	}
}

func TestApplyWritesManagedFilesSectionsAndState(t *testing.T) {
	dir := t.TempDir()
	des := &model.Desired{
		Files: []model.Output{{
			Path: filepath.Join(".opencode", "commands", "explore.md"),
			Body: "Explore\n",
		}},
		Sections: []model.Section{{
			Path: "AGENTS.md",
			ID:   "core",
			Body: "Core rules\n",
		}},
	}

	if err := Apply(dir, "opencode", des); err != nil {
		t.Fatalf("apply desired state: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(dir, ".opencode", "commands", "explore.md"))
	if err != nil {
		t.Fatalf("read managed file: %v", err)
	}
	if string(raw) != "Explore\n" {
		t.Fatalf("got file body %q, want %q", string(raw), "Explore\n")
	}

	agents, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	text := string(agents)
	if !strings.Contains(text, "<!-- agentspec:section:start core -->") {
		t.Fatalf("missing section start marker in %q", text)
	}
	if !strings.Contains(text, "Core rules\n") {
		t.Fatalf("missing section body in %q", text)
	}
	if !strings.Contains(text, "<!-- agentspec:section:end core -->") {
		t.Fatalf("missing section end marker in %q", text)
	}

	if _, err := os.Stat(filepath.Join(dir, ".agentspec", "state", "opencode.json")); err != nil {
		t.Fatalf("stat state file: %v", err)
	}
}

func TestApplyTracksOpenCodeSkillStateUnderTargetNativeRoot(t *testing.T) {
	dir := t.TempDir()
	des := &model.Desired{
		Files: []model.Output{{
			Path: filepath.Join(".opencode", "skills", "debug", "SKILL.md"),
			Body: "---\nname: debug\ndescription: Debug skill\n---\n\n# Debug\n",
		}},
	}

	if err := Apply(dir, "opencode", des); err != nil {
		t.Fatalf("apply desired state: %v", err)
	}

	plan, err := Preview(dir, "opencode", des)
	if err != nil {
		t.Fatalf("preview desired state: %v", err)
	}
	if len(plan.Changes) != 0 || len(plan.Conflicts) != 0 {
		t.Fatalf("got plan %#v, want no changes or conflicts", plan)
	}
}

func TestPreviewMigratesLegacyOpenCodeSkillStateToTargetNativeRoot(t *testing.T) {
	dir := t.TempDir()
	legacyPath := filepath.Join(".agents", "skills", "debug", "SKILL.md")
	body := "---\nname: debug\ndescription: Debug skill\n---\n\n# Debug\n"

	if err := os.MkdirAll(filepath.Join(dir, ".agents", "skills", "debug"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".agentspec", "state"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, legacyPath), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	stateBody := `{"owner":"agentspec","version":3,"files":[{"path":"` + legacyPath + `","hash":"` + hashText(body) + `"}],"sections":[]}`
	if err := os.WriteFile(filepath.Join(dir, ".agentspec", "state", "opencode.json"), []byte(stateBody), 0o644); err != nil {
		t.Fatal(err)
	}

	des := &model.Desired{
		Files: []model.Output{{
			Path: filepath.Join(".opencode", "skills", "debug", "SKILL.md"),
			Body: body,
		}},
	}

	plan, err := Preview(dir, "opencode", des)
	if err != nil {
		t.Fatalf("preview desired state: %v", err)
	}
	if len(plan.Changes) != 2 {
		t.Fatalf("got %d changes, want %d", len(plan.Changes), 2)
	}
	if plan.Changes[0].Kind != Create || plan.Changes[0].Path != filepath.Join(".opencode", "skills", "debug", "SKILL.md") {
		t.Fatalf("got first change %#v", plan.Changes[0])
	}
	if plan.Changes[1].Kind != Delete || plan.Changes[1].Path != legacyPath {
		t.Fatalf("got second change %#v", plan.Changes[1])
	}
}

func TestApplyRejectsForeignFileCollision(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".opencode", "commands")
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(path, "explore.md")
	if err := os.WriteFile(file, []byte("foreign\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	des := &model.Desired{
		Files: []model.Output{{
			Path: filepath.Join(".opencode", "commands", "explore.md"),
			Body: "Explore\n",
		}},
	}

	err := Apply(dir, "opencode", des)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestApplyRejectsForeignSectionMarkersWithoutMutation(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("<!-- foreign:section:start core -->\nCore rules\n<!-- foreign:section:end core -->\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	des := &model.Desired{
		Sections: []model.Section{{
			Path: "AGENTS.md",
			ID:   "core",
			Body: "Core rules\n",
		}},
	}

	err := Apply(dir, "opencode", des)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "foreign section file") {
		t.Fatalf("got error %q, want foreign section conflict", err)
	}

	raw, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read preserved AGENTS.md: %v", err)
	}
	text := string(raw)
	if !strings.Contains(text, "<!-- foreign:section:start core -->") {
		t.Fatalf("foreign marker unexpectedly changed in %q", text)
	}
	if strings.Contains(text, "<!-- agentspec:section:start core -->") {
		t.Fatalf("unexpected agentspec marker in %q", text)
	}
}

func TestApplyRemovesOwnedOrphansAndPreservesForeignContent(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("Intro\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	first := &model.Desired{
		Files: []model.Output{{
			Path: filepath.Join(".opencode", "commands", "explore.md"),
			Body: "Explore\n",
		}},
		Sections: []model.Section{{
			Path: "AGENTS.md",
			ID:   "core",
			Body: "Core rules\n",
		}},
	}

	if err := Apply(dir, "opencode", first); err != nil {
		t.Fatalf("apply initial desired state: %v", err)
	}

	if err := Apply(dir, "opencode", &model.Desired{}); err != nil {
		t.Fatalf("apply empty desired state: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".opencode", "commands", "explore.md")); !os.IsNotExist(err) {
		t.Fatalf("expected managed file removed, got err %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	if string(raw) != "Intro\n" {
		t.Fatalf("got AGENTS.md %q, want %q", string(raw), "Intro\n")
	}
}

func TestApplyRejectsForeignStateFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".agentspec", "state")
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "opencode.json"), []byte(`{"owner":"other","version":3,"files":[],"sections":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	err := Apply(dir, "opencode", &model.Desired{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "foreign state") {
		t.Fatalf("got error %q, want foreign state", err)
	}
}

func TestApplyDoesNotMutateWhenPreflightFails(t *testing.T) {
	dir := t.TempDir()
	first := &model.Desired{
		Files: []model.Output{{
			Path: filepath.Join(".opencode", "commands", "explore.md"),
			Body: "Explore\n",
		}},
		Sections: []model.Section{{
			Path: "AGENTS.md",
			ID:   "core",
			Body: "Core rules\n",
		}},
	}

	if err := Apply(dir, "opencode", first); err != nil {
		t.Fatalf("apply initial desired state: %v", err)
	}

	path := filepath.Join(dir, ".opencode", "commands")
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "review.md"), []byte("foreign\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	next := &model.Desired{
		Files: []model.Output{{
			Path: filepath.Join(".opencode", "commands", "review.md"),
			Body: "Review\n",
		}},
	}

	err := Apply(dir, "opencode", next)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	raw, err := os.ReadFile(filepath.Join(dir, ".opencode", "commands", "explore.md"))
	if err != nil {
		t.Fatalf("read preserved managed file: %v", err)
	}
	if string(raw) != "Explore\n" {
		t.Fatalf("got preserved file %q, want %q", string(raw), "Explore\n")
	}

	agents, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read preserved AGENTS.md: %v", err)
	}
	text := string(agents)
	if !strings.Contains(text, "<!-- agentspec:section:start core -->") {
		t.Fatalf("lost preserved section in %q", text)
	}
}

func TestApplyRejectsTamperedStateFilePath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".agentspec", "state")
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "opencode.json"), []byte(`{"owner":"agentspec","version":3,"files":[{"path":"../../outside","hash":"`+strings.Repeat("a", 64)+`"}],"sections":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	err := Apply(dir, "opencode", &model.Desired{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid state file path") {
		t.Fatalf("got error %q, want invalid state file path", err)
	}
}

func TestApplyRejectsTamperedStateSectionPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".agentspec", "state")
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "opencode.json"), []byte(`{"owner":"agentspec","version":3,"files":[],"sections":[{"path":"CLAUDE.md","id":"core"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	err := Apply(dir, "opencode", &model.Desired{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid state section path") {
		t.Fatalf("got error %q, want invalid state section path", err)
	}
}

func TestApplyRejectsTamperedManagedRootClaimOnPrune(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".opencode", "commands"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".agentspec", "state"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".opencode", "commands", "foreign.md"), []byte("foreign\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".agentspec", "state", "opencode.json"), []byte(`{"owner":"agentspec","version":3,"files":[{"path":".opencode/commands/foreign.md","hash":"`+strings.Repeat("a", 64)+`"}],"sections":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	err := Apply(dir, "opencode", &model.Desired{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "mismatched ownership fingerprint") {
		t.Fatalf("got error %q, want mismatched ownership fingerprint", err)
	}

	raw, err := os.ReadFile(filepath.Join(dir, ".opencode", "commands", "foreign.md"))
	if err != nil {
		t.Fatalf("read preserved foreign file: %v", err)
	}
	if string(raw) != "foreign\n" {
		t.Fatalf("got preserved file %q, want %q", string(raw), "foreign\n")
	}
}

func TestApplyRejectsTamperedManagedRootClaimOnOverwrite(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".opencode", "commands"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".agentspec", "state"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".opencode", "commands", "foreign.md"), []byte("foreign\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".agentspec", "state", "opencode.json"), []byte(`{"owner":"agentspec","version":3,"files":[{"path":".opencode/commands/foreign.md","hash":"`+strings.Repeat("a", 64)+`"}],"sections":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	des := &model.Desired{
		Files: []model.Output{{
			Path: filepath.Join(".opencode", "commands", "foreign.md"),
			Body: "owned\n",
		}},
	}

	err := Apply(dir, "opencode", des)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "mismatched ownership fingerprint") {
		t.Fatalf("got error %q, want mismatched ownership fingerprint", err)
	}

	raw, err := os.ReadFile(filepath.Join(dir, ".opencode", "commands", "foreign.md"))
	if err != nil {
		t.Fatalf("read preserved foreign file: %v", err)
	}
	if string(raw) != "foreign\n" {
		t.Fatalf("got preserved file %q, want %q", string(raw), "foreign\n")
	}
}
