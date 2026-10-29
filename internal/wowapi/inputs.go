package wowapi

import (
	"fmt"
	"slices"
	"strings"

	"wow-teal-types/internal/sourcelock"
)

// maxInputFile bounds every generation input read.
const maxInputFile = 128 << 20

// generationInputs is every input file of one generation, read and decoded
// before conversion starts. Conversion itself performs no I/O.
type generationInputs struct {
	inputFiles int
	questie    []annotationDocument
	native     []nativeDocument
	frameXML   frameXMLSupplement
	overrides  catalogueOverrides
}

func (p Project) readInputs() (generationInputs, error) {
	lock, err := sourcelock.ReadSources(p.SourcesPath())
	if err != nil {
		return generationInputs{}, err
	}
	in := generationInputs{}
	for _, source := range lock {
		in.inputFiles += len(source.Files)
	}
	if in.questie, err = p.readQuestieDocuments(lock); err != nil {
		return generationInputs{}, err
	}
	if in.native, err = p.readNativeDocuments(lock); err != nil {
		return generationInputs{}, err
	}
	if err = readInputJSON(p.path(frameXMLFile), &in.frameXML); err != nil {
		return generationInputs{}, err
	}
	if err = readInputJSON(p.path(overridesFile), &in.overrides); err != nil {
		return generationInputs{}, err
	}
	return in, nil
}

func readInputJSON(path string, target any) error {
	data, err := sourcelock.ReadRegular(path, maxInputFile)
	if err != nil {
		return err
	}
	if err = sourcelock.DecodeStrict(data, target); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}

// readQuestieDocuments returns the locked annotation files in name order.
func (p Project) readQuestieDocuments(lock sourcelock.Sources) ([]annotationDocument, error) {
	rows := slices.Clone(lock["questie"].Files)
	slices.SortStableFunc(rows, func(a, b sourcelock.SourceFile) int { return strings.Compare(a.Name, b.Name) })
	documents := []annotationDocument{}
	for _, row := range rows {
		if !strings.HasSuffix(row.Name, ".lua") {
			continue
		}
		path, err := p.InputPath("questie", row.Name)
		if err != nil {
			return nil, err
		}
		data, err := sourcelock.ReadRegular(path, maxInputFile)
		if err != nil {
			return nil, err
		}
		doc, err := parseAnnotations(data, row.Name)
		if err != nil {
			return nil, err
		}
		documents = append(documents, doc)
	}
	return documents, nil
}

// readNativeDocuments returns the verified native documentation files in name order.
func (p Project) readNativeDocuments(lock sourcelock.Sources) ([]nativeDocument, error) {
	documents := []nativeDocument{}
	for _, source := range lock["blizzard"].Files {
		path, err := p.InputPath("blizzard", source.Name)
		if err != nil {
			return nil, err
		}
		data, err := sourcelock.ReadRegular(path, maxInputFile)
		if err != nil {
			return nil, err
		}
		if !source.Matches(data) {
			return nil, fmt.Errorf("modified native source: %s", source.Name)
		}
		if !strings.HasSuffix(source.Name, ".lua") {
			continue
		}
		raw, err := parseLua(data, source.Name)
		if err != nil {
			return nil, err
		}
		metadata, err := decodeNativeMetadata(raw)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", source.Name, err)
		}
		documents = append(documents, nativeDocument{Name: source.Name, Metadata: metadata})
	}
	slices.SortStableFunc(documents, func(a, b nativeDocument) int { return strings.Compare(a.Name, b.Name) })
	return documents, nil
}
