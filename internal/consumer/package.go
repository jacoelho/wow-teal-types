package consumer

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"time"
)

func Package(root string) (PackageResult, error) {
	manifest, err := Verify(root)
	if err != nil {
		return PackageResult{}, err
	}
	names := map[string]bool{}
	for name := range manifest.Files {
		names[name] = true
	}
	for _, name := range []string{"package.json", "LICENSE", "README.md", "UPDATING.md", "tlconfig.lua", "docs/api/teal-catalogue.md", "docs/architecture.md", "tools/teal/tests/compiler.lua", "tools/teal/README.md", "tools/teal/provenance.json", "tools/teal/unknown.patch", "tools/teal/rebuild.py", "tools/teal/upstream/tl.lua", "examples/minimal/WowTealExample.tl", "examples/minimal/WowTealExample.toc", "check-all.sh", "Makefile", "rockspecs/wow-tl-type-" + manifest.Version + "-1.rockspec.in"} {
		names[name] = true
	}
	// Tests exercise the same consumer closure without importing producer packages.
	files, err := filepath.Glob(filepath.Join(root, "internal/consumer/*_test.go"))
	if err != nil {
		return PackageResult{}, err
	}
	for _, path := range files {
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return PackageResult{}, err
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
			return PackageResult{}, err
		}
	}
	output := filepath.Join(root, "dist")
	if err = os.MkdirAll(output, 0755); err != nil {
		return PackageResult{}, err
	}
	temporary, err := os.CreateTemp(output, ".consumer-*.zip")
	if err != nil {
		return PackageResult{}, err
	}
	defer os.Remove(temporary.Name())
	zipped := zip.NewWriter(temporary)
	zipped.RegisterCompressor(zip.Deflate, func(writer io.Writer) (io.WriteCloser, error) { return flate.NewWriter(writer, flate.BestCompression) })
	for _, name := range slices.Sorted(maps.Keys(names)) {
		path, e := checkedPath(root, name)
		if e != nil {
			_ = zipped.Close()
			_ = temporary.Close()
			return PackageResult{}, e
		}
		data, e := os.ReadFile(path)
		if e != nil {
			_ = zipped.Close()
			_ = temporary.Close()
			return PackageResult{}, e
		}
		if name == "Makefile" {
			data = []byte(consumerMakefile)
		}
		header := &zip.FileHeader{Name: "wow-teal-types/" + name, Method: zip.Deflate}
		header.SetModTime(time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC))
		mode := consumerFileMode(name)
		header.SetMode(mode)
		entry, e := zipped.CreateHeader(header)
		if e == nil {
			_, e = entry.Write(data)
		}
		if e != nil {
			_ = zipped.Close()
			_ = temporary.Close()
			return PackageResult{}, e
		}
	}
	if err = zipped.Close(); err != nil {
		_ = temporary.Close()
		return PackageResult{}, err
	}
	if err = temporary.Close(); err != nil {
		return PackageResult{}, err
	}
	reader, err := zip.OpenReader(temporary.Name())
	if err != nil {
		return PackageResult{}, err
	}
	bound := manifest.Files
	expectedNames := slices.Sorted(maps.Keys(names))
	if len(reader.File) != len(expectedNames) {
		_ = reader.Close()
		return PackageResult{}, fmt.Errorf("archive member inventory mismatch")
	}
	for i, entry := range reader.File {
		if entry.Name != "wow-teal-types/"+expectedNames[i] {
			_ = reader.Close()
			return PackageResult{}, fmt.Errorf("archive member order mismatch")
		}
		if expected, exists := bound[expectedNames[i]]; exists {
			source, e := entry.Open()
			if e != nil {
				_ = reader.Close()
				return PackageResult{}, e
			}
			data, e := io.ReadAll(source)
			closeErr := source.Close()
			if e != nil {
				_ = reader.Close()
				return PackageResult{}, e
			}
			if closeErr != nil {
				_ = reader.Close()
				return PackageResult{}, closeErr
			}
			if fingerprint(data) != expected {
				_ = reader.Close()
				return PackageResult{}, fmt.Errorf("archive differs from package manifest: %s", expectedNames[i])
			}
		}
	}
	if err = reader.Close(); err != nil {
		return PackageResult{}, err
	}
	current, err := Verify(root)
	if err != nil {
		return PackageResult{}, err
	}
	if !sameManifest(current, manifest) {
		return PackageResult{}, fmt.Errorf("package changed during archiving")
	}
	data, err := os.ReadFile(temporary.Name())
	if err != nil {
		return PackageResult{}, err
	}
	target := filepath.Join(output, "wow-teal-types-"+manifest.Version+"-retail-"+manifest.WoWBuild+".zip")
	if err = os.Rename(temporary.Name(), target); err != nil {
		return PackageResult{}, err
	}
	hash := fingerprint(data)
	if err = atomicWrite(target+".sha256", bytes.NewBufferString(hash+"  "+filepath.Base(target)+"\n").Bytes(), 0644); err != nil {
		return PackageResult{}, err
	}
	return PackageResult{Archive: target, SHA256: hash, Members: len(names)}, nil
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
