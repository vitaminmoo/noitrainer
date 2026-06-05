package noitadata

import (
	"io/fs"
	"path"
	"strings"
	"testing"
)

// These tests run against the user's local Noita install. They are
// skipped if no install is found so the package remains testable on
// machines without Noita.

func mustFS(t *testing.T) *FS {
	t.Helper()
	n, err := OpenAuto()
	if err != nil {
		t.Skipf("noita install not available: %v", err)
	}
	t.Cleanup(func() { n.Close() })
	return n
}

func TestWakOpensAndIndexesFiles(t *testing.T) {
	n := mustFS(t)
	if n.Wak().Len() == 0 {
		t.Fatal("wak has 0 entries")
	}
}

func TestKnownFilesReadable(t *testing.T) {
	n := mustFS(t)
	// credits.txt is the first file observed in the hex dump.
	b, err := n.ReadFile("data/credits.txt")
	if err != nil {
		t.Fatalf("ReadFile credits.txt: %v", err)
	}
	if len(b) == 0 {
		t.Fatal("credits.txt is empty")
	}
}

func TestCaseInsensitiveWakLookup(t *testing.T) {
	n := mustFS(t)
	want, err := n.ReadFile("data/credits.txt")
	if err != nil {
		t.Fatal(err)
	}
	got, err := n.ReadFile("DATA/Credits.TXT")
	if err != nil {
		t.Fatalf("uppercase lookup failed: %v", err)
	}
	if string(got) != string(want) {
		t.Fatal("case variants returned different bytes")
	}
}

func TestDiskShadowsWak(t *testing.T) {
	n := mustFS(t)
	// config.xml lives on disk at the install root. It's not in the wak
	// but confirms that top-level disk reads work.
	if _, err := n.ReadFile("config.xml"); err != nil {
		t.Fatalf("ReadFile config.xml: %v", err)
	}
}

func TestReadDirMergesDiskAndWak(t *testing.T) {
	n := mustFS(t)
	entries, err := fs.ReadDir(n, "data")
	if err != nil {
		t.Fatalf("ReadDir data: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("data/ has no entries")
	}
	// audio is present on disk; credits.txt comes from the wak only.
	saw := map[string]bool{}
	for _, e := range entries {
		saw[e.Name()] = true
	}
	if !saw["audio"] {
		t.Error("expected disk dir 'audio' in data/")
	}
	if !saw["credits.txt"] {
		t.Error("expected wak file 'credits.txt' in data/")
	}
}

func TestWalkDirCoversArchive(t *testing.T) {
	n := mustFS(t)
	count := 0
	err := fs.WalkDir(n, "data", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			count++
			if count > 100 {
				return fs.SkipAll
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if count == 0 {
		t.Fatal("WalkDir found no files under data/")
	}
}

func TestWakContainsKnownEntry(t *testing.T) {
	n := mustFS(t)
	// credits.txt is the first entry observed in the raw on-disk index
	// (pre-sort); verify it survives parsing regardless of sort order.
	if _, _, ok := n.Wak().Lookup("data/credits.txt"); !ok {
		t.Fatal("data/credits.txt missing from parsed index")
	}
}

func TestNoLeadingSlashes(t *testing.T) {
	n := mustFS(t)
	for _, e := range n.Wak().Entries() {
		if strings.HasPrefix(e.name, "/") {
			t.Fatalf("entry %q has leading slash", e.name)
		}
		if p := path.Clean(e.name); p != e.name {
			t.Fatalf("entry %q is not clean (%q)", e.name, p)
		}
	}
}
