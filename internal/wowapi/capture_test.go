package wowapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
)

func captureTestRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("test source unavailable")
	}
	root, err := filepath.Abs(filepath.Join(filepath.Dir(file), "../.."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}
func captureTestJSON(t *testing.T, v any) []byte {
	t.Helper()
	data, err := jsonPretty(v)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
func captureTestObject(t *testing.T, v any) M {
	t.Helper()
	data := captureTestJSON(t, v)
	result, err := readCaptureData(data, false)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestCaptureClassificationMatchesFrozenTransportCases(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(captureTestRoot(t), "internal/wowapi/testdata/capture_cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Report, Manifest, Receipt, Classification M
		ManifestBytes                             string `json:"manifest_bytes"`
		Reject                                    string
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err = decoder.Decode(&cases); err != nil {
		t.Fatal(err)
	}
	for index, tc := range cases {
		t.Run(fmt.Sprintf("case_%03d", index), func(t *testing.T) {
			before := captureTestJSON(t, tc.Report)
			result, err := validateCapture(tc.Report, tc.Manifest, tc.Receipt, []byte(tc.ManifestBytes))
			if tc.Reject != "" {
				if err == nil {
					t.Fatalf("accepted capture rejected by baseline: %s", tc.Reject)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if !evEqual(result, tc.Classification) {
					t.Fatalf("classification changed: got %v want %v", result, tc.Classification)
				}
			}
			if !bytes.Equal(before, captureTestJSON(t, tc.Report)) {
				t.Fatal("capture validation rewrote source observations")
			}
		})
	}
}
func TestCaptureLiteralAndJSONBoundaries(t *testing.T) {
	for _, raw := range []string{`WoWAPIProbeDB = os.execute("bad")`, `WoWAPIProbeDB = {x=1,x=2}`, `WoWAPIProbeDB = {x=1+2}`, `WoWAPIProbeDB = {x=Enum.Secret}`, `WoWAPIProbeDB = {x="\z"}`, `WoWAPIProbeDB = {x="\x41"}`, `WoWAPIProbeDB = {[1]="value"}`, `WoWAPIProbeDB = {x=1} print("bad")`, `WoWAPIProbeDB = {x=1, "mixed"}`, `WoWAPIProbeDB = {x=` + strings.Repeat("{", 65) + strings.Repeat("}", 65) + `}`} {
		if _, err := readCaptureData([]byte(raw), true); err == nil {
			t.Errorf("accepted executable, ambiguous or unbounded capture %q", raw)
		}
	}
	for _, raw := range []string{`{"x":1,"x":2}`, `{"x":{"$ref":"Enum.Secret"}}`, `{"x":NaN}`, `{"x":1e999}`, `{"x":"` + strings.Repeat("x", 513) + `"}`, `{"x":true} {}`, `{"x":` + strings.Repeat("[", 65) + strings.Repeat("]", 65) + `}`} {
		if _, err := readCaptureData([]byte(raw), false); err == nil {
			t.Errorf("accepted invalid capture JSON %q", raw)
		}
	}
	if _, err := readCaptureData(make([]byte, captureMaxBytes+1), false); err == nil {
		t.Fatal("accepted oversized capture")
	}
	parsed, err := readCaptureData([]byte(`WoWAPIProbeDB = { first = false, second = -17, third = "A\000B", values = {true, false} }`), true)
	if err != nil {
		t.Fatal(err)
	}
	if parsed["first"] != false || asInt(parsed["second"]) != -17 || parsed["third"] != "A\x00B" || !evEqual(parsed["values"], []any{true, false}) {
		t.Fatalf("literal values changed: %v", parsed)
	}
}
func TestHistoricalCapturesRemainImmutableAndIngestionIsIdempotent(t *testing.T) {
	root := captureTestRoot(t)
	directory := filepath.Join(root, "references/wow-api/native-observations/12.1.5.70575/presence-1.0.0")
	capture := filepath.Join(directory, "capture.lua")
	manifest := filepath.Join(directory, "probe-build.json")
	receipt := filepath.Join(directory, "installed-receipt.json")
	output := filepath.Join(t.TempDir(), "evidence")
	first, err := IngestCapture(capture, manifest, receipt, output)
	if err != nil {
		t.Fatal(err)
	}
	second, err := IngestCapture(capture, manifest, receipt, output)
	if err != nil {
		t.Fatal(err)
	}
	if !evEqual(first, second) {
		t.Fatal("idempotent import changed classification")
	}
	path := filepath.Join(output, "report.json")
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	changed := append(append([]byte{}, before...), ' ')
	if err = os.WriteFile(path, changed, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err = IngestCapture(capture, manifest, receipt, output); err == nil {
		t.Fatal("overwrote changed historical evidence")
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(changed, after) {
		t.Fatal("failed immutable import changed files")
	}
}
func TestCaptureAcquisitionRejectsNonregularAndOversizedFiles(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "large.json")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = file.Truncate(captureMaxBytes * 32); err != nil {
		t.Fatal(err)
	}
	if err = file.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = readCaptureFile(path); err == nil {
		t.Fatal("accepted oversized capture")
	}
	fifo := filepath.Join(root, "capture.lua")
	if err = syscall.Mkfifo(fifo, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = readCaptureFile(fifo); err == nil {
		t.Fatal("accepted nonregular capture")
	}
}
