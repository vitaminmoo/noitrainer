package noita

//go:generate go run github.com/vitaminmoo/memtools/cmd/hexpatgen -i noita.hexpat -o noita_gen.go -pkg noita

import (
	"encoding/binary"
	"fmt"
	"io"

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

// Component type IDs (runtime-assigned, resolved via PTypeName from live process).
const (
	TypeAbilityComponent                    = 3
	TypeAnimalAIComponent                   = 6
	TypeAudioComponent                      = 10
	TypeAudioListenerComponent              = 11
	TypeAudioLoopComponent                  = 12
	TypeBiomeTrackerComponent               = 13
	TypeCameraBoundComponent                = 18
	TypeCharacterCollisionComponent         = 21
	TypeCharacterDataComponent              = 22
	TypeCharacterPlatformingComponent       = 23
	TypeCollisionTriggerComponent           = 25
	TypeControlsComponent                   = 28
	TypeDamageModelComponent                = 31
	TypeDrugEffectComponent                 = 38
	TypeElectricityReceiverComponent        = 42
	TypeExplodeOnDamageComponent            = 46
	TypeGameEffectComponent                 = 53
	TypeGameLogComponent                    = 54
	TypeGameStatsComponent                  = 55
	TypeGenomeDataComponent                 = 57
	TypeGunComponent                        = 60
	TypeHitboxComponent                     = 63
	TypeHotspotComponent                    = 65
	TypeIngestionComponent                  = 70
	TypeInheritTransformComponent           = 71
	TypeInteractableComponent               = 72
	TypeInventory2Component                 = 73
	TypeInventoryGuiComponent               = 75
	TypeItemActionComponent                 = 77
	TypeItemAlchemyComponent                = 78
	TypeItemChestComponent                  = 79
	TypeItemComponent                       = 80
	TypeItemCostComponent                   = 81
	TypeItemPickUpperComponent              = 82
	TypeKickComponent                       = 85
	TypeLifetimeComponent                   = 88
	TypeLightComponent                      = 89
	TypeLiquidDisplacerComponent            = 92
	TypeLuaComponent                        = 96
	TypeManaReloaderComponent               = 99
	TypeMaterialAreaCheckerComponent        = 100
	TypeMaterialInventoryComponent          = 101
	TypeMaterialSuckerComponent             = 103
	TypeMusicEnergyAffectorComponent        = 105
	TypeParticleEmitterComponent            = 109
	TypePathFindingComponent                = 110
	TypePathFindingGridMarkerComponent      = 111
	TypePhysicsBody2Component               = 113
	TypePhysicsBodyCollisionDamageComponent = 114
	TypePhysicsBodyComponent                = 115
	TypePhysicsImageShapeComponent          = 116
	TypePhysicsJointComponent               = 119
	TypePhysicsPickUpComponent              = 121
	TypePhysicsThrowableComponent           = 124
	TypePixelSpriteComponent                = 126
	TypePlatformShooterPlayerComponent      = 127
	TypePlayerCollisionComponent            = 128
	TypePlayerStatsComponent                = 129
	TypePositionSeedComponent               = 130
	TypePotionComponent                     = 131
	TypeProjectileComponent                 = 133
	TypeSimplePhysicsComponent              = 138
	TypeSpriteAnimatorComponent             = 140
	TypeSpriteComponent                     = 141
	TypeSpriteOffsetAnimatorComponent       = 142
	TypeSpriteParticleEmitterComponent      = 143
	TypeSpriteStainsComponent               = 144
	TypeStatusEffectDataComponent           = 145
	TypeStreamingKeepAliveComponent         = 146
	TypeTorchComponent                      = 151
	TypeUIInfoComponent                     = 153
	TypeVariableStorageComponent            = 154
	TypeVelocityComponent                   = 155
	TypeVerletPhysicsComponent              = 156
	TypeVerletWorldJointComponent           = 158
	TypeWalletComponent                     = 159
	TypeWorldStateComponent                 = 161
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
	Entity       *Entity
	Name         string
	Ptr          uint32
	HasHP        bool
	HasWallet    bool
	HasAbility   bool
	HasCharData  bool
	ComponentIDs []int // all component type IDs present on this entity
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
	Children []*EntitySummary
}

// ComponentBufferInfo holds metadata about a component buffer (type).
type ComponentBufferInfo struct {
	TypeID      int
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
		n := s.Length
		if n > 16 {
			n = 16
		}
		return string(s.Data[:n])
	}
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

// readEM reads the EntityManager via its static pointer.
func (r *Reader) readEM() *EntityManager {
	emPtr, err := r.readU32(AddrEntityManager)
	if err != nil || emPtr == 0 {
		return nil
	}
	em, _ := ReadEntityManager(r.Ctx, uintptr(emPtr))
	return em
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

		// Read camera from WorldManager view rect
		if gs.Globals != nil && gs.Globals.PWorldManager != 0 {
			if vr, _ := ReadWorldManagerViewRect(r.Ctx, uintptr(gs.Globals.PWorldManager)); vr != nil {
				gs.ViewW = vr.ViewWidth
				gs.ViewH = vr.ViewHeight
				gs.CameraX = vr.ViewX + vr.ViewWidth*0.5
				gs.CameraY = vr.ViewY + vr.ViewHeight*0.5
			}
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
	gs.PlayerHP = readComponent[DamageModelComponent](r, em, gs.PlayerEntity.SlotIndex, TypeDamageModelComponent, ReadDamageModelComponent)
	gs.PlayerWallet = readComponent[WalletComponent](r, em, gs.PlayerEntity.SlotIndex, TypeWalletComponent, ReadWalletComponent)
	gs.PlayerChar = readComponent[CharacterDataComponent](r, em, gs.PlayerEntity.SlotIndex, TypeCharacterDataComponent, ReadCharacterDataComponent)
	gs.PlayerInv = readComponent[Inventory2Component](r, em, gs.PlayerEntity.SlotIndex, TypeInventory2Component, ReadInventory2Component)

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
	globalsPtr, err := r.readU32(AddrGameGlobals)
	if err != nil || globalsPtr == 0 {
		return fmt.Sprintf("mat_%d", matID)
	}
	globals, _ := ReadGameGlobals(r.Ctx, uintptr(globalsPtr))
	if globals == nil || globals.PCellFactory == 0 {
		return fmt.Sprintf("mat_%d", matID)
	}
	cf, _ := ReadCellFactory(r.Ctx, uintptr(globals.PCellFactory))
	if cf == nil || cf.CellDataArrayPtr == 0 {
		return fmt.Sprintf("mat_%d", matID)
	}
	addr := uintptr(cf.CellDataArrayPtr) + uintptr(matID)*cellDataStride
	cd, _ := ReadCellData(r.Ctx, addr)
	if cd == nil {
		return fmt.Sprintf("mat_%d", matID)
	}
	name := MsvcStringValue(&cd.Name, r.Ctx)
	if name == "" {
		return fmt.Sprintf("mat_%d", matID)
	}
	return name
}

// readPotionContents reads MaterialInventoryComponent for an entity and returns non-zero materials.
func (r *Reader) readPotionContents(em *EntityManager, slotIndex int32) []MaterialContent {
	mic := readComponent[MaterialInventoryComponent](r, em, slotIndex, TypeMaterialInventoryComponent, ReadMaterialInventoryComponent)
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
func lookupComponentBufferPtr(em *EntityManager, typeID int) uint32 {
	if em == nil || typeID >= len(em.ComponentBuffers.Elements) {
		return 0
	}
	return em.ComponentBuffers.Elements[typeID]
}

// findComponentPtr resolves a component pointer using lazy ComponentBuffer reads.
func (r *Reader) findComponentPtr(em *EntityManager, slotIndex int32, typeID int) uint32 {
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
func (r *Reader) hasComponent(em *EntityManager, slotIndex int32, typeID int) bool {
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
func (r *Reader) readChildEntityPtrs(childrenPtr uint32) []uint32 {
	if childrenPtr == 0 {
		return nil
	}
	cc, _ := ReadChildrenContainer(r.Ctx, uintptr(childrenPtr))
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

	childPtrs := r.readChildEntityPtrs(player.ChildrenPtr)
	for _, cp := range childPtrs {
		if cp == 0 {
			continue
		}
		child, _ := ReadEntity(r.Ctx, uintptr(cp))
		if child == nil || child.PendingKill >= 1 {
			continue
		}

		if ac := readComponent[AbilityComponent](r, em, child.SlotIndex, TypeAbilityComponent, ReadAbilityComponent); ac != nil {
			item := &InventoryItem{Entity: child, Ability: ac}
			if !item.IsWand() {
				item.Contents = r.readPotionContents(em, child.SlotIndex)
			}
			items = append(items, item)
		}

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
		nameStr, _ := er.Name().Read()

		// Build a minimal Entity for the summary (avoids full eager read)
		entity := &Entity{
			EntityId:  entityId,
			SlotIndex: slotIndex,
			PosX:      posX,
			PosY:      posY,
		}
		if nameStr != nil {
			entity.Name = *nameStr
		}

		name := MsvcStringValue(&entity.Name, r.Ctx)
		compIDs := r.FindEntityComponentIDs(em, slotIndex)
		summary := &EntitySummary{
			Entity:       entity,
			Name:         name,
			Ptr:          ePtr,
			HasHP:        r.hasComponent(em, slotIndex, TypeDamageModelComponent),
			HasWallet:    r.hasComponent(em, slotIndex, TypeWalletComponent),
			HasAbility:   r.hasComponent(em, slotIndex, TypeAbilityComponent),
			HasCharData:  r.hasComponent(em, slotIndex, TypeCharacterDataComponent),
			ComponentIDs: compIDs,
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
		Entity:  e,
		Name:    MsvcStringValue(&e.Name, r.Ctx),
		HP:      readComponent[DamageModelComponent](r, em, e.SlotIndex, TypeDamageModelComponent, ReadDamageModelComponent),
		Wallet:  readComponent[WalletComponent](r, em, e.SlotIndex, TypeWalletComponent, ReadWalletComponent),
		Char:    readComponent[CharacterDataComponent](r, em, e.SlotIndex, TypeCharacterDataComponent, ReadCharacterDataComponent),
		Inv:     readComponent[Inventory2Component](r, em, e.SlotIndex, TypeInventory2Component, ReadInventory2Component),
		Ability: readComponent[AbilityComponent](r, em, e.SlotIndex, TypeAbilityComponent, ReadAbilityComponent),
	}

	// Read children
	childPtrs := r.readChildEntityPtrs(e.ChildrenPtr)
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
			Name:   MsvcStringValue(&child.Name, r.Ctx),
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
			name = MsvcStringValue(nameMs, r.Ctx)
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
			TypeID:      i,
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
	emPtr, err := r.readU32(AddrEntityManager)
	if err != nil || emPtr == 0 {
		return nil, 0
	}
	em, _ := ReadEntityManager(r.Ctx, uintptr(emPtr))
	return em, emPtr
}

// FindEntityComponentIDs returns all component type IDs that an entity has.
func (r *Reader) FindEntityComponentIDs(em *EntityManager, slotIndex int32) []int {
	if slotIndex < 0 || em == nil {
		return nil
	}

	var ids []int
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
		ids = append(ids, typeID)
	}
	return ids
}

// ReadRawComponent reads raw bytes for an arbitrary component type at a given entity slot.
func (r *Reader) ReadRawComponent(em *EntityManager, slotIndex int32, typeID int, size int) (uint32, []byte) {
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
