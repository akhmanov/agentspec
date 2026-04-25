package resolve

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	pathpkg "path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"agentspec/internal/config"
	"agentspec/internal/model"
	"gopkg.in/yaml.v3"
)

const (
	defaultHTTPTimeout = 10 * time.Second
	defaultGitTimeout  = 30 * time.Second
	defaultMaxBytes    = 1 << 20
)

type httpDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type remoteLoader struct {
	client             httpDoer
	maxBytes           int64
	httpBaseURL        string
	githubRawBaseURL   string
	githubCloneBaseURL string
	runGit             func(dir string, env []string, args ...string) error
	runGitOutput       func(dir string, env []string, args ...string) ([]byte, error)
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

type gitTreeEntry struct {
	Mode   string
	Type   string
	Object string
	Path   string
}

type notDirectoryError struct {
	path string
}

func newRemoteLoader(client httpDoer) *remoteLoader {
	if client == nil {
		client = &http.Client{Timeout: defaultHTTPTimeout}
	}

	return &remoteLoader{
		client:             client,
		maxBytes:           defaultMaxBytes,
		githubRawBaseURL:   "https://raw.githubusercontent.com",
		githubCloneBaseURL: "https://github.com",
		runGit:             runGitCommand,
		runGitOutput:       runGitCommandOutput,
	}
}

func (e *statusError) Error() string {
	return fmt.Sprintf("fetch http %q: unexpected status %s", e.url, e.status)
}

func (e *notDirectoryError) Error() string {
	return fmt.Sprintf("skill path %q: not a directory", e.path)
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

func (r *remoteLoader) githubFileURL(src *config.GitSource) (string, error) {
	base, err := url.Parse(r.githubRawBaseURL)
	if err != nil {
		return "", fmt.Errorf("parse github raw base url: %w", err)
	}
	base.Path = pathpkg.Join(base.Path, src.Repo, src.Ref, src.Path)
	return base.String(), nil
}

func (r *remoteLoader) githubCloneURL(src *config.GitSource) (string, error) {
	base, err := url.Parse(r.githubCloneBaseURL)
	if err != nil {
		return "", fmt.Errorf("parse github clone base url: %w", err)
	}
	base.Path = pathpkg.Join(base.Path, src.Repo+".git")
	return base.String(), nil
}

func Resolve(root string, cfg *config.Config) (*model.Resolved, error) {
	return resolveWithLoader(root, cfg, newRemoteLoader(nil))
}

func resolveWithLoader(root string, cfg *config.Config, loader *remoteLoader) (*model.Resolved, error) {
	res := &model.Resolved{}
	configDir := localBaseDir(root, cfg)

	sections, err := resolveDocs(configDir, cfg.Sections, cfg.SectionOrder, loader)
	if err != nil {
		return nil, fmt.Errorf("resolve sections: %w", err)
	}
	res.Sections = sections

	commands, err := resolveDocs(configDir, cfg.Commands, cfg.CommandOrder, loader)
	if err != nil {
		return nil, fmt.Errorf("resolve commands: %w", err)
	}
	res.Commands = commands

	agents, err := resolveDocs(configDir, cfg.Agents, cfg.AgentOrder, loader)
	if err != nil {
		return nil, fmt.Errorf("resolve agents: %w", err)
	}
	res.Agents = agents

	skills, err := resolveSkills(configDir, cfg.Skills, cfg.SkillOrder, loader)
	if err != nil {
		return nil, fmt.Errorf("resolve skills: %w", err)
	}
	res.Skills = skills

	return res, nil
}

func localBaseDir(root string, cfg *config.Config) string {
	if cfg != nil && cfg.BaseDir != "" {
		return cfg.BaseDir
	}

	return root
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

	return readSkillFiles(resolvePath(root, src.Path), src.Path)
}

func resolveGitHubSkillFiles(src *config.GitSource, loader *remoteLoader) ([]model.File, error) {
	raw, err := loader.fetchGitHubFile(src)
	if err == nil {
		return []model.File{{Path: "SKILL.md", Body: string(raw)}}, nil
	}
	if !notFound(err) {
		return nil, err
	}
	rawErr := err

	files, err := loader.fetchGitHubBundle(src)
	if err != nil {
		var notDir *notDirectoryError
		if errors.As(err, &notDir) {
			return nil, rawErr
		}
		return nil, fmt.Errorf("github skill path %q: %w", src.Path, err)
	}

	return files, nil
}

func (r *remoteLoader) fetchGitHubBundle(src *config.GitSource) ([]model.File, error) {
	repoDir, cleanup, err := r.fetchGitHubRepo(src)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	entry, err := r.gitTreeEntry(repoDir, src.Path)
	if err != nil {
		return nil, err
	}
	if entry == nil || entry.Mode != "040000" || entry.Type != "tree" {
		return nil, &notDirectoryError{path: src.Path}
	}

	return r.readGitBundleFiles(repoDir, src.Path)
}

func (r *remoteLoader) fetchGitHubRepo(src *config.GitSource) (string, func(), error) {
	repoURL, err := r.githubCloneURL(src)
	if err != nil {
		return "", nil, err
	}

	tmp, err := os.MkdirTemp("", "agentspec-github-skill-*")
	if err != nil {
		return "", nil, fmt.Errorf("create temp repo dir: %w", err)
	}
	cleanup := func() {
		_ = os.RemoveAll(tmp)
	}

	repoDir := filepath.Join(tmp, "repo")
	env := []string{"GIT_LFS_SKIP_SMUDGE=1"}
	if err := r.runGit("", env, "init", repoDir); err != nil {
		cleanup()
		return "", nil, err
	}
	if err := r.runGit(repoDir, env, "remote", "add", "origin", repoURL); err != nil {
		cleanup()
		return "", nil, err
	}
	if err := r.runGit(repoDir, env, "fetch", "--depth", "1", "--filter=blob:none", "origin", src.Ref); err != nil {
		cleanup()
		return "", nil, err
	}

	return repoDir, cleanup, nil
}

func (r *remoteLoader) gitTreeEntry(repoDir, remotePath string) (*gitTreeEntry, error) {
	entries, err := r.gitTreeEntries(repoDir, remotePath, false)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, nil
	}
	return &entries[0], nil
}

func (r *remoteLoader) gitTreeEntries(repoDir, remotePath string, recursive bool) ([]gitTreeEntry, error) {
	args := []string{"ls-tree", "-z"}
	if recursive {
		args = append(args, "-r", "--full-tree")
	}
	args = append(args, "FETCH_HEAD", "--", literalGitPathspec(remotePath))

	raw, err := r.runGitOutput(repoDir, []string{"GIT_LFS_SKIP_SMUDGE=1"}, args...)
	if err != nil {
		return nil, err
	}

	return parseGitTreeEntries(raw)
}

func (r *remoteLoader) readGitBundleFiles(repoDir, prefix string) ([]model.File, error) {
	entries, err := r.gitTreeEntries(repoDir, prefix, true)
	if err != nil {
		return nil, err
	}

	files := []model.File{}
	hasRoot := false
	for _, entry := range entries {
		rel, err := bundleRelativePath(entry.Path, prefix)
		if err != nil {
			return nil, err
		}
		if entry.Mode == "120000" {
			return nil, fmt.Errorf("skill path %q: symlink bundle path %q", prefix, rel)
		}
		if entry.Type != "blob" || (entry.Mode != "100644" && entry.Mode != "100755") {
			return nil, fmt.Errorf("skill path %q: non-regular bundle path %q", prefix, rel)
		}

		raw, err := r.runGitOutput(repoDir, []string{"GIT_LFS_SKIP_SMUDGE=1"}, "cat-file", "blob", entry.Object)
		if err != nil {
			return nil, fmt.Errorf("read skill file %q: %w", rel, err)
		}
		if rel == "SKILL.md" {
			hasRoot = true
		}
		files = append(files, model.File{Path: filepath.FromSlash(rel), Body: string(raw)})
	}

	if !hasRoot {
		return nil, fmt.Errorf("skill path %q: missing root SKILL.md", prefix)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})

	return files, nil
}

func readSkillFiles(full, label string) ([]model.File, error) {
	info, err := os.Stat(full)
	if err != nil {
		return nil, fmt.Errorf("stat skill path %q: %w", label, err)
	}

	if !info.IsDir() {
		raw, err := os.ReadFile(full)
		if err != nil {
			return nil, fmt.Errorf("read skill path %q: %w", label, err)
		}
		return []model.File{{Path: "SKILL.md", Body: string(raw)}}, nil
	}

	return readSkillBundleDir(full, label)
}

func readSkillBundleDir(full, label string) ([]model.File, error) {
	files := []model.File{}
	paths := []string{}
	err := filepath.WalkDir(full, func(child string, d fs.DirEntry, err error) error {
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
		return nil, fmt.Errorf("walk skill path %q: %w", label, err)
	}

	sort.Strings(paths)
	hasRoot := false
	for _, rel := range paths {
		if !safeBundlePath(filepath.ToSlash(rel)) {
			return nil, fmt.Errorf("skill path %q: unsafe bundle path %q", label, rel)
		}
		info, err := os.Lstat(filepath.Join(full, rel))
		if err != nil {
			return nil, fmt.Errorf("stat skill file %q: %w", rel, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("skill path %q: symlink bundle path %q", label, rel)
		}
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("skill path %q: non-regular bundle path %q", label, rel)
		}
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
		return nil, fmt.Errorf("skill path %q: missing root SKILL.md", label)
	}

	return files, nil
}

func bundleRelativePath(name, prefix string) (string, error) {
	if !strings.HasPrefix(name, prefix+"/") {
		return "", fmt.Errorf("skill path %q: unexpected bundle path %q", prefix, name)
	}

	rel := strings.TrimPrefix(name, prefix+"/")
	if !safeBundlePath(rel) {
		return "", fmt.Errorf("skill path %q: unsafe bundle path %q", prefix, rel)
	}

	return rel, nil
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
		if !safeBundlePath(filepath.ToSlash(file.Path)) {
			return fmt.Errorf("unsafe bundle path %q", file.Path)
		}
	}

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

func runGitCommand(dir string, env []string, args ...string) error {
	_, err := runGitCommandOutput(dir, env, args...)
	return err
}

func runGitCommandOutput(dir string, env []string, args ...string) ([]byte, error) {
	timeout := defaultGitTimeout
	if raw := os.Getenv("AGENTSPEC_GIT_TIMEOUT"); raw != "" {
		if parsed, err := time.ParseDuration(raw); err == nil && parsed > 0 {
			timeout = parsed
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)

	output, err := cmd.CombinedOutput()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return nil, fmt.Errorf("run git %q: timed out after %s", strings.Join(args, " "), timeout)
	}
	if err == nil {
		return output, nil
	}

	msg := strings.TrimSpace(string(output))
	if msg == "" {
		return nil, fmt.Errorf("run git %q: %w", strings.Join(args, " "), err)
	}
	return nil, fmt.Errorf("run git %q: %w: %s", strings.Join(args, " "), err, msg)
}

func literalGitPathspec(path string) string {
	return ":(literal)" + path
}

func parseGitTreeEntries(raw []byte) ([]gitTreeEntry, error) {
	parts := bytes.Split(raw, []byte{0})
	entries := make([]gitTreeEntry, 0, len(parts))
	for _, part := range parts {
		if len(part) == 0 {
			continue
		}

		meta, path, ok := bytes.Cut(part, []byte{'\t'})
		if !ok {
			return nil, fmt.Errorf("parse git tree entry %q: missing path separator", string(part))
		}
		fields := strings.Fields(string(meta))
		if len(fields) != 3 {
			return nil, fmt.Errorf("parse git tree entry %q: unexpected metadata", string(part))
		}

		entries = append(entries, gitTreeEntry{
			Mode:   fields[0],
			Type:   fields[1],
			Object: fields[2],
			Path:   string(path),
		})
	}

	return entries, nil
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
