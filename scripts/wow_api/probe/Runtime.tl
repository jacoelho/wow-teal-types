local type W = require("wow")
local type N = require("native_values")
local type T = require("probe_types")
local _, private = ... as (string, T.Namespace)
local plan = private.Plan
local last: T.Report
local capturing = false

local type ObjectTypeQuery = function(W.Object): N.TextLike
local type IsObjectTypeQuery = function(W.Object, string): boolean | W.OpaqueValue
local type DisableBinding = function(W.DurationTextBindingObject)
local type Observer = function(string): string
local type Queries = function(string): {T.Observation}
local record Resource
   state: string
   observe: Observer
   queries: Queries
   cleanup: function(): string
end

local function scalar(value: unknown): string
   if issecretvalue(value) or not canaccessvalue(value) then return "unavailable" end
   if value is string and #value <= 96 then return value end
   return "unavailable"
end

local function number(value: unknown): number
   if issecretvalue(value) or not canaccessvalue(value) or not (value is number) then return nil end
   if value ~= value or value == math.huge or value == -math.huge then return nil end
   return value
end

local function observeBinding(object: W.DurationTextBindingObject): Observer
   return function(name: string): string
      local ok, state = pcall(function(): string
         if issecretvalue(object) or not canaccessvalue(object) then return "inaccessible" end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then return "invalid_object" end
         if kind == "table" and not canaccesstable(object) then return "inaccessible_table" end
         local found = __wow_observe_field(object, name)
         if issecretvalue(found) or not canaccessvalue(found) then return "inaccessible" end
         if found is nil then return "absent" end
         return type(found) == "function" and "present" or "not_function"
      end)
      return ok and state or "lookup_error"
   end
end

local function observeClock(object: W.LuaDurationManualClock): Observer
   return function(name: string): string
      local ok, state = pcall(function(): string
         if issecretvalue(object) or not canaccessvalue(object) then return "inaccessible" end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then return "invalid_object" end
         if kind == "table" and not canaccesstable(object) then return "inaccessible_table" end
         local found = __wow_observe_field(object, name)
         if issecretvalue(found) or not canaccessvalue(found) then return "inaccessible" end
         if found is nil then return "absent" end
         return type(found) == "function" and "present" or "not_function"
      end)
      return ok and state or "lookup_error"
   end
end

local function observeLocale(object: W.LuaLocaleContext): Observer
   return function(name: string): string
      local ok, state = pcall(function(): string
         if issecretvalue(object) or not canaccessvalue(object) then return "inaccessible" end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then return "invalid_object" end
         if kind == "table" and not canaccesstable(object) then return "inaccessible_table" end
         local found = __wow_observe_field(object, name)
         if issecretvalue(found) or not canaccessvalue(found) then return "inaccessible" end
         if found is nil then return "absent" end
         return type(found) == "function" and "present" or "not_function"
      end)
      return ok and state or "lookup_error"
   end
end

local function observeSignals(object: W.TimedSignalMap): Observer
   return function(name: string): string
      local ok, state = pcall(function(): string
         if issecretvalue(object) or not canaccessvalue(object) then return "inaccessible" end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then return "invalid_object" end
         if kind == "table" and not canaccesstable(object) then return "inaccessible_table" end
         local found = __wow_observe_field(object, name)
         if issecretvalue(found) or not canaccessvalue(found) then return "inaccessible" end
         if found is nil then return "absent" end
         return type(found) == "function" and "present" or "not_function"
      end)
      return ok and state or "lookup_error"
   end
end

local function observeRadial(object: W.RadialProgress): Observer
   return function(name: string): string
      local ok, state = pcall(function(): string
         if issecretvalue(object) or not canaccessvalue(object) then return "inaccessible" end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then return "invalid_object" end
         if kind == "table" and not canaccesstable(object) then return "inaccessible_table" end
         local found = __wow_observe_field(object, name)
         if issecretvalue(found) or not canaccessvalue(found) then return "inaccessible" end
         if found is nil then return "absent" end
         return type(found) == "function" and "present" or "not_function"
      end)
      return ok and state or "lookup_error"
   end
end

local function observeVector(object: W.VectorGraphics): Observer
   return function(name: string): string
      local ok, state = pcall(function(): string
         if issecretvalue(object) or not canaccessvalue(object) then return "inaccessible" end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then return "invalid_object" end
         if kind == "table" and not canaccesstable(object) then return "inaccessible_table" end
         local found = __wow_observe_field(object, name)
         if issecretvalue(found) or not canaccessvalue(found) then return "inaccessible" end
         if found is nil then return "absent" end
         return type(found) == "function" and "present" or "not_function"
      end)
      return ok and state or "lookup_error"
   end
end

local function observeFrame(object: W.Frame): Observer
   return function(name: string): string
      local ok, state = pcall(function(): string
         if issecretvalue(object) or not canaccessvalue(object) then return "inaccessible" end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then return "invalid_object" end
         if kind == "table" and not canaccesstable(object) then return "inaccessible_table" end
         local found = __wow_observe_field(object, name)
         if issecretvalue(found) or not canaccessvalue(found) then return "inaccessible" end
         if found is nil then return "absent" end
         return type(found) == "function" and "present" or "not_function"
      end)
      return ok and state or "lookup_error"
   end
end

local function observeFont(object: W.FontString): Observer
   return function(name: string): string
      local ok, state = pcall(function(): string
         if issecretvalue(object) or not canaccessvalue(object) then return "inaccessible" end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then return "invalid_object" end
         if kind == "table" and not canaccesstable(object) then return "inaccessible_table" end
         local found = __wow_observe_field(object, name)
         if issecretvalue(found) or not canaccessvalue(found) then return "inaccessible" end
         if found is nil then return "absent" end
         return type(found) == "function" and "present" or "not_function"
      end)
      return ok and state or "lookup_error"
   end
end

local function observeTexture(object: W.Texture): Observer
   return function(name: string): string
      local ok, state = pcall(function(): string
         if issecretvalue(object) or not canaccessvalue(object) then return "inaccessible" end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then return "invalid_object" end
         if kind == "table" and not canaccesstable(object) then return "inaccessible_table" end
         local found = __wow_observe_field(object, name)
         if issecretvalue(found) or not canaccessvalue(found) then return "inaccessible" end
         if found is nil then return "absent" end
         return type(found) == "function" and "present" or "not_function"
      end)
      return ok and state or "lookup_error"
   end
end

local record UI
   frame: W.Frame
   group: W.AnimationGroup
   radial: W.RadialProgress
   vector: W.VectorGraphics
   font: W.FontString
   texture: W.Texture
   attempted: {string: boolean}
   states: {string: string}
end
local ui: UI = {attempted = {}, states = {}}

local function makeFrame(): W.Frame, string
   local state = "lookup_error"
   local ok, object = pcall(function(): W.Frame
      local fn = CreateFrame
      if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
      if fn == nil then state = "absent"; return nil end
      if type(fn) ~= "function" then state = "unavailable"; return nil end
      state = "call_error"
      return fn("Frame")
   end)
   if not ok or state ~= "call_error" then return nil, state end
   if issecretvalue(object) or not canaccessvalue(object) then return nil, "inaccessible" end
   if object == nil then return nil, "no_result" end
   local kind = type(object)
   if kind ~= "table" and kind ~= "userdata" then return nil, "invalid_object" end
   if kind == "table" and not canaccesstable(object) then return nil, "inaccessible_table" end
   return object, "available"
end

local function getFrame(): W.Frame, string
   if ui.attempted.frame then return ui.frame, ui.states.frame end
   ui.attempted.frame = true
   local object, state = makeFrame()
   if state == "available" then
      local ok, hidden = pcall(function(): boolean
         if issecretvalue(object) or not canaccessvalue(object) then return false end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then return false end
         if kind == "table" and not canaccesstable(object) then return false end
         local hide = object.Hide
         if issecretvalue(hide) or not canaccessvalue(hide) or type(hide) ~= "function" then return false end
         hide(object)
         return true
      end)
      if not ok then object, state = nil, "hide_error"
      elseif not hidden then object, state = nil, "hide_unavailable" end
   end
   ui.frame, ui.states.frame = object, state
   return object, state
end

local function makeGroup(): W.AnimationGroup, string
   local owner, ownerState = getFrame()
   if ownerState ~= "available" then return nil, ownerState end
   local state = "lookup_error"
   local ok, object = pcall(function(): W.AnimationGroup
      if issecretvalue(owner) or not canaccessvalue(owner) then state = "inaccessible"; return nil end
      local kind = type(owner)
      if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
      if kind == "table" and not canaccesstable(owner) then state = "inaccessible_table"; return nil end
      local fn = owner.CreateAnimationGroup
      if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
      if fn == nil then state = "absent"; return nil end
      if type(fn) ~= "function" then state = "unavailable"; return nil end
      state = "call_error"
      return fn(owner)
   end)
   if not ok or state ~= "call_error" then return nil, state end
   if issecretvalue(object) or not canaccessvalue(object) then return nil, "inaccessible" end
   if object == nil then return nil, "no_result" end
   local kind = type(object)
   if kind ~= "table" and kind ~= "userdata" then return nil, "invalid_object" end
   if kind == "table" and not canaccesstable(object) then return nil, "inaccessible_table" end
   return object, "available"
end

local function getGroup(): W.AnimationGroup, string
   if ui.attempted.group then return ui.group, ui.states.group end
   ui.attempted.group = true
   local object, state = makeGroup()
   ui.group, ui.states.group = object, state
   return object, state
end

local function makeRadial(): W.RadialProgress, string
   local owner, ownerState = getGroup()
   if ownerState ~= "available" then return nil, ownerState end
   local state = "lookup_error"
   local ok, object = pcall(function(): W.RadialProgress
      if issecretvalue(owner) or not canaccessvalue(owner) then state = "inaccessible"; return nil end
      local kind = type(owner)
      if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
      if kind == "table" and not canaccesstable(owner) then state = "inaccessible_table"; return nil end
      local fn = owner.CreateAnimation
      if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
      if fn == nil then state = "absent"; return nil end
      if type(fn) ~= "function" then state = "unavailable"; return nil end
      state = "call_error"
      return fn(owner, "RadialProgress")
   end)
   if not ok or state ~= "call_error" then return nil, state end
   if issecretvalue(object) or not canaccessvalue(object) then return nil, "inaccessible" end
   if object == nil then return nil, "no_result" end
   local kind = type(object)
   if kind ~= "table" and kind ~= "userdata" then return nil, "invalid_object" end
   if kind == "table" and not canaccesstable(object) then return nil, "inaccessible_table" end
   return object, "available"
end

local function getRadial(): W.RadialProgress, string
   if ui.attempted.radial then return ui.radial, ui.states.radial end
   ui.attempted.radial = true
   local object, state = makeRadial()
   ui.radial, ui.states.radial = object, state
   return object, state
end

local function makeVector(): W.VectorGraphics, string
   local owner, ownerState = getFrame()
   if ownerState ~= "available" then return nil, ownerState end
   local state = "lookup_error"
   local ok, object = pcall(function(): W.VectorGraphics
      if issecretvalue(owner) or not canaccessvalue(owner) then state = "inaccessible"; return nil end
      local kind = type(owner)
      if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
      if kind == "table" and not canaccesstable(owner) then state = "inaccessible_table"; return nil end
      local fn = owner.CreateVectorGraphics
      if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
      if fn == nil then state = "absent"; return nil end
      if type(fn) ~= "function" then state = "unavailable"; return nil end
      state = "call_error"
      return fn(owner, nil, "ARTWORK")
   end)
   if not ok or state ~= "call_error" then return nil, state end
   if issecretvalue(object) or not canaccessvalue(object) then return nil, "inaccessible" end
   if object == nil then return nil, "no_result" end
   local kind = type(object)
   if kind ~= "table" and kind ~= "userdata" then return nil, "invalid_object" end
   if kind == "table" and not canaccesstable(object) then return nil, "inaccessible_table" end
   return object, "available"
end

local function getVector(): W.VectorGraphics, string
   if ui.attempted.vector then return ui.vector, ui.states.vector end
   ui.attempted.vector = true
   local object, state = makeVector()
   ui.vector, ui.states.vector = object, state
   return object, state
end

local function makeFont(): W.FontString, string
   local owner, ownerState = getFrame()
   if ownerState ~= "available" then return nil, ownerState end
   local state = "lookup_error"
   local ok, object = pcall(function(): W.FontString
      if issecretvalue(owner) or not canaccessvalue(owner) then state = "inaccessible"; return nil end
      local kind = type(owner)
      if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
      if kind == "table" and not canaccesstable(owner) then state = "inaccessible_table"; return nil end
      local fn = owner.CreateFontString
      if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
      if fn == nil then state = "absent"; return nil end
      if type(fn) ~= "function" then state = "unavailable"; return nil end
      state = "call_error"
      return fn(owner, nil, "ARTWORK")
   end)
   if not ok or state ~= "call_error" then return nil, state end
   if issecretvalue(object) or not canaccessvalue(object) then return nil, "inaccessible" end
   if object == nil then return nil, "no_result" end
   local kind = type(object)
   if kind ~= "table" and kind ~= "userdata" then return nil, "invalid_object" end
   if kind == "table" and not canaccesstable(object) then return nil, "inaccessible_table" end
   return object, "available"
end

local function getFont(): W.FontString, string
   if ui.attempted.font then return ui.font, ui.states.font end
   ui.attempted.font = true
   local object, state = makeFont()
   ui.font, ui.states.font = object, state
   return object, state
end

local function makeTexture(): W.Texture, string
   local owner, ownerState = getFrame()
   if ownerState ~= "available" then return nil, ownerState end
   local state = "lookup_error"
   local ok, object = pcall(function(): W.Texture
      if issecretvalue(owner) or not canaccessvalue(owner) then state = "inaccessible"; return nil end
      local kind = type(owner)
      if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
      if kind == "table" and not canaccesstable(owner) then state = "inaccessible_table"; return nil end
      local fn = owner.CreateTexture
      if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
      if fn == nil then state = "absent"; return nil end
      if type(fn) ~= "function" then state = "unavailable"; return nil end
      state = "call_error"
      return fn(owner, nil, "ARTWORK")
   end)
   if not ok or state ~= "call_error" then return nil, state end
   if issecretvalue(object) or not canaccessvalue(object) then return nil, "inaccessible" end
   if object == nil then return nil, "no_result" end
   local kind = type(object)
   if kind ~= "table" and kind ~= "userdata" then return nil, "invalid_object" end
   if kind == "table" and not canaccesstable(object) then return nil, "inaccessible_table" end
   return object, "available"
end

local function getTexture(): W.Texture, string
   if ui.attempted.texture then return ui.texture, ui.states.texture end
   ui.attempted.texture = true
   local object, state = makeTexture()
   ui.texture, ui.states.texture = object, state
   return object, state
end

local function makeBinding(): W.DurationTextBindingObject, string
   local state = "lookup_error"
   local ok, object = pcall(function(): W.DurationTextBindingObject
      local namespace = C_DurationUtil
      if issecretvalue(namespace) or not canaccessvalue(namespace) then state = "inaccessible"; return nil end
      if namespace == nil then state = "absent"; return nil end
      local kind = type(namespace)
      if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
      if kind == "table" and not canaccesstable(namespace) then state = "inaccessible_table"; return nil end
      local fn = namespace.CreateDurationTextBinding
      if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
      if fn == nil then state = "absent"; return nil end
      if type(fn) ~= "function" then state = "unavailable"; return nil end
      state = "call_error"
      return fn()
   end)
   if not ok or state ~= "call_error" then return nil, state end
   if issecretvalue(object) or not canaccessvalue(object) then return nil, "inaccessible" end
   if object == nil then return nil, "no_result" end
   local kind = type(object)
   if kind ~= "table" and kind ~= "userdata" then return nil, "invalid_object" end
   if kind == "table" and not canaccesstable(object) then return nil, "inaccessible_table" end
   return object, "available"
end

local function makeClock(): W.LuaDurationManualClock, string
   local state = "lookup_error"
   local ok, object = pcall(function(): W.LuaDurationManualClock
      local namespace = C_DurationUtil
      if issecretvalue(namespace) or not canaccessvalue(namespace) then state = "inaccessible"; return nil end
      if namespace == nil then state = "absent"; return nil end
      local kind = type(namespace)
      if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
      if kind == "table" and not canaccesstable(namespace) then state = "inaccessible_table"; return nil end
      local fn = namespace.CreateManualClock
      if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
      if fn == nil then state = "absent"; return nil end
      if type(fn) ~= "function" then state = "unavailable"; return nil end
      state = "call_error"
      return fn()
   end)
   if not ok or state ~= "call_error" then return nil, state end
   if issecretvalue(object) or not canaccessvalue(object) then return nil, "inaccessible" end
   if object == nil then return nil, "no_result" end
   local kind = type(object)
   if kind ~= "table" and kind ~= "userdata" then return nil, "invalid_object" end
   if kind == "table" and not canaccesstable(object) then return nil, "inaccessible_table" end
   return object, "available"
end

local function makeLocale(): W.LuaLocaleContext, string
   local state = "lookup_error"
   local ok, object = pcall(function(): W.LuaLocaleContext
      local namespace = C_Intl
      if issecretvalue(namespace) or not canaccessvalue(namespace) then state = "inaccessible"; return nil end
      if namespace == nil then state = "absent"; return nil end
      local kind = type(namespace)
      if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
      if kind == "table" and not canaccesstable(namespace) then state = "inaccessible_table"; return nil end
      local fn = namespace.CreateLocaleContext
      if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
      if fn == nil then state = "absent"; return nil end
      if type(fn) ~= "function" then state = "unavailable"; return nil end
      state = "call_error"
      return fn("enUS")
   end)
   if not ok or state ~= "call_error" then return nil, state end
   if issecretvalue(object) or not canaccessvalue(object) then return nil, "inaccessible" end
   if object == nil then return nil, "no_result" end
   local kind = type(object)
   if kind ~= "table" and kind ~= "userdata" then return nil, "invalid_object" end
   if kind == "table" and not canaccesstable(object) then return nil, "inaccessible_table" end
   return object, "available"
end

local function makeSignals(): W.TimedSignalMap, string
   local state = "lookup_error"
   local ok, object = pcall(function(): W.TimedSignalMap
      local namespace = C_Timer
      if issecretvalue(namespace) or not canaccessvalue(namespace) then state = "inaccessible"; return nil end
      if namespace == nil then state = "absent"; return nil end
      local kind = type(namespace)
      if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
      if kind == "table" and not canaccesstable(namespace) then state = "inaccessible_table"; return nil end
      local fn = namespace.NewTimedSignalMap
      if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
      if fn == nil then state = "absent"; return nil end
      if type(fn) ~= "function" then state = "unavailable"; return nil end
      state = "call_error"
      return fn(function(_: N.NumberLike) end)
   end)
   if not ok or state ~= "call_error" then return nil, state end
   if issecretvalue(object) or not canaccessvalue(object) then return nil, "inaccessible" end
   if object == nil then return nil, "no_result" end
   local kind = type(object)
   if kind ~= "table" and kind ~= "userdata" then return nil, "invalid_object" end
   if kind == "table" and not canaccesstable(object) then return nil, "inaccessible_table" end
   return object, "available"
end

local function makeDuration(): W.DurationObject, string
   local state = "lookup_error"
   local ok, object = pcall(function(): W.DurationObject
      local namespace = C_DurationUtil
      if issecretvalue(namespace) or not canaccessvalue(namespace) then state = "inaccessible"; return nil end
      if namespace == nil then state = "absent"; return nil end
      local kind = type(namespace)
      if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
      if kind == "table" and not canaccesstable(namespace) then state = "inaccessible_table"; return nil end
      local fn = namespace.CreateDuration
      if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
      if fn == nil then state = "absent"; return nil end
      if type(fn) ~= "function" then state = "unavailable"; return nil end
      state = "call_error"
      return fn()
   end)
   if not ok or state ~= "call_error" then return nil, state end
   if issecretvalue(object) or not canaccessvalue(object) then return nil, "inaccessible" end
   if object == nil then return nil, "no_result" end
   local kind = type(object)
   if kind ~= "table" and kind ~= "userdata" then return nil, "invalid_object" end
   if kind == "table" and not canaccesstable(object) then return nil, "inaccessible_table" end
   return object, "available"
end

local function makeFormatter(): W.SecondsFormatter, string
   local state = "lookup_error"
   local ok, object = pcall(function(): W.SecondsFormatter
      local namespace = C_StringUtil
      if issecretvalue(namespace) or not canaccessvalue(namespace) then state = "inaccessible"; return nil end
      if namespace == nil then state = "absent"; return nil end
      local kind = type(namespace)
      if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
      if kind == "table" and not canaccesstable(namespace) then state = "inaccessible_table"; return nil end
      local fn = namespace.CreateSecondsFormatter
      if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
      if fn == nil then state = "absent"; return nil end
      if type(fn) ~= "function" then state = "unavailable"; return nil end
      state = "call_error"
      return fn()
   end)
   if not ok or state ~= "call_error" then return nil, state end
   if issecretvalue(object) or not canaccessvalue(object) then return nil, "inaccessible" end
   if object == nil then return nil, "no_result" end
   local kind = type(object)
   if kind ~= "table" and kind ~= "userdata" then return nil, "invalid_object" end
   if kind == "table" and not canaccesstable(object) then return nil, "inaccessible_table" end
   return object, "available"
end

local function makeCurve(): W.ColorCurveObject, string
   local state = "lookup_error"
   local ok, object = pcall(function(): W.ColorCurveObject
      local namespace = C_CurveUtil
      if issecretvalue(namespace) or not canaccessvalue(namespace) then state = "inaccessible"; return nil end
      if namespace == nil then state = "absent"; return nil end
      local kind = type(namespace)
      if kind ~= "table" and kind ~= "userdata" then state = "invalid_object"; return nil end
      if kind == "table" and not canaccesstable(namespace) then state = "inaccessible_table"; return nil end
      local fn = namespace.CreateColorCurve
      if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"; return nil end
      if fn == nil then state = "absent"; return nil end
      if type(fn) ~= "function" then state = "unavailable"; return nil end
      state = "call_error"
      return fn()
   end)
   if not ok or state ~= "call_error" then return nil, state end
   if issecretvalue(object) or not canaccessvalue(object) then return nil, "inaccessible" end
   if object == nil then return nil, "no_result" end
   local kind = type(object)
   if kind ~= "table" and kind ~= "userdata" then return nil, "invalid_object" end
   if kind == "table" and not canaccesstable(object) then return nil, "inaccessible_table" end
   return object, "available"
end

local function getFontObject(): W.Font, string
   local object = GameFontNormal
   if issecretvalue(object) or not canaccessvalue(object) then return nil, "inaccessible" end
   if object == nil then return nil, "no_result" end
   local kind = type(object)
   if kind ~= "table" and kind ~= "userdata" then return nil, "invalid_object" end
   if kind == "table" and not canaccesstable(object) then return nil, "inaccessible_table" end
   return object, "available"
end

local queryNames = {"Animation", "RadialProgress", "Region", "LayeredRegion", "ScriptObject", "UIObject", "VectorGraphics"}

-- Only the Object contract authorizes these calls. Other native families retain
-- presence evidence for the query fields without inventing a call signature.
local function observeObjectQueries(object: W.Object, canonical: string): {T.Observation}
   local rows: {T.Observation} = {}
   local function objectType(): unknown, string
      local ok, fn, state = pcall(function(): ObjectTypeQuery, string
         if issecretvalue(object) or not canaccessvalue(object) then return nil, "inaccessible" end
         if object == nil then return nil, "absent" end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then return nil, "invalid_object" end
         if kind == "table" and not canaccesstable(object) then return nil, "inaccessible_table" end
         local found = object.GetObjectType
         if issecretvalue(found) or not canaccessvalue(found) then return nil, "inaccessible" end
         if found == nil then return nil, "absent" end
         return found, "available"
      end)
      if not ok then return nil, "lookup_error" end
      if state ~= "available" then return nil, state end
      if issecretvalue(fn) or not canaccessvalue(fn) then return nil, "inaccessible" end
      if type(fn) ~= "function" then return nil, "unavailable" end
      local called, value = pcall(fn, object)
      if not called then return nil, "call_error" end
      if issecretvalue(value) or not canaccessvalue(value) then return nil, "inaccessible" end
      if value is nil then return nil, "no_result" end
      return value, "available"
   end
   local value, state = objectType()
   local row: T.Observation = {caseID = canonical .. ".query.GetObjectType", state = state}
   if state == "available" then
      if not issecretvalue(value) and canaccessvalue(value) and value is string and #value <= 96 then row.value = value
      else row.state = "invalid_result" end
   end
   rows[#rows + 1] = row
   for _, name in ipairs(queryNames) do
      local ok, fn, foundState = pcall(function(): IsObjectTypeQuery, string
         if issecretvalue(object) or not canaccessvalue(object) then return nil, "inaccessible" end
         if object == nil then return nil, "absent" end
         local kind = type(object)
         if kind ~= "table" and kind ~= "userdata" then return nil, "invalid_object" end
         if kind == "table" and not canaccesstable(object) then return nil, "inaccessible_table" end
         local found = object.IsObjectType
         if issecretvalue(found) or not canaccessvalue(found) then return nil, "inaccessible" end
         if found == nil then return nil, "absent" end
         return found, "available"
      end)
      local answer: unknown
      state = foundState
      if not ok then state = "lookup_error"
      elseif state == "available" then
         if issecretvalue(fn) or not canaccessvalue(fn) then state = "inaccessible"
         elseif type(fn) ~= "function" then state = "unavailable"
         else
            local called, result = pcall(fn, object, name)
            if not called then state = "call_error"
            elseif issecretvalue(result) or not canaccessvalue(result) then state = "inaccessible"
            elseif result is nil then state = "no_result"
            else answer = result end
         end
      end
      row = {caseID = canonical .. ".query.IsObjectType." .. name, state = state}
      if state == "available" then
         if not issecretvalue(answer) and canaccessvalue(answer) and answer is boolean then row.value = answer
         else row.state = "invalid_result" end
      end
      rows[#rows + 1] = row
   end
   return rows
end

local function observeUncontractedQueries(observe: Observer, canonical: string): {T.Observation}
   local rows: {T.Observation} = {}
   local function state(name: string): string
      local observed = observe(name)
      if observed == "present" then return "unverified_signature" end
      if observed == "not_function" then return "unavailable" end
      return observed
   end
   rows[#rows + 1] = {caseID = canonical .. ".query.GetObjectType", state = state("GetObjectType")}
   for _, name in ipairs(queryNames) do
      rows[#rows + 1] = {caseID = canonical .. ".query.IsObjectType." .. name, state = state("IsObjectType")}
   end
   return rows
end

local function cleanupBinding(binding: W.DurationTextBindingObject): string
   local ok, fn, state = pcall(function(): DisableBinding, string
      if issecretvalue(binding) or not canaccessvalue(binding) then return nil, "inaccessible" end
      if binding == nil then return nil, "absent" end
      local kind = type(binding)
      if kind ~= "table" and kind ~= "userdata" then return nil, "invalid_object" end
      if kind == "table" and not canaccesstable(binding) then return nil, "inaccessible_table" end
      local found = binding.Disable
      if issecretvalue(found) or not canaccessvalue(found) then return nil, "inaccessible" end
      if found == nil then return nil, "absent" end
      return found, "available"
   end)
   if not ok then return "lookup_error" end
   if state ~= "available" then return state end
   if issecretvalue(fn) or not canaccessvalue(fn) or type(fn) ~= "function" then return "not_function" end
   return pcall(fn, binding) and "complete" or "call_error"
end

local function observeHierarchy(resource: Resource, canonical: string, factoryState?: string): {T.Observation}
   local rows: {T.Observation} = {}
   for _, requirement in ipairs(plan.hierarchy) do
      if requirement.canonical == canonical then
         for _, name in ipairs(requirement.methods) do
            local state: string
            if factoryState and factoryState ~= "available" then state = "factory_" .. factoryState
            else state = resource.observe(name) end
            rows[#rows + 1] = {caseID = canonical .. ".parent." .. requirement.parent .. "." .. name, state = state}
         end
      end
   end
   return rows
end

local function resource(state: string, observe: Observer, queries?: Queries): Resource
   return {state = state, observe = observe, queries = queries or function(name: string): {T.Observation} return observeUncontractedQueries(observe, name) end}
end

local function uiResource(canonical: string): Resource
   if canonical == "RadialProgress" then
      local object, state = getRadial()
      return resource(state, observeRadial(object), function(name: string): {T.Observation} return observeObjectQueries(object, name) end)
   elseif canonical == "VectorGraphics" then
      local object, state = getVector()
      return resource(state, observeVector(object), function(name: string): {T.Observation} return observeObjectQueries(object, name) end)
   elseif canonical == "Frame" then
      local object, state = getFrame()
      return resource(state, observeFrame(object))
   elseif canonical == "FontString" then
      local object, state = getFont()
      return resource(state, observeFont(object))
   elseif canonical == "Texture" then
      local object, state = getTexture()
      return resource(state, observeTexture(object))
   else return resource("unsupported_family", function(_: string): string return "absent" end) end
end

local function observeRepresentatives(result: T.Report)
   for _, requirement in ipairs(plan.representatives) do
      local object = uiResource(requirement.canonical)
      for _, row in ipairs(observeHierarchy(object, requirement.canonical, object.state)) do
         result.hierarchy[#result.hierarchy + 1] = row
      end
   end
end

-- Clock identity is shared within one capture. Native mutable resources remain
-- owned by that capture; only UI objects have load-long lifetimes.
local function invocationHost(): T.InvocationHost
   local clock: W.LuaDurationManualClock
   local clockState: string
   return {
      binding = makeBinding,
      clock = function(): W.LuaDurationManualClock, string
         if clockState then return clock, clockState end
         clock, clockState = makeClock()
         return clock, clockState
      end,
      locale = makeLocale, signals = makeSignals,
      radial = getRadial, vector = getVector, font = getFont,
      duration = makeDuration, formatter = makeFormatter, curve = makeCurve,
      fontObject = getFontObject,
      observeHierarchy = function(canonical: string): {T.Observation}
         return observeHierarchy(uiResource(canonical), canonical)
      end,
   }
end

local function construct(host: T.InvocationHost, family: T.Family): Resource
   local canonical = family.canonical
   if canonical == "DurationTextBindingObject" then
      local object, state = host.binding()
      local one = resource(state, observeBinding(object))
      if not issecretvalue(object) and canaccessvalue(object) and object ~= nil then
         one.cleanup = function(): string return cleanupBinding(object) end
      end
      return one
   elseif canonical == "LuaDurationClockObject" or canonical == "LuaDurationManualClock" then
      local object, state = host.clock()
      return resource(state, observeClock(object))
   elseif canonical == "LuaLocaleContext" then
      local object, state = host.locale()
      return resource(state, observeLocale(object))
   elseif canonical == "TimedSignalMap" then
      local object, state = host.signals()
      return resource(state, observeSignals(object))
   else return uiResource(canonical) end
end

local function report(): T.Report
   local version, build, date, interface = GetBuildInfo()
   local project = WOW_PROJECT_ID
   local result: T.Report = {
      schema = 1, probeVersion = plan.version, expectedBuild = plan.expectedBuild,
      sourceFingerprint = plan.sourceFingerprint, declarationFingerprint = plan.declarationFingerprint,
      clientVersion = scalar(version), clientBuild = scalar(build), clientDate = scalar(date),
      clientInterface = number(interface), projectID = number(project),
      context = "out_of_combat_public_owned_objects", outcome = "incomplete",
      expectedMethods = plan.expectedMethods, limits = plan.limits, observedMethods = 0, presentMethods = 0,
      capabilityCaptureComplete = false, verifiesSignatures = false, verifiesSecrecy = false,
      cleanup = {}, families = {},
      mode = "presence", invocations = {}, invokedMethods = 0, verifiedMethods = 0,
      invocationCaptureComplete = false, fixtureSHA256 = plan.fixtureSHA256,
      hierarchy = {}, bindingDiagnostics = {}, bindingTextSamples = {}, bindingExperimentChecks = {},
   }
   return result
end

local function invokeCapture(): T.Report
   local result = report()
   result.mode = "public_invocations"
   -- Refuse an incompatible client before creating or mutating owned fixtures.
   if result.clientVersion .. "." .. result.clientBuild ~= plan.expectedBuild then
      result.outcome = "build_mismatch"
      return result
   end
   if result.projectID ~= plan.expectedProjectID or result.clientInterface ~= plan.expectedInterface then
      result.outcome = "client_context_mismatch"
      return result
   end
   private.RunInvocations(invocationHost(), result)
   observeRepresentatives(result)
   return result
end

local function capture(): T.Report
   local result = report()
   local host = invocationHost()
   local binding: Resource
   local ok = pcall(function()
      for _, family in ipairs(plan.families) do
         local object = construct(host, family)
         local state = object.state
         if family.name == "DurationTextBindingObjectAPI" then binding = object end
         local one: T.FamilyResult = {family = family.name, factory = state, methods = {}, queries = {}}
         result.families[#result.families + 1] = one
         for _, name in ipairs(family.methods) do
            local observed = "factory_" .. state
            if state == "available" then observed = object.observe(name) end
            local row: T.Observation = {caseID = family.canonical .. "." .. name, state = observed}
            one.methods[#one.methods + 1] = row
            result.observedMethods = result.observedMethods + 1
            if observed == "present" then result.presentMethods = result.presentMethods + 1 end
         end
         if state == "available" then
            one.queries = object.queries(family.canonical)
            for _, row in ipairs(observeHierarchy(object, family.canonical)) do result.hierarchy[#result.hierarchy + 1] = row end
         end
      end
      observeRepresentatives(result)
   end)
   if binding and binding.cleanup then
      result.cleanup[#result.cleanup + 1] = {caseID = "cleanup.DurationTextBinding.Disable", state = binding.cleanup()}
   end
   if not ok then result.outcome = "capture_error"; return result end
   local clean = true
   for _, row in ipairs(result.cleanup) do if row.state ~= "complete" then clean = false end end
   result.capabilityCaptureComplete = result.observedMethods == plan.expectedMethods and result.presentMethods == plan.expectedMethods and clean
   if result.clientVersion .. "." .. result.clientBuild ~= plan.expectedBuild then
      result.capabilityCaptureComplete = false
      result.outcome = "build_mismatch"
   elseif result.projectID ~= plan.expectedProjectID or result.clientInterface ~= plan.expectedInterface then
      result.capabilityCaptureComplete = false
      result.outcome = "client_context_mismatch"
   elseif result.capabilityCaptureComplete then result.outcome = "capabilities_collected"
   end
   return result
end

SLASH_WOWAPIPROBE1 = "/api-probe"
SlashCmdList.WOWAPIPROBE = function(command: string)
   if issecretvalue(command) or not canaccessvalue(command) then return end
   if command == "capture" or command == "invoke" then
      if capturing then print("WoWAPIProbe: capture already running"); return end
      local ok, combat = pcall(InCombatLockdown)
      if not ok or issecretvalue(combat) or not canaccessvalue(combat) or combat ~= false then print("WoWAPIProbe: capture unavailable during combat or unknown context"); return end
      capturing = true
      local success, value = pcall(command == "invoke" and invokeCapture or capture)
      capturing = false
      if success then
         last = value
         WoWAPIProbeDB = value
         if command == "invoke" then
            print("WoWAPIProbe: " .. last.outcome .. "; successful invocations " .. tostring(last.invokedMethods) .. "/" .. tostring(last.expectedMethods) .. "; property checks " .. tostring(last.verifiedMethods))
         else print("WoWAPIProbe: " .. last.outcome .. "; method presence " .. tostring(last.presentMethods) .. "/" .. tostring(last.expectedMethods)) end
         print("WoWAPIProbe: signature and secrecy behavior remain unverified; logout/reload writes WoWAPIProbe.lua SavedVariables")
      else print("WoWAPIProbe: capture failed; previous saved report retained")
      end
   elseif command == "clear" then
      WoWAPIProbeDB, last = nil, nil
      print("WoWAPIProbe: saved capture cleared")
   elseif command == "status" then
      if last then print("WoWAPIProbe: " .. last.outcome)
      else print("WoWAPIProbe: no capture in this load; older saved data is not trusted or resumed") end
   else print("WoWAPIProbe: /api-probe capture | invoke | status | clear")
   end
end
