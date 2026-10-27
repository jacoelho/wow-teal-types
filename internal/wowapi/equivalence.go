package wowapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"wow-teal-types/internal/sourcelock"
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

// PolicyHashes binds the maintained interpretation and compiler closure. Tests are
// not policy inputs; each production source file and its inventory is bound.
func PolicyHashes(root string) (map[string]string, error) {
	names := []string{"go.mod", "go.sum", "scripts/wow_api/overrides.json", "scripts/wow_api/metadata_fields.json", "tools/check_secrets.lua", "tools/check_runtime_imports.lua", "types/wow/native_values.d.tl", "tlconfig.lua", "tools/teal/tl", "tools/teal/cli.lua", "tools/teal/tl.lua", "tools/teal/argparse.lua", "tools/teal/contract.lua", "tools/teal/check_capabilities.lua"}
	for _, directory := range []string{"internal/wowapi", "internal/release", "internal/consumer", "internal/sourcelock", "tools/wow-types", "tools/wow-teal"} {
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
	r := map[string]string{}
	for _, name := range names {
		b, e := checkedEquivalenceBytes(root, name)
		if e != nil {
			return nil, e
		}
		r[name] = sha256hex(b)
	}
	return r, nil
}
func equivalenceRows(source sourcelock.SourceGroup, group string) (map[string]sourcelock.SourceFile, error) {
	rows := map[string]sourcelock.SourceFile{}
	for _, row := range source.Files {
		if _, exists := rows[row.Name]; exists {
			return nil, fmt.Errorf("duplicate equivalence source input: %s:%s", group, row.Name)
		}
		rows[row.Name] = row
	}
	return rows, nil
}
func (p Project) equivalenceSourceBytes(group, name string, row sourcelock.SourceFile) ([]byte, error) {
	input, ok := p.inputs[group]
	if !ok {
		return nil, fmt.Errorf("unrecognized equivalence input group: %s", group)
	}
	data, e := checkedEquivalenceBytes(input, name)
	if e != nil {
		return nil, e
	}
	if len(data) != row.Size || blobSHA(data) != row.SHA {
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
func (p Project) equivalenceVersion(sources sourcelock.Sources) (string, error) {
	rows, e := equivalenceRows(sources["blizzard_version"], "blizzard_version")
	if e != nil {
		return "", e
	}
	if !exactEvidenceKeys(rows, []string{"version.txt"}) {
		return "", fmt.Errorf("equivalence requires one literal version.txt input")
	}
	data, e := p.equivalenceSourceBytes("blizzard_version", "version.txt", rows["version.txt"])
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
func (p Project) proveEquivalentInputs(baseline, target sourcelock.Sources, baselineVersion, targetVersion string) (int, error) {
	if !slices.Equal(sortedKeys(baseline), sortedKeys(target)) {
		return 0, fmt.Errorf("source-equivalence group inventory changed")
	}
	for _, group := range []string{"blizzard", "blizzard_support", "blizzard_version"} {
		if _, exists := target[group]; !exists {
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
		native := slices.Contains([]string{"blizzard", "blizzard_support", "blizzard_version"}, group)
		before, after := baseline[group], target[group]
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
		if !slices.Equal(sortedKeys(old), sortedKeys(newRows)) {
			return 0, fmt.Errorf("source-equivalence complete input inventory changed: %s", group)
		}
		if group == "blizzard_version" {
			if _, e := validEvidenceVersion(baselineVersion); e != nil {
				return 0, e
			}
			expected := []byte(baselineVersion + "\n")
			row := old["version.txt"]
			if row.SHA != blobSHA(expected) || row.Size != len(expected) {
				return 0, fmt.Errorf("source-equivalence baseline version literal changed")
			}
			continue
		}
		for _, name := range sortedKeys(newRows) {
			data, e := p.equivalenceSourceBytes(group, name, newRows[name])
			if e != nil {
				return 0, e
			}
			if native {
				row := old[name]
				matches := len(data) == row.Size && blobSHA(data) == row.SHA
				base := strings.TrimRight(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
				for count := 0; count < 16 && !matches; count++ {
					lf := base + strings.Repeat("\n", count)
					for _, value := range []string{lf, strings.ReplaceAll(lf, "\n", "\r\n")} {
						matches = matches || len(value) == row.Size && blobSHA([]byte(value)) == row.SHA
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
func (p Project) equivalenceProof(receipt EquivalenceReceipt, report CrosscheckReport, ledger ReconciliationLedger, closure sourceClosure, selectedVersion string) (*EquivalenceProof, error) {
	if receipt.SchemaVersion != 1 || receipt.Policy != equivalencePolicy {
		return nil, fmt.Errorf("unsupported source-equivalence receipt policy")
	}
	policies, e := PolicyHashes(p.Root)
	if e != nil {
		return nil, e
	}
	if !maps.Equal(receipt.PolicySHA256, policies) {
		return nil, fmt.Errorf("source-equivalence conversion or secret-check policy changed")
	}
	ledgerData, e := checkedEquivalenceBytes(p.Root, "scripts/wow_api/tsm_reconciliation.json")
	if e != nil {
		return nil, e
	}
	recorded, e := decodeEvidenceJSON[ReconciliationLedger](ledgerData)
	if e != nil || !evEqual(recorded, ledger) {
		return nil, fmt.Errorf("source-equivalence ledger argument differs from recorded ledger")
	}
	if receipt.LedgerSHA256 != sha256hex(ledgerData) {
		return nil, fmt.Errorf("source-equivalence original ledger changed")
	}
	selected := selectedVersion
	if selected == "" {
		selected, e = selectedEvidenceVersion(p.Root)
		if e != nil {
			return nil, e
		}
	}
	if receipt.TargetVersion != selected {
		return nil, fmt.Errorf("source-equivalence target version differs from .wow-version")
	}
	target := closure.Native
	recordedSources, e := sourcelock.ReadSources(filepath.Join(p.Root, "scripts/wow_api/sources.json"))
	if e != nil {
		return nil, e
	}
	if !evEqual(target, recordedSources) {
		return nil, fmt.Errorf("source-equivalence target sources differ from selected source lock")
	}
	if receipt.TargetSourcesSHA256 != evDigest(target) || report.SourceLocksSHA256 != evDigest(closure) {
		return nil, fmt.Errorf("source-equivalence target source identity or closure changed")
	}
	baseline := receipt.BaselineSources
	baselineClosure := closure
	baselineClosure.Native = baseline
	expected := evDigest(baselineClosure)
	if expected != ledger.SourceLocksSHA256 || receipt.BaselineSourceLocksSHA256 != expected {
		return nil, fmt.Errorf("source-equivalence TSM, secondary or resolved evidence changed")
	}
	checked, e := p.proveEquivalentInputs(baseline, target, receipt.BaselineVersion, selected)
	if e != nil {
		return nil, e
	}
	review, e := p.reconcile(report, ledger, nil)
	if e != nil {
		return nil, e
	}
	if !review.DeclarationGraphMatches || len(review.OrphanedResolutions) > 0 || review.UnresolvedContracts > 0 {
		return nil, fmt.Errorf("source equivalence cannot approve changed contracts, resolutions or declaration graph")
	}
	return &EquivalenceProof{Policy: equivalencePolicy, BaselineVersion: receipt.BaselineVersion, TargetVersion: selected, CheckedNativeFiles: checked, ReceiptSHA256: evDigest(receipt), LedgerSHA256: receipt.LedgerSHA256}, nil
}
func (p Project) validateEquivalence(report CrosscheckReport, ledger ReconciliationLedger, closure sourceClosure) (*EquivalenceProof, error) {
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
	receipt, e := decodeEvidenceJSON[EquivalenceReceipt](b)
	if e != nil {
		return nil, e
	}
	return p.equivalenceProof(receipt, report, ledger, closure, "")
}

// CreateEquivalence proposes source-only evidence; generation revalidates every
// bound fact before it can affect publication.
func CreateEquivalence(baselineRoot, targetRoot string) (EquivalenceReceipt, error) {
	baseline, target := NewProject(baselineRoot), NewProject(targetRoot)
	old, e := sourcelock.ReadSources(filepath.Join(baselineRoot, "scripts/wow_api/sources.json"))
	if e != nil {
		return EquivalenceReceipt{}, e
	}
	current, e := sourcelock.ReadSources(filepath.Join(targetRoot, "scripts/wow_api/sources.json"))
	if e != nil {
		return EquivalenceReceipt{}, e
	}
	ledger, e := readEvidenceJSON[ReconciliationLedger](filepath.Join(baselineRoot, "scripts/wow_api/tsm_reconciliation.json"))
	if e != nil {
		return EquivalenceReceipt{}, e
	}
	oldLedger, e := checkedEquivalenceBytes(baselineRoot, "scripts/wow_api/tsm_reconciliation.json")
	if e != nil {
		return EquivalenceReceipt{}, e
	}
	newLedger, e := checkedEquivalenceBytes(targetRoot, "scripts/wow_api/tsm_reconciliation.json")
	if e != nil {
		return EquivalenceReceipt{}, e
	}
	if string(oldLedger) != string(newLedger) {
		return EquivalenceReceipt{}, fmt.Errorf("source-equivalence ledger bytes differ between workspaces")
	}
	policies, e := PolicyHashes(baselineRoot)
	if e != nil {
		return EquivalenceReceipt{}, e
	}
	newPolicies, e := PolicyHashes(targetRoot)
	if e != nil {
		return EquivalenceReceipt{}, e
	}
	if !evEqual(policies, newPolicies) {
		return EquivalenceReceipt{}, fmt.Errorf("source-equivalence conversion or secret-check policy changed")
	}
	baselineVersion, e := baseline.equivalenceVersion(old)
	if e != nil {
		return EquivalenceReceipt{}, e
	}
	report, e := readReconciliationEvidence(filepath.Join(baselineRoot, "evidence/generated/wow_tsm_crosscheck.json"))
	if e != nil {
		return EquivalenceReceipt{}, e
	}
	tsm, e := sourcelock.ReadTSM(filepath.Join(baselineRoot, "scripts/wow_api/tsm_sources.json"))
	if e != nil {
		return EquivalenceReceipt{}, e
	}
	closure := sourceClosure{Native: old, TSM: tsm, Secondary: report.NativeSecondaryEvidence.Sources, Resolved: report.TSMResolvedSources}
	review, e := baseline.reconcile(report, ledger, nil)
	if e != nil {
		return EquivalenceReceipt{}, e
	}
	if !review.DeclarationGraphMatches || len(review.OrphanedResolutions) > 0 || review.UnresolvedContracts > 0 {
		return EquivalenceReceipt{}, fmt.Errorf("source equivalence cannot approve changed contracts, resolutions or declaration graph")
	}
	if _, e := os.Lstat(filepath.Join(baselineRoot, equivalenceReceipt)); e == nil {
		prior, e := readEvidenceJSON[EquivalenceReceipt](filepath.Join(baselineRoot, equivalenceReceipt))
		if e != nil {
			return EquivalenceReceipt{}, e
		}
		if _, e = baseline.equivalenceProof(prior, report, ledger, closure, baselineVersion); e != nil {
			return EquivalenceReceipt{}, e
		}
		old = prior.BaselineSources
		baselineVersion = prior.BaselineVersion
	} else if !os.IsNotExist(e) {
		return EquivalenceReceipt{}, e
	}
	closure.Native = old
	if evDigest(closure) != ledger.SourceLocksSHA256 {
		return EquivalenceReceipt{}, fmt.Errorf("source-equivalence baseline source closure was not reviewed by original ledger")
	}
	targetVersion, e := selectedEvidenceVersion(targetRoot)
	if e != nil {
		return EquivalenceReceipt{}, e
	}
	if _, e = target.proveEquivalentInputs(old, current, baselineVersion, targetVersion); e != nil {
		return EquivalenceReceipt{}, e
	}
	return EquivalenceReceipt{SchemaVersion: 1, Policy: equivalencePolicy, PolicySHA256: policies, LedgerSHA256: sha256hex(oldLedger), BaselineSources: old, BaselineVersion: baselineVersion, BaselineSourceLocksSHA256: ledger.SourceLocksSHA256, TargetSourcesSHA256: evDigest(current), TargetVersion: targetVersion}, nil
}

func readEvidenceJSON[T any](path string) (result T, err error) {
	data, err := readEvidenceFile(path, 128<<20)
	if err != nil {
		return result, err
	}
	return decodeEvidenceJSON[T](data)
}

// Source equivalence consumes reviewed contract identities and source closure.
// Diagnostic details remain immutable evidence; decoding them into output-only
// comparison variants would create a second interpretation of that evidence.
func readReconciliationEvidence(path string) (CrosscheckReport, error) {
	data, err := readEvidenceFile(path, 128<<20)
	if err != nil {
		return CrosscheckReport{}, err
	}
	if trimmed := bytes.TrimSpace(data); len(trimmed) == 0 || trimmed[0] != '{' {
		return CrosscheckReport{}, fmt.Errorf("reconciliation evidence must be an object")
	}
	var recorded struct {
		Contracts               []reconciledContract `json:"contracts"`
		DeclarationGraphSHA256  string               `json:"declaration_graph_sha256"`
		NativeSecondaryEvidence struct {
			Sources NativeEvidenceSources `json:"sources"`
		} `json:"native_secondary_evidence"`
		SourceLocksSHA256  string          `json:"source_locks_sha256"`
		TSMResolvedSources resolvedSources `json:"tsm_resolved_sources"`
	}
	// This is an intentional read projection. Unknown diagnostic fields belong to
	// the report producer; encoding/json still rejects wrong consumed field types
	// and trailing documents.
	if err = json.Unmarshal(data, &recorded); err != nil {
		return CrosscheckReport{}, fmt.Errorf("read reconciliation identities: %w", err)
	}
	if recorded.Contracts == nil || recorded.DeclarationGraphSHA256 == "" || recorded.SourceLocksSHA256 == "" {
		return CrosscheckReport{}, fmt.Errorf("incomplete reconciliation identities")
	}
	for _, contract := range recorded.Contracts {
		if contract.Symbol == "" || contract.Fingerprint == "" || contract.Discrepancies == nil {
			return CrosscheckReport{}, fmt.Errorf("incomplete reconciliation contract")
		}
	}
	return CrosscheckReport{Contracts: recorded.Contracts, DeclarationGraphSHA256: recorded.DeclarationGraphSHA256, SourceLocksSHA256: recorded.SourceLocksSHA256, NativeSecondaryEvidence: NativeEvidence{Sources: recorded.NativeSecondaryEvidence.Sources}, TSMResolvedSources: recorded.TSMResolvedSources}, nil
}
