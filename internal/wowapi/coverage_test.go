package wowapi

import (
	"bytes"
	"context"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"wow-teal-types/internal/sourcelock"
	"wow-teal-types/internal/wowapi/literal"
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
	manifest := Manifest{Types: []ManifestType{{Export: "Frame"}}}
	public := "local type W = require(\"wow\")\nglobal Root: W.Frame\nglobal Make: function(W.NumberLike): W.Frame\n"
	write("wow_env.d.tl", public)
	write("wow_secure_env.d.tl", public+"global SecureCall: function(W.Frame)\n")
	report, err := declarationExports(manifest, directory)
	if err != nil {
		t.Fatal(err)
	}
	expected := map[string][]string{"public": {"Make", "Root"}, "secure": {"Make", "Root", "SecureCall"}}
	if !evEqual(report.AmbientGlobals, expected) {
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
	data := []byte(`local Example = { Name = "Example", Type = "System", Namespace = "C_Example", Functions = {{ Name = "Get", Type = "Function", SecretReturns = true, SecretReturnsForAspect = { Enum.SecretAspect.Cooldown } }}, }; APIDocumentation:AddDocumentationTable(Example);`)
	if err := os.WriteFile(filepath.Join(native, "Example.lua"), data, 0644); err != nil {
		t.Fatal(err)
	}
	source := "blizzard:Example.lua"
	metadata := literal.Object{"Name": literal.Text("Get"), "Type": literal.Text("Function"), "SecretReturns": literal.Bool(true)}
	aspect, err := literal.Ref("Enum.SecretAspect.Cooldown")
	if err != nil {
		t.Fatal(err)
	}
	metadata["SecretReturnsForAspect"] = literal.List([]literal.Value{aspect})
	function, e := decodeNativeFunction(metadata)
	if e != nil {
		t.Fatal(e)
	}
	manifest := Manifest{NativeTables: []ManifestNativeTables{{Source: source, Tables: []literal.Value{}, Predicates: []literal.Value{}}}, Symbols: []ManifestSymbol{{Name: "C_Example.Get", Source: source, Function: &function}}, Statistics: CatalogueStatistics{NativeFunctions: 1}}
	lock := sourcelock.Sources{"blizzard": {Files: []sourcelock.SourceFile{{Name: "Example.lua", SHA: blobSHA(data)}}}}
	snapshot := sourcelock.Snapshot{Files: []sourcelock.ClientFile{{Path: coverageNativePrefix + "Example.lua"}}}
	p := NewProject(root).WithInputs(map[string]string{"blizzard": native})
	run := func(input Manifest) (contracts []NativeCoverageContract, err error) {
		defer captureError(&err)
		data, err := jsonPretty(input)
		if err != nil {
			return nil, err
		}
		input, err = ParseManifest(data)
		if err != nil {
			return nil, err
		}
		_, contracts, err = p.coverageNative(context.Background(), input, lock, snapshot)
		return
	}
	contracts, err := run(manifest)
	if err != nil || len(contracts) != 1 || contracts[0].Name != "C_Example.Get" {
		t.Fatalf("correct native contract rejected: %v %v", contracts, err)
	}
	for _, mode := range []string{"omit", "clear-secret", "changed-aspect"} {
		changed := manifest
		changed.Symbols = slices.Clone(manifest.Symbols)
		if mode == "omit" {
			changed.Symbols = nil
		} else {
			changedFunction := function
			changedFunction.Raw = maps.Clone(function.Raw)
			if mode == "clear-secret" {
				changedFunction.Raw["SecretReturns"] = literal.Bool(false)
			} else {
				changedFunction.Raw["SecretReturnsForAspect"] = literal.List([]literal.Value{literal.Record(literal.Object{"$ref": literal.Text("Enum.SecretAspect.Health")})})
			}
			changed.Symbols[0].Function = &changedFunction
		}
		if _, err := run(changed); err == nil || !strings.Contains(err.Error(), "native contract metadata differs") {
			t.Fatalf("changed contract claimed native coverage: %s %v", mode, err)
		}
	}
}
func TestCoverageMatchesReviewedCompleteClientReport(t *testing.T) {
	root := captureTestRoot(t)
	snapshot := filepath.Join(root, "build/wow-client/70575")
	if _, err := os.Stat(snapshot); os.IsNotExist(err) {
		t.Skip("full locked client integration requires acquired snapshot")
	}
	baseline := filepath.Join(root, "evidence/generated/wow_client_coverage.json")
	expected, err := os.ReadFile(baseline)
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := ReadManifest(filepath.Join(root, "evidence/generated/wow_manifest.json"))
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
		t.Fatalf("coverage differs from reviewed report; actual %s", path)
	}
}
