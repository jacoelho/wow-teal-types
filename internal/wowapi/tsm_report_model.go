package wowapi

import (
	"encoding/json"
	"wow-teal-types/internal/sourcelock"
	"wow-teal-types/internal/wowapi/literal"
)

type nativeMatch struct {
	AddedFiles            []string `json:"added_files"`
	ChangedOrRemovedFiles []string `json:"changed_or_removed_files"`
	MatchedFiles          int      `json:"matched_files"`
}
type nativeSnapshot struct {
	nativeMatch
	TSMBuild        string                     `json:"tsm_build"`
	TSMCommit       string                     `json:"tsm_commit"`
	WorkspaceBuild  string                     `json:"workspace_build"`
	WorkspaceCommit string                     `json:"workspace_commit,omitempty"`
	WorkspaceSource *sourcelock.SourceIdentity `json:"workspace_source,omitempty"`
}
type tsmIdentity struct {
	BlobVersion   int    `json:"blob_version"`
	Commit        string `json:"commit"`
	EmbeddedFiles int    `json:"embedded_files"`
	Release       string `json:"release"`
	Repository    string `json:"repository"`
}
type tsmCoverage struct {
	Compared                  []string         `json:"compared"`
	EmbeddedSecretAnnotations map[string]int   `json:"embedded_secret_annotations"`
	Limits                    []string         `json:"limits"`
	UnsupportedAnnotations    []stubDiagnostic `json:"unsupported_annotations"`
}
type emittedPolicy struct {
	NativeHandleTypes         []string `json:"native_handle_types"`
	SecretDataTypes           []string `json:"secret_data_types"`
	StaticCarrierSourceSHA256 string   `json:"static_carrier_source_sha256"`
}
type symbolContext struct {
	NativeSource string `json:"native_source"`
	Symbol       string `json:"symbol"`
}
type symbolDifferences struct {
	symbolContext
	Differences []contractDifference `json:"differences"`
	Disposition string               `json:"disposition"`
}
type reconciledContract struct {
	symbolContext
	Coverage           string   `json:"coverage"`
	Discrepancies      []string `json:"discrepancies"`
	Fingerprint        string   `json:"fingerprint"`
	RequiresResolution bool     `json:"requires_resolution"`
}
type typePosition struct {
	ContractSource string       `json:"contract_source,omitempty"`
	EffectiveFact  *typeFact    `json:"effective_fact,omitempty"`
	Fact           positionFact `json:"fact"`
	Position       int          `json:"position"`
	Side           string       `json:"side"`
}

// Tree is omitted only for secondary type-position explanations, as in the
// maintained evidence format; the full fact remains available for comparison.
type positionFact struct {
	Value    typeFact
	OmitTree bool
}

func (p positionFact) MarshalJSON() ([]byte, error) {
	b, e := json.Marshal(p.Value)
	if e != nil || !p.OmitTree {
		return b, e
	}
	raw, e := literal.ParseJSON(b)
	if e != nil {
		return nil, e
	}
	object, ok := raw.Object()
	if !ok {
		panic("type fact must encode object")
	}
	delete(object, "tree")
	return json.Marshal(object)
}

type comparisonVariant struct {
	Branch                    string               `json:"branch"`
	Differences               []contractDifference `json:"differences"`
	EffectiveFunctionSecurity map[string][]string  `json:"effective_function_security"`
	Line                      int                  `json:"line"`
	PossiblePrimarySources    []stubSource         `json:"possible_primary_sources"`
	RawDifferences            []contractDifference `json:"raw_differences"`
	RawSecurity               map[string][]string  `json:"raw_security"`
	Source                    string               `json:"source"`
	TypeFacts                 []typePosition       `json:"type_facts"`
}
type symbolComparison struct {
	symbolContext
	Candidates []tsmStub           `json:"candidates"`
	Selection  stubDefinitions     `json:"selection"`
	TSMSymbol  string              `json:"tsm_symbol"`
	Variants   []comparisonVariant `json:"variants"`
}
type nativeOnlyContract struct {
	symbolContext
	EmittedContract     effectiveDeclaration  `json:"emitted_contract"`
	Evidence            []string              `json:"evidence"`
	Reason              string                `json:"reason"`
	ScriptObject        string                `json:"script_object"`
	SecondaryCrosscheck *NativeMethodEvidence `json:"secondary_crosscheck"`
}
type contractSnapshot struct {
	symbolContext
	Candidates          []tsmStub              `json:"candidates"`
	Comparisons         []comparisonVariant    `json:"comparisons"`
	Differences         []contractDifference   `json:"differences"`
	Emitted             effectiveDeclaration   `json:"emitted"`
	Native              literal.Object         `json:"native"`
	SecondaryCrosscheck **NativeMethodEvidence `json:"secondary_crosscheck,omitempty"`
	Selection           stubDefinitions        `json:"selection"`
}
type sourceClosure struct {
	Native    sourcelock.Sources    `json:"native"`
	Resolved  resolvedSources       `json:"resolved"`
	Secondary NativeEvidenceSources `json:"secondary"`
	TSM       sourcelock.TSM        `json:"tsm"`
}
type CrosscheckReport struct {
	Comparisons             []symbolComparison      `json:"comparisons"`
	Contracts               []reconciledContract    `json:"contracts"`
	Coverage                tsmCoverage             `json:"coverage"`
	DeclarationGraphSHA256  string                  `json:"declaration_graph_sha256"`
	DifferenceKinds         map[string]int          `json:"difference_kinds"`
	Differences             []symbolDifferences     `json:"differences"`
	EffectiveDeclarations   []effectiveDeclaration  `json:"effective_declarations"`
	EffectivePolicy         emittedPolicy           `json:"effective_policy"`
	NativeOnly              []nativeOnlyContract    `json:"native_only"`
	NativeSecondaryEvidence NativeEvidence          `json:"native_secondary_evidence"`
	NativeSnapshot          nativeSnapshot          `json:"native_snapshot"`
	Purpose                 string                  `json:"purpose"`
	Reconciliation          ReconciliationReport    `json:"reconciliation"`
	SchemaVersion           int                     `json:"schema_version"`
	SourceLocksSHA256       string                  `json:"source_locks_sha256"`
	Statistics              map[string]int          `json:"statistics"`
	TSM                     tsmIdentity             `json:"tsm"`
	TSMResolvedSources      resolvedSources         `json:"tsm_resolved_sources"`
	TSMSources              []sourcelock.LockedFile `json:"tsm_sources"`
}
type resolutionPolicy struct {
	Decision string   `json:"decision"`
	Evidence []string `json:"evidence"`
	Reason   string   `json:"reason"`
}
type positionResolution struct {
	DifferenceSHA256 string         `json:"difference_sha256"`
	Facts            literal.Object `json:"facts"`
	Policy           string         `json:"policy"`
}
type parserCorrection struct {
	PreviousFingerprint string `json:"previous_fingerprint"`
	Reason              string `json:"reason"`
}

type contractApproval struct {
	ParserCorrection *parserCorrection    `json:"parser_correction,omitempty"`
	Coverage         string               `json:"coverage"`
	Fingerprint      string               `json:"fingerprint"`
	NativeSource     string               `json:"native_source"`
	Policies         []string             `json:"policies"`
	Resolutions      []positionResolution `json:"resolutions"`
}
type ReconciliationLedger struct {
	Contracts              map[string]contractApproval `json:"contracts"`
	DeclarationGraphSHA256 string                      `json:"declaration_graph_sha256"`
	Policies               map[string]resolutionPolicy `json:"policies"`
	ReviewDate             string                      `json:"review_date"`
	ReviewScope            string                      `json:"review_scope"`
	SchemaVersion          int                         `json:"schema_version"`
	SourceLocksSHA256      string                      `json:"source_locks_sha256"`
}
type reconciliationDecision struct {
	Policies []string `json:"policies"`
	State    string   `json:"state"`
	Symbol   string   `json:"symbol"`
}
type ReconciliationReport struct {
	Decisions               []reconciliationDecision    `json:"decisions"`
	DeclarationGraphMatches bool                        `json:"declaration_graph_matches"`
	OrphanedResolutions     []string                    `json:"orphaned_resolutions"`
	Passed                  bool                        `json:"passed"`
	Policies                map[string]resolutionPolicy `json:"policies"`
	SourceEquivalence       *EquivalenceProof           `json:"source_equivalence,omitempty"`
	SourceLocksMatch        bool                        `json:"source_locks_match"`
	UnresolvedContracts     int                         `json:"unresolved_contracts"`
}
