package wowapi

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
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
	if doc["Enabled"] != false || asMap(asList(doc["Arguments"])[0])["Default"] != false || doc["Other"] != "raw" || doc["Escaped"] != "A\n" || doc["Arithmetic"] != int64(1) {
		t.Fatalf("literal values changed: %#v", doc)
	}
	expected := M{"$expression": []any{M{"$ref": "Enum.Test.One"}, "+", int64(1)}}
	if !reflect.DeepEqual(doc["Value"], expected) {
		t.Fatalf("reference arithmetic: %#v", doc["Value"])
	}
	doc, e = ParseLua([]byte("local Example={Documentation={\"line\\\nnext\"}}"), "physical-newline.lua")
	if e != nil || !reflect.DeepEqual(doc["Documentation"], []any{"line\nnext"}) {
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
	convert := func(raw string) (string, error) {
		return captureString(func() string {
			return (typeConverter{raw: raw, self: "Owner", resolve: func(s string) string { return s }, literal: func(values []string) string { return "Enum_" + strings.Join(values, "_") }, record: func(fields []recordField) string { return "NamedRecord" }}).walk(raw, 0)
		})
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
func TestAnnotationInventoryRetainsOwnerTypesAndDiagnostics(t *testing.T) {
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
	functions := asList(doc["functions"])
	if len(functions) != 2 {
		t.Fatalf("functions: %#v", functions)
	}
	method := asMap(functions[0])
	if method["owner"] != "Example" || method["raw_owner"] != "E" || method["method"] != true {
		t.Fatalf("owner lost: %#v", method)
	}
	param := asMap(asList(method["params"])[0])
	if param["type"] != "table<string, number>" || param["optional"] != true || param["description"] != "explanation" {
		t.Fatalf("parameter lost: %#v", param)
	}
	ret := asMap(asList(method["returns"])[0])
	if ret["type"] != "number" || ret["name"] != "count" || ret["description"] != "explanation" {
		t.Fatalf("return lost: %#v", ret)
	}
	diagnostics := asList(doc["diagnostics"])
	if len(diagnostics) != 1 || asMap(diagnostics[0])["code"] != "unknown_annotation" {
		t.Fatalf("unknown metadata disappeared: %#v", diagnostics)
	}
}

func TestLuaIntegerArithmeticDoesNotRoundThroughFloat(t *testing.T) {
	doc, e := ParseLua([]byte(`local Example={Large=9007199254740993+1,Negative=-9007199254740993,Wide=9223372036854775807+1}`), "integer.lua")
	if e != nil {
		t.Fatal(e)
	}
	if doc["Large"] != int64(9007199254740994) || doc["Negative"] != int64(-9007199254740993) || doc["Wide"] != json.Number("9223372036854775808") {
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
