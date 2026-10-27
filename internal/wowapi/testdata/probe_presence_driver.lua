
local mode, root = arg[1], arg[2]
local expected = {DurationTextBindingObject = {"Assign", "CanFormatText", "CanUpdateFontString", "ClearTextColorCurve", "Copy", "Disable", "Enable", "GetDuration", "GetExpiredText", "GetFontString", "GetFormattedText", "GetFormattedTextColor", "GetTextColorCurve", "GetTimeModifier", "GetUpdateInterval", "GetZeroDurationText", "HasSecretValues", "IsEnabled", "SetDuration", "SetEnabled", "SetExpiredText", "SetFontString", "SetFormatter", "SetTextColorCurve", "SetTextFormat", "SetTimeModifier", "SetToDefaults", "SetUpdateInterval", "SetZeroDurationText", "UpdateFontString"}, LuaDurationClockObject = {"GetTime"}, LuaDurationManualClock = {"AdvanceTime", "ResetTime", "RewindTime", "SetTime"}, LuaLocaleContext = {"CompareStrings", "FindBreaks", "FindStringMatches", "FoldCase", "FormatCurrency", "FormatDate", "FormatDateTime", "FormatNumber", "FormatTime", "GetCurrencyName", "GetDisplayName", "GetLocale", "GetSortKey", "Length", "ParseCurrency", "ParseNumber", "SelectPlural", "SetLocale", "ToLower", "ToTitle", "ToUpper", "TransformLocale"}, RadialProgress = {"GetFromPercent", "GetToPercent", "SetFromPercent", "SetToPercent"}, VectorGraphics = {"ClearSVG", "GetSVGFileID", "HasSVG", "SetSVG"}, TimedSignalMap = {"CancelAllSignals", "CancelSignal", "GetNextSignal", "GetSignalCount", "GetSignalTime", "HasSignal", "SignalAfter", "SignalAt"}}
local private, resources, counts = {}, {}, {}
local forbidden, disabled, messages = 0, 0, {}
local secret = {_secret=true}
local proxies, tableProof, firstCaller = {}, {}, {}
local attemptedIndexes, unauthorizedIndexes = 0, 0
local function tick(key) counts[key]=(counts[key] or 0)+1 end
function issecretvalue(value) return type(value)=='table' and rawget(value,'_secret')==true end
function canaccessvalue(value) return not (type(value)=='table' and rawget(value,'_inaccessible')==true) end
function canaccesstable(value)
 if proxies[value] then
  local caller=debug.getinfo(2,'f').func
  firstCaller[value]=firstCaller[value] or caller
  if mode=='caller_denied' and caller~=firstCaller[value] then return false end
  tableProof[value]=caller
 end
 return rawget(value,'_tableInaccessible')~=true
end
local function object(name)
  local result={}
  resources[result]=true
  for _,method in ipairs(expected[name] or {}) do
    result[method]=function() forbidden=forbidden+1; error('method invocation forbidden') end
  end
  result.GetObjectType=function() return name end
  result.IsObjectType=function(_,query)
    assert(type(query)=='string' and #query<32)
    return query==name
  end
  if name=='DurationTextBindingObject' then
    result.Disable=function()
      disabled=disabled+1
      if mode=='cleanup_error' then error(secret) end
    end
    if mode=='absent' then result.ClearTextColorCurve=nil end
    if mode=='inaccessible' then result._inaccessible=true end
    if mode=='secret' then result._secret=true end
    if mode=='table_inaccessible' then result._tableInaccessible=true end
  end
  if name=='RadialProgress' and mode=='query_secret' then result.GetObjectType=function() return secret end end
  if name=='RadialProgress' and mode=='query_malformed' then result.GetObjectType=function() return 42 end end
  if name=='LuaLocaleContext' and mode=='query_absent' then result.GetObjectType=nil; result.IsObjectType=nil end
  if name=='VectorGraphics' then
    result.GetRotation=function() forbidden=forbidden+1; error('candidate invocation forbidden') end
    result.SetRotation=function() forbidden=forbidden+1; error('candidate invocation forbidden') end
    if mode=='absent' then result.GetRotation=nil; result.SetRotation=nil end
    if mode=='inaccessible' then result.GetRotation={_inaccessible=true} end
    if mode=='secret' then result.GetRotation=secret end
  end
  if name=='DurationTextBindingObject' and (mode=='caller_scoped' or mode=='caller_denied' or mode=='caller_lookup_error') then
    local backing=result
    result=setmetatable({}, {__index=function(proxy,key)
      attemptedIndexes=attemptedIndexes+1
      if tableProof[proxy]~=debug.getinfo(2,'f').func then
        unauthorizedIndexes=unauthorizedIndexes+1
        error(secret)
      end
      if mode=='caller_lookup_error' and key=='ClearTextColorCurve' then error(secret) end
      return rawget(backing,key)
    end})
    proxies[result]=true
    resources[result]=true
  end
  return result
end
C_DurationUtil={
 CreateDurationTextBinding=function() tick('binding'); return object('DurationTextBindingObject') end,
 CreateManualClock=function()
   tick('clock'); local clock=object('LuaDurationManualClock')
   clock.GetTime=function() forbidden=forbidden+1; error('clock invocation forbidden') end
   return clock
 end,
}
C_Intl={CreateLocaleContext=function(locale)
 assert(locale=='enUS'); tick('locale')
 if mode=='constructor_error' then error(secret) end
 if mode=='constructor_malformed' then return 42 end
 return object('LuaLocaleContext')
end}
C_Timer={NewTimedSignalMap=function(callback)
 assert(type(callback)=='function'); tick('map'); return object('TimedSignalMap')
end}
function CreateFrame(kind)
 assert(kind=='Frame'); tick('frame')
 if mode=='frame_error' then error(secret) end
 local function representative(name)
  local result=object(name)
  for _,method in ipairs({'SetPreventSecretValues','CreateAnimationGroup','GetAnimationGroups','StopAnimating'}) do
   result[method]=function() forbidden=forbidden+1;error('representative candidate invocation forbidden') end
  end
  if name=='Texture' and mode=='inaccessible' then result.SetPreventSecretValues={_inaccessible=true} end
  if name=='FontString' and mode=='absent' then result.GetAnimationGroups=nil end
  if name=='FontString' and mode=='secret' then result.StopAnimating=secret end
  return result
 end
 local frame={
  Hide=function() tick('hide') end,
  CreateAnimationGroup=function()
   tick('group'); return {CreateAnimation=function(_,kind)
    assert(kind=='RadialProgress'); tick('radial'); return object('RadialProgress')
   end}
  end,
  CreateVectorGraphics=function(_,name,layer)
   assert(name==nil and layer=='ARTWORK'); tick('vector'); return object('VectorGraphics')
  end,
  CreateTexture=function(_,name,layer)
   assert(name==nil and layer=='ARTWORK');tick('texture');return representative('Texture')
  end,
  CreateFontString=function(_,name,layer)
   assert(name==nil and layer=='ARTWORK');tick('font');return representative('FontString')
  end,
 }
 for _,method in ipairs({'SetPreventSecretValues','GetAnimationGroups','StopAnimating'}) do
  frame[method]=function() forbidden=forbidden+1;error('representative candidate invocation forbidden') end
 end
 if mode=='absent' then frame.SetPreventSecretValues=nil end
 if mode=='query_malformed' then frame.StopAnimating=42 end
 resources[frame]=true
 return frame
end
function GetBuildInfo()
 if mode=='report_error' then error(secret) end
 return '12.1.5', mode=='wrong_build' and '00000' or '70575', 'Oct 23 2026', 120105
end
WOW_PROJECT_ID=1
function InCombatLockdown() return mode=='combat' end
SlashCmdList={}
function print(value) assert(type(value)=='string'); messages[#messages+1]=value end
local old={untrusted=true}
WoWAPIProbeDB=old
assert(loadfile(root..'/Plan.lua'))('WoWAPIProbe',private)
assert(loadfile(root..'/Runtime.lua'))('WoWAPIProbe',private)
assert(next(counts)==nil and WoWAPIProbeDB==old,'load must not capture or allocate')
assert(SLASH_WOWAPIPROBE1=='/api-probe')
SlashCmdList.WOWAPIPROBE('status')
assert(WoWAPIProbeDB==old,'status must not trust or replace old data')
SlashCmdList.WOWAPIPROBE('capture')
if mode=='combat' or mode=='report_error' then
 assert(WoWAPIProbeDB==old and next(counts)==nil)
 return
end
local result=WoWAPIProbeDB
assert(result~=old and result.schema==1)
assert(result.expectedBuild=='12.1.5.70575' and result.expectedMethods==73)
assert(result.probeVersion=='1.5.0' and #result.sourceFingerprint==64 and #result.declarationFingerprint==64)
assert(#result.bindingDiagnostics==0,'presence capture must not run binding diagnostics')
assert(#result.bindingTextSamples==0 and #result.bindingExperimentChecks==0)
assert(result.verifiesSignatures==false and result.verifiesSecrecy==false)
assert(result.context=='out_of_combat_public_owned_objects')
assert(result.observedMethods==73 and #result.families==7)
local seen, families={},{}
for _,family in ipairs(result.families) do
 families[family.family]=family
 for _,row in ipairs(family.methods) do
  assert(not seen[row.caseID]); seen[row.caseID]=row
 end
end
local expectedCount=0
for name,methods in pairs(expected) do
 for _,method in ipairs(methods) do
  assert(seen[name..'.'..method], 'missing independently expected case')
  expectedCount=expectedCount+1
 end
end
assert(expectedCount==73)
local candidates={}
local representatives={}
local historical={}
for _,row in ipairs(result.hierarchy) do
 if row.caseID:match('^VectorGraphics%.parent%.ketho%.RotationCandidates%.') then candidates[row.caseID]=row.state end
 if row.caseID:match('%.parent%.native%.RepresentativeCapabilities%.') then representatives[row.caseID]=row.state end
 if row.caseID:match('%.parent%.native%.HistoricalAbsences%.') then historical[row.caseID]=row.state end
end
for _,canonical in ipairs({'Frame','Texture','FontString'}) do
 for _,name in ipairs({'SetPreventSecretValues','CreateAnimationGroup','GetAnimationGroups','StopAnimating'}) do
  local id=canonical..'.parent.native.RepresentativeCapabilities.'..name
  assert(representatives[id], 'missing independently expected representative case')
  local expectedState='present'
  if mode=='frame_error' then expectedState='factory_call_error'
  elseif mode=='absent' and ((canonical=='Frame' and name=='SetPreventSecretValues') or (canonical=='FontString' and name=='GetAnimationGroups')) then expectedState='absent'
  elseif mode=='inaccessible' and canonical=='Texture' and name=='SetPreventSecretValues' then expectedState='inaccessible'
  elseif mode=='secret' and canonical=='FontString' and name=='StopAnimating' then expectedState='inaccessible'
  elseif mode=='query_malformed' and canonical=='Frame' and name=='StopAnimating' then expectedState='not_function' end
  assert(representatives[id]==expectedState,id..':'..representatives[id])
 end
end
if mode~='frame_error' then
 for _,id in ipairs({'RadialProgress.parent.native.HistoricalAbsences.SetPreventSecretValues','VectorGraphics.parent.native.HistoricalAbsences.SetPreventSecretValues','VectorGraphics.parent.native.HistoricalAbsences.CreateAnimationGroup','VectorGraphics.parent.native.HistoricalAbsences.GetAnimationGroups','VectorGraphics.parent.native.HistoricalAbsences.StopAnimating'}) do
  assert(historical[id]=='absent','canonical correction must not drop historical negative queries')
 end
end
local rotation='VectorGraphics.parent.ketho.RotationCandidates.'
if mode=='absent' then
 assert(candidates[rotation..'GetRotation']=='absent' and candidates[rotation..'SetRotation']=='absent')
elseif mode=='inaccessible' or mode=='secret' then
 assert(candidates[rotation..'GetRotation']=='inaccessible' and candidates[rotation..'SetRotation']=='present')
elseif mode~='frame_error' then
 assert(candidates[rotation..'GetRotation']=='present' and candidates[rotation..'SetRotation']=='present')
end
local function sanitized(value)
 assert(value~=secret and not resources[value])
 local kind=type(value)
 assert(kind=='table' or kind=='number' or kind=='string' or kind=='boolean' or kind=='nil')
 if kind=='table' then
  for key,item in pairs(value) do sanitized(key); sanitized(item) end
 end
end
sanitized(result)
local binding=families.DurationTextBindingObjectAPI
local locale=families.LuaLocaleContextAPI
local radial=families.SimpleAnimRadialProgressAPI
if mode=='caller_denied' then
 assert(binding.factory=='available' and binding.methods[1].state=='inaccessible_table')
 assert(attemptedIndexes==0 and unauthorizedIndexes==0)
elseif mode=='caller_lookup_error' then
 assert(seen['DurationTextBindingObject.ClearTextColorCurve'].state=='lookup_error' and result.presentMethods==72)
 assert(attemptedIndexes>0 and unauthorizedIndexes==0)
elseif mode=='caller_scoped' then
 assert(attemptedIndexes>0 and unauthorizedIndexes==0)
elseif mode=='absent' then
 assert(seen['DurationTextBindingObject.ClearTextColorCurve'].state=='absent' and result.presentMethods==72)
elseif mode=='constructor_error' then
 assert(locale.factory=='call_error' and locale.methods[1].state=='factory_call_error')
elseif mode=='constructor_malformed' then
 assert(locale.factory=='invalid_object' and locale.methods[1].state=='factory_invalid_object')
elseif mode=='secret' or mode=='inaccessible' then
 assert(binding.factory=='inaccessible' and binding.methods[1].state=='factory_inaccessible')
elseif mode=='table_inaccessible' then
 assert(binding.factory=='inaccessible_table')
elseif mode=='query_secret' then
 assert(radial.queries[1].state=='inaccessible' and radial.queries[1].value==nil)
elseif mode=='query_malformed' then
 assert(radial.queries[1].state=='invalid_result' and radial.queries[1].value==nil)
elseif mode=='query_absent' then
 assert(#locale.queries==8)
 for _,query in ipairs(locale.queries) do assert(query.state=='absent') end
elseif mode=='wrong_build' then
 assert(result.outcome=='build_mismatch' and result.capabilityCaptureComplete==false)
elseif mode=='cleanup_error' then
 assert(result.cleanup[1].state=='call_error' and result.capabilityCaptureComplete==false)
elseif mode=='frame_error' then
 assert(radial.factory=='call_error' and families.SimpleVectorGraphicsAPI.factory=='call_error')
end
if mode=='normal' or mode=='query_absent' or mode=='query_secret' or mode=='query_malformed' or mode=='caller_scoped' then
 assert(result.outcome=='capabilities_collected' and result.capabilityCaptureComplete==true)
 assert(result.presentMethods==73)
else assert(result.capabilityCaptureComplete==false) end
if mode=='normal' or mode=='caller_scoped' then
 local queryCount=0
 for _,family in ipairs(result.families) do
  for _,query in ipairs(family.queries) do assert(query.state=='available'); queryCount=queryCount+1 end
 end
 assert(queryCount==56)
 assert(radial.queries[2].caseID=='RadialProgress.query.IsObjectType.Animation')
 assert(radial.queries[6].caseID=='RadialProgress.query.IsObjectType.ScriptObject')
 assert(radial.queries[3].value==true)
end
SlashCmdList.WOWAPIPROBE('capture')
assert(WoWAPIProbeDB~=result,'replace the bounded last report atomically')
assert(counts.binding==2 and counts.clock==2 and counts.locale==2 and counts.map==2)
assert(counts.frame==1,'cached UI construction must be bounded across captures, including failures')
if mode~='frame_error' then assert(counts.group==1 and counts.radial==1 and counts.vector==1 and counts.font==1 and counts.texture==1) end
assert(disabled==((mode=='secret' or mode=='inaccessible' or mode=='table_inaccessible' or mode=='caller_denied') and 0 or 2))
assert(forbidden==0,'method presence must not invoke catalogue methods')
SlashCmdList.WOWAPIPROBE('clear')
assert(WoWAPIProbeDB==nil)
