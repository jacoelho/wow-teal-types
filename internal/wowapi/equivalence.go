package wowapi

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const equivalenceReceipt = "scripts/wow_api/source_equivalence.json"
const equivalencePolicy = "native-source-byte-equivalence-v1"

func checkedEquivalenceBytes(root, name string) ([]byte, error) {
	if !evidenceSafeRelative(name) {
		return nil, fmt.Errorf("unsafe equivalence input path: %s", name)
	}
	path := filepath.Join(root, filepath.FromSlash(name))
	for parent := path; parent != filepath.Dir(root); parent = filepath.Dir(parent) {
		if info, e := os.Lstat(parent); e != nil {
			return nil, e
		} else if info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("equivalence inputs cannot contain symlinks: %s", path)
		}
		if parent == root {
			break
		}
	}
	return readEvidenceFile(path, 8<<20)
}

// PolicyHashes binds the complete maintained Go interpretation closure. Tests are
// not policy inputs; each production source file and its inventory is bound.
func PolicyHashes(root string) (M, error) {
	names := []string{"go.mod", "go.sum", "scripts/wow_api/overrides.json", "scripts/wow_api/metadata_fields.json", "tools/check_secrets.lua", "tools/check_runtime_imports.lua", "types/wow/native_values.d.tl", "tlconfig.lua"}
	for _, directory := range []string{"internal/wowapi", "internal/release", "internal/consumer", "tools/wow-types", "tools/wow-teal"} {
		e := filepath.WalkDir(filepath.Join(root, directory), func(path string, d fs.DirEntry, e error) error {
			if e != nil {
				return e
			}
			if d.Type()&os.ModeSymlink != 0 {
				return fmt.Errorf("symlink in conversion policy: %s", path)
			}
			if !d.IsDir() && strings.HasSuffix(d.Name(), ".go") && !strings.HasSuffix(d.Name(), "_test.go") {
				name, e := filepath.Rel(root, path)
				if e != nil {
					return e
				}
				names = append(names, filepath.ToSlash(name))
			}
			return nil
		})
		if e != nil {
			return nil, e
		}
	}
	sort.Strings(names)
	r := M{}
	for _, name := range names {
		b, e := checkedEquivalenceBytes(root, name)
		if e != nil {
			return nil, e
		}
		r[name] = sha256hex(b)
	}
	return r, nil
}
func equivalenceRows(source M, group string) (M, error) {
	rows := M{}
	for _, v := range asList(source["files"]) {
		row := asMap(v)
		name := asString(row["name"])
		if rows[name] != nil {
			return nil, fmt.Errorf("duplicate equivalence source input: %s:%s", group, name)
		}
		rows[name] = row
	}
	return rows, nil
}
func (p Project) equivalenceSourceBytes(group, name string, row M) ([]byte, error) {
	input, ok := p.inputs[group]
	if !ok {
		return nil, fmt.Errorf("unrecognized equivalence input group: %s", group)
	}
	data, e := checkedEquivalenceBytes(input, name)
	if e != nil {
		return nil, e
	}
	if len(data) != asInt(row["size"]) || blobSHA(data) != row["sha"] {
		return nil, fmt.Errorf("modified equivalence source input: %s:%s", group, name)
	}
	return data, nil
}
func validEvidenceVersion(v string) (string, error) {
	if !regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+$`).MatchString(v) {
		return "", fmt.Errorf("invalid equivalence client version")
	}
	return v, nil
}
func (p Project) equivalenceVersion(sources M) (string, error) {
	rows, e := equivalenceRows(asMap(sources["blizzard_version"]), "blizzard_version")
	if e != nil {
		return "", e
	}
	if !exactEvidenceKeys(rows, []string{"version.txt"}) {
		return "", fmt.Errorf("equivalence requires one literal version.txt input")
	}
	data, e := p.equivalenceSourceBytes("blizzard_version", "version.txt", asMap(rows["version.txt"]))
	if e != nil {
		return "", e
	}
	v, e := validEvidenceVersion(strings.TrimSuffix(string(data), "\n"))
	if e != nil {
		return "", e
	}
	if string(data) != v+"\n" {
		return "", fmt.Errorf("equivalence requires the exact client version literal")
	}
	return v, nil
}
func selectedEvidenceVersion(root string) (string, error) {
	b, e := checkedEquivalenceBytes(root, ".wow-version")
	if e != nil {
		return "", e
	}
	if len(b) > 64 {
		return "", fmt.Errorf("invalid .wow-version")
	}
	return validEvidenceVersion(strings.TrimSuffix(string(b), "\n"))
}
func (p Project) proveEquivalentInputs(baseline, target M, baselineVersion, targetVersion string) (int, error) {
	if !evEqual(evList(sortedKeys(baseline)), evList(sortedKeys(target))) {
		return 0, fmt.Errorf("source-equivalence group inventory changed")
	}
	for _, group := range []string{"blizzard", "blizzard_support", "blizzard_version"} {
		if target[group] == nil {
			return 0, fmt.Errorf("source-equivalence group inventory changed")
		}
	}
	version, e := p.equivalenceVersion(target)
	if e != nil {
		return 0, e
	}
	if version != targetVersion {
		return 0, fmt.Errorf("source-equivalence target version differs from .wow-version")
	}
	checked := 0
	for _, group := range sortedKeys(target) {
		native := evContains([]string{"blizzard", "blizzard_support", "blizzard_version"}, group)
		before, after := asMap(baseline[group]), asMap(target[group])
		if !native && !evEqual(before, after) {
			return 0, fmt.Errorf("source-equivalence supplemental identity changed: %s", group)
		}
		old, e := equivalenceRows(before, group)
		if e != nil {
			return 0, e
		}
		newRows, e := equivalenceRows(after, group)
		if e != nil {
			return 0, e
		}
		if !evEqual(evList(sortedKeys(old)), evList(sortedKeys(newRows))) {
			return 0, fmt.Errorf("source-equivalence complete input inventory changed: %s", group)
		}
		if group == "blizzard_version" {
			if _, e := validEvidenceVersion(baselineVersion); e != nil {
				return 0, e
			}
			expected := []byte(baselineVersion + "\n")
			row := asMap(old["version.txt"])
			if row["sha"] != blobSHA(expected) || asInt(row["size"]) != len(expected) {
				return 0, fmt.Errorf("source-equivalence baseline version literal changed")
			}
			continue
		}
		for _, name := range sortedKeys(newRows) {
			data, e := p.equivalenceSourceBytes(group, name, asMap(newRows[name]))
			if e != nil {
				return 0, e
			}
			if native {
				row := asMap(old[name])
				matches := len(data) == asInt(row["size"]) && blobSHA(data) == row["sha"]
				base := strings.TrimRight(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
				for count := 0; count < 16 && !matches; count++ {
					lf := base + strings.Repeat("\n", count)
					for _, value := range []string{lf, strings.ReplaceAll(lf, "\n", "\r\n")} {
						matches = matches || len(value) == asInt(row["size"]) && blobSHA([]byte(value)) == row["sha"]
					}
				}
				if !matches {
					return 0, fmt.Errorf("source-equivalence native metadata or support bytes changed: %s:%s", group, name)
				}
				checked++
			}
		}
	}
	return checked, nil
}
func (p Project) equivalenceProof(receipt, report, ledger, closure M, selectedVersion string) (M, error) {
	if asInt(receipt["schema_version"]) != 1 || receipt["policy"] != equivalencePolicy {
		return nil, fmt.Errorf("unsupported source-equivalence receipt policy")
	}
	policies, e := PolicyHashes(p.Root)
	if e != nil {
		return nil, e
	}
	if !evEqual(receipt["policy_sha256"], policies) {
		return nil, fmt.Errorf("source-equivalence conversion or secret-check policy changed")
	}
	ledgerData, e := checkedEquivalenceBytes(p.Root, "scripts/wow_api/tsm_reconciliation.json")
	if e != nil {
		return nil, e
	}
	recorded, e := decodeEvidenceJSON(ledgerData)
	if e != nil || !evEqual(recorded, ledger) {
		return nil, fmt.Errorf("source-equivalence ledger argument differs from recorded ledger")
	}
	if receipt["ledger_sha256"] != sha256hex(ledgerData) {
		return nil, fmt.Errorf("source-equivalence original ledger changed")
	}
	selected := selectedVersion
	if selected == "" {
		selected, e = selectedEvidenceVersion(p.Root)
		if e != nil {
			return nil, e
		}
	}
	if receipt["target_version"] != selected {
		return nil, fmt.Errorf("source-equivalence target version differs from .wow-version")
	}
	target := asMap(closure["native"])
	recorded, e = p.Lock("sources.json")
	if e != nil {
		return nil, e
	}
	if !evEqual(target, recorded) {
		return nil, fmt.Errorf("source-equivalence target sources differ from selected source lock")
	}
	if receipt["target_sources_sha256"] != evDigest(target) || report["source_locks_sha256"] != evDigest(closure) {
		return nil, fmt.Errorf("source-equivalence target source identity or closure changed")
	}
	baseline := asMap(receipt["baseline_sources"])
	expected := evDigest(evMerge(closure, M{"native": baseline}))
	if expected != ledger["source_locks_sha256"] || receipt["baseline_source_locks_sha256"] != expected {
		return nil, fmt.Errorf("source-equivalence TSM, secondary or resolved evidence changed")
	}
	checked, e := p.proveEquivalentInputs(baseline, target, asString(receipt["baseline_version"]), selected)
	if e != nil {
		return nil, e
	}
	review, e := p.reconcile(report, ledger, nil)
	if e != nil {
		return nil, e
	}
	if !truth(review["declaration_graph_matches"]) || truth(review["orphaned_resolutions"]) || truth(review["unresolved_contracts"]) {
		return nil, fmt.Errorf("source equivalence cannot approve changed contracts, resolutions or declaration graph")
	}
	return M{"policy": equivalencePolicy, "baseline_version": receipt["baseline_version"], "target_version": selected, "checked_native_files": checked, "receipt_sha256": evDigest(receipt), "ledger_sha256": receipt["ledger_sha256"]}, nil
}
func (p Project) validateEquivalence(report, ledger, closure M) (M, error) {
	path := filepath.Join(p.Root, equivalenceReceipt)
	if _, e := os.Lstat(path); os.IsNotExist(e) {
		return nil, nil
	} else if e != nil {
		return nil, e
	}
	b, e := checkedEquivalenceBytes(p.Root, equivalenceReceipt)
	if e != nil {
		return nil, e
	}
	receipt, e := decodeEvidenceJSON(b)
	if e != nil {
		return nil, e
	}
	return p.equivalenceProof(receipt, report, ledger, closure, "")
}

// CreateEquivalence proposes source-only evidence; generation revalidates every
// bound fact before it can affect publication.
func CreateEquivalence(baselineRoot, targetRoot string) (M, error) {
	baseline, target := NewProject(baselineRoot), NewProject(targetRoot)
	old, e := baseline.Lock("sources.json")
	if e != nil {
		return nil, e
	}
	current, e := target.Lock("sources.json")
	if e != nil {
		return nil, e
	}
	ledger, e := baseline.Lock("tsm_reconciliation.json")
	if e != nil {
		return nil, e
	}
	oldLedger, e := checkedEquivalenceBytes(baselineRoot, "scripts/wow_api/tsm_reconciliation.json")
	if e != nil {
		return nil, e
	}
	newLedger, e := checkedEquivalenceBytes(targetRoot, "scripts/wow_api/tsm_reconciliation.json")
	if e != nil {
		return nil, e
	}
	if string(oldLedger) != string(newLedger) {
		return nil, fmt.Errorf("source-equivalence ledger bytes differ between workspaces")
	}
	policies, e := PolicyHashes(baselineRoot)
	if e != nil {
		return nil, e
	}
	newPolicies, e := PolicyHashes(targetRoot)
	if e != nil {
		return nil, e
	}
	if !evEqual(policies, newPolicies) {
		return nil, fmt.Errorf("source-equivalence conversion or secret-check policy changed")
	}
	baselineVersion, e := baseline.equivalenceVersion(old)
	if e != nil {
		return nil, e
	}
	report, e := readJSON(filepath.Join(baselineRoot, "evidence/generated/wow_tsm_crosscheck.json"))
	if e != nil {
		return nil, e
	}
	tsm, e := baseline.Lock("tsm_sources.json")
	if e != nil {
		return nil, e
	}
	closure := M{"native": old, "tsm": tsm, "secondary": asMap(report["native_secondary_evidence"])["sources"], "resolved": report["tsm_resolved_sources"]}
	review, e := baseline.reconcile(report, ledger, nil)
	if e != nil {
		return nil, e
	}
	if !truth(review["declaration_graph_matches"]) || truth(review["orphaned_resolutions"]) || truth(review["unresolved_contracts"]) {
		return nil, fmt.Errorf("source equivalence cannot approve changed contracts, resolutions or declaration graph")
	}
	if _, e := os.Lstat(filepath.Join(baselineRoot, equivalenceReceipt)); e == nil {
		prior, e := readJSON(filepath.Join(baselineRoot, equivalenceReceipt))
		if e != nil {
			return nil, e
		}
		if _, e = baseline.equivalenceProof(prior, report, ledger, closure, baselineVersion); e != nil {
			return nil, e
		}
		old = asMap(prior["baseline_sources"])
		baselineVersion = asString(prior["baseline_version"])
	} else if !os.IsNotExist(e) {
		return nil, e
	}
	if evDigest(evMerge(closure, M{"native": old})) != ledger["source_locks_sha256"] {
		return nil, fmt.Errorf("source-equivalence baseline source closure was not reviewed by original ledger")
	}
	targetVersion, e := selectedEvidenceVersion(targetRoot)
	if e != nil {
		return nil, e
	}
	if _, e = target.proveEquivalentInputs(old, current, baselineVersion, targetVersion); e != nil {
		return nil, e
	}
	return M{"schema_version": 1, "policy": equivalencePolicy, "policy_sha256": policies, "ledger_sha256": sha256hex(oldLedger), "baseline_sources": old, "baseline_version": baselineVersion, "baseline_source_locks_sha256": ledger["source_locks_sha256"], "target_sources_sha256": evDigest(current), "target_version": targetVersion}, nil
}
