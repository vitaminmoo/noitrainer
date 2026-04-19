// Runtime tools attach to the running noita.exe process and mirror the
// introspection surface of cmd/cli. Each tool opens a fresh Reader per
// call; there is no shared state between tools.
package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
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
}
