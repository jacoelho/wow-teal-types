
local _, private = ...





local function pack(...)
   return { count = select("#", ...), values = { ... } }
end

local function numeric(value)
   if issecretvalue(value) or not canaccessvalue(value) or not (type(value) == "number") then return nil end
   if value ~= value or value == math.huge or value == -math.huge then return nil end
   return value
end
local function text(value)
   if issecretvalue(value) or not canaccessvalue(value) or not (type(value) == "string") then return nil end
   if #value > 4096 then return nil end
   return value
end
local function boolean(value)
   if issecretvalue(value) or not canaccessvalue(value) or not (type(value) == "boolean") then return nil end
   return value
end
local function same(left, right)
   if issecretvalue(left) or issecretvalue(right) or not canaccessallvalues(left, right) then return false end
   return left == right
end
local function arrayCount(value)
   if issecretvalue(value) or not canaccessvalue(value) or type(value) ~= "table" or not canaccesstable(value) then return nil end
   local array = value
   local count = #array
   if count > 16 then return nil end
   for index = 1, count do
      local item = array[index]
      if issecretvalue(item) or not canaccessvalue(item) or not (type(item) == "number") then return nil end
   end
   return count
end



private.RunInvocations = function(host, result)
   local objects = {}
   local states = {}
   local bindings = {}
   local signalMap
   local rows = {}
   local seen = {}
   local diagnostics = {}
   for _, id in ipairs(private.Plan.bindingDiagnosticCases) do
      local row = { caseID = id, state = "not_run" }
      result.bindingDiagnostics[#result.bindingDiagnostics + 1] = row
      diagnostics[id] = row
   end
   for _, id in ipairs(private.Plan.bindingExperimentCases) do
      local row = { caseID = id, state = "not_run" }
      result.bindingExperimentChecks[#result.bindingExperimentChecks + 1] = row
      diagnostics[id] = row
   end
   local samples = {}
   for _, id in ipairs(private.Plan.bindingTextSampleCases) do
      local row = { caseID = id, state = "not_run" }
      result.bindingTextSamples[#result.bindingTextSamples + 1] = row
      samples[id] = row
   end
   local function call(object, name, args, argumentCount)
      local fn, state = host.field(object, name)
      if state ~= "available" then return nil, state end
      if issecretvalue(fn) or not canaccessvalue(fn) or type(fn) ~= "function" then return nil, "not_function" end
      local arguments = args or {}
      local packed = pack(pcall((fn), object, unpack(arguments, 1, argumentCount or #arguments)))
      if packed.values[1] ~= true then return nil, "call_error" end
      local values = {}
      for index = 2, packed.count do values[index - 1] = packed.values[index] end
      return { count = packed.count - 1, values = values }, "called"
   end
   local function value(object, name)
      local packed = call(object, name)
      return packed and packed.values[1]
   end
   local function apply(object, name, ...)
      local _, state = call(object, name, { ... })
      return state == "called"
   end
   local function diagnosticSet(id, object, name, args, argumentCount)
      local packed, state = call(object, name, args, argumentCount)
      diagnostics[id].state = state == "called" and (packed.count == 0 and "called_void" or "unexpected_tuple") or state
   end
   local function readText(object, name)
      local packed, state = call(object, name)
      if state ~= "called" then return nil, state end
      if packed.count ~= 1 then return nil, "unexpected_tuple" end
      local item = packed.values[1]
      if issecretvalue(item) then return nil, "secret" end
      if not canaccessvalue(item) then return nil, "inaccessible" end
      if item == nil then return nil, "nil" end
      if not (type(item) == "string") then return nil, "wrong_type" end
      if #item > 4096 then return nil, "oversized" end
      return item, #item == 0 and "public_empty" or "public_nonempty"
   end
   local function diagnosticText(id, object, name)
      local item, state = readText(object, name)
      diagnostics[id].state = state
      return item
   end
   local function sampleText(id, item, state)
      local row = samples[id]
      row.state = state
      if state ~= "public_empty" and state ~= "public_nonempty" then return end



      if issecretvalue(item) then row.state = "secret"; return end
      if not canaccessvalue(item) then row.state = "inaccessible"; return end
      if not (type(item) == "string") then row.state = "wrong_type"; return end
      if #item > 128 then row.state = "oversized"; return end
      local hex = ""
      local firstNul = 0
      for index = 1, #item do
         local byte = string.byte(item, index)
         hex = hex .. string.format("%02x", byte)
         if byte == 0 and firstNul == 0 then firstNul = index end
      end
      row.state, row.hex, row.byteLength, row.firstNul = "public", hex, #item, firstNul
   end
   local function sampleRead(id, object, name)
      local item, state = readText(object, name)
      sampleText(id, item, state)
      return item
   end
   local function diagnosticCompare(id, left, right)


      if issecretvalue(left) or issecretvalue(right) or not canaccessallvalues(left, right) then
         diagnostics[id].state = "not_comparable"
      elseif not (type(left) == "string") or not (type(right) == "string") then diagnostics[id].state = "not_comparable"
      else diagnostics[id].state = left == right and "matched" or "mismatch" end
   end
   local function sampleCompare(id, leftID, rightID, left, right)
      if samples[leftID].state ~= "public" or samples[rightID].state ~= "public" then
         diagnostics[id].state = "not_comparable"
      else diagnosticCompare(id, left, right) end
   end
   local function diagnosticFact(id, object, name, expected, kind)
      local packed, state = call(object, name)
      if state ~= "called" or packed.count ~= 1 then diagnostics[id].state = "not_comparable"; return end
      local item = packed.values[1]
      if issecretvalue(item) or issecretvalue(expected) or not canaccessallvalues(item, expected) then diagnostics[id].state = "not_comparable"; return end
      if kind == "number" then
         if not (type(item) == "number") or item ~= item or item == math.huge or item == -math.huge then diagnostics[id].state = "not_comparable"; return end
      else
         local itemKind, expectedKind = type(item), type(expected)
         if (itemKind ~= "table" and itemKind ~= "userdata") or (expectedKind ~= "table" and expectedKind ~= "userdata") then diagnostics[id].state = "not_comparable"; return end
      end
      diagnostics[id].state = item == expected and "matched" or "mismatch"
   end
   local function fixture(namespace, name, ...)
      local fn = host.resolve(namespace, name)
      local object = host.create(fn, ...)
      return object
   end
   local function run(canonical, object, name, args, check, oracle)
      local id = canonical .. "." .. name
      if seen[id] then error("duplicate authored probe case") end
      seen[id] = true
      local row = { caseID = id, state = "not_run", invoked = false, returnCount = 0, returns = {}, check = check, behaviorVerified = false }
      rows[#rows + 1] = row
      local packed, state = call(object, name, args)
      row.state = state
      if state ~= "called" then return end
      row.invoked, row.returnCount = true, packed.count
      if packed.count > 8 then row.state = "unsupported_tuple"; return end
      for index = 1, packed.count do
         local item = packed.values[index]
         if issecretvalue(item) then row.returns[index] = "secret"; row.state = "secret_result"
         elseif not canaccessvalue(item) then row.returns[index] = "inaccessible"; row.state = "inaccessible_result"
         else row.returns[index] = type(item) end
      end
      if row.state ~= "called" then return end
      local ok, matched = pcall(oracle, packed.values, packed.count)
      row.behaviorVerified = ok and matched == true
      row.state = row.behaviorVerified and "passed" or "property_mismatch"
   end
   local function stringResult(values, count)
      local decoded = text(values[1])
      return count == 1 and decoded ~= nil and #decoded > 0
   end
   local function exactNumber(expected)
      return function(values, count) return count == 1 and numeric(values[1]) == expected end
   end
   local function exactText(expected)
      return function(values, count) return count == 1 and text(values[1]) == expected end
   end
   local function exactBoolean(expected)
      return function(values, count) return count == 1 and boolean(values[1]) == expected end
   end
   local ok = pcall(function()
      local binding = host.construct("DurationTextBindingObjectAPI", objects, states)
      if not issecretvalue(binding) and canaccessvalue(binding) and binding ~= nil then bindings[#bindings + 1] = binding end
      local clock = host.construct("LuaDurationManualClockAPI", objects, states)
      local duration = fixture("C_DurationUtil", "CreateDuration")
      local formatter = fixture("C_StringUtil", "CreateSecondsFormatter")
      local curve = fixture("C_CurveUtil", "CreateColorCurve")
      local font = host.uiObject("font")
      local fontObject = host.resolve("GameFontNormal")
      diagnosticSet("setup.font_object_set", font, "SetFontObject", { fontObject })
      diagnosticSet("setup.clock_set_time", clock, "SetTime", { 12.5 })
      diagnosticSet("setup.duration_set_clock", duration, "SetClock", { clock })
      diagnosticSet("setup.duration_set_span", duration, "SetTimeSpan", { 0, 60 })
      diagnosticFact("setup.font_object_identity", font, "GetFontObject", fontObject, "object")
      diagnosticFact("setup.duration_clock_identity", duration, "GetClock", clock, "object")
      diagnosticFact("setup.duration_clock_time", duration, "GetClockTime", 12.5, "number")
      diagnosticFact("setup.duration_remaining", duration, "GetRemainingDuration", 47.5, "number")

      apply(curve, "AddPoint", 0, { r = 1, g = 1, b = 1, a = 1 })
      apply(curve, "AddPoint", 1, { r = 1, g = 1, b = 1, a = 1 })
      local canonical = "DurationTextBindingObject"
      run(canonical, binding, "SetToDefaults", {}, "clears_duration", function(_, count)
         return count == 0 and same(value(binding, "GetDuration"), nil)
      end)
      run(canonical, binding, "SetDuration", { duration }, "duration_identity", function(_, count) return count == 0 and same(value(binding, "GetDuration"), duration) end)
      run(canonical, binding, "GetDuration", {}, "duration_identity", function(values, count) return count == 1 and same(values[1], duration) end)
      run(canonical, binding, "SetFontString", { font }, "font_identity", function(_, count) return count == 0 and same(value(binding, "GetFontString"), font) end)
      run(canonical, binding, "GetFontString", {}, "font_identity", function(values, count) return count == 1 and same(values[1], font) end)
      run(canonical, binding, "SetFormatter", { formatter }, "can_format", function(_, count) return count == 0 and boolean(value(binding, "CanFormatText")) == true end)
      run(canonical, binding, "SetTextFormat", { "{}", { { property = 0, formatter = formatter } } }, "can_format_components", function(_, count) return count == 0 and boolean(value(binding, "CanFormatText")) == true end)
      run(canonical, binding, "CanFormatText", {}, "configured_true", exactBoolean(true))
      run(canonical, binding, "CanUpdateFontString", {}, "configured_true", exactBoolean(true))
      run(canonical, binding, "SetExpiredText", { "EXPIRED" }, "text_roundtrip", function(_, count) return count == 0 and text(value(binding, "GetExpiredText")) == "EXPIRED" end)
      run(canonical, binding, "GetExpiredText", {}, "text_roundtrip", exactText("EXPIRED"))
      run(canonical, binding, "SetZeroDurationText", { "ZERO" }, "text_roundtrip", function(_, count) return count == 0 and text(value(binding, "GetZeroDurationText")) == "ZERO" end)
      run(canonical, binding, "GetZeroDurationText", {}, "text_roundtrip", exactText("ZERO"))
      run(canonical, binding, "SetTimeModifier", { 1 }, "modifier_roundtrip", function(_, count) return count == 0 and numeric(value(binding, "GetTimeModifier")) == 1 end)
      run(canonical, binding, "GetTimeModifier", {}, "modifier_roundtrip", exactNumber(1))
      run(canonical, binding, "SetUpdateInterval", { 0.25 }, "interval_roundtrip", function(_, count) return count == 0 and numeric(value(binding, "GetUpdateInterval")) == 0.25 end)
      run(canonical, binding, "GetUpdateInterval", {}, "interval_roundtrip", exactNumber(0.25))
      run(canonical, binding, "SetTextColorCurve", { curve, 0 }, "curve_property_tuple", function(_, count)
         local returned = call(binding, "GetTextColorCurve")
         return count == 0 and returned ~= nil and returned.count == 2 and same(returned.values[1], curve) and numeric(returned.values[2]) == 0
      end)
      run(canonical, binding, "GetTextColorCurve", {}, "curve_property_tuple", function(values, count) return count == 2 and same(values[1], curve) and numeric(values[2]) == 0 end)
      run(canonical, binding, "GetFormattedTextColor", {}, "public_white_rgba", function(values, count)
         local color = values[1]
         local r = host.field(color, "r")
         local g = host.field(color, "g")
         local b = host.field(color, "b")
         local a = host.field(color, "a")
         return count == 1 and numeric(r) == 1 and numeric(g) == 1 and numeric(b) == 1 and numeric(a) == 1
      end)
      run(canonical, binding, "ClearTextColorCurve", {}, "cleared_curve_tuple", function(_, count)
         local returned = call(binding, "GetTextColorCurve")
         return count == 0 and returned ~= nil and returned.count == 0
      end)
      run(canonical, binding, "GetFormattedText", {}, "public_nonempty_text", stringResult)
      run(canonical, binding, "Enable", {}, "enabled_true", function(_, count) return count == 0 and boolean(value(binding, "IsEnabled")) == true end)
      run(canonical, binding, "IsEnabled", {}, "enabled_true", exactBoolean(true))
      local expectedBefore = diagnosticText("baseline.expected_before", binding, "GetFormattedText")
      run(canonical, binding, "UpdateFontString", {}, "font_text_matches", function(_, count) local expected = text(value(binding, "GetFormattedText")); return count == 0 and expected ~= nil and text(value(font, "GetText")) == expected end)
      local expectedAfter = diagnosticText("baseline.expected_after", binding, "GetFormattedText")
      local actual = diagnosticText("baseline.font_after", font, "GetText")
      sampleText("baseline.expected", expectedAfter, diagnostics["baseline.expected_after"].state)
      sampleText("baseline.font", actual, diagnostics["baseline.font_after"].state)
      local baselineExpected = expectedAfter
      diagnosticCompare("baseline.expected_stable", expectedBefore, expectedAfter)
      diagnosticCompare("baseline.font_matches", actual, expectedAfter)


      diagnosticSet("sentinel.set", font, "SetText", { "PROBE_SENTINEL" })
      actual = diagnosticText("sentinel.read", font, "GetText")
      diagnosticCompare("sentinel.matches", actual, "PROBE_SENTINEL")
      diagnosticSet("retry.update", binding, "UpdateFontString", {})
      expectedAfter = diagnosticText("retry.expected", binding, "GetFormattedText")
      actual = diagnosticText("retry.font", font, "GetText")
      diagnosticCompare("retry.matches", actual, expectedAfter)
      diagnosticSet("zero_interval.set", binding, "SetUpdateInterval", { 0 })
      diagnosticSet("zero_interval.update", binding, "UpdateFontString", {})
      expectedAfter = diagnosticText("zero_interval.expected", binding, "GetFormattedText")
      actual = diagnosticText("zero_interval.font", font, "GetText")
      diagnosticCompare("zero_interval.matches", actual, expectedAfter)


      if not issecretvalue(baselineExpected) and canaccessvalue(baselineExpected) and type(baselineExpected) == "string" and #baselineExpected <= 128 then
         diagnosticSet("projection.set", font, "SetText", { baselineExpected })
      else diagnostics["projection.set"].state = "not_run" end
      actual = sampleRead("projection.font", font, "GetText")
      sampleCompare("projection.matches", "projection.font", "baseline.expected", actual, baselineExpected)
      diagnosticSet("manual_step.set_time", clock, "SetTime", { 20 })
      diagnosticFact("manual_step.clock_time", duration, "GetClockTime", 20, "number")
      diagnosticFact("manual_step.remaining", duration, "GetRemainingDuration", 40, "number")
      diagnosticSet("manual_step.update", binding, "UpdateFontString", {})
      expectedAfter = sampleRead("manual_step.expected", binding, "GetFormattedText")
      actual = sampleRead("manual_step.font", font, "GetText")
      sampleCompare("manual_step.matches", "manual_step.font", "manual_step.expected", actual, expectedAfter)
      diagnosticSet("manual_step.restore_time", clock, "SetTime", { 12.5 })
      local nowFn = host.resolve("GetTime")
      local now
      if not issecretvalue(nowFn) and canaccessvalue(nowFn) and type(nowFn) == "function" then
         local okNow, rawNow = pcall((nowFn))
         if okNow and not issecretvalue(rawNow) and canaccessvalue(rawNow) and type(rawNow) == "number" then
            if rawNow == rawNow and rawNow ~= math.huge and rawNow ~= -math.huge then now = rawNow end
         end
      end
      if now == nil then diagnostics["default_clock.time"].state = "not_comparable"
      else
         diagnostics["default_clock.time"].state = "matched"


         diagnosticSet("default_clock.clear", duration, "SetClock", {}, 1)
         diagnosticSet("default_clock.span", duration, "SetTimeSpan", { now, now + 60 })
         diagnosticFact("default_clock.clock_time", duration, "GetClockTime", now, "number")
         diagnosticFact("default_clock.remaining", duration, "GetRemainingDuration", 60, "number")
         diagnosticSet("default_clock.update", binding, "UpdateFontString", {})
         expectedAfter = sampleRead("default_clock.expected", binding, "GetFormattedText")
         actual = sampleRead("default_clock.font", font, "GetText")
         sampleCompare("default_clock.matches", "default_clock.font", "default_clock.expected", actual, expectedAfter)
      end
      run(canonical, binding, "HasSecretValues", {}, "public_fixture_false", exactBoolean(false))
      run(canonical, binding, "SetEnabled", { false }, "enabled_false", function(_, count) return count == 0 and boolean(value(binding, "IsEnabled")) == false end)
      run(canonical, binding, "Disable", {}, "enabled_false", function(_, count) return count == 0 and boolean(value(binding, "IsEnabled")) == false end)
      local copied
      run(canonical, binding, "Copy", {}, "copy_preserves_text", function(values, count)
         copied = values[1]
         if issecretvalue(copied) or not canaccessvalue(copied) or copied == nil then return false end
         bindings[#bindings + 1] = copied
         return count == 1 and text(value(copied, "GetExpiredText")) == "EXPIRED"
      end)
      apply(copied, "SetExpiredText", "ASSIGNED")
      run(canonical, binding, "Assign", { copied }, "assignment_preserves_text", function(_, count) return count == 0 and text(value(binding, "GetExpiredText")) == "ASSIGNED" end)

      run("LuaDurationClockObject", clock, "GetTime", {}, "manual_time_12_5", exactNumber(12.5))
      canonical = "LuaDurationManualClock"
      run(canonical, clock, "SetTime", { 12.5 }, "manual_time_12_5", function(_, count) return count == 0 and numeric(value(clock, "GetTime")) == 12.5 end)
      run(canonical, clock, "AdvanceTime", { 3 }, "manual_time_15_5", function(_, count) return count == 0 and numeric(value(clock, "GetTime")) == 15.5 end)
      run(canonical, clock, "RewindTime", { 1 }, "manual_time_14_5", function(_, count) return count == 0 and numeric(value(clock, "GetTime")) == 14.5 end)
      run(canonical, clock, "ResetTime", {}, "manual_time_zero", function(_, count) return count == 0 and numeric(value(clock, "GetTime")) == 0 end)

      local locale = host.construct("LuaLocaleContextAPI", objects, states)
      canonical = "LuaLocaleContext"
      run(canonical, locale, "SetLocale", { "en_US" }, "locale_accepted", exactBoolean(true))
      run(canonical, locale, "GetLocale", {}, "en_us_locale", function(values, count) local decoded = text(values[1]); return count == 1 and (decoded == "en_US" or decoded == "enUS") end)
      run(canonical, locale, "CompareStrings", { "abc", "abc", 4 }, "equal_collation_zero", exactNumber(0))
      run(canonical, locale, "FindBreaks", { "abc", 0 }, "four_character_boundaries", function(values, count) return count == 1 and arrayCount(values[1]) == 4 end)
      run(canonical, locale, "FindStringMatches", { "abc abc", "abc", 4 }, "two_nonoverlapping_matches", function(values, count) return count == 1 and arrayCount(values[1]) == 2 end)
      run(canonical, locale, "FoldCase", { "ABC" }, "ascii_fold", exactText("abc"))
      run(canonical, locale, "FormatCurrency", { 12.5, "USD" }, "public_currency_text", stringResult)
      run(canonical, locale, "FormatDate", { 0, 3, "UTC" }, "public_date_text", stringResult)
      run(canonical, locale, "FormatDateTime", { 0, 3, 1, "UTC" }, "public_datetime_text", stringResult)
      run(canonical, locale, "FormatNumber", { 12.5, 0 }, "ascii_decimal", exactText("12.5"))
      run(canonical, locale, "FormatTime", { 0, 1, "UTC" }, "public_time_text", stringResult)
      run(canonical, locale, "GetCurrencyName", { "USD", 0 }, "public_currency_name", stringResult)
      run(canonical, locale, "GetDisplayName", { "en_US" }, "public_display_name", stringResult)
      run(canonical, locale, "GetSortKey", { "abc", 4 }, "public_nonempty_sort_key", stringResult)
      run(canonical, locale, "Length", { "abc" }, "three_characters", exactNumber(3))
      run(canonical, locale, "ParseCurrency", { "$12.50" }, "currency_amount_code", function(values, count)
         local amount = host.field(values[1], "amount")
         local code = host.field(values[1], "currencyCode")
         return count == 1 and numeric(amount) == 12.5 and text(code) == "USD"
      end)
      run(canonical, locale, "ParseNumber", { "12.5", 0 }, "decimal_parse", exactNumber(12.5))
      run(canonical, locale, "SelectPlural", { 1, 0 }, "singular_one", exactText("one"))
      run(canonical, locale, "ToLower", { "ABC" }, "ascii_lower", exactText("abc"))
      run(canonical, locale, "ToTitle", { "hello world" }, "ascii_title", exactText("Hello World"))
      run(canonical, locale, "ToUpper", { "abc" }, "ascii_upper", exactText("ABC"))
      run(canonical, locale, "TransformLocale", { 3 }, "language_en", exactText("en"))

      signalMap = host.construct("TimedSignalMapAPI", objects, states)
      canonical = "TimedSignalMap"
      local nowFn = host.resolve("GetTime")
      local okNow, now = pcall((nowFn))
      local deadline = okNow and numeric(now)
      if deadline then deadline = deadline + 86400 end
      run(canonical, signalMap, "SignalAt", { 101, deadline }, "scheduled_key_present", function(_, count) local returned = call(signalMap, "HasSignal", { 101 }); return count == 0 and returned ~= nil and boolean(returned.values[1]) == true end)
      run(canonical, signalMap, "SignalAfter", { 202, 172800 }, "scheduled_key_present", function(_, count) local returned = call(signalMap, "HasSignal", { 202 }); return count == 0 and returned ~= nil and boolean(returned.values[1]) == true end)
      run(canonical, signalMap, "HasSignal", { 101 }, "scheduled_key_present", exactBoolean(true))
      run(canonical, signalMap, "GetSignalCount", {}, "two_scheduled_keys", exactNumber(2))
      run(canonical, signalMap, "GetSignalTime", { 101 }, "absolute_deadline", function(values, count) return count == 1 and numeric(values[1]) == deadline end)
      run(canonical, signalMap, "GetNextSignal", {}, "key_time_tuple", function(values, count) return count == 2 and numeric(values[1]) == 101 and numeric(values[2]) == deadline end)
      run(canonical, signalMap, "CancelSignal", { 101 }, "one_key_remains", function(_, count) return count == 0 and numeric(value(signalMap, "GetSignalCount")) == 1 end)
      run(canonical, signalMap, "CancelAllSignals", {}, "zero_keys_remain", function(_, count) return count == 0 and numeric(value(signalMap, "GetSignalCount")) == 0 end)

      local radial = host.construct("SimpleAnimRadialProgressAPI", objects, states)
      canonical = "RadialProgress"
      run(canonical, radial, "SetFromPercent", { 0.25 }, "from_roundtrip", function(_, count) return count == 0 and numeric(value(radial, "GetFromPercent")) == 0.25 end)
      run(canonical, radial, "SetToPercent", { 0.75 }, "to_roundtrip", function(_, count) return count == 0 and numeric(value(radial, "GetToPercent")) == 0.75 end)
      run(canonical, radial, "GetFromPercent", {}, "from_roundtrip", exactNumber(0.25))
      run(canonical, radial, "GetToPercent", {}, "to_roundtrip", exactNumber(0.75))
      for _, row in ipairs(host.observeHierarchy(radial, canonical)) do result.hierarchy[#result.hierarchy + 1] = row end

      local vector = host.construct("SimpleVectorGraphicsAPI", objects, states)
      canonical = "VectorGraphics"
      run(canonical, vector, "SetSVG", { "Interface\\AddOns\\WoWAPIProbe\\Fixture.svg" }, "owned_svg_accepted", exactBoolean(true))
      run(canonical, vector, "GetSVGFileID", {}, "public_fileid_carrier", function(values, count) return count == 1 and numeric(values[1]) ~= nil end)
      run(canonical, vector, "HasSVG", {}, "owned_svg_present", exactBoolean(true))
      run(canonical, vector, "ClearSVG", {}, "svg_cleared", function(_, count) return count == 0 and boolean(value(vector, "HasSVG")) == false end)
      for _, row in ipairs(host.observeHierarchy(vector, canonical)) do result.hierarchy[#result.hierarchy + 1] = row end
   end)

   for index, binding in ipairs(bindings) do
      local _, state = call(binding, "Disable")
      result.cleanup[#result.cleanup + 1] = { caseID = "cleanup.invocation.binding." .. tostring(index), state = state == "called" and "complete" or state }
   end
   if not issecretvalue(signalMap) and canaccessvalue(signalMap) and signalMap ~= nil then
      local _, state = call(signalMap, "CancelAllSignals")
      result.cleanup[#result.cleanup + 1] = { caseID = "cleanup.invocation.signals", state = state == "called" and "complete" or state }
   end
   result.invocations = rows
   local complete = ok
   for _, row in ipairs(rows) do
      if row.invoked then result.invokedMethods = result.invokedMethods + 1 end
      if row.behaviorVerified then result.verifiedMethods = result.verifiedMethods + 1 end
      if row.caseID == "VectorGraphics.SetSVG" and row.state ~= "passed" then row.state = "fixture_unavailable" end
      if row.state ~= "passed" then complete = false end
   end
   for _, family in ipairs(private.Plan.families) do
      for _, name in ipairs(family.methods) do
         local id = family.canonical .. "." .. name
         if not seen[id] then
            complete = false
            rows[#rows + 1] = { caseID = id, state = "fixture_setup_failed", invoked = false, returnCount = 0, returns = {}, check = "not_run", behaviorVerified = false }
         end
      end
   end
   for _, row in ipairs(result.cleanup) do if row.state ~= "complete" then complete = false end end
   result.invocationCaptureComplete = complete and #rows == private.Plan.expectedMethods
   result.outcome = result.invocationCaptureComplete and "public_invocations_verified" or "public_invocations_incomplete"
end
