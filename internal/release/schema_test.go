package release

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"wow-teal-types/internal/sourcelock"
)

func TestExtractorConfigWireContract(t *testing.T) {
	want, err := os.ReadFile(filepath.Join("..", "..", "tools", "cdn", "testdata", "extractor-config.json"))
	if err != nil {
		t.Fatal(err)
	}
	id := sourcelock.ClientIdentity{Product: "wow", Region: "eu", Locale: "enUS", Build: 70575, Version: "12.1.5.70575", BuildConfig: strings.Repeat("a", 32), CDNConfig: strings.Repeat("b", 32), ProductConfig: strings.Repeat("c", 32)}
	got, err := jsonBytes(newExtractorConfig(id))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("release config.json =\n%s\nwant tools/cdn fixture\n%s", got, want)
	}
}

func TestExtractorReceiptWireContract(t *testing.T) {
	receipt, err := readJSON[extractorReceipt](filepath.Join("..", "..", "tools", "cdn", "testdata", "extractor-result.json"))
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Product != "wow" || receipt.Version != "12.1.5.70575" || receipt.Build != 70575 || len(receipt.Files) != 1 {
		t.Fatalf("receipt identity or inventory lost: %+v", receipt)
	}
	file := receipt.Files[0]
	if file.Path != "Interface/example.lua" || file.Size != 5 || file.SHA256 != "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824" || file.ContentKey != "5d41402abc4b2a76b9719d911017c592" {
		t.Fatalf("receipt file integrity lost: %+v", file)
	}
}
