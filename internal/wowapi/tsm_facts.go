package wowapi

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func evStrings(v any) []string {
	r := []string{}
	switch a := v.(type) {
	case []string:
		return append(r, a...)
	case []any:
		for _, x := range a {
			r = append(r, asString(x))
		}
	}
	return r
}
func evList(a []string) []any {
	r := make([]any, len(a))
	for i, v := range a {
		r[i] = v
	}
	return r
}
func evEqual(a, b any) bool {
	x, e := jsonCompact(a)
	if e != nil {
		return false
	}
	y, e := jsonCompact(b)
	return e == nil && string(x) == string(y)
}
func evSortedUnique(a []string) []string {
	m := map[string]bool{}
	for _, s := range a {
		m[s] = true
	}
	return sortedKeys(m)
}
func evContains(a []string, s string) bool {
	for _, v := range a {
		if v == s {
			return true
		}
	}
	return false
}
func evGet(m M, k string, defaultValue any) any {
	if v, ok := m[k]; ok {
		return v
	}
	return defaultValue
}
func evMapWithout(m M, exclude ...string) M {
	r := M{}
	for k, v := range m {
		if !evContains(exclude, k) {
			r[k] = v
		}
	}
	return r
}
func evMerge(a M, others ...M) M {
	r := M{}
	for k, v := range a {
		r[k] = v
	}
	for _, b := range others {
		for k, v := range b {
			r[k] = v
		}
	}
	return r
}
func evDigest(v any) string {
	s, e := digest(v)
	if e != nil {
		factFail("invalid evidence JSON: %s", e)
	}
	return s
}
func evTier(path string) int {
	for i, p := range []string{"vendor/", "generated/", "overrides/"} {
		if strings.HasPrefix(path, p) {
			return i
		}
	}
	factFail("unknown alias scan tier: %s", path)
	return 0
}
func scopedPolicy(tag string, parameters []any) M {
	words := strings.Fields(tag)
	if len(words) == 0 || !evContains([]string{"none", "untainted", "tainted"}, words[0]) {
		return M{"raw": tag, "unsupported": "unknown secret argument policy"}
	}
	names := []string{}
	for _, v := range parameters {
		if s, ok := v.(string); ok {
			names = append(names, s)
		} else {
			names = append(names, asString(asMap(v)["name"]))
		}
	}
	requested := words[1:]
	unknown := []string{}
	positions := []any{}
	for _, name := range requested {
		found := false
		for i, v := range names {
			if name == v {
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
		positions = []any{}
	} else if len(requested) == 0 {
		for i := range names {
			positions = append(positions, i+1)
		}
	}
	sort.Slice(positions, func(i, j int) bool { return asInt(positions[i]) < asInt(positions[j]) })
	uniq := []any{}
	for _, v := range positions {
		if len(uniq) == 0 || v != uniq[len(uniq)-1] {
			uniq = append(uniq, v)
		}
	}
	scope := "named"
	if len(requested) == 0 {
		scope = "all"
	}
	return M{"raw": tag, "policy": words[0], "scope": scope, "positions": uniq, "unknown_names": evList(evSortedUnique(unknown))}
}
func renderFact(n factNode) string {
	if len(n) == 0 {
		factFail("empty type node")
	}
	parts := n[1:]
	switch kindOf(n) {
	case "name":
		return asString(parts[0])
	case "literal":
		b, e := encodeJSON(parts[0], false, false, false)
		if e != nil {
			factFail("invalid literal")
		}
		return string(b)
	case "variable", "pack", "class-name":
		v := "$" + strconv.Itoa(asInt(parts[0]))
		if kindOf(n) == "pack" {
			return "..." + v
		}
		if kindOf(n) == "class-name" {
			return "`" + v + "`"
		}
		return v
	case "union", "intersection":
		a := []string{}
		for _, x := range parts {
			p := childNode(x)
			s := renderFact(p)
			if kindOf(p) == "callable" || kindOf(p) == "union" {
				s = "(" + s + ")"
			}
			a = append(a, s)
		}
		op := "|"
		prefix := ""
		if kindOf(n) == "intersection" {
			op = "&"
			if len(parts) == 1 {
				prefix = "&"
			}
		}
		return prefix + strings.Join(a, op)
	case "array", "variadic":
		p := childNode(parts[0])
		s := renderFact(p)
		if evContains([]string{"union", "intersection", "callable"}, kindOf(p)) {
			s = "(" + s + ")"
		}
		if kindOf(n) == "array" {
			return s + "[]"
		}
		return "..." + s
	case "generic":
		a := []string{}
		for _, v := range parts[1:] {
			a = append(a, renderFact(childNode(v)))
		}
		return asString(parts[0]) + "<" + strings.Join(a, ",") + ">"
	case "callable":
		params := []string{}
		for i, x := range asList(parts[0]) {
			p := childNode(x)
			name := "arg" + strconv.Itoa(i)
			if truth(p[2]) {
				name = "..."
			} else if truth(p[1]) {
				name += "?"
			}
			params = append(params, name+":"+renderFact(childNode(p[3])))
		}
		returns := []string{}
		for _, v := range asList(parts[1]) {
			returns = append(returns, renderFact(childNode(v)))
		}
		s := "fun(" + strings.Join(params, ",") + ")"
		if len(returns) > 0 {
			s += ":" + strings.Join(returns, ",")
		}
		return s
	}
	factFail("unknown type node: %s", kindOf(n))
	return ""
}
func unionFact(nodes []any) factNode {
	flat := []any{}
	for _, x := range nodes {
		n := childNode(x)
		if kindOf(n) == "union" {
			flat = append(flat, n[1:]...)
		} else {
			flat = append(flat, n)
		}
	}
	unique := map[string]any{}
	for _, x := range flat {
		unique[renderFact(childNode(x))] = x
	}
	ordered := []any{}
	for _, k := range sortedKeys(unique) {
		ordered = append(ordered, unique[k])
	}
	if len(ordered) == 1 {
		return childNode(ordered[0])
	}
	return fn("union", ordered...)
}
func nullableFact(n factNode) bool {
	if evEqual(n, fn("name", "nil")) {
		return true
	}
	if kindOf(n) == "union" {
		for _, x := range n[1:] {
			if evEqual(x, fn("name", "nil")) {
				return true
			}
		}
	}
	return false
}
func carrierFact(n factNode) factNode {
	kind := kindOf(n)
	switch kind {
	case "generic":
		if asString(n[1]) == "secret" && len(n) == 3 {
			return carrierFact(childNode(n[2]))
		}
		a := []any{n[1]}
		for _, v := range n[2:] {
			a = append(a, carrierFact(childNode(v)))
		}
		return fn(kind, a...)
	case "union":
		a := []any{}
		for _, v := range n[1:] {
			a = append(a, carrierFact(childNode(v)))
		}
		return unionFact(a)
	case "array", "variadic", "intersection":
		a := []any{}
		for _, v := range n[1:] {
			a = append(a, carrierFact(childNode(v)))
		}
		return fn(kind, a...)
	case "callable":
		params := []any{}
		for _, v := range asList(n[1]) {
			p := childNode(v)
			params = append(params, fn("parameter", p[1], p[2], carrierFact(childNode(p[3]))))
		}
		returns := []any{}
		for _, v := range asList(n[2]) {
			returns = append(returns, carrierFact(childNode(v)))
		}
		return fn(kind, params, returns)
	}
	return n
}
func withoutNilFact(n factNode) factNode {
	if kindOf(n) != "union" {
		return n
	}
	a := []any{}
	for _, v := range n[1:] {
		if !evEqual(v, fn("name", "nil")) {
			a = append(a, v)
		}
	}
	if len(a) == 0 {
		return fn("name", "nil")
	}
	return unionFact(a)
}

type typeFacts struct {
	catalogue *catalogue
	aliases   map[string][]M
	enums     M
	callbacks map[string]M
}

func newTypeFacts(files map[string]string, c *catalogue) *typeFacts {
	f := &typeFacts{catalogue: c, aliases: map[string][]M{}, enums: M{}, callbacks: map[string]M{}}
	if c != nil {
		for _, d := range c.nativeDocs {
			for _, v := range asList(d.Metadata["Tables"]) {
				item := asMap(v)
				if asString(item["Type"]) == "CallbackType" {
					f.callbacks[asString(item["Name"])] = M{"item": item, "source": "blizzard:" + d.Name}
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
			valid, any := true, false
			for _, field := range fields {
				field = strings.TrimSpace(field)
				if field == "" {
					continue
				}
				any = true
				if !regexp.MustCompile(`^\w+\s*=\s*-?(?:0x[0-9a-fA-F]+|\d+(?:\.\d+)?)$`).MatchString(field) {
					valid = false
				}
			}
			if valid && any {
				f.enums[match(1)] = M{"name": match(1), "source": path, "line": strings.Count(text[:ix[0]], "\n") + 1, "representation": "numeric enum carrier"}
			}
		}
		var pending M
		for i, raw := range strings.Split(text, "\n") {
			m := aliasPattern.FindStringSubmatch(raw)
			if m != nil {
				pending = M{"name": m[1], "members": []any{}, "source": path, "line": i + 1}
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
func aliasMember(pending M, expression string) {
	v, _, e := splitSourceType(expression)
	if e != nil {
		pending["unsupported"] = expression
	} else if v != "" {
		pending["members"] = append(asList(pending["members"]), v)
	}
}
func (f *typeFacts) fact(expression, source string, generics []string) (result M) {
	evidence := M{}
	defer func() {
		if p := recover(); p != nil {
			if e, ok := p.(factFailure); ok {
				rows := []any{}
				for _, k := range sortedKeys(evidence) {
					rows = append(rows, evidence[k])
				}
				result = M{"raw": expression, "type": nil, "tree": nil, "unsupported": e.message, "evidence": rows}
			} else {
				panic(p)
			}
		}
	}()
	if !evContains([]string{"tsm", "native", "effective"}, source) {
		factFail("type fact source must be tsm, native or effective")
	}
	env := map[string]factBinding{}
	rawBinders := []M{}
	for _, declaration := range generics {
		values, e := SplitTopLevel(declaration, ',')
		if e != nil {
			factFail("%s", e)
		}
		for _, v := range values {
			m := regexp.MustCompile(`^\s*(\.\.\.[A-Za-z_]\w*|[A-Za-z_]\w*)\s*(?::\s*(.+))?$`).FindStringSubmatch(v)
			if m == nil {
				factFail("invalid generic constraint: %s", v)
			}
			pack := strings.HasPrefix(m[1], "...")
			name := strings.TrimPrefix(m[1], "...")
			if _, ok := env[name]; ok {
				factFail("invalid or duplicate generic binder: %s", name)
			}
			env[name] = factBinding{index: len(rawBinders), pack: pack}
			rawBinders = append(rawBinders, M{"pack": pack, "raw": m[2]})
		}
	}
	binders := []any{}
	for _, b := range rawBinders {
		var constraint any
		if asString(b["raw"]) != "" {
			constraint = f.parse(asString(b["raw"]), source, nil, evidence, 0, env)
		}
		binders = append(binders, M{"pack": b["pack"], "constraint": constraint})
	}
	n := f.parse(expression, source, nil, evidence, 0, env)
	wrappers := []any{}
	features := map[string]bool{}
	var visit func(factNode, []any)
	visit = func(n factNode, path []any) {
		kind := kindOf(n)
		if evContains([]string{"callable", "variable", "class-name", "pack", "intersection"}, kind) {
			feature := kind
			if feature == "variable" {
				feature = "bound-generic"
			}
			features[feature] = true
		}
		switch kind {
		case "generic":
			if asString(n[1]) == "secret" {
				if len(n) != 3 {
					factFail("secret wrapper requires one type")
				}
				wrappers = append(wrappers, M{"path": path, "type": renderFact(childNode(n[2]))})
			}
			for i, v := range n[2:] {
				visit(childNode(v), append(append([]any{}, path...), "argument", i))
			}
		case "array", "variadic":
			label := "element"
			if kind == "variadic" {
				label = "tail"
			}
			visit(childNode(n[1]), append(append([]any{}, path...), label))
		case "union", "intersection":
			for i, v := range n[1:] {
				visit(childNode(v), append(append([]any{}, path...), kind, i))
			}
		case "callable":
			for i, v := range asList(n[1]) {
				visit(childNode(childNode(v)[3]), append(append([]any{}, path...), "callback-argument", i))
			}
			for i, v := range asList(n[2]) {
				visit(childNode(v), append(append([]any{}, path...), "callback-return", i))
			}
		}
	}
	visit(n, []any{})
	rows := []any{}
	for _, k := range sortedKeys(evidence) {
		rows = append(rows, evidence[k])
	}
	result = M{"raw": expression, "type": renderFact(n), "tree": n, "carrier_type": renderFact(withoutNilFact(carrierFact(n))), "nilable": nullableFact(n), "variadic": kindOf(n) == "variadic" || kindOf(n) == "pack", "secret": len(wrappers) > 0, "secret_wrappers": wrappers, "evidence": rows, "features": evList(sortedKeys(features))}
	if len(binders) > 0 {
		result["binders"] = binders
	}
	return result
}

type factBinding struct {
	index int
	pack  bool
}

func (f *typeFacts) parse(expression, source string, seen []string, evidence M, depth int, env map[string]factBinding) factNode {
	if depth > 48 {
		factFail("type fact nesting exceeds 48 levels")
	}
	n, _, e := readSourceType(expression, false)
	if e != nil {
		factFail("%s", e)
	}
	return f.resolve(n, source, seen, evidence, depth, env)
}
func (f *typeFacts) resolve(n factNode, source string, seen []string, evidence M, depth int, env map[string]factBinding) factNode {
	if depth > 48 {
		factFail("type fact nesting exceeds 48 levels")
	}
	walk := func(v any) factNode { return f.resolve(childNode(v), source, seen, evidence, depth+1, env) }
	kind := kindOf(n)
	switch kind {
	case "union":
		a := []any{}
		for _, v := range n[1:] {
			a = append(a, walk(v))
		}
		return unionFact(a)
	case "class-name":
		b, ok := env[asString(n[1])]
		if !ok || b.pack {
			factFail("unbound class-name generic: %s", n[1])
		}
		return fn(kind, b.index)
	case "variadic":
		child := childNode(n[1])
		if kindOf(child) == "name" {
			if b, ok := env[asString(child[1])]; ok && b.pack {
				return fn("pack", b.index)
			}
		}
		return fn(kind, walk(n[1]))
	case "array", "intersection":
		a := []any{}
		for _, v := range n[1:] {
			a = append(a, walk(v))
		}
		return fn(kind, a...)
	case "generic":
		if asString(n[1]) == "secret" && len(n) != 3 {
			factFail("secret wrapper requires one type")
		}
		a := []any{n[1]}
		for _, v := range n[2:] {
			a = append(a, walk(v))
		}
		return fn(kind, a...)
	case "callable":
		params := []any{}
		for _, v := range asList(n[1]) {
			p := childNode(v)
			params = append(params, fn("parameter", p[1], p[2], walk(p[3])))
		}
		returns := []any{}
		for _, v := range asList(n[2]) {
			returns = append(returns, walk(v))
		}
		return fn(kind, params, returns)
	}
	if kind != "name" {
		return n
	}
	text := asString(n[1])
	if mapped, ok := map[string]string{"integer": "number", "cstring": "string", "bool": "boolean", "luaIndex": "number"}[text]; ok {
		text = mapped
	}
	if b, ok := env[text]; ok {
		if b.pack {
			factFail("generic pack requires expansion: %s", text)
		}
		return fn("variable", b.index)
	}
	if evContains(seen, text) {
		factFail("cyclic alias: %s", strings.Join(append(append([]string{}, seen...), text), " -> "))
	}
	if source == "native" {
		if entry := f.callbacks[text]; entry != nil {
			item := asMap(entry["item"])
			evidence["native-callback:"+text] = M{"name": text, "source": entry["source"], "metadata": item}
			return f.nativeCallback(item, source, append(append([]string{}, seen...), text), evidence, depth+1, env)
		}
	}
	if (source == "native" || source == "effective") && f.catalogue != nil && f.catalogue.defs[text] == nil {
		alias := asMap(f.catalogue.aliases[text])
		if sig := asMap(alias["signature"]); sig != nil {
			evidence["native:"+text] = M{"name": text, "source": alias["source"], "metadata": sig}
			return f.nativeCallback(sig, source, append(append([]string{}, seen...), text), evidence, depth+1, env)
		}
		if s, ok := alias["expression"].(string); ok {
			evidence["native:"+text] = M{"name": text, "source": alias["source"], "expression": s}
			return f.parse(s, source, append(append([]string{}, seen...), text), evidence, depth+1, env)
		}
	}
	if source == "tsm" && len(f.aliases[text]) > 0 {
		aliases := f.aliases[text]
		tier := -1
		for _, a := range aliases {
			tier = max(tier, evTier(asString(a["source"])))
		}
		perFile := map[string]M{}
		for _, a := range aliases {
			path := asString(a["source"])
			if evTier(path) == tier && (perFile[path] == nil || asInt(perFile[path]["line"]) < asInt(a["line"])) {
				perFile[path] = a
			}
		}
		expressions := map[string]bool{}
		unsupported := false
		for _, a := range perFile {
			if truth(a["unsupported"]) {
				unsupported = true
			} else {
				expressions[strings.Join(evStrings(a["members"]), "|")] = true
			}
		}
		for _, a := range aliases {
			fact := evMapWithout(a, "members")
			if len(aliases) > 1 {
				selected := evEqual(perFile[asString(a["source"])], a)
				fact["expression"] = strings.Join(evStrings(a["members"]), "|")
				fact["selected"] = selected
				fact["selection_basis"] = "last-wins by known scan tier and within-file order; cross-file order unavailable"
			}
			evidence[asString(a["source"])+":"+strconv.Itoa(asInt(a["line"]))] = fact
		}
		if len(expressions) != 1 || unsupported {
			factFail("conflicting same-tier or unsupported alias: %s", text)
		}
		return f.parse(sortedKeys(expressions)[0], source, append(append([]string{}, seen...), text), evidence, depth+1, env)
	}
	if source == "tsm" && f.enums[text] != nil {
		item := asMap(f.enums[text])
		evidence[asString(item["source"])+":"+strconv.Itoa(asInt(item["line"]))] = item
		return fn("name", "number")
	}
	return fn("name", text)
}
func (f *typeFacts) nativeCallback(item M, source string, seen []string, evidence M, depth int, env map[string]factBinding) factNode {
	valueType := func(value M, returned bool) factNode {
		expression := asString(evGet(value, "Mixin", value["Type"]))
		if asString(value["Type"]) == "table" && truth(value["InnerType"]) {
			expression = asString(value["InnerType"]) + "[]"
		}
		n := f.parse(expression, source, seen, evidence, depth+1, env)
		secret := truth(value["ConditionalSecret"]) || truth(value["SecretValue"]) || (returned && truth(item["SecretReturns"]))
		if secret && !truth(value["NeverSecret"]) && !(returned && truth(item["ReturnsNeverSecret"])) {
			n = fn("generic", "secret", n)
		}
		if returned && truth(value["Nilable"]) {
			n = unionFact([]any{n, fn("name", "nil")})
		}
		if returned && truth(value["StrideIndex"]) {
			n = fn("variadic", n)
		}
		return n
	}
	params := []any{}
	for _, v := range asList(item["Arguments"]) {
		m := asMap(v)
		_, def := m["Default"]
		params = append(params, fn("parameter", truth(m["Nilable"]) || def, truth(m["StrideIndex"]), valueType(m, false)))
	}
	returns := []any{}
	for _, v := range asList(item["Returns"]) {
		returns = append(returns, valueType(asMap(v), true))
	}
	return fn("callable", params, returns)
}
func (f *typeFacts) equivalent(a, b M) bool {
	return a["unsupported"] == nil && b["unsupported"] == nil && evEqual(a["tree"], b["tree"]) && evEqual(evGet(a, "binders", []any{}), evGet(b, "binders", []any{}))
}
func evError(p any) error {
	if e, ok := p.(factFailure); ok {
		return fmt.Errorf("%s", e.message)
	}
	panic(p)
}
