package wowapi

import (
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
		case primitive == "string" || asString(asMap(c.defs[names[member]])["kind"]) == "enum":
			value = "W.TextLike"
		case primitive == "boolean":
			value = "boolean | OpaqueValue"
		case primitive == "any" || strings.HasPrefix(member, "{") || strings.HasPrefix(member, "function") || c.isHandle(firstNonempty(names[member], member)):
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
func (c *catalogue) signature(item M, owner string) (string, error) {
	return captureString(func() string { return c.signatureValue(item, owner) })
}
func (c *catalogue) signatureValue(item M, owner string) string {
	receiverSecret := c.secretStructs[owner] && !truth(item["ReturnsNeverSecret"])
	if signature, ok := item["resolved_signature"]; ok {
		value := asString(signature)
		if receiverSecret {
			return c.projectSignature(value)
		}
		return value
	}
	args, returns := []string{}, []string{}
	generics, genericDeclarations := []string{}, []string{}
	if truth(item["native"]) {
		if truth(item["method"]) {
			args = append(args, "self")
		}
		nativeArgs := asList(item["Arguments"])
		for index, v := range nativeArgs {
			argument := asMap(v)
			if truth(argument["StrideIndex"]) {
				tail := []string{}
				for _, v := range nativeArgs[index:] {
					tail = append(tail, c.nativeTypeValue(asMap(v), false))
				}
				args = append(args, strings.Join(uniqueStrings(tail), " | ")+"...")
				break
			}
			raw := asString(argument["Type"])
			if tuple, ok := c.abiTuples[raw]; ok {
				args = append(args, stringsOf(tuple)...)
				continue
			}
			acceptsSecret := asString(item["SecretArguments"]) == "AllowedWhenTainted" && !truth(argument["NeverSecret"])
			value := c.nativeTypeValue(argument, acceptsSecret && c.primitive(raw) != "")
			if acceptsSecret && containsString([]string{"TextureAsset", "TextureAssetDisk"}, raw) {
				value += " | W.OpaqueNumber"
			}
			if truth(argument["StrideIndex"]) {
				value += "..."
			} else if _, hasDefault := argument["Default"]; truth(argument["Nilable"]) || hasDefault {
				value = "? " + strings.TrimSuffix(value, " | nil")
			}
			args = append(args, value)
		}
		nativeReturns := asList(item["Returns"])
		absent := c.mayReturnNothing(item)
		if truth(item["ReturnsNeverSecret"]) {
			next := []any{}
			for _, v := range nativeReturns {
				next = append(next, withFields(asMap(v), M{"NeverSecret": true}))
			}
			nativeReturns = next
		}
		for index, v := range nativeReturns {
			output := asMap(v)
			if truth(output["StrideIndex"]) {
				tail := []string{}
				for _, v := range nativeReturns[index:] {
					value := asMap(v)
					tail = append(tail, c.nativeTypeValue(withFields(value, M{"Nilable": truth(value["Nilable"]) || absent}), receiverSecret || c.resultSecret(item)))
				}
				returns = append(returns, strings.Join(uniqueStrings(tail), " | ")+"...")
				break
			}
			raw := asString(output["Type"])
			if tuple, ok := c.abiTuples[raw]; ok {
				for _, value := range stringsOf(tuple) {
					if !truth(output["NeverSecret"]) && (receiverSecret || c.resultSecret(item) || secretResult(output)) {
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
			if truth(output["StrideIndex"]) {
				value += "..."
			}
			returns = append(returns, value)
		}
	} else {
		item = mapCopy(item)
		declarations := []string{}
		for _, group := range stringsOf(item["generics"]) {
			declarations = append(declarations, splitTop(group, ',')...)
		}
		for _, decl := range declarations {
			name, bound, separator := strings.Cut(decl, ":")
			name, bound = strings.TrimSpace(name), strings.TrimSpace(bound)
			if separator && bound != "table" && (containsString([]string{"string", "number", "boolean"}, bound) || strings.Contains(bound, "OptionsGetter")) {
				pattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(name) + `\b`)
				for _, values := range [][]any{asList(item["params"]), asList(item["returns"])} {
					for _, v := range values {
						value := asMap(v)
						value["type"] = pattern.ReplaceAllString(asString(value["type"]), bound)
					}
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
		if truth(item["method"]) {
			args = append(args, "self")
		}
		params := M{}
		for _, v := range asList(item["params"]) {
			param := asMap(v)
			params[asString(param["name"])] = param
		}
		for _, name := range stringsOf(item["arguments"]) {
			param := asMap(params[name])
			if param == nil && name == "..." {
				param = asMap(params["..."])
			}
			value := "any"
			if param == nil {
				c.notes = append(c.notes, M{"symbol": getOr(item, "qualified_name", nil), "rule": "source parameter has no type annotation", "parameter": name, "source": item["source"]})
			} else {
				value = c.tealValue(asString(param["type"]), owner, generics)
			}
			if name == "..." {
				value += "..."
			} else if param != nil && (truth(param["optional"]) || containsString(splitTop(value, '|'), "nil") || containsString(splitTop(asString(param["type"]), '|'), "nil")) {
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
		for _, v := range asList(item["returns"]) {
			output := asMap(v)
			raw := asString(output["type"])
			if raw == "..." {
				value := "any..."
				if receiverSecret {
					value = "OpaqueValue..."
				}
				returns = append(returns, value)
			} else {
				value := c.tealValue(raw, owner, generics)
				if receiverSecret && !truth(output["NeverSecret"]) {
					value = c.projectSecret(value)
				}
				if truth(output["variadic"]) {
					value += "..."
				}
				returns = append(returns, value)
			}
		}
		if len(asList(item["returns"])) == 0 && asString(item["source"]) == "questie:Data/Wiki.lua" {
			value := "any..."
			if receiverSecret {
				value = "OpaqueValue..."
			}
			returns = append(returns, value)
			c.notes = append(c.notes, M{"symbol": getOr(item, "qualified_name", nil), "rule": "wiki stub has no result annotation; result tuple remains dynamic", "source": item["source"]})
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
func (c *catalogue) signatures(item M, owner string) (values []string, err error) {
	defer captureError(&err)
	return c.signaturesValue(item, owner), nil
}
func (c *catalogue) signaturesValue(item M, owner string) []string {
	values := []string{}
	variants := append([]any{item}, asList(item["variants"])...)
	for _, v := range variants {
		variant := asMap(v)
		choices := [][]string{}
		count := 1
		for _, v := range asList(variant["params"]) {
			param := asMap(v)
			raw := asString(param["type"])
			members := splitTop(raw, '|')
			tables := []string{}
			for _, value := range members {
				if !containsString([]string{"string", "table", "math"}, value) && c.defs[value] != nil && containsString([]string{"record", "interface"}, asString(asMap(c.defs[value])["kind"])) {
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
			unsupported(asString(variant["qualified_name"]), "more than 32 input overload combinations")
		}
		var emit func(int, []string)
		emit = func(index int, combination []string) {
			if index < len(choices) {
				for _, choice := range choices[index] {
					emit(index+1, append(combination, choice))
				}
				return
			}
			expanded := mapCopy(variant)
			for i, v := range asList(expanded["params"]) {
				asMap(v)["type"] = combination[i]
			}
			values = append(values, c.signatureValue(expanded, owner))
		}
		emit(0, nil)
	}
	for _, raw := range stringsOf(item["overloads"]) {
		converted := c.tealValue(raw, owner, nil)
		signature := converted
		if def := asMap(c.defs[converted]); asString(def["kind"]) == "callback" {
			signature = asString(def["signature"])
		}
		if c.secretStructs[owner] && !truth(item["ReturnsNeverSecret"]) {
			signature = c.projectSignature(signature)
		}
		values = append(values, signature)
	}
	return uniqueStrings(values)
}
func (c *catalogue) prepareGlobals() {
	for _, name := range append([]string{}, c.globalOrder...) {
		item := asMap(c.globals[name])
		if item == nil || item["type_name"] != nil {
			continue
		}
		signatures := c.signaturesValue(item, "")
		if len(signatures) < 2 {
			continue
		}
		key := "Callable_" + ident(name)
		record := c.record(key, asString(item["source"]))
		calls := []string{}
		for _, signature := range signatures {
			call := strings.Replace(signature, "(", "(self, ", 1)
			call = strings.ReplaceAll(call, "self, )", "self)")
			calls = append(calls, call)
		}
		variants := []any{}
		for _, call := range calls[1:] {
			variants = append(variants, M{"resolved_signature": call})
		}
		asMap(record["methods"])["metamethod __call"] = M{"resolved_signature": calls[0], "variants": variants}
		c.setGlobal(name, M{"type_name": key, "source": item["source"], "callable_contract": item})
	}
}
func (c *catalogue) reconcileInheritance() {
	type expanded struct{ fields, methods M }
	cache := map[string]expanded{}
	var expand func(string, []string) expanded
	expand = func(name string, stack []string) expanded {
		if result, ok := cache[name]; ok {
			return result
		}
		if containsString(stack, name) {
			reject("cyclic inheritance: %s", strings.Join(append(stack, name), " -> "))
		}
		record := asMap(c.defs[name])
		fields, methods := M{}, M{}
		for _, base := range stringsOf(record["bases"]) {
			visited := map[string]bool{}
			for c.defs[base] == nil && c.aliases[base] != nil {
				if visited[base] {
					reject("cyclic base alias: %s", base)
				}
				visited[base] = true
				base = asString(getOr(asMap(c.aliases[base]), "expression", base))
			}
			if c.defs[base] == nil {
				reject("%s: unknown base %s", name, base)
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
		record["inherited_fields"] = inheritedFields
		for k, v := range asMap(record["fields"]) {
			fields[k] = v
		}
		for _, method := range sortedKeys(asMap(record["methods"])) {
			item := asMap(asMap(record["methods"])[method])
			if truth(asMap(methods[method])["native"]) && !truth(item["native"]) {
				c.notes = append(c.notes, M{"symbol": name + "." + method, "rule": "inherit canonical native ancestor contract"})
				continue
			}
			methods[method] = item
		}
		result := expanded{fields, methods}
		cache[name] = result
		return result
	}
	for _, name := range append([]string{}, c.defOrder...) {
		record := asMap(c.defs[name])
		if containsString([]string{"record", "interface"}, asString(record["kind"])) {
			expand(name, nil)
		}
	}
	for _, name := range sortedKeys(cache) {
		record := asMap(c.defs[name])
		methods := M{}
		for key, value := range asMap(record["methods"]) {
			candidate := asMap(cache[name].methods[key])
			actual := asMap(value)
			if sameMapIdentity(candidate, actual) {
				methods[key] = value
			}
		}
		record["methods"] = methods
		fields := M{}
		inherited, _ := record["inherited_fields"].(map[string]bool)
		for key, value := range asMap(record["fields"]) {
			if !inherited[key] || truth(asMap(value)["native"]) {
				fields[key] = value
			}
		}
		record["fields"] = fields
		if truth(record["native_object"]) {
			record["kind"], record["opaque"] = "interface", true
		}
	}
	for _, name := range c.defOrder {
		record := asMap(c.defs[name])
		for _, base := range stringsOf(record["bases"]) {
			value := asMap(c.defs[base])
			if asString(value["kind"]) == "record" && !truth(value["fields"]) && !truth(value["opaque"]) {
				value["kind"] = "interface"
			}
		}
	}
}
func sameMapIdentity(a, b M) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return mapPointer(a) == mapPointer(b)
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
				setAdd(out, targets(asString(asMap(c.aliases[token])["expression"]), next))
			} else if record := asMap(c.defs[token]); record != nil && !c.isHandle(token) && containsString([]string{"record", "interface", "converted_record"}, asString(record["kind"])) {
				out[token] = true
			}
		}
		return out
	}
	seed := func(item M, inherited bool) map[string]bool {
		if truth(item["NeverSecret"]) || truth(item["NeverSecretContents"]) {
			return map[string]bool{}
		}
		expression := ""
		if _, ok := item["Type"]; ok || truth(item["native"]) {
			expression = asString(getOr(item, "InnerType", getOr(item, "Type", "")))
		} else {
			expression = asString(getOr(item, "type", getOr(item, "resolved", "")))
		}
		if inherited || secretResult(item) {
			return targets(expression, map[string]bool{})
		}
		return map[string]bool{}
	}
	c.secretStructs = map[string]bool{}
	functions := []M{}
	for _, name := range c.globalOrder {
		functions = append(functions, asMap(c.globals[name]))
	}
	for _, name := range c.defOrder {
		record := asMap(c.defs[name])
		for _, method := range sortedKeys(asMap(record["methods"])) {
			functions = append(functions, asMap(asMap(record["methods"])[method]))
		}
		for _, field := range asMap(record["fields"]) {
			if item := asMap(field); item != nil {
				setAdd(c.secretStructs, seed(item, false))
			}
		}
	}
	for _, item := range functions {
		for _, output := range asList(item["Returns"]) {
			if !truth(item["ReturnsNeverSecret"]) {
				setAdd(c.secretStructs, seed(asMap(output), c.resultSecret(item)))
			}
		}
	}
	for {
		previous := map[string]bool{}
		setAdd(previous, c.secretStructs)
		for _, name := range append([]string{}, c.defOrder...) {
			record := asMap(c.defs[name])
			if record == nil || c.isHandle(name) {
				continue
			}
			bases := map[string]bool{}
			for _, base := range stringsOf(record["bases"]) {
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
			for _, v := range asMap(record["fields"]) {
				if field, ok := v.(string); ok {
					setAdd(c.secretStructs, targets(field, map[string]bool{}))
				} else {
					field := asMap(v)
					if !truth(field["NeverSecret"]) && !truth(field["NeverSecretContents"]) {
						if truth(field["native"]) {
							setAdd(c.secretStructs, seed(field, true))
						} else {
							expression := asString(field["resolved"])
							if _, ok := field["resolved"]; !ok {
								expression = c.tealValue(asString(field["type"]), name, nil)
							}
							setAdd(c.secretStructs, targets(expression, map[string]bool{}))
						}
					}
				}
			}
			for _, method := range sortedKeys(asMap(record["methods"])) {
				item := asMap(asMap(record["methods"])[method])
				if truth(item["ReturnsNeverSecret"]) {
					continue
				}
				for _, output := range asList(getOr(item, "Returns", getOr(item, "returns", []any{}))) {
					setAdd(c.secretStructs, seed(asMap(output), true))
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
