package wowapi

import (
	"fmt"
	"maps"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"wow-teal-types/internal/wowapi/literal"
)

func evSortedUnique(values []string) []string {
	set := map[string]bool{}
	for _, value := range values {
		set[value] = true
	}
	return sortedKeys(set)
}
func evContains(values []string, value string) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}
func evEqual[T any](a, b T) bool {
	x, err := jsonCompact(a)
	if err != nil {
		return false
	}
	y, err := jsonCompact(b)
	return err == nil && string(x) == string(y)
}
func evDigest[T any](value T) string {
	hash, err := digest(value)
	if err != nil {
		factFail(fmt.Sprintf("invalid evidence JSON: %s", err))
	}
	return hash
}
func evTier(path string) int {
	for i, p := range []string{"vendor/", "generated/", "overrides/"} {
		if strings.HasPrefix(path, p) {
			return i
		}
	}
	factFail("unknown alias scan tier: " + path)
	return 0
}
func scopedPolicy(tag string, names []string) policyScope {
	words := strings.Fields(tag)
	if len(words) == 0 || !evContains([]string{"none", "untainted", "tainted"}, words[0]) {
		return policyScope{Raw: tag, Unsupported: "unknown secret argument policy"}
	}
	requested := words[1:]
	unknown := []string{}
	positions := []int{}
	for _, name := range requested {
		found := false
		for i, value := range names {
			if name == value {
				positions = append(positions, i+1)
				found = true
				break
			}
		}
		if !found {
			unknown = append(unknown, name)
		}
	}
	if len(unknown) > 0 {
		positions = []int{}
	} else if len(requested) == 0 {
		for i := range names {
			positions = append(positions, i+1)
		}
	}
	sort.Ints(positions)
	unique := []int{}
	for _, value := range positions {
		if len(unique) == 0 || value != unique[len(unique)-1] {
			unique = append(unique, value)
		}
	}
	scope := "named"
	if len(requested) == 0 {
		scope = "all"
	}
	return policyScope{Raw: tag, Policy: words[0], Scope: scope, Positions: unique, UnknownNames: evSortedUnique(unknown)}
}
func renderFact(node factNode) string {
	if node == nil {
		factFail("empty type node")
	}
	switch node.kind {
	case "name":
		return node.text
	case "literal":
		data, err := encodeJSON(node.text, false, false, false)
		if err != nil {
			factFail("invalid literal")
		}
		return string(data)
	case "variable", "pack", "class-name":
		value := "$" + strconv.Itoa(node.index)
		if node.kind == "pack" {
			return "..." + value
		}
		if node.kind == "class-name" {
			return "`" + value + "`"
		}
		return value
	case "union", "intersection":
		parts := []string{}
		for _, child := range node.children {
			part := renderFact(child)
			if child.kind == "callable" || child.kind == "union" {
				part = "(" + part + ")"
			}
			parts = append(parts, part)
		}
		separator, prefix := "|", ""
		if node.kind == "intersection" {
			separator = "&"
			if len(parts) == 1 {
				prefix = "&"
			}
		}
		return prefix + strings.Join(parts, separator)
	case "array", "variadic":
		child := node.children[0]
		value := renderFact(child)
		if evContains([]string{"union", "intersection", "callable"}, child.kind) {
			value = "(" + value + ")"
		}
		if node.kind == "array" {
			return value + "[]"
		}
		return "..." + value
	case "generic":
		parts := []string{}
		for _, child := range node.children {
			parts = append(parts, renderFact(child))
		}
		return node.text + "<" + strings.Join(parts, ",") + ">"
	case "callable":
		parameters := []string{}
		for i, p := range node.parameters {
			name := "arg" + strconv.Itoa(i)
			if p.variadic {
				name = "..."
			} else if p.optional {
				name += "?"
			}
			parameters = append(parameters, name+":"+renderFact(p.value))
		}
		returns := []string{}
		for _, value := range node.returns {
			returns = append(returns, renderFact(value))
		}
		result := "fun(" + strings.Join(parameters, ",") + ")"
		if len(returns) > 0 {
			result += ":" + strings.Join(returns, ",")
		}
		return result
	}
	factFail("unknown type node: " + node.kind)
	return ""
}
func unionFact(nodes []factNode) factNode {
	flat := []factNode{}
	for _, node := range nodes {
		if node.kind == "union" {
			flat = append(flat, node.children...)
		} else {
			flat = append(flat, node)
		}
	}
	unique := map[string]factNode{}
	for _, node := range flat {
		unique[renderFact(node)] = node
	}
	ordered := []factNode{}
	for _, key := range sortedKeys(unique) {
		ordered = append(ordered, unique[key])
	}
	if len(ordered) == 1 {
		return ordered[0]
	}
	return compoundFact("union", ordered...)
}
func nullableFact(node factNode) bool {
	if node.kind == "name" && node.text == "nil" {
		return true
	}
	if node.kind == "union" {
		for _, child := range node.children {
			if child.kind == "name" && child.text == "nil" {
				return true
			}
		}
	}
	return false
}
func carrierFact(node factNode) factNode {
	switch node.kind {
	case "generic":
		if node.text == "secret" && len(node.children) == 1 {
			return carrierFact(node.children[0])
		}
		children := []factNode{}
		for _, child := range node.children {
			children = append(children, carrierFact(child))
		}
		return genericFact(node.text, children...)
	case "union", "array", "variadic", "intersection":
		children := []factNode{}
		for _, child := range node.children {
			children = append(children, carrierFact(child))
		}
		if node.kind == "union" {
			return unionFact(children)
		}
		return compoundFact(node.kind, children...)
	case "callable":
		parameters := []factParameter{}
		for _, p := range node.parameters {
			parameters = append(parameters, factParameter{p.optional, p.variadic, carrierFact(p.value)})
		}
		returns := []factNode{}
		for _, value := range node.returns {
			returns = append(returns, carrierFact(value))
		}
		return callableFact(parameters, returns)
	}
	return node
}
func withoutNilFact(node factNode) factNode {
	if node.kind != "union" {
		return node
	}
	children := []factNode{}
	for _, child := range node.children {
		if child.kind != "name" || child.text != "nil" {
			children = append(children, child)
		}
	}
	if len(children) == 0 {
		return nameFact("nil")
	}
	return unionFact(children)
}

type typeFacts struct {
	catalogue *catalogue
	aliases   map[string][]*factAlias
	enums     map[string]factEvidence
	callbacks map[string]nativeCallbackEvidence
}

func newTypeFacts(files map[string]string, c *catalogue) *typeFacts {
	f := &typeFacts{catalogue: c, aliases: map[string][]*factAlias{}, enums: map[string]factEvidence{}, callbacks: map[string]nativeCallbackEvidence{}}
	if c != nil {
		for _, d := range c.nativeDocs {
			for _, item := range d.Metadata.Tables {
				if item.Type == "CallbackType" {
					f.callbacks[item.Name] = nativeCallbackEvidence{item.Callback(), "blizzard:" + d.Name}
				}
			}
		}
	}
	enumPattern := regexp.MustCompile(`(?m)---@enum\s+([\w.]+)\s*\n([\w.]+)\s*=\s*\{([^{}]*)\}`)
	aliasPattern := regexp.MustCompile(`^\s*---@alias\s+([\w.-]+)(?:\s+(.*))?$`)
	for _, path := range sortedKeys(files) {
		text := files[path]
		for _, ix := range enumPattern.FindAllStringSubmatchIndex(text, -1) {
			match := func(i int) string { return text[ix[i*2]:ix[i*2+1]] }
			fields := strings.Split(match(3), ",")
			valid, hasField := true, false
			for _, field := range fields {
				field = strings.TrimSpace(field)
				if field == "" {
					continue
				}
				hasField = true
				if !regexp.MustCompile(`^\w+\s*=\s*-?(?:0x[0-9a-fA-F]+|\d+(?:\.\d+)?)$`).MatchString(field) {
					valid = false
				}
			}
			if valid && hasField {
				f.enums[match(1)] = factEvidence{Name: match(1), Source: path, Line: new(strings.Count(text[:ix[0]], "\n") + 1), Representation: "numeric enum carrier"}
			}
		}
		var pending *factAlias
		for i, raw := range strings.Split(text, "\n") {
			m := aliasPattern.FindStringSubmatch(raw)
			if m != nil {
				pending = &factAlias{Name: m[1], Members: []string{}, Source: path, Line: i + 1}
				f.aliases[m[1]] = append(f.aliases[m[1]], pending)
				if m[2] != "" {
					aliasMember(pending, m[2])
				}
			} else if pending != nil && strings.HasPrefix(strings.TrimSpace(raw), "---|") {
				aliasMember(pending, strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(raw), "---|")))
			} else if strings.TrimSpace(raw) != "" && !strings.HasPrefix(strings.TrimSpace(raw), "--") {
				pending = nil
			}
		}
	}
	return f
}

type nativeCallbackEvidence struct {
	Function NativeFunction
	Source   string
}

func aliasMember(alias *factAlias, expression string) {
	value, _, err := splitSourceType(expression)
	if err != nil {
		alias.Unsupported = expression
	} else if value != "" {
		alias.Members = append(alias.Members, value)
	}
}
func (f *typeFacts) fact(expression, source string, generics []string) (result typeFact) {
	evidence := map[string]factEvidence{}
	evidenceRows := func() []factEvidence {
		rows := []factEvidence{}
		for _, key := range sortedKeys(evidence) {
			rows = append(rows, evidence[key])
		}
		return rows
	}
	defer func() {
		if caught := recover(); caught != nil {
			if failure, ok := caught.(factFailure); ok {
				result = typeFact{Raw: expression, Unsupported: failure.message, Evidence: evidenceRows()}
			} else {
				panic(caught)
			}
		}
	}()
	if !evContains([]string{"tsm", "native", "effective"}, source) {
		factFail("type fact source must be tsm, native or effective")
	}
	env := map[string]factBinding{}
	type binderDeclaration struct {
		pack bool
		raw  string
	}
	declarations := []binderDeclaration{}
	for _, declaration := range generics {
		values, err := SplitTopLevel(declaration, ',')
		if err != nil {
			factFail(err.Error())
		}
		for _, value := range values {
			match := regexp.MustCompile(`^\s*(\.\.\.[A-Za-z_]\w*|[A-Za-z_]\w*)\s*(?::\s*(.+))?$`).FindStringSubmatch(value)
			if match == nil {
				factFail("invalid generic constraint: " + value)
			}
			pack := strings.HasPrefix(match[1], "...")
			name := strings.TrimPrefix(match[1], "...")
			if _, exists := env[name]; exists {
				factFail("invalid or duplicate generic binder: " + name)
			}
			env[name] = factBinding{index: len(declarations), pack: pack}
			declarations = append(declarations, binderDeclaration{pack, match[2]})
		}
	}
	binders := []genericBinder{}
	for _, declaration := range declarations {
		var constraint factNode
		if declaration.raw != "" {
			constraint = f.parse(declaration.raw, source, nil, evidence, 0, env)
		}
		binders = append(binders, genericBinder{Constraint: constraint, Pack: declaration.pack})
	}
	node := f.parse(expression, source, nil, evidence, 0, env)
	wrappers := []secretWrapper{}
	features := map[string]bool{}
	var visit func(factNode, []factPathPart)
	visit = func(node factNode, path []factPathPart) {
		kind := node.kind
		if evContains([]string{"callable", "variable", "class-name", "pack", "intersection"}, kind) {
			feature := kind
			if feature == "variable" {
				feature = "bound-generic"
			}
			features[feature] = true
		}
		switch kind {
		case "generic":
			if node.text == "secret" {
				if len(node.children) != 1 {
					factFail("secret wrapper requires one type")
				}
				wrappers = append(wrappers, secretWrapper{path, renderFact(node.children[0])})
			}
			for i, child := range node.children {
				visit(child, factPath(path, "argument", &i))
			}
		case "array", "variadic":
			label := "element"
			if kind == "variadic" {
				label = "tail"
			}
			visit(node.children[0], factPath(path, label, nil))
		case "union", "intersection":
			for i, child := range node.children {
				visit(child, factPath(path, kind, &i))
			}
		case "callable":
			for i, p := range node.parameters {
				visit(p.value, factPath(path, "callback-argument", &i))
			}
			for i, child := range node.returns {
				visit(child, factPath(path, "callback-return", &i))
			}
		}
	}
	visit(node, []factPathPart{})
	result = typeFact{Raw: expression, Type: renderFact(node), Tree: node, CarrierType: renderFact(withoutNilFact(carrierFact(node))), Nilable: nullableFact(node), Variadic: node.kind == "variadic" || node.kind == "pack", Secret: len(wrappers) > 0, SecretWrappers: wrappers, Evidence: evidenceRows(), Features: sortedKeys(features)}
	if len(binders) > 0 {
		result.Binders = binders
	}
	return result
}

type factBinding struct {
	index int
	pack  bool
}

func (f *typeFacts) parse(expression, source string, seen []string, evidence map[string]factEvidence, depth int, env map[string]factBinding) factNode {
	if depth > 48 {
		factFail("type fact nesting exceeds 48 levels")
	}
	node, _, err := readSourceType(expression, false)
	if err != nil {
		factFail(err.Error())
	}
	return f.resolve(node, source, seen, evidence, depth, env)
}
func (f *typeFacts) resolve(node factNode, source string, seen []string, evidence map[string]factEvidence, depth int, env map[string]factBinding) factNode {
	if depth > 48 {
		factFail("type fact nesting exceeds 48 levels")
	}
	walk := func(value factNode) factNode { return f.resolve(value, source, seen, evidence, depth+1, env) }
	switch node.kind {
	case "union", "array", "intersection":
		children := []factNode{}
		for _, child := range node.children {
			children = append(children, walk(child))
		}
		if node.kind == "union" {
			return unionFact(children)
		}
		return compoundFact(node.kind, children...)
	case "class-name":
		binding, exists := env[node.text]
		if !exists || binding.pack {
			factFail("unbound class-name generic: " + node.text)
		}
		return indexedFact(node.kind, binding.index)
	case "variadic":
		child := node.children[0]
		if child.kind == "name" {
			if binding, exists := env[child.text]; exists && binding.pack {
				return indexedFact("pack", binding.index)
			}
		}
		return compoundFact(node.kind, walk(child))
	case "generic":
		if node.text == "secret" && len(node.children) != 1 {
			factFail("secret wrapper requires one type")
		}
		children := []factNode{}
		for _, child := range node.children {
			children = append(children, walk(child))
		}
		return genericFact(node.text, children...)
	case "callable":
		parameters := []factParameter{}
		for _, p := range node.parameters {
			parameters = append(parameters, factParameter{p.optional, p.variadic, walk(p.value)})
		}
		returns := []factNode{}
		for _, value := range node.returns {
			returns = append(returns, walk(value))
		}
		return callableFact(parameters, returns)
	}
	if node.kind != "name" {
		return node
	}
	text := node.text
	if mapped, exists := map[string]string{"integer": "number", "cstring": "string", "bool": "boolean", "luaIndex": "number"}[text]; exists {
		text = mapped
	}
	if binding, exists := env[text]; exists {
		if binding.pack {
			factFail("generic pack requires expansion: " + text)
		}
		return indexedFact("variable", binding.index)
	}
	if evContains(seen, text) {
		factFail("cyclic alias: " + strings.Join(append(append([]string{}, seen...), text), " -> "))
	}
	nextSeen := append(append([]string{}, seen...), text)
	if source == "native" {
		if entry, exists := f.callbacks[text]; exists {
			evidence["native-callback:"+text] = factEvidence{Name: text, Source: entry.Source, Metadata: entry.Function.Raw}
			return f.nativeCallback(entry.Function, source, nextSeen, evidence, depth+1, env)
		}
	}
	if (source == "native" || source == "effective") && f.catalogue != nil && f.catalogue.defs[text] == nil {
		alias := f.catalogue.aliases[text]
		if alias != nil && alias.Signature != nil {
			native := alias.Signature.Native
			if native == nil {
				factFail("native callback alias lacks native contract: " + text)
			}
			// Persisted callback evidence includes its interpretation provenance;
			// keep those annotations outside the immutable native source record.
			metadata := maps.Clone(native.Raw)
			if metadata == nil {
				metadata = literal.Object{}
			}
			metadata["native"] = literal.Bool(true)
			metadata["source"] = literal.Text(alias.Source)
			evidence["native:"+text] = factEvidence{Name: text, Source: alias.Source, Metadata: metadata}
			return f.nativeCallback(*native, source, nextSeen, evidence, depth+1, env)
		}
		if alias != nil && alias.Expression != "" {
			evidence["native:"+text] = factEvidence{Name: text, Source: alias.Source, Expression: &alias.Expression}
			return f.parse(alias.Expression, source, nextSeen, evidence, depth+1, env)
		}
	}
	if source == "tsm" && len(f.aliases[text]) > 0 {
		aliases := f.aliases[text]
		tier := -1
		for _, alias := range aliases {
			tier = max(tier, evTier(alias.Source))
		}
		perFile := map[string]*factAlias{}
		for _, alias := range aliases {
			if evTier(alias.Source) == tier && (perFile[alias.Source] == nil || perFile[alias.Source].Line < alias.Line) {
				perFile[alias.Source] = alias
			}
		}
		expressions := map[string]bool{}
		unsupported := false
		for _, alias := range perFile {
			if alias.Unsupported != "" {
				unsupported = true
			} else {
				expressions[strings.Join(alias.Members, "|")] = true
			}
		}
		for _, alias := range aliases {
			row := factEvidence{Name: alias.Name, Source: alias.Source, Line: &alias.Line, Unsupported: alias.Unsupported}
			if len(aliases) > 1 {
				row.Expression = new(strings.Join(alias.Members, "|"))
				row.Selected = new(perFile[alias.Source] == alias)
				row.SelectionBasis = "last-wins by known scan tier and within-file order; cross-file order unavailable"
			}
			evidence[alias.Source+":"+strconv.Itoa(alias.Line)] = row
		}
		if len(expressions) != 1 || unsupported {
			factFail("conflicting same-tier or unsupported alias: " + text)
		}
		return f.parse(sortedKeys(expressions)[0], source, nextSeen, evidence, depth+1, env)
	}
	if source == "tsm" {
		if item, exists := f.enums[text]; exists {
			evidence[item.Source+":"+strconv.Itoa(*item.Line)] = item
			return nameFact("number")
		}
	}
	return nameFact(text)
}
func (f *typeFacts) nativeCallback(item NativeFunction, source string, seen []string, evidence map[string]factEvidence, depth int, env map[string]factBinding) factNode {
	valueType := func(value NativeParameter, returned bool) factNode {
		expression := value.Type
		if value.Mixin != "" {
			expression = value.Mixin
		}
		if value.Type == "table" && value.InnerType != "" {
			expression = value.InnerType + "[]"
		}
		node := f.parse(expression, source, seen, evidence, depth+1, env)
		secret := enabled(value.ConditionalSecret) || enabled(value.SecretValue) || (returned && enabled(item.Policy.SecretReturns))
		if secret && !enabled(value.NeverSecret) && !(returned && enabled(item.Policy.ReturnsNeverSecret)) {
			node = genericFact("secret", node)
		}
		if returned && enabled(value.Nilable) {
			node = unionFact([]factNode{node, nameFact("nil")})
		}
		if returned && value.StrideIndex != nil && *value.StrideIndex != 0 {
			node = compoundFact("variadic", node)
		}
		return node
	}
	parameters := []factParameter{}
	for _, value := range item.Arguments {
		parameters = append(parameters, factParameter{enabled(value.Nilable) || value.HasDefault, value.StrideIndex != nil && *value.StrideIndex != 0, valueType(value, false)})
	}
	returns := []factNode{}
	for _, value := range item.Returns {
		returns = append(returns, valueType(value, true))
	}
	return callableFact(parameters, returns)
}
func (f *typeFacts) equivalent(a, b typeFact) bool {
	return a.Unsupported == "" && b.Unsupported == "" && evEqual(a.Tree, b.Tree) && evEqual(a.Binders, b.Binders)
}
