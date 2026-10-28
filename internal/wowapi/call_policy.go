package wowapi

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"wow-teal-types/internal/consumer"
)

// callPolicy is finalized with its signature. Source permission fields never
// enter the renderer or the consumer checker as a second interpretation path.
type callPolicy struct {
	Symbol           string
	SecretParameters []int
	SecretVarargFrom int
	Guard            string
}

func (p *callPolicy) equal(other *callPolicy) bool {
	if p == nil || other == nil {
		return p == other
	}
	return p.Symbol == other.Symbol && p.Guard == other.Guard && p.SecretVarargFrom == other.SecretVarargFrom && slices.Equal(p.SecretParameters, other.SecretParameters)
}

func (c *catalogue) projectCallable(item *catalogueFunction, owner, symbol string) []projectedSignature {
	out := []projectedSignature{}
	for _, choice := range c.distinctChoices(item, owner, symbol) {
		out = append(out, choice.signature)
	}
	return out
}

// distinctChoice is one overload in its projected identity, with the contract
// that produced it.
type distinctChoice struct {
	signature projectedSignature
	contract  *catalogueFunction
}

// distinctChoices owns overload identity: variants that differ only by nil are
// one overload, and must agree on permissions.
func (c *catalogue) distinctChoices(item *catalogueFunction, owner, symbol string) []distinctChoice {
	out := []distinctChoice{}
	seen := map[projectedType]int{}
	for _, choice := range c.signatureChoices(item, owner) {
		expression := projectType(choice.expression)
		policy := c.signaturePolicy(choice.contract, symbol, string(expression))
		if index, ok := seen[expression]; ok {
			if !out[index].signature.policy.equal(policy) {
				reject(fmt.Sprintf("ambiguous permissions for identical overload: %s", symbol))
			}
			continue
		}
		seen[expression] = len(out)
		out = append(out, distinctChoice{signature: projectedSignature{expression: expression, policy: policy}, contract: choice.contract})
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
	policy := &callPolicy{Symbol: symbol, SecretParameters: []int{}}
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
	slots := c.argumentSlots(contract.Native)
	arguments, ok := signatureParameters(signature)
	if !ok {
		reject(fmt.Sprintf("invalid finalized native signature: %s", symbol))
	}
	// The rendered parameters must be exactly the native slots, so every grant
	// lands on the parameter whose declared type the slot produced.
	want := 0
	if item.Method {
		want++
	}
	for _, slot := range slots {
		if slot.variadic {
			want++
		} else {
			want += slot.width
		}
	}
	if len(arguments) != want {
		reject(fmt.Sprintf("native signature %s renders %d parameters for %d native slots: %s", symbol, len(arguments), want, signature))
	}
	position := 1
	if item.Method {
		position++
	}
	for _, slot := range slots {
		if slot.variadic {
			if slot.acceptsSecret {
				policy.SecretVarargFrom = position
			}
			break
		}
		for range slot.width {
			if slot.acceptsSecret {
				policy.SecretParameters = append(policy.SecretParameters, position)
			}
			position++
		}
	}
	return policy
}

// signatureParameters splits a function signature's parameter list.
func signatureParameters(signature string) ([]string, bool) {
	opening := strings.IndexByte(signature, '(')
	if opening < 0 {
		return nil, false
	}
	closing := matchingClose(signature, opening)
	if closing < 0 {
		return nil, false
	}
	parameters := splitTop(signature[opening+1:closing], ',')
	if len(parameters) == 1 && parameters[0] == "" {
		return nil, true
	}
	return parameters, true
}

// argumentSlot is one native argument's secret permission. The declared
// parameter types and call_policy.lua both derive from it, so a permitted
// position is a position whose source-rendered type admits a Secret. Reviewed
// overloads may narrow a permitted position to a type that rejects a Secret;
// the grant then never applies.
type argumentSlot struct {
	width         int // declared parameters; an ABI tuple spans several
	acceptsSecret bool
	variadic      bool // the collapsed repeating tail; always the last slot
}

// argumentSlots returns one slot per native argument up to the repeating tail.
// No frozen source establishes recursive input permission. Output secrecy
// flags and the absence of a prohibition cannot grant that permission.
func (c *catalogue) argumentSlots(native *nativeFunction) []argumentSlot {
	tainted := stringValue(native.Policy.SecretArguments) == "AllowedWhenTainted"
	slots := []argumentSlot{}
	for index, argument := range native.Arguments {
		if argument.StrideIndex != nil && *argument.StrideIndex != 0 {
			// A repeating source tuple becomes one union vararg. Every member
			// must permit a secret root before that collapsed slot can permit it.
			accepts := tainted
			for _, member := range native.Arguments[index:] {
				accepts = accepts && !enabled(member.NeverSecret)
			}
			return append(slots, argumentSlot{width: 1, acceptsSecret: accepts, variadic: true})
		}
		width := 1
		if tuple, ok := c.abiTuples[argument.Type]; ok {
			width = len(tuple)
		}
		slots = append(slots, argumentSlot{width: width, acceptsSecret: tainted && !enabled(argument.NeverSecret)})
	}
	return slots
}

func renderCallPolicy(locations map[string]map[string]callPolicy) []byte {
	var out strings.Builder
	out.WriteString("-- Generated native-call permissions, bound by the consumer artifact manifest.\nreturn {\n   schema = 1,\n   compiler = " + luaQuote(consumer.TealVersion) + ",\n   declarations = {\n")
	for _, path := range slices.Sorted(maps.Keys(locations)) {
		fmt.Fprintf(&out, "      [%s] = {\n", luaQuote(path))
		for _, position := range slices.Sorted(maps.Keys(locations[path])) {
			policy := locations[path][position]
			fmt.Fprintf(&out, "         [%s] = { symbol = %s, secret_parameters = %s", luaQuote(position), luaQuote(policy.Symbol), luaPositions(policy.SecretParameters))
			if policy.SecretVarargFrom > 0 {
				fmt.Fprintf(&out, ", secret_vararg_from = %d", policy.SecretVarargFrom)
			}
			if policy.Guard != "" {
				fmt.Fprintf(&out, ", guard = %s", luaQuote(policy.Guard))
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
