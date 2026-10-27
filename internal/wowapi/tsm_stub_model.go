package wowapi

import "encoding/json"

type tsmParameter struct {
	Name     string `json:"name"`
	Optional bool   `json:"optional"`
	Type     string `json:"type"`
}
type tsmReturn struct {
	Label string `json:"label"`
	Type  string `json:"type"`
}
type unsupportedAnnotation struct {
	Annotation string `json:"annotation"`
	Reason     string `json:"reason"`
}
type invalidParameter struct {
	Position  int    `json:"position,omitempty"`
	Positions []int  `json:"positions,omitempty"`
	Raw       string `json:"raw,omitempty"`
	Reason    string `json:"reason"`
}
type sourceDefect struct {
	ComparisonScope string             `json:"comparison_scope"`
	Declaration     string             `json:"declaration"`
	Parameters      []invalidParameter `json:"parameters"`
}
type tsmStub struct {
	Arguments    []string                `json:"arguments"`
	Generics     []string                `json:"generics"`
	Line         int                     `json:"line"`
	Method       bool                    `json:"method"`
	Overloads    []string                `json:"overloads"`
	Parameters   []tsmParameter          `json:"parameters"`
	Returns      []tsmReturn             `json:"returns"`
	Security     map[string][]string     `json:"security"`
	Source       string                  `json:"source"`
	SourceDefect *sourceDefect           `json:"source_defect,omitempty"`
	Unsupported  []unsupportedAnnotation `json:"unsupported,omitempty"`
}
type stubDiagnostic struct {
	Annotations []unsupportedAnnotation `json:"annotations"`
	Line        int                     `json:"line"`
	Source      string                  `json:"source"`
	Symbol      string                  `json:"symbol"`
}
type simpleFact struct {
	Nilable bool   `json:"nilable"`
	Secret  bool   `json:"secret"`
	Type    string `json:"type"`
}
type stubSource struct {
	Line   int    `json:"line"`
	Source string `json:"source"`
}
type primarySecurityOption struct {
	Security map[string][]string `json:"security"`
	Sources  []stubSource        `json:"sources"`
}
type stubDefinitions struct {
	Active                 []tsmStub               `json:"active"`
	Basis                  []string                `json:"basis"`
	Discarded              []tsmStub               `json:"discarded"`
	Mode                   string                  `json:"mode"`
	OrderUncertain         bool                    `json:"order_uncertain"`
	PossiblePrimary        []tsmStub               `json:"possible_primary"`
	PrimarySecurityOptions []primarySecurityOption `json:"primary_security_options"`
	ResolvedBinding        *resolvedBinding        `json:"resolved_binding,omitempty"`
}

func (s stubDefinitions) MarshalJSON() ([]byte, error) {
	if s.Mode == "native-only" {
		return json.Marshal(struct {
			Active    []tsmStub `json:"active"`
			Discarded []tsmStub `json:"discarded"`
			Mode      string    `json:"mode"`
		}{s.Active, s.Discarded, s.Mode})
	}
	type full stubDefinitions
	return json.Marshal(full(s))
}
