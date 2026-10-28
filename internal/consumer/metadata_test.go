package consumer

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func recipeFixture(t *testing.T) addonFixture {
	t.Helper()
	f := newAddon(t)
	repo := repositoryRoot(t)
	writeTest(t, filepath.Join(f.root, "LICENSE"), fileText(t, filepath.Join(repo, "LICENSE")))
	manifest := readManifest(t, filepath.Join(f.root, "package.json"))
	recipe, err := Recipe(f.root, manifest, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	writeTest(t, filepath.Join(f.root, "rockspecs/wow-tl-type-"+manifest.Version+"-1.rockspec.in"), recipe)
	return f
}
func recipeCommand(t *testing.T, f addonFixture, options RockspecOptions, reject string) string {
	t.Helper()
	options.Output = filepath.Join(f.project, "candidate")
	var out bytes.Buffer
	return expect(t, Rockspecs(f.root, options, &out, &out), &out, reject)
}
func TestRecipeInstallsOnlyFourModulesWithMatchingToolkit(t *testing.T) {
	f := recipeFixture(t)
	manifest := readManifest(t, filepath.Join(f.root, "package.json"))
	recipeCommand(t, f, RockspecOptions{Check: true}, "")
	recipeCommand(t, f, RockspecOptions{Repository: "example/wow-teal-types"}, "")
	path := filepath.Join(f.project, "candidate/wow-tl-type-"+manifest.Version+"-1.rockspec")
	before := fileText(t, path)
	recipeCommand(t, f, RockspecOptions{Repository: "example/wow-teal-types"}, "")
	if before != fileText(t, path) {
		t.Fatal("recipe is nondeterministic")
	}
	recipeCommand(t, f, RockspecOptions{Repository: "example/wow-teal-types", Check: true}, "")
	inspector := filepath.Join(f.project, "Inspect.lua")
	writeTest(t, inspector, `local data = {}
local chunk
if setfenv then
   chunk = assert(loadfile(arg[1]))
   setfenv(chunk, data)
else
   chunk = assert(loadfile(arg[1], "t", data))
end
chunk()
print(data.package)
print(data.version)
print(data.source.url)
print(data.source.tag)
print(data.description.license)
assert(#data.dependencies == 2 and data.dependencies[1] == "lua >= 5.1" and data.dependencies[2] == "tl == 0.24.8")
for _, path in ipairs(data.build.install.lua) do print(path) end
`)
	command := exec.CommandContext(t.Context(), teal(t), "run", inspector, "--", path)
	command.Dir = f.project
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("recipe Lua parse: %v %s", err, output)
	}
	expected := "wow-tl-type\n" + manifest.Version + "-1\ngit+https://github.com/example/wow-teal-types\nv" + manifest.Version + "\nMIT\ntypes/wow/native_values.d.tl\ntypes/wow/wow.d.tl\ntypes/wow/wow_env.d.tl\ntypes/wow/wow_secure_env.d.tl\n"
	if string(output) != expected {
		t.Fatalf("unexpected recipe installation contract:\n%s", output)
	}
}
func TestStaleRecipesAreRejectedWithoutRewrites(t *testing.T) {
	f := recipeFixture(t)
	recipeCommand(t, f, RockspecOptions{Repository: "example/wow-teal-types"}, "")
	path := filepath.Join(f.project, "candidate/wow-tl-type-0.1.0-1.rockspec")
	writeTest(t, path, fileText(t, path)+"-- stale\n")
	before := fileText(t, path)
	recipeCommand(t, f, RockspecOptions{Repository: "example/wow-teal-types", Check: true}, "missing or stale")
	if fileText(t, path) != before {
		t.Fatal("candidate check rewrote output")
	}
	path = filepath.Join(f.root, "rockspecs/wow-tl-type-0.1.0-1.rockspec.in")
	writeTest(t, path, fileText(t, path)+"-- stale\n")
	before = fileText(t, path)
	recipeCommand(t, f, RockspecOptions{Check: true}, "template is stale")
	if fileText(t, path) != before {
		t.Fatal("template check rewrote output")
	}
}
func TestRecipeRequiresSourceIdentityAndIntactDeclarations(t *testing.T) {
	f := recipeFixture(t)
	recipeCommand(t, f, RockspecOptions{}, "requires --repository")
	if _, err := os.Stat(filepath.Join(f.project, "candidate")); !os.IsNotExist(err) {
		t.Fatal("missing source identity created candidate")
	}
	path := filepath.Join(f.root, "types/wow/wow.d.tl")
	writeTest(t, path, fileText(t, path)+"-- changed\n")
	recipeCommand(t, f, RockspecOptions{Repository: "example/wow-teal-types"}, "integrity mismatch")
}
func TestOwnerDescriptionRoundTripsAndCannotOverrideCoreFields(t *testing.T) {
	f := recipeFixture(t)
	value := "Maintainer \"Quoted\" \\ José"
	metadata, err := json.Marshal(map[string]string{"maintainer": value})
	if err != nil {
		t.Fatal(err)
	}
	options := RockspecOptions{Repository: "example/wow-teal-types", DescriptionJSON: string(metadata)}
	recipeCommand(t, f, options, "")
	options.Check = true
	recipeCommand(t, f, options, "")
	path := filepath.Join(f.project, "candidate/wow-tl-type-0.1.0-1.rockspec")
	inspector := filepath.Join(f.project, "Metadata.lua")
	writeTest(t, inspector, `local data = {}
local chunk
if setfenv then chunk=assert(loadfile(arg[1]));setfenv(chunk,data)
else chunk=assert(loadfile(arg[1],"t",data)) end
chunk()
print(data.description.maintainer)
`)
	command := exec.CommandContext(t.Context(), teal(t), "run", inspector, "--", path)
	command.Dir = f.project
	output, err := command.CombinedOutput()
	if err != nil || strings.TrimSuffix(string(output), "\n") != value {
		t.Fatalf("owner metadata did not roundtrip: %v %s", err, output)
	}
	before := fileText(t, path)
	for _, invalid := range []string{`{"summary":"override"}`, `{"license":"other"}`, `{"maintainer":[]}`, `{"maintainer":"line\nbreak"}`, `null`, `{"maintainer":null}`, `{"maintainer":17}`, `[]`} {
		recipeCommand(t, f, RockspecOptions{Repository: "example/wow-teal-types", DescriptionJSON: invalid}, "description")
		if fileText(t, path) != before {
			t.Fatal("invalid description changed candidate")
		}
	}
}

func TestSourceAndDependencyBounds(t *testing.T) {
	f := newAddon(t)
	f.write(t, "Good.tl", "print(\"checked\")\n")
	f.write(t, "Big.tl", strings.Repeat(" ", maxSourceBytes+1))
	f.check(t, Selection{Sources: []string{f.source}}, "source exceeds 2 MiB")
	if err := os.Remove(filepath.Join(f.source, "Big.tl")); err != nil {
		t.Fatal(err)
	}
	deps := filepath.Join(f.project, "types")
	writeTest(t, filepath.Join(deps, "Big.d.tl"), strings.Repeat(" ", maxSourceBytes+1))
	f.check(t, Selection{Sources: []string{f.source}, Include: []string{deps}}, "regular file of at most 2 MiB")
}
