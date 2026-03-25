// Noita Memory Structures — ImHex Pattern Language
// For use with memory dumps from noita.exe (PE x86 32-bit LE, image base 0x00400000)
//
// Usage: Open a memory dump in ImHex, set base address to match the dump region,
//        then place patterns at the appropriate addresses.
//
// All offsets validated against live memory dumps (2026-03-18).

import std.mem;
import std.string;

// =============================================================================
// MSVC Types
// =============================================================================

// MSVC std::string with Small String Optimization (SSO)
// Total size: 24 bytes (NOT 28 — there is no trailing padding)
// If capacity <= 15: data inline in first 16 bytes
// If capacity > 15:  first 4 bytes = heap pointer, data on heap
struct MsvcString {
    char data[16]   [[comment("Inline if capacity<=15, else first 4 bytes = heap ptr")]];
    u32 length;
    u32 capacity;
} [[static, format("format_msvc_string")]];

fn format_msvc_string(MsvcString s) {
    if (s.length == 0) return "";
    if (s.capacity <= 15) {
        str result = "";
        for (u32 i = 0, i < s.length, i = i + 1) {
            result = result + s.data[i];
        }
        return result;
    }
    u32 heap_ptr = s.data[0] | (s.data[1] << 8) | (s.data[2] << 16) | (s.data[3] << 24);
    if (heap_ptr == 0) return "";
    return std::mem::read_string(heap_ptr, s.length);
};

// MSVC std::vector<T> header (12 bytes) — opaque, for vectors we don't index into
struct StdVectorHeader {
    u32 begin_ptr;
    u32 end_ptr;
    u32 capacity_ptr;
} [[sealed, static, format("format_vector")]];

fn format_vector(StdVectorHeader v) {
    if (v.begin_ptr == 0) return "empty";
    u32 count = (v.end_ptr - v.begin_ptr) / 4;
    return std::format("{} elements @ 0x{:08X}", count, v.begin_ptr);
};

// Typed vectors — element data read from heap at begin_ptr
struct U32Vector {
    u32 begin_ptr;
    u32 end_ptr;
    u32 capacity_ptr;
    u32 elements[(end_ptr - begin_ptr) / 4] @ begin_ptr;
} [[static]];

struct S32Vector {
    u32 begin_ptr;
    u32 end_ptr;
    u32 capacity_ptr;
    s32 elements[(end_ptr - begin_ptr) / 4] @ begin_ptr;
} [[static]];

struct F64Vector {
    u32 begin_ptr;
    u32 end_ptr;
    u32 capacity_ptr;
    double elements[(end_ptr - begin_ptr) / 8] @ begin_ptr;
} [[static]];

// =============================================================================
// Forward declarations for pointer types
// =============================================================================

using Entity;
using ChildrenContainer;

// =============================================================================
// Entity System
// =============================================================================

// Entity struct (0x98 bytes)
// Position is stored directly here, NOT in a component.
struct Entity {
    s32 entity_id           [[comment("Unique ID, 1-based")]];
    s32 slot_index          [[comment("Index in EntityManager.entityArray")]];
    u32 unknown_08;
    s32 pending_kill        [[comment("<1 = alive"), format("format_alive")]];
    u32 flags_10            [[comment("Low byte may be active flag, upper bytes vary")]];
    MsvcString name         [[comment("e.g. 'DEBUG_NAME:player'")]];
    u32 unknown_2c          [[comment("Gap between name and tag_bitset")]];

    // +0x30: Tag bitset (64 bytes = 512 bits)
    u8 tag_bitset[64]       [[comment("Bitset512 for entity tags")]];

    // +0x70: Transform block
    float pos_x             [[comment("World X position"), color("FF4444")]];
    float pos_y             [[comment("World Y position"), color("FF4444")]];
    float rot_cos           [[comment("cos(rotation)")]];
    float rot_sin           [[comment("sin(rotation)")]];
    float rot_neg_sin       [[comment("-sin(rotation)")]];
    float rot_cos2          [[comment("cos(rotation) duplicate")]];
    float scale_x;
    float scale_y;

    // +0x90
    ChildrenContainer *children_ptr : u32 [[comment("ChildrenContainer* (Entity** array), or NULL")]];
    Entity *parent_entity_ptr : u32      [[comment("Entity* parent, or NULL")]];
} [[color("4488FF"), static]];

fn format_alive(s32 v) {
    if (v < 1) return "alive";
    return "dead/pending_kill";
};

// ChildrenContainer (0x0C bytes, heap-allocated at entity+0x90)
// Stores Entity* pointers (NOT entity IDs)
struct ChildrenContainer {
    u32 begin_ptr       [[comment("Entity** — first child pointer")]];
    u32 end_ptr         [[comment("Entity** — past last child")]];
    u32 capacity_ptr;
    u32 children[(end_ptr - begin_ptr) / 4] @ begin_ptr;
} [[static, format("format_children")]];

fn format_children(ChildrenContainer c) {
    if (c.begin_ptr == 0) return "no children";
    u32 count = (c.end_ptr - c.begin_ptr) / 4;
    return std::format("{} children", count);
};

// EntityManager (0x3C bytes)
struct EntityManager {
    u32 vtable;
    s32 next_entity_id          [[comment("Auto-incrementing, 1-based")]];

    // Free slot stack (std::vector<int>)
    StdVectorHeader free_slot_stack [[name("freeSlotStack")]];

    // Entity array (std::vector<Entity*>) — can be 100k+ entries, read lazily
    StdVectorHeader entity_array    [[name("entityArray"), color("44FF44")]];

    // Tag groups (12-byte stride entries)
    StdVectorHeader tag_groups      [[name("tagGroups"), comment("NOT for component lookup")]];

    // Component buffers (std::vector<ComponentBuffer*>) — ~162 entries
    U32Vector component_buffers     [[name("componentBuffers"), color("FF8844"),
        comment("Indexed by runtime typeId. Use component+0x08 to discover type names.")]];

    u32 p_event_manager             [[name("pEventManager")]];
} [[color("44FF44")]];

// =============================================================================
// Component Buffer (sparse-dense indexed array)
// =============================================================================

// ComponentBuffer<T> — per-type component storage (~0xC4 bytes)
// Lookup: sparseIndex[entity.slotIndex] → denseIndex → components[denseIndex]
struct ComponentBuffer {
    u32 vtable;
    s32 sentinel                [[comment("Always -1 (not-found marker)")]];
    s32 initial_capacity        [[comment("Default 0x40")]];
    u32 unknown_0c;

    // Sparse array: indexed by entity.slotIndex → dense index (or -1)
    // These are too large to read eagerly (100k+ entries) — use StdVectorHeader
    // and index manually via readU32/readS32.
    StdVectorHeader sparse_index    [[name("sparseIndex"), color("FFAA44"),
        comment("sparseIndex[entity.slotIndex] = denseIndex or -1")]];

    // Dense arrays (parallel)
    StdVectorHeader entity_refs     [[name("entityRefs"), comment("Entity* per dense slot")]];
    StdVectorHeader prev_index      [[name("prevIndex"), comment("Doubly-linked chain")]];
    StdVectorHeader next_index      [[name("nextIndex"),
        comment("Follow chain for multiple components of same type on one entity")]];
    StdVectorHeader components      [[name("components"), color("44FFAA"),
        comment("Component* per dense slot")]];

    // Free slot deque (0x14 bytes)
    padding[0x14]               [[name("freeSlotDeque")]];

    // Handle/generation system
    StdVectorHeader handle_map;
    StdVectorHeader generations;
    StdVectorHeader reverse_handle_map;

    padding[0x14]               [[name("freeHandleDeque")]];

    s32 active_count            [[color("FFFF44")]];
    s32 capacity_limit;
    u32 unknown_a0;
    u32 p_entity_manager;
    u32 p_event_manager;

    MsvcString name_string      [[comment("Component type name (may be empty — use component+0x08 instead)")]];
} [[color("FF8844")]];

// =============================================================================
// Component Base Header (common to ALL component types, 0x48 bytes)
// =============================================================================

struct ComponentHeader {
    u32 vtable;
    s32 buffer_index            [[comment("Dense index in ComponentBuffer")]];
    u32 p_type_name             [[comment("char* to type name C string")]];
    s32 type_id                 [[comment("Runtime-assigned by ComponentTypeRegistry")]];
    u32 unknown_10;
    bool active                 [[color("44FF44"), comment("Must be nonzero to use")]];
    padding[3];
    u8 component_tags[32]       [[comment("Bitset256")]];
    u32 unknown_38;
    u32 unknown_3c;
    u32 unknown_40;
    u32 unknown_44;
    // Component-specific fields start at +0x48
};

// =============================================================================
// DamageModelComponent (0x2DC bytes, typeId=31)
// =============================================================================

struct DamageModelComponent {
    ComponentHeader header      [[inline]];

    // +0x48
    double hp                   [[color("FF4444"), comment("Multiply by 25 for display value")]];
    double max_hp               [[color("FF4444")]];
    double max_hp_cap;
    double max_hp_old;

    // +0x68: unknown field (constant 0x0100FD84 across runs, NOT a multiplier)
    u32 unknown_68              [[comment("Constant across runs — possibly a pointer or type info")]];

    // +0x6C: 16 damage multipliers (validated: 0.35=electricity, 1.5=material in dumps)
    float dmg_mult_melee        [[color("FF8844")]];
    float dmg_mult_projectile;
    float dmg_mult_explosion;
    float dmg_mult_electricity;
    float dmg_mult_fire;
    float dmg_mult_drill;
    float dmg_mult_slice;
    float dmg_mult_ice;
    float dmg_mult_healing;
    float dmg_mult_physics_hit;
    float dmg_mult_radioactive;
    float dmg_mult_poison;
    float dmg_mult_holy;
    float dmg_mult_curse;
    float dmg_mult_overeating;
    float dmg_mult_material;

    // +0xAC
    s32 invincibility_frames    [[comment("Frames of i-frames remaining")]];

    // Remaining 0x2DC - 0xB0 = 0x22C bytes of fields
    // (blood, fire, fall damage, ragdoll, etc.)
    padding[0x22C];
} [[color("FF4444")]];

// =============================================================================
// WalletComponent (0x64 bytes, typeId=159)
// =============================================================================

struct WalletComponent {
    ComponentHeader header      [[inline]];

    // +0x48
    s64 money                   [[color("FFD700"), comment("Current gold")]];
    s64 money_spent;
    s64 money_prev_frame        [[comment("Previous frame gold (infinite gold HAX detection)")]];
    bool has_reached_inf        [[comment("Gold overflow flag")]];
    padding[3];
} [[color("FFD700")]];

// =============================================================================
// CharacterDataComponent (0x118 bytes, typeId=22)
// =============================================================================

struct CharacterDataComponent {
    ComponentHeader header      [[inline]];

    // +0x48 to +0x87: platforming, AABB fields
    padding[0x40];

    // +0x88
    float gravity;
    float fly_time;

    // +0x90 to +0xB7
    padding[0x28];

    // +0xB8
    bool is_on_ground;
    padding[0x4F];

    // +0x108
    float velocity_x            [[color("44AAFF")]];
    float velocity_y            [[color("44AAFF")]];

    padding[0x08];
} [[color("44AAFF")]];

// =============================================================================
// AbilityComponent (0x3BC bytes, typeId=3) — Wand data
// =============================================================================

// ConfigGun sub-struct (20 bytes, at AbilityComponent+0x110)
struct ConfigGun {
    u32 vtable;
    s32 actions_per_round       [[comment("Spells per click")]];
    bool shuffle_deck_when_empty [[comment("Shuffle vs non-shuffle wand")]];
    padding[3];
    s32 reload_time             [[comment("Wand reload (frames), default 40")]];
    s32 deck_capacity           [[comment("Max spell slots")]];
} [[color("AA44FF")]];

struct AbilityComponent {
    ComponentHeader header      [[inline]];

    // +0x48
    s32 cooldown_frames;
    MsvcString entity_file      [[comment("Projectile entity XML")]];
    MsvcString sprite_file;
    s32 entity_count;
    bool never_reload;
    padding[3];
    s32 reload_time_frames      [[comment("Wand base reload")]];
    float mana                  [[color("4444FF"), comment("Current mana")]];
    float mana_max              [[color("4444FF")]];
    float mana_charge_speed     [[color("4444FF"), comment("Mana regen per frame")]];

    // +0x94
    bool rotate_in_hand;
    padding[3];
    float rotate_in_hand_amount;
    float rotate_hand_amount;
    bool fast_projectile;
    padding[3];
    float swim_propel_amount;
    s32 max_charged_actions;
    s32 charge_wait_frames;
    float item_recoil_recovery_speed;
    float item_recoil_max;
    float item_recoil_offset_coeff;
    float item_recoil_rotation_coeff;
    MsvcString base_item_file;
    bool use_entity_file_as_projectile_info_proxy;
    bool click_to_use;
    padding[2];
    s32 stat_times_player_has_shot;
    s32 stat_times_player_has_edited;
    bool shooting_reduces_amount_in_inventory;
    bool throw_as_item;
    bool simulate_throw_as_item;
    padding[1];
    s32 max_amount_in_inventory;
    s32 amount_in_inventory;
    bool drop_as_item_on_death;
    padding[3];
    MsvcString ui_name;

    // +0x10C
    bool use_gun_script;
    bool is_petris_gun;
    padding[2];

    // +0x110
    ConfigGun gun_config        [[color("AA44FF")]];

    // +0x124: ConfigGunActionInfo (572 bytes) — per-shot spell accumulator
    // Too large to fully inline; key fields below
    padding[0x23C]              [[name("gunaction_config"),
        comment("ConfigGunActionInfo: 572 bytes. fire_rate_wait at +0x100, damage fields at +0x124-0x14C")]];

    // +0x360
    s32 gun_level;
    MsvcString add_these_child_actions;
    s32 current_slot_durability;
    MsvcString slot_consumption_function;

    // +0x398: Runtime state
    s32 next_frame_usable       [[comment("Frame when cast delay ends")]];
    s32 cast_delay_start_frame;
    s32 ammo_left;
    s32 reload_frames_left      [[comment("Frames remaining in reload")]];
    s32 reload_next_frame_usable;
    s32 charge_count;
    s32 next_charge_frame;
    float item_recoil;
    bool is_initialized;
    padding[3];
} [[color("AA44FF")]];

// =============================================================================
// HitboxComponent (0x68 bytes, typeId=63) — AABB collision hitbox
// =============================================================================

struct HitboxComponent {
    ComponentHeader header      [[inline]];

    // +0x48
    bool is_player              [[comment("Is a player hitbox")]];
    bool is_enemy               [[comment("Is an enemy hitbox (default: true)")]];
    bool is_item                [[comment("Is an item hitbox")]];
    padding[1];                 // alignment

    // +0x4C
    float aabb_min_x            [[comment("Bounding box min X relative to entity (default: -5)")]];
    float aabb_max_x            [[comment("Bounding box max X relative to entity (default: 5)")]];
    float aabb_min_y            [[comment("Bounding box min Y relative to entity (default: -5)")]];
    float aabb_max_y            [[comment("Bounding box max Y relative to entity (default: 5)")]];
    float damage_multiplier     [[comment("Damage multiplier for hits to this hitbox (default: 1.0)")]];
    float offset_x              [[comment("Hitbox offset X from entity center")]];
    float offset_y              [[comment("Hitbox offset Y from entity center")]];
} [[color("FF8844")]];

// =============================================================================
// CollisionTriggerComponent (typeId=25) — trigger zone with width/height/radius
// =============================================================================

struct CollisionTriggerComponent {
    ComponentHeader header      [[inline]];

    // +0x48
    float width                 [[comment("Trigger zone width (default: 32)")]];
    float height                [[comment("Trigger zone height (default: 32)")]];
    float radius                [[comment("Trigger zone radius (default: 32)")]];
    MsvcString required_tag     [[comment("Tag required to trigger (default: mortal)")]];
} [[color("6272A4")]];

// =============================================================================
// SpriteComponent (0xF0 bytes, typeId=141) — visual sprite
// =============================================================================

struct SpriteComponent {
    ComponentHeader header      [[inline]];

    // +0x48
    MsvcString image_file       [[comment("Path to sprite image file")]];
    // remaining fields not needed
} [[color("AAAAAA")]];

// =============================================================================
// ItemComponent (typeId=80) — item in world or inventory
// =============================================================================

struct ItemComponent {
    ComponentHeader header      [[inline]];

    // +0x48
    MsvcString item_name        [[comment("Name of the item (e.g. $item_potion)")]];
    // +0x60
    bool is_stackable;
    bool is_consumable;
    bool stats_count_as_item_pick_up;
    bool auto_pickup;
    u32 unknown_64              [[hidden]];
    // +0x68
    s32 uses_remaining          [[comment("Uses left; -1=unlimited")]];
    bool is_identified;
    bool is_frozen;
} [[color("F1FA8C")]];

// =============================================================================
// VelocityComponent (typeId=155) — physics velocity and gravity
// =============================================================================

struct VelocityComponent {
    ComponentHeader header      [[inline]];

    // +0x48
    float gravity_x             [[comment("Horizontal gravity (default: 0)")]];
    float gravity_y             [[comment("Vertical gravity (default: 400)")]];
    float mass                  [[comment("Physics mass (default: 0.05)")]];
    float air_friction          [[comment("Air resistance 0-1 (default: 0.55)")]];
    float terminal_velocity     [[comment("Max velocity (default: 1000)")]];
} [[color("8BE9FD")]];

// =============================================================================
// LightComponent (typeId=89) — light source
// =============================================================================

struct LightComponent {
    ComponentHeader header      [[inline]];

    // +0x48
    u32 internal_ptr             [[hidden]];
    // +0x4C
    float radius                [[comment("Light radius in pixels (default: 0)")]];
    u32 r                       [[color("FF0000"), comment("Red 0-255 (default: 255)")]];
    u32 g                       [[color("00FF00"), comment("Green 0-255 (default: 178)")]];
    u32 b                       [[color("0000FF"), comment("Blue 0-255 (default: 118)")]];
    float offset_x;
    float offset_y;
} [[color("FFD700")]];

// =============================================================================
// GameEffectComponent (typeId=53) — status effect on entity
// =============================================================================

struct GameEffectComponent {
    ComponentHeader header      [[inline]];

    // +0x48
    s32 effect                  [[comment("GAME_EFFECT enum value")]];
    s32 frames                  [[comment("Duration in frames; -1=forever")]];
} [[color("BD93F9")]];

// =============================================================================
// Inventory2Component (~0x80 bytes, typeId=73) — on player entity
// =============================================================================

struct Inventory2Component {
    ComponentHeader header      [[inline]];

    // +0x48
    s32 quick_inventory_slots   [[comment("Number of wand slots (default 4)")]];
    s32 full_inventory_slots_x;
    s32 full_inventory_slots_y;
    s32 saved_active_item_index;
    s32 active_item             [[color("FFAA44"), comment("Entity ID of held wand/item")]];
    s32 actual_active_item;
    s32 active_stash;
    s32 throw_item;
    bool item_holstered;
    bool initialized;
    bool force_refresh;
    bool dont_log_next_item_equip;
    float smoothed_item_x_offset;
    s32 last_item_switch_frame;
    float intro_equip_item_lerp;
    float smoothed_item_angle_x;
    float smoothed_item_angle_y;
} [[color("FFAA44")]];

// =============================================================================
// WorldStateComponent (0x1C8 bytes, typeId=161)
// =============================================================================

struct WorldStateComponent {
    ComponentHeader header      [[inline]];

    // +0x48 to +0x107: time, weather, day_count fields
    padding[0xC0];

    // +0x108: Gods counters
    s32 biome_crypt_count       [[name("BIOME_CRYPT")]];
    s32 gods_afraid             [[color("FF44FF"), comment("Polymorph count → angry/very angry")]];
    s32 gods_impressed          [[color("FF44FF"), comment("Kill counter")]];
    s32 gods_afraid_damage;
    s32 gods_enraged            [[color("FF44FF"), comment("Death counter → curious")]];

    // +0x11C to +0x1C7
    padding[0xAC];
} [[color("FF44FF")]];

// =============================================================================
// GameGlobals (0x1A0 bytes)
// =============================================================================

struct GameGlobals {
    s32 frame_count             [[color("FFFFFF")]];
    s32 physics_step_count      [[comment("Physics ticks (~6x frame_count at 360 physics fps)")]];
    float game_time             [[comment("Game time in seconds")]];

    // +0x0C: Manager pointers
    WorldManagerViewRect *p_world_manager : u32 [[color("88FFAA"), comment("WorldManager*, 726 xrefs")]];
    u32 p_chunk_system          [[color("88FFAA")]];
    u32 p_cell_grid             [[color("88FFAA")]];
    CellFactory *p_cell_factory : u32 [[color("88FFAA")]];
    u32 unknown_1c;
    u32 p_physics_world         [[color("88FFAA")]];
    u32 p_audio_manager         [[color("88FFAA")]];

    // +0x28 to +0x17F: internal state (camera fields at +0x130/+0x158 are NOT
    // reliable — use WorldManager view rect for camera position instead)
    padding[0x158];

    // +0x180: Viewport bounds (world coords, validated in dumps)
    float viewport_left;
    float viewport_top;
    float viewport_right;
    float viewport_bottom;

    // +0x190 to +0x19F
    padding[0x10];
} [[color("AAAAAA")]];

// =============================================================================
// WorldManager (view rect at start)
// =============================================================================

struct WorldManagerViewRect {
    float view_x                [[comment("View rect left edge (world coords)")]];
    float view_y                [[comment("View rect top edge")]];
    float view_width            [[comment("View rect width (world pixels)")]];
    float view_height           [[comment("View rect height")]];
};

// =============================================================================
// MaterialInventoryComponent (typeId=101)
// =============================================================================

struct MaterialInventoryComponent {
    ComponentHeader header      [[inline]];

    // +0x48 to +0x7F: unknown fields
    padding[0x38];

    // +0x80: std::vector<double> — material amounts indexed by material ID
    F64Vector count_per_material_type [[name("countPerMaterialType"),
        comment("std::vector<double>: amount[material_id]")]];

    // +0x8C to end: remaining fields
    padding[0x34];
};

// =============================================================================
// CellFactory / CellData — material name lookup
// =============================================================================

// CellData (0x290 bytes stride) — one per material type in CellFactory array
struct CellData {
    MsvcString name             [[comment("Material name string")]];
    padding[0x278]              [[comment("Remaining CellData fields")]];
} [[static]];

// CellFactory — only the fields we need for material name lookups
struct CellFactory {
    // +0x00 to +0x17: unknown
    padding[0x18];

    // +0x18: pointer to CellData array base
    u32 cell_data_array_ptr     [[comment("CellData* — base of material array, stride 0x290")]];
};

// =============================================================================
// DeathMatchApp (0xD0 bytes)
// =============================================================================

struct DeathMatchApp {
    padding[0x58];

    // +0x58: Player entity vector (std::vector<Entity*>)
    U32Vector player_entities       [[color("44FF44"),
        comment("Entity* pointers (NOT entity IDs). Single-player has 1 entry.")]];

    // +0x64 to +0xCF
    padding[0x6C];
} [[color("44FF44")]];

// =============================================================================
// Static Globals
// =============================================================================

// Component type IDs (runtime-assigned, observed via PTypeName from live process).
// Only IDs confirmed at runtime are included; unobserved types are omitted.
// Named "TypeID" so hexpatgen produces Go constants like TypeIDAbilityComponent.
enum TypeID : s32 {
    AbilityComponent                    =   3,
    AnimalAIComponent                   =   6,
    AudioComponent                      =  10,
    AudioListenerComponent              =  11,
    AudioLoopComponent                  =  12,
    BiomeTrackerComponent               =  13,
    CameraBoundComponent                =  18,
    CharacterCollisionComponent         =  21,
    CharacterDataComponent              =  22,
    CharacterPlatformingComponent       =  23,
    CollisionTriggerComponent           =  25,
    ControlsComponent                   =  28,
    DamageModelComponent                =  31,
    DrugEffectComponent                 =  38,
    ElectricityReceiverComponent        =  42,
    ExplodeOnDamageComponent            =  46,
    GameEffectComponent                 =  53,
    GameLogComponent                    =  54,
    GameStatsComponent                  =  55,
    GenomeDataComponent                 =  57,
    GunComponent                        =  60,
    HitboxComponent                     =  63,
    HotspotComponent                    =  65,
    IngestionComponent                  =  70,
    InheritTransformComponent           =  71,
    InteractableComponent               =  72,
    Inventory2Component                 =  73,
    InventoryGuiComponent               =  75,
    ItemActionComponent                 =  77,
    ItemAlchemyComponent                =  78,
    ItemChestComponent                  =  79,
    ItemComponent                       =  80,
    ItemCostComponent                   =  81,
    ItemPickUpperComponent              =  82,
    KickComponent                       =  85,
    LifetimeComponent                   =  88,
    LightComponent                      =  89,
    LiquidDisplacerComponent            =  92,
    LuaComponent                        =  96,
    ManaReloaderComponent               =  99,
    MaterialAreaCheckerComponent        = 100,
    MaterialInventoryComponent          = 101,
    MaterialSuckerComponent             = 103,
    MusicEnergyAffectorComponent        = 105,
    ParticleEmitterComponent            = 109,
    PathFindingComponent                = 110,
    PathFindingGridMarkerComponent      = 111,
    PhysicsBody2Component               = 113,
    PhysicsBodyCollisionDamageComponent = 114,
    PhysicsBodyComponent                = 115,
    PhysicsImageShapeComponent          = 116,
    PhysicsJointComponent               = 119,
    PhysicsPickUpComponent              = 121,
    PhysicsThrowableComponent           = 124,
    PixelSpriteComponent                = 126,
    PlatformShooterPlayerComponent      = 127,
    PlayerCollisionComponent            = 128,
    PlayerStatsComponent                = 129,
    PositionSeedComponent               = 130,
    PotionComponent                     = 131,
    ProjectileComponent                 = 133,
    SimplePhysicsComponent              = 138,
    SpriteAnimatorComponent             = 140,
    SpriteComponent                     = 141,
    SpriteOffsetAnimatorComponent       = 142,
    SpriteParticleEmitterComponent      = 143,
    SpriteStainsComponent               = 144,
    StatusEffectDataComponent           = 145,
    StreamingKeepAliveComponent         = 146,
    TorchComponent                      = 151,
    UIInfoComponent                     = 153,
    VariableStorageComponent            = 154,
    VelocityComponent                   = 155,
    VerletPhysicsComponent              = 156,
    VerletWorldJointComponent           = 158,
    WalletComponent                     = 159,
    WorldStateComponent                 = 161
};

// =============================================================================
// Placement templates — uncomment ONE section matching your data source.
// =============================================================================

// --- Open a .sparse.bin file — all VAs work directly, no base address needed ---

// Static globals (stable across all runs)
u32 g_worldSeed              @ 0x01205004 [[color("AAFFAA"), comment("Current world seed (was incorrectly documented at 0x01206FE8)")]];
s32 g_deathCount             @ 0x01208AF8 [[color("AAFFAA"), comment("Night sky stars")]];
s32 g_numOrbsTotal           @ 0x01152544 [[color("AAFFAA"), comment("Total orb count (default 10)")]];
EntityManager *g_entityManager : u32 @ 0x01204B98 [[color("44FF44")]];
DeathMatchApp *g_deathMatchApp : u32 @ 0x01204BC0 [[color("44FF44")]];
GameGlobals *g_gameGlobals : u32     @ 0x0122374C [[color("AAAAAA")]];
WorldStateComponent *g_worldState : u32 @ 0x01205010 [[color("FF44FF")]];
u32 g_orbPersistence_ptr     @ 0x01207404 [[color("AAFFAA"), comment("-> OrbPersistenceData")]];

// Stable singletons (same heap address across observed runs)
GameGlobals   gameglobals_heap   @ 0x01BBAE58;
DeathMatchApp deathmatchapp_heap @ 0x01B93510;
EntityManager entitymanager_heap @ 0x06BA7F88;

// Per-run heap objects — uncomment the set matching your dump.
// Dump 1 (noita.exe_1773868853_24788):
Entity player_entity             @ 0x1673E6C8;
DamageModelComponent player_dmg  @ 0x3334B780;
AbilityComponent wand0           @ 0x0648EED0;
AbilityComponent wand1           @ 0x0648F2B0;
// Dump 2 (noita.exe_1773868900_25625):
// Entity player_entity           @ 0x334DF910;
// DamageModelComponent player_dmg @ 0x3352FC58;
