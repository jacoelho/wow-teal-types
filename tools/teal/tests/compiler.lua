local here = assert(arg[0]:match("^(.*)/tests/compiler.lua$"))
local tl = assert(loadfile(here .. "/tl.lua"))()
assert(tl.version() == "0.24.8-wow.1")
assert(tl.wow_capabilities.unknown == 1 and tl.wow_capabilities.type_report == 1)
assert(tl.typecodes.PUBLIC_UNKNOWN ~= tl.typecodes.UNKNOWN)

local cases = {}
local function accept(name, source) cases[#cases + 1] = {name, source, true} end
local function reject(name, source, diagnostic) cases[#cases + 1] = {name, source, false, diagnostic} end
accept("transport", [[local record R x:number end
local record U is userdata end
local function identity(x:unknown):unknown return x end
local function f(u:U)
 local a=identity(1)
 local b=identity("x")
 local c=identity(false)
 local d=identity(nil)
 local e=identity({1,2})
 local r:R={x=1}
 local g=identity(r)
 local h=identity(function(x:number):number return x end)
 local i=identity(u)
end]])
accept("alias", [[local type Value=unknown
local function f(x:Value):unknown return x end
local function g(x:unknown):Value return x end]])
accept("generic_transport", [[local function id<T>(x:T):T return x end
local function f(x:unknown):unknown return id(x) end
local a:number=id(1)
local b:string=id("x")]])
accept("tuple_transport", [[local function f(x:unknown,y:unknown):unknown,unknown return x,y end
local function g(x:unknown,y:unknown):unknown,unknown return f(x,y) end]])
accept("safe_callback", [[local function broad(x:unknown):number if x is number then return x end return 0 end
local function invoke(f:function(number):unknown):unknown return f(1) end
local result=invoke(broad)]])
accept("same_array_storage", [[local function write(a:{unknown},v:unknown) a[1]=v end
local function f(a:{unknown},v:unknown) write(a,v) end]])
accept("same_map_storage", [[local function write(a:{string:unknown},v:unknown) a.x=v end
local function f(a:{string:unknown},v:unknown) write(a,v) end]])
accept("same_record_storage", [[local record R x:unknown end
local function write(a:R,v:unknown) a.x=v end
local function f(a:R,v:unknown) write(a,v) end]])
accept("arity_count", [[local function count(...:unknown):integer return select("#",...) end]])
accept("numeric_select_opaque", [[local function tail(...:unknown):unknown... return select(2,...) end]])
accept("typed_pcall", [[local function id(x:unknown):unknown return x end
local function f(x:unknown):boolean,unknown return pcall(id,x) end]])
accept("typed_unpack", [[local function f(x:{unknown}):unknown... return table.unpack(x) end]])
accept("legitimate_result_generic", [[global fresh:function<T>():{T}
local values:{number}=fresh()]])
for _, primitive in ipairs({"number","string","boolean","nil"}) do
 accept("narrow_"..primitive, "local function f(x:unknown):boolean if x is "..primitive.." then return true end return false end")
end
for _, target in ipairs({"number","string","boolean","{number}","function(number):number","any"}) do
 reject("conversion_"..target,"local function f(x:unknown):"..target.." return x end","unknown")
 reject("cast_"..target,"local function f(x:unknown):"..target.." return x as "..target.." end","unknown")
end
for _, expression in ipairs({"x+1","1+x","x..'a'","'a'..x","#x","-x","not x","x and true","true and x","x or false","false or x","x<1","x.a","x[1]","x(1)","({1})[x]"}) do
 reject("operation_"..expression,"local function f(x:unknown) local result="..expression.." end","unknown")
end
for _, source in ipairs({"if x then end","while x do break end","repeat break until x","for i=x,1 do end","for i=1,x do end","for i=1,2,x do end","for v in x do end"}) do
 reject("condition_"..source,"local function f(x:unknown) "..source.." end")
end
for _, builtin in ipairs({"print","tostring","tonumber","assert","pairs","ipairs","next"}) do
 reject("builtin_"..builtin,"local function f(x:unknown) "..builtin.."(x) end")
 reject("builtin_alias_"..builtin,"local inspect="..builtin.."\nlocal function f(x:unknown) inspect(x) end")
end
reject("assert_callback", [[local function inspect(f:function(unknown):unknown,x:unknown):unknown return f(x) end
local function f(x:unknown):unknown return inspect(assert,x) end]])
reject("legacy_callback", [[local function inspect(f:function(unknown),x:unknown) f(x) end
local function f(x:unknown) inspect(print,x) end]])
reject("legacy_container", [[local function f(x:{unknown}) print(x) end]],"unknown")
reject("legacy_container_cast", [[local function f(x:{unknown}):{number} return x as {number} end]],"unknown")
reject("record_storage", [[local record A x:number end
local record B x:unknown end
local function write(b:B) b.x="wrong" end
local function f(a:A) write(a) end]])
reject("array_storage", [[local function write(a:{unknown}) a[1]="wrong" end
local function f(a:{number}) write(a) end]])
reject("map_value_storage", [[local function write(a:{string:unknown}) a.x="wrong" end
local function f(a:{string:number}) write(a) end]])
reject("map_key_storage", [[local function write(a:{unknown:string}) a["wrong"]="wrong" end
local function f(a:{number:string}) write(a) end]])
reject("callback_parameter", [[local function inspect(f:function(unknown):unknown):unknown return f("wrong") end
local function narrow(x:number):number return x+1 end
local v=inspect(narrow)]])
reject("callback_result", [[local function broad(x:number):unknown return "wrong" end
local function inspect(f:function(number):number):number return f(1) end
local v=inspect(broad)]])
reject("callback_variadic_parameter", [[local function f(x:number,...:number):number return x end
local g:function(number,unknown):number=f]])
reject("callback_variadic_result", [[global f:function():unknown...
local g:function():number,number=f]])
reject("generic_result", [[local function id<T>(x:T):T return x end
local function f(x:unknown):number return id(x) end]],"unknown")
reject("tuple_second_result", [[global f:function():number,unknown
local a,b:number,number=f()]],"unknown")
reject("tuple_second_argument", [[global f:function():number,unknown
local function consume(a:number,b:number) end
consume(f())]],"unknown")
reject("negative_narrow", [[local function f(x:unknown):string if not (x is number) then return x end return "" end]],"unknown")
reject("union_unknown", [[local function f(x:unknown|number):number return x end]],"unknown")
for _, target in ipairs({"integer","table","function","R"}) do
 reject("structural_narrow_"..target,"local record R x:number end\nlocal function f(x:unknown) if x is "..target.." then end end","unknown")
end
for _, order in ipairs({false,true}) do
 local first="f: function(number):number"
 local second="f: function(unknown):unknown"
 if order then first,second=second,first end
 reject("overload_unknown_"..tostring(order),"local record R\n"..first.."\n"..second.."\nend\nlocal function use(r:R,x:unknown):number return r.f(x) end")
 accept("overload_number_"..tostring(order),"local record R\n"..first.."\n"..second.."\nend\nlocal function use(r:R):unknown return r.f(1) end")
end
accept("generic_fresh_identities", [[local function id<T>(v:T):T return v end
local function f(v:unknown):unknown
 local opaque=id(v)
 local n:number=id(1)
 local s:string=id("x")
 return opaque
end]])
reject("alias_nested_array_cast", [[local type V=unknown
local function f(v:{{V}}):{{number}} return v as {{number}} end]],"unknown")
reject("method_storage", [[local record Narrow value:number end
local record Wide value:unknown end
local record Writer set:function(Writer,Wide) end
local function f(w:Writer,n:Narrow) w:set(n) end]])
reject("interface_storage", [[local interface Wide value:unknown end
local record Narrow value:number end
local function write(w:Wide) w.value="wrong" end
local function f(n:Narrow) write(n) end]])
reject("generic_storage", [[local record Box<T> value:T end
local function write(w:Box<unknown>) w.value="wrong" end
local function f(n:Box<number>) write(n) end]])
reject("generic_callback", [[local function identity<T>(f:T):T return f end
local function narrow(x:number):number return x+1 end
local wide:function(unknown):unknown=identity(narrow)]])
for _,order in ipairs({false,true}) do
 local a="f:function<T>(T,number):T"
 local b="f:function<T>(number,T):T"
 if order then a,b=b,a end
 accept("overload_inference_rollback_"..tostring(order),"local record R\n"..a.."\n"..b.."\nend\nlocal function f(r:R,x:unknown):unknown return r.f(1,x) end")
 reject("overload_inference_retains_unknown_"..tostring(order),"local record R\n"..a.."\n"..b.."\nend\nlocal function f(r:R,x:unknown):number return r.f(1,x) end")
end
reject("callable_record_parameter", [[local record F
 metamethod __call:function(F,number):number
end
local function f(c:F,v:unknown):number return c(v) end]],"unknown")
reject("callable_record_result", [[local record F
 metamethod __call:function(F):unknown
end
local function f(c:F):number return c() end]],"unknown")
accept("callable_record_transport", [[local record F
 metamethod __call:function(F,unknown):unknown
end
local function f(c:F,v:unknown):unknown return c(v) end]])
for _,body in ipairs({"print(x)","assert(x)","if x then end","while x do break end","local n=x as number","local r=x as R","local f=x as function():number","local n=tonumber(x)","local s=tostring(x)","local t:{number}=x as {number}"}) do
 reject("generic_body_"..body,"local record R x:number end\nlocal function f<T>(x:T) "..body.." end")
end
reject("generic_nested_legacy", [[local function f<T>(x:{T}) print(x) end]])
reject("generic_callback_legacy", [[local function f<T>(x:function():T) print(x) end]])
accept("constrained_interface", [[local interface Named name:string end
local function get<T is Named>(x:T):string return x.name end]])
accept("tag_only", [[local function f(x:unknown):string return type(x) end]])
accept("equality_observation", [[local function f(x:unknown,y:unknown):boolean return x==y end]])
reject("tag_does_not_narrow", [[local record R x:number end
local function f(v:unknown):R if type(v)=="table" then return v as R end return {} as R end]],"unknown")
accept("observe_record", [[local record R value:number end
local function f(r:R):unknown return __wow_observe_field(r,"other") end]])
accept("observe_userdata", [[local record R userdata end
local function f(r:R):unknown return __wow_observe_field(r,"other") end]])
accept("observe_map", [[local function f(r:{string:number},k:string):unknown return __wow_observe_field(r,k) end]])
reject("observe_unknown", [[local function f(r:unknown):unknown return __wow_observe_field(r,"x") end]],"unknown receiver")
reject("observe_generic", [[local function f<T>(r:T):unknown return __wow_observe_field(r,"x") end]],"unknown receiver")
reject("observe_function", [[local function f(r:function()):unknown return __wow_observe_field(r,"x") end]],"statically known")
reject("observe_alias", [[local f=__wow_observe_field]],"direct call")
reject("observe_cannot_call", [[local record R value:number end
local function f(r:R) __wow_observe_field(r,"x")() end]],"unknown")
reject("observe_shadow", [[local function __wow_observe_field(x:number,y:string):number return x end
local function f(r:unknown):number return __wow_observe_field(r,"x") end]],"unknown")
for _,order in ipairs({false,true}) do
 local precise="create:function(self,Choice):Child"
 local broad="create:function(self,string):Base"
 if order then precise,broad=broad,precise end
 local declarations="local enum Choice \"child\" end\nlocal record Base base:number end\nlocal record Child base:number child:number end\nlocal record Factory\n"..precise.."\n"..broad.."\nend\n"
 accept("aliased_method_overload_"..tostring(order),declarations.."local function f(owner:Factory):Child local method=owner.create return method(owner,\"child\") end")
 reject("aliased_method_wrong_receiver_"..tostring(order),declarations.."local function f(owner:Factory,wrong:Base):Child local method=owner.create return method(wrong,\"child\") end")
 reject("aliased_method_wrong_literal_"..tostring(order),declarations.."local function f(owner:Factory):Child local method=owner.create return method(owner,\"other\") end")
end
for _,kind in ipairs({"record","interface"}) do
 reject("inherited_storage_"..kind,"local interface Wide value:unknown end\nlocal "..kind.." Narrow is Wide value:number end\nlocal function write(w:Wide) w.value=\"bad\" end\nlocal function f(n:Narrow):number local alias:Wide=n write(alias) return n.value+1 end", "does not match definition")
 for _,field in ipairs({"number","unknown"}) do
  accept("same_inherited_field_"..kind..field,"local interface Base value:"..field.." read:function(self):number end\nlocal "..kind.." Derived is Base value:"..field.." read:function(self):number end\nlocal function f(n:Derived):number local base:Base=n return base:read() end")
 end
end
reject("generic_inherited_storage", [[local interface Box<T> value:T end
local record Narrow is Box<unknown> value:number end
local function write(w:Box<unknown>) w.value="bad" end
local function f(n:Narrow):number local alias:Box<unknown> = n write(alias) return n.value+1 end]],"does not match definition")
reject("transitive_inherited_storage", [[local interface Wide value:unknown end
local interface Middle is Wide end
local record Narrow is Middle value:number end
local function write(w:Wide) w.value="bad" end
local function f(n:Narrow):number write(n) return n.value+1 end]],"does not match definition")
reject("writable_method_override", [[local interface Wide run:function(self,number):unknown end
local record Narrow is Wide run:function(self,unknown):number end
local function write(w:Wide) w.run=function(_:Wide,_:number):unknown return "bad" end end
local function f(n:Narrow):number write(n) return n:run("ok")+1 end]],"does not match definition")
reject("method_receiver_erasure", [[local record Narrow value:number read:function(self):number end
local function erase(n:Narrow):function(unknown):number return n.read end
local function f(n:Narrow,x:unknown):number local call=erase(n) return call(x) end]])
accept("method_receiver_preserved", [[local record Narrow value:number read:function(self):number end
local function extract(n:Narrow):function(Narrow):number return n.read end
local function f(n:Narrow):number local call=extract(n) return call(n) end]])
for _,pair in ipairs({{"function(number,?number):number","function(number,number):number"},{"function(number...):number","function(number):number"},{"function():number...","function():number"}}) do
 reject("function_storage_"..pair[1],"local interface Wide callback:"..pair[1].." end\nlocal record Narrow is Wide callback:"..pair[2].." end", "does not match definition")
end
reject("generic_wrapper_record_cast", [[local record R f:function<T>(T):unknown end
local record N f:function<T>(T):number end
local function cast(r:R):N return r as N end
local function test(r:R):number return cast(r).f("bad")+1 end]],"unknown")
reject("generic_wrapper_map_cast", [[local function cast(r:{string:function<T>(T):unknown}):{string:function<T>(T):number} return r as {string:function<T>(T):number} end]],"unknown")
reject("generic_wrapper_map_legacy", [[local function f(r:{string:function<T>(T):unknown}):any return r end]],"unknown")
reject("generic_wrapper_constraint", [[local interface Box value:unknown end
local function f(r:{string:function<T is Box>(T):number}):any return r end]],"unknown")
reject("generic_alias_argument", [[local record Box<T> value:T end
local type Hidden=Box<unknown>
local function f(r:Hidden):any return r end]],"unknown")
accept("generic_wrapper_record_transport", [[local record R f:function<T>(T):unknown end
local function id(r:R):R return r as R end]])
accept("generic_wrapper_map_transport", [[local function id(r:{string:function<T>(T):unknown}):{string:function<T>(T):unknown} return r end]])
accept("generic_wrapper_fresh_calls", [[local record R f:function<T>(T):T end
local function use(r:R):number,string local n:number=r.f(1) local s:string=r.f("x") return n,s end]])
local failures=0
for _,case in ipairs(cases) do
 local env=assert(tl.new_env({defaults={feat_arity="on",feat_lax="off",gen_target="5.1",gen_compat="off"}}))
 local ok,result=pcall(tl.check_string,case[2],env,case[1]..".tl")
 local diagnostics={}
 if ok then
  for _,e in ipairs(result.type_errors or {}) do diagnostics[#diagnostics+1]=e.msg;assert(e.y and e.x,"missing diagnostic location") end
 end
 local syntax=ok and #(result.syntax_errors or {})>0
 local valid=ok and not syntax and (#diagnostics==0)==case[3]
 if valid and not case[3] and case[4] then valid=table.concat(diagnostics,"\n"):find(case[4],1,true)~=nil end
 if not valid then
  failures=failures+1
  io.stderr:write(case[1],": ",not ok and tostring(result) or syntax and "syntax failure" or table.concat(diagnostics," | "),"\n")
 end
end
local runtime=[[local function classify(x:unknown):string
 if x is number then return "n"..tostring(x*2) end
 if x is string then return "s"..x end
 if x is boolean then if x then return "true" end return "false" end
 if x is nil then return "nil" end
 return "opaque"
end
assert(classify(3)=="n6")
assert(classify("hello")=="shello")
assert(classify(false)=="false")
assert(classify(nil)=="nil")
assert(classify({1})=="opaque")]]
runtime=runtime..[[
local interface Parent value:number read:function(self):number end
local record Child is Parent value:number read:function(self):number end
local child:Child={value=7,read=function(self:Child):number return self.value end}
local parent:Parent=child
assert(parent:read()==7)
parent.value=9
assert(child:read()==9)
local record R
 value:number
end
local receivers,keys=0,0
global provided:function():R
local object:R=provided()
local function receiver():R receivers=receivers+1 return object end
local function key():string keys=keys+1 return "missing" end
local observed=__wow_observe_field(receiver(),key())
assert(receivers==1 and keys==1)
assert(observed is number)
assert(observed==23)
local function failing():unknown return __wow_observe_field(object,"explode") end
local ok=pcall(failing)
assert(not ok)
]]
local env=assert(tl.new_env({defaults={feat_arity="on",gen_target="5.1",gen_compat="off"}}))
local code,result=tl.gen(runtime,env,{gen_target="5.1"})
for _,err in ipairs(result.syntax_errors or {}) do io.stderr:write(err.msg,"\n") end
for _,err in ipairs(result.type_errors or {}) do io.stderr:write(err.msg,"\n") end
assert(code and #result.syntax_errors==0 and #result.type_errors==0,"runtime positive failed typecheck")
assert(not code:find("math.type",1,true),"generated unavailable Lua5.1 math.type")
local runtime_env=setmetatable({provided=function()
 return setmetatable({value=7},{__index=function(_,key)
  if key=="explode" then error("observed failure") end
  return 23
 end})
end},{__index=_G})
setfenv(assert(loadstring(code)),runtime_env)()
assert(failures==0,tostring(failures).." compiler regression failures")
print(tostring(#cases).." compiler contracts and Lua5.1 execution passed")
