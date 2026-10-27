package release

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func fixtureFile(t *testing.T, root, name string, b []byte) {
	t.Helper()
	if e := atomicWrite(filepath.Join(root, filepath.FromSlash(name)), b, 0644); e != nil {
		t.Fatal(e)
	}
}
func publicationFixture(t *testing.T) (string, string, string) {
	t.Helper()
	base := t.TempDir()
	root, next := filepath.Join(base, "checkout"), filepath.Join(base, "candidate")
	if e := os.MkdirAll(root, 0755); e != nil {
		t.Fatal(e)
	}
	if e := os.MkdirAll(next, 0755); e != nil {
		t.Fatal(e)
	}
	fixtureFile(t, root, ".wow-version", []byte("12.1.5.70575"))
	for _, name := range []string{"types/wow/wow.d.tl", "types/wow/wow_env.d.tl", "types/wow/wow_secure_env.d.tl", "types/wow/call_policy.lua", "internal/wowapi/inputs/sources.json", "internal/wowapi/inputs/client_sources.json", "README.md", "examples/minimal/WowTealExample.toc", "package.json", "rockspecs/wow-tl-type-1.2.3-1.rockspec.in"} {
		old, new := []byte("old "+name), []byte("new "+name)
		if name == "package.json" {
			old = []byte(`{"schema_version":1,"version":"1.2.3","name":"old","license":"MIT","teal_version":"0.24.8","wow_build":"12.1.5.70575","files":{}}`)
			new = []byte(`{"schema_version":1,"version":"1.2.3","name":"new","license":"MIT","teal_version":"0.24.8","wow_build":"12.1.5.70575","files":{}}`)
		}
		fixtureFile(t, root, name, old)
		fixtureFile(t, next, name, new)
	}
	for _, name := range []string{"types/wow/native_values.d.tl", "internal/wowapi/inputs/framexml.json", "internal/wowapi/inputs/overrides.json", "tools/check_secrets.lua"} {
		fixtureFile(t, root, name, []byte("maintained input "+name))
		fixtureFile(t, next, name, []byte("candidate must not publish this"))
	}
	if e := os.Chmod(filepath.Join(root, "types/wow/wow.d.tl"), 0640); e != nil {
		t.Fatal(e)
	}
	run := filepath.Join(root, "build/wow-update/trial")
	if e := os.MkdirAll(run, 0755); e != nil {
		t.Fatal(e)
	}
	return root, next, run
}
func TestPublicationOwnsOnlyDerivedOutputsAndIsRepeatable(t *testing.T) {
	root, next, run := publicationFixture(t)
	old, e := captureInputs(root)
	if e != nil {
		t.Fatal(e)
	}
	var order []string
	replace := func(a, b string) error { order = append(order, filepath.Base(b)); return os.Rename(a, b) }
	changed, e := publish(t.Context(), root, next, run, old, replace)
	if e != nil {
		t.Fatal(e)
	}
	want := []string{"types/wow/wow.d.tl", "types/wow/wow_env.d.tl", "types/wow/wow_secure_env.d.tl", "types/wow/call_policy.lua", "internal/wowapi/inputs/sources.json", "internal/wowapi/inputs/client_sources.json", "README.md", "examples/minimal/WowTealExample.toc", "package.json", "rockspecs/wow-tl-type-1.2.3-1.rockspec.in"}
	slices.Sort(want)
	got := slices.Clone(changed)
	slices.Sort(got)
	if !slices.Equal(got, want) || order[len(order)-1] != "package.json" {
		t.Fatalf("publication inventory/order: %v", changed)
	}
	for _, name := range changed {
		a, e := os.ReadFile(filepath.Join(root, name))
		if e != nil {
			t.Fatal(e)
		}
		b, e := os.ReadFile(filepath.Join(next, name))
		if e != nil {
			t.Fatal(e)
		}
		if string(a) != string(b) {
			t.Fatal(name)
		}
	}
	for _, name := range []string{"types/wow/native_values.d.tl", "internal/wowapi/inputs/framexml.json", "internal/wowapi/inputs/overrides.json", "tools/check_secrets.lua"} {
		b, _ := os.ReadFile(filepath.Join(root, name))
		if string(b) != "maintained input "+name {
			t.Fatalf("changed maintained input: %s", name)
		}
	}
	version, _ := os.ReadFile(filepath.Join(root, ".wow-version"))
	if string(version) != "12.1.5.70575" {
		t.Fatalf("version bytes changed")
	}
	st, _ := os.Stat(filepath.Join(root, "types/wow/wow.d.tl"))
	if st.Mode().Perm() != 0640 {
		t.Fatal("mode changed")
	}
	after, _ := captureInputs(root)
	changes, e := publish(t.Context(), root, next, run, after, os.Rename)
	if e != nil || len(changes) != 0 {
		t.Fatalf("repeat: %v %v", changes, e)
	}
}
func TestPublicationRollsBackAfterActualReplacementAndCancellation(t *testing.T) {
	for _, cancelled := range []bool{false, true} {
		t.Run(map[bool]string{false: "disk_failure", true: "interrupt"}[cancelled], func(t *testing.T) {
			root, next, run := publicationFixture(t)
			newName := "rockspecs/wow-tl-type-1.2.3-1.rockspec.in"
			if e := os.Remove(filepath.Join(root, newName)); e != nil {
				t.Fatal(e)
			}
			before, _ := captureInputs(root)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			count := 0
			replace := func(a, b string) error {
				if e := os.Rename(a, b); e != nil {
					return e
				}
				count++
				if count == 3 {
					if cancelled {
						cancel()
					} else {
						return errors.New("disk full after rename")
					}
				}
				return nil
			}
			_, e := publish(ctx, root, next, run, before, replace)
			if e == nil {
				t.Fatal("failure accepted")
			}
			after, err := captureInputs(root)
			if err != nil {
				t.Fatal(err)
			}
			if !maps.Equal(before, after) {
				t.Fatal("rollback lost original bytes, absence or modes")
			}
			if exists(filepath.Join(root, newName)) {
				t.Fatal("new output survived rollback")
			}
			if exists(filepath.Join(root, "build/wow-update/pending.json")) {
				t.Fatal("clean rollback retained pending journal")
			}
		})
	}
}
func TestPublicationRefusesStaleVersionOrSource(t *testing.T) {
	for _, name := range []string{".wow-version", "internal/wowapi/new.go"} {
		t.Run(name, func(t *testing.T) {
			root, next, run := publicationFixture(t)
			before, _ := captureInputs(root)
			fixtureFile(t, root, name, []byte("external edit"))
			_, e := publish(t.Context(), root, next, run, before, os.Rename)
			if e == nil || !strings.Contains(e.Error(), "active inputs changed") {
				t.Fatalf("stale accepted: %v", e)
			}
			b, _ := os.ReadFile(filepath.Join(root, "types/wow/wow.d.tl"))
			if string(b) != "old types/wow/wow.d.tl" {
				t.Fatal("published stale candidate")
			}
		})
	}
}
func TestConcurrentUpdateLockFailsWithoutWaiting(t *testing.T) {
	root := t.TempDir()
	unlock, e := updateLock(root)
	if e != nil {
		t.Fatal(e)
	}
	defer unlock()
	second, e := updateLock(root)
	if e == nil {
		second()
		t.Fatal("concurrent update accepted")
	}
}
func TestRecoveryPreservesExternalEditAndRestoresAfterItIsResolved(t *testing.T) {
	root, next, run := publicationFixture(t)
	before, _ := captureInputs(root)
	var edited string
	replace := func(a, b string) error {
		if e := os.Rename(a, b); e != nil {
			return e
		}
		edited = b
		fixtureFile(t, root, strings.TrimPrefix(b, root+string(filepath.Separator)), []byte("external edit"))
		return errors.New("interrupted publication")
	}
	_, e := publish(t.Context(), root, next, run, before, replace)
	if e == nil || !strings.Contains(e.Error(), "requires recovery") {
		t.Fatalf("external edit lost: %v", e)
	}
	b, _ := os.ReadFile(edited)
	if string(b) != "external edit" {
		t.Fatal("rollback overwrote external edit")
	}
	if _, e = Recover(root); e == nil {
		t.Fatal("recovery overwrote unresolved external edit")
	}
	relative, _ := filepath.Rel(root, edited)
	checked, _ := os.ReadFile(filepath.Join(next, relative))
	if e = atomicWrite(edited, checked, 0644); e != nil {
		t.Fatal(e)
	}
	if _, e = Recover(root); e != nil {
		t.Fatal(e)
	}
	after, _ := captureInputs(root)
	if !maps.Equal(before, after) {
		t.Fatal("recovery did not restore prior publication")
	}
}
func TestFailedCandidateRetainsDiagnosticsAndActiveVersion(t *testing.T) {
	root := t.TempDir()
	fixtureFile(t, root, ".wow-version", []byte("12.1.5.70575"))
	fixtureFile(t, root, "internal/wowapi/inputs/client_sources.json", []byte(`{"version":"12.1.5.70575","build":70575}`))
	fixtureFile(t, root, "types/wow/wow.d.tl", []byte("usable previous declarations"))
	_, e := Update(t.Context(), root)
	if e == nil {
		t.Fatal("incomplete candidate accepted")
	}
	b, _ := os.ReadFile(filepath.Join(root, "types/wow/wow.d.tl"))
	if string(b) != "usable previous declarations" {
		t.Fatal("failed candidate changed declarations")
	}
	matches, _ := filepath.Glob(filepath.Join(root, "build/wow-update/*/failure.json"))
	if len(matches) != 1 {
		t.Fatalf("missing candidate diagnostics: %v", matches)
	}
	version, _ := os.ReadFile(filepath.Join(root, ".wow-version"))
	if string(version) != "12.1.5.70575" {
		t.Fatal("selected version rewritten")
	}
}

func TestMalformedRecoveryRetainsJournalAndOutputs(t *testing.T) {
	for _, kind := range []string{"null", "array", "trailing", "missing_after", "null_after", "fractional_mode", "null_mode", "overflow_mode"} {
		t.Run(kind, func(t *testing.T) {
			root, _, run := publicationFixture(t)
			id, err := identityOf(filepath.Join(root, "types/wow/wow.d.tl"))
			if err != nil {
				t.Fatal(err)
			}
			body, err := jsonBytes(journal{Schema: 1, Root: root, Files: []publicationRow{{Path: "types/wow/wow.d.tl", Before: id, After: id, Backup: "0"}}})
			if err != nil {
				t.Fatal(err)
			}
			switch kind {
			case "null":
				body = []byte("null")
			case "array":
				body = []byte("[]")
			case "trailing":
				body = append(body, []byte("{}")...)
			case "missing_after":
				body = []byte(fmt.Sprintf(`{"schema":1,"root":%q,"files":[{"path":"types/wow/wow.d.tl","before":null,"backup":"0"}]}`, root))
			case "null_after":
				body, err = jsonBytes(journal{Schema: 1, Root: root, Files: []publicationRow{{Path: "types/wow/wow.d.tl", Before: id, After: nil, Backup: "0"}}})
				if err != nil {
					t.Fatal(err)
				}
			case "fractional_mode":
				body = []byte(strings.Replace(string(body), `"mode": 416`, `"mode": 416.5`, 1))
			case "null_mode":
				body = []byte(strings.Replace(string(body), `"mode": 416`, `"mode": null`, 1))
			case "overflow_mode":
				body = []byte(strings.Replace(string(body), `"mode": 416`, `"mode": 4294967296`, 1))
			}
			journalPath := filepath.Join(run, "publication/journal.json")
			fixtureFile(t, filepath.Dir(journalPath), "journal.json", body)
			pointer := filepath.Join(root, "build/wow-update/pending.json")
			if err = writeObject(pointer, publicationPointer{Journal: journalPath, SHA256: sha(body)}); err != nil {
				t.Fatal(err)
			}
			before := metadataInputs(t, root)
			if _, err = Recover(root); err == nil {
				t.Fatal("malformed recovery accepted")
			}
			if !maps.Equal(before, metadataInputs(t, root)) {
				t.Fatal("malformed recovery changed outputs")
			}
			actual, err := os.ReadFile(journalPath)
			if err != nil || string(actual) != string(body) {
				t.Fatal("malformed journal changed")
			}
			if _, err = os.Stat(pointer); err != nil {
				t.Fatal("pending marker removed")
			}
		})
	}
}
