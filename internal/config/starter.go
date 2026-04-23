package config

import "os"

const starter = "sections: {}\ncommands: {}\nagents: {}\nskills: {}\n"

func WriteStarter(path string) error {
	return os.WriteFile(path, []byte(starter), 0o644)
}
