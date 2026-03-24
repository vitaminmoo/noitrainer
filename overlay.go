package main

import (
	"context"
	"image/color"
	"log"
	"strings"
	"sync/atomic"
	"time"

	overlay "github.com/vitaminmoo/go-overlay"
	"github.com/vitaminmoo/memtools/process"
	"noitrainer/noita"
)

// overlayData is the snapshot drawn each frame.
type overlayData struct {
	EntityX, EntityY float32 // selected entity world position
	CameraX, CameraY float32
	ViewW, ViewH     float32
	HasSelection     bool
}

// overlaySelection is set by the TUI to tell the overlay which entity to track.
type overlaySelection struct {
	EntityPtr uint32 // memory pointer of selected entity (0 = none)
}

type overlayScene struct {
	data      atomic.Pointer[overlayData]
	selection atomic.Pointer[overlaySelection]
	proc      atomic.Pointer[process.Process]
	drawCount int
}

func (s *overlayScene) Update(ctx context.Context) {
	proc := s.proc.Load()
	if proc == nil {
		return
	}

	reader := noita.NewReader(proc)

	// Read only camera (lightweight — no entity list scan).
	cam := reader.ReadCamera()
	if cam == nil {
		return
	}

	d := &overlayData{
		CameraX: cam.CameraX,
		CameraY: cam.CameraY,
		ViewW:   cam.ViewW,
		ViewH:   cam.ViewH,
	}

	// Read only the selected entity's position.
	if sel := s.selection.Load(); sel != nil && sel.EntityPtr != 0 {
		if ent, _ := noita.ReadEntity(reader.Ctx, uintptr(sel.EntityPtr)); ent != nil {
			d.HasSelection = true
			d.EntityX = ent.PosX
			d.EntityY = ent.PosY
		}
	}

	s.data.Store(d)
}

func (s *overlayScene) Draw(c *overlay.Canvas) {
	s.drawCount++
	if s.drawCount%60 == 1 {
		log.Printf("overlay draw #%d: screen=%dx%d", s.drawCount, c.Width(), c.Height())
	}

	wr, ok := c.WindowRect()
	if !ok {
		if s.drawCount%60 == 1 {
			log.Printf("overlay draw: no window rect")
		}
		return
	}

	d := s.data.Load()
	if d == nil || d.ViewW == 0 || d.ViewH == 0 {
		if s.drawCount%60 == 1 {
			log.Printf("overlay draw: no data (d=%v)", d != nil)
		}
		return
	}

	wx := float64(wr.X)
	wy := float64(wr.Y)
	ww := float64(wr.W)
	wh := float64(wr.H)

	if s.drawCount%60 == 1 {
		log.Printf("overlay draw: wr={%.0f,%.0f %.0fx%.0f} cam=%.0f,%.0f view=%.0fx%.0f",
			wx, wy, ww, wh, d.CameraX, d.CameraY, d.ViewW, d.ViewH)
	}

	// Debug: draw red rectangle around the detected window bounds.
	red := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	c.Line(wx, wy, wx+ww, wy, red)        // top
	c.Line(wx+ww, wy, wx+ww, wy+wh, red)  // right
	c.Line(wx+ww, wy+wh, wx, wy+wh, red)  // bottom
	c.Line(wx, wy+wh, wx, wy, red)         // left

	// Map game world coordinates to screen pixels within the window.
	scaleX := ww / float64(d.ViewW)
	scaleY := wh / float64(d.ViewH)

	// Draw circle around selected entity.
	if d.HasSelection {
		sx := wx + float64(d.EntityX-d.CameraX)*scaleX + ww/2
		sy := wy + float64(d.EntityY-d.CameraY)*scaleY + wh/2

		c.Ellipse(sx, sy, 12, 12, color.RGBA{R: 255, G: 100, B: 100, A: 220})
	}

	// Draw green marker at world origin (0,0).
	ox := wx + float64(0-d.CameraX)*scaleX + ww/2
	oy := wy + float64(0-d.CameraY)*scaleY + wh/2
	c.Ellipse(ox, oy, 8, 8, color.RGBA{R: 0, G: 255, B: 0, A: 255})
}

func startOverlay(ctx context.Context) *overlayScene {
	scene := &overlayScene{}

	opts := overlay.Options{
		WindowMatcher: overlay.NewKWinMatcher(func(name string) bool {
			return strings.HasPrefix(name, "Noita - Build")
		}),
		UpdateRate: 16 * time.Millisecond, // ~60Hz
	}

	go func() {
		log.Printf("overlay: starting")
		err := overlay.Run(ctx, opts, scene)
		if err != nil {
			log.Printf("overlay: exited with error: %v", err)
		} else {
			log.Printf("overlay: exited cleanly")
		}
	}()

	return scene
}
