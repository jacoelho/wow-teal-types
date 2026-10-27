package wowapi

import (
	"fmt"
	"strconv"
	"strings"
)

// callPolicy is finalized with its signature. Source permission fields never
// enter the renderer or the consumer checker as a second interpretation path.
type callPolicy struct {
	Symbol                   string
	SecretParameters         []int
	SecretVarargFrom         int
	SecretContentsParameters []int
	SecretContentsVarargFrom int
	Guard                    string
}

func uniqueSignatureChoices(values []signatureChoice) []signatureChoice {
	out := []signatureChoice{}
	seen := map[string]bool{}
	for _, value := range values {
		if !seen[value.expression] {
			seen[value.expression] = true
			out = append(out, value)
		}
	}
	return out
}
func (c *catalogue) projectCallable(item *catalogueFunction, owner, symbol string) []projectedSignature {
	out := []projectedSignature{}
	seen := map[string]int{}
	for _, choice := range c.signatureChoices(item, owner) {
		policy := c.signaturePolicy(choice.contract, symbol, choice.expression)
		if index, ok := seen[choice.expression]; ok {
			if !evEqual(out[index].policy, policy) {
				reject(fmt.Sprintf("ambiguous permissions for identical overload: %s", symbol))
			}
			continue
		}
		seen[choice.expression] = len(out)
		out = append(out, projectedSignature{expression: choice.expression, policy: policy})
	}
	return out
}
func (c *catalogue) signaturePolicy(item *catalogueFunction, symbol, signature string) *callPolicy {
	if item.PolicySymbol != "" {
		symbol = item.PolicySymbol
	}
	contract := item.contract()
	if contract.Native == nil {
		return nil
	}
	policy := &callPolicy{Symbol: symbol, SecretParameters: []int{}, SecretContentsParameters: []int{}}
	switch symbol {
	case "canaccessvalue":
		policy.Guard = "value"
	case "canaccessallvalues":
		policy.Guard = "values"
	case "canaccesstable":
		policy.Guard = "table"
	case "issecretvalue":
		policy.Guard = "secret"
	}
	if stringValue(contract.Native.Policy.SecretArguments) != "AllowedWhenTainted" {
		return policy
	}
	opening := strings.IndexByte(signature, '(')
	if opening < 0 {
		return policy
	}
	closing := matchingClose(signature, opening)
	if closing < 0 {
		reject(fmt.Sprintf("invalid finalized native signature: %s", symbol))
	}
	arguments := splitTop(signature[opening+1:closing], ',')
	if len(arguments) == 1 && arguments[0] == "" {
		arguments = nil
	}
	position := 1
	if item.Method {
		position++
	}
	for index, parameter := range contract.Native.Arguments {
		root := !enabled(parameter.NeverSecret)
		// No frozen source establishes recursive input permission. Output secrecy
		// flags and the absence of a prohibition cannot grant that permission.
		width := 1
		if tuple, ok := c.abiTuples[parameter.Type]; ok {
			width = len(tuple)
		}
		variadic := parameter.StrideIndex != nil && *parameter.StrideIndex != 0
		if variadic {
			// A repeating source tuple becomes one union vararg. Every member
			// must permit a secret root before that collapsed slot can permit it.
			for _, member := range contract.Native.Arguments[index:] {
				root = root && !enabled(member.NeverSecret)
			}
			if root {
				policy.SecretVarargFrom = position
			}
			break
		}
		for range width {
			if position <= len(arguments) {
				if root {
					policy.SecretParameters = append(policy.SecretParameters, position)
				}
			}
			position++
		}
	}
	return policy
}

func (p declarationProjection) renderCallPolicy(files map[string][]byte) []byte {
	locations := map[string]map[string]callPolicy{}
	add := func(path string, line, column int, policy *callPolicy) {
		if policy == nil {
			return
		}
		if locations[path] == nil {
			locations[path] = map[string]callPolicy{}
		}
		key := strconv.Itoa(line) + ":" + strconv.Itoa(column)
		if previous, ok := locations[path][key]; ok && !evEqual(previous, *policy) {
			reject(fmt.Sprintf("conflicting native policy origin: %s:%s", path, key))
		}
		locations[path][key] = *policy
	}
	type origin struct {
		policy *callPolicy
		column int
	}
	methods := map[string]map[string]origin{}
	for _, declaration := range p.declarations {
		rows := map[string]origin{}
		for _, method := range declaration.methods {
			prefix := "      " + method.name + ": "
			for _, signature := range method.signatures {
				if signature.policy != nil {
					rows[prefix+signature.expression] = origin{signature.policy, len(prefix) + 1}
				}
			}
		}
		methods[declaration.name] = rows
	}
	globals := map[string]origin{}
	ambient := map[string]origin{}
	for _, global := range p.globals {
		if global.policy == nil {
			continue
		}
		prefix := "   type Global_" + ident(global.name) + " = "
		globals[prefix+global.expression] = origin{global.policy, len(prefix) + 1}
		if global.public {
			prefix = "global " + global.name + ": "
			ambient[prefix+global.ambientExpression] = origin{global.policy, len(prefix) + 1}
		}
	}
	for _, file := range []string{"wow.d.tl", "wow_env.d.tl", "wow_secure_env.d.tl"} {
		path := "types/wow/" + file
		locations[path] = map[string]callPolicy{}
		owner := ""
		for index, line := range strings.Split(string(files[file]), "\n") {
			if file == "wow.d.tl" {
				if strings.HasPrefix(line, "   record ") {
					owner = strings.TrimPrefix(line, "   record ")
				} else if strings.HasPrefix(line, "   interface ") {
					owner = strings.TrimPrefix(line, "   interface ")
				} else if line == "   end" {
					owner = ""
				}
				if value, ok := methods[owner][line]; ok {
					add(path, index+1, value.column, value.policy)
				}
				if value, ok := globals[line]; ok {
					add(path, index+1, value.column, value.policy)
				}
			} else if value, ok := ambient[line]; ok {
				add(path, index+1, value.column, value.policy)
			}
		}
	}
	var out strings.Builder
	out.WriteString("-- Generated native-call permissions, bound by the consumer artifact manifest.\nreturn {\n   schema = 1,\n   compiler = \"0.24.8-wow.1\",\n   declarations = {\n")
	for _, path := range sortedKeys(locations) {
		fmt.Fprintf(&out, "      [%s] = {\n", requireStringJSON(path))
		for _, position := range sortedKeys(locations[path]) {
			policy := locations[path][position]
			fmt.Fprintf(&out, "         [%s] = { symbol = %s, secret_parameters = %s, secret_contents_parameters = %s", requireStringJSON(position), requireStringJSON(policy.Symbol), luaPositions(policy.SecretParameters), luaPositions(policy.SecretContentsParameters))
			if policy.SecretVarargFrom > 0 {
				fmt.Fprintf(&out, ", secret_vararg_from = %d", policy.SecretVarargFrom)
			}
			if policy.SecretContentsVarargFrom > 0 {
				fmt.Fprintf(&out, ", secret_contents_vararg_from = %d", policy.SecretContentsVarargFrom)
			}
			if policy.Guard != "" {
				fmt.Fprintf(&out, ", guard = %s", requireStringJSON(policy.Guard))
			}
			out.WriteString(" },\n")
		}
		out.WriteString("      },\n")
	}
	out.WriteString("   },\n}\n")
	return []byte(out.String())
}
func luaPositions(values []int) string {
	parts := make([]string, len(values))
	for i, value := range values {
		parts[i] = strconv.Itoa(value)
	}
	return "{" + strings.Join(parts, ", ") + "}"
}
