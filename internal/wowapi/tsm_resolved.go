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
	"strconv"
	"strings"
	"time"
)

func evidenceInteger(v any) (int, bool) {
	switch x := v.(type) {
	case int:
		return x, true
	case int64:
		return int(x), int64(int(x)) == x
	case json.Number:
		n, e := strconv.ParseInt(string(x), 10, 64)
		return int(n), e == nil && int64(int(n)) == n
	}
	return 0, false
}

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
func decodeEvidenceJSON(data []byte) (M, error) {
	var m M
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	if e := d.Decode(&m); e != nil {
		return nil, e
	}
	var extra any
	if e := d.Decode(&extra); e != io.EOF {
		return nil, fmt.Errorf("trailing evidence JSON data")
	}
	if m == nil {
		return nil, fmt.Errorf("evidence requires JSON object")
	}
	return m, nil
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
func (p Project) resolvedIdentity() (M, error) {
	data, e := readEvidenceFile(filepath.Join(p.Root, "scripts/wow_api/tsm_sources.json"), 2<<20)
	if e != nil {
		return nil, e
	}
	lock, e := decodeEvidenceJSON(data)
	if e != nil {
		return nil, e
	}
	if lock["repository"] != "TradeSkillMaster/wowlua-ls" || !regexp.MustCompile(`^[0-9a-f]{40}$`).MatchString(asString(lock["commit"])) || asInt(lock["blob_version"]) != 44 {
		return nil, fmt.Errorf("invalid resolved TSM source identity or reader schema")
	}
	var blob M
	for _, v := range asList(lock["files"]) {
		row := asMap(v)
		if row == nil {
			return nil, fmt.Errorf("invalid resolved TSM source file inventory")
		}
		if row["name"] == "precomputed.bin.zst" {
			if blob != nil {
				return nil, fmt.Errorf("resolved TSM source lock requires one resolved blob")
			}
			blob = row
		}
	}
	size, validSize := evidenceInteger(blob["size"])
	version, validVersion := evidenceInteger(lock["blob_version"])
	if !validVersion || version != 44 {
		return nil, fmt.Errorf("unsupported resolved TSM reader schema")
	}
	sha := asString(blob["sha256"])
	if !validSize || size < 8 || size > 64<<20 || !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(sha) {
		return nil, fmt.Errorf("invalid resolved TSM blob identity")
	}
	return M{"repository": lock["repository"], "commit": lock["commit"], "blob_version": lock["blob_version"], "blob_sha256": sha, "blob_size": blob["size"]}, nil
}
func (p Project) exporterSources() (M, error) {
	r := M{}
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
		metadata, e := decodeEvidenceJSON(data)
		if e != nil {
			return "", e
		}
		path = filepath.Join(asString(metadata["target_directory"]), "debug/tsm-resolved-export")
	}
	info, e := os.Stat(path)
	if e != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 {
		return "", fmt.Errorf("resolved TSM exporter is unavailable; build the pinned reader")
	}
	return filepath.Abs(path)
}
func selectedPolicy(security M, actual any, name string) (M, error) {
	expected := M{"when": []any{}, "args": nil, "aspects": []any{}, "guard": nil, "unless": nil, "clears": nil, "restriction_guard": nil, "satisfies": nil, "preconditions": []any{}}
	for key := range security {
		if !evContains([]string{"secret-when", "secret-args", "secret-aspect", "secret-unless", "secret-precondition"}, key) {
			return nil, fmt.Errorf("unsupported selected resolved TSM policy forms: %s", name)
		}
	}
	splitFirst := func(s string) []string { return regexp.MustCompile(`\s+`).Split(strings.TrimSpace(s), 2) }
	for _, text := range evStrings(security["secret-when"]) {
		if strings.TrimSpace(text) == "" {
			return nil, fmt.Errorf("empty selected resolved TSM predicate: %s", name)
		}
		words := splitFirst(text)
		var doc any
		if len(words) > 1 {
			doc = words[1]
		}
		expected["when"] = append(asList(expected["when"]), M{"name": words[0], "doc": doc})
	}
	for _, pair := range [][2]string{{"secret-args", "args"}, {"secret-unless", "unless"}} {
		values := evStrings(security[pair[0]])
		if len(values) > 1 {
			return nil, fmt.Errorf("ambiguous selected resolved TSM policy: %s", name)
		}
		if len(values) == 0 {
			continue
		}
		words := strings.Fields(values[0])
		if len(words) == 0 {
			return nil, fmt.Errorf("empty selected resolved TSM policy: %s", name)
		}
		if pair[1] == "args" {
			policy, ok := map[string]string{"none": "NotAllowed", "untainted": "AllowedWhenUntainted", "tainted": "AllowedWhenTainted"}[words[0]]
			if !ok {
				return nil, fmt.Errorf("unknown selected resolved TSM argument policy: %s", name)
			}
			expected["args"] = M{"policy": policy, "params": evList(words[1:])}
		} else {
			expected["unless"] = M{"param": words[0], "values": evList(words[1:])}
		}
	}
	expected["aspects"] = evGet(security, "secret-aspect", []any{})
	for _, text := range evStrings(security["secret-precondition"]) {
		if strings.TrimSpace(text) == "" {
			return nil, fmt.Errorf("empty selected resolved TSM precondition: %s", name)
		}
		words := splitFirst(text)
		rest := ""
		if len(words) > 1 {
			rest = words[1]
		}
		tail := splitFirst(rest)
		var failure, doc any
		if len(tail) > 0 && evContains([]string{"ReturnNothing", "ReturnWithError", "Error"}, tail[0]) {
			failure = tail[0]
			if len(tail) > 1 && tail[1] != "" {
				doc = tail[1]
			}
		} else if rest != "" {
			doc = rest
		}
		expected["preconditions"] = append(asList(expected["preconditions"]), M{"name": words[0], "failure": failure, "doc": doc})
	}
	var expectedValue any = expected
	if len(security) == 0 {
		expectedValue = nil
	}
	if !evEqual(expectedValue, actual) {
		return nil, fmt.Errorf("resolved TSM selected-policy mismatch: %s", name)
	}
	return M{"state": "match", "source_security_sha256": evDigest(security), "resolved_secret_sha256": evDigest(actual)}, nil
}
func (p Project) validateResolved(export M, requests []M, inventory map[string][]M, files map[string]string) (M, error) {
	locked, e := p.resolvedIdentity()
	if e != nil {
		return nil, e
	}
	if asInt(export["schema_version"]) != 1 || export["tsm_commit"] != locked["commit"] || !evEqual(export["blob_version"], locked["blob_version"]) || export["blob_sha256"] != locked["blob_sha256"] || !evEqual(export["blob_size"], locked["blob_size"]) {
		return nil, fmt.Errorf("incompatible resolved TSM export provenance")
	}
	sources, e := p.exporterSources()
	if e != nil {
		return nil, e
	}
	if !evEqual(export["exporter_sources"], sources) {
		return nil, fmt.Errorf("stale resolved TSM exporter; rebuild the pinned-schema reader")
	}
	requested := M{}
	for _, r := range requests {
		symbol := asString(r["symbol"])
		if _, ok := requested[symbol]; ok {
			return nil, fmt.Errorf("duplicate resolved TSM requests")
		}
		requested[symbol] = r["tsm_symbol"]
	}
	bindings := asMap(export["bindings"])
	if !evEqual(evList(sortedKeys(bindings)), evList(sortedKeys(requested))) {
		return nil, fmt.Errorf("resolved TSM binding inventory mismatch")
	}
	result := M{}
	for _, name := range sortedKeys(requested) {
		tsmName := asString(requested[name])
		row := asMap(bindings[name])
		if row["symbol"] != name || row["tsm_symbol"] != tsmName {
			return nil, fmt.Errorf("resolved TSM symbol identity mismatch: %s", name)
		}
		location := asMap(row["location"])
		path := asString(location["path"])
		text, ok := files[path]
		if !ok {
			return nil, fmt.Errorf("unknown resolved TSM source: %s", path)
		}
		start, validStart := evidenceInteger(location["start"])
		end, validEnd := evidenceInteger(location["end"])
		if !validStart || !validEnd || start < 0 || start >= end || end > len(text) {
			return nil, fmt.Errorf("invalid resolved TSM source location: %s", name)
		}
		line := strings.Count(text[:start], "\n") + 1
		candidates := []M{}
		for _, v := range inventory[tsmName] {
			if v["source"] == path && asInt(v["line"]) == line {
				candidates = append(candidates, v)
			}
		}
		if len(candidates) != 1 {
			return nil, fmt.Errorf("resolved TSM source does not identify one archived candidate: %s", name)
		}
		candidate := candidates[0]
		function := asMap(row["function"])
		_, hasSecret := function["secret"]
		if function == nil || asList(function["overloads"]) == nil || !hasSecret {
			return nil, fmt.Errorf("resolved TSM function facts are incomplete: %s", name)
		}
		check, e := selectedPolicy(asMap(candidate["security"]), function["secret"], name)
		if e != nil {
			return nil, e
		}
		result[name] = evMerge(asMap(cloneValue(row)), M{"source": path, "line": line, "candidate": cloneValue(candidate), "candidate_sha256": evDigest(candidate), "policy_check": check})
	}
	return result, nil
}
func (p Project) loadResolved(ctx context.Context, requests []M, inventory map[string][]M, files map[string]string) (M, error) {
	locked, e := p.resolvedIdentity()
	if e != nil {
		return nil, e
	}
	blobPath := filepath.Join(p.Root, "references/wow-api/tsm/precomputed.bin.zst")
	data, e := readEvidenceFile(blobPath, 64<<20)
	if e != nil {
		return nil, e
	}
	if len(data) != asInt(locked["blob_size"]) || sha256hex(data) != locked["blob_sha256"] {
		return nil, fmt.Errorf("resolved TSM blob integrity failure")
	}
	artifactPath := filepath.Join(p.Root, "references/wow-api/tsm/resolved-bindings.json")
	artifact, e := readEvidenceFile(artifactPath, 16<<20)
	if e != nil {
		return nil, e
	}
	expected, e := decodeEvidenceJSON(artifact)
	if e != nil {
		return nil, e
	}
	decodeRequests := []any{}
	for _, name := range sortedKeys(asMap(expected["bindings"])) {
		decodeRequests = append(decodeRequests, M{"symbol": name, "tsm_symbol": asMap(asMap(expected["bindings"])[name])["tsm_symbol"]})
	}
	encoded, e := jsonDefault(decodeRequests)
	if e != nil {
		return nil, e
	}
	tool, e := p.resolvedExecutable(ctx)
	if e != nil {
		return nil, e
	}
	tmp, e := os.MkdirTemp("", "tsm-reader-")
	if e != nil {
		return nil, e
	}
	defer os.RemoveAll(tmp)
	requestPath := filepath.Join(tmp, "requests.json")
	if e = os.WriteFile(requestPath, encoded, 0600); e != nil {
		return nil, e
	}
	rawBytes, e := runEvidenceCommand(ctx, []string{tool, blobPath, requestPath}, nil, 16<<20)
	if e != nil {
		return nil, fmt.Errorf("resolved TSM exporter failed: %w", e)
	}
	raw, e := decodeEvidenceJSON(rawBytes)
	if e != nil {
		return nil, e
	}
	if !evEqual(raw, expected) {
		return nil, fmt.Errorf("stale resolved TSM binding evidence; review and regenerate its export")
	}
	names := map[string]bool{}
	for _, row := range requests {
		names[asString(row["symbol"])] = true
	}
	selected := M{}
	for name, row := range asMap(raw["bindings"]) {
		if names[name] {
			selected[name] = row
		}
	}
	raw["bindings"] = selected
	bindings, e := p.validateResolved(raw, requests, inventory, files)
	if e != nil {
		return nil, e
	}
	exporter, e := p.exporterSources()
	if e != nil {
		return nil, e
	}
	sources := evMerge(locked, M{"artifact_sha256": sha256hex(artifact), "exporter_sources": exporter})
	return M{"bindings": bindings, "sources": sources}, nil
}
