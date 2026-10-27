package wowapi

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"wow-teal-types/internal/wowapi/literal"
)

const annotationName = `[A-Za-z_][A-Za-z_0-9.\-]*`

var annotationFunction = regexp.MustCompile(`^(?:local\s+)?function\s+(` + annotationName + `(?::` + annotationName + `)?)\s*\(([^)]*)\)`)
var annotationBinding = regexp.MustCompile(`^(local\s+)?(` + annotationName + `)\s*=\s*(.*)$`)
var annotationTag = regexp.MustCompile(`^---@([A-Za-z]+)\s*(.*)$`)
var annotationKnown = wordSet("meta alias class field param return enum generic overload deprecated nodiscard diagnostic see paramsig type name protected")
var annotationClass = regexp.MustCompile(`^(?:\(([^)]*)\)\s*)?(` + annotationName + `)(?:\s*:\s*(.*))?$`)
var annotationAlias = regexp.MustCompile(`^(` + annotationName + `)(?:\s+(.*))?$`)
var annotationTypePrefix = regexp.MustCompile(`^(?:\.\.\.|-?(?:0x[0-9A-Fa-f]+|\d+(?:\.\d+)?)|` + annotationName + `)`)
var annotationParameterName = regexp.MustCompile(`^(?:\.\.\.|` + annotationName + `)\??`)
var annotationVisibility = regexp.MustCompile(`^(private|protected|public|package)\s+`)
var annotationReturnLabel = regexp.MustCompile(`^(\.\.\.|[A-Za-z_][A-Za-z_0-9]*)(?:\s+(.*))?$`)
var annotationEnumValue = regexp.MustCompile(`^(` + annotationName + `|\[.*?\])\s*=\s*(.*?)(?:,)?\s*(?:--.*)?$`)

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
		match := annotationTypePrefix.FindString(text[i:])
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
func SplitAnnotationType(text string) (raw, tail string, err error) {
	defer captureError(&err)
	raw, tail = splitAnnotationType(text)
	return
}
func namedAnnotationType(text string, field bool) AnnotationParameter {
	var visibility *string
	if field {
		if match := annotationVisibility.FindStringSubmatch(text); match != nil {
			visibility = stringPointer(match[1])
			text = text[len(match[0]):]
		}
	}
	end := 0
	if strings.HasPrefix(text, "[") {
		end = annotationBalancedEnd(text, 0)
	} else {
		match := annotationParameterName.FindString(text)
		if match == "" {
			reject("missing field/parameter name")
		}
		end = len(match)
	}
	name := text[:end]
	raw, description := splitAnnotationType(strings.TrimSpace(text[end:]))
	if raw == "" {
		reject(fmt.Sprintf("missing type for %s", name))
	}
	return AnnotationParameter{Name: strings.TrimSuffix(name, "?"), Type: raw, Optional: strings.HasSuffix(name, "?"), Variadic: strings.HasPrefix(name, "..."), Description: description, Visibility: visibility}
}
func annotationLiteral(text string) AnnotationConstant {
	switch text {
	case "true":
		return AnnotationConstant{kind: constantBoolean, boolean: true}
	case "false":
		return AnnotationConstant{kind: constantBoolean}
	case "nil":
		return AnnotationConstant{}
	}
	result := AnnotationConstant{kind: constantString, text: text}
	func() {
		defer func() {
			if r := recover(); r != nil {
				if _, ok := r.(conversionError); !ok {
					panic(r)
				}
			}
		}()
		if strings.HasPrefix(text, "\"") || strings.HasPrefix(text, "'") {
			result.text, _ = quotedLiteral(text)
			return
		}
		if strings.Contains(text, ".") {
			if n, e := strconv.ParseFloat(text, 64); e == nil {
				result = AnnotationConstant{kind: constantFloat, number: literal.Float(n)}
			}
		} else {
			base := 10
			value := text
			if strings.HasPrefix(strings.ToLower(text), "0x") {
				base = 16
				value = text[2:]
			}
			if n, e := strconv.ParseInt(value, base, 64); e == nil {
				result = AnnotationConstant{kind: constantInteger, integer: n}
			}
		}
	}()
	return result
}

// ParseAnnotations retains declaration metadata and unsupported annotations without executing Lua.
func ParseAnnotations(data []byte, relative string) (result AnnotationDocument, err error) {
	defer captureError(&err)
	if !utf8.Valid(data) {
		reject(fmt.Sprintf("%s: invalid UTF-8 annotations", relative))
	}
	result = AnnotationDocument{Path: relative, Aliases: []*AnnotationAlias{}, Annotations: []AnnotationTag{}, Bindings: []AnnotationBinding{}, Classes: []*AnnotationClass{}, Diagnostics: []AnnotationDiagnostic{}, Enums: []*AnnotationEnum{}, Functions: []*AnnotationFunction{}}
	source := strings.ReplaceAll(string(data), "\r\n", "\n")
	source = strings.ReplaceAll(source, "\r", "\n")
	pending := []AnnotationTag{}
	var activeClass, bindingClass *AnnotationClass
	var activeAlias *AnnotationAlias
	var activeEnum *AnnotationEnum
	bindings := map[string]string{}
	classes := map[string]*AnnotationClass{}
	diagnostic := func(number int, code, message, raw string) {
		result.Diagnostics = append(result.Diagnostics, AnnotationDiagnostic{Line: number, Code: code, Message: message, Raw: raw})
	}
	for index, original := range strings.Split(source, "\n") {
		number := index + 1
		line := strings.TrimSpace(original)
		if tag := annotationTag.FindStringSubmatch(line); tag != nil {
			kind, text := tag[1], tag[2]
			annotation := AnnotationTag{Kind: kind, Value: text, Line: number, Raw: original}
			result.Annotations = append(result.Annotations, annotation)
			if !annotationKnown[kind] {
				diagnostic(number, "unknown_annotation", kind, original)
			}
			func() {
				defer func() {
					if r := recover(); r != nil {
						if e, ok := r.(conversionError); ok {
							diagnostic(number, "annotation_parse", e.Error(), original)
							pending = append(pending, annotation)
						} else {
							panic(r)
						}
					}
				}()
				switch kind {
				case "class":
					match := annotationClass.FindStringSubmatch(text)
					if match == nil {
						reject("unrecognized class declaration")
					}
					bases := []string{}
					for _, base := range strings.Split(match[3], ",") {
						if s := strings.TrimSpace(base); s != "" {
							bases = append(bases, s)
						}
					}
					var attrs *string
					if match[1] != "" {
						attrs = stringPointer(match[1])
					}
					activeClass = &AnnotationClass{Name: match[2], Bases: bases, Attributes: attrs, Fields: []AnnotationParameter{}, Methods: []*AnnotationFunction{}, Line: number, Raw: original}
					classes[match[2]] = activeClass
					result.Classes = append(result.Classes, activeClass)
					bindingClass = activeClass
					activeAlias = nil
				case "field":
					value := namedAnnotationType(text, true)
					value.Line, value.Raw = number, original
					if activeClass == nil {
						diagnostic(number, "orphan_field", "field has no preceding class", original)
					} else {
						activeClass.Fields = append(activeClass.Fields, value)
					}
				case "alias":
					match := annotationAlias.FindStringSubmatch(text)
					if match == nil {
						reject("unrecognized alias declaration")
					}
					raw, description := splitAnnotationType(match[2])
					activeAlias = &AnnotationAlias{Name: match[1], Type: raw, Description: description, Alternatives: []AnnotationAlternative{}, Line: number, Raw: original}
					result.Aliases = append(result.Aliases, activeAlias)
					bindingClass = nil
				case "enum":
					activeEnum = &AnnotationEnum{Name: text, Values: []AnnotationEnumValue{}, Line: number, Raw: original}
					result.Enums = append(result.Enums, activeEnum)
					activeAlias = nil
				default:
					if kind != "meta" && kind != "diagnostic" && kind != "see" {
						pending = append(pending, annotation)
						activeAlias = nil
					}
				}
			}()
			continue
		}
		if strings.HasPrefix(line, "---|") {
			if activeAlias == nil {
				diagnostic(number, "orphan_alternative", "alternative has no preceding alias", original)
			} else {
				func() {
					defer func() {
						if r := recover(); r != nil {
							if e, ok := r.(conversionError); ok {
								diagnostic(number, "alternative_parse", e.Error(), original)
							} else {
								panic(r)
							}
						}
					}()
					alt := strings.TrimSpace(line[4:])
					isDefault := strings.HasPrefix(alt, ">")
					expression, description := splitAnnotationType(strings.TrimPrefix(alt, ">"))
					activeAlias.Alternatives = append(activeAlias.Alternatives, AnnotationAlternative{Type: expression, Description: description, Default: isDefault, Line: number, Raw: original})
				}()
			}
			continue
		}
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}
		activeAlias = nil
		if function := annotationFunction.FindStringSubmatch(line); function != nil {
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
			var ownerValue, rawOwnerValue *string
			if owner != "" {
				ownerValue = stringPointer(owner)
			}
			if rawOwner != "" {
				rawOwnerValue = stringPointer(rawOwner)
			}
			arguments := []string{}
			for _, part := range strings.Split(args, ",") {
				if s := strings.TrimSpace(part); s != "" {
					arguments = append(arguments, s)
				}
			}
			item := &AnnotationFunction{Name: member, QualifiedName: name, Owner: ownerValue, RawOwner: rawOwnerValue, Method: method, Arguments: arguments, Params: []AnnotationParameter{}, Returns: []AnnotationReturn{}, Overloads: []string{}, Generics: []string{}, Annotations: pending, Line: number, Raw: original}
			for _, v := range pending {
				ann := v
				kind, text := ann.Kind, ann.Value
				func() {
					defer func() {
						if r := recover(); r != nil {
							if e, ok := r.(conversionError); ok {
								diagnostic(ann.Line, "function_annotation_parse", e.Error(), ann.Raw)
							} else {
								panic(r)
							}
						}
					}()
					switch kind {
					case "param":
						value := namedAnnotationType(text, false)
						value.Line, value.Raw = ann.Line, ann.Raw
						item.Params = append(item.Params, value)
					case "return":
						raw, tail := splitAnnotationType(text)
						label := annotationReturnLabel.FindStringSubmatch(tail)
						var name *string
						description := tail
						variadic := false
						if label != nil {
							name = stringPointer(label[1])
							description = label[2]
							variadic = label[1] == "..."
						}
						item.Returns = append(item.Returns, AnnotationReturn{Type: raw, Name: name, Variadic: variadic, Description: stringPointer(description), Line: ann.Line, Raw: ann.Raw})
					case "generic":
						item.Generics = append(item.Generics, text)
					case "overload":
						item.Overloads = append(item.Overloads, text)
					}
				}()
			}
			result.Functions = append(result.Functions, item)
			if cls := classes[owner]; cls != nil {
				cls.Methods = append(cls.Methods, item)
			}
			pending = []AnnotationTag{}
			bindingClass = nil
			continue
		}
		if activeEnum != nil && activeEnum.Binding != nil {
			if strings.HasPrefix(line, "}") {
				activeEnum = nil
			} else if value := annotationEnumValue.FindStringSubmatch(line); value != nil {
				raw := strings.TrimSpace(strings.TrimRight(value[2], ","))
				activeEnum.Values = append(activeEnum.Values, AnnotationEnumValue{Name: value[1], Value: annotationLiteral(raw), RawValue: raw, Line: number, Raw: original})
			} else {
				diagnostic(number, "enum_value_parse", "unrecognized enum value", original)
			}
			continue
		}
		if binding := annotationBinding.FindStringSubmatch(line); binding != nil {
			local, name, value := binding[1] != "", binding[2], binding[3]
			item := AnnotationBinding{Name: name, Value: value, Local: local, Line: number, Raw: original}
			if bindingClass != nil {
				item.Class = stringPointer(bindingClass.Name)
				bindingClass.Binding = stringPointer(name)
				bindings[name] = bindingClass.Name
				bindingClass = nil
			}
			result.Bindings = append(result.Bindings, item)
			if activeEnum != nil {
				activeEnum.Binding = stringPointer(name)
				if strings.HasSuffix(strings.TrimSpace(value), "}") {
					activeEnum = nil
				}
			}
			pending = []AnnotationTag{}
			continue
		}
		if strings.Contains(line, "function") {
			diagnostic(number, "function_parse", "unrecognized function declaration", original)
		} else if !strings.HasPrefix(line, "end") && !strings.HasPrefix(line, "}") && !strings.HasPrefix(line, "{") {
			diagnostic(number, "unparsed_code", "code retained without declaration interpretation", original)
		}
		pending = []AnnotationTag{}
		bindingClass = nil
	}
	for _, value := range pending {
		ann := value
		diagnostic(ann.Line, "orphan_annotation", "annotation not followed by a declaration", ann.Raw)
	}
	return
}
func parseAnnotationFile(path, relative string) (AnnotationDocument, error) {
	data, e := os.ReadFile(path)
	if e != nil {
		return AnnotationDocument{}, e
	}
	return ParseAnnotations(data, relative)
}
