package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"wow-teal-types/internal/consumer"
	"wow-teal-types/internal/release"
	"wow-teal-types/internal/wowapi"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if e := run(ctx, os.Args[1:]); e != nil {
		fmt.Fprintln(os.Stderr, "wow-types:", e)
		os.Exit(1)
	}
}
func run(ctx context.Context, args []string) error {
	root, cwd := "", ""
	for len(args) >= 2 && (args[0] == "--root" || args[0] == "--cwd") {
		if args[0] == "--root" {
			root = args[1]
		} else {
			cwd = args[1]
		}
		args = args[2:]
	}
	if root == "" {
		var e error
		root, e = os.Getwd()
		if e != nil {
			return e
		}
	}
	var e error
	root, e = filepath.Abs(root)
	if e != nil {
		return e
	}
	if cwd != "" {
		if e = os.Chdir(cwd); e != nil {
			return e
		}
	}
	if len(args) == 0 {
		return fmt.Errorf("expected generate, check-generator, bootstrap, acquire-client, stage-release, capture, probe, recover or version")
	}
	command := args[0]
	options, positionals, e := parseOptions(args[1:])
	if e != nil {
		return e
	}
	if e = validateOptions(command, options, positionals); e != nil {
		return e
	}
	if e = resolvePaths(command, options, positionals); e != nil {
		return e
	}
	selected, e := consumer.ReadVersion(root)
	if e != nil {
		return e
	}
	snapshot := options["client-source"]
	if snapshot == "" {
		snapshot = options["snapshot"]
	}
	if snapshot == "" {
		snapshot = filepath.Join(root, "build/wow-client", strconv.Itoa(selected.Build))
	}
	snapshot, e = filepath.Abs(snapshot)
	if e != nil {
		return e
	}
	var result any
	switch command {
	case "generate":
		result, e = release.Update(ctx, root)
	case "recover":
		result, e = release.Recover(root)
	case "version":
		manifest, err := consumer.Verify(root)
		if err != nil {
			return err
		}
		fmt.Println(manifest["version"])
		return nil
	case "bootstrap":
		_, e = release.Bootstrap(ctx, root, snapshot)
		result = map[string]any{"status": "verified", "snapshot": snapshot}
	case "acquire-client":
		result, e = release.Acquire(ctx, root, snapshot, options["cache"])
	case "stage-client":
		result, e = release.StageSelected(ctx, root, snapshot, options["cache"], options["region"])
	case "verify-client":
		if len(positionals) > 0 {
			snapshot = positionals[0]
		}
		result, e = release.VerifySnapshot(snapshot, nil)
	case "build-extractor":
		var binary string
		var receipt map[string]any
		binary, receipt, e = release.BuildExtractor(ctx, root)
		if e == nil {
			receipt["entry"] = binary
			result = receipt
		}
	case "check-generator":
		result, e = release.CheckGenerator(ctx, root, snapshot)
	case "stage-release":
		output := options["output"]
		if output == "" {
			output = filepath.Join(root, "build/wow-api-candidates", selected.Full)
		}
		refs := map[string]string{}
		for _, name := range []string{"blizzard", "questie", "tsm", "wowless", "backup"} {
			refs[name] = options[name+"-ref"]
		}
		result, e = release.StageRelease(ctx, root, output, snapshot, refs)
	case "capture":
		if len(positionals) != 1 {
			return fmt.Errorf("capture requires one SavedVariables/JSON file")
		}
		for _, n := range []string{"manifest", "receipt", "output"} {
			if options[n] == "" {
				return fmt.Errorf("capture requires --%s", n)
			}
		}
		result, e = wowapi.IngestCapture(positionals[0], options["manifest"], options["receipt"], options["output"])
	case "probe":
		if len(positionals) == 0 {
			return fmt.Errorf("probe requires plan or build")
		}
		result, e = release.Probe(ctx, root, positionals[0], options["output"])
	default:
		return fmt.Errorf("unknown command %q", command)
	}
	if e != nil {
		return e
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
func parseOptions(args []string) (map[string]string, []string, error) {
	options := map[string]string{}
	var rest []string
	allowed := map[string]bool{}
	for _, n := range strings.Fields("output client-source cache region snapshot questie-ref tsm-ref wowless-ref backup-ref blizzard-ref manifest receipt") {
		allowed[n] = true
	}
	for n := 0; n < len(args); n++ {
		a := args[n]
		if a == "--" {
			rest = append(rest, args[n+1:]...)
			break
		}
		if !strings.HasPrefix(a, "--") {
			rest = append(rest, a)
			continue
		}
		key, value, eq := strings.Cut(strings.TrimPrefix(a, "--"), "=")
		if !allowed[key] {
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

func validateOptions(command string, options map[string]string, positionals []string) error {
	allowed := map[string]string{
		"generate": "", "recover": "", "version": "",
		"bootstrap": "client-source snapshot", "acquire-client": "client-source snapshot cache",
		"stage-client": "client-source snapshot cache region", "verify-client": "client-source snapshot",
		"build-extractor": "",
		"check-generator": "client-source snapshot",
		"stage-release":   "client-source snapshot output blizzard-ref questie-ref tsm-ref wowless-ref backup-ref",
		"capture":         "manifest receipt output", "probe": "output",
	}
	names, known := allowed[command]
	if !known {
		return fmt.Errorf("unknown command %q", command)
	}
	for option := range options {
		found := false
		for _, name := range strings.Fields(names) {
			if name == option {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("--%s is not supported by %s", option, command)
		}
	}
	switch command {
	case "capture":
		if len(positionals) != 1 {
			return fmt.Errorf("capture requires one file")
		}
	case "probe":
		if len(positionals) != 1 {
			return fmt.Errorf("probe requires plan or build")
		}
		if positionals[0] == "plan" && len(options) > 0 {
			return fmt.Errorf("probe plan does not accept output options")
		}
	case "verify-client":
		if len(positionals) > 1 {
			return fmt.Errorf("verify-client accepts at most one snapshot")
		}
	default:
		if len(positionals) > 0 {
			return fmt.Errorf("%s does not accept positional arguments", command)
		}
	}
	return nil
}

// resolvePaths fixes caller-relative paths before acquisition changes subprocess directories.
func resolvePaths(command string, options map[string]string, positionals []string) error {
	for _, name := range []string{"client-source", "snapshot", "cache", "output", "manifest", "receipt"} {
		if path := options[name]; path != "" {
			absolute, err := filepath.Abs(path)
			if err != nil {
				return fmt.Errorf("--%s: %w", name, err)
			}
			options[name] = absolute
		}
	}
	if command == "capture" || command == "verify-client" {
		for index, path := range positionals {
			absolute, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			positionals[index] = absolute
		}
	}
	return nil
}
