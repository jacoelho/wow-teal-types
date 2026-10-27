package release

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"go.yaml.in/yaml/v3"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"wow-teal-types/internal/consumer"
	"wow-teal-types/internal/sourcelock"
	"wow-teal-types/internal/wowapi"
)

var maintainedDirs = []string{"internal", "tools", "scripts/wow_api", "types/wow", "bin", "tests", "docs", "examples", "rockspecs", ".github/workflows", "evidence/generated", "evidence/reviews", "references/wow-api/native-observations"}
var maintainedFiles = []string{".wow-version", "go.mod", "go.sum", "tlconfig.lua", "LICENSE", "README.md", "UPDATING.md", "package.json", "provenance.json", "Makefile", "check-all.sh", "references/wow-api/tsm/resolved-bindings.json"}

func copyMaintained(root, target string) error {
	if exists(target) {
		return fmt.Errorf("candidate output already exists: %s", target)
	}
	if e := os.MkdirAll(target, 0755); e != nil {
		return e
	}
	for _, name := range maintainedDirs {
		src := filepath.Join(root, filepath.FromSlash(name))
		if exists(src) {
			if e := copyDirectory(src, filepath.Join(target, filepath.FromSlash(name))); e != nil {
				return e
			}
		}
	}
	for _, name := range maintainedFiles {
		src, e := safeJoin(root, name)
		if e != nil {
			return e
		}
		if exists(src) {
			dst, e := safeJoin(target, name)
			if e != nil {
				return e
			}
			if e = copyFile(src, dst); e != nil {
				return e
			}
		}
	}
	return nil
}
func sourcePaths(p wowapi.Project, group, name string) string { return p.InputPath(group, name) }
func stageNative(root, snapshot string) (sourcelock.Sources, error) {
	p := wowapi.NewProject(root)
	lock, e := sourcelock.ReadSources(filepath.Join(root, "scripts/wow_api/sources.json"))
	if e != nil {
		return nil, e
	}
	record, e := VerifySnapshot(snapshot, nil)
	if e != nil {
		return nil, e
	}
	if e = projectVersion(root, record.Identity()); e != nil {
		return nil, e
	}
	receipt, e := readFile(filepath.Join(snapshot, "snapshot.json"), maxFile)
	if e != nil {
		return nil, e
	}
	files := map[string]bool{}
	var names []string
	prefix := "Interface/AddOns/Blizzard_APIDocumentationGenerated/"
	for _, row := range record.Files {
		n := row.Path
		files[n] = true
		if strings.HasPrefix(n, prefix) {
			names = append(names, strings.TrimPrefix(n, prefix))
		}
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("empty packaged native inventory")
	}
	sortStrings(names)
	type selected struct {
		name, path         string
		normalized, legacy []byte
	}
	acquired := map[string][]selected{}
	preserved := true
	for _, group := range []string{"blizzard", "blizzard_support", "blizzard_version"} {
		source := lock[group]
		if source.Kind == "casc" {
			preserved = false
		}
		old := map[string]sourcelock.SourceFile{}
		selectedNames := []string{}
		for _, r := range source.Files {
			old[r.Name] = r
			selectedNames = append(selectedNames, r.Name)
		}
		if group == "blizzard" {
			selectedNames = names
		}
		if len(selectedNames) != len(old) {
			preserved = false
		}
		for _, name := range selectedNames {
			var packaged string
			var raw []byte
			if group == "blizzard_version" {
				raw = []byte(record.Version + "\n")
			} else {
				relative := name
				if group == "blizzard" {
					relative = "Blizzard_APIDocumentationGenerated/" + name
				}
				packaged, e = clientPath(relative, files)
				if e != nil {
					return nil, e
				}
				if packaged == "" {
					return nil, fmt.Errorf("required support removed: %s", name)
				}
				raw, e = readFile(filepath.Join(snapshot, "raw", filepath.FromSlash(packaged)), maxFile)
				if e != nil {
					return nil, e
				}
			}
			var legacy []byte
			if r, ok := old[name]; ok {
				if len(raw) == r.Size && blob(raw) == r.SHA {
					legacy = raw
				} else {
					legacy = reviewedBytes(raw, r)
				}
			}
			if legacy == nil {
				preserved = false
			}
			acquired[group] = append(acquired[group], selected{name, packaged, normalized(raw, 1), legacy})
		}
	}
	for _, group := range []string{"blizzard", "blizzard_support", "blizzard_version"} {
		old := lock[group]
		rows := []sourcelock.SourceFile{}
		for _, r := range acquired[group] {
			data := r.normalized
			if preserved {
				data = r.legacy
			}
			origin := sourcelock.ClientPath(r.path)
			rows = append(rows, sourcelock.SourceFile{ClientPath: origin, Name: r.name, SHA: blob(data), Size: len(data)})
			if e = atomicWrite(p.InputPath(group, r.name), data, 0644); e != nil {
				return nil, e
			}
		}
		if !preserved {
			lock[group] = sourcelock.SourceGroup{Kind: "casc", ClientLockSHA256: sha(receipt), Path: old.Path, Normalization: "lf-final-newline", Files: rows}
		}
	}
	if e = atomicWrite(filepath.Join(root, "scripts/wow_api/client_sources.json"), receipt, 0644); e != nil {
		return nil, e
	}
	if e = writeObject(filepath.Join(root, "scripts/wow_api/sources.json"), lock); e != nil {
		return nil, e
	}
	return lock, nil
}
func sortStrings(v []string) {
	for n := 1; n < len(v); n++ {
		for j := n; j > 0 && v[j] < v[j-1]; j-- {
			v[j], v[j-1] = v[j-1], v[j]
		}
	}
}

type githubEntry struct {
	Mode string `json:"mode"`
	Path string `json:"path"`
	SHA  string `json:"sha"`
	Size *int   `json:"size,omitempty"`
	Type string `json:"type"`
	URL  string `json:"url"`
}
type githubResponse struct {
	SHA       string        `json:"sha"`
	TagName   string        `json:"tag_name"`
	Tree      []githubEntry `json:"tree"`
	Truncated *bool         `json:"truncated"`
}
type github struct {
	trees map[string]map[string]githubEntry
}

func (g *github) api(ctx context.Context, path string) (githubResponse, error) {
	data, err := fetch(ctx, "https://api.github.com"+path, 16<<20)
	if err != nil {
		return githubResponse{}, err
	}
	// GitHub owns extra response fields; only this typed projection enters acquisition.
	var response *githubResponse
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err = decoder.Decode(&response); err != nil {
		return githubResponse{}, err
	}
	if err = decoder.Decode(new(json.RawMessage)); err != io.EOF {
		return githubResponse{}, fmt.Errorf("trailing GitHub JSON")
	}
	if response == nil {
		return githubResponse{}, fmt.Errorf("expected GitHub response object")
	}
	return *response, nil
}
func (g *github) resolve(ctx context.Context, repo, ref string) (string, error) {
	m, e := g.api(ctx, "/repos/"+repo+"/commits/"+url.PathEscape(ref))
	if e != nil {
		return "", e
	}
	revision := m.SHA
	if !hex40.MatchString(revision) {
		return "", fmt.Errorf("invalid resolved commit: %s", repo)
	}
	return revision, nil
}
func (g *github) tree(ctx context.Context, repo, revision string) (map[string]githubEntry, error) {
	if !hex40.MatchString(revision) {
		return nil, fmt.Errorf("inventory requires immutable revision")
	}
	key := repo + "@" + revision
	if g.trees == nil {
		g.trees = map[string]map[string]githubEntry{}
	}
	if rows := g.trees[key]; rows != nil {
		return rows, nil
	}
	m, e := g.api(ctx, "/repos/"+repo+"/git/trees/"+revision+"?recursive=1")
	if e != nil {
		return nil, e
	}
	if m.Truncated == nil || *m.Truncated || m.Tree == nil || len(m.Tree) > 20000 {
		return nil, fmt.Errorf("incomplete or oversized recursive tree: %s", repo)
	}
	rows := map[string]githubEntry{}
	for _, r := range m.Tree {
		name := r.Path
		if _, e = safeJoin("/", name); e != nil {
			return nil, e
		}
		if _, duplicate := rows[name]; duplicate || !hex40.MatchString(r.SHA) || (r.Type == "blob" && (r.Size == nil || *r.Size < 0)) {
			return nil, fmt.Errorf("invalid recursive tree row: %s", name)
		}
		rows[name] = r
	}
	g.trees[key] = rows
	return rows, nil
}
func (g *github) file(ctx context.Context, repo, revision, name string, limit int) ([]byte, error) {
	tree, e := g.tree(ctx, repo, revision)
	if e != nil {
		return nil, e
	}
	row, found := tree[name]
	if !found || row.Type != "blob" || (row.Mode != "100644" && row.Mode != "100755") || row.Size == nil || *row.Size > limit {
		return nil, fmt.Errorf("missing or unsupported source entry: %s", name)
	}
	b, e := fetch(ctx, "https://raw.githubusercontent.com/"+repo+"/"+revision+"/"+escapePath(name), int64(limit))
	if e != nil {
		return nil, e
	}
	if len(b) != *row.Size || blob(b) != row.SHA {
		return nil, fmt.Errorf("Git blob identity mismatch: %s", name)
	}
	return b, nil
}
func (g *github) inventory(ctx context.Context, repo, revision, prefix string) ([]sourcelock.SourceFile, error) {
	tree, e := g.tree(ctx, repo, revision)
	if e != nil {
		return nil, e
	}
	prefix = strings.TrimSuffix(prefix, "/") + "/"
	rows := []sourcelock.SourceFile{}
	for _, name := range keys(tree) {
		r := tree[name]
		if !strings.HasPrefix(name, prefix) || r.Type == "tree" {
			continue
		}
		if r.Type != "blob" || (r.Mode != "100644" && r.Mode != "100755") {
			return nil, fmt.Errorf("unsupported source inventory entry: %s", name)
		}
		rows = append(rows, sourcelock.SourceFile{Name: strings.TrimPrefix(name, prefix), SHA: r.SHA, Size: *r.Size})
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("empty upstream source inventory")
	}
	return rows, nil
}
func refreshCommunity(ctx context.Context, root string, g *github, refs map[string]string) error {
	p := wowapi.NewProject(root)
	lock, e := sourcelock.ReadSources(filepath.Join(root, "scripts/wow_api/sources.json"))
	if e != nil {
		return e
	}
	revisionByRepo := map[string]string{}
	for _, group := range keys(lock) {
		if group == "blizzard" || group == "blizzard_support" || group == "blizzard_version" {
			continue
		}
		source := lock[group]
		repo := source.Repository
		revision := revisionByRepo[repo]
		if revision == "" {
			ref := refs["questie"]
			if ref == "" {
				ref = source.Ref
			}
			if ref == "" {
				ref = source.Commit
			}
			revision, e = g.resolve(ctx, repo, ref)
			if e != nil {
				return e
			}
			revisionByRepo[repo] = revision
			source.Ref = ref
		}
		source.Commit = revision
		if group == "questie" {
			source.Files, e = g.inventory(ctx, repo, revision, source.Path)
			if e != nil {
				return e
			}
		} else {
			tree, err := g.tree(ctx, repo, revision)
			if err != nil {
				return err
			}
			rows := []sourcelock.SourceFile{}
			for _, old := range source.Files {
				name := old.Name
				entry, found := tree[strings.Trim(source.Path+"/"+name, "/")]
				if !found || entry.Size == nil {
					return fmt.Errorf("required supporting source removed: %s", name)
				}
				rows = append(rows, sourcelock.SourceFile{Name: name, SHA: entry.SHA, Size: *entry.Size})
			}
			source.Files = rows
		}
		for _, row := range source.Files {
			name := row.Name
			b, e := g.file(ctx, repo, revision, strings.Trim(source.Path+"/"+name, "/"), maxFile)
			if e != nil {
				return e
			}
			if e = atomicWrite(p.InputPath(group, name), b, 0644); e != nil {
				return e
			}
		}
		lock[group] = source
	}
	return writeObject(filepath.Join(root, "scripts/wow_api/sources.json"), lock)
}

// StageRelease acquires an immutable candidate, including explicit third-party revisions.
// It never updates approval ledgers or active outputs.
func StageRelease(ctx context.Context, root, output, snapshot string, refs map[string]string) (StageResult, error) {
	before, e := captureInputs(root)
	if e != nil {
		return StageResult{}, e
	}
	if exists(output) {
		return StageResult{}, fmt.Errorf("candidate destination already exists")
	}
	workspace := filepath.Join(output, "workspace")
	if e = copyMaintained(root, workspace); e != nil {
		return StageResult{}, e
	}
	failed := true
	defer func() {
		if failed {
			_ = writeObject(filepath.Join(output, "failure.json"), stageFailure{Status: "failed", Workspace: workspace})
		}
	}()
	if e = seedCaches(root, workspace); e != nil {
		return StageResult{}, e
	}
	native, e := stageNative(workspace, snapshot)
	if e != nil {
		return StageResult{}, e
	}
	g := &github{}
	if e = refreshCommunity(ctx, workspace, g, refs); e != nil {
		return StageResult{}, e
	}
	blockers := []string{}
	if e = refreshTSM(ctx, workspace, g, refs["tsm"], &blockers); e != nil {
		return StageResult{}, e
	}
	if e = refreshSecondary(ctx, workspace, g, refs, &blockers); e != nil {
		return StageResult{}, e
	}
	if e = refreshBackup(ctx, workspace, g, refs["backup"], native, &blockers); e != nil {
		return StageResult{}, e
	}
	if e = requireUnchanged(root, before); e != nil {
		return StageResult{}, e
	}
	report := StageResult{Status: "acquired-review-required", Workspace: workspace, ReviewRequired: true, Blockers: blockers, ValidationRequired: []string{"schema-matched reader build", "resolved binding regeneration when changed", "exact reconciliation review", "deterministic generation and compiler checks", "maintainer and external consumer checks", "build-matched native evidence"}}
	v, e := consumerVersion(workspace)
	if e != nil {
		return StageResult{}, e
	}
	report.NativeBuild = v
	for key, tree := range g.trees {
		if e = writeObject(filepath.Join(output, "upstream-inventories", strings.ReplaceAll(strings.ReplaceAll(key, "/", "--"), "@", "-")+".json"), tree); e != nil {
			return StageResult{}, e
		}
	}
	if e = writeObject(filepath.Join(output, "release-report.json"), report); e != nil {
		return StageResult{}, e
	}
	failed = false
	return report, nil
}
func refreshTSM(ctx context.Context, root string, g *github, ref string, blockers *[]string) error {
	path := filepath.Join(root, "scripts/wow_api/tsm_sources.json")
	lock, e := sourcelock.ReadTSM(path)
	if e != nil {
		return e
	}
	old := lock.Commit
	repo := lock.Repository
	if ref == "" {
		r, e := g.api(ctx, "/repos/"+repo+"/releases/latest")
		if e != nil {
			return e
		}
		ref = r.TagName
		if ref == "" {
			return fmt.Errorf("TSM latest release lacks tag")
		}
	}
	rev, e := g.resolve(ctx, repo, ref)
	if e != nil {
		return e
	}
	payloads := map[string][]byte{}
	for index := range lock.Files {
		row := &lock.Files[index]
		up := row.Path
		if up == "" {
			u, e := url.Parse(row.URL)
			if e != nil {
				return e
			}
			parts := strings.SplitN(strings.TrimPrefix(u.Path, "/"), "/", 4)
			if len(parts) != 4 {
				return fmt.Errorf("invalid TSM upstream URL")
			}
			up = parts[3]
		}
		b, e := g.file(ctx, repo, rev, up, maxPayload)
		if e != nil {
			return e
		}
		if bytes.HasPrefix(b, []byte("version https://git-lfs.github.com/")) {
			parts := strings.Split(string(b), "\n")
			if len(parts) != 4 || !strings.HasPrefix(parts[1], "oid sha256:") {
				return fmt.Errorf("invalid TSM LFS pointer")
			}
			expected := strings.TrimPrefix(parts[1], "oid sha256:")
			b, e = fetch(ctx, "https://media.githubusercontent.com/media/"+repo+"/"+rev+"/"+escapePath(up), maxPayload)
			if e != nil {
				return e
			}
			if sha(b) != expected {
				return fmt.Errorf("TSM LFS identity mismatch")
			}
		}
		row.Path = up
		row.URL = ""
		row.SHA256 = sha(b)
		row.Size = len(b)
		payloads[row.Name] = b
		if e = atomicWrite(filepath.Join(root, "references/wow-api/tsm", row.Name), b, 0644); e != nil {
			return e
		}
	}
	provenance := string(payloads["precomputed-provenance.txt"])
	match := regexp.MustCompile(`(?m)^# Source: https://github\.com/Gethe/wow-ui-source\.git @ ([0-9a-f]{7,40}) \(branch: live\)$`).FindAllStringSubmatch(provenance, -1)
	counts := regexp.MustCompile(`(?m)^# Embedded source files: ([0-9]+)$`).FindAllStringSubmatch(provenance, -1)
	if len(match) != 1 || len(counts) != 1 {
		return fmt.Errorf("TSM provenance lacks Retail revision/file count")
	}
	baseline := &lock.Blizzard
	baseline.Commit, e = g.resolve(ctx, baseline.Repository, match[0][1])
	if e != nil {
		return e
	}
	version, e := g.file(ctx, baseline.Repository, baseline.Commit, "version.txt", maxFile)
	if e != nil {
		return e
	}
	baseline.Build = strings.TrimSpace(string(version))
	baseline.VersionSHA256 = sha(version)
	baseline.Files, e = g.inventory(ctx, baseline.Repository, baseline.Commit, baseline.Path)
	if e != nil {
		return e
	}
	if e = atomicWrite(filepath.Join(root, "references/wow-api/tsm/blizzard-version.txt"), version, 0644); e != nil {
		return e
	}
	files, resolved := payloads["precomputed-files.bin.zst"], payloads["precomputed.bin.zst"]
	if len(files) < 4 || len(resolved) < 8 || string(resolved[:4]) != "_WOW" || binary.LittleEndian.Uint32(files) != 44 || binary.LittleEndian.Uint32(resolved[4:]) != 44 {
		return fmt.Errorf("unsupported TSM serialized schema; review reader")
	}
	lock.BlobVersion = 44
	var count int
	fmt.Sscanf(counts[0][1], "%d", &count)
	lock.EmbeddedFiles = count
	lock.Commit = rev
	lock.Release = ref
	if rev != old {
		cargo := filepath.Join(root, "scripts/wow_api/tsm_resolved_export/Cargo.toml")
		b, e := readFile(cargo, maxFile)
		if e != nil {
			return e
		}
		needle := "rev = \"" + old + "\""
		if strings.Count(string(b), needle) != 2 {
			return fmt.Errorf("TSM Cargo dependency requires manual review")
		}
		if e = atomicWrite(cargo, []byte(strings.ReplaceAll(string(b), needle, "rev = \""+rev+"\"")), 0644); e != nil {
			return e
		}
		*blockers = append(*blockers, "regenerate staged Cargo.lock and review schema-matched reader", "refresh selected binding evidence and exact reconciliation decisions")
	}
	return writeObject(path, lock)
}
func refreshSecondary(ctx context.Context, root string, g *github, refs map[string]string, blockers *[]string) error {
	path := filepath.Join(root, "scripts/wow_api/native_crosscheck_sources.json")
	lock, e := sourcelock.ReadCrosscheck(path)
	if e != nil {
		return e
	}
	revisions := map[string]string{}
	for _, name := range []string{"blizzard", "wowless"} {
		cohort := &lock.Blizzard
		if name == "wowless" {
			cohort = &lock.Wowless
		}
		repo := cohort.Repository
		ref := refs[name]
		if ref == "" {
			ref = cohort.Ref
		}
		if ref == "" {
			ref = cohort.Commit
		}
		rev := revisions[repo]
		if rev == "" {
			rev, e = g.resolve(ctx, repo, ref)
			if e != nil {
				return e
			}
			revisions[repo] = rev
		}
		cohort.Commit = rev
		if name == "blizzard" {
			b, e := g.file(ctx, repo, rev, "version.txt", maxFile)
			if e != nil {
				return e
			}
			cohort.Build = strings.TrimSpace(string(b))
			v, e := consumerVersion(root)
			if e != nil {
				return e
			}
			if cohort.Build != v {
				*blockers = append(*blockers, "Gethe secondary build differs from selected client")
			}
		}
	}
	for index := range lock.Files {
		row := &lock.Files[index]
		repo := row.Repository
		rev := revisions[repo]
		b, e := g.file(ctx, repo, rev, row.UpstreamPath, maxFile)
		if e != nil {
			return e
		}
		row.Commit = rev
		row.Size = len(b)
		row.SHA256 = sha(b)
		row.GitBlobSHA1 = blob(b)
		row.URL = "https://raw.githubusercontent.com/" + repo + "/" + rev + "/" + row.UpstreamPath
		if e = atomicWrite(filepath.Join(root, "references/wow-api/native-crosscheck", row.Name), b, 0644); e != nil {
			return e
		}
	}
	data, err := readFile(filepath.Join(root, "references/wow-api/native-crosscheck/wowless/data/products/wow/build.yaml"), maxFile)
	if err != nil {
		return err
	}
	var build struct {
		Build      string `yaml:"build"`
		Date       string `yaml:"date"`
		Hash       string `yaml:"hash"`
		Test       bool   `yaml:"test"`
		TOCVersion int    `yaml:"tocversion"`
		Version    string `yaml:"version"`
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err = decoder.Decode(&build); err != nil {
		return err
	}
	if build.Version == "" || build.Build == "" {
		return fmt.Errorf("wowless build lacks version identity")
	}
	lock.Wowless.Build = build.Version + "." + build.Build
	return writeObject(path, lock)
}
func refreshBackup(ctx context.Context, root string, g *github, ref string, native sourcelock.Sources, blockers *[]string) error {
	path := filepath.Join(root, "scripts/wow_api/native_backup_sources.json")
	lock, e := sourcelock.ReadBackup(path)
	if e != nil {
		return e
	}
	if ref == "" {
		ref = lock.Commit
	}
	rev, e := g.resolve(ctx, lock.Repository, ref)
	if e != nil {
		return e
	}
	changed := rev != lock.Commit
	lock.Commit = rev
	lock.Inventory, e = g.inventory(ctx, lock.Repository, rev, lock.Path)
	if e != nil {
		return e
	}
	if changed {
		lock.BuildEvidence = sourcelock.BuildEvidence{State: "requires-build-review", PreviousBuild: lock.ClaimedBuild}
		lock.ClaimedBuild = nil
		*blockers = append(*blockers, "review changed backup build provenance")
	}
	for index := range lock.Files {
		row := &lock.Files[index]
		b, e := g.file(ctx, lock.Repository, rev, lock.Path+"/"+row.Name, maxFile)
		if e != nil {
			return e
		}
		row.SHA = blob(b)
		row.SHA256 = sha(b)
		row.Size = len(b)
		if e = atomicWrite(filepath.Join(root, "references/wow-api/native-backup", row.Name), b, 0644); e != nil {
			return e
		}
	}
	return writeObject(path, lock)
}

func consumerVersion(root string) (string, error) {
	v, e := consumer.ReadVersion(root)
	return v.Full, e
}
