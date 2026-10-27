package wowapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"unicode/utf16"

	"wow-teal-types/internal/wowapi/literal"
)

// The three formats are existing byte contracts: review digests, interned names,
// and human-readable reports. Native metadata separately retains literal UTF-8.
func jsonCompact[T any](v T) ([]byte, error)    { return encodeJSON(v, false, false, true) }
func jsonDefault[T any](v T) ([]byte, error)    { return encodeJSON(v, false, true, true) }
func jsonPretty[T any](v T) ([]byte, error)     { return encodeJSON(v, true, false, true) }
func jsonPrettyUTF8[T any](v T) ([]byte, error) { return encodeJSON(v, true, false, false) }
func digest[T any](v T) (string, error) {
	b, e := jsonCompact(v)
	if e != nil {
		return "", e
	}
	return sha256hex(b), nil
}
func encodeJSON[T any](v T, pretty, spaced, ascii bool) ([]byte, error) {
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
	if pretty {
		var indented bytes.Buffer
		if err = json.Indent(&indented, data, "", "  "); err != nil {
			return nil, err
		}
		data = indented.Bytes()
	}
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
