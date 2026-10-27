package release

import (
	"path/filepath"
	"testing"
)

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
