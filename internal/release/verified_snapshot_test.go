package release

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"wow-teal-types/internal/sourcelock"
	"wow-teal-types/internal/wowapi"
)

const fixtureDocument = "Interface/AddOns/Blizzard_APIDocumentationGenerated/A.lua"

// A verified snapshot's reads return only bytes that match the verified record,
// however the directory changes after verification.
func TestVerifiedSnapshotReadsOnlyVerifiedBytes(t *testing.T) {
	raw := []byte("local A = {}\r\nAPIDocumentation:AddDocumentationTable(A)\r\n")
	for _, tc := range []struct {
		name, request string
		change        func(t *testing.T, snapshot string)
		wantErr       bool
	}{
		{name: "unchanged", request: "Blizzard_APIDocumentationGenerated/A.lua"},
		{name: "same-size mutation", request: "Blizzard_APIDocumentationGenerated/A.lua", wantErr: true, change: func(t *testing.T, snapshot string) {
			fixtureFile(t, snapshot, "raw/"+fixtureDocument, bytes.ToUpper(raw))
		}},
		{name: "truncated", request: "Blizzard_APIDocumentationGenerated/A.lua", wantErr: true, change: func(t *testing.T, snapshot string) {
			fixtureFile(t, snapshot, "raw/"+fixtureDocument, raw[:len(raw)-1])
		}},
		{name: "extended", request: "Blizzard_APIDocumentationGenerated/A.lua", wantErr: true, change: func(t *testing.T, snapshot string) {
			fixtureFile(t, snapshot, "raw/"+fixtureDocument, append(bytes.Clone(raw), '\n'))
		}},
		{name: "deleted", request: "Blizzard_APIDocumentationGenerated/A.lua", wantErr: true, change: func(t *testing.T, snapshot string) {
			if err := os.Remove(filepath.Join(snapshot, "raw", fixtureDocument)); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "symlink to identical bytes", request: "Blizzard_APIDocumentationGenerated/A.lua", wantErr: true, change: func(t *testing.T, snapshot string) {
			path := filepath.Join(snapshot, "raw", fixtureDocument)
			fixtureFile(t, snapshot, "copy.lua", raw)
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(filepath.Join(snapshot, "copy.lua"), path); err != nil {
				t.Fatal(err)
			}
		}},
		// Files outside the record are absent, even when they appear on disk.
		{name: "created after verification", request: "Blizzard_APIDocumentationGenerated/B.lua", change: func(t *testing.T, snapshot string) {
			fixtureFile(t, snapshot, "raw/Interface/AddOns/Blizzard_APIDocumentationGenerated/B.lua", raw)
		}},
		{name: "escaping name", request: "../../snapshot.json"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, snapshot, _ := snapshotFixture(t)
			verified := mustVerify(t, snapshot)
			if tc.change != nil {
				tc.change(t, snapshot)
			}
			path, data, err := verified.packagedFile(tc.request)
			switch {
			case tc.wantErr:
				if err == nil || data != nil {
					t.Fatalf("packagedFile = %q, %q, %v; want an error and no data", path, data, err)
				}
			case strings.HasSuffix(tc.request, "A.lua"):
				if err != nil || path != fixtureDocument || !bytes.Equal(data, raw) {
					t.Fatalf("packagedFile = %q, %q, %v; want the raw verified bytes", path, data, err)
				}
			default:
				if err != nil || path != "" || data != nil {
					t.Fatalf("packagedFile = %q, %q, %v; want an absent file", path, data, err)
				}
			}
		})
	}
}

// writeNativeSources writes a source selection whose native rows do not match
// the fixture client, so staging replaces them with the client's bytes.
func writeNativeSources(t *testing.T, root string, blizzard sourcelock.SourceGroup) sourcelock.Sources {
	t.Helper()
	sources := sourcelock.Sources{"blizzard": blizzard, "blizzard_version": sourceFixture("version.txt", []byte("12.1.4.0\n"))}
	if err := writeObject(wowapi.NewProject(root).SourcesPath(), sources); err != nil {
		t.Fatal(err)
	}
	return sources
}

func TestStageNativeWritesTheVerifiedRecord(t *testing.T) {
	for _, swap := range []bool{false, true} {
		t.Run(map[bool]string{false: "reformatted receipt", true: "receipt replaced after verification"}[swap], func(t *testing.T) {
			root, snapshot, lock := snapshotFixture(t)
			receipt := filepath.Join(snapshot, "snapshot.json")
			want, err := os.ReadFile(receipt)
			if err != nil {
				t.Fatal(err)
			}
			reformatted, err := json.MarshalIndent(lock, "", "\t")
			if err != nil {
				t.Fatal(err)
			}
			fixtureFile(t, snapshot, "snapshot.json", reformatted)
			writeNativeSources(t, root, sourceFixture("A.lua", []byte("other\n")))
			verified := mustVerify(t, snapshot)
			if swap {
				other := lock
				other.Region = "us"
				if err := writeObject(receipt, other); err != nil {
					t.Fatal(err)
				}
			}
			if err := stageNative(root, verified); err != nil {
				t.Fatal(err)
			}
			p := wowapi.NewProject(root)
			got, err := os.ReadFile(p.ClientSourcesPath())
			if err != nil || !bytes.Equal(got, want) {
				t.Fatalf("client lock = %s, %v; want the canonical verified record %s", got, err, want)
			}
			sources, err := sourcelock.ReadSources(p.SourcesPath())
			if err != nil {
				t.Fatal(err)
			}
			if blizzard := sources["blizzard"]; blizzard.Kind != "casc" || blizzard.ClientLockSHA256 != sha(got) {
				t.Fatalf("blizzard selection = %+v; want casc bound to client lock %s", blizzard, sha(got))
			}
			if _, err := projectSources(root, verified); err != nil {
				t.Fatalf("staged sources do not project: %v", err)
			}
		})
	}
}

// A lock row that matches no representation of the client bytes is replaced,
// including a zero-size row that the empty result would otherwise match.
func TestStageNativeReplacesUnmatchedLockRows(t *testing.T) {
	root, snapshot, _ := snapshotFixture(t)
	writeNativeSources(t, root, sourcelock.SourceGroup{Repository: "fixture/source", Commit: strings.Repeat("a", 40),
		Files: []sourcelock.SourceFile{{Name: "A.lua", SHA: "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391", Size: 0}}})
	if err := stageNative(root, mustVerify(t, snapshot)); err != nil {
		t.Fatal(err)
	}
	sources, err := sourcelock.ReadSources(wowapi.NewProject(root).SourcesPath())
	if err != nil {
		t.Fatal(err)
	}
	normalized := []byte("local A = {}\nAPIDocumentation:AddDocumentationTable(A)\n")
	if rows := sources["blizzard"].Files; len(rows) != 1 || !rows[0].Matches(normalized) {
		t.Fatalf("blizzard rows = %+v; want the normalized client document", rows)
	}
}

// A zero-size lock row must not accept a non-empty client document.
func TestProjectionRejectsZeroSizeRowForNonEmptyDocument(t *testing.T) {
	root, snapshot, lock := snapshotFixture(t)
	if err := writeObject(wowapi.NewProject(root).ClientSourcesPath(), lock); err != nil {
		t.Fatal(err)
	}
	empty := sourcelock.SourceFile{Name: "A.lua", SHA: "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391", Size: 0}
	source := sourcelock.Sources{
		"blizzard":         sourcelock.SourceGroup{Repository: "fixture/source", Commit: strings.Repeat("a", 40), Files: []sourcelock.SourceFile{empty}},
		"blizzard_version": sourceFixture("version.txt", []byte("12.1.5.70575\n")),
	}
	if err := writeObject(wowapi.NewProject(root).SourcesPath(), source); err != nil {
		t.Fatal(err)
	}
	_, err := projectSources(root, mustVerify(t, snapshot))
	if err == nil || !strings.Contains(err.Error(), "client source differs from locked contract") {
		t.Fatalf("projectSources = %v; want a locked-contract mismatch", err)
	}
	if exists(filepath.Join(root, "build/wow-api-source")) {
		t.Fatal("rejected projection was written")
	}
}

// A client file changed after verification fails every consumer of the
// verified snapshot before it writes anything.
func TestSnapshotConsumersRejectFilesChangedAfterVerification(t *testing.T) {
	mutate := func(t *testing.T, snapshot string) {
		b, err := os.ReadFile(filepath.Join(snapshot, "raw", fixtureDocument))
		if err != nil {
			t.Fatal(err)
		}
		fixtureFile(t, snapshot, "raw/"+fixtureDocument, bytes.ToUpper(b))
	}
	t.Run("projectSources", func(t *testing.T) {
		root, snapshot, lock := snapshotFixture(t)
		if err := writeObject(wowapi.NewProject(root).ClientSourcesPath(), lock); err != nil {
			t.Fatal(err)
		}
		doc := []byte("local A = {}\nAPIDocumentation:AddDocumentationTable(A)\n")
		source := sourcelock.Sources{"blizzard": sourceFixture("A.lua", doc), "blizzard_version": sourceFixture("version.txt", []byte("12.1.5.70575\n"))}
		if err := writeObject(wowapi.NewProject(root).SourcesPath(), source); err != nil {
			t.Fatal(err)
		}
		verified := mustVerify(t, snapshot)
		mutate(t, snapshot)
		if _, err := projectSources(root, verified); err == nil || !strings.Contains(err.Error(), "changed after verification") {
			t.Fatalf("projectSources = %v; want a changed-after-verification rejection", err)
		}
		if exists(filepath.Join(root, "build/wow-api-source")) {
			t.Fatal("rejected projection was written")
		}
	})
	t.Run("stageNative", func(t *testing.T) {
		root, snapshot, _ := snapshotFixture(t)
		p := wowapi.NewProject(root)
		writeNativeSources(t, root, sourceFixture("A.lua", []byte("other\n")))
		before, err := os.ReadFile(p.SourcesPath())
		if err != nil {
			t.Fatal(err)
		}
		verified := mustVerify(t, snapshot)
		mutate(t, snapshot)
		if err := stageNative(root, verified); err == nil || !strings.Contains(err.Error(), "changed after verification") {
			t.Fatalf("stageNative = %v; want a changed-after-verification rejection", err)
		}
		if after, err := os.ReadFile(p.SourcesPath()); err != nil || !bytes.Equal(after, before) {
			t.Fatalf("source selection changed: %v", err)
		}
		if exists(p.ClientSourcesPath()) {
			t.Fatal("client lock written")
		}
	})
}
