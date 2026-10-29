package main

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"sort"
)

const (
	archiveIndexLimit = 8 << 20
	// archiveIndexTotal is the default budget for retained archive index
	// bytes. It is a distinct limit from maxMetadata that happens to be equal.
	archiveIndexTotal = 384 << 20
	archiveBlockSize  = 4096
	archiveEntrySize  = 24
)

type archiveLocation struct {
	archive key
	offset  uint32
	size    uint32
}

type archiveIndex struct {
	archive key
	data    []byte
	entries uint32
	blocks  int
}

// The key binds the footer through CASC's MD5 identity chain. Each
// block and the complete TOC are checked before any location is exposed.
func parseArchiveIndex(data []byte, archive key) (archiveIndex, error) {
	fail := func(reason string) (archiveIndex, error) {
		return archiveIndex{}, fmt.Errorf("archive %x index: %s", archive, reason)
	}
	if len(data) < 28 || len(data) > archiveIndexLimit {
		return fail("length outside supported bounds")
	}
	foot := data[len(data)-20:]
	if !bytes.Equal(foot[:8], []byte{1, 0, 0, 4, 4, 4, 16, 8}) {
		return fail("unsupported footer layout")
	}
	if md5.Sum(data[len(data)-28:]) != [16]byte(archive) {
		return fail("footer identity mismatch")
	}
	var footerInput [20]byte
	copy(footerInput[:12], foot[:12])
	footerHash := md5.Sum(footerInput[:])
	if !bytes.Equal(footerHash[:8], foot[12:]) {
		return fail("footer checksum mismatch")
	}
	entries := binary.LittleEndian.Uint32(foot[8:12])
	perBlock := uint64(archiveBlockSize / archiveEntrySize)
	blocks := (uint64(entries) + perBlock - 1) / perBlock
	if entries == 0 || blocks*(archiveBlockSize+24)+28 != uint64(len(data)) {
		return fail("entry count and exact layout length disagree")
	}
	tocStart := int(blocks) * archiveBlockSize
	toc := data[tocStart : len(data)-28]
	tocHash := md5.Sum(toc)
	if !bytes.Equal(tocHash[:8], data[len(data)-28:len(data)-20]) {
		return fail("table-of-contents checksum mismatch")
	}
	var previous []byte
	var zero key
	for block := range int(blocks) {
		body := data[block*archiveBlockSize : (block+1)*archiveBlockSize]
		hash := md5.Sum(body)
		checksum := toc[int(blocks)*16+block*8 : int(blocks)*16+(block+1)*8]
		if !bytes.Equal(hash[:8], checksum) {
			return fail(fmt.Sprintf("block %d checksum mismatch", block))
		}
		count := min(int(perBlock), int(entries)-block*int(perBlock))
		for entry := range count {
			record := body[entry*archiveEntrySize : (entry+1)*archiveEntrySize]
			k := record[:16]
			if bytes.Equal(k, zero[:]) || (previous != nil && bytes.Compare(previous, k) > 0) {
				return fail("empty or unsorted encoding key")
			}
			size := binary.BigEndian.Uint32(record[16:20])
			offset := binary.BigEndian.Uint32(record[20:24])
			if size == 0 || uint64(offset)+uint64(size) > 1<<31-1 {
				return fail("entry range exceeds supported archive bounds")
			}
			previous = k
		}
		if !bytes.Equal(previous, toc[block*16:(block+1)*16]) {
			return fail(fmt.Sprintf("block %d TOC end key mismatch", block))
		}
		for _, value := range body[count*archiveEntrySize:] {
			if value != 0 {
				return fail(fmt.Sprintf("block %d has nonzero padding", block))
			}
		}
	}
	return archiveIndex{archive: archive, data: data, entries: entries, blocks: int(blocks)}, nil
}

func (index archiveIndex) locate(k key) (archiveLocation, bool) {
	toc := index.data[index.blocks*archiveBlockSize:]
	block := sort.Search(index.blocks, func(i int) bool {
		return bytes.Compare(toc[i*16:(i+1)*16], k[:]) >= 0
	})
	if block == index.blocks {
		return archiveLocation{}, false
	}
	perBlock := archiveBlockSize / archiveEntrySize
	count := min(perBlock, int(index.entries)-block*perBlock)
	body := index.data[block*archiveBlockSize : (block+1)*archiveBlockSize]
	entry := sort.Search(count, func(i int) bool {
		return bytes.Compare(body[i*archiveEntrySize:i*archiveEntrySize+16], k[:]) >= 0
	})
	if entry == count || !bytes.Equal(body[entry*archiveEntrySize:entry*archiveEntrySize+16], k[:]) {
		return archiveLocation{}, false
	}
	record := body[entry*archiveEntrySize : (entry+1)*archiveEntrySize]
	return archiveLocation{archive: index.archive, size: binary.BigEndian.Uint32(record[16:20]),
		offset: binary.BigEndian.Uint32(record[20:24])}, true
}
