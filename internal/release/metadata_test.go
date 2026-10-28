package release

import (
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"wow-teal-types/internal/consumer"
)

func metadataFixture(t *testing.T) string {
	t.Helper()
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate repository fixture")
	}
	repository := filepath.Clean(filepath.Join(filepath.Dir(source), "../.."))
	root := t.TempDir()
	copyFile := func(name string) {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(repository, name))
		if err != nil {
			t.Fatal(err)
		}
		fixtureFile(t, root, name, data)
	}
	for _, name := range append(consumer.RequiredFiles(), consumer.ManifestFile, "LICENSE", "README.md", "examples/WowTealExample/WowTealExample.toc", "internal/wowapi/inputs/sources.json", "internal/wowapi/inputs/client_sources.json") {
		copyFile(name)
	}
	return root
}

func metadataInputs(t *testing.T, root string) inputState {
	t.Helper()
	state, err := captureInputs(root)
	if err != nil {
		t.Fatal(err)
	}
	return state
}

func TestMetadataPreparationPreservesInputsAndExactVersionBytes(t *testing.T) {
	root := metadataFixture(t)
	fixtureFile(t, root, ".wow-version", []byte("12.1.5.70575"))
	before := metadataInputs(t, root)
	artifacts, err := prepareMetadata(root)
	if err != nil {
		t.Fatal(err)
	}
	if !maps.Equal(before, metadataInputs(t, root)) {
		t.Fatal("metadata preparation changed inputs")
	}
	manifest, err := consumer.ParsePackageManifest(artifacts[consumer.ManifestFile])
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Files[".wow-version"] != sha([]byte("12.1.5.70575")) {
		t.Fatal("metadata fingerprint normalized the version bytes")
	}
	if err = writeCandidateArtifacts(root, artifacts); err != nil {
		t.Fatal(err)
	}
	if _, err = consumer.Verify(root); err != nil {
		t.Fatal(err)
	}
	fixtureFile(t, root, ".wow-version", []byte("12.1.5.70576\n"))
	before = metadataInputs(t, root)
	if _, err = prepareMetadata(root); err == nil || !strings.Contains(err.Error(), "source cohort") {
		t.Fatalf("expected cohort refusal, got %v", err)
	}
	if !maps.Equal(before, metadataInputs(t, root)) {
		t.Fatal("rejected cohort changed inputs")
	}
}

func TestMetadataComparisonRejectsStaleFilesWithoutRefresh(t *testing.T) {
	root := metadataFixture(t)
	artifacts, err := prepareMetadata(root)
	if err != nil {
		t.Fatal(err)
	}
	if err = writeCandidateArtifacts(root, artifacts); err != nil {
		t.Fatal(err)
	}
	if err = compareArtifacts(root, artifacts); err != nil {
		t.Fatal(err)
	}
	toc := "examples/WowTealExample/WowTealExample.toc"
	stale := strings.Replace(string(artifacts[toc]), "## Interface: 120105", "## Interface: 120104", 1)
	fixtureFile(t, root, toc, []byte(stale))
	before := metadataInputs(t, root)
	artifacts, err = prepareMetadata(root)
	if err != nil {
		t.Fatal(err)
	}
	if err = compareArtifacts(root, artifacts); err == nil || !strings.Contains(err.Error(), "stale generated output: "+toc) {
		t.Fatalf("expected stale metadata, got %v", err)
	}
	if !maps.Equal(before, metadataInputs(t, root)) {
		t.Fatal("metadata comparison refreshed stale inputs")
	}
}

func TestPackageManifestWireBytes(t *testing.T) {
	manifest := consumer.PackageManifest{Files: map[string]string{"b": "two", "a": "one"}, License: "MIT", Name: "name<&", SchemaVersion: 1, TealVersion: "0.24.8", Version: "0.1.0", WoWBuild: "12.1.5.70575"}
	actual, err := jsonBytes(manifest)
	if err != nil {
		t.Fatal(err)
	}
	const wantManifest = `{
  "files": {
    "a": "one",
    "b": "two"
  },
  "license": "MIT",
  "name": "name<&",
  "schema_version": 1,
  "teal_version": "0.24.8",
  "version": "0.1.0",
  "wow_build": "12.1.5.70575"
}
`
	if string(actual) != wantManifest {
		t.Fatalf("manifest bytes changed:\n%s", actual)
	}
}
