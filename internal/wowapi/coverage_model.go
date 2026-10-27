package wowapi

import "wow-teal-types/internal/sourcelock"

type DeclarationSurface struct {
	SHA256                 string              `json:"sha256"`
	TotalTypeNames         int                 `json:"total_type_names"`
	GlobalSignatureAliases []string            `json:"global_signature_aliases"`
	OpaquePolicyExports    []string            `json:"opaque_policy_exports"`
	AmbientGlobals         map[string][]string `json:"ambient_globals"`
}
type NativeCoverageDocument struct {
	Source          string `json:"source"`
	ClientPath      string `json:"client_path"`
	ReviewedBlobSHA string `json:"reviewed_blob_sha"`
	FunctionCount   int    `json:"function_count"`
	EventCount      int    `json:"event_count"`
	TableCount      int    `json:"table_count"`
}
type NativeCoverageContract struct {
	Kind           string `json:"kind"`
	Name           string `json:"name"`
	Source         string `json:"source"`
	SourceCategory string `json:"source_category"`
	ClientPath     string `json:"client_path"`
	MetadataSHA256 string `json:"metadata_sha256"`
}
type CoverageType struct {
	ManifestType
	SourceCategory        string   `json:"source_category"`
	DirectNativeContracts int      `json:"direct_native_contracts"`
	NativeContractBases   []string `json:"native_contract_bases"`
}
type CoverageSupplement struct {
	Name           string `json:"name"`
	Source         string `json:"source"`
	Line           int    `json:"line"`
	SelectedSource string `json:"selected_source"`
}
type CoverageLibrary struct {
	Name               string                    `json:"name"`
	SourceIdentity     sourcelock.SourceIdentity `json:"source_identity"`
	DeriveFromClient   bool                      `json:"derive_from_client"`
	Types              []string                  `json:"types"`
	SymbolDeclarations []string                  `json:"symbol_declarations"`
	Files              []sourcelock.SourceFile   `json:"files"`
}
type coverageClient struct {
	sourcelock.ClientIdentity
	Extractor sourcelock.Extractor `json:"extractor"`
	Manifests map[string]int       `json:"manifests"`
	Schema    int                  `json:"schema"`
}
type coverageInputs struct {
	ClientLockSHA256          string                               `json:"client_lock_sha256"`
	DeclarationManifestSHA256 string                               `json:"declaration_manifest_sha256"`
	SourceLockSHA256          string                               `json:"source_lock_sha256"`
	Client                    coverageClient                       `json:"client"`
	SourceCohort              map[string]sourcelock.SourceIdentity `json:"source_cohort"`
}
type coverageStatistics struct {
	ClientFiles                      int            `json:"client_files"`
	NativeLuaDocuments               int            `json:"native_lua_documents"`
	NativeFunctions                  int            `json:"native_functions"`
	NativeEvents                     int            `json:"native_events"`
	CatalogueTypeExports             int            `json:"catalogue_type_exports"`
	TypeOrigins                      map[string]int `json:"type_origins"`
	RetainedSupplementalDeclarations int            `json:"retained_supplemental_declarations"`
	RetainedSupplementalCategories   map[string]int `json:"retained_supplemental_categories"`
	ThirdPartyTypeExports            int            `json:"third_party_type_exports"`
	ThirdPartySymbolDeclarations     int            `json:"third_party_symbol_declarations"`
}
type CoverageReport struct {
	SchemaVersion                int                              `json:"schema_version"`
	Claim                        string                           `json:"claim"`
	SourceCategories             map[string]string                `json:"source_categories"`
	Inputs                       coverageInputs                   `json:"inputs"`
	Statistics                   coverageStatistics               `json:"statistics"`
	DeclarationSurface           DeclarationSurface               `json:"declaration_surface"`
	NativeDocuments              []NativeCoverageDocument         `json:"native_documents"`
	NativeContracts              []NativeCoverageContract         `json:"native_contracts"`
	Types                        []CoverageType                   `json:"types"`
	RemainingAnnotationContracts map[string][]CoverageSupplement  `json:"remaining_annotation_contracts"`
	ThirdPartyLibraries          []CoverageLibrary                `json:"third_party_libraries"`
	ReviewedFrameXML             frameXMLSupplement               `json:"reviewed_framexml"`
	RepresentationDecisions      []RepresentationNote             `json:"representation_decisions"`
	SupplementalMethodExclusions map[string]SupplementalExclusion `json:"supplemental_method_exclusions"`
	Limits                       []string                         `json:"limits"`
}
