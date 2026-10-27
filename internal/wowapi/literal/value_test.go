package literal

import (
	"math"
	"testing"
)

func TestLiteralKindsAndCanonicalNumbers(t *testing.T) {
	for _, test := range []struct {
		source  string
		kind    Kind
		encoded string
	}{
		{"null", Null, "null"}, {"false", Boolean, "false"}, {"0", Integer, "0"}, {"1.0", FloatKind, "1.0"}, {"-0.0", FloatKind, "-0.0"},
		{"[]", Array, "[]"}, {"{}", ObjectKind, "{}"}, {"1000000.0", FloatKind, "1000000.0"}, {"0.00001", FloatKind, "1e-05"}, {"1e16", FloatKind, "1e+16"},
		{"9007199254740993", Integer, "9007199254740993"},
	} {
		t.Run(test.source, func(t *testing.T) {
			value, err := ParseJSON([]byte(test.source))
			if err != nil {
				t.Fatal(err)
			}
			if value.Kind() != test.kind {
				t.Fatalf("kind = %v, want %v", value.Kind(), test.kind)
			}
			data, err := value.MarshalJSON()
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != test.encoded {
				t.Fatalf("JSON = %s, want %s", data, test.encoded)
			}
		})
	}
}
func TestLiteralReferencesAndExpressions(t *testing.T) {
	reference, err := Ref("Enum.Example.First")
	if err != nil {
		t.Fatal(err)
	}
	expression, err := Binary("+", reference, Int(1))
	if err != nil {
		t.Fatal(err)
	}
	data, err := expression.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if expected := `{"$expression":[{"$ref":"Enum.Example.First"},"+",1]}`; string(data) != expected {
		t.Fatalf("JSON = %s, want %s", data, expected)
	}
	for _, value := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		if _, err := Number(value); err == nil {
			t.Error("accepted nonfinite number")
		}
	}
	if _, err := Binary("+", Text("text"), Int(1)); err == nil {
		t.Error("accepted arithmetic on text")
	}
	if _, err := Ref("Enum..First"); err == nil {
		t.Error("accepted malformed reference")
	}
}
func TestJSONRejectsAmbiguousOrInvalidLiterals(t *testing.T) {
	for _, source := range []string{`{"a":1,"a":2}`, `[1,]`, `{"a":1,}`, `01`, `1e999`, `"\x20"`, `true false`} {
		if _, err := ParseJSON([]byte(source)); err == nil {
			t.Errorf("accepted %s", source)
		}
	}
}
