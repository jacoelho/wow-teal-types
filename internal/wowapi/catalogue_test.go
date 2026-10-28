package wowapi

import (
	"strconv"
	"strings"
	"testing"

	"wow-teal-types/internal/sourcelock"
)

func TestFrameXMLMethodContracts(t *testing.T) {
	for _, test := range []struct {
		name, expression string
		valid            bool
	}{
		{"primitive", "number", false},
		{"bare function", "function", false},
		{"missing receiver", "fun():number", false},
		{"implicit receiver", "fun(self):number", false},
		{"wrong receiver", "fun(self:Other):number", false},
		{"callable union", "fun(self:Widget)|number", false},
		{"owning receiver", "fun(self:Widget):number", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			var supplement frameXMLSupplement
			input := `{"classes":[{"name":"Widget","methods":{"Read":` + strconv.Quote(test.expression) + `}}]}`
			if err := sourcelock.DecodeStrict([]byte(input), &supplement); err != nil {
				t.Fatal(err)
			}
			c := newCatalogue()
			err := c.loadFrameXML(supplement)
			if !test.valid {
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
			if !strings.Contains(declarations, "Read: function(self): number") {
				t.Fatalf("owning receiver missing from declarations:\n%s", declarations)
			}
		})
	}
}
