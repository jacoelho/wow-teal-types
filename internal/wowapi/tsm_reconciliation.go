package wowapi

import (
	"fmt"
	"sort"
)

func (p Project) reconcile(report, ledger, closure M) (M, error) {
	policies, approvals := asMap(ledger["policies"]), asMap(ledger["contracts"])
	actual := M{}
	for _, v := range asList(report["contracts"]) {
		row := asMap(v)
		symbol := asString(row["symbol"])
		if actual[symbol] != nil {
			return nil, fmt.Errorf("duplicate native symbols in TSM reconciliation")
		}
		actual[symbol] = row
	}
	results := []any{}
	unresolved := 0
	for _, name := range sortedKeys(actual) {
		contract, approval := asMap(actual[name]), asMap(approvals[name])
		state := "reviewed"
		if approval == nil {
			state = "unreviewed-contract"
		} else if approval["fingerprint"] != contract["fingerprint"] {
			state = "changed-contract"
		} else {
			unknown := false
			for _, rule := range evStrings(approval["policies"]) {
				unknown = unknown || policies[rule] == nil
			}
			for _, v := range asList(approval["resolutions"]) {
				unknown = unknown || policies[asString(asMap(v)["policy"])] == nil
			}
			if unknown {
				state = "unknown-resolution-policy"
			} else {
				deltas := []string{}
				altered := false
				for _, v := range asList(approval["resolutions"]) {
					row := asMap(v)
					deltas = append(deltas, asString(row["difference_sha256"]))
					altered = altered || evDigest(row["facts"]) != row["difference_sha256"]
				}
				sort.Strings(deltas)
				if !evEqual(evList(deltas), contract["discrepancies"]) {
					state = "missing-position-resolution"
				} else if altered {
					state = "altered-resolution-facts"
				} else if truth(contract["requires_resolution"]) && !truth(approval["policies"]) {
					state = "missing-resolution"
				}
			}
		}
		if state != "reviewed" {
			unresolved++
		}
		results = append(results, M{"symbol": name, "state": state, "policies": evGet(approval, "policies", []any{})})
	}
	orphans := []any{}
	for _, name := range sortedKeys(approvals) {
		if actual[name] == nil {
			orphans = append(orphans, name)
		}
	}
	graphMatches := ledger["declaration_graph_sha256"] == report["declaration_graph_sha256"]
	sourceMatches := ledger["source_locks_sha256"] == report["source_locks_sha256"]
	var equivalence M
	var e error
	if !sourceMatches && closure != nil {
		equivalence, e = p.validateEquivalence(report, ledger, closure)
		if e != nil {
			return nil, e
		}
	}
	r := M{"decisions": results, "orphaned_resolutions": orphans, "declaration_graph_matches": graphMatches, "source_locks_match": sourceMatches, "unresolved_contracts": unresolved, "passed": unresolved == 0 && len(orphans) == 0 && graphMatches && (sourceMatches || equivalence != nil), "policies": policies}
	if equivalence != nil {
		r["source_equivalence"] = equivalence
	}
	return r, nil
}
