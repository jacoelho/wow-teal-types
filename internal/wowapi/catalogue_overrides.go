package wowapi

import (
	"fmt"
	"maps"
	"slices"
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
	FrameFactory                 frameFactoryOverride        `json:"frame_factory"`
}

// frameFactoryOverride lists the CreateFrame frame types and templates whose
// literal names select a narrower result than Frame.
type frameFactoryOverride struct {
	FrameTypes []string                `json:"frame_types"`
	Templates  []frameTemplateOverride `json:"templates"`
}

// frameTemplateOverride applies the template's mixin Type to frames of the
// listed types created with exactly that template name.
type frameTemplateOverride struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	FrameTypes []string `json:"frame_types"`
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
	Returns       *[]nativeParameter `json:"returns"`
	ArgumentTypes map[string]string  `json:"argument_types"`
}
type supplementalExclusion struct {
	Owner     string `json:"owner"`
	Method    string `json:"method"`
	Source    string `json:"source"`
	Signature string `json:"signature"`
}
// receiverHierarchyOverride replaces owner's reviewed bases. Owner keeps the
// own methods of each InheritMethodsFrom base it detaches from, unless it
// declares them itself.
type receiverHierarchyOverride struct {
	Owner              string   `json:"owner"`
	ExpectedBases      []string `json:"expected_bases"`
	Bases              []string `json:"bases"`
	InheritMethodsFrom []string `json:"inherit_methods_from"`
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
func nativeArgument(item *catalogueFunction, name string) *nativeParameter {
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
func annotationArgument(item *catalogueFunction, name string) *annotationParameter {
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
func (c *catalogue) applyOverrides(overrides catalogueOverrides) {
	c.abiTuples = overrides.ABITuples
	for _, name := range slices.Sorted(maps.Keys(overrides.Aliases)) {
		item := overrides.Aliases[name]
		c.aliases[name] = &catalogueAlias{Expression: item.Type, Source: "reviewed override"}
	}
	for _, name := range slices.Sorted(maps.Keys(overrides.OpaqueObjects)) {
		item := overrides.OpaqueObjects[name]
		c.record(name, "reviewed opaque native object", kindRecord).Opaque = item.Representation != "table"
	}
	c.applyFrameFactory(overrides.FrameFactory)
	// Native arity and restrictions still apply to numeric IDs, paths and texture objects.
	c.method("TextureBase", "SetTexture").Native.Arguments[0].Type = "TextureAsset"
	c.aliases["fileID"] = &catalogueAlias{Expression: "number", Source: "Lua 5.1 numeric asset-ID carrier; native ID validation still applies"}
	// FrameXML also uses coordinate-only and omitted-relative-point shorthands.
	point := c.method("ScriptRegionResizing", "SetPoint")
	point.ResolvedSignature = new("function(self, FramePoint, ? ScriptRegion | string, ? FramePoint, ? number, ? number)")
	point.Variants = []*catalogueFunction{{ResolvedSignature: new("function(self, FramePoint, ? number, ? number)")}, {ResolvedSignature: new("function(self, FramePoint, ScriptRegion | string, ? number, ? number)")}}
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
		v.Type, v.Nilable = "LuaValueVariant", new(true)
	}
	c.applyFormattedText()
	// FrameXML calls SetText with only text; color, alpha and wrap are optional.
	for i := 1; i < len(c.method("GameTooltip", "SetText").Native.Arguments); i++ {
		c.method("GameTooltip", "SetText").Native.Arguments[i].Nilable = new(true)
	}
	// Numeric curves retain numeric results; general curves may return other carriers.
	numericCurves := []*catalogueFunction{c.global("UnitHealthPercent"), c.global("UnitPowerPercent")}
	durationMethods := c.definition("DurationObject").Methods
	for _, name := range slices.Sorted(maps.Keys(durationMethods)) {
		if strings.HasPrefix(name, "Evaluate") {
			numericCurves = append(numericCurves, durationMethods[name])
		}
	}
	for _, item := range numericCurves {
		var target *nativeParameter
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
		item.Annotation.Returns[0].Type = "any?"
		if name == "GetLibrary" && len(item.Annotation.Returns) == 1 {
			item.Annotation.Returns = append(item.Annotation.Returns, annotationReturn{Type: "number?"})
		}
		if name == "NewLibrary" {
			annotationArgument(item, "minor").Type = "number|string"
		}
	}
	item := c.global("LibStub")
	item.Annotation.Generics = []string{}
	item.Annotation.Params[0].Type = "string"
	item.Annotation.Returns[0].Type = "any?"
	if len(item.Annotation.Returns) == 1 {
		item.Annotation.Returns = append(item.Annotation.Returns, annotationReturn{Type: "number?"})
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
	c.definition("AceHook-3.0").Fields["hooks"].Type = "table<any,any>"
	// Repair the malformed annotation to accept a macro name or numeric index.
	annotationArgument(c.global("SetBindingMacro"), "name").Type = "number|string"
	for i := range c.global("GetCurrenciesString").Annotation.Returns {
		output := &c.global("GetCurrenciesString").Annotation.Returns[i]
		if strings.ReplaceAll(output.Type, " ", "") == "number[]|table[]" {
			output.Type = "any"
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
				item.Native.Returns = append([]nativeParameter{}, (*correction.Returns)...)
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
	animation.ResolvedSignature = new("function(self, " + c.literal([]string{"RadialProgress"}) + ", ? string, ? string): RadialProgress")
	animation.Variants = append([]*catalogueFunction{genericAnimation}, animation.Variants...)
	c.applyReceiverCorrections(overrides)
}

// applyFormattedText appends the string.format arguments that native
// documentation omits from every SetFormattedText. The arguments share the
// method's secret permission: FrameXML passes UnitName results to
// FontString:SetFormattedText (ChatFrameEditBox.lua, RaidFinder.lua).
func (c *catalogue) applyFormattedText() {
	c.aliases["FormatArgument"] = &catalogueAlias{Expression: "string | number", Source: "Lua 5.1 string.format argument"}
	owners := 0
	for _, name := range slices.Sorted(maps.Keys(c.defs)) {
		item := c.defs[name].Methods["SetFormattedText"]
		if item == nil {
			continue
		}
		if item.Native == nil {
			reject(fmt.Sprintf("%s.SetFormattedText has no native contract; review the format argument override", name))
		}
		arguments := item.Native.Arguments
		if len(arguments) != 1 || arguments[0].Name != "text" || arguments[0].StrideIndex != nil {
			reject(fmt.Sprintf("%s.SetFormattedText native arguments changed; review the format argument override", name))
		}
		item.Native.Arguments = append(arguments, nativeParameter{Name: "formatArguments", Type: "FormatArgument", Nilable: new(true), StrideIndex: new(1)})
		owners++
	}
	if owners == 0 {
		reject("no SetFormattedText method found; review the format argument override")
	}
}

// applyFrameFactory declares CreateFrame overloads, most specific first: a
// frame-type literal with a template literal, a frame-type literal, then any
// string returning Frame. Literals are case-sensitive while the client is not,
// so other spellings, template lists and unknown templates select a wider result.
func (c *catalogue) applyFrameFactory(factory frameFactoryOverride) {
	const source = "reviewed native frame factory"
	if len(factory.FrameTypes) == 0 {
		reject("frame factory lists no frame types")
	}
	signature := func(frameType, template, result string) *catalogueFunction {
		templateArgument := "? string"
		if template != "" {
			templateArgument = c.literal([]string{template})
		}
		return &catalogueFunction{Source: source, ResolvedSignature: new("function(" + c.literal([]string{frameType}) + ", ? string, ? Frame, " + templateArgument + ", ? number): " + result)}
	}
	overloads := []*catalogueFunction{}
	for _, template := range factory.Templates {
		if c.defs[template.Type] == nil {
			reject(fmt.Sprintf("frame template %s type %s is not defined", template.Name, template.Type))
		}
		for _, frameType := range template.FrameTypes {
			if !slices.Contains(factory.FrameTypes, frameType) {
				reject(fmt.Sprintf("frame template %s lists unknown frame type %s", template.Name, frameType))
			}
			name := frameType + "_" + template.Name
			if c.defs[name] != nil || c.aliases[name] != nil {
				reject(fmt.Sprintf("frame template type %s collides with a declaration", name))
			}
			c.record(name, source, kindInterface).Bases = []string{frameType, template.Type}
			overloads = append(overloads, signature(frameType, template.Name, name))
		}
	}
	for _, frameType := range factory.FrameTypes {
		if !c.extends(frameType, "Frame") {
			reject(fmt.Sprintf("frame factory type %s does not extend Frame", frameType))
		}
		overloads = append(overloads, signature(frameType, "", frameType))
	}
	fallback := &catalogueFunction{Source: source, ResolvedSignature: new("function(string, ? string, ? Frame, ? string, ? number): Frame")}
	overloads[0].Variants = append(overloads[1:], fallback)
	c.setGlobal("CreateFrame", overloads[0])
}

// extends reports whether the definition name is base or inherits from it.
func (c *catalogue) extends(name, base string) bool {
	seen := map[string]bool{}
	pending := []string{name}
	for len(pending) > 0 {
		current := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		if current == base {
			return true
		}
		if seen[current] {
			continue
		}
		seen[current] = true
		pending = append(pending, definitionBases(c.defs[current])...)
	}
	return false
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
		if actual := c.definition(owner).Bases; !slices.Equal(actual, rule.ExpectedBases) {
			reject(fmt.Sprintf("review changed receiver hierarchy: %s has bases %q, expected %q", owner, actual, rule.ExpectedBases))
		}
		for _, base := range rule.Bases {
			if c.defs[base] == nil {
				reject(fmt.Sprintf("review changed receiver hierarchy: %s base %s is not defined", owner, base))
			}
		}
		methods := c.definition(owner).Methods
		for _, base := range rule.InheritMethodsFrom {
			if !slices.Contains(rule.ExpectedBases, base) || slices.Contains(rule.Bases, base) {
				reject(fmt.Sprintf("review changed receiver hierarchy: %s inherits methods from %s, which it does not detach", owner, base))
			}
			for name, method := range c.definition(base).Methods {
				if methods[name] == nil {
					methods[name] = method
				}
			}
		}
		c.definition(owner).Bases = append([]string{}, rule.Bases...)
	}
}

// applyNativeTableContracts restores documented container relations that the
// native schema's table/any labels cannot carry. Mutable maps and arrays stay
// generic instead of using untyped containers as supertypes.
func (c *catalogue) applyNativeTableContracts() {
	for _, contract := range []struct{ name, mapSignature, arraySignature string }{
		{"keys", "function<K, V>({K: V}): {K}", "function<V>({V}): {number}"},
		{"values", "function<K, V>({K: V}): {V}", "function<V>({V}): {V}"},
		{"removeunordered", "", "function<V>({V}, ? number): V | nil"},
		{"freeze", "function<K, V>({K: V}): {K: V}", "function<V>({V}): {V}"},
		{"count", "function<K, V>({K: V})", "function<V>({V})"},
		{"getcountinfo", "function<K, V>({K: V})", "function<V>({V})"},
		{"contains", "function<K, V>({K: V}, ? any)", "function<V>({V}, ? any)"},
		{"indexof", "function<K, V>({K: V}, ? any)", "function<V>({V}, ? any)"},
		{"isempty", "function<K, V>({K: V})", "function<V>({V})"},
		{"isfrozen", "function<K, V>({K: V})", "function<V>({V})"},
		{"removevalue", "function<K, V>({K: V}, ? any)", "function<V>({V}, ? any)"},
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
		item.ResolvedSignature = new(signatures[0])
		for _, signature := range signatures[1:] {
			variant := item.clone()
			variant.Variants = nil
			variant.ResolvedSignature = new(signature)
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
	item.Variants = append(item.Variants, &catalogueFunction{ResolvedSignature: new("function(string, any...): " + dataType + "..."), Source: item.Source})
}
