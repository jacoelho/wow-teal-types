package consumer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
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

func (tc secretCase) valid() bool {
	reject := tc.Verdict == "reject"
	return (reject || tc.Verdict == "accept") && reject == (tc.Reject != "") && (reject || !tc.Compiler)
}

// These authored cases exercise language and secrecy contracts independently of
// Go command internals. Cases without modules share one checker run, as the
// sources of an addon do, so each case also shows that diagnostics stay on the
// source that caused them. Every checker rejection is paired with a stock `tl
// check` control that must pass, so each row proves the checker adds the
// rejection rather than repeating a compiler error.
func TestSecretOperationContracts(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	tl := teal(t)
	checker := filepath.Join(root, "tools/check_secrets.lua")
	declarations := filepath.Join(root, "types/wow")
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
	shared := secretGroup{dir: t.TempDir(), search: []string{declarations}}
	for index, tc := range cases {
		if !tc.valid() {
			t.Fatalf("malformed case %03d: verdict %q, reject %q, compiler %v", index, tc.Verdict, tc.Reject, tc.Compiler)
		}
		if len(tc.Modules) == 0 {
			shared.add(t, caseSource(index), tc)
		}
	}
	diagnostics := shared.check(t, tl, checker)
	for index, tc := range cases {
		name := fmt.Sprintf("%s/%03d", strings.TrimPrefix(tc.Name, "test_"), index)
		if len(tc.Modules) == 0 {
			t.Run(name, func(t *testing.T) { assertSecretCase(t, tc, diagnostics[caseSource(index)]) })
			continue
		}
		// Modules resolve from the case directory, so such a case runs alone.
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			group := secretGroup{dir: t.TempDir()}
			group.search = []string{declarations, group.dir}
			for module, body := range tc.Modules {
				path := filepath.Join(group.dir, module)
				writeTest(t, path, body)
				group.modules = append(group.modules, path)
			}
			group.add(t, "caller.tl", tc)
			assertSecretCase(t, tc, group.check(t, tl, checker)["caller.tl"])
		})
	}
}

// caseSource names a case's source after its index, as its subtest is named.
func caseSource(index int) string { return fmt.Sprintf("case_%03d.tl", index) }

// secretGroup is the sources of one checker run, written to dir.
type secretGroup struct {
	dir     string
	search  []string
	sources []string
	// controls are the sources whose rejection the checker owns.
	controls []string
	// modules are the imported files, which Teal reports by their path.
	modules []string
}

func (g *secretGroup) add(t *testing.T, name string, tc secretCase) {
	t.Helper()
	writeTest(t, filepath.Join(g.dir, name), tc.Source)
	g.sources = append(g.sources, name)
	if tc.Verdict == "reject" && !tc.Compiler {
		g.controls = append(g.controls, name)
	}
}

// check runs the stock control and the checker over the group and returns each
// source's diagnostics. Output that names no source of the group, or an exit
// status that disagrees with the diagnostics, fails the test: either would let
// a rejection be lost or credited to the wrong case.
func (g secretGroup) check(t *testing.T, tl, checker string) map[string][]diagnostic {
	t.Helper()
	run := func(args []string) (string, error) {
		command := exec.CommandContext(t.Context(), tl, args...)
		command.Dir = g.dir
		output, err := command.CombinedOutput()
		return string(output), err
	}
	if len(g.controls) > 0 {
		// One stock run shares globals between files; a leak can only make the
		// control pass a case that the isolated checker run then rejects early.
		output, err := run(append(compilerFlags(g.search), append([]string{"--global-env-def", "wow_env", "check"}, g.controls...)...))
		if err != nil {
			errors := stockErrors(output)
			if len(errors) == 0 {
				errors = []string{output}
			}
			t.Errorf("stock control rejected checker cases: %v\n%s", err, strings.Join(errors, "\n"))
		}
	}
	output, err := run(secretCheckArgs(g.search, checker, "wow_env", "wow_env", g.sources))
	diagnostics, parseErr := parseDiagnostics(output, slices.Concat(g.sources, g.modules))
	if parseErr != nil {
		t.Fatalf("%v\n%s", parseErr, output)
	}
	failed := false
	for _, list := range diagnostics {
		failed = failed || slices.ContainsFunc(list, diagnostic.failure)
	}
	if (err != nil) != failed {
		t.Fatalf("checker exit %v disagrees with its diagnostics:\n%s", err, output)
	}
	return diagnostics
}

func assertSecretCase(t *testing.T, tc secretCase, diagnostics []diagnostic) {
	t.Helper()
	report := func() string {
		lines := make([]string, len(diagnostics))
		for index, d := range diagnostics {
			lines[index] = d.line
		}
		return strings.Join(lines, "\n")
	}
	if tc.Verdict == "accept" {
		if slices.ContainsFunc(diagnostics, diagnostic.failure) {
			t.Fatalf("unexpected rejection:\n%s", report())
		}
		return
	}
	// A checker category starts its message; stock Teal may phrase a compiler
	// rejection inside a longer message.
	if !slices.ContainsFunc(diagnostics, func(d diagnostic) bool {
		return d.failure() && (strings.HasPrefix(d.message, tc.Reject) || tc.Compiler && strings.Contains(d.message, tc.Reject))
	}) {
		t.Fatalf("missing diagnostic %q:\n%s", tc.Reject, report())
	}
	for _, want := range tc.Extra {
		if !slices.ContainsFunc(diagnostics, func(d diagnostic) bool { return strings.Contains(d.line, want) }) {
			t.Fatalf("missing diagnostic detail %q:\n%s", want, report())
		}
	}
}

// diagnostic is one `SOURCE:LINE:COL: message` line of checker output.
type diagnostic struct{ line, message string }

func (d diagnostic) failure() bool { return !strings.HasPrefix(d.message, "warning: ") }

var diagnosticLine = regexp.MustCompile(`^([^:]+):\d+:\d+: (.*)$`)

// parseDiagnostics groups checker output by file and rejects any line that does
// not name one of files.
func parseDiagnostics(output string, files []string) (map[string][]diagnostic, error) {
	diagnostics := map[string][]diagnostic{}
	for line := range strings.Lines(output) {
		line = strings.TrimSuffix(line, "\n")
		match := diagnosticLine.FindStringSubmatch(line)
		if match == nil || !slices.Contains(files, match[1]) {
			return nil, fmt.Errorf("checker output names no checked file: %q", line)
		}
		diagnostics[match[1]] = append(diagnostics[match[1]], diagnostic{line, match[2]})
	}
	return diagnostics, nil
}

var stockErrorHeader = regexp.MustCompile(`^\d+ errors?:$`)

// stockErrors returns the error lines of `tl check` output, which lists each
// file's warnings and errors in separate sections.
func stockErrors(output string) []string {
	var errors []string
	inErrors := false
	for line := range strings.Lines(output) {
		line = strings.TrimSuffix(line, "\n")
		switch {
		case stockErrorHeader.MatchString(line):
			inErrors = true
		case strings.HasPrefix(line, "----"), strings.HasPrefix(line, "===="):
			inErrors = false
		case inErrors:
			errors = append(errors, line)
		}
	}
	return errors
}

// Sources checked in one run are independent: a global that one source declares
// is unknown to the next, as it is when that source is checked alone.
func TestSecretCheckSourcesAreIndependent(t *testing.T) {
	t.Parallel()
	output, diagnostics, err := runNativeCheck(t, nativeFixture(t, nil), "",
		checkSource{"declare.tl", "global probe_counter: number = 0\nprint(probe_counter)\n"},
		checkSource{"read.tl", "print(probe_counter + 1)\n"})
	failures := slices.DeleteFunc(slices.Clone(diagnostics["read.tl"]), func(d diagnostic) bool { return !d.failure() })
	if err == nil || slices.ContainsFunc(diagnostics["declare.tl"], diagnostic.failure) ||
		len(failures) != 1 || failures[0].line != "read.tl:1:7: unknown variable: probe_counter" {
		t.Fatalf("sources were not checked independently: %v\n%s", err, output)
	}
}
