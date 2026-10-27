package wowapi

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

var typeNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z_0-9.\-]*$`)
var numericTypePattern = regexp.MustCompile(`^-?(?:0x[0-9A-Fa-f]+|\d+(?:\.\d+)?)$`)
var genericTypePattern = regexp.MustCompile(`^([A-Za-z_][A-Za-z_0-9.\-]*)\s*<(.*)>$`)
var simpleFieldPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z_0-9]*\??$`)

func unsupported(raw, reason string) { reject("%q: %s", raw, reason) }
func SplitTopLevel(text string, delimiter byte) (parts []string, err error) {
	defer captureError(&err)
	return splitTop(text, delimiter), nil
}
func splitTop(text string, delimiter byte) []string {
	pairs := map[byte]byte{'(': ')', '[': ']', '{': '}', '<': '>'}
	stack := []byte{}
	quote := byte(0)
	escaped := false
	start := 0
	parts := []string{}
	for i := 0; i < len(text); i++ {
		c := text[i]
		if quote != 0 {
			if escaped {
				escaped = false
			} else if c == '\\' {
				escaped = true
			} else if c == quote {
				quote = 0
			}
			continue
		}
		if c == '"' || c == '\'' || c == '`' {
			quote = c
		} else if close, ok := pairs[c]; ok {
			stack = append(stack, close)
			if len(stack) > 64 {
				unsupported(text, "type nesting exceeds 64 levels")
			}
		} else if strings.ContainsRune(")]}>", rune(c)) {
			if len(stack) == 0 || stack[len(stack)-1] != c {
				unsupported(text, "unbalanced type delimiters")
			}
			stack = stack[:len(stack)-1]
		} else if c == delimiter && len(stack) == 0 {
			parts = append(parts, strings.TrimSpace(text[start:i]))
			start = i + 1
		}
	}
	if quote != 0 || len(stack) > 0 {
		unsupported(text, "unterminated type literal or delimiter")
	}
	return append(parts, strings.TrimSpace(text[start:]))
}
func matchingClose(text string, opening int) int {
	pairs := map[byte]byte{'(': ')', '[': ']', '{': '}', '<': '>'}
	if opening < 0 || opening >= len(text) {
		unsupported(text, "missing type delimiter")
	}
	stack := []byte{pairs[text[opening]]}
	quote := byte(0)
	escaped := false
	for i := opening + 1; i < len(text); i++ {
		c := text[i]
		if quote != 0 {
			if escaped {
				escaped = false
			} else if c == '\\' {
				escaped = true
			} else if c == quote {
				quote = 0
			}
		} else if c == '"' || c == '\'' || c == '`' {
			quote = c
		} else if close, ok := pairs[c]; ok {
			stack = append(stack, close)
			if len(stack) > 64 {
				unsupported(text, "type nesting exceeds 64 levels")
			}
		} else if strings.ContainsRune(")]}>", rune(c)) {
			if len(stack) == 0 || stack[len(stack)-1] != c {
				unsupported(text, "unbalanced type delimiters")
			}
			stack = stack[:len(stack)-1]
			if len(stack) == 0 {
				return i
			}
		}
	}
	unsupported(text, "unterminated type delimiter")
	return 0
}
func quotedLiteral(text string) (string, bool) {
	if len(text) == 0 || text[0] != '"' && text[0] != '\'' {
		return "", false
	}
	if len(text) < 2 || text[len(text)-1] != text[0] {
		unsupported(text, "unsupported quoted literal escape")
	}
	body := text[1 : len(text)-1]
	var out strings.Builder
	for i := 0; i < len(body); i++ {
		c := body[i]
		if c != '\\' {
			if c == text[0] || c == '\n' || c == '\r' {
				unsupported(text, "unsupported quoted literal escape")
			}
			out.WriteByte(c)
			continue
		}
		i++
		if i >= len(body) {
			unsupported(text, "unsupported quoted literal escape")
		}
		c = body[i]
		switch c {
		case 'n':
			out.WriteByte('\n')
		case 'r':
			out.WriteByte('\r')
		case 't':
			out.WriteByte('\t')
		case 'a':
			out.WriteByte('\a')
		case 'b':
			out.WriteByte('\b')
		case 'f':
			out.WriteByte('\f')
		case 'v':
			out.WriteByte('\v')
		case '\\', '\'', '"':
			out.WriteByte(c)
		case '\n':
		case 'x', 'u', 'U':
			size := 2
			if c == 'u' {
				size = 4
			}
			if c == 'U' {
				size = 8
			}
			if i+size >= len(body) {
				unsupported(text, "unsupported quoted literal escape")
			}
			n, e := strconv.ParseUint(body[i+1:i+1+size], 16, 32)
			if e != nil || !utf8.ValidRune(rune(n)) {
				unsupported(text, "unsupported quoted literal escape")
			}
			out.WriteRune(rune(n))
			i += size
		default:
			if c >= '0' && c <= '7' {
				start := i
				for i+1 < len(body) && i-start < 2 && body[i+1] >= '0' && body[i+1] <= '7' {
					i++
				}
				n, _ := strconv.ParseInt(body[start:i+1], 8, 32)
				out.WriteRune(rune(n))
			} else {
				out.WriteByte('\\')
				out.WriteByte(c)
			}
		}
	}
	return out.String(), true
}
func uniqueStrings(values []string) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, v := range values {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}
func containsString(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}
func toAnyStrings(values []string) []any {
	out := make([]any, len(values))
	for i, s := range values {
		out[i] = s
	}
	return out
}
func stringsOf(v any) []string {
	out := []string{}
	switch x := v.(type) {
	case []string:
		return x
	case []any:
		for _, s := range x {
			out = append(out, asString(s))
		}
	}
	return out
}

type recordField struct {
	name, typeName string
	optional       bool
}
type typeConverter struct {
	raw, self string
	resolve   func(string) string
	literal   func([]string) string
	record    func([]recordField) string
	callback  func(string) string
	union     func([]string) string
}

func (c typeConverter) walk(text string, depth int) string {
	text = strings.TrimSpace(text)
	if depth > 64 {
		unsupported(c.raw, "type nesting exceeds 64 levels")
	}
	if text == "" {
		unsupported(c.raw, "empty type expression")
	}
	if strings.HasPrefix(text, "fun(") || strings.HasPrefix(text, "fun ") {
		return c.function(text, depth+1)
	}
	members := splitTop(text, '|')
	for i, member := range members {
		if strings.HasPrefix(member, "fun(") || strings.HasPrefix(member, "fun ") {
			close := matchingClose(member, strings.IndexByte(member, '('))
			if strings.HasPrefix(strings.TrimSpace(member[close+1:]), ":") {
				members = append(members[:i], strings.Join(members[i:], " | "))
				break
			}
		}
	}
	if len(members) > 1 {
		literals, ordinary, functions := []string{}, []string{}, []string{}
		for _, member := range members {
			optional := strings.HasSuffix(member, "?") && (strings.HasPrefix(member, "\"") || strings.HasPrefix(member, "'"))
			literal := member
			if optional {
				literal = strings.TrimSuffix(literal, "?")
			}
			if value, ok := quotedLiteral(literal); ok {
				literals = append(literals, value)
				if optional {
					ordinary = append(ordinary, "nil")
				}
			} else if strings.HasPrefix(member, "fun(") || strings.HasPrefix(member, "fun ") {
				functions = append(functions, member)
			} else {
				ordinary = append(ordinary, member)
			}
		}
		if len(functions) > 1 {
			unsupported(c.raw, "multiple callback variants need an explicit overload declaration")
		}
		converted := []string{}
		if len(literals) > 0 && !containsString(ordinary, "string") {
			if c.literal == nil {
				unsupported(c.raw, "string literals require one caller-interned enum")
			}
			for _, member := range ordinary {
				converted = append(converted, c.walk(member, depth+1))
			}
			literals = uniqueStrings(literals)
			sort.Strings(literals)
			converted = append(converted, c.literal(literals))
		} else {
			for _, member := range ordinary {
				converted = append(converted, c.walk(member, depth+1))
			}
		}
		for _, member := range functions {
			converted = append(converted, c.walk(member, depth+1))
		}
		converted = uniqueStrings(converted)
		if c.union != nil {
			return c.union(converted)
		}
		tables := 0
		for _, member := range converted {
			if strings.HasPrefix(member, "{") {
				tables++
			}
		}
		if tables > 1 {
			unsupported(c.raw, "multiple anonymous table variants cannot be discriminated by Teal")
		}
		return strings.Join(converted, " | ")
	}
	if strings.HasSuffix(text, "?") {
		inner := c.walk(strings.TrimSuffix(text, "?"), depth+1)
		if strings.HasPrefix(inner, "function(") {
			return "nil | " + inner
		}
		return inner + " | nil"
	}
	if strings.HasSuffix(text, "[]") {
		return "{" + c.walk(strings.TrimSuffix(text, "[]"), depth+1) + "}"
	}
	if strings.HasPrefix(text, "(") {
		if matchingClose(text, 0) == len(text)-1 {
			return c.walk(text[1:len(text)-1], depth+1)
		}
		unsupported(c.raw, "unexpected suffix after grouped type")
	}
	if strings.HasPrefix(text, "{") {
		if matchingClose(text, 0) != len(text)-1 {
			unsupported(c.raw, "unexpected suffix after table type")
		}
		return c.table(text[1:len(text)-1], depth+1)
	}
	if value, ok := quotedLiteral(text); ok {
		if c.literal == nil {
			unsupported(c.raw, "string literal requires caller-interned enum")
		}
		return c.literal([]string{value})
	}
	if numericTypePattern.MatchString(text) {
		return "number"
	}
	switch text {
	case "true", "false":
		return "boolean"
	case "number", "integer", "string", "boolean", "nil", "any":
		return text
	case "table":
		return "{any: any}"
	case "function":
		s := "function(any...): any..."
		if c.callback != nil {
			return c.callback(s)
		}
		return s
	case "self":
		if c.self == "" {
			unsupported(c.raw, "self type requires an owning receiver")
		}
		return c.self
	}
	if strings.HasPrefix(text, "`") {
		unsupported(c.raw, "LuaLS string-dependent generics have no direct Teal equivalent")
	}
	if match := genericTypePattern.FindStringSubmatch(text); match != nil {
		args := splitTop(match[2], ',')
		if match[1] == "table" {
			if len(args) != 2 {
				unsupported(c.raw, "table generic requires explicit key and value types")
			}
			return "{" + c.walk(args[0], depth+1) + ": " + c.walk(args[1], depth+1) + "}"
		}
		converted := []string{}
		for _, arg := range args {
			converted = append(converted, c.walk(arg, depth+1))
		}
		return c.resolve(match[1]) + "<" + strings.Join(converted, ", ") + ">"
	}
	if !typeNamePattern.MatchString(text) {
		unsupported(c.raw, "unrecognized type syntax")
	}
	return c.resolve(text)
}
func (c typeConverter) function(text string, depth int) string {
	opening := strings.IndexByte(text, '(')
	if opening < 0 {
		unsupported(c.raw, "callback has no parameter list")
	}
	closing := matchingClose(text, opening)
	suffix := strings.TrimSpace(text[closing+1:])
	optional := suffix == "?"
	if suffix != "" && !optional && !strings.HasPrefix(suffix, ":") {
		unsupported(c.raw, "unrecognized callback suffix")
	}
	params := []string{}
	body := strings.TrimSpace(text[opening+1 : closing])
	if body != "" {
		for _, part := range splitTop(body, ',') {
			pair := splitTop(part, ':')
			if len(pair) == 1 {
				if part == "..." {
					params = append(params, "any...")
				} else if part == "self" && c.self != "" {
					params = append(params, c.self)
				} else {
					unsupported(c.raw, "untyped callback parameter: "+part)
				}
				continue
			}
			if len(pair) != 2 {
				unsupported(c.raw, "invalid callback parameter")
			}
			name := pair[0]
			value := c.walk(pair[1], depth+1)
			if strings.HasPrefix(name, "...") {
				value += "..."
			} else if strings.HasSuffix(name, "?") {
				value = "? " + value
			}
			params = append(params, value)
		}
	}
	signature := "function(" + strings.Join(params, ", ") + ")"
	if strings.HasPrefix(suffix, ":") {
		ret := strings.TrimSpace(suffix[1:])
		if ret == "..." {
			signature += ": any..."
		} else {
			values := []string{}
			for _, part := range splitTop(ret, ',') {
				values = append(values, c.walk(part, depth+1))
			}
			signature += ": " + strings.Join(values, ", ")
		}
	}
	if c.callback != nil {
		signature = c.callback(signature)
	}
	if optional {
		return "nil | " + signature
	}
	return signature
}
func (c typeConverter) table(body string, depth int) string {
	parts := splitTop(body, ',')
	if len(parts) == 1 && parts[0] == "" {
		if c.record != nil {
			return c.record([]recordField{})
		}
		unsupported(c.raw, "empty anonymous record needs a named declaration")
	}
	numeric := map[int]string{}
	fields := []recordField{}
	mapType := ""
	for _, part := range parts {
		pair := splitTop(part, ':')
		if len(pair) != 2 {
			unsupported(c.raw, "anonymous table member needs key and value")
		}
		key, value := pair[0], c.walk(pair[1], depth+1)
		if strings.HasPrefix(key, "[") && strings.HasSuffix(key, "]") {
			key = strings.TrimSpace(key[1 : len(key)-1])
			n, e := strconv.Atoi(key)
			if e == nil && n >= 0 && regexp.MustCompile(`^[0-9]+$`).MatchString(key) {
				if _, ok := numeric[n]; ok {
					unsupported(c.raw, "duplicate tuple index")
				}
				numeric[n] = value
			} else {
				if mapType != "" {
					unsupported(c.raw, "multiple map key types need an explicit declaration")
				}
				mapType = "{" + c.walk(key, depth+1) + ": " + value + "}"
			}
		} else if simpleFieldPattern.MatchString(key) {
			fields = append(fields, recordField{strings.TrimSuffix(key, "?"), value, strings.HasSuffix(key, "?")})
		} else {
			unsupported(c.raw, "unsupported anonymous record key: "+key)
		}
	}
	if len(numeric) > 0 && len(fields) == 0 && mapType == "" {
		values := []string{}
		for i := 1; i <= len(numeric); i++ {
			v, ok := numeric[i]
			if !ok {
				unsupported(c.raw, "sparse numeric tuple cannot be represented as a Teal tuple")
			}
			values = append(values, v)
		}
		return "{" + strings.Join(values, ", ") + "}"
	}
	if mapType != "" && len(fields) == 0 && len(numeric) == 0 {
		return mapType
	}
	if len(fields) > 0 && len(numeric) == 0 && mapType == "" && c.record != nil {
		return c.record(fields)
	}
	unsupported(c.raw, "anonymous record requires a caller-owned declaration")
	return ""
}
