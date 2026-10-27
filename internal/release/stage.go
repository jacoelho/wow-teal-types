package release

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"wow-teal-types/internal/consumer"
	"wow-teal-types/internal/wowapi"
)

var maintainedDirs = []string{"internal", "tools", "scripts/wow_api", "types/wow", "bin", "tests", "docs", "examples", "rockspecs", ".github/workflows", "evidence/generated", "references/wow-api/native-observations"}
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
func stageNative(root, snapshot string) (M, error) {
	p := wowapi.NewProject(root)
	lock, e := p.Lock("sources.json")
	if e != nil {
		return nil, e
	}
	record, e := VerifySnapshot(snapshot, nil)
	if e != nil {
		return nil, e
	}
	if e = projectVersion(root, record); e != nil {
		return nil, e
	}
	receipt, e := readFile(filepath.Join(snapshot, "snapshot.json"), maxFile)
	if e != nil {
		return nil, e
	}
	files := map[string]bool{}
	var names []string
	prefix := "Interface/AddOns/Blizzard_APIDocumentationGenerated/"
	for _, v := range arr(record["files"]) {
		n := s(obj(v)["path"])
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
		source := obj(lock[group])
		if s(source["kind"]) == "casc" {
			preserved = false
		}
		old := map[string]M{}
		selectedNames := []string{}
		for _, v := range arr(source["files"]) {
			r := obj(v)
			old[s(r["name"])] = r
			selectedNames = append(selectedNames, s(r["name"]))
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
				raw = []byte(s(record["version"]) + "\n")
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
			if r := old[name]; r != nil {
				if len(raw) == i(r["size"]) && blob(raw) == s(r["sha"]) {
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
		old := obj(lock[group])
		rows := []any{}
		for _, r := range acquired[group] {
			data := r.normalized
			if preserved {
				data = r.legacy
			}
			var origin any
			if r.path != "" {
				origin = r.path
			}
			rows = append(rows, M{"name": r.name, "client_path": origin, "sha": blob(data), "size": len(data)})
			if e = atomicWrite(p.InputPath(group, r.name), data, 0644); e != nil {
				return nil, e
			}
		}
		if !preserved {
			lock[group] = M{"kind": "casc", "client_lock_sha256": sha(receipt), "path": old["path"], "normalization": "lf-final-newline", "files": rows}
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

type github struct{ trees map[string]map[string]M }

func (g *github) api(ctx context.Context, path string) (M, error) {
	b, e := fetch(ctx, "https://api.github.com"+path, 16<<20)
	if e != nil {
		return nil, e
	}
	return decodeObject(b)
}
func decodeObject(b []byte) (M, error) {
	var m M
	d := newJSONDecoder(b)
	if e := d.Decode(&m); e != nil {
		return nil, e
	}
	return m, nil
}
func (g *github) resolve(ctx context.Context, repo, ref string) (string, error) {
	m, e := g.api(ctx, "/repos/"+repo+"/commits/"+url.PathEscape(ref))
	if e != nil {
		return "", e
	}
	revision := s(m["sha"])
	if !hex40.MatchString(revision) {
		return "", fmt.Errorf("invalid resolved commit: %s", repo)
	}
	return revision, nil
}
func (g *github) tree(ctx context.Context, repo, revision string) (map[string]M, error) {
	if !hex40.MatchString(revision) {
		return nil, fmt.Errorf("inventory requires immutable revision")
	}
	key := repo + "@" + revision
	if g.trees == nil {
		g.trees = map[string]map[string]M{}
	}
	if rows := g.trees[key]; rows != nil {
		return rows, nil
	}
	m, e := g.api(ctx, "/repos/"+repo+"/git/trees/"+revision+"?recursive=1")
	if e != nil {
		return nil, e
	}
	if m["truncated"] != false || len(arr(m["tree"])) > 20000 {
		return nil, fmt.Errorf("incomplete or oversized recursive tree: %s", repo)
	}
	rows := map[string]M{}
	for _, v := range arr(m["tree"]) {
		r := obj(v)
		name := s(r["path"])
		if _, e = safeJoin("/", name); e != nil {
			return nil, e
		}
		if rows[name] != nil || !hex40.MatchString(s(r["sha"])) {
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
	row := tree[name]
	if row == nil || row["type"] != "blob" || (row["mode"] != "100644" && row["mode"] != "100755") || i(row["size"]) > limit {
		return nil, fmt.Errorf("missing or unsupported source entry: %s", name)
	}
	b, e := fetch(ctx, "https://raw.githubusercontent.com/"+repo+"/"+revision+"/"+escapePath(name), int64(limit))
	if e != nil {
		return nil, e
	}
	if len(b) != i(row["size"]) || blob(b) != s(row["sha"]) {
		return nil, fmt.Errorf("Git blob identity mismatch: %s", name)
	}
	return b, nil
}
func (g *github) inventory(ctx context.Context, repo, revision, prefix string) ([]any, error) {
	tree, e := g.tree(ctx, repo, revision)
	if e != nil {
		return nil, e
	}
	prefix = strings.TrimSuffix(prefix, "/") + "/"
	rows := []any{}
	for _, name := range keys(tree) {
		r := tree[name]
		if !strings.HasPrefix(name, prefix) || r["type"] == "tree" {
			continue
		}
		if r["type"] != "blob" || (r["mode"] != "100644" && r["mode"] != "100755") {
			return nil, fmt.Errorf("unsupported source inventory entry: %s", name)
		}
		rows = append(rows, M{"name": strings.TrimPrefix(name, prefix), "sha": r["sha"], "size": r["size"]})
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("empty upstream source inventory")
	}
	return rows, nil
}
func refreshCommunity(ctx context.Context, root string, g *github, refs map[string]string) error {
	p := wowapi.NewProject(root)
	lock, e := p.Lock("sources.json")
	if e != nil {
		return e
	}
	revisionByRepo := map[string]string{}
	for _, group := range keys(lock) {
		if group == "blizzard" || group == "blizzard_support" || group == "blizzard_version" {
			continue
		}
		source := obj(lock[group])
		repo := s(source["repository"])
		revision := revisionByRepo[repo]
		if revision == "" {
			ref := refs["questie"]
			if ref == "" {
				ref = s(source["ref"])
			}
			if ref == "" {
				ref = s(source["commit"])
			}
			revision, e = g.resolve(ctx, repo, ref)
			if e != nil {
				return e
			}
			revisionByRepo[repo] = revision
			source["ref"] = ref
		}
		source["commit"] = revision
		if group == "questie" {
			source["files"], e = g.inventory(ctx, repo, revision, s(source["path"]))
			if e != nil {
				return e
			}
		} else {
			tree, err := g.tree(ctx, repo, revision)
			if err != nil {
				return err
			}
			rows := []any{}
			for _, v := range arr(source["files"]) {
				old := obj(v)
				name := s(old["name"])
				entry := tree[strings.Trim(s(source["path"])+"/"+name, "/")]
				if entry == nil {
					return fmt.Errorf("required supporting source removed: %s", name)
				}
				rows = append(rows, M{"name": name, "sha": entry["sha"], "size": entry["size"]})
			}
			source["files"] = rows
		}
		for _, v := range arr(source["files"]) {
			row := obj(v)
			name := s(row["name"])
			b, e := g.file(ctx, repo, revision, strings.Trim(s(source["path"])+"/"+name, "/"), maxFile)
			if e != nil {
				return e
			}
			if e = atomicWrite(p.InputPath(group, name), b, 0644); e != nil {
				return e
			}
		}
	}
	return writeObject(filepath.Join(root, "scripts/wow_api/sources.json"), lock)
}

// StageRelease acquires an immutable candidate, including explicit third-party revisions.
// It never updates approval ledgers or active outputs.
func StageRelease(ctx context.Context, root, output, snapshot string, refs map[string]string) (M, error) {
	before, e := captureInputs(root)
	if e != nil {
		return nil, e
	}
	if exists(output) {
		return nil, fmt.Errorf("candidate destination already exists")
	}
	workspace := filepath.Join(output, "workspace")
	if e = copyMaintained(root, workspace); e != nil {
		return nil, e
	}
	failed := true
	defer func() {
		if failed {
			_ = writeObject(filepath.Join(output, "failure.json"), M{"status": "failed", "workspace": workspace})
		}
	}()
	if e = seedCaches(root, workspace); e != nil {
		return nil, e
	}
	native, e := stageNative(workspace, snapshot)
	if e != nil {
		return nil, e
	}
	g := &github{}
	if e = refreshCommunity(ctx, workspace, g, refs); e != nil {
		return nil, e
	}
	blockers := []any{}
	if e = refreshTSM(ctx, workspace, g, refs["tsm"], &blockers); e != nil {
		return nil, e
	}
	if e = refreshSecondary(ctx, workspace, g, refs, &blockers); e != nil {
		return nil, e
	}
	if e = refreshBackup(ctx, workspace, g, refs["backup"], native, &blockers); e != nil {
		return nil, e
	}
	if e = requireUnchanged(root, before); e != nil {
		return nil, e
	}
	report := M{"status": "acquired-review-required", "workspace": workspace, "nativeBuild": s(obj(native["blizzard_version"])["version"]), "reviewRequired": true, "blockers": blockers, "validationRequired": []any{"schema-matched reader build", "resolved binding regeneration when changed", "exact reconciliation review", "deterministic generation and compiler checks", "maintainer and external consumer checks", "build-matched native evidence"}}
	v, e := consumerVersion(workspace)
	if e != nil {
		return nil, e
	}
	report["nativeBuild"] = v
	for key, tree := range g.trees {
		if e = writeObject(filepath.Join(output, "upstream-inventories", strings.ReplaceAll(strings.ReplaceAll(key, "/", "--"), "@", "-")+".json"), tree); e != nil {
			return nil, e
		}
	}
	if e = writeObject(filepath.Join(output, "release-report.json"), report); e != nil {
		return nil, e
	}
	failed = false
	return report, nil
}
func refreshTSM(ctx context.Context, root string, g *github, ref string, blockers *[]any) error {
	path := filepath.Join(root, "scripts/wow_api/tsm_sources.json")
	lock, e := readObject(path)
	if e != nil {
		return e
	}
	old := s(lock["commit"])
	repo := s(lock["repository"])
	if ref == "" {
		r, e := g.api(ctx, "/repos/"+repo+"/releases/latest")
		if e != nil {
			return e
		}
		ref = s(r["tag_name"])
		if ref == "" {
			return fmt.Errorf("TSM latest release lacks tag")
		}
	}
	rev, e := g.resolve(ctx, repo, ref)
	if e != nil {
		return e
	}
	payloads := map[string][]byte{}
	for _, v := range arr(lock["files"]) {
		row := obj(v)
		up := s(row["path"])
		if up == "" {
			u, e := url.Parse(s(row["url"]))
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
		row["path"] = up
		delete(row, "url")
		row["sha256"] = sha(b)
		row["size"] = len(b)
		payloads[s(row["name"])] = b
		if e = atomicWrite(filepath.Join(root, "references/wow-api/tsm", s(row["name"])), b, 0644); e != nil {
			return e
		}
	}
	provenance := string(payloads["precomputed-provenance.txt"])
	match := regexp.MustCompile(`(?m)^# Source: https://github\.com/Gethe/wow-ui-source\.git @ ([0-9a-f]{7,40}) \(branch: live\)$`).FindAllStringSubmatch(provenance, -1)
	counts := regexp.MustCompile(`(?m)^# Embedded source files: ([0-9]+)$`).FindAllStringSubmatch(provenance, -1)
	if len(match) != 1 || len(counts) != 1 {
		return fmt.Errorf("TSM provenance lacks Retail revision/file count")
	}
	baseline := obj(lock["blizzard"])
	baseline["commit"], e = g.resolve(ctx, s(baseline["repository"]), match[0][1])
	if e != nil {
		return e
	}
	version, e := g.file(ctx, s(baseline["repository"]), s(baseline["commit"]), "version.txt", maxFile)
	if e != nil {
		return e
	}
	baseline["build"] = strings.TrimSpace(string(version))
	baseline["version_sha256"] = sha(version)
	baseline["files"], e = g.inventory(ctx, s(baseline["repository"]), s(baseline["commit"]), s(baseline["path"]))
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
	lock["blob_version"] = 44
	var count int
	fmt.Sscanf(counts[0][1], "%d", &count)
	lock["embedded_files"] = count
	lock["commit"] = rev
	lock["release"] = ref
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
func refreshSecondary(ctx context.Context, root string, g *github, refs map[string]string, blockers *[]any) error {
	path := filepath.Join(root, "scripts/wow_api/native_crosscheck_sources.json")
	lock, e := readObject(path)
	if e != nil {
		return e
	}
	revisions := map[string]string{}
	for _, name := range keys(lock) {
		cohort := obj(lock[name])
		repo := s(cohort["repository"])
		if repo == "" || s(cohort["commit"]) == "" {
			continue
		}
		ref := refs[name]
		if ref == "" {
			ref = s(cohort["ref"])
		}
		if ref == "" {
			ref = s(cohort["commit"])
		}
		rev := revisions[repo]
		if rev == "" {
			rev, e = g.resolve(ctx, repo, ref)
			if e != nil {
				return e
			}
			revisions[repo] = rev
		}
		cohort["commit"] = rev
		if name == "blizzard" {
			b, e := g.file(ctx, repo, rev, "version.txt", maxFile)
			if e != nil {
				return e
			}
			cohort["build"] = strings.TrimSpace(string(b))
			v, e := consumerVersion(root)
			if e != nil {
				return e
			}
			if s(cohort["build"]) != v {
				*blockers = append(*blockers, "Gethe secondary build differs from selected client")
			}
		}
	}
	for _, v := range arr(lock["files"]) {
		row := obj(v)
		repo := s(row["repository"])
		rev := revisions[repo]
		b, e := g.file(ctx, repo, rev, s(row["upstream_path"]), maxFile)
		if e != nil {
			return e
		}
		row["commit"] = rev
		row["size"] = len(b)
		row["sha256"] = sha(b)
		row["git_blob_sha1"] = blob(b)
		row["url"] = "https://raw.githubusercontent.com/" + repo + "/" + rev + "/" + s(row["upstream_path"])
		if e = atomicWrite(filepath.Join(root, "references/wow-api/native-crosscheck", s(row["name"])), b, 0644); e != nil {
			return e
		}
	}
	if cohort := obj(lock["wowless"]); cohort != nil {
		data, err := readFile(filepath.Join(root, "references/wow-api/native-crosscheck/wowless/data/products/wow/build.yaml"), maxFile)
		if err != nil {
			return err
		}
		value, err := wowapi.ParseEvidenceYAML(string(data))
		if err != nil {
			return err
		}
		build := obj(value)
		cohort["build"] = fmt.Sprint(build["version"]) + "." + fmt.Sprint(build["build"])
	}
	return writeObject(path, lock)
}
func refreshBackup(ctx context.Context, root string, g *github, ref string, native M, blockers *[]any) error {
	path := filepath.Join(root, "scripts/wow_api/native_backup_sources.json")
	lock, e := readObject(path)
	if e != nil {
		return e
	}
	if ref == "" {
		ref = s(lock["commit"])
	}
	rev, e := g.resolve(ctx, s(lock["repository"]), ref)
	if e != nil {
		return e
	}
	changed := rev != s(lock["commit"])
	lock["commit"] = rev
	lock["inventory"], e = g.inventory(ctx, s(lock["repository"]), rev, s(lock["path"]))
	if e != nil {
		return e
	}
	if changed {
		lock["build_evidence"] = M{"state": "requires-build-review", "previous_build": lock["claimed_build"]}
		lock["claimed_build"] = nil
		*blockers = append(*blockers, "review changed backup build provenance")
	}
	for _, v := range arr(lock["files"]) {
		row := obj(v)
		b, e := g.file(ctx, s(lock["repository"]), rev, s(lock["path"])+"/"+s(row["name"]), maxFile)
		if e != nil {
			return e
		}
		row["sha"] = blob(b)
		row["sha256"] = sha(b)
		row["size"] = len(b)
		if e = atomicWrite(filepath.Join(root, "references/wow-api/native-backup", s(row["name"])), b, 0644); e != nil {
			return e
		}
	}
	return writeObject(path, lock)
}

func consumerVersion(root string) (string, error) {
	v, e := consumer.ReadVersion(root)
	return v.Full, e
}
