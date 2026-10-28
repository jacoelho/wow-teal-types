package wowapi

import (
	"slices"
	"strings"
	"testing"
)

// literalFactoryFixture defines a frame factory global and an animation
// factory method whose valid configurations the rejection rows each break once.
func literalFactoryFixture() *catalogue {
	c := newCatalogue()
	c.record("Frame", "fixture", kindRecord)
	c.record("Button", "fixture", kindRecord).Bases = []string{"Frame"}
	c.record("Texture", "fixture", kindRecord)
	c.record("BackdropTemplate", "fixture", kindInterface)
	c.record("Button_Taken", "fixture", kindInterface)
	c.globals["CreateFrame"] = &catalogueFunction{Source: "fixture", ResolvedSignature: new("function(string): Frame")}
	c.record("Animation", "fixture", kindRecord)
	c.record("Alpha", "fixture", kindRecord).Bases = []string{"Animation"}
	c.record("AnimationGroup", "fixture", kindRecord).Methods["CreateAnimation"] = &catalogueFunction{Source: "fixture", Method: true, ResolvedSignature: new("function(self, string): Animation")}
	return c
}

func TestLiteralFactoriesRejectUnreviewedConfiguration(t *testing.T) {
	valid := map[string]literalFactoryOverride{
		"global": {Symbol: "CreateFrame", Base: "Frame", Overload: "function({type}, {template}): {result}", Types: []string{"Frame", "Button"},
			Templates: []literalTemplateOverride{{Name: "BackdropTemplate", Type: "BackdropTemplate", Types: []string{"Button"}}}},
		"method": {Symbol: "AnimationGroup.CreateAnimation", Base: "Animation", Overload: "function(self, {type}, {template}): {result}", Types: []string{"Animation", "Alpha"},
			Templates: []literalTemplateOverride{{Name: "BackdropTemplate", Type: "BackdropTemplate", Types: []string{"Alpha"}}}},
	}
	for _, test := range []struct {
		name   string
		break_ func(*literalFactoryOverride)
		reject string
	}{
		{"no types", func(f *literalFactoryOverride) { f.Types, f.Templates = nil, nil }, "lists no types"},
		{"type outside base", func(f *literalFactoryOverride) { f.Types = append(f.Types, "Texture") }, "type Texture does not extend"},
		{"type listed twice", func(f *literalFactoryOverride) { f.Types = append(f.Types, f.Types[0]) }, "twice"},
		{"undefined template type", func(f *literalFactoryOverride) { f.Templates[0].Type = "Missing" }, "type Missing is not defined"},
		{"template lists unlisted type", func(f *literalFactoryOverride) { f.Templates[0].Types = []string{"Texture"} }, "lists unknown type Texture"},
		{"synthetic name collides", func(f *literalFactoryOverride) {
			f.Types, f.Templates[0].Name, f.Templates[0].Types = append(f.Types, "Button"), "Taken", []string{"Button"}
		}, "Button_Taken collides"},
		{"overload lacks type", func(f *literalFactoryOverride) { f.Overload = strings.Replace(f.Overload, "{type}", "string", 1) }, "lacks a placeholder"},
		{"overload lacks result", func(f *literalFactoryOverride) { f.Overload = strings.Replace(f.Overload, "{result}", "Frame", 1) }, "lacks a placeholder"},
		{"overload lacks template", func(f *literalFactoryOverride) { f.Overload = strings.Replace(f.Overload, "{template}", "string", 1) }, "lacks a placeholder"},
		{"unknown symbol", func(f *literalFactoryOverride) { f.Symbol += "Missing" }, "missing required"},
	} {
		for kind, factory := range valid {
			t.Run(test.name+"/"+kind, func(t *testing.T) {
				factory.Types = slices.Clone(factory.Types)
				factory.Templates = []literalTemplateOverride{factory.Templates[0]}
				factory.Templates[0].Types = slices.Clone(factory.Templates[0].Types)
				test.break_(&factory)
				err := attempt(func() { literalFactoryFixture().applyLiteralFactories([]literalFactoryOverride{factory}) })
				if err == nil || !strings.Contains(err.Error(), test.reject) {
					t.Fatalf("error = %v, want %q", err, test.reject)
				}
			})
		}
	}
	t.Run("symbol declared twice", func(t *testing.T) {
		err := attempt(func() { literalFactoryFixture().applyLiteralFactories([]literalFactoryOverride{valid["global"], valid["global"]}) })
		if err == nil || !strings.Contains(err.Error(), "declared twice") {
			t.Fatalf("error = %v, want duplicate rejection", err)
		}
	})
}

// Teal selects the first matching overload, so order is part of the contract:
// template literals, then type literals in listed order, then the fallback.
// Every overload keeps the factory's native permission.
func TestLiteralFactoryOverloadsKeepOrderAndNativePolicy(t *testing.T) {
	for _, test := range []struct {
		name, generic, fallback string
	}{
		{"source fallback", "", "function(self, W.MaybeSecret<string>): Widget"},
		{"reviewed generic fallback", "function(self, string, ? string): Widget", "function(self, string, ? string): Widget"},
	} {
		t.Run(test.name, func(t *testing.T) {
			c := newCatalogue()
			c.record("Widget", "fixture", kindRecord)
			c.record("A", "fixture", kindRecord).Bases = []string{"Widget"}
			c.record("B", "fixture", kindRecord).Bases = []string{"Widget"}
			c.record("Tpl", "fixture", kindInterface)
			c.record("Factory", "fixture", kindRecord).Methods["Make"] = &catalogueFunction{Source: "fixture", Method: true, Native: &nativeFunction{
				Arguments: []nativeParameter{{Name: "kind", Type: "string"}},
				Returns:   []nativeParameter{{Name: "made", Type: "Widget"}},
				Policy:    nativePolicy{SecretArguments: new("AllowedWhenTainted")},
			}}
			c.applyLiteralFactories([]literalFactoryOverride{{
				Symbol: "Factory.Make", Base: "Widget", Generic: test.generic,
				Overload: "function(self, {type}, {template}): {result}", Types: []string{"A", "B"},
				Templates: []literalTemplateOverride{{Name: "Tpl", Type: "Tpl", Types: []string{"B"}}},
			}})
			projection, err := c.finalize(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			outputs := projection.render()
			var got []string
			for line := range strings.SplitSeq(string(outputs["wow.d.tl"]), "\n") {
				if signature, ok := strings.CutPrefix(line, "      Make: "); ok {
					got = append(got, signature)
				}
			}
			literal := func(value string) string { return c.literal([]string{value}) }
			want := []string{
				"function(self, " + literal("B") + ", " + literal("Tpl") + "): B_Tpl",
				"function(self, " + literal("A") + ", ? string): A",
				"function(self, " + literal("B") + ", ? string): B",
				test.fallback,
			}
			if !slices.Equal(got, want) {
				t.Fatalf("overloads = %q, want %q", got, want)
			}
			if policy := string(outputs["call_policy.lua"]); strings.Count(policy, `symbol = "Factory.Make"`) != len(want) || strings.Count(policy, "secret_parameters = {2}") != len(want) {
				t.Errorf("want %d Factory.Make entries granting argument 2:\n%s", len(want), policy)
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
