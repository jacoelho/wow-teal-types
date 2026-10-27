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
	"wow-teal-types/internal/sourcelock"
)

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
func resolvedRequests(c *catalogue, inventory map[string][]tsmStub, mapping map[string]string) ([]resolvedRequest, error) {
	requests := []resolvedRequest{}
	for _, doc := range c.nativeDocs {
		widget := doc.Metadata.Type == "ScriptObject"
		owner := doc.Metadata.Namespace
		mapped := true
		workspaceOwner := owner
		if widget {
			owner, mapped = mapping[doc.Metadata.Name]
		}
		for _, f := range doc.Metadata.Functions {
			if widget {
				workspaceOwner = nativeMethodOwner(doc.Metadata.Name, f.Name)
			}
			tsmName := f.Name
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
				if selection.OrderUncertain {
					symbol := f.Name
					if workspaceOwner != "" {
						symbol = workspaceOwner + "." + symbol
					}
					requests = append(requests, resolvedRequest{Symbol: symbol, TSMSymbol: tsmName})
				}
			}
		}
	}
	sort.Slice(requests, func(i, j int) bool { return requests[i].Symbol < requests[j].Symbol })
	return requests, nil
}
func (p Project) loadTSMSources() (sourcelock.TSM, nativeMatch, error) {
	lock, e := sourcelock.ReadTSM(filepath.Join(p.Root, "scripts/wow_api/tsm_sources.json"))
	if e != nil {
		return lock, nativeMatch{}, e
	}
	fail := func(err error) (sourcelock.TSM, nativeMatch, error) { return lock, nativeMatch{}, err }
	for _, row := range lock.Files {
		if !evidenceSafeRelative(row.Name) {
			return fail(fmt.Errorf("invalid TSM source path"))
		}
		b, e := readEvidenceFile(filepath.Join(p.Root, "references/wow-api/tsm", row.Name), 64<<20)
		if e != nil {
			return fail(e)
		}
		if len(b) != row.Size || sha256hex(b) != row.SHA256 {
			return fail(fmt.Errorf("TSM source integrity failure: %s", row.Name))
		}
	}
	version, e := readEvidenceFile(filepath.Join(p.Root, "references/wow-api/tsm/blizzard-version.txt"), 1024)
	if e != nil {
		return fail(e)
	}
	baseline := lock.Blizzard
	if sha256hex(version) != baseline.VersionSHA256 || strings.TrimSpace(string(version)) != baseline.Build {
		return fail(fmt.Errorf("TSM Blizzard baseline version integrity failure"))
	}
	native, e := sourcelock.ReadSources(filepath.Join(p.Root, "scripts/wow_api/sources.json"))
	if e != nil {
		return fail(e)
	}
	current := map[string]sourcelock.SourceFile{}
	for _, row := range native["blizzard"].Files {
		current[row.Name] = row
	}
	matching := nativeMatch{AddedFiles: []string{}, ChangedOrRemovedFiles: []string{}}
	old := map[string]bool{}
	for _, row := range baseline.Files {
		old[row.Name] = true
		local, ok := current[row.Name]
		if !ok || local.SHA != row.SHA {
			matching.ChangedOrRemovedFiles = append(matching.ChangedOrRemovedFiles, row.Name)
		} else {
			b, e := os.ReadFile(p.InputPath("blizzard", row.Name))
			if e != nil {
				return fail(e)
			}
			if blobSHA(b) != row.SHA {
				return fail(fmt.Errorf("modified matched Blizzard source: %s", row.Name))
			}
			matching.MatchedFiles++
		}
	}
	for _, name := range sortedKeys(current) {
		if !old[name] {
			matching.AddedFiles = append(matching.AddedFiles, name)
		}
	}
	return lock, matching, nil
}
func sortedStubJSON(rows []tsmStub) []tsmStub {
	a := append([]tsmStub{}, rows...)
	sort.SliceStable(a, func(i, j int) bool {
		if a[i].Source != a[j].Source {
			return a[i].Source < a[j].Source
		}
		return a[i].Line < a[j].Line
	})
	return a
}
func (p Project) crosscheck(ctx context.Context, c *catalogue, manifest Manifest, projection declarationProjection, rendered map[string][]byte) (result CrosscheckReport, err error) {
	defer recoverFact(&err)
	lock, matching, e := p.loadTSMSources()
	if e != nil {
		return result, e
	}
	nativeEvidence, e := p.loadNativeEvidence(c, projection)
	if e != nil {
		return result, e
	}
	data, e := readEvidenceFile(filepath.Join(p.Root, "references/wow-api/tsm/precomputed-files.bin.zst"), 64<<20)
	if e != nil {
		return result, e
	}
	files, e := decodeTSMFiles(ctx, data, lock.BlobVersion, lock.EmbeddedFiles)
	if e != nil {
		return result, e
	}
	inventory, diagnostics, tags := parseTSMStubs(files)
	facts := newTypeFacts(files, c)
	mapping, e := p.scriptObjectMapping()
	if e != nil {
		return result, e
	}
	requests, e := resolvedRequests(c, inventory, mapping)
	if e != nil {
		return result, e
	}
	resolved, e := p.loadResolved(ctx, requests, inventory, files)
	if e != nil {
		return result, e
	}
	predicates := map[string]string{}
	for _, doc := range c.nativeDocs {
		if doc.Name == "SecretPredicatesDocumentation.lua" {
			for _, v := range doc.Metadata.Predicates {
				predicates[v.Name] = v.Type
			}
		}
	}
	counts := map[string]int{"native_functions": 0, "compared_functions": 0, "native_only_functions": 0, "unresolved_functions": 0}
	inc := func(key string, yes bool) {
		if yes {
			counts[key]++
		} else {
			counts[key] += 0
		}
	}
	differences, effective, contracts, comparisons, nativeOnly := []symbolDifferences{}, []effectiveDeclaration{}, []reconciledContract{}, []symbolComparison{}, []nativeOnlyContract{}
	carrierSource, e := os.ReadFile(filepath.Join(p.Root, "types/wow/native_values.d.tl"))
	if e != nil {
		return result, e
	}
	carrierAliases := map[string]string{}
	for _, m := range regexp.MustCompile(`(?m)^\s*type (\w+) = (.+)$`).FindAllStringSubmatch(string(carrierSource), -1) {
		carrierAliases[m[1]] = m[2]
	}
	projectionNotes := map[string]map[string]RepresentationNote{}
	for _, note := range c.notes {
		if note.Symbol != nil {
			if projectionNotes[*note.Symbol] == nil {
				projectionNotes[*note.Symbol] = map[string]RepresentationNote{}
			}
			b, e := jsonDefault(note)
			if e != nil {
				return result, e
			}
			projectionNotes[*note.Symbol][string(b)] = note
		}
	}
	for _, document := range c.nativeDocs {
		doc := document.Metadata
		widget := doc.Type == "ScriptObject"
		owner := doc.Namespace
		mapped := true
		if widget {
			owner, mapped = mapping[doc.Name]
		}
		for _, native := range doc.Functions {
			counts["native_functions"]++
			workspaceOwner := owner
			if widget {
				workspaceOwner = nativeMethodOwner(doc.Name, native.Name)
			}
			name := native.Name
			if workspaceOwner != "" {
				name = workspaceOwner + "." + name
			}
			tsmName := native.Name
			if owner != "" {
				tsmName = owner + "." + tsmName
			}
			context := symbolContext{Symbol: name, NativeSource: "blizzard:" + document.Name}
			var item *catalogueFunction
			if workspaceOwner != "" {
				if def := c.defs[workspaceOwner]; def != nil {
					item = def.Methods[native.Name]
				}
			} else {
				item = c.globals[native.Name]
			}
			if item == nil {
				return result, fmt.Errorf("missing emitted native contract: %s", name)
			}
			metadata := item.contract()
			signatures, e := projection.signatures(name)
			if e != nil {
				return result, e
			}
			notes := []RepresentationNote{}
			for _, key := range sortedKeys(projectionNotes[name]) {
				notes = append(notes, projectionNotes[name][key])
			}
			emitted := effectiveDeclaration{Symbol: name, EmittedSignatures: signatures, WorkspaceResultSecret: c.resultSecret(metadata), EffectiveMayReturnNothing: new(c.mayReturnNothing(metadata)), ProjectionNotes: notes}
			candidates := inventory[tsmName]
			if widget && !mapped {
				candidates = nil
			}
			var selection stubDefinitions
			deltas, sourceFacts := []contractDifference{}, []comparisonVariant{}
			var absence nativeOnlyContract
			if len(candidates) == 0 {
				reason := "no archived function declaration"
				if widget && !mapped {
					reason = "TSM generator omits ScriptObjects outside its fixed class map"
				}
				absence = nativeOnlyContract{symbolContext: context, ScriptObject: doc.Name, Reason: reason, Evidence: []string{"tsm:blizzard.rs:569", "tsm:stub_gen_mod.rs:316"}, EmittedContract: emitted, SecondaryCrosscheck: nativeEvidence.Methods[name]}
				nativeOnly = append(nativeOnly, absence)
				counts["native_only_functions"]++
				selection = stubDefinitions{Mode: "native-only", Active: []tsmStub{}, Discarded: []tsmStub{}}
				deltas = append(deltas, difference("upstream-object-omission", textDifference(doc.Name), differenceValue{}, 0))
				if secondary := absence.SecondaryCrosscheck; secondary != nil {
					counts["secondary_compared_functions"]++
					for _, d := range secondary.SignatureComparison.Differences {
						deltas = append(deltas, difference("secondary-signature-difference", differenceValue{secondary: &d}, differenceValue{}, 0))
					}
					for _, d := range secondary.ReceiverHierarchy.Differences {
						deltas = append(deltas, difference("secondary-receiver-hierarchy", differenceValue{hierarchy: &d}, differenceValue{}, 0))
					}
				}
			} else {
				var binding *resolvedBinding
				if v, ok := resolved.Bindings[name]; ok {
					binding = &v
				}
				selection, e = definitionSet(tsmName, candidates, binding)
				if e != nil {
					return result, e
				}
				inc("resolved_primary_bindings", binding != nil)
				counts["compared_functions"]++
				if widget {
					counts["compared_widgets"]++
				} else {
					counts["compared_apis"]++
				}
				inc("overload_sets", selection.Mode == "typed-namespace-overloads")
				inc("order_uncertain_symbols", selection.OrderUncertain)
				type branch struct {
					stub   tsmStub
					suffix string
				}
				branches := []branch{}
				for _, candidate := range selection.Active {
					branches = append(branches, branch{candidate, ""})
					if !containsStub(selection.PossiblePrimary, candidate) {
						continue
					}
					for index, raw := range candidate.Overloads {
						suffix := "#overload" + strconv.Itoa(index+1)
						overload, e := explicitOverload(raw, candidate)
						if e != nil {
							d := difference("unsupported-tsm-overload", differenceValue{}, textDifference(raw), 0)
							d.Variant = stubID(candidate) + suffix
							d.Reason = e.Error()
							deltas = append(deltas, d)
						} else {
							branches = append(branches, branch{overload, suffix})
						}
					}
				}
				for _, b := range branches {
					inc("explicit_overloads", b.suffix != "")
					options := selection.PrimarySecurityOptions
					if b.suffix != "" || !strings.Contains(tsmName, ".") || strings.HasPrefix(selection.Mode, "override") {
						options = []primarySecurityOption{{Security: b.stub.Security, Sources: []stubSource{{Source: b.stub.Source, Line: b.stub.Line}}}}
					}
					for _, option := range options {
						primaryPolicy := option.Security
						sourceKey := stubID(b.stub) + b.suffix
						if len(selection.PrimarySecurityOptions) > 1 {
							sourceKey += "#policy" + evDigest(primaryPolicy)
						}
						comparedStub := b.stub
						comparedStub.Security = primaryPolicy
						raw, e := compareFunction(native, comparedStub, predicates, widget, nil, nil)
						if e != nil {
							return result, e
						}
						mayReturnNothing := c.mayReturnNothing(metadata)
						normalized, e := compareFunction(native, comparedStub, predicates, widget, facts, &mayReturnNothing)
						if e != nil {
							return result, e
						}
						if b.stub.SourceDefect != nil {
							normalized = append(normalized, difference("upstream-source-defect", differenceValue{raw: native.Raw}, differenceValue{defect: b.stub.SourceDefect}, 0))
						}
						projected, e := effectiveFact(c, projection, name, b.stub, widget, carrierAliases)
						if e != nil {
							return result, e
						}
						projected.TSMSource = b.stub.Source
						projected.TSMLine = b.stub.Line
						projected.TSMBranch = &b.suffix
						effective = append(effective, projected)
						positions := []typePosition{}
						for _, side := range []string{"argument", "return"} {
							expressions := []string{}
							if side == "argument" {
								for _, v := range b.stub.Parameters {
									expressions = append(expressions, v.Type)
								}
							} else {
								for _, v := range b.stub.Returns {
									expressions = append(expressions, v.Type)
								}
							}
							for i, expression := range expressions {
								fact := facts.fact(expression, "tsm", b.stub.Generics)
								if len(fact.Evidence) > 0 || len(fact.SecretWrappers) > 0 || len(fact.Features) > 0 || fact.Unsupported != "" {
									positions = append(positions, typePosition{Side: side, Position: i + 1, Fact: positionFact{Value: fact, OmitTree: true}})
								}
								if fact.Unsupported != "" {
									d := difference("bounded-resolver-limit", differenceValue{}, textDifference(expression), i+1)
									d.Side = side
									d.Reason = fact.Unsupported
									normalized = append(normalized, d)
								}
							}
						}
						for _, side := range []struct {
							name   string
							values []NativeParameter
						}{{"argument", native.Arguments}, {"return", native.Returns}} {
							for i, value := range side.values {
								nativeFact, e := nativeTypeFact(value, false)
								if e != nil {
									return result, e
								}
								expression := nativeFact.Type
								fact := facts.fact(expression, "native", nil)
								if len(fact.Features) > 0 {
									effectiveFact := facts.fact(expression, "effective", nil)
									positions = append(positions, typePosition{Side: side.name, Position: i + 1, ContractSource: "native", Fact: positionFact{Value: fact}, EffectiveFact: &effectiveFact})
								}
								if fact.Unsupported != "" {
									d := difference("bounded-native-resolver-limit", textDifference(expression), differenceValue{}, i+1)
									d.Side = side.name
									d.Reason = fact.Unsupported
									normalized = append(normalized, d)
								}
							}
						}
						if len(b.stub.Unsupported) > 0 {
							normalized = append(normalized, difference("unsupported-tsm-annotation", differenceValue{}, differenceValue{unsupported: &b.stub.Unsupported}, 0))
						}
						for _, opacity := range projected.OpacityDifferences {
							if opacity.Disposition == "review-effective-projection-of-tsm-secret-return" {
								normalized = append(normalized, contractDifference{Kind: "effective-secret-projection", Disposition: opacity.Disposition, Position: opacity.Position, TSMSecretWrapper: &opacity.TSMSecretWrapper, WorkspaceOpaqueCarrier: &opacity.WorkspaceOpaqueCarrier})
							}
						}
						for _, delta := range normalized {
							delta.Variant = sourceKey
							deltas = append(deltas, delta)
						}
						sourceFacts = append(sourceFacts, comparisonVariant{Source: b.stub.Source, Line: b.stub.Line, RawDifferences: raw, Branch: b.suffix, PossiblePrimarySources: option.Sources, Differences: normalized, TypeFacts: positions, RawSecurity: b.stub.Security, EffectiveFunctionSecurity: primaryPolicy})
					}
				}
				comparisons = append(comparisons, symbolComparison{symbolContext: context, TSMSymbol: tsmName, Selection: selection, Candidates: sortedStubJSON(candidates), Variants: sourceFacts})
				if len(deltas) == 0 {
					counts["matching_functions"]++
				}
				bounded, activeDefect, sourceDefect := false, false, false
				for _, delta := range deltas {
					bounded = bounded || strings.HasPrefix(delta.Kind, "unsupported-") || strings.HasPrefix(delta.Kind, "bounded-")
					activeDefect = activeDefect || delta.Kind == "upstream-source-defect"
				}
				for _, candidate := range candidates {
					sourceDefect = sourceDefect || candidate.SourceDefect != nil
				}
				inc("bounded_comparison_functions", bounded)
				inc("upstream_source_defect_functions", sourceDefect)
				inc("active_source_defect_functions", activeDefect)
			}
			audit := emitted
			audit.NativeAudit = true
			effective = append(effective, audit)
			if len(deltas) > 0 {
				differences = append(differences, symbolDifferences{symbolContext: context, Disposition: "requires-exact-reviewed-resolution", Differences: deltas})
			}
			snapshot := contractSnapshot{symbolContext: context, Native: native.Raw, Selection: selection, Candidates: sortedStubJSON(candidates), Emitted: emitted, Comparisons: sourceFacts, Differences: deltas}
			if len(candidates) == 0 {
				snapshot.SecondaryCrosscheck = &absence.SecondaryCrosscheck
			}
			discrepancies := []string{}
			for _, delta := range deltas {
				discrepancies = append(discrepancies, evDigest(delta))
			}
			sort.Strings(discrepancies)
			coverage := "tsm-source-comparison"
			if len(candidates) == 0 {
				coverage = "native-only-upstream-omission"
				if absence.SecondaryCrosscheck != nil {
					coverage = "secondary-source-comparison"
				}
			}
			contracts = append(contracts, reconciledContract{symbolContext: context, Fingerprint: evDigest(snapshot), RequiresResolution: len(deltas) > 0, Discrepancies: discrepancies, Coverage: coverage})
		}
	}
	counts["accounted_functions"] = len(contracts)
	counts["symbols_with_differences"] = len(differences)
	if len(contracts) != counts["compared_functions"]+counts["native_only_functions"] || len(contracts) != counts["native_functions"] {
		return result, fmt.Errorf("TSM native coverage accounting mismatch")
	}
	graph := map[string]string{}
	for k, v := range rendered {
		graph[k] = string(v)
	}
	sourceLock, e := sourcelock.ReadSources(filepath.Join(p.Root, "scripts/wow_api/sources.json"))
	if e != nil {
		return result, e
	}
	workspaceNative := sourceLock["blizzard"]
	snapshot := nativeSnapshot{nativeMatch: matching, TSMCommit: lock.Blizzard.Commit, TSMBuild: lock.Blizzard.Build}
	if workspaceNative.Kind == "casc" {
		identity := workspaceNative.Identity()
		snapshot.WorkspaceSource = &identity
	} else {
		snapshot.WorkspaceCommit = workspaceNative.Commit
	}
	version, e := os.ReadFile(p.InputPath("blizzard_version", "version.txt"))
	if e != nil {
		return result, e
	}
	snapshot.WorkspaceBuild = strings.TrimSpace(string(version))
	coverage, e := decodeEvidenceJSON[tsmCoverage]([]byte(tsmCoverageDescriptionsJSON))
	if e != nil {
		return result, e
	}
	coverage.EmbeddedSecretAnnotations = tags
	coverage.UnsupportedAnnotations = diagnostics
	differenceKinds := map[string]int{}
	for _, row := range differences {
		for _, d := range row.Differences {
			differenceKinds[d.Kind]++
		}
	}
	sort.SliceStable(differences, func(i, j int) bool { return differences[i].Symbol < differences[j].Symbol })
	sort.SliceStable(comparisons, func(i, j int) bool { return comparisons[i].Symbol < comparisons[j].Symbol })
	sort.SliceStable(nativeOnly, func(i, j int) bool { return nativeOnly[i].Symbol < nativeOnly[j].Symbol })
	sort.SliceStable(contracts, func(i, j int) bool { return contracts[i].Symbol < contracts[j].Symbol })
	sort.SliceStable(effective, func(i, j int) bool {
		if effective[i].Symbol != effective[j].Symbol {
			return effective[i].Symbol < effective[j].Symbol
		}
		return effective[i].TSMSource < effective[j].TSMSource
	})
	handles := []string{}
	for name := range c.defs {
		if c.isHandle(name) {
			handles = append(handles, name)
		}
	}
	sort.Strings(handles)
	closure := sourceClosure{Native: sourceLock, TSM: lock, Secondary: nativeEvidence.Sources, Resolved: resolved.Sources}
	nativeEvidence.Methods = nil
	result = CrosscheckReport{SchemaVersion: 2, Purpose: "Reviewed source-contract reconciliation; TSM never overrides native types or permissions.", TSM: tsmIdentity{Repository: lock.Repository, Commit: lock.Commit, Release: lock.Release, BlobVersion: lock.BlobVersion, EmbeddedFiles: lock.EmbeddedFiles}, TSMSources: lock.Files, TSMResolvedSources: resolved.Sources, NativeSnapshot: snapshot, Coverage: coverage, Statistics: counts, DifferenceKinds: differenceKinds, Differences: differences, Comparisons: comparisons, NativeOnly: nativeOnly, NativeSecondaryEvidence: nativeEvidence, Contracts: contracts, EffectiveDeclarations: effective, EffectivePolicy: emittedPolicy{SecretDataTypes: sortedKeys(c.secretStructs), NativeHandleTypes: handles, StaticCarrierSourceSHA256: sha256hex(carrierSource)}, DeclarationGraphSHA256: evDigest(struct {
		Generated    map[string]string `json:"generated"`
		NativeValues string            `json:"native_values"`
	}{graph, string(carrierSource)}), SourceLocksSHA256: evDigest(closure)}
	ledgerData, e := readEvidenceFile(filepath.Join(p.Root, "scripts/wow_api/tsm_reconciliation.json"), 64<<20)
	if e != nil {
		return result, e
	}
	ledger, e := decodeEvidenceJSON[ReconciliationLedger](ledgerData)
	if e != nil {
		return result, e
	}
	review, e := p.reconcile(result, ledger, &closure)
	if e != nil {
		return result, e
	}
	result.Reconciliation = review
	counts["unresolved_functions"] = review.UnresolvedContracts
	counts["reviewed_functions"] = len(contracts) - review.UnresolvedContracts
	if !review.Passed {
		changed := []string{}
		for _, decision := range review.Decisions {
			if decision.State != "reviewed" {
				changed = append(changed, decision.Symbol)
			}
		}
		detail, _ := jsonDefault(struct {
			Contracts               []string `json:"contracts"`
			DeclarationGraphMatches bool     `json:"declaration_graph_matches"`
			Orphans                 []string `json:"orphans"`
			SourceLocksMatch        bool     `json:"source_locks_match"`
		}{changed, review.DeclarationGraphMatches, review.OrphanedResolutions, review.SourceLocksMatch})
		return result, fmt.Errorf("unreviewed TSM reconciliation: %s", detail)
	}
	return result, nil
}
