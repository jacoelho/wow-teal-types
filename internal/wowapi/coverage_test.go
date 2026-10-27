package wowapi

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnvironmentCoverageAccountsForGlobalsAndReferences(t *testing.T) {
	directory := t.TempDir()
	write := func(name, text string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(directory, name), []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
	}
	write("wow.d.tl", `local record W
   interface Frame
   end
   type OpaqueNumber = userdata
   type OpaqueText = userdata
   type NumberLike = number | OpaqueNumber
   type TextLike = string | OpaqueText
end
return W
`)
	manifest := M{"types": []any{M{"export": "Frame"}}}
	public := "local type W = require(\"wow\")\nglobal Root: W.Frame\nglobal Make: function(W.NumberLike): W.Frame\n"
	write("wow_env.d.tl", public)
	write("wow_secure_env.d.tl", public+"global SecureCall: function(W.Frame)\n")
	report, err := declarationExports(manifest, directory)
	if err != nil {
		t.Fatal(err)
	}
	expected := M{"public": []any{"Make", "Root"}, "secure": []any{"Make", "Root", "SecureCall"}}
	if !evEqual(report["ambient_globals"], expected) {
		t.Fatalf("unaccounted globals: %v", report)
	}
	for _, tc := range []struct{ body, want string }{{strings.ReplaceAll(public, "W.Frame", "W.UnknownExport"), "unexported declaration"}, {public + "global Root: W.Frame\n", "duplicate public ambient global"}, {"local type W = require(\"wow\")\n", "unaccounted ambient global"}} {
		write("wow_env.d.tl", tc.body)
		_, err := declarationExports(manifest, directory)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Fatalf("expected %s, got %v", tc.want, err)
		}
	}
}
func TestNativeCoverageRejectsMissingContractsAndChangedSecrecy(t *testing.T) {
	root := t.TempDir()
	native := filepath.Join(root, "native")
	if err := os.Mkdir(native, 0755); err != nil {
		t.Fatal(err)
	}
	data := []byte(`local Example = { Name = "Example", Type = "System", Namespace = "C_Example", Functions = {{ Name = "Get", Type = "Function", SecretReturns = true }}, }; APIDocumentation:AddDocumentationTable(Example);`)
	if err := os.WriteFile(filepath.Join(native, "Example.lua"), data, 0644); err != nil {
		t.Fatal(err)
	}
	source := "blizzard:Example.lua"
	metadata := M{"Name": "Get", "Type": "Function", "SecretReturns": true}
	manifest := M{"native_tables": []any{M{"source": source, "tables": []any{}, "predicates": []any{}}}, "symbols": []any{M{"name": "C_Example.Get", "source": source, "metadata": metadata}}, "statistics": M{"native_functions": 1, "native_events": 0}}
	lock := M{"blizzard": M{"files": []any{M{"name": "Example.lua", "sha": blobSHA(data)}}}}
	snapshot := M{"files": []any{M{"path": coverageNativePrefix + "Example.lua"}}}
	p := NewProject(root).WithInputs(map[string]string{"blizzard": native})
	run := func(input M) (contracts []any, err error) {
		defer captureError(&err)
		_, contracts, err = p.coverageNative(context.Background(), input, lock, snapshot)
		return
	}
	contracts, err := run(manifest)
	if err != nil || len(contracts) != 1 || asMap(contracts[0])["name"] != "C_Example.Get" {
		t.Fatalf("correct native contract rejected: %v %v", contracts, err)
	}
	for _, mode := range []string{"omit", "clear-secret"} {
		changed := cloneValue(manifest).(map[string]any)
		if mode == "omit" {
			changed["symbols"] = []any{}
		} else {
			asMap(asMap(asList(changed["symbols"])[0])["metadata"])["SecretReturns"] = false
		}
		if _, err := run(changed); err == nil || !strings.Contains(err.Error(), "native contract metadata differs") {
			t.Fatalf("changed contract claimed native coverage: %s %v", mode, err)
		}
	}
}
func TestCoverageMatchesFrozenCompleteClientReport(t *testing.T) {
	root := captureTestRoot(t)
	snapshot := filepath.Join(root, "build/wow-client/70575")
	if _, err := os.Stat(snapshot); os.IsNotExist(err) {
		t.Skip("full locked client integration requires acquired snapshot")
	}
	baseline := filepath.Join(root, "build/go-migration-baseline/evidence/generated/wow_client_coverage.json")
	expected, err := os.ReadFile(baseline)
	if os.IsNotExist(err) {
		t.Skip("migration-only independent Python oracle not present")
	}
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := readJSON(filepath.Join(root, "evidence/generated/wow_manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	files, err := NewProject(root).Coverage(t.Context(), snapshot, manifest)
	if err != nil {
		t.Fatal(err)
	}
	actual := files["evidence/generated/wow_client_coverage.json"]
	if !bytes.Equal(actual, expected) {
		path := filepath.Join(root, "build/go-migration-coverage-actual.json")
		if err = os.WriteFile(path, actual, 0644); err != nil {
			t.Fatal(err)
		}
		t.Fatalf("Go coverage differs from frozen Python report; actual %s", path)
	}
}
