package wowapi

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestNativeTableContractsPreserveConcreteTypes(t *testing.T) {
	root := catalogueTestRoot(t)
	c := newCatalogue(Project{})
	c.aliases["bool"] = &catalogueAlias{Expression: "boolean"}
	c.aliases["LuaValueVariant"] = &catalogueAlias{Expression: "unknown"}
	for _, name := range []string{"LuaTableExtensionsDocumentation.lua", "AddOnProfilerDocumentation.lua"} {
		data, err := os.ReadFile(filepath.Join(root, "references/wow-api/blizzard", name))
		if err != nil {
			t.Fatal(err)
		}
		raw, err := ParseLua(data, name)
		if err != nil {
			t.Fatal(err)
		}
		metadata, err := decodeNativeMetadata(raw)
		if err != nil {
			t.Fatal(err)
		}
		namespace := c.record(metadata.Namespace, name)
		for _, function := range metadata.Functions {
			if metadata.Namespace == "table" || function.Name == "MeasureCall" {
				namespace.Methods[function.Name] = functionFromNative(function, name, false)
			}
		}
	}
	c.record("AddOnProfilerCallResults", "fixture")
	c.applyNativeTableContracts()
	projection, err := c.finalize(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	fixture := t.TempDir()
	dir := filepath.Join(fixture, "types/wow")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	for name, data := range projection.render() {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	native, err := os.ReadFile(filepath.Join(root, "types/wow/native_values.d.tl"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "native_values.d.tl"), native, 0644); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name, body string
		reject     bool
	}{
		{"map and array contracts", `local map:{string:number}={a=1}
local array:{boolean}={true,false}
local keys:{string}=tbl.keys(map)
local values:{number}=tbl.values(map)
local arrayKeys:{number}=tbl.keys(array)
local arrayValues:{boolean}=tbl.values(array)
local removed:boolean|nil=tbl.removeunordered(array)
local indexed:boolean|nil=tbl.removeunordered(array,1)
local frozenMap:{string:number}=tbl.freeze(map)
local frozenArray:{boolean}=tbl.freeze(array)
local count:W.NumberLike=tbl.count(map)
local arrayCount:W.NumberLike=tbl.count(array)
local n,k,m:W.NumberLike,W.NumberLike,W.NumberLike=tbl.getcountinfo(map)
local an,ak,am:W.NumberLike,W.NumberLike,W.NumberLike=tbl.getcountinfo(array)
local contains:boolean=tbl.contains(map,"not a number")
local arrayContains:boolean=tbl.contains(array,99)
local index:number|nil=tbl.indexof(map,false)
local arrayIndex:number|nil=tbl.indexof(array,"not a boolean")
local empty:boolean|W.OpaqueValue=tbl.isempty(map)
local arrayEmpty:boolean|W.OpaqueValue=tbl.isempty(array)
local frozen:boolean=tbl.isfrozen(map)
local arrayFrozen:boolean=tbl.isfrozen(array)
local removedCount:number=tbl.removevalue(map,"missing")
local arrayRemovedCount:number=tbl.removevalue(array,10)
local measurement:W.AddOnProfilerCallResults=profiler.MeasureCall(function():number return 1 end)`, false},
		{"keys are not values", `local map:{string:number}={a=1}
local wrong:{number}=tbl.keys(map)`, true},
		{"values are not keys", `local map:{string:number}={a=1}
local wrong:{string}=tbl.values(map)`, true},
		{"removed element type", `local array:{boolean}={true}
local wrong:string=tbl.removeunordered(array)`, true},
		{"removal requires array", `local wrong=tbl.removeunordered(1)`, true},
		{"freeze preserves map", `local map:{string:number}={a=1}
local wrong:{string:string}=tbl.freeze(map)`, true},
		{"freeze preserves array", `local array:{boolean}={true}
local wrong:{string}=tbl.freeze(array)`, true},
		{"freeze requires table", `local wrong=tbl.freeze(1)`, true},
		{"count remains restricted", `local map:{string:number}={a=1}
local wrong:number=tbl.count(map)`, true},
		{"measurement tail remains unknown", `local measurement,value:W.AddOnProfilerCallResults,number=profiler.MeasureCall(function():number return 1 end)`, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			source := `local type W=require("wow")
local function check(tbl:W.Namespace_table,profiler:W.C_AddOnProfiler)
` + tc.body + "\nend\n"
			caller := filepath.Join(fixture, "caller.tl")
			if err := os.WriteFile(caller, []byte(source), 0644); err != nil {
				t.Fatal(err)
			}
			command := exec.CommandContext(t.Context(), filepath.Join(root, "tools/teal/tl"), "-I", dir, "--gen-target", "5.1", "--gen-compat", "off", "check", caller)
			output, err := command.CombinedOutput()
			if !tc.reject {
				if err != nil {
					t.Fatalf("unexpected rejection: %v\n%s", err, output)
				}
			} else if err == nil || !strings.Contains(string(output), "caller.tl:") || (!strings.Contains(string(output), "got") && !strings.Contains(string(output), "unknown value must be narrowed")) {
				t.Fatalf("expected caller type mismatch: %v\n%s", err, output)
			}
		})
	}
}
