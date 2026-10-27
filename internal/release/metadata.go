package release

import (
	"fmt"
	"path/filepath"
	"regexp"

	"wow-teal-types/internal/consumer"
	"wow-teal-types/internal/sourcelock"
)

func replaceOnce(data []byte, pattern, replacement, context string) ([]byte, error) {
	re := regexp.MustCompile(pattern)
	if len(re.FindAllIndex(data, -1)) != 1 {
		return nil, fmt.Errorf("expected one generated %s", context)
	}
	return re.ReplaceAll(data, []byte(replacement)), nil
}

// prepareMetadata derives a complete metadata cohort without changing its inputs.
func prepareMetadata(root string) (map[string][]byte, error) {
	var err error
	root, err = filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	selected, err := consumer.ReadVersion(root)
	if err != nil {
		return nil, err
	}
	native, err := sourcelock.ReadSources(filepath.Join(root, "internal/wowapi/inputs/sources.json"))
	if err != nil {
		return nil, err
	}
	client, err := sourcelock.ReadSnapshot(filepath.Join(root, "internal/wowapi/inputs/client_sources.json"))
	if err != nil {
		return nil, err
	}
	if err = selected.Require(client.Version, client.Build); err != nil {
		return nil, err
	}
	versions := native["blizzard_version"].Files
	versionBytes := []byte(selected.Full + "\n")
	if len(versions) != 1 || versions[0].Name != "version.txt" || versions[0].SHA != blob(versionBytes) || versions[0].Size != len(versionBytes) {
		return nil, fmt.Errorf("native source version lock does not match .wow-version")
	}
	manifestData, err := readFile(filepath.Join(root, "package.json"), maxFile)
	if err != nil {
		return nil, err
	}
	manifest, err := consumer.ParsePackageManifest(manifestData)
	if err != nil {
		return nil, err
	}
	files := map[string]string{}
	for _, name := range consumer.RequiredFiles() {
		files[name] = ""
	}
	manifest.Files = files
	manifest.WoWBuild = selected.Full
	manifest.TealVersion = consumer.TealVersion
	outputs := map[string][]byte{}
	read := func(name string) ([]byte, error) {
		path, e := safeJoin(root, name)
		if e != nil {
			return nil, e
		}
		return readFile(path, maxPayload)
	}
	data, err := read("README.md")
	if err != nil {
		return nil, err
	}
	outputs["README.md"], err = replaceOnce(data, `(?m)^Targets Retail \*\*[^*\r\n]+\*\*`, "Targets Retail **"+selected.Full+"**", "README target-version line")
	if err != nil {
		return nil, err
	}
	data, err = read("examples/minimal/WowTealExample.toc")
	if err != nil {
		return nil, err
	}
	outputs["examples/minimal/WowTealExample.toc"], err = replaceOnce(data, `(?m)^## Interface: [^\r\n]+`, fmt.Sprintf("## Interface: %d", selected.Interface), "example TOC interface line")
	if err != nil {
		return nil, err
	}
	for _, name := range keys(files) {
		if name == "package.json" {
			return nil, fmt.Errorf("package metadata cannot fingerprint itself")
		}
		data, exists := outputs[name]
		if !exists {
			data, err = read(name)
			if err != nil {
				return nil, err
			}
		}
		files[name] = sha(data)
	}
	template := "rockspecs/wow-tl-type-" + manifest.Version + "-1.rockspec.in"
	templates, err := filepath.Glob(filepath.Join(root, "rockspecs/*.rockspec.in"))
	if err != nil {
		return nil, err
	}
	for _, path := range templates {
		if path != filepath.Join(root, template) {
			return nil, fmt.Errorf("remove superseded recipe templates after reviewing the package version change")
		}
	}
	recipe, err := consumer.Recipe(root, manifest, "", nil)
	if err != nil {
		return nil, err
	}
	outputs[template] = []byte(recipe)
	outputs["package.json"], err = jsonBytes(manifest)
	if err != nil {
		return nil, err
	}
	return outputs, nil
}
