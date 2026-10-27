package release

import (
	"bytes"
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
	for _, name := range append(consumer.RequiredFiles(), "package.json", "LICENSE", "README.md", "examples/minimal/WowTealExample.toc", "scripts/wow_api/sources.json", "scripts/wow_api/client_sources.json") {
		copyFile(name)
	}
	provenance, err := readJSON[provenanceRecord](filepath.Join(root, "provenance.json"))
	if err != nil {
		t.Fatal(err)
	}
	for name := range provenance.SourceLockSHA256 {
		if name != "extractor.json" {
			copyFile("scripts/wow_api/" + name)
		}
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
	manifest, err := consumer.ParsePackageManifest(artifacts["package.json"])
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
	stale := strings.Replace(string(artifacts["README.md"]), "Targets Retail **12.1.5.70575**", "Targets Retail **12.1.5.70574**", 1)
	fixtureFile(t, root, "README.md", []byte(stale))
	before := metadataInputs(t, root)
	artifacts, err = prepareMetadata(root)
	if err != nil {
		t.Fatal(err)
	}
	if err = compareArtifacts(root, artifacts); err == nil || !strings.Contains(err.Error(), "stale generated output: README.md") {
		t.Fatalf("expected stale metadata, got %v", err)
	}
	if !maps.Equal(before, metadataInputs(t, root)) {
		t.Fatal("metadata comparison refreshed stale inputs")
	}
}

func TestProvenanceSerializationPreservesApprovedBytes(t *testing.T) {
	root := metadataFixture(t)
	path := filepath.Join(root, "provenance.json")
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	value, err := readJSON[provenanceRecord](path)
	if err != nil {
		t.Fatal(err)
	}
	after, err := jsonBytes(value)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("concrete provenance representation changed approved bytes")
	}
}
