package wowapi

import (
	"encoding/json"
	"fmt"
	"wow-teal-types/internal/wowapi/literal"
)

// differenceValue is the closed set of facts compared by the audit. Raw retains
// native source evidence only; interpreted comparisons use the specific fields.
type differenceValue struct {
	text        *string
	boolean     *bool
	count       *int
	texts       *[]string
	positions   *[]int
	fact        *typeFact
	scope       *policyScope
	raw         literal.Object
	defect      *sourceDefect
	unsupported *[]unsupportedAnnotation
	secondary   *secondaryDifference
	hierarchy   *hierarchyDifference
}

func textDifference(v string) differenceValue    { return differenceValue{text: &v} }
func boolDifference(v bool) differenceValue      { return differenceValue{boolean: &v} }
func countDifference(v int) differenceValue      { return differenceValue{count: &v} }
func textsDifference(v []string) differenceValue { return differenceValue{texts: &v} }
func (v differenceValue) MarshalJSON() ([]byte, error) {
	switch {
	case v.text != nil:
		return json.Marshal(*v.text)
	case v.boolean != nil:
		return json.Marshal(*v.boolean)
	case v.count != nil:
		return json.Marshal(*v.count)
	case v.texts != nil:
		return json.Marshal(*v.texts)
	case v.positions != nil:
		return json.Marshal(*v.positions)
	case v.fact != nil:
		return json.Marshal(*v.fact)
	case v.scope != nil:
		return json.Marshal(*v.scope)
	case v.raw != nil:
		return json.Marshal(v.raw)
	case v.defect != nil:
		return json.Marshal(*v.defect)
	case v.unsupported != nil:
		return json.Marshal(*v.unsupported)
	case v.secondary != nil:
		return json.Marshal(*v.secondary)
	case v.hierarchy != nil:
		return json.Marshal(*v.hierarchy)
	default:
		return []byte("null"), nil
	}
}

type comparedValues struct {
	Native differenceValue `json:"native"`
	TSM    differenceValue `json:"tsm"`
}
type contractDifference struct {
	*comparedValues
	Basis                  string   `json:"basis,omitempty"`
	Disposition            string   `json:"disposition,omitempty"`
	Features               []string `json:"features,omitzero"`
	Kind                   string   `json:"kind"`
	Position               int      `json:"position,omitempty"`
	Reason                 string   `json:"reason,omitempty"`
	Side                   string   `json:"side,omitempty"`
	TSMSecretWrapper       *bool    `json:"tsm_secret_wrapper,omitempty"`
	Variant                string   `json:"variant,omitempty"`
	WorkspaceOpaqueCarrier *bool    `json:"workspace_opaque_carrier,omitempty"`
}

func difference(kind string, native, tsm differenceValue, position int) contractDifference {
	return contractDifference{Kind: kind, Position: position, comparedValues: &comparedValues{Native: native, TSM: tsm}}
}
func evMustFact(f simpleFact, err error) simpleFact {
	if err != nil {
		factFail(err.Error())
	}
	return f
}
func recoverFact(err *error) {
	if v := recover(); v != nil {
		if failure, ok := v.(factFailure); ok {
			*err = fmt.Errorf("%s", failure.message)
		} else {
			panic(v)
		}
	}
}

type opacityDifference struct {
	Disposition            string `json:"disposition"`
	Position               int    `json:"position"`
	TSMSecretWrapper       bool   `json:"tsm_secret_wrapper"`
	WorkspaceOpaqueCarrier bool   `json:"workspace_opaque_carrier"`
}
type uncheckedReturn struct {
	Position int    `json:"position"`
	Reason   string `json:"reason"`
}
type effectiveDeclaration struct {
	EffectiveMayReturnNothing *bool                `json:"effective_may_return_nothing,omitempty"`
	EmittedSignatures         []string             `json:"emitted_signatures"`
	NativeAudit               bool                 `json:"native_audit,omitempty"`
	OpacityDifferences        []opacityDifference  `json:"opacity_differences,omitzero"`
	ProjectionNotes           []RepresentationNote `json:"projection_notes,omitzero"`
	ReviewedSecretEffect      *bool                `json:"reviewed_secret_effect,omitempty"`
	SecretDataReturns         []string             `json:"secret_data_returns,omitzero"`
	Symbol                    string               `json:"symbol"`
	TSMBranch                 *string              `json:"tsm_branch,omitempty"`
	TSMLine                   int                  `json:"tsm_line,omitempty"`
	TSMSecretReturns          []bool               `json:"tsm_secret_returns,omitzero"`
	TSMSource                 string               `json:"tsm_source,omitempty"`
	UncheckedReturnPositions  []uncheckedReturn    `json:"unchecked_return_positions,omitzero"`
	WorkspaceResultSecret     bool                 `json:"workspace_result_secret"`
}

var nativeSecretPredicateNames = []string{"SecretInActivePvPMatch", "SecretInChatMessagingLockdown", "SecretWhenAnchoringSecret", "SecretWhenAurasRestricted", "SecretWhenCooldownsRestricted", "SecretWhenCurveSecret", "SecretWhenEncounterEvent", "SecretWhenInCombat", "SecretWhenLossOfControlInfoRestricted", "SecretWhenLuaTableHasSecretKeys", "SecretWhenNumericFormatterSecret", "SecretWhenTotemSlotSecret", "SecretWhenUnitAuraRestricted", "SecretWhenUnitComparisonRestricted", "SecretWhenUnitHealthMaxRestricted", "SecretWhenUnitIdentityRestricted", "SecretWhenUnitNameIdentityRestricted", "SecretWhenUnitPossessionRestricted", "SecretWhenUnitPowerMaxRestricted", "SecretWhenUnitPowerRestricted", "SecretWhenUnitSpellCastRestricted", "SecretWhenUnitStatsRestricted", "SecretWhenUnitThreatStateRestricted", "SecretWhenUnitThreatValuesRestricted"}
