// Package wowapi converts verified native and annotation inputs into Teal declarations.
package wowapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path"
	"path/filepath"
)

// Root-relative, slash-separated locations of the generator inputs. Release
// tooling reads and rewrites the same lock files.
const (
	InputsDir         = "internal/wowapi/inputs"
	SourcesFile       = InputsDir + "/sources.json"
	ClientSourcesFile = InputsDir + "/client_sources.json"
	frameXMLFile      = InputsDir + "/framexml.json"
	overridesFile     = InputsDir + "/overrides.json"
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
func (p Project) path(relative string) string {
	return filepath.Join(p.Root, filepath.FromSlash(relative))
}

// SourcesPath is the lock of every source file a generation reads.
func (p Project) SourcesPath() string { return p.path(SourcesFile) }

// ClientSourcesPath is the lock of the extracted client snapshot.
func (p Project) ClientSourcesPath() string { return p.path(ClientSourcesFile) }

// Generate converts the project's locked inputs into the Teal declaration files.
func (p Project) Generate(ctx context.Context) (Artifacts, error) {
	if err := ctx.Err(); err != nil {
		return Artifacts{}, err
	}
	in, err := p.readInputs()
	if err != nil {
		return Artifacts{}, err
	}
	return generate(ctx, in)
}

func generate(ctx context.Context, in generationInputs) (artifacts Artifacts, err error) {
	defer recoverConversion(&err)
	c := newCatalogue()
	c.inputStats.InputFiles = in.inputFiles
	c.loadQuestie(in.questie)
	c.loadNative(in.native)
	if err = c.loadFrameXML(in.frameXML); err != nil {
		return Artifacts{}, err
	}
	c.applyOverrides(in.overrides)
	projection, err := c.finalize(ctx)
	if err != nil {
		return Artifacts{}, err
	}
	files := map[string][]byte{}
	for name, b := range projection.render() {
		files[path.Join("types/wow", name)] = b
	}
	return Artifacts{Files: files, Statistics: c.statistics()}, nil
}

func sha256hex(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }

type CatalogueStatistics struct {
	InputFiles          int `json:"input_files"`
	NativeEvents        int `json:"native_events"`
	NativeFunctions     int `json:"native_functions"`
	QuestieDeclarations int `json:"questie_declarations"`
	TypeExports         int `json:"type_exports"`
}
