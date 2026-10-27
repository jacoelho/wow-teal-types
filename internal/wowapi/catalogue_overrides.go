package wowapi

import (
	"fmt"
	"path/filepath"
	"strings"
)

func mapPointer(m M) string { return fmt.Sprintf("%p", m) }
func (c *catalogue) definition(name string) M {
	m := asMap(c.defs[name])
	if m == nil {
		reject("missing required definition %s", name)
	}
	return m
}
func (c *catalogue) method(owner, name string) M {
	m := asMap(asMap(c.definition(owner)["methods"])[name])
	if m == nil {
		reject("missing required method %s.%s", owner, name)
	}
	return m
}
func (c *catalogue) global(name string) M {
	m := asMap(c.globals[name])
	if m == nil {
		reject("missing required global %s", name)
	}
	return m
}
func argumentNamed(item M, key, name string) M {
	field := "name"
	if key == "Arguments" {
		field = "Name"
	}
	for _, v := range asList(item[key]) {
		value := asMap(v)
		if asString(value[field]) == name {
			return value
		}
	}
	reject("missing required %s argument %s", key, name)
	return nil
}
func (c *catalogue) note(symbol, rule string, extra ...M) {
	m := M{"symbol": symbol, "rule": rule}
	for _, fields := range extra {
		for k, v := range fields {
			m[k] = v
		}
	}
	c.notes = append(c.notes, m)
}
func (c *catalogue) applyOverrides() {
	c.definition("ModelSceneFrameActor")["bases"] = []any{"ModelSceneFrameActorBase"}
	legacyActor := c.definition("ModelSceneActor")
	legacyActor["bases"] = []any{"Object"}
	for name, value := range asMap(c.definition("ModelSceneActorBase")["methods"]) {
		methods := asMap(legacyActor["methods"])
		if methods[name] == nil {
			methods[name] = value
		}
	}
	c.note("ModelSceneFrameActor", "keep native frame actor separate from Questie's legacy world-actor alias; their SetPlayerModelFromGlues argument lists differ")
	c.aliases["unknown"] = M{"expression": "OpaqueValue", "source": "LuaLS unknown: opaque until explicitly narrowed"}
	overrides := mustMap(c.project.Lock("overrides.json"))
	c.abiTuples = asMap(getOr(overrides, "abi_tuples", M{}))
	for _, name := range sortedKeys(asMap(overrides["aliases"])) {
		item := asMap(asMap(overrides["aliases"])[name])
		c.aliases[name] = M{"expression": item["type"], "source": "reviewed override"}
		c.notes = append(c.notes, withFields(item, M{"symbol": name}))
	}
	for _, name := range sortedKeys(asMap(overrides["opaque_objects"])) {
		item := asMap(asMap(overrides["opaque_objects"])[name])
		c.record(name, "reviewed opaque native object")["opaque"] = asString(item["representation"]) != "table"
		c.notes = append(c.notes, withFields(item, M{"symbol": name}))
	}
	c.setGlobal("CreateFrame", M{"resolved_signature": "function(string, ? string, ? Frame, ? string, ? number): Frame", "source": "reviewed conservative factory"})
	asMap(asList(c.method("TextureBase", "SetTexture")["Arguments"])[0])["Type"] = "TextureAsset"
	c.note("TextureBase.SetTexture", "retain native arity/restrictions; correct texture carrier to numeric file ID, path, or texture object", M{"evidence": "Questie Widget/Base/TextureBase.lua and Blizzard numeric icon-file call sites"})
	c.aliases["fileID"] = M{"expression": "number", "source": "Lua 5.1 numeric asset-ID carrier; native ID validation still applies"}
	c.note("fileID", "numeric Lua 5.1 carrier, matching existing opaque-number sinks; native ID validation remains required")
	point := c.method("ScriptRegionResizing", "SetPoint")
	point["resolved_signature"] = "function(self, FramePoint, ? ScriptRegion | string, ? FramePoint, ? number, ? number)"
	point["variants"] = []any{M{"resolved_signature": "function(self, FramePoint, ? number, ? number)"}, M{"resolved_signature": "function(self, FramePoint, ScriptRegion | string, ? number, ? number)"}}
	c.note("ScriptRegionResizing.SetPoint", "retain documented native call shorthands from Questie and Blizzard FrameXML call sites")
	for _, name := range []string{"NewTimer", "NewTicker"} {
		asMap(asList(c.method("C_Timer", name)["Returns"])[0])["Type"] = "FunctionContainer"
	}
	c.aliases["TickerCallback"] = M{"expression": "FunctionContainer|fun(cb:FunctionContainer)", "source": "reviewed timer callback/handle carrier"}
	c.note("C_Timer.NewTimer/NewTicker", "native timer results are cancellable FunctionContainer handles; callback receives that handle", M{"evidence": "Questie Type/FunctionContainer.lua; locked RingedFrameTemplate.lua and LoopingSoundEffect.lua"})
	for _, owner := range []string{"Font", "FontString", "EditBox", "SimpleHTML"} {
		argumentNamed(c.method(owner, "SetFontObject"), "Arguments", "font")["Type"] = "FontObject"
	}
	for _, name := range []string{"SetAttribute", "SetAttributeNoHandler"} {
		v := argumentNamed(c.method("Frame", name), "Arguments", "value")
		v["Type"], v["Nilable"] = "LuaValueVariant", true
	}
	c.note("Frame.SetAttribute/SetAttributeNoHandler", "native attributes accept arbitrary Lua values, including FrameXML frame/table/bool attributes; preserve explicit dynamic boundary", M{"evidence": "locked Blizzard_RestrictedAddOnEnvironment/SecureHandlers.lua"})
	c.note("Font/FontString/EditBox/SimpleHTML.SetFontObject", "accept a named font or Font object", M{"evidence": "Questie FontObject alias and locked FrameXML font call sites"})
	for _, argument := range asList(c.method("GameTooltip", "SetText")["Arguments"])[1:] {
		asMap(argument)["Nilable"] = true
	}
	c.note("GameTooltip.SetText", "color, alpha and wrap arguments are optional", M{"evidence": "Questie Widget/Frame/GameTooltip.lua; locked Blizzard_SettingsPanel.lua GameTooltip:SetText(categoryTutorial.tooltip)"})
	appendField(c.definition("LuaDurationManualClock"), "bases", "LuaDurationClockObject")
	appendField(c.definition("SecondsFormatter"), "bases", "NumericFormatter")
	c.note("LuaDurationManualClock/SecondsFormatter", "preserve base clock/formatter handle contracts required by native duration consumers")
	numericCurves := []M{c.global("UnitHealthPercent"), c.global("UnitPowerPercent")}
	durationMethods := asMap(c.definition("DurationObject")["methods"])
	for _, name := range sortedKeys(durationMethods) {
		if strings.HasPrefix(name, "Evaluate") {
			numericCurves = append(numericCurves, asMap(durationMethods[name]))
		}
	}
	for _, item := range numericCurves {
		var target M
		for _, v := range asList(item["Arguments"]) {
			argument := asMap(v)
			if asString(argument["Type"]) == "LuaCurveObjectBase" {
				target = argument
				break
			}
		}
		if target == nil {
			continue
		}
		original := mapCopy(item)
		target["Type"] = "LuaCurveObject"
		asMap(asList(item["Returns"])[0])["Type"] = "number"
		item["variants"] = []any{original}
	}
	c.note("UnitHealthPercent/UnitPowerPercent/DurationObject.Evaluate*", "numeric curves retain a number-like result; general base/color curves retain the opaque native variant", M{"evidence": "native LuaCurveEvaluatedResult documentation distinguishes supplied-curve result carriers"})
	asMap(asList(c.method("RootMenuDescriptionProxy", "AddMenuReleasedCallback")["params"])[0])["type"] = "fun(menu:MenuProxy,reason:number)"
	c.note("RootMenuDescriptionProxy.AddMenuReleasedCallback", "callback receives released menu proxy and numeric close reason", M{"evidence": "locked Blizzard_Menu/Menu.lua onReleased invocation"})
	argumentNamed(c.global("SetupTextureKitOnFrame"), "params", "setVisibility")["optional"] = true
	c.note("SetupTextureKitOnFrame.setVisibility", "optional: native helper tests its truthiness before changing visibility")
	for _, name := range []string{"NewLibrary", "GetLibrary"} {
		item := c.method("LibStub", name)
		item["generics"] = []any{}
		asMap(asList(item["params"])[0])["type"] = "string"
		asMap(asList(item["returns"])[0])["type"] = "unknown?"
		if name == "GetLibrary" && len(asList(item["returns"])) == 1 {
			appendField(item, "returns", M{"type": "number?", "name": "minor", "variadic": false})
		}
		if name == "NewLibrary" {
			argumentNamed(item, "params", "minor")["type"] = "number|string"
		}
	}
	item := c.global("LibStub")
	item["generics"] = []any{}
	asMap(asList(item["params"])[0])["type"] = "string"
	asMap(asList(item["returns"])[0])["type"] = "unknown?"
	if len(asList(item["returns"])) == 1 {
		appendField(item, "returns", M{"type": "number?", "name": "minor", "variadic": false})
	}
	c.note("LibStub", "dependent library-name inference requires explicit consumer type assertion")
	for _, name := range []string{"CreateEnumRadioMenu", "CreateEnumRadioContextMenu"} {
		argumentNamed(c.method("MenuUtil", name), "params", "enum")["type"] = "table<string,V>"
	}
	for _, owner := range []string{"Menu", "MenuUtil"} {
		item := c.method(owner, "CreateRootMenuDescription")
		item["generics"] = []any{}
		asMap(asList(item["params"])[0])["type"] = "table"
		asMap(asList(item["returns"])[0])["type"] = "RootMenuDescriptionProxy"
	}
	c.note("Menu.CreateRootMenuDescription", "mixin input is a table; result retains the guaranteed root-description surface. Extra mixin members require a consumer-owned type.")
	v := argumentNamed(c.method("AceComm-3.0", "SendCommMessage"), "params", "callbackArg")
	v["type"], v["optional"] = "any", true
	asMap(asMap(c.definition("AceHook-3.0")["fields"])["hooks"])["type"] = "table<any,unknown>"
	c.note("AceHook-3.0.hooks", "undiscriminable hook-registry values remain opaque")
	argumentNamed(c.global("SetBindingMacro"), "params", "name")["type"] = "number|string"
	for _, v := range asList(c.global("GetCurrenciesString")["returns"]) {
		output := asMap(v)
		if strings.ReplaceAll(asString(output["type"]), " ", "") == "number[]|table[]" {
			output["type"] = "unknown"
		}
	}
	argumentNamed(c.global("GetCurrenciesString"), "params", "currencies")["type"] = "(number|table)[]"
	asMap(asList(c.global("GetCurrenciesString")["returns"])[0])["type"] = "string?"
	c.aliases["KStringDiscordUserName"] = M{"expression": "string", "source": "Blizzard Discord user-name text carrier"}
	for _, note := range []M{{"symbol": "CreateFrame", "rule": "return common Frame; template composition requires reviewed assertion"}, {"symbol": "MenuUtil.CreateEnumRadioMenu", "rule": "enum is a string-keyed table of V, iterated through pairs by FrameXML"}, {"symbol": "AceComm-3.0.SendCommMessage.callbackArg", "rule": "single optional arbitrary callback argument, not a variadic type pack"}, {"symbol": "SetBindingMacro.name", "rule": "repair malformed annotation to macro name or index"}, {"symbol": "GetCurrenciesString", "rule": "accept numeric IDs, keyed currency entries or positional argument tables; an empty list returns nil"}} {
		c.notes = append(c.notes, note)
	}
	call := c.global("LibStub")
	call["method"] = true
	asMap(c.definition("LibStub")["methods"])["metamethod __call"] = call
	c.setGlobal("LibStub", M{"type_name": "LibStub", "source": call["source"]})
	for _, v := range asList(overrides["secret_effects"]) {
		effect := asMap(v)
		for _, symbol := range stringsOf(effect["symbols"]) {
			item := c.symbolItem(symbol)
			variants := append([]any{item}, asList(item["variants"])...)
			for _, v := range variants {
				asMap(v)["secret_effect"] = true
			}
			c.note(symbol, asString(effect["reason"]), M{"evidence": effect["evidence"]})
		}
	}
	for _, v := range asList(overrides["native_contracts"]) {
		correction := asMap(v)
		for _, symbol := range stringsOf(correction["symbols"]) {
			item := c.symbolItem(symbol)
			if value, ok := correction["signature"]; ok {
				item["resolved_signature"] = value
			}
			if value, ok := correction["returns"]; ok {
				item["Returns"] = cloneValue(value)
			}
			for argument, typeName := range asMap(correction["argument_types"]) {
				argumentNamed(item, "Arguments", argument)["Type"] = typeName
			}
			c.note(symbol, asString(correction["reason"]), M{"evidence": correction["evidence"]})
		}
	}
	c.applyObservedReceiverContracts(overrides)
}
func (c *catalogue) symbolItem(symbol string) M {
	index := strings.LastIndexByte(symbol, '.')
	if index < 0 {
		return c.global(symbol)
	}
	return c.method(symbol[:index], symbol[index+1:])
}
func (c *catalogue) applyObservedReceiverContracts(overrides M) {
	exclusions, hierarchies := asList(overrides["supplemental_method_exclusions"]), asList(overrides["receiver_hierarchy"])
	if len(exclusions) == 0 && len(hierarchies) == 0 {
		return
	}
	observed := mustMap(c.project.LoadObservations())
	for _, v := range exclusions {
		rule := asMap(v)
		owner, name := asString(rule["owner"]), asString(rule["method"])
		symbol := owner + "." + name
		item := asMap(asMap(c.definition(owner)["methods"])[name])
		if item == nil || truth(item["native"]) || asString(item["source"]) != asString(rule["source"]) || c.signatureValue(item, owner) != asString(rule["signature"]) {
			reject("review changed supplemental exclusion contract: %s", symbol)
		}
		source := asString(rule["source"])
		relative := strings.TrimPrefix(source, "questie:")
		if !strings.HasPrefix(source, "questie:") || filepath.IsAbs(relative) || containsString(strings.Split(filepath.ToSlash(relative), "/"), "..") {
			reject("invalid supplemental exclusion source: %s", symbol)
		}
		if sha256hex(mustFile(c.project.InputPath("questie", relative))) != asString(rule["source_sha256"]) || asString(asMap(observed["sources"])["build"]) != asString(rule["build"]) {
			reject("review changed supplemental exclusion provenance: %s", symbol)
		}
		cases := stringsOf(rule["negative_cases"])
		if len(cases) == 0 {
			reject("supplemental exclusion lacks negative native observations: %s", symbol)
		}
		for _, name := range cases {
			if asString(asMap(asMap(observed["hierarchy"])[name])["state"]) != "absent" {
				reject("supplemental exclusion lacks negative native observations: %s", symbol)
			}
		}
		c.excludedMethods[symbol] = withFields(rule, M{"original": mapCopy(item), "observation_sources": observed["sources"], "current_source_cohort_matches": observed["applicable"]})
		delete(asMap(c.definition(owner)["methods"]), name)
		c.note(symbol, "exclude unsupported common supplemental guarantee", M{"reason": rule["reason"], "evidence": rule["evidence"]})
	}
	for _, v := range hierarchies {
		rule := asMap(v)
		owner := asString(rule["owner"])
		if requireStringJSON(getOr(c.definition(owner), "bases", []any{})) != requireStringJSON(rule["expected_bases"]) {
			reject("review changed receiver hierarchy: %s", owner)
		}
		for _, base := range stringsOf(rule["bases"]) {
			if c.defs[base] == nil {
				reject("review changed receiver hierarchy: %s", owner)
			}
		}
		c.definition(owner)["bases"] = cloneValue(rule["bases"])
		c.note(owner, asString(rule["reason"]), M{"evidence": rule["evidence"], "observed_build": asMap(observed["sources"])["build"], "current_source_cohort_matches": observed["applicable"]})
	}
}
