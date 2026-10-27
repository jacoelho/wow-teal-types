package wowapi

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type factFailure struct{ message string }

func factFail(format string, args ...any) { panic(factFailure{fmt.Sprintf(format, args...)}) }

type factNode = []any

func fn(kind string, values ...any) factNode { return append([]any{kind}, values...) }
func kindOf(n factNode) string {
	if len(n) == 0 {
		return ""
	}
	return asString(n[0])
}
func childNode(v any) factNode { return asList(v) }

type sourceTypeReader struct {
	text       string
	pos, depth int
}

func (r *sourceTypeReader) space() {
	for r.pos < len(r.text) && unicode.IsSpace(rune(r.text[r.pos])) {
		r.pos++
	}
}
func (r *sourceTypeReader) token(s string) bool {
	saved := r.pos
	r.space()
	if strings.HasPrefix(r.text[r.pos:], s) {
		r.pos += len(s)
		return true
	}
	r.pos = saved
	return false
}
func (r *sourceTypeReader) require(s string) {
	if !r.token(s) {
		factFail("expected %s at %s", s, r.tail40())
	}
}
func (r *sourceTypeReader) tail40() string { return r.text[r.pos:min(len(r.text), r.pos+40)] }

var sourceIdent = regexp.MustCompile(`^[A-Za-z_]\w*(?:[.-]\w+)*`)
var sourceNumber = regexp.MustCompile(`^-?(?:0x[0-9a-fA-F]+|\d+(?:\.\d+)?)`)

func (r *sourceTypeReader) identifier() string {
	r.space()
	s := sourceIdent.FindString(r.text[r.pos:])
	if s == "" {
		factFail("expected type name at %s", r.tail40())
	}
	r.pos += len(s)
	return s
}
func (r *sourceTypeReader) expression() factNode {
	r.depth++
	if r.depth > 48 {
		factFail("type fact nesting exceeds 48 levels")
	}
	defer func() { r.depth-- }()
	a := []any{r.intersection()}
	for r.token("|") {
		a = append(a, r.intersection())
	}
	if len(a) == 1 {
		return childNode(a[0])
	}
	return fn("union", a...)
}
func (r *sourceTypeReader) intersection() factNode {
	unary := r.token("&")
	a := []any{r.atom()}
	for r.token("&") {
		a = append(a, r.atom())
	}
	if unary && (len(a) != 1 || kindOf(childNode(a[0])) != "variadic") {
		factFail("unary intersection requires a type pack")
	}
	if unary || len(a) > 1 {
		return fn("intersection", a...)
	}
	return childNode(a[0])
}
func (r *sourceTypeReader) atom() factNode {
	r.space()
	if r.token("...") {
		r.space()
		v := fn("name", "untyped-vararg")
		if r.pos < len(r.text) && !strings.ContainsRune(",)|&#", rune(r.text[r.pos])) {
			v = r.atom()
		}
		return fn("variadic", v)
	}
	var value factNode
	if r.token("(") {
		value = r.expression()
		r.require(")")
	} else if r.pos < len(r.text) && strings.ContainsRune("\"'`", rune(r.text[r.pos])) {
		start := r.pos
		q := r.text[r.pos]
		r.pos++
		closed := false
		for r.pos < len(r.text) {
			c := r.text[r.pos]
			r.pos++
			if c == '\\' {
				r.pos++
			} else if c == q {
				closed = true
				break
			}
		}
		if !closed {
			factFail("unterminated type literal")
		}
		raw := r.text[start:r.pos]
		if q == '`' {
			s := raw[1 : len(raw)-1]
			if !regexp.MustCompile(`^[A-Za-z_]\w*$`).MatchString(s) {
				factFail("backtick requires a generic class-name binder")
			}
			value = fn("class-name", s)
		} else {
			value = fn("literal", unquoteSource(raw))
		}
	} else {
		r.space()
		num := sourceNumber.FindString(r.text[r.pos:])
		if num != "" {
			r.pos += len(num)
			value = fn("name", "number")
		} else {
			name := r.identifier()
			if (name == "fun" || name == "function") && r.token("(") {
				value = r.callable()
			} else {
				if name == "true" || name == "false" {
					name = "boolean"
				}
				value = fn("name", name)
				if r.token("<") {
					args := []any{r.expression()}
					for r.token(",") {
						args = append(args, r.expression())
					}
					r.require(">")
					if name == "table" && len(args) != 2 {
						factFail("table generic requires key and value types")
					}
					value = fn("generic", append([]any{name}, args...)...)
				}
			}
		}
	}
	for {
		if r.token("[]") {
			value = fn("array", value)
		} else if r.token("?") {
			value = fn("union", value, fn("name", "nil"))
		} else if r.token("...") {
			value = fn("variadic", value)
		} else {
			break
		}
	}
	return value
}
func (r *sourceTypeReader) callable() factNode {
	params := []any{}
	if !r.token(")") {
		for {
			variadic := r.token("...")
			name := "..."
			if !variadic {
				name = r.identifier()
			}
			optional := r.token("?")
			var v factNode
			if r.token(":") {
				v = r.expression()
			} else if variadic {
				v = fn("name", "untyped-vararg")
			} else {
				factFail("untyped callback parameter: %s", name)
			}
			params = append(params, fn("parameter", optional, variadic, v))
			if r.token(")") {
				break
			}
			r.require(",")
			if variadic {
				factFail("callback vararg must be the final parameter")
			}
		}
	}
	returns := []any{}
	if r.token(":") {
		returns = append(returns, r.expression())
		for r.token(",") {
			returns = append(returns, r.expression())
		}
	}
	return fn("callable", params, returns)
}
func readSourceType(text string, prefix bool) (n factNode, end int, err error) {
	defer func() {
		if p := recover(); p != nil {
			if f, ok := p.(factFailure); ok {
				err = fmt.Errorf("%s", f.message)
			} else {
				panic(p)
			}
		}
	}()
	r := sourceTypeReader{text: text}
	r.space()
	if r.pos == len(text) || text[r.pos] == '#' {
		if prefix {
			return nil, r.pos, nil
		}
		factFail("empty type expression")
	}
	n = r.expression()
	end = r.pos
	r.space()
	if !prefix && r.pos != len(text) {
		factFail("unsupported type suffix: %s", text[r.pos:])
	}
	return n, end, nil
}
func splitSourceType(text string) (string, string, error) {
	_, end, e := readSourceType(text, true)
	if e != nil {
		return "", "", e
	}
	return strings.TrimSpace(text[:end]), strings.TrimSpace(text[end:]), nil
}
func unquoteSource(raw string) string {
	if len(raw) < 2 {
		factFail("unterminated type literal")
	}
	if raw[0] == '\'' {
		body := strings.ReplaceAll(raw[1:len(raw)-1], `\'`, `'`)
		body = strings.ReplaceAll(body, `"`, `\"`)
		raw = `"` + body + `"`
	}
	value, err := strconv.Unquote(raw)
	if err != nil {
		factFail("invalid type literal: %s", raw)
	}
	return value
}
