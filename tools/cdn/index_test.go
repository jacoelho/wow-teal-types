package main

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"testing"
)

// One-record vectors were calculated independently from the published CASC
// layout. Fixed checksums ensure the parser cannot validate its own encoder.
var indexVectors = [][4]string{
	{"7c7b7ea4022025fb", "9d84dbd7fefd18fb", "00b680444b47ec4d", "603562acda6d94ac0b153ec26cc35ce6"},
	{"2bb038b6e6419f48", "2de0b4e604195d27", "00b680444b47ec4d", "3aaadaaea4b910281b8f32bb08365575"},
	{"b379d69e49bcfed8", "64714ce00184b4a4", "00b680444b47ec4d", "9ff4855eca76e59135b44ec31472b485"},
	{"a29027cc7527de2b", "de6540b754d28740", "00b680444b47ec4d", "e696d6f3c1908d09dd4e2c7c50aaaff1"},
	{"cae184a4cd78da81", "6fe7250d0c2a8cd5", "00b680444b47ec4d", "2e7e1c2852b85bb3440c8b6e78ddd214"},
	{"07ecb45aeca9bd11", "e3d16652c52f124a", "00b680444b47ec4d", "aa488f47af53bc4e5d2c02d9fae8f450"},
	{"c7d8e5b27f970676", "4f08a6acf287ea84", "00b680444b47ec4d", "5af7561b5fa8e7bc511f72c142f5bd41"},
	{"4040e1e2ab48bb43", "1351885463019e9b", "00b680444b47ec4d", "4863a60fe9bc67ccc4cd2281435fda84"},
}

func literalIndex(t *testing.T, tag int) ([]byte, key) {
	t.Helper()
	data := make([]byte, 4148)
	data[0], data[19], data[23], data[4096] = byte(tag), 32, 64, byte(tag)
	for i, position := range []int{4112, 4120, 4140} {
		value, err := hex.DecodeString(indexVectors[tag-1][i])
		if err != nil {
			t.Fatal(err)
		}
		copy(data[position:], value)
	}
	copy(data[4128:], []byte{1, 0, 0, 4, 4, 4, 16, 8, 1, 0, 0, 0})
	k, err := cdnKey(indexVectors[tag-1][3])
	if err != nil {
		t.Fatal(err)
	}
	return data, k
}

func TestArchiveIndexLiteralLocationsAndIntegrity(t *testing.T) {
	data, k := literalIndex(t, 1)
	index, err := parseArchiveIndex(data, k)
	if err != nil {
		t.Fatal(err)
	}
	location, ok := index.locate(key{1})
	if !ok || location.archive != k || location.offset != 64 || location.size != 32 {
		t.Fatalf("location = %+v, found = %v", location, ok)
	}
	if _, ok := index.locate(key{2}); ok {
		t.Fatal("unlisted encoding key resolved")
	}
	for _, position := range []int{0, 100, 4096, 4112, 4120, 4128, 4140} {
		changed := append([]byte(nil), data...)
		changed[position] ^= 1
		if _, err := parseArchiveIndex(changed, k); err == nil {
			t.Fatalf("mutation at byte %d was accepted", position)
		}
	}
	for _, size := range []int{0, 27, len(data) - 1} {
		if _, err := parseArchiveIndex(data[:size], k); err == nil {
			t.Fatalf("truncation to %d bytes was accepted", size)
		}
	}
}

// This fixture writer belongs only to HTTP tests. Parser correctness above is
// checked against fixed independent vectors rather than these computed hashes.
func httpIndex(k key, size uint32) ([]byte, key) {
	data := make([]byte, 4148)
	copy(data[:16], k[:])
	binary.BigEndian.PutUint32(data[16:20], size)
	binary.BigEndian.PutUint32(data[20:24], 64)
	copy(data[4096:4112], k[:])
	hash := md5.Sum(data[:4096])
	copy(data[4112:4120], hash[:8])
	hash = md5.Sum(data[4096:4120])
	copy(data[4120:4128], hash[:8])
	copy(data[4128:], []byte{1, 0, 0, 4, 4, 4, 16, 8, 1, 0, 0, 0})
	footerInput := append(append([]byte(nil), data[4128:4140]...), make([]byte, 8)...)
	hash = md5.Sum(footerInput)
	copy(data[4140:], hash[:8])
	return data, key(md5.Sum(data[4120:]))
}
