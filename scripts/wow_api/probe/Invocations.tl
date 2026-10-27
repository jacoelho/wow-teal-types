local type W = require("wow")
local type T = require("probe_types")
local _, private = ... as (string, T.Namespace)

-- Preserve actual native return arity before applying typed property oracles.
local function pack(...: unknown): T.Packed
   return {count = select("#", ...), values = {...}}
end

local function numeric(value: unknown): number
   if issecretvalue(value) or not canaccessvalue(value) or not (value is number) then return nil end
   if value ~= value or value == math.huge or value == -math.huge then return nil end
   return value
end
local function text(value: unknown): string
   if issecretvalue(value) or not canaccessvalue(value) or not (value is string) then return nil end
   if #value > 4096 then return nil end
   return value
end
local function boolean(value: unknown): boolean
   if issecretvalue(value) or not canaccessvalue(value) or not (value is boolean) then return nil end
   return value
end
local function same(left: unknown, right: unknown): boolean
   if issecretvalue(left) or issecretvalue(right) or not canaccessallvalues(left, right) then return false end
   return left == right
end
local function arrayCount(value: {number}): integer
   if issecretvalue(value) or not canaccessvalue(value) or type(value) ~= "table" or not canaccesstable(value) then return nil end
   local array = value
   local count: integer = #array
   if count > 16 then return nil end
   for index = 1, count do
      local item = array[index]
      if issecretvalue(item) or not canaccessvalue(item) or not (item is number) then return nil end
   end
   return count
end

local function packBinding(...: W.DurationTextBindingObject): T.Packed
   local packed = pack(...)
   packed.binding = (...)
   return packed
end

local function packDuration(...: W.DurationObject): T.Packed
   local packed = pack(...)
   packed.duration = (...)
   return packed
end

local function packClock(...: W.LuaDurationClockObject): T.Packed
   local packed = pack(...)
   packed.clock = (...)
   return packed
end

local function packFont(...: W.FontString): T.Packed
   local packed = pack(...)
   packed.font = (...)
   return packed
end

local function packFontObject(...: W.Font): T.Packed
   local packed = pack(...)
   packed.fontObject = (...)
   return packed
end

local function packCurve(...: W.ColorCurveObject | number | nil): T.Packed
   local packed = pack(...)
   local first = ...
   if not issecretvalue(first) and canaccessvalue(first) and not (first is number) and not (first is nil) then packed.curve = first end
   return packed
end

local function packColor(...: W.colorRGBA): T.Packed
   local packed = pack(...)
   packed.color = (...)
   return packed
end

local function packCurrency(...: W.CurrencyParseResult): T.Packed
   local packed = pack(...)
   packed.currency = (...)
   return packed
end

local function packNumbers(...: {number}): T.Packed
   local packed = pack(...)
   packed.numbers = (...)
   return packed
end

-- Oracles receive public scalar facts and retained typed handles. Opaque tuple
-- transport stays inside the observation boundary, never an external callback.
local function checkedResult(packed: T.Packed): T.CheckedResult
   local checked: T.CheckedResult = {
      count = packed.count, values = {}, binding = packed.binding,
      duration = packed.duration, clock = packed.clock, font = packed.font,
      fontObject = packed.fontObject, curve = packed.curve, numbers = packed.numbers,
   }
   for index = 1, packed.count do
      local item = packed.values[index]
      if not issecretvalue(item) and canaccessvalue(item) then
         if item is number then checked.values[index] = item
         elseif item is string then checked.values[index] = item
         elseif item is boolean then checked.values[index] = item end
      end
   end
   local color = packed.color
   if not issecretvalue(color) and canaccessvalue(color) and type(color) == "table" and canaccesstable(color) then
      checked.color = {r = numeric(color.r), g = numeric(color.g), b = numeric(color.b), a = numeric(color.a)}
   end
   local currency = packed.currency
   if not issecretvalue(currency) and canaccessvalue(currency) and type(currency) == "table" and canaccesstable(currency) then
      checked.currency = {amount = numeric(currency.amount), currencyCode = text(currency.currencyCode)}
   end
   return checked
end

local record Operations
   ColorCurveObjectAddPoint: function(W.ColorCurveObject, number, T.PublicColor): T.Operation
   DurationObjectGetClock: function(W.DurationObject): T.Operation
   DurationObjectGetClockTime: function(W.DurationObject): T.Operation
   DurationObjectGetRemainingDuration: function(W.DurationObject): T.Operation
   DurationObjectSetClock: function(W.DurationObject, ? W.LuaDurationClockObject): T.Operation
   DurationObjectSetTimeSpan: function(W.DurationObject, number, number): T.Operation
   DurationTextBindingObjectAssign: function(W.DurationTextBindingObject, W.DurationTextBindingObject): T.Operation
   DurationTextBindingObjectCanFormatText: function(W.DurationTextBindingObject): T.Operation
   DurationTextBindingObjectCanUpdateFontString: function(W.DurationTextBindingObject): T.Operation
   DurationTextBindingObjectClearTextColorCurve: function(W.DurationTextBindingObject): T.Operation
   DurationTextBindingObjectCopy: function(W.DurationTextBindingObject): T.Operation
   DurationTextBindingObjectDisable: function(W.DurationTextBindingObject): T.Operation
   DurationTextBindingObjectEnable: function(W.DurationTextBindingObject): T.Operation
   DurationTextBindingObjectGetDuration: function(W.DurationTextBindingObject): T.Operation
   DurationTextBindingObjectGetExpiredText: function(W.DurationTextBindingObject): T.Operation
   DurationTextBindingObjectGetFontString: function(W.DurationTextBindingObject): T.Operation
   DurationTextBindingObjectGetFormattedText: function(W.DurationTextBindingObject): T.Operation
   DurationTextBindingObjectGetFormattedTextColor: function(W.DurationTextBindingObject): T.Operation
   DurationTextBindingObjectGetTextColorCurve: function(W.DurationTextBindingObject): T.Operation
   DurationTextBindingObjectGetTimeModifier: function(W.DurationTextBindingObject): T.Operation
   DurationTextBindingObjectGetUpdateInterval: function(W.DurationTextBindingObject): T.Operation
   DurationTextBindingObjectGetZeroDurationText: function(W.DurationTextBindingObject): T.Operation
   DurationTextBindingObjectHasSecretValues: function(W.DurationTextBindingObject): T.Operation
   DurationTextBindingObjectIsEnabled: function(W.DurationTextBindingObject): T.Operation
   DurationTextBindingObjectSetDuration: function(W.DurationTextBindingObject, W.DurationObject): T.Operation
   DurationTextBindingObjectSetEnabled: function(W.DurationTextBindingObject, boolean): T.Operation
   DurationTextBindingObjectSetExpiredText: function(W.DurationTextBindingObject, ? string): T.Operation
   DurationTextBindingObjectSetFontString: function(W.DurationTextBindingObject, W.FontString): T.Operation
   DurationTextBindingObjectSetFormatter: function(W.DurationTextBindingObject, W.NumericFormatter): T.Operation
   DurationTextBindingObjectSetTextColorCurve: function(W.DurationTextBindingObject, W.ColorCurveObject, number): T.Operation
   DurationTextBindingObjectSetTextFormat: function(W.DurationTextBindingObject, string, {W.DurationTextBindingFormatComponent}): T.Operation
   DurationTextBindingObjectSetTimeModifier: function(W.DurationTextBindingObject, number): T.Operation
   DurationTextBindingObjectSetToDefaults: function(W.DurationTextBindingObject): T.Operation
   DurationTextBindingObjectSetUpdateInterval: function(W.DurationTextBindingObject, number): T.Operation
   DurationTextBindingObjectSetZeroDurationText: function(W.DurationTextBindingObject, ? string): T.Operation
   DurationTextBindingObjectUpdateFontString: function(W.DurationTextBindingObject): T.Operation
   FontStringGetFontObject: function(W.FontString): T.Operation
   FontStringGetText: function(W.FontString): T.Operation
   FontStringSetFontObject: function(W.FontString, string | W.Font): T.Operation
   FontStringSetText: function(W.FontString, ? string): T.Operation
   LuaDurationManualClockAdvanceTime: function(W.LuaDurationManualClock, number): T.Operation
   LuaDurationManualClockGetTime: function(W.LuaDurationManualClock): T.Operation
   LuaDurationManualClockResetTime: function(W.LuaDurationManualClock): T.Operation
   LuaDurationManualClockRewindTime: function(W.LuaDurationManualClock, number): T.Operation
   LuaDurationManualClockSetTime: function(W.LuaDurationManualClock, number): T.Operation
   LuaLocaleContextCompareStrings: function(W.LuaLocaleContext, string, string, number): T.Operation
   LuaLocaleContextFindBreaks: function(W.LuaLocaleContext, string, number): T.Operation
   LuaLocaleContextFindStringMatches: function(W.LuaLocaleContext, string, string, number): T.Operation
   LuaLocaleContextFoldCase: function(W.LuaLocaleContext, string): T.Operation
   LuaLocaleContextFormatCurrency: function(W.LuaLocaleContext, number, string): T.Operation
   LuaLocaleContextFormatDate: function(W.LuaLocaleContext, number, number, string): T.Operation
   LuaLocaleContextFormatDateTime: function(W.LuaLocaleContext, number, number, number, string): T.Operation
   LuaLocaleContextFormatNumber: function(W.LuaLocaleContext, number, number): T.Operation
   LuaLocaleContextFormatTime: function(W.LuaLocaleContext, number, number, string): T.Operation
   LuaLocaleContextGetCurrencyName: function(W.LuaLocaleContext, string, number): T.Operation
   LuaLocaleContextGetDisplayName: function(W.LuaLocaleContext, string): T.Operation
   LuaLocaleContextGetLocale: function(W.LuaLocaleContext): T.Operation
   LuaLocaleContextGetSortKey: function(W.LuaLocaleContext, string, number): T.Operation
   LuaLocaleContextLength: function(W.LuaLocaleContext, string): T.Operation
   LuaLocaleContextParseCurrency: function(W.LuaLocaleContext, string): T.Operation
   LuaLocaleContextParseNumber: function(W.LuaLocaleContext, string, number): T.Operation
   LuaLocaleContextSelectPlural: function(W.LuaLocaleContext, number, number): T.Operation
   LuaLocaleContextSetLocale: function(W.LuaLocaleContext, string): T.Operation
   LuaLocaleContextToLower: function(W.LuaLocaleContext, string): T.Operation
   LuaLocaleContextToTitle: function(W.LuaLocaleContext, string): T.Operation
   LuaLocaleContextToUpper: function(W.LuaLocaleContext, string): T.Operation
   LuaLocaleContextTransformLocale: function(W.LuaLocaleContext, number): T.Operation
   RadialProgressGetFromPercent: function(W.RadialProgress): T.Operation
   RadialProgressGetToPercent: function(W.RadialProgress): T.Operation
   RadialProgressSetFromPercent: function(W.RadialProgress, number): T.Operation
   RadialProgressSetToPercent: function(W.RadialProgress, number): T.Operation
   TimedSignalMapCancelAllSignals: function(W.TimedSignalMap): T.Operation
   TimedSignalMapCancelSignal: function(W.TimedSignalMap, number): T.Operation
   TimedSignalMapGetNextSignal: function(W.TimedSignalMap): T.Operation
   TimedSignalMapGetSignalCount: function(W.TimedSignalMap): T.Operation
   TimedSignalMapGetSignalTime: function(W.TimedSignalMap, number): T.Operation
   TimedSignalMapHasSignal: function(W.TimedSignalMap, number): T.Operation
   TimedSignalMapSignalAfter: function(W.TimedSignalMap, number, number): T.Operation
   TimedSignalMapSignalAt: function(W.TimedSignalMap, number, number): T.Operation
   VectorGraphicsClearSVG: function(W.VectorGraphics): T.Operation
   VectorGraphicsGetSVGFileID: function(W.VectorGraphics): T.Operation
   VectorGraphicsHasSVG: function(W.VectorGraphics): T.Operation
   VectorGraphicsSetSVG: function(W.VectorGraphics, string): T.Operation
end
local operations: Operations = {}

-- Each closure keeps lookup, access proof and invocation in one protected call.
-- Caller-scoped native permissions cannot be established by a different helper.
-- The state records whether an exception arose during lookup or invocation.
function operations.ColorCurveObjectAddPoint(object: W.ColorCurveObject, arg1: number, arg2: T.PublicColor): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.AddPoint
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1, {r = arg2.r, g = arg2.g, b = arg2.b, a = arg2.a}))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.DurationObjectGetClock(object: W.DurationObject): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.GetClock
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return packClock(fn(object))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.DurationObjectGetClockTime(object: W.DurationObject): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.GetClockTime
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.DurationObjectGetRemainingDuration(object: W.DurationObject): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.GetRemainingDuration
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.DurationObjectSetClock(object: W.DurationObject, arg1?: W.LuaDurationClockObject): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.SetClock
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.DurationObjectSetTimeSpan(object: W.DurationObject, arg1: number, arg2: number): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.SetTimeSpan
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1, arg2))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.DurationTextBindingObjectAssign(object: W.DurationTextBindingObject, arg1: W.DurationTextBindingObject): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.Assign
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.DurationTextBindingObjectCanFormatText(object: W.DurationTextBindingObject): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.CanFormatText
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.DurationTextBindingObjectCanUpdateFontString(object: W.DurationTextBindingObject): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.CanUpdateFontString
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.DurationTextBindingObjectClearTextColorCurve(object: W.DurationTextBindingObject): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.ClearTextColorCurve
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.DurationTextBindingObjectCopy(object: W.DurationTextBindingObject): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.Copy
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return packBinding(fn(object))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.DurationTextBindingObjectDisable(object: W.DurationTextBindingObject): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.Disable
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.DurationTextBindingObjectEnable(object: W.DurationTextBindingObject): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.Enable
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.DurationTextBindingObjectGetDuration(object: W.DurationTextBindingObject): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.GetDuration
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return packDuration(fn(object))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.DurationTextBindingObjectGetExpiredText(object: W.DurationTextBindingObject): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.GetExpiredText
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.DurationTextBindingObjectGetFontString(object: W.DurationTextBindingObject): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.GetFontString
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return packFont(fn(object))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.DurationTextBindingObjectGetFormattedText(object: W.DurationTextBindingObject): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.GetFormattedText
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.DurationTextBindingObjectGetFormattedTextColor(object: W.DurationTextBindingObject): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.GetFormattedTextColor
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return packColor(fn(object))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.DurationTextBindingObjectGetTextColorCurve(object: W.DurationTextBindingObject): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.GetTextColorCurve
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return packCurve(fn(object))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.DurationTextBindingObjectGetTimeModifier(object: W.DurationTextBindingObject): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.GetTimeModifier
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.DurationTextBindingObjectGetUpdateInterval(object: W.DurationTextBindingObject): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.GetUpdateInterval
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.DurationTextBindingObjectGetZeroDurationText(object: W.DurationTextBindingObject): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.GetZeroDurationText
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.DurationTextBindingObjectHasSecretValues(object: W.DurationTextBindingObject): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.HasSecretValues
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.DurationTextBindingObjectIsEnabled(object: W.DurationTextBindingObject): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.IsEnabled
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.DurationTextBindingObjectSetDuration(object: W.DurationTextBindingObject, arg1: W.DurationObject): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.SetDuration
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.DurationTextBindingObjectSetEnabled(object: W.DurationTextBindingObject, arg1: boolean): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.SetEnabled
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.DurationTextBindingObjectSetExpiredText(object: W.DurationTextBindingObject, arg1?: string): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.SetExpiredText
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.DurationTextBindingObjectSetFontString(object: W.DurationTextBindingObject, arg1: W.FontString): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.SetFontString
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.DurationTextBindingObjectSetFormatter(object: W.DurationTextBindingObject, arg1: W.NumericFormatter): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.SetFormatter
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.DurationTextBindingObjectSetTextColorCurve(object: W.DurationTextBindingObject, arg1: W.ColorCurveObject, arg2: number): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.SetTextColorCurve
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1, arg2))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.DurationTextBindingObjectSetTextFormat(object: W.DurationTextBindingObject, arg1: string, arg2: {W.DurationTextBindingFormatComponent}): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.SetTextFormat
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1, arg2))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.DurationTextBindingObjectSetTimeModifier(object: W.DurationTextBindingObject, arg1: number): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.SetTimeModifier
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.DurationTextBindingObjectSetToDefaults(object: W.DurationTextBindingObject): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.SetToDefaults
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.DurationTextBindingObjectSetUpdateInterval(object: W.DurationTextBindingObject, arg1: number): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.SetUpdateInterval
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.DurationTextBindingObjectSetZeroDurationText(object: W.DurationTextBindingObject, arg1?: string): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.SetZeroDurationText
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.DurationTextBindingObjectUpdateFontString(object: W.DurationTextBindingObject): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.UpdateFontString
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.FontStringGetFontObject(object: W.FontString): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.GetFontObject
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return packFontObject(fn(object))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.FontStringGetText(object: W.FontString): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.GetText
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.FontStringSetFontObject(object: W.FontString, arg1: string | W.Font): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.SetFontObject
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.FontStringSetText(object: W.FontString, arg1?: string): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.SetText
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.LuaDurationManualClockAdvanceTime(object: W.LuaDurationManualClock, arg1: number): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.AdvanceTime
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.LuaDurationManualClockGetTime(object: W.LuaDurationManualClock): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.GetTime
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.LuaDurationManualClockResetTime(object: W.LuaDurationManualClock): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.ResetTime
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.LuaDurationManualClockRewindTime(object: W.LuaDurationManualClock, arg1: number): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.RewindTime
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.LuaDurationManualClockSetTime(object: W.LuaDurationManualClock, arg1: number): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.SetTime
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.LuaLocaleContextCompareStrings(object: W.LuaLocaleContext, arg1: string, arg2: string, arg3: number): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.CompareStrings
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1, arg2, arg3))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.LuaLocaleContextFindBreaks(object: W.LuaLocaleContext, arg1: string, arg2: number): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.FindBreaks
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return packNumbers(fn(object, arg1, arg2))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.LuaLocaleContextFindStringMatches(object: W.LuaLocaleContext, arg1: string, arg2: string, arg3: number): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.FindStringMatches
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return packNumbers(fn(object, arg1, arg2, arg3))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.LuaLocaleContextFoldCase(object: W.LuaLocaleContext, arg1: string): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.FoldCase
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.LuaLocaleContextFormatCurrency(object: W.LuaLocaleContext, arg1: number, arg2: string): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.FormatCurrency
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1, arg2))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.LuaLocaleContextFormatDate(object: W.LuaLocaleContext, arg1: number, arg2: number, arg3: string): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.FormatDate
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1, arg2, arg3))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.LuaLocaleContextFormatDateTime(object: W.LuaLocaleContext, arg1: number, arg2: number, arg3: number, arg4: string): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.FormatDateTime
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1, arg2, arg3, arg4))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.LuaLocaleContextFormatNumber(object: W.LuaLocaleContext, arg1: number, arg2: number): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.FormatNumber
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1, arg2))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.LuaLocaleContextFormatTime(object: W.LuaLocaleContext, arg1: number, arg2: number, arg3: string): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.FormatTime
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1, arg2, arg3))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.LuaLocaleContextGetCurrencyName(object: W.LuaLocaleContext, arg1: string, arg2: number): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.GetCurrencyName
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1, arg2))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.LuaLocaleContextGetDisplayName(object: W.LuaLocaleContext, arg1: string): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.GetDisplayName
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.LuaLocaleContextGetLocale(object: W.LuaLocaleContext): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.GetLocale
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.LuaLocaleContextGetSortKey(object: W.LuaLocaleContext, arg1: string, arg2: number): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.GetSortKey
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1, arg2))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.LuaLocaleContextLength(object: W.LuaLocaleContext, arg1: string): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.Length
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.LuaLocaleContextParseCurrency(object: W.LuaLocaleContext, arg1: string): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.ParseCurrency
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return packCurrency(fn(object, arg1))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.LuaLocaleContextParseNumber(object: W.LuaLocaleContext, arg1: string, arg2: number): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.ParseNumber
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1, arg2))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.LuaLocaleContextSelectPlural(object: W.LuaLocaleContext, arg1: number, arg2: number): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.SelectPlural
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1, arg2))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.LuaLocaleContextSetLocale(object: W.LuaLocaleContext, arg1: string): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.SetLocale
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.LuaLocaleContextToLower(object: W.LuaLocaleContext, arg1: string): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.ToLower
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.LuaLocaleContextToTitle(object: W.LuaLocaleContext, arg1: string): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.ToTitle
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.LuaLocaleContextToUpper(object: W.LuaLocaleContext, arg1: string): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.ToUpper
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.LuaLocaleContextTransformLocale(object: W.LuaLocaleContext, arg1: number): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.TransformLocale
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.RadialProgressGetFromPercent(object: W.RadialProgress): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.GetFromPercent
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.RadialProgressGetToPercent(object: W.RadialProgress): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.GetToPercent
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.RadialProgressSetFromPercent(object: W.RadialProgress, arg1: number): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.SetFromPercent
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.RadialProgressSetToPercent(object: W.RadialProgress, arg1: number): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.SetToPercent
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.TimedSignalMapCancelAllSignals(object: W.TimedSignalMap): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.CancelAllSignals
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.TimedSignalMapCancelSignal(object: W.TimedSignalMap, arg1: number): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.CancelSignal
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.TimedSignalMapGetNextSignal(object: W.TimedSignalMap): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.GetNextSignal
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.TimedSignalMapGetSignalCount(object: W.TimedSignalMap): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.GetSignalCount
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.TimedSignalMapGetSignalTime(object: W.TimedSignalMap, arg1: number): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.GetSignalTime
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.TimedSignalMapHasSignal(object: W.TimedSignalMap, arg1: number): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.HasSignal
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.TimedSignalMapSignalAfter(object: W.TimedSignalMap, arg1: number, arg2: number): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.SignalAfter
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1, arg2))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.TimedSignalMapSignalAt(object: W.TimedSignalMap, arg1: number, arg2: number): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.SignalAt
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1, arg2))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.VectorGraphicsClearSVG(object: W.VectorGraphics): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.ClearSVG
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.VectorGraphicsGetSVGFileID(object: W.VectorGraphics): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.GetSVGFileID
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.VectorGraphicsHasSVG(object: W.VectorGraphics): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.HasSVG
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

function operations.VectorGraphicsSetSVG(object: W.VectorGraphics, arg1: string): T.Operation
   return function(): T.Packed, string
      local state = "lookup_error"
      local ok, packed = pcall(function(): T.Packed
         if issecretvalue(object) or not canaccessvalue(object) then state = "inaccessible"; return nil end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
         if kind == "table" and not canaccesstable(object) then state = "inaccessible_table"; return nil end
         local fn = object.SetSVG
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
         if fn == nil then state = "absent"; return nil end
         if type(fn) ~= "function" then state = "not_function"; return nil end
         state = "call_error"
         return pack(fn(object, arg1))
      end)
      if not ok then return nil, state end
      if state == "call_error" then return packed, "called" end
      return nil, state
   end
end

-- Each case owns a concrete public fixture and a bounded oracle. Metadata
-- supplies the coverage inventory, never the expected behavioral result.
private.RunInvocations = function(host: T.InvocationHost, result: T.Report)
   local bindings: {W.DurationTextBindingObject} = {}
   local signalMap: W.TimedSignalMap
   local pendingProjection: T.Invocation
   local rows: {T.Invocation} = {}
   local seen: {string: boolean} = {}
   local diagnostics: {string: T.BindingDiagnostic} = {}
   for _, id in ipairs(private.Plan.bindingDiagnosticCases) do
      local row: T.BindingDiagnostic = {caseID = id, state = "not_run"}
      result.bindingDiagnostics[#result.bindingDiagnostics + 1] = row
      diagnostics[id] = row
   end
   for _, id in ipairs(private.Plan.bindingExperimentCases) do
      local row: T.BindingDiagnostic = {caseID = id, state = "not_run"}
      result.bindingExperimentChecks[#result.bindingExperimentChecks + 1] = row
      diagnostics[id] = row
   end
   local samples: {string: T.BindingTextSample} = {}
   for _, id in ipairs(private.Plan.bindingTextSampleCases) do
      local row: T.BindingTextSample = {caseID = id, state = "not_run"}
      result.bindingTextSamples[#result.bindingTextSamples + 1] = row
      samples[id] = row
   end
   local function call(operation: T.Operation): T.Packed, string
      local ok, packed, state = pcall(operation)
      if not ok then return nil, "call_error" end
      return packed, state
   end
   local function value(operation: T.Operation): unknown
      local packed = call(operation)
      if packed then return packed.values[1] end
      return nil
   end
   local function durationValue(operation: T.Operation): W.DurationObject
      local packed = call(operation)
      return packed and packed.duration
   end
   local function fontValue(operation: T.Operation): W.FontString
      local packed = call(operation)
      return packed and packed.font
   end
   local function apply(operation: T.Operation): boolean
      local _, state = call(operation)
      return state == "called"
   end
   local function diagnosticSet(id: string, operation: T.Operation)
      local packed, state = call(operation)
      diagnostics[id].state = state == "called" and (packed.count == 0 and "called_void" or "unexpected_tuple") or state
   end
   local function readText(operation: T.Operation): string, string
      local packed, state = call(operation)
      if state ~= "called" then return nil, state end
      if packed.count ~= 1 then return nil, "unexpected_tuple" end
      local item = packed.values[1]
      if issecretvalue(item) then return nil, "secret" end
      if not canaccessvalue(item) then return nil, "inaccessible" end
      if item == nil then return nil, "nil" end
      if not (item is string) then return nil, "wrong_type" end
      if #item > 4096 then return nil, "oversized" end
      return item, #item == 0 and "public_empty" or "public_nonempty"
   end
   local function diagnosticText(id: string, operation: T.Operation): string
      local item, state = readText(operation)
      diagnostics[id].state = state
      return item
   end
   local function sampleText(id: string, item: unknown, state: string)
      local row = samples[id]
      row.state = state
      if state ~= "public_empty" and state ~= "public_nonempty" then return end
      -- This bounded exception samples only this probe's owned public text.
      -- Recheck here before byte operations and persistence; no secret sample
      -- or truncated prefix may become evidence through an earlier guard.
      if issecretvalue(item) then row.state = "secret"; return end
      if not canaccessvalue(item) then row.state = "inaccessible"; return end
      if not (item is string) then row.state = "wrong_type"; return end
      if #item > 128 then row.state = "oversized"; return end
      local hex = ""
      local firstNul: integer = 0
      for index = 1, #item do
         local byte = string.byte(item, index)
         hex = hex .. string.format("%02x", byte)
         if byte == 0 and firstNul == 0 then firstNul = index end
      end
      row.state, row.hex, row.byteLength, row.firstNul = "public", hex, #item, firstNul
   end
   local function sampleRead(id: string, operation: T.Operation): string
      local item, state = readText(operation)
      sampleText(id, item, state)
      return item
   end
   local function diagnosticCompare(id: string, left: unknown, right: unknown)
      -- Permission and carrier checks stay with the equality operation, even
      -- when another diagnostic already classified one of these values.
      if issecretvalue(left) or issecretvalue(right) or not canaccessallvalues(left, right) then
         diagnostics[id].state = "not_comparable"
      elseif not (left is string) or not (right is string) then diagnostics[id].state = "not_comparable"
      else diagnostics[id].state = left == right and "matched" or "mismatch" end
   end
   local function sampleCompare(id: string, leftID: string, rightID: string, left: unknown, right: unknown)
      if samples[leftID].state ~= "public" or samples[rightID].state ~= "public" then
         diagnostics[id].state = "not_comparable"
      else diagnosticCompare(id, left, right) end
   end
   local function diagnosticFact(id: string, operation: T.Operation, expected: unknown, kind: string)
      local packed, state = call(operation)
      if state ~= "called" or packed.count ~= 1 then diagnostics[id].state = "not_comparable"; return end
      local item = packed.values[1]
      if issecretvalue(item) or issecretvalue(expected) or not canaccessallvalues(item, expected) then diagnostics[id].state = "not_comparable"; return end
      if kind == "number" then
         if not (item is number) or item ~= item or item == math.huge or item == -math.huge then diagnostics[id].state = "not_comparable"; return end
      else
         local itemKind, expectedKind = type(item), type(expected)
         if (itemKind ~= "table" and itemKind ~= "userdata") or (expectedKind ~= "table" and expectedKind ~= "userdata") then diagnostics[id].state = "not_comparable"; return end
      end
      diagnostics[id].state = item == expected and "matched" or "mismatch"
   end
   local function run(canonical: string, name: string, operation: T.Operation, check: string, oracle?: T.Oracle): T.Invocation
      local id = canonical .. "." .. name
      if seen[id] then error("duplicate authored probe case") end
      seen[id] = true
      local row: T.Invocation = {caseID = id, state = "not_run", invoked = false, returnCount = 0, returns = {}, check = check, behaviorVerified = false}
      rows[#rows + 1] = row
      local packed, state = call(operation)
      row.state = state
      if state ~= "called" then return row end
      row.invoked, row.returnCount = true, packed.count
      if packed.count > 8 then row.state = "unsupported_tuple"; return row end
      for index = 1, packed.count do
         local item = packed.values[index]
         if issecretvalue(item) then row.returns[index] = "secret"; row.state = "secret_result"
         elseif not canaccessvalue(item) then row.returns[index] = "inaccessible"; row.state = "inaccessible_result"
         else row.returns[index] = type(item) end
      end
      if row.state ~= "called" then return row end
      if not oracle then
         row.state = packed.count == 0 and "pending_projection" or "property_mismatch"
         return row
      end
      local converted, checked = pcall(checkedResult, packed)
      if not converted then row.state = "property_mismatch"; return row end
      local ok, matched = pcall(oracle, checked)
      row.behaviorVerified = ok and matched == true
      row.state = row.behaviorVerified and "passed" or "property_mismatch"
      return row
   end
   local function finishProjection(baseline: unknown, projected: unknown)
      if not pendingProjection or pendingProjection.state ~= "pending_projection" then return end
      local ok, state = pcall(function(): string
         if diagnostics["sentinel.set"].state ~= "called_void" or diagnostics["sentinel.matches"].state ~= "matched" or
            diagnostics["baseline.expected_stable"].state ~= "matched" or diagnostics["projection.set"].state ~= "called_void" or
            samples["baseline.expected"].state ~= "public" or samples["baseline.font"].state ~= "public" or samples["projection.font"].state ~= "public" then
            return "projection_unavailable"
         end
         -- Compare the binding's saved output with the same font's native
         -- projection. Native grammar expansion stays outside this probe.
         if issecretvalue(baseline) or issecretvalue(projected) or not canaccessallvalues(baseline, projected) then return "projection_unavailable" end
         if not (baseline is string) or not (projected is string) or #baseline > 128 or #projected > 128 then return "projection_unavailable" end
         if projected == "PROBE_SENTINEL" then return "projection_unavailable" end
         return baseline == projected and "passed" or "property_mismatch"
      end)
      pendingProjection.state = ok and state or "projection_unavailable"
      pendingProjection.behaviorVerified = pendingProjection.state == "passed"
   end
   local function stringResult(packed: T.CheckedResult): boolean
      local decoded = text(packed.values[1])
      return packed.count == 1 and decoded ~= nil and #decoded > 0
   end
   local function exactNumber(expected: number): T.Oracle
      return function(packed: T.CheckedResult): boolean return packed.count == 1 and numeric(packed.values[1]) == expected end
   end
   local function exactText(expected: string): T.Oracle
      return function(packed: T.CheckedResult): boolean return packed.count == 1 and text(packed.values[1]) == expected end
   end
   local function exactBoolean(expected: boolean): T.Oracle
      return function(packed: T.CheckedResult): boolean return packed.count == 1 and boolean(packed.values[1]) == expected end
   end
   local ok = pcall(function()
      local binding = host.binding()
      if not issecretvalue(binding) and canaccessvalue(binding) and binding ~= nil then bindings[#bindings + 1] = binding end
      local clock = host.clock()
      local duration = host.duration()
      local formatter = host.formatter()
      local curve = host.curve()
      local font = host.font()
      local fontObject = host.fontObject()
      diagnosticSet("setup.font_object_set", operations.FontStringSetFontObject(font, fontObject))
      diagnosticSet("setup.clock_set_time", operations.LuaDurationManualClockSetTime(clock, 12.5))
      diagnosticSet("setup.duration_set_clock", operations.DurationObjectSetClock(duration, clock))
      diagnosticSet("setup.duration_set_span", operations.DurationObjectSetTimeSpan(duration, 0, 60))
      diagnosticFact("setup.font_object_identity", operations.FontStringGetFontObject(font), fontObject, "object")
      diagnosticFact("setup.duration_clock_identity", operations.DurationObjectGetClock(duration), clock, "object")
      diagnosticFact("setup.duration_clock_time", operations.DurationObjectGetClockTime(duration), 12.5, "number")
      diagnosticFact("setup.duration_remaining", operations.DurationObjectGetRemainingDuration(duration), 47.5, "number")
      -- ColorMixin-compatible public input avoids sampling native units.
      apply(operations.ColorCurveObjectAddPoint(curve, 0, {r = 1, g = 1, b = 1, a = 1}))
      apply(operations.ColorCurveObjectAddPoint(curve, 1, {r = 1, g = 1, b = 1, a = 1}))
      local canonical = "DurationTextBindingObject"
      run(canonical, "SetToDefaults", operations.DurationTextBindingObjectSetToDefaults(binding), "clears_duration", function(packed: T.CheckedResult): boolean
         local count = packed.count
         return count == 0 and same(durationValue(operations.DurationTextBindingObjectGetDuration(binding)), nil)
      end)
      run(canonical, "SetDuration", operations.DurationTextBindingObjectSetDuration(binding, duration), "duration_identity", function(packed: T.CheckedResult): boolean
         local count = packed.count return count == 0 and same(durationValue(operations.DurationTextBindingObjectGetDuration(binding)), duration) end)
      run(canonical, "GetDuration", operations.DurationTextBindingObjectGetDuration(binding), "duration_identity", function(packed: T.CheckedResult): boolean
         local count = packed.count return count == 1 and same(packed.duration, duration) end)
      run(canonical, "SetFontString", operations.DurationTextBindingObjectSetFontString(binding, font), "font_identity", function(packed: T.CheckedResult): boolean
         local count = packed.count return count == 0 and same(fontValue(operations.DurationTextBindingObjectGetFontString(binding)), font) end)
      run(canonical, "GetFontString", operations.DurationTextBindingObjectGetFontString(binding), "font_identity", function(packed: T.CheckedResult): boolean
         local count = packed.count return count == 1 and same(packed.font, font) end)
      run(canonical, "SetFormatter", operations.DurationTextBindingObjectSetFormatter(binding, formatter), "can_format", function(packed: T.CheckedResult): boolean
         local count = packed.count return count == 0 and boolean(value(operations.DurationTextBindingObjectCanFormatText(binding))) == true end)
      run(canonical, "SetTextFormat", operations.DurationTextBindingObjectSetTextFormat(binding, "{}", {{property = 0, formatter = formatter}}), "can_format_components", function(packed: T.CheckedResult): boolean
         local count = packed.count return count == 0 and boolean(value(operations.DurationTextBindingObjectCanFormatText(binding))) == true end)
      run(canonical, "CanFormatText", operations.DurationTextBindingObjectCanFormatText(binding), "configured_true", exactBoolean(true))
      run(canonical, "CanUpdateFontString", operations.DurationTextBindingObjectCanUpdateFontString(binding), "configured_true", exactBoolean(true))
      run(canonical, "SetExpiredText", operations.DurationTextBindingObjectSetExpiredText(binding, "EXPIRED"), "text_roundtrip", function(packed: T.CheckedResult): boolean
         local count = packed.count return count == 0 and text(value(operations.DurationTextBindingObjectGetExpiredText(binding))) == "EXPIRED" end)
      run(canonical, "GetExpiredText", operations.DurationTextBindingObjectGetExpiredText(binding), "text_roundtrip", exactText("EXPIRED"))
      run(canonical, "SetZeroDurationText", operations.DurationTextBindingObjectSetZeroDurationText(binding, "ZERO"), "text_roundtrip", function(packed: T.CheckedResult): boolean
         local count = packed.count return count == 0 and text(value(operations.DurationTextBindingObjectGetZeroDurationText(binding))) == "ZERO" end)
      run(canonical, "GetZeroDurationText", operations.DurationTextBindingObjectGetZeroDurationText(binding), "text_roundtrip", exactText("ZERO"))
      run(canonical, "SetTimeModifier", operations.DurationTextBindingObjectSetTimeModifier(binding, 1), "modifier_roundtrip", function(packed: T.CheckedResult): boolean
         local count = packed.count return count == 0 and numeric(value(operations.DurationTextBindingObjectGetTimeModifier(binding))) == 1 end)
      run(canonical, "GetTimeModifier", operations.DurationTextBindingObjectGetTimeModifier(binding), "modifier_roundtrip", exactNumber(1))
      run(canonical, "SetUpdateInterval", operations.DurationTextBindingObjectSetUpdateInterval(binding, 0.25), "interval_roundtrip", function(packed: T.CheckedResult): boolean
         local count = packed.count return count == 0 and numeric(value(operations.DurationTextBindingObjectGetUpdateInterval(binding))) == 0.25 end)
      run(canonical, "GetUpdateInterval", operations.DurationTextBindingObjectGetUpdateInterval(binding), "interval_roundtrip", exactNumber(0.25))
      run(canonical, "SetTextColorCurve", operations.DurationTextBindingObjectSetTextColorCurve(binding, curve, 0), "curve_property_tuple", function(packed: T.CheckedResult): boolean
         local count = packed.count
         local returned = call(operations.DurationTextBindingObjectGetTextColorCurve(binding))
         return count == 0 and returned ~= nil and returned.count == 2 and same(returned.curve, curve) and numeric(returned.values[2]) == 0
      end)
      run(canonical, "GetTextColorCurve", operations.DurationTextBindingObjectGetTextColorCurve(binding), "curve_property_tuple", function(packed: T.CheckedResult): boolean
         local values, count = packed.values, packed.count return count == 2 and same(packed.curve, curve) and numeric(values[2]) == 0 end)
      run(canonical, "GetFormattedTextColor", operations.DurationTextBindingObjectGetFormattedTextColor(binding), "public_white_rgba", function(packed: T.CheckedResult): boolean
         local count = packed.count
         local color = packed.color
         if issecretvalue(color) or not canaccessvalue(color) or type(color) ~= "table" or not canaccesstable(color) then return false end
         local r, g, b, a = color.r, color.g, color.b, color.a
         return count == 1 and numeric(r) == 1 and numeric(g) == 1 and numeric(b) == 1 and numeric(a) == 1
      end)
      run(canonical, "ClearTextColorCurve", operations.DurationTextBindingObjectClearTextColorCurve(binding), "cleared_curve_tuple", function(packed: T.CheckedResult): boolean
         local count = packed.count
         local returned = call(operations.DurationTextBindingObjectGetTextColorCurve(binding))
         return count == 0 and returned ~= nil and returned.count == 0
      end)
      run(canonical, "GetFormattedText", operations.DurationTextBindingObjectGetFormattedText(binding), "public_nonempty_text", stringResult)
      run(canonical, "Enable", operations.DurationTextBindingObjectEnable(binding), "enabled_true", function(packed: T.CheckedResult): boolean
         local count = packed.count return count == 0 and boolean(value(operations.DurationTextBindingObjectIsEnabled(binding))) == true end)
      run(canonical, "IsEnabled", operations.DurationTextBindingObjectIsEnabled(binding), "enabled_true", exactBoolean(true))
      local expectedBefore = diagnosticText("baseline.expected_before", operations.DurationTextBindingObjectGetFormattedText(binding))
      pendingProjection = run(canonical, "UpdateFontString", operations.DurationTextBindingObjectUpdateFontString(binding), "font_matches_native_projection")
      local expectedAfter = diagnosticText("baseline.expected_after", operations.DurationTextBindingObjectGetFormattedText(binding))
      sampleText("baseline.expected", expectedAfter, diagnostics["baseline.expected_after"].state)
      local actual = diagnosticText("baseline.font_after", operations.FontStringGetText(font))
      sampleText("baseline.font", actual, diagnostics["baseline.font_after"].state)
      local baselineExpected = expectedAfter
      local baselineFont = actual
      diagnosticCompare("baseline.expected_stable", expectedBefore, expectedAfter)
      diagnosticCompare("baseline.font_matches", actual, expectedAfter)
      diagnosticSet("sentinel.set", operations.FontStringSetText(font, "PROBE_SENTINEL"))
      actual = diagnosticText("sentinel.read", operations.FontStringGetText(font))
      diagnosticCompare("sentinel.matches", actual, "PROBE_SENTINEL")
      -- Put the projection directly after the sentinel so a no-op control
      -- cannot accidentally validate a binding's unchanged output against itself.
      if diagnostics["sentinel.set"].state == "called_void" and diagnostics["sentinel.matches"].state == "matched" and
         diagnostics["baseline.expected_stable"].state == "matched" and samples["baseline.expected"].state == "public" and samples["baseline.font"].state == "public" and
         not issecretvalue(baselineExpected) and canaccessvalue(baselineExpected) and baselineExpected is string and #baselineExpected <= 128 then
         diagnosticSet("projection.set", operations.FontStringSetText(font, baselineExpected))
      else diagnostics["projection.set"].state = "not_run" end
      actual = sampleRead("projection.font", operations.FontStringGetText(font))
      sampleCompare("projection.matches", "projection.font", "baseline.expected", actual, baselineExpected)
      finishProjection(baselineFont, actual)
      diagnosticSet("retry.update", operations.DurationTextBindingObjectUpdateFontString(binding))
      expectedAfter = diagnosticText("retry.expected", operations.DurationTextBindingObjectGetFormattedText(binding))
      actual = diagnosticText("retry.font", operations.FontStringGetText(font))
      diagnosticCompare("retry.matches", actual, expectedAfter)
      diagnosticSet("zero_interval.set", operations.DurationTextBindingObjectSetUpdateInterval(binding, 0))
      diagnosticSet("zero_interval.update", operations.DurationTextBindingObjectUpdateFontString(binding))
      expectedAfter = diagnosticText("zero_interval.expected", operations.DurationTextBindingObjectGetFormattedText(binding))
      actual = diagnosticText("zero_interval.font", operations.FontStringGetText(font))
      diagnosticCompare("zero_interval.matches", actual, expectedAfter)
      diagnosticSet("manual_step.set_time", operations.LuaDurationManualClockSetTime(clock, 20))
      diagnosticFact("manual_step.clock_time", operations.DurationObjectGetClockTime(duration), 20, "number")
      diagnosticFact("manual_step.remaining", operations.DurationObjectGetRemainingDuration(duration), 40, "number")
      diagnosticSet("manual_step.update", operations.DurationTextBindingObjectUpdateFontString(binding))
      expectedAfter = sampleRead("manual_step.expected", operations.DurationTextBindingObjectGetFormattedText(binding))
      actual = sampleRead("manual_step.font", operations.FontStringGetText(font))
      sampleCompare("manual_step.matches", "manual_step.font", "manual_step.expected", actual, expectedAfter)
      diagnosticSet("manual_step.restore_time", operations.LuaDurationManualClockSetTime(clock, 12.5))
      local nowFn = GetTime
      local now: number
      if not issecretvalue(nowFn) and canaccessvalue(nowFn) and type(nowFn) == "function" then
         local okNow, rawNow = pcall(nowFn)
         if okNow and not issecretvalue(rawNow) and canaccessvalue(rawNow) and rawNow is number then
            if rawNow == rawNow and rawNow ~= math.huge and rawNow ~= -math.huge then now = rawNow end
         end
      end
      if now == nil then diagnostics["default_clock.time"].state = "not_comparable"
      else
         diagnostics["default_clock.time"].state = "matched"
         -- Default clock uses GetTime; this span is current even if native font
         -- updates do not honor the manual clock used by the baseline fixture.
         diagnosticSet("default_clock.clear", operations.DurationObjectSetClock(duration, nil))
         diagnosticSet("default_clock.span", operations.DurationObjectSetTimeSpan(duration, now, now + 60))
         diagnosticFact("default_clock.clock_time", operations.DurationObjectGetClockTime(duration), now, "number")
         diagnosticFact("default_clock.remaining", operations.DurationObjectGetRemainingDuration(duration), 60, "number")
         diagnosticSet("default_clock.update", operations.DurationTextBindingObjectUpdateFontString(binding))
         expectedAfter = sampleRead("default_clock.expected", operations.DurationTextBindingObjectGetFormattedText(binding))
         actual = sampleRead("default_clock.font", operations.FontStringGetText(font))
         sampleCompare("default_clock.matches", "default_clock.font", "default_clock.expected", actual, expectedAfter)
      end
      run(canonical, "HasSecretValues", operations.DurationTextBindingObjectHasSecretValues(binding), "public_fixture_false", exactBoolean(false))
      run(canonical, "SetEnabled", operations.DurationTextBindingObjectSetEnabled(binding, false), "enabled_false", function(packed: T.CheckedResult): boolean
         local count = packed.count return count == 0 and boolean(value(operations.DurationTextBindingObjectIsEnabled(binding))) == false end)
      run(canonical, "Disable", operations.DurationTextBindingObjectDisable(binding), "enabled_false", function(packed: T.CheckedResult): boolean
         local count = packed.count return count == 0 and boolean(value(operations.DurationTextBindingObjectIsEnabled(binding))) == false end)
      local copied: W.DurationTextBindingObject
      run(canonical, "Copy", operations.DurationTextBindingObjectCopy(binding), "copy_preserves_text", function(packed: T.CheckedResult): boolean
         local count = packed.count
         copied = packed.binding
         if issecretvalue(copied) or not canaccessvalue(copied) or copied == nil then return false end
         bindings[#bindings + 1] = copied
         return count == 1 and text(value(operations.DurationTextBindingObjectGetExpiredText(copied))) == "EXPIRED"
      end)
      apply(operations.DurationTextBindingObjectSetExpiredText(copied, "ASSIGNED"))
      run(canonical, "Assign", operations.DurationTextBindingObjectAssign(binding, copied), "assignment_preserves_text", function(packed: T.CheckedResult): boolean
         local count = packed.count return count == 0 and text(value(operations.DurationTextBindingObjectGetExpiredText(binding))) == "ASSIGNED" end)

      run("LuaDurationClockObject", "GetTime", operations.LuaDurationManualClockGetTime(clock), "manual_time_12_5", exactNumber(12.5))
      canonical = "LuaDurationManualClock"
      run(canonical, "SetTime", operations.LuaDurationManualClockSetTime(clock, 12.5), "manual_time_12_5", function(packed: T.CheckedResult): boolean
         local count = packed.count return count == 0 and numeric(value(operations.LuaDurationManualClockGetTime(clock))) == 12.5 end)
      run(canonical, "AdvanceTime", operations.LuaDurationManualClockAdvanceTime(clock, 3), "manual_time_15_5", function(packed: T.CheckedResult): boolean
         local count = packed.count return count == 0 and numeric(value(operations.LuaDurationManualClockGetTime(clock))) == 15.5 end)
      run(canonical, "RewindTime", operations.LuaDurationManualClockRewindTime(clock, 1), "manual_time_14_5", function(packed: T.CheckedResult): boolean
         local count = packed.count return count == 0 and numeric(value(operations.LuaDurationManualClockGetTime(clock))) == 14.5 end)
      run(canonical, "ResetTime", operations.LuaDurationManualClockResetTime(clock), "manual_time_zero", function(packed: T.CheckedResult): boolean
         local count = packed.count return count == 0 and numeric(value(operations.LuaDurationManualClockGetTime(clock))) == 0 end)

      local locale = host.locale()
      canonical = "LuaLocaleContext"
      run(canonical, "SetLocale", operations.LuaLocaleContextSetLocale(locale, "en_US"), "locale_accepted", exactBoolean(true))
      run(canonical, "GetLocale", operations.LuaLocaleContextGetLocale(locale), "en_us_locale", function(packed: T.CheckedResult): boolean
         local values, count = packed.values, packed.count local decoded = text(values[1]); return count == 1 and (decoded == "en_US" or decoded == "enUS") end)
      run(canonical, "CompareStrings", operations.LuaLocaleContextCompareStrings(locale, "abc", "abc", 4), "equal_collation_zero", exactNumber(0))
      run(canonical, "FindBreaks", operations.LuaLocaleContextFindBreaks(locale, "abc", 0), "four_character_boundaries", function(packed: T.CheckedResult): boolean
         local count = packed.count return count == 1 and arrayCount(packed.numbers) == 4 end)
      run(canonical, "FindStringMatches", operations.LuaLocaleContextFindStringMatches(locale, "abc abc", "abc", 4), "two_nonoverlapping_matches", function(packed: T.CheckedResult): boolean
         local count = packed.count return count == 1 and arrayCount(packed.numbers) == 2 end)
      run(canonical, "FoldCase", operations.LuaLocaleContextFoldCase(locale, "ABC"), "ascii_fold", exactText("abc"))
      run(canonical, "FormatCurrency", operations.LuaLocaleContextFormatCurrency(locale, 12.5, "USD"), "public_currency_text", stringResult)
      run(canonical, "FormatDate", operations.LuaLocaleContextFormatDate(locale, 0, 3, "UTC"), "public_date_text", stringResult)
      run(canonical, "FormatDateTime", operations.LuaLocaleContextFormatDateTime(locale, 0, 3, 1, "UTC"), "public_datetime_text", stringResult)
      run(canonical, "FormatNumber", operations.LuaLocaleContextFormatNumber(locale, 12.5, 0), "ascii_decimal", exactText("12.5"))
      run(canonical, "FormatTime", operations.LuaLocaleContextFormatTime(locale, 0, 1, "UTC"), "public_time_text", stringResult)
      run(canonical, "GetCurrencyName", operations.LuaLocaleContextGetCurrencyName(locale, "USD", 0), "public_currency_name", stringResult)
      run(canonical, "GetDisplayName", operations.LuaLocaleContextGetDisplayName(locale, "en_US"), "public_display_name", stringResult)
      run(canonical, "GetSortKey", operations.LuaLocaleContextGetSortKey(locale, "abc", 4), "public_nonempty_sort_key", stringResult)
      run(canonical, "Length", operations.LuaLocaleContextLength(locale, "abc"), "three_characters", exactNumber(3))
      run(canonical, "ParseCurrency", operations.LuaLocaleContextParseCurrency(locale, "$12.50"), "currency_amount_code", function(packed: T.CheckedResult): boolean
         local count = packed.count
         local currency = packed.currency
         if issecretvalue(currency) or not canaccessvalue(currency) or type(currency) ~= "table" or not canaccesstable(currency) then return false end
         local amount, code = currency.amount, currency.currencyCode
         return count == 1 and numeric(amount) == 12.5 and text(code) == "USD"
      end)
      run(canonical, "ParseNumber", operations.LuaLocaleContextParseNumber(locale, "12.5", 0), "decimal_parse", exactNumber(12.5))
      run(canonical, "SelectPlural", operations.LuaLocaleContextSelectPlural(locale, 1, 0), "singular_one", exactText("one"))
      run(canonical, "ToLower", operations.LuaLocaleContextToLower(locale, "ABC"), "ascii_lower", exactText("abc"))
      run(canonical, "ToTitle", operations.LuaLocaleContextToTitle(locale, "hello world"), "ascii_title", exactText("Hello World"))
      run(canonical, "ToUpper", operations.LuaLocaleContextToUpper(locale, "abc"), "ascii_upper", exactText("ABC"))
      run(canonical, "TransformLocale", operations.LuaLocaleContextTransformLocale(locale, 3), "language_en", exactText("en"))

      signalMap = host.signals()
      canonical = "TimedSignalMap"
      local signalTime = GetTime
      local deadline: number
      if not issecretvalue(signalTime) and canaccessvalue(signalTime) and type(signalTime) == "function" then
         local okNow, signalNow = pcall(signalTime)
         if okNow then deadline = numeric(signalNow) end
      end
      if deadline then deadline = deadline + 86400 end
      run(canonical, "SignalAt", operations.TimedSignalMapSignalAt(signalMap, 101, deadline), "scheduled_key_present", function(packed: T.CheckedResult): boolean
         local count = packed.count local returned = call(operations.TimedSignalMapHasSignal(signalMap, 101)); return count == 0 and returned ~= nil and boolean(returned.values[1]) == true end)
      run(canonical, "SignalAfter", operations.TimedSignalMapSignalAfter(signalMap, 202, 172800), "scheduled_key_present", function(packed: T.CheckedResult): boolean
         local count = packed.count local returned = call(operations.TimedSignalMapHasSignal(signalMap, 202)); return count == 0 and returned ~= nil and boolean(returned.values[1]) == true end)
      run(canonical, "HasSignal", operations.TimedSignalMapHasSignal(signalMap, 101), "scheduled_key_present", exactBoolean(true))
      run(canonical, "GetSignalCount", operations.TimedSignalMapGetSignalCount(signalMap), "two_scheduled_keys", exactNumber(2))
      run(canonical, "GetSignalTime", operations.TimedSignalMapGetSignalTime(signalMap, 101), "absolute_deadline", function(packed: T.CheckedResult): boolean
         local values, count = packed.values, packed.count return count == 1 and numeric(values[1]) == deadline end)
      run(canonical, "GetNextSignal", operations.TimedSignalMapGetNextSignal(signalMap), "key_time_tuple", function(packed: T.CheckedResult): boolean
         local values, count = packed.values, packed.count return count == 2 and numeric(values[1]) == 101 and numeric(values[2]) == deadline end)
      run(canonical, "CancelSignal", operations.TimedSignalMapCancelSignal(signalMap, 101), "one_key_remains", function(packed: T.CheckedResult): boolean
         local count = packed.count return count == 0 and numeric(value(operations.TimedSignalMapGetSignalCount(signalMap))) == 1 end)
      run(canonical, "CancelAllSignals", operations.TimedSignalMapCancelAllSignals(signalMap), "zero_keys_remain", function(packed: T.CheckedResult): boolean
         local count = packed.count return count == 0 and numeric(value(operations.TimedSignalMapGetSignalCount(signalMap))) == 0 end)

      local radial = host.radial()
      canonical = "RadialProgress"
      run(canonical, "SetFromPercent", operations.RadialProgressSetFromPercent(radial, 0.25), "from_roundtrip", function(packed: T.CheckedResult): boolean
         local count = packed.count return count == 0 and numeric(value(operations.RadialProgressGetFromPercent(radial))) == 0.25 end)
      run(canonical, "SetToPercent", operations.RadialProgressSetToPercent(radial, 0.75), "to_roundtrip", function(packed: T.CheckedResult): boolean
         local count = packed.count return count == 0 and numeric(value(operations.RadialProgressGetToPercent(radial))) == 0.75 end)
      run(canonical, "GetFromPercent", operations.RadialProgressGetFromPercent(radial), "from_roundtrip", exactNumber(0.25))
      run(canonical, "GetToPercent", operations.RadialProgressGetToPercent(radial), "to_roundtrip", exactNumber(0.75))
      for _, row in ipairs(host.observeHierarchy(canonical)) do result.hierarchy[#result.hierarchy + 1] = row end

      local vector = host.vector()
      canonical = "VectorGraphics"
      run(canonical, "SetSVG", operations.VectorGraphicsSetSVG(vector, "Interface\\AddOns\\WoWAPIProbe\\Fixture.svg"), "owned_svg_accepted", exactBoolean(true))
      run(canonical, "GetSVGFileID", operations.VectorGraphicsGetSVGFileID(vector), "public_fileid_carrier", function(packed: T.CheckedResult): boolean
         local values, count = packed.values, packed.count return count == 1 and numeric(values[1]) ~= nil end)
      run(canonical, "HasSVG", operations.VectorGraphicsHasSVG(vector), "owned_svg_present", exactBoolean(true))
      run(canonical, "ClearSVG", operations.VectorGraphicsClearSVG(vector), "svg_cleared", function(packed: T.CheckedResult): boolean
         local count = packed.count return count == 0 and boolean(value(operations.VectorGraphicsHasSVG(vector))) == false end)
      for _, row in ipairs(host.observeHierarchy(canonical)) do result.hierarchy[#result.hierarchy + 1] = row end
   end)
   -- Always release mutable native work, including after a setup/oracle error.
   for index, binding in ipairs(bindings) do
      local _, state = call(operations.DurationTextBindingObjectDisable(binding))
      result.cleanup[#result.cleanup + 1] = {caseID = "cleanup.invocation.binding." .. tostring(index), state = state == "called" and "complete" or state}
   end
   if not issecretvalue(signalMap) and canaccessvalue(signalMap) and signalMap ~= nil then
      local _, state = call(operations.TimedSignalMapCancelAllSignals(signalMap))
      result.cleanup[#result.cleanup + 1] = {caseID = "cleanup.invocation.signals", state = state == "called" and "complete" or state}
   end
   result.invocations = rows
   if pendingProjection and pendingProjection.state == "pending_projection" then pendingProjection.state = "projection_unavailable" end
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
            rows[#rows + 1] = {caseID = id, state = "fixture_setup_failed", invoked = false, returnCount = 0, returns = {}, check = "not_run", behaviorVerified = false}
         end
      end
   end
   for _, row in ipairs(result.cleanup) do if row.state ~= "complete" then complete = false end end
   result.invocationCaptureComplete = complete and #rows == private.Plan.expectedMethods
   result.outcome = result.invocationCaptureComplete and "public_invocations_verified" or "public_invocations_incomplete"
end
