package wowapi

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEvidenceYAMLRejectsAmbiguousAndUnsafeShapes(t *testing.T) {
	for _, text := range []string{"value: 1\nvalue: 2\n", "shared: &ref [1]\ncopy: *ref\n", "a: !unknown value\n", "value: .inf\n", "value: .nan\n", "first: 1\n---\nsecond: 2\n", "true: value\n", strings.Repeat("[", 50) + "1" + strings.Repeat("]", 50)} {
		t.Run(text, func(t *testing.T) {
			if _, e := safeYAML([]byte(text)); e == nil {
				t.Fatal("unsafe or ambiguous evidence accepted")
			}
		})
	}
	value, e := safeYAML([]byte("flag: false\nnumber: 0\ntext: 'false'\nempty: null\n"))
	if e != nil {
		t.Fatal(e)
	}
	m := asMap(value)
	if m["flag"] != false || m["number"] != int64(0) || m["text"] != "false" {
		t.Fatalf("YAML scalar distinctions lost: %#v", m)
	}
	if _, exists := m["empty"]; !exists || m["empty"] != nil {
		t.Fatal("explicit null lost")
	}
	if _, e := secondaryType(M{"unknown": "number"}); e == nil {
		t.Fatal("unrecognized secondary type silently broadened")
	}
}
func TestTypeFactsKeepSecretLocationOptionalsAndVariadicRelations(t *testing.T) {
	facts := newTypeFacts(map[string]string{}, nil)
	fact := func(text string) M { return facts.fact(text, "tsm", nil) }
	element, container := fact("secret<number>[]?"), fact("secret<number[]>?")
	if element["carrier_type"] != "number[]" || element["nilable"] != true {
		t.Fatalf("array fact: %#v", element)
	}
	if facts.equivalent(element, container) {
		t.Fatal("element secrecy collapsed into container secrecy")
	}
	if !evEqual(element["secret_wrappers"], []any{M{"path": []any{"union", 1, "element"}, "type": "number"}}) {
		t.Fatalf("secret path lost: %#v", element["secret_wrappers"])
	}
	if !evEqual(container["secret_wrappers"], []any{M{"path": []any{"union", 1}, "type": "number[]"}}) {
		t.Fatalf("secret container path lost: %#v", container["secret_wrappers"])
	}
	if facts.equivalent(fact("number?"), fact("number")) {
		t.Fatal("nilability erased")
	}
	if !facts.equivalent(fact("number..."), fact("...number")) || facts.equivalent(fact("number..."), fact("number")) {
		t.Fatal("variadic relationship lost")
	}
	left := fact("fun(value?: secret<number>, ...: string): boolean, secret<string>")
	renamed := fact("fun(other?: secret<number>, ...: string): boolean, secret<string>")
	if !facts.equivalent(left, renamed) {
		t.Fatal("callback parameter labels changed type identity")
	}
	for _, changed := range []string{"fun(value: secret<number>, ...: string): boolean, secret<string>", "fun(value?: secret<number>, tail: string): boolean, secret<string>", "fun(value?: secret<number>, ...: string): secret<string>, boolean", "fun(value?: number, ...: string): boolean, secret<string>", "fun(value?: secret<number>, ...: string): boolean, string"} {
		if facts.equivalent(left, fact(changed)) {
			t.Errorf("changed callback contract considered equivalent: %s", changed)
		}
	}
	if fact("fun(): number | nil")["nilable"] != false || fact("(fun(): number) | nil")["nilable"] != true {
		t.Fatal("inner result absence and outer callback absence collapsed")
	}
}
func TestTypeFactsResolveAliasesWithoutInventingInheritance(t *testing.T) {
	native := newCatalogue(NewProject(t.TempDir()))
	native.aliases["Seconds"] = M{"expression": "number", "source": "native primitive"}
	facts := newTypeFacts(map[string]string{"vendor/aliases.lua": "---@alias Span Seconds\n---@alias Seconds number\n---@class Child : Parent\n"}, native)
	if facts.fact("Span", "tsm", nil)["carrier_type"] != "number" || facts.fact("Seconds", "native", nil)["carrier_type"] != "number" {
		t.Fatal("carrier alias resolution lost")
	}
	if facts.equivalent(facts.fact("Child", "tsm", nil), facts.fact("Parent", "tsm", nil)) {
		t.Fatal("inheritance was treated as type equality")
	}
	if facts.fact("table<string,Span>", "tsm", nil)["carrier_type"] != "table<string,number>" {
		t.Fatal("nested alias not resolved")
	}
	conflicts := newTypeFacts(map[string]string{"vendor/a.lua": "---@alias Broken number\n", "vendor/b.lua": "---@alias Broken string\n"}, nil)
	for _, expression := range []string{"Broken", "fun(x: number", "& number", "`T`", "table<number", "keyof T"} {
		result := conflicts.fact(expression, "tsm", nil)
		if result["type"] != nil || result["unsupported"] == nil {
			t.Errorf("unsupported fact broadened: %s %#v", expression, result)
		}
	}
	if conflicts.fact("any", "tsm", nil)["type"] != "any" {
		t.Fatal("source-explicit any is not distinguished from unresolved evidence")
	}
}
func TestScopedPolicyRetainsExactParameterIdentity(t *testing.T) {
	expected := M{"raw": "untainted i j", "policy": "untainted", "scope": "named", "positions": []any{3, 4}, "unknown_names": []any{}}
	if got := scopedPolicy("untainted i j", []any{"list", "separator", "i", "j"}); !evEqual(got, expected) {
		t.Fatalf("policy positions: %#v", got)
	}
	if evEqual(scopedPolicy("untainted i", []any{"list", "i", "j"}), scopedPolicy("untainted j", []any{"list", "i", "j"})) {
		t.Fatal("policy scope identities collapsed")
	}
	if scopedPolicy("maybe", []any{"value"})["unsupported"] == nil {
		t.Fatal("unknown policy accepted")
	}
}
func reconciliationFixture() (M, M) {
	facts := M{"kind": "arguments-type", "position": 1, "native": "number", "tsm": "string"}
	difference := evDigest(facts)
	report := M{"declaration_graph_sha256": "unchanged-graph", "source_locks_sha256": "unchanged-source", "contracts": []any{M{"symbol": "Read", "fingerprint": "reviewed-fingerprint", "requires_resolution": true, "discrepancies": []any{difference}}}}
	ledger := M{"declaration_graph_sha256": "unchanged-graph", "source_locks_sha256": "unchanged-source", "policies": M{"reviewed": M{"reason": "retain precise native argument"}}, "contracts": M{"Read": M{"fingerprint": "reviewed-fingerprint", "policies": []any{"reviewed"}, "resolutions": []any{M{"policy": "reviewed", "difference_sha256": difference, "facts": facts}}}}}
	return report, ledger
}
func TestReconciliationCannotSelfApproveChangedFactsOrGraph(t *testing.T) {
	project := NewProject(t.TempDir())
	report, ledger := reconciliationFixture()
	before, _ := jsonPretty(ledger)
	result, e := project.reconcile(report, ledger, nil)
	if e != nil || result["passed"] != true {
		t.Fatalf("reviewed contract rejected: %#v %v", result, e)
	}
	for _, mode := range []string{"fingerprint", "facts", "policy", "missing-resolution", "graph", "source", "orphan", "duplicate"} {
		t.Run(mode, func(t *testing.T) {
			r, l := mapCopy(report), mapCopy(ledger)
			switch mode {
			case "fingerprint":
				asMap(asList(r["contracts"])[0])["fingerprint"] = "changed-permission"
			case "facts":
				asMap(asMap(asList(asMap(asMap(l["contracts"])["Read"])["resolutions"])[0])["facts"])["position"] = 2
			case "policy":
				delete(asMap(l["policies"]), "reviewed")
			case "missing-resolution":
				asMap(asMap(l["contracts"])["Read"])["resolutions"] = []any{}
			case "graph":
				r["declaration_graph_sha256"] = "weakened-types"
			case "source":
				r["source_locks_sha256"] = "changed-source"
			case "orphan":
				asMap(l["contracts"])["Removed"] = M{"fingerprint": "unused"}
			case "duplicate":
				r["contracts"] = append(asList(r["contracts"]), asList(r["contracts"])[0])
			}
			result, e := project.reconcile(r, l, nil)
			if e == nil && result["passed"] == true {
				t.Fatal("changed contract or approval facts were accepted")
			}
		})
	}
	after, _ := jsonPretty(ledger)
	if !bytes.Equal(before, after) {
		t.Fatal("reconciliation rewrote original ledger")
	}
}
func equivalenceFixture(t *testing.T) (string, string, M, M, M) {
	t.Helper()
	directory := t.TempDir()
	baseline, target := filepath.Join(directory, "baseline"), filepath.Join(directory, "candidate")
	root := catalogueTestRoot(t)
	policies, e := PolicyHashes(root)
	if e != nil {
		t.Fatal(e)
	}
	write := func(root, name string, data []byte) {
		t.Helper()
		path := filepath.Join(root, filepath.FromSlash(name))
		if e := os.MkdirAll(filepath.Dir(path), 0755); e != nil {
			t.Fatal(e)
		}
		if e := os.WriteFile(path, data, 0644); e != nil {
			t.Fatal(e)
		}
	}
	writeJSON := func(root, name string, value any) {
		t.Helper()
		b, e := jsonPretty(value)
		if e != nil {
			t.Fatal(e)
		}
		write(root, name, b)
	}
	policyData := map[string][]byte{}
	for name := range policies {
		data, e := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
		if e != nil {
			t.Fatal(e)
		}
		policyData[name] = data
	}
	sources := map[string]M{}
	for _, item := range []struct{ root, version string }{{baseline, "12.1.5.70575"}, {target, "13.0.0.80001"}} {
		write(item.root, ".wow-version", []byte(item.version+"\n"))
		for _, directory := range []string{"internal/wowapi", "internal/release", "internal/consumer", "tools/wow-types", "tools/wow-teal"} {
			if e := os.MkdirAll(filepath.Join(item.root, directory), 0755); e != nil {
				t.Fatal(e)
			}
		}
		for name, data := range policyData {
			write(item.root, name, data)
		}
		p := NewProject(item.root)
		files := map[string]struct {
			name string
			data []byte
		}{"blizzard": {"Example.lua", []byte("local Example={Name=\"Example\",Functions={{Name=\"Read\",SecretReturns=false,RequiresAccess=true}}}\n")}, "blizzard_support": {"Access.lua", []byte("local function CanRead(value) return not issecretvalue(value) end\n")}, "blizzard_version": {"version.txt", []byte(item.version + "\n")}, "questie": {"Annotations.lua", []byte("---@param value boolean\nfunction Read(value) end\n")}, "questie_runtime": {"Runtime.lua", []byte("local Library = {}\n")}}
		groups := M{}
		for _, group := range []string{"blizzard", "blizzard_support", "blizzard_version", "questie", "questie_runtime"} {
			file := files[group]
			data := file.data
			if item.root == target && group == "blizzard" {
				data = append(bytes.ReplaceAll(data, []byte("\n"), []byte("\r\n")), []byte("\r\n")...)
			}
			path := p.InputPath(group, file.name)
			relative, e := filepath.Rel(item.root, path)
			if e != nil {
				t.Fatal(e)
			}
			write(item.root, relative, data)
			identity := M{"repository": "fixture/" + group, "commit": strings.Repeat("a", 40)}
			if item.root == target && containsString([]string{"blizzard", "blizzard_support", "blizzard_version"}, group) {
				identity = M{"kind": "casc", "client_lock_sha256": strings.Repeat("c", 64), "normalization": "lf-final-newline"}
			}
			identity["path"] = group
			identity["files"] = []any{M{"name": file.name, "size": len(data), "sha": blobSHA(data)}}
			groups[group] = identity
		}
		sources[item.root] = groups
		writeJSON(item.root, "scripts/wow_api/sources.json", groups)
		writeJSON(item.root, "scripts/wow_api/tsm_sources.json", M{"commit": "tsm-reviewed"})
	}
	oldClosure := M{"native": sources[baseline], "tsm": M{"commit": "tsm-reviewed"}, "secondary": M{"native_capture_build": "12.1.5.70575"}, "resolved": M{"blob_sha256": strings.Repeat("r", 64)}}
	ledger := M{"source_locks_sha256": evDigest(oldClosure), "declaration_graph_sha256": strings.Repeat("g", 64), "policies": M{}, "contracts": M{"Read": M{"fingerprint": strings.Repeat("f", 64), "policies": []any{}, "resolutions": []any{}}}}
	report := M{"source_locks_sha256": evDigest(oldClosure), "declaration_graph_sha256": strings.Repeat("g", 64), "native_secondary_evidence": M{"sources": oldClosure["secondary"]}, "tsm_resolved_sources": oldClosure["resolved"], "contracts": []any{M{"symbol": "Read", "fingerprint": strings.Repeat("f", 64), "requires_resolution": false, "discrepancies": []any{}}}}
	writeJSON(baseline, "evidence/generated/wow_tsm_crosscheck.json", report)
	writeJSON(baseline, "scripts/wow_api/tsm_reconciliation.json", ledger)
	writeJSON(target, "scripts/wow_api/tsm_reconciliation.json", ledger)
	closure := evMerge(oldClosure, M{"native": sources[target]})
	currentReport := evMerge(report, M{"source_locks_sha256": evDigest(closure)})
	return baseline, target, ledger, closure, currentReport
}
func TestByteEquivalentSourcesReuseOnlyUnchangedApproval(t *testing.T) {
	baseline, target, ledger, closure, report := equivalenceFixture(t)
	ledgerPath := filepath.Join(target, "scripts/wow_api/tsm_reconciliation.json")
	before, e := os.ReadFile(ledgerPath)
	if e != nil {
		t.Fatal(e)
	}
	p := NewProject(target)
	if result, e := p.validateEquivalence(report, ledger, closure); e != nil || result != nil {
		t.Fatalf("unrecorded equivalence accepted: %#v %v", result, e)
	}
	receipt, e := CreateEquivalence(baseline, target)
	if e != nil {
		t.Fatal(e)
	}
	b, e := jsonPretty(receipt)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(target, equivalenceReceipt), b, 0644); e != nil {
		t.Fatal(e)
	}
	result, e := p.validateEquivalence(report, ledger, closure)
	if e != nil || result["baseline_version"] != "12.1.5.70575" || result["target_version"] != "13.0.0.80001" || asInt(result["checked_native_files"]) != 2 {
		t.Fatalf("equivalence: %#v %v", result, e)
	}
	after, e := os.ReadFile(ledgerPath)
	if e != nil || !bytes.Equal(before, after) {
		t.Fatal("equivalence rewrote original ledger")
	}
}
func TestChangedPermissionBlocksDespiteIdenticalEmittedGraph(t *testing.T) {
	baseline, target, ledger, closure, report := equivalenceFixture(t)
	if report["declaration_graph_sha256"] != ledger["declaration_graph_sha256"] {
		t.Fatal("fixture graph must remain unchanged")
	}
	p := NewProject(target)
	path := p.InputPath("blizzard", "Example.lua")
	data, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	data = bytes.ReplaceAll(data, []byte("RequiresAccess=true"), []byte("RequiresAccess=false"))
	if e = os.WriteFile(path, data, 0644); e != nil {
		t.Fatal(e)
	}
	sources := mapCopy(asMap(closure["native"]))
	row := asMap(asList(asMap(sources["blizzard"])["files"])[0])
	row["size"], row["sha"] = len(data), blobSHA(data)
	b, e := jsonPretty(sources)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(target, "scripts/wow_api/sources.json"), b, 0644); e != nil {
		t.Fatal(e)
	}
	if _, e = CreateEquivalence(baseline, target); e == nil || !strings.Contains(e.Error(), "native metadata or support bytes changed") {
		t.Fatalf("permission change accepted or wrong failure: %v", e)
	}
	if _, e = os.Stat(filepath.Join(target, equivalenceReceipt)); !os.IsNotExist(e) {
		t.Fatal("failed comparison created an approval receipt")
	}
}
func TestEquivalenceCannotApproveChangedPolicyLedgerOrRuntimeFacts(t *testing.T) {
	for _, mode := range []string{"policy", "ledger", "graph", "runtime", "closure", "receipt", "annotation"} {
		t.Run(mode, func(t *testing.T) {
			baseline, target, ledger, closure, report := equivalenceFixture(t)
			receipt, e := CreateEquivalence(baseline, target)
			if e != nil {
				t.Fatal(e)
			}
			if mode == "receipt" {
				asMap(asMap(receipt["baseline_sources"])["blizzard"])["files"] = []any{}
			}
			b, e := jsonPretty(receipt)
			if e != nil {
				t.Fatal(e)
			}
			if e = os.WriteFile(filepath.Join(target, equivalenceReceipt), b, 0644); e != nil {
				t.Fatal(e)
			}
			p := NewProject(target)
			switch mode {
			case "policy":
				e = os.WriteFile(filepath.Join(target, "tools/check_secrets.lua"), []byte("weakened checker\n"), 0644)
			case "ledger":
				path := filepath.Join(target, "scripts/wow_api/tsm_reconciliation.json")
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				e = os.WriteFile(path, append(data, '\n'), 0644)
			case "graph":
				report["declaration_graph_sha256"] = "changed types"
			case "runtime":
				asMap(asList(report["contracts"])[0])["fingerprint"] = "new build lacks matching invocation evidence"
			case "closure":
				closure["resolved"] = M{"blob_sha256": "different evidence"}
				report["source_locks_sha256"] = evDigest(closure)
			case "annotation":
				e = os.WriteFile(p.InputPath("questie", "Annotations.lua"), []byte("---@param value number\nfunction Read(value) end\n"), 0644)
			}
			if e != nil {
				t.Fatal(e)
			}
			if _, e = p.validateEquivalence(report, ledger, closure); e == nil {
				t.Fatal("changed evidence was self-approved")
			}
		})
	}
}
