package wowapi

import (
	"math"
	"testing"
)

func TestReviewJSONByteContracts(t *testing.T) {
	value := M{"empty": []any{}, "float": float64(1), "negative_zero": math.Copysign(0, -1), "text": "©😀<\x7f", "none": nil}
	compact, e := jsonCompact(value)
	if e != nil {
		t.Fatal(e)
	}
	expected := `{"empty":[],"float":1.0,"negative_zero":-0.0,"none":null,"text":"\u00a9\ud83d\ude00<\u007f"}`
	if string(compact) != expected {
		t.Fatalf("compact = %s; want %s", compact, expected)
	}
	spaced, e := jsonDefault([]any{"a,b:c", float64(1000000), float64(0.00001), float64(1e16)})
	if e != nil {
		t.Fatal(e)
	}
	if string(spaced) != `["a,b:c", 1000000.0, 1e-05, 1e+16]` {
		t.Fatalf("interning JSON: %s", spaced)
	}
	pretty, e := jsonPretty(M{"a": false, "b": []any{}})
	if e != nil {
		t.Fatal(e)
	}
	if string(pretty) != "{\n  \"a\": false,\n  \"b\": []\n}\n" {
		t.Fatalf("report JSON: %q", pretty)
	}
}
