package wowapi

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

func TestFinalizedDeclarationsResolveGeneratedTypesInheritanceAndSecrecy(t *testing.T) {
	c := newCatalogue()
	base := c.record("Base", "fixture", kindRecord)
	base.Fields["count"] = &catalogueField{Native: &nativeParameter{Type: "number"}}
	base.Methods["Get"] = &catalogueFunction{Method: true, Native: &nativeFunction{Returns: []nativeParameter{{Type: "number"}}}}
	child := c.record("Child", "fixture", kindRecord)
	child.Bases = []string{"Base"}
	child.Fields["count"] = &catalogueField{Type: "number"}
	child.Methods["Get"] = &catalogueFunction{Annotation: &annotationFunction{Returns: []annotationReturn{{Type: "string"}}}}
	payload := c.record("Payload", "fixture", kindRecord)
	payload.Fields["count"] = &catalogueField{Native: &nativeParameter{Type: "number"}}
	payload.Fields["visible"] = &catalogueField{Native: &nativeParameter{Type: "boolean", NeverSecret: new(true)}}
	holder := c.record("Holder", "fixture", kindRecord)
	holder.Fields["left"] = &catalogueField{Type: "{value:number}"}
	holder.Fields["right"] = &catalogueField{Type: "{value:number}"}
	holder.Fields["callback"] = &catalogueField{Type: "fun(value:number):string"}
	c.setGlobal("GetPayload", &catalogueFunction{Source: "fixture", Native: &nativeFunction{Returns: []nativeParameter{{Type: "Payload", SecretValue: new(true)}}}})

	projection, err := c.finalize(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	outputs := projection.render()
	declarations := string(outputs["wow.d.tl"])
	for _, expected := range []string{
		"   record Base\n      count: number\n      Get: function(self): number\n   end",
		"   record Child\n      is Base\n   end",
		"   record Payload\n      count: W.MaybeSecret<number>\n      visible: boolean\n   end",
		"   record Holder\n      callback: function(number): (string)\n      left: Shape_bc707ec91b8a\n      right: Shape_bc707ec91b8a\n   end",
		"   record Shape_bc707ec91b8a\n      value: number\n   end",
		"   type Global_GetPayload = function(): Payload",
	} {
		if !strings.Contains(declarations, expected) {
			t.Errorf("missing declaration:\n%s\nactual:\n%s", expected, declarations)
		}
	}
	if strings.Count(declarations, "record Shape_") != 1 {
		t.Error("repeated source types were not interned once")
	}
	if got := string(outputs["wow_env.d.tl"]); !strings.Contains(got, "global GetPayload: function(): W.Payload\n") {
		t.Errorf("public return type was not qualified: %s", got)
	}

	// A finalized projection is independent of subsequent source changes.
	payload.Fields["count"] = &catalogueField{Native: &nativeParameter{Type: "string"}}
	c.globals["GetPayload"] = &catalogueFunction{ResolvedSignature: new("function(): string")}
	for name, expected := range outputs {
		if !bytes.Equal(projection.render()[name], expected) {
			t.Errorf("source mutation changed finalized %s", name)
		}
	}
}

// A function reached through a secret structure is projected to Data, so
// nothing named only in its signature becomes secret; a value beside it does.
func TestFunctionsReachedThroughSecretStructuresStayOpaque(t *testing.T) {
	c := newCatalogue()
	payload := c.record("Payload", "fixture", kindRecord)
	payload.Fields["count"] = &catalogueField{Native: &nativeParameter{Type: "number"}}
	payload.Fields["cb"] = &catalogueField{Type: "fun(o:Other)?"}
	payload.Fields["cb2"] = &catalogueField{Type: "fun(o:Other2):string"}
	payload.Fields["cb3"] = &catalogueField{Type: "fun():Other3"}
	payload.Fields["cb4"] = &catalogueField{Type: "(fun():string)|Beside2"}
	// Methods reach Data through projection before secrecy scans them; Pair
	// and Make pin that rendered contract, not the field-path scan above.
	payload.Methods["Pair"] = &catalogueFunction{Method: true, ResolvedSignature: new("function(self): function(): (string), Beside")}
	payload.Methods["Make"] = &catalogueFunction{Method: true, ResolvedSignature: new("function(self): function(): (Made)")}
	source := `---@param cb fun():string
---@param n number
---@return number
---@return string
---@overload fun(self:Payload, cb:(fun():string), n:number):number,string
---@overload fun(self:Payload):(fun():number)
function Payload:Run(cb, n) end
`
	doc, err := parseAnnotations([]byte(source), "payload.lua")
	if err != nil {
		t.Fatal(err)
	}
	payload.Methods["Run"] = functionFromAnnotation(doc.Functions[0], "fixture")
	for _, name := range []string{"Other", "Other2", "Other3", "Beside", "Beside2", "Made"} {
		c.record(name, "fixture", kindRecord).Fields["n"] = &catalogueField{Native: &nativeParameter{Type: "number"}}
	}
	c.setGlobal("GetPayload", &catalogueFunction{Source: "fixture", Native: &nativeFunction{Returns: []nativeParameter{{Type: "Payload", SecretValue: new(true)}}}})

	projection, err := c.finalize(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	declarations := string(projection.render()["wow.d.tl"])
	for _, expected := range []string{
		"      cb: W.SecretData\n",
		"      cb2: W.SecretData\n",
		"      cb3: W.SecretData\n",
		"      Make: function(self): W.SecretData\n",
		"      Pair: function(self): W.SecretData, Beside\n",
		"      Run: function(self, function(): (string), number): W.MaybeSecret<number>, W.MaybeSecret<string>\n",
		"      Run: function(self): W.SecretData\n",
		"   record Other\n      n: number\n   end",
		"   record Other2\n      n: number\n   end",
		"   record Other3\n      n: number\n   end",
		"   record Made\n      n: number\n   end",
		"   record Beside\n      n: W.MaybeSecret<number>\n   end",
		"   record Beside2\n      n: W.MaybeSecret<number>\n   end",
	} {
		if !strings.Contains(declarations, expected) {
			t.Errorf("missing declaration:\n%s\nactual:\n%s", expected, declarations)
		}
	}
	if strings.Count(declarations, "      Run: ") != 2 || strings.Contains(declarations, "function(Payload") {
		t.Errorf("Run overloads must render the receiver as self and collapse:\n%s", declarations)
	}
	// How a union with a function member projects is not a contract; only
	// that the function stays opaque is.
	if _, line, found := strings.Cut(declarations, "      cb4: "); !found || strings.Contains(line[:strings.IndexByte(line, '\n')], "function(") {
		t.Errorf("cb4 must stay opaque:\n%s", declarations)
	}
}

// Secrecy follows the rendered types: a structure named only inside a function
// type stays plain whichever path reaches it, and a structure is found under
// the name it renders as.
func TestSecrecyReadsRenderedTypesOnly(t *testing.T) {
	c := newCatalogue()
	payload := c.record("Payload", "fixture", kindRecord)
	payload.Fields["count"] = &catalogueField{Native: &nativeParameter{Type: "number"}}
	payload.Fields["renamed"] = &catalogueField{Type: "Other-11.0"}
	payload.Fields["plain"] = &catalogueField{Type: "Other12"}
	payload.Fields["handler"] = &catalogueField{Type: "Handler"}
	source := `---@return table<string, fun():Other5>
function Payload:Map() end
---@return Other7|fun():Other8
function Payload:Either() end
`
	doc, err := parseAnnotations([]byte(source), "payload.lua")
	if err != nil {
		t.Fatal(err)
	}
	payload.Methods["Map"] = functionFromAnnotation(doc.Functions[0], "fixture")
	payload.Methods["Either"] = functionFromAnnotation(doc.Functions[1], "fixture")
	c.aliases["Handlers"] = &catalogueAlias{Expression: "table<string, fun():Other6>", Source: "fixture"}
	c.aliases["Mixed"] = &catalogueAlias{Expression: "Other9|fun():Other10", Source: "fixture"}
	c.aliases["Handler"] = &catalogueAlias{Signature: functionFromNative(nativeFunction{Returns: []nativeParameter{{Type: "Other13"}}}, "fixture", false), Source: "fixture"}
	for _, name := range []string{"Other5", "Other6", "Other7", "Other8", "Other9", "Other10", "Other-11.0", "Other12", "Other13"} {
		c.record(name, "fixture", kindRecord).Fields["n"] = &catalogueField{Native: &nativeParameter{Type: "number"}}
	}
	for global, result := range map[string]string{"GetPayload": "Payload", "GetHandlers": "Handlers", "GetMixed": "Mixed"} {
		c.setGlobal(global, &catalogueFunction{Source: "fixture", Native: &nativeFunction{Returns: []nativeParameter{{Type: result, SecretValue: new(true)}}}})
	}

	projection, err := c.finalize(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	declarations := string(projection.render()["wow.d.tl"])
	for _, expected := range []string{
		"   record Other5\n      n: number\n   end",
		"   record Other6\n      n: number\n   end",
		"   record Other8\n      n: number\n   end",
		"   record Other10\n      n: number\n   end",
		"   record Other_11_0\n      n: W.MaybeSecret<number>\n   end",
		"   record Other12\n      n: W.MaybeSecret<number>\n   end",
		"   record Other13\n      n: number\n   end",
	} {
		if !strings.Contains(declarations, expected) {
			t.Errorf("missing declaration:\n%s\nactual:\n%s", expected, declarations)
		}
	}
	// How a union with a function member projects is not a contract; a
	// structure beside the function is secret exactly when it is rendered.
	for _, test := range []struct{ prefix, name string }{
		{"      Either: ", "Other7"},
		{" Global_GetMixed = ", "Other9"},
	} {
		_, line, found := strings.Cut(declarations, test.prefix)
		if !found {
			t.Fatalf("missing %q:\n%s", test.prefix, declarations)
		}
		named := slices.Contains(dottedIdentifiers.FindAllString(line[:strings.IndexByte(line, '\n')], -1), test.name)
		secret := strings.Contains(declarations, "   record "+test.name+"\n      n: W.MaybeSecret<number>\n   end")
		if named != secret {
			t.Errorf("%s rendered %v but secret %v:\n%s", test.name, named, secret, declarations)
		}
	}
}

func TestFinalizationRejectsInheritanceCycles(t *testing.T) {
	c := newCatalogue()
	c.record("First", "fixture", kindRecord).Bases = []string{"Second"}
	c.record("Second", "fixture", kindRecord).Bases = []string{"First"}
	if _, err := c.finalize(t.Context()); err == nil || !strings.Contains(err.Error(), "cyclic inheritance") {
		t.Fatalf("inheritance cycle error = %v", err)
	}
}

func TestMultilineSignaturesCannotShiftPolicyOrigins(t *testing.T) {
	for _, newline := range []string{"\n", "\r"} {
		t.Run(strconv.Quote(newline), func(t *testing.T) {
			c := newCatalogue()
			c.setGlobal("first", &catalogueFunction{ResolvedSignature: new("function(" + newline + "number): number")})
			c.setGlobal("second", &catalogueFunction{Native: &nativeFunction{Arguments: []nativeParameter{{Type: "number"}}}})
			_, err := func() (outputs map[string][]byte, err error) {
				defer recoverConversion(&err)
				projection, err := c.finalize(t.Context())
				if err != nil {
					return nil, err
				}
				return projection.render(), nil
			}()
			if err == nil || !strings.Contains(err.Error(), "one physical line") {
				t.Fatalf("multiline declaration must fail before publishing policy: %v", err)
			}
		})
	}
}

// Each declaration's source comment directly precedes it, whichever pass
// renders the declaration.
func TestSourceCommentsPrecedeTheirDeclarations(t *testing.T) {
	c := newCatalogue()
	c.setDef("E", &catalogueDefinition{Kind: kindEnum, Values: []string{`"a"`}, Source: "src-enum"})
	c.aliases["A"] = &catalogueAlias{Expression: "string | number", Source: "src-alias"}
	c.record("R", "src-record", kindRecord)
	c.record("I", "src-interface", kindInterface)
	projection, err := c.finalize(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(projection.render()["wow.d.tl"]), "\n")
	for source, declaration := range map[string]string{
		"src-enum": "   enum E", "src-alias": "   type A =",
		"src-record": "   record R", "src-interface": "   interface I",
	} {
		found := 0
		for i, line := range lines {
			if line == "   -- "+source {
				found++
				if i+1 == len(lines) || !strings.HasPrefix(lines[i+1], declaration) {
					t.Errorf("-- %s is followed by %q; want %q", source, lines[min(i+1, len(lines)-1)], declaration)
				}
			}
		}
		if found != 1 {
			t.Errorf("-- %s appears %d times; want once", source, found)
		}
	}
}

// A native callback alias declares its results bare, as a whole declaration
// does, and groups them where it is written inside another type, so they stay
// its results.
func TestNestedCallbackAliasesGroupResults(t *testing.T) {
	c := newCatalogue()
	c.aliases["Handler"] = &catalogueAlias{Signature: functionFromNative(nativeFunction{Returns: []nativeParameter{{Type: "Result"}, {Type: "number"}}}, "fixture", false), Source: "fixture"}
	c.record("Result", "fixture", kindRecord).Fields["n"] = &catalogueField{Native: &nativeParameter{Type: "number"}}
	c.record("Holder", "fixture", kindRecord).Fields["handlers"] = &catalogueField{Type: "Handler[]"}
	projection, err := c.finalize(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	outputs := projection.render()
	declarations := string(outputs["wow.d.tl"])
	for _, expected := range []string{
		"   type Handler = function(): Result, number\n",
		"   record Holder\n      handlers: {function(): (Result, number)}\n   end\n",
	} {
		if !strings.Contains(declarations, expected) {
			t.Errorf("missing declaration:\n%s\nactual:\n%s", expected, declarations)
		}
	}

	root := catalogueTestRoot(t)
	dir := filepath.Join(t.TempDir(), "types/wow")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	native, err := os.ReadFile(filepath.Join(root, "types/wow/native_values.d.tl"))
	if err != nil {
		t.Fatal(err)
	}
	outputs["native_values.d.tl"] = native
	for name, data := range outputs {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	caller := filepath.Join(dir, "caller.tl")
	source := `local type W=require("wow")
local function call(holder:W.Holder):(W.Result, number)
   return holder.handlers[1]()
end
`
	if err := os.WriteFile(caller, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	command := exec.CommandContext(t.Context(), StockTeal(t), "-I", dir, "--gen-target", "5.1", "--gen-compat", "off", "check", caller)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("nested callback results must stay its results: %v\n%s", err, output)
	}
}
