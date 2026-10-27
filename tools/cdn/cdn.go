package main

import (
	"bytes"
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
	cache        string
	installation string
	client       *http.Client
	hosts        []string
	indices      []archiveIndex
	cacheLocks   [64]sync.Mutex
}

func newCDN(cache, installation string) *cdnClient {
	return &cdnClient{cache: cache, installation: installation,
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
		}, func() ([]byte, error) { return c.download(ctx, cdnPath(namespace, kind, k), 8<<20, nil, nil) })
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
					return c.download(ctx, cdnPath("wow", "data", location.archive), limit, &location, nil)
				}
			}
			return c.download(ctx, cdnPath("wow", "data", k), limit, nil, nil)
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
	run, cancel := context.WithCancel(ctx)
	defer cancel()
	var consumed, next atomic.Int64
	account := func(count int) error {
		if consumed.Add(int64(count)) > archiveIndexTotal {
			return errors.New("archive index reads exceed 384 MiB")
		}
		return nil
	}
	indices := make([]archiveIndex, len(keys))
	var firstError error
	var once sync.Once
	var workers sync.WaitGroup
	for range min(4, len(keys)) {
		workers.Go(func() {
			for {
				if run.Err() != nil {
					return
				}
				i := int(next.Add(1) - 1)
				if i >= len(keys) {
					return
				}
				index, err := c.archiveIndex(run, keys[i], account)
				if err != nil {
					once.Do(func() { firstError = err; cancel() })
					return
				}
				indices[i] = index
			}
		})
	}
	workers.Wait()
	if firstError != nil {
		return firstError
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	c.indices = indices
	return nil
}

func (c *cdnClient) archiveIndex(ctx context.Context, k key, account func(int) error) (archiveIndex, error) {
	var parsed archiveIndex
	charged := false
	_, err := c.cached(ctx, filepath.Join("tpr", "wow", "data", hex.EncodeToString(k[:])+".index"), archiveIndexLimit,
		func(data []byte) error {
			if !charged {
				if err := account(len(data)); err != nil {
					return err
				}
			}
			index, err := parseArchiveIndex(data, k)
			if err == nil {
				parsed = index
			}
			return err
		}, func() ([]byte, error) {
			if c.installation != "" {
				path := filepath.Join(c.installation, "Data", "indices", hex.EncodeToString(k[:])+".index")
				data, exists, err := cdnReadFile(path, archiveIndexLimit)
				if err != nil {
					return nil, err
				}
				if exists {
					charged = true
					return data, account(len(data))
				}
			}
			data, err := c.download(ctx, cdnPath("wow", "data", k)+".index", archiveIndexLimit, nil, account)
			charged = err == nil
			return data, err
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
	if err := cdnCheckParents(path); err != nil {
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
	if err := cdnWriteFile(path, data); err != nil {
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
	if err != nil || int64(len(data)) > limit {
		return nil, false, fmt.Errorf("read bounded CDN cache %s: %w", path, errors.Join(err, errors.New("content limit or read failure")))
	}
	return data, true, nil
}

func cdnCheckParents(path string) error {
	for current := filepath.Dir(path); ; current = filepath.Dir(current) {
		info, err := os.Lstat(current)
		if err == nil && (info.Mode()&os.ModeSymlink != 0 || !info.IsDir()) {
			return fmt.Errorf("CDN cache parent must be a directory without symlinks: %s", current)
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		if filepath.Dir(current) == current {
			return nil
		}
	}
}

func cdnWriteFile(path string, data []byte) error {
	if err := cdnCheckParents(path); err != nil {
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

func (c *cdnClient) download(ctx context.Context, path string, limit int64, location *archiveLocation, account func(int) error) ([]byte, error) {
	var last error
	for _, host := range c.hosts {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		data, err := c.request(ctx, host+"/"+path, limit, location, account)
		if err == nil {
			return data, nil
		}
		last = err
	}
	return nil, fmt.Errorf("CDN acquisition %s failed on %d configured hosts: %w", path, len(c.hosts), last)
}

func (c *cdnClient) request(ctx context.Context, url string, limit int64, location *archiveLocation, account func(int) error) ([]byte, error) {
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
	var body bytes.Buffer
	buffer := make([]byte, 64<<10)
	for {
		count, readErr := response.Body.Read(buffer)
		if count > 0 {
			if account != nil {
				if err := account(count); err != nil {
					return nil, err
				}
			}
			if int64(body.Len())+int64(count) > limit {
				return nil, fmt.Errorf("%s response exceeds %d bytes", url, limit)
			}
			body.Write(buffer[:count])
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, fmt.Errorf("read %s: %w", url, readErr)
		}
	}
	if location != nil && body.Len() != int(location.size) {
		return nil, fmt.Errorf("%s range body length differs from index", url)
	}
	return body.Bytes(), nil
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
