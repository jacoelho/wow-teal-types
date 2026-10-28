package wowapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"unicode/utf16"

	"wow-teal-types/internal/wowapi/literal"
)

// Stable field ordering and ASCII escaping preserve interned declaration names.
func jsonCompact(v any) ([]byte, error) { return encodeJSON(v, false) }
func jsonDefault(v any) ([]byte, error) { return encodeJSON(v, true) }
func encodeJSON(v any, spaced bool) ([]byte, error) {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	// The closed syntax tree is local to the codec. Domain inputs remain their
	// concrete types; sorting is a persisted-byte rule, including struct fields.
	syntax, err := literal.ParseJSON(b.Bytes())
	if err != nil {
		return nil, err
	}
	data, err := syntax.MarshalJSON()
	if err != nil {
		return nil, err
	}
	if spaced {
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
	{
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
	return data, nil
}

func decodeJSON(data []byte, value any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("trailing JSON data")
	}
	return nil
}
