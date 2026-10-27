package release

import "wow-teal-types/internal/sourcelock"

type provenanceClient struct {
	Build     int    `json:"build"`
	Flavor    string `json:"flavor"`
	Interface int    `json:"interface"`
	Version   string `json:"version"`
}
type provenanceCompiler struct {
	Compatibility string `json:"compatibility"`
	LuaTarget     string `json:"lua_target"`
	Name          string `json:"name"`
	StrictArity   *bool  `json:"strict_arity"`
	Version       string `json:"version"`
}
type provenanceExport struct {
	CheckerPolicy    string `json:"checker_policy"`
	DeclarationBytes string `json:"declaration_bytes"`
	Generation       string `json:"generation"`
}
type sourceProvenance struct {
	ClientLockSHA256 string `json:"client_lock_sha256,omitempty"`
	Commit           string `json:"commit,omitempty"`
	Kind             string `json:"kind,omitempty"`
	Normalization    string `json:"normalization,omitempty"`
	Ref              string `json:"ref,omitempty"`
	Repository       string `json:"repository,omitempty"`
}

func sourceSummary(source sourcelock.SourceGroup) sourceProvenance {
	return sourceProvenance{source.ClientLockSHA256, source.Commit, source.Kind, source.Normalization, source.Ref, source.Repository}
}

type clientProvenance struct {
	Build         int                  `json:"build"`
	BuildConfig   string               `json:"build_config"`
	CDNConfig     string               `json:"cdn_config"`
	Extractor     sourcelock.Extractor `json:"extractor"`
	Locale        string               `json:"locale"`
	Manifests     map[string]int       `json:"manifests"`
	Product       string               `json:"product"`
	ProductConfig string               `json:"product_config"`
	Region        string               `json:"region"`
	Schema        int                  `json:"schema"`
	Version       string               `json:"version"`
}

func clientSummary(s sourcelock.Snapshot) clientProvenance {
	return clientProvenance{s.Build, s.BuildConfig, s.CDNConfig, s.Extractor, s.Locale, s.Manifests, s.Product, s.ProductConfig, s.Region, s.Schema, s.Version}
}

type provenanceSources struct {
	Blizzard        sourceProvenance `json:"blizzard"`
	BlizzardSupport sourceProvenance `json:"blizzard_support"`
	BlizzardVersion sourceProvenance `json:"blizzard_version"`
	ClientCASC      clientProvenance `json:"client_casc"`
	Questie         sourceProvenance `json:"questie"`
	QuestieRuntime  sourceProvenance `json:"questie_runtime"`
}
type provenanceRecord struct {
	Client           provenanceClient   `json:"client"`
	Compiler         provenanceCompiler `json:"compiler"`
	Export           provenanceExport   `json:"export"`
	Limits           []string           `json:"limits"`
	PackageName      string             `json:"package_name"`
	PackageVersion   string             `json:"package_version"`
	SchemaVersion    int                `json:"schema_version"`
	SourceLockSHA256 map[string]string  `json:"source_lock_sha256"`
	Sources          provenanceSources  `json:"sources"`
}
