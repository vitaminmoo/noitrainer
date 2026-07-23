// Command petriview renders a Noita save's world chunks to a PNG and dumps a
// JSON index of everything the chunk files contain.
//
//	petriview -world ~/.../save00/world -out /tmp/save
//
// writes /tmp/save.png (the composited world) and /tmp/save.json (chunks,
// materials, rigid bodies and joints). With -html it also writes a
// self-contained explorer page that embeds both.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"noitrainer/noitadata"
	"noitrainer/noitasave"
)

type jsonWorld struct {
	Bounds    jsonRect     `json:"bounds"`
	ChunkSize int          `json:"chunkSize"`
	Chunks    []jsonChunk  `json:"chunks"`
	Materials []jsonMatUse `json:"materials"`
	Totals    jsonTotals   `json:"totals"`
}

type jsonRect struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

type jsonTotals struct {
	Chunks       int `json:"chunks"`
	Objects      int `json:"objects"`
	Joints       int `json:"joints"`
	CustomColors int `json:"customColors"`
	SolidCells   int `json:"solidCells"`
}

// jsonMatUse is a world-wide material tally, for the legend.
type jsonMatUse struct {
	Name  string `json:"name"`
	Cells int    `json:"cells"`
	Color string `json:"color"`
	Known bool   `json:"known"`
}

type jsonChunk struct {
	File       string       `json:"file"`
	X          int          `json:"x"`
	Y          int          `json:"y"`
	Version    uint32       `json:"version"`
	W          uint32       `json:"w"`
	H          uint32       `json:"h"`
	SolidCells int          `json:"solidCells"`
	Materials  []jsonMatUse `json:"materials"`
	Colors     int          `json:"customColors"`
	Objects    []jsonObject `json:"objects"`
	Joints     []jsonJoint  `json:"joints"`
}

type jsonObject struct {
	// IDs are u64 and exceed JSON's exact integer range, so they travel as
	// strings.
	ID       string  `json:"id"`
	Material string  `json:"material"`
	X        float32 `json:"x"`
	Y        float32 `json:"y"`
	Rot      float32 `json:"rot"`
	W        uint32  `json:"w"`
	H        uint32  `json:"h"`
	Pixels   int     `json:"pixels"`
	IsStatic bool    `json:"isStatic"`
	IsCircle bool    `json:"isCircle"`
	Radius   float64 `json:"radius"`
	Z        float32 `json:"z"`
}

type jsonJoint struct {
	ID    string `json:"id"`
	Kind  string `json:"kind"`
	BodyA string `json:"bodyA"`
	BodyB string `json:"bodyB"`
	Local bool   `json:"local"`
}

func jointKindName(k uint32) string {
	switch k {
	case noitasave.JointRevolute:
		return "REVOLUTE"
	case noitasave.JointWeld:
		return "WELD"
	case noitasave.JointRagdollA:
		return "RAGDOLL_A"
	case noitasave.JointRagdollB:
		return "RAGDOLL_B"
	case noitasave.JointPlainWeld:
		return "PLAIN_WELD"
	case noitasave.JointPlainRevolute:
		return "PLAIN_REVOLUTE"
	case noitasave.JointRevoluteAttachToNearby:
		return "REVOLUTE_ATTACH_TO_SURFACE"
	case noitasave.JointWeldAttachToNearbySurface:
		return "WELD_ATTACH_TO_SURFACE"
	}
	return fmt.Sprintf("UNKNOWN_0x%x", k)
}

func main() {
	world := flag.String("world", "", "path to a save00/world directory (required)")
	install := flag.String("install", "", "Noita install directory (default: auto-detect)")
	out := flag.String("out", "world", "output path prefix")
	noPhysics := flag.Bool("no-physics", false, "omit rigid bodies from the render")
	writeHTML := flag.Bool("html", false, "also write a self-contained explorer page")
	flag.Parse()

	if *world == "" {
		fmt.Fprintln(os.Stderr, "petriview: -world is required")
		flag.Usage()
		os.Exit(2)
	}

	w, problems, err := noitasave.LoadWorld(*world)
	for _, p := range problems {
		fmt.Fprintf(os.Stderr, "warning: %v\n", p)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "petriview: %v\n", err)
		os.Exit(1)
	}

	var fsys *noitadata.FS
	if *install != "" {
		fsys, err = noitadata.Open(*install)
	} else {
		fsys, err = noitadata.OpenAuto()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "petriview: opening Noita install: %v\n"+
			"pass -install if auto-detection failed\n", err)
		os.Exit(1)
	}
	defer fsys.Close()

	palette, err := noitasave.LoadPalette(fsys)
	if err != nil {
		fmt.Fprintf(os.Stderr, "petriview: %v\n", err)
		os.Exit(1)
	}

	opts := noitasave.DefaultRenderOptions()
	opts.DrawPhysics = !*noPhysics
	img := w.Render(palette, opts)

	pngPath := *out + ".png"
	if err := noitasave.WritePNG(pngPath, img); err != nil {
		fmt.Fprintf(os.Stderr, "petriview: %v\n", err)
		os.Exit(1)
	}

	index := buildIndex(w, palette)
	jsonPath := *out + ".json"
	blob, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "petriview: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(jsonPath, blob, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "petriview: %v\n", err)
		os.Exit(1)
	}

	b := img.Bounds()
	fmt.Printf("%s  %dx%d px, world (%d,%d)-(%d,%d)\n",
		pngPath, b.Dx(), b.Dy(), b.Min.X, b.Min.Y, b.Max.X, b.Max.Y)
	fmt.Printf("%s  %d chunks, %d bodies, %d joints, %d materials, %d solid cells\n",
		jsonPath, index.Totals.Chunks, index.Totals.Objects, index.Totals.Joints,
		len(index.Materials), index.Totals.SolidCells)

	if *writeHTML {
		htmlPath := *out + ".html"
		if err := writeExplorer(htmlPath, pngPath, index); err != nil {
			fmt.Fprintf(os.Stderr, "petriview: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("%s  self-contained explorer\n", htmlPath)
	}
}

func buildIndex(w *noitasave.World, p *noitasave.Palette) jsonWorld {
	b := w.Bounds()
	idx := jsonWorld{
		Bounds:    jsonRect{X: b.Min.X, Y: b.Min.Y, W: b.Dx(), H: b.Dy()},
		ChunkSize: noitasave.ChunkSize,
	}

	worldMats := map[string]int{}

	coords := make([]noitasave.ChunkCoord, 0, len(w.Chunks))
	for c := range w.Chunks {
		coords = append(coords, c)
	}
	sort.Slice(coords, func(i, j int) bool {
		if coords[i].Y != coords[j].Y {
			return coords[i].Y < coords[j].Y
		}
		return coords[i].X < coords[j].X
	})

	for _, coord := range coords {
		c := w.Chunks[coord]
		jc := jsonChunk{
			File:    filepath.Base(w.Paths[coord]),
			X:       coord.X,
			Y:       coord.Y,
			Version: c.Version,
			W:       c.Width,
			H:       c.Height,
			Colors:  len(c.CustomColors),
			// Always emit arrays, never null: the explorer iterates these
			// directly and a null would abort its script.
			Objects: []jsonObject{},
			Joints:  []jsonJoint{},
		}

		counts := map[string]int{}
		for _, cell := range c.Cells {
			if cell == 0 {
				continue
			}
			mi, _ := noitasave.CellMaterial(cell)
			if int(mi) >= len(c.MaterialNames) {
				continue
			}
			counts[c.MaterialNames[mi]]++
			jc.SolidCells++
		}
		jc.Materials = tally(counts, p)
		for name, n := range counts {
			worldMats[name] += n
		}

		bodies := map[uint64]bool{}
		for _, o := range c.PhysicsObjects {
			bodies[o.ID] = true
		}

		for _, o := range c.PhysicsObjects {
			mat := ""
			if int(o.MaterialIndex) < len(c.MaterialNames) {
				mat = c.MaterialNames[o.MaterialIndex]
			}
			jc.Objects = append(jc.Objects, jsonObject{
				ID:       strconv.FormatUint(o.ID, 10),
				Material: mat,
				X:        o.X, Y: o.Y, Rot: o.Rot,
				W: o.Width, H: o.Height, Pixels: len(o.Colors),
				IsStatic: o.IsStatic,
				IsCircle: o.CircleRadius > 0,
				Radius:   o.CircleRadius,
				Z:        o.Z,
			})
		}
		for _, j := range c.Joints {
			jc.Joints = append(jc.Joints, jsonJoint{
				ID:    strconv.FormatUint(j.ID, 10),
				Kind:  jointKindName(j.Kind),
				BodyA: strconv.FormatUint(j.BodyAID, 10),
				BodyB: strconv.FormatUint(j.BodyBID, 10),
				Local: bodies[j.BodyAID] && bodies[j.BodyBID],
			})
		}

		idx.Totals.Chunks++
		idx.Totals.Objects += len(jc.Objects)
		idx.Totals.Joints += len(jc.Joints)
		idx.Totals.CustomColors += jc.Colors
		idx.Totals.SolidCells += jc.SolidCells
		idx.Chunks = append(idx.Chunks, jc)
	}

	idx.Materials = tally(worldMats, p)
	return idx
}

// tally turns a name->cell-count map into a sorted, colour-annotated slice.
func tally(counts map[string]int, p *noitasave.Palette) []jsonMatUse {
	out := make([]jsonMatUse, 0, len(counts))
	for name, n := range counts {
		use := jsonMatUse{Name: name, Cells: n, Color: "#ff00ff"}
		if m, ok := p.Material(name); ok {
			use.Known = true
			if m.HasColor {
				use.Color = fmt.Sprintf("#%02x%02x%02x", m.Color.R, m.Color.G, m.Color.B)
			}
		}
		out = append(out, use)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Cells != out[j].Cells {
			return out[i].Cells > out[j].Cells
		}
		return out[i].Name < out[j].Name
	})
	return out
}
