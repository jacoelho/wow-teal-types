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

// TealVersion is the exact stock Teal release the checker and generated
// declarations are written against. The compiler is pinned by version, not bytes.
const TealVersion = "0.24.8"

// tealInstall is the remedy named whenever the external compiler is unusable.
const tealInstall = "luarocks install tl " + TealVersion

var modules = []string{"native_values", "wow", "wow_env", "wow_secure_env"}
var excluded = map[string]bool{".git": true, ".venv": true, "__pycache__": true, "build": true, "dist": true, "vendor": true, "node_modules": true}
var moduleName = regexp.MustCompile(`^[A-Za-z_][A-Za-z_0-9]*(\.[A-Za-z_][A-Za-z_0-9]*)*$`)

// RequiredFiles is the complete immutable consumer execution closure. The
// producer, CDN extractor, test fixtures and the external compiler are not part of it.
func RequiredFiles() []string {
	names := []string{".wow-version", "bin/wow-teal", "go.mod", "tools/check_secrets.lua", "tools/check_runtime_imports.lua", "tools/wow-teal/main.go", "types/wow/call_policy.lua",
		"internal/consumer/consumer.go", "internal/consumer/schema.go", "internal/consumer/files.go", "internal/consumer/version.go", "internal/consumer/snapshot.go", "internal/consumer/rockspec.go", "internal/consumer/package.go"}
	for _, name := range modules {
		names = append(names, "types/wow/"+name+".d.tl")
	}
	sort.Strings(names)
	return names
}

// Verify checks the package manifest, every file it binds and the selected
// WoW version without copying anything.
func Verify(root string) (PackageManifest, error) {
	verified, err := readPackage(context.Background(), root, nil)
	return verified.manifest, err
}

// Selection names caller-owned Teal sources for Check and Build.
type Selection struct {
	// Sources are Teal files or directories; Build requires exactly one directory.
	Sources []string
	// Include adds dependency search roots captured with the sources.
	Include []string
	// Global names an addon declaration module merged with the native environment.
	Global string
	// Secure selects the secure native environment.
	Secure bool
}

// CheckTypes type-checks the package's own declaration modules.
func CheckTypes(ctx context.Context, root string, out, errout io.Writer) error {
	root, err := canonicalExisting(root)
	if err != nil {
		return err
	}
	inputs, err := captureBuildInputs(ctx, root, nil, nil)
	if err != nil {
		return err
	}
	defer inputs.close()
	tl, err := resolveCompiler(ctx, root, inputs, errout)
	if err != nil {
		return err
	}
	for _, name := range modules {
		fmt.Fprintln(out, "Checking "+filepath.Join(root, "types/wow", name+".d.tl"))
		if _, err = tl.run(append(compilerFlags(inputs.search), "check", filepath.Join(inputs.packageRoot, "types/wow", name+".d.tl"))...); err != nil {
			return err
		}
	}
	fmt.Fprintf(out, "Passed declaration checks for %d files.\n", len(modules))
	return nil
}

// Check type-checks and secret-checks the selected sources.
func Check(ctx context.Context, root string, selected Selection, out, errout io.Writer) error {
	root, err := canonicalExisting(root)
	if err != nil {
		return err
	}
	session, err := openSession(ctx, root, selected, errout)
	if err != nil {
		return err
	}
	defer session.inputs.close()
	if err = session.checkSources(out); err != nil {
		return err
	}
	fmt.Fprintf(out, "Passed type and secret checks for %d selected source files.\n", len(session.sources))
	return nil
}

// Build checks one source directory and atomically replaces the owned output
// directory with its generated Lua.
func Build(ctx context.Context, root string, selected Selection, output string, out, errout io.Writer) error {
	root, err := canonicalExisting(root)
	if err != nil {
		return err
	}
	if output == "" {
		return fmt.Errorf("build requires --output")
	}
	if len(selected.Sources) != 1 {
		return fmt.Errorf("build requires one source directory")
	}
	info, err := os.Stat(selected.Sources[0])
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("build requires one source directory")
	}
	sourceRoot, err := canonicalExisting(selected.Sources[0])
	if err != nil {
		return err
	}
	destination, err := canonicalOutput(output)
	if err != nil {
		return err
	}
	if inside(destination, sourceRoot) || inside(sourceRoot, destination) || inside(root, destination) {
		return fmt.Errorf("output must not overlap source or package roots")
	}
	if err = ownedOutput(destination); err != nil {
		return err
	}
	session, err := openSession(ctx, root, selected, errout)
	if err != nil {
		return err
	}
	defer session.inputs.close()
	if err = session.checkSources(out); err != nil {
		return err
	}
	count, err := session.build(sourceRoot, destination)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "Generated %d Lua files in %s\n", count, destination)
	fmt.Fprintf(out, "Passed type and secret checks for %d selected source files.\n", len(session.sources))
	return nil
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

// compiler runs one external Teal executable inside a captured workspace.
type compiler struct {
	ctx         context.Context
	path        string
	work        string
	environment []string
	paths       *strings.Replacer
	errout      io.Writer
}

// resolveCompiler locates TL, else tl on PATH, and admits it only when its
// version output is exactly TealVersion. Callers capture and verify package
// inputs first, so no compiler runs against unverified bytes.
func resolveCompiler(ctx context.Context, root string, inputs *buildInputs, errout io.Writer) (compiler, error) {
	name := os.Getenv("TL")
	if name == "" {
		name = "tl"
	}
	path, err := exec.LookPath(name)
	if err != nil {
		return compiler{}, fmt.Errorf("locate Teal compiler (set TL or run %s): %w", tealInstall, err)
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return compiler{}, err
	}
	patterns := []string{}
	for _, directory := range inputs.search {
		if strings.ContainsAny(directory, ";?\n\r") {
			return compiler{}, fmt.Errorf("unsupported Teal search-path character: %s", directory)
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
	c := compiler{
		ctx:         ctx,
		path:        path,
		work:        inputs.directory,
		environment: environment,
		paths:       strings.NewReplacer(inputs.packageRoot, root, filepath.Join(inputs.directory, "inputs")+string(filepath.Separator), string(filepath.Separator)),
		errout:      errout,
	}
	version, err := c.run("--version")
	if err != nil {
		return compiler{}, fmt.Errorf("unsupported Teal version: %s --version failed; expected %s (%s): %w", path, TealVersion, tealInstall, err)
	}
	if err = admitTealVersion(version); err != nil {
		return compiler{}, err
	}
	return c, nil
}

// admitTealVersion accepts only TealVersion, ignoring surrounding whitespace.
// The checker reads compiler internals, so a prefix or suffix match is unsafe.
func admitTealVersion(output string) error {
	if strings.TrimSpace(output) != TealVersion {
		return fmt.Errorf("unsupported Teal version: %q; expected %s (%s)", output, TealVersion, tealInstall)
	}
	return nil
}

func (c compiler) run(args ...string) (string, error) {
	child, cancel := context.WithTimeout(c.ctx, 60*time.Second)
	defer cancel()
	process := exec.CommandContext(child, c.path, args...)
	process.Dir = c.work
	process.Env = c.environment
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
		return "", fmt.Errorf("%s: %w", c.paths.Replace(detail), err)
	}
	if stderr.Len() != 0 {
		if _, err = io.WriteString(c.errout, c.paths.Replace(stderr.String())); err != nil {
			return "", err
		}
	}
	return stdout.String(), nil
}

// compilerFlags are the language options of every consumer compiler invocation.
func compilerFlags(search []string) []string {
	flags := []string{"--feat-arity", "on", "--gen-target", "5.1", "--gen-compat", "off"}
	for _, directory := range search {
		flags = append(flags, "-I", directory)
	}
	return flags
}

// secretCheckArgs is the one checker invocation shape: the selected type
// environment, the source, and the native environment whose identity it pins.
func secretCheckArgs(search []string, checker, environment, source, native string) []string {
	return append(compilerFlags(search), "run", checker, "--", environment, source, native)
}

// session owns the captured inputs and admitted compiler for one Check or Build.
type session struct {
	root        string
	inputs      *buildInputs
	tl          compiler
	sources     []string // caller paths, sorted
	captured    []string // captured copies, parallel to sources
	environment string   // module selected as the global environment
	native      string
}

func openSession(ctx context.Context, root string, selected Selection, errout io.Writer) (_ *session, err error) {
	if len(selected.Sources) == 0 {
		return nil, fmt.Errorf("expected owned Teal sources")
	}
	if selected.Global != "" && !moduleName.MatchString(selected.Global) {
		return nil, fmt.Errorf("addon global environment must be a module name")
	}
	if slices.Contains(modules, selected.Global) {
		return nil, fmt.Errorf("use --secure to select the secure native environment")
	}
	sources, err := sourcesFor(selected.Sources)
	if err != nil {
		return nil, err
	}
	include, err := searchRoots(selected, sources)
	if err != nil {
		return nil, err
	}
	inputs, err := captureBuildInputs(ctx, root, sources, include)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			inputs.close()
		}
	}()
	s := &session{root: root, inputs: inputs, sources: sources, native: "wow_env"}
	if selected.Secure {
		s.native = "wow_secure_env"
	}
	s.environment = s.native
	if selected.Global != "" {
		s.environment = "wow_teal_combined_env"
		body := "require(\"" + s.native + "\")\nrequire(\"" + selected.Global + "\")\n"
		if err = os.WriteFile(filepath.Join(inputs.directory, "environment", s.environment+".d.tl"), []byte(body), 0644); err != nil {
			return nil, err
		}
	}
	for _, source := range sources {
		s.captured = append(s.captured, inputs.source(source))
	}
	if s.tl, err = resolveCompiler(ctx, root, inputs, errout); err != nil {
		return nil, err
	}
	return s, nil
}

// searchRoots orders explicit includes, selected directories, then the
// parents of selected files; earlier roots take precedence.
func searchRoots(selected Selection, sources []string) ([]string, error) {
	include := []string{}
	for _, name := range selected.Include {
		path, err := inputRoot(name)
		if err != nil {
			return nil, err
		}
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("include paths must be directories")
		}
		include = append(include, path)
	}
	for _, name := range selected.Sources {
		info, err := os.Stat(name)
		if err != nil {
			return nil, err
		}
		if info.IsDir() {
			path, err := canonicalExisting(name)
			if err != nil {
				return nil, err
			}
			include = append(include, path)
		}
	}
	parents := map[string]bool{}
	for _, source := range sources {
		parents[filepath.Dir(source)] = true
	}
	return append(include, slices.Sorted(maps.Keys(parents))...), nil
}

func (s *session) checkSources(out io.Writer) error {
	checker := filepath.Join(s.inputs.packageRoot, "tools/check_secrets.lua")
	for i, source := range s.captured {
		fmt.Fprintln(out, "Checking "+s.sources[i])
		if _, err := s.tl.run(secretCheckArgs(s.inputs.search, checker, s.environment, source, s.native)...); err != nil {
			return err
		}
	}
	return nil
}

// build generates Lua for every checked source, then publishes it with an
// ownership receipt. It returns the number of generated files.
func (s *session) build(sourceRoot, destination string) (int, error) {
	generated := filepath.Join(s.inputs.directory, "generated")
	if err := os.Mkdir(generated, 0755); err != nil {
		return 0, err
	}
	flags := compilerFlags(s.inputs.search)
	imports := filepath.Join(s.inputs.packageRoot, "tools/check_runtime_imports.lua")
	for i, source := range s.captured {
		relative, err := filepath.Rel(sourceRoot, s.sources[i])
		if err != nil {
			return 0, err
		}
		target := filepath.Join(generated, strings.TrimSuffix(relative, ".tl")+".lua")
		if err = os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return 0, err
		}
		if _, err = s.tl.run(append(slices.Clone(flags), "--global-env-def", s.environment, "gen", "--check", "-o", target, source)...); err != nil {
			return 0, err
		}
		if _, err = s.tl.run(append(slices.Clone(flags), "run", imports, "--", target)...); err != nil {
			return 0, err
		}
	}
	files := map[string]string{}
	err := filepath.WalkDir(generated, func(path string, entry os.DirEntry, e error) error {
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
		return 0, err
	}
	receipt, err := receiptBytes(BuildReceipt{Files: files, PackageManifestSHA256: s.inputs.manifestHash, PackageVersion: s.inputs.manifest.Version})
	if err != nil {
		return 0, err
	}
	if err = os.WriteFile(filepath.Join(generated, "wow-teal-build.json"), receipt, 0644); err != nil {
		return 0, err
	}
	if err = publishBuild(generated, destination, os.Rename); err != nil {
		return 0, err
	}
	return len(files), nil
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
