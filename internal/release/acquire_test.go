package release

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"wow-teal-types/internal/consumer"
	"wow-teal-types/internal/sourcelock"
	"wow-teal-types/internal/wowapi"
)

// snapshotFixture returns a root with a .wow-version and an extractor source,
// plus a valid snapshot at <root>/snapshot.
func snapshotFixture(t *testing.T) (string, string, sourcelock.Snapshot) {
	t.Helper()
	root := t.TempDir()
	fixtureFile(t, root, ".wow-version", []byte("12.1.5.70575\n"))
	fixtureFile(t, root, "tools/cdn/go.mod", []byte("module wow-teal-types/tools/cdn\n"))
	fixtureFile(t, root, "tools/cdn/main.go", []byte("package main\n"))
	snapshot := filepath.Join(root, "snapshot")
	return root, snapshot, writeSnapshotFixture(t, snapshot)
}

func writeSnapshotFixture(t *testing.T, snapshot string) sourcelock.Snapshot {
	t.Helper()
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
	record := sourcelock.NewSnapshot(sourcelock.ClientIdentity{Product: "wow", Region: "eu", Locale: "enUS", Build: 70575, Version: "12.1.5.70575", BuildConfig: strings.Repeat("a", 32), CDNConfig: strings.Repeat("b", 32), ProductConfig: strings.Repeat("c", 32)}, rows, manifests)
	if e := writeObject(filepath.Join(snapshot, "snapshot.json"), record); e != nil {
		t.Fatal(e)
	}
	lock, e := sourcelock.ReadSnapshot(filepath.Join(snapshot, "snapshot.json"))
	if e != nil {
		t.Fatal(e)
	}
	return lock
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
	if e := writeObject(wowapi.NewProject(root).ClientSourcesPath(), lock); e != nil {
		t.Fatal(e)
	}
	doc := []byte("local A = {}\nAPIDocumentation:AddDocumentationTable(A)\n")
	version := []byte("12.1.5.70575\n")
	source := sourcelock.Sources{"blizzard": sourceFixture("A.lua", doc), "blizzard_version": sourceFixture("version.txt", version)}
	if e := writeObject(wowapi.NewProject(root).SourcesPath(), source); e != nil {
		t.Fatal(e)
	}
	p, e := projectSources(root, mustVerify(t, snapshot))
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
	if _, e = projectSources(root, mustVerify(t, snapshot)); e != nil {
		t.Fatal(e)
	}
	selected := []byte("local A = {}\nAPIDocumentation:AddDocumentationTable(A)\n\n")
	source["blizzard"] = sourceFixture("A.lua", selected)
	if err := writeObject(wowapi.NewProject(root).SourcesPath(), source); err != nil {
		t.Fatal(err)
	}
	next, err := projectSources(root, mustVerify(t, snapshot))
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
	if _, e = projectSources(root, mustVerify(t, snapshot)); e == nil {
		t.Fatal("modified projection accepted")
	}
}
func TestBootstrapUsesVerifiedProjectionDespiteStaleOrdinaryCache(t *testing.T) {
	root, snapshot, lock := snapshotFixture(t)
	if err := writeObject(wowapi.NewProject(root).ClientSourcesPath(), lock); err != nil {
		t.Fatal(err)
	}
	doc := []byte("local A = {}\nAPIDocumentation:AddDocumentationTable(A)\n")
	sources := sourcelock.Sources{
		"blizzard":         sourceFixture("A.lua", doc),
		"blizzard_version": sourceFixture("version.txt", []byte("12.1.5.70575\n")),
	}
	if err := writeObject(wowapi.NewProject(root).SourcesPath(), sources); err != nil {
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
	_, e := clientPath("Blizzard_Test/Binding.lua", map[string]sourcelock.ClientFile{"Interface/AddOns/Blizzard_Test/One/Binding.lua": {}, "Interface/AddOns/Blizzard_Test/Two/Binding.lua": {}})
	if e == nil {
		t.Fatal("ambiguous source accepted")
	}
}

func mustVerify(t *testing.T, snapshot string) verifiedSnapshot {
	t.Helper()
	v, e := verifySnapshot(snapshot, nil)
	if e != nil {
		t.Fatal(e)
	}
	return v
}

func sourceFixture(name string, data []byte) sourcelock.SourceGroup {
	return sourcelock.SourceGroup{Repository: "fixture/source", Commit: strings.Repeat("a", 40), Files: []sourcelock.SourceFile{{Name: name, SHA: sourcelock.BlobSHA(data), Size: len(data)}}}
}

func TestParseVersionRow(t *testing.T) {
	const row = "eu|aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa|bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb||70575|12.1.5.70575|cccccccccccccccccccccccccccccccc"
	const header = "Region!STRING:0|BuildConfig!HEX:16|CDNConfig!HEX:16|KeyRing!HEX:16|BuildId!DEC:4|VersionsName!String:0|ProductConfig!HEX:16\n## seqn = 1\n"
	want := sourcelock.ClientIdentity{Product: "wow", Region: "eu", Locale: "enUS", Build: 70575, Version: "12.1.5.70575", BuildConfig: strings.Repeat("a", 32), CDNConfig: strings.Repeat("b", 32), ProductConfig: strings.Repeat("c", 32)}
	for _, tc := range []struct {
		name, body, wantErr string
		build               int
	}{
		{name: "valid", body: header + row + "\nus|x\n", build: 70575},
		{name: "any build", body: header + row + "\n", build: 0},
		{name: "non-numeric build", body: header + strings.Replace(row, "|70575|", "|abc|", 1), wantErr: "build field"},
		{name: "build with suffix", body: header + strings.Replace(row, "|70575|", "|70575abc|", 1), wantErr: "build field"},
		{name: "field count", body: header + row + "|extra", wantErr: "fields"},
		{name: "missing region", body: header + "us" + row[2:], wantErr: "region"},
		{name: "duplicate region", body: header + row + "\n" + row, wantErr: "region"},
		{name: "build mismatch", body: header + row, build: 70576, wantErr: "not currently offered"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseVersionRow([]byte(tc.body), "eu", tc.build)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("error = %v; want one naming %q", err, tc.wantErr)
				}
				return
			}
			if err != nil || got != want {
				t.Fatalf("identity = %+v, %v; want %+v", got, err, want)
			}
		})
	}
}

func TestFetchRetriesOnlyTransientFailures(t *testing.T) {
	type waitCall struct{ attempt, requestsBefore int }
	for _, tc := range []struct {
		name      string
		statuses  []int
		body      string
		limit     int64
		http      bool
		cancel    bool
		wantErr   string
		requests  int
		waitCalls []waitCall
	}{
		{name: "transient then success", statuses: []int{503, 429, 200}, body: "ok", requests: 3, waitCalls: []waitCall{{1, 1}, {2, 2}}},
		{name: "permanent status", statuses: []int{404}, wantErr: "HTTP 404", requests: 1},
		{name: "exhausted", statuses: []int{500, 502, 503}, wantErr: "HTTP 503", requests: 3, waitCalls: []waitCall{{1, 1}, {2, 2}}},
		{name: "cancelled during wait", statuses: []int{503, 200}, cancel: true, wantErr: context.Canceled.Error(), requests: 1, waitCalls: []waitCall{{1, 1}}},
		{name: "oversized body", statuses: []int{200, 200}, body: "12345", limit: 4, wantErr: "exceeds", requests: 1},
		{name: "plain http", statuses: []int{200}, http: true, wantErr: "unsupported acquisition URL", requests: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var served atomic.Int64
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				status := tc.statuses[min(int(served.Add(1))-1, len(tc.statuses)-1)]
				w.WriteHeader(status)
				if status == http.StatusOK {
					_, _ = w.Write([]byte(tc.body))
				}
			}))
			defer srv.Close()
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			var waits []waitCall
			f := fetcher{client: srv.Client(), wait: func(ctx context.Context, attempt int) error {
				waits = append(waits, waitCall{attempt, int(served.Load())})
				if tc.cancel {
					cancel()
					return ctx.Err()
				}
				return nil
			}}
			address := srv.URL
			if tc.http {
				address = "http" + strings.TrimPrefix(address, "https")
			}
			limit := tc.limit
			if limit == 0 {
				limit = 1 << 10
			}
			got, err := f.get(ctx, address, limit)
			if tc.wantErr == "" {
				if err != nil || string(got) != tc.body {
					t.Fatalf("get = %q, %v; want %q", got, err, tc.body)
				}
			} else if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("error = %v; want one containing %q", err, tc.wantErr)
			}
			if tc.cancel && !errors.Is(err, context.Canceled) {
				t.Fatalf("error = %v; want context.Canceled", err)
			}
			if requests := int(served.Load()); requests != tc.requests || !slices.Equal(waits, tc.waitCalls) {
				t.Fatalf("requests = %d, waits = %v; want %d, %v", requests, waits, tc.requests, tc.waitCalls)
			}
		})
	}
}

func TestResolveSnapshotReusesOrExtractsOneCacheEntry(t *testing.T) {
	// {sha} is the fixture root's extractor source hash.
	const provided, cached = "build/wow-client/70575", "build/wow-update/snapshots/12.1.5.70575/{sha}"
	type stageCall struct {
		output   string
		identity sourcelock.ClientIdentity
	}
	for _, tc := range []struct {
		name      string
		place     string // root-relative location of the fixture snapshot; "" leaves it unplaced
		stale     bool   // also cache a valid snapshot under another source hash of the same version
		selection string // "lock", "version" or "other"
		want      string
		wantErr   bool
		stages    int
	}{
		{name: "provided snapshot", place: provided, selection: "lock", want: provided},
		{name: "cached snapshot via lock", place: cached, selection: "lock", want: cached},
		{name: "cached snapshot via .wow-version", place: cached, selection: "version", want: cached},
		{name: "cached snapshot of another build", place: "build/wow-update/snapshots/12.1.5.70576/{sha}", selection: "other", wantErr: true},
		{name: "missing snapshot", selection: "lock", want: cached, stages: 1},
		{name: "cached snapshot under another source hash", stale: true, selection: "lock", want: cached, stages: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, fixture, lock := snapshotFixture(t)
			sha, err := extractorSourceSHA(root)
			if err != nil {
				t.Fatal(err)
			}
			expand := func(p string) string { return strings.ReplaceAll(p, "{sha}", sha) }
			if tc.place != "" {
				if err := os.MkdirAll(filepath.Dir(filepath.Join(root, expand(tc.place))), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.Rename(fixture, filepath.Join(root, expand(tc.place))); err != nil {
					t.Fatal(err)
				}
			}
			if tc.stale {
				writeSnapshotFixture(t, filepath.Join(root, "build/wow-update/snapshots/12.1.5.70575", strings.Repeat("e", 64)))
			}
			var selection clientSelection
			switch tc.selection {
			case "lock":
				selection = lockedSelection(lock)
			case "version":
				selected, err := consumer.ReadVersion(root)
				if err != nil {
					t.Fatal(err)
				}
				selection = offeredSelection(selected, fetcher{}, "eu")
			case "other":
				selection = offeredSelection(consumer.Version{Full: "12.1.5.70576", Build: 70576}, fetcher{}, "eu")
			}
			var calls []stageCall
			stage := func(_ context.Context, output string, identity sourcelock.ClientIdentity, _ *sourcelock.Snapshot) (verifiedSnapshot, error) {
				calls = append(calls, stageCall{output, identity})
				if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
					return verifiedSnapshot{}, err
				}
				if err := os.Rename(fixture, output); err != nil {
					return verifiedSnapshot{}, err
				}
				return verifySnapshot(output, &lock)
			}
			got, err := resolveSnapshot(t.Context(), root, selection, stage)
			if tc.wantErr {
				if err == nil {
					t.Fatal("mismatched snapshot accepted")
				}
			} else if err != nil || got.dir != filepath.Join(root, expand(tc.want)) || !got.record.Equal(lock) {
				t.Fatalf("resolveSnapshot = %q, %v; want %q", got.dir, err, filepath.Join(root, expand(tc.want)))
			}
			if len(calls) != tc.stages {
				t.Fatalf("stage calls = %v; want %d", calls, tc.stages)
			}
			if tc.stages == 1 {
				if want := (stageCall{filepath.Join(root, expand(cached)), lock.Identity()}); calls[0] != want {
					t.Fatalf("stage call = %+v; want %+v", calls[0], want)
				}
				if _, err := VerifySnapshot(filepath.Join(root, expand(cached)), &lock); err != nil {
					t.Fatalf("extracted snapshot not in cache: %v", err)
				}
			}
		})
	}
}

// Locked and offered selections both key the cache by the current extractor
// source, so an extractor change re-extracts instead of reusing old output.
func TestResolveSnapshotCacheFollowsExtractorSource(t *testing.T) {
	for _, locked := range []bool{false, true} {
		t.Run(map[bool]string{false: "offered", true: "locked"}[locked], func(t *testing.T) {
			root, fixture, lock := snapshotFixture(t)
			identities := 0
			selection := clientSelection{want: consumer.Version{Full: "12.1.5.70575", Build: 70575}, identity: func(context.Context) (sourcelock.ClientIdentity, error) {
				identities++
				return lock.Identity(), nil
			}}
			if locked {
				selection.lock = &lock
			}
			var outputs []string
			stage := func(_ context.Context, output string, identity sourcelock.ClientIdentity, _ *sourcelock.Snapshot) (verifiedSnapshot, error) {
				outputs = append(outputs, output)
				writeSnapshotFixture(t, output)
				return verifySnapshot(output, nil)
			}
			sha, err := extractorSourceSHA(root)
			if err != nil {
				t.Fatal(err)
			}
			first := filepath.Join(root, "build/wow-update/snapshots/12.1.5.70575", sha)
			if err := os.MkdirAll(filepath.Dir(first), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.Rename(fixture, first); err != nil {
				t.Fatal(err)
			}
			got, err := resolveSnapshot(t.Context(), root, selection, stage)
			if err != nil || got.dir != first || identities != 0 || len(outputs) != 0 {
				t.Fatalf("unchanged extractor: resolveSnapshot = %q, %v, identities %d, stages %v; want cache hit %q", got.dir, err, identities, outputs, first)
			}

			fixtureFile(t, root, "tools/cdn/main.go", []byte("package main\n\nfunc main() {}\n"))
			got, err = resolveSnapshot(t.Context(), root, selection, stage)
			if err != nil {
				t.Fatal(err)
			}
			if identities != 1 || len(outputs) != 1 || got.dir != outputs[0] {
				t.Fatalf("changed extractor: resolveSnapshot = %q, identities %d, stages %v; want one extraction", got.dir, identities, outputs)
			}
			if got.dir == first || filepath.Dir(got.dir) != filepath.Dir(first) {
				t.Fatalf("changed extractor resolved %q; want a new entry beside %q", got.dir, first)
			}
			if _, err := VerifySnapshot(first, &lock); err != nil {
				t.Fatalf("previous cache entry disturbed: %v", err)
			}
		})
	}
}
