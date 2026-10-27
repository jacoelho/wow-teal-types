package wowapi

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"wow-teal-types/internal/wowapi/literal"
)

type factFailure struct{ message string }

func factFail(message string) { panic(factFailure{message}) }

// factNode is the interpreted type-expression tree, separate from source literals.
type factNode = *typeFactNode
type typeFactNode struct {
	kind       string
	text       string
	index      int
	children   []factNode
	parameters []factParameter
	returns    []factNode
}
type factParameter struct {
	optional, variadic bool
	value              factNode
}

func nameFact(name string) factNode               { return &typeFactNode{kind: "name", text: name} }
func textFact(kind, text string) factNode         { return &typeFactNode{kind: kind, text: text} }
func indexedFact(kind string, index int) factNode { return &typeFactNode{kind: kind, index: index} }
func compoundFact(kind string, children ...factNode) factNode {
	return &typeFactNode{kind: kind, children: children}
}
func genericFact(name string, children ...factNode) factNode {
	return &typeFactNode{kind: "generic", text: name, children: children}
}
func callableFact(parameters []factParameter, returns []factNode) factNode {
	return &typeFactNode{kind: "callable", parameters: parameters, returns: returns}
}
func kindOf(node factNode) string {
	if node == nil {
		return ""
	}
	return node.kind
}
func (node *typeFactNode) literal() literal.Value {
	if node == nil {
		return literal.Value{}
	}
	values := []literal.Value{literal.Text(node.kind)}
	switch node.kind {
	case "name", "literal":
		values = append(values, literal.Text(node.text))
	case "class-name":
		if node.text != "" {
			values = append(values, literal.Text(node.text))
		} else {
			values = append(values, literal.Int(int64(node.index)))
		}
	case "variable", "pack":
		values = append(values, literal.Int(int64(node.index)))
	case "generic":
		values = append(values, literal.Text(node.text))
		for _, child := range node.children {
			values = append(values, child.literal())
		}
	case "callable":
		parameters := make([]literal.Value, 0, len(node.parameters))
		for _, p := range node.parameters {
			parameters = append(parameters, literal.List([]literal.Value{literal.Text("parameter"), literal.Bool(p.optional), literal.Bool(p.variadic), p.value.literal()}))
		}
		returns := make([]literal.Value, 0, len(node.returns))
		for _, r := range node.returns {
			returns = append(returns, r.literal())
		}
		values = append(values, literal.List(parameters), literal.List(returns))
	default:
		for _, child := range node.children {
			values = append(values, child.literal())
		}
	}
	return literal.List(values)
}
func (node *typeFactNode) MarshalJSON() ([]byte, error) { return json.Marshal(node.literal()) }

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
		factFail(fmt.Sprintf("expected %s at %s", s, r.tail40()))
	}
}
func (r *sourceTypeReader) tail40() string { return r.text[r.pos:min(len(r.text), r.pos+40)] }

var sourceIdent = regexp.MustCompile(`^[A-Za-z_]\w*(?:[.-]\w+)*`)
var sourceNumber = regexp.MustCompile(`^-?(?:0x[0-9a-fA-F]+|\d+(?:\.\d+)?)`)

func (r *sourceTypeReader) identifier() string {
	r.space()
	s := sourceIdent.FindString(r.text[r.pos:])
	if s == "" {
		factFail(fmt.Sprintf("expected type name at %s", r.tail40()))
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
	a := []factNode{r.intersection()}
	for r.token("|") {
		a = append(a, r.intersection())
	}
	if len(a) == 1 {
		return a[0]
	}
	return compoundFact("union", a...)
}
func (r *sourceTypeReader) intersection() factNode {
	unary := r.token("&")
	a := []factNode{r.atom()}
	for r.token("&") {
		a = append(a, r.atom())
	}
	if unary && (len(a) != 1 || kindOf(a[0]) != "variadic") {
		factFail("unary intersection requires a type pack")
	}
	if unary || len(a) > 1 {
		return compoundFact("intersection", a...)
	}
	return a[0]
}
func (r *sourceTypeReader) atom() factNode {
	r.space()
	if r.token("...") {
		r.space()
		v := nameFact("untyped-vararg")
		if r.pos < len(r.text) && !strings.ContainsRune(",)|&#", rune(r.text[r.pos])) {
			v = r.atom()
		}
		return compoundFact("variadic", v)
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
			value = textFact("class-name", s)
		} else {
			value = textFact("literal", unquoteSource(raw))
		}
	} else {
		r.space()
		num := sourceNumber.FindString(r.text[r.pos:])
		if num != "" {
			r.pos += len(num)
			value = nameFact("number")
		} else {
			name := r.identifier()
			if (name == "fun" || name == "function") && r.token("(") {
				value = r.callable()
			} else {
				if name == "true" || name == "false" {
					name = "boolean"
				}
				value = nameFact(name)
				if r.token("<") {
					args := []factNode{r.expression()}
					for r.token(",") {
						args = append(args, r.expression())
					}
					r.require(">")
					if name == "table" && len(args) != 2 {
						factFail("table generic requires key and value types")
					}
					value = genericFact(name, args...)
				}
			}
		}
	}
	for {
		if r.token("[]") {
			value = compoundFact("array", value)
		} else if r.token("?") {
			value = compoundFact("union", value, nameFact("nil"))
		} else if r.token("...") {
			value = compoundFact("variadic", value)
		} else {
			break
		}
	}
	return value
}
func (r *sourceTypeReader) callable() factNode {
	params := []factParameter{}
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
				v = nameFact("untyped-vararg")
			} else {
				factFail(fmt.Sprintf("untyped callback parameter: %s", name))
			}
			params = append(params, factParameter{optional, variadic, v})
			if r.token(")") {
				break
			}
			r.require(",")
			if variadic {
				factFail("callback vararg must be the final parameter")
			}
		}
	}
	returns := []factNode{}
	if r.token(":") {
		returns = append(returns, r.expression())
		for r.token(",") {
			returns = append(returns, r.expression())
		}
	}
	return callableFact(params, returns)
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
		factFail(fmt.Sprintf("unsupported type suffix: %s", text[r.pos:]))
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
		factFail(fmt.Sprintf("invalid type literal: %s", raw))
	}
	return value
}
