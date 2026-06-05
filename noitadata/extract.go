package noitadata

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ExtractDir materializes the subtree rooted at srcDir (an fs.FS path,
// forward slashes) into the on-disk directory dstDir. Existing files are
// overwritten. Returns the number of files written.
//
// This is useful for clients that need real OS paths (XML tools,
// editors) — the returned tree mirrors what the game sees when reading
// that subtree.
func ExtractDir(n *FS, srcDir, dstDir string) (int, error) {
	srcDir = strings.Trim(srcDir, "/")
	if srcDir == "" {
		srcDir = "."
	}
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return 0, err
	}
	count := 0
	walkErr := fs.WalkDir(n, srcDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(p, srcDir)
		rel = strings.TrimPrefix(rel, "/")
		target := filepath.Join(dstDir, filepath.FromSlash(rel))
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		src, err := n.Open(p)
		if err != nil {
			return err
		}
		defer src.Close()
		dst, err := os.Create(target)
		if err != nil {
			return err
		}
		if _, err := io.Copy(dst, src); err != nil {
			dst.Close()
			return err
		}
		if err := dst.Close(); err != nil {
			return err
		}
		count++
		return nil
	})
	return count, walkErr
}
