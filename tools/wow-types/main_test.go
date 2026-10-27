package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMutatingCommandsRejectIgnoredOptions(t *testing.T) {
	for _, command := range []string{"generate", "recover", "acquire-client", "bootstrap", "stage-release"} {
		if e := validateOptions(command, map[string]string{"check": "true"}, nil); e == nil {
			t.Fatalf("%s ignored --check", command)
		}
	}
	if e := validateOptions("generate", map[string]string{}, []string{"unexpected"}); e == nil {
		t.Fatal("ignored positional argument")
	}

}

func TestRemovedPublicationCommandsRejectWithoutWrites(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"catalogue", "coverage", "metadata", "capture", "probe"} {
		if err := run(t.Context(), []string{"--root", root, name}); err == nil || !strings.Contains(err.Error(), "unknown command") {
			t.Fatalf("removed %s command: %v", name, err)
		}
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatal("rejected commands changed the active tree")
	}
}
func TestCachePathUsesCallerDirectory(t *testing.T) {
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	for _, name := range []string{".wow-version", "internal/wowapi/inputs/client_sources.json"} {
		data, err := os.ReadFile(filepath.Join(repository, name))
		if err != nil {
			t.Fatal(err)
		}
		writeCommandFixture(t, filepath.Join(root, name), data)
	}
	writeCommandFixture(t, filepath.Join(root, "tools/cdn/go.mod"), []byte("module wow-teal-types/tools/cdn\n\ngo 1.27\n"))
	writeCommandFixture(t, filepath.Join(root, "tools/cdn/main.go"), []byte(`package main
import "os"
func main() {
	if err := os.WriteFile(os.Getenv("WOW_TEST_CACHE_ARGUMENT"), []byte(os.Args[3]), 0600); err != nil { panic(err) }
	os.Exit(1)
}
`))
	t.Chdir(t.TempDir())
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	observed := filepath.Join(cwd, "cache-argument")
	t.Setenv("WOW_TEST_CACHE_ARGUMENT", observed)
	// The extractor records its cache argument and stops before any CDN request.
	if err := run(t.Context(), []string{"--root", root, "acquire-client", "--cache", "build/cache"}); err == nil {
		t.Fatal("expected extractor failure")
	}
	argument, err := os.ReadFile(observed)
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(cwd, "build/cache"); string(argument) != want {
		t.Fatalf("extractor cache = %q; want caller path %q", argument, want)
	}
}

func writeCommandFixture(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
}
