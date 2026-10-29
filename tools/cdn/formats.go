package main

import (
	"bytes"
	"compress/zlib"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"math/bits"
	"slices"
	"strings"
)

type key [16]byte

type encodingEntry struct {
	Size uint64
	Keys []key
}
type encodingTable struct {
	data     []byte
	pageSize int
	index    []key
	pages    int
}

// decodeBLTE accepts the unencrypted N/Z encodings used by the UI catalogue.
// Unknown/encrypted modes fail; unavailable plaintext is never filled with zeros.
// expected == 0 means unknown; the acquisition boundary independently checks the
// content key, including legitimately empty files.
func decodeBLTE(data []byte, limit, expected uint64) ([]byte, error) {
	if limit > maxMetadata || expected > limit {
		return nil, fmt.Errorf("BLTE: invalid header or size limit")
	}
	header, count, decoded, err := blteLayout(data)
	if err != nil {
		return nil, err
	}
	if header == 0 {
		result, err := decodeChunk(data[8:], limit)
		if err != nil {
			return nil, err
		}
		if expected != 0 && uint64(len(result)) != expected {
			return nil, fmt.Errorf("BLTE: decoded size %d, expected %d", len(result), expected)
		}
		return result, nil
	}
	if decoded > limit || expected != 0 && decoded != expected {
		return nil, fmt.Errorf("BLTE: complete encoded/decoded size mismatch")
	}
	output := make([]byte, 0, int(decoded))
	cursor := int(header)
	for n := range int(count) {
		p := 12 + 24*n
		compressed := int(binary.BigEndian.Uint32(data[p : p+4]))
		plain := uint64(binary.BigEndian.Uint32(data[p+4 : p+8]))
		block := data[cursor : cursor+compressed]
		result, err := decodeChunk(block, plain)
		if err != nil {
			return nil, fmt.Errorf("BLTE chunk %d: %w", n, err)
		}
		if uint64(len(result)) != plain {
			return nil, fmt.Errorf("BLTE: chunk %d decoded size mismatch", n)
		}
		output = append(output, result...)
		cursor += compressed
	}
	return output, nil
}

// Chunked EKeys identify the header only. Verify the complete envelope before
// caching so intact headers cannot conceal damaged compressed payloads.
func verifyBLTEEnvelope(data []byte) error {
	_, _, _, err := blteLayout(data)
	return err
}

func blteLayout(data []byte) (header, count, decoded uint64, err error) {
	if len(data) < 9 || len(data) > maxMetadata || string(data[:4]) != "BLTE" {
		return 0, 0, 0, fmt.Errorf("BLTE: invalid header or encoded size limit")
	}
	header = uint64(binary.BigEndian.Uint32(data[4:8]))
	if header == 0 {
		return 0, 0, 0, nil
	}
	if len(data) < 12 || data[8] != 15 {
		return 0, 0, 0, fmt.Errorf("BLTE: unsupported chunk table")
	}
	count = uint64(data[9])<<16 | uint64(data[10])<<8 | uint64(data[11])
	if count == 0 || count > 1<<20 || header != 12+24*count || header > uint64(len(data)) {
		return 0, 0, 0, fmt.Errorf("BLTE: invalid chunk table length")
	}
	cursor := header
	for n := range int(count) {
		p := 12 + 24*n
		compressed := uint64(binary.BigEndian.Uint32(data[p : p+4]))
		decoded += uint64(binary.BigEndian.Uint32(data[p+4 : p+8]))
		if compressed == 0 || compressed > uint64(len(data))-cursor || decoded > maxMetadata {
			return 0, 0, 0, fmt.Errorf("BLTE: chunk %d exceeds bounds", n)
		}
		checksum := md5.Sum(data[cursor : cursor+compressed])
		if !bytes.Equal(checksum[:], data[p+8:p+24]) {
			return 0, 0, 0, fmt.Errorf("BLTE: chunk %d checksum mismatch", n)
		}
		cursor += compressed
	}
	if cursor != uint64(len(data)) {
		return 0, 0, 0, fmt.Errorf("BLTE: complete encoded size mismatch")
	}
	return header, count, decoded, nil
}

func decodeChunk(block []byte, limit uint64) ([]byte, error) {
	if len(block) == 0 {
		return nil, fmt.Errorf("BLTE: empty encoded chunk")
	}
	switch block[0] {
	case 'N':
		if uint64(len(block)-1) > limit {
			return nil, fmt.Errorf("BLTE: plain chunk exceeds decoded limit")
		}
		return bytes.Clone(block[1:]), nil
	case 'Z':
		source := bytes.NewReader(block[1:])
		reader, err := zlib.NewReader(source)
		if err != nil {
			return nil, fmt.Errorf("BLTE zlib header: %w", err)
		}
		result, readErr := io.ReadAll(io.LimitReader(reader, int64(limit)+1))
		closeErr := reader.Close()
		if readErr != nil {
			return nil, fmt.Errorf("BLTE zlib data: %w", readErr)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("BLTE zlib close: %w", closeErr)
		}
		if uint64(len(result)) > limit {
			return nil, fmt.Errorf("BLTE: compressed chunk exceeds decoded limit")
		}
		if source.Len() != 0 {
			return nil, fmt.Errorf("BLTE: trailing zlib bytes")
		}
		return result, nil
	default:
		return nil, fmt.Errorf("BLTE: unsupported chunk mode %q", block[0])
	}
}

func uint40(data []byte) uint64 {
	return uint64(data[0])<<32 | uint64(binary.BigEndian.Uint32(data[1:5]))
}

// parseEncoding takes ownership of data: callers must not mutate it while the
// table exists. Page checksums are checked once before bounded binary lookups.
func parseEncoding(data []byte) (*encodingTable, error) {
	if len(data) < 22 || len(data) > maxMetadata || string(data[:2]) != "EN" || data[2] != 1 || data[3] != 16 || data[4] != 16 || data[17] != 0 {
		return nil, fmt.Errorf("encoding: unsupported header")
	}
	contentSize := uint64(binary.BigEndian.Uint16(data[5:7])) * 1024
	encodedSize := uint64(binary.BigEndian.Uint16(data[7:9])) * 1024
	contentCount := uint64(binary.BigEndian.Uint32(data[9:13]))
	encodedCount := uint64(binary.BigEndian.Uint32(data[13:17]))
	specSize := uint64(binary.BigEndian.Uint32(data[18:22]))
	if contentSize == 0 || encodedSize == 0 || contentSize > 1<<20 || encodedSize > 1<<20 || contentCount == 0 || encodedCount == 0 {
		return nil, fmt.Errorf("encoding: invalid page dimensions")
	}
	contentIndex := uint64(22) + specSize
	contentPages := contentIndex + contentCount*32
	encodedIndex := contentPages + contentCount*contentSize
	encodedPages := encodedIndex + encodedCount*32
	end := encodedPages + encodedCount*encodedSize
	if end > uint64(len(data)) || specSize == 0 {
		return nil, fmt.Errorf("encoding: truncated tables")
	}
	specs := data[22:contentIndex]
	if specs[len(specs)-1] != 0 {
		return nil, fmt.Errorf("encoding: unterminated encoding specifications")
	}
	specCount := bytes.Count(specs, []byte{0})
	trailer := data[end:]
	if len(trailer) > 64<<10 {
		return nil, fmt.Errorf("encoding: oversized self-encoding specification")
	}
	for _, b := range trailer {
		if b < 32 || b > 126 {
			return nil, fmt.Errorf("encoding: invalid self-encoding specification")
		}
	}
	table := &encodingTable{data: data, pageSize: int(contentSize), pages: int(contentPages), index: make([]key, int(contentCount))}
	var previous key
	for n := range int(contentCount) {
		entry := data[int(contentIndex)+32*n : int(contentIndex)+32*(n+1)]
		first := key(entry[:16])
		page := data[int(contentPages)+n*int(contentSize) : int(contentPages)+(n+1)*int(contentSize)]
		checksum := md5.Sum(page)
		if !bytes.Equal(checksum[:], entry[16:]) {
			return nil, fmt.Errorf("encoding: content page %d checksum mismatch", n)
		}
		if n > 0 && bytes.Compare(previous[:], first[:]) >= 0 {
			return nil, fmt.Errorf("encoding: unsorted content page index")
		}
		table.index[n] = first
		offset := 0
		var last key
		records := 0
		for offset < len(page) && page[offset] != 0 {
			count := int(page[offset])
			length := 22 + 16*count
			if offset+length > len(page) {
				return nil, fmt.Errorf("encoding: content page %d truncated record", n)
			}
			current := key(page[offset+6 : offset+22])
			if records == 0 && current != first || records > 0 && bytes.Compare(last[:], current[:]) >= 0 {
				return nil, fmt.Errorf("encoding: content page %d record order/index mismatch", n)
			}
			if n > 0 && records == 0 && bytes.Compare(previous[:], current[:]) >= 0 {
				return nil, fmt.Errorf("encoding: content page ranges overlap")
			}
			for i := range count {
				if key(page[offset+22+16*i:offset+38+16*i]) == (key{}) {
					return nil, fmt.Errorf("encoding: empty encoding key")
				}
			}
			last = current
			offset += length
			records++
		}
		if records == 0 || !allZero(page[offset:]) {
			return nil, fmt.Errorf("encoding: content page %d invalid padding", n)
		}
		previous = last
	}
	var lastEncoded key
	for n := range int(encodedCount) {
		entry := data[int(encodedIndex)+32*n : int(encodedIndex)+32*(n+1)]
		first := key(entry[:16])
		page := data[int(encodedPages)+n*int(encodedSize) : int(encodedPages)+(n+1)*int(encodedSize)]
		checksum := md5.Sum(page)
		if !bytes.Equal(checksum[:], entry[16:]) {
			return nil, fmt.Errorf("encoding: encoding page %d checksum mismatch", n)
		}
		offset := 0
		records := 0
		for offset+25 <= len(page) {
			current := key(page[offset : offset+16])
			if current == (key{}) {
				// The final encoding page can terminate with an empty key,
				// specification index -1 and zero size before zero padding.
				if binary.BigEndian.Uint32(page[offset+16:offset+20]) == math.MaxUint32 && allZero(page[offset+20:offset+25]) {
					offset += 25
				}
				break
			}
			if records == 0 && current != first || (n > 0 || records > 0) && bytes.Compare(lastEncoded[:], current[:]) >= 0 {
				return nil, fmt.Errorf("encoding: encoding page %d record order/index mismatch", n)
			}
			spec := binary.BigEndian.Uint32(page[offset+16 : offset+20])
			if uint64(spec) >= uint64(specCount) {
				return nil, fmt.Errorf("encoding: invalid encoding specification index")
			}
			lastEncoded = current
			offset += 25
			records++
		}
		if records == 0 || !allZero(page[offset:]) {
			return nil, fmt.Errorf("encoding: encoding page %d invalid padding", n)
		}
	}
	return table, nil
}

func allZero(data []byte) bool {
	for _, b := range data {
		if b != 0 {
			return false
		}
	}
	return true
}

func (table *encodingTable) lookup(content key) (encodingEntry, error) {
	// The page index holds each page's first key; the candidate page is the
	// last one whose first key is not after content.
	n, found := slices.BinarySearchFunc(table.index, content, func(first, target key) int { return bytes.Compare(first[:], target[:]) })
	if !found {
		n--
	}
	if n < 0 {
		return encodingEntry{}, fmt.Errorf("encoding: content key %x not found", content)
	}
	page := table.data[table.pages+n*table.pageSize : table.pages+(n+1)*table.pageSize]
	for offset := 0; offset < len(page) && page[offset] != 0; {
		count := int(page[offset])
		current := key(page[offset+6 : offset+22])
		if current == content {
			result := encodingEntry{Size: uint40(page[offset+1 : offset+6]), Keys: make([]key, count)}
			for i := range count {
				result.Keys[i] = key(page[offset+22+16*i : offset+38+16*i])
			}
			return result, nil
		}
		if bytes.Compare(current[:], content[:]) > 0 {
			break
		}
		offset += 22 + 16*count
	}
	return encodingEntry{}, fmt.Errorf("encoding: content key %x not found", content)
}

// scanRoot validates a complete TSFM versions 1/2 manifest but retains only
// requested IDs and names. Older/unknown layouts fail rather than being guessed.
// Each retained value is the unique eligible enUS content identity.
func scanRoot(data []byte, ids map[uint32]bool, hashes map[uint64]bool) (map[uint32]key, map[uint64]key, error) {
	byID := make(map[uint32]key)
	byHash := make(map[uint64]key)
	if len(data) < 20 || len(data) > maxMetadata || string(data[:4]) != "TSFM" {
		return nil, nil, fmt.Errorf("root: unsupported header")
	}
	offset := int(binary.LittleEndian.Uint32(data[4:8]))
	version := binary.LittleEndian.Uint32(data[8:12])
	if offset != 20 && offset != 24 || version != 1 && version != 2 || len(data) < offset || offset == 24 && !allZero(data[20:24]) {
		return nil, nil, fmt.Errorf("root: unsupported version/header")
	}
	total := binary.LittleEndian.Uint32(data[12:16])
	named := binary.LittleEndian.Uint32(data[16:20])
	var actualTotal, actualNamed uint64
	blocks := 0
	for offset < len(data) {
		header := 12
		if version == 2 {
			header = 17
		}
		if len(data)-offset < header {
			return nil, nil, fmt.Errorf("root: truncated block %d header", blocks)
		}
		count := uint64(binary.LittleEndian.Uint32(data[offset : offset+4]))
		var locale, flags uint32
		if version == 2 {
			locale = binary.LittleEndian.Uint32(data[offset+4 : offset+8])
			flags = binary.LittleEndian.Uint32(data[offset+8:offset+12]) | binary.LittleEndian.Uint32(data[offset+12:offset+16]) | uint32(data[offset+16])<<17
		} else {
			flags = binary.LittleEndian.Uint32(data[offset+4 : offset+8])
			locale = binary.LittleEndian.Uint32(data[offset+8 : offset+12])
		}
		offset += header
		hasNames := flags&0x10000000 == 0
		stride := uint64(20)
		if hasNames {
			stride = 28
		}
		size := count * stride
		if count == 0 || size > uint64(len(data)-offset) {
			return nil, nil, fmt.Errorf("root: truncated or empty block %d", blocks)
		}
		actualTotal += count
		if hasNames {
			actualNamed += count
		}
		if actualTotal > uint64(total) || actualNamed > uint64(named) {
			return nil, nil, fmt.Errorf("root: declared entry counts exceeded")
		}
		hashesOffset := offset + int(count)*20
		keysOffset := offset + int(count)*4
		eligible := locale&2 != 0 && flags&(0x80|0x100) == 0
		var next uint64
		for i := range int(count) {
			id := next + uint64(binary.LittleEndian.Uint32(data[offset+4*i:offset+4*i+4]))
			if id > math.MaxUint32 {
				return nil, nil, fmt.Errorf("root: file-data ID overflow")
			}
			next = id + 1
			content := key(data[keysOffset+16*i : keysOffset+16*(i+1)])
			if content == (key{}) {
				return nil, nil, fmt.Errorf("root: empty content key")
			}
			if !eligible {
				continue
			}
			if ids[uint32(id)] {
				if old, ok := byID[uint32(id)]; ok && old != content {
					return nil, nil, fmt.Errorf("root: ambiguous enUS file-data ID %d", id)
				}
				byID[uint32(id)] = content
			}
			if hasNames {
				hash := binary.LittleEndian.Uint64(data[hashesOffset+8*i : hashesOffset+8*(i+1)])
				if hashes[hash] {
					if old, ok := byHash[hash]; ok && old != content {
						return nil, nil, fmt.Errorf("root: ambiguous enUS path hash %x", hash)
					}
					byHash[hash] = content
				}
			}
		}
		offset += int(size)
		blocks++
	}
	if actualTotal != uint64(total) || actualNamed != uint64(named) {
		return nil, nil, fmt.Errorf("root: complete entry counts differ")
	}
	for id := range ids {
		if _, ok := byID[id]; !ok {
			return nil, nil, fmt.Errorf("root: missing enUS file-data ID %d", id)
		}
	}
	for hash := range hashes {
		if _, ok := byHash[hash]; !ok {
			return nil, nil, fmt.Errorf("root: missing enUS path hash %x", hash)
		}
	}
	return byID, byHash, nil
}

// pathHash is lookup3/hashlittle2's (pc<<32)|pb over canonical WoW paths.
func pathHash(path string) uint64 {
	data := []byte(strings.ToUpper(strings.ReplaceAll(path, "/", "\\")))
	a := uint32(0xdeadbeef) + uint32(len(data))
	b, c := a, a
	for len(data) > 12 {
		a += binary.LittleEndian.Uint32(data[:4])
		b += binary.LittleEndian.Uint32(data[4:8])
		c += binary.LittleEndian.Uint32(data[8:12])
		a -= c
		a ^= bits.RotateLeft32(c, 4)
		c += b
		b -= a
		b ^= bits.RotateLeft32(a, 6)
		a += c
		c -= b
		c ^= bits.RotateLeft32(b, 8)
		b += a
		a -= c
		a ^= bits.RotateLeft32(c, 16)
		c += b
		b -= a
		b ^= bits.RotateLeft32(a, 19)
		a += c
		c -= b
		c ^= bits.RotateLeft32(b, 4)
		b += a
		data = data[12:]
	}
	if len(data) == 0 {
		return uint64(c)<<32 | uint64(b)
	}
	var tail [12]byte
	copy(tail[:], data)
	a += binary.LittleEndian.Uint32(tail[:4])
	b += binary.LittleEndian.Uint32(tail[4:8])
	c += binary.LittleEndian.Uint32(tail[8:12])
	c ^= b
	c -= bits.RotateLeft32(b, 14)
	a ^= c
	a -= bits.RotateLeft32(c, 11)
	b ^= a
	b -= bits.RotateLeft32(a, 25)
	c ^= b
	c -= bits.RotateLeft32(b, 16)
	a ^= c
	a -= bits.RotateLeft32(c, 4)
	b ^= a
	b -= bits.RotateLeft32(a, 14)
	c ^= b
	c -= bits.RotateLeft32(b, 24)
	return uint64(c)<<32 | uint64(b)
}
