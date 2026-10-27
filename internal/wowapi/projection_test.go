package wowapi

import (
	"bytes"
	"strings"
	"testing"
)

func TestFinalizedDeclarationsResolveGeneratedTypesInheritanceAndSecrecy(t *testing.T) {
	c := newCatalogue(Project{})
	base := c.record("Base", "fixture")
	asMap(base["fields"])["count"] = M{"native": true, "Type": "number"}
	asMap(base["methods"])["Get"] = M{"native": true, "method": true, "Returns": []any{M{"Type": "number"}}}
	child := c.record("Child", "fixture")
	child["bases"] = []any{"Base"}
	asMap(child["fields"])["count"] = M{"type": "number"}
	asMap(child["methods"])["Get"] = M{"returns": []any{M{"type": "string"}}}
	payload := c.record("Payload", "fixture")
	asMap(payload["fields"])["count"] = M{"native": true, "Type": "number"}
	asMap(payload["fields"])["visible"] = M{"native": true, "Type": "boolean", "NeverSecret": true}
	holder := c.record("Holder", "fixture")
	asMap(holder["fields"])["left"] = M{"type": "{value:number}"}
	asMap(holder["fields"])["right"] = M{"type": "{value:number}"}
	asMap(holder["fields"])["callback"] = M{"type": "fun(value:number):string"}
	c.setGlobal("GetPayload", M{"source": "fixture", "native": true, "Returns": []any{M{"Type": "Payload", "SecretValue": true}}})

	projection, err := c.finalize(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	outputs := projection.render()
	declarations := string(outputs["wow.d.tl"])
	for _, expected := range []string{
		"   record Base\n      count: number\n      Get: function(self): number\n   end",
		"   record Child\n      is Base\n   end",
		"   record Payload\n      count: W.NumberLike\n      visible: boolean\n   end",
		"   record Holder\n      callback: Callback_4e54a059defc\n      left: Shape_bc707ec91b8a\n      right: Shape_bc707ec91b8a\n   end",
		"   record Shape_bc707ec91b8a\n      value: number\n   end",
		"   type Callback_4e54a059defc = function(number): string",
		"   type Global_GetPayload = function(): Payload",
	} {
		if !strings.Contains(declarations, expected) {
			t.Errorf("missing declaration:\n%s\nactual:\n%s", expected, declarations)
		}
	}
	if strings.Count(declarations, "record Shape_") != 1 || strings.Count(declarations, "type Callback_") != 1 {
		t.Error("repeated source types were not interned once")
	}
	if got := string(outputs["wow_env.d.tl"]); !strings.Contains(got, "global GetPayload: function(): W.Payload\n") {
		t.Errorf("public return type was not qualified: %s", got)
	}

	// A finalized projection is independent of subsequent source changes.
	asMap(payload["fields"])["count"] = M{"native": true, "Type": "string"}
	c.globals["GetPayload"] = M{"resolved_signature": "function(): string"}
	for name, expected := range outputs {
		if !bytes.Equal(projection.render()[name], expected) {
			t.Errorf("source mutation changed finalized %s", name)
		}
	}
}

func TestFinalizationRejectsInheritanceCycles(t *testing.T) {
	c := newCatalogue(Project{})
	c.record("First", "fixture")["bases"] = []any{"Second"}
	c.record("Second", "fixture")["bases"] = []any{"First"}
	if _, err := c.finalize(t.Context()); err == nil || !strings.Contains(err.Error(), "cyclic inheritance") {
		t.Fatalf("inheritance cycle error = %v", err)
	}
}
