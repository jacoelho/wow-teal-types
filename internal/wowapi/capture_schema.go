package wowapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"wow-teal-types/internal/wowapi/literal"
)

type ProbePlan struct {
	BindingDiagnosticCases  []string                `json:"bindingDiagnosticCases"`
	BindingExperimentCases  []string                `json:"bindingExperimentCases"`
	BindingTextSampleCases  []string                `json:"bindingTextSampleCases"`
	CandidateEvidence       *ProbeCandidateEvidence `json:"candidateEvidence"`
	DeclarationFingerprint  string                  `json:"declarationFingerprint"`
	ExpectedBuild           string                  `json:"expectedBuild"`
	ExpectedInterface       int                     `json:"expectedInterface"`
	ExpectedMethods         int                     `json:"expectedMethods"`
	ExpectedProjectID       int                     `json:"expectedProjectID"`
	Families                []ProbeFamily           `json:"families"`
	FixtureSHA256           string                  `json:"fixtureSHA256"`
	Hierarchy               []ProbeHierarchy        `json:"hierarchy"`
	Limits                  map[string]int          `json:"limits"`
	NativeCandidateEvidence *ProbeNativeEvidence    `json:"nativeCandidateEvidence"`
	Representatives         []ProbeRepresentative   `json:"representatives"`
	SourceFingerprint       string                  `json:"sourceFingerprint"`
	Version                 string                  `json:"version"`
}

type ProbeCandidateEvidence struct {
	Build       string   `json:"build"`
	Commit      string   `json:"commit"`
	GitBlobSHA1 string   `json:"git_blob_sha1"`
	Methods     []string `json:"methods"`
	Path        string   `json:"path"`
	Repository  string   `json:"repository"`
	Role        string   `json:"role"`
	SHA256      string   `json:"sha256"`
}

type ProbeNativeEvidence struct {
	Build   string `json:"build"`
	Capture string `json:"capture"`
	Role    string `json:"role"`
	SHA256  string `json:"sha256"`
}

type ProbeRepresentative struct {
	Canonical   string `json:"canonical"`
	Constructor string `json:"constructor"`
	ResourceKey string `json:"resourceKey"`
}

type ProbeFamily struct {
	Canonical string   `json:"canonical"`
	Methods   []string `json:"methods"`
	Name      string   `json:"name"`
}

type ProbeHierarchy struct {
	Canonical string   `json:"canonical"`
	Methods   []string `json:"methods"`
	Parent    string   `json:"parent"`
}

type ProbeBuildManifest struct {
	Files map[string]string `json:"files"`
	Plan  ProbePlan         `json:"plan"`
	Probe string            `json:"probe"`
}

type ProbeReceipt struct {
	Installed           string            `json:"installed,omitempty"`
	Purpose             string            `json:"purpose,omitempty"`
	BuildManifestSHA256 string            `json:"build_manifest_sha256"`
	Files               map[string]string `json:"files"`
}

type ProbeBuildResult struct {
	Families  int    `json:"families"`
	Installed bool   `json:"installed"`
	Methods   int    `json:"methods"`
	Output    string `json:"output"`
}

type CaptureResult struct {
	HierarchyCases              *[]string `json:"hierarchy_cases,omitempty"`
	Kind                        string    `json:"kind"`
	MethodCases                 []string  `json:"method_cases"`
	PresentMethods              int       `json:"present_methods"`
	ProbeReportedPropertyPasses int       `json:"probe_reported_property_passes"`
	SuccessfulInvocations       int       `json:"successful_invocations"`
	UnobservedHierarchy         bool      `json:"unobserved_hierarchy,omitempty"`
	VerifiesSecrecy             bool      `json:"verifies_secrecy"`
}

type CaptureReport struct {
	BindingDiagnostics        []CaptureState       `json:"bindingDiagnostics"`
	BindingExperimentChecks   []CaptureState       `json:"bindingExperimentChecks"`
	BindingTextSamples        []BindingTextSample  `json:"bindingTextSamples"`
	CapabilityCaptureComplete bool                 `json:"capabilityCaptureComplete"`
	Cleanup                   []CaptureObservation `json:"cleanup"`
	ClientBuild               string               `json:"clientBuild"`
	ClientDate                string               `json:"clientDate"`
	ClientInterface           int                  `json:"clientInterface"`
	ClientVersion             string               `json:"clientVersion"`
	Context                   string               `json:"context"`
	DeclarationFingerprint    string               `json:"declarationFingerprint"`
	ExpectedBuild             string               `json:"expectedBuild"`
	ExpectedMethods           int                  `json:"expectedMethods"`
	Families                  []CaptureFamily      `json:"families"`
	FixtureSHA256             string               `json:"fixtureSHA256"`
	Hierarchy                 []CaptureObservation `json:"hierarchy"`
	InvocationCaptureComplete bool                 `json:"invocationCaptureComplete"`
	Invocations               []CaptureInvocation  `json:"invocations"`
	InvokedMethods            int                  `json:"invokedMethods"`
	Limits                    map[string]int       `json:"limits"`
	Mode                      string               `json:"mode"`
	ObservedMethods           int                  `json:"observedMethods"`
	Outcome                   string               `json:"outcome"`
	PresentMethods            int                  `json:"presentMethods"`
	ProbeVersion              string               `json:"probeVersion"`
	ProjectID                 int                  `json:"projectID"`
	Schema                    int                  `json:"schema"`
	SourceFingerprint         string               `json:"sourceFingerprint"`
	VerifiedMethods           int                  `json:"verifiedMethods"`
	VerifiesSecrecy           bool                 `json:"verifiesSecrecy"`
	VerifiesSignatures        bool                 `json:"verifiesSignatures"`
}

type CaptureFamily struct {
	Factory string               `json:"factory"`
	Family  string               `json:"family"`
	Methods []CaptureObservation `json:"methods"`
	Queries []CaptureObservation `json:"queries"`
}

type CaptureObservation struct {
	CaseID string             `json:"caseID"`
	State  string             `json:"state"`
	Value  *CaptureQueryValue `json:"value,omitempty"`
}

type CaptureState struct {
	CaseID string `json:"caseID"`
	State  string `json:"state"`
}

type CaptureInvocation struct {
	BehaviorVerified bool     `json:"behaviorVerified"`
	CaseID           string   `json:"caseID"`
	Check            string   `json:"check"`
	Invoked          bool     `json:"invoked"`
	ReturnCount      int      `json:"returnCount"`
	Returns          []string `json:"returns"`
	State            string   `json:"state"`
}

type BindingTextSample struct {
	ByteLength int    `json:"byteLength"`
	CaseID     string `json:"caseID"`
	FirstNul   int    `json:"firstNul"`
	Hex        string `json:"hex"`
	State      string `json:"state"`
}

type CaptureEvidenceFile struct {
	SHA256 string `json:"sha256"`
	Size   int    `json:"size"`
}

type CaptureEvidenceIndex struct {
	Files         map[string]CaptureEvidenceFile `json:"files"`
	SchemaVersion int                            `json:"schema_version"`
	Scope         string                         `json:"scope"`
}

// CaptureQueryValue is the public scalar returned by a probe query. Presence is
// represented by the pointer on CaptureObservation; null is never a query result.
type CaptureQueryValue struct {
	text      string
	boolean   bool
	isBoolean bool
}

func (v CaptureQueryValue) MarshalJSON() ([]byte, error) {
	if v.isBoolean {
		return json.Marshal(v.boolean)
	}
	return json.Marshal(v.text)
}
func (v *CaptureQueryValue) UnmarshalJSON(data []byte) error {
	raw, err := literal.ParseJSON(data)
	if err != nil {
		return err
	}
	if text, ok := raw.Text(); ok {
		*v = CaptureQueryValue{text: text}
		return nil
	}
	if b, ok := raw.Bool(); ok {
		*v = CaptureQueryValue{boolean: b, isBoolean: true}
		return nil
	}
	return fmt.Errorf("capture query value must be string or boolean")
}

// decodeCaptureObject converts bounded syntax into one concrete transport shape.
// Null is invalid in these evidence schemas, including optional fields when present.
func decodeCaptureObject[T any](raw literal.Object) (result T, err error) {
	defer captureError(&err)
	var nonnull func(literal.Value)
	nonnull = func(v literal.Value) {
		if v.Kind() == literal.Null {
			reject("null capture field")
		}
		if rows, ok := v.Array(); ok {
			for _, row := range rows {
				nonnull(row)
			}
		}
		if fields, ok := v.Object(); ok {
			for _, field := range fields {
				nonnull(field)
			}
		}
	}
	nonnull(literal.Record(raw))
	data, err := json.Marshal(raw)
	if err != nil {
		return result, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&result)
	return result, err
}
func captureObject(v literal.Value) literal.Object {
	obj, ok := v.Object()
	if !ok {
		reject("capture expected object")
	}
	return obj
}
func captureRows(v literal.Value) []literal.Value {
	rows, ok := v.Array()
	if !ok {
		reject("capture expected array")
	}
	return rows
}
func captureFields(raw literal.Object, required string) {
	for key := range wordSet(required) {
		if _, ok := raw[key]; !ok {
			reject(fmt.Sprintf("missing capture field: %s", key))
		}
	}
}
func captureReportFields(version string) map[string]bool {
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
	return fields
}
func decodeCaptureReport(raw literal.Object) (report CaptureReport, err error) {
	defer captureError(&err)
	version, ok := raw["probeVersion"].Text()
	if !ok {
		reject("capture expected probe version")
	}
	capExact(raw, captureReportFields(version), "capture")
	for _, family := range captureRows(raw["families"]) {
		row := captureObject(family)
		capExact(row, wordSet("family factory methods queries"), "family")
		for _, key := range []string{"methods", "queries"} {
			validateObservationShape(row[key])
		}
	}
	validateObservationShape(raw["cleanup"])
	if version != "1.0.0" {
		validateObservationShape(raw["hierarchy"])
		for _, v := range captureRows(raw["invocations"]) {
			capExact(captureObject(v), wordSet("caseID state invoked returnCount returns check behaviorVerified"), "invocation")
		}
	}
	for _, key := range []string{"bindingDiagnostics", "bindingExperimentChecks"} {
		if v, ok := raw[key]; ok {
			for _, row := range captureRows(v) {
				capExact(captureObject(row), wordSet("caseID state"), key)
			}
		}
	}
	if v, ok := raw["bindingTextSamples"]; ok {
		for _, v := range captureRows(v) {
			row := captureObject(v)
			state, _ := row["state"].Text()
			fields := wordSet("caseID state")
			if state == "public" {
				fields = wordSet("caseID state hex byteLength firstNul")
			}
			capExact(row, fields, "binding text sample")
		}
	}
	return decodeCaptureObject[CaptureReport](raw)
}
func validateObservationShape(v literal.Value) {
	for _, v := range captureRows(v) {
		row := captureObject(v)
		fields := wordSet("caseID state")
		if _, ok := row["value"]; ok {
			fields["value"] = true
		}
		capExact(row, fields, "observation")
	}
}
func decodeProbeManifest(raw literal.Object) (manifest ProbeBuildManifest, err error) {
	defer captureError(&err)
	captureFields(raw, "files plan probe")
	plan := captureObject(raw["plan"])
	captureFields(plan, "version declarationFingerprint expectedBuild expectedInterface expectedMethods expectedProjectID families limits sourceFingerprint")
	for _, v := range captureRows(plan["families"]) {
		captureFields(captureObject(v), "canonical methods name")
	}
	if rows, ok := plan["hierarchy"]; ok {
		for _, v := range captureRows(rows) {
			captureFields(captureObject(v), "canonical methods parent")
		}
	}
	return decodeCaptureObject[ProbeBuildManifest](raw)
}
func decodeProbeReceipt(raw literal.Object) (result ProbeReceipt, err error) {
	defer captureError(&err)
	captureFields(raw, "build_manifest_sha256 files")
	return decodeCaptureObject[ProbeReceipt](raw)
}

func captureMatches[T any](raw literal.Object, value T) bool {
	data, err := json.Marshal(value)
	if err != nil {
		reject(fmt.Sprintf("encode capture comparison: %s", err))
	}
	encoded, err := literal.ParseJSON(data)
	if err != nil {
		reject(fmt.Sprintf("decode capture comparison: %s", err))
	}
	return literal.Record(raw).Equal(encoded)
}
