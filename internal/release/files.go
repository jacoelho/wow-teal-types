// Package release owns verified acquisition, isolated candidates, and publication.
package release

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"
)

type M = map[string]any

const maxFile = 8 << 20
const maxPayload = 64 << 20
const maxTotal = 384 << 20

func s(v any) string { x, _ := v.(string); return x }
func i(v any) int {
	switch n := v.(type) {
	case json.Number:
		k, _ := n.Int64()
		return int(k)
	case int:
		return n
	case float64:
		return int(n)
	}
	return 0
}
func obj(v any) M     { x, _ := v.(map[string]any); return x }
func arr(v any) []any { x, _ := v.([]any); return x }
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
func blob(b []byte) string {
	h := sha1.New()
	fmt.Fprintf(h, "blob %d\x00", len(b))
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}
func jsonBytes(v any) ([]byte, error) {
	var b bytes.Buffer
	e := json.NewEncoder(&b)
	e.SetEscapeHTML(false)
	e.SetIndent("", "  ")
	if err := e.Encode(v); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
func readObject(path string) (M, error) {
	b, e := readFile(path, 128<<20)
	if e != nil {
		return nil, e
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	var m M
	if e = d.Decode(&m); e != nil {
		return nil, fmt.Errorf("%s: %w", path, e)
	}
	var rest any
	if e = d.Decode(&rest); e != io.EOF {
		return nil, fmt.Errorf("%s: trailing JSON", path)
	}
	if m == nil {
		return nil, fmt.Errorf("%s: expected object", path)
	}
	return m, nil
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
func readFile(path string, limit int64) ([]byte, error) {
	if _, e := safeJoin(filepath.Dir(path), filepath.Base(path)); e != nil {
		return nil, e
	}
	st, e := os.Stat(path)
	if e != nil {
		return nil, e
	}
	if !st.Mode().IsRegular() || st.Size() > limit {
		return nil, fmt.Errorf("nonregular or oversized file: %s", path)
	}
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	b, e := io.ReadAll(io.LimitReader(f, limit+1))
	if e != nil {
		return nil, e
	}
	if int64(len(b)) > limit {
		return nil, fmt.Errorf("oversized file: %s", path)
	}
	return b, nil
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
func writeObject(path string, v any) error {
	b, e := jsonBytes(v)
	if e != nil {
		return e
	}
	return atomicWrite(path, b, 0644)
}
func equal(a, b any) bool { return reflect.DeepEqual(a, b) }
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
	b, e := readFile(src, maxPayload)
	if e != nil {
		return e
	}
	st, e := os.Stat(src)
	if e != nil {
		return e
	}
	return atomicWrite(dst, b, st.Mode().Perm())
}
func copyDirectory(src, dst string) error {
	total := int64(0)
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if path != src && d.IsDir() && (d.Name() == "target" || d.Name() == ".git" || d.Name() == "__pycache__") {
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

func newJSONDecoder(b []byte) *json.Decoder {
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	return d
}
