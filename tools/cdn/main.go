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
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const maxFile = 8 << 20
const maxMetadata = 384 << 20
const maxFiles = 20000

type sourceConfig struct {
	Product       string `json:"product"`
	Region        string `json:"region"`
	Locale        string `json:"locale"`
	Build         int    `json:"build"`
	Version       string `json:"version"`
	BuildConfig   string `json:"build_config"`
	CDNConfig     string `json:"cdn_config"`
	ProductConfig string `json:"product_config"`
	Installation  string `json:"installation,omitzero"`
}
type manifestID struct {
	ID   uint32 `json:"file_data_id"`
	Path string `json:"path"`
}

var manifests = []manifestID{{6067012, "Interface/ui-code-list.txt"}, {6067013, "Interface/ui-toc-list.txt"}, {6076661, "Interface/ui-gen-addon-list.txt"}}

type fileReceipt struct {
	Path       string `json:"path"`
	Size       int    `json:"size"`
	SHA256     string `json:"sha256"`
	ContentKey string `json:"content_key"`
}
type preparedFile struct {
	path    string
	content key
	size    uint64
}

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
func validateConfig(c sourceConfig) error {
	regions := map[string]bool{"eu": true, "us": true, "kr": true, "cn": true, "tw": true}
	if c.Product != "wow" || c.Locale != "enUS" || !regions[c.Region] {
		return fmt.Errorf("only explicit wow/enUS and supported regions are accepted")
	}
	version := regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+\.([0-9]+)$`).FindStringSubmatch(c.Version)
	if len(version) != 2 {
		return fmt.Errorf("invalid client version %q", c.Version)
	}
	build, err := strconv.Atoi(version[1])
	if err != nil || c.Build != build || build <= 0 {
		return fmt.Errorf("version and build disagree")
	}
	for _, s := range []string{c.BuildConfig, c.CDNConfig, c.ProductConfig} {
		if _, err := parseKey(s); err != nil {
			return err
		}
	}
	return nil
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
func extract(ctx context.Context, config sourceConfig, output, cache string) error {
	c, err := newCDN(cache, config.Installation)
	if err != nil {
		return err
	}
	defer c.close()
	bkey, _ := parseKey(config.BuildConfig)
	ckey, _ := parseKey(config.CDNConfig)
	pkey, _ := parseKey(config.ProductConfig)
	data, err := c.config(ctx, "wow", bkey)
	if err != nil {
		return err
	}
	build, err := configValues(data)
	if err != nil {
		return err
	}
	data, err = c.config(ctx, "wow", ckey)
	if err != nil {
		return err
	}
	cdn, err := configValues(data)
	if err != nil {
		return err
	}
	if _, err = c.config(ctx, "configs", pkey); err != nil {
		return err
	}
	expected := fmt.Sprintf("WOW-%dpatch%s_Retail", config.Build, config.Version[:strings.LastIndexByte(config.Version, '.')])
	if len(build["build-name"]) != 1 || build["build-name"][0] != expected || len(build["build-uid"]) != 1 || build["build-uid"][0] != config.Product {
		return fmt.Errorf("verified build configuration does not identify %s", expected)
	}
	fmt.Printf("Loading pinned %s for enUS...\n", expected)
	verification, differences, err := c.prepareIndexes(ctx, cdn)
	if err != nil {
		return err
	}
	keys := build["encoding"]
	if len(keys) != 2 {
		return fmt.Errorf("build requires encoding content and encoding keys")
	}
	encodingContent, err := parseKey(keys[0])
	if err != nil {
		return err
	}
	encodingKey, err := parseKey(keys[1])
	if err != nil {
		return err
	}
	sizes := build["encoding-size"]
	if len(sizes) != 2 {
		return fmt.Errorf("build requires decoded and encoded encoding sizes")
	}
	decodedSize, err := strconv.ParseUint(sizes[0], 10, 64)
	if err != nil || decodedSize == 0 || decodedSize > maxMetadata {
		return fmt.Errorf("unsupported encoding decoded size")
	}
	encodedSize, err := strconv.ParseUint(sizes[1], 10, 64)
	if err != nil || encodedSize == 0 || encodedSize > maxMetadata {
		return fmt.Errorf("unsupported encoding encoded size")
	}
	encoded, err := c.encoded(ctx, encodingKey, maxMetadata)
	if err != nil {
		return err
	}
	if uint64(len(encoded)) != encodedSize {
		return fmt.Errorf("encoding manifest encoded size mismatch")
	}
	decoded, err := decodeBLTE(encoded, maxMetadata, decodedSize)
	if err != nil {
		return fmt.Errorf("encoding: %w", err)
	}
	if key(md5.Sum(decoded)) != encodingContent {
		return fmt.Errorf("encoding manifest content key mismatch")
	}
	encoding, err := parseEncoding(decoded)
	if err != nil {
		return err
	}
	rootKey, err := oneKey(build, "root")
	if err != nil {
		return err
	}
	root, err := readContent(ctx, c, encoding, rootKey, maxMetadata)
	if err != nil {
		return fmt.Errorf("root manifest: %w", err)
	}
	wanted := map[uint32]bool{}
	for _, manifest := range manifests {
		wanted[manifest.ID] = true
	}
	byID, _, err := scanRoot(root, wanted, nil)
	if err != nil {
		return err
	}
	paths := map[string]string{}
	receipts := []fileReceipt{}
	prepared := map[string]preparedFile{}
	var total uint64
	addPath := func(name string) error {
		if err := safePath(name); err != nil {
			return err
		}
		folded := strings.ToLower(name)
		if old, ok := paths[folded]; ok && old != name {
			return fmt.Errorf("case-colliding manifest paths %q and %q", old, name)
		}
		paths[folded] = name
		if len(paths) > maxFiles {
			return fmt.Errorf("UI manifest file count exceeds %d", maxFiles)
		}
		return nil
	}
	prepare := func(name string, k key) (preparedFile, error) {
		entry, err := encoding.lookup(k)
		if err != nil {
			return preparedFile{}, err
		}
		if entry.Size > maxFile {
			return preparedFile{}, fmt.Errorf("%s exceeds per-file limit", name)
		}
		total += entry.Size
		if total > maxMetadata {
			return preparedFile{}, fmt.Errorf("UI files exceed aggregate limit")
		}
		return preparedFile{name, k, entry.Size}, nil
	}
	for _, manifest := range manifests {
		k, ok := byID[manifest.ID]
		if !ok {
			return fmt.Errorf("%s: no eligible enUS root entry for FDID %d", manifest.Path, manifest.ID)
		}
		if err := addPath(manifest.Path); err != nil {
			return err
		}
		file, err := prepare(manifest.Path, k)
		if err != nil {
			return err
		}
		data, err := readContent(ctx, c, encoding, k, maxFile)
		if err != nil {
			return err
		}
		receipt, err := writeSource(output, file, data)
		if err != nil {
			return err
		}
		receipts = append(receipts, receipt)
		prepared[file.path] = file
		seen := map[string]bool{}
		for line := range strings.SplitSeq(strings.TrimPrefix(string(data), "\xef\xbb\xbf"), "\n") {
			name := strings.ReplaceAll(strings.TrimSpace(line), "\\", "/")
			if name == "" {
				continue
			}
			if seen[name] {
				return fmt.Errorf("%s: duplicate manifest path %q", manifest.Path, name)
			}
			seen[name] = true
			if err := addPath(name); err != nil {
				return err
			}
		}
	}
	if len(paths) <= len(manifests) {
		return fmt.Errorf("empty complete UI manifest inventory")
	}
	names := make([]string, 0, len(paths))
	hashes := map[uint64]bool{}
	for _, name := range paths {
		names = append(names, name)
		if _, ok := prepared[name]; !ok {
			hashes[pathHash(name)] = true
		}
	}
	sort.Strings(names)
	_, byHash, err := scanRoot(root, nil, hashes)
	if err != nil {
		return err
	}
	files := []preparedFile{}
	for _, name := range names {
		if _, ok := prepared[name]; ok {
			continue
		}
		k, ok := byHash[pathHash(name)]
		if !ok {
			return fmt.Errorf("%s: no unambiguous named enUS root entry", name)
		}
		file, err := prepare(name, k)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		files = append(files, file)
	}
	fmt.Printf("Extracting %d files (%d bytes), maximum 4 workers...\n", len(paths), total)
	workCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	jobs := make(chan int)
	results := make([]fileReceipt, len(files))
	var wg sync.WaitGroup
	var first error
	var once sync.Once
	for range 4 {
		wg.Go(func() {
			for i := range jobs {
				if workCtx.Err() != nil {
					continue
				}
				file := files[i]
				data, err := readContent(workCtx, c, encoding, file.content, maxFile)
				if err == nil {
					results[i], err = writeSource(output, file, data)
				}
				if err != nil {
					once.Do(func() { first = fmt.Errorf("%s: %w", file.path, err); cancel() })
				}
			}
		})
	}
sendFiles:
	for i := range files {
		select {
		case jobs <- i:
		case <-workCtx.Done():
			break sendFiles
		}
	}
	close(jobs)
	wg.Wait()
	if first != nil {
		return first
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	receipts = append(receipts, results...)
	sort.Slice(receipts, func(i, j int) bool { return receipts[i].Path < receipts[j].Path })
	receipt := struct {
		sourceConfig
		Manifests    []manifestID         `json:"manifests"`
		Files        []fileReceipt        `json:"files"`
		Verification indexVerification    `json:"archive_index_verification"`
		Differences  []sizeHintDifference `json:"archive_index_size_hints"`
	}{config, manifests, receipts, verification, differences}
	data, err = json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	path := filepath.Join(output, "extractor-result.json")
	if err := os.WriteFile(path+".tmp", data, 0644); err != nil {
		return err
	}
	if err := os.Rename(path+".tmp", path); err != nil {
		return err
	}
	fmt.Printf("Verified extraction receipt: %s\n", path)
	return nil
}
func run(ctx context.Context, args []string) error {
	if len(args) != 3 {
		return fmt.Errorf("usage: wow-cdn CONFIG_JSON OUTPUT_DIR CACHE_DIR")
	}
	f, err := os.Open(args[0])
	if err != nil {
		return err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return err
	}
	if info.Size() > 64<<10 {
		return fmt.Errorf("configuration exceeds 64 KiB")
	}
	var config sourceConfig
	dec := json.NewDecoder(io.LimitReader(f, (64<<10)+1))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&config); err != nil {
		return err
	}
	if err := dec.Decode(new(json.RawMessage)); err != io.EOF {
		return fmt.Errorf("unexpected configuration suffix")
	}
	if err := validateConfig(config); err != nil {
		return err
	}
	output, err := filepath.Abs(args[1])
	if err != nil {
		return err
	}
	cache, err := filepath.Abs(args[2])
	if err != nil {
		return err
	}
	for _, pair := range [][2]string{{output, cache}, {cache, output}} {
		relative, err := filepath.Rel(pair[0], pair[1])
		if err != nil {
			return err
		}
		if relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))) {
			return fmt.Errorf("cache and extraction output must not overlap")
		}
	}
	entries, err := os.ReadDir(output)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if len(entries) != 0 {
		return fmt.Errorf("extraction output must be empty")
	}
	if err := os.MkdirAll(output, 0755); err != nil {
		return err
	}
	return extract(ctx, config, output, cache)
}
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, 20*time.Minute)
	defer cancel()
	if err := run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "source extraction failed:", err)
		os.Exit(1)
	}
}
