package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
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
	if err = consumer.Run(ctx, root, args, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "wow-teal:", err)
		os.Exit(1)
	}
}
