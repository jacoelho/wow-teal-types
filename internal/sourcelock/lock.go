// Package sourcelock owns the persisted identities of acquired source inputs.
package sourcelock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
)

// Optional string fields must be absent or nonempty; null is never an alias for absence.
type nonemptyString string

func (value *nonemptyString) UnmarshalJSON(data []byte) error {
	var text string
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return fmt.Errorf("expected nonempty string")
	}
	if err := json.Unmarshal(data, &text); err != nil {
		return err
	}
	if text == "" {
		return fmt.Errorf("expected nonempty string")
	}
	*value = nonemptyString(text)
	return nil
}

// OptionalPath distinguishes an absent Git field from a CASC null version path.
type OptionalPath struct {
	Present bool
	Value   *string
}

func (p OptionalPath) IsZero() bool                 { return !p.Present }
func (p OptionalPath) MarshalJSON() ([]byte, error) { return json.Marshal(p.Value) }
func (p *OptionalPath) UnmarshalJSON(data []byte) error {
	p.Present = true
	return json.Unmarshal(data, &p.Value)
}
func ClientPath(path string) OptionalPath {
	p := OptionalPath{Present: true}
	if path != "" {
		p.Value = &path
	}
	return p
}
func (p OptionalPath) Equal(other OptionalPath) bool {
	return p.Present == other.Present && (p.Value == nil && other.Value == nil || p.Value != nil && other.Value != nil && *p.Value == *other.Value)
}

type SourceFile struct {
	ClientPath OptionalPath `json:"client_path,omitzero"`
	Name       string       `json:"name"`
	SHA        string       `json:"sha"`
	Size       int          `json:"size"`
}

func (f *SourceFile) UnmarshalJSON(data []byte) error {
	var row struct {
		ClientPath OptionalPath `json:"client_path"`
		Name       string       `json:"name"`
		SHA        string       `json:"sha"`
		Size       *int         `json:"size"`
	}
	if err := decode(data, &row); err != nil {
		return err
	}
	if row.Size == nil || *row.Size < 0 {
		return fmt.Errorf("source file requires nonnegative integer size")
	}
	if !fs.ValidPath(row.Name) || !hex40.MatchString(row.SHA) {
		return fmt.Errorf("invalid source file identity")
	}
	if row.ClientPath.Present && row.ClientPath.Value != nil && !fs.ValidPath(*row.ClientPath.Value) {
		return fmt.Errorf("invalid source client path")
	}
	*f = SourceFile{row.ClientPath, row.Name, row.SHA, *row.Size}
	return nil
}

type SourceGroup struct {
	ClientLockSHA256 string       `json:"client_lock_sha256,omitempty"`
	Commit           string       `json:"commit,omitempty"`
	Files            []SourceFile `json:"files"`
	Kind             string       `json:"kind,omitempty"`
	Normalization    string       `json:"normalization,omitempty"`
	Path             string       `json:"path"`
	Ref              string       `json:"ref,omitempty"`
	Repository       string       `json:"repository,omitempty"`
}

func (g *SourceGroup) UnmarshalJSON(data []byte) error {
	var value struct {
		ClientLockSHA256 nonemptyString `json:"client_lock_sha256"`
		Commit           nonemptyString `json:"commit"`
		Files            []SourceFile   `json:"files"`
		Kind             nonemptyString `json:"kind"`
		Normalization    nonemptyString `json:"normalization"`
		Path             *string        `json:"path"`
		Ref              nonemptyString `json:"ref"`
		Repository       nonemptyString `json:"repository"`
	}
	if err := decode(data, &value); err != nil {
		return err
	}
	if value.Path == nil {
		return fmt.Errorf("source group requires path")
	}
	if *value.Path != "" && !fs.ValidPath(*value.Path) {
		return fmt.Errorf("invalid source group path")
	}
	*g = SourceGroup{string(value.ClientLockSHA256), string(value.Commit), value.Files, string(value.Kind), string(value.Normalization), *value.Path, string(value.Ref), string(value.Repository)}
	return nil
}

type Sources map[string]SourceGroup

var hex32 = regexp.MustCompile(`^[0-9a-f]{32}$`)
var hex40 = regexp.MustCompile(`^[0-9a-f]{40}$`)
var hex64 = regexp.MustCompile(`^[0-9a-f]{64}$`)

func ParseSources(data []byte) (Sources, error) {
	var sources Sources
	if err := decode(data, &sources); err != nil {
		return nil, err
	}
	if len(sources) == 0 {
		return nil, fmt.Errorf("empty source groups")
	}
	for name, group := range sources {
		if !fs.ValidPath(name) || group.Files == nil {
			return nil, fmt.Errorf("source group %q lacks file inventory", name)
		}
		switch group.Kind {
		case "":
			if group.Repository == "" || !hex40.MatchString(group.Commit) || group.ClientLockSHA256 != "" || group.Normalization != "" {
				return nil, fmt.Errorf("invalid Git source group %s", name)
			}
		case "casc":
			if !hex64.MatchString(group.ClientLockSHA256) || group.Normalization != "lf-final-newline" || group.Commit != "" || group.Repository != "" || group.Ref != "" {
				return nil, fmt.Errorf("invalid CASC source group %s", name)
			}
		default:
			return nil, fmt.Errorf("unsupported source kind %q", group.Kind)
		}
		seen := map[string]bool{}
		for _, file := range group.Files {
			if seen[file.Name] {
				return nil, fmt.Errorf("duplicate source file %s", file.Name)
			}
			seen[file.Name] = true
			if (group.Kind == "casc") != file.ClientPath.Present {
				return nil, fmt.Errorf("source kind/path mismatch: %s", file.Name)
			}
			if group.Kind == "casc" && ((name == "blizzard_version") != (file.ClientPath.Value == nil)) {
				return nil, fmt.Errorf("invalid CASC client path: %s", file.Name)
			}
		}
	}
	return sources, nil
}
func ReadSources(path string) (Sources, error) {
	b, e := read(path)
	if e != nil {
		return nil, e
	}
	v, e := ParseSources(b)
	if e != nil {
		return nil, fmt.Errorf("%s: %w", path, e)
	}
	return v, nil
}

type ClientIdentity struct {
	Build         int    `json:"build"`
	BuildConfig   string `json:"build_config"`
	CDNConfig     string `json:"cdn_config"`
	Locale        string `json:"locale"`
	Product       string `json:"product"`
	ProductConfig string `json:"product_config"`
	Region        string `json:"region"`
	Version       string `json:"version"`
}

func (id ClientIdentity) Validate() error {
	if id.Product != "wow" || id.Locale != "enUS" {
		return fmt.Errorf("unsupported client product/locale")
	}
	if !slices.Contains([]string{"eu", "us", "kr", "cn", "tw"}, id.Region) {
		return fmt.Errorf("unsupported client region")
	}
	if id.Build < 1 || !regexp.MustCompile(fmt.Sprintf(`^\d+\.\d+\.\d+\.%d$`, id.Build)).MatchString(id.Version) {
		return fmt.Errorf("client build/version disagree")
	}
	if !hex32.MatchString(id.BuildConfig) || !hex32.MatchString(id.CDNConfig) || !hex32.MatchString(id.ProductConfig) {
		return fmt.Errorf("invalid client configuration")
	}
	return nil
}

type Extractor struct {
	Implementation string `json:"implementation"`
	SourceSHA256   string `json:"source_sha256"`
}
type ClientFile struct {
	ContentKey string `json:"content_key"`
	Path       string `json:"path"`
	SHA256     string `json:"sha256"`
	Size       int    `json:"size"`
}

func (f *ClientFile) UnmarshalJSON(data []byte) error {
	var row struct {
		ContentKey string `json:"content_key"`
		Path       string `json:"path"`
		SHA256     string `json:"sha256"`
		Size       *int   `json:"size"`
	}
	if e := decode(data, &row); e != nil {
		return e
	}
	if row.Size == nil || *row.Size < 0 || !fs.ValidPath(row.Path) || !hex32.MatchString(row.ContentKey) || !hex64.MatchString(row.SHA256) {
		return fmt.Errorf("invalid client file identity")
	}
	*f = ClientFile{row.ContentKey, row.Path, row.SHA256, *row.Size}
	return nil
}

// Snapshot preserves the alphabetical field order of published client locks.
type Snapshot struct {
	Build         int            `json:"build"`
	BuildConfig   string         `json:"build_config"`
	CDNConfig     string         `json:"cdn_config"`
	Extractor     Extractor      `json:"extractor"`
	Files         []ClientFile   `json:"files"`
	Locale        string         `json:"locale"`
	Manifests     map[string]int `json:"manifests"`
	Product       string         `json:"product"`
	ProductConfig string         `json:"product_config"`
	Region        string         `json:"region"`
	Schema        int            `json:"schema"`
	Version       string         `json:"version"`
}

func (s Snapshot) Identity() ClientIdentity {
	return ClientIdentity{s.Build, s.BuildConfig, s.CDNConfig, s.Locale, s.Product, s.ProductConfig, s.Region, s.Version}
}
func NewSnapshot(id ClientIdentity, extractor Extractor, files []ClientFile, manifests map[string]int) Snapshot {
	return Snapshot{id.Build, id.BuildConfig, id.CDNConfig, extractor, files, id.Locale, manifests, id.Product, id.ProductConfig, id.Region, 1, id.Version}
}
func (s Snapshot) Equal(other Snapshot) bool {
	return s.Identity() == other.Identity() && s.Extractor == other.Extractor && s.Schema == other.Schema && maps.Equal(s.Manifests, other.Manifests) && slices.Equal(s.Files, other.Files)
}
func ParseSnapshot(data []byte) (Snapshot, error) {
	var v Snapshot
	if e := decode(data, &v); e != nil {
		return v, e
	}
	if e := v.Identity().Validate(); e != nil {
		return v, e
	}
	if v.Schema != 1 || v.Files == nil || v.Manifests == nil || v.Extractor.Implementation == "" || !hex64.MatchString(v.Extractor.SourceSHA256) {
		return v, fmt.Errorf("invalid client snapshot schema")
	}
	return v, nil
}
func ReadSnapshot(path string) (Snapshot, error) {
	b, e := read(path)
	if e != nil {
		return Snapshot{}, e
	}
	v, e := ParseSnapshot(b)
	if e != nil {
		return v, fmt.Errorf("%s: %w", path, e)
	}
	return v, nil
}

func decode[T any](data []byte, value *T) error {
	if len(bytes.TrimSpace(data)) == 0 || bytes.TrimSpace(data)[0] != '{' {
		return fmt.Errorf("expected JSON object")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if e := decoder.Decode(value); e != nil {
		return e
	}
	if e := decoder.Decode(new(json.RawMessage)); e != io.EOF {
		return fmt.Errorf("trailing JSON input")
	}
	return nil
}
func read(path string) ([]byte, error) {
	for _, name := range []string{path, filepath.Dir(path)} {
		st, err := os.Lstat(name)
		if err != nil {
			return nil, err
		}
		if st.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("source lock path contains symlink: %s", name)
		}
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !st.Mode().IsRegular() || st.Size() > 128<<20 {
		return nil, fmt.Errorf("nonregular or oversized source lock: %s", path)
	}
	data, err := io.ReadAll(io.LimitReader(f, (128<<20)+1))
	if len(data) > 128<<20 {
		return nil, fmt.Errorf("source lock exceeds size limit")
	}
	return data, err
}

func IsSHA256(value string) bool { return hex64.MatchString(value) }
