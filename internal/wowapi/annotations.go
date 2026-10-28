package wowapi

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

const annotationName = `[A-Za-z_][A-Za-z_0-9.\-]*`

var annotationFunctionPattern = regexp.MustCompile(`^(?:local\s+)?function\s+(` + annotationName + `(?::` + annotationName + `)?)\s*\(([^)]*)\)`)
var annotationBindingPattern = regexp.MustCompile(`^(local\s+)?(` + annotationName + `)\s*=\s*(.*)$`)
var annotationTagPattern = regexp.MustCompile(`^---@([A-Za-z]+)\s*(.*)$`)
var annotationClassPattern = regexp.MustCompile(`^(?:\(([^)]*)\)\s*)?(` + annotationName + `)(?:\s*:\s*(.*))?$`)
var annotationAliasPattern = regexp.MustCompile(`^(` + annotationName + `)(?:\s+(.*))?$`)
var annotationTypePrefixPattern = regexp.MustCompile(`^(?:\.\.\.|-?(?:0x[0-9A-Fa-f]+|\d+(?:\.\d+)?)|` + annotationName + `)`)
var annotationParameterNamePattern = regexp.MustCompile(`^(?:\.\.\.|` + annotationName + `)\??`)
var annotationVisibilityPattern = regexp.MustCompile(`^(private|protected|public|package)\s+`)
var annotationReturnLabelPattern = regexp.MustCompile(`^(\.\.\.|[A-Za-z_][A-Za-z_0-9]*)(?:\s+(.*))?$`)
var annotationEnumValuePattern = regexp.MustCompile(`^(` + annotationName + `|\[.*?\])\s*=\s*(.*?)(?:,)?\s*(?:--.*)?$`)

func annotationQuotedEnd(text string, start int) int {
	quote := text[start]
	for i := start + 1; i < len(text); i++ {
		if text[i] == '\\' {
			i++
		} else if text[i] == quote {
			return i + 1
		}
	}
	reject("unterminated quoted literal")
	return 0
}
func annotationBalancedEnd(text string, start int) int { return matchingClose(text, start) + 1 }
func annotationTypeEnd(text string, start int) int {
	i := start
	for i < len(text) && strings.ContainsRune(" \t\r\n", rune(text[i])) {
		i++
	}
	if i >= len(text) || text[i] == '#' {
		return i
	}
	if strings.HasPrefix(text[i:], "fun(") || strings.HasPrefix(text[i:], "fun ") {
		opening := strings.IndexByte(text[i+3:], '(')
		if opening < 0 {
			reject("function type lacks parameter list")
		}
		opening += i + 3
		i = annotationBalancedEnd(text, opening)
		after := i
		for after < len(text) && strings.ContainsRune(" \t\r\n", rune(text[after])) {
			after++
		}
		if after < len(text) && text[after] == ':' {
			i = annotationTypeEnd(text, after+1)
			for {
				after = i
				for after < len(text) && strings.ContainsRune(" \t\r\n", rune(text[after])) {
					after++
				}
				if after >= len(text) || text[after] != ',' {
					break
				}
				i = annotationTypeEnd(text, after+1)
			}
		}
	} else if text[i] == '"' || text[i] == '\'' || text[i] == '`' {
		i = annotationQuotedEnd(text, i)
	} else if text[i] == '(' || text[i] == '{' {
		i = annotationBalancedEnd(text, i)
	} else {
		match := annotationTypePrefixPattern.FindString(text[i:])
		if match == "" {
			reject(fmt.Sprintf("unrecognized type prefix: %s", text[i:min(i+40, len(text))]))
		}
		i += len(match)
		after := i
		for after < len(text) && strings.ContainsRune(" \t\r\n", rune(text[after])) {
			after++
		}
		if after < len(text) && text[after] == '<' {
			i = annotationBalancedEnd(text, after)
		} else if match == "..." && after < len(text) && text[after] == ':' {
			i = annotationTypeEnd(text, after+1)
		}
	}
	for i < len(text) {
		if strings.HasPrefix(text[i:], "[]") {
			i += 2
		} else if text[i] == '?' {
			i++
		} else {
			break
		}
	}
	after := i
	for after < len(text) && strings.ContainsRune(" \t\r\n", rune(text[after])) {
		after++
	}
	if after < len(text) && text[after] == '|' {
		return annotationTypeEnd(text, after+1)
	}
	return i
}
func splitAnnotationType(text string) (string, string) {
	end := annotationTypeEnd(text, 0)
	return strings.TrimSpace(text[:end]), strings.TrimSpace(text[end:])
}
func namedAnnotationType(text string, field bool) annotationParameter {
	if field {
		if match := annotationVisibilityPattern.FindStringSubmatch(text); match != nil {
			text = text[len(match[0]):]
		}
	}
	end := 0
	if strings.HasPrefix(text, "[") {
		end = annotationBalancedEnd(text, 0)
	} else {
		match := annotationParameterNamePattern.FindString(text)
		if match == "" {
			reject("missing field/parameter name")
		}
		end = len(match)
	}
	name := text[:end]
	raw, _ := splitAnnotationType(strings.TrimSpace(text[end:]))
	if raw == "" {
		reject(fmt.Sprintf("missing type for %s", name))
	}
	return annotationParameter{Name: strings.TrimSuffix(name, "?"), Type: raw, Optional: strings.HasSuffix(name, "?"), Variadic: strings.HasPrefix(name, "...")}
}

// annotationNumeric classifies the enum values used by the generated contract.
// Uninterpreted expressions retain the string classification.
func annotationNumeric(text string) bool {
	if text == "true" || text == "false" {
		return true
	}
	if strings.Contains(text, ".") {
		_, err := strconv.ParseFloat(text, 64)
		return err == nil
	}
	base := 10
	if strings.HasPrefix(strings.ToLower(text), "0x") {
		base, text = 16, text[2:]
	}
	_, err := strconv.ParseInt(text, base, 64)
	return err == nil
}

type annotationDirective struct {
	kind  string
	line  int
	value string
}

// parseAnnotations extracts declaration contracts without executing Lua.
// Unsupported annotation tags and ordinary Lua code do not define contracts.
func parseAnnotations(data []byte, relative string) (result annotationDocument, err error) {
	lineNumber := 1
	// Runs after recoverConversion: every failure is a rejection at lineNumber.
	defer func() {
		if err != nil {
			err = fmt.Errorf("%s:%d: %w", relative, lineNumber, err)
		}
	}()
	defer recoverConversion(&err)
	if !utf8.Valid(data) {
		reject("invalid UTF-8 annotations")
	}
	result = annotationDocument{Path: relative, Aliases: []*annotationAlias{}, Bindings: []annotationBinding{}, Classes: []*annotationClass{}, Enums: []*annotationEnum{}, Functions: []*annotationFunction{}}
	source := strings.ReplaceAll(string(data), "\r\n", "\n")
	source = strings.ReplaceAll(source, "\r", "\n")
	pending := []annotationDirective{}
	var activeClass, bindingClass *annotationClass
	var activeAlias *annotationAlias
	var activeEnum *annotationEnum
	enumBound := false
	bindings := map[string]string{}
	for index, original := range strings.Split(source, "\n") {
		lineNumber = index + 1
		line := strings.TrimSpace(original)
		if tag := annotationTagPattern.FindStringSubmatch(line); tag != nil {
			kind, text := tag[1], tag[2]
			switch kind {
			case "class":
				match := annotationClassPattern.FindStringSubmatch(text)
				if match == nil {
					reject("unrecognized class declaration")
				}
				bases := []string{}
				for _, base := range strings.Split(match[3], ",") {
					if s := strings.TrimSpace(base); s != "" {
						bases = append(bases, s)
					}
				}
				activeClass = &annotationClass{Name: match[2], Bases: bases, Fields: []annotationParameter{}}
				result.Classes = append(result.Classes, activeClass)
				bindingClass = activeClass
				activeAlias = nil
			case "field":
				value := namedAnnotationType(text, true)
				if activeClass == nil {
					reject("field has no preceding class")
				}
				activeClass.Fields = append(activeClass.Fields, value)
			case "alias":
				match := annotationAliasPattern.FindStringSubmatch(text)
				if match == nil {
					reject("unrecognized alias declaration")
				}
				raw, _ := splitAnnotationType(match[2])
				activeAlias = &annotationAlias{Name: match[1], Type: raw, Alternatives: []string{}, Line: lineNumber}
				result.Aliases = append(result.Aliases, activeAlias)
				bindingClass = nil
			case "enum":
				if !plainIdentifier.MatchString(strings.ReplaceAll(text, ".", "_")) {
					reject("unrecognized enum declaration")
				}
				activeEnum = &annotationEnum{Name: text, Values: []annotationEnumValue{}}
				enumBound = false
				result.Enums = append(result.Enums, activeEnum)
				activeAlias = nil
			case "param", "return", "generic", "overload":
				pending = append(pending, annotationDirective{kind: kind, value: text, line: lineNumber})
				activeAlias = nil
			}
			continue
		}
		if strings.HasPrefix(line, "---|") {
			if activeAlias == nil {
				reject("alternative has no preceding alias")
			}
			alt := strings.TrimSpace(line[4:])
			expression, _ := splitAnnotationType(strings.TrimPrefix(alt, ">"))
			if expression == "" {
				reject("missing alternative type")
			}
			activeAlias.Alternatives = append(activeAlias.Alternatives, expression)
			continue
		}
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}
		activeAlias = nil
		if function := annotationFunctionPattern.FindStringSubmatch(line); function != nil {
			name, args := function[1], function[2]
			method := strings.Contains(name, ":")
			rawOwner, member := "", name
			if method {
				cut := strings.LastIndexByte(name, ':')
				rawOwner, member = name[:cut], name[cut+1:]
			} else if cut := strings.LastIndexByte(name, '.'); cut >= 0 {
				rawOwner, member = name[:cut], name[cut+1:]
			}
			owner := rawOwner
			if value, ok := bindings[owner]; ok {
				owner = value
			}
			var ownerValue *string
			if owner != "" {
				ownerValue = new(owner)
			}
			arguments := []string{}
			for _, part := range strings.Split(args, ",") {
				if s := strings.TrimSpace(part); s != "" {
					arguments = append(arguments, s)
				}
			}
			item := &annotationFunction{Name: member, QualifiedName: name, Owner: ownerValue, Method: method, Local: strings.HasPrefix(line, "local function "), Arguments: arguments, Params: []annotationParameter{}, Returns: []annotationReturn{}, Overloads: []string{}, Generics: []string{}}
			for _, ann := range pending {
				lineNumber = ann.line
				switch ann.kind {
				case "param":
					item.Params = append(item.Params, namedAnnotationType(ann.value, false))
				case "return":
					raw, tail := splitAnnotationType(ann.value)
					if raw == "" {
						reject("missing return type")
					}
					label := annotationReturnLabelPattern.FindStringSubmatch(tail)
					variadic := false
					if label != nil {
						variadic = label[1] == "..."
					}
					item.Returns = append(item.Returns, annotationReturn{Type: raw, Variadic: variadic})
				case "generic":
					item.Generics = append(item.Generics, ann.value)
				case "overload":
					item.Overloads = append(item.Overloads, ann.value)
				}
			}
			result.Functions = append(result.Functions, item)
			pending = nil
			bindingClass = nil
			continue
		}
		if activeEnum != nil && enumBound {
			if strings.HasPrefix(line, "}") {
				activeEnum = nil
			} else if value := annotationEnumValuePattern.FindStringSubmatch(line); value != nil {
				raw := strings.TrimSpace(strings.TrimRight(value[2], ","))
				activeEnum.Values = append(activeEnum.Values, annotationEnumValue{Name: value[1], Numeric: annotationNumeric(raw)})
			} else {
				reject("unrecognized enum value")
			}
			continue
		}
		if binding := annotationBindingPattern.FindStringSubmatch(line); binding != nil {
			local, name, value := binding[1] != "", binding[2], binding[3]
			item := annotationBinding{Name: name, Value: value, Local: local, TopLevel: original[0] != ' ' && original[0] != '\t'}
			if bindingClass != nil {
				item.Class = new(bindingClass.Name)
				bindings[name] = bindingClass.Name
				bindingClass = nil
			}
			result.Bindings = append(result.Bindings, item)
			if activeEnum != nil {
				enumBound = true
				if strings.HasSuffix(strings.TrimSpace(value), "}") {
					activeEnum = nil
				}
			}
			pending = nil
			continue
		}
		pending = nil
		bindingClass = nil
	}
	return
}
