package wowapi

import (
	"path/filepath"
	"runtime"
	"testing"
)

func catalogueTestRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("missing test location")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../.."))
}
