package wowapi

import (
	"fmt"
	"path/filepath"
	"strings"
)

type catalogueOverrides struct {
	Aliases                      map[string]RepresentationNote `json:"aliases"`
	OpaqueObjects                map[string]RepresentationNote `json:"opaque_objects"`
	ABITuples                    map[string][]string           `json:"abi_tuples"`
	ABITupleEvidence             map[string]RepresentationNote `json:"abi_tuple_evidence"`
	SecretEffects                []RepresentationNote          `json:"secret_effects"`
	NativeContracts              []nativeContractOverride      `json:"native_contracts"`
	SupplementalMethodExclusions []SupplementalExclusion       `json:"supplemental_method_exclusions"`
	ReceiverHierarchy            []receiverHierarchyOverride   `json:"receiver_hierarchy"`
}
type nativeContractOverride struct {
	Symbols       []string           `json:"symbols"`
	Signature     *string            `json:"signature"`
	Returns       *[]NativeParameter `json:"returns"`
	ArgumentTypes map[string]string  `json:"argument_types"`
	Reason        string             `json:"reason"`
	Evidence      EvidenceReferences `json:"evidence"`
}
type receiverHierarchyOverride struct {
	Owner         string             `json:"owner"`
	ExpectedBases []string           `json:"expected_bases"`
	Bases         []string           `json:"bases"`
	Reason        string             `json:"reason"`
	Evidence      EvidenceReferences `json:"evidence"`
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
func (c *catalogue) note(symbol, rule string, extra ...RepresentationNote) {
	note := RepresentationNote{Symbol: stringPointer(symbol), Rule: rule}
	if len(extra) > 0 {
		note = extra[0]
		note.Symbol = stringPointer(symbol)
		note.Rule = rule
	}
	c.notes = append(c.notes, note)
}
func evidenceText(text string) *EvidenceReferences { return &EvidenceReferences{Text: &text} }
func (c *catalogue) applyOverrides() {
	c.definition("ModelSceneFrameActor").Bases = []string{"ModelSceneFrameActorBase"}
	legacyActor := c.definition("ModelSceneActor")
	legacyActor.Bases = []string{"Object"}
	for name, value := range c.definition("ModelSceneActorBase").Methods {
		methods := legacyActor.Methods
		if methods[name] == nil {
			methods[name] = value
		}
	}
	c.note("ModelSceneFrameActor", "keep native frame actor separate from Questie's legacy world-actor alias; their SetPlayerModelFromGlues argument lists differ")
	var overrides catalogueOverrides
	if e := decodeJSON(mustFile(filepath.Join(c.project.Root, "scripts/wow_api/overrides.json")), &overrides); e != nil {
		reject(e.Error())
	}
	c.abiTuples = overrides.ABITuples
	for _, name := range sortedKeys(overrides.Aliases) {
		item := overrides.Aliases[name]
		c.aliases[name] = &catalogueAlias{Expression: item.Type, Source: "reviewed override"}
		item.Symbol = stringPointer(name)
		c.notes = append(c.notes, item)
	}
	for _, name := range sortedKeys(overrides.OpaqueObjects) {
		item := overrides.OpaqueObjects[name]
		c.record(name, "reviewed opaque native object").Opaque = item.Representation != "table"
		item.Symbol = stringPointer(name)
		c.notes = append(c.notes, item)
	}
	c.setGlobal("CreateFrame", &catalogueFunction{
		ResolvedSignature: stringPointer("function(" + c.literal([]string{"StatusBar"}) + ", ? string, ? Frame, ? string, ? number): StatusBar"),
		Source:            "reviewed native frame factory",
		Variants: []*catalogueFunction{{
			ResolvedSignature: stringPointer("function(string, ? string, ? Frame, ? string, ? number): Frame"),
		}},
	})
	c.method("TextureBase", "SetTexture").Native.Arguments[0].Type = "TextureAsset"
	c.note("TextureBase.SetTexture", "retain native arity/restrictions; correct texture carrier to numeric file ID, path, or texture object", RepresentationNote{Evidence: evidenceText("Questie Widget/Base/TextureBase.lua and Blizzard numeric icon-file call sites")})
	c.aliases["fileID"] = &catalogueAlias{Expression: "number", Source: "Lua 5.1 numeric asset-ID carrier; native ID validation still applies"}
	c.note("fileID", "numeric Lua 5.1 carrier, matching existing opaque-number sinks; native ID validation remains required")
	point := c.method("ScriptRegionResizing", "SetPoint")
	point.ResolvedSignature = stringPointer("function(self, FramePoint, ? ScriptRegion | string, ? FramePoint, ? number, ? number)")
	point.Variants = []*catalogueFunction{{ResolvedSignature: stringPointer("function(self, FramePoint, ? number, ? number)")}, {ResolvedSignature: stringPointer("function(self, FramePoint, ScriptRegion | string, ? number, ? number)")}}
	c.note("ScriptRegionResizing.SetPoint", "retain documented native call shorthands from Questie and Blizzard FrameXML call sites")
	for _, name := range []string{"NewTimer", "NewTicker"} {
		c.method("C_Timer", name).Native.Returns[0].Type = "FunctionContainer"
	}
	c.aliases["TickerCallback"] = &catalogueAlias{Expression: "FunctionContainer|fun(cb:FunctionContainer)", Source: "reviewed timer callback/handle carrier"}
	c.note("C_Timer.NewTimer/NewTicker", "native timer results are cancellable FunctionContainer handles; callback receives that handle", RepresentationNote{Evidence: evidenceText("Questie Type/FunctionContainer.lua; locked RingedFrameTemplate.lua and LoopingSoundEffect.lua")})
	for _, owner := range []string{"Font", "FontString", "EditBox", "SimpleHTML"} {
		nativeArgument(c.method(owner, "SetFontObject"), "font").Type = "FontObject"
	}
	for _, name := range []string{"SetAttribute", "SetAttributeNoHandler"} {
		v := nativeArgument(c.method("Frame", name), "value")
		v.Type, v.Nilable = "LuaValueVariant", boolPointer(true)
	}
	c.note("Frame.SetAttribute/SetAttributeNoHandler", "native attributes accept arbitrary Lua values, including FrameXML frame/table/bool attributes; preserve explicit dynamic boundary", RepresentationNote{Evidence: evidenceText("locked Blizzard_RestrictedAddOnEnvironment/SecureHandlers.lua")})
	c.note("Font/FontString/EditBox/SimpleHTML.SetFontObject", "accept a named font or Font object", RepresentationNote{Evidence: evidenceText("Questie FontObject alias and locked FrameXML font call sites")})
	for i := 1; i < len(c.method("GameTooltip", "SetText").Native.Arguments); i++ {
		c.method("GameTooltip", "SetText").Native.Arguments[i].Nilable = boolPointer(true)
	}
	c.note("GameTooltip.SetText", "color, alpha and wrap arguments are optional", RepresentationNote{Evidence: evidenceText("Questie Widget/Frame/GameTooltip.lua; locked Blizzard_SettingsPanel.lua GameTooltip:SetText(categoryTutorial.tooltip)")})
	c.definition("LuaDurationManualClock").Bases = append(c.definition("LuaDurationManualClock").Bases, "LuaDurationClockObject")
	c.definition("SecondsFormatter").Bases = append(c.definition("SecondsFormatter").Bases, "NumericFormatter")
	c.note("LuaDurationManualClock/SecondsFormatter", "preserve base clock/formatter handle contracts required by native duration consumers")
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
	c.note("UnitHealthPercent/UnitPowerPercent/DurationObject.Evaluate*", "numeric curves retain a number-like result; general base/color curves retain the opaque native variant", RepresentationNote{Evidence: evidenceText("native LuaCurveEvaluatedResult documentation distinguishes supplied-curve result carriers")})
	c.method("RootMenuDescriptionProxy", "AddMenuReleasedCallback").Annotation.Params[0].Type = "fun(menu:MenuProxy,reason:number)"
	c.note("RootMenuDescriptionProxy.AddMenuReleasedCallback", "callback receives released menu proxy and numeric close reason", RepresentationNote{Evidence: evidenceText("locked Blizzard_Menu/Menu.lua onReleased invocation")})
	annotationArgument(c.global("SetupTextureKitOnFrame"), "setVisibility").Optional = true
	c.note("SetupTextureKitOnFrame.setVisibility", "optional: native helper tests its truthiness before changing visibility")
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
	c.note("LibStub", "dependent library-name inference requires explicit consumer type assertion")
	for _, name := range []string{"CreateEnumRadioMenu", "CreateEnumRadioContextMenu"} {
		annotationArgument(c.method("MenuUtil", name), "enum").Type = "table<string,V>"
	}
	for _, owner := range []string{"Menu", "MenuUtil"} {
		item := c.method(owner, "CreateRootMenuDescription")
		item.Annotation.Generics = []string{}
		item.Annotation.Params[0].Type = "table"
		item.Annotation.Returns[0].Type = "RootMenuDescriptionProxy"
	}
	c.note("Menu.CreateRootMenuDescription", "mixin input is a table; result retains the guaranteed root-description surface. Extra mixin members require a consumer-owned type.")
	v := annotationArgument(c.method("AceComm-3.0", "SendCommMessage"), "callbackArg")
	v.Type, v.Optional = "any", true
	c.definition("AceHook-3.0").Fields["hooks"].Type = "table<any,unknown>"
	c.note("AceHook-3.0.hooks", "undiscriminable hook-registry values remain opaque")
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
	c.note("CreateFrame", "literal StatusBar construction retains its native subtype; other frame names return the common Frame surface, and template composition is not inferred", RepresentationNote{Evidence: evidenceText("questie:Data/Wiki.lua:3602 and Blizzard_Menu/Compositor.lua:191")})
	c.note("MenuUtil.CreateEnumRadioMenu", "enum is a string-keyed table of V, iterated through pairs by FrameXML")
	c.note("AceComm-3.0.SendCommMessage.callbackArg", "single optional arbitrary callback argument, not a variadic type pack")
	c.note("SetBindingMacro.name", "repair malformed annotation to macro name or index")
	c.note("GetCurrenciesString", "accept numeric IDs, keyed currency entries or positional argument tables; an empty list returns nil")
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
			c.note(symbol, effect.Reason, RepresentationNote{Evidence: effect.Evidence})
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
			c.note(symbol, correction.Reason, RepresentationNote{Evidence: &correction.Evidence})
		}
	}
	c.applyNativeTableContracts()
	c.applySecureCallContracts()
	// The recorded native fixture constructs this exact animation subtype.
	animation := c.method("AnimationGroup", "CreateAnimation")
	genericAnimation := animation.clone()
	animation.ResolvedSignature = stringPointer("function(self, " + c.literal([]string{"RadialProgress"}) + ", ? string, ? string): RadialProgress")
	animation.Variants = append([]*catalogueFunction{genericAnimation}, animation.Variants...)
	c.note("AnimationGroup.CreateAnimation", "literal RadialProgress construction retains the native receiver subtype", RepresentationNote{Evidence: evidenceText("recorded public RadialProgress native fixture")})
	c.applyObservedReceiverContracts(overrides)
	c.applyLayeredRegionHierarchy()
}
func (c *catalogue) symbolItem(symbol string) *catalogueFunction {
	index := strings.LastIndexByte(symbol, '.')
	if index < 0 {
		return c.global(symbol)
	}
	return c.method(symbol[:index], symbol[index+1:])
}
func (c *catalogue) applyObservedReceiverContracts(overrides catalogueOverrides) {
	exclusions, hierarchies := overrides.SupplementalMethodExclusions, overrides.ReceiverHierarchy
	if len(exclusions) == 0 && len(hierarchies) == 0 {
		return
	}
	observed, e := c.project.LoadObservations()
	if e != nil {
		reject(e.Error())
	}
	for _, rule := range exclusions {
		owner, name := rule.Owner, rule.Method
		symbol := owner + "." + name
		item := c.definition(owner).Methods[name]
		if item == nil || item.Native != nil || item.Source != rule.Source || c.signatureValue(item, owner) != rule.Signature {
			reject(fmt.Sprintf("review changed supplemental exclusion contract: %s", symbol))
		}
		source := rule.Source
		relative := strings.TrimPrefix(source, "questie:")
		if !strings.HasPrefix(source, "questie:") || filepath.IsAbs(relative) || containsString(strings.Split(filepath.ToSlash(relative), "/"), "..") {
			reject(fmt.Sprintf("invalid supplemental exclusion source: %s", symbol))
		}
		if sha256hex(mustFile(c.project.InputPath("questie", relative))) != rule.SourceSHA256 || observed.Sources.Build != rule.Build {
			reject(fmt.Sprintf("review changed supplemental exclusion provenance: %s", symbol))
		}
		if len(rule.NegativeCases) == 0 {
			reject(fmt.Sprintf("supplemental exclusion lacks negative native observations: %s", symbol))
		}
		for _, name := range rule.NegativeCases {
			if observed.Hierarchy[name].State != "absent" {
				reject(fmt.Sprintf("supplemental exclusion lacks negative native observations: %s", symbol))
			}
		}
		rule.Original = ExcludedAnnotationFunction{AnnotationFunction: *item.clone().Annotation, Source: item.Source}
		rule.ObservationSources = observed.Sources
		rule.CurrentSourceCohortMatches = observed.Applicable
		c.excludedMethods[symbol] = rule
		delete(c.definition(owner).Methods, name)
		c.note(symbol, "exclude unsupported common supplemental guarantee", RepresentationNote{Reason: rule.Reason, Evidence: &EvidenceReferences{List: rule.Evidence}})
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
		c.note(owner, rule.Reason, RepresentationNote{Evidence: &rule.Evidence, ObservedBuild: observed.Sources.Build, CurrentSourceCohortMatches: boolPointer(observed.Applicable)})
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
		c.note("table."+contract.name, "preserve documented container inputs and input/output type relations", RepresentationNote{Evidence: evidenceText("locked LuaTableExtensionsDocumentation.lua function and return documentation")})
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
	c.note("securecall", "function-name lookup has opaque results; only a supplied callback establishes its result relation")
}
