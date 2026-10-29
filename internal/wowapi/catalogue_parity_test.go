package wowapi_test

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"wow-teal-types/internal/release"
	"wow-teal-types/internal/sourcelock"
	"wow-teal-types/internal/wowapi"
)

func catalogueTestRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("missing test location")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../.."))
}
func catalogueTestProject(t *testing.T) wowapi.Project {
	t.Helper()
	root := catalogueTestRoot(t)
	lock, err := sourcelock.ReadSnapshot(filepath.Join(root, "internal/wowapi/inputs/client_sources.json"))
	if err != nil {
		t.Fatal(err)
	}
	snapshot := filepath.Join(root, "build/wow-client", strconv.Itoa(lock.Build))
	if _, err := os.Stat(snapshot); os.IsNotExist(err) {
		t.Skip("client snapshot absent; run bootstrap first")
	} else if err != nil {
		t.Fatal(err)
	}
	project, err := release.Bootstrap(t.Context(), root, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	return project
}

func TestGenerateMatchesCheckedDeclarationsAndIsDeterministic(t *testing.T) {
	root := catalogueTestRoot(t)
	p := catalogueTestProject(t)
	first, err := p.Generate(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	names := []string{"types/wow/wow.d.tl", "types/wow/wow_env.d.tl", "types/wow/wow_secure_env.d.tl", "types/wow/call_policy.lua"}
	if len(first.Files) != len(names) {
		t.Fatalf("generated %d files, want %d", len(first.Files), len(names))
	}
	for _, name := range names {
		actual := first.Files[name]
		expected, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(actual, expected) {
			t.Errorf("generated artifact differs from checked bytes: %s", name)
		}
	}

	second, err := catalogueTestProject(t).Generate(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Files) != len(second.Files) {
		t.Fatal("independent generation changed artifact inventory")
	}
	for name, expected := range first.Files {
		if !bytes.Equal(second.Files[name], expected) {
			t.Errorf("independent generation changed %s", name)
		}
	}
}
