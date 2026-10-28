package consumer

import (
	"errors"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildUsesCapturedInputsAndManifest(t *testing.T) {
	f := newAddon(t)
	source := f.write(t, "Main.tl", "local n: number = MyCounter\nprint(n)\nprint(\"captured\")\n")
	declarations := filepath.Join(f.project, "types")
	declaration := filepath.Join(declarations, "addon_env.d.tl")
	writeTest(t, declaration, "global MyCounter: number\n")
	manifest := filepath.Join(f.root, "package.json")
	before := fileText(t, manifest)
	actualTL := teal(t)
	wrapper := filepath.Join(f.project, "tl-wrapper")
	writeTest(t, wrapper, `#!/bin/sh
"$SNAPSHOT_REAL_TL" "$@" || exit $?
for arg do
 case "$arg" in
 */tools/check_secrets.lua)
  printf 'local invalid: number = "changed"\n' > "$SNAPSHOT_SOURCE"
  printf 'global MyCounter: string\n' > "$SNAPSHOT_DECLARATION"
  printf '\n' >> "$SNAPSHOT_MANIFEST"
  ;;
 esac
done
`)
	if err := os.Chmod(wrapper, 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SNAPSHOT_REAL_TL", actualTL)
	t.Setenv("SNAPSHOT_SOURCE", source)
	t.Setenv("SNAPSHOT_DECLARATION", declaration)
	t.Setenv("SNAPSHOT_MANIFEST", manifest)
	t.Setenv("TL", wrapper)
	f.build(t, Selection{Sources: []string{f.source}, Include: []string{declarations}, Global: "addon_env"}, f.output, "")
	if got := fileText(t, filepath.Join(f.output, "Main.lua")); !strings.Contains(got, `print("captured")`) {
		t.Fatalf("generation did not use checked source: %s", got)
	}
	receipt, err := readBuildReceipt(filepath.Join(f.output, "wow-teal-build.json"))
	if err != nil {
		t.Fatal(err)
	}
	if receipt.PackageManifestSHA256 != fingerprint([]byte(before)) {
		t.Fatal("receipt did not bind captured manifest")
	}
	t.Setenv("TL", actualTL)
	previous := outputBytes(t, f.output)
	f.build(t, Selection{Sources: []string{f.source}, Include: []string{declarations}, Global: "addon_env"}, f.output, "expected number")
	if !maps.Equal(previous, outputBytes(t, f.output)) {
		t.Fatal("failed later invocation changed output")
	}
}

func TestBuildRejectsChangedPackageBeforeCompiler(t *testing.T) {
	f := newAddon(t)
	f.write(t, "Main.tl", "print(\"checked\")\n")
	f.build(t, Selection{Sources: []string{f.source}}, f.output, "")
	before := outputBytes(t, f.output)
	writeTest(t, filepath.Join(f.root, "tools/check_secrets.lua"), "error('modified checker')\n")
	t.Setenv("TL", "/must-not-run-compiler")
	f.build(t, Selection{Sources: []string{f.source}}, f.output, "package integrity mismatch: tools/check_secrets.lua")
	if !maps.Equal(before, outputBytes(t, f.output)) {
		t.Fatal("invalid package changed output")
	}
}

func TestInputLinksRejectButAncestorAliasesResolve(t *testing.T) {
	f := newAddon(t)
	f.write(t, "Main.tl", "print(\"checked\")\n")
	alias := filepath.Join(f.project, "alias")
	if err := os.Symlink(f.source, alias); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{alias, alias + string(filepath.Separator)} {
		f.check(t, Selection{Sources: []string{path}}, "must not be a symlink")
		f.check(t, Selection{Sources: []string{f.source}, Include: []string{path}}, "must not be a symlink")
	}
	parentAlias := filepath.Join(t.TempDir(), "project")
	if err := os.Symlink(f.project, parentAlias); err != nil {
		t.Fatal(err)
	}
	f.check(t, Selection{Sources: []string{filepath.Join(parentAlias, "src")}}, "")
	nested := filepath.Join(f.source, "linked")
	if err := os.Symlink(filepath.Join(f.project, "absent"), nested); err != nil {
		t.Fatal(err)
	}
	f.check(t, Selection{Sources: []string{f.source}}, "must not be a symlink")
	if err := os.Remove(nested); err != nil {
		t.Fatal(err)
	}
	deps := filepath.Join(f.project, "types")
	writeTest(t, filepath.Join(deps, "real.d.tl"), "return {}\n")
	if err := os.Symlink(filepath.Join(deps, "real.d.tl"), filepath.Join(deps, "linked.d.tl")); err != nil {
		t.Fatal(err)
	}
	f.check(t, Selection{Sources: []string{f.source}, Include: []string{deps}}, "must not be a symlink")
}

func TestExcludedDependenciesRequireExplicitRoot(t *testing.T) {
	f := newAddon(t)
	f.write(t, "Main.tl", "local helper = require(\"Helper\")\nlocal n: number = helper.value\nprint(n)\n")
	deps := filepath.Join(f.source, "vendor/Library")
	writeTest(t, filepath.Join(deps, "Helper.tl"), "return { value = 17 }\n")
	f.check(t, Selection{Sources: []string{f.source}}, "module not found")
	f.check(t, Selection{Sources: []string{f.source}, Include: []string{deps}}, "")
}

func TestCapturedRootsPreserveTealSuffixPrecedence(t *testing.T) {
	f := newAddon(t)
	f.write(t, "Main.tl", "local helper = require(\"Helper\")\nlocal n: number = helper.value\nprint(n)\n")
	first, second := filepath.Join(f.project, "first"), filepath.Join(f.project, "second")
	writeTest(t, filepath.Join(first, "Helper.tl"), "return { value = \"wrong\" }\n")
	writeTest(t, filepath.Join(second, "Helper.d.tl"), "local record Helper\n value: number\nend\nreturn Helper\n")
	f.check(t, Selection{Sources: []string{f.source}, Include: []string{first, second, first}}, "")
}

func TestPublicationFailureRetainsRecoverableBackup(t *testing.T) {
	for _, restoreFails := range []bool{false, true} {
		t.Run(map[bool]string{false: "restored", true: "retained"}[restoreFails], func(t *testing.T) {
			parent := t.TempDir()
			destination, generated := filepath.Join(parent, "output"), filepath.Join(parent, "generated")
			writeTest(t, filepath.Join(destination, "Main.lua"), "previous")
			receipt, err := receiptBytes(BuildReceipt{Files: map[string]string{"Main.lua": fingerprint([]byte("previous"))}})
			if err != nil {
				t.Fatal(err)
			}
			writeTest(t, filepath.Join(destination, "wow-teal-build.json"), string(receipt))
			writeTest(t, filepath.Join(generated, "Main.lua"), "next")
			before := outputBytes(t, destination)
			installErr, restoreErr := errors.New("install failed"), errors.New("restore failed")
			backup := ""
			rename := func(from, to string) error {
				if from == destination {
					backup = to
					return os.Rename(from, to)
				}
				if from == backup {
					if restoreFails {
						return restoreErr
					}
					return os.Rename(from, to)
				}
				return installErr
			}
			err = publishBuild(generated, destination, rename)
			if !errors.Is(err, installErr) {
				t.Fatalf("lost install cause: %v", err)
			}
			if restoreFails {
				if !errors.Is(err, restoreErr) || !strings.Contains(err.Error(), backup) {
					t.Fatalf("missing recovery evidence: %v", err)
				}
				if !maps.Equal(before, outputBytes(t, backup)) {
					t.Fatal("backup was lost")
				}
			} else {
				if !maps.Equal(before, outputBytes(t, destination)) {
					t.Fatal("previous output not restored")
				}
				matches, err := filepath.Glob(filepath.Join(parent, ".wow-teal-publish-*"))
				if err != nil || len(matches) != 0 {
					t.Fatalf("successful rollback leaked stage: %v %v", matches, err)
				}
			}
		})
	}
}

func TestDependencyAggregateBound(t *testing.T) {
	f := newAddon(t)
	f.write(t, "Main.tl", "print(1)\n")
	deps := filepath.Join(f.project, "types")
	for _, name := range []string{"a", "b", "c", "d", "e", "f", "g", "h", "i"} {
		writeTest(t, filepath.Join(deps, name+".d.tl"), strings.Repeat(" ", 2<<20))
	}
	t.Setenv("TL", "/must-not-run-compiler")
	f.check(t, Selection{Sources: []string{f.source}, Include: []string{deps}}, "source/dependency inputs exceed 16 MiB")
}

func TestFilesCreatedAfterCaptureCannotShadowDependencies(t *testing.T) {
	f := newAddon(t)
	f.write(t, "Main.tl", "local helper = require(\"Helper\")\nlocal n: number = helper.value\nprint(n)\n")
	deps := filepath.Join(f.project, "dependencies")
	writeTest(t, filepath.Join(deps, "Helper.tl"), "return { value = 17 }\n")
	wrapper := filepath.Join(f.project, "tl-wrapper")
	actualTL := teal(t)
	writeTest(t, wrapper, `#!/bin/sh
"$SNAPSHOT_REAL_TL" "$@" || exit $?
for arg do
 case "$arg" in
 */tools/check_secrets.lua)
  printf 'local record Helper\n value: string\nend\nreturn Helper\n' > "$SNAPSHOT_SHADOW"
  ;;
 esac
done
`)
	if err := os.Chmod(wrapper, 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SNAPSHOT_REAL_TL", actualTL)
	t.Setenv("SNAPSHOT_SHADOW", filepath.Join(deps, "Helper.d.tl"))
	t.Setenv("TL", wrapper)
	f.build(t, Selection{Sources: []string{f.source}, Include: []string{deps}}, f.output, "")
	t.Setenv("TL", actualTL)
	f.check(t, Selection{Sources: []string{f.source}, Include: []string{deps}}, "expected number")
}
