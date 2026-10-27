package wowapi

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type conversionError struct{ message string }

func (e conversionError) Error() string { return e.message }
func reject(format string, args ...any) { panic(conversionError{fmt.Sprintf(format, args...)}) }
func captureError(target *error) {
	if r := recover(); r != nil {
		if e, ok := r.(conversionError); ok {
			*target = e
		} else {
			panic(r)
		}
	}
}
func captureString(fn func() string) (value string, err error) {
	defer captureError(&err)
	value = fn()
	return
}

type luaToken struct {
	text string
	line int
}
type luaParser struct {
	tokens []luaToken
	pos    int
	source string
}

var luaKeywords = wordSet("and break do else elseif end false for function if in local nil not or repeat return then true until while")

func wordSet(s string) map[string]bool {
	m := map[string]bool{}
	for word := range strings.FieldsSeq(s) {
		m[word] = true
	}
	return m
}
func identifier(s string) bool {
	if len(s) == 0 || luaKeywords[s] || !(s[0] == '_' || s[0] >= 'A' && s[0] <= 'Z' || s[0] >= 'a' && s[0] <= 'z') {
		return false
	}
	for _, c := range s[1:] {
		if !(c == '_' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9') {
			return false
		}
	}
	return true
}
func longBracket(text string, start int) (int, int, bool) {
	if start >= len(text) || text[start] != '[' {
		return 0, 0, false
	}
	i := start + 1
	for i < len(text) && text[i] == '=' {
		i++
	}
	if i >= len(text) || text[i] != '[' {
		return 0, 0, false
	}
	closing := "]" + text[start+1:i] + "]"
	end := strings.Index(text[i+1:], closing)
	if end < 0 {
		return i + 1, 0, true
	}
	return i + 1, i + 1 + end + len(closing), true
}
func luaTokens(text, source string) []luaToken {
	tokens := []luaToken{}
	line := 1
	for i := 0; i < len(text); {
		start := i
		c := text[i]
		if unicode.IsSpace(rune(c)) {
			if c == '\n' {
				line++
			}
			i++
			continue
		}
		if strings.HasPrefix(text[i:], "--") {
			if _, end, long := longBracket(text, i+2); long {
				if end == 0 {
					reject("%s: unterminated long comment", source)
				}
				line += strings.Count(text[i:end], "\n")
				i = end
			} else {
				end := strings.IndexByte(text[i:], '\n')
				if end < 0 {
					break
				}
				i += end
			}
			continue
		}
		tokenLine := line
		if _, end, long := longBracket(text, i); long {
			if end == 0 {
				reject("%s: unterminated long string", source)
			}
			i = end
			line += strings.Count(text[start:i], "\n")
		} else if c == '\'' || c == '"' {
			i++
			closed := false
			for i < len(text) {
				if text[i] == '\\' {
					i += 2
					continue
				}
				if text[i] == c {
					i++
					closed = true
					break
				}
				i++
			}
			if !closed || i > len(text) {
				reject("%s: unterminated string", source)
			}
			line += strings.Count(text[start:i], "\n")
		} else if c >= '0' && c <= '9' {
			i++
			if c == '0' && i < len(text) && (text[i] == 'x' || text[i] == 'X') {
				i++
				for i < len(text) && strings.ContainsRune("0123456789abcdefABCDEF", rune(text[i])) {
					i++
				}
			} else {
				for i < len(text) && text[i] >= '0' && text[i] <= '9' {
					i++
				}
				if i < len(text) && text[i] == '.' {
					i++
					for i < len(text) && text[i] >= '0' && text[i] <= '9' {
						i++
					}
				}
				if i < len(text) && (text[i] == 'e' || text[i] == 'E') {
					i++
					if i < len(text) && (text[i] == '+' || text[i] == '-') {
						i++
					}
					begin := i
					for i < len(text) && text[i] >= '0' && text[i] <= '9' {
						i++
					}
					if i == begin {
						reject("%s: malformed exponent", source)
					}
				}
			}
		} else if c == '_' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' {
			i++
			for i < len(text) && (text[i] == '_' || text[i] >= 'A' && text[i] <= 'Z' || text[i] >= 'a' && text[i] <= 'z' || text[i] >= '0' && text[i] <= '9') {
				i++
			}
		} else if strings.ContainsRune("{}[](),;=.:+-", rune(c)) {
			i++
		} else {
			reject("%s: unsupported Lua near %q", source, text[i:min(i+60, len(text))])
		}
		tokens = append(tokens, luaToken{text[start:i], tokenLine})
	}
	return tokens
}
func decodeLuaString(raw, source string) string {
	if raw[0] == '[' {
		start, _, _ := longBracket(raw, 0)
		end := len(raw) - start
		return strings.TrimPrefix(raw[start:end], "\n")
	}
	body := raw[1 : len(raw)-1]
	var out strings.Builder
	escapes := map[byte]byte{'a': '\a', 'b': '\b', 'f': '\f', 'n': '\n', 'r': '\r', 't': '\t', 'v': '\v', '\\': '\\', '"': '"', '\'': '\'', '\n': '\n'}
	for i := 0; i < len(body); {
		c := body[i]
		i++
		if c != '\\' {
			if c == '\r' || c == '\n' {
				reject("%s: unescaped newline in quoted string", source)
			}
			out.WriteByte(c)
			continue
		}
		if i == len(body) {
			reject("%s: incomplete string escape", source)
		}
		c = body[i]
		i++
		if c >= '0' && c <= '9' {
			start := i - 1
			for i < len(body) && i-start < 3 && body[i] >= '0' && body[i] <= '9' {
				i++
			}
			value, _ := strconv.Atoi(body[start:i])
			if value > 127 {
				reject("%s: non-ASCII byte escape is unsupported in UTF-8 metadata", source)
			}
			out.WriteByte(byte(value))
		} else if value, ok := escapes[c]; ok {
			out.WriteByte(value)
		} else {
			reject("%s: unsupported string escape %q", source, c)
		}
	}
	return out.String()
}
func (p *luaParser) peek() string {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos].text
	}
	return ""
}
func (p *luaParser) take(expected ...string) string {
	if p.pos >= len(p.tokens) {
		reject("%s: unexpected end of documentation", p.source)
	}
	t := p.tokens[p.pos]
	p.pos++
	if len(expected) > 0 && t.text != expected[0] {
		reject("%s:%d: expected %q, got %q", p.source, t.line, expected[0], t.text)
	}
	return t.text
}
func luaNumber(v any) (float64, bool, bool) {
	switch x := v.(type) {
	case int:
		return float64(x), true, true
	case int64:
		return float64(x), true, true
	case float64:
		return x, true, false
	case json.Number:
		n, e := x.Float64()
		return n, e == nil, !strings.ContainsAny(string(x), ".eE")
	}
	return 0, false, false
}
func arithmeticReference(v any) bool {
	if _, ok, _ := luaNumber(v); ok {
		return true
	}
	m, ok := v.(map[string]any)
	return ok && len(m) == 1 && (m["$ref"] != nil || m["$expression"] != nil)
}
func (p *luaParser) value(depth int) any {
	value := p.atom(depth)
	for p.peek() == "+" || p.peek() == "-" {
		op := p.take()
		right := p.atom(depth)
		aInt, ai := luaInteger(value)
		bInt, bi := luaInteger(right)
		a, an, _ := luaNumber(value)
		b, bn, _ := luaNumber(right)
		if ai && bi {
			if op == "+" {
				aInt.Add(aInt, bInt)
			} else {
				aInt.Sub(aInt, bInt)
			}
			value = luaIntegerValue(aInt)
		} else if an && bn {
			if op == "+" {
				a += b
			} else {
				a -= b
			}
			if math.IsInf(a, 0) || math.IsNaN(a) {
				reject("%s: nonfinite documentation number", p.source)
			}
			value = a
		} else {
			if !arithmeticReference(value) || !arithmeticReference(right) {
				reject("%s: arithmetic requires numbers or named references", p.source)
			}
			value = M{"$expression": []any{value, op, right}}
		}
	}
	return value
}
func (p *luaParser) atom(depth int) any {
	if depth > 64 {
		reject("%s: table nesting exceeds 64", p.source)
	}
	token := p.take()
	if token == "{" {
		entries := []any{}
		fields := M{}
		for p.peek() != "}" {
			if p.peek() == "[" {
				p.take("[")
				key, ok := p.value(depth + 1).(string)
				p.take("]")
				p.take("=")
				if !ok {
					reject("%s: only string documentation table keys are supported", p.source)
				}
				if strings.HasPrefix(key, "$") {
					reject("%s: reserved metadata representation key %q", p.source, key)
				}
				if _, ok := fields[key]; ok {
					reject("%s: duplicate table key %s", p.source, key)
				}
				fields[key] = p.value(depth + 1)
			} else if p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].text == "=" {
				key := p.take()
				if !identifier(key) {
					reject("%s: invalid documentation key %q", p.source, key)
				}
				p.take("=")
				if _, ok := fields[key]; ok {
					reject("%s: duplicate table key %s", p.source, key)
				}
				fields[key] = p.value(depth + 1)
			} else {
				entries = append(entries, p.value(depth+1))
			}
			if p.peek() == "," || p.peek() == ";" {
				p.take()
			} else if p.peek() != "}" {
				reject("%s: expected table separator, got %s", p.source, p.peek())
			}
		}
		p.take("}")
		if len(fields) > 0 && len(entries) > 0 {
			reject("%s: mixed table keys unsupported", p.source)
		}
		if len(fields) > 0 {
			return fields
		}
		return entries
	}
	if token[0] == '\'' || token[0] == '"' || token[0] == '[' && len(token) > 1 {
		return decodeLuaString(token, p.source)
	}
	switch token {
	case "true":
		return true
	case "false":
		return false
	case "nil":
		return nil
	case "-", "+":
		v := p.atom(depth + 1)
		if integer, ok := luaInteger(v); ok {
			if token == "-" {
				integer.Neg(integer)
			}
			return luaIntegerValue(integer)
		}
		n, ok, _ := luaNumber(v)
		if !ok {
			reject("%s: unary operator on non-number", p.source)
		}
		if token == "-" {
			n = -n
		}
		return n
	}
	if token[0] >= '0' && token[0] <= '9' {
		if strings.HasPrefix(strings.ToLower(token), "0x") {
			n, ok := new(big.Int).SetString(token[2:], 16)
			if !ok {
				reject("%s: unsupported number", p.source)
			}
			return luaIntegerValue(n)
		}
		if !strings.ContainsAny(token, ".eE") {
			if len(token) > 4300 {
				reject("%s: documentation integer exceeds4300digits", p.source)
			}
			n, ok := new(big.Int).SetString(token, 10)
			if !ok {
				reject("%s: unsupported number", p.source)
			}
			return luaIntegerValue(n)
		}
		n, e := strconv.ParseFloat(token, 64)
		if e != nil || math.IsInf(n, 0) || math.IsNaN(n) {
			reject("%s: nonfinite documentation number", p.source)
		}
		return n
	}
	if !identifier(token) {
		reject("%s: expected a literal or named reference, got %q", p.source, token)
	}
	parts := []string{token}
	for p.peek() == "." {
		p.take(".")
		part := p.take()
		if !identifier(part) {
			reject("%s: invalid named reference component %q", p.source, part)
		}
		parts = append(parts, part)
	}
	return M{"$ref": strings.Join(parts, ".")}
}

// ParseLua reads a closed documentation-table grammar without executing Lua.
func ParseLua(data []byte, source string) (result M, err error) {
	defer captureError(&err)
	if !utf8.Valid(data) {
		reject("%s: invalid UTF-8 metadata", source)
	}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	p := luaParser{tokens: luaTokens(text, source), source: source}
	p.take("local")
	name := p.take()
	if !identifier(name) {
		reject("%s: invalid documentation table name %q", source, name)
	}
	p.take("=")
	value := p.value(0)
	result, ok := value.(map[string]any)
	if !ok {
		reject("%s: documentation must be a keyed table", source)
	}
	if p.peek() == ";" {
		p.take()
	}
	if p.peek() == "APIDocumentation" {
		for _, token := range []string{"APIDocumentation", ":", "AddDocumentationTable", "(", name, ")"} {
			p.take(token)
		}
		if p.peek() == ";" {
			p.take()
		}
	}
	if p.peek() != "" {
		reject("%s: unsupported statement after documentation table: %q", source, p.peek())
	}
	return
}
func parseLuaFile(path string) (M, error) {
	data, e := os.ReadFile(path)
	if e != nil {
		return nil, e
	}
	return ParseLua(data, path)
}

func luaInteger(value any) (*big.Int, bool) {
	switch v := value.(type) {
	case int:
		return big.NewInt(int64(v)), true
	case int64:
		return big.NewInt(v), true
	case json.Number:
		if !strings.ContainsAny(string(v), ".eE") {
			n, ok := new(big.Int).SetString(string(v), 10)
			return n, ok
		}
	}
	return nil, false
}
func luaIntegerValue(value *big.Int) any {
	if value.IsInt64() {
		return value.Int64()
	}
	return json.Number(value.String())
}
