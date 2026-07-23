// Package noitasave reads and writes Noita's on-disk save files.
//
// Every blob under save00/ (world/*.png_petri, world/*.bin, world_tree.bin, ...)
// shares one container written by Noita_WriteSerializedBufferToFile
// (noita.exe @ 0x0056ccc0). The container header is LITTLE-endian:
//
//	le u32 compressedSize    // == filesize - 8
//	le u32 uncompressedSize
//	u8    data[compressedSize]   // FastLZ, level from the marker in data[0]
//
// Payloads under 0x80 bytes are stored uncompressed, signalled by
// compressedSize == uncompressedSize.
//
// The payload itself is written through network_utils::ISerializer, which is
// BIG-endian, so container ints and payload ints have opposite byte order.
// Serialized aggregates use a uniform shape:
//
//	vector<T>     => be u32 count, then count elements
//	std::string   => be u32 length, then length bytes (no NUL terminator)
//
// Note this is NOT the in-memory MSVC layout (see noita/noita.hexpat for the
// 24-byte SSO std::string); serialization flattens both.
package noitasave

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// ErrShortFile is returned for files too small to hold a container header.
var ErrShortFile = errors.New("noitasave: file shorter than 8-byte container header")

// DecodeContainer unwraps a save blob, returning the decompressed payload.
func DecodeContainer(b []byte) ([]byte, error) {
	if len(b) < 8 {
		return nil, ErrShortFile
	}
	compressed := binary.LittleEndian.Uint32(b[0:4])
	uncompressed := binary.LittleEndian.Uint32(b[4:8])

	if int64(compressed)+8 != int64(len(b)) {
		return nil, fmt.Errorf("noitasave: header says %d compressed bytes but file holds %d",
			compressed, len(b)-8)
	}

	body := b[8:]

	// Stored-raw form: the game writes this for payloads under 0x80 bytes.
	if compressed == uncompressed {
		out := make([]byte, len(body))
		copy(out, body)
		return out, nil
	}

	out, err := fastlzDecompress(body, int(uncompressed))
	if err != nil {
		return nil, err
	}
	if len(out) != int(uncompressed) {
		return nil, fmt.Errorf("noitasave: decompressed %d bytes, header declared %d",
			len(out), uncompressed)
	}
	return out, nil
}

// EncodeContainer wraps a payload the way the game does: stored raw below
// 0x80 bytes, FastLZ-compressed above it.
func EncodeContainer(payload []byte) ([]byte, error) {
	var body []byte
	if len(payload) < 0x80 {
		body = payload
	} else {
		body = fastlzCompress(payload)
		if len(body) == 0 {
			return nil, errors.New("noitasave: fastlz produced an empty stream")
		}
	}

	out := make([]byte, 8, 8+len(body))
	binary.LittleEndian.PutUint32(out[0:4], uint32(len(body)))
	binary.LittleEndian.PutUint32(out[4:8], uint32(len(payload)))
	return append(out, body...), nil
}
