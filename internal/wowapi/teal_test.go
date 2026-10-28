package wowapi

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"wow-teal-types/internal/consumer"
)

// StockTeal returns the path of the stock Teal compiler, from TL or PATH.
// Exported so external test packages in this directory share one resolver.
func StockTeal(t testing.TB) string {
	t.Helper()
	install := "install it with `luarocks install tl " + consumer.TealVersion + "` or set TL"
	path := os.Getenv("TL")
	if path == "" {
		found, err := exec.LookPath("tl")
		if err != nil {
			t.Fatalf("stock Teal %s not found on PATH: %v; %s", consumer.TealVersion, err, install)
		}
		path = found
	}
	output, err := exec.CommandContext(t.Context(), path, "--version").CombinedOutput()
	if err != nil {
		t.Fatalf("%s --version: %v\n%s\n%s", path, err, output, install)
	}
	if version := strings.TrimSpace(string(output)); version != consumer.TealVersion {
		t.Fatalf("%s reports Teal %q, want %q; %s", path, version, consumer.TealVersion, install)
	}
	return path
}
