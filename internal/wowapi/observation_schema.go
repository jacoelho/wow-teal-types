package wowapi

// ObservationLock binds a capture bundle to immutable acquisition evidence and
// the source cohort against which its observations may be applied.
type ObservationLock struct {
	Acquisition        string `json:"acquisition"`
	AcquisitionSHA256  string `json:"acquisition_sha256"`
	Build              string `json:"build"`
	Bundle             string `json:"bundle"`
	EvidenceSHA256     string `json:"evidence_sha256"`
	SchemaVersion      int    `json:"schema_version"`
	SourceInputsSHA256 string `json:"source_inputs_sha256"`
}
type BehaviorReviewRule struct {
	BindingSample    string `json:"binding_sample"`
	CorrectedCheck   string `json:"corrected_check"`
	Disposition      string `json:"disposition"`
	ProjectionSample string `json:"projection_sample"`
	RawSample        string `json:"raw_sample"`
	Reason           string `json:"reason"`
	RecordedCheck    string `json:"recorded_check"`
	Symbol           string `json:"symbol"`
}
type BehaviorReviewSources struct {
	Observation   ObservationLock    `json:"observation"`
	Review        BehaviorReviewRule `json:"review"`
	SchemaVersion int                `json:"schema_version"`
}
type BehaviorReview struct {
	BehaviorVerified bool            `json:"behavior_verified"`
	Build            string          `json:"build"`
	Check            string          `json:"check"`
	Disposition      string          `json:"disposition"`
	Observation      ObservationLock `json:"observation"`
	Reason           string          `json:"reason"`
	Scope            string          `json:"scope"`
	Symbol           string          `json:"symbol"`
}
type NativeObservations struct {
	Sources               ObservationLock
	Report                CaptureReport
	Plan                  ProbePlan
	Applicable            bool
	Methods               map[string]CaptureInvocation
	Hierarchy             map[string]CaptureObservation
	BehaviorReviewSources BehaviorReviewSources
	BehaviorReviews       map[string]BehaviorReview
}
type captureAcquisition struct {
	Acquisition                 string   `json:"acquisition"`
	Bundle                      string   `json:"bundle"`
	BundleIndexSHA256           string   `json:"bundle_index_sha256"`
	CleanupComplete             bool     `json:"cleanup_complete"`
	ClientBuild                 string   `json:"client_build"`
	Context                     string   `json:"context"`
	InstalledFilesVerified      bool     `json:"installed_files_verified_against_retained_receipt"`
	Limits                      []string `json:"limits"`
	LoaderResult                string   `json:"loader_result_reported_by_user"`
	ProbeVersion                string   `json:"probe_version"`
	PropertyChecksPassed        int      `json:"property_checks_passed"`
	PropertyMismatches          []string `json:"property_mismatches"`
	Schema                      int      `json:"schema"`
	SuccessfulMethodInvocations int      `json:"successful_method_invocations"`
}
