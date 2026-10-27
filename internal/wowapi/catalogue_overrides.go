package wowapi

import (
	"fmt"
	"path/filepath"
	"strings"
)

type catalogueOverrides struct {
	Aliases                      map[string]aliasOverride    `json:"aliases"`
	OpaqueObjects                map[string]opaqueOverride   `json:"opaque_objects"`
	ABITuples                    map[string][]string         `json:"abi_tuples"`
	SecretEffects                []secretEffectOverride      `json:"secret_effects"`
	NativeContracts              []nativeContractOverride    `json:"native_contracts"`
	SupplementalMethodExclusions []supplementalExclusion     `json:"supplemental_method_exclusions"`
	ReceiverHierarchy            []receiverHierarchyOverride `json:"receiver_hierarchy"`
}
type aliasOverride struct {
	Type string `json:"type"`
}
type opaqueOverride struct {
	Representation string `json:"representation"`
}
type secretEffectOverride struct {
	Symbols []string `json:"symbols"`
}
type nativeContractOverride struct {
	Symbols       []string           `json:"symbols"`
	Signature     *string            `json:"signature"`
	Returns       *[]NativeParameter `json:"returns"`
	ArgumentTypes map[string]string  `json:"argument_types"`
}
type supplementalExclusion struct {
	Owner     string `json:"owner"`
	Method    string `json:"method"`
	Source    string `json:"source"`
	Signature string `json:"signature"`
}
type receiverHierarchyOverride struct {
	Owner         string   `json:"owner"`
	ExpectedBases []string `json:"expected_bases"`
	Bases         []string `json:"bases"`
}

func (c *catalogue) definition(name string) *catalogueDefinition {
	m := c.defs[name]
	if m == nil {
		reject(fmt.Sprintf("missing required definition %s", name))
	}
	return m
}
func (c *catalogue) method(owner, name string) *catalogueFunction {
	m := c.definition(owner).Methods[name]
	if m == nil {
		reject(fmt.Sprintf("missing required method %s.%s", owner, name))
	}
	return m
}
func (c *catalogue) global(name string) *catalogueFunction {
	m := c.globals[name]
	if m == nil {
		reject(fmt.Sprintf("missing required global %s", name))
	}
	return m
}
func nativeArgument(item *catalogueFunction, name string) *NativeParameter {
	if item.Native != nil {
		for i := range item.Native.Arguments {
			if item.Native.Arguments[i].Name == name {
				return &item.Native.Arguments[i]
			}
		}
	}
	reject(fmt.Sprintf("missing required Arguments argument %s", name))
	return nil
}
func annotationArgument(item *catalogueFunction, name string) *AnnotationParameter {
	if item.Annotation != nil {
		for i := range item.Annotation.Params {
			if item.Annotation.Params[i].Name == name {
				return &item.Annotation.Params[i]
			}
		}
	}
	reject(fmt.Sprintf("missing required params argument %s", name))
	return nil
}
func (c *catalogue) applyOverrides() {
	// Frame actors and legacy world actors have different SetPlayerModelFromGlues arguments.
	c.definition("ModelSceneFrameActor").Bases = []string{"ModelSceneFrameActorBase"}
	legacyActor := c.definition("ModelSceneActor")
	legacyActor.Bases = []string{"Object"}
	for name, value := range c.definition("ModelSceneActorBase").Methods {
		methods := legacyActor.Methods
		if methods[name] == nil {
			methods[name] = value
		}
	}
	var overrides catalogueOverrides
	if e := decodeJSON(mustFile(filepath.Join(c.project.Root, "internal/wowapi/inputs/overrides.json")), &overrides); e != nil {
		reject(e.Error())
	}
	c.abiTuples = overrides.ABITuples
	for _, name := range sortedKeys(overrides.Aliases) {
		item := overrides.Aliases[name]
		c.aliases[name] = &catalogueAlias{Expression: item.Type, Source: "reviewed override"}
	}
	for _, name := range sortedKeys(overrides.OpaqueObjects) {
		item := overrides.OpaqueObjects[name]
		c.record(name, "reviewed opaque native object").Opaque = item.Representation != "table"
	}
	c.setGlobal("CreateFrame", &catalogueFunction{
		ResolvedSignature: stringPointer("function(" + c.literal([]string{"StatusBar"}) + ", ? string, ? Frame, ? string, ? number): StatusBar"),
		Source:            "reviewed native frame factory",
		Variants: []*catalogueFunction{{
			ResolvedSignature: stringPointer("function(string, ? string, ? Frame, ? string, ? number): Frame"),
		}},
	})
	// Native arity and restrictions still apply to numeric IDs, paths and texture objects.
	c.method("TextureBase", "SetTexture").Native.Arguments[0].Type = "TextureAsset"
	c.aliases["fileID"] = &catalogueAlias{Expression: "number", Source: "Lua 5.1 numeric asset-ID carrier; native ID validation still applies"}
	// FrameXML also uses coordinate-only and omitted-relative-point shorthands.
	point := c.method("ScriptRegionResizing", "SetPoint")
	point.ResolvedSignature = stringPointer("function(self, FramePoint, ? ScriptRegion | string, ? FramePoint, ? number, ? number)")
	point.Variants = []*catalogueFunction{{ResolvedSignature: stringPointer("function(self, FramePoint, ? number, ? number)")}, {ResolvedSignature: stringPointer("function(self, FramePoint, ScriptRegion | string, ? number, ? number)")}}
	// Timer results are cancellable handles; callbacks receive the same handle.
	for _, name := range []string{"NewTimer", "NewTicker"} {
		c.method("C_Timer", name).Native.Returns[0].Type = "FunctionContainer"
	}
	c.aliases["TickerCallback"] = &catalogueAlias{Expression: "FunctionContainer|fun(cb:FunctionContainer)", Source: "reviewed timer callback/handle carrier"}
	for _, owner := range []string{"Font", "FontString", "EditBox", "SimpleHTML"} {
		nativeArgument(c.method(owner, "SetFontObject"), "font").Type = "FontObject"
	}
	// SecureHandlers.lua stores arbitrary Lua values in frame attributes.
	for _, name := range []string{"SetAttribute", "SetAttributeNoHandler"} {
		v := nativeArgument(c.method("Frame", name), "value")
		v.Type, v.Nilable = "LuaValueVariant", boolPointer(true)
	}
	// FrameXML calls SetText with only text; color, alpha and wrap are optional.
	for i := 1; i < len(c.method("GameTooltip", "SetText").Native.Arguments); i++ {
		c.method("GameTooltip", "SetText").Native.Arguments[i].Nilable = boolPointer(true)
	}
	c.definition("LuaDurationManualClock").Bases = append(c.definition("LuaDurationManualClock").Bases, "LuaDurationClockObject")
	c.definition("SecondsFormatter").Bases = append(c.definition("SecondsFormatter").Bases, "NumericFormatter")
	// Numeric curves retain numeric results; general curves may return other carriers.
	numericCurves := []*catalogueFunction{c.global("UnitHealthPercent"), c.global("UnitPowerPercent")}
	durationMethods := c.definition("DurationObject").Methods
	for _, name := range sortedKeys(durationMethods) {
		if strings.HasPrefix(name, "Evaluate") {
			numericCurves = append(numericCurves, durationMethods[name])
		}
	}
	for _, item := range numericCurves {
		var target *NativeParameter
		for i := range item.Native.Arguments {
			argument := &item.Native.Arguments[i]
			if argument.Type == "LuaCurveObjectBase" {
				target = argument
				break
			}
		}
		if target == nil {
			continue
		}
		original := item.clone()
		target.Type = "LuaCurveObject"
		item.Native.Returns[0].Type = "number"
		item.Variants = []*catalogueFunction{original}
	}
	c.method("RootMenuDescriptionProxy", "AddMenuReleasedCallback").Annotation.Params[0].Type = "fun(menu:MenuProxy,reason:number)"
	annotationArgument(c.global("SetupTextureKitOnFrame"), "setVisibility").Optional = true
	// Library-name-dependent results require an explicit consumer type assertion.
	for _, name := range []string{"NewLibrary", "GetLibrary"} {
		item := c.method("LibStub", name)
		item.Annotation.Generics = []string{}
		item.Annotation.Params[0].Type = "string"
		item.Annotation.Returns[0].Type = "unknown?"
		if name == "GetLibrary" && len(item.Annotation.Returns) == 1 {
			item.Annotation.Returns = append(item.Annotation.Returns, AnnotationReturn{Type: "number?", Name: stringPointer("minor")})
		}
		if name == "NewLibrary" {
			annotationArgument(item, "minor").Type = "number|string"
		}
	}
	item := c.global("LibStub")
	item.Annotation.Generics = []string{}
	item.Annotation.Params[0].Type = "string"
	item.Annotation.Returns[0].Type = "unknown?"
	if len(item.Annotation.Returns) == 1 {
		item.Annotation.Returns = append(item.Annotation.Returns, AnnotationReturn{Type: "number?", Name: stringPointer("minor")})
	}
	for _, name := range []string{"CreateEnumRadioMenu", "CreateEnumRadioContextMenu"} {
		annotationArgument(c.method("MenuUtil", name), "enum").Type = "table<string,V>"
	}
	// Extra mixin members are consumer-owned, beyond the guaranteed root surface.
	for _, owner := range []string{"Menu", "MenuUtil"} {
		item := c.method(owner, "CreateRootMenuDescription")
		item.Annotation.Generics = []string{}
		item.Annotation.Params[0].Type = "table"
		item.Annotation.Returns[0].Type = "RootMenuDescriptionProxy"
	}
	v := annotationArgument(c.method("AceComm-3.0", "SendCommMessage"), "callbackArg")
	v.Type, v.Optional = "any", true
	c.definition("AceHook-3.0").Fields["hooks"].Type = "table<any,unknown>"
	// Repair the malformed annotation to accept a macro name or numeric index.
	annotationArgument(c.global("SetBindingMacro"), "name").Type = "number|string"
	for i := range c.global("GetCurrenciesString").Annotation.Returns {
		output := &c.global("GetCurrenciesString").Annotation.Returns[i]
		if strings.ReplaceAll(output.Type, " ", "") == "number[]|table[]" {
			output.Type = "unknown"
		}
	}
	annotationArgument(c.global("GetCurrenciesString"), "currencies").Type = "(number|table)[]"
	c.global("GetCurrenciesString").Annotation.Returns[0].Type = "string?"
	c.aliases["KStringDiscordUserName"] = &catalogueAlias{Expression: "string", Source: "Blizzard Discord user-name text carrier"}
	call := c.global("LibStub")
	call.Method = true
	c.definition("LibStub").Methods["metamethod __call"] = call
	c.setGlobal("LibStub", &catalogueFunction{TypeName: "LibStub", Source: call.Source})
	for _, effect := range overrides.SecretEffects {
		for _, symbol := range effect.Symbols {
			item := c.symbolItem(symbol)
			variants := append([]*catalogueFunction{item}, item.Variants...)
			for _, variant := range variants {
				variant.SecretEffect = true
			}
		}
	}
	for _, correction := range overrides.NativeContracts {
		for _, symbol := range correction.Symbols {
			item := c.symbolItem(symbol)
			if correction.Signature != nil {
				item.ResolvedSignature = correction.Signature
			}
			if correction.Returns != nil {
				item.Native.Returns = append([]NativeParameter{}, (*correction.Returns)...)
			}
			for argument, typeName := range correction.ArgumentTypes {
				nativeArgument(item, argument).Type = typeName
			}
		}
	}
	c.applyNativeTableContracts()
	c.applySecureCallContracts()
	// Preserve the subtype selected by the literal animation name.
	animation := c.method("AnimationGroup", "CreateAnimation")
	genericAnimation := animation.clone()
	animation.ResolvedSignature = stringPointer("function(self, " + c.literal([]string{"RadialProgress"}) + ", ? string, ? string): RadialProgress")
	animation.Variants = append([]*catalogueFunction{genericAnimation}, animation.Variants...)
	c.applyReceiverCorrections(overrides)
	c.applyLayeredRegionHierarchy()
}
func (c *catalogue) symbolItem(symbol string) *catalogueFunction {
	index := strings.LastIndexByte(symbol, '.')
	if index < 0 {
		return c.global(symbol)
	}
	return c.method(symbol[:index], symbol[index+1:])
}
func (c *catalogue) applyReceiverCorrections(overrides catalogueOverrides) {
	exclusions, hierarchies := overrides.SupplementalMethodExclusions, overrides.ReceiverHierarchy
	if len(exclusions) == 0 && len(hierarchies) == 0 {
		return
	}
	for _, rule := range exclusions {
		owner, name := rule.Owner, rule.Method
		symbol := owner + "." + name
		item := c.definition(owner).Methods[name]
		if item == nil || item.Native != nil || item.Source != rule.Source || c.signatureValue(item, owner) != rule.Signature {
			reject(fmt.Sprintf("review changed supplemental exclusion contract: %s", symbol))
		}
		delete(c.definition(owner).Methods, name)
	}
	for _, rule := range hierarchies {
		owner := rule.Owner
		if requireStringJSON(c.definition(owner).Bases) != requireStringJSON(rule.ExpectedBases) {
			reject(fmt.Sprintf("review changed receiver hierarchy: %s", owner))
		}
		for _, base := range rule.Bases {
			if c.defs[base] == nil {
				reject(fmt.Sprintf("review changed receiver hierarchy: %s", owner))
			}
		}
		c.definition(owner).Bases = append([]string{}, rule.Bases...)
	}
}

// applyNativeTableContracts restores documented container relations that the
// native schema's table/any labels cannot carry. Mutable maps and arrays stay
// generic instead of using invariant unknown-valued containers as supertypes.
func (c *catalogue) applyNativeTableContracts() {
	for _, contract := range []struct{ name, mapSignature, arraySignature string }{
		{"keys", "function<K, V>({K: V}): {K}", "function<V>({V}): {number}"},
		{"values", "function<K, V>({K: V}): {V}", "function<V>({V}): {V}"},
		{"removeunordered", "", "function<V>({V}, ? number): V | nil"},
		{"freeze", "function<K, V>({K: V}): {K: V}", "function<V>({V}): {V}"},
		{"count", "function<K, V>({K: V})", "function<V>({V})"},
		{"getcountinfo", "function<K, V>({K: V})", "function<V>({V})"},
		{"contains", "function<K, V>({K: V}, ? unknown)", "function<V>({V}, ? unknown)"},
		{"indexof", "function<K, V>({K: V}, ? unknown)", "function<V>({V}, ? unknown)"},
		{"isempty", "function<K, V>({K: V})", "function<V>({V})"},
		{"isfrozen", "function<K, V>({K: V})", "function<V>({V})"},
		{"removevalue", "function<K, V>({K: V}, ? unknown)", "function<V>({V}, ? unknown)"},
	} {
		item := c.method("table", contract.name)
		original := c.signatureValue(item, "")
		results := splitTop(original, ':')
		suffix := ""
		if len(results) > 1 {
			suffix = ": " + strings.Join(results[1:], ":")
		}
		signatures := []string{}
		for _, signature := range []string{contract.mapSignature, contract.arraySignature} {
			if signature == "" {
				continue
			}
			if len(splitTop(signature, ':')) == 1 {
				signature += suffix
			}
			signatures = append(signatures, signature)
		}
		item.ResolvedSignature = stringPointer(signatures[0])
		for _, signature := range signatures[1:] {
			variant := item.clone()
			variant.Variants = nil
			variant.ResolvedSignature = stringPointer(signature)
			item.Variants = append(item.Variants, variant)
		}
	}
}

// A function-name lookup cannot bind the callback's result parameter. Keep it
// separate from the source's homogeneous callback argument/result relation.
func (c *catalogue) applySecureCallContracts() {
	item := c.global("securecall")
	argument := annotationArgument(item, "funcOrKey")
	members := splitTop(argument.Type, '|')
	if len(members) != 2 || members[0] != "S" || !strings.HasPrefix(members[1], "fun(") {
		reject("review changed securecall callback/name contract")
	}
	argument.Type = members[1]
	item.Variants = append(item.Variants, &catalogueFunction{ResolvedSignature: stringPointer("function(string, unknown...): unknown..."), Source: item.Source})
}
