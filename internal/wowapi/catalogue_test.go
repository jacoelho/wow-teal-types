package wowapi

import (
	"strconv"
	"strings"
	"testing"

	"wow-teal-types/internal/sourcelock"
)

func TestFrameXMLMethodContracts(t *testing.T) {
	// want is the rendered method; empty means the method is rejected.
	for _, test := range []struct{ name, expression, want string }{
		{"primitive", "number", ""},
		{"bare function", "function", ""},
		{"missing receiver", "fun():number", ""},
		{"implicit receiver", "fun(self):number", ""},
		{"wrong receiver", "fun(self:Other):number", ""},
		{"callable union", "fun(self:Widget)|number", ""},
		{"owning receiver", "fun(self:Widget):number", "Read: function(self): number\n"},
		{"several results", "fun(self:Widget):number,string", "Read: function(self): number, string\n"},
		{"callback parameter", "fun(self:Widget, cb:(fun():string), n:number):number", "Read: function(self, function(): (string), number): number\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var supplement frameXMLSupplement
			input := `{"classes":[{"name":"Widget","methods":{"Read":` + strconv.Quote(test.expression) + `}}]}`
			if err := sourcelock.DecodeStrict([]byte(input), &supplement); err != nil {
				t.Fatal(err)
			}
			c := newCatalogue()
			err := c.loadFrameXML(supplement)
			if test.want == "" {
				if err == nil || !strings.Contains(err.Error(), "Widget.Read:") {
					t.Fatalf("invalid method error = %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			projection, err := c.finalize(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			declarations := string(projection.render()["wow.d.tl"])
			if !strings.Contains(declarations, test.want) {
				t.Fatalf("missing %q from declarations:\n%s", test.want, declarations)
			}
		})
	}
}

// A supplement global declares a whole signature, so its results are not a
// grouped result list.
func TestFrameXMLGlobalsDeclareWholeSignatures(t *testing.T) {
	c := newCatalogue()
	if err := c.loadFrameXML(frameXMLSupplement{Globals: map[string]string{"Measure": "fun(a:number):string,number"}}); err != nil {
		t.Fatal(err)
	}
	projection, err := c.finalize(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if got := string(projection.render()["wow_env.d.tl"]); !strings.Contains(got, "global Measure: function(number): string, number\n") {
		t.Fatalf("supplement global is not a whole signature:\n%s", got)
	}
}
