package consumer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"
)

// PackageManifest binds the complete consumer execution closure to one release.
// Field order is part of the published JSON byte contract.
type PackageManifest struct {
	Files         map[string]string `json:"files"`
	License       string            `json:"license"`
	Name          string            `json:"name"`
	SchemaVersion int               `json:"schema_version"`
	TealVersion   string            `json:"teal_version"`
	Version       string            `json:"version"`
	WoWBuild      string            `json:"wow_build"`
}

type BuildReceipt struct {
	Files                 map[string]string `json:"files"`
	PackageManifestSHA256 string            `json:"package_manifest_sha256"`
	PackageVersion        string            `json:"package_version"`
}

type PackageResult struct {
	Archive string `json:"archive"`
	Members int    `json:"members"`
	SHA256  string `json:"sha256"`
}

// jsonString rejects null instead of accepting encoding/json's zero-string
// substitution. Missing optional fields still have their ordinary zero value.
type jsonString string

func (s *jsonString) UnmarshalJSON(data []byte) error {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return fmt.Errorf("expected string, got null")
	}
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*s = jsonString(value)
	return nil
}

type manifestJSON struct {
	Files         map[string]jsonString `json:"files"`
	License       jsonString            `json:"license"`
	Name          jsonString            `json:"name"`
	SchemaVersion int                   `json:"schema_version"`
	TealVersion   jsonString            `json:"teal_version"`
	Version       jsonString            `json:"version"`
	WoWBuild      jsonString            `json:"wow_build"`
}
type receiptJSON struct {
	Files                 map[string]jsonString `json:"files"`
	PackageManifestSHA256 jsonString            `json:"package_manifest_sha256"`
	PackageVersion        jsonString            `json:"package_version"`
}

func fileHashes(files map[string]jsonString) (map[string]string, error) {
	if files == nil {
		return nil, fmt.Errorf("files must be a JSON object")
	}
	result := make(map[string]string, len(files))
	for name, hash := range files {
		if hash == "" {
			return nil, fmt.Errorf("file hash must be a nonempty string: %s", name)
		}
		result[name] = string(hash)
	}
	return result, nil
}

// ParsePackageManifest validates the manifest schema without reading its files.
// Release preparation uses this before updating the cohort's file fingerprints.
func ParsePackageManifest(data []byte) (PackageManifest, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var value *manifestJSON
	if err := decoder.Decode(&value); err != nil {
		return PackageManifest{}, fmt.Errorf("package manifest: %w", err)
	}
	if err := decoder.Decode(new(json.RawMessage)); err != io.EOF {
		return PackageManifest{}, fmt.Errorf("package manifest: trailing JSON input")
	}
	if value == nil {
		return PackageManifest{}, fmt.Errorf("package manifest: expected JSON object")
	}
	if value.License == "" || value.Name == "" || value.TealVersion == "" || value.Version == "" || value.WoWBuild == "" {
		return PackageManifest{}, fmt.Errorf("package manifest: missing required string field")
	}
	if value.SchemaVersion != 1 {
		return PackageManifest{}, fmt.Errorf("package manifest: unsupported schema_version %d", value.SchemaVersion)
	}
	files, err := fileHashes(value.Files)
	if err != nil {
		return PackageManifest{}, fmt.Errorf("package manifest: %w", err)
	}
	return PackageManifest{files, string(value.License), string(value.Name), value.SchemaVersion, string(value.TealVersion), string(value.Version), string(value.WoWBuild)}, nil
}
func readManifest(path string) (PackageManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return PackageManifest{}, err
	}
	manifest, err := ParsePackageManifest(data)
	if err != nil {
		return PackageManifest{}, fmt.Errorf("%s: %w", path, err)
	}
	return manifest, nil
}
func parseBuildReceipt(data []byte) (BuildReceipt, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var value *receiptJSON
	if err := decoder.Decode(&value); err != nil {
		return BuildReceipt{}, fmt.Errorf("build receipt: %w", err)
	}
	if err := decoder.Decode(new(json.RawMessage)); err != io.EOF {
		return BuildReceipt{}, fmt.Errorf("build receipt: trailing JSON input")
	}
	if value == nil {
		return BuildReceipt{}, fmt.Errorf("build receipt: expected JSON object")
	}
	files, err := fileHashes(value.Files)
	if err != nil {
		return BuildReceipt{}, fmt.Errorf("build receipt: %w", err)
	}
	return BuildReceipt{files, string(value.PackageManifestSHA256), string(value.PackageVersion)}, nil
}
func readBuildReceipt(path string) (BuildReceipt, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return BuildReceipt{}, err
	}
	return parseBuildReceipt(data)
}
func sameManifest(a, b PackageManifest) bool {
	return maps.Equal(a.Files, b.Files) && a.License == b.License && a.Name == b.Name && a.SchemaVersion == b.SchemaVersion && a.TealVersion == b.TealVersion && a.Version == b.Version && a.WoWBuild == b.WoWBuild
}

func manifestBytes(value PackageManifest) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
func receiptBytes(value BuildReceipt) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
func writePackageResult(out io.Writer, value PackageResult) error {
	encoder := json.NewEncoder(out)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}
