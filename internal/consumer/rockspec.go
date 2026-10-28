package consumer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"unicode"
)

var releaseVersion = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)
var repositoryName = regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9-]{0,38})/[A-Za-z0-9_.-]+$`)
var descriptionKey = regexp.MustCompile(`^[a-z_]+$`)

func printable(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}
func quoteJSON(value string) string {
	var b bytes.Buffer
	e := json.NewEncoder(&b)
	e.SetEscapeHTML(false)
	_ = e.Encode(value)
	return strings.TrimSuffix(b.String(), "\n")
}
func descriptionFields(encoded string) (map[string]string, error) {
	fields := map[string]string{}
	if len(encoded) > 16384 {
		return nil, fmt.Errorf("registry description metadata exceeds 16384 characters")
	}
	var values map[string]string
	if err := json.Unmarshal([]byte(encoded), &values); err != nil {
		return nil, fmt.Errorf("registry description metadata: %w", err)
	}
	if values == nil || len(values) > 32 {
		return nil, fmt.Errorf("registry description metadata must be a JSON string-to-string object")
	}
	for key, value := range values {
		if !descriptionKey.MatchString(key) {
			return nil, fmt.Errorf("registry description keys must use lowercase letters and underscores")
		}
		switch key {
		case "summary", "detailed", "homepage", "license":
			return nil, fmt.Errorf("registry description metadata cannot replace %s", key)
		}
		if !printable(value) {
			return nil, fmt.Errorf("registry description values must be nonempty printable strings")
		}
		fields[key] = value
	}
	return fields, nil
}

func Recipe(root string, manifest PackageManifest, repository string, description map[string]string) (string, error) {
	if repository == "" {
		repository = "@REPOSITORY@"
	}
	version, compiler := manifest.Version, manifest.TealVersion
	license := manifest.License
	if !printable(license) {
		return "", fmt.Errorf("package license must be a nonempty printable string")
	}
	if _, err := checkedPath(root, "LICENSE"); err != nil {
		return "", fmt.Errorf("project LICENSE file is required: %w", err)
	}
	if !releaseVersion.MatchString(version) || compiler != TealVersion {
		return "", fmt.Errorf("package version must use major.minor.patch and compiler must match %s", TealVersion)
	}
	files := manifest.Files
	expected := []string{}
	for _, name := range modules {
		expected = append(expected, "types/wow/"+name+".d.tl")
	}
	sort.Strings(expected)
	declared := []string{}
	for name := range files {
		if strings.HasPrefix(name, "types/") && strings.HasSuffix(name, ".d.tl") {
			declared = append(declared, name)
		}
	}
	sort.Strings(declared)
	if strings.Join(expected, "\n") != strings.Join(declared, "\n") {
		return "", fmt.Errorf("declaration inventory differs from the four reviewed WoW modules")
	}
	var installs strings.Builder
	for _, name := range expected {
		path, err := checkedPath(root, name)
		if err != nil {
			return "", err
		}
		hash, err := fileHash(path)
		if err != nil {
			return "", err
		}
		if hash != files[name] {
			return "", fmt.Errorf("declaration integrity mismatch: %s", name)
		}
		fmt.Fprintf(&installs, "         %q,\n", name)
	}
	var additions strings.Builder
	for _, key := range slices.Sorted(maps.Keys(description)) {
		fmt.Fprintf(&additions, "   [%s] = %s,\n", quoteJSON(key), quoteJSON(description[key]))
	}
	return fmt.Sprintf(`rockspec_format = "3.0"
package = "wow-tl-type"
version = "%s-1"

source = {
   url = "git+https://github.com/%s",
   tag = "v%s",
}

description = {
   summary = "Retail World of Warcraft declaration files for Teal",
   detailed = [[
      Declaration-only snapshot for Retail %s.
      Requires the matching wow-teal toolkit, which supplies the required
      secret-operation checks, and Teal %s.
   ]],
   homepage = "https://github.com/%s",
   license = %s,
%s}

dependencies = {
   "lua >= 5.1",
   "tl == %s",
}

build = {
   type = "builtin",
   modules = {},
   install = {
      lua = {
%s      },
   },
}
`, version, repository, version, manifest.WoWBuild, compiler, repository, quoteJSON(license), additions.String(), compiler, installs.String()), nil
}

// RecipeTemplate returns the reviewed template's package-relative path and its
// expected text for manifest, rejecting templates left from another version.
func RecipeTemplate(root string, manifest PackageManifest) (string, string, error) {
	name := "rockspecs/wow-tl-type-" + manifest.Version + "-1.rockspec.in"
	if err := checkRecipeTemplates(root, name); err != nil {
		return "", "", err
	}
	text, err := Recipe(root, manifest, "", nil)
	if err != nil {
		return "", "", err
	}
	return name, text, nil
}

func checkRecipeTemplates(root, name string) error {
	templates, err := filepath.Glob(filepath.Join(root, "rockspecs/*.rockspec.in"))
	if err != nil {
		return err
	}
	for _, path := range templates {
		if path != filepath.Join(root, filepath.FromSlash(name)) {
			return fmt.Errorf("remove superseded recipe templates after reviewing the package version change")
		}
	}
	return nil
}

// RockspecOptions selects candidate recipe generation or checking.
type RockspecOptions struct {
	// Repository is the GitHub OWNER/REPOSITORY; empty checks only the template.
	Repository string
	// DescriptionJSON adds registry description fields; empty adds none.
	DescriptionJSON string
	// Output is the candidate directory; empty selects build/rockspecs.
	Output string
	// Check compares instead of writing.
	Check bool
}

func Rockspecs(root string, options RockspecOptions, out, errout io.Writer) error {
	root, err := canonicalExisting(root)
	if err != nil {
		return err
	}
	repository, output := options.Repository, options.Output
	if output == "" {
		output = filepath.Join(root, "build/rockspecs")
	}
	if repository != "" && !repositoryName.MatchString(repository) {
		return fmt.Errorf("--repository must be an actual GitHub OWNER/REPOSITORY")
	}
	var description map[string]string
	if options.DescriptionJSON != "" {
		description, err = descriptionFields(options.DescriptionJSON)
		if err != nil {
			return err
		}
		if repository == "" {
			return fmt.Errorf("registry description metadata requires --repository OWNER/REPOSITORY")
		}
	}
	manifest, err := Verify(root)
	if err != nil {
		return err
	}
	name, expected, err := RecipeTemplate(root, manifest)
	if err != nil {
		return err
	}
	existing, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
	if err != nil || string(existing) != expected {
		return fmt.Errorf("recipe template is stale; regenerate the reviewed cohort with wow-types generate")
	}
	if options.Check && repository == "" {
		fmt.Fprintln(out, "Recipe template matches the reviewed declaration inventory and version.")
		return nil
	}
	if repository == "" {
		return fmt.Errorf("generation requires --repository OWNER/REPOSITORY; no public repository has been assumed")
	}
	concrete, err := Recipe(root, manifest, repository, description)
	if err != nil {
		return err
	}
	target := filepath.Join(output, strings.TrimSuffix(filepath.Base(name), ".in"))
	if options.Check {
		data, e := os.ReadFile(target)
		if e != nil || string(data) != concrete {
			return fmt.Errorf("generated candidate is missing or stale: %s", target)
		}
		fmt.Fprintln(out, "Candidate recipe matches "+repository+" and the reviewed snapshot.")
		return nil
	}
	if err = atomicWrite(target, []byte(concrete), 0644); err != nil {
		return err
	}
	fmt.Fprintln(out, target)
	fmt.Fprintln(errout, "LuaRocks lint is required before uploading.")
	return nil
}
