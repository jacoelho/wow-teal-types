package wowapi

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

func (c *catalogue) projectSecret(expression string) string {
	projected := []string{}
	names := map[string]string{}
	for name := range c.defs {
		names[ident(name)] = name
	}
	for _, member := range splitTop(expression, '|') {
		value := member
		primitive := c.primitive(member)
		switch {
		case primitive == "number":
			value = "W.NumberLike"
		case primitive == "string" || definitionKind(c.defs[names[member]]) == "enum":
			value = "W.TextLike"
		case primitive == "boolean":
			value = "boolean | OpaqueValue"
		case primitive == "unknown":
			value = "unknown"
		case strings.HasPrefix(member, "{") || strings.HasPrefix(member, "function") || c.isHandle(firstNonempty(names[member], member)):
			value = "OpaqueValue"
		}
		projected = append(projected, splitTop(value, '|')...)
	}
	return strings.Join(uniqueStrings(projected), " | ")
}
func firstNonempty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
func (c *catalogue) projectSignature(signature string) string {
	parts := splitTop(signature, ':')
	if len(parts) < 2 {
		return signature
	}
	results := []string{}
	for _, result := range splitTop(strings.Join(parts[1:], ":"), ',') {
		variadic := strings.HasSuffix(result, "...")
		if variadic {
			result = strings.TrimSuffix(result, "...")
		}
		value := c.projectSecret(result)
		if variadic {
			value += "..."
		}
		results = append(results, value)
	}
	return parts[0] + ": " + strings.Join(results, ", ")
}
func (c *catalogue) signature(item *catalogueFunction, owner string) (string, error) {
	return captureString(func() string { return c.signatureValue(item, owner) })
}
func (c *catalogue) signatureValue(item *catalogueFunction, owner string) string {
	receiverSecret := c.secretStructs[owner] && !item.returnsNeverSecret()
	if item.ResolvedSignature != nil {
		value := *item.ResolvedSignature
		if receiverSecret {
			return c.projectSignature(value)
		}
		return value
	}
	args, returns := []string{}, []string{}
	generics, genericDeclarations := []string{}, []string{}
	if item.Native != nil {
		native := item.Native
		if item.Method {
			args = append(args, "self")
		}
		nativeArgs := native.Arguments
		for index, argument := range nativeArgs {
			if argument.StrideIndex != nil && *argument.StrideIndex != 0 {
				tail := []string{}
				for _, v := range nativeArgs[index:] {
					tail = append(tail, c.nativeTypeValue(v, false))
				}
				args = append(args, strings.Join(uniqueStrings(tail), " | ")+"...")
				break
			}
			raw := argument.Type
			acceptsSecret := stringValue(native.Policy.SecretArguments) == "AllowedWhenTainted" && !enabled(argument.NeverSecret)
			if tuple, ok := c.abiTuples[raw]; ok {
				for _, value := range tuple {
					if acceptsSecret {
						value = c.projectSecret(value)
					}
					args = append(args, value)
				}
				continue
			}
			value := c.nativeTypeValue(argument, acceptsSecret && c.primitive(raw) != "")
			if acceptsSecret && containsString([]string{"TextureAsset", "TextureAssetDisk"}, raw) {
				value += " | W.OpaqueNumber"
			}
			if enabled(argument.Nilable) || argument.HasDefault {
				value = "? " + strings.TrimSuffix(value, " | nil")
			}
			args = append(args, value)
		}
		nativeReturns := append([]NativeParameter{}, native.Returns...)
		absent := c.mayReturnNothing(item)
		if item.returnsNeverSecret() {
			for i := range nativeReturns {
				nativeReturns[i].NeverSecret = boolPointer(true)
			}
		}
		for index, output := range nativeReturns {
			if output.StrideIndex != nil && *output.StrideIndex != 0 {
				tail := []string{}
				for _, value := range nativeReturns[index:] {
					value.Nilable = boolPointer(enabled(value.Nilable) || absent)
					tail = append(tail, c.nativeTypeValue(value, receiverSecret || c.resultSecret(item)))
				}
				returns = append(returns, strings.Join(uniqueStrings(tail), " | ")+"...")
				break
			}
			raw := output.Type
			if tuple, ok := c.abiTuples[raw]; ok {
				for _, value := range tuple {
					if !enabled(output.NeverSecret) && (receiverSecret || c.resultSecret(item) || output.secretResult()) {
						value = c.projectSecret(value)
					}
					if absent {
						value += " | nil"
					}
					returns = append(returns, value)
				}
				continue
			}
			value := c.nativeTypeValue(output, receiverSecret || c.resultSecret(item))
			if absent && !containsString(splitTop(value, '|'), "nil") {
				value += " | nil"
			}
			returns = append(returns, value)
		}
	} else {
		annotation := item.clone().Annotation
		if annotation == nil {
			annotation = &AnnotationFunction{}
		}
		declarations := []string{}
		for _, group := range annotation.Generics {
			declarations = append(declarations, splitTop(group, ',')...)
		}
		for _, decl := range declarations {
			name, bound, separator := strings.Cut(decl, ":")
			name, bound = strings.TrimSpace(name), strings.TrimSpace(bound)
			if separator && bound != "table" && (containsString([]string{"string", "number", "boolean"}, bound) || strings.Contains(bound, "OptionsGetter")) {
				pattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(name) + `\b`)
				for i := range annotation.Params {
					annotation.Params[i].Type = pattern.ReplaceAllString(annotation.Params[i].Type, bound)
				}
				for i := range annotation.Returns {
					annotation.Returns[i].Type = pattern.ReplaceAllString(annotation.Returns[i].Type, bound)
				}
				continue
			}
			// A caller cannot choose a result-only generic. Preserve its declared
			// bound, or expose an opaque result when the source supplies none.
			pattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(name) + `\b`)
			inputBound := false
			for _, parameter := range annotation.Params {
				inputBound = inputBound || pattern.MatchString(parameter.Type)
			}
			if !inputBound {
				replacement := "unknown"
				if separator {
					replacement = bound
				}
				for i := range annotation.Returns {
					annotation.Returns[i].Type = pattern.ReplaceAllString(annotation.Returns[i].Type, replacement)
				}
				continue
			}
			generics = append(generics, name)
			constraint := ""
			if bound == "table" {
				constraint = "CatalogueTable"
			} else if separator {
				constraint = c.resolve(bound)
			}
			declaration := ident(name)
			if constraint != "" {
				declaration += " is " + constraint
			}
			genericDeclarations = append(genericDeclarations, declaration)
		}
		if item.Method {
			args = append(args, "self")
		}
		params := map[string]AnnotationParameter{}
		for _, param := range annotation.Params {
			params[param.Name] = param
		}
		for _, name := range annotation.Arguments {
			param, present := params[name]
			value := "unknown"
			if !present {
				var symbol *string
				if annotation.QualifiedName != "" {
					symbol = &annotation.QualifiedName
				}
				c.notes = append(c.notes, RepresentationNote{Symbol: symbol, Rule: "source parameter has no type annotation", Parameter: name, Source: item.Source})
			} else {
				value = c.tealValue(param.Type, owner, generics)
			}
			if name == "..." {
				value += "..."
			} else if present && (param.Optional || containsString(splitTop(value, '|'), "nil") || containsString(splitTop(param.Type, '|'), "nil")) {
				members := []string{}
				for _, member := range splitTop(value, '|') {
					if member != "nil" {
						members = append(members, member)
					}
				}
				value = "? " + strings.Join(members, " | ")
			}
			args = append(args, value)
		}
		for _, output := range annotation.Returns {
			raw := output.Type
			if raw == "..." {
				value := "unknown..."
				if receiverSecret {
					value = "OpaqueValue..."
				}
				returns = append(returns, value)
			} else {
				value := c.tealValue(raw, owner, generics)
				if receiverSecret && !output.NeverSecret {
					value = c.projectSecret(value)
				}
				if output.Variadic {
					value += "..."
				}
				returns = append(returns, value)
			}
		}
		if len(annotation.Returns) == 0 && item.Source == "questie:Data/Wiki.lua" {
			value := "unknown..."
			if receiverSecret {
				value = "OpaqueValue..."
			}
			returns = append(returns, value)
			var symbol *string
			if annotation.QualifiedName != "" {
				symbol = &annotation.QualifiedName
			}
			c.notes = append(c.notes, RepresentationNote{Symbol: symbol, Rule: "wiki stub has no result annotation; result tuple remains dynamic", Source: item.Source})
		}
	}
	signature := "function"
	if len(genericDeclarations) > 0 {
		signature += "<" + strings.Join(genericDeclarations, ", ") + ">"
	}
	signature += "(" + strings.Join(args, ", ") + ")"
	if len(returns) > 0 {
		signature += ": " + strings.Join(returns, ", ")
	}
	return signature
}
func (c *catalogue) signatures(item *catalogueFunction, owner string) (values []string, err error) {
	defer captureError(&err)
	return c.signaturesValue(item, owner), nil
}

type signatureChoice struct {
	expression string
	contract   *catalogueFunction
}

func (c *catalogue) signaturesValue(item *catalogueFunction, owner string) []string {
	values := []string{}
	for _, choice := range c.signatureChoices(item, owner) {
		values = append(values, choice.expression)
	}
	return uniqueStrings(values)
}
func (c *catalogue) signatureChoices(item *catalogueFunction, owner string) []signatureChoice {
	values := []signatureChoice{}
	variants := append([]*catalogueFunction{item}, item.Variants...)
	for _, variant := range variants {
		choices := [][]string{}
		count := 1
		var params []AnnotationParameter
		if variant.Annotation != nil {
			params = variant.Annotation.Params
		}
		for _, param := range params {
			raw := param.Type
			members := splitTop(raw, '|')
			tables := []string{}
			for _, value := range members {
				if !containsString([]string{"string", "table", "math"}, value) && containsString([]string{"record", "interface"}, definitionKind(c.defs[value])) {
					tables = append(tables, value)
				}
			}
			options := []string{raw}
			if len(tables) > 1 && !containsString(members, "table") {
				options = []string{}
				for _, table := range tables {
					ordinary := []string{}
					for _, member := range members {
						if !containsString(tables, member) {
							ordinary = append(ordinary, member)
						}
					}
					options = append(options, strings.Join(append(ordinary, table), " | "))
				}
			}
			choices = append(choices, options)
			count *= len(options)
		}
		if count > 32 {
			qualified := ""
			if variant.Annotation != nil {
				qualified = variant.Annotation.QualifiedName
			}
			unsupported(qualified, "more than 32 input overload combinations")
		}
		var emit func(int, []string)
		emit = func(index int, combination []string) {
			if index < len(choices) {
				for _, choice := range choices[index] {
					emit(index+1, append(combination, choice))
				}
				return
			}
			expanded := variant.clone()
			if expanded.Annotation != nil {
				for i := range expanded.Annotation.Params {
					expanded.Annotation.Params[i].Type = combination[i]
				}
			}
			values = append(values, signatureChoice{expression: c.signatureValue(expanded, owner), contract: variant})
		}
		emit(0, nil)
	}
	var overloads []string
	if item.Annotation != nil {
		overloads = item.Annotation.Overloads
	}
	for _, raw := range overloads {
		converted := c.tealValue(raw, owner, nil)
		signature := converted
		if def := c.defs[converted]; definitionKind(def) == "callback" {
			signature = def.Signature
		}
		if c.secretStructs[owner] && !item.returnsNeverSecret() {
			signature = c.projectSignature(signature)
		}
		values = append(values, signatureChoice{expression: signature, contract: item})
	}
	return values
}
func (c *catalogue) prepareGlobals() {
	for _, name := range append([]string{}, c.globalOrder...) {
		item := c.globals[name]
		if item == nil || item.TypeName != "" {
			continue
		}
		choices := c.signatureChoices(item, "")
		if len(uniqueSignatureChoices(choices)) < 2 {
			continue
		}
		key := "Callable_" + ident(name)
		record := c.record(key, item.Source)
		calls := []*catalogueFunction{}
		for _, choice := range choices {
			call := strings.Replace(choice.expression, "(", "(self, ", 1)
			call = strings.ReplaceAll(call, "self, )", "self)")
			calls = append(calls, &catalogueFunction{ResolvedSignature: &call, CallableContract: choice.contract, Method: true, PolicySymbol: name})
		}
		calls[0].Variants = calls[1:]
		record.Methods["metamethod __call"] = calls[0]
		c.setGlobal(name, &catalogueFunction{TypeName: key, Source: item.Source, CallableContract: item})
	}
}
func (c *catalogue) reconcileInheritance() {
	type expanded struct {
		fields  map[string]*catalogueField
		methods map[string]*catalogueFunction
	}
	cache := map[string]expanded{}
	var expand func(string, []string) expanded
	expand = func(name string, stack []string) expanded {
		if result, ok := cache[name]; ok {
			return result
		}
		if containsString(stack, name) {
			reject(fmt.Sprintf("cyclic inheritance: %s", strings.Join(append(stack, name), " -> ")))
		}
		record := c.defs[name]
		fields, methods := map[string]*catalogueField{}, map[string]*catalogueFunction{}
		for _, base := range record.Bases {
			visited := map[string]bool{}
			for c.defs[base] == nil && c.aliases[base] != nil {
				if visited[base] {
					reject(fmt.Sprintf("cyclic base alias: %s", base))
				}
				visited[base] = true
				base = c.aliases[base].Expression
			}
			if c.defs[base] == nil {
				reject(fmt.Sprintf("%s: unknown base %s", name, base))
			}
			inherited := expand(base, append(append([]string{}, stack...), name))
			for k, v := range inherited.fields {
				fields[k] = v
			}
			for k, v := range inherited.methods {
				methods[k] = v
			}
		}
		inheritedFields := map[string]bool{}
		for k := range fields {
			inheritedFields[k] = true
		}
		record.InheritedFields = inheritedFields
		for k, v := range record.Fields {
			fields[k] = v
		}
		for _, method := range sortedKeys(record.Methods) {
			item := record.Methods[method]
			if previous := methods[method]; previous != nil && previous.Native != nil && item.Native == nil {
				c.notes = append(c.notes, RepresentationNote{Symbol: stringPointer(name + "." + method), Rule: "inherit canonical native ancestor contract"})
				continue
			}
			methods[method] = item
		}
		result := expanded{fields, methods}
		cache[name] = result
		return result
	}
	for _, name := range append([]string{}, c.defOrder...) {
		if containsString([]string{"record", "interface"}, definitionKind(c.defs[name])) {
			expand(name, nil)
		}
	}
	for _, name := range sortedKeys(cache) {
		record := c.defs[name]
		methods := map[string]*catalogueFunction{}
		for key, value := range record.Methods {
			if cache[name].methods[key] == value {
				methods[key] = value
			}
		}
		record.Methods = methods
		fields := map[string]*catalogueField{}
		for key, value := range record.Fields {
			if !record.InheritedFields[key] || value.Native != nil {
				fields[key] = value
			}
		}
		record.Fields = fields
		if record.NativeObject {
			record.Kind, record.Opaque = "interface", true
		}
	}
	for _, name := range c.defOrder {
		record := c.defs[name]
		for _, base := range definitionBases(record) {
			value := c.defs[base]
			if definitionKind(value) == "record" && len(value.Fields) == 0 && !value.Opaque {
				value.Kind = "interface"
			}
		}
	}
}
func (c *catalogue) prepareSecrecy() {
	var targets func(string, map[string]bool) map[string]bool
	targets = func(expression string, seen map[string]bool) map[string]bool {
		out := map[string]bool{}
		if seen[expression] || strings.HasPrefix(expression, "fun(") || strings.HasPrefix(expression, "function(") {
			return out
		}
		next := map[string]bool{}
		setAdd(next, seen)
		next[expression] = true
		for _, token := range dottedIdentifiers.FindAllString(expression, -1) {
			if c.primitive(token) != "" || containsString([]string{"nil", "table", "self"}, token) {
				continue
			}
			if c.aliases[token] != nil && c.defs[token] == nil {
				setAdd(out, targets(c.aliases[token].Expression, next))
			} else if record := c.defs[token]; record != nil && !c.isHandle(token) && containsString([]string{"record", "interface", "converted_record"}, record.Kind) {
				out[token] = true
			}
		}
		return out
	}
	seed := func(item NativeParameter, inherited bool) map[string]bool {
		if enabled(item.NeverSecret) || enabled(item.NeverSecretContents) {
			return map[string]bool{}
		}
		if inherited || item.secretResult() {
			return targets(firstNonempty(item.InnerType, item.Type), map[string]bool{})
		}
		return map[string]bool{}
	}
	c.secretStructs = map[string]bool{}
	functions := []*catalogueFunction{}
	for _, name := range c.globalOrder {
		functions = append(functions, c.globals[name])
	}
	for _, name := range c.defOrder {
		record := c.defs[name]
		if record == nil {
			continue
		}
		for _, method := range sortedKeys(record.Methods) {
			functions = append(functions, record.Methods[method])
		}
		for _, field := range record.Fields {
			if field.Native != nil {
				setAdd(c.secretStructs, seed(*field.Native, false))
			}
		}
	}
	for _, item := range functions {
		if item.Native != nil && !item.returnsNeverSecret() {
			for _, output := range item.Native.Returns {
				setAdd(c.secretStructs, seed(output, c.resultSecret(item)))
			}
		}
	}
	for {
		previous := map[string]bool{}
		setAdd(previous, c.secretStructs)
		for _, name := range append([]string{}, c.defOrder...) {
			record := c.defs[name]
			if record == nil || c.isHandle(name) {
				continue
			}
			bases := map[string]bool{}
			for _, base := range record.Bases {
				setAdd(bases, targets(base, map[string]bool{}))
			}
			for base := range bases {
				if c.secretStructs[base] {
					c.secretStructs[name] = true
				}
			}
			if !c.secretStructs[name] {
				continue
			}
			setAdd(c.secretStructs, bases)
			for _, field := range record.Fields {
				if field.Native != nil {
					setAdd(c.secretStructs, seed(*field.Native, true))
				} else {
					expression := field.Type
					if field.Resolved != nil {
						expression = *field.Resolved
					} else if record.Kind != "converted_record" {
						expression = c.tealValue(field.Type, name, nil)
					}
					setAdd(c.secretStructs, targets(expression, map[string]bool{}))
				}
			}
			for _, method := range sortedKeys(record.Methods) {
				item := record.Methods[method]
				if item.returnsNeverSecret() {
					continue
				}
				if item.Native != nil {
					for _, output := range item.Native.Returns {
						setAdd(c.secretStructs, seed(output, true))
					}
				} else if item.Annotation != nil {
					for _, output := range item.Annotation.Returns {
						if !output.NeverSecret {
							setAdd(c.secretStructs, targets(output.Type, map[string]bool{}))
						}
					}
				}
				for _, signature := range c.signaturesValue(item, name) {
					parts := splitTop(signature, ':')
					if len(parts) > 1 {
						setAdd(c.secretStructs, targets(strings.Join(parts[1:], ":"), map[string]bool{}))
					}
				}
			}
		}
		if setEqual(previous, c.secretStructs) {
			break
		}
	}
}
func sortedSet(values map[string]bool) []string {
	out := []string{}
	for k, v := range values {
		if v {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}
