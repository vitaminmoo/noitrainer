package main

import (
	"container/list"
	"fmt"
	"image"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"

	"noitrainer/noitasave"
)

// chunkEntry is a decoded chunk plus its custom-color index, which is too
// expensive to rebuild per sampled pixel.
type chunkEntry struct {
	c        *noitasave.Chunk
	colorIdx []int32
}

type lruItem struct {
	coord noitasave.ChunkCoord
	entry *chunkEntry
	err   error
}

// store maps chunk coordinates to files and keeps a bounded number of decoded
// chunks in memory, so tile rendering never needs the whole world at once.
type store struct {
	paths  map[noitasave.ChunkCoord]string
	coords []noitasave.ChunkCoord // sorted by y, then x
	bounds image.Rectangle

	mu    sync.Mutex
	cache map[noitasave.ChunkCoord]*list.Element
	lru   *list.List // of *lruItem, front = most recent
	cap   int

	loads atomic.Int64 // chunk files read since startup
}

// cached reports how many decoded chunks are currently held in memory.
func (s *store) cached() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lru.Len()
}

func openStore(dir string, capChunks int) (*store, []error, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.png_petri"))
	if err != nil {
		return nil, nil, err
	}
	sort.Strings(paths)

	s := &store{
		paths: make(map[noitasave.ChunkCoord]string, len(paths)),
		cache: make(map[noitasave.ChunkCoord]*list.Element),
		lru:   list.New(),
		cap:   capChunks,
	}
	var problems []error
	first := true
	for _, p := range paths {
		coord, ok := noitasave.ParseChunkName(p)
		if !ok {
			problems = append(problems, fmt.Errorf("%s: unrecognised chunk filename", filepath.Base(p)))
			continue
		}
		s.paths[coord] = p
		s.coords = append(s.coords, coord)
		cr := image.Rect(coord.X, coord.Y, coord.X+noitasave.ChunkSize, coord.Y+noitasave.ChunkSize)
		if first {
			s.bounds, first = cr, false
		} else {
			s.bounds = s.bounds.Union(cr)
		}
	}
	if len(s.coords) == 0 {
		return nil, problems, fmt.Errorf("no readable chunks in %s", dir)
	}
	sort.Slice(s.coords, func(i, j int) bool {
		if s.coords[i].Y != s.coords[j].Y {
			return s.coords[i].Y < s.coords[j].Y
		}
		return s.coords[i].X < s.coords[j].X
	})
	return s, problems, nil
}

// chunk returns the decoded chunk at coord, loading and caching it on demand.
// A nil entry with nil error means no chunk exists there. Load failures are
// cached too, so a corrupt chunk doesn't get re-read for every tile.
func (s *store) chunk(coord noitasave.ChunkCoord) (*chunkEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if el, ok := s.cache[coord]; ok {
		s.lru.MoveToFront(el)
		it := el.Value.(*lruItem)
		return it.entry, it.err
	}
	p, ok := s.paths[coord]
	if !ok {
		return nil, nil
	}
	s.loads.Add(1)
	c, err := noitasave.LoadChunk(p)
	var e *chunkEntry
	if err == nil {
		e = &chunkEntry{c: c, colorIdx: c.CustomColorIndex()}
	}
	s.cache[coord] = s.lru.PushFront(&lruItem{coord: coord, entry: e, err: err})
	if s.lru.Len() > s.cap {
		old := s.lru.Remove(s.lru.Back()).(*lruItem)
		delete(s.cache, old.coord)
	}
	return e, err
}

// floorDiv is integer division rounding toward negative infinity, for mapping
// negative world coordinates onto chunk origins.
func floorDiv(a, b int) int {
	q := a / b
	if a%b != 0 && (a < 0) != (b < 0) {
		q--
	}
	return q
}

// chunkOrigin returns the origin of the chunk containing world pixel (wx, wy).
func chunkOrigin(wx, wy int) noitasave.ChunkCoord {
	cs := noitasave.ChunkSize
	return noitasave.ChunkCoord{X: floorDiv(wx, cs) * cs, Y: floorDiv(wy, cs) * cs}
}
