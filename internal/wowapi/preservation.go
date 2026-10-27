package wowapi

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"wow-teal-types/internal/sourcelock"
	"wow-teal-types/internal/wowapi/literal"
)

var metadataComponent = regexp.MustCompile(`^(?:[A-Za-z_][A-Za-z_0-9]*|\$(?:ref|expression))$`)

type MetadataPolicy struct {
	BaselineSourceLockSHA256 string              `json:"baseline_source_lock_sha256"`
	Policy                   string              `json:"policy"`
	SchemaVersion            int                 `json:"schema_version"`
	ValueTypes               map[string][]string `json:"value_types"`
	Description              string              `json:"description"`
}
type MetadataReport struct {
	SchemaVersion        int                       `json:"schema_version"`
	Claim                string                    `json:"claim"`
	Sources              sourcelock.SourceIdentity `json:"sources"`
	MetadataPolicySHA256 string                    `json:"metadata_policy_sha256"`
	Statistics           MetadataStatistics        `json:"statistics"`
	Documents            []MetadataDocument        `json:"documents"`
	ManifestAccounted    bool                      `json:"manifest_accounted"`
	RepresentationNotes  []RepresentationNote      `json:"representation_notes"`
	ProjectionLimits     []string                  `json:"projection_limits"`
	UnknownFields        []UnknownMetadata         `json:"unknown_fields"`
	Passed               bool                      `json:"passed"`
}
type MetadataStatistics struct {
	NativeDocuments int `json:"native_documents"`
	NativeFunctions int `json:"native_functions"`
	NativeEvents    int `json:"native_events"`
	UnknownValues   int `json:"unknown_values"`
}
type MetadataDocument struct {
	Source          string         `json:"source"`
	ReviewedBlobSHA string         `json:"reviewed_blob_sha"`
	SourceSHA256    string         `json:"source_sha256"`
	Size            int            `json:"size"`
	Metadata        literal.Object `json:"metadata"`
}
type UnknownMetadata struct {
	Source        string        `json:"source"`
	Path          string        `json:"path"`
	SchemaPath    string        `json:"schema_path"`
	ValueType     string        `json:"value_type"`
	ExpectedTypes []string      `json:"expected_types"`
	Value         literal.Value `json:"value"`
}

func metadataValueType(value literal.Value) string {
	switch value.Kind() {
	case literal.Null:
		return "null"
	case literal.Boolean:
		return "boolean"
	case literal.Integer:
		return "integer"
	case literal.FloatKind:
		return "number"
	case literal.String:
		return "string"
	case literal.Array:
		return "array"
	case literal.ObjectKind, literal.Reference, literal.Expression:
		return "object"
	}
	reject("unsupported native metadata literal")
	return ""
}
func metadataWalk(value literal.Value, schemaPath, path string, visit func(string, string, literal.Value)) {
	visit(schemaPath, path, value)
	switch value.Kind() {
	case literal.ObjectKind:
		object, _ := value.Object()
		for _, name := range sortedKeys(object) {
			component := "." + name
			if !metadataComponent.MatchString(name) {
				component = "[" + requireStringJSON(name) + "]"
			}
			metadataWalk(object[name], schemaPath+component, path+component, visit)
		}
	case literal.Array:
		array, _ := value.Array()
		for i, child := range array {
			metadataWalk(child, schemaPath+"[]", path+"["+strconv.Itoa(i)+"]", visit)
		}
	}
}

type metadataIdentity struct{ source, kind, name string }

func metadataEqual[T any](a, b T) bool {
	aa, e := jsonCompact(a)
	if e != nil {
		return false
	}
	bb, e := jsonCompact(b)
	return e == nil && string(aa) == string(bb)
}
func rawArray(object literal.Object, key string) []literal.Value {
	if value, ok := object[key]; ok {
		if array, ok := value.Array(); ok {
			return array
		}
	}
	return []literal.Value{}
}

// AuditMetadata retains every literal and verifies the manifest before permitting publication.
func (p Project) AuditMetadata(manifest *Manifest) (report MetadataReport, err error) {
	defer captureError(&err)
	sourceLock, e := sourcelock.ReadSources(filepath.Join(p.Root, "scripts/wow_api/sources.json"))
	if e != nil {
		return report, e
	}
	policyPath := filepath.Join(p.Root, "scripts/wow_api/metadata_fields.json")
	policyData := mustFile(policyPath)
	var policy MetadataPolicy
	if e := decodeJSON(policyData, &policy); e != nil {
		return report, e
	}
	shapes := policy.ValueTypes
	validTypes := wordSet("null boolean integer number string array object")
	if policy.SchemaVersion != 1 || len(shapes) == 0 || shapes["$"] == nil {
		reject("invalid native metadata field policy")
	}
	for _, types := range shapes {
		if len(types) == 0 {
			reject("invalid native metadata field policy")
		}
		for _, kind := range types {
			if !validTypes[kind] {
				reject("invalid native metadata field policy")
			}
		}
	}
	expected := map[metadataIdentity]literal.Object{}
	documents, unknown := []MetadataDocument{}, []UnknownMetadata{}
	tables := map[string]ManifestNativeTables{}
	counts := map[string]int{"function": 0, "event": 0}
	rows := []sourcelock.SourceFile{}
	for _, row := range sourceLock["blizzard"].Files {
		if strings.HasSuffix(row.Name, ".lua") {
			rows = append(rows, row)
		}
	}
	if len(rows) == 0 {
		reject("native metadata input inventory is empty")
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Name < rows[j].Name })
	names := map[string]bool{}
	for _, row := range rows {
		name := row.Name
		if filepath.IsAbs(name) || containsString(strings.Split(filepath.ToSlash(name), "/"), "..") || names[name] {
			reject(fmt.Sprintf("invalid or duplicate native metadata input: %s", name))
		}
		names[name] = true
		path := p.InputPath("blizzard", name)
		data := mustFile(path)
		if blobSHA(data) != row.SHA || len(data) != row.Size {
			reject(fmt.Sprintf("native metadata source identity mismatch: %s", name))
		}
		metadata, e := ParseLua(data, path)
		if e != nil {
			return report, e
		}
		source := "blizzard:" + name
		documents = append(documents, MetadataDocument{Source: source, ReviewedBlobSHA: row.SHA, SourceSHA256: sha256hex(data), Size: len(data), Metadata: metadata})
		encoded, e := literal.Record(metadata).MarshalJSON()
		if e != nil {
			return report, e
		}
		walkValue, e := literal.ParseJSON(encoded)
		if e != nil {
			return report, e
		}
		metadataWalk(walkValue, "$", "$", func(schemaPath, location string, value literal.Value) {
			kind := metadataValueType(value)
			if !containsString(shapes[schemaPath], kind) {
				types := shapes[schemaPath]
				if types == nil {
					types = []string{}
				}
				unknown = append(unknown, UnknownMetadata{Source: source, Path: location, SchemaPath: schemaPath, ValueType: kind, ExpectedTypes: types, Value: value})
			}
		})
		for _, entry := range []struct{ key, kind string }{{"Functions", "function"}, {"Events", "event"}} {
			for i, v := range rawArray(metadata, entry.key) {
				item, ok := v.Object()
				if ok {
					for _, required := range []string{"Name", "Type"} {
						if _, ok := item[required]; !ok {
							unknown = append(unknown, UnknownMetadata{Source: source, Path: fmt.Sprintf("$.%s[%d].%s", entry.key, i, required), SchemaPath: "$." + entry.key + "[]." + required, ValueType: "absent", ExpectedTypes: []string{"string"}})
						}
					}
				}
				name, ok := item["Name"].Text()
				if !ok {
					continue
				}
				identity := metadataIdentity{source, entry.kind, name}
				if expected[identity] != nil {
					reject(fmt.Sprintf("duplicate native metadata contract: %v", identity))
				}
				expected[identity] = item
				counts[entry.kind]++
			}
		}
		tables[source] = ManifestNativeTables{Source: source, Tables: rawArray(metadata, "Tables"), Predicates: rawArray(metadata, "Predicates")}
	}
	if manifest != nil {
		observed := map[metadataIdentity]literal.Object{}
		for _, item := range manifest.Symbols {
			kind, name := "function", ""
			var metadata literal.Object
			if item.Event != nil {
				kind, name, metadata = "event", item.Event.Name, item.Event.Raw
			} else if item.Function != nil {
				name, metadata = item.Function.Name, item.Function.Raw
			}
			identity := metadataIdentity{item.Source, kind, name}
			if observed[identity] != nil {
				reject(fmt.Sprintf("duplicate manifest native metadata contract: %v", identity))
			}
			observed[identity] = metadata
		}
		if len(observed) != len(expected) {
			reject("manifest does not preserve native contracts: missing or extra contracts")
		}
		for key, value := range expected {
			if !metadataEqual(observed[key], value) {
				reject(fmt.Sprintf("manifest does not preserve native contracts: altered %s.%s", key.source, key.name))
			}
		}
		actualTables := map[string]ManifestNativeTables{}
		for _, item := range manifest.NativeTables {
			if _, ok := actualTables[item.Source]; ok {
				reject(fmt.Sprintf("duplicate manifest native metadata document: %s", item.Source))
			}
			actualTables[item.Source] = item
		}
		if !metadataEqual(actualTables, tables) {
			reject("manifest does not preserve complete native tables and predicates")
		}
		if counts["function"] != manifest.Statistics.NativeFunctions || counts["event"] != manifest.Statistics.NativeEvents {
			reject("native metadata contract counts differ from the manifest")
		}
	}
	notes := []RepresentationNote{}
	if manifest != nil {
		notes = manifest.Representations
	}
	report = MetadataReport{SchemaVersion: 1, Claim: "All selected native literal metadata is retained here. Original source bytes are separately bound by the client/source locks. Teal declarations alone do not express every source distinction.", Sources: sourceLock["blizzard"].Identity(), MetadataPolicySHA256: sha256hex(policyData), Statistics: MetadataStatistics{NativeDocuments: len(documents), NativeFunctions: counts["function"], NativeEvents: counts["event"], UnknownValues: len(unknown)}, Documents: documents, ManifestAccounted: manifest != nil, RepresentationNotes: notes, ProjectionLimits: []string{"Comments, lexical spelling, line endings and executable registration are retained in the locked source bytes, rather than in the parsed metadata values.", "Known metadata shapes establish parser coverage, not automatic approval of changed permissions or secret-value behavior.", "Teal nilability, repeated tuples, opaque markers and source-untyped values retain the representation limits recorded by the declaration manifest."}, UnknownFields: unknown, Passed: len(unknown) == 0}
	return
}
func RequireKnownMetadata(report MetadataReport) error {
	if len(report.UnknownFields) == 0 {
		return nil
	}
	type summary struct {
		Source        string   `json:"source"`
		Path          string   `json:"path"`
		ValueType     string   `json:"value_type"`
		ExpectedTypes []string `json:"expected_types"`
	}
	values := []summary{}
	for _, item := range report.UnknownFields[:min(10, len(report.UnknownFields))] {
		values = append(values, summary{item.Source, item.Path, item.ValueType, item.ExpectedTypes})
	}
	return fmt.Errorf("unreviewed native metadata fields or value types: %s", requireStringJSON(values))
}

// WriteMetadata writes a retained diagnostic report, including unknown values.
func WriteMetadata(path string, report MetadataReport) error {
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

// MetadataJSON retains UTF-8 spelling in the complete native metadata artifact.
func MetadataJSON(report MetadataReport) ([]byte, error) { return jsonPrettyUTF8(report) }
