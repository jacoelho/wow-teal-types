package wowapi

type AnnotationDocument struct {
	Aliases   []*AnnotationAlias    `json:"aliases"`
	Bindings  []AnnotationBinding   `json:"bindings"`
	Classes   []*AnnotationClass    `json:"classes"`
	Enums     []*AnnotationEnum     `json:"enums"`
	Functions []*AnnotationFunction `json:"functions"`
	Path      string                `json:"path"`
}
type AnnotationParameter struct {
	Name     string `json:"name"`
	Optional bool   `json:"optional"`
	Type     string `json:"type"`
	Variadic bool   `json:"variadic"`
}
type AnnotationReturn struct {
	Type        string `json:"type"`
	Variadic    bool   `json:"variadic"`
	NeverSecret bool   `json:"-"`
}
type AnnotationFunction struct {
	Arguments     []string              `json:"arguments"`
	Generics      []string              `json:"generics"`
	Method        bool                  `json:"method"`
	Name          string                `json:"name"`
	Overloads     []string              `json:"overloads"`
	Owner         *string               `json:"owner"`
	Params        []AnnotationParameter `json:"params"`
	QualifiedName string                `json:"qualified_name"`
	Local         bool                  `json:"local"`
	Returns       []AnnotationReturn    `json:"returns"`
}
type AnnotationClass struct {
	Bases  []string              `json:"bases"`
	Fields []AnnotationParameter `json:"fields"`
	Name   string                `json:"name"`
}
type AnnotationAlias struct {
	Alternatives []string `json:"alternatives"`
	Line         int      `json:"line"`
	Name         string   `json:"name"`
	Type         string   `json:"type"`
}
type AnnotationEnum struct {
	Name   string                `json:"name"`
	Values []AnnotationEnumValue `json:"values"`
}
type AnnotationEnumValue struct {
	Name    string `json:"name"`
	Numeric bool   `json:"numeric"`
}
type AnnotationBinding struct {
	Class    *string `json:"class"`
	Local    bool    `json:"local"`
	Name     string  `json:"name"`
	TopLevel bool    `json:"top_level"`
	Value    string  `json:"value"`
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
func stringPointer(value string) *string { return &value }
func boolPointer(value bool) *bool       { return &value }
