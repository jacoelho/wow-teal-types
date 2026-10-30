package wowapi

import (
	"maps"
	"slices"
)

// definitionKind is the declaration form of a catalogue definition. The zero
// value marks an absent definition (see kindOf).
type definitionKind uint8

const (
	kindAbsent definitionKind = iota
	kindEnum
	kindRecord
	kindInterface
	// kindConvertedRecord is an interned anonymous LuaLS table whose field
	// types are already Teal expressions.
	kindConvertedRecord
)

var definitionKindNames = [...]string{"absent", "enum", "record", "interface", "converted_record"}

func (k definitionKind) String() string { return definitionKindNames[k] }

// isTableLike reports whether values of the kind are Lua tables with fields.
func (k definitionKind) isTableLike() bool {
	return k == kindRecord || k == kindInterface || k == kindConvertedRecord
}

func kindOf(v *catalogueDefinition) definitionKind {
	if v == nil {
		return kindAbsent
	}
	return v.Kind
}

type catalogueDefinition struct {
	Kind                               definitionKind
	Source                             string
	Fields                             map[string]*catalogueField
	Methods                            map[string]*catalogueFunction
	Bases                              []string
	Opaque, NativeHandle, NativeObject bool
	Values                             []string
}
type catalogueAlias struct {
	Expression, Source string
	Signature          *catalogueFunction
}
type catalogueField struct {
	Type     string
	Resolved *string
	Native   *nativeParameter
}
type catalogueFunction struct {
	PolicySymbol      string
	Native            *nativeFunction
	Annotation        *annotationFunction
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
		native.Policy.Flags = maps.Clone(native.Policy.Flags)
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
		copy.Annotation = &annotation
	}
	copy.Variants = make([]*catalogueFunction, len(function.Variants))
	for i, variant := range function.Variants {
		copy.Variants[i] = variant.clone()
	}
	copy.CallableContract = function.CallableContract.clone()
	return &copy
}
func functionFromNative(native nativeFunction, source string, method bool) *catalogueFunction {
	return (&catalogueFunction{Native: &native, Source: source, Method: method}).clone()
}
func functionFromAnnotation(annotation *annotationFunction, source string) *catalogueFunction {
	return (&catalogueFunction{Annotation: annotation, Source: source, Method: annotation.Method}).clone()
}
