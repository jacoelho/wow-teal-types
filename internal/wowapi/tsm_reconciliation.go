package wowapi

import (
	"fmt"
	"sort"
)

func (p Project) reconcile(report CrosscheckReport, ledger ReconciliationLedger, closure *sourceClosure) (ReconciliationReport, error) {
	actual := map[string]reconciledContract{}
	for _, row := range report.Contracts {
		if _, ok := actual[row.Symbol]; ok {
			return ReconciliationReport{}, fmt.Errorf("duplicate native symbols in TSM reconciliation")
		}
		actual[row.Symbol] = row
	}
	results := []reconciliationDecision{}
	unresolved := 0
	for _, name := range sortedKeys(actual) {
		contract := actual[name]
		approval, ok := ledger.Contracts[name]
		state := "reviewed"
		if !ok {
			state = "unreviewed-contract"
		} else if approval.Fingerprint != contract.Fingerprint {
			state = "changed-contract"
		} else {
			unknown := false
			for _, rule := range approval.Policies {
				_, present := ledger.Policies[rule]
				unknown = unknown || !present
			}
			for _, row := range approval.Resolutions {
				_, present := ledger.Policies[row.Policy]
				unknown = unknown || !present
			}
			if unknown {
				state = "unknown-resolution-policy"
			} else {
				deltas := []string{}
				altered := false
				for _, row := range approval.Resolutions {
					deltas = append(deltas, row.DifferenceSHA256)
					altered = altered || evDigest(row.Facts) != row.DifferenceSHA256
				}
				sort.Strings(deltas)
				if !evEqual(deltas, contract.Discrepancies) {
					state = "missing-position-resolution"
				} else if altered {
					state = "altered-resolution-facts"
				} else if contract.RequiresResolution && len(approval.Policies) == 0 {
					state = "missing-resolution"
				}
			}
		}
		if state != "reviewed" {
			unresolved++
		}
		policies := approval.Policies
		if !ok {
			policies = []string{}
		}
		results = append(results, reconciliationDecision{Symbol: name, State: state, Policies: policies})
	}
	orphans := []string{}
	for _, name := range sortedKeys(ledger.Contracts) {
		if _, ok := actual[name]; !ok {
			orphans = append(orphans, name)
		}
	}
	graphMatches := ledger.DeclarationGraphSHA256 == report.DeclarationGraphSHA256
	sourceMatches := ledger.SourceLocksSHA256 == report.SourceLocksSHA256
	var equivalence *EquivalenceProof
	var err error
	if !sourceMatches && closure != nil {
		equivalence, err = p.validateEquivalence(report, ledger, *closure)
		if err != nil {
			return ReconciliationReport{}, err
		}
	}
	return ReconciliationReport{Decisions: results, OrphanedResolutions: orphans, DeclarationGraphMatches: graphMatches, SourceLocksMatch: sourceMatches, UnresolvedContracts: unresolved, Passed: unresolved == 0 && len(orphans) == 0 && graphMatches && (sourceMatches || equivalence != nil), Policies: ledger.Policies, SourceEquivalence: equivalence}, nil
}
