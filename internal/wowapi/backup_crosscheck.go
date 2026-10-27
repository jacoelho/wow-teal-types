package wowapi

import (
	"fmt"
	"path/filepath"
	"strings"
	"wow-teal-types/internal/sourcelock"
	"wow-teal-types/internal/wowapi/literal"
)

type metadataDifference struct {
	Path           string         `json:"path"`
	Kind           string         `json:"kind"`
	Primary        *literal.Value `json:"primary,omitempty"`
	Backup         *literal.Value `json:"backup,omitempty"`
	PrimaryPresent *bool          `json:"primary_present,omitempty"`
	BackupPresent  *bool          `json:"backup_present,omitempty"`
}

func compareMetadata(primary, backup literal.Value, path string) []metadataDifference {
	if primary.Kind() != backup.Kind() {
		return []metadataDifference{{Path: path, Kind: "type", Primary: &primary, Backup: &backup}}
	}
	switch primary.Kind() {
	case literal.ObjectKind:
		a, _ := primary.Object()
		b, _ := backup.Object()
		names := map[string]bool{}
		for k := range a {
			names[k] = true
		}
		for k := range b {
			names[k] = true
		}
		diff := []metadataDifference{}
		for _, key := range sortedKeys(names) {
			where := key
			if path != "" {
				where = path + "." + key
			}
			left, lp := a[key]
			right, rp := b[key]
			if !lp || !rp {
				diff = append(diff, metadataDifference{Path: where, Kind: "presence", PrimaryPresent: &lp, BackupPresent: &rp})
			} else {
				diff = append(diff, compareMetadata(left, right, where)...)
			}
		}
		return diff
	case literal.Array:
		a, _ := primary.Array()
		b, _ := backup.Array()
		diff := []metadataDifference{}
		if len(a) != len(b) {
			left, right := literal.Int(int64(len(a))), literal.Int(int64(len(b)))
			diff = append(diff, metadataDifference{Path: path, Kind: "length", Primary: &left, Backup: &right})
		}
		for i := 0; i < min(len(a), len(b)); i++ {
			diff = append(diff, compareMetadata(a[i], b[i], fmt.Sprintf("%s[%d]", path, i+1))...)
		}
		return diff
	}
	if primary.Equal(backup) {
		return []metadataDifference{}
	}
	return []metadataDifference{{Path: path, Kind: "value", Primary: &primary, Backup: &backup}}
}
func (p Project) loadBackup() (sourcelock.Backup, map[string]NativeDocument, error) {
	lock, e := sourcelock.ReadBackup(filepath.Join(p.Root, "scripts/wow_api/native_backup_sources.json"))
	if e != nil {
		return lock, nil, e
	}
	if lock.SchemaVersion != 1 {
		return lock, nil, fmt.Errorf("unsupported native backup schema")
	}
	primary, e := sourcelock.ReadSources(filepath.Join(p.Root, "scripts/wow_api/sources.json"))
	if e != nil {
		return lock, nil, e
	}
	type fileIdentity struct {
		SHA  string
		Size int
	}
	expected := map[string]fileIdentity{}
	for _, row := range primary["blizzard"].Files {
		expected[row.Name] = fileIdentity{row.SHA, row.Size}
	}
	inventory := map[string]fileIdentity{}
	for _, row := range lock.Inventory {
		name := row.Name
		if _, present := inventory[name]; present || !evidenceSafeRelative(name) {
			return lock, nil, fmt.Errorf("duplicate or unsafe native backup inventory: %s", name)
		}
		inventory[name] = fileIdentity{row.SHA, row.Size}
	}
	if !evEqual(inventory, expected) {
		return lock, nil, fmt.Errorf("native backup complete metadata inventory differs from primary")
	}
	version, e := readEvidenceFile(p.InputPath("blizzard_version", "version.txt"), 1024)
	if e != nil {
		return lock, nil, e
	}
	versionRows := primary["blizzard_version"].Files
	if len(versionRows) != 1 || blobSHA(version) != versionRows[0].SHA {
		return lock, nil, fmt.Errorf("native backup primary version integrity failure")
	}
	if lock.ClaimedBuild == nil || *lock.ClaimedBuild != strings.TrimSpace(string(version)) {
		return lock, nil, fmt.Errorf("native backup build provenance differs from primary")
	}
	documents := map[string]NativeDocument{}
	for _, row := range lock.Files {
		name := row.Name
		exp, exists := expected[name]
		if _, present := documents[name]; present || !exists || row.SHA != exp.SHA || row.Size != exp.Size {
			return lock, nil, fmt.Errorf("invalid native backup file inventory: %s", name)
		}
		data, e := readEvidenceFile(filepath.Join(p.Root, "references/wow-api/native-backup", name), 1<<20)
		if e != nil {
			return lock, nil, e
		}
		if len(data) != row.Size || blobSHA(data) != row.SHA || sha256hex(data) != row.SHA256 {
			return lock, nil, fmt.Errorf("native backup integrity failure: %s", name)
		}
		raw, e := ParseLua(data, name)
		if e != nil {
			return lock, nil, e
		}
		metadata, e := decodeNativeMetadata(raw)
		if e != nil {
			return lock, nil, e
		}
		documents[name] = NativeDocument{Name: name, Raw: raw, Metadata: metadata}
	}
	return lock, documents, nil
}
func corroborateBackup(documents map[string]NativeDocument, docs []NativeDocument) (backupContext, error) {
	source := map[string]NativeDocument{}
	for _, doc := range docs {
		source[doc.Name] = doc
	}
	for name, backup := range documents {
		original, ok := source[name]
		if !ok || len(compareMetadata(literal.Record(original.Raw), literal.Record(backup.Raw), "")) > 0 {
			return backupContext{}, fmt.Errorf("native backup metadata differs: %s", name)
		}
	}
	for _, name := range []string{"TimedSignalMapSharedDocumentation.lua", "SecretPredicatesDocumentation.lua", "SecretAspectConstantsDocumentation.lua"} {
		if _, ok := documents[name]; !ok {
			return backupContext{}, fmt.Errorf("native backup lacks predicate/callback context")
		}
	}
	predicates, backupPredicates := map[string]NativePredicate{}, map[string]NativePredicate{}
	definitions, backupDefinitions := map[string]NativeTable{}, map[string]NativeTable{}
	for _, doc := range source {
		for _, p := range doc.Metadata.Predicates {
			predicates[p.Name] = p
		}
		for _, d := range doc.Metadata.Tables {
			definitions[d.Name] = d
		}
	}
	for _, doc := range documents {
		for _, p := range doc.Metadata.Predicates {
			backupPredicates[p.Name] = p
		}
		for _, d := range doc.Metadata.Tables {
			backupDefinitions[d.Name] = d
		}
	}
	pending := []NativeParameter{}
	for _, doc := range documents {
		for _, method := range doc.Metadata.Functions {
			pending = append(pending, method.Arguments...)
			pending = append(pending, method.Returns...)
			for _, key := range backupPolicyPredicates(method.Policy) {
				primary, present := predicates[key]
				backup, backed := backupPredicates[key]
				if !present || !backed || len(compareMetadata(literal.Record(primary.Raw), literal.Record(backup.Raw), "")) > 0 {
					return backupContext{}, fmt.Errorf("native backup lacks matching referenced predicate: %s", key)
				}
			}
		}
	}
	seen := map[string]bool{}
	for len(pending) > 0 {
		value := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		for _, name := range []string{value.Type, value.InnerType} {
			definition, ok := definitions[name]
			if !ok || seen[name] {
				continue
			}
			seen[name] = true
			if len(seen) > 2048 {
				return backupContext{}, fmt.Errorf("native backup dependency graph exceeds type bound")
			}
			pending = append(pending, definition.Arguments...)
			pending = append(pending, definition.Fields...)
			pending = append(pending, definition.Values...)
			if definition.Type == "CallbackType" {
				backup, present := backupDefinitions[name]
				if !present || len(compareMetadata(literal.Record(definition.Raw), literal.Record(backup.Raw), "")) > 0 {
					return backupContext{}, fmt.Errorf("native backup lacks matching referenced callback: %s", name)
				}
			}
		}
	}
	return backupContext{Documents: sortedKeys(documents), DefinitionComparison: "exact", Scope: "separately maintained metadata distribution; not independent native semantics"}, nil
}

func backupPolicyPredicates(policy NativePolicy) []string {
	out := []string{}
	if policy.ChecksForbiddenAspects != nil {
		out = append(out, "ChecksForbiddenAspects")
	}
	if policy.CheckAllowChangeParent != nil {
		out = append(out, "CheckAllowChangeParent")
	}
	if policy.CheckAllowInheritForbiddenLayoutAspects != nil {
		out = append(out, "CheckAllowInheritForbiddenLayoutAspects")
	}
	if policy.RequiresActiveCommentator != nil {
		out = append(out, "RequiresActiveCommentator")
	}
	if policy.RequiresAssignableScript != nil {
		out = append(out, "RequiresAssignableScript")
	}
	if policy.RequiresCanChangeHitTestPoints != nil {
		out = append(out, "RequiresCanChangeHitTestPoints")
	}
	if policy.RequiresClubsInitialized != nil {
		out = append(out, "RequiresClubsInitialized")
	}
	if policy.RequiresCommentator != nil {
		out = append(out, "RequiresCommentator")
	}
	if policy.RequiresComparableUnitTokens != nil {
		out = append(out, "RequiresComparableUnitTokens")
	}
	if policy.RequiresDeclassifiedUnitIdentity != nil {
		out = append(out, "RequiresDeclassifiedUnitIdentity")
	}
	if policy.RequiresFontStringTextAccess != nil {
		out = append(out, "RequiresFontStringTextAccess")
	}
	if policy.RequiresFriendList != nil {
		out = append(out, "RequiresFriendList")
	}
	if policy.RequiresIndexInRange != nil {
		out = append(out, "RequiresIndexInRange")
	}
	if policy.RequiresLimitedInput != nil {
		out = append(out, "RequiresLimitedInput")
	}
	if policy.RequiresNonReadOnlyCVar != nil {
		out = append(out, "RequiresNonReadOnlyCVar")
	}
	if policy.RequiresNonSecretAura != nil {
		out = append(out, "RequiresNonSecretAura")
	}
	if policy.RequiresNonSecureCVar != nil {
		out = append(out, "RequiresNonSecureCVar")
	}
	if policy.RequiresRecentAllies != nil {
		out = append(out, "RequiresRecentAllies")
	}
	if policy.RequiresRestrictedAbbreviationBreakpoints != nil {
		out = append(out, "RequiresRestrictedAbbreviationBreakpoints")
	}
	if policy.RequiresScriptObjectAlphaAccess != nil {
		out = append(out, "RequiresScriptObjectAlphaAccess")
	}
	if policy.RequiresScriptObjectDesaturationAccess != nil {
		out = append(out, "RequiresScriptObjectDesaturationAccess")
	}
	if policy.RequiresSpellDiminishUI != nil {
		out = append(out, "RequiresSpellDiminishUI")
	}
	if policy.RequiresStatusBarDesaturationAccess != nil {
		out = append(out, "RequiresStatusBarDesaturationAccess")
	}
	if policy.RequiresSupportedScript != nil {
		out = append(out, "RequiresSupportedScript")
	}
	if policy.RequiresTimedSignalMapAccess != nil {
		out = append(out, "RequiresTimedSignalMapAccess")
	}
	if policy.RequiresUnitAuraAccess != nil {
		out = append(out, "RequiresUnitAuraAccess")
	}
	if policy.RequiresUnitIdentityAccess != nil {
		out = append(out, "RequiresUnitIdentityAccess")
	}
	if policy.RequiresValidAbbreviationBreakpoints != nil {
		out = append(out, "RequiresValidAbbreviationBreakpoints")
	}
	if policy.RequiresValidActionSlot != nil {
		out = append(out, "RequiresValidActionSlot")
	}
	if policy.RequiresValidAndPublicCVar != nil {
		out = append(out, "RequiresValidAndPublicCVar")
	}
	if policy.RequiresValidFontAsset != nil {
		out = append(out, "RequiresValidFontAsset")
	}
	if policy.RequiresValidFontHeight != nil {
		out = append(out, "RequiresValidFontHeight")
	}
	if policy.RequiresValidInviteTarget != nil {
		out = append(out, "RequiresValidInviteTarget")
	}
	if policy.RequiresValidTimelineEvent != nil {
		out = append(out, "RequiresValidTimelineEvent")
	}
	if policy.RequiresValidUnitAuraInstance != nil {
		out = append(out, "RequiresValidUnitAuraInstance")
	}
	return out
}
