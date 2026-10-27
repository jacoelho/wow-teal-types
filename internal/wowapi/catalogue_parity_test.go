package wowapi

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func catalogueTestRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("missing test location")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../.."))
}
func TestGenerateMatchesCheckedDeclarationsAndIsDeterministic(t *testing.T) {
	root := catalogueTestRoot(t)
	p := NewProject(root)
	if _, err := os.Stat(p.InputPath("blizzard", "BuildDocumentation.lua")); err != nil {
		t.Skip("native inputs absent; run bootstrap first")
	}
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

	second, err := NewProject(root).Generate(t.Context())
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
