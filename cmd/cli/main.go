package main

import (
	"encoding/hex"
	"fmt"
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
  watch              Watch game state changes (original behavior)
  entities           List all entities with component info
  entity <id>        Show detailed info for entity by ID
  tree <id>          Show entity parent/child tree
  buffers            List all component buffers (type registry)
  dump <id> <type>   Hex dump raw component bytes for entity ID + type ID
  components <id>    List all component types on an entity
  categorize         Categorize entities by name and component signature
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
		fmt.Printf("    Fly Time:   %.1f\n", details.Char.FlyTime)
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
		fmt.Printf("    UiName:     %s\n", noita.MsvcStringValue(&a.UiName, reader.Ctx))
		fmt.Printf("    EntityFile: %s\n", noita.MsvcStringValue(&a.EntityFile, reader.Ctx))
		fmt.Printf("    SpriteFile: %s\n", noita.MsvcStringValue(&a.SpriteFile, reader.Ctx))
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
		{"Melee", d.DmgMultMelee}, {"Projectile", d.DmgMultProjectile},
		{"Explosion", d.DmgMultExplosion}, {"Electricity", d.DmgMultElectricity},
		{"Fire", d.DmgMultFire}, {"Drill", d.DmgMultDrill},
		{"Slice", d.DmgMultSlice}, {"Ice", d.DmgMultIce},
		{"Healing", d.DmgMultHealing}, {"Physics", d.DmgMultPhysicsHit},
		{"Radioactive", d.DmgMultRadioactive}, {"Poison", d.DmgMultPoison},
		{"Holy", d.DmgMultHoly}, {"Curse", d.DmgMultCurse},
		{"Overeating", d.DmgMultOvereating}, {"Material", d.DmgMultMaterial},
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
		parentName := noita.MsvcStringValue(&parent.Name, reader.Ctx)
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

// ── helpers ────────────────────────────────────────────────────────

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
				m[key] = fmt.Sprintf("%q", noita.MsvcStringValue(ms, reader.Ctx))
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
