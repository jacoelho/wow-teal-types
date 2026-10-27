package wowapi

import (
	"math"
	"testing"
	"wow-teal-types/internal/wowapi/literal"
)

func TestReviewJSONByteContracts(t *testing.T) {
	value := struct {
		Empty        []string      `json:"empty"`
		Float        literal.Float `json:"float"`
		NegativeZero literal.Float `json:"negative_zero"`
		Text         string        `json:"text"`
		None         *string       `json:"none"`
	}{[]string{}, 1, literal.Float(math.Copysign(0, -1)), "©😀<\x7f", nil}
	compact, e := jsonCompact(value)
	if e != nil {
		t.Fatal(e)
	}
	expected := `{"empty":[],"float":1.0,"negative_zero":-0.0,"none":null,"text":"\u00a9\ud83d\ude00<\u007f"}`
	if string(compact) != expected {
		t.Fatalf("compact = %s; want %s", compact, expected)
	}
	spaced, e := jsonDefault(literal.List([]literal.Value{literal.Text("a,b:c"), mustLiteral(literal.Number(1000000)), mustLiteral(literal.Number(0.00001)), mustLiteral(literal.Number(1e16))}))
	if e != nil {
		t.Fatal(e)
	}
	if string(spaced) != `["a,b:c", 1000000.0, 1e-05, 1e+16]` {
		t.Fatalf("interning JSON: %s", spaced)
	}
	pretty, e := jsonPretty(struct {
		A bool     `json:"a"`
		B []string `json:"b"`
	}{false, []string{}})
	if e != nil {
		t.Fatal(e)
	}
	if string(pretty) != "{\n  \"a\": false,\n  \"b\": []\n}\n" {
		t.Fatalf("report JSON: %q", pretty)
	}
}
