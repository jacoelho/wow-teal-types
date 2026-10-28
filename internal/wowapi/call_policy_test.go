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
	c := newCatalogue()
	allowed := nativeParameter{Type: "number", ConditionalSecretContents: new(true)}
	denied := nativeParameter{Type: "unknown", NeverSecret: new(true)}
	native := func(arguments ...nativeParameter) *catalogueFunction {
		return &catalogueFunction{Native: &nativeFunction{Arguments: arguments, Policy: nativePolicy{SecretArguments: new("AllowedWhenTainted")}}}
	}
	c.setGlobal("source", &catalogueFunction{Native: &nativeFunction{Returns: []nativeParameter{{Type: "number", SecretValue: new(true)}}}})
	c.setGlobal("sourceUnknown", &catalogueFunction{Native: &nativeFunction{Returns: []nativeParameter{{Type: "unknown", ConditionalSecret: new(true)}}}})
	c.setGlobal("allow", native(allowed))
	c.setGlobal("deny", native(denied))
	c.setGlobal("secureAllow", native(allowed))
	c.setGlobal("secureDeny", native(denied))
	c.secureGlobals["secureAllow"] = true
	c.secureGlobals["secureDeny"] = true
	c.setGlobal("sourceTable", &catalogueFunction{Native: &nativeFunction{Returns: []nativeParameter{{Type: "table", InnerType: "number", ConditionalSecretContents: new(true)}}}})
	c.setGlobal("publicTable", native(nativeParameter{Type: "table", InnerType: "number", NeverSecretContents: new(true)}))
	overload := native(allowed)
	overload.Variants = []*catalogueFunction{native(allowed, denied)}
	c.setGlobal("overloaded", overload)
	method := native(allowed, denied)
	method.Method = true
	c.record("Receiver", "fixture", kindRecord).Methods["Accept"] = method
	c.setGlobal("receiver", &catalogueFunction{Native: &nativeFunction{Returns: []nativeParameter{{Type: "Receiver", NeverSecret: new(true)}}}})
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
		c.setGlobal(name, &catalogueFunction{Native: &nativeFunction{Arguments: []nativeParameter{{Type: "unknown"}}, Returns: []nativeParameter{{Type: "boolean", NeverSecret: new(true)}}}})
	}
	projection, err := c.finalize(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	outputs := projection.render()
	// Finalized contracts own permission facts as well as their declaration text.
	*c.globals["allow"].Native.Policy.SecretArguments = "AllowedWhenUntainted"
	c.globals["allow"].Native.Arguments[0].NeverSecret = new(true)
	c.globals["allow"].Native.Arguments[0].ConditionalSecretContents = new(false)
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
	for _, name := range []string{"types/wow/native_values.d.tl", "tools/check_secrets.lua"} {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		write(name, data)
	}
	cases := []struct{ name, source, rejection string }{
		{"secure origin allows secret", `secureAllow(source())`, ""},
		{"secure origin denies secret", `secureDeny(source())`, "secret-value"},
		{"allowed secret scalar", `allow(source())`, ""},
		{"never secret", `deny(source())`, "secret-value"},
		{"secret table contents denied", `publicTable(sourceTable())`, "got"},
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
	tl := StockTeal(t)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			caller := filepath.Join(fixture, "caller.tl")
			write("caller.tl", []byte(tc.source))
			environment := "wow_env"
			if strings.HasPrefix(tc.name, "secure origin") {
				environment = "wow_secure_env"
			}
			command := exec.CommandContext(t.Context(), tl, "-I", filepath.Join(fixture, "types/wow"), "--gen-target", "5.1", "--gen-compat", "off", "run", filepath.Join(fixture, "tools/check_secrets.lua"), "--", environment, caller)
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

func TestOverloadsDifferingOnlyByNilCollapseUnlessPermissionsDiffer(t *testing.T) {
	native := func(secretArguments string) *catalogueFunction {
		return &catalogueFunction{Native: &nativeFunction{Arguments: []nativeParameter{{Type: "number"}}, Policy: nativePolicy{SecretArguments: new(secretArguments)}}}
	}
	resolved := func(item *catalogueFunction, signature string) *catalogueFunction {
		item.ResolvedSignature = new(signature)
		return item
	}
	t.Run("same permission", func(t *testing.T) {
		c := newCatalogue()
		item := resolved(native("NotAllowed"), "function(number): string")
		item.Variants = []*catalogueFunction{resolved(native("NotAllowed"), "function(number): string | nil")}
		c.setGlobal("collapse", item)
		projection, err := c.finalize(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		if got := string(projection.render()["wow_env.d.tl"]); !strings.Contains(got, "global collapse: function(number): string\n") {
			t.Errorf("nil-only variants were not collapsed to one overload:\n%s", got)
		}
	})
	t.Run("different permission", func(t *testing.T) {
		c := newCatalogue()
		item := resolved(native("NotAllowed"), "function(number): string")
		item.Variants = []*catalogueFunction{resolved(native("AllowedWhenTainted"), "function(number): string | nil")}
		c.setGlobal("collapse", item)
		_, err := c.finalize(t.Context())
		if err == nil || !strings.Contains(err.Error(), "ambiguous permissions for identical overload") {
			t.Fatalf("error = %v", err)
		}
	})
}
