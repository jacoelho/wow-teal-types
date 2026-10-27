package consumer

import (
	"bytes"
	"strings"
	"testing"
)

func TestManifestSchemaRejectsMalformedInput(t *testing.T) {
	const valid = `{"files":{"x":"digest"},"license":"MIT","name":"wow-teal-types","schema_version":1,"teal_version":"0.24.8","version":"0.1.0","wow_build":"12.1.5.70575"}`
	if _, err := ParsePackageManifest([]byte(valid)); err != nil {
		t.Fatal(err)
	}
	cases := []struct{ name, document string }{
		{"null", `null`}, {"array", `[]`}, {"empty", `{}`},
		{"missing files", strings.Replace(valid, `"files":{"x":"digest"},`, "", 1)},
		{"null files", strings.Replace(valid, `{"x":"digest"}`, `null`, 1)},
		{"array files", strings.Replace(valid, `{"x":"digest"}`, `[]`, 1)},
		{"null hash", strings.Replace(valid, `"digest"`, `null`, 1)},
		{"numeric hash", strings.Replace(valid, `"digest"`, `17`, 1)},
		{"empty hash", strings.Replace(valid, `"digest"`, `""`, 1)},
		{"null required string", strings.Replace(valid, `"MIT"`, `null`, 1)},
		{"numeric required string", strings.Replace(valid, `"MIT"`, `17`, 1)},
		{"missing required string", strings.Replace(valid, `"license":"MIT",`, "", 1)},
		{"null schema", strings.Replace(valid, `"schema_version":1`, `"schema_version":null`, 1)},
		{"unknown schema", strings.Replace(valid, `"schema_version":1`, `"schema_version":2`, 1)},
		{"unknown field", strings.TrimSuffix(valid, "}") + `,"unexpected":true}`},
		{"trailing document", valid + ` {}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ParsePackageManifest([]byte(tc.document)); err == nil {
				t.Fatalf("accepted malformed manifest: %s", tc.document)
			}
		})
	}
}

func TestBuildReceiptSchemaAcceptsMinimalOwnershipAndRejectsMalformedInput(t *testing.T) {
	receipt, err := parseBuildReceipt([]byte(`{"files":{"Main.lua":"digest"}}`))
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Files["Main.lua"] != "digest" || receipt.PackageVersion != "" || receipt.PackageManifestSHA256 != "" {
		t.Fatalf("unexpected minimal receipt: %+v", receipt)
	}
	for _, document := range []string{`null`, `[]`, `{}`, `{"files":null}`, `{"files":[]}`, `{"files":{"Main.lua":null}}`, `{"files":{"Main.lua":17}}`, `{"files":{},"package_version":null}`, `{"files":{},"package_manifest_sha256":null}`, `{"files":{},"unknown":true}`, `{"files":{}} {}`} {
		if _, err := parseBuildReceipt([]byte(document)); err == nil {
			t.Errorf("accepted malformed receipt: %s", document)
		}
	}
}

func TestConsumerSchemaWireBytes(t *testing.T) {
	manifest := PackageManifest{Files: map[string]string{"b": "two", "a": "one"}, License: "MIT", Name: "name<&", SchemaVersion: 1, TealVersion: "0.24.8", Version: "0.1.0", WoWBuild: "12.1.5.70575"}
	actual, err := manifestBytes(manifest)
	if err != nil {
		t.Fatal(err)
	}
	const wantManifest = `{
  "files": {
    "a": "one",
    "b": "two"
  },
  "license": "MIT",
  "name": "name<&",
  "schema_version": 1,
  "teal_version": "0.24.8",
  "version": "0.1.0",
  "wow_build": "12.1.5.70575"
}
`
	if string(actual) != wantManifest {
		t.Fatalf("manifest bytes changed:\n%s", actual)
	}
	actual, err = receiptBytes(BuildReceipt{Files: map[string]string{"<Main>.lua": "digest"}, PackageManifestSHA256: "manifest", PackageVersion: "0.1.0"})
	if err != nil {
		t.Fatal(err)
	}
	const wantReceipt = `{
  "files": {
    "<Main>.lua": "digest"
  },
  "package_manifest_sha256": "manifest",
  "package_version": "0.1.0"
}
`
	if string(actual) != wantReceipt {
		t.Fatalf("receipt bytes changed:\n%s", actual)
	}
	var out bytes.Buffer
	if err = writePackageResult(&out, PackageResult{Archive: "<archive>.zip", Members: 3, SHA256: "digest"}); err != nil {
		t.Fatal(err)
	}
	const wantResult = `{
  "archive": "<archive>.zip",
  "members": 3,
  "sha256": "digest"
}
`
	if out.String() != wantResult {
		t.Fatalf("result bytes changed:\n%s", out.String())
	}
}

func TestVersionRequiresBothIdentityParts(t *testing.T) {
	selected := Version{Full: "12.1.5.70575", Build: 70575}
	if err := selected.Require("12.1.5.70575", 70575); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		version string
		build   int
	}{{"12.1.6.70575", 70575}, {"12.1.5.70575", 70576}} {
		if err := selected.Require(tc.version, tc.build); err == nil {
			t.Errorf("accepted mismatched version %s/build %d", tc.version, tc.build)
		}
	}
}
