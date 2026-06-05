# noitrainer

Live introspection tooling for [Noita](https://noitagame.com/). It attaches to a
running `noita.exe`, walks the game's in-memory structures, and exposes them three
ways:

- a **terminal UI** (`./noitrainer`) showing player / entity / wand / world state,
  with an optional click-through on-screen overlay that draws entity positions and
  hitboxes on top of the game;
- a **CLI** (`cmd/cli`) for one-shot introspection (entities, components, raw
  memory peek/deref, materials, biomes, chunks, pixel scenes, NG+ count, ...);
- an **MCP server** (`cmd/noita-mcp`) that serves both the static install
  (`data/data.wak` + the on-disk tree, with XML semantic queries) and the same
  runtime introspection surface, so an agent can read the game files *and* query
  the live game.

> [!WARNING]
> **Pre-alpha, and built for one machine.** This is a personal project shaped
> entirely around the author's setup (Linux + Wayland/sway, Noita via Steam
> Proton). It is shared in case it's useful, not as a polished release. Expect to
> read and edit the source to get it working anywhere else. APIs, layout, and the
> memory offsets it depends on will change without notice.

## Status & limitations

- **Linux only for anything that touches the running game.** Memory access goes
  through [`memtools`](https://github.com/vitaminmoo/memtools), which reads
  `/proc` and uses the Linux `process_vm_readv` syscall. It has no macOS/Windows
  implementation and will not compile there (see [Porting](#porting-to-macos--windows)).
- **The overlay is Linux + Wayland only**, via
  [`go-overlay`](https://github.com/vitaminmoo/go-overlay) (`wlr-layer-shell`).
  Best on sway; KDE/KWin is partial; X11/macOS/Windows are unsupported. The TUI
  still works without it.
- **Pinned to one Noita build.** Root pointers are hardcoded absolute addresses
  (`noita/noita_gen.go`, e.g. `AddrGEntityManager = 0x01204B98`) against the
  32-bit x86 `noita.exe`, last validated **2026-03-18**. A Noita update almost
  certainly breaks them — there is no signature/pattern scanning yet. You will
  need to re-find the offsets (the `.hexpat` and Ghidra comments in
  `noita/reader.go` document where they came from).
- **No stability guarantees** on the API, the offsets, or behavior.

## Requirements

- Go (see `go.mod` for the version)
- Linux
- Noita (the 32-bit build) running — on Linux that's via Steam Play / Proton,
  where the process is still named `noita.exe`
- For the overlay: a Wayland compositor implementing `wlr-layer-shell` (sway
  recommended)
- To attach to the process you must be able to read its memory
  (`process_vm_readv` / `ptrace`). Run as the same user; on hardened kernels you
  may need to relax `kernel.yama.ptrace_scope` or grant `CAP_SYS_PTRACE`.

## Build

```sh
go build -o noitrainer .              # the TUI
go build -o noitrainer-cli ./cmd/cli  # the CLI
go build -o noita-mcp ./cmd/noita-mcp # the MCP server
```

## Usage

### TUI

```sh
./noitrainer
```

Tabs: `Player`, `Entities`, `Wands`, `World`, `Overlay`, `Log`. `tab` /
`shift+tab` switch tabs, `q` or `ctrl+c` quits. It waits for `noita.exe` and
starts reading once the game is up.

> [!NOTE]
> The TUI unconditionally starts a pprof HTTP server on `:4200`
> (`main.go`). If that port is taken, or you don't want it listening, edit or
> remove that block.

### CLI

```sh
./noitrainer-cli              # prints the command list
./noitrainer-cli entities
./noitrainer-cli entity 123
./noitrainer-cli material water
./noitrainer-cli peek 0x01204B98 64
```

### MCP server

Speaks MCP over stdio. The static tools (filesystem + XML queries over the Noita
install) work on any platform; the runtime tools require a running game on Linux.
Install detection order:

1. `--install <dir>` flag, else
2. `$NOITA_PATH`, else
3. platform Steam defaults (`noitadata/install.go`).

A directory "looks like Noita" if it contains `data/data.wak`.

## Layout

| Path             | What it is                                                        |
|------------------|-------------------------------------------------------------------|
| `main.go`        | TUI (bubbletea), process attach, render loop                      |
| `overlay.go`     | Wayland overlay glue (entity dots + hitboxes)                     |
| `scheduler.go`   | Read scheduling for the TUI/overlay                               |
| `noita/`         | Memory model: `noita.hexpat` (ImHex pattern), `noita_gen.go` (generated reader), `reader.go` (higher-level reads) |
| `noitadata/`     | Static install access: `data.wak` unpack, disk FS, XML entity index — **no memory access, cross-platform** |
| `cmd/cli/`       | One-shot introspection CLI                                         |
| `cmd/noita-mcp/` | MCP server: static (`main.go`) + runtime (`runtime.go`) tools     |

`noita/noita_gen.go` is generated from `noita/noita.hexpat`; regenerate with
`go generate` after editing offsets.

## Landmines / known rough edges

These are the things most likely to bite, beyond the platform limits above:

- **Hardcoded offsets rot silently.** When the game updates, reads return garbage
  rather than failing loudly. If numbers look wrong, suspect the offsets first.
- **Process name is hardcoded** to `noita.exe` (`main.go`, `cmd/cli`,
  `cmd/noita-mcp/runtime.go`). Fine under Proton; a native macOS/Windows build
  would use a different name.
- **The pprof `:4200` listener** in the TUI is always on (see note above).
- **`noita-mcp` mixes cross-platform and Linux-only tools in one binary**, so the
  whole thing currently only builds on Linux even though the static half doesn't
  need memory access. See below.
- **CPU cost.** Live reads are syscall-heavy; see `PERFORMANCE-TODO.md` for the
  profiled hotspots and the remaining optimization ideas.

## Porting to macOS / Windows

The work splits cleanly along the memory boundary:

- **Already portable:** the `noitadata` package and the *static* MCP tools — they
  only read files. The blocker to building `noita-mcp` elsewhere is that it
  registers the runtime (memory) tools in the same binary. Putting the runtime
  registration behind a build tag (or a separate binary) would let the static
  server build and run on macOS/Windows today.
- **Needs a platform backend:** `memtools` itself — it would need a macOS
  (`mach_vm_read` / `task_for_pid`, which requires entitlements or disabling SIP)
  and Windows (`ReadProcessMemory` / `OpenProcess`) implementation behind build
  tags. Until then, the TUI, the CLI, and the runtime MCP tools are Linux-only.
- **The overlay** would need a non-Wayland backend (or to be made optional) for
  the TUI to render its on-screen layer off Linux.
- **Offsets** are build-specific, not OS-specific — but the native macOS/Windows
  Noita builds differ from the Proton-run `noita.exe`, so the addresses would
  have to be re-derived per target anyway.

## License

[MIT](LICENSE).
