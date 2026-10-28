package consumer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type secretCase struct {
	Name    string            `json:"name"`
	Source  string            `json:"source"`
	Modules map[string]string `json:"modules"`
	// Verdict is "accept" or "reject".
	Verdict string `json:"verdict"`
	// Compiler marks rejections owned by stock Teal rather than the checker.
	Compiler bool     `json:"compiler"`
	Reject   string   `json:"reject"`
	Extra    []string `json:"extra"`
}

// These authored cases exercise language and secrecy contracts independently of
// Go command internals. Every checker rejection is paired with a stock `tl
// check` control that must pass, so each row proves the checker adds the
// rejection rather than repeating a compiler error.
func TestSecretOperationContracts(t *testing.T) {
	root := repositoryRoot(t)
	tl := teal(t)
	checker := filepath.Join(root, "tools/check_secrets.lua")
	data, err := os.ReadFile(filepath.Join(root, "internal/consumer/testdata/secret_cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cases []secretCase
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&cases); err != nil {
		t.Fatal(err)
	}
	for index, tc := range cases {
		t.Run(fmt.Sprintf("%s/%03d", strings.TrimPrefix(tc.Name, "test_"), index), func(t *testing.T) {
			t.Parallel()
			reject := tc.Verdict == "reject"
			if tc.Verdict != "accept" && !reject || reject == (tc.Reject == "") || tc.Compiler && !reject {
				t.Fatalf("malformed case: verdict %q, reject %q, compiler %v", tc.Verdict, tc.Reject, tc.Compiler)
			}
			work := t.TempDir()
			writeTest(t, filepath.Join(work, "caller.tl"), tc.Source)
			search := []string{filepath.Join(root, "types/wow")}
			if len(tc.Modules) > 0 {
				for name, body := range tc.Modules {
					writeTest(t, filepath.Join(work, name), body)
				}
				search = append(search, work)
			}
			run := func(args []string) (string, error) {
				command := exec.CommandContext(t.Context(), tl, args...)
				command.Dir = work
				output, err := command.CombinedOutput()
				return string(output), err
			}
			if reject && !tc.Compiler {
				if output, err := run(append(compilerFlags(search), "--global-env-def", "wow_env", "check", "caller.tl")); err != nil {
					t.Fatalf("stock control rejected a checker case: %v\n%s", err, output)
				}
			}
			output, err := run(secretCheckArgs(search, checker, "wow_env", "caller.tl", "wow_env"))
			if !reject {
				if err != nil {
					t.Fatalf("unexpected rejection: %v\n%s", err, output)
				}
				return
			}
			if err == nil {
				t.Fatalf("accepted source requiring %q\n%s", tc.Reject, output)
			}
			if !strings.Contains(output, "caller.tl:") || !strings.Contains(output, tc.Reject) || !tc.Compiler && !hasDiagnostic(output, tc.Reject) {
				t.Fatalf("missing diagnostic %q at caller.tl:\n%s", tc.Reject, output)
			}
			for _, want := range tc.Extra {
				if !strings.Contains(output, want) {
					t.Fatalf("missing diagnostic detail %q:\n%s", want, output)
				}
			}
		})
	}
}

// hasDiagnostic reports a `caller.tl:LINE:COL: reject` line, so the category
// must start the diagnostic rather than appear inside another message.
func hasDiagnostic(output, reject string) bool {
	for line := range strings.SplitSeq(output, "\n") {
		position, message, ok := strings.Cut(line, ": ")
		if ok && strings.HasPrefix(position, "caller.tl:") && strings.HasPrefix(message, reject) {
			return true
		}
	}
	return false
}
