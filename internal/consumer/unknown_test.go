package consumer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestUnknownCompilerContracts(t *testing.T) {
	root := repositoryRoot(t)
	command := exec.CommandContext(t.Context(), "luajit", filepath.Join(root, "tools/teal/tests/compiler.lua"))
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("compiler contracts: %v\n%s", err, output)
	}
}

func unknownFixture(t *testing.T) (string, string) {
	t.Helper()
	root := repositoryRoot(t)
	fixture := t.TempDir()
	for _, name := range []string{"tools/check_secrets.lua", "tools/teal/contract.lua"} {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		writeTest(t, filepath.Join(fixture, name), string(data))
	}
	writeTest(t, filepath.Join(fixture, "types/wow/native_values.d.tl"), `local record Native
 record OpaqueNumber
  userdata
 end
 record OpaqueText
  userdata
 end
end
return Native
`)
	writeTest(t, filepath.Join(fixture, "types/wow/wow.d.tl"), `local type N = require("native_values")
local record W
 record OpaqueValue
  userdata
 end
 type NumberLike = number | N.OpaqueNumber
end
return W
`)
	declarations := []struct{ name, signature, policy string }{
		{"canaccessvalue", "function(unknown):boolean", `guard="value",secret_parameters={},secret_contents_parameters={}`},
		{"canaccessallvalues", "function(unknown...):boolean", `guard="values",secret_parameters={},secret_contents_parameters={}`},
		{"canaccesstable", "function(unknown):boolean", `guard="table",secret_parameters={},secret_contents_parameters={}`},
		{"issecretvalue", "function(unknown):boolean", `guard="secret",secret_parameters={},secret_contents_parameters={}`},
		{"source", "function():unknown", `secret_parameters={},secret_contents_parameters={}`},
		{"deny", "function(unknown)", `secret_parameters={},secret_contents_parameters={}`},
		{"allow", "function(unknown)", `secret_parameters={1},secret_contents_parameters={1}`},
		{"rootOnly", "function(unknown)", `secret_parameters={1},secret_contents_parameters={}`},
		{"nested", "function({unknown})", `secret_parameters={},secret_contents_parameters={}`},
		{"allowContent", "function({unknown})", `secret_parameters={},secret_contents_parameters={1}`},
		{"tuple", "function():number,unknown", `secret_parameters={},secret_contents_parameters={}`},
		{"consume", "function(number,unknown)", `secret_parameters={},secret_contents_parameters={}`},
	}
	var env, policy strings.Builder
	env.WriteString("local type W = require(\"wow\")\n")
	policy.WriteString(`return {schema=1,compiler="0.24.8-wow.1",declarations={["types/wow/wow_env.d.tl"]={`)
	for index, declaration := range declarations {
		line := "global " + declaration.name + ": " + declaration.signature
		env.WriteString(line + "\n")
		fmt.Fprintf(&policy, "[%q]={symbol=%q,%s},\n", fmt.Sprintf("%d:%d", index+2, strings.Index(line, "function")+1), declaration.name, declaration.policy)
	}
	for index, name := range []string{"CallableDeny", "CallableAllow"} {
		line := " metamethod __call: function(" + name + ",unknown)"
		fmt.Fprintf(&env, "local record %s\n%s\nend\nglobal callable%s:%s\n", name, line, name, name)
		permissions := "{}"
		if index == 1 {
			permissions = "{2}"
		}
		fmt.Fprintf(&policy, "[%q]={symbol=%q,secret_parameters=%s,secret_contents_parameters=%s},\n", fmt.Sprintf("%d:%d", len(declarations)+3+index*4, strings.Index(line, "function")+1), name, permissions, permissions)
	}
	policy.WriteString("}}}\n")
	writeTest(t, filepath.Join(fixture, "types/wow/wow_env.d.tl"), env.String())
	writeTest(t, filepath.Join(fixture, "types/wow/call_policy.lua"), policy.String())
	return fixture, filepath.Join(root, "tools/teal/tl")
}
func TestUnknownSecrecyContracts(t *testing.T) {
	cases := []struct{ name, source, reject string }{
		{"callable deny", `callableCallableDeny(source())`, "secret-value"},
		{"callable allow", `callableCallableAllow(source())`, ""},
		{"generic type unguarded", `local function kind<T>(v:T):string return type(v) end
local name=kind(source())`, "secret-value"},
		{"generic native sink", `local function forward<T>(v:T) deny(v) end
forward(source())`, "secret-value"},
		{"generic callback sink", `local function forward<T>(v:T,f:function(unknown)) f(v) end`, "secret-value"},
		{"generic identity retains secrecy", `local function id<T>(v:T):T return v end
local v=id(source())
if v is number then print(v+1) end`, "secret-value"},
		{"generic guarded tag", `local function kind<T>(v:T):string
if canaccessvalue(v) and not issecretvalue(v) then return type(v) end
return "inaccessible"
end
local name=kind(source())`, ""},
		{"generic constrained field", `local interface Box value:unknown end
local function inspect<T is Box>(box:T):string return type(box.value) end`, "secret-value"},
		{"tag guarded", `local v=source()
if canaccessvalue(v) and not issecretvalue(v) then local tag=type(v) end`, ""},
		{"tag unguarded", `local tag=type(source())`, "secret-value"},
		{"tag wrong value", `local a,b=source(),source()
if canaccessvalue(a) and not issecretvalue(a) then local tag=type(b) end`, "secret-value"},
		{"equality guarded", `local a,b=source(),source()
if canaccessvalue(a) and not issecretvalue(a) and canaccessvalue(b) and not issecretvalue(b) then local equal=a==b end`, ""},
		{"equality one side", `local a,b=source(),source()
if canaccessvalue(a) and not issecretvalue(a) then local equal=a==b end`, "secret-value"},
		{"equality no public proof", `local a,b=source(),source()
if canaccessvalue(a) and canaccessvalue(b) then local equal=a==b end`, "secret-value"},
		{"field observation", `local record R value:number end
local function observe(r:R):unknown return __wow_observe_field(r,"missing") end`, ""},
		{"field result opaque", `local record R value:number end
local function observe(r:R):string return type(__wow_observe_field(r,"missing")) end`, "secret-value"},
		{"pass through", `local function id(v:unknown):unknown return v end
local result=id(source())`, ""},
		{"helper narrowing", `local function id(v:unknown):unknown return v end
local v=id(source())
if v is number then print(v+1) end`, "secret-value"},
		{"guarded helper narrowing", `local function id(v:unknown):unknown return v end
local v=id(source())
if canaccessvalue(v) and not issecretvalue(v) and v is number then print(v+1) end`, ""},
		{"legacy sink", `local function id(v:unknown):unknown return v end
print(id(source()))`, "unknown value cannot escape"},
		{"native deny", `deny(source())`, "secret-value"},
		{"native allow", `allow(source())`, ""},
		{"native root permission is not contents", `rootOnly(source())`, "call argument contents"},
		{"guarded unknown still has opaque contents", `local v=source()
if canaccessvalue(v) and not issecretvalue(v) then deny(v) end`, "call argument contents"},
		{"native guarded primitive", `local v=source()
if canaccessvalue(v) and not issecretvalue(v) and v is number then deny(v) end`, ""},
		{"native alias deny", `local sink=deny
sink(source())`, "secret-value"},
		{"native alias allow", `local sink=allow
sink(source())`, ""},
		{"callback deny", `local function f(sink:function(unknown),v:unknown) sink(v) end`, "secret-value"},
		{"callback guarded", `local function f(sink:function(unknown),v:unknown)
if canaccessvalue(v) and not issecretvalue(v) and v is number then sink(v) end
end`, ""},
		{"fresh public literal", `nested({1,2})`, ""},
		{"fresh opaque literal", `nested({source()})`, "call argument contents"},
		{"fresh nested public literal", `nested({{1,2}})`, ""},
		{"fresh nested opaque literal", `nested({{source()}})`, "call argument contents"},
		{"container deny", `local values:{unknown}={source()}
nested(values)`, "call argument contents"},
		{"container allow", `local values:{unknown}={source()}
allowContent(values)`, ""},
		{"field narrowing", `local record R value:unknown end
local function f(r:R)
local v=r.value
if v is number then print(v+1) end
end`, "secret-value"},
		{"later tuple result", `local n,v=tuple()
if v is number then print(n+v) end`, "secret-value"},
		{"later tuple argument", `consume(tuple())`, "secret-value"},
		{"alias guard", `local guard=canaccessvalue
local public=issecretvalue
local v=source()
if guard(v) and not public(v) and v is number then print(v+1) end`, ""},
		{"shadow guard", `local function canaccessvalue(_:unknown):boolean return true end
local v=source()
if canaccessvalue(v) and not issecretvalue(v) and v is number then print(v+1) end`, "secret-value"},
		{"wrong value", `local a,b=source(),source()
if canaccessvalue(a) and not issecretvalue(a) and b is number then print(b+1) end`, "secret-value"},
		{"reassignment", `local v=source()
if canaccessvalue(v) and not issecretvalue(v) then v=source() if v is number then print(v+1) end end`, "secret-value"},
		{"closure mutation", `local v=source()
local function change() v=source() end
if canaccessvalue(v) and not issecretvalue(v) then change() if v is number then print(v+1) end end`, "secret-value"},
		{"loop mutation", `local v=source()
if canaccessvalue(v) and not issecretvalue(v) then
for i=1,2 do if v is number then print(v+1) end v=source() end
end`, "secret-value"},
		{"pcall deny", `local ok=pcall(deny,source())`, "secret-value"},
		{"pcall allow", `local ok=pcall(allow,source())`, ""},
		{"pcall local transport", `local function id(v:unknown):unknown return v end
local ok,v=pcall(id,source())`, ""},
		{"vararg counting", `local function count(...:unknown):integer return select("#",...) end
local n=count(source())`, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fixture, compiler := unknownFixture(t)
			source := filepath.Join(fixture, "caller.tl")
			writeTest(t, source, tc.source)
			command := exec.CommandContext(t.Context(), compiler, "-I", filepath.Join(fixture, "types/wow"), "--gen-target", "5.1", "--gen-compat", "off", "run", filepath.Join(fixture, "tools/check_secrets.lua"), "--", "wow_env", source)
			output, err := command.CombinedOutput()
			if tc.reject == "" {
				if err != nil {
					t.Fatalf("unexpected rejection: %v\n%s", err, output)
				}
			} else if err == nil || !strings.Contains(string(output), tc.reject) || !strings.Contains(string(output), "caller.tl:") {
				t.Fatalf("expected %q rejection, got %v\n%s", tc.reject, err, output)
			}
		})
	}
}

func TestUnknownReportProtocolFailsClosed(t *testing.T) {
	for _, tc := range []struct{ name, patch, want string }{
		{"capability", `tl.wow_capabilities.unknown=0`, "paired Teal"},
		{"schema", `local original=tl.get_types
function tl.get_types(result) local report=original(result) report.wow_schema=nil return report end`, "report schema"},
		{"generic bound", `local original=tl.get_types
function tl.get_types(result)
 local report=original(result)
 for _,info in pairs(report.types) do
  if info.t==tl.typecodes.TYPE_VARIABLE then info.opaque_bound=nil info.constraint=nil end
 end
 return report
end`, "generic bound"},
		{"later tuple member", `local original=tl.get_types
function tl.get_types(result)
 local report=original(result)
 for _,info in pairs(report.types) do
  if info.t==tl.typecodes.FUNCTION and info.file and info.file:match("wow_env.d.tl$") and info.y==12 then
   assert(info.rets and info.rets[2],"test fixture changed")
   report.types[info.rets[2][1]]=nil
   break
  end
 end
 return report
end`, "tuple"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture, compiler := unknownFixture(t)
			source := filepath.Join(fixture, "caller.tl")
			writeTest(t, source, "local function id<T>(value:T):T return value end\nconsume(tuple())\n")
			checker := filepath.Join(fixture, "tools/check_secrets.lua")
			wrapper := filepath.Join(fixture, "protocol.lua")
			writeTest(t, wrapper, "local tl=require(\"tl\")\n"+tc.patch+fmt.Sprintf("\narg[0]=%q\ndofile(arg[0])\n", checker))
			command := exec.CommandContext(t.Context(), compiler, "-I", filepath.Join(fixture, "types/wow"), "run", wrapper, "--", "wow_env", source)
			output, err := command.CombinedOutput()
			if err == nil || !strings.Contains(string(output), tc.want) {
				t.Fatalf("missing fail-closed protocol rejection %q: %v\n%s", tc.want, err, output)
			}
		})
	}
}
