package release

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
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
	for _, name := range append(append([]string{}, ownedOutputs...), "rockspecs/wow-tl-type-1.2.3-1.rockspec.in") {
		old, new := []byte("old "+name), []byte("new "+name)
		if name == "package.json" {
			old = []byte(`{"version":"1.2.3","state":"old"}`)
			new = []byte(`{"version":"1.2.3","state":"new"}`)
		}
		fixtureFile(t, root, name, old)
		fixtureFile(t, next, name, new)
	}
	for _, name := range []string{"types/wow/native_values.d.tl", "scripts/wow_api/tsm_reconciliation.json", "scripts/wow_api/overrides.json", "tools/check_secrets.lua", "references/wow-api/native-observations/old/report.json"} {
		fixtureFile(t, root, name, []byte("approved immutable "+name))
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
	if len(changed) != len(ownedOutputs)+1 || order[len(order)-1] != "package.json" {
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
	for _, name := range []string{"types/wow/native_values.d.tl", "scripts/wow_api/tsm_reconciliation.json", "scripts/wow_api/overrides.json", "tools/check_secrets.lua"} {
		b, _ := os.ReadFile(filepath.Join(root, name))
		if string(b) != "approved immutable "+name {
			t.Fatalf("changed frozen policy: %s", name)
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
			if !reflect.DeepEqual(before, after) {
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
	if !reflect.DeepEqual(before, after) {
		t.Fatal("recovery did not restore prior publication")
	}
}
func TestFailedCandidateRetainsDiagnosticsAndActiveVersion(t *testing.T) {
	root := t.TempDir()
	fixtureFile(t, root, ".wow-version", []byte("12.1.5.70575"))
	fixtureFile(t, root, "scripts/wow_api/client_sources.json", []byte(`{"version":"12.1.5.70575","build":70575}`))
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
