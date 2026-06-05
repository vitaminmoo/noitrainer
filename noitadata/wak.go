// Package noitadata exposes Noita's game files as a single fs.FS rooted at
// the install directory, transparently merging the on-disk tree with entries
// stored inside data/data.wak.
//
// data.wak format (verified against Noita's own WizardPak_ParseFileIndex):
//
//	offset  size  field
//	0x00    4     version / unused (observed 0)
//	0x04    4     file_count (LE u32)
//	0x08    4     first_data_offset (LE u32) - start of file data region
//	0x0C    4     padding
//	0x10    ...   entry[0], entry[1], ..., entry[file_count-1]
//	...           file data (referenced by entry.offset)
//
// Each entry:
//
//	u32 offset    // absolute byte offset into the .wak of this file's data
//	u32 size      // bytes
//	u32 name_len  // bytes
//	char name[name_len]  // not null-terminated, lowercase, forward slashes
//
// No compression, no encryption. Lookups are case-insensitive (Noita
// lowercases the requested path before searching its sorted index).
package noitadata

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// wakEntry describes a single file inside data.wak.
type wakEntry struct {
	name   string // lowercase, forward-slash separated
	offset int64
	size   int64
}

// Wak is a read-only view over a data.wak archive. Safe for concurrent use.
type Wak struct {
	f       *os.File
	size    int64
	entries []wakEntry // sorted by name

	// dirs maps a directory path (no trailing slash, "" for root) to the
	// names of its direct children (files + subdirs). Populated lazily.
	dirs map[string][]string
}

// OpenWak opens and parses the index of a data.wak archive at path.
func OpenWak(path string) (*Wak, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	st, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	w := &Wak{f: f, size: st.Size()}
	if err := w.readIndex(); err != nil {
		f.Close()
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	w.buildDirs()
	return w, nil
}

// Close releases the underlying file handle.
func (w *Wak) Close() error { return w.f.Close() }

// Len returns the number of files in the archive.
func (w *Wak) Len() int { return len(w.entries) }

func (w *Wak) readIndex() error {
	var header [16]byte
	if _, err := io.ReadFull(io.NewSectionReader(w.f, 0, 16), header[:]); err != nil {
		return fmt.Errorf("reading header: %w", err)
	}
	count := binary.LittleEndian.Uint32(header[4:8])
	firstData := int64(binary.LittleEndian.Uint32(header[8:12]))
	if firstData < 16 || firstData > w.size {
		return fmt.Errorf("implausible first_data_offset %d (file size %d)", firstData, w.size)
	}

	// Entries occupy bytes [16, firstData). Stream them sequentially.
	idx := io.NewSectionReader(w.f, 16, firstData-16)
	br := bufReader{r: idx}
	w.entries = make([]wakEntry, 0, count)
	for i := range count {
		var hdr [12]byte
		if _, err := io.ReadFull(&br, hdr[:]); err != nil {
			return fmt.Errorf("entry %d header: %w", i, err)
		}
		off := int64(binary.LittleEndian.Uint32(hdr[0:4]))
		sz := int64(binary.LittleEndian.Uint32(hdr[4:8]))
		nameLen := binary.LittleEndian.Uint32(hdr[8:12])
		if nameLen == 0 || nameLen > 1024 {
			return fmt.Errorf("entry %d: implausible name length %d", i, nameLen)
		}
		name := make([]byte, nameLen)
		if _, err := io.ReadFull(&br, name); err != nil {
			return fmt.Errorf("entry %d name: %w", i, err)
		}
		if off < firstData || off+sz > w.size {
			return fmt.Errorf("entry %d %q: data range [%d,%d) out of bounds", i, name, off, off+sz)
		}
		w.entries = append(w.entries, wakEntry{
			name:   strings.ToLower(string(name)),
			offset: off,
			size:   sz,
		})
	}
	// Noita stores entries sorted for binary search; sort defensively.
	sort.Slice(w.entries, func(i, j int) bool { return w.entries[i].name < w.entries[j].name })
	return nil
}

func (w *Wak) buildDirs() {
	w.dirs = make(map[string][]string)
	seen := make(map[string]map[string]struct{}) // dir -> set of child names
	add := func(dir, child string) {
		m, ok := seen[dir]
		if !ok {
			m = make(map[string]struct{})
			seen[dir] = m
		}
		m[child] = struct{}{}
	}
	for _, e := range w.entries {
		parts := strings.Split(e.name, "/")
		for i := range parts {
			dir := strings.Join(parts[:i], "/")
			add(dir, parts[i])
		}
	}
	for dir, set := range seen {
		names := make([]string, 0, len(set))
		for n := range set {
			names = append(names, n)
		}
		sort.Strings(names)
		w.dirs[dir] = names
	}
}

// Lookup returns the entry for name (case-insensitive) or ok=false.
func (w *Wak) Lookup(name string) (offset, size int64, ok bool) {
	key := strings.ToLower(name)
	i := sort.Search(len(w.entries), func(i int) bool { return w.entries[i].name >= key })
	if i < len(w.entries) && w.entries[i].name == key {
		return w.entries[i].offset, w.entries[i].size, true
	}
	return 0, 0, false
}

// ReadFile returns the bytes of name or ErrNotExist.
func (w *Wak) ReadFile(name string) ([]byte, error) {
	off, sz, ok := w.Lookup(name)
	if !ok {
		return nil, &os.PathError{Op: "read", Path: name, Err: os.ErrNotExist}
	}
	buf := make([]byte, sz)
	if _, err := w.f.ReadAt(buf, off); err != nil {
		return nil, err
	}
	return buf, nil
}

// ReaderAt returns a reader over name's bytes.
func (w *Wak) ReaderAt(name string) (*io.SectionReader, int64, error) {
	off, sz, ok := w.Lookup(name)
	if !ok {
		return nil, 0, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
	}
	return io.NewSectionReader(w.f, off, sz), sz, nil
}

// DirEntries returns the direct children (file or subdir names) of dir
// (case-insensitive, "" for archive root). Returns nil if dir is not a
// directory in the archive.
func (w *Wak) DirEntries(dir string) []string {
	key := strings.ToLower(strings.Trim(dir, "/"))
	names, ok := w.dirs[key]
	if !ok {
		return nil
	}
	out := make([]string, len(names))
	copy(out, names)
	return out
}

// IsDir reports whether dir exists as a directory in the archive.
func (w *Wak) IsDir(dir string) bool {
	_, ok := w.dirs[strings.ToLower(strings.Trim(dir, "/"))]
	return ok
}

// Entries returns all entries in sorted order. The returned slice aliases
// internal state and must not be modified.
func (w *Wak) Entries() []wakEntry { return w.entries }

// bufReader is a small buffered reader to avoid per-field syscalls during
// index parsing. We can't use bufio.Reader because we want to pass it to
// io.ReadFull which needs io.Reader, which bufio satisfies — but we want
// to keep allocations predictable and avoid the larger default buffer.
type bufReader struct {
	r   io.Reader
	buf [4096]byte
	n   int
	off int
}

func (b *bufReader) Read(p []byte) (int, error) {
	if b.off == b.n {
		n, err := b.r.Read(b.buf[:])
		if n == 0 {
			if err == nil {
				err = io.EOF
			}
			return 0, err
		}
		b.off, b.n = 0, n
	}
	n := copy(p, b.buf[b.off:b.n])
	b.off += n
	return n, nil
}

// sentinel to ensure errors.Is works with os.ErrNotExist for callers.
var _ = errors.Is
