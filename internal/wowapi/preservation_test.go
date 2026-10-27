package wowapi

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"wow-teal-types/internal/sourcelock"
	"wow-teal-types/internal/wowapi/literal"
)

const preservedFixture = `local Example = {
 Name="Example",Type="System",Environment="All",
 Functions={{Name="Inspect",Type="Function",SecretReturns=false,
 Arguments={{Name="enabled",Type="bool",Default=false}},
 Returns={{Name="value",Type="number",Nilable=false}}},
 {Name="Other",Type="Function"}},
 Tables={{Name="Handler",Type="CallbackType",Arguments={{Name="item",Type="string",ConditionalSecret=true}}}},
 Predicates={{Name="RequireAccess",Type="Predicate",FailureMode="Error"}}
};APIDocumentation:AddDocumentationTable(Example);`

func preservationFixture(t *testing.T, text string) (Project, Manifest) {
	t.Helper()
	root := t.TempDir()
	write := func(name string, data []byte) {
		t.Helper()
		path := filepath.Join(root, name)
		if e := os.MkdirAll(filepath.Dir(path), 0755); e != nil {
			t.Fatal(e)
		}
		if e := os.WriteFile(path, data, 0644); e != nil {
			t.Fatal(e)
		}
	}
	write("references/wow-api/blizzard/Example.lua", []byte(text))
	lock := sourcelock.Sources{"blizzard": {Repository: "fixture/native", Commit: strings.Repeat("a", 40), Path: "", Files: []sourcelock.SourceFile{{Name: "Example.lua", SHA: blobSHA([]byte(text)), Size: len(text)}}}}
	b, e := jsonPretty(lock)
	if e != nil {
		t.Fatal(e)
	}
	write("scripts/wow_api/sources.json", b)
	policy, e := os.ReadFile(filepath.Join(catalogueTestRoot(t), "scripts/wow_api/metadata_fields.json"))
	if e != nil {
		t.Fatal(e)
	}
	write("scripts/wow_api/metadata_fields.json", policy)
	manifest, e := ParseManifest([]byte(`{"symbols":[{"source":"blizzard:Example.lua","name":"Inspect","arguments":[{"Name":"enabled","Type":"bool","Default":false}],"returns":[{"Name":"value","Type":"number","Nilable":false}],"may_return_nothing":false,"restrictions":{},"metadata":{"Name":"Inspect","Type":"Function","SecretReturns":false,"Arguments":[{"Name":"enabled","Type":"bool","Default":false}],"Returns":[{"Name":"value","Type":"number","Nilable":false}]}},{"source":"blizzard:Example.lua","name":"Other","arguments":[],"returns":[],"may_return_nothing":false,"restrictions":{},"metadata":{"Name":"Other","Type":"Function"}}],"native_tables":[{"source":"blizzard:Example.lua","tables":[{"Name":"Handler","Type":"CallbackType","Arguments":[{"Name":"item","Type":"string","ConditionalSecret":true}]}],"predicates":[{"Name":"RequireAccess","Type":"Predicate","FailureMode":"Error"}]}],"statistics":{"native_functions":2,"native_events":0},"representations":[{"source":"fixture","rule":"Projection distinction remains explicit."}]}`))
	if e != nil {
		t.Fatal(e)
	}
	return NewProject(root), manifest
}
func TestCompleteMetadataRetainsFalseAbsenceAndNestedRestrictions(t *testing.T) {
	p, manifest := preservationFixture(t, preservedFixture)
	report, e := p.AuditMetadata(&manifest)
	if e != nil {
		t.Fatal(e)
	}
	if e = RequireKnownMetadata(report); e != nil {
		t.Fatal(e)
	}
	doc := report.Documents[0].Metadata
	if !doc["Environment"].Equal(literal.Text("All")) || !rawArgument(t, doc, "Functions", 0)["SecretReturns"].Equal(literal.Bool(false)) {
		t.Fatal("header or explicit false lost")
	}
	if _, exists := rawArgument(t, doc, "Functions", 1)["SecretReturns"]; exists {
		t.Fatal("absence became an explicit false")
	}
	arg := rawArgument(t, rawArgument(t, doc, "Functions", 0), "Arguments", 0)
	if !arg["Default"].Equal(literal.Bool(false)) {
		t.Fatal("false default lost")
	}
	callback := rawArgument(t, doc, "Tables", 0)
	if !rawArgument(t, callback, "Arguments", 0)["ConditionalSecret"].Equal(literal.Bool(true)) {
		t.Fatal("nested secrecy lost")
	}
	if report.Statistics.NativeFunctions != 2 || report.Statistics.UnknownValues != 0 {
		t.Fatalf("counts: %#v", report.Statistics)
	}
	output := filepath.Join(p.Root, "candidate.json")
	if e = WriteMetadata(output, report); e != nil {
		t.Fatal(e)
	}
	var saved MetadataReport
	data, e := os.ReadFile(output)
	if e == nil {
		e = decodeJSON(data, &saved)
	}
	if e != nil || !metadataEqual(saved, report) {
		t.Fatalf("retained report: %v", e)
	}
}
func TestUnknownMetadataIsRetainedBeforePublicationRefusal(t *testing.T) {
	text := strings.Replace(preservedFixture, "SecretReturns=false,", "SecretReturns=\"conditional\",NewPermission=false,", 1)
	p, manifest := preservationFixture(t, text)
	item := manifest.Symbols[0].Function.Raw
	item["SecretReturns"], item["NewPermission"] = literal.Text("conditional"), literal.Bool(false)
	report, e := p.AuditMetadata(&manifest)
	if e != nil {
		t.Fatal(e)
	}
	if e = RequireKnownMetadata(report); e == nil {
		t.Fatal("unknown permission and changed metadata type accepted")
	}
	unknown := report.UnknownFields
	paths := map[string]bool{}
	for _, v := range unknown {
		paths[v.Path] = true
	}
	if len(paths) != 2 || !paths["$.Functions[0].NewPermission"] || !paths["$.Functions[0].SecretReturns"] {
		t.Fatalf("unknown values lost: %#v", unknown)
	}
	doc := report.Documents[0].Metadata
	function := rawArgument(t, doc, "Functions", 0)
	if !function["SecretReturns"].Equal(literal.Text("conditional")) || !function["NewPermission"].Equal(literal.Bool(false)) {
		t.Fatal("rejected metadata lost from retained evidence")
	}
}
func TestAlteredOrOmittedManifestCannotClaimPreservation(t *testing.T) {
	for _, mode := range []string{"omit-function", "alter-secret", "omit-predicate"} {
		t.Run(mode, func(t *testing.T) {
			p, manifest := preservationFixture(t, preservedFixture)
			switch mode {
			case "omit-function":
				manifest.Symbols = manifest.Symbols[:1]
			case "alter-secret":
				manifest.Symbols[0].Function.Raw["SecretReturns"] = literal.Bool(true)
			case "omit-predicate":
				manifest.NativeTables[0].Predicates = []literal.Value{}
			}
			if _, e := p.AuditMetadata(&manifest); e == nil {
				t.Fatal("omitted or altered metadata accepted")
			}
		})
	}
}
func TestUnsupportedMetadataShapeIsRetainedInPreflight(t *testing.T) {
	p, _ := preservationFixture(t, `local Example={Name="Example",Type="System",Functions=false}`)
	report, e := p.AuditMetadata(nil)
	if e != nil {
		t.Fatal(e)
	}
	if e = RequireKnownMetadata(report); e == nil {
		t.Fatal("unsupported Functions shape accepted")
	}
	doc := report.Documents[0].Metadata
	if !doc["Functions"].Equal(literal.Bool(false)) {
		t.Fatal("unsupported false shape lost")
	}
}
