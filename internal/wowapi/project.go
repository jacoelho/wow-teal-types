// Package wowapi converts verified native and annotation inputs into Teal declarations.
package wowapi

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"sort"
)

// Project selects explicit, immutable input roots for one generation.
type Project struct {
	Root   string
	inputs map[string]string
}

type Artifacts struct {
	Files      map[string][]byte
	Statistics CatalogueStatistics
}

func NewProject(root string) Project {
	return Project{Root: root, inputs: map[string]string{
		"questie": filepath.Join(root, "references/questie-annotations/Core"),
	}}
}
func (p Project) WithInputs(paths map[string]string) Project {
	next := NewProject(p.Root)
	for k, v := range p.inputs {
		next.inputs[k] = v
	}
	for k, v := range paths {
		next.inputs[k] = v
	}
	return next
}
func (p Project) InputPath(group, name string) (string, error) {
	root := p.inputs[group]
	if root == "" {
		return "", fmt.Errorf("input root not selected: %s", group)
	}
	return filepath.Join(root, filepath.FromSlash(name)), nil
}
func sha256hex(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }
func blobSHA(b []byte) string {
	h := sha1.New()
	fmt.Fprintf(h, "blob %d\x00", len(b))
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

type CatalogueStatistics struct {
	InputFiles          int `json:"input_files"`
	NativeEvents        int `json:"native_events"`
	NativeFunctions     int `json:"native_functions"`
	QuestieDeclarations int `json:"questie_declarations"`
	TypeExports         int `json:"type_exports"`
}
