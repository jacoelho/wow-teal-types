package release

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"wow-teal-types/internal/sourcelock"
	"wow-teal-types/internal/wowapi"
)

func snapshotFixture(t *testing.T) (string, string, sourcelock.Snapshot) {
	t.Helper()
	root := t.TempDir()
	fixtureFile(t, root, ".wow-version", []byte("12.1.5.70575\n"))
	snapshot := filepath.Join(root, "snapshot")
	files := map[string][]byte{"Interface/AddOns/Blizzard_APIDocumentationGenerated/A.lua": []byte("local A = {}\r\nAPIDocumentation:AddDocumentationTable(A)\r\n"), "Interface/AddOns/Blizzard_SharedXML/UI.xsd": []byte("schema\r\n")}
	body := ""
	for _, name := range keys(files) {
		body += name + "\n"
	}
	for name := range manifests {
		files[name] = []byte(body)
	}
	rows := []sourcelock.ClientFile{}
	for _, name := range keys(files) {
		b := files[name]
		fixtureFile(t, snapshot, "raw/"+name, b)
		rows = append(rows, sourcelock.ClientFile{Path: name, Size: len(b), SHA256: sha(b), ContentKey: contentKey(b)})
	}
	record := sourcelock.NewSnapshot(sourcelock.ClientIdentity{Product: "wow", Region: "eu", Locale: "enUS", Build: 70575, Version: "12.1.5.70575", BuildConfig: strings.Repeat("a", 32), CDNConfig: strings.Repeat("b", 32), ProductConfig: strings.Repeat("c", 32)}, sourcelock.Extractor{Implementation: "fixture", SourceSHA256: strings.Repeat("d", 64)}, rows, manifests)
	if e := writeObject(filepath.Join(snapshot, "snapshot.json"), record); e != nil {
		t.Fatal(e)
	}
	lock, e := sourcelock.ReadSnapshot(filepath.Join(snapshot, "snapshot.json"))
	if e != nil {
		t.Fatal(e)
	}
	return root, snapshot, lock
}
func TestCompleteSnapshotRejectsMissingChangedExtraAndSymlinkInputs(t *testing.T) {
	for _, mode := range []string{"valid", "changed", "missing", "extra", "symlink", "wrong_manifest"} {
		t.Run(mode, func(t *testing.T) {
			_, snapshot, lock := snapshotFixture(t)
			path := filepath.Join(snapshot, "raw/Interface/AddOns/Blizzard_APIDocumentationGenerated/A.lua")
			switch mode {
			case "changed":
				fixtureFile(t, snapshot, "raw/Interface/AddOns/Blizzard_APIDocumentationGenerated/A.lua", []byte("changed"))
			case "missing":
				os.Remove(path)
			case "extra":
				fixtureFile(t, snapshot, "raw/Interface/unlisted.lua", []byte("extra"))
			case "symlink":
				os.Remove(path)
				os.Symlink(filepath.Join(snapshot, "snapshot.json"), path)
			case "wrong_manifest":
				lock.Manifests = map[string]int{"Interface/ui-code-list.txt": 0}
				writeObject(filepath.Join(snapshot, "snapshot.json"), lock)
			}
			_, e := VerifySnapshot(snapshot, nil)
			if mode == "valid" && e != nil {
				t.Fatal(e)
			}
			if mode != "valid" && e == nil {
				t.Fatal("invalid snapshot accepted")
			}
		})
	}
}
func TestProjectionRetainsRawBytesAndReviewedNewlineRepresentation(t *testing.T) {
	root, snapshot, lock := snapshotFixture(t)
	if e := writeObject(filepath.Join(root, "scripts/wow_api/client_sources.json"), lock); e != nil {
		t.Fatal(e)
	}
	doc := []byte("local A = {}\nAPIDocumentation:AddDocumentationTable(A)\n")
	support := []byte("schema\n")
	version := []byte("12.1.5.70575\n")
	source := sourcelock.Sources{"blizzard": sourceFixture("A.lua", doc), "blizzard_support": sourceFixture("Blizzard_SharedXML/UI.xsd", support), "blizzard_version": sourceFixture("version.txt", version)}
	if e := writeObject(filepath.Join(root, "scripts/wow_api/sources.json"), source); e != nil {
		t.Fatal(e)
	}
	p, e := Projection(root, snapshot, filepath.Join(root, "scripts/wow_api/client_sources.json"))
	if e != nil {
		t.Fatal(e)
	}
	b, e := os.ReadFile(p.InputPath("blizzard", "A.lua"))
	if e != nil || !bytes.Equal(b, doc) {
		t.Fatalf("reviewed representation lost: %q %v", b, e)
	}
	raw, e := os.ReadFile(filepath.Join(snapshot, "raw/Interface/AddOns/Blizzard_APIDocumentationGenerated/A.lua"))
	if e != nil || !bytes.Contains(raw, []byte("\r\n")) {
		t.Fatal("raw client bytes changed")
	}
	if _, e = Projection(root, snapshot, filepath.Join(root, "scripts/wow_api/client_sources.json")); e != nil {
		t.Fatal(e)
	}
	fixtureFile(t, filepath.Dir(p.InputPath("blizzard", "A.lua")), "A.lua", []byte("modified"))
	if _, e = Projection(root, snapshot, filepath.Join(root, "scripts/wow_api/client_sources.json")); e == nil {
		t.Fatal("modified projection accepted")
	}
}
func TestBoundedResponsesAndCancelledWorkerPool(t *testing.T) {
	if _, e := readBounded(strings.NewReader("12345"), 4); e == nil {
		t.Fatal("oversized response accepted")
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	called := 0
	e := parallel(ctx, []sourceJob{{}, {}, {}}, func(context.Context, sourceJob) error { called++; return nil })
	if e == nil || called != 0 {
		t.Fatalf("cancelled pool: %d %v", called, e)
	}
}
func TestClientPathAmbiguityIsNotResolvedArbitrarily(t *testing.T) {
	_, e := clientPath("Blizzard_Test/Binding.lua", map[string]bool{"Interface/AddOns/Blizzard_Test/One/Binding.lua": true, "Interface/AddOns/Blizzard_Test/Two/Binding.lua": true})
	if e == nil {
		t.Fatal("ambiguous source accepted")
	}
}

func TestRejectedNewMetadataIsRetainedBeforeEquivalence(t *testing.T) {
	root := t.TempDir()
	literal := []byte("local A = {Name = \"A\", NewPermission = false}\nAPIDocumentation:AddDocumentationTable(A)\n")
	fixtureFile(t, root, "references/wow-api/blizzard/A.lua", literal)
	source := sourcelock.Sources{"blizzard": sourceFixture("A.lua", literal)}
	if e := writeObject(filepath.Join(root, "scripts/wow_api/sources.json"), source); e != nil {
		t.Fatal(e)
	}
	fixtureFile(t, root, "scripts/wow_api/metadata_fields.json", []byte(`{"schema_version":1,"value_types":{"$":["object"],"$.Name":["string"]}}`))
	fixtureFile(t, root, "evidence/generated/wow_native_metadata.json", []byte(`{"copied_old_build":true}`))
	e := prepareContractReview(wowapi.NewProject(root), filepath.Join(root, "missing-baseline"), true)
	if e == nil || !strings.Contains(e.Error(), "unreviewed native metadata") {
		t.Fatalf("unreviewed metadata accepted: %v", e)
	}
	data, e := os.ReadFile(filepath.Join(root, "evidence/generated/wow_native_metadata.json"))
	if e != nil {
		t.Fatal(e)
	}
	var report struct {
		Documents []struct {
			Metadata struct {
				NewPermission *bool `json:"NewPermission"`
			} `json:"metadata"`
		} `json:"documents"`
		UnknownFields []struct {
			Path string `json:"path"`
		} `json:"unknown_fields"`
	}
	if e = json.Unmarshal(data, &report); e != nil {
		t.Fatal(e)
	}
	if len(report.Documents) != 1 || report.Documents[0].Metadata.NewPermission == nil || *report.Documents[0].Metadata.NewPermission || len(report.UnknownFields) != 1 {
		t.Fatalf("rejected metadata was lost: %s", data)
	}

}

func TestConcurrentPinnedSecondaryAcquisitionIsIndependent(t *testing.T) {
	root := t.TempDir()
	commit := strings.Repeat("a", 40)
	rows := []sourcelock.LockedFile{}
	for n := range 8 {
		name := fmt.Sprintf("input-%d.lua", n)
		b := []byte(name)
		fixtureFile(t, root, "references/wow-api/tsm/"+name, b)
		rows = append(rows, sourcelock.LockedFile{Name: name, Path: name, SHA256: sha(b), Size: len(b)})
	}
	version := []byte("12.1.5.70575\n")
	fixtureFile(t, root, "references/wow-api/tsm/blizzard-version.txt", version)
	if e := writeObject(filepath.Join(root, "scripts/wow_api/tsm_sources.json"), sourcelock.TSM{Repository: "test/repo", Commit: commit, Files: rows, Blizzard: sourcelock.TSMBlizzard{Repository: "test/repo", Commit: commit, VersionSHA256: sha(version)}}); e != nil {
		t.Fatal(e)
	}
	cohort := sourcelock.Cohort{Repository: "test/repo", Commit: commit}
	if e := writeObject(filepath.Join(root, "scripts/wow_api/native_crosscheck_sources.json"), sourcelock.Crosscheck{SchemaVersion: 1, Blizzard: cohort, Wowless: cohort, Files: []sourcelock.LockedFile{}}); e != nil {
		t.Fatal(e)
	}
	if e := writeObject(filepath.Join(root, "scripts/wow_api/native_backup_sources.json"), sourcelock.Backup{SchemaVersion: 1, Repository: "test/repo", Commit: commit, Files: []sourcelock.LockedFile{}, Inventory: []sourcelock.SourceFile{}}); e != nil {
		t.Fatal(e)
	}
	if e := syncSecondary(t.Context(), root, false); e != nil {
		t.Fatal(e)
	}
	rows[3].Path = "../escape"
	if e := writeObject(filepath.Join(root, "scripts/wow_api/tsm_sources.json"), sourcelock.TSM{Repository: "test/repo", Commit: commit, Files: rows, Blizzard: sourcelock.TSMBlizzard{Repository: "test/repo", Commit: commit, VersionSHA256: sha(version)}}); e != nil {
		t.Fatal(e)
	}
	if e := syncSecondary(t.Context(), root, false); e == nil {
		t.Fatal("invalid concurrent source path was accepted")
	}
}

func sourceFixture(name string, data []byte) sourcelock.SourceGroup {
	return sourcelock.SourceGroup{Repository: "fixture/source", Commit: strings.Repeat("a", 40), Files: []sourcelock.SourceFile{{Name: name, SHA: blob(data), Size: len(data)}}}
}
