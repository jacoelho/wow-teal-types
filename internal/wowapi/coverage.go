package wowapi

import (
	"cmp"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"wow-teal-types/internal/sourcelock"
	"wow-teal-types/internal/wowapi/literal"
)

const coverageNativePrefix = "Interface/AddOns/Blizzard_APIDocumentationGenerated/"

var coverageValueExports = wordSet("OpaqueNumber OpaqueText NumberLike TextLike")
var coverageExport = regexp.MustCompile(`(?m)^   (?:record|interface|type|enum) ([A-Za-z_]\w*)`)
var coverageImport = regexp.MustCompile(`(?m)^local type ([A-Za-z_]\w*) = require\("wow"\)$`)
var coverageGlobal = regexp.MustCompile(`(?m)^global ([A-Za-z_]\w*): (.+)$`)

func coverageDigest[T any](v T) string {
	data, err := jsonPretty(v)
	if err != nil {
		reject(fmt.Sprintf("invalid coverage JSON: %s", err))
	}
	return sha256hex(data)
}
func coverageCategory(source string) string {
	if strings.HasPrefix(source, "blizzard:") {
		return "native"
	}
	if strings.HasPrefix(source, "questie:") {
		return "annotation-supplement"
	}
	return "representation-policy"
}
func coverageUnique[T any](rows []T, key func(T) string, label string) map[string]T {
	result := map[string]T{}
	for _, row := range rows {
		identity := key(row)
		if _, exists := result[identity]; exists {
			reject(fmt.Sprintf("duplicate %s: %s", label, identity))
		}
		result[identity] = row
	}
	return result
}
func coverageDifference(a, b map[string]bool) []string {
	out := []string{}
	for key := range a {
		if !b[key] {
			out = append(out, key)
		}
	}
	sort.Strings(out)
	return out
}
func coverageClientPath(name string, files map[string]bool) string {
	exact := "Interface/AddOns/" + name
	if files[exact] {
		return exact
	}
	addon, tail, _ := strings.Cut(name, "/")
	matches := []string{}
	for path := range files {
		if strings.HasPrefix(path, "Interface/AddOns/"+addon+"/") && filepath.Base(path) == filepath.Base(tail) {
			matches = append(matches, path)
		}
	}
	if len(matches) > 1 {
		reject(fmt.Sprintf("ambiguous packaged client path: %s", name))
	}
	if len(matches) == 1 {
		return matches[0]
	}
	return ""
}
func declarationExports(manifest Manifest, directory string) (result DeclarationSurface, err error) {
	defer captureError(&err)
	data, e := os.ReadFile(filepath.Join(directory, "wow.d.tl"))
	if e != nil {
		return DeclarationSurface{}, e
	}
	declared := map[string]bool{}
	for _, match := range coverageExport.FindAllStringSubmatch(string(data), -1) {
		if declared[match[1]] {
			reject("duplicate emitted declaration export")
		}
		declared[match[1]] = true
	}
	expected := map[string]bool{}
	for _, item := range manifest.Types {
		expected[item.Export] = true
	}
	missing := coverageDifference(expected, declared)
	unexpected := []string{}
	for name := range declared {
		if !expected[name] && !coverageValueExports[name] && !strings.HasPrefix(name, "Global_") {
			unexpected = append(unexpected, name)
		}
	}
	sort.Strings(unexpected)
	if len(missing) > 0 || len(unexpected) > 0 {
		reject(fmt.Sprintf("declaration/manifest export mismatch: missing=%v, unexpected=%v", missing[:min(10, len(missing))], unexpected[:min(10, len(unexpected))]))
	}
	for name := range coverageValueExports {
		if !declared[name] {
			reject("missing canonical opaque-value exports")
		}
	}
	ambient := map[string][]string{}
	for filename, label := range map[string]string{"wow_env.d.tl": "public", "wow_secure_env.d.tl": "secure"} {
		body, e := os.ReadFile(filepath.Join(directory, filename))
		if e != nil {
			return DeclarationSurface{}, e
		}
		text := string(body)
		imported := coverageImport.FindStringSubmatch(text)
		if imported == nil {
			reject(fmt.Sprintf("ambient environment does not import the canonical catalogue: %s", filename))
		}
		rows := coverageGlobal.FindAllStringSubmatch(text, -1)
		count := 0
		for line := range strings.SplitSeq(text, "\n") {
			if strings.HasPrefix(line, "global ") {
				count++
			}
		}
		if len(rows) == 0 || len(rows) != count {
			reject(fmt.Sprintf("unaccounted ambient global declarations: %s", filename))
		}
		names := map[string]bool{}
		reference := regexp.MustCompile(`\b` + regexp.QuoteMeta(imported[1]) + `\.([A-Za-z_]\w*)`)
		for _, row := range rows {
			if names[row[1]] {
				reject(fmt.Sprintf("duplicate %s ambient global: %s", label, row[1]))
			}
			names[row[1]] = true
			for _, match := range reference.FindAllStringSubmatch(row[2], -1) {
				if !declared[match[1]] {
					reject(fmt.Sprintf("ambient environment references an unexported declaration: %s", filename))
				}
			}
		}
		ambient[label] = sortedKeys(names)
	}
	aliases := []string{}
	for _, name := range sortedKeys(declared) {
		if strings.HasPrefix(name, "Global_") {
			aliases = append(aliases, name)
		}
	}
	return DeclarationSurface{SHA256: sha256hex(data), TotalTypeNames: len(declared), GlobalSignatureAliases: aliases, OpaquePolicyExports: sortedKeys(coverageValueExports), AmbientGlobals: ambient}, nil
}
func (p Project) coverageNative(ctx context.Context, manifest Manifest, sourceLock sourcelock.Sources, snapshot sourcelock.Snapshot) ([]NativeCoverageDocument, []NativeCoverageContract, error) {
	expected := coverageUnique(sourceLock["blizzard"].Files, func(row sourcelock.SourceFile) string { return row.Name }, "locked native document")
	for name := range expected {
		if !strings.HasSuffix(name, ".lua") {
			delete(expected, name)
		}
	}
	clientFiles := map[string]bool{}
	for _, item := range snapshot.Files {
		name := item.Path
		if clientFiles[name] {
			reject("duplicate client source path")
		}
		clientFiles[name] = true
	}
	mapped := map[string]string{}
	observed := map[string]bool{}
	mappedSet := map[string]bool{}
	missing := []string{}
	for name := range expected {
		path := coverageClientPath("Blizzard_APIDocumentationGenerated/"+name, clientFiles)
		mapped[name] = path
		if path == "" {
			missing = append(missing, name)
		} else {
			mappedSet[path] = true
		}
	}
	for path := range clientFiles {
		if strings.HasPrefix(path, coverageNativePrefix) && strings.HasSuffix(path, ".lua") {
			observed[path] = true
		}
	}
	if len(missing) > 0 || !capEqualSet(mappedSet, observed) {
		reject("native client document inventory mismatch")
	}
	tables := coverageUnique(manifest.NativeTables, func(row ManifestNativeTables) string { return row.Source }, "manifest native document")
	if len(tables) != len(expected) {
		reject("manifest does not account for every locked native Lua document")
	}
	for name := range expected {
		if _, exists := tables["blizzard:"+name]; !exists {
			reject("manifest does not account for every locked native Lua document")
		}
	}
	identity := func(row ManifestSymbol) string {
		kind := "function"
		if row.Event != nil {
			kind = "event"
		}
		return kind + "\x00" + row.Name
	}
	actual := coverageUnique(manifest.Symbols, identity, "native contract")
	contracts := map[string]NativeCoverageContract{}
	documents := []NativeCoverageDocument{}
	functions, events := 0, 0
	for _, name := range sortedKeys(expected) {
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}
		data, err := os.ReadFile(p.InputPath("blizzard", name))
		if err != nil {
			return nil, nil, err
		}
		if blobSHA(data) != expected[name].SHA {
			return nil, nil, fmt.Errorf("modified native coverage source: %s", name)
		}
		raw, err := ParseLua(data, name)
		if err != nil {
			return nil, nil, err
		}
		doc, err := decodeNativeMetadata(raw)
		if err != nil {
			return nil, nil, err
		}
		source := "blizzard:" + name
		tableValues := []literal.Value{}
		for _, table := range doc.Tables {
			tableValues = append(tableValues, literal.Record(table.Raw))
		}
		predicateValues := []literal.Value{}
		for _, predicate := range doc.Predicates {
			predicateValues = append(predicateValues, literal.Record(predicate.Raw))
		}
		metadata := ManifestNativeTables{Source: source, Tables: tableValues, Predicates: predicateValues}
		if !evEqual(tables[source], metadata) {
			reject(fmt.Sprintf("native table/predicate metadata differs: %s", source))
		}
		retain := func(kind, symbol string, raw literal.Object) {
			key := kind + "\x00" + symbol
			if _, exists := contracts[key]; exists {
				reject(fmt.Sprintf("duplicate client native contract: %s", symbol))
			}
			recorded, exists := actual[key]
			var evidence literal.Object
			if kind == "function" && recorded.Function != nil {
				evidence = recorded.Function.Raw
			}
			if kind == "event" && recorded.Event != nil {
				evidence = recorded.Event.Raw
			}
			// Persisted metadata represents symbolic Lua values as JSON objects.
			if !exists || recorded.Source != source || !evEqual(evidence, raw) {
				reject(fmt.Sprintf("native contract metadata differs: %s", symbol))
			}
			contracts[key] = NativeCoverageContract{Kind: kind, Name: symbol, Source: source, SourceCategory: "native", ClientPath: mapped[name], MetadataSHA256: coverageDigest(raw)}
		}
		for _, function := range doc.Functions {
			owner := doc.Namespace
			if doc.Type == "ScriptObject" {
				owner = nativeMethodOwner(doc.Name, function.Name)
			}
			symbol := function.Name
			if owner != "" {
				symbol = owner + "." + symbol
			}
			retain("function", symbol, function.Raw)
			functions++
		}
		for _, event := range doc.Events {
			symbol := event.Name
			if event.LiteralName != "" {
				symbol = event.LiteralName
			}
			retain("event", symbol, event.Raw)
			events++
		}
		documents = append(documents, NativeCoverageDocument{Source: source, ClientPath: mapped[name], ReviewedBlobSHA: expected[name].SHA, FunctionCount: len(doc.Functions), EventCount: len(doc.Events), TableCount: len(doc.Tables)})
	}
	if len(actual) != len(contracts) {
		reject("manifest contains native contracts absent from the selected client documents")
	}
	stats := manifest.Statistics
	if functions != stats.NativeFunctions || events != stats.NativeEvents {
		reject("native manifest statistics differ from the verified contract inventory")
	}
	rows := []NativeCoverageContract{}
	for _, key := range sortedKeys(contracts) {
		rows = append(rows, contracts[key])
	}
	return documents, rows, nil
}
func coverageTypes(manifest Manifest, contracts []NativeCoverageContract) []CoverageType {
	rows := coverageUnique(manifest.Types, func(row ManifestType) string { return row.Name }, "manifest type name")
	coverageUnique(manifest.Types, func(row ManifestType) string { return row.Export }, "manifest type export")
	if len(rows) != manifest.Statistics.TypeExports {
		reject("manifest export statistic differs from its type inventory")
	}
	owners := map[string]int{}
	for _, row := range contracts {
		name := row.Name
		if index := strings.LastIndexByte(name, '.'); row.Kind == "function" && index >= 0 {
			owners[name[:index]]++
		}
	}
	var nativeBases func(string, map[string]bool) map[string]bool
	nativeBases = func(name string, seen map[string]bool) map[string]bool {
		if seen[name] || len(seen) > 512 {
			reject(fmt.Sprintf("cyclic exported type inheritance: %s", name))
		}
		next := capUnion(seen)
		next[name] = true
		out := map[string]bool{}
		for _, base := range rows[name].Bases {
			if owners[base] > 0 {
				out[base] = true
			}
			for parent := range nativeBases(base, next) {
				out[parent] = true
			}
		}
		return out
	}
	result := []CoverageType{}
	for _, name := range sortedKeys(rows) {
		row := rows[name]
		result = append(result, CoverageType{ManifestType: row, SourceCategory: coverageCategory(row.Source), DirectNativeContracts: owners[name], NativeContractBases: sortedKeys(nativeBases(name, map[string]bool{}))})
	}
	return result
}
func coverageSupplements(manifest Manifest, types []CoverageType, sourceLock sourcelock.Sources) (map[string][]CoverageSupplement, []CoverageLibrary) {
	groups := map[string][]CoverageSupplement{}
	type libraryState struct {
		types, symbols []string
		files          map[string]bool
	}
	libraries := map[string]*libraryState{}
	files := coverageUnique(sourceLock["questie"].Files, func(row sourcelock.SourceFile) string { return row.Name }, "Questie input")
	declarations := coverageUnique(manifest.QuestieSymbols, func(row ManifestAnnotationSymbol) string {
		return row.Source + "\x00" + fmt.Sprint(row.Line) + "\x00" + row.Name
	}, "supplemental declaration")
	if len(declarations) != manifest.Statistics.QuestieDeclarations {
		reject("supplemental manifest statistics differ from its declaration inventory")
	}
	library := func(name string) *libraryState {
		if libraries[name] == nil {
			libraries[name] = &libraryState{types: []string{}, symbols: []string{}, files: map[string]bool{}}
		}
		return libraries[name]
	}
	for _, key := range sortedKeys(declarations) {
		row := declarations[key]
		selected := row.SelectedSource
		if !strings.HasPrefix(selected, "questie:") {
			continue
		}
		relative := strings.TrimPrefix(selected, "questie:")
		if _, exists := files[relative]; !exists {
			reject(fmt.Sprintf("supplement selects an unlocked annotation input: %s", selected))
		}
		parts := strings.Split(relative, "/")
		category := parts[0]
		groups[category] = append(groups[category], CoverageSupplement{Name: row.Name, Source: row.Source, Line: row.Line, SelectedSource: row.SelectedSource})
		if category == "Libraries" {
			if len(parts) < 2 {
				reject("invalid library annotation path")
			}
			item := library(parts[1])
			item.symbols = append(item.symbols, row.Name)
			item.files[relative] = true
		}
	}
	for _, row := range types {
		if !strings.HasPrefix(row.Source, "questie:Libraries/") {
			continue
		}
		relative := strings.TrimPrefix(row.Source, "questie:")
		if _, exists := files[relative]; !exists {
			reject(fmt.Sprintf("library type selects an unlocked annotation input: %s", relative))
		}
		parts := strings.Split(relative, "/")
		item := library(parts[1])
		item.types = append(item.types, row.Export)
		item.files[relative] = true
	}
	for _, rows := range groups {
		slices.SortStableFunc(rows, func(a, b CoverageSupplement) int {
			if n := cmp.Compare(a.Name, b.Name); n != 0 {
				return n
			}
			if n := cmp.Compare(a.Source, b.Source); n != 0 {
				return n
			}
			return cmp.Compare(a.Line, b.Line)
		})
	}
	identity := sourceLock["questie"].Identity()
	result := []CoverageLibrary{}
	for _, name := range sortedKeys(libraries) {
		item := libraries[name]
		rows := []sourcelock.SourceFile{}
		for _, path := range sortedKeys(item.files) {
			rows = append(rows, files[path])
		}
		sort.Strings(item.types)
		sort.Strings(item.symbols)
		result = append(result, CoverageLibrary{Name: name, SourceIdentity: identity, Types: item.types, SymbolDeclarations: item.symbols, Files: rows})
	}
	return groups, result
}

// Coverage accounts for every emitted contract after its caller verifies the
// complete snapshot and supplies a Project using that snapshot's projection.
func (p Project) Coverage(ctx context.Context, snapshot string, manifest Manifest) (files map[string][]byte, err error) {
	defer captureError(&err)
	verified, e := sourcelock.ReadSnapshot(filepath.Join(snapshot, "snapshot.json"))
	if e != nil {
		return nil, e
	}
	client, e := sourcelock.ReadSnapshot(filepath.Join(p.Root, "scripts/wow_api/client_sources.json"))
	if e != nil {
		return nil, e
	}
	if !verified.Equal(client) {
		return nil, fmt.Errorf("coverage snapshot differs from selected source lock")
	}
	sourceLock, e := sourcelock.ReadSources(filepath.Join(p.Root, "scripts/wow_api/sources.json"))
	if e != nil {
		return nil, e
	}
	cohort := map[string]sourcelock.SourceIdentity{}
	for key, value := range sourceLock {
		cohort[key] = value.Identity()
	}
	if !evEqual(manifest.Sources, cohort) {
		reject("declaration manifest and reviewed source lock describe different cohorts")
	}
	documents, contracts, e := p.coverageNative(ctx, manifest, sourceLock, verified)
	if e != nil {
		return nil, e
	}
	types := coverageTypes(manifest, contracts)
	exports, e := declarationExports(manifest, filepath.Join(p.Root, "types/wow"))
	if e != nil {
		return nil, e
	}
	supplements, libraries := coverageSupplements(manifest, types, sourceLock)
	categories := map[string]int{}
	for _, item := range types {
		category := item.SourceCategory
		categories[category]++
	}
	nativeFunctions, nativeEvents := 0, 0
	for _, item := range contracts {
		if item.Kind == "function" {
			nativeFunctions++
		} else {
			nativeEvents++
		}
	}
	retained := 0
	categoryCounts := map[string]int{}
	for key, rows := range supplements {
		count := len(rows)
		retained += count
		categoryCounts[key] = count
	}
	thirdTypes, thirdSymbols := 0, 0
	for _, item := range libraries {
		row := item
		thirdTypes += len(row.Types)
		thirdSymbols += len(row.SymbolDeclarations)
	}
	manifestData, e := os.ReadFile(filepath.Join(p.Root, "evidence/generated/wow_manifest.json"))
	if e != nil {
		return nil, e
	}
	recordedManifest, e := ReadManifest(filepath.Join(p.Root, "evidence/generated/wow_manifest.json"))
	if e != nil {
		return nil, e
	}
	if !evEqual(recordedManifest, manifest) {
		reject("coverage manifest differs from generated candidate")
	}
	clientBytes, e := os.ReadFile(filepath.Join(p.Root, "scripts/wow_api/client_sources.json"))
	if e != nil {
		return nil, e
	}
	sourceBytes, e := os.ReadFile(filepath.Join(p.Root, "scripts/wow_api/sources.json"))
	if e != nil {
		return nil, e
	}
	frameXMLData, e := os.ReadFile(filepath.Join(p.Root, "scripts/wow_api/framexml.json"))
	if e != nil {
		return nil, e
	}
	var frameXML frameXMLSupplement
	if e = decodeJSON(frameXMLData, &frameXML); e != nil {
		return nil, e
	}
	report := CoverageReport{SchemaVersion: 1, Claim: "Every maintained catalogue export is accounted for; all selected native documentation contracts match the direct client snapshot. Supplemental contracts and representation policy remain explicit. This is not inference of every client Lua/XML declaration.", SourceCategories: map[string]string{"native": "Recorded native metadata origin; the source document is validated against the direct snapshot.", "annotation-supplement": "Recorded annotation origin, not a claim that the merged type is wholly supplemental. direct_native_contracts and native_contract_bases identify native contributions.", "representation-policy": "Reviewed FrameXML assertions, aliases, opaque identities or compiler representation helpers; not automatically inferred client types."},
		Inputs:             coverageInputs{ClientLockSHA256: sha256hex(clientBytes), DeclarationManifestSHA256: sha256hex(manifestData), SourceLockSHA256: sha256hex(sourceBytes), Client: coverageClient{ClientIdentity: verified.Identity(), Extractor: verified.Extractor, Manifests: verified.Manifests, Schema: verified.Schema}, SourceCohort: cohort},
		Statistics:         coverageStatistics{ClientFiles: len(verified.Files), NativeLuaDocuments: len(documents), NativeFunctions: nativeFunctions, NativeEvents: nativeEvents, CatalogueTypeExports: len(types), TypeOrigins: categories, RetainedSupplementalDeclarations: retained, RetainedSupplementalCategories: categoryCounts, ThirdPartyTypeExports: thirdTypes, ThirdPartySymbolDeclarations: thirdSymbols},
		DeclarationSurface: exports, NativeDocuments: documents, NativeContracts: contracts, Types: types, RemainingAnnotationContracts: supplements, ThirdPartyLibraries: libraries, ReviewedFrameXML: frameXML, RepresentationDecisions: manifest.Representations, SupplementalMethodExclusions: manifest.SupplementalMethodExclusions,
		Limits: []string{"Lua/XML syntax does not generally determine complete parameter, return, nilability, generic or secrecy contracts.", "Third-party libraries, legacy supplements and undocumented C++ APIs cannot all be derived from packaged Blizzard UI files.", "Origin counts are not percentages of automatic type inference; merged interfaces may contain native methods and inherited native capabilities.", "Static opaque identities and conservative secrecy effects remain maintained policy; public fixture evidence does not authorize secret-context operations."}}
	data, e := jsonPretty(report)
	if e != nil {
		return nil, e
	}
	return map[string][]byte{"evidence/generated/wow_client_coverage.json": data}, nil
}
