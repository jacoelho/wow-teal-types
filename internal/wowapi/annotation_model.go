package wowapi

import (
	"encoding/json"
	"fmt"

	"wow-teal-types/internal/wowapi/literal"
)

type AnnotationDocument struct {
	Aliases     []*AnnotationAlias     `json:"aliases"`
	Annotations []AnnotationTag        `json:"annotations"`
	Bindings    []AnnotationBinding    `json:"bindings"`
	Classes     []*AnnotationClass     `json:"classes"`
	Diagnostics []AnnotationDiagnostic `json:"diagnostics"`
	Enums       []*AnnotationEnum      `json:"enums"`
	Functions   []*AnnotationFunction  `json:"functions"`
	Path        string                 `json:"path"`
}
type AnnotationTag struct {
	Kind  string `json:"kind"`
	Line  int    `json:"line"`
	Raw   string `json:"raw"`
	Value string `json:"value"`
}
type AnnotationParameter struct {
	Description string  `json:"description"`
	Line        int     `json:"line,omitempty"`
	Name        string  `json:"name"`
	Optional    bool    `json:"optional"`
	Raw         string  `json:"raw,omitempty"`
	Type        string  `json:"type"`
	Variadic    bool    `json:"variadic"`
	Visibility  *string `json:"visibility"`
}
type AnnotationReturn struct {
	Description *string `json:"description,omitempty"`
	Line        int     `json:"line,omitempty"`
	Name        *string `json:"name"`
	Raw         string  `json:"raw,omitempty"`
	Type        string  `json:"type"`
	Variadic    bool    `json:"variadic"`
	NeverSecret bool    `json:"-"`
}
type AnnotationFunction struct {
	Annotations   []AnnotationTag       `json:"annotations"`
	Arguments     []string              `json:"arguments"`
	Generics      []string              `json:"generics"`
	Line          int                   `json:"line"`
	Method        bool                  `json:"method"`
	Name          string                `json:"name"`
	Overloads     []string              `json:"overloads"`
	Owner         *string               `json:"owner"`
	Params        []AnnotationParameter `json:"params"`
	QualifiedName string                `json:"qualified_name"`
	Raw           string                `json:"raw"`
	RawOwner      *string               `json:"raw_owner"`
	Returns       []AnnotationReturn    `json:"returns"`
}
type AnnotationClass struct {
	Attributes *string               `json:"attributes"`
	Bases      []string              `json:"bases"`
	Binding    *string               `json:"binding"`
	Fields     []AnnotationParameter `json:"fields"`
	Line       int                   `json:"line"`
	Methods    []*AnnotationFunction `json:"methods"`
	Name       string                `json:"name"`
	Raw        string                `json:"raw"`
}
type AnnotationAlias struct {
	Alternatives []AnnotationAlternative `json:"alternatives"`
	Description  string                  `json:"description"`
	Line         int                     `json:"line"`
	Name         string                  `json:"name"`
	Raw          string                  `json:"raw"`
	Type         string                  `json:"type"`
}
type AnnotationAlternative struct {
	Default     bool   `json:"default"`
	Description string `json:"description"`
	Line        int    `json:"line"`
	Raw         string `json:"raw"`
	Type        string `json:"type"`
}
type AnnotationEnum struct {
	Binding *string               `json:"binding"`
	Line    int                   `json:"line"`
	Name    string                `json:"name"`
	Raw     string                `json:"raw"`
	Values  []AnnotationEnumValue `json:"values"`
}
type AnnotationEnumValue struct {
	Line     int                `json:"line"`
	Name     string             `json:"name"`
	Raw      string             `json:"raw"`
	RawValue string             `json:"raw_value"`
	Value    AnnotationConstant `json:"value"`
}
type AnnotationBinding struct {
	Class *string `json:"class"`
	Line  int     `json:"line"`
	Local bool    `json:"local"`
	Name  string  `json:"name"`
	Raw   string  `json:"raw"`
	Value string  `json:"value"`
}
type AnnotationDiagnostic struct {
	Code    string `json:"code"`
	Line    int    `json:"line"`
	Message string `json:"message"`
	Raw     string `json:"raw"`
	Source  string `json:"source,omitempty"`
}

type constantKind uint8

const (
	constantNull constantKind = iota
	constantBoolean
	constantInteger
	constantFloat
	constantString
)

// AnnotationConstant is the bounded scalar value of an annotation enum member.
// Uninterpreted enum expressions remain their source text.
type AnnotationConstant struct {
	kind    constantKind
	boolean bool
	integer int64
	number  literal.Float
	text    string
}

func (value AnnotationConstant) numeric() bool {
	return value.kind == constantBoolean || value.kind == constantInteger || value.kind == constantFloat
}
func (value AnnotationConstant) MarshalJSON() ([]byte, error) {
	switch value.kind {
	case constantNull:
		return []byte("null"), nil
	case constantBoolean:
		return json.Marshal(value.boolean)
	case constantInteger:
		return json.Marshal(value.integer)
	case constantFloat:
		return value.number.MarshalJSON()
	case constantString:
		return json.Marshal(value.text)
	default:
		return nil, fmt.Errorf("invalid annotation constant")
	}
}
func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
func stringPointer(value string) *string { return &value }
func boolPointer(value bool) *bool       { return &value }
