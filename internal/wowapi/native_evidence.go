package wowapi

import (
	"encoding/json"
	"fmt"
	"strings"
)

func primaryObservations(files map[string]string) (map[string]primaryObservation, error) {
	path := "blizzard/Blizzard_CustomAuraButton.lua"
	text := files[path]
	var required []string
	e := json.Unmarshal([]byte(nativeCallsiteRequiredJSON), &required)
	if e != nil {
		return nil, e
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			return nil, fmt.Errorf("primary call-site ownership or data-flow evidence changed")
		}
	}
	out := map[string]primaryObservation{}
	for _, row := range nativeCallsites {
		i := strings.Index(text, row.Snippet)
		if i < 0 {
			return nil, fmt.Errorf("primary call-site evidence changed: %s", row.Name)
		}
		out[row.Name] = primaryObservation{Source: path, Line: strings.Count(text[:i], "\n") + 1, Call: row.Snippet, ObservedArgumentCount: row.ArgumentCount, ArgumentRoles: row.ArgumentRoles, Scope: "Static Blizzard use corroborates this invocation and data flow; complete arity, return, secret permission and runtime behavior are not established."}
	}
	return out, nil
}

type capabilityReceiver struct {
	Methods []string
	Bases   []string
}

func inheritedCapabilityNames(owner string, definitions map[string]capabilityReceiver, seen []string) (map[string]bool, error) {
	if evContains(seen, owner) || len(seen) > 32 {
		return nil, fmt.Errorf("cyclic or unbounded native capability composition")
	}
	definition, present := definitions[owner]
	if !present {
		return nil, fmt.Errorf("missing native capability receiver: %s", owner)
	}
	names := map[string]bool{}
	for _, key := range definition.Methods {
		names[key] = true
	}
	for _, parent := range definition.Bases {
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
func (p Project) loadNativeEvidence(c *catalogue, projection declarationProjection) (NativeEvidence, error) {
	observed, e := p.LoadObservations()
	if e != nil {
		return NativeEvidence{}, e
	}
	lock, files, e := p.loadNativeSecondary()
	if e != nil {
		return NativeEvidence{}, e
	}
	docs := c.nativeDocs
	available := map[string]bool{}
	for _, doc := range docs {
		available[doc.Metadata.Name] = true
	}
	all := true
	for family := range nativeFamilies {
		all = all && available[family]
	}
	if !all {
		docs, e = p.nativeDocuments()
		if e != nil {
			return NativeEvidence{}, e
		}
	}
	backupLock, backupDocs, e := p.loadBackup()
	if e != nil {
		return NativeEvidence{}, e
	}
	backupContext, e := corroborateBackup(backupDocs, docs)
	if e != nil {
		return NativeEvidence{}, e
	}
	tables := map[string]string{}
	nativeDocs := map[string]NativeDocument{}
	for _, doc := range docs {
		for _, table := range doc.Metadata.Tables {
			tables[table.Name] = table.Type
		}
		if _, ok := nativeFamilies[doc.Metadata.Name]; ok {
			nativeDocs[doc.Metadata.Name] = doc
		}
	}
	prefix := "wowless/data/products/wow/"
	secondary := map[string]map[string]SecondaryReceiver{}
	secondaryCapabilities := map[string]map[string]capabilityReceiver{}
	for _, name := range []string{"luaobjects.yaml", "uiobjects.yaml"} {
		receivers, e := parseSecondaryReceivers([]byte(files[prefix+name]))
		if e != nil {
			return NativeEvidence{}, e
		}
		secondary[name] = receivers
		capabilities := map[string]capabilityReceiver{}
		for name, receiver := range receivers {
			capabilities[name] = capabilityReceiver{Methods: sortedKeys(receiver.Methods), Bases: receiver.Inherits}
		}
		secondaryCapabilities[name] = capabilities
	}
	neededTypes := map[string]bool{}
	for _, doc := range nativeDocs {
		for _, fn := range doc.Metadata.Functions {
			for _, tuple := range [][]NativeParameter{fn.Arguments, fn.Returns} {
				for _, parameter := range tuple {
					neededTypes[parameter.Type] = true
					neededTypes[parameter.InnerType] = true
				}
			}
		}
	}
	typedefs, e := parseSecondaryTypedefs([]byte(files[prefix+"docs.yaml"]), neededTypes)
	if e != nil {
		return NativeEvidence{}, e
	}
	workspaceCapabilities := map[string]capabilityReceiver{}
	for name, def := range c.defs {
		workspaceCapabilities[name] = capabilityReceiver{Methods: sortedKeys(def.Methods), Bases: def.Bases}
	}
	observations, e := primaryObservations(files)
	if e != nil {
		return NativeEvidence{}, e
	}
	methods, hierarchy := map[string]*NativeMethodEvidence{}, map[string]receiverHierarchy{}
	for _, family := range sortedKeys(nativeFamilies) {
		spec := nativeFamilies[family]
		doc := nativeDocs[family]
		backupDoc, backed := backupDocs[doc.Name]
		if doc.Name == "" || !backed {
			return NativeEvidence{}, fmt.Errorf("native backup lacks receiver metadata: %s", family)
		}
		nativeMethods, backupMethods := map[string]NativeFunction{}, map[string]NativeFunction{}
		for _, method := range doc.Metadata.Functions {
			nativeMethods[method.Name] = method
		}
		for _, method := range backupDoc.Metadata.Functions {
			backupMethods[method.Name] = method
		}
		if len(backupMethods) != len(backupDoc.Metadata.Functions) || !evEqual(sortedKeys(backupMethods), sortedKeys(nativeMethods)) {
			return NativeEvidence{}, fmt.Errorf("native backup method inventory differs: %s", family)
		}
		if len(nativeMethods) != len(doc.Metadata.Functions) || len(nativeMethods) != spec.count {
			return NativeEvidence{}, fmt.Errorf("native secondary method inventory changed: %s", family)
		}
		receiver := secondary[spec.source][spec.owner]
		receiverMethods := receiver.Methods
		if !evEqual(sortedKeys(receiverMethods), sortedKeys(nativeMethods)) {
			return NativeEvidence{}, fmt.Errorf("secondary method inventory differs: %s", family)
		}
		bases := definitionBases(c.defs[spec.canonical])
		parentage, e := compareHierarchy(bases, receiver)
		if e != nil {
			return NativeEvidence{}, e
		}

		if (family == "SimpleAnimRadialProgressAPI" || family == "SimpleVectorGraphicsAPI") && observed.Applicable {
			workspaceNames, secondaryNames := map[string]bool{}, map[string]bool{}
			for _, base := range bases {
				names, e := inheritedCapabilityNames(base, workspaceCapabilities, nil)
				if e != nil {
					return NativeEvidence{}, e
				}
				for key := range names {
					workspaceNames[key] = true
				}
			}
			for _, base := range parentage.SecondaryBases {
				names, e := inheritedCapabilityNames(base, secondaryCapabilities[spec.source], nil)
				if e != nil {
					return NativeEvidence{}, e
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
					if observed.Hierarchy[spec.canonical+".parent."+parent+"."+name].State != "present" {
						present = false
					}
				}
			}
			match := evEqual(sortedKeys(workspaceNames), sortedKeys(secondaryNames))
			parentage.NativeCapabilityConfirmation = &capabilityConfirmation{Build: observed.Sources.Build, WorkspaceMethods: sortedKeys(workspaceNames), SecondaryMethods: sortedKeys(secondaryNames), MethodSetsMatch: match, AllMethodsObservedPresent: present, Scope: "method availability on the owned native receiver; virtual names are not native type identity claims"}
			if present && match {
				for i := range parentage.Differences {
					parentage.Differences[i].Disposition = "reviewed composition naming difference; equivalent method sets corroborated by build-matched native lookup"
				}
			}
		}
		hierarchy[family] = parentage
		for _, name := range sortedKeys(nativeMethods) {
			native := nativeMethods[name]
			symbol := spec.canonical + "." + name
			compared, e := compareSecondaryMethod(native, receiverMethods[name], typedefs, tables)
			if e != nil {
				return NativeEvidence{}, e
			}
			var signatures []string
			if definition := c.defs[spec.canonical]; definition != nil && definition.Methods[name] != nil {
				signatures, e = projection.signatures(symbol)
				if e != nil {
					return NativeEvidence{}, e
				}
			}
			var invocation *CaptureInvocation
			if observed.Applicable {
				if value, present := observed.Methods[symbol]; present {
					invocation = &value
				}
			}
			var observation *primaryObservation
			if value, present := observations[symbol]; present {
				observation = &value
			}
			row := &NativeMethodEvidence{Symbol: symbol, Family: family, Coverage: "Blizzard-derived secondary translation consistency", NativeSource: "blizzard:" + doc.Name, NativeMetadata: native.Raw, SecondarySource: prefix + spec.source, SecondaryReceiver: spec.owner, SecondaryMethod: receiverMethods[name], SignatureComparison: compared, ReceiverHierarchy: parentage, NativeSecurityNotRepresentedBySecondary: nativeSecurity(native), BackupMetadata: backupMethodEvidence{Source: "native-backup:" + doc.Name, MetadataComparison: "exact", Security: nativeSecurity(backupMethods[name]), Scope: "separately maintained metadata distribution; not independent native semantics"}, EmittedSignatures: signatures, PrimaryObservation: observation, NativeInvocationObservation: invocation, Limits: []string{"Secondary signatures derive from the same Blizzard documentation source family.", "Missing secondary secrecy does not establish public values, permitted arguments or preconditions.", "Static sources do not establish live runtime availability or access."}}
			if review, present := observed.BehaviorReviews[symbol]; present {
				row.ReviewedBehaviorObservation = &review
			}
			methods[symbol] = row

		}
	}
	if len(methods) != 73 {
		return NativeEvidence{}, fmt.Errorf("secondary or primary method coverage changed")
	}
	for symbol := range observations {
		if methods[symbol] == nil {
			return NativeEvidence{}, fmt.Errorf("secondary or primary method coverage changed")
		}
	}
	statistics := map[string]int{"secondary_methods": len(methods), "signature_differences": 0, "receiver_families_with_differences": 0, "methods_with_unrepresented_security": 0, "methods_with_backup_metadata": len(methods), "security_contracts_without_build_matched_corroboration": 0, "methods_with_native_invocations": 0, "methods_without_native_invocations": 0, "native_property_mismatches": 0, "methods_with_accepted_public_behavior_observations": 0, "methods_without_accepted_public_behavior_observations": 0, "receiver_families_without_native_capability_corroboration": 0, "primary_callsite_methods": len(observations), "methods_without_primary_callsites": len(methods) - len(observations)}
	inc := func(key string, yes bool) {
		if yes {
			statistics[key] = statistics[key] + 1
		}
	}
	for _, row := range methods {
		invocation := row.NativeInvocationObservation
		invoked := invocation != nil && invocation.Invoked
		accepted := invocation != nil && invocation.BehaviorVerified || row.ReviewedBehaviorObservation != nil && row.ReviewedBehaviorObservation.BehaviorVerified
		inc("signature_differences", len(row.SignatureComparison.Differences) > 0)
		inc("methods_with_unrepresented_security", !row.NativeSecurityNotRepresentedBySecondary.empty())
		inc("methods_with_native_invocations", invoked)
		inc("methods_without_native_invocations", !invoked)
		inc("native_property_mismatches", invocation != nil && invocation.State == "property_mismatch")
		inc("methods_with_accepted_public_behavior_observations", accepted)
		inc("methods_without_accepted_public_behavior_observations", !accepted)
	}
	for _, parentage := range hierarchy {
		inc("receiver_families_with_differences", len(parentage.Differences) > 0)
	}
	for _, family := range []string{"SimpleAnimRadialProgressAPI", "SimpleVectorGraphicsAPI"} {
		confirmation := hierarchy[family].NativeCapabilityConfirmation
		inc("receiver_families_without_native_capability_corroboration", confirmation == nil || !confirmation.MethodSetsMatch || !confirmation.AllMethodsObservedPresent)
	}
	sources := NativeEvidenceSources{Crosscheck: lock, Backup: backupLock, NativeObservations: observed.Sources, NativeBehaviorReviews: observed.BehaviorReviewSources}
	coverage, e := decodeEvidenceJSON[nativeCoverage]([]byte(nativeCoverageJSON))
	if e != nil {
		return NativeEvidence{}, e
	}
	return NativeEvidence{Sources: sources, BackupContext: backupContext, Methods: methods, ReceiverHierarchy: hierarchy, Statistics: statistics, Coverage: coverage}, nil
}
