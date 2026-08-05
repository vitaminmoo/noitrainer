package main

import (
	"bytes"
	"container/list"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"runtime"
	"sync"
	"sync/atomic"

	"noitrainer/noitasave"
)

const (
	tileSize = 256
	// nativeZoom is the Leaflet zoom level at which one tile pixel is one
	// world pixel. Zoom 0 then covers 256*2^10 = 262144 world px per tile,
	// enough to hold any explored world in a handful of tiles.
	nativeZoom = 10
)

var (
	hlColor      = color.RGBA{R: 255, G: 60, B: 240, A: 255}
	missingColor = color.RGBA{R: 255, B: 255, A: 255}
)

type tileServer struct {
	st  *store
	pal *noitasave.Palette

	// sem bounds concurrent renders so a burst of tile requests (a fast
	// zoom-out) queues instead of thrashing the chunk cache; queued requests
	// drop out as soon as the client aborts them.
	sem chan struct{}

	mu    sync.Mutex
	cache map[string]*list.Element
	lru   *list.List // of *tileItem
	cap   int

	queued    atomic.Int64 // waiting for a render slot
	rendering atomic.Int64 // holding a render slot
	rendered  atomic.Int64 // tiles rendered since startup
	hits      atomic.Int64 // tile-cache hits since startup
}

type tileItem struct {
	key string
	png []byte
}

func newTileServer(st *store, pal *noitasave.Palette, capTiles int) *tileServer {
	return &tileServer{
		st:    st,
		pal:   pal,
		sem:   make(chan struct{}, max(2, runtime.NumCPU()/2)),
		cache: make(map[string]*list.Element),
		lru:   list.New(),
		cap:   capTiles,
	}
}

// tile returns the encoded PNG for tile (z, tx, ty). hl highlights one
// material; dim darkens everything else; phys draws rigid bodies (native
// zoom only, where their pixels are resolvable). Rendering is abandoned as
// soon as ctx is cancelled (the client scrolled away).
func (t *tileServer) tile(ctx context.Context, z, tx, ty int, hl string, dim, phys bool) ([]byte, error) {
	key := fmt.Sprintf("%d/%d/%d|%s|%v|%v", z, tx, ty, hl, dim, phys)

	t.mu.Lock()
	if el, ok := t.cache[key]; ok {
		t.lru.MoveToFront(el)
		b := el.Value.(*tileItem).png
		t.mu.Unlock()
		t.hits.Add(1)
		return b, nil
	}
	t.mu.Unlock()

	t.queued.Add(1)
	select {
	case t.sem <- struct{}{}:
		t.queued.Add(-1)
		defer func() { <-t.sem }()
	case <-ctx.Done():
		t.queued.Add(-1)
		return nil, ctx.Err()
	}

	t.rendering.Add(1)
	img, err := t.render(ctx, z, tx, ty, hl, dim, phys)
	t.rendering.Add(-1)
	if err != nil {
		return nil, err
	}
	t.rendered.Add(1)
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	b := buf.Bytes()

	t.mu.Lock()
	if _, ok := t.cache[key]; !ok {
		t.cache[key] = t.lru.PushFront(&tileItem{key: key, png: b})
		if t.lru.Len() > t.cap {
			old := t.lru.Remove(t.lru.Back()).(*tileItem)
			delete(t.cache, old.key)
		}
	}
	t.mu.Unlock()
	return b, nil
}

func (t *tileServer) render(ctx context.Context, z, tx, ty int, hl string, dim, phys bool) (*image.RGBA, error) {
	stride := 1 << (nativeZoom - z)
	x0 := tx * tileSize * stride
	y0 := ty * tileSize * stride
	worldRect := image.Rect(x0, y0, x0+tileSize*stride, y0+tileSize*stride)

	// The image lives in world coordinates at native zoom so physics bodies
	// can be stamped directly; at coarser zooms it is plain tile space.
	var img *image.RGBA
	if stride == 1 {
		img = image.NewRGBA(worldRect)
	} else {
		img = image.NewRGBA(image.Rect(0, 0, tileSize, tileSize))
	}
	if !worldRect.Overlaps(t.st.bounds) {
		return img, nil
	}

	var hlMask []bool
	if hl != "" && dim {
		hlMask = make([]bool, tileSize*tileSize)
	}

	cs := noitasave.ChunkSize
	var cur *chunkEntry
	var curCoord noitasave.ChunkCoord
	haveCur := false

	for py := range tileSize {
		if py%32 == 0 && ctx.Err() != nil {
			return nil, ctx.Err()
		}
		wy := y0 + py*stride
		cy := floorDiv(wy, cs) * cs
		for px := range tileSize {
			wx := x0 + px*stride
			coord := noitasave.ChunkCoord{X: floorDiv(wx, cs) * cs, Y: cy}
			if !haveCur || coord != curCoord {
				e, err := t.st.chunk(coord)
				if err != nil {
					e = nil // corrupt chunk: render as empty
				}
				cur, curCoord, haveCur = e, coord, true
			}
			if cur == nil {
				continue
			}
			c := cur.c
			lx, ly := wx-coord.X, wy-coord.Y
			if lx >= int(c.Width) || ly >= int(c.Height) {
				continue
			}
			i := ly*int(c.Width) + lx
			cell := c.Cells[i]
			if cell == 0 {
				continue
			}

			name := ""
			if mi, _ := noitasave.CellMaterial(cell); int(mi) < len(c.MaterialNames) {
				name = c.MaterialNames[mi]
			}

			var col color.RGBA
			if ci := cur.colorIdx[i]; ci >= 0 && int(ci) < len(c.CustomColors) {
				col = noitasave.ColorFromCellCustom(c.CustomColors[ci])
				col.A = 255
			} else if pc, ok := t.pal.ColorAt(name, wx, wy); ok {
				col = pc
			} else {
				col = missingColor
			}

			if hl != "" && name == hl {
				col = hlColor
				if hlMask != nil {
					hlMask[py*tileSize+px] = true
				}
			}

			if stride == 1 {
				img.SetRGBA(wx, wy, col)
			} else {
				img.SetRGBA(px, py, col)
			}
		}
	}

	if phys && stride == 1 {
		// Bodies can hang over their owning chunk's edge, so scan chunks a
		// little beyond the tile.
		pad := worldRect.Inset(-cs)
		for cy := floorDiv(pad.Min.Y, cs) * cs; cy < pad.Max.Y; cy += cs {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			for cx := floorDiv(pad.Min.X, cs) * cs; cx < pad.Max.X; cx += cs {
				e, err := t.st.chunk(noitasave.ChunkCoord{X: cx, Y: cy})
				if e == nil || err != nil {
					continue
				}
				for i := range e.c.PhysicsObjects {
					noitasave.DrawPhysicsObject(img, &e.c.PhysicsObjects[i])
				}
			}
		}
	}

	if hlMask != nil {
		// Match the explorer's old "dim others" look: everything that isn't
		// the highlighted material fades toward black.
		for py := range tileSize {
			for px := range tileSize {
				if hlMask[py*tileSize+px] {
					continue
				}
				o := img.PixOffset(img.Rect.Min.X+px, img.Rect.Min.Y+py)
				p := img.Pix[o : o+4 : o+4]
				if p[3] == 0 {
					continue
				}
				p[0] = uint8(uint32(p[0]) * 80 / 255)
				p[1] = uint8(uint32(p[1]) * 80 / 255)
				p[2] = uint8(uint32(p[2]) * 80 / 255)
			}
		}
	}
	return img, nil
}
