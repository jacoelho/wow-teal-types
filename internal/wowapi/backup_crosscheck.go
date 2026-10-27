package wowapi

import (
	"fmt"
	"path/filepath"
	"strings"
)

func compareMetadata(primary, backup any, path string) []any {
	if metadataKind(primary) != metadataKind(backup) {
		return []any{M{"path": path, "kind": "type", "primary": primary, "backup": backup}}
	}
	switch a := primary.(type) {
	case M:
		b := asMap(backup)
		names := map[string]bool{}
		for k := range a {
			names[k] = true
		}
		for k := range b {
			names[k] = true
		}
		diff := []any{}
		for _, key := range sortedKeys(names) {
			where := key
			if path != "" {
				where = path + "." + key
			}
			left, lp := a[key]
			right, rp := b[key]
			if !lp || !rp {
				diff = append(diff, M{"path": where, "kind": "presence", "primary_present": lp, "backup_present": rp})
			} else {
				diff = append(diff, compareMetadata(left, right, where)...)
			}
		}
		return diff
	case []any:
		b := asList(backup)
		diff := []any{}
		if len(a) != len(b) {
			diff = append(diff, M{"path": path, "kind": "length", "primary": len(a), "backup": len(b)})
		}
		for i := 0; i < min(len(a), len(b)); i++ {
			diff = append(diff, compareMetadata(a[i], b[i], fmt.Sprintf("%s[%d]", path, i+1))...)
		}
		return diff
	}
	if evEqual(primary, backup) {
		return []any{}
	}
	return []any{M{"path": path, "kind": "value", "primary": primary, "backup": backup}}
}
func metadataKind(v any) string {
	switch v.(type) {
	case nil:
		return "nil"
	case bool:
		return "bool"
	case int, int64:
		return "int"
	case float64:
		return "float"
	case string:
		return "string"
	case M:
		return "map"
	case []any:
		return "array"
	}
	return fmt.Sprintf("%T", v)
}
func (p Project) loadBackup() (M, map[string]M, error) {
	lock, e := p.Lock("native_backup_sources.json")
	if e != nil {
		return nil, nil, e
	}
	if asInt(lock["schema_version"]) != 1 {
		return nil, nil, fmt.Errorf("unsupported native backup schema")
	}
	primary, e := p.Lock("sources.json")
	if e != nil {
		return nil, nil, e
	}
	expected := M{}
	for _, v := range asList(asMap(primary["blizzard"])["files"]) {
		row := asMap(v)
		expected[asString(row["name"])] = M{"sha": row["sha"], "size": row["size"]}
	}
	inventory := M{}
	for _, v := range asList(lock["inventory"]) {
		row := asMap(v)
		name := asString(row["name"])
		if inventory[name] != nil || !evidenceSafeRelative(name) {
			return nil, nil, fmt.Errorf("duplicate or unsafe native backup inventory: %s", name)
		}
		inventory[name] = M{"sha": row["sha"], "size": row["size"]}
	}
	if !evEqual(inventory, expected) {
		return nil, nil, fmt.Errorf("native backup complete metadata inventory differs from primary")
	}
	version, e := readEvidenceFile(p.InputPath("blizzard_version", "version.txt"), 1024)
	if e != nil {
		return nil, nil, e
	}
	versionRows := asList(asMap(primary["blizzard_version"])["files"])
	if len(versionRows) != 1 || blobSHA(version) != asMap(versionRows[0])["sha"] {
		return nil, nil, fmt.Errorf("native backup primary version integrity failure")
	}
	if lock["claimed_build"] != strings.TrimSpace(string(version)) {
		return nil, nil, fmt.Errorf("native backup build provenance differs from primary")
	}
	documents := map[string]M{}
	for _, v := range asList(lock["files"]) {
		row := asMap(v)
		name := asString(row["name"])
		exp := asMap(expected[name])
		if documents[name] != nil || exp == nil || row["sha"] != exp["sha"] || !evEqual(row["size"], exp["size"]) {
			return nil, nil, fmt.Errorf("invalid native backup file inventory: %s", name)
		}
		data, e := readEvidenceFile(filepath.Join(p.Root, "references/wow-api/native-backup", name), 1<<20)
		if e != nil {
			return nil, nil, e
		}
		if len(data) != asInt(row["size"]) || blobSHA(data) != row["sha"] || sha256hex(data) != row["sha256"] {
			return nil, nil, fmt.Errorf("native backup integrity failure: %s", name)
		}
		doc, e := ParseLua(data, name)
		if e != nil {
			return nil, nil, e
		}
		documents[name] = doc
	}
	return lock, documents, nil
}
func corroborateBackup(documents map[string]M, docs []nativeDocument) (M, error) {
	source := map[string]M{}
	for _, doc := range docs {
		source[doc.Name] = doc.Metadata
	}
	for name, backup := range documents {
		if source[name] == nil || len(compareMetadata(source[name], backup, "")) > 0 {
			return nil, fmt.Errorf("native backup metadata differs: %s", name)
		}
	}
	for _, name := range []string{"TimedSignalMapSharedDocumentation.lua", "SecretPredicatesDocumentation.lua", "SecretAspectConstantsDocumentation.lua"} {
		if documents[name] == nil {
			return nil, fmt.Errorf("native backup lacks predicate/callback context")
		}
	}
	collect := func(d map[string]M, field string) map[string]M {
		r := map[string]M{}
		for _, doc := range d {
			for _, v := range asList(doc[field]) {
				m := asMap(v)
				r[asString(m["Name"])] = m
			}
		}
		return r
	}
	predicates, backupPredicates := collect(source, "Predicates"), collect(documents, "Predicates")
	pending := []any{}
	for _, doc := range documents {
		for _, v := range asList(doc["Functions"]) {
			method := asMap(v)
			pending = append(pending, method)
			for key := range method {
				if strings.HasPrefix(key, "Requires") || strings.HasPrefix(key, "Check") {
					if predicates[key] == nil || backupPredicates[key] == nil || len(compareMetadata(predicates[key], backupPredicates[key], "")) > 0 {
						return nil, fmt.Errorf("native backup lacks matching referenced predicate: %s", key)
					}
				}
			}
		}
	}
	definitions, backupDefinitions := collect(source, "Tables"), collect(documents, "Tables")
	seen := map[string]bool{}
	for len(pending) > 0 {
		value := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		switch v := value.(type) {
		case []any:
			pending = append(pending, v...)
		case M:
			for _, item := range v {
				switch item.(type) {
				case []any, M:
					pending = append(pending, item)
				}
			}
			for _, key := range []string{"Type", "InnerType"} {
				name, ok := v[key].(string)
				if !ok || definitions[name] == nil || seen[name] {
					continue
				}
				seen[name] = true
				if len(seen) > 2048 {
					return nil, fmt.Errorf("native backup dependency graph exceeds type bound")
				}
				definition := definitions[name]
				pending = append(pending, definition)
				if definition["Type"] == "CallbackType" && (backupDefinitions[name] == nil || len(compareMetadata(definition, backupDefinitions[name], "")) > 0) {
					return nil, fmt.Errorf("native backup lacks matching referenced callback: %s", name)
				}
			}
		}
	}
	return M{"documents": evList(sortedKeys(documents)), "definition_comparison": "exact", "scope": "separately maintained metadata distribution; not independent native semantics"}, nil
}
