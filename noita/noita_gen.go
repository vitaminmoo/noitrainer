package noita

import (
	"encoding/binary"
	"github.com/vitaminmoo/memtools/hexpat/runtime"
	"math"
)

type MsvcString struct {
	Data     [16]byte
	Length   uint32
	Capacity uint32
}

type CellData struct {
	Name MsvcString
}

type U32Vector struct {
	BeginPtr    uint32
	EndPtr      uint32
	CapacityPtr uint32
	Elements    []uint32
}

type DeathMatchApp struct {
	PlayerEntities U32Vector
}

type F64Vector struct {
	BeginPtr    uint32
	EndPtr      uint32
	CapacityPtr uint32
	Elements    []float64
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

type CellFactory struct {
	CellDataArrayPtr uint32
}

type S32Vector struct {
	BeginPtr    uint32
	EndPtr      uint32
	CapacityPtr uint32
	Elements    []int32
}

type ChildrenContainer struct {
	BeginPtr    uint32
	EndPtr      uint32
	CapacityPtr uint32
	Children    []uint32
}

type StdVectorHeader struct {
	BeginPtr    uint32
	EndPtr      uint32
	CapacityPtr uint32
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

type WorldStateComponent struct {
	Header           ComponentHeader
	BiomeCryptCount  int32
	GodsAfraid       int32
	GodsImpressed    int32
	GodsAfraidDamage int32
	GodsEnraged      int32
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

type EntityManager struct {
	Vtable           uint32
	NextEntityId     int32
	FreeSlotStack    StdVectorHeader
	EntityArray      StdVectorHeader
	TagGroups        StdVectorHeader
	ComponentBuffers U32Vector
	PEventManager    uint32
}

type WalletComponent struct {
	Header         ComponentHeader
	Money          int64
	MoneySpent     int64
	MoneyPrevFrame int64
	HasReachedInf  bool
}

type MaterialInventoryComponent struct {
	Header               ComponentHeader
	CountPerMaterialType F64Vector
}

func ReadMsvcString(ctx *runtime.ReadContext, addr uintptr) (*MsvcString, runtime.Errors) {
	var errs runtime.Errors
	result := &MsvcString{}
	var buf [4]byte

	// Field: Data (array[16]) at int64(addr)+0
	if _, err := ctx.ReadAt(result.Data[:], int64(addr)+0); err != nil {
		errs.Add("MsvcString.Data", uintptr(int64(addr)+0), err)
	}

	// Field: Length at int64(addr)+16
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+16); err != nil {
		errs.Add("MsvcString.Length", uintptr(int64(addr)+16), err)
	} else {
		result.Length = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: Capacity at int64(addr)+20
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+20); err != nil {
		errs.Add("MsvcString.Capacity", uintptr(int64(addr)+20), err)
	} else {
		result.Capacity = binary.LittleEndian.Uint32(buf[:4])
	}

	return result, errs
}

func ReadCellData(ctx *runtime.ReadContext, addr uintptr) (*CellData, runtime.Errors) {
	var errs runtime.Errors
	result := &CellData{}

	// Field: Name at int64(addr)+0
	{
		child, childErrs := ReadMsvcString(ctx, uintptr(int64(addr)+0))
		if child != nil {
			result.Name = *child
		}
		errs = append(errs, childErrs...)
	}

	return result, errs
}

func ReadU32Vector(ctx *runtime.ReadContext, addr uintptr) (*U32Vector, runtime.Errors) {
	var errs runtime.Errors
	result := &U32Vector{}
	var buf [4]byte
	offset := int64(0)

	// Field: BeginPtr at int64(addr)+offset
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+offset); err != nil {
		errs.Add("U32Vector.BeginPtr", uintptr(int64(addr)+offset), err)
	} else {
		result.BeginPtr = binary.LittleEndian.Uint32(buf[:4])
	}

	offset += 4
	// Field: EndPtr at int64(addr)+offset
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+offset); err != nil {
		errs.Add("U32Vector.EndPtr", uintptr(int64(addr)+offset), err)
	} else {
		result.EndPtr = binary.LittleEndian.Uint32(buf[:4])
	}

	offset += 4
	// Field: CapacityPtr at int64(addr)+offset
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+offset); err != nil {
		errs.Add("U32Vector.CapacityPtr", uintptr(int64(addr)+offset), err)
	} else {
		result.CapacityPtr = binary.LittleEndian.Uint32(buf[:4])
	}

	offset += 4
	// Field: Elements (dynamic array) at int64(result.BeginPtr)
	result.Elements = make([]uint32, int(((result.EndPtr - result.BeginPtr) / 4)))
	for i := range result.Elements {
		if _, err := ctx.ReadAt(buf[:4], int64(result.BeginPtr)+int64(i)*4); err != nil {
			errs.Add("U32Vector.Elements", uintptr(int64(result.BeginPtr)+int64(i)*4), err)
		} else {
			result.Elements[i] = binary.LittleEndian.Uint32(buf[:4])
		}
	}

	return result, errs
}

func ReadDeathMatchApp(ctx *runtime.ReadContext, addr uintptr) (*DeathMatchApp, runtime.Errors) {
	var errs runtime.Errors
	result := &DeathMatchApp{}

	// Field: PlayerEntities at int64(addr)+88
	{
		child, childErrs := ReadU32Vector(ctx, uintptr(int64(addr)+88))
		if child != nil {
			result.PlayerEntities = *child
		}
		errs = append(errs, childErrs...)
	}

	return result, errs
}

func ReadF64Vector(ctx *runtime.ReadContext, addr uintptr) (*F64Vector, runtime.Errors) {
	var errs runtime.Errors
	result := &F64Vector{}
	var buf [8]byte
	offset := int64(0)

	// Field: BeginPtr at int64(addr)+offset
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+offset); err != nil {
		errs.Add("F64Vector.BeginPtr", uintptr(int64(addr)+offset), err)
	} else {
		result.BeginPtr = binary.LittleEndian.Uint32(buf[:4])
	}

	offset += 4
	// Field: EndPtr at int64(addr)+offset
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+offset); err != nil {
		errs.Add("F64Vector.EndPtr", uintptr(int64(addr)+offset), err)
	} else {
		result.EndPtr = binary.LittleEndian.Uint32(buf[:4])
	}

	offset += 4
	// Field: CapacityPtr at int64(addr)+offset
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+offset); err != nil {
		errs.Add("F64Vector.CapacityPtr", uintptr(int64(addr)+offset), err)
	} else {
		result.CapacityPtr = binary.LittleEndian.Uint32(buf[:4])
	}

	offset += 4
	// Field: Elements (dynamic array) at int64(result.BeginPtr)
	result.Elements = make([]float64, int(((result.EndPtr - result.BeginPtr) / 8)))
	for i := range result.Elements {
		if _, err := ctx.ReadAt(buf[:8], int64(result.BeginPtr)+int64(i)*8); err != nil {
			errs.Add("F64Vector.Elements", uintptr(int64(result.BeginPtr)+int64(i)*8), err)
		} else {
			result.Elements[i] = math.Float64frombits(binary.LittleEndian.Uint64(buf[:8]))
		}
	}

	return result, errs
}

func ReadComponentHeader(ctx *runtime.ReadContext, addr uintptr) (*ComponentHeader, runtime.Errors) {
	var errs runtime.Errors
	result := &ComponentHeader{}
	var buf [4]byte

	// Field: Vtable at int64(addr)+0
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+0); err != nil {
		errs.Add("ComponentHeader.Vtable", uintptr(int64(addr)+0), err)
	} else {
		result.Vtable = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: BufferIndex at int64(addr)+4
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+4); err != nil {
		errs.Add("ComponentHeader.BufferIndex", uintptr(int64(addr)+4), err)
	} else {
		result.BufferIndex = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: PTypeName at int64(addr)+8
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+8); err != nil {
		errs.Add("ComponentHeader.PTypeName", uintptr(int64(addr)+8), err)
	} else {
		result.PTypeName = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: TypeId at int64(addr)+12
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+12); err != nil {
		errs.Add("ComponentHeader.TypeId", uintptr(int64(addr)+12), err)
	} else {
		result.TypeId = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: Unknown10 at int64(addr)+16
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+16); err != nil {
		errs.Add("ComponentHeader.Unknown10", uintptr(int64(addr)+16), err)
	} else {
		result.Unknown10 = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: Active at int64(addr)+20
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+20); err != nil {
		errs.Add("ComponentHeader.Active", uintptr(int64(addr)+20), err)
	} else {
		result.Active = buf[0] != 0
	}

	// Field: ComponentTags (array[32]) at int64(addr)+24
	if _, err := ctx.ReadAt(result.ComponentTags[:], int64(addr)+24); err != nil {
		errs.Add("ComponentHeader.ComponentTags", uintptr(int64(addr)+24), err)
	}

	// Field: Unknown38 at int64(addr)+56
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+56); err != nil {
		errs.Add("ComponentHeader.Unknown38", uintptr(int64(addr)+56), err)
	} else {
		result.Unknown38 = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: Unknown3c at int64(addr)+60
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+60); err != nil {
		errs.Add("ComponentHeader.Unknown3c", uintptr(int64(addr)+60), err)
	} else {
		result.Unknown3c = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: Unknown40 at int64(addr)+64
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+64); err != nil {
		errs.Add("ComponentHeader.Unknown40", uintptr(int64(addr)+64), err)
	} else {
		result.Unknown40 = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: Unknown44 at int64(addr)+68
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+68); err != nil {
		errs.Add("ComponentHeader.Unknown44", uintptr(int64(addr)+68), err)
	} else {
		result.Unknown44 = binary.LittleEndian.Uint32(buf[:4])
	}

	return result, errs
}

func ReadDamageModelComponent(ctx *runtime.ReadContext, addr uintptr) (*DamageModelComponent, runtime.Errors) {
	var errs runtime.Errors
	result := &DamageModelComponent{}
	var buf [8]byte

	// Field: Header at int64(addr)+0
	{
		child, childErrs := ReadComponentHeader(ctx, uintptr(int64(addr)+0))
		if child != nil {
			result.Header = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: Hp at int64(addr)+72
	if _, err := ctx.ReadAt(buf[:8], int64(addr)+72); err != nil {
		errs.Add("DamageModelComponent.Hp", uintptr(int64(addr)+72), err)
	} else {
		result.Hp = math.Float64frombits(binary.LittleEndian.Uint64(buf[:8]))
	}

	// Field: MaxHp at int64(addr)+80
	if _, err := ctx.ReadAt(buf[:8], int64(addr)+80); err != nil {
		errs.Add("DamageModelComponent.MaxHp", uintptr(int64(addr)+80), err)
	} else {
		result.MaxHp = math.Float64frombits(binary.LittleEndian.Uint64(buf[:8]))
	}

	// Field: MaxHpCap at int64(addr)+88
	if _, err := ctx.ReadAt(buf[:8], int64(addr)+88); err != nil {
		errs.Add("DamageModelComponent.MaxHpCap", uintptr(int64(addr)+88), err)
	} else {
		result.MaxHpCap = math.Float64frombits(binary.LittleEndian.Uint64(buf[:8]))
	}

	// Field: MaxHpOld at int64(addr)+96
	if _, err := ctx.ReadAt(buf[:8], int64(addr)+96); err != nil {
		errs.Add("DamageModelComponent.MaxHpOld", uintptr(int64(addr)+96), err)
	} else {
		result.MaxHpOld = math.Float64frombits(binary.LittleEndian.Uint64(buf[:8]))
	}

	// Field: Unknown68 at int64(addr)+104
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+104); err != nil {
		errs.Add("DamageModelComponent.Unknown68", uintptr(int64(addr)+104), err)
	} else {
		result.Unknown68 = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: DmgMultMelee at int64(addr)+108
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+108); err != nil {
		errs.Add("DamageModelComponent.DmgMultMelee", uintptr(int64(addr)+108), err)
	} else {
		result.DmgMultMelee = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: DmgMultProjectile at int64(addr)+112
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+112); err != nil {
		errs.Add("DamageModelComponent.DmgMultProjectile", uintptr(int64(addr)+112), err)
	} else {
		result.DmgMultProjectile = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: DmgMultExplosion at int64(addr)+116
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+116); err != nil {
		errs.Add("DamageModelComponent.DmgMultExplosion", uintptr(int64(addr)+116), err)
	} else {
		result.DmgMultExplosion = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: DmgMultElectricity at int64(addr)+120
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+120); err != nil {
		errs.Add("DamageModelComponent.DmgMultElectricity", uintptr(int64(addr)+120), err)
	} else {
		result.DmgMultElectricity = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: DmgMultFire at int64(addr)+124
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+124); err != nil {
		errs.Add("DamageModelComponent.DmgMultFire", uintptr(int64(addr)+124), err)
	} else {
		result.DmgMultFire = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: DmgMultDrill at int64(addr)+128
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+128); err != nil {
		errs.Add("DamageModelComponent.DmgMultDrill", uintptr(int64(addr)+128), err)
	} else {
		result.DmgMultDrill = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: DmgMultSlice at int64(addr)+132
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+132); err != nil {
		errs.Add("DamageModelComponent.DmgMultSlice", uintptr(int64(addr)+132), err)
	} else {
		result.DmgMultSlice = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: DmgMultIce at int64(addr)+136
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+136); err != nil {
		errs.Add("DamageModelComponent.DmgMultIce", uintptr(int64(addr)+136), err)
	} else {
		result.DmgMultIce = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: DmgMultHealing at int64(addr)+140
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+140); err != nil {
		errs.Add("DamageModelComponent.DmgMultHealing", uintptr(int64(addr)+140), err)
	} else {
		result.DmgMultHealing = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: DmgMultPhysicsHit at int64(addr)+144
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+144); err != nil {
		errs.Add("DamageModelComponent.DmgMultPhysicsHit", uintptr(int64(addr)+144), err)
	} else {
		result.DmgMultPhysicsHit = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: DmgMultRadioactive at int64(addr)+148
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+148); err != nil {
		errs.Add("DamageModelComponent.DmgMultRadioactive", uintptr(int64(addr)+148), err)
	} else {
		result.DmgMultRadioactive = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: DmgMultPoison at int64(addr)+152
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+152); err != nil {
		errs.Add("DamageModelComponent.DmgMultPoison", uintptr(int64(addr)+152), err)
	} else {
		result.DmgMultPoison = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: DmgMultHoly at int64(addr)+156
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+156); err != nil {
		errs.Add("DamageModelComponent.DmgMultHoly", uintptr(int64(addr)+156), err)
	} else {
		result.DmgMultHoly = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: DmgMultCurse at int64(addr)+160
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+160); err != nil {
		errs.Add("DamageModelComponent.DmgMultCurse", uintptr(int64(addr)+160), err)
	} else {
		result.DmgMultCurse = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: DmgMultOvereating at int64(addr)+164
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+164); err != nil {
		errs.Add("DamageModelComponent.DmgMultOvereating", uintptr(int64(addr)+164), err)
	} else {
		result.DmgMultOvereating = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: DmgMultMaterial at int64(addr)+168
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+168); err != nil {
		errs.Add("DamageModelComponent.DmgMultMaterial", uintptr(int64(addr)+168), err)
	} else {
		result.DmgMultMaterial = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: InvincibilityFrames at int64(addr)+172
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+172); err != nil {
		errs.Add("DamageModelComponent.InvincibilityFrames", uintptr(int64(addr)+172), err)
	} else {
		result.InvincibilityFrames = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	return result, errs
}

func ReadCharacterDataComponent(ctx *runtime.ReadContext, addr uintptr) (*CharacterDataComponent, runtime.Errors) {
	var errs runtime.Errors
	result := &CharacterDataComponent{}
	var buf [4]byte

	// Field: Header at int64(addr)+0
	{
		child, childErrs := ReadComponentHeader(ctx, uintptr(int64(addr)+0))
		if child != nil {
			result.Header = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: Gravity at int64(addr)+136
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+136); err != nil {
		errs.Add("CharacterDataComponent.Gravity", uintptr(int64(addr)+136), err)
	} else {
		result.Gravity = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: FlyTime at int64(addr)+140
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+140); err != nil {
		errs.Add("CharacterDataComponent.FlyTime", uintptr(int64(addr)+140), err)
	} else {
		result.FlyTime = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: IsOnGround at int64(addr)+184
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+184); err != nil {
		errs.Add("CharacterDataComponent.IsOnGround", uintptr(int64(addr)+184), err)
	} else {
		result.IsOnGround = buf[0] != 0
	}

	// Field: VelocityX at int64(addr)+264
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+264); err != nil {
		errs.Add("CharacterDataComponent.VelocityX", uintptr(int64(addr)+264), err)
	} else {
		result.VelocityX = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: VelocityY at int64(addr)+268
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

	// Field: Vtable at int64(addr)+0
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+0); err != nil {
		errs.Add("ConfigGun.Vtable", uintptr(int64(addr)+0), err)
	} else {
		result.Vtable = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: ActionsPerRound at int64(addr)+4
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+4); err != nil {
		errs.Add("ConfigGun.ActionsPerRound", uintptr(int64(addr)+4), err)
	} else {
		result.ActionsPerRound = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ShuffleDeckWhenEmpty at int64(addr)+8
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+8); err != nil {
		errs.Add("ConfigGun.ShuffleDeckWhenEmpty", uintptr(int64(addr)+8), err)
	} else {
		result.ShuffleDeckWhenEmpty = buf[0] != 0
	}

	// Field: ReloadTime at int64(addr)+12
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+12); err != nil {
		errs.Add("ConfigGun.ReloadTime", uintptr(int64(addr)+12), err)
	} else {
		result.ReloadTime = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: DeckCapacity at int64(addr)+16
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+16); err != nil {
		errs.Add("ConfigGun.DeckCapacity", uintptr(int64(addr)+16), err)
	} else {
		result.DeckCapacity = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	return result, errs
}

func ReadAbilityComponent(ctx *runtime.ReadContext, addr uintptr) (*AbilityComponent, runtime.Errors) {
	var errs runtime.Errors
	result := &AbilityComponent{}
	var buf [4]byte

	// Field: Header at int64(addr)+0
	{
		child, childErrs := ReadComponentHeader(ctx, uintptr(int64(addr)+0))
		if child != nil {
			result.Header = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: CooldownFrames at int64(addr)+72
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+72); err != nil {
		errs.Add("AbilityComponent.CooldownFrames", uintptr(int64(addr)+72), err)
	} else {
		result.CooldownFrames = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: EntityFile at int64(addr)+76
	{
		child, childErrs := ReadMsvcString(ctx, uintptr(int64(addr)+76))
		if child != nil {
			result.EntityFile = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: SpriteFile at int64(addr)+100
	{
		child, childErrs := ReadMsvcString(ctx, uintptr(int64(addr)+100))
		if child != nil {
			result.SpriteFile = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: EntityCount at int64(addr)+124
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+124); err != nil {
		errs.Add("AbilityComponent.EntityCount", uintptr(int64(addr)+124), err)
	} else {
		result.EntityCount = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: NeverReload at int64(addr)+128
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+128); err != nil {
		errs.Add("AbilityComponent.NeverReload", uintptr(int64(addr)+128), err)
	} else {
		result.NeverReload = buf[0] != 0
	}

	// Field: ReloadTimeFrames at int64(addr)+132
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+132); err != nil {
		errs.Add("AbilityComponent.ReloadTimeFrames", uintptr(int64(addr)+132), err)
	} else {
		result.ReloadTimeFrames = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: Mana at int64(addr)+136
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+136); err != nil {
		errs.Add("AbilityComponent.Mana", uintptr(int64(addr)+136), err)
	} else {
		result.Mana = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ManaMax at int64(addr)+140
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+140); err != nil {
		errs.Add("AbilityComponent.ManaMax", uintptr(int64(addr)+140), err)
	} else {
		result.ManaMax = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ManaChargeSpeed at int64(addr)+144
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+144); err != nil {
		errs.Add("AbilityComponent.ManaChargeSpeed", uintptr(int64(addr)+144), err)
	} else {
		result.ManaChargeSpeed = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: RotateInHand at int64(addr)+148
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+148); err != nil {
		errs.Add("AbilityComponent.RotateInHand", uintptr(int64(addr)+148), err)
	} else {
		result.RotateInHand = buf[0] != 0
	}

	// Field: RotateInHandAmount at int64(addr)+152
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+152); err != nil {
		errs.Add("AbilityComponent.RotateInHandAmount", uintptr(int64(addr)+152), err)
	} else {
		result.RotateInHandAmount = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: RotateHandAmount at int64(addr)+156
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+156); err != nil {
		errs.Add("AbilityComponent.RotateHandAmount", uintptr(int64(addr)+156), err)
	} else {
		result.RotateHandAmount = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: FastProjectile at int64(addr)+160
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+160); err != nil {
		errs.Add("AbilityComponent.FastProjectile", uintptr(int64(addr)+160), err)
	} else {
		result.FastProjectile = buf[0] != 0
	}

	// Field: SwimPropelAmount at int64(addr)+164
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+164); err != nil {
		errs.Add("AbilityComponent.SwimPropelAmount", uintptr(int64(addr)+164), err)
	} else {
		result.SwimPropelAmount = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: MaxChargedActions at int64(addr)+168
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+168); err != nil {
		errs.Add("AbilityComponent.MaxChargedActions", uintptr(int64(addr)+168), err)
	} else {
		result.MaxChargedActions = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ChargeWaitFrames at int64(addr)+172
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+172); err != nil {
		errs.Add("AbilityComponent.ChargeWaitFrames", uintptr(int64(addr)+172), err)
	} else {
		result.ChargeWaitFrames = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ItemRecoilRecoverySpeed at int64(addr)+176
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+176); err != nil {
		errs.Add("AbilityComponent.ItemRecoilRecoverySpeed", uintptr(int64(addr)+176), err)
	} else {
		result.ItemRecoilRecoverySpeed = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ItemRecoilMax at int64(addr)+180
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+180); err != nil {
		errs.Add("AbilityComponent.ItemRecoilMax", uintptr(int64(addr)+180), err)
	} else {
		result.ItemRecoilMax = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ItemRecoilOffsetCoeff at int64(addr)+184
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+184); err != nil {
		errs.Add("AbilityComponent.ItemRecoilOffsetCoeff", uintptr(int64(addr)+184), err)
	} else {
		result.ItemRecoilOffsetCoeff = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ItemRecoilRotationCoeff at int64(addr)+188
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+188); err != nil {
		errs.Add("AbilityComponent.ItemRecoilRotationCoeff", uintptr(int64(addr)+188), err)
	} else {
		result.ItemRecoilRotationCoeff = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: BaseItemFile at int64(addr)+192
	{
		child, childErrs := ReadMsvcString(ctx, uintptr(int64(addr)+192))
		if child != nil {
			result.BaseItemFile = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: UseEntityFileAsProjectileInfoProxy at int64(addr)+216
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+216); err != nil {
		errs.Add("AbilityComponent.UseEntityFileAsProjectileInfoProxy", uintptr(int64(addr)+216), err)
	} else {
		result.UseEntityFileAsProjectileInfoProxy = buf[0] != 0
	}

	// Field: ClickToUse at int64(addr)+217
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+217); err != nil {
		errs.Add("AbilityComponent.ClickToUse", uintptr(int64(addr)+217), err)
	} else {
		result.ClickToUse = buf[0] != 0
	}

	// Field: StatTimesPlayerHasShot at int64(addr)+220
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+220); err != nil {
		errs.Add("AbilityComponent.StatTimesPlayerHasShot", uintptr(int64(addr)+220), err)
	} else {
		result.StatTimesPlayerHasShot = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: StatTimesPlayerHasEdited at int64(addr)+224
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+224); err != nil {
		errs.Add("AbilityComponent.StatTimesPlayerHasEdited", uintptr(int64(addr)+224), err)
	} else {
		result.StatTimesPlayerHasEdited = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ShootingReducesAmountInInventory at int64(addr)+228
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+228); err != nil {
		errs.Add("AbilityComponent.ShootingReducesAmountInInventory", uintptr(int64(addr)+228), err)
	} else {
		result.ShootingReducesAmountInInventory = buf[0] != 0
	}

	// Field: ThrowAsItem at int64(addr)+229
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+229); err != nil {
		errs.Add("AbilityComponent.ThrowAsItem", uintptr(int64(addr)+229), err)
	} else {
		result.ThrowAsItem = buf[0] != 0
	}

	// Field: SimulateThrowAsItem at int64(addr)+230
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+230); err != nil {
		errs.Add("AbilityComponent.SimulateThrowAsItem", uintptr(int64(addr)+230), err)
	} else {
		result.SimulateThrowAsItem = buf[0] != 0
	}

	// Field: MaxAmountInInventory at int64(addr)+232
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+232); err != nil {
		errs.Add("AbilityComponent.MaxAmountInInventory", uintptr(int64(addr)+232), err)
	} else {
		result.MaxAmountInInventory = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: AmountInInventory at int64(addr)+236
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+236); err != nil {
		errs.Add("AbilityComponent.AmountInInventory", uintptr(int64(addr)+236), err)
	} else {
		result.AmountInInventory = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: DropAsItemOnDeath at int64(addr)+240
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+240); err != nil {
		errs.Add("AbilityComponent.DropAsItemOnDeath", uintptr(int64(addr)+240), err)
	} else {
		result.DropAsItemOnDeath = buf[0] != 0
	}

	// Field: UiName at int64(addr)+244
	{
		child, childErrs := ReadMsvcString(ctx, uintptr(int64(addr)+244))
		if child != nil {
			result.UiName = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: UseGunScript at int64(addr)+268
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+268); err != nil {
		errs.Add("AbilityComponent.UseGunScript", uintptr(int64(addr)+268), err)
	} else {
		result.UseGunScript = buf[0] != 0
	}

	// Field: IsPetrisGun at int64(addr)+269
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+269); err != nil {
		errs.Add("AbilityComponent.IsPetrisGun", uintptr(int64(addr)+269), err)
	} else {
		result.IsPetrisGun = buf[0] != 0
	}

	// Field: GunConfig at int64(addr)+272
	{
		child, childErrs := ReadConfigGun(ctx, uintptr(int64(addr)+272))
		if child != nil {
			result.GunConfig = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: GunLevel at int64(addr)+864
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+864); err != nil {
		errs.Add("AbilityComponent.GunLevel", uintptr(int64(addr)+864), err)
	} else {
		result.GunLevel = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: AddTheseChildActions at int64(addr)+868
	{
		child, childErrs := ReadMsvcString(ctx, uintptr(int64(addr)+868))
		if child != nil {
			result.AddTheseChildActions = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: CurrentSlotDurability at int64(addr)+892
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+892); err != nil {
		errs.Add("AbilityComponent.CurrentSlotDurability", uintptr(int64(addr)+892), err)
	} else {
		result.CurrentSlotDurability = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: SlotConsumptionFunction at int64(addr)+896
	{
		child, childErrs := ReadMsvcString(ctx, uintptr(int64(addr)+896))
		if child != nil {
			result.SlotConsumptionFunction = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: NextFrameUsable at int64(addr)+920
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+920); err != nil {
		errs.Add("AbilityComponent.NextFrameUsable", uintptr(int64(addr)+920), err)
	} else {
		result.NextFrameUsable = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: CastDelayStartFrame at int64(addr)+924
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+924); err != nil {
		errs.Add("AbilityComponent.CastDelayStartFrame", uintptr(int64(addr)+924), err)
	} else {
		result.CastDelayStartFrame = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: AmmoLeft at int64(addr)+928
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+928); err != nil {
		errs.Add("AbilityComponent.AmmoLeft", uintptr(int64(addr)+928), err)
	} else {
		result.AmmoLeft = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ReloadFramesLeft at int64(addr)+932
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+932); err != nil {
		errs.Add("AbilityComponent.ReloadFramesLeft", uintptr(int64(addr)+932), err)
	} else {
		result.ReloadFramesLeft = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ReloadNextFrameUsable at int64(addr)+936
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+936); err != nil {
		errs.Add("AbilityComponent.ReloadNextFrameUsable", uintptr(int64(addr)+936), err)
	} else {
		result.ReloadNextFrameUsable = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ChargeCount at int64(addr)+940
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+940); err != nil {
		errs.Add("AbilityComponent.ChargeCount", uintptr(int64(addr)+940), err)
	} else {
		result.ChargeCount = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: NextChargeFrame at int64(addr)+944
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+944); err != nil {
		errs.Add("AbilityComponent.NextChargeFrame", uintptr(int64(addr)+944), err)
	} else {
		result.NextChargeFrame = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ItemRecoil at int64(addr)+948
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+948); err != nil {
		errs.Add("AbilityComponent.ItemRecoil", uintptr(int64(addr)+948), err)
	} else {
		result.ItemRecoil = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: IsInitialized at int64(addr)+952
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+952); err != nil {
		errs.Add("AbilityComponent.IsInitialized", uintptr(int64(addr)+952), err)
	} else {
		result.IsInitialized = buf[0] != 0
	}

	return result, errs
}

func ReadInventory2Component(ctx *runtime.ReadContext, addr uintptr) (*Inventory2Component, runtime.Errors) {
	var errs runtime.Errors
	result := &Inventory2Component{}
	var buf [4]byte

	// Field: Header at int64(addr)+0
	{
		child, childErrs := ReadComponentHeader(ctx, uintptr(int64(addr)+0))
		if child != nil {
			result.Header = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: QuickInventorySlots at int64(addr)+72
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+72); err != nil {
		errs.Add("Inventory2Component.QuickInventorySlots", uintptr(int64(addr)+72), err)
	} else {
		result.QuickInventorySlots = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: FullInventorySlotsX at int64(addr)+76
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+76); err != nil {
		errs.Add("Inventory2Component.FullInventorySlotsX", uintptr(int64(addr)+76), err)
	} else {
		result.FullInventorySlotsX = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: FullInventorySlotsY at int64(addr)+80
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+80); err != nil {
		errs.Add("Inventory2Component.FullInventorySlotsY", uintptr(int64(addr)+80), err)
	} else {
		result.FullInventorySlotsY = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: SavedActiveItemIndex at int64(addr)+84
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+84); err != nil {
		errs.Add("Inventory2Component.SavedActiveItemIndex", uintptr(int64(addr)+84), err)
	} else {
		result.SavedActiveItemIndex = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ActiveItem at int64(addr)+88
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+88); err != nil {
		errs.Add("Inventory2Component.ActiveItem", uintptr(int64(addr)+88), err)
	} else {
		result.ActiveItem = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ActualActiveItem at int64(addr)+92
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+92); err != nil {
		errs.Add("Inventory2Component.ActualActiveItem", uintptr(int64(addr)+92), err)
	} else {
		result.ActualActiveItem = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ActiveStash at int64(addr)+96
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+96); err != nil {
		errs.Add("Inventory2Component.ActiveStash", uintptr(int64(addr)+96), err)
	} else {
		result.ActiveStash = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ThrowItem at int64(addr)+100
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+100); err != nil {
		errs.Add("Inventory2Component.ThrowItem", uintptr(int64(addr)+100), err)
	} else {
		result.ThrowItem = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ItemHolstered at int64(addr)+104
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+104); err != nil {
		errs.Add("Inventory2Component.ItemHolstered", uintptr(int64(addr)+104), err)
	} else {
		result.ItemHolstered = buf[0] != 0
	}

	// Field: Initialized at int64(addr)+105
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+105); err != nil {
		errs.Add("Inventory2Component.Initialized", uintptr(int64(addr)+105), err)
	} else {
		result.Initialized = buf[0] != 0
	}

	// Field: ForceRefresh at int64(addr)+106
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+106); err != nil {
		errs.Add("Inventory2Component.ForceRefresh", uintptr(int64(addr)+106), err)
	} else {
		result.ForceRefresh = buf[0] != 0
	}

	// Field: DontLogNextItemEquip at int64(addr)+107
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+107); err != nil {
		errs.Add("Inventory2Component.DontLogNextItemEquip", uintptr(int64(addr)+107), err)
	} else {
		result.DontLogNextItemEquip = buf[0] != 0
	}

	// Field: SmoothedItemXOffset at int64(addr)+108
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+108); err != nil {
		errs.Add("Inventory2Component.SmoothedItemXOffset", uintptr(int64(addr)+108), err)
	} else {
		result.SmoothedItemXOffset = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: LastItemSwitchFrame at int64(addr)+112
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+112); err != nil {
		errs.Add("Inventory2Component.LastItemSwitchFrame", uintptr(int64(addr)+112), err)
	} else {
		result.LastItemSwitchFrame = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: IntroEquipItemLerp at int64(addr)+116
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+116); err != nil {
		errs.Add("Inventory2Component.IntroEquipItemLerp", uintptr(int64(addr)+116), err)
	} else {
		result.IntroEquipItemLerp = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: SmoothedItemAngleX at int64(addr)+120
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+120); err != nil {
		errs.Add("Inventory2Component.SmoothedItemAngleX", uintptr(int64(addr)+120), err)
	} else {
		result.SmoothedItemAngleX = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: SmoothedItemAngleY at int64(addr)+124
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

	// Field: ViewX at int64(addr)+0
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+0); err != nil {
		errs.Add("WorldManagerViewRect.ViewX", uintptr(int64(addr)+0), err)
	} else {
		result.ViewX = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ViewY at int64(addr)+4
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+4); err != nil {
		errs.Add("WorldManagerViewRect.ViewY", uintptr(int64(addr)+4), err)
	} else {
		result.ViewY = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ViewWidth at int64(addr)+8
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+8); err != nil {
		errs.Add("WorldManagerViewRect.ViewWidth", uintptr(int64(addr)+8), err)
	} else {
		result.ViewWidth = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ViewHeight at int64(addr)+12
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+12); err != nil {
		errs.Add("WorldManagerViewRect.ViewHeight", uintptr(int64(addr)+12), err)
	} else {
		result.ViewHeight = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	return result, errs
}

func ReadCellFactory(ctx *runtime.ReadContext, addr uintptr) (*CellFactory, runtime.Errors) {
	var errs runtime.Errors
	result := &CellFactory{}
	var buf [4]byte

	// Field: CellDataArrayPtr at int64(addr)+24
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+24); err != nil {
		errs.Add("CellFactory.CellDataArrayPtr", uintptr(int64(addr)+24), err)
	} else {
		result.CellDataArrayPtr = binary.LittleEndian.Uint32(buf[:4])
	}

	return result, errs
}

func ReadS32Vector(ctx *runtime.ReadContext, addr uintptr) (*S32Vector, runtime.Errors) {
	var errs runtime.Errors
	result := &S32Vector{}
	var buf [4]byte
	offset := int64(0)

	// Field: BeginPtr at int64(addr)+offset
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+offset); err != nil {
		errs.Add("S32Vector.BeginPtr", uintptr(int64(addr)+offset), err)
	} else {
		result.BeginPtr = binary.LittleEndian.Uint32(buf[:4])
	}

	offset += 4
	// Field: EndPtr at int64(addr)+offset
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+offset); err != nil {
		errs.Add("S32Vector.EndPtr", uintptr(int64(addr)+offset), err)
	} else {
		result.EndPtr = binary.LittleEndian.Uint32(buf[:4])
	}

	offset += 4
	// Field: CapacityPtr at int64(addr)+offset
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+offset); err != nil {
		errs.Add("S32Vector.CapacityPtr", uintptr(int64(addr)+offset), err)
	} else {
		result.CapacityPtr = binary.LittleEndian.Uint32(buf[:4])
	}

	offset += 4
	// Field: Elements (dynamic array) at int64(result.BeginPtr)
	result.Elements = make([]int32, int(((result.EndPtr - result.BeginPtr) / 4)))
	for i := range result.Elements {
		if _, err := ctx.ReadAt(buf[:4], int64(result.BeginPtr)+int64(i)*4); err != nil {
			errs.Add("S32Vector.Elements", uintptr(int64(result.BeginPtr)+int64(i)*4), err)
		} else {
			result.Elements[i] = int32(binary.LittleEndian.Uint32(buf[:4]))
		}
	}

	return result, errs
}

func ReadChildrenContainer(ctx *runtime.ReadContext, addr uintptr) (*ChildrenContainer, runtime.Errors) {
	var errs runtime.Errors
	result := &ChildrenContainer{}
	var buf [4]byte
	offset := int64(0)

	// Field: BeginPtr at int64(addr)+offset
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+offset); err != nil {
		errs.Add("ChildrenContainer.BeginPtr", uintptr(int64(addr)+offset), err)
	} else {
		result.BeginPtr = binary.LittleEndian.Uint32(buf[:4])
	}

	offset += 4
	// Field: EndPtr at int64(addr)+offset
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+offset); err != nil {
		errs.Add("ChildrenContainer.EndPtr", uintptr(int64(addr)+offset), err)
	} else {
		result.EndPtr = binary.LittleEndian.Uint32(buf[:4])
	}

	offset += 4
	// Field: CapacityPtr at int64(addr)+offset
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+offset); err != nil {
		errs.Add("ChildrenContainer.CapacityPtr", uintptr(int64(addr)+offset), err)
	} else {
		result.CapacityPtr = binary.LittleEndian.Uint32(buf[:4])
	}

	offset += 4
	// Field: Children (dynamic array) at int64(result.BeginPtr)
	result.Children = make([]uint32, int(((result.EndPtr - result.BeginPtr) / 4)))
	for i := range result.Children {
		if _, err := ctx.ReadAt(buf[:4], int64(result.BeginPtr)+int64(i)*4); err != nil {
			errs.Add("ChildrenContainer.Children", uintptr(int64(result.BeginPtr)+int64(i)*4), err)
		} else {
			result.Children[i] = binary.LittleEndian.Uint32(buf[:4])
		}
	}

	return result, errs
}

func ReadStdVectorHeader(ctx *runtime.ReadContext, addr uintptr) (*StdVectorHeader, runtime.Errors) {
	var errs runtime.Errors
	result := &StdVectorHeader{}
	var buf [4]byte

	// Field: BeginPtr at int64(addr)+0
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+0); err != nil {
		errs.Add("StdVectorHeader.BeginPtr", uintptr(int64(addr)+0), err)
	} else {
		result.BeginPtr = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: EndPtr at int64(addr)+4
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+4); err != nil {
		errs.Add("StdVectorHeader.EndPtr", uintptr(int64(addr)+4), err)
	} else {
		result.EndPtr = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: CapacityPtr at int64(addr)+8
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+8); err != nil {
		errs.Add("StdVectorHeader.CapacityPtr", uintptr(int64(addr)+8), err)
	} else {
		result.CapacityPtr = binary.LittleEndian.Uint32(buf[:4])
	}

	return result, errs
}

func ReadComponentBuffer(ctx *runtime.ReadContext, addr uintptr) (*ComponentBuffer, runtime.Errors) {
	var errs runtime.Errors
	result := &ComponentBuffer{}
	var buf [4]byte

	// Field: Vtable at int64(addr)+0
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+0); err != nil {
		errs.Add("ComponentBuffer.Vtable", uintptr(int64(addr)+0), err)
	} else {
		result.Vtable = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: Sentinel at int64(addr)+4
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+4); err != nil {
		errs.Add("ComponentBuffer.Sentinel", uintptr(int64(addr)+4), err)
	} else {
		result.Sentinel = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: InitialCapacity at int64(addr)+8
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+8); err != nil {
		errs.Add("ComponentBuffer.InitialCapacity", uintptr(int64(addr)+8), err)
	} else {
		result.InitialCapacity = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: Unknown0c at int64(addr)+12
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+12); err != nil {
		errs.Add("ComponentBuffer.Unknown0c", uintptr(int64(addr)+12), err)
	} else {
		result.Unknown0c = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: SparseIndex at int64(addr)+16
	{
		child, childErrs := ReadStdVectorHeader(ctx, uintptr(int64(addr)+16))
		if child != nil {
			result.SparseIndex = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: EntityRefs at int64(addr)+28
	{
		child, childErrs := ReadStdVectorHeader(ctx, uintptr(int64(addr)+28))
		if child != nil {
			result.EntityRefs = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: PrevIndex at int64(addr)+40
	{
		child, childErrs := ReadStdVectorHeader(ctx, uintptr(int64(addr)+40))
		if child != nil {
			result.PrevIndex = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: NextIndex at int64(addr)+52
	{
		child, childErrs := ReadStdVectorHeader(ctx, uintptr(int64(addr)+52))
		if child != nil {
			result.NextIndex = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: Components at int64(addr)+64
	{
		child, childErrs := ReadStdVectorHeader(ctx, uintptr(int64(addr)+64))
		if child != nil {
			result.Components = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: HandleMap at int64(addr)+96
	{
		child, childErrs := ReadStdVectorHeader(ctx, uintptr(int64(addr)+96))
		if child != nil {
			result.HandleMap = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: Generations at int64(addr)+108
	{
		child, childErrs := ReadStdVectorHeader(ctx, uintptr(int64(addr)+108))
		if child != nil {
			result.Generations = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: ReverseHandleMap at int64(addr)+120
	{
		child, childErrs := ReadStdVectorHeader(ctx, uintptr(int64(addr)+120))
		if child != nil {
			result.ReverseHandleMap = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: ActiveCount at int64(addr)+152
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+152); err != nil {
		errs.Add("ComponentBuffer.ActiveCount", uintptr(int64(addr)+152), err)
	} else {
		result.ActiveCount = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: CapacityLimit at int64(addr)+156
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+156); err != nil {
		errs.Add("ComponentBuffer.CapacityLimit", uintptr(int64(addr)+156), err)
	} else {
		result.CapacityLimit = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: UnknownA0 at int64(addr)+160
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+160); err != nil {
		errs.Add("ComponentBuffer.UnknownA0", uintptr(int64(addr)+160), err)
	} else {
		result.UnknownA0 = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: PEntityManager at int64(addr)+164
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+164); err != nil {
		errs.Add("ComponentBuffer.PEntityManager", uintptr(int64(addr)+164), err)
	} else {
		result.PEntityManager = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: PEventManager at int64(addr)+168
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+168); err != nil {
		errs.Add("ComponentBuffer.PEventManager", uintptr(int64(addr)+168), err)
	} else {
		result.PEventManager = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: NameString at int64(addr)+172
	{
		child, childErrs := ReadMsvcString(ctx, uintptr(int64(addr)+172))
		if child != nil {
			result.NameString = *child
		}
		errs = append(errs, childErrs...)
	}

	return result, errs
}

func ReadWorldStateComponent(ctx *runtime.ReadContext, addr uintptr) (*WorldStateComponent, runtime.Errors) {
	var errs runtime.Errors
	result := &WorldStateComponent{}
	var buf [4]byte

	// Field: Header at int64(addr)+0
	{
		child, childErrs := ReadComponentHeader(ctx, uintptr(int64(addr)+0))
		if child != nil {
			result.Header = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: BiomeCryptCount at int64(addr)+264
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+264); err != nil {
		errs.Add("WorldStateComponent.BiomeCryptCount", uintptr(int64(addr)+264), err)
	} else {
		result.BiomeCryptCount = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: GodsAfraid at int64(addr)+268
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+268); err != nil {
		errs.Add("WorldStateComponent.GodsAfraid", uintptr(int64(addr)+268), err)
	} else {
		result.GodsAfraid = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: GodsImpressed at int64(addr)+272
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+272); err != nil {
		errs.Add("WorldStateComponent.GodsImpressed", uintptr(int64(addr)+272), err)
	} else {
		result.GodsImpressed = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: GodsAfraidDamage at int64(addr)+276
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+276); err != nil {
		errs.Add("WorldStateComponent.GodsAfraidDamage", uintptr(int64(addr)+276), err)
	} else {
		result.GodsAfraidDamage = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: GodsEnraged at int64(addr)+280
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+280); err != nil {
		errs.Add("WorldStateComponent.GodsEnraged", uintptr(int64(addr)+280), err)
	} else {
		result.GodsEnraged = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	return result, errs
}

func ReadGameGlobals(ctx *runtime.ReadContext, addr uintptr) (*GameGlobals, runtime.Errors) {
	var errs runtime.Errors
	result := &GameGlobals{}
	var buf [4]byte

	// Field: FrameCount at int64(addr)+0
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+0); err != nil {
		errs.Add("GameGlobals.FrameCount", uintptr(int64(addr)+0), err)
	} else {
		result.FrameCount = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: PhysicsStepCount at int64(addr)+4
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+4); err != nil {
		errs.Add("GameGlobals.PhysicsStepCount", uintptr(int64(addr)+4), err)
	} else {
		result.PhysicsStepCount = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: GameTime at int64(addr)+8
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+8); err != nil {
		errs.Add("GameGlobals.GameTime", uintptr(int64(addr)+8), err)
	} else {
		result.GameTime = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: PWorldManager at int64(addr)+12
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+12); err != nil {
		errs.Add("GameGlobals.PWorldManager", uintptr(int64(addr)+12), err)
	} else {
		result.PWorldManager = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: PChunkSystem at int64(addr)+16
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+16); err != nil {
		errs.Add("GameGlobals.PChunkSystem", uintptr(int64(addr)+16), err)
	} else {
		result.PChunkSystem = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: PCellGrid at int64(addr)+20
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+20); err != nil {
		errs.Add("GameGlobals.PCellGrid", uintptr(int64(addr)+20), err)
	} else {
		result.PCellGrid = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: PCellFactory at int64(addr)+24
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+24); err != nil {
		errs.Add("GameGlobals.PCellFactory", uintptr(int64(addr)+24), err)
	} else {
		result.PCellFactory = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: Unknown1c at int64(addr)+28
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+28); err != nil {
		errs.Add("GameGlobals.Unknown1c", uintptr(int64(addr)+28), err)
	} else {
		result.Unknown1c = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: PPhysicsWorld at int64(addr)+32
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+32); err != nil {
		errs.Add("GameGlobals.PPhysicsWorld", uintptr(int64(addr)+32), err)
	} else {
		result.PPhysicsWorld = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: PAudioManager at int64(addr)+36
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+36); err != nil {
		errs.Add("GameGlobals.PAudioManager", uintptr(int64(addr)+36), err)
	} else {
		result.PAudioManager = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: ViewportLeft at int64(addr)+384
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+384); err != nil {
		errs.Add("GameGlobals.ViewportLeft", uintptr(int64(addr)+384), err)
	} else {
		result.ViewportLeft = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ViewportTop at int64(addr)+388
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+388); err != nil {
		errs.Add("GameGlobals.ViewportTop", uintptr(int64(addr)+388), err)
	} else {
		result.ViewportTop = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ViewportRight at int64(addr)+392
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+392); err != nil {
		errs.Add("GameGlobals.ViewportRight", uintptr(int64(addr)+392), err)
	} else {
		result.ViewportRight = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ViewportBottom at int64(addr)+396
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+396); err != nil {
		errs.Add("GameGlobals.ViewportBottom", uintptr(int64(addr)+396), err)
	} else {
		result.ViewportBottom = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	return result, errs
}

func ReadEntity(ctx *runtime.ReadContext, addr uintptr) (*Entity, runtime.Errors) {
	var errs runtime.Errors
	result := &Entity{}
	var buf [4]byte

	// Field: EntityId at int64(addr)+0
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+0); err != nil {
		errs.Add("Entity.EntityId", uintptr(int64(addr)+0), err)
	} else {
		result.EntityId = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: SlotIndex at int64(addr)+4
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+4); err != nil {
		errs.Add("Entity.SlotIndex", uintptr(int64(addr)+4), err)
	} else {
		result.SlotIndex = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: Unknown08 at int64(addr)+8
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+8); err != nil {
		errs.Add("Entity.Unknown08", uintptr(int64(addr)+8), err)
	} else {
		result.Unknown08 = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: PendingKill at int64(addr)+12
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+12); err != nil {
		errs.Add("Entity.PendingKill", uintptr(int64(addr)+12), err)
	} else {
		result.PendingKill = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: Flags10 at int64(addr)+16
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+16); err != nil {
		errs.Add("Entity.Flags10", uintptr(int64(addr)+16), err)
	} else {
		result.Flags10 = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: Name at int64(addr)+20
	{
		child, childErrs := ReadMsvcString(ctx, uintptr(int64(addr)+20))
		if child != nil {
			result.Name = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: Unknown2c at int64(addr)+44
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+44); err != nil {
		errs.Add("Entity.Unknown2c", uintptr(int64(addr)+44), err)
	} else {
		result.Unknown2c = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: TagBitset (array[64]) at int64(addr)+48
	if _, err := ctx.ReadAt(result.TagBitset[:], int64(addr)+48); err != nil {
		errs.Add("Entity.TagBitset", uintptr(int64(addr)+48), err)
	}

	// Field: PosX at int64(addr)+112
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+112); err != nil {
		errs.Add("Entity.PosX", uintptr(int64(addr)+112), err)
	} else {
		result.PosX = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: PosY at int64(addr)+116
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+116); err != nil {
		errs.Add("Entity.PosY", uintptr(int64(addr)+116), err)
	} else {
		result.PosY = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: RotCos at int64(addr)+120
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+120); err != nil {
		errs.Add("Entity.RotCos", uintptr(int64(addr)+120), err)
	} else {
		result.RotCos = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: RotSin at int64(addr)+124
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+124); err != nil {
		errs.Add("Entity.RotSin", uintptr(int64(addr)+124), err)
	} else {
		result.RotSin = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: RotNegSin at int64(addr)+128
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+128); err != nil {
		errs.Add("Entity.RotNegSin", uintptr(int64(addr)+128), err)
	} else {
		result.RotNegSin = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: RotCos2 at int64(addr)+132
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+132); err != nil {
		errs.Add("Entity.RotCos2", uintptr(int64(addr)+132), err)
	} else {
		result.RotCos2 = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ScaleX at int64(addr)+136
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+136); err != nil {
		errs.Add("Entity.ScaleX", uintptr(int64(addr)+136), err)
	} else {
		result.ScaleX = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ScaleY at int64(addr)+140
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+140); err != nil {
		errs.Add("Entity.ScaleY", uintptr(int64(addr)+140), err)
	} else {
		result.ScaleY = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: ChildrenPtr at int64(addr)+144
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+144); err != nil {
		errs.Add("Entity.ChildrenPtr", uintptr(int64(addr)+144), err)
	} else {
		result.ChildrenPtr = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: ParentEntityPtr at int64(addr)+148
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+148); err != nil {
		errs.Add("Entity.ParentEntityPtr", uintptr(int64(addr)+148), err)
	} else {
		result.ParentEntityPtr = binary.LittleEndian.Uint32(buf[:4])
	}

	return result, errs
}

func ReadEntityManager(ctx *runtime.ReadContext, addr uintptr) (*EntityManager, runtime.Errors) {
	var errs runtime.Errors
	result := &EntityManager{}
	var buf [4]byte

	// Field: Vtable at int64(addr)+0
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+0); err != nil {
		errs.Add("EntityManager.Vtable", uintptr(int64(addr)+0), err)
	} else {
		result.Vtable = binary.LittleEndian.Uint32(buf[:4])
	}

	// Field: NextEntityId at int64(addr)+4
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+4); err != nil {
		errs.Add("EntityManager.NextEntityId", uintptr(int64(addr)+4), err)
	} else {
		result.NextEntityId = int32(binary.LittleEndian.Uint32(buf[:4]))
	}

	// Field: FreeSlotStack at int64(addr)+8
	{
		child, childErrs := ReadStdVectorHeader(ctx, uintptr(int64(addr)+8))
		if child != nil {
			result.FreeSlotStack = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: EntityArray at int64(addr)+20
	{
		child, childErrs := ReadStdVectorHeader(ctx, uintptr(int64(addr)+20))
		if child != nil {
			result.EntityArray = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: TagGroups at int64(addr)+32
	{
		child, childErrs := ReadStdVectorHeader(ctx, uintptr(int64(addr)+32))
		if child != nil {
			result.TagGroups = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: ComponentBuffers at int64(addr)+44
	{
		child, childErrs := ReadU32Vector(ctx, uintptr(int64(addr)+44))
		if child != nil {
			result.ComponentBuffers = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: PEventManager at int64(addr)+56
	if _, err := ctx.ReadAt(buf[:4], int64(addr)+56); err != nil {
		errs.Add("EntityManager.PEventManager", uintptr(int64(addr)+56), err)
	} else {
		result.PEventManager = binary.LittleEndian.Uint32(buf[:4])
	}

	return result, errs
}

func ReadWalletComponent(ctx *runtime.ReadContext, addr uintptr) (*WalletComponent, runtime.Errors) {
	var errs runtime.Errors
	result := &WalletComponent{}
	var buf [8]byte

	// Field: Header at int64(addr)+0
	{
		child, childErrs := ReadComponentHeader(ctx, uintptr(int64(addr)+0))
		if child != nil {
			result.Header = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: Money at int64(addr)+72
	if _, err := ctx.ReadAt(buf[:8], int64(addr)+72); err != nil {
		errs.Add("WalletComponent.Money", uintptr(int64(addr)+72), err)
	} else {
		result.Money = int64(binary.LittleEndian.Uint64(buf[:8]))
	}

	// Field: MoneySpent at int64(addr)+80
	if _, err := ctx.ReadAt(buf[:8], int64(addr)+80); err != nil {
		errs.Add("WalletComponent.MoneySpent", uintptr(int64(addr)+80), err)
	} else {
		result.MoneySpent = int64(binary.LittleEndian.Uint64(buf[:8]))
	}

	// Field: MoneyPrevFrame at int64(addr)+88
	if _, err := ctx.ReadAt(buf[:8], int64(addr)+88); err != nil {
		errs.Add("WalletComponent.MoneyPrevFrame", uintptr(int64(addr)+88), err)
	} else {
		result.MoneyPrevFrame = int64(binary.LittleEndian.Uint64(buf[:8]))
	}

	// Field: HasReachedInf at int64(addr)+96
	if _, err := ctx.ReadAt(buf[:1], int64(addr)+96); err != nil {
		errs.Add("WalletComponent.HasReachedInf", uintptr(int64(addr)+96), err)
	} else {
		result.HasReachedInf = buf[0] != 0
	}

	return result, errs
}

func ReadMaterialInventoryComponent(ctx *runtime.ReadContext, addr uintptr) (*MaterialInventoryComponent, runtime.Errors) {
	var errs runtime.Errors
	result := &MaterialInventoryComponent{}

	// Field: Header at int64(addr)+0
	{
		child, childErrs := ReadComponentHeader(ctx, uintptr(int64(addr)+0))
		if child != nil {
			result.Header = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: CountPerMaterialType at int64(addr)+128
	{
		child, childErrs := ReadF64Vector(ctx, uintptr(int64(addr)+128))
		if child != nil {
			result.CountPerMaterialType = *child
		}
		errs = append(errs, childErrs...)
	}

	return result, errs
}

// MsvcStringReader provides lazy, field-level access to MsvcString without reading the entire struct.
type MsvcStringReader struct {
	ctx  *runtime.ReadContext
	addr uintptr
}

// NewMsvcStringReader creates a lazy reader for MsvcString at the given address.
func NewMsvcStringReader(ctx *runtime.ReadContext, addr uintptr) *MsvcStringReader {
	return &MsvcStringReader{ctx: ctx, addr: addr}
}

// Addr returns the base address of this MsvcString.
func (r *MsvcStringReader) Addr() uintptr {
	return r.addr
}

// Read materializes the full MsvcString struct eagerly.
func (r *MsvcStringReader) Read() (*MsvcString, runtime.Errors) {
	return ReadMsvcString(r.ctx, r.addr)
}

func (r *MsvcStringReader) Data() ([16]byte, error) {
	var result [16]byte
	if _, err := r.ctx.ReadAt(result[:], int64(r.addr)+0); err != nil {
		return result, err
	}
	return result, nil
}

func (r *MsvcStringReader) Length() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+16); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *MsvcStringReader) Capacity() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+20); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

// CellDataReader provides lazy, field-level access to CellData without reading the entire struct.
type CellDataReader struct {
	ctx  *runtime.ReadContext
	addr uintptr
}

// NewCellDataReader creates a lazy reader for CellData at the given address.
func NewCellDataReader(ctx *runtime.ReadContext, addr uintptr) *CellDataReader {
	return &CellDataReader{ctx: ctx, addr: addr}
}

// Addr returns the base address of this CellData.
func (r *CellDataReader) Addr() uintptr {
	return r.addr
}

// Read materializes the full CellData struct eagerly.
func (r *CellDataReader) Read() (*CellData, runtime.Errors) {
	return ReadCellData(r.ctx, r.addr)
}

// Name returns a lazy reader for the nested MsvcString (zero I/O).
func (r *CellDataReader) Name() *MsvcStringReader {
	return NewMsvcStringReader(r.ctx, uintptr(int64(r.addr)+0))
}

// DeathMatchAppReader provides lazy, field-level access to DeathMatchApp without reading the entire struct.
type DeathMatchAppReader struct {
	ctx  *runtime.ReadContext
	addr uintptr
}

// NewDeathMatchAppReader creates a lazy reader for DeathMatchApp at the given address.
func NewDeathMatchAppReader(ctx *runtime.ReadContext, addr uintptr) *DeathMatchAppReader {
	return &DeathMatchAppReader{ctx: ctx, addr: addr}
}

// Addr returns the base address of this DeathMatchApp.
func (r *DeathMatchAppReader) Addr() uintptr {
	return r.addr
}

// Read materializes the full DeathMatchApp struct eagerly.
func (r *DeathMatchAppReader) Read() (*DeathMatchApp, runtime.Errors) {
	return ReadDeathMatchApp(r.ctx, r.addr)
}

// PlayerEntities eagerly reads the nested U32Vector (no lazy reader available for this type).
func (r *DeathMatchAppReader) PlayerEntities() (*U32Vector, runtime.Errors) {
	return ReadU32Vector(r.ctx, uintptr(int64(r.addr)+88))
}

// ComponentHeaderReader provides lazy, field-level access to ComponentHeader without reading the entire struct.
type ComponentHeaderReader struct {
	ctx  *runtime.ReadContext
	addr uintptr
}

// NewComponentHeaderReader creates a lazy reader for ComponentHeader at the given address.
func NewComponentHeaderReader(ctx *runtime.ReadContext, addr uintptr) *ComponentHeaderReader {
	return &ComponentHeaderReader{ctx: ctx, addr: addr}
}

// Addr returns the base address of this ComponentHeader.
func (r *ComponentHeaderReader) Addr() uintptr {
	return r.addr
}

// Read materializes the full ComponentHeader struct eagerly.
func (r *ComponentHeaderReader) Read() (*ComponentHeader, runtime.Errors) {
	return ReadComponentHeader(r.ctx, r.addr)
}

func (r *ComponentHeaderReader) Vtable() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+0); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *ComponentHeaderReader) BufferIndex() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+4); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *ComponentHeaderReader) PTypeName() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+8); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *ComponentHeaderReader) TypeId() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+12); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *ComponentHeaderReader) Unknown10() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+16); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *ComponentHeaderReader) Active() (bool, error) {
	var buf [1]byte
	if _, err := r.ctx.ReadAt(buf[:1], int64(r.addr)+20); err != nil {
		return false, err
	}
	return buf[0] != 0, nil
}

func (r *ComponentHeaderReader) ComponentTags() ([32]uint8, error) {
	var result [32]uint8
	if _, err := r.ctx.ReadAt(result[:], int64(r.addr)+24); err != nil {
		return result, err
	}
	return result, nil
}

func (r *ComponentHeaderReader) Unknown38() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+56); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *ComponentHeaderReader) Unknown3c() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+60); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *ComponentHeaderReader) Unknown40() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+64); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *ComponentHeaderReader) Unknown44() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+68); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

// DamageModelComponentReader provides lazy, field-level access to DamageModelComponent without reading the entire struct.
type DamageModelComponentReader struct {
	ctx  *runtime.ReadContext
	addr uintptr
}

// NewDamageModelComponentReader creates a lazy reader for DamageModelComponent at the given address.
func NewDamageModelComponentReader(ctx *runtime.ReadContext, addr uintptr) *DamageModelComponentReader {
	return &DamageModelComponentReader{ctx: ctx, addr: addr}
}

// Addr returns the base address of this DamageModelComponent.
func (r *DamageModelComponentReader) Addr() uintptr {
	return r.addr
}

// Read materializes the full DamageModelComponent struct eagerly.
func (r *DamageModelComponentReader) Read() (*DamageModelComponent, runtime.Errors) {
	return ReadDamageModelComponent(r.ctx, r.addr)
}

// Header returns a lazy reader for the nested ComponentHeader (zero I/O).
func (r *DamageModelComponentReader) Header() *ComponentHeaderReader {
	return NewComponentHeaderReader(r.ctx, uintptr(int64(r.addr)+0))
}

func (r *DamageModelComponentReader) Hp() (float64, error) {
	var buf [8]byte
	if _, err := r.ctx.ReadAt(buf[:8], int64(r.addr)+72); err != nil {
		return 0, err
	}
	return math.Float64frombits(binary.LittleEndian.Uint64(buf[:8])), nil
}

func (r *DamageModelComponentReader) MaxHp() (float64, error) {
	var buf [8]byte
	if _, err := r.ctx.ReadAt(buf[:8], int64(r.addr)+80); err != nil {
		return 0, err
	}
	return math.Float64frombits(binary.LittleEndian.Uint64(buf[:8])), nil
}

func (r *DamageModelComponentReader) MaxHpCap() (float64, error) {
	var buf [8]byte
	if _, err := r.ctx.ReadAt(buf[:8], int64(r.addr)+88); err != nil {
		return 0, err
	}
	return math.Float64frombits(binary.LittleEndian.Uint64(buf[:8])), nil
}

func (r *DamageModelComponentReader) MaxHpOld() (float64, error) {
	var buf [8]byte
	if _, err := r.ctx.ReadAt(buf[:8], int64(r.addr)+96); err != nil {
		return 0, err
	}
	return math.Float64frombits(binary.LittleEndian.Uint64(buf[:8])), nil
}

func (r *DamageModelComponentReader) Unknown68() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+104); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *DamageModelComponentReader) DmgMultMelee() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+108); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *DamageModelComponentReader) DmgMultProjectile() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+112); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *DamageModelComponentReader) DmgMultExplosion() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+116); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *DamageModelComponentReader) DmgMultElectricity() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+120); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *DamageModelComponentReader) DmgMultFire() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+124); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *DamageModelComponentReader) DmgMultDrill() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+128); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *DamageModelComponentReader) DmgMultSlice() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+132); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *DamageModelComponentReader) DmgMultIce() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+136); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *DamageModelComponentReader) DmgMultHealing() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+140); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *DamageModelComponentReader) DmgMultPhysicsHit() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+144); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *DamageModelComponentReader) DmgMultRadioactive() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+148); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *DamageModelComponentReader) DmgMultPoison() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+152); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *DamageModelComponentReader) DmgMultHoly() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+156); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *DamageModelComponentReader) DmgMultCurse() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+160); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *DamageModelComponentReader) DmgMultOvereating() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+164); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *DamageModelComponentReader) DmgMultMaterial() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+168); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *DamageModelComponentReader) InvincibilityFrames() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+172); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

// CharacterDataComponentReader provides lazy, field-level access to CharacterDataComponent without reading the entire struct.
type CharacterDataComponentReader struct {
	ctx  *runtime.ReadContext
	addr uintptr
}

// NewCharacterDataComponentReader creates a lazy reader for CharacterDataComponent at the given address.
func NewCharacterDataComponentReader(ctx *runtime.ReadContext, addr uintptr) *CharacterDataComponentReader {
	return &CharacterDataComponentReader{ctx: ctx, addr: addr}
}

// Addr returns the base address of this CharacterDataComponent.
func (r *CharacterDataComponentReader) Addr() uintptr {
	return r.addr
}

// Read materializes the full CharacterDataComponent struct eagerly.
func (r *CharacterDataComponentReader) Read() (*CharacterDataComponent, runtime.Errors) {
	return ReadCharacterDataComponent(r.ctx, r.addr)
}

// Header returns a lazy reader for the nested ComponentHeader (zero I/O).
func (r *CharacterDataComponentReader) Header() *ComponentHeaderReader {
	return NewComponentHeaderReader(r.ctx, uintptr(int64(r.addr)+0))
}

func (r *CharacterDataComponentReader) Gravity() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+136); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *CharacterDataComponentReader) FlyTime() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+140); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *CharacterDataComponentReader) IsOnGround() (bool, error) {
	var buf [1]byte
	if _, err := r.ctx.ReadAt(buf[:1], int64(r.addr)+184); err != nil {
		return false, err
	}
	return buf[0] != 0, nil
}

func (r *CharacterDataComponentReader) VelocityX() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+264); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *CharacterDataComponentReader) VelocityY() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+268); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

// ConfigGunReader provides lazy, field-level access to ConfigGun without reading the entire struct.
type ConfigGunReader struct {
	ctx  *runtime.ReadContext
	addr uintptr
}

// NewConfigGunReader creates a lazy reader for ConfigGun at the given address.
func NewConfigGunReader(ctx *runtime.ReadContext, addr uintptr) *ConfigGunReader {
	return &ConfigGunReader{ctx: ctx, addr: addr}
}

// Addr returns the base address of this ConfigGun.
func (r *ConfigGunReader) Addr() uintptr {
	return r.addr
}

// Read materializes the full ConfigGun struct eagerly.
func (r *ConfigGunReader) Read() (*ConfigGun, runtime.Errors) {
	return ReadConfigGun(r.ctx, r.addr)
}

func (r *ConfigGunReader) Vtable() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+0); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *ConfigGunReader) ActionsPerRound() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+4); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *ConfigGunReader) ShuffleDeckWhenEmpty() (bool, error) {
	var buf [1]byte
	if _, err := r.ctx.ReadAt(buf[:1], int64(r.addr)+8); err != nil {
		return false, err
	}
	return buf[0] != 0, nil
}

func (r *ConfigGunReader) ReloadTime() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+12); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *ConfigGunReader) DeckCapacity() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+16); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

// AbilityComponentReader provides lazy, field-level access to AbilityComponent without reading the entire struct.
type AbilityComponentReader struct {
	ctx  *runtime.ReadContext
	addr uintptr
}

// NewAbilityComponentReader creates a lazy reader for AbilityComponent at the given address.
func NewAbilityComponentReader(ctx *runtime.ReadContext, addr uintptr) *AbilityComponentReader {
	return &AbilityComponentReader{ctx: ctx, addr: addr}
}

// Addr returns the base address of this AbilityComponent.
func (r *AbilityComponentReader) Addr() uintptr {
	return r.addr
}

// Read materializes the full AbilityComponent struct eagerly.
func (r *AbilityComponentReader) Read() (*AbilityComponent, runtime.Errors) {
	return ReadAbilityComponent(r.ctx, r.addr)
}

// Header returns a lazy reader for the nested ComponentHeader (zero I/O).
func (r *AbilityComponentReader) Header() *ComponentHeaderReader {
	return NewComponentHeaderReader(r.ctx, uintptr(int64(r.addr)+0))
}

func (r *AbilityComponentReader) CooldownFrames() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+72); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

// EntityFile returns a lazy reader for the nested MsvcString (zero I/O).
func (r *AbilityComponentReader) EntityFile() *MsvcStringReader {
	return NewMsvcStringReader(r.ctx, uintptr(int64(r.addr)+76))
}

// SpriteFile returns a lazy reader for the nested MsvcString (zero I/O).
func (r *AbilityComponentReader) SpriteFile() *MsvcStringReader {
	return NewMsvcStringReader(r.ctx, uintptr(int64(r.addr)+100))
}

func (r *AbilityComponentReader) EntityCount() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+124); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *AbilityComponentReader) NeverReload() (bool, error) {
	var buf [1]byte
	if _, err := r.ctx.ReadAt(buf[:1], int64(r.addr)+128); err != nil {
		return false, err
	}
	return buf[0] != 0, nil
}

func (r *AbilityComponentReader) ReloadTimeFrames() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+132); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *AbilityComponentReader) Mana() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+136); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *AbilityComponentReader) ManaMax() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+140); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *AbilityComponentReader) ManaChargeSpeed() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+144); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *AbilityComponentReader) RotateInHand() (bool, error) {
	var buf [1]byte
	if _, err := r.ctx.ReadAt(buf[:1], int64(r.addr)+148); err != nil {
		return false, err
	}
	return buf[0] != 0, nil
}

func (r *AbilityComponentReader) RotateInHandAmount() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+152); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *AbilityComponentReader) RotateHandAmount() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+156); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *AbilityComponentReader) FastProjectile() (bool, error) {
	var buf [1]byte
	if _, err := r.ctx.ReadAt(buf[:1], int64(r.addr)+160); err != nil {
		return false, err
	}
	return buf[0] != 0, nil
}

func (r *AbilityComponentReader) SwimPropelAmount() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+164); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *AbilityComponentReader) MaxChargedActions() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+168); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *AbilityComponentReader) ChargeWaitFrames() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+172); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *AbilityComponentReader) ItemRecoilRecoverySpeed() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+176); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *AbilityComponentReader) ItemRecoilMax() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+180); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *AbilityComponentReader) ItemRecoilOffsetCoeff() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+184); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *AbilityComponentReader) ItemRecoilRotationCoeff() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+188); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

// BaseItemFile returns a lazy reader for the nested MsvcString (zero I/O).
func (r *AbilityComponentReader) BaseItemFile() *MsvcStringReader {
	return NewMsvcStringReader(r.ctx, uintptr(int64(r.addr)+192))
}

func (r *AbilityComponentReader) UseEntityFileAsProjectileInfoProxy() (bool, error) {
	var buf [1]byte
	if _, err := r.ctx.ReadAt(buf[:1], int64(r.addr)+216); err != nil {
		return false, err
	}
	return buf[0] != 0, nil
}

func (r *AbilityComponentReader) ClickToUse() (bool, error) {
	var buf [1]byte
	if _, err := r.ctx.ReadAt(buf[:1], int64(r.addr)+217); err != nil {
		return false, err
	}
	return buf[0] != 0, nil
}

func (r *AbilityComponentReader) StatTimesPlayerHasShot() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+220); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *AbilityComponentReader) StatTimesPlayerHasEdited() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+224); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *AbilityComponentReader) ShootingReducesAmountInInventory() (bool, error) {
	var buf [1]byte
	if _, err := r.ctx.ReadAt(buf[:1], int64(r.addr)+228); err != nil {
		return false, err
	}
	return buf[0] != 0, nil
}

func (r *AbilityComponentReader) ThrowAsItem() (bool, error) {
	var buf [1]byte
	if _, err := r.ctx.ReadAt(buf[:1], int64(r.addr)+229); err != nil {
		return false, err
	}
	return buf[0] != 0, nil
}

func (r *AbilityComponentReader) SimulateThrowAsItem() (bool, error) {
	var buf [1]byte
	if _, err := r.ctx.ReadAt(buf[:1], int64(r.addr)+230); err != nil {
		return false, err
	}
	return buf[0] != 0, nil
}

func (r *AbilityComponentReader) MaxAmountInInventory() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+232); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *AbilityComponentReader) AmountInInventory() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+236); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *AbilityComponentReader) DropAsItemOnDeath() (bool, error) {
	var buf [1]byte
	if _, err := r.ctx.ReadAt(buf[:1], int64(r.addr)+240); err != nil {
		return false, err
	}
	return buf[0] != 0, nil
}

// UiName returns a lazy reader for the nested MsvcString (zero I/O).
func (r *AbilityComponentReader) UiName() *MsvcStringReader {
	return NewMsvcStringReader(r.ctx, uintptr(int64(r.addr)+244))
}

func (r *AbilityComponentReader) UseGunScript() (bool, error) {
	var buf [1]byte
	if _, err := r.ctx.ReadAt(buf[:1], int64(r.addr)+268); err != nil {
		return false, err
	}
	return buf[0] != 0, nil
}

func (r *AbilityComponentReader) IsPetrisGun() (bool, error) {
	var buf [1]byte
	if _, err := r.ctx.ReadAt(buf[:1], int64(r.addr)+269); err != nil {
		return false, err
	}
	return buf[0] != 0, nil
}

// GunConfig returns a lazy reader for the nested ConfigGun (zero I/O).
func (r *AbilityComponentReader) GunConfig() *ConfigGunReader {
	return NewConfigGunReader(r.ctx, uintptr(int64(r.addr)+272))
}

func (r *AbilityComponentReader) GunLevel() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+864); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

// AddTheseChildActions returns a lazy reader for the nested MsvcString (zero I/O).
func (r *AbilityComponentReader) AddTheseChildActions() *MsvcStringReader {
	return NewMsvcStringReader(r.ctx, uintptr(int64(r.addr)+868))
}

func (r *AbilityComponentReader) CurrentSlotDurability() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+892); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

// SlotConsumptionFunction returns a lazy reader for the nested MsvcString (zero I/O).
func (r *AbilityComponentReader) SlotConsumptionFunction() *MsvcStringReader {
	return NewMsvcStringReader(r.ctx, uintptr(int64(r.addr)+896))
}

func (r *AbilityComponentReader) NextFrameUsable() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+920); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *AbilityComponentReader) CastDelayStartFrame() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+924); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *AbilityComponentReader) AmmoLeft() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+928); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *AbilityComponentReader) ReloadFramesLeft() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+932); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *AbilityComponentReader) ReloadNextFrameUsable() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+936); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *AbilityComponentReader) ChargeCount() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+940); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *AbilityComponentReader) NextChargeFrame() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+944); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *AbilityComponentReader) ItemRecoil() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+948); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *AbilityComponentReader) IsInitialized() (bool, error) {
	var buf [1]byte
	if _, err := r.ctx.ReadAt(buf[:1], int64(r.addr)+952); err != nil {
		return false, err
	}
	return buf[0] != 0, nil
}

// Inventory2ComponentReader provides lazy, field-level access to Inventory2Component without reading the entire struct.
type Inventory2ComponentReader struct {
	ctx  *runtime.ReadContext
	addr uintptr
}

// NewInventory2ComponentReader creates a lazy reader for Inventory2Component at the given address.
func NewInventory2ComponentReader(ctx *runtime.ReadContext, addr uintptr) *Inventory2ComponentReader {
	return &Inventory2ComponentReader{ctx: ctx, addr: addr}
}

// Addr returns the base address of this Inventory2Component.
func (r *Inventory2ComponentReader) Addr() uintptr {
	return r.addr
}

// Read materializes the full Inventory2Component struct eagerly.
func (r *Inventory2ComponentReader) Read() (*Inventory2Component, runtime.Errors) {
	return ReadInventory2Component(r.ctx, r.addr)
}

// Header returns a lazy reader for the nested ComponentHeader (zero I/O).
func (r *Inventory2ComponentReader) Header() *ComponentHeaderReader {
	return NewComponentHeaderReader(r.ctx, uintptr(int64(r.addr)+0))
}

func (r *Inventory2ComponentReader) QuickInventorySlots() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+72); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *Inventory2ComponentReader) FullInventorySlotsX() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+76); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *Inventory2ComponentReader) FullInventorySlotsY() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+80); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *Inventory2ComponentReader) SavedActiveItemIndex() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+84); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *Inventory2ComponentReader) ActiveItem() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+88); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *Inventory2ComponentReader) ActualActiveItem() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+92); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *Inventory2ComponentReader) ActiveStash() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+96); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *Inventory2ComponentReader) ThrowItem() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+100); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *Inventory2ComponentReader) ItemHolstered() (bool, error) {
	var buf [1]byte
	if _, err := r.ctx.ReadAt(buf[:1], int64(r.addr)+104); err != nil {
		return false, err
	}
	return buf[0] != 0, nil
}

func (r *Inventory2ComponentReader) Initialized() (bool, error) {
	var buf [1]byte
	if _, err := r.ctx.ReadAt(buf[:1], int64(r.addr)+105); err != nil {
		return false, err
	}
	return buf[0] != 0, nil
}

func (r *Inventory2ComponentReader) ForceRefresh() (bool, error) {
	var buf [1]byte
	if _, err := r.ctx.ReadAt(buf[:1], int64(r.addr)+106); err != nil {
		return false, err
	}
	return buf[0] != 0, nil
}

func (r *Inventory2ComponentReader) DontLogNextItemEquip() (bool, error) {
	var buf [1]byte
	if _, err := r.ctx.ReadAt(buf[:1], int64(r.addr)+107); err != nil {
		return false, err
	}
	return buf[0] != 0, nil
}

func (r *Inventory2ComponentReader) SmoothedItemXOffset() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+108); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *Inventory2ComponentReader) LastItemSwitchFrame() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+112); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *Inventory2ComponentReader) IntroEquipItemLerp() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+116); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *Inventory2ComponentReader) SmoothedItemAngleX() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+120); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *Inventory2ComponentReader) SmoothedItemAngleY() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+124); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

// WorldManagerViewRectReader provides lazy, field-level access to WorldManagerViewRect without reading the entire struct.
type WorldManagerViewRectReader struct {
	ctx  *runtime.ReadContext
	addr uintptr
}

// NewWorldManagerViewRectReader creates a lazy reader for WorldManagerViewRect at the given address.
func NewWorldManagerViewRectReader(ctx *runtime.ReadContext, addr uintptr) *WorldManagerViewRectReader {
	return &WorldManagerViewRectReader{ctx: ctx, addr: addr}
}

// Addr returns the base address of this WorldManagerViewRect.
func (r *WorldManagerViewRectReader) Addr() uintptr {
	return r.addr
}

// Read materializes the full WorldManagerViewRect struct eagerly.
func (r *WorldManagerViewRectReader) Read() (*WorldManagerViewRect, runtime.Errors) {
	return ReadWorldManagerViewRect(r.ctx, r.addr)
}

func (r *WorldManagerViewRectReader) ViewX() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+0); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *WorldManagerViewRectReader) ViewY() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+4); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *WorldManagerViewRectReader) ViewWidth() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+8); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *WorldManagerViewRectReader) ViewHeight() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+12); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

// CellFactoryReader provides lazy, field-level access to CellFactory without reading the entire struct.
type CellFactoryReader struct {
	ctx  *runtime.ReadContext
	addr uintptr
}

// NewCellFactoryReader creates a lazy reader for CellFactory at the given address.
func NewCellFactoryReader(ctx *runtime.ReadContext, addr uintptr) *CellFactoryReader {
	return &CellFactoryReader{ctx: ctx, addr: addr}
}

// Addr returns the base address of this CellFactory.
func (r *CellFactoryReader) Addr() uintptr {
	return r.addr
}

// Read materializes the full CellFactory struct eagerly.
func (r *CellFactoryReader) Read() (*CellFactory, runtime.Errors) {
	return ReadCellFactory(r.ctx, r.addr)
}

func (r *CellFactoryReader) CellDataArrayPtr() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+24); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

// StdVectorHeaderReader provides lazy, field-level access to StdVectorHeader without reading the entire struct.
type StdVectorHeaderReader struct {
	ctx  *runtime.ReadContext
	addr uintptr
}

// NewStdVectorHeaderReader creates a lazy reader for StdVectorHeader at the given address.
func NewStdVectorHeaderReader(ctx *runtime.ReadContext, addr uintptr) *StdVectorHeaderReader {
	return &StdVectorHeaderReader{ctx: ctx, addr: addr}
}

// Addr returns the base address of this StdVectorHeader.
func (r *StdVectorHeaderReader) Addr() uintptr {
	return r.addr
}

// Read materializes the full StdVectorHeader struct eagerly.
func (r *StdVectorHeaderReader) Read() (*StdVectorHeader, runtime.Errors) {
	return ReadStdVectorHeader(r.ctx, r.addr)
}

func (r *StdVectorHeaderReader) BeginPtr() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+0); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *StdVectorHeaderReader) EndPtr() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+4); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *StdVectorHeaderReader) CapacityPtr() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+8); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

// ComponentBufferReader provides lazy, field-level access to ComponentBuffer without reading the entire struct.
type ComponentBufferReader struct {
	ctx  *runtime.ReadContext
	addr uintptr
}

// NewComponentBufferReader creates a lazy reader for ComponentBuffer at the given address.
func NewComponentBufferReader(ctx *runtime.ReadContext, addr uintptr) *ComponentBufferReader {
	return &ComponentBufferReader{ctx: ctx, addr: addr}
}

// Addr returns the base address of this ComponentBuffer.
func (r *ComponentBufferReader) Addr() uintptr {
	return r.addr
}

// Read materializes the full ComponentBuffer struct eagerly.
func (r *ComponentBufferReader) Read() (*ComponentBuffer, runtime.Errors) {
	return ReadComponentBuffer(r.ctx, r.addr)
}

func (r *ComponentBufferReader) Vtable() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+0); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *ComponentBufferReader) Sentinel() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+4); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *ComponentBufferReader) InitialCapacity() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+8); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *ComponentBufferReader) Unknown0c() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+12); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

// SparseIndex returns a lazy reader for the nested StdVectorHeader (zero I/O).
func (r *ComponentBufferReader) SparseIndex() *StdVectorHeaderReader {
	return NewStdVectorHeaderReader(r.ctx, uintptr(int64(r.addr)+16))
}

// EntityRefs returns a lazy reader for the nested StdVectorHeader (zero I/O).
func (r *ComponentBufferReader) EntityRefs() *StdVectorHeaderReader {
	return NewStdVectorHeaderReader(r.ctx, uintptr(int64(r.addr)+28))
}

// PrevIndex returns a lazy reader for the nested StdVectorHeader (zero I/O).
func (r *ComponentBufferReader) PrevIndex() *StdVectorHeaderReader {
	return NewStdVectorHeaderReader(r.ctx, uintptr(int64(r.addr)+40))
}

// NextIndex returns a lazy reader for the nested StdVectorHeader (zero I/O).
func (r *ComponentBufferReader) NextIndex() *StdVectorHeaderReader {
	return NewStdVectorHeaderReader(r.ctx, uintptr(int64(r.addr)+52))
}

// Components returns a lazy reader for the nested StdVectorHeader (zero I/O).
func (r *ComponentBufferReader) Components() *StdVectorHeaderReader {
	return NewStdVectorHeaderReader(r.ctx, uintptr(int64(r.addr)+64))
}

// HandleMap returns a lazy reader for the nested StdVectorHeader (zero I/O).
func (r *ComponentBufferReader) HandleMap() *StdVectorHeaderReader {
	return NewStdVectorHeaderReader(r.ctx, uintptr(int64(r.addr)+96))
}

// Generations returns a lazy reader for the nested StdVectorHeader (zero I/O).
func (r *ComponentBufferReader) Generations() *StdVectorHeaderReader {
	return NewStdVectorHeaderReader(r.ctx, uintptr(int64(r.addr)+108))
}

// ReverseHandleMap returns a lazy reader for the nested StdVectorHeader (zero I/O).
func (r *ComponentBufferReader) ReverseHandleMap() *StdVectorHeaderReader {
	return NewStdVectorHeaderReader(r.ctx, uintptr(int64(r.addr)+120))
}

func (r *ComponentBufferReader) ActiveCount() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+152); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *ComponentBufferReader) CapacityLimit() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+156); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *ComponentBufferReader) UnknownA0() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+160); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *ComponentBufferReader) PEntityManager() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+164); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *ComponentBufferReader) PEventManager() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+168); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

// NameString returns a lazy reader for the nested MsvcString (zero I/O).
func (r *ComponentBufferReader) NameString() *MsvcStringReader {
	return NewMsvcStringReader(r.ctx, uintptr(int64(r.addr)+172))
}

// WorldStateComponentReader provides lazy, field-level access to WorldStateComponent without reading the entire struct.
type WorldStateComponentReader struct {
	ctx  *runtime.ReadContext
	addr uintptr
}

// NewWorldStateComponentReader creates a lazy reader for WorldStateComponent at the given address.
func NewWorldStateComponentReader(ctx *runtime.ReadContext, addr uintptr) *WorldStateComponentReader {
	return &WorldStateComponentReader{ctx: ctx, addr: addr}
}

// Addr returns the base address of this WorldStateComponent.
func (r *WorldStateComponentReader) Addr() uintptr {
	return r.addr
}

// Read materializes the full WorldStateComponent struct eagerly.
func (r *WorldStateComponentReader) Read() (*WorldStateComponent, runtime.Errors) {
	return ReadWorldStateComponent(r.ctx, r.addr)
}

// Header returns a lazy reader for the nested ComponentHeader (zero I/O).
func (r *WorldStateComponentReader) Header() *ComponentHeaderReader {
	return NewComponentHeaderReader(r.ctx, uintptr(int64(r.addr)+0))
}

func (r *WorldStateComponentReader) BiomeCryptCount() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+264); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *WorldStateComponentReader) GodsAfraid() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+268); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *WorldStateComponentReader) GodsImpressed() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+272); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *WorldStateComponentReader) GodsAfraidDamage() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+276); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *WorldStateComponentReader) GodsEnraged() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+280); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

// GameGlobalsReader provides lazy, field-level access to GameGlobals without reading the entire struct.
type GameGlobalsReader struct {
	ctx  *runtime.ReadContext
	addr uintptr
}

// NewGameGlobalsReader creates a lazy reader for GameGlobals at the given address.
func NewGameGlobalsReader(ctx *runtime.ReadContext, addr uintptr) *GameGlobalsReader {
	return &GameGlobalsReader{ctx: ctx, addr: addr}
}

// Addr returns the base address of this GameGlobals.
func (r *GameGlobalsReader) Addr() uintptr {
	return r.addr
}

// Read materializes the full GameGlobals struct eagerly.
func (r *GameGlobalsReader) Read() (*GameGlobals, runtime.Errors) {
	return ReadGameGlobals(r.ctx, r.addr)
}

func (r *GameGlobalsReader) FrameCount() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+0); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *GameGlobalsReader) PhysicsStepCount() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+4); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *GameGlobalsReader) GameTime() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+8); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *GameGlobalsReader) PWorldManager() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+12); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *GameGlobalsReader) PChunkSystem() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+16); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *GameGlobalsReader) PCellGrid() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+20); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *GameGlobalsReader) PCellFactory() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+24); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *GameGlobalsReader) Unknown1c() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+28); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *GameGlobalsReader) PPhysicsWorld() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+32); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *GameGlobalsReader) PAudioManager() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+36); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *GameGlobalsReader) ViewportLeft() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+384); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *GameGlobalsReader) ViewportTop() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+388); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *GameGlobalsReader) ViewportRight() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+392); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *GameGlobalsReader) ViewportBottom() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+396); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

// EntityReader provides lazy, field-level access to Entity without reading the entire struct.
type EntityReader struct {
	ctx  *runtime.ReadContext
	addr uintptr
}

// NewEntityReader creates a lazy reader for Entity at the given address.
func NewEntityReader(ctx *runtime.ReadContext, addr uintptr) *EntityReader {
	return &EntityReader{ctx: ctx, addr: addr}
}

// Addr returns the base address of this Entity.
func (r *EntityReader) Addr() uintptr {
	return r.addr
}

// Read materializes the full Entity struct eagerly.
func (r *EntityReader) Read() (*Entity, runtime.Errors) {
	return ReadEntity(r.ctx, r.addr)
}

func (r *EntityReader) EntityId() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+0); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *EntityReader) SlotIndex() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+4); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *EntityReader) Unknown08() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+8); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *EntityReader) PendingKill() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+12); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *EntityReader) Flags10() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+16); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

// Name returns a lazy reader for the nested MsvcString (zero I/O).
func (r *EntityReader) Name() *MsvcStringReader {
	return NewMsvcStringReader(r.ctx, uintptr(int64(r.addr)+20))
}

func (r *EntityReader) Unknown2c() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+44); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *EntityReader) TagBitset() ([64]uint8, error) {
	var result [64]uint8
	if _, err := r.ctx.ReadAt(result[:], int64(r.addr)+48); err != nil {
		return result, err
	}
	return result, nil
}

func (r *EntityReader) PosX() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+112); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *EntityReader) PosY() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+116); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *EntityReader) RotCos() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+120); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *EntityReader) RotSin() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+124); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *EntityReader) RotNegSin() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+128); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *EntityReader) RotCos2() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+132); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *EntityReader) ScaleX() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+136); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *EntityReader) ScaleY() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+140); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *EntityReader) ChildrenPtr() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+144); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *EntityReader) ParentEntityPtr() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+148); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

// EntityManagerReader provides lazy, field-level access to EntityManager without reading the entire struct.
type EntityManagerReader struct {
	ctx  *runtime.ReadContext
	addr uintptr
}

// NewEntityManagerReader creates a lazy reader for EntityManager at the given address.
func NewEntityManagerReader(ctx *runtime.ReadContext, addr uintptr) *EntityManagerReader {
	return &EntityManagerReader{ctx: ctx, addr: addr}
}

// Addr returns the base address of this EntityManager.
func (r *EntityManagerReader) Addr() uintptr {
	return r.addr
}

// Read materializes the full EntityManager struct eagerly.
func (r *EntityManagerReader) Read() (*EntityManager, runtime.Errors) {
	return ReadEntityManager(r.ctx, r.addr)
}

func (r *EntityManagerReader) Vtable() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+0); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *EntityManagerReader) NextEntityId() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+4); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

// FreeSlotStack returns a lazy reader for the nested StdVectorHeader (zero I/O).
func (r *EntityManagerReader) FreeSlotStack() *StdVectorHeaderReader {
	return NewStdVectorHeaderReader(r.ctx, uintptr(int64(r.addr)+8))
}

// EntityArray returns a lazy reader for the nested StdVectorHeader (zero I/O).
func (r *EntityManagerReader) EntityArray() *StdVectorHeaderReader {
	return NewStdVectorHeaderReader(r.ctx, uintptr(int64(r.addr)+20))
}

// TagGroups returns a lazy reader for the nested StdVectorHeader (zero I/O).
func (r *EntityManagerReader) TagGroups() *StdVectorHeaderReader {
	return NewStdVectorHeaderReader(r.ctx, uintptr(int64(r.addr)+32))
}

// ComponentBuffers eagerly reads the nested U32Vector (no lazy reader available for this type).
func (r *EntityManagerReader) ComponentBuffers() (*U32Vector, runtime.Errors) {
	return ReadU32Vector(r.ctx, uintptr(int64(r.addr)+44))
}

func (r *EntityManagerReader) PEventManager() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+56); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

// WalletComponentReader provides lazy, field-level access to WalletComponent without reading the entire struct.
type WalletComponentReader struct {
	ctx  *runtime.ReadContext
	addr uintptr
}

// NewWalletComponentReader creates a lazy reader for WalletComponent at the given address.
func NewWalletComponentReader(ctx *runtime.ReadContext, addr uintptr) *WalletComponentReader {
	return &WalletComponentReader{ctx: ctx, addr: addr}
}

// Addr returns the base address of this WalletComponent.
func (r *WalletComponentReader) Addr() uintptr {
	return r.addr
}

// Read materializes the full WalletComponent struct eagerly.
func (r *WalletComponentReader) Read() (*WalletComponent, runtime.Errors) {
	return ReadWalletComponent(r.ctx, r.addr)
}

// Header returns a lazy reader for the nested ComponentHeader (zero I/O).
func (r *WalletComponentReader) Header() *ComponentHeaderReader {
	return NewComponentHeaderReader(r.ctx, uintptr(int64(r.addr)+0))
}

func (r *WalletComponentReader) Money() (int64, error) {
	var buf [8]byte
	if _, err := r.ctx.ReadAt(buf[:8], int64(r.addr)+72); err != nil {
		return 0, err
	}
	return int64(binary.LittleEndian.Uint64(buf[:8])), nil
}

func (r *WalletComponentReader) MoneySpent() (int64, error) {
	var buf [8]byte
	if _, err := r.ctx.ReadAt(buf[:8], int64(r.addr)+80); err != nil {
		return 0, err
	}
	return int64(binary.LittleEndian.Uint64(buf[:8])), nil
}

func (r *WalletComponentReader) MoneyPrevFrame() (int64, error) {
	var buf [8]byte
	if _, err := r.ctx.ReadAt(buf[:8], int64(r.addr)+88); err != nil {
		return 0, err
	}
	return int64(binary.LittleEndian.Uint64(buf[:8])), nil
}

func (r *WalletComponentReader) HasReachedInf() (bool, error) {
	var buf [1]byte
	if _, err := r.ctx.ReadAt(buf[:1], int64(r.addr)+96); err != nil {
		return false, err
	}
	return buf[0] != 0, nil
}

// MaterialInventoryComponentReader provides lazy, field-level access to MaterialInventoryComponent without reading the entire struct.
type MaterialInventoryComponentReader struct {
	ctx  *runtime.ReadContext
	addr uintptr
}

// NewMaterialInventoryComponentReader creates a lazy reader for MaterialInventoryComponent at the given address.
func NewMaterialInventoryComponentReader(ctx *runtime.ReadContext, addr uintptr) *MaterialInventoryComponentReader {
	return &MaterialInventoryComponentReader{ctx: ctx, addr: addr}
}

// Addr returns the base address of this MaterialInventoryComponent.
func (r *MaterialInventoryComponentReader) Addr() uintptr {
	return r.addr
}

// Read materializes the full MaterialInventoryComponent struct eagerly.
func (r *MaterialInventoryComponentReader) Read() (*MaterialInventoryComponent, runtime.Errors) {
	return ReadMaterialInventoryComponent(r.ctx, r.addr)
}

// Header returns a lazy reader for the nested ComponentHeader (zero I/O).
func (r *MaterialInventoryComponentReader) Header() *ComponentHeaderReader {
	return NewComponentHeaderReader(r.ctx, uintptr(int64(r.addr)+0))
}

// CountPerMaterialType eagerly reads the nested F64Vector (no lazy reader available for this type).
func (r *MaterialInventoryComponentReader) CountPerMaterialType() (*F64Vector, runtime.Errors) {
	return ReadF64Vector(r.ctx, uintptr(int64(r.addr)+128))
}

// Ensure imports are used.
var (
	_ = binary.LittleEndian
	_ = math.Float32frombits
)
