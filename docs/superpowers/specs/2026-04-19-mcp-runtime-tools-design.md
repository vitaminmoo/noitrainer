# MCP runtime tools — exposing the CLI surface over MCP

## Context

`cmd/cli/main.go` is a rich introspection CLI for a running `noita.exe`:
entity/component browsing, raw-memory peek/deref, material/biome/chunk
lookups, pixel-scene dumps, NG+ count. It connects via `memtools/process` +
`noita.Reader`.

`cmd/noita-mcp/main.go` is an MCP server exposing the static Noita install
(data.wak + disk tree) — 9 tools covering filesystem + XML semantic queries.

Zero overlap exists today: an MCP client agent can read game XML but cannot
query the running game. This spec adds the runtime surface to the MCP server.

## Goal

Expose every CLI command that fits request/response semantics as an MCP tool
in the same `noita-mcp` binary.

## Scope

### In

All CLI commands except the three noted below, re-registered as MCP tools
(23 total):

| CLI command         | MCP tool name         | Notes                                    |
|---------------------|-----------------------|------------------------------------------|
| `entities`          | `entities`            |                                          |
| `entities-dump`     | `entities_dump`       | Returns JSON array (was NDJSON)          |
| `entity <id>`       | `entity`              |                                          |
| `tree <id>`         | `tree`                |                                          |
| `buffers`           | `buffers`             |                                          |
| `dump <id> <type>`  | `dump`                | Size input optional, default 256         |
| `components <id>`   | `components`          |                                          |
| `categorize`        | `categorize`          |                                          |
| `materials [f]`     | `materials`           | Optional `filter`                        |
| `material <id>`     | `material`            |                                          |
| `cell <wx> <wy>`    | `cell`                |                                          |
| `chunks [N]`        | `chunks`              | Optional `samples`, default 8            |
| `biome-grid`        | `biome_grid`          | Always JSON                              |
| `biome-chunk`       | `biome_chunk`         | Always JSON                              |
| `biome-at`          | `biome_at`            | Always JSON                              |
| `biome-dump [f]`    | `biome_dump`          | Optional `filter`                        |
| `biome-flags`       | `biome_flags`         | JSON array (was NDJSON)                  |
| `biome-at-many`     | `biome_at_many`       | `coords: [{wx,wy},...]` → JSON array     |
| `pixel-scenes`      | `pixel_scenes`        | JSON array (was NDJSON)                  |
| `ngplus`            | `ngplus`              |                                          |
| `peek <addr> [sz]`  | `peek`                | Addresses as strings, decimal or `0x…`   |
| `deref <addr> [sz]` | `deref`                |                                          |
| `read <type> <addr>`| `read_memory`         | Renamed to avoid `read_file` confusion   |

### Out

- `watch` — streaming loop, doesn't fit request/response. Agents that want
  current state can call `entities`/`components` etc. directly.
- Refactoring `cmd/cli/main.go` to share formatters. Minor duplication is
  accepted; both surfaces evolve independently and the CLI is stable.
- Structured `any` return from MCP tools. Existing tools in this server
  return text only; new tools match. JSON is emitted as text content.

## Design

### File layout

New file: `cmd/noita-mcp/runtime.go`. `registerRuntimeTools(s *mcp.Server)`
is called from `main()` alongside `registerTools` and `registerSemanticTools`.
Keeps ~600+ lines of handler code separated from the existing data tools.

### Connection lifecycle

Lazy, per-call. Each tool handler calls a new helper:

```go
func connectRuntime() (*noita.Reader, *process.Process, error)
```

that returns `process.FromName("noita.exe")[0]` wrapped in `noita.NewReader`,
or an error. Matches the CLI's per-command pattern, and makes each call
robust to the game being closed or restarted. No shared state across tools.

### Install path handling (startup)

Currently `main()` calls `log.Fatalf` if the install path can't be opened.
Relax this: log a warning, leave `n` nil, and let each static tool handler
return a `toolErr` if `n == nil`. Runtime tools are unaffected. Users who
only want runtime tools (game running, install not mounted) get a working
server.

### Code reuse strategy

Runtime handlers are written fresh against `noita.Reader`, producing text
via `strings.Builder`. They do *not* import CLI package code. This
duplicates formatting logic (hex dumps, damage multipliers, biome chunk
printers), which is accepted; a future refactor can extract shared
formatters to a `noita/cliformat` package if drift becomes a problem.

### Output shape

Text content in `CallToolResult` (matches existing server). Three sub-cases:

1. **Human-readable tables/summaries** — `entities`, `entity`, `tree`,
   `buffers`, `components`, `categorize`, `materials`, `material`, `cell`,
   `chunks`, `biome_dump`, `ngplus`, `peek`, `deref`, `read_memory`,
   `dump` — mirror the CLI's existing format as-is.
2. **Always-JSON** — `biome_grid`, `biome_chunk`, `biome_at` — always
   return JSON (CLI has `--json` flag; no reason to offer text in MCP).
3. **Bulk JSON** — `entities_dump`, `biome_flags`, `pixel_scenes`,
   `biome_at_many` — CLI emits NDJSON; MCP returns a single JSON **array**
   as text (cleaner for agents that will `JSON.parse` the result).

### Error handling

Replace CLI's `os.Exit(1)` call sites with `toolErr(err)` returns. No
panics, no exits. `connectRuntime` failures (Noita not running, permission
denied) surface as `{IsError: true, Content: "<error>"}`.

### Inputs

All input structs follow the existing `jsonschema:"..."` convention. Raw
addresses pass as strings (parsed with the CLI's `parseAddr`) so agents can
use `"0x00abcdef"` or `"123456"`.

Example:

```go
type peekInput struct {
    Addr string `json:"addr" jsonschema:"address, decimal or 0x-prefixed hex"`
    Size int    `json:"size,omitempty" jsonschema:"bytes to dump (default 128)"`
}

type biomeAtManyInput struct {
    Coords []struct {
        WX int32 `json:"wx"`
        WY int32 `json:"wy"`
    } `json:"coords" jsonschema:"list of world-pixel coordinates"`
}
```

## Non-goals

- Persistent/cached `Reader`. Overhead of per-call `FromName` + reader
  creation is negligible and the robustness win is real.
- Any write/mutation tool. All tools are read-only.
- Test coverage for each handler. Formatting is wide and shallow; testing
  every output shape adds maintenance burden for little value. Smoke-test
  that the server registers all tools and a representative handful of
  tools produce output against a running game (manual).

## Risks

- **Formatter drift** between CLI and MCP. Mitigated by: neither surface
  is the source of truth for on-wire data (the `noita.Reader` types are),
  so drift only affects human-readable presentation, not correctness.
- **MCP server no longer hard-requires a valid install path.** If an
  agent calls a static tool when the install failed to open, it gets an
  error — but this is the right failure mode (server usable for runtime
  tools alone).

## Verification

Manual, post-implementation:

1. `go build ./cmd/noita-mcp` — must compile.
2. Start server with Noita running and install path valid. Confirm all
   23 new tools are listed.
3. Call `entities`, `biome_at {wx:0,wy:0}`, `peek {addr:"0x00400000"}` —
   verify plausible output matching CLI equivalents.
4. Stop Noita; call `entities` — verify error (not panic).
5. Start server with bogus `--install` path; call `list_dir` — verify
   error (not panic) and `entities` still works if Noita is running.
