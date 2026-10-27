package wowapi

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

func evMustFact(m M, e error) M {
	if e != nil {
		factFail("%s", e)
	}
	return m
}
func compareFunction(native, stub M, predicates M, widget bool, facts *typeFacts, absent *bool) (differences []any, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = evError(p)
		}
	}()
	differences = []any{}
	difference := func(kind string, expected, actual any, position int) {
		m := M{"kind": kind, "native": expected, "tsm": actual}
		if position > 0 {
			m["position"] = position
		}
		differences = append(differences, m)
	}
	expand := func(raw any) []any {
		r := []any{}
		for _, v := range asList(raw) {
			m := asMap(v)
			if m["Type"] == "uiRect" {
				for range 4 {
					r = append(r, evMerge(m, M{"Type": "number", "Mixin": "number"}))
				}
			} else {
				r = append(r, m)
			}
		}
		return r
	}
	arguments, returns := expand(native["Arguments"]), expand(native["Returns"])
	generics := evStrings(stub["generics"])
	carrier := func(value, side string) string {
		if facts == nil {
			return value
		}
		g := []string{}
		if side == "tsm" {
			g = generics
		}
		result := facts.fact(value, side, g)
		if s, ok := result["carrier_type"].(string); ok {
			return s
		}
		return value
	}
	positionExpression := func(expression string, variadic, nilable bool) string {
		if facts != nil && variadic && !truth(facts.fact(expression, "tsm", generics)["variadic"]) {
			if nilable {
				expression = "(" + expression + ")|nil"
			}
			return "(" + expression + ")..."
		}
		return expression
	}
	positionNilable := func(expression string, fallback bool) bool {
		if facts == nil {
			return fallback
		}
		fact := facts.fact(expression, "tsm", generics)
		n := childNode(fact["tree"])
		if kindOf(n) == "variadic" {
			return nullableFact(childNode(n[1]))
		}
		if v, ok := fact["nilable"].(bool); ok {
			return v
		}
		return fallback
	}
	structural := func(side string, position int, nativeExpression, tsmExpression string) {
		if facts == nil {
			return
		}
		expected := facts.fact(nativeExpression, "native", nil)
		actual := facts.fact(tsmExpression, "tsm", generics)
		if expected["unsupported"] != nil || actual["unsupported"] != nil {
			return
		}
		features := evSortedUnique(append(evStrings(expected["features"]), evStrings(actual["features"])...))
		if len(features) == 0 {
			return
		}
		if !evEqual(expected["tree"], actual["tree"]) || !evEqual(evGet(expected, "binders", []any{}), evGet(actual, "binders", []any{})) {
			differences = append(differences, M{"kind": side + "-structural-contract", "position": position, "native": expected, "tsm": actual, "features": evList(features)})
		}
	}
	stubArgs := asList(stub["arguments"])
	stubReturns := asList(stub["returns"])
	if len(arguments) != len(stubArgs) {
		difference("argument-count", len(arguments), len(stubArgs), 0)
	}
	annotated := M{}
	for _, v := range asList(stub["parameters"]) {
		m := asMap(v)
		annotated[asString(m["name"])] = m
	}
	for i, v := range arguments {
		if i >= len(stubArgs) {
			break
		}
		value := asMap(v)
		name := asString(stubArgs[i])
		actual := asMap(annotated[name])
		if actual == nil {
			difference("argument-annotation-missing", value["Type"], name, i+1)
			continue
		}
		expectedFact := evMustFact(nativeTypeFact(value, false))
		actualFact := evMustFact(simpleTypeFact(asString(actual["type"])))
		expectedExpression := positionExpression(asString(expectedFact["type"]), truth(value["StrideIndex"]), truth(value["Nilable"]))
		actualExpression := positionExpression(asString(actual["type"]), strings.HasPrefix(name, "..."), false)
		structural("argument", i+1, expectedExpression, actualExpression)
		actualCarrierExpression := actualExpression
		if facts == nil {
			actualCarrierExpression = asString(actualFact["type"])
		}
		if carrier(expectedExpression, "native") != carrier(actualCarrierExpression, "tsm") {
			difference("argument-type", expectedFact["type"], actualFact["type"], i+1)
		}
		_, hasDefault := value["Default"]
		optional := truth(value["Nilable"]) || hasDefault
		actualNilable := positionNilable(actualExpression, truth(actualFact["nilable"]))
		if optional != (truth(actual["optional"]) || actualNilable) {
			difference("argument-optional", optional, truth(actual["optional"]) || actualNilable, i+1)
		}
	}
	if len(returns) != len(stubReturns) {
		difference("return-annotation-count", len(returns), len(stubReturns), 0)
	}
	expectedWhen := []string{}
	for key, value := range native {
		if truth(value) && (predicates[key] == "Secret" || strings.HasPrefix(key, "SecretWhen") || strings.HasPrefix(key, "SecretIn") || strings.HasPrefix(key, "SecretOn")) {
			expectedWhen = append(expectedWhen, key)
		}
	}
	sort.Strings(expectedWhen)
	functionSecret := len(expectedWhen) > 0 || truth(native["SecretReturns"])
	for i := 0; i < min(len(returns), len(stubReturns)); i++ {
		value, actual := asMap(returns[i]), asMap(stubReturns[i])
		absence := truth(native["MayReturnNothing"])
		if absent != nil {
			absence = *absent
		}
		expectedFact := evMustFact(nativeTypeFact(value, absence))
		actualFact := evMustFact(simpleTypeFact(asString(actual["type"])))
		expectedExpression := positionExpression(asString(expectedFact["type"]), truth(value["StrideIndex"]), truth(value["Nilable"]))
		structural("return", i+1, expectedExpression, asString(actual["type"]))
		for _, key := range []string{"type", "nilable"} {
			var expected, actualValue any
			if key == "type" {
				expected = carrier(expectedExpression, "native")
				expression := asString(actual["type"])
				if facts == nil {
					expression = asString(actualFact["type"])
				}
				actualValue = carrier(expression, "tsm")
			} else {
				expected = expectedFact[key]
				actualValue = actualFact[key]
				if facts != nil {
					actualValue = positionNilable(asString(actual["type"]), truth(actualFact[key]))
				}
			}
			if !evEqual(expected, actualValue) {
				difference("return-"+key, expectedFact[key], actualFact[key], i+1)
			}
		}
		aggregate := value["Type"] == "table" || !evContains([]string{"number", "boolean", "string", "WOWGUID", "fileID", "time_t", "normalizedValue", "Seconds"}, asString(expectedFact["type"]))
		if !widget && !aggregate {
			secret := (functionSecret || truth(value["ConditionalSecret"]) || truth(value["SecretValue"])) && !truth(value["NeverSecret"]) && !truth(native["ReturnsNeverSecret"])
			if secret != truth(actualFact["secret"]) {
				difference("return-secret-scalar", secret, actualFact["secret"], i+1)
			}
		}
	}
	security := asMap(stub["security"])
	firstWords := func(v any) []string {
		r := []string{}
		for _, text := range evStrings(v) {
			words := strings.Fields(text)
			if len(words) > 0 {
				r = append(r, words[0])
			}
		}
		sort.Strings(r)
		return r
	}
	actualWhen := firstWords(security["secret-when"])
	if !evEqual(evList(expectedWhen), evList(actualWhen)) {
		difference("secret-predicates", evList(expectedWhen), evList(actualWhen), 0)
	}
	expectedPre := []string{}
	for key, kind := range predicates {
		if kind == "Precondition" && truth(native[key]) {
			expectedPre = append(expectedPre, key)
		}
	}
	sort.Strings(expectedPre)
	actualPre := firstWords(security["secret-precondition"])
	if !evEqual(evList(expectedPre), evList(actualPre)) {
		difference("secret-preconditions", evList(expectedPre), evList(actualPre), 0)
	}
	aspects := []string{}
	for _, key := range []string{"SecretArgumentsAddAspect", "SecretReturnsForAspect"} {
		for _, aspect := range asList(native[key]) {
			s := asString(aspect)
			if m := asMap(aspect); m != nil {
				s = asString(m["$ref"])
			}
			parts := strings.Split(s, ".")
			aspects = append(aspects, parts[len(parts)-1])
		}
	}
	aspects = evSortedUnique(aspects)
	actualAspects := evStrings(security["secret-aspect"])
	sort.Strings(actualAspects)
	if !evEqual(evList(aspects), evList(actualAspects)) {
		difference("secret-aspects", evList(aspects), evList(actualAspects), 0)
	}
	if policy := native["SecretArguments"]; policy != nil {
		expectedPolicy := policy
		if mapped, ok := map[string]string{"AllowedWhenTainted": "tainted", "AllowedWhenUntainted": "untainted", "NotAllowed": "none"}[asString(policy)]; ok {
			expectedPolicy = mapped
		}
		var actualPolicy any
		args := evStrings(security["secret-args"])
		if len(args) > 0 {
			w := strings.Fields(args[0])
			if len(w) > 0 {
				actualPolicy = w[0]
			}
		}
		if !evEqual(expectedPolicy, actualPolicy) {
			difference("secret-argument-policy", expectedPolicy, actualPolicy, 0)
			if actualPolicy == "tainted" && (truth(native["ConstSecretAccessor"]) || truth(native["SecretArgumentsAddAspect"])) {
				row := asMap(differences[len(differences)-1])
				row["disposition"] = "documented-tsm-const-accessor-or-aspect-setter-policy-override"
				row["basis"] = "TSM secret_stubs.rs entry_secrecy deliberately promotes these argument policies to tainted."
			}
		}
		allPositions := []any{}
		for i := range arguments {
			allPositions = append(allPositions, i+1)
		}
		for _, tag := range args {
			scope := scopedPolicy(tag, stubArgs)
			if truth(scope["unknown_names"]) || truth(scope["unsupported"]) {
				difference("unsupported-secret-argument-scope", nil, scope, 0)
			} else if scope["scope"] == "named" && !evEqual(scope["positions"], allPositions) {
				difference("secret-argument-scope", allPositions, scope, 0)
			}
		}
	}
	return differences, nil
}
func effectiveFact(c *catalogue, projection declarationProjection, name string, stub M, widget bool, carrierAliases map[string]string) (M, error) {
	owner, method := "", name
	if i := strings.LastIndex(name, "."); i >= 0 {
		owner, method = name[:i], name[i+1:]
	}
	var item M
	if owner != "" {
		item = asMap(asMap(asMap(c.defs[owner])["methods"])[method])
	} else {
		item = asMap(c.globals[method])
	}
	if item == nil {
		return nil, fmt.Errorf("native symbol missing from effective catalogue: %s", name)
	}
	metadata := asMap(evGet(item, "callable_contract", item))
	signatures, e := projection.signatures(name)
	if e != nil {
		return nil, e
	}
	signatureReturns := [][]string{}
	for _, s := range signatures {
		parts, e := SplitTopLevel(s, ':')
		if e != nil {
			return nil, e
		}
		values := []string{}
		if len(parts) > 1 {
			values, e = SplitTopLevel(strings.Join(parts[1:], ":"), ',')
			if e != nil {
				return nil, e
			}
		}
		signatureReturns = append(signatureReturns, values)
	}
	tsmSecrets := []any{}
	for _, v := range asList(stub["returns"]) {
		fact, e := simpleTypeFact(asString(asMap(v)["type"]))
		if e != nil {
			return nil, e
		}
		tsmSecrets = append(tsmSecrets, fact["secret"])
	}
	var opaqueCarrier func(string, []string) bool
	opaqueCarrier = func(expression string, seen []string) bool {
		members, e := SplitTopLevel(strings.TrimSuffix(expression, "..."), '|')
		if e != nil {
			return false
		}
		if evContains(members, "any") {
			return false
		}
		for _, token := range members {
			token = strings.TrimPrefix(strings.TrimSpace(token), "W.")
			if evContains([]string{"OpaqueNumber", "OpaqueText", "OpaqueValue"}, token) {
				return true
			}
			if alias, ok := carrierAliases[token]; ok && !evContains(seen, token) && opaqueCarrier(alias, append(append([]string{}, seen...), token)) {
				return true
			}
		}
		return false
	}
	opacity, unchecked := []any{}, []any{}
	for i, secret := range tsmSecrets {
		comparable := []string{}
		for _, values := range signatureReturns {
			if i < len(values) {
				comparable = append(comparable, strings.TrimSuffix(values[i], "..."))
			} else if len(values) > 0 && strings.HasSuffix(values[len(values)-1], "...") {
				comparable = append(comparable, strings.TrimSuffix(values[len(values)-1], "..."))
			}
		}
		if len(comparable) == 0 {
			unchecked = append(unchecked, M{"position": i + 1, "reason": "return tuple differs; inspect the explicit arity discrepancy"})
			continue
		}
		opaque := true
		for _, v := range comparable {
			opaque = opaque && opaqueCarrier(v, nil)
		}
		if opaque != truth(secret) {
			disposition := "workspace-opaque-carrier-versus-tsm-public-type"
			if truth(secret) {
				disposition = "review-effective-projection-of-tsm-secret-return"
			} else if truth(metadata["secret_effect"]) {
				disposition = "intentional-reviewed-handle-secret-effect"
			} else if widget {
				disposition = "intentional-conservative-native-widget-projection"
			}
			opacity = append(opacity, M{"position": i + 1, "tsm_secret_wrapper": secret, "workspace_opaque_carrier": opaque, "disposition": disposition})
		}
	}
	dataReturns := []string{}
	for _, values := range signatureReturns {
		for _, v := range values {
			for _, token := range regexp.MustCompile(`\b[A-Za-z_]\w*\b`).FindAllString(v, -1) {
				if c.secretStructs[token] {
					dataReturns = append(dataReturns, token)
				}
			}
		}
	}
	return M{"symbol": name, "emitted_signatures": evList(signatures), "workspace_result_secret": c.resultSecret(metadata), "reviewed_secret_effect": truth(metadata["secret_effect"]), "secret_data_returns": evList(evSortedUnique(dataReturns)), "tsm_secret_returns": tsmSecrets, "opacity_differences": opacity, "unchecked_return_positions": unchecked}, nil
}
