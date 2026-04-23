package resolve

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	pathpkg "path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"aw/internal/config"
	"aw/internal/model"
	"gopkg.in/yaml.v3"
)

const (
	defaultHTTPTimeout = 10 * time.Second
	defaultMaxBytes    = 1 << 20
)

type httpDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type remoteLoader struct {
	client               httpDoer
	maxBytes             int64
	httpBaseURL          string
	githubRawBaseURL     string
	githubArchiveBaseURL string
}

type statusError struct {
	url    string
	code   int
	status string
}

type skillMeta struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

func newRemoteLoader(client httpDoer) *remoteLoader {
	if client == nil {
		client = &http.Client{Timeout: defaultHTTPTimeout}
	}

	return &remoteLoader{
		client:               client,
		maxBytes:             defaultMaxBytes,
		githubRawBaseURL:     envOr("AW_GITHUB_RAW_BASE_URL", "https://raw.githubusercontent.com"),
		githubArchiveBaseURL: envOr("AW_GITHUB_ARCHIVE_BASE_URL", "https://codeload.github.com"),
	}
}

func (e *statusError) Error() string {
	return fmt.Sprintf("fetch http %q: unexpected status %s", e.url, e.status)
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func (r *remoteLoader) fetchHTTP(addr string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, addr, nil)
	if err != nil {
		return nil, fmt.Errorf("build http request %q: %w", addr, err)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch http %q: %w", addr, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &statusError{url: addr, code: resp.StatusCode, status: resp.Status}
	}

	raw, err := io.ReadAll(io.LimitReader(resp.Body, r.maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read http %q: %w", addr, err)
	}
	if int64(len(raw)) > r.maxBytes {
		return nil, fmt.Errorf("read http %q: response too large", addr)
	}

	return raw, nil
}

func (r *remoteLoader) fetchGitHubFile(src *config.GitSource) ([]byte, error) {
	addr, err := r.githubFileURL(src)
	if err != nil {
		return nil, err
	}

	return r.fetchHTTP(addr)
}

func (r *remoteLoader) fetchGitHubArchive(src *config.GitSource) ([]byte, error) {
	addr, err := r.githubArchiveURL(src)
	if err != nil {
		return nil, err
	}

	return r.fetchHTTP(addr)
}

func (r *remoteLoader) githubFileURL(src *config.GitSource) (string, error) {
	base, err := url.Parse(r.githubRawBaseURL)
	if err != nil {
		return "", fmt.Errorf("parse github raw base url: %w", err)
	}
	base.Path = pathpkg.Join(base.Path, src.Repo, src.Ref, src.Path)
	return base.String(), nil
}

func (r *remoteLoader) githubArchiveURL(src *config.GitSource) (string, error) {
	base, err := url.Parse(r.githubArchiveBaseURL)
	if err != nil {
		return "", fmt.Errorf("parse github archive base url: %w", err)
	}
	base.Path = pathpkg.Join(base.Path, src.Repo, "tar.gz", src.Ref)
	return base.String(), nil
}

func Resolve(root string, cfg *config.Config) (*model.Resolved, error) {
	return resolveWithLoader(root, cfg, newRemoteLoader(nil))
}

func resolveWithLoader(root string, cfg *config.Config, loader *remoteLoader) (*model.Resolved, error) {
	res := &model.Resolved{}

	sections, err := resolveDocs(root, cfg.Sections, cfg.SectionOrder, loader)
	if err != nil {
		return nil, fmt.Errorf("resolve sections: %w", err)
	}
	res.Sections = sections

	commands, err := resolveDocs(root, cfg.Commands, cfg.CommandOrder, loader)
	if err != nil {
		return nil, fmt.Errorf("resolve commands: %w", err)
	}
	res.Commands = commands

	agents, err := resolveDocs(root, cfg.Agents, cfg.AgentOrder, loader)
	if err != nil {
		return nil, fmt.Errorf("resolve agents: %w", err)
	}
	res.Agents = agents

	skills, err := resolveSkills(root, cfg.Skills, cfg.SkillOrder, loader)
	if err != nil {
		return nil, fmt.Errorf("resolve skills: %w", err)
	}
	res.Skills = skills

	return res, nil
}

func resolveDocs(root string, items map[string]config.Source, order []string, loader *remoteLoader) ([]model.Document, error) {
	docs := make([]model.Document, 0, len(order))
	for _, id := range order {
		body, err := resolveBody(root, items[id], loader)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", id, err)
		}
		docs = append(docs, model.Document{ID: id, Body: body})
	}
	return docs, nil
}

func resolveSkills(root string, items map[string]config.Source, order []string, loader *remoteLoader) ([]model.Skill, error) {
	skills := make([]model.Skill, 0, len(order))
	for _, id := range order {
		src := items[id]
		if src.Inline != "" {
			files := []model.File{{
				Path: "SKILL.md",
				Body: src.Inline,
			}}
			if err := validateSkillFiles(files); err != nil {
				return nil, fmt.Errorf("%s: %w", id, err)
			}
			skills = append(skills, model.Skill{
				ID:    id,
				Files: files,
			})
			continue
		}

		files, err := resolveSkillFiles(root, src, loader)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", id, err)
		}
		if err := validateSkillFiles(files); err != nil {
			return nil, fmt.Errorf("%s: %w", id, err)
		}

		skills = append(skills, model.Skill{ID: id, Files: files})
	}

	return skills, nil
}

func resolveBody(root string, src config.Source, loader *remoteLoader) (string, error) {
	if src.Inline != "" {
		return src.Inline, nil
	}
	if src.HTTP != "" {
		raw, err := loader.fetchHTTP(src.HTTP)
		if err != nil {
			return "", err
		}
		return string(raw), nil
	}
	if src.GitHub != nil {
		raw, err := loader.fetchGitHubFile(src.GitHub)
		if err != nil {
			return "", err
		}
		return string(raw), nil
	}

	raw, err := os.ReadFile(resolvePath(root, src.Path))
	if err != nil {
		return "", fmt.Errorf("read path %q: %w", src.Path, err)
	}

	return string(raw), nil
}

func resolvePath(root, path string) string {
	if filepath.IsAbs(path) {
		return path
	}

	return filepath.Join(root, path)
}

func resolveSkillFiles(root string, src config.Source, loader *remoteLoader) ([]model.File, error) {
	if src.HTTP != "" {
		raw, err := loader.fetchHTTP(src.HTTP)
		if err != nil {
			return nil, err
		}
		return []model.File{{Path: "SKILL.md", Body: string(raw)}}, nil
	}
	if src.GitHub != nil {
		return resolveGitHubSkillFiles(src.GitHub, loader)
	}

	full := resolvePath(root, src.Path)
	info, err := os.Stat(full)
	if err != nil {
		return nil, fmt.Errorf("stat skill path %q: %w", src.Path, err)
	}

	if !info.IsDir() {
		raw, err := os.ReadFile(full)
		if err != nil {
			return nil, fmt.Errorf("read skill path %q: %w", src.Path, err)
		}
		return []model.File{{Path: "SKILL.md", Body: string(raw)}}, nil
	}

	files := []model.File{}
	paths := []string{}
	err = filepath.WalkDir(full, func(child string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(full, child)
		if err != nil {
			return err
		}
		paths = append(paths, rel)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk skill path %q: %w", src.Path, err)
	}

	sort.Strings(paths)
	hasRoot := false
	for _, rel := range paths {
		if rel == "SKILL.md" {
			hasRoot = true
		}

		raw, err := os.ReadFile(filepath.Join(full, rel))
		if err != nil {
			return nil, fmt.Errorf("read skill file %q: %w", rel, err)
		}
		files = append(files, model.File{Path: rel, Body: string(raw)})
	}

	if !hasRoot {
		return nil, fmt.Errorf("skill path %q: missing root SKILL.md", src.Path)
	}

	return files, nil
}

func resolveGitHubSkillFiles(src *config.GitSource, loader *remoteLoader) ([]model.File, error) {
	raw, err := loader.fetchGitHubFile(src)
	if err == nil {
		return []model.File{{Path: "SKILL.md", Body: string(raw)}}, nil
	}
	if !notFound(err) {
		return nil, err
	}

	archive, err := loader.fetchGitHubArchive(src)
	if err != nil {
		return nil, err
	}

	files, err := extractGitHubSkillBundle(archive, src.Path)
	if err != nil {
		return nil, fmt.Errorf("github skill path %q: %w", src.Path, err)
	}

	return files, nil
}

func extractGitHubSkillBundle(raw []byte, prefix string) ([]model.File, error) {
	zr, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("open archive: %w", err)
	}
	defer zr.Close()

	tr := tar.NewReader(zr)
	prefix = pathpkg.Clean(prefix)
	files := []model.File{}
	hasRoot := false
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read archive entry: %w", err)
		}
		if hdr.FileInfo().IsDir() {
			continue
		}

		rel, ok, err := bundlePath(hdr.Name, prefix)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}

		body, err := io.ReadAll(tr)
		if err != nil {
			return nil, fmt.Errorf("read archive body %q: %w", hdr.Name, err)
		}
		if rel == "SKILL.md" {
			hasRoot = true
		}
		files = append(files, model.File{Path: rel, Body: string(body)})
	}
	if !hasRoot {
		return nil, fmt.Errorf("missing root SKILL.md")
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})

	return files, nil
}

func bundlePath(name, prefix string) (string, bool, error) {
	parts := strings.SplitN(name, "/", 2)
	if len(parts) != 2 {
		return "", false, nil
	}
	rest := parts[1]
	if !safeBundlePath(rest) {
		return "", false, fmt.Errorf("unsafe bundle path %q", rest)
	}
	if rest == prefix {
		return "SKILL.md", true, nil
	}
	if !strings.HasPrefix(rest, prefix+"/") {
		return "", false, nil
	}

	rel := strings.TrimPrefix(rest, prefix+"/")
	if !safeBundlePath(rel) {
		return "", false, fmt.Errorf("unsafe bundle path %q", rel)
	}

	return filepath.FromSlash(rel), true, nil
}

func safeBundlePath(raw string) bool {
	if raw == "" || pathpkg.IsAbs(raw) {
		return false
	}
	clean := pathpkg.Clean(raw)
	if clean != raw || clean == "." {
		return false
	}
	return clean != ".." && !strings.HasPrefix(clean, "../")
}

func notFound(err error) bool {
	var status *statusError
	return errors.As(err, &status) && status.code == http.StatusNotFound
}

func validateSkillFiles(files []model.File) error {
	for _, file := range files {
		if file.Path != "SKILL.md" {
			continue
		}
		if validSkill(file.Body) {
			return nil
		}
		return fmt.Errorf("invalid SKILL.md: missing skill frontmatter")
	}

	return fmt.Errorf("invalid SKILL.md: missing root skill file")
}

func validSkill(body string) bool {
	body = strings.ReplaceAll(body, "\r\n", "\n")
	body = strings.ReplaceAll(body, "\r", "\n")

	parts := strings.SplitN(body, "\n---\n", 2)
	if len(parts) != 2 || !strings.HasPrefix(body, "---\n") {
		return false
	}

	meta := strings.TrimPrefix(parts[0], "---\n")
	var cfg skillMeta
	if err := yaml.Unmarshal([]byte(meta), &cfg); err != nil {
		return false
	}

	return cfg.Name != "" && cfg.Description != ""
}
