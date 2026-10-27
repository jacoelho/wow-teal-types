package wowapi

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestProbePlanMatchesFrozenIndependentPlan(t *testing.T) {
	root := captureTestRoot(t)
	plan, err := NewProject(root).ProbePlan(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	expected, err := readCaptureJSON[ProbePlan](filepath.Join(root, "internal/wowapi/testdata/probe_plan.json"))
	if err != nil {
		t.Fatal(err)
	}
	declarations := map[string]string{}
	for _, name := range []string{"wow.d.tl", "wow_env.d.tl", "wow_secure_env.d.tl", "native_values.d.tl"} {
		data, err := os.ReadFile(filepath.Join(root, "types/wow", name))
		if err != nil {
			t.Fatal(err)
		}
		declarations[name] = fmt.Sprintf("%x", sha256.Sum256(data))
	}
	encoded, err := json.Marshal(declarations)
	if err != nil {
		t.Fatal(err)
	}
	if want := fmt.Sprintf("%x", sha256.Sum256(encoded)); plan.DeclarationFingerprint != want {
		t.Fatalf("probe declaration fingerprint = %s, want %s", plan.DeclarationFingerprint, want)
	}
	// The retained plan fixes the inventory; its historical declaration bytes
	// intentionally differ from this explicitly reviewed compiler migration.
	plan.DeclarationFingerprint, expected.DeclarationFingerprint = "", ""
	if !evEqual(plan, expected) {
		actual, err := jsonPretty(plan)
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(root, "build/go-migration-probe-plan-actual.json")
		if err = os.WriteFile(path, actual, 0644); err != nil {
			t.Fatal(err)
		}
		t.Fatalf("probe plan differs from frozen Python plan: %s", path)
	}
}
func TestGeneratedProbeRuntimeAndOwnedPublication(t *testing.T) {
	root := captureTestRoot(t)
	p := NewProject(root)
	output := filepath.Join(t.TempDir(), "built")
	if _, err := p.BuildProbe(t.Context(), output, ""); err != nil {
		t.Fatal(err)
	}
	manifest, err := readCaptureJSON[ProbeBuildManifest](filepath.Join(output, "probe-build.json"))
	if err != nil {
		t.Fatal(err)
	}
	expected := map[string]bool{}
	for _, name := range probeRuntimeFiles {
		expected[name] = true
	}
	if !capEqualSet(expected, func() map[string]bool {
		m := map[string]bool{}
		for name := range manifest.Files {
			m[name] = true
		}
		return m
	}()) {
		t.Fatal("probe runtime inventory changed")
	}
	for name, expected := range manifest.Files {
		data, err := os.ReadFile(filepath.Join(output, name))
		if err != nil {
			t.Fatal(err)
		}
		if sha256hex(data) != expected {
			t.Fatalf("probe manifest does not bind %s", name)
		}
		if strings.HasSuffix(name, ".lua") {
			for _, module := range []string{"wow", "probe_types", "compat53"} {
				if strings.Contains(string(data), "require(\""+module+"\")") {
					t.Fatalf("runtime type dependency in %s", name)
				}
			}
		}
	}
	for driver, modes := range map[string][]string{
		"probe_presence_driver.lua":   {"normal", "combat", "report_error", "absent", "constructor_error", "constructor_malformed", "secret", "inaccessible", "table_inaccessible", "query_secret", "query_malformed", "query_absent", "unexpected_query", "wrong_build", "cleanup_error", "frame_error", "caller_scoped", "caller_denied", "caller_lookup_error"},
		"probe_invocation_driver.lua": {"normal", "combat", "unknown_context", "wrong_build", "wrong_interface", "wrong_project", "wrong_getter", "wrong_font", "tuple_short", "tuple_swap", "secret_return", "inaccessible_return", "native_error", "svg_unavailable", "svg_error", "cleanup_error", "diag_noop", "diag_zero_interval", "diag_changing", "diag_sentinel_error", "diag_sentinel_noop", "diag_setup_error", "diag_setter_tuple", "diag_fact_wrong_identity", "diag_fact_wrong_number", "diag_fact_secret", "diag_fact_inaccessible", "diag_fact_wrong_type", "diag_read_call_error", "diag_read_zero", "diag_read_multi", "diag_read_nil", "diag_read_secret", "diag_read_inaccessible", "diag_read_wrong_type", "diag_read_oversized", "diag_read_empty", "diag_read_nonempty", "clock_split", "sample_bytes", "sample_empty", "sample_128", "sample_129", "experiment_error", "default_time_secret", "native_plural", "projection_noop", "projection_noop_stale", "projection_error", "update_tuple", "update_error", "pending_abort"},
	} {
		for _, mode := range modes {
			t.Run(driver+"/"+mode, func(t *testing.T) {
				command := exec.CommandContext(t.Context(), "luajit", filepath.Join(root, "internal/wowapi/testdata", driver), mode, filepath.Join(output, "WoWAPIProbe"))
				data, err := command.CombinedOutput()
				if err != nil {
					t.Fatalf("offline probe behavior: %v\n%s", err, data)
				}
			})
		}
	}
	before := captureProbeFiles(t, output)
	if _, err = p.BuildProbe(t.Context(), output, ""); err != nil {
		t.Fatal(err)
	}
	after := captureProbeFiles(t, output)
	if !maps.Equal(before, after) {
		t.Fatal("repeated probe build changed owned output")
	}
	path := filepath.Join(output, "WoWAPIProbe/Runtime.lua")
	write := []byte("-- preserve local changes\n")
	if err = os.WriteFile(path, write, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err = p.BuildProbe(t.Context(), output, ""); err == nil {
		t.Fatal("probe replaced edited output")
	}
	data, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(data, write) {
		t.Fatal("probe lost local changes")
	}
}
func captureProbeFiles(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if entry.IsDir() {
			return nil
		}
		relative, e := filepath.Rel(root, path)
		if e != nil {
			return e
		}
		data, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		out[filepath.ToSlash(relative)] = string(data)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}
func TestPreviousProbeOwnershipLayoutIsCheckedBeforeUpgrade(t *testing.T) {
	p := NewProject(captureTestRoot(t))
	for _, mutation := range []string{"", "edited", "extra"} {
		t.Run(mutation, func(t *testing.T) {
			output := filepath.Join(t.TempDir(), "previous")
			files := map[string]string{}
			for _, name := range probePreviousRuntimeFiles {
				data := []byte("-- previous owned runtime\n")
				path := filepath.Join(output, name)
				if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, data, 0644); err != nil {
					t.Fatal(err)
				}
				files[name] = sha256hex(data)
			}
			encoded, err := jsonPretty(struct {
				Files map[string]string `json:"files"`
				Probe string            `json:"probe"`
			}{files, "1.0.0"})
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(filepath.Join(output, "probe-build.json"), encoded, 0644); err != nil {
				t.Fatal(err)
			}
			if mutation != "" {
				name := "notes.txt"
				if mutation == "edited" {
					name = "WoWAPIProbe/Runtime.lua"
				}
				path := filepath.Join(output, name)
				if err = os.WriteFile(path, []byte("preserve me"), 0644); err != nil {
					t.Fatal(err)
				}
				if _, err = p.BuildProbe(t.Context(), output, ""); err == nil {
					t.Fatal("probe upgraded unowned or changed output")
				}
				data, err := os.ReadFile(path)
				if err != nil || string(data) != "preserve me" {
					t.Fatal("probe upgrade lost user data")
				}
			} else {
				if _, err = p.BuildProbe(t.Context(), output, ""); err != nil {
					t.Fatal(err)
				}
				if err = checkExistingProbe(output); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
func TestProbeCompileFailurePreservesPreviousOutput(t *testing.T) {
	root := captureTestRoot(t)
	p := NewProject(root)
	output := filepath.Join(t.TempDir(), "probe")
	if _, err := p.BuildProbe(t.Context(), output, ""); err != nil {
		t.Fatal(err)
	}
	before := captureProbeFiles(t, output)
	compiler := filepath.Join(t.TempDir(), "failing-compiler")
	body := "#!/bin/sh\nif [ \"$1\" = \"--version\" ]; then printf '0.24.8-wow.1\\n'; exit 0; fi\nprintf 'synthetic compiler failure\\n' >&2\nexit 1\n"
	if err := os.WriteFile(compiler, []byte(body), 0755); err != nil {
		t.Fatal(err)
	}
	if _, err := p.BuildProbe(t.Context(), output, compiler); err == nil || !strings.Contains(err.Error(), "synthetic compiler failure") {
		t.Fatalf("expected failed strict compilation: %v", err)
	}
	if !evEqual(before, captureProbeFiles(t, output)) {
		t.Fatal("failed probe compilation changed owned output")
	}
}

func TestProbeDefaultOutputPreservesUnownedDirectory(t *testing.T) {
	root := t.TempDir()
	output := filepath.Join(root, "build/generated/wow-api-probe")
	if err := os.MkdirAll(output, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(output, "notes.txt")
	if err := os.WriteFile(path, []byte("preserve me"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := NewProject(root).BuildProbe(t.Context(), "", "/compiler-must-not-be-invoked"); err == nil || !strings.Contains(err.Error(), output) {
		t.Fatalf("expected default output ownership failure: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "preserve me" {
		t.Fatal("default probe output lost unowned data")
	}
}

func TestProbePublicationFailurePreservesPreviousOutput(t *testing.T) {
	for _, restoreFails := range []bool{false, true} {
		name := "restored"
		if restoreFails {
			name = "recovery retained"
		}
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			output := filepath.Join(root, "output")
			stage := filepath.Join(root, "stage")
			publication := filepath.Join(stage, "publication")
			backup := filepath.Join(stage, "previous")
			if err := os.MkdirAll(publication, 0755); err != nil {
				t.Fatal(err)
			}
			files := map[string]string{}
			previous := []byte("-- previous owned probe bytes\n")
			for _, name := range probePreviousRuntimeFiles {
				path := filepath.Join(output, name)
				if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, previous, 0644); err != nil {
					t.Fatal(err)
				}
				files[name] = sha256hex(previous)
			}
			manifest, err := jsonPretty(struct {
				Files map[string]string `json:"files"`
				Probe string            `json:"probe"`
			}{files, "1.0.0"})
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(filepath.Join(output, "probe-build.json"), manifest, 0644); err != nil {
				t.Fatal(err)
			}
			installErr := errors.New("injected install failure")
			restoreErr := errors.New("injected restore failure")
			err = publishProbe(stage, output, func(from, to string) error {
				if from == publication {
					return installErr
				}
				if from == backup && restoreFails {
					return restoreErr
				}
				return os.Rename(from, to)
			})
			if !errors.Is(err, installErr) {
				t.Fatalf("publication error = %v", err)
			}
			preserved := output
			if restoreFails {
				if !errors.Is(err, restoreErr) || !strings.Contains(err.Error(), backup) {
					t.Fatalf("rollback error does not identify recovery: %v", err)
				}
				preserved = backup
			} else if _, err := os.Stat(stage); !os.IsNotExist(err) {
				t.Fatalf("restored publication left staging files: %v", err)
			}
			for _, name := range probePreviousRuntimeFiles {
				data, err := os.ReadFile(filepath.Join(preserved, name))
				if err != nil || !bytes.Equal(data, previous) {
					t.Fatalf("previous %s was lost: %v", name, err)
				}
			}
			data, err := os.ReadFile(filepath.Join(preserved, "probe-build.json"))
			if err != nil || !bytes.Equal(data, manifest) {
				t.Fatalf("previous ownership manifest was lost: %v", err)
			}
		})
	}
}
