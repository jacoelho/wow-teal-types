// Package consumer owns immutable package checking and checked addon compilation.
package consumer

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"
)

const maxSources = 512
const maxEntries = 10000
const maxSourceBytes = 2 << 20

// TealVersion identifies the bundled compiler/checker contract, not stock Teal.
const TealVersion = "0.24.8-wow.1"

var modules = []string{"native_values", "wow", "wow_env", "wow_secure_env"}
var excluded = map[string]bool{".git": true, ".venv": true, "__pycache__": true, "build": true, "dist": true, "vendor": true, "node_modules": true}
var moduleName = regexp.MustCompile(`^[A-Za-z_][A-Za-z_0-9]*(\.[A-Za-z_][A-Za-z_0-9]*)*$`)

// RequiredFiles is the complete immutable consumer execution closure. The
// producer, CDN extractor, Rust reader and test fixtures are not runtime inputs.
func RequiredFiles() []string {
	names := []string{".wow-version", "bin/wow-teal", "go.mod", "go.sum", "provenance.json", "tools/check_secrets.lua", "tools/check_runtime_imports.lua", "tools/wow-teal/main.go",
		"tools/teal/tl", "tools/teal/cli.lua", "tools/teal/tl.lua", "tools/teal/argparse.lua", "tools/teal/contract.lua", "tools/teal/check_capabilities.lua", "tools/teal/LICENSE", "types/wow/call_policy.lua",
		"internal/consumer/consumer.go", "internal/consumer/schema.go", "internal/consumer/files.go", "internal/consumer/version.go", "internal/consumer/snapshot.go", "internal/consumer/rockspec.go", "internal/consumer/package.go"}
	for _, name := range modules {
		names = append(names, "types/wow/"+name+".d.tl")
	}
	sort.Strings(names)
	return names
}

func Verify(root string) (PackageManifest, error) {
	manifest, err := readManifest(filepath.Join(root, "package.json"))
	if err != nil {
		return PackageManifest{}, err
	}
	files := manifest.Files
	if !slices.Equal(slices.Sorted(maps.Keys(files)), RequiredFiles()) {
		return PackageManifest{}, fmt.Errorf("package manifest inventory mismatch")
	}
	for _, name := range slices.Sorted(maps.Keys(files)) {
		path, err := checkedPath(root, name, false)
		if err != nil {
			return PackageManifest{}, fmt.Errorf("package integrity mismatch: %s: %w", name, err)
		}
		actual, err := fileHash(path)
		if err != nil {
			return PackageManifest{}, err
		}
		if actual != files[name] {
			return PackageManifest{}, fmt.Errorf("package integrity mismatch: %s", name)
		}
	}
	selected, err := ReadVersion(root)
	if err != nil {
		return PackageManifest{}, err
	}
	if manifest.WoWBuild != selected.Full {
		return PackageManifest{}, fmt.Errorf("package WoW version differs from .wow-version")
	}
	return manifest, nil
}

type selection struct {
	command          string
	sources, include []string
	output, global   string
	secure           bool
}

func parseSelection(args []string) (selection, error) {
	if len(args) == 0 {
		return selection{}, fmt.Errorf("expected version, verify, check-types, check, build, package or rockspecs")
	}
	selected := selection{command: args[0]}
	for i := 1; i < len(args); i++ {
		a := args[i]
		switch a {
		case "--secure":
			selected.secure = true
		case "-I", "--include", "-o", "--output", "--global-env-def":
			if i+1 == len(args) {
				return selected, fmt.Errorf("missing argument for %s", a)
			}
			i++
			switch a {
			case "-I", "--include":
				selected.include = append(selected.include, args[i])
			case "-o", "--output":
				selected.output = args[i]
			default:
				selected.global = args[i]
			}
		default:
			if strings.HasPrefix(a, "-") {
				return selected, fmt.Errorf("unknown option: %s", a)
			}
			selected.sources = append(selected.sources, a)
		}
	}
	switch selected.command {
	case "verify", "check-types":
		if len(args) != 1 {
			return selected, fmt.Errorf("%s does not accept source arguments", selected.command)
		}
	case "check", "build":
		if len(selected.sources) == 0 {
			return selected, fmt.Errorf("expected owned Teal sources")
		}
		if selected.command == "build" && selected.output == "" {
			return selected, fmt.Errorf("build requires --output")
		}
	default:
		return selected, fmt.Errorf("unknown command: %s", selected.command)
	}
	return selected, nil
}

func Run(ctx context.Context, root string, args []string, out, errout io.Writer) error {
	var err error
	root, err = canonicalExisting(root)
	if err != nil {
		return err
	}
	if len(args) > 0 {
		switch args[0] {
		case "version":
			if len(args) != 1 {
				return fmt.Errorf("version takes no arguments")
			}
			manifest, err := Verify(root)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(out, manifest.Version)
			return err
		case "package":
			if len(args) != 1 {
				return fmt.Errorf("package takes no arguments")
			}
			value, err := Package(root)
			if err != nil {
				return err
			}
			return writePackageResult(out, value)
		case "rockspecs":
			return Rockspecs(root, args[1:], out, errout)
		}
	}
	selected, err := parseSelection(args)
	if err != nil {
		return err
	}
	return execute(ctx, root, selected, out, errout)
}

func sourcesFor(arguments []string) ([]string, error) {
	sources := map[string]bool{}
	entries := 0
	for _, argument := range arguments {
		argument = filepath.Clean(argument)
		info, err := os.Lstat(argument)
		if err != nil {
			return nil, err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("source selection must not be a symlink: %s", argument)
		}
		selected, err := canonicalExisting(argument)
		if err != nil {
			return nil, err
		}
		if info.IsDir() {
			err = filepath.WalkDir(selected, func(path string, entry os.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				if path == selected {
					return nil
				}
				entries++
				if entries > maxEntries {
					return fmt.Errorf("source traversal exceeds 10000 entries")
				}
				if entry.Type()&os.ModeSymlink != 0 {
					return fmt.Errorf("source must not be a symlink: %s", path)
				}
				if entry.IsDir() {
					if excluded[entry.Name()] {
						return filepath.SkipDir
					}
					return nil
				}
				if strings.HasSuffix(path, ".tl") && !strings.HasSuffix(path, ".d.tl") {
					info, err := entry.Info()
					if err != nil {
						return err
					}
					if !info.Mode().IsRegular() {
						return fmt.Errorf("source must be a regular file: %s", path)
					}
					sources[path] = true
					if len(sources) > maxSources {
						return fmt.Errorf("source selection exceeds 512 files")
					}
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("source must be a regular file: %s", selected)
		} else if strings.HasSuffix(selected, ".tl") && !strings.HasSuffix(selected, ".d.tl") {
			sources[selected] = true
		} else {
			return nil, fmt.Errorf("expected a Teal source file or directory: %s", selected)
		}
	}
	if len(sources) == 0 {
		return nil, fmt.Errorf("no Teal source files selected")
	}
	if len(sources) > maxSources {
		return nil, fmt.Errorf("source selection exceeds 512 files")
	}
	for path := range sources {
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if info.Size() > maxSourceBytes {
			return nil, fmt.Errorf("source exceeds 2 MiB: %s", path)
		}
	}
	return slices.Sorted(maps.Keys(sources)), nil
}

func ownedOutput(destination string) error {
	info, err := os.Lstat(destination)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("output must be an owned build directory")
	}
	receipt := filepath.Join(destination, "wow-teal-build.json")
	info, err = os.Lstat(receipt)
	if err != nil || !info.Mode().IsRegular() {
		return fmt.Errorf("existing output has no wow-teal ownership receipt")
	}
	value, err := readBuildReceipt(receipt)
	if err != nil {
		return err
	}
	files := value.Files
	actual := map[string]bool{}
	entries := 0
	err = filepath.WalkDir(destination, func(path string, entry os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		entries++
		if entries > maxEntries {
			return fmt.Errorf("output exceeds 10000 entries")
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("output contains a symlink")
		}
		if entry.IsDir() {
			return nil
		}
		info, e := entry.Info()
		if e != nil {
			return e
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("output contains a nonregular file")
		}
		relative, e := filepath.Rel(destination, path)
		if e != nil {
			return e
		}
		actual[filepath.ToSlash(relative)] = true
		return nil
	})
	if err != nil {
		return err
	}
	if len(actual) != len(files)+1 || !actual["wow-teal-build.json"] {
		return fmt.Errorf("output contains unowned files")
	}
	for name, expected := range files {
		if !validRelative(name) || !actual[name] {
			return fmt.Errorf("output contains unowned files")
		}
		hash, e := fileHash(filepath.Join(destination, filepath.FromSlash(name)))
		if e != nil {
			return e
		}
		if hash != expected {
			return fmt.Errorf("output has local changes: %s", name)
		}
	}
	return nil
}

func compilerRun(ctx context.Context, command []string, cwd string, environment []string, paths *strings.Replacer, errout io.Writer) (string, error) {
	child, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	process := exec.CommandContext(child, command[0], command[1:]...)
	process.Dir = cwd
	process.Env = environment
	var stdout, stderr bytes.Buffer
	process.Stdout = &stdout
	process.Stderr = &stderr
	err := process.Run()
	if err != nil {
		if child.Err() != nil {
			return "", fmt.Errorf("compiler command: %w", child.Err())
		}
		detail := strings.TrimSpace(stdout.String() + stderr.String())
		if detail == "" {
			detail = "compiler failed without diagnostics: " + err.Error()
		}
		return "", fmt.Errorf("%s: %w", paths.Replace(detail), err)
	}
	if stderr.Len() != 0 {
		if _, err = io.WriteString(errout, paths.Replace(stderr.String())); err != nil {
			return "", err
		}
	}
	return stdout.String(), nil
}

func execute(ctx context.Context, root string, args selection, out, errout io.Writer) error {
	if args.command == "verify" {
		manifest, err := Verify(root)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(out, "Verified wow-teal %s / Retail %s\n", manifest.Version, manifest.WoWBuild)
		return err
	}
	declarationCheck := args.command == "check-types"
	types := filepath.Join(root, "types/wow")
	var sources []string
	var err error
	if declarationCheck {
		for _, name := range modules {
			sources = append(sources, filepath.Join(types, name+".d.tl"))
		}
	} else {
		sources, err = sourcesFor(args.sources)
	}
	if err != nil {
		return err
	}
	destination, sourceRoot := "", ""
	if args.command == "build" {
		if len(args.sources) != 1 {
			return fmt.Errorf("build requires one source directory")
		}
		info, e := os.Stat(args.sources[0])
		if e != nil {
			return e
		}
		if !info.IsDir() {
			return fmt.Errorf("build requires one source directory")
		}
		sourceRoot, err = canonicalExisting(args.sources[0])
		if err != nil {
			return err
		}
		destination, err = canonicalOutput(args.output)
		if err != nil {
			return err
		}
		if inside(destination, sourceRoot) || inside(sourceRoot, destination) || inside(root, destination) {
			return fmt.Errorf("output must not overlap source or package roots")
		}
		if err = ownedOutput(destination); err != nil {
			return err
		}
	}
	include := []string{}
	for _, name := range args.include {
		path, e := inputRoot(name)
		if e != nil {
			return e
		}
		info, e := os.Stat(path)
		if e != nil {
			return e
		}
		if !info.IsDir() {
			return fmt.Errorf("include paths must be directories")
		}
		include = append(include, path)
	}
	if !declarationCheck {
		for _, name := range args.sources {
			info, e := os.Stat(name)
			if e != nil {
				return e
			}
			if info.IsDir() {
				path, e := canonicalExisting(name)
				if e != nil {
					return e
				}
				include = append(include, path)
			}
		}
	}
	parents := map[string]bool{}
	for _, source := range sources {
		parents[filepath.Dir(source)] = true
	}
	include = append(include, slices.Sorted(maps.Keys(parents))...)
	callerSources := sources
	if declarationCheck {
		callerSources = nil
		include = nil
	}
	inputs, err := captureBuildInputs(ctx, root, callerSources, include)
	if err != nil {
		return err
	}
	defer inputs.close()
	manifest := inputs.manifest
	work := inputs.directory
	capturedSources := make([]string, len(sources))
	for i, source := range sources {
		if declarationCheck {
			capturedSources[i] = filepath.Join(inputs.packageRoot, "types/wow", filepath.Base(source))
		} else {
			capturedSources[i] = inputs.source(source)
		}
	}
	compiler := os.Getenv("TL")
	if compiler == "" {
		compiler = filepath.Join(inputs.packageRoot, "tools/teal/tl")
	}
	compiler, err = exec.LookPath(compiler)
	if err != nil {
		return fmt.Errorf("locate bundled Teal compiler: %w", err)
	}
	compiler, err = filepath.Abs(compiler)
	if err != nil {
		return err
	}
	if args.global != "" && !moduleName.MatchString(args.global) {
		return fmt.Errorf("addon global environment must be a module name")
	}
	for _, name := range modules {
		if args.global == name {
			return fmt.Errorf("use --secure to select the secure native environment")
		}
	}
	native := "wow_env"
	if args.secure {
		native = "wow_secure_env"
	}
	selectedEnv := native
	if args.global != "" {
		selectedEnv = "wow_teal_combined_env"
		body := "require(\"" + native + "\")\nrequire(\"" + args.global + "\")\n"
		if err = os.WriteFile(filepath.Join(work, "environment", selectedEnv+".d.tl"), []byte(body), 0644); err != nil {
			return err
		}
	}
	search := inputs.search
	patterns := []string{}
	for _, directory := range search {
		if strings.ContainsAny(directory, ";?\n\r") {
			return fmt.Errorf("unsupported Teal search-path character: %s", directory)
		}
		patterns = append(patterns, filepath.ToSlash(directory)+"/?.lua", filepath.ToSlash(directory)+"/?/init.lua")
	}
	environment := []string{}
	for _, value := range os.Environ() {
		if !strings.HasPrefix(value, "TL_PATH=") {
			environment = append(environment, value)
		}
	}
	environment = append(environment, "TL_PATH="+strings.Join(patterns, ";"))
	diagnosticPaths := strings.NewReplacer(inputs.packageRoot, root, filepath.Join(work, "inputs")+string(filepath.Separator), string(filepath.Separator))
	run := func(command []string) (string, error) {
		return compilerRun(ctx, command, work, environment, diagnosticPaths, errout)
	}
	version, err := run([]string{compiler, "--version"})
	if err != nil {
		return err
	}
	if strings.TrimSpace(version) != manifest.TealVersion {
		return fmt.Errorf("unsupported Teal version: %s; expected %v", strings.TrimSpace(version), manifest.TealVersion)
	}
	if _, err = run([]string{compiler, "run", filepath.Join(inputs.packageRoot, "tools/teal/check_capabilities.lua")}); err != nil {
		return err
	}
	flags := []string{compiler, "--feat-arity", "on", "--gen-target", "5.1", "--gen-compat", "off"}
	for _, directory := range search {
		flags = append(flags, "-I", directory)
	}
	invoke := func(arguments ...string) error {
		_, e := run(append(append([]string{}, flags...), arguments...))
		return e
	}
	for i, source := range capturedSources {
		fmt.Fprintln(out, "Checking "+sources[i])
		if declarationCheck {
			err = invoke("check", source)
		} else {
			err = invoke("run", filepath.Join(inputs.packageRoot, "tools/check_secrets.lua"), "--", selectedEnv, source, native)
		}
		if err != nil {
			return err
		}
	}
	generated := filepath.Join(work, "generated")
	if err = os.Mkdir(generated, 0755); err != nil {
		return err
	}
	if destination != "" {
		for i, source := range capturedSources {
			relative, e := filepath.Rel(sourceRoot, sources[i])
			if e != nil {
				return e
			}
			target := filepath.Join(generated, strings.TrimSuffix(relative, ".tl")+".lua")
			if err = os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			if err = invoke("--global-env-def", selectedEnv, "gen", "--check", "-o", target, source); err != nil {
				return err
			}
			if err = invoke("run", filepath.Join(inputs.packageRoot, "tools/check_runtime_imports.lua"), "--", target); err != nil {
				return err
			}
		}
	}
	if destination != "" {
		files := map[string]string{}
		err = filepath.WalkDir(generated, func(path string, entry os.DirEntry, e error) error {
			if e != nil {
				return e
			}
			if entry.IsDir() {
				return nil
			}
			relative, e := filepath.Rel(generated, path)
			if e != nil {
				return e
			}
			hash, e := fileHash(path)
			if e != nil {
				return e
			}
			files[filepath.ToSlash(relative)] = hash
			return nil
		})
		if err != nil {
			return err
		}
		receipt, err := receiptBytes(BuildReceipt{Files: files, PackageManifestSHA256: inputs.manifestHash, PackageVersion: manifest.Version})
		if err != nil {
			return err
		}
		if err = os.WriteFile(filepath.Join(generated, "wow-teal-build.json"), receipt, 0644); err != nil {
			return err
		}
		if err = publishBuild(generated, destination, os.Rename); err != nil {
			return err
		}
		fmt.Fprintf(out, "Generated %d Lua files in %s\n", len(files), destination)
	}
	if declarationCheck {
		fmt.Fprintf(out, "Passed declaration checks for %d files.\n", len(sources))
	} else {
		fmt.Fprintf(out, "Passed type and secret checks for %d selected source files.\n", len(sources))
	}
	return nil
}

func publishBuild(generated, destination string, rename func(string, string) error) error {
	if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
		return err
	}
	publication, err := os.MkdirTemp(filepath.Dir(destination), ".wow-teal-publish-")
	if err != nil {
		return err
	}
	retain := false
	defer func() {
		if !retain {
			_ = os.RemoveAll(publication)
		}
	}()
	stage := filepath.Join(publication, "next")
	if err = copyTree(generated, stage); err != nil {
		return err
	}
	if err = ownedOutput(destination); err != nil {
		return err
	}
	previous := filepath.Join(publication, "previous")
	exists := false
	if _, err = os.Lstat(destination); err == nil {
		if err = rename(destination, previous); err != nil {
			return err
		}
		exists = true
	} else if !os.IsNotExist(err) {
		return err
	}
	if err = rename(stage, destination); err != nil {
		if exists {
			if restore := rename(previous, destination); restore != nil {
				retain = true
				return errors.Join(err, fmt.Errorf("restore previous build (backup retained at %s): %w", previous, restore))
			}
		}
		return err
	}
	return nil
}
