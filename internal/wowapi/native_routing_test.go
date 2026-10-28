package wowapi_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"wow-teal-types/internal/wowapi"
)

func TestLayeredRegionContractsPreserveSourceAndReceiverCapabilities(t *testing.T) {
	root := catalogueTestRoot(t)
	artifacts, err := catalogueTestProject(t).Generate(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	names := []string{"GetDrawLayer", "SetDrawLayer", "GetVertexColor", "SetVertexColor", "SetVertexColorFromBoolean"}
	outputs := artifacts.Files
	policy := string(outputs["types/wow/call_policy.lua"])
	for _, name := range names {
		symbol := "LayeredRegion." + name
		if strings.Count(policy, `symbol = "`+symbol+`"`) != 1 || strings.Contains(policy, `symbol = "Region.`+name+`"`) {
			t.Errorf("policy owner mismatch: %s", symbol)
		}
	}
	if !strings.Contains(policy, `symbol = "LayeredRegion.SetVertexColor", secret_parameters = {2, 3, 4, 5}`) {
		t.Error("moved vertex-color contract lost receiver-offset permissions")
	}
	fixture := t.TempDir()
	dir := filepath.Join(fixture, "types/wow")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	for name, data := range outputs {
		if err := os.WriteFile(filepath.Join(fixture, name), data, 0644); err != nil {
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
		{"supported receiver contracts", `frame:SetAlpha(1)
frame:Hide()
model:SetDrawLayer("ARTWORK")
font:SetDrawLayer("ARTWORK",2)
texture:SetDrawLayer("ARTWORK",2)
vector:SetDrawLayer("ARTWORK",2)
font:SetVertexColor(1,1,1)
texture:SetVertexColor(1,1,1)
vector:SetVertexColor(1,1,1)`, false},
		{"Frame has no layer", `frame:SetDrawLayer("ARTWORK",2)`, true},
		{"Frame has no vertex color", `frame:SetVertexColor(1,1,1)`, true},
		{"ModelScene keeps narrower native arity", `model:SetDrawLayer("ARTWORK",2)`, true},
	}
	tl := wowapi.StockTeal(t)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			source := `local type W=require("wow")
local function check(frame:W.Frame,model:W.ModelScene,font:W.FontString,texture:W.TextureBase,vector:W.VectorGraphics)
` + tc.body + "\nend\n"
			path := filepath.Join(fixture, "caller.tl")
			if err := os.WriteFile(path, []byte(source), 0644); err != nil {
				t.Fatal(err)
			}
			command := exec.CommandContext(t.Context(), tl, "-I", dir, "--gen-target", "5.1", "--gen-compat", "off", "check", path)
			output, err := command.CombinedOutput()
			if !tc.reject {
				if err != nil {
					t.Fatalf("unexpected rejection: %v\n%s", err, output)
				}
			} else if err == nil || !strings.Contains(string(output), "caller.tl:") || strings.Contains(string(output), "syntax error") {
				t.Fatalf("expected receiver contract rejection: %v\n%s", err, output)
			}
		})
	}
}
