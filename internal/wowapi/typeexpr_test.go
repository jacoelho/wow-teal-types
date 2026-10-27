package wowapi

import (
	"strings"
	"testing"
	"wow-teal-types/internal/wowapi/literal"
)

func TestClosedLuaMetadataGrammar(t *testing.T) {
	source := `-- retained literals
local Example={Name="Example", Enabled=false, Arguments={{Name="value",Type="number",Default=false}}, Value=Enum.Test.One+1, Arithmetic=-1+2, Other=[=[
raw]=], Escaped="\065\n"};
APIDocumentation:AddDocumentationTable(Example);`
	doc, e := ParseLua([]byte(source), "fixture.lua")
	if e != nil {
		t.Fatal(e)
	}
	if !doc["Enabled"].Equal(literal.Bool(false)) || !rawArgument(t, doc, "Arguments", 0)["Default"].Equal(literal.Bool(false)) || !doc["Other"].Equal(literal.Text("raw")) || !doc["Escaped"].Equal(literal.Text("A\n")) || !doc["Arithmetic"].Equal(literal.Int(1)) {
		t.Fatalf("literal values changed: %#v", doc)
	}
	encoded, _ := doc["Value"].MarshalJSON()
	if string(encoded) != `{"$expression":[{"$ref":"Enum.Test.One"},"+",1]}` {
		t.Fatalf("reference arithmetic: %#v", doc["Value"])
	}
	doc, e = ParseLua([]byte("local Example={Documentation={\"line\\\nnext\"}}"), "physical-newline.lua")
	if e != nil || !doc["Documentation"].Equal(literal.List([]literal.Value{literal.Text("line\nnext")})) {
		t.Fatalf("escaped physical newline: %#v, %v", doc, e)
	}
	for _, input := range []string{`local Example={Name="a",["Name"]="b"}`, `local Example={[1]="a",["1"]="b"}`, `local Example={["$ref"]="ambiguous"}`, `Run(); local Example={Name="a"}`, `local Example={Name="a"};Example.Name="b"`, `local Example={Name="a"};APIDocumentation:AddDocumentationTable(Other)`, `local Example={Value=true+1}`, `local Example={Value=1e999}`, `local Example={Value=1e308+1e308}`, `local Example={Value="\x41"}`, `local Example={Value="\255"}`, "local Example={Value=\"line\nnext\"}", `local Example={Name="a"};--[[unclosed`, `local Example={Name=[[unclosed}`, `local Example={Value=function() end}`} {
		t.Run(input, func(t *testing.T) {
			if _, e := ParseLua([]byte(input), "rejected.lua"); e == nil {
				t.Fatal("accepted executable, ambiguous or unsupported metadata")
			}
		})
	}
}
func TestTypeExpressionsPreserveExplicitContracts(t *testing.T) {
	convert := func(raw string) (value string, err error) {
		defer captureError(&err)
		return (typeConverter{raw: raw, self: "Owner", resolve: func(s string) string { return s }, literal: func(values []string) string { return "Enum_" + strings.Join(values, "_") }, record: func(fields []recordField) string { return "NamedRecord" }}).walk(raw, 0), nil
	}
	for _, test := range []struct{ raw, want string }{{"number|nil", "number | nil"}, {"table<string,number>", "{string: number}"}, {"(number|string)[]", "{number | string}"}, {"{[1]:string,[2]:number}", "{string, number}"}, {"fun(self, enabled?:boolean,...:string): number,string", "function(Owner, ? boolean, string...): number, string"}, {"'b'|'a'|'a'", "Enum_a_b"}, {"'one'?|'two'", "nil | Enum_one_two"}, {"{value:number}", "NamedRecord"}, {"true|5", "boolean | number"}} {
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
	doc, e := ParseAnnotations([]byte(source), "Example.lua")
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
			_, err := ParseAnnotations([]byte(annotation+"\nfunction Broken(value) end"), "broken.lua")
			if err == nil || !strings.Contains(err.Error(), "broken.lua:1:") {
				t.Fatalf("malformed contract did not fail at its source: %v", err)
			}
		})
	}
}

func TestAnnotationsIgnoreOrdinaryLua(t *testing.T) {
	doc, err := ParseAnnotations([]byte("---@strange ignored\nCallSomething()\nif ready then\n---@return number\nfunction Valid() end\nend"), "ordinary.lua")
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Functions) != 1 || doc.Functions[0].Name != "Valid" || len(doc.Functions[0].Returns) != 1 || doc.Functions[0].Returns[0].Type != "number" {
		t.Fatalf("ordinary Lua changed the declaration contract: %#v", doc.Functions)
	}
}

func TestLuaIntegerArithmeticDoesNotRoundThroughFloat(t *testing.T) {
	doc, e := ParseLua([]byte(`local Example={Large=9007199254740993+1,Negative=-9007199254740993,Wide=9223372036854775807+1}`), "integer.lua")
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
	if _, e := ParseLua(data, "bad.lua"); e == nil {
		t.Fatal("invalid UTF8 native metadata accepted")
	}
	if _, e := ParseAnnotations([]byte{'-', '-', '-', '@', 'c', 'l', 'a', 's', 's', ' ', 255}, "bad.lua"); e == nil {
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
