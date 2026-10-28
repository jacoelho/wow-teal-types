package wowapi

import (
	"cmp"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"
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
func requireStringJSON(v any) string {
	b, e := jsonDefault(v)
	if e != nil {
		reject(fmt.Sprintf("JSON encoding: %s", e))
	}
	return string(b)
}

type catalogue struct {
	defs                 map[string]*catalogueDefinition
	aliases              map[string]*catalogueAlias
	globals              map[string]*catalogueFunction
	errors               []string
	inputStats           CatalogueStatistics
	secretStructs        map[string]bool
	resolving            map[string]bool
	secureGlobals        map[string]bool
	absencePreconditions map[string]bool
	abiTuples            map[string][]string
	defOrder             []string
	globalOrder          []string
	// defNames maps declared identifiers back to definition names. It is
	// derived from defs and dropped whenever defs changes.
	defNames map[string]string
}

func newCatalogue() *catalogue {
	return &catalogue{
		defs:                 map[string]*catalogueDefinition{},
		aliases:              map[string]*catalogueAlias{},
		globals:              map[string]*catalogueFunction{},
		errors:               []string{},
		secretStructs:        map[string]bool{},
		resolving:            map[string]bool{},
		secureGlobals:        map[string]bool{},
		absencePreconditions: map[string]bool{},
		abiTuples:            map[string][]string{},
		defOrder:             []string{},
		globalOrder:          []string{},
	}
}
func (c *catalogue) setDef(name string, item *catalogueDefinition) {
	if _, ok := c.defs[name]; !ok {
		c.defOrder = append(c.defOrder, name)
	}
	c.defs[name] = item
	c.defNames = nil
}
func (c *catalogue) deleteDef(name string) {
	delete(c.defs, name)
	c.defNames = nil
}

// defNamed returns the definition declared as identifier.
func (c *catalogue) defNamed(identifier string) (string, bool) {
	if c.defNames == nil {
		c.defNames = make(map[string]string, len(c.defs))
		for name := range c.defs {
			c.defNames[ident(name)] = name
		}
	}
	name, ok := c.defNames[identifier]
	return name, ok
}
func (c *catalogue) setGlobal(name string, item *catalogueFunction) {
	if _, ok := c.globals[name]; !ok {
		c.globalOrder = append(c.globalOrder, name)
	}
	c.globals[name] = item
}

// record returns the named definition, creating it with kind when absent.
func (c *catalogue) record(name, source string, kind definitionKind) *catalogueDefinition {
	delete(c.aliases, name)
	if v := c.defs[name]; v != nil {
		return v
	}
	v := &catalogueDefinition{Kind: kind, Source: source, Fields: map[string]*catalogueField{}, Methods: map[string]*catalogueFunction{}, Bases: []string{}}
	c.setDef(name, v)
	return v
}
func (c *catalogue) literal(values []string) string {
	key := "Literal_" + sha256hex([]byte(requireStringJSON(values)))[:12]
	if c.defs[key] == nil {
		c.setDef(key, &catalogueDefinition{Kind: kindEnum, Values: values, Source: "interned LuaLS string literals"})
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
		c.setDef(key, &catalogueDefinition{Kind: kindConvertedRecord, Fields: mapped, Source: "interned LuaLS record"})
	}
	return ident(key)
}
func (c *catalogue) resolve(name string, position typePosition) string {
	if name == "self" {
		unsupported(name, "self requires receiver context")
	}
	if name == "any" || name == "unknown" {
		return position.untyped()
	}
	if slices.Contains([]string{"number", "integer", "string", "boolean", "nil", "thread"}, name) {
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
		return c.tealValue(alias.Expression, "", nil, position)
	}
	if c.defs[name] != nil || c.aliases[name] != nil {
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
func (c *catalogue) tealValue(expression, owner string, generics []string, position typePosition) string {
	converter := typeConverter{raw: expression, self: ident(owner), position: position, literal: c.literal, record: c.anonymous}
	converter.resolve = func(name string, position typePosition) string {
		if slices.Contains(generics, name) {
			return ident(name)
		}
		return c.resolve(name, position)
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
			c.setDef(key, &catalogueDefinition{Kind: kindCallback, Signature: signature, Generics: usedGenerics, Source: "named LuaLS callback"})
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
		// An untyped member absorbs the union: Data already has table and
		// function members, so no other table or callback could discriminate.
		if slices.Contains(values, "any") {
			return "any"
		}
		if slices.Contains(values, dataType) {
			if slices.Contains(values, "nil") {
				return dataType + " | nil"
			}
			return dataType
		}
		named := map[string]*catalogueDefinition{}
		for name, value := range c.defs {
			named[ident(name)] = value
		}
		enums := []string{}
		for _, value := range values {
			if named[value] != nil && named[value].Kind == kindEnum {
				enums = append(enums, value)
			}
		}
		if len(enums) > 0 {
			next := []string{}
			for _, value := range values {
				if !slices.Contains(enums, value) {
					next = append(next, value)
				}
			}
			values = next
			if !slices.Contains(values, "string") {
				members := []string{}
				for _, value := range enums {
					members = append(members, named[value].Values...)
				}
				members = uniqueStrings(members)
				slices.Sort(members)
				values = append(values, c.literal(members))
			}
		}
		tables := []string{}
		for _, value := range values {
			record := named[value]
			if strings.HasPrefix(value, "{") || record != nil && record.Kind.isTableLike() && !record.Opaque {
				tables = append(tables, value)
			}
		}
		if len(tables) > 1 {
			ordered := append([]string{}, tables...)
			slices.Sort(ordered)
			replacement := "TableVariants_" + sha256hex([]byte(strings.Join(ordered, "|")))[:12]
			if c.defs[replacement] == nil {
				c.setDef(replacement, &catalogueDefinition{Kind: kindInterface, Source: "opaque result surface of " + strings.Join(tables, " | "), Fields: map[string]*catalogueField{}, Methods: map[string]*catalogueFunction{}, Bases: []string{}})
			}
			for _, table := range tables {
				record := named[table]
				if record == nil {
					reject(fmt.Sprintf("unknown table union member %s", table))
				}
				if !slices.Contains(record.Bases, replacement) {
					record.Bases = append(record.Bases, replacement)
				}
			}
			next := []string{}
			for _, v := range values {
				if !slices.Contains(tables, v) {
					next = append(next, v)
				}
			}
			values = append(next, replacement)
		}
		return strings.Join(values, " | ")
	}
	return converter.walk(expression, 0)
}
func (c *catalogue) loadQuestie(docs []annotationDocument) {
	for _, doc := range docs {
		source := "questie:" + doc.Path
		if strings.HasPrefix(doc.Path, "Lua/") {
			continue
		}
		for _, alias := range doc.Aliases {
			alternatives := []string{}
			if alias.Type != "" {
				alternatives = append(alternatives, alias.Type)
			}
			for _, v := range alias.Alternatives {
				alternatives = append(alternatives, strings.TrimRight(v, "| "))
			}
			expression := strings.Join(alternatives, " | ")
			if expression == "" {
				c.errors = append(c.errors, fmt.Sprintf("%s:%d: empty alias %s", source, alias.Line, alias.Name))
				continue
			}
			if strings.HasSuffix(expression, "|") {
				expression = strings.TrimRight(expression, "| ")
			}
			c.aliases[alias.Name] = &catalogueAlias{Expression: expression, Source: source}
		}
		for _, cls := range doc.Classes {
			record := c.record(cls.Name, source, kindInterface)
			for _, base := range cls.Bases {
				if !slices.Contains(record.Bases, base) {
					record.Bases = append(record.Bases, base)
				}
			}
			for _, field := range cls.Fields {
				record.Fields[field.Name] = &catalogueField{Type: field.Type}
			}
		}
		for _, enum := range doc.Enums {
			numeric := true
			for _, v := range enum.Values {
				if !v.Numeric {
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
			record := c.record(recordName, source, kindRecord)
			for _, v := range enum.Values {
				typ := "string"
				if v.Numeric {
					typ = "number"
				}
				record.Fields[v.Name] = &catalogueField{Type: typ}
			}
			c.bindGlobal(name, recordName, source)
		}
		for _, function := range doc.Functions {
			if function.Local {
				continue
			}
			c.inputStats.QuestieDeclarations++
			item := functionFromAnnotation(function, source)
			owner := stringValue(function.Owner)
			name := function.Name
			if owner == "Region" {
				owner = nativeMethodOwner("SimpleRegionAPI", name)
			}
			if owner != "" {
				methods := c.record(owner, source, kindRecord).Methods
				if previous := methods[name]; previous != nil {
					item.Variants = append(append([]*catalogueFunction{}, previous.Variants...), previous)
				}
				methods[name] = item
			} else {
				c.setGlobal(name, item)
			}
		}
		for _, binding := range doc.Bindings {
			if binding.Local || !binding.TopLevel {
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
					record := c.record(name, source, kindInterface)
					for _, m := range strings.Split(mixins, ",") {
						record.Bases = append(record.Bases, strings.TrimSpace(m))
					}
				case strings.HasPrefix(value, "{}"):
					target = name
					c.record(name, source, kindRecord)
				case numericTypePattern.MatchString(value):
					target = "number"
				}
			}
			if target != "" && c.globals[name] == nil {
				c.bindGlobal(name, target, source)
			}
		}
	}
	for name, expression := range map[string]string{"bool": "boolean", "cstring": "string", "luaIndex": "number", "UnitTokenType": "string", "UnitTokenVariant": "string", "LuaFunctionReference": "function", "Seconds": "number"} {
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
	record := c.record(parent, source, kindRecord)
	if len(parts) != 2 {
		c.errors = append(c.errors, source+": unsupported global nesting "+name)
		return
	}
	record.Fields[parts[1]] = &catalogueField{Resolved: new(ident(typeName))}
}

// scriptReceivers names the receivers whose documentation group does not
// follow the Simple*/FrameAPI*/*API naming convention.
var scriptReceivers = map[string]string{"SimpleCheckboxAPI": "CheckButton", "SimpleHTMLAPI": "SimpleHTML", "SimpleAnimAPI": "Animation", "SimpleAnimGroupAPI": "AnimationGroup", "SimpleAnimatableObjectAPI": "AnimatableObject", "SimpleFrameScriptObjectAPI": "FrameScriptObject", "SimpleControlPointAPI": "ControlPoint", "SimpleTextureBaseAPI": "TextureBase", "SimpleScriptRegionAPI": "ScriptRegion", "SimpleScriptRegionResizingAPI": "ScriptRegionResizing", "SimpleObjectAPI": "Object", "SimpleRegionAPI": "Region", "SimpleMovieAPI": "MovieFrame", "MinimapFrameAPI": "Minimap", "FrameAPITooltip": "GameTooltip", "FrameAPIModelSceneFrame": "ModelScene", "FrameAPIModelSceneFrameActor": "ModelSceneFrameActor", "FrameAPIModelSceneFrameActorBase": "ModelSceneFrameActorBase", "FrameAPIArchaeologyDigSiteFrame": "ArchaeologyDigSiteFrame", "FrameAPIQuestPOI": "QuestPOIFrame", "FrameAPIScenarioPOI": "ScenarioPOIFrame", "LuaDurationObjectAPI": "DurationObject", "LuaCurveObjectAPI": "CurveObject", "LuaColorCurveObjectAPI": "ColorCurveObject", "LuaCurveObjectBaseAPI": "CurveObjectBase", "UnitHealPredictionCalculatorAPI": "UnitHealPredictionCalculator", "AbbreviateConfigAPI": "AbbreviateConfig"}

func scriptName(name string) string {
	if v, ok := scriptReceivers[name]; ok {
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
func (c *catalogue) loadNative(docs []nativeDocument) {
	nativeNames, nativeFunctions := map[string]bool{}, map[string]bool{}
	for _, document := range docs {
		doc := document.Metadata
		c.inputStats.NativeFunctions += len(doc.Functions)
		c.inputStats.NativeEvents += len(doc.EventNames)
		for _, predicate := range doc.Predicates {
			if predicate.Type == "Precondition" && slices.Contains([]string{"ReturnNothing", "ReturnWithError"}, predicate.FailureMode) {
				c.absencePreconditions[predicate.Name] = true
			}
		}
		source := "blizzard:" + document.Name
		for _, table := range doc.Tables {
			name, kind := table.Name, table.Type
			if nativeNames[name] {
				reject(fmt.Sprintf("%s: duplicate native type %s", source, name))
			}
			nativeNames[name] = true
			switch kind {
			case "Enumeration":
				c.deleteDef(name)
				if name != "UnitTokenType" {
					c.aliases[name] = &catalogueAlias{Expression: "number", Source: source}
				}
				c.aliases["Enum."+name] = &catalogueAlias{Expression: "number", Source: source}
				record := c.record("Enum."+name+"_Values", source, kindRecord)
				record.Fields = map[string]*catalogueField{}
				for _, field := range table.Fields {
					record.Fields[field.Name] = &catalogueField{Type: "number"}
				}
				c.bindGlobal("Enum."+name, "Enum."+name+"_Values", source)
				metadata := map[string]*catalogueField{}
				for _, key := range table.EnumMetadata {
					metadata[key] = &catalogueField{Type: "number"}
				}
				if len(metadata) > 0 {
					metaName := "Enum." + name + "Meta_Values"
					c.record(metaName, source, kindRecord).Fields = metadata
					c.bindGlobal("Enum."+name+"Meta", metaName, source)
				}
			case "Structure":
				record := c.record(name, source, kindRecord)
				record.Kind, record.Source, record.Bases = kindRecord, source, []string{}
				record.Fields = map[string]*catalogueField{}
				for _, field := range table.Fields {
					record.Fields[field.Name] = &catalogueField{Native: &field}
				}
			case "CallbackType":
				c.aliases[name] = &catalogueAlias{Signature: functionFromNative(table.callback(), source, false), Source: source}
			case "Constants":
				record := c.record("Constants."+name+"_Values", source, kindRecord)
				record.Fields = map[string]*catalogueField{}
				for _, field := range table.Fields {
					record.Fields[field.Name] = &catalogueField{Native: &field}
				}
				c.bindGlobal("Constants."+name, "Constants."+name+"_Values", source)
			default:
				c.errors = append(c.errors, source+": unsupported documentation table "+kind)
			}
		}
		namespace, owner := doc.Namespace, ""
		if doc.Type == "ScriptObject" {
			owner = scriptName(doc.Name)
			record := c.record(owner, source, kindInterface)
			record.NativeHandle, record.NativeObject = true, doc.ObjectType == "Userdata"
			alias := strings.TrimSuffix(doc.Name, "API")
			if c.aliases[alias] == nil {
				c.aliases[alias] = &catalogueAlias{Expression: owner, Source: source}
			}
		} else if namespace != "" {
			owner = namespace
			c.record(owner, source, kindRecord)
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
					record := c.record(methodOwner, source, kindInterface)
					record.Kind = kindInterface
					record.NativeHandle, record.NativeObject = true, doc.ObjectType == "Userdata"
				}
			}
			owner := methodOwner
			symbol := name
			if owner != "" {
				symbol = owner + "." + name
			}
			if nativeFunctions[symbol] {
				reject(fmt.Sprintf("%s: duplicate native function %s", source, symbol))
			}
			nativeFunctions[symbol] = true
			item := functionFromNative(function, source, doc.Type == "ScriptObject")
			if item.Native.Policy.Environment == nil {
				item.Native.Policy.Environment = new(cmp.Or(doc.Environment, "All"))
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
		}
	}
	c.absencePreconditions["RequiresNonSecretAura"] = true
}
func (c *catalogue) mayReturnNothing(item *catalogueFunction) bool {
	if item == nil || item.Native == nil {
		return false
	}
	policy := item.Native.Policy
	if policy.Flags[nativeMayReturnNothing] {
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
	case "any", "unknown":
		// Untyped: no primitive is known, but the value is not a named type.
		return "any"
	case "string", "boolean":
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
func (c *catalogue) resultSecret(item *catalogueFunction) bool {
	return item != nil && (!item.returnsNeverSecret()) && ((item.Native != nil && item.Native.Policy.secretResult()) || item.SecretEffect)
}
func (item *catalogueFunction) returnsNeverSecret() bool {
	return item != nil && item.Native != nil && item.Native.Policy.Flags[nativeReturnsNeverSecret]
}

// isUserdata reports whether a named type is declared `is userdata`. Teal
// cannot discriminate a union of two userdata types, so such a type can only
// be wrapped as Secret<T>, never as MaybeSecret<T>.
func (c *catalogue) isUserdata(name string) bool {
	return c.isHandle(name) || c.defs[name] != nil && c.defs[name].Opaque
}

// nativeTypeValue declares one native parameter or result at position.
func (c *catalogue) nativeTypeValue(item nativeParameter, secret bool, position typePosition) string {
	raw := item.Type
	secret = !enabled(item.NeverSecret) && (secret || item.secretResult())
	result := ""
	switch {
	case raw == "table" && (secret || enabled(item.ConditionalSecretContents)) && !enabled(item.NeverSecretContents):
		// Possibly secret contents stay marked after canaccesstable unwraps the table.
		result = secretTableType
		if position == inputPosition {
			result = "any"
		}
	case raw == "table" && item.InnerType == "":
		result = position.untyped()
	case raw == "table":
		value := c.nativeTypeValue(nativeParameter{Type: item.InnerType}, !enabled(item.NeverSecretContents) && (secret || enabled(item.ConditionalSecretContents)), position)
		key := ""
		if item.KeyType != "" {
			key = c.nativeTypeValue(nativeParameter{Type: item.KeyType}, false, position) + ": "
		}
		result = "{" + key + value + "}"
	default:
		result = c.resolve(raw, position)
		if secret {
			result = c.secretNative(raw, result, position)
		}
	}
	if enabled(item.Nilable) && result != "any" {
		result += " | nil"
	}
	return result
}

// secretNative declares a possibly secret native value of raw type, already
// resolved to result.
func (c *catalogue) secretNative(raw, result string, position typePosition) string {
	switch primitive := c.primitive(raw); primitive {
	case "number", "string", "boolean":
		return "W.MaybeSecret<" + primitive + ">"
	case "any":
		if position == inputPosition {
			return "any"
		}
		return secretDataType
	}
	if c.isUserdata(raw) {
		return "W.Secret<" + result + ">"
	}
	// Secret structures project their own fields instead.
	if name, ok := c.defNamed(result); ok && c.defs[name].Kind.isTableLike() {
		return result
	}
	return "W.MaybeSecret<" + result + ">"
}

type frameXMLSupplement struct {
	Description string            `json:"description"`
	Aliases     map[string]string `json:"aliases"`
	Classes     []frameXMLClass   `json:"classes"`
	Globals     map[string]string `json:"globals"`
}
type frameXMLClass struct {
	Name    string            `json:"name"`
	Bases   []string          `json:"bases"`
	Fields  map[string]string `json:"fields"`
	Methods map[string]string `json:"methods"`
}

func (c *catalogue) loadFrameXML(supplement frameXMLSupplement) (err error) {
	defer recoverConversion(&err)
	source := "reviewed FrameXML supplement"
	for _, name := range slices.Sorted(maps.Keys(supplement.Aliases)) {
		c.aliases[name] = &catalogueAlias{Expression: supplement.Aliases[name], Source: source}
	}
	for _, cls := range supplement.Classes {
		record := c.record(cls.Name, source, kindInterface)
		for _, base := range cls.Bases {
			if !slices.Contains(record.Bases, base) {
				record.Bases = append(record.Bases, base)
			}
		}
	}
	for _, cls := range supplement.Classes {
		owner := cls.Name
		record := c.defs[owner]
		for _, name := range slices.Sorted(maps.Keys(cls.Fields)) {
			record.Fields[name] = &catalogueField{Type: cls.Fields[name]}
		}
		for _, name := range slices.Sorted(maps.Keys(cls.Methods)) {
			signature, err := c.frameXMLMethod(cls.Methods[name], owner)
			if err != nil {
				return fmt.Errorf("%s.%s: %w", owner, name, err)
			}
			record.Methods[name] = &catalogueFunction{ResolvedSignature: &signature, Source: source}
		}
	}
	for _, name := range slices.Sorted(maps.Keys(supplement.Globals)) {
		c.setGlobal(name, &catalogueFunction{ResolvedSignature: new(c.tealValue(supplement.Globals[name], "", nil, outputPosition)), Source: source})
	}
	return nil
}
func (c *catalogue) frameXMLMethod(expression, owner string) (signature string, err error) {
	defer recoverConversion(&err)
	pattern := regexp.MustCompile(`^fun\(\s*self\s*:\s*` + regexp.QuoteMeta(owner) + `\s*[,)]`)
	if !pattern.MatchString(expression) {
		reject("expected an explicit owning receiver")
	}
	// The method is a value the game provides: its parameters are inputs.
	callback := c.defs[c.tealValue(expression, owner, nil, outputPosition)]
	if callback == nil || callback.Kind != kindCallback {
		reject("expected a method function")
	}
	signature = callback.Signature
	receiver := "function(" + ident(owner)
	if !strings.HasPrefix(signature, receiver) || len(signature) <= len(receiver) || !strings.ContainsRune(",)", rune(signature[len(receiver)])) {
		reject("expected an explicit owning receiver")
	}
	return "function(self" + signature[len(receiver):], nil
}

func (c *catalogue) statistics() CatalogueStatistics {
	stats := c.inputStats
	stats.TypeExports = len(c.defs) + len(c.aliases)
	return stats
}
