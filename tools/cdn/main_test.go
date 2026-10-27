package main

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const configFixture = `{"product":"wow","region":"eu","locale":"enUS","build":70575,"version":"12.1.5.70575","build_config":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","cdn_config":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","product_config":"cccccccccccccccccccccccccccccccc"}`

func TestRunRejectsOverlappingDirectoriesBeforeAcquisition(t *testing.T) {
	for _, mode := range []string{"identical", "nested-cache", "nested-output", "cache-alias", "output-alias", "missing-suffix"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			configPath := filepath.Join(root, "config.json")
			if err := os.WriteFile(configPath, []byte(configFixture), 0o600); err != nil {
				t.Fatal(err)
			}
			destination := filepath.Join(root, "destination")
			if err := os.Mkdir(destination, 0o700); err != nil {
				t.Fatal(err)
			}
			alias := filepath.Join(root, "alias")
			if err := os.Symlink(destination, alias); err != nil {
				t.Fatal(err)
			}
			output, cache := destination, destination
			switch mode {
			case "nested-cache":
				cache = filepath.Join(destination, "cache")
			case "nested-output":
				output = filepath.Join(destination, "output")
			case "cache-alias":
				cache = alias
			case "output-alias":
				output = alias
			case "missing-suffix":
				output = filepath.Join(destination, "missing", "output")
				cache = filepath.Join(alias, "missing", "output", "cache")
			}
			// Cancellation prevents network work if path validation regresses.
			ctx, cancel := context.WithCancel(t.Context())
			cancel()
			err := run(ctx, []string{configPath, output, cache})
			if err == nil || !strings.Contains(err.Error(), "must not overlap") {
				t.Fatalf("overlap error = %v", err)
			}
			requireNoAcquiredFiles(t, root, configPath)
		})
	}
}

func requireNoAcquiredFiles(t *testing.T, root, configPath string) {
	t.Helper()
	if err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path != configPath && entry.Type().IsRegular() {
			t.Errorf("overlap acquired file %s", path)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestPreparedDirectoriesAcceptDisjointDestinations(t *testing.T) {
	root := t.TempDir()
	alias := filepath.Join(root, "alias")
	if err := os.Symlink(root, alias); err != nil {
		t.Fatal(err)
	}
	resolved, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for _, relative := range []string{"output", "cache/nested"} {
		got, err := prepareDirectory(filepath.Join(alias, relative))
		want := filepath.Join(resolved, relative)
		if err != nil || got != want {
			t.Fatalf("prepared directory = %q, %v; want %q", got, err, want)
		}
		paths = append(paths, got)
	}
	if err := checkDirectorySeparation(paths[0], paths[1]); err != nil {
		t.Fatalf("disjoint directories rejected: %v", err)
	}
}

func TestRunRejectsCaseAliasedDirectoriesBeforeAcquisition(t *testing.T) {
	for _, mode := range []string{"existing", "missing-output", "missing-cache", "symlink-ancestor"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			probe := filepath.Join(root, "CaseProbe")
			if err := os.Mkdir(probe, 0o700); err != nil {
				t.Fatal(err)
			}
			original, err := os.Stat(probe)
			if err != nil {
				t.Fatal(err)
			}
			alias, err := os.Stat(filepath.Join(root, "caseprobe"))
			if os.IsNotExist(err) {
				t.Skip("filesystem preserves case-distinct directory identities")
			}
			if err != nil {
				t.Fatal(err)
			}
			if !os.SameFile(original, alias) {
				t.Skip("filesystem preserves case-distinct directory identities")
			}
			configPath := filepath.Join(root, "config.json")
			if err := os.WriteFile(configPath, []byte(configFixture), 0o600); err != nil {
				t.Fatal(err)
			}
			output, cache := filepath.Join(root, "Output"), filepath.Join(root, "output")
			switch mode {
			case "existing":
				if err := os.Mkdir(output, 0o700); err != nil {
					t.Fatal(err)
				}
			case "missing-output":
				cache = filepath.Join(cache, "cache")
			case "missing-cache":
				output, cache = filepath.Join(root, "cache", "output"), filepath.Join(root, "Cache")
			case "symlink-ancestor":
				real, link := filepath.Join(root, "real"), filepath.Join(root, "alias")
				if err := os.Mkdir(real, 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(real, link); err != nil {
					t.Fatal(err)
				}
				output, cache = filepath.Join(real, "Output"), filepath.Join(link, "output", "cache")
			}
			ctx, cancel := context.WithCancel(t.Context())
			cancel()
			err = run(ctx, []string{configPath, output, cache})
			if err == nil || !strings.Contains(err.Error(), "must not overlap") {
				t.Fatalf("case-alias overlap error = %v", err)
			}
			requireNoAcquiredFiles(t, root, configPath)
		})
	}
}
