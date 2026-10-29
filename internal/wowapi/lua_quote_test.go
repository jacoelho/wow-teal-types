package wowapi

import (
	"strings"
	"testing"
)

func TestLuaQuoteProducesLua51StringLiterals(t *testing.T) {
	cases := []struct {
		name, input, want string
	}{
		{"double quote", `a"b`, `"a\"b"`},
		{"backslash", `\`, `"\\"`},
		{"newline", "\n", `"\010"`},
		{"carriage return", "\r", `"\013"`},
		{"NUL before digit", "\x001", `"\0001"`},
		{"low control byte", "\x01", `"\001"`},
		{"high control byte", "\x1f", `"\031"`},
		{"DEL", "\x7f", `"\127"`},
		{"two-byte UTF-8", "©", `"\194\169"`},
		{"four-byte UTF-8", "😀", `"\240\159\152\128"`},
		{"line separator", " ", `"\226\128\168"`},
		{"long-bracket and comment markers", "]]--", `"]]--"`},
		{"invalid UTF-8", "\xff", `"\255"`},
		{"plain ASCII", "Enum.Foo bar", `"Enum.Foo bar"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := luaQuote(tc.input)
			if got != tc.want {
				t.Fatalf("luaQuote(%q) = %s, want %s", tc.input, got, tc.want)
			}
			if strings.Contains(got, `\u`) {
				t.Fatalf("luaQuote(%q) = %s contains a \\u escape", tc.input, got)
			}
		})
	}
}
