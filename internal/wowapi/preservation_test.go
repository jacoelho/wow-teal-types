package wowapi

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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

func preservationFixture(t *testing.T, text string) (Project, M) {
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
	lock := M{"blizzard": M{"repository": "fixture/native", "commit": strings.Repeat("a", 40), "files": []any{M{"name": "Example.lua", "sha": blobSHA([]byte(text)), "size": len(text)}}}}
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
	inspect := M{"Name": "Inspect", "Type": "Function", "SecretReturns": false, "Arguments": []any{M{"Name": "enabled", "Type": "bool", "Default": false}}, "Returns": []any{M{"Name": "value", "Type": "number", "Nilable": false}}}
	other := M{"Name": "Other", "Type": "Function"}
	callback := M{"Name": "Handler", "Type": "CallbackType", "Arguments": []any{M{"Name": "item", "Type": "string", "ConditionalSecret": true}}}
	predicate := M{"Name": "RequireAccess", "Type": "Predicate", "FailureMode": "Error"}
	manifest := M{"symbols": []any{M{"source": "blizzard:Example.lua", "name": "Inspect", "metadata": inspect}, M{"source": "blizzard:Example.lua", "name": "Other", "metadata": other}}, "native_tables": []any{M{"source": "blizzard:Example.lua", "tables": []any{callback}, "predicates": []any{predicate}}}, "statistics": M{"native_functions": 2, "native_events": 0}, "representations": []any{M{"source": "fixture", "note": "Projection distinction remains explicit."}}}
	return NewProject(root), manifest
}
func TestCompleteMetadataRetainsFalseAbsenceAndNestedRestrictions(t *testing.T) {
	p, manifest := preservationFixture(t, preservedFixture)
	report, e := p.AuditMetadata(manifest)
	if e != nil {
		t.Fatal(e)
	}
	if e = RequireKnownMetadata(report); e != nil {
		t.Fatal(e)
	}
	doc := asMap(asMap(asList(report["documents"])[0])["metadata"])
	functions := asList(doc["Functions"])
	if doc["Environment"] != "All" || asMap(functions[0])["SecretReturns"] != false {
		t.Fatal("header or explicit false lost")
	}
	if _, exists := asMap(functions[1])["SecretReturns"]; exists {
		t.Fatal("absence became an explicit false")
	}
	arg := asMap(asList(asMap(functions[0])["Arguments"])[0])
	if arg["Default"] != false {
		t.Fatal("false default lost")
	}
	callback := asMap(asList(doc["Tables"])[0])
	if asMap(asList(callback["Arguments"])[0])["ConditionalSecret"] != true {
		t.Fatal("nested secrecy lost")
	}
	if asMap(report["statistics"])["native_functions"] != 2 || asMap(report["statistics"])["unknown_values"] != 0 {
		t.Fatalf("counts: %#v", report["statistics"])
	}
	output := filepath.Join(p.Root, "candidate.json")
	if e = WriteMetadata(output, report); e != nil {
		t.Fatal(e)
	}
	saved, e := readJSON(output)
	if e != nil || !metadataEqual(saved, report) {
		t.Fatalf("retained report: %v", e)
	}
}
func TestUnknownMetadataIsRetainedBeforePublicationRefusal(t *testing.T) {
	text := strings.Replace(preservedFixture, "SecretReturns=false,", "SecretReturns=\"conditional\",NewPermission=false,", 1)
	p, manifest := preservationFixture(t, text)
	item := asMap(asMap(asList(manifest["symbols"])[0])["metadata"])
	item["SecretReturns"], item["NewPermission"] = "conditional", false
	report, e := p.AuditMetadata(manifest)
	if e != nil {
		t.Fatal(e)
	}
	if e = RequireKnownMetadata(report); e == nil {
		t.Fatal("unknown permission and changed metadata type accepted")
	}
	unknown := asList(report["unknown_fields"])
	paths := map[string]bool{}
	for _, v := range unknown {
		paths[asString(asMap(v)["path"])] = true
	}
	if len(paths) != 2 || !paths["$.Functions[0].NewPermission"] || !paths["$.Functions[0].SecretReturns"] {
		t.Fatalf("unknown values lost: %#v", unknown)
	}
	doc := asMap(asMap(asList(report["documents"])[0])["metadata"])
	function := asMap(asList(doc["Functions"])[0])
	if function["SecretReturns"] != "conditional" || function["NewPermission"] != false {
		t.Fatal("rejected metadata lost from retained evidence")
	}
}
func TestAlteredOrOmittedManifestCannotClaimPreservation(t *testing.T) {
	for _, mode := range []string{"omit-function", "alter-secret", "omit-predicate"} {
		t.Run(mode, func(t *testing.T) {
			p, manifest := preservationFixture(t, preservedFixture)
			switch mode {
			case "omit-function":
				manifest["symbols"] = asList(manifest["symbols"])[:1]
			case "alter-secret":
				asMap(asMap(asList(manifest["symbols"])[0])["metadata"])["SecretReturns"] = true
			case "omit-predicate":
				asMap(asList(manifest["native_tables"])[0])["predicates"] = []any{}
			}
			if _, e := p.AuditMetadata(manifest); e == nil {
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
	doc := asMap(asMap(asList(report["documents"])[0])["metadata"])
	if doc["Functions"] != false {
		t.Fatal("unsupported false shape lost")
	}
}
