package consumer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
)

const maxInputBytes = 16 << 20

// buildInputs owns the captured namespace for one compiler invocation sequence.
// The compiler may write output beside it, but never into its search roots.
type buildInputs struct {
	directory    string
	packageRoot  string
	manifest     map[string]any
	manifestHash string
	search       []string
}

func (b *buildInputs) close() { _ = os.RemoveAll(b.directory) }

func (b *buildInputs) source(path string) string {
	return filepath.Join(b.directory, "inputs", strings.TrimPrefix(path, string(filepath.Separator)))
}

func readInput(path string, limit int64) ([]byte, error) {
	// Nonblocking open lets the descriptor check reject a FIFO even if the path
	// changed after directory traversal. Links are never admitted as file inputs.
	fd, err := syscall.Open(path, syscall.O_RDONLY|syscall.O_NONBLOCK|syscall.O_NOFOLLOW, 0)
	if err != nil {
		return nil, fmt.Errorf("read input %s: %w", path, err)
	}
	f := os.NewFile(uintptr(fd), path)
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > limit {
		return nil, fmt.Errorf("input must be a regular file of at most %d bytes: %s", limit, path)
	}
	data, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return nil, fmt.Errorf("read input %s: %w", path, err)
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("input exceeds %d bytes: %s", limit, path)
	}
	return data, nil
}

func inputRoot(path string) (string, error) {
	path = filepath.Clean(path)
	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("input root must not be a symlink: %s", path)
	}
	return canonicalExisting(path)
}

func captureBuildInputs(ctx context.Context, root string, sources, include []string) (_ *buildInputs, err error) {
	directory, err := os.MkdirTemp("", "wow-teal-")
	if err != nil {
		return nil, err
	}
	// Canonical filenames are also used by the Lua checker's declaration identity
	// checks; normalize the OS temporary-directory alias once at this boundary.
	physical, err := canonicalExisting(directory)
	if err != nil {
		_ = os.RemoveAll(directory)
		return nil, err
	}
	directory = physical
	b := &buildInputs{directory: directory, packageRoot: filepath.Join(directory, "package")}
	defer func() {
		if err != nil {
			b.close()
		}
	}()
	if err = b.capturePackage(ctx, root); err != nil {
		return nil, err
	}
	if err = b.captureSources(ctx, sources, include); err != nil {
		return nil, err
	}
	if err = os.Mkdir(filepath.Join(directory, "environment"), 0755); err != nil {
		return nil, err
	}
	if err = os.WriteFile(filepath.Join(directory, "tlconfig.lua"), []byte("return {}\n"), 0644); err != nil {
		return nil, err
	}
	b.search = []string{filepath.Join(b.packageRoot, "types/wow"), filepath.Join(directory, "environment")}
	for _, path := range include {
		b.search = append(b.search, b.source(path))
	}
	return b, nil
}

func (b *buildInputs) capturePackage(ctx context.Context, root string) error {
	path, err := checkedPath(root, "package.json", false)
	if err != nil {
		return err
	}
	data, err := readInput(path, maxInputBytes)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err = decoder.Decode(&b.manifest); err != nil {
		return fmt.Errorf("package manifest: %w", err)
	}
	var tail any
	if err = decoder.Decode(&tail); err != io.EOF {
		return fmt.Errorf("package manifest: trailing JSON input")
	}
	files := object(b.manifest["files"])
	if !reflect.DeepEqual(sortedKeys(files), RequiredFiles()) {
		return fmt.Errorf("package manifest inventory mismatch")
	}
	b.manifestHash = fingerprint(data)
	if err = atomicWrite(filepath.Join(b.packageRoot, "package.json"), data, 0644); err != nil {
		return err
	}
	total := len(data)
	for _, name := range RequiredFiles() {
		if err = ctx.Err(); err != nil {
			return err
		}
		path, err := checkedPath(root, name, false)
		if err != nil {
			return fmt.Errorf("package integrity mismatch: %s: %w", name, err)
		}
		data, err := readInput(path, int64(maxInputBytes-total))
		if err != nil {
			return err
		}
		if fingerprint(data) != files[name] {
			return fmt.Errorf("package integrity mismatch: %s", name)
		}
		total += len(data)
		if err = atomicWrite(filepath.Join(b.packageRoot, name), data, 0644); err != nil {
			return err
		}
	}
	selected, err := ReadVersion(b.packageRoot)
	if err != nil {
		return err
	}
	if b.manifest["wow_build"] != selected.Full {
		return fmt.Errorf("package WoW version differs from .wow-version")
	}
	return nil
}

func (b *buildInputs) captureSources(ctx context.Context, sources, include []string) error {
	seen := map[string]bool{}
	visited := map[string]bool{}
	total := 0
	copyInput := func(path string) error {
		if seen[path] {
			return nil
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		data, err := readInput(path, maxSourceBytes)
		if err != nil {
			return fmt.Errorf("dependency must be a regular file of at most 2 MiB: %w", err)
		}
		total += len(data)
		if total > maxInputBytes {
			return fmt.Errorf("source/dependency inputs exceed 16 MiB")
		}
		seen[path] = true
		return atomicWrite(b.source(path), data, 0644)
	}
	for _, source := range sources {
		if err := copyInput(source); err != nil {
			return err
		}
	}
	for _, directory := range include {
		// Even empty include roots must exist in the captured namespace.
		if err := os.MkdirAll(b.source(directory), 0755); err != nil {
			return err
		}
		err := filepath.WalkDir(directory, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if err = ctx.Err(); err != nil {
				return err
			}
			if path == b.directory {
				return filepath.SkipDir
			}
			if visited[path] {
				if entry.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if path != directory && excluded[entry.Name()] && entry.IsDir() {
				return filepath.SkipDir
			}
			visited[path] = true
			if len(visited) > maxEntries {
				return fmt.Errorf("dependency traversal exceeds 10000 entries")
			}
			if entry.Type()&os.ModeSymlink != 0 {
				return fmt.Errorf("input must not be a symlink: %s", path)
			}
			if entry.IsDir() {
				return nil
			}
			if strings.HasSuffix(path, ".tl") || strings.HasSuffix(path, ".lua") {
				return copyInput(path)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return ctx.Err()
}
