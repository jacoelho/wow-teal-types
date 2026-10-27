package wowapi

import (
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
)

func (p Project) observationPath(name string) (string, error) {
	if !evidenceSafeRelative(name) || !strings.HasPrefix(name, "references/wow-api/native-observations/") {
		return "", fmt.Errorf("invalid native observation evidence path")
	}
	return filepath.Join(p.Root, filepath.FromSlash(name)), nil
}
func exactEvidenceKeys(m M, names []string) bool {
	if m == nil || len(m) != len(names) {
		return false
	}
	for _, name := range names {
		if _, ok := m[name]; !ok {
			return false
		}
	}
	return true
}
func (p Project) loadCapture(lock M) (M, error) {
	if !exactEvidenceKeys(lock, []string{"schema_version", "bundle", "evidence_sha256", "acquisition", "acquisition_sha256", "build", "source_inputs_sha256"}) || asInt(lock["schema_version"]) != 1 {
		return nil, fmt.Errorf("unsupported native observation lock")
	}
	bundle, e := p.observationPath(asString(lock["bundle"]))
	if e != nil {
		return nil, e
	}
	acqPath, e := p.observationPath(asString(lock["acquisition"]))
	if e != nil {
		return nil, e
	}
	indexBytes, e := readEvidenceFile(filepath.Join(bundle, "evidence.json"), 256<<10)
	if e != nil {
		return nil, e
	}
	acquisition, e := readEvidenceFile(acqPath, 256<<10)
	if e != nil {
		return nil, e
	}
	if sha256hex(indexBytes) != lock["evidence_sha256"] || sha256hex(acquisition) != lock["acquisition_sha256"] {
		return nil, fmt.Errorf("native observation acquisition identity changed")
	}
	index, e := readCaptureData(indexBytes, false)
	if e != nil {
		return nil, e
	}
	names := []string{"capture.lua", "report.json", "probe-build.json", "installed-receipt.json", "classification.json"}
	inventory := asMap(index["files"])
	if asInt(index["schema_version"]) != 1 || !exactEvidenceKeys(inventory, names) {
		return nil, fmt.Errorf("native observation bundle inventory changed")
	}
	files := map[string][]byte{}
	parsed := map[string]M{}
	for _, name := range names {
		data, e := readEvidenceFile(filepath.Join(bundle, name), 256<<10)
		if e != nil {
			return nil, e
		}
		expected := asMap(inventory[name])
		if len(data) != asInt(expected["size"]) || sha256hex(data) != expected["sha256"] {
			return nil, fmt.Errorf("native observation bundle bytes changed: %s", name)
		}
		files[name] = data
		parsed[name], e = readCaptureData(data, name == "capture.lua")
		if e != nil {
			return nil, e
		}
	}
	report, manifest, receipt := parsed["report.json"], parsed["probe-build.json"], parsed["installed-receipt.json"]
	classification, e := validateCapture(report, manifest, receipt, files["probe-build.json"])
	if e != nil {
		return nil, e
	}
	if !evEqual(parsed["capture.lua"], report) || !evEqual(parsed["classification.json"], classification) {
		return nil, fmt.Errorf("native observation representations disagree")
	}
	acq, e := readCaptureData(acquisition, false)
	if e != nil {
		return nil, e
	}
	build := asString(report["clientVersion"]) + "." + asString(report["clientBuild"])
	if build != lock["build"] || report["mode"] != "public_invocations" || acq["bundle_index_sha256"] != lock["evidence_sha256"] || acq["client_build"] != build {
		return nil, fmt.Errorf("native observation build or acquisition context changed")
	}
	sources, e := p.Lock("sources.json")
	if e != nil {
		return nil, e
	}
	hierarchy, e := p.Lock("native_crosscheck_sources.json")
	if e != nil {
		return nil, e
	}
	version, e := readEvidenceFile(p.InputPath("blizzard_version", "version.txt"), 1024)
	if e != nil {
		return nil, e
	}
	applicable := build == strings.TrimSpace(string(version)) && evDigest(M{"catalogue": sources, "hierarchy": hierarchy}) == lock["source_inputs_sha256"]
	methods, capabilities := M{}, M{}
	for _, v := range asList(report["invocations"]) {
		row := asMap(v)
		methods[asString(row["caseID"])] = row
	}
	for _, v := range asList(report["hierarchy"]) {
		row := asMap(v)
		capabilities[asString(row["caseID"])] = row
	}
	return M{"sources": lock, "report": report, "plan": manifest["plan"], "applicable": applicable, "methods": methods, "hierarchy": capabilities}, nil
}
func (p Project) LoadObservations() (M, error) {
	lock, e := p.Lock("native_observations.json")
	if e != nil {
		return nil, e
	}
	observed, e := p.loadCapture(lock)
	if e != nil {
		return nil, e
	}
	sources, e := p.Lock("native_behavior_reviews.json")
	if e != nil {
		return nil, e
	}
	if !exactEvidenceKeys(sources, []string{"schema_version", "observation", "review"}) || asInt(sources["schema_version"]) != 1 {
		return nil, fmt.Errorf("unsupported native behavior review")
	}
	rule := asMap(sources["review"])
	expected := M{"symbol": "DurationTextBindingObject.UpdateFontString", "recorded_check": "font_text_matches", "corrected_check": "font_matches_native_projection", "disposition": "verified-through-native-font-projection", "raw_sample": "baseline.expected", "binding_sample": "baseline.font", "projection_sample": "projection.font"}
	if len(rule) != len(expected)+1 || asString(rule["reason"]) == "" {
		return nil, fmt.Errorf("unsupported native behavior review rule")
	}
	for key, value := range expected {
		if !evEqual(rule[key], value) {
			return nil, fmt.Errorf("unsupported native behavior review rule")
		}
	}
	capture, e := p.loadCapture(asMap(sources["observation"]))
	if e != nil {
		return nil, e
	}
	report := asMap(capture["report"])
	method := asMap(asMap(capture["methods"])[asString(rule["symbol"])])
	states := func(rows any) M {
		r := M{}
		for _, v := range asList(rows) {
			m := asMap(v)
			r[asString(m["caseID"])] = m["state"]
		}
		return r
	}
	diagnostics, controls, cleanup := states(report["bindingDiagnostics"]), states(report["bindingExperimentChecks"]), states(report["cleanup"])
	samples := M{}
	for _, v := range asList(report["bindingTextSamples"]) {
		m := asMap(v)
		samples[asString(m["caseID"])] = m
	}
	bad := report["probeVersion"] != "1.4.0" || method["check"] != rule["recorded_check"] || method["state"] != "property_mismatch" || method["invoked"] != true || method["behaviorVerified"] != false || asInt(method["returnCount"]) != 0 || diagnostics["baseline.expected_stable"] != "matched" || controls["projection.set"] != "called_void" || !exactEvidenceKeys(cleanup, []string{"cleanup.invocation.binding.1", "cleanup.invocation.binding.2", "cleanup.invocation.signals"})
	for _, state := range cleanup {
		bad = bad || state != "complete"
	}
	required := []string{asString(rule["raw_sample"]), asString(rule["binding_sample"]), asString(rule["projection_sample"])}
	values := [][]byte{}
	for _, name := range required {
		sample := asMap(samples[name])
		if sample["state"] != "public" {
			bad = true
		}
		b, e := hex.DecodeString(asString(sample["hex"]))
		if e != nil {
			return nil, e
		}
		values = append(values, b)
	}
	if bad {
		return nil, fmt.Errorf("native behavior review lacks its required observations")
	}
	raw, binding, projection := string(values[0]), string(values[1]), string(values[2])
	if raw == "" || binding == "" || raw == binding || binding != projection {
		return nil, fmt.Errorf("native behavior review does not prove native font projection")
	}
	review := M{"symbol": rule["symbol"], "check": rule["corrected_check"], "behavior_verified": true, "disposition": rule["disposition"], "reason": rule["reason"], "build": asMap(capture["sources"])["build"], "observation": capture["sources"], "scope": "owned public fixture postcondition: captured binding text equals captured native SetText projection; no causality or secrecy claim; original raw mismatch preserved"}
	observed["behavior_review_sources"] = sources
	observed["behavior_reviews"] = M{}
	if truth(capture["applicable"]) {
		observed["behavior_reviews"] = M{asString(rule["symbol"]): review}
	}
	return observed, nil
}
