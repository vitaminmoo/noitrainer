package noita

//go:generate go run github.com/vitaminmoo/memtools/cmd/hexpatgen -i noita.hexpat -o noita_gen.go -pkg noita

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/vitaminmoo/memtools/hexpat/runtime"
)

// CellData stride in the CellFactory material array.
const cellDataStride = 0x290

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

	// All entities in the world
	Entities []*EntitySummary

	// Camera from WorldManager
	CameraX float32
	CameraY float32
	ViewW   float32
	ViewH   float32
}

// EntitySummary holds basic info about an entity for list display.
type EntitySummary struct {
	Entity           *Entity
	Name             string
	Ptr              uint32
	HasHP            bool
	HasWallet        bool
	HasAbility       bool
	HasCharData      bool
	Hitbox           *HitboxComponent           // nil if no hitbox
	CollisionTrigger *CollisionTriggerComponent // nil if no trigger
	Sprite           *SpriteComponent           // nil if no sprite
	Lua              *LuaComponent              // nil if no lua
	Item             *ItemComponent             // nil if no item
	Contents         []MaterialContent          // potion/flask contents
	ComponentIDs     []TypeID                   // all component type IDs present on this entity
}

// EntityDetails holds full component data for a selected entity.
type EntityDetails struct {
	Entity   *Entity
	Name     string
	HP       *DamageModelComponent
	Wallet   *WalletComponent
	Char     *CharacterDataComponent
	Inv      *Inventory2Component
	Ability  *AbilityComponent
	Sprite   *SpriteComponent
	Item     *ItemComponent
	Velocity *VelocityComponent
	Light    *LightComponent
	Effect   *GameEffectComponent
	Lua      *LuaComponent
	Contents []MaterialContent
	Children []*EntitySummary
}

// ComponentBufferInfo holds metadata about a component buffer (type).
type ComponentBufferInfo struct {
	TypeIndex   int
	Name        string
	ActiveCount int32
	Capacity    int32
	Ptr         uint32
}

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
	name := item.Ability.UiName.FormatMsvcString(ctx)
	if name == "" {
		name = item.Ability.SpriteFile.FormatMsvcString(ctx)
	}
	return name
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

// readEM reads the EntityManager via its static pointer.
func (r *Reader) readEM() *EntityManager {
	em, _ := ReadGEntityManager(r.Ctx)
	return em
}

// CameraState holds just the camera position and view dimensions.
type CameraState struct {
	CameraX, CameraY float32
	ViewW, ViewH     float32
}

// ReadCamera reads only the camera position and view rect (lightweight).
func (r *Reader) ReadCamera() *CameraState {
	globals, _ := ReadGGameGlobals(r.Ctx)
	if globals == nil {
		return nil
	}
	vr, _ := globals.ReadPWorldManager(r.Ctx)
	if vr == nil {
		return nil
	}
	return &CameraState{
		CameraX: vr.ViewX + vr.ViewWidth*0.5,
		CameraY: vr.ViewY + vr.ViewHeight*0.5,
		ViewW:   vr.ViewWidth,
		ViewH:   vr.ViewHeight,
	}
}

// ReadState reads a complete game state snapshot.
func (r *Reader) ReadState() *GameState {
	gs := &GameState{Connected: true}

	// Read static globals
	if v, err := ReadGWorldSeed(r.Ctx); err == nil {
		gs.WorldSeed = v
	} else {
		gs.Error = fmt.Sprintf("read world seed: %v", err)
		gs.Connected = false
		return gs
	}

	gs.DeathCount, _ = ReadGDeathCount(r.Ctx)
	gs.NumOrbsTotal, _ = ReadGNumOrbsTotal(r.Ctx)

	// Read GameGlobals (pointer indirection)
	gs.Globals, _ = ReadGGameGlobals(r.Ctx)
	if gs.Globals != nil {
		// Read camera from WorldManager view rect
		if vr, _ := gs.Globals.ReadPWorldManager(r.Ctx); vr != nil {
			gs.ViewW = vr.ViewWidth
			gs.ViewH = vr.ViewHeight
			gs.CameraX = vr.ViewX + vr.ViewWidth*0.5
			gs.CameraY = vr.ViewY + vr.ViewHeight*0.5
		}
	}

	// Read WorldStateComponent (pointer indirection)
	gs.WorldState, _ = ReadGWorldState(r.Ctx)

	// Find player entity via DeathMatchApp -> player_entities vector
	dma, _ := ReadGDeathMatchApp(r.Ctx)
	if dma == nil || len(dma.PlayerEntities.Elements) == 0 {
		return gs
	}

	playerEntityPtr := dma.PlayerEntities.Elements[0]
	if playerEntityPtr == 0 {
		return gs
	}

	gs.PlayerEntity, _ = ReadEntity(r.Ctx, uintptr(playerEntityPtr))
	if gs.PlayerEntity == nil {
		return gs
	}

	em := r.readEM()
	if em == nil {
		return gs
	}

	// Read components for the player entity
	gs.PlayerHP = readComponent[DamageModelComponent](r, em, gs.PlayerEntity.SlotIndex, TypeIDDamageModelComponent, ReadDamageModelComponent)
	gs.PlayerWallet = readComponent[WalletComponent](r, em, gs.PlayerEntity.SlotIndex, TypeIDWalletComponent, ReadWalletComponent)
	gs.PlayerChar = readComponent[CharacterDataComponent](r, em, gs.PlayerEntity.SlotIndex, TypeIDCharacterDataComponent, ReadCharacterDataComponent)
	gs.PlayerInv = readComponent[Inventory2Component](r, em, gs.PlayerEntity.SlotIndex, TypeIDInventory2Component, ReadInventory2Component)

	// Read inventory
	allItems := r.findInventoryItems(em, gs.PlayerEntity)
	for _, item := range allItems {
		if item.IsWand() {
			gs.Wands = append(gs.Wands, item)
		} else {
			gs.Items = append(gs.Items, item)
		}
	}

	// Read all entities
	gs.Entities = r.ReadEntityList()

	return gs
}

// readMaterialName reads the material name for a given material ID from CellFactory.
func (r *Reader) readMaterialName(matID int) string {
	globals, _ := ReadGGameGlobals(r.Ctx)
	if globals == nil {
		return fmt.Sprintf("mat_%d", matID)
	}
	cf, _ := globals.ReadPCellFactory(r.Ctx)
	if cf == nil || cf.CellDataArrayPtr == 0 {
		return fmt.Sprintf("mat_%d", matID)
	}
	addr := uintptr(cf.CellDataArrayPtr) + uintptr(matID)*cellDataStride
	cd, _ := ReadCellData(r.Ctx, addr)
	if cd == nil {
		return fmt.Sprintf("mat_%d", matID)
	}
	name := cd.Name.FormatMsvcString(r.Ctx)
	if name == "" {
		return fmt.Sprintf("mat_%d", matID)
	}
	return name
}

// readPotionContents reads MaterialInventoryComponent for an entity and returns non-zero materials.
func (r *Reader) readPotionContents(em *EntityManager, slotIndex int32) []MaterialContent {
	mic := readComponent[MaterialInventoryComponent](r, em, slotIndex, TypeIDMaterialInventoryComponent, ReadMaterialInventoryComponent)
	if mic == nil {
		return nil
	}
	var contents []MaterialContent
	for i, amount := range mic.CountPerMaterialType.Elements {
		if amount > 0 {
			contents = append(contents, MaterialContent{
				MaterialID: i,
				Name:       r.readMaterialName(i),
				Amount:     amount,
			})
		}
	}
	return contents
}

// lookupComponentBufferPtr returns the ComponentBuffer pointer for a given type ID.
func lookupComponentBufferPtr(em *EntityManager, typeID TypeID) uint32 {
	if em == nil || int(typeID) >= len(em.ComponentBuffers.Elements) {
		return 0
	}
	return em.ComponentBuffers.Elements[int(typeID)]
}

// findComponentPtr resolves a component pointer using lazy ComponentBuffer reads.
func (r *Reader) findComponentPtr(em *EntityManager, slotIndex int32, typeID TypeID) uint32 {
	bufferPtr := lookupComponentBufferPtr(em, typeID)
	if bufferPtr == 0 {
		return 0
	}
	cbr := NewComponentBufferReader(r.Ctx, uintptr(bufferPtr))
	sparse := cbr.SparseIndex()
	beginPtr, _ := sparse.BeginPtr()
	endPtr, _ := sparse.EndPtr()
	numSparse := (endPtr - beginPtr) / 4
	if uint32(slotIndex) >= numSparse {
		return 0
	}
	denseIdx, err := r.readS32(int64(beginPtr) + int64(slotIndex)*4)
	if err != nil || denseIdx < 0 {
		return 0
	}
	comps := cbr.Components()
	compBegin, _ := comps.BeginPtr()
	compEnd, _ := comps.EndPtr()
	numComps := (compEnd - compBegin) / 4
	if uint32(denseIdx) >= numComps {
		return 0
	}
	compPtr, _ := r.readU32(int64(compBegin) + int64(denseIdx)*4)
	return compPtr
}

// hasComponent checks if an entity has a component type, reading minimal data.
func (r *Reader) hasComponent(em *EntityManager, slotIndex int32, typeID TypeID) bool {
	bufferPtr := lookupComponentBufferPtr(em, typeID)
	if bufferPtr == 0 {
		return false
	}
	cbr := NewComponentBufferReader(r.Ctx, uintptr(bufferPtr))
	sparse := cbr.SparseIndex()
	beginPtr, _ := sparse.BeginPtr()
	endPtr, _ := sparse.EndPtr()
	numSparse := (endPtr - beginPtr) / 4
	if uint32(slotIndex) >= numSparse {
		return false
	}
	denseIdx, err := r.readS32(int64(beginPtr) + int64(slotIndex)*4)
	return err == nil && denseIdx >= 0
}

// readChildEntityPtrs reads the Entity* pointers from a ChildrenContainer.
func (r *Reader) readChildEntityPtrs(entity *Entity) []uint32 {
	cc, _ := entity.ReadChildrenPtr(r.Ctx)
	if cc == nil || len(cc.Children) == 0 {
		return nil
	}
	if len(cc.Children) > 1000 {
		return nil // sanity limit
	}
	return cc.Children
}

// findInventoryItems traverses player children (and grandchildren) looking for entities with AbilityComponent.
func (r *Reader) findInventoryItems(em *EntityManager, player *Entity) []*InventoryItem {
	if em == nil || player == nil {
		return nil
	}

	var items []*InventoryItem

	childPtrs := r.readChildEntityPtrs(player)
	for _, cp := range childPtrs {
		if cp == 0 {
			continue
		}
		child, _ := ReadEntity(r.Ctx, uintptr(cp))
		if child == nil || child.PendingKill >= 1 {
			continue
		}

		if ac := readComponent[AbilityComponent](r, em, child.SlotIndex, TypeIDAbilityComponent, ReadAbilityComponent); ac != nil {
			item := &InventoryItem{Entity: child, Ability: ac}
			if !item.IsWand() {
				item.Contents = r.readPotionContents(em, child.SlotIndex)
			}
			items = append(items, item)
		}

		grandchildPtrs := r.readChildEntityPtrs(child)
		for _, gcp := range grandchildPtrs {
			if gcp == 0 {
				continue
			}
			grandchild, _ := ReadEntity(r.Ctx, uintptr(gcp))
			if grandchild == nil || grandchild.PendingKill >= 1 {
				continue
			}
			if ac := readComponent[AbilityComponent](r, em, grandchild.SlotIndex, TypeIDAbilityComponent, ReadAbilityComponent); ac != nil {
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

// ReadEntityList reads all entities from the EntityManager and returns summaries.
func (r *Reader) ReadEntityList() []*EntitySummary {
	em := r.readEM()
	if em == nil {
		return nil
	}

	count := (em.EntityArray.EndPtr - em.EntityArray.BeginPtr) / 4
	if count == 0 || count > 100000 {
		return nil
	}

	// Batch read entity pointers
	ptrBuf := make([]byte, count*4)
	if _, err := r.Ctx.ReadAt(ptrBuf, int64(em.EntityArray.BeginPtr)); err != nil {
		return nil
	}

	var summaries []*EntitySummary
	for i := uint32(0); i < count; i++ {
		ePtr := binary.LittleEndian.Uint32(ptrBuf[i*4 : i*4+4])
		if ePtr == 0 {
			continue
		}
		// Use lazy reader — only read the fields we need
		er := NewEntityReader(r.Ctx, uintptr(ePtr))
		pendingKill, _ := er.PendingKill()
		if pendingKill >= 1 {
			continue
		}
		slotIndex, _ := er.SlotIndex()
		entityId, _ := er.EntityId()
		posX, _ := er.PosX()
		posY, _ := er.PosY()
		parentPtr, _ := er.ParentEntityPtr()
		nameStr, _ := er.Name().Read()

		// Build a minimal Entity for the summary (avoids full eager read)
		entity := &Entity{
			EntityId:        entityId,
			SlotIndex:       slotIndex,
			PosX:            posX,
			PosY:            posY,
			ParentEntityPtr: parentPtr,
		}
		if nameStr != nil {
			entity.Name = *nameStr
		}

		name := entity.Name.FormatMsvcString(r.Ctx)
		compIDs := r.FindEntityComponentIDs(em, slotIndex)
		summary := &EntitySummary{
			Entity:           entity,
			Name:             name,
			Ptr:              ePtr,
			HasHP:            r.hasComponent(em, slotIndex, TypeIDDamageModelComponent),
			HasWallet:        r.hasComponent(em, slotIndex, TypeIDWalletComponent),
			HasAbility:       r.hasComponent(em, slotIndex, TypeIDAbilityComponent),
			HasCharData:      r.hasComponent(em, slotIndex, TypeIDCharacterDataComponent),
			Hitbox:           readComponent[HitboxComponent](r, em, slotIndex, TypeIDHitboxComponent, ReadHitboxComponent),
			CollisionTrigger: readComponent[CollisionTriggerComponent](r, em, slotIndex, TypeIDCollisionTriggerComponent, ReadCollisionTriggerComponent),
			Sprite:           readComponent[SpriteComponent](r, em, slotIndex, TypeIDSpriteComponent, ReadSpriteComponent),
			Lua:              readComponent[LuaComponent](r, em, slotIndex, TypeIDLuaComponent, ReadLuaComponent),
			Item:             readComponent[ItemComponent](r, em, slotIndex, TypeIDItemComponent, ReadItemComponent),
			Contents:         r.readPotionContents(em, slotIndex),
			ComponentIDs:     compIDs,
		}
		summaries = append(summaries, summary)
	}
	return summaries
}

// ReadEntityDetails reads full component data for a specific entity.
func (r *Reader) ReadEntityDetails(entityPtr uint32) *EntityDetails {
	if entityPtr == 0 {
		return nil
	}
	e, _ := ReadEntity(r.Ctx, uintptr(entityPtr))
	if e == nil {
		return nil
	}

	em := r.readEM()
	if em == nil {
		return nil
	}

	details := &EntityDetails{
		Entity:   e,
		Name:     e.Name.FormatMsvcString(r.Ctx),
		HP:       readComponent[DamageModelComponent](r, em, e.SlotIndex, TypeIDDamageModelComponent, ReadDamageModelComponent),
		Wallet:   readComponent[WalletComponent](r, em, e.SlotIndex, TypeIDWalletComponent, ReadWalletComponent),
		Char:     readComponent[CharacterDataComponent](r, em, e.SlotIndex, TypeIDCharacterDataComponent, ReadCharacterDataComponent),
		Inv:      readComponent[Inventory2Component](r, em, e.SlotIndex, TypeIDInventory2Component, ReadInventory2Component),
		Ability:  readComponent[AbilityComponent](r, em, e.SlotIndex, TypeIDAbilityComponent, ReadAbilityComponent),
		Sprite:   readComponent[SpriteComponent](r, em, e.SlotIndex, TypeIDSpriteComponent, ReadSpriteComponent),
		Item:     readComponent[ItemComponent](r, em, e.SlotIndex, TypeIDItemComponent, ReadItemComponent),
		Velocity: readComponent[VelocityComponent](r, em, e.SlotIndex, TypeIDVelocityComponent, ReadVelocityComponent),
		Light:    readComponent[LightComponent](r, em, e.SlotIndex, TypeIDLightComponent, ReadLightComponent),
		Effect:   readComponent[GameEffectComponent](r, em, e.SlotIndex, TypeIDGameEffectComponent, ReadGameEffectComponent),
		Lua:      readComponent[LuaComponent](r, em, e.SlotIndex, TypeIDLuaComponent, ReadLuaComponent),
	}
	details.Contents = r.readPotionContents(em, e.SlotIndex)

	// Read children
	childPtrs := r.readChildEntityPtrs(e)
	for _, cp := range childPtrs {
		if cp == 0 {
			continue
		}
		child, _ := ReadEntity(r.Ctx, uintptr(cp))
		if child == nil || child.PendingKill >= 1 {
			continue
		}
		details.Children = append(details.Children, &EntitySummary{
			Entity: child,
			Name:   child.Name.FormatMsvcString(r.Ctx),
			Ptr:    cp,
		})
	}

	return details
}

// ReadComponentTypeName reads the C string at a component's PTypeName pointer.
func (r *Reader) ReadComponentTypeName(compPtr uint32) string {
	if compPtr == 0 {
		return ""
	}
	hdr := NewComponentHeaderReader(r.Ctx, uintptr(compPtr))
	pTypeName, err := hdr.PTypeName()
	if err != nil || pTypeName == 0 {
		return ""
	}
	buf := make([]byte, 128)
	if _, err := r.Ctx.ReadAt(buf, int64(pTypeName)); err != nil {
		return ""
	}
	for i, b := range buf {
		if b == 0 {
			return string(buf[:i])
		}
	}
	return string(buf)
}

// ReadComponentBuffers returns metadata for all component buffer types in the ECS.
func (r *Reader) ReadComponentBuffers() []*ComponentBufferInfo {
	em := r.readEM()
	if em == nil {
		return nil
	}

	if len(em.ComponentBuffers.Elements) == 0 {
		return nil
	}

	var infos []*ComponentBufferInfo
	for i, bufPtr := range em.ComponentBuffers.Elements {
		if bufPtr == 0 {
			continue
		}
		cbr := NewComponentBufferReader(r.Ctx, uintptr(bufPtr))
		activeCount, _ := cbr.ActiveCount()
		nameMs, _ := cbr.NameString().Read()
		var name string
		if nameMs != nil {
			name = nameMs.FormatMsvcString(r.Ctx)
		}
		// If NameString is empty, resolve via PTypeName from an active component
		if name == "" && activeCount > 0 {
			comps := cbr.Components()
			compBegin, _ := comps.BeginPtr()
			compEnd, _ := comps.EndPtr()
			if compBegin != 0 {
				numComps := (compEnd - compBegin) / 4
				for j := uint32(0); j < numComps && j < 16; j++ {
					compPtr, err := r.readU32(int64(compBegin) + int64(j)*4)
					if err != nil || compPtr == 0 {
						continue
					}
					if n := r.ReadComponentTypeName(compPtr); n != "" {
						name = n
						break
					}
				}
			}
		}
		if activeCount == 0 && name == "" {
			continue
		}
		capacity, _ := cbr.CapacityLimit()
		infos = append(infos, &ComponentBufferInfo{
			TypeIndex:   i,
			Name:        name,
			ActiveCount: activeCount,
			Capacity:    capacity,
			Ptr:         bufPtr,
		})
	}
	return infos
}

// ReadEntityManagerPtr returns the entity manager. Exported for CLI tools.
func (r *Reader) ReadEntityManagerPtr() (*EntityManager, uint32) {
	em, _ := ReadGEntityManager(r.Ctx)
	return em, AddrGEntityManager
}

// FindEntityComponentIDs returns all component type IDs that an entity has.
func (r *Reader) FindEntityComponentIDs(em *EntityManager, slotIndex int32) []TypeID {
	if slotIndex < 0 || em == nil {
		return nil
	}

	var ids []TypeID
	for typeID, bufPtr := range em.ComponentBuffers.Elements {
		if bufPtr == 0 {
			continue
		}
		// Lazy read — only access ActiveCount and SparseIndex
		cbr := NewComponentBufferReader(r.Ctx, uintptr(bufPtr))
		activeCount, _ := cbr.ActiveCount()
		if activeCount == 0 {
			continue
		}
		sparse := cbr.SparseIndex()
		beginPtr, _ := sparse.BeginPtr()
		endPtr, _ := sparse.EndPtr()
		numSparse := (endPtr - beginPtr) / 4
		if uint32(slotIndex) >= numSparse {
			continue
		}
		denseIdx, err := r.readS32(int64(beginPtr) + int64(slotIndex)*4)
		if err != nil || denseIdx < 0 {
			continue
		}
		ids = append(ids, TypeID(typeID))
	}
	return ids
}

// ReadRawComponent reads raw bytes for an arbitrary component type at a given entity slot.
func (r *Reader) ReadRawComponent(em *EntityManager, slotIndex int32, typeID TypeID, size int) (uint32, []byte) {
	compPtr := r.findComponentPtr(em, slotIndex, typeID)
	if compPtr == 0 {
		return 0, nil
	}
	buf := make([]byte, size)
	if _, err := r.Ctx.ReadAt(buf, int64(compPtr)); err != nil {
		return compPtr, nil
	}
	return compPtr, buf
}

type readFunc[T any] func(ctx *runtime.ReadContext, addr uintptr) (*T, runtime.Errors)

// readComponent looks up a single component for an entity via the ECS sparse-dense index.
func readComponent[T any](r *Reader, em *EntityManager, slotIndex int32, typeID TypeID, readFn readFunc[T]) *T {
	components := readAllComponents(r, em, slotIndex, typeID, readFn)
	if len(components) > 0 {
		return components[0]
	}
	return nil
}

// readAllComponents reads all components of a given type for an entity, following the linked chain.
func readAllComponents[T any](r *Reader, em *EntityManager, slotIndex int32, typeID TypeID, readFn readFunc[T]) []*T {
	if slotIndex < 0 {
		return nil
	}

	bufferPtr := lookupComponentBufferPtr(em, typeID)
	if bufferPtr == 0 {
		return nil
	}

	cbr := NewComponentBufferReader(r.Ctx, uintptr(bufferPtr))
	sparse := cbr.SparseIndex()
	sparseBegin, _ := sparse.BeginPtr()
	sparseEnd, _ := sparse.EndPtr()
	numSparse := (sparseEnd - sparseBegin) / 4
	if uint32(slotIndex) >= numSparse {
		return nil
	}
	denseIdx, err := r.readS32(int64(sparseBegin) + int64(slotIndex)*4)
	if err != nil || denseIdx < 0 {
		return nil
	}

	comps := cbr.Components()
	compBegin, _ := comps.BeginPtr()
	compEnd, _ := comps.EndPtr()
	next := cbr.NextIndex()
	nextBegin, _ := next.BeginPtr()
	nextEnd, _ := next.EndPtr()

	var results []*T

	for denseIdx >= 0 {
		numComponents := (compEnd - compBegin) / 4
		if uint32(denseIdx) >= numComponents {
			break
		}

		compPtr, err := r.readU32(int64(compBegin) + int64(denseIdx)*4)
		if err != nil || compPtr == 0 {
			break
		}

		comp, _ := readFn(r.Ctx, uintptr(compPtr))
		if comp != nil {
			results = append(results, comp)
		}

		// Follow nextIndex chain for multiple components of same type
		numNext := (nextEnd - nextBegin) / 4
		if uint32(denseIdx) >= numNext {
			break
		}
		nextIdx, err := r.readS32(int64(nextBegin) + int64(denseIdx)*4)
		if err != nil {
			break
		}
		denseIdx = nextIdx
	}

	return results
}
