package wowapi

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func resolvedFixture(t *testing.T) (Project, M, []M, map[string][]M, map[string]string) {
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
	lock := M{"repository": "TradeSkillMaster/wowlua-ls", "commit": strings.Repeat("a", 40), "blob_version": 44, "files": []any{M{"name": "precomputed.bin.zst", "size": 8, "sha256": strings.Repeat("b", 64)}}}
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
	candidate := M{"source": "vendor/example.lua", "line": 1, "security": M{}, "parameters": []any{}, "returns": []any{}}
	files := map[string]string{"vendor/example.lua": "function Read() end\n"}
	inventory := map[string][]M{"Read": {candidate}}
	requests := []M{{"symbol": "Read", "tsm_symbol": "Read"}}
	export := M{"schema_version": 1, "tsm_commit": strings.Repeat("a", 40), "blob_version": 44, "blob_sha256": strings.Repeat("b", 64), "blob_size": 8, "exporter_sources": sources, "bindings": M{"Read": M{"symbol": "Read", "tsm_symbol": "Read", "location": M{"path": "vendor/example.lua", "start": 0, "end": 19}, "function": M{"overloads": []any{}, "secret": nil}}}}
	return p, export, requests, inventory, files
}
func TestResolvedBindingsRequireExactSourceLocationPolicyAndReaderIdentity(t *testing.T) {
	p, export, requests, inventory, files := resolvedFixture(t)
	bindings, e := p.validateResolved(export, requests, inventory, files)
	if e != nil || len(bindings) != 1 {
		t.Fatalf("valid exact source selection failed: %v", e)
	}
	cases := map[string]func(M){"schema": func(v M) { v["schema_version"] = 2 }, "commit": func(v M) { v["tsm_commit"] = strings.Repeat("c", 40) }, "source closure": func(v M) { asMap(v["exporter_sources"])["Cargo.lock"] = "changed" }, "blob identity": func(v M) { v["blob_sha256"] = strings.Repeat("c", 64) }, "missing binding": func(v M) { delete(asMap(v["bindings"]), "Read") }, "unknown source": func(v M) { asMap(asMap(asMap(v["bindings"])["Read"])["location"])["path"] = "unknown.lua" }, "fractional offset": func(v M) { asMap(asMap(asMap(v["bindings"])["Read"])["location"])["start"] = 0.5 }, "boolean offset": func(v M) { asMap(asMap(asMap(v["bindings"])["Read"])["location"])["start"] = false }, "outside source": func(v M) { asMap(asMap(asMap(v["bindings"])["Read"])["location"])["end"] = 999 }, "missing overload facts": func(v M) { delete(asMap(asMap(asMap(v["bindings"])["Read"])["function"]), "overloads") }, "changed policy": func(v M) { asMap(asMap(asMap(v["bindings"])["Read"])["function"])["secret"] = M{} }, "nil versus false": func(v M) { asMap(asMap(asMap(v["bindings"])["Read"])["function"])["secret"] = false }}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			changed := asMap(cloneValue(export))
			change(changed)
			if _, e := p.validateResolved(changed, requests, inventory, files); e == nil {
				t.Fatal("altered resolved source/policy facts were accepted")
			}
		})
	}
	if _, e := p.validateResolved(export, append(requests, requests[0]), inventory, files); e == nil {
		t.Fatal("duplicate resolved requests accepted")
	}
}
func TestSelectedPolicyRetainsAbsentFieldsAndNamedScopes(t *testing.T) {
	security := M{"secret-when": []any{"SecretWhenFoo source explanation"}, "secret-args": []any{"untainted value"}, "secret-aspect": []any{"Color"}, "secret-unless": []any{"index 1 2"}, "secret-precondition": []any{"RequiresAccess ReturnNothing inaccessible"}}
	expected := M{"when": []any{M{"name": "SecretWhenFoo", "doc": "source explanation"}}, "args": M{"policy": "AllowedWhenUntainted", "params": []any{"value"}}, "aspects": []any{"Color"}, "guard": nil, "unless": M{"param": "index", "values": []any{"1", "2"}}, "clears": nil, "restriction_guard": nil, "satisfies": nil, "preconditions": []any{M{"name": "RequiresAccess", "failure": "ReturnNothing", "doc": "inaccessible"}}}
	if _, e := selectedPolicy(security, expected, "Read"); e != nil {
		t.Fatal(e)
	}
	for _, field := range []string{"when", "args", "aspects", "guard", "unless", "clears", "restriction_guard", "satisfies", "preconditions"} {
		t.Run(field, func(t *testing.T) {
			changed := asMap(cloneValue(expected))
			delete(changed, field)
			if _, e := selectedPolicy(security, changed, "Read"); e == nil {
				t.Fatal("absence was treated as equivalent policy")
			}
		})
	}
	changed := asMap(cloneValue(expected))
	asMap(changed["args"])["params"] = []any{"another"}
	if _, e := selectedPolicy(security, changed, "Read"); e == nil {
		t.Fatal("named argument scope was relaxed")
	}
	if _, e := selectedPolicy(M{"secret-new-rule": []any{"opaque"}}, nil, "Read"); e == nil {
		t.Fatal("unknown security metadata accepted")
	}
}
func TestResolvedSourceIdentityRejectsNonintegerSchemaAndSize(t *testing.T) {
	for _, mutation := range []func(M){func(m M) { m["blob_version"] = 44.0 }, func(m M) { asMap(asList(m["files"])[0])["size"] = 8.5 }, func(m M) { asMap(asList(m["files"])[0])["size"] = true }, func(m M) { m["commit"] = "moving-reference" }} {
		p, _, _, _, _ := resolvedFixture(t)
		lock, e := p.Lock("tsm_sources.json")
		if e != nil {
			t.Fatal(e)
		}
		mutation(lock)
		data, e := jsonPretty(lock)
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(filepath.Join(p.Root, "scripts/wow_api/tsm_sources.json"), data, 0644); e != nil {
			t.Fatal(e)
		}
		if _, e = p.resolvedIdentity(); e == nil {
			t.Fatal("invalid reader source identity was accepted")
		}
	}
}
