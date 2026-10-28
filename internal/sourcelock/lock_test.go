package sourcelock

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const gitSource = `{"blizzard":{"commit":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","files":[{"name":"A.lua","sha":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","size":0}],"path":"","repository":"fixture/source"}}`
const cascVersion = `{"blizzard_version":{"client_lock_sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","files":[{"client_path":null,"name":"version.txt","sha":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","size":0}],"kind":"casc","normalization":"lf-final-newline","path":""}}`

func TestSourceSchemaDistinguishesAbsentNullAndZero(t *testing.T) {
	git, err := parseSources([]byte(gitSource))
	if err != nil {
		t.Fatal(err)
	}
	if git["blizzard"].Files[0].ClientPath.Present || git["blizzard"].Files[0].Size != 0 {
		t.Fatal("Git absence or zero size lost")
	}
	casc, err := parseSources([]byte(cascVersion))
	if err != nil {
		t.Fatal(err)
	}
	file := casc["blizzard_version"].Files[0]
	if !file.ClientPath.Present || file.ClientPath.Value != nil {
		t.Fatal("CASC null version path lost")
	}
	for _, source := range []Sources{git, casc} {
		data, err := json.Marshal(source)
		if err != nil {
			t.Fatal(err)
		}
		again, err := parseSources(data)
		if err != nil {
			t.Fatal(err)
		}
		for name, group := range source {
			if !group.Files[0].ClientPath.Equal(again[name].Files[0].ClientPath) {
				t.Fatal("source path presence changed")
			}
		}
	}
}
func TestSourceSchemaRejectsMalformedShapes(t *testing.T) {
	for _, input := range []string{
		`null`, `[]`, gitSource + `{}`, strings.Replace(gitSource, `"size":0`, `"size":null`, 1), strings.Replace(gitSource, `,"size":0`, "", 1), strings.Replace(gitSource, `"size":0`, `"size":0.5`, 1), strings.Replace(gitSource, `"size":0`, `"size":99999999999999999999999`, 1), strings.Replace(gitSource, `"name":"A.lua"`, `"name":true`, 1), strings.Replace(gitSource, `"path":""`, `"path":null`, 1), strings.Replace(gitSource, `"name":"A.lua"`, `"client_path":null,"name":"A.lua"`, 1), strings.Replace(cascVersion, `"client_path":null,`, "", 1), strings.Replace(cascVersion, `"client_path":null`, `"client_path":12`, 1), strings.Replace(cascVersion, `"client_path":null`, `"client_path":"../outside"`, 1),
	} {
		if _, err := parseSources([]byte(input)); err == nil {
			t.Fatalf("accepted malformed lock: %s", input)
		}
	}
}
func TestPublishedClientLockRoundTripPreservesBytes(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("../..", "internal/wowapi/inputs/client_sources.json"))
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := parseSnapshot(data)
	if err != nil {
		t.Fatal(err)
	}
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err = encoder.Encode(snapshot); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buffer.Bytes(), data) {
		t.Fatal("client lock serialization changed")
	}
	for _, malformed := range []string{`null`, `[]`, string(data) + `{}`, strings.Replace(string(data), `"build": 70575`, `"build": 70575.5`, 1), strings.Replace(string(data), `"schema": 1`, `"schema": null`, 1), strings.Replace(string(data), `"size": 2641`, `"size": null`, 1), strings.Replace(string(data), `"files": [`, `"extractor": {"implementation": "wow-teal-types/tools/cdn"},
  "files": [`, 1)} {
		if _, err := parseSnapshot([]byte(malformed)); err == nil {
			t.Fatal("accepted malformed client lock")
		}
	}
}
