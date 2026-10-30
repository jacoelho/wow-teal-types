package wowapi

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

var typeNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z_0-9.\-]*$`)
var numericTypePattern = regexp.MustCompile(`^-?(?:0x[0-9A-Fa-f]+|\d+(?:\.\d+)?)$`)
var genericTypePattern = regexp.MustCompile(`^([A-Za-z_][A-Za-z_0-9.\-]*)\s*<(.*)>$`)
var decimalDigitsPattern = regexp.MustCompile(`^[0-9]+$`)
var simpleFieldPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z_0-9]*\??$`)

func unsupported(raw, reason string) { reject(fmt.Sprintf("%q: %s", raw, reason)) }
func typeCloser(c byte) byte {
	switch c {
	case '(':
		return ')'
	case '[':
		return ']'
	case '{':
		return '}'
	case '<':
		return '>'
	}
	return 0
}

// scanTypeDelimiters walks text from start, skipping quoted literals and
// tracking bracket nesting. visit receives each unquoted byte with the
// nesting depth after that byte and returns false to stop. The result
// reports whether the scan stopped early or ended with every quote and
// bracket closed; otherwise the caller rejects the unterminated text.
func scanTypeDelimiters(text string, start int, visit func(i, depth int) bool) bool {
	stack := []byte{}
	quote := byte(0)
	escaped := false
	for i := start; i < len(text); i++ {
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
			continue
		}
		if closer := typeCloser(c); closer != 0 {
			stack = append(stack, closer)
			if len(stack) > 64 {
				unsupported(text, "type nesting exceeds 64 levels")
			}
		} else if strings.IndexByte(")]}>", c) >= 0 {
			if len(stack) == 0 || stack[len(stack)-1] != c {
				unsupported(text, "unbalanced type delimiters")
			}
			stack = stack[:len(stack)-1]
		}
		if !visit(i, len(stack)) {
			return true
		}
	}
	return quote == 0 && len(stack) == 0
}

// nilMember matches a nil union member with its adjacent separator. nil is a
// reserved word, so the token cannot be part of a type name, and a separator
// next to it always belongs to the same union.
var nilMember = regexp.MustCompile(`\s*\|\s*nil\b|\bnil\s*\|\s*`)

// omitNil removes nil union members from a generated type expression. Teal
// admits nil for every type but cannot narrow `T | nil` through truthiness, so
// the explicit member blocks idiomatic absence checks without adding safety.
// A lone nil is kept; optional parameters keep their `? T` marker.
func omitNil(expression string) string {
	if strings.ContainsAny(expression, "\"'`") {
		unsupported(expression, "quoted literal in generated type expression")
	}
	return nilMember.ReplaceAllString(expression, "")
}

// splitTop splits text at delimiter bytes outside quotes and brackets.
func splitTop(text string, delimiter byte) []string {
	start := 0
	parts := []string{}
	closed := scanTypeDelimiters(text, 0, func(i, depth int) bool {
		if depth == 0 && text[i] == delimiter {
			parts = append(parts, strings.TrimSpace(text[start:i]))
			start = i + 1
		}
		return true
	})
	if !closed {
		unsupported(text, "unterminated type literal or delimiter")
	}
	return append(parts, strings.TrimSpace(text[start:]))
}

// matchingClose returns the index of the bracket closing text[opening].
func matchingClose(text string, opening int) int {
	if opening < 0 || opening >= len(text) || typeCloser(text[opening]) == 0 {
		unsupported(text, "missing type delimiter")
	}
	closing := -1
	scanTypeDelimiters(text, opening, func(i, depth int) bool {
		if depth == 0 {
			closing = i
			return false
		}
		return true
	})
	if closing < 0 {
		unsupported(text, "unterminated type delimiter")
	}
	return closing
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

// typePosition is the direction a value crosses the API boundary. It decides
// how a type without source information is declared.
type typePosition uint8

const (
	// inputPosition values are supplied by the addon: parameters, and the
	// results of callbacks the game invokes.
	inputPosition typePosition = iota + 1
	// outputPosition values are produced by the game: results, fields, and
	// the arguments the game passes to callbacks.
	outputPosition
)

// nativeValue is a native_values.d.tl type that wow re-exports, so generated
// environment files and addons name it through the W alias of the wow module.
type nativeValue struct{ name, parameters string }

// nativeValues are every re-exported native value type, in declaration order.
var nativeValues = []nativeValue{{"Data", ""}, {"Atom", ""}, {"SecretData", ""}, {"SecretTable", ""}, {"MaybeSecret", "<T>"}, {"Secret", "<T>"}}

func isNativeValue(name string) bool {
	return slices.ContainsFunc(nativeValues, func(v nativeValue) bool { return v.name == name })
}

// Generated spellings of the native value types.
const (
	dataType          = "W.Data"
	secretDataType    = "W.SecretData"
	secretTableType   = "W.Secret<W.SecretTable>"
	secretPrefix      = "W.Secret<"
	maybeSecretPrefix = "W.MaybeSecret<"
)

func secretOf(t string) string      { return secretPrefix + t + ">" }
func maybeSecretOf(t string) string { return maybeSecretPrefix + t + ">" }

func (p typePosition) flip() typePosition {
	if p == inputPosition {
		return outputPosition
	}
	return inputPosition
}

// untyped is the declaration of a value with no type information. Inputs stay
// any so records, functions and table literals remain acceptable; outputs are
// public plain values of undocumented shape.
func (p typePosition) untyped() string {
	if p == inputPosition {
		return "any"
	}
	return dataType
}

type recordField struct {
	name, typeName string
	optional       bool
}

// typeConverter translates one LuaLS type expression at one position. Every
// callback is required: the catalogue owns interning and name resolution.
// Function parameters take the flipped position; everything else nested in
// the expression, including anonymous record fields, keeps it.
type typeConverter struct {
	raw, self string
	position  typePosition
	resolve   func(string, typePosition) string
	literal   func([]string) string
	record    func([]recordField) string
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
		if len(literals) > 0 && !slices.Contains(ordinary, "string") {
			for _, member := range ordinary {
				converted = append(converted, c.walk(member, depth+1))
			}
			literals = uniqueStrings(literals)
			slices.Sort(literals)
			converted = append(converted, c.literal(literals))
		} else {
			for _, member := range ordinary {
				converted = append(converted, c.walk(member, depth+1))
			}
		}
		for _, member := range functions {
			converted = append(converted, c.walk(member, depth+1))
		}
		return c.union(uniqueStrings(converted))
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
		return c.literal([]string{value})
	}
	if numericTypePattern.MatchString(text) {
		return "number"
	}
	switch text {
	case "true", "false":
		return "boolean"
	case "any", "unknown", "table", "function":
		// Bare table and function identity establish no shape or signature.
		return c.position.untyped()
	case "number", "integer", "string", "boolean", "nil":
		return text
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
		return c.resolve(match[1], c.position) + "<" + strings.Join(converted, ", ") + ">"
	}
	if !typeNamePattern.MatchString(text) {
		unsupported(c.raw, "unrecognized type syntax")
	}
	return c.resolve(text, c.position)
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
	parameter := c
	parameter.position = c.position.flip()
	body := strings.TrimSpace(text[opening+1 : closing])
	if body != "" {
		for _, part := range splitTop(body, ',') {
			pair := splitTop(part, ':')
			if len(pair) == 1 {
				if part == "..." {
					params = append(params, parameter.position.untyped()+"...")
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
			value := parameter.walk(pair[1], depth+1)
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
		values := []string{}
		parts := splitTop(ret, ',')
		for index, part := range parts {
			pair := splitTop(part, ':')
			variadic := pair[0] == "..."
			if variadic && index != len(parts)-1 {
				unsupported(c.raw, "variadic callback return must be last")
			}
			switch {
			case variadic && len(pair) == 1:
				values = append(values, c.position.untyped()+"...")
			case variadic && len(pair) == 2:
				values = append(values, c.walk(pair[1], depth+1)+"...")
			case len(pair) == 1:
				values = append(values, c.walk(part, depth+1))
			default:
				unsupported(c.raw, "invalid callback return")
			}
		}
		// Teal result lists are greedy: unparenthesized, a following parameter,
		// result or union member would become another result. Callers compose
		// this value into those positions, so the group always closes it.
		signature += ": (" + strings.Join(values, ", ") + ")"
	}
	if optional {
		return "nil | " + signature
	}
	return signature
}

// wholeSignature spells a converted function type used as a complete
// declaration: its result group is the end of the declaration, so it is
// written like every other generated signature and compares equal to them.
func wholeSignature(converted string) string {
	if !strings.HasPrefix(converted, "function(") {
		return converted
	}
	closing := matchingClose(converted, len("function"))
	results := converted[closing+1:]
	if !strings.HasPrefix(results, ": (") || matchingClose(converted, closing+3) != len(converted)-1 {
		return converted
	}
	return converted[:closing+1] + ": " + converted[closing+4:len(converted)-1]
}

// groupedSignature is the inverse of wholeSignature: a function type inside
// another type needs its results grouped, or they would read as part of the
// enclosing type.
func groupedSignature(signature string) string {
	if match := functionTypeKeyword.FindStringIndex(signature); match == nil || match[0] != 0 {
		return signature
	}
	parameters := len("function")
	if signature[parameters] == '<' {
		parameters = matchingClose(signature, parameters) + 1
	}
	closing := matchingClose(signature, parameters)
	if !strings.HasPrefix(signature[closing+1:], ": ") {
		return signature
	}
	return signature[:closing+1] + ": (" + signature[closing+3:] + ")"
}

// functionTypeKeyword starts a converted function type; functionResultGroup
// opens its result group, however the colon is spaced.
var (
	functionTypeKeyword = regexp.MustCompile(`\bfunction[(<]`)
	functionResultGroup = regexp.MustCompile(`^\s*:\s*\(`)
)

// withoutFunctionTypes removes every function type from expression, which is
// converted Teal text. A converted function type ends at its parameter list or
// result group, so the span is exact at any depth.
func withoutFunctionTypes(expression string) string {
	var out strings.Builder
	for {
		match := functionTypeKeyword.FindStringIndex(expression)
		if match == nil {
			out.WriteString(expression)
			return out.String()
		}
		out.WriteString(expression[:match[0]])
		end := match[1] - 1
		if expression[end] == '<' {
			end = matchingClose(expression, end) + 1
		}
		end = matchingClose(expression, end) + 1
		if group := functionResultGroup.FindStringIndex(expression[end:]); group != nil {
			end = matchingClose(expression, end+group[1]-1) + 1
		}
		expression = expression[end:]
	}
}
func (c typeConverter) table(body string, depth int) string {
	parts := splitTop(body, ',')
	if len(parts) == 1 && parts[0] == "" {
		return c.record([]recordField{})
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
			if e == nil && n >= 0 && decimalDigitsPattern.MatchString(key) {
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
	if len(fields) > 0 && len(numeric) == 0 && mapType == "" {
		return c.record(fields)
	}
	unsupported(c.raw, "anonymous record requires a caller-owned declaration")
	return ""
}
