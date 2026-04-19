# MCP Runtime Tools Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Expose the runtime introspection surface from `cmd/cli/main.go` (23 commands) as MCP tools on the existing `cmd/noita-mcp` server.

**Architecture:** New file `cmd/noita-mcp/runtime.go` with one `registerRuntimeTools(s *mcp.Server)` function invoked from `main()`. Each tool handler lazily connects to `noita.exe` via a new `connectRuntime()` helper that returns errors instead of exiting. `main()` is relaxed to soft-fail on install-path open errors so runtime tools remain usable without a valid install mount. Formatters are freshly written to emit text (matching existing MCP tool output); CLI code is *not* refactored for sharing. Per the spec, JSON outputs are returned as text content — bulk tools return a JSON array (not NDJSON).

**Tech Stack:** Go 1.x; `github.com/modelcontextprotocol/go-sdk/mcp`; `noitrainer/noita`; `github.com/vitaminmoo/memtools/process`.

**Spec:** `docs/superpowers/specs/2026-04-19-mcp-runtime-tools-design.md`

**Testing approach:** Per the approved spec, runtime tools are not unit-tested — they depend on a live `noita.exe` process. Each task ends with a compile check (`go build ./cmd/noita-mcp`) and registers the expected tools (verified via `tools/list` over stdio where practical). Final task is a manual smoke test against a running game.

---

## Task 1: Scaffolding — `connectRuntime`, main() relaxation, empty registration

**Files:**
- Create: `cmd/noita-mcp/runtime.go`
- Modify: `cmd/noita-mcp/main.go`

### Step 1: Create `runtime.go` with connection helper and stub registration

- [ ] Write `cmd/noita-mcp/runtime.go`:

```go
// Runtime tools attach to the running noita.exe process and mirror the
// introspection surface of cmd/cli. Each tool opens a fresh Reader per
// call; there is no shared state between tools.
package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"noitrainer/noita"

	"github.com/vitaminmoo/memtools/process"
)

// connectRuntime finds the running noita.exe and returns a Reader bound to
// it. Returns an error when noita is not running or cannot be attached.
func connectRuntime() (*noita.Reader, *process.Process, error) {
	procs, err := process.FromName("noita.exe")
	if err != nil {
		return nil, nil, fmt.Errorf("noita.exe not found: %w", err)
	}
	if len(procs) == 0 {
		return nil, nil, fmt.Errorf("noita.exe not running")
	}
	proc := procs[0]
	return noita.NewReader(proc), proc, nil
}

// parseRuntimeAddr accepts decimal or 0x-prefixed hex as a 32-bit address.
// Mirrors cmd/cli/main.go:parseAddr.
func parseRuntimeAddr(s string) (uint32, error) {
	s = strings.TrimSpace(s)
	base := 10
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		s = s[2:]
		base = 16
	}
	v, err := strconv.ParseUint(s, base, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid address %q: %w", s, err)
	}
	return uint32(v), nil
}

// truncateStr mirrors cmd/cli/main.go:truncate.
func truncateStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-2] + ".."
}

// registerRuntimeTools registers every runtime introspection tool.
// Implementations are added in subsequent plan tasks.
func registerRuntimeTools(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "ngplus",
		Description: "Read the NG+ count from the running game.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in struct{}) (*mcp.CallToolResult, any, error) {
		reader, _, err := connectRuntime()
		if err != nil {
			return toolErr(err)
		}
		v, err := noita.ReadGNgPlusCount(reader.Ctx)
		if err != nil {
			return toolErr(fmt.Errorf("read ng+ count: %w", err))
		}
		return textResult(fmt.Sprintf("NG+ count: %d\n", v)), nil, nil
	})
}
```

### Step 2: Relax `main()` so install-open failures don't kill the server

- [ ] Edit `cmd/noita-mcp/main.go` lines 31-50 — replace the install-open block and add the runtime registration:

```go
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
```

### Step 3: Guard static-tool handlers against nil `n`

- [ ] In `cmd/noita-mcp/main.go`, add this helper near `toolErr`:

```go
func requireFS(n *noitadata.FS) error {
	if n == nil {
		return fmt.Errorf("noita install unavailable; pass --install or set NOITA_PATH")
	}
	return nil
}
```

- [ ] Inside each of the six handlers in `registerTools` (list_dir, stat, read_file, glob, extract_dir, search) and each of the three handlers in `registerSemanticTools` (describe_entity, xrefs, find_entities), add as the first statement inside the handler:

```go
		if err := requireFS(n); err != nil {
			return toolErr(err)
		}
```

### Step 4: Build

- [ ] Run: `go build ./cmd/noita-mcp`
- [ ] Expected: clean build, no errors.

### Step 5: Smoke-verify tool registration

- [ ] Run (with the game not running is fine):

```bash
printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"plan-test","version":"0"}}}' \
  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' \
  | ./noita-mcp 2>/dev/null | tail -1 | python3 -c 'import sys,json; r=json.load(sys.stdin); print("\n".join(sorted(t["name"] for t in r["result"]["tools"])))'
```

- [ ] Expected: a sorted list including `ngplus` alongside the 9 existing tools.

### Step 6: Commit

```bash
git add cmd/noita-mcp/runtime.go cmd/noita-mcp/main.go docs/superpowers/specs/2026-04-19-mcp-runtime-tools-design.md docs/superpowers/plans/2026-04-19-mcp-runtime-tools.md
git commit -m "noita-mcp: scaffold runtime tools (ngplus) and soften install hard-fail"
```

---

## Task 2: Entity inspection tools

Adds eight tools: `entities`, `entities_dump`, `entity`, `tree`, `buffers`, `dump`, `components`, `categorize`.

**Files:**
- Modify: `cmd/noita-mcp/runtime.go`

### Step 1: Add shared helpers

- [ ] Append to `runtime.go`:

```go
// buildRuntimeBufferNames mirrors cmd/cli/main.go:buildBufferNameMap.
func buildRuntimeBufferNames(reader *noita.Reader) map[noita.TypeID]string {
	names := make(map[noita.TypeID]string)
	for _, b := range reader.ReadComponentBuffers() {
		if b.Name != "" {
			names[noita.TypeID(b.TypeIndex)] = b.Name
		}
	}
	return names
}

// findEntityByIDRT mirrors cmd/cli/main.go:findEntityByID.
func findEntityByIDRT(reader *noita.Reader, id int32) *noita.EntitySummary {
	for _, e := range reader.ReadEntityList() {
		if e.Entity.EntityId == id {
			return e
		}
	}
	return nil
}
```

### Step 2: Add tool input structs

- [ ] Append to `runtime.go`:

```go
type entityIDInput struct {
	EntityID int32 `json:"entity_id" jsonschema:"EntityId of a live entity"`
}

type dumpInput struct {
	EntityID int32 `json:"entity_id"`
	TypeID   int   `json:"type_id" jsonschema:"component type index"`
	Size     int   `json:"size,omitempty" jsonschema:"bytes to dump (default 256)"`
}
```

### Step 3: Register `entities`, `entities_dump`, `buffers`, `components`, `dump`

- [ ] Append to the body of `registerRuntimeTools`:

```go
	mcp.AddTool(s, &mcp.Tool{
		Name:        "entities",
		Description: "List all live entities with id, name, position, and component names.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in struct{}) (*mcp.CallToolResult, any, error) {
		reader, _, err := connectRuntime()
		if err != nil {
			return toolErr(err)
		}
		nameMap := buildRuntimeBufferNames(reader)
		entities := reader.ReadEntityList()
		var b strings.Builder
		fmt.Fprintf(&b, "Found %d entities\n\n", len(entities))
		fmt.Fprintf(&b, "%-8s %-30s %-20s %s\n", "ID", "Name", "Position", "Components")
		fmt.Fprintf(&b, "%-8s %-30s %-20s %s\n", "──", "────", "────────", "──────────")
		for _, e := range entities {
			name := e.Name
			if name == "" {
				name = "(unnamed)"
			}
			pos := fmt.Sprintf("%.0f, %.0f", e.Entity.PosX, e.Entity.PosY)
			var compNames []string
			for _, cid := range e.ComponentIDs {
				if n, ok := nameMap[cid]; ok {
					compNames = append(compNames, n)
				} else {
					compNames = append(compNames, fmt.Sprintf("type_%d", cid))
				}
			}
			fmt.Fprintf(&b, "%-8d %-30s %-20s %s\n",
				e.Entity.EntityId, truncateStr(name, 29), pos, strings.Join(compNames, ", "))
		}
		return textResult(b.String()), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "entities_dump",
		Description: "Return a JSON array, one object per live entity: {entityId,name,x,y,ptr,components}.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in struct{}) (*mcp.CallToolResult, any, error) {
		reader, _, err := connectRuntime()
		if err != nil {
			return toolErr(err)
		}
		nameMap := buildRuntimeBufferNames(reader)
		type row struct {
			EntityID   int32    `json:"entityId"`
			Name       string   `json:"name"`
			X          float32  `json:"x"`
			Y          float32  `json:"y"`
			Ptr        string   `json:"ptr"`
			Components []string `json:"components"`
		}
		entities := reader.ReadEntityList()
		out := make([]row, 0, len(entities))
		for _, e := range entities {
			compNames := make([]string, 0, len(e.ComponentIDs))
			for _, cid := range e.ComponentIDs {
				if n, ok := nameMap[cid]; ok {
					compNames = append(compNames, n)
				} else {
					compNames = append(compNames, fmt.Sprintf("type_%d", cid))
				}
			}
			out = append(out, row{
				EntityID:   e.Entity.EntityId,
				Name:       e.Name,
				X:          e.Entity.PosX,
				Y:          e.Entity.PosY,
				Ptr:        fmt.Sprintf("0x%08x", e.Ptr),
				Components: compNames,
			})
		}
		data, err := json.Marshal(out)
		if err != nil {
			return toolErr(err)
		}
		return textResult(string(data)), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "buffers",
		Description: "List all component buffers (the type registry): TypeID, Name, ActiveCount, Capacity, Ptr.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in struct{}) (*mcp.CallToolResult, any, error) {
		reader, _, err := connectRuntime()
		if err != nil {
			return toolErr(err)
		}
		buffers := reader.ReadComponentBuffers()
		var b strings.Builder
		fmt.Fprintf(&b, "Found %d component buffers\n\n", len(buffers))
		fmt.Fprintf(&b, "%-6s %-45s %-10s %-10s %s\n", "TypeID", "Name", "Active", "Capacity", "Ptr")
		fmt.Fprintf(&b, "%-6s %-45s %-10s %-10s %s\n", "──────", "────", "──────", "────────", "───")
		for _, bu := range buffers {
			name := bu.Name
			if name == "" {
				name = "(unnamed)"
			}
			fmt.Fprintf(&b, "%-6d %-45s %-10d %-10d 0x%08X\n",
				bu.TypeIndex, truncateStr(name, 44), bu.ActiveCount, bu.Capacity, bu.Ptr)
		}
		return textResult(b.String()), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "components",
		Description: "List component types present on an entity, with their in-memory pointers.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in entityIDInput) (*mcp.CallToolResult, any, error) {
		reader, _, err := connectRuntime()
		if err != nil {
			return toolErr(err)
		}
		nameMap := buildRuntimeBufferNames(reader)
		e := findEntityByIDRT(reader, in.EntityID)
		if e == nil {
			return toolErr(fmt.Errorf("entity %d not found", in.EntityID))
		}
		em, _ := reader.ReadEntityManagerPtr()
		if em == nil {
			return toolErr(fmt.Errorf("failed to read EntityManager"))
		}
		compIDs := reader.FindEntityComponentIDs(em, e.Entity.SlotIndex)
		name := e.Name
		if name == "" {
			name = "(unnamed)"
		}
		var b strings.Builder
		fmt.Fprintf(&b, "Entity %d (%s) has %d component types:\n\n", in.EntityID, name, len(compIDs))
		fmt.Fprintf(&b, "%-6s %-45s %s\n", "TypeID", "Name", "Ptr")
		fmt.Fprintf(&b, "%-6s %-45s %s\n", "──────", "────", "───")
		for _, cid := range compIDs {
			compName := fmt.Sprintf("type_%d", cid)
			if n, ok := nameMap[cid]; ok {
				compName = n
			}
			compPtr, _ := reader.ReadRawComponent(em, e.Entity.SlotIndex, cid, 0)
			fmt.Fprintf(&b, "%-6d %-45s 0x%08X\n", cid, compName, compPtr)
		}
		return textResult(b.String()), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "dump",
		Description: "Hex-dump raw component bytes for entity + type id. size defaults to 256.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in dumpInput) (*mcp.CallToolResult, any, error) {
		reader, _, err := connectRuntime()
		if err != nil {
			return toolErr(err)
		}
		e := findEntityByIDRT(reader, in.EntityID)
		if e == nil {
			return toolErr(fmt.Errorf("entity %d not found", in.EntityID))
		}
		em, _ := reader.ReadEntityManagerPtr()
		if em == nil {
			return toolErr(fmt.Errorf("failed to read EntityManager"))
		}
		nameMap := buildRuntimeBufferNames(reader)
		typeID := noita.TypeID(in.TypeID)
		compName := fmt.Sprintf("type_%d", typeID)
		if n, ok := nameMap[typeID]; ok {
			compName = n
		}
		size := in.Size
		if size <= 0 {
			size = 256
		}
		compPtr, data := reader.ReadRawComponent(em, e.Entity.SlotIndex, typeID, size)
		if data == nil {
			return toolErr(fmt.Errorf("entity %d has no component of type %d (%s)", in.EntityID, typeID, compName))
		}
		var b strings.Builder
		fmt.Fprintf(&b, "Entity %d, %s (type %d) @ 0x%08X, %d bytes:\n\n", in.EntityID, compName, typeID, compPtr, size)
		b.WriteString(hex.Dump(data))
		return textResult(b.String()), nil, nil
	})
```

### Step 4: Register `entity` and `tree`

These call deeper APIs. Implement by porting `cmdEntity` from `cmd/cli/main.go:444-545` and `cmdTree` from `cmd/cli/main.go:574-643` — same logic, but:
- Input: `entityIDInput`.
- Output: `strings.Builder` → `textResult(b.String())`.
- Replace `os.Exit(1)` call sites with `return toolErr(err)`.
- Replace `fmt.Fprintf(os.Stderr, ...)` with `fmt.Errorf(...)` passed to `toolErr`.
- Helper `printDmgMults` from `cmd/cli/main.go:547-570` becomes `appendDmgMultsRT(b *strings.Builder, d *noita.DamageModelComponent)` — same body but writes to `b`.
- Helper `printChildTree` becomes `appendChildTreeRT(b *strings.Builder, reader *noita.Reader, parent *noita.EntitySummary, depth, maxDepth int)` — same body with writes redirected to `b`.

- [ ] Add the two helpers and the two tool registrations. Signatures:

```go
func appendDmgMultsRT(b *strings.Builder, d *noita.DamageModelComponent) { /* body mirrors printDmgMults */ }
func appendChildTreeRT(b *strings.Builder, reader *noita.Reader, parent *noita.EntitySummary, depth, maxDepth int) { /* body mirrors printChildTree */ }

mcp.AddTool(s, &mcp.Tool{
    Name:        "entity",
    Description: "Show detailed info for an entity (position, flags, HP, character, wallet, inventory, ability, children).",
}, func(ctx context.Context, req *mcp.CallToolRequest, in entityIDInput) (*mcp.CallToolResult, any, error) { /* port cmdEntity */ })

mcp.AddTool(s, &mcp.Tool{
    Name:        "tree",
    Description: "Show an entity's parent-chain (root-first) then its subtree to depth 2.",
}, func(ctx context.Context, req *mcp.CallToolRequest, in entityIDInput) (*mcp.CallToolResult, any, error) { /* port cmdTree */ })
```

### Step 5: Register `categorize`

- [ ] Port `cmdCategorize` from `cmd/cli/main.go:737-840` — all reads are `reader.ReadEntityList()` + `nameMap`. No external state. Output goes to a `strings.Builder`, returned via `textResult`. Input is empty struct.

```go
mcp.AddTool(s, &mcp.Tool{
    Name:        "categorize",
    Description: "Group live entities by name and by component signature; show counts and top members.",
}, func(ctx context.Context, req *mcp.CallToolRequest, in struct{}) (*mcp.CallToolResult, any, error) { /* port cmdCategorize */ })
```

### Step 6: Update imports

- [ ] Ensure `runtime.go` imports include `encoding/hex` and `encoding/json`. Verify the compiler reports no unused imports.

### Step 7: Build

- [ ] Run: `go build ./cmd/noita-mcp`
- [ ] Expected: clean build.

### Step 8: Verify tool registration

- [ ] Repeat the `tools/list` smoke from Task 1, Step 5 and confirm all eight new names appear:
  `entities, entities_dump, entity, tree, buffers, dump, components, categorize`.

### Step 9: Commit

```bash
git add cmd/noita-mcp/runtime.go
git commit -m "noita-mcp: add entity inspection tools"
```

---

## Task 3: Material / chunk / cell tools

Adds four tools: `materials`, `material`, `cell`, `chunks`.

**Files:**
- Modify: `cmd/noita-mcp/runtime.go`

### Step 1: Input structs and helper

- [ ] Append:

```go
type materialsInput struct {
	Filter string `json:"filter,omitempty" jsonschema:"optional substring match on material name"`
}
type materialInput struct {
	MaterialID int `json:"material_id"`
}
type cellInput struct {
	WX int32 `json:"wx"`
	WY int32 `json:"wy"`
}
type chunksInput struct {
	Samples int `json:"samples,omitempty" jsonschema:"max loaded chunks to sample (default 8)"`
}

// indentHexDumpRT mirrors cmd/cli/main.go:indentHexDump.
func indentHexDumpRT(buf []byte, base uint32) string {
	dump := hex.Dump(buf)
	lines := strings.Split(strings.TrimRight(dump, "\n"), "\n")
	var out strings.Builder
	for _, ln := range lines {
		if len(ln) >= 10 {
			off64, err := strconv.ParseUint(strings.TrimSpace(ln[:8]), 16, 32)
			if err == nil {
				fmt.Fprintf(&out, "  %08x%s\n", uint32(off64)+base, ln[8:])
				continue
			}
		}
		out.WriteString("  ")
		out.WriteString(ln)
		out.WriteByte('\n')
	}
	return out.String()
}
```

### Step 2: Register four tools

- [ ] Port the handlers from `cmd/cli/main.go`:
  - `materials` ← `cmdMaterials` (lines 844-872)
  - `material` ← `cmdMaterial` (lines 874-913)
  - `cell` ← `cmdCell` (lines 917-949)
  - `chunks` ← `cmdChunks` (lines 951-971)

All four follow the same translation rules as Task 2 Step 4. The `material` handler uses `indentHexDumpRT` for its final byte preview (replacing `cmdCell`'s `indentHexDump`). Descriptions:

```go
mcp.AddTool(s, &mcp.Tool{Name: "materials", Description: "List CellFactory materials; optional name substring filter."}, ...)
mcp.AddTool(s, &mcp.Tool{Name: "material",  Description: "Show full CellData for a material id."}, ...)
mcp.AddTool(s, &mcp.Tool{Name: "cell",      Description: "Resolve a world pixel to its chunk/cell pointers; dump first 0x40 bytes of the cell."}, ...)
mcp.AddTool(s, &mcp.Tool{Name: "chunks",    Description: "Show ChunkSystem stats (total chunks, coord range) and sample some loaded chunks."}, ...)
```

### Step 3: Build

- [ ] Run: `go build ./cmd/noita-mcp` — expect clean build.

### Step 4: Verify registration

- [ ] `tools/list` shows `materials`, `material`, `cell`, `chunks`.

### Step 5: Commit

```bash
git add cmd/noita-mcp/runtime.go
git commit -m "noita-mcp: add material/chunk/cell tools"
```

---

## Task 4: Biome + pixel-scene tools

Adds seven tools: `biome_grid`, `biome_chunk`, `biome_at`, `biome_dump`, `biome_flags`, `biome_at_many`, `pixel_scenes`.

Important shape change from CLI: the `--json` flag on `biome_grid`/`biome_chunk`/`biome_at` is removed — these three *always* emit JSON. `biome_flags`, `biome_at_many`, and `pixel_scenes` emit a single JSON **array** (not NDJSON).

**Files:**
- Modify: `cmd/noita-mcp/runtime.go`

### Step 1: Input structs

- [ ] Append:

```go
type biomeChunkInput struct {
	CX int32 `json:"cx"`
	CY int32 `json:"cy"`
}
type biomeAtInput struct {
	WX int32 `json:"wx"`
	WY int32 `json:"wy"`
}
type biomeDumpInput struct {
	Filter string `json:"filter,omitempty" jsonschema:"optional substring match on biome name"`
}
type coord struct {
	WX int32 `json:"wx"`
	WY int32 `json:"wy"`
}
type biomeAtManyInput struct {
	Coords []coord `json:"coords" jsonschema:"list of world-pixel coordinates"`
}
```

### Step 2: Register `biome_grid`, `biome_chunk`, `biome_at`

- [ ] These three always return JSON. Example (`biome_grid`):

```go
mcp.AddTool(s, &mcp.Tool{
    Name:        "biome_grid",
    Description: "JSON: biome chunk grid header (width, height, shifts, chunks_ptr) with loaded-chunk count.",
}, func(ctx context.Context, req *mcp.CallToolRequest, in struct{}) (*mcp.CallToolResult, any, error) {
    reader, _, err := connectRuntime()
    if err != nil {
        return toolErr(err)
    }
    g := reader.ReadBiomeGridInfo()
    if g == nil {
        return toolErr(fmt.Errorf("biome grid unavailable (WorldManager.pBackgroundGrid is null)"))
    }
    loaded := 0
    reader.IterateBiomeChunks(func(*noita.BiomeChunkInfo) bool { loaded++; return true })
    data, err := json.Marshal(struct {
        *noita.BiomeGridInfo
        Loaded int `json:"loaded"`
    }{g, loaded})
    if err != nil {
        return toolErr(err)
    }
    return textResult(string(data)), nil, nil
})
```

- [ ] `biome_chunk`: call `reader.ReadBiomeChunkInfo(in.CX, in.CY)`, JSON-marshal the `*BiomeChunkInfo` (nil marshals to `null`), return as text.
- [ ] `biome_at`: call `reader.ResolveBiomeAt(in.WX, in.WY)`; error if the grid is unavailable (check `reader.ReadBiomeGridInfo() == nil` first); JSON-marshal result.

### Step 3: Register `biome_dump`

- [ ] Port `cmdBiomeDump` from `cmd/cli/main.go:1175-1199` to a tool accepting `biomeDumpInput`. Output is human-readable (matches CLI). Same translation rules.

### Step 4: Register `biome_flags`, `pixel_scenes` (JSON-array bulk tools)

- [ ] `biome_flags` — iterate via `reader.IterateBiomeChunks`, skip chunks whose `Name` is `""`/`"_EMPTY_"`/`"???"`, collect the same struct as `cmdBiomeFlags` (`cmd/cli/main.go:1090-1122`) into a slice, then emit `json.Marshal(slice)`.

- [ ] `pixel_scenes` — iterate via `reader.IteratePixelScenes`, collect `*PixelSceneInfo` values into a slice, `json.Marshal`.

Both must fail with a reasonable error if `reader.ReadBiomeGridInfo() == nil`.

### Step 5: Register `biome_at_many`

- [ ] Accepts `biomeAtManyInput`, returns a JSON array of `BiomeAtResult`:

```go
mcp.AddTool(s, &mcp.Tool{
    Name:        "biome_at_many",
    Description: "Resolve biomes at many world-pixel coordinates; returns a JSON array of BiomeAtResult objects in input order.",
}, func(ctx context.Context, req *mcp.CallToolRequest, in biomeAtManyInput) (*mcp.CallToolResult, any, error) {
    reader, _, err := connectRuntime()
    if err != nil {
        return toolErr(err)
    }
    if reader.ReadBiomeGridInfo() == nil {
        return toolErr(fmt.Errorf("biome grid unavailable"))
    }
    out := make([]*noita.BiomeAtResult, 0, len(in.Coords))
    for _, c := range in.Coords {
        out = append(out, reader.ResolveBiomeAt(c.WX, c.WY))
    }
    data, err := json.Marshal(out)
    if err != nil {
        return toolErr(err)
    }
    return textResult(string(data)), nil, nil
})
```

### Step 6: Build

- [ ] Run: `go build ./cmd/noita-mcp` — expect clean build.

### Step 7: Verify registration

- [ ] `tools/list` shows all seven new names.

### Step 8: Commit

```bash
git add cmd/noita-mcp/runtime.go
git commit -m "noita-mcp: add biome and pixel-scene tools"
```

---

## Task 5: Raw-memory tools

Adds three tools: `peek`, `deref`, `read_memory`.

**Files:**
- Modify: `cmd/noita-mcp/runtime.go`

### Step 1: Input structs

- [ ] Append:

```go
type peekInput struct {
	Addr string `json:"addr" jsonschema:"address, decimal or 0x-prefixed hex"`
	Size int    `json:"size,omitempty" jsonschema:"bytes to dump (default 128)"`
}
type derefInput struct {
	Addr string `json:"addr" jsonschema:"address of a u32 pointer to dereference"`
	Size int    `json:"size,omitempty" jsonschema:"bytes to dump at the pointee (default 128)"`
}
type readMemInput struct {
	Type string `json:"type" jsonschema:"u8|u16|u32|u64|s32|f32|f64|str|ptr"`
	Addr string `json:"addr"`
}
```

### Step 2: Register `peek`

- [ ] Append:

```go
mcp.AddTool(s, &mcp.Tool{
    Name:        "peek",
    Description: "Hex-dump arbitrary virtual memory at addr. size defaults to 128.",
}, func(ctx context.Context, req *mcp.CallToolRequest, in peekInput) (*mcp.CallToolResult, any, error) {
    reader, _, err := connectRuntime()
    if err != nil {
        return toolErr(err)
    }
    addr, err := parseRuntimeAddr(in.Addr)
    if err != nil {
        return toolErr(err)
    }
    size := in.Size
    if size <= 0 {
        size = 128
    }
    buf := make([]byte, size)
    n, err := reader.Ctx.ReadAt(buf, int64(addr))
    if err != nil {
        return toolErr(fmt.Errorf("read at 0x%08X: %w", addr, err))
    }
    var b strings.Builder
    fmt.Fprintf(&b, "0x%08X, %d bytes:\n\n%s", addr, n, indentHexDumpRT(buf[:n], addr))
    return textResult(b.String()), nil, nil
})
```

### Step 3: Register `deref`

- [ ] Port `cmdDeref` (`cmd/cli/main.go:1267-1287`) with the same translation rules. Input: `derefInput`. Use `binary.LittleEndian.Uint32` — add `"encoding/binary"` to imports.

### Step 4: Register `read_memory`

- [ ] Port `cmdRead` (`cmd/cli/main.go:1289-1339`). Input: `readMemInput`. All `fmt.Printf` → `fmt.Fprintf(&b, ...)`; all `os.Exit(1)` → `return toolErr(...)`. Use `math.Float32frombits`/`Float64frombits` — add `"math"` to imports. For the `"str"` case, use `noita.ReadMsvcString(reader.Ctx, uintptr(addr))` and return both its metadata and `FormatMsvcString(reader.Ctx)` output.

Description:

```go
mcp.AddTool(s, &mcp.Tool{
    Name:        "read_memory",
    Description: "Read a typed value at addr. type is one of: u8, u16, u32, u64, s32, f32, f64, str, ptr.",
}, ...)
```

### Step 5: Build

- [ ] Run: `go build ./cmd/noita-mcp` — expect clean build.

### Step 6: Verify registration

- [ ] `tools/list` shows `peek`, `deref`, `read_memory`.

### Step 7: Commit

```bash
git add cmd/noita-mcp/runtime.go
git commit -m "noita-mcp: add raw-memory peek/deref/read tools"
```

---

## Task 6: Final verification

**Files:** none.

### Step 1: Full tool-list check

- [ ] With the server started and `tools/list` called, verify exactly these **32** tool names appear:

Existing (9): `list_dir stat read_file glob extract_dir search describe_entity xrefs find_entities`

New (23): `entities entities_dump entity tree buffers dump components categorize materials material cell chunks biome_grid biome_chunk biome_at biome_dump biome_flags biome_at_many pixel_scenes ngplus peek deref read_memory`

### Step 2: Smoke against a running game

Start Noita and launch a run so the world is loaded. Then drive the server over stdio and sanity-check output for a representative subset:

- [ ] `ngplus` — matches `./noitrainer-cli ngplus`.
- [ ] `entities` — non-empty list including a `player_unknown`-prefixed entity.
- [ ] `entity {"entity_id": <player id from entities>}` — contains `DamageModelComponent`, `CharacterDataComponent`.
- [ ] `biome_at {"wx": 0, "wy": 0}` — emits JSON with `.original` and `.resolved` populated.
- [ ] `biome_flags` — JSON array, non-empty.
- [ ] `pixel_scenes` — JSON array (may be empty early in a run; not an error).
- [ ] `peek {"addr": "0x00400000", "size": 64}` — 64 bytes of hex.
- [ ] `read_memory {"type": "u32", "addr": "0x00400000"}` — prints a u32 value.

### Step 3: Failure-mode check

- [ ] Stop Noita. Call `entities` — must return `IsError: true` with a human-readable "noita.exe not running" message.
- [ ] Start the server with `--install=/tmp/definitely-not-a-noita-install`. Call `list_dir {"path": "."}` — must return `IsError: true` with "install unavailable". Call `entities` (with Noita running) — must still succeed.

### Step 4: Commit

```bash
# The plan file was already tracked in Task 1's commit; if the plan or spec
# was amended during execution, commit those changes now; otherwise skip.
git status
```

---

## Self-review

**Spec coverage.** All 23 in-scope CLI commands have a task. `watch` is explicitly omitted (spec §Out). Install-path soft-fail is covered (Task 1 Steps 2-3). Bulk JSON tools (`entities_dump`, `biome_flags`, `pixel_scenes`, `biome_at_many`) return single arrays, per spec §Output shape #3.

**Placeholder scan.** No "TBD"/"TODO". Large-port steps (e.g., Task 2 Step 4 for `entity`/`tree`) name the exact source lines, the exact transformation rules, and specify the output helper signatures so the engineer has a complete recipe.

**Type consistency.** `entityIDInput`, `biomeChunkInput`, `biomeAtInput`, `coord`, `biomeAtManyInput`, `peekInput`, `derefInput`, `readMemInput`, `dumpInput`, `materialsInput`, `materialInput`, `cellInput`, `chunksInput`, `biomeDumpInput` — each defined once and referenced as-is downstream. Helpers `connectRuntime`, `buildRuntimeBufferNames`, `findEntityByIDRT`, `parseRuntimeAddr`, `truncateStr`, `indentHexDumpRT`, `appendDmgMultsRT`, `appendChildTreeRT` — each defined before use.
