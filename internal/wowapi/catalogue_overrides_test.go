package wowapi

import (
	"strings"
	"testing"
)

func TestFrameFactoryRejectsUnreviewedConfiguration(t *testing.T) {
	backdrop := frameTemplateOverride{Name: "BackdropTemplate", Type: "BackdropTemplate", FrameTypes: []string{"Button"}}
	for _, test := range []struct {
		name    string
		factory frameFactoryOverride
		reject  string
	}{
		{"empty", frameFactoryOverride{}, "lists no frame types"},
		{"not a frame", frameFactoryOverride{FrameTypes: []string{"Texture"}}, "Texture does not extend Frame"},
		{"undefined template type", frameFactoryOverride{FrameTypes: []string{"Button"}, Templates: []frameTemplateOverride{{Name: "Missing", Type: "Missing", FrameTypes: []string{"Button"}}}}, "type Missing is not defined"},
		{"template frame type not listed", frameFactoryOverride{FrameTypes: []string{"Frame"}, Templates: []frameTemplateOverride{backdrop}}, "unknown frame type Button"},
		{"synthetic name collides", frameFactoryOverride{FrameTypes: []string{"Button"}, Templates: []frameTemplateOverride{{Name: "Template", Type: "BackdropTemplate", FrameTypes: []string{"Button"}}}}, "Button_Template collides"},
	} {
		t.Run(test.name, func(t *testing.T) {
			c := newCatalogue()
			c.record("Frame", "fixture", kindRecord)
			c.record("Button", "fixture", kindRecord).Bases = []string{"Frame"}
			c.record("Texture", "fixture", kindRecord)
			c.record("BackdropTemplate", "fixture", kindInterface)
			c.record("Button_Template", "fixture", kindInterface)
			err := attempt(func() { c.applyFrameFactory(test.factory) })
			if err == nil || !strings.Contains(err.Error(), test.reject) {
				t.Fatalf("error = %v, want %q", err, test.reject)
			}
		})
	}
}

func TestFormattedTextArgumentsShareTheMethodSecretPermission(t *testing.T) {
	c := newCatalogue()
	setFormattedText := func(secretArguments string) *catalogueFunction {
		return &catalogueFunction{Method: true, Native: &nativeFunction{
			Arguments: []nativeParameter{{Name: "text", Type: "string"}},
			Policy:    nativePolicy{SecretArguments: new(secretArguments)},
		}}
	}
	c.record("Label", "fixture", kindRecord).Methods["SetFormattedText"] = setFormattedText("AllowedWhenTainted")
	c.record("Clickable", "fixture", kindRecord).Methods["SetFormattedText"] = setFormattedText("NotAllowed")
	c.applyFormattedText()

	projection, err := c.finalize(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	outputs := projection.render()
	declarations := string(outputs["wow.d.tl"])
	for _, expected := range []string{
		"SetFormattedText: function(self, W.MaybeSecret<string>, W.MaybeSecret<string> | W.MaybeSecret<number>...)",
		"SetFormattedText: function(self, string, string | number...)",
	} {
		if !strings.Contains(declarations, expected) {
			t.Errorf("missing %q in:\n%s", expected, declarations)
		}
	}
	// The vararg type and the checker policy must agree: only the method whose
	// fixed argument accepts a secret grants secret format arguments.
	if policy := string(outputs["call_policy.lua"]); strings.Count(policy, "secret_vararg_from = 3") != 1 {
		t.Errorf("want one secret vararg grant:\n%s", policy)
	}
}

func TestFormattedTextRejectsChangedNativeContracts(t *testing.T) {
	for _, test := range []struct {
		name   string
		method *catalogueFunction
		reject string
	}{
		{"no owner", nil, "no SetFormattedText method found"},
		{"annotation only", &catalogueFunction{Method: true, Annotation: &annotationFunction{}}, "has no native contract"},
		{"extra argument", &catalogueFunction{Method: true, Native: &nativeFunction{Arguments: []nativeParameter{{Name: "text", Type: "string"}, {Name: "extra", Type: "number"}}}}, "native arguments changed"},
		{"renamed argument", &catalogueFunction{Method: true, Native: &nativeFunction{Arguments: []nativeParameter{{Name: "format", Type: "string"}}}}, "native arguments changed"},
		{"already variadic", &catalogueFunction{Method: true, Native: &nativeFunction{Arguments: []nativeParameter{{Name: "text", Type: "string", StrideIndex: new(1)}}}}, "native arguments changed"},
	} {
		t.Run(test.name, func(t *testing.T) {
			c := newCatalogue()
			label := c.record("Label", "fixture", kindRecord)
			if test.method != nil {
				label.Methods["SetFormattedText"] = test.method
			}
			err := attempt(c.applyFormattedText)
			if err == nil || !strings.Contains(err.Error(), test.reject) {
				t.Fatalf("error = %v, want %q", err, test.reject)
			}
		})
	}
}
