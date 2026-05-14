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

// MSVC std::map<K,V> header (8 bytes). Pair-compressed allocator means only
// the _Tree_val members are actually stored: a pointer to the head sentinel
// node, and the element count.
//
// Each tree node is 64 bytes:
//   +0x00 _Left   (u32)   — child or head sentinel for nil
//   +0x04 _Parent (u32)
//   +0x08 _Right  (u32)
//   +0x0C _Color  (u8)    — 0=red, 1=black
//   +0x0D _Isnil  (u8)    — head sentinel has 1, real nodes 0
//   +0x0E pad
//   +0x10 _Myval = pair<K,V>
//
// For std::map<std::string, std::string> _Myval is two MsvcStrings (24+24).
// Iterate from head->_Parent (the root); skip nodes whose _Isnil byte is set.
struct StdMapHeader {
    u32 head_ptr;
    u32 size;
} [[sealed, static]];

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

    // +0x68: damage_multipliers sub-object (ConfigDamagesByType, 64 bytes).
    // XML uses <damage_multipliers melee="..." ...>; game registers 15 fields
    // in this order. The leading u32 is the sub-object's C++ vtable, which
    // stays constant within a build (observed 0x0100FD84).
    u32 damage_multipliers_vtable        [[comment("ConfigDamagesByType vtable")]];

    // +0x6C: 15 damage multipliers (verified against player dump:
    // explosion=0.35, holy=1.5 matches player XML's damage_multipliers)
    float damage_multipliers_melee       [[color("FF8844")]];
    float damage_multipliers_projectile;
    float damage_multipliers_explosion;
    float damage_multipliers_electricity;
    float damage_multipliers_fire;
    float damage_multipliers_drill;
    float damage_multipliers_slice;
    float damage_multipliers_ice;
    float damage_multipliers_healing;
    float damage_multipliers_physics_hit;
    float damage_multipliers_radioactive;
    float damage_multipliers_poison;
    float damage_multipliers_overeating;
    float damage_multipliers_curse;
    float damage_multipliers_holy;

    // +0xA8
    float critical_damage_resistance    [[comment("0.0=full crit applied, 1.0=no crit applied")]];

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
    float fly_time_max          [[comment("XML fly_time_max; actual LensValueFloat (12B), reading only the first float")]];

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
    // +0x4C
    MsvcString custom_effect_id [[comment("Effect ID for effect=CUSTOM (e.g. \"NECROBOT\")")]];
    // +0x64
    s32 frames                  [[comment("Duration in frames; -1=forever")]];
} [[color("BD93F9")]];

// =============================================================================
// LuaComponent (typeId=96) — Lua scripting
// =============================================================================

struct LuaComponent {
    ComponentHeader header      [[inline]];

    // +0x48 — internal state before script_source_file
    padding[0xAC];
    // +0xF4
    MsvcString script_source_file [[comment("Path to .lua script file")]];
} [[color("50FA7B")]];

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

    // +0x48..+0x9B: weather, wind, lighting, lightning, player_spawn_location.
    // (Not broken out yet; would be is_initialized + 18 floats + a CVector2.)
    padding[0x54];

    // +0x9C: lua_globals — std::map<std::string,std::string> backing
    // GlobalsGetValue / GlobalsSetValue. Holds run-persistent script state:
    // fungal_shift_iteration, fungal_shift_last_frame, HOLY_MOUNTAIN_DEPTH,
    // perk_pickup_count, mods, etc.
    StdMapHeader lua_globals    [[name("luaGlobals"),
        color("50FA7B"),
        comment("std::map<string,string> from GlobalsGetValue/SetValue")]];

    // +0xA4..+0xCB: pending_portals + next_portal_id + apparitions_per_level
    // + npc_parties — left opaque.
    padding[0x28];

    // +0xCC..+0xE3: session_stat_file (std::string, 24 bytes).
    padding[0x18];

    // +0xE4: orbs_found_thisrun — std::vector<int> of orb ids picked up.
    S32Vector orbs_found_thisrun        [[name("orbsFoundThisRun"),
        color("F1FA8C"),
        comment("Orb IDs picked up this run")]];

    // +0xF0: flags — std::vector<std::string> of run-milestone flags
    // (card_unlocked_*, game_won, ending98_completed, ...).
    StdVectorHeader flags               [[name("flagsVec"),
        color("8BE9FD"),
        comment("Run flags from AddFlagPersistent / similar")]];

    // +0xFC: changed_materials — std::vector<std::string> of alternating
    // from/to material names. Each ConvertMaterialEverywhere call (e.g.
    // fungal shifts via data/scripts/magic/fungal_shift.lua) appends
    // exactly two MsvcStrings (from, then to). Element stride 24 bytes.
    StdVectorHeader changed_materials   [[name("changedMaterials"),
        color("FF8844"),
        comment("Alternating from/to material names from ConvertMaterialEverywhere (fungal shifts)")]];

    // +0x108: Five player counters that drive god reactions.
    // Names follow the schema dump (data/schemas/*.xml). The TUI currently
    // groups them as "gods afraid/impressed/enraged" — these are the raw
    // counters that feed those concepts.
    s32 player_polymorph_count          [[color("FF44FF"),
        comment("Polymorphs taken — gods afraid threshold")]];
    s32 player_polymorph_random_count   [[color("FF44FF"),
        comment("Random polymorphs — gods very afraid threshold")]];
    s32 player_did_infinite_spell_count [[color("FF44FF"),
        comment("Infinite-spell exploits triggered")]];
    s32 player_did_damage_over_1milj    [[color("FF44FF"),
        comment("Single hits dealing >1M damage — gods impressed")]];
    s32 player_living_with_minus_hp     [[color("FF44FF"),
        comment("Frames lived at <0 HP")]];

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
    ChunkSystem *p_chunk_system : u32 [[color("88FFAA"), comment("Owns embedded CellGrid at +0x500")]];
    u32 p_cell_grid             [[color("88FFAA"), comment("Different pointer than p_chunk_system->cell_grid; noita_hook uses the embedded one")]];
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
//
// pBackgroundGrid offset confirmed by LuaImpl_BiomeMapGetName @ 0x007c8f50
// (assembly: MOV ECX,[EAX + 0xc] ; MOV ECX,[ECX + 0x48]).

struct WorldManagerViewRect {
    float view_x                [[comment("View rect left edge (world coords)")]];
    float view_y                [[comment("View rect top edge")]];
    float view_width            [[comment("View rect width (world pixels)")]];
    float view_height           [[comment("View rect height")]];

    // +0x10..+0x47: many sprite/scene pointers, not mapped here.
    padding[0x38];

    // +0x48: BackgroundGrid* — the biome chunk grid (NOT the cell chunk grid).
    BiomeGrid *p_background_grid : u32 [[color("FFCC44"),
        comment("BiomeGrid* — drives BiomeMapGetName / wobbled biome lookup")]];
};

// =============================================================================
// Biome chunk grid — drives ChunkGrid_ResolveChunkAtPosition
// (noita.exe @ 0x0087d9a0, the wobble).
// Distinct from the cell chunk grid in ChunkSystem.
// =============================================================================
//
// Resolution algorithm (ChunkGrid_ResolveChunkAtPosition @ 0x0087d9a0):
//   shifted_x = wx + grid.x_shift     // x_shift = world_size_pixels / 2
//   shifted_y = wy + grid.y_shift     // y_shift = 14 * 512 = BASE_Y
//   chunk_cx  = floor(shifted_x / 512) mod grid.width    (clamped >=0)
//   chunk_cy  = clamp(floor(shifted_y / 512), 0, grid.height-1)
//   sub_x     = shifted_x mod 512
//   sub_y     = shifted_y mod 512
//   chunk     = chunks[chunk_cy * grid.width + chunk_cx]
//   if chunk.wobble_eligible == 0 || chunk.force_original != 0: return chunk
//   probe neighbors based on (sub_x,sub_y) thresholds 42 / 470 (0x2A / 0x1D6).
//   first neighbor with a different chunk* triggers the wobble. otherwise
//   return the original chunk (no wobble needed).
//   if either side has wavy_edge==0: simplex-only offset.
//   else: full sin/cos + simplex offset.
//   resolved = chunks[(shifted_x + dx) >> 9][(shifted_y + dy) >> 9]
//
// For the 2025-01-25 build, grid is 70x48 in normal mode.

struct BiomeChunk {
    u32 vtable;                                                        // +0x00
    u32 unknown_04;                                                    // +0x04
    MsvcString biome_name       [[name("biomeName"),
        comment("e.g. \"$biome_lake\", \"$biome_holymountain\"")]];   // +0x08..+0x1F

    // +0x20..+0xC3: many fields (texture handles, biome XML data, etc.)
    padding[0xA4];

    // Edge-noise control bytes — read by ChunkGrid_ResolveChunkAtPosition
    // (noita.exe @ 0x0087d9a0).
    u8 wobble_eligible          [[name("wobbleEligible"),
        comment("If 0, return original chunk (no wobble at all)")]];   // +0xC4
    u8 wavy_edge                [[name("wavyEdge"),
        comment("If 0 on either side of an edge, only simplex noise (no sin/cos wobble)")]]; // +0xC5
    u8 force_original           [[name("forceOriginal"),
        comment("If non-zero, short-circuit to original chunk")]];     // +0xC6
    u8 unknown_C7;                                                     // +0xC7

    // +0xC8..+0x2A3: more biome state (background sprite, music, etc.)
    padding[0x1DC];

    // +0x2A4: biome data pointer — non-NULL gates pixel-scene placement
    // (PixelScene_PlaceEntryIfValid at noita.exe 0x0087d7e0 places only
    // if this is non-null).
    u32 biome_data_ptr          [[name("biomeDataPtr"),
        comment("Pixel-scene placement gate (non-null => allowed)")]];  // +0x2A4

    // +0x2A8..+0x2D7: more biome state (texture/audio handles, etc.)
    padding[0x30];

    // +0x2D8: XML path the biome was loaded from — e.g.
    // "data/biomes/coalmine.xml", "data/biomes/tower_coalmine.xml".
    // Populated by ProceduralTerrain_Init (noita.exe @ 0x0087a...) from the
    // <BiomesToLoad> list in data/biomes/_pixel_scenes.xml. The allocation
    // size (Biome_InitializeFromConfig @ 0x0086b9f0 does operator_new(0x2F0))
    // places this string as the last member of the struct.
    MsvcString xml_path         [[name("xmlPath"),
        comment("Full XML file path, e.g. \"data/biomes/tower_coalmine.xml\"")]]; // +0x2D8..+0x2EF
};

// PixelSceneEntry — what `LoadPixelScene` queues. 0x90 bytes / entry.
// Three MsvcStrings (materials filename, colors filename, background filename)
// follow the position. The flag at +0x58 distinguishes background scenes.
struct PixelSceneEntry {
    u32 chunk_system_back_ref   [[comment("Back-pointer to ChunkSystem")]];   // +0x00
    s32 x                       [[comment("World X (top-left)")]];            // +0x04
    s32 y                       [[comment("World Y (top-left)")]];            // +0x08
    MsvcString materials_filename [[name("materialsFile")]];                  // +0x0C..+0x23
    MsvcString colors_filename    [[name("colorsFile")]];                     // +0x24..+0x3B
    MsvcString background_filename [[name("backgroundFile")]];                // +0x3C..+0x53
    padding[0x04];                                                            // +0x54..+0x57
    // +0x58/+0x59 are the two boolean parameters threaded through
    // `LoadPixelScene(... skip_biome_checks, skip_edge_textures, ...)` (see
    // LuaImpl_LoadPixelScene). Names are best-guess from that signature; in
    // dumps they don't perfectly correlate with which scenes look like
    // backgrounds-only, so treat as raw opaque flags until further verified.
    u8 flag_skip_biome_checks   [[name("flagSkipBiomeChecks")]];              // +0x58
    u8 flag_skip_edge_textures  [[name("flagSkipEdgeTextures")]];             // +0x59
    padding[0x36];                                                            // +0x5A..+0x8F
};

struct BiomeGrid {
    // The biome grid object also owns the pixel scene queue at +0x00 and an
    // alternate-scenes vector at +0x0C. Both are std::vector<PixelSceneEntry>
    // (24 bytes each, stride 0x90).
    u32 scenes_begin            [[name("scenesBegin"),
        comment("PixelSceneEntry* — start of placed pixel scenes")]];   // +0x00
    u32 scenes_end              [[name("scenesEnd")]];                  // +0x04
    u32 scenes_capacity_end     [[name("scenesCapacityEnd")]];          // +0x08
    u32 scenes_alt_begin        [[name("scenesAltBegin"),
        comment("Second PixelSceneEntry vector (alt / background?)")]]; // +0x0C
    u32 scenes_alt_end          [[name("scenesAltEnd")]];               // +0x10
    u32 scenes_alt_capacity_end [[name("scenesAltCapacityEnd")]];       // +0x14

    padding[0x20];                                                      // +0x18..+0x37

    // +0x38: world->grid coordinate shift (added to input X before chunk indexing).
    // Equals pixel-width / 2 of the biome map (e.g. 35*512 = 17920 normal).
    double x_shift              [[name("xShift")]];                    // +0x38
    // +0x40: Y shift, matches BASE_Y = 14*512 = 7168.
    double y_shift              [[name("yShift")]];                    // +0x40

    padding[0x08];                                                     // +0x48..+0x4F

    s32 width                   [[comment("Grid width in chunks (70 normal, may differ NG+)")]];   // +0x50
    s32 height                  [[comment("Grid height in chunks (48 normal)")]];                  // +0x54
    s32 total_count             [[comment("width * height")]];                                     // +0x58
    s32 unknown_5C;                                                    // +0x5C
    u32 unknown_60;                                                    // +0x60
    u32 unknown_64;                                                    // +0x64

    // +0x68: chunks pointer — array of BiomeChunk* of size width*height,
    // indexed as chunks[chunk_cy * width + chunk_cx].
    u32 chunks_ptr              [[name("chunksPtr"),
        color("FFCC44"),
        comment("BiomeChunk*[width*height]")]];                        // +0x68
    s32 chunks_count            [[name("chunksCount")]];               // +0x6C
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
// CellFactory / CellData — material lookup (names, colors, textures)
// =============================================================================

// CellTexture — per-material tile texture. Pointed to from CellData+0x88.
// Pixel format is BGRA8 (byte order in memory: B, G, R, A per pixel).
struct CellTexture {
    s32 width                   [[comment("Texture width in pixels")]];
    s32 height                  [[comment("Texture height in pixels")]];
    u32 unknown_08;
    u32 pixel_data_ptr          [[comment("u8[] BGRA8 pixels, size = width*height*4")]];
};

// CellData (0x290 bytes stride) — one per material type in CellFactory array
struct CellData {
    MsvcString name             [[comment("Material name string")]];  // +0x00..+0x17
    padding[0x4C];                                                    // +0x18..+0x63
    u32 fallback_color          [[comment("ARGB color used when texture_ptr is NULL")]];  // +0x64
    padding[0x20];                                                    // +0x68..+0x87
    CellTexture *texture_ptr : u32  [[comment("Per-material tile texture, or NULL")]];    // +0x88
    padding[0x204];                                                   // +0x8C..+0x28F
} [[static]];

// CellFactory — material table and per-material lookups
struct CellFactory {
    padding[0x18];                                                    // +0x00..+0x17
    u32 cell_data_array_ptr     [[comment("CellData* — base of material array, stride 0x290")]];  // +0x18
    padding[0x08];                                                    // +0x1C..+0x23
    s32 material_count          [[comment("Number of materials in cell_data_array")]];  // +0x24
    u32 material_0_color        [[comment("ARGB color for material ID 0")]];            // +0x28
};

// =============================================================================
// Chunk System / CellGrid — pixel-world storage (512x512 cells per chunk)
// =============================================================================
//
// World→cell lookup (from CellGrid_GetCellPtr @ 0x0089c2d0):
//   chunk_cx  = ((wx >> 9) - 0x100) & 0x1ff
//   chunk_cy  = ((wy >> 9) - 0x100) & 0x1ff
//   chunk_idx = chunk_cy * 0x200 + chunk_cx      // into chunk_table_ptr
//   cell_idx  = (wy & 0x1ff) * 512 + (wx & 0x1ff) // into chunk->cell_slots_ptr
//
// A cell slot value of 0 means "air" (unloaded chunks also appear as air).
// Non-NULL slots point to a polymorphic Cell* with a vtable at +0.

// Chunk — heap-allocated, one per loaded 512x512 region.
struct Chunk {
    u32 cell_slots_ptr          [[comment("u32[512*512] — each slot is a Cell* (0 = air)")]];
    // remaining fields (dirty rects, pixel_scene state, etc.) unmapped
};

// CellGrid — embedded in ChunkSystem at +0x500; also returned by ChunkSystem vtable[3].
// The chunk table is 512x512 = 262144 Chunk* entries (1MB), not defined inline.
struct CellGrid {
    u32 vtable;
    u32 unknown_04;
    u32 chunk_table_ptr         [[comment("Chunk*[0x40000] — NULL entry = chunk not loaded")]];
};

// ChunkSystem — top-level owner of the cell grid. Pointed to by GameGlobals+0x10.
// vtable[3] (slot offset 0x0C) returns the embedded CellGrid at +0x500.
struct ChunkSystem {
    u32 vtable;
    padding[0x4FC];
    CellGrid cell_grid          [[name("cellGrid"), comment("Embedded at +0x500 — same object as vtable[3]()")]];
};

// CellMaterialInfo — returned by Cell->vtable[12]() (__thiscall).
// Only the material ID offset is mapped.
struct CellMaterialInfo {
    padding[0x30];
    s32 material_id             [[comment("Index into CellFactory.cell_data_array_ptr")]];
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
s32 g_ngPlusCount            @ 0x01205024 [[color("AAFFAA"), comment("NG+ replay count, 0 on fresh run")]];
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
