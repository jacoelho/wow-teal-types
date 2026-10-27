package release

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"wow-teal-types/internal/wowapi"
)

func snapshotFixture(t *testing.T) (string, string, M) {
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
	rows := []any{}
	for _, name := range keys(files) {
		b := files[name]
		fixtureFile(t, snapshot, "raw/"+name, b)
		rows = append(rows, M{"path": name, "size": len(b), "sha256": sha(b), "content_key": contentKey(b)})
	}
	record := M{"schema": 1, "product": "wow", "region": "eu", "locale": "enUS", "build": 70575, "version": "12.1.5.70575", "build_config": strings.Repeat("a", 32), "cdn_config": strings.Repeat("b", 32), "product_config": strings.Repeat("c", 32), "manifests": manifests, "extractor": M{"source_sha256": "pinned"}, "files": rows}
	if e := writeObject(filepath.Join(snapshot, "snapshot.json"), record); e != nil {
		t.Fatal(e)
	}
	lock, e := readObject(filepath.Join(snapshot, "snapshot.json"))
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
				lock["manifests"] = M{"Interface/ui-code-list.txt": 0}
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
	row := func(n string, b []byte) M { return M{"name": n, "sha": blob(b), "size": len(b)} }
	source := M{"blizzard": M{"files": []any{row("A.lua", doc)}}, "blizzard_support": M{"files": []any{row("Blizzard_SharedXML/UI.xsd", support)}}, "blizzard_version": M{"files": []any{row("version.txt", version)}}}
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
	e := parallel(ctx, []int{1, 2, 3}, func(context.Context, int) error { called++; return nil })
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
	source := M{"blizzard": M{"files": []any{M{"name": "A.lua", "sha": blob(literal), "size": len(literal)}}}}
	if e := writeObject(filepath.Join(root, "scripts/wow_api/sources.json"), source); e != nil {
		t.Fatal(e)
	}
	if e := writeObject(filepath.Join(root, "scripts/wow_api/metadata_fields.json"), M{"schema_version": 1, "value_types": M{"$": []any{"object"}, "$.Name": []any{"string"}}}); e != nil {
		t.Fatal(e)
	}
	fixtureFile(t, root, "evidence/generated/wow_native_metadata.json", []byte(`{"copied_old_build":true}`))
	e := prepareContractReview(wowapi.NewProject(root), filepath.Join(root, "missing-baseline"), true)
	if e == nil || !strings.Contains(e.Error(), "unreviewed native metadata") {
		t.Fatalf("unreviewed metadata accepted: %v", e)
	}
	report, e := readObject(filepath.Join(root, "evidence/generated/wow_native_metadata.json"))
	if e != nil {
		t.Fatal(e)
	}
	docs := arr(report["documents"])
	if len(docs) != 1 || obj(obj(docs[0])["metadata"])["NewPermission"] != false || len(arr(report["unknown_fields"])) != 1 {
		t.Fatalf("rejected metadata was lost: %v", report)
	}
}

func TestConcurrentPinnedSecondaryAcquisitionIsIndependent(t *testing.T) {
	root := t.TempDir()
	commit := strings.Repeat("a", 40)
	rows := []any{}
	for n := range 8 {
		name := fmt.Sprintf("input-%d.lua", n)
		b := []byte(name)
		fixtureFile(t, root, "references/wow-api/tsm/"+name, b)
		rows = append(rows, M{"name": name, "path": name, "sha256": sha(b), "size": len(b)})
	}
	version := []byte("12.1.5.70575\n")
	fixtureFile(t, root, "references/wow-api/tsm/blizzard-version.txt", version)
	if e := writeObject(filepath.Join(root, "scripts/wow_api/tsm_sources.json"), M{"repository": "test/repo", "commit": commit, "files": rows, "blizzard": M{"repository": "test/repo", "commit": commit, "version_sha256": sha(version)}}); e != nil {
		t.Fatal(e)
	}
	for _, name := range []string{"native_crosscheck_sources.json", "native_backup_sources.json"} {
		if e := writeObject(filepath.Join(root, "scripts/wow_api", name), M{"files": []any{}}); e != nil {
			t.Fatal(e)
		}
	}
	if e := syncSecondary(t.Context(), root, false); e != nil {
		t.Fatal(e)
	}
	obj(rows[3])["path"] = "../escape"
	if e := writeObject(filepath.Join(root, "scripts/wow_api/tsm_sources.json"), M{"repository": "test/repo", "commit": commit, "files": rows, "blizzard": M{"repository": "test/repo", "commit": commit, "version_sha256": sha(version)}}); e != nil {
		t.Fatal(e)
	}
	if e := syncSecondary(t.Context(), root, false); e == nil {
		t.Fatal("invalid concurrent source path was accepted")
	}
}
