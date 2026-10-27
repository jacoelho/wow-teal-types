package main

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func formatHex(t *testing.T, text string) []byte {
	t.Helper()
	data, err := hex.DecodeString(strings.Join(strings.Fields(text), ""))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestBLTEPlainCompressedAndChunkChecks(t *testing.T) {
	fixtures := []struct {
		name, encoded, text string
		size                uint64
	}{
		{"plain", "424c5445000000004e616263", "abc", 3},
		{"zlib", "424c5445000000005a789ccb48cdc9c90700062c0215", "hello", 5},
		{"chunked", "424c5445000000240f0000010000000400000003cc62f2c3819619698137cc31f9d02b424e616263", "abc", 3},
	}
	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			data := formatHex(t, fixture.encoded)
			if err := verifyBLTEEnvelope(data); err != nil {
				t.Fatal(err)
			}
			actual, err := decodeBLTE(data, fixture.size, fixture.size)
			if err != nil || string(actual) != fixture.text {
				t.Fatalf("decode=%q, %v", actual, err)
			}
			if _, err := decodeBLTE(data, fixture.size-1, 0); err == nil {
				t.Fatal("accepted decoded-size overflow")
			}
			if _, err := decodeBLTE(data, 100, fixture.size+1); err == nil {
				t.Fatal("accepted wrong expected size")
			}
			for n := range len(data) {
				if _, err := decodeBLTE(data[:n], 100, fixture.size); err == nil {
					t.Fatalf("accepted truncated prefix %d", n)
				}
			}
		})
	}
	corrupted := formatHex(t, "424c5445000000240f0000010000000400000003cc62f2c3819619698137cc31f9d02b424e616263")
	corrupted[len(corrupted)-1] ^= 1
	if err := verifyBLTEEnvelope(corrupted); err == nil || !strings.Contains(err.Error(), "checksum") {
		t.Fatalf("intact header with corrupt payload accepted before decoding: %v", err)
	}
	for _, data := range [][]byte{
		formatHex(t, "424c54450000000045616263"), formatHex(t, "424c54450000000046616263"),
		formatHex(t, "424c5445000000005a789ccb48cdc9c90700062c021500"),
		formatHex(t, "424c5445000000240f0000010000000400000003000000000000000000000000000000004e616263"),
		formatHex(t, "424c5445000000240f0000010000000400000003cc62f2c3819619698137cc31f9d02b424e61626300"),
	} {
		if _, err := decodeBLTE(data, 100, 0); err == nil {
			t.Fatalf("accepted invalid BLTE %x", data)
		}
	}
}

func encodingFixture(t *testing.T) []byte {
	t.Helper()
	// One content record (size3, one EKey), one encoding record, 1KiB pages.
	content := make([]byte, 1024)
	copy(content, formatHex(t, "0100000000031112131415161718191a1b1c1d1e1f202122232425262728292a2b2c2d2e2f30"))
	encoded := make([]byte, 1024)
	copy(encoded, formatHex(t, "2122232425262728292a2b2c2d2e2f3000000000000000000c"))
	// A null EKey and specification index -1 terminate the final page.
	binary.BigEndian.PutUint32(encoded[41:45], 0xffffffff)
	data := formatHex(t, "454e0110100001000100000001000000010000000002")
	data = append(data, 'n', 0)
	csum := md5.Sum(content)
	esum := md5.Sum(encoded)
	data = append(data, content[6:22]...)
	data = append(data, csum[:]...)
	data = append(data, content...)
	data = append(data, encoded[:16]...)
	data = append(data, esum[:]...)
	return append(data, encoded...)
}

func TestEncodingLookupAndIntegrity(t *testing.T) {
	data := encodingFixture(t)
	table, err := parseEncoding(data)
	if err != nil {
		t.Fatal(err)
	}
	content := key(formatHex(t, "1112131415161718191a1b1c1d1e1f20"))
	encoded := key(formatHex(t, "2122232425262728292a2b2c2d2e2f30"))
	actual, err := table.lookup(content)
	if err != nil || actual.Size != 3 || len(actual.Keys) != 1 || actual.Keys[0] != encoded {
		t.Fatalf("lookup=%+v, %v", actual, err)
	}
	if _, err := table.lookup(key{}); err == nil {
		t.Fatal("accepted missing content")
	}
	for _, n := range []int{0, 1, 21, 24, 55, len(data) - 1} {
		if _, err := parseEncoding(data[:n]); err == nil {
			t.Fatalf("accepted truncated length %d", n)
		}
	}
	for _, n := range []int{2, 3, 9, 17, 21, 24 + 16, 24 + 32 + 4, 24 + 32 + 1024 + 16, 24 + 32 + 1024 + 32 + 16} {
		changed := bytes.Clone(data)
		changed[n] ^= 1
		if _, err := parseEncoding(changed); err == nil {
			t.Fatalf("accepted changed byte %d", n)
		}
	}
}

func rootFixture(t *testing.T) []byte {
	t.Helper()
	return formatHex(t, `5453464d 18000000 02000000 01000000 01000000 00000000
 01000000 02000000 00000000 00000000 00
 07000000 1112131415161718191a1b1c1d1e1f20 0807060504030201`)
}

func TestRootSelectionAndMalformedVariants(t *testing.T) {
	data := rootFixture(t)
	wanted := map[uint32]bool{7: true}
	names := map[uint64]bool{0x0102030405060708: true}
	ids, hashes, err := scanRoot(data, wanted, names)
	expected := key(formatHex(t, "1112131415161718191a1b1c1d1e1f20"))
	if err != nil || ids[7] != expected || hashes[0x0102030405060708] != expected {
		t.Fatalf("root=%v %v, %v", ids, hashes, err)
	}
	for n := range len(data) {
		if _, _, err := scanRoot(data[:n], wanted, names); err == nil {
			t.Fatalf("accepted truncated root %d", n)
		}
	}
	for _, change := range []struct {
		name   string
		offset int
		value  byte
	}{
		{"unsupported version", 8, 3}, {"unknown header", 20, 1}, {"count mismatch", 12, 2},
		{"wrong locale", 28, 4}, {"low violence", 32, 0x80}, {"do not load", 33, 1},
	} {
		t.Run(change.name, func(t *testing.T) {
			bad := bytes.Clone(data)
			bad[change.offset] = change.value
			if _, _, err := scanRoot(bad, wanted, names); err == nil {
				t.Fatal("accepted invalid/ineligible root")
			}
		})
	}
	ambiguous := append(bytes.Clone(data), data[24:]...)
	binary.LittleEndian.PutUint32(ambiguous[12:16], 2)
	binary.LittleEndian.PutUint32(ambiguous[16:20], 2)
	ambiguous[len(data)+21] ^= 1
	if _, _, err := scanRoot(ambiguous, wanted, names); err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("ambiguous root: %v", err)
	}
	if _, _, err := scanRoot(data, map[uint32]bool{8: true}, nil); err == nil {
		t.Fatal("accepted missing requested file ID")
	}
}

func TestPathHashRecordedClientVectors(t *testing.T) {
	// Expected hashes were read from the independently decoded 70575 root entries.
	vectors := []struct {
		path string
		hash uint64
	}{
		{"Interface/ui-code-list.txt", 0xf1ce4163ed49e783},
		{"Interface/ui-toc-list.txt", 0x955e5dbb2d4c3d22},
		{"Interface/ui-gen-addon-list.txt", 0x0562952677b34b56},
		{"Interface/AddOns/Blizzard_APIDocumentationGenerated/UnitDocumentation.lua", 0x4bbb224dcdc6e064},
	}
	for _, v := range vectors {
		if got := pathHash(v.path); got != v.hash {
			t.Fatalf("%s: %016x, expected %016x", v.path, got, v.hash)
		}
		if pathHash(strings.ToUpper(strings.ReplaceAll(v.path, "/", "\\"))) != v.hash {
			t.Fatal("path normalization changed lookup")
		}
	}
	if pathHash("") != 0xdeadbeefdeadbeef {
		t.Fatal("empty lookup3 vector")
	}
}

func TestRootV1UnnamedIDSelection(t *testing.T) {
	data := formatHex(t, `5453464d 14000000 01000000 01000000 00000000
 01000000 00000010 02000000
 09000000 1112131415161718191a1b1c1d1e1f20`)
	ids, names, err := scanRoot(data, map[uint32]bool{9: true}, nil)
	if err != nil || ids[9] != key(formatHex(t, "1112131415161718191a1b1c1d1e1f20")) || len(names) != 0 {
		t.Fatalf("unnamed ID selection = %v, %v, %v", ids, names, err)
	}
	if _, _, err := scanRoot(data, nil, map[uint64]bool{1: true}); err == nil {
		t.Fatal("unnamed ID invented a path lookup")
	}
}

func TestRecordedClientFormats(t *testing.T) {
	root := os.Getenv("WOW_CDN_FORMAT_REFERENCE")
	if root == "" {
		t.Skip("set WOW_CDN_FORMAT_REFERENCE to a 70575 reference cache with snapshot.json")
	}
	data, err := os.ReadFile(filepath.Join(root, "snapshot.json"))
	if err != nil {
		t.Fatal(err)
	}
	var snapshot struct {
		Files []struct {
			Path    string `json:"path"`
			Content string `json:"content_key"`
			Size    uint64 `json:"size"`
		} `json:"files"`
	}
	if err := json.Unmarshal(data, &snapshot); err != nil {
		t.Fatal(err)
	}
	metadata := filepath.Join(root, "metadata")
	enc, err := os.ReadFile(filepath.Join(metadata, "09bbef079e2e00af0b4796c3594fe938.decoded"))
	if err != nil {
		t.Fatal(err)
	}
	table, err := parseEncoding(enc)
	if err != nil {
		t.Fatal(err)
	}
	rawRoot, err := os.ReadFile(filepath.Join(metadata, "0eb99647827081ef97088c189f1a6375.decoded"))
	if err != nil {
		t.Fatal(err)
	}
	ids, _, err := scanRoot(rawRoot, map[uint32]bool{6067012: true, 6067013: true, 6076661: true}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if ids[6067012] != key(formatHex(t, "1ff7756eb3265ff38ca80c4073f97725")) || ids[6067013] != key(formatHex(t, "bb325df4e00903a0ef542bf80a8f7349")) || ids[6076661] != key(formatHex(t, "81051bcc2cf1bedf378224b0a93e2877")) {
		t.Fatal("manifest content keys differ from verified snapshot")
	}
	wanted := make(map[uint64]bool, len(snapshot.Files))
	for _, row := range snapshot.Files {
		wanted[pathHash(row.Path)] = true
	}
	_, names, err := scanRoot(rawRoot, nil, wanted)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range snapshot.Files {
		expected := key(formatHex(t, row.Content))
		if names[pathHash(row.Path)] != expected {
			t.Fatalf("root content mismatch: %s", row.Path)
		}
		entry, err := table.lookup(expected)
		if err != nil || entry.Size != row.Size {
			t.Fatalf("encoding %s size %d expected %d: %v", row.Path, entry.Size, row.Size, err)
		}
	}
	t.Logf("all %d root content keys and encoding sizes match independently acquired snapshot", len(snapshot.Files))
}

func FuzzFormatBounds(f *testing.F) {
	for _, text := range []string{
		"424c5445000000004e616263",
		"424c5445000000240f0000010000000400000003cc62f2c3819619698137cc31f9d02b424e616263",
		"5453464d18000000020000000100000001000000000000000100000002000000000000000000000000070000001112131415161718191a1b1c1d1e1f200807060504030201",
		"454e0110100001000100000001000000010000000002",
	} {
		data, err := hex.DecodeString(text)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(data)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 8192 {
			t.Skip()
		}
		result, _ := decodeBLTE(data, 8192, 0)
		if len(result) > 8192 {
			t.Fatal("decoded output exceeded its limit")
		}
		_, _ = parseEncoding(data)
		_, _, _ = scanRoot(data, nil, nil)
		_ = verifyBLTEEnvelope(data)
	})
}
