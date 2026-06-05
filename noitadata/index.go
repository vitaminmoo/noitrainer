package noitadata

import (
	"encoding/xml"
	"errors"
	"io/fs"
	"sort"
	"strings"
	"sync"
)

// Index is a cross-reference index over every *.xml under data/.
// Built once; concurrent reads are safe.
type Index struct {
	// Outbound: file path -> sorted list of file paths referenced
	// from its attributes (including <Base file="...">).
	Outbound map[string][]string
	// Inbound: file path -> sorted list of files that reference it.
	Inbound map[string][]string
	// Components: component type name (e.g. "DamageModelComponent") ->
	// sorted list of entity files where that component appears directly
	// or inside a <Base> block. Does NOT include components inherited
	// transitively from deeper ancestors (run ResolveEntity for that).
	Components map[string][]string
	// Entities: sorted list of all files whose root element is <Entity>.
	Entities []string
}

var (
	indexOnce sync.Once
	indexErr  error
	builtIdx  *Index
)

// Index returns the xref index for this FS, building it on first call.
// Subsequent calls return the cached index.
func (n *FS) Index() (*Index, error) {
	indexOnce.Do(func() {
		builtIdx, indexErr = buildIndex(n)
	})
	return builtIdx, indexErr
}

func buildIndex(n *FS) (*Index, error) {
	outbound := map[string]map[string]bool{}
	components := map[string]map[string]bool{}
	entities := map[string]bool{}

	err := fs.WalkDir(n, "data", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(p), ".xml") {
			return nil
		}
		data, err := n.ReadFile(p)
		if err != nil {
			return nil
		}
		info := scanXML(stripXMLComments(data))
		if info.isEntity {
			entities[p] = true
			for _, comp := range info.componentTypes {
				if components[comp] == nil {
					components[comp] = map[string]bool{}
				}
				components[comp][p] = true
			}
		}
		if len(info.refs) > 0 {
			set := outbound[p]
			if set == nil {
				set = map[string]bool{}
				outbound[p] = set
			}
			for _, r := range info.refs {
				set[r] = true
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	idx := &Index{
		Outbound:   make(map[string][]string, len(outbound)),
		Inbound:    map[string][]string{},
		Components: make(map[string][]string, len(components)),
	}
	inboundSet := map[string]map[string]bool{}
	for src, refs := range outbound {
		out := make([]string, 0, len(refs))
		for r := range refs {
			out = append(out, r)
			if inboundSet[r] == nil {
				inboundSet[r] = map[string]bool{}
			}
			inboundSet[r][src] = true
		}
		sort.Strings(out)
		idx.Outbound[src] = out
	}
	for r, srcs := range inboundSet {
		out := make([]string, 0, len(srcs))
		for s := range srcs {
			out = append(out, s)
		}
		sort.Strings(out)
		idx.Inbound[r] = out
	}
	for comp, paths := range components {
		out := make([]string, 0, len(paths))
		for p := range paths {
			out = append(out, p)
		}
		sort.Strings(out)
		idx.Components[comp] = out
	}
	for p := range entities {
		idx.Entities = append(idx.Entities, p)
	}
	sort.Strings(idx.Entities)
	return idx, nil
}

// xmlInfo is the raw result of scanXML: refs is every path-looking
// attribute value; componentTypes is every element name seen at Entity
// depth==2 (direct components) plus those inside <Base> blocks. isEntity
// reports whether the root element is <Entity>.
type xmlInfo struct {
	isEntity       bool
	componentTypes []string
	refs           []string
}

func scanXML(data []byte) xmlInfo {
	dec := xml.NewDecoder(strings.NewReader(string(data)))
	dec.Strict = false

	var info xmlInfo
	compSet := map[string]bool{}
	refSet := map[string]bool{}

	depth := 0
	// elemStack tracks element names for the current path so we can
	// detect "direct child of Entity" (depth 2) and "direct child of
	// <Base> which itself is a direct child of Entity" (depth 3 where
	// parent is Base, grandparent is Entity).
	elemStack := make([]string, 0, 8)
	rootSeen := false

	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			elemStack = append(elemStack, t.Name.Local)
			if !rootSeen {
				rootSeen = true
				if t.Name.Local == "Entity" {
					info.isEntity = true
				}
			}
			// Collect refs from any attribute, regardless of root type.
			for _, a := range t.Attr {
				for _, r := range extractRefs(a.Value) {
					refSet[r] = true
				}
			}
			if info.isEntity {
				// Component types: elements directly under <Entity> that
				// aren't <Base>, and elements directly under a <Base>
				// that's itself directly under <Entity>.
				if depth == 2 && t.Name.Local != "Base" {
					compSet[t.Name.Local] = true
				} else if depth == 3 && elemStack[1] == "Base" {
					compSet[t.Name.Local] = true
				}
			}
		case xml.EndElement:
			depth--
			if len(elemStack) > 0 {
				elemStack = elemStack[:len(elemStack)-1]
			}
		}
	}

	info.componentTypes = make([]string, 0, len(compSet))
	for c := range compSet {
		info.componentTypes = append(info.componentTypes, c)
	}
	sort.Strings(info.componentTypes)
	info.refs = make([]string, 0, len(refSet))
	for r := range refSet {
		info.refs = append(info.refs, r)
	}
	sort.Strings(info.refs)
	return info
}

// FindEntitiesWith returns entity paths whose parsed component set
// includes componentType and, for each (attr, value) in attrMatch, at
// least one component of componentType has attrs[attr] equal to (or,
// if value is empty, attr is present on) the component. Value matching
// is a substring test on the component's attribute string.
//
// Uses the index to narrow candidates, then re-parses matched files.
func (n *FS) FindEntitiesWith(componentType string, attrMatch map[string]string) ([]string, error) {
	idx, err := n.Index()
	if err != nil {
		return nil, err
	}
	candidates := idx.Components[componentType]
	if len(attrMatch) == 0 {
		out := make([]string, len(candidates))
		copy(out, candidates)
		return out, nil
	}
	var hits []string
	for _, p := range candidates {
		e, err := ParseEntity(n, p)
		if err != nil {
			continue
		}
		comps := append([]Component(nil), e.directComponents...)
		for _, bb := range e.rawBaseFiles {
			comps = append(comps, bb.overrides...)
		}
		if componentMatches(comps, componentType, attrMatch) {
			hits = append(hits, p)
		}
	}
	return hits, nil
}

func componentMatches(comps []Component, typ string, attrMatch map[string]string) bool {
	for _, c := range comps {
		if c.Type != typ {
			continue
		}
		ok := true
		for k, want := range attrMatch {
			got, present := c.Attrs[k]
			if !present {
				ok = false
				break
			}
			if want != "" && !strings.Contains(got, want) {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}

// stripXMLComments returns a copy of data with every <!-- ... --> span
// replaced with whitespace. Noita's XML contains hand-authored comments
// like `<!------------ MATERIALS -------------------->` whose `--`
// sequences inside the comment body are invalid per the XML spec and
// cause encoding/xml to spin with its non-strict decoder. We preempt
// that by scanning byte-wise for comment markers (which are
// straightforward to locate) before feeding to the decoder.
func stripXMLComments(data []byte) []byte {
	out := make([]byte, len(data))
	copy(out, data)
	for i := 0; i+4 <= len(out); {
		if out[i] == '<' && out[i+1] == '!' && out[i+2] == '-' && out[i+3] == '-' {
			end := i + 4
			for end+3 <= len(out) {
				if out[end] == '-' && out[end+1] == '-' && out[end+2] == '>' {
					end += 3
					break
				}
				end++
			}
			if end > len(out) {
				end = len(out)
			}
			for j := i; j < end; j++ {
				if out[j] != '\n' {
					out[j] = ' '
				}
			}
			i = end
			continue
		}
		i++
	}
	return out
}
