package main

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"noitrainer/noita"

	"github.com/vitaminmoo/memtools/process"
)

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: noitrainer-cli <command> [args]

Commands:
  watch                  Watch game state changes (original behavior)
  entities               List all entities with component info
  entities-dump          Emit NDJSON for every entity (id, name, pos, components)
  entity <id>            Show detailed info for entity by ID
  tree <id>              Show entity parent/child tree
  buffers                List all component buffers (type registry)
  dump <id> <type>       Hex dump raw component bytes for entity ID + type ID
  components <id>        List all component types on an entity
  categorize             Categorize entities by name and component signature
  materials [filter]     List CellFactory materials (optional name substring filter)
  material <id>          Show full CellData for a material ID
  cell <wx> <wy>         Resolve a world pixel to its chunk/cell pointers
  chunks [N]             Show ChunkSystem stats; sample up to N loaded chunks (default 8)
  biome-grid             Show biome chunk grid header (width/height/shifts/chunks ptr)
  biome-chunk <cx> <cy>  Show biome chunk at grid coords (name + wobble flags)
  biome-at <wx> <wy>     Resolve biome at a world pixel (with wobble decision)
  biome-dump [filter]    List every loaded biome chunk (cx,cy,name,flags); optional name substring
  biome-flags            Emit JSON-lines (cx,cy,name,wobble flags) for every named chunk
  biome-at-many          Read "wx wy" lines from stdin; emit one JSON line of resolution per coord
  pixel-scenes           List every pixel scene currently queued in the BiomeGrid (NDJSON)
  ngplus                 Show NG+ count

Most biome-* commands accept --json before/after positional args to emit a
single JSON object instead of the human-readable layout.
  peek <addr> [size]     Hex dump arbitrary virtual memory (size default 128)
  deref <addr> [size]    Read u32 at <addr>, then hex dump [size] bytes at that pointer
  read <type> <addr>     Read typed value at address. type=u8|u16|u32|u64|s32|f32|f64|str|ptr

Addresses accept decimal or 0x-prefixed hex.
`)
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	cmd := os.Args[1]

	switch cmd {
	case "watch":
		cmdWatch()
	case "entities":
		cmdEntities()
	case "entities-dump":
		cmdEntitiesDump()
	case "entity":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Usage: noitrainer-cli entity <entity-id>\n")
			os.Exit(1)
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid entity ID: %s\n", os.Args[2])
			os.Exit(1)
		}
		cmdEntity(int32(id))
	case "tree":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Usage: noitrainer-cli tree <entity-id>\n")
			os.Exit(1)
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid entity ID: %s\n", os.Args[2])
			os.Exit(1)
		}
		cmdTree(int32(id))
	case "buffers":
		cmdBuffers()
	case "dump":
		if len(os.Args) < 4 {
			fmt.Fprintf(os.Stderr, "Usage: noitrainer-cli dump <entity-id> <type-id> [size]\n")
			os.Exit(1)
		}
		entityID, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid entity ID: %s\n", os.Args[2])
			os.Exit(1)
		}
		typeID, err := strconv.Atoi(os.Args[3])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid type ID: %s\n", os.Args[3])
			os.Exit(1)
		}
		size := 256
		if len(os.Args) >= 5 {
			size, err = strconv.Atoi(os.Args[4])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid size: %s\n", os.Args[4])
				os.Exit(1)
			}
		}
		cmdDump(int32(entityID), noita.TypeID(typeID), size)
	case "components":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Usage: noitrainer-cli components <entity-id>\n")
			os.Exit(1)
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid entity ID: %s\n", os.Args[2])
			os.Exit(1)
		}
		cmdComponents(int32(id))
	case "categorize":
		cmdCategorize()
	case "materials":
		filter := ""
		if len(os.Args) >= 3 {
			filter = os.Args[2]
		}
		cmdMaterials(filter)
	case "material":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Usage: noitrainer-cli material <material-id>\n")
			os.Exit(1)
		}
		matID, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid material ID: %s\n", os.Args[2])
			os.Exit(1)
		}
		cmdMaterial(matID)
	case "cell":
		if len(os.Args) < 4 {
			fmt.Fprintf(os.Stderr, "Usage: noitrainer-cli cell <wx> <wy>\n")
			os.Exit(1)
		}
		wx, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid wx: %s\n", os.Args[2])
			os.Exit(1)
		}
		wy, err := strconv.Atoi(os.Args[3])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid wy: %s\n", os.Args[3])
			os.Exit(1)
		}
		cmdCell(int32(wx), int32(wy))
	case "chunks":
		samples := 8
		if len(os.Args) >= 3 {
			n, err := strconv.Atoi(os.Args[2])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid sample count: %s\n", os.Args[2])
				os.Exit(1)
			}
			samples = n
		}
		cmdChunks(samples)
	case "biome-grid":
		jsonOut, _ := extractJSONFlag(os.Args[2:])
		cmdBiomeGrid(jsonOut)
	case "biome-chunk":
		jsonOut, rest := extractJSONFlag(os.Args[2:])
		if len(rest) < 2 {
			fmt.Fprintf(os.Stderr, "Usage: noitrainer-cli biome-chunk [--json] <cx> <cy>\n")
			os.Exit(1)
		}
		cx, err := strconv.Atoi(rest[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid cx: %s\n", rest[0])
			os.Exit(1)
		}
		cy, err := strconv.Atoi(rest[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid cy: %s\n", rest[1])
			os.Exit(1)
		}
		cmdBiomeChunk(int32(cx), int32(cy), jsonOut)
	case "biome-at":
		jsonOut, rest := extractJSONFlag(os.Args[2:])
		if len(rest) < 2 {
			fmt.Fprintf(os.Stderr, "Usage: noitrainer-cli biome-at [--json] <wx> <wy>\n")
			os.Exit(1)
		}
		wx, err := strconv.Atoi(rest[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid wx: %s\n", rest[0])
			os.Exit(1)
		}
		wy, err := strconv.Atoi(rest[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid wy: %s\n", rest[1])
			os.Exit(1)
		}
		cmdBiomeAt(int32(wx), int32(wy), jsonOut)
	case "biome-dump":
		filter := ""
		if len(os.Args) >= 3 {
			filter = os.Args[2]
		}
		cmdBiomeDump(filter)
	case "biome-flags":
		cmdBiomeFlags()
	case "biome-at-many":
		cmdBiomeAtMany()
	case "pixel-scenes":
		cmdPixelScenes()
	case "ngplus":
		cmdNgPlus()
	case "peek":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Usage: noitrainer-cli peek <addr> [size]\n")
			os.Exit(1)
		}
		addr, err := parseAddr(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid address %q: %v\n", os.Args[2], err)
			os.Exit(1)
		}
		size := 128
		if len(os.Args) >= 4 {
			n, err := strconv.Atoi(os.Args[3])
			if err != nil || n <= 0 {
				fmt.Fprintf(os.Stderr, "Invalid size: %s\n", os.Args[3])
				os.Exit(1)
			}
			size = n
		}
		cmdPeek(addr, size)
	case "deref":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Usage: noitrainer-cli deref <addr> [size]\n")
			os.Exit(1)
		}
		addr, err := parseAddr(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid address %q: %v\n", os.Args[2], err)
			os.Exit(1)
		}
		size := 128
		if len(os.Args) >= 4 {
			n, err := strconv.Atoi(os.Args[3])
			if err != nil || n <= 0 {
				fmt.Fprintf(os.Stderr, "Invalid size: %s\n", os.Args[3])
				os.Exit(1)
			}
			size = n
		}
		cmdDeref(addr, size)
	case "read":
		if len(os.Args) < 4 {
			fmt.Fprintf(os.Stderr, "Usage: noitrainer-cli read <type> <addr>\n")
			os.Exit(1)
		}
		typ := os.Args[2]
		addr, err := parseAddr(os.Args[3])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid address %q: %v\n", os.Args[3], err)
			os.Exit(1)
		}
		cmdRead(typ, addr)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		usage()
	}
}

func connect() (*noita.Reader, *process.Process) {
	fmt.Fprintf(os.Stderr, "Looking for noita.exe...\n")
	procs, err := process.FromName("noita.exe")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	proc := procs[0]
	fmt.Fprintf(os.Stderr, "Found noita.exe (PID %d)\n", proc.PID)
	return noita.NewReader(proc), proc
}

// buildBufferNameMap returns a map of type ID -> component name from the buffer registry.
func buildBufferNameMap(reader *noita.Reader) map[noita.TypeID]string {
	names := make(map[noita.TypeID]string)
	for _, b := range reader.ReadComponentBuffers() {
		if b.Name != "" {
			names[noita.TypeID(b.TypeIndex)] = b.Name
		}
	}
	return names
}

// findEntityByID scans entities to find one with a given EntityId.
func findEntityByID(reader *noita.Reader, id int32) *noita.EntitySummary {
	entities := reader.ReadEntityList()
	for _, e := range entities {
		if e.Entity.EntityId == id {
			return e
		}
	}
	return nil
}

// ── watch ──────────────────────────────────────────────────────────

func cmdWatch() {
	reader, _ := connect()

	state := reader.ReadState()
	if !state.Connected {
		fmt.Fprintf(os.Stderr, "Failed to connect: %s\n", state.Error)
		os.Exit(1)
	}

	prev := flattenState(state, reader)
	fmt.Println("=== Initial State ===")
	keys := sortedKeys(prev)
	for _, k := range keys {
		fmt.Printf("  %-40s %s\n", k, prev[k])
	}
	fmt.Printf("\n(%d values)\n", len(prev))
	fmt.Println("\n=== Watching for changes (1s interval) ===")

	for {
		time.Sleep(1 * time.Second)
		state = reader.ReadState()
		if !state.Connected {
			fmt.Println("[disconnected]")
			continue
		}
		curr := flattenState(state, reader)
		changes := diff(prev, curr)
		if len(changes) > 0 {
			ts := time.Now().Format("15:04:05")
			fmt.Printf("\n[%s] %d changed:\n", ts, len(changes))
			for _, k := range sortedKeys(changes) {
				fmt.Printf("  %-40s %s -> %s\n", k, prev[k], changes[k])
			}
		}
		prev = curr
	}
}

// ── entities ───────────────────────────────────────────────────────

func cmdEntities() {
	reader, _ := connect()
	nameMap := buildBufferNameMap(reader)
	entities := reader.ReadEntityList()

	fmt.Printf("Found %d entities\n\n", len(entities))
	fmt.Printf("%-8s %-30s %-20s %s\n", "ID", "Name", "Position", "Components")
	fmt.Printf("%-8s %-30s %-20s %s\n", "──", "────", "────────", "──────────")

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

		fmt.Printf("%-8d %-30s %-20s %s\n",
			e.Entity.EntityId,
			truncate(name, 29),
			pos,
			strings.Join(compNames, ", "),
		)
	}
}

// cmdEntitiesDump emits NDJSON (one line per entity) suitable for diffing
// against a telescope-side prediction. Ground-truth counterpart to
// `scripts/dump.mjs entities` in the noita-telescope repo.
//
// Output schema:
//
//	{"entityId":12345,"name":"worm","x":-1024,"y":512,"components":["VelocityComponent",...]}
func cmdEntitiesDump() {
	reader, _ := connect()
	nameMap := buildBufferNameMap(reader)
	entities := reader.ReadEntityList()

	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	count := 0
	for _, e := range entities {
		compNames := make([]string, 0, len(e.ComponentIDs))
		for _, cid := range e.ComponentIDs {
			if n, ok := nameMap[cid]; ok {
				compNames = append(compNames, n)
			} else {
				compNames = append(compNames, fmt.Sprintf("type_%d", cid))
			}
		}
		row := struct {
			EntityId   int32    `json:"entityId"`
			Name       string   `json:"name"`
			X          float32  `json:"x"`
			Y          float32  `json:"y"`
			Ptr        string   `json:"ptr"`
			Components []string `json:"components"`
		}{
			EntityId:   e.Entity.EntityId,
			Name:       e.Name,
			X:          e.Entity.PosX,
			Y:          e.Entity.PosY,
			Ptr:        fmt.Sprintf("0x%08x", e.Ptr),
			Components: compNames,
		}
		_ = enc.Encode(&row)
		count++
	}
	fmt.Fprintf(os.Stderr, "%d entities emitted\n", count)
}

// ── entity ─────────────────────────────────────────────────────────

func cmdEntity(id int32) {
	reader, _ := connect()
	nameMap := buildBufferNameMap(reader)

	e := findEntityByID(reader, id)
	if e == nil {
		fmt.Fprintf(os.Stderr, "Entity %d not found\n", id)
		os.Exit(1)
	}

	details := reader.ReadEntityDetails(e.Ptr)
	if details == nil {
		fmt.Fprintf(os.Stderr, "Failed to read entity details\n")
		os.Exit(1)
	}

	fmt.Printf("=== Entity %d ===\n", id)
	fmt.Printf("  Name:         %s\n", details.Name)
	fmt.Printf("  Ptr:          0x%08X\n", e.Ptr)
	fmt.Printf("  SlotIndex:    %d\n", details.Entity.SlotIndex)
	fmt.Printf("  Position:     %.2f, %.2f\n", details.Entity.PosX, details.Entity.PosY)
	fmt.Printf("  Scale:        %.2f, %.2f\n", details.Entity.ScaleX, details.Entity.ScaleY)
	fmt.Printf("  Rotation:     cos=%.3f sin=%.3f\n", details.Entity.RotCos, details.Entity.RotSin)
	fmt.Printf("  Flags:        0x%08X\n", details.Entity.Flags10)
	fmt.Printf("  PendingKill:  %d\n", details.Entity.PendingKill)
	fmt.Printf("  ParentPtr:    0x%08X\n", details.Entity.ParentEntityPtr)
	fmt.Printf("  ChildrenPtr:  0x%08X\n", details.Entity.ChildrenPtr)

	// Tag bitset
	var setBits []int
	for i, b := range details.Entity.TagBitset {
		for bit := 0; bit < 8; bit++ {
			if b&(1<<bit) != 0 {
				setBits = append(setBits, i*8+bit)
			}
		}
	}
	if len(setBits) > 0 {
		fmt.Printf("  TagBits:      %v\n", setBits)
	}

	// Component list
	fmt.Printf("\n  Components (%d):\n", len(e.ComponentIDs))
	for _, cid := range e.ComponentIDs {
		name := fmt.Sprintf("type_%d", cid)
		if n, ok := nameMap[cid]; ok {
			name = n
		}
		fmt.Printf("    [%3d] %s\n", cid, name)
	}

	// Known component details
	if details.HP != nil {
		fmt.Printf("\n  DamageModelComponent:\n")
		fmt.Printf("    HP:         %.0f / %.0f (cap %.0f)\n", details.HP.Hp*25, details.HP.MaxHp*25, details.HP.MaxHpCap*25)
		fmt.Printf("    I-Frames:   %d\n", details.HP.InvincibilityFrames)
		printDmgMults(details.HP)
	}

	if details.Char != nil {
		fmt.Printf("\n  CharacterDataComponent:\n")
		fmt.Printf("    Velocity:   %.1f, %.1f\n", details.Char.VelocityX, details.Char.VelocityY)
		fmt.Printf("    On Ground:  %v\n", details.Char.IsOnGround)
		fmt.Printf("    Gravity:    %.2f\n", details.Char.Gravity)
		fmt.Printf("    Fly Time:   %.1f\n", details.Char.FlyTimeMax)
	}

	if details.Wallet != nil {
		fmt.Printf("\n  WalletComponent:\n")
		fmt.Printf("    Gold:       %d (spent %d)\n", details.Wallet.Money, details.Wallet.MoneySpent)
	}

	if details.Inv != nil {
		fmt.Printf("\n  Inventory2Component:\n")
		fmt.Printf("    Wand Slots: %d\n", details.Inv.QuickInventorySlots)
		fmt.Printf("    Active:     %d\n", details.Inv.ActiveItem)
	}

	if details.Ability != nil {
		a := details.Ability
		fmt.Printf("\n  AbilityComponent:\n")
		fmt.Printf("    UiName:     %s\n", a.UiName.FormatMsvcString(reader.Ctx))
		fmt.Printf("    EntityFile: %s\n", a.EntityFile.FormatMsvcString(reader.Ctx))
		fmt.Printf("    SpriteFile: %s\n", a.SpriteFile.FormatMsvcString(reader.Ctx))
		fmt.Printf("    Mana:       %.0f / %.0f (regen %.0f/s)\n", a.Mana, a.ManaMax, a.ManaChargeSpeed*60)
		fmt.Printf("    UseGun:     %v\n", a.UseGunScript)
		gc := a.GunConfig
		fmt.Printf("    Actions:    %d  Deck: %d  Shuffle: %v  Reload: %d\n",
			gc.ActionsPerRound, gc.DeckCapacity, gc.ShuffleDeckWhenEmpty, gc.ReloadTime)
	}

	if len(details.Children) > 0 {
		fmt.Printf("\n  Children (%d):\n", len(details.Children))
		for _, child := range details.Children {
			name := child.Name
			if name == "" {
				name = "(unnamed)"
			}
			fmt.Printf("    [%d] %s @ 0x%08X\n", child.Entity.EntityId, name, child.Ptr)
		}
	}
}

func printDmgMults(d *noita.DamageModelComponent) {
	mults := []struct {
		name string
		val  float32
	}{
		{"Melee", d.DamageMultipliersMelee}, {"Projectile", d.DamageMultipliersProjectile},
		{"Explosion", d.DamageMultipliersExplosion}, {"Electricity", d.DamageMultipliersElectricity},
		{"Fire", d.DamageMultipliersFire}, {"Drill", d.DamageMultipliersDrill},
		{"Slice", d.DamageMultipliersSlice}, {"Ice", d.DamageMultipliersIce},
		{"Healing", d.DamageMultipliersHealing}, {"Physics", d.DamageMultipliersPhysicsHit},
		{"Radioactive", d.DamageMultipliersRadioactive}, {"Poison", d.DamageMultipliersPoison},
		{"Holy", d.DamageMultipliersHoly}, {"Curse", d.DamageMultipliersCurse},
		{"Overeating", d.DamageMultipliersOvereating},
	}
	var nonDefault []string
	for _, m := range mults {
		if m.val != 1.0 {
			nonDefault = append(nonDefault, fmt.Sprintf("%s:%.2f", m.name, m.val))
		}
	}
	if len(nonDefault) > 0 {
		fmt.Printf("    Dmg Mults:  %s\n", strings.Join(nonDefault, " "))
	}
}

// ── tree ───────────────────────────────────────────────────────────

func cmdTree(id int32) {
	reader, _ := connect()

	e := findEntityByID(reader, id)
	if e == nil {
		fmt.Fprintf(os.Stderr, "Entity %d not found\n", id)
		os.Exit(1)
	}

	// Walk up to root
	var chain []*noita.EntitySummary
	chain = append(chain, e)
	current := e
	for current.Entity.ParentEntityPtr != 0 {
		parent, _ := noita.ReadEntity(reader.Ctx, uintptr(current.Entity.ParentEntityPtr))
		if parent == nil {
			break
		}
		parentName := parent.Name.FormatMsvcString(reader.Ctx)
		parentSummary := &noita.EntitySummary{
			Entity: parent,
			Name:   parentName,
			Ptr:    current.Entity.ParentEntityPtr,
		}
		chain = append([]*noita.EntitySummary{parentSummary}, chain...)
		current = parentSummary
	}

	// Print the chain up to our entity
	for i, node := range chain {
		indent := strings.Repeat("  ", i)
		name := node.Name
		if name == "" {
			name = "(unnamed)"
		}
		marker := ""
		if node.Entity.EntityId == id {
			marker = " <<<"
		}
		fmt.Printf("%s[%d] %s (slot=%d pos=%.0f,%.0f)%s\n",
			indent, node.Entity.EntityId, name,
			node.Entity.SlotIndex, node.Entity.PosX, node.Entity.PosY, marker)
	}

	// Print children tree from our entity
	printChildTree(reader, e, len(chain)-1, 2)
}

func printChildTree(reader *noita.Reader, parent *noita.EntitySummary, depth int, maxDepth int) {
	if depth > 10 {
		return
	}
	details := reader.ReadEntityDetails(parent.Ptr)
	if details == nil {
		return
	}
	for _, child := range details.Children {
		indent := strings.Repeat("  ", depth+1)
		name := child.Name
		if name == "" {
			name = "(unnamed)"
		}
		fmt.Printf("%s[%d] %s (slot=%d pos=%.0f,%.0f)\n",
			indent, child.Entity.EntityId, name,
			child.Entity.SlotIndex, child.Entity.PosX, child.Entity.PosY)
		if depth < maxDepth {
			printChildTree(reader, child, depth+1, maxDepth)
		}
	}
}

// ── buffers ────────────────────────────────────────────────────────

func cmdBuffers() {
	reader, _ := connect()
	buffers := reader.ReadComponentBuffers()

	fmt.Printf("Found %d component buffers\n\n", len(buffers))
	fmt.Printf("%-6s %-45s %-10s %-10s %s\n", "TypeID", "Name", "Active", "Capacity", "Ptr")
	fmt.Printf("%-6s %-45s %-10s %-10s %s\n", "──────", "────", "──────", "────────", "───")

	for _, b := range buffers {
		name := b.Name
		if name == "" {
			name = "(unnamed)"
		}
		fmt.Printf("%-6d %-45s %-10d %-10d 0x%08X\n",
			b.TypeIndex, truncate(name, 44), b.ActiveCount, b.Capacity, b.Ptr)
	}
}

// ── dump ───────────────────────────────────────────────────────────

func cmdDump(entityID int32, typeID noita.TypeID, size int) {
	reader, _ := connect()

	e := findEntityByID(reader, entityID)
	if e == nil {
		fmt.Fprintf(os.Stderr, "Entity %d not found\n", entityID)
		os.Exit(1)
	}

	em, _ := reader.ReadEntityManagerPtr()
	if em == nil {
		fmt.Fprintf(os.Stderr, "Failed to read EntityManager\n")
		os.Exit(1)
	}

	nameMap := buildBufferNameMap(reader)
	compName := fmt.Sprintf("type_%d", typeID)
	if n, ok := nameMap[typeID]; ok {
		compName = n
	}

	compPtr, data := reader.ReadRawComponent(em, e.Entity.SlotIndex, typeID, size)
	if data == nil {
		fmt.Fprintf(os.Stderr, "Entity %d has no component of type %d (%s)\n", entityID, typeID, compName)
		os.Exit(1)
	}

	fmt.Printf("Entity %d, %s (type %d) @ 0x%08X, %d bytes:\n\n", entityID, compName, typeID, compPtr, size)
	fmt.Print(hex.Dump(data))
}

// ── components ─────────────────────────────────────────────────────

func cmdComponents(id int32) {
	reader, _ := connect()
	nameMap := buildBufferNameMap(reader)

	e := findEntityByID(reader, id)
	if e == nil {
		fmt.Fprintf(os.Stderr, "Entity %d not found\n", id)
		os.Exit(1)
	}

	em, _ := reader.ReadEntityManagerPtr()
	if em == nil {
		fmt.Fprintf(os.Stderr, "Failed to read EntityManager\n")
		os.Exit(1)
	}

	compIDs := reader.FindEntityComponentIDs(em, e.Entity.SlotIndex)
	name := e.Name
	if name == "" {
		name = "(unnamed)"
	}
	fmt.Printf("Entity %d (%s) has %d component types:\n\n", id, name, len(compIDs))
	fmt.Printf("%-6s %-45s %s\n", "TypeID", "Name", "Ptr")
	fmt.Printf("%-6s %-45s %s\n", "──────", "────", "───")

	for _, cid := range compIDs {
		compName := fmt.Sprintf("type_%d", cid)
		if n, ok := nameMap[cid]; ok {
			compName = n
		}
		compPtr, _ := reader.ReadRawComponent(em, e.Entity.SlotIndex, cid, 0)
		fmt.Printf("%-6d %-45s 0x%08X\n", cid, compName, compPtr)
	}
}

// ── categorize ─────────────────────────────────────────────────────

func cmdCategorize() {
	reader, _ := connect()
	nameMap := buildBufferNameMap(reader)
	entities := reader.ReadEntityList()

	// Group by name
	byName := make(map[string][]*noita.EntitySummary)
	for _, e := range entities {
		name := e.Name
		if name == "" {
			name = "(unnamed)"
		}
		byName[name] = append(byName[name], e)
	}

	// Group by component signature
	type sigGroup struct {
		sig      string
		names    map[string]int
		count    int
		compIDs  []noita.TypeID
	}
	bySig := make(map[string]*sigGroup)
	for _, e := range entities {
		var sigParts []string
		for _, cid := range e.ComponentIDs {
			sigParts = append(sigParts, fmt.Sprintf("%d", cid))
		}
		sig := strings.Join(sigParts, ",")
		g, ok := bySig[sig]
		if !ok {
			g = &sigGroup{sig: sig, names: make(map[string]int), compIDs: e.ComponentIDs}
			bySig[sig] = g
		}
		g.count++
		name := e.Name
		if name == "" {
			name = "(unnamed)"
		}
		g.names[name]++
	}

	// Print by name
	fmt.Printf("=== Entities by Name (%d unique names, %d total) ===\n\n", len(byName), len(entities))
	nameKeys := sortedKeys2(byName)
	for _, name := range nameKeys {
		group := byName[name]
		if len(group) == 1 {
			e := group[0]
			fmt.Printf("  %-40s (id=%d, pos=%.0f,%.0f, %d components)\n",
				name, e.Entity.EntityId, e.Entity.PosX, e.Entity.PosY, len(e.ComponentIDs))
		} else {
			fmt.Printf("  %-40s (%d instances)\n", name, len(group))
		}
	}

	// Print by component signature
	fmt.Printf("\n=== Entities by Component Signature (%d unique signatures) ===\n\n", len(bySig))

	// Sort signatures by count descending
	type sigEntry struct {
		sig string
		g   *sigGroup
	}
	var sigEntries []sigEntry
	for sig, g := range bySig {
		sigEntries = append(sigEntries, sigEntry{sig, g})
	}
	sort.Slice(sigEntries, func(i, j int) bool {
		return sigEntries[i].g.count > sigEntries[j].g.count
	})

	for _, se := range sigEntries {
		g := se.g

		// Format component names
		var compNames []string
		for _, cid := range g.compIDs {
			if n, ok := nameMap[cid]; ok {
				compNames = append(compNames, n)
			} else {
				compNames = append(compNames, fmt.Sprintf("type_%d", cid))
			}
		}

		// Top entity names in this group
		var topNames []string
		for name, count := range g.names {
			if count > 1 {
				topNames = append(topNames, fmt.Sprintf("%s(%d)", name, count))
			} else {
				topNames = append(topNames, name)
			}
		}
		sort.Strings(topNames)
		nameList := strings.Join(topNames, ", ")
		if len(nameList) > 80 {
			nameList = nameList[:77] + "..."
		}

		fmt.Printf("  [%d entities] %s\n", g.count, nameList)
		fmt.Printf("    Components: %s\n\n", strings.Join(compNames, ", "))
	}
}

// ── materials ──────────────────────────────────────────────────────

func cmdMaterials(filter string) {
	reader, _ := connect()
	mats := reader.ReadMaterials()
	if len(mats) == 0 {
		fmt.Fprintf(os.Stderr, "No materials found (CellFactory unavailable?)\n")
		os.Exit(1)
	}
	filter = strings.ToLower(filter)
	fmt.Printf("%-5s %-32s %-10s %-11s %s\n", "ID", "Name", "Fallback", "Texture", "CellData")
	fmt.Printf("%-5s %-32s %-10s %-11s %s\n", "──", "────", "────────", "───────", "────────")
	shown := 0
	for _, m := range mats {
		if filter != "" && !strings.Contains(strings.ToLower(m.Name), filter) {
			continue
		}
		tex := "—"
		if m.TexturePtr != 0 {
			tex = fmt.Sprintf("%dx%d", m.TexW, m.TexH)
		}
		name := m.Name
		if name == "" {
			name = "(unnamed)"
		}
		fmt.Printf("%-5d %-32s 0x%08X %-11s 0x%08X\n",
			m.ID, truncate(name, 31), m.FallbackColor, tex, uint32(m.Addr))
		shown++
	}
	fmt.Printf("\n%d / %d materials shown\n", shown, len(mats))
}

func cmdMaterial(id int) {
	reader, _ := connect()
	mats := reader.ReadMaterials()
	var mat *noita.MaterialInfo
	for _, m := range mats {
		if m.ID == id {
			mat = m
			break
		}
	}
	if mat == nil {
		fmt.Fprintf(os.Stderr, "Material %d not found (have %d materials)\n", id, len(mats))
		os.Exit(1)
	}
	name := mat.Name
	if name == "" {
		name = "(unnamed)"
	}
	fmt.Printf("=== Material %d: %s ===\n", mat.ID, name)
	fmt.Printf("  CellData @     0x%08X\n", uint32(mat.Addr))
	fmt.Printf("  FallbackColor  0x%08X (A=%d R=%d G=%d B=%d)\n",
		mat.FallbackColor,
		(mat.FallbackColor>>24)&0xFF, (mat.FallbackColor>>16)&0xFF,
		(mat.FallbackColor>>8)&0xFF, mat.FallbackColor&0xFF)
	if mat.TexturePtr == 0 {
		fmt.Printf("  Texture:       (none)\n")
		return
	}
	fmt.Printf("  Texture @      0x%08X\n", mat.TexturePtr)
	fmt.Printf("    Size:        %dx%d (%d BGRA pixels)\n",
		mat.TexW, mat.TexH, int64(mat.TexW)*int64(mat.TexH))
	fmt.Printf("    PixelData @  0x%08X\n", mat.PixelDataPtr)
	if mat.PixelDataPtr != 0 && mat.TexW > 0 && mat.TexH > 0 {
		// Print first pixel (BGRA order in memory).
		var px [4]byte
		if _, err := reader.Ctx.ReadAt(px[:], int64(mat.PixelDataPtr)); err == nil {
			fmt.Printf("    pixel[0,0]:  B=%d G=%d R=%d A=%d\n", px[0], px[1], px[2], px[3])
		}
	}
}

// ── chunks / cells ─────────────────────────────────────────────────

func cmdCell(wx, wy int32) {
	reader, _ := connect()
	info := reader.ReadCellAt(wx, wy)
	if info == nil {
		fmt.Fprintf(os.Stderr, "ChunkSystem unavailable\n")
		os.Exit(1)
	}
	fmt.Printf("=== Cell lookup at (%d, %d) ===\n", wx, wy)
	fmt.Printf("  Chunk coord:   (%d, %d)  table idx %d / 0x%X\n",
		info.ChunkCX, info.ChunkCY, info.ChunkIdx, info.ChunkIdx)
	if info.ChunkPtr == 0 {
		fmt.Printf("  Chunk:         (unloaded — air)\n")
		return
	}
	fmt.Printf("  Chunk @        0x%08X\n", info.ChunkPtr)
	if info.CellSlotsPtr == 0 {
		fmt.Printf("  CellSlots:     (none)\n")
		return
	}
	fmt.Printf("  CellSlots @    0x%08X\n", info.CellSlotsPtr)
	fmt.Printf("  Cell idx:      %d  (x%%512=%d, y%%512=%d)\n",
		info.CellIdx, uint32(wx&0x1FF), uint32(wy&0x1FF))
	if info.CellPtr == 0 {
		fmt.Printf("  Cell:          0 (air)\n")
		return
	}
	fmt.Printf("  Cell @         0x%08X\n", info.CellPtr)
	// Dump the first 0x40 bytes of the Cell for inspection.
	buf := make([]byte, 0x40)
	if _, err := reader.Ctx.ReadAt(buf, int64(info.CellPtr)); err == nil {
		fmt.Printf("\n  First 0x40 bytes:\n%s", indentHexDump(buf, info.CellPtr))
	}
}

func cmdChunks(sampleLimit int) {
	reader, _ := connect()
	stats := reader.ReadChunkStats(sampleLimit)
	if stats == nil {
		fmt.Fprintf(os.Stderr, "ChunkSystem unavailable\n")
		os.Exit(1)
	}
	fmt.Printf("=== ChunkSystem ===\n")
	fmt.Printf("  chunk_table @   0x%08X (%d entries)\n", stats.ChunkTablePtr, stats.TableEntries)
	fmt.Printf("  loaded chunks:  %d\n", stats.Loaded)
	if stats.Loaded > 0 {
		fmt.Printf("  loaded coord range: cx [%d..%d] cy [%d..%d]\n",
			stats.MinCX, stats.MaxCX, stats.MinCY, stats.MaxCY)
	}
	if len(stats.Samples) > 0 {
		fmt.Printf("\n  Samples (first %d):\n", len(stats.Samples))
		for _, s := range stats.Samples {
			fmt.Printf("    cx=%-3d cy=%-3d  Chunk* 0x%08X\n", s.CX, s.CY, s.ChunkPtr)
		}
	}
}

// ── biome grid ────────────────────────────────────────────────────

// extractJSONFlag pulls a `--json` (or `-j`) flag out of an args slice and
// returns the bool + remaining args (preserving order).
func extractJSONFlag(args []string) (bool, []string) {
	jsonOut := false
	rest := make([]string, 0, len(args))
	for _, a := range args {
		if a == "--json" || a == "-j" {
			jsonOut = true
			continue
		}
		rest = append(rest, a)
	}
	return jsonOut, rest
}

func writeJSON(v interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		fmt.Fprintf(os.Stderr, "json encode: %v\n", err)
		os.Exit(1)
	}
}

func cmdBiomeGrid(jsonOut bool) {
	reader, _ := connect()
	g := reader.ReadBiomeGridInfo()
	if g == nil {
		fmt.Fprintf(os.Stderr, "Biome grid unavailable (WorldManager.pBackgroundGrid is null)\n")
		os.Exit(1)
	}
	if jsonOut {
		loaded := 0
		reader.IterateBiomeChunks(func(*noita.BiomeChunkInfo) bool {
			loaded++
			return true
		})
		writeJSON(struct {
			*noita.BiomeGridInfo
			Loaded int `json:"loaded"`
		}{g, loaded})
		return
	}
	fmt.Printf("=== BiomeGrid @ 0x%08X ===\n", g.Ptr)
	fmt.Printf("  width:    %d  height:  %d  total:  %d\n", g.Width, g.Height, g.TotalCount)
	fmt.Printf("  x_shift:  %g  y_shift:  %g\n", g.XShift, g.YShift)
	fmt.Printf("  chunks_ptr: 0x%08X\n", g.ChunksPtr)

	loaded := 0
	reader.IterateBiomeChunks(func(*noita.BiomeChunkInfo) bool {
		loaded++
		return true
	})
	fmt.Printf("  loaded chunks: %d\n", loaded)
}

func cmdBiomeChunk(cx, cy int32, jsonOut bool) {
	reader, _ := connect()
	c := reader.ReadBiomeChunkInfo(cx, cy)
	if jsonOut {
		writeJSON(c) // nil prints as "null"
		return
	}
	if c == nil {
		fmt.Printf("(no biome chunk at cx=%d cy=%d)\n", cx, cy)
		return
	}
	printBiomeChunk("", c)
}

func cmdBiomeAt(wx, wy int32, jsonOut bool) {
	reader, _ := connect()
	res := reader.ResolveBiomeAt(wx, wy)
	if res == nil {
		fmt.Fprintf(os.Stderr, "Biome grid unavailable\n")
		os.Exit(1)
	}
	if jsonOut {
		writeJSON(res)
		return
	}
	fmt.Printf("=== Biome lookup at (%d, %d) ===\n", res.WX, res.WY)
	fmt.Printf("  grid:        %dx%d  shift=(%g, %g)\n",
		res.GridWidth, res.GridHeight, res.XShift, res.YShift)
	fmt.Printf("  chunk_coord: (%d, %d)  sub=(%d, %d)\n",
		res.OrigCX, res.OrigCY, res.SubX, res.SubY)
	fmt.Printf("\n-- Original chunk --\n")
	printBiomeChunk("  ", res.Original)
	fmt.Printf("\n-- Wobble decision --\n")
	fmt.Printf("  type: %s\n", res.WobbleType)
	if res.NeighborDir != "" {
		fmt.Printf("  neighbor:  %s -> (cx=%d, cy=%d)\n",
			res.NeighborDir, res.NeighborCX, res.NeighborCY)
	}
	if res.Wobbled && res.Original != nil && res.Resolved != nil && res.Resolved.Ptr != res.Original.Ptr {
		fmt.Printf("\n-- Resolved chunk (after wobble) --\n")
		printBiomeChunk("  ", res.Resolved)
	} else {
		fmt.Printf("\n  (resolved == original; no wobble offset computed here)\n")
		fmt.Printf("  (telescope's wobble math is the part to verify against this metadata)\n")
	}
}

// cmdBiomeFlags emits one JSON line per loaded chunk that has a real biome
// name. Output is JSON-lines (NDJSON), so consumers can stream it directly.
//
// Output schema:
//   {"cx":35,"cy":14,"name":"$biome_tower","xmlName":"tower_coalmine",
//    "ptr":"0x17115d20","biomeDataPtr":"0x170a1708",
//    "wobbleEligible":true,"wavyEdge":true,"forceOriginal":false}
//
// xmlName is the biome's XML filename (path + extension stripped): this is
// the per-XML identity and distinguishes biomes that share a translation key
// (e.g. all tower_*.xml show name="$biome_tower"; the_end vs the_sky both
// show name="$biome_boss_victoryroom").
func cmdBiomeFlags() {
	reader, _ := connect()
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	count := 0
	reader.IterateBiomeChunks(func(c *noita.BiomeChunkInfo) bool {
		if c.Name == "_EMPTY_" || c.Name == "???" || c.Name == "" {
			return true
		}
		row := struct {
			CX             int32  `json:"cx"`
			CY             int32  `json:"cy"`
			Name           string `json:"name"`
			XmlName        string `json:"xmlName"`
			Ptr            string `json:"ptr"`
			BiomeDataPtr   string `json:"biomeDataPtr"`
			WobbleEligible bool   `json:"wobbleEligible"`
			WavyEdge       bool   `json:"wavyEdge"`
			ForceOriginal  bool   `json:"forceOriginal"`
		}{
			CX: c.CX, CY: c.CY, Name: c.Name, XmlName: c.XmlName,
			Ptr:            fmt.Sprintf("0x%08x", c.Ptr),
			BiomeDataPtr:   fmt.Sprintf("0x%08x", c.BiomeDataPtr),
			WobbleEligible: c.WobbleEligibe,
			WavyEdge:       c.WavyEdge,
			ForceOriginal:  c.ForceOriginal,
		}
		_ = enc.Encode(&row)
		count++
		return true
	})
	fmt.Fprintf(os.Stderr, "%d chunks emitted\n", count)
}

// cmdBiomeAtMany reads `wx wy` (whitespace- or comma-separated) lines from
// stdin and emits one JSON line per coord. Designed to feed comparison
// scripts that test telescope's wobble math against ground truth at scale.
//
// Output schema (one per line):
//   {"wx":-1024,"wy":1023,"origCX":33,"origCY":15,"subX":0,"subY":511,
//    "wobbleType":"skipped-flags","neighborDir":"bottom","wobbled":false,
//    "original":{"name":"$biome_coalmine_alt", ...},
//    "resolved":{"name":"$biome_coalmine_alt", ...}}
func cmdBiomeAtMany() {
	reader, _ := connect()
	if reader.ReadBiomeGridInfo() == nil {
		fmt.Fprintf(os.Stderr, "Biome grid unavailable\n")
		os.Exit(1)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 0, 1<<16), 1<<20) // tolerate long lines
	count := 0
	bad := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Accept "wx wy", "wx,wy", "wx\twy".
		fields := strings.FieldsFunc(line, func(r rune) bool {
			return r == ' ' || r == '\t' || r == ','
		})
		if len(fields) < 2 {
			bad++
			continue
		}
		wx, err1 := strconv.ParseInt(fields[0], 10, 32)
		wy, err2 := strconv.ParseInt(fields[1], 10, 32)
		if err1 != nil || err2 != nil {
			bad++
			continue
		}
		res := reader.ResolveBiomeAt(int32(wx), int32(wy))
		_ = enc.Encode(res)
		count++
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "stdin: %v\n", err)
	}
	fmt.Fprintf(os.Stderr, "%d coords resolved, %d skipped\n", count, bad)
}

func cmdBiomeDump(filter string) {
	reader, _ := connect()
	g := reader.ReadBiomeGridInfo()
	if g == nil {
		fmt.Fprintf(os.Stderr, "Biome grid unavailable\n")
		os.Exit(1)
	}
	fmt.Printf("# BiomeGrid %dx%d (shift=%g,%g) chunks_ptr=0x%08X\n",
		g.Width, g.Height, g.XShift, g.YShift, g.ChunksPtr)
	fmt.Printf("# %-3s %-3s %-10s %-10s %-32s %s\n",
		"cx", "cy", "ChunkPtr", "BiomeData", "Name", "Flags(eligible/wavy/forced)")
	count := 0
	filterLower := strings.ToLower(filter)
	reader.IterateBiomeChunks(func(c *noita.BiomeChunkInfo) bool {
		if filter != "" && !strings.Contains(strings.ToLower(c.Name), filterLower) {
			return true
		}
		fmt.Printf("  %-3d %-3d 0x%08X 0x%08X %-32s e=%d w=%d f=%d\n",
			c.CX, c.CY, c.Ptr, c.BiomeDataPtr, truncate(c.Name, 31),
			boolToInt(c.WobbleEligibe), boolToInt(c.WavyEdge), boolToInt(c.ForceOriginal))
		count++
		return true
	})
	fmt.Printf("\n%d chunks shown\n", count)
}

func printBiomeChunk(indent string, c *noita.BiomeChunkInfo) {
	if c == nil {
		fmt.Printf("%s(null)\n", indent)
		return
	}
	fmt.Printf("%schunk @ 0x%08X  cx=%d cy=%d\n", indent, c.Ptr, c.CX, c.CY)
	fmt.Printf("%s  name:           %q\n", indent, c.Name)
	fmt.Printf("%s  wobble_eligible:%v  wavy_edge:%v  force_original:%v\n", indent,
		c.WobbleEligibe, c.WavyEdge, c.ForceOriginal)
	fmt.Printf("%s  biome_data_ptr: 0x%08X (non-null gates pixel-scene placement)\n",
		indent, c.BiomeDataPtr)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// cmdPixelScenes emits NDJSON for every entry in the BiomeGrid pixel-scene
// vectors. This is what the running game has actually placed/queued — the
// authoritative answer to "did this scene spawn?" for telescope's
// loadPixelScene gate to be diffed against.
func cmdPixelScenes() {
	reader, _ := connect()
	if reader.ReadBiomeGridInfo() == nil {
		fmt.Fprintf(os.Stderr, "Biome grid unavailable\n")
		os.Exit(1)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	count := 0
	reader.IteratePixelScenes(func(p *noita.PixelSceneInfo) bool {
		_ = enc.Encode(p)
		count++
		return true
	})
	fmt.Fprintf(os.Stderr, "%d pixel scenes emitted\n", count)
}

// ── ngplus ─────────────────────────────────────────────────────────

func cmdNgPlus() {
	reader, _ := connect()
	v, err := noita.ReadGNewGamePlusCount(reader.Ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Read failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("NG+ count: %d\n", v)
}

// ── peek / deref / read ────────────────────────────────────────────

func cmdPeek(addr uint32, size int) {
	reader, _ := connect()
	buf := make([]byte, size)
	n, err := reader.Ctx.ReadAt(buf, int64(addr))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Read at 0x%08X: %v\n", addr, err)
		os.Exit(1)
	}
	fmt.Printf("0x%08X, %d bytes:\n\n%s", addr, n, indentHexDump(buf[:n], addr))
}

func cmdDeref(addr uint32, size int) {
	reader, _ := connect()
	var p [4]byte
	if _, err := reader.Ctx.ReadAt(p[:], int64(addr)); err != nil {
		fmt.Fprintf(os.Stderr, "Read ptr at 0x%08X: %v\n", addr, err)
		os.Exit(1)
	}
	target := binary.LittleEndian.Uint32(p[:])
	fmt.Printf("*(u32*)0x%08X = 0x%08X\n\n", addr, target)
	if target == 0 {
		fmt.Printf("(null pointer)\n")
		return
	}
	buf := make([]byte, size)
	n, err := reader.Ctx.ReadAt(buf, int64(target))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Read at target 0x%08X: %v\n", target, err)
		os.Exit(1)
	}
	fmt.Printf("0x%08X, %d bytes:\n\n%s", target, n, indentHexDump(buf[:n], target))
}

func cmdRead(typ string, addr uint32) {
	reader, _ := connect()
	read := func(n int) []byte {
		buf := make([]byte, n)
		if _, err := reader.Ctx.ReadAt(buf, int64(addr)); err != nil {
			fmt.Fprintf(os.Stderr, "Read at 0x%08X: %v\n", addr, err)
			os.Exit(1)
		}
		return buf
	}
	switch strings.ToLower(typ) {
	case "u8":
		b := read(1)
		fmt.Printf("u8  @ 0x%08X = %d (0x%02X)\n", addr, b[0], b[0])
	case "u16":
		b := read(2)
		v := binary.LittleEndian.Uint16(b)
		fmt.Printf("u16 @ 0x%08X = %d (0x%04X)\n", addr, v, v)
	case "u32", "ptr":
		b := read(4)
		v := binary.LittleEndian.Uint32(b)
		fmt.Printf("u32 @ 0x%08X = %d (0x%08X)\n", addr, v, v)
	case "u64":
		b := read(8)
		v := binary.LittleEndian.Uint64(b)
		fmt.Printf("u64 @ 0x%08X = %d (0x%016X)\n", addr, v, v)
	case "s32":
		b := read(4)
		v := int32(binary.LittleEndian.Uint32(b))
		fmt.Printf("s32 @ 0x%08X = %d\n", addr, v)
	case "f32":
		b := read(4)
		bits := binary.LittleEndian.Uint32(b)
		fmt.Printf("f32 @ 0x%08X = %g (bits 0x%08X)\n", addr, math.Float32frombits(bits), bits)
	case "f64":
		b := read(8)
		bits := binary.LittleEndian.Uint64(b)
		fmt.Printf("f64 @ 0x%08X = %g (bits 0x%016X)\n", addr, math.Float64frombits(bits), bits)
	case "str":
		ms, _ := noita.ReadMsvcString(reader.Ctx, uintptr(addr))
		if ms == nil {
			fmt.Fprintf(os.Stderr, "Read MsvcString at 0x%08X failed\n", addr)
			os.Exit(1)
		}
		fmt.Printf("MsvcString @ 0x%08X: len=%d cap=%d %q\n",
			addr, ms.Length, ms.Capacity, ms.FormatMsvcString(reader.Ctx))
	default:
		fmt.Fprintf(os.Stderr, "Unknown type %q (want u8|u16|u32|u64|s32|f32|f64|str|ptr)\n", typ)
		os.Exit(1)
	}
}

// ── helpers ────────────────────────────────────────────────────────

// parseAddr accepts decimal or 0x-prefixed hex as a 32-bit address.
func parseAddr(s string) (uint32, error) {
	s = strings.TrimSpace(s)
	base := 10
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		s = s[2:]
		base = 16
	}
	v, err := strconv.ParseUint(s, base, 64)
	if err != nil {
		return 0, err
	}
	return uint32(v), nil
}

// indentHexDump reformats hex.Dump output with a 2-space prefix and replaces
// the leading byte-offset with the absolute VA.
func indentHexDump(buf []byte, base uint32) string {
	dump := hex.Dump(buf)
	lines := strings.Split(strings.TrimRight(dump, "\n"), "\n")
	var out strings.Builder
	for _, ln := range lines {
		// hex.Dump prefixes each line with an 8-digit offset + two spaces.
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

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-2] + ".."
}

func sortedKeys2[V any](m map[string][]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// ── watch helpers (original code) ──────────────────────────────────

func flattenState(gs *noita.GameState, reader *noita.Reader) map[string]string {
	m := make(map[string]string)

	m["WorldSeed"] = fmt.Sprintf("%d", gs.WorldSeed)
	m["NgPlusCount"] = fmt.Sprintf("%d", gs.NgPlusCount)
	m["DeathCount"] = fmt.Sprintf("%d", gs.DeathCount)
	m["NumOrbsTotal"] = fmt.Sprintf("%d", gs.NumOrbsTotal)
	m["Camera.X"] = fmt.Sprintf("%.1f", gs.CameraX)
	m["Camera.Y"] = fmt.Sprintf("%.1f", gs.CameraY)
	m["Camera.ViewW"] = fmt.Sprintf("%.0f", gs.ViewW)
	m["Camera.ViewH"] = fmt.Sprintf("%.0f", gs.ViewH)

	if g := gs.Globals; g != nil {
		flattenStruct(m, "Globals", g, reader)
	}
	if ws := gs.WorldState; ws != nil {
		flattenStruct(m, "WorldState", ws, reader)
	}
	if e := gs.PlayerEntity; e != nil {
		flattenStruct(m, "Player", e, reader)
	}
	if d := gs.PlayerHP; d != nil {
		flattenStruct(m, "HP", d, reader)
	}
	if w := gs.PlayerWallet; w != nil {
		flattenStruct(m, "Wallet", w, reader)
	}
	if c := gs.PlayerChar; c != nil {
		flattenStruct(m, "Char", c, reader)
	}
	if inv := gs.PlayerInv; inv != nil {
		flattenStruct(m, "Inv", inv, reader)
	}
	for i, item := range gs.Wands {
		m[fmt.Sprintf("Wand%d.Name", i)] = fmt.Sprintf("%q", item.Name(reader.Ctx))
		flattenStruct(m, fmt.Sprintf("Wand%d", i), item.Ability, reader)
	}
	for i, item := range gs.Items {
		m[fmt.Sprintf("Item%d.Name", i)] = fmt.Sprintf("%q", item.Name(reader.Ctx))
		for _, mat := range item.Contents {
			m[fmt.Sprintf("Item%d.%s", i, mat.Name)] = fmt.Sprintf("%.0f", mat.Amount)
		}
	}

	return m
}

func flattenStruct(m map[string]string, prefix string, v any, reader *noita.Reader) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return
		}
		rv = rv.Elem()
	}
	rt := rv.Type()

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		fv := rv.Field(i)
		key := prefix + "." + field.Name

		switch fv.Kind() {
		case reflect.Struct:
			if field.Type == reflect.TypeOf(noita.MsvcString{}) {
				ms := fv.Addr().Interface().(*noita.MsvcString)
				m[key] = fmt.Sprintf("%q", ms.FormatMsvcString(reader.Ctx))
			} else if field.Type == reflect.TypeOf(noita.ComponentHeader{}) {
				hdr := fv.Interface().(noita.ComponentHeader)
				m[key+".TypeId"] = fmt.Sprintf("%d", hdr.TypeId)
				m[key+".Active"] = fmt.Sprintf("%v", hdr.Active)
			} else {
				flattenStruct(m, key, fv.Addr().Interface(), reader)
			}
		case reflect.Array:
			if field.Type.Elem().Kind() == reflect.Uint8 && field.Type.Len() > 16 {
				m[key] = fmt.Sprintf("[%d bytes]", field.Type.Len())
			} else {
				m[key] = fmt.Sprintf("%v", fv.Interface())
			}
		case reflect.Bool:
			m[key] = fmt.Sprintf("%v", fv.Bool())
		case reflect.Float32, reflect.Float64:
			m[key] = fmt.Sprintf("%g", fv.Float())
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			m[key] = fmt.Sprintf("%d", fv.Int())
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			val := fv.Uint()
			name := strings.ToLower(field.Name)
			if val > 0xFFFF && (strings.Contains(name, "ptr") || strings.HasPrefix(name, "p") || strings.Contains(name, "vtable")) {
				m[key] = fmt.Sprintf("0x%08X", val)
			} else {
				m[key] = fmt.Sprintf("%d", val)
			}
		default:
			m[key] = fmt.Sprintf("%v", fv.Interface())
		}
	}
}

func diff(prev, curr map[string]string) map[string]string {
	changes := make(map[string]string)
	for k, v := range curr {
		if pv, ok := prev[k]; !ok || pv != v {
			changes[k] = v
		}
	}
	for k := range prev {
		if _, ok := curr[k]; !ok {
			changes[k] = "<removed>"
		}
	}
	return changes
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
