package claudecode

import (
	"path/filepath"
	"testing"

	"agentspec/internal/model"
)

func TestBuildRendersClaudeCodePaths(t *testing.T) {
	res := &model.Resolved{
		Sections: []model.Document{{ID: "core", Body: "Core rules\n"}},
		Commands: []model.Document{{ID: "explore", Body: "Explore\n"}},
		Agents:   []model.Document{{ID: "reviewer", Body: "Review\n"}},
		Skills: []model.Skill{{
			ID: "debug",
			Files: []model.File{
				{Path: "SKILL.md", Body: "# Debug\n"},
				{Path: filepath.Join("notes", "guide.md"), Body: "Guide\n"},
			},
		}},
	}

	des := Build(res)

	if len(des.Sections) != 1 {
		t.Fatalf("got %d sections, want %d", len(des.Sections), 1)
	}
	if des.Sections[0].Path != "CLAUDE.md" {
		t.Fatalf("got section path %q, want %q", des.Sections[0].Path, "CLAUDE.md")
	}
	if des.Sections[0].ID != "core" {
		t.Fatalf("got section id %q, want %q", des.Sections[0].ID, "core")
	}

	if len(des.Files) != 4 {
		t.Fatalf("got %d files, want %d", len(des.Files), 4)
	}

	want := []string{
		filepath.Join(".claude", "commands", "explore.md"),
		filepath.Join(".claude", "agents", "reviewer.md"),
		filepath.Join(".claude", "skills", "debug", "SKILL.md"),
		filepath.Join(".claude", "skills", "debug", "notes", "guide.md"),
	}

	for i, path := range want {
		if des.Files[i].Path != path {
			t.Fatalf("got file %d path %q, want %q", i, des.Files[i].Path, path)
		}
	}
}
