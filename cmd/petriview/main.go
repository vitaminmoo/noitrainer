// Command petriview serves a Noita save's world chunks as a browsable map.
//
//	petriview -world ~/.../save00/world
//
// then open http://127.0.0.1:8940. Tiles are rendered on demand from the
// .png_petri chunks (zoomed-out levels sample every Nth world pixel), so
// memory stays bounded no matter how much of the world has been explored.
package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"noitrainer/noitadata"
	"noitrainer/noitasave"
)

//go:embed assets
var assets embed.FS

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
	addr := flag.String("addr", "127.0.0.1:8940", "HTTP listen address")
	flag.Parse()

	if *world == "" {
		fmt.Fprintln(os.Stderr, "petriview: -world is required")
		flag.Usage()
		os.Exit(2)
	}

	st, problems, err := openStore(*world, 768)
	for _, p := range problems {
		log.Printf("warning: %v", p)
	}
	if err != nil {
		log.Fatalf("petriview: %v", err)
	}

	var fsys *noitadata.FS
	if *install != "" {
		fsys, err = noitadata.Open(*install)
	} else {
		fsys, err = noitadata.OpenAuto()
	}
	if err != nil {
		log.Fatalf("petriview: opening Noita install: %v\npass -install if auto-detection failed", err)
	}
	defer fsys.Close()

	palette, err := noitasave.LoadPalette(fsys)
	if err != nil {
		log.Fatalf("petriview: %v", err)
	}

	start := time.Now()
	log.Printf("indexing %d chunks...", len(st.coords))
	idx, errs := buildIndex(st, palette)
	for _, e := range errs {
		log.Printf("warning: %v", e)
	}
	indexJSON, err := json.Marshal(idx)
	if err != nil {
		log.Fatalf("petriview: %v", err)
	}
	log.Printf("indexed in %.1fs: %d chunks, %d bodies, %d joints, %d materials, %d solid cells",
		time.Since(start).Seconds(), idx.Totals.Chunks, idx.Totals.Objects,
		idx.Totals.Joints, len(idx.Materials), idx.Totals.SolidCells)

	ts := newTileServer(st, palette, 1024)

	sub, err := fs.Sub(assets, "assets")
	if err != nil {
		log.Fatalf("petriview: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("GET /vendor/", http.FileServerFS(sub))
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, sub, "index.html")
	})
	mux.HandleFunc("GET /index.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(indexJSON)
	})
	mux.HandleFunc("GET /tile/{z}/{x}/{y}", ts.handleTile)
	mux.HandleFunc("GET /material", handleMaterial(st))
	mux.HandleFunc("GET /stats", func(w http.ResponseWriter, r *http.Request) {
		ts.mu.Lock()
		tileCount := ts.lru.Len()
		ts.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int64{
			"tilesQueued":    ts.queued.Load(),
			"tilesRendering": ts.rendering.Load(),
			"tilesRendered":  ts.rendered.Load(),
			"tileCacheHits":  ts.hits.Load(),
			"tilesCached":    int64(tileCount),
			"chunksCached":   int64(st.cached()),
			"chunkCacheCap":  int64(st.cap),
			"chunkLoads":     st.loads.Load(),
			"chunksTotal":    int64(len(st.coords)),
		})
	})

	log.Printf("serving %s on http://%s", *world, *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatalf("petriview: %v", err)
	}
}

func (t *tileServer) handleTile(w http.ResponseWriter, r *http.Request) {
	z, err1 := strconv.Atoi(r.PathValue("z"))
	tx, err2 := strconv.Atoi(r.PathValue("x"))
	ty, err3 := strconv.Atoi(strings.TrimSuffix(r.PathValue("y"), ".png"))
	if err1 != nil || err2 != nil || err3 != nil || z < 0 || z > nativeZoom {
		http.Error(w, "bad tile address", http.StatusBadRequest)
		return
	}
	q := r.URL.Query()
	b, err := t.tile(r.Context(), z, tx, ty, q.Get("hl"), q.Get("dim") == "1", q.Get("phys") != "0")
	if err != nil {
		if r.Context().Err() != nil {
			return // client gave up; nothing to answer
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-cache")
	w.Write(b)
}

func handleMaterial(st *store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		wx, err1 := strconv.Atoi(q.Get("x"))
		wy, err2 := strconv.Atoi(q.Get("y"))
		if err1 != nil || err2 != nil {
			http.Error(w, "bad coordinate", http.StatusBadRequest)
			return
		}
		var resp struct {
			Material string `json:"material,omitempty"`
			Custom   bool   `json:"custom,omitempty"`
			Chunk    string `json:"chunk,omitempty"`
		}
		coord := chunkOrigin(wx, wy)
		if e, err := st.chunk(coord); e != nil && err == nil {
			lx, ly := wx-coord.X, wy-coord.Y
			if lx < int(e.c.Width) && ly < int(e.c.Height) {
				i := ly*int(e.c.Width) + lx
				if cell := e.c.Cells[i]; cell != 0 {
					if mi, _ := noitasave.CellMaterial(cell); int(mi) < len(e.c.MaterialNames) {
						resp.Material = e.c.MaterialNames[mi]
					}
					resp.Custom = e.colorIdx[i] >= 0
					resp.Chunk = filepath.Base(st.paths[coord])
				}
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// buildIndex summarises every chunk without retaining any of them: chunks are
// loaded, tallied, and dropped, a bounded number at a time.
func buildIndex(st *store, p *noitasave.Palette) (jsonWorld, []error) {
	idx := jsonWorld{
		Bounds: jsonRect{
			X: st.bounds.Min.X, Y: st.bounds.Min.Y,
			W: st.bounds.Dx(), H: st.bounds.Dy(),
		},
		ChunkSize: noitasave.ChunkSize,
	}

	type result struct {
		jc     jsonChunk
		counts map[string]int
		err    error
	}
	results := make([]result, len(st.coords))
	sem := make(chan struct{}, runtime.NumCPU())
	var wg sync.WaitGroup
	for i, coord := range st.coords {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			path := st.paths[coord]
			c, err := noitasave.LoadChunk(path)
			if err != nil {
				results[i] = result{err: err}
				return
			}
			jc, counts := indexChunk(coord, path, c, p)
			results[i] = result{jc: jc, counts: counts}
		}()
	}
	wg.Wait()

	var errs []error
	worldMats := map[string]int{}
	for _, r := range results {
		if r.err != nil {
			errs = append(errs, r.err)
			continue
		}
		for name, n := range r.counts {
			worldMats[name] += n
		}
		idx.Totals.Chunks++
		idx.Totals.Objects += len(r.jc.Objects)
		idx.Totals.Joints += len(r.jc.Joints)
		idx.Totals.CustomColors += r.jc.Colors
		idx.Totals.SolidCells += r.jc.SolidCells
		idx.Chunks = append(idx.Chunks, r.jc)
	}
	idx.Materials = tally(worldMats, p)
	return idx, errs
}

func indexChunk(coord noitasave.ChunkCoord, path string, c *noitasave.Chunk, p *noitasave.Palette) (jsonChunk, map[string]int) {
	jc := jsonChunk{
		File:    filepath.Base(path),
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
	return jc, counts
}

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
