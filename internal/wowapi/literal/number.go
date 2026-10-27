package literal

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Float keeps Python's historical shortest float spelling, including 1.0 and
// -0.0. Known numeric DTO fields use this scalar rather than a raw Value.
type Float float64

func (v Float) MarshalJSON() ([]byte, error) {
	if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
		return nil, fmt.Errorf("non-finite JSON number")
	}
	style := byte('f')
	a := math.Abs(float64(v))
	if a != 0 && (a < 1e-4 || a >= 1e16) {
		style = 'e'
	}
	text := strconv.FormatFloat(float64(v), style, -1, 64)
	if style == 'f' && !strings.Contains(text, ".") {
		text += ".0"
	}
	return []byte(text), nil
}
func (v *Float) UnmarshalJSON(data []byte) error {
	var value float64
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return fmt.Errorf("non-finite JSON number")
	}
	*v = Float(value)
	return nil
}

// Numeric preserves the integer-versus-float form of a known numeric field.
// It cannot contain a string, collection, or arbitrary source value.
type Numeric struct {
	integer  string
	floating Float
	isFloat  bool
}

func IntegerNumber(value int64) Numeric { return Numeric{integer: strconv.FormatInt(value, 10)} }
func FloatNumber(value float64) (Numeric, error) {
	if _, err := Float(value).MarshalJSON(); err != nil {
		return Numeric{}, err
	}
	return Numeric{floating: Float(value), isFloat: true}, nil
}
func (n Numeric) MarshalJSON() ([]byte, error) {
	if n.isFloat {
		return n.floating.MarshalJSON()
	}
	if n.integer == "" {
		return []byte("0"), nil
	}
	return []byte(n.integer), nil
}
func (n *Numeric) UnmarshalJSON(data []byte) error {
	value, err := ParseJSON(data)
	if err != nil {
		return err
	}
	if integer, ok := value.Integer(); ok {
		*n = Numeric{integer: integer}
		return nil
	}
	if number, ok := value.Number(); ok {
		*n = Numeric{floating: Float(number), isFloat: true}
		return nil
	}
	return fmt.Errorf("expected JSON number")
}
func (n Numeric) Float64() (float64, error) {
	if n.isFloat {
		return float64(n.floating), nil
	}
	text := n.integer
	if text == "" {
		text = "0"
	}
	return strconv.ParseFloat(text, 64)
}
func (n Numeric) Integer() (string, bool) {
	if n.isFloat {
		return "", false
	}
	if n.integer == "" {
		return "0", true
	}
	return n.integer, true
}
