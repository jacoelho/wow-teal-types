package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"wow-teal-types/internal/consumer"
	"wow-teal-types/internal/release"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if e := run(ctx, os.Args[1:]); e != nil {
		fmt.Fprintln(os.Stderr, "wow-types:", e)
		os.Exit(1)
	}
}

// globalOptions precede the command and apply to every command.
var globalOptions = []string{"root", "cwd"}

// invocation is a validated command line: options are restricted to the
// command's table entry and caller-relative paths are already absolute.
type invocation struct {
	root        string
	options     map[string]string
	positionals []string
}

type command struct {
	name           string
	options        []string
	maxPositionals int
	run            func(context.Context, invocation) error
}

// commands is the only definition of the command line: dispatch, option
// validation and usage all derive from it.
var commands = []command{
	{name: "generate", run: func(ctx context.Context, in invocation) error {
		return writeResult(release.Update(ctx, in.root))
	}},
	{name: "check-generator", options: []string{"client-source", "snapshot"}, run: func(ctx context.Context, in invocation) error {
		snapshot, err := in.snapshot()
		if err != nil {
			return err
		}
		return writeResult(release.CheckGenerator(ctx, in.root, snapshot))
	}},
	{name: "bootstrap", options: []string{"client-source", "snapshot"}, run: func(ctx context.Context, in invocation) error {
		snapshot, err := in.snapshot()
		if err != nil {
			return err
		}
		_, err = release.Bootstrap(ctx, in.root, snapshot)
		return writeResult(struct {
			Snapshot string `json:"snapshot"`
			Status   string `json:"status"`
		}{snapshot, "verified"}, err)
	}},
	{name: "acquire-client", options: []string{"client-source", "snapshot", "cache"}, run: func(ctx context.Context, in invocation) error {
		snapshot, err := in.snapshot()
		if err != nil {
			return err
		}
		return writeResult(release.Acquire(ctx, in.root, snapshot, in.options["cache"]))
	}},
	{name: "stage-client", options: []string{"client-source", "snapshot", "cache", "region"}, run: func(ctx context.Context, in invocation) error {
		snapshot, err := in.snapshot()
		if err != nil {
			return err
		}
		return writeResult(release.StageSelected(ctx, in.root, snapshot, in.options["cache"], in.options["region"]))
	}},
	{name: "verify-client", options: []string{"client-source", "snapshot"}, maxPositionals: 1, run: func(ctx context.Context, in invocation) error {
		snapshot, err := in.snapshot()
		if err != nil {
			return err
		}
		return writeResult(release.VerifySnapshot(snapshot, nil))
	}},
	{name: "stage-release", options: []string{"client-source", "snapshot", "output", "questie-ref"}, run: func(ctx context.Context, in invocation) error {
		snapshot, err := in.snapshot()
		if err != nil {
			return err
		}
		output := in.options["output"]
		if output == "" {
			selected, err := consumer.ReadVersion(in.root)
			if err != nil {
				return err
			}
			output = filepath.Join(in.root, "build/wow-api-candidates", selected.Full)
		}
		return writeResult(release.StageRelease(ctx, in.root, output, snapshot, in.options["questie-ref"]))
	}},
	{name: "build-extractor", run: func(ctx context.Context, in invocation) error {
		binary, err := release.BuildExtractor(ctx, in.root)
		return writeResult(struct {
			Entry string `json:"entry"`
		}{binary}, err)
	}},
	{name: "recover", run: func(ctx context.Context, in invocation) error {
		return writeResult(release.Recover(in.root))
	}},
	{name: "version", run: func(ctx context.Context, in invocation) error {
		manifest, err := consumer.Verify(in.root)
		if err != nil {
			return err
		}
		fmt.Println(manifest.Version)
		return nil
	}},
}

func usage() string {
	var out strings.Builder
	out.WriteString("usage: wow-types [--root DIR] [--cwd DIR] COMMAND [OPTIONS]\ncommands:")
	for _, cmd := range commands {
		out.WriteString("\n  " + cmd.name)
		if cmd.maxPositionals > 0 {
			out.WriteString(" [SNAPSHOT]")
		}
		for _, option := range cmd.options {
			out.WriteString(" [--" + option + " VALUE]")
		}
	}
	return out.String()
}

func run(ctx context.Context, args []string) error {
	global, args, err := parseOptions(args, globalOptions, true)
	if err != nil {
		return err
	}
	root := global["root"]
	if root == "" {
		if root, err = os.Getwd(); err != nil {
			return err
		}
	}
	if root, err = filepath.Abs(root); err != nil {
		return err
	}
	if cwd := global["cwd"]; cwd != "" {
		if err = os.Chdir(cwd); err != nil {
			return err
		}
	}
	if len(args) == 0 {
		return fmt.Errorf("expected a command\n%s", usage())
	}
	cmd, err := lookupCommand(args[0])
	if err != nil {
		return err
	}
	options, positionals, err := parseOptions(args[1:], cmd.options, false)
	if err != nil {
		return fmt.Errorf("%s: %w", cmd.name, err)
	}
	if err = validatePositionals(cmd, positionals); err != nil {
		return err
	}
	if err = resolvePaths(options, positionals); err != nil {
		return err
	}
	return cmd.run(ctx, invocation{root: root, options: options, positionals: positionals})
}

func lookupCommand(name string) (command, error) {
	for _, cmd := range commands {
		if cmd.name == name {
			return cmd, nil
		}
	}
	return command{}, fmt.Errorf("unknown command %q\n%s", name, usage())
}

func validatePositionals(cmd command, positionals []string) error {
	if len(positionals) <= cmd.maxPositionals {
		return nil
	}
	if cmd.maxPositionals == 0 {
		return fmt.Errorf("%s does not accept positional arguments", cmd.name)
	}
	return fmt.Errorf("%s accepts at most %d positional argument(s)", cmd.name, cmd.maxPositionals)
}

// snapshot selects the client snapshot: a positional argument, then
// --client-source or --snapshot, then the build named by .wow-version.
func (in invocation) snapshot() (string, error) {
	if len(in.positionals) > 0 {
		return in.positionals[0], nil
	}
	if snapshot := in.options["client-source"]; snapshot != "" {
		return snapshot, nil
	}
	if snapshot := in.options["snapshot"]; snapshot != "" {
		return snapshot, nil
	}
	selected, err := consumer.ReadVersion(in.root)
	if err != nil {
		return "", err
	}
	return release.ClientSnapshotPath(in.root, selected.Build), nil
}

func writeResult(result any, err error) error {
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

// parseOptions reads --name VALUE and --name=VALUE options restricted to
// allowed. With stopAtPositional, parsing ends at the first positional
// argument and the remainder is returned untouched; otherwise positionals
// are collected and "--" ends option parsing.
func parseOptions(args, allowed []string, stopAtPositional bool) (map[string]string, []string, error) {
	options := map[string]string{}
	var rest []string
	for n := 0; n < len(args); n++ {
		a := args[n]
		if !stopAtPositional && a == "--" {
			rest = append(rest, args[n+1:]...)
			break
		}
		if !strings.HasPrefix(a, "--") {
			if stopAtPositional {
				return options, args[n:], nil
			}
			rest = append(rest, a)
			continue
		}
		key, value, eq := strings.Cut(strings.TrimPrefix(a, "--"), "=")
		if !slices.Contains(allowed, key) {
			return nil, nil, fmt.Errorf("unknown option --%s", key)
		}
		if !eq {
			n++
			if n >= len(args) {
				return nil, nil, fmt.Errorf("missing --%s value", key)
			}
			value = args[n]
		}
		options[key] = value
	}
	return options, rest, nil
}

// resolvePaths fixes caller-relative paths before acquisition changes subprocess directories.
func resolvePaths(options map[string]string, positionals []string) error {
	for _, name := range []string{"client-source", "snapshot", "cache", "output"} {
		if path := options[name]; path != "" {
			absolute, err := filepath.Abs(path)
			if err != nil {
				return fmt.Errorf("--%s: %w", name, err)
			}
			options[name] = absolute
		}
	}
	for index, path := range positionals {
		absolute, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		positionals[index] = absolute
	}
	return nil
}
