package main

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type cdnClient struct {
	// cache must be a resolved path (prepareDirectory); cdnCheckParents trusts
	// it as the root of the symlink-free tree.
	cache   string
	client  *http.Client
	hosts   []string
	indices []archiveIndex
	// indexBudget bounds the aggregate bytes of retained archive indexes.
	indexBudget int64
	cacheLocks  [64]sync.Mutex
}

func newCDN(cache string) *cdnClient {
	return &cdnClient{cache: cache, indexBudget: archiveIndexTotal,
		client: &http.Client{Transport: &http.Transport{MaxConnsPerHost: 4, MaxIdleConns: 8,
			MaxIdleConnsPerHost: 4, IdleConnTimeout: 60 * time.Second, DisableCompression: true},
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return errors.New("CDN redirects are not permitted") }},
		hosts: []string{"https://us.cdn.blizzard.com", "https://level3.ssl.blizzard.com"}}
}

func (c *cdnClient) close() { c.client.CloseIdleConnections() }

func cdnPath(namespace, kind string, k key) string {
	text := hex.EncodeToString(k[:])
	return "tpr/" + namespace + "/" + kind + "/" + text[:2] + "/" + text[2:4] + "/" + text
}

func (c *cdnClient) config(ctx context.Context, namespace string, k key) ([]byte, error) {
	if namespace != "wow" && namespace != "configs" {
		return nil, fmt.Errorf("unsupported config namespace %q", namespace)
	}
	kind := "config"
	if namespace == "configs" {
		kind = "data"
	}
	return c.cached(ctx, filepath.Join("tpr", namespace, kind, hex.EncodeToString(k[:])), 8<<20,
		func(data []byte) error {
			if md5.Sum(data) != [16]byte(k) {
				return errors.New("config MD5 mismatch")
			}
			return nil
		}, func() ([]byte, error) { return c.download(ctx, cdnPath(namespace, kind, k), 8<<20, nil) })
}

func (c *cdnClient) encoded(ctx context.Context, k key, limit int64) ([]byte, error) {
	if limit < 9 || limit > 512<<20 {
		return nil, errors.New("encoded file limit outside supported bounds")
	}
	return c.cached(ctx, filepath.Join("tpr", "wow", "data", hex.EncodeToString(k[:])), limit,
		func(data []byte) error { return verifyEncodingKey(data, k) }, func() ([]byte, error) {
			for _, index := range c.indices {
				if location, ok := index.locate(k); ok {
					if int64(location.size) > limit {
						return nil, fmt.Errorf("encoded file %x exceeds limit %d", k, limit)
					}
					return c.download(ctx, cdnPath("wow", "data", location.archive), limit, &location)
				}
			}
			return c.download(ctx, cdnPath("wow", "data", k), limit, nil)
		})
}

func verifyEncodingKey(data []byte, expected key) error {
	if len(data) < 9 || string(data[:4]) != "BLTE" {
		return errors.New("invalid BLTE envelope")
	}
	header := uint64(binary.BigEndian.Uint32(data[4:8]))
	identity := data
	if header != 0 {
		if header < 12 || header > uint64(len(data)) {
			return errors.New("invalid BLTE header size")
		}
		identity = data[:int(header)]
	}
	if md5.Sum(identity) != [16]byte(expected) {
		return fmt.Errorf("BLTE encoding key mismatch for %x", expected)
	}
	return verifyBLTEEnvelope(data)
}

func (c *cdnClient) prepareIndexes(ctx context.Context, values map[string][]string) error {
	archives, sizes := values["archives"], values["archives-index-size"]
	if len(archives) == 0 || len(archives) > 10_000 || len(sizes) != len(archives) {
		return errors.New("CDN config must contain 1..10000 archives with matching size hints")
	}
	keys := make([]key, len(archives))
	seen := make(map[key]bool, len(archives))
	for i, archive := range archives {
		k, err := parseKey(archive)
		if err != nil {
			return err
		}
		if seen[k] {
			return fmt.Errorf("duplicate archive %s", archive)
		}
		seen[k], keys[i] = true, k
		size, err := strconv.ParseInt(sizes[i], 10, 64)
		if err != nil || size <= 0 {
			return fmt.Errorf("archive %s has invalid size hint %q", archive, sizes[i])
		}
	}
	var consumed atomic.Int64
	indices := make([]archiveIndex, len(keys))
	err := forEach(ctx, len(keys), maxWorkers, func(ctx context.Context, i int) error {
		index, err := c.archiveIndex(ctx, keys[i])
		if err != nil {
			return err
		}
		// Charged once per verified index, whether cached or downloaded, so a
		// failed attempt on one host never consumes budget.
		if consumed.Add(int64(len(index.data))) > c.indexBudget {
			return fmt.Errorf("archive index bytes exceed %d", c.indexBudget)
		}
		indices[i] = index
		return nil
	})
	if err != nil {
		return err
	}
	c.indices = indices
	return nil
}

func (c *cdnClient) archiveIndex(ctx context.Context, k key) (archiveIndex, error) {
	var parsed archiveIndex
	_, err := c.cached(ctx, filepath.Join("tpr", "wow", "data", hex.EncodeToString(k[:])+".index"), archiveIndexLimit,
		func(data []byte) error {
			index, err := parseArchiveIndex(data, k)
			if err == nil {
				parsed = index
			}
			return err
		}, func() ([]byte, error) {
			return c.download(ctx, cdnPath("wow", "data", k)+".index", archiveIndexLimit, nil)
		})
	return parsed, err
}

func (c *cdnClient) cached(ctx context.Context, relative string, limit int64, validate func([]byte) error, acquire func() ([]byte, error)) ([]byte, error) {
	lock := &c.cacheLocks[byte(md5.Sum([]byte(relative))[0])%byte(len(c.cacheLocks))]
	lock.Lock()
	defer lock.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	path := filepath.Join(c.cache, relative)
	if err := cdnCheckParents(c.cache, path); err != nil {
		return nil, err
	}
	data, exists, err := cdnReadFile(path, limit)
	if err != nil {
		return nil, err
	}
	if exists {
		if err := validate(data); err != nil {
			return nil, fmt.Errorf("modified CDN cache %s: %w", path, err)
		}
		return data, nil
	}
	data, err = acquire()
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("CDN content exceeds %d bytes", limit)
	}
	if err := validate(data); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := cdnWriteFile(c.cache, path, data); err != nil {
		return nil, err
	}
	return data, nil
}

func cdnReadFile(path string, limit int64) ([]byte, bool, error) {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	if !info.Mode().IsRegular() || info.Size() > limit {
		return nil, false, fmt.Errorf("nonregular or oversized CDN cache: %s", path)
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, false, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil {
		return nil, false, fmt.Errorf("read CDN cache %s: %w", path, err)
	}
	if int64(len(data)) > limit {
		return nil, false, fmt.Errorf("CDN cache %s grew beyond %d bytes while reading", path, limit)
	}
	return data, true, nil
}

// cdnCheckParents rejects symlinks and non-directories between path and the
// resolved cache root. Ancestors above the root were resolved by EvalSymlinks
// when the root was prepared.
func cdnCheckParents(root, path string) error {
	for current := filepath.Dir(path); ; current = filepath.Dir(current) {
		info, err := os.Lstat(current)
		if err == nil && (info.Mode()&os.ModeSymlink != 0 || !info.IsDir()) {
			return fmt.Errorf("CDN cache parent must be a directory without symlinks: %s", current)
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		if current == root {
			return nil
		}
		if filepath.Dir(current) == current {
			return fmt.Errorf("CDN cache path %s is outside cache root %s", path, root)
		}
	}
}

func cdnWriteFile(root, path string, data []byte) error {
	if err := cdnCheckParents(root, path); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".cdn-")
	if err != nil {
		return err
	}
	temporary := file.Name()
	defer os.Remove(temporary)
	if _, err := file.Write(data); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(temporary, path)
}

func (c *cdnClient) download(ctx context.Context, path string, limit int64, location *archiveLocation) ([]byte, error) {
	var last error
	for _, host := range c.hosts {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		data, err := c.request(ctx, host+"/"+path, limit, location)
		if err == nil {
			return data, nil
		}
		last = err
	}
	return nil, fmt.Errorf("CDN acquisition %s failed on %d configured hosts: %w", path, len(c.hosts), last)
}

func (c *cdnClient) request(ctx context.Context, url string, limit int64, location *archiveLocation) ([]byte, error) {
	requestContext, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(requestContext, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "wow-teal-types-cdn")
	request.Header.Set("Accept-Encoding", "identity")
	wantStatus := http.StatusOK
	if location != nil {
		wantStatus = http.StatusPartialContent
		request.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", location.offset, uint64(location.offset)+uint64(location.size)-1))
	}
	response, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != wantStatus {
		return nil, fmt.Errorf("%s returned HTTP %d, expected %d", url, response.StatusCode, wantStatus)
	}
	if value := response.Header.Get("Content-Encoding"); value != "" && value != "identity" {
		return nil, fmt.Errorf("%s returned unsupported content encoding %q", url, value)
	}
	if response.ContentLength > limit {
		return nil, fmt.Errorf("%s response exceeds %d bytes", url, limit)
	}
	if location != nil {
		if err := checkContentRange(response.Header.Get("Content-Range"), *location); err != nil {
			return nil, fmt.Errorf("%s: %w", url, err)
		}
		if response.ContentLength >= 0 && response.ContentLength != int64(location.size) {
			return nil, fmt.Errorf("%s range Content-Length differs from index", url)
		}
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", url, err)
	}
	if int64(len(body)) > limit {
		return nil, fmt.Errorf("%s response exceeds %d bytes", url, limit)
	}
	if location != nil && len(body) != int(location.size) {
		return nil, fmt.Errorf("%s range body length differs from index", url)
	}
	return body, nil
}

func checkContentRange(text string, location archiveLocation) error {
	rangeText, ok := strings.CutPrefix(text, "bytes ")
	if !ok {
		return errors.New("missing or invalid Content-Range")
	}
	interval, totalText, ok := strings.Cut(rangeText, "/")
	if !ok {
		return errors.New("invalid Content-Range total")
	}
	startText, endText, ok := strings.Cut(interval, "-")
	if !ok {
		return errors.New("invalid Content-Range interval")
	}
	start, a := strconv.ParseUint(startText, 10, 64)
	end, b := strconv.ParseUint(endText, 10, 64)
	total, d := strconv.ParseUint(totalText, 10, 64)
	if a != nil || b != nil || d != nil || start != uint64(location.offset) || end != start+uint64(location.size)-1 || total <= end {
		return errors.New("Content-Range does not match the archive index")
	}
	return nil
}
