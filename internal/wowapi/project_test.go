package wowapi

import (
	"strings"
	"testing"
)

func TestInputPathRejectsUnselectedNativeRoot(t *testing.T) {
	project := NewProject(t.TempDir())
	for _, test := range []struct {
		name    string
		project Project
	}{
		{"missing", project},
		{"empty", project.WithInputs(map[string]string{"blizzard": ""})},
	} {
		t.Run(test.name, func(t *testing.T) {
			path, err := test.project.InputPath("blizzard", "A.lua")
			if path != "" || err == nil || !strings.Contains(err.Error(), "blizzard") {
				t.Fatalf("InputPath = %q, %v; want no path and missing group error", path, err)
			}
		})
	}
}
