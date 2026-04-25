package claudecode

import (
	"path/filepath"

	"agentspec/internal/model"
)

func Build(res *model.Resolved) *model.Desired {
	des := &model.Desired{}

	for _, section := range res.Sections {
		des.Sections = append(des.Sections, model.Section{
			Path: "CLAUDE.md",
			ID:   section.ID,
			Body: section.Body,
		})
	}

	for _, cmd := range res.Commands {
		des.Files = append(des.Files, model.Output{
			Path: filepath.Join(".claude", "commands", cmd.ID+".md"),
			Body: cmd.Body,
		})
	}

	for _, agent := range res.Agents {
		des.Files = append(des.Files, model.Output{
			Path: filepath.Join(".claude", "agents", agent.ID+".md"),
			Body: agent.Body,
		})
	}

	for _, skill := range res.Skills {
		for _, file := range skill.Files {
			des.Files = append(des.Files, model.Output{
				Path: filepath.Join(".claude", "skills", skill.ID, file.Path),
				Body: file.Body,
			})
		}
	}

	return des
}
