package noitasave

// Whole-save view: every .png_petri in a save00/world directory, composited
// into one image.
//
// A chunk file is named world_<x>_<y>.png_petri, where x and y are the world
// pixel coordinates of the chunk's top-left corner (the same values the game
// builds the path from in WorldSave_LoadChunkImage). Cells are row-major from
// that corner, so world position is simply chunk origin + cell offset.

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
)

// ChunkCoord is a chunk's top-left corner in world pixels.
type ChunkCoord struct{ X, Y int }

// World is a set of loaded chunks keyed by origin.
type World struct {
	Chunks map[ChunkCoord]*Chunk
	Paths  map[ChunkCoord]string
}

var chunkNameRE = regexp.MustCompile(`^world_(-?\d+)_(-?\d+)\.png_petri$`)

// ParseChunkName extracts the origin from a chunk filename.
func ParseChunkName(name string) (ChunkCoord, bool) {
	m := chunkNameRE.FindStringSubmatch(filepath.Base(name))
	if m == nil {
		return ChunkCoord{}, false
	}
	x, err1 := strconv.Atoi(m[1])
	y, err2 := strconv.Atoi(m[2])
	if err1 != nil || err2 != nil {
		return ChunkCoord{}, false
	}
	return ChunkCoord{X: x, Y: y}, true
}

// LoadWorld reads every chunk in a save00/world directory. Chunks that fail to
// parse are reported but do not abort the load, so one bad file cannot hide
// the rest of the save.
func LoadWorld(dir string) (*World, []error, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.png_petri"))
	if err != nil {
		return nil, nil, err
	}
	sort.Strings(paths)

	w := &World{
		Chunks: make(map[ChunkCoord]*Chunk, len(paths)),
		Paths:  make(map[ChunkCoord]string, len(paths)),
	}
	var problems []error
	for _, p := range paths {
		coord, ok := ParseChunkName(p)
		if !ok {
			problems = append(problems, fmt.Errorf("%s: unrecognised chunk filename", filepath.Base(p)))
			continue
		}
		c, err := LoadChunk(p)
		if err != nil {
			problems = append(problems, err)
			continue
		}
		w.Chunks[coord] = c
		w.Paths[coord] = p
	}
	if len(w.Chunks) == 0 {
		return w, problems, fmt.Errorf("noitasave: no readable chunks in %s", dir)
	}
	return w, problems, nil
}

// Bounds returns the world-pixel rectangle covered by all loaded chunks.
func (w *World) Bounds() image.Rectangle {
	first := true
	var r image.Rectangle
	for coord, c := range w.Chunks {
		cr := image.Rect(coord.X, coord.Y, coord.X+int(c.Width), coord.Y+int(c.Height))
		if first {
			r, first = cr, false
		} else {
			r = r.Union(cr)
		}
	}
	return r
}

// RenderOptions controls what a render includes.
type RenderOptions struct {
	// DrawPhysics draws rigid bodies over the cell grid.
	DrawPhysics bool
	// Background is painted where no chunk or cell covers a pixel.
	Background color.RGBA
	// MissingMaterial is used for a cell whose material is not in the
	// palette; a zero value skips such cells instead.
	MissingMaterial color.RGBA
}

// DefaultRenderOptions draws bodies over a dark background, with unknown
// materials in magenta so they are obvious rather than invisible.
func DefaultRenderOptions() RenderOptions {
	return RenderOptions{
		DrawPhysics:     true,
		Background:      color.RGBA{A: 255},
		MissingMaterial: color.RGBA{R: 255, B: 255, A: 255},
	}
}

// Render composites the whole world into one image. The returned image's
// bounds are in world coordinates, so image pixel (x, y) is world pixel
// (x, y); use Bounds().Min to convert to a zero-based offset.
func (w *World) Render(p *Palette, opts RenderOptions) *image.RGBA {
	b := w.Bounds()
	img := image.NewRGBA(b)

	if opts.Background.A > 0 {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				img.SetRGBA(x, y, opts.Background)
			}
		}
	}

	coords := w.sortedCoords()
	for _, coord := range coords {
		w.drawChunkCells(img, coord, w.Chunks[coord], p, opts)
	}
	if opts.DrawPhysics {
		for _, coord := range coords {
			for i := range w.Chunks[coord].PhysicsObjects {
				DrawPhysicsObject(img, &w.Chunks[coord].PhysicsObjects[i])
			}
		}
	}
	return img
}

func (w *World) sortedCoords() []ChunkCoord {
	coords := make([]ChunkCoord, 0, len(w.Chunks))
	for c := range w.Chunks {
		coords = append(coords, c)
	}
	sort.Slice(coords, func(i, j int) bool {
		if coords[i].Y != coords[j].Y {
			return coords[i].Y < coords[j].Y
		}
		return coords[i].X < coords[j].X
	})
	return coords
}

func (w *World) drawChunkCells(img *image.RGBA, coord ChunkCoord, c *Chunk, p *Palette, opts RenderOptions) {
	colorIdx := c.CustomColorIndex()
	width := int(c.Width)

	for i, cell := range c.Cells {
		if cell == 0 {
			continue
		}
		x := coord.X + i%width
		y := coord.Y + i/width

		// A custom colour overrides the material's appearance entirely.
		if ci := colorIdx[i]; ci >= 0 && int(ci) < len(c.CustomColors) {
			img.SetRGBA(x, y, ColorFromCellCustom(c.CustomColors[ci]))
			continue
		}

		idx, _ := CellMaterial(cell)
		if int(idx) >= len(c.MaterialNames) {
			continue
		}
		if col, ok := p.ColorAt(c.MaterialNames[idx], x, y); ok {
			img.SetRGBA(x, y, col)
		} else if opts.MissingMaterial.A > 0 {
			img.SetRGBA(x, y, opts.MissingMaterial)
		}
	}
}

// DrawPhysicsObject stamps a body's pixel image at its saved pose, rotating
// about the image centre. Destination pixels are inverse-mapped so the result
// has no gaps.
func DrawPhysicsObject(img *image.RGBA, o *PhysicsObject) {
	iw, ih := int(o.Width), int(o.Height)
	if iw <= 0 || ih <= 0 || len(o.Colors) < iw*ih {
		return
	}

	sin, cos := math.Sincos(float64(o.Rot))
	cx, cy := float64(iw)/2, float64(ih)/2

	// Rotated half-extent of the image, so the scan covers every landing pixel.
	half := (math.Abs(float64(iw))*math.Abs(cos) + math.Abs(float64(ih))*math.Abs(sin)) / 2
	halfY := (math.Abs(float64(iw))*math.Abs(sin) + math.Abs(float64(ih))*math.Abs(cos)) / 2
	ox, oy := float64(o.X), float64(o.Y)

	minX, maxX := int(math.Floor(ox-half))-1, int(math.Ceil(ox+half))+1
	minY, maxY := int(math.Floor(oy-halfY))-1, int(math.Ceil(oy+halfY))+1

	for dy := minY; dy <= maxY; dy++ {
		for dx := minX; dx <= maxX; dx++ {
			if !(image.Point{dx, dy}).In(img.Rect) {
				continue
			}
			// Inverse-rotate this destination pixel into image space.
			rx := float64(dx) - ox
			ry := float64(dy) - oy
			sx := int(math.Floor(rx*cos + ry*sin + cx))
			sy := int(math.Floor(-rx*sin + ry*cos + cy))
			if sx < 0 || sy < 0 || sx >= iw || sy >= ih {
				continue
			}
			c := ColorFromCellCustom(o.Colors[sy*iw+sx])
			if c.A == 0 {
				continue
			}
			img.SetRGBA(dx, dy, color.RGBA{R: c.R, G: c.G, B: c.B, A: 255})
		}
	}
}

// RenderMaterialMap renders a per-pixel lookup image the same size as Render's
// output, encoding which material occupies each world pixel:
//
//	R = id & 0xff, G = id >> 8   1-based id from ids; 0 means empty
//	B = 1 when the cell carries a custom colour, else 0
//
// It is a data channel, not something to look at: a viewer can read a pixel to
// name the material under the cursor, or scan for one id to highlight every
// cell of a material. Only cells are included, not rigid bodies, so it stays
// consistent with per-material cell counts.
func (w *World) RenderMaterialMap(ids map[string]int) *image.RGBA {
	b := w.Bounds()
	img := image.NewRGBA(b)

	for _, coord := range w.sortedCoords() {
		c := w.Chunks[coord]
		colorIdx := c.CustomColorIndex()
		width := int(c.Width)

		for i, cell := range c.Cells {
			if cell == 0 {
				continue
			}
			mi, _ := CellMaterial(cell)
			if int(mi) >= len(c.MaterialNames) {
				continue
			}
			id := ids[c.MaterialNames[mi]]
			if id == 0 {
				continue
			}
			custom := uint8(0)
			if colorIdx[i] >= 0 {
				custom = 1
			}
			img.SetRGBA(coord.X+i%width, coord.Y+i/width, color.RGBA{
				R: uint8(id),
				G: uint8(id >> 8),
				B: custom,
				A: 255,
			})
		}
	}
	return img
}

// WritePNG encodes an image to a file.
func WritePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
