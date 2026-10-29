package consumer

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

type nativeDeclaration struct{ name, signature, policy string }

// nativeDeclarations is a small native environment bound by a call policy. The
// policy keys each declaration by its function-type location, as generated.
var nativeDeclarations = []nativeDeclaration{
	{"canaccessvalue", "function(any): boolean", `guard="value",secret_parameters={}`},
	{"canaccessallvalues", "function(any...): boolean", `guard="values",secret_parameters={}`},
	{"canaccesstable", "function(any): boolean", `guard="table",secret_parameters={}`},
	{"issecretvalue", "function(any): boolean", `guard="secret",secret_parameters={}`},
	{"source", "function(): W.MaybeSecret<number>", `secret_parameters={}`},
	{"tableSource", "function(): W.Secret<W.SecretTable>", `secret_parameters={}`},
	{"deny", "function(any)", `secret_parameters={}`},
	{"allow", "function(any)", `secret_parameters={1}`},
	{"allowRest", "function(number, any...)", `secret_parameters={},secret_vararg_from=2`},
	{"nested", "function({any})", `secret_parameters={}`},
	{"tuple", "function(): number, W.MaybeSecret<number>", `secret_parameters={}`},
	{"consume", "function(number, any)", `secret_parameters={}`},
	{"onEach", "function<T>(function(T))", `secret_parameters={}`},
	{"onUnion", "function<T>(function(T) | string)", `secret_parameters={}`},
	{"withValue", "function<T>(T, function(T))", `secret_parameters={}`},
	{"withUnion", "function<T>(T, function(T) | string)", `secret_parameters={}`},
	{"withEither", "function<T>(T | function(T))", `secret_parameters={}`},
	{"withTwo", "function<A, B>(A, B, function(A, B))", `secret_parameters={}`},
	{"onEachArr", "function<T>(function({T}))", `secret_parameters={}`},
	{"loop", "function<T>(function(T): T)", `secret_parameters={}`},
	{"withOptional", "function<T>(function(T), ? T)", `secret_parameters={}`},
	{"withOptional2", "function<A, B>(function(A, B), ? A)", `secret_parameters={}`},
	{"withOptionalPair", "function<A, B>(function(A, B), ? A, ? B)", `secret_parameters={}`},
}

// nativeLocation is the policy key of a declaration: line 1 is the wow import.
func nativeLocation(index int) string {
	declaration := nativeDeclarations[index]
	line := "global " + declaration.name + ": " + declaration.signature
	return fmt.Sprintf("%d:%d", index+2, strings.Index(line, "function")+1)
}

func nativeIndex(t *testing.T, name string) int {
	t.Helper()
	for index, declaration := range nativeDeclarations {
		if declaration.name == name {
			return index
		}
	}
	t.Fatalf("no native declaration %s", name)
	return 0
}

// nativeFixture copies the real checker and marker declarations next to a
// synthetic environment. edit may rewrite any fixture file before use.
func nativeFixture(t *testing.T, edit func(files map[string]string)) string {
	t.Helper()
	root := repositoryRoot(t)
	fixture := t.TempDir()
	files := map[string]string{
		"types/wow/wow.d.tl": `local type N = require("native_values")
local record W
   type Data = N.Data
   type Atom = N.Atom
   type SecretData = N.SecretData
   type SecretTable = N.SecretTable
   type MaybeSecret<T> = N.MaybeSecret<T>
   type Secret<T> = N.Secret<T>
end
return W
`,
	}
	for _, name := range []string{"tools/check_secrets.lua", "types/wow/native_values.d.tl"} {
		files[name] = fileText(t, filepath.Join(root, name))
	}
	var env, policy strings.Builder
	env.WriteString("local type W = require(\"wow\")\n")
	fmt.Fprintf(&policy, "return {schema=1,compiler=%q,declarations={[\"types/wow/wow_env.d.tl\"]={\n", TealVersion)
	for index, declaration := range nativeDeclarations {
		env.WriteString("global " + declaration.name + ": " + declaration.signature + "\n")
		fmt.Fprintf(&policy, "[%q]={symbol=%q,%s},\n", nativeLocation(index), declaration.name, declaration.policy)
	}
	for index, name := range []string{"CallableDeny", "CallableAllow"} {
		line := "   metamethod __call: function(" + name + ", any)"
		fmt.Fprintf(&env, "local record %s\n%s\nend\nglobal callable%s: %s\n", name, line, name, name)
		permissions := "{}"
		if index == 1 {
			permissions = "{2}"
		}
		fmt.Fprintf(&policy, "[%q]={symbol=%q,secret_parameters=%s},\n", fmt.Sprintf("%d:%d", len(nativeDeclarations)+3+index*4, strings.Index(line, "function")+1), name, permissions)
	}
	policy.WriteString("}}}\n")
	files["types/wow/wow_env.d.tl"] = env.String()
	files["types/wow/call_policy.lua"] = policy.String()
	if edit != nil {
		edit(files)
	}
	for name, text := range files {
		writeTest(t, filepath.Join(fixture, filepath.FromSlash(name)), text)
	}
	return fixture
}

// checkSource is a source file of one checker run.
type checkSource struct{ name, text string }

// caller is a fixture source that imports the wow declarations.
func caller(source string) checkSource {
	return checkSource{"caller.tl", "local type W = require(\"wow\")\n" + source + "\n"}
}

// execNativeCheck runs the fixture checker over sources, optionally through a
// Lua wrapper that patches the compiler before the checker loads. It returns
// the output and the source names.
func execNativeCheck(t *testing.T, fixture, wrapper string, sources ...checkSource) (string, []string, error) {
	t.Helper()
	names := make([]string, len(sources))
	for index, source := range sources {
		writeTest(t, filepath.Join(fixture, source.name), source.text)
		names[index] = source.name
	}
	checker := filepath.Join(fixture, "tools/check_secrets.lua")
	script := checker
	if wrapper != "" {
		script = filepath.Join(fixture, "protocol.lua")
		// The wrapper patches the compiler instance `tl run` already loaded.
		writeTest(t, script, "local tl = package.loaded[\"tl\"]\n"+wrapper+fmt.Sprintf("\narg[0] = %q\ndofile(arg[0])\n", checker))
	}
	command := exec.CommandContext(t.Context(), teal(t), secretCheckArgs([]string{filepath.Join(fixture, "types/wow")}, script, "wow_env", "wow_env", names)...)
	command.Dir = fixture
	output, err := command.CombinedOutput()
	return string(output), names, err
}

// runNativeCheck is execNativeCheck with the output grouped by source. Output
// that names no source fails the test.
func runNativeCheck(t *testing.T, fixture, wrapper string, sources ...checkSource) (string, map[string][]diagnostic, error) {
	t.Helper()
	output, names, err := execNativeCheck(t, fixture, wrapper, sources...)
	diagnostics, parseErr := parseDiagnostics(output, names)
	if parseErr != nil {
		t.Fatalf("%v\n%s", parseErr, output)
	}
	return output, diagnostics, err
}

func TestNativePolicyContracts(t *testing.T) {
	cases := []struct{ name, source, reject string }{
		{"callable deny", `callableCallableDeny(source())`, "secret-value flow"},
		{"callable allow", `callableCallableAllow(source())`, ""},
		{"native deny", `deny(source())`, "secret-value flow"},
		{"native allow", `allow(source())`, ""},
		{"native vararg allow", `allowRest(1, source(), source())`, ""},
		{"native vararg before range", `allowRest(source() as number, source())`, "secret-value"},
		{"native alias deny", `local sink = deny
sink(source())`, "secret-value flow"},
		{"native alias allow", `local sink = allow
sink(source())`, ""},
		{"policy bound to another location does not allow", `allow(source())`, "secret-value flow"},
		{"generic type variable is a sink", `local function kind<T>(v: T): string return type(v) end
local name = kind(source())`, "secret-value flow"},
		{"generic native sink", `local function forward<T>(v: T) deny(v) end
forward(source())`, "secret-value flow"},
		{"generic identity", `local function id<T>(v: T): T return v end
local v = id(source())`, "secret-value flow"},
		{"generic maybe secret guarded tag", `local function kind<T>(v: W.MaybeSecret<T>): string
   if canaccessvalue(v) and not issecretvalue(v) then return type(v) end
   return "inaccessible"
end
local name = kind(source())`, ""},
		{"tag guarded", `local v = source()
if canaccessvalue(v) and not issecretvalue(v) then local tag = type(v) end`, ""},
		{"tag unguarded", `local tag = type(source())`, "secret-value"},
		{"tag wrong value", `local a, b = source(), source()
if canaccessvalue(a) and not issecretvalue(a) then local tag = type(b) end`, "secret-value"},
		{"equality guarded", `local a, b = source(), source()
if canaccessvalue(a) and not issecretvalue(a) and canaccessvalue(b) and not issecretvalue(b) then local equal = a == b end`, ""},
		{"equality one side", `local a, b = source(), source()
if canaccessvalue(a) and not issecretvalue(a) then local equal = a == b end`, "secret-value comparison"},
		{"equality needs access only", `local a, b = source(), source()
if canaccessvalue(a) and canaccessvalue(b) then local equal = a == b end`, ""},
		{"any pass through is a sink", `local function id(v: any): any return v end
local result = id(source())`, "secret-value flow"},
		{"guarded value is still secret typed", `local v = source()
if canaccessvalue(v) and not issecretvalue(v) then deny(v) end`, "secret-value flow"},
		{"guarded unwrap reaches native", `local v = source()
if canaccessvalue(v) and not issecretvalue(v) then deny(v as number) end`, ""},
		{"fresh public literal", `nested({1, 2})`, ""},
		{"fresh secret literal", `nested({source()})`, "secret-value flow"},
		{"fresh nested secret literal", `local function f(n: function({{any}})) n({{source()}}) end`, "secret-value flow"},
		{"secret container into any container", `local values: {W.MaybeSecret<number>} = {source()}
nested(values)`, "secret-value flow"},
		{"callback deny", `local function f(sink: function(any), v: W.MaybeSecret<number>) sink(v) end`, "secret-value flow"},
		{"callback guarded unwrap", `local function f(sink: function(any), v: W.MaybeSecret<number>)
   if canaccessvalue(v) and not issecretvalue(v) then sink(v as number) end
end`, ""},
		// A native chooses a type variable no argument binds, so what it hands a
		// callback may not reach a sink.
		{"unbound native type variable to literal callback", `onEach(function(x: any) end)`, "secret-value flow: generic transport: type variable reaches any"},
		{"unbound native type variable to named callback", `local function h(x: any) end
onEach(h)`, "secret-value flow: generic transport: type variable reaches any"},
		{"unbound native type variable to typed callback", `onEach(function(x: number) end)`, ""},
		// A callback member of a declared union is compared like a callback.
		{"native union callback member compared", `onUnion(function(x: any) end)`, "secret-value flow: argument 1: callback parameter 1"},
		{"native union named callback member compared", `local function h(x: any) end
onUnion(h)`, "secret-value flow: argument 1: callback parameter 1"},
		{"native union typed callback member", `onUnion(function(x: number) end)`, ""},
		{"native union callback position takes its plain member", `onUnion("s")`, ""},
		{"native type-variable-or-callback union", `withEither(function(x: any) end)`, "secret-value flow: argument 1: callback parameter 1"},
		// A bound native type variable transports the bound value.
		{"native transport of secret to callback", `withValue(source(), function(x: any) end)`, "secret-value flow: generic transport: Secret reaches any"},
		{"native transport of plain value to callback", `withValue(1, function(x: any) end)`, ""},
		{"native transport beside nil callback", `withValue(source(), nil)`, ""},
		// A vararg callback parameter receives every later value.
		{"native transport of secret to vararg callback", `withTwo(1, source(), function(...: any) end)`, "secret-value flow: generic transport: Secret reaches any"},
		{"native transport of plain values to vararg callback", `withTwo(1, 2, function(...: any) end)`, ""},
		{"native transport of secret past callback arity", `withTwo(1, source(), function(a: any) end)`, ""},
		// An omitted argument is nil, so it binds its type variable as nil does;
		// a last argument with variadic results binds every later position to
		// its own results.
		{"omitted optional binds native type variable to nil", `withOptional(function(x: any) end)`, ""},
		{"supplied plain optional", `withOptional(function(x: any) end, 1)`, ""},
		{"supplied secret optional reaches callback", `withOptional(function(x: any) end, source())`, "secret-value flow: generic transport: Secret reaches any"},
		{"omission binds only its own type variable", `withOptional2(function(a: number, b: any) end)`, "secret-value flow: generic transport: type variable reaches any"},
		{"variadic last argument fills an omitted position", `local function g(): W.MaybeSecret<number>... return source(), source() end
withOptionalPair(function(a: W.MaybeSecret<number>, b: any) end, g())`, "secret-value flow: generic transport: Secret reaches any"},
		{"vararg expression as last argument fills an omitted position", `local function f(...: W.MaybeSecret<number>) withOptionalPair(function(a: W.MaybeSecret<number>, b: any) end, ...) end`, "secret-value flow: generic transport: Secret reaches any"},
		{"fixed-result last argument leaves later positions omitted", `local function g(): W.MaybeSecret<number> return source() end
withOptionalPair(function(a: W.MaybeSecret<number>, b: any) end, g())`, ""},
		// What a callback returns or receives in a container does not bind the
		// type variable: the native still chooses what it passes the callback.
		{"callback container does not bind native type variable", `onEachArr(function(t: {any}) end)`, "secret-value flow: generic transport: type variable reaches any"},
		{"typed callback container", `onEachArr(function(t: {number}) end)`, ""},
		{"callback result does not bind native type variable", `loop(function(x: any): any return x end)`, "secret-value flow: generic transport: type variable reaches any"},
		{"typed callback result", `loop(function(x: number): number return x end)`, ""},
		// A callback result flows into the type variable: it is checked as a
		// source and is not a destination of the callee's choice.
		{"callback result is not a type variable destination", `loop(function(x: number): any return x end)`, ""},
		{"callback result is checked as a type variable source", `loop(function(x: number): W.MaybeSecret<number> return source() end)`, "secret-value flow: generic transport: Secret reaches"},
		{"unbound type variable reaches a callback parameter beside a plain result", `loop(function(x: any): number return 1 end)`, "secret-value flow: generic transport: type variable reaches any"},
		// The union callback rejects even a plain value only because the checker
		// does not model that shape, so withUnion(1, ...) is not pinned.
		{"native transport of secret to union callback", `withUnion(source(), function(x: any) end)`, "secret-value flow: argument 2: callback parameter 1"},
		{"native transport of secret to named union callback", `local function h(x: any) end
withUnion(source(), h)`, "secret-value flow: argument 2: callback parameter 1"},
		{"native transport of secret to typed union callback", `withUnion(source(), function(x: number) end)`, "secret-value flow"},
		{"field narrowing", `local record R value: W.MaybeSecret<number> end
local function f(r: R)
   local v = r.value
   if v is number then print(v + 1) end
end`, "secret-value"},
		{"later tuple result", `local n, v = tuple()
if v is number then print(n + v) end`, "secret-value"},
		{"later tuple argument", `consume(tuple())`, "secret-value flow: argument 2"},
		{"alias guard", `local guard = canaccessvalue
local public = issecretvalue
local v = source()
if guard(v) and not public(v) then print((v as number) + 1) end`, ""},
		{"shadow guard", `local function canaccessvalue(_: any): boolean return true end
local v = source()
if canaccessvalue(v) and not issecretvalue(v) then print((v as number) + 1) end`, "secret-value"},
		{"wrong value", `local a, b = source(), source()
if canaccessvalue(a) and not issecretvalue(a) then print((b as number) + 1) end`, "secret-value"},
		{"reassignment", `local v = source()
if canaccessvalue(v) and not issecretvalue(v) then v = source() print((v as number) + 1) end`, "secret-value"},
		{"closure mutation", `local v = source()
local function change() v = source() end
if canaccessvalue(v) and not issecretvalue(v) then change() print((v as number) + 1) end`, "secret-value"},
		{"loop mutation", `local v = source()
if canaccessvalue(v) and not issecretvalue(v) then
   for _ = 1, 2 do print((v as number) + 1) v = source() end
end`, "secret-value"},
		{"pcall deny", `local ok = pcall(deny, source())`, "secret-value flow"},
		{"pcall allow", `local ok = pcall(allow, source())`, ""},
		{"pcall typed transport", `local function id(v: W.MaybeSecret<number>): W.MaybeSecret<number> return v end
local ok, v = pcall(id, source())`, ""},
		{"vararg counting", `local function count(...: W.MaybeSecret<number>): integer return select("#", ...) end
local n = count(source())`, ""},
		{"secret table guarded", `local raw = tableSource()
if canaccesstable(raw) then local t = raw as W.SecretTable local field = t["key"] end`, ""},
		{"secret table unguarded", `local t = tableSource() as W.SecretTable`, "secret-value type assertion"},
		{"lookalike marker is plain", `local record Secret<T> is userdata end
local function f(s: Secret<number>) if s then print(1) end end`, ""},
		{"native marker is opaque", `local function f(s: W.Secret<number>) if s then print(1) end end`, "secret-value truthiness"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var edit func(map[string]string)
			if tc.name == "policy bound to another location does not allow" {
				// Permissions follow the declaration location, not the symbol.
				edit = func(files map[string]string) {
					name, key := "types/wow/call_policy.lua", fmt.Sprintf("[%q]={symbol=\"allow\"", nativeLocation(nativeIndex(t, "allow")))
					if !strings.Contains(files[name], key) {
						t.Fatalf("fixture policy lacks %s", key)
					}
					files[name] = strings.Replace(files[name], key, `["1:1"]={symbol="allow"`, 1)
				}
			}
			output, diagnostics, err := runNativeCheck(t, nativeFixture(t, edit), "", caller(tc.source))
			if tc.reject == "" {
				if err != nil {
					t.Fatalf("unexpected rejection: %v\n%s", err, output)
				}
			} else if err == nil || !slices.ContainsFunc(diagnostics["caller.tl"], func(d diagnostic) bool {
				return d.failure() && strings.HasPrefix(d.message, tc.reject)
			}) {
				t.Fatalf("expected %q rejection at caller.tl, got %v\n%s", tc.reject, err, output)
			}
		})
	}
}

// The checker reads compiler internals; any drift in the compiler, the marker
// declarations, the stdlib identities or the policy fails closed.
func TestNativePolicyProtocolFailsClosed(t *testing.T) {
	const source = `deny(source())`
	type protocolCase struct {
		name, wrapper, want string
		edit                func(t *testing.T, files map[string]string)
	}
	var cases []protocolCase
	// Opacity follows the native_values declarations Teal loaded. Renaming a
	// marker consistently still compiles, so only the checker can reject it.
	for _, marker := range []string{"Secret", "MaybeSecret", "Data", "Atom", "SecretData", "SecretTable"} {
		cases = append(cases, protocolCase{name: "missing marker " + marker, want: "opaque marker definition changed: " + marker, edit: func(t *testing.T, files map[string]string) {
			const values, wow = "types/wow/native_values.d.tl", "types/wow/wow.d.tl"
			renamed := regexp.MustCompile(`\b`+marker+`\b`).ReplaceAllString(files[values], marker+"Renamed")
			if renamed == files[values] {
				t.Fatalf("fixture lacks %s", marker)
			}
			files[values] = renamed
			files[wow] = strings.Replace(files[wow], "N."+marker+"\n", "N."+marker+"Renamed\n", 1)
			files[wow] = strings.Replace(files[wow], "N."+marker+"<T>", "N."+marker+"Renamed<T>", 1)
		}})
	}
	cases = append(cases, protocolCase{name: "markers outside native_values", want: "declaration identity mismatch: native_values", edit: func(t *testing.T, files map[string]string) {
		// wow.d.tl declares the same vocabulary itself and never loads native_values.
		const wow = "types/wow/wow.d.tl"
		files[wow] = strings.ReplaceAll(files["types/wow/native_values.d.tl"], "Values", "W")
	}})
	for _, tc := range append(cases, []protocolCase{
		{name: "compiler version", wrapper: `tl.version = function() return "0.24.8-wow.1" end`, want: "unsupported Teal version"},
		{name: "declaration load", wrapper: `tl.new_env = function() return nil, "declarations unavailable" end`, want: "declarations unavailable"},
		{name: "stdlib identity", wrapper: `local new_env = tl.new_env
tl.new_env = function(...)
   local env, err = new_env(...)
   if env then env.globals.pcall = nil end
   return env, err
end`, want: "compiler omitted stdlib identity: pcall"},
		{name: "expression identity", wrapper: `local process = tl.process
tl.process = function(filename, env, ...)
   env.report_types = false
   return process(filename, env, ...)
end`, want: "compiler omitted expression identity"},
		{name: "policy compiler", want: "missing or incompatible call policy", edit: func(t *testing.T, files map[string]string) {
			name := "types/wow/call_policy.lua"
			files[name] = strings.Replace(files[name], fmt.Sprintf("compiler=%q", TealVersion), `compiler="0.24.8-wow.1"`, 1)
		}},
		{name: "contents policy", want: "unsupported secret-contents policy", edit: func(t *testing.T, files map[string]string) {
			name := "types/wow/call_policy.lua"
			files[name] = strings.Replace(files[name], `symbol="allow",secret_parameters={1}`, `symbol="allow",secret_parameters={1},secret_contents_parameters={1}`, 1)
		}},
		{name: "guard symbol", want: "invalid native guard policy", edit: func(t *testing.T, files map[string]string) {
			name := "types/wow/call_policy.lua"
			files[name] = strings.Replace(files[name], `symbol="allow",secret_parameters={1}`, `symbol="allow",guard="value",secret_parameters={1}`, 1)
		}},
	}...) {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var edit func(map[string]string)
			if tc.edit != nil {
				edit = func(files map[string]string) { tc.edit(t, files) }
			}
			// Drift affects every source, so each one reports it.
			output, diagnostics, err := runNativeCheck(t, nativeFixture(t, edit), tc.wrapper, caller(source), checkSource{"other.tl", "print(1)\n"})
			for _, name := range []string{"caller.tl", "other.tl"} {
				if err == nil || !slices.ContainsFunc(diagnostics[name], func(d diagnostic) bool { return strings.Contains(d.message, tc.want) }) {
					t.Fatalf("missing fail-closed rejection %q on %s: %v\n%s", tc.want, name, err, output)
				}
			}
		})
	}
}

// A type error in the shared declarations is reported where Teal found it, and
// fails every source.
func TestDeclarationErrorsFailEverySource(t *testing.T) {
	t.Parallel()
	fixture := nativeFixture(t, func(files map[string]string) {
		files["types/wow/wow_env.d.tl"] += "global broken: nosuchtype\n"
	})
	output, names, err := execNativeCheck(t, fixture, "", caller(`deny(source())`), checkSource{"other.tl", "print(1)\n"})
	declarations := filepath.Join(fixture, "types/wow/wow_env.d.tl")
	diagnostics, parseErr := parseDiagnostics(output, append(names, declarations))
	if parseErr != nil {
		t.Fatalf("%v\n%s", parseErr, output)
	}
	contains := func(text string) func(diagnostic) bool {
		return func(d diagnostic) bool { return strings.Contains(d.message, text) }
	}
	if err == nil || !slices.ContainsFunc(diagnostics[declarations], contains("unknown type nosuchtype")) {
		t.Fatalf("declaration error not reported: %v\n%s", err, output)
	}
	for _, name := range names {
		if !slices.ContainsFunc(diagnostics[name], contains("declarations do not type-check")) {
			t.Fatalf("%s not failed: %v\n%s", name, err, output)
		}
	}
}

// Teal may type a call by a later overload than the first that fits, so
// generated overloads whose arities overlap must agree on the opacity of every
// result. A generated set that disagrees fails every source.
func TestGeneratedOverloadOpacityDisagreementFailsEverySource(t *testing.T) {
	cases := []struct{ name, first, second, want string }{
		{"overlapping arities", "function(self): number", "function(self, ? number): N.Secret<number>", "result 1"},
		{"opaque contents", "function(self): {number}", "function(self, ? number): {N.Secret<number>}", "result 1"},
		{"later result", "function(self): number, number", "function(self, ? number): number, N.Secret<number>", "result 2"},
		{"disjoint arities", "function(self): number", "function(self, number): N.Secret<number>", ""},
		{"agreeing opacity", "function(self): number", "function(self, ? number): number", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fixture := nativeFixture(t, func(files map[string]string) {
				const wow, end = "types/wow/wow.d.tl", "end\nreturn W\n"
				if !strings.HasSuffix(files[wow], end) {
					t.Fatalf("fixture %s does not end with %q", wow, end)
				}
				probe := "   record Probe\n      get: " + tc.first + "\n      get: " + tc.second + "\n   end\n"
				files[wow] = strings.TrimSuffix(files[wow], end) + probe + end
			})
			// The check runs once per checker run; the second source shows a
			// failure does not mark it done.
			output, diagnostics, err := runNativeCheck(t, fixture, "", caller(`print(1)`), checkSource{"other.tl", "print(1)\n"})
			if tc.want == "" {
				if err != nil {
					t.Fatalf("unexpected rejection: %v\n%s", err, output)
				}
				return
			}
			want := "generated overloads of wow.Probe.get disagree on the opacity of " + tc.want
			for _, name := range []string{"caller.tl", "other.tl"} {
				if err == nil || !slices.ContainsFunc(diagnostics[name], func(d diagnostic) bool { return strings.Contains(d.message, want) }) {
					t.Fatalf("missing fail-closed rejection %q on %s: %v\n%s", want, name, err, output)
				}
			}
		})
	}
}
