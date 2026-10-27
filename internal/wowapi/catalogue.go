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
func restrictions(item M) M {
	out := M{}
	for k, v := range item {
		if strings.HasPrefix(k, "Secret") || strings.HasPrefix(k, "Require") || strings.HasPrefix(k, "Check") || strings.HasPrefix(k, "NeverSecret") || containsString([]string{"Environment", "HasRestrictions", "IsProtectedFunction", "ReturnsNeverSecret", "ConditionalSecret", "ConditionalSecretContents", "ConstSecretAccessor", "SecureHooksAllowed", "AddsForbiddenAspects", "RestrictedForMacroChatMessages"}, k) {
			out[k] = v
		}
	}
	return out
}
func secretResult(item M) bool {
	if truth(item["NeverSecret"]) || truth(item["ReturnsNeverSecret"]) {
		return false
	}
	for k, v := range item {
		if truth(v) && (strings.HasPrefix(k, "SecretWhen") || strings.HasPrefix(k, "SecretReturns") || strings.HasPrefix(k, "SecretIn") || containsString([]string{"SecretValue", "ConditionalSecret", "ConditionalSecretContents"}, k)) {
			return true
		}
	}
	return false
}
func mapCopy(item M) M {
	if item == nil {
		return M{}
	}
	return cloneValue(item).(map[string]any)
}
func withFields(item M, fields M) M {
	out := mapCopy(item)
	for k, v := range fields {
		out[k] = v
	}
	return out
}
func appendField(item M, key string, value any) { item[key] = append(asList(item[key]), value) }
func requireStringJSON(v any) string {
	b, e := jsonDefault(v)
	if e != nil {
		reject("JSON encoding: %s", e)
	}
	return string(b)
}
func mustMap(v M, e error) M {
	if e != nil {
		reject("%s", e)
	}
	return v
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
func getOr(item M, key string, defaultValue any) any {
	v, ok := item[key]
	if !ok {
		return defaultValue
	}
	return v
}

type nativeDocument struct {
	Name     string
	Metadata M
}
type catalogue struct {
	project                                                                                           Project
	defs, aliases, globals                                                                            M
	symbols, files, notes                                                                             []any
	errors                                                                                            []string
	questie                                                                                           []M
	nativeDocs                                                                                        []nativeDocument
	used, secretStructs, nativeNames, nativeFunctions, resolving, secureGlobals, absencePreconditions map[string]bool
	abiTuples, excludedMethods                                                                        M
	defOrder, globalOrder                                                                             []string
}

func newCatalogue(p Project) *catalogue {
	return &catalogue{project: p, defs: M{"OpaqueValue": M{"kind": "external", "source": "opaque native/LuaLS unknown carrier"}}, aliases: M{}, globals: M{}, symbols: []any{}, files: []any{}, notes: []any{}, errors: []string{}, questie: []M{}, nativeDocs: []nativeDocument{}, used: map[string]bool{}, secretStructs: map[string]bool{}, nativeNames: map[string]bool{}, nativeFunctions: map[string]bool{}, resolving: map[string]bool{}, secureGlobals: map[string]bool{}, absencePreconditions: map[string]bool{}, abiTuples: M{}, excludedMethods: M{}, defOrder: []string{"OpaqueValue"}, globalOrder: []string{}}
}
func (c *catalogue) setDef(name string, item M) {
	if _, ok := c.defs[name]; !ok {
		c.defOrder = append(c.defOrder, name)
	}
	c.defs[name] = item
}
func (c *catalogue) setGlobal(name string, item M) {
	if _, ok := c.globals[name]; !ok {
		c.globalOrder = append(c.globalOrder, name)
	}
	c.globals[name] = item
}
func (c *catalogue) record(name, source string, kind ...string) M {
	delete(c.aliases, name)
	if value, ok := c.defs[name]; ok {
		return asMap(value)
	}
	k := "record"
	if len(kind) > 0 {
		k = kind[0]
	}
	record := M{"kind": k, "source": source, "fields": M{}, "methods": M{}, "bases": []any{}}
	c.setDef(name, record)
	return record
}
func (c *catalogue) literal(values []string) string {
	key := "Literal_" + sha256hex([]byte(requireStringJSON(toAnyStrings(values))))[:12]
	if _, ok := c.defs[key]; !ok {
		c.setDef(key, M{"kind": "enum", "values": toAnyStrings(values), "source": "interned LuaLS string literals"})
	}
	return ident(key)
}
func (c *catalogue) anonymous(fields []recordField) string {
	rows := []any{}
	mapped := M{}
	for _, f := range fields {
		rows = append(rows, []any{f.name, f.typeName, f.optional})
		mapped[f.name] = f.typeName
	}
	key := "Shape_" + sha256hex([]byte(requireStringJSON(rows)))[:12]
	if _, ok := c.defs[key]; !ok {
		c.setDef(key, M{"kind": "converted_record", "fields": mapped, "source": "interned LuaLS record"})
	}
	return ident(key)
}
func (c *catalogue) resolve(name string) string {
	if name == "self" {
		unsupported(name, "self requires receiver context")
	}
	if containsString([]string{"number", "integer", "string", "boolean", "nil", "any", "thread"}, name) {
		return name
	}
	if c.defs[name] == nil && c.aliases[name] == nil && c.aliases["Enum."+name] != nil {
		c.aliases[name] = M{"expression": "Enum." + name, "source": "Questie numeric enum name"}
	}
	if c.aliases[name] != nil && c.defs[name] == nil {
		if c.resolving[name] {
			unsupported(name, "cyclic type alias")
		}
		c.resolving[name] = true
		defer delete(c.resolving, name)
		alias := asMap(c.aliases[name])
		if sig := asMap(alias["signature"]); sig != nil {
			return c.signatureValue(sig, "")
		}
		return c.tealValue(asString(alias["expression"]), "", nil)
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
		key := "Callback_" + sha256hex([]byte(signature + pythonTuple(generics)))[:12]
		if _, ok := c.defs[key]; !ok {
			c.setDef(key, M{"kind": "callback", "signature": signature, "generics": toAnyStrings(generics), "source": "named LuaLS callback"})
		}
		if len(generics) > 0 {
			return key + "<" + strings.Join(generics, ", ") + ">"
		}
		return key
	}
	converter.union = func(values []string) string {
		all := []string{}
		for _, value := range values {
			all = append(all, splitTop(value, '|')...)
		}
		values = uniqueStrings(all)
		if containsString(values, "any") {
			return "any"
		}
		named := map[string]M{}
		for name, value := range c.defs {
			named[ident(name)] = asMap(value)
		}
		enums := []string{}
		for _, value := range values {
			if asString(named[value]["kind"]) == "enum" {
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
					members = append(members, stringsOf(named[value]["values"])...)
				}
				members = uniqueStrings(members)
				sort.Strings(members)
				values = append(values, c.literal(members))
			}
		}
		tables := []string{}
		for _, value := range values {
			record := named[value]
			if strings.HasPrefix(value, "{") || containsString([]string{"record", "interface", "converted_record"}, asString(record["kind"])) && !truth(record["opaque"]) {
				tables = append(tables, value)
			}
		}
		if len(tables) > 1 {
			replacement := "{any: any}"
			if !containsString(tables, replacement) {
				ordered := append([]string{}, tables...)
				sort.Strings(ordered)
				replacement = "TableVariants_" + sha256hex([]byte(strings.Join(ordered, "|")))[:12]
				if c.defs[replacement] == nil {
					c.setDef(replacement, M{"kind": "interface", "source": "opaque result surface of " + strings.Join(tables, " | "), "fields": M{}, "methods": M{}, "bases": []any{}})
				}
				for _, table := range tables {
					record := named[table]
					if record == nil {
						reject("unknown table union member %s", table)
					}
					if !containsString(stringsOf(record["bases"]), replacement) {
						appendField(record, "bases", replacement)
					}
				}
			}
			c.notes = append(c.notes, M{"expression": expression, "representation": replacement, "rule": "common table surface of source union"})
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
func (c *catalogue) loadQuestie() {
	lock := mustMap(c.project.Lock("sources.json"))
	rows := asList(asMap(lock["questie"])["files"])
	sort.SliceStable(rows, func(i, j int) bool { return asString(asMap(rows[i])["name"]) < asString(asMap(rows[j])["name"]) })
	for _, row := range rows {
		name := asString(asMap(row)["name"])
		if strings.HasSuffix(name, ".lua") {
			doc, e := parseAnnotationFile(c.project.InputPath("questie", name), name)
			if e != nil {
				reject("%s", e)
			}
			c.questie = append(c.questie, doc)
		}
	}
	for _, doc := range c.questie {
		path := asString(doc["path"])
		source := "questie:" + path
		if strings.HasPrefix(path, "Lua/") {
			c.files = append(c.files, M{"source": source, "disposition": "compiler-builtins"})
			continue
		}
		c.files = append(c.files, M{"source": source, "disposition": "converted"})
		for _, raw := range asList(doc["aliases"]) {
			alias := asMap(raw)
			alternatives := []string{}
			if value := asString(alias["type"]); value != "" {
				alternatives = append(alternatives, value)
			}
			for _, v := range asList(alias["alternatives"]) {
				alternatives = append(alternatives, strings.TrimRight(asString(asMap(v)["type"]), "| "))
			}
			expression := strings.Join(alternatives, " | ")
			if expression == "" {
				c.errors = append(c.errors, fmt.Sprintf("%s:%d: empty alias %s", source, asInt(alias["line"]), asString(alias["name"])))
				continue
			}
			if strings.HasSuffix(expression, "|") {
				c.notes = append(c.notes, M{"symbol": alias["name"], "rule": "remove empty trailing union member", "source": source})
				expression = strings.TrimRight(expression, "| ")
			}
			c.aliases[asString(alias["name"])] = M{"expression": expression, "source": source, "line": alias["line"]}
		}
		for _, v := range asList(doc["classes"]) {
			cls := asMap(v)
			record := c.record(asString(cls["name"]), source, "interface")
			for _, base := range asList(cls["bases"]) {
				if !containsString(stringsOf(record["bases"]), asString(base)) {
					appendField(record, "bases", base)
				}
			}
			for _, value := range asList(cls["fields"]) {
				field := asMap(value)
				asMap(record["fields"])[asString(field["name"])] = withFields(field, M{"source": source, "native": false})
			}
		}
		for _, v := range asList(doc["enums"]) {
			enum := asMap(v)
			values := asList(enum["values"])
			numeric := true
			for _, v := range values {
				if !annotationNumeric(asMap(v)["value"]) {
					numeric = false
				}
			}
			expression := "string"
			if numeric {
				expression = "number"
			}
			name := asString(enum["name"])
			c.aliases[name] = M{"expression": expression, "source": source, "values": values}
			recordName := name + "_Values"
			record := c.record(recordName, source)
			for _, v := range values {
				value := asMap(v)
				typ := "string"
				if annotationNumeric(value["value"]) {
					typ = "number"
				}
				asMap(record["fields"])[asString(value["name"])] = M{"type": typ, "native": false, "source": source}
			}
			c.bindGlobal(name, recordName, source)
		}
		for _, v := range asList(doc["functions"]) {
			function := asMap(v)
			if strings.HasPrefix(strings.TrimLeft(asString(function["raw"]), " \t"), "local function ") {
				continue
			}
			item := withFields(function, M{"native": false, "source": source})
			owner := asString(function["owner"])
			name := asString(function["name"])
			if owner != "" {
				methods := asMap(c.record(owner, source)["methods"])
				if previous := asMap(methods[name]); previous != nil {
					item["variants"] = append(append([]any{}, asList(previous["variants"])...), previous)
				}
				methods[name] = item
			} else {
				c.setGlobal(name, item)
			}
		}
		for _, v := range asList(doc["bindings"]) {
			binding := asMap(v)
			raw := asString(binding["raw"])
			if truth(binding["local"]) || len(raw) > 0 && (raw[0] == ' ' || raw[0] == '\t') {
				continue
			}
			name, value := strings.TrimPrefix(asString(binding["name"]), "_G."), asString(binding["value"])
			target := asString(binding["class"])
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
						appendField(record, "bases", strings.TrimSpace(m))
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
	for name, expression := range map[string]string{"bool": "boolean", "cstring": "string", "luaIndex": "number", "UnitTokenType": "string", "UnitTokenVariant": "string", "LuaFunctionReference": "function", "Secret": "OpaqueValue", "Seconds": "number"} {
		if c.aliases[name] == nil {
			c.aliases[name] = M{"expression": expression, "source": "Blizzard metadata primitive"}
		}
	}
}
func annotationNumeric(value any) bool {
	_, ok, _ := luaNumber(value)
	if _, b := value.(bool); b {
		return true
	}
	return ok
}
func (c *catalogue) bindGlobal(name, typeName, source string) {
	parts := strings.Split(name, ".")
	if len(parts) == 1 {
		c.setGlobal(name, M{"type_name": typeName, "source": source})
		return
	}
	parent := parts[0] + "_Namespace"
	c.setGlobal(parts[0], M{"type_name": parent, "source": source})
	record := c.record(parent, source)
	if len(parts) != 2 {
		c.errors = append(c.errors, source+": unsupported global nesting "+name)
		return
	}
	asMap(record["fields"])[parts[1]] = M{"resolved": ident(typeName), "source": source}
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
	lock := mustMap(c.project.Lock("sources.json"))
	for _, key := range sortedKeys(lock) {
		source := asMap(lock[key])
		for _, v := range asList(source["files"]) {
			item := asMap(v)
			name := asString(item["name"])
			if key != "blizzard" && key != "questie" || key == "blizzard" && !strings.HasSuffix(name, ".lua") {
				disposition := "verification-evidence"
				if key == "blizzard" {
					disposition = "registration-index"
				}
				c.files = append(c.files, M{"source": key + ":" + name, "disposition": disposition})
			}
		}
	}
	rows := asList(asMap(lock["blizzard"])["files"])
	sort.SliceStable(rows, func(i, j int) bool { return asString(asMap(rows[i])["name"]) < asString(asMap(rows[j])["name"]) })
	for _, v := range rows {
		entry := asMap(v)
		name := asString(entry["name"])
		if !strings.HasSuffix(name, ".lua") {
			continue
		}
		doc, e := parseLuaFile(c.project.InputPath("blizzard", name))
		if e != nil {
			reject("%s", e)
		}
		c.nativeDocs = append(c.nativeDocs, nativeDocument{name, doc})
		source := "blizzard:" + name
		c.files = append(c.files, M{"source": source, "disposition": "converted"})
		for _, v := range asList(doc["Tables"]) {
			table := asMap(v)
			name, kind := asString(table["Name"]), asString(table["Type"])
			if c.nativeNames[name] {
				reject("%s: duplicate native type %s", source, name)
			}
			c.nativeNames[name] = true
			fields := asList(table["Fields"])
			switch kind {
			case "Enumeration":
				delete(c.defs, name)
				if name != "UnitTokenType" {
					c.aliases[name] = M{"expression": "number", "source": source, "values": getOr(table, "Fields", []any{})}
				}
				c.aliases["Enum."+name] = M{"expression": "number", "source": source, "values": getOr(table, "Fields", []any{})}
				record := c.record("Enum."+name+"_Values", source)
				record["fields"] = M{}
				for _, v := range fields {
					field := asMap(v)
					asMap(record["fields"])[asString(field["Name"])] = M{"type": "number", "native": false, "source": source}
				}
				c.bindGlobal("Enum."+name, "Enum."+name+"_Values", source)
				metadata := M{}
				for _, key := range []string{"NumValues", "MinValue", "MaxValue"} {
					if _, ok := table[key]; ok {
						metadata[key] = M{"type": "number", "native": false, "source": source}
					}
				}
				if len(metadata) > 0 {
					metaName := "Enum." + name + "Meta_Values"
					c.record(metaName, source)["fields"] = metadata
					c.bindGlobal("Enum."+name+"Meta", metaName, source)
				}
			case "Structure":
				record := c.record(name, source)
				record["kind"], record["source"], record["bases"] = "record", source, []any{}
				record["fields"] = M{}
				for _, v := range fields {
					field := asMap(v)
					asMap(record["fields"])[asString(field["Name"])] = withFields(field, M{"source": source, "native": true})
				}
			case "CallbackType":
				c.aliases[name] = M{"signature": withFields(table, M{"native": true, "source": source}), "source": source}
			case "Constants":
				record := c.record("Constants."+name+"_Values", source)
				record["fields"] = M{}
				for _, v := range fields {
					field := asMap(v)
					asMap(record["fields"])[asString(field["Name"])] = withFields(field, M{"source": source, "native": true})
				}
				c.bindGlobal("Constants."+name, "Constants."+name+"_Values", source)
			default:
				c.errors = append(c.errors, source+": unsupported documentation table "+kind)
			}
		}
		namespace := asString(doc["Namespace"])
		owner := ""
		if asString(doc["Type"]) == "ScriptObject" {
			owner = scriptName(asString(doc["Name"]))
			record := c.record(owner, source, "interface")
			record["native_handle"], record["native_object"] = true, asString(doc["ObjectType"]) == "Userdata"
			alias := strings.TrimSuffix(asString(doc["Name"]), "API")
			if c.aliases[alias] == nil {
				c.aliases[alias] = M{"expression": owner, "source": source}
			}
		} else if namespace != "" {
			owner = namespace
			c.record(owner, source)
			c.setGlobal(namespace, M{"type_name": owner, "source": source})
			if asString(doc["Environment"]) == "SecureOnly" {
				c.secureGlobals[namespace] = true
			}
		}
		for _, v := range asList(doc["Functions"]) {
			function := asMap(v)
			name := asString(function["Name"])
			symbol := name
			if owner != "" {
				symbol = owner + "." + name
			}
			if c.nativeFunctions[symbol] {
				reject("%s: duplicate native function %s", source, symbol)
			}
			c.nativeFunctions[symbol] = true
			var ownerValue any
			if owner != "" {
				ownerValue = owner
			}
			item := withFields(function, M{"source": source, "native": true, "owner": ownerValue, "method": asString(doc["Type"]) == "ScriptObject", "Environment": getOr(function, "Environment", getOr(doc, "Environment", "All"))})
			if asString(item["Environment"]) == "SecureOnly" {
				if owner != "" && asString(doc["Environment"]) != "SecureOnly" {
					reject("%s: %s: mixed-visibility methods need a separate public namespace projection", source, symbol)
				}
				if owner == "" {
					c.secureGlobals[name] = true
				}
			}
			if owner != "" {
				asMap(asMap(c.defs[owner])["methods"])[name] = item
			} else {
				c.setGlobal(name, item)
			}
			c.symbols = append(c.symbols, M{"name": symbol, "source": source, "arguments": getOr(function, "Arguments", []any{}), "returns": getOr(function, "Returns", []any{}), "may_return_nothing": getOr(function, "MayReturnNothing", false), "restrictions": restrictions(item), "metadata": mapCopy(function)})
		}
		for _, v := range asList(doc["Events"]) {
			event := asMap(v)
			c.symbols = append(c.symbols, M{"name": getOr(event, "LiteralName", event["Name"]), "source": source, "event": event})
		}
	}
	for _, doc := range c.nativeDocs {
		for _, v := range asList(doc.Metadata["Predicates"]) {
			p := asMap(v)
			if asString(p["Type"]) == "Precondition" && containsString([]string{"ReturnNothing", "ReturnWithError"}, asString(p["FailureMode"])) {
				c.absencePreconditions[asString(p["Name"])] = true
			}
		}
	}
	c.absencePreconditions["RequiresNonSecretAura"] = true
}
func (c *catalogue) mayReturnNothing(item M) bool {
	if truth(item["MayReturnNothing"]) {
		return true
	}
	for name := range c.absencePreconditions {
		if truth(item[name]) {
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
	case "string", "boolean", "any":
		return name
	}
	raw := asString(asMap(c.aliases[name])["expression"])
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
	record := asMap(c.defs[name])
	if truth(record["native_handle"]) {
		return true
	}
	seen[name] = true
	alias := asString(asMap(c.aliases[name])["expression"])
	if alias != "" && c.handleSeen(alias, seen) {
		return true
	}
	for _, base := range stringsOf(record["bases"]) {
		if c.handleSeen(base, seen) {
			return true
		}
	}
	return false
}
func (c *catalogue) resultSecret(item M) bool {
	return !truth(item["ReturnsNeverSecret"]) && (secretResult(item) || truth(item["secret_effect"]))
}
func (c *catalogue) nativeType(item M, secret bool) (string, error) {
	return captureString(func() string { return c.nativeTypeValue(item, secret) })
}
func (c *catalogue) nativeTypeValue(item M, secret bool) string {
	raw := asString(item["Type"])
	secret = !truth(item["NeverSecret"]) && (secret || secretResult(item))
	if raw == "table" && (secret || truth(item["ConditionalSecretContents"])) && !truth(item["NeverSecretContents"]) {
		result := "OpaqueValue"
		if truth(item["Nilable"]) {
			result += " | nil"
		}
		return result
	}
	result := ""
	if raw == "table" {
		inner := asString(item["InnerType"])
		if inner == "" {
			result = "{any: any}"
		} else {
			value := c.nativeTypeValue(M{"Type": inner}, !truth(item["NeverSecretContents"]) && (secret || truth(item["ConditionalSecretContents"])))
			key := ""
			if truth(item["KeyType"]) {
				key = c.nativeTypeValue(M{"Type": item["KeyType"]}, false) + ": "
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
		case "any":
			result = "OpaqueValue"
		default:
			if c.isHandle(raw) {
				result = "OpaqueValue"
			} else if !containsString([]string{"table", "Secret", "LuaValueReference", "LuaValueVariant"}, raw) {
				known := false
				for name, v := range c.defs {
					if result == ident(name) && containsString([]string{"record", "interface", "converted_record"}, asString(asMap(v)["kind"])) {
						known = true
					}
				}
				if !known {
					result = "OpaqueValue"
				}
			}
		}
	}
	if truth(item["Nilable"]) {
		result += " | nil"
	}
	return result
}
func (c *catalogue) loadFrameXML() {
	supplement := mustMap(c.project.Lock("framexml.json"))
	source := "reviewed FrameXML supplement"
	for _, name := range sortedKeys(asMap(supplement["aliases"])) {
		c.aliases[name] = M{"expression": asMap(supplement["aliases"])[name], "source": source}
	}
	for _, v := range asList(supplement["classes"]) {
		cls := asMap(v)
		record := c.record(asString(cls["name"]), source, "interface")
		for _, base := range asList(cls["bases"]) {
			if !containsString(stringsOf(record["bases"]), asString(base)) {
				appendField(record, "bases", base)
			}
		}
	}
	for _, v := range asList(supplement["classes"]) {
		cls := asMap(v)
		owner := asString(cls["name"])
		record := asMap(c.defs[owner])
		for _, name := range sortedKeys(asMap(cls["fields"])) {
			asMap(record["fields"])[name] = M{"type": asMap(cls["fields"])[name], "native": false, "source": source}
		}
		for _, name := range sortedKeys(asMap(cls["methods"])) {
			expression := asString(asMap(cls["methods"])[name])
			callback := c.tealValue(expression, owner, nil)
			signature := asString(asMap(c.defs[callback])["signature"])
			receiver := "function(" + ident(owner)
			pattern := regexp.MustCompile(`^fun\(\s*self\s*:\s*` + regexp.QuoteMeta(owner) + `\s*[,)]`)
			if !pattern.MatchString(expression) || !strings.HasPrefix(signature, receiver) || len(signature) <= len(receiver) || !strings.ContainsRune(",)", rune(signature[len(receiver)])) {
				reject("%s.%s: expected an explicit owning receiver", owner, name)
			}
			signature = "function(self" + signature[len(receiver):]
			asMap(record["methods"])[name] = M{"resolved_signature": signature, "source": source}
		}
	}
	for _, name := range sortedKeys(asMap(supplement["globals"])) {
		c.setGlobal(name, M{"resolved_signature": c.tealValue(asString(asMap(supplement["globals"])[name]), "", nil), "source": source})
	}
	c.notes = append(c.notes, asList(supplement["evidence"])...)
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
	preserved, e := p.AuditMetadata(manifest)
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
	reports := map[string]M{"wow_manifest.json": manifest, "wow_tsm_crosscheck.json": crosscheck}
	for name, r := range reports {
		b, e := jsonPretty(r)
		if e != nil {
			return Artifacts{}, e
		}
		files["evidence/generated/"+name] = b
	}
	files["evidence/generated/wow_native_metadata.json"] = mustJSONUTF8(preserved)
	return Artifacts{Files: files, Statistics: asMap(manifest["statistics"])}, nil
}
func mustJSONUTF8(v any) []byte {
	b, e := jsonPrettyUTF8(v)
	if e != nil {
		reject("%s", e)
	}
	return b
}
func mustFile(path string) []byte {
	b, e := os.ReadFile(path)
	if e != nil {
		reject("%s", e)
	}
	return b
}
