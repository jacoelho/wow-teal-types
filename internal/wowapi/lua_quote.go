package wowapi

import "strconv"

// luaQuote renders s as a double-quoted Lua 5.1 string literal. Lua has no
// \u escape, so every control byte, DEL, and non-ASCII byte is written as a
// fixed-width decimal escape; the fixed width keeps a following digit from
// being read as part of the escape. Bytes are copied verbatim, so invalid
// UTF-8 survives unchanged.
func luaQuote(s string) string {
	out := make([]byte, 0, len(s)+2)
	out = append(out, '"')
	for i := 0; i < len(s); i++ {
		b := s[i]
		switch {
		case b == '"' || b == '\\':
			out = append(out, '\\', b)
		case b < 0x20 || b >= 0x7f:
			if b < 100 {
				out = append(out, '\\', '0')
				if b < 10 {
					out = append(out, '0')
				}
			} else {
				out = append(out, '\\')
			}
			out = strconv.AppendUint(out, uint64(b), 10)
		default:
			out = append(out, b)
		}
	}
	return string(append(out, '"'))
}
