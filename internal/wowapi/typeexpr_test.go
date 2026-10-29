package wowapi

import (
	"maps"
	"regexp"
	"slices"
	"strings"
	"testing"
	"wow-teal-types/internal/wowapi/literal"
)

func TestClosedLuaMetadataGrammar(t *testing.T) {
	source := `-- retained literals
local Example={Name="Example", Enabled=false, Arguments={{Name="value",Type="number",Default=false}}, Value=Enum.Test.One+1, Arithmetic=-1+2, Other=[=[
raw]=], Escaped="\065\n"};
APIDocumentation:AddDocumentationTable(Example);`
	doc, e := parseLua([]byte(source), "fixture.lua")
	if e != nil {
		t.Fatal(e)
	}
	if !doc["Enabled"].Equal(literal.Bool(false)) || !rawArgument(t, doc, "Arguments", 0)["Default"].Equal(literal.Bool(false)) || !doc["Other"].Equal(literal.Text("raw")) || !doc["Escaped"].Equal(literal.Text("A\n")) || !doc["Arithmetic"].Equal(mustLiteral(literal.BigInteger("1"))) {
		t.Fatalf("literal values changed: %#v", doc)
	}
	encoded, _ := doc["Value"].MarshalJSON()
	if string(encoded) != `{"$expression":[{"$ref":"Enum.Test.One"},"+",1]}` {
		t.Fatalf("reference arithmetic: %#v", doc["Value"])
	}
	doc, e = parseLua([]byte("local Example={Documentation={\"line\\\nnext\"}}"), "physical-newline.lua")
	if e != nil || !doc["Documentation"].Equal(literal.List([]literal.Value{literal.Text("line\nnext")})) {
		t.Fatalf("escaped physical newline: %#v, %v", doc, e)
	}
	for _, input := range []string{`local Example={Name="a",["Name"]="b"}`, `local Example={[1]="a",["1"]="b"}`, `local Example={["$ref"]="ambiguous"}`, `Run(); local Example={Name="a"}`, `local Example={Name="a"};Example.Name="b"`, `local Example={Name="a"};APIDocumentation:AddDocumentationTable(Other)`, `local Example={Value=true+1}`, `local Example={Value=1e999}`, `local Example={Value=1e308+1e308}`, `local Example={Value="\x41"}`, `local Example={Value="\255"}`, "local Example={Value=\"line\nnext\"}", `local Example={Name="a"};--[[unclosed`, `local Example={Name=[[unclosed}`, `local Example={Value=function() end}`} {
		t.Run(input, func(t *testing.T) {
			if _, e := parseLua([]byte(input), "rejected.lua"); e == nil {
				t.Fatal("accepted executable, ambiguous or unsupported metadata")
			}
		})
	}
}
func TestTypeExpressionsPreserveExplicitContracts(t *testing.T) {
	// Interned names are content hashes; describe expands them into the
	// contract they carry so expectations do not depend on hash values.
	interned := regexp.MustCompile(`\b(?:Literal|Shape|Callback)_[0-9a-f]{12}\b`)
	var describe func(c *catalogue, value string) string
	describe = func(c *catalogue, value string) string {
		return interned.ReplaceAllStringFunc(value, func(name string) string {
			item := c.defs[name]
			switch item.Kind {
			case kindEnum:
				return "enum(" + strings.Join(item.Values, ",") + ")"
			case kindCallback:
				return "callback(" + describe(c, item.Signature) + ")"
			case kindConvertedRecord:
				fields := []string{}
				for _, field := range slices.Sorted(maps.Keys(item.Fields)) {
					fields = append(fields, field+": "+describe(c, item.Fields[field].Type))
				}
				return "shape{" + strings.Join(fields, ", ") + "}"
			}
			t.Fatalf("unexpected interned kind %s for %s", item.Kind, name)
			return ""
		})
	}
	convert := func(raw string) (value string, err error) {
		defer recoverConversion(&err)
		c := newCatalogue()
		return describe(c, c.tealValue(raw, "Owner", nil, outputPosition)), nil
	}
	for _, test := range []struct{ raw, want string }{{"number|nil", "number | nil"}, {"table<string,number>", "{string: number}"}, {"(number|string)[]", "{number | string}"}, {"{[1]:string,[2]:number}", "{string, number}"}, {"fun(self, enabled?:boolean,...:string): number,string", "callback(function(Owner, ? boolean, string...): number, string)"}, {"'b'|'a'|'a'", "enum(a,b)"}, {"'one'?|'two'", "nil | enum(one,two)"}, {"{value:number}", "shape{value: number}"}, {"true|5", "boolean | number"}} {
		got, e := convert(test.raw)
		if e != nil || got != test.want {
			t.Errorf("%s: %q, %v; want %q", test.raw, got, e, test.want)
		}
	}
	for _, raw := range []string{"`T`", "table<number>", "{[2]:number}", "{[number]:number,[string]:string}", "fun(value)", "(number]", "number|", "fun(value:number):"} {
		if _, e := convert(raw); e == nil {
			t.Errorf("accepted unsupported type %q", raw)
		}
	}
}
func TestOmitNilRemovesOnlyNilUnionMembers(t *testing.T) {
	omit := func(expression string) (value string, err error) {
		defer recoverConversion(&err)
		return omitNil(expression), nil
	}
	for _, test := range []struct{ expression, want string }{
		{"number | nil", "number"},
		{"nil | number", "number"},
		{"A | nil | B", "A | B"},
		{"function(? number): string | nil", "function(? number): string"},
		{"{string: nil | number}", "{string: number}"},
		{"number | nil...", "number..."},
		{"nil", "nil"},
		{"nilCount | number", "nilCount | number"},
		{"function(): nil", "function(): nil"},
	} {
		if got, e := omit(test.expression); e != nil || got != test.want {
			t.Errorf("%s: %q, %v; want %q", test.expression, got, e, test.want)
		}
	}
	for _, expression := range []string{`"a" | nil`, "'a'", "`a`"} {
		if _, e := omit(expression); e == nil {
			t.Errorf("accepted quoted literal %q", expression)
		}
	}
}

func TestAnnotationsPreserveContractsAndIgnoreUnsupportedTags(t *testing.T) {
	source := `---@class Example: Base
---@field private enabled? boolean caption
local E = {}
---@param value? table<string, number> explanation
---@return number count explanation
---@generic T: table
---@overload fun(self:Example,value:string): number
function E:Read(value) end
---@strange future value
function Unknown() end
`
	doc, e := parseAnnotations([]byte(source), "Example.lua")
	if e != nil {
		t.Fatal(e)
	}
	functions := doc.Functions
	if len(functions) != 2 {
		t.Fatalf("functions: %#v", functions)
	}
	method := functions[0]
	if stringValue(method.Owner) != "Example" || !method.Method {
		t.Fatalf("owner lost: %#v", method)
	}
	param := method.Params[0]
	if param.Type != "table<string, number>" || !param.Optional {
		t.Fatalf("parameter lost: %#v", param)
	}
	ret := method.Returns[0]
	if ret.Type != "number" {
		t.Fatalf("return lost: %#v", ret)
	}

}

func TestAnnotationsRejectMalformedFunctionContracts(t *testing.T) {
	for _, annotation := range []string{"---@return @invalid\n---@return number", "---@return \"unterminated", "---@return", "---@param", "---@param value"} {
		t.Run(annotation, func(t *testing.T) {
			_, err := parseAnnotations([]byte(annotation+"\nfunction Broken(value) end"), "broken.lua")
			if err == nil || !strings.Contains(err.Error(), "broken.lua:1:") {
				t.Fatalf("malformed contract did not fail at its source: %v", err)
			}
		})
	}
}

func TestAnnotationsIgnoreOrdinaryLua(t *testing.T) {
	doc, err := parseAnnotations([]byte("---@strange ignored\nCallSomething()\nif ready then\n---@return number\nfunction Valid() end\nend"), "ordinary.lua")
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Functions) != 1 || doc.Functions[0].Name != "Valid" || len(doc.Functions[0].Returns) != 1 || doc.Functions[0].Returns[0].Type != "number" {
		t.Fatalf("ordinary Lua changed the declaration contract: %#v", doc.Functions)
	}
}

func TestLuaIntegerArithmeticDoesNotRoundThroughFloat(t *testing.T) {
	doc, e := parseLua([]byte(`local Example={Large=9007199254740993+1,Negative=-9007199254740993,Wide=9223372036854775807+1}`), "integer.lua")
	if e != nil {
		t.Fatal(e)
	}
	large, _ := doc["Large"].Integer()
	negative, _ := doc["Negative"].Integer()
	wide, _ := doc["Wide"].Integer()
	if large != "9007199254740994" || negative != "-9007199254740993" || wide != "9223372036854775808" {
		t.Fatalf("integer precision lost: %#v", doc)
	}
}
func TestInvalidUTF8CannotBecomeReplacementCharacters(t *testing.T) {
	data := append([]byte(`local Example={Name="`), 255)
	data = append(data, []byte(`"}`)...)
	if _, e := parseLua(data, "bad.lua"); e == nil {
		t.Fatal("invalid UTF8 native metadata accepted")
	}
	if _, e := parseAnnotations([]byte{'-', '-', '-', '@', 'c', 'l', 'a', 's', 's', ' ', 255}, "bad.lua"); e == nil {
		t.Fatal("invalid UTF8 annotation accepted")
	}
}

func rawArgument(t *testing.T, doc literal.Object, key string, index int) literal.Object {
	t.Helper()
	values, ok := doc[key].Array()
	if !ok || index >= len(values) {
		t.Fatalf("missing raw %s[%d]", key, index)
	}
	value, ok := values[index].Object()
	if !ok {
		t.Fatalf("raw %s[%d] is not an object", key, index)
	}
	return value
}
