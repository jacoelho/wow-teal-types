package wowapi

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"wow-teal-types/internal/sourcelock"
)

var tealKeywords = wordSet("and break do else elseif end false for function goto if in local nil not or repeat return then true until while global type record enum interface is as where")
var identifierReplace = regexp.MustCompile(`[^A-Za-z_0-9]`)
var plainIdentifier = regexp.MustCompile(`^[A-Za-z_]\w*$`)
var dottedIdentifiers = regexp.MustCompile(`[A-Za-z_]\w*(?:\.[A-Za-z_]\w*)*`)

func ident(name string) string {
	value := identifierReplace.ReplaceAllString(name, "_")
	switch value {
	case "string", "table", "math", "coroutine", "debug", "bit":
		return "Namespace_" + value
	}
	if len(value) > 0 && (value[0] >= '0' && value[0] <= '9' || tealKeywords[value]) {
		return "T_" + value
	}
	return value
}
func requireStringJSON[T any](v T) string {
	b, e := jsonDefault(v)
	if e != nil {
		reject(fmt.Sprintf("JSON encoding: %s", e))
	}
	return string(b)
}
func stringSet(values ...string) map[string]bool {
	out := map[string]bool{}
	for _, v := range values {
		out[v] = true
	}
	return out
}
func setAdd(target map[string]bool, source map[string]bool) {
	for k, v := range source {
		if v {
			target[k] = true
		}
	}
}
func setEqual(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}

type catalogue struct {
	project                                                                                           Project
	defs                                                                                              map[string]*catalogueDefinition
	aliases                                                                                           map[string]*catalogueAlias
	globals                                                                                           map[string]*catalogueFunction
	symbols                                                                                           []ManifestSymbol
	files                                                                                             []ManifestFile
	notes                                                                                             []RepresentationNote
	errors                                                                                            []string
	questie                                                                                           []AnnotationDocument
	nativeDocs                                                                                        []NativeDocument
	used, secretStructs, nativeNames, nativeFunctions, resolving, secureGlobals, absencePreconditions map[string]bool
	abiTuples                                                                                         map[string][]string
	excludedMethods                                                                                   map[string]SupplementalExclusion
	defOrder, globalOrder                                                                             []string
}

func newCatalogue(p Project) *catalogue {
	return &catalogue{project: p, defs: map[string]*catalogueDefinition{"OpaqueValue": {Kind: "external", Source: "opaque native/LuaLS unknown carrier"}}, aliases: map[string]*catalogueAlias{}, globals: map[string]*catalogueFunction{}, symbols: []ManifestSymbol{}, files: []ManifestFile{}, notes: []RepresentationNote{}, errors: []string{}, questie: []AnnotationDocument{}, nativeDocs: []NativeDocument{}, used: map[string]bool{}, secretStructs: map[string]bool{}, nativeNames: map[string]bool{}, nativeFunctions: map[string]bool{}, resolving: map[string]bool{}, secureGlobals: map[string]bool{}, absencePreconditions: map[string]bool{}, abiTuples: map[string][]string{}, excludedMethods: map[string]SupplementalExclusion{}, defOrder: []string{"OpaqueValue"}, globalOrder: []string{}}
}
func (c *catalogue) setDef(name string, item *catalogueDefinition) {
	if _, ok := c.defs[name]; !ok {
		c.defOrder = append(c.defOrder, name)
	}
	c.defs[name] = item
}
func (c *catalogue) setGlobal(name string, item *catalogueFunction) {
	if _, ok := c.globals[name]; !ok {
		c.globalOrder = append(c.globalOrder, name)
	}
	c.globals[name] = item
}
func (c *catalogue) record(name, source string, kind ...string) *catalogueDefinition {
	delete(c.aliases, name)
	if v := c.defs[name]; v != nil {
		return v
	}
	k := "record"
	if len(kind) > 0 {
		k = kind[0]
	}
	v := &catalogueDefinition{Kind: k, Source: source, Fields: map[string]*catalogueField{}, Methods: map[string]*catalogueFunction{}, Bases: []string{}}
	c.setDef(name, v)
	return v
}
func (c *catalogue) literal(values []string) string {
	key := "Literal_" + sha256hex([]byte(requireStringJSON(values)))[:12]
	if c.defs[key] == nil {
		c.setDef(key, &catalogueDefinition{Kind: "enum", Values: values, Source: "interned LuaLS string literals"})
	}
	return ident(key)
}

type anonymousFieldKey struct {
	name, typeName string
	optional       bool
}

func (v anonymousFieldKey) MarshalJSON() ([]byte, error) {
	name, e := jsonCompact(v.name)
	if e != nil {
		return nil, e
	}
	typ, e := jsonCompact(v.typeName)
	if e != nil {
		return nil, e
	}
	optional := "false"
	if v.optional {
		optional = "true"
	}
	return []byte("[" + string(name) + "," + string(typ) + "," + optional + "]"), nil
}
func (c *catalogue) anonymous(fields []recordField) string {
	rows := []anonymousFieldKey{}
	mapped := map[string]*catalogueField{}
	for _, f := range fields {
		rows = append(rows, anonymousFieldKey{f.name, f.typeName, f.optional})
		mapped[f.name] = &catalogueField{Type: f.typeName}
	}
	key := "Shape_" + sha256hex([]byte(requireStringJSON(rows)))[:12]
	if c.defs[key] == nil {
		c.setDef(key, &catalogueDefinition{Kind: "converted_record", Fields: mapped, Source: "interned LuaLS record"})
	}
	return ident(key)
}
func (c *catalogue) resolve(name string) string {
	if name == "self" {
		unsupported(name, "self requires receiver context")
	}
	if name == "any" {
		return "unknown"
	}
	if containsString([]string{"number", "integer", "string", "boolean", "nil", "unknown", "thread"}, name) {
		return name
	}
	if c.defs[name] == nil && c.aliases[name] == nil && c.aliases["Enum."+name] != nil {
		c.aliases[name] = &catalogueAlias{Expression: "Enum." + name, Source: "Questie numeric enum name"}
	}
	if c.aliases[name] != nil && c.defs[name] == nil {
		if c.resolving[name] {
			unsupported(name, "cyclic type alias")
		}
		c.resolving[name] = true
		defer delete(c.resolving, name)
		alias := c.aliases[name]
		if alias.Signature != nil {
			return c.signatureValue(alias.Signature, "")
		}
		return c.tealValue(alias.Expression, "", nil)
	}
	if c.defs[name] != nil || c.aliases[name] != nil {
		c.used[name] = true
		return ident(name)
	}
	unsupported(name, "unresolved named type")
	return ""
}
func pythonTuple(values []string) string {
	if len(values) == 0 {
		return "()"
	}
	out := []string{}
	for _, v := range values {
		quote := "'"
		if strings.Contains(v, "'") && !strings.Contains(v, "\"") {
			quote = "\""
		}
		v = strings.ReplaceAll(v, "\\", "\\\\")
		v = strings.ReplaceAll(v, quote, "\\"+quote)
		out = append(out, quote+v+quote)
	}
	if len(out) == 1 {
		return "(" + out[0] + ",)"
	}
	return "(" + strings.Join(out, ", ") + ")"
}
func (c *catalogue) teal(expression, owner string, generics []string) (string, error) {
	return captureString(func() string { return c.tealValue(expression, owner, generics) })
}
func (c *catalogue) tealValue(expression, owner string, generics []string) string {
	converter := typeConverter{raw: expression, self: ident(owner), literal: c.literal, record: c.anonymous}
	converter.resolve = func(name string) string {
		if containsString(generics, name) {
			return ident(name)
		}
		return c.resolve(name)
	}
	converter.callback = func(signature string) string {
		usedGenerics := []string{}
		for _, generic := range generics {
			if regexp.MustCompile(`\b` + regexp.QuoteMeta(generic) + `\b`).MatchString(signature) {
				usedGenerics = append(usedGenerics, generic)
			}
		}
		key := "Callback_" + sha256hex([]byte(signature + pythonTuple(usedGenerics)))[:12]
		if _, ok := c.defs[key]; !ok {
			c.setDef(key, &catalogueDefinition{Kind: "callback", Signature: signature, Generics: usedGenerics, Source: "named LuaLS callback"})
		}
		if len(usedGenerics) > 0 {
			return key + "<" + strings.Join(usedGenerics, ", ") + ">"
		}
		return key
	}
	converter.union = func(values []string) string {
		all := []string{}
		for _, value := range values {
			all = append(all, splitTop(value, '|')...)
		}
		values = uniqueStrings(all)
		if containsString(values, "unknown") {
			return "unknown"
		}
		named := map[string]*catalogueDefinition{}
		for name, value := range c.defs {
			named[ident(name)] = value
		}
		enums := []string{}
		for _, value := range values {
			if named[value] != nil && named[value].Kind == "enum" {
				enums = append(enums, value)
			}
		}
		if len(enums) > 0 {
			next := []string{}
			for _, value := range values {
				if !containsString(enums, value) {
					next = append(next, value)
				}
			}
			values = next
			if !containsString(values, "string") {
				members := []string{}
				for _, value := range enums {
					members = append(members, named[value].Values...)
				}
				members = uniqueStrings(members)
				sort.Strings(members)
				values = append(values, c.literal(members))
			}
		}
		tables := []string{}
		for _, value := range values {
			record := named[value]
			if strings.HasPrefix(value, "{") || record != nil && containsString([]string{"record", "interface", "converted_record"}, record.Kind) && !record.Opaque {
				tables = append(tables, value)
			}
		}
		if len(tables) > 1 {
			replacement := "{unknown: unknown}"
			if !containsString(tables, replacement) {
				ordered := append([]string{}, tables...)
				sort.Strings(ordered)
				replacement = "TableVariants_" + sha256hex([]byte(strings.Join(ordered, "|")))[:12]
				if c.defs[replacement] == nil {
					c.setDef(replacement, &catalogueDefinition{Kind: "interface", Source: "opaque result surface of " + strings.Join(tables, " | "), Fields: map[string]*catalogueField{}, Methods: map[string]*catalogueFunction{}, Bases: []string{}})
				}
				for _, table := range tables {
					record := named[table]
					if record == nil {
						reject(fmt.Sprintf("unknown table union member %s", table))
					}
					if !containsString(record.Bases, replacement) {
						record.Bases = append(record.Bases, replacement)
					}
				}
			}
			c.notes = append(c.notes, RepresentationNote{Expression: expression, Representation: replacement, Rule: "common table surface of source union"})
			next := []string{}
			for _, v := range values {
				if !containsString(tables, v) {
					next = append(next, v)
				}
			}
			values = append(next, replacement)
		}
		return strings.Join(values, " | ")
	}
	return converter.walk(expression, 0)
}
func (c *catalogue) sourceLocks() sourcelock.Sources {
	locks, e := sourcelock.ReadSources(filepath.Join(c.project.Root, "scripts/wow_api/sources.json"))
	if e != nil {
		reject(e.Error())
	}
	return locks
}
func (c *catalogue) loadQuestie() {
	rows := c.sourceLocks()["questie"].Files
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].Name < rows[j].Name })
	for _, row := range rows {
		if strings.HasSuffix(row.Name, ".lua") {
			doc, e := parseAnnotationFile(c.project.InputPath("questie", row.Name), row.Name)
			if e != nil {
				reject(e.Error())
			}
			c.questie = append(c.questie, doc)
		}
	}
	for _, doc := range c.questie {
		path := doc.Path
		source := "questie:" + path
		if strings.HasPrefix(path, "Lua/") {
			c.files = append(c.files, ManifestFile{Source: source, Disposition: "compiler-builtins"})
			continue
		}
		c.files = append(c.files, ManifestFile{Source: source, Disposition: "converted"})
		for _, alias := range doc.Aliases {
			alternatives := []string{}
			if alias.Type != "" {
				alternatives = append(alternatives, alias.Type)
			}
			for _, v := range alias.Alternatives {
				alternatives = append(alternatives, strings.TrimRight(v.Type, "| "))
			}
			expression := strings.Join(alternatives, " | ")
			if expression == "" {
				c.errors = append(c.errors, fmt.Sprintf("%s:%d: empty alias %s", source, alias.Line, alias.Name))
				continue
			}
			if strings.HasSuffix(expression, "|") {
				c.notes = append(c.notes, RepresentationNote{Symbol: stringPointer(alias.Name), Rule: "remove empty trailing union member", Source: source})
				expression = strings.TrimRight(expression, "| ")
			}
			c.aliases[alias.Name] = &catalogueAlias{Expression: expression, Source: source, Line: alias.Line}
		}
		for _, cls := range doc.Classes {
			record := c.record(cls.Name, source, "interface")
			for _, base := range cls.Bases {
				if !containsString(record.Bases, base) {
					record.Bases = append(record.Bases, base)
				}
			}
			for _, field := range cls.Fields {
				record.Fields[field.Name] = &catalogueField{Type: field.Type, Source: source}
			}
		}
		for _, enum := range doc.Enums {
			numeric := true
			for _, v := range enum.Values {
				if !v.Value.numeric() {
					numeric = false
				}
			}
			expression := "string"
			if numeric {
				expression = "number"
			}
			name := enum.Name
			c.aliases[name] = &catalogueAlias{Expression: expression, Source: source}
			recordName := name + "_Values"
			record := c.record(recordName, source)
			for _, v := range enum.Values {
				typ := "string"
				if v.Value.numeric() {
					typ = "number"
				}
				record.Fields[v.Name] = &catalogueField{Type: typ, Source: source}
			}
			c.bindGlobal(name, recordName, source)
		}
		for _, function := range doc.Functions {
			if strings.HasPrefix(strings.TrimLeft(function.Raw, " \t"), "local function ") {
				continue
			}
			item := functionFromAnnotation(function, source)
			owner := stringValue(function.Owner)
			name := function.Name
			if owner == "Region" {
				owner = nativeMethodOwner("SimpleRegionAPI", name)
			}
			if owner != "" {
				methods := c.record(owner, source).Methods
				if previous := methods[name]; previous != nil {
					item.Variants = append(append([]*catalogueFunction{}, previous.Variants...), previous)
				}
				methods[name] = item
			} else {
				c.setGlobal(name, item)
			}
		}
		for _, binding := range doc.Bindings {
			raw := binding.Raw
			if binding.Local || len(raw) > 0 && (raw[0] == ' ' || raw[0] == '\t') {
				continue
			}
			name, value := strings.TrimPrefix(binding.Name, "_G."), binding.Value
			target := stringValue(binding.Class)
			if target == "" {
				switch {
				case strings.HasPrefix(value, "CreateColor("):
					target = "ColorMixin"
				case strings.HasPrefix(value, "CreateFont("):
					target = "Font"
				case strings.HasPrefix(value, "CreateFromMixins("):
					target = name
					mixins := strings.SplitN(strings.TrimPrefix(value, "CreateFromMixins("), ")", 2)[0]
					record := c.record(name, source, "interface")
					for _, m := range strings.Split(mixins, ",") {
						record.Bases = append(record.Bases, strings.TrimSpace(m))
					}
				case strings.HasPrefix(value, "{}"):
					target = name
					c.record(name, source)
				case numericTypePattern.MatchString(value):
					target = "number"
				}
			}
			if target != "" && c.globals[name] == nil {
				c.bindGlobal(name, target, source)
			}
		}
	}
	for name, expression := range map[string]string{"bool": "boolean", "cstring": "string", "luaIndex": "number", "UnitTokenType": "string", "UnitTokenVariant": "string", "LuaFunctionReference": "function", "Secret": "unknown", "Seconds": "number"} {
		if c.aliases[name] == nil {
			c.aliases[name] = &catalogueAlias{Expression: expression, Source: "Blizzard metadata primitive"}
		}
	}
}
func (c *catalogue) bindGlobal(name, typeName, source string) {
	parts := strings.Split(name, ".")
	if len(parts) == 1 {
		c.setGlobal(name, &catalogueFunction{TypeName: typeName, Source: source})
		return
	}
	parent := parts[0] + "_Namespace"
	c.setGlobal(parts[0], &catalogueFunction{TypeName: parent, Source: source})
	record := c.record(parent, source)
	if len(parts) != 2 {
		c.errors = append(c.errors, source+": unsupported global nesting "+name)
		return
	}
	record.Fields[parts[1]] = &catalogueField{Resolved: stringPointer(ident(typeName)), Source: source}
}
func scriptName(name string) string {
	direct := map[string]string{"SimpleCheckboxAPI": "CheckButton", "SimpleHTMLAPI": "SimpleHTML", "SimpleAnimAPI": "Animation", "SimpleAnimGroupAPI": "AnimationGroup", "SimpleAnimatableObjectAPI": "AnimatableObject", "SimpleFrameScriptObjectAPI": "FrameScriptObject", "SimpleControlPointAPI": "ControlPoint", "SimpleTextureBaseAPI": "TextureBase", "SimpleScriptRegionAPI": "ScriptRegion", "SimpleScriptRegionResizingAPI": "ScriptRegionResizing", "SimpleObjectAPI": "Object", "SimpleRegionAPI": "Region", "SimpleMovieAPI": "MovieFrame", "MinimapFrameAPI": "Minimap", "FrameAPITooltip": "GameTooltip", "FrameAPIModelSceneFrame": "ModelScene", "FrameAPIModelSceneFrameActor": "ModelSceneFrameActor", "FrameAPIModelSceneFrameActorBase": "ModelSceneFrameActorBase", "FrameAPIArchaeologyDigSiteFrame": "ArchaeologyDigSiteFrame", "FrameAPIQuestPOI": "QuestPOIFrame", "FrameAPIScenarioPOI": "ScenarioPOIFrame", "LuaDurationObjectAPI": "DurationObject", "LuaCurveObjectAPI": "CurveObject", "LuaColorCurveObjectAPI": "ColorCurveObject", "LuaCurveObjectBaseAPI": "CurveObjectBase", "UnitHealPredictionCalculatorAPI": "UnitHealPredictionCalculator", "AbbreviateConfigAPI": "AbbreviateConfig"}
	if v, ok := direct[name]; ok {
		return v
	}
	if strings.HasPrefix(name, "SimpleAnim") {
		return strings.TrimSuffix(strings.TrimPrefix(name, "SimpleAnim"), "API")
	}
	if strings.HasPrefix(name, "Simple") {
		return strings.TrimSuffix(strings.TrimPrefix(name, "Simple"), "API")
	}
	if strings.HasPrefix(name, "FrameAPI") {
		return strings.TrimPrefix(name, "FrameAPI")
	}
	return strings.TrimSuffix(name, "API")
}
func (c *catalogue) loadNative() {
	lock := c.sourceLocks()
	for _, key := range sortedKeys(lock) {
		for _, file := range lock[key].Files {
			name := file.Name
			if key != "blizzard" && key != "questie" || key == "blizzard" && !strings.HasSuffix(name, ".lua") {
				disposition := "verification-evidence"
				if key == "blizzard" {
					disposition = "registration-index"
				}
				c.files = append(c.files, ManifestFile{Source: key + ":" + name, Disposition: disposition})
			}
		}
	}
	docs, e := c.project.readNativeDocuments()
	if e != nil {
		reject(e.Error())
	}
	sort.SliceStable(docs, func(i, j int) bool { return docs[i].Name < docs[j].Name })
	c.nativeDocs = docs
	for _, document := range docs {
		doc := document.Metadata
		source := "blizzard:" + document.Name
		c.files = append(c.files, ManifestFile{Source: source, Disposition: "converted"})
		for _, table := range doc.Tables {
			name, kind := table.Name, table.Type
			if c.nativeNames[name] {
				reject(fmt.Sprintf("%s: duplicate native type %s", source, name))
			}
			c.nativeNames[name] = true
			switch kind {
			case "Enumeration":
				delete(c.defs, name)
				if name != "UnitTokenType" {
					c.aliases[name] = &catalogueAlias{Expression: "number", Source: source}
				}
				c.aliases["Enum."+name] = &catalogueAlias{Expression: "number", Source: source}
				record := c.record("Enum."+name+"_Values", source)
				record.Fields = map[string]*catalogueField{}
				for _, field := range table.Fields {
					record.Fields[field.Name] = &catalogueField{Type: "number", Source: source}
				}
				c.bindGlobal("Enum."+name, "Enum."+name+"_Values", source)
				metadata := map[string]*catalogueField{}
				for _, key := range table.EnumMetadata {
					metadata[key] = &catalogueField{Type: "number", Source: source}
				}
				if len(metadata) > 0 {
					metaName := "Enum." + name + "Meta_Values"
					c.record(metaName, source).Fields = metadata
					c.bindGlobal("Enum."+name+"Meta", metaName, source)
				}
			case "Structure":
				record := c.record(name, source)
				record.Kind, record.Source, record.Bases = "record", source, []string{}
				record.Fields = map[string]*catalogueField{}
				for _, field := range table.Fields {
					record.Fields[field.Name] = &catalogueField{Native: &field, Source: source}
				}
			case "CallbackType":
				c.aliases[name] = &catalogueAlias{Signature: functionFromNative(table.Callback(), source, false), Source: source}
			case "Constants":
				record := c.record("Constants."+name+"_Values", source)
				record.Fields = map[string]*catalogueField{}
				for _, field := range table.Fields {
					record.Fields[field.Name] = &catalogueField{Native: &field, Source: source}
				}
				c.bindGlobal("Constants."+name, "Constants."+name+"_Values", source)
			default:
				c.errors = append(c.errors, source+": unsupported documentation table "+kind)
			}
		}
		namespace, owner := doc.Namespace, ""
		if doc.Type == "ScriptObject" {
			owner = scriptName(doc.Name)
			record := c.record(owner, source, "interface")
			record.NativeHandle, record.NativeObject = true, doc.ObjectType == "Userdata"
			alias := strings.TrimSuffix(doc.Name, "API")
			if c.aliases[alias] == nil {
				c.aliases[alias] = &catalogueAlias{Expression: owner, Source: source}
			}
		} else if namespace != "" {
			owner = namespace
			c.record(owner, source)
			c.setGlobal(namespace, &catalogueFunction{TypeName: owner, Source: source})
			if doc.Environment == "SecureOnly" {
				c.secureGlobals[namespace] = true
			}
		}
		for _, function := range doc.Functions {
			name := function.Name
			methodOwner := owner
			if doc.Type == "ScriptObject" {
				methodOwner = nativeMethodOwner(doc.Name, name)
				if methodOwner != owner {
					record := c.record(methodOwner, source, "interface")
					record.Kind = "interface"
					record.NativeHandle, record.NativeObject = true, doc.ObjectType == "Userdata"
				}
			}
			owner := methodOwner
			symbol := name
			if owner != "" {
				symbol = owner + "." + name
			}
			if c.nativeFunctions[symbol] {
				reject(fmt.Sprintf("%s: duplicate native function %s", source, symbol))
			}
			c.nativeFunctions[symbol] = true
			item := functionFromNative(function, source, doc.Type == "ScriptObject")
			if item.Native.Policy.Environment == nil {
				item.Native.Policy.Environment = stringPointer(firstNonempty(doc.Environment, "All"))
			}
			if stringValue(item.Native.Policy.Environment) == "SecureOnly" {
				if owner != "" && doc.Environment != "SecureOnly" {
					reject(fmt.Sprintf("%s: %s: mixed-visibility methods need a separate public namespace projection", source, symbol))
				}
				if owner == "" {
					c.secureGlobals[name] = true
				}
			}
			if owner != "" {
				c.defs[owner].Methods[name] = item
			} else {
				c.setGlobal(name, item)
			}
			c.symbols = append(c.symbols, ManifestSymbol{Name: symbol, Source: source, Function: &function, Restrictions: item.Native.Policy.restrictions()})
		}
		for _, event := range doc.Events {
			c.symbols = append(c.symbols, ManifestSymbol{Name: firstNonempty(event.LiteralName, event.Name), Source: source, Event: &event})
		}
	}
	for _, doc := range c.nativeDocs {
		for _, predicate := range doc.Metadata.Predicates {
			if predicate.Type == "Precondition" && containsString([]string{"ReturnNothing", "ReturnWithError"}, predicate.FailureMode) {
				c.absencePreconditions[predicate.Name] = true
			}
		}
	}
	c.absencePreconditions["RequiresNonSecretAura"] = true
}
func (c *catalogue) mayReturnNothing(item *catalogueFunction) bool {
	if item == nil || item.Native == nil {
		return false
	}
	policy := item.Native.Policy
	if enabled(policy.MayReturnNothing) {
		return true
	}
	for name := range c.absencePreconditions {
		if policy.precondition(name) {
			return true
		}
	}
	return false
}
func (c *catalogue) primitive(name string) string { return c.primitiveSeen(name, map[string]bool{}) }
func (c *catalogue) primitiveSeen(name string, seen map[string]bool) string {
	switch name {
	case "number", "integer":
		return "number"
	case "any":
		return "unknown"
	case "string", "boolean", "unknown":
		return name
	}
	raw := ""
	if alias := c.aliases[name]; alias != nil {
		raw = alias.Expression
	}
	if raw != "" && !seen[raw] && plainIdentifier.MatchString(raw) {
		seen[name] = true
		return c.primitiveSeen(raw, seen)
	}
	return ""
}
func (c *catalogue) isHandle(name string) bool { return c.handleSeen(name, map[string]bool{}) }
func (c *catalogue) handleSeen(name string, seen map[string]bool) bool {
	if seen[name] || c.primitive(name) != "" {
		return false
	}
	record := c.defs[name]
	if record != nil && record.NativeHandle {
		return true
	}
	seen[name] = true
	alias := ""
	if value := c.aliases[name]; value != nil {
		alias = value.Expression
	}
	if alias != "" && c.handleSeen(alias, seen) {
		return true
	}
	for _, base := range definitionBases(record) {
		if c.handleSeen(base, seen) {
			return true
		}
	}
	return false
}
func definitionBases(v *catalogueDefinition) []string {
	if v == nil {
		return nil
	}
	return v.Bases
}
func definitionKind(v *catalogueDefinition) string {
	if v == nil {
		return ""
	}
	return v.Kind
}
func (c *catalogue) resultSecret(item *catalogueFunction) bool {
	return item != nil && (!item.returnsNeverSecret()) && ((item.Native != nil && item.Native.Policy.secretResult()) || item.SecretEffect)
}
func (item *catalogueFunction) returnsNeverSecret() bool {
	return item != nil && item.Native != nil && enabled(item.Native.Policy.ReturnsNeverSecret)
}
func (c *catalogue) nativeType(item NativeParameter, secret bool) (string, error) {
	return captureString(func() string { return c.nativeTypeValue(item, secret) })
}
func (c *catalogue) nativeTypeValue(item NativeParameter, secret bool) string {
	raw := item.Type
	secret = !enabled(item.NeverSecret) && (secret || item.secretResult())
	if raw == "table" && (secret || enabled(item.ConditionalSecretContents)) && !enabled(item.NeverSecretContents) {
		result := "OpaqueValue"
		if enabled(item.Nilable) {
			result += " | nil"
		}
		return result
	}
	result := ""
	if raw == "table" {
		inner := item.InnerType
		if inner == "" {
			result = "{unknown: unknown}"
		} else {
			value := c.nativeTypeValue(NativeParameter{Type: inner}, !enabled(item.NeverSecretContents) && (secret || enabled(item.ConditionalSecretContents)))
			key := ""
			if item.KeyType != "" {
				key = c.nativeTypeValue(NativeParameter{Type: item.KeyType}, false) + ": "
			}
			result = "{" + key + value + "}"
		}
	} else {
		result = c.resolve(raw)
	}
	if secret {
		switch c.primitive(raw) {
		case "number":
			result = "W.NumberLike"
		case "string":
			result = "W.TextLike"
		case "boolean":
			result = "boolean | OpaqueValue"
		case "unknown":
			result = "unknown"
		default:
			if c.isHandle(raw) {
				result = "OpaqueValue"
			} else if !containsString([]string{"table", "Secret", "LuaValueReference", "LuaValueVariant"}, raw) {
				known := false
				for name, v := range c.defs {
					if result == ident(name) && containsString([]string{"record", "interface", "converted_record"}, v.Kind) {
						known = true
					}
				}
				if !known {
					result = "OpaqueValue"
				}
			}
		}
	}
	if enabled(item.Nilable) {
		result += " | nil"
	}
	return result
}

type frameXMLSupplement struct {
	Description string               `json:"description"`
	Aliases     map[string]string    `json:"aliases"`
	Classes     []frameXMLClass      `json:"classes"`
	Globals     map[string]string    `json:"globals"`
	Evidence    []RepresentationNote `json:"evidence"`
}
type frameXMLClass struct {
	Name    string            `json:"name"`
	Bases   []string          `json:"bases"`
	Fields  map[string]string `json:"fields"`
	Methods map[string]string `json:"methods"`
}

func (c *catalogue) loadFrameXML() {
	var supplement frameXMLSupplement
	if e := decodeJSON(mustFile(filepath.Join(c.project.Root, "scripts/wow_api/framexml.json")), &supplement); e != nil {
		reject(e.Error())
	}
	source := "reviewed FrameXML supplement"
	for _, name := range sortedKeys(supplement.Aliases) {
		c.aliases[name] = &catalogueAlias{Expression: supplement.Aliases[name], Source: source}
	}
	for _, cls := range supplement.Classes {
		record := c.record(cls.Name, source, "interface")
		for _, base := range cls.Bases {
			if !containsString(record.Bases, base) {
				record.Bases = append(record.Bases, base)
			}
		}
	}
	for _, cls := range supplement.Classes {
		owner := cls.Name
		record := c.defs[owner]
		for _, name := range sortedKeys(cls.Fields) {
			record.Fields[name] = &catalogueField{Type: cls.Fields[name], Source: source}
		}
		for _, name := range sortedKeys(cls.Methods) {
			expression := cls.Methods[name]
			callback := c.tealValue(expression, owner, nil)
			signature := c.defs[callback].Signature
			receiver := "function(" + ident(owner)
			pattern := regexp.MustCompile(`^fun\(\s*self\s*:\s*` + regexp.QuoteMeta(owner) + `\s*[,)]`)
			if !pattern.MatchString(expression) || !strings.HasPrefix(signature, receiver) || len(signature) <= len(receiver) || !strings.ContainsRune(",)", rune(signature[len(receiver)])) {
				reject(fmt.Sprintf("%s.%s: expected an explicit owning receiver", owner, name))
			}
			signature = "function(self" + signature[len(receiver):]
			record.Methods[name] = &catalogueFunction{ResolvedSignature: &signature, Source: source}
		}
	}
	for _, name := range sortedKeys(supplement.Globals) {
		c.setGlobal(name, &catalogueFunction{ResolvedSignature: stringPointer(c.tealValue(supplement.Globals[name], "", nil)), Source: source})
	}
	c.notes = append(c.notes, supplement.Evidence...)
}
func (p Project) Generate(ctx context.Context) (artifacts Artifacts, err error) {
	defer captureError(&err)
	if e := ctx.Err(); e != nil {
		return Artifacts{}, e
	}
	c := newCatalogue(p)
	c.loadQuestie()
	c.loadNative()
	c.loadFrameXML()
	c.applyOverrides()
	projection, e := c.finalize(ctx)
	if e != nil {
		return Artifacts{}, e
	}
	outputs := projection.render()
	manifest := c.manifest()
	preserved, e := p.AuditMetadata(&manifest)
	if e != nil {
		return Artifacts{}, e
	}
	if e = RequireKnownMetadata(preserved); e != nil {
		return Artifacts{Files: map[string][]byte{"evidence/generated/wow_native_metadata.json": mustJSONUTF8(preserved)}}, e
	}
	crosscheck, e := p.crosscheck(ctx, c, manifest, projection, outputs)
	if e != nil {
		return Artifacts{}, e
	}
	files := map[string][]byte{}
	for name, b := range outputs {
		files[filepath.ToSlash(filepath.Join("types/wow", name))] = b
	}
	manifestBytes, e := jsonPretty(manifest)
	if e != nil {
		return Artifacts{}, e
	}
	files["evidence/generated/wow_manifest.json"] = manifestBytes
	crosscheckBytes, e := jsonPretty(crosscheck)
	if e != nil {
		return Artifacts{}, e
	}
	files["evidence/generated/wow_tsm_crosscheck.json"] = crosscheckBytes
	files["evidence/generated/wow_native_metadata.json"] = mustJSONUTF8(preserved)
	return Artifacts{Files: files, Statistics: manifest.Statistics}, nil
}
func mustJSONUTF8[T any](v T) []byte {
	b, e := jsonPrettyUTF8(v)
	if e != nil {
		reject(fmt.Sprintf("%s", e))
	}
	return b
}
func mustFile(path string) []byte {
	b, e := os.ReadFile(path)
	if e != nil {
		reject(fmt.Sprintf("%s", e))
	}
	return b
}
