package wowapi

import (
	"bytes"
	"strings"
	"testing"
)

func TestFinalizedDeclarationsResolveGeneratedTypesInheritanceAndSecrecy(t *testing.T) {
	c := newCatalogue(Project{})
	base := c.record("Base", "fixture")
	base.Fields["count"] = &catalogueField{Native: &NativeParameter{Type: "number"}}
	base.Methods["Get"] = &catalogueFunction{Method: true, Native: &NativeFunction{Returns: []NativeParameter{{Type: "number"}}}}
	child := c.record("Child", "fixture")
	child.Bases = []string{"Base"}
	child.Fields["count"] = &catalogueField{Type: "number"}
	child.Methods["Get"] = &catalogueFunction{Annotation: &AnnotationFunction{Returns: []AnnotationReturn{{Type: "string"}}}}
	payload := c.record("Payload", "fixture")
	payload.Fields["count"] = &catalogueField{Native: &NativeParameter{Type: "number"}}
	payload.Fields["visible"] = &catalogueField{Native: &NativeParameter{Type: "boolean", NeverSecret: boolPointer(true)}}
	holder := c.record("Holder", "fixture")
	holder.Fields["left"] = &catalogueField{Type: "{value:number}"}
	holder.Fields["right"] = &catalogueField{Type: "{value:number}"}
	holder.Fields["callback"] = &catalogueField{Type: "fun(value:number):string"}
	c.setGlobal("GetPayload", &catalogueFunction{Source: "fixture", Native: &NativeFunction{Returns: []NativeParameter{{Type: "Payload", SecretValue: boolPointer(true)}}}})

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
	payload.Fields["count"] = &catalogueField{Native: &NativeParameter{Type: "string"}}
	c.globals["GetPayload"] = &catalogueFunction{ResolvedSignature: stringPointer("function(): string")}
	for name, expected := range outputs {
		if !bytes.Equal(projection.render()[name], expected) {
			t.Errorf("source mutation changed finalized %s", name)
		}
	}
}

func TestFinalizationRejectsInheritanceCycles(t *testing.T) {
	c := newCatalogue(Project{})
	c.record("First", "fixture").Bases = []string{"Second"}
	c.record("Second", "fixture").Bases = []string{"First"}
	if _, err := c.finalize(t.Context()); err == nil || !strings.Contains(err.Error(), "cyclic inheritance") {
		t.Fatalf("inheritance cycle error = %v", err)
	}
}
