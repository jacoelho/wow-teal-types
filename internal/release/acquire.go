package release

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"wow-teal-types/internal/consumer"
	"wow-teal-types/internal/sourcelock"
	"wow-teal-types/internal/wowapi"
)

var manifests = map[string]int{"Interface/ui-code-list.txt": 6067012, "Interface/ui-toc-list.txt": 6067013, "Interface/ui-gen-addon-list.txt": 6076661}
var hex40 = regexp.MustCompile(`^[0-9a-f]{40}$`)

func clientPath(name string, files map[string]bool) (string, error) {
	exact := "Interface/AddOns/" + name
	if files[exact] {
		return exact, nil
	}
	addon, _, _ := strings.Cut(name, "/")
	var found string
	for p := range files {
		if strings.HasPrefix(p, "Interface/AddOns/"+addon+"/") && filepath.Base(p) == filepath.Base(name) {
			if found != "" {
				return "", fmt.Errorf("ambiguous packaged client path: %s", name)
			}
			found = p
		}
	}
	return found, nil
}
func manifestPaths(raw string) ([]string, error) {
	found := map[string]string{}
	for _, name := range keys(manifests) {
		path, e := safeJoin(raw, name)
		if e != nil {
			return nil, e
		}
		b, e := readFile(path, maxFile)
		if e != nil {
			return nil, e
		}
		seen := map[string]bool{}
		for line := range strings.SplitSeq(strings.TrimPrefix(string(b), "\ufeff"), "\n") {
			name := strings.ReplaceAll(strings.TrimSpace(line), "\\", "/")
			if name == "" {
				continue
			}
			if !strings.HasPrefix(name, "Interface/") {
				return nil, fmt.Errorf("unsafe client path %q", name)
			}
			if _, e = safeJoin(raw, name); e != nil {
				return nil, e
			}
			if strings.ContainsAny(name, ";*?\"<>|") {
				return nil, fmt.Errorf("unsafe client path: %s", name)
			}
			for _, r := range name {
				if r < 32 || r > 126 {
					return nil, fmt.Errorf("unsafe client path: %s", name)
				}
			}
			if seen[name] {
				return nil, fmt.Errorf("duplicate UI manifest entry: %s", name)
			}
			seen[name] = true
			k := strings.ToLower(name)
			if old := found[k]; old != "" && old != name {
				return nil, fmt.Errorf("case-colliding UI manifest paths")
			}
			found[k] = name
			if len(found)+len(manifests) > 20000 {
				return nil, fmt.Errorf("client manifest exceeds file limit")
			}
		}
	}
	if len(found) == 0 {
		return nil, fmt.Errorf("empty client manifest inventory")
	}
	for name := range manifests {
		found[strings.ToLower(name)] = name
	}
	out := make([]string, 0, len(found))
	for _, n := range found {
		out = append(out, n)
	}
	sort.Strings(out)
	return out, nil
}

// VerifySnapshot validates all raw bytes and the complete manifest-selected inventory.
func VerifySnapshot(snapshot string, expected *sourcelock.Snapshot) (sourcelock.Snapshot, error) {
	recorded, e := sourcelock.ReadSnapshot(filepath.Join(snapshot, "snapshot.json"))
	if e != nil {
		return sourcelock.Snapshot{}, e
	}
	if expected != nil && !recorded.Equal(*expected) {
		return sourcelock.Snapshot{}, fmt.Errorf("client snapshot differs from selected source lock")
	}
	if e = recorded.Identity().Validate(); e != nil {
		return sourcelock.Snapshot{}, e
	}
	recordedManifests := recorded.Manifests
	if len(recordedManifests) != len(manifests) {
		return sourcelock.Snapshot{}, fmt.Errorf("unsupported manifest identities")
	}
	for name, id := range manifests {
		if recordedManifests[name] != id {
			return sourcelock.Snapshot{}, fmt.Errorf("unsupported manifest identities")
		}
	}
	if recorded.Schema != 1 {
		return sourcelock.Snapshot{}, fmt.Errorf("unsupported client snapshot schema")
	}
	raw := filepath.Join(snapshot, "raw")
	names, e := manifestPaths(raw)
	if e != nil {
		return sourcelock.Snapshot{}, e
	}
	rows := recorded.Files
	if len(rows) != len(names) {
		return sourcelock.Snapshot{}, fmt.Errorf("client snapshot inventory does not match manifests")
	}
	inventory := map[string]bool{}
	total := 0
	for _, row := range rows {
		name := row.Path
		if inventory[name] {
			return sourcelock.Snapshot{}, fmt.Errorf("duplicate client snapshot entry: %s", name)
		}
		p, e := safeJoin(raw, name)
		if e != nil {
			return sourcelock.Snapshot{}, e
		}
		b, e := readFile(p, maxFile)
		if e != nil {
			return sourcelock.Snapshot{}, e
		}
		total += len(b)
		if total > maxTotal {
			return sourcelock.Snapshot{}, fmt.Errorf("client snapshot exceeds aggregate size")
		}
		if len(b) != row.Size || sha(b) != row.SHA256 || contentKey(b) != row.ContentKey {
			return sourcelock.Snapshot{}, fmt.Errorf("client content identity mismatch: %s", name)
		}
		inventory[name] = true
	}
	for _, name := range names {
		if !inventory[name] {
			return sourcelock.Snapshot{}, fmt.Errorf("missing manifest-selected input: %s", name)
		}
	}
	disk := 0
	e = filepath.WalkDir(raw, func(p string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("client snapshot contains symlink: %s", p)
		}
		if d.IsDir() {
			return nil
		}
		r, e := filepath.Rel(raw, p)
		if e != nil {
			return e
		}
		n := filepath.ToSlash(r)
		if n == "extractor-result.json" {
			return nil
		}
		if !inventory[n] {
			return fmt.Errorf("unlisted client snapshot file: %s", n)
		}
		disk++
		return nil
	})
	if e != nil {
		return sourcelock.Snapshot{}, e
	}
	if disk != len(names) {
		return sourcelock.Snapshot{}, fmt.Errorf("client snapshot inventory incomplete")
	}
	return recorded, nil
}
func projectVersion(root string, identity sourcelock.ClientIdentity) error {
	v, e := consumer.ReadVersion(root)
	if e != nil {
		return e
	}
	if v.Full != identity.Version || v.Build != identity.Build {
		return fmt.Errorf(".wow-version selects %s; source cohort is %s", v.Full, identity.Version)
	}
	return nil
}
func ResolveBuild(ctx context.Context, region string, build int) (sourcelock.ClientIdentity, error) {
	b, e := fetch(ctx, "https://us.version.battle.net/wow/versions", 64<<10)
	if e != nil {
		return sourcelock.ClientIdentity{}, e
	}
	var matches []string
	for line := range strings.SplitSeq(string(b), "\n") {
		if strings.HasPrefix(line, region+"|") {
			matches = append(matches, line)
		}
	}
	if len(matches) != 1 {
		return sourcelock.ClientIdentity{}, fmt.Errorf("missing or ambiguous Battle.net region: %s", region)
	}
	parts := strings.Split(strings.TrimSpace(matches[0]), "|")
	if len(parts) != 7 {
		return sourcelock.ClientIdentity{}, fmt.Errorf("invalid Battle.net version row")
	}
	var offered int
	fmt.Sscanf(parts[4], "%d", &offered)
	if build != 0 && offered != build {
		return sourcelock.ClientIdentity{}, fmt.Errorf("requested build %d is not currently offered (%d); supply exact-build lock", build, offered)
	}
	identity := sourcelock.ClientIdentity{Product: "wow", Region: region, Locale: "enUS", Build: offered, Version: parts[5], BuildConfig: parts[1], CDNConfig: parts[2], ProductConfig: parts[6]}
	return identity, identity.Validate()
}
func BuildExtractor(ctx context.Context, root string) (string, sourcelock.Extractor, error) {
	dir := filepath.Join(root, "tools/cdn")
	files, e := os.ReadDir(dir)
	if e != nil {
		return "", sourcelock.Extractor{}, e
	}
	var source []byte
	for _, f := range files {
		n := f.Name()
		if n != "go.mod" && (!strings.HasSuffix(n, ".go") || strings.HasSuffix(n, "_test.go")) {
			continue
		}
		b, e := readFile(filepath.Join(dir, n), maxFile)
		if e != nil {
			return "", sourcelock.Extractor{}, e
		}
		source = append(source, []byte(n)...)
		source = append(source, 0)
		source = append(source, b...)
	}
	if len(source) == 0 {
		return "", sourcelock.Extractor{}, fmt.Errorf("invalid extractor source inventory")
	}
	moduleCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	list := exec.CommandContext(moduleCtx, "go", "list", "-m", "all")
	list.Dir = dir
	list.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOWORK=off", "GOFLAGS=")
	modules, err := list.Output()
	if err != nil {
		return "", sourcelock.Extractor{}, fmt.Errorf("extractor module audit: %w", err)
	}
	if string(bytes.TrimSpace(modules)) != "wow-teal-types/tools/cdn" {
		return "", sourcelock.Extractor{}, fmt.Errorf("CDN extractor must have no external module dependencies")
	}
	binary := filepath.Join(root, "build/direct-source-tools/wow-cdn")
	e = command(ctx, dir, filepath.Join(root, "build/direct-source-tools/build.log"), 3*time.Minute, "go", "build", "-mod=readonly", "-trimpath", "-buildvcs=false", "-o", binary, ".")
	if e != nil {
		return "", sourcelock.Extractor{}, e
	}
	return binary, sourcelock.Extractor{Implementation: "wow-teal-types/tools/cdn", SourceSHA256: sha(source)}, nil
}
func Stage(ctx context.Context, root, output, cache string, identity sourcelock.ClientIdentity, expected *sourcelock.Snapshot) (sourcelock.Snapshot, error) {
	if exists(output) {
		return sourcelock.Snapshot{}, fmt.Errorf("client snapshot destination already exists: %s", output)
	}
	id := identity
	if e := id.Validate(); e != nil {
		return sourcelock.Snapshot{}, e
	}
	binary, tool, e := BuildExtractor(ctx, root)
	if e != nil {
		return sourcelock.Snapshot{}, e
	}
	if e = os.MkdirAll(filepath.Dir(output), 0755); e != nil {
		return sourcelock.Snapshot{}, e
	}
	stage, e := os.MkdirTemp(filepath.Dir(output), ".client-source-")
	if e != nil {
		return sourcelock.Snapshot{}, e
	}
	success := false
	defer func() {
		if !success {
			os.Remove(filepath.Join(stage, "snapshot.json"))
			if !exists(output + ".failed") {
				_ = os.Rename(stage, output+".failed")
			}
		}
	}()
	if e = writeObject(filepath.Join(stage, "config.json"), id); e != nil {
		return sourcelock.Snapshot{}, e
	}
	if cache == "" {
		cache = filepath.Join(root, "build/direct-source-tools/cache")
	}
	raw := filepath.Join(stage, "raw")
	if e = command(ctx, stage, filepath.Join(stage, "extraction.log"), 20*time.Minute, binary, filepath.Join(stage, "config.json"), raw, cache); e != nil {
		return sourcelock.Snapshot{}, e
	}
	receipt, e := readJSON[extractorReceipt](filepath.Join(raw, "extractor-result.json"))
	if e != nil {
		return sourcelock.Snapshot{}, e
	}
	actual := receipt.ClientIdentity
	if e = actual.Validate(); e != nil || actual != id {
		return sourcelock.Snapshot{}, fmt.Errorf("extractor returned different client identity")
	}
	record := sourcelock.NewSnapshot(id, tool, receipt.Files, manifests)
	if e = writeObject(filepath.Join(stage, "snapshot.json"), record); e != nil {
		return sourcelock.Snapshot{}, e
	}
	if _, e = VerifySnapshot(stage, expected); e != nil {
		return sourcelock.Snapshot{}, e
	}
	if e = os.Rename(stage, output); e != nil {
		return sourcelock.Snapshot{}, e
	}
	success = true
	return sourcelock.ReadSnapshot(filepath.Join(output, "snapshot.json"))
}

// Projection bridges packaged bytes to the locked source representation.
func Projection(root, snapshot, lockPath string) (wowapi.Project, error) {
	p := wowapi.NewProject(root)
	lock, e := sourcelock.ReadSnapshot(lockPath)
	if e != nil {
		return p, e
	}
	record, e := VerifySnapshot(snapshot, &lock)
	if e != nil {
		return p, e
	}
	if e = projectVersion(root, record.Identity()); e != nil {
		return p, e
	}
	sources, e := sourcelock.ReadSources(filepath.Join(root, "internal/wowapi/inputs/sources.json"))
	if e != nil {
		return p, e
	}
	lockBytes, e := readFile(lockPath, maxFile)
	if e != nil {
		return p, e
	}
	identity := sha(lockBytes)
	dest := filepath.Join(root, "build/wow-api-source", identity)
	files := map[string]bool{}
	for _, row := range record.Files {
		files[row.Path] = true
	}
	pending := map[string][]byte{}
	paths := map[string]string{}
	for _, group := range []string{"blizzard", "blizzard_support", "blizzard_version"} {
		source := sources[group]
		casc := source.Kind == "casc"
		if casc && (source.ClientLockSHA256 != identity || source.Normalization != "lf-final-newline") {
			return p, fmt.Errorf("native source selection differs from client lock: %s", group)
		}
		for _, row := range source.Files {
			name := row.Name
			if _, e = safeJoin(dest, group+"/"+name); e != nil {
				return p, e
			}
			relative := name
			if group == "blizzard" {
				relative = "Blizzard_APIDocumentationGenerated/" + name
			}
			origin := sourcelock.ClientPath("")
			var b []byte
			if group == "blizzard_version" {
				b = []byte(record.Version + "\n")
			} else {
				path, err := clientPath(relative, files)
				if err != nil {
					return p, err
				}
				if path == "" {
					return p, fmt.Errorf("locked source absent from client: %s", relative)
				}
				origin = sourcelock.ClientPath(path)
				b, e = readFile(filepath.Join(snapshot, "raw", filepath.FromSlash(path)), maxFile)
				if e != nil {
					return p, e
				}
			}
			if casc {
				if !row.ClientPath.Equal(origin) {
					return p, fmt.Errorf("native source path differs from client selection: %s", name)
				}
				b = normalized(b, 1)
			} else if blob(b) != row.SHA {
				b = lockedBytes(b, row)
			}
			if len(b) != row.Size || blob(b) != row.SHA {
				return p, fmt.Errorf("client source differs from locked contract: %s", relative)
			}
			pending[group+"/"+name] = b
		}
		paths[group] = filepath.Join(dest, group)
	}
	if exists(dest) {
		disk := map[string]bool{}
		e = filepath.WalkDir(dest, func(path string, d fs.DirEntry, e error) error {
			if e != nil {
				return e
			}
			if d.Type()&os.ModeSymlink != 0 {
				return fmt.Errorf("projection contains symlink")
			}
			if d.IsDir() {
				return nil
			}
			r, e := filepath.Rel(dest, path)
			if e != nil {
				return e
			}
			name := filepath.ToSlash(r)
			expected, ok := pending[name]
			if !ok {
				return fmt.Errorf("unlisted projection file: %s", name)
			}
			b, e := readFile(path, maxFile)
			if e != nil {
				return e
			}
			if !bytes.Equal(b, expected) {
				return fmt.Errorf("modified client source projection: %s", name)
			}
			disk[name] = true
			return nil
		})
		if e != nil {
			return p, e
		}
		if len(disk) != len(pending) {
			return p, fmt.Errorf("incomplete client source projection")
		}
	} else {
		if e = os.MkdirAll(filepath.Dir(dest), 0755); e != nil {
			return p, e
		}
		stage, e := os.MkdirTemp(filepath.Dir(dest), ".projection-")
		if e != nil {
			return p, e
		}
		defer os.RemoveAll(stage)
		for _, name := range keys(pending) {
			path, e := safeJoin(stage, name)
			if e != nil {
				return p, e
			}
			if e = atomicWrite(path, pending[name], 0644); e != nil {
				return p, e
			}
		}
		if e = os.Rename(stage, dest); e != nil {
			return p, e
		}
	}
	return p.WithInputs(paths), nil
}
func normalized(b []byte, count int) []byte {
	base := bytes.TrimRight(bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n")), "\n")
	return append(append([]byte{}, base...), bytes.Repeat([]byte("\n"), count)...)
}
func lockedBytes(b []byte, row sourcelock.SourceFile) []byte {
	for n := range 16 {
		v := normalized(b, n)
		if len(v) == row.Size && blob(v) == row.SHA {
			return v
		}
	}
	return nil
}

func fetch(ctx context.Context, address string, limit int64) ([]byte, error) {
	u, e := url.Parse(address)
	if e != nil || u.Scheme != "https" {
		return nil, fmt.Errorf("unsupported acquisition URL: %s", address)
	}
	var last error
	for attempt := range 3 {
		req, e := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
		if e != nil {
			return nil, e
		}
		req.Header.Set("User-Agent", "wow-teal-types")
		res, e := acquireHTTPClient().Do(req)
		if e == nil {
			if res.StatusCode != http.StatusOK {
				last = fmt.Errorf("HTTP %d: %s", res.StatusCode, address)
				res.Body.Close()
				if res.StatusCode != 429 && res.StatusCode < 500 {
					return nil, last
				}
			} else {
				b, err := readBounded(res.Body, limit)
				res.Body.Close()
				return b, err
			}
		} else {
			last = e
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		_ = attempt
	}
	return nil, fmt.Errorf("acquire %s: %w", address, last)
}
