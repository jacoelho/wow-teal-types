package wowapi

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"unicode/utf16"
)

func catalogueTestRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("missing test location")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../.."))
}
func TestRecordedNativeLiteralsAreLossless(t *testing.T) {
	root := catalogueTestRoot(t)
	p := NewProject(root)
	if _, e := os.Stat(p.InputPath("blizzard", "BuildDocumentation.lua")); e != nil {
		t.Skip("native inputs absent; run bootstrap first")
	}
	report, e := readJSON(filepath.Join(root, "evidence/generated/wow_native_metadata.json"))
	if e != nil {
		t.Fatal(e)
	}
	for _, v := range asList(report["documents"]) {
		document := asMap(v)
		source := asString(document["source"])
		name := source[len("blizzard:"):]
		data, e := os.ReadFile(p.InputPath("blizzard", name))
		if e != nil {
			t.Fatal(e)
		}
		actual, e := ParseLua(data, name)
		if e != nil {
			t.Fatalf("%s: %v", name, e)
		}
		want := document["metadata"]
		a, _ := jsonCompact(actual)
		b, _ := jsonCompact(want)
		if !bytes.Equal(a, b) {
			t.Fatalf("native literal metadata changed: %s", name)
		}
	}
}
func TestGenerateMatchesReviewedArtifactsAndIsDeterministic(t *testing.T) {
	root := catalogueTestRoot(t)
	p := NewProject(root)
	if _, err := os.Stat(p.InputPath("blizzard", "BuildDocumentation.lua")); err != nil {
		t.Skip("native inputs absent; run bootstrap first")
	}
	first, err := p.Generate(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	for name, actual := range first.Files {
		expected, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(actual, expected) {
			t.Errorf("generated artifact differs from reviewed bytes: %s", name)
		}
	}

	// Compute the audit's byte contract independently of the renderer and the
	// evidence serializer, from the files returned to the generation caller.
	graph := map[string]string{}
	for _, name := range []string{"wow.d.tl", "wow_env.d.tl", "wow_secure_env.d.tl"} {
		graph[name] = string(first.Files["types/wow/"+name])
	}
	carrier, err := os.ReadFile(filepath.Join(root, "types/wow/native_values.d.tl"))
	if err != nil {
		t.Fatal(err)
	}
	var encoded bytes.Buffer
	encoder := json.NewEncoder(&encoded)
	encoder.SetEscapeHTML(false)
	if err = encoder.Encode(map[string]any{"generated": graph, "native_values": string(carrier)}); err != nil {
		t.Fatal(err)
	}
	var canonical bytes.Buffer
	for _, r := range string(bytes.TrimSuffix(encoded.Bytes(), []byte("\n"))) {
		if r < 128 {
			canonical.WriteByte(byte(r))
		} else {
			for _, unit := range utf16.Encode([]rune{r}) {
				fmt.Fprintf(&canonical, `\u%04x`, unit)
			}
		}
	}
	var report struct {
		GraphSHA256 string `json:"declaration_graph_sha256"`
	}
	if err = json.Unmarshal(first.Files["evidence/generated/wow_tsm_crosscheck.json"], &report); err != nil {
		t.Fatal(err)
	}
	if expected := fmt.Sprintf("%x", sha256.Sum256(canonical.Bytes())); report.GraphSHA256 != expected {
		t.Fatalf("audit graph fingerprint %s does not bind returned declaration bytes: %s", report.GraphSHA256, expected)
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
