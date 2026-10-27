local type T = require("probe_types")
local _, private = ... as (string, T.Namespace)
local plan = private.Plan
local globals = _G as {string: any}
local last: T.Report
local capturing = false

-- Native access checks belong to the function that performs the operation.
-- The protected lookup checks permission inside its own indexing closure.
local function field(object: any, name: string): any, string
   local ok, value, state = pcall(function(): any, string
      if issecretvalue(object) or not canaccessvalue(object) then return nil, "inaccessible" end
      local kind = type(object)
      if kind ~= "table" and kind ~= "userdata" then return nil, "invalid_object" end
      if kind == "table" and not canaccesstable(object) then return nil, "inaccessible_table" end
      local found = (object as {string: any})[name]
      if issecretvalue(found) or not canaccessvalue(found) then return nil, "inaccessible" end
      if found == nil then return nil, "absent" end
      return found, "available"
   end)
   if not ok then return nil, "lookup_error" end
   return value, state
end

local function resolve(namespace: string, name?: string): any, string
   local object, state = field(globals, namespace)
   if state ~= "available" then return nil, state end
   if name then return field(object, name) end
   return object, state
end

local function invoke(fn: any, ...: any): any, string
   if issecretvalue(fn) or not canaccessvalue(fn) then return nil, "inaccessible" end
   if type(fn) ~= "function" then return nil, "unavailable" end
   local ok, value = pcall((fn as function(any...): any), ...)
   if not ok then return nil, "call_error" end
   if issecretvalue(value) or not canaccessvalue(value) then return nil, "inaccessible" end
   if value == nil then return nil, "no_result" end
   return value, "available"
end

local function owned(object: any, state: string): any, string
   if state ~= "available" then return nil, state end
   if issecretvalue(object) or not canaccessvalue(object) then return nil, "inaccessible" end
   local kind = type(object)
   if kind ~= "table" and kind ~= "userdata" then return nil, "invalid_object" end
   if kind == "table" and not canaccesstable(object) then return nil, "inaccessible_table" end
   return object, state
end

local function create(fn: any, ...: any): any, string
   local object, state = invoke(fn, ...)
   return owned(object, state)
end

local function method(object: any, name: string, ...: any): any, string
   local fn, state = field(object, name)
   if state ~= "available" then return nil, state end
   return invoke(fn, object, ...)
end

local function scalar(value: any): string
   if issecretvalue(value) or not canaccessvalue(value) then return "unavailable" end
   if value is string and #value <= 96 then return value end
   return "unavailable"
end

local function number(value: any): number
   if issecretvalue(value) or not canaccessvalue(value) or not (value is number) then return nil end
   if value ~= value or value == math.huge or value == -math.huge then return nil end
   return value
end

-- WoW frames/children are not destroyable. Each construction is attempted at
-- most once per load, including failures, so repeated captures cannot leak UI.
local ui: {string: any} = {}
local tried: {string: boolean} = {}
local uiState: {string: string} = {}
local function uiObject(key: string): any, string
   if tried[key] then return ui[key], uiState[key] end
   tried[key] = true
   local object: any
   local state: string
   if key == "frame" then
      local fn = resolve("CreateFrame")
      object, state = create(fn, "Frame")
      if state == "available" then
         local hide, hideState = field(object, "Hide")
         if hideState ~= "available" or issecretvalue(hide) or not canaccessvalue(hide) or type(hide) ~= "function" then
            object, state = nil, "hide_unavailable"
         elseif not pcall((hide as function(any...): any), object) then
            object, state = nil, "hide_error"
         end
      end
   else
      local owner: any
      owner, state = uiObject(key == "radial" and "group" or "frame")
      if state == "available" then
         if key == "group" then object, state = method(owner, "CreateAnimationGroup")
         elseif key == "radial" then object, state = method(owner, "CreateAnimation", "RadialProgress")
         elseif key == "vector" then object, state = method(owner, "CreateVectorGraphics", nil, "ARTWORK")
         elseif key == "font" then object, state = method(owner, "CreateFontString", nil, "ARTWORK")
         elseif key == "texture" then object, state = method(owner, "CreateTexture", nil, "ARTWORK")
         end
      end
   end
   object, state = owned(object, state)
   ui[key], uiState[key] = object, state
   return object, state
end

local function construct(family: string, objects: {string: any}, states: {string: string}): any, string
   if family == "SimpleAnimRadialProgressAPI" then return uiObject("radial") end
   if family == "SimpleVectorGraphicsAPI" then return uiObject("vector") end
   if family == "LuaDurationClockObjectAPI" or family == "LuaDurationManualClockAPI" then
      if states.clock then return objects.clock, states.clock end
      local fn = resolve("C_DurationUtil", "CreateManualClock")
      objects.clock, states.clock = create(fn)
      return objects.clock, states.clock
   end
   local fn: any
   if family == "DurationTextBindingObjectAPI" then fn = resolve("C_DurationUtil", "CreateDurationTextBinding")
   elseif family == "LuaLocaleContextAPI" then
      fn = resolve("C_Intl", "CreateLocaleContext")
      return create(fn, "enUS")
   elseif family == "TimedSignalMapAPI" then
      fn = resolve("C_Timer", "NewTimedSignalMap")
      return create(fn, function(_: any) end)
   else return nil, "unsupported_family"
   end
   return create(fn)
end

local queryNames = {"Animation", "RadialProgress", "Region", "LayeredRegion", "ScriptObject", "UIObject", "VectorGraphics"}
local function observeQueries(object: any, canonical: string): {T.Observation}
   local rows: {T.Observation} = {}
   local value, state = method(object, "GetObjectType")
   local row: T.Observation = {caseID = canonical .. ".query.GetObjectType", state = state}
   if state == "available" then
      if not issecretvalue(value) and canaccessvalue(value) and value is string and #value <= 96 then row.value = value
      else row.state = "invalid_result" end
   end
   rows[#rows + 1] = row
   for _, name in ipairs(queryNames) do
      value, state = method(object, "IsObjectType", name)
      row = {caseID = canonical .. ".query.IsObjectType." .. name, state = state}
      if state == "available" then
         if not issecretvalue(value) and canaccessvalue(value) and value is boolean then row.value = value
         else row.state = "invalid_result" end
      end
      rows[#rows + 1] = row
   end
   return rows
end

local function cleanup(objects: {string: any}): {T.Observation}
   local result: {T.Observation} = {}
   local binding = objects.binding
   if issecretvalue(binding) or not canaccessvalue(binding) then
      result[#result + 1] = {caseID = "cleanup.DurationTextBinding.Disable", state = "inaccessible"}
   elseif binding ~= nil then
      local fn, state = field(binding, "Disable")
      if state == "available" and not issecretvalue(fn) and canaccessvalue(fn) and type(fn) == "function" then
         local ok = pcall((fn as function(any...): any), binding)
         result[#result + 1] = {caseID = "cleanup.DurationTextBinding.Disable", state = ok and "complete" or "call_error"}
      else result[#result + 1] = {caseID = "cleanup.DurationTextBinding.Disable", state = state == "available" and "not_function" or state}
      end
   end
   return result
end

local function observeHierarchy(object: any, canonical: string, factoryState?: string): {T.Observation}
   local rows: {T.Observation} = {}
   for _, requirement in ipairs(plan.hierarchy) do
      if requirement.canonical == canonical then
         for _, name in ipairs(requirement.methods) do
            local value: any
            local state: string
            if factoryState and factoryState ~= "available" then state = "factory_" .. factoryState
            else value, state = field(object, name) end
            if state == "available" then
               if issecretvalue(value) or not canaccessvalue(value) then state = "inaccessible"
               else state = type(value) == "function" and "present" or "not_function" end
            end
            rows[#rows + 1] = {caseID = canonical .. ".parent." .. requirement.parent .. "." .. name, state = state}
         end
      end
   end
   return rows
end

local function observeRepresentatives(result: T.Report)
   for _, requirement in ipairs(plan.representatives) do
      local object, state = uiObject(requirement.resourceKey)
      for _, row in ipairs(observeHierarchy(object, requirement.canonical, state)) do
         result.hierarchy[#result.hierarchy + 1] = row
      end
   end
end

local function report(): T.Report
   local version, build, date, interface = GetBuildInfo()
   local project = resolve("WOW_PROJECT_ID")
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
   private.RunInvocations({field = field, resolve = resolve, create = create, construct = construct, uiObject = uiObject, observeHierarchy = observeHierarchy}, result)
   observeRepresentatives(result)
   return result
end

local function capture(): T.Report
   local result = report()
   local objects: {string: any} = {}
   local states: {string: string} = {}
   local ok = pcall(function()
      for _, family in ipairs(plan.families) do
         local object, state = construct(family.name, objects, states)
         if family.name == "DurationTextBindingObjectAPI" then objects.binding = object end
         local one: T.FamilyResult = {family = family.name, factory = state, methods = {}, queries = {}}
         result.families[#result.families + 1] = one
         for _, name in ipairs(family.methods) do
            local value: any
            local observed = "factory_" .. state
            if state == "available" then
               value, observed = field(object, name)
               if observed == "available" then
                  if issecretvalue(value) or not canaccessvalue(value) then observed = "inaccessible"
                  else observed = type(value) == "function" and "present" or "not_function" end
               end
            end
            local row: T.Observation = {caseID = family.canonical .. "." .. name, state = observed}
            one.methods[#one.methods + 1] = row
            result.observedMethods = result.observedMethods + 1
            if observed == "present" then result.presentMethods = result.presentMethods + 1 end
         end
         if state == "available" then
            one.queries = observeQueries(object, family.canonical)
            for _, row in ipairs(observeHierarchy(object, family.canonical)) do result.hierarchy[#result.hierarchy + 1] = row end
         end
      end
      observeRepresentatives(result)
   end)
   result.cleanup = cleanup(objects)
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
