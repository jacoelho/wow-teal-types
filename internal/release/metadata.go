package release

import (
	"fmt"
	"path/filepath"
	"regexp"

	"wow-teal-types/internal/consumer"
	"wow-teal-types/internal/sourcelock"
	"wow-teal-types/internal/wowapi"
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
	project := wowapi.NewProject(root)
	native, err := sourcelock.ReadSources(project.SourcesPath())
	if err != nil {
		return nil, err
	}
	client, err := sourcelock.ReadSnapshot(project.ClientSourcesPath())
	if err != nil {
		return nil, err
	}
	if err = selected.Require(client.Version, client.Build); err != nil {
		return nil, err
	}
	versions := native["blizzard_version"].Files
	versionBytes := []byte(selected.Full + "\n")
	if len(versions) != 1 || versions[0].Name != "version.txt" || !versions[0].Matches(versionBytes) {
		return nil, fmt.Errorf("native source version lock does not match .wow-version")
	}
	manifestData, err := sourcelock.ReadRegular(filepath.Join(root, consumer.ManifestFile), maxFile)
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
		return sourcelock.ReadRegular(path, maxPayload)
	}
	data, err := read("examples/WowTealExample/WowTealExample.toc")
	if err != nil {
		return nil, err
	}
	outputs["examples/WowTealExample/WowTealExample.toc"], err = replaceOnce(data, `(?m)^## Interface: [^\r\n]+`, fmt.Sprintf("## Interface: %d", selected.Interface), "example TOC interface line")
	if err != nil {
		return nil, err
	}
	for _, name := range keys(files) {
		data, exists := outputs[name]
		if !exists {
			data, err = read(name)
			if err != nil {
				return nil, err
			}
		}
		files[name] = sha(data)
	}
	template, recipe, err := consumer.RecipeTemplate(root, manifest)
	if err != nil {
		return nil, err
	}
	outputs[template] = []byte(recipe)
	outputs[consumer.ManifestFile], err = jsonBytes(manifest)
	if err != nil {
		return nil, err
	}
	return outputs, nil
}
