package wowapi

import (
	"bytes"
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"wow-teal-types/internal/sourcelock"
	"wow-teal-types/internal/wowapi/literal"
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
	raw, e := yamlRaw(value)
	if e != nil {
		t.Fatal(e)
	}
	m, ok := raw.Object()
	if !ok {
		t.Fatal("expected YAML object")
	}
	if !m["flag"].Equal(literal.Bool(false)) || !m["number"].Equal(literal.Int(0)) || !m["text"].Equal(literal.Text("false")) {
		t.Fatalf("YAML scalar distinctions lost: %#v", m)
	}
	if _, exists := m["empty"]; !exists || m["empty"].Kind() != literal.Null {
		t.Fatal("explicit null lost")
	}
	bad, e := safeYAML([]byte("unknown: number"))
	if e != nil {
		t.Fatal(e)
	}
	if _, e := parseSecondaryType(bad); e == nil {
		t.Fatal("unrecognized secondary type silently broadened")
	}
}
func TestTypeFactsKeepSecretLocationOptionalsAndVariadicRelations(t *testing.T) {
	facts := newTypeFacts(map[string]string{}, nil)
	fact := func(text string) typeFact { return facts.fact(text, "tsm", nil) }
	element, container := fact("secret<number>[]?"), fact("secret<number[]>?")
	if element.CarrierType != "number[]" || element.Nilable != true {
		t.Fatalf("array fact: %#v", element)
	}
	if facts.equivalent(element, container) {
		t.Fatal("element secrecy collapsed into container secrecy")
	}
	if string(testEvidenceJSON(t, element.SecretWrappers)) != `[{"path":["union",1,"element"],"type":"number"}]` {
		t.Fatalf("secret path lost: %#v", element.SecretWrappers)
	}
	if string(testEvidenceJSON(t, container.SecretWrappers)) != `[{"path":["union",1],"type":"number[]"}]` {
		t.Fatalf("secret container path lost: %#v", container.SecretWrappers)
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
	if fact("fun(): number | nil").Nilable != false || fact("(fun(): number) | nil").Nilable != true {
		t.Fatal("inner result absence and outer callback absence collapsed")
	}
}
func TestTypeFactsResolveAliasesWithoutInventingInheritance(t *testing.T) {
	native := newCatalogue(NewProject(t.TempDir()))
	native.aliases["Seconds"] = &catalogueAlias{Expression: "number", Source: "native primitive"}
	facts := newTypeFacts(map[string]string{"vendor/aliases.lua": "---@alias Span Seconds\n---@alias Seconds number\n---@class Child : Parent\n"}, native)
	if facts.fact("Span", "tsm", nil).CarrierType != "number" || facts.fact("Seconds", "native", nil).CarrierType != "number" {
		t.Fatal("carrier alias resolution lost")
	}
	if facts.equivalent(facts.fact("Child", "tsm", nil), facts.fact("Parent", "tsm", nil)) {
		t.Fatal("inheritance was treated as type equality")
	}
	if facts.fact("table<string,Span>", "tsm", nil).CarrierType != "table<string,number>" {
		t.Fatal("nested alias not resolved")
	}
	conflicts := newTypeFacts(map[string]string{"vendor/a.lua": "---@alias Broken number\n", "vendor/b.lua": "---@alias Broken string\n"}, nil)
	for _, expression := range []string{"Broken", "fun(x: number", "& number", "`T`", "table<number", "keyof T"} {
		result := conflicts.fact(expression, "tsm", nil)
		if result.Type != "" || result.Unsupported == "" {
			t.Errorf("unsupported fact broadened: %s %#v", expression, result)
		}
	}
	if conflicts.fact("any", "tsm", nil).Type != "any" {
		t.Fatal("source-explicit any is not distinguished from unresolved evidence")
	}
}
func TestScopedPolicyRetainsExactParameterIdentity(t *testing.T) {
	expected := policyScope{Raw: "untainted i j", Policy: "untainted", Scope: "named", Positions: []int{3, 4}, UnknownNames: []string{}}
	if got := scopedPolicy("untainted i j", []string{"list", "separator", "i", "j"}); !evEqual(got, expected) {
		t.Fatalf("policy positions: %#v", got)
	}
	if evEqual(scopedPolicy("untainted i", []string{"list", "i", "j"}), scopedPolicy("untainted j", []string{"list", "i", "j"})) {
		t.Fatal("policy scope identities collapsed")
	}
	if scopedPolicy("maybe", []string{"value"}).Unsupported == "" {
		t.Fatal("unknown policy accepted")
	}
}
func reconciliationFixture() (CrosscheckReport, ReconciliationLedger) {
	facts := literal.Object{"kind": literal.Text("arguments-type"), "position": literal.Int(1), "native": literal.Text("number"), "tsm": literal.Text("string")}
	difference := evDigest(facts)
	report := CrosscheckReport{DeclarationGraphSHA256: "unchanged-graph", SourceLocksSHA256: "unchanged-source", Contracts: []reconciledContract{{symbolContext: symbolContext{Symbol: "Read"}, Fingerprint: "reviewed-fingerprint", RequiresResolution: true, Discrepancies: []string{difference}}}}
	ledger := ReconciliationLedger{DeclarationGraphSHA256: "unchanged-graph", SourceLocksSHA256: "unchanged-source", Policies: map[string]resolutionPolicy{"reviewed": {Reason: "retain precise native argument"}}, Contracts: map[string]contractApproval{"Read": {Fingerprint: "reviewed-fingerprint", Policies: []string{"reviewed"}, Resolutions: []positionResolution{{Policy: "reviewed", DifferenceSHA256: difference, Facts: facts}}}}}
	return report, ledger
}
func TestReconciliationCannotSelfApproveChangedFactsOrGraph(t *testing.T) {
	project := NewProject(t.TempDir())
	report, ledger := reconciliationFixture()
	before, _ := jsonPretty(ledger)
	result, e := project.reconcile(report, ledger, nil)
	if e != nil || !result.Passed {
		t.Fatalf("reviewed contract rejected: %#v %v", result, e)
	}
	for _, mode := range []string{"fingerprint", "facts", "policy", "missing-resolution", "graph", "source", "orphan", "duplicate"} {
		t.Run(mode, func(t *testing.T) {
			r, l := reconciliationFixture()
			switch mode {
			case "fingerprint":
				r.Contracts[0].Fingerprint = "changed-permission"
			case "facts":
				l.Contracts["Read"].Resolutions[0].Facts["position"] = literal.Int(2)
			case "policy":
				delete(l.Policies, "reviewed")
			case "missing-resolution":
				approval := l.Contracts["Read"]
				approval.Resolutions = nil
				l.Contracts["Read"] = approval
			case "graph":
				r.DeclarationGraphSHA256 = "weakened-types"
			case "source":
				r.SourceLocksSHA256 = "changed-source"
			case "orphan":
				l.Contracts["Removed"] = contractApproval{Fingerprint: "unused"}
			case "duplicate":
				r.Contracts = append(r.Contracts, r.Contracts[0])
			}
			result, e := project.reconcile(r, l, nil)
			if e == nil && result.Passed {
				t.Fatal("changed contract or approval facts were accepted")
			}
		})
	}
	after, _ := jsonPretty(ledger)
	if !bytes.Equal(before, after) {
		t.Fatal("reconciliation rewrote original ledger")
	}
}
func equivalenceFixture(t *testing.T) (string, string, ReconciliationLedger, sourceClosure, CrosscheckReport) {
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
	policyData := map[string][]byte{}
	for name := range policies {
		data, e := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
		if e != nil {
			t.Fatal(e)
		}
		policyData[name] = data
	}
	sources := map[string]sourcelock.Sources{}
	tsm := sourcelock.TSM{Commit: strings.Repeat("a", 40), Repository: "fixture/tsm", Files: []sourcelock.LockedFile{}, Blizzard: sourcelock.TSMBlizzard{Repository: "fixture/native", Commit: strings.Repeat("b", 40), VersionSHA256: strings.Repeat("c", 64)}}
	for _, item := range []struct{ root, version string }{{baseline, "12.1.5.70575"}, {target, "13.0.0.80001"}} {
		write(item.root, ".wow-version", []byte(item.version+"\n"))
		for _, directory := range []string{"internal/wowapi", "internal/release", "internal/consumer", "internal/sourcelock", "tools/wow-types", "tools/wow-teal"} {
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
		groups := sourcelock.Sources{}
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
			identity := sourcelock.SourceGroup{Repository: "fixture/" + group, Commit: strings.Repeat("a", 40), Path: group, Files: []sourcelock.SourceFile{{Name: file.name, Size: len(data), SHA: blobSHA(data)}}}
			if item.root == target && slices.Contains([]string{"blizzard", "blizzard_support", "blizzard_version"}, group) {
				identity.Repository = ""
				identity.Commit = ""
				identity.Kind = "casc"
				identity.ClientLockSHA256 = strings.Repeat("c", 64)
				identity.Normalization = "lf-final-newline"
				clientPath := "Interface/AddOns/" + group + "/" + file.name
				if group == "blizzard_version" {
					clientPath = ""
				}
				identity.Files[0].ClientPath = sourcelock.ClientPath(clientPath)
			}
			groups[group] = identity
		}
		sources[item.root] = groups
		writeEvidenceJSON(t, item.root, "scripts/wow_api/sources.json", groups)
		writeEvidenceJSON(t, item.root, "scripts/wow_api/tsm_sources.json", tsm)
	}
	oldClosure := sourceClosure{Native: sources[baseline], TSM: tsm, Secondary: NativeEvidenceSources{NativeObservations: ObservationLock{Build: "12.1.5.70575"}}, Resolved: resolvedSources{BlobSHA256: strings.Repeat("r", 64)}}
	ledger := ReconciliationLedger{SourceLocksSHA256: evDigest(oldClosure), DeclarationGraphSHA256: strings.Repeat("g", 64), Policies: map[string]resolutionPolicy{}, Contracts: map[string]contractApproval{"Read": {Fingerprint: strings.Repeat("f", 64), Policies: []string{}, Resolutions: []positionResolution{}}}}
	report := CrosscheckReport{SourceLocksSHA256: evDigest(oldClosure), DeclarationGraphSHA256: strings.Repeat("g", 64), NativeSecondaryEvidence: NativeEvidence{Sources: oldClosure.Secondary}, TSMResolvedSources: oldClosure.Resolved, Contracts: []reconciledContract{{symbolContext: symbolContext{Symbol: "Read"}, Fingerprint: strings.Repeat("f", 64), Discrepancies: []string{}}}}
	writeEvidenceJSON(t, baseline, "evidence/generated/wow_tsm_crosscheck.json", report)
	writeEvidenceJSON(t, baseline, "scripts/wow_api/tsm_reconciliation.json", ledger)
	writeEvidenceJSON(t, target, "scripts/wow_api/tsm_reconciliation.json", ledger)
	closure := oldClosure
	closure.Native = sources[target]
	currentReport := report
	currentReport.SourceLocksSHA256 = evDigest(closure)
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
	if e != nil || result.BaselineVersion != "12.1.5.70575" || result.TargetVersion != "13.0.0.80001" || result.CheckedNativeFiles != 2 {
		t.Fatalf("equivalence: %#v %v", result, e)
	}
	after, e := os.ReadFile(ledgerPath)
	if e != nil || !bytes.Equal(before, after) {
		t.Fatal("equivalence rewrote original ledger")
	}
}
func TestChangedPermissionBlocksDespiteIdenticalEmittedGraph(t *testing.T) {
	baseline, target, ledger, closure, report := equivalenceFixture(t)
	if report.DeclarationGraphSHA256 != ledger.DeclarationGraphSHA256 {
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
	sources := maps.Clone(closure.Native)
	group := sources["blizzard"]
	group.Files = slices.Clone(group.Files)
	group.Files[0].Size = len(data)
	group.Files[0].SHA = blobSHA(data)
	sources["blizzard"] = group
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
				group := receipt.BaselineSources["blizzard"]
				group.Files = []sourcelock.SourceFile{}
				receipt.BaselineSources["blizzard"] = group
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
				report.DeclarationGraphSHA256 = "changed types"
			case "runtime":
				report.Contracts[0].Fingerprint = "new build lacks matching invocation evidence"
			case "closure":
				closure.Resolved.BlobSHA256 = "different evidence"
				report.SourceLocksSHA256 = evDigest(closure)
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

func testEvidenceJSON[T any](t *testing.T, value T) []byte {
	t.Helper()
	data, err := jsonCompact(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
func writeEvidenceJSON[T any](t *testing.T, root, name string, value T) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, name)
	if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
}

func TestSecondaryBuildUsesCanonicalQuotedOrIntegerIdentity(t *testing.T) {
	for _, raw := range []string{"70529", "'70529'"} {
		got, err := secondaryBuild([]byte("version: 12.1.5\nbuild: " + raw + "\n"))
		if err != nil || got != "12.1.5.70529" {
			t.Fatalf("build %s: got %q, %v", raw, got, err)
		}
	}
	for _, raw := range []string{"null", "70529.5", "'release'", "'070529'", "'+70529'", "'0'", "'999999999999999999999999999999'"} {
		if _, err := secondaryBuild([]byte("version: 12.1.5\nbuild: " + raw + "\n")); err == nil {
			t.Errorf("accepted invalid build %s", raw)
		}
	}
}

func TestSecondaryTypedefSelectionDoesNotInterpretUnreferencedContracts(t *testing.T) {
	data := []byte("typedefs:\n  Known:\n    type: number\n  Unrelated:\n    type: unsupported-carrier\n")
	got, err := parseSecondaryTypedefs(data, map[string]bool{"Known": true})
	if err != nil || len(got) != 1 || got["Known"].Type.Primitive != "number" {
		t.Fatalf("selected typedefs = %#v, %v", got, err)
	}
	if _, err := parseSecondaryTypedefs(data, map[string]bool{"Unrelated": true}); err == nil {
		t.Fatal("interpreted an unsupported selected carrier")
	}
}
