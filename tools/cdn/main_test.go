package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// testdata/extractor-config.json is the release-written config.json; release's
// schema test asserts it produces these exact bytes.
func configFixture(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/extractor-config.json")
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestRunRejectsOverlappingDirectoriesBeforeAcquisition(t *testing.T) {
	for _, mode := range []string{"identical", "nested-cache", "nested-output", "cache-alias", "output-alias", "missing-suffix"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			configPath := filepath.Join(root, "config.json")
			if err := os.WriteFile(configPath, configFixture(t), 0o600); err != nil {
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
			err := run(ctx, []string{configPath, output, cache}, io.Discard)
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
			if err := os.WriteFile(configPath, configFixture(t), 0o600); err != nil {
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
			err = run(ctx, []string{configPath, output, cache}, io.Discard)
			if err == nil || !strings.Contains(err.Error(), "must not overlap") {
				t.Fatalf("case-alias overlap error = %v", err)
			}
			requireNoAcquiredFiles(t, root, configPath)
		})
	}
}

func TestParseKeyRejectsMalformedKeys(t *testing.T) {
	for _, value := range []string{"", "a", strings.Repeat("a", 30), strings.Repeat("a", 34), strings.Repeat("A", 32), strings.Repeat("g", 32), strings.Repeat("0", 32)} {
		if _, err := parseKey(value); err == nil {
			t.Fatalf("accepted malformed key %q", value)
		}
	}
	got, err := parseKey("0102030405060708090a0b0c0d0e0f10")
	if err != nil || got != (key{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}) {
		t.Fatalf("valid key = %x, %v", got, err)
	}
}

func TestExtractorReceiptWireContract(t *testing.T) {
	data, err := os.ReadFile("testdata/extractor-result.json")
	if err != nil {
		t.Fatal(err)
	}
	receipt := extractorReceipt{
		clientIdentity: clientIdentity{Product: "wow", Region: "eu", Locale: "enUS", Build: 70575, Version: "12.1.5.70575", BuildConfig: strings.Repeat("a", 32), CDNConfig: strings.Repeat("b", 32), ProductConfig: strings.Repeat("c", 32)},
		Files:          []fileReceipt{{Path: "Interface/example.lua", Size: 5, SHA256: "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824", ContentKey: "5d41402abc4b2a76b9719d911017c592"}},
	}
	encoded, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	var got, want map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &want); err != nil {
		t.Fatal(err)
	}
	if len(got) != len(want) {
		t.Fatalf("receipt fields = %s; want %s", encoded, data)
	}
	for name, expected := range want {
		actual, ok := got[name]
		if !ok {
			t.Fatalf("receipt missing %s", name)
		}
		if name != "files" {
			if !bytes.Equal(actual, expected) {
				t.Errorf("%s = %s; want %s", name, actual, expected)
			}
			continue
		}
		var actualFiles, expectedFiles []fileReceipt
		if err := json.Unmarshal(actual, &actualFiles); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(expected, &expectedFiles); err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(actualFiles, expectedFiles) {
			t.Fatalf("receipt files = %+v; want %+v", actualFiles, expectedFiles)
		}
	}
}

func TestExtractorConfigWireContract(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, configFixture(t), 0o600); err != nil {
		t.Fatal(err)
	}
	job, err := readConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	want := []manifestID{{6067012, "Interface/ui-code-list.txt"}, {6076661, "Interface/ui-gen-addon-list.txt"}, {6067013, "Interface/ui-toc-list.txt"}}
	if !slices.Equal(job.manifests, want) {
		t.Fatalf("manifests = %+v; want %+v", job.manifests, want)
	}
	if job.identity.Build != 70575 || job.identity.Region != "eu" || job.buildName != "WOW-70575patch12.1.5_Retail" {
		t.Fatalf("identity = %+v, build name %q", job.identity, job.buildName)
	}
}

func TestExtractorConfigRejectsInvalidManifests(t *testing.T) {
	var base map[string]json.RawMessage
	if err := json.Unmarshal(configFixture(t), &base); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name  string
		field string
		value string
	}{
		{"missing", "manifests", ""},
		{"empty", "manifests", `{}`},
		{"null", "manifests", `null`},
		{"duplicate-id", "manifests", `{"Interface/a.txt": 7, "Interface/b.txt": 7}`},
		{"zero-id", "manifests", `{"Interface/a.txt": 0}`},
		{"negative-id", "manifests", `{"Interface/a.txt": -1}`},
		{"unsafe-path", "manifests", `{"Interface/../a.txt": 7}`},
		{"unknown-key", "installation", `"/Applications/World of Warcraft"`},
	} {
		t.Run(test.name, func(t *testing.T) {
			fields := maps.Clone(base)
			if test.value == "" {
				delete(fields, test.field)
			} else {
				fields[test.field] = json.RawMessage(test.value)
			}
			data, err := json.Marshal(fields)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(t.TempDir(), "config.json")
			if err := os.WriteFile(path, data, 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := readConfig(path); err == nil {
				t.Fatalf("accepted %s", data)
			}
		})
	}
}
