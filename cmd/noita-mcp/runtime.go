// Runtime tools attach to the running noita.exe process and mirror the
// introspection surface of cmd/cli. Each tool opens a fresh Reader per
// call; there is no shared state between tools.
package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"sort"
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

type entityIDInput struct {
	EntityID int32 `json:"entity_id" jsonschema:"EntityId of a live entity"`
}

type dumpInput struct {
	EntityID int32 `json:"entity_id"`
	TypeID   int   `json:"type_id" jsonschema:"component type index"`
	Size     int   `json:"size,omitempty" jsonschema:"bytes to dump (default 256)"`
}

// sortedKeys2RT mirrors cmd/cli/main.go:sortedKeys2.
func sortedKeys2RT[V any](m map[string][]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// appendDmgMultsRT mirrors cmd/cli/main.go:printDmgMults but writes to b.
func appendDmgMultsRT(b *strings.Builder, d *noita.DamageModelComponent) {
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
		fmt.Fprintf(b, "    Dmg Mults:  %s\n", strings.Join(nonDefault, " "))
	}
}

// appendChildTreeRT mirrors cmd/cli/main.go:printChildTree but writes to b.
func appendChildTreeRT(b *strings.Builder, reader *noita.Reader, parent *noita.EntitySummary, depth, maxDepth int) {
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
		fmt.Fprintf(b, "%s[%d] %s (slot=%d pos=%.0f,%.0f)\n",
			indent, child.Entity.EntityId, name,
			child.Entity.SlotIndex, child.Entity.PosX, child.Entity.PosY)
		if depth < maxDepth {
			appendChildTreeRT(b, reader, child, depth+1, maxDepth)
		}
	}
}

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

func boolToIntRT(b bool) int {
	if b {
		return 1
	}
	return 0
}

type peekInput struct {
	Addr string `json:"addr" jsonschema:"address, decimal or 0x-prefixed hex"`
	Size int    `json:"size,omitempty" jsonschema:"bytes to dump (default 128)"`
}

type cellColorInput struct {
	WX int32 `json:"wx"`
	WY int32 `json:"wy"`
}

type cellgridBlitInput struct {
	WX int32 `json:"wx"`
	WY int32 `json:"wy"`
	W  int32 `json:"w" jsonschema:"width in world pixels (max 1024)"`
	H  int32 `json:"h" jsonschema:"height in world pixels (max 1024)"`
}
type derefInput struct {
	Addr string `json:"addr" jsonschema:"address of a u32 pointer to dereference"`
	Size int    `json:"size,omitempty" jsonschema:"bytes to dump at the pointee (default 128)"`
}
type readMemInput struct {
	Type string `json:"type" jsonschema:"u8|u16|u32|u64|s32|f32|f64|str|ptr"`
	Addr string `json:"addr"`
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
		v, err := noita.ReadGNewGamePlusCount(reader.Ctx)
		if err != nil {
			return toolErr(fmt.Errorf("read ng+ count: %w", err))
		}
		return textResult(fmt.Sprintf("NG+ count: %d\n", v)), nil, nil
	})

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

	mcp.AddTool(s, &mcp.Tool{
		Name:        "entity",
		Description: "Show detailed info for an entity (position, flags, HP, character, wallet, inventory, ability, children).",
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

		details := reader.ReadEntityDetails(e.Ptr)
		if details == nil {
			return toolErr(fmt.Errorf("failed to read entity details"))
		}

		var b strings.Builder
		fmt.Fprintf(&b, "=== Entity %d ===\n", in.EntityID)
		fmt.Fprintf(&b, "  Name:         %s\n", details.Name)
		fmt.Fprintf(&b, "  Ptr:          0x%08X\n", e.Ptr)
		fmt.Fprintf(&b, "  SlotIndex:    %d\n", details.Entity.SlotIndex)
		fmt.Fprintf(&b, "  Position:     %.2f, %.2f\n", details.Entity.PosX, details.Entity.PosY)
		fmt.Fprintf(&b, "  Scale:        %.2f, %.2f\n", details.Entity.ScaleX, details.Entity.ScaleY)
		fmt.Fprintf(&b, "  Rotation:     cos=%.3f sin=%.3f\n", details.Entity.RotCos, details.Entity.RotSin)
		fmt.Fprintf(&b, "  Flags:        0x%08X\n", details.Entity.Flags10)
		fmt.Fprintf(&b, "  PendingKill:  %d\n", details.Entity.PendingKill)
		fmt.Fprintf(&b, "  ParentPtr:    0x%08X\n", details.Entity.ParentEntityPtr)
		fmt.Fprintf(&b, "  ChildrenPtr:  0x%08X\n", details.Entity.ChildrenPtr)

		// Tag bitset
		var setBits []int
		for i, bt := range details.Entity.TagBitset {
			for bit := 0; bit < 8; bit++ {
				if bt&(1<<bit) != 0 {
					setBits = append(setBits, i*8+bit)
				}
			}
		}
		if len(setBits) > 0 {
			fmt.Fprintf(&b, "  TagBits:      %v\n", setBits)
		}

		// Component list
		fmt.Fprintf(&b, "\n  Components (%d):\n", len(e.ComponentIDs))
		for _, cid := range e.ComponentIDs {
			name := fmt.Sprintf("type_%d", cid)
			if n, ok := nameMap[cid]; ok {
				name = n
			}
			fmt.Fprintf(&b, "    [%3d] %s\n", cid, name)
		}

		// Known component details
		if details.HP != nil {
			fmt.Fprintf(&b, "\n  DamageModelComponent:\n")
			fmt.Fprintf(&b, "    HP:         %.0f / %.0f (cap %.0f)\n", details.HP.Hp*25, details.HP.MaxHp*25, details.HP.MaxHpCap*25)
			fmt.Fprintf(&b, "    I-Frames:   %d\n", details.HP.InvincibilityFrames)
			appendDmgMultsRT(&b, details.HP)
		}

		if details.Char != nil {
			fmt.Fprintf(&b, "\n  CharacterDataComponent:\n")
			fmt.Fprintf(&b, "    Velocity:   %.1f, %.1f\n", details.Char.VelocityX, details.Char.VelocityY)
			fmt.Fprintf(&b, "    On Ground:  %v\n", details.Char.IsOnGround)
			fmt.Fprintf(&b, "    Gravity:    %.2f\n", details.Char.Gravity)
			fmt.Fprintf(&b, "    Fly Time:   %.1f\n", details.Char.FlyTimeMax)
		}

		if details.Wallet != nil {
			fmt.Fprintf(&b, "\n  WalletComponent:\n")
			fmt.Fprintf(&b, "    Gold:       %d (spent %d)\n", details.Wallet.Money, details.Wallet.MoneySpent)
		}

		if details.Inv != nil {
			fmt.Fprintf(&b, "\n  Inventory2Component:\n")
			fmt.Fprintf(&b, "    Wand Slots: %d\n", details.Inv.QuickInventorySlots)
			fmt.Fprintf(&b, "    Active:     %d\n", details.Inv.ActiveItem)
		}

		if details.Ability != nil {
			a := details.Ability
			fmt.Fprintf(&b, "\n  AbilityComponent:\n")
			fmt.Fprintf(&b, "    UiName:     %s\n", a.UiName.FormatMsvcString(reader.Ctx))
			fmt.Fprintf(&b, "    EntityFile: %s\n", a.EntityFile.FormatMsvcString(reader.Ctx))
			fmt.Fprintf(&b, "    SpriteFile: %s\n", a.SpriteFile.FormatMsvcString(reader.Ctx))
			fmt.Fprintf(&b, "    Mana:       %.0f / %.0f (regen %.0f/s)\n", a.Mana, a.ManaMax, a.ManaChargeSpeed*60)
			fmt.Fprintf(&b, "    UseGun:     %v\n", a.UseGunScript)
			gc := a.GunConfig
			fmt.Fprintf(&b, "    Actions:    %d  Deck: %d  Shuffle: %v  Reload: %d\n",
				gc.ActionsPerRound, gc.DeckCapacity, gc.ShuffleDeckWhenEmpty, gc.ReloadTime)
		}

		if len(details.Children) > 0 {
			fmt.Fprintf(&b, "\n  Children (%d):\n", len(details.Children))
			for _, child := range details.Children {
				name := child.Name
				if name == "" {
					name = "(unnamed)"
				}
				fmt.Fprintf(&b, "    [%d] %s @ 0x%08X\n", child.Entity.EntityId, name, child.Ptr)
			}
		}

		return textResult(b.String()), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "tree",
		Description: "Show an entity's parent-chain (root-first) then its subtree to depth 2.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in entityIDInput) (*mcp.CallToolResult, any, error) {
		reader, _, err := connectRuntime()
		if err != nil {
			return toolErr(err)
		}

		e := findEntityByIDRT(reader, in.EntityID)
		if e == nil {
			return toolErr(fmt.Errorf("entity %d not found", in.EntityID))
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

		var b strings.Builder
		// Print the chain up to our entity
		for i, node := range chain {
			indent := strings.Repeat("  ", i)
			name := node.Name
			if name == "" {
				name = "(unnamed)"
			}
			marker := ""
			if node.Entity.EntityId == in.EntityID {
				marker = " <<<"
			}
			fmt.Fprintf(&b, "%s[%d] %s (slot=%d pos=%.0f,%.0f)%s\n",
				indent, node.Entity.EntityId, name,
				node.Entity.SlotIndex, node.Entity.PosX, node.Entity.PosY, marker)
		}

		// Print children tree from our entity
		appendChildTreeRT(&b, reader, e, len(chain)-1, 2)

		return textResult(b.String()), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "categorize",
		Description: "Group live entities by name and by component signature; show counts and top members.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in struct{}) (*mcp.CallToolResult, any, error) {
		reader, _, err := connectRuntime()
		if err != nil {
			return toolErr(err)
		}
		nameMap := buildRuntimeBufferNames(reader)
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
			sig     string
			names   map[string]int
			count   int
			compIDs []noita.TypeID
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

		var b strings.Builder

		// Print by name
		fmt.Fprintf(&b, "=== Entities by Name (%d unique names, %d total) ===\n\n", len(byName), len(entities))
		nameKeys := sortedKeys2RT(byName)
		for _, name := range nameKeys {
			group := byName[name]
			if len(group) == 1 {
				e := group[0]
				fmt.Fprintf(&b, "  %-40s (id=%d, pos=%.0f,%.0f, %d components)\n",
					name, e.Entity.EntityId, e.Entity.PosX, e.Entity.PosY, len(e.ComponentIDs))
			} else {
				fmt.Fprintf(&b, "  %-40s (%d instances)\n", name, len(group))
			}
		}

		// Print by component signature
		fmt.Fprintf(&b, "\n=== Entities by Component Signature (%d unique signatures) ===\n\n", len(bySig))

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

			fmt.Fprintf(&b, "  [%d entities] %s\n", g.count, nameList)
			fmt.Fprintf(&b, "    Components: %s\n\n", strings.Join(compNames, ", "))
		}

		return textResult(b.String()), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "materials",
		Description: "List CellFactory materials; optional name substring filter.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in materialsInput) (*mcp.CallToolResult, any, error) {
		reader, _, err := connectRuntime()
		if err != nil {
			return toolErr(err)
		}
		mats := reader.ReadMaterials()
		if len(mats) == 0 {
			return toolErr(fmt.Errorf("no materials found (CellFactory unavailable?)"))
		}
		filter := strings.ToLower(in.Filter)
		var b strings.Builder
		fmt.Fprintf(&b, "%-5s %-32s %-10s %-11s %s\n", "ID", "Name", "Fallback", "Texture", "CellData")
		fmt.Fprintf(&b, "%-5s %-32s %-10s %-11s %s\n", "──", "────", "────────", "───────", "────────")
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
			fmt.Fprintf(&b, "%-5d %-32s 0x%08X %-11s 0x%08X\n",
				m.ID, truncateStr(name, 31), m.FallbackColor, tex, uint32(m.Addr))
			shown++
		}
		fmt.Fprintf(&b, "\n%d / %d materials shown\n", shown, len(mats))
		return textResult(b.String()), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "material",
		Description: "Show full CellData for a material id.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in materialInput) (*mcp.CallToolResult, any, error) {
		reader, _, err := connectRuntime()
		if err != nil {
			return toolErr(err)
		}
		mats := reader.ReadMaterials()
		var mat *noita.MaterialInfo
		for _, m := range mats {
			if m.ID == in.MaterialID {
				mat = m
				break
			}
		}
		if mat == nil {
			return toolErr(fmt.Errorf("material %d not found (have %d materials)", in.MaterialID, len(mats)))
		}
		name := mat.Name
		if name == "" {
			name = "(unnamed)"
		}
		var b strings.Builder
		fmt.Fprintf(&b, "=== Material %d: %s ===\n", mat.ID, name)
		fmt.Fprintf(&b, "  CellData @     0x%08X\n", uint32(mat.Addr))
		fmt.Fprintf(&b, "  FallbackColor  0x%08X (A=%d R=%d G=%d B=%d)\n",
			mat.FallbackColor,
			(mat.FallbackColor>>24)&0xFF, (mat.FallbackColor>>16)&0xFF,
			(mat.FallbackColor>>8)&0xFF, mat.FallbackColor&0xFF)
		if mat.TexturePtr == 0 {
			fmt.Fprintf(&b, "  Texture:       (none)\n")
			return textResult(b.String()), nil, nil
		}
		fmt.Fprintf(&b, "  Texture @      0x%08X\n", mat.TexturePtr)
		fmt.Fprintf(&b, "    Size:        %dx%d (%d BGRA pixels)\n",
			mat.TexW, mat.TexH, int64(mat.TexW)*int64(mat.TexH))
		fmt.Fprintf(&b, "    PixelData @  0x%08X\n", mat.PixelDataPtr)
		if mat.PixelDataPtr != 0 && mat.TexW > 0 && mat.TexH > 0 {
			var px [4]byte
			if _, err := reader.Ctx.ReadAt(px[:], int64(mat.PixelDataPtr)); err == nil {
				fmt.Fprintf(&b, "    pixel[0,0]:  B=%d G=%d R=%d A=%d\n", px[0], px[1], px[2], px[3])
			}
		}
		return textResult(b.String()), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "cell",
		Description: "Resolve a world pixel to its chunk/cell pointers; dump first 0x40 bytes of the cell.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in cellInput) (*mcp.CallToolResult, any, error) {
		reader, _, err := connectRuntime()
		if err != nil {
			return toolErr(err)
		}
		info := reader.ReadCellAt(in.WX, in.WY)
		if info == nil {
			return toolErr(fmt.Errorf("ChunkSystem unavailable"))
		}
		var b strings.Builder
		fmt.Fprintf(&b, "=== Cell lookup at (%d, %d) ===\n", in.WX, in.WY)
		fmt.Fprintf(&b, "  Chunk coord:   (%d, %d)  table idx %d / 0x%X\n",
			info.ChunkCX, info.ChunkCY, info.ChunkIdx, info.ChunkIdx)
		if info.ChunkPtr == 0 {
			fmt.Fprintf(&b, "  Chunk:         (unloaded — air)\n")
			return textResult(b.String()), nil, nil
		}
		fmt.Fprintf(&b, "  Chunk @        0x%08X\n", info.ChunkPtr)
		if info.CellSlotsPtr == 0 {
			fmt.Fprintf(&b, "  CellSlots:     (none)\n")
			return textResult(b.String()), nil, nil
		}
		fmt.Fprintf(&b, "  CellSlots @    0x%08X\n", info.CellSlotsPtr)
		fmt.Fprintf(&b, "  Cell idx:      %d  (x%%512=%d, y%%512=%d)\n",
			info.CellIdx, uint32(in.WX&0x1FF), uint32(in.WY&0x1FF))
		if info.CellPtr == 0 {
			fmt.Fprintf(&b, "  Cell:          0 (air)\n")
			return textResult(b.String()), nil, nil
		}
		fmt.Fprintf(&b, "  Cell @         0x%08X\n", info.CellPtr)
		buf := make([]byte, 0x40)
		if _, err := reader.Ctx.ReadAt(buf, int64(info.CellPtr)); err == nil {
			fmt.Fprintf(&b, "\n  First 0x40 bytes:\n%s", indentHexDumpRT(buf, info.CellPtr))
		}
		return textResult(b.String()), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "chunks",
		Description: "Show ChunkSystem stats (total chunks, coord range) and sample some loaded chunks.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in chunksInput) (*mcp.CallToolResult, any, error) {
		reader, _, err := connectRuntime()
		if err != nil {
			return toolErr(err)
		}
		samples := in.Samples
		if samples <= 0 {
			samples = 8
		}
		stats := reader.ReadChunkStats(samples)
		if stats == nil {
			return toolErr(fmt.Errorf("ChunkSystem unavailable"))
		}
		var b strings.Builder
		fmt.Fprintf(&b, "=== ChunkSystem ===\n")
		fmt.Fprintf(&b, "  chunk_table @   0x%08X (%d entries)\n", stats.ChunkTablePtr, stats.TableEntries)
		fmt.Fprintf(&b, "  loaded chunks:  %d\n", stats.Loaded)
		if stats.Loaded > 0 {
			fmt.Fprintf(&b, "  loaded coord range: cx [%d..%d] cy [%d..%d]\n",
				stats.MinCX, stats.MaxCX, stats.MinCY, stats.MaxCY)
		}
		if len(stats.Samples) > 0 {
			fmt.Fprintf(&b, "\n  Samples (first %d):\n", len(stats.Samples))
			for _, sa := range stats.Samples {
				fmt.Fprintf(&b, "    cx=%-3d cy=%-3d  Chunk* 0x%08X\n", sa.CX, sa.CY, sa.ChunkPtr)
			}
		}
		return textResult(b.String()), nil, nil
	})

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

	mcp.AddTool(s, &mcp.Tool{
		Name:        "biome_chunk",
		Description: "JSON: metadata for the biome chunk at (cx, cy). Returns null if not loaded.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in biomeChunkInput) (*mcp.CallToolResult, any, error) {
		reader, _, err := connectRuntime()
		if err != nil {
			return toolErr(err)
		}
		c := reader.ReadBiomeChunkInfo(in.CX, in.CY)
		data, err := json.Marshal(c)
		if err != nil {
			return toolErr(err)
		}
		return textResult(string(data)), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "biome_at",
		Description: "JSON: resolve a world-pixel coordinate to its biome (original + wobble-resolved).",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in biomeAtInput) (*mcp.CallToolResult, any, error) {
		reader, _, err := connectRuntime()
		if err != nil {
			return toolErr(err)
		}
		if reader.ReadBiomeGridInfo() == nil {
			return toolErr(fmt.Errorf("biome grid unavailable"))
		}
		res := reader.ResolveBiomeAt(in.WX, in.WY)
		data, err := json.Marshal(res)
		if err != nil {
			return toolErr(err)
		}
		return textResult(string(data)), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "biome_dump",
		Description: "List every loaded biome chunk; optional name substring filter.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in biomeDumpInput) (*mcp.CallToolResult, any, error) {
		reader, _, err := connectRuntime()
		if err != nil {
			return toolErr(err)
		}
		g := reader.ReadBiomeGridInfo()
		if g == nil {
			return toolErr(fmt.Errorf("biome grid unavailable"))
		}
		var b strings.Builder
		fmt.Fprintf(&b, "# BiomeGrid %dx%d (shift=%g,%g) chunks_ptr=0x%08X\n",
			g.Width, g.Height, g.XShift, g.YShift, g.ChunksPtr)
		fmt.Fprintf(&b, "# %-3s %-3s %-10s %-10s %-32s %s\n",
			"cx", "cy", "ChunkPtr", "BiomeData", "Name", "Flags(eligible/wavy/forced)")
		count := 0
		filterLower := strings.ToLower(in.Filter)
		reader.IterateBiomeChunks(func(c *noita.BiomeChunkInfo) bool {
			if in.Filter != "" && !strings.Contains(strings.ToLower(c.Name), filterLower) {
				return true
			}
			fmt.Fprintf(&b, "  %-3d %-3d 0x%08X 0x%08X %-32s e=%d w=%d f=%d\n",
				c.CX, c.CY, c.Ptr, c.BiomeDataPtr, truncateStr(c.Name, 31),
				boolToIntRT(c.WobbleEligibe), boolToIntRT(c.WavyEdge), boolToIntRT(c.ForceOriginal))
			count++
			return true
		})
		fmt.Fprintf(&b, "\n%d chunks shown\n", count)
		return textResult(b.String()), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "biome_flags",
		Description: "JSON array: one entry per loaded biome chunk with a real name (cx, cy, name, xmlName, ptrs, wobble flags).",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in struct{}) (*mcp.CallToolResult, any, error) {
		reader, _, err := connectRuntime()
		if err != nil {
			return toolErr(err)
		}
		if reader.ReadBiomeGridInfo() == nil {
			return toolErr(fmt.Errorf("biome grid unavailable"))
		}
		type row struct {
			CX             int32  `json:"cx"`
			CY             int32  `json:"cy"`
			Name           string `json:"name"`
			XmlName        string `json:"xmlName"`
			Ptr            string `json:"ptr"`
			BiomeDataPtr   string `json:"biomeDataPtr"`
			WobbleEligible bool   `json:"wobbleEligible"`
			WavyEdge       bool   `json:"wavyEdge"`
			ForceOriginal  bool   `json:"forceOriginal"`
		}
		out := make([]row, 0)
		reader.IterateBiomeChunks(func(c *noita.BiomeChunkInfo) bool {
			if c.Name == "_EMPTY_" || c.Name == "???" || c.Name == "" {
				return true
			}
			out = append(out, row{
				CX: c.CX, CY: c.CY, Name: c.Name, XmlName: c.XmlName,
				Ptr:            fmt.Sprintf("0x%08x", c.Ptr),
				BiomeDataPtr:   fmt.Sprintf("0x%08x", c.BiomeDataPtr),
				WobbleEligible: c.WobbleEligibe,
				WavyEdge:       c.WavyEdge,
				ForceOriginal:  c.ForceOriginal,
			})
			return true
		})
		data, err := json.Marshal(out)
		if err != nil {
			return toolErr(err)
		}
		return textResult(string(data)), nil, nil
	})

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

	mcp.AddTool(s, &mcp.Tool{
		Name:        "pixel_scenes",
		Description: "JSON array: every BiomeGrid pixel-scene entry currently placed/queued by the running game.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in struct{}) (*mcp.CallToolResult, any, error) {
		reader, _, err := connectRuntime()
		if err != nil {
			return toolErr(err)
		}
		if reader.ReadBiomeGridInfo() == nil {
			return toolErr(fmt.Errorf("biome grid unavailable"))
		}
		out := make([]*noita.PixelSceneInfo, 0)
		reader.IteratePixelScenes(func(p *noita.PixelSceneInfo) bool {
			out = append(out, p)
			return true
		})
		data, err := json.Marshal(out)
		if err != nil {
			return toolErr(err)
		}
		return textResult(string(data)), nil, nil
	})

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

	mcp.AddTool(s, &mcp.Tool{
		Name:        "deref",
		Description: "Read a u32 pointer at addr and hex-dump memory at the pointee. size defaults to 128.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in derefInput) (*mcp.CallToolResult, any, error) {
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
		var p [4]byte
		if _, err := reader.Ctx.ReadAt(p[:], int64(addr)); err != nil {
			return toolErr(fmt.Errorf("read ptr at 0x%08X: %w", addr, err))
		}
		target := binary.LittleEndian.Uint32(p[:])
		var b strings.Builder
		fmt.Fprintf(&b, "*(u32*)0x%08X = 0x%08X\n\n", addr, target)
		if target == 0 {
			b.WriteString("(null pointer)\n")
			return textResult(b.String()), nil, nil
		}
		buf := make([]byte, size)
		n, err := reader.Ctx.ReadAt(buf, int64(target))
		if err != nil {
			return toolErr(fmt.Errorf("read at target 0x%08X: %w", target, err))
		}
		fmt.Fprintf(&b, "0x%08X, %d bytes:\n\n%s", target, n, indentHexDumpRT(buf[:n], target))
		return textResult(b.String()), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "read_memory",
		Description: "Read a typed value at addr. type is one of: u8, u16, u32, u64, s32, f32, f64, str, ptr.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in readMemInput) (*mcp.CallToolResult, any, error) {
		reader, _, err := connectRuntime()
		if err != nil {
			return toolErr(err)
		}
		addr, err := parseRuntimeAddr(in.Addr)
		if err != nil {
			return toolErr(err)
		}
		read := func(n int) ([]byte, error) {
			buf := make([]byte, n)
			if _, err := reader.Ctx.ReadAt(buf, int64(addr)); err != nil {
				return nil, fmt.Errorf("read at 0x%08X: %w", addr, err)
			}
			return buf, nil
		}
		var b strings.Builder
		switch strings.ToLower(in.Type) {
		case "u8":
			buf, err := read(1)
			if err != nil {
				return toolErr(err)
			}
			fmt.Fprintf(&b, "u8  @ 0x%08X = %d (0x%02X)\n", addr, buf[0], buf[0])
		case "u16":
			buf, err := read(2)
			if err != nil {
				return toolErr(err)
			}
			v := binary.LittleEndian.Uint16(buf)
			fmt.Fprintf(&b, "u16 @ 0x%08X = %d (0x%04X)\n", addr, v, v)
		case "u32", "ptr":
			buf, err := read(4)
			if err != nil {
				return toolErr(err)
			}
			v := binary.LittleEndian.Uint32(buf)
			fmt.Fprintf(&b, "u32 @ 0x%08X = %d (0x%08X)\n", addr, v, v)
		case "u64":
			buf, err := read(8)
			if err != nil {
				return toolErr(err)
			}
			v := binary.LittleEndian.Uint64(buf)
			fmt.Fprintf(&b, "u64 @ 0x%08X = %d (0x%016X)\n", addr, v, v)
		case "s32":
			buf, err := read(4)
			if err != nil {
				return toolErr(err)
			}
			v := int32(binary.LittleEndian.Uint32(buf))
			fmt.Fprintf(&b, "s32 @ 0x%08X = %d\n", addr, v)
		case "f32":
			buf, err := read(4)
			if err != nil {
				return toolErr(err)
			}
			bits := binary.LittleEndian.Uint32(buf)
			fmt.Fprintf(&b, "f32 @ 0x%08X = %g (bits 0x%08X)\n", addr, math.Float32frombits(bits), bits)
		case "f64":
			buf, err := read(8)
			if err != nil {
				return toolErr(err)
			}
			bits := binary.LittleEndian.Uint64(buf)
			fmt.Fprintf(&b, "f64 @ 0x%08X = %g (bits 0x%016X)\n", addr, math.Float64frombits(bits), bits)
		case "str":
			ms, _ := noita.ReadMsvcString(reader.Ctx, uintptr(addr))
			if ms == nil {
				return toolErr(fmt.Errorf("read MsvcString at 0x%08X failed", addr))
			}
			fmt.Fprintf(&b, "MsvcString @ 0x%08X: len=%d cap=%d %q\n",
				addr, ms.Length, ms.Capacity, ms.FormatMsvcString(reader.Ctx))
		default:
			return toolErr(fmt.Errorf("unknown type %q (want u8|u16|u32|u64|s32|f32|f64|str|ptr)", in.Type))
		}
		return textResult(b.String()), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name: "cell_color",
		Description: "Return the rendered cell color at world pixel (wx, wy). Reads cell+0x30 — the engine's mColor slot, which is the texture-sampled-and-edge-stamped pixel that ends up in the GPU sprite_cellgrid texture. Memory layout is BGRA8 per the CellTexture convention. Returns JSON with {raw_hex, b, g, r, a, mirror_hex} where mirror_hex is the same field at cell+0x34 (a backup the engine maintains, often equal). Air cells (no live cell at that coord) return {present: false}.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in cellColorInput) (*mcp.CallToolResult, any, error) {
		reader, _, err := connectRuntime()
		if err != nil {
			return toolErr(err)
		}
		info := reader.ReadCellAt(in.WX, in.WY)
		if info == nil {
			return toolErr(fmt.Errorf("ChunkSystem unavailable"))
		}
		if info.CellPtr == 0 {
			data, _ := json.Marshal(map[string]any{
				"wx": in.WX, "wy": in.WY, "present": false, "reason": "air",
			})
			return textResult(string(data)), nil, nil
		}
		var buf [8]byte
		if _, err := reader.Ctx.ReadAt(buf[:], int64(info.CellPtr)+0x30); err != nil {
			return toolErr(fmt.Errorf("read cell color at 0x%08X: %w", info.CellPtr+0x30, err))
		}
		// Engine stores cell pixels BGRA in memory (matches CellTexture in
		// noita.hexpat). Decode both slots so callers can confirm the
		// engine maintains the mirror at +0x34 (often a snapshot for
		// diff/staining vs the live mColor at +0x30).
		decode := func(b []byte) map[string]any {
			c := binary.LittleEndian.Uint32(b)
			return map[string]any{
				"raw_hex": fmt.Sprintf("0x%08X", c),
				"b":       int(b[0]),
				"g":       int(b[1]),
				"r":       int(b[2]),
				"a":       int(b[3]),
			}
		}
		out := map[string]any{
			"wx":         in.WX,
			"wy":         in.WY,
			"present":    true,
			"cell_ptr":   fmt.Sprintf("0x%08X", info.CellPtr),
			"primary":    decode(buf[0:4]),
			"mirror":     decode(buf[4:8]),
		}
		data, _ := json.Marshal(out)
		return textResult(string(data)), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name: "cellgrid_blit",
		Description: "Read a w×h block of the engine's per-cell rendered colors (cell+0x30, BGRA8) and return a base64-encoded PNG of the result. Air cells become transparent. Use to compare the engine's actual on-screen materials/edges against an offline renderer — the data here matches what the engine uploads to the GPU sprite_cellgrid texture (before the post bilinear/dilate shaders). Max 1024×1024.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in cellgridBlitInput) (*mcp.CallToolResult, any, error) {
		if in.W <= 0 || in.H <= 0 {
			return toolErr(fmt.Errorf("w and h must be > 0"))
		}
		if in.W > 1024 || in.H > 1024 {
			return toolErr(fmt.Errorf("w and h must each be ≤ 1024 (got %dx%d)", in.W, in.H))
		}
		reader, _, err := connectRuntime()
		if err != nil {
			return toolErr(err)
		}
		// Per-chunk batching: every chunk overlapping the rect is loaded
		// once. Cell-slots is a 512*512 array of u32 cell pointers (1 MiB
		// per chunk) — read once per chunk, then per-cell mColor is a
		// 4-byte read. Cuts ChunkSystem-traversal overhead from
		// O(W*H) ReadCellAt calls down to ~4 chunk slot reads + W*H 4-byte
		// reads. Reading cell mColors individually keeps the impl simple;
		// it's still fast enough for sub-1024² rects.
		const chunkSide = 512
		img := image.NewNRGBA(image.Rect(0, 0, int(in.W), int(in.H)))

		// Walk chunks covered by the rect. World coords use the
		// chunkCx = ((wx >> 9) - 0x100) & 0x1ff convention from
		// noita.hexpat.
		minWX, minWY := in.WX, in.WY
		maxWX, maxWY := in.WX+in.W-1, in.WY+in.H-1
		minCX := minWX >> 9
		minCY := minWY >> 9
		maxCX := maxWX >> 9
		maxCY := maxWY >> 9

		// Cache: chunk_idx → cell-slots blob (262144 u32 little-endian).
		// Memory: 1 MiB per loaded chunk. Bounded by (maxCX-minCX+1)*
		// (maxCY-minCY+1) which for a 1024² rect is at most 9 chunks → 9 MiB.
		type chunkBlob struct {
			loaded    bool
			cellSlots []byte // 4 * chunkSide * chunkSide bytes, or nil if chunk unloaded
		}
		blobs := make(map[uint64]*chunkBlob)

		blobAt := func(cx, cy int32) *chunkBlob {
			key := (uint64(uint32(cx)) << 32) | uint64(uint32(cy))
			if b, ok := blobs[key]; ok {
				return b
			}
			b := &chunkBlob{}
			blobs[key] = b
			info := reader.ReadCellAt(cx*chunkSide, cy*chunkSide)
			if info == nil || info.ChunkPtr == 0 || info.CellSlotsPtr == 0 {
				return b
			}
			buf := make([]byte, 4*chunkSide*chunkSide)
			if _, err := reader.Ctx.ReadAt(buf, int64(info.CellSlotsPtr)); err != nil {
				return b
			}
			b.loaded = true
			b.cellSlots = buf
			return b
		}

		for cy := minCY; cy <= maxCY; cy++ {
			for cx := minCX; cx <= maxCX; cx++ {
				blob := blobAt(cx, cy)
				if !blob.loaded {
					continue
				}
				// Pixel range within this chunk.
				cxBaseWX := cx * chunkSide
				cyBaseWY := cy * chunkSide
				x0 := minWX
				if cxBaseWX > x0 {
					x0 = cxBaseWX
				}
				y0 := minWY
				if cyBaseWY > y0 {
					y0 = cyBaseWY
				}
				x1 := maxWX
				if cxBaseWX+chunkSide-1 < x1 {
					x1 = cxBaseWX + chunkSide - 1
				}
				y1 := maxWY
				if cyBaseWY+chunkSide-1 < y1 {
					y1 = cyBaseWY + chunkSide - 1
				}
				for wy := y0; wy <= y1; wy++ {
					localY := wy - cyBaseWY
					for wx := x0; wx <= x1; wx++ {
						localX := wx - cxBaseWX
						slotOff := (localY*chunkSide + localX) * 4
						cellPtr := binary.LittleEndian.Uint32(blob.cellSlots[slotOff : slotOff+4])
						if cellPtr == 0 {
							// Air — leave transparent.
							continue
						}
						var px [4]byte
						if _, err := reader.Ctx.ReadAt(px[:], int64(cellPtr)+0x30); err != nil {
							continue
						}
						// BGRA in memory.
						img.SetNRGBA(int(wx-minWX), int(wy-minWY), color.NRGBA{
							R: px[2],
							G: px[1],
							B: px[0],
							A: px[3],
						})
					}
				}
			}
		}

		var pngBuf bytes.Buffer
		if err := png.Encode(&pngBuf, img); err != nil {
			return toolErr(fmt.Errorf("png encode: %w", err))
		}
		out := map[string]any{
			"wx":         in.WX,
			"wy":         in.WY,
			"w":          in.W,
			"h":          in.H,
			"png_base64": base64.StdEncoding.EncodeToString(pngBuf.Bytes()),
			"png_bytes":  pngBuf.Len(),
		}
		data, _ := json.Marshal(out)
		return textResult(string(data)), nil, nil
	})
}
