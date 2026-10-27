package wowapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"wow-teal-types/internal/sourcelock"
	"wow-teal-types/internal/wowapi/literal"
)

type Manifest struct {
	Files                        []ManifestFile                       `json:"files"`
	NativeTables                 []ManifestNativeTables               `json:"native_tables"`
	QuestieSymbols               []ManifestAnnotationSymbol           `json:"questie_symbols"`
	Representations              []RepresentationNote                 `json:"representations"`
	SourceDiagnostics            []AnnotationDiagnostic               `json:"source_diagnostics"`
	Sources                      map[string]sourcelock.SourceIdentity `json:"sources"`
	Statistics                   CatalogueStatistics                  `json:"statistics"`
	SupplementalMethodExclusions map[string]SupplementalExclusion     `json:"supplemental_method_exclusions"`
	Symbols                      []ManifestSymbol                     `json:"symbols"`
	Types                        []ManifestType                       `json:"types"`
}
type ManifestFile struct {
	Disposition string `json:"disposition"`
	Source      string `json:"source"`
}
type ManifestType struct {
	Bases  []string `json:"bases"`
	Export string   `json:"export"`
	Kind   string   `json:"kind"`
	Name   string   `json:"name"`
	Source string   `json:"source"`
}
type ManifestAnnotationSymbol struct {
	Excluded       bool                  `json:"excluded,omitempty"`
	Line           int                   `json:"line"`
	Name           string                `json:"name"`
	Parameters     []AnnotationParameter `json:"parameters"`
	Returns        []AnnotationReturn    `json:"returns"`
	SelectedSource string                `json:"selected_source"`
	Source         string                `json:"source"`
}

// ManifestNativeTables retains raw source tables/predicates; declarations use
// NativeTable and NativePredicate in the catalogue instead.
type ManifestNativeTables struct {
	Predicates []literal.Value `json:"predicates"`
	Source     string          `json:"source"`
	Tables     []literal.Value `json:"tables"`
}
type ManifestSymbol struct {
	Name, Source string
	Function     *NativeFunction
	Event        *NativeEvent
	Restrictions NativePolicy
}

func (symbol ManifestSymbol) MarshalJSON() ([]byte, error) {
	if symbol.Event != nil {
		return json.Marshal(struct {
			Event  literal.Value `json:"event"`
			Name   string        `json:"name"`
			Source string        `json:"source"`
		}{literal.Record(symbol.Event.Raw), symbol.Name, symbol.Source})
	}
	if symbol.Function == nil {
		return nil, fmt.Errorf("manifest symbol %s lacks contract", symbol.Name)
	}
	arguments := rawNativeParameters(symbol.Function.Arguments)
	returns := rawNativeParameters(symbol.Function.Returns)
	return json.Marshal(struct {
		Arguments        []literal.Value `json:"arguments"`
		MayReturnNothing bool            `json:"may_return_nothing"`
		Metadata         literal.Value   `json:"metadata"`
		Name             string          `json:"name"`
		Restrictions     NativePolicy    `json:"restrictions"`
		Returns          []literal.Value `json:"returns"`
		Source           string          `json:"source"`
	}{arguments, enabled(symbol.Function.Policy.MayReturnNothing), literal.Record(symbol.Function.Raw), symbol.Name, symbol.Restrictions, returns, symbol.Source})
}
func (symbol *ManifestSymbol) UnmarshalJSON(data []byte) error {
	var row struct {
		Arguments        []literal.Value `json:"arguments"`
		MayReturnNothing *bool           `json:"may_return_nothing"`
		Metadata         *literal.Value  `json:"metadata"`
		Event            *literal.Value  `json:"event"`
		Name             string          `json:"name"`
		Restrictions     NativePolicy    `json:"restrictions"`
		Returns          []literal.Value `json:"returns"`
		Source           string          `json:"source"`
	}
	if err := decodeJSON(data, &row); err != nil {
		return err
	}
	*symbol = ManifestSymbol{Name: row.Name, Source: row.Source, Restrictions: row.Restrictions}
	if row.Event != nil {
		raw, ok := row.Event.Object()
		if !ok {
			return fmt.Errorf("manifest event metadata must be an object")
		}
		event, err := decodeNativeEvent(raw)
		if err != nil {
			return err
		}
		symbol.Event = &event
		return nil
	}
	if row.Metadata == nil {
		return fmt.Errorf("manifest symbol lacks metadata")
	}
	raw, ok := row.Metadata.Object()
	if !ok {
		return fmt.Errorf("manifest function metadata must be an object")
	}
	function, err := decodeNativeFunction(raw)
	if err != nil {
		return err
	}
	symbol.Function = &function
	if !literal.List(row.Arguments).Equal(literal.List(rawNativeParameters(function.Arguments))) || !literal.List(row.Returns).Equal(literal.List(rawNativeParameters(function.Returns))) || row.MayReturnNothing == nil || *row.MayReturnNothing != enabled(function.Policy.MayReturnNothing) {
		return fmt.Errorf("manifest symbol projected metadata differs")
	}
	return nil
}
func rawNativeParameters(parameters []NativeParameter) []literal.Value {
	out := make([]literal.Value, len(parameters))
	for i, parameter := range parameters {
		out[i] = literal.Record(parameter.Raw)
	}
	return out
}

// EvidenceReferences preserves the established scalar/list form of a citation.
// Both forms identify references; neither can carry arbitrary domain payloads.
type EvidenceReferences struct {
	Text *string
	List []string
}

func (references EvidenceReferences) MarshalJSON() ([]byte, error) {
	if references.Text != nil {
		return json.Marshal(*references.Text)
	}
	return json.Marshal(references.List)
}
func (references *EvidenceReferences) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '"' {
		var text string
		if err := json.Unmarshal(data, &text); err != nil {
			return err
		}
		*references = EvidenceReferences{Text: &text}
		return nil
	}
	var list []string
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}
	*references = EvidenceReferences{List: list}
	return nil
}

type RepresentationNote struct {
	CurrentSourceCohortMatches *bool               `json:"current_source_cohort_matches,omitempty"`
	Evidence                   *EvidenceReferences `json:"evidence,omitempty"`
	Expression                 string              `json:"expression,omitempty"`
	ObservedBuild              string              `json:"observed_build,omitempty"`
	Parameter                  string              `json:"parameter,omitempty"`
	Path                       string              `json:"path,omitempty"`
	Reason                     string              `json:"reason,omitempty"`
	Representation             string              `json:"representation,omitempty"`
	Rule                       string              `json:"rule,omitempty"`
	Source                     string              `json:"source,omitempty"`
	Symbol                     *string             `json:"symbol,omitempty"`
	Symbols                    []string            `json:"symbols,omitempty"`
	Type                       string              `json:"type,omitempty"`
}
type SupplementalExclusion struct {
	Build                      string                     `json:"build"`
	CurrentSourceCohortMatches bool                       `json:"current_source_cohort_matches"`
	Evidence                   []string                   `json:"evidence"`
	Method                     string                     `json:"method"`
	NegativeCases              []string                   `json:"negative_cases"`
	ObservationSources         ObservationLock            `json:"observation_sources"`
	Original                   ExcludedAnnotationFunction `json:"original"`
	Owner                      string                     `json:"owner"`
	Reason                     string                     `json:"reason"`
	Signature                  string                     `json:"signature"`
	Source                     string                     `json:"source"`
	SourceSHA256               string                     `json:"source_sha256"`
}
type ExcludedAnnotationFunction struct {
	AnnotationFunction
	Native bool   `json:"native"`
	Source string `json:"source"`
}

func decodeJSON[T any](data []byte, value *T) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("trailing JSON data")
	}
	return nil
}
func ReadManifest(path string) (Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, err
	}
	return ParseManifest(data)
}
func ParseManifest(data []byte) (Manifest, error) {
	var manifest Manifest
	err := decodeJSON(data, &manifest)
	return manifest, err
}
