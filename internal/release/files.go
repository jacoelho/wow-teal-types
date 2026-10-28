// Package release owns verified acquisition, isolated candidates, and publication.
package release

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"wow-teal-types/internal/sourcelock"
)

const maxFile = 8 << 20
const maxPayload = 64 << 20
const maxTotal = 384 << 20

func keys[V any](m map[string]V) []string {
	r := make([]string, 0, len(m))
	for k := range m {
		r = append(r, k)
	}
	sort.Strings(r)
	return r
}
func sha(b []byte) string        { x := sha256.Sum256(b); return hex.EncodeToString(x[:]) }
func contentKey(b []byte) string { x := md5.Sum(b); return hex.EncodeToString(x[:]) }
func jsonBytes[T any](v T) ([]byte, error) {
	var b bytes.Buffer
	e := json.NewEncoder(&b)
	e.SetEscapeHTML(false)
	e.SetIndent("", "  ")
	if err := e.Encode(v); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
func readJSON[T any](path string) (T, error) {
	var value T
	data, err := sourcelock.ReadRegular(path, 128<<20)
	if err != nil {
		return value, err
	}
	if err = sourcelock.DecodeStrict(data, &value); err != nil {
		return value, fmt.Errorf("%s: %w", path, err)
	}
	return value, nil
}
func safeJoin(root, name string) (string, error) {
	if name == "" || filepath.IsAbs(name) || strings.ContainsAny(name, "\\\x00") || strings.Contains(name, ":") {
		return "", fmt.Errorf("unsafe path: %q", name)
	}
	for part := range strings.SplitSeq(name, "/") {
		if part == "" || part == "." || part == ".." {
			return "", fmt.Errorf("unsafe path: %q", name)
		}
	}
	p := filepath.Join(root, filepath.FromSlash(name))
	for q := p; ; q = filepath.Dir(q) {
		st, e := os.Lstat(q)
		if e == nil && st.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("path contains symlink: %s", q)
		}
		if e != nil && !os.IsNotExist(e) {
			return "", e
		}
		if q == filepath.Dir(q) || q == root {
			break
		}
	}
	return p, nil
}

func atomicWrite(path string, b []byte, mode fs.FileMode) error {
	if e := os.MkdirAll(filepath.Dir(path), 0755); e != nil {
		return e
	}
	f, e := os.CreateTemp(filepath.Dir(path), ".wow-types-")
	if e != nil {
		return e
	}
	name := f.Name()
	defer os.Remove(name)
	_, we := f.Write(b)
	ce := f.Close()
	if e = errors.Join(we, ce); e != nil {
		return e
	}
	if e = os.Chmod(name, mode); e != nil {
		return e
	}
	return os.Rename(name, path)
}
func writeObject[T any](path string, v T) error {
	b, e := jsonBytes(v)
	if e != nil {
		return e
	}
	return atomicWrite(path, b, 0644)
}
func command(ctx context.Context, root, log string, timeout time.Duration, args ...string) error {
	child, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if e := os.MkdirAll(filepath.Dir(log), 0755); e != nil {
		return e
	}
	f, e := os.Create(log)
	if e != nil {
		return e
	}
	cmd := exec.CommandContext(child, args[0], args[1:]...)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOWORK=off", "GOFLAGS=")
	cmd.Stdout = f
	cmd.Stderr = f
	e = cmd.Run()
	ce := f.Close()
	if e = errors.Join(e, ce); e != nil {
		b, _ := os.ReadFile(log)
		if len(b) > 4096 {
			b = b[len(b)-4096:]
		}
		return fmt.Errorf("command %s failed; see %s: %w: %s", strings.Join(args, " "), log, e, strings.TrimSpace(string(b)))
	}
	return nil
}
func exists(path string) bool { _, e := os.Lstat(path); return e == nil }
func copyFile(src, dst string) error {
	b, e := sourcelock.ReadRegular(src, maxPayload)
	if e != nil {
		return e
	}
	st, e := os.Stat(src)
	if e != nil {
		return e
	}
	return atomicWrite(dst, b, st.Mode().Perm())
}

// unmaintainedDir names build and VCS directories that maintained-tree walks skip.
func unmaintainedDir(name string) bool {
	return name == "target" || name == ".git" || name == "__pycache__"
}
func copyDirectory(src, dst string) error {
	total := int64(0)
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if path != src && d.IsDir() && unmaintainedDir(d.Name()) {
			return filepath.SkipDir
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("maintained source contains symlink: %s", path)
		}
		if d.IsDir() {
			return nil
		}
		st, e := d.Info()
		if e != nil {
			return e
		}
		if !st.Mode().IsRegular() {
			return fmt.Errorf("nonregular maintained file: %s", path)
		}
		total += st.Size()
		if total > maxTotal {
			return fmt.Errorf("maintained source exceeds aggregate limit")
		}
		r, e := filepath.Rel(src, path)
		if e != nil {
			return e
		}
		return copyFile(path, filepath.Join(dst, r))
	})
}
