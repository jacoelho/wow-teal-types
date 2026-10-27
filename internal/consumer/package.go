package consumer

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"time"
)

func Package(root string) (map[string]any, error) {
	manifest, err := Verify(root)
	if err != nil {
		return nil, err
	}
	names := map[string]bool{}
	for name := range object(manifest["files"]) {
		names[name] = true
	}
	for _, name := range []string{"package.json", "LICENSE", "README.md", "UPDATING.md", "tlconfig.lua", "docs/api/teal-catalogue.md", "docs/architecture.md", "examples/minimal/WowTealExample.tl", "examples/minimal/WowTealExample.toc", "check-all.sh", "Makefile", "rockspecs/wow-tl-type-" + str(manifest["version"]) + "-1.rockspec.in"} {
		names[name] = true
	}
	// Tests exercise the same consumer closure without importing producer packages.
	files, err := filepath.Glob(filepath.Join(root, "internal/consumer/*_test.go"))
	if err != nil {
		return nil, err
	}
	for _, path := range files {
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return nil, err
		}
		names[filepath.ToSlash(relative)] = true
	}
	fixtureRoot := filepath.Join(root, "internal/consumer/testdata")
	if info, e := os.Stat(fixtureRoot); e == nil && info.IsDir() {
		err = filepath.WalkDir(fixtureRoot, func(path string, entry os.DirEntry, e error) error {
			if e != nil {
				return e
			}
			if entry.Type()&os.ModeSymlink != 0 {
				return fmt.Errorf("package test fixture contains a symlink")
			}
			if entry.IsDir() {
				return nil
			}
			relative, e := filepath.Rel(root, path)
			if e != nil {
				return e
			}
			names[filepath.ToSlash(relative)] = true
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	output := filepath.Join(root, "dist")
	if err = os.MkdirAll(output, 0755); err != nil {
		return nil, err
	}
	temporary, err := os.CreateTemp(output, ".consumer-*.zip")
	if err != nil {
		return nil, err
	}
	defer os.Remove(temporary.Name())
	zipped := zip.NewWriter(temporary)
	zipped.RegisterCompressor(zip.Deflate, func(writer io.Writer) (io.WriteCloser, error) { return flate.NewWriter(writer, flate.BestCompression) })
	for _, name := range sortedKeys(names) {
		path, e := checkedPath(root, name, false)
		if e != nil {
			_ = zipped.Close()
			_ = temporary.Close()
			return nil, e
		}
		data, e := os.ReadFile(path)
		if e != nil {
			_ = zipped.Close()
			_ = temporary.Close()
			return nil, e
		}
		if name == "Makefile" {
			data = []byte(consumerMakefile)
		}
		header := &zip.FileHeader{Name: "wow-teal-types/" + name, Method: zip.Deflate}
		header.SetModTime(time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC))
		mode := os.FileMode(0644)
		if name == "bin/wow-teal" || name == "check-all.sh" {
			mode = 0755
		}
		header.SetMode(mode)
		entry, e := zipped.CreateHeader(header)
		if e == nil {
			_, e = entry.Write(data)
		}
		if e != nil {
			_ = zipped.Close()
			_ = temporary.Close()
			return nil, e
		}
	}
	if err = zipped.Close(); err != nil {
		_ = temporary.Close()
		return nil, err
	}
	if err = temporary.Close(); err != nil {
		return nil, err
	}
	reader, err := zip.OpenReader(temporary.Name())
	if err != nil {
		return nil, err
	}
	bound := object(manifest["files"])
	expectedNames := sortedKeys(names)
	if len(reader.File) != len(expectedNames) {
		_ = reader.Close()
		return nil, fmt.Errorf("archive member inventory mismatch")
	}
	for i, entry := range reader.File {
		if entry.Name != "wow-teal-types/"+expectedNames[i] {
			_ = reader.Close()
			return nil, fmt.Errorf("archive member order mismatch")
		}
		if expected, exists := bound[expectedNames[i]]; exists {
			source, e := entry.Open()
			if e != nil {
				_ = reader.Close()
				return nil, e
			}
			data, e := io.ReadAll(source)
			closeErr := source.Close()
			if e != nil {
				_ = reader.Close()
				return nil, e
			}
			if closeErr != nil {
				_ = reader.Close()
				return nil, closeErr
			}
			if fingerprint(data) != expected {
				_ = reader.Close()
				return nil, fmt.Errorf("archive differs from package manifest: %s", expectedNames[i])
			}
		}
	}
	if err = reader.Close(); err != nil {
		return nil, err
	}
	current, err := Verify(root)
	if err != nil {
		return nil, err
	}
	if !reflect.DeepEqual(current, manifest) {
		return nil, fmt.Errorf("package changed during archiving")
	}
	data, err := os.ReadFile(temporary.Name())
	if err != nil {
		return nil, err
	}
	target := filepath.Join(output, "wow-teal-types-"+str(manifest["version"])+"-retail-"+str(manifest["wow_build"])+".zip")
	if err = os.Rename(temporary.Name(), target); err != nil {
		return nil, err
	}
	hash := fingerprint(data)
	if err = atomicWrite(target+".sha256", bytes.NewBufferString(hash+"  "+filepath.Base(target)+"\n").Bytes(), 0644); err != nil {
		return nil, err
	}
	return map[string]any{"archive": target, "sha256": hash, "members": len(names)}, nil
}

// The archive exposes only commands whose implementation and inputs it contains.
const consumerMakefile = `.PHONY: check test example package

check:
	./check-all.sh

test:
	GOTOOLCHAIN=local GOWORK=off go test ./...

example:
	bin/wow-teal build examples/minimal --output build/WowTealExample

package:
	bin/wow-teal package
`
