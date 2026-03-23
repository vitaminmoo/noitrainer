package noita

//go:generate go run github.com/vitaminmoo/memtools/cmd/hexpatgen@latest -i noita.hexpat -o noita_gen.go -pkg noita

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"github.com/vitaminmoo/memtools/hexpat/runtime"
)

// Known static addresses in noita.exe (stable across runs).
const (
	AddrWorldSeed      = 0x01205004
	AddrDeathCount     = 0x01208AF8
	AddrNumOrbsTotal   = 0x01152544
	AddrEntityManager  = 0x01204B98
	AddrDeathMatchApp  = 0x01204BC0
	AddrGameGlobals    = 0x0122374C
	AddrWorldState     = 0x01205010
	AddrOrbPersistence = 0x01207404
)

// Component type IDs (runtime-assigned, validated from dumps).
const (
	TypeAbilityComponent           = 3
	TypeCharacterDataComponent     = 22
	TypeDamageModelComponent       = 31
	TypeInventory2Component        = 73
	TypeMaterialInventoryComponent = 101
	TypeWalletComponent            = 159
	TypeWorldStateComponent        = 161
)

// CellFactory material array: inline CellData structs at stride 0x290, base at CellFactory+0x18.
// CellData+0x00 = name (24-byte MsvcString).
const (
	cellDataStride         = 0x290
	cellFactoryArrayOffset = 0x18
)

// GameState holds a snapshot of all interesting game data.
type GameState struct {
	Connected bool
	Error     string

	WorldSeed    uint32
	DeathCount   int32
	NumOrbsTotal int32

	Globals    *GameGlobals
	WorldState *WorldStateComponent

	PlayerEntity *Entity
	PlayerHP     *DamageModelComponent
	PlayerWallet *WalletComponent
	PlayerChar   *CharacterDataComponent
	PlayerInv    *Inventory2Component
	Wands        []*InventoryItem
	Items        []*InventoryItem

	// Camera from WorldManager (GameGlobals.pWorldManager -> viewX/Y/W/H)
	CameraX float32
	CameraY float32
	ViewW   float32
	ViewH   float32
}

// MsvcStringValue extracts the Go string from an MsvcString.
// MaterialContent represents a material and its amount in a container.
type MaterialContent struct {
	MaterialID int
	Name       string
	Amount     float64
}

// InventoryItem wraps an AbilityComponent with its parent entity info.
type InventoryItem struct {
	Entity   *Entity
	Ability  *AbilityComponent
	Contents []MaterialContent // populated for potions/flasks
}

// IsWand returns true if this item is a wand (has gun script), false for potions/consumables.
func (item *InventoryItem) IsWand() bool {
	return item.Ability.UseGunScript
}

// Name returns a display name.
func (item *InventoryItem) Name(ctx *runtime.ReadContext) string {
	name := MsvcStringValue(&item.Ability.UiName, ctx)
	if name == "" {
		name = MsvcStringValue(&item.Ability.SpriteFile, ctx)
	}
	return name
}

func MsvcStringValue(s *MsvcString, ctx *runtime.ReadContext) string {
	if s.Length == 0 {
		return ""
	}
	if s.Capacity <= 15 {
		// Inline SSO: data is in the first Length bytes
		n := s.Length
		if n > 16 {
			n = 16
		}
		return string(s.Data[:n])
	}
	// Heap-allocated: first 4 bytes of Data are a pointer
	heapPtr := binary.LittleEndian.Uint32(s.Data[:4])
	if heapPtr == 0 {
		return "<null>"
	}
	n := s.Length
	if n > 4096 {
		n = 4096
	}
	buf := make([]byte, n)
	if _, err := ctx.ReadAt(buf, int64(heapPtr)); err != nil {
		return fmt.Sprintf("<heap@0x%08X len=%d err=%v>", heapPtr, s.Length, err)
	}
	return string(buf)
}

// Reader reads Noita game state from process memory.
type Reader struct {
	proc io.ReadSeeker
	Ctx  *runtime.ReadContext
}

func NewReader(proc io.ReadSeeker) *Reader {
	return &Reader{
		proc: proc,
		Ctx:  runtime.NewReadContext(proc),
	}
}

func (r *Reader) readU32(addr int64) (uint32, error) {
	var buf [4]byte
	if _, err := r.Ctx.ReadAt(buf[:], addr); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:]), nil
}

func (r *Reader) readS32(addr int64) (int32, error) {
	v, err := r.readU32(addr)
	return int32(v), err
}

func (r *Reader) readF32(addr int64) (float32, error) {
	var buf [4]byte
	if _, err := r.Ctx.ReadAt(buf[:], addr); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:])), nil
}

// ReadState reads a complete game state snapshot.
func (r *Reader) ReadState() *GameState {
	gs := &GameState{Connected: true}

	// Read static globals
	if v, err := r.readU32(AddrWorldSeed); err == nil {
		gs.WorldSeed = v
	} else {
		gs.Error = fmt.Sprintf("read world seed: %v", err)
		gs.Connected = false
		return gs
	}

	gs.DeathCount, _ = r.readS32(AddrDeathCount)
	gs.NumOrbsTotal, _ = r.readS32(AddrNumOrbsTotal)

	// Read GameGlobals (pointer indirection)
	if globalsPtr, err := r.readU32(AddrGameGlobals); err == nil && globalsPtr != 0 {
		gs.Globals, _ = ReadGameGlobals(r.Ctx, uintptr(globalsPtr))

		// Read camera from WorldManager (pWorldManager -> viewX/Y/W/H at +0x00..+0x0C)
		if gs.Globals != nil && gs.Globals.PWorldManager != 0 {
			wm := int64(gs.Globals.PWorldManager)
			gs.ViewW, _ = r.readF32(wm + 0x08)
			gs.ViewH, _ = r.readF32(wm + 0x0C)
			viewX, _ := r.readF32(wm + 0x00)
			viewY, _ := r.readF32(wm + 0x04)
			gs.CameraX = viewX + gs.ViewW*0.5
			gs.CameraY = viewY + gs.ViewH*0.5
		}
	}

	// Read WorldStateComponent (pointer indirection)
	if wsPtr, err := r.readU32(AddrWorldState); err == nil && wsPtr != 0 {
		gs.WorldState, _ = ReadWorldStateComponent(r.Ctx, uintptr(wsPtr))
	}

	// Find player entity via DeathMatchApp -> player_entities vector
	dmaPtr, err := r.readU32(AddrDeathMatchApp)
	if err != nil || dmaPtr == 0 {
		return gs
	}
	dma, _ := ReadDeathMatchApp(r.Ctx, uintptr(dmaPtr))
	if dma == nil || dma.PlayerEntities.BeginPtr == 0 {
		return gs
	}

	// Read first player entity pointer
	playerEntityPtr, err := r.readU32(int64(dma.PlayerEntities.BeginPtr))
	if err != nil || playerEntityPtr == 0 {
		return gs
	}

	gs.PlayerEntity, _ = ReadEntity(r.Ctx, uintptr(playerEntityPtr))
	if gs.PlayerEntity == nil {
		return gs
	}

	// Look up components via EntityManager
	emPtr, err := r.readU32(AddrEntityManager)
	if err != nil || emPtr == 0 {
		return gs
	}
	em, _ := ReadEntityManager(r.Ctx, uintptr(emPtr))
	if em == nil {
		return gs
	}

	// Read components for the player entity
	gs.PlayerHP = readComponent[DamageModelComponent](r, em, gs.PlayerEntity.SlotIndex, TypeDamageModelComponent, ReadDamageModelComponent)
	gs.PlayerWallet = readComponent[WalletComponent](r, em, gs.PlayerEntity.SlotIndex, TypeWalletComponent, ReadWalletComponent)
	gs.PlayerChar = readComponent[CharacterDataComponent](r, em, gs.PlayerEntity.SlotIndex, TypeCharacterDataComponent, ReadCharacterDataComponent)
	gs.PlayerInv = readComponent[Inventory2Component](r, em, gs.PlayerEntity.SlotIndex, TypeInventory2Component, ReadInventory2Component)

	// Read inventory: traverse player → children → inventory containers → children → AbilityComponent
	// Wands/items are NOT on the player entity; they're child entities in the inventory hierarchy.
	allItems := r.findInventoryItems(em, gs.PlayerEntity)
	for _, item := range allItems {
		if item.IsWand() {
			gs.Wands = append(gs.Wands, item)
		} else {
			gs.Items = append(gs.Items, item)
		}
	}

	return gs
}

// readMaterialName reads the material name for a given material ID from CellFactory.
func (r *Reader) readMaterialName(matID int) string {
	globalsPtr, err := r.readU32(AddrGameGlobals)
	if err != nil || globalsPtr == 0 {
		return fmt.Sprintf("mat_%d", matID)
	}
	cfPtr, err := r.readU32(int64(globalsPtr) + 0x18)
	if err != nil || cfPtr == 0 {
		return fmt.Sprintf("mat_%d", matID)
	}
	// Read base pointer to CellData array at CellFactory+0x18
	arrayBase, err := r.readU32(int64(cfPtr) + cellFactoryArrayOffset)
	if err != nil || arrayBase == 0 {
		return fmt.Sprintf("mat_%d", matID)
	}
	// CellData[matID] at arrayBase + matID * 0x290, name is MsvcString at +0x00
	addr := int64(arrayBase) + int64(matID)*cellDataStride
	ms, errs := ReadMsvcString(r.Ctx, uintptr(addr))
	if errs.HasFatal() || ms == nil {
		return fmt.Sprintf("mat_%d", matID)
	}
	name := MsvcStringValue(ms, r.Ctx)
	if name == "" {
		return fmt.Sprintf("mat_%d", matID)
	}
	return name
}

// readPotionContents reads MaterialInventoryComponent for an entity and returns non-zero materials.
func (r *Reader) readPotionContents(em *EntityManager, slotIndex int32) []MaterialContent {
	compPtr := r.findComponentPtr(em, slotIndex, TypeMaterialInventoryComponent)
	if compPtr == 0 {
		return nil
	}
	// Material count vector at MaterialInventoryComponent+0x80 (std::vector<double>)
	vecBegin, _ := r.readU32(int64(compPtr) + 0x80)
	vecEnd, _ := r.readU32(int64(compPtr) + 0x84)
	if vecBegin == 0 || vecEnd <= vecBegin {
		return nil
	}
	byteLen := vecEnd - vecBegin
	numMaterials := byteLen / 8
	if numMaterials > 1000 {
		return nil
	}
	vecData := make([]byte, byteLen)
	if _, err := r.Ctx.ReadAt(vecData, int64(vecBegin)); err != nil {
		return nil
	}
	var contents []MaterialContent
	for i := uint32(0); i < numMaterials; i++ {
		amount := math.Float64frombits(binary.LittleEndian.Uint64(vecData[i*8 : i*8+8]))
		if amount > 0 {
			contents = append(contents, MaterialContent{
				MaterialID: int(i),
				Name:       r.readMaterialName(int(i)),
				Amount:     amount,
			})
		}
	}
	return contents
}

// findComponentPtr looks up a component pointer for an entity slot + type ID.
func (r *Reader) findComponentPtr(em *EntityManager, slotIndex int32, typeID int) uint32 {
	if slotIndex < 0 || em == nil {
		return 0
	}
	numBuffers := (em.ComponentBuffers.EndPtr - em.ComponentBuffers.BeginPtr) / 4
	if uint32(typeID) >= numBuffers {
		return 0
	}
	bufferPtr, err := r.readU32(int64(em.ComponentBuffers.BeginPtr) + int64(typeID)*4)
	if err != nil || bufferPtr == 0 {
		return 0
	}
	cb, _ := ReadComponentBuffer(r.Ctx, uintptr(bufferPtr))
	if cb == nil {
		return 0
	}
	numSparse := (cb.SparseIndex.EndPtr - cb.SparseIndex.BeginPtr) / 4
	if uint32(slotIndex) >= numSparse {
		return 0
	}
	denseIdx, err := r.readS32(int64(cb.SparseIndex.BeginPtr) + int64(slotIndex)*4)
	if err != nil || denseIdx < 0 {
		return 0
	}
	compPtr, _ := r.readU32(int64(cb.Components.BeginPtr) + int64(denseIdx)*4)
	return compPtr
}

// readChildEntityPtrs reads the Entity* pointers from a ChildrenContainer.
func (r *Reader) readChildEntityPtrs(childrenPtr uint32) []uint32 {
	if childrenPtr == 0 {
		return nil
	}
	cc, _ := ReadChildrenContainer(r.Ctx, uintptr(childrenPtr))
	if cc == nil || cc.BeginPtr == 0 || cc.EndPtr <= cc.BeginPtr {
		return nil
	}
	count := (cc.EndPtr - cc.BeginPtr) / 4
	if count > 1000 {
		return nil // sanity limit
	}
	ptrs := make([]uint32, count)
	for i := uint32(0); i < count; i++ {
		p, err := r.readU32(int64(cc.BeginPtr) + int64(i)*4)
		if err != nil {
			break
		}
		ptrs[i] = p
	}
	return ptrs
}

// findInventoryItems traverses player children (and grandchildren) looking for entities with AbilityComponent.
func (r *Reader) findInventoryItems(em *EntityManager, player *Entity) []*InventoryItem {
	if em == nil || player == nil {
		return nil
	}

	var items []*InventoryItem

	// Player → children (includes inventory container entities)
	childPtrs := r.readChildEntityPtrs(player.ChildrenPtr)
	for _, cp := range childPtrs {
		if cp == 0 {
			continue
		}
		child, _ := ReadEntity(r.Ctx, uintptr(cp))
		if child == nil || child.PendingKill >= 1 {
			continue
		}

		// Check if this child has AbilityComponent
		if ac := readComponent[AbilityComponent](r, em, child.SlotIndex, TypeAbilityComponent, ReadAbilityComponent); ac != nil {
			item := &InventoryItem{Entity: child, Ability: ac}
			if !item.IsWand() {
				item.Contents = r.readPotionContents(em, child.SlotIndex)
			}
			items = append(items, item)
		}

		// Check grandchildren (inventory container → item entities)
		grandchildPtrs := r.readChildEntityPtrs(child.ChildrenPtr)
		for _, gcp := range grandchildPtrs {
			if gcp == 0 {
				continue
			}
			grandchild, _ := ReadEntity(r.Ctx, uintptr(gcp))
			if grandchild == nil || grandchild.PendingKill >= 1 {
				continue
			}
			if ac := readComponent[AbilityComponent](r, em, grandchild.SlotIndex, TypeAbilityComponent, ReadAbilityComponent); ac != nil {
				item := &InventoryItem{Entity: grandchild, Ability: ac}
				if !item.IsWand() {
					item.Contents = r.readPotionContents(em, grandchild.SlotIndex)
				}
				items = append(items, item)
			}
		}
	}

	return items
}

type readFunc[T any] func(ctx *runtime.ReadContext, addr uintptr) (*T, runtime.Errors)

// readComponent looks up a single component for an entity via the ECS sparse-dense index.
func readComponent[T any](r *Reader, em *EntityManager, slotIndex int32, typeID int, readFn readFunc[T]) *T {
	components := readAllComponents(r, em, slotIndex, typeID, readFn)
	if len(components) > 0 {
		return components[0]
	}
	return nil
}

// readAllComponents reads all components of a given type for an entity, following the linked chain.
func readAllComponents[T any](r *Reader, em *EntityManager, slotIndex int32, typeID int, readFn readFunc[T]) []*T {
	if slotIndex < 0 {
		return nil
	}

	// Get component buffer pointer: componentBuffers[typeID]
	numBuffers := (em.ComponentBuffers.EndPtr - em.ComponentBuffers.BeginPtr) / 4
	if uint32(typeID) >= numBuffers {
		return nil
	}

	bufferPtr, err := r.readU32(int64(em.ComponentBuffers.BeginPtr) + int64(typeID)*4)
	if err != nil || bufferPtr == 0 {
		return nil
	}

	cb, _ := ReadComponentBuffer(r.Ctx, uintptr(bufferPtr))
	if cb == nil {
		return nil
	}

	// Sparse lookup: sparseIndex[slotIndex] -> denseIndex
	numSparse := (cb.SparseIndex.EndPtr - cb.SparseIndex.BeginPtr) / 4
	if uint32(slotIndex) >= numSparse {
		return nil
	}

	denseIdx, err := r.readS32(int64(cb.SparseIndex.BeginPtr) + int64(slotIndex)*4)
	if err != nil || denseIdx < 0 {
		return nil
	}

	var results []*T

	for denseIdx >= 0 {
		// components[denseIdx] -> component pointer
		numComponents := (cb.Components.EndPtr - cb.Components.BeginPtr) / 4
		if uint32(denseIdx) >= numComponents {
			break
		}

		compPtr, err := r.readU32(int64(cb.Components.BeginPtr) + int64(denseIdx)*4)
		if err != nil || compPtr == 0 {
			break
		}

		comp, _ := readFn(r.Ctx, uintptr(compPtr))
		if comp != nil {
			results = append(results, comp)
		}

		// Follow nextIndex chain for multiple components of same type
		numNext := (cb.NextIndex.EndPtr - cb.NextIndex.BeginPtr) / 4
		if uint32(denseIdx) >= numNext {
			break
		}
		nextIdx, err := r.readS32(int64(cb.NextIndex.BeginPtr) + int64(denseIdx)*4)
		if err != nil {
			break
		}
		denseIdx = nextIdx
	}

	return results
}
