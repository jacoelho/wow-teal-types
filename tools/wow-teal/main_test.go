package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"wow-teal-types/internal/consumer"
)

// Malformed command lines fail before any package, compiler or file access, so
// an empty root proves no command reached the consumer.
func TestMalformedCommandLinesFailBeforeConsumer(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want string
	}{
		{nil, "expected version, verify"},
		{[]string{"publish"}, "unknown command: publish"},
		{[]string{"version", "extra"}, "version takes no arguments"},
		{[]string{"verify", "extra"}, "verify takes no arguments"},
		{[]string{"check-types", "extra"}, "check-types takes no arguments"},
		{[]string{"package", "extra"}, "package takes no arguments"},
		{[]string{"rockspecs", "--refresh-template"}, "unknown rockspec option: --refresh-template"},
		{[]string{"rockspecs", "--output"}, "missing argument for --output"},
		{[]string{"rockspecs", "--repository"}, "missing argument for --repository"},
		{[]string{"rockspecs", "--description-json"}, "missing argument for --description-json"},
		{[]string{"check", "src", "-I"}, "missing argument for -I"},
		{[]string{"check", "--global-env-def"}, "missing argument for --global-env-def"},
		{[]string{"check", "--strict", "src"}, "unknown option: --strict"},
		{[]string{"check", "src", "--output", "out"}, "check does not accept --output"},
		{[]string{"build", "src", "-o"}, "missing argument for -o"},
		{[]string{"check", "src", "--include"}, "missing argument for --include"},
		{[]string{"build", "src", "--output"}, "missing argument for --output"},
	} {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			var out, errout bytes.Buffer
			err := run(t.Context(), t.TempDir(), tc.args, &out, &errout)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got %v, want %q", err, tc.want)
			}
			if out.Len() != 0 || errout.Len() != 0 {
				t.Fatalf("rejected command wrote output: %q %q", out.String(), errout.String())
			}
		})
	}
}

// The archive script captures `wow-teal version` as the package version, so the
// output must be exactly the manifest's version line.
func TestVersionPrintsPackageVersionLine(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, consumer.ManifestFile))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Version string `json:"version"`
	}
	if err = json.Unmarshal(data, &manifest); err != nil || manifest.Version == "" {
		t.Fatalf("manifest.json version: %q %v", manifest.Version, err)
	}
	var out bytes.Buffer
	if err = run(t.Context(), root, []string{"version"}, &out, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if out.String() != manifest.Version+"\n" {
		t.Fatalf("version printed %q, want %q", out.String(), manifest.Version+"\n")
	}
}
