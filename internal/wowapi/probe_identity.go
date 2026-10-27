package wowapi

import "wow-teal-types/internal/sourcelock"

// Field order preserves the canonical fingerprint used by retained captures.
type probeSourceIdentity struct {
	BindingDiagnosticCases []string               `json:"bindingDiagnosticCases"`
	BindingExperimentCases []string               `json:"bindingExperimentCases"`
	BindingTextSampleBytes int                    `json:"bindingTextSampleBytes"`
	BindingTextSampleCases []string               `json:"bindingTextSampleCases"`
	Candidates             ProbeCandidateEvidence `json:"candidates"`
	Catalogue              sourcelock.Sources     `json:"catalogue"`
	Hierarchy              sourcelock.Crosscheck  `json:"hierarchy"`
	NativeCandidates       ProbeNativeEvidence    `json:"nativeCandidates"`
	NativeRequirements     []ProbeHierarchy       `json:"nativeRequirements"`
	RepresentativeMethods  []string               `json:"representativeMethods"`
	Representatives        []ProbeRepresentative  `json:"representatives"`
}
