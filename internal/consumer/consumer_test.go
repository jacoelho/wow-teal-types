package consumer

import (
	"archive/zip"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"testing"
)

type addonFixture struct{ root, project, source, output string }

func newAddon(t *testing.T) addonFixture {
	t.Helper()
	repo := repositoryRoot(t)
	project := filepath.Join(t.TempDir(), "external addon with spaces")
	root := filepath.Join(project, "vendor/wow-teal-types")
	source := filepath.Join(project, "src")
	manifest, err := readJSON(filepath.Join(repo, "package.json"))
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]any{}
	for _, name := range RequiredFiles() {
		data, err := os.ReadFile(filepath.Join(repo, filepath.FromSlash(name)))
		if err != nil {
			t.Fatal(err)
		}
		writeTest(t, filepath.Join(root, filepath.FromSlash(name)), string(data))
		files[name] = fingerprint(data)
	}
	manifest["files"] = files
	data, err := jsonBytes(manifest)
	if err != nil {
		t.Fatal(err)
	}
	writeTest(t, filepath.Join(root, "package.json"), string(data))
	if err = os.MkdirAll(source, 0755); err != nil {
		t.Fatal(err)
	}
	return addonFixture{root, project, source, filepath.Join(project, "build/lua")}
}
func (f addonFixture) write(t *testing.T, name, text string) string {
	t.Helper()
	path := filepath.Join(f.source, name)
	writeTest(t, path, text)
	return path
}
func (f addonFixture) run(t *testing.T, args []string, reject string) string {
	t.Helper()
	return runConsumer(t, t.Context(), f.root, args, reject)
}
func fileText(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
func outputBytes(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if !entry.IsDir() {
			relative, e := filepath.Rel(root, path)
			if e != nil {
				return e
			}
			data, e := os.ReadFile(path)
			if e != nil {
				return e
			}
			out[relative] = string(data)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestVersionCommandUsesVerifiedPackageWithoutCompiler(t *testing.T) {
	f := newAddon(t)
	t.Setenv("TL", "/compiler-must-not-be-invoked")
	if got := f.run(t, []string{"version"}, ""); got != "0.1.0\n" {
		t.Fatalf("version output: %q", got)
	}
	f.run(t, []string{"version", "extra"}, "version takes no arguments")
	writeTest(t, filepath.Join(f.root, "types/wow/wow.d.tl"), "changed declaration\n")
	f.run(t, []string{"version"}, "package integrity mismatch")
}

func TestRelocatedAddonBuildAndRepeat(t *testing.T) {
	f := newAddon(t)
	f.write(t, "WowTealExample.tl", fileText(t, filepath.Join(repositoryRoot(t), "examples/minimal/WowTealExample.tl")))
	writeTest(t, filepath.Join(f.project, "tlconfig.lua"), "error(\"ambient config must not run\")\n")
	t.Setenv("TL_PATH", "/missing-and-untrusted/?.lua")
	text := f.run(t, []string{"build", f.source, "-o", f.output}, "")
	if !strings.Contains(text, "Passed type and secret checks for 1 selected source files.") {
		t.Fatal(text)
	}
	result := fileText(t, filepath.Join(f.output, "WowTealExample.lua"))
	for _, want := range []string{"CreateFrame(\"Frame\"", "bar:SetValue(health)"} {
		if !strings.Contains(result, want) {
			t.Fatalf("generated addon missing %q", want)
		}
	}
	if strings.Contains(result, "require(") {
		t.Fatal("type import survived generation")
	}
	before := outputBytes(t, f.output)
	f.run(t, []string{"build", f.source, "-o", f.output}, "")
	if !reflect.DeepEqual(before, outputBytes(t, f.output)) {
		t.Fatal("repeat changed checked build")
	}
}
func TestSelectedImportedBodiesReceiveSecretChecks(t *testing.T) {
	f := newAddon(t)
	f.write(t, "Entry.tl", "local helpers = require(\"Helpers\")\nhelpers.inspect(UnitHealth(\"player\"))\n")
	f.write(t, "Helpers.tl", `local type W = require("wow")
local record Helpers
   inspect: function(W.NumberLike)
end
local helpers: Helpers = {}
function helpers.inspect(value: W.NumberLike)
   if value == 0 then print("bad") end
end
return helpers
`)
	f.run(t, []string{"check", f.source}, "secret-value comparison")
}
func TestInvalidSourceSelectionAndTypeErrors(t *testing.T) {
	f := newAddon(t)
	f.write(t, "Bad.tl", "C_Spell.GetBaseSpell(true)\n")
	f.run(t, []string{"check", f.source}, "argument")
	f.run(t, []string{"check", filepath.Join(f.source, "Missing.tl")}, "no such file")
	empty := filepath.Join(f.project, "empty")
	if err := os.Mkdir(empty, 0755); err != nil {
		t.Fatal(err)
	}
	f.run(t, []string{"check", empty}, "no Teal source files")
	f.run(t, []string{"check", filepath.Join(f.root, "types/wow/wow.d.tl")}, "expected a Teal source")
}
func TestManifestBindsTypesCheckersAndVersion(t *testing.T) {
	f := newAddon(t)
	for _, name := range []string{"types/wow/wow.d.tl", "tools/check_secrets.lua", ".wow-version", "internal/consumer/consumer.go"} {
		path := filepath.Join(f.root, name)
		before := fileText(t, path)
		writeTest(t, path, before+"\n-- changed\n")
		f.run(t, []string{"verify"}, "package integrity mismatch")
		writeTest(t, path, before)
	}
	f.run(t, []string{"verify"}, "")
}
func TestPackageResolutionAndDirectShadowFailure(t *testing.T) {
	f := newAddon(t)
	source := f.write(t, "Unsafe.tl", "local type W = require(\"wow\")\nlocal function f(v: W.NumberLike) if v then print(v) end end\n")
	shadow := filepath.Join(f.project, "shadow")
	writeTest(t, filepath.Join(shadow, "wow.d.tl"), "local record Fake\n type NumberLike = number\nend\nreturn Fake\n")
	f.run(t, []string{"check", f.source, "-I", shadow}, "secret-value truthiness")
	command := exec.CommandContext(t.Context(), teal(t), "run", filepath.Join(f.root, "tools/check_secrets.lua"), "--", "wow_env", source)
	command.Dir = f.project
	command.Env = append(os.Environ(), "TL_PATH="+shadow+"/?.lua;"+filepath.Join(f.root, "types/wow/?.lua"))
	data, err := command.CombinedOutput()
	if err == nil || !strings.Contains(string(data), "declaration identity mismatch: wow") {
		t.Fatalf("shadow was not rejected: %v\n%s", err, data)
	}
}
func TestAddonGlobalsExtendNativeEnvironment(t *testing.T) {
	f := newAddon(t)
	declarations := filepath.Join(f.project, "addon_types")
	writeTest(t, filepath.Join(declarations, "addon_env.d.tl"), "global MyAddonCounter: number\n")
	f.write(t, "Main.tl", "MyAddonCounter = C_Spell.GetBaseSpell(17)\n")
	f.run(t, []string{"build", f.source, "-I", declarations, "--global-env-def", "addon_env", "-o", f.output}, "")
	if strings.Contains(fileText(t, filepath.Join(f.output, "Main.lua")), "require(") {
		t.Fatal("declaration import survived")
	}
	f.run(t, []string{"check", f.source}, "MyAddonCounter")
}
func TestFailedBuildPreservesOutputAndRejectsForeignDirectory(t *testing.T) {
	f := newAddon(t)
	f.write(t, "Good.tl", "print(\"checked\")\n")
	f.run(t, []string{"build", f.source, "-o", f.output}, "")
	before := outputBytes(t, f.output)
	f.write(t, "Bad.tl", "local type W = require(\"wow\")\nlocal function f(v: W.NumberLike) if v == 0 then print(v) end end\n")
	f.run(t, []string{"build", f.source, "-o", f.output}, "secret-value comparison")
	if !reflect.DeepEqual(before, outputBytes(t, f.output)) {
		t.Fatal("failed build changed previous output")
	}
	f.run(t, []string{"build", f.source, "-o", f.source}, "output must not overlap")
	foreign := filepath.Join(f.project, "foreign")
	writeTest(t, filepath.Join(foreign, "mine.txt"), "preserve")
	f.run(t, []string{"build", f.source, "-o", foreign}, "no wow-teal ownership receipt")
	if fileText(t, filepath.Join(foreign, "mine.txt")) != "preserve" {
		t.Fatal("foreign output changed")
	}
}
func TestWrongCompilerAndEditedOutputDoNotPublish(t *testing.T) {
	f := newAddon(t)
	f.write(t, "Good.tl", "print(\"checked\")\n")
	f.run(t, []string{"build", f.source, "-o", f.output}, "")
	before := fileText(t, filepath.Join(f.output, "Good.lua"))
	compiler := filepath.Join(f.project, "future-tl")
	writeTest(t, compiler, "#!/bin/sh\nprintf '99.0.0\\n'\n")
	if err := os.Chmod(compiler, 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TL", compiler)
	f.run(t, []string{"build", f.source, "-o", f.output}, "unsupported Teal version")
	if fileText(t, filepath.Join(f.output, "Good.lua")) != before {
		t.Fatal("wrong compiler changed output")
	}
	writeTest(t, filepath.Join(f.output, "Good.lua"), "-- user edit\n")
	f.run(t, []string{"build", f.source, "-o", f.output}, "output has local changes")
	if fileText(t, filepath.Join(f.output, "Good.lua")) != "-- user edit\n" {
		t.Fatal("user edit replaced")
	}
}
func TestSymlinkAndExcludedSourceTrees(t *testing.T) {
	f := newAddon(t)
	source := f.write(t, "Good.tl", "print(\"checked\")\n")
	f.write(t, "build/Bad.tl", "not valid Teal")
	f.write(t, "vendor/Bad.tl", "not valid Teal")
	f.run(t, []string{"check", f.source}, "")
	if err := os.Symlink(source, filepath.Join(f.source, "Linked.tl")); err != nil {
		t.Fatal(err)
	}
	f.run(t, []string{"check", f.source}, "must not be a symlink")
}
func TestSecureEnvironmentSelection(t *testing.T) {
	f := newAddon(t)
	f.write(t, "Main.tl", "local forbidden = GetForbiddenObjectTable(CreateFrame(\"Frame\"))\nprint(forbidden)\n")
	f.run(t, []string{"check", f.source}, "GetForbiddenObjectTable")
	f.run(t, []string{"check", f.source, "--secure"}, "")
}
func TestNestedModulesResolveFromSelectedRoot(t *testing.T) {
	f := newAddon(t)
	f.write(t, "myaddon/Entry.tl", "local helper = require(\"myaddon.Helpers\")\nprint(helper.value)\n")
	f.write(t, "myaddon/Helpers.tl", "local helper = { value = 17 }\nreturn helper\n")
	f.run(t, []string{"build", f.source, "-o", f.output}, "")
	for _, name := range []string{"myaddon/Entry.lua", "myaddon/Helpers.lua"} {
		if _, err := os.Stat(filepath.Join(f.output, name)); err != nil {
			t.Fatal(err)
		}
	}
}
func TestNonregularSourcesAndDependenciesRejectBeforeReading(t *testing.T) {
	f := newAddon(t)
	f.write(t, "Good.tl", "print(\"checked\")\n")
	f.run(t, []string{"build", f.source, "-o", f.output}, "")
	before := outputBytes(t, f.output)
	fifo := filepath.Join(f.source, "Input.tl")
	if err := syscall.Mkfifo(fifo, 0600); err != nil {
		t.Fatal(err)
	}
	f.run(t, []string{"check", fifo}, "regular file")
	f.run(t, []string{"build", f.source, "-o", f.output}, "regular file")
	if err := os.Remove(fifo); err != nil {
		t.Fatal(err)
	}
	deps := filepath.Join(f.project, "dependencies")
	if err := os.Mkdir(deps, 0755); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Mkfifo(filepath.Join(deps, "Native.d.tl"), 0600); err != nil {
		t.Fatal(err)
	}
	f.run(t, []string{"build", f.source, "-I", deps, "-o", f.output}, "regular file")
	if !reflect.DeepEqual(before, outputBytes(t, f.output)) {
		t.Fatal("nonregular input changed previous output")
	}
}
func TestRuntimeImportsRejectWithoutConfusingCommentsOrStrings(t *testing.T) {
	f := newAddon(t)
	f.write(t, "Main.tl", "-- require(\"wow\") is text.\nprint('require(\"wow\")')\nprint([[require(\"native_values\")]])\n")
	f.run(t, []string{"build", f.source, "-o", f.output}, "")
	before := outputBytes(t, f.output)
	for _, expression := range []string{`require("wow")`, `require [[wow]]`, `require([=[wow]=])`, `_G.require("\119\111\119")`, `require("native_values")`, `(require)("wow")`, `(_G.require)("wow")`, `_G["require"]("wow")`, `require(("wow"))`} {
		t.Run(expression, func(t *testing.T) {
			f.write(t, "Main.tl", "local imported = "+expression+"\nprint(imported)\n")
			reject := "runtime declaration/compatibility import"
			if strings.Contains(expression, `\119`) {
				reject = "module not found"
			}
			f.run(t, []string{"build", f.source, "-o", f.output}, reject)
			if !reflect.DeepEqual(before, outputBytes(t, f.output)) {
				t.Fatal("runtime import changed previous output")
			}
			literal := filepath.Join(f.project, "Literal.lua")
			writeTest(t, literal, "local imported = "+expression+"\nprint(imported)\n")
			command := exec.CommandContext(t.Context(), teal(t), "run", filepath.Join(f.root, "tools/check_runtime_imports.lua"), "--", literal)
			command.Dir = f.project
			data, err := command.CombinedOutput()
			if err == nil || !strings.Contains(string(data), "runtime declaration/compatibility import") {
				t.Fatalf("literal import accepted: %v %s", err, data)
			}
		})
	}
}

func TestVersionByteAndFormatContract(t *testing.T) {
	root := t.TempDir()
	for _, raw := range []string{"12.1.5.70575", "12.1.5.70575\n"} {
		writeTest(t, filepath.Join(root, ".wow-version"), raw)
		version, err := ReadVersion(root)
		if err != nil {
			t.Fatal(err)
		}
		if version.Full != "12.1.5.70575" || version.Interface != 120105 || version.Build != 70575 {
			t.Fatalf("unexpected version: %+v", version)
		}
		if fileText(t, filepath.Join(root, ".wow-version")) != raw {
			t.Fatal("version parser mutated source bytes")
		}
	}
	for _, raw := range []string{"012.1.5.70575\n", "12.01.5.70575\n", "12.1.5.070575\n", "12.1.5.70575\n\n", "12.1.5.70575\r\n"} {
		writeTest(t, filepath.Join(root, ".wow-version"), raw)
		if _, err := ReadVersion(root); err == nil {
			t.Fatalf("accepted %q", raw)
		}
	}
}

func TestConsumerCLIUsesAddonWorkingDirectory(t *testing.T) {
	f := newAddon(t)
	repo := repositoryRoot(t)
	binary := filepath.Join(t.TempDir(), "wow-teal")
	command := exec.CommandContext(t.Context(), "go", "build", "-mod=readonly", "-o", binary, "./tools/wow-teal")
	command.Dir = repo
	command.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOWORK=off")
	data, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("build consumer: %v %s", err, data)
	}
	f.write(t, "Main.tl", "print(\"checked\")\n")
	command = exec.CommandContext(t.Context(), binary, "--root", f.root, "build", "src", "-o", "build/lua")
	command.Dir = f.project
	data, err = command.CombinedOutput()
	if err != nil {
		t.Fatalf("external CLI: %v %s", err, data)
	}
	if _, err = os.Stat(filepath.Join(f.output, "Main.lua")); err != nil {
		t.Fatal(err)
	}
}

func TestConsumerPackageIsReproducibleAndExcludesProducer(t *testing.T) {
	f := newAddon(t)
	repo := repositoryRoot(t)
	for _, name := range []string{"LICENSE", "README.md", "UPDATING.md", "tlconfig.lua", "docs/api/teal-catalogue.md", "docs/architecture.md", "examples/minimal/WowTealExample.tl", "examples/minimal/WowTealExample.toc", "Makefile", "check-all.sh", "rockspecs/wow-tl-type-0.1.0-1.rockspec.in"} {
		writeTest(t, filepath.Join(f.root, name), fileText(t, filepath.Join(repo, name)))
	}
	first, err := Package(f.root)
	if err != nil {
		t.Fatal(err)
	}
	before := fileText(t, str(first["archive"]))
	second, err := Package(f.root)
	if err != nil {
		t.Fatal(err)
	}
	if first["sha256"] != second["sha256"] || before != fileText(t, str(second["archive"])) {
		t.Fatal("consumer package is not repeatable")
	}
	archive, err := zip.OpenReader(str(first["archive"]))
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()
	for _, member := range archive.File {
		if strings.HasSuffix(member.Name, ".py") || strings.Contains(member.Name, "/scripts/") || strings.Contains(member.Name, "/tools/cdn/") {
			t.Fatalf("producer or Python in consumer archive: %s", member.Name)
		}
	}
}
