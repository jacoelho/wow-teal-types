package wowapi

import (
	"context"
	"fmt"
	"strings"
)

type declarationKind uint8

const (
	declarationExternal declarationKind = iota + 1
	declarationEnum
	declarationAlias
	declarationCallback
	declarationRecord
	declarationInterface
)

type projectedSignature struct {
	expression string
	policy     *callPolicy
}

type projectedField struct {
	name, expression string
}
type projectedMethod struct {
	name, restrictions string
	signatures         []projectedSignature
}
type projectedDeclaration struct {
	kind             declarationKind
	name, source     string
	expression       string
	generics, values []string
	bases            []string
	fields           []projectedField
	methods          []projectedMethod
}
type projectedGlobal struct {
	name, expression, ambientExpression string
	policy                              *callPolicy
	public                              bool
}

// declarationProjection owns the normalized values consumed by rendering and
// call-policy generation. Source metadata never crosses into the renderer.
type declarationProjection struct {
	declarations  []projectedDeclaration
	globals       []projectedGlobal
	secureGlobals []string
	contracts     map[string][]projectedSignature
}

func (p declarationProjection) signatures(symbol string) ([]string, error) {
	values, ok := p.contracts[symbol]
	if !ok {
		return nil, fmt.Errorf("native symbol missing from finalized projection: %s", symbol)
	}
	out := make([]string, len(values))
	for i, value := range values {
		out[i] = value.expression
	}
	return out, nil
}

func emittedSignatures(c *catalogue, item *catalogueFunction, owner string) ([]string, error) {
	if item.TypeName == "" {
		return c.signatures(item, owner)
	}
	typeName := item.TypeName
	record := c.defs[typeName]
	var call *catalogueFunction
	if record != nil {
		call = record.Methods["metamethod __call"]
	}
	if call == nil {
		return nil, fmt.Errorf("native function binding is not callable: %s", typeName)
	}
	signatures, e := c.signatures(call, typeName)
	if e != nil {
		return nil, e
	}
	for i, s := range signatures {
		s = strings.Replace(s, "(self, ", "(", 1)
		s = strings.Replace(s, "(self)", "()", 1)
		signatures[i] = s
	}
	return signatures, nil
}
func projectSignatures(values []string) []projectedSignature {
	out := make([]projectedSignature, len(values))
	for i, value := range values {
		out[i] = projectedSignature{expression: value}
	}
	return out
}

// finalize resolves the generated-type and secrecy closure before publishing any
// normalized declaration. Repeated passes are bounded and compare semantic input
// state; rendering is never used as a discovery pass.
func (c *catalogue) finalize(ctx context.Context) (projection declarationProjection, err error) {
	defer captureError(&err)
	c.reconcileInheritance()
	c.prepareGlobals()
	for pass := 0; pass < 32; pass++ {
		if err = ctx.Err(); err != nil {
			return declarationProjection{}, err
		}
		before := c.projectionState()
		projection = c.projectDeclarations()
		c.prepareSecrecy()
		if len(c.errors) > 0 {
			return declarationProjection{}, fmt.Errorf("%d conversion errors: %s", len(c.errors), strings.Join(uniqueStrings(c.errors), "\n"))
		}
		if before == c.projectionState() {
			return projection, nil
		}
	}
	return declarationProjection{}, fmt.Errorf("declaration projection did not converge within 32 passes")
}

func (c *catalogue) projectionState() string {
	return jsonDigest(struct {
		Definitions   map[string]*catalogueDefinition
		Aliases       map[string]*catalogueAlias
		Globals       map[string]*catalogueFunction
		SecretStructs map[string]bool
	}{c.defs, c.aliases, c.globals, c.secretStructs})
}

func (c *catalogue) projectAttempt(label string, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(conversionError); ok {
				c.errors = append(c.errors, label+": "+e.Error())
			} else {
				panic(r)
			}
		}
	}()
	fn()
}

func (c *catalogue) projectDeclarations() declarationProjection {
	projection := declarationProjection{contracts: map[string][]projectedSignature{}, secureGlobals: sortedSet(c.secureGlobals)}
	for _, name := range append([]string{}, c.globalOrder...) {
		item := c.globals[name]
		if item.TypeName == "" {
			c.projectAttempt("global "+name, func() { c.signatureValue(item, "") })
		}
	}
	seen := map[string]bool{}
	for {
		pending := map[string]bool{}
		for name := range c.defs {
			if !seen[name] {
				pending[name] = true
			}
		}
		for name := range c.aliases {
			if !seen[name] {
				pending[name] = true
			}
		}
		if len(pending) == 0 {
			break
		}
		if len(seen)+len(pending) > 32768 {
			reject("declaration projection exceeds 32768 types")
		}
		for _, name := range c.orderDeclarations(pending) {
			seen[name] = true
			if name == "OpaqueValue" {
				continue
			}
			c.projectAttempt(name, func() {
				declaration, emit := c.projectDeclaration(name)
				if !emit {
					return
				}
				projection.declarations = append(projection.declarations, declaration)
				for _, method := range declaration.methods {
					projection.contracts[name+"."+method.name] = method.signatures
				}
			})
		}
	}
	exports := map[string]bool{"OpaqueNumber": true, "OpaqueText": true, "NumberLike": true, "TextLike": true}
	for name := range c.defs {
		if name != "unknown" && name != "any" {
			exports[ident(name)] = true
		}
	}
	for name := range c.aliases {
		if name != "unknown" && name != "any" {
			exports[ident(name)] = true
		}
	}
	for _, name := range sortedKeys(c.globals) {
		c.projectAttempt("global "+name, func() {
			if strings.Contains(name, ".") || !plainIdentifier.MatchString(name) {
				reject(fmt.Sprintf("unrepresentable global %s", name))
			}
			item := c.globals[name]
			value := item.TypeName
			if value != "" {
				value = ident(value)
			} else {
				value = c.signatureValue(item, "")
			}
			global := projectedGlobal{name: name, expression: value, public: !containsString([]string{"math", "string", "table"}, name) && !c.secureGlobals[name] && !strings.HasPrefix(item.Source, "questie:Libraries/")}
			direct := value
			if def := c.defs[value]; def != nil && def.Signature != "" {
				direct = def.Signature
			}
			if strings.HasPrefix(direct, "function") {
				global.ambientExpression = qualifyExports(direct, exports)
			} else {
				global.ambientExpression = "W.Global_" + ident(name)
			}
			if item.TypeName == "" {
				global.policy = c.signaturePolicy(item, name, value)
			}
			projection.globals = append(projection.globals, global)
			if item.TypeName == "" || item.CallableContract != nil {
				signatures, e := emittedSignatures(c, item, "")
				if e != nil {
					reject(fmt.Sprintf("%s", e))
				}
				projection.contracts[name] = projectSignatures(signatures)
			}
		})
	}
	return projection
}

func (c *catalogue) orderDeclarations(pending map[string]bool) []string {
	ordered := []string{}
	visited, visiting := map[string]bool{}, map[string]bool{}
	var visit func(string, int)
	visit = func(name string, depth int) {
		if visiting[name] || depth > 256 {
			reject(fmt.Sprintf("cyclic or unbounded declaration inheritance: %s", name))
		}
		if visited[name] {
			return
		}
		visiting[name] = true
		for _, base := range definitionBases(c.defs[name]) {
			aliases := map[string]bool{}
			for c.defs[base] == nil && c.aliases[base] != nil {
				if aliases[base] {
					reject(fmt.Sprintf("cyclic base alias: %s", base))
				}
				aliases[base] = true
				base = c.aliases[base].Expression
			}
			if pending[base] {
				visit(base, depth+1)
			}
		}
		ordered = append(ordered, name)
		visited[name] = true
		delete(visiting, name)
	}
	for _, name := range sortedKeys(pending) {
		visit(name, 0)
	}
	return ordered
}

func (c *catalogue) projectDeclaration(name string) (projectedDeclaration, bool) {
	declaration := projectedDeclaration{name: ident(name)}
	item := c.defs[name]
	if item == nil {
		alias := c.aliases[name]
		if alias.Expression == name {
			return declaration, false
		}
		declaration.kind, declaration.source = declarationAlias, alias.Source
		if value := alias.Signature; value != nil {
			declaration.expression = c.signatureValue(value, "")
		} else {
			declaration.expression = c.tealValue(alias.Expression, "", nil)
		}
		return declaration, true
	}
	declaration.source = item.Source
	kind := item.Kind
	switch kind {
	case "external":
		declaration.kind = declarationExternal
	case "enum":
		declaration.kind = declarationEnum
		for _, value := range item.Values {
			declaration.values = append(declaration.values, requireStringJSON(value))
		}
	case "callback":
		declaration.kind = declarationCallback
		declaration.name = name
		declaration.generics = item.Generics
		declaration.expression = item.Signature
	case "record", "interface", "converted_record":
		declaration.kind = declarationRecord
		if kind == "interface" {
			declaration.kind = declarationInterface
		}
		if item.Opaque {
			declaration.bases = append(declaration.bases, "userdata")
		}
		for _, base := range item.Bases {
			declaration.bases = append(declaration.bases, c.resolve(base))
		}
		fields, methods := item.Fields, item.Methods
		for _, nameOfField := range sortedKeys(fields) {
			value := fields[nameOfField]
			field := value
			if methods[nameOfField] != nil && field.Native == nil {
				continue
			}
			expression := ""
			switch {
			case kind == "converted_record":
				expression = field.Type
			case field.Resolved != nil:
				expression = *field.Resolved
			case field.Native != nil:
				expression = c.nativeTypeValue(*field.Native, c.secretStructs[name])
			default:
				expression = c.tealValue(field.Type, name, nil)
			}
			if c.secretStructs[name] && !(field.Native != nil && (enabled(field.Native.NeverSecret) || enabled(field.Native.NeverSecretContents))) {
				expression = c.projectSecret(expression)
			}
			if !plainIdentifier.MatchString(nameOfField) || tealKeywords[nameOfField] {
				nameOfField = "[" + requireStringJSON(strings.Trim(nameOfField, "[]\"'")) + "]"
			}
			declaration.fields = append(declaration.fields, projectedField{name: nameOfField, expression: expression})
		}
		for _, nameOfMethod := range sortedKeys(methods) {
			method := methods[nameOfMethod]
			projected := projectedMethod{name: nameOfMethod, signatures: c.projectCallable(method, name, name+"."+nameOfMethod)}
			if method.Native != nil {
				restricted := requireStringJSON(method.Native.Policy.restrictions())
				if restricted != "{}" {
					projected.restrictions = restricted
				}
			}
			declaration.methods = append(declaration.methods, projected)
		}
	default:
		reject(fmt.Sprintf("unsupported declaration kind %q for %s", kind, name))
	}
	return declaration, true
}
