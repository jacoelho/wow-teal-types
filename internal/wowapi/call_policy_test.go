package wowapi

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestFinalizedNativeCallPermissions(t *testing.T) {
	c := newCatalogue(Project{})
	allowed := NativeParameter{Type: "number", ConditionalSecretContents: boolPointer(true)}
	denied := NativeParameter{Type: "unknown", NeverSecret: boolPointer(true)}
	native := func(arguments ...NativeParameter) *catalogueFunction {
		return &catalogueFunction{Native: &NativeFunction{Arguments: arguments, Policy: NativePolicy{SecretArguments: stringPointer("AllowedWhenTainted")}}}
	}
	c.setGlobal("source", &catalogueFunction{Native: &NativeFunction{Returns: []NativeParameter{{Type: "number", SecretValue: boolPointer(true)}}}})
	c.setGlobal("sourceUnknown", &catalogueFunction{Native: &NativeFunction{Returns: []NativeParameter{{Type: "unknown"}}}})
	c.setGlobal("allow", native(allowed))
	c.setGlobal("deny", native(denied))
	rootOnly := allowed
	rootOnly.Type = "unknown"
	rootOnly.NeverSecretContents = boolPointer(true)
	c.setGlobal("rootOnly", native(rootOnly))
	overload := native(allowed)
	overload.Variants = []*catalogueFunction{native(allowed, denied)}
	c.setGlobal("overloaded", overload)
	method := native(allowed, denied)
	method.Method = true
	c.record("Receiver", "fixture").Methods["Accept"] = method
	c.setGlobal("receiver", &catalogueFunction{Native: &NativeFunction{Returns: []NativeParameter{{Type: "Receiver", NeverSecret: boolPointer(true)}}}})
	c.abiTuples["Pair"] = []string{"number", "number"}
	pair := allowed
	pair.Type = "Pair"
	c.setGlobal("expanded", native(pair, denied))
	stride := 1
	rest := allowed
	rest.Type = "unknown"
	rest.StrideIndex = &stride
	c.setGlobal("variadic", native(denied, rest))
	c.setGlobal("mixedVariadic", native(rest, denied))
	for _, name := range []string{"canaccessvalue", "issecretvalue", "pretendGuard"} {
		c.setGlobal(name, &catalogueFunction{Native: &NativeFunction{Arguments: []NativeParameter{{Type: "unknown"}}, Returns: []NativeParameter{{Type: "boolean", NeverSecret: boolPointer(true)}}}})
	}
	projection, err := c.finalize(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	outputs := projection.render()
	// Finalized contracts own permission facts as well as their declaration text.
	*c.globals["allow"].Native.Policy.SecretArguments = "AllowedWhenUntainted"
	c.globals["allow"].Native.Arguments[0].NeverSecret = boolPointer(true)
	c.globals["allow"].Native.Arguments[0].ConditionalSecretContents = boolPointer(false)
	for name, expected := range outputs {
		if !bytes.Equal(projection.render()[name], expected) {
			t.Fatalf("builder permission mutation changed %s", name)
		}
	}
	root := catalogueTestRoot(t)
	fixture := t.TempDir()
	write := func(name string, data []byte) {
		t.Helper()
		path := filepath.Join(fixture, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	for name, data := range outputs {
		write("types/wow/"+name, data)
	}
	for _, name := range []string{"types/wow/native_values.d.tl", "tools/check_secrets.lua", "tools/teal/contract.lua"} {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		write(name, data)
	}
	cases := []struct{ name, source, rejection string }{
		{"allowed secret scalar", `allow(source())`, ""},
		{"never secret", `deny(source())`, "secret-value"},
		{"contents denied", `rootOnly(sourceUnknown())`, "call argument contents"},
		{"overload single", `overloaded(source())`, ""},
		{"overload second allowed", `overloaded(source(),1)`, ""},
		{"overload second denied", `overloaded(1,source())`, "secret-value"},
		{"method receiver offset", `receiver():Accept(source(),1)`, ""},
		{"method denied slot", `receiver():Accept(1,source())`, "secret-value"},
		{"ABI expanded slots", `expanded(source(),source(),1)`, ""},
		{"ABI following denied slot", `expanded(1,1,source())`, "secret-value"},
		{"variadic permission", `variadic(1,source(),source())`, ""},
		{"mixed variadic denied member", `mixedVariadic(source())`, "secret-value"},
		{"variadic fixed denied slot", `variadic(source(),1)`, "secret-value"},
		{"canonical guards", `local value=sourceUnknown()
if canaccessvalue(value) and not issecretvalue(value) and value is number then deny(value) end`, ""},
		{"guard lookalike", `local value=sourceUnknown()
if pretendGuard(value) and not issecretvalue(value) and value is number then deny(value) end`, "secret-value"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			caller := filepath.Join(fixture, "caller.tl")
			write("caller.tl", []byte(tc.source))
			command := exec.CommandContext(t.Context(), filepath.Join(root, "tools/teal/tl"), "-I", filepath.Join(fixture, "types/wow"), "--gen-target", "5.1", "--gen-compat", "off", "run", filepath.Join(fixture, "tools/check_secrets.lua"), "--", "wow_env", caller)
			output, err := command.CombinedOutput()
			if tc.rejection == "" {
				if err != nil {
					t.Fatalf("unexpected rejection: %v\n%s", err, output)
				}
			} else if err == nil || !strings.Contains(string(output), tc.rejection) || !strings.Contains(string(output), "caller.tl:") {
				t.Fatalf("expected %q rejection: %v\n%s", tc.rejection, err, output)
			}
		})
	}
}
