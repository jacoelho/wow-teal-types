// Package wowapi converts verified native and annotation inputs into one audited catalogue.
package wowapi

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type M = map[string]any

// Project selects explicit, immutable input roots for one generation.
type Project struct {
	Root   string
	inputs map[string]string
	reader string
}

type Artifacts struct {
	Files      map[string][]byte
	Statistics M
}

func NewProject(root string) Project {
	return Project{Root: root, inputs: map[string]string{
		"blizzard":         filepath.Join(root, "references/wow-api/blizzard"),
		"blizzard_support": filepath.Join(root, "references/wow-api/support"),
		"blizzard_version": filepath.Join(root, "references/wow-api/version"),
		"questie":          filepath.Join(root, "references/questie-annotations/Core"),
		"questie_runtime":  filepath.Join(root, "references/wow-api/questie-runtime"),
	}}
}
func (p Project) WithInputs(paths map[string]string) Project {
	next := NewProject(p.Root)
	next.reader = p.reader
	for k, v := range p.inputs {
		next.inputs[k] = v
	}
	for k, v := range paths {
		next.inputs[k] = v
	}
	return next
}
func (p Project) InputPath(group, name string) string {
	return filepath.Join(p.inputs[group], filepath.FromSlash(name))
}
func (p Project) Lock(name string) (M, error) {
	return readJSON(filepath.Join(p.Root, "scripts/wow_api", name))
}
func readJSON(path string) (M, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	dec := json.NewDecoder(io.LimitReader(f, 128<<20))
	dec.UseNumber()
	var v M
	if err = dec.Decode(&v); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	var extra any
	if err = dec.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf("%s: trailing JSON data", path)
	}
	if v == nil {
		return nil, fmt.Errorf("%s: expected JSON object", path)
	}
	return v, nil
}
func sha256hex(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }
func blobSHA(b []byte) string {
	h := sha1.New()
	fmt.Fprintf(h, "blob %d\x00", len(b))
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}
func asMap(v any) M         { m, _ := v.(map[string]any); return m }
func asList(v any) []any    { a, _ := v.([]any); return a }
func asString(v any) string { s, _ := v.(string); return s }
func asInt(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	}
	return 0
}
func truth(v any) bool {
	switch x := v.(type) {
	case nil:
		return false
	case bool:
		return x
	case string:
		return x != ""
	case int:
		return x != 0
	case int64:
		return x != 0
	case float64:
		return x != 0
	case json.Number:
		return x != "0" && x != "0.0"
	case []any:
		return len(x) != 0
	case M:
		return len(x) != 0
	}
	return true
}
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
func cloneValue(v any) any {
	switch x := v.(type) {
	case M:
		r := M{}
		for k, v := range x {
			r[k] = cloneValue(v)
		}
		return r
	case []any:
		r := make([]any, len(x))
		for i, v := range x {
			r[i] = cloneValue(v)
		}
		return r
	case []string:
		return append([]string{}, x...)
	}
	return v
}
func relativeName(root, path string) (string, error) {
	r, e := filepath.Rel(root, path)
	if e != nil || r == ".." || strings.HasPrefix(r, "../") {
		return "", fmt.Errorf("path escapes project: %s", path)
	}
	return filepath.ToSlash(r), nil
}

// WithReader selects the independently verified schema-matched TSM reader.
func (p Project) WithReader(binary string) Project { p.reader = binary; return p }

// MetadataJSON retains UTF-8 spelling in the complete native metadata artifact.
func MetadataJSON(report M) ([]byte, error) { return jsonPrettyUTF8(report) }
