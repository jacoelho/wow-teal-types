package wowapi

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const coverageNativePrefix = "Interface/AddOns/Blizzard_APIDocumentationGenerated/"

var coverageValueExports = wordSet("OpaqueNumber OpaqueText NumberLike TextLike")
var coverageExport = regexp.MustCompile(`(?m)^   (?:record|interface|type|enum) ([A-Za-z_]\w*)`)
var coverageImport = regexp.MustCompile(`(?m)^local type ([A-Za-z_]\w*) = require\("wow"\)$`)
var coverageGlobal = regexp.MustCompile(`(?m)^global ([A-Za-z_]\w*): (.+)$`)

func coverageDigest(v any) string {
	data, err := jsonPretty(v)
	if err != nil {
		reject("invalid coverage JSON: %s", err)
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
func coverageUnique(rows []any, key func(M) string, label string) map[string]M {
	result := map[string]M{}
	for _, item := range rows {
		row := asMap(item)
		if row == nil {
			reject("invalid %s", label)
		}
		identity := key(row)
		if _, exists := result[identity]; exists {
			reject("duplicate %s: %s", label, identity)
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
		reject("ambiguous packaged client path: %s", name)
	}
	if len(matches) == 1 {
		return matches[0]
	}
	return ""
}
func declarationExports(manifest M, directory string) (result M, err error) {
	defer captureError(&err)
	data, e := os.ReadFile(filepath.Join(directory, "wow.d.tl"))
	if e != nil {
		return nil, e
	}
	declared := map[string]bool{}
	for _, match := range coverageExport.FindAllStringSubmatch(string(data), -1) {
		if declared[match[1]] {
			reject("duplicate emitted declaration export")
		}
		declared[match[1]] = true
	}
	expected := map[string]bool{}
	for _, item := range asList(manifest["types"]) {
		expected[asString(asMap(item)["export"])] = true
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
		reject("declaration/manifest export mismatch: missing=%v, unexpected=%v", missing[:min(10, len(missing))], unexpected[:min(10, len(unexpected))])
	}
	for name := range coverageValueExports {
		if !declared[name] {
			reject("missing canonical opaque-value exports")
		}
	}
	ambient := M{}
	for filename, label := range map[string]string{"wow_env.d.tl": "public", "wow_secure_env.d.tl": "secure"} {
		body, e := os.ReadFile(filepath.Join(directory, filename))
		if e != nil {
			return nil, e
		}
		text := string(body)
		imported := coverageImport.FindStringSubmatch(text)
		if imported == nil {
			reject("ambient environment does not import the canonical catalogue: %s", filename)
		}
		rows := coverageGlobal.FindAllStringSubmatch(text, -1)
		count := 0
		for line := range strings.SplitSeq(text, "\n") {
			if strings.HasPrefix(line, "global ") {
				count++
			}
		}
		if len(rows) == 0 || len(rows) != count {
			reject("unaccounted ambient global declarations: %s", filename)
		}
		names := map[string]bool{}
		reference := regexp.MustCompile(`\b` + regexp.QuoteMeta(imported[1]) + `\.([A-Za-z_]\w*)`)
		for _, row := range rows {
			if names[row[1]] {
				reject("duplicate %s ambient global: %s", label, row[1])
			}
			names[row[1]] = true
			for _, match := range reference.FindAllStringSubmatch(row[2], -1) {
				if !declared[match[1]] {
					reject("ambient environment references an unexported declaration: %s", filename)
				}
			}
		}
		ambient[label] = evList(sortedKeys(names))
	}
	aliases := []string{}
	for _, name := range sortedKeys(declared) {
		if strings.HasPrefix(name, "Global_") {
			aliases = append(aliases, name)
		}
	}
	return M{"sha256": sha256hex(data), "total_type_names": len(declared), "global_signature_aliases": evList(aliases), "opaque_policy_exports": evList(sortedKeys(coverageValueExports)), "ambient_globals": ambient}, nil
}
func (p Project) coverageNative(ctx context.Context, manifest, sourceLock, snapshot M) ([]any, []any, error) {
	expected := coverageUnique(asList(asMap(sourceLock["blizzard"])["files"]), func(row M) string { return asString(row["name"]) }, "locked native document")
	for name := range expected {
		if !strings.HasSuffix(name, ".lua") {
			delete(expected, name)
		}
	}
	clientFiles := map[string]bool{}
	for _, item := range asList(snapshot["files"]) {
		name := asString(asMap(item)["path"])
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
	tables := coverageUnique(asList(manifest["native_tables"]), func(row M) string { return asString(row["source"]) }, "manifest native document")
	if len(tables) != len(expected) {
		reject("manifest does not account for every locked native Lua document")
	}
	for name := range expected {
		if tables["blizzard:"+name] == nil {
			reject("manifest does not account for every locked native Lua document")
		}
	}
	identity := func(row M) string {
		kind := "function"
		if _, ok := row["event"]; ok {
			kind = "event"
		}
		return kind + "\x00" + asString(row["name"])
	}
	actual := coverageUnique(asList(manifest["symbols"]), identity, "native contract")
	contracts := map[string]M{}
	documents := []any{}
	functions, events := 0, 0
	for _, name := range sortedKeys(expected) {
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}
		data, err := os.ReadFile(p.InputPath("blizzard", name))
		if err != nil {
			return nil, nil, err
		}
		if blobSHA(data) != expected[name]["sha"] {
			return nil, nil, fmt.Errorf("modified native coverage source: %s", name)
		}
		doc, err := ParseLua(data, name)
		if err != nil {
			return nil, nil, err
		}
		source := "blizzard:" + name
		metadata := M{"source": source, "tables": evGet(doc, "Tables", []any{}), "predicates": evGet(doc, "Predicates", []any{})}
		if !evEqual(tables[source], metadata) {
			reject("native table/predicate metadata differs: %s", source)
		}
		owner := asString(doc["Namespace"])
		if doc["Type"] == "ScriptObject" {
			owner = scriptName(asString(doc["Name"]))
		}
		for _, kind := range []string{"function", "event"} {
			key := "Functions"
			if kind == "event" {
				key = "Events"
			}
			for _, item := range asList(doc[key]) {
				row := asMap(item)
				symbol := asString(row["Name"])
				if kind == "function" && owner != "" {
					symbol = owner + "." + symbol
				} else if kind == "event" {
					symbol = asString(evGet(row, "LiteralName", symbol))
				}
				key := kind + "\x00" + symbol
				if contracts[key] != nil {
					reject("duplicate client native contract: %s", symbol)
				}
				recorded := actual[key]
				fact := "metadata"
				if kind == "event" {
					fact = "event"
				}
				if recorded == nil || recorded["source"] != source || !evEqual(recorded[fact], row) {
					reject("native contract metadata differs: %s", symbol)
				}
				contracts[key] = M{"kind": kind, "name": symbol, "source": source, "source_category": "native", "client_path": mapped[name], "metadata_sha256": coverageDigest(row)}
				if kind == "function" {
					functions++
				} else {
					events++
				}
			}
		}
		documents = append(documents, M{"source": source, "client_path": mapped[name], "reviewed_blob_sha": expected[name]["sha"], "function_count": len(asList(doc["Functions"])), "event_count": len(asList(doc["Events"])), "table_count": len(asList(doc["Tables"]))})
	}
	if len(actual) != len(contracts) {
		reject("manifest contains native contracts absent from the selected client documents")
	}
	stats := asMap(manifest["statistics"])
	if functions != asInt(stats["native_functions"]) || events != asInt(stats["native_events"]) {
		reject("native manifest statistics differ from the verified contract inventory")
	}
	rows := []any{}
	for _, key := range sortedKeys(contracts) {
		rows = append(rows, contracts[key])
	}
	return documents, rows, nil
}
func coverageTypes(manifest M, contracts []any) []any {
	rows := coverageUnique(asList(manifest["types"]), func(row M) string { return asString(row["name"]) }, "manifest type name")
	coverageUnique(asList(manifest["types"]), func(row M) string { return asString(row["export"]) }, "manifest type export")
	if len(rows) != asInt(asMap(manifest["statistics"])["type_exports"]) {
		reject("manifest export statistic differs from its type inventory")
	}
	owners := map[string]int{}
	for _, item := range contracts {
		row := asMap(item)
		name := asString(row["name"])
		if index := strings.LastIndexByte(name, '.'); row["kind"] == "function" && index >= 0 {
			owners[name[:index]]++
		}
	}
	var nativeBases func(string, map[string]bool) map[string]bool
	nativeBases = func(name string, seen map[string]bool) map[string]bool {
		if seen[name] || len(seen) > 512 {
			reject("cyclic exported type inheritance: %s", name)
		}
		next := capUnion(seen)
		next[name] = true
		out := map[string]bool{}
		for _, base := range evStrings(rows[name]["bases"]) {
			if owners[base] > 0 {
				out[base] = true
			}
			for parent := range nativeBases(base, next) {
				out[parent] = true
			}
		}
		return out
	}
	result := []any{}
	for _, name := range sortedKeys(rows) {
		row := mapCopy(rows[name])
		row["source_category"] = coverageCategory(asString(row["source"]))
		row["direct_native_contracts"] = owners[name]
		row["native_contract_bases"] = evList(sortedKeys(nativeBases(name, map[string]bool{})))
		result = append(result, row)
	}
	return result
}
func coverageSupplements(manifest M, types []any, sourceLock M) (M, []any) {
	groups := M{}
	libraries := map[string]M{}
	files := coverageUnique(asList(asMap(sourceLock["questie"])["files"]), func(row M) string { return asString(row["name"]) }, "Questie input")
	declarations := coverageUnique(asList(manifest["questie_symbols"]), func(row M) string {
		return asString(row["source"]) + "\x00" + fmt.Sprint(row["line"]) + "\x00" + asString(row["name"])
	}, "supplemental declaration")
	if len(declarations) != asInt(asMap(manifest["statistics"])["questie_declarations"]) {
		reject("supplemental manifest statistics differ from its declaration inventory")
	}
	library := func(name string) M {
		if libraries[name] == nil {
			libraries[name] = M{"types": []any{}, "symbols": []any{}, "files": map[string]bool{}}
		}
		return libraries[name]
	}
	for _, key := range sortedKeys(declarations) {
		row := declarations[key]
		selected := asString(row["selected_source"])
		if !strings.HasPrefix(selected, "questie:") {
			continue
		}
		relative := strings.TrimPrefix(selected, "questie:")
		if files[relative] == nil {
			reject("supplement selects an unlocked annotation input: %s", selected)
		}
		parts := strings.Split(relative, "/")
		category := parts[0]
		groups[category] = append(asList(groups[category]), M{"name": row["name"], "source": row["source"], "line": row["line"], "selected_source": row["selected_source"]})
		if category == "Libraries" {
			if len(parts) < 2 {
				reject("invalid library annotation path")
			}
			item := library(parts[1])
			item["symbols"] = append(asList(item["symbols"]), row["name"])
			item["files"].(map[string]bool)[relative] = true
		}
	}
	for _, item := range types {
		row := asMap(item)
		source := asString(row["source"])
		if !strings.HasPrefix(source, "questie:Libraries/") {
			continue
		}
		relative := strings.TrimPrefix(source, "questie:")
		if files[relative] == nil {
			reject("library type selects an unlocked annotation input: %s", relative)
		}
		parts := strings.Split(relative, "/")
		item := library(parts[1])
		item["types"] = append(asList(item["types"]), row["export"])
		item["files"].(map[string]bool)[relative] = true
	}
	for key, value := range groups {
		rows := asList(value)
		sort.SliceStable(rows, func(i, j int) bool {
			a, b := asMap(rows[i]), asMap(rows[j])
			for _, field := range []string{"name", "source"} {
				if a[field] != b[field] {
					return asString(a[field]) < asString(b[field])
				}
			}
			return asInt(a["line"]) < asInt(b["line"])
		})
		groups[key] = rows
	}
	identity := evMapWithout(asMap(sourceLock["questie"]), "files")
	result := []any{}
	for _, name := range sortedKeys(libraries) {
		item := libraries[name]
		paths := item["files"].(map[string]bool)
		rows := []any{}
		for _, path := range sortedKeys(paths) {
			rows = append(rows, files[path])
		}
		typeNames := evStrings(item["types"])
		symbols := evStrings(item["symbols"])
		sort.Strings(typeNames)
		sort.Strings(symbols)
		result = append(result, M{"name": name, "source_identity": identity, "derive_from_client": false, "types": evList(typeNames), "symbol_declarations": evList(symbols), "files": rows})
	}
	return groups, result
}

// Coverage accounts for every emitted contract after its caller verifies the
// complete snapshot and supplies a Project using that snapshot's projection.
func (p Project) Coverage(ctx context.Context, snapshot string, manifest M) (files map[string][]byte, err error) {
	defer captureError(&err)
	verified, e := readJSON(filepath.Join(snapshot, "snapshot.json"))
	if e != nil {
		return nil, e
	}
	client, e := p.Lock("client_sources.json")
	if e != nil {
		return nil, e
	}
	if !evEqual(verified, client) {
		return nil, fmt.Errorf("coverage snapshot differs from selected source lock")
	}
	sourceLock, e := p.Lock("sources.json")
	if e != nil {
		return nil, e
	}
	cohort := M{}
	for key, value := range sourceLock {
		cohort[key] = evMapWithout(asMap(value), "files")
	}
	if !evEqual(manifest["sources"], cohort) {
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
	categories := M{}
	for _, item := range types {
		category := asString(asMap(item)["source_category"])
		categories[category] = asInt(categories[category]) + 1
	}
	nativeFunctions, nativeEvents := 0, 0
	for _, item := range contracts {
		if asMap(item)["kind"] == "function" {
			nativeFunctions++
		} else {
			nativeEvents++
		}
	}
	retained := 0
	categoryCounts := M{}
	for key, rows := range supplements {
		count := len(asList(rows))
		retained += count
		categoryCounts[key] = count
	}
	thirdTypes, thirdSymbols := 0, 0
	for _, item := range libraries {
		row := asMap(item)
		thirdTypes += len(asList(row["types"]))
		thirdSymbols += len(asList(row["symbol_declarations"]))
	}
	manifestData, e := os.ReadFile(filepath.Join(p.Root, "evidence/generated/wow_manifest.json"))
	if e != nil {
		return nil, e
	}
	recordedManifest, e := readJSON(filepath.Join(p.Root, "evidence/generated/wow_manifest.json"))
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
	frameXML, e := p.Lock("framexml.json")
	if e != nil {
		return nil, e
	}
	report := M{"schema_version": 1, "claim": "Every maintained catalogue export is accounted for; all selected native documentation contracts match the direct client snapshot. Supplemental contracts and representation policy remain explicit. This is not inference of every client Lua/XML declaration.",
		"source_categories":   M{"native": "Recorded native metadata origin; the source document is validated against the direct snapshot.", "annotation-supplement": "Recorded annotation origin, not a claim that the merged type is wholly supplemental. direct_native_contracts and native_contract_bases identify native contributions.", "representation-policy": "Reviewed FrameXML assertions, aliases, opaque identities or compiler representation helpers; not automatically inferred client types."},
		"inputs":              M{"client_lock_sha256": sha256hex(clientBytes), "declaration_manifest_sha256": sha256hex(manifestData), "source_lock_sha256": sha256hex(sourceBytes), "client": evMapWithout(verified, "files"), "source_cohort": cohort},
		"statistics":          M{"client_files": len(asList(verified["files"])), "native_lua_documents": len(documents), "native_functions": nativeFunctions, "native_events": nativeEvents, "catalogue_type_exports": len(types), "type_origins": categories, "retained_supplemental_declarations": retained, "retained_supplemental_categories": categoryCounts, "third_party_type_exports": thirdTypes, "third_party_symbol_declarations": thirdSymbols},
		"declaration_surface": exports, "native_documents": documents, "native_contracts": contracts, "types": types, "remaining_annotation_contracts": supplements, "third_party_libraries": libraries, "reviewed_framexml": frameXML, "representation_decisions": manifest["representations"], "supplemental_method_exclusions": manifest["supplemental_method_exclusions"],
		"limits": []any{"Lua/XML syntax does not generally determine complete parameter, return, nilability, generic or secrecy contracts.", "Third-party libraries, legacy supplements and undocumented C++ APIs cannot all be derived from packaged Blizzard UI files.", "Origin counts are not percentages of automatic type inference; merged interfaces may contain native methods and inherited native capabilities.", "Static opaque identities and conservative secrecy effects remain maintained policy; public fixture evidence does not authorize secret-context operations."}}
	data, e := jsonPretty(report)
	if e != nil {
		return nil, e
	}
	return map[string][]byte{"evidence/generated/wow_client_coverage.json": data}, nil
}
