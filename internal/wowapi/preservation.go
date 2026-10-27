package wowapi

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var metadataComponent = regexp.MustCompile(`^(?:[A-Za-z_][A-Za-z_0-9]*|\$(?:ref|expression))$`)

func metadataValueType(value any) string {
	switch v := value.(type) {
	case nil:
		return "null"
	case bool:
		return "boolean"
	case int, int64:
		return "integer"
	case float64:
		return "number"
	case json.Number:
		if strings.ContainsAny(string(v), ".eE") {
			return "number"
		}
		return "integer"
	case string:
		return "string"
	case []any:
		return "array"
	case M:
		return "object"
	}
	reject("unsupported native metadata value %T", value)
	return ""
}
func metadataWalk(value any, schemaPath, path string, visit func(string, string, any)) {
	visit(schemaPath, path, value)
	switch v := value.(type) {
	case M:
		for _, name := range sortedKeys(v) {
			component := "." + name
			if !metadataComponent.MatchString(name) {
				component = "[" + requireStringJSON(name) + "]"
			}
			metadataWalk(v[name], schemaPath+component, path+component, visit)
		}
	case []any:
		for i, child := range v {
			metadataWalk(child, schemaPath+"[]", path+"["+strconv.Itoa(i)+"]", visit)
		}
	}
}

type metadataIdentity struct{ source, kind, name string }

func metadataEqual(a, b any) bool {
	aa, e := jsonCompact(a)
	if e != nil {
		return false
	}
	bb, e := jsonCompact(b)
	return e == nil && string(aa) == string(bb)
}

// AuditMetadata retains every literal and verifies the manifest before permitting publication.
func (p Project) AuditMetadata(manifest M) (report M, err error) {
	defer captureError(&err)
	sourceLock := mustMap(p.Lock("sources.json"))
	policyPath := filepath.Join(p.Root, "scripts/wow_api/metadata_fields.json")
	policyData := mustFile(policyPath)
	policy := mustMap(readJSON(policyPath))
	shapes := asMap(policy["value_types"])
	validTypes := wordSet("null boolean integer number string array object")
	if asInt(policy["schema_version"]) != 1 || len(shapes) == 0 || shapes["$"] == nil {
		reject("invalid native metadata field policy")
	}
	for _, v := range shapes {
		types, ok := v.([]any)
		if !ok || len(types) == 0 {
			reject("invalid native metadata field policy")
		}
		for _, kind := range types {
			if !validTypes[asString(kind)] {
				reject("invalid native metadata field policy")
			}
		}
	}
	expected := map[metadataIdentity]M{}
	documents, unknown := []any{}, []any{}
	tables := M{}
	counts := map[string]int{"function": 0, "event": 0}
	rows := []M{}
	for _, v := range asList(asMap(sourceLock["blizzard"])["files"]) {
		row := asMap(v)
		if strings.HasSuffix(asString(row["name"]), ".lua") {
			rows = append(rows, row)
		}
	}
	if len(rows) == 0 {
		reject("native metadata input inventory is empty")
	}
	sort.Slice(rows, func(i, j int) bool { return asString(rows[i]["name"]) < asString(rows[j]["name"]) })
	names := map[string]bool{}
	for _, row := range rows {
		name := asString(row["name"])
		if filepath.IsAbs(name) || containsString(strings.Split(filepath.ToSlash(name), "/"), "..") || names[name] {
			reject("invalid or duplicate native metadata input: %s", name)
		}
		names[name] = true
		path := p.InputPath("blizzard", name)
		data := mustFile(path)
		if blobSHA(data) != asString(row["sha"]) || len(data) != asInt(row["size"]) {
			reject("native metadata source identity mismatch: %s", name)
		}
		metadata, e := ParseLua(data, path)
		if e != nil {
			return nil, e
		}
		source := "blizzard:" + name
		documents = append(documents, M{"source": source, "reviewed_blob_sha": row["sha"], "source_sha256": sha256hex(data), "size": len(data), "metadata": metadata})
		metadataWalk(metadata, "$", "$", func(schemaPath, location string, value any) {
			kind := metadataValueType(value)
			if !containsString(stringsOf(shapes[schemaPath]), kind) {
				unknown = append(unknown, M{"source": source, "path": location, "schema_path": schemaPath, "value_type": kind, "expected_types": getOr(shapes, schemaPath, []any{}), "value": value})
			}
		})
		for _, entry := range []struct{ key, kind string }{{"Functions", "function"}, {"Events", "event"}} {
			items, ok := getOr(metadata, entry.key, []any{}).([]any)
			if !ok {
				continue
			}
			for i, v := range items {
				item := asMap(v)
				if item != nil {
					for _, required := range []string{"Name", "Type"} {
						if _, ok := item[required]; !ok {
							unknown = append(unknown, M{"source": source, "path": fmt.Sprintf("$.%s[%d].%s", entry.key, i, required), "schema_path": "$." + entry.key + "[]." + required, "value_type": "absent", "expected_types": []any{"string"}, "value": nil})
						}
					}
				}
				name, ok := item["Name"].(string)
				if !ok {
					continue
				}
				identity := metadataIdentity{source, entry.kind, name}
				if expected[identity] != nil {
					reject("duplicate native metadata contract: %v", identity)
				}
				expected[identity] = item
				counts[entry.kind]++
			}
		}
		tables[source] = M{"source": source, "tables": getOr(metadata, "Tables", []any{}), "predicates": getOr(metadata, "Predicates", []any{})}
	}
	if manifest != nil {
		observed := map[metadataIdentity]M{}
		for _, v := range asList(manifest["symbols"]) {
			item := asMap(v)
			kind, key := "function", "metadata"
			if _, ok := item["event"]; ok {
				kind, key = "event", "event"
			}
			metadata := asMap(item[key])
			identity := metadataIdentity{asString(item["source"]), kind, asString(metadata["Name"])}
			if observed[identity] != nil {
				reject("duplicate manifest native metadata contract: %v", identity)
			}
			observed[identity] = metadata
		}
		if len(observed) != len(expected) {
			reject("manifest does not preserve native contracts: missing or extra contracts")
		}
		for key, value := range expected {
			if !metadataEqual(observed[key], value) {
				reject("manifest does not preserve native contracts: altered %s.%s", key.source, key.name)
			}
		}
		actualTables := M{}
		for _, v := range asList(manifest["native_tables"]) {
			item := asMap(v)
			source := asString(item["source"])
			if actualTables[source] != nil {
				reject("duplicate manifest native metadata document: %s", source)
			}
			actualTables[source] = item
		}
		if !metadataEqual(actualTables, tables) {
			reject("manifest does not preserve complete native tables and predicates")
		}
		statistics := asMap(manifest["statistics"])
		if counts["function"] != asInt(statistics["native_functions"]) || counts["event"] != asInt(statistics["native_events"]) {
			reject("native metadata contract counts differ from the manifest")
		}
	}
	sources := mapCopy(asMap(sourceLock["blizzard"]))
	delete(sources, "files")
	notes := []any{}
	if manifest != nil {
		notes = asList(manifest["representations"])
	}
	report = M{"schema_version": 1, "claim": "All selected native literal metadata is retained here. Original source bytes are separately bound by the client/source locks. Teal declarations alone do not express every source distinction.", "sources": sources, "metadata_policy_sha256": sha256hex(policyData), "statistics": M{"native_documents": len(documents), "native_functions": counts["function"], "native_events": counts["event"], "unknown_values": len(unknown)}, "documents": documents, "manifest_accounted": manifest != nil, "representation_notes": notes, "projection_limits": []any{"Comments, lexical spelling, line endings and executable registration are retained in the locked source bytes, rather than in the parsed metadata values.", "Known metadata shapes establish parser coverage, not automatic approval of changed permissions or secret-value behavior.", "Teal nilability, repeated tuples, opaque markers and source-untyped values retain the representation limits recorded by the declaration manifest."}, "unknown_fields": unknown, "passed": len(unknown) == 0}
	return
}
func RequireKnownMetadata(report M) error {
	if len(asList(report["unknown_fields"])) == 0 {
		return nil
	}
	values := []any{}
	for _, v := range asList(report["unknown_fields"])[:min(10, len(asList(report["unknown_fields"])))] {
		item := asMap(v)
		values = append(values, M{"source": item["source"], "path": item["path"], "value_type": item["value_type"], "expected_types": item["expected_types"]})
	}
	return fmt.Errorf("unreviewed native metadata fields or value types: %s", requireStringJSON(values))
}

// WriteMetadata writes a retained diagnostic report, including unknown values.
func WriteMetadata(path string, report M) error {
	b, e := jsonPrettyUTF8(report)
	if e != nil {
		return e
	}
	if e = os.MkdirAll(filepath.Dir(path), 0755); e != nil {
		return e
	}
	file, e := os.CreateTemp(filepath.Dir(path), ".native-metadata-")
	if e != nil {
		return e
	}
	name := file.Name()
	defer os.Remove(name)
	if _, e = file.Write(b); e != nil {
		file.Close()
		return e
	}
	if e = file.Close(); e != nil {
		return e
	}
	return os.Rename(name, path)
}
