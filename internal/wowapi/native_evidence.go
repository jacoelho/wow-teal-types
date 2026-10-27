package wowapi

import (
	"encoding/json"
	"fmt"
	"strings"
)

func staticEvidenceValue(raw string) any {
	var v any
	if e := json.Unmarshal([]byte(raw), &v); e != nil {
		panic(e)
	}
	return v
}
func primaryObservations(files map[string]string) (M, error) {
	path := "blizzard/Blizzard_CustomAuraButton.lua"
	text := files[path]
	for _, snippet := range evStrings(staticEvidenceValue(nativeCallsiteRequiredJSON)) {
		if !strings.Contains(text, snippet) {
			return nil, fmt.Errorf("primary call-site ownership or data-flow evidence changed")
		}
	}
	r := M{}
	for _, v := range asList(staticEvidenceValue(nativeCallsiteRowsJSON)) {
		row := asList(v)
		name, snippet := asString(row[0]), asString(row[1])
		i := strings.Index(text, snippet)
		if i < 0 {
			return nil, fmt.Errorf("primary call-site evidence changed: %s", name)
		}
		r[name] = M{"source": path, "line": strings.Count(text[:i], "\n") + 1, "call": snippet, "observed_argument_count": asInt(row[2]), "argument_roles": row[3], "scope": "Static Blizzard use corroborates this invocation and data flow; complete arity, return, secret permission and runtime behavior are not established."}
	}
	return r, nil
}
func inheritedCapabilityNames(owner string, definitions M, seen []string) (map[string]bool, error) {
	if evContains(seen, owner) || len(seen) > 32 {
		return nil, fmt.Errorf("cyclic or unbounded native capability composition")
	}
	definition := asMap(definitions[owner])
	if definition == nil {
		return nil, fmt.Errorf("missing native capability receiver: %s", owner)
	}
	names := map[string]bool{}
	for key := range asMap(definition["methods"]) {
		names[key] = true
	}
	parents := evGet(definition, "bases", definition["inherits"])
	parentNames := evStrings(parents)
	if s, ok := parents.(string); ok {
		parentNames = []string{s}
	} else if m := asMap(parents); m != nil {
		parentNames = sortedKeys(m)
	}
	for _, parent := range parentNames {
		more, e := inheritedCapabilityNames(parent, definitions, append(append([]string{}, seen...), owner))
		if e != nil {
			return nil, e
		}
		for key := range more {
			names[key] = true
		}
	}
	if len(names) > 512 {
		return nil, fmt.Errorf("native capability composition exceeds limit")
	}
	return names, nil
}
func (p Project) loadNativeEvidence(c *catalogue, projection declarationProjection) (M, error) {
	observed, e := p.LoadObservations()
	if e != nil {
		return nil, e
	}
	lock, files, e := p.loadNativeSecondary()
	if e != nil {
		return nil, e
	}
	docs := c.nativeDocs
	available := map[string]bool{}
	for _, doc := range docs {
		available[asString(doc.Metadata["Name"])] = true
	}
	all := true
	for family := range nativeFamilies {
		all = all && available[family]
	}
	if !all {
		docs, e = p.nativeDocuments()
		if e != nil {
			return nil, e
		}
	}
	backupLock, backupDocs, e := p.loadBackup()
	if e != nil {
		return nil, e
	}
	backupContext, e := corroborateBackup(backupDocs, docs)
	if e != nil {
		return nil, e
	}
	tables := map[string]string{}
	nativeDocs := map[string]nativeDocument{}
	for _, doc := range docs {
		for _, v := range asList(doc.Metadata["Tables"]) {
			m := asMap(v)
			tables[asString(m["Name"])] = asString(m["Type"])
		}
		if _, ok := nativeFamilies[asString(doc.Metadata["Name"])]; ok {
			nativeDocs[asString(doc.Metadata["Name"])] = doc
		}
	}
	prefix := "wowless/data/products/wow/"
	secondary := map[string]M{}
	for _, name := range []string{"luaobjects.yaml", "uiobjects.yaml"} {
		v, e := safeYAML([]byte(files[prefix+name]))
		if e != nil {
			return nil, e
		}
		secondary[name] = asMap(v)
	}
	v, e := safeYAML([]byte(files[prefix+"docs.yaml"]))
	if e != nil {
		return nil, e
	}
	typedefs := asMap(asMap(v)["typedefs"])
	observations, e := primaryObservations(files)
	if e != nil {
		return nil, e
	}
	methods, hierarchy := M{}, M{}
	for _, family := range sortedKeys(nativeFamilies) {
		spec := nativeFamilies[family]
		doc := nativeDocs[family]
		if doc.Metadata == nil || backupDocs[doc.Name] == nil {
			return nil, fmt.Errorf("native backup lacks receiver metadata: %s", family)
		}
		nativeMethods, backupMethods := M{}, M{}
		for _, v := range asList(doc.Metadata["Functions"]) {
			m := asMap(v)
			nativeMethods[asString(m["Name"])] = m
		}
		for _, v := range asList(backupDocs[doc.Name]["Functions"]) {
			m := asMap(v)
			backupMethods[asString(m["Name"])] = m
		}
		if len(backupMethods) != len(asList(backupDocs[doc.Name]["Functions"])) || !evEqual(evList(sortedKeys(backupMethods)), evList(sortedKeys(nativeMethods))) {
			return nil, fmt.Errorf("native backup method inventory differs: %s", family)
		}
		if len(nativeMethods) != len(asList(doc.Metadata["Functions"])) || len(nativeMethods) != spec.count {
			return nil, fmt.Errorf("native secondary method inventory changed: %s", family)
		}
		receiver, e := evStrictMapping(secondary[spec.source][spec.owner], []string{"methods", "inherits", "fields", "virtual"}, "secondary receiver")
		if e != nil {
			return nil, e
		}
		receiverMethods := asMap(receiver["methods"])
		if !evEqual(evList(sortedKeys(receiverMethods)), evList(sortedKeys(nativeMethods))) {
			return nil, fmt.Errorf("secondary method inventory differs: %s", family)
		}
		bases := evStrings(asMap(c.defs[spec.canonical])["bases"])
		parentage, e := compareHierarchy(bases, receiver)
		if e != nil {
			return nil, e
		}
		hierarchy[family] = parentage
		if (family == "SimpleAnimRadialProgressAPI" || family == "SimpleVectorGraphicsAPI") && truth(observed["applicable"]) {
			workspaceNames, secondaryNames := map[string]bool{}, map[string]bool{}
			for _, base := range bases {
				names, e := inheritedCapabilityNames(base, c.defs, nil)
				if e != nil {
					return nil, e
				}
				for key := range names {
					workspaceNames[key] = true
				}
			}
			for _, base := range evStrings(parentage["secondary_bases"]) {
				names, e := inheritedCapabilityNames(base, secondary[spec.source], nil)
				if e != nil {
					return nil, e
				}
				for key := range names {
					secondaryNames[key] = true
				}
			}
			parent := "Region"
			if spec.canonical == "RadialProgress" {
				parent = "Animation"
			}
			present := true
			for _, names := range []map[string]bool{workspaceNames, secondaryNames} {
				for name := range names {
					if asMap(asMap(observed["hierarchy"])[spec.canonical+".parent."+parent+"."+name])["state"] != "present" {
						present = false
					}
				}
			}
			match := evEqual(evList(sortedKeys(workspaceNames)), evList(sortedKeys(secondaryNames)))
			parentage["native_capability_confirmation"] = M{"build": asMap(observed["sources"])["build"], "workspace_methods": evList(sortedKeys(workspaceNames)), "secondary_methods": evList(sortedKeys(secondaryNames)), "method_sets_match": match, "all_methods_observed_present": present, "scope": "method availability on the owned native receiver; virtual names are not native type identity claims"}
			if present && match {
				for _, v := range asList(parentage["differences"]) {
					asMap(v)["disposition"] = "reviewed composition naming difference; equivalent method sets corroborated by build-matched native lookup"
				}
			}
		}
		for _, name := range sortedKeys(nativeMethods) {
			native := asMap(nativeMethods[name])
			symbol := spec.canonical + "." + name
			compared, e := compareSecondaryMethod(native, asMap(receiverMethods[name]), typedefs, tables)
			if e != nil {
				return nil, e
			}
			item := asMap(asMap(asMap(c.defs[spec.canonical])["methods"])[name])
			var signatures any
			if item != nil {
				a, e := projection.signatures(symbol)
				if e != nil {
					return nil, e
				}
				signatures = evList(a)
			}
			var invocation any
			if truth(observed["applicable"]) {
				invocation = asMap(observed["methods"])[symbol]
			}
			row := M{"symbol": symbol, "family": family, "coverage": "Blizzard-derived secondary translation consistency", "native_source": "blizzard:" + doc.Name, "native_metadata": cloneValue(native), "secondary_source": prefix + spec.source, "secondary_receiver": spec.owner, "secondary_method": cloneValue(receiverMethods[name]), "signature_comparison": compared, "receiver_hierarchy": parentage, "native_security_not_represented_by_secondary": nativeSecurity(native), "backup_metadata": M{"source": "native-backup:" + doc.Name, "metadata_comparison": "exact", "security": nativeSecurity(asMap(backupMethods[name])), "scope": "separately maintained metadata distribution; not independent native semantics"}, "emitted_signatures": signatures, "primary_observation": observations[symbol], "native_invocation_observation": invocation, "limits": []any{"Secondary signatures derive from the same Blizzard documentation source family.", "Missing secondary secrecy does not establish public values, permitted arguments or preconditions.", "Static sources do not establish live runtime availability or access."}}
			if review := asMap(observed["behavior_reviews"])[symbol]; review != nil {
				row["reviewed_behavior_observation"] = review
			}
			methods[symbol] = row
		}
	}
	if len(methods) != 73 {
		return nil, fmt.Errorf("secondary or primary method coverage changed")
	}
	for symbol := range observations {
		if methods[symbol] == nil {
			return nil, fmt.Errorf("secondary or primary method coverage changed")
		}
	}
	statistics := M{"secondary_methods": len(methods), "signature_differences": 0, "receiver_families_with_differences": 0, "methods_with_unrepresented_security": 0, "methods_with_backup_metadata": len(methods), "security_contracts_without_build_matched_corroboration": 0, "methods_with_native_invocations": 0, "methods_without_native_invocations": 0, "native_property_mismatches": 0, "methods_with_accepted_public_behavior_observations": 0, "methods_without_accepted_public_behavior_observations": 0, "receiver_families_without_native_capability_corroboration": 0, "primary_callsite_methods": len(observations), "methods_without_primary_callsites": len(methods) - len(observations)}
	inc := func(key string, yes bool) {
		if yes {
			statistics[key] = asInt(statistics[key]) + 1
		}
	}
	for _, v := range methods {
		row := asMap(v)
		invocation := asMap(row["native_invocation_observation"])
		invoked := invocation != nil && truth(invocation["invoked"])
		accepted := truth(invocation["behaviorVerified"]) || truth(asMap(row["reviewed_behavior_observation"])["behavior_verified"])
		inc("signature_differences", truth(asMap(row["signature_comparison"])["differences"]))
		inc("methods_with_unrepresented_security", truth(row["native_security_not_represented_by_secondary"]))
		inc("methods_with_native_invocations", invoked)
		inc("methods_without_native_invocations", !invoked)
		inc("native_property_mismatches", invocation["state"] == "property_mismatch")
		inc("methods_with_accepted_public_behavior_observations", accepted)
		inc("methods_without_accepted_public_behavior_observations", !accepted)
	}
	for _, v := range hierarchy {
		inc("receiver_families_with_differences", truth(asMap(v)["differences"]))
	}
	for _, family := range []string{"SimpleAnimRadialProgressAPI", "SimpleVectorGraphicsAPI"} {
		m := asMap(asMap(hierarchy[family])["native_capability_confirmation"])
		inc("receiver_families_without_native_capability_corroboration", !truth(m["method_sets_match"]) || !truth(m["all_methods_observed_present"]))
	}
	sources := evMerge(lock, M{"backup": backupLock, "native_observations": observed["sources"], "native_behavior_reviews": observed["behavior_review_sources"]})
	return M{"sources": sources, "backup_context": backupContext, "methods": methods, "receiver_hierarchy": hierarchy, "statistics": statistics, "coverage": staticEvidenceValue(nativeCoverageJSON)}, nil
}
