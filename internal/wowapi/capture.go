package wowapi

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"unicode/utf8"
)

const captureMaxBytes = 256 << 10
const captureMaxNodes = 20000

var captureQueryTypes = []string{"Animation", "RadialProgress", "Region", "LayeredRegion", "ScriptObject", "UIObject", "VectorGraphics"}
var captureBaseFields = wordSet("schema probeVersion expectedBuild sourceFingerprint declarationFingerprint clientVersion clientBuild clientDate clientInterface projectID context outcome expectedMethods limits observedMethods presentMethods capabilityCaptureComplete verifiesSignatures verifiesSecrecy cleanup families")
var captureNewFields = wordSet("mode invocations invokedMethods verifiedMethods invocationCaptureComplete fixtureSHA256 hierarchy")
var diagnosticSetters = wordSet("setup.font_object_set setup.clock_set_time setup.duration_set_clock setup.duration_set_span sentinel.set retry.update zero_interval.set zero_interval.update")
var diagnosticReads = wordSet("baseline.expected_before baseline.expected_after baseline.font_after sentinel.read retry.expected retry.font zero_interval.expected zero_interval.font")
var diagnosticComparisons = wordSet("setup.font_object_identity setup.duration_clock_identity setup.duration_clock_time setup.duration_remaining baseline.expected_stable baseline.font_matches sentinel.matches retry.matches zero_interval.matches")
var diagnosticFailures = wordSet("not_run absent inaccessible inaccessible_table invalid_object lookup_error not_function call_error")
var bindingSamples = wordSet("baseline.expected baseline.font projection.font manual_step.expected manual_step.font default_clock.expected default_clock.font")
var experimentSetters = wordSet("projection.set manual_step.set_time manual_step.update manual_step.restore_time default_clock.clear default_clock.span default_clock.update")
var experimentComparisons = wordSet("projection.matches manual_step.clock_time manual_step.remaining manual_step.matches default_clock.time default_clock.clock_time default_clock.remaining default_clock.matches")
var captureState = regexp.MustCompile(`^[a-z_]{1,64}$`)
var captureCheck = regexp.MustCompile(`^[A-Za-z0-9_.-]{1,96}$`)
var captureHash = regexp.MustCompile(`^[0-9a-f]{64}$`)
var captureHex = regexp.MustCompile(`^([0-9a-f]{2}){0,128}$`)

func capUnion(sets ...map[string]bool) map[string]bool {
	result := map[string]bool{}
	for _, set := range sets {
		for key := range set {
			result[key] = true
		}
	}
	return result
}
func capEqualSet(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}
func capExact(value M, fields map[string]bool, label string) {
	if value == nil || len(value) != len(fields) {
		reject("unexpected %s fields", label)
	}
	for key := range value {
		if !fields[key] {
			reject("unexpected %s fields", label)
		}
	}
}
func capInteger(value any) bool {
	switch x := value.(type) {
	case int, int64:
		return true
	case json.Number:
		_, err := x.Int64()
		return err == nil && !strings.ContainsAny(string(x), ".eE")
	}
	return false
}
func capInt(value any) int {
	if !capInteger(value) {
		reject("capture expected integer")
	}
	return asInt(value)
}
func capString(value any) string {
	s, ok := value.(string)
	if !ok {
		reject("capture expected string")
	}
	return s
}
func capBool(value any) bool {
	b, ok := value.(bool)
	if !ok {
		reject("capture expected boolean")
	}
	return b
}
func capArray(value any, label string) []any {
	rows, ok := value.([]any)
	if !ok {
		reject("invalid %s", label)
	}
	return rows
}
func capSet(values any, label string) map[string]bool {
	result := map[string]bool{}
	for _, v := range capArray(values, label) {
		s := capString(v)
		if result[s] {
			reject("duplicate %s", label)
		}
		result[s] = true
	}
	return result
}
func capRequired(plan M, name string, expected map[string]bool) {
	if !capEqualSet(capSet(plan[name], "binding requirements: "+name), expected) {
		reject("invalid binding requirements: %s", name)
	}
}
func capBounded(v any, depth int, budget *int) {
	*budget--
	if *budget < 0 || depth > 64 {
		reject("capture exceeds node/depth limit")
	}
	switch x := v.(type) {
	case M:
		for k, v := range x {
			if k == "$ref" || k == "$expression" {
				reject("nonliteral capture key")
			}
			capBounded(v, depth+1, budget)
		}
	case []any:
		for _, v := range x {
			capBounded(v, depth+1, budget)
		}
	case string:
		if utf8.RuneCountInString(x) > 512 {
			reject("capture string exceeds 512 characters")
		}
	case float64:
		if math.IsNaN(x) || math.IsInf(x, 0) {
			reject("nonfinite capture value")
		}
	case json.Number:
		if strings.ContainsAny(string(x), ".eE") {
			value, err := x.Float64()
			if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
				reject("nonfinite capture value")
			}
		}
	case nil, bool, int, int64:
	default:
		reject("unsupported capture value")
	}
}
func captureJSONValue(decoder *json.Decoder, depth int, budget *int) any {
	*budget--
	if *budget < 0 || depth > 64 {
		reject("capture exceeds node/depth limit")
	}
	token, err := decoder.Token()
	if err != nil {
		reject("invalid capture JSON: %s", err)
	}
	if delimiter, ok := token.(json.Delim); ok {
		switch delimiter {
		case '{':
			value := M{}
			for decoder.More() {
				key, err := decoder.Token()
				if err != nil {
					reject("invalid capture JSON: %s", err)
				}
				name, ok := key.(string)
				if !ok {
					reject("invalid capture JSON key")
				}
				if _, found := value[name]; found {
					reject("duplicate capture JSON key")
				}
				value[name] = captureJSONValue(decoder, depth+1, budget)
			}
			if _, err = decoder.Token(); err != nil {
				reject("invalid capture JSON: %s", err)
			}
			return value
		case '[':
			value := []any{}
			for decoder.More() {
				value = append(value, captureJSONValue(decoder, depth+1, budget))
			}
			if _, err = decoder.Token(); err != nil {
				reject("invalid capture JSON: %s", err)
			}
			return value
		default:
			reject("invalid capture JSON delimiter")
		}
	}
	return token
}
func captureLiteral(p *luaParser, depth int) any {
	if depth > 64 {
		reject("capture nesting exceeds 64")
	}
	if p.peek() != "{" {
		token := p.peek()
		if token == "" {
			reject("capture must contain literal data only")
		}
		if token == "+" || token == "-" {
			sign := p.take()
			raw := p.peek()
			if raw == "" || raw[0] < '0' || raw[0] > '9' {
				reject("invalid signed capture literal")
			}
			value := p.atom(depth)
			n, ok, integer := luaNumber(value)
			if !ok {
				reject("invalid signed capture literal")
			}
			if sign == "-" {
				n = -n
			}
			if integer {
				return int64(n)
			}
			return n
		}
		if token == "true" || token == "false" || token == "nil" || token[0] >= '0' && token[0] <= '9' {
			return p.atom(depth)
		}
		if token[0] == '\'' || token[0] == '"' {
			for i := 1; i < len(token)-1; i++ {
				if token[i] == '\\' {
					i++
					if i >= len(token)-1 || !strings.ContainsRune("0123456789abfnrtv\\\"'", rune(token[i])) {
						reject("unsupported capture string escape")
					}
				}
			}
			return p.atom(depth)
		}
		reject("capture must contain literal data only")
	}
	p.take("{")
	entries := []any{}
	fields := M{}
	for p.peek() != "}" {
		key := ""
		hasKey := false
		if p.peek() == "[" {
			p.take()
			value := captureLiteral(p, depth+1)
			var ok bool
			key, ok = value.(string)
			if !ok {
				reject("duplicate or unsupported capture key")
			}
			hasKey = true
			p.take("]")
			p.take("=")
		} else if p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].text == "=" {
			key = p.take()
			hasKey = true
			p.take("=")
		}
		if hasKey {
			if _, exists := fields[key]; exists {
				reject("duplicate or unsupported capture key")
			}
			fields[key] = captureLiteral(p, depth+1)
		} else {
			entries = append(entries, captureLiteral(p, depth+1))
		}
		if p.peek() == "," || p.peek() == ";" {
			p.take()
		} else if p.peek() != "}" {
			reject("expected literal table separator")
		}
	}
	p.take("}")
	if len(entries) > 0 && len(fields) > 0 {
		reject("mixed capture table")
	}
	if len(fields) > 0 {
		return fields
	}
	return entries
}
func readCaptureData(data []byte, savedVariables bool) (result M, err error) {
	defer captureError(&err)
	if len(data) > captureMaxBytes {
		reject("capture exceeds byte limit")
	}
	if !utf8.Valid(data) {
		reject("capture must be UTF-8")
	}
	var value any
	if savedVariables {
		p := luaParser{tokens: luaTokens(string(data), "capture"), source: "capture"}
		if len(p.tokens) > captureMaxNodes*8 {
			reject("capture exceeds token limit")
		}
		p.take("WoWAPIProbeDB")
		p.take("=")
		value = captureLiteral(&p, 0)
		if p.peek() != "" {
			reject("trailing capture statements")
		}
	} else {
		decoder := json.NewDecoder(bytes.NewReader(data))
		decoder.UseNumber()
		budget := captureMaxNodes
		value = captureJSONValue(decoder, 0, &budget)
		var tail any
		if e := decoder.Decode(&tail); e != io.EOF {
			reject("trailing capture JSON data")
		}
	}
	budget := captureMaxNodes
	capBounded(value, 0, &budget)
	result, ok := value.(map[string]any)
	if !ok {
		reject("capture must be an object")
	}
	return result, nil
}
func capObservations(value any, allowed, valueCases map[string]bool) map[string]M {
	result := map[string]M{}
	for _, item := range capArray(value, "capture observations array") {
		row := asMap(item)
		if row == nil || row["caseID"] == nil || row["state"] == nil || len(row) < 2 || len(row) > 3 {
			reject("unexpected observation fields")
		}
		for key := range row {
			if key != "caseID" && key != "state" && key != "value" {
				reject("unexpected observation fields")
			}
		}
		name, state := capString(row["caseID"]), capString(row["state"])
		if !allowed[name] || result[name] != nil || !captureState.MatchString(state) {
			reject("duplicate, unknown or invalid observation")
		}
		payload, present := row["value"]
		if present {
			_, s := payload.(string)
			_, b := payload.(bool)
			if !valueCases[name] || state != "available" || !(s || b) {
				reject("unexpected observation payload")
			}
		}
		if valueCases[name] && state == "available" && !present {
			reject("missing available query value")
		}
		result[name] = row
	}
	return result
}
func capStates(rows any, allowed, setters, reads map[string]bool, label string) map[string]string {
	found := map[string]string{}
	for _, item := range capArray(rows, "binding "+label) {
		row := asMap(item)
		capExact(row, wordSet("caseID state"), "binding "+label)
		name, state := capString(row["caseID"]), capString(row["state"])
		if !allowed[name] || found[name] != "" {
			reject("duplicate or unknown binding %s", label)
		}
		states := wordSet("matched mismatch not_comparable not_run")
		if setters[name] {
			states = capUnion(wordSet("called_void unexpected_tuple"), diagnosticFailures)
		} else if reads[name] {
			states = capUnion(wordSet("public_empty public_nonempty nil secret wrong_type oversized unexpected_tuple"), diagnosticFailures)
		}
		if !states[state] {
			reject("invalid binding %s state: %s", label, name)
		}
		found[name] = state
	}
	return found
}
func capDiagnostics(report, plan M, mode string) map[string]string {
	required := capUnion(diagnosticSetters, diagnosticReads, diagnosticComparisons)
	capRequired(plan, "bindingDiagnosticCases", required)
	found := capStates(report["bindingDiagnostics"], required, diagnosticSetters, diagnosticReads, "diagnostic")
	if mode == "public_invocations" && len(found) != len(required) || mode == "presence" && len(found) > 0 {
		reject("invalid binding diagnostic coverage")
	}
	if len(found) == 0 {
		return found
	}
	inputs := map[string][]string{"baseline.expected_stable": {"baseline.expected_before", "baseline.expected_after"}, "baseline.font_matches": {"baseline.expected_after", "baseline.font_after"}, "sentinel.matches": {"sentinel.read"}, "retry.matches": {"retry.expected", "retry.font"}, "zero_interval.matches": {"zero_interval.expected", "zero_interval.font"}}
	for name, operands := range inputs {
		if found[name] == "matched" || found[name] == "mismatch" {
			for _, operand := range operands {
				if found[operand] != "public_empty" && found[operand] != "public_nonempty" {
					reject("unavailable diagnostic inputs cannot establish equality: %s", name)
				}
			}
		}
	}
	return found
}
func capExperiments(report, plan M, mode string, diagnostics map[string]string) (map[string]string, map[string]M) {
	required := capUnion(experimentSetters, experimentComparisons)
	capRequired(plan, "bindingTextSampleCases", bindingSamples)
	capRequired(plan, "bindingExperimentCases", required)
	found := map[string]string{}
	public := map[string][]byte{}
	samples := map[string]M{}
	failures := capUnion(diagnosticFailures, wordSet("nil secret wrong_type oversized unexpected_tuple"))
	for _, item := range capArray(report["bindingTextSamples"], "binding experiment arrays") {
		row := asMap(item)
		state := capString(row["state"])
		fields := wordSet("caseID state")
		if state == "public" {
			fields = wordSet("caseID state hex byteLength firstNul")
		}
		capExact(row, fields, "binding text sample")
		name := capString(row["caseID"])
		if !bindingSamples[name] || found[name] != "" {
			reject("duplicate or unknown binding text sample")
		}
		if state == "public" {
			text := capString(row["hex"])
			if !captureHex.MatchString(text) {
				reject("invalid or oversized public fixture bytes")
			}
			value, _ := hex.DecodeString(text)
			if capInt(row["byteLength"]) != len(value) || capInt(row["firstNul"]) != bytes.IndexByte(value, 0)+1 {
				reject("inconsistent public fixture byte metadata")
			}
			public[name] = value
		} else if !failures[state] {
			reject("invalid binding text sample state")
		}
		found[name] = state
		samples[name] = row
	}
	checks := capStates(report["bindingExperimentChecks"], required, experimentSetters, nil, "experiment check")
	if mode == "presence" {
		if len(found) > 0 || len(checks) > 0 {
			reject("presence cannot contain binding experiment evidence")
		}
		return checks, samples
	}
	if len(found) != len(bindingSamples) || len(checks) != len(required) {
		reject("missing binding experiment evidence")
	}
	for sample, source := range map[string]string{"baseline.expected": "baseline.expected_after", "baseline.font": "baseline.font_after"} {
		state := diagnostics[source]
		if state != "public_empty" && state != "public_nonempty" {
			if found[sample] != state {
				reject("baseline sample contradicts its source read state")
			}
		} else if found[sample] == "public" && (len(public[sample]) == 0) != (state == "public_empty") {
			reject("baseline sample contradicts its source text length")
		}
	}
	pairs := map[string][2]string{"projection.matches": {"baseline.expected", "projection.font"}, "manual_step.matches": {"manual_step.expected", "manual_step.font"}, "default_clock.matches": {"default_clock.expected", "default_clock.font"}}
	for name, pair := range pairs {
		if checks[name] == "matched" || checks[name] == "mismatch" {
			left, lok := public[pair[0]]
			right, rok := public[pair[1]]
			if !lok || !rok {
				reject("unavailable fixture bytes cannot establish equality")
			}
			if (checks[name] == "matched") != bytes.Equal(left, right) {
				reject("binding experiment contradicts exact public fixture bytes")
			}
		}
	}
	left, lok := public["baseline.expected"]
	right, rok := public["baseline.font"]
	if lok && rok && (diagnostics["baseline.font_matches"] == "matched" || diagnostics["baseline.font_matches"] == "mismatch") && (diagnostics["baseline.font_matches"] == "matched") != bytes.Equal(left, right) {
		reject("baseline diagnostic contradicts exact public fixture bytes")
	}
	return checks, samples
}

func validateCapture(report, manifest, receipt M, manifestBytes []byte) (classification M, err error) {
	defer captureError(&err)
	plan := asMap(manifest["plan"])
	version := capString(plan["version"])
	if !wordSet("1.0.0 1.1.0 1.2.0 1.3.0 1.4.0 1.5.0")[version] || manifest["probe"] != version {
		reject("unsupported probe version")
	}
	fields := capUnion(captureBaseFields)
	if version != "1.0.0" {
		fields = capUnion(fields, captureNewFields)
	}
	if version == "1.3.0" || version == "1.4.0" || version == "1.5.0" {
		fields["bindingDiagnostics"] = true
	}
	if version == "1.4.0" || version == "1.5.0" {
		fields["bindingTextSamples"] = true
		fields["bindingExperimentChecks"] = true
	}
	capExact(report, fields, "capture")
	if receipt["build_manifest_sha256"] != sha256hex(manifestBytes) || !evEqual(receipt["files"], manifest["files"]) {
		reject("installed probe receipt identity mismatch")
	}
	runtimeFiles := asMap(manifest["files"])
	if runtimeFiles == nil {
		reject("invalid installed probe file identity")
	}
	for name, value := range runtimeFiles {
		if filepath.IsAbs(name) || strings.Contains(name, "\\") {
			reject("invalid installed probe file identity")
		}
		for component := range strings.SplitSeq(name, "/") {
			if component == ".." || component == "." || component == "" {
				reject("invalid installed probe file identity")
			}
		}
		if !captureHash.MatchString(capString(value)) {
			reject("invalid installed probe file identity")
		}
	}
	for _, key := range []string{"probeVersion", "expectedBuild", "sourceFingerprint", "declarationFingerprint"} {
		expected := plan[key]
		if key == "probeVersion" {
			expected = version
		}
		if !evEqual(report[key], expected) {
			reject("capture identity mismatch: %s", key)
		}
	}
	if capInt(report["schema"]) != 1 {
		reject("unsupported capture schema")
	}
	if capString(report["clientVersion"])+"."+capString(report["clientBuild"]) != plan["expectedBuild"] {
		reject("client build mismatch")
	}
	for key, expected := range map[string]any{"clientInterface": plan["expectedInterface"], "projectID": plan["expectedProjectID"]} {
		if capInt(report[key]) != capInt(expected) {
			reject("client identity mismatch: %s", key)
		}
	}
	limits := asMap(report["limits"])
	if report["context"] != "out_of_combat_public_owned_objects" || limits == nil || !evEqual(limits, plan["limits"]) {
		reject("capture context/limits mismatch")
	}
	for _, value := range limits {
		if !capInteger(value) {
			reject("capture context/limits mismatch")
		}
	}
	if report["verifiesSignatures"] != false || report["verifiesSecrecy"] != false {
		reject("capture cannot certify signatures or secrecy")
	}
	expected := map[string]M{}
	methods := map[string]bool{}
	for _, item := range capArray(plan["families"], "capture method requirements") {
		row := asMap(item)
		name := capString(row["name"])
		if expected[name] != nil {
			reject("invalid capture method requirements")
		}
		expected[name] = row
		for _, method := range capArray(row["methods"], "capture method requirements") {
			methods[capString(row["canonical"])+"."+capString(method)] = true
		}
	}
	if len(methods) != capInt(plan["expectedMethods"]) || capInt(report["expectedMethods"]) != len(methods) {
		reject("invalid capture method requirements")
	}
	mode := "presence"
	if value, ok := report["mode"]; ok {
		mode = capString(value)
	}
	if mode != "presence" && mode != "public_invocations" {
		reject("unsupported capture mode")
	}
	diagnostics := map[string]string{}
	checks := map[string]string{}
	samples := map[string]M{}
	if fields["bindingDiagnostics"] {
		diagnostics = capDiagnostics(report, plan, mode)
	}
	if fields["bindingTextSamples"] {
		checks, samples = capExperiments(report, plan, mode, diagnostics)
	}
	observed := map[string]M{}
	families := map[string]bool{}
	present := 0
	factoriesAvailable := true
	for _, item := range capArray(report["families"], "capture families") {
		family := asMap(item)
		capExact(family, wordSet("family factory methods queries"), "family")
		name := capString(family["family"])
		requirement, ok := expected[name]
		if !ok || families[name] {
			reject("duplicate or unknown capture family")
		}
		families[name] = true
		factory := capString(family["factory"])
		if !captureState.MatchString(factory) {
			reject("invalid constructor state")
		}
		factoriesAvailable = factoriesAvailable && factory == "available"
		ids := map[string]bool{}
		canonical := capString(requirement["canonical"])
		for _, method := range capArray(requirement["methods"], "capture method requirements") {
			ids[canonical+"."+capString(method)] = true
		}
		rows := capObservations(family["methods"], ids, nil)
		if len(rows) != len(ids) {
			reject("missing capture method case")
		}
		for key, row := range rows {
			observed[key] = row
			if row["state"] == "present" {
				present++
			}
		}
		prefix := canonical + ".query."
		queries := map[string]bool{prefix + "GetObjectType": true}
		for _, kind := range captureQueryTypes {
			queries[prefix+"IsObjectType."+kind] = true
		}
		queryRows := capObservations(family["queries"], queries, queries)
		if factory == "available" && len(queryRows) != len(queries) {
			reject("missing capture query case")
		}
		for name, row := range queryRows {
			if value, ok := row["value"]; ok {
				if strings.HasSuffix(name, "GetObjectType") {
					if _, ok = value.(string); !ok {
						reject("invalid native query result type")
					}
				} else if _, ok = value.(bool); !ok {
					reject("invalid native query result type")
				}
			}
		}
	}
	for key, count := range map[string]int{"observedMethods": len(observed), "presentMethods": present} {
		if capInt(report[key]) != count {
			reject("inconsistent capture summary: %s", key)
		}
	}
	cleanupIDs := wordSet("cleanup.DurationTextBinding.Disable cleanup.invocation.binding.1 cleanup.invocation.binding.2 cleanup.invocation.signals")
	cleanup := capObservations(report["cleanup"], cleanupIDs, nil)
	clean := true
	for _, row := range cleanup {
		clean = clean && row["state"] == "complete"
	}
	presenceComplete := mode == "presence" && len(families) == len(expected) && factoriesAvailable && present == len(methods) && clean && cleanup["cleanup.DurationTextBinding.Disable"] != nil
	if capBool(report["capabilityCaptureComplete"]) != presenceComplete {
		reject("inconsistent presence completion")
	}
	if version == "1.0.0" {
		outcome := "incomplete"
		if presenceComplete {
			outcome = "capabilities_collected"
		}
		if report["outcome"] != outcome {
			reject("inconsistent presence outcome")
		}
		return M{"kind": "presence", "present_methods": present, "successful_invocations": 0, "probe_reported_property_passes": 0, "method_cases": evList(sortedKeys(methods)), "unobserved_hierarchy": true, "verifies_secrecy": false}, nil
	}
	if report["fixtureSHA256"] != plan["fixtureSHA256"] {
		reject("fixture identity mismatch")
	}
	hierarchy := map[string]bool{}
	for _, item := range capArray(plan["hierarchy"], "hierarchy requirements") {
		row := asMap(item)
		for _, method := range capArray(row["methods"], "hierarchy requirements") {
			hierarchy[capString(row["canonical"])+".parent."+capString(row["parent"])+"."+capString(method)] = true
		}
	}
	hierarchyRows := capObservations(report["hierarchy"], hierarchy, nil)
	if presenceComplete && len(hierarchyRows) != len(hierarchy) {
		reject("missing hierarchy cases")
	}
	invoked, verified := 0, 0
	seen := map[string]bool{}
	invocationStates := wordSet("not_run absent inaccessible inaccessible_table lookup_error invalid_object not_function call_error passed property_mismatch projection_unavailable secret_result inaccessible_result unsupported_tuple fixture_unavailable fixture_setup_failed")
	validLabels := wordSet("nil boolean number string table userdata function thread secret inaccessible")
	successfulStates := wordSet("passed property_mismatch projection_unavailable secret_result inaccessible_result unsupported_tuple fixture_unavailable")
	for _, item := range capArray(report["invocations"], "invocation rows") {
		row := asMap(item)
		capExact(row, wordSet("caseID state invoked returnCount returns check behaviorVerified"), "invocation")
		name, state := capString(row["caseID"]), capString(row["state"])
		if !methods[name] || seen[name] {
			reject("duplicate or unknown invocation")
		}
		seen[name] = true
		if !invocationStates[state] {
			reject("invalid invocation state")
		}
		if !captureCheck.MatchString(capString(row["check"])) {
			reject("invalid invocation check identity")
		}
		called, behavior := capBool(row["invoked"]), capBool(row["behaviorVerified"])
		count := capInt(row["returnCount"])
		if count < 0 || count > 8 {
			reject("invalid invocation types/count")
		}
		labels := capArray(row["returns"], "invocation tuple labels")
		if len(labels) != count {
			reject("invalid invocation tuple labels")
		}
		opaque := map[string]bool{}
		for _, label := range labels {
			s := capString(label)
			if !validLabels[s] {
				reject("invalid invocation tuple labels")
			}
			if s == "secret" || s == "inaccessible" {
				opaque[s] = true
			}
		}
		passed := state == "passed"
		if len(opaque) > 0 && (behavior || !wordSet("secret_result inaccessible_result fixture_unavailable")[state]) {
			reject("opaque results cannot pass a property oracle")
		}
		if state == "secret_result" && !opaque["secret"] {
			reject("secret result state lacks secret result evidence")
		}
		if state == "inaccessible_result" && !opaque["inaccessible"] {
			reject("inaccessible result state lacks inaccessible result evidence")
		}
		if behavior != passed || passed && !called || !called && count != 0 {
			reject("contradictory invocation result")
		}
		if called && !successfulStates[state] {
			reject("failed call cannot be a successful invocation")
		}
		projection := version == "1.5.0" && name == "DurationTextBindingObject.UpdateFontString"
		if state == "projection_unavailable" && (!projection || !called || count != 0) {
			reject("invalid unavailable native projection")
		}
		if projection && called {
			if row["check"] != "font_matches_native_projection" {
				reject("incorrect native projection oracle")
			}
			if passed {
				valid := count == 0 && diagnostics["baseline.expected_stable"] == "matched" && diagnostics["sentinel.set"] == "called_void" && diagnostics["sentinel.matches"] == "matched" && checks["projection.set"] == "called_void"
				for _, sample := range []string{"baseline.expected", "baseline.font", "projection.font"} {
					valid = valid && samples[sample]["state"] == "public"
				}
				valid = valid && samples["baseline.font"]["hex"] == samples["projection.font"]["hex"] && samples["projection.font"]["hex"] != hex.EncodeToString([]byte("PROBE_SENTINEL"))
				if !valid {
					reject("native projection result lacks its required public controls")
				}
			}
		}
		if called {
			invoked++
		}
		if passed {
			verified++
		}
	}
	for key, count := range map[string]int{"invokedMethods": invoked, "verifiedMethods": verified} {
		if capInt(report[key]) != count {
			reject("inconsistent invocation summary")
		}
	}
	requiredCleanup := wordSet("cleanup.invocation.binding.1 cleanup.invocation.binding.2 cleanup.invocation.signals")
	allCleanup := true
	for name := range requiredCleanup {
		allCleanup = allCleanup && cleanup[name] != nil
	}
	complete := mode == "public_invocations" && len(seen) == len(methods) && verified == len(methods) && clean && allCleanup
	if complete && len(hierarchyRows) != len(hierarchy) {
		reject("missing hierarchy cases")
	}
	if capBool(report["invocationCaptureComplete"]) != complete {
		reject("inconsistent invocation completion")
	}
	if mode == "public_invocations" {
		outcome := "public_invocations_incomplete"
		if complete {
			outcome = "public_invocations_verified"
		}
		if len(seen) != len(methods) || len(observed) > 0 || len(families) > 0 || report["outcome"] != outcome {
			reject("invalid invocation coverage/outcome")
		}
	} else {
		outcome := "incomplete"
		if presenceComplete {
			outcome = "capabilities_collected"
		}
		if len(seen) > 0 || invoked != 0 || verified != 0 || report["outcome"] != outcome {
			reject("presence cannot contain invocation evidence")
		}
	}
	return M{"kind": mode, "present_methods": present, "successful_invocations": invoked, "probe_reported_property_passes": verified, "method_cases": evList(sortedKeys(methods)), "hierarchy_cases": evList(sortedKeys(hierarchyRows)), "verifies_secrecy": false}, nil
}

func readCaptureFile(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("capture evidence must be a regular file: %s", path)
	}
	descriptor, err := syscall.Open(path, syscall.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(descriptor), path)
	defer file.Close()
	info, err = file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("capture evidence must be a regular file: %s", path)
	}
	data, err := io.ReadAll(io.LimitReader(file, captureMaxBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > captureMaxBytes {
		return nil, fmt.Errorf("capture evidence exceeds byte limit: %s", path)
	}
	return data, nil
}

// IngestCapture validates caller-supplied data and retains its exact immutable
// input bytes. It does not authenticate observations or grant secret permissions.
func IngestCapture(capture, manifestPath, receiptPath, output string) (M, error) {
	raw, err := readCaptureFile(capture)
	if err != nil {
		return nil, err
	}
	manifestBytes, err := readCaptureFile(manifestPath)
	if err != nil {
		return nil, err
	}
	receiptBytes, err := readCaptureFile(receiptPath)
	if err != nil {
		return nil, err
	}
	report, err := readCaptureData(raw, filepath.Ext(capture) == ".lua")
	if err != nil {
		return nil, err
	}
	manifest, err := readCaptureData(manifestBytes, false)
	if err != nil {
		return nil, err
	}
	receipt, err := readCaptureData(receiptBytes, false)
	if err != nil {
		return nil, err
	}
	classification, err := validateCapture(report, manifest, receipt, manifestBytes)
	if err != nil {
		return nil, err
	}
	files := map[string][]byte{"capture" + filepath.Ext(capture): raw, "probe-build.json": manifestBytes, "installed-receipt.json": receiptBytes}
	for name, value := range map[string]any{"report.json": report, "classification.json": classification} {
		files[name], err = jsonPretty(value)
		if err != nil {
			return nil, err
		}
	}
	index := M{}
	for name, data := range files {
		index[name] = M{"sha256": sha256hex(data), "size": len(data)}
	}
	files["evidence.json"], err = jsonPretty(M{"schema_version": 1, "files": index, "scope": "validated caller-supplied data; acquisition receipt must be independently established"})
	if err != nil {
		return nil, err
	}
	if info, e := os.Lstat(output); e == nil {
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("existing capture evidence differs; never overwrite")
		}
		entries, e := os.ReadDir(output)
		if e != nil {
			return nil, e
		}
		if len(entries) != len(files) {
			return nil, fmt.Errorf("existing capture evidence differs; never overwrite")
		}
		for _, entry := range entries {
			expected, exists := files[entry.Name()]
			if !exists {
				return nil, fmt.Errorf("existing capture evidence differs; never overwrite")
			}
			data, e := readCaptureFile(filepath.Join(output, entry.Name()))
			if e != nil {
				return nil, e
			}
			if !bytes.Equal(data, expected) {
				return nil, fmt.Errorf("existing capture evidence differs; never overwrite")
			}
		}
		return classification, nil
	} else if !os.IsNotExist(e) {
		return nil, e
	}
	if err = os.MkdirAll(filepath.Dir(output), 0755); err != nil {
		return nil, err
	}
	stage, err := os.MkdirTemp(filepath.Dir(output), "probe-capture-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(stage)
	evidence := filepath.Join(stage, "evidence")
	if err = os.Mkdir(evidence, 0755); err != nil {
		return nil, err
	}
	for name, data := range files {
		if err = os.WriteFile(filepath.Join(evidence, name), data, 0644); err != nil {
			return nil, err
		}
	}
	if err = os.Rename(evidence, output); err != nil {
		return nil, err
	}
	return classification, nil
}
