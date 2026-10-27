package wowapi

import (
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"

	"wow-teal-types/internal/sourcelock"
	"wow-teal-types/internal/wowapi/literal"
)

func (p Project) observationPath(name string) (string, error) {
	if !evidenceSafeRelative(name) || !strings.HasPrefix(name, "references/wow-api/native-observations/") {
		return "", fmt.Errorf("invalid native observation evidence path")
	}
	return filepath.Join(p.Root, filepath.FromSlash(name)), nil
}
func exactEvidenceKeys[T any](m map[string]T, names []string) bool {
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
func parseCaptureJSON[T any](data []byte) (result T, err error) {
	raw, err := readCaptureData(data, false)
	if err != nil {
		return result, err
	}
	return decodeCaptureObject[T](raw)
}
func readCaptureJSON[T any](path string) (result T, err error) {
	data, err := readCaptureFile(path)
	if err != nil {
		return result, err
	}
	return parseCaptureJSON[T](data)
}
func (p Project) loadCapture(lock ObservationLock) (observed NativeObservations, err error) {
	defer captureError(&err)
	if lock.SchemaVersion != 1 || !captureHash.MatchString(lock.EvidenceSHA256) || !captureHash.MatchString(lock.AcquisitionSHA256) || !captureHash.MatchString(lock.SourceInputsSHA256) {
		return observed, fmt.Errorf("unsupported native observation lock")
	}
	bundle, e := p.observationPath(lock.Bundle)
	if e != nil {
		return observed, e
	}
	acqPath, e := p.observationPath(lock.Acquisition)
	if e != nil {
		return observed, e
	}
	indexBytes, e := readEvidenceFile(filepath.Join(bundle, "evidence.json"), 256<<10)
	if e != nil {
		return observed, e
	}
	acquisition, e := readEvidenceFile(acqPath, 256<<10)
	if e != nil {
		return observed, e
	}
	if sha256hex(indexBytes) != lock.EvidenceSHA256 || sha256hex(acquisition) != lock.AcquisitionSHA256 {
		return observed, fmt.Errorf("native observation acquisition identity changed")
	}
	index, e := parseCaptureJSON[CaptureEvidenceIndex](indexBytes)
	if e != nil {
		return observed, e
	}
	names := []string{"capture.lua", "report.json", "probe-build.json", "installed-receipt.json", "classification.json"}
	if index.SchemaVersion != 1 || !exactEvidenceKeys(index.Files, names) {
		return observed, fmt.Errorf("native observation bundle inventory changed")
	}
	files := map[string][]byte{}
	for _, name := range names {
		data, e := readEvidenceFile(filepath.Join(bundle, name), 256<<10)
		if e != nil {
			return observed, e
		}
		expected := index.Files[name]
		if len(data) != expected.Size || sha256hex(data) != expected.SHA256 {
			return observed, fmt.Errorf("native observation bundle bytes changed: %s", name)
		}
		files[name] = data
	}
	rawReport, e := readCaptureData(files["report.json"], false)
	if e != nil {
		return observed, e
	}
	report, e := decodeCaptureReport(rawReport)
	if e != nil {
		return observed, e
	}
	rawManifest, e := readCaptureData(files["probe-build.json"], false)
	if e != nil {
		return observed, e
	}
	manifest, e := decodeProbeManifest(rawManifest)
	if e != nil {
		return observed, e
	}
	rawReceipt, e := readCaptureData(files["installed-receipt.json"], false)
	if e != nil {
		return observed, e
	}
	receipt, e := decodeProbeReceipt(rawReceipt)
	if e != nil {
		return observed, e
	}
	classification, e := validateCapture(report, manifest, receipt, files["probe-build.json"])
	if e != nil {
		return observed, e
	}
	rawCapture, e := readCaptureData(files["capture.lua"], true)
	if e != nil {
		return observed, e
	}
	rawClassification, e := readCaptureData(files["classification.json"], false)
	if e != nil {
		return observed, e
	}
	if !literal.Record(rawCapture).Equal(literal.Record(rawReport)) || !captureMatches(rawClassification, classification) {
		return observed, fmt.Errorf("native observation representations disagree")
	}
	acq, e := parseCaptureJSON[captureAcquisition](acquisition)
	if e != nil {
		return observed, e
	}
	build := report.ClientVersion + "." + report.ClientBuild
	if build != lock.Build || report.Mode != "public_invocations" || acq.BundleIndexSHA256 != lock.EvidenceSHA256 || acq.ClientBuild != build {
		return observed, fmt.Errorf("native observation build or acquisition context changed")
	}
	sources, e := sourcelock.ReadSources(filepath.Join(p.Root, "scripts/wow_api/sources.json"))
	if e != nil {
		return observed, e
	}
	hierarchy, e := sourcelock.ReadCrosscheck(filepath.Join(p.Root, "scripts/wow_api/native_crosscheck_sources.json"))
	if e != nil {
		return observed, e
	}
	version, e := readEvidenceFile(p.InputPath("blizzard_version", "version.txt"), 1024)
	if e != nil {
		return observed, e
	}
	identity := struct {
		Catalogue sourcelock.Sources    `json:"catalogue"`
		Hierarchy sourcelock.Crosscheck `json:"hierarchy"`
	}{sources, hierarchy}
	observed = NativeObservations{Sources: lock, Report: report, Plan: manifest.Plan, Applicable: build == strings.TrimSpace(string(version)) && evDigest(identity) == lock.SourceInputsSHA256, Methods: map[string]CaptureInvocation{}, Hierarchy: map[string]CaptureObservation{}}
	for _, row := range report.Invocations {
		observed.Methods[row.CaseID] = row
	}
	for _, row := range report.Hierarchy {
		observed.Hierarchy[row.CaseID] = row
	}
	return observed, nil
}
func (p Project) LoadObservations() (observed NativeObservations, err error) {
	defer captureError(&err)
	lock, e := readCaptureJSON[ObservationLock](filepath.Join(p.Root, "scripts/wow_api/native_observations.json"))
	if e != nil {
		return observed, e
	}
	observed, e = p.loadCapture(lock)
	if e != nil {
		return observed, e
	}
	sources, e := readCaptureJSON[BehaviorReviewSources](filepath.Join(p.Root, "scripts/wow_api/native_behavior_reviews.json"))
	if e != nil {
		return observed, e
	}
	if sources.SchemaVersion != 1 {
		return observed, fmt.Errorf("unsupported native behavior review")
	}
	rule := sources.Review
	expected := BehaviorReviewRule{Symbol: "DurationTextBindingObject.UpdateFontString", RecordedCheck: "font_text_matches", CorrectedCheck: "font_matches_native_projection", Disposition: "verified-through-native-font-projection", RawSample: "baseline.expected", BindingSample: "baseline.font", ProjectionSample: "projection.font", Reason: rule.Reason}
	if rule != expected || rule.Reason == "" {
		return observed, fmt.Errorf("unsupported native behavior review rule")
	}
	capture, e := p.loadCapture(sources.Observation)
	if e != nil {
		return observed, e
	}
	report := capture.Report
	method := capture.Methods[rule.Symbol]
	states := func(rows []CaptureState) map[string]string {
		result := map[string]string{}
		for _, row := range rows {
			result[row.CaseID] = row.State
		}
		return result
	}
	diagnostics, controls := states(report.BindingDiagnostics), states(report.BindingExperimentChecks)
	cleanup := map[string]string{}
	for _, row := range report.Cleanup {
		cleanup[row.CaseID] = row.State
	}
	samples := map[string]BindingTextSample{}
	for _, row := range report.BindingTextSamples {
		samples[row.CaseID] = row
	}
	bad := report.ProbeVersion != "1.4.0" || method.Check != rule.RecordedCheck || method.State != "property_mismatch" || !method.Invoked || method.BehaviorVerified || method.ReturnCount != 0 || diagnostics["baseline.expected_stable"] != "matched" || controls["projection.set"] != "called_void" || !exactEvidenceKeys(cleanup, []string{"cleanup.invocation.binding.1", "cleanup.invocation.binding.2", "cleanup.invocation.signals"})
	for _, state := range cleanup {
		bad = bad || state != "complete"
	}
	values := [][]byte{}
	for _, name := range []string{rule.RawSample, rule.BindingSample, rule.ProjectionSample} {
		sample := samples[name]
		bad = bad || sample.State != "public"
		value, e := hex.DecodeString(sample.Hex)
		if e != nil {
			return observed, e
		}
		values = append(values, value)
	}
	if bad {
		return observed, fmt.Errorf("native behavior review lacks its required observations")
	}
	raw, binding, projection := string(values[0]), string(values[1]), string(values[2])
	if raw == "" || binding == "" || raw == binding || binding != projection {
		return observed, fmt.Errorf("native behavior review does not prove native font projection")
	}
	review := BehaviorReview{Symbol: rule.Symbol, Check: rule.CorrectedCheck, BehaviorVerified: true, Disposition: rule.Disposition, Reason: rule.Reason, Build: capture.Sources.Build, Observation: capture.Sources, Scope: "owned public fixture postcondition: captured binding text equals captured native SetText projection; no causality or secrecy claim; original raw mismatch preserved"}
	observed.BehaviorReviewSources = sources
	observed.BehaviorReviews = map[string]BehaviorReview{}
	if capture.Applicable {
		observed.BehaviorReviews[rule.Symbol] = review
	}
	return observed, nil
}
