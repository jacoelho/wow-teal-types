package wowapi

import (
	"fmt"
	"math/big"
	"slices"
	"strings"
	"wow-teal-types/internal/wowapi/literal"
)

// NativeDocument pairs a source name with its decoded contracts.
type NativeDocument struct {
	Name     string
	Metadata NativeMetadata
}
type NativeMetadata struct {
	Environment string
	Events      []NativeEvent
	Functions   []NativeFunction
	Name        string
	Namespace   string
	ObjectType  string
	Predicates  []NativePredicate
	Tables      []NativeTable
	Type        string
}
type NativeFunction struct {
	Name, Type, Namespace string
	Arguments, Returns    []NativeParameter
	Policy                NativePolicy
}
type NativeParameter struct {
	HasDefault                                                 bool
	Name, Type, InnerType, KeyType, Mixin                      string
	Nilable, NilableContents, NeverSecret, NeverSecretContents *bool
	ConditionalSecret, ConditionalSecretContents, SecretValue  *bool
	StrideIndex                                                *int
}
type NativeEvent struct {
	Name, Type, LiteralName string
	Payload                 []NativeParameter
	Policy                  NativePolicy
}
type NativeTable struct {
	EnumMetadata                       []string
	Policy                             NativePolicy
	Name, Type                         string
	Arguments, Fields, Values, Returns []NativeParameter
}
type NativePredicate struct {
	Name, Type, FailureMode string
}
type NativeReference struct{ Name string }
type NativeAspectArgument struct {
	Argument string
	Aspect   NativeReference
}

type NativePolicy struct {
	AddsForbiddenAspects                      []NativeAspectArgument `json:"AddsForbiddenAspects,omitzero"`
	ChecksForbiddenAspects                    []NativeAspectArgument `json:"ChecksForbiddenAspects,omitzero"`
	Environment                               *string                `json:"Environment,omitzero"`
	SecretArguments                           *string                `json:"SecretArguments,omitzero"`
	SecretArgumentsAddAspect                  []NativeReference      `json:"SecretArgumentsAddAspect,omitzero"`
	SecretReturnsForAspect                    []NativeReference      `json:"SecretReturnsForAspect,omitzero"`
	CallbackEvent                             *bool                  `json:"CallbackEvent,omitzero"`
	CheckAllowChangeParent                    *bool                  `json:"CheckAllowChangeParent,omitzero"`
	CheckAllowInheritForbiddenLayoutAspects   *bool                  `json:"CheckAllowInheritForbiddenLayoutAspects,omitzero"`
	ConstSecretAccessor                       *bool                  `json:"ConstSecretAccessor,omitzero"`
	HasRestrictions                           *bool                  `json:"HasRestrictions,omitzero"`
	IsProtectedFunction                       *bool                  `json:"IsProtectedFunction,omitzero"`
	MayReturnNothing                          *bool                  `json:"MayReturnNothing,omitzero"`
	MouseFocusValidForLimitedInput            *bool                  `json:"MouseFocusValidForLimitedInput,omitzero"`
	RequireNPERestricted                      *bool                  `json:"RequireNPERestricted,omitzero"`
	RequiresActiveCommentator                 *bool                  `json:"RequiresActiveCommentator,omitzero"`
	RequiresAssignableScript                  *bool                  `json:"RequiresAssignableScript,omitzero"`
	RequiresCanChangeHitTestPoints            *bool                  `json:"RequiresCanChangeHitTestPoints,omitzero"`
	RequiresClubsInitialized                  *bool                  `json:"RequiresClubsInitialized,omitzero"`
	RequiresCommentator                       *bool                  `json:"RequiresCommentator,omitzero"`
	RequiresComparableUnitTokens              *bool                  `json:"RequiresComparableUnitTokens,omitzero"`
	RequiresDeclassifiedUnitIdentity          *bool                  `json:"RequiresDeclassifiedUnitIdentity,omitzero"`
	RequiresFontStringTextAccess              *bool                  `json:"RequiresFontStringTextAccess,omitzero"`
	RequiresFriendList                        *bool                  `json:"RequiresFriendList,omitzero"`
	RequiresIndexInRange                      *bool                  `json:"RequiresIndexInRange,omitzero"`
	RequiresLimitedInput                      *bool                  `json:"RequiresLimitedInput,omitzero"`
	RequiresNonReadOnlyCVar                   *bool                  `json:"RequiresNonReadOnlyCVar,omitzero"`
	RequiresNonSecretAura                     *bool                  `json:"RequiresNonSecretAura,omitzero"`
	RequiresNonSecureCVar                     *bool                  `json:"RequiresNonSecureCVar,omitzero"`
	RequiresRecentAllies                      *bool                  `json:"RequiresRecentAllies,omitzero"`
	RequiresRestrictedAbbreviationBreakpoints *bool                  `json:"RequiresRestrictedAbbreviationBreakpoints,omitzero"`
	RequiresScriptObjectAlphaAccess           *bool                  `json:"RequiresScriptObjectAlphaAccess,omitzero"`
	RequiresScriptObjectDesaturationAccess    *bool                  `json:"RequiresScriptObjectDesaturationAccess,omitzero"`
	RequiresSpellDiminishUI                   *bool                  `json:"RequiresSpellDiminishUI,omitzero"`
	RequiresStatusBarDesaturationAccess       *bool                  `json:"RequiresStatusBarDesaturationAccess,omitzero"`
	RequiresSupportedScript                   *bool                  `json:"RequiresSupportedScript,omitzero"`
	RequiresTimedSignalMapAccess              *bool                  `json:"RequiresTimedSignalMapAccess,omitzero"`
	RequiresUnitAuraAccess                    *bool                  `json:"RequiresUnitAuraAccess,omitzero"`
	RequiresUnitIdentityAccess                *bool                  `json:"RequiresUnitIdentityAccess,omitzero"`
	RequiresValidAbbreviationBreakpoints      *bool                  `json:"RequiresValidAbbreviationBreakpoints,omitzero"`
	RequiresValidActionSlot                   *bool                  `json:"RequiresValidActionSlot,omitzero"`
	RequiresValidAndPublicCVar                *bool                  `json:"RequiresValidAndPublicCVar,omitzero"`
	RequiresValidFontAsset                    *bool                  `json:"RequiresValidFontAsset,omitzero"`
	RequiresValidFontHeight                   *bool                  `json:"RequiresValidFontHeight,omitzero"`
	RequiresValidInviteTarget                 *bool                  `json:"RequiresValidInviteTarget,omitzero"`
	RequiresValidTimelineEvent                *bool                  `json:"RequiresValidTimelineEvent,omitzero"`
	RequiresValidUnitAuraInstance             *bool                  `json:"RequiresValidUnitAuraInstance,omitzero"`
	RestrictedForMacroChatMessages            *bool                  `json:"RestrictedForMacroChatMessages,omitzero"`
	ReturnsNeverSecret                        *bool                  `json:"ReturnsNeverSecret,omitzero"`
	SecretInActivePvPMatch                    *bool                  `json:"SecretInActivePvPMatch,omitzero"`
	SecretInChatMessagingLockdown             *bool                  `json:"SecretInChatMessagingLockdown,omitzero"`
	SecretPayloads                            *bool                  `json:"SecretPayloads,omitzero"`
	SecretReturns                             *bool                  `json:"SecretReturns,omitzero"`
	SecretWhenAnchoringSecret                 *bool                  `json:"SecretWhenAnchoringSecret,omitzero"`
	SecretWhenAurasRestricted                 *bool                  `json:"SecretWhenAurasRestricted,omitzero"`
	SecretWhenCooldownsRestricted             *bool                  `json:"SecretWhenCooldownsRestricted,omitzero"`
	SecretWhenCurveSecret                     *bool                  `json:"SecretWhenCurveSecret,omitzero"`
	SecretWhenEncounterEvent                  *bool                  `json:"SecretWhenEncounterEvent,omitzero"`
	SecretWhenInCombat                        *bool                  `json:"SecretWhenInCombat,omitzero"`
	SecretWhenLossOfControlInfoRestricted     *bool                  `json:"SecretWhenLossOfControlInfoRestricted,omitzero"`
	SecretWhenLuaTableHasSecretKeys           *bool                  `json:"SecretWhenLuaTableHasSecretKeys,omitzero"`
	SecretWhenNumericFormatterSecret          *bool                  `json:"SecretWhenNumericFormatterSecret,omitzero"`
	SecretWhenTotemSlotSecret                 *bool                  `json:"SecretWhenTotemSlotSecret,omitzero"`
	SecretWhenUnitAuraRestricted              *bool                  `json:"SecretWhenUnitAuraRestricted,omitzero"`
	SecretWhenUnitComparisonRestricted        *bool                  `json:"SecretWhenUnitComparisonRestricted,omitzero"`
	SecretWhenUnitHealthMaxRestricted         *bool                  `json:"SecretWhenUnitHealthMaxRestricted,omitzero"`
	SecretWhenUnitIdentityRestricted          *bool                  `json:"SecretWhenUnitIdentityRestricted,omitzero"`
	SecretWhenUnitNameIdentityRestricted      *bool                  `json:"SecretWhenUnitNameIdentityRestricted,omitzero"`
	SecretWhenUnitPossessionRestricted        *bool                  `json:"SecretWhenUnitPossessionRestricted,omitzero"`
	SecretWhenUnitPowerMaxRestricted          *bool                  `json:"SecretWhenUnitPowerMaxRestricted,omitzero"`
	SecretWhenUnitPowerRestricted             *bool                  `json:"SecretWhenUnitPowerRestricted,omitzero"`
	SecretWhenUnitSpellCastRestricted         *bool                  `json:"SecretWhenUnitSpellCastRestricted,omitzero"`
	SecretWhenUnitStatsRestricted             *bool                  `json:"SecretWhenUnitStatsRestricted,omitzero"`
	SecretWhenUnitThreatStateRestricted       *bool                  `json:"SecretWhenUnitThreatStateRestricted,omitzero"`
	SecretWhenUnitThreatValuesRestricted      *bool                  `json:"SecretWhenUnitThreatValuesRestricted,omitzero"`
	SecureHooksAllowed                        *bool                  `json:"SecureHooksAllowed,omitzero"`
	SynchronousEvent                          *bool                  `json:"SynchronousEvent,omitzero"`
	TransliteratorAllowed                     *bool                  `json:"TransliteratorAllowed,omitzero"`
	UniqueEvent                               *bool                  `json:"UniqueEvent,omitzero"`
}

func enabled(value *bool) bool { return value != nil && *value }
func nativeText(raw literal.Object, name string) (string, error) {
	value, present := raw[name]
	if !present {
		return "", nil
	}
	text, ok := value.Text()
	if !ok {
		return "", fmt.Errorf("native %s must be a string", name)
	}
	return text, nil
}
func nativeEnum(raw literal.Object, name string, allowed ...string) (*string, error) {
	if _, present := raw[name]; !present {
		return nil, nil
	}
	text, err := nativeText(raw, name)
	if err != nil {
		return nil, err
	}
	if !slices.Contains(allowed, text) {
		return nil, fmt.Errorf("unsupported native %s value %q", name, text)
	}
	return &text, nil
}
func nativeKind(raw literal.Object, allowed ...string) (string, error) {
	value, err := nativeEnum(raw, "Type", allowed...)
	if err != nil {
		return "", err
	}
	if value == nil {
		return "", fmt.Errorf("native Type must be a nonempty string")
	}
	return *value, nil
}
func nativeEnvironment(raw literal.Object) (*string, error) {
	return nativeEnum(raw, "Environment", "All", "SecureOnly")
}
func nativeOptionalBool(raw literal.Object, name string) (*bool, error) {
	value, present := raw[name]
	if !present {
		return nil, nil
	}
	boolean, ok := value.Bool()
	if !ok {
		return nil, fmt.Errorf("native %s must be a boolean", name)
	}
	return &boolean, nil
}
func nativeOptionalInt(raw literal.Object, name string) (*int, error) {
	value, present := raw[name]
	if !present {
		return nil, nil
	}
	text, ok := value.Integer()
	if !ok {
		return nil, fmt.Errorf("native %s must be an integer", name)
	}
	number, ok := new(big.Int).SetString(text, 10)
	if !ok || !number.IsInt64() {
		return nil, fmt.Errorf("native %s exceeds integer range", name)
	}
	result := int(number.Int64())
	if int64(result) != number.Int64() {
		return nil, fmt.Errorf("native %s exceeds integer range", name)
	}
	return &result, nil
}
func validateNativeDocumentation(raw literal.Object) error {
	value, present := raw["Documentation"]
	if !present {
		return nil
	}
	array, ok := value.Array()
	if !ok {
		return fmt.Errorf("native Documentation must be an array")
	}
	for i, item := range array {
		if _, ok := item.Text(); !ok {
			return fmt.Errorf("native Documentation[%d] must be a string", i)
		}
	}
	return nil
}
func nativeItems[T any](raw literal.Object, name string, decode func(literal.Object) (T, error)) ([]T, error) {
	value, present := raw[name]
	if !present {
		return nil, nil
	}
	array, ok := value.Array()
	if !ok {
		return nil, fmt.Errorf("native %s must be an array", name)
	}
	values := make([]T, len(array))
	for i, item := range array {
		object, ok := item.Object()
		if !ok {
			return nil, fmt.Errorf("native %s[%d] must be an object", name, i)
		}
		value, err := decode(object)
		if err != nil {
			return nil, fmt.Errorf("native %s[%d]: %w", name, i, err)
		}
		values[i] = value
	}
	return values, nil
}
func decodeNativeReference(value literal.Value) (NativeReference, error) {
	if name, ok := value.Reference(); ok {
		return NativeReference{name}, nil
	}
	if object, ok := value.Object(); ok && len(object) == 1 {
		if name, ok := object["$ref"].Text(); ok {
			return NativeReference{name}, nil
		}
	}
	return NativeReference{}, fmt.Errorf("native aspect must be a named reference")
}
func nativeReferences(raw literal.Object, name string) ([]NativeReference, error) {
	value, present := raw[name]
	if !present {
		return nil, nil
	}
	array, ok := value.Array()
	if !ok {
		return nil, fmt.Errorf("native %s must be an array", name)
	}
	values := make([]NativeReference, len(array))
	for i, item := range array {
		value, err := decodeNativeReference(item)
		if err != nil {
			return nil, fmt.Errorf("native %s[%d]: %w", name, i, err)
		}
		values[i] = value
	}
	return values, nil
}
func decodeNativeAspect(raw literal.Object) (NativeAspectArgument, error) {
	if err := nativeFields(raw, "Argument Aspect", false); err != nil {
		return NativeAspectArgument{}, err
	}
	name, err := nativeText(raw, "Argument")
	if err != nil {
		return NativeAspectArgument{}, err
	}
	aspect, err := decodeNativeReference(raw["Aspect"])
	if err != nil {
		return NativeAspectArgument{}, err
	}
	return NativeAspectArgument{name, aspect}, nil
}
func decodeNativePolicy(raw literal.Object) (policy NativePolicy, err error) {
	if policy.CallbackEvent, err = nativeOptionalBool(raw, "CallbackEvent"); err != nil {
		return policy, err
	}
	if policy.CheckAllowChangeParent, err = nativeOptionalBool(raw, "CheckAllowChangeParent"); err != nil {
		return policy, err
	}
	if policy.CheckAllowInheritForbiddenLayoutAspects, err = nativeOptionalBool(raw, "CheckAllowInheritForbiddenLayoutAspects"); err != nil {
		return policy, err
	}
	if policy.ConstSecretAccessor, err = nativeOptionalBool(raw, "ConstSecretAccessor"); err != nil {
		return policy, err
	}
	if policy.HasRestrictions, err = nativeOptionalBool(raw, "HasRestrictions"); err != nil {
		return policy, err
	}
	if policy.IsProtectedFunction, err = nativeOptionalBool(raw, "IsProtectedFunction"); err != nil {
		return policy, err
	}
	if policy.MayReturnNothing, err = nativeOptionalBool(raw, "MayReturnNothing"); err != nil {
		return policy, err
	}
	if policy.MouseFocusValidForLimitedInput, err = nativeOptionalBool(raw, "MouseFocusValidForLimitedInput"); err != nil {
		return policy, err
	}
	if policy.RequireNPERestricted, err = nativeOptionalBool(raw, "RequireNPERestricted"); err != nil {
		return policy, err
	}
	if policy.RequiresActiveCommentator, err = nativeOptionalBool(raw, "RequiresActiveCommentator"); err != nil {
		return policy, err
	}
	if policy.RequiresAssignableScript, err = nativeOptionalBool(raw, "RequiresAssignableScript"); err != nil {
		return policy, err
	}
	if policy.RequiresCanChangeHitTestPoints, err = nativeOptionalBool(raw, "RequiresCanChangeHitTestPoints"); err != nil {
		return policy, err
	}
	if policy.RequiresClubsInitialized, err = nativeOptionalBool(raw, "RequiresClubsInitialized"); err != nil {
		return policy, err
	}
	if policy.RequiresCommentator, err = nativeOptionalBool(raw, "RequiresCommentator"); err != nil {
		return policy, err
	}
	if policy.RequiresComparableUnitTokens, err = nativeOptionalBool(raw, "RequiresComparableUnitTokens"); err != nil {
		return policy, err
	}
	if policy.RequiresDeclassifiedUnitIdentity, err = nativeOptionalBool(raw, "RequiresDeclassifiedUnitIdentity"); err != nil {
		return policy, err
	}
	if policy.RequiresFontStringTextAccess, err = nativeOptionalBool(raw, "RequiresFontStringTextAccess"); err != nil {
		return policy, err
	}
	if policy.RequiresFriendList, err = nativeOptionalBool(raw, "RequiresFriendList"); err != nil {
		return policy, err
	}
	if policy.RequiresIndexInRange, err = nativeOptionalBool(raw, "RequiresIndexInRange"); err != nil {
		return policy, err
	}
	if policy.RequiresLimitedInput, err = nativeOptionalBool(raw, "RequiresLimitedInput"); err != nil {
		return policy, err
	}
	if policy.RequiresNonReadOnlyCVar, err = nativeOptionalBool(raw, "RequiresNonReadOnlyCVar"); err != nil {
		return policy, err
	}
	if policy.RequiresNonSecretAura, err = nativeOptionalBool(raw, "RequiresNonSecretAura"); err != nil {
		return policy, err
	}
	if policy.RequiresNonSecureCVar, err = nativeOptionalBool(raw, "RequiresNonSecureCVar"); err != nil {
		return policy, err
	}
	if policy.RequiresRecentAllies, err = nativeOptionalBool(raw, "RequiresRecentAllies"); err != nil {
		return policy, err
	}
	if policy.RequiresRestrictedAbbreviationBreakpoints, err = nativeOptionalBool(raw, "RequiresRestrictedAbbreviationBreakpoints"); err != nil {
		return policy, err
	}
	if policy.RequiresScriptObjectAlphaAccess, err = nativeOptionalBool(raw, "RequiresScriptObjectAlphaAccess"); err != nil {
		return policy, err
	}
	if policy.RequiresScriptObjectDesaturationAccess, err = nativeOptionalBool(raw, "RequiresScriptObjectDesaturationAccess"); err != nil {
		return policy, err
	}
	if policy.RequiresSpellDiminishUI, err = nativeOptionalBool(raw, "RequiresSpellDiminishUI"); err != nil {
		return policy, err
	}
	if policy.RequiresStatusBarDesaturationAccess, err = nativeOptionalBool(raw, "RequiresStatusBarDesaturationAccess"); err != nil {
		return policy, err
	}
	if policy.RequiresSupportedScript, err = nativeOptionalBool(raw, "RequiresSupportedScript"); err != nil {
		return policy, err
	}
	if policy.RequiresTimedSignalMapAccess, err = nativeOptionalBool(raw, "RequiresTimedSignalMapAccess"); err != nil {
		return policy, err
	}
	if policy.RequiresUnitAuraAccess, err = nativeOptionalBool(raw, "RequiresUnitAuraAccess"); err != nil {
		return policy, err
	}
	if policy.RequiresUnitIdentityAccess, err = nativeOptionalBool(raw, "RequiresUnitIdentityAccess"); err != nil {
		return policy, err
	}
	if policy.RequiresValidAbbreviationBreakpoints, err = nativeOptionalBool(raw, "RequiresValidAbbreviationBreakpoints"); err != nil {
		return policy, err
	}
	if policy.RequiresValidActionSlot, err = nativeOptionalBool(raw, "RequiresValidActionSlot"); err != nil {
		return policy, err
	}
	if policy.RequiresValidAndPublicCVar, err = nativeOptionalBool(raw, "RequiresValidAndPublicCVar"); err != nil {
		return policy, err
	}
	if policy.RequiresValidFontAsset, err = nativeOptionalBool(raw, "RequiresValidFontAsset"); err != nil {
		return policy, err
	}
	if policy.RequiresValidFontHeight, err = nativeOptionalBool(raw, "RequiresValidFontHeight"); err != nil {
		return policy, err
	}
	if policy.RequiresValidInviteTarget, err = nativeOptionalBool(raw, "RequiresValidInviteTarget"); err != nil {
		return policy, err
	}
	if policy.RequiresValidTimelineEvent, err = nativeOptionalBool(raw, "RequiresValidTimelineEvent"); err != nil {
		return policy, err
	}
	if policy.RequiresValidUnitAuraInstance, err = nativeOptionalBool(raw, "RequiresValidUnitAuraInstance"); err != nil {
		return policy, err
	}
	if policy.RestrictedForMacroChatMessages, err = nativeOptionalBool(raw, "RestrictedForMacroChatMessages"); err != nil {
		return policy, err
	}
	if policy.ReturnsNeverSecret, err = nativeOptionalBool(raw, "ReturnsNeverSecret"); err != nil {
		return policy, err
	}
	if policy.SecretInActivePvPMatch, err = nativeOptionalBool(raw, "SecretInActivePvPMatch"); err != nil {
		return policy, err
	}
	if policy.SecretInChatMessagingLockdown, err = nativeOptionalBool(raw, "SecretInChatMessagingLockdown"); err != nil {
		return policy, err
	}
	if policy.SecretPayloads, err = nativeOptionalBool(raw, "SecretPayloads"); err != nil {
		return policy, err
	}
	if policy.SecretReturns, err = nativeOptionalBool(raw, "SecretReturns"); err != nil {
		return policy, err
	}
	if policy.SecretWhenAnchoringSecret, err = nativeOptionalBool(raw, "SecretWhenAnchoringSecret"); err != nil {
		return policy, err
	}
	if policy.SecretWhenAurasRestricted, err = nativeOptionalBool(raw, "SecretWhenAurasRestricted"); err != nil {
		return policy, err
	}
	if policy.SecretWhenCooldownsRestricted, err = nativeOptionalBool(raw, "SecretWhenCooldownsRestricted"); err != nil {
		return policy, err
	}
	if policy.SecretWhenCurveSecret, err = nativeOptionalBool(raw, "SecretWhenCurveSecret"); err != nil {
		return policy, err
	}
	if policy.SecretWhenEncounterEvent, err = nativeOptionalBool(raw, "SecretWhenEncounterEvent"); err != nil {
		return policy, err
	}
	if policy.SecretWhenInCombat, err = nativeOptionalBool(raw, "SecretWhenInCombat"); err != nil {
		return policy, err
	}
	if policy.SecretWhenLossOfControlInfoRestricted, err = nativeOptionalBool(raw, "SecretWhenLossOfControlInfoRestricted"); err != nil {
		return policy, err
	}
	if policy.SecretWhenLuaTableHasSecretKeys, err = nativeOptionalBool(raw, "SecretWhenLuaTableHasSecretKeys"); err != nil {
		return policy, err
	}
	if policy.SecretWhenNumericFormatterSecret, err = nativeOptionalBool(raw, "SecretWhenNumericFormatterSecret"); err != nil {
		return policy, err
	}
	if policy.SecretWhenTotemSlotSecret, err = nativeOptionalBool(raw, "SecretWhenTotemSlotSecret"); err != nil {
		return policy, err
	}
	if policy.SecretWhenUnitAuraRestricted, err = nativeOptionalBool(raw, "SecretWhenUnitAuraRestricted"); err != nil {
		return policy, err
	}
	if policy.SecretWhenUnitComparisonRestricted, err = nativeOptionalBool(raw, "SecretWhenUnitComparisonRestricted"); err != nil {
		return policy, err
	}
	if policy.SecretWhenUnitHealthMaxRestricted, err = nativeOptionalBool(raw, "SecretWhenUnitHealthMaxRestricted"); err != nil {
		return policy, err
	}
	if policy.SecretWhenUnitIdentityRestricted, err = nativeOptionalBool(raw, "SecretWhenUnitIdentityRestricted"); err != nil {
		return policy, err
	}
	if policy.SecretWhenUnitNameIdentityRestricted, err = nativeOptionalBool(raw, "SecretWhenUnitNameIdentityRestricted"); err != nil {
		return policy, err
	}
	if policy.SecretWhenUnitPossessionRestricted, err = nativeOptionalBool(raw, "SecretWhenUnitPossessionRestricted"); err != nil {
		return policy, err
	}
	if policy.SecretWhenUnitPowerMaxRestricted, err = nativeOptionalBool(raw, "SecretWhenUnitPowerMaxRestricted"); err != nil {
		return policy, err
	}
	if policy.SecretWhenUnitPowerRestricted, err = nativeOptionalBool(raw, "SecretWhenUnitPowerRestricted"); err != nil {
		return policy, err
	}
	if policy.SecretWhenUnitSpellCastRestricted, err = nativeOptionalBool(raw, "SecretWhenUnitSpellCastRestricted"); err != nil {
		return policy, err
	}
	if policy.SecretWhenUnitStatsRestricted, err = nativeOptionalBool(raw, "SecretWhenUnitStatsRestricted"); err != nil {
		return policy, err
	}
	if policy.SecretWhenUnitThreatStateRestricted, err = nativeOptionalBool(raw, "SecretWhenUnitThreatStateRestricted"); err != nil {
		return policy, err
	}
	if policy.SecretWhenUnitThreatValuesRestricted, err = nativeOptionalBool(raw, "SecretWhenUnitThreatValuesRestricted"); err != nil {
		return policy, err
	}
	if policy.SecureHooksAllowed, err = nativeOptionalBool(raw, "SecureHooksAllowed"); err != nil {
		return policy, err
	}
	if policy.SynchronousEvent, err = nativeOptionalBool(raw, "SynchronousEvent"); err != nil {
		return policy, err
	}
	if policy.TransliteratorAllowed, err = nativeOptionalBool(raw, "TransliteratorAllowed"); err != nil {
		return policy, err
	}
	if policy.UniqueEvent, err = nativeOptionalBool(raw, "UniqueEvent"); err != nil {
		return policy, err
	}
	if policy.Environment, err = nativeEnvironment(raw); err != nil {
		return policy, err
	}
	if policy.SecretArguments, err = nativeEnum(raw, "SecretArguments", "AllowedWhenTainted", "AllowedWhenUntainted", "NotAllowed"); err != nil {
		return policy, err
	}
	if policy.SecretArgumentsAddAspect, err = nativeReferences(raw, "SecretArgumentsAddAspect"); err != nil {
		return policy, err
	}
	if policy.SecretReturnsForAspect, err = nativeReferences(raw, "SecretReturnsForAspect"); err != nil {
		return policy, err
	}
	if policy.AddsForbiddenAspects, err = nativeItems(raw, "AddsForbiddenAspects", decodeNativeAspect); err != nil {
		return policy, err
	}
	if policy.ChecksForbiddenAspects, err = nativeItems(raw, "ChecksForbiddenAspects", decodeNativeAspect); err != nil {
		return policy, err
	}
	return policy, nil
}
func decodeNativeMetadata(raw literal.Object) (value NativeMetadata, err error) {
	if err = nativeFields(raw, "Name Type Namespace ObjectType Environment Documentation Functions Events Tables Predicates", false); err != nil {
		return value, err
	}
	if value.Name, err = nativeText(raw, "Name"); err != nil {
		return value, err
	}
	kind, err := nativeEnum(raw, "Type", "System", "ScriptObject")
	if err != nil {
		return value, err
	}
	if kind != nil {
		value.Type = *kind
	}
	if value.Namespace, err = nativeText(raw, "Namespace"); err != nil {
		return value, err
	}
	objectType, err := nativeEnum(raw, "ObjectType", "Userdata")
	if err != nil {
		return value, err
	}
	if objectType != nil {
		value.ObjectType = *objectType
	}
	environment, err := nativeEnvironment(raw)
	if err != nil {
		return value, err
	}
	if environment != nil {
		value.Environment = *environment
	}
	if err = validateNativeDocumentation(raw); err != nil {
		return value, err
	}
	if value.Functions, err = nativeItems(raw, "Functions", decodeNativeFunction); err != nil {
		return value, err
	}
	if value.Events, err = nativeItems(raw, "Events", decodeNativeEvent); err != nil {
		return value, err
	}
	if value.Tables, err = nativeItems(raw, "Tables", decodeNativeTable); err != nil {
		return value, err
	}
	if value.Predicates, err = nativeItems(raw, "Predicates", decodeNativePredicate); err != nil {
		return value, err
	}
	functions := map[string]bool{}
	for _, function := range value.Functions {
		if functions[function.Name] {
			return value, fmt.Errorf("duplicate native function %s", function.Name)
		}
		functions[function.Name] = true
	}
	events := map[string]bool{}
	for _, event := range value.Events {
		if events[event.Name] {
			return value, fmt.Errorf("duplicate native event %s", event.Name)
		}
		events[event.Name] = true
	}

	return value, nil
}
func decodeNativeFunction(raw literal.Object) (value NativeFunction, err error) {
	if err = nativeIdentity(raw); err != nil {
		return value, err
	}
	if err = nativeFields(raw, "Name Type Namespace Documentation Arguments Returns", true); err != nil {
		return value, err
	}
	if value.Name, err = nativeText(raw, "Name"); err != nil {
		return value, err
	}
	if value.Type, err = nativeKind(raw, "Function"); err != nil {
		return value, err
	}
	if value.Namespace, err = nativeText(raw, "Namespace"); err != nil {
		return value, err
	}
	if err = validateNativeDocumentation(raw); err != nil {
		return value, err
	}
	if value.Arguments, err = nativeItems(raw, "Arguments", decodeNativeParameter); err != nil {
		return value, err
	}
	if value.Returns, err = nativeItems(raw, "Returns", decodeNativeParameter); err != nil {
		return value, err
	}
	if value.Policy, err = decodeNativePolicy(raw); err != nil {
		return value, err
	}
	return value, nil
}
func decodeNativeParameter(raw literal.Object) (value NativeParameter, err error) {
	if _, err = nativeOptionalInt(raw, "EnumValue"); err != nil {
		return value, err
	}
	for _, name := range []string{"Default", "Value"} {
		if err = nativeScalar(raw, name); err != nil {
			return value, err
		}
	}
	if err = nativeFields(raw, "Name Type InnerType KeyType Mixin Documentation Nilable NilableContents NeverSecret NeverSecretContents ConditionalSecret ConditionalSecretContents SecretValue StrideIndex Default Value EnumValue", false); err != nil {
		return value, err
	}
	_, value.HasDefault = raw["Default"]
	if value.Name, err = nativeText(raw, "Name"); err != nil {
		return value, err
	}
	if value.Type, err = nativeText(raw, "Type"); err != nil {
		return value, err
	}
	if value.InnerType, err = nativeText(raw, "InnerType"); err != nil {
		return value, err
	}
	if value.KeyType, err = nativeText(raw, "KeyType"); err != nil {
		return value, err
	}
	if value.Mixin, err = nativeText(raw, "Mixin"); err != nil {
		return value, err
	}
	if err = validateNativeDocumentation(raw); err != nil {
		return value, err
	}
	if value.Nilable, err = nativeOptionalBool(raw, "Nilable"); err != nil {
		return value, err
	}
	if value.NilableContents, err = nativeOptionalBool(raw, "NilableContents"); err != nil {
		return value, err
	}
	if value.NeverSecret, err = nativeOptionalBool(raw, "NeverSecret"); err != nil {
		return value, err
	}
	if value.NeverSecretContents, err = nativeOptionalBool(raw, "NeverSecretContents"); err != nil {
		return value, err
	}
	if value.ConditionalSecret, err = nativeOptionalBool(raw, "ConditionalSecret"); err != nil {
		return value, err
	}
	if value.ConditionalSecretContents, err = nativeOptionalBool(raw, "ConditionalSecretContents"); err != nil {
		return value, err
	}
	if value.SecretValue, err = nativeOptionalBool(raw, "SecretValue"); err != nil {
		return value, err
	}
	if value.StrideIndex, err = nativeOptionalInt(raw, "StrideIndex"); err != nil {
		return value, err
	}
	return value, nil
}
func decodeNativeEvent(raw literal.Object) (value NativeEvent, err error) {
	if err = nativeIdentity(raw); err != nil {
		return value, err
	}
	if err = nativeFields(raw, "Name Type LiteralName Documentation Payload", true); err != nil {
		return value, err
	}
	if value.Name, err = nativeText(raw, "Name"); err != nil {
		return value, err
	}
	if value.Type, err = nativeKind(raw, "Event"); err != nil {
		return value, err
	}
	if value.LiteralName, err = nativeText(raw, "LiteralName"); err != nil {
		return value, err
	}
	if err = validateNativeDocumentation(raw); err != nil {
		return value, err
	}
	if value.Payload, err = nativeItems(raw, "Payload", decodeNativeParameter); err != nil {
		return value, err
	}
	if value.Policy, err = decodeNativePolicy(raw); err != nil {
		return value, err
	}
	return value, nil
}
func decodeNativeTable(raw literal.Object) (value NativeTable, err error) {
	for _, name := range []string{"NumValues", "MinValue", "MaxValue"} {
		if _, err = nativeOptionalInt(raw, name); err != nil {
			return value, err
		}
	}
	if err = nativeFields(raw, "Name Type Documentation Arguments Fields Values Returns NumValues MinValue MaxValue", true); err != nil {
		return value, err
	}
	if value.Name, err = nativeText(raw, "Name"); err != nil {
		return value, err
	}
	if value.Type, err = nativeKind(raw, "Enumeration", "Structure", "CallbackType", "Constants"); err != nil {
		return value, err
	}
	if err = validateNativeDocumentation(raw); err != nil {
		return value, err
	}
	if value.Arguments, err = nativeItems(raw, "Arguments", decodeNativeParameter); err != nil {
		return value, err
	}
	if value.Fields, err = nativeItems(raw, "Fields", decodeNativeParameter); err != nil {
		return value, err
	}
	if value.Values, err = nativeItems(raw, "Values", decodeNativeParameter); err != nil {
		return value, err
	}
	if value.Returns, err = nativeItems(raw, "Returns", decodeNativeParameter); err != nil {
		return value, err
	}
	if value.Policy, err = decodeNativePolicy(raw); err != nil {
		return value, err
	}
	for _, key := range []string{"NumValues", "MinValue", "MaxValue"} {
		if _, present := raw[key]; present {
			value.EnumMetadata = append(value.EnumMetadata, key)
		}
	}
	return value, nil
}
func decodeNativePredicate(raw literal.Object) (value NativePredicate, err error) {
	if err = nativeFields(raw, "Name Type FailureMode Documentation", false); err != nil {
		return value, err
	}
	if value.Name, err = nativeText(raw, "Name"); err != nil {
		return value, err
	}
	if value.Type, err = nativeKind(raw, "Precondition", "Secret"); err != nil {
		return value, err
	}
	failureMode, err := nativeEnum(raw, "FailureMode", "Error", "ReturnNothing", "ReturnWithError")
	if err != nil {
		return value, err
	}
	if failureMode != nil {
		value.FailureMode = *failureMode
	}
	if err = validateNativeDocumentation(raw); err != nil {
		return value, err
	}
	return value, nil
}

func (table NativeTable) Callback() NativeFunction {
	return NativeFunction{Name: table.Name, Type: table.Type, Arguments: table.Arguments, Returns: table.Returns, Policy: table.Policy}
}

func (p NativePolicy) restrictions() NativePolicy {
	return NativePolicy{
		CheckAllowChangeParent:                    p.CheckAllowChangeParent,
		CheckAllowInheritForbiddenLayoutAspects:   p.CheckAllowInheritForbiddenLayoutAspects,
		ConstSecretAccessor:                       p.ConstSecretAccessor,
		HasRestrictions:                           p.HasRestrictions,
		IsProtectedFunction:                       p.IsProtectedFunction,
		RequireNPERestricted:                      p.RequireNPERestricted,
		RequiresActiveCommentator:                 p.RequiresActiveCommentator,
		RequiresAssignableScript:                  p.RequiresAssignableScript,
		RequiresCanChangeHitTestPoints:            p.RequiresCanChangeHitTestPoints,
		RequiresClubsInitialized:                  p.RequiresClubsInitialized,
		RequiresCommentator:                       p.RequiresCommentator,
		RequiresComparableUnitTokens:              p.RequiresComparableUnitTokens,
		RequiresDeclassifiedUnitIdentity:          p.RequiresDeclassifiedUnitIdentity,
		RequiresFontStringTextAccess:              p.RequiresFontStringTextAccess,
		RequiresFriendList:                        p.RequiresFriendList,
		RequiresIndexInRange:                      p.RequiresIndexInRange,
		RequiresLimitedInput:                      p.RequiresLimitedInput,
		RequiresNonReadOnlyCVar:                   p.RequiresNonReadOnlyCVar,
		RequiresNonSecretAura:                     p.RequiresNonSecretAura,
		RequiresNonSecureCVar:                     p.RequiresNonSecureCVar,
		RequiresRecentAllies:                      p.RequiresRecentAllies,
		RequiresRestrictedAbbreviationBreakpoints: p.RequiresRestrictedAbbreviationBreakpoints,
		RequiresScriptObjectAlphaAccess:           p.RequiresScriptObjectAlphaAccess,
		RequiresScriptObjectDesaturationAccess:    p.RequiresScriptObjectDesaturationAccess,
		RequiresSpellDiminishUI:                   p.RequiresSpellDiminishUI,
		RequiresStatusBarDesaturationAccess:       p.RequiresStatusBarDesaturationAccess,
		RequiresSupportedScript:                   p.RequiresSupportedScript,
		RequiresTimedSignalMapAccess:              p.RequiresTimedSignalMapAccess,
		RequiresUnitAuraAccess:                    p.RequiresUnitAuraAccess,
		RequiresUnitIdentityAccess:                p.RequiresUnitIdentityAccess,
		RequiresValidAbbreviationBreakpoints:      p.RequiresValidAbbreviationBreakpoints,
		RequiresValidActionSlot:                   p.RequiresValidActionSlot,
		RequiresValidAndPublicCVar:                p.RequiresValidAndPublicCVar,
		RequiresValidFontAsset:                    p.RequiresValidFontAsset,
		RequiresValidFontHeight:                   p.RequiresValidFontHeight,
		RequiresValidInviteTarget:                 p.RequiresValidInviteTarget,
		RequiresValidTimelineEvent:                p.RequiresValidTimelineEvent,
		RequiresValidUnitAuraInstance:             p.RequiresValidUnitAuraInstance,
		RestrictedForMacroChatMessages:            p.RestrictedForMacroChatMessages,
		ReturnsNeverSecret:                        p.ReturnsNeverSecret,
		SecretInActivePvPMatch:                    p.SecretInActivePvPMatch,
		SecretInChatMessagingLockdown:             p.SecretInChatMessagingLockdown,
		SecretPayloads:                            p.SecretPayloads,
		SecretReturns:                             p.SecretReturns,
		SecretWhenAnchoringSecret:                 p.SecretWhenAnchoringSecret,
		SecretWhenAurasRestricted:                 p.SecretWhenAurasRestricted,
		SecretWhenCooldownsRestricted:             p.SecretWhenCooldownsRestricted,
		SecretWhenCurveSecret:                     p.SecretWhenCurveSecret,
		SecretWhenEncounterEvent:                  p.SecretWhenEncounterEvent,
		SecretWhenInCombat:                        p.SecretWhenInCombat,
		SecretWhenLossOfControlInfoRestricted:     p.SecretWhenLossOfControlInfoRestricted,
		SecretWhenLuaTableHasSecretKeys:           p.SecretWhenLuaTableHasSecretKeys,
		SecretWhenNumericFormatterSecret:          p.SecretWhenNumericFormatterSecret,
		SecretWhenTotemSlotSecret:                 p.SecretWhenTotemSlotSecret,
		SecretWhenUnitAuraRestricted:              p.SecretWhenUnitAuraRestricted,
		SecretWhenUnitComparisonRestricted:        p.SecretWhenUnitComparisonRestricted,
		SecretWhenUnitHealthMaxRestricted:         p.SecretWhenUnitHealthMaxRestricted,
		SecretWhenUnitIdentityRestricted:          p.SecretWhenUnitIdentityRestricted,
		SecretWhenUnitNameIdentityRestricted:      p.SecretWhenUnitNameIdentityRestricted,
		SecretWhenUnitPossessionRestricted:        p.SecretWhenUnitPossessionRestricted,
		SecretWhenUnitPowerMaxRestricted:          p.SecretWhenUnitPowerMaxRestricted,
		SecretWhenUnitPowerRestricted:             p.SecretWhenUnitPowerRestricted,
		SecretWhenUnitSpellCastRestricted:         p.SecretWhenUnitSpellCastRestricted,
		SecretWhenUnitStatsRestricted:             p.SecretWhenUnitStatsRestricted,
		SecretWhenUnitThreatStateRestricted:       p.SecretWhenUnitThreatStateRestricted,
		SecretWhenUnitThreatValuesRestricted:      p.SecretWhenUnitThreatValuesRestricted,
		SecureHooksAllowed:                        p.SecureHooksAllowed,
		Environment:                               p.Environment,
		SecretArguments:                           p.SecretArguments,
		SecretArgumentsAddAspect:                  p.SecretArgumentsAddAspect,
		SecretReturnsForAspect:                    p.SecretReturnsForAspect,
		AddsForbiddenAspects:                      p.AddsForbiddenAspects,
		ChecksForbiddenAspects:                    p.ChecksForbiddenAspects,
	}
}
func (p NativePolicy) secretResult() bool {
	if enabled(p.ReturnsNeverSecret) {
		return false
	}
	return enabled(p.SecretInActivePvPMatch) || enabled(p.SecretInChatMessagingLockdown) || enabled(p.SecretReturns) || enabled(p.SecretWhenAnchoringSecret) || enabled(p.SecretWhenAurasRestricted) || enabled(p.SecretWhenCooldownsRestricted) || enabled(p.SecretWhenCurveSecret) || enabled(p.SecretWhenEncounterEvent) || enabled(p.SecretWhenInCombat) || enabled(p.SecretWhenLossOfControlInfoRestricted) || enabled(p.SecretWhenLuaTableHasSecretKeys) || enabled(p.SecretWhenNumericFormatterSecret) || enabled(p.SecretWhenTotemSlotSecret) || enabled(p.SecretWhenUnitAuraRestricted) || enabled(p.SecretWhenUnitComparisonRestricted) || enabled(p.SecretWhenUnitHealthMaxRestricted) || enabled(p.SecretWhenUnitIdentityRestricted) || enabled(p.SecretWhenUnitNameIdentityRestricted) || enabled(p.SecretWhenUnitPossessionRestricted) || enabled(p.SecretWhenUnitPowerMaxRestricted) || enabled(p.SecretWhenUnitPowerRestricted) || enabled(p.SecretWhenUnitSpellCastRestricted) || enabled(p.SecretWhenUnitStatsRestricted) || enabled(p.SecretWhenUnitThreatStateRestricted) || enabled(p.SecretWhenUnitThreatValuesRestricted) || len(p.SecretReturnsForAspect) > 0
}
func (p NativePolicy) precondition(name string) bool {
	switch name {
	case "CallbackEvent":
		return enabled(p.CallbackEvent)
	case "CheckAllowChangeParent":
		return enabled(p.CheckAllowChangeParent)
	case "CheckAllowInheritForbiddenLayoutAspects":
		return enabled(p.CheckAllowInheritForbiddenLayoutAspects)
	case "ConstSecretAccessor":
		return enabled(p.ConstSecretAccessor)
	case "HasRestrictions":
		return enabled(p.HasRestrictions)
	case "IsProtectedFunction":
		return enabled(p.IsProtectedFunction)
	case "MayReturnNothing":
		return enabled(p.MayReturnNothing)
	case "MouseFocusValidForLimitedInput":
		return enabled(p.MouseFocusValidForLimitedInput)
	case "RequireNPERestricted":
		return enabled(p.RequireNPERestricted)
	case "RequiresActiveCommentator":
		return enabled(p.RequiresActiveCommentator)
	case "RequiresAssignableScript":
		return enabled(p.RequiresAssignableScript)
	case "RequiresCanChangeHitTestPoints":
		return enabled(p.RequiresCanChangeHitTestPoints)
	case "RequiresClubsInitialized":
		return enabled(p.RequiresClubsInitialized)
	case "RequiresCommentator":
		return enabled(p.RequiresCommentator)
	case "RequiresComparableUnitTokens":
		return enabled(p.RequiresComparableUnitTokens)
	case "RequiresDeclassifiedUnitIdentity":
		return enabled(p.RequiresDeclassifiedUnitIdentity)
	case "RequiresFontStringTextAccess":
		return enabled(p.RequiresFontStringTextAccess)
	case "RequiresFriendList":
		return enabled(p.RequiresFriendList)
	case "RequiresIndexInRange":
		return enabled(p.RequiresIndexInRange)
	case "RequiresLimitedInput":
		return enabled(p.RequiresLimitedInput)
	case "RequiresNonReadOnlyCVar":
		return enabled(p.RequiresNonReadOnlyCVar)
	case "RequiresNonSecretAura":
		return enabled(p.RequiresNonSecretAura)
	case "RequiresNonSecureCVar":
		return enabled(p.RequiresNonSecureCVar)
	case "RequiresRecentAllies":
		return enabled(p.RequiresRecentAllies)
	case "RequiresRestrictedAbbreviationBreakpoints":
		return enabled(p.RequiresRestrictedAbbreviationBreakpoints)
	case "RequiresScriptObjectAlphaAccess":
		return enabled(p.RequiresScriptObjectAlphaAccess)
	case "RequiresScriptObjectDesaturationAccess":
		return enabled(p.RequiresScriptObjectDesaturationAccess)
	case "RequiresSpellDiminishUI":
		return enabled(p.RequiresSpellDiminishUI)
	case "RequiresStatusBarDesaturationAccess":
		return enabled(p.RequiresStatusBarDesaturationAccess)
	case "RequiresSupportedScript":
		return enabled(p.RequiresSupportedScript)
	case "RequiresTimedSignalMapAccess":
		return enabled(p.RequiresTimedSignalMapAccess)
	case "RequiresUnitAuraAccess":
		return enabled(p.RequiresUnitAuraAccess)
	case "RequiresUnitIdentityAccess":
		return enabled(p.RequiresUnitIdentityAccess)
	case "RequiresValidAbbreviationBreakpoints":
		return enabled(p.RequiresValidAbbreviationBreakpoints)
	case "RequiresValidActionSlot":
		return enabled(p.RequiresValidActionSlot)
	case "RequiresValidAndPublicCVar":
		return enabled(p.RequiresValidAndPublicCVar)
	case "RequiresValidFontAsset":
		return enabled(p.RequiresValidFontAsset)
	case "RequiresValidFontHeight":
		return enabled(p.RequiresValidFontHeight)
	case "RequiresValidInviteTarget":
		return enabled(p.RequiresValidInviteTarget)
	case "RequiresValidTimelineEvent":
		return enabled(p.RequiresValidTimelineEvent)
	case "RequiresValidUnitAuraInstance":
		return enabled(p.RequiresValidUnitAuraInstance)
	case "RestrictedForMacroChatMessages":
		return enabled(p.RestrictedForMacroChatMessages)
	case "ReturnsNeverSecret":
		return enabled(p.ReturnsNeverSecret)
	case "SecretInActivePvPMatch":
		return enabled(p.SecretInActivePvPMatch)
	case "SecretInChatMessagingLockdown":
		return enabled(p.SecretInChatMessagingLockdown)
	case "SecretPayloads":
		return enabled(p.SecretPayloads)
	case "SecretReturns":
		return enabled(p.SecretReturns)
	case "SecretWhenAnchoringSecret":
		return enabled(p.SecretWhenAnchoringSecret)
	case "SecretWhenAurasRestricted":
		return enabled(p.SecretWhenAurasRestricted)
	case "SecretWhenCooldownsRestricted":
		return enabled(p.SecretWhenCooldownsRestricted)
	case "SecretWhenCurveSecret":
		return enabled(p.SecretWhenCurveSecret)
	case "SecretWhenEncounterEvent":
		return enabled(p.SecretWhenEncounterEvent)
	case "SecretWhenInCombat":
		return enabled(p.SecretWhenInCombat)
	case "SecretWhenLossOfControlInfoRestricted":
		return enabled(p.SecretWhenLossOfControlInfoRestricted)
	case "SecretWhenLuaTableHasSecretKeys":
		return enabled(p.SecretWhenLuaTableHasSecretKeys)
	case "SecretWhenNumericFormatterSecret":
		return enabled(p.SecretWhenNumericFormatterSecret)
	case "SecretWhenTotemSlotSecret":
		return enabled(p.SecretWhenTotemSlotSecret)
	case "SecretWhenUnitAuraRestricted":
		return enabled(p.SecretWhenUnitAuraRestricted)
	case "SecretWhenUnitComparisonRestricted":
		return enabled(p.SecretWhenUnitComparisonRestricted)
	case "SecretWhenUnitHealthMaxRestricted":
		return enabled(p.SecretWhenUnitHealthMaxRestricted)
	case "SecretWhenUnitIdentityRestricted":
		return enabled(p.SecretWhenUnitIdentityRestricted)
	case "SecretWhenUnitNameIdentityRestricted":
		return enabled(p.SecretWhenUnitNameIdentityRestricted)
	case "SecretWhenUnitPossessionRestricted":
		return enabled(p.SecretWhenUnitPossessionRestricted)
	case "SecretWhenUnitPowerMaxRestricted":
		return enabled(p.SecretWhenUnitPowerMaxRestricted)
	case "SecretWhenUnitPowerRestricted":
		return enabled(p.SecretWhenUnitPowerRestricted)
	case "SecretWhenUnitSpellCastRestricted":
		return enabled(p.SecretWhenUnitSpellCastRestricted)
	case "SecretWhenUnitStatsRestricted":
		return enabled(p.SecretWhenUnitStatsRestricted)
	case "SecretWhenUnitThreatStateRestricted":
		return enabled(p.SecretWhenUnitThreatStateRestricted)
	case "SecretWhenUnitThreatValuesRestricted":
		return enabled(p.SecretWhenUnitThreatValuesRestricted)
	case "SecureHooksAllowed":
		return enabled(p.SecureHooksAllowed)
	case "SynchronousEvent":
		return enabled(p.SynchronousEvent)
	case "TransliteratorAllowed":
		return enabled(p.TransliteratorAllowed)
	case "UniqueEvent":
		return enabled(p.UniqueEvent)
	}
	return false
}
func (p NativeParameter) secretResult() bool {
	return !enabled(p.NeverSecret) && (enabled(p.SecretValue) || enabled(p.ConditionalSecret) || enabled(p.ConditionalSecretContents))
}
func (p NativeReference) MarshalJSON() ([]byte, error) {
	return jsonCompact(struct {
		Ref string `json:"$ref"`
	}{p.Name})
}
func (p *NativeReference) UnmarshalJSON(data []byte) error {
	var ref struct {
		Ref string `json:"$ref"`
	}
	if err := decodeJSON(data, &ref); err != nil {
		return err
	}
	p.Name = ref.Ref
	return nil
}

func (p *NativeParameter) UnmarshalJSON(data []byte) error {
	raw, e := literal.ParseJSON(data)
	if e != nil {
		return e
	}
	object, ok := raw.Object()
	if !ok {
		return fmt.Errorf("native parameter must be object")
	}
	value, e := decodeNativeParameter(object)
	if e != nil {
		return e
	}
	*p = value
	return nil
}

// Unknown metadata can change call permissions. Reject it before interpreting
// contracts so a source update cannot silently weaken the generated policy.
func nativeFields(raw literal.Object, allowed string, policy bool) error {
	fields := strings.Fields(allowed)
	for _, name := range sortedKeys(raw) {
		if !slices.Contains(fields, name) && !(policy && nativePolicyField(name)) {
			return fmt.Errorf("unsupported native field %s", name)
		}
	}
	return nil
}
func nativePolicyField(name string) bool {
	switch name {
	case "AddsForbiddenAspects",
		"ChecksForbiddenAspects",
		"Environment",
		"SecretArguments",
		"SecretArgumentsAddAspect",
		"SecretReturnsForAspect",
		"CallbackEvent",
		"CheckAllowChangeParent",
		"CheckAllowInheritForbiddenLayoutAspects",
		"ConstSecretAccessor",
		"HasRestrictions",
		"IsProtectedFunction",
		"MayReturnNothing",
		"MouseFocusValidForLimitedInput",
		"RequireNPERestricted",
		"RequiresActiveCommentator",
		"RequiresAssignableScript",
		"RequiresCanChangeHitTestPoints",
		"RequiresClubsInitialized",
		"RequiresCommentator",
		"RequiresComparableUnitTokens",
		"RequiresDeclassifiedUnitIdentity",
		"RequiresFontStringTextAccess",
		"RequiresFriendList",
		"RequiresIndexInRange",
		"RequiresLimitedInput",
		"RequiresNonReadOnlyCVar",
		"RequiresNonSecretAura",
		"RequiresNonSecureCVar",
		"RequiresRecentAllies",
		"RequiresRestrictedAbbreviationBreakpoints",
		"RequiresScriptObjectAlphaAccess",
		"RequiresScriptObjectDesaturationAccess",
		"RequiresSpellDiminishUI",
		"RequiresStatusBarDesaturationAccess",
		"RequiresSupportedScript",
		"RequiresTimedSignalMapAccess",
		"RequiresUnitAuraAccess",
		"RequiresUnitIdentityAccess",
		"RequiresValidAbbreviationBreakpoints",
		"RequiresValidActionSlot",
		"RequiresValidAndPublicCVar",
		"RequiresValidFontAsset",
		"RequiresValidFontHeight",
		"RequiresValidInviteTarget",
		"RequiresValidTimelineEvent",
		"RequiresValidUnitAuraInstance",
		"RestrictedForMacroChatMessages",
		"ReturnsNeverSecret",
		"SecretInActivePvPMatch",
		"SecretInChatMessagingLockdown",
		"SecretPayloads",
		"SecretReturns",
		"SecretWhenAnchoringSecret",
		"SecretWhenAurasRestricted",
		"SecretWhenCooldownsRestricted",
		"SecretWhenCurveSecret",
		"SecretWhenEncounterEvent",
		"SecretWhenInCombat",
		"SecretWhenLossOfControlInfoRestricted",
		"SecretWhenLuaTableHasSecretKeys",
		"SecretWhenNumericFormatterSecret",
		"SecretWhenTotemSlotSecret",
		"SecretWhenUnitAuraRestricted",
		"SecretWhenUnitComparisonRestricted",
		"SecretWhenUnitHealthMaxRestricted",
		"SecretWhenUnitIdentityRestricted",
		"SecretWhenUnitNameIdentityRestricted",
		"SecretWhenUnitPossessionRestricted",
		"SecretWhenUnitPowerMaxRestricted",
		"SecretWhenUnitPowerRestricted",
		"SecretWhenUnitSpellCastRestricted",
		"SecretWhenUnitStatsRestricted",
		"SecretWhenUnitThreatStateRestricted",
		"SecretWhenUnitThreatValuesRestricted",
		"SecureHooksAllowed",
		"SynchronousEvent",
		"TransliteratorAllowed",
		"UniqueEvent":
		return true
	}
	return false
}
func nativeIdentity(raw literal.Object) error {
	for _, name := range []string{"Name", "Type"} {
		value, err := nativeText(raw, name)
		if err != nil {
			return err
		}
		if value == "" {
			return fmt.Errorf("native %s must be a nonempty string", name)
		}
	}
	return nil
}
func nativeScalar(raw literal.Object, name string) error {
	value, present := raw[name]
	if !present {
		return nil
	}
	switch value.Kind() {
	case literal.Boolean, literal.String, literal.Integer, literal.FloatKind, literal.Reference, literal.Expression:
		return nil
	default:
		return fmt.Errorf("native %s must be a scalar or named constant", name)
	}
}
