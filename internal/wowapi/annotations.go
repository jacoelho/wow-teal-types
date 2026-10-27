package wowapi

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
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
			reject("unrecognized type prefix: %s", text[i:min(i+40, len(text))])
		}
		i += len(match)
		after := i
		for after < len(text) && strings.ContainsRune(" \t\r\n", rune(text[after])) {
			after++
		}
		if after < len(text) && text[after] == '<' {
			i = annotationBalancedEnd(text, after)
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
func namedAnnotationType(text string, field bool) M {
	var visibility any
	if field {
		if match := annotationVisibility.FindStringSubmatch(text); match != nil {
			visibility = match[1]
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
		reject("missing type for %s", name)
	}
	return M{"name": strings.TrimSuffix(name, "?"), "type": raw, "optional": strings.HasSuffix(name, "?"), "variadic": strings.HasPrefix(name, "..."), "description": description, "visibility": visibility}
}
func annotationLiteral(text string) any {
	switch text {
	case "true":
		return true
	case "false":
		return false
	case "nil":
		return nil
	}
	var result any = text
	func() {
		defer func() {
			if r := recover(); r != nil {
				if _, ok := r.(conversionError); !ok {
					panic(r)
				}
			}
		}()
		if strings.HasPrefix(text, "\"") || strings.HasPrefix(text, "'") {
			result, _ = quotedLiteral(text)
			return
		}
		if strings.Contains(text, ".") {
			if n, e := strconv.ParseFloat(text, 64); e == nil {
				result = n
			}
		} else {
			base := 10
			value := text
			if strings.HasPrefix(strings.ToLower(text), "0x") {
				base = 16
				value = text[2:]
			}
			if n, e := strconv.ParseInt(value, base, 64); e == nil {
				result = n
			}
		}
	}()
	return result
}

// ParseAnnotations retains declaration metadata and unsupported annotations without executing Lua.
func ParseAnnotations(data []byte, relative string) (result M, err error) {
	defer captureError(&err)
	if !utf8.Valid(data) {
		reject("%s: invalid UTF-8 annotations", relative)
	}
	result = M{"path": relative}
	for _, key := range []string{"aliases", "classes", "functions", "enums", "bindings", "annotations", "diagnostics"} {
		result[key] = []any{}
	}
	source := strings.ReplaceAll(string(data), "\r\n", "\n")
	source = strings.ReplaceAll(source, "\r", "\n")
	pending := []any{}
	var activeClass, bindingClass, activeAlias, activeEnum M
	bindings := map[string]string{}
	classes := map[string]M{}
	diagnostic := func(number any, code, message, raw string) {
		result["diagnostics"] = append(asList(result["diagnostics"]), M{"line": number, "code": code, "message": message, "raw": raw})
	}
	for index, original := range strings.Split(source, "\n") {
		number := index + 1
		line := strings.TrimSpace(original)
		if tag := annotationTag.FindStringSubmatch(line); tag != nil {
			kind, text := tag[1], tag[2]
			annotation := M{"kind": kind, "value": text, "line": number, "raw": original}
			result["annotations"] = append(asList(result["annotations"]), annotation)
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
					bases := []any{}
					for _, base := range strings.Split(match[3], ",") {
						if s := strings.TrimSpace(base); s != "" {
							bases = append(bases, s)
						}
					}
					var attrs any
					if match[1] != "" {
						attrs = match[1]
					}
					activeClass = M{"name": match[2], "bases": bases, "attributes": attrs, "fields": []any{}, "methods": []any{}, "binding": nil, "line": number, "raw": original}
					classes[match[2]] = activeClass
					result["classes"] = append(asList(result["classes"]), activeClass)
					bindingClass = activeClass
					activeAlias = nil
				case "field":
					value := namedAnnotationType(text, true)
					value["line"], value["raw"] = number, original
					if activeClass == nil {
						diagnostic(number, "orphan_field", "field has no preceding class", original)
					} else {
						activeClass["fields"] = append(asList(activeClass["fields"]), value)
					}
				case "alias":
					match := annotationAlias.FindStringSubmatch(text)
					if match == nil {
						reject("unrecognized alias declaration")
					}
					raw, description := splitAnnotationType(match[2])
					activeAlias = M{"name": match[1], "type": raw, "description": description, "alternatives": []any{}, "line": number, "raw": original}
					result["aliases"] = append(asList(result["aliases"]), activeAlias)
					bindingClass = nil
				case "enum":
					activeEnum = M{"name": text, "binding": nil, "values": []any{}, "line": number, "raw": original}
					result["enums"] = append(asList(result["enums"]), activeEnum)
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
					activeAlias["alternatives"] = append(asList(activeAlias["alternatives"]), M{"type": expression, "description": description, "default": isDefault, "line": number, "raw": original})
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
			var ownerValue, rawOwnerValue any
			if owner != "" {
				ownerValue = owner
			}
			if rawOwner != "" {
				rawOwnerValue = rawOwner
			}
			arguments := []any{}
			for _, part := range strings.Split(args, ",") {
				if s := strings.TrimSpace(part); s != "" {
					arguments = append(arguments, s)
				}
			}
			item := M{"name": member, "qualified_name": name, "owner": ownerValue, "raw_owner": rawOwnerValue, "method": method, "arguments": arguments, "params": []any{}, "returns": []any{}, "overloads": []any{}, "generics": []any{}, "annotations": pending, "line": number, "raw": original}
			for _, v := range pending {
				ann := asMap(v)
				kind, text := asString(ann["kind"]), asString(ann["value"])
				func() {
					defer func() {
						if r := recover(); r != nil {
							if e, ok := r.(conversionError); ok {
								diagnostic(ann["line"], "function_annotation_parse", e.Error(), asString(ann["raw"]))
							} else {
								panic(r)
							}
						}
					}()
					switch kind {
					case "param":
						value := namedAnnotationType(text, false)
						value["line"], value["raw"] = ann["line"], ann["raw"]
						item["params"] = append(asList(item["params"]), value)
					case "return":
						raw, tail := splitAnnotationType(text)
						label := annotationReturnLabel.FindStringSubmatch(tail)
						var name any
						description := tail
						variadic := false
						if label != nil {
							name = label[1]
							description = label[2]
							variadic = label[1] == "..."
						}
						item["returns"] = append(asList(item["returns"]), M{"type": raw, "name": name, "variadic": variadic, "description": description, "line": ann["line"], "raw": ann["raw"]})
					case "generic":
						item["generics"] = append(asList(item["generics"]), text)
					case "overload":
						item["overloads"] = append(asList(item["overloads"]), text)
					}
				}()
			}
			result["functions"] = append(asList(result["functions"]), item)
			if cls := classes[owner]; cls != nil {
				cls["methods"] = append(asList(cls["methods"]), item)
			}
			pending = []any{}
			bindingClass = nil
			continue
		}
		if activeEnum != nil && activeEnum["binding"] != nil {
			if strings.HasPrefix(line, "}") {
				activeEnum = nil
			} else if value := annotationEnumValue.FindStringSubmatch(line); value != nil {
				raw := strings.TrimSpace(strings.TrimRight(value[2], ","))
				activeEnum["values"] = append(asList(activeEnum["values"]), M{"name": value[1], "value": annotationLiteral(raw), "raw_value": raw, "line": number, "raw": original})
			} else {
				diagnostic(number, "enum_value_parse", "unrecognized enum value", original)
			}
			continue
		}
		if binding := annotationBinding.FindStringSubmatch(line); binding != nil {
			local, name, value := binding[1] != "", binding[2], binding[3]
			item := M{"name": name, "value": value, "local": local, "line": number, "raw": original, "class": nil}
			if bindingClass != nil {
				item["class"] = bindingClass["name"]
				bindingClass["binding"] = name
				bindings[name] = asString(bindingClass["name"])
				bindingClass = nil
			}
			result["bindings"] = append(asList(result["bindings"]), item)
			if activeEnum != nil {
				activeEnum["binding"] = name
				if strings.HasSuffix(strings.TrimSpace(value), "}") {
					activeEnum = nil
				}
			}
			pending = []any{}
			continue
		}
		if strings.Contains(line, "function") {
			diagnostic(number, "function_parse", "unrecognized function declaration", original)
		} else if !strings.HasPrefix(line, "end") && !strings.HasPrefix(line, "}") && !strings.HasPrefix(line, "{") {
			diagnostic(number, "unparsed_code", "code retained without declaration interpretation", original)
		}
		pending = []any{}
		bindingClass = nil
	}
	for _, value := range pending {
		ann := asMap(value)
		diagnostic(ann["line"], "orphan_annotation", "annotation not followed by a declaration", asString(ann["raw"]))
	}
	return
}
func parseAnnotationFile(path, relative string) (M, error) {
	data, e := os.ReadFile(path)
	if e != nil {
		return nil, e
	}
	return ParseAnnotations(data, relative)
}
