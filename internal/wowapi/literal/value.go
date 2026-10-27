// Package literal retains source syntax whose fields are not known until review.
// Interpreted API contracts use concrete types outside this package.
package literal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"slices"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Kind uint8

const (
	Null Kind = iota
	Boolean
	String
	Integer
	FloatKind
	Array
	ObjectKind
	Reference
	Expression
)

// Object is the keyed syntax of a raw evidence document, never a domain model.
type Object map[string]Value

// Value is a closed raw literal. Its zero value is null. Constructors and accessors
// copy containers, so a retained source value cannot change through an alias.
type Value struct {
	kind        Kind
	text        string
	boolean     bool
	number      float64
	array       []Value
	object      Object
	left, right *Value
}

func Bool(v bool) Value   { return Value{kind: Boolean, boolean: v} }
func Text(v string) Value { return Value{kind: String, text: v} }
func Int(v int64) Value   { return Value{kind: Integer, text: strconv.FormatInt(v, 10)} }
func BigInteger(decimal string) (Value, error) {
	if len(strings.TrimPrefix(decimal, "-")) > 4300 {
		return Value{}, fmt.Errorf("literal integer exceeds 4300 digits")
	}
	n, ok := new(big.Int).SetString(decimal, 10)
	if !ok {
		return Value{}, fmt.Errorf("invalid literal integer %q", decimal)
	}
	return Value{kind: Integer, text: n.String()}, nil
}
func Number(v float64) (Value, error) {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return Value{}, fmt.Errorf("non-finite literal number")
	}
	return Value{kind: FloatKind, number: v}, nil
}
func List(v []Value) Value { return Value{kind: Array, array: slices.Clone(v)} }
func Record(v Object) Value {
	copy := make(Object, len(v))
	for key, value := range v {
		copy[key] = value
	}
	return Value{kind: ObjectKind, object: copy}
}
func Ref(name string) (Value, error) {
	for _, part := range strings.Split(name, ".") {
		if part == "" {
			return Value{}, fmt.Errorf("invalid literal reference %q", name)
		}
		for i, r := range part {
			if !(r == '_' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || i > 0 && r >= '0' && r <= '9') {
				return Value{}, fmt.Errorf("invalid literal reference %q", name)
			}
		}
	}
	return Value{kind: Reference, text: name}, nil
}
func Binary(operator string, left, right Value) (Value, error) {
	if operator != "+" && operator != "-" {
		return Value{}, fmt.Errorf("unsupported literal operator %q", operator)
	}
	for _, value := range []Value{left, right} {
		if value.kind != Integer && value.kind != FloatKind && value.kind != Reference && value.kind != Expression {
			return Value{}, fmt.Errorf("literal arithmetic requires numbers or named references")
		}
	}
	return Value{kind: Expression, text: operator, left: &left, right: &right}, nil
}
func (v Value) Kind() Kind              { return v.kind }
func (v Value) Bool() (bool, bool)      { return v.boolean, v.kind == Boolean }
func (v Value) Text() (string, bool)    { return v.text, v.kind == String }
func (v Value) Integer() (string, bool) { return v.text, v.kind == Integer }
func (v Value) Number() (float64, bool) { return v.number, v.kind == FloatKind }
func (v Value) Array() ([]Value, bool)  { return slices.Clone(v.array), v.kind == Array }
func (v Value) Object() (Object, bool) {
	if v.kind != ObjectKind {
		return nil, false
	}
	out := make(Object, len(v.object))
	for key, value := range v.object {
		out[key] = value
	}
	return out, true
}
func (v Value) Reference() (string, bool) { return v.text, v.kind == Reference }
func (v Value) Expression() (string, Value, Value, bool) {
	if v.kind != Expression {
		return "", Value{}, Value{}, false
	}
	return v.text, *v.left, *v.right, true
}
func (v Value) Equal(other Value) bool {
	if v.kind != other.kind {
		return false
	}
	switch v.kind {
	case Null:
		return true
	case Boolean:
		return v.boolean == other.boolean
	case String, Integer, Reference:
		return v.text == other.text
	case FloatKind:
		return math.Float64bits(v.number) == math.Float64bits(other.number)
	case Array:
		return slices.EqualFunc(v.array, other.array, Value.Equal)
	case ObjectKind:
		if len(v.object) != len(other.object) {
			return false
		}
		for key, value := range v.object {
			right, ok := other.object[key]
			if !ok || !value.Equal(right) {
				return false
			}
		}
		return true
	case Expression:
		return v.text == other.text && v.left.Equal(*other.left) && v.right.Equal(*other.right)
	}
	return false
}

// MarshalJSON retains the historical integer/float distinction and symbolic Lua
// representation. Formatting and ASCII escaping belong to the enclosing codec.
func (v Value) MarshalJSON() ([]byte, error) {
	var out bytes.Buffer
	if err := v.writeJSON(&out, 0); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
func writeText(out *bytes.Buffer, text string) error {
	if !utf8.ValidString(text) {
		return fmt.Errorf("invalid UTF-8 literal string")
	}
	var encoded bytes.Buffer
	encoder := json.NewEncoder(&encoded)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(text); err != nil {
		return err
	}
	out.Write(bytes.TrimSuffix(encoded.Bytes(), []byte{'\n'}))
	return nil
}
func (v Value) writeJSON(out *bytes.Buffer, depth int) error {
	if depth > 64 {
		return fmt.Errorf("literal nesting exceeds 64")
	}
	switch v.kind {
	case Null:
		out.WriteString("null")
	case Boolean:
		out.WriteString(strconv.FormatBool(v.boolean))
	case String:
		return writeText(out, v.text)
	case Integer:
		out.WriteString(v.text)
	case FloatKind:
		data, err := Float(v.number).MarshalJSON()
		if err != nil {
			return err
		}
		out.Write(data)
	case Array:
		out.WriteByte('[')
		for i, value := range v.array {
			if i > 0 {
				out.WriteByte(',')
			}
			if err := value.writeJSON(out, depth+1); err != nil {
				return err
			}
		}
		out.WriteByte(']')
	case ObjectKind:
		keys := make([]string, 0, len(v.object))
		for key := range v.object {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		out.WriteByte('{')
		for i, key := range keys {
			if i > 0 {
				out.WriteByte(',')
			}
			if err := writeText(out, key); err != nil {
				return err
			}
			out.WriteByte(':')
			if err := v.object[key].writeJSON(out, depth+1); err != nil {
				return err
			}
		}
		out.WriteByte('}')
	case Reference:
		out.WriteString(`{"$ref":`)
		if err := writeText(out, v.text); err != nil {
			return err
		}
		out.WriteByte('}')
	case Expression:
		out.WriteString(`{"$expression":[`)
		if err := v.left.writeJSON(out, depth+1); err != nil {
			return err
		}
		out.WriteByte(',')
		if err := writeText(out, v.text); err != nil {
			return err
		}
		out.WriteByte(',')
		if err := v.right.writeJSON(out, depth+1); err != nil {
			return err
		}
		out.WriteString("]}")
	default:
		return fmt.Errorf("invalid literal kind %d", v.kind)
	}
	return nil
}

func (v *Value) UnmarshalJSON(data []byte) error {
	value, err := ParseJSON(data)
	if err != nil {
		return err
	}
	*v = value
	return nil
}
