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
	EventNames  []string
	Functions   []NativeFunction
	Name        string
	Namespace   string
	ObjectType  string
	Predicates  []NativePredicate
	Tables      []NativeTable
	Type        string
}
type NativeFunction struct {
	Name               string
	Arguments, Returns []NativeParameter
	Policy             NativePolicy
}
type NativeParameter struct {
	HasDefault                                                 bool
	Name, Type, InnerType, KeyType                             string
	Nilable, NilableContents, NeverSecret, NeverSecretContents *bool
	ConditionalSecret, ConditionalSecretContents, SecretValue  *bool
	StrideIndex                                                *int
}
type NativeTable struct {
	EnumMetadata               []string
	Policy                     NativePolicy
	Name, Type                 string
	Arguments, Fields, Returns []NativeParameter
}
type NativePredicate struct {
	Name, Type, FailureMode string
}
type NativeReference struct{ Name string }
type NativeAspectArgument struct {
	Argument string
	Aspect   NativeReference
}

type nativeFlag string

const (
	nativeCallbackEvent                             nativeFlag = "CallbackEvent"
	nativeCheckAllowChangeParent                    nativeFlag = "CheckAllowChangeParent"
	nativeCheckAllowInheritForbiddenLayoutAspects   nativeFlag = "CheckAllowInheritForbiddenLayoutAspects"
	nativeConstSecretAccessor                       nativeFlag = "ConstSecretAccessor"
	nativeHasRestrictions                           nativeFlag = "HasRestrictions"
	nativeIsProtectedFunction                       nativeFlag = "IsProtectedFunction"
	nativeMayReturnNothing                          nativeFlag = "MayReturnNothing"
	nativeMouseFocusValidForLimitedInput            nativeFlag = "MouseFocusValidForLimitedInput"
	nativeRequireNPERestricted                      nativeFlag = "RequireNPERestricted"
	nativeRequiresActiveCommentator                 nativeFlag = "RequiresActiveCommentator"
	nativeRequiresAssignableScript                  nativeFlag = "RequiresAssignableScript"
	nativeRequiresCanChangeHitTestPoints            nativeFlag = "RequiresCanChangeHitTestPoints"
	nativeRequiresClubsInitialized                  nativeFlag = "RequiresClubsInitialized"
	nativeRequiresCommentator                       nativeFlag = "RequiresCommentator"
	nativeRequiresComparableUnitTokens              nativeFlag = "RequiresComparableUnitTokens"
	nativeRequiresDeclassifiedUnitIdentity          nativeFlag = "RequiresDeclassifiedUnitIdentity"
	nativeRequiresFontStringTextAccess              nativeFlag = "RequiresFontStringTextAccess"
	nativeRequiresFriendList                        nativeFlag = "RequiresFriendList"
	nativeRequiresIndexInRange                      nativeFlag = "RequiresIndexInRange"
	nativeRequiresLimitedInput                      nativeFlag = "RequiresLimitedInput"
	nativeRequiresNonReadOnlyCVar                   nativeFlag = "RequiresNonReadOnlyCVar"
	nativeRequiresNonSecretAura                     nativeFlag = "RequiresNonSecretAura"
	nativeRequiresNonSecureCVar                     nativeFlag = "RequiresNonSecureCVar"
	nativeRequiresRecentAllies                      nativeFlag = "RequiresRecentAllies"
	nativeRequiresRestrictedAbbreviationBreakpoints nativeFlag = "RequiresRestrictedAbbreviationBreakpoints"
	nativeRequiresScriptObjectAlphaAccess           nativeFlag = "RequiresScriptObjectAlphaAccess"
	nativeRequiresScriptObjectDesaturationAccess    nativeFlag = "RequiresScriptObjectDesaturationAccess"
	nativeRequiresSpellDiminishUI                   nativeFlag = "RequiresSpellDiminishUI"
	nativeRequiresStatusBarDesaturationAccess       nativeFlag = "RequiresStatusBarDesaturationAccess"
	nativeRequiresSupportedScript                   nativeFlag = "RequiresSupportedScript"
	nativeRequiresTimedSignalMapAccess              nativeFlag = "RequiresTimedSignalMapAccess"
	nativeRequiresUnitAuraAccess                    nativeFlag = "RequiresUnitAuraAccess"
	nativeRequiresUnitIdentityAccess                nativeFlag = "RequiresUnitIdentityAccess"
	nativeRequiresValidAbbreviationBreakpoints      nativeFlag = "RequiresValidAbbreviationBreakpoints"
	nativeRequiresValidActionSlot                   nativeFlag = "RequiresValidActionSlot"
	nativeRequiresValidAndPublicCVar                nativeFlag = "RequiresValidAndPublicCVar"
	nativeRequiresValidFontAsset                    nativeFlag = "RequiresValidFontAsset"
	nativeRequiresValidFontHeight                   nativeFlag = "RequiresValidFontHeight"
	nativeRequiresValidInviteTarget                 nativeFlag = "RequiresValidInviteTarget"
	nativeRequiresValidTimelineEvent                nativeFlag = "RequiresValidTimelineEvent"
	nativeRequiresValidUnitAuraInstance             nativeFlag = "RequiresValidUnitAuraInstance"
	nativeRestrictedForMacroChatMessages            nativeFlag = "RestrictedForMacroChatMessages"
	nativeReturnsNeverSecret                        nativeFlag = "ReturnsNeverSecret"
	nativeSecretInActivePvPMatch                    nativeFlag = "SecretInActivePvPMatch"
	nativeSecretInChatMessagingLockdown             nativeFlag = "SecretInChatMessagingLockdown"
	nativeSecretPayloads                            nativeFlag = "SecretPayloads"
	nativeSecretReturns                             nativeFlag = "SecretReturns"
	nativeSecretWhenAnchoringSecret                 nativeFlag = "SecretWhenAnchoringSecret"
	nativeSecretWhenAurasRestricted                 nativeFlag = "SecretWhenAurasRestricted"
	nativeSecretWhenCooldownsRestricted             nativeFlag = "SecretWhenCooldownsRestricted"
	nativeSecretWhenCurveSecret                     nativeFlag = "SecretWhenCurveSecret"
	nativeSecretWhenEncounterEvent                  nativeFlag = "SecretWhenEncounterEvent"
	nativeSecretWhenInCombat                        nativeFlag = "SecretWhenInCombat"
	nativeSecretWhenLossOfControlInfoRestricted     nativeFlag = "SecretWhenLossOfControlInfoRestricted"
	nativeSecretWhenLuaTableHasSecretKeys           nativeFlag = "SecretWhenLuaTableHasSecretKeys"
	nativeSecretWhenNumericFormatterSecret          nativeFlag = "SecretWhenNumericFormatterSecret"
	nativeSecretWhenTotemSlotSecret                 nativeFlag = "SecretWhenTotemSlotSecret"
	nativeSecretWhenUnitAuraRestricted              nativeFlag = "SecretWhenUnitAuraRestricted"
	nativeSecretWhenUnitComparisonRestricted        nativeFlag = "SecretWhenUnitComparisonRestricted"
	nativeSecretWhenUnitHealthMaxRestricted         nativeFlag = "SecretWhenUnitHealthMaxRestricted"
	nativeSecretWhenUnitIdentityRestricted          nativeFlag = "SecretWhenUnitIdentityRestricted"
	nativeSecretWhenUnitNameIdentityRestricted      nativeFlag = "SecretWhenUnitNameIdentityRestricted"
	nativeSecretWhenUnitPossessionRestricted        nativeFlag = "SecretWhenUnitPossessionRestricted"
	nativeSecretWhenUnitPowerMaxRestricted          nativeFlag = "SecretWhenUnitPowerMaxRestricted"
	nativeSecretWhenUnitPowerRestricted             nativeFlag = "SecretWhenUnitPowerRestricted"
	nativeSecretWhenUnitSpellCastRestricted         nativeFlag = "SecretWhenUnitSpellCastRestricted"
	nativeSecretWhenUnitStatsRestricted             nativeFlag = "SecretWhenUnitStatsRestricted"
	nativeSecretWhenUnitThreatStateRestricted       nativeFlag = "SecretWhenUnitThreatStateRestricted"
	nativeSecretWhenUnitThreatValuesRestricted      nativeFlag = "SecretWhenUnitThreatValuesRestricted"
	nativeSecureHooksAllowed                        nativeFlag = "SecureHooksAllowed"
	nativeSynchronousEvent                          nativeFlag = "SynchronousEvent"
	nativeTransliteratorAllowed                     nativeFlag = "TransliteratorAllowed"
	nativeUniqueEvent                               nativeFlag = "UniqueEvent"
)

type nativeFlagCategory uint8

const (
	nativeRestriction nativeFlagCategory = 1 << iota
	nativeSecretResult
)

// This table owns Boolean policy acceptance and semantic classification.
var nativeFlags = map[nativeFlag]nativeFlagCategory{
	nativeCallbackEvent:                             0,
	nativeCheckAllowChangeParent:                    nativeRestriction,
	nativeCheckAllowInheritForbiddenLayoutAspects:   nativeRestriction,
	nativeConstSecretAccessor:                       nativeRestriction,
	nativeHasRestrictions:                           nativeRestriction,
	nativeIsProtectedFunction:                       nativeRestriction,
	nativeMayReturnNothing:                          0,
	nativeMouseFocusValidForLimitedInput:            0,
	nativeRequireNPERestricted:                      nativeRestriction,
	nativeRequiresActiveCommentator:                 nativeRestriction,
	nativeRequiresAssignableScript:                  nativeRestriction,
	nativeRequiresCanChangeHitTestPoints:            nativeRestriction,
	nativeRequiresClubsInitialized:                  nativeRestriction,
	nativeRequiresCommentator:                       nativeRestriction,
	nativeRequiresComparableUnitTokens:              nativeRestriction,
	nativeRequiresDeclassifiedUnitIdentity:          nativeRestriction,
	nativeRequiresFontStringTextAccess:              nativeRestriction,
	nativeRequiresFriendList:                        nativeRestriction,
	nativeRequiresIndexInRange:                      nativeRestriction,
	nativeRequiresLimitedInput:                      nativeRestriction,
	nativeRequiresNonReadOnlyCVar:                   nativeRestriction,
	nativeRequiresNonSecretAura:                     nativeRestriction,
	nativeRequiresNonSecureCVar:                     nativeRestriction,
	nativeRequiresRecentAllies:                      nativeRestriction,
	nativeRequiresRestrictedAbbreviationBreakpoints: nativeRestriction,
	nativeRequiresScriptObjectAlphaAccess:           nativeRestriction,
	nativeRequiresScriptObjectDesaturationAccess:    nativeRestriction,
	nativeRequiresSpellDiminishUI:                   nativeRestriction,
	nativeRequiresStatusBarDesaturationAccess:       nativeRestriction,
	nativeRequiresSupportedScript:                   nativeRestriction,
	nativeRequiresTimedSignalMapAccess:              nativeRestriction,
	nativeRequiresUnitAuraAccess:                    nativeRestriction,
	nativeRequiresUnitIdentityAccess:                nativeRestriction,
	nativeRequiresValidAbbreviationBreakpoints:      nativeRestriction,
	nativeRequiresValidActionSlot:                   nativeRestriction,
	nativeRequiresValidAndPublicCVar:                nativeRestriction,
	nativeRequiresValidFontAsset:                    nativeRestriction,
	nativeRequiresValidFontHeight:                   nativeRestriction,
	nativeRequiresValidInviteTarget:                 nativeRestriction,
	nativeRequiresValidTimelineEvent:                nativeRestriction,
	nativeRequiresValidUnitAuraInstance:             nativeRestriction,
	nativeRestrictedForMacroChatMessages:            nativeRestriction,
	nativeReturnsNeverSecret:                        nativeRestriction,
	nativeSecretInActivePvPMatch:                    nativeRestriction | nativeSecretResult,
	nativeSecretInChatMessagingLockdown:             nativeRestriction | nativeSecretResult,
	nativeSecretPayloads:                            nativeRestriction,
	nativeSecretReturns:                             nativeRestriction | nativeSecretResult,
	nativeSecretWhenAnchoringSecret:                 nativeRestriction | nativeSecretResult,
	nativeSecretWhenAurasRestricted:                 nativeRestriction | nativeSecretResult,
	nativeSecretWhenCooldownsRestricted:             nativeRestriction | nativeSecretResult,
	nativeSecretWhenCurveSecret:                     nativeRestriction | nativeSecretResult,
	nativeSecretWhenEncounterEvent:                  nativeRestriction | nativeSecretResult,
	nativeSecretWhenInCombat:                        nativeRestriction | nativeSecretResult,
	nativeSecretWhenLossOfControlInfoRestricted:     nativeRestriction | nativeSecretResult,
	nativeSecretWhenLuaTableHasSecretKeys:           nativeRestriction | nativeSecretResult,
	nativeSecretWhenNumericFormatterSecret:          nativeRestriction | nativeSecretResult,
	nativeSecretWhenTotemSlotSecret:                 nativeRestriction | nativeSecretResult,
	nativeSecretWhenUnitAuraRestricted:              nativeRestriction | nativeSecretResult,
	nativeSecretWhenUnitComparisonRestricted:        nativeRestriction | nativeSecretResult,
	nativeSecretWhenUnitHealthMaxRestricted:         nativeRestriction | nativeSecretResult,
	nativeSecretWhenUnitIdentityRestricted:          nativeRestriction | nativeSecretResult,
	nativeSecretWhenUnitNameIdentityRestricted:      nativeRestriction | nativeSecretResult,
	nativeSecretWhenUnitPossessionRestricted:        nativeRestriction | nativeSecretResult,
	nativeSecretWhenUnitPowerMaxRestricted:          nativeRestriction | nativeSecretResult,
	nativeSecretWhenUnitPowerRestricted:             nativeRestriction | nativeSecretResult,
	nativeSecretWhenUnitSpellCastRestricted:         nativeRestriction | nativeSecretResult,
	nativeSecretWhenUnitStatsRestricted:             nativeRestriction | nativeSecretResult,
	nativeSecretWhenUnitThreatStateRestricted:       nativeRestriction | nativeSecretResult,
	nativeSecretWhenUnitThreatValuesRestricted:      nativeRestriction | nativeSecretResult,
	nativeSecureHooksAllowed:                        nativeRestriction,
	nativeSynchronousEvent:                          0,
	nativeTransliteratorAllowed:                     0,
	nativeUniqueEvent:                               0,
}

type NativePolicy struct {
	Flags                    map[nativeFlag]bool    `json:"Flags,omitzero"`
	AddsForbiddenAspects     []NativeAspectArgument `json:"AddsForbiddenAspects,omitzero"`
	ChecksForbiddenAspects   []NativeAspectArgument `json:"ChecksForbiddenAspects,omitzero"`
	Environment              *string                `json:"Environment,omitzero"`
	SecretArguments          *string                `json:"SecretArguments,omitzero"`
	SecretArgumentsAddAspect []NativeReference      `json:"SecretArgumentsAddAspect,omitzero"`
	SecretReturnsForAspect   []NativeReference      `json:"SecretReturnsForAspect,omitzero"`
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
	for _, name := range sortedKeys(raw) {
		flag := nativeFlag(name)
		if _, recognized := nativeFlags[flag]; !recognized {
			continue
		}
		value, err := nativeOptionalBool(raw, name)
		if err != nil {
			return policy, err
		}
		if policy.Flags == nil {
			policy.Flags = make(map[nativeFlag]bool)
		}
		policy.Flags[flag] = *value
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
	if value.EventNames, err = nativeItems(raw, "Events", decodeNativeEvent); err != nil {
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
	for _, event := range value.EventNames {
		if events[event] {
			return value, fmt.Errorf("duplicate native event %s", event)
		}
		events[event] = true
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
	if _, err = nativeKind(raw, "Function"); err != nil {
		return value, err
	}
	if _, err = nativeText(raw, "Namespace"); err != nil {
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
	if _, err = nativeText(raw, "Mixin"); err != nil {
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
func decodeNativeEvent(raw literal.Object) (name string, err error) {
	if err = nativeIdentity(raw); err != nil {
		return name, err
	}
	if err = nativeFields(raw, "Name Type LiteralName Documentation Payload", true); err != nil {
		return name, err
	}
	if name, err = nativeText(raw, "Name"); err != nil {
		return name, err
	}
	if _, err = nativeKind(raw, "Event"); err != nil {
		return name, err
	}
	if _, err = nativeText(raw, "LiteralName"); err != nil {
		return name, err
	}
	if err = validateNativeDocumentation(raw); err != nil {
		return name, err
	}
	if _, err = nativeItems(raw, "Payload", decodeNativeParameter); err != nil {
		return name, err
	}
	if _, err = decodeNativePolicy(raw); err != nil {
		return name, err
	}
	return name, nil
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
	if _, err = nativeItems(raw, "Values", decodeNativeParameter); err != nil {
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
	return NativeFunction{Name: table.Name, Arguments: table.Arguments, Returns: table.Returns, Policy: table.Policy}
}

// restrictions translates policy into the flat source vocabulary used by
// generated comments. Presence is retained, including explicit false values.
func (p NativePolicy) restrictions() literal.Object {
	out := literal.Object{}
	for flag, value := range p.Flags {
		if nativeFlags[flag]&nativeRestriction != 0 {
			out[string(flag)] = literal.Bool(value)
		}
	}
	if p.Environment != nil {
		out["Environment"] = literal.Text(*p.Environment)
	}
	if p.SecretArguments != nil {
		out["SecretArguments"] = literal.Text(*p.SecretArguments)
	}
	for name, references := range map[string][]NativeReference{
		"SecretArgumentsAddAspect": p.SecretArgumentsAddAspect,
		"SecretReturnsForAspect":   p.SecretReturnsForAspect,
	} {
		if references != nil {
			values := make([]literal.Value, len(references))
			for i, reference := range references {
				values[i] = literal.Record(literal.Object{"$ref": literal.Text(reference.Name)})
			}
			out[name] = literal.List(values)
		}
	}
	for name, arguments := range map[string][]NativeAspectArgument{
		"AddsForbiddenAspects":   p.AddsForbiddenAspects,
		"ChecksForbiddenAspects": p.ChecksForbiddenAspects,
	} {
		if arguments != nil {
			values := make([]literal.Value, len(arguments))
			for i, argument := range arguments {
				values[i] = literal.Record(literal.Object{
					"Argument": literal.Text(argument.Argument),
					"Aspect":   literal.Record(literal.Object{"$ref": literal.Text(argument.Aspect.Name)}),
				})
			}
			out[name] = literal.List(values)
		}
	}
	return out
}
func (p NativePolicy) secretResult() bool {
	if p.Flags[nativeReturnsNeverSecret] {
		return false
	}
	if len(p.SecretReturnsForAspect) > 0 {
		return true
	}
	for flag, enabled := range p.Flags {
		if enabled && nativeFlags[flag]&nativeSecretResult != 0 {
			return true
		}
	}
	return false
}
func (p NativePolicy) precondition(name string) bool {
	return p.Flags[nativeFlag(name)]
}
func (p NativeParameter) secretResult() bool {
	return !enabled(p.NeverSecret) && (enabled(p.SecretValue) || enabled(p.ConditionalSecret) || enabled(p.ConditionalSecretContents))
}
func (p NativeReference) MarshalJSON() ([]byte, error) {
	return jsonCompact(struct {
		Ref string `json:"$ref"`
	}{p.Name})
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
	if _, recognized := nativeFlags[nativeFlag(name)]; recognized {
		return true
	}
	switch name {
	case "AddsForbiddenAspects", "ChecksForbiddenAspects", "Environment", "SecretArguments", "SecretArgumentsAddAspect", "SecretReturnsForAspect":
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
