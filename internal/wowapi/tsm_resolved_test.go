package wowapi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"wow-teal-types/internal/sourcelock"
)

func resolvedFixture(t *testing.T) (Project, resolvedExport, []resolvedRequest, map[string][]tsmStub, map[string]string) {
	t.Helper()
	p := NewProject(t.TempDir())
	write := func(name string, data []byte) {
		t.Helper()
		path := filepath.Join(p.Root, name)
		if e := os.MkdirAll(filepath.Dir(path), 0755); e != nil {
			t.Fatal(e)
		}
		if e := os.WriteFile(path, data, 0644); e != nil {
			t.Fatal(e)
		}
	}
	lock := sourcelock.TSM{Repository: "TradeSkillMaster/wowlua-ls", Commit: strings.Repeat("a", 40), BlobVersion: 44, Blizzard: sourcelock.TSMBlizzard{Repository: "Gethe/wow-ui-source", Commit: strings.Repeat("a", 40), VersionSHA256: strings.Repeat("c", 64)}, Files: []sourcelock.LockedFile{{Name: "precomputed.bin.zst", Size: 8, SHA256: strings.Repeat("b", 64)}}}
	data, e := jsonPretty(lock)
	if e != nil {
		t.Fatal(e)
	}
	write("scripts/wow_api/tsm_sources.json", data)
	for _, name := range []string{"Cargo.toml", "Cargo.lock", "src/main.rs"} {
		write("scripts/wow_api/tsm_resolved_export/"+name, []byte("independent reader source: "+name))
	}
	sources, e := p.exporterSources()
	if e != nil {
		t.Fatal(e)
	}
	candidate := tsmStub{Source: "vendor/example.lua", Line: 1, Security: map[string][]string{}, Parameters: []tsmParameter{}, Returns: []tsmReturn{}}
	files := map[string]string{"vendor/example.lua": "function Read() end\n"}
	inventory := map[string][]tsmStub{"Read": {candidate}}
	requests := []resolvedRequest{{Symbol: "Read", TSMSymbol: "Read"}}
	export := resolvedExport{SchemaVersion: 1, TSMCommit: strings.Repeat("a", 40), BlobVersion: 44, BlobSHA256: strings.Repeat("b", 64), BlobSize: 8, ExporterSources: sources, Bindings: map[string]resolvedBinding{"Read": {Symbol: "Read", TSMSymbol: "Read", Location: resolvedLocation{Path: "vendor/example.lua", Start: 0, End: 19}, Function: &resolvedFunction{}}}}
	return p, export, requests, inventory, files
}
func TestResolvedBindingsRequireExactSourceLocationPolicyAndReaderIdentity(t *testing.T) {
	p, export, requests, inventory, files := resolvedFixture(t)
	bindings, e := p.validateResolved(export, requests, inventory, files)
	if e != nil || len(bindings) != 1 {
		t.Fatalf("valid exact source selection failed: %v", e)
	}
	cases := map[string]func(*resolvedExport){"schema": func(v *resolvedExport) { v.SchemaVersion = 2 }, "commit": func(v *resolvedExport) { v.TSMCommit = strings.Repeat("c", 40) }, "source closure": func(v *resolvedExport) { v.ExporterSources["Cargo.lock"] = "changed" }, "blob identity": func(v *resolvedExport) { v.BlobSHA256 = strings.Repeat("c", 64) }, "missing binding": func(v *resolvedExport) { delete(v.Bindings, "Read") }, "unknown source": func(v *resolvedExport) {
		r := v.Bindings["Read"]
		r.Location.Path = "unknown.lua"
		v.Bindings["Read"] = r
	}, "outside source": func(v *resolvedExport) { r := v.Bindings["Read"]; r.Location.End = 999; v.Bindings["Read"] = r }, "missing function": func(v *resolvedExport) { r := v.Bindings["Read"]; r.Function = nil; v.Bindings["Read"] = r }, "changed policy": func(v *resolvedExport) {
		r := v.Bindings["Read"]
		r.Function.Secret = &resolvedPolicy{Aspects: []string{"Color"}, When: []policyPredicate{}, Preconditions: []policyPrecondition{}}
		v.Bindings["Read"] = r
	}}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			_, changed, _, _, _ := resolvedFixture(t)
			change(&changed)
			if _, e := p.validateResolved(changed, requests, inventory, files); e == nil {
				t.Fatal("altered resolved source/policy facts were accepted")
			}
		})
	}
	if _, e := p.validateResolved(export, append(requests, requests[0]), inventory, files); e == nil {
		t.Fatal("duplicate resolved requests accepted")
	}
}
func TestResolvedBindingDecoderRejectsInvalidLocationsAndPolicyShape(t *testing.T) {
	valid := `{"symbol":"Read","tsm_symbol":"Read","location":{"path":"example.lua","start":0,"end":19},"function":{"overloads":[],"secret":null}}`
	var binding resolvedBinding
	if e := json.Unmarshal([]byte(valid), &binding); e != nil {
		t.Fatal(e)
	}
	for name, pair := range map[string][2]string{"fractional offset": {`"start":0`, `"start":0.5`}, "boolean offset": {`"start":0`, `"start":false`}, "missing offset": {`"start":0,`, ``}, "overflow offset": {`"start":0`, `"start":99999999999999999999999999`}, "missing overload facts": {`"overloads":[],`, ``}, "empty policy": {`"secret":null`, `"secret":{}`}, "false policy": {`"secret":null`, `"secret":false`}} {
		t.Run(name, func(t *testing.T) {
			var got resolvedBinding
			if e := json.Unmarshal([]byte(strings.Replace(valid, pair[0], pair[1], 1)), &got); e == nil {
				t.Fatal("invalid binding was accepted")
			}
		})
	}
}
func TestSelectedPolicyRetainsAbsentFieldsAndNamedScopes(t *testing.T) {
	security := map[string][]string{"secret-when": {"SecretWhenFoo source explanation"}, "secret-args": {"untainted value"}, "secret-aspect": {"Color"}, "secret-unless": {"index 1 2"}, "secret-precondition": {"RequiresAccess ReturnNothing inaccessible"}}
	raw := `{"when":[{"name":"SecretWhenFoo","doc":"source explanation"}],"args":{"policy":"AllowedWhenUntainted","params":["value"]},"aspects":["Color"],"guard":null,"unless":{"param":"index","values":["1","2"]},"clears":null,"restriction_guard":null,"satisfies":null,"preconditions":[{"name":"RequiresAccess","failure":"ReturnNothing","doc":"inaccessible"}]}`
	var expected resolvedPolicy
	if e := json.Unmarshal([]byte(raw), &expected); e != nil {
		t.Fatal(e)
	}
	if _, e := selectedPolicy(security, &expected, "Read"); e != nil {
		t.Fatal(e)
	}
	// Each required field is omitted from literal transport JSON, not represented
	// by a zero Go field; absence and null are distinct source contracts.
	for _, field := range []string{"when", "args", "aspects", "guard", "unless", "clears", "restriction_guard", "satisfies", "preconditions"} {
		t.Run(field, func(t *testing.T) {
			pairs := []string{`"when":[],`, `"args":null,`, `"aspects":[],`, `"guard":null,`, `"unless":null,`, `"clears":null,`, `"restriction_guard":null,`, `"satisfies":null,`, `"preconditions":[]`}
			var kept []string
			for _, pair := range pairs {
				if !strings.HasPrefix(pair, `"`+field+`":`) {
					kept = append(kept, strings.TrimSuffix(pair, ","))
				}
			}
			var got resolvedPolicy
			if e := json.Unmarshal([]byte("{"+strings.Join(kept, ",")+"}"), &got); e == nil {
				t.Fatal("missing policy field accepted")
			}
		})
	}
	expected.Args.Params = []string{"another"}
	if _, e := selectedPolicy(security, &expected, "Read"); e == nil {
		t.Fatal("named argument scope was relaxed")
	}
	if _, e := selectedPolicy(map[string][]string{"secret-new-rule": {"opaque"}}, nil, "Read"); e == nil {
		t.Fatal("unknown security metadata accepted")
	}
}
func TestResolvedSourceIdentityRejectsNonintegerSchemaAndSize(t *testing.T) {
	for _, pair := range [][2]string{{`"blob_version": 44`, `"blob_version": 44.0`}, {`"size": 8`, `"size": 8.5`}, {`"size": 8`, `"size": true`}, {`"commit": "` + strings.Repeat("a", 40) + `"`, `"commit": "moving-reference"`}} {
		p, _, _, _, _ := resolvedFixture(t)
		path := filepath.Join(p.Root, "scripts/wow_api/tsm_sources.json")
		data, e := os.ReadFile(path)
		if e != nil {
			t.Fatal(e)
		}
		changed := strings.Replace(string(data), pair[0], pair[1], 1)
		if changed == string(data) {
			t.Fatal("test mutation did not match transport JSON")
		}
		if e := os.WriteFile(path, []byte(changed), 0644); e != nil {
			t.Fatal(e)
		}
		if _, e := p.resolvedIdentity(); e == nil {
			t.Fatal("invalid reader source identity was accepted")
		}
	}
}
