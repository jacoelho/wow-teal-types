package wowapi

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"wow-teal-types/internal/consumer"
)

const probeVersion = "1.5.0"

var probeRuntimeFiles = []string{"WoWAPIProbe/Plan.lua", "WoWAPIProbe/Runtime.lua", "WoWAPIProbe/Invocations.lua", "WoWAPIProbe/Fixture.svg", "WoWAPIProbe/WoWAPIProbe.toc"}
var probePreviousRuntimeFiles = []string{"WoWAPIProbe/Plan.lua", "WoWAPIProbe/Runtime.lua", "WoWAPIProbe/WoWAPIProbe.toc"}
var probeCandidateEvidence = M{"repository": "Ketho/BlizzardInterfaceResources", "commit": "f3b334b33bd19fa6a0250f7922ecfa83db7fe356", "build": "12.1.5.70448", "path": "Resources/WidgetAPI.lua", "git_blob_sha1": "adfe0e698f9f71e9f6ef1f0a0e8f889a202fd0f9", "sha256": "87b57900d2b21f21fc27c82f102a0eb4be97479f9dd8af687aeb68f45009d150", "role": "method-presence candidates; no inherited contract or invocation", "methods": []any{"GetRotation", "SetRotation"}}
var probeNativeCandidateEvidence = M{"capture": "references/wow-api/native-observations/12.1.5.70575/public-invocations-1.1.0/report.json", "sha256": "83fdc1d36d4e3e9f4c8cfe541f3391c3271621d2e0d93bf646c6e22cee4e7ef2", "build": "12.1.5.70575", "role": "historical native disagreements nominate lookup-only regression cases"}
var probeNativeCandidates = []any{M{"canonical": "RadialProgress", "parent": "native.HistoricalAbsences", "methods": []any{"SetPreventSecretValues"}}, M{"canonical": "VectorGraphics", "parent": "native.HistoricalAbsences", "methods": []any{"CreateAnimationGroup", "GetAnimationGroups", "SetPreventSecretValues", "StopAnimating"}}}
var probeRepresentativeMethods = []any{"SetPreventSecretValues", "CreateAnimationGroup", "GetAnimationGroups", "StopAnimating"}
var probeRepresentatives = []any{M{"canonical": "Frame", "resourceKey": "frame", "constructor": "CreateFrame('Frame')"}, M{"canonical": "Texture", "resourceKey": "texture", "constructor": "owned Frame:CreateTexture(nil, 'ARTWORK')"}, M{"canonical": "FontString", "resourceKey": "font", "constructor": "owned Frame:CreateFontString(nil, 'ARTWORK')"}}
var probeDiagnosticCases = []any{"setup.font_object_set", "setup.clock_set_time", "setup.duration_set_clock", "setup.duration_set_span", "setup.font_object_identity", "setup.duration_clock_identity", "setup.duration_clock_time", "setup.duration_remaining", "baseline.expected_before", "baseline.expected_after", "baseline.font_after", "baseline.expected_stable", "baseline.font_matches", "sentinel.set", "sentinel.read", "sentinel.matches", "retry.update", "retry.expected", "retry.font", "retry.matches", "zero_interval.set", "zero_interval.update", "zero_interval.expected", "zero_interval.font", "zero_interval.matches"}
var probeSampleCases = []any{"baseline.expected", "baseline.font", "projection.font", "manual_step.expected", "manual_step.font", "default_clock.expected", "default_clock.font"}
var probeExperimentCases = []any{"projection.set", "projection.matches", "manual_step.set_time", "manual_step.clock_time", "manual_step.remaining", "manual_step.update", "manual_step.matches", "manual_step.restore_time", "default_clock.time", "default_clock.clear", "default_clock.span", "default_clock.clock_time", "default_clock.remaining", "default_clock.update", "default_clock.matches"}

func (p Project) ProbePlan(ctx context.Context) (plan M, err error) {
	defer captureError(&err)
	data, e := readEvidenceFile(filepath.Join(p.Root, asString(probeNativeCandidateEvidence["capture"])), 1<<20)
	if e != nil {
		return nil, e
	}
	if sha256hex(data) != probeNativeCandidateEvidence["sha256"] {
		reject("native candidate origin integrity failure: %s", probeNativeCandidateEvidence["capture"])
	}
	lock, e := p.Lock("sources.json")
	if e != nil {
		return nil, e
	}
	versionBytes, e := os.ReadFile(p.InputPath("blizzard_version", "version.txt"))
	if e != nil {
		return nil, e
	}
	versions := asList(asMap(lock["blizzard_version"])["files"])
	if len(versions) != 1 || blobSHA(versionBytes) != asMap(versions[0])["sha"] {
		reject("modified probe client-version source")
	}
	documents, e := p.nativeDocuments()
	if e != nil {
		return nil, e
	}
	docs := map[string]M{}
	for _, doc := range documents {
		if _, exists := nativeFamilies[asString(doc.Metadata["Name"])]; exists {
			docs[asString(doc.Metadata["Name"])] = doc.Metadata
		}
	}
	families := []any{}
	count := 0
	for _, name := range sortedKeys(nativeFamilies) {
		definition := nativeFamilies[name]
		methods := []string{}
		for _, item := range asList(docs[name]["Functions"]) {
			methods = append(methods, asString(asMap(item)["Name"]))
		}
		sort.Strings(methods)
		if len(evSortedUnique(methods)) != definition.count || len(methods) != definition.count {
			reject("probe method inventory changed: %s", name)
		}
		count += len(methods)
		families = append(families, M{"name": name, "canonical": definition.canonical, "methods": evList(methods)})
	}
	if count != 73 {
		reject("probe requires review of changed native inventory")
	}
	declarations := M{}
	for _, name := range []string{"wow.d.tl", "wow_env.d.tl", "wow_secure_env.d.tl", "native_values.d.tl"} {
		data, e := os.ReadFile(filepath.Join(p.Root, "types/wow", name))
		if e != nil {
			return nil, e
		}
		declarations[name] = sha256hex(data)
	}
	c := newCatalogue(p)
	c.loadQuestie()
	c.loadNative()
	c.loadFrameXML()
	c.applyOverrides()
	c.reconcileInheritance()
	var inherited func(string, map[string]bool) map[string]bool
	inherited = func(owner string, seen map[string]bool) map[string]bool {
		if seen[owner] || len(seen) > 32 {
			reject("cyclic or unbounded probe hierarchy")
		}
		if ctx.Err() != nil {
			reject("probe plan: %s", ctx.Err())
		}
		definition := asMap(c.defs[owner])
		if definition == nil {
			reject("missing probe hierarchy owner: %s", owner)
		}
		methods := map[string]bool{}
		for name := range asMap(definition["methods"]) {
			methods[name] = true
		}
		next := capUnion(seen)
		next[owner] = true
		for _, base := range evStrings(definition["bases"]) {
			for name := range inherited(base, next) {
				methods[name] = true
			}
		}
		if len(methods) > 512 {
			reject("probe hierarchy exceeds method bound")
		}
		return methods
	}
	hierarchy := []any{}
	for _, row := range [][2]string{{"RadialProgress", "Animation"}, {"VectorGraphics", "Region"}} {
		hierarchy = append(hierarchy, M{"canonical": row[0], "parent": row[1], "methods": evList(sortedKeys(inherited(row[1], map[string]bool{})))})
	}
	hierarchyLock, hierarchyFiles, e := p.loadNativeSecondary()
	if e != nil {
		return nil, e
	}
	facetsValue, e := safeYAML([]byte(hierarchyFiles["wowless/data/products/wow/uiobjects.yaml"]))
	if e != nil {
		return nil, e
	}
	facets := asMap(facetsValue)
	var facetMethods func(string, map[string]bool) map[string]bool
	facetMethods = func(owner string, seen map[string]bool) map[string]bool {
		if seen[owner] || len(seen) > 32 {
			reject("cyclic or unbounded facet hierarchy")
		}
		definition := asMap(facets[owner])
		if definition == nil {
			reject("missing probe facet: %s", owner)
		}
		methods := map[string]bool{}
		for name := range asMap(definition["methods"]) {
			methods[name] = true
		}
		parents := []string{}
		switch value := definition["inherits"].(type) {
		case nil:
		case string:
			parents = []string{value}
		case M:
			for name, v := range value {
				if v != nil {
					reject("unsupported probe facet inheritance")
				}
				parents = append(parents, name)
			}
		default:
			reject("unsupported probe facet inheritance")
		}
		next := capUnion(seen)
		next[owner] = true
		for _, parent := range parents {
			for name := range facetMethods(parent, next) {
				methods[name] = true
			}
		}
		if len(methods) > 512 {
			reject("probe facet exceeds method bound")
		}
		return methods
	}
	for _, parent := range []string{"LayeredRegion", "ScriptObject"} {
		hierarchy = append(hierarchy, M{"canonical": "VectorGraphics", "parent": "wowless." + parent, "methods": evList(sortedKeys(facetMethods(parent, map[string]bool{})))})
	}
	hierarchy = append(hierarchy, M{"canonical": "VectorGraphics", "parent": "ketho.RotationCandidates", "methods": cloneValue(probeCandidateEvidence["methods"])})
	for _, row := range probeNativeCandidates {
		hierarchy = append(hierarchy, cloneValue(row))
	}
	for _, row := range probeRepresentatives {
		hierarchy = append(hierarchy, M{"canonical": asMap(row)["canonical"], "parent": "native.RepresentativeCapabilities", "methods": cloneValue(probeRepresentativeMethods)})
	}
	selected, e := consumer.ReadVersion(p.Root)
	if e != nil {
		return nil, e
	}
	buildVersion := strings.TrimSpace(string(versionBytes))
	if buildVersion != selected.Full {
		reject("probe source version differs from .wow-version")
	}
	hierarchyCases := 0
	for _, row := range hierarchy {
		hierarchyCases += len(asList(asMap(row)["methods"]))
	}
	fixture, e := os.ReadFile(filepath.Join(p.Root, "scripts/wow_api/probe/Fixture.svg"))
	if e != nil {
		return nil, e
	}
	sourceFingerprint := evDigest(M{"catalogue": lock, "hierarchy": hierarchyLock, "candidates": probeCandidateEvidence, "nativeCandidates": probeNativeCandidateEvidence, "nativeRequirements": probeNativeCandidates, "representatives": probeRepresentatives, "representativeMethods": probeRepresentativeMethods, "bindingDiagnosticCases": probeDiagnosticCases, "bindingTextSampleCases": probeSampleCases, "bindingTextSampleBytes": 128, "bindingExperimentCases": probeExperimentCases})
	return M{"version": probeVersion, "expectedBuild": buildVersion, "expectedInterface": selected.Interface, "expectedProjectID": 1,
		"limits":        M{"families": 7, "methods": 73, "queryCases": 56, "savedCaptures": 1, "hierarchyCases": hierarchyCases, "invocationCases": count, "returnTuple": 8, "frame": 1, "animationGroup": 1, "radialProgress": 1, "vectorGraphics": 1, "texture": 1, "durationBinding": 1, "manualClock": 1, "localeContext": 1, "timedSignalMap": 1, "invocationBindingCopies": 1, "fontString": 1, "duration": 1, "formatter": 1, "colorCurve": 1, "retainedUIHandles": 6, "presencePeakHandles": 10, "invocationPeakHandles": 14, "representativeCases": 12, "bindingDiagnosticCases": 25, "additionalNativeCalls": 31, "bindingTextSampleCases": 7, "bindingTextSampleBytes": 128, "bindingExperimentCases": 15},
		"fixtureSHA256": sha256hex(fixture), "sourceFingerprint": sourceFingerprint, "declarationFingerprint": evDigest(declarations), "candidateEvidence": cloneValue(probeCandidateEvidence), "nativeCandidateEvidence": cloneValue(probeNativeCandidateEvidence), "representatives": cloneValue(probeRepresentatives), "bindingDiagnosticCases": cloneValue(probeDiagnosticCases), "bindingTextSampleCases": cloneValue(probeSampleCases), "bindingExperimentCases": cloneValue(probeExperimentCases), "expectedMethods": count, "families": families, "hierarchy": hierarchy}, nil
}
func probeLiteral(value any) string {
	switch x := value.(type) {
	case string:
		data, e := jsonCompact(x)
		if e != nil {
			reject("invalid probe plan string")
		}
		return string(data)
	case []any:
		items := []string{}
		for _, row := range x {
			items = append(items, probeLiteral(row))
		}
		return "{" + strings.Join(items, ", ") + "}"
	case M:
		items := []string{}
		for _, name := range sortedKeys(x) {
			if !identifier(name) {
				reject("invalid probe plan field: %s", name)
			}
			items = append(items, name+" = "+probeLiteral(x[name]))
		}
		return "{" + strings.Join(items, ", ") + "}"
	case int:
		return fmt.Sprint(x)
	case int64:
		return fmt.Sprint(x)
	}
	reject("unsupported generated probe-plan value")
	return ""
}
func checkExistingProbe(output string) error {
	info, err := os.Lstat(output)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("probe output exists without an ownership manifest: %s", output)
	}
	manifest, err := readJSON(filepath.Join(output, "probe-build.json"))
	if err != nil {
		return fmt.Errorf("probe output exists without an ownership manifest: %s", output)
	}
	files := asMap(manifest["files"])
	current := map[string]bool{}
	for name := range files {
		current[name] = true
	}
	full := map[string]bool{}
	for _, name := range probeRuntimeFiles {
		full[name] = true
	}
	old := map[string]bool{}
	for _, name := range probePreviousRuntimeFiles {
		old[name] = true
	}
	if !capEqualSet(current, full) && !capEqualSet(current, old) {
		return fmt.Errorf("probe output contains unowned files")
	}
	actual := map[string]bool{}
	entries := 0
	err = filepath.WalkDir(output, func(path string, entry os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		entries++
		if entries > 64 {
			return fmt.Errorf("probe output contains unowned files")
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("probe output contains a symlink")
		}
		if entry.IsDir() {
			return nil
		}
		info, e := entry.Info()
		if e != nil {
			return e
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("probe output contains nonregular file")
		}
		relative, e := filepath.Rel(output, path)
		if e != nil {
			return e
		}
		actual[filepath.ToSlash(relative)] = true
		return nil
	})
	if err != nil {
		return err
	}
	expected := capUnion(current)
	expected["probe-build.json"] = true
	if !capEqualSet(actual, expected) {
		return fmt.Errorf("probe output contains unowned files")
	}
	for name, expected := range files {
		data, err := readEvidenceFile(filepath.Join(output, name), 2<<20)
		if err != nil {
			return err
		}
		if sha256hex(data) != expected {
			return fmt.Errorf("probe output has local edits: %s", name)
		}
	}
	return nil
}
func probeCommand(ctx context.Context, compiler string, args []string, cwd string) error {
	child, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()
	command := exec.CommandContext(child, compiler, args...)
	command.Dir = cwd
	environment := []string{}
	for _, value := range os.Environ() {
		if !strings.HasPrefix(value, "TL_PATH=") {
			environment = append(environment, value)
		}
	}
	patterns := []string{}
	for index, value := range args {
		if value == "-I" && index+1 < len(args) {
			directory := args[index+1]
			if strings.ContainsAny(directory, ";?\n\r") {
				return fmt.Errorf("unsupported Teal search-path character: %s", directory)
			}
			patterns = append(patterns, filepath.ToSlash(directory)+"/?.lua", filepath.ToSlash(directory)+"/?/init.lua")
		}
	}
	command.Env = append(environment, "TL_PATH="+strings.Join(patterns, ";"))
	data, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("probe compiler: %w: %s", err, strings.TrimSpace(string(data)))
	}
	return nil
}

// BuildProbe writes a checked, owned probe package; it never installs the addon.
func (p Project) BuildProbe(ctx context.Context, output, compiler string) (result M, err error) {
	defer captureError(&err)
	if output == "" {
		output = filepath.Join(p.Root, "build/generated/wow-api-probe")
	}
	output, e := filepath.Abs(output)
	if e != nil {
		return nil, e
	}
	if e = checkExistingProbe(output); e != nil {
		return nil, e
	}
	if compiler == "" {
		compiler = os.Getenv("TL")
		if compiler == "" {
			compiler = "tl"
		}
	}
	compiler, e = exec.LookPath(compiler)
	if e != nil {
		return nil, e
	}
	compiler, e = filepath.Abs(compiler)
	if e != nil {
		return nil, e
	}
	versionContext, cancelVersion := context.WithTimeout(ctx, 60*time.Second)
	defer cancelVersion()
	command := exec.CommandContext(versionContext, compiler, "--version")
	version, e := command.Output()
	if e != nil {
		return nil, e
	}
	if strings.TrimSpace(string(version)) != "0.24.8" {
		return nil, fmt.Errorf("probe requires Teal 0.24.8")
	}
	plan, e := p.ProbePlan(ctx)
	if e != nil {
		return nil, e
	}
	if e = os.MkdirAll(filepath.Dir(output), 0755); e != nil {
		return nil, e
	}
	stage, e := os.MkdirTemp(filepath.Dir(output), "wow-api-probe-")
	if e != nil {
		return nil, e
	}
	cleanupStage := true
	defer func() {
		if cleanupStage {
			os.RemoveAll(stage)
		}
	}()
	addon := filepath.Join(stage, "WoWAPIProbe")
	if e = os.Mkdir(addon, 0755); e != nil {
		return nil, e
	}
	source := filepath.Join(p.Root, "scripts/wow_api/probe")
	planSource := filepath.Join(stage, "Plan.tl")
	body := "local type T = require(\"probe_types\")\nlocal _, private = ... as (string, T.Namespace)\nprivate.Plan = " + probeLiteral(plan) + "\n"
	if e = os.WriteFile(planSource, []byte(body), 0644); e != nil {
		return nil, e
	}
	environment, e := os.ReadFile(filepath.Join(source, "env.d.tl"))
	if e != nil {
		return nil, e
	}
	environment = append([]byte("require(\"wow_env\")\n"), environment...)
	if e = os.WriteFile(filepath.Join(stage, "probe_env.d.tl"), environment, 0644); e != nil {
		return nil, e
	}
	args := []string{"--feat-arity", "on", "--gen-target", "5.1", "--gen-compat", "off", "-I", filepath.Join(p.Root, "types/wow"), "-I", source, "-I", stage, "--global-env-def", "probe_env"}
	for _, row := range [][2]string{{planSource, "Plan.lua"}, {filepath.Join(source, "Runtime.tl"), "Runtime.lua"}, {filepath.Join(source, "Invocations.tl"), "Invocations.lua"}} {
		if e = probeCommand(ctx, compiler, append(append([]string{}, args...), "run", filepath.Join(p.Root, "tools/check_secrets.lua"), "--", "probe_env", row[0], "wow_env"), stage); e != nil {
			return nil, e
		}
		if e = probeCommand(ctx, compiler, append(append([]string{}, args...), "gen", "--check", "-o", filepath.Join(addon, row[1]), row[0]), stage); e != nil {
			return nil, e
		}
		generated, e := os.ReadFile(filepath.Join(addon, row[1]))
		if e != nil {
			return nil, e
		}
		for _, module := range []string{"wow", "probe_types", "compat53"} {
			if strings.Contains(string(generated), "require(\""+module+"\")") {
				return nil, fmt.Errorf("probe generated an unwanted runtime type dependency")
			}
		}
	}
	toc := fmt.Sprintf("## Interface: %d\n## Title: WoW API Capability Probe\n## Version: %s\n## LoadOnDemand: 1\n## SavedVariables: WoWAPIProbeDB\nPlan.lua\nRuntime.lua\nInvocations.lua\n", asInt(plan["expectedInterface"]), probeVersion)
	if e = os.WriteFile(filepath.Join(addon, "WoWAPIProbe.toc"), []byte(toc), 0644); e != nil {
		return nil, e
	}
	fixture, e := os.ReadFile(filepath.Join(source, "Fixture.svg"))
	if e != nil {
		return nil, e
	}
	if e = os.WriteFile(filepath.Join(addon, "Fixture.svg"), fixture, 0644); e != nil {
		return nil, e
	}
	files := M{}
	for _, name := range probeRuntimeFiles {
		data, e := os.ReadFile(filepath.Join(stage, name))
		if e != nil {
			return nil, e
		}
		files[name] = sha256hex(data)
	}
	manifest, e := jsonPretty(M{"probe": probeVersion, "plan": plan, "files": files})
	if e != nil {
		return nil, e
	}
	publication := filepath.Join(stage, "publication")
	if e = os.Mkdir(publication, 0755); e != nil {
		return nil, e
	}
	if e = os.Rename(addon, filepath.Join(publication, "WoWAPIProbe")); e != nil {
		return nil, e
	}
	if e = os.WriteFile(filepath.Join(publication, "probe-build.json"), manifest, 0644); e != nil {
		return nil, e
	}
	// Publication takes ownership of the stage, including any recoverable backup.
	cleanupStage = false
	if e = publishProbe(stage, output, os.Rename); e != nil {
		return nil, e
	}
	return M{"output": output, "methods": plan["expectedMethods"], "families": len(asList(plan["families"])), "installed": false}, nil
}

// publishProbe owns stage and removes it unless restoration fails. In that case,
// the previous output remains at the recovery path reported by the error.
func publishProbe(stage, output string, rename func(string, string) error) error {
	cleanup := true
	defer func() {
		if cleanup {
			os.RemoveAll(stage)
		}
	}()
	if err := checkExistingProbe(output); err != nil {
		return err
	}
	backup := filepath.Join(stage, "previous")
	exists := false
	if _, err := os.Lstat(output); err == nil {
		if err = rename(output, backup); err != nil {
			return fmt.Errorf("preserve previous probe: %w", err)
		}
		exists = true
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := rename(filepath.Join(stage, "publication"), output); err != nil {
		if exists {
			if restore := rename(backup, output); restore != nil {
				cleanup = false
				return fmt.Errorf("publish probe failed; previous probe retained at %s: %w", backup, errors.Join(err, fmt.Errorf("restore previous probe: %w", restore)))
			}
		}
		return fmt.Errorf("publish probe: %w", err)
	}
	return nil
}
