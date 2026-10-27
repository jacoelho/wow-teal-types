package release

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"wow-teal-types/internal/sourcelock"
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
	if e := writeObject(filepath.Join(root, "internal/wowapi/inputs/client_sources.json"), lock); e != nil {
		t.Fatal(e)
	}
	doc := []byte("local A = {}\nAPIDocumentation:AddDocumentationTable(A)\n")
	version := []byte("12.1.5.70575\n")
	source := sourcelock.Sources{"blizzard": sourceFixture("A.lua", doc), "blizzard_version": sourceFixture("version.txt", version)}
	if e := writeObject(filepath.Join(root, "internal/wowapi/inputs/sources.json"), source); e != nil {
		t.Fatal(e)
	}
	p, e := Projection(root, snapshot, filepath.Join(root, "internal/wowapi/inputs/client_sources.json"))
	if e != nil {
		t.Fatal(e)
	}
	path, e := p.InputPath("blizzard", "A.lua")
	if e != nil {
		t.Fatal(e)
	}
	b, e := os.ReadFile(path)
	if e != nil || !bytes.Equal(b, doc) {
		t.Fatalf("locked representation lost: %q %v", b, e)
	}
	raw, e := os.ReadFile(filepath.Join(snapshot, "raw/Interface/AddOns/Blizzard_APIDocumentationGenerated/A.lua"))
	if e != nil || !bytes.Contains(raw, []byte("\r\n")) {
		t.Fatal("raw client bytes changed")
	}
	if _, e = Projection(root, snapshot, filepath.Join(root, "internal/wowapi/inputs/client_sources.json")); e != nil {
		t.Fatal(e)
	}
	selected := []byte("local A = {}\nAPIDocumentation:AddDocumentationTable(A)\n\n")
	source["blizzard"] = sourceFixture("A.lua", selected)
	if err := writeObject(filepath.Join(root, "internal/wowapi/inputs/sources.json"), source); err != nil {
		t.Fatal(err)
	}
	next, err := Projection(root, snapshot, filepath.Join(root, "internal/wowapi/inputs/client_sources.json"))
	if err != nil {
		t.Fatal(err)
	}
	nextPath, err := next.InputPath("blizzard", "A.lua")
	if err != nil {
		t.Fatal(err)
	}
	b, err = os.ReadFile(nextPath)
	if err != nil || !bytes.Equal(b, selected) {
		t.Fatalf("new source selection = %q, %v; want %q", b, err, selected)
	}
	b, err = os.ReadFile(path)
	if err != nil || !bytes.Equal(b, doc) {
		t.Fatalf("previous source selection changed: %q, %v", b, err)
	}
	fixtureFile(t, filepath.Dir(nextPath), "A.lua", []byte("modified"))
	if _, e = Projection(root, snapshot, filepath.Join(root, "internal/wowapi/inputs/client_sources.json")); e == nil {
		t.Fatal("modified projection accepted")
	}
}
func TestBootstrapUsesVerifiedProjectionDespiteStaleOrdinaryCache(t *testing.T) {
	root, snapshot, lock := snapshotFixture(t)
	if err := writeObject(filepath.Join(root, "internal/wowapi/inputs/client_sources.json"), lock); err != nil {
		t.Fatal(err)
	}
	doc := []byte("local A = {}\nAPIDocumentation:AddDocumentationTable(A)\n")
	sources := sourcelock.Sources{
		"blizzard":         sourceFixture("A.lua", doc),
		"blizzard_version": sourceFixture("version.txt", []byte("12.1.5.70575\n")),
	}
	if err := writeObject(filepath.Join(root, "internal/wowapi/inputs/sources.json"), sources); err != nil {
		t.Fatal(err)
	}
	old := []byte("previous build\n")
	ordinary := "references/wow-api/blizzard/A.lua"
	fixtureFile(t, root, ordinary, old)
	p, err := Bootstrap(t.Context(), root, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	path, err := p.InputPath("blizzard", "A.lua")
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(got, doc) {
		t.Fatalf("verified projection = %q, %v; want %q", got, err, doc)
	}
	got, err = os.ReadFile(filepath.Join(root, ordinary))
	if err != nil || !bytes.Equal(got, old) {
		t.Fatalf("ordinary cache changed: %q, %v", got, err)
	}
	if _, err := os.Stat(filepath.Join(root, "references/wow-api/version/version.txt")); !os.IsNotExist(err) {
		t.Fatalf("unexpected native cache mirror: %v", err)
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

func sourceFixture(name string, data []byte) sourcelock.SourceGroup {
	return sourcelock.SourceGroup{Repository: "fixture/source", Commit: strings.Repeat("a", 40), Files: []sourcelock.SourceFile{{Name: name, SHA: blob(data), Size: len(data)}}}
}
