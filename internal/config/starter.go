package config

import "os"

const starter = "# Managed sections inserted into AGENTS.md / CLAUDE.md\nsections: {}\n\n# Command documents materialized for the target\ncommands: {}\n\n# Agent documents materialized for the target\nagents: {}\n\n# Skill bundles materialized for the target\nskills: {}\n"

func WriteStarter(path string) error {
	return os.WriteFile(path, []byte(starter), 0o644)
}
