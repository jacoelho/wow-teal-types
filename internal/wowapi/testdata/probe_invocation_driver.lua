
local mode, root = arg[1], arg[2]
local expected = {DurationTextBindingObject = {"Assign", "CanFormatText", "CanUpdateFontString", "ClearTextColorCurve", "Copy", "Disable", "Enable", "GetDuration", "GetExpiredText", "GetFontString", "GetFormattedText", "GetFormattedTextColor", "GetTextColorCurve", "GetTimeModifier", "GetUpdateInterval", "GetZeroDurationText", "HasSecretValues", "IsEnabled", "SetDuration", "SetEnabled", "SetExpiredText", "SetFontString", "SetFormatter", "SetTextColorCurve", "SetTextFormat", "SetTimeModifier", "SetToDefaults", "SetUpdateInterval", "SetZeroDurationText", "UpdateFontString"}, LuaDurationClockObject = {"GetTime"}, LuaDurationManualClock = {"AdvanceTime", "ResetTime", "RewindTime", "SetTime"}, LuaLocaleContext = {"CompareStrings", "FindBreaks", "FindStringMatches", "FoldCase", "FormatCurrency", "FormatDate", "FormatDateTime", "FormatNumber", "FormatTime", "GetCurrencyName", "GetDisplayName", "GetLocale", "GetSortKey", "Length", "ParseCurrency", "ParseNumber", "SelectPlural", "SetLocale", "ToLower", "ToTitle", "ToUpper", "TransformLocale"}, RadialProgress = {"GetFromPercent", "GetToPercent", "SetFromPercent", "SetToPercent"}, VectorGraphics = {"ClearSVG", "GetSVGFileID", "HasSVG", "SetSVG"}, TimedSignalMap = {"CancelAllSignals", "CancelSignal", "GetNextSignal", "GetSignalCount", "GetSignalTime", "HasSignal", "SignalAfter", "SignalAt"}}
local expectedDiagnostics = {"setup.font_object_set", "setup.clock_set_time", "setup.duration_set_clock", "setup.duration_set_span", "setup.font_object_identity", "setup.duration_clock_identity", "setup.duration_clock_time", "setup.duration_remaining", "baseline.expected_before", "baseline.expected_after", "baseline.font_after", "baseline.expected_stable", "baseline.font_matches", "sentinel.set", "sentinel.read", "sentinel.matches", "retry.update", "retry.expected", "retry.font", "retry.matches", "zero_interval.set", "zero_interval.update", "zero_interval.expected", "zero_interval.font", "zero_interval.matches"}
local expectedSamples, expectedExperiments = {"baseline.expected", "baseline.font", "projection.font", "manual_step.expected", "manual_step.font", "default_clock.expected", "default_clock.font"}, {"projection.set", "projection.matches", "manual_step.set_time", "manual_step.clock_time", "manual_step.remaining", "manual_step.update", "manual_step.matches", "manual_step.restore_time", "default_clock.time", "default_clock.clear", "default_clock.span", "default_clock.clock_time", "default_clock.remaining", "default_clock.update", "default_clock.matches"}
local private, handles, counts, bindings, maps = {}, {}, {}, {}, {}
local secret = {_secret=true}
local constructors, scheduled, mutations = 0, 0, 0
local candidateCalls=0
local nativeCalls={}
local function countCall(name) nativeCalls[name]=(nativeCalls[name] or 0)+1 end
local function nativeText(value)
 if mode=='native_plural' then
  local projections={['48 |4Second:Seconds;']='48 Seconds',['40 |4Second:Seconds;']='40 Seconds',['1 |4Minute:Minutes;']='1 Minute'}
  return projections[value] or value
 end
 if mode~='sample_bytes' then return value end
 return value:match('^[^%z]*')
end
function issecretvalue(v)
 if mode=='pending_abort' and v=='ABORT_AFTER_PENDING' then error(secret) end
 return type(v)=='table' and rawget(v,'_secret')==true
end
function canaccessvalue(v) return not (type(v)=='table' and rawget(v,'_inaccessible')==true) end
function canaccessallvalues(...)
 for i=1,select('#',...) do if not canaccessvalue(select(i,...)) then return false end end
 return true
end
function canaccesstable(v) return not (type(v)=='table' and rawget(v,'_blocked')==true) end
local function object(kind)
 constructors=constructors+1
 local v={kind=kind}; handles[v]=true
 v.GetObjectType=function() return kind end
 v.IsObjectType=function(_,query) return query==kind end
 if kind=='frame' or kind=='font' or kind=='texture' then
  for _,name in ipairs({'SetPreventSecretValues','CreateAnimationGroup','GetAnimationGroups','StopAnimating'}) do
   v[name]=function() candidateCalls=candidateCalls+1;error('representative candidate invocation forbidden') end
  end
 end
 return v
end
local function binding()
 local v=object('binding'); bindings[#bindings+1]=v
 v.enabled=false;v.interval=0;v.modifier=0
 v.SetToDefaults=function(self) self.duration=nil; self.font=nil; self.formatter=nil; self.expired=nil;self.zero=nil;self.color=nil end
 v.SetDuration=function(self,duration) assert(handles[duration]); self.duration=duration end
 v.GetDuration=function(self) return self.duration end
 v.SetFontString=function(self,font) assert(handles[font]); self.font=font end
 v.GetFontString=function(self) return self.font end
 v.SetFormatter=function(self,formatter) assert(handles[formatter]);self.formatter=formatter end
 v.SetTextFormat=function(self,format,parts) assert(format=='{}' and #parts==1 and parts[1].property==0 and handles[parts[1].formatter]);self.formatter=parts[1].formatter end
 v.CanFormatText=function(self) return self.duration~=nil and self.formatter~=nil end
 v.CanUpdateFontString=function(self) return self:CanFormatText() and self.font~=nil end
 v.SetExpiredText=function(self,text) self.expired=text end
 v.GetExpiredText=function(self) return self.expired end
 v.SetZeroDurationText=function(self,text) self.zero=text end
 v.GetZeroDurationText=function(self) return self.zero end
 v.SetTimeModifier=function(self,n) assert(n==1);self.modifier=n end
 v.GetTimeModifier=function(self) return self.modifier end
 v.SetUpdateInterval=function(self,n) countCall('SetUpdateInterval');assert(n==0.25 or n==0);self.interval=n end
 v.GetUpdateInterval=function(self) return self.interval end
 v.SetTextColorCurve=function(self,curve,property) assert(handles[curve] and property==0);self.color=curve end
 v.GetTextColorCurve=function(self) if self.color then return self.color,0 end end
 v.GetFormattedTextColor=function(self) assert(self.color);return {r=1,g=1,b=1,a=1} end
 v.ClearTextColorCurve=function(self) self.color=nil end
 local formattedReads=0
 local function formattedText()
  if mode=='secret_return' then return secret end
  if mode=='inaccessible_return' then return {_inaccessible=true} end
  if mode=='native_error' then error(secret) end
  if mode=='diag_changing' then formattedReads=formattedReads+1;return 'changing '..formattedReads end
  if mode=='sample_empty' then return '' end
  if mode=='sample_128' then return string.rep('X',128) end
  if mode=='sample_129' then return string.rep('X',129) end
  if mode=='sample_bytes' then return 'A\0\195\169\255' end
  local d=v.duration
  local clockTime=d.clock and d.clock.time or 100
  if mode=='native_plural' then
   local text={[47.5]='48 |4Second:Seconds;',[40]='40 |4Second:Seconds;',[60]='1 |4Minute:Minutes;'}
   return assert(text[d.b-clockTime])
  end
  return tostring(math.max(0,d.b-clockTime))..' sec'
 end
 v.GetFormattedText=function() countCall('GetFormattedText');return formattedText() end
 v.UpdateFontString=function(self)
  countCall('UpdateFontString')
  assert(self.font)
  if not self.enabled then return end
  if mode=='update_error' then error(secret) end
  if mode=='diag_noop' or mode=='projection_noop_stale' or (mode=='diag_zero_interval' and self.interval~=0) then return end
  if mode=='experiment_error' and nativeCalls.UpdateFontString==4 then error(secret) end
  if mode=='clock_split' then self.font.text=self.duration.b<=100 and 'EXPIRED' or tostring(self.duration.b-100)..' sec';return end
  local output=mode=='wrong_font' and 'CORRUPTED' or formattedText()
  self.font.text=type(output)=='string' and nativeText(output) or output
  if mode=='update_tuple' then return true end
 end
 v.HasSecretValues=function() return false end
 v.Enable=function(self) self.enabled=true end
 v.IsEnabled=function(self) return self.enabled end
 v.SetEnabled=function(self,on) assert(on==false);self.enabled=on end
 v.Disable=function(self)
  self.enabled=false
  self.disables=(self.disables or 0)+1
  if mode=='cleanup_error' and self.disables>1 then error(secret) end
 end
 v.Copy=function(self)
  local copy=binding(); copy.expired=self.expired;copy.enabled=self.enabled
  return copy
 end
 v.Assign=function(self,copy) assert(handles[copy]);self.expired=copy.expired end
 return v
end
C_DurationUtil={
 CreateDurationTextBinding=binding,
 CreateManualClock=function()
  local clock=object('clock');clock.time=0
  clock.SetTime=function(self,n) countCall('SetTime');self.time=n end
  clock.AdvanceTime=function(self,n) self.time=self.time+n end
  clock.RewindTime=function(self,n) self.time=self.time-n end
  clock.ResetTime=function(self) self.time=0 end
  clock.GetTime=function(self) return self.time end
  return clock
 end,
 CreateDuration=function()
  local d=object('duration')
  d.SetClock=function(self,...)
   countCall('SetClock');assert(select('#',...)==1,'SetClock must receive explicit nil when clearing')
   local c=...;assert(c==nil or handles[c]);self.clock=c
  end
  d.SetTimeSpan=function(self,a,b) countCall('SetTimeSpan');assert((a==0 and b==60) or (a==100 and b==160));self.a=a;self.b=b end
  d.GetClock=function(self) countCall('GetClock');return self.clock end
  d.GetClockTime=function(self) countCall('GetClockTime');return self.clock and self.clock.time or 100 end
  d.GetRemainingDuration=function(self)
   countCall('GetRemainingDuration')
   if mode=='diag_fact_secret' then return secret end
   if mode=='diag_fact_inaccessible' then return {_inaccessible=true} end
   if mode=='diag_fact_wrong_type' then return '47.5' end
   if mode=='diag_fact_wrong_number' then return 48 end
   return self.b-(self.clock and self.clock.time or 100)
  end
  return d
 end,
}
C_StringUtil={CreateSecondsFormatter=function() return object('formatter') end}
C_CurveUtil={CreateColorCurve=function()
 local curve=object('curve')
 curve.AddPoint=function(_,x,color) assert((x==0 or x==1) and color.r==1 and color.a==1) end
 return curve
end}
C_Intl={CreateLocaleContext=function(locale)
 assert(locale=='enUS')
 local context=object('locale');context.locale=locale
 context.SetLocale=function(self,id) assert(id=='en_US');self.locale=id;return true end
 context.GetLocale=function(self) return self.locale end
 context.CompareStrings=function(_,a,b,strength) assert(a=='abc' and b=='abc' and strength==4);return 0 end
 context.FindBreaks=function(_,value,kind) assert(value=='abc' and kind==0);return {0,1,2,3} end
 context.FindStringMatches=function(_,value,pattern,strength) assert(value=='abc abc' and pattern=='abc' and strength==4);return {0,4} end
 context.FoldCase=function(_,s) assert(s=='ABC');return 'abc' end
 context.FormatCurrency=function(_,n,code) assert(n==12.5 and code=='USD');return '$12.50' end
 context.FormatDate=function(_,n,style,zone) assert(n==0 and style==3 and zone=='UTC');return 'January 1, 1970' end
 context.FormatDateTime=function(_,n,d,t,zone) assert(n==0 and d==3 and t==1 and zone=='UTC');return 'January 1, 1970 12:00 AM' end
 context.FormatNumber=function(_,n,style) assert(n==12.5 and style==0);return '12.5' end
 context.FormatTime=function(_,n,style,zone) assert(n==0 and style==1 and zone=='UTC');return '12:00 AM' end
 context.GetCurrencyName=function(_,code,style) assert(code=='USD' and style==0);return '$' end
 context.GetDisplayName=function(_,id) assert(id=='en_US');return 'English (United States)' end
 context.GetSortKey=function(_,s,strength) assert(s=='abc' and strength==4);return 'abc\0key' end
 context.Length=function(_,s) assert(s=='abc');return 3 end
 context.ParseCurrency=function(_,s) assert(s=='$12.50');return {amount=12.5,currencyCode='USD'} end
 context.ParseNumber=function(_,s,style) assert(s=='12.5' and style==0);return 12.5 end
 context.SelectPlural=function(_,n,kind) assert(n==1 and kind==0);return 'one' end
 context.ToLower=function(_,s) assert(s=='ABC');return 'abc' end
 context.ToTitle=function(_,s) assert(s=='hello world');return 'Hello World' end
 context.ToUpper=function(_,s) assert(s=='abc');return 'ABC' end
 context.TransformLocale=function(_,kind) assert(kind==3);return 'en' end
 return context
end}
C_Timer={NewTimedSignalMap=function(callback)
 assert(type(callback)=='function')
 local map=object('map');map.signals={};maps[#maps+1]=map
 map.SignalAt=function(self,key,time) assert(key==101 and time==86500);scheduled=scheduled+1;self.signals[key]=time end
 map.SignalAfter=function(self,key,delay) assert(key==202 and delay==172800);scheduled=scheduled+1;self.signals[key]=100+delay end
 map.HasSignal=function(self,key) return self.signals[key]~=nil end
 map.GetSignalCount=function(self) local n=0;for _ in pairs(self.signals) do n=n+1 end;return n end
 map.GetSignalTime=function(self,key) return self.signals[key] end
 map.GetNextSignal=function(self)
  if mode=='tuple_short' then return 101 end
  if mode=='tuple_swap' then return 86500,101 end
  return 101,86500
 end
 map.CancelSignal=function(self,key) assert(key==101);self.signals[key]=nil end
 map.CancelAllSignals=function(self) self.signals={} end
 return map
end}
GameFontNormal={}
function CreateFrame(kind)
 assert(kind=='Frame');counts.frame=(counts.frame or 0)+1
 local frame=object('frame')
 frame.Hide=function() end
 frame.CreateFontString=function(_,name,layer)
  assert(name==nil and layer=='ARTWORK');counts.font=(counts.font or 0)+1
  local font=object('font');font.text=mode=='projection_noop_stale' and 'PROBE_SENTINEL' or ''
  font.SetFontObject=function(self,fontObject)
   assert(fontObject==GameFontNormal)
   if mode=='diag_setup_error' then error(secret) end
   self.fontObject=fontObject
   if mode=='diag_setter_tuple' then return true end
  end
  font.GetFontObject=function(self)
   countCall('GetFontObject')
   if mode=='diag_fact_wrong_identity' then return {} end
   return self.fontObject
  end
  font.SetText=function(self,value)
   countCall('SetText');assert(type(value)=='string' and #value<=128)
   if mode=='diag_sentinel_error' then error(secret) end
   if mode=='projection_error' and value~='PROBE_SENTINEL' then error(secret) end
   if (mode=='projection_noop' or mode=='projection_noop_stale') and value~='PROBE_SENTINEL' then return end
   if mode~='diag_sentinel_noop' then self.text=nativeText(value) end
  end
  font.GetText=function(self)
   countCall('GetText')
   if mode=='pending_abort' then return 'ABORT_AFTER_PENDING' end
   if mode=='diag_read_call_error' then error(secret) end
   if mode=='diag_read_zero' then return end
   if mode=='diag_read_multi' then return self.text,'extra' end
   if mode=='diag_read_nil' then return nil end
   if mode=='diag_read_secret' then return secret end
   if mode=='diag_read_inaccessible' then return {_inaccessible=true} end
   if mode=='diag_read_wrong_type' then return 47.5 end
   if mode=='diag_read_oversized' then return string.rep('X',4097) end
   if mode=='diag_read_empty' then return '' end
   if mode=='diag_read_nonempty' then return 'NATIVE_CONTENT_SHOULD_STAY_PRIVATE' end
   return self.text
  end
  return font
 end
 frame.CreateTexture=function(_,name,layer)
  assert(name==nil and layer=='ARTWORK');counts.texture=(counts.texture or 0)+1
  return object('texture')
 end
 frame.CreateAnimationGroup=function()
  counts.group=(counts.group or 0)+1
  local group=object('group')
  group.CreateAnimation=function(_,name)
   assert(name=='RadialProgress');counts.radial=(counts.radial or 0)+1
   local radial=object('radial')
   radial.SetFromPercent=function(self,p) assert(p==0.25);self.from=p end
   radial.SetToPercent=function(self,p) assert(p==0.75);self.to=p end
   radial.GetFromPercent=function(self) if mode=='wrong_getter' then return 0.5 end;return self.from end
   radial.GetToPercent=function(self) return self.to end
   return radial
  end
  return group
 end
 frame.CreateVectorGraphics=function(_,name,layer)
  assert(name==nil and layer=='ARTWORK');counts.vector=(counts.vector or 0)+1
  local vector=object('vector')
  vector.SetSVG=function(self,path)
   assert(path=='Interface\\AddOns\\WoWAPIProbe\\Fixture.svg')
   if mode=='svg_error' then error(secret) end
   if mode=='svg_unavailable' then return false end
   self.svg=true;return true
  end
  vector.GetSVGFileID=function() return 0 end
  vector.HasSVG=function(self) return self.svg==true end
  vector.ClearSVG=function(self) self.svg=false end
  return vector
 end
 return frame
end
function GetTime()
 countCall('GetTime')
 if mode=='default_time_secret' and nativeCalls.GetTime==1 then return secret end
 return 100
end
function GetBuildInfo() return '12.1.5',mode=='wrong_build' and '0' or '70575','Oct 23 2026',mode=='wrong_interface' and 0 or 120105 end
function InCombatLockdown() if mode=='unknown_context' then return nil end;return mode=='combat' end
WOW_PROJECT_ID=mode=='wrong_project' and 2 or 1
SlashCmdList={}
function print() end
local old={untrusted=true}
WoWAPIProbeDB=old
for _,file in ipairs({'Plan.lua','Runtime.lua','Invocations.lua'}) do assert(loadfile(root..'/'..file))('WoWAPIProbe',private) end
assert(constructors==0 and scheduled==0,'load must remain inert')
SlashCmdList.WOWAPIPROBE('invoke')
if mode=='combat' or mode=='unknown_context' then assert(WoWAPIProbeDB==old and constructors==0 and scheduled==0);return end
local result=WoWAPIProbeDB
if mode=='wrong_build' then assert(result.outcome=='build_mismatch' and constructors==0 and scheduled==0 and #result.bindingDiagnostics==0 and #result.bindingTextSamples==0 and #result.bindingExperimentChecks==0);return end
if mode=='wrong_interface' or mode=='wrong_project' then assert(result.outcome=='client_context_mismatch' and constructors==0 and scheduled==0 and #result.bindingDiagnostics==0 and #result.bindingTextSamples==0 and #result.bindingExperimentChecks==0);return end
assert(result.mode=='public_invocations' and result.verifiesSecrecy==false and result.verifiesSignatures==false)
assert(#result.invocations==73 and result.expectedMethods==73)
local seen={}
for _,row in ipairs(result.invocations) do assert(not seen[row.caseID] and row.state~='pending_projection');seen[row.caseID]=row end
assert(seen['DurationTextBindingObject.UpdateFontString'].check=='font_matches_native_projection')
for canonical,methods in pairs(expected) do
 for _,method in ipairs(methods) do assert(seen[canonical..'.'..method],'missing independent invocation case') end
end
local diagnostics={}
assert(#result.bindingDiagnostics==25)
for _,row in ipairs(result.bindingDiagnostics) do
 assert(not diagnostics[row.caseID]);diagnostics[row.caseID]=row.state
 for key in pairs(row) do assert(key=='caseID' or key=='state','diagnostic payload must contain exactly ID/state') end
end
for _,id in ipairs(expectedDiagnostics) do assert(diagnostics[id],'missing literal diagnostic ID '..id) end
local experiments,samples={},{}
assert(#result.bindingExperimentChecks==15 and #result.bindingTextSamples==7)
for _,row in ipairs(result.bindingExperimentChecks) do
 assert(not experiments[row.caseID]);experiments[row.caseID]=row.state
 for key in pairs(row) do assert(key=='caseID' or key=='state') end
end
for _,id in ipairs(expectedExperiments) do assert(experiments[id],'missing experiment '..id) end
for _,row in ipairs(result.bindingTextSamples) do
 assert(not samples[row.caseID]);samples[row.caseID]=row
 local fields=0
 for key in pairs(row) do
  fields=fields+1
  assert(key=='caseID' or key=='state' or (row.state=='public' and (key=='hex' or key=='byteLength' or key=='firstNul')),'sample fields')
 end
 assert(fields==(row.state=='public' and 5 or 2),'exact public/unavailable sample shape')
 if row.state=='public' then
  assert(type(row.hex)=='string' and row.hex:match('^[0-9a-f]*$') and #row.hex==row.byteLength*2 and row.byteLength<=128)
  assert(row.firstNul>=0 and row.firstNul<=row.byteLength)
 end
end
for _,id in ipairs(expectedSamples) do assert(samples[id],'missing sample '..id) end
local function sampleIs(id,hex,length,nul)
 local row=samples[id]
 assert(row.state=='public' and row.hex==hex and row.byteLength==length and row.firstNul==nul,id)
end
local function sanitized(v)
 assert(not handles[v] and v~=secret)
 assert(v~='PROBE_SENTINEL' and v~='47.5 sec' and v~='NATIVE_CONTENT_SHOULD_STAY_PRIVATE')
 local kind=type(v);assert(kind=='nil' or kind=='table' or kind=='string' or kind=='number' or kind=='boolean')
 if kind=='table' then for key,item in pairs(v) do sanitized(key);sanitized(item) end end
end
sanitized(result)
for _,v in ipairs(bindings) do assert(v.enabled==false and (v.disables or 0)>0,'every binding must be disabled') end
for _,v in ipairs(maps) do assert(next(v.signals)==nil,'no pending native signals may survive') end
assert(candidateCalls==0,'representative capabilities must remain lookup-only')
if mode=='pending_abort' then
 local row=seen['DurationTextBindingObject.UpdateFontString']
 assert(row.state=='projection_unavailable' and row.invoked and row.returnCount==0 and not row.behaviorVerified)
 assert(seen['LuaDurationManualClock.RewindTime'].state=='fixture_setup_failed' and not seen['VectorGraphics.ClearSVG'].invoked)
 local invoked,verified=0,0
 for _,one in ipairs(result.invocations) do if one.invoked then invoked=invoked+1 end;if one.behaviorVerified then verified=verified+1 end end
 assert(result.invokedMethods==invoked and result.verifiedMethods==verified and invoked>0 and verified<invoked)
 assert(not result.invocationCaptureComplete)
 SlashCmdList.WOWAPIPROBE('invoke')
 for _,one in ipairs(WoWAPIProbeDB.invocations) do assert(one.state~='pending_projection') end
 assert(counts.frame==1 and counts.font==1 and counts.texture==1 and counts.group==nil)
 for _,one in ipairs(bindings) do assert(one.enabled==false and (one.disables or 0)>0) end
 return
end
assert(seen['LuaDurationManualClock.RewindTime'].state=='passed')
assert(seen['LuaDurationManualClock.ResetTime'].state=='passed')
assert(seen['VectorGraphics.ClearSVG'].invoked==true,'unrelated later cases must still run')
if mode=='normal' then
 assert(result.outcome=='public_invocations_verified' and result.invokedMethods==73 and result.verifiedMethods==73)
 assert(result.invocationCaptureComplete==true)
 for _,row in ipairs(result.invocations) do assert(row.state=='passed',row.caseID..':'..row.state) end
 for _,id in ipairs({'setup.font_object_set','setup.clock_set_time','setup.duration_set_clock','setup.duration_set_span','sentinel.set','retry.update','zero_interval.set','zero_interval.update'}) do assert(diagnostics[id]=='called_void',id) end
 for _,id in ipairs({'setup.font_object_identity','setup.duration_clock_identity','setup.duration_clock_time','setup.duration_remaining','baseline.expected_stable','baseline.font_matches','sentinel.matches','retry.matches','zero_interval.matches'}) do assert(diagnostics[id]=='matched',id) end
 for _,id in ipairs({'baseline.expected_before','baseline.expected_after','baseline.font_after','sentinel.read','retry.expected','retry.font','zero_interval.expected','zero_interval.font'}) do assert(diagnostics[id]=='public_nonempty',id) end
 local originalCalls=10 -- pre-1.3: five binding/font calls, two SetTime, SetClock, SetTimeSpan, GetTime.
 local observed=0
 for name,n in pairs({GetFontObject=1,GetClock=1,GetClockTime=3,GetRemainingDuration=3,GetFormattedText=7,GetText=7,SetText=2,UpdateFontString=5,SetUpdateInterval=2,SetTime=4,SetClock=2,SetTimeSpan=2,GetTime=2}) do
  assert(nativeCalls[name]==n,name..' call bound');observed=observed+n
 end
 assert(observed-originalCalls==31,'exact additional native-call budget')
 sampleIs('baseline.expected','34372e3520736563',8,0)
 sampleIs('baseline.font','34372e3520736563',8,0)
 sampleIs('projection.font','34372e3520736563',8,0)
 sampleIs('manual_step.expected','343020736563',6,0)
 sampleIs('manual_step.font','343020736563',6,0)
 sampleIs('default_clock.expected','363020736563',6,0)
 sampleIs('default_clock.font','363020736563',6,0)
 for _,id in ipairs({'projection.set','manual_step.set_time','manual_step.update','manual_step.restore_time','default_clock.clear','default_clock.span','default_clock.update'}) do assert(experiments[id]=='called_void',id) end
 for _,id in ipairs({'projection.matches','manual_step.clock_time','manual_step.remaining','manual_step.matches','default_clock.time','default_clock.clock_time','default_clock.remaining','default_clock.matches'}) do assert(experiments[id]=='matched',id) end
elseif mode=='wrong_getter' then assert(seen['RadialProgress.GetFromPercent'].state=='property_mismatch')
elseif mode=='wrong_font' then assert(seen['DurationTextBindingObject.UpdateFontString'].state=='property_mismatch' and seen['DurationTextBindingObject.UpdateFontString'].invoked==true)
elseif mode=='tuple_short' or mode=='tuple_swap' then assert(seen['TimedSignalMap.GetNextSignal'].state=='property_mismatch')
elseif mode=='secret_return' then assert(seen['DurationTextBindingObject.GetFormattedText'].state=='secret_result')
elseif mode=='inaccessible_return' then assert(seen['DurationTextBindingObject.GetFormattedText'].state=='inaccessible_result')
elseif mode=='native_error' then assert(seen['DurationTextBindingObject.GetFormattedText'].state=='call_error')
elseif mode=='svg_error' or mode=='svg_unavailable' then assert(seen['VectorGraphics.SetSVG'].state=='fixture_unavailable')
elseif mode=='cleanup_error' then assert(result.invocationCaptureComplete==false)
elseif mode=='diag_noop' or mode=='diag_zero_interval' then
 assert(seen['DurationTextBindingObject.UpdateFontString'].state=='property_mismatch' and result.verifiedMethods==72)
 assert(diagnostics['sentinel.matches']=='matched' and diagnostics['retry.matches']=='matched' and diagnostics['zero_interval.matches']=='matched')
 assert(result.invocationCaptureComplete==false,'diagnostic success must not promote original failure')
elseif mode=='diag_changing' then
 assert(diagnostics['baseline.expected_stable']=='mismatch')
 assert(experiments['projection.set']=='not_run' and seen['DurationTextBindingObject.UpdateFontString'].state=='projection_unavailable','unstable expected text cannot verify projection')
elseif mode=='diag_sentinel_error' or mode=='diag_sentinel_noop' then
 assert(diagnostics['sentinel.matches']=='mismatch' and result.verifiedMethods==72 and seen['DurationTextBindingObject.UpdateFontString'].state=='projection_unavailable')
 assert(diagnostics['sentinel.set']==(mode=='diag_sentinel_error' and 'call_error' or 'called_void'))
elseif mode=='diag_setup_error' then assert(diagnostics['setup.font_object_set']=='call_error' and diagnostics['setup.font_object_identity']=='not_comparable')
elseif mode=='diag_setter_tuple' then assert(diagnostics['setup.font_object_set']=='unexpected_tuple')
elseif mode=='diag_fact_wrong_identity' then assert(diagnostics['setup.font_object_identity']=='mismatch')
elseif mode=='diag_fact_wrong_number' then assert(diagnostics['setup.duration_remaining']=='mismatch')
elseif mode:match('^diag_fact_') then assert(diagnostics['setup.duration_remaining']=='not_comparable')
elseif mode:match('^diag_read_') then
 local states={call_error='call_error',zero='unexpected_tuple',multi='unexpected_tuple',['nil']='nil',secret='secret',inaccessible='inaccessible',wrong_type='wrong_type',oversized='oversized',empty='public_empty',nonempty='public_nonempty'}
 local state=states[mode:sub(11)]
 for _,id in ipairs({'baseline.font_after','sentinel.read','retry.font','zero_interval.font'}) do assert(diagnostics[id]==state,id..':'..diagnostics[id]) end
 assert(diagnostics['baseline.font_matches']==((state=='public_empty' or state=='public_nonempty') and 'mismatch' or 'not_comparable'))
 local sampleState=(state=='public_empty' or state=='public_nonempty') and 'public' or state
 for _,id in ipairs({'baseline.font','projection.font','manual_step.font','default_clock.font'}) do assert(samples[id].state==sampleState,id) end
 if sampleState~='public' then assert(experiments['projection.matches']=='not_comparable' and experiments['manual_step.matches']=='not_comparable' and experiments['default_clock.matches']=='not_comparable') end
elseif mode=='clock_split' then
 assert(seen['DurationTextBindingObject.UpdateFontString'].state=='property_mismatch' and result.verifiedMethods==72)
 sampleIs('baseline.expected','34372e3520736563',8,0)
 sampleIs('baseline.font','45585049524544',7,0)
 sampleIs('manual_step.expected','343020736563',6,0)
 sampleIs('manual_step.font','45585049524544',7,0)
 assert(experiments['projection.matches']=='matched' and experiments['manual_step.clock_time']=='matched' and experiments['manual_step.remaining']=='matched' and experiments['manual_step.matches']=='mismatch')
 assert(experiments['manual_step.restore_time']=='called_void' and experiments['default_clock.matches']=='matched')
elseif mode=='sample_bytes' then
 sampleIs('baseline.expected','4100c3a9ff',5,2)
 sampleIs('baseline.font','41',1,0)
 sampleIs('projection.font','41',1,0)
 assert(experiments['projection.matches']=='mismatch')
 assert(seen['DurationTextBindingObject.UpdateFontString'].state=='passed','native C-string projection is the expected output')
elseif mode=='sample_empty' or mode=='sample_128' then
 for _,id in ipairs(expectedSamples) do sampleIs(id,mode=='sample_empty' and '' or string.rep('58',128),mode=='sample_empty' and 0 or 128,0) end
elseif mode=='sample_129' then
 for _,id in ipairs(expectedSamples) do if id~='projection.font' then assert(samples[id].state=='oversized',id) end end
 sampleIs('projection.font','50524f42455f53454e54494e454c',14,0)
 assert(diagnostics['baseline.font_matches']=='matched','original4096 comparison must remain intact')
 assert(experiments['projection.set']=='not_run' and experiments['projection.matches']=='not_comparable' and experiments['manual_step.matches']=='not_comparable' and experiments['default_clock.matches']=='not_comparable')
 assert(seen['DurationTextBindingObject.UpdateFontString'].state=='projection_unavailable')
elseif mode=='experiment_error' then
 assert(experiments['manual_step.update']=='call_error' and experiments['manual_step.restore_time']=='called_void')
 assert(experiments['default_clock.matches']=='matched')
elseif mode=='default_time_secret' then
 assert(experiments['default_clock.time']=='not_comparable' and samples['default_clock.expected'].state=='not_run' and samples['default_clock.font'].state=='not_run')
 for _,id in ipairs({'default_clock.clear','default_clock.span','default_clock.clock_time','default_clock.remaining','default_clock.update','default_clock.matches'}) do assert(experiments[id]=='not_run',id) end
elseif mode=='native_plural' then
 sampleIs('baseline.expected','3438207c345365636f6e643a5365636f6e64733b',20,0)
 sampleIs('baseline.font','3438205365636f6e6473',10,0)
 sampleIs('projection.font','3438205365636f6e6473',10,0)
 assert(diagnostics['baseline.font_matches']=='mismatch' and experiments['projection.matches']=='mismatch')
 assert(seen['DurationTextBindingObject.UpdateFontString'].state=='passed' and result.verifiedMethods==73)
elseif mode=='projection_noop' or mode=='projection_noop_stale' or mode=='projection_error' then
 assert(diagnostics['sentinel.matches']=='matched' and seen['DurationTextBindingObject.UpdateFontString'].state=='projection_unavailable')
 sampleIs('projection.font','50524f42455f53454e54494e454c',14,0)
 if mode=='projection_noop_stale' then assert(samples['baseline.font'].hex==samples['projection.font'].hex,'even equal stale sentinel must never validate itself') end
elseif mode=='update_tuple' then
 local row=seen['DurationTextBindingObject.UpdateFontString']
 assert(row.state=='property_mismatch' and row.invoked and row.returnCount==1 and not row.behaviorVerified)
elseif mode=='update_error' then
 local row=seen['DurationTextBindingObject.UpdateFontString']
 assert(row.state=='call_error' and not row.invoked and not row.behaviorVerified)
end
local primaryUnaffected=mode=='normal' or mode=='diag_setup_error' or mode=='diag_setter_tuple' or mode:match('^diag_fact_')~=nil or mode=='sample_128' or mode=='sample_bytes' or mode=='experiment_error' or mode=='default_time_secret' or mode=='native_plural'
assert(result.invocationCaptureComplete==(primaryUnaffected==true),'diagnostics must not change primary outcome')
SlashCmdList.WOWAPIPROBE('invoke')
assert(counts.frame==1 and counts.font==1 and counts.texture==1 and counts.group==1 and counts.radial==1 and counts.vector==1,'UI ownership must remain bounded across captures')
assert(candidateCalls==0,'candidate lookup must not invoke otherwise callable methods')
for _,v in ipairs(bindings) do assert(v.enabled==false) end
for _,v in ipairs(maps) do assert(next(v.signals)==nil) end
