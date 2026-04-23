package config

import (
	"fmt"
	"net/url"
	"os"
	pathpkg "path"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var validID = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
var validGitHubRepo = regexp.MustCompile(`^[A-Za-z0-9._-]+/[A-Za-z0-9._-]+$`)

type Source struct {
	Inline string     `yaml:"inline,omitempty"`
	Path   string     `yaml:"path,omitempty"`
	HTTP   string     `yaml:"http,omitempty"`
	GitHub *GitSource `yaml:"github,omitempty"`
	GitLab *GitSource `yaml:"gitlab,omitempty"`
}

type GitSource struct {
	Repo string `yaml:"repo,omitempty"`
	Ref  string `yaml:"ref,omitempty"`
	Path string `yaml:"path,omitempty"`
}

type Config struct {
	Sections     map[string]Source `yaml:"sections"`
	Commands     map[string]Source `yaml:"commands"`
	Agents       map[string]Source `yaml:"agents"`
	Skills       map[string]Source `yaml:"skills"`
	SectionOrder []string          `yaml:"-"`
	CommandOrder []string          `yaml:"-"`
	AgentOrder   []string          `yaml:"-"`
	SkillOrder   []string          `yaml:"-"`
}

func Load(path string) (*Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parse config %q: %w", path, err)
	}

	var node yaml.Node
	if err := yaml.Unmarshal(raw, &node); err != nil {
		return nil, fmt.Errorf("parse config order %q: %w", path, err)
	}

	if cfg.Sections == nil {
		cfg.Sections = map[string]Source{}
	}
	if cfg.Commands == nil {
		cfg.Commands = map[string]Source{}
	}
	if cfg.Agents == nil {
		cfg.Agents = map[string]Source{}
	}
	if cfg.Skills == nil {
		cfg.Skills = map[string]Source{}
	}

	cfg.SectionOrder = orderedKeys(&node, "sections")
	cfg.CommandOrder = orderedKeys(&node, "commands")
	cfg.AgentOrder = orderedKeys(&node, "agents")
	cfg.SkillOrder = orderedKeys(&node, "skills")

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if err := validateGroup("sections", c.Sections); err != nil {
		return err
	}
	if err := validateGroup("commands", c.Commands); err != nil {
		return err
	}
	if err := validateGroup("agents", c.Agents); err != nil {
		return err
	}
	return validateGroup("skills", c.Skills)
}

func validateGroup(kind string, items map[string]Source) error {
	for id, src := range items {
		if !validID.MatchString(id) {
			return fmt.Errorf("invalid resource id for %s.%s", kind, id)
		}
		if selectors(src) != 1 {
			return fmt.Errorf("invalid selectors for %s.%s: expected exactly one selector", kind, id)
		}
		if err := validateHTTP(kind, id, src.HTTP); err != nil {
			return err
		}
		if err := validateGitHub(kind, id, src.GitHub); err != nil {
			return err
		}
		if src.GitLab != nil {
			return fmt.Errorf("unsupported selector for %s.%s: gitlab", kind, id)
		}
	}

	return nil
}

func validateHTTP(kind, id, raw string) error {
	if raw == "" {
		return nil
	}

	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("invalid http selector for %s.%s: invalid url", kind, id)
	}
	if parsed.Scheme != "https" {
		return fmt.Errorf("invalid http selector for %s.%s: https required", kind, id)
	}

	return nil
}

func validateGitHub(kind, id string, src *GitSource) error {
	if src == nil {
		return nil
	}
	if src.Repo == "" {
		return fmt.Errorf("invalid github selector for %s.%s: missing repo", kind, id)
	}
	if !validGitHubRepo.MatchString(src.Repo) {
		return fmt.Errorf("invalid github selector for %s.%s: invalid repo", kind, id)
	}
	if src.Ref == "" {
		return fmt.Errorf("invalid github selector for %s.%s: missing ref", kind, id)
	}
	if src.Path == "" {
		return fmt.Errorf("invalid github selector for %s.%s: missing path", kind, id)
	}
	if !safeRemotePath(src.Path) {
		return fmt.Errorf("invalid github selector for %s.%s: unsafe path", kind, id)
	}

	return nil
}

func safeRemotePath(raw string) bool {
	if raw == "" || strings.HasPrefix(raw, "/") {
		return false
	}

	clean := pathpkg.Clean(raw)
	if clean != raw || clean == "." {
		return false
	}

	return clean != ".." && !strings.HasPrefix(clean, "../")
}

func selectors(src Source) int {
	n := 0
	if src.Inline != "" {
		n++
	}
	if src.Path != "" {
		n++
	}
	if src.HTTP != "" {
		n++
	}
	if src.GitHub != nil {
		n++
	}
	if src.GitLab != nil {
		n++
	}
	return n
}

func orderedKeys(doc *yaml.Node, key string) []string {
	if len(doc.Content) == 0 {
		return nil
	}

	root := doc.Content[0]
	for i := 0; i+1 < len(root.Content); i += 2 {
		if root.Content[i].Value != key {
			continue
		}

		group := root.Content[i+1]
		keys := make([]string, 0, len(group.Content)/2)
		for j := 0; j+1 < len(group.Content); j += 2 {
			keys = append(keys, group.Content[j].Value)
		}
		return keys
	}

	return nil
}
