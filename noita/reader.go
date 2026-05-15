package noita

//go:generate go run github.com/vitaminmoo/memtools/cmd/hexpatgen -i noita.hexpat -o noita_gen.go -pkg noita

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"path"
	"strings"
	"time"

	"github.com/vitaminmoo/memtools/hexpat/runtime"
)


// CellData stride in the CellFactory material array.
const cellDataStride = 0x290

// Domain identifies a subset of game state that can be read independently.
type Domain int

const (
	DomainStatics          Domain = iota // WorldSeed, NgPlusCount, DeathCount, NumOrbsTotal
	DomainGlobalsAndCamera               // GameGlobals + camera view rect
	DomainWorldState                     // WorldStateComponent + fungal shifts + lua globals + flags + orbs
	DomainPlayerCore                     // player entity + DamageModel + Wallet + Char + Inventory2
	DomainPlayerInventory                // wands, items, wand spell decks (deps PlayerCore)
	DomainPlayerEffects                  // active GameEffectComponents (deps PlayerCore)
	DomainEntities                       // full entity list
)

// AllDomains is the order ReadState() uses for the merged snapshot.
var AllDomains = []Domain{
	DomainStatics,
	DomainGlobalsAndCamera,
	DomainWorldState,
	DomainPlayerCore,
	DomainPlayerInventory,
	DomainPlayerEffects,
	DomainEntities,
}

// String returns a stable name for the domain (for logging / metrics).
func (d Domain) String() string {
	switch d {
	case DomainStatics:
		return "statics"
	case DomainGlobalsAndCamera:
		return "globals"
	case DomainWorldState:
		return "world"
	case DomainPlayerCore:
		return "player"
	case DomainPlayerInventory:
		return "inventory"
	case DomainPlayerEffects:
		return "effects"
	case DomainEntities:
		return "entities"
	}
	return fmt.Sprintf("domain(%d)", int(d))
}

// GameState holds a snapshot of all interesting game data.
type GameState struct {
	Connected bool
	Error     string

	// Domains records the last time each domain was successfully refreshed.
	// nil for snapshots produced before any domain has run.
	Domains map[Domain]time.Time

	WorldSeed    uint32
	NgPlusCount  int32
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

	// Fungal shifts (and any other ConvertMaterialEverywhere calls).
	// Logged in WorldStateComponent.changed_materials, oldest first.
	FungalShifts []FungalShift

	// Persistent Lua globals (GlobalsGetValue / GlobalsSetValue).
	// Includes fungal_shift_iteration, HOLY_MOUNTAIN_DEPTH, perk picks, etc.
	LuaGlobals map[string]string

	// Run-milestone flags from WorldStateComponent.flags.
	Flags []string

	// Orb IDs picked up this run.
	OrbsFoundThisRun []int32

	// Active GameEffectComponents on the player (status effects, perks).
	PlayerEffects []ActiveEffect

	// Spell-card names loaded into each wand (parallel to Wands).
	WandSpells [][]string
}


// FungalShift represents a single ConvertMaterialEverywhere call. Each fungal
// shift script invocation typically produces multiple entries (one per
// converted material in the source group).
type FungalShift struct {
	From string
	To   string
}


// ActiveEffect is a GameEffectComponent attached to (a child of) the player.
type ActiveEffect struct {
	Name           string // entity name of the carrier (often empty)
	CustomEffectID string // populated for effect=CUSTOM perks like PROTECTION_RADIOACTIVITY
	Effect         int32  // GAME_EFFECT enum index
	Frames         int32  // -1 = forever
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

// MaterialInfo is one entry from the CellFactory material table.
type MaterialInfo struct {
	ID            int
	Name          string
	FallbackColor uint32
	TexturePtr    uint32
	TexW          int32
	TexH          int32
	PixelDataPtr  uint32
	Addr          uintptr
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


// bufferMeta caches per-frame metadata for a single ComponentBuffer.
// These fields are buffer-level (not entity-level) and don't change
// between entities within a single tick, so we read them once.
type bufferMeta struct {
	valid       bool
	activeCount int32
	sparseBegin uint32
	sparseEnd   uint32
	compBegin   uint32
	compEnd     uint32
	nextBegin   uint32
	nextEnd     uint32
}

// Reader reads Noita game state from process memory.
type Reader struct {
	proc io.ReadSeeker
	Ctx  *runtime.ReadContext
	// bufCache is populated once per ReadEntityList call and used by
	// FindEntityComponentIDs, hasComponent, readAllComponents, etc.
	// Indexed by TypeID. Nil outside of ReadEntityList.
	bufCache []bufferMeta
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
// ReadFungalShifts decodes WorldStateComponent.changed_materials, the flat
// std::vector<std::string> updated by ConvertMaterialEverywhere. The vector
// stores alternating from/to names — every call appends the pair, so a single
// fungal shift that converts a 3-material source group produces 3 entries.
// Returns shifts in oldest-first order.
// ReadLuaGlobals walks the MSVC red-black tree backing
// WorldStateComponent.lua_globals and returns every {key, value} pair. Used by
// the script API GlobalsGetValue / GlobalsSetValue. Returns nil if empty.
//
// MSVC tree node layout (64 bytes):
//
//	+0x00 _Left   uint32
//	+0x04 _Parent uint32
//	+0x08 _Right  uint32
//	+0x0C _Color  uint8
//	+0x0D _Isnil  uint8
//	+0x10 key     MsvcString (24 bytes)
//	+0x28 value   MsvcString (24 bytes)
//
// The map header points at the head sentinel. head._Parent is the real root.
// Real nodes have _Isnil == 0; nil children point back to the head sentinel.
func (r *Reader) ReadLuaGlobals(m *StdMapHeader) map[string]string {
	if m == nil || m.HeadPtr == 0 || m.Size == 0 {
		return nil
	}
	if m.Size > 8192 {
		return nil // sanity
	}

	// Read head sentinel; root = head._Parent.
	var head [16]byte
	if _, err := r.Ctx.ReadAt(head[:], int64(m.HeadPtr)); err != nil {
		return nil
	}
	root := binary.LittleEndian.Uint32(head[4:8])
	if root == 0 || root == m.HeadPtr {
		return nil
	}

	out := make(map[string]string, m.Size)
	stack := []uint32{root}
	visited := 0
	limit := int(m.Size) * 2 // soft bound to survive races / cycles
	if limit > 16384 {
		limit = 16384
	}

	for len(stack) > 0 && visited < limit {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if node == 0 || node == m.HeadPtr {
			continue
		}

		var raw [16]byte
		if _, err := r.Ctx.ReadAt(raw[:], int64(node)); err != nil {
			continue
		}
		if raw[13] != 0 { // _Isnil
			continue
		}
		left := binary.LittleEndian.Uint32(raw[0:4])
		right := binary.LittleEndian.Uint32(raw[8:12])
		if left != 0 && left != m.HeadPtr {
			stack = append(stack, left)
		}
		if right != 0 && right != m.HeadPtr {
			stack = append(stack, right)
		}

		key, _ := ReadMsvcString(r.Ctx, uintptr(int64(node)+0x10))
		val, _ := ReadMsvcString(r.Ctx, uintptr(int64(node)+0x28))
		if key == nil {
			continue
		}
		k := key.FormatMsvcString(r.Ctx)
		if k == "" {
			continue
		}
		v := ""
		if val != nil {
			v = val.FormatMsvcString(r.Ctx)
		}
		out[k] = v
		visited++
	}
	return out
}

// ReadMsvcStringVector decodes a std::vector<std::string> by reading each
// 24-byte MsvcString in place. Used for run flags, biome tags, etc.
func (r *Reader) ReadMsvcStringVector(vec *StdVectorHeader) []string {
	if vec == nil || vec.BeginPtr == 0 || vec.EndPtr <= vec.BeginPtr {
		return nil
	}
	const stride = 24
	count := int((vec.EndPtr - vec.BeginPtr) / stride)
	if count <= 0 || count > 4096 {
		return nil
	}
	out := make([]string, 0, count)
	for i := 0; i < count; i++ {
		s, _ := ReadMsvcString(r.Ctx, uintptr(int64(vec.BeginPtr)+int64(i*stride)))
		if s == nil {
			out = append(out, "")
			continue
		}
		out = append(out, s.FormatMsvcString(r.Ctx))
	}
	return out
}

// ReadPlayerEffects collects every GameEffectComponent in the player's child
// tree (effects are attached to short-lived child entities). Returns the
// component fields plus the holding entity's name (e.g. "effect_drunk").
func (r *Reader) ReadPlayerEffects(em *EntityManager, player *Entity) []ActiveEffect {
	if em == nil || player == nil {
		return nil
	}
	var out []ActiveEffect

	// Effects can be on the player itself or on its child entities.
	visit := func(e *Entity) {
		if e == nil || e.PendingKill >= 1 {
			return
		}
		effects := readAllComponents[GameEffectComponent](r, em, e.SlotIndex,
			TypeIDGameEffectComponent, ReadGameEffectComponent)
		for _, gec := range effects {
			if gec == nil || !gec.Header.Active {
				continue
			}
			out = append(out, ActiveEffect{
				Name:           e.Name.FormatMsvcString(r.Ctx),
				CustomEffectID: gec.CustomEffectId.FormatMsvcString(r.Ctx),
				Effect:         gec.Effect,
				Frames:         gec.Frames,
			})
		}
	}

	visit(player)
	for _, cp := range r.readChildEntityPtrs(player) {
		if cp == 0 {
			continue
		}
		c, _ := ReadEntity(r.Ctx, uintptr(cp))
		visit(c)
	}
	return out
}

// ReadWandSpellNames returns the spell action_id strings loaded into the given
// wand entity, in slot order. Each spell card is a child entity of the wand
// with an ItemActionComponent; that component's only field (after the standard
// 0x48 header) is action_id (an MsvcString).
func (r *Reader) ReadWandSpellNames(em *EntityManager, wand *Entity) []string {
	if em == nil || wand == nil {
		return nil
	}
	var out []string
	for _, cp := range r.readChildEntityPtrs(wand) {
		if cp == 0 {
			continue
		}
		child, _ := ReadEntity(r.Ctx, uintptr(cp))
		if child == nil || child.PendingKill >= 1 {
			continue
		}
		compPtr := r.findComponentPtr(em, child.SlotIndex, TypeIDItemActionComponent)
		if compPtr == 0 {
			continue
		}
		// action_id MsvcString starts at component_ptr + 0x48 (after the
		// shared ComponentHeader).
		s, _ := ReadMsvcString(r.Ctx, uintptr(int64(compPtr)+0x48))
		if s == nil {
			continue
		}
		out = append(out, s.FormatMsvcString(r.Ctx))
	}
	return out
}

func (r *Reader) ReadFungalShifts(vec *StdVectorHeader) []FungalShift {
	if vec == nil || vec.BeginPtr == 0 || vec.EndPtr <= vec.BeginPtr {
		return nil
	}
	const stride = 24 // sizeof(MsvcString)
	count := int((vec.EndPtr - vec.BeginPtr) / stride)
	if count <= 0 || count > 4096 {
		return nil
	}
	pairs := count / 2
	if pairs == 0 {
		return nil
	}
	out := make([]FungalShift, 0, pairs)
	for i := 0; i < pairs; i++ {
		fromAddr := uintptr(int64(vec.BeginPtr) + int64(i*2*stride))
		toAddr := uintptr(int64(vec.BeginPtr) + int64((i*2+1)*stride))
		from, _ := ReadMsvcString(r.Ctx, fromAddr)
		to, _ := ReadMsvcString(r.Ctx, toAddr)
		shift := FungalShift{}
		if from != nil {
			shift.From = from.FormatMsvcString(r.Ctx)
		}
		if to != nil {
			shift.To = to.FormatMsvcString(r.Ctx)
		}
		out = append(out, shift)
	}
	return out
}

// ReadState reads a complete game-state snapshot by running every domain.
// Equivalent to RunDomains(nil, AllDomains).
func (r *Reader) ReadState() *GameState {
	return r.RunDomains(nil, AllDomains)
}

// RunDomains refreshes the requested domains, carrying over fields from prev
// for domains that aren't being refreshed this call. If prev is nil, returns
// a fresh snapshot.
//
// If DomainStatics is requested and the connect-canary read fails, the
// returned state has Connected=false and remaining domains are skipped.
func (r *Reader) RunDomains(prev *GameState, domains []Domain) *GameState {
	gs := &GameState{Connected: true}
	if prev != nil {
		*gs = *prev
		gs.Connected = true
		gs.Error = ""
	}
	// Copy Domains map so callers' prior snapshots are not mutated.
	domainTimes := make(map[Domain]time.Time, len(gs.Domains)+len(domains))
	for k, v := range gs.Domains {
		domainTimes[k] = v
	}
	gs.Domains = domainTimes

	// Statics first: it carries the connect-canary check.
	requested := domainSet(domains)
	if requested[DomainStatics] {
		if err := r.readStatics(gs); err != nil {
			return gs
		}
		gs.Domains[DomainStatics] = time.Now()
	}

	// Player-related domains share the EntityManager pointer; resolve once.
	var em *EntityManager
	if requested[DomainPlayerCore] || requested[DomainPlayerInventory] || requested[DomainPlayerEffects] {
		em = r.readEM()
	}

	now := time.Now()
	for _, d := range domains {
		switch d {
		case DomainStatics:
			// already handled above
			continue
		case DomainGlobalsAndCamera:
			r.readGlobalsAndCamera(gs)
		case DomainWorldState:
			r.readWorldState(gs)
		case DomainPlayerCore:
			r.readPlayerCore(gs, em)
		case DomainPlayerInventory:
			r.readPlayerInventory(gs, em)
		case DomainPlayerEffects:
			r.readPlayerEffectsDomain(gs, em)
		case DomainEntities:
			gs.Entities = r.ReadEntityList()
		default:
			continue
		}
		gs.Domains[d] = now
	}
	return gs
}

func domainSet(domains []Domain) map[Domain]bool {
	m := make(map[Domain]bool, len(domains))
	for _, d := range domains {
		m[d] = true
	}
	return m
}

// readStatics fills WorldSeed/NgPlusCount/DeathCount/NumOrbsTotal. WorldSeed
// doubles as the connect-canary: if it fails to read, mark disconnected.
func (r *Reader) readStatics(gs *GameState) error {
	v, err := ReadGWorldSeed(r.Ctx)
	if err != nil {
		gs.Error = fmt.Sprintf("read world seed: %v", err)
		gs.Connected = false
		return err
	}
	gs.WorldSeed = v
	gs.NgPlusCount, _ = ReadGNgPlusCount(r.Ctx)
	gs.DeathCount, _ = ReadGDeathCount(r.Ctx)
	gs.NumOrbsTotal, _ = ReadGNumOrbsTotal(r.Ctx)
	return nil
}

func (r *Reader) readGlobalsAndCamera(gs *GameState) {
	gs.Globals, _ = ReadGGameGlobals(r.Ctx)
	if gs.Globals == nil {
		return
	}
	vr, _ := gs.Globals.ReadPWorldManager(r.Ctx)
	if vr == nil {
		return
	}
	gs.ViewW = vr.ViewWidth
	gs.ViewH = vr.ViewHeight
	gs.CameraX = vr.ViewX + vr.ViewWidth*0.5
	gs.CameraY = vr.ViewY + vr.ViewHeight*0.5
}

func (r *Reader) readWorldState(gs *GameState) {
	gs.WorldState, _ = ReadGWorldState(r.Ctx)
	if gs.WorldState == nil {
		gs.FungalShifts = nil
		gs.LuaGlobals = nil
		gs.Flags = nil
		gs.OrbsFoundThisRun = nil
		return
	}
	gs.FungalShifts = r.ReadFungalShifts(&gs.WorldState.ChangedMaterials)
	gs.LuaGlobals = r.ReadLuaGlobals(&gs.WorldState.LuaGlobals)
	gs.Flags = r.ReadMsvcStringVector(&gs.WorldState.Flags)
	if elems := gs.WorldState.OrbsFoundThisrun.Elements; len(elems) > 0 {
		gs.OrbsFoundThisRun = append([]int32(nil), elems...)
	} else {
		gs.OrbsFoundThisRun = nil
	}
}

func (r *Reader) readPlayerCore(gs *GameState, em *EntityManager) {
	gs.PlayerEntity = nil
	gs.PlayerHP = nil
	gs.PlayerWallet = nil
	gs.PlayerChar = nil
	gs.PlayerInv = nil

	dma, _ := ReadGDeathMatchApp(r.Ctx)
	if dma == nil || len(dma.PlayerEntities.Elements) == 0 {
		return
	}
	playerEntityPtr := dma.PlayerEntities.Elements[0]
	if playerEntityPtr == 0 {
		return
	}
	pe, _ := ReadEntity(r.Ctx, uintptr(playerEntityPtr))
	gs.PlayerEntity = pe
	if pe == nil || em == nil {
		return
	}
	gs.PlayerHP = readComponent[DamageModelComponent](r, em, pe.SlotIndex, TypeIDDamageModelComponent, ReadDamageModelComponent)
	gs.PlayerWallet = readComponent[WalletComponent](r, em, pe.SlotIndex, TypeIDWalletComponent, ReadWalletComponent)
	gs.PlayerChar = readComponent[CharacterDataComponent](r, em, pe.SlotIndex, TypeIDCharacterDataComponent, ReadCharacterDataComponent)
	gs.PlayerInv = readComponent[Inventory2Component](r, em, pe.SlotIndex, TypeIDInventory2Component, ReadInventory2Component)
}

func (r *Reader) readPlayerInventory(gs *GameState, em *EntityManager) {
	gs.Wands = nil
	gs.Items = nil
	gs.WandSpells = nil
	if em == nil || gs.PlayerEntity == nil {
		return
	}
	for _, item := range r.findInventoryItems(em, gs.PlayerEntity) {
		if item.IsWand() {
			gs.Wands = append(gs.Wands, item)
		} else {
			gs.Items = append(gs.Items, item)
		}
	}
	if len(gs.Wands) == 0 {
		return
	}
	gs.WandSpells = make([][]string, len(gs.Wands))
	for i, w := range gs.Wands {
		if w == nil || w.Entity == nil {
			continue
		}
		gs.WandSpells[i] = r.ReadWandSpellNames(em, w.Entity)
	}
}

func (r *Reader) readPlayerEffectsDomain(gs *GameState, em *EntityManager) {
	gs.PlayerEffects = nil
	if em == nil || gs.PlayerEntity == nil {
		return
	}
	gs.PlayerEffects = r.ReadPlayerEffects(em, gs.PlayerEntity)
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

// ReadMaterials enumerates the CellFactory material table.
func (r *Reader) ReadMaterials() []*MaterialInfo {
	globals, _ := ReadGGameGlobals(r.Ctx)
	if globals == nil {
		return nil
	}
	cf, _ := globals.ReadPCellFactory(r.Ctx)
	if cf == nil || cf.CellDataArrayPtr == 0 || cf.MaterialCount <= 0 {
		return nil
	}
	count := int(cf.MaterialCount)
	if count > 4096 {
		count = 4096
	}
	out := make([]*MaterialInfo, 0, count)
	for i := 0; i < count; i++ {
		addr := uintptr(cf.CellDataArrayPtr) + uintptr(i)*cellDataStride
		cd, _ := ReadCellData(r.Ctx, addr)
		if cd == nil {
			continue
		}
		mi := &MaterialInfo{
			ID:            i,
			Name:          cd.Name.FormatMsvcString(r.Ctx),
			FallbackColor: cd.FallbackColor,
			TexturePtr:    cd.TexturePtr,
			Addr:          addr,
		}
		if cd.TexturePtr != 0 {
			if tex, _ := ReadCellTexture(r.Ctx, uintptr(cd.TexturePtr)); tex != nil {
				mi.TexW = tex.Width
				mi.TexH = tex.Height
				mi.PixelDataPtr = tex.PixelDataPtr
			}
		}
		out = append(out, mi)
	}
	return out
}

// CellInfo describes a single pixel-world cell lookup result.
type CellInfo struct {
	WorldX, WorldY int32
	ChunkCX        uint32
	ChunkCY        uint32
	ChunkIdx       uint32
	ChunkPtr       uint32 // Chunk* (0 = unloaded/air)
	CellSlotsPtr   uint32 // heap base of 512*512 Cell* slots
	CellIdx        uint32
	CellPtr        uint32 // Cell* (0 = air)
}

// ReadCellAt resolves a world-pixel coordinate to its chunk/cell pointers.
// See noita.hexpat "Chunk System" for the address math.
func (r *Reader) ReadCellAt(wx, wy int32) *CellInfo {
	globals, _ := ReadGGameGlobals(r.Ctx)
	if globals == nil || globals.PChunkSystem == 0 {
		return nil
	}
	// CellGrid is embedded at ChunkSystem+0x500; chunk_table_ptr at +0x08.
	chunkTablePtrAddr := int64(globals.PChunkSystem) + 0x500 + 0x08
	chunkTablePtr, err := r.readU32(chunkTablePtrAddr)
	if err != nil || chunkTablePtr == 0 {
		return nil
	}

	chunkCx := uint32(((wx >> 9) - 0x100)) & 0x1FF
	chunkCy := uint32(((wy >> 9) - 0x100)) & 0x1FF
	chunkIdx := chunkCy*0x200 + chunkCx

	info := &CellInfo{
		WorldX: wx, WorldY: wy,
		ChunkCX: chunkCx, ChunkCY: chunkCy, ChunkIdx: chunkIdx,
	}
	chunkPtr, err := r.readU32(int64(chunkTablePtr) + int64(chunkIdx)*4)
	if err != nil {
		return info
	}
	info.ChunkPtr = chunkPtr
	if chunkPtr == 0 {
		return info
	}
	chunk, _ := ReadChunk(r.Ctx, uintptr(chunkPtr))
	if chunk == nil || chunk.CellSlotsPtr == 0 {
		return info
	}
	info.CellSlotsPtr = chunk.CellSlotsPtr
	info.CellIdx = uint32(wy&0x1FF)*512 + uint32(wx&0x1FF)
	cellPtr, err := r.readU32(int64(chunk.CellSlotsPtr) + int64(info.CellIdx)*4)
	if err == nil {
		info.CellPtr = cellPtr
	}
	return info
}

// ChunkStats summarizes the loaded chunk table.
type ChunkStats struct {
	ChunkTablePtr uint32
	TableEntries  int // always 0x40000
	Loaded        int
	MinCX, MinCY  uint32
	MaxCX, MaxCY  uint32
	Samples       []ChunkSample
}

// ChunkSample is one loaded-chunk record.
type ChunkSample struct {
	CX, CY   uint32
	ChunkPtr uint32
}

// ReadChunkStats scans the chunk table and returns loaded-chunk statistics.
// Up to `sampleLimit` chunk samples are included; pass 0 to disable sampling.
func (r *Reader) ReadChunkStats(sampleLimit int) *ChunkStats {
	globals, _ := ReadGGameGlobals(r.Ctx)
	if globals == nil || globals.PChunkSystem == 0 {
		return nil
	}
	chunkTablePtr, err := r.readU32(int64(globals.PChunkSystem) + 0x500 + 0x08)
	if err != nil || chunkTablePtr == 0 {
		return nil
	}
	const entries = 0x40000
	buf := make([]byte, entries*4)
	if _, err := r.Ctx.ReadAt(buf, int64(chunkTablePtr)); err != nil {
		return nil
	}
	stats := &ChunkStats{
		ChunkTablePtr: chunkTablePtr,
		TableEntries:  entries,
		MinCX:         0x1FF, MinCY: 0x1FF,
	}
	for i := 0; i < entries; i++ {
		p := binary.LittleEndian.Uint32(buf[i*4:])
		if p == 0 {
			continue
		}
		cx := uint32(i) & 0x1FF
		cy := uint32(i) >> 9
		stats.Loaded++
		if cx < stats.MinCX {
			stats.MinCX = cx
		}
		if cx > stats.MaxCX {
			stats.MaxCX = cx
		}
		if cy < stats.MinCY {
			stats.MinCY = cy
		}
		if cy > stats.MaxCY {
			stats.MaxCY = cy
		}
		if len(stats.Samples) < sampleLimit {
			stats.Samples = append(stats.Samples, ChunkSample{CX: cx, CY: cy, ChunkPtr: p})
		}
	}
	if stats.Loaded == 0 {
		stats.MinCX, stats.MinCY = 0, 0
	}
	return stats
}

// =============================================================================
// Biome chunk grid (the wobble lookup) — see noita.hexpat "Biome chunk grid".
// =============================================================================

// BiomeChunkInfo summarizes a single biome chunk's flags and data pointer.
type BiomeChunkInfo struct {
	CX            int32  `json:"cx"`
	CY            int32  `json:"cy"`
	Ptr           uint32 `json:"ptr"`
	Name          string `json:"name"`
	XmlName       string `json:"xmlName"`
	WobbleEligibe bool   `json:"wobbleEligible"`
	WavyEdge      bool   `json:"wavyEdge"`
	ForceOriginal bool   `json:"forceOriginal"`
	BiomeDataPtr  uint32 `json:"biomeDataPtr"`
}

// BiomeGridInfo describes the biome grid header.
type BiomeGridInfo struct {
	Ptr        uint32  `json:"ptr"`
	Width      int32   `json:"width"`
	Height     int32   `json:"height"`
	XShift     float64 `json:"xShift"`
	YShift     float64 `json:"yShift"`
	ChunksPtr  uint32  `json:"chunksPtr"`
	TotalCount int32   `json:"totalCount"`
}

// readBiomeGrid resolves the biome grid pointer from globals.
// Returns nil if the world manager isn't set up yet.
func (r *Reader) readBiomeGrid() (*BiomeGrid, uint32) {
	globals, _ := ReadGGameGlobals(r.Ctx)
	if globals == nil || globals.PWorldManager == 0 {
		return nil, 0
	}
	wm, _ := ReadWorldManagerViewRect(r.Ctx, uintptr(globals.PWorldManager))
	if wm == nil || wm.PBackgroundGrid == 0 {
		return nil, 0
	}
	grid, _ := ReadBiomeGrid(r.Ctx, uintptr(wm.PBackgroundGrid))
	return grid, wm.PBackgroundGrid
}

// ReadBiomeGridInfo returns the grid header (dimensions, shifts, chunks ptr).
func (r *Reader) ReadBiomeGridInfo() *BiomeGridInfo {
	grid, ptr := r.readBiomeGrid()
	if grid == nil {
		return nil
	}
	return &BiomeGridInfo{
		Ptr:        ptr,
		Width:      grid.Width,
		Height:     grid.Height,
		XShift:     grid.XShift,
		YShift:     grid.YShift,
		ChunksPtr:  grid.ChunksPtr,
		TotalCount: grid.TotalCount,
	}
}

// readBiomeChunkAtIdx reads one chunk pointer from the grid array and returns
// its info. Returns nil if the slot is null.
func (r *Reader) readBiomeChunkAtIdx(grid *BiomeGrid, cx, cy int32) *BiomeChunkInfo {
	if grid == nil || grid.Width == 0 || grid.ChunksPtr == 0 {
		return nil
	}
	if cx < 0 || cy < 0 || cx >= grid.Width || cy >= grid.Height {
		return nil
	}
	idx := int64(cy)*int64(grid.Width) + int64(cx)
	chunkPtr, err := r.readU32(int64(grid.ChunksPtr) + idx*4)
	if err != nil || chunkPtr == 0 {
		return nil
	}
	return r.readBiomeChunk(uintptr(chunkPtr), cx, cy)
}

func (r *Reader) readBiomeChunk(addr uintptr, cx, cy int32) *BiomeChunkInfo {
	bc, _ := ReadBiomeChunk(r.Ctx, addr)
	if bc == nil {
		return nil
	}
	return &BiomeChunkInfo{
		CX: cx, CY: cy,
		Ptr:           uint32(addr),
		Name:          bc.BiomeName.FormatMsvcString(r.Ctx),
		XmlName:       stripBiomeXMLPath(bc.XmlPath.FormatMsvcString(r.Ctx)),
		WobbleEligibe: bc.WobbleEligible != 0,
		WavyEdge:      bc.WavyEdge != 0,
		ForceOriginal: bc.ForceOriginal != 0,
		BiomeDataPtr:  bc.BiomeDataPtr,
	}
}

// stripBiomeXMLPath turns "data/biomes/tower_coalmine.xml" into "tower_coalmine".
// Empty input returns empty. Uses path.Base (forward slashes) since Noita
// stores these paths with forward slashes regardless of host OS.
func stripBiomeXMLPath(p string) string {
	if p == "" {
		return ""
	}
	return strings.TrimSuffix(path.Base(p), ".xml")
}

// ReadBiomeChunk returns the chunk at grid coords (cx, cy), or nil if empty.
func (r *Reader) ReadBiomeChunkInfo(cx, cy int32) *BiomeChunkInfo {
	grid, _ := r.readBiomeGrid()
	if grid == nil {
		return nil
	}
	return r.readBiomeChunkAtIdx(grid, cx, cy)
}

// BiomeAtResult is what `biome-at` returns: original chunk + final (post-wobble)
// chunk + the actual decision path the binary took.
type BiomeAtResult struct {
	WX          int32           `json:"wx"`
	WY          int32           `json:"wy"`
	GridWidth   int32           `json:"gridWidth"`
	GridHeight  int32           `json:"gridHeight"`
	XShift      float64         `json:"xShift"`
	YShift      float64         `json:"yShift"`
	OrigCX      int32           `json:"origCX"`
	OrigCY      int32           `json:"origCY"`
	SubX        int32           `json:"subX"`
	SubY        int32           `json:"subY"`
	Original    *BiomeChunkInfo `json:"original,omitempty"`
	Resolved    *BiomeChunkInfo `json:"resolved,omitempty"`
	Wobbled     bool            `json:"wobbled"`
	WobbleType  string          `json:"wobbleType"`
	NeighborDir string          `json:"neighborDir,omitempty"`
	NeighborCX  int32           `json:"neighborCX,omitempty"`
	NeighborCY  int32           `json:"neighborCY,omitempty"`
	WobbleDX    float64         `json:"wobbleDX,omitempty"`
	WobbleDY    float64         `json:"wobbleDY,omitempty"`
}

// ResolveBiomeAt mirrors ChunkGrid_ResolveChunkAtPosition (Noita @ 0x0087d9a0):
// it returns Noita's resolved biome chunk for a world coordinate, including
// the wobble decision (skipped, simplex-only, sin-cos+simplex) and which
// neighbor (if any) triggered the wobble.
func (r *Reader) ResolveBiomeAt(wx, wy int32) *BiomeAtResult {
	grid, _ := r.readBiomeGrid()
	if grid == nil {
		return nil
	}
	res := &BiomeAtResult{
		WX: wx, WY: wy,
		GridWidth: grid.Width, GridHeight: grid.Height,
		XShift: grid.XShift, YShift: grid.YShift,
		WobbleType: "none",
	}

	// Match the binary's coordinate math (uses doubles internally).
	sx := float64(wx) + grid.XShift
	sy := float64(wy) + grid.YShift
	res.SubX = int32(int64(sx) & 0x1FF)
	res.SubY = int32(int64(sy) & 0x1FF)

	cx := int32(int64(sx)>>9) % grid.Width
	if cx < 0 {
		cx += grid.Width
	}
	cy := int32(int64(sy) >> 9)
	if cy >= grid.Height {
		cy = grid.Height - 1
	}
	if cy < 0 {
		cy = 0
	}
	res.OrigCX, res.OrigCY = cx, cy

	orig := r.readBiomeChunkAtIdx(grid, cx, cy)
	res.Original = orig
	res.Resolved = orig
	if orig == nil {
		return res
	}

	// Short-circuits matching the binary's first two checks.
	if !orig.WobbleEligibe || orig.ForceOriginal {
		res.WobbleType = "skipped-flags"
		return res
	}

	// Find the first neighbor with a different chunk pointer, in the same
	// order the binary probes (left, top, right, bottom, then NW/SW only if
	// sub_x<42, then NE/SE in their respective corners).
	type cand struct {
		dir    string
		cx, cy int32
	}
	subX, subY := res.SubX, res.SubY
	var probes []cand
	if subX < 0x2A {
		probes = append(probes, cand{"left", cx - 1, cy})
	}
	if subY < 0x2A {
		probes = append(probes, cand{"top", cx, cy - 1})
	}
	if subX > 0x1D6 {
		probes = append(probes, cand{"right", cx + 1, cy})
	}
	if subY > 0x1D6 {
		probes = append(probes, cand{"bottom", cx, cy + 1})
	}
	if subX < 0x2A {
		if subY < 0x2A {
			probes = append(probes, cand{"top-left", cx - 1, cy - 1})
		}
		if subY > 0x1D6 {
			probes = append(probes, cand{"bottom-left", cx - 1, cy + 1})
		}
	}
	if subX > 0x1D6 {
		if subY < 0x2A {
			probes = append(probes, cand{"top-right", cx + 1, cy - 1})
		}
		if subY > 0x1D6 {
			probes = append(probes, cand{"bottom-right", cx + 1, cy + 1})
		}
	}

	var neighbor *BiomeChunkInfo
	for _, c := range probes {
		ncx, ncy := wrapCX(c.cx, grid.Width), clampCY(c.cy, grid.Height)
		ne := r.readBiomeChunkAtIdx(grid, ncx, ncy)
		if ne == nil || ne.Ptr == orig.Ptr {
			continue
		}
		neighbor = ne
		res.NeighborDir = c.dir
		res.NeighborCX, res.NeighborCY = ncx, ncy
		break
	}
	if neighbor == nil {
		res.WobbleType = "no-differing-neighbor"
		return res
	}
	if !neighbor.WobbleEligibe {
		res.WobbleType = "skipped-flags"
		return res
	}

	// Wobble math from ChunkGrid_ResolveChunkAtPosition (noita.exe @
	// 0x0087d9a0; this branch starts at LAB_0087dc01 inside it).
	//   simplex_val = simplex(sx*0.05, sy*0.05) * 70
	//   simplex-only (either side wavy_edge=0):
	//       dx = dy = simplex_val * 2.5
	//   sin-cos + simplex (both wavy):
	//       dx = cos(sx*0.005)*30 + simplex_val*11
	//       dy = sin(sy*0.005)*30 + simplex_val*11
	var dx, dy float64
	simplexVal := computeSimplex2D(sx*0.05, sy*0.05) * 70.0
	if !orig.WavyEdge || !neighbor.WavyEdge {
		dx = simplexVal * 2.5
		dy = dx
		res.WobbleType = "simplex-only"
	} else {
		dx = math.Cos(sx*0.005)*30.0 + simplexVal*11.0
		dy = math.Sin(sy*0.005)*30.0 + simplexVal*11.0
		res.WobbleType = "sin-cos+simplex"
	}
	res.WobbleDX = dx
	res.WobbleDY = dy
	res.Wobbled = true

	// The binary swaps X/Y wobble in the chunk lookup — X chunk index uses
	// (Y_wobble + sx), Y uses (X_wobble + sy). See
	// `BiomeGrid_GetChunkAt (noita.exe @ 0x0087d870):
	// (grid, (dVar10+dVar13)>>9, (dVar12+dVar9)>>9)` where dVar10 is the
	// sin/y branch and dVar12 is the cos/x branch.
	wcx := wrapCX(int32(int64(math.Floor(sx+dy))>>9), grid.Width)
	wcy := clampCY(int32(int64(math.Floor(sy+dx))>>9), grid.Height)
	wobbled := r.readBiomeChunkAtIdx(grid, wcx, wcy)
	// Only adopt the wobbled chunk if it exists and is wobble-eligible.
	if wobbled != nil && wobbled.WobbleEligibe {
		res.Resolved = wobbled
	}
	return res
}

// computeSimplex2D is the 2D simplex noise helper from
// ChunkGrid_ResolveChunkAtPosition (noita.exe @ 0x0087d9a0). Returned
// value is the raw simplex contribution before amplitude scaling.
func computeSimplex2D(x, y float64) float64 {
	sqrt312 := (math.Sqrt(3) - 1) / 2
	sqrt336 := (3 - math.Sqrt(3)) / 6

	// edge noise tables (the static 256-byte permutation + the *0xC mod
	// table). Initialized lazily on first call.
	if edgeNoise2 == nil {
		edgeNoise2 = make([]int32, 512)
		edgeNoiseM12 = make([]int32, 512)
		for i := 0; i < 512; i++ {
			t := int32(edgeNoiseRaw[i&0xff])
			edgeNoise2[i] = t
			edgeNoiseM12[i] = t % 0xC
		}
	}

	dVar7 := (x + y) * sqrt312
	dVar6a := dVar7 + x
	uVar2 := uint32(dVar6a)
	if dVar6a < float64(uVar2) {
		uVar2--
	}
	dVar7 = dVar7 + y
	uVar1 := uint32(dVar7)
	if dVar7 < float64(uVar1) {
		uVar1--
	}
	dVar6 := float64(uVar1+uVar2) * sqrt336
	dVar10 := x - (float64(uVar2) - dVar6)
	dVar9 := y - (float64(uVar1) - dVar6)
	uVar1 &= 0xff
	uVar2 &= 0xff
	bIfDVar9LtDVar10 := int32(0)
	if dVar9 < dVar10 {
		bIfDVar9LtDVar10 = 1
	}
	bIfDVar10LeDVar9 := int32(0)
	if dVar10 <= dVar9 {
		bIfDVar10LeDVar9 = 1
	}
	dVar7b := (dVar10 - float64(bIfDVar9LtDVar10)) + sqrt336
	dVar3 := (dVar9 - float64(bIfDVar10LeDVar9)) + sqrt336
	dVar11 := (dVar10 - 1) + sqrt336*2
	dVar4 := (dVar9 - 1) + sqrt336*2

	dVar5 := 0.5 - dVar10*dVar10 - dVar9*dVar9
	var part1, part2, part3 float64
	if dVar5 >= 0 {
		idx := edgeNoiseM12[edgeNoise2[uVar1]+int32(uVar2)] * 4
		part1 = (float64(edgeSigns[idx+1])*dVar9 + float64(edgeSigns[idx])*dVar10) *
			dVar5 * dVar5 * dVar5 * dVar5
	}
	dVar5 = 0.5 - dVar7b*dVar7b - dVar3*dVar3
	if dVar5 >= 0 {
		idx := edgeNoiseM12[edgeNoise2[int32(uVar1)+bIfDVar10LeDVar9]+int32(uVar2)+bIfDVar9LtDVar10] * 4
		part2 = (float64(edgeSigns[idx+1])*dVar3 + float64(edgeSigns[idx])*dVar7b) *
			dVar5 * dVar5 * dVar5 * dVar5
	}
	dVar3a := 0.5 - dVar11*dVar11 - dVar4*dVar4
	if dVar3a >= 0 {
		idx := edgeNoiseM12[edgeNoise2[int32(uVar1)+1]+int32(uVar2)+1] * 4
		part3 = (float64(edgeSigns[idx+1])*dVar4 + float64(edgeSigns[idx])*dVar11) *
			dVar3a * dVar3a * dVar3a * dVar3a
	}
	return part1 + part2 + part3
}

var (
	edgeNoise2   []int32
	edgeNoiseM12 []int32
)

var edgeNoiseRaw = [256]uint8{
	0x97, 0xa0, 0x89, 0x5b, 0x5a, 0x0f, 0x83, 0x0d, 0xc9, 0x5f, 0x60, 0x35, 0xc2, 0xe9, 0x07, 0xe1,
	0x8c, 0x24, 0x67, 0x1e, 0x45, 0x8e, 0x08, 0x63, 0x25, 0xf0, 0x15, 0x0a, 0x17, 0xbe, 0x06, 0x94,
	0xf7, 0x78, 0xea, 0x4b, 0x00, 0x1a, 0xc5, 0x3e, 0x5e, 0xfc, 0xdb, 0xcb, 0x75, 0x23, 0x0b, 0x20,
	0x39, 0xb1, 0x21, 0x58, 0xed, 0x95, 0x38, 0x57, 0xae, 0x14, 0x7d, 0x88, 0xab, 0xa8, 0x44, 0xaf,
	0x4a, 0xa5, 0x47, 0x86, 0x8b, 0x30, 0x1b, 0xa6, 0x4d, 0x92, 0x9e, 0xe7, 0x53, 0x6f, 0xe5, 0x7a,
	0x3c, 0xd3, 0x85, 0xe6, 0xdc, 0x69, 0x5c, 0x29, 0x37, 0x2e, 0xf5, 0x28, 0xf4, 0x66, 0x8f, 0x36,
	0x41, 0x19, 0x3f, 0xa1, 0x01, 0xd8, 0x50, 0x49, 0xd1, 0x4c, 0x84, 0xbb, 0xd0, 0x59, 0x12, 0xa9,
	0xc8, 0xc4, 0x87, 0x82, 0x74, 0xbc, 0x9f, 0x56, 0xa4, 0x64, 0x6d, 0xc6, 0xad, 0xba, 0x03, 0x40,
	0x34, 0xd9, 0xe2, 0xfa, 0x7c, 0x7b, 0x05, 0xca, 0x26, 0x93, 0x76, 0x7e, 0xff, 0x52, 0x55, 0xd4,
	0xcf, 0xce, 0x3b, 0xe3, 0x2f, 0x10, 0x3a, 0x11, 0xb6, 0xbd, 0x1c, 0x2a, 0xdf, 0xb7, 0xaa, 0xd5,
	0x77, 0xf8, 0x98, 0x02, 0x2c, 0x9a, 0xa3, 0x46, 0xdd, 0x99, 0x65, 0x9b, 0xa7, 0x2b, 0xac, 0x09,
	0x81, 0x16, 0x27, 0xfd, 0x13, 0x62, 0x6c, 0x6e, 0x4f, 0x71, 0xe0, 0xe8, 0xb2, 0xb9, 0x70, 0x68,
	0xda, 0xf6, 0x61, 0xe4, 0xfb, 0x22, 0xf2, 0xc1, 0xee, 0xd2, 0x90, 0x0c, 0xbf, 0xb3, 0xa2, 0xf1,
	0x51, 0x33, 0x91, 0xeb, 0xf9, 0x0e, 0xef, 0x6b, 0x31, 0xc0, 0xd6, 0x1f, 0xb5, 0xc7, 0x6a, 0x9d,
	0xb8, 0x54, 0xcc, 0xb0, 0x73, 0x79, 0x32, 0x2d, 0x7f, 0x04, 0x96, 0xfe, 0x8a, 0xec, 0xcd, 0x5d,
	0xde, 0x72, 0x43, 0x1d, 0x18, 0x48, 0xf3, 0x8d, 0x80, 0xc3, 0x4e, 0x42, 0xd7, 0x3d, 0x9c, 0xb4,
}

var edgeSigns = [48]int32{
	1, 1, 0, 0, -1, 1, 0, 0, 1, -1, 0, 0, -1, -1, 0, 0,
	1, 0, 1, 0, -1, 0, 1, 0, 1, 0, -1, 0, -1, 0, -1, 0,
	0, 1, 1, 0, 0, -1, 1, 0, 0, 1, -1, 0, 0, -1, -1, 0,
}

// PixelSceneInfo is one queued/placed pixel scene from noita's runtime.
type PixelSceneInfo struct {
	X                  int32  `json:"x"`
	Y                  int32  `json:"y"`
	MaterialsFile      string `json:"materialsFile"`
	ColorsFile         string `json:"colorsFile"`
	BackgroundFile     string `json:"backgroundFile,omitempty"`
	FlagSkipBiomeCheck bool   `json:"flagSkipBiomeChecks,omitempty"`
	FlagSkipEdgeTex    bool   `json:"flagSkipEdgeTextures,omitempty"`
	Index              int32  `json:"index"`
	Vec                string `json:"vec"` // "main" or "alt"
}

// iteratePixelSceneVec walks one of the two PixelSceneEntry vectors that
// live inside BiomeGrid.
func (r *Reader) iteratePixelSceneVec(begin, end uint32, vecName string, fn func(*PixelSceneInfo) bool) {
	if begin == 0 || end <= begin {
		return
	}
	const stride = 0x90
	count := int((end - begin) / stride)
	if count <= 0 || count > 1<<20 {
		return
	}
	buf := make([]byte, count*stride)
	if _, err := r.Ctx.ReadAt(buf, int64(begin)); err != nil {
		return
	}
	for i := 0; i < count; i++ {
		base := i * stride
		x := int32(binary.LittleEndian.Uint32(buf[base+0x04:]))
		y := int32(binary.LittleEndian.Uint32(buf[base+0x08:]))
		mat, _ := ReadMsvcString(r.Ctx, uintptr(int64(begin)+int64(base)+0x0C))
		col, _ := ReadMsvcString(r.Ctx, uintptr(int64(begin)+int64(base)+0x24))
		bg, _ := ReadMsvcString(r.Ctx, uintptr(int64(begin)+int64(base)+0x3C))
		info := &PixelSceneInfo{
			X: x, Y: y,
			Index: int32(i),
			Vec:   vecName,
		}
		if mat != nil {
			info.MaterialsFile = mat.FormatMsvcString(r.Ctx)
		}
		if col != nil {
			info.ColorsFile = col.FormatMsvcString(r.Ctx)
		}
		if bg != nil {
			info.BackgroundFile = bg.FormatMsvcString(r.Ctx)
		}
		info.FlagSkipBiomeCheck = buf[base+0x58] != 0
		info.FlagSkipEdgeTex = buf[base+0x59] != 0
		if !fn(info) {
			return
		}
	}
}

// IteratePixelScenes walks both pixel-scene vectors in BiomeGrid (the "main"
// queue used by procedural Lua placements and the "alt" / background queue).
// fn may return false to stop iteration.
func (r *Reader) IteratePixelScenes(fn func(*PixelSceneInfo) bool) {
	grid, _ := r.readBiomeGrid()
	if grid == nil {
		return
	}
	stop := false
	r.iteratePixelSceneVec(grid.ScenesBegin, grid.ScenesEnd, "main", func(p *PixelSceneInfo) bool {
		if !fn(p) {
			stop = true
			return false
		}
		return true
	})
	if stop {
		return
	}
	r.iteratePixelSceneVec(grid.ScenesAltBegin, grid.ScenesAltEnd, "alt", fn)
}

// IterateBiomeChunks calls fn for every non-null chunk in the biome grid.
// fn receives the chunk info and may return false to stop iteration.
func (r *Reader) IterateBiomeChunks(fn func(*BiomeChunkInfo) bool) {
	grid, _ := r.readBiomeGrid()
	if grid == nil || grid.ChunksPtr == 0 || grid.Width == 0 {
		return
	}
	total := int(grid.Width) * int(grid.Height)
	if total <= 0 {
		return
	}
	buf := make([]byte, total*4)
	if _, err := r.Ctx.ReadAt(buf, int64(grid.ChunksPtr)); err != nil {
		return
	}
	for i := 0; i < total; i++ {
		p := binary.LittleEndian.Uint32(buf[i*4:])
		if p == 0 {
			continue
		}
		cx := int32(i % int(grid.Width))
		cy := int32(i / int(grid.Width))
		info := r.readBiomeChunk(uintptr(p), cx, cy)
		if info == nil {
			continue
		}
		if !fn(info) {
			return
		}
	}
}

func wrapCX(cx, width int32) int32 {
	if width <= 0 {
		return 0
	}
	cx = cx % width
	if cx < 0 {
		cx += width
	}
	return cx
}

func clampCY(cy, height int32) int32 {
	if height <= 0 {
		return 0
	}
	if cy < 0 {
		return 0
	}
	if cy > height-1 {
		return height - 1
	}
	return cy
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

// buildBufferCache bulk-reads metadata for all component buffers in one
// syscall each. The fields cached (ActiveCount, SparseIndex, Components,
// NextIndex begin/end pointers) are buffer-level, not entity-level, so
// they're identical for every entity within a single tick.
func (r *Reader) buildBufferCache(em *EntityManager) {
	n := len(em.ComponentBuffers.Elements)
	r.bufCache = make([]bufferMeta, n)
	for i, bufPtr := range em.ComponentBuffers.Elements {
		if bufPtr == 0 {
			continue
		}
		// Read offsets 16..155 (140 bytes) covering SparseIndex through ActiveCount.
		var raw [140]byte
		if _, err := r.Ctx.ReadAt(raw[:], int64(bufPtr)+16); err != nil {
			continue
		}
		// Offsets relative to raw (raw[0] = ComponentBuffer offset 16):
		//   SparseIndex.BeginPtr = CB+16  → raw[0]
		//   SparseIndex.EndPtr   = CB+20  → raw[4]
		//   NextIndex.BeginPtr   = CB+52  → raw[36]
		//   NextIndex.EndPtr     = CB+56  → raw[40]
		//   Components.BeginPtr  = CB+64  → raw[48]
		//   Components.EndPtr    = CB+68  → raw[52]
		//   ActiveCount          = CB+152 → raw[136]
		r.bufCache[i] = bufferMeta{
			valid:       true,
			sparseBegin: binary.LittleEndian.Uint32(raw[0:]),
			sparseEnd:   binary.LittleEndian.Uint32(raw[4:]),
			nextBegin:   binary.LittleEndian.Uint32(raw[36:]),
			nextEnd:     binary.LittleEndian.Uint32(raw[40:]),
			compBegin:   binary.LittleEndian.Uint32(raw[48:]),
			compEnd:     binary.LittleEndian.Uint32(raw[52:]),
			activeCount: int32(binary.LittleEndian.Uint32(raw[136:])),
		}
	}
}

// cachedMeta returns the cached buffer metadata for a type ID, or an
// invalid entry if no cache is available or the typeID is out of range.
func (r *Reader) cachedMeta(em *EntityManager, typeID TypeID) bufferMeta {
	if r.bufCache != nil && int(typeID) < len(r.bufCache) {
		return r.bufCache[int(typeID)]
	}
	// Fallback: no cache, do live reads (for callers outside ReadEntityList).
	bufferPtr := lookupComponentBufferPtr(em, typeID)
	if bufferPtr == 0 {
		return bufferMeta{}
	}
	var raw [140]byte
	if _, err := r.Ctx.ReadAt(raw[:], int64(bufferPtr)+16); err != nil {
		return bufferMeta{}
	}
	return bufferMeta{
		valid:       true,
		sparseBegin: binary.LittleEndian.Uint32(raw[0:]),
		sparseEnd:   binary.LittleEndian.Uint32(raw[4:]),
		nextBegin:   binary.LittleEndian.Uint32(raw[36:]),
		nextEnd:     binary.LittleEndian.Uint32(raw[40:]),
		compBegin:   binary.LittleEndian.Uint32(raw[48:]),
		compEnd:     binary.LittleEndian.Uint32(raw[52:]),
		activeCount: int32(binary.LittleEndian.Uint32(raw[136:])),
	}
}

// findComponentPtr resolves a component pointer using cached ComponentBuffer metadata.
func (r *Reader) findComponentPtr(em *EntityManager, slotIndex int32, typeID TypeID) uint32 {
	meta := r.cachedMeta(em, typeID)
	if !meta.valid {
		return 0
	}
	numSparse := (meta.sparseEnd - meta.sparseBegin) / 4
	if uint32(slotIndex) >= numSparse {
		return 0
	}
	denseIdx, err := r.readS32(int64(meta.sparseBegin) + int64(slotIndex)*4)
	if err != nil || denseIdx < 0 {
		return 0
	}
	numComps := (meta.compEnd - meta.compBegin) / 4
	if uint32(denseIdx) >= numComps {
		return 0
	}
	compPtr, _ := r.readU32(int64(meta.compBegin) + int64(denseIdx)*4)
	return compPtr
}

// hasComponent checks if an entity has a component type using cached metadata.
func (r *Reader) hasComponent(em *EntityManager, slotIndex int32, typeID TypeID) bool {
	meta := r.cachedMeta(em, typeID)
	if !meta.valid {
		return false
	}
	numSparse := (meta.sparseEnd - meta.sparseBegin) / 4
	if uint32(slotIndex) >= numSparse {
		return false
	}
	denseIdx, err := r.readS32(int64(meta.sparseBegin) + int64(slotIndex)*4)
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

// entityHeaderSize is the on-the-wire size of the Entity struct (matches
// ReadEntity's buf[152]byte). All header fields live within this window.
const entityHeaderSize = 152

// ReadEntityList reads all entities from the EntityManager and returns
// summaries. The entity-pointer table and all entity-header reads are
// coalesced into two batched syscalls; per-entity component lookups still
// use the in-frame bufCache.
func (r *Reader) ReadEntityList() []*EntitySummary {
	em := r.readEM()
	if em == nil {
		return nil
	}

	// Cache component buffer metadata for the duration of this call.
	r.buildBufferCache(em)
	defer func() { r.bufCache = nil }()

	count := (em.EntityArray.EndPtr - em.EntityArray.BeginPtr) / 4
	if count == 0 || count > 100000 {
		return nil
	}

	// Syscall 1: entity-pointer table.
	ptrBuf := make([]byte, count*4)
	if _, err := r.Ctx.ReadAt(ptrBuf, int64(em.EntityArray.BeginPtr)); err != nil {
		return nil
	}

	// Syscall 2 (chunked): batch all entity headers. Each header is 152 bytes,
	// so 1000 entities = 152 KB across a single process_vm_readv (with one
	// follow-up if we exceed UIO_MAXIOV iovecs).
	type entRef struct {
		ptr uint32
		buf []byte
	}
	refs := make([]entRef, 0, count)
	coll := runtime.NewCollector(r.Ctx)
	for i := uint32(0); i < count; i++ {
		ePtr := binary.LittleEndian.Uint32(ptrBuf[i*4 : i*4+4])
		if ePtr == 0 {
			continue
		}
		refs = append(refs, entRef{ptr: ePtr, buf: coll.Add(uintptr(ePtr), entityHeaderSize)})
	}
	if err := coll.Flush(); err != nil {
		// Batch failed (rare; usually a transient unmapped page). Fall back to
		// per-entity reads so we still return whatever's currently readable.
		ptrs := make([]uint32, len(refs))
		for i, ref := range refs {
			ptrs[i] = ref.ptr
		}
		return r.readEntityListPerEntity(em, ptrs)
	}

	summaries := make([]*EntitySummary, 0, len(refs))
	for _, ref := range refs {
		ent := decodeEntityHeader(ref.buf)
		if ent.PendingKill >= 1 {
			continue
		}
		summaries = append(summaries, r.buildEntitySummary(em, ref.ptr, ent))
	}
	return summaries
}

// readEntityListPerEntity is the fallback used when the batched header read
// fails; it reproduces the original per-entity ReadAt path.
func (r *Reader) readEntityListPerEntity(em *EntityManager, ptrs []uint32) []*EntitySummary {
	summaries := make([]*EntitySummary, 0, len(ptrs))
	for _, p := range ptrs {
		ent, _ := ReadEntity(r.Ctx, uintptr(p))
		if ent == nil || ent.PendingKill >= 1 {
			continue
		}
		summaries = append(summaries, r.buildEntitySummary(em, p, ent))
	}
	return summaries
}

func (r *Reader) buildEntitySummary(em *EntityManager, ePtr uint32, ent *Entity) *EntitySummary {
	return &EntitySummary{
		Entity:           ent,
		Name:             ent.Name.FormatMsvcString(r.Ctx),
		Ptr:              ePtr,
		HasHP:            r.hasComponent(em, ent.SlotIndex, TypeIDDamageModelComponent),
		HasWallet:        r.hasComponent(em, ent.SlotIndex, TypeIDWalletComponent),
		HasAbility:       r.hasComponent(em, ent.SlotIndex, TypeIDAbilityComponent),
		HasCharData:      r.hasComponent(em, ent.SlotIndex, TypeIDCharacterDataComponent),
		Hitbox:           readComponent[HitboxComponent](r, em, ent.SlotIndex, TypeIDHitboxComponent, ReadHitboxComponent),
		CollisionTrigger: readComponent[CollisionTriggerComponent](r, em, ent.SlotIndex, TypeIDCollisionTriggerComponent, ReadCollisionTriggerComponent),
		Sprite:           readComponent[SpriteComponent](r, em, ent.SlotIndex, TypeIDSpriteComponent, ReadSpriteComponent),
		Lua:              readComponent[LuaComponent](r, em, ent.SlotIndex, TypeIDLuaComponent, ReadLuaComponent),
		Item:             readComponent[ItemComponent](r, em, ent.SlotIndex, TypeIDItemComponent, ReadItemComponent),
		Contents:         r.readPotionContents(em, ent.SlotIndex),
		ComponentIDs:     r.FindEntityComponentIDs(em, ent.SlotIndex),
	}
}

// decodeEntityHeader mirrors ReadEntity's decode against a pre-fetched
// 152-byte buffer. The Name MsvcString header is extracted from the buffer
// directly; FormatMsvcString resolves heap-allocated string content on
// demand (and most entity names are inline so need no further reads).
func decodeEntityHeader(buf []byte) *Entity {
	e := &Entity{
		EntityId:    int32(binary.LittleEndian.Uint32(buf[0:])),
		SlotIndex:   int32(binary.LittleEndian.Uint32(buf[4:])),
		Unknown08:   binary.LittleEndian.Uint32(buf[8:]),
		PendingKill: int32(binary.LittleEndian.Uint32(buf[12:])),
		Flags10:     binary.LittleEndian.Uint32(buf[16:]),
	}
	copy(e.Name.Data[:], buf[20:36])
	e.Name.Length = binary.LittleEndian.Uint32(buf[36:])
	e.Name.Capacity = binary.LittleEndian.Uint32(buf[40:])
	e.Unknown2c = binary.LittleEndian.Uint32(buf[44:])
	copy(e.TagBitset[:], buf[48:112])
	e.PosX = math.Float32frombits(binary.LittleEndian.Uint32(buf[112:]))
	e.PosY = math.Float32frombits(binary.LittleEndian.Uint32(buf[116:]))
	e.RotCos = math.Float32frombits(binary.LittleEndian.Uint32(buf[120:]))
	e.RotSin = math.Float32frombits(binary.LittleEndian.Uint32(buf[124:]))
	e.RotNegSin = math.Float32frombits(binary.LittleEndian.Uint32(buf[128:]))
	e.RotCos2 = math.Float32frombits(binary.LittleEndian.Uint32(buf[132:]))
	e.ScaleX = math.Float32frombits(binary.LittleEndian.Uint32(buf[136:]))
	e.ScaleY = math.Float32frombits(binary.LittleEndian.Uint32(buf[140:]))
	e.ChildrenPtr = binary.LittleEndian.Uint32(buf[144:])
	e.ParentEntityPtr = binary.LittleEndian.Uint32(buf[148:])
	return e
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
// Uses bufCache when available (within ReadEntityList) to avoid repeated
// syscalls for buffer metadata that doesn't change between entities.
func (r *Reader) FindEntityComponentIDs(em *EntityManager, slotIndex int32) []TypeID {
	if slotIndex < 0 || em == nil {
		return nil
	}

	var ids []TypeID
	for typeID := range em.ComponentBuffers.Elements {
		meta := r.cachedMeta(em, TypeID(typeID))
		if !meta.valid || meta.activeCount == 0 {
			continue
		}
		numSparse := (meta.sparseEnd - meta.sparseBegin) / 4
		if uint32(slotIndex) >= numSparse {
			continue
		}
		denseIdx, err := r.readS32(int64(meta.sparseBegin) + int64(slotIndex)*4)
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
// Uses cached buffer metadata when available.
func readAllComponents[T any](r *Reader, em *EntityManager, slotIndex int32, typeID TypeID, readFn readFunc[T]) []*T {
	if slotIndex < 0 {
		return nil
	}

	meta := r.cachedMeta(em, typeID)
	if !meta.valid {
		return nil
	}

	numSparse := (meta.sparseEnd - meta.sparseBegin) / 4
	if uint32(slotIndex) >= numSparse {
		return nil
	}
	denseIdx, err := r.readS32(int64(meta.sparseBegin) + int64(slotIndex)*4)
	if err != nil || denseIdx < 0 {
		return nil
	}

	var results []*T

	for denseIdx >= 0 {
		numComponents := (meta.compEnd - meta.compBegin) / 4
		if uint32(denseIdx) >= numComponents {
			break
		}

		compPtr, err := r.readU32(int64(meta.compBegin) + int64(denseIdx)*4)
		if err != nil || compPtr == 0 {
			break
		}

		comp, _ := readFn(r.Ctx, uintptr(compPtr))
		if comp != nil {
			results = append(results, comp)
		}

		// Follow nextIndex chain for multiple components of same type
		numNext := (meta.nextEnd - meta.nextBegin) / 4
		if uint32(denseIdx) >= numNext {
			break
		}
		nextIdx, err := r.readS32(int64(meta.nextBegin) + int64(denseIdx)*4)
		if err != nil {
			break
		}
		denseIdx = nextIdx
	}

	return results
}
