package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"wow-teal-types/internal/consumer"
)

func main() {
	args := os.Args[1:]
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
		var err error
		root, err = os.Getwd()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	root, err := filepath.Abs(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if cwd != "" {
		if err = os.Chdir(cwd); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err = run(ctx, root, args, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "wow-teal:", err)
		os.Exit(1)
	}
}

// run translates one command line into a typed consumer call.
func run(ctx context.Context, root string, args []string, out, errout io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("expected version, verify, check-types, check, build, package or rockspecs")
	}
	command, rest := args[0], args[1:]
	switch command {
	case "version", "verify", "check-types", "package":
		if len(rest) != 0 {
			return fmt.Errorf("%s takes no arguments", command)
		}
	}
	switch command {
	case "version":
		manifest, err := consumer.Verify(root)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(out, manifest.Version)
		return err
	case "verify":
		manifest, err := consumer.Verify(root)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(out, "Verified wow-teal %s / Retail %s\n", manifest.Version, manifest.WoWBuild)
		return err
	case "check-types":
		return consumer.CheckTypes(ctx, root, out, errout)
	case "package":
		result, err := consumer.Package(root)
		if err != nil {
			return err
		}
		return consumer.WritePackageResult(out, result)
	case "rockspecs":
		options, err := parseRockspecs(rest)
		if err != nil {
			return err
		}
		return consumer.Rockspecs(root, options, out, errout)
	case "check", "build":
		selected, output, err := parseSelection(rest)
		if err != nil {
			return err
		}
		if command == "check" {
			if output != "" {
				return fmt.Errorf("check does not accept --output")
			}
			return consumer.Check(ctx, root, selected, out, errout)
		}
		return consumer.Build(ctx, root, selected, output, out, errout)
	}
	return fmt.Errorf("unknown command: %s", command)
}

func parseSelection(args []string) (consumer.Selection, string, error) {
	var selected consumer.Selection
	output := ""
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch a {
		case "--secure":
			selected.Secure = true
		case "-I", "--include", "-o", "--output", "--global-env-def":
			if i+1 == len(args) {
				return selected, "", fmt.Errorf("missing argument for %s", a)
			}
			i++
			switch a {
			case "-I", "--include":
				selected.Include = append(selected.Include, args[i])
			case "-o", "--output":
				output = args[i]
			default:
				selected.Global = args[i]
			}
		default:
			if strings.HasPrefix(a, "-") {
				return selected, "", fmt.Errorf("unknown option: %s", a)
			}
			selected.Sources = append(selected.Sources, a)
		}
	}
	return selected, output, nil
}

func parseRockspecs(args []string) (consumer.RockspecOptions, error) {
	var options consumer.RockspecOptions
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--check":
			options.Check = true
		case "--repository", "--description-json", "--output":
			if i+1 == len(args) {
				return options, fmt.Errorf("missing argument for %s", args[i])
			}
			option := args[i]
			i++
			switch option {
			case "--repository":
				options.Repository = args[i]
			case "--description-json":
				options.DescriptionJSON = args[i]
			case "--output":
				options.Output = args[i]
			}
		default:
			return options, fmt.Errorf("unknown rockspec option: %s", args[i])
		}
	}
	return options, nil
}
