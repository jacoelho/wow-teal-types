package wowapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"wow-teal-types/internal/wowapi/literal"
)

type resolvedIdentity struct {
	BlobSHA256  string `json:"blob_sha256"`
	BlobSize    int    `json:"blob_size"`
	BlobVersion int    `json:"blob_version"`
	Commit      string `json:"commit"`
	Repository  string `json:"repository"`
}
type resolvedRequest struct {
	Symbol    string `json:"symbol"`
	TSMSymbol string `json:"tsm_symbol"`
}
type policyPredicate struct {
	Doc  *string `json:"doc"`
	Name string  `json:"name"`
}
type policyPrecondition struct {
	Doc     *string `json:"doc"`
	Failure *string `json:"failure"`
	Name    string  `json:"name"`
}
type policyArguments struct {
	Params []string `json:"params"`
	Policy string   `json:"policy"`
}
type policyUnless struct {
	Param  string   `json:"param"`
	Values []string `json:"values"`
}

// These vendor policy fields have no supported non-null representation yet.
type nullPolicy struct{}

func (nullPolicy) MarshalJSON() ([]byte, error) { return []byte("null"), nil }
func (*nullPolicy) UnmarshalJSON(data []byte) error {
	if !bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return fmt.Errorf("unsupported resolved policy form")
	}
	return nil
}

type resolvedPolicy struct {
	Args             *policyArguments     `json:"args"`
	Aspects          []string             `json:"aspects"`
	Clears           nullPolicy           `json:"clears"`
	Guard            nullPolicy           `json:"guard"`
	Preconditions    []policyPrecondition `json:"preconditions"`
	RestrictionGuard nullPolicy           `json:"restriction_guard"`
	Satisfies        nullPolicy           `json:"satisfies"`
	Unless           *policyUnless        `json:"unless"`
	When             []policyPredicate    `json:"when"`
}

func (policy *resolvedPolicy) UnmarshalJSON(data []byte) error {
	raw, err := literal.ParseJSON(data)
	if err != nil {
		return err
	}
	object, ok := raw.Object()
	if !ok {
		return fmt.Errorf("resolved policy must be object")
	}
	required := []string{"args", "aspects", "clears", "guard", "preconditions", "restriction_guard", "satisfies", "unless", "when"}
	if len(object) != len(required) {
		return fmt.Errorf("incomplete resolved policy")
	}
	for _, name := range required {
		if _, ok := object[name]; !ok {
			return fmt.Errorf("missing resolved policy %s", name)
		}
	}
	type policyFields resolvedPolicy
	var decoded policyFields
	if err = json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	if decoded.Aspects == nil || decoded.Preconditions == nil || decoded.When == nil {
		return fmt.Errorf("resolved policy arrays must be present")
	}
	*policy = resolvedPolicy(decoded)
	return nil
}

type policyCheck struct {
	ResolvedSecretSHA256 string `json:"resolved_secret_sha256"`
	SourceSecuritySHA256 string `json:"source_security_sha256"`
	State                string `json:"state"`
}
type resolvedLocation struct {
	End       int    `json:"end"`
	NameEnd   int    `json:"name_end,omitempty"`
	NameStart int    `json:"name_start,omitempty"`
	Path      string `json:"path"`
	Start     int    `json:"start"`
}

func (location *resolvedLocation) UnmarshalJSON(data []byte) error {
	var row struct {
		End       *int   `json:"end"`
		NameEnd   int    `json:"name_end"`
		NameStart int    `json:"name_start"`
		Path      string `json:"path"`
		Start     *int   `json:"start"`
	}
	if err := json.Unmarshal(data, &row); err != nil {
		return err
	}
	if row.End == nil || row.Start == nil || row.Path == "" {
		return fmt.Errorf("incomplete resolved location")
	}
	*location = resolvedLocation{*row.End, row.NameEnd, row.NameStart, row.Path, *row.Start}
	return nil
}

type resolvedFunction struct {
	OverloadCount int
	Secret        *resolvedPolicy
}

func (function *resolvedFunction) UnmarshalJSON(data []byte) error {
	raw, err := literal.ParseJSON(data)
	if err != nil {
		return err
	}
	object, ok := raw.Object()
	if !ok {
		return fmt.Errorf("resolved function must be object")
	}
	if _, present := object["secret"]; !present {
		return fmt.Errorf("resolved function lacks secret facts")
	}
	overloads, ok := object["overloads"].Array()
	if !ok {
		return fmt.Errorf("resolved function lacks overload facts")
	}
	var row struct {
		Secret *resolvedPolicy `json:"secret"`
	}
	if err = json.Unmarshal(data, &row); err != nil {
		return err
	}
	*function = resolvedFunction{OverloadCount: len(overloads), Secret: row.Secret}
	return nil
}
func (function resolvedFunction) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Overloads []struct{}      `json:"overloads"`
		Secret    *resolvedPolicy `json:"secret"`
	}{make([]struct{}, function.OverloadCount), function.Secret})
}

type resolvedBinding struct {
	Raw               literal.Object
	Symbol, TSMSymbol string
	Location          resolvedLocation
	Function          *resolvedFunction
	Source            string
	Line              int
	Candidate         *tsmStub
	CandidateSHA256   string
	PolicyCheck       *policyCheck
}

func (binding *resolvedBinding) UnmarshalJSON(data []byte) error {
	raw, err := literal.ParseJSON(data)
	if err != nil {
		return err
	}
	object, ok := raw.Object()
	if !ok {
		return fmt.Errorf("resolved binding must be object")
	}
	var row struct {
		Symbol    string            `json:"symbol"`
		TSMSymbol string            `json:"tsm_symbol"`
		Location  resolvedLocation  `json:"location"`
		Function  *resolvedFunction `json:"function"`
	}
	if err = json.Unmarshal(data, &row); err != nil {
		return err
	}
	if row.Symbol == "" || row.TSMSymbol == "" || row.Function == nil {
		return fmt.Errorf("incomplete resolved binding")
	}
	*binding = resolvedBinding{Raw: object, Symbol: row.Symbol, TSMSymbol: row.TSMSymbol, Location: row.Location, Function: row.Function}
	return nil
}
func (binding resolvedBinding) MarshalJSON() ([]byte, error) {
	object := maps.Clone(binding.Raw)
	if object == nil {
		data, err := json.Marshal(struct {
			Function  *resolvedFunction `json:"function"`
			Location  resolvedLocation  `json:"location"`
			Symbol    string            `json:"symbol"`
			TSMSymbol string            `json:"tsm_symbol"`
		}{binding.Function, binding.Location, binding.Symbol, binding.TSMSymbol})
		if err != nil {
			return nil, err
		}
		syntax, err := literal.ParseJSON(data)
		if err != nil {
			return nil, err
		}
		object, _ = syntax.Object()
	}
	if binding.Source != "" {
		data, err := json.Marshal(struct {
			Candidate       *tsmStub     `json:"candidate,omitempty"`
			CandidateSHA256 string       `json:"candidate_sha256"`
			Line            int          `json:"line"`
			PolicyCheck     *policyCheck `json:"policy_check"`
			Source          string       `json:"source"`
		}{binding.Candidate, binding.CandidateSHA256, binding.Line, binding.PolicyCheck, binding.Source})
		if err != nil {
			return nil, err
		}
		syntax, err := literal.ParseJSON(data)
		if err != nil {
			return nil, err
		}
		extra, _ := syntax.Object()
		for name, value := range extra {
			object[name] = value
		}
	}
	return json.Marshal(object)
}

type resolvedExport struct {
	Bindings        map[string]resolvedBinding `json:"bindings"`
	BlobSHA256      string                     `json:"blob_sha256"`
	BlobSize        int                        `json:"blob_size"`
	BlobVersion     int                        `json:"blob_version"`
	ExporterSources map[string]string          `json:"exporter_sources"`
	SchemaVersion   int                        `json:"schema_version"`
	TSMCommit       string                     `json:"tsm_commit"`
}
type resolvedSources struct {
	ArtifactSHA256  string            `json:"artifact_sha256"`
	BlobSHA256      string            `json:"blob_sha256"`
	BlobSize        int               `json:"blob_size"`
	BlobVersion     int               `json:"blob_version"`
	Commit          string            `json:"commit"`
	ExporterSources map[string]string `json:"exporter_sources"`
	Repository      string            `json:"repository"`
}
type resolvedEvidence struct {
	Bindings map[string]resolvedBinding `json:"bindings"`
	Sources  resolvedSources            `json:"sources"`
}
