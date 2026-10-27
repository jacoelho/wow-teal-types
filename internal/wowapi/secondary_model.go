package wowapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go.yaml.in/yaml/v3"
	"io"
	"math"
	"strconv"
	"strings"
	"wow-teal-types/internal/wowapi/literal"
)

type secondaryCarrier struct {
	Primitive   string
	Constructor string
	Name        string
	Element     *secondaryCarrier
}

func (t secondaryCarrier) MarshalJSON() ([]byte, error) {
	if t.Primitive != "" {
		return json.Marshal(t.Primitive)
	}
	if t.Constructor == "arrayof" {
		return json.Marshal(struct {
			ArrayOf *secondaryCarrier `json:"arrayof"`
		}{t.Element})
	}
	return json.Marshal(map[string]string{t.Constructor: t.Name})
}

type defaultLiteral struct {
	Text      *string
	Boolean   *bool
	Number    *literal.Numeric
	Reference string
}

func (d defaultLiteral) MarshalJSON() ([]byte, error) {
	switch {
	case d.Text != nil:
		return json.Marshal(*d.Text)
	case d.Boolean != nil:
		return json.Marshal(*d.Boolean)
	case d.Number != nil:
		return json.Marshal(*d.Number)
	case d.Reference != "":
		return json.Marshal(struct {
			Reference string `json:"$ref"`
		}{d.Reference})
	default:
		return []byte("null"), nil
	}
}

type parameterDefault struct {
	Present bool            `json:"present"`
	Value   *defaultLiteral `json:"value,omitempty"`
}
type secondaryParameterFact struct {
	Default parameterDefault `json:"default"`
	Name    string           `json:"name"`
	Nilable bool             `json:"nilable"`
	Stride  *int             `json:"stride"`
	Type    secondaryCarrier `json:"type"`
}
type SecondaryParameter struct {
	Name    string
	Type    secondaryCarrier
	Nilable bool
	Default parameterDefault
	Stride  *int
}
type SecondaryMethod struct {
	Inputs, Outputs  []SecondaryParameter
	MayReturnNothing bool
	Raw              literal.Object
}

func (m SecondaryMethod) MarshalJSON() ([]byte, error) { return json.Marshal(m.Raw) }

type SecondaryReceiver struct {
	Methods  map[string]SecondaryMethod
	Inherits []string
}
type secondaryTypedef struct{ Type secondaryCarrier }
type secondaryComparedValue struct {
	Text    *string
	Boolean *bool
	Integer *int
	Type    *secondaryCarrier
	Default *parameterDefault
}

func (v secondaryComparedValue) MarshalJSON() ([]byte, error) {
	switch {
	case v.Text != nil:
		return json.Marshal(*v.Text)
	case v.Boolean != nil:
		return json.Marshal(*v.Boolean)
	case v.Integer != nil:
		return json.Marshal(*v.Integer)
	case v.Type != nil:
		return json.Marshal(*v.Type)
	case v.Default != nil:
		return json.Marshal(*v.Default)
	default:
		return []byte("null"), nil
	}
}

type secondaryDifference struct {
	Kind      string                 `json:"kind"`
	Native    secondaryComparedValue `json:"native"`
	Position  int                    `json:"position,omitempty"`
	Secondary secondaryComparedValue `json:"secondary"`
}
type secondaryTuple struct {
	Native    []secondaryParameterFact `json:"native"`
	Secondary []secondaryParameterFact `json:"secondary"`
}
type secondaryTuples struct {
	Arguments secondaryTuple `json:"arguments"`
	Returns   secondaryTuple `json:"returns"`
}
type secondaryComparison struct {
	Differences []secondaryDifference `json:"differences"`
	Tuples      secondaryTuples       `json:"tuples"`
}
type hierarchyDifference struct {
	Disposition string   `json:"disposition"`
	Kind        string   `json:"kind"`
	Native      []string `json:"native"`
	Secondary   []string `json:"secondary"`
}
type capabilityConfirmation struct {
	AllMethodsObservedPresent bool     `json:"all_methods_observed_present"`
	Build                     string   `json:"build"`
	MethodSetsMatch           bool     `json:"method_sets_match"`
	Scope                     string   `json:"scope"`
	SecondaryMethods          []string `json:"secondary_methods"`
	WorkspaceMethods          []string `json:"workspace_methods"`
}
type receiverHierarchy struct {
	Differences                  []hierarchyDifference   `json:"differences"`
	NativeCapabilityConfirmation *capabilityConfirmation `json:"native_capability_confirmation,omitempty"`
	SecondaryBases               []string                `json:"secondary_bases"`
	SecondaryNamespace           string                  `json:"secondary_namespace"`
	WorkspaceBases               []string                `json:"workspace_bases"`
}

// safeYAML preserves syntax only at the evidence boundary. Domain decoders below
// interpret specific schemas rather than passing untyped YAML into comparisons.
func safeYAML(data []byte) (*yaml.Node, error) {
	if len(data) > 1<<20 {
		return nil, fmt.Errorf("YAML evidence exceeds 1 MiB")
	}
	var document yaml.Node
	dec := yaml.NewDecoder(bytes.NewReader(data))
	if e := dec.Decode(&document); e != nil {
		return nil, fmt.Errorf("invalid pinned YAML evidence: %w", e)
	}
	var extra yaml.Node
	if e := dec.Decode(&extra); e != io.EOF {
		return nil, fmt.Errorf("multiple YAML evidence documents are unsupported")
	}
	var validate func(*yaml.Node, int) error
	validate = func(n *yaml.Node, depth int) error {
		if depth > 48 {
			return fmt.Errorf("YAML evidence nesting exceeds 48 levels")
		}
		if n.Alias != nil || n.Kind == yaml.AliasNode {
			return fmt.Errorf("YAML aliases are not supported in pinned evidence")
		}
		switch n.Kind {
		case yaml.DocumentNode:
			if len(n.Content) != 1 {
				return fmt.Errorf("invalid YAML evidence document")
			}
		case yaml.MappingNode:
			seen := map[string]bool{}
			for i := 0; i < len(n.Content); i += 2 {
				k := n.Content[i]
				if k.Kind != yaml.ScalarNode || (k.Tag != "!!str" && k.Tag != "!!int") {
					return fmt.Errorf("YAML evidence mapping key must be a string or integer at line %d", k.Line)
				}
				key := k.Tag + ":" + k.Value
				if seen[key] {
					return fmt.Errorf("duplicate YAML evidence key: %s", k.Value)
				}
				seen[key] = true
			}
		case yaml.SequenceNode:
		case yaml.ScalarNode:
			if _, e := yamlScalar(n); e != nil {
				return e
			}
		default:
			return fmt.Errorf("unsupported YAML evidence node")
		}
		for _, child := range n.Content {
			if e := validate(child, depth+1); e != nil {
				return e
			}
		}
		return nil
	}
	if e := validate(&document, 0); e != nil {
		return nil, e
	}
	return document.Content[0], nil
}
func yamlScalar(n *yaml.Node) (literal.Value, error) {
	if n.Style == 0 && n.Tag == "!!str" {
		switch n.Value {
		case "yes", "Yes", "YES", "on", "On", "ON":
			return literal.Bool(true), nil
		case "no", "No", "NO", "off", "Off", "OFF":
			return literal.Bool(false), nil
		}
	}
	switch n.Tag {
	case "!!str":
		return literal.Text(n.Value), nil
	case "!!null":
		return literal.Value{}, nil
	case "!!bool":
		switch strings.ToLower(n.Value) {
		case "true", "yes", "on":
			return literal.Bool(true), nil
		case "false", "no", "off":
			return literal.Bool(false), nil
		}
	case "!!int":
		var v int64
		if e := n.Decode(&v); e != nil {
			return literal.Value{}, e
		}
		return literal.Int(v), nil
	case "!!float":
		var v float64
		if e := n.Decode(&v); e != nil {
			return literal.Value{}, e
		}
		if !math.IsNaN(v) && !math.IsInf(v, 0) {
			return literal.Number(v)
		}
	}
	return literal.Value{}, fmt.Errorf("unsupported YAML evidence scalar: %s", n.Tag)
}
func yamlRaw(n *yaml.Node) (literal.Value, error) {
	if n.Kind == yaml.ScalarNode {
		return yamlScalar(n)
	}
	if n.Kind == yaml.SequenceNode {
		values := []literal.Value{}
		for _, child := range n.Content {
			v, e := yamlRaw(child)
			if e != nil {
				return literal.Value{}, e
			}
			values = append(values, v)
		}
		return literal.List(values), nil
	}
	if n.Kind == yaml.MappingNode {
		values := literal.Object{}
		for i := 0; i < len(n.Content); i += 2 {
			if n.Content[i].Tag != "!!str" {
				return literal.Value{}, fmt.Errorf("raw evidence object requires string keys")
			}
			v, e := yamlRaw(n.Content[i+1])
			if e != nil {
				return literal.Value{}, e
			}
			values[n.Content[i].Value] = v
		}
		return literal.Record(values), nil
	}
	return literal.Value{}, fmt.Errorf("unsupported YAML evidence node")
}
func yamlMapping(n *yaml.Node, allowed []string, context string) (map[string]*yaml.Node, error) {
	if n == nil || n.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("unsupported %s schema", context)
	}
	m := map[string]*yaml.Node{}
	for i := 0; i < len(n.Content); i += 2 {
		k := n.Content[i]
		if k.Tag != "!!str" || (allowed != nil && !evContains(allowed, k.Value)) {
			return nil, fmt.Errorf("unsupported %s schema: %s", context, k.Value)
		}
		m[k.Value] = n.Content[i+1]
	}
	return m, nil
}
func yamlText(n *yaml.Node) (string, error) {
	if n == nil || n.Kind != yaml.ScalarNode {
		return "", fmt.Errorf("expected YAML string")
	}
	v, e := yamlScalar(n)
	if e != nil {
		return "", e
	}
	s, ok := v.Text()
	if !ok {
		return "", fmt.Errorf("expected YAML string")
	}
	return s, nil
}
func yamlBool(n *yaml.Node) (bool, error) {
	if n == nil {
		return false, nil
	}
	v, e := yamlScalar(n)
	if e != nil {
		return false, e
	}
	b, ok := v.Bool()
	if !ok {
		return false, fmt.Errorf("expected YAML Boolean")
	}
	return b, nil
}
func yamlStride(n *yaml.Node) (*int, error) {
	if n == nil || n.Tag == "!!null" {
		return nil, nil
	}
	if n.Tag != "!!int" {
		return nil, fmt.Errorf("stride must be a positive integer")
	}
	var v int
	if e := n.Decode(&v); e != nil || v <= 0 {
		return nil, fmt.Errorf("stride must be a positive integer")
	}
	return &v, nil
}
func parseSecondaryType(n *yaml.Node) (secondaryCarrier, error) {
	if n != nil && n.Kind == yaml.ScalarNode {
		s, e := yamlText(n)
		if e != nil {
			return secondaryCarrier{}, e
		}
		if !evContains([]string{"number", "string", "boolean", "function", "table", "any", "FileAsset"}, s) {
			return secondaryCarrier{}, fmt.Errorf("unsupported secondary primitive: %s", s)
		}
		return secondaryCarrier{Primitive: s}, nil
	}
	m, e := yamlMapping(n, []string{"arrayof", "enum", "luaobject", "uiobject", "structure", "stringenum"}, "secondary type")
	if e != nil {
		return secondaryCarrier{}, e
	}
	if len(m) != 1 {
		return secondaryCarrier{}, fmt.Errorf("secondary type requires one constructor")
	}
	for key, inner := range m {
		if key == "arrayof" {
			value, e := parseSecondaryType(inner)
			return secondaryCarrier{Constructor: key, Element: &value}, e
		}
		name, e := yamlText(inner)
		if e != nil || name == "" {
			return secondaryCarrier{}, fmt.Errorf("secondary nominal type requires a name")
		}
		return secondaryCarrier{Constructor: key, Name: name}, nil
	}
	panic("unreachable carrier")
}
func parseDefault(value literal.Value) (defaultLiteral, error) {
	switch value.Kind() {
	case literal.Null:
		return defaultLiteral{}, nil
	case literal.String:
		text, _ := value.Text()
		return defaultLiteral{Text: &text}, nil
	case literal.Boolean:
		b, _ := value.Bool()
		return defaultLiteral{Boolean: &b}, nil
	case literal.Integer, literal.FloatKind:
		b, e := json.Marshal(value)
		if e != nil {
			return defaultLiteral{}, e
		}
		var n literal.Numeric
		if e = json.Unmarshal(b, &n); e != nil {
			return defaultLiteral{}, e
		}
		return defaultLiteral{Number: &n}, nil
	case literal.Reference:
		name, _ := value.Reference()
		return defaultLiteral{Reference: name}, nil
	case literal.ObjectKind:
		o, _ := value.Object()
		if len(o) == 1 {
			if name, ok := o["$ref"].Text(); ok && name != "" {
				return defaultLiteral{Reference: name}, nil
			}
		}
	}
	return defaultLiteral{}, fmt.Errorf("unsupported default literal or enum reference")
}
func parseSecondaryParameter(n *yaml.Node) (SecondaryParameter, error) {
	var p SecondaryParameter
	m, e := yamlMapping(n, []string{"name", "type", "nilable", "default", "stride"}, "secondary parameter")
	if e != nil {
		return p, e
	}
	p.Name, e = yamlText(m["name"])
	if e != nil || p.Name == "" {
		return p, fmt.Errorf("secondary parameter requires name and type")
	}
	p.Type, e = parseSecondaryType(m["type"])
	if e != nil {
		return p, e
	}
	p.Nilable, e = yamlBool(m["nilable"])
	if e != nil {
		return p, e
	}
	p.Stride, e = yamlStride(m["stride"])
	if e != nil {
		return p, e
	}
	if n, ok := m["default"]; ok {
		raw, e := yamlRaw(n)
		if e != nil {
			return p, e
		}
		v, e := parseDefault(raw)
		if e != nil {
			return p, e
		}
		p.Default = parameterDefault{Present: true, Value: &v}
	}
	return p, nil
}
func parseSecondaryMethod(n *yaml.Node) (SecondaryMethod, error) {
	var method SecondaryMethod
	m, e := yamlMapping(n, []string{"inputs", "outputs", "mayreturnnothing"}, "secondary method")
	if e != nil {
		return method, e
	}
	for _, pair := range []struct {
		key    string
		target *[]SecondaryParameter
	}{{"inputs", &method.Inputs}, {"outputs", &method.Outputs}} {
		node, ok := m[pair.key]
		if !ok {
			return method, fmt.Errorf("secondary method requires both input and output tuples")
		}
		if node.Tag == "!!null" {
			continue
		}
		if node.Kind != yaml.SequenceNode {
			return method, fmt.Errorf("secondary tuple must be a list or null")
		}
		*pair.target = []SecondaryParameter{}
		for _, child := range node.Content {
			v, e := parseSecondaryParameter(child)
			if e != nil {
				return method, e
			}
			*pair.target = append(*pair.target, v)
		}
	}
	method.MayReturnNothing, e = yamlBool(m["mayreturnnothing"])
	if e != nil {
		return method, e
	}
	raw, e := yamlRaw(n)
	if e != nil {
		return method, e
	}
	method.Raw, _ = raw.Object()
	return method, nil
}
func parseSecondaryReceivers(data []byte) (map[string]SecondaryReceiver, error) {
	root, e := safeYAML(data)
	if e != nil {
		return nil, e
	}
	receivers, e := yamlMapping(root, nil, "secondary receivers")
	if e != nil {
		return nil, e
	}
	out := map[string]SecondaryReceiver{}
	for name, node := range receivers {
		selected := false
		for _, spec := range nativeFamilies {
			selected = selected || spec.owner == name
		}
		var allowed []string
		if selected {
			allowed = []string{"methods", "inherits", "fields", "virtual"}
		}
		m, e := yamlMapping(node, allowed, "secondary receiver")
		if e != nil {
			return nil, e
		}
		receiver := SecondaryReceiver{Methods: map[string]SecondaryMethod{}, Inherits: []string{}}
		if methods := m["methods"]; methods != nil && methods.Tag != "!!null" {
			rows, e := yamlMapping(methods, nil, "secondary methods")
			if e != nil {
				return nil, e
			}
			for key, node := range rows {
				if !selected {
					receiver.Methods[key] = SecondaryMethod{}
					continue
				}
				method, e := parseSecondaryMethod(node)
				if e != nil {
					return nil, fmt.Errorf("secondary %s.%s: %w", name, key, e)
				}
				receiver.Methods[key] = method
			}
		}
		if parent := m["inherits"]; parent != nil && parent.Tag != "!!null" {
			if parent.Kind == yaml.ScalarNode {
				value, e := yamlText(parent)
				if e != nil {
					return nil, e
				}
				receiver.Inherits = append(receiver.Inherits, value)
			} else {
				parents, e := yamlMapping(parent, nil, "secondary inheritance")
				if e != nil {
					return nil, e
				}
				for _, key := range sortedKeys(parents) {
					if parents[key].Tag != "!!null" {
						return nil, fmt.Errorf("unsupported secondary receiver inheritance")
					}
					receiver.Inherits = append(receiver.Inherits, key)
				}
			}
		}
		out[name] = receiver
	}
	return out, nil
}
func parseSecondaryTypedefs(data []byte, needed map[string]bool) (map[string]secondaryTypedef, error) {
	root, e := safeYAML(data)
	if e != nil {
		return nil, e
	}
	doc, e := yamlMapping(root, nil, "secondary docs")
	if e != nil {
		return nil, e
	}
	rows, e := yamlMapping(doc["typedefs"], nil, "secondary typedefs")
	if e != nil {
		return nil, e
	}
	out := map[string]secondaryTypedef{}
	for key, node := range rows {
		if !needed[key] {
			continue
		}
		fields, e := yamlMapping(node, nil, "secondary typedef")
		if e != nil {
			return nil, e
		}
		typ, e := parseSecondaryType(fields["type"])
		if e != nil {
			return nil, e
		}
		out[key] = secondaryTypedef{Type: typ}
	}
	return out, nil
}
func secondaryBuild(data []byte) (string, error) {
	root, e := safeYAML(data)
	if e != nil {
		return "", e
	}
	m, e := yamlMapping(root, nil, "secondary build")
	if e != nil {
		return "", e
	}
	version, e := yamlText(m["version"])
	if e != nil {
		return "", e
	}
	build := m["build"]
	if build == nil || build.Kind != yaml.ScalarNode || (build.Tag != "!!int" && build.Tag != "!!str") {
		return "", fmt.Errorf("secondary build must be a decimal integer or string")
	}
	var value int
	if build.Tag == "!!str" {
		value, e = strconv.Atoi(build.Value)
		if e != nil || strconv.Itoa(value) != build.Value {
			return "", fmt.Errorf("secondary build must be a decimal integer string")
		}
	} else if e = build.Decode(&value); e != nil {
		return "", fmt.Errorf("secondary build integer: %w", e)
	}
	if value <= 0 {
		return "", fmt.Errorf("secondary build must be positive integer")
	}
	return version + "." + strconv.Itoa(value), nil
}
