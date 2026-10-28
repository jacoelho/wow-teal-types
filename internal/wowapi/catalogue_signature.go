package wowapi

import (
	"cmp"
	"context"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"
)

// projectSecret marks every member of expression as possibly secret. Plain
// values merge into one MaybeSecret so the union holds a single Secret
// userdata; userdata handles can only be Secret, and possibly secret tables
// keep their contents marked. Secret structures keep their name because their
// own fields carry the marking. Already projected members are kept, so the
// projection is idempotent.
func (c *catalogue) projectSecret(expression string, position typePosition) string {
	plain, kept := []string{}, []string{}
	nilable := false
	for _, member := range splitTop(expression, '|') {
		primitive := c.primitive(member)
		defName, _ := c.defNamed(member)
		name := cmp.Or(defName, member)
		switch {
		case member == "nil":
			nilable = true
		case member == "any" || member == secretDataType || strings.HasPrefix(member, "W.MaybeSecret<") || strings.HasPrefix(member, "W.Secret<"):
			kept = append(kept, member)
		case member == dataType || primitive == "any" || strings.HasPrefix(member, "function"):
			plain = append(plain, dataType)
		case primitive != "" || kindOf(c.defs[defName]) == kindEnum:
			plain = append(plain, member)
		case strings.HasPrefix(member, "{"):
			if position == inputPosition {
				kept = append(kept, "any")
			} else {
				kept = append(kept, secretTableType)
			}
		case c.isUserdata(name):
			kept = append(kept, "W.Secret<"+member+">")
		default:
			kept = append(kept, member)
		}
	}
	if slices.Contains(kept, "any") {
		return "any"
	}
	projected := []string{}
	if plain = uniqueStrings(plain); slices.Contains(plain, dataType) {
		// SecretData already covers every plain member and secret tables.
		projected = append(projected, secretDataType)
		kept = slices.DeleteFunc(kept, func(member string) bool { return member == secretTableType || member == secretDataType })
	} else if len(plain) > 0 && position == inputPosition {
		// Secret<T> is invariant, so Secret<A | B> rejects Secret<A>; an input
		// accepts each member's secret separately.
		for _, member := range plain {
			projected = append(projected, "W.MaybeSecret<"+member+">")
		}
	} else if len(plain) > 0 {
		projected = append(projected, "W.MaybeSecret<"+strings.Join(plain, " | ")+">")
	}
	projected = append(projected, uniqueStrings(kept)...)
	if nilable {
		projected = append(projected, "nil")
	}
	return strings.Join(projected, " | ")
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
		value := c.projectSecret(result, outputPosition)
		if variadic {
			value += "..."
		}
		results = append(results, value)
	}
	return parts[0] + ": " + strings.Join(results, ", ")
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
		taintedSecrets := stringValue(native.Policy.SecretArguments) == "AllowedWhenTainted"
		for index, argument := range nativeArgs {
			if argument.StrideIndex != nil && *argument.StrideIndex != 0 {
				// The collapsed vararg accepts secrets exactly when signaturePolicy
				// grants secret_vararg_from: every member must permit them.
				acceptsSecret := taintedSecrets
				tail := []string{}
				for _, v := range nativeArgs[index:] {
					acceptsSecret = acceptsSecret && !enabled(v.NeverSecret)
					tail = append(tail, c.nativeTypeValue(v, false, inputPosition))
				}
				value := strings.Join(uniqueStrings(tail), " | ")
				if acceptsSecret {
					value = c.projectSecret(value, inputPosition)
				}
				args = append(args, value+"...")
				break
			}
			raw := argument.Type
			acceptsSecret := taintedSecrets && !enabled(argument.NeverSecret)
			if tuple, ok := c.abiTuples[raw]; ok {
				for _, value := range tuple {
					if acceptsSecret {
						value = c.projectSecret(value, inputPosition)
					}
					args = append(args, value)
				}
				continue
			}
			value := c.nativeTypeValue(argument, acceptsSecret && c.primitive(raw) != "", inputPosition)
			if acceptsSecret && slices.Contains([]string{"TextureAsset", "TextureAssetDisk"}, raw) {
				value += " | W.Secret<number>"
			}
			if enabled(argument.Nilable) || argument.HasDefault {
				value = "? " + strings.TrimSuffix(value, " | nil")
			}
			args = append(args, value)
		}
		nativeReturns := append([]nativeParameter{}, native.Returns...)
		absent := c.mayReturnNothing(item)
		if item.returnsNeverSecret() {
			for i := range nativeReturns {
				nativeReturns[i].NeverSecret = new(true)
			}
		}
		for index, output := range nativeReturns {
			if output.StrideIndex != nil && *output.StrideIndex != 0 {
				tail := []string{}
				for _, value := range nativeReturns[index:] {
					value.Nilable = new(enabled(value.Nilable) || absent)
					tail = append(tail, c.nativeTypeValue(value, receiverSecret || c.resultSecret(item), outputPosition))
				}
				returns = append(returns, strings.Join(uniqueStrings(tail), " | ")+"...")
				break
			}
			raw := output.Type
			if tuple, ok := c.abiTuples[raw]; ok {
				for _, value := range tuple {
					if !enabled(output.NeverSecret) && (receiverSecret || c.resultSecret(item) || output.secretResult()) {
						value = c.projectSecret(value, outputPosition)
					}
					if absent {
						value += " | nil"
					}
					returns = append(returns, value)
				}
				continue
			}
			value := c.nativeTypeValue(output, receiverSecret || c.resultSecret(item), outputPosition)
			if absent && !slices.Contains(splitTop(value, '|'), "nil") {
				value += " | nil"
			}
			returns = append(returns, value)
		}
	} else {
		annotation := item.clone().Annotation
		if annotation == nil {
			annotation = &annotationFunction{}
		}
		declarations := []string{}
		for _, group := range annotation.Generics {
			declarations = append(declarations, splitTop(group, ',')...)
		}
		for _, decl := range declarations {
			name, bound, separator := strings.Cut(decl, ":")
			name, bound = strings.TrimSpace(name), strings.TrimSpace(bound)
			if separator && bound != "table" && (slices.Contains([]string{"string", "number", "boolean"}, bound) || strings.Contains(bound, "OptionsGetter")) {
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
				replacement := "any"
				if separator {
					replacement = bound
				}
				for i := range annotation.Returns {
					annotation.Returns[i].Type = pattern.ReplaceAllString(annotation.Returns[i].Type, replacement)
				}
				continue
			}
			generics = append(generics, name)
			// A table bound stays unconstrained: Teal records do not satisfy a
			// map constraint.
			constraint := ""
			if separator && bound != "table" {
				constraint = c.resolve(bound, inputPosition)
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
		params := map[string]annotationParameter{}
		for _, param := range annotation.Params {
			params[param.Name] = param
		}
		for _, name := range annotation.Arguments {
			param, present := params[name]
			value := inputPosition.untyped()
			if present {
				value = c.tealValue(param.Type, owner, generics, inputPosition)
			}
			if name == "..." {
				value += "..."
			} else if present && (param.Optional || slices.Contains(splitTop(value, '|'), "nil") || slices.Contains(splitTop(param.Type, '|'), "nil")) {
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
				value := outputPosition.untyped()
				if receiverSecret {
					value = c.projectSecret(value, outputPosition)
				}
				returns = append(returns, value+"...")
			} else {
				value := c.tealValue(raw, owner, generics, outputPosition)
				if receiverSecret {
					value = c.projectSecret(value, outputPosition)
				}
				if output.Variadic {
					value += "..."
				}
				returns = append(returns, value)
			}
		}
		if len(annotation.Returns) == 0 && item.Source == "questie:Data/Wiki.lua" {
			value := outputPosition.untyped()
			if receiverSecret {
				value = c.projectSecret(value, outputPosition)
			}
			returns = append(returns, value+"...")
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
		var params []annotationParameter
		if variant.Annotation != nil {
			params = variant.Annotation.Params
		}
		for _, param := range params {
			raw := param.Type
			members := splitTop(raw, '|')
			tables := []string{}
			for _, value := range members {
				if !slices.Contains([]string{"string", "table", "math"}, value) && kindOf(c.defs[value]).isTableLike() {
					tables = append(tables, value)
				}
			}
			options := []string{raw}
			if len(tables) > 1 && !slices.Contains(members, "table") {
				options = []string{}
				for _, table := range tables {
					ordinary := []string{}
					for _, member := range members {
						if !slices.Contains(tables, member) {
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
		// An overload is a function the game provides: its parameters are inputs.
		converted := c.tealValue(raw, owner, nil, outputPosition)
		signature := converted
		if def := c.defs[converted]; kindOf(def) == kindCallback {
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
		// projectCallable owns overload identity: variants that differ only by
		// nil are one overload, and must agree on permissions.
		if len(c.projectCallable(item, "", name)) < 2 {
			continue
		}
		choices := c.signatureChoices(item, "")
		key := "Callable_" + ident(name)
		record := c.record(key, item.Source, kindRecord)
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
		fields          map[string]*catalogueField
		methods         map[string]*catalogueFunction
		inheritedFields map[string]bool
	}
	cache := map[string]expanded{}
	var expand func(string, []string) expanded
	expand = func(name string, stack []string) expanded {
		if result, ok := cache[name]; ok {
			return result
		}
		if slices.Contains(stack, name) {
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
		for k, v := range record.Fields {
			fields[k] = v
		}
		for _, method := range slices.Sorted(maps.Keys(record.Methods)) {
			item := record.Methods[method]
			if previous := methods[method]; previous != nil && previous.Native != nil && item.Native == nil {
				continue
			}
			methods[method] = item
		}
		result := expanded{fields, methods, inheritedFields}
		cache[name] = result
		return result
	}
	for _, name := range append([]string{}, c.defOrder...) {
		// Converted records are excluded: their fields are already Teal
		// expressions and their only bases are member-less TableVariants
		// interfaces, so there is nothing to inherit.
		if kind := kindOf(c.defs[name]); kind == kindRecord || kind == kindInterface {
			expand(name, nil)
		}
	}
	for _, name := range slices.Sorted(maps.Keys(cache)) {
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
			if !cache[name].inheritedFields[key] || value.Native != nil {
				fields[key] = value
			}
		}
		record.Fields = fields
		if record.NativeObject {
			record.Kind, record.Opaque = kindInterface, true
		}
	}
	for _, name := range c.defOrder {
		record := c.defs[name]
		for _, base := range definitionBases(record) {
			value := c.defs[base]
			if kindOf(value) == kindRecord && len(value.Fields) == 0 && !value.Opaque {
				value.Kind = kindInterface
			}
		}
	}
}

// secrecyPassLimit bounds the secrecy fixed point. Every pass but the last
// marks at least one more definition, so the limit only trips on a defect.
const secrecyPassLimit = 32768

func (c *catalogue) prepareSecrecy(ctx context.Context) {
	var targets func(string, map[string]bool) map[string]bool
	targets = func(expression string, seen map[string]bool) map[string]bool {
		out := map[string]bool{}
		if seen[expression] || strings.HasPrefix(expression, "fun(") || strings.HasPrefix(expression, "function(") {
			return out
		}
		next := map[string]bool{}
		maps.Copy(next, seen)
		next[expression] = true
		for _, token := range dottedIdentifiers.FindAllString(expression, -1) {
			if c.primitive(token) != "" || slices.Contains([]string{"nil", "table", "self"}, token) {
				continue
			}
			if c.aliases[token] != nil && c.defs[token] == nil {
				maps.Copy(out, targets(c.aliases[token].Expression, next))
			} else if record := c.defs[token]; record != nil && !c.isHandle(token) && record.Kind.isTableLike() {
				out[token] = true
			}
		}
		return out
	}
	seed := func(item nativeParameter, inherited bool) map[string]bool {
		if enabled(item.NeverSecret) || enabled(item.NeverSecretContents) {
			return map[string]bool{}
		}
		if inherited || item.secretResult() {
			return targets(cmp.Or(item.InnerType, item.Type), map[string]bool{})
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
		for _, method := range slices.Sorted(maps.Keys(record.Methods)) {
			functions = append(functions, record.Methods[method])
		}
		for _, field := range record.Fields {
			if field.Native != nil {
				maps.Copy(c.secretStructs, seed(*field.Native, false))
			}
		}
	}
	for _, item := range functions {
		if item.Native != nil && !item.returnsNeverSecret() {
			for _, output := range item.Native.Returns {
				maps.Copy(c.secretStructs, seed(output, c.resultSecret(item)))
			}
		}
	}
	for pass := 0; ; pass++ {
		if err := ctx.Err(); err != nil {
			rejectErr(err)
		}
		if pass == secrecyPassLimit {
			reject(fmt.Sprintf("secrecy propagation did not converge within %d passes", secrecyPassLimit))
		}
		previous := maps.Clone(c.secretStructs)
		for _, name := range append([]string{}, c.defOrder...) {
			record := c.defs[name]
			if record == nil || c.isHandle(name) {
				continue
			}
			bases := map[string]bool{}
			for _, base := range record.Bases {
				maps.Copy(bases, targets(base, map[string]bool{}))
			}
			for base := range bases {
				if c.secretStructs[base] {
					c.secretStructs[name] = true
				}
			}
			if !c.secretStructs[name] {
				continue
			}
			maps.Copy(c.secretStructs, bases)
			for _, field := range record.Fields {
				if field.Native != nil {
					maps.Copy(c.secretStructs, seed(*field.Native, true))
				} else {
					expression := field.Type
					if field.Resolved != nil {
						expression = *field.Resolved
					} else if record.Kind != kindConvertedRecord {
						expression = c.tealValue(field.Type, name, nil, outputPosition)
					}
					maps.Copy(c.secretStructs, targets(expression, map[string]bool{}))
				}
			}
			for _, method := range slices.Sorted(maps.Keys(record.Methods)) {
				item := record.Methods[method]
				if item.returnsNeverSecret() {
					continue
				}
				if item.Native != nil {
					for _, output := range item.Native.Returns {
						maps.Copy(c.secretStructs, seed(output, true))
					}
				} else if item.Annotation != nil {
					for _, output := range item.Annotation.Returns {
						maps.Copy(c.secretStructs, targets(output.Type, map[string]bool{}))
					}
				}
				for _, signature := range c.signaturesValue(item, name) {
					parts := splitTop(signature, ':')
					if len(parts) > 1 {
						maps.Copy(c.secretStructs, targets(strings.Join(parts[1:], ":"), map[string]bool{}))
					}
				}
			}
		}
		if maps.Equal(previous, c.secretStructs) {
			break
		}
	}
}
