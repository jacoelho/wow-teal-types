package wowapi

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"wow-teal-types/internal/sourcelock"
	"wow-teal-types/internal/wowapi/literal"
)

var nativeFamilies = map[string]struct {
	source, owner, canonical string
	count                    int
}{"DurationTextBindingObjectAPI": {"luaobjects.yaml", "DurationTextBinding", "DurationTextBindingObject", 30}, "LuaDurationClockObjectAPI": {"luaobjects.yaml", "LuaDurationClock", "LuaDurationClockObject", 1}, "LuaDurationManualClockAPI": {"luaobjects.yaml", "LuaDurationManualClock", "LuaDurationManualClock", 4}, "LuaLocaleContextAPI": {"luaobjects.yaml", "LuaLocaleContext", "LuaLocaleContext", 22}, "TimedSignalMapAPI": {"luaobjects.yaml", "LuaTimedSignalMapObject", "TimedSignalMap", 8}, "SimpleAnimRadialProgressAPI": {"uiobjects.yaml", "RadialProgress", "RadialProgress", 4}, "SimpleVectorGraphicsAPI": {"uiobjects.yaml", "VectorGraphics", "VectorGraphics", 4}}

func secondaryNativeType(v NativeParameter, typedefs map[string]secondaryTypedef, tables map[string]string) (secondaryCarrier, error) {
	if v.InnerType != "" {
		inner, e := secondaryNativeType(NativeParameter{Type: v.InnerType}, typedefs, tables)
		return secondaryCarrier{Constructor: "arrayof", Element: &inner}, e
	}
	if def, ok := typedefs[v.Type]; ok {
		return def.Type, nil
	}
	kind := tables[v.Type]
	if kind == "" && v.Type == "colorRGBA" {
		kind = "Structure"
	}
	if kind == "Constants" {
		return secondaryCarrier{Primitive: "number"}, nil
	}
	if constructor := map[string]string{"Enumeration": "enum", "Structure": "structure", "stringenum": "stringenum"}[kind]; constructor != "" {
		return secondaryCarrier{Constructor: constructor, Name: v.Type}, nil
	}
	if kind == "CallbackType" {
		return secondaryCarrier{Primitive: "function"}, nil
	}
	return secondaryCarrier{}, fmt.Errorf("unmapped native secondary carrier: %s", v.Type)
}
func nativeParameterFact(value NativeParameter, typedefs map[string]secondaryTypedef, tables map[string]string) (secondaryParameterFact, error) {
	var fact secondaryParameterFact
	if value.Name == "" || value.Type == "" {
		return fact, fmt.Errorf("secondary parameter requires name and type")
	}
	typ, e := secondaryNativeType(value, typedefs, tables)
	if e != nil {
		return fact, e
	}
	def := parameterDefault{}
	if value.HasDefault {
		v, e := parseDefault(value.Raw["Default"])
		if e != nil {
			return fact, e
		}
		def = parameterDefault{Present: true, Value: &v}
	}
	if value.StrideIndex != nil && *value.StrideIndex <= 0 {
		return fact, fmt.Errorf("stride must be a positive integer")
	}
	return secondaryParameterFact{Name: value.Name, Type: typ, Nilable: enabled(value.Nilable), Default: def, Stride: value.StrideIndex}, nil
}
func compareSecondaryMethod(native NativeFunction, secondary SecondaryMethod, typedefs map[string]secondaryTypedef, tables map[string]string) (secondaryComparison, error) {
	result := secondaryComparison{Differences: []secondaryDifference{}}
	add := func(kind string, left, right secondaryComparedValue, pos int) {
		result.Differences = append(result.Differences, secondaryDifference{Kind: kind, Native: left, Secondary: right, Position: pos})
	}
	for _, side := range []struct {
		name      string
		native    []NativeParameter
		secondary []SecondaryParameter
		tuple     *secondaryTuple
	}{{"arguments", native.Arguments, secondary.Inputs, &result.Tuples.Arguments}, {"returns", native.Returns, secondary.Outputs, &result.Tuples.Returns}} {
		expected, actual := []secondaryParameterFact{}, []secondaryParameterFact{}
		for _, v := range side.secondary {
			actual = append(actual, secondaryParameterFact{Name: v.Name, Type: v.Type, Nilable: v.Nilable, Default: v.Default, Stride: v.Stride})
		}
		for _, v := range side.native {
			fact, e := nativeParameterFact(v, typedefs, tables)
			if e != nil {
				return result, e
			}
			expected = append(expected, fact)
		}
		*side.tuple = secondaryTuple{Native: expected, Secondary: actual}
		if len(expected) != len(actual) {
			add(side.name+"-count", secondaryComparedValue{Integer: new(len(expected))}, secondaryComparedValue{Integer: new(len(actual))}, 0)
		}
		for i := 0; i < min(len(expected), len(actual)); i++ {
			left, right := expected[i], actual[i]
			if left.Name != right.Name {
				add(side.name+"-name", secondaryComparedValue{Text: &left.Name}, secondaryComparedValue{Text: &right.Name}, i+1)
			}
			if !evEqual(left.Type, right.Type) {
				add(side.name+"-type", secondaryComparedValue{Type: &left.Type}, secondaryComparedValue{Type: &right.Type}, i+1)
			}
			if left.Nilable != right.Nilable {
				add(side.name+"-nilable", secondaryComparedValue{Boolean: &left.Nilable}, secondaryComparedValue{Boolean: &right.Nilable}, i+1)
			}
			if !evEqual(left.Default, right.Default) {
				add(side.name+"-default", secondaryComparedValue{Default: &left.Default}, secondaryComparedValue{Default: &right.Default}, i+1)
			}
			if !evEqual(left.Stride, right.Stride) {
				add(side.name+"-stride", secondaryComparedValue{Integer: left.Stride}, secondaryComparedValue{Integer: right.Stride}, i+1)
			}
		}
	}
	if enabled(native.Policy.MayReturnNothing) != secondary.MayReturnNothing {
		add("may-return-nothing", secondaryComparedValue{Boolean: new(enabled(native.Policy.MayReturnNothing))}, secondaryComparedValue{Boolean: &secondary.MayReturnNothing}, 0)
	}
	return result, nil
}
func compareHierarchy(bases []string, secondary SecondaryReceiver) (receiverHierarchy, error) {
	parents := append([]string{}, secondary.Inherits...)
	for i, s := range parents {
		if s == "LuaDurationClock" {
			parents[i] = "LuaDurationClockObject"
		}
	}
	sort.Strings(parents)
	bases = append([]string{}, bases...)
	sort.Strings(bases)
	diff := []hierarchyDifference{}
	if !evEqual(bases, parents) {
		diff = append(diff, hierarchyDifference{Kind: "receiver-hierarchy", Native: bases, Secondary: parents, Disposition: "retain-workspace-bases; secondary hierarchy lacks independent native proof"})
	}
	return receiverHierarchy{WorkspaceBases: bases, SecondaryBases: parents, SecondaryNamespace: "wowless receiver/facet names; only explicit native aliases are mapped", Differences: diff}, nil
}

type securityPosition struct {
	Metadata literal.Object `json:"metadata"`
	Position int            `json:"position"`
}
type securityEvidence struct {
	Arguments []securityPosition `json:"arguments,omitempty"`
	Function  literal.Object     `json:"function,omitempty"`
	Returns   []securityPosition `json:"returns,omitempty"`
}

func nativeSecurity(native NativeFunction) securityEvidence {
	restricted := func(raw literal.Object) literal.Object {
		out := literal.Object{}
		for key, value := range raw {
			if strings.Contains(key, "Secret") || strings.HasPrefix(key, "Requires") || strings.HasPrefix(key, "Check") {
				out[key] = value
			}
		}
		return out
	}
	result := securityEvidence{Function: restricted(native.Raw)}
	for _, side := range []struct {
		values []NativeParameter
		target *[]securityPosition
	}{{native.Arguments, &result.Arguments}, {native.Returns, &result.Returns}} {
		for i, value := range side.values {
			m := restricted(value.Raw)
			if len(m) > 0 {
				*side.target = append(*side.target, securityPosition{Metadata: m, Position: i + 1})
			}
		}
	}
	if len(result.Function) == 0 && (len(result.Arguments) > 0 || len(result.Returns) > 0) {
		result.Function = literal.Object{}
	}
	return result
}
func (s securityEvidence) empty() bool {
	return len(s.Function) == 0 && len(s.Arguments) == 0 && len(s.Returns) == 0
}
func (p Project) loadNativeSecondary() (sourcelock.Crosscheck, map[string]string, error) {
	lock, e := sourcelock.ReadCrosscheck(filepath.Join(p.Root, "scripts/wow_api/native_crosscheck_sources.json"))
	if e != nil {
		return lock, nil, e
	}
	files := map[string]string{}
	for _, row := range lock.Files {
		if !evidenceSafeRelative(row.Name) {
			return lock, nil, fmt.Errorf("invalid native evidence path")
		}
		if _, ok := files[row.Name]; ok {
			return lock, nil, fmt.Errorf("duplicate native evidence path")
		}
		data, e := readEvidenceFile(filepath.Join(p.Root, "references/wow-api/native-crosscheck", row.Name), 1<<20)
		if e != nil {
			return lock, nil, e
		}
		if len(data) != row.Size || sha256hex(data) != row.SHA256 || blobSHA(data) != row.GitBlobSHA1 {
			return lock, nil, fmt.Errorf("native cross-check integrity failure: %s", row.Name)
		}
		files[row.Name] = string(data)
	}
	build, e := secondaryBuild([]byte(files["wowless/data/products/wow/build.yaml"]))
	if e != nil {
		return lock, nil, e
	}
	if build != lock.Wowless.Build {
		return lock, nil, fmt.Errorf("native cross-check build provenance mismatch")
	}
	return lock, files, nil
}
func (p Project) nativeDocuments() ([]NativeDocument, error) { return p.readNativeDocuments() }

func (s securityEvidence) MarshalJSON() ([]byte, error) {
	if s.empty() {
		return []byte("{}"), nil
	}
	type report struct {
		Arguments []securityPosition `json:"arguments,omitempty"`
		Function  literal.Object     `json:"function"`
		Returns   []securityPosition `json:"returns,omitempty"`
	}
	return json.Marshal(report{s.Arguments, s.Function, s.Returns})
}
