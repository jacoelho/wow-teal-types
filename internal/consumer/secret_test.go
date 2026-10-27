package consumer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate test source")
	}
	root, err := filepath.Abs(filepath.Join(filepath.Dir(file), "../.."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}
func teal(t *testing.T) string {
	t.Helper()
	name := os.Getenv("TL")
	if name == "" {
		name = filepath.Join(repositoryRoot(t), "tools/teal/tl")
	}
	path, err := exec.LookPath(name)
	if err != nil {
		t.Fatal("bundled Teal compiler required:", err)
	}
	return path
}
func writeTest(t *testing.T, path, source string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
}

// These authored cases exercise language and secrecy contracts independently of
// Go command internals. Unknown-type cases reject at the compiler boundary;
// precise marker cases separately exercise the secret checker.
func TestSecretOperationContracts(t *testing.T) {
	root := repositoryRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "internal/consumer/testdata/secret_cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name, Source, Reject string
		Plain                bool
		Modules              map[string]string
		Extra                []string
	}
	if err = json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for index, tc := range cases {
		t.Run(fmt.Sprintf("%s/%03d", strings.TrimPrefix(tc.Name, "test_"), index), func(t *testing.T) {
			work := t.TempDir()
			source := filepath.Join(work, "caller.tl")
			writeTest(t, source, tc.Source)
			args := []string{"-I", filepath.Join(root, "types/wow"), "--gen-target", "5.1", "--gen-compat", "off"}
			if len(tc.Modules) > 0 {
				for name, body := range tc.Modules {
					writeTest(t, filepath.Join(work, name), body)
				}
				args = append(args, "-I", work)
			}
			if tc.Plain {
				args = append(args, "--global-env-def", "wow_env", "check", source)
			} else {
				args = append(args, "run", filepath.Join(root, "tools/check_secrets.lua"), "--", "wow_env", source)
			}
			command := exec.CommandContext(t.Context(), teal(t), args...)
			command.Dir = root
			var stdout, stderr bytes.Buffer
			command.Stdout = &stdout
			command.Stderr = &stderr
			err := command.Run()
			output := stdout.String() + stderr.String()
			if tc.Reject == "" {
				if err != nil {
					t.Fatalf("unexpected rejection: %v\n%s", err, output)
				}
			} else {
				if err == nil {
					t.Fatalf("accepted source requiring %q", tc.Reject)
				}
				for _, want := range append([]string{tc.Reject, "caller.tl:"}, tc.Extra...) {
					if !strings.Contains(output, want) {
						t.Fatalf("missing diagnostic %q:\n%s", want, output)
					}
				}
			}
		})
	}
}

func runConsumer(t *testing.T, ctx context.Context, root string, args []string, reject string) string {
	t.Helper()
	var output bytes.Buffer
	err := Run(ctx, root, args, &output, &output)
	text := output.String()
	if reject == "" {
		if err != nil {
			t.Fatalf("consumer command %v: %v\n%s", args, err, text)
		}
	} else {
		if err == nil {
			t.Fatalf("consumer accepted %v; expected %q", args, reject)
		}
		text += "\n" + err.Error()
		if !strings.Contains(text, reject) {
			t.Fatalf("expected %q in %s", reject, text)
		}
	}
	return text
}
