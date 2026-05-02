package noita

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/vitaminmoo/memtools/hexpat/runtime"
	"math"
)

type TypeID int32

const (
	TypeIDAbilityComponent                    TypeID = 3
	TypeIDAnimalAIComponent                   TypeID = 6
	TypeIDAudioComponent                      TypeID = 10
	TypeIDAudioListenerComponent              TypeID = 11
	TypeIDAudioLoopComponent                  TypeID = 12
	TypeIDBiomeTrackerComponent               TypeID = 13
	TypeIDCameraBoundComponent                TypeID = 18
	TypeIDCharacterCollisionComponent         TypeID = 21
	TypeIDCharacterDataComponent              TypeID = 22
	TypeIDCharacterPlatformingComponent       TypeID = 23
	TypeIDCollisionTriggerComponent           TypeID = 25
	TypeIDControlsComponent                   TypeID = 28
	TypeIDDamageModelComponent                TypeID = 31
	TypeIDDrugEffectComponent                 TypeID = 38
	TypeIDElectricityReceiverComponent        TypeID = 42
	TypeIDExplodeOnDamageComponent            TypeID = 46
	TypeIDGameEffectComponent                 TypeID = 53
	TypeIDGameLogComponent                    TypeID = 54
	TypeIDGameStatsComponent                  TypeID = 55
	TypeIDGenomeDataComponent                 TypeID = 57
	TypeIDGunComponent                        TypeID = 60
	TypeIDHitboxComponent                     TypeID = 63
	TypeIDHotspotComponent                    TypeID = 65
	TypeIDIngestionComponent                  TypeID = 70
	TypeIDInheritTransformComponent           TypeID = 71
	TypeIDInteractableComponent               TypeID = 72
	TypeIDInventory2Component                 TypeID = 73
	TypeIDInventoryGuiComponent               TypeID = 75
	TypeIDItemActionComponent                 TypeID = 77
	TypeIDItemAlchemyComponent                TypeID = 78
	TypeIDItemChestComponent                  TypeID = 79
	TypeIDItemComponent                       TypeID = 80
	TypeIDItemCostComponent                   TypeID = 81
	TypeIDItemPickUpperComponent              TypeID = 82
	TypeIDKickComponent                       TypeID = 85
	TypeIDLifetimeComponent                   TypeID = 88
	TypeIDLightComponent                      TypeID = 89
	TypeIDLiquidDisplacerComponent            TypeID = 92
	TypeIDLuaComponent                        TypeID = 96
	TypeIDManaReloaderComponent               TypeID = 99
	TypeIDMaterialAreaCheckerComponent        TypeID = 100
	TypeIDMaterialInventoryComponent          TypeID = 101
	TypeIDMaterialSuckerComponent             TypeID = 103
	TypeIDMusicEnergyAffectorComponent        TypeID = 105
	TypeIDParticleEmitterComponent            TypeID = 109
	TypeIDPathFindingComponent                TypeID = 110
	TypeIDPathFindingGridMarkerComponent      TypeID = 111
	TypeIDPhysicsBody2Component               TypeID = 113
	TypeIDPhysicsBodyCollisionDamageComponent TypeID = 114
	TypeIDPhysicsBodyComponent                TypeID = 115
	TypeIDPhysicsImageShapeComponent          TypeID = 116
	TypeIDPhysicsJointComponent               TypeID = 119
	TypeIDPhysicsPickUpComponent              TypeID = 121
	TypeIDPhysicsThrowableComponent           TypeID = 124
	TypeIDPixelSpriteComponent                TypeID = 126
	TypeIDPlatformShooterPlayerComponent      TypeID = 127
	TypeIDPlayerCollisionComponent            TypeID = 128
	TypeIDPlayerStatsComponent                TypeID = 129
	TypeIDPositionSeedComponent               TypeID = 130
	TypeIDPotionComponent                     TypeID = 131
	TypeIDProjectileComponent                 TypeID = 133
	TypeIDSimplePhysicsComponent              TypeID = 138
	TypeIDSpriteAnimatorComponent             TypeID = 140
	TypeIDSpriteComponent                     TypeID = 141
	TypeIDSpriteOffsetAnimatorComponent       TypeID = 142
	TypeIDSpriteParticleEmitterComponent      TypeID = 143
	TypeIDSpriteStainsComponent               TypeID = 144
	TypeIDStatusEffectDataComponent           TypeID = 145
	TypeIDStreamingKeepAliveComponent         TypeID = 146
	TypeIDTorchComponent                      TypeID = 151
	TypeIDUIInfoComponent                     TypeID = 153
	TypeIDVariableStorageComponent            TypeID = 154
	TypeIDVelocityComponent                   TypeID = 155
	TypeIDVerletPhysicsComponent              TypeID = 156
	TypeIDVerletWorldJointComponent           TypeID = 158
	TypeIDWalletComponent                     TypeID = 159
	TypeIDWorldStateComponent                 TypeID = 161
)

func (e TypeID) String() string {
	switch e {
	case TypeIDAbilityComponent:
		return fmt.Sprintf("AbilityComponent (%d)", int32(e))
	case TypeIDAnimalAIComponent:
		return fmt.Sprintf("AnimalAIComponent (%d)", int32(e))
	case TypeIDAudioComponent:
		return fmt.Sprintf("AudioComponent (%d)", int32(e))
	case TypeIDAudioListenerComponent:
		return fmt.Sprintf("AudioListenerComponent (%d)", int32(e))
	case TypeIDAudioLoopComponent:
		return fmt.Sprintf("AudioLoopComponent (%d)", int32(e))
	case TypeIDBiomeTrackerComponent:
		return fmt.Sprintf("BiomeTrackerComponent (%d)", int32(e))
	case TypeIDCameraBoundComponent:
		return fmt.Sprintf("CameraBoundComponent (%d)", int32(e))
	case TypeIDCharacterCollisionComponent:
		return fmt.Sprintf("CharacterCollisionComponent (%d)", int32(e))
	case TypeIDCharacterDataComponent:
		return fmt.Sprintf("CharacterDataComponent (%d)", int32(e))
	case TypeIDCharacterPlatformingComponent:
		return fmt.Sprintf("CharacterPlatformingComponent (%d)", int32(e))
	case TypeIDCollisionTriggerComponent:
		return fmt.Sprintf("CollisionTriggerComponent (%d)", int32(e))
	case TypeIDControlsComponent:
		return fmt.Sprintf("ControlsComponent (%d)", int32(e))
	case TypeIDDamageModelComponent:
		return fmt.Sprintf("DamageModelComponent (%d)", int32(e))
	case TypeIDDrugEffectComponent:
		return fmt.Sprintf("DrugEffectComponent (%d)", int32(e))
	case TypeIDElectricityReceiverComponent:
		return fmt.Sprintf("ElectricityReceiverComponent (%d)", int32(e))
	case TypeIDExplodeOnDamageComponent:
		return fmt.Sprintf("ExplodeOnDamageComponent (%d)", int32(e))
	case TypeIDGameEffectComponent:
		return fmt.Sprintf("GameEffectComponent (%d)", int32(e))
	case TypeIDGameLogComponent:
		return fmt.Sprintf("GameLogComponent (%d)", int32(e))
	case TypeIDGameStatsComponent:
		return fmt.Sprintf("GameStatsComponent (%d)", int32(e))
	case TypeIDGenomeDataComponent:
		return fmt.Sprintf("GenomeDataComponent (%d)", int32(e))
	case TypeIDGunComponent:
		return fmt.Sprintf("GunComponent (%d)", int32(e))
	case TypeIDHitboxComponent:
		return fmt.Sprintf("HitboxComponent (%d)", int32(e))
	case TypeIDHotspotComponent:
		return fmt.Sprintf("HotspotComponent (%d)", int32(e))
	case TypeIDIngestionComponent:
		return fmt.Sprintf("IngestionComponent (%d)", int32(e))
	case TypeIDInheritTransformComponent:
		return fmt.Sprintf("InheritTransformComponent (%d)", int32(e))
	case TypeIDInteractableComponent:
		return fmt.Sprintf("InteractableComponent (%d)", int32(e))
	case TypeIDInventory2Component:
		return fmt.Sprintf("Inventory2Component (%d)", int32(e))
	case TypeIDInventoryGuiComponent:
		return fmt.Sprintf("InventoryGuiComponent (%d)", int32(e))
	case TypeIDItemActionComponent:
		return fmt.Sprintf("ItemActionComponent (%d)", int32(e))
	case TypeIDItemAlchemyComponent:
		return fmt.Sprintf("ItemAlchemyComponent (%d)", int32(e))
	case TypeIDItemChestComponent:
		return fmt.Sprintf("ItemChestComponent (%d)", int32(e))
	case TypeIDItemComponent:
		return fmt.Sprintf("ItemComponent (%d)", int32(e))
	case TypeIDItemCostComponent:
		return fmt.Sprintf("ItemCostComponent (%d)", int32(e))
	case TypeIDItemPickUpperComponent:
		return fmt.Sprintf("ItemPickUpperComponent (%d)", int32(e))
	case TypeIDKickComponent:
		return fmt.Sprintf("KickComponent (%d)", int32(e))
	case TypeIDLifetimeComponent:
		return fmt.Sprintf("LifetimeComponent (%d)", int32(e))
	case TypeIDLightComponent:
		return fmt.Sprintf("LightComponent (%d)", int32(e))
	case TypeIDLiquidDisplacerComponent:
		return fmt.Sprintf("LiquidDisplacerComponent (%d)", int32(e))
	case TypeIDLuaComponent:
		return fmt.Sprintf("LuaComponent (%d)", int32(e))
	case TypeIDManaReloaderComponent:
		return fmt.Sprintf("ManaReloaderComponent (%d)", int32(e))
	case TypeIDMaterialAreaCheckerComponent:
		return fmt.Sprintf("MaterialAreaCheckerComponent (%d)", int32(e))
	case TypeIDMaterialInventoryComponent:
		return fmt.Sprintf("MaterialInventoryComponent (%d)", int32(e))
	case TypeIDMaterialSuckerComponent:
		return fmt.Sprintf("MaterialSuckerComponent (%d)", int32(e))
	case TypeIDMusicEnergyAffectorComponent:
		return fmt.Sprintf("MusicEnergyAffectorComponent (%d)", int32(e))
	case TypeIDParticleEmitterComponent:
		return fmt.Sprintf("ParticleEmitterComponent (%d)", int32(e))
	case TypeIDPathFindingComponent:
		return fmt.Sprintf("PathFindingComponent (%d)", int32(e))
	case TypeIDPathFindingGridMarkerComponent:
		return fmt.Sprintf("PathFindingGridMarkerComponent (%d)", int32(e))
	case TypeIDPhysicsBody2Component:
		return fmt.Sprintf("PhysicsBody2Component (%d)", int32(e))
	case TypeIDPhysicsBodyCollisionDamageComponent:
		return fmt.Sprintf("PhysicsBodyCollisionDamageComponent (%d)", int32(e))
	case TypeIDPhysicsBodyComponent:
		return fmt.Sprintf("PhysicsBodyComponent (%d)", int32(e))
	case TypeIDPhysicsImageShapeComponent:
		return fmt.Sprintf("PhysicsImageShapeComponent (%d)", int32(e))
	case TypeIDPhysicsJointComponent:
		return fmt.Sprintf("PhysicsJointComponent (%d)", int32(e))
	case TypeIDPhysicsPickUpComponent:
		return fmt.Sprintf("PhysicsPickUpComponent (%d)", int32(e))
	case TypeIDPhysicsThrowableComponent:
		return fmt.Sprintf("PhysicsThrowableComponent (%d)", int32(e))
	case TypeIDPixelSpriteComponent:
		return fmt.Sprintf("PixelSpriteComponent (%d)", int32(e))
	case TypeIDPlatformShooterPlayerComponent:
		return fmt.Sprintf("PlatformShooterPlayerComponent (%d)", int32(e))
	case TypeIDPlayerCollisionComponent:
		return fmt.Sprintf("PlayerCollisionComponent (%d)", int32(e))
	case TypeIDPlayerStatsComponent:
		return fmt.Sprintf("PlayerStatsComponent (%d)", int32(e))
	case TypeIDPositionSeedComponent:
		return fmt.Sprintf("PositionSeedComponent (%d)", int32(e))
	case TypeIDPotionComponent:
		return fmt.Sprintf("PotionComponent (%d)", int32(e))
	case TypeIDProjectileComponent:
		return fmt.Sprintf("ProjectileComponent (%d)", int32(e))
	case TypeIDSimplePhysicsComponent:
		return fmt.Sprintf("SimplePhysicsComponent (%d)", int32(e))
	case TypeIDSpriteAnimatorComponent:
		return fmt.Sprintf("SpriteAnimatorComponent (%d)", int32(e))
	case TypeIDSpriteComponent:
		return fmt.Sprintf("SpriteComponent (%d)", int32(e))
	case TypeIDSpriteOffsetAnimatorComponent:
		return fmt.Sprintf("SpriteOffsetAnimatorComponent (%d)", int32(e))
	case TypeIDSpriteParticleEmitterComponent:
		return fmt.Sprintf("SpriteParticleEmitterComponent (%d)", int32(e))
	case TypeIDSpriteStainsComponent:
		return fmt.Sprintf("SpriteStainsComponent (%d)", int32(e))
	case TypeIDStatusEffectDataComponent:
		return fmt.Sprintf("StatusEffectDataComponent (%d)", int32(e))
	case TypeIDStreamingKeepAliveComponent:
		return fmt.Sprintf("StreamingKeepAliveComponent (%d)", int32(e))
	case TypeIDTorchComponent:
		return fmt.Sprintf("TorchComponent (%d)", int32(e))
	case TypeIDUIInfoComponent:
		return fmt.Sprintf("UIInfoComponent (%d)", int32(e))
	case TypeIDVariableStorageComponent:
		return fmt.Sprintf("VariableStorageComponent (%d)", int32(e))
	case TypeIDVelocityComponent:
		return fmt.Sprintf("VelocityComponent (%d)", int32(e))
	case TypeIDVerletPhysicsComponent:
		return fmt.Sprintf("VerletPhysicsComponent (%d)", int32(e))
	case TypeIDVerletWorldJointComponent:
		return fmt.Sprintf("VerletWorldJointComponent (%d)", int32(e))
	case TypeIDWalletComponent:
		return fmt.Sprintf("WalletComponent (%d)", int32(e))
	case TypeIDWorldStateComponent:
		return fmt.Sprintf("WorldStateComponent (%d)", int32(e))
	default:
		return fmt.Sprintf("unknown (%d)", int32(e))
	}
}

func (e TypeID) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())
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

type MsvcString struct {
	Data     [16]byte
	Length   uint32
	Capacity uint32
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

type SpriteComponent struct {
	Header    ComponentHeader
	ImageFile MsvcString
}

type StdVectorHeader struct {
	BeginPtr    uint32
	EndPtr      uint32
	CapacityPtr uint32
}

type WorldStateComponent struct {
	Header           ComponentHeader
	ChangedMaterials StdVectorHeader
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
	Children    []uint32
}

type LuaComponent struct {
	Header           ComponentHeader
	ScriptSourceFile MsvcString
}

type BiomeChunk struct {
	Vtable         uint32
	Unknown04      uint32
	BiomeName      MsvcString
	WobbleEligible uint8
	WavyEdge       uint8
	ForceOriginal  uint8
	UnknownC7      uint8
	BiomeDataPtr   uint32
	XmlPath        MsvcString
}

type BiomeGrid struct {
	ScenesBegin          uint32
	ScenesEnd            uint32
	ScenesCapacityEnd    uint32
	ScenesAltBegin       uint32
	ScenesAltEnd         uint32
	ScenesAltCapacityEnd uint32
	XShift               float64
	YShift               float64
	Width                int32
	Height               int32
	TotalCount           int32
	Unknown5C            int32
	Unknown60            uint32
	Unknown64            uint32
	ChunksPtr            uint32
	ChunksCount          int32
}

type CellTexture struct {
	Width        int32
	Height       int32
	Unknown08    uint32
	PixelDataPtr uint32
}

type S32Vector struct {
	BeginPtr    uint32
	EndPtr      uint32
	CapacityPtr uint32
	Elements    []int32
}

type U32Vector struct {
	BeginPtr    uint32
	EndPtr      uint32
	CapacityPtr uint32
	Elements    []uint32
}

type LightComponent struct {
	Header      ComponentHeader
	InternalPtr uint32
	Radius      float32
	R           uint32
	G           uint32
	B           uint32
	OffsetX     float32
	OffsetY     float32
}

type GameEffectComponent struct {
	Header ComponentHeader
	Effect int32
	Frames int32
}

type CellFactory struct {
	CellDataArrayPtr uint32
	MaterialCount    int32
	Material0Color   uint32
}

type Chunk struct {
	CellSlotsPtr uint32
}

type HitboxComponent struct {
	Header           ComponentHeader
	IsPlayer         bool
	IsEnemy          bool
	IsItem           bool
	AabbMinX         float32
	AabbMaxX         float32
	AabbMinY         float32
	AabbMaxY         float32
	DamageMultiplier float32
	OffsetX          float32
	OffsetY          float32
}

type ItemComponent struct {
	Header                 ComponentHeader
	ItemName               MsvcString
	IsStackable            bool
	IsConsumable           bool
	StatsCountAsItemPickUp bool
	AutoPickup             bool
	Unknown64              uint32
	UsesRemaining          int32
	IsIdentified           bool
	IsFrozen               bool
}

type WorldManagerViewRect struct {
	ViewX           float32
	ViewY           float32
	ViewWidth       float32
	ViewHeight      float32
	PBackgroundGrid uint32
}

type CellData struct {
	Name          MsvcString
	FallbackColor uint32
	TexturePtr    uint32
}

type CellGrid struct {
	Vtable        uint32
	Unknown04     uint32
	ChunkTablePtr uint32
}

type CellMaterialInfo struct {
	MaterialId int32
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

type CharacterDataComponent struct {
	Header     ComponentHeader
	Gravity    float32
	FlyTimeMax float32
	IsOnGround bool
	VelocityX  float32
	VelocityY  float32
}

type VelocityComponent struct {
	Header           ComponentHeader
	GravityX         float32
	GravityY         float32
	Mass             float32
	AirFriction      float32
	TerminalVelocity float32
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

type MaterialInventoryComponent struct {
	Header               ComponentHeader
	CountPerMaterialType F64Vector
}

type ChunkSystem struct {
	Vtable   uint32
	CellGrid CellGrid
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

type WalletComponent struct {
	Header         ComponentHeader
	Money          int64
	MoneySpent     int64
	MoneyPrevFrame int64
	HasReachedInf  bool
}

type CollisionTriggerComponent struct {
	Header      ComponentHeader
	Width       float32
	Height      float32
	Radius      float32
	RequiredTag MsvcString
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

type PixelSceneEntry struct {
	ChunkSystemBackRef   uint32
	X                    int32
	Y                    int32
	MaterialsFilename    MsvcString
	ColorsFilename       MsvcString
	BackgroundFilename   MsvcString
	FlagSkipBiomeChecks  uint8
	FlagSkipEdgeTextures uint8
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

type DamageModelComponent struct {
	Header                       ComponentHeader
	Hp                           float64
	MaxHp                        float64
	MaxHpCap                     float64
	MaxHpOld                     float64
	DamageMultipliersVtable      uint32
	DamageMultipliersMelee       float32
	DamageMultipliersProjectile  float32
	DamageMultipliersExplosion   float32
	DamageMultipliersElectricity float32
	DamageMultipliersFire        float32
	DamageMultipliersDrill       float32
	DamageMultipliersSlice       float32
	DamageMultipliersIce         float32
	DamageMultipliersHealing     float32
	DamageMultipliersPhysicsHit  float32
	DamageMultipliersRadioactive float32
	DamageMultipliersPoison      float32
	DamageMultipliersOvereating  float32
	DamageMultipliersCurse       float32
	DamageMultipliersHoly        float32
	CriticalDamageResistance     float32
	InvincibilityFrames          int32
}

func ReadComponentHeader(ctx *runtime.ReadContext, addr uintptr) (*ComponentHeader, runtime.Errors) {
	var errs runtime.Errors
	result := &ComponentHeader{}
	var buf [72]byte

	if _, err := ctx.ReadAt(buf[:], int64(addr)); err != nil {
		errs.Add("ComponentHeader", uintptr(addr), err)
		return result, errs
	}

	result.Vtable = binary.LittleEndian.Uint32(buf[0:])
	result.BufferIndex = int32(binary.LittleEndian.Uint32(buf[4:]))
	result.PTypeName = binary.LittleEndian.Uint32(buf[8:])
	result.TypeId = int32(binary.LittleEndian.Uint32(buf[12:]))
	result.Unknown10 = binary.LittleEndian.Uint32(buf[16:])
	result.Active = buf[20] != 0
	copy(result.ComponentTags[:], buf[24:56])
	result.Unknown38 = binary.LittleEndian.Uint32(buf[56:])
	result.Unknown3c = binary.LittleEndian.Uint32(buf[60:])
	result.Unknown40 = binary.LittleEndian.Uint32(buf[64:])
	result.Unknown44 = binary.LittleEndian.Uint32(buf[68:])
	return result, errs
}

func ReadMsvcString(ctx *runtime.ReadContext, addr uintptr) (*MsvcString, runtime.Errors) {
	var errs runtime.Errors
	result := &MsvcString{}
	var buf [24]byte

	if _, err := ctx.ReadAt(buf[:], int64(addr)); err != nil {
		errs.Add("MsvcString", uintptr(addr), err)
		return result, errs
	}

	copy(result.Data[:], buf[0:16])
	result.Length = binary.LittleEndian.Uint32(buf[16:])
	result.Capacity = binary.LittleEndian.Uint32(buf[20:])
	return result, errs
}

func ReadConfigGun(ctx *runtime.ReadContext, addr uintptr) (*ConfigGun, runtime.Errors) {
	var errs runtime.Errors
	result := &ConfigGun{}
	var buf [20]byte

	if _, err := ctx.ReadAt(buf[:], int64(addr)); err != nil {
		errs.Add("ConfigGun", uintptr(addr), err)
		return result, errs
	}

	result.Vtable = binary.LittleEndian.Uint32(buf[0:])
	result.ActionsPerRound = int32(binary.LittleEndian.Uint32(buf[4:]))
	result.ShuffleDeckWhenEmpty = buf[8] != 0
	result.ReloadTime = int32(binary.LittleEndian.Uint32(buf[12:]))
	result.DeckCapacity = int32(binary.LittleEndian.Uint32(buf[16:]))
	return result, errs
}

func ReadAbilityComponent(ctx *runtime.ReadContext, addr uintptr) (*AbilityComponent, runtime.Errors) {
	var errs runtime.Errors
	result := &AbilityComponent{}
	var buf [956]byte

	if _, err := ctx.ReadAt(buf[:], int64(addr)); err != nil {
		errs.Add("AbilityComponent", uintptr(addr), err)
		return result, errs
	}

	// Field: Header at int64(addr)+0
	{
		child, childErrs := ReadComponentHeader(ctx, uintptr(int64(addr)+0))
		if child != nil {
			result.Header = *child
		}
		errs = append(errs, childErrs...)
	}

	result.CooldownFrames = int32(binary.LittleEndian.Uint32(buf[72:]))
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

	result.EntityCount = int32(binary.LittleEndian.Uint32(buf[124:]))
	result.NeverReload = buf[128] != 0
	result.ReloadTimeFrames = int32(binary.LittleEndian.Uint32(buf[132:]))
	result.Mana = math.Float32frombits(binary.LittleEndian.Uint32(buf[136:]))
	result.ManaMax = math.Float32frombits(binary.LittleEndian.Uint32(buf[140:]))
	result.ManaChargeSpeed = math.Float32frombits(binary.LittleEndian.Uint32(buf[144:]))
	result.RotateInHand = buf[148] != 0
	result.RotateInHandAmount = math.Float32frombits(binary.LittleEndian.Uint32(buf[152:]))
	result.RotateHandAmount = math.Float32frombits(binary.LittleEndian.Uint32(buf[156:]))
	result.FastProjectile = buf[160] != 0
	result.SwimPropelAmount = math.Float32frombits(binary.LittleEndian.Uint32(buf[164:]))
	result.MaxChargedActions = int32(binary.LittleEndian.Uint32(buf[168:]))
	result.ChargeWaitFrames = int32(binary.LittleEndian.Uint32(buf[172:]))
	result.ItemRecoilRecoverySpeed = math.Float32frombits(binary.LittleEndian.Uint32(buf[176:]))
	result.ItemRecoilMax = math.Float32frombits(binary.LittleEndian.Uint32(buf[180:]))
	result.ItemRecoilOffsetCoeff = math.Float32frombits(binary.LittleEndian.Uint32(buf[184:]))
	result.ItemRecoilRotationCoeff = math.Float32frombits(binary.LittleEndian.Uint32(buf[188:]))
	// Field: BaseItemFile at int64(addr)+192
	{
		child, childErrs := ReadMsvcString(ctx, uintptr(int64(addr)+192))
		if child != nil {
			result.BaseItemFile = *child
		}
		errs = append(errs, childErrs...)
	}

	result.UseEntityFileAsProjectileInfoProxy = buf[216] != 0
	result.ClickToUse = buf[217] != 0
	result.StatTimesPlayerHasShot = int32(binary.LittleEndian.Uint32(buf[220:]))
	result.StatTimesPlayerHasEdited = int32(binary.LittleEndian.Uint32(buf[224:]))
	result.ShootingReducesAmountInInventory = buf[228] != 0
	result.ThrowAsItem = buf[229] != 0
	result.SimulateThrowAsItem = buf[230] != 0
	result.MaxAmountInInventory = int32(binary.LittleEndian.Uint32(buf[232:]))
	result.AmountInInventory = int32(binary.LittleEndian.Uint32(buf[236:]))
	result.DropAsItemOnDeath = buf[240] != 0
	// Field: UiName at int64(addr)+244
	{
		child, childErrs := ReadMsvcString(ctx, uintptr(int64(addr)+244))
		if child != nil {
			result.UiName = *child
		}
		errs = append(errs, childErrs...)
	}

	result.UseGunScript = buf[268] != 0
	result.IsPetrisGun = buf[269] != 0
	// Field: GunConfig at int64(addr)+272
	{
		child, childErrs := ReadConfigGun(ctx, uintptr(int64(addr)+272))
		if child != nil {
			result.GunConfig = *child
		}
		errs = append(errs, childErrs...)
	}

	result.GunLevel = int32(binary.LittleEndian.Uint32(buf[864:]))
	// Field: AddTheseChildActions at int64(addr)+868
	{
		child, childErrs := ReadMsvcString(ctx, uintptr(int64(addr)+868))
		if child != nil {
			result.AddTheseChildActions = *child
		}
		errs = append(errs, childErrs...)
	}

	result.CurrentSlotDurability = int32(binary.LittleEndian.Uint32(buf[892:]))
	// Field: SlotConsumptionFunction at int64(addr)+896
	{
		child, childErrs := ReadMsvcString(ctx, uintptr(int64(addr)+896))
		if child != nil {
			result.SlotConsumptionFunction = *child
		}
		errs = append(errs, childErrs...)
	}

	result.NextFrameUsable = int32(binary.LittleEndian.Uint32(buf[920:]))
	result.CastDelayStartFrame = int32(binary.LittleEndian.Uint32(buf[924:]))
	result.AmmoLeft = int32(binary.LittleEndian.Uint32(buf[928:]))
	result.ReloadFramesLeft = int32(binary.LittleEndian.Uint32(buf[932:]))
	result.ReloadNextFrameUsable = int32(binary.LittleEndian.Uint32(buf[936:]))
	result.ChargeCount = int32(binary.LittleEndian.Uint32(buf[940:]))
	result.NextChargeFrame = int32(binary.LittleEndian.Uint32(buf[944:]))
	result.ItemRecoil = math.Float32frombits(binary.LittleEndian.Uint32(buf[948:]))
	result.IsInitialized = buf[952] != 0
	return result, errs
}

func ReadSpriteComponent(ctx *runtime.ReadContext, addr uintptr) (*SpriteComponent, runtime.Errors) {
	var errs runtime.Errors
	result := &SpriteComponent{}
	var buf [96]byte

	if _, err := ctx.ReadAt(buf[:], int64(addr)); err != nil {
		errs.Add("SpriteComponent", uintptr(addr), err)
		return result, errs
	}

	// Field: Header at int64(addr)+0
	{
		child, childErrs := ReadComponentHeader(ctx, uintptr(int64(addr)+0))
		if child != nil {
			result.Header = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: ImageFile at int64(addr)+72
	{
		child, childErrs := ReadMsvcString(ctx, uintptr(int64(addr)+72))
		if child != nil {
			result.ImageFile = *child
		}
		errs = append(errs, childErrs...)
	}

	return result, errs
}

func ReadStdVectorHeader(ctx *runtime.ReadContext, addr uintptr) (*StdVectorHeader, runtime.Errors) {
	var errs runtime.Errors
	result := &StdVectorHeader{}
	var buf [12]byte

	if _, err := ctx.ReadAt(buf[:], int64(addr)); err != nil {
		errs.Add("StdVectorHeader", uintptr(addr), err)
		return result, errs
	}

	result.BeginPtr = binary.LittleEndian.Uint32(buf[0:])
	result.EndPtr = binary.LittleEndian.Uint32(buf[4:])
	result.CapacityPtr = binary.LittleEndian.Uint32(buf[8:])
	return result, errs
}

func ReadWorldStateComponent(ctx *runtime.ReadContext, addr uintptr) (*WorldStateComponent, runtime.Errors) {
	var errs runtime.Errors
	result := &WorldStateComponent{}
	var buf [456]byte

	if _, err := ctx.ReadAt(buf[:], int64(addr)); err != nil {
		errs.Add("WorldStateComponent", uintptr(addr), err)
		return result, errs
	}

	// Field: Header at int64(addr)+0
	{
		child, childErrs := ReadComponentHeader(ctx, uintptr(int64(addr)+0))
		if child != nil {
			result.Header = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: ChangedMaterials at int64(addr)+252
	{
		child, childErrs := ReadStdVectorHeader(ctx, uintptr(int64(addr)+252))
		if child != nil {
			result.ChangedMaterials = *child
		}
		errs = append(errs, childErrs...)
	}

	result.BiomeCryptCount = int32(binary.LittleEndian.Uint32(buf[264:]))
	result.GodsAfraid = int32(binary.LittleEndian.Uint32(buf[268:]))
	result.GodsImpressed = int32(binary.LittleEndian.Uint32(buf[272:]))
	result.GodsAfraidDamage = int32(binary.LittleEndian.Uint32(buf[276:]))
	result.GodsEnraged = int32(binary.LittleEndian.Uint32(buf[280:]))
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

func ReadLuaComponent(ctx *runtime.ReadContext, addr uintptr) (*LuaComponent, runtime.Errors) {
	var errs runtime.Errors
	result := &LuaComponent{}
	var buf [268]byte

	if _, err := ctx.ReadAt(buf[:], int64(addr)); err != nil {
		errs.Add("LuaComponent", uintptr(addr), err)
		return result, errs
	}

	// Field: Header at int64(addr)+0
	{
		child, childErrs := ReadComponentHeader(ctx, uintptr(int64(addr)+0))
		if child != nil {
			result.Header = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: ScriptSourceFile at int64(addr)+244
	{
		child, childErrs := ReadMsvcString(ctx, uintptr(int64(addr)+244))
		if child != nil {
			result.ScriptSourceFile = *child
		}
		errs = append(errs, childErrs...)
	}

	return result, errs
}

func ReadBiomeChunk(ctx *runtime.ReadContext, addr uintptr) (*BiomeChunk, runtime.Errors) {
	var errs runtime.Errors
	result := &BiomeChunk{}
	var buf [752]byte

	if _, err := ctx.ReadAt(buf[:], int64(addr)); err != nil {
		errs.Add("BiomeChunk", uintptr(addr), err)
		return result, errs
	}

	result.Vtable = binary.LittleEndian.Uint32(buf[0:])
	result.Unknown04 = binary.LittleEndian.Uint32(buf[4:])
	// Field: BiomeName at int64(addr)+8
	{
		child, childErrs := ReadMsvcString(ctx, uintptr(int64(addr)+8))
		if child != nil {
			result.BiomeName = *child
		}
		errs = append(errs, childErrs...)
	}

	result.WobbleEligible = buf[196]
	result.WavyEdge = buf[197]
	result.ForceOriginal = buf[198]
	result.UnknownC7 = buf[199]
	result.BiomeDataPtr = binary.LittleEndian.Uint32(buf[676:])
	// Field: XmlPath at int64(addr)+728
	{
		child, childErrs := ReadMsvcString(ctx, uintptr(int64(addr)+728))
		if child != nil {
			result.XmlPath = *child
		}
		errs = append(errs, childErrs...)
	}

	return result, errs
}

func ReadBiomeGrid(ctx *runtime.ReadContext, addr uintptr) (*BiomeGrid, runtime.Errors) {
	var errs runtime.Errors
	result := &BiomeGrid{}
	var buf [112]byte

	if _, err := ctx.ReadAt(buf[:], int64(addr)); err != nil {
		errs.Add("BiomeGrid", uintptr(addr), err)
		return result, errs
	}

	result.ScenesBegin = binary.LittleEndian.Uint32(buf[0:])
	result.ScenesEnd = binary.LittleEndian.Uint32(buf[4:])
	result.ScenesCapacityEnd = binary.LittleEndian.Uint32(buf[8:])
	result.ScenesAltBegin = binary.LittleEndian.Uint32(buf[12:])
	result.ScenesAltEnd = binary.LittleEndian.Uint32(buf[16:])
	result.ScenesAltCapacityEnd = binary.LittleEndian.Uint32(buf[20:])
	result.XShift = math.Float64frombits(binary.LittleEndian.Uint64(buf[56:]))
	result.YShift = math.Float64frombits(binary.LittleEndian.Uint64(buf[64:]))
	result.Width = int32(binary.LittleEndian.Uint32(buf[80:]))
	result.Height = int32(binary.LittleEndian.Uint32(buf[84:]))
	result.TotalCount = int32(binary.LittleEndian.Uint32(buf[88:]))
	result.Unknown5C = int32(binary.LittleEndian.Uint32(buf[92:]))
	result.Unknown60 = binary.LittleEndian.Uint32(buf[96:])
	result.Unknown64 = binary.LittleEndian.Uint32(buf[100:])
	result.ChunksPtr = binary.LittleEndian.Uint32(buf[104:])
	result.ChunksCount = int32(binary.LittleEndian.Uint32(buf[108:]))
	return result, errs
}

func ReadCellTexture(ctx *runtime.ReadContext, addr uintptr) (*CellTexture, runtime.Errors) {
	var errs runtime.Errors
	result := &CellTexture{}
	var buf [16]byte

	if _, err := ctx.ReadAt(buf[:], int64(addr)); err != nil {
		errs.Add("CellTexture", uintptr(addr), err)
		return result, errs
	}

	result.Width = int32(binary.LittleEndian.Uint32(buf[0:]))
	result.Height = int32(binary.LittleEndian.Uint32(buf[4:]))
	result.Unknown08 = binary.LittleEndian.Uint32(buf[8:])
	result.PixelDataPtr = binary.LittleEndian.Uint32(buf[12:])
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

func ReadLightComponent(ctx *runtime.ReadContext, addr uintptr) (*LightComponent, runtime.Errors) {
	var errs runtime.Errors
	result := &LightComponent{}
	var buf [100]byte

	if _, err := ctx.ReadAt(buf[:], int64(addr)); err != nil {
		errs.Add("LightComponent", uintptr(addr), err)
		return result, errs
	}

	// Field: Header at int64(addr)+0
	{
		child, childErrs := ReadComponentHeader(ctx, uintptr(int64(addr)+0))
		if child != nil {
			result.Header = *child
		}
		errs = append(errs, childErrs...)
	}

	result.InternalPtr = binary.LittleEndian.Uint32(buf[72:])
	result.Radius = math.Float32frombits(binary.LittleEndian.Uint32(buf[76:]))
	result.R = binary.LittleEndian.Uint32(buf[80:])
	result.G = binary.LittleEndian.Uint32(buf[84:])
	result.B = binary.LittleEndian.Uint32(buf[88:])
	result.OffsetX = math.Float32frombits(binary.LittleEndian.Uint32(buf[92:]))
	result.OffsetY = math.Float32frombits(binary.LittleEndian.Uint32(buf[96:]))
	return result, errs
}

func ReadGameEffectComponent(ctx *runtime.ReadContext, addr uintptr) (*GameEffectComponent, runtime.Errors) {
	var errs runtime.Errors
	result := &GameEffectComponent{}
	var buf [80]byte

	if _, err := ctx.ReadAt(buf[:], int64(addr)); err != nil {
		errs.Add("GameEffectComponent", uintptr(addr), err)
		return result, errs
	}

	// Field: Header at int64(addr)+0
	{
		child, childErrs := ReadComponentHeader(ctx, uintptr(int64(addr)+0))
		if child != nil {
			result.Header = *child
		}
		errs = append(errs, childErrs...)
	}

	result.Effect = int32(binary.LittleEndian.Uint32(buf[72:]))
	result.Frames = int32(binary.LittleEndian.Uint32(buf[76:]))
	return result, errs
}

func ReadCellFactory(ctx *runtime.ReadContext, addr uintptr) (*CellFactory, runtime.Errors) {
	var errs runtime.Errors
	result := &CellFactory{}
	var buf [44]byte

	if _, err := ctx.ReadAt(buf[:], int64(addr)); err != nil {
		errs.Add("CellFactory", uintptr(addr), err)
		return result, errs
	}

	result.CellDataArrayPtr = binary.LittleEndian.Uint32(buf[24:])
	result.MaterialCount = int32(binary.LittleEndian.Uint32(buf[36:]))
	result.Material0Color = binary.LittleEndian.Uint32(buf[40:])
	return result, errs
}

func ReadChunk(ctx *runtime.ReadContext, addr uintptr) (*Chunk, runtime.Errors) {
	var errs runtime.Errors
	result := &Chunk{}
	var buf [4]byte

	if _, err := ctx.ReadAt(buf[:], int64(addr)); err != nil {
		errs.Add("Chunk", uintptr(addr), err)
		return result, errs
	}

	result.CellSlotsPtr = binary.LittleEndian.Uint32(buf[0:])
	return result, errs
}

func ReadHitboxComponent(ctx *runtime.ReadContext, addr uintptr) (*HitboxComponent, runtime.Errors) {
	var errs runtime.Errors
	result := &HitboxComponent{}
	var buf [104]byte

	if _, err := ctx.ReadAt(buf[:], int64(addr)); err != nil {
		errs.Add("HitboxComponent", uintptr(addr), err)
		return result, errs
	}

	// Field: Header at int64(addr)+0
	{
		child, childErrs := ReadComponentHeader(ctx, uintptr(int64(addr)+0))
		if child != nil {
			result.Header = *child
		}
		errs = append(errs, childErrs...)
	}

	result.IsPlayer = buf[72] != 0
	result.IsEnemy = buf[73] != 0
	result.IsItem = buf[74] != 0
	result.AabbMinX = math.Float32frombits(binary.LittleEndian.Uint32(buf[76:]))
	result.AabbMaxX = math.Float32frombits(binary.LittleEndian.Uint32(buf[80:]))
	result.AabbMinY = math.Float32frombits(binary.LittleEndian.Uint32(buf[84:]))
	result.AabbMaxY = math.Float32frombits(binary.LittleEndian.Uint32(buf[88:]))
	result.DamageMultiplier = math.Float32frombits(binary.LittleEndian.Uint32(buf[92:]))
	result.OffsetX = math.Float32frombits(binary.LittleEndian.Uint32(buf[96:]))
	result.OffsetY = math.Float32frombits(binary.LittleEndian.Uint32(buf[100:]))
	return result, errs
}

func ReadItemComponent(ctx *runtime.ReadContext, addr uintptr) (*ItemComponent, runtime.Errors) {
	var errs runtime.Errors
	result := &ItemComponent{}
	var buf [110]byte

	if _, err := ctx.ReadAt(buf[:], int64(addr)); err != nil {
		errs.Add("ItemComponent", uintptr(addr), err)
		return result, errs
	}

	// Field: Header at int64(addr)+0
	{
		child, childErrs := ReadComponentHeader(ctx, uintptr(int64(addr)+0))
		if child != nil {
			result.Header = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: ItemName at int64(addr)+72
	{
		child, childErrs := ReadMsvcString(ctx, uintptr(int64(addr)+72))
		if child != nil {
			result.ItemName = *child
		}
		errs = append(errs, childErrs...)
	}

	result.IsStackable = buf[96] != 0
	result.IsConsumable = buf[97] != 0
	result.StatsCountAsItemPickUp = buf[98] != 0
	result.AutoPickup = buf[99] != 0
	result.Unknown64 = binary.LittleEndian.Uint32(buf[100:])
	result.UsesRemaining = int32(binary.LittleEndian.Uint32(buf[104:]))
	result.IsIdentified = buf[108] != 0
	result.IsFrozen = buf[109] != 0
	return result, errs
}

func ReadWorldManagerViewRect(ctx *runtime.ReadContext, addr uintptr) (*WorldManagerViewRect, runtime.Errors) {
	var errs runtime.Errors
	result := &WorldManagerViewRect{}
	var buf [76]byte

	if _, err := ctx.ReadAt(buf[:], int64(addr)); err != nil {
		errs.Add("WorldManagerViewRect", uintptr(addr), err)
		return result, errs
	}

	result.ViewX = math.Float32frombits(binary.LittleEndian.Uint32(buf[0:]))
	result.ViewY = math.Float32frombits(binary.LittleEndian.Uint32(buf[4:]))
	result.ViewWidth = math.Float32frombits(binary.LittleEndian.Uint32(buf[8:]))
	result.ViewHeight = math.Float32frombits(binary.LittleEndian.Uint32(buf[12:]))
	result.PBackgroundGrid = binary.LittleEndian.Uint32(buf[72:])
	return result, errs
}

func ReadCellData(ctx *runtime.ReadContext, addr uintptr) (*CellData, runtime.Errors) {
	var errs runtime.Errors
	result := &CellData{}
	var buf [656]byte

	if _, err := ctx.ReadAt(buf[:], int64(addr)); err != nil {
		errs.Add("CellData", uintptr(addr), err)
		return result, errs
	}

	// Field: Name at int64(addr)+0
	{
		child, childErrs := ReadMsvcString(ctx, uintptr(int64(addr)+0))
		if child != nil {
			result.Name = *child
		}
		errs = append(errs, childErrs...)
	}

	result.FallbackColor = binary.LittleEndian.Uint32(buf[100:])
	result.TexturePtr = binary.LittleEndian.Uint32(buf[136:])
	return result, errs
}

func ReadCellGrid(ctx *runtime.ReadContext, addr uintptr) (*CellGrid, runtime.Errors) {
	var errs runtime.Errors
	result := &CellGrid{}
	var buf [12]byte

	if _, err := ctx.ReadAt(buf[:], int64(addr)); err != nil {
		errs.Add("CellGrid", uintptr(addr), err)
		return result, errs
	}

	result.Vtable = binary.LittleEndian.Uint32(buf[0:])
	result.Unknown04 = binary.LittleEndian.Uint32(buf[4:])
	result.ChunkTablePtr = binary.LittleEndian.Uint32(buf[8:])
	return result, errs
}

func ReadCellMaterialInfo(ctx *runtime.ReadContext, addr uintptr) (*CellMaterialInfo, runtime.Errors) {
	var errs runtime.Errors
	result := &CellMaterialInfo{}
	var buf [52]byte

	if _, err := ctx.ReadAt(buf[:], int64(addr)); err != nil {
		errs.Add("CellMaterialInfo", uintptr(addr), err)
		return result, errs
	}

	result.MaterialId = int32(binary.LittleEndian.Uint32(buf[48:]))
	return result, errs
}

func ReadDeathMatchApp(ctx *runtime.ReadContext, addr uintptr) (*DeathMatchApp, runtime.Errors) {
	var errs runtime.Errors
	result := &DeathMatchApp{}
	var buf [208]byte

	if _, err := ctx.ReadAt(buf[:], int64(addr)); err != nil {
		errs.Add("DeathMatchApp", uintptr(addr), err)
		return result, errs
	}

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

func ReadCharacterDataComponent(ctx *runtime.ReadContext, addr uintptr) (*CharacterDataComponent, runtime.Errors) {
	var errs runtime.Errors
	result := &CharacterDataComponent{}
	var buf [280]byte

	if _, err := ctx.ReadAt(buf[:], int64(addr)); err != nil {
		errs.Add("CharacterDataComponent", uintptr(addr), err)
		return result, errs
	}

	// Field: Header at int64(addr)+0
	{
		child, childErrs := ReadComponentHeader(ctx, uintptr(int64(addr)+0))
		if child != nil {
			result.Header = *child
		}
		errs = append(errs, childErrs...)
	}

	result.Gravity = math.Float32frombits(binary.LittleEndian.Uint32(buf[136:]))
	result.FlyTimeMax = math.Float32frombits(binary.LittleEndian.Uint32(buf[140:]))
	result.IsOnGround = buf[184] != 0
	result.VelocityX = math.Float32frombits(binary.LittleEndian.Uint32(buf[264:]))
	result.VelocityY = math.Float32frombits(binary.LittleEndian.Uint32(buf[268:]))
	return result, errs
}

func ReadVelocityComponent(ctx *runtime.ReadContext, addr uintptr) (*VelocityComponent, runtime.Errors) {
	var errs runtime.Errors
	result := &VelocityComponent{}
	var buf [92]byte

	if _, err := ctx.ReadAt(buf[:], int64(addr)); err != nil {
		errs.Add("VelocityComponent", uintptr(addr), err)
		return result, errs
	}

	// Field: Header at int64(addr)+0
	{
		child, childErrs := ReadComponentHeader(ctx, uintptr(int64(addr)+0))
		if child != nil {
			result.Header = *child
		}
		errs = append(errs, childErrs...)
	}

	result.GravityX = math.Float32frombits(binary.LittleEndian.Uint32(buf[72:]))
	result.GravityY = math.Float32frombits(binary.LittleEndian.Uint32(buf[76:]))
	result.Mass = math.Float32frombits(binary.LittleEndian.Uint32(buf[80:]))
	result.AirFriction = math.Float32frombits(binary.LittleEndian.Uint32(buf[84:]))
	result.TerminalVelocity = math.Float32frombits(binary.LittleEndian.Uint32(buf[88:]))
	return result, errs
}

func ReadGameGlobals(ctx *runtime.ReadContext, addr uintptr) (*GameGlobals, runtime.Errors) {
	var errs runtime.Errors
	result := &GameGlobals{}
	var buf [416]byte

	if _, err := ctx.ReadAt(buf[:], int64(addr)); err != nil {
		errs.Add("GameGlobals", uintptr(addr), err)
		return result, errs
	}

	result.FrameCount = int32(binary.LittleEndian.Uint32(buf[0:]))
	result.PhysicsStepCount = int32(binary.LittleEndian.Uint32(buf[4:]))
	result.GameTime = math.Float32frombits(binary.LittleEndian.Uint32(buf[8:]))
	result.PWorldManager = binary.LittleEndian.Uint32(buf[12:])
	result.PChunkSystem = binary.LittleEndian.Uint32(buf[16:])
	result.PCellGrid = binary.LittleEndian.Uint32(buf[20:])
	result.PCellFactory = binary.LittleEndian.Uint32(buf[24:])
	result.Unknown1c = binary.LittleEndian.Uint32(buf[28:])
	result.PPhysicsWorld = binary.LittleEndian.Uint32(buf[32:])
	result.PAudioManager = binary.LittleEndian.Uint32(buf[36:])
	result.ViewportLeft = math.Float32frombits(binary.LittleEndian.Uint32(buf[384:]))
	result.ViewportTop = math.Float32frombits(binary.LittleEndian.Uint32(buf[388:]))
	result.ViewportRight = math.Float32frombits(binary.LittleEndian.Uint32(buf[392:]))
	result.ViewportBottom = math.Float32frombits(binary.LittleEndian.Uint32(buf[396:]))
	return result, errs
}

func ReadMaterialInventoryComponent(ctx *runtime.ReadContext, addr uintptr) (*MaterialInventoryComponent, runtime.Errors) {
	var errs runtime.Errors
	result := &MaterialInventoryComponent{}
	var buf [192]byte

	if _, err := ctx.ReadAt(buf[:], int64(addr)); err != nil {
		errs.Add("MaterialInventoryComponent", uintptr(addr), err)
		return result, errs
	}

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

func ReadChunkSystem(ctx *runtime.ReadContext, addr uintptr) (*ChunkSystem, runtime.Errors) {
	var errs runtime.Errors
	result := &ChunkSystem{}
	var buf [1292]byte

	if _, err := ctx.ReadAt(buf[:], int64(addr)); err != nil {
		errs.Add("ChunkSystem", uintptr(addr), err)
		return result, errs
	}

	result.Vtable = binary.LittleEndian.Uint32(buf[0:])
	// Field: CellGrid at int64(addr)+1280
	{
		child, childErrs := ReadCellGrid(ctx, uintptr(int64(addr)+1280))
		if child != nil {
			result.CellGrid = *child
		}
		errs = append(errs, childErrs...)
	}

	return result, errs
}

func ReadEntity(ctx *runtime.ReadContext, addr uintptr) (*Entity, runtime.Errors) {
	var errs runtime.Errors
	result := &Entity{}
	var buf [152]byte

	if _, err := ctx.ReadAt(buf[:], int64(addr)); err != nil {
		errs.Add("Entity", uintptr(addr), err)
		return result, errs
	}

	result.EntityId = int32(binary.LittleEndian.Uint32(buf[0:]))
	result.SlotIndex = int32(binary.LittleEndian.Uint32(buf[4:]))
	result.Unknown08 = binary.LittleEndian.Uint32(buf[8:])
	result.PendingKill = int32(binary.LittleEndian.Uint32(buf[12:]))
	result.Flags10 = binary.LittleEndian.Uint32(buf[16:])
	// Field: Name at int64(addr)+20
	{
		child, childErrs := ReadMsvcString(ctx, uintptr(int64(addr)+20))
		if child != nil {
			result.Name = *child
		}
		errs = append(errs, childErrs...)
	}

	result.Unknown2c = binary.LittleEndian.Uint32(buf[44:])
	copy(result.TagBitset[:], buf[48:112])
	result.PosX = math.Float32frombits(binary.LittleEndian.Uint32(buf[112:]))
	result.PosY = math.Float32frombits(binary.LittleEndian.Uint32(buf[116:]))
	result.RotCos = math.Float32frombits(binary.LittleEndian.Uint32(buf[120:]))
	result.RotSin = math.Float32frombits(binary.LittleEndian.Uint32(buf[124:]))
	result.RotNegSin = math.Float32frombits(binary.LittleEndian.Uint32(buf[128:]))
	result.RotCos2 = math.Float32frombits(binary.LittleEndian.Uint32(buf[132:]))
	result.ScaleX = math.Float32frombits(binary.LittleEndian.Uint32(buf[136:]))
	result.ScaleY = math.Float32frombits(binary.LittleEndian.Uint32(buf[140:]))
	result.ChildrenPtr = binary.LittleEndian.Uint32(buf[144:])
	result.ParentEntityPtr = binary.LittleEndian.Uint32(buf[148:])
	return result, errs
}

func ReadComponentBuffer(ctx *runtime.ReadContext, addr uintptr) (*ComponentBuffer, runtime.Errors) {
	var errs runtime.Errors
	result := &ComponentBuffer{}
	var buf [196]byte

	if _, err := ctx.ReadAt(buf[:], int64(addr)); err != nil {
		errs.Add("ComponentBuffer", uintptr(addr), err)
		return result, errs
	}

	result.Vtable = binary.LittleEndian.Uint32(buf[0:])
	result.Sentinel = int32(binary.LittleEndian.Uint32(buf[4:]))
	result.InitialCapacity = int32(binary.LittleEndian.Uint32(buf[8:]))
	result.Unknown0c = binary.LittleEndian.Uint32(buf[12:])
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

	result.ActiveCount = int32(binary.LittleEndian.Uint32(buf[152:]))
	result.CapacityLimit = int32(binary.LittleEndian.Uint32(buf[156:]))
	result.UnknownA0 = binary.LittleEndian.Uint32(buf[160:])
	result.PEntityManager = binary.LittleEndian.Uint32(buf[164:])
	result.PEventManager = binary.LittleEndian.Uint32(buf[168:])
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

func ReadWalletComponent(ctx *runtime.ReadContext, addr uintptr) (*WalletComponent, runtime.Errors) {
	var errs runtime.Errors
	result := &WalletComponent{}
	var buf [100]byte

	if _, err := ctx.ReadAt(buf[:], int64(addr)); err != nil {
		errs.Add("WalletComponent", uintptr(addr), err)
		return result, errs
	}

	// Field: Header at int64(addr)+0
	{
		child, childErrs := ReadComponentHeader(ctx, uintptr(int64(addr)+0))
		if child != nil {
			result.Header = *child
		}
		errs = append(errs, childErrs...)
	}

	result.Money = int64(binary.LittleEndian.Uint64(buf[72:]))
	result.MoneySpent = int64(binary.LittleEndian.Uint64(buf[80:]))
	result.MoneyPrevFrame = int64(binary.LittleEndian.Uint64(buf[88:]))
	result.HasReachedInf = buf[96] != 0
	return result, errs
}

func ReadCollisionTriggerComponent(ctx *runtime.ReadContext, addr uintptr) (*CollisionTriggerComponent, runtime.Errors) {
	var errs runtime.Errors
	result := &CollisionTriggerComponent{}
	var buf [108]byte

	if _, err := ctx.ReadAt(buf[:], int64(addr)); err != nil {
		errs.Add("CollisionTriggerComponent", uintptr(addr), err)
		return result, errs
	}

	// Field: Header at int64(addr)+0
	{
		child, childErrs := ReadComponentHeader(ctx, uintptr(int64(addr)+0))
		if child != nil {
			result.Header = *child
		}
		errs = append(errs, childErrs...)
	}

	result.Width = math.Float32frombits(binary.LittleEndian.Uint32(buf[72:]))
	result.Height = math.Float32frombits(binary.LittleEndian.Uint32(buf[76:]))
	result.Radius = math.Float32frombits(binary.LittleEndian.Uint32(buf[80:]))
	// Field: RequiredTag at int64(addr)+84
	{
		child, childErrs := ReadMsvcString(ctx, uintptr(int64(addr)+84))
		if child != nil {
			result.RequiredTag = *child
		}
		errs = append(errs, childErrs...)
	}

	return result, errs
}

func ReadInventory2Component(ctx *runtime.ReadContext, addr uintptr) (*Inventory2Component, runtime.Errors) {
	var errs runtime.Errors
	result := &Inventory2Component{}
	var buf [128]byte

	if _, err := ctx.ReadAt(buf[:], int64(addr)); err != nil {
		errs.Add("Inventory2Component", uintptr(addr), err)
		return result, errs
	}

	// Field: Header at int64(addr)+0
	{
		child, childErrs := ReadComponentHeader(ctx, uintptr(int64(addr)+0))
		if child != nil {
			result.Header = *child
		}
		errs = append(errs, childErrs...)
	}

	result.QuickInventorySlots = int32(binary.LittleEndian.Uint32(buf[72:]))
	result.FullInventorySlotsX = int32(binary.LittleEndian.Uint32(buf[76:]))
	result.FullInventorySlotsY = int32(binary.LittleEndian.Uint32(buf[80:]))
	result.SavedActiveItemIndex = int32(binary.LittleEndian.Uint32(buf[84:]))
	result.ActiveItem = int32(binary.LittleEndian.Uint32(buf[88:]))
	result.ActualActiveItem = int32(binary.LittleEndian.Uint32(buf[92:]))
	result.ActiveStash = int32(binary.LittleEndian.Uint32(buf[96:]))
	result.ThrowItem = int32(binary.LittleEndian.Uint32(buf[100:]))
	result.ItemHolstered = buf[104] != 0
	result.Initialized = buf[105] != 0
	result.ForceRefresh = buf[106] != 0
	result.DontLogNextItemEquip = buf[107] != 0
	result.SmoothedItemXOffset = math.Float32frombits(binary.LittleEndian.Uint32(buf[108:]))
	result.LastItemSwitchFrame = int32(binary.LittleEndian.Uint32(buf[112:]))
	result.IntroEquipItemLerp = math.Float32frombits(binary.LittleEndian.Uint32(buf[116:]))
	result.SmoothedItemAngleX = math.Float32frombits(binary.LittleEndian.Uint32(buf[120:]))
	result.SmoothedItemAngleY = math.Float32frombits(binary.LittleEndian.Uint32(buf[124:]))
	return result, errs
}

func ReadPixelSceneEntry(ctx *runtime.ReadContext, addr uintptr) (*PixelSceneEntry, runtime.Errors) {
	var errs runtime.Errors
	result := &PixelSceneEntry{}
	var buf [144]byte

	if _, err := ctx.ReadAt(buf[:], int64(addr)); err != nil {
		errs.Add("PixelSceneEntry", uintptr(addr), err)
		return result, errs
	}

	result.ChunkSystemBackRef = binary.LittleEndian.Uint32(buf[0:])
	result.X = int32(binary.LittleEndian.Uint32(buf[4:]))
	result.Y = int32(binary.LittleEndian.Uint32(buf[8:]))
	// Field: MaterialsFilename at int64(addr)+12
	{
		child, childErrs := ReadMsvcString(ctx, uintptr(int64(addr)+12))
		if child != nil {
			result.MaterialsFilename = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: ColorsFilename at int64(addr)+36
	{
		child, childErrs := ReadMsvcString(ctx, uintptr(int64(addr)+36))
		if child != nil {
			result.ColorsFilename = *child
		}
		errs = append(errs, childErrs...)
	}

	// Field: BackgroundFilename at int64(addr)+60
	{
		child, childErrs := ReadMsvcString(ctx, uintptr(int64(addr)+60))
		if child != nil {
			result.BackgroundFilename = *child
		}
		errs = append(errs, childErrs...)
	}

	result.FlagSkipBiomeChecks = buf[88]
	result.FlagSkipEdgeTextures = buf[89]
	return result, errs
}

func ReadEntityManager(ctx *runtime.ReadContext, addr uintptr) (*EntityManager, runtime.Errors) {
	var errs runtime.Errors
	result := &EntityManager{}
	var buf [60]byte

	if _, err := ctx.ReadAt(buf[:], int64(addr)); err != nil {
		errs.Add("EntityManager", uintptr(addr), err)
		return result, errs
	}

	result.Vtable = binary.LittleEndian.Uint32(buf[0:])
	result.NextEntityId = int32(binary.LittleEndian.Uint32(buf[4:]))
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

	result.PEventManager = binary.LittleEndian.Uint32(buf[56:])
	return result, errs
}

func ReadDamageModelComponent(ctx *runtime.ReadContext, addr uintptr) (*DamageModelComponent, runtime.Errors) {
	var errs runtime.Errors
	result := &DamageModelComponent{}
	var buf [732]byte

	if _, err := ctx.ReadAt(buf[:], int64(addr)); err != nil {
		errs.Add("DamageModelComponent", uintptr(addr), err)
		return result, errs
	}

	// Field: Header at int64(addr)+0
	{
		child, childErrs := ReadComponentHeader(ctx, uintptr(int64(addr)+0))
		if child != nil {
			result.Header = *child
		}
		errs = append(errs, childErrs...)
	}

	result.Hp = math.Float64frombits(binary.LittleEndian.Uint64(buf[72:]))
	result.MaxHp = math.Float64frombits(binary.LittleEndian.Uint64(buf[80:]))
	result.MaxHpCap = math.Float64frombits(binary.LittleEndian.Uint64(buf[88:]))
	result.MaxHpOld = math.Float64frombits(binary.LittleEndian.Uint64(buf[96:]))
	result.DamageMultipliersVtable = binary.LittleEndian.Uint32(buf[104:])
	result.DamageMultipliersMelee = math.Float32frombits(binary.LittleEndian.Uint32(buf[108:]))
	result.DamageMultipliersProjectile = math.Float32frombits(binary.LittleEndian.Uint32(buf[112:]))
	result.DamageMultipliersExplosion = math.Float32frombits(binary.LittleEndian.Uint32(buf[116:]))
	result.DamageMultipliersElectricity = math.Float32frombits(binary.LittleEndian.Uint32(buf[120:]))
	result.DamageMultipliersFire = math.Float32frombits(binary.LittleEndian.Uint32(buf[124:]))
	result.DamageMultipliersDrill = math.Float32frombits(binary.LittleEndian.Uint32(buf[128:]))
	result.DamageMultipliersSlice = math.Float32frombits(binary.LittleEndian.Uint32(buf[132:]))
	result.DamageMultipliersIce = math.Float32frombits(binary.LittleEndian.Uint32(buf[136:]))
	result.DamageMultipliersHealing = math.Float32frombits(binary.LittleEndian.Uint32(buf[140:]))
	result.DamageMultipliersPhysicsHit = math.Float32frombits(binary.LittleEndian.Uint32(buf[144:]))
	result.DamageMultipliersRadioactive = math.Float32frombits(binary.LittleEndian.Uint32(buf[148:]))
	result.DamageMultipliersPoison = math.Float32frombits(binary.LittleEndian.Uint32(buf[152:]))
	result.DamageMultipliersOvereating = math.Float32frombits(binary.LittleEndian.Uint32(buf[156:]))
	result.DamageMultipliersCurse = math.Float32frombits(binary.LittleEndian.Uint32(buf[160:]))
	result.DamageMultipliersHoly = math.Float32frombits(binary.LittleEndian.Uint32(buf[164:]))
	result.CriticalDamageResistance = math.Float32frombits(binary.LittleEndian.Uint32(buf[168:]))
	result.InvincibilityFrames = int32(binary.LittleEndian.Uint32(buf[172:]))
	return result, errs
}

// ReadPBackgroundGrid follows the PBackgroundGrid pointer and reads the target BiomeGrid.
func (s *WorldManagerViewRect) ReadPBackgroundGrid(ctx *runtime.ReadContext) (*BiomeGrid, runtime.Errors) {
	if s.PBackgroundGrid == 0 {
		return nil, nil
	}
	return ReadBiomeGrid(ctx, uintptr(s.PBackgroundGrid))
}

// ReadTexturePtr follows the TexturePtr pointer and reads the target CellTexture.
func (s *CellData) ReadTexturePtr(ctx *runtime.ReadContext) (*CellTexture, runtime.Errors) {
	if s.TexturePtr == 0 {
		return nil, nil
	}
	return ReadCellTexture(ctx, uintptr(s.TexturePtr))
}

// ReadPWorldManager follows the PWorldManager pointer and reads the target WorldManagerViewRect.
func (s *GameGlobals) ReadPWorldManager(ctx *runtime.ReadContext) (*WorldManagerViewRect, runtime.Errors) {
	if s.PWorldManager == 0 {
		return nil, nil
	}
	return ReadWorldManagerViewRect(ctx, uintptr(s.PWorldManager))
}

// ReadPChunkSystem follows the PChunkSystem pointer and reads the target ChunkSystem.
func (s *GameGlobals) ReadPChunkSystem(ctx *runtime.ReadContext) (*ChunkSystem, runtime.Errors) {
	if s.PChunkSystem == 0 {
		return nil, nil
	}
	return ReadChunkSystem(ctx, uintptr(s.PChunkSystem))
}

// ReadPCellFactory follows the PCellFactory pointer and reads the target CellFactory.
func (s *GameGlobals) ReadPCellFactory(ctx *runtime.ReadContext) (*CellFactory, runtime.Errors) {
	if s.PCellFactory == 0 {
		return nil, nil
	}
	return ReadCellFactory(ctx, uintptr(s.PCellFactory))
}

// ReadChildrenPtr follows the ChildrenPtr pointer and reads the target ChildrenContainer.
func (s *Entity) ReadChildrenPtr(ctx *runtime.ReadContext) (*ChildrenContainer, runtime.Errors) {
	if s.ChildrenPtr == 0 {
		return nil, nil
	}
	return ReadChildrenContainer(ctx, uintptr(s.ChildrenPtr))
}

// ReadParentEntityPtr follows the ParentEntityPtr pointer and reads the target Entity.
func (s *Entity) ReadParentEntityPtr(ctx *runtime.ReadContext) (*Entity, runtime.Errors) {
	if s.ParentEntityPtr == 0 {
		return nil, nil
	}
	return ReadEntity(ctx, uintptr(s.ParentEntityPtr))
}

// Static address constants for top-level placements.
const (
	AddrGWorldSeed         uint32 = 0x01205004
	AddrGNgPlusCount       uint32 = 0x01205024
	AddrGDeathCount        uint32 = 0x01208AF8
	AddrGNumOrbsTotal      uint32 = 0x01152544
	AddrGEntityManager     uint32 = 0x01204B98
	AddrGDeathMatchApp     uint32 = 0x01204BC0
	AddrGGameGlobals       uint32 = 0x0122374C
	AddrGWorldState        uint32 = 0x01205010
	AddrGOrbPersistencePtr uint32 = 0x01207404
	AddrGameglobalsHeap    uint32 = 0x01BBAE58
	AddrDeathmatchappHeap  uint32 = 0x01B93510
	AddrEntitymanagerHeap  uint32 = 0x06BA7F88
	AddrPlayerEntity       uint32 = 0x1673E6C8
	AddrPlayerDmg          uint32 = 0x3334B780
	AddrWand0              uint32 = 0x0648EED0
	AddrWand1              uint32 = 0x0648F2B0
)

// ReadGWorldSeed reads the uint32 at AddrGWorldSeed.
func ReadGWorldSeed(ctx *runtime.ReadContext) (uint32, error) {
	var buf [4]byte
	if _, err := ctx.ReadAt(buf[:], int64(AddrGWorldSeed)); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:]), nil
}

// ReadGNgPlusCount reads the int32 at AddrGNgPlusCount.
func ReadGNgPlusCount(ctx *runtime.ReadContext) (int32, error) {
	var buf [4]byte
	if _, err := ctx.ReadAt(buf[:], int64(AddrGNgPlusCount)); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:])), nil
}

// ReadGDeathCount reads the int32 at AddrGDeathCount.
func ReadGDeathCount(ctx *runtime.ReadContext) (int32, error) {
	var buf [4]byte
	if _, err := ctx.ReadAt(buf[:], int64(AddrGDeathCount)); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:])), nil
}

// ReadGNumOrbsTotal reads the int32 at AddrGNumOrbsTotal.
func ReadGNumOrbsTotal(ctx *runtime.ReadContext) (int32, error) {
	var buf [4]byte
	if _, err := ctx.ReadAt(buf[:], int64(AddrGNumOrbsTotal)); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:])), nil
}

// ReadGEntityManager reads the pointer at AddrGEntityManager and follows it to EntityManager.
func ReadGEntityManager(ctx *runtime.ReadContext) (*EntityManager, runtime.Errors) {
	var buf [4]byte
	if _, err := ctx.ReadAt(buf[:], int64(AddrGEntityManager)); err != nil {
		var errs runtime.Errors
		errs.Add("g_entityManager", uintptr(AddrGEntityManager), err)
		return nil, errs
	}
	ptr := binary.LittleEndian.Uint32(buf[:])
	if ptr == 0 {
		return nil, nil
	}
	return ReadEntityManager(ctx, uintptr(ptr))
}

// ReadGDeathMatchApp reads the pointer at AddrGDeathMatchApp and follows it to DeathMatchApp.
func ReadGDeathMatchApp(ctx *runtime.ReadContext) (*DeathMatchApp, runtime.Errors) {
	var buf [4]byte
	if _, err := ctx.ReadAt(buf[:], int64(AddrGDeathMatchApp)); err != nil {
		var errs runtime.Errors
		errs.Add("g_deathMatchApp", uintptr(AddrGDeathMatchApp), err)
		return nil, errs
	}
	ptr := binary.LittleEndian.Uint32(buf[:])
	if ptr == 0 {
		return nil, nil
	}
	return ReadDeathMatchApp(ctx, uintptr(ptr))
}

// ReadGGameGlobals reads the pointer at AddrGGameGlobals and follows it to GameGlobals.
func ReadGGameGlobals(ctx *runtime.ReadContext) (*GameGlobals, runtime.Errors) {
	var buf [4]byte
	if _, err := ctx.ReadAt(buf[:], int64(AddrGGameGlobals)); err != nil {
		var errs runtime.Errors
		errs.Add("g_gameGlobals", uintptr(AddrGGameGlobals), err)
		return nil, errs
	}
	ptr := binary.LittleEndian.Uint32(buf[:])
	if ptr == 0 {
		return nil, nil
	}
	return ReadGameGlobals(ctx, uintptr(ptr))
}

// ReadGWorldState reads the pointer at AddrGWorldState and follows it to WorldStateComponent.
func ReadGWorldState(ctx *runtime.ReadContext) (*WorldStateComponent, runtime.Errors) {
	var buf [4]byte
	if _, err := ctx.ReadAt(buf[:], int64(AddrGWorldState)); err != nil {
		var errs runtime.Errors
		errs.Add("g_worldState", uintptr(AddrGWorldState), err)
		return nil, errs
	}
	ptr := binary.LittleEndian.Uint32(buf[:])
	if ptr == 0 {
		return nil, nil
	}
	return ReadWorldStateComponent(ctx, uintptr(ptr))
}

// ReadGOrbPersistencePtr reads the uint32 at AddrGOrbPersistencePtr.
func ReadGOrbPersistencePtr(ctx *runtime.ReadContext) (uint32, error) {
	var buf [4]byte
	if _, err := ctx.ReadAt(buf[:], int64(AddrGOrbPersistencePtr)); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:]), nil
}

// ReadGameglobalsHeap reads GameGlobals at AddrGameglobalsHeap.
func ReadGameglobalsHeap(ctx *runtime.ReadContext) (*GameGlobals, runtime.Errors) {
	return ReadGameGlobals(ctx, uintptr(AddrGameglobalsHeap))
}

// ReadDeathmatchappHeap reads DeathMatchApp at AddrDeathmatchappHeap.
func ReadDeathmatchappHeap(ctx *runtime.ReadContext) (*DeathMatchApp, runtime.Errors) {
	return ReadDeathMatchApp(ctx, uintptr(AddrDeathmatchappHeap))
}

// ReadEntitymanagerHeap reads EntityManager at AddrEntitymanagerHeap.
func ReadEntitymanagerHeap(ctx *runtime.ReadContext) (*EntityManager, runtime.Errors) {
	return ReadEntityManager(ctx, uintptr(AddrEntitymanagerHeap))
}

// ReadPlayerEntity reads Entity at AddrPlayerEntity.
func ReadPlayerEntity(ctx *runtime.ReadContext) (*Entity, runtime.Errors) {
	return ReadEntity(ctx, uintptr(AddrPlayerEntity))
}

// ReadPlayerDmg reads DamageModelComponent at AddrPlayerDmg.
func ReadPlayerDmg(ctx *runtime.ReadContext) (*DamageModelComponent, runtime.Errors) {
	return ReadDamageModelComponent(ctx, uintptr(AddrPlayerDmg))
}

// ReadWand0 reads AbilityComponent at AddrWand0.
func ReadWand0(ctx *runtime.ReadContext) (*AbilityComponent, runtime.Errors) {
	return ReadAbilityComponent(ctx, uintptr(AddrWand0))
}

// ReadWand1 reads AbilityComponent at AddrWand1.
func ReadWand1(ctx *runtime.ReadContext) (*AbilityComponent, runtime.Errors) {
	return ReadAbilityComponent(ctx, uintptr(AddrWand1))
}

// FormatVector is transpiled from hexpat function format_vector.
func (s *StdVectorHeader) FormatVector() string {
	if s.BeginPtr == 0 {
		return "empty"
	}
	count := ((s.EndPtr - s.BeginPtr) / 4)
	return fmt.Sprintf("%v elements @ 0x%08X", count, s.BeginPtr)
}

// FormatChildren is transpiled from hexpat function format_children.
func (s *ChildrenContainer) FormatChildren() string {
	if s.BeginPtr == 0 {
		return "no children"
	}
	count := ((s.EndPtr - s.BeginPtr) / 4)
	return fmt.Sprintf("%v children", count)
}

// FormatMsvcString is transpiled from hexpat function format_msvc_string.
func (s *MsvcString) FormatMsvcString(ctx *runtime.ReadContext) string {
	if s.Length == 0 {
		return ""
	}
	if s.Capacity <= 15 {
		result := ""
		for i := uint32(0); i < s.Length; i = (i + 1) {
			result = result + string(s.Data[i])
		}
		return result
	}
	heapPtr := (((uint32(s.Data[0]) | (uint32(s.Data[1]) << 8)) | (uint32(s.Data[2]) << 16)) | (uint32(s.Data[3]) << 24))
	if heapPtr == 0 {
		return ""
	}
	return _memReadString(ctx, uint64(heapPtr), uint64(s.Length))
}

// _memReadString reads a string of the given length from an address in process memory.
func _memReadString(ctx *runtime.ReadContext, addr, length uint64) string {
	if length == 0 {
		return ""
	}
	if length > 4096 {
		length = 4096
	}
	buf := make([]byte, length)
	if _, err := ctx.ReadAt(buf, int64(addr)); err != nil {
		return ""
	}
	return string(buf)
}

// _memReadUnsigned reads an unsigned integer of the given byte size from an address.
func _memReadUnsigned(ctx *runtime.ReadContext, addr, size uint64) uint64 {
	var buf [8]byte
	if size > 8 {
		size = 8
	}
	if _, err := ctx.ReadAt(buf[:size], int64(addr)); err != nil {
		return 0
	}
	switch size {
	case 1:
		return uint64(buf[0])
	case 2:
		return uint64(binary.LittleEndian.Uint16(buf[:2]))
	case 4:
		return uint64(binary.LittleEndian.Uint32(buf[:4]))
	case 8:
		return binary.LittleEndian.Uint64(buf[:8])
	default:
		return 0
	}
}

// _memReadSigned reads a signed integer of the given byte size from an address.
func _memReadSigned(ctx *runtime.ReadContext, addr, size uint64) int64 {
	return int64(_memReadUnsigned(ctx, addr, size))
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

// SpriteComponentReader provides lazy, field-level access to SpriteComponent without reading the entire struct.
type SpriteComponentReader struct {
	ctx  *runtime.ReadContext
	addr uintptr
}

// NewSpriteComponentReader creates a lazy reader for SpriteComponent at the given address.
func NewSpriteComponentReader(ctx *runtime.ReadContext, addr uintptr) *SpriteComponentReader {
	return &SpriteComponentReader{ctx: ctx, addr: addr}
}

// Addr returns the base address of this SpriteComponent.
func (r *SpriteComponentReader) Addr() uintptr {
	return r.addr
}

// Read materializes the full SpriteComponent struct eagerly.
func (r *SpriteComponentReader) Read() (*SpriteComponent, runtime.Errors) {
	return ReadSpriteComponent(r.ctx, r.addr)
}

// Header returns a lazy reader for the nested ComponentHeader (zero I/O).
func (r *SpriteComponentReader) Header() *ComponentHeaderReader {
	return NewComponentHeaderReader(r.ctx, uintptr(int64(r.addr)+0))
}

// ImageFile returns a lazy reader for the nested MsvcString (zero I/O).
func (r *SpriteComponentReader) ImageFile() *MsvcStringReader {
	return NewMsvcStringReader(r.ctx, uintptr(int64(r.addr)+72))
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

// ChangedMaterials returns a lazy reader for the nested StdVectorHeader (zero I/O).
func (r *WorldStateComponentReader) ChangedMaterials() *StdVectorHeaderReader {
	return NewStdVectorHeaderReader(r.ctx, uintptr(int64(r.addr)+252))
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

// LuaComponentReader provides lazy, field-level access to LuaComponent without reading the entire struct.
type LuaComponentReader struct {
	ctx  *runtime.ReadContext
	addr uintptr
}

// NewLuaComponentReader creates a lazy reader for LuaComponent at the given address.
func NewLuaComponentReader(ctx *runtime.ReadContext, addr uintptr) *LuaComponentReader {
	return &LuaComponentReader{ctx: ctx, addr: addr}
}

// Addr returns the base address of this LuaComponent.
func (r *LuaComponentReader) Addr() uintptr {
	return r.addr
}

// Read materializes the full LuaComponent struct eagerly.
func (r *LuaComponentReader) Read() (*LuaComponent, runtime.Errors) {
	return ReadLuaComponent(r.ctx, r.addr)
}

// Header returns a lazy reader for the nested ComponentHeader (zero I/O).
func (r *LuaComponentReader) Header() *ComponentHeaderReader {
	return NewComponentHeaderReader(r.ctx, uintptr(int64(r.addr)+0))
}

// ScriptSourceFile returns a lazy reader for the nested MsvcString (zero I/O).
func (r *LuaComponentReader) ScriptSourceFile() *MsvcStringReader {
	return NewMsvcStringReader(r.ctx, uintptr(int64(r.addr)+244))
}

// BiomeChunkReader provides lazy, field-level access to BiomeChunk without reading the entire struct.
type BiomeChunkReader struct {
	ctx  *runtime.ReadContext
	addr uintptr
}

// NewBiomeChunkReader creates a lazy reader for BiomeChunk at the given address.
func NewBiomeChunkReader(ctx *runtime.ReadContext, addr uintptr) *BiomeChunkReader {
	return &BiomeChunkReader{ctx: ctx, addr: addr}
}

// Addr returns the base address of this BiomeChunk.
func (r *BiomeChunkReader) Addr() uintptr {
	return r.addr
}

// Read materializes the full BiomeChunk struct eagerly.
func (r *BiomeChunkReader) Read() (*BiomeChunk, runtime.Errors) {
	return ReadBiomeChunk(r.ctx, r.addr)
}

func (r *BiomeChunkReader) Vtable() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+0); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *BiomeChunkReader) Unknown04() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+4); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

// BiomeName returns a lazy reader for the nested MsvcString (zero I/O).
func (r *BiomeChunkReader) BiomeName() *MsvcStringReader {
	return NewMsvcStringReader(r.ctx, uintptr(int64(r.addr)+8))
}

func (r *BiomeChunkReader) WobbleEligible() (uint8, error) {
	var buf [1]byte
	if _, err := r.ctx.ReadAt(buf[:1], int64(r.addr)+196); err != nil {
		return 0, err
	}
	return buf[0], nil
}

func (r *BiomeChunkReader) WavyEdge() (uint8, error) {
	var buf [1]byte
	if _, err := r.ctx.ReadAt(buf[:1], int64(r.addr)+197); err != nil {
		return 0, err
	}
	return buf[0], nil
}

func (r *BiomeChunkReader) ForceOriginal() (uint8, error) {
	var buf [1]byte
	if _, err := r.ctx.ReadAt(buf[:1], int64(r.addr)+198); err != nil {
		return 0, err
	}
	return buf[0], nil
}

func (r *BiomeChunkReader) UnknownC7() (uint8, error) {
	var buf [1]byte
	if _, err := r.ctx.ReadAt(buf[:1], int64(r.addr)+199); err != nil {
		return 0, err
	}
	return buf[0], nil
}

func (r *BiomeChunkReader) BiomeDataPtr() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+676); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

// XmlPath returns a lazy reader for the nested MsvcString (zero I/O).
func (r *BiomeChunkReader) XmlPath() *MsvcStringReader {
	return NewMsvcStringReader(r.ctx, uintptr(int64(r.addr)+728))
}

// BiomeGridReader provides lazy, field-level access to BiomeGrid without reading the entire struct.
type BiomeGridReader struct {
	ctx  *runtime.ReadContext
	addr uintptr
}

// NewBiomeGridReader creates a lazy reader for BiomeGrid at the given address.
func NewBiomeGridReader(ctx *runtime.ReadContext, addr uintptr) *BiomeGridReader {
	return &BiomeGridReader{ctx: ctx, addr: addr}
}

// Addr returns the base address of this BiomeGrid.
func (r *BiomeGridReader) Addr() uintptr {
	return r.addr
}

// Read materializes the full BiomeGrid struct eagerly.
func (r *BiomeGridReader) Read() (*BiomeGrid, runtime.Errors) {
	return ReadBiomeGrid(r.ctx, r.addr)
}

func (r *BiomeGridReader) ScenesBegin() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+0); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *BiomeGridReader) ScenesEnd() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+4); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *BiomeGridReader) ScenesCapacityEnd() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+8); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *BiomeGridReader) ScenesAltBegin() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+12); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *BiomeGridReader) ScenesAltEnd() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+16); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *BiomeGridReader) ScenesAltCapacityEnd() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+20); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *BiomeGridReader) XShift() (float64, error) {
	var buf [8]byte
	if _, err := r.ctx.ReadAt(buf[:8], int64(r.addr)+56); err != nil {
		return 0, err
	}
	return math.Float64frombits(binary.LittleEndian.Uint64(buf[:8])), nil
}

func (r *BiomeGridReader) YShift() (float64, error) {
	var buf [8]byte
	if _, err := r.ctx.ReadAt(buf[:8], int64(r.addr)+64); err != nil {
		return 0, err
	}
	return math.Float64frombits(binary.LittleEndian.Uint64(buf[:8])), nil
}

func (r *BiomeGridReader) Width() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+80); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *BiomeGridReader) Height() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+84); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *BiomeGridReader) TotalCount() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+88); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *BiomeGridReader) Unknown5C() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+92); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *BiomeGridReader) Unknown60() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+96); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *BiomeGridReader) Unknown64() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+100); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *BiomeGridReader) ChunksPtr() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+104); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *BiomeGridReader) ChunksCount() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+108); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

// CellTextureReader provides lazy, field-level access to CellTexture without reading the entire struct.
type CellTextureReader struct {
	ctx  *runtime.ReadContext
	addr uintptr
}

// NewCellTextureReader creates a lazy reader for CellTexture at the given address.
func NewCellTextureReader(ctx *runtime.ReadContext, addr uintptr) *CellTextureReader {
	return &CellTextureReader{ctx: ctx, addr: addr}
}

// Addr returns the base address of this CellTexture.
func (r *CellTextureReader) Addr() uintptr {
	return r.addr
}

// Read materializes the full CellTexture struct eagerly.
func (r *CellTextureReader) Read() (*CellTexture, runtime.Errors) {
	return ReadCellTexture(r.ctx, r.addr)
}

func (r *CellTextureReader) Width() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+0); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *CellTextureReader) Height() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+4); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *CellTextureReader) Unknown08() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+8); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *CellTextureReader) PixelDataPtr() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+12); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

// LightComponentReader provides lazy, field-level access to LightComponent without reading the entire struct.
type LightComponentReader struct {
	ctx  *runtime.ReadContext
	addr uintptr
}

// NewLightComponentReader creates a lazy reader for LightComponent at the given address.
func NewLightComponentReader(ctx *runtime.ReadContext, addr uintptr) *LightComponentReader {
	return &LightComponentReader{ctx: ctx, addr: addr}
}

// Addr returns the base address of this LightComponent.
func (r *LightComponentReader) Addr() uintptr {
	return r.addr
}

// Read materializes the full LightComponent struct eagerly.
func (r *LightComponentReader) Read() (*LightComponent, runtime.Errors) {
	return ReadLightComponent(r.ctx, r.addr)
}

// Header returns a lazy reader for the nested ComponentHeader (zero I/O).
func (r *LightComponentReader) Header() *ComponentHeaderReader {
	return NewComponentHeaderReader(r.ctx, uintptr(int64(r.addr)+0))
}

func (r *LightComponentReader) InternalPtr() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+72); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *LightComponentReader) Radius() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+76); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *LightComponentReader) R() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+80); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *LightComponentReader) G() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+84); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *LightComponentReader) B() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+88); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *LightComponentReader) OffsetX() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+92); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *LightComponentReader) OffsetY() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+96); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

// GameEffectComponentReader provides lazy, field-level access to GameEffectComponent without reading the entire struct.
type GameEffectComponentReader struct {
	ctx  *runtime.ReadContext
	addr uintptr
}

// NewGameEffectComponentReader creates a lazy reader for GameEffectComponent at the given address.
func NewGameEffectComponentReader(ctx *runtime.ReadContext, addr uintptr) *GameEffectComponentReader {
	return &GameEffectComponentReader{ctx: ctx, addr: addr}
}

// Addr returns the base address of this GameEffectComponent.
func (r *GameEffectComponentReader) Addr() uintptr {
	return r.addr
}

// Read materializes the full GameEffectComponent struct eagerly.
func (r *GameEffectComponentReader) Read() (*GameEffectComponent, runtime.Errors) {
	return ReadGameEffectComponent(r.ctx, r.addr)
}

// Header returns a lazy reader for the nested ComponentHeader (zero I/O).
func (r *GameEffectComponentReader) Header() *ComponentHeaderReader {
	return NewComponentHeaderReader(r.ctx, uintptr(int64(r.addr)+0))
}

func (r *GameEffectComponentReader) Effect() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+72); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *GameEffectComponentReader) Frames() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+76); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
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

func (r *CellFactoryReader) MaterialCount() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+36); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *CellFactoryReader) Material0Color() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+40); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

// ChunkReader provides lazy, field-level access to Chunk without reading the entire struct.
type ChunkReader struct {
	ctx  *runtime.ReadContext
	addr uintptr
}

// NewChunkReader creates a lazy reader for Chunk at the given address.
func NewChunkReader(ctx *runtime.ReadContext, addr uintptr) *ChunkReader {
	return &ChunkReader{ctx: ctx, addr: addr}
}

// Addr returns the base address of this Chunk.
func (r *ChunkReader) Addr() uintptr {
	return r.addr
}

// Read materializes the full Chunk struct eagerly.
func (r *ChunkReader) Read() (*Chunk, runtime.Errors) {
	return ReadChunk(r.ctx, r.addr)
}

func (r *ChunkReader) CellSlotsPtr() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+0); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

// HitboxComponentReader provides lazy, field-level access to HitboxComponent without reading the entire struct.
type HitboxComponentReader struct {
	ctx  *runtime.ReadContext
	addr uintptr
}

// NewHitboxComponentReader creates a lazy reader for HitboxComponent at the given address.
func NewHitboxComponentReader(ctx *runtime.ReadContext, addr uintptr) *HitboxComponentReader {
	return &HitboxComponentReader{ctx: ctx, addr: addr}
}

// Addr returns the base address of this HitboxComponent.
func (r *HitboxComponentReader) Addr() uintptr {
	return r.addr
}

// Read materializes the full HitboxComponent struct eagerly.
func (r *HitboxComponentReader) Read() (*HitboxComponent, runtime.Errors) {
	return ReadHitboxComponent(r.ctx, r.addr)
}

// Header returns a lazy reader for the nested ComponentHeader (zero I/O).
func (r *HitboxComponentReader) Header() *ComponentHeaderReader {
	return NewComponentHeaderReader(r.ctx, uintptr(int64(r.addr)+0))
}

func (r *HitboxComponentReader) IsPlayer() (bool, error) {
	var buf [1]byte
	if _, err := r.ctx.ReadAt(buf[:1], int64(r.addr)+72); err != nil {
		return false, err
	}
	return buf[0] != 0, nil
}

func (r *HitboxComponentReader) IsEnemy() (bool, error) {
	var buf [1]byte
	if _, err := r.ctx.ReadAt(buf[:1], int64(r.addr)+73); err != nil {
		return false, err
	}
	return buf[0] != 0, nil
}

func (r *HitboxComponentReader) IsItem() (bool, error) {
	var buf [1]byte
	if _, err := r.ctx.ReadAt(buf[:1], int64(r.addr)+74); err != nil {
		return false, err
	}
	return buf[0] != 0, nil
}

func (r *HitboxComponentReader) AabbMinX() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+76); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *HitboxComponentReader) AabbMaxX() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+80); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *HitboxComponentReader) AabbMinY() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+84); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *HitboxComponentReader) AabbMaxY() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+88); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *HitboxComponentReader) DamageMultiplier() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+92); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *HitboxComponentReader) OffsetX() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+96); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *HitboxComponentReader) OffsetY() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+100); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

// ItemComponentReader provides lazy, field-level access to ItemComponent without reading the entire struct.
type ItemComponentReader struct {
	ctx  *runtime.ReadContext
	addr uintptr
}

// NewItemComponentReader creates a lazy reader for ItemComponent at the given address.
func NewItemComponentReader(ctx *runtime.ReadContext, addr uintptr) *ItemComponentReader {
	return &ItemComponentReader{ctx: ctx, addr: addr}
}

// Addr returns the base address of this ItemComponent.
func (r *ItemComponentReader) Addr() uintptr {
	return r.addr
}

// Read materializes the full ItemComponent struct eagerly.
func (r *ItemComponentReader) Read() (*ItemComponent, runtime.Errors) {
	return ReadItemComponent(r.ctx, r.addr)
}

// Header returns a lazy reader for the nested ComponentHeader (zero I/O).
func (r *ItemComponentReader) Header() *ComponentHeaderReader {
	return NewComponentHeaderReader(r.ctx, uintptr(int64(r.addr)+0))
}

// ItemName returns a lazy reader for the nested MsvcString (zero I/O).
func (r *ItemComponentReader) ItemName() *MsvcStringReader {
	return NewMsvcStringReader(r.ctx, uintptr(int64(r.addr)+72))
}

func (r *ItemComponentReader) IsStackable() (bool, error) {
	var buf [1]byte
	if _, err := r.ctx.ReadAt(buf[:1], int64(r.addr)+96); err != nil {
		return false, err
	}
	return buf[0] != 0, nil
}

func (r *ItemComponentReader) IsConsumable() (bool, error) {
	var buf [1]byte
	if _, err := r.ctx.ReadAt(buf[:1], int64(r.addr)+97); err != nil {
		return false, err
	}
	return buf[0] != 0, nil
}

func (r *ItemComponentReader) StatsCountAsItemPickUp() (bool, error) {
	var buf [1]byte
	if _, err := r.ctx.ReadAt(buf[:1], int64(r.addr)+98); err != nil {
		return false, err
	}
	return buf[0] != 0, nil
}

func (r *ItemComponentReader) AutoPickup() (bool, error) {
	var buf [1]byte
	if _, err := r.ctx.ReadAt(buf[:1], int64(r.addr)+99); err != nil {
		return false, err
	}
	return buf[0] != 0, nil
}

func (r *ItemComponentReader) Unknown64() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+100); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *ItemComponentReader) UsesRemaining() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+104); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *ItemComponentReader) IsIdentified() (bool, error) {
	var buf [1]byte
	if _, err := r.ctx.ReadAt(buf[:1], int64(r.addr)+108); err != nil {
		return false, err
	}
	return buf[0] != 0, nil
}

func (r *ItemComponentReader) IsFrozen() (bool, error) {
	var buf [1]byte
	if _, err := r.ctx.ReadAt(buf[:1], int64(r.addr)+109); err != nil {
		return false, err
	}
	return buf[0] != 0, nil
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

func (r *WorldManagerViewRectReader) PBackgroundGrid() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+72); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

// FollowPBackgroundGrid reads the PBackgroundGrid pointer and follows it to the target BiomeGrid.
func (r *WorldManagerViewRectReader) FollowPBackgroundGrid() (*BiomeGrid, runtime.Errors) {
	ptr, err := r.PBackgroundGrid()
	if err != nil || ptr == 0 {
		if err != nil {
			var errs runtime.Errors
			errs.Add("WorldManagerViewRect.PBackgroundGrid", r.addr, err)
			return nil, errs
		}
		return nil, nil
	}
	return ReadBiomeGrid(r.ctx, uintptr(ptr))
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

func (r *CellDataReader) FallbackColor() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+100); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *CellDataReader) TexturePtr() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+136); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

// FollowTexturePtr reads the TexturePtr pointer and follows it to the target CellTexture.
func (r *CellDataReader) FollowTexturePtr() (*CellTexture, runtime.Errors) {
	ptr, err := r.TexturePtr()
	if err != nil || ptr == 0 {
		if err != nil {
			var errs runtime.Errors
			errs.Add("CellData.TexturePtr", r.addr, err)
			return nil, errs
		}
		return nil, nil
	}
	return ReadCellTexture(r.ctx, uintptr(ptr))
}

// CellGridReader provides lazy, field-level access to CellGrid without reading the entire struct.
type CellGridReader struct {
	ctx  *runtime.ReadContext
	addr uintptr
}

// NewCellGridReader creates a lazy reader for CellGrid at the given address.
func NewCellGridReader(ctx *runtime.ReadContext, addr uintptr) *CellGridReader {
	return &CellGridReader{ctx: ctx, addr: addr}
}

// Addr returns the base address of this CellGrid.
func (r *CellGridReader) Addr() uintptr {
	return r.addr
}

// Read materializes the full CellGrid struct eagerly.
func (r *CellGridReader) Read() (*CellGrid, runtime.Errors) {
	return ReadCellGrid(r.ctx, r.addr)
}

func (r *CellGridReader) Vtable() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+0); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *CellGridReader) Unknown04() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+4); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *CellGridReader) ChunkTablePtr() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+8); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

// CellMaterialInfoReader provides lazy, field-level access to CellMaterialInfo without reading the entire struct.
type CellMaterialInfoReader struct {
	ctx  *runtime.ReadContext
	addr uintptr
}

// NewCellMaterialInfoReader creates a lazy reader for CellMaterialInfo at the given address.
func NewCellMaterialInfoReader(ctx *runtime.ReadContext, addr uintptr) *CellMaterialInfoReader {
	return &CellMaterialInfoReader{ctx: ctx, addr: addr}
}

// Addr returns the base address of this CellMaterialInfo.
func (r *CellMaterialInfoReader) Addr() uintptr {
	return r.addr
}

// Read materializes the full CellMaterialInfo struct eagerly.
func (r *CellMaterialInfoReader) Read() (*CellMaterialInfo, runtime.Errors) {
	return ReadCellMaterialInfo(r.ctx, r.addr)
}

func (r *CellMaterialInfoReader) MaterialId() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+48); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
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

func (r *CharacterDataComponentReader) FlyTimeMax() (float32, error) {
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

// VelocityComponentReader provides lazy, field-level access to VelocityComponent without reading the entire struct.
type VelocityComponentReader struct {
	ctx  *runtime.ReadContext
	addr uintptr
}

// NewVelocityComponentReader creates a lazy reader for VelocityComponent at the given address.
func NewVelocityComponentReader(ctx *runtime.ReadContext, addr uintptr) *VelocityComponentReader {
	return &VelocityComponentReader{ctx: ctx, addr: addr}
}

// Addr returns the base address of this VelocityComponent.
func (r *VelocityComponentReader) Addr() uintptr {
	return r.addr
}

// Read materializes the full VelocityComponent struct eagerly.
func (r *VelocityComponentReader) Read() (*VelocityComponent, runtime.Errors) {
	return ReadVelocityComponent(r.ctx, r.addr)
}

// Header returns a lazy reader for the nested ComponentHeader (zero I/O).
func (r *VelocityComponentReader) Header() *ComponentHeaderReader {
	return NewComponentHeaderReader(r.ctx, uintptr(int64(r.addr)+0))
}

func (r *VelocityComponentReader) GravityX() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+72); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *VelocityComponentReader) GravityY() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+76); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *VelocityComponentReader) Mass() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+80); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *VelocityComponentReader) AirFriction() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+84); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *VelocityComponentReader) TerminalVelocity() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+88); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
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

// FollowPWorldManager reads the PWorldManager pointer and follows it to the target WorldManagerViewRect.
func (r *GameGlobalsReader) FollowPWorldManager() (*WorldManagerViewRect, runtime.Errors) {
	ptr, err := r.PWorldManager()
	if err != nil || ptr == 0 {
		if err != nil {
			var errs runtime.Errors
			errs.Add("GameGlobals.PWorldManager", r.addr, err)
			return nil, errs
		}
		return nil, nil
	}
	return ReadWorldManagerViewRect(r.ctx, uintptr(ptr))
}

func (r *GameGlobalsReader) PChunkSystem() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+16); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

// FollowPChunkSystem reads the PChunkSystem pointer and follows it to the target ChunkSystem.
func (r *GameGlobalsReader) FollowPChunkSystem() (*ChunkSystem, runtime.Errors) {
	ptr, err := r.PChunkSystem()
	if err != nil || ptr == 0 {
		if err != nil {
			var errs runtime.Errors
			errs.Add("GameGlobals.PChunkSystem", r.addr, err)
			return nil, errs
		}
		return nil, nil
	}
	return ReadChunkSystem(r.ctx, uintptr(ptr))
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

// FollowPCellFactory reads the PCellFactory pointer and follows it to the target CellFactory.
func (r *GameGlobalsReader) FollowPCellFactory() (*CellFactory, runtime.Errors) {
	ptr, err := r.PCellFactory()
	if err != nil || ptr == 0 {
		if err != nil {
			var errs runtime.Errors
			errs.Add("GameGlobals.PCellFactory", r.addr, err)
			return nil, errs
		}
		return nil, nil
	}
	return ReadCellFactory(r.ctx, uintptr(ptr))
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

// ChunkSystemReader provides lazy, field-level access to ChunkSystem without reading the entire struct.
type ChunkSystemReader struct {
	ctx  *runtime.ReadContext
	addr uintptr
}

// NewChunkSystemReader creates a lazy reader for ChunkSystem at the given address.
func NewChunkSystemReader(ctx *runtime.ReadContext, addr uintptr) *ChunkSystemReader {
	return &ChunkSystemReader{ctx: ctx, addr: addr}
}

// Addr returns the base address of this ChunkSystem.
func (r *ChunkSystemReader) Addr() uintptr {
	return r.addr
}

// Read materializes the full ChunkSystem struct eagerly.
func (r *ChunkSystemReader) Read() (*ChunkSystem, runtime.Errors) {
	return ReadChunkSystem(r.ctx, r.addr)
}

func (r *ChunkSystemReader) Vtable() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+0); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

// CellGrid returns a lazy reader for the nested CellGrid (zero I/O).
func (r *ChunkSystemReader) CellGrid() *CellGridReader {
	return NewCellGridReader(r.ctx, uintptr(int64(r.addr)+1280))
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

// FollowChildrenPtr reads the ChildrenPtr pointer and follows it to the target ChildrenContainer.
func (r *EntityReader) FollowChildrenPtr() (*ChildrenContainer, runtime.Errors) {
	ptr, err := r.ChildrenPtr()
	if err != nil || ptr == 0 {
		if err != nil {
			var errs runtime.Errors
			errs.Add("Entity.ChildrenPtr", r.addr, err)
			return nil, errs
		}
		return nil, nil
	}
	return ReadChildrenContainer(r.ctx, uintptr(ptr))
}

func (r *EntityReader) ParentEntityPtr() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+148); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

// FollowParentEntityPtr reads the ParentEntityPtr pointer and follows it to the target Entity.
func (r *EntityReader) FollowParentEntityPtr() (*Entity, runtime.Errors) {
	ptr, err := r.ParentEntityPtr()
	if err != nil || ptr == 0 {
		if err != nil {
			var errs runtime.Errors
			errs.Add("Entity.ParentEntityPtr", r.addr, err)
			return nil, errs
		}
		return nil, nil
	}
	return ReadEntity(r.ctx, uintptr(ptr))
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

// CollisionTriggerComponentReader provides lazy, field-level access to CollisionTriggerComponent without reading the entire struct.
type CollisionTriggerComponentReader struct {
	ctx  *runtime.ReadContext
	addr uintptr
}

// NewCollisionTriggerComponentReader creates a lazy reader for CollisionTriggerComponent at the given address.
func NewCollisionTriggerComponentReader(ctx *runtime.ReadContext, addr uintptr) *CollisionTriggerComponentReader {
	return &CollisionTriggerComponentReader{ctx: ctx, addr: addr}
}

// Addr returns the base address of this CollisionTriggerComponent.
func (r *CollisionTriggerComponentReader) Addr() uintptr {
	return r.addr
}

// Read materializes the full CollisionTriggerComponent struct eagerly.
func (r *CollisionTriggerComponentReader) Read() (*CollisionTriggerComponent, runtime.Errors) {
	return ReadCollisionTriggerComponent(r.ctx, r.addr)
}

// Header returns a lazy reader for the nested ComponentHeader (zero I/O).
func (r *CollisionTriggerComponentReader) Header() *ComponentHeaderReader {
	return NewComponentHeaderReader(r.ctx, uintptr(int64(r.addr)+0))
}

func (r *CollisionTriggerComponentReader) Width() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+72); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *CollisionTriggerComponentReader) Height() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+76); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *CollisionTriggerComponentReader) Radius() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+80); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

// RequiredTag returns a lazy reader for the nested MsvcString (zero I/O).
func (r *CollisionTriggerComponentReader) RequiredTag() *MsvcStringReader {
	return NewMsvcStringReader(r.ctx, uintptr(int64(r.addr)+84))
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

// PixelSceneEntryReader provides lazy, field-level access to PixelSceneEntry without reading the entire struct.
type PixelSceneEntryReader struct {
	ctx  *runtime.ReadContext
	addr uintptr
}

// NewPixelSceneEntryReader creates a lazy reader for PixelSceneEntry at the given address.
func NewPixelSceneEntryReader(ctx *runtime.ReadContext, addr uintptr) *PixelSceneEntryReader {
	return &PixelSceneEntryReader{ctx: ctx, addr: addr}
}

// Addr returns the base address of this PixelSceneEntry.
func (r *PixelSceneEntryReader) Addr() uintptr {
	return r.addr
}

// Read materializes the full PixelSceneEntry struct eagerly.
func (r *PixelSceneEntryReader) Read() (*PixelSceneEntry, runtime.Errors) {
	return ReadPixelSceneEntry(r.ctx, r.addr)
}

func (r *PixelSceneEntryReader) ChunkSystemBackRef() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+0); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *PixelSceneEntryReader) X() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+4); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *PixelSceneEntryReader) Y() (int32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+8); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:4])), nil
}

// MaterialsFilename returns a lazy reader for the nested MsvcString (zero I/O).
func (r *PixelSceneEntryReader) MaterialsFilename() *MsvcStringReader {
	return NewMsvcStringReader(r.ctx, uintptr(int64(r.addr)+12))
}

// ColorsFilename returns a lazy reader for the nested MsvcString (zero I/O).
func (r *PixelSceneEntryReader) ColorsFilename() *MsvcStringReader {
	return NewMsvcStringReader(r.ctx, uintptr(int64(r.addr)+36))
}

// BackgroundFilename returns a lazy reader for the nested MsvcString (zero I/O).
func (r *PixelSceneEntryReader) BackgroundFilename() *MsvcStringReader {
	return NewMsvcStringReader(r.ctx, uintptr(int64(r.addr)+60))
}

func (r *PixelSceneEntryReader) FlagSkipBiomeChecks() (uint8, error) {
	var buf [1]byte
	if _, err := r.ctx.ReadAt(buf[:1], int64(r.addr)+88); err != nil {
		return 0, err
	}
	return buf[0], nil
}

func (r *PixelSceneEntryReader) FlagSkipEdgeTextures() (uint8, error) {
	var buf [1]byte
	if _, err := r.ctx.ReadAt(buf[:1], int64(r.addr)+89); err != nil {
		return 0, err
	}
	return buf[0], nil
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

func (r *DamageModelComponentReader) DamageMultipliersVtable() (uint32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+104); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:4]), nil
}

func (r *DamageModelComponentReader) DamageMultipliersMelee() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+108); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *DamageModelComponentReader) DamageMultipliersProjectile() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+112); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *DamageModelComponentReader) DamageMultipliersExplosion() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+116); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *DamageModelComponentReader) DamageMultipliersElectricity() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+120); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *DamageModelComponentReader) DamageMultipliersFire() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+124); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *DamageModelComponentReader) DamageMultipliersDrill() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+128); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *DamageModelComponentReader) DamageMultipliersSlice() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+132); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *DamageModelComponentReader) DamageMultipliersIce() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+136); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *DamageModelComponentReader) DamageMultipliersHealing() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+140); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *DamageModelComponentReader) DamageMultipliersPhysicsHit() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+144); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *DamageModelComponentReader) DamageMultipliersRadioactive() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+148); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *DamageModelComponentReader) DamageMultipliersPoison() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+152); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *DamageModelComponentReader) DamageMultipliersOvereating() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+156); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *DamageModelComponentReader) DamageMultipliersCurse() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+160); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *DamageModelComponentReader) DamageMultipliersHoly() (float32, error) {
	var buf [4]byte
	if _, err := r.ctx.ReadAt(buf[:4], int64(r.addr)+164); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(buf[:4])), nil
}

func (r *DamageModelComponentReader) CriticalDamageResistance() (float32, error) {
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

// Ensure imports are used.
var (
	_ = binary.LittleEndian
	_ = json.Marshal
	_ = fmt.Sprintf
	_ = math.Float32frombits
)
