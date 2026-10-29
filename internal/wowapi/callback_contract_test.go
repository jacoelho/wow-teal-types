package wowapi

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCallbackContractsBindResultsToSource(t *testing.T) {
	source := `---@generic A, R
---@generic S : string
---@param funcOrKey S|fun(...: A): ...:R
---@param ... A
---@return R ...
function securecall(funcOrKey, ...) end
---@generic A, R
---@param func fun(...: A): ...:R
---@param ... A
---@return R ...
function securecallfunction(func, ...) end
---@return function
function opaqueGetter() end
---@return fun(value:number):string
function typedGetter() end
---@generic T
---@return T
function unboundResult() end
---@generic K,V
---@return fun(state:table<K,V>,index?:K):K,V
---@return table<K,V>
function unboundIterator() end
`
	doc, err := parseAnnotations([]byte(source), "callbacks.lua")
	if err != nil {
		t.Fatal(err)
	}
	c := newCatalogue()
	for _, function := range doc.Functions {
		c.setGlobal(function.Name, functionFromAnnotation(function, "fixture"))
	}
	c.applySecureCallContracts()
	projection, err := c.finalize(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	root := catalogueTestRoot(t)
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
	callbacks := `local type W=require("wow")
local function echo(...:number):string... return "value" end
local function numeric(...:number):number... return 1 end
local function opaque(...:number):W.Data... return "value" end
`
	cases := []struct {
		name, body string
		reject     bool
	}{
		{"homogeneous callback relations", `local text:string=securecallfunction(echo,1)
local other:string=securecall(echo,1)
local n:number=securecallfunction(numeric,2)
local m:number=securecall(numeric,2)
local callback=typedGetter()
local value:string=callback(1)
local hidden:W.Data=opaqueGetter()
local unbound:any=unboundResult()
local iterator,state:any,any=unboundIterator()
local lookup:W.Data=securecall("DynamicName",1)`, false},
		{"function callback cannot forge result", `local wrong:number=securecallfunction(echo,1)`, true},
		{"callable overload cannot forge result", `local wrong:number=securecall(echo,1)`, true},
		// Stock Teal 0.24.8 lets an expected type instantiate R through a
		// callback returning Data, so only the uncontextualised result is bound.
		{"opaque callback result stays data", `local value=securecallfunction(opaque,1)
local wrong:number=value`, true},
		{"string lookup cannot bind result", `local wrong:number=securecall("DynamicName",1)`, true},
		{"unproven callable cannot be invoked", `local callback=opaqueGetter()
callback()`, true},
		{"known callback retains argument contract", `local callback=typedGetter()
local wrong=callback("text")`, true},
		{"output generic cannot forge scalar", `local wrong:number=unboundResult()`, true},
		{"output generics cannot forge state", `local iterator,state:any,{string:number}=unboundIterator()`, true},
	}
	tl := StockTeal(t)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			caller := filepath.Join(fixture, "caller.tl")
			if err := os.WriteFile(caller, []byte(callbacks+tc.body), 0644); err != nil {
				t.Fatal(err)
			}
			command := exec.CommandContext(t.Context(), tl, "-I", dir, "--global-env-def", "wow_env", "--gen-target", "5.1", "--gen-compat", "off", "check", caller)
			output, err := command.CombinedOutput()
			if !tc.reject {
				if err != nil {
					t.Fatalf("unexpected rejection: %v\n%s", err, output)
				}
			} else if err == nil || !strings.Contains(string(output), "caller.tl:") || strings.Contains(string(output), "syntax error") {
				t.Fatalf("expected caller type rejection: %v\n%s", err, output)
			}
		})
	}
}
