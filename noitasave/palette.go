package noitasave

// Material appearance, resolved from the game's data/materials.xml.
//
// Materials are declared as <CellData> plus <CellDataChild>, where a child
// inherits anything it does not override from the _parent it names. Two colour
// sources matter for rendering:
//
//	<Graphics color="aarrggbb" texture_file="data/materials_gfx/x.png">
//	wang_color="aarrggbb"   (fallback; used by worldgen)
//
// Noita paints solid materials by sampling texture_file in WORLD space, so the
// texture tiles seamlessly across chunk boundaries. Sampling the same way gets
// a render close to what the game shows; materials without a texture fall back
// to the flat Graphics colour and then to wang_color.

import (
	"encoding/xml"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"io/fs"
	"strconv"
	"strings"
)

// Material is one entry from materials.xml with inheritance already resolved.
type Material struct {
	Name        string
	Parent      string
	Color       color.RGBA // Graphics color, else wang_color
	HasColor    bool
	TextureFile string
}

// Palette resolves material names to pixels.
type Palette struct {
	fsys      fs.FS
	materials map[string]*Material
	textures  map[string]*image.RGBA // texture_file -> decoded, nil if unusable
}

type xmlGraphics struct {
	Color       string `xml:"color,attr"`
	TextureFile string `xml:"texture_file,attr"`
}

type xmlCellData struct {
	Name      string       `xml:"name,attr"`
	Parent    string       `xml:"_parent,attr"`
	WangColor string       `xml:"wang_color,attr"`
	Graphics  *xmlGraphics `xml:"Graphics"`
}

type xmlMaterials struct {
	CellData      []xmlCellData `xml:"CellData"`
	CellDataChild []xmlCellData `xml:"CellDataChild"`
}

// LoadPalette parses data/materials.xml from a Noita install FS.
func LoadPalette(fsys fs.FS) (*Palette, error) {
	raw, err := fs.ReadFile(fsys, "data/materials.xml")
	if err != nil {
		return nil, fmt.Errorf("noitasave: reading materials.xml: %w", err)
	}

	var doc xmlMaterials
	if err := xml.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("noitasave: parsing materials.xml: %w", err)
	}

	p := &Palette{
		fsys:      fsys,
		materials: make(map[string]*Material),
		textures:  make(map[string]*image.RGBA),
	}
	for _, group := range [][]xmlCellData{doc.CellData, doc.CellDataChild} {
		for _, cd := range group {
			if cd.Name == "" {
				continue
			}
			m := &Material{Name: cd.Name, Parent: cd.Parent}
			if cd.Graphics != nil {
				m.TextureFile = cd.Graphics.TextureFile
				if c, ok := parseNoitaColor(cd.Graphics.Color); ok {
					m.Color, m.HasColor = c, true
				}
			}
			if !m.HasColor {
				if c, ok := parseNoitaColor(cd.WangColor); ok {
					m.Color, m.HasColor = c, true
				}
			}
			p.materials[cd.Name] = m
		}
	}

	// Resolve inheritance: pull colour and texture down from ancestors.
	for name := range p.materials {
		p.resolve(name, 0)
	}
	return p, nil
}

// resolve fills in a material's missing appearance from its parent chain.
// depth guards against a cycle in _parent.
func (p *Palette) resolve(name string, depth int) *Material {
	m := p.materials[name]
	if m == nil || depth > 16 {
		return m
	}
	if m.HasColor && m.TextureFile != "" {
		return m
	}
	if m.Parent == "" || m.Parent == name {
		return m
	}
	parent := p.resolve(m.Parent, depth+1)
	if parent == nil {
		return m
	}
	if !m.HasColor && parent.HasColor {
		m.Color, m.HasColor = parent.Color, true
	}
	if m.TextureFile == "" {
		m.TextureFile = parent.TextureFile
	}
	return m
}

// Material returns the resolved material, if known.
func (p *Palette) Material(name string) (*Material, bool) {
	m, ok := p.materials[name]
	return m, ok
}

// Len reports how many materials were loaded.
func (p *Palette) Len() int { return len(p.materials) }

// texture decodes and caches a material texture. A nil result means the
// texture is missing or undecodable, and the caller should use a flat colour.
func (p *Palette) texture(path string) *image.RGBA {
	if path == "" {
		return nil
	}
	if img, seen := p.textures[path]; seen {
		return img
	}
	p.textures[path] = nil // negative-cache before attempting

	f, err := p.fsys.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	src, _, err := image.Decode(f)
	if err != nil {
		return nil
	}
	b := src.Bounds()
	if b.Empty() {
		return nil
	}
	rgba := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			rgba.Set(x, y, src.At(b.Min.X+x, b.Min.Y+y))
		}
	}
	p.textures[path] = rgba
	return rgba
}

// ColorAt returns the colour to paint for a material at world pixel (wx, wy).
// Textures are sampled in world space so they tile across chunk seams.
func (p *Palette) ColorAt(name string, wx, wy int) (color.RGBA, bool) {
	m, ok := p.materials[name]
	if !ok {
		return color.RGBA{}, false
	}
	if tex := p.texture(m.TextureFile); tex != nil {
		w, h := tex.Rect.Dx(), tex.Rect.Dy()
		c := tex.RGBAAt(mod(wx, w), mod(wy, h))
		if c.A > 0 {
			return c, true
		}
	}
	if m.HasColor {
		return m.Color, true
	}
	return color.RGBA{}, false
}

func mod(a, n int) int {
	if n <= 0 {
		return 0
	}
	m := a % n
	if m < 0 {
		m += n
	}
	return m
}

// parseNoitaColor parses Noita's "aarrggbb" (or "rrggbb") hex attribute.
func parseNoitaColor(s string) (color.RGBA, bool) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "#")
	if len(s) != 8 && len(s) != 6 {
		return color.RGBA{}, false
	}
	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return color.RGBA{}, false
	}
	if len(s) == 6 {
		return color.RGBA{R: uint8(v >> 16), G: uint8(v >> 8), B: uint8(v), A: 255}, true
	}
	return color.RGBA{
		A: uint8(v >> 24),
		R: uint8(v >> 16),
		G: uint8(v >> 8),
		B: uint8(v),
	}, true
}

// ColorFromCellCustom converts a stored colour word — a custom cell colour or
// a physics-body pixel — to an RGBA.
//
// These are packed 0xAABBGGRR, i.e. RGBA bytes as they sit in the game's
// little-endian memory, which is the OPPOSITE channel order from the
// 0xAARRGGBB attributes in materials.xml. Reading them as ARGB renders wood as
// blue and snow with gold fringes, which is how this was caught.
func ColorFromCellCustom(v uint32) color.RGBA {
	return color.RGBA{
		A: uint8(v >> 24),
		B: uint8(v >> 16),
		G: uint8(v >> 8),
		R: uint8(v),
	}
}
