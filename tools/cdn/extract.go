package main

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

// sourceInventory is the complete manifest-selected file set. Manifest bytes
// are already read and verified because their contents define the inventory.
type sourceInventory struct {
	manifests []manifestSource
	files     []preparedFile
	bytes     uint64
}

type manifestSource struct {
	file preparedFile
	data []byte
}

func extract(ctx context.Context, job extraction, output, cache string, progress io.Writer) error {
	c := newCDN(cache)
	defer c.close()
	build, cdn, err := loadConfigs(ctx, c, job)
	if err != nil {
		return err
	}
	fmt.Fprintf(progress, "Loading pinned %s for enUS...\n", job.buildName)
	if err := c.prepareIndexes(ctx, cdn); err != nil {
		return fmt.Errorf("archive indexes: %w", err)
	}
	encoding, err := loadEncoding(ctx, c, build)
	if err != nil {
		return err
	}
	rootKey, err := oneKey(build, "root")
	if err != nil {
		return fmt.Errorf("build config: %w", err)
	}
	root, err := readContent(ctx, c, encoding, rootKey, maxMetadata)
	if err != nil {
		return fmt.Errorf("root manifest: %w", err)
	}
	inv, err := inventory(ctx, c, encoding, root, job.manifests)
	if err != nil {
		return err
	}
	fmt.Fprintf(progress, "Extracting %d files (%d bytes), maximum %d workers...\n", len(inv.manifests)+len(inv.files), inv.bytes, maxWorkers)
	receipts := make([]fileReceipt, 0, len(inv.manifests)+len(inv.files))
	for _, manifest := range inv.manifests {
		receipt, err := writeSource(output, manifest.file, manifest.data)
		if err != nil {
			return err
		}
		receipts = append(receipts, receipt)
	}
	files, err := extractFiles(ctx, c, encoding, output, inv.files)
	if err != nil {
		return err
	}
	path, err := writeReceipt(output, job.identity, append(receipts, files...))
	if err != nil {
		return err
	}
	fmt.Fprintf(progress, "Verified extraction receipt: %s\n", path)
	return nil
}

// loadConfigs fetches the three pinned configurations and checks that the
// build configuration names the requested client.
func loadConfigs(ctx context.Context, c *cdnClient, job extraction) (build, cdn map[string][]string, err error) {
	if build, err = loadConfig(ctx, c, "build", job.buildConfig); err != nil {
		return nil, nil, err
	}
	if cdn, err = loadConfig(ctx, c, "CDN", job.cdnConfig); err != nil {
		return nil, nil, err
	}
	if _, err := c.config(ctx, "configs", job.productConfig); err != nil {
		return nil, nil, fmt.Errorf("product config %s: %w", keyText(job.productConfig), err)
	}
	if !slices.Equal(build["build-name"], []string{job.buildName}) || !slices.Equal(build["build-uid"], []string{job.identity.Product}) {
		return nil, nil, fmt.Errorf("verified build configuration does not identify %s", job.buildName)
	}
	return build, cdn, nil
}

func loadConfig(ctx context.Context, c *cdnClient, name string, k key) (map[string][]string, error) {
	data, err := c.config(ctx, "wow", k)
	if err != nil {
		return nil, fmt.Errorf("%s config %s: %w", name, keyText(k), err)
	}
	values, err := configValues(data)
	if err != nil {
		return nil, fmt.Errorf("%s config %s: %w", name, keyText(k), err)
	}
	return values, nil
}

// loadEncoding fetches the encoding manifest named by the build configuration
// and verifies its encoded size, decoded size and content key.
func loadEncoding(ctx context.Context, c *cdnClient, build map[string][]string) (*encodingTable, error) {
	keys := build["encoding"]
	if len(keys) != 2 {
		return nil, fmt.Errorf("build config requires encoding content and encoding keys")
	}
	contentKey, err := parseKey(keys[0])
	if err != nil {
		return nil, fmt.Errorf("build config encoding content key: %w", err)
	}
	encodingKey, err := parseKey(keys[1])
	if err != nil {
		return nil, fmt.Errorf("build config encoding key: %w", err)
	}
	sizes := build["encoding-size"]
	if len(sizes) != 2 {
		return nil, fmt.Errorf("build config requires decoded and encoded encoding sizes")
	}
	decodedSize, err := strconv.ParseUint(sizes[0], 10, 64)
	if err != nil || decodedSize == 0 || decodedSize > maxMetadata {
		return nil, fmt.Errorf("unsupported encoding decoded size %q", sizes[0])
	}
	encodedSize, err := strconv.ParseUint(sizes[1], 10, 64)
	if err != nil || encodedSize == 0 || encodedSize > maxMetadata {
		return nil, fmt.Errorf("unsupported encoding encoded size %q", sizes[1])
	}
	encoded, err := c.encoded(ctx, encodingKey, maxMetadata)
	if err != nil {
		return nil, fmt.Errorf("encoding manifest: %w", err)
	}
	if uint64(len(encoded)) != encodedSize {
		return nil, fmt.Errorf("encoding manifest encoded size mismatch")
	}
	decoded, err := decodeBLTE(encoded, maxMetadata, decodedSize)
	if err != nil {
		return nil, fmt.Errorf("encoding manifest: %w", err)
	}
	if key(md5.Sum(decoded)) != contentKey {
		return nil, fmt.Errorf("encoding manifest content key mismatch")
	}
	encoding, err := parseEncoding(decoded)
	if err != nil {
		return nil, fmt.Errorf("encoding manifest: %w", err)
	}
	return encoding, nil
}

// inventory resolves the selected manifests by file data ID, reads them, and
// resolves every path they list by name hash. It enforces path safety,
// case-collision, per-file, file-count and aggregate limits before any write.
func inventory(ctx context.Context, c *cdnClient, encoding *encodingTable, root []byte, manifests []manifestID) (sourceInventory, error) {
	wanted := make(map[uint32]bool, len(manifests))
	for _, manifest := range manifests {
		wanted[manifest.ID] = true
	}
	byID, _, err := scanRoot(root, wanted, nil)
	if err != nil {
		return sourceInventory{}, err
	}
	var inv sourceInventory
	paths := map[string]string{}
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
		inv.bytes += entry.Size
		if inv.bytes > maxSourceTotal {
			return preparedFile{}, fmt.Errorf("UI files exceed aggregate limit")
		}
		return preparedFile{name, k, entry.Size}, nil
	}
	selected := map[string]bool{}
	for _, manifest := range manifests {
		k, ok := byID[manifest.ID]
		if !ok {
			return sourceInventory{}, fmt.Errorf("%s: no eligible enUS root entry for FDID %d", manifest.Path, manifest.ID)
		}
		if err := addPath(manifest.Path); err != nil {
			return sourceInventory{}, err
		}
		file, err := prepare(manifest.Path, k)
		if err != nil {
			return sourceInventory{}, err
		}
		data, err := readContent(ctx, c, encoding, k, maxFile)
		if err != nil {
			return sourceInventory{}, fmt.Errorf("%s: %w", manifest.Path, err)
		}
		inv.manifests = append(inv.manifests, manifestSource{file, data})
		selected[manifest.Path] = true
		seen := map[string]bool{}
		for line := range strings.SplitSeq(strings.TrimPrefix(string(data), "\xef\xbb\xbf"), "\n") {
			name := strings.ReplaceAll(strings.TrimSpace(line), "\\", "/")
			if name == "" {
				continue
			}
			if seen[name] {
				return sourceInventory{}, fmt.Errorf("%s: duplicate manifest path %q", manifest.Path, name)
			}
			seen[name] = true
			if err := addPath(name); err != nil {
				return sourceInventory{}, err
			}
		}
	}
	if len(paths) <= len(manifests) {
		return sourceInventory{}, fmt.Errorf("empty complete UI manifest inventory")
	}
	names := make([]string, 0, len(paths))
	hashes := map[uint64]bool{}
	for _, name := range paths {
		if !selected[name] {
			names = append(names, name)
			hashes[pathHash(name)] = true
		}
	}
	slices.Sort(names)
	_, byHash, err := scanRoot(root, nil, hashes)
	if err != nil {
		return sourceInventory{}, err
	}
	inv.files = make([]preparedFile, 0, len(names))
	for _, name := range names {
		k, ok := byHash[pathHash(name)]
		if !ok {
			return sourceInventory{}, fmt.Errorf("%s: no unambiguous named enUS root entry", name)
		}
		file, err := prepare(name, k)
		if err != nil {
			return sourceInventory{}, fmt.Errorf("%s: %w", name, err)
		}
		inv.files = append(inv.files, file)
	}
	return inv, nil
}

func extractFiles(ctx context.Context, c *cdnClient, encoding *encodingTable, output string, files []preparedFile) ([]fileReceipt, error) {
	receipts := make([]fileReceipt, len(files))
	err := forEach(ctx, len(files), maxWorkers, func(ctx context.Context, i int) error {
		file := files[i]
		data, err := readContent(ctx, c, encoding, file.content, maxFile)
		if err == nil {
			receipts[i], err = writeSource(output, file, data)
		}
		if err != nil {
			return fmt.Errorf("%s: %w", file.path, err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return receipts, nil
}

func writeReceipt(output string, identity clientIdentity, receipts []fileReceipt) (string, error) {
	slices.SortFunc(receipts, func(a, b fileReceipt) int { return strings.Compare(a.Path, b.Path) })
	data, err := json.MarshalIndent(extractorReceipt{identity, receipts}, "", "  ")
	if err != nil {
		return "", err
	}
	data = append(data, '\n')
	path := filepath.Join(output, "extractor-result.json")
	if err := os.WriteFile(path+".tmp", data, 0644); err != nil {
		return "", err
	}
	if err := os.Rename(path+".tmp", path); err != nil {
		return "", err
	}
	return path, nil
}

// forEach runs fn for every index in [0, n) on at most workers goroutines. The
// first failure cancels the context passed to fn, stops further indexes from
// starting and is returned; otherwise the parent context's error is returned.
func forEach(ctx context.Context, n, workers int, fn func(context.Context, int) error) error {
	run, cancel := context.WithCancel(ctx)
	defer cancel()
	var next atomic.Int64
	var first error
	var once sync.Once
	var wg sync.WaitGroup
	for range min(workers, n) {
		wg.Go(func() {
			for run.Err() == nil {
				i := int(next.Add(1) - 1)
				if i >= n {
					return
				}
				if err := fn(run, i); err != nil {
					once.Do(func() { first = err; cancel() })
					return
				}
			}
		})
	}
	wg.Wait()
	if first != nil {
		return first
	}
	return ctx.Err()
}
