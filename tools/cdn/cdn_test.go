package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func testCDN(t *testing.T, server *httptest.Server) *cdnClient {
	t.Helper()
	cache, err := prepareDirectory(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	c := newCDN(cache, "")
	c.hosts = []string{server.URL}
	t.Cleanup(c.close)
	return c
}

func TestCDNConfigAndLooseContentVerifyCachedBytes(t *testing.T) {
	config := []byte("build-name = fixture\n")
	blte := []byte("BLTE\x00\x00\x00\x00Nhello")
	k, ekey := key(md5.Sum(config)), key(md5.Sum(blte))
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if strings.Contains(r.URL.Path, "/config/") {
			w.Write(config)
		} else {
			w.Write(blte)
		}
	}))
	defer server.Close()
	c := testCDN(t, server)
	for range 2 {
		data, err := c.config(t.Context(), "wow", k)
		if err != nil || !bytes.Equal(data, config) {
			t.Fatalf("config = %q, error = %v", data, err)
		}
		data, err = c.encoded(t.Context(), ekey, 1024)
		if err != nil || !bytes.Equal(data, blte) {
			t.Fatalf("encoded content = %q, error = %v", data, err)
		}
	}
	if requests.Load() != 2 {
		t.Fatalf("cache reuse issued %d HTTP requests", requests.Load())
	}
	path := filepath.Join(c.cache, "tpr/wow/data", hex.EncodeToString(ekey[:]))
	if err := os.WriteFile(path, append(blte[:len(blte)-1:len(blte)-1], 'x'), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := c.encoded(t.Context(), ekey, 1024); err == nil || !strings.Contains(err.Error(), "modified CDN cache") {
		t.Fatalf("modified cache error = %v", err)
	}
	if requests.Load() != 2 {
		t.Fatal("modified cache silently downloaded replacement bytes")
	}
}

func TestCDNArchiveRangeBoundary(t *testing.T) {
	blte := []byte("BLTE\x00\x00\x00\x00Nhello")
	ek := key(md5.Sum(blte))
	index, archive := httpIndex(ek, uint32(len(blte)))
	for _, mode := range []string{"correct", "ignored", "wrong-range", "short", "oversize", "corrupt"} {
		t.Run(mode, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.URL.Path, ".index") {
					w.Write(index)
					return
				}
				if got := r.Header.Get("Range"); got != "bytes=64-77" {
					t.Errorf("Range = %q", got)
				}
				if mode == "ignored" {
					w.Write(blte)
					return
				}
				contentRange := "bytes 64-77/128"
				if mode == "wrong-range" {
					contentRange = "bytes 63-76/128"
				}
				w.Header().Set("Content-Range", contentRange)
				w.WriteHeader(http.StatusPartialContent)
				switch mode {
				case "short":
					w.Write(blte[:len(blte)-1])
				case "oversize":
					w.(http.Flusher).Flush()
					w.Write(bytes.Repeat([]byte{'x'}, 100))
				case "corrupt":
					changed := append([]byte(nil), blte...)
					changed[len(changed)-1] = 'x'
					w.Write(changed)
				default:
					w.Write(blte)
				}
			}))
			defer server.Close()
			c := testCDN(t, server)
			facts, differences, err := c.prepareIndexes(t.Context(), map[string][]string{
				"archives": {hex.EncodeToString(archive[:])}, "archives-index-size": {"1"},
			})
			if err != nil || facts.ArchiveCount != 1 || facts.ArchiveEntries != 1 || len(differences) != 1 {
				t.Fatalf("index facts = %+v, differences = %+v, error = %v", facts, differences, err)
			}
			data, err := c.encoded(t.Context(), ek, 32)
			if mode == "correct" {
				if err != nil || !bytes.Equal(data, blte) {
					t.Fatalf("range content = %q, error = %v", data, err)
				}
			} else {
				if err == nil {
					t.Fatal("invalid archive response was accepted")
				}
				if _, statErr := os.Stat(filepath.Join(c.cache, "tpr/wow/data", hex.EncodeToString(ek[:]))); !errors.Is(statErr, os.ErrNotExist) {
					t.Fatalf("invalid response entered cache: %v", statErr)
				}
			}
		})
	}
}

func TestCDNChunkedPayloadIntegrityPrecedesCachePublication(t *testing.T) {
	data := make([]byte, 42)
	copy(data, "BLTE")
	binary.BigEndian.PutUint32(data[4:8], 36)
	data[8], data[11] = 15, 1
	binary.BigEndian.PutUint32(data[12:16], 6)
	binary.BigEndian.PutUint32(data[16:20], 5)
	copy(data[36:], "Nhello")
	chunkHash := md5.Sum(data[36:])
	copy(data[20:36], chunkHash[:])
	ek := key(md5.Sum(data[:36]))
	for _, corrupt := range []bool{false, true} {
		t.Run(strconv.FormatBool(corrupt), func(t *testing.T) {
			response := append([]byte(nil), data...)
			if corrupt {
				response[len(response)-1] ^= 1
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write(response) }))
			defer server.Close()
			c := testCDN(t, server)
			actual, err := c.encoded(t.Context(), ek, 1024)
			if !corrupt {
				if err != nil || !bytes.Equal(actual, data) {
					t.Fatalf("valid chunked response = %x, error = %v", actual, err)
				}
			} else {
				if err == nil {
					t.Fatal("header-matched but corrupt chunk payload was accepted")
				}
				if _, err := os.Stat(filepath.Join(c.cache, "tpr/wow/data", hex.EncodeToString(ek[:]))); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("corrupt chunk payload entered cache: %v", err)
				}
			}
		})
	}
}

func TestCDNIndexWorkersAreBoundedAndCancellationLeavesNoContent(t *testing.T) {
	var active, peak atomic.Int32
	started, release := make(chan struct{}, 8), make(chan struct{})
	data := make(map[string][]byte)
	values := map[string][]string{}
	for tag := 1; tag <= 8; tag++ {
		body, k := literalIndex(t, tag)
		name := hex.EncodeToString(k[:])
		data[name+".index"] = body
		values["archives"] = append(values["archives"], name)
		values["archives-index-size"] = append(values["archives-index-size"], strconv.Itoa(len(body)))
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := active.Add(1)
		defer active.Add(-1)
		for old := peak.Load(); n > old && !peak.CompareAndSwap(old, n); old = peak.Load() {
		}
		started <- struct{}{}
		select {
		case <-release:
			w.Write(data[filepath.Base(r.URL.Path)])
		case <-r.Context().Done():
		}
	}))
	defer server.Close()
	c := testCDN(t, server)
	done := make(chan error, 1)
	go func() { _, _, err := c.prepareIndexes(t.Context(), values); done <- err }()
	for range 4 {
		select {
		case <-started:
		case <-time.After(5 * time.Second):
			close(release)
			t.Fatal("four workers did not reach the observable HTTP boundary")
		}
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if peak.Load() != 4 {
		t.Fatalf("maximum concurrent archive requests = %d", peak.Load())
	}

	entered := make(chan struct{})
	blocked := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		<-r.Context().Done()
	}))
	defer blocked.Close()
	other := testCDN(t, blocked)
	ctx, cancel := context.WithCancel(t.Context())
	go func() { _, err := other.encoded(ctx, key{1}, 1024); done <- err }()
	<-entered
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("cancellation error = %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("cancelled CDN work did not stop")
	}
	if err := filepath.WalkDir(other.cache, func(path string, entry os.DirEntry, err error) error {
		if err == nil && !entry.IsDir() {
			return fmt.Errorf("cancelled request left cache content %s", path)
		}
		return err
	}); err != nil {
		t.Fatal(err)
	}
}
