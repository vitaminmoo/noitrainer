package noitasave

import (
	"errors"
	"fmt"
)

// FastLZ (Ariya Hidayat's reference implementation, the version Noita statically
// links) ported to Go. Noita compresses save blobs with fastlz_compress, which
// picks level 1 for payloads < 64 KiB and level 2 above that; the decompressor
// dispatches on a marker in the top 3 bits of the first byte, so both are
// implemented here.
//
// Compression reproduces the reference encoder's output byte-for-byte, so a
// re-written save is indistinguishable from one the game wrote itself
// (cross-checked against the C reference in fastlz_test.go).

const (
	flzMaxCopy     = 32
	flzMaxLen      = 264 // 256 + 8
	flzMaxDistance = 8192

	// Level 2 shortens the near-window by one and adds a far window.
	flzMaxDistanceL2  = 8191
	flzMaxFarDistance = 65535 + flzMaxDistanceL2 - 1

	flzHashLog  = 13
	flzHashSize = 1 << flzHashLog
	flzHashMask = flzHashSize - 1
)

var (
	errFastlzTruncated = errors.New("fastlz: compressed stream truncated")
	errFastlzOverflow  = errors.New("fastlz: output exceeds declared size")
	errFastlzCorrupt   = errors.New("fastlz: back-reference before start of output")
)

// fastlzLevel reports the compression level encoded in the stream's first byte.
func fastlzLevel(in []byte) (int, error) {
	if len(in) == 0 {
		return 0, errFastlzTruncated
	}
	level := int(in[0]>>5) + 1
	if level != 1 && level != 2 {
		return 0, fmt.Errorf("fastlz: unsupported compression level %d", level)
	}
	return level, nil
}

// fastlzDecompress expands in into at most maxout bytes. maxout comes from the
// container header, so a well-formed stream lands exactly on it.
func fastlzDecompress(in []byte, maxout int) ([]byte, error) {
	level, err := fastlzLevel(in)
	if err != nil {
		return nil, err
	}

	out := make([]byte, 0, maxout)
	ip, ipLimit := 0, len(in)
	ctrl := int(in[ip]) & 31
	ip++

	for loop := true; loop; {
		if ctrl >= 32 {
			// Back-reference: length in the top 3 bits, distance split
			// across the low 5 bits and follow-up bytes.
			length := (ctrl >> 5) - 1
			ofs := (ctrl & 31) << 8
			ref := len(out) - ofs

			if length == 7-1 {
				if level == 1 {
					if ip >= ipLimit {
						return nil, errFastlzTruncated
					}
					length += int(in[ip])
					ip++
				} else {
					for {
						if ip >= ipLimit {
							return nil, errFastlzTruncated
						}
						code := int(in[ip])
						ip++
						length += code
						if code != 255 {
							break
						}
					}
				}
			}

			if ip >= ipLimit {
				return nil, errFastlzTruncated
			}
			code := int(in[ip])
			ip++
			ref -= code

			// Level 2 escapes to a 16-bit far distance.
			if level == 2 && code == 255 && ofs == 31<<8 {
				if ip+1 >= ipLimit {
					return nil, errFastlzTruncated
				}
				ofs = int(in[ip]) << 8
				ip++
				ofs += int(in[ip])
				ip++
				ref = len(out) - ofs - flzMaxDistanceL2
			}

			if len(out)+length+3 > maxout {
				return nil, errFastlzOverflow
			}
			if ref-1 < 0 {
				return nil, errFastlzCorrupt
			}

			if ip < ipLimit {
				ctrl = int(in[ip])
				ip++
			} else {
				loop = false
			}

			if ref == len(out) {
				// Distance 0 encodes a run of the previous byte.
				b := out[ref-1]
				for i := 0; i < length+3; i++ {
					out = append(out, b)
				}
			} else {
				ref--
				// Byte-at-a-time: source and destination may overlap.
				for i := 0; i < length+3; i++ {
					out = append(out, out[ref])
					ref++
				}
			}
		} else {
			// Literal run of ctrl+1 bytes.
			n := ctrl + 1
			if len(out)+n > maxout {
				return nil, errFastlzOverflow
			}
			if ip+n > ipLimit {
				return nil, errFastlzTruncated
			}
			out = append(out, in[ip:ip+n]...)
			ip += n

			loop = ip < ipLimit
			if loop {
				ctrl = int(in[ip])
				ip++
			}
		}
	}
	return out, nil
}

func flzReadU16(b []byte, i int) int { return int(b[i]) | int(b[i+1])<<8 }

func flzHash(b []byte, i int) int {
	v := flzReadU16(b, i)
	v ^= flzReadU16(b, i+1) ^ (v >> (16 - flzHashLog))
	return v & flzHashMask
}

// fastlzCompress compresses in with the level the game itself would choose:
// level 1 below 64 KiB, level 2 at or above it.
func fastlzCompress(in []byte) []byte {
	if len(in) < 65536 {
		return fastlzCompressLevel(1, in)
	}
	return fastlzCompressLevel(2, in)
}

// fastlzCompressLevel is a direct port of the reference FASTLZ_COMPRESSOR.
// It reproduces the reference encoder's byte stream exactly.
func fastlzCompressLevel(level int, in []byte) []byte {
	length := len(in)

	// Runt inputs are emitted as a single literal run.
	if length < 4 {
		if length == 0 {
			return nil
		}
		out := make([]byte, 0, length+1)
		out = append(out, byte(length-1))
		out = append(out, in...)
		return out
	}

	maxDistance := flzMaxDistance
	maxFarDistance := 0
	if level == 2 {
		maxDistance = flzMaxDistanceL2
		maxFarDistance = flzMaxFarDistance
	}

	out := make([]byte, 0, length+length/16+64)

	htab := make([]int, flzHashSize)
	ipBound := length - 2
	ipLimit := length - 12
	ip := 0

	// The reference seeds every hash slot with the input start.
	for i := range htab {
		htab[i] = 0
	}

	// Start with a two-byte literal run; the count byte is patched later.
	copyCount := 2
	out = append(out, flzMaxCopy-1)
	out = append(out, in[ip])
	ip++
	out = append(out, in[ip])
	ip++

	for ip < ipLimit {
		anchor := ip
		matchLen := 3
		ref := 0
		distance := 0
		matched := false

		if level == 2 {
			// A run of one repeated byte is its own (distance 1) match.
			if in[ip] == in[ip-1] && flzReadU16(in, ip-1) == flzReadU16(in, ip+1) {
				distance = 1
				ref = anchor - 1 + 3
				matched = true
			}
		}

		if !matched {
			hval := flzHash(in, ip)
			ref = htab[hval]
			distance = anchor - ref
			htab[hval] = anchor

			limit := maxDistance
			if level == 2 {
				limit = maxFarDistance
			}

			ok := distance != 0 && distance < limit &&
				in[ref] == in[ip] && in[ref+1] == in[ip+1] && in[ref+2] == in[ip+2]

			if ok && level == 2 && distance >= flzMaxDistanceL2 {
				// Far matches must be at least 5 bytes to be worth encoding.
				if in[ref+3] != in[ip+3] || in[ref+4] != in[ip+4] {
					ok = false
				} else {
					matchLen += 2
				}
			}

			if !ok {
				// Literal byte; retry from the next position.
				out = append(out, in[anchor])
				ip = anchor + 1
				copyCount++
				if copyCount == flzMaxCopy {
					copyCount = 0
					out = append(out, flzMaxCopy-1)
				}
				continue
			}
			ref += matchLen
			matched = true
		}

		// Extend the match. Both cursors sit matchLen bytes past their starts.
		ip = anchor + matchLen
		distance--

		if distance == 0 {
			x := in[ip-1]
			for ip < ipBound && in[ref] == x {
				ref++
				ip++
			}
		} else {
			// The reference unrolls 8 unchecked compares, then bounds-checks;
			// clamping to ipBound throughout is equivalent for valid input and
			// cannot read out of range.
			for n := 0; n < 8; n++ {
				if ref >= length || ip >= length || in[ref] != in[ip] {
					ref++
					ip++
					goto encode
				}
				ref++
				ip++
			}
			for ip < ipBound {
				if in[ref] != in[ip] {
					ref++
					ip++
					break
				}
				ref++
				ip++
			}
		}

	encode:
		// Patch (or drop) the pending literal-run count.
		if copyCount != 0 {
			out[len(out)-copyCount-1] = byte(copyCount - 1)
		} else {
			out = out[:len(out)-1]
		}
		copyCount = 0

		ip -= 3
		mlen := ip - anchor

		if level == 1 {
			for mlen > flzMaxLen-2 {
				out = append(out, byte((7<<5)+(distance>>8)))
				out = append(out, byte(flzMaxLen-2-7-2))
				out = append(out, byte(distance&255))
				mlen -= flzMaxLen - 2
			}
			if mlen < 7 {
				out = append(out, byte((mlen<<5)+(distance>>8)))
				out = append(out, byte(distance&255))
			} else {
				out = append(out, byte((7<<5)+(distance>>8)))
				out = append(out, byte(mlen-7))
				out = append(out, byte(distance&255))
			}
		} else {
			if distance < flzMaxDistanceL2 {
				if mlen < 7 {
					out = append(out, byte((mlen<<5)+(distance>>8)))
					out = append(out, byte(distance&255))
				} else {
					out = append(out, byte((7<<5)+(distance>>8)))
					for mlen -= 7; mlen >= 255; mlen -= 255 {
						out = append(out, 255)
					}
					out = append(out, byte(mlen))
					out = append(out, byte(distance&255))
				}
			} else {
				if mlen < 7 {
					distance -= flzMaxDistanceL2
					out = append(out, byte((mlen<<5)+31))
					out = append(out, 255)
					out = append(out, byte(distance>>8))
					out = append(out, byte(distance&255))
				} else {
					distance -= flzMaxDistanceL2
					out = append(out, byte((7<<5)+31))
					for mlen -= 7; mlen >= 255; mlen -= 255 {
						out = append(out, 255)
					}
					out = append(out, byte(mlen))
					out = append(out, 255)
					out = append(out, byte(distance>>8))
					out = append(out, byte(distance&255))
				}
			}
		}

		// Refresh the hash at the match boundary. Skipping the update when the
		// 3-byte hash window would run off the end is safe: that only happens
		// past ipLimit, where the outer loop is about to exit anyway.
		if ip+2 < length {
			htab[flzHash(in, ip)] = ip
		}
		ip++
		if ip+2 < length {
			htab[flzHash(in, ip)] = ip
		}
		ip++

		// Assume a literal run follows; the count byte is patched later.
		out = append(out, flzMaxCopy-1)
	}

	// Flush the tail as literals. The reference widens ipBound by one here,
	// after the match loops have finished using the narrower bound.
	ipBound++
	for ip <= ipBound {
		out = append(out, in[ip])
		ip++
		copyCount++
		if copyCount == flzMaxCopy {
			copyCount = 0
			out = append(out, flzMaxCopy-1)
		}
	}

	if copyCount != 0 {
		out[len(out)-copyCount-1] = byte(copyCount - 1)
	} else {
		out = out[:len(out)-1]
	}

	if level == 2 {
		out[0] |= 1 << 5 // level-2 marker
	}
	return out
}
