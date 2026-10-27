package literal

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

// ParseJSON keeps source number kinds and rejects duplicate keys before a typed
// evidence decoder interprets the resulting syntax. It never decodes into an
// erased Go payload.
func ParseJSON(data []byte) (Value, error) {
	if len(data) > 128<<20 {
		return Value{}, fmt.Errorf("JSON evidence exceeds 128 MiB")
	}
	if !utf8.Valid(data) {
		return Value{}, fmt.Errorf("invalid UTF-8 JSON evidence")
	}
	p := jsonParser{data: data, budget: 1 << 20}
	value, err := p.value(0)
	if err != nil {
		return Value{}, err
	}
	p.space()
	if p.at != len(data) {
		return Value{}, fmt.Errorf("trailing JSON data at byte %d", p.at)
	}
	return value, nil
}

type jsonParser struct {
	data       []byte
	at, budget int
}

func (p *jsonParser) space() {
	for p.at < len(p.data) && strings.ContainsRune(" \t\r\n", rune(p.data[p.at])) {
		p.at++
	}
}
func (p *jsonParser) take(c byte) bool {
	p.space()
	if p.at < len(p.data) && p.data[p.at] == c {
		p.at++
		return true
	}
	return false
}
func (p *jsonParser) text() (string, error) {
	p.space()
	start := p.at
	if p.at >= len(p.data) || p.data[p.at] != '"' {
		return "", fmt.Errorf("expected JSON string at byte %d", p.at)
	}
	p.at++
	for p.at < len(p.data) {
		c := p.data[p.at]
		p.at++
		if c == '\\' {
			if p.at < len(p.data) {
				p.at++
			}
			continue
		}
		if c == '"' {
			var value string
			if err := json.Unmarshal(p.data[start:p.at], &value); err != nil {
				return "", err
			}
			return value, nil
		}
	}
	return "", fmt.Errorf("unterminated JSON string")
}
func (p *jsonParser) value(depth int) (Value, error) {
	p.budget--
	if depth > 64 || p.budget < 0 {
		return Value{}, fmt.Errorf("JSON evidence exceeds nesting or value bound")
	}
	p.space()
	if p.at >= len(p.data) {
		return Value{}, fmt.Errorf("unexpected end of JSON evidence")
	}
	switch p.data[p.at] {
	case '"':
		text, err := p.text()
		return Text(text), err
	case '{':
		p.at++
		fields := Object{}
		if p.take('}') {
			return Record(fields), nil
		}
		for {
			key, err := p.text()
			if err != nil {
				return Value{}, err
			}
			if _, exists := fields[key]; exists {
				return Value{}, fmt.Errorf("duplicate JSON key %q", key)
			}
			if !p.take(':') {
				return Value{}, fmt.Errorf("expected JSON key separator at byte %d", p.at)
			}
			value, err := p.value(depth + 1)
			if err != nil {
				return Value{}, err
			}
			fields[key] = value
			if p.take('}') {
				return Record(fields), nil
			}
			if !p.take(',') {
				return Value{}, fmt.Errorf("expected JSON object separator at byte %d", p.at)
			}
		}
	case '[':
		p.at++
		values := []Value{}
		if p.take(']') {
			return List(values), nil
		}
		for {
			value, err := p.value(depth + 1)
			if err != nil {
				return Value{}, err
			}
			values = append(values, value)
			if p.take(']') {
				return List(values), nil
			}
			if !p.take(',') {
				return Value{}, fmt.Errorf("expected JSON array separator at byte %d", p.at)
			}
		}
	case 't':
		if p.word("true") {
			return Bool(true), nil
		}
	case 'f':
		if p.word("false") {
			return Bool(false), nil
		}
	case 'n':
		if p.word("null") {
			return Value{}, nil
		}
	default:
		start := p.at
		for p.at < len(p.data) && strings.ContainsRune("0123456789+-.eE", rune(p.data[p.at])) {
			p.at++
		}
		if start == p.at {
			break
		}
		token := p.data[start:p.at]
		if !json.Valid(token) {
			return Value{}, fmt.Errorf("invalid JSON number at byte %d", start)
		}
		if !strings.ContainsAny(string(token), ".eE") {
			return BigInteger(string(token))
		}
		number, err := strconv.ParseFloat(string(token), 64)
		if err != nil {
			return Value{}, err
		}
		return Number(number)
	}
	return Value{}, fmt.Errorf("invalid JSON value at byte %d", p.at)
}
func (p *jsonParser) word(word string) bool {
	// Converting the remaining input for every keyword makes parsing quadratic.
	if len(p.data)-p.at >= len(word) && string(p.data[p.at:p.at+len(word)]) == word {
		p.at += len(word)
		return true
	}
	return false
}
