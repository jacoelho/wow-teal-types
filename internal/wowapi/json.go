package wowapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf16"
)

// The three formats are existing byte contracts: review digests, interned names,
// and human-readable reports. Native metadata separately retains literal UTF-8.
func jsonCompact(v any) ([]byte, error)    { return encodeJSON(v, false, false, true) }
func jsonDefault(v any) ([]byte, error)    { return encodeJSON(v, false, true, true) }
func jsonPretty(v any) ([]byte, error)     { return encodeJSON(v, true, false, true) }
func jsonPrettyUTF8(v any) ([]byte, error) { return encodeJSON(v, true, false, false) }
func digest(v any) (string, error) {
	b, e := jsonCompact(v)
	if e != nil {
		return "", e
	}
	return sha256hex(b), nil
}
func encodeJSON(v any, pretty, spaced, ascii bool) ([]byte, error) {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	if pretty {
		enc.SetIndent("", "  ")
	}
	normalized, err := pythonNumbers(reflect.ValueOf(v))
	if err != nil {
		return nil, err
	}
	if err := enc.Encode(normalized); err != nil {
		return nil, err
	}
	data := bytes.TrimSuffix(b.Bytes(), []byte("\n"))
	if spaced && !pretty {
		var out bytes.Buffer
		quoted, escaped := false, false
		for _, c := range data {
			out.WriteByte(c)
			if quoted {
				if escaped {
					escaped = false
				} else if c == '\\' {
					escaped = true
				} else if c == '"' {
					quoted = false
				}
			} else if c == '"' {
				quoted = true
			} else if c == ',' || c == ':' {
				out.WriteByte(' ')
			}
		}
		data = out.Bytes()
	}
	if ascii {
		var out bytes.Buffer
		for _, r := range string(data) {
			if r < 127 {
				out.WriteByte(byte(r))
			} else if r <= 0xffff {
				fmt.Fprintf(&out, "\\u%04x", r)
			} else {
				a, b := utf16.EncodeRune(r)
				fmt.Fprintf(&out, "\\u%04x\\u%04x", a, b)
			}
		}
		data = out.Bytes()
	}
	if pretty {
		data = append(data, '\n')
	}
	return data, nil
}

// Python source fingerprints distinguish an integer from an integral float.
// Keep the shortest round-tripping float with Python's fixed/exponent bounds.
func pythonNumbers(v reflect.Value) (any, error) {
	if !v.IsValid() {
		return nil, nil
	}
	if v.Kind() == reflect.Interface || v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil, nil
		}
		return pythonNumbers(v.Elem())
	}
	switch v.Kind() {
	case reflect.Float32, reflect.Float64:
		f := v.Float()
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return nil, fmt.Errorf("non-finite JSON number")
		}
		a := math.Abs(f)
		style := byte('f')
		if a != 0 && (a < 1e-4 || a >= 1e16) {
			style = 'e'
		}
		text := strconv.FormatFloat(f, style, -1, 64)
		if style == 'f' && !strings.Contains(text, ".") {
			text += ".0"
		}
		return json.Number(text), nil
	case reflect.Map:
		if v.IsNil() {
			return nil, nil
		}
		if v.Type().Key().Kind() != reflect.String {
			return nil, fmt.Errorf("JSON object keys must be strings")
		}
		m := map[string]any{}
		iter := v.MapRange()
		for iter.Next() {
			value, e := pythonNumbers(iter.Value())
			if e != nil {
				return nil, e
			}
			m[iter.Key().String()] = value
		}
		return m, nil
	case reflect.Slice, reflect.Array:
		if v.Kind() == reflect.Slice && v.IsNil() {
			return nil, nil
		}
		r := make([]any, v.Len())
		for i := range r {
			value, e := pythonNumbers(v.Index(i))
			if e != nil {
				return nil, e
			}
			r[i] = value
		}
		return r, nil
	default:
		return v.Interface(), nil
	}
}
