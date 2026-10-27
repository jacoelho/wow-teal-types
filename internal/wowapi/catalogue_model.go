package wowapi

import "slices"

type catalogueDefinition struct {
	Kind, Source                       string
	Fields                             map[string]*catalogueField
	Methods                            map[string]*catalogueFunction
	Bases                              []string
	Opaque, NativeHandle, NativeObject bool
	Values                             []string
	Signature                          string
	Generics                           []string
	InheritedFields                    map[string]bool
}
type catalogueAlias struct {
	Expression, Source string
	Signature          *catalogueFunction
	Line               int
}
type catalogueField struct {
	Type, Source string
	Resolved     *string
	Native       *NativeParameter
}
type catalogueFunction struct {
	PolicySymbol      string
	Native            *NativeFunction
	Annotation        *AnnotationFunction
	Source            string
	Method            bool
	ResolvedSignature *string
	Variants          []*catalogueFunction
	TypeName          string
	CallableContract  *catalogueFunction
	SecretEffect      bool
}

func (function *catalogueFunction) contract() *catalogueFunction {
	if function.CallableContract != nil {
		return function.CallableContract
	}
	return function
}
func (function *catalogueFunction) clone() *catalogueFunction {
	if function == nil {
		return nil
	}
	copy := *function
	if function.Native != nil {
		native := *function.Native
		native.Arguments = slices.Clone(native.Arguments)
		native.Returns = slices.Clone(native.Returns)
		copy.Native = &native
	}
	if function.Annotation != nil {
		annotation := *function.Annotation
		annotation.Params = slices.Clone(annotation.Params)
		annotation.Returns = slices.Clone(annotation.Returns)
		annotation.Generics = slices.Clone(annotation.Generics)
		annotation.Arguments = slices.Clone(annotation.Arguments)
		annotation.Overloads = slices.Clone(annotation.Overloads)
		annotation.Annotations = slices.Clone(annotation.Annotations)
		copy.Annotation = &annotation
	}
	copy.Variants = make([]*catalogueFunction, len(function.Variants))
	for i, variant := range function.Variants {
		copy.Variants[i] = variant.clone()
	}
	copy.CallableContract = function.CallableContract.clone()
	return &copy
}
func functionFromNative(native NativeFunction, source string, method bool) *catalogueFunction {
	return (&catalogueFunction{Native: &native, Source: source, Method: method}).clone()
}
func functionFromAnnotation(annotation *AnnotationFunction, source string) *catalogueFunction {
	return (&catalogueFunction{Annotation: annotation, Source: source, Method: annotation.Method}).clone()
}
