package wowapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go.yaml.in/yaml/v3"
)

var nativeFamilies = map[string]struct {
	source, owner, canonical string
	count                    int
}{"DurationTextBindingObjectAPI": {"luaobjects.yaml", "DurationTextBinding", "DurationTextBindingObject", 30}, "LuaDurationClockObjectAPI": {"luaobjects.yaml", "LuaDurationClock", "LuaDurationClockObject", 1}, "LuaDurationManualClockAPI": {"luaobjects.yaml", "LuaDurationManualClock", "LuaDurationManualClock", 4}, "LuaLocaleContextAPI": {"luaobjects.yaml", "LuaLocaleContext", "LuaLocaleContext", 22}, "TimedSignalMapAPI": {"luaobjects.yaml", "LuaTimedSignalMapObject", "TimedSignalMap", 8}, "SimpleAnimRadialProgressAPI": {"uiobjects.yaml", "RadialProgress", "RadialProgress", 4}, "SimpleVectorGraphicsAPI": {"uiobjects.yaml", "VectorGraphics", "VectorGraphics", 4}}

func safeYAML(data []byte) (any, error) {
	if len(data) > 1<<20 {
		return nil, fmt.Errorf("YAML evidence exceeds 1 MiB")
	}
	var document yaml.Node
	dec := yaml.NewDecoder(bytes.NewReader(data))
	if e := dec.Decode(&document); e != nil {
		return nil, fmt.Errorf("invalid pinned YAML evidence: %w", e)
	}
	var extra yaml.Node
	if e := dec.Decode(&extra); e != io.EOF {
		return nil, fmt.Errorf("multiple YAML evidence documents are unsupported")
	}
	var convert func(*yaml.Node, int) (any, error)
	convert = func(n *yaml.Node, depth int) (any, error) {
		if depth > 48 {
			return nil, fmt.Errorf("YAML evidence nesting exceeds 48 levels")
		}
		if n.Alias != nil || n.Kind == yaml.AliasNode {
			return nil, fmt.Errorf("YAML aliases are not supported in pinned evidence")
		}
		switch n.Kind {
		case yaml.DocumentNode:
			if len(n.Content) != 1 {
				return nil, fmt.Errorf("invalid YAML evidence document")
			}
			return convert(n.Content[0], depth+1)
		case yaml.SequenceNode:
			a := []any{}
			for _, v := range n.Content {
				x, e := convert(v, depth+1)
				if e != nil {
					return nil, e
				}
				a = append(a, x)
			}
			return a, nil
		case yaml.MappingNode:
			m := M{}
			typed := map[any]any{}
			allStrings := true
			for i := 0; i < len(n.Content); i += 2 {
				key, e := convert(n.Content[i], depth+1)
				if e != nil {
					return nil, e
				}
				switch key.(type) {
				case string, int64:
				default:
					return nil, fmt.Errorf("YAML evidence mapping key must be a string or integer at line %d", n.Content[i].Line)
				}
				if _, ok := typed[key]; ok {
					return nil, fmt.Errorf("duplicate YAML evidence key: %v", key)
				}
				value, e := convert(n.Content[i+1], depth+1)
				if e != nil {
					return nil, e
				}
				typed[key] = value
				if s, ok := key.(string); ok {
					m[s] = value
				} else {
					allStrings = false
				}
			}
			if allStrings {
				return m, nil
			}
			return typed, nil
		case yaml.ScalarNode:
			if n.Style == 0 && n.Tag == "!!str" {
				// Match SafeLoader's YAML 1.1 implicit Boolean spellings, rather
				// than accepting arbitrary mixed-case strings as new policies.
				switch n.Value {
				case "yes", "Yes", "YES", "on", "On", "ON":
					return true, nil
				case "no", "No", "NO", "off", "Off", "OFF":
					return false, nil
				}
			}
			switch n.Tag {
			case "!!str":
				return n.Value, nil
			case "!!null":
				return nil, nil
			case "!!bool":
				switch strings.ToLower(n.Value) {
				case "true", "yes", "on":
					return true, nil
				case "false", "no", "off":
					return false, nil
				default:
					return nil, fmt.Errorf("invalid YAML evidence Boolean")
				}
			case "!!int":
				var v int64
				if e := n.Decode(&v); e != nil {
					return nil, e
				}
				return v, nil
			case "!!float":
				var v float64
				if e := n.Decode(&v); e != nil {
					return nil, e
				}
				if math.IsNaN(v) || math.IsInf(v, 0) {
					return nil, fmt.Errorf("nonfinite YAML evidence number")
				}
				return v, nil
			default:
				return nil, fmt.Errorf("unsupported YAML evidence tag: %s", n.Tag)
			}
		}
		return nil, fmt.Errorf("unsupported YAML evidence node")
	}
	return convert(&document, 0)
}
func evStrictMapping(value any, allowed []string, context string) (M, error) {
	m := asMap(value)
	if m == nil {
		return nil, fmt.Errorf("unsupported %s schema", context)
	}
	for k := range m {
		if !evContains(allowed, k) {
			return nil, fmt.Errorf("unsupported %s schema: %s", context, k)
		}
	}
	return m, nil
}
func secondaryType(v any) (any, error) {
	if s, ok := v.(string); ok {
		if !evContains([]string{"number", "string", "boolean", "function", "table", "any", "FileAsset"}, s) {
			return nil, fmt.Errorf("unsupported secondary primitive: %s", s)
		}
		return s, nil
	}
	m, e := evStrictMapping(v, []string{"arrayof", "enum", "luaobject", "uiobject", "structure", "stringenum"}, "secondary type")
	if e != nil {
		return nil, e
	}
	if len(m) != 1 {
		return nil, fmt.Errorf("secondary type requires one constructor")
	}
	for key, inner := range m {
		if key == "arrayof" {
			x, e := secondaryType(inner)
			return M{key: x}, e
		}
		if asString(inner) == "" {
			return nil, fmt.Errorf("secondary nominal type requires a name")
		}
		return M{key: inner}, nil
	}
	return nil, fmt.Errorf("empty secondary type")
}
func secondaryNativeType(v, typedefs M, tables map[string]string) (any, error) {
	name := asString(v["Type"])
	if inner, ok := v["InnerType"]; ok {
		x, e := secondaryNativeType(M{"Type": inner}, typedefs, tables)
		return M{"arrayof": x}, e
	}
	if def := asMap(typedefs[name]); def != nil {
		return secondaryType(def["type"])
	}
	kind := tables[name]
	if kind == "" && name == "colorRGBA" {
		kind = "Structure"
	}
	if kind == "Constants" {
		return "number", nil
	}
	if constructor := map[string]string{"Enumeration": "enum", "Structure": "structure", "stringenum": "stringenum"}[kind]; constructor != "" {
		return M{constructor: name}, nil
	}
	if kind == "CallbackType" {
		return "function", nil
	}
	return nil, fmt.Errorf("unmapped native secondary carrier: %s", name)
}
func nativeDefault(v any) (any, error) {
	switch x := v.(type) {
	case nil, string, bool, int, int64, json.Number:
		return v, nil
	case float64:
		if !math.IsInf(x, 0) && !math.IsNaN(x) {
			return v, nil
		}
	case M:
		if len(x) == 1 && asString(x["$ref"]) != "" {
			return v, nil
		}
	}
	return nil, fmt.Errorf("unsupported default literal or enum reference")
}
func nativeStride(v any) (any, error) {
	if v == nil {
		return nil, nil
	}
	if n, ok := evidenceInteger(v); ok && n > 0 {
		return v, nil
	}
	return nil, fmt.Errorf("stride must be a positive integer")
}
func secondaryParameter(value M, native bool, typedefs M, tables map[string]string) (M, error) {
	nameKey, typeKey, nilKey, defaultKey, strideKey := "name", "type", "nilable", "default", "stride"
	if native {
		nameKey, typeKey, nilKey, defaultKey, strideKey = "Name", "Type", "Nilable", "Default", "StrideIndex"
	} else {
		if _, e := evStrictMapping(value, []string{"name", "type", "nilable", "default", "stride"}, "secondary parameter"); e != nil {
			return nil, e
		}
	}
	if asString(value[nameKey]) == "" || value[typeKey] == nil {
		return nil, fmt.Errorf("secondary parameter requires name and type")
	}
	nilable := false
	if v, ok := value[nilKey]; ok {
		var yes bool
		nilable, yes = v.(bool)
		if !yes {
			return nil, fmt.Errorf("nilability must be Boolean")
		}
	}
	var typ any
	var e error
	if native {
		typ, e = secondaryNativeType(value, typedefs, tables)
	} else {
		typ, e = secondaryType(value[typeKey])
	}
	if e != nil {
		return nil, e
	}
	def := M{"present": false}
	if v, ok := value[defaultKey]; ok {
		def["present"] = true
		def["value"], e = nativeDefault(v)
		if e != nil {
			return nil, e
		}
	}
	stride, e := nativeStride(value[strideKey])
	if e != nil {
		return nil, e
	}
	return M{"name": value[nameKey], "type": typ, "nilable": nilable, "default": def, "stride": stride}, nil
}
func compareSecondaryMethod(native, secondary, typedefs M, tables map[string]string) (M, error) {
	if _, e := evStrictMapping(secondary, []string{"inputs", "outputs", "mayreturnnothing"}, "secondary method"); e != nil {
		return nil, e
	}
	for _, key := range []string{"inputs", "outputs"} {
		if _, ok := secondary[key]; !ok {
			return nil, fmt.Errorf("secondary method requires both input and output tuples")
		}
	}
	for _, pair := range []struct {
		m M
		k string
	}{{secondary, "mayreturnnothing"}, {native, "MayReturnNothing"}} {
		if v, ok := pair.m[pair.k]; ok {
			if _, ok := v.(bool); !ok {
				return nil, fmt.Errorf("absence flag must be Boolean")
			}
		}
	}
	differences := []any{}
	tuples := M{}
	for _, side := range []struct{ name, raw, source string }{{"arguments", "Arguments", "inputs"}, {"returns", "Returns", "outputs"}} {
		raw := secondary[side.source]
		if raw != nil && asList(raw) == nil {
			return nil, fmt.Errorf("secondary tuple must be a list or null")
		}
		actual, expected := []any{}, []any{}
		for _, v := range asList(raw) {
			m, e := secondaryParameter(asMap(v), false, typedefs, tables)
			if e != nil {
				return nil, e
			}
			actual = append(actual, m)
		}
		for _, v := range asList(native[side.raw]) {
			m, e := secondaryParameter(asMap(v), true, typedefs, tables)
			if e != nil {
				return nil, e
			}
			expected = append(expected, m)
		}
		tuples[side.name] = M{"native": expected, "secondary": actual}
		if len(expected) != len(actual) {
			differences = append(differences, M{"kind": side.name + "-count", "native": len(expected), "secondary": len(actual)})
		}
		for i := 0; i < min(len(expected), len(actual)); i++ {
			left, right := asMap(expected[i]), asMap(actual[i])
			for _, key := range []string{"name", "type", "nilable", "default", "stride"} {
				if !evEqual(left[key], right[key]) {
					differences = append(differences, M{"kind": side.name + "-" + key, "position": i + 1, "native": left[key], "secondary": right[key]})
				}
			}
		}
	}
	if truth(native["MayReturnNothing"]) != truth(secondary["mayreturnnothing"]) {
		differences = append(differences, M{"kind": "may-return-nothing", "native": truth(native["MayReturnNothing"]), "secondary": truth(secondary["mayreturnnothing"])})
	}
	return M{"tuples": tuples, "differences": differences}, nil
}
func compareHierarchy(bases []string, secondary M) (M, error) {
	parents := []string{}
	switch v := secondary["inherits"].(type) {
	case string:
		parents = append(parents, v)
	case M:
		for _, key := range sortedKeys(v) {
			if v[key] != nil {
				return nil, fmt.Errorf("unsupported secondary receiver inheritance")
			}
			parents = append(parents, key)
		}
	case nil:
	default:
		return nil, fmt.Errorf("unsupported secondary receiver inheritance")
	}
	for i, s := range parents {
		if s == "LuaDurationClock" {
			parents[i] = "LuaDurationClockObject"
		}
	}
	sort.Strings(parents)
	bases = append([]string{}, bases...)
	sort.Strings(bases)
	diff := []any{}
	if !evEqual(evList(bases), evList(parents)) {
		diff = append(diff, M{"kind": "receiver-hierarchy", "native": evList(bases), "secondary": evList(parents), "disposition": "retain-workspace-bases; secondary hierarchy lacks independent native proof"})
	}
	return M{"workspace_bases": evList(bases), "secondary_bases": evList(parents), "secondary_namespace": "wowless receiver/facet names; only explicit native aliases are mapped", "differences": diff}, nil
}
func nativeSecurity(native M) M {
	restricted := func(v M) M {
		r := M{}
		for key, value := range v {
			if strings.Contains(key, "Secret") || strings.HasPrefix(key, "Requires") || strings.HasPrefix(key, "Check") {
				r[key] = value
			}
		}
		return r
	}
	result := M{"function": restricted(native)}
	for _, side := range []string{"Arguments", "Returns"} {
		positions := []any{}
		for i, v := range asList(native[side]) {
			m := restricted(asMap(v))
			if len(m) > 0 {
				positions = append(positions, M{"position": i + 1, "metadata": m})
			}
		}
		if len(positions) > 0 {
			result[strings.ToLower(side)] = positions
		}
	}
	if len(asMap(result["function"])) == 0 && len(result) == 1 {
		return M{}
	}
	return result
}
func (p Project) loadNativeSecondary() (M, map[string]string, error) {
	lock, e := p.Lock("native_crosscheck_sources.json")
	if e != nil {
		return nil, nil, e
	}
	files := map[string]string{}
	for _, v := range asList(lock["files"]) {
		row := asMap(v)
		name := asString(row["name"])
		if !evidenceSafeRelative(name) {
			return nil, nil, fmt.Errorf("invalid native evidence path")
		}
		if _, ok := files[name]; ok {
			return nil, nil, fmt.Errorf("duplicate native evidence path")
		}
		data, e := readEvidenceFile(filepath.Join(p.Root, "references/wow-api/native-crosscheck", name), 1<<20)
		if e != nil {
			return nil, nil, e
		}
		if len(data) != asInt(row["size"]) || sha256hex(data) != row["sha256"] || blobSHA(data) != row["git_blob_sha1"] {
			return nil, nil, fmt.Errorf("native cross-check integrity failure: %s", name)
		}
		files[name] = string(data)
	}
	buildData, e := safeYAML([]byte(files["wowless/data/products/wow/build.yaml"]))
	if e != nil {
		return nil, nil, e
	}
	build := asMap(buildData)
	if asString(build["version"])+"."+fmt.Sprint(build["build"]) != asMap(lock["wowless"])["build"] {
		return nil, nil, fmt.Errorf("native cross-check build provenance mismatch")
	}
	return lock, files, nil
}
func (p Project) nativeDocuments() ([]nativeDocument, error) {
	lock, e := p.Lock("sources.json")
	if e != nil {
		return nil, e
	}
	r := []nativeDocument{}
	for _, v := range asList(asMap(lock["blizzard"])["files"]) {
		row := asMap(v)
		name := asString(row["name"])
		data, e := os.ReadFile(p.InputPath("blizzard", name))
		if e != nil {
			return nil, e
		}
		if blobSHA(data) != row["sha"] {
			return nil, fmt.Errorf("modified native cross-check source: %s", name)
		}
		if strings.HasSuffix(name, ".lua") {
			m, e := ParseLua(data, name)
			if e != nil {
				return nil, e
			}
			r = append(r, nativeDocument{Name: name, Metadata: m})
		}
	}
	return r, nil
}

// ParseEvidenceYAML admits the same bounded, alias-free evidence shape used by cross-checks.
func ParseEvidenceYAML(text string) (any, error) { return safeYAML([]byte(text)) }
