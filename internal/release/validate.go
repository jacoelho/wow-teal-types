package release

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"wow-teal-types/internal/consumer"
)

// Validate the newly emitted declarations rather than resolving an older checked
// catalogue from the active tree. Consumer secret checks remain a separate gate.
func validateDeclarations(ctx context.Context, root string, files map[string][]byte) error {
	stage, e := os.MkdirTemp("", "wow-catalogue-go-")
	if e != nil {
		return e
	}
	defer os.RemoveAll(stage)
	for name, data := range files {
		if strings.HasPrefix(name, "types/wow/") {
			if e = atomicWrite(filepath.Join(stage, filepath.Base(name)), data, 0644); e != nil {
				return e
			}
		}
	}
	native, e := readFile(filepath.Join(root, "types/wow/native_values.d.tl"), maxFile)
	if e != nil {
		return e
	}
	if e = atomicWrite(filepath.Join(stage, "native_values.d.tl"), native, 0644); e != nil {
		return e
	}
	if e = atomicWrite(filepath.Join(stage, "tlconfig.lua"), []byte("return {}\n"), 0644); e != nil {
		return e
	}
	compiler := os.Getenv("TL")
	if compiler == "" {
		compiler = filepath.Join(root, "tools/teal/tl")
	}
	compiler, e = exec.LookPath(compiler)
	if e != nil {
		return e
	}
	run := func(args ...string) error {
		child, cancel := context.WithTimeout(ctx, 60*time.Second)
		defer cancel()
		cmd := exec.CommandContext(child, compiler, args...)
		cmd.Dir = stage
		env := []string{}
		for _, v := range os.Environ() {
			if !strings.HasPrefix(v, "TL_PATH=") {
				env = append(env, v)
			}
		}
		cmd.Env = append(env, "TL_PATH="+filepath.ToSlash(stage)+"/?.lua;"+filepath.ToSlash(stage)+"/?/init.lua")
		b, e := cmd.CombinedOutput()
		if e != nil {
			return fmt.Errorf("generated declaration check: %w: %s", e, strings.TrimSpace(string(b)))
		}
		if len(args) == 1 && args[0] == "--version" && strings.TrimSpace(string(b)) != consumer.TealVersion {
			return fmt.Errorf("catalogue requires Teal %s", consumer.TealVersion)
		}
		return nil
	}
	if e = run("--version"); e != nil {
		return e
	}
	if e = run("run", filepath.Join(root, "tools/teal/check_capabilities.lua")); e != nil {
		return e
	}
	for _, name := range []string{"native_values.d.tl", "wow.d.tl", "wow_env.d.tl", "wow_secure_env.d.tl"} {
		if e = run("--feat-arity", "on", "--gen-target", "5.1", "--gen-compat", "off", "-I", stage, "check", filepath.Join(stage, name)); e != nil {
			return e
		}
	}
	return nil
}
