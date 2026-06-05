package noitadata

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"time"
)

// FS is an fs.FS rooted at the Noita install directory. Paths under
// data/ are resolved by consulting the on-disk data/ tree first and
// falling back to data/data.wak (matching Noita's own VFS behavior,
// where registered providers are consulted in order and disk wins).
// Paths outside data/ are served purely from disk.
type FS struct {
	root string
	wak  *Wak
}

// Open returns an FS rooted at the Noita install at installDir (the
// directory containing data/, tools_modding/, noita.exe, ...).
func Open(installDir string) (*FS, error) {
	clean := filepath.Clean(installDir)
	wak, err := OpenWak(filepath.Join(clean, "data", "data.wak"))
	if err != nil {
		return nil, err
	}
	return &FS{root: clean, wak: wak}, nil
}

// OpenAuto calls FindInstall and Open.
func OpenAuto() (*FS, error) {
	dir, err := FindInstall()
	if err != nil {
		return nil, err
	}
	return Open(dir)
}

// Root returns the install directory.
func (n *FS) Root() string { return n.root }

// Wak returns the underlying archive.
func (n *FS) Wak() *Wak { return n.wak }

// Close releases resources.
func (n *FS) Close() error { return n.wak.Close() }

// Open implements fs.FS. name uses forward slashes and must be valid per
// fs.ValidPath (no leading /, no ".", no "..").
func (n *FS) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}
	if name == "." {
		return n.openDir(".")
	}

	// Try disk first (mirrors the game's registration order).
	disk := filepath.Join(n.root, filepath.FromSlash(name))
	if info, err := os.Stat(disk); err == nil {
		if info.IsDir() {
			return n.openDir(name)
		}
		return os.Open(disk)
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: err}
	}

	// Fall back to the wak.
	if off, sz, ok := n.wak.Lookup(name); ok {
		return &wakFile{
			name: path.Base(name),
			sr:   io.NewSectionReader(n.wak.f, off, sz),
			size: sz,
		}, nil
	}
	if n.isDataDir(name) {
		return n.openDir(name)
	}
	return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
}

// ReadFile implements fs.ReadFileFS.
func (n *FS) ReadFile(name string) ([]byte, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}
	disk := filepath.Join(n.root, filepath.FromSlash(name))
	if b, err := os.ReadFile(disk); err == nil {
		return b, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	return n.wak.ReadFile(name)
}

// Stat implements fs.StatFS.
func (n *FS) Stat(name string) (fs.FileInfo, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "stat", Path: name, Err: fs.ErrInvalid}
	}
	if name == "." {
		return dirInfo("."), nil
	}
	disk := filepath.Join(n.root, filepath.FromSlash(name))
	if info, err := os.Stat(disk); err == nil {
		return info, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, &fs.PathError{Op: "stat", Path: name, Err: err}
	}
	if off, sz, ok := n.wak.Lookup(name); ok {
		_ = off
		return fileInfo{name: path.Base(name), size: sz}, nil
	}
	if n.isDataDir(name) {
		return dirInfo(path.Base(name)), nil
	}
	return nil, &fs.PathError{Op: "stat", Path: name, Err: fs.ErrNotExist}
}

// ReadDir implements fs.ReadDirFS, unioning disk and wak entries.
func (n *FS) ReadDir(name string) ([]fs.DirEntry, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrInvalid}
	}
	set := map[string]fs.DirEntry{}

	// Disk entries.
	diskPath := n.root
	if name != "." {
		diskPath = filepath.Join(n.root, filepath.FromSlash(name))
	}
	if entries, err := os.ReadDir(diskPath); err == nil {
		for _, e := range entries {
			set[e.Name()] = e
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: err}
	}

	// Wak entries: the wak namespace is a 1:1 subset of the fs namespace
	// (all entries begin with "data/"). The fs path is also the wak path.
	wakDir := name
	if name == "." {
		wakDir = ""
	}
	for _, child := range n.wak.DirEntries(wakDir) {
		if _, exists := set[child]; exists {
			continue
		}
		childPath := child
		if wakDir != "" {
			childPath = wakDir + "/" + child
		}
		if n.wak.IsDir(childPath) {
			set[child] = dirEntry{name: child, dir: true}
		} else if _, sz, ok := n.wak.Lookup(childPath); ok {
			set[child] = dirEntry{name: child, dir: false, size: sz}
		}
	}

	if len(set) == 0 {
		// If the directory doesn't exist on disk and isn't in the wak,
		// surface a real not-exist error.
		if _, err := os.Stat(diskPath); errors.Is(err, os.ErrNotExist) && !n.isDataDir(name) {
			return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrNotExist}
		}
	}

	out := make([]fs.DirEntry, 0, len(set))
	for _, e := range set {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name() < out[j].Name() })
	return out, nil
}

// isDataDir reports whether name names a directory reachable through
// data.wak. name is in fs.FS form (forward slashes, no leading /).
func (n *FS) isDataDir(name string) bool {
	if name == "." || name == "" {
		return true
	}
	return n.wak.IsDir(name)
}

type wakFile struct {
	name string
	sr   *io.SectionReader
	size int64
}

func (f *wakFile) Stat() (fs.FileInfo, error) { return fileInfo{name: f.name, size: f.size}, nil }
func (f *wakFile) Read(p []byte) (int, error) { return f.sr.Read(p) }
func (f *wakFile) Close() error               { return nil }

func (n *FS) openDir(name string) (fs.File, error) {
	entries, err := n.ReadDir(name)
	if err != nil {
		return nil, err
	}
	return &dirFile{name: path.Base(name), entries: entries}, nil
}

type dirFile struct {
	name    string
	entries []fs.DirEntry
	off     int
}

func (d *dirFile) Stat() (fs.FileInfo, error) { return dirInfo(d.name), nil }
func (d *dirFile) Read([]byte) (int, error) {
	return 0, &fs.PathError{Op: "read", Path: d.name, Err: errors.New("is a directory")}
}
func (d *dirFile) Close() error { return nil }
func (d *dirFile) ReadDir(count int) ([]fs.DirEntry, error) {
	remaining := len(d.entries) - d.off
	if count <= 0 {
		out := d.entries[d.off:]
		d.off = len(d.entries)
		return out, nil
	}
	if remaining == 0 {
		return nil, io.EOF
	}
	if count > remaining {
		count = remaining
	}
	out := d.entries[d.off : d.off+count]
	d.off += count
	return out, nil
}

type dirEntry struct {
	name string
	dir  bool
	size int64
}

func (d dirEntry) Name() string { return d.name }
func (d dirEntry) IsDir() bool  { return d.dir }
func (d dirEntry) Type() fs.FileMode {
	if d.dir {
		return fs.ModeDir
	}
	return 0
}
func (d dirEntry) Info() (fs.FileInfo, error) {
	if d.dir {
		return dirInfo(d.name), nil
	}
	return fileInfo{name: d.name, size: d.size}, nil
}

type fileInfo struct {
	name string
	size int64
}

func (f fileInfo) Name() string       { return f.name }
func (f fileInfo) Size() int64        { return f.size }
func (f fileInfo) Mode() fs.FileMode  { return 0o444 }
func (f fileInfo) ModTime() time.Time { return time.Time{} }
func (f fileInfo) IsDir() bool        { return false }
func (f fileInfo) Sys() any           { return nil }

type dirInfo string

func (d dirInfo) Name() string       { return string(d) }
func (d dirInfo) Size() int64        { return 0 }
func (d dirInfo) Mode() fs.FileMode  { return fs.ModeDir | 0o555 }
func (d dirInfo) ModTime() time.Time { return time.Time{} }
func (d dirInfo) IsDir() bool        { return true }
func (d dirInfo) Sys() any           { return nil }
