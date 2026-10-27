package wowapi

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func (p Project) loadTSMSources() (M, M, error) {
	lock, e := p.Lock("tsm_sources.json")
	if e != nil {
		return nil, nil, e
	}
	for _, v := range asList(lock["files"]) {
		row := asMap(v)
		name := asString(row["name"])
		if !evidenceSafeRelative(name) {
			return nil, nil, fmt.Errorf("invalid TSM source path")
		}
		b, e := readEvidenceFile(filepath.Join(p.Root, "references/wow-api/tsm", name), 64<<20)
		if e != nil {
			return nil, nil, e
		}
		if len(b) != asInt(row["size"]) || sha256hex(b) != row["sha256"] {
			return nil, nil, fmt.Errorf("TSM source integrity failure: %s", name)
		}
	}
	version, e := readEvidenceFile(filepath.Join(p.Root, "references/wow-api/tsm/blizzard-version.txt"), 1024)
	if e != nil {
		return nil, nil, e
	}
	baseline := asMap(lock["blizzard"])
	if sha256hex(version) != baseline["version_sha256"] || strings.TrimSpace(string(version)) != baseline["build"] {
		return nil, nil, fmt.Errorf("TSM Blizzard baseline version integrity failure")
	}
	native, e := p.Lock("sources.json")
	if e != nil {
		return nil, nil, e
	}
	current := M{}
	for _, v := range asList(asMap(native["blizzard"])["files"]) {
		m := asMap(v)
		current[asString(m["name"])] = m
	}
	matched := 0
	drift := []any{}
	old := map[string]bool{}
	for _, v := range asList(baseline["files"]) {
		row := asMap(v)
		name := asString(row["name"])
		old[name] = true
		local := asMap(current[name])
		if local == nil || local["sha"] != row["sha"] {
			drift = append(drift, name)
		} else {
			b, e := os.ReadFile(p.InputPath("blizzard", name))
			if e != nil {
				return nil, nil, e
			}
			if blobSHA(b) != row["sha"] {
				return nil, nil, fmt.Errorf("modified matched Blizzard source: %s", name)
			}
			matched++
		}
	}
	added := []any{}
	for _, name := range sortedKeys(current) {
		if !old[name] {
			added = append(added, name)
		}
	}
	return lock, M{"matched_files": matched, "changed_or_removed_files": drift, "added_files": added}, nil
}
func (p Project) scriptObjectMapping() (map[string]string, error) {
	data, e := readEvidenceFile(filepath.Join(p.Root, "references/wow-api/tsm/stub_gen_mod.rs"), 8<<20)
	if e != nil {
		return nil, e
	}
	text := string(data)
	_, body, ok := strings.Cut(text, "const SCRIPTOBJECT_CLASS_MAP:")
	if !ok {
		return nil, fmt.Errorf("missing TSM ScriptObject class mapping")
	}
	body, _, ok = strings.Cut(body, "];")
	if !ok {
		return nil, fmt.Errorf("unterminated TSM ScriptObject class mapping")
	}
	r := map[string]string{}
	for _, m := range regexp.MustCompile(`\("([\w]+)", "([\w]+)"\)`).FindAllStringSubmatch(body, -1) {
		r[m[1]] = m[2]
	}
	return r, nil
}
func resolvedRequests(c *catalogue, inventory map[string][]M, mapping map[string]string) ([]M, error) {
	requests := []M{}
	for _, doc := range c.nativeDocs {
		widget := doc.Metadata["Type"] == "ScriptObject"
		owner := asString(doc.Metadata["Namespace"])
		mapped := true
		workspaceOwner := owner
		if widget {
			owner, mapped = mapping[asString(doc.Metadata["Name"])]
			workspaceOwner = scriptName(asString(doc.Metadata["Name"]))
		}
		for _, v := range asList(doc.Metadata["Functions"]) {
			f := asMap(v)
			tsmName := asString(f["Name"])
			if owner != "" {
				tsmName = owner + "." + tsmName
			}
			candidates := inventory[tsmName]
			if widget && !mapped {
				candidates = nil
			}
			if len(candidates) > 0 {
				selection, e := definitionSet(tsmName, candidates, nil)
				if e != nil {
					return nil, e
				}
				if truth(selection["order_uncertain"]) {
					symbol := asString(f["Name"])
					if workspaceOwner != "" {
						symbol = workspaceOwner + "." + symbol
					}
					requests = append(requests, M{"symbol": symbol, "tsm_symbol": tsmName})
				}
			}
		}
	}
	sort.Slice(requests, func(i, j int) bool { return asString(requests[i]["symbol"]) < asString(requests[j]["symbol"]) })
	return requests, nil
}
func sortedStubJSON(rows []M) []any {
	a := append([]M{}, rows...)
	sort.SliceStable(a, func(i, j int) bool {
		if a[i]["source"] != a[j]["source"] {
			return asString(a[i]["source"]) < asString(a[j]["source"])
		}
		return asInt(a[i]["line"]) < asInt(a[j]["line"])
	})
	out := []any{}
	for _, row := range a {
		out = append(out, row)
	}
	return out
}
func sortEvidenceRows(a []any, key, second string) {
	sort.SliceStable(a, func(i, j int) bool {
		l, r := asMap(a[i]), asMap(a[j])
		if l[key] != r[key] {
			return asString(l[key]) < asString(r[key])
		}
		return asString(l[second]) < asString(r[second])
	})
}
func (p Project) crosscheck(ctx context.Context, c *catalogue, manifest M, projection declarationProjection, rendered map[string][]byte) (result M, err error) {
	defer func() {
		if x := recover(); x != nil {
			err = evError(x)
		}
	}()
	lock, matching, e := p.loadTSMSources()
	if e != nil {
		return nil, e
	}
	nativeEvidence, e := p.loadNativeEvidence(c, projection)
	if e != nil {
		return nil, e
	}
	data, e := readEvidenceFile(filepath.Join(p.Root, "references/wow-api/tsm/precomputed-files.bin.zst"), 64<<20)
	if e != nil {
		return nil, e
	}
	files, e := decodeTSMFiles(ctx, data, asInt(lock["blob_version"]), asInt(lock["embedded_files"]))
	if e != nil {
		return nil, e
	}
	inventory, diagnostics, tags := parseTSMStubs(files)
	facts := newTypeFacts(files, c)
	mapping, e := p.scriptObjectMapping()
	if e != nil {
		return nil, e
	}
	requests, e := resolvedRequests(c, inventory, mapping)
	if e != nil {
		return nil, e
	}
	resolved, e := p.loadResolved(ctx, requests, inventory, files)
	if e != nil {
		return nil, e
	}
	predicates := M{}
	for _, doc := range c.nativeDocs {
		if doc.Name == "SecretPredicatesDocumentation.lua" {
			for _, v := range asList(doc.Metadata["Predicates"]) {
				m := asMap(v)
				predicates[asString(m["Name"])] = m["Type"]
			}
		}
	}
	counts := M{"native_functions": 0, "compared_functions": 0, "native_only_functions": 0, "unresolved_functions": 0}
	inc := func(k string, n int) { counts[k] = asInt(counts[k]) + n }
	bint := func(b bool) int {
		if b {
			return 1
		}
		return 0
	}
	differences, effective, contracts, comparisons, nativeOnly := []any{}, []any{}, []any{}, []any{}, []any{}
	carrierSource, e := os.ReadFile(filepath.Join(p.Root, "types/wow/native_values.d.tl"))
	if e != nil {
		return nil, e
	}
	carrierAliases := map[string]string{}
	for _, m := range regexp.MustCompile(`(?m)^\s*type (\w+) = (.+)$`).FindAllStringSubmatch(string(carrierSource), -1) {
		carrierAliases[m[1]] = m[2]
	}
	projectionNotes := map[string]M{}
	for _, v := range c.notes {
		note := asMap(v)
		if name, ok := note["symbol"].(string); ok {
			if projectionNotes[name] == nil {
				projectionNotes[name] = M{}
			}
			b, e := jsonDefault(note)
			if e != nil {
				return nil, e
			}
			projectionNotes[name][string(b)] = note
		}
	}
	for _, document := range c.nativeDocs {
		doc := document.Metadata
		widget := doc["Type"] == "ScriptObject"
		owner := asString(doc["Namespace"])
		mapped := true
		if widget {
			owner, mapped = mapping[asString(doc["Name"])]
		}
		for _, rawNative := range asList(doc["Functions"]) {
			native := asMap(rawNative)
			inc("native_functions", 1)
			workspaceOwner := owner
			if widget {
				workspaceOwner = scriptName(asString(doc["Name"]))
			}
			name := asString(native["Name"])
			if workspaceOwner != "" {
				name = workspaceOwner + "." + name
			}
			tsmName := asString(native["Name"])
			if owner != "" {
				tsmName = owner + "." + tsmName
			}
			context := M{"symbol": name, "native_source": "blizzard:" + document.Name}
			var item M
			if workspaceOwner != "" {
				item = asMap(asMap(asMap(c.defs[workspaceOwner])["methods"])[asString(native["Name"])])
			} else {
				item = asMap(c.globals[asString(native["Name"])])
			}
			if item == nil {
				return nil, fmt.Errorf("missing emitted native contract: %s", name)
			}
			metadata := asMap(evGet(item, "callable_contract", item))
			signatures, e := projection.signatures(name)
			if e != nil {
				return nil, e
			}
			notes := []any{}
			for _, key := range sortedKeys(projectionNotes[name]) {
				notes = append(notes, projectionNotes[name][key])
			}
			emitted := M{"symbol": name, "emitted_signatures": evList(signatures), "workspace_result_secret": c.resultSecret(metadata), "effective_may_return_nothing": c.mayReturnNothing(metadata), "projection_notes": notes}
			candidates := inventory[tsmName]
			if widget && !mapped {
				candidates = nil
			}
			var selection M
			deltas, sourceFacts := []any{}, []any{}
			var absence M
			if len(candidates) == 0 {
				reason := "no archived function declaration"
				if widget && !mapped {
					reason = "TSM generator omits ScriptObjects outside its fixed class map"
				}
				absence = evMerge(context, M{"script_object": doc["Name"], "reason": reason, "evidence": []any{"tsm:blizzard.rs:569", "tsm:stub_gen_mod.rs:316"}, "emitted_contract": emitted, "secondary_crosscheck": asMap(nativeEvidence["methods"])[name]})
				nativeOnly = append(nativeOnly, absence)
				inc("native_only_functions", 1)
				selection = M{"mode": "native-only", "active": []any{}, "discarded": []any{}}
				deltas = append(deltas, M{"kind": "upstream-object-omission", "native": doc["Name"], "tsm": nil})
				if secondary := asMap(absence["secondary_crosscheck"]); secondary != nil {
					inc("secondary_compared_functions", 1)
					for _, difference := range asList(asMap(secondary["signature_comparison"])["differences"]) {
						deltas = append(deltas, M{"kind": "secondary-signature-difference", "native": difference, "tsm": nil})
					}
					for _, difference := range asList(asMap(secondary["receiver_hierarchy"])["differences"]) {
						deltas = append(deltas, M{"kind": "secondary-receiver-hierarchy", "native": difference, "tsm": nil})
					}
				}
			} else {
				binding := asMap(asMap(resolved["bindings"])[name])
				selection, e = definitionSet(tsmName, candidates, binding)
				if e != nil {
					return nil, e
				}
				inc("resolved_primary_bindings", bint(binding != nil))
				inc("compared_functions", 1)
				if widget {
					inc("compared_widgets", 1)
				} else {
					inc("compared_apis", 1)
				}
				inc("overload_sets", bint(selection["mode"] == "typed-namespace-overloads"))
				inc("order_uncertain_symbols", bint(truth(selection["order_uncertain"])))
				type branch struct {
					stub   M
					suffix string
				}
				branches := []branch{}
				for _, v := range asList(selection["active"]) {
					candidate := asMap(v)
					branches = append(branches, branch{candidate, ""})
					primary := false
					for _, p := range asList(selection["possible_primary"]) {
						primary = primary || evEqual(candidate, p)
					}
					if !primary {
						continue
					}
					for index, raw := range evStrings(candidate["overloads"]) {
						suffix := "#overload" + strconv.Itoa(index+1)
						overload, e := explicitOverload(raw, candidate)
						if e != nil {
							deltas = append(deltas, M{"kind": "unsupported-tsm-overload", "variant": stubID(candidate) + suffix, "native": nil, "tsm": raw, "reason": e.Error()})
						} else {
							branches = append(branches, branch{overload, suffix})
						}
					}
				}
				for _, b := range branches {
					inc("explicit_overloads", bint(b.suffix != ""))
					options := asList(selection["primary_security_options"])
					if b.suffix != "" || !strings.Contains(tsmName, ".") || strings.HasPrefix(asString(selection["mode"]), "override") {
						options = []any{M{"security": b.stub["security"], "sources": []any{M{"source": b.stub["source"], "line": b.stub["line"]}}}}
					}
					for _, v := range options {
						option := asMap(v)
						primaryPolicy := asMap(option["security"])
						sourceKey := stubID(b.stub) + b.suffix
						if len(asList(selection["primary_security_options"])) > 1 {
							sourceKey += "#policy" + evDigest(primaryPolicy)
						}
						comparedStub := evMerge(b.stub, M{"security": primaryPolicy})
						raw, e := compareFunction(native, comparedStub, predicates, widget, nil, nil)
						if e != nil {
							return nil, e
						}
						mayReturnNothing := c.mayReturnNothing(metadata)
						normalized, e := compareFunction(native, comparedStub, predicates, widget, facts, &mayReturnNothing)
						if e != nil {
							return nil, e
						}
						if truth(b.stub["source_defect"]) {
							normalized = append(normalized, M{"kind": "upstream-source-defect", "native": native, "tsm": b.stub["source_defect"]})
						}
						projection, e := effectiveFact(c, projection, name, b.stub, widget, carrierAliases)
						if e != nil {
							return nil, e
						}
						effective = append(effective, evMerge(projection, M{"tsm_source": b.stub["source"], "tsm_line": b.stub["line"], "tsm_branch": b.suffix}))
						positions := []any{}
						for _, side := range []struct{ name, field string }{{"argument", "parameters"}, {"return", "returns"}} {
							for i, v := range asList(b.stub[side.field]) {
								value := asMap(v)
								expression := asString(value["type"])
								fact := facts.fact(expression, "tsm", evStrings(b.stub["generics"]))
								if truth(fact["evidence"]) || truth(fact["secret_wrappers"]) || truth(fact["features"]) || fact["unsupported"] != nil {
									positions = append(positions, M{"side": side.name, "position": i + 1, "fact": evMapWithout(fact, "tree")})
								}
								if fact["unsupported"] != nil {
									normalized = append(normalized, M{"kind": "bounded-resolver-limit", "side": side.name, "position": i + 1, "native": nil, "tsm": value["type"], "reason": fact["unsupported"]})
								}
							}
						}
						for _, side := range []struct{ name, field string }{{"argument", "Arguments"}, {"return", "Returns"}} {
							for i, v := range asList(native[side.field]) {
								value := asMap(v)
								nativeFact, e := nativeTypeFact(value, false)
								if e != nil {
									return nil, e
								}
								expression := asString(nativeFact["type"])
								fact := facts.fact(expression, "native", nil)
								if truth(fact["features"]) {
									positions = append(positions, M{"side": side.name, "position": i + 1, "contract_source": "native", "fact": fact, "effective_fact": facts.fact(expression, "effective", nil)})
								}
								if fact["unsupported"] != nil {
									normalized = append(normalized, M{"kind": "bounded-native-resolver-limit", "side": side.name, "position": i + 1, "native": expression, "tsm": nil, "reason": fact["unsupported"]})
								}
							}
						}
						if truth(b.stub["unsupported"]) {
							normalized = append(normalized, M{"kind": "unsupported-tsm-annotation", "native": nil, "tsm": b.stub["unsupported"]})
						}
						for _, v := range asList(projection["opacity_differences"]) {
							opacity := asMap(v)
							if opacity["disposition"] == "review-effective-projection-of-tsm-secret-return" {
								normalized = append(normalized, evMerge(M{"kind": "effective-secret-projection"}, opacity))
							}
						}
						for _, delta := range normalized {
							deltas = append(deltas, evMerge(asMap(delta), M{"variant": sourceKey}))
						}
						sourceFacts = append(sourceFacts, M{"source": b.stub["source"], "line": b.stub["line"], "raw_differences": raw, "branch": b.suffix, "possible_primary_sources": option["sources"], "differences": normalized, "type_facts": positions, "raw_security": b.stub["security"], "effective_function_security": primaryPolicy})
					}
				}
				comparisons = append(comparisons, evMerge(context, M{"tsm_symbol": tsmName, "selection": selection, "candidates": sortedStubJSON(candidates), "variants": sourceFacts}))
				if len(deltas) == 0 {
					inc("matching_functions", 1)
				}
				bounded, activeDefect, sourceDefect := false, false, false
				for _, v := range deltas {
					kind := asString(asMap(v)["kind"])
					bounded = bounded || strings.HasPrefix(kind, "unsupported-") || strings.HasPrefix(kind, "bounded-")
					activeDefect = activeDefect || kind == "upstream-source-defect"
				}
				for _, candidate := range candidates {
					sourceDefect = sourceDefect || truth(candidate["source_defect"])
				}
				inc("bounded_comparison_functions", bint(bounded))
				inc("upstream_source_defect_functions", bint(sourceDefect))
				inc("active_source_defect_functions", bint(activeDefect))
			}
			effective = append(effective, evMerge(emitted, M{"native_audit": true}))
			if len(deltas) > 0 {
				differences = append(differences, evMerge(context, M{"disposition": "requires-exact-reviewed-resolution", "differences": deltas}))
			}
			snapshot := evMerge(context, M{"native": native, "selection": selection, "candidates": sortedStubJSON(candidates), "emitted": emitted, "comparisons": sourceFacts, "differences": deltas})
			if len(candidates) == 0 {
				snapshot["secondary_crosscheck"] = absence["secondary_crosscheck"]
			}
			discrepancies := []string{}
			for _, delta := range deltas {
				discrepancies = append(discrepancies, evDigest(delta))
			}
			sort.Strings(discrepancies)
			coverage := "tsm-source-comparison"
			if len(candidates) == 0 {
				coverage = "native-only-upstream-omission"
				if truth(absence["secondary_crosscheck"]) {
					coverage = "secondary-source-comparison"
				}
			}
			contracts = append(contracts, evMerge(context, M{"fingerprint": evDigest(snapshot), "requires_resolution": len(deltas) > 0, "discrepancies": evList(discrepancies), "coverage": coverage}))
		}
	}
	counts["accounted_functions"] = len(contracts)
	counts["symbols_with_differences"] = len(differences)
	if len(contracts) != asInt(counts["compared_functions"])+asInt(counts["native_only_functions"]) || len(contracts) != asInt(counts["native_functions"]) {
		return nil, fmt.Errorf("TSM native coverage accounting mismatch")
	}
	graph := M{}
	for k, v := range rendered {
		graph[k] = string(v)
	}
	sourceLock, e := p.Lock("sources.json")
	if e != nil {
		return nil, e
	}
	workspaceNative := asMap(sourceLock["blizzard"])
	snapshot := M{"tsm_commit": asMap(lock["blizzard"])["commit"], "tsm_build": asMap(lock["blizzard"])["build"]}
	if workspaceNative["kind"] == "casc" {
		snapshot["workspace_source"] = evMapWithout(workspaceNative, "files")
	} else {
		snapshot["workspace_commit"] = workspaceNative["commit"]
	}
	version, e := os.ReadFile(p.InputPath("blizzard_version", "version.txt"))
	if e != nil {
		return nil, e
	}
	snapshot["workspace_build"] = strings.TrimSpace(string(version))
	snapshot = evMerge(snapshot, matching)
	coverage := asMap(staticEvidenceValue(tsmCoverageDescriptionsJSON))
	coverage["embedded_secret_annotations"] = tags
	coverage["unsupported_annotations"] = diagnostics
	differenceKinds := M{}
	for _, v := range differences {
		for _, delta := range asList(asMap(v)["differences"]) {
			kind := asString(asMap(delta)["kind"])
			differenceKinds[kind] = asInt(differenceKinds[kind]) + 1
		}
	}
	for _, rows := range [][]any{differences, comparisons, nativeOnly, contracts} {
		sortEvidenceRows(rows, "symbol", "")
	}
	sortEvidenceRows(effective, "symbol", "tsm_source")
	secretTypes := evList(sortedKeys(c.secretStructs))
	handles := []string{}
	for name := range c.defs {
		if c.isHandle(name) {
			handles = append(handles, name)
		}
	}
	sort.Strings(handles)
	closure := M{"native": sourceLock, "tsm": lock, "secondary": nativeEvidence["sources"], "resolved": resolved["sources"]}
	result = M{"schema_version": 2, "purpose": "Reviewed source-contract reconciliation; TSM never overrides native types or permissions.", "tsm": M{"repository": lock["repository"], "commit": lock["commit"], "release": lock["release"], "blob_version": lock["blob_version"], "embedded_files": lock["embedded_files"]}, "tsm_sources": lock["files"], "tsm_resolved_sources": resolved["sources"], "native_snapshot": snapshot, "coverage": coverage, "statistics": counts, "difference_kinds": differenceKinds, "differences": differences, "comparisons": comparisons, "native_only": nativeOnly, "native_secondary_evidence": evMapWithout(nativeEvidence, "methods"), "contracts": contracts, "effective_declarations": effective, "effective_policy": M{"secret_data_types": secretTypes, "native_handle_types": evList(handles), "static_carrier_source_sha256": sha256hex(carrierSource)}, "declaration_graph_sha256": evDigest(M{"generated": graph, "native_values": string(carrierSource)}), "source_locks_sha256": evDigest(closure)}
	ledger, e := p.Lock("tsm_reconciliation.json")
	if e != nil {
		return nil, e
	}
	review, e := p.reconcile(result, ledger, closure)
	if e != nil {
		return nil, e
	}
	result["reconciliation"] = review
	counts["unresolved_functions"] = review["unresolved_contracts"]
	counts["reviewed_functions"] = len(contracts) - asInt(review["unresolved_contracts"])
	if !truth(review["passed"]) {
		changed := []any{}
		for _, v := range asList(review["decisions"]) {
			m := asMap(v)
			if m["state"] != "reviewed" {
				changed = append(changed, m["symbol"])
			}
		}
		detail, _ := jsonDefault(M{"contracts": changed, "orphans": review["orphaned_resolutions"], "declaration_graph_matches": review["declaration_graph_matches"], "source_locks_match": review["source_locks_match"]})
		return result, fmt.Errorf("unreviewed TSM reconciliation: %s", detail)
	}
	return result, nil
}
