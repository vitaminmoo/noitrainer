package noitadata

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"regexp"
	"sort"
	"strings"
)

// Entity is a parsed Noita entity XML, optionally with its <Base> chain
// resolved (merged). A resolved Entity's Components reflect the full
// inheritance chain with overrides applied on top of each ancestor.
type Entity struct {
	// SourcePath is the fs path the entity was parsed from.
	SourcePath string
	// Name is the "name" attribute of the root <Entity> element (may be a
	// $-prefixed translation key).
	Name string
	// Tags is the "tags" attribute split on comma.
	Tags []string
	// Components are the components of the entity. When Resolved is true
	// this includes the full inherited set with overrides.
	Components []Component
	// BaseChain lists the <Base file="..."/> paths walked to resolve this
	// entity, root-first. Empty if the entity has no <Base>.
	BaseChain []string
	// Resolved reports whether <Base> inheritance has been merged.
	Resolved bool

	// Intermediate state used only during parse/resolve, not part of the
	// public API. Populated by parseEntityBytes, consumed by resolveEntity.
	rawBaseFiles     []baseBlock
	directComponents []Component
}

// Component is a single component on an entity.
type Component struct {
	// Type is the XML element name (e.g. "DamageModelComponent").
	Type string
	// Tags is the "_tags" attribute split on comma (empty if absent).
	Tags []string
	// Attrs holds all attributes on the component element excluding _tags.
	Attrs map[string]string
	// Children are nested elements directly inside this component (e.g.
	// <damage_multipliers ...> inside <DamageModelComponent>).
	Children []Component
}

// ParseEntity reads an Entity XML from fsys and returns the raw parsed
// form (without resolving <Base>). Use ResolveEntity to get the merged
// inheritance.
func ParseEntity(fsys fs.FS, p string) (*Entity, error) {
	data, err := fs.ReadFile(fsys, p)
	if err != nil {
		return nil, err
	}
	return parseEntityBytes(p, data)
}

// ResolveEntity parses the entity at p and merges its <Base> chain.
// Each Base's components are loaded first, then the overrides nested
// inside the <Base> element are merged on top (by Type, first-match),
// then the entity's direct top-level components are appended.
func ResolveEntity(fsys fs.FS, p string) (*Entity, error) {
	return resolveEntity(fsys, p, map[string]bool{})
}

func resolveEntity(fsys fs.FS, p string, seen map[string]bool) (*Entity, error) {
	if seen[p] {
		return nil, fmt.Errorf("cycle in base chain at %s", p)
	}
	seen[p] = true

	ent, err := ParseEntity(fsys, p)
	if err != nil {
		return nil, err
	}

	// If the entity declared any <Base>, resolve them in order and merge.
	merged := []Component{}
	for _, bf := range ent.rawBaseFiles {
		parent, err := resolveEntity(fsys, bf.path, seen)
		if err != nil {
			// Missing bases are common (placeholders, conditional files).
			// Keep going with what we have and record the chain.
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return nil, err
		}
		merged = append(merged, parent.Components...)
		ent.BaseChain = append(ent.BaseChain, parent.BaseChain...)
		ent.BaseChain = append(ent.BaseChain, bf.path)
		for _, override := range bf.overrides {
			merged = applyOverride(merged, override)
		}
	}

	merged = append(merged, ent.directComponents...)
	ent.Components = merged
	ent.Resolved = true
	ent.rawBaseFiles = nil
	ent.directComponents = nil
	return ent, nil
}

// applyOverride merges o into the first component of matching Type in
// comps (attributes and children), or appends it if no match.
func applyOverride(comps []Component, o Component) []Component {
	for i := range comps {
		if comps[i].Type == o.Type && tagsOverlap(comps[i].Tags, o.Tags) {
			for k, v := range o.Attrs {
				if comps[i].Attrs == nil {
					comps[i].Attrs = map[string]string{}
				}
				comps[i].Attrs[k] = v
			}
			for _, c := range o.Children {
				comps[i].Children = applyOverride(comps[i].Children, c)
			}
			return comps
		}
	}
	return append(comps, o)
}

// tagsOverlap is true if a and b share any tag, or either is empty (so
// an untagged override matches any component of the same type, which is
// Noita's common case).
func tagsOverlap(a, b []string) bool {
	if len(a) == 0 || len(b) == 0 {
		return true
	}
	set := make(map[string]bool, len(a))
	for _, t := range a {
		set[t] = true
	}
	for _, t := range b {
		if set[t] {
			return true
		}
	}
	return false
}

// --- XML parsing ---

type baseBlock struct {
	path      string
	overrides []Component
}

// parseEntityBytes parses a Noita Entity XML document.
func parseEntityBytes(p string, data []byte) (*Entity, error) {
	data = stripXMLComments(data)
	dec := xml.NewDecoder(strings.NewReader(string(data)))
	dec.Strict = false

	ent := &Entity{SourcePath: p}
	depth := 0
	for {
		tok, err := dec.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			// Bail on any parse error — don't loop on syntax errors.
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			if depth == 1 {
				// Root must be <Entity>
				if t.Name.Local != "Entity" {
					return nil, fmt.Errorf("%s: root element is %q, want Entity", p, t.Name.Local)
				}
				for _, a := range t.Attr {
					switch a.Name.Local {
					case "name":
						ent.Name = a.Value
					case "tags":
						ent.Tags = splitCSV(a.Value)
					}
				}
				continue
			}
			if depth == 2 {
				if t.Name.Local == "Base" {
					bb, err := parseBaseBlock(dec, t)
					if err != nil {
						return nil, err
					}
					ent.rawBaseFiles = append(ent.rawBaseFiles, bb)
					depth--
					continue
				}
				comp, err := parseComponent(dec, t)
				if err != nil {
					return nil, err
				}
				ent.directComponents = append(ent.directComponents, comp)
				depth--
				continue
			}
			// depth > 2 here means we're skipping; shouldn't normally hit.
		case xml.EndElement:
			depth--
			if depth < 0 {
				break
			}
		}
	}
	return ent, nil
}

func parseBaseBlock(dec *xml.Decoder, start xml.StartElement) (baseBlock, error) {
	bb := baseBlock{}
	for _, a := range start.Attr {
		if a.Name.Local == "file" {
			bb.path = a.Value
		}
	}
	for {
		tok, err := dec.Token()
		if err != nil {
			return bb, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			c, err := parseComponent(dec, t)
			if err != nil {
				return bb, err
			}
			bb.overrides = append(bb.overrides, c)
		case xml.EndElement:
			if t.Name.Local == start.Name.Local {
				return bb, nil
			}
		}
	}
}

func parseComponent(dec *xml.Decoder, start xml.StartElement) (Component, error) {
	c := Component{Type: start.Name.Local, Attrs: map[string]string{}}
	for _, a := range start.Attr {
		if a.Name.Local == "_tags" {
			c.Tags = splitCSV(a.Value)
			continue
		}
		c.Attrs[a.Name.Local] = a.Value
	}
	for {
		tok, err := dec.Token()
		if err != nil {
			return c, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			child, err := parseComponent(dec, t)
			if err != nil {
				return c, err
			}
			c.Children = append(c.Children, child)
		case xml.EndElement:
			if t.Name.Local == start.Name.Local {
				return c, nil
			}
		}
	}
}

func splitCSV(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := parts[:0]
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}


// --- File reference extraction ---

// pathExts recognized as Noita asset references when they appear as an
// attribute value.
var pathExts = map[string]bool{
	".xml": true, ".png": true, ".lua": true, ".bank": true,
	".frag": true, ".vert": true, ".bmp": true, ".csv": true,
	".txt": true, ".plz": true, ".psd": true,
}

// refPattern captures the leading path-looking part of an attribute value
// before any dollar-interpolation placeholder.
var refPattern = regexp.MustCompile(`(?s)^([^\s<>]+)`)

// Refs returns the outbound file references present in this component
// and its children. Values are normalized paths (forward slashes) with
// Noita's $-placeholders preserved.
func (c Component) Refs() []string {
	seen := map[string]bool{}
	var out []string
	collectRefs(c, seen, &out)
	sort.Strings(out)
	return out
}

// Refs returns the outbound file references for the entire entity
// (across all components, and including the Base chain if resolved).
func (e *Entity) Refs() []string {
	seen := map[string]bool{}
	var out []string
	for _, bp := range e.BaseChain {
		if !seen[bp] {
			seen[bp] = true
			out = append(out, bp)
		}
	}
	for _, c := range e.Components {
		collectRefs(c, seen, &out)
	}
	sort.Strings(out)
	return out
}

func collectRefs(c Component, seen map[string]bool, out *[]string) {
	for _, v := range c.Attrs {
		for _, r := range extractRefs(v) {
			if !seen[r] {
				seen[r] = true
				*out = append(*out, r)
			}
		}
	}
	for _, ch := range c.Children {
		collectRefs(ch, seen, out)
	}
}

// extractRefs pulls path-looking references out of a single attribute
// value. An attribute can contain a single reference ("data/foo.xml") or
// a space/comma separated list; Noita occasionally stuffs multiple
// references into one attribute, so handle that too.
func extractRefs(v string) []string {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	var out []string
	for _, tok := range strings.FieldsFunc(v, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\n' || r == ','
	}) {
		m := refPattern.FindStringSubmatch(tok)
		if m == nil {
			continue
		}
		s := m[1]
		if looksLikePath(s) {
			out = append(out, path.Clean(s))
		}
	}
	return out
}

func looksLikePath(s string) bool {
	if strings.HasPrefix(s, "data/") {
		return true
	}
	ext := strings.ToLower(path.Ext(s))
	if pathExts[ext] {
		// require at least one slash to avoid matching bare filenames
		// that aren't fs references (e.g. "zombie.xml" in a comment).
		return strings.Contains(s, "/")
	}
	return false
}
