package sync

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"agentspec/internal/model"
)

var validID = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
var validHash = regexp.MustCompile(`^[a-f0-9]{64}$`)

const (
	currentOwner         = "agentspec"
	currentStateVersion  = 3
	currentStateDirName  = ".agentspec"
	currentSectionPrefix = "agentspec"
)

type state struct {
	Owner    string         `json:"owner"`
	Version  int            `json:"version"`
	Files    []fileState    `json:"files"`
	Sections []sectionState `json:"sections"`
}

type fileState struct {
	Path string `json:"path"`
	Hash string `json:"hash"`
}

type sectionState struct {
	Path string `json:"path"`
	ID   string `json:"id"`
}

type ChangeKind string

const (
	Create ChangeKind = "create"
	Update ChangeKind = "update"
	Delete ChangeKind = "delete"
)

type Change struct {
	Kind ChangeKind
	Path string
}

type Conflict struct {
	Path   string
	Reason string
}

type Plan struct {
	Changes   []Change
	Conflicts []Conflict
}

func Apply(root, target string, des *model.Desired) error {
	plan, prev, err := inspect(root, target, des)
	if err != nil {
		return err
	}
	if len(plan.Conflicts) != 0 {
		item := plan.Conflicts[0]
		return fmt.Errorf("%s for %q", item.Reason, item.Path)
	}

	for _, file := range des.Files {
		path := filepath.Join(root, file.Path)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return fmt.Errorf("mkdir file dir %q: %w", file.Path, err)
		}
		if err := os.WriteFile(path, []byte(file.Body), 0o644); err != nil {
			return fmt.Errorf("write file %q: %w", file.Path, err)
		}
	}

	for _, section := range des.Sections {
		path := filepath.Join(root, section.Path)
		raw, err := os.ReadFile(path)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("read section file %q: %w", section.Path, err)
		}

		body := upsertSection(string(raw), section)
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			return fmt.Errorf("write section file %q: %w", section.Path, err)
		}
	}

	if err := pruneSections(root, prev.Sections, des.Sections); err != nil {
		return err
	}
	if err := pruneFiles(root, prev.Files, des.Files); err != nil {
		return err
	}

	st := state{Owner: currentOwner, Version: currentStateVersion}
	for _, file := range des.Files {
		st.Files = append(st.Files, fileState{Path: file.Path, Hash: hashText(file.Body)})
	}
	for _, section := range des.Sections {
		st.Sections = append(st.Sections, sectionState{Path: section.Path, ID: section.ID})
	}

	if err := saveState(root, target, st); err != nil {
		return err
	}

	return nil
}

func Preview(root, target string, des *model.Desired) (*Plan, error) {
	plan, _, err := inspect(root, target, des)
	if err != nil {
		return nil, err
	}
	return plan, nil
}

func inspect(root, target string, des *model.Desired) (*Plan, state, error) {
	prev, err := loadState(root, target)
	if err != nil {
		return nil, state{}, err
	}

	plan := &Plan{}
	inspectFiles(root, prev, des, plan)
	if err := inspectSections(root, prev, des, plan); err != nil {
		return nil, state{}, err
	}

	return plan, prev, nil
}

func inspectFiles(root string, prev state, des *model.Desired, plan *Plan) {
	next := map[string]string{}
	for _, file := range des.Files {
		next[file.Path] = file.Body
		path := filepath.Join(root, file.Path)
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			plan.Changes = append(plan.Changes, Change{Kind: Create, Path: file.Path})
			continue
		}
		if err != nil {
			plan.Conflicts = append(plan.Conflicts, Conflict{Path: file.Path, Reason: fmt.Sprintf("stat failed: %v", err)})
			continue
		}
		if info.IsDir() {
			plan.Conflicts = append(plan.Conflicts, Conflict{Path: file.Path, Reason: "refusing to overwrite foreign file"})
			continue
		}

		item, ok := owned(prev.Files, file.Path)
		if !ok {
			plan.Conflicts = append(plan.Conflicts, Conflict{Path: file.Path, Reason: "refusing to overwrite foreign file"})
			continue
		}

		hash, err := hashFile(path)
		if err != nil {
			plan.Conflicts = append(plan.Conflicts, Conflict{Path: file.Path, Reason: fmt.Sprintf("hash file failed: %v", err)})
			continue
		}
		if hash != item.Hash {
			plan.Conflicts = append(plan.Conflicts, Conflict{Path: file.Path, Reason: "mismatched ownership fingerprint"})
			continue
		}
		if hashText(file.Body) != hash {
			plan.Changes = append(plan.Changes, Change{Kind: Update, Path: file.Path})
		}
	}

	for _, file := range prev.Files {
		if _, ok := next[file.Path]; ok {
			continue
		}

		full := filepath.Join(root, file.Path)
		hash, err := hashFile(full)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			plan.Conflicts = append(plan.Conflicts, Conflict{Path: file.Path, Reason: fmt.Sprintf("hash file failed: %v", err)})
			continue
		}
		if hash != file.Hash {
			plan.Conflicts = append(plan.Conflicts, Conflict{Path: file.Path, Reason: "mismatched ownership fingerprint"})
			continue
		}
		plan.Changes = append(plan.Changes, Change{Kind: Delete, Path: file.Path})
	}
}

func inspectSections(root string, prev state, des *model.Desired, plan *Plan) error {
	next := map[sectionState]struct{}{}
	for _, section := range des.Sections {
		next[sectionState{Path: section.Path, ID: section.ID}] = struct{}{}

		path := filepath.Join(root, section.Path)
		raw, err := os.ReadFile(path)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("read section file %q: %w", section.Path, err)
		}

		before := string(raw)
		if hasForeignSectionMarker(before, section.ID) {
			plan.Conflicts = append(plan.Conflicts, Conflict{Path: sectionKey(section.Path, section.ID), Reason: "refusing to overwrite foreign section file"})
			continue
		}
		after := upsertSection(before, section)
		if before == after {
			continue
		}

		kind := Create
		if hasManagedSection(before, section.ID) {
			kind = Update
		}
		plan.Changes = append(plan.Changes, Change{Kind: kind, Path: sectionKey(section.Path, section.ID)})
	}

	for _, section := range prev.Sections {
		if _, ok := next[section]; ok {
			continue
		}

		path := filepath.Join(root, section.Path)
		raw, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("read section file %q: %w", section.Path, err)
		}

		before := string(raw)
		after := removeSection(before, section)
		if before == after {
			continue
		}
		plan.Changes = append(plan.Changes, Change{Kind: Delete, Path: sectionKey(section.Path, section.ID)})
	}

	return nil
}

func loadState(root, target string) (state, error) {
	path := statePath(root, target)
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return state{}, nil
	}
	if err != nil {
		return state{}, fmt.Errorf("read state: %w", err)
	}

	st, err := parseState(path, raw, currentOwner, currentStateVersion)
	if err != nil {
		return state{}, err
	}

	return st, nil
}

func saveState(root, target string, st state) error {
	path := statePath(root, target)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir state dir: %w", err)
	}

	raw, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}

	if err := os.WriteFile(path, raw, 0o644); err != nil {
		return fmt.Errorf("write state: %w", err)
	}

	return nil
}

func statePath(root, target string) string {
	return filepath.Join(root, currentStateDirName, "state", target+".json")
}

func parseState(path string, raw []byte, owner string, version int) (state, error) {
	var st state
	if err := json.Unmarshal(raw, &st); err != nil {
		return state{}, fmt.Errorf("parse state: %w", err)
	}
	if st.Owner != owner || st.Version != version {
		return state{}, fmt.Errorf("foreign state file at %q", path)
	}
	for _, file := range st.Files {
		if !validStateFile(file.Path) {
			return state{}, fmt.Errorf("invalid state file path %q", file.Path)
		}
		if !validHash.MatchString(file.Hash) {
			return state{}, fmt.Errorf("invalid state file hash for %q", file.Path)
		}
	}
	for _, section := range st.Sections {
		if !validStateSection(section) {
			return state{}, fmt.Errorf("invalid state section path %q", section.Path)
		}
	}

	return st, nil
}

func upsertSection(raw string, section model.Section) string {
	start := sectionStart(section.ID)
	end := sectionEnd(section.ID)
	block := start + "\n" + section.Body + end + "\n"

	out, ok := replaceManagedSection(raw, start, end, block)
	if ok {
		return out
	}

	if raw == "" {
		return block
	}
	if !strings.HasSuffix(raw, "\n") {
		raw += "\n"
	}
	return raw + "\n" + block
}

func replaceManagedSection(raw, start, end, block string) (string, bool) {
	i := strings.Index(raw, start)
	j := strings.Index(raw, end)
	if i < 0 || j < 0 || j < i {
		return raw, false
	}

	j += len(end)
	if j < len(raw) && raw[j] == '\n' {
		j++
	}

	return raw[:i] + block + raw[j:], true
}

func owned(files []fileState, path string) (fileState, bool) {
	for _, file := range files {
		if file.Path == path {
			return file, true
		}
	}
	return fileState{}, false
}

func pruneFiles(root string, prev []fileState, next []model.Output) error {
	want := map[string]struct{}{}
	for _, file := range next {
		want[file.Path] = struct{}{}
	}

	for _, file := range prev {
		if _, ok := want[file.Path]; ok {
			continue
		}

		full := filepath.Join(root, file.Path)
		hash, err := hashFile(full)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("hash file %q: %w", file.Path, err)
		}
		if hash != file.Hash {
			return fmt.Errorf("mismatched ownership fingerprint for %q", file.Path)
		}
		if err := os.Remove(full); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove orphan file %q: %w", file.Path, err)
		}
		pruneDirs(root, filepath.Dir(full))
	}

	return nil
}

func pruneSections(root string, prev []sectionState, next []model.Section) error {
	want := map[sectionState]struct{}{}
	for _, section := range next {
		want[sectionState{Path: section.Path, ID: section.ID}] = struct{}{}
	}

	for _, section := range prev {
		if _, ok := want[section]; ok {
			continue
		}

		path := filepath.Join(root, section.Path)
		raw, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("read section file %q: %w", section.Path, err)
		}

		if err := os.WriteFile(path, []byte(removeSection(string(raw), section)), 0o644); err != nil {
			return fmt.Errorf("write section file %q: %w", section.Path, err)
		}
	}

	return nil
}

func removeSection(raw string, section sectionState) string {
	out, ok := removeManagedSection(raw, sectionStart(section.ID), sectionEnd(section.ID))
	if ok {
		return out
	}

	return raw
}

func removeManagedSection(raw, start, end string) (string, bool) {
	i := strings.Index(raw, start)
	j := strings.Index(raw, end)
	if i < 0 || j < 0 || j < i {
		return raw, false
	}

	j += len(end)
	if j < len(raw) && raw[j] == '\n' {
		j++
	}

	out := raw[:i] + raw[j:]
	for strings.Contains(out, "\n\n\n") {
		out = strings.ReplaceAll(out, "\n\n\n", "\n\n")
	}
	if strings.HasSuffix(out, "\n\n") {
		out = strings.TrimSuffix(out, "\n")
	}
	return out, true
}

func pruneDirs(root, dir string) {
	for dir != root && dir != "." {
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) != 0 {
			return
		}
		if err := os.Remove(dir); err != nil {
			return
		}
		dir = filepath.Dir(dir)
	}
}

func validStateFile(path string) bool {
	if path == "" || filepath.IsAbs(path) {
		return false
	}

	clean := filepath.Clean(path)
	if clean != path || clean == "." {
		return false
	}
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return false
	}

	roots := []string{
		filepath.Join(".opencode", "commands") + string(filepath.Separator),
		filepath.Join(".opencode", "agents") + string(filepath.Separator),
		filepath.Join(".agents", "skills") + string(filepath.Separator),
	}
	for _, root := range roots {
		if strings.HasPrefix(clean, root) {
			return true
		}
	}

	return false
}

func validStateSection(section sectionState) bool {
	return section.Path == "AGENTS.md" && validID.MatchString(section.ID)
}

func hashText(body string) string {
	sum := sha256.Sum256([]byte(body))
	return hex.EncodeToString(sum[:])
}

func hashFile(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return hashText(string(raw)), nil
}

func sectionKey(path, id string) string {
	return path + "#" + id
}

func sectionStart(id string) string {
	return sectionMarker(currentSectionPrefix, "start", id)
}

func sectionEnd(id string) string {
	return sectionMarker(currentSectionPrefix, "end", id)
}

func sectionMarker(prefix, edge, id string) string {
	return "<!-- " + prefix + ":section:" + edge + " " + id + " -->"
}

func hasManagedSection(raw, id string) bool {
	return strings.Contains(raw, sectionStart(id))
}

func hasForeignSectionMarker(raw, id string) bool {
	matches := regexp.MustCompile(`<!-- ([a-zA-Z0-9._-]+):section:start ` + regexp.QuoteMeta(id) + ` -->`).FindAllStringSubmatch(raw, -1)
	for _, match := range matches {
		if len(match) == 2 && match[1] != currentSectionPrefix {
			return true
		}
	}

	return false
}
