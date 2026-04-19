// noita-mcp exposes Noita's game files (disk tree + data/data.wak) as MCP
// tools so an agent can list directories, read files, glob, and extract
// subtrees to real paths.
package main

import (
	"context"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"noitrainer/noitadata"
)

func main() {
	var (
		installFlag = flag.String("install", "", "override Noita install directory (else auto-detected or $NOITA_PATH)")
	)
	flag.Parse()

	var n *noitadata.FS
	var err error
	if *installFlag != "" {
		n, err = noitadata.Open(*installFlag)
	} else {
		n, err = noitadata.OpenAuto()
	}
	if err != nil {
		log.Printf("noita-mcp: install unavailable (%v); static tools will return errors", err)
		n = nil
	} else {
		defer n.Close()
		log.Printf("noita-mcp: serving %s (wak: %d files)", n.Root(), n.Wak().Len())
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "noita-data",
		Version: "0.1.0",
	}, nil)

	registerTools(server, n)
	registerSemanticTools(server, n)
	registerRuntimeTools(server)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("mcp server: %v", err)
	}
}

// --- tool inputs/outputs ---

type listDirInput struct {
	Path string `json:"path" jsonschema:"directory path relative to the Noita install root, forward slashes, empty or '.' for the root"`
}

type statInput struct {
	Path string `json:"path" jsonschema:"file or directory path relative to the Noita install root"`
}

type readFileInput struct {
	Path     string `json:"path" jsonschema:"file path relative to the Noita install root"`
	Encoding string `json:"encoding,omitempty" jsonschema:"'text' (default; errors on non-UTF8) or 'base64' for binary files"`
	Offset   int64  `json:"offset,omitempty" jsonschema:"byte offset into the file (default 0)"`
	Length   int64  `json:"length,omitempty" jsonschema:"max bytes to return (default: whole file from offset)"`
}

type globInput struct {
	Pattern string `json:"pattern" jsonschema:"fs.Glob-style pattern (e.g. 'data/entities/**/*.xml'). ** is supported."`
	Limit   int    `json:"limit,omitempty" jsonschema:"max matches to return (default 500)"`
}

type extractDirInput struct {
	SourceDir string `json:"source_dir" jsonschema:"directory path in the noita fs (e.g. 'data/entities/animals')"`
	DestDir   string `json:"dest_dir" jsonschema:"absolute on-disk directory to write into; will be created"`
}

type searchInput struct {
	Pattern    string `json:"pattern" jsonschema:"substring to search for (case-sensitive)"`
	PathGlob   string `json:"path_glob,omitempty" jsonschema:"restrict search to files matching this glob (default 'data/**/*.xml')"`
	MaxMatches int    `json:"max_matches,omitempty" jsonschema:"stop after this many matches (default 200)"`
}

// --- registration ---

func registerTools(s *mcp.Server, n *noitadata.FS) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_dir",
		Description: "List the direct contents of a directory. Returns names and whether each entry is a directory.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in listDirInput) (*mcp.CallToolResult, any, error) {
		if err := requireFS(n); err != nil {
			return toolErr(err)
		}
		p := normalizePath(in.Path)
		entries, err := fs.ReadDir(n, p)
		if err != nil {
			return toolErr(err)
		}
		var b strings.Builder
		fmt.Fprintf(&b, "%s  (%d entries)\n", p, len(entries))
		for _, e := range entries {
			kind := "file"
			if e.IsDir() {
				kind = "dir "
			}
			info, err := e.Info()
			size := int64(-1)
			if err == nil {
				size = info.Size()
			}
			if e.IsDir() {
				fmt.Fprintf(&b, "  [%s] %s\n", kind, e.Name())
			} else {
				fmt.Fprintf(&b, "  [%s] %s (%d bytes)\n", kind, e.Name(), size)
			}
		}
		return textResult(b.String()), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "stat",
		Description: "Return metadata (size, is_dir) about a path.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in statInput) (*mcp.CallToolResult, any, error) {
		if err := requireFS(n); err != nil {
			return toolErr(err)
		}
		p := normalizePath(in.Path)
		info, err := fs.Stat(n, p)
		if err != nil {
			return toolErr(err)
		}
		return textResult(fmt.Sprintf("%s\n  is_dir: %v\n  size: %d\n", p, info.IsDir(), info.Size())), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "read_file",
		Description: "Read a file's contents. Default returns UTF-8 text; use encoding='base64' for binary files.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in readFileInput) (*mcp.CallToolResult, any, error) {
		if err := requireFS(n); err != nil {
			return toolErr(err)
		}
		p := normalizePath(in.Path)
		data, err := n.ReadFile(p)
		if err != nil {
			return toolErr(err)
		}
		if in.Offset < 0 || in.Offset > int64(len(data)) {
			return toolErr(fmt.Errorf("offset %d out of range (file is %d bytes)", in.Offset, len(data)))
		}
		data = data[in.Offset:]
		if in.Length > 0 && in.Length < int64(len(data)) {
			data = data[:in.Length]
		}
		enc := in.Encoding
		if enc == "" {
			enc = "text"
		}
		switch enc {
		case "text":
			if !utf8.Valid(data) {
				return toolErr(fmt.Errorf("%s is not valid UTF-8; retry with encoding='base64'", p))
			}
			return textResult(string(data)), nil, nil
		case "base64":
			return textResult(base64.StdEncoding.EncodeToString(data)), nil, nil
		default:
			return toolErr(fmt.Errorf("unknown encoding %q (use 'text' or 'base64')", enc))
		}
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "glob",
		Description: "Find files matching a glob. Supports ** for recursive wildcards. Example: 'data/entities/**/*.xml'.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in globInput) (*mcp.CallToolResult, any, error) {
		if err := requireFS(n); err != nil {
			return toolErr(err)
		}
		limit := in.Limit
		if limit <= 0 {
			limit = 500
		}
		matches, err := doubleGlob(n, in.Pattern, limit)
		if err != nil {
			return toolErr(err)
		}
		sort.Strings(matches)
		var b strings.Builder
		fmt.Fprintf(&b, "%d match(es) for %q\n", len(matches), in.Pattern)
		for _, m := range matches {
			fmt.Fprintln(&b, m)
		}
		return textResult(b.String()), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "extract_dir",
		Description: "Copy a subtree of the noita fs to a real on-disk directory. Useful when external tools need native file paths.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in extractDirInput) (*mcp.CallToolResult, any, error) {
		if err := requireFS(n); err != nil {
			return toolErr(err)
		}
		if in.DestDir == "" || !filepath.IsAbs(in.DestDir) {
			return toolErr(fmt.Errorf("dest_dir must be an absolute path"))
		}
		count, err := noitadata.ExtractDir(n, in.SourceDir, in.DestDir)
		if err != nil {
			return toolErr(err)
		}
		return textResult(fmt.Sprintf("wrote %d file(s) from %s to %s\n", count, in.SourceDir, in.DestDir)), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "search",
		Description: "Substring search across files matching path_glob (default data/**/*.xml). Returns file:line matches.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in searchInput) (*mcp.CallToolResult, any, error) {
		if err := requireFS(n); err != nil {
			return toolErr(err)
		}
		if in.Pattern == "" {
			return toolErr(fmt.Errorf("pattern is required"))
		}
		pathGlob := in.PathGlob
		if pathGlob == "" {
			pathGlob = "data/**/*.xml"
		}
		max := in.MaxMatches
		if max <= 0 {
			max = 200
		}
		files, err := doubleGlob(n, pathGlob, 100000)
		if err != nil {
			return toolErr(err)
		}
		var b strings.Builder
		hits := 0
		needle := []byte(in.Pattern)
	outer:
		for _, f := range files {
			data, err := n.ReadFile(f)
			if err != nil {
				continue
			}
			line := 1
			for start := 0; start < len(data); {
				nl := indexByteFrom(data, '\n', start)
				end := nl
				if end < 0 {
					end = len(data)
				}
				if containsBytes(data[start:end], needle) {
					fmt.Fprintf(&b, "%s:%d: %s\n", f, line, strings.TrimSpace(string(data[start:end])))
					hits++
					if hits >= max {
						break outer
					}
				}
				line++
				if nl < 0 {
					break
				}
				start = nl + 1
			}
		}
		if hits == 0 {
			return textResult(fmt.Sprintf("no matches for %q in %s\n", in.Pattern, pathGlob)), nil, nil
		}
		return textResult(b.String()), nil, nil
	})
}

// --- Noita-specific semantic tools ---

type describeEntityInput struct {
	Path string `json:"path" jsonschema:"entity XML path, e.g. 'data/entities/animals/zombie.xml'"`
	Raw  bool   `json:"raw,omitempty" jsonschema:"when true, skip <Base> resolution and return only the entity's own content"`
}

type xrefsInput struct {
	Path string `json:"path" jsonschema:"file path to examine references for"`
}

type findEntitiesInput struct {
	Component string            `json:"component" jsonschema:"component type name, e.g. 'DamageModelComponent'"`
	Where     map[string]string `json:"where,omitempty" jsonschema:"optional attribute filters; value is substring-matched. Use empty string to match presence only."`
	Limit     int               `json:"limit,omitempty" jsonschema:"max results (default 200)"`
}

func registerSemanticTools(s *mcp.Server, n *noitadata.FS) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "describe_entity",
		Description: "Parse a Noita Entity XML with optional <Base> inheritance resolution. Returns name, tags, inheritance chain, and a flat list of components with their attributes.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in describeEntityInput) (*mcp.CallToolResult, any, error) {
		if err := requireFS(n); err != nil {
			return toolErr(err)
		}
		p := normalizePath(in.Path)
		var (
			e   *noitadata.Entity
			err error
		)
		if in.Raw {
			e, err = noitadata.ParseEntity(n, p)
		} else {
			e, err = noitadata.ResolveEntity(n, p)
		}
		if err != nil {
			return toolErr(err)
		}
		return textResult(formatEntity(e)), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "xrefs",
		Description: "List files that this path references (outbound) and files that reference this path (inbound). Backed by a one-time XML index built on startup.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in xrefsInput) (*mcp.CallToolResult, any, error) {
		if err := requireFS(n); err != nil {
			return toolErr(err)
		}
		p := normalizePath(in.Path)
		idx, err := n.Index()
		if err != nil {
			return toolErr(err)
		}
		var b strings.Builder
		fmt.Fprintf(&b, "# %s\n\n", p)
		fmt.Fprintf(&b, "Outbound (%d):\n", len(idx.Outbound[p]))
		for _, r := range idx.Outbound[p] {
			fmt.Fprintf(&b, "  -> %s\n", r)
		}
		fmt.Fprintf(&b, "\nInbound (%d):\n", len(idx.Inbound[p]))
		for _, r := range idx.Inbound[p] {
			fmt.Fprintf(&b, "  <- %s\n", r)
		}
		return textResult(b.String()), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "find_entities",
		Description: "Find entity XMLs containing a given component type, optionally filtered by attribute substring matches. Example: component='DamageModelComponent', where={'hp':'0.5'}.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in findEntitiesInput) (*mcp.CallToolResult, any, error) {
		if err := requireFS(n); err != nil {
			return toolErr(err)
		}
		if in.Component == "" {
			return toolErr(fmt.Errorf("component is required"))
		}
		limit := in.Limit
		if limit <= 0 {
			limit = 200
		}
		hits, err := n.FindEntitiesWith(in.Component, in.Where)
		if err != nil {
			return toolErr(err)
		}
		sort.Strings(hits)
		if len(hits) > limit {
			hits = hits[:limit]
		}
		var b strings.Builder
		fmt.Fprintf(&b, "%d entity file(s) match\n", len(hits))
		for _, h := range hits {
			fmt.Fprintln(&b, h)
		}
		return textResult(b.String()), nil, nil
	})
}

// formatEntity renders an Entity in a compact human-readable form for
// MCP text output.
func formatEntity(e *noitadata.Entity) string {
	var b strings.Builder
	fmt.Fprintf(&b, "entity: %s\n", e.SourcePath)
	fmt.Fprintf(&b, "  name: %s\n", e.Name)
	if len(e.Tags) > 0 {
		fmt.Fprintf(&b, "  tags: %s\n", strings.Join(e.Tags, ", "))
	}
	if len(e.BaseChain) > 0 {
		fmt.Fprintf(&b, "  base chain (root-first):\n")
		for _, p := range e.BaseChain {
			fmt.Fprintf(&b, "    - %s\n", p)
		}
	}
	fmt.Fprintf(&b, "  components (%d)%s:\n", len(e.Components), resolvedSuffix(e.Resolved))
	for _, c := range e.Components {
		fmt.Fprintf(&b, "    - %s", c.Type)
		if len(c.Tags) > 0 {
			fmt.Fprintf(&b, " [%s]", strings.Join(c.Tags, ","))
		}
		b.WriteByte('\n')
		// Keep attribute output deterministic.
		keys := make([]string, 0, len(c.Attrs))
		for k := range c.Attrs {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintf(&b, "        %s = %s\n", k, c.Attrs[k])
		}
		for _, ch := range c.Children {
			fmt.Fprintf(&b, "        <%s", ch.Type)
			chKeys := make([]string, 0, len(ch.Attrs))
			for k := range ch.Attrs {
				chKeys = append(chKeys, k)
			}
			sort.Strings(chKeys)
			for _, k := range chKeys {
				fmt.Fprintf(&b, " %s=%q", k, ch.Attrs[k])
			}
			fmt.Fprintf(&b, ">\n")
		}
	}
	refs := e.Refs()
	if len(refs) > 0 {
		fmt.Fprintf(&b, "  references (%d):\n", len(refs))
		for _, r := range refs {
			fmt.Fprintf(&b, "    - %s\n", r)
		}
	}
	return b.String()
}

func resolvedSuffix(resolved bool) string {
	if resolved {
		return " (resolved)"
	}
	return " (raw, no <Base> merge)"
}

// --- helpers ---

func normalizePath(p string) string {
	p = strings.TrimSpace(p)
	p = strings.Trim(p, "/")
	if p == "" {
		return "."
	}
	return path.Clean(p)
}

func textResult(s string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: s}},
	}
}

func requireFS(n *noitadata.FS) error {
	if n == nil {
		return fmt.Errorf("noita install unavailable; pass --install or set NOITA_PATH")
	}
	return nil
}

func toolErr(err error) (*mcp.CallToolResult, any, error) {
	if errors.Is(err, fs.ErrNotExist) {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{&mcp.TextContent{Text: "not found: " + err.Error()}},
		}, nil, nil
	}
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
	}, nil, nil
}

// doubleGlob implements a glob that supports ** segments by walking the fs.
func doubleGlob(fsys fs.FS, pattern string, limit int) ([]string, error) {
	if !strings.Contains(pattern, "**") {
		// Fast path via stdlib.
		out, err := fs.Glob(fsys, pattern)
		if err != nil {
			return nil, err
		}
		if limit > 0 && len(out) > limit {
			out = out[:limit]
		}
		return out, nil
	}
	parts := strings.Split(pattern, "/")
	// Root of the walk: longest leading prefix with no wildcard.
	var rootParts []string
	for _, p := range parts {
		if strings.ContainsAny(p, "*?[") {
			break
		}
		rootParts = append(rootParts, p)
	}
	root := "."
	if len(rootParts) > 0 {
		root = path.Join(rootParts...)
	}
	var out []string
	err := fs.WalkDir(fsys, root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return fs.SkipDir
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		ok, err := doubleMatch(pattern, p)
		if err != nil {
			return err
		}
		if ok {
			out = append(out, p)
			if len(out) >= limit {
				return fs.SkipAll
			}
		}
		return nil
	})
	return out, err
}

// doubleMatch matches path segment-by-segment, treating "**" as "zero or
// more segments".
func doubleMatch(pattern, name string) (bool, error) {
	pp := strings.Split(pattern, "/")
	np := strings.Split(name, "/")
	return segMatch(pp, np)
}

func segMatch(pp, np []string) (bool, error) {
	for len(pp) > 0 {
		if pp[0] == "**" {
			if len(pp) == 1 {
				return true, nil
			}
			for i := 0; i <= len(np); i++ {
				ok, err := segMatch(pp[1:], np[i:])
				if err != nil {
					return false, err
				}
				if ok {
					return true, nil
				}
			}
			return false, nil
		}
		if len(np) == 0 {
			return false, nil
		}
		ok, err := path.Match(pp[0], np[0])
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
		pp = pp[1:]
		np = np[1:]
	}
	return len(np) == 0, nil
}

func indexByteFrom(b []byte, c byte, start int) int {
	for i := start; i < len(b); i++ {
		if b[i] == c {
			return i
		}
	}
	return -1
}

func containsBytes(haystack, needle []byte) bool {
	if len(needle) == 0 {
		return true
	}
	for i := 0; i+len(needle) <= len(haystack); i++ {
		match := true
		for j := 0; j < len(needle); j++ {
			if haystack[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

