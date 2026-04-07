# Performance TODO

Profiled at ~81% of one CPU core over 30 seconds. Three cost centers:
process memory reads (~80%), overlay buffer clearing (~9%), drawing (~3%).

## 1. Cache component buffer metadata per-frame (noitrainer/noita)

`FindEntityComponentIDs` iterates ~50 component types per entity, reading
`ActiveCount`, `BeginPtr`, `EndPtr` individually each time. These are
properties of the component buffer, not the entity — they don't change
between entities within a single tick.

Read all component buffer metadata once per frame into a local cache,
then do in-memory lookups per entity instead of syscalls.

**Estimated savings: ~40% of total CPU**

Files: `noita/reader.go` (`FindEntityComponentIDs`, `hasComponent`,
`readAllComponents`)

## 2. ~~Bulk struct reads in generated code (memtools/hexpat + noita_gen.go)~~ DONE

hexpat codegen now detects static structs (no dynamic fields, known size)
and generates a single `ReadAt` for the entire struct, decoding fields
locally. Regenerated via `go generate`.

- `ReadComponentHeader`: 11 syscalls → 1 (72-byte bulk read)
- `ReadF64Vector` and similar flat structs: N syscalls → 1
- Nested children: parent + child = 2 total instead of N+M
- Dynamic structs (conditionals, variable-length arrays) unchanged
- Lazy `XReader` types still do per-field reads (lower priority)

## 3. Bulk F64Vector element reads (noita_gen.go)

`ReadF64Vector` was not covered by the codegen optimization (dynamic
array length, pointer-chased element storage). The 3 header fields are
still individual reads, and the element loop still does 1 syscall per
float64 (~60 for materials).

Two changes needed, both local to this codebase or as a manual override
of the generated code:
- Header: bulk-read the 12-byte header (BeginPtr/EndPtr/CapacityPtr) in 1 call
- Elements: single bulk read of `(EndPtr - BeginPtr)` bytes, then decode locally

**Estimated savings: ~7% of total CPU**

Files: `noita_gen.go` (`ReadF64Vector`)

## 4. Skip overlay frames when scene unchanged (go-overlay)

The overlay renders every vsync unconditionally — clearing ~8MB of pixel
buffer and redrawing even when the scene data hasn't changed. Add a
generation counter or dirty flag; skip render/clear/commit when nothing
changed.

**Estimated savings: ~9% of total CPU**

Files: `go-overlay/wayland.go` (`render`, frame callback)

## 5. Dirty region tracking in overlay (go-overlay)

`clear(fimg.Pix)` zeroes the entire buffer every frame and
`DamageBuffer(0, 0, w, h)` reports full-surface damage. Track which
pixel regions were drawn last frame, clear only those, and report the
dirty bounding box instead.

**Estimated savings: ~5% of total CPU**

Files: `go-overlay/wayland.go`, `go-overlay/pixbuf.go`

## 6. Defer expensive reads for non-displayed entities (noitrainer/noita)

`readPotionContents` reads `MaterialInventoryComponent` with its full
F64Vector for every entity that has one (~8% CPU). If potion contents are
only needed for displayed/selected entities, defer this read to
`updateOverlay` where filtering has already happened.

Files: `noita/reader.go` (`readPotionContents`, `ReadEntityList`)

## 7. Fast-path axis-aligned lines in overlay (go-overlay)

`strokeLine` uses Bresenham with per-pixel bounds checking. For
horizontal/vertical lines (common with hitbox rectangles), a bulk
memset-style write would be faster.

Files: `go-overlay/draw.go` (`strokeLine`, `setPixel`)
