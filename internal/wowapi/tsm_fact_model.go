package wowapi

import (
	"encoding/json"
	"wow-teal-types/internal/wowapi/literal"
)

type policyScope struct {
	Policy       string   `json:"policy,omitempty"`
	Positions    []int    `json:"positions,omitzero"`
	Raw          string   `json:"raw"`
	Scope        string   `json:"scope,omitempty"`
	UnknownNames []string `json:"unknown_names,omitzero"`
	Unsupported  string   `json:"unsupported,omitempty"`
}
type factAlias struct {
	Name, Source string
	Line         int
	Members      []string
	Unsupported  string
}
type factEvidence struct {
	Expression     *string        `json:"expression,omitempty"`
	Line           *int           `json:"line,omitempty"`
	Metadata       literal.Object `json:"metadata,omitempty"`
	Name           string         `json:"name"`
	Representation string         `json:"representation,omitempty"`
	Selected       *bool          `json:"selected,omitempty"`
	SelectionBasis string         `json:"selection_basis,omitempty"`
	Source         string         `json:"source"`
	Unsupported    string         `json:"unsupported,omitempty"`
}
type genericBinder struct {
	Constraint factNode `json:"constraint"`
	Pack       bool     `json:"pack"`
}
type factPathPart struct {
	label string
	index *int
}

func (p factPathPart) MarshalJSON() ([]byte, error) {
	if p.index != nil {
		return json.Marshal(*p.index)
	}
	return json.Marshal(p.label)
}
func factPath(path []factPathPart, label string, index *int) []factPathPart {
	out := append(append([]factPathPart{}, path...), factPathPart{label: label})
	if index != nil {
		out = append(out, factPathPart{index: index})
	}
	return out
}

type secretWrapper struct {
	Path []factPathPart `json:"path"`
	Type string         `json:"type"`
}
type typeFact struct {
	Binders        []genericBinder `json:"binders,omitempty"`
	CarrierType    string          `json:"carrier_type"`
	Evidence       []factEvidence  `json:"evidence"`
	Features       []string        `json:"features"`
	Nilable        bool            `json:"nilable"`
	Raw            string          `json:"raw"`
	Secret         bool            `json:"secret"`
	SecretWrappers []secretWrapper `json:"secret_wrappers"`
	Tree           factNode        `json:"tree"`
	Type           string          `json:"type"`
	Unsupported    string          `json:"-"`
	Variadic       bool            `json:"variadic"`
}

func (f typeFact) MarshalJSON() ([]byte, error) {
	if f.Unsupported != "" {
		return json.Marshal(struct {
			Evidence    []factEvidence `json:"evidence"`
			Raw         string         `json:"raw"`
			Tree        factNode       `json:"tree"`
			Type        *string        `json:"type"`
			Unsupported string         `json:"unsupported"`
		}{Evidence: f.Evidence, Raw: f.Raw, Unsupported: f.Unsupported})
	}
	type success typeFact
	return json.Marshal(success(f))
}
