
WoWAPIProbeDB = {
["expectedBuild"] = "12.1.5.70575",
["capabilityCaptureComplete"] = false,
["verifiesSignatures"] = false,
["probeVersion"] = "1.1.0",
["verifiesSecrecy"] = false,
["expectedMethods"] = 73,
["sourceFingerprint"] = "668d281223c9f886b282cdba8fcf07a561075c722e208732ce2ea5669ef8dd99",
["outcome"] = "public_invocations_incomplete",
["limits"] = {
["colorCurve"] = 1,
["animationGroup"] = 1,
["invocationBindingCopies"] = 1,
["invocationCases"] = 73,
["duration"] = 1,
["methods"] = 73,
["returnTuple"] = 8,
["formatter"] = 1,
["hierarchyCases"] = 281,
["families"] = 7,
["timedSignalMap"] = 1,
["durationBinding"] = 1,
["savedCaptures"] = 1,
["queryCases"] = 56,
["localeContext"] = 1,
["manualClock"] = 1,
["vectorGraphics"] = 1,
["frame"] = 1,
["radialProgress"] = 1,
["fontString"] = 1,
},
["mode"] = "public_invocations",
["context"] = "out_of_combat_public_owned_objects",
["fixtureSHA256"] = "e933cd46074a01a9c4c718886bbb62fda666843449b6d8730b66c115a2ee219c",
["clientDate"] = "Oct 22 2026",
["declarationFingerprint"] = "3f4d79c8b5e98aa2d452e0de377a3135373d4c9fb524f79376fb6a522217ad0e",
["invocationCaptureComplete"] = false,
["verifiedMethods"] = 72,
["clientBuild"] = "70575",
["families"] = {
},
["invocations"] = {
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
},
["check"] = "clears_duration",
["caseID"] = "DurationTextBindingObject.SetToDefaults",
["returnCount"] = 0,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
},
["check"] = "duration_identity",
["caseID"] = "DurationTextBindingObject.SetDuration",
["returnCount"] = 0,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"userdata",
},
["check"] = "duration_identity",
["caseID"] = "DurationTextBindingObject.GetDuration",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
},
["check"] = "font_identity",
["caseID"] = "DurationTextBindingObject.SetFontString",
["returnCount"] = 0,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"table",
},
["check"] = "font_identity",
["caseID"] = "DurationTextBindingObject.GetFontString",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
},
["check"] = "can_format",
["caseID"] = "DurationTextBindingObject.SetFormatter",
["returnCount"] = 0,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
},
["check"] = "can_format_components",
["caseID"] = "DurationTextBindingObject.SetTextFormat",
["returnCount"] = 0,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"boolean",
},
["check"] = "configured_true",
["caseID"] = "DurationTextBindingObject.CanFormatText",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"boolean",
},
["check"] = "configured_true",
["caseID"] = "DurationTextBindingObject.CanUpdateFontString",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
},
["check"] = "text_roundtrip",
["caseID"] = "DurationTextBindingObject.SetExpiredText",
["returnCount"] = 0,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"string",
},
["check"] = "text_roundtrip",
["caseID"] = "DurationTextBindingObject.GetExpiredText",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
},
["check"] = "text_roundtrip",
["caseID"] = "DurationTextBindingObject.SetZeroDurationText",
["returnCount"] = 0,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"string",
},
["check"] = "text_roundtrip",
["caseID"] = "DurationTextBindingObject.GetZeroDurationText",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
},
["check"] = "modifier_roundtrip",
["caseID"] = "DurationTextBindingObject.SetTimeModifier",
["returnCount"] = 0,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"number",
},
["check"] = "modifier_roundtrip",
["caseID"] = "DurationTextBindingObject.GetTimeModifier",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
},
["check"] = "interval_roundtrip",
["caseID"] = "DurationTextBindingObject.SetUpdateInterval",
["returnCount"] = 0,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"number",
},
["check"] = "interval_roundtrip",
["caseID"] = "DurationTextBindingObject.GetUpdateInterval",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
},
["check"] = "curve_property_tuple",
["caseID"] = "DurationTextBindingObject.SetTextColorCurve",
["returnCount"] = 0,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"userdata",
"number",
},
["check"] = "curve_property_tuple",
["caseID"] = "DurationTextBindingObject.GetTextColorCurve",
["returnCount"] = 2,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"table",
},
["check"] = "public_white_rgba",
["caseID"] = "DurationTextBindingObject.GetFormattedTextColor",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
},
["check"] = "cleared_curve_tuple",
["caseID"] = "DurationTextBindingObject.ClearTextColorCurve",
["returnCount"] = 0,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"string",
},
["check"] = "public_nonempty_text",
["caseID"] = "DurationTextBindingObject.GetFormattedText",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "property_mismatch",
["behaviorVerified"] = false,
["returns"] = {
},
["check"] = "font_text_matches",
["caseID"] = "DurationTextBindingObject.UpdateFontString",
["returnCount"] = 0,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"boolean",
},
["check"] = "public_fixture_false",
["caseID"] = "DurationTextBindingObject.HasSecretValues",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
},
["check"] = "enabled_true",
["caseID"] = "DurationTextBindingObject.Enable",
["returnCount"] = 0,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"boolean",
},
["check"] = "enabled_true",
["caseID"] = "DurationTextBindingObject.IsEnabled",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
},
["check"] = "enabled_false",
["caseID"] = "DurationTextBindingObject.SetEnabled",
["returnCount"] = 0,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
},
["check"] = "enabled_false",
["caseID"] = "DurationTextBindingObject.Disable",
["returnCount"] = 0,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"userdata",
},
["check"] = "copy_preserves_text",
["caseID"] = "DurationTextBindingObject.Copy",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
},
["check"] = "assignment_preserves_text",
["caseID"] = "DurationTextBindingObject.Assign",
["returnCount"] = 0,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"number",
},
["check"] = "manual_time_12_5",
["caseID"] = "LuaDurationClockObject.GetTime",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
},
["check"] = "manual_time_12_5",
["caseID"] = "LuaDurationManualClock.SetTime",
["returnCount"] = 0,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
},
["check"] = "manual_time_15_5",
["caseID"] = "LuaDurationManualClock.AdvanceTime",
["returnCount"] = 0,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
},
["check"] = "manual_time_14_5",
["caseID"] = "LuaDurationManualClock.RewindTime",
["returnCount"] = 0,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
},
["check"] = "manual_time_zero",
["caseID"] = "LuaDurationManualClock.ResetTime",
["returnCount"] = 0,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"boolean",
},
["check"] = "locale_accepted",
["caseID"] = "LuaLocaleContext.SetLocale",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"string",
},
["check"] = "en_us_locale",
["caseID"] = "LuaLocaleContext.GetLocale",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"number",
},
["check"] = "equal_collation_zero",
["caseID"] = "LuaLocaleContext.CompareStrings",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"table",
},
["check"] = "four_character_boundaries",
["caseID"] = "LuaLocaleContext.FindBreaks",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"table",
},
["check"] = "two_nonoverlapping_matches",
["caseID"] = "LuaLocaleContext.FindStringMatches",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"string",
},
["check"] = "ascii_fold",
["caseID"] = "LuaLocaleContext.FoldCase",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"string",
},
["check"] = "public_currency_text",
["caseID"] = "LuaLocaleContext.FormatCurrency",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"string",
},
["check"] = "public_date_text",
["caseID"] = "LuaLocaleContext.FormatDate",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"string",
},
["check"] = "public_datetime_text",
["caseID"] = "LuaLocaleContext.FormatDateTime",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"string",
},
["check"] = "ascii_decimal",
["caseID"] = "LuaLocaleContext.FormatNumber",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"string",
},
["check"] = "public_time_text",
["caseID"] = "LuaLocaleContext.FormatTime",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"string",
},
["check"] = "public_currency_name",
["caseID"] = "LuaLocaleContext.GetCurrencyName",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"string",
},
["check"] = "public_display_name",
["caseID"] = "LuaLocaleContext.GetDisplayName",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"string",
},
["check"] = "public_nonempty_sort_key",
["caseID"] = "LuaLocaleContext.GetSortKey",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"number",
},
["check"] = "three_characters",
["caseID"] = "LuaLocaleContext.Length",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"table",
},
["check"] = "currency_amount_code",
["caseID"] = "LuaLocaleContext.ParseCurrency",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"number",
},
["check"] = "decimal_parse",
["caseID"] = "LuaLocaleContext.ParseNumber",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"string",
},
["check"] = "singular_one",
["caseID"] = "LuaLocaleContext.SelectPlural",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"string",
},
["check"] = "ascii_lower",
["caseID"] = "LuaLocaleContext.ToLower",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"string",
},
["check"] = "ascii_title",
["caseID"] = "LuaLocaleContext.ToTitle",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"string",
},
["check"] = "ascii_upper",
["caseID"] = "LuaLocaleContext.ToUpper",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"string",
},
["check"] = "language_en",
["caseID"] = "LuaLocaleContext.TransformLocale",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
},
["check"] = "scheduled_key_present",
["caseID"] = "TimedSignalMap.SignalAt",
["returnCount"] = 0,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
},
["check"] = "scheduled_key_present",
["caseID"] = "TimedSignalMap.SignalAfter",
["returnCount"] = 0,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"boolean",
},
["check"] = "scheduled_key_present",
["caseID"] = "TimedSignalMap.HasSignal",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"number",
},
["check"] = "two_scheduled_keys",
["caseID"] = "TimedSignalMap.GetSignalCount",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"number",
},
["check"] = "absolute_deadline",
["caseID"] = "TimedSignalMap.GetSignalTime",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"number",
"number",
},
["check"] = "key_time_tuple",
["caseID"] = "TimedSignalMap.GetNextSignal",
["returnCount"] = 2,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
},
["check"] = "one_key_remains",
["caseID"] = "TimedSignalMap.CancelSignal",
["returnCount"] = 0,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
},
["check"] = "zero_keys_remain",
["caseID"] = "TimedSignalMap.CancelAllSignals",
["returnCount"] = 0,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
},
["check"] = "from_roundtrip",
["caseID"] = "RadialProgress.SetFromPercent",
["returnCount"] = 0,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
},
["check"] = "to_roundtrip",
["caseID"] = "RadialProgress.SetToPercent",
["returnCount"] = 0,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"number",
},
["check"] = "from_roundtrip",
["caseID"] = "RadialProgress.GetFromPercent",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"number",
},
["check"] = "to_roundtrip",
["caseID"] = "RadialProgress.GetToPercent",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"boolean",
},
["check"] = "owned_svg_accepted",
["caseID"] = "VectorGraphics.SetSVG",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"number",
},
["check"] = "public_fileid_carrier",
["caseID"] = "VectorGraphics.GetSVGFileID",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
"boolean",
},
["check"] = "owned_svg_present",
["caseID"] = "VectorGraphics.HasSVG",
["returnCount"] = 1,
["invoked"] = true,
},
{
["state"] = "passed",
["behaviorVerified"] = true,
["returns"] = {
},
["check"] = "svg_cleared",
["caseID"] = "VectorGraphics.ClearSVG",
["returnCount"] = 0,
["invoked"] = true,
},
},
["cleanup"] = {
{
["state"] = "complete",
["caseID"] = "cleanup.invocation.binding.1",
},
{
["state"] = "complete",
["caseID"] = "cleanup.invocation.binding.2",
},
{
["state"] = "complete",
["caseID"] = "cleanup.invocation.signals",
},
},
["presentMethods"] = 0,
["observedMethods"] = 0,
["clientInterface"] = 120105,
["hierarchy"] = {
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.AddAccessRestrictions",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.AddForbiddenAspects",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.AddSecretAspect",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.CanBeAccessedInContext",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.ClearParentKey",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.GetAccessRestrictions",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.GetDebugName",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.GetDuration",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.GetElapsed",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.GetEndDelay",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.GetForbiddenAspects",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.GetInheritableForbiddenAspects",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.GetName",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.GetObjectTable",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.GetObjectType",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.GetOrder",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.GetParent",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.GetParentKey",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.GetProgress",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.GetRegionParent",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.GetScript",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.GetSmoothProgress",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.GetSmoothing",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.GetStartDelay",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.GetTarget",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.HasAccessConstraints",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.HasAnyAccessRestrictions",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.HasAnyForbiddenAspects",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.HasAnySecretAspect",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.HasScript",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.HasSecretAspect",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.HasSecretValues",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.HookScript",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.IsDelaying",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.IsDone",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.IsForbidden",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.IsObjectType",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.IsPaused",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.IsPlaying",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.IsPreventingSecretValues",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.IsStopped",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.Pause",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.Play",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.Restart",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.SetChildKey",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.SetDuration",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.SetEndDelay",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.SetForbidden",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.SetOrder",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.SetParent",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.SetParentKey",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.SetPlaying",
},
{
["state"] = "absent",
["caseID"] = "RadialProgress.parent.Animation.SetPreventSecretValues",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.SetScript",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.SetSmoothProgress",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.SetSmoothing",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.SetStartDelay",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.SetTarget",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.SetTargetKey",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.SetTargetName",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.SetTargetParent",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.SetToDefaults",
},
{
["state"] = "present",
["caseID"] = "RadialProgress.parent.Animation.Stop",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.AddAccessRestrictions",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.AddForbiddenAspects",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.AddSecretAspect",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.AdjustPointsOffset",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.CanBeAccessedInContext",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.CanChangeProtectedState",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.CanPropagateMouseClicks",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.CanPropagateMouseMotion",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.ClearAllPoints",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.ClearParentKey",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.ClearPoint",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.ClearPointsOffset",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.ClearScripts",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.CollapsesLayout",
},
{
["state"] = "absent",
["caseID"] = "VectorGraphics.parent.Region.CreateAnimationGroup",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.EnableMouse",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.EnableMouseMotion",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.EnableMouseWheel",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.GetAccessRestrictions",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.GetAlpha",
},
{
["state"] = "absent",
["caseID"] = "VectorGraphics.parent.Region.GetAnimationGroups",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.GetBottom",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.GetCenter",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.GetDebugName",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.GetDrawLayer",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.GetEffectiveScale",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.GetForbiddenAspects",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.GetHeight",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.GetInheritableForbiddenAspects",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.GetLeft",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.GetName",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.GetNumPoints",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.GetObjectTable",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.GetObjectType",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.GetParent",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.GetParentKey",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.GetPoint",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.GetPointByName",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.GetRect",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.GetRight",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.GetRoundLayoutToNearestPixel",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.GetScale",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.GetScaledRect",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.GetScript",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.GetSize",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.GetSourceLocation",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.GetTop",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.GetVertexColor",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.GetWidth",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.HasAccessConstraints",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.HasAnyAccessRestrictions",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.HasAnyForbiddenAspects",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.HasAnySecretAspect",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.HasScript",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.HasSecretAspect",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.HasSecretValues",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.Hide",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.HookScript",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.Intersects",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.IsAnchoringRestricted",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.IsAnchoringSecret",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.IsCollapsed",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.IsDragging",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.IsForbidden",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.IsIgnoringParentAlpha",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.IsIgnoringParentScale",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.IsMouseClickEnabled",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.IsMouseEnabled",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.IsMouseMotionEnabled",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.IsMouseMotionFocus",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.IsMouseOver",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.IsMouseWheelEnabled",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.IsObjectLoaded",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.IsObjectType",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.IsPreventingSecretValues",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.IsProtected",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.IsRectValid",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.IsShown",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.IsVisible",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.SetAllPoints",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.SetAlpha",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.SetAlphaFromBoolean",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.SetCollapsesLayout",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.SetDrawLayer",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.SetForbidden",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.SetHeight",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.SetIgnoreParentAlpha",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.SetIgnoreParentScale",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.SetMouseClickEnabled",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.SetMouseMotionEnabled",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.SetParent",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.SetParentKey",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.SetPassThroughButtons",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.SetPoint",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.SetPointsOffset",
},
{
["state"] = "absent",
["caseID"] = "VectorGraphics.parent.Region.SetPreventSecretValues",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.SetPropagateMouseClicks",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.SetPropagateMouseMotion",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.SetRoundLayoutToNearestPixel",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.SetScale",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.SetScript",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.SetShown",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.SetSize",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.SetToDefaults",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.SetVertexColor",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.SetVertexColorFromBoolean",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.SetWidth",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.ShouldButtonPassThrough",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.Region.Show",
},
{
["state"] = "absent",
["caseID"] = "VectorGraphics.parent.Region.StopAnimating",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.AddAccessRestrictions",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.AddForbiddenAspects",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.AddSecretAspect",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.AdjustPointsOffset",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.CanBeAccessedInContext",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.CanChangeProtectedState",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.CanPropagateMouseClicks",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.CanPropagateMouseMotion",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.ClearAllPoints",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.ClearParentKey",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.ClearPoint",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.ClearPointsOffset",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.ClearScripts",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.CollapsesLayout",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.EnableMouse",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.EnableMouseMotion",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.EnableMouseWheel",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.GetAccessRestrictions",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.GetAlpha",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.GetBottom",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.GetCenter",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.GetDebugName",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.GetDrawLayer",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.GetEffectiveScale",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.GetForbiddenAspects",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.GetHeight",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.GetInheritableForbiddenAspects",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.GetLeft",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.GetName",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.GetNumPoints",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.GetObjectTable",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.GetObjectType",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.GetParent",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.GetParentKey",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.GetPoint",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.GetPointByName",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.GetRect",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.GetRight",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.GetRoundLayoutToNearestPixel",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.GetScale",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.GetScaledRect",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.GetSize",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.GetSourceLocation",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.GetTop",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.GetVertexColor",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.GetWidth",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.HasAccessConstraints",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.HasAnyAccessRestrictions",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.HasAnyForbiddenAspects",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.HasAnySecretAspect",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.HasSecretAspect",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.HasSecretValues",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.Hide",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.Intersects",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.IsAnchoringRestricted",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.IsAnchoringSecret",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.IsCollapsed",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.IsDragging",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.IsForbidden",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.IsIgnoringParentAlpha",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.IsIgnoringParentScale",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.IsMouseClickEnabled",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.IsMouseEnabled",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.IsMouseMotionEnabled",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.IsMouseMotionFocus",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.IsMouseOver",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.IsMouseWheelEnabled",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.IsObjectLoaded",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.IsObjectType",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.IsPreventingSecretValues",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.IsProtected",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.IsRectValid",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.IsShown",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.IsVisible",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.SetAllPoints",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.SetAlpha",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.SetAlphaFromBoolean",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.SetCollapsesLayout",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.SetDrawLayer",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.SetForbidden",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.SetHeight",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.SetIgnoreParentAlpha",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.SetIgnoreParentScale",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.SetMouseClickEnabled",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.SetMouseMotionEnabled",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.SetParent",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.SetParentKey",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.SetPassThroughButtons",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.SetPoint",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.SetPointsOffset",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.SetPropagateMouseClicks",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.SetPropagateMouseMotion",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.SetRoundLayoutToNearestPixel",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.SetScale",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.SetShown",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.SetSize",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.SetToDefaults",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.SetVertexColor",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.SetVertexColorFromBoolean",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.SetWidth",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.ShouldButtonPassThrough",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.LayeredRegion.Show",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.ScriptObject.GetScript",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.ScriptObject.HasScript",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.ScriptObject.HookScript",
},
{
["state"] = "present",
["caseID"] = "VectorGraphics.parent.wowless.ScriptObject.SetScript",
},
{
["state"] = "absent",
["caseID"] = "VectorGraphics.parent.ketho.RotationCandidates.GetRotation",
},
{
["state"] = "absent",
["caseID"] = "VectorGraphics.parent.ketho.RotationCandidates.SetRotation",
},
},
["invokedMethods"] = 73,
["clientVersion"] = "12.1.5",
["schema"] = 1,
["projectID"] = 1,
}
