package wowapi

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"wow-teal-types/internal/sourcelock"
)

func (p Project) readNativeDocuments() ([]NativeDocument, error) {
	lock, err := sourcelock.ReadSources(filepath.Join(p.Root, "scripts/wow_api/sources.json"))
	if err != nil {
		return nil, err
	}
	documents := []NativeDocument{}
	for _, source := range lock["blizzard"].Files {
		data, err := os.ReadFile(p.InputPath("blizzard", source.Name))
		if err != nil {
			return nil, err
		}
		if blobSHA(data) != source.SHA || len(data) != source.Size {
			return nil, fmt.Errorf("modified native source: %s", source.Name)
		}
		if !strings.HasSuffix(source.Name, ".lua") {
			continue
		}
		raw, err := ParseLua(data, source.Name)
		if err != nil {
			return nil, err
		}
		metadata, err := decodeNativeMetadata(raw)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", source.Name, err)
		}
		documents = append(documents, NativeDocument{Name: source.Name, Metadata: metadata})
	}
	return documents, nil
}
