package release

import (
	"fmt"
	"path/filepath"
	"regexp"

	"wow-teal-types/internal/consumer"
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
	load := func(name string) (map[string]any, error) {
		path, e := safeJoin(root, name)
		if e != nil {
			return nil, e
		}
		return readObject(path)
	}
	native, err := load("scripts/wow_api/sources.json")
	if err != nil {
		return nil, err
	}
	client, err := load("scripts/wow_api/client_sources.json")
	if err != nil {
		return nil, err
	}
	if err = selected.Require(client); err != nil {
		return nil, err
	}
	versions, ok := obj(native["blizzard_version"])["files"].([]any)
	versionBytes := []byte(selected.Full + "\n")
	if !ok || len(versions) != 1 {
		return nil, fmt.Errorf("native source version lock does not match .wow-version")
	}
	version := obj(versions[0])
	if version["name"] != "version.txt" || version["sha"] != blob(versionBytes) || i(version["size"]) != len(versionBytes) {
		return nil, fmt.Errorf("native source version lock does not match .wow-version")
	}
	manifest, err := load("package.json")
	if err != nil {
		return nil, err
	}
	provenance, err := load("provenance.json")
	if err != nil {
		return nil, err
	}
	files := map[string]any{}
	for _, name := range consumer.RequiredFiles() {
		files[name] = ""
	}
	manifest["files"] = files
	manifest["wow_build"] = selected.Full
	provenance["package_version"] = manifest["version"]
	compiler := obj(provenance["compiler"])
	if compiler == nil {
		compiler = map[string]any{}
		provenance["compiler"] = compiler
	}
	compiler["version"] = manifest["teal_version"]
	provenance["client"] = map[string]any{"version": selected.Version, "build": selected.Build, "interface": selected.Interface, "flavor": "retail"}
	lockNames := map[string]bool{"client_sources.json": true}
	for name := range obj(provenance["source_lock_sha256"]) {
		if name != "extractor.json" {
			lockNames[name] = true
		}
	}
	locks := map[string]any{}
	for _, name := range keys(lockNames) {
		path, e := safeJoin(root, "scripts/wow_api/"+name)
		if e != nil {
			return nil, e
		}
		data, e := readFile(path, maxPayload)
		if e != nil {
			return nil, e
		}
		locks[name] = sha(data)
	}
	provenance["source_lock_sha256"] = locks
	sources := obj(provenance["sources"])
	if sources == nil {
		sources = map[string]any{}
		provenance["sources"] = sources
	}
	for name, value := range native {
		copy := map[string]any{}
		for key, value := range obj(value) {
			if key != "files" && key != "path" {
				copy[key] = value
			}
		}
		sources[name] = copy
	}
	clientSource := map[string]any{}
	for key, value := range client {
		if key != "files" {
			clientSource[key] = value
		}
	}
	sources["client_casc"] = clientSource
	outputs := map[string][]byte{}
	outputs["provenance.json"], err = jsonBytes(provenance)
	if err != nil {
		return nil, err
	}
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
	template := "rockspecs/wow-tl-type-" + s(manifest["version"]) + "-1.rockspec.in"
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
