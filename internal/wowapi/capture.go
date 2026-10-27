package wowapi

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"unicode/utf8"
	"wow-teal-types/internal/wowapi/literal"
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
func capExact(value literal.Object, fields map[string]bool, label string) {
	if value == nil || len(value) != len(fields) {
		reject(fmt.Sprintf("unexpected %s fields", label))
	}
	for key := range value {
		if !fields[key] {
			reject(fmt.Sprintf("unexpected %s fields", label))
		}
	}
}
func capSet(values []string, label string) map[string]bool {
	result := map[string]bool{}
	for _, s := range values {
		if result[s] {
			reject(fmt.Sprintf("duplicate %s", label))
		}
		result[s] = true
	}
	return result
}
func capRequired(values []string, name string, expected map[string]bool) {
	if !capEqualSet(capSet(values, "binding requirements: "+name), expected) {
		reject(fmt.Sprintf("invalid binding requirements: %s", name))
	}
}
func capBounded(v literal.Value, depth int, budget *int) {
	*budget--
	if *budget < 0 || depth > 64 {
		reject("capture exceeds node/depth limit")
	}
	switch v.Kind() {
	case literal.ObjectKind:
		fields, _ := v.Object()
		for key, value := range fields {
			if key == "$ref" || key == "$expression" {
				reject("nonliteral capture key")
			}
			capBounded(value, depth+1, budget)
		}
	case literal.Array:
		rows, _ := v.Array()
		for _, value := range rows {
			capBounded(value, depth+1, budget)
		}
	case literal.String:
		text, _ := v.Text()
		if utf8.RuneCountInString(text) > 512 {
			reject("capture string exceeds 512 characters")
		}
	case literal.Null, literal.Boolean, literal.Integer, literal.FloatKind:
	default:
		reject("unsupported capture value")
	}
}
func captureLiteral(p *luaParser, depth int) literal.Value {
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
			if n, ok := value.Integer(); ok {
				if sign == "-" {
					if strings.HasPrefix(n, "-") {
						n = strings.TrimPrefix(n, "-")
					} else {
						n = "-" + n
					}
				}
				result, e := literal.BigInteger(n)
				if e != nil {
					reject(fmt.Sprintf("invalid signed capture literal: %s", e))
				}
				return result
			}
			n, ok := value.Number()
			if !ok {
				reject("invalid signed capture literal")
			}
			if sign == "-" {
				n = -n
			}
			result, e := literal.Number(n)
			if e != nil {
				reject(fmt.Sprintf("invalid signed capture literal: %s", e))
			}
			return result
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
	entries := []literal.Value{}
	fields := literal.Object{}
	for p.peek() != "}" {
		key := ""
		hasKey := false
		if p.peek() == "[" {
			p.take()
			value := captureLiteral(p, depth+1)
			var ok bool
			key, ok = value.Text()
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
		return literal.Record(fields)
	}
	return literal.List(entries)
}
func readCaptureData(data []byte, savedVariables bool) (result literal.Object, err error) {
	defer captureError(&err)
	if len(data) > captureMaxBytes {
		reject("capture exceeds byte limit")
	}
	if !utf8.Valid(data) {
		reject("capture must be UTF-8")
	}
	var value literal.Value
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
		var e error
		value, e = literal.ParseJSON(data)
		if e != nil {
			reject(fmt.Sprintf("invalid capture JSON: %s", e))
		}
	}
	budget := captureMaxNodes
	capBounded(value, 0, &budget)
	result, ok := value.Object()
	if !ok {
		reject("capture must be an object")
	}
	return result, nil
}
func capObservations(rows []CaptureObservation, allowed, valueCases map[string]bool) map[string]CaptureObservation {
	result := map[string]CaptureObservation{}
	for _, row := range rows {
		name, state := row.CaseID, row.State
		if _, exists := result[name]; !allowed[name] || exists || !captureState.MatchString(state) {
			reject("duplicate, unknown or invalid observation")
		}
		if row.Value != nil && (!valueCases[name] || state != "available") {
			reject("unexpected observation payload")
		}
		if valueCases[name] && state == "available" && row.Value == nil {
			reject("missing available query value")
		}
		result[name] = row
	}
	return result
}
func capStates(rows []CaptureState, allowed, setters, reads map[string]bool, label string) map[string]string {
	found := map[string]string{}
	for _, row := range rows {
		name, state := row.CaseID, row.State
		if !allowed[name] || found[name] != "" {
			reject(fmt.Sprintf("duplicate or unknown binding %s", label))
		}
		states := wordSet("matched mismatch not_comparable not_run")
		if setters[name] {
			states = capUnion(wordSet("called_void unexpected_tuple"), diagnosticFailures)
		} else if reads[name] {
			states = capUnion(wordSet("public_empty public_nonempty nil secret wrong_type oversized unexpected_tuple"), diagnosticFailures)
		}
		if !states[state] {
			reject(fmt.Sprintf("invalid binding %s state: %s", label, name))
		}
		found[name] = state
	}
	return found
}
func capDiagnostics(report CaptureReport, plan ProbePlan, mode string) map[string]string {
	required := capUnion(diagnosticSetters, diagnosticReads, diagnosticComparisons)
	capRequired(plan.BindingDiagnosticCases, "bindingDiagnosticCases", required)
	found := capStates(report.BindingDiagnostics, required, diagnosticSetters, diagnosticReads, "diagnostic")
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
					reject(fmt.Sprintf("unavailable diagnostic inputs cannot establish equality: %s", name))
				}
			}
		}
	}
	return found
}
func capExperiments(report CaptureReport, plan ProbePlan, mode string, diagnostics map[string]string) (map[string]string, map[string]BindingTextSample) {
	required := capUnion(experimentSetters, experimentComparisons)
	capRequired(plan.BindingTextSampleCases, "bindingTextSampleCases", bindingSamples)
	capRequired(plan.BindingExperimentCases, "bindingExperimentCases", required)
	found := map[string]string{}
	public := map[string][]byte{}
	samples := map[string]BindingTextSample{}
	failures := capUnion(diagnosticFailures, wordSet("nil secret wrong_type oversized unexpected_tuple"))
	for _, row := range report.BindingTextSamples {
		state := row.State

		name := row.CaseID
		if !bindingSamples[name] || found[name] != "" {
			reject("duplicate or unknown binding text sample")
		}
		if state == "public" {
			text := row.Hex
			if !captureHex.MatchString(text) {
				reject("invalid or oversized public fixture bytes")
			}
			value, _ := hex.DecodeString(text)
			if row.ByteLength != len(value) || row.FirstNul != bytes.IndexByte(value, 0)+1 {
				reject("inconsistent public fixture byte metadata")
			}
			public[name] = value
		} else if !failures[state] {
			reject("invalid binding text sample state")
		}
		found[name] = state
		samples[name] = row
	}
	checks := capStates(report.BindingExperimentChecks, required, experimentSetters, nil, "experiment check")
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

func validateCapture(report CaptureReport, manifest ProbeBuildManifest, receipt ProbeReceipt, manifestBytes []byte) (classification CaptureResult, err error) {
	defer captureError(&err)
	plan := manifest.Plan
	version := plan.Version
	if !wordSet("1.0.0 1.1.0 1.2.0 1.3.0 1.4.0 1.5.0")[version] || manifest.Probe != version {
		reject("unsupported probe version")
	}
	fields := captureReportFields(version)
	if receipt.BuildManifestSHA256 != sha256hex(manifestBytes) || !maps.Equal(receipt.Files, manifest.Files) {
		reject("installed probe receipt identity mismatch")
	}
	runtimeFiles := manifest.Files
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
		if !captureHash.MatchString(value) {
			reject("invalid installed probe file identity")
		}
	}
	for key, pair := range map[string][2]string{"probeVersion": {report.ProbeVersion, version}, "expectedBuild": {report.ExpectedBuild, plan.ExpectedBuild}, "sourceFingerprint": {report.SourceFingerprint, plan.SourceFingerprint}, "declarationFingerprint": {report.DeclarationFingerprint, plan.DeclarationFingerprint}} {
		if pair[0] != pair[1] {
			reject(fmt.Sprintf("capture identity mismatch: %s", key))
		}
	}
	if report.Schema != 1 {
		reject("unsupported capture schema")
	}
	if report.ClientVersion+"."+report.ClientBuild != plan.ExpectedBuild {
		reject("client build mismatch")
	}
	for key, pair := range map[string][2]int{"clientInterface": {report.ClientInterface, plan.ExpectedInterface}, "projectID": {report.ProjectID, plan.ExpectedProjectID}} {
		if pair[0] != pair[1] {
			reject(fmt.Sprintf("client identity mismatch: %s", key))
		}
	}
	if report.Context != "out_of_combat_public_owned_objects" || report.Limits == nil || !maps.Equal(report.Limits, plan.Limits) {
		reject("capture context/limits mismatch")
	}

	if report.VerifiesSignatures != false || report.VerifiesSecrecy != false {
		reject("capture cannot certify signatures or secrecy")
	}
	expected := map[string]ProbeFamily{}
	methods := map[string]bool{}
	for _, row := range plan.Families {
		name := row.Name
		if _, exists := expected[name]; exists {
			reject("invalid capture method requirements")
		}
		expected[name] = row
		for _, method := range row.Methods {
			methods[row.Canonical+"."+method] = true
		}
	}
	if len(methods) != plan.ExpectedMethods || report.ExpectedMethods != len(methods) {
		reject("invalid capture method requirements")
	}
	mode := report.Mode
	if version == "1.0.0" {
		mode = "presence"
	}
	if mode != "presence" && mode != "public_invocations" {
		reject("unsupported capture mode")
	}
	diagnostics := map[string]string{}
	checks := map[string]string{}
	samples := map[string]BindingTextSample{}
	if fields["bindingDiagnostics"] {
		diagnostics = capDiagnostics(report, plan, mode)
	}
	if fields["bindingTextSamples"] {
		checks, samples = capExperiments(report, plan, mode, diagnostics)
	}
	observed := map[string]CaptureObservation{}
	families := map[string]bool{}
	present := 0
	factoriesAvailable := true
	for _, family := range report.Families {
		name := family.Family
		requirement, ok := expected[name]
		if !ok || families[name] {
			reject("duplicate or unknown capture family")
		}
		families[name] = true
		factory := family.Factory
		if !captureState.MatchString(factory) {
			reject("invalid constructor state")
		}
		factoriesAvailable = factoriesAvailable && factory == "available"
		ids := map[string]bool{}
		canonical := requirement.Canonical
		for _, method := range requirement.Methods {
			ids[canonical+"."+method] = true
		}
		rows := capObservations(family.Methods, ids, nil)
		if len(rows) != len(ids) {
			reject("missing capture method case")
		}
		for key, row := range rows {
			observed[key] = row
			if row.State == "present" {
				present++
			}
		}
		prefix := canonical + ".query."
		queries := map[string]bool{prefix + "GetObjectType": true}
		for _, kind := range captureQueryTypes {
			queries[prefix+"IsObjectType."+kind] = true
		}
		queryRows := capObservations(family.Queries, queries, queries)
		if factory == "available" && len(queryRows) != len(queries) {
			reject("missing capture query case")
		}
		for name, row := range queryRows {
			if row.Value != nil && row.Value.isBoolean == strings.HasSuffix(name, "GetObjectType") {
				reject("invalid native query result type")
			}
		}
	}
	for key, count := range map[string]int{"observedMethods": len(observed), "presentMethods": present} {
		if map[string]int{"observedMethods": report.ObservedMethods, "presentMethods": report.PresentMethods, "invokedMethods": report.InvokedMethods, "verifiedMethods": report.VerifiedMethods}[key] != count {
			reject(fmt.Sprintf("inconsistent capture summary: %s", key))
		}
	}
	cleanupIDs := wordSet("cleanup.DurationTextBinding.Disable cleanup.invocation.binding.1 cleanup.invocation.binding.2 cleanup.invocation.signals")
	cleanup := capObservations(report.Cleanup, cleanupIDs, nil)
	clean := true
	for _, row := range cleanup {
		clean = clean && row.State == "complete"
	}
	presenceComplete := mode == "presence" && len(families) == len(expected) && factoriesAvailable && present == len(methods) && clean && cleanup["cleanup.DurationTextBinding.Disable"].CaseID != ""
	if report.CapabilityCaptureComplete != presenceComplete {
		reject("inconsistent presence completion")
	}
	if version == "1.0.0" {
		outcome := "incomplete"
		if presenceComplete {
			outcome = "capabilities_collected"
		}
		if report.Outcome != outcome {
			reject("inconsistent presence outcome")
		}
		return CaptureResult{Kind: "presence", PresentMethods: present, MethodCases: sortedKeys(methods), UnobservedHierarchy: true}, nil
	}
	if report.FixtureSHA256 != plan.FixtureSHA256 {
		reject("fixture identity mismatch")
	}
	hierarchy := map[string]bool{}
	for _, row := range plan.Hierarchy {
		for _, method := range row.Methods {
			hierarchy[row.Canonical+".parent."+row.Parent+"."+method] = true
		}
	}
	hierarchyRows := capObservations(report.Hierarchy, hierarchy, nil)
	if presenceComplete && len(hierarchyRows) != len(hierarchy) {
		reject("missing hierarchy cases")
	}
	invoked, verified := 0, 0
	seen := map[string]bool{}
	invocationStates := wordSet("not_run absent inaccessible inaccessible_table lookup_error invalid_object not_function call_error passed property_mismatch projection_unavailable secret_result inaccessible_result unsupported_tuple fixture_unavailable fixture_setup_failed")
	validLabels := wordSet("nil boolean number string table userdata function thread secret inaccessible")
	successfulStates := wordSet("passed property_mismatch projection_unavailable secret_result inaccessible_result unsupported_tuple fixture_unavailable")
	for _, row := range report.Invocations {
		name, state := row.CaseID, row.State
		if !methods[name] || seen[name] {
			reject("duplicate or unknown invocation")
		}
		seen[name] = true
		if !invocationStates[state] {
			reject("invalid invocation state")
		}
		if !captureCheck.MatchString(row.Check) {
			reject("invalid invocation check identity")
		}
		called, behavior := row.Invoked, row.BehaviorVerified
		count := row.ReturnCount
		if count < 0 || count > 8 {
			reject("invalid invocation types/count")
		}
		labels := row.Returns
		if len(labels) != count {
			reject("invalid invocation tuple labels")
		}
		opaque := map[string]bool{}
		for _, label := range labels {
			s := label
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
			if row.Check != "font_matches_native_projection" {
				reject("incorrect native projection oracle")
			}
			if passed {
				valid := count == 0 && diagnostics["baseline.expected_stable"] == "matched" && diagnostics["sentinel.set"] == "called_void" && diagnostics["sentinel.matches"] == "matched" && checks["projection.set"] == "called_void"
				for _, sample := range []string{"baseline.expected", "baseline.font", "projection.font"} {
					valid = valid && samples[sample].State == "public"
				}
				valid = valid && samples["baseline.font"].Hex == samples["projection.font"].Hex && samples["projection.font"].Hex != hex.EncodeToString([]byte("PROBE_SENTINEL"))
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
		if map[string]int{"observedMethods": report.ObservedMethods, "presentMethods": report.PresentMethods, "invokedMethods": report.InvokedMethods, "verifiedMethods": report.VerifiedMethods}[key] != count {
			reject("inconsistent invocation summary")
		}
	}
	requiredCleanup := wordSet("cleanup.invocation.binding.1 cleanup.invocation.binding.2 cleanup.invocation.signals")
	allCleanup := true
	for name := range requiredCleanup {
		allCleanup = allCleanup && cleanup[name].CaseID != ""
	}
	complete := mode == "public_invocations" && len(seen) == len(methods) && verified == len(methods) && clean && allCleanup
	if complete && len(hierarchyRows) != len(hierarchy) {
		reject("missing hierarchy cases")
	}
	if report.InvocationCaptureComplete != complete {
		reject("inconsistent invocation completion")
	}
	if mode == "public_invocations" {
		outcome := "public_invocations_incomplete"
		if complete {
			outcome = "public_invocations_verified"
		}
		if len(seen) != len(methods) || len(observed) > 0 || len(families) > 0 || report.Outcome != outcome {
			reject("invalid invocation coverage/outcome")
		}
	} else {
		outcome := "incomplete"
		if presenceComplete {
			outcome = "capabilities_collected"
		}
		if len(seen) > 0 || invoked != 0 || verified != 0 || report.Outcome != outcome {
			reject("presence cannot contain invocation evidence")
		}
	}
	cases := sortedKeys(hierarchyRows)
	return CaptureResult{Kind: mode, PresentMethods: present, SuccessfulInvocations: invoked, ProbeReportedPropertyPasses: verified, MethodCases: sortedKeys(methods), HierarchyCases: &cases}, nil
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
func IngestCapture(capture, manifestPath, receiptPath, output string) (CaptureResult, error) {
	raw, err := readCaptureFile(capture)
	if err != nil {
		return CaptureResult{}, err
	}
	manifestBytes, err := readCaptureFile(manifestPath)
	if err != nil {
		return CaptureResult{}, err
	}
	receiptBytes, err := readCaptureFile(receiptPath)
	if err != nil {
		return CaptureResult{}, err
	}
	rawReport, err := readCaptureData(raw, filepath.Ext(capture) == ".lua")
	if err != nil {
		return CaptureResult{}, err
	}
	rawManifest, err := readCaptureData(manifestBytes, false)
	if err != nil {
		return CaptureResult{}, err
	}
	rawReceipt, err := readCaptureData(receiptBytes, false)
	if err != nil {
		return CaptureResult{}, err
	}
	report, err := decodeCaptureReport(rawReport)
	if err != nil {
		return CaptureResult{}, err
	}
	manifest, err := decodeProbeManifest(rawManifest)
	if err != nil {
		return CaptureResult{}, err
	}
	receipt, err := decodeProbeReceipt(rawReceipt)
	if err != nil {
		return CaptureResult{}, err
	}
	classification, err := validateCapture(report, manifest, receipt, manifestBytes)
	if err != nil {
		return CaptureResult{}, err
	}
	files := map[string][]byte{"capture" + filepath.Ext(capture): raw, "probe-build.json": manifestBytes, "installed-receipt.json": receiptBytes}
	files["report.json"], err = jsonPretty(rawReport)
	if err != nil {
		return CaptureResult{}, err
	}
	files["classification.json"], err = jsonPretty(classification)
	if err != nil {
		return CaptureResult{}, err
	}
	index := map[string]CaptureEvidenceFile{}
	for name, data := range files {
		index[name] = CaptureEvidenceFile{SHA256: sha256hex(data), Size: len(data)}
	}
	files["evidence.json"], err = jsonPretty(CaptureEvidenceIndex{Files: index, SchemaVersion: 1, Scope: "validated caller-supplied data; acquisition receipt must be independently established"})
	if err != nil {
		return CaptureResult{}, err
	}
	if info, e := os.Lstat(output); e == nil {
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return CaptureResult{}, fmt.Errorf("existing capture evidence differs; never overwrite")
		}
		entries, e := os.ReadDir(output)
		if e != nil {
			return CaptureResult{}, e
		}
		if len(entries) != len(files) {
			return CaptureResult{}, fmt.Errorf("existing capture evidence differs; never overwrite")
		}
		for _, entry := range entries {
			expected, exists := files[entry.Name()]
			if !exists {
				return CaptureResult{}, fmt.Errorf("existing capture evidence differs; never overwrite")
			}
			data, e := readCaptureFile(filepath.Join(output, entry.Name()))
			if e != nil {
				return CaptureResult{}, e
			}
			if !bytes.Equal(data, expected) {
				return CaptureResult{}, fmt.Errorf("existing capture evidence differs; never overwrite")
			}
		}
		return classification, nil
	} else if !os.IsNotExist(e) {
		return CaptureResult{}, e
	}
	if err = os.MkdirAll(filepath.Dir(output), 0755); err != nil {
		return CaptureResult{}, err
	}
	stage, err := os.MkdirTemp(filepath.Dir(output), "probe-capture-")
	if err != nil {
		return CaptureResult{}, err
	}
	defer os.RemoveAll(stage)
	evidence := filepath.Join(stage, "evidence")
	if err = os.Mkdir(evidence, 0755); err != nil {
		return CaptureResult{}, err
	}
	for name, data := range files {
		if err = os.WriteFile(filepath.Join(evidence, name), data, 0644); err != nil {
			return CaptureResult{}, err
		}
	}
	if err = os.Rename(evidence, output); err != nil {
		return CaptureResult{}, err
	}
	return classification, nil
}
