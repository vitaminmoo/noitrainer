package main

import (
	"context"
	"image/color"
	"log"
	"os"
	"strings"
	"sync/atomic"
	"time"

	overlay "github.com/vitaminmoo/go-overlay"
	"github.com/vitaminmoo/memtools/process"
	"noitrainer/noita"
)

// overlayEntity is a single entity to render on the overlay.
type overlayEntity struct {
	X, Y       float32
	Name       string
	Color      color.RGBA
	HasHitbox  bool
	AabbMinX   float32
	AabbMaxX   float32
	AabbMinY   float32
	AabbMaxY   float32
	HitOffsetX float32
	HitOffsetY float32
}

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
	entities  atomic.Pointer[[]overlayEntity]
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

	// Draw all tracked entities with labels.
	if ents := s.entities.Load(); ents != nil {
		lineH := float64(c.TextLineHeight())
		for _, e := range *ents {
			sx := wx + float64(e.X-d.CameraX)*scaleX + ww/2
			sy := wy + float64(e.Y-d.CameraY)*scaleY + wh/2
			// Skip if off-screen.
			if sx < wx-20 || sx > wx+ww+20 || sy < wy-20 || sy > wy+wh+20 {
				continue
			}

			if e.HasHitbox {
				// Draw AABB rectangle (hitbox-relative coords scaled to screen).
				ox := float64(e.HitOffsetX) * scaleX
				oy := float64(e.HitOffsetY) * scaleY
				x0 := sx + ox + float64(e.AabbMinX)*scaleX
				y0 := sy + oy + float64(e.AabbMinY)*scaleY
				x1 := sx + ox + float64(e.AabbMaxX)*scaleX
				y1 := sy + oy + float64(e.AabbMaxY)*scaleY
				c.Line(x0, y0, x1, y0, e.Color) // top
				c.Line(x1, y0, x1, y1, e.Color) // right
				c.Line(x1, y1, x0, y1, e.Color) // bottom
				c.Line(x0, y1, x0, y0, e.Color) // left
				// Label centered above the hitbox, clamped to window top
				// so it stays readable when the hitbox top is off-screen.
				if e.Name != "" {
					tw := float64(c.TextWidth(e.Name))
					cx := (x0 + x1) / 2
					labelY := y0 - lineH - 2
					if labelY < wy && y1 > wy {
						labelY = wy + 2
					}
					c.Text(e.Name, cx-tw/2, labelY, e.Color)
				}
			} else {
				c.Ellipse(sx, sy, 6, 6, e.Color)
				if e.Name != "" {
					c.Text(e.Name, sx+8, sy-lineH/2, e.Color)
				}
			}
		}
	}

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

	matchFunc := func(name string) bool {
		return strings.HasPrefix(name, "Noita - Build")
	}

	var matcher overlay.WindowMatcher
	if os.Getenv("SWAYSOCK") != "" {
		matcher = overlay.NewSwayMatcher(matchFunc)
	} else {
		matcher = overlay.NewKWinMatcher(matchFunc)
	}

	opts := overlay.Options{
		WindowMatcher: matcher,
		UpdateRate:    16 * time.Millisecond, // ~60Hz
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
