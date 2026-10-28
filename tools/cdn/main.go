// wow-cdn extracts a complete, pinned Retail UI source snapshot.
package main

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"syscall"
)

const (
	maxFile = 8 << 20
	// maxMetadata bounds one encoded or decoded metadata blob (encoding, root)
	// and is the BLTE decoder's ceiling.
	maxMetadata = 384 << 20
	// maxSourceTotal bounds aggregate extracted UI bytes. It is a separate limit
	// from maxMetadata that mirrors release's maxTotal, which VerifySnapshot
	// enforces over the same files; the two must stay equal.
	maxSourceTotal = 384 << 20
	maxFiles       = 20000
	maxManifests   = 16
	maxWorkers     = 4
	maxConfigBytes = 64 << 10
)

type clientIdentity struct {
	Product       string `json:"product"`
	Region        string `json:"region"`
	Locale        string `json:"locale"`
	Build         int    `json:"build"`
	Version       string `json:"version"`
	BuildConfig   string `json:"build_config"`
	CDNConfig     string `json:"cdn_config"`
	ProductConfig string `json:"product_config"`
}

// sourceConfig is the release-written config.json. Release owns the manifest
// selection (path to file data ID); the extractor validates and follows it.
type sourceConfig struct {
	clientIdentity
	Manifests map[string]uint32 `json:"manifests"`
}
type manifestID struct {
	ID   uint32
	Path string
}

// extraction is a validated sourceConfig: every key is parsed and the manifest
// selection is ordered by path, so extraction has no remaining parse failures.
type extraction struct {
	identity      clientIdentity
	buildConfig   key
	cdnConfig     key
	productConfig key
	buildName     string
	manifests     []manifestID
}

type fileReceipt struct {
	Path       string `json:"path"`
	Size       int    `json:"size"`
	SHA256     string `json:"sha256"`
	ContentKey string `json:"content_key"`
}
type extractorReceipt struct {
	clientIdentity
	Files []fileReceipt `json:"files"`
}
type preparedFile struct {
	path    string
	content key
	size    uint64
}

var clientVersion = regexp.MustCompile(`^([0-9]+\.[0-9]+\.[0-9]+)\.([0-9]+)$`)

func parseKey(s string) (key, error) {
	var k key
	b, err := hex.DecodeString(s)
	if err != nil || len(b) != len(k) || strings.ToLower(s) != s {
		return k, fmt.Errorf("invalid lowercase content key %q", s)
	}
	copy(k[:], b)
	if k == (key{}) {
		return k, fmt.Errorf("zero content key is unsupported")
	}
	return k, nil
}
func keyText(k key) string { return hex.EncodeToString(k[:]) }
func safePath(s string) error {
	if !strings.HasPrefix(s, "Interface/") {
		return fmt.Errorf("unsafe source path %q", s)
	}
	for _, c := range s {
		if c < 32 || c > 126 || strings.ContainsRune("\\:;*?\"<>|", c) {
			return fmt.Errorf("unsafe source path %q", s)
		}
	}
	for part := range strings.SplitSeq(s, "/") {
		if part == "" || part == "." || part == ".." {
			return fmt.Errorf("unsafe source path %q", s)
		}
	}
	return nil
}
func configValues(data []byte) (map[string][]string, error) {
	result := map[string][]string{}
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(strings.SplitN(line, "#", 2)[0])
		if line == "" {
			continue
		}
		name, value, ok := strings.Cut(line, "=")
		name = strings.TrimSpace(name)
		if !ok || name == "" {
			return nil, fmt.Errorf("malformed configuration line %q", line)
		}
		if _, exists := result[name]; exists {
			return nil, fmt.Errorf("duplicate configuration field %q", name)
		}
		result[name] = strings.Fields(value)
	}
	return result, nil
}
func oneKey(values map[string][]string, name string) (key, error) {
	entries := values[name]
	if len(entries) != 1 {
		return key{}, fmt.Errorf("configuration requires one %s value", name)
	}
	return parseKey(entries[0])
}
func validateConfig(c sourceConfig) (extraction, error) {
	regions := map[string]bool{"eu": true, "us": true, "kr": true, "cn": true, "tw": true}
	if c.Product != "wow" || c.Locale != "enUS" || !regions[c.Region] {
		return extraction{}, fmt.Errorf("only explicit wow/enUS and supported regions are accepted")
	}
	version := clientVersion.FindStringSubmatch(c.Version)
	if version == nil {
		return extraction{}, fmt.Errorf("invalid client version %q", c.Version)
	}
	build, err := strconv.Atoi(version[2])
	if err != nil || c.Build != build || build <= 0 {
		return extraction{}, fmt.Errorf("version %q and build %d disagree", c.Version, c.Build)
	}
	var keys [3]key
	for i, field := range [3]struct{ name, value string }{{"build_config", c.BuildConfig}, {"cdn_config", c.CDNConfig}, {"product_config", c.ProductConfig}} {
		if keys[i], err = parseKey(field.value); err != nil {
			return extraction{}, fmt.Errorf("%s: %w", field.name, err)
		}
	}
	if len(c.Manifests) == 0 || len(c.Manifests) > maxManifests {
		return extraction{}, fmt.Errorf("configuration requires 1..%d manifests", maxManifests)
	}
	manifests := make([]manifestID, 0, len(c.Manifests))
	for path, id := range c.Manifests {
		manifests = append(manifests, manifestID{id, path})
	}
	slices.SortFunc(manifests, func(a, b manifestID) int { return strings.Compare(a.Path, b.Path) })
	ids := make(map[uint32]string, len(manifests))
	for _, manifest := range manifests {
		if err := safePath(manifest.Path); err != nil {
			return extraction{}, fmt.Errorf("manifest: %w", err)
		}
		if manifest.ID == 0 {
			return extraction{}, fmt.Errorf("manifest %s has zero file data ID", manifest.Path)
		}
		if other, ok := ids[manifest.ID]; ok {
			return extraction{}, fmt.Errorf("manifests %s and %s share file data ID %d", other, manifest.Path, manifest.ID)
		}
		ids[manifest.ID] = manifest.Path
	}
	return extraction{
		identity:      c.clientIdentity,
		buildConfig:   keys[0],
		cdnConfig:     keys[1],
		productConfig: keys[2],
		buildName:     fmt.Sprintf("WOW-%dpatch%s_Retail", build, version[1]),
		manifests:     manifests,
	}, nil
}
func readContent(ctx context.Context, c *cdnClient, e *encodingTable, k key, limit uint64) ([]byte, error) {
	entry, err := e.lookup(k)
	if err != nil {
		return nil, fmt.Errorf("content %s: %w", keyText(k), err)
	}
	if entry.Size > limit {
		return nil, fmt.Errorf("content %s decoded size %d exceeds %d", keyText(k), entry.Size, limit)
	}
	var failures []error
	for _, ekey := range entry.Keys {
		encoded, err := c.encoded(ctx, ekey, int64(limit+limit/4+(1<<20)))
		if err == nil {
			var data []byte
			data, err = decodeBLTE(encoded, limit, entry.Size)
			if err == nil {
				sum := md5.Sum(data)
				if key(sum) != k || uint64(len(data)) != entry.Size {
					err = fmt.Errorf("decoded content key or encoding size mismatch")
				} else {
					return data, nil
				}
			}
		}
		failures = append(failures, fmt.Errorf("encoding %s: %w", keyText(ekey), err))
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
	}
	return nil, fmt.Errorf("content %s: %w", keyText(k), errors.Join(failures...))
}
func writeSource(output string, file preparedFile, data []byte) (fileReceipt, error) {
	if uint64(len(data)) != file.size || key(md5.Sum(data)) != file.content {
		return fileReceipt{}, fmt.Errorf("%s: size or content key mismatch", file.path)
	}
	path := filepath.Join(output, filepath.FromSlash(file.path))
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fileReceipt{}, err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return fileReceipt{}, err
	}
	_, writeErr := f.Write(data)
	closeErr := f.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		return fileReceipt{}, err
	}
	sum := sha256.Sum256(data)
	return fileReceipt{file.path, len(data), hex.EncodeToString(sum[:]), keyText(file.content)}, nil
}

// Prepare directories before comparing identities so missing path components
// and filesystem aliases are resolved by the filesystem itself.
func prepareDirectory(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(absolute, 0o755); err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(absolute)
}

func checkDirectorySeparation(output, cache string) error {
	for _, pair := range [][2]string{{output, cache}, {cache, output}} {
		root, err := os.Stat(pair[0])
		if err != nil {
			return err
		}
		for ancestor := pair[1]; ; {
			info, err := os.Stat(ancestor)
			if err != nil {
				return err
			}
			if os.SameFile(root, info) {
				return fmt.Errorf("cache and extraction output must not overlap")
			}
			parent := filepath.Dir(ancestor)
			if parent == ancestor {
				break
			}
			ancestor = parent
		}
	}
	return nil
}

func run(ctx context.Context, args []string, progress io.Writer) error {
	if len(args) != 3 {
		return fmt.Errorf("usage: wow-cdn CONFIG_JSON OUTPUT_DIR CACHE_DIR")
	}
	job, err := readConfig(args[0])
	if err != nil {
		return fmt.Errorf("configuration %s: %w", args[0], err)
	}
	output, err := prepareDirectory(args[1])
	if err != nil {
		return err
	}
	cache, err := prepareDirectory(args[2])
	if err != nil {
		return err
	}
	if err := checkDirectorySeparation(output, cache); err != nil {
		return err
	}
	entries, err := os.ReadDir(output)
	if err != nil {
		return err
	}
	if len(entries) != 0 {
		return fmt.Errorf("extraction output must be empty")
	}
	return extract(ctx, job, output, cache, progress)
}

func readConfig(path string) (extraction, error) {
	f, err := os.Open(path)
	if err != nil {
		return extraction{}, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return extraction{}, err
	}
	if info.Size() > maxConfigBytes {
		return extraction{}, fmt.Errorf("exceeds %d bytes", maxConfigBytes)
	}
	var config sourceConfig
	dec := json.NewDecoder(io.LimitReader(f, maxConfigBytes+1))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&config); err != nil {
		return extraction{}, err
	}
	if err := dec.Decode(new(json.RawMessage)); err != io.EOF {
		return extraction{}, fmt.Errorf("unexpected suffix")
	}
	return validateConfig(config)
}

// The release process owns the extraction deadline; this command only stops on
// interruption.
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "source extraction failed:", err)
		os.Exit(1)
	}
}
