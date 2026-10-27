package consumer

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func fingerprint(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }
func fileHash(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return fingerprint(data), nil
}
func blobSHA(data []byte) string {
	sum := sha1.Sum(append(fmt.Appendf(nil, "blob %d%c", len(data), 0), data...))
	return hex.EncodeToString(sum[:])
}
func intValue(value any) int {
	switch x := value.(type) {
	case json.Number:
		n, _ := x.Int64()
		return int(n)
	case int:
		return x
	case float64:
		return int(x)
	}
	return -1
}
func object(value any) map[string]any { m, _ := value.(map[string]any); return m }
func str(value any) string            { s, _ := value.(string); return s }
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
func readJSON(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var value map[string]any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err = decoder.Decode(&value); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	var tail any
	if err = decoder.Decode(&tail); err != io.EOF {
		return nil, fmt.Errorf("%s: trailing JSON input", path)
	}
	if value == nil {
		return nil, fmt.Errorf("%s: expected JSON object", path)
	}
	return value, nil
}
func jsonBytes(value any) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
func validRelative(name string) bool {
	if name == "" || filepath.IsAbs(name) || strings.ContainsAny(name, "\\\x00") {
		return false
	}
	for part := range strings.SplitSeq(name, "/") {
		if part == "" || part == "." || part == ".." {
			return false
		}
	}
	return true
}
func checkedPath(root, name string, missing bool) (string, error) {
	if !validRelative(name) {
		return "", fmt.Errorf("invalid package manifest path: %s", name)
	}
	path := filepath.Join(root, filepath.FromSlash(name))
	current := path
	for {
		info, err := os.Lstat(current)
		if err == nil && info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("package path contains a symlink: %s", name)
		}
		if err != nil && !os.IsNotExist(err) {
			return "", err
		}
		if current == root {
			break
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("path escaped package root: %s", name)
		}
		current = parent
	}
	info, err := os.Lstat(path)
	if os.IsNotExist(err) && missing {
		return path, nil
	}
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("package input must be a regular file: %s", name)
	}
	return path, nil
}
func atomicWrite(path string, data []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".wow-teal-")
	if err != nil {
		return err
	}
	name := file.Name()
	defer os.Remove(name)
	_, writeErr := file.Write(data)
	closeErr := file.Close()
	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return closeErr
	}
	if err = os.Chmod(name, mode); err != nil {
		return err
	}
	return os.Rename(name, path)
}
func inside(path, parent string) bool {
	relative, err := filepath.Rel(parent, path)
	return err == nil && (relative == "." || relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)))
}
func canonicalExisting(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(absolute)
}
func canonicalOutput(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if info, e := os.Lstat(absolute); e == nil && info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("output must not be a symlink")
	}
	current := absolute
	var suffix []string
	for {
		resolved, e := filepath.EvalSymlinks(current)
		if e == nil {
			for i := len(suffix) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, suffix[i])
			}
			return resolved, nil
		}
		if !os.IsNotExist(e) {
			return "", e
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", e
		}
		suffix = append(suffix, filepath.Base(current))
		current = parent
	}
}
func copyTree(source, target string) error {
	return filepath.WalkDir(source, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		dest := filepath.Join(target, relative)
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("source contains a symlink: %s", path)
		}
		if entry.IsDir() {
			return os.MkdirAll(dest, 0755)
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("source must be a regular file: %s", path)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dest, data, info.Mode().Perm())
	})
}
