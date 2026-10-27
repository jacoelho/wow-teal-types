package wowapi

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

func effectiveFact(c *catalogue, projection declarationProjection, name string, stub tsmStub, widget bool, carrierAliases map[string]string) (effectiveDeclaration, error) {
	owner, method := "", name
	if i := strings.LastIndex(name, "."); i >= 0 {
		owner, method = name[:i], name[i+1:]
	}
	var item *catalogueFunction
	if owner != "" {
		if def := c.defs[owner]; def != nil {
			item = def.Methods[method]
		}
	} else {
		item = c.globals[method]
	}
	if item == nil {
		return effectiveDeclaration{}, fmt.Errorf("native symbol missing from effective catalogue: %s", name)
	}
	metadata := item.contract()
	signatures, e := projection.signatures(name)
	if e != nil {
		return effectiveDeclaration{}, e
	}
	signatureReturns := [][]string{}
	for _, s := range signatures {
		parts, e := SplitTopLevel(s, ':')
		if e != nil {
			return effectiveDeclaration{}, e
		}
		values := []string{}
		if len(parts) > 1 {
			values, e = SplitTopLevel(strings.Join(parts[1:], ":"), ',')
			if e != nil {
				return effectiveDeclaration{}, e
			}
		}
		signatureReturns = append(signatureReturns, values)
	}
	tsmSecrets := []bool{}
	for _, v := range stub.Returns {
		fact, e := simpleTypeFact(v.Type)
		if e != nil {
			return effectiveDeclaration{}, e
		}
		tsmSecrets = append(tsmSecrets, fact.Secret)
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
			if evContains([]string{"OpaqueNumber", "OpaqueText", "OpaqueValue", "unknown"}, token) {
				return true
			}
			if alias, ok := carrierAliases[token]; ok && !evContains(seen, token) && opaqueCarrier(alias, append(append([]string{}, seen...), token)) {
				return true
			}
		}
		return false
	}
	opacity, unchecked := []opacityDifference{}, []uncheckedReturn{}
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
			unchecked = append(unchecked, uncheckedReturn{Position: i + 1, Reason: "return tuple differs; inspect the explicit arity discrepancy"})
			continue
		}
		opaque := true
		for _, v := range comparable {
			opaque = opaque && opaqueCarrier(v, nil)
		}
		if opaque != secret {
			disposition := "workspace-opaque-carrier-versus-tsm-public-type"
			if secret {
				disposition = "review-effective-projection-of-tsm-secret-return"
			} else if metadata.SecretEffect {
				disposition = "intentional-reviewed-handle-secret-effect"
			} else if widget {
				disposition = "intentional-conservative-native-widget-projection"
			}
			opacity = append(opacity, opacityDifference{Position: i + 1, TSMSecretWrapper: secret, WorkspaceOpaqueCarrier: opaque, Disposition: disposition})
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
	return effectiveDeclaration{Symbol: name, EmittedSignatures: signatures, WorkspaceResultSecret: c.resultSecret(metadata), ReviewedSecretEffect: new(metadata.SecretEffect), SecretDataReturns: evSortedUnique(dataReturns), TSMSecretReturns: tsmSecrets, OpacityDifferences: opacity, UncheckedReturnPositions: unchecked}, nil
}

func compareFunction(native NativeFunction, stub tsmStub, predicates map[string]string, widget bool, facts *typeFacts, absent *bool) (differences []contractDifference, err error) {
	defer recoverFact(&err)
	differences = []contractDifference{}
	add := func(kind string, expected, actual differenceValue, position int) {
		differences = append(differences, difference(kind, expected, actual, position))
	}
	expand := func(raw []NativeParameter) []NativeParameter {
		r := []NativeParameter{}
		for _, v := range raw {
			if v.Type == "uiRect" {
				v.Type = "number"
				v.Mixin = "number"
				for range 4 {
					r = append(r, v)
				}
			} else {
				r = append(r, v)
			}
		}
		return r
	}
	arguments, returns := expand(native.Arguments), expand(native.Returns)
	generics := stub.Generics
	carrier := func(value, side string) string {
		if facts == nil {
			return value
		}
		var g []string
		if side == "tsm" {
			g = generics
		}
		f := facts.fact(value, side, g)
		if f.Unsupported == "" {
			return f.CarrierType
		}
		return value
	}
	positionExpression := func(expression string, variadic, nilable bool) string {
		if facts != nil && variadic && !facts.fact(expression, "tsm", generics).Variadic {
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
		f := facts.fact(expression, "tsm", generics)
		if kindOf(f.Tree) == "variadic" {
			return nullableFact(f.Tree.children[0])
		}
		if f.Unsupported == "" {
			return f.Nilable
		}
		return fallback
	}
	structural := func(side string, position int, nativeExpression, tsmExpression string) {
		if facts == nil {
			return
		}
		expected, actual := facts.fact(nativeExpression, "native", nil), facts.fact(tsmExpression, "tsm", generics)
		if expected.Unsupported != "" || actual.Unsupported != "" {
			return
		}
		features := evSortedUnique(append(append([]string{}, expected.Features...), actual.Features...))
		if len(features) == 0 {
			return
		}
		if !evEqual(expected.Tree, actual.Tree) || !evEqual(expected.Binders, actual.Binders) {
			d := difference(side+"-structural-contract", differenceValue{fact: &expected}, differenceValue{fact: &actual}, position)
			d.Features = features
			differences = append(differences, d)
		}
	}
	stubArgs, stubReturns := stub.Arguments, stub.Returns
	if len(arguments) != len(stubArgs) {
		add("argument-count", countDifference(len(arguments)), countDifference(len(stubArgs)), 0)
	}
	annotated := map[string]tsmParameter{}
	for _, v := range stub.Parameters {
		annotated[v.Name] = v
	}
	for i, value := range arguments {
		if i >= len(stubArgs) {
			break
		}
		name := stubArgs[i]
		actual, ok := annotated[name]
		if !ok {
			add("argument-annotation-missing", textDifference(value.Type), textDifference(name), i+1)
			continue
		}
		expectedFact, actualFact := evMustFact(nativeTypeFact(value, false)), evMustFact(simpleTypeFact(actual.Type))
		expectedExpression := positionExpression(expectedFact.Type, value.StrideIndex != nil && *value.StrideIndex != 0, enabled(value.Nilable))
		actualExpression := positionExpression(actual.Type, strings.HasPrefix(name, "..."), false)
		structural("argument", i+1, expectedExpression, actualExpression)
		actualCarrierExpression := actualExpression
		if facts == nil {
			actualCarrierExpression = actualFact.Type
		}
		if carrier(expectedExpression, "native") != carrier(actualCarrierExpression, "tsm") {
			add("argument-type", textDifference(expectedFact.Type), textDifference(actualFact.Type), i+1)
		}
		optional := enabled(value.Nilable) || value.HasDefault
		actualOptional := actual.Optional || positionNilable(actualExpression, actualFact.Nilable)
		if optional != actualOptional {
			add("argument-optional", boolDifference(optional), boolDifference(actualOptional), i+1)
		}
	}
	if len(returns) != len(stubReturns) {
		add("return-annotation-count", countDifference(len(returns)), countDifference(len(stubReturns)), 0)
	}
	expectedWhen := []string{}
	for _, key := range nativeSecretPredicateNames {
		if native.Policy.precondition(key) {
			expectedWhen = append(expectedWhen, key)
		}
	}
	for key, kind := range predicates {
		if kind == "Secret" && native.Policy.precondition(key) && !evContains(expectedWhen, key) {
			expectedWhen = append(expectedWhen, key)
		}
	}
	sort.Strings(expectedWhen)
	functionSecret := len(expectedWhen) > 0 || enabled(native.Policy.SecretReturns)
	for i := 0; i < min(len(returns), len(stubReturns)); i++ {
		value, actual := returns[i], stubReturns[i]
		absence := enabled(native.Policy.MayReturnNothing)
		if absent != nil {
			absence = *absent
		}
		expectedFact, actualFact := evMustFact(nativeTypeFact(value, absence)), evMustFact(simpleTypeFact(actual.Type))
		expectedExpression := positionExpression(expectedFact.Type, value.StrideIndex != nil && *value.StrideIndex != 0, enabled(value.Nilable))
		structural("return", i+1, expectedExpression, actual.Type)
		expression := actual.Type
		if facts == nil {
			expression = actualFact.Type
		}
		if carrier(expectedExpression, "native") != carrier(expression, "tsm") {
			add("return-type", textDifference(expectedFact.Type), textDifference(actualFact.Type), i+1)
		}
		if expectedFact.Nilable != positionNilable(actual.Type, actualFact.Nilable) {
			add("return-nilable", boolDifference(expectedFact.Nilable), boolDifference(actualFact.Nilable), i+1)
		}
		aggregate := value.Type == "table" || !evContains([]string{"number", "boolean", "string", "WOWGUID", "fileID", "time_t", "normalizedValue", "Seconds"}, expectedFact.Type)
		if !widget && !aggregate {
			secret := (functionSecret || enabled(value.ConditionalSecret) || enabled(value.SecretValue)) && !enabled(value.NeverSecret) && !enabled(native.Policy.ReturnsNeverSecret)
			if secret != actualFact.Secret {
				add("return-secret-scalar", boolDifference(secret), boolDifference(actualFact.Secret), i+1)
			}
		}
	}
	security := stub.Security
	firstWords := func(values []string) []string {
		r := []string{}
		for _, text := range values {
			words := strings.Fields(text)
			if len(words) > 0 {
				r = append(r, words[0])
			}
		}
		sort.Strings(r)
		return r
	}
	actualWhen := firstWords(security["secret-when"])
	if !evEqual(expectedWhen, actualWhen) {
		add("secret-predicates", textsDifference(expectedWhen), textsDifference(actualWhen), 0)
	}
	expectedPre := []string{}
	for key, kind := range predicates {
		if kind == "Precondition" && native.Policy.precondition(key) {
			expectedPre = append(expectedPre, key)
		}
	}
	sort.Strings(expectedPre)
	actualPre := firstWords(security["secret-precondition"])
	if !evEqual(expectedPre, actualPre) {
		add("secret-preconditions", textsDifference(expectedPre), textsDifference(actualPre), 0)
	}
	aspects := []string{}
	for _, group := range [][]NativeReference{native.Policy.SecretArgumentsAddAspect, native.Policy.SecretReturnsForAspect} {
		for _, aspect := range group {
			parts := strings.Split(aspect.Name, ".")
			aspects = append(aspects, parts[len(parts)-1])
		}
	}
	aspects = evSortedUnique(aspects)
	actualAspects := append([]string{}, security["secret-aspect"]...)
	sort.Strings(actualAspects)
	if !evEqual(aspects, actualAspects) {
		add("secret-aspects", textsDifference(aspects), textsDifference(actualAspects), 0)
	}
	if policy := native.Policy.SecretArguments; policy != nil {
		expectedPolicy := *policy
		if mapped, ok := map[string]string{"AllowedWhenTainted": "tainted", "AllowedWhenUntainted": "untainted", "NotAllowed": "none"}[*policy]; ok {
			expectedPolicy = mapped
		}
		var actualPolicy *string
		args := security["secret-args"]
		if len(args) > 0 {
			words := strings.Fields(args[0])
			if len(words) > 0 {
				actualPolicy = &words[0]
			}
		}
		if actualPolicy == nil || expectedPolicy != *actualPolicy {
			add("secret-argument-policy", textDifference(expectedPolicy), differenceValue{text: actualPolicy}, 0)
			if actualPolicy != nil && *actualPolicy == "tainted" && (enabled(native.Policy.ConstSecretAccessor) || len(native.Policy.SecretArgumentsAddAspect) > 0) {
				row := &differences[len(differences)-1]
				row.Disposition = "documented-tsm-const-accessor-or-aspect-setter-policy-override"
				row.Basis = "TSM secret_stubs.rs entry_secrecy deliberately promotes these argument policies to tainted."
			}
		}
		positions := []int{}
		for i := range arguments {
			positions = append(positions, i+1)
		}
		for _, tag := range args {
			scope := scopedPolicy(tag, stubArgs)
			if len(scope.UnknownNames) > 0 || scope.Unsupported != "" {
				add("unsupported-secret-argument-scope", differenceValue{}, differenceValue{scope: &scope}, 0)
			} else if scope.Scope == "named" && !evEqual(scope.Positions, positions) {
				add("secret-argument-scope", differenceValue{positions: &positions}, differenceValue{scope: &scope}, 0)
			}
		}
	}
	return differences, nil
}
