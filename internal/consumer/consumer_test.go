package consumer

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
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
	manifest := readManifest(t, filepath.Join(repo, "package.json"))
	files := map[string]string{}
	for _, name := range RequiredFiles() {
		data, err := os.ReadFile(filepath.Join(repo, filepath.FromSlash(name)))
		if err != nil {
			t.Fatal(err)
		}
		writeTest(t, filepath.Join(root, filepath.FromSlash(name)), string(data))
		files[name] = fingerprint(data)
	}
	manifest.Files = files
	data, err := json.Marshal(manifest)
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
func (f addonFixture) check(t *testing.T, selected Selection, reject string) string {
	t.Helper()
	var out bytes.Buffer
	return expect(t, Check(t.Context(), f.root, selected, &out, &out), &out, reject)
}
func (f addonFixture) build(t *testing.T, selected Selection, output, reject string) string {
	t.Helper()
	var out bytes.Buffer
	return expect(t, Build(t.Context(), f.root, selected, output, &out, &out), &out, reject)
}
func (f addonFixture) verify(t *testing.T, reject string) string {
	t.Helper()
	manifest, err := Verify(f.root)
	return expect(t, err, bytes.NewBufferString(manifest.Version), reject)
}

// expect checks a consumer result: success when reject is empty, otherwise an
// error whose text or captured output contains reject.
func expect(t *testing.T, err error, output *bytes.Buffer, reject string) string {
	t.Helper()
	text := output.String()
	if reject == "" {
		if err != nil {
			t.Fatalf("unexpected rejection: %v\n%s", err, text)
		}
		return text
	}
	if err == nil {
		t.Fatalf("accepted; expected %q\n%s", reject, text)
	}
	text += "\n" + err.Error()
	if !strings.Contains(text, reject) {
		t.Fatalf("expected %q in %s", reject, text)
	}
	return text
}
func readManifest(t *testing.T, path string) PackageManifest {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := ParsePackageManifest(data)
	if err != nil {
		t.Fatal(err)
	}
	return manifest
}
func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate test source")
	}
	root, err := filepath.Abs(filepath.Join(filepath.Dir(file), "../.."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

// teal resolves the compiler by the consumer rule (TL, else tl on PATH) and
// requires the pinned stock version.
func teal(t *testing.T) string {
	t.Helper()
	name := os.Getenv("TL")
	if name == "" {
		name = "tl"
	}
	path, err := exec.LookPath(name)
	if err != nil {
		t.Fatalf("stock Teal %s required (%s): %v", TealVersion, tealInstall, err)
	}
	output, err := exec.CommandContext(t.Context(), path, "--version").Output()
	if err != nil {
		t.Fatalf("stock Teal %s required (%s): %v", TealVersion, tealInstall, err)
	}
	if err = admitTealVersion(string(output)); err != nil {
		t.Fatal(err)
	}
	return path
}
func writeTest(t *testing.T, path, source string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
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

func TestVerifyUsesPackageWithoutCompiler(t *testing.T) {
	f := newAddon(t)
	t.Setenv("TL", "/compiler-must-not-be-invoked")
	if got := f.verify(t, ""); got != "0.1.0" {
		t.Fatalf("verified version: %q", got)
	}
	writeTest(t, filepath.Join(f.root, "types/wow/wow.d.tl"), "changed declaration\n")
	f.verify(t, "package integrity mismatch")
}

func TestNativeFrameFactoryContracts(t *testing.T) {
	f := newAddon(t)
	f.write(t, "Main.tl", `local type W = require("wow")
local frame:W.Frame = CreateFrame("Frame")
local bar:W.StatusBar = CreateFrame("StatusBar", nil, frame)
bar:SetValue(1)
local button:W.Button = CreateFrame("Button")
local backdrop = CreateFrame("Frame", nil, frame, "BackdropTemplate")
backdrop:SetBackdrop({})
local backdropButton = CreateFrame("Button", "Named", frame, "BackdropTemplate", 1)
backdropButton:SetBackdrop({})
backdropButton:Click()
local kind:string = "Frame"
local any:W.Frame = CreateFrame(kind)
print(button, any)
`)
	f.check(t, Selection{Sources: []string{f.source}}, "")
	for _, test := range []struct{ source, reject string }{
		{`local type W = require("wow")
local bar:W.StatusBar = CreateFrame("Frame")`, "StatusBar"},
		{`local type W = require("wow")
local function construct(kind:string):W.StatusBar return CreateFrame(kind) end`, "StatusBar"},
		// BackdropTemplate narrows only the listed frame types and exact template name.
		{`CreateFrame("Cooldown", nil, nil, "BackdropTemplate"):SetBackdrop({})`, "SetBackdrop"},
		{`CreateFrame("Frame", nil, nil, "OtherTemplate"):SetBackdrop({})`, "SetBackdrop"},
	} {
		f.write(t, "Main.tl", test.source)
		f.check(t, Selection{Sources: []string{f.source}}, test.reject)
	}
}

// Teal admits nil for every type but cannot narrow `T | nil` by truthiness, so
// generated declarations state absence only through optional parameters.
func TestGeneratedDeclarationsOmitNilUnionMembers(t *testing.T) {
	member := regexp.MustCompile(`\|\s*nil\b|\bnil\s*\|`)
	for _, name := range []string{"wow.d.tl", "wow_env.d.tl", "wow_secure_env.d.tl"} {
		for number, line := range strings.Split(fileText(t, filepath.Join(repositoryRoot(t), "types/wow", name)), "\n") {
			if !strings.HasPrefix(strings.TrimSpace(line), "--") && member.MatchString(line) {
				t.Errorf("%s:%d declares a nil union member: %s", name, number+1, strings.TrimSpace(line))
			}
		}
	}
}

func TestFrameXMLSupplementContracts(t *testing.T) {
	f := newAddon(t)
	f.write(t, "Main.tl", `local type W = require("wow")
local mage = RAID_CLASS_COLORS["MAGE"]
local red:W.MaybeSecret<number> = mage.r
local alpha:W.MaybeSecret<number> = GRAY_FONT_COLOR.a
ChatFrame1:AddMessage("first")
ChatFrame10:AddMessage("last")
local unit:boolean = GameTooltip:IsTooltipType(Enum.TooltipDataType.Unit)
RaidWarningFrame:OnEvent("CHAT_MSG_RAID_WARNING", "message", "sender")
`)
	f.check(t, Selection{Sources: []string{f.source}}, "")
}

func TestRelocatedAddonBuildAndRepeat(t *testing.T) {
	f := newAddon(t)
	f.write(t, "WowTealExample.tl", fileText(t, filepath.Join(repositoryRoot(t), "examples/minimal/WowTealExample.tl")))
	writeTest(t, filepath.Join(f.project, "tlconfig.lua"), "error(\"ambient config must not run\")\n")
	t.Setenv("TL_PATH", "/missing-and-untrusted/?.lua")
	text := f.build(t, Selection{Sources: []string{f.source}}, f.output, "")
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
	f.build(t, Selection{Sources: []string{f.source}}, f.output, "")
	if !maps.Equal(before, outputBytes(t, f.output)) {
		t.Fatal("repeat changed checked build")
	}
}
func TestSelectedImportedBodiesReceiveSecretChecks(t *testing.T) {
	f := newAddon(t)
	f.write(t, "Entry.tl", "local helpers = require(\"Helpers\")\nhelpers.inspect(1)\n")
	f.write(t, "Helpers.tl", `local type W = require("wow")
local record Helpers
   inspect: function(W.MaybeSecret<number>)
end
local helpers: Helpers = {}
function helpers.inspect(value: W.MaybeSecret<number>)
   if value == 0 then print("bad") end
end
return helpers
`)
	f.check(t, Selection{Sources: []string{f.source}}, "secret-value comparison")
}
func TestInvalidSourceSelectionAndTypeErrors(t *testing.T) {
	f := newAddon(t)
	f.write(t, "Bad.tl", "C_Spell.GetBaseSpell(true)\n")
	f.check(t, Selection{Sources: []string{f.source}}, "argument")
	f.check(t, Selection{Sources: []string{filepath.Join(f.source, "Missing.tl")}}, "no such file")
	empty := filepath.Join(f.project, "empty")
	if err := os.Mkdir(empty, 0755); err != nil {
		t.Fatal(err)
	}
	f.check(t, Selection{Sources: []string{empty}}, "no Teal source files")
	f.check(t, Selection{Sources: []string{filepath.Join(f.root, "types/wow/wow.d.tl")}}, "expected a Teal source")
}
func TestManifestBindsTypesCheckersAndVersion(t *testing.T) {
	f := newAddon(t)
	for _, name := range []string{"types/wow/wow.d.tl", "tools/check_secrets.lua", ".wow-version", "internal/consumer/consumer.go"} {
		path := filepath.Join(f.root, name)
		before := fileText(t, path)
		writeTest(t, path, before+"\n-- changed\n")
		f.verify(t, "package integrity mismatch")
		writeTest(t, path, before)
	}
	f.verify(t, "")
}
func TestPackageResolutionAndDirectShadowFailure(t *testing.T) {
	f := newAddon(t)
	source := f.write(t, "Unsafe.tl", "local type W = require(\"wow\")\nlocal function f(v: W.MaybeSecret<number>) if v then print(v) end end\n")
	shadow := filepath.Join(f.project, "shadow")
	writeTest(t, filepath.Join(shadow, "wow.d.tl"), "local record Fake\n type MaybeSecret<T> = T\nend\nreturn Fake\n")
	f.check(t, Selection{Sources: []string{f.source}, Include: []string{shadow}}, "secret-value truthiness")
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
	f.build(t, Selection{Sources: []string{f.source}, Include: []string{declarations}, Global: "addon_env"}, f.output, "")
	if strings.Contains(fileText(t, filepath.Join(f.output, "Main.lua")), "require(") {
		t.Fatal("declaration import survived")
	}
	f.check(t, Selection{Sources: []string{f.source}}, "MyAddonCounter")
}
func TestFailedBuildPreservesOutputAndRejectsForeignDirectory(t *testing.T) {
	f := newAddon(t)
	f.write(t, "Good.tl", "print(\"checked\")\n")
	f.build(t, Selection{Sources: []string{f.source}}, f.output, "")
	before := outputBytes(t, f.output)
	f.write(t, "Bad.tl", "local type W = require(\"wow\")\nlocal function f(v: W.MaybeSecret<number>) if v == 0 then print(v) end end\n")
	f.build(t, Selection{Sources: []string{f.source}}, f.output, "secret-value comparison")
	if !maps.Equal(before, outputBytes(t, f.output)) {
		t.Fatal("failed build changed previous output")
	}
	f.build(t, Selection{Sources: []string{f.source}}, f.source, "output must not overlap")
	foreign := filepath.Join(f.project, "foreign")
	writeTest(t, filepath.Join(foreign, "mine.txt"), "preserve")
	f.build(t, Selection{Sources: []string{f.source}}, foreign, "no wow-teal ownership receipt")
	if fileText(t, filepath.Join(foreign, "mine.txt")) != "preserve" {
		t.Fatal("foreign output changed")
	}
}
func TestWrongCompilerAndEditedOutputDoNotPublish(t *testing.T) {
	f := newAddon(t)
	stock := teal(t)
	f.write(t, "Good.tl", "print(\"checked\")\n")
	f.build(t, Selection{Sources: []string{f.source}}, f.output, "")
	before := fileText(t, filepath.Join(f.output, "Good.lua"))
	for name, script := range map[string]string{
		"fork":    "printf '0.24.8-wow.1\\n'",
		"older":   "printf '0.24.7\\n'",
		"prefix":  "printf '0.24.80\\n'",
		"empty":   "true",
		"failing": "printf '0.24.8\\n'; exit 1",
		"future":  "printf '99.0.0\\n'",
	} {
		compiler := filepath.Join(f.project, name+"-tl")
		writeTest(t, compiler, "#!/bin/sh\n"+script+"\n")
		if err := os.Chmod(compiler, 0755); err != nil {
			t.Fatal(err)
		}
		t.Setenv("TL", compiler)
		text := f.build(t, Selection{Sources: []string{f.source}}, f.output, "unsupported Teal version")
		if !strings.Contains(text, tealInstall) {
			t.Fatalf("%s: rejection does not name %q:\n%s", name, tealInstall, text)
		}
		if fileText(t, filepath.Join(f.output, "Good.lua")) != before {
			t.Fatalf("%s compiler changed output", name)
		}
	}
	t.Setenv("TL", stock)
	writeTest(t, filepath.Join(f.output, "Good.lua"), "-- user edit\n")
	f.build(t, Selection{Sources: []string{f.source}}, f.output, "output has local changes")
	if fileText(t, filepath.Join(f.output, "Good.lua")) != "-- user edit\n" {
		t.Fatal("user edit replaced")
	}
}
func TestSymlinkAndExcludedSourceTrees(t *testing.T) {
	f := newAddon(t)
	source := f.write(t, "Good.tl", "print(\"checked\")\n")
	f.write(t, "build/Bad.tl", "not valid Teal")
	f.write(t, "vendor/Bad.tl", "not valid Teal")
	f.check(t, Selection{Sources: []string{f.source}}, "")
	if err := os.Symlink(source, filepath.Join(f.source, "Linked.tl")); err != nil {
		t.Fatal(err)
	}
	f.check(t, Selection{Sources: []string{f.source}}, "must not be a symlink")
}
func TestSecureEnvironmentSelection(t *testing.T) {
	f := newAddon(t)
	f.write(t, "Main.tl", "local forbidden = GetForbiddenObjectTable(CreateFrame(\"Frame\"))\nprint(forbidden)\n")
	f.check(t, Selection{Sources: []string{f.source}}, "GetForbiddenObjectTable")
	f.check(t, Selection{Sources: []string{f.source}, Secure: true}, "")
}
func TestNestedModulesResolveFromSelectedRoot(t *testing.T) {
	f := newAddon(t)
	f.write(t, "myaddon/Entry.tl", "local helper = require(\"myaddon.Helpers\")\nprint(helper.value)\n")
	f.write(t, "myaddon/Helpers.tl", "local helper = { value = 17 }\nreturn helper\n")
	f.build(t, Selection{Sources: []string{f.source}}, f.output, "")
	for _, name := range []string{"myaddon/Entry.lua", "myaddon/Helpers.lua"} {
		if _, err := os.Stat(filepath.Join(f.output, name)); err != nil {
			t.Fatal(err)
		}
	}
}
func TestNonregularSourcesAndDependenciesRejectBeforeReading(t *testing.T) {
	f := newAddon(t)
	f.write(t, "Good.tl", "print(\"checked\")\n")
	f.build(t, Selection{Sources: []string{f.source}}, f.output, "")
	before := outputBytes(t, f.output)
	fifo := filepath.Join(f.source, "Input.tl")
	if err := syscall.Mkfifo(fifo, 0600); err != nil {
		t.Fatal(err)
	}
	f.check(t, Selection{Sources: []string{fifo}}, "regular file")
	f.build(t, Selection{Sources: []string{f.source}}, f.output, "regular file")
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
	f.build(t, Selection{Sources: []string{f.source}, Include: []string{deps}}, f.output, "regular file")
	if !maps.Equal(before, outputBytes(t, f.output)) {
		t.Fatal("nonregular input changed previous output")
	}
}
func TestRuntimeImportsRejectWithoutConfusingCommentsOrStrings(t *testing.T) {
	f := newAddon(t)
	f.write(t, "Main.tl", "-- require(\"wow\") is text.\nprint('require(\"wow\")')\nprint([[require(\"native_values\")]])\n")
	f.build(t, Selection{Sources: []string{f.source}}, f.output, "")
	before := outputBytes(t, f.output)
	for _, expression := range []string{`require("wow")`, `require [[wow]]`, `require([=[wow]=])`, `_G.require("\119\111\119")`, `require("native_values")`, `(require)("wow")`, `(_G.require)("wow")`, `_G["require"]("wow")`, `require(("wow"))`} {
		t.Run(expression, func(t *testing.T) {
			f.write(t, "Main.tl", "local imported = "+expression+"\n")
			reject := "runtime declaration/compatibility import"
			if strings.Contains(expression, `\119`) {
				reject = "module not found"
			}
			f.build(t, Selection{Sources: []string{f.source}}, f.output, reject)
			if !maps.Equal(before, outputBytes(t, f.output)) {
				t.Fatal("runtime import changed previous output")
			}
			literal := filepath.Join(f.project, "Literal.lua")
			writeTest(t, literal, "local imported = "+expression+"\n")
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
	before := fileText(t, first.Archive)
	second, err := Package(f.root)
	if err != nil {
		t.Fatal(err)
	}
	if first.SHA256 != second.SHA256 || before != fileText(t, second.Archive) {
		t.Fatal("consumer package is not repeatable")
	}
	archive, err := zip.OpenReader(first.Archive)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()
	for _, member := range archive.File {
		if strings.HasSuffix(member.Name, ".py") || strings.Contains(member.Name, "/tools/teal/") || strings.Contains(member.Name, "/scripts/") || strings.Contains(member.Name, "/tools/cdn/") {
			t.Fatalf("producer or Python in consumer archive: %s", member.Name)
		}
	}
}
