package wowapi

// Static explanatory text and call-site discriminators from the maintained policy.
type nativeCallsite struct {
	Name, Snippet string
	ArgumentCount int
	ArgumentRoles []string
}

var nativeCallsites = []nativeCallsite{
	{"DurationTextBindingObject.Assign", "binding:Assign(options.binding);", 1, []string{"existing duration text binding from processed options"}},
	{"DurationTextBindingObject.SetToDefaults", "binding:SetToDefaults();", 0, []string{}},
	{"DurationTextBindingObject.SetFormatter", "binding:SetFormatter(addonTable.DefaultAuraDurationFormatter);", 1, []string{"default aura duration formatter"}},
	{"DurationTextBindingObject.SetFontString", "binding:SetFontString(fontString);", 1, []string{"FontString validated by RequireObjectType before initialization"}},
	{"DurationTextBindingObject.SetDuration", "binding:SetDuration(self:GetAuraDuration());", 1, []string{"duration returned by GetAuraDuration"}},
	{"DurationTextBindingObject.SetTextFormat", "binding:SetTextFormat(options.textFormat.formatString, options.textFormat.components);", 2, []string{"format string", "ordered component collection"}},
	{"DurationTextBindingObject.SetTextColorCurve", "binding:SetTextColorCurve(options.textColor.curve, options.textColor.property);", 2, []string{"color curve", "selected duration property"}},
	{"DurationTextBindingObject.SetEnabled", "self:GetDurationTextBinding():SetEnabled(false);", 1, []string{"literal Boolean false"}},
	{"TimedSignalMap.SignalAt", "AuraButtonTimedSignalMap:SignalAt(pandemicUpdateSignal, pandemicStartTime);", 2, []string{"signal identity", "pandemic deadline compared with GetTime"}},
	{"TimedSignalMap.CancelSignal", "AuraButtonTimedSignalMap:CancelSignal(pandemicUpdateSignal);", 1, []string{"the same signal identity used by SignalAt"}},
}

const nativeCallsiteRequiredJSON = `["local binding = self:GetDurationTextBinding();","self.durationTextBinding = C_DurationUtil.CreateDurationTextBinding();","AuraContainerUtil.RequireObjectType(\"FontString\")","if timeNow < pandemicStartTime then"]`
const nativeCoverageJSON = `{"corroborated":["receiver/method existence","ordered input/output carriers","nilability","default presence","explicit MayReturnNothing"],"native_observation_scope":"Recorded public-fixture invocations and receiver capability lookups apply only to the exact matched build/source cohort. Historical raw mismatches remain explicit; a separately reviewed later capture may establish the bounded behavior without rewriting earlier observations.","not_independently_established":["native secret-value behavior and permissions","complete inherited method signatures","contexts beyond recorded public owned fixtures","full contract from a static call site"]}`
const tsmCoverageDescriptionsJSON = `{"compared":["native functions and mapped object methods","all typed duplicate namespace variants","ordered arguments/results and ABI tuples","carrier aliases and structural callback/generic/pack/intersection facts","recursive secret wrapper locations","preconditions, absence, aspects and scoped argument permissions","every native emitted contract and full declaration graph","all seven TSM-omitted object families against a pinned secondary translation"],"limits":["Resolved bindings establish primary selection and preserve serialized policies/overloads for 262 formerly ambiguous symbols; general post-scan assignability is not established.","Secondary object signatures derive from Blizzard documentation and do not independently verify native secrecy or permissions.","Malformed upstream declarations retain explicit source-defect findings, not a fabricated valid function contract.","General assignability and runtime generic inference are not established by structural source comparison.","Aggregate field/event payload runtime accessibility is not proven by source comparison.","Duplicate overloads do not merge function-level secret permissions."]}`
