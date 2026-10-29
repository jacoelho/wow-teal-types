package wowapi

type annotationDocument struct {
	Aliases   []*annotationAlias    `json:"aliases"`
	Bindings  []annotationBinding   `json:"bindings"`
	Classes   []*annotationClass    `json:"classes"`
	Enums     []*annotationEnum     `json:"enums"`
	Functions []*annotationFunction `json:"functions"`
	Path      string                `json:"path"`
}
type annotationParameter struct {
	Name     string `json:"name"`
	Optional bool   `json:"optional"`
	Type     string `json:"type"`
	Variadic bool   `json:"variadic"`
}
type annotationReturn struct {
	Type     string `json:"type"`
	Variadic bool   `json:"variadic"`
}
type annotationFunction struct {
	Arguments     []string              `json:"arguments"`
	Generics      []string              `json:"generics"`
	Method        bool                  `json:"method"`
	Name          string                `json:"name"`
	Overloads     []string              `json:"overloads"`
	Owner         *string               `json:"owner"`
	Params        []annotationParameter `json:"params"`
	QualifiedName string                `json:"qualified_name"`
	Local         bool                  `json:"local"`
	Returns       []annotationReturn    `json:"returns"`
}
type annotationClass struct {
	Bases  []string              `json:"bases"`
	Fields []annotationParameter `json:"fields"`
	Name   string                `json:"name"`
}
type annotationAlias struct {
	Alternatives []string `json:"alternatives"`
	Line         int      `json:"line"`
	Name         string   `json:"name"`
	Type         string   `json:"type"`
}
type annotationEnum struct {
	Name   string                `json:"name"`
	Values []annotationEnumValue `json:"values"`
}
type annotationEnumValue struct {
	Name    string `json:"name"`
	Numeric bool   `json:"numeric"`
}
type annotationBinding struct {
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
