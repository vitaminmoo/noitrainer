package noita

import (
	"encoding/binary"
	"github.com/vitaminmoo/memtools/hexpatgen/runtime"
	"math"
)

type MsvcString struct {
	Data     [16]byte
	Length   uint32
	Capacity uint32
}

type Entity struct {
	EntityId        int32
	SlotIndex       int32
	Unknown08       uint32
	PendingKill     int32
	Flags10         uint32
	Name            MsvcString
	Unknown2c       uint32
	TagBitset       [64]uint8
	PosX            float32
	PosY            float32
	RotCos          float32
	RotSin          float32
	RotNegSin       float32
	RotCos2         float32
	ScaleX          float32
	ScaleY          float32
	ChildrenPtr     uint32
	ParentEntityPtr uint32
}

type ComponentHeader struct {
	Vtable        uint32
	BufferIndex   int32
	PTypeName     uint32
	TypeId        int32
	Unknown10     uint32
	Active        bool
	ComponentTags [32]uint8
	Unknown38     uint32
	Unknown3c     uint32
	Unknown40     uint32
	Unknown44     uint32
}

type WalletComponent struct {
	Header         ComponentHeader
	Money          int64
	MoneySpent     int64
	MoneyPrevFrame int64
	HasReachedInf  bool
}

type CharacterDataComponent struct {
	Header     ComponentHeader
	Gravity    float32
	FlyTime    float32
	IsOnGround bool
	VelocityX  float32
	VelocityY  float32
}

type ConfigGun struct {
	Vtable               uint32
	ActionsPerRound      int32
	ShuffleDeckWhenEmpty bool
	ReloadTime           int32
	DeckCapacity         int32
}

type StdVectorHeader struct {
	BeginPtr    uint32
	EndPtr      uint32
	CapacityPtr uint32
}

type EntityManager struct {
	Vtable           uint32
	NextEntityId     int32
	FreeSlotStack    StdVectorHeader
	EntityArray      StdVectorHeader
	TagGroups        StdVectorHeader
	ComponentBuffers StdVectorHeader
	PEventManager    uint32
}

type ComponentBuffer struct {
	Vtable           uint32
	Sentinel         int32
	InitialCapacity  int32
	Unknown0c        uint32
	SparseIndex      StdVectorHeader
	EntityRefs       StdVectorHeader
	PrevIndex        StdVectorHeader
	NextIndex        StdVectorHeader
	Components       StdVectorHeader
	HandleMap        StdVectorHeader
	Generations      StdVectorHeader
	ReverseHandleMap StdVectorHeader
	ActiveCount      int32
	CapacityLimit    int32
	UnknownA0        uint32
	PEntityManager   uint32
	PEventManager    uint32
	NameString       MsvcString
}

type DamageModelComponent struct {
	Header              ComponentHeader
	Hp                  float64
	MaxHp               float64
	MaxHpCap            float64
	MaxHpOld            float64
	Unknown68           uint32
	DmgMultMelee        float32
	DmgMultProjectile   float32
	DmgMultExplosion    float32
	DmgMultElectricity  float32
	DmgMultFire         float32
	DmgMultDrill        float32
	DmgMultSlice        float32
	DmgMultIce          float32
	DmgMultHealing      float32
	DmgMultPhysicsHit   float32
	DmgMultRadioactive  float32
	DmgMultPoison       float32
	DmgMultHoly         float32
	DmgMultCurse        float32
	DmgMultOvereating   float32
	DmgMultMaterial     float32
	InvincibilityFrames int32
}

type Inventory2Component struct {
	Header               ComponentHeader
	QuickInventorySlots  int32
	FullInventorySlotsX  int32
	FullInventorySlotsY  int32
	SavedActiveItemIndex int32
	ActiveItem           int32
	ActualActiveItem     int32
	ActiveStash          int32
	ThrowItem            int32
	ItemHolstered        bool
	Initialized          bool
	ForceRefresh         bool
	DontLogNextItemEquip bool
	SmoothedItemXOffset  float32
	LastItemSwitchFrame  int32
	IntroEquipItemLerp   float32
	SmoothedItemAngleX   float32
	SmoothedItemAngleY   float32
}

type WorldManagerViewRect struct {
	ViewX      float32
	ViewY      float32
	ViewWidth  float32
	ViewHeight float32
}

type DeathMatchApp struct {
	PlayerEntities StdVectorHeader
}

type WorldStateComponent struct {
	Header           ComponentHeader
	BiomeCryptCount  int32
	GodsAfraid       int32
	GodsImpressed    int32
	GodsAfraidDamage int32
	GodsEnraged      int32
}

type ChildrenContainer struct {
	BeginPtr    uint32
	EndPtr      uint32
	CapacityPtr uint32
}

type AbilityComponent struct {
	Header                             ComponentHeader
	CooldownFrames                     int32
	EntityFile                         MsvcString
	SpriteFile                         MsvcString
	EntityCount                        int32
	NeverReload                        bool
	ReloadTimeFrames                   int32
	Mana                               float32
	ManaMax                            float32
	ManaChargeSpeed                    float32
	RotateInHand                       bool
	RotateInHandAmount                 float32
	RotateHandAmount                   float32
	FastProjectile                     bool
	SwimPropelAmount                   float32
	MaxChargedActions                  int32
	ChargeWaitFrames                   int32
	ItemRecoilRecoverySpeed            float32
	ItemRecoilMax                      float32
	ItemRecoilOffsetCoeff              float32
	ItemRecoilRotationCoeff            float32
	BaseItemFile                       MsvcString
	UseEntityFileAsProjectileInfoProxy bool
	ClickToUse                         bool
	StatTimesPlayerHasShot             int32
	StatTimesPlayerHasEdited           int32
	ShootingReducesAmountInInventory   bool
	ThrowAsItem                        bool
	SimulateThrowAsItem                bool
	MaxAmountInInventory               int32
	AmountInInventory                  int32
	DropAsItemOnDeath                  bool
	UiName                             MsvcString
	UseGunScript                       bool
	IsPetrisGun                        bool
	GunConfig                          ConfigGun
	GunLevel                           int32
	AddTheseChildActions               MsvcString
	CurrentSlotDurability              int32
	SlotConsumptionFunction            MsvcString
	NextFrameUsable                    int32
	CastDelayStartFrame                int32
	AmmoLeft                           int32
	ReloadFramesLeft                   int32
	ReloadNextFrameUsable              int32
	ChargeCount                        int32
	NextChargeFrame                    int32
	ItemRecoil                         float32
	IsInitialized                      bool
}

type GameGlobals struct {
	FrameCount       int32
	PhysicsStepCount int32
	GameTime         float32
	PWorldManager    uint32
	PChunkSystem     uint32
	PCellGrid        uint32
	PCellFactory     uint32
	Unknown1c        uint32
	PPhysicsWorld    uint32
	PAudioManager    uint32
	ViewportLeft     float32
	ViewportTop      float32
	ViewportRight    float32
	ViewportBottom   float32
}

func ReadMsvcString(ctx *runtime.ReadContext, addr uintptr) (*MsvcString, runtime.Errors) {
	var errs runtime.Errors
	result := &MsvcString{}
	var buf [4]byte

	// Field: Data (array[16]) at offset 0
	if _, err := ctx.ReadAt(result.Data[:], int64(addr)+0); err != nil {
		errs.Add("MsvcString.Data", uintptr(int64(addr)+0), err)
	}

	// Field: Length at offset 16
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+16); err != nil {
		errs.Add("MsvcString.Length", uintptr(int64(addr)+16), err)
	} else {
		result.Length = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: Capacity at offset 20
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+20); err != nil {
		errs.Add("MsvcString.Capacity", uintptr(int64(addr)+20), err)
	} else {
		result.Capacity = binary.LittleEndian.Uint32(buf[:4])
	}

	return result, errs
}

func ReadEntity(ctx *runtime.ReadContext, addr uintptr) (*Entity, runtime.Errors) {
	var errs runtime.Errors
	result := &Entity{}
	var buf [4]byte

	// Field: EntityId at offset 0
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+0); err != nil {
		errs.Add("Entity.EntityId", uintptr(int64(addr)+0), err)
	} else {
		result.EntityId = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: SlotIndex at offset 4
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+4); err != nil {
		errs.Add("Entity.SlotIndex", uintptr(int64(addr)+4), err)
	} else {
		result.SlotIndex = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: Unknown08 at offset 8
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+8); err != nil {
		errs.Add("Entity.Unknown08", uintptr(int64(addr)+8), err)
	} else {
		result.Unknown08 = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: PendingKill at offset 12
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+12); err != nil {
		errs.Add("Entity.PendingKill", uintptr(int64(addr)+12), err)
	} else {
		result.PendingKill = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: Flags10 at offset 16
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+16); err != nil {
		errs.Add("Entity.Flags10", uintptr(int64(addr)+16), err)
	} else {
		result.Flags10 = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: Name at offset 20
	{
		child, childErrs := ReadMsvcString(ctx, uintptr(int64(addr)+20))
		if child != nil {
			result.Name = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: Unknown2c at offset 44
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+44); err != nil {
		errs.Add("Entity.Unknown2c", uintptr(int64(addr)+44), err)
	} else {
		result.Unknown2c = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: TagBitset (array[64]) at offset 48
	if _, err := ctx.ReadAt(result.TagBitset[:], int64(addr)+48); err != nil {
		errs.Add("Entity.TagBitset", uintptr(int64(addr)+48), err)
	}

	// Field: PosX at offset 112
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+112); err != nil {
		errs.Add("Entity.PosX", uintptr(int64(addr)+112), err)
	} else {
		result.PosX = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: PosY at offset 116
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+116); err != nil {
		errs.Add("Entity.PosY", uintptr(int64(addr)+116), err)
	} else {
		result.PosY = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: RotCos at offset 120
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+120); err != nil {
		errs.Add("Entity.RotCos", uintptr(int64(addr)+120), err)
	} else {
		result.RotCos = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: RotSin at offset 124
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+124); err != nil {
		errs.Add("Entity.RotSin", uintptr(int64(addr)+124), err)
	} else {
		result.RotSin = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: RotNegSin at offset 128
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+128); err != nil {
		errs.Add("Entity.RotNegSin", uintptr(int64(addr)+128), err)
	} else {
		result.RotNegSin = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: RotCos2 at offset 132
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+132); err != nil {
		errs.Add("Entity.RotCos2", uintptr(int64(addr)+132), err)
	} else {
		result.RotCos2 = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ScaleX at offset 136
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+136); err != nil {
		errs.Add("Entity.ScaleX", uintptr(int64(addr)+136), err)
	} else {
		result.ScaleX = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ScaleY at offset 140
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+140); err != nil {
		errs.Add("Entity.ScaleY", uintptr(int64(addr)+140), err)
	} else {
		result.ScaleY = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ChildrenPtr at offset 144
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+144); err != nil {
		errs.Add("Entity.ChildrenPtr", uintptr(int64(addr)+144), err)
	} else {
		result.ChildrenPtr = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: ParentEntityPtr at offset 148
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+148); err != nil {
		errs.Add("Entity.ParentEntityPtr", uintptr(int64(addr)+148), err)
	} else {
		result.ParentEntityPtr = binary.LittleEndian.Uint32(buf[:4])
	}

	return result, errs
}

func ReadComponentHeader(ctx *runtime.ReadContext, addr uintptr) (*ComponentHeader, runtime.Errors) {
	var errs runtime.Errors
	result := &ComponentHeader{}
	var buf [4]byte

	// Field: Vtable at offset 0
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+0); err != nil {
		errs.Add("ComponentHeader.Vtable", uintptr(int64(addr)+0), err)
	} else {
		result.Vtable = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: BufferIndex at offset 4
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+4); err != nil {
		errs.Add("ComponentHeader.BufferIndex", uintptr(int64(addr)+4), err)
	} else {
		result.BufferIndex = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: PTypeName at offset 8
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+8); err != nil {
		errs.Add("ComponentHeader.PTypeName", uintptr(int64(addr)+8), err)
	} else {
		result.PTypeName = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: TypeId at offset 12
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+12); err != nil {
		errs.Add("ComponentHeader.TypeId", uintptr(int64(addr)+12), err)
	} else {
		result.TypeId = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: Unknown10 at offset 16
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+16); err != nil {
		errs.Add("ComponentHeader.Unknown10", uintptr(int64(addr)+16), err)
	} else {
		result.Unknown10 = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: Active at offset 20
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+20); err != nil {
		errs.Add("ComponentHeader.Active", uintptr(int64(addr)+20), err)
	} else {
		result.Active = buf[0] != 0
	}

	// Field: ComponentTags (array[32]) at offset 24
	if _, err := ctx.ReadAt(result.ComponentTags[:], int64(addr)+24); err != nil {
		errs.Add("ComponentHeader.ComponentTags", uintptr(int64(addr)+24), err)
	}

	// Field: Unknown38 at offset 56
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+56); err != nil {
		errs.Add("ComponentHeader.Unknown38", uintptr(int64(addr)+56), err)
	} else {
		result.Unknown38 = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: Unknown3c at offset 60
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+60); err != nil {
		errs.Add("ComponentHeader.Unknown3c", uintptr(int64(addr)+60), err)
	} else {
		result.Unknown3c = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: Unknown40 at offset 64
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+64); err != nil {
		errs.Add("ComponentHeader.Unknown40", uintptr(int64(addr)+64), err)
	} else {
		result.Unknown40 = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: Unknown44 at offset 68
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+68); err != nil {
		errs.Add("ComponentHeader.Unknown44", uintptr(int64(addr)+68), err)
	} else {
		result.Unknown44 = binary.LittleEndian.Uint32(buf[:4])
	}

	return result, errs
}

func ReadWalletComponent(ctx *runtime.ReadContext, addr uintptr) (*WalletComponent, runtime.Errors) {
	var errs runtime.Errors
	result := &WalletComponent{}
	var buf [8]byte

	// Field: Header at offset 0
	{
		child, childErrs := ReadComponentHeader(ctx, uintptr(int64(addr)+0))
		if child != nil {
			result.Header = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: Money at offset 72
	if _, err := ctx.ReadAt(buf[:8], int64(addr)+72); err != nil {
		errs.Add("WalletComponent.Money", uintptr(int64(addr)+72), err)
	} else {
		result.Money = int64(binary.LittleEndian.Uint64(buf[:8]))
	}

	// Field: MoneySpent at offset 80
	if _, err := ctx.ReadAt(buf[:8], int64(addr)+80); err != nil {
		errs.Add("WalletComponent.MoneySpent", uintptr(int64(addr)+80), err)
	} else {
		result.MoneySpent = int64(binary.LittleEndian.Uint64(buf[:8]))
	}

	// Field: MoneyPrevFrame at offset 88
	if _, err := ctx.ReadAt(buf[:8], int64(addr)+88); err != nil {
		errs.Add("WalletComponent.MoneyPrevFrame", uintptr(int64(addr)+88), err)
	} else {
		result.MoneyPrevFrame = int64(binary.LittleEndian.Uint64(buf[:8]))
	}

	// Field: HasReachedInf at offset 96
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+96); err != nil {
		errs.Add("WalletComponent.HasReachedInf", uintptr(int64(addr)+96), err)
	} else {
		result.HasReachedInf = buf[0] != 0
	}

	return result, errs
}

func ReadCharacterDataComponent(ctx *runtime.ReadContext, addr uintptr) (*CharacterDataComponent, runtime.Errors) {
	var errs runtime.Errors
	result := &CharacterDataComponent{}
	var buf [4]byte

	// Field: Header at offset 0
	{
		child, childErrs := ReadComponentHeader(ctx, uintptr(int64(addr)+0))
		if child != nil {
			result.Header = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: Gravity at offset 136
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+136); err != nil {
		errs.Add("CharacterDataComponent.Gravity", uintptr(int64(addr)+136), err)
	} else {
		result.Gravity = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: FlyTime at offset 140
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+140); err != nil {
		errs.Add("CharacterDataComponent.FlyTime", uintptr(int64(addr)+140), err)
	} else {
		result.FlyTime = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: IsOnGround at offset 184
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+184); err != nil {
		errs.Add("CharacterDataComponent.IsOnGround", uintptr(int64(addr)+184), err)
	} else {
		result.IsOnGround = buf[0] != 0
	}

	// Field: VelocityX at offset 264
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+264); err != nil {
		errs.Add("CharacterDataComponent.VelocityX", uintptr(int64(addr)+264), err)
	} else {
		result.VelocityX = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: VelocityY at offset 268
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+268); err != nil {
		errs.Add("CharacterDataComponent.VelocityY", uintptr(int64(addr)+268), err)
	} else {
		result.VelocityY = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	return result, errs
}

func ReadConfigGun(ctx *runtime.ReadContext, addr uintptr) (*ConfigGun, runtime.Errors) {
	var errs runtime.Errors
	result := &ConfigGun{}
	var buf [4]byte

	// Field: Vtable at offset 0
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+0); err != nil {
		errs.Add("ConfigGun.Vtable", uintptr(int64(addr)+0), err)
	} else {
		result.Vtable = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: ActionsPerRound at offset 4
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+4); err != nil {
		errs.Add("ConfigGun.ActionsPerRound", uintptr(int64(addr)+4), err)
	} else {
		result.ActionsPerRound = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ShuffleDeckWhenEmpty at offset 8
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+8); err != nil {
		errs.Add("ConfigGun.ShuffleDeckWhenEmpty", uintptr(int64(addr)+8), err)
	} else {
		result.ShuffleDeckWhenEmpty = buf[0] != 0
	}

	// Field: ReloadTime at offset 12
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+12); err != nil {
		errs.Add("ConfigGun.ReloadTime", uintptr(int64(addr)+12), err)
	} else {
		result.ReloadTime = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: DeckCapacity at offset 16
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+16); err != nil {
		errs.Add("ConfigGun.DeckCapacity", uintptr(int64(addr)+16), err)
	} else {
		result.DeckCapacity = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	return result, errs
}

func ReadStdVectorHeader(ctx *runtime.ReadContext, addr uintptr) (*StdVectorHeader, runtime.Errors) {
	var errs runtime.Errors
	result := &StdVectorHeader{}
	var buf [4]byte

	// Field: BeginPtr at offset 0
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+0); err != nil {
		errs.Add("StdVectorHeader.BeginPtr", uintptr(int64(addr)+0), err)
	} else {
		result.BeginPtr = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: EndPtr at offset 4
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+4); err != nil {
		errs.Add("StdVectorHeader.EndPtr", uintptr(int64(addr)+4), err)
	} else {
		result.EndPtr = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: CapacityPtr at offset 8
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+8); err != nil {
		errs.Add("StdVectorHeader.CapacityPtr", uintptr(int64(addr)+8), err)
	} else {
		result.CapacityPtr = binary.LittleEndian.Uint32(buf[:4])
	}

	return result, errs
}

func ReadEntityManager(ctx *runtime.ReadContext, addr uintptr) (*EntityManager, runtime.Errors) {
	var errs runtime.Errors
	result := &EntityManager{}
	var buf [4]byte

	// Field: Vtable at offset 0
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+0); err != nil {
		errs.Add("EntityManager.Vtable", uintptr(int64(addr)+0), err)
	} else {
		result.Vtable = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: NextEntityId at offset 4
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+4); err != nil {
		errs.Add("EntityManager.NextEntityId", uintptr(int64(addr)+4), err)
	} else {
		result.NextEntityId = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: FreeSlotStack at offset 8
	{
		child, childErrs := ReadStdVectorHeader(ctx, uintptr(int64(addr)+8))
		if child != nil {
			result.FreeSlotStack = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: EntityArray at offset 20
	{
		child, childErrs := ReadStdVectorHeader(ctx, uintptr(int64(addr)+20))
		if child != nil {
			result.EntityArray = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: TagGroups at offset 32
	{
		child, childErrs := ReadStdVectorHeader(ctx, uintptr(int64(addr)+32))
		if child != nil {
			result.TagGroups = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: ComponentBuffers at offset 44
	{
		child, childErrs := ReadStdVectorHeader(ctx, uintptr(int64(addr)+44))
		if child != nil {
			result.ComponentBuffers = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: PEventManager at offset 56
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+56); err != nil {
		errs.Add("EntityManager.PEventManager", uintptr(int64(addr)+56), err)
	} else {
		result.PEventManager = binary.LittleEndian.Uint32(buf[:4])
	}

	return result, errs
}

func ReadComponentBuffer(ctx *runtime.ReadContext, addr uintptr) (*ComponentBuffer, runtime.Errors) {
	var errs runtime.Errors
	result := &ComponentBuffer{}
	var buf [4]byte

	// Field: Vtable at offset 0
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+0); err != nil {
		errs.Add("ComponentBuffer.Vtable", uintptr(int64(addr)+0), err)
	} else {
		result.Vtable = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: Sentinel at offset 4
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+4); err != nil {
		errs.Add("ComponentBuffer.Sentinel", uintptr(int64(addr)+4), err)
	} else {
		result.Sentinel = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: InitialCapacity at offset 8
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+8); err != nil {
		errs.Add("ComponentBuffer.InitialCapacity", uintptr(int64(addr)+8), err)
	} else {
		result.InitialCapacity = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: Unknown0c at offset 12
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+12); err != nil {
		errs.Add("ComponentBuffer.Unknown0c", uintptr(int64(addr)+12), err)
	} else {
		result.Unknown0c = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: SparseIndex at offset 16
	{
		child, childErrs := ReadStdVectorHeader(ctx, uintptr(int64(addr)+16))
		if child != nil {
			result.SparseIndex = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: EntityRefs at offset 28
	{
		child, childErrs := ReadStdVectorHeader(ctx, uintptr(int64(addr)+28))
		if child != nil {
			result.EntityRefs = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: PrevIndex at offset 40
	{
		child, childErrs := ReadStdVectorHeader(ctx, uintptr(int64(addr)+40))
		if child != nil {
			result.PrevIndex = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: NextIndex at offset 52
	{
		child, childErrs := ReadStdVectorHeader(ctx, uintptr(int64(addr)+52))
		if child != nil {
			result.NextIndex = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: Components at offset 64
	{
		child, childErrs := ReadStdVectorHeader(ctx, uintptr(int64(addr)+64))
		if child != nil {
			result.Components = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: HandleMap at offset 96
	{
		child, childErrs := ReadStdVectorHeader(ctx, uintptr(int64(addr)+96))
		if child != nil {
			result.HandleMap = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: Generations at offset 108
	{
		child, childErrs := ReadStdVectorHeader(ctx, uintptr(int64(addr)+108))
		if child != nil {
			result.Generations = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: ReverseHandleMap at offset 120
	{
		child, childErrs := ReadStdVectorHeader(ctx, uintptr(int64(addr)+120))
		if child != nil {
			result.ReverseHandleMap = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: ActiveCount at offset 152
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+152); err != nil {
		errs.Add("ComponentBuffer.ActiveCount", uintptr(int64(addr)+152), err)
	} else {
		result.ActiveCount = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: CapacityLimit at offset 156
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+156); err != nil {
		errs.Add("ComponentBuffer.CapacityLimit", uintptr(int64(addr)+156), err)
	} else {
		result.CapacityLimit = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: UnknownA0 at offset 160
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+160); err != nil {
		errs.Add("ComponentBuffer.UnknownA0", uintptr(int64(addr)+160), err)
	} else {
		result.UnknownA0 = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: PEntityManager at offset 164
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+164); err != nil {
		errs.Add("ComponentBuffer.PEntityManager", uintptr(int64(addr)+164), err)
	} else {
		result.PEntityManager = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: PEventManager at offset 168
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+168); err != nil {
		errs.Add("ComponentBuffer.PEventManager", uintptr(int64(addr)+168), err)
	} else {
		result.PEventManager = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: NameString at offset 172
	{
		child, childErrs := ReadMsvcString(ctx, uintptr(int64(addr)+172))
		if child != nil {
			result.NameString = *child
		}
		errs = append(errs, childErrs...)
	}

	return result, errs
}

func ReadDamageModelComponent(ctx *runtime.ReadContext, addr uintptr) (*DamageModelComponent, runtime.Errors) {
	var errs runtime.Errors
	result := &DamageModelComponent{}
	var buf [8]byte

	// Field: Header at offset 0
	{
		child, childErrs := ReadComponentHeader(ctx, uintptr(int64(addr)+0))
		if child != nil {
			result.Header = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: Hp at offset 72
	if _, err := ctx.ReadAt(buf[:8], int64(addr)+72); err != nil {
		errs.Add("DamageModelComponent.Hp", uintptr(int64(addr)+72), err)
	} else {
		result.Hp = math.Float64frombits(binary.LittleEndian.Uint64(buf[:8]))
	}

	// Field: MaxHp at offset 80
	if _, err := ctx.ReadAt(buf[:8], int64(addr)+80); err != nil {
		errs.Add("DamageModelComponent.MaxHp", uintptr(int64(addr)+80), err)
	} else {
		result.MaxHp = math.Float64frombits(binary.LittleEndian.Uint64(buf[:8]))
	}

	// Field: MaxHpCap at offset 88
	if _, err := ctx.ReadAt(buf[:8], int64(addr)+88); err != nil {
		errs.Add("DamageModelComponent.MaxHpCap", uintptr(int64(addr)+88), err)
	} else {
		result.MaxHpCap = math.Float64frombits(binary.LittleEndian.Uint64(buf[:8]))
	}

	// Field: MaxHpOld at offset 96
	if _, err := ctx.ReadAt(buf[:8], int64(addr)+96); err != nil {
		errs.Add("DamageModelComponent.MaxHpOld", uintptr(int64(addr)+96), err)
	} else {
		result.MaxHpOld = math.Float64frombits(binary.LittleEndian.Uint64(buf[:8]))
	}

	// Field: Unknown68 at offset 104
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+104); err != nil {
		errs.Add("DamageModelComponent.Unknown68", uintptr(int64(addr)+104), err)
	} else {
		result.Unknown68 = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: DmgMultMelee at offset 108
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+108); err != nil {
		errs.Add("DamageModelComponent.DmgMultMelee", uintptr(int64(addr)+108), err)
	} else {
		result.DmgMultMelee = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: DmgMultProjectile at offset 112
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+112); err != nil {
		errs.Add("DamageModelComponent.DmgMultProjectile", uintptr(int64(addr)+112), err)
	} else {
		result.DmgMultProjectile = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: DmgMultExplosion at offset 116
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+116); err != nil {
		errs.Add("DamageModelComponent.DmgMultExplosion", uintptr(int64(addr)+116), err)
	} else {
		result.DmgMultExplosion = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: DmgMultElectricity at offset 120
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+120); err != nil {
		errs.Add("DamageModelComponent.DmgMultElectricity", uintptr(int64(addr)+120), err)
	} else {
		result.DmgMultElectricity = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: DmgMultFire at offset 124
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+124); err != nil {
		errs.Add("DamageModelComponent.DmgMultFire", uintptr(int64(addr)+124), err)
	} else {
		result.DmgMultFire = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: DmgMultDrill at offset 128
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+128); err != nil {
		errs.Add("DamageModelComponent.DmgMultDrill", uintptr(int64(addr)+128), err)
	} else {
		result.DmgMultDrill = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: DmgMultSlice at offset 132
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+132); err != nil {
		errs.Add("DamageModelComponent.DmgMultSlice", uintptr(int64(addr)+132), err)
	} else {
		result.DmgMultSlice = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: DmgMultIce at offset 136
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+136); err != nil {
		errs.Add("DamageModelComponent.DmgMultIce", uintptr(int64(addr)+136), err)
	} else {
		result.DmgMultIce = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: DmgMultHealing at offset 140
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+140); err != nil {
		errs.Add("DamageModelComponent.DmgMultHealing", uintptr(int64(addr)+140), err)
	} else {
		result.DmgMultHealing = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: DmgMultPhysicsHit at offset 144
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+144); err != nil {
		errs.Add("DamageModelComponent.DmgMultPhysicsHit", uintptr(int64(addr)+144), err)
	} else {
		result.DmgMultPhysicsHit = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: DmgMultRadioactive at offset 148
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+148); err != nil {
		errs.Add("DamageModelComponent.DmgMultRadioactive", uintptr(int64(addr)+148), err)
	} else {
		result.DmgMultRadioactive = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: DmgMultPoison at offset 152
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+152); err != nil {
		errs.Add("DamageModelComponent.DmgMultPoison", uintptr(int64(addr)+152), err)
	} else {
		result.DmgMultPoison = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: DmgMultHoly at offset 156
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+156); err != nil {
		errs.Add("DamageModelComponent.DmgMultHoly", uintptr(int64(addr)+156), err)
	} else {
		result.DmgMultHoly = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: DmgMultCurse at offset 160
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+160); err != nil {
		errs.Add("DamageModelComponent.DmgMultCurse", uintptr(int64(addr)+160), err)
	} else {
		result.DmgMultCurse = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: DmgMultOvereating at offset 164
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+164); err != nil {
		errs.Add("DamageModelComponent.DmgMultOvereating", uintptr(int64(addr)+164), err)
	} else {
		result.DmgMultOvereating = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: DmgMultMaterial at offset 168
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+168); err != nil {
		errs.Add("DamageModelComponent.DmgMultMaterial", uintptr(int64(addr)+168), err)
	} else {
		result.DmgMultMaterial = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: InvincibilityFrames at offset 172
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+172); err != nil {
		errs.Add("DamageModelComponent.InvincibilityFrames", uintptr(int64(addr)+172), err)
	} else {
		result.InvincibilityFrames = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	return result, errs
}

func ReadInventory2Component(ctx *runtime.ReadContext, addr uintptr) (*Inventory2Component, runtime.Errors) {
	var errs runtime.Errors
	result := &Inventory2Component{}
	var buf [4]byte

	// Field: Header at offset 0
	{
		child, childErrs := ReadComponentHeader(ctx, uintptr(int64(addr)+0))
		if child != nil {
			result.Header = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: QuickInventorySlots at offset 72
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+72); err != nil {
		errs.Add("Inventory2Component.QuickInventorySlots", uintptr(int64(addr)+72), err)
	} else {
		result.QuickInventorySlots = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: FullInventorySlotsX at offset 76
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+76); err != nil {
		errs.Add("Inventory2Component.FullInventorySlotsX", uintptr(int64(addr)+76), err)
	} else {
		result.FullInventorySlotsX = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: FullInventorySlotsY at offset 80
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+80); err != nil {
		errs.Add("Inventory2Component.FullInventorySlotsY", uintptr(int64(addr)+80), err)
	} else {
		result.FullInventorySlotsY = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: SavedActiveItemIndex at offset 84
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+84); err != nil {
		errs.Add("Inventory2Component.SavedActiveItemIndex", uintptr(int64(addr)+84), err)
	} else {
		result.SavedActiveItemIndex = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ActiveItem at offset 88
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+88); err != nil {
		errs.Add("Inventory2Component.ActiveItem", uintptr(int64(addr)+88), err)
	} else {
		result.ActiveItem = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ActualActiveItem at offset 92
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+92); err != nil {
		errs.Add("Inventory2Component.ActualActiveItem", uintptr(int64(addr)+92), err)
	} else {
		result.ActualActiveItem = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ActiveStash at offset 96
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+96); err != nil {
		errs.Add("Inventory2Component.ActiveStash", uintptr(int64(addr)+96), err)
	} else {
		result.ActiveStash = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ThrowItem at offset 100
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+100); err != nil {
		errs.Add("Inventory2Component.ThrowItem", uintptr(int64(addr)+100), err)
	} else {
		result.ThrowItem = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ItemHolstered at offset 104
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+104); err != nil {
		errs.Add("Inventory2Component.ItemHolstered", uintptr(int64(addr)+104), err)
	} else {
		result.ItemHolstered = buf[0] != 0
	}

	// Field: Initialized at offset 105
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+105); err != nil {
		errs.Add("Inventory2Component.Initialized", uintptr(int64(addr)+105), err)
	} else {
		result.Initialized = buf[0] != 0
	}

	// Field: ForceRefresh at offset 106
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+106); err != nil {
		errs.Add("Inventory2Component.ForceRefresh", uintptr(int64(addr)+106), err)
	} else {
		result.ForceRefresh = buf[0] != 0
	}

	// Field: DontLogNextItemEquip at offset 107
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+107); err != nil {
		errs.Add("Inventory2Component.DontLogNextItemEquip", uintptr(int64(addr)+107), err)
	} else {
		result.DontLogNextItemEquip = buf[0] != 0
	}

	// Field: SmoothedItemXOffset at offset 108
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+108); err != nil {
		errs.Add("Inventory2Component.SmoothedItemXOffset", uintptr(int64(addr)+108), err)
	} else {
		result.SmoothedItemXOffset = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: LastItemSwitchFrame at offset 112
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+112); err != nil {
		errs.Add("Inventory2Component.LastItemSwitchFrame", uintptr(int64(addr)+112), err)
	} else {
		result.LastItemSwitchFrame = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: IntroEquipItemLerp at offset 116
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+116); err != nil {
		errs.Add("Inventory2Component.IntroEquipItemLerp", uintptr(int64(addr)+116), err)
	} else {
		result.IntroEquipItemLerp = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: SmoothedItemAngleX at offset 120
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+120); err != nil {
		errs.Add("Inventory2Component.SmoothedItemAngleX", uintptr(int64(addr)+120), err)
	} else {
		result.SmoothedItemAngleX = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: SmoothedItemAngleY at offset 124
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+124); err != nil {
		errs.Add("Inventory2Component.SmoothedItemAngleY", uintptr(int64(addr)+124), err)
	} else {
		result.SmoothedItemAngleY = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	return result, errs
}

func ReadWorldManagerViewRect(ctx *runtime.ReadContext, addr uintptr) (*WorldManagerViewRect, runtime.Errors) {
	var errs runtime.Errors
	result := &WorldManagerViewRect{}
	var buf [4]byte

	// Field: ViewX at offset 0
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+0); err != nil {
		errs.Add("WorldManagerViewRect.ViewX", uintptr(int64(addr)+0), err)
	} else {
		result.ViewX = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ViewY at offset 4
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+4); err != nil {
		errs.Add("WorldManagerViewRect.ViewY", uintptr(int64(addr)+4), err)
	} else {
		result.ViewY = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ViewWidth at offset 8
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+8); err != nil {
		errs.Add("WorldManagerViewRect.ViewWidth", uintptr(int64(addr)+8), err)
	} else {
		result.ViewWidth = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ViewHeight at offset 12
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+12); err != nil {
		errs.Add("WorldManagerViewRect.ViewHeight", uintptr(int64(addr)+12), err)
	} else {
		result.ViewHeight = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	return result, errs
}

func ReadDeathMatchApp(ctx *runtime.ReadContext, addr uintptr) (*DeathMatchApp, runtime.Errors) {
	var errs runtime.Errors
	result := &DeathMatchApp{}

	// Field: PlayerEntities at offset 88
	{
		child, childErrs := ReadStdVectorHeader(ctx, uintptr(int64(addr)+88))
		if child != nil {
			result.PlayerEntities = *child
		}
		errs = append(errs, childErrs...)
	}

	return result, errs
}

func ReadWorldStateComponent(ctx *runtime.ReadContext, addr uintptr) (*WorldStateComponent, runtime.Errors) {
	var errs runtime.Errors
	result := &WorldStateComponent{}
	var buf [4]byte

	// Field: Header at offset 0
	{
		child, childErrs := ReadComponentHeader(ctx, uintptr(int64(addr)+0))
		if child != nil {
			result.Header = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: BiomeCryptCount at offset 264
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+264); err != nil {
		errs.Add("WorldStateComponent.BiomeCryptCount", uintptr(int64(addr)+264), err)
	} else {
		result.BiomeCryptCount = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: GodsAfraid at offset 268
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+268); err != nil {
		errs.Add("WorldStateComponent.GodsAfraid", uintptr(int64(addr)+268), err)
	} else {
		result.GodsAfraid = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: GodsImpressed at offset 272
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+272); err != nil {
		errs.Add("WorldStateComponent.GodsImpressed", uintptr(int64(addr)+272), err)
	} else {
		result.GodsImpressed = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: GodsAfraidDamage at offset 276
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+276); err != nil {
		errs.Add("WorldStateComponent.GodsAfraidDamage", uintptr(int64(addr)+276), err)
	} else {
		result.GodsAfraidDamage = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: GodsEnraged at offset 280
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+280); err != nil {
		errs.Add("WorldStateComponent.GodsEnraged", uintptr(int64(addr)+280), err)
	} else {
		result.GodsEnraged = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	return result, errs
}

func ReadChildrenContainer(ctx *runtime.ReadContext, addr uintptr) (*ChildrenContainer, runtime.Errors) {
	var errs runtime.Errors
	result := &ChildrenContainer{}
	var buf [4]byte

	// Field: BeginPtr at offset 0
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+0); err != nil {
		errs.Add("ChildrenContainer.BeginPtr", uintptr(int64(addr)+0), err)
	} else {
		result.BeginPtr = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: EndPtr at offset 4
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+4); err != nil {
		errs.Add("ChildrenContainer.EndPtr", uintptr(int64(addr)+4), err)
	} else {
		result.EndPtr = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: CapacityPtr at offset 8
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+8); err != nil {
		errs.Add("ChildrenContainer.CapacityPtr", uintptr(int64(addr)+8), err)
	} else {
		result.CapacityPtr = binary.LittleEndian.Uint32(buf[:4])
	}

	return result, errs
}

func ReadAbilityComponent(ctx *runtime.ReadContext, addr uintptr) (*AbilityComponent, runtime.Errors) {
	var errs runtime.Errors
	result := &AbilityComponent{}
	var buf [4]byte

	// Field: Header at offset 0
	{
		child, childErrs := ReadComponentHeader(ctx, uintptr(int64(addr)+0))
		if child != nil {
			result.Header = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: CooldownFrames at offset 72
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+72); err != nil {
		errs.Add("AbilityComponent.CooldownFrames", uintptr(int64(addr)+72), err)
	} else {
		result.CooldownFrames = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: EntityFile at offset 76
	{
		child, childErrs := ReadMsvcString(ctx, uintptr(int64(addr)+76))
		if child != nil {
			result.EntityFile = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: SpriteFile at offset 100
	{
		child, childErrs := ReadMsvcString(ctx, uintptr(int64(addr)+100))
		if child != nil {
			result.SpriteFile = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: EntityCount at offset 124
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+124); err != nil {
		errs.Add("AbilityComponent.EntityCount", uintptr(int64(addr)+124), err)
	} else {
		result.EntityCount = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: NeverReload at offset 128
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+128); err != nil {
		errs.Add("AbilityComponent.NeverReload", uintptr(int64(addr)+128), err)
	} else {
		result.NeverReload = buf[0] != 0
	}

	// Field: ReloadTimeFrames at offset 132
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+132); err != nil {
		errs.Add("AbilityComponent.ReloadTimeFrames", uintptr(int64(addr)+132), err)
	} else {
		result.ReloadTimeFrames = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: Mana at offset 136
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+136); err != nil {
		errs.Add("AbilityComponent.Mana", uintptr(int64(addr)+136), err)
	} else {
		result.Mana = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ManaMax at offset 140
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+140); err != nil {
		errs.Add("AbilityComponent.ManaMax", uintptr(int64(addr)+140), err)
	} else {
		result.ManaMax = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ManaChargeSpeed at offset 144
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+144); err != nil {
		errs.Add("AbilityComponent.ManaChargeSpeed", uintptr(int64(addr)+144), err)
	} else {
		result.ManaChargeSpeed = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: RotateInHand at offset 148
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+148); err != nil {
		errs.Add("AbilityComponent.RotateInHand", uintptr(int64(addr)+148), err)
	} else {
		result.RotateInHand = buf[0] != 0
	}

	// Field: RotateInHandAmount at offset 152
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+152); err != nil {
		errs.Add("AbilityComponent.RotateInHandAmount", uintptr(int64(addr)+152), err)
	} else {
		result.RotateInHandAmount = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: RotateHandAmount at offset 156
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+156); err != nil {
		errs.Add("AbilityComponent.RotateHandAmount", uintptr(int64(addr)+156), err)
	} else {
		result.RotateHandAmount = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: FastProjectile at offset 160
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+160); err != nil {
		errs.Add("AbilityComponent.FastProjectile", uintptr(int64(addr)+160), err)
	} else {
		result.FastProjectile = buf[0] != 0
	}

	// Field: SwimPropelAmount at offset 164
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+164); err != nil {
		errs.Add("AbilityComponent.SwimPropelAmount", uintptr(int64(addr)+164), err)
	} else {
		result.SwimPropelAmount = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: MaxChargedActions at offset 168
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+168); err != nil {
		errs.Add("AbilityComponent.MaxChargedActions", uintptr(int64(addr)+168), err)
	} else {
		result.MaxChargedActions = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ChargeWaitFrames at offset 172
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+172); err != nil {
		errs.Add("AbilityComponent.ChargeWaitFrames", uintptr(int64(addr)+172), err)
	} else {
		result.ChargeWaitFrames = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ItemRecoilRecoverySpeed at offset 176
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+176); err != nil {
		errs.Add("AbilityComponent.ItemRecoilRecoverySpeed", uintptr(int64(addr)+176), err)
	} else {
		result.ItemRecoilRecoverySpeed = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ItemRecoilMax at offset 180
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+180); err != nil {
		errs.Add("AbilityComponent.ItemRecoilMax", uintptr(int64(addr)+180), err)
	} else {
		result.ItemRecoilMax = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ItemRecoilOffsetCoeff at offset 184
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+184); err != nil {
		errs.Add("AbilityComponent.ItemRecoilOffsetCoeff", uintptr(int64(addr)+184), err)
	} else {
		result.ItemRecoilOffsetCoeff = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ItemRecoilRotationCoeff at offset 188
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+188); err != nil {
		errs.Add("AbilityComponent.ItemRecoilRotationCoeff", uintptr(int64(addr)+188), err)
	} else {
		result.ItemRecoilRotationCoeff = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: BaseItemFile at offset 192
	{
		child, childErrs := ReadMsvcString(ctx, uintptr(int64(addr)+192))
		if child != nil {
			result.BaseItemFile = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: UseEntityFileAsProjectileInfoProxy at offset 216
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+216); err != nil {
		errs.Add("AbilityComponent.UseEntityFileAsProjectileInfoProxy", uintptr(int64(addr)+216), err)
	} else {
		result.UseEntityFileAsProjectileInfoProxy = buf[0] != 0
	}

	// Field: ClickToUse at offset 217
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+217); err != nil {
		errs.Add("AbilityComponent.ClickToUse", uintptr(int64(addr)+217), err)
	} else {
		result.ClickToUse = buf[0] != 0
	}

	// Field: StatTimesPlayerHasShot at offset 220
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+220); err != nil {
		errs.Add("AbilityComponent.StatTimesPlayerHasShot", uintptr(int64(addr)+220), err)
	} else {
		result.StatTimesPlayerHasShot = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: StatTimesPlayerHasEdited at offset 224
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+224); err != nil {
		errs.Add("AbilityComponent.StatTimesPlayerHasEdited", uintptr(int64(addr)+224), err)
	} else {
		result.StatTimesPlayerHasEdited = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ShootingReducesAmountInInventory at offset 228
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+228); err != nil {
		errs.Add("AbilityComponent.ShootingReducesAmountInInventory", uintptr(int64(addr)+228), err)
	} else {
		result.ShootingReducesAmountInInventory = buf[0] != 0
	}

	// Field: ThrowAsItem at offset 229
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+229); err != nil {
		errs.Add("AbilityComponent.ThrowAsItem", uintptr(int64(addr)+229), err)
	} else {
		result.ThrowAsItem = buf[0] != 0
	}

	// Field: SimulateThrowAsItem at offset 230
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+230); err != nil {
		errs.Add("AbilityComponent.SimulateThrowAsItem", uintptr(int64(addr)+230), err)
	} else {
		result.SimulateThrowAsItem = buf[0] != 0
	}

	// Field: MaxAmountInInventory at offset 232
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+232); err != nil {
		errs.Add("AbilityComponent.MaxAmountInInventory", uintptr(int64(addr)+232), err)
	} else {
		result.MaxAmountInInventory = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: AmountInInventory at offset 236
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+236); err != nil {
		errs.Add("AbilityComponent.AmountInInventory", uintptr(int64(addr)+236), err)
	} else {
		result.AmountInInventory = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: DropAsItemOnDeath at offset 240
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+240); err != nil {
		errs.Add("AbilityComponent.DropAsItemOnDeath", uintptr(int64(addr)+240), err)
	} else {
		result.DropAsItemOnDeath = buf[0] != 0
	}

	// Field: UiName at offset 244
	{
		child, childErrs := ReadMsvcString(ctx, uintptr(int64(addr)+244))
		if child != nil {
			result.UiName = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: UseGunScript at offset 268
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+268); err != nil {
		errs.Add("AbilityComponent.UseGunScript", uintptr(int64(addr)+268), err)
	} else {
		result.UseGunScript = buf[0] != 0
	}

	// Field: IsPetrisGun at offset 269
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+269); err != nil {
		errs.Add("AbilityComponent.IsPetrisGun", uintptr(int64(addr)+269), err)
	} else {
		result.IsPetrisGun = buf[0] != 0
	}

	// Field: GunConfig at offset 272
	{
		child, childErrs := ReadConfigGun(ctx, uintptr(int64(addr)+272))
		if child != nil {
			result.GunConfig = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: GunLevel at offset 864
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+864); err != nil {
		errs.Add("AbilityComponent.GunLevel", uintptr(int64(addr)+864), err)
	} else {
		result.GunLevel = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: AddTheseChildActions at offset 868
	{
		child, childErrs := ReadMsvcString(ctx, uintptr(int64(addr)+868))
		if child != nil {
			result.AddTheseChildActions = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: CurrentSlotDurability at offset 892
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+892); err != nil {
		errs.Add("AbilityComponent.CurrentSlotDurability", uintptr(int64(addr)+892), err)
	} else {
		result.CurrentSlotDurability = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: SlotConsumptionFunction at offset 896
	{
		child, childErrs := ReadMsvcString(ctx, uintptr(int64(addr)+896))
		if child != nil {
			result.SlotConsumptionFunction = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: NextFrameUsable at offset 920
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+920); err != nil {
		errs.Add("AbilityComponent.NextFrameUsable", uintptr(int64(addr)+920), err)
	} else {
		result.NextFrameUsable = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: CastDelayStartFrame at offset 924
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+924); err != nil {
		errs.Add("AbilityComponent.CastDelayStartFrame", uintptr(int64(addr)+924), err)
	} else {
		result.CastDelayStartFrame = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: AmmoLeft at offset 928
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+928); err != nil {
		errs.Add("AbilityComponent.AmmoLeft", uintptr(int64(addr)+928), err)
	} else {
		result.AmmoLeft = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ReloadFramesLeft at offset 932
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+932); err != nil {
		errs.Add("AbilityComponent.ReloadFramesLeft", uintptr(int64(addr)+932), err)
	} else {
		result.ReloadFramesLeft = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ReloadNextFrameUsable at offset 936
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+936); err != nil {
		errs.Add("AbilityComponent.ReloadNextFrameUsable", uintptr(int64(addr)+936), err)
	} else {
		result.ReloadNextFrameUsable = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ChargeCount at offset 940
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+940); err != nil {
		errs.Add("AbilityComponent.ChargeCount", uintptr(int64(addr)+940), err)
	} else {
		result.ChargeCount = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: NextChargeFrame at offset 944
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+944); err != nil {
		errs.Add("AbilityComponent.NextChargeFrame", uintptr(int64(addr)+944), err)
	} else {
		result.NextChargeFrame = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ItemRecoil at offset 948
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+948); err != nil {
		errs.Add("AbilityComponent.ItemRecoil", uintptr(int64(addr)+948), err)
	} else {
		result.ItemRecoil = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: IsInitialized at offset 952
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+952); err != nil {
		errs.Add("AbilityComponent.IsInitialized", uintptr(int64(addr)+952), err)
	} else {
		result.IsInitialized = buf[0] != 0
	}

	return result, errs
}

func ReadGameGlobals(ctx *runtime.ReadContext, addr uintptr) (*GameGlobals, runtime.Errors) {
	var errs runtime.Errors
	result := &GameGlobals{}
	var buf [4]byte

	// Field: FrameCount at offset 0
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+0); err != nil {
		errs.Add("GameGlobals.FrameCount", uintptr(int64(addr)+0), err)
	} else {
		result.FrameCount = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: PhysicsStepCount at offset 4
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+4); err != nil {
		errs.Add("GameGlobals.PhysicsStepCount", uintptr(int64(addr)+4), err)
	} else {
		result.PhysicsStepCount = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: GameTime at offset 8
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+8); err != nil {
		errs.Add("GameGlobals.GameTime", uintptr(int64(addr)+8), err)
	} else {
		result.GameTime = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: PWorldManager at offset 12
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+12); err != nil {
		errs.Add("GameGlobals.PWorldManager", uintptr(int64(addr)+12), err)
	} else {
		result.PWorldManager = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: PChunkSystem at offset 16
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+16); err != nil {
		errs.Add("GameGlobals.PChunkSystem", uintptr(int64(addr)+16), err)
	} else {
		result.PChunkSystem = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: PCellGrid at offset 20
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+20); err != nil {
		errs.Add("GameGlobals.PCellGrid", uintptr(int64(addr)+20), err)
	} else {
		result.PCellGrid = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: PCellFactory at offset 24
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+24); err != nil {
		errs.Add("GameGlobals.PCellFactory", uintptr(int64(addr)+24), err)
	} else {
		result.PCellFactory = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: Unknown1c at offset 28
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+28); err != nil {
		errs.Add("GameGlobals.Unknown1c", uintptr(int64(addr)+28), err)
	} else {
		result.Unknown1c = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: PPhysicsWorld at offset 32
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+32); err != nil {
		errs.Add("GameGlobals.PPhysicsWorld", uintptr(int64(addr)+32), err)
	} else {
		result.PPhysicsWorld = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: PAudioManager at offset 36
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+36); err != nil {
		errs.Add("GameGlobals.PAudioManager", uintptr(int64(addr)+36), err)
	} else {
		result.PAudioManager = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: ViewportLeft at offset 384
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+384); err != nil {
		errs.Add("GameGlobals.ViewportLeft", uintptr(int64(addr)+384), err)
	} else {
		result.ViewportLeft = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ViewportTop at offset 388
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+388); err != nil {
		errs.Add("GameGlobals.ViewportTop", uintptr(int64(addr)+388), err)
	} else {
		result.ViewportTop = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ViewportRight at offset 392
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+392); err != nil {
		errs.Add("GameGlobals.ViewportRight", uintptr(int64(addr)+392), err)
	} else {
		result.ViewportRight = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ViewportBottom at offset 396
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+396); err != nil {
		errs.Add("GameGlobals.ViewportBottom", uintptr(int64(addr)+396), err)
	} else {
		result.ViewportBottom = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	return result, errs
}

// Ensure imports are used.
var (
	_ = binary.LittleEndian
	_ = math.Float32frombits
)
