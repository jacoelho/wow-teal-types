package wowapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"wow-teal-types/internal/sourcelock"
)

func readEvidenceBounded(r io.Reader, limit int64) ([]byte, error) {
	data, e := io.ReadAll(io.LimitReader(r, limit+1))
	if e != nil {
		return nil, e
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("evidence exceeds %d byte limit", limit)
	}
	return data, nil
}
func readEvidenceFile(path string, limit int64) ([]byte, error) {
	info, e := os.Lstat(path)
	if e != nil {
		return nil, e
	}
	if !info.Mode().IsRegular() || info.Size() > limit {
		return nil, fmt.Errorf("nonregular or oversized evidence input: %s", path)
	}
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	return readEvidenceBounded(f, limit)
}
func evidenceSafeRelative(name string) bool {
	if name == "" || filepath.IsAbs(name) || strings.ContainsAny(name, "\\\x00") {
		return false
	}
	for _, s := range strings.Split(name, "/") {
		if s == "" || s == "." || s == ".." {
			return false
		}
	}
	return true
}
func decodeEvidenceJSON[T any](data []byte) (T, error) {
	var value T
	if len(bytes.TrimSpace(data)) == 0 || bytes.TrimSpace(data)[0] != '{' {
		return value, fmt.Errorf("evidence requires JSON object")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return value, err
	}
	if err := decoder.Decode(new(json.RawMessage)); err != io.EOF {
		return value, fmt.Errorf("trailing evidence JSON data")
	}
	return value, nil
}

type boundedEvidenceLog struct {
	data  []byte
	limit int
}

func (b *boundedEvidenceLog) Write(p []byte) (int, error) {
	if remaining := b.limit - len(b.data); remaining > 0 {
		b.data = append(b.data, p[:min(len(p), remaining)]...)
	}
	return len(p), nil
}
func runEvidenceCommand(ctx context.Context, args []string, stdin []byte, limit int64) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	log := boundedEvidenceLog{limit: 4096}
	cmd.Stderr = &log
	pipe, e := cmd.StdoutPipe()
	if e != nil {
		return nil, e
	}
	if e = cmd.Start(); e != nil {
		return nil, e
	}
	data, readErr := readEvidenceBounded(pipe, limit)
	if readErr != nil {
		_ = cmd.Process.Kill()
	}
	e = cmd.Wait()
	if readErr != nil {
		return nil, readErr
	}
	if e != nil {
		return nil, fmt.Errorf("%s: %w: %s", args[0], e, string(log.data))
	}
	return data, nil
}
func (p Project) resolvedIdentity() (resolvedIdentity, error) {
	lock, err := sourcelock.ReadTSM(filepath.Join(p.Root, "scripts/wow_api/tsm_sources.json"))
	if err != nil {
		return resolvedIdentity{}, err
	}
	if lock.Repository != "TradeSkillMaster/wowlua-ls" || !regexp.MustCompile(`^[0-9a-f]{40}$`).MatchString(lock.Commit) || lock.BlobVersion != 44 {
		return resolvedIdentity{}, fmt.Errorf("invalid resolved TSM source identity or reader schema")
	}
	var blob *sourcelock.LockedFile
	for i := range lock.Files {
		if lock.Files[i].Name == "precomputed.bin.zst" {
			if blob != nil {
				return resolvedIdentity{}, fmt.Errorf("resolved TSM source lock requires one resolved blob")
			}
			blob = &lock.Files[i]
		}
	}
	if blob == nil || blob.Size < 8 || blob.Size > 64<<20 || !sourcelock.IsSHA256(blob.SHA256) {
		return resolvedIdentity{}, fmt.Errorf("invalid resolved TSM blob identity")
	}
	return resolvedIdentity{Repository: lock.Repository, Commit: lock.Commit, BlobVersion: lock.BlobVersion, BlobSHA256: blob.SHA256, BlobSize: blob.Size}, nil
}
func (p Project) exporterSources() (map[string]string, error) {
	r := map[string]string{}
	base := filepath.Join(p.Root, "scripts/wow_api/tsm_resolved_export")
	for _, name := range []string{"Cargo.toml", "Cargo.lock", "src/main.rs"} {
		b, e := readEvidenceFile(filepath.Join(base, name), 2<<20)
		if e != nil {
			return nil, e
		}
		r[name] = sha256hex(b)
	}
	data, e := readEvidenceFile(filepath.Join(p.Root, "scripts/wow_api/tsm_sources.json"), 2<<20)
	if e != nil {
		return nil, e
	}
	r["tsm_sources.json"] = sha256hex(data)
	return r, nil
}
func (p Project) resolvedExecutable(ctx context.Context) (string, error) {
	path := p.reader
	if path == "" {
		path = os.Getenv("TSM_RESOLVED_EXPORT")
	}
	if path == "" {
		data, e := runEvidenceCommand(ctx, []string{"cargo", "metadata", "--locked", "--offline", "--no-deps", "--format-version", "1", "--manifest-path", filepath.Join(p.Root, "scripts/wow_api/tsm_resolved_export/Cargo.toml")}, nil, 2<<20)
		if e != nil {
			return "", fmt.Errorf("cannot locate resolved TSM exporter; build pinned reader or set TSM_RESOLVED_EXPORT: %w", e)
		}
		var metadata struct {
			TargetDirectory string `json:"target_directory"`
		}
		e = json.Unmarshal(data, &metadata)
		if e != nil {
			return "", e
		}
		path = filepath.Join(metadata.TargetDirectory, "debug/tsm-resolved-export")
	}
	info, e := os.Stat(path)
	if e != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 {
		return "", fmt.Errorf("resolved TSM exporter is unavailable; build the pinned reader")
	}
	return filepath.Abs(path)
}
func selectedPolicy(security map[string][]string, actual *resolvedPolicy, name string) (policyCheck, error) {
	expected := &resolvedPolicy{When: []policyPredicate{}, Aspects: []string{}, Preconditions: []policyPrecondition{}}
	for key := range security {
		if !evContains([]string{"secret-when", "secret-args", "secret-aspect", "secret-unless", "secret-precondition"}, key) {
			return policyCheck{}, fmt.Errorf("unsupported selected resolved TSM policy forms: %s", name)
		}
	}
	splitFirst := func(value string) []string { return regexp.MustCompile(`\s+`).Split(strings.TrimSpace(value), 2) }
	for _, text := range security["secret-when"] {
		if strings.TrimSpace(text) == "" {
			return policyCheck{}, fmt.Errorf("empty selected resolved TSM predicate: %s", name)
		}
		words := splitFirst(text)
		var doc *string
		if len(words) > 1 {
			doc = &words[1]
		}
		expected.When = append(expected.When, policyPredicate{Name: words[0], Doc: doc})
	}
	for _, key := range []string{"secret-args", "secret-unless"} {
		values := security[key]
		if len(values) > 1 {
			return policyCheck{}, fmt.Errorf("ambiguous selected resolved TSM policy: %s", name)
		}
		if len(values) == 0 {
			continue
		}
		words := strings.Fields(values[0])
		if len(words) == 0 {
			return policyCheck{}, fmt.Errorf("empty selected resolved TSM policy: %s", name)
		}
		if key == "secret-args" {
			policy, exists := map[string]string{"none": "NotAllowed", "untainted": "AllowedWhenUntainted", "tainted": "AllowedWhenTainted"}[words[0]]
			if !exists {
				return policyCheck{}, fmt.Errorf("unknown selected resolved TSM argument policy: %s", name)
			}
			expected.Args = &policyArguments{Policy: policy, Params: words[1:]}
		} else {
			expected.Unless = &policyUnless{Param: words[0], Values: words[1:]}
		}
	}
	if values, exists := security["secret-aspect"]; exists {
		expected.Aspects = values
	}
	for _, text := range security["secret-precondition"] {
		if strings.TrimSpace(text) == "" {
			return policyCheck{}, fmt.Errorf("empty selected resolved TSM precondition: %s", name)
		}
		words := splitFirst(text)
		rest := ""
		if len(words) > 1 {
			rest = words[1]
		}
		tail := splitFirst(rest)
		var failure, doc *string
		if len(tail) > 0 && evContains([]string{"ReturnNothing", "ReturnWithError", "Error"}, tail[0]) {
			failure = &tail[0]
			if len(tail) > 1 && tail[1] != "" {
				doc = &tail[1]
			}
		} else if rest != "" {
			doc = &rest
		}
		expected.Preconditions = append(expected.Preconditions, policyPrecondition{Name: words[0], Failure: failure, Doc: doc})
	}
	if len(security) == 0 {
		expected = nil
	}
	if !evEqual(expected, actual) {
		return policyCheck{}, fmt.Errorf("resolved TSM selected-policy mismatch: %s", name)
	}
	return policyCheck{State: "match", SourceSecuritySHA256: evDigest(security), ResolvedSecretSHA256: evDigest(actual)}, nil
}
func (p Project) validateResolved(export resolvedExport, requests []resolvedRequest, inventory map[string][]tsmStub, files map[string]string) (map[string]resolvedBinding, error) {
	locked, err := p.resolvedIdentity()
	if err != nil {
		return nil, err
	}
	if export.SchemaVersion != 1 || export.TSMCommit != locked.Commit || export.BlobVersion != locked.BlobVersion || export.BlobSHA256 != locked.BlobSHA256 || export.BlobSize != locked.BlobSize {
		return nil, fmt.Errorf("incompatible resolved TSM export provenance")
	}
	sources, err := p.exporterSources()
	if err != nil {
		return nil, err
	}
	if !evEqual(export.ExporterSources, sources) {
		return nil, fmt.Errorf("stale resolved TSM exporter; rebuild the pinned-schema reader")
	}
	requested := map[string]string{}
	for _, request := range requests {
		if _, exists := requested[request.Symbol]; exists {
			return nil, fmt.Errorf("duplicate resolved TSM requests")
		}
		requested[request.Symbol] = request.TSMSymbol
	}
	if !evEqual(sortedKeys(export.Bindings), sortedKeys(requested)) {
		return nil, fmt.Errorf("resolved TSM binding inventory mismatch")
	}
	result := map[string]resolvedBinding{}
	for _, name := range sortedKeys(requested) {
		tsmName := requested[name]
		row := export.Bindings[name]
		if row.Symbol != name || row.TSMSymbol != tsmName {
			return nil, fmt.Errorf("resolved TSM symbol identity mismatch: %s", name)
		}
		location := row.Location
		text, exists := files[location.Path]
		if !exists {
			return nil, fmt.Errorf("unknown resolved TSM source: %s", location.Path)
		}
		if location.Start < 0 || location.Start >= location.End || location.End > len(text) {
			return nil, fmt.Errorf("invalid resolved TSM source location: %s", name)
		}
		line := strings.Count(text[:location.Start], "\n") + 1
		candidates := []tsmStub{}
		for _, candidate := range inventory[tsmName] {
			if candidate.Source == location.Path && candidate.Line == line {
				candidates = append(candidates, candidate)
			}
		}
		if len(candidates) != 1 {
			return nil, fmt.Errorf("resolved TSM source does not identify one archived candidate: %s", name)
		}
		candidate := candidates[0]
		if row.Function == nil {
			return nil, fmt.Errorf("resolved TSM function facts are incomplete: %s", name)
		}
		check, err := selectedPolicy(candidate.Security, row.Function.Secret, name)
		if err != nil {
			return nil, err
		}
		row.Source = location.Path
		row.Line = line
		row.Candidate = &candidate
		row.CandidateSHA256 = evDigest(candidate)
		row.PolicyCheck = &check
		result[name] = row
	}
	return result, nil
}
func (p Project) loadResolved(ctx context.Context, requests []resolvedRequest, inventory map[string][]tsmStub, files map[string]string) (resolvedEvidence, error) {
	locked, e := p.resolvedIdentity()
	if e != nil {
		return resolvedEvidence{}, e
	}
	blobPath := filepath.Join(p.Root, "references/wow-api/tsm/precomputed.bin.zst")
	data, e := readEvidenceFile(blobPath, 64<<20)
	if e != nil {
		return resolvedEvidence{}, e
	}
	if len(data) != locked.BlobSize || sha256hex(data) != locked.BlobSHA256 {
		return resolvedEvidence{}, fmt.Errorf("resolved TSM blob integrity failure")
	}
	artifactPath := filepath.Join(p.Root, "references/wow-api/tsm/resolved-bindings.json")
	artifact, e := readEvidenceFile(artifactPath, 16<<20)
	if e != nil {
		return resolvedEvidence{}, e
	}
	expected, e := decodeEvidenceJSON[resolvedExport](artifact)
	if e != nil {
		return resolvedEvidence{}, e
	}
	decodeRequests := []resolvedRequest{}
	for _, name := range sortedKeys(expected.Bindings) {
		decodeRequests = append(decodeRequests, resolvedRequest{Symbol: name, TSMSymbol: expected.Bindings[name].TSMSymbol})
	}
	encoded, e := jsonDefault(decodeRequests)
	if e != nil {
		return resolvedEvidence{}, e
	}
	tool, e := p.resolvedExecutable(ctx)
	if e != nil {
		return resolvedEvidence{}, e
	}
	tmp, e := os.MkdirTemp("", "tsm-reader-")
	if e != nil {
		return resolvedEvidence{}, e
	}
	defer os.RemoveAll(tmp)
	requestPath := filepath.Join(tmp, "requests.json")
	if e = os.WriteFile(requestPath, encoded, 0600); e != nil {
		return resolvedEvidence{}, e
	}
	rawBytes, e := runEvidenceCommand(ctx, []string{tool, blobPath, requestPath}, nil, 16<<20)
	if e != nil {
		return resolvedEvidence{}, fmt.Errorf("resolved TSM exporter failed: %w", e)
	}
	raw, e := decodeEvidenceJSON[resolvedExport](rawBytes)
	if e != nil {
		return resolvedEvidence{}, e
	}
	if !evEqual(raw, expected) {
		return resolvedEvidence{}, fmt.Errorf("stale resolved TSM binding evidence; review and regenerate its export")
	}
	names := map[string]bool{}
	for _, row := range requests {
		names[row.Symbol] = true
	}
	selected := map[string]resolvedBinding{}
	for name, row := range raw.Bindings {
		if names[name] {
			selected[name] = row
		}
	}
	raw.Bindings = selected
	bindings, e := p.validateResolved(raw, requests, inventory, files)
	if e != nil {
		return resolvedEvidence{}, e
	}
	exporter, e := p.exporterSources()
	if e != nil {
		return resolvedEvidence{}, e
	}
	sources := resolvedSources{ArtifactSHA256: sha256hex(artifact), BlobSHA256: locked.BlobSHA256, BlobSize: locked.BlobSize, BlobVersion: locked.BlobVersion, Commit: locked.Commit, ExporterSources: exporter, Repository: locked.Repository}
	return resolvedEvidence{Bindings: bindings, Sources: sources}, nil
}
