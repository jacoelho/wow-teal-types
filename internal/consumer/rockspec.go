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
      Requires the matching wow-teal toolkit and its bundled compiler %s.
      The toolkit also supplies the required secret-operation checks.
   ]],
   homepage = "https://github.com/%s",
   license = %s,
%s}

dependencies = {
   "lua >= 5.1",
}

build = {
   type = "builtin",
   modules = {},
   install = {
      lua = {
%s      },
   },
}
`, version, repository, version, manifest.WoWBuild, compiler, repository, quoteJSON(license), additions.String(), installs.String()), nil
}

func Rockspecs(root string, args []string, out, errout io.Writer) error {
	repository, encoded, output := "", "", filepath.Join(root, "build/rockspecs")
	descriptionSet, check := false, false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--check":
			check = true
		case "--repository", "--description-json", "--output":
			if i+1 == len(args) {
				return fmt.Errorf("missing argument for %s", args[i])
			}
			option := args[i]
			i++
			switch option {
			case "--repository":
				repository = args[i]
			case "--description-json":
				encoded = args[i]
				descriptionSet = true
			case "--output":
				output = args[i]
			}
		default:
			return fmt.Errorf("unknown rockspec option: %s", args[i])
		}
	}
	if repository != "" && !repositoryName.MatchString(repository) {
		return fmt.Errorf("--repository must be an actual GitHub OWNER/REPOSITORY")
	}
	var description map[string]string
	var err error
	if descriptionSet {
		description, err = descriptionFields(encoded)
		if err != nil {
			return err
		}
		if repository == "" {
			return fmt.Errorf("registry description metadata requires --repository OWNER/REPOSITORY")
		}
	}
	manifest, err := readManifest(filepath.Join(root, "package.json"))
	if err != nil {
		return err
	}
	expected, err := Recipe(root, manifest, "", nil)
	if err != nil {
		return err
	}
	filename := "wow-tl-type-" + manifest.Version + "-1.rockspec"
	template := filepath.Join(root, "rockspecs", filename+".in")
	templates, err := filepath.Glob(filepath.Join(root, "rockspecs/*.rockspec.in"))
	if err != nil {
		return err
	}
	for _, path := range templates {
		if path != template {
			return fmt.Errorf("remove superseded recipe templates after reviewing the version change")
		}
	}
	existing, err := os.ReadFile(template)
	if err != nil || string(existing) != expected {
		return fmt.Errorf("recipe template is stale; regenerate the reviewed cohort with wow-types generate")
	}
	if check && repository == "" {
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
	target := filepath.Join(output, filename)
	if check {
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
