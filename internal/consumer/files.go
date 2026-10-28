package consumer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
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
func checkedPath(root, name string) (string, error) {
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

func consumerFileMode(name string) os.FileMode {
	if name == "bin/wow-teal" || name == "check-all.sh" {
		return 0755
	}
	return 0644
}
