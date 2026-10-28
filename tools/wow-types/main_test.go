package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRecoveryDoesNotRequireSelectedVersion(t *testing.T) {
	for _, version := range []string{"", "invalid\n"} {
		t.Run(version, func(t *testing.T) {
			root := t.TempDir()
			if version != "" {
				writeCommandFixture(t, filepath.Join(root, ".wow-version"), []byte(version))
			}
			if err := run(t.Context(), []string{"--root", root, "recover"}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestExplicitSnapshotVerificationDoesNotRequireSelectedVersion(t *testing.T) {
	for _, version := range []string{"", "invalid\n"} {
		for _, option := range []string{"", "--snapshot", "--client-source"} {
			t.Run(version+option, func(t *testing.T) {
				root := t.TempDir()
				if version != "" {
					writeCommandFixture(t, filepath.Join(root, ".wow-version"), []byte(version))
				}
				snapshot := filepath.Join(t.TempDir(), "missing-snapshot")
				args := []string{"--root", root, "verify-client"}
				if option != "" {
					args = append(args, option)
				}
				args = append(args, snapshot)
				err := run(t.Context(), args)
				var pathError *os.PathError
				if !errors.As(err, &pathError) || !os.IsNotExist(err) || pathError.Path != filepath.Join(snapshot, "snapshot.json") {
					t.Fatalf("expected missing snapshot.json error, got %v", err)
				}
			})
		}
	}
}

func TestDefaultSnapshotRequiresSelectedVersion(t *testing.T) {
	for _, command := range []string{"bootstrap", "acquire-client", "stage-client", "verify-client", "check-generator", "stage-release"} {
		t.Run(command, func(t *testing.T) {
			err := run(t.Context(), []string{"--root", t.TempDir(), command})
			if err == nil || !strings.Contains(err.Error(), "missing or invalid .wow-version") {
				t.Fatalf("expected version prerequisite, got %v", err)
			}
		})
	}
}

func TestExtractorBuildDoesNotRequireSelectedVersion(t *testing.T) {
	root := t.TempDir()
	writeCommandFixture(t, filepath.Join(root, "tools/cdn/go.mod"), []byte("module wow-teal-types/tools/cdn\n\ngo 1.27\n"))
	writeCommandFixture(t, filepath.Join(root, "tools/cdn/main.go"), []byte("package main\nfunc main() {}\n"))
	if err := run(t.Context(), []string{"--root", root, "build-extractor"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "build/direct-source-tools/wow-cdn")); err != nil {
		t.Fatal(err)
	}
}

func TestMutatingCommandsRejectIgnoredOptions(t *testing.T) {
	root := t.TempDir()
	for _, command := range []string{"generate", "recover", "acquire-client", "bootstrap", "stage-release"} {
		if e := run(t.Context(), []string{"--root", root, command, "--check", "true"}); e == nil || !strings.Contains(e.Error(), "unknown option --check") {
			t.Fatalf("%s ignored --check: %v", command, e)
		}
	}
	if e := run(t.Context(), []string{"--root", root, "generate", "unexpected"}); e == nil || !strings.Contains(e.Error(), "does not accept positional arguments") {
		t.Fatalf("ignored positional argument: %v", e)
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
