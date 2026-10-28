package release

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"wow-teal-types/internal/consumer"
	"wow-teal-types/internal/sourcelock"
	"wow-teal-types/internal/wowapi"
)

// manifests is the authoritative UI manifest selection (path to file data ID).
// Release passes it to the extractor through config.json and records it in the
// client snapshot.
var manifests = map[string]int{"Interface/ui-code-list.txt": 6067012, "Interface/ui-toc-list.txt": 6067013, "Interface/ui-gen-addon-list.txt": 6076661}

// extractorConfig is the config.json contract read by tools/cdn.
type extractorConfig struct {
	sourcelock.ClientIdentity
	Manifests map[string]int `json:"manifests"`
	MaxFile   int            `json:"max_file"`
	MaxTotal  int            `json:"max_total"`
}

// newExtractorConfig passes the limits VerifySnapshot enforces to the
// extractor, so extraction fails at the same bounds before downloading.
func newExtractorConfig(id sourcelock.ClientIdentity) extractorConfig {
	return extractorConfig{id, manifests, maxFile, maxTotal}
}

func clientPath(name string, files map[string]sourcelock.ClientFile) (string, error) {
	exact := "Interface/AddOns/" + name
	if _, ok := files[exact]; ok {
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
		b, e := sourcelock.ReadRegular(path, maxFile)
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
	v, e := verifySnapshot(snapshot, expected)
	return v.record, e
}

// verifiedSnapshot is a client snapshot whose record and complete raw inventory
// passed verification once. Its only raw read, packagedFile, re-checks each
// file against the record, so a later on-disk change fails at use instead of
// requiring the whole snapshot to be hashed again.
type verifiedSnapshot struct {
	dir    string
	record sourcelock.Snapshot
	files  map[string]sourcelock.ClientFile
}

func verifySnapshot(snapshot string, expected *sourcelock.Snapshot) (verifiedSnapshot, error) {
	recorded, e := sourcelock.ReadSnapshot(filepath.Join(snapshot, "snapshot.json"))
	if e != nil {
		return verifiedSnapshot{}, e
	}
	if expected != nil && !recorded.Equal(*expected) {
		return verifiedSnapshot{}, fmt.Errorf("client snapshot differs from selected source lock")
	}
	if e = recorded.Identity().Validate(); e != nil {
		return verifiedSnapshot{}, e
	}
	recordedManifests := recorded.Manifests
	if len(recordedManifests) != len(manifests) {
		return verifiedSnapshot{}, fmt.Errorf("unsupported manifest identities")
	}
	for name, id := range manifests {
		if recordedManifests[name] != id {
			return verifiedSnapshot{}, fmt.Errorf("unsupported manifest identities")
		}
	}
	if recorded.Schema != 1 {
		return verifiedSnapshot{}, fmt.Errorf("unsupported client snapshot schema")
	}
	raw := filepath.Join(snapshot, "raw")
	names, e := manifestPaths(raw)
	if e != nil {
		return verifiedSnapshot{}, e
	}
	rows := recorded.Files
	if len(rows) != len(names) {
		return verifiedSnapshot{}, fmt.Errorf("client snapshot inventory does not match manifests")
	}
	inventory := map[string]sourcelock.ClientFile{}
	total := 0
	for _, row := range rows {
		name := row.Path
		if _, ok := inventory[name]; ok {
			return verifiedSnapshot{}, fmt.Errorf("duplicate client snapshot entry: %s", name)
		}
		p, e := safeJoin(raw, name)
		if e != nil {
			return verifiedSnapshot{}, e
		}
		b, e := sourcelock.ReadRegular(p, maxFile)
		if e != nil {
			return verifiedSnapshot{}, e
		}
		total += len(b)
		if total > maxTotal {
			return verifiedSnapshot{}, fmt.Errorf("client snapshot exceeds aggregate size")
		}
		if !clientContent(row, b) {
			return verifiedSnapshot{}, fmt.Errorf("client content identity mismatch: %s", name)
		}
		inventory[name] = row
	}
	for _, name := range names {
		if _, ok := inventory[name]; !ok {
			return verifiedSnapshot{}, fmt.Errorf("missing manifest-selected input: %s", name)
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
		if _, ok := inventory[n]; !ok {
			return fmt.Errorf("unlisted client snapshot file: %s", n)
		}
		disk++
		return nil
	})
	if e != nil {
		return verifiedSnapshot{}, e
	}
	if disk != len(names) {
		return verifiedSnapshot{}, fmt.Errorf("client snapshot inventory incomplete")
	}
	return verifiedSnapshot{snapshot, recorded, inventory}, nil
}

func clientContent(row sourcelock.ClientFile, b []byte) bool {
	return len(b) == row.Size && sha(b) == row.SHA256 && contentKey(b) == row.ContentKey
}

// packagedFile reads the addon file that name selects ("<addon>/<file>") and
// returns its packaged path. An empty path means the client lacks the file.
func (v verifiedSnapshot) packagedFile(name string) (string, []byte, error) {
	path, e := clientPath(name, v.files)
	if e != nil || path == "" {
		return "", nil, e
	}
	b, e := sourcelock.ReadRegular(filepath.Join(v.dir, "raw", filepath.FromSlash(path)), maxFile)
	if e != nil {
		return "", nil, e
	}
	if !clientContent(v.files[path], b) {
		return "", nil, fmt.Errorf("client snapshot changed after verification: %s", path)
	}
	return path, b, nil
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
func resolveBuild(ctx context.Context, f fetcher, region string, build int) (sourcelock.ClientIdentity, error) {
	b, e := f.get(ctx, "https://us.version.battle.net/wow/versions", 64<<10)
	if e != nil {
		return sourcelock.ClientIdentity{}, e
	}
	return parseVersionRow(b, region, build)
}

// parseVersionRow selects the region's row from a Battle.net version listing.
// wantBuild 0 accepts whichever build is offered.
func parseVersionRow(body []byte, region string, wantBuild int) (sourcelock.ClientIdentity, error) {
	var matches []string
	for line := range strings.SplitSeq(string(body), "\n") {
		if strings.HasPrefix(line, region+"|") {
			matches = append(matches, line)
		}
	}
	if len(matches) != 1 {
		return sourcelock.ClientIdentity{}, fmt.Errorf("missing or ambiguous Battle.net region: %s", region)
	}
	parts := strings.Split(strings.TrimSpace(matches[0]), "|")
	if len(parts) != 7 {
		return sourcelock.ClientIdentity{}, fmt.Errorf("invalid Battle.net version row: %d fields", len(parts))
	}
	offered, e := strconv.Atoi(parts[4])
	if e != nil {
		return sourcelock.ClientIdentity{}, fmt.Errorf("invalid Battle.net build field %q: %w", parts[4], e)
	}
	if wantBuild != 0 && offered != wantBuild {
		return sourcelock.ClientIdentity{}, fmt.Errorf("requested build %d is not currently offered (%d); supply exact-build lock", wantBuild, offered)
	}
	identity := sourcelock.ClientIdentity{Product: "wow", Region: region, Locale: "enUS", Build: offered, Version: parts[5], BuildConfig: parts[1], CDNConfig: parts[2], ProductConfig: parts[6]}
	if e = identity.Validate(); e != nil {
		return sourcelock.ClientIdentity{}, fmt.Errorf("Battle.net %s row: %w", region, e)
	}
	return identity, nil
}

// extractorSourceSHA identifies the extractor build inputs that key cached snapshots.
func extractorSourceSHA(root string) (string, error) {
	dir := filepath.Join(root, "tools/cdn")
	files, e := os.ReadDir(dir)
	if e != nil {
		return "", e
	}
	var source []byte
	for _, f := range files {
		n := f.Name()
		if n != "go.mod" && (!strings.HasSuffix(n, ".go") || strings.HasSuffix(n, "_test.go")) {
			continue
		}
		b, e := sourcelock.ReadRegular(filepath.Join(dir, n), maxFile)
		if e != nil {
			return "", e
		}
		source = append(source, []byte(n)...)
		source = append(source, 0)
		source = append(source, b...)
	}
	if len(source) == 0 {
		return "", fmt.Errorf("invalid extractor source inventory")
	}
	return sha(source), nil
}

func BuildExtractor(ctx context.Context, root string) (string, error) {
	dir := filepath.Join(root, "tools/cdn")
	moduleCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	list := exec.CommandContext(moduleCtx, "go", "list", "-m", "all")
	list.Dir = dir
	list.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOWORK=off", "GOFLAGS=")
	modules, err := list.Output()
	if err != nil {
		return "", fmt.Errorf("extractor module audit: %w", err)
	}
	if string(bytes.TrimSpace(modules)) != "wow-teal-types/tools/cdn" {
		return "", fmt.Errorf("CDN extractor must have no external module dependencies")
	}
	binary := filepath.Join(root, "build/direct-source-tools/wow-cdn")
	e := command(ctx, dir, filepath.Join(root, "build/direct-source-tools/build.log"), 3*time.Minute, "go", "build", "-mod=readonly", "-trimpath", "-buildvcs=false", "-o", binary, ".")
	if e != nil {
		return "", e
	}
	return binary, nil
}
func extractSnapshot(ctx context.Context, root, output, cache string, identity sourcelock.ClientIdentity, expected *sourcelock.Snapshot) (verifiedSnapshot, error) {
	if exists(output) {
		return verifiedSnapshot{}, fmt.Errorf("client snapshot destination already exists: %s", output)
	}
	id := identity
	if e := id.Validate(); e != nil {
		return verifiedSnapshot{}, e
	}
	binary, e := BuildExtractor(ctx, root)
	if e != nil {
		return verifiedSnapshot{}, e
	}
	if e = os.MkdirAll(filepath.Dir(output), 0755); e != nil {
		return verifiedSnapshot{}, e
	}
	stage, e := os.MkdirTemp(filepath.Dir(output), ".client-source-")
	if e != nil {
		return verifiedSnapshot{}, e
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
	if e = writeObject(filepath.Join(stage, "config.json"), newExtractorConfig(id)); e != nil {
		return verifiedSnapshot{}, e
	}
	if cache == "" {
		cache = filepath.Join(root, "build/direct-source-tools/cache")
	}
	raw := filepath.Join(stage, "raw")
	if e = command(ctx, stage, filepath.Join(stage, "extraction.log"), 20*time.Minute, binary, filepath.Join(stage, "config.json"), raw, cache); e != nil {
		return verifiedSnapshot{}, e
	}
	receipt, e := readJSON[extractorReceipt](filepath.Join(raw, "extractor-result.json"))
	if e != nil {
		return verifiedSnapshot{}, e
	}
	actual := receipt.ClientIdentity
	if e = actual.Validate(); e != nil || actual != id {
		return verifiedSnapshot{}, fmt.Errorf("extractor returned different client identity")
	}
	record := sourcelock.NewSnapshot(id, receipt.Files, manifests)
	if e = writeObject(filepath.Join(stage, "snapshot.json"), record); e != nil {
		return verifiedSnapshot{}, e
	}
	verified, e := verifySnapshot(stage, expected)
	if e != nil {
		return verifiedSnapshot{}, e
	}
	if e = os.Rename(stage, output); e != nil {
		return verifiedSnapshot{}, e
	}
	success = true
	verified.dir = output
	return verified, nil
}

// projectSources bridges packaged bytes to the locked source representation.
func projectSources(root string, snapshot verifiedSnapshot) (wowapi.Project, error) {
	p := wowapi.NewProject(root)
	lockPath := p.ClientSourcesPath()
	lock, e := sourcelock.ReadSnapshot(lockPath)
	if e != nil {
		return p, e
	}
	record := snapshot.record
	if !record.Equal(lock) {
		return p, fmt.Errorf("client snapshot differs from selected source lock")
	}
	if e = projectVersion(root, record.Identity()); e != nil {
		return p, e
	}
	sources, e := sourcelock.ReadSources(p.SourcesPath())
	if e != nil {
		return p, e
	}
	lockBytes, e := sourcelock.ReadRegular(lockPath, maxFile)
	if e != nil {
		return p, e
	}
	identity := sha(lockBytes)
	selection, e := jsonBytes(sourcelock.Sources{
		"blizzard": sources["blizzard"], "blizzard_version": sources["blizzard_version"],
	})
	if e != nil {
		return p, e
	}
	// The selected representation can change independently of the raw snapshot.
	dest := filepath.Join(root, "build/wow-api-source", identity+"-"+sha(selection))
	pending := map[string][]byte{}
	paths := map[string]string{}
	for _, group := range []string{"blizzard", "blizzard_version"} {
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
				path, raw, err := snapshot.packagedFile(relative)
				if err != nil {
					return p, err
				}
				if path == "" {
					return p, fmt.Errorf("locked source absent from client: %s", relative)
				}
				origin = sourcelock.ClientPath(path)
				b = raw
			}
			if casc {
				if !row.ClientPath.Equal(origin) {
					return p, fmt.Errorf("native source path differs from client selection: %s", name)
				}
				b = normalized(b, 1)
			} else if locked, ok := lockedBytes(b, row); ok {
				b = locked
			}
			if !row.Matches(b) {
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
			b, e := sourcelock.ReadRegular(path, maxFile)
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
// lockedBytes returns b in the locked representation: unchanged, or with
// normalized line endings and a final-newline count the lock records.
func lockedBytes(b []byte, row sourcelock.SourceFile) ([]byte, bool) {
	if row.Matches(b) {
		return b, true
	}
	for n := range 16 {
		v := normalized(b, n)
		if row.Matches(v) {
			return v, true
		}
	}
	return nil, false
}

// fetcher owns HTTPS acquisition. One client serves every attempt; wait paces
// retries and is never called before the first attempt or after the last.
type fetcher struct {
	client *http.Client
	wait   func(ctx context.Context, attempt int) error
}

const fetchAttempts = 3

func newFetcher() fetcher {
	client := &http.Client{Timeout: 45 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 3 || req.URL.Scheme != "https" {
			return fmt.Errorf("unsupported acquisition redirect")
		}
		return nil
	}}
	return fetcher{client: client, wait: backoff}
}

// backoff waits 1s before the second attempt and 2s before the third.
func backoff(ctx context.Context, attempt int) error {
	t := time.NewTimer(time.Second << (attempt - 1))
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// get retries transport failures, 429 and 5xx; other statuses and oversized
// bodies fail at once.
func (f fetcher) get(ctx context.Context, address string, limit int64) ([]byte, error) {
	u, e := url.Parse(address)
	if e != nil || u.Scheme != "https" {
		return nil, fmt.Errorf("unsupported acquisition URL: %s", address)
	}
	var last error
	for attempt := range fetchAttempts {
		if attempt > 0 {
			if e := f.wait(ctx, attempt); e != nil {
				return nil, fmt.Errorf("acquire %s: retry after %v: %w", address, last, e)
			}
		}
		req, e := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
		if e != nil {
			return nil, e
		}
		req.Header.Set("User-Agent", "wow-teal-types")
		res, e := f.client.Do(req)
		if e != nil {
			if ctx.Err() != nil {
				return nil, fmt.Errorf("acquire %s: %w", address, ctx.Err())
			}
			last = e
			continue
		}
		if res.StatusCode == http.StatusOK {
			b, e := readBounded(res.Body, limit)
			res.Body.Close()
			if e != nil {
				return nil, fmt.Errorf("acquire %s: %w", address, e)
			}
			return b, nil
		}
		res.Body.Close()
		last = fmt.Errorf("HTTP %d", res.StatusCode)
		if res.StatusCode != http.StatusTooManyRequests && res.StatusCode < 500 {
			return nil, fmt.Errorf("acquire %s: %w", address, last)
		}
	}
	return nil, fmt.Errorf("acquire %s after %d attempts: %w", address, fetchAttempts, last)
}

func readBounded(r io.Reader, limit int64) ([]byte, error) {
	b, e := io.ReadAll(io.LimitReader(r, limit+1))
	if e != nil {
		return nil, e
	}
	if int64(len(b)) > limit {
		return nil, fmt.Errorf("acquisition response exceeds %d bytes", limit)
	}
	return b, nil
}
