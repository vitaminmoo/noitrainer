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

// #############################################################################
// # Ghidra-synced gameplay structs (auto-ported from the /Noita type category)
// #
// # These mirror the struct layouts in the ghidra database 1:1 (field offsets,
// # sizes, and total size all match). Unlike the hand-curated structs above,
// # most of these have NOT been re-validated against live memory dumps — field
// # names come from ghidra and unmodeled regions are filled with padding[] to
// # preserve exact size. *Component structs inline the 0x48-byte ComponentHeader
// # and model their body from +0x48 (a couple of undersized components that
// # predate that layout are modeled literally and noted inline).
// #############################################################################

// -----------------------------------------------------------------------------
// Shared helper / geometry types (Vec2, AABB, Transform2D, Bitset, ...)
// -----------------------------------------------------------------------------
// =============================================================================
// Vec2 (0x08 bytes)
// =============================================================================
struct Vec2 {
    float x;
    float y;
} [[color("AAAAAA")]];

// =============================================================================
// Vec2i (0x08 bytes)
// =============================================================================
struct Vec2i {
    s32 x;
    s32 y;
} [[color("AAAAAA")]];

// =============================================================================
// AABB (0x10 bytes)
// =============================================================================
struct AABB {
    float min_x;
    float min_y;
    float max_x;
    float max_y;
} [[color("AAAAAA")]];

// =============================================================================
// IAABB (0x10 bytes)
// =============================================================================
struct IAABB {
    s32 min_x;
    s32 min_y;
    s32 max_x;
    s32 max_y;
} [[color("AAAAAA")]];

// =============================================================================
// FColor (0x10 bytes)
// =============================================================================
struct FColor {
    float r;
    float g;
    float b;
    float a;
} [[color("AAAAAA")]];

// =============================================================================
// Transform2D (0x20 bytes)
// =============================================================================
struct Transform2D {
    float posX;
    float posY;
    float m00;
    float m01;
    float m10;
    float m11;
    float scaleX;
    float scaleY;
} [[color("AAAAAA")]];

// =============================================================================
// XForm (0x20 bytes)
// =============================================================================
struct XForm {
    float pos_x;
    float pos_y;
    float rotation;
    float scale_x;
    float scale_y;
    padding[0x20 - 0x14];                            // tail to exact size
} [[color("AAAAAA")]];

// =============================================================================
// PixelRect (0x10 bytes)
// =============================================================================
struct PixelRect {
    s32 minX;
    s32 minY;
    s32 maxX;
    s32 maxY;
} [[color("AAAAAA")]];

// =============================================================================
// RaycastResult (0x1C bytes)
// =============================================================================
struct RaycastResult {
    u8 status;
    padding[3];                                      // 0x01..0x03
    u32 hitCell [[comment("undefined* hit cell")]];
    s32 hitX;
    s32 hitY;
    bool reachedEnd;
    padding[3];                                      // 0x11..0x13
    s32 blockedCount;
    s32 passedCount;
} [[color("AAAAAA")]];

// =============================================================================
// ValueRange (0x08 bytes)
// =============================================================================
struct ValueRange {
    float min;
    float max;
} [[color("AAAAAA")]];

// =============================================================================
// ValueRangeInt (0x08 bytes)
// =============================================================================
struct ValueRangeInt {
    s32 min;
    s32 max;
} [[color("AAAAAA")]];

// =============================================================================
// LensValueFloat (0x0C bytes)
// =============================================================================
struct LensValueFloat {
    float value;
    float target;
    float speed;
} [[color("AAAAAA")]];

// =============================================================================
// LensValueInt (0x0C bytes)
// =============================================================================
struct LensValueInt {
    s32 value;
    s32 target;
    float speed;
} [[color("AAAAAA")]];

// =============================================================================
// LensValueBool (0x0C bytes)
// =============================================================================
struct LensValueBool {
    bool value;
    padding[3];                                      // 0x01..0x03
    bool target;
    padding[3];                                      // 0x05..0x07
    float speed;
} [[color("AAAAAA")]];

// =============================================================================
// Bitset256 (0x20 bytes)
// =============================================================================
struct Bitset256 {
    u32 words[8];
} [[color("AAAAAA")]];

// =============================================================================
// Bitset512 (0x40 bytes)
// =============================================================================
struct Bitset512 {
    u32 words[16];
} [[color("AAAAAA")]];

// -----------------------------------------------------------------------------
// Config & gun data structs (Config*, GunActionData, SpellDeckRecord, ...)
// -----------------------------------------------------------------------------
// =============================================================================
// batch_02 — non-component data/config structs (all fields from offset 0x00)
// =============================================================================

// =============================================================================
// ConfigDamageCritical (0x10 bytes)
// =============================================================================
struct ConfigDamageCritical {
    float chance;
    float damage_multiplier;
    padding[0x10 - 0x08];                          // tail to exact size
} [[color("FF4444")]];

// =============================================================================
// ConfigDamagesByType (0x40 bytes)
// =============================================================================
struct ConfigDamagesByType {
    u32 vtable [[comment("vtable*")]];
    float melee;
    float projectile;
    float explosion;
    float electricity;
    float fire;
    float drill;
    float slice;
    float ice;
    float healing;
    float physics_hit;
    float radioactive;
    float poison;
    float overeating;
    float curse;
    float holy;
} [[color("FF4444")]];

// =============================================================================
// ConfigDrugFx (0x1C bytes)
// =============================================================================
struct ConfigDrugFx {
    u32 vftable [[comment("vtable*")]];
    u32 field_0x4;
    u32 field_0x8;
    u32 field_0xc;
    u32 field_0x10;
    u32 field_0x14;
    u32 field_0x18;
} [[color("8888AA")]];

// =============================================================================
// ConfigExplosion (0x174 = 372 bytes)
// =============================================================================
struct ConfigExplosion {
    padding[0x44];                                 // 0x00..0x43 unmodeled base
    // +0x44
    float direction_x;
    float direction_y;
    float world_x;
    float world_y;
    float knockback_force;
    s32 source_entity;
    s32 shooter_entity;
    padding[0xA4 - 0x60];                           // 0x60..0xa3 gap
    float blood_radius_multiplier;                  // 0xA4
    float blood_amount_multiplier;
    padding[0xB0 - 0xAC];                           // 0xac..0xaf gap
    bool play_hit_sound;                            // 0xB0
    padding[0xB8 - 0xB1];                           // 0xb1..0xb7 gap
    s32 critical_chance_percent;                    // 0xB8
    float critical_damage_multiplier;
    bool is_critical_hit;                           // 0xC0
    padding[0xC4 - 0xC1];                           // 0xc1..0xc3 gap
    float projectile_speed;                         // 0xC4
    MsvcString source_description_override;          // 0xC8 (24B)
    padding[0x174 - 0xE0];                          // tail to exact size
} [[color("FF4444")]];

// =============================================================================
// ConfigGunActionInfo (0x23C = 572 bytes)
// =============================================================================
struct ConfigGunActionInfo {
    u32 vtable [[comment("vtable*")]];
    MsvcString action_id;
    MsvcString action_name;
    MsvcString action_description;
    MsvcString action_sprite_filename;
    MsvcString action_unidentified_sprite_filename;
    s32 action_type;
    MsvcString action_spawn_level;
    MsvcString action_spawn_probability;
    MsvcString action_spawn_requires_flag;
    bool action_spawn_manual_unlock;                // 0xC8
    padding[3];                                     // 0xc9..0xcb
    s32 action_max_uses;                            // 0xCC
    MsvcString custom_xml_file;                     // 0xD0
    float action_mana_drain;                        // 0xE8
    bool action_is_dangerous_blast;                 // 0xEC
    padding[3];                                     // 0xed..0xef
    s32 action_draw_many_count;                     // 0xF0
    bool action_ai_never_uses;                      // 0xF4
    bool action_never_unlimited;
    bool state_shuffled;
    padding[1];                                     // 0xf7
    s32 state_cards_drawn;                          // 0xF8
    bool state_discarded_action;                    // 0xFC
    bool state_destroyed_action;
    padding[2];                                     // 0xfe..0xff
    s32 fire_rate_wait;                             // 0x100
    float speed_multiplier;
    float child_speed_multiplier;
    float dampening;
    float explosion_radius;
    float spread_degrees;
    float pattern_degrees;
    float screenshake;
    float recoil;
    float damage_melee_add;
    float damage_projectile_add;
    float damage_electricity_add;
    float damage_fire_add;
    float damage_explosion_add;
    float damage_ice_add;
    float damage_slice_add;
    float damage_healing_add;
    float damage_curse_add;
    float damage_drill_add;
    float damage_null_all;
    s32 damage_critical_chance;
    float damage_critical_multiplier;
    float explosion_damage_to_materials;
    float knockback_force;
    s32 reload_time;
    s32 lightning_count;
    MsvcString material;                            // 0x168
    s32 material_amount;                            // 0x180
    MsvcString trail_material;                      // 0x184
    s32 trail_material_amount;                      // 0x19C
    s32 bounces;
    float gravity;
    float light;
    float blood_count_multiplier;
    s32 gore_particles;
    s32 ragdoll_fx;
    bool friendly_fire;                             // 0x1B8
    padding[3];                                     // 0x1b9..0x1bb
    float physics_impulse_coeff;                    // 0x1BC
    s32 lifetime_add;
    MsvcString sprite;                              // 0x1C4
    MsvcString extra_entities;
    MsvcString game_effect_entities;
    MsvcString sound_loop_tag;
    MsvcString projectile_file;                     // 0x224..0x23B
} [[color("FF4444")]];

// =============================================================================
// ConfigLaser (0x32 = 50 bytes)
// =============================================================================
struct ConfigLaser {
    u32 vftable [[comment("vtable*")]];
    u32 field_0x4;
    u32 field_0x8;
    u32 field_0xc;
    u32 field_0x10;
    u32 field_0x14;
    u32 field_0x18;
    u32 field_0x1c;
    u8 field_0x20;                                  // 0x20
    padding[3];                                     // 0x21..0x23
    u32 field_0x24;                                 // 0x24
    u16 field_0x28;                                 // 0x28
    padding[2];                                     // 0x2a..0x2b
    u32 field_0x2c;                                 // 0x2c
    u16 field_0x30;                                 // 0x30..0x31
} [[color("8888AA")]];

// =============================================================================
// ConfigNpcParty (0x30 = 48 bytes)
// =============================================================================
struct ConfigNpcParty {
    u32 vftable [[comment("vtable*")]];
    u32 field_0x4;
    u32 field_0x8;
    u8 field_0xc;                                   // 0xc
    padding[3];                                     // 0xd..0xf
    u32 field_0x10;
    u32 field_0x14;
    u32 field_0x18;
    u32 field_0x1c;
    u32 field_0x20;
    u32 field_0x24;
    u32 field_0x28;
    u32 field_0x2c;
} [[color("8888AA")]];

// =============================================================================
// ConfigPendingPortal (0x3C = 60 bytes)
// =============================================================================
struct ConfigPendingPortal {
    u32 vtable [[comment("vtable*")]];
    float position_x;
    float position_y;
    float target_position_x;
    float target_position_y;
    u32 id;
    u32 target_id;
    bool is_at_home;                                // 0x1c
    padding[0x38 - 0x1D];                           // 0x1d..0x37 gap
    u32 entity [[comment("Entity* / pointer")]];    // 0x38
} [[color("8888AA")]];

// =============================================================================
// ConfigSliders (0x12C = 300 bytes)
// =============================================================================
struct ConfigSliders {
    u32 vftable [[comment("vtable*")]];
    padding[0xE8 - 0x04];                           // 0x04..0xe7 gap
    u32 field_0xe8;                                 // 0xe8
    u32 field_0xec;
    u32 field_0xf0;
    u32 field_0xf4;
    u32 field_0xf8;
    u32 field_0xfc;
    u32 field_0x100;
    u32 field_0x104;
    padding[0x110 - 0x108];                         // 0x108..0x10f gap
    u32 field_0x110;                                // 0x110
    u32 field_0x114;
    u32 field_0x118;
    u32 field_0x11c;
    u32 field_0x120;
    u32 field_0x124;
    u32 field_0x128;                                // 0x128..0x12b
} [[color("8888AA")]];

// =============================================================================
// ConfigSound (0x3C = 60 bytes)
// =============================================================================
struct ConfigSound {
    u32 vftable [[comment("vtable*")]];
    u8 field_0x4;                                   // 0x4
    padding[0x14 - 0x05];                           // 0x5..0x13 gap
    u32 field_0x14;                                 // 0x14
    u32 field_0x18;
    u32 field_0x1c;
    u16 field_0x20;                                 // 0x20
    u8 field_0x22;                                  // 0x22
    padding[1];                                     // 0x23
    u32 field_0x24;                                 // 0x24
    u32 field_0x28;
    u32 field_0x2c;
    u32 field_0x30;
    u32 field_0x34;
    u32 field_0x38;                                 // 0x38..0x3b
} [[color("8888AA")]];

// =============================================================================
// StainLevelConfig (0xA0 = 160 bytes)
// =============================================================================
struct StainLevelConfig {
    MsvcString effectId;
    MsvcString uiName;
    MsvcString uiDescription;
    MsvcString uiIcon;
    bool fireAccelerantFlag;                        // 0x60
    bool clearsAmountFlag;                          // 0x61
    padding[2];                                     // 0x62..0x63
    MsvcString effectName;                          // 0x64
    float threshold;                                // 0x7c
    float duration;                                 // 0x80
    MsvcString auxString;                           // 0x84
    bool flag9c;                                    // 0x9c
    bool poisonLikeFlag;                            // 0x9d
    padding[0xA0 - 0x9E];                           // tail to exact size
} [[color("8888AA")]];

// =============================================================================
// GunActionData (0x278 = 632 bytes)
// =============================================================================
struct GunActionData {
    MsvcString action_id;
    padding[572] [[name("config"), comment("ConfigGunActionInfo, 572 bytes")]]; // 0x18
    StdVectorHeader child_actions_a;                // 0x254 (std::vector, 12B)
    StdVectorHeader child_actions_b;                // 0x260
    StdVectorHeader child_actions_c;                // 0x26c..0x277
} [[color("8888AA")]];

// =============================================================================
// GunShotState (0x98 = 152 bytes)
// =============================================================================
struct GunShotState {
    u8 action_type;                                 // 0x0
    padding[3];                                     // 0x1..0x3
    StdVectorHeader container_4;                     // 0x4 (std::vector, 12B)
    StdVectorHeader container_10;                    // 0x10
    StdMapHeader spell_action_map;                   // 0x1c (std::map, 8B)
    MsvcString extra_string;                         // 0x24
    u32 shot_effects;                                // 0x3c
    float recoil_force;                              // 0x40
    StdVectorHeader pending_instructions;            // 0x44
    StdVectorHeader drawn_action_results;            // 0x50
    StdVectorHeader shot_spell_items;                // 0x5c
    StdVectorHeader visualizer_queue;                // 0x68
    s32 recharge_frames;                             // 0x74
    bool insufficient_mana;                          // 0x78
    bool consumed_action;                            // 0x79
    bool ended_out_of_mana;                          // 0x7a
    padding[1];                                      // 0x7b
    s32 consumed_at_least_one_action;                // 0x7c
    bool shooter_is_player;                          // 0x80
    padding[3];                                      // 0x81..0x83
    StdVectorHeader extra_vector;                    // 0x84
    StdMapHeader trigger_strings;                    // 0x90..0x97 (std::map, 8B)
} [[color("4444FF")]];

// =============================================================================
// SpellDeckRecord (0x14 = 20 bytes)
// =============================================================================
struct SpellDeckRecord {
    StdVectorHeader deck;                            // 0x0 (std::vector, 12B)
    s32 param;                                       // 0xc
    bool flag;                                       // 0x10
    padding[0x14 - 0x11];                            // tail to exact size
} [[color("8888AA")]];

// =============================================================================
// OrbPersistenceData (0x38 = 56 bytes)
// =============================================================================
struct OrbPersistenceData {
    u32 persistenceVtableOrData [[comment("vtable* / data pointer")]];
    s32 orbCountAllTime;
    padding[0x38 - 0x08];                            // tail to exact size
} [[color("8888AA")]];

// -----------------------------------------------------------------------------
// Stats / world data structs + Message_* (part 1)
// -----------------------------------------------------------------------------
// =============================================================================
// GameStats (0xC0 / 192 bytes)
// =============================================================================
struct GameStats {
    bool dead;
    padding[3];                                  // pad_01
    s32 death_count;
    s32 streaks;
    s32 world_seed;
    u8 killed_by[24];
    u8 killed_by_extra[24];
    s32 death_pos_x;
    s32 death_pos_y;
    padding[4];                                  // pad_4c
    double playtime;
    char playtime_str[24];
    s32 places_visited;
    s32 enemies_killed;
    s32 heart_containers;
    padding[4];                                  // pad_7c
    double hp;
    double gold;
    double gold_all;
    bool gold_infinite;
    padding[3];                                  // pad_99
    s32 items;
    s32 projectiles_shot;
    s32 kicks;
    double damage_taken;
    double healed;
    s32 teleports;
    s32 wands_edited;
    s32 biomes_visited_with_wands;
} [[color("FF4444")]];

// =============================================================================
// GlobalStats (0x340 / 832 bytes)
// =============================================================================
struct GlobalStats {
    u32 vtable                 [[comment("vftable*")]];
    s32 STATS_VERSION;
    s32 DEBUG_HOW_MANY_TIMES_DONE;
    bool DEBUG_IS_ON;
    padding[3];                                  // pad_0d
    s32 DEBUG_HOW_MANY_RESETS;
    bool DEBUG_FIXED_STATS;
    bool session_dead;
    padding[2];                                  // pad_16
    u64 KEY_VALUE_STATS;
    u8 session[200]                [[comment("GameStats-like blob, 200 bytes")]];
    u8 highest[200]                [[comment("GameStats-like blob, 200 bytes")]];
    u8 global[200]                 [[comment("GameStats-like blob, 200 bytes")]];
    u8 prev_best[200]              [[comment("GameStats-like blob, 200 bytes")]];
} [[color("8888AA")]];

// =============================================================================
// EntityTemplate (0x88 / 136 bytes)
// =============================================================================
struct EntityTemplate {
    MsvcString xmlFilename;                       // 0x00 (24B)
    padding[24];                                 // 0x18 gap to transform
    Transform2D transform;                        // 0x30 (32B)
    MsvcString tag;                               // 0x50 (24B)
    s32 internId;                                 // 0x68
    bool enabledFlag;                             // 0x6C
    padding[136 - 0x6D];                          // tail to exact size
} [[color("AAAAAA")]];

// =============================================================================
// MetaData (0x30 / 48 bytes)
// =============================================================================
struct MetaData {
    u32 vtable                 [[comment("vftable*")]];
    s32 field_04;
    s32 field_08;
    s32 field_0C;
    s32 field_10;
    s32 field_14;
    s32 field_18;
    s32 field_1C;
    u32 pSentinel_20           [[comment("pointer")]];
    u32 pSentinel_24           [[comment("pointer")]];
    u32 pSentinel_28           [[comment("pointer")]];
    s32 field_2C;
} [[color("AAAAAA")]];

// =============================================================================
// CellReaction (0x44 / 68 bytes)
// =============================================================================
struct CellReaction {
    padding[4];                                  // 0x00 gap (no field)
    s32 probability;
    s32 input_cell1_id;
    s32 input_cell2_id;
    s32 output_cell1_id;
    s32 output_cell2_id;
    bool has_input_cell3;
    padding[3];                                  // 0x19 gap
    s32 input_cell3_id;
    s32 output_cell3_id;
    s32 cosmetic_particle_id;
    s32 req_lifetime;
    u8 cosmetic_mode_cell1;
    u8 cosmetic_mode_cell2;
    u8 cosmetic_param1;
    u8 cosmetic_param2;
    bool destroy_horizontally_lonely_pixels;
    bool blob_mode;
    padding[2];                                  // 0x32 gap
    s32 entity_id;
    s32 direction;
    u32 pParticleConfig        [[comment("pointer")]];
    float audio_fx_volume_1;
} [[color("8888AA")]];

// =============================================================================
// CaveStructure (0x3C / 60 bytes)
// =============================================================================
struct CaveStructure {
    u32 vftable                [[comment("vftable*")]];
    u8 field_0x4;
    padding[15];                                 // 0x05 gap
    u32 field_0x14;
    u32 field_0x18;
    u32 field_0x1c;
    u32 field_0x20;
    u32 field_0x24;
    u32 field_0x28;
    u32 field_0x2c;
    u32 field_0x30;
    u32 field_0x34;
    u32 field_0x38;
} [[color("AAAAAA")]];

// =============================================================================
// Biome (0x243 / 579 bytes)
// =============================================================================
struct Biome {
    u32 vftable                [[comment("vftable*")]];
    padding[574];                                // 0x04 .. 0x241 unmodeled
    u8 unknown_242;                              // 0x242
} [[color("AAAAAA")]];

// =============================================================================
// ProceduralBiome (0x154 / 340 bytes)
// =============================================================================
struct ProceduralBiome {
    u32 pBiomeNode             [[comment("pointer")]];
    char wangTemplateFile_data[16];
    s32 wangTemplateFile_size;
    s32 wangTemplateFile_capacity;
    char wangScriptsFile_data[16];
    s32 wangScriptsFile_size;
    s32 wangScriptsFile_capacity;
    s32 wangMapWidth;
    s32 wangMapHeight;
    s32 biomeRegionMinX;
    s32 biomeRegionMinY;
    s32 biomeRegionMaxX;
    s32 biomeRegionMaxY;
    s32 seed;
    u8 staticTileFlag;
    padding[3];                                  // 0x51 gap
    s32 wangOut_width;
    s32 wangOut_height;
    s32 wangOut_totalPixels;
    s32 wangOut_unk60;
    s32 wangOut_unk64;
    s32 wangOut_unk68;
    u32 wangOut_pixelData      [[comment("int*")]];
    s32 wangOut_capacity;
    s32 unk74;
    s32 unk78;
    u32 pixelSceneEntries_begin    [[comment("pointer")]];
    u32 pixelSceneEntries_end      [[comment("pointer")]];
    s32 pixelSceneEntries_capacity;
    s32 unk88;
    u32 extraLayerPixelData    [[comment("pointer")]];
    s32 destGrid_width;
    s32 destGrid_height;
    s32 destGrid_totalPixels;
    s32 destGrid_unk9C;
    s32 destGrid_unkA0;
    s32 destGrid_unkA4;
    u32 destGrid_pixelData     [[comment("int*")]];
    s32 destGrid_capacity;
    padding[144];                                // 0xB0 .. 0x13F unmodeled
    u32 pMaterialLookupFunc    [[comment("pointer")]];
    u32 pSpawnScriptCallback   [[comment("pointer")]];
    u32 pCachedTemplateData    [[comment("pointer")]];
    s32 cachedTemplateWidth;
    s32 cachedTemplateHeight;
} [[color("AAAAAA")]];

// =============================================================================
// Message_DamageAboutToBeReceived (0x38 / 56 bytes)
// =============================================================================
struct Message_DamageAboutToBeReceived {
    u32 vftable                [[comment("vftable*")]];
    u8 field_0x4;
    padding[3];                                  // 0x05 gap
    u32 field_0x8;
    u32 field_0xc;
    u32 field_0x10;
    u32 field_0x14;
    u32 field_0x18;
    u32 field_0x1c;
    u32 field_0x20;
    u32 field_0x24;
    u32 field_0x28;
    u8 field_0x2c;
    padding[3];                                  // 0x2D gap
    u32 field_0x30;
    u32 field_0x34;
} [[color("AA8888")]];

// =============================================================================
// Message_DamageAboutToBeReceivedPost (0x38 / 56 bytes)
// =============================================================================
struct Message_DamageAboutToBeReceivedPost {
    u32 vftable                [[comment("vftable*")]];
    u8 field_0x4;
    padding[3];                                  // 0x05 gap
    u32 field_0x8;
    u32 field_0xc;
    u32 field_0x10;
    u32 field_0x14;
    u32 field_0x18;
    u32 field_0x1c;
    u32 field_0x20;
    u32 field_0x24;
    u32 field_0x28;
    u8 field_0x2c;
    padding[3];                                  // 0x2D gap
    u32 field_0x30;
    u32 field_0x34;
} [[color("AA8888")]];

// =============================================================================
// Message_DamageReceived (0x54 / 84 bytes)
// =============================================================================
struct Message_DamageReceived {
    u32 vftable                [[comment("vftable*")]];
    u8 field_0x4;
    padding[3];                                  // 0x05 gap
    u32 field_0x8;
    u32 field_0xc;
    u32 field_0x10;
    u32 field_0x14;
    u32 field_0x18;
    u8 field_0x1c;
    padding[15];                                 // 0x1D gap
    u32 field_0x2c;
    u32 field_0x30;
    u8 field_0x34;
    padding[3];                                  // 0x35 gap
    u32 field_0x38;
    u32 field_0x3c;
    u32 field_0x40;
    u32 field_0x44;
    u32 field_0x48;
    u8 field_0x4c;
    padding[3];                                  // 0x4D gap
    u32 field_0x50;
} [[color("AA8888")]];

// =============================================================================
// Message_Explosion (0x20 / 32 bytes)
// =============================================================================
struct Message_Explosion {
    u32 vftable                [[comment("vftable*")]];
    u8 field_0x4;
    padding[3];                                  // 0x05 gap
    u32 field_0x8;
    u32 field_0xc;
    u32 field_0x10;
    u32 field_0x14;
    u32 field_0x18;
    u32 field_0x1c;
} [[color("AA8888")]];

// =============================================================================
// Message_ItemAllowedToPickUp (0x37 / 55 bytes)
// =============================================================================
struct Message_ItemAllowedToPickUp {
    u32 vftable                [[comment("vftable*")]];
    padding[50];                                 // 0x04 .. 0x35 unmodeled
    u8 field_0x36;                               // 0x36
} [[color("AA8888")]];

// -----------------------------------------------------------------------------
// Message_* (part 2) + AI/combat components
// -----------------------------------------------------------------------------
// =============================================================================
// batch_04 — Message_* data structs + *Component structs
// =============================================================================

// =============================================================================
// Message_Kick (0x0C bytes)
// =============================================================================
struct Message_Kick {
    u32 vftable;                                 // 0x00
    u8 field_0x4;                                // 0x04
    padding[3];                                  // 0x05..0x07
    u32 field_0x8;                               // 0x08
} [[color("AA8888")]];

// =============================================================================
// Message_PhysicsBodyCollision (0x2C bytes)
// =============================================================================
struct Message_PhysicsBodyCollision {
    u32 vftable;                                 // 0x00
    u8 field_0x4;                                // 0x04
    padding[3];                                  // 0x05..0x07
    u32 field_0x8;                               // 0x08
    u32 field_0xc;                               // 0x0c
    u32 field_0x10;                              // 0x10
    u32 field_0x14;                              // 0x14
    u32 field_0x18;                              // 0x18
    u32 field_0x1c;                              // 0x1c
    u32 field_0x20;                              // 0x20
    u8 field_0x24;                               // 0x24
    padding[3];                                  // 0x25..0x27
    u32 field_0x28;                              // 0x28
} [[color("AA8888")]];

// =============================================================================
// Message_PlayerDied (0xC4 bytes)
// =============================================================================
struct Message_PlayerDied {
    u32 vftable;                                 // 0x00
    padding[144];                                // 0x04..0x93
    u32 field_0x94;                              // 0x94
    padding[4];                                  // 0x98..0x9b
    u32 field_0x9c;                              // 0x9c
    padding[32];                                 // 0xa0..0xbf
    u32 field_0xc0;                              // 0xc0
} [[color("AA8888")]];

// =============================================================================
// Message_PlayerSpawned (0xAC bytes)
// =============================================================================
struct Message_PlayerSpawned {
    u32 vftable;                                 // 0x00
    padding[24];                                 // 0x04..0x1b
    u8 field_0x1c;                               // 0x1c
    padding[15];                                 // 0x1d..0x2b
    u32 field_0x2c;                              // 0x2c
    padding[116];                                // 0x30..0xa3
    u8 field_0xa4;                               // 0xa4
    padding[3];                                  // 0xa5..0xa7
    u32 field_0xa8;                              // 0xa8
} [[color("AA8888")]];

// =============================================================================
// Message_PolymorphingTo (0x30 bytes)
// =============================================================================
struct Message_PolymorphingTo {
    u32 vftable;                                 // 0x00
    u8 field_0x4;                                // 0x04
    padding[3];                                  // 0x05..0x07
    u32 field_0x8;                               // 0x08
    u32 field_0xc;                               // 0x0c
    u32 field_0x10;                              // 0x10
    u32 field_0x14;                              // 0x14
    u8 field_0x18;                               // 0x18
    padding[15];                                 // 0x19..0x27
    u32 field_0x28;                              // 0x28
    u32 field_0x2c;                              // 0x2c
} [[color("AA8888")]];

// =============================================================================
// Message_Teleported (0x28 bytes)
// =============================================================================
struct Message_Teleported {
    u32 vftable;                                 // 0x00
    u8 field_0x4;                                // 0x04
    padding[3];                                  // 0x05..0x07
    u32 field_0x8;                               // 0x08
    u32 field_0xc;                               // 0x0c
    u32 field_0x10;                              // 0x10
    u8 field_0x14;                               // 0x14
    padding[3];                                  // 0x15..0x17
    u32 field_0x18;                              // 0x18
    u32 field_0x1c;                              // 0x1c
    u32 field_0x20;                              // 0x20
    u32 field_0x24;                              // 0x24
} [[color("AA8888")]];

// =============================================================================
// Message_UseItem (0x28 bytes)
// =============================================================================
struct Message_UseItem {
    u32 vftable;                                 // 0x00
    u8 field_0x4;                                // 0x04
    padding[3];                                  // 0x05..0x07
    u32 field_0x8;                               // 0x08
    u32 field_0xc;                               // 0x0c
    u32 field_0x10;                              // 0x10
    u16 field_0x14;                              // 0x14
    u8 field_0x16;                               // 0x16
    padding[1];                                  // 0x17
    u32 field_0x18;                              // 0x18
    u32 field_0x1c;                              // 0x1c
    u32 field_0x20;                              // 0x20
    u32 field_0x24;                              // 0x24
} [[color("AA8888")]];

// =============================================================================
// Message_VisitedNewBiome (0x44 bytes)
// =============================================================================
struct Message_VisitedNewBiome {
    u32 vftable;                                 // 0x00
    u8 field_0x4;                                // 0x04
    padding[3];                                  // 0x05..0x07
    u32 field_0x8;                               // 0x08
    u32 field_0xc;                               // 0x0c
    u32 field_0x10;                              // 0x10
    u8 field_0x14;                               // 0x14
    padding[15];                                 // 0x15..0x23
    u32 field_0x24;                              // 0x24
    u32 field_0x28;                              // 0x28
    u8 field_0x2c;                               // 0x2c
    padding[15];                                 // 0x2d..0x3b
    u32 field_0x3c;                              // 0x3c
    u32 field_0x40;                              // 0x40
} [[color("AA8888")]];

// =============================================================================
// AdvancedFishAIComponent (0x89 bytes)
// =============================================================================
struct AdvancedFishAIComponent {
    ComponentHeader header      [[inline]];      // 0x00..0x47
    // +0x48
    float move_check_range_min;                  // 0x48
    float move_check_range_max;                  // 0x4c
    bool flock;                                  // 0x50
    bool avoid_predators;                        // 0x51
    bool mHasTargetDirection;                    // 0x52
    padding[1];                                  // 0x53
    Vec2 mTargetPos;                             // 0x54
    Vec2 mTargetVec;                             // 0x5c
    Vec2 mLastFramesMovementAreaMin;             // 0x64
    Vec2 mLastFramesMovementAreaMax;             // 0x6c
    s32 mNumFailedTargetSearches;                // 0x74
    s32 mNextFrameCheckAreWeStuck;               // 0x78
    s32 mNextFrameCheckFlockWants;               // 0x7c
    s32 mNextFramePredatorAvoidance;             // 0x80
    float mScared;                               // 0x84
    bool mWantsToBeInFlock;                      // 0x88
} [[color("AAAAAA")]];

// =============================================================================
// AIAttackComponent (0xCC bytes)
// =============================================================================
struct AIAttackComponent {
    ComponentHeader header      [[inline]];      // 0x00..0x47
    // +0x48
    s32 use_probability;                         // 0x48
    float min_distance;                          // 0x4c
    float max_distance;                          // 0x50
    float angular_range_deg;                     // 0x54
    s32 state_duration_frames;                   // 0x58
    s32 frames_between;                          // 0x5c
    s32 frames_between_global;                    // 0x60
    MsvcString animation_name;                   // 0x64
    bool attack_landing_ranged_enabled;          // 0x7c
    padding[3];                                  // 0x7d..0x7f
    s32 attack_ranged_action_frame;              // 0x80
    float attack_ranged_offset_x;                // 0x84
    float attack_ranged_offset_y;                // 0x88
    float attack_ranged_root_offset_x;           // 0x8c
    float attack_ranged_root_offset_y;           // 0x90
    bool attack_ranged_use_message;              // 0x94
    bool attack_ranged_predict;                  // 0x95
    padding[2];                                  // 0x96..0x97
    MsvcString attack_ranged_entity_file;        // 0x98
    s32 attack_ranged_entity_count_min;          // 0xb0
    s32 attack_ranged_entity_count_max;          // 0xb4
    bool attack_ranged_use_laser_sight;          // 0xb8
    bool attack_ranged_aim_rotation_enabled;     // 0xb9
    padding[2];                                  // 0xba..0xbb
    float attack_ranged_aim_rotation_speed;      // 0xbc
    float attack_ranged_aim_rotation_shooting_ok_angle_deg; // 0xc0
    float mRangedAttackCurrentAimAngle;          // 0xc4
    s32 mNextFrameUsable;                        // 0xc8
} [[color("AAAAAA")]];

// =============================================================================
// AltarComponent (0xF8 bytes)
// =============================================================================
struct AltarComponent {
    ComponentHeader header      [[inline]];      // 0x00..0x47
    // +0x48
    u8 field_0x48;                               // 0x48
    padding[15];                                 // 0x49..0x57
    u32 field_0x58;                              // 0x58
    u32 field_0x5c;                              // 0x5c
    u32 field_0x60;                              // 0x60
    u32 field_0x64;                              // 0x64
    u32 field_0x68;                              // 0x68
    u32 field_0x6c;                              // 0x6c
    u32 field_0x70;                              // 0x70
    u32 field_0x74;                              // 0x74
    u32 field_0x78;                              // 0x78
    u32 field_0x7c;                              // 0x7c
    u32 field_0x80;                              // 0x80
    u32 field_0x84;                              // 0x84
    u32 field_0x88;                              // 0x88
    u32 field_0x8c;                              // 0x8c
    u32 field_0x90;                              // 0x90
    u32 field_0x94;                              // 0x94
    u32 field_0x98;                              // 0x98
    u32 field_0x9c;                              // 0x9c
    u32 field_0xa0;                              // 0xa0
    u32 field_0xa4;                              // 0xa4
    u32 field_0xa8;                              // 0xa8
    u32 field_0xac;                              // 0xac
    u32 field_0xb0;                              // 0xb0
    padding[4];                                  // 0xb4..0xb7
    u32 field_0xb8;                              // 0xb8
    u32 field_0xbc;                              // 0xbc
    u32 field_0xc0;                              // 0xc0
    u32 field_0xc4;                              // 0xc4
    u32 field_0xc8;                              // 0xc8
    u32 field_0xcc;                              // 0xcc
    u32 field_0xd0;                              // 0xd0
    u32 field_0xd4;                              // 0xd4
    u32 field_0xd8;                              // 0xd8
    u32 field_0xdc;                              // 0xdc
    u32 field_0xe0;                              // 0xe0
    u32 field_0xe4;                              // 0xe4
    u32 field_0xe8;                              // 0xe8
    u32 field_0xec;                              // 0xec
    u32 field_0xf0;                              // 0xf0
    u32 field_0xf4;                              // 0xf4
} [[color("AAAAAA")]];

// =============================================================================
// AnimalAIComponent (0x3BC bytes)
// =============================================================================
struct AnimalAIComponent {
    ComponentHeader header      [[inline]];      // 0x00..0x47
    // +0x48
    s32 ai_state;                                // 0x48
    s32 ai_state_timer;                          // 0x4c
    bool keep_state_alive_when_enabled;          // 0x50
    padding[3];                                  // 0x51..0x53
    MsvcString preferred_job;                    // 0x54
    float escape_if_damaged_probability;         // 0x6c
    float attack_if_damaged_probability;         // 0x70
    float eye_offset_x;                          // 0x74
    float eye_offset_y;                          // 0x78
    bool attack_only_if_attacked;                // 0x7c
    bool dont_counter_attack_own_herd;           // 0x7d
    padding[2];                                  // 0x7e..0x7f
    float creature_detection_range_x;            // 0x80
    float creature_detection_range_y;            // 0x84
    float creature_detection_angular_range_deg;  // 0x88
    s32 creature_detection_check_every_x_frames; // 0x8c
    float max_distance_to_cam_to_start_hunting;  // 0x90
    s32 pathfinding_max_depth_no_target;         // 0x94
    s32 pathfinding_max_depth_has_target;        // 0x98
    float aggressiveness_min;                    // 0x9c
    float aggressiveness_max;                    // 0xa0
    bool tries_to_ranged_attack_friends;         // 0xa4
    bool attack_melee_enabled;                   // 0xa5
    bool attack_dash_enabled;                    // 0xa6
    bool attack_landing_ranged_enabled;          // 0xa7
    bool attack_ranged_enabled;                  // 0xa8
    padding[3];                                  // 0xa9..0xab
    float attack_knockback_multiplier;           // 0xac
    bool is_static_turret;                       // 0xb0
    padding[3];                                  // 0xb1..0xb3
    s32 attack_melee_max_distance;               // 0xb4
    s32 attack_melee_action_frame;               // 0xb8
    s32 attack_melee_frames_between;             // 0xbc
    float attack_melee_damage_min;               // 0xc0
    float attack_melee_damage_max;               // 0xc4
    float attack_melee_impulse_vector_x;         // 0xc8
    float attack_melee_impulse_vector_y;         // 0xcc
    float attack_melee_impulse_multiplier;       // 0xd0
    float attack_melee_offset_x;                 // 0xd4
    float attack_melee_offset_y;                 // 0xd8
    bool attack_melee_finish_enabled;            // 0xdc
    padding[3];                                  // 0xdd..0xdf
    padding[372] [[name("attack_melee_finish_config_explosion"), comment("ConfigExplosion, 372 bytes")]]; // 0xe0
    s32 attack_melee_finish_action_frame;        // 0x254
    float attack_dash_distance;                  // 0x258
    s32 attack_dash_frames_between;              // 0x25c
    float attack_dash_damage;                    // 0x260
    float attack_dash_speed;                     // 0x264
    float attack_dash_lob;                       // 0x268
    float attack_ranged_min_distance;            // 0x26c
    float attack_ranged_max_distance;            // 0x270
    s32 attack_ranged_action_frame;              // 0x274
    LensValueInt attack_ranged_frames_between;   // 0x278
    float attack_ranged_offset_x;                // 0x284
    float attack_ranged_offset_y;                // 0x288
    bool attack_ranged_use_message;              // 0x28c
    bool attack_ranged_predict;                  // 0x28d
    padding[2];                                  // 0x28e..0x28f
    MsvcString attack_ranged_entity_file;        // 0x290
    s32 attack_ranged_entity_count_min;          // 0x2a8
    s32 attack_ranged_entity_count_max;          // 0x2ac
    bool attack_ranged_use_laser_sight;          // 0x2b0
    bool attack_ranged_laser_sight_beam_kind;    // 0x2b1
    bool attack_ranged_aim_rotation_enabled;     // 0x2b2
    padding[1];                                  // 0x2b3
    float attack_ranged_aim_rotation_speed;      // 0x2b4
    float attack_ranged_aim_rotation_shooting_ok_angle_deg; // 0x2b8
    s32 attack_ranged_state_duration_frames;     // 0x2bc
    bool hide_from_prey;                         // 0x2c0
    padding[3];                                  // 0x2c1..0x2c3
    float hide_from_prey_target_distance;        // 0x2c4
    s32 hide_from_prey_time;                     // 0x2c8
    s32 food_material;                           // 0x2cc
    s32 food_particle_effect_material;           // 0x2d0
    bool food_eating_create_particles;           // 0x2d4
    padding[3];                                  // 0x2d5..0x2d7
    s32 eating_area_radius_x;                    // 0x2d8
    s32 eating_area_radius_y;                    // 0x2dc
    s32 mouth_offset_x;                          // 0x2e0
    s32 mouth_offset_y;                          // 0x2e4
    bool defecates_and_pees;                     // 0x2e8
    padding[3];                                  // 0x2e9..0x2eb
    s32 butt_offset_x;                           // 0x2ec
    s32 butt_offset_y;                           // 0x2f0
    float pee_velocity_x;                        // 0x2f4
    float pee_velocity_y;                        // 0x2f8
    bool needs_food;                             // 0x2fc
    bool sense_creatures;                        // 0x2fd
    bool sense_creatures_through_walls;          // 0x2fe
    bool can_fly;                                // 0x2ff
    bool can_walk;                               // 0x300
    padding[3];                                  // 0x301..0x303
    s32 path_distance_to_target_node_to_turn_around; // 0x304
    float path_cleanup_explosion_radius;         // 0x308
    float max_distance_to_move_from_home;        // 0x30c
    u32 mAiStateStack_begin [[comment("pointer")]]; // 0x310
    u32 mAiStateStack_end [[comment("pointer")]];   // 0x314
    u32 mAiStateStack_cap [[comment("pointer")]];   // 0x318
    s32 mAiStateLastSwitchFrame;                 // 0x31c
    s32 mAiStatePrev;                            // 0x320
    s32 mCreatureDetectionNextCheck;             // 0x324
    s32 mGreatestThreat;                         // 0x328
    s32 mGreatestPrey;                           // 0x32c
    s32 mSelectedMultiAttack;                    // 0x330
    bool mHasFoundPrey;                          // 0x334
    bool mHasBeenAttackedByPlayer;               // 0x335
    bool mHasStartedAttacking;                   // 0x336
    padding[1];                                  // 0x337
    s32 mNearbyFoodCount;                        // 0x338
    s32 mEatNextFrame;                           // 0x33c
    s32 mEatTime;                                // 0x340
    s32 mFrameNextGiveUp;                        // 0x344
    float mLastFramesMovementAreaMin_x;          // 0x348
    float mLastFramesMovementAreaMin_y;          // 0x34c
    Vec2 mLastFramesMovementAreaMax;             // 0x350
    s32 mFoodMaterialId;                         // 0x358
    s32 mFoodParticleEffectMaterialId;           // 0x35c
    LensValueFloat mAggression;                  // 0x360
    float mNextJumpLob;                          // 0x36c
    Vec2 mNextJumpTarget;                        // 0x370
    bool mNextJumpHasVelocity;                   // 0x378
    padding[3];                                  // 0x379..0x37b
    s32 mLastFrameJumped;                        // 0x37c
    s32 mFramesWithoutTarget;                    // 0x380
    s32 mLastFrameCanDamageOwnHerd;              // 0x384
    Vec2 mHomePosition;                          // 0x388
    s32 mLastFrameAttackWasDone;                 // 0x390
    s32 mNextFrameCanCallFriend;                 // 0x394
    s32 mNextFrameRespondFriend;                 // 0x398
    bool mHasNoticedPlayer;                      // 0x39c
    padding[3];                                  // 0x39d..0x39f
    float mRangedAttackCurrentAimAngle;          // 0x3a0
    s32 mRangedAttackNextFrame;                  // 0x3a4
    s32 mMeleeAttackNextFrame;                   // 0x3a8
    float mNextMeleeAttackDamage;                // 0x3ac
    bool mMeleeAttacking;                        // 0x3b0
    padding[3];                                  // 0x3b1..0x3b3
    s32 mMeleeAttackDashNextFrame;               // 0x3b4
    s32 mCurrentJob;                             // 0x3b8
} [[color("FF4444")]];

// =============================================================================
// AreaDamageComponent (0x9C bytes)
// =============================================================================
struct AreaDamageComponent {
    ComponentHeader header      [[inline]];      // 0x00..0x47
    // +0x48
    u32 field_0x48;                              // 0x48
    u32 field_0x4c;                              // 0x4c
    u32 field_0x50;                              // 0x50
    u32 field_0x54;                              // 0x54
    u32 field_0x58;                              // 0x58
    u32 field_0x5c;                              // 0x5c
    u32 field_0x60;                              // 0x60
    u32 field_0x64;                              // 0x64
    u32 field_0x68;                              // 0x68
    u8 field_0x6c;                               // 0x6c
    padding[15];                                 // 0x6d..0x7b
    u32 field_0x7c;                              // 0x7c
    u32 field_0x80;                              // 0x80
    u8 field_0x84;                               // 0x84
    padding[15];                                 // 0x85..0x93
    u32 field_0x94;                              // 0x94
    u32 field_0x98;                              // 0x98
} [[color("FF4444")]];

// -----------------------------------------------------------------------------
// Components (Attach–Controller)
// -----------------------------------------------------------------------------
// =============================================================================
// AttachToEntityComponent (0x94 = 148 bytes)
// =============================================================================
struct AttachToEntityComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    u32 unknown_48;
    u8  unknown_4c;
    padding[3];
    u8  unknown_50;
    padding[15];                                 // 0x51..0x5f
    u32 unknown_60;
    u32 unknown_64;
    u32 unknown_68;
    u16 unknown_6c;
    padding[2];                                  // 0x6e..0x6f
    u32 unknown_70;
    u32 unknown_74;
    u32 unknown_78;
    u32 unknown_7c;
    u32 unknown_80;
    u32 unknown_84;
    u32 unknown_88;
    u32 unknown_8c;
    u32 unknown_90;
} [[color("AAAAAA")]];

// =============================================================================
// AudioComponent (0x9c = 156 bytes)
// =============================================================================
struct AudioComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    padding[24];                                 // 0x48..0x5f
    u32  event_root [[comment("pointer")]];
    padding[20];                                 // 0x64..0x77
    u32  audio_physics_material [[comment("pointer")]];
    padding[20];                                 // 0x7c..0x8f
    bool set_latest_event_position;
    bool remove_latest_event_on_destroyed;
    bool send_message_on_event_dead;
    bool play_only_if_visible;
    s32  m_audio_physics_material;
    s32  m_latest_source;
} [[color("AAAAAA")]];

// =============================================================================
// AudioLoopComponent (0x98 = 152 bytes)
// =============================================================================
struct AudioLoopComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    MsvcString file;
    MsvcString event_name;
    bool auto_play;
    bool auto_play_if_enabled;
    bool play_on_component_enable;
    bool calculate_material_lowpass;
    bool set_speed_parameter;
    bool set_speed_parameter_only_based_on_x_movement;
    bool set_speed_parameter_only_based_on_y_movement;
    padding[1];                                  // 0x7f
    float volume_autofade_speed;
    float m_volume;
    float m_intensity;
    float m_intensity2;
    s32   m_source;
    s32   m_frame_created;
} [[color("AAAAAA")]];

// =============================================================================
// BiomeTrackerComponent (0x68 = 104 bytes)
// =============================================================================
struct BiomeTrackerComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    s32 limit_to_every_n_frame;
    padding[4];                                  // 0x4c..0x4f
    MsvcString current_biome_name;
} [[color("AAAAAA")]];

// =============================================================================
// BlackHoleComponent (0x58 = 88 bytes)
// =============================================================================
struct BlackHoleComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    float radius;
    float particle_attractor_force;
    float damage_probability;
    float damage_amount;
} [[color("FF4444")]];

// =============================================================================
// BossDragonComponent (0x131 = 305 bytes)
// =============================================================================
struct BossDragonComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    float speed;
    float speed_hunt;
    float acceleration;
    float direction_adjust_speed;
    float direction_adjust_speed_hunt;
    float gravity;
    float tail_gravity;
    float part_distance;
    s32   ground_check_offset;
    float eat_ground_radius;
    bool  eat_ground;
    padding[3];                                  // 0x71..0x73
    float hitbox_radius;
    float bite_damage;
    float target_kill_radius;
    float target_kill_ragdoll_force;
    float hunt_box_radius;
    float random_target_box_radius;
    s32   new_hunt_target_check_every;
    s32   new_random_target_check_every;
    float jump_cam_shake;
    float jump_cam_shake_distance;
    float eat_anim_wait_mult;
    MsvcString projectile_1;
    s32   projectile_1_count;
    MsvcString projectile_2;
    s32   projectile_2_count;
    MsvcString ragdoll_filename;
    s32   mTargetEntityId;
    Vec2  mTargetVec;
    float mGravVelocity;
    float mSpeed;
    Vec2  mRandomTarget;
    Vec2  mLastLivingTargetPos;
    s32   mNextTargetCheckFrame;
    s32   mNextHuntTargetCheckFrame;
    bool  mOnGroundPrev;
    padding[3];                                  // 0x11d..0x11f
    s32   mMaterialIdPrev;
    s32   mPhase;
    s32   mNextPhaseSwitchTime;
    float mPartDistance;
    bool  mIsInitialized;
} [[color("FF4444")]];

// =============================================================================
// BossHealthBarComponent (0x4b = 75 bytes)
// =============================================================================
struct BossHealthBarComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47 (incl. mOldSpritesDestroyed @0x05)
    // +0x48
    bool gui;
    bool gui_special_final_boss;
    bool in_world;
} [[color("AAAAAA")]];

// =============================================================================
// CameraBoundComponent (0x5c = 92 bytes)
// =============================================================================
struct CameraBoundComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    bool enabled;
    padding[3];                                  // 0x49..0x4b
    float distance;
    float distance_border;
    s32   max_count;
    bool  freeze_on_distance_kill;
    bool  freeze_on_max_count_kill;
    padding[2];                                  // 0x5a..0x5b
} [[color("AAAAAA")]];

// =============================================================================
// CardinalMovementComponent (0x54 = 84 bytes)
// =============================================================================
struct CardinalMovementComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    bool horizontal_movement;
    bool vertical_movement;
    bool intercardinal_movement;
    padding[1];                                  // 0x4b
    Vec2 mPrevPos;
} [[color("AAAAAA")]];

// =============================================================================
// CellEaterComponent (0x70 = 112 bytes)
// =============================================================================
struct CellEaterComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    float radius;
    s32   eat_probability;
    s32   ignored_material;
    bool  only_stain;
    bool  eat_dynamic_physics_bodies;
    bool  limited_materials;
    padding[1];                                  // 0x57
    MsvcString ignored_material_tag;
} [[color("AAAAAA")]];

// =============================================================================
// CharacterStatsComponent (0x90 = 144 bytes)
// =============================================================================
struct CharacterStatsComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    u8  unknown_48;
    padding[15];                                 // 0x49..0x57
    u32 unknown_58;
    u32 unknown_5c;
    u8  unknown_60;
    padding[15];                                 // 0x61..0x6f
    u32 unknown_70;
    u32 unknown_74;
    u8  unknown_78;
    padding[15];                                 // 0x79..0x87
    u32 unknown_88;
    u32 unknown_8c;
} [[color("AAAAAA")]];

// =============================================================================
// ConsumableTeleportComponent (0x68 = 104 bytes)
// =============================================================================
struct ConsumableTeleportComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    bool create_other_end;
    bool is_at_home;
    padding[2];                                  // 0x4a..0x4b
    float collision_radius;
    float target_location_x;
    float target_location_y;
    u32   target_id;
    u32   id;
    s32   mNextUsableFrame;
    bool  mHasOtherEnd;
    padding[3];                                  // 0x65..0x67
} [[color("AAAAAA")]];

// =============================================================================
// ControllerGoombaAIComponent (0x80 = 128 bytes)
// =============================================================================
struct ControllerGoombaAIComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    s32  mChangingDirectionCounter;
    bool auto_turn_around_enabled;
    padding[3];                                  // 0x4d..0x4f
    s32  wait_to_turn_around;
    s32  wall_hit_wait;
    bool check_wall_detection;
    padding[3];                                  // 0x59..0x5b
    float wall_detection_aabb_min_x;
    float wall_detection_aabb_max_x;
    padding[4];                                  // 0x64..0x67 (wall_detection_aabb_min_y gap)
    float wall_detection_aabb_max_y;
    bool  check_floor_detection;
    padding[3];                                  // 0x6d..0x6f
    float floor_detection_aabb_min_x;
    float floor_detection_aabb_max_x;
    float floor_detection_aabb_min_y;
    float floor_detection_aabb_max_y;
} [[color("AAAAAA")]];

// -----------------------------------------------------------------------------
// Components (Controls–Fossil)
// -----------------------------------------------------------------------------
// =============================================================================
// ControlsComponent (0x1D0 / 464 bytes)
// =============================================================================
struct ControlsComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    bool mButtonDownFire;            padding[3];
    s32  mButtonFrameFire;
    s32  mButtonLastFrameFire;
    bool mButtonDownFire2;           padding[3];
    s32  mButtonFrameFire2;
    bool mButtonDownAction;          padding[3];
    s32  mButtonFrameAction;
    bool mButtonDownThrow;           padding[3];
    s32  mButtonFrameThrow;
    bool mButtonDownInteract;        padding[3];
    s32  mButtonFrameInteract;
    bool mButtonDownLeft;            padding[3];
    s32  mButtonFrameLeft;
    bool mButtonDownRight;           padding[3];
    s32  mButtonFrameRight;
    bool mButtonDownUp;              padding[3];
    s32  mButtonFrameUp;
    bool mButtonDownDown;            padding[3];
    s32  mButtonFrameDown;
    bool mButtonDownJump;            padding[3];
    s32  mButtonFrameJump;
    bool mButtonDownRun;             padding[3];
    s32  mButtonFrameRun;
    bool mButtonDownFly;             padding[3];
    s32  mButtonFrameFly;
    bool mButtonDownDig;             padding[3];
    s32  mButtonFrameDig;
    bool mButtonDownChangeItemR;     padding[3];
    s32  mButtonFrameChangeItemR;
    s32  mButtonCountChangeItemR;
    bool mButtonDownChangeItemL;     padding[3];
    s32  mButtonFrameChangeItemL;
    s32  mButtonCountChangeItemL;
    bool mButtonDownInventory;       padding[3];
    s32  mButtonFrameInventory;
    bool mButtonDownHolsterItem;     padding[3];
    s32  mButtonFrameHolsterItem;
    bool mButtonDownDropItem;        padding[3];
    s32  mButtonFrameDropItem;
    bool mButtonDownKick;            padding[3];
    s32  mButtonFrameKick;
    bool mButtonDownEat;             padding[3];
    s32  mButtonFrameEat;
    bool mButtonDownLeftClick;       padding[3];
    s32  mButtonFrameLeftClick;
    bool mButtonDownRightClick;      padding[3];     // 252..255
    bool mButtonFrameRightClick;     padding[47];    // 256, gap to 304
    float mAimingVector_x;
    float mAimingVector_y;
    float mAimingVectorNormalized_x;
    float mAimingVectorNormalized_y;
    float mAimingVectorNonZeroLatest_x;
    float mAimingVectorNonZeroLatest_y;
    float mGamepadAimingVectorRaw_x;
    float mGamepadAimingVectorRaw_y;
    float mJumpVelocity_x;
    float mJumpVelocity_y;
    float mMousePosition_x;
    float mMousePosition_y;
    float mMousePositionRaw_x;
    float mMousePositionRaw_y;
    float mMousePositionRawPrev_x;
    float mMousePositionRawPrev_y;
    Vec2  mMouseDelta;                               // 368
    bool  mGamepadIndirectAiming;    padding[7];     // 376, gap to 384
    float mGamePadCursorInWorld_x;
    float mGamePadCursorInWorld_y;   padding[40];    // 392, gap to 432
    bool  polymorph_hax;             padding[3];
    s32   polymorph_next_attack_frame;
    s32   input_latency_frames;      padding[8];     // 444, gap to 452
    bool  enabled;
    bool  gamepad_indirect_aiming_enabled;
    bool  gamepad_fire_on_thumbstick_extend;  padding[1];  // 454, gap to 456
    float gamepad_fire_on_thumbstick_extend_threshold;     // 456..459
    padding[4];                                      // tail to 464
} [[color("AAAAAA")]];

// =============================================================================
// CrawlerAnimalComponent (0xF8 / 248 bytes)
// =============================================================================
struct CrawlerAnimalComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    float ray_length;
    s32   ray_count;
    float gravity;
    float terminal_velocity;
    float speed;
    s32   give_up_area_radius;
    s32   give_up_time;
    float attack_from_ceiling_check_ray_length;
    s32   attack_from_ceiling_check_every_n_frames;
    float collision_damage;
    float collision_damage_radius;
    s32   collision_damage_frames_between;
    bool  animate;
    bool  mDir;                      padding[2];     // 121, gap to 124
    s32   mFrameNextGiveUp;
    s32   mFrameNextDamage;
    s32   mFrameNextAttackFromCeilingCheck;
    Vec2  mMin;
    Vec2  mMax;
    Vec2  mPrevNonSnappedPosition;
    padding[72] [[name("field_0xa0_block"), comment("18x undefined4, 0xa0..0xe7")]];  // 160..232
    bool  mPrevFalling;
    bool  mIsInitialized;            padding[2];     // 233, gap to 236
    float mVelocityY;
    float mAngle;
    float mMovementStepAccumulator;                  // 244..248
} [[color("FF4444")]];

// =============================================================================
// DamageNearbyEntitiesComponent (0xA4 / 164 bytes)
// =============================================================================
struct DamageNearbyEntitiesComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    float radius;
    float damage_min;
    float damage_max;
    float target_vec_max_len;
    float knockback_multiplier;
    s32   time_between_damaging;
    padding[4] [[name("field_0x60")]];
    MsvcString damage_description;                   // 100
    MsvcString target_tag;                           // 124
    padding[4] [[name("field_0x94")]];
    Vec2  mVelocity;                                 // 152
    s32   mNextDamageFrame;                          // 160..164
} [[color("FF4444")]];

// =============================================================================
// DebugSpatialVisualizerComponent (0x5C / 92 bytes)
// =============================================================================
struct DebugSpatialVisualizerComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    float min_x;
    float min_y;
    float max_x;
    float max_y;
    s32   color;                                     // 88..92
} [[color("AAAAAA")]];

// =============================================================================
// ElectricChargeComponent (0x60 / 96 bytes)
// =============================================================================
struct ElectricChargeComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    s32   charge_time_frames;
    float fx_velocity_max;
    s32   electricity_emission_interval_frames;
    s32   fx_emission_interval_min;
    s32   fx_emission_interval_max;
    s32   charge;                                    // 92..96
} [[color("AAAAAA")]];

// =============================================================================
// ElectricityComponent (0x67 / 103 bytes)
// =============================================================================
struct ElectricityComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    padding[30];                                     // 72, gap to 102
    u8 unknown_66;                                   // 102..103
} [[color("AAAAAA")]];

// =============================================================================
// ElectricityReceiverComponent (0x6C / 108 bytes)
// =============================================================================
struct ElectricityReceiverComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    s32 offset_x;
    s32 offset_y;
    s32 radius;
    s32 active_time_frames;
    s32 switch_on_msg_interval_frames;
    s32 electrified_msg_interval_frames;
    s32 mLastFrameElectrified;
    s32 mNextElectrifiedMsgFrame;
    s32 mNextSwitchOnMsgFrame;                       // 104..108
} [[color("AAAAAA")]];

// =============================================================================
// ElectricitySourceComponent (0x54 / 84 bytes)
// =============================================================================
struct ElectricitySourceComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    s32 radius;
    s32 emission_interval_frames;
    s32 mNextFrameEmitElectricity;                   // 80..84
} [[color("AAAAAA")]];

// =============================================================================
// EnergyShieldComponent (0x6C / 108 bytes)
// =============================================================================
struct EnergyShieldComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    float radius;
    float damage_multiplier;
    float max_energy;
    float energy_required_to_shield;
    float recharge_speed;
    float sector_degrees;
    float energy;
    Vec2  mPrevPosition;                             // 100..108
} [[color("FF4444")]];

// =============================================================================
// ExplodeOnDamageComponent (0x1CD / 461 bytes)
// =============================================================================
struct ExplodeOnDamageComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    padding[388] [[name("body"), comment("undefined block 0x48..0x1cc, incl ConfigExplosion etc.")]];  // 72..460
    u8 unknown_1cc;                                  // 460..461
} [[color("FF4444")]];

// =============================================================================
// ExplosionComponent (0x1D0 / 464 bytes)
// =============================================================================
struct ExplosionComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    padding[4] [[name("field_0x48")]];
    padding[372] [[name("body"), comment("undefined block 0x4c..0x1bf, incl ConfigExplosion etc.")]];  // 76..448
    s32  timeout_frames;
    s32  timeout_frames_random;
    bool kill_entity;                padding[3];     // 456, gap to 460
    s32  mTimerTriggerFrame;                         // 460..464
} [[color("AAAAAA")]];

// =============================================================================
// FishAIComponent (0x74 / 116 bytes)
// =============================================================================
struct FishAIComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    s32   direction;
    float speed;
    Vec2  aabb_min;
    Vec2  aabb_max;
    Vec2  velocity;
    s32   stuck_counter;
    Vec2  mLastCheckPos;                             // 108..116
} [[color("AAAAAA")]];

// =============================================================================
// FlyingComponent (0x5C / 92 bytes)
// =============================================================================
struct FlyingComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    s32   type;
    float perlin_freq;
    float perlin_time_freq;
    float perlin_wind_x;
    float perlin_wind_y;                             // 88..92
} [[color("AAAAAA")]];

// =============================================================================
// FossilComponent (0x48 / 72 bytes)
// NB: atypical — fields begin right after the vtable (no 72-byte ComponentHeader).
// Modeled literally from the ghidra layout.
// =============================================================================
struct FossilComponent {
    u32 vftable;
    s32 count_min;
    s32 count_max;
    s32 seed;
    MsvcString material;
    MsvcString image_file;
    s32 material_id;
    u32 field_0x44;
} [[color("AAAAAA")]];

// -----------------------------------------------------------------------------
// Components (GameLog–InheritTransform)
// -----------------------------------------------------------------------------
// =============================================================================
// GameLogComponent (0x50 bytes)
// =============================================================================
struct GameLogComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    bool report_death;
    bool report_damage;
    bool report_new_biomes;
    padding[1];
    s32 mNewBiomeCheckFrame;
    // total 0x50
} [[color("AAAAAA")]];

// =============================================================================
// GameStatsComponent (0x9C bytes)
// =============================================================================
struct GameStatsComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    MsvcString name;
    MsvcString stats_filename;
    bool is_player;
    padding[3];
    MsvcString extra_death_msg;
    bool dont_do_logplayerkill;
    padding[3];
    s32 player_polymorph_count;
    // total 0x9C
} [[color("AAAAAA")]];

// =============================================================================
// GenomeDataComponent (0x70 bytes)
// =============================================================================
struct GenomeDataComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    LensValueInt herd_id;
    bool is_predator;
    padding[3];
    float food_chain_rank;
    u32 friend_thundermage [[comment("Entity* pointer")]];
    padding[4];
    u32 friend_firemage [[comment("Entity* pointer")]];
    bool berserk_dont_attack_friends;
    padding[3];
    // total 0x70
} [[color("AAAAAA")]];

// =============================================================================
// GhostComponent (0xA0 bytes)
// =============================================================================
struct GhostComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    float speed;
    Vec2 velocity;
    s32 new_hunt_target_check_every;
    float hunt_box_radius;
    float aggressiveness;
    s32 mEntityHome;
    float max_distance_from_home;
    bool die_if_no_home;
    padding[3];
    s32 mFramesWithoutHome;
    MsvcString target_tag;
    Vec2 mTargetPosition;
    s32 mTargetEntityId;
    Vec2 mRandomTarget;
    s32 mNextTargetCheckFrame;
    // total 0xA0
} [[color("AAAAAA")]];

// =============================================================================
// GunComponent (0x50 bytes)
// =============================================================================
struct GunComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    u32 mLuaManager [[comment("LuaManager* pointer")]];
    padding[4];
    // total 0x50
} [[color("AAAAAA")]];

// =============================================================================
// HitEffectComponent (0x70 bytes)
// =============================================================================
struct HitEffectComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    u32 unknown_48;
    u32 unknown_4c;
    u32 unknown_50;
    u32 unknown_54;
    u8  unknown_58;
    padding[15];
    u32 unknown_68;
    u32 unknown_6c;
    // total 0x70
} [[color("AAAAAA")]];

// =============================================================================
// HomingComponent (0x7D bytes)
// =============================================================================
struct HomingComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    MsvcString target_tag;
    bool target_who_shot;
    padding[3];
    float detect_distance;
    float homing_velocity_multiplier;
    float homing_targeting_coeff;
    bool just_rotate_towards_target;
    padding[3];
    float max_turn_rate;
    s32 predefined_target;
    bool look_for_root_entities_only;
    // total 0x7D
} [[color("AAAAAA")]];

// =============================================================================
// HotspotComponent (0x6C bytes)
// =============================================================================
struct HotspotComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    Vec2 offset;
    bool transform_with_scale;
    padding[3];
    MsvcString sprite_hotspot_name;
    // total 0x6C
} [[color("AAAAAA")]];

// =============================================================================
// IKLimbAttackerComponent (0x84 bytes)
// =============================================================================
struct IKLimbAttackerComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    float radius;
    float leg_velocity_coeff;
    float targeting_radius;
    bool targeting_raytrace;
    padding[3];
    MsvcString target_entities_with_tag;
    Vec2 mTarget;
    s32 mTargetEntity;
    u32 unknown_7c;
    float mStateTimer;
    // total 0x84
} [[color("AAAAAA")]];

// =============================================================================
// IKLimbComponent (0x84 bytes)
// =============================================================================
struct IKLimbComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    float length;
    float thigh_extra_lenght;
    Vec2 end_position;
    float mJointSideInterpolation;
    Vec2 mJointWorldPos;
    Vec2 mPart0PrevPos;
    float mPart0PrevRotation;
    Vec2 mPart1PrevPos;
    float mPart1PrevRotation;
    Vec2 mEndPrevPos;
    // total 0x84
} [[color("AAAAAA")]];

// =============================================================================
// IKLimbsAnimatorComponent (0x251 bytes)
// =============================================================================
struct IKLimbsAnimatorComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    u32 unknown_48;
    u32 unknown_4c;
    u32 unknown_50;
    u8  unknown_54;
    padding[3];
    u32 unknown_58;
    u32 unknown_5c;
    u32 unknown_60;
    u8  unknown_64;
    padding[3];
    u32 unknown_68;
    u32 unknown_6c;
    padding[480];
    u8  unknown_250;
    // total 0x251
} [[color("AAAAAA")]];

// =============================================================================
// IKLimbWalkerComponent (0x80 bytes)
// =============================================================================
struct IKLimbWalkerComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    float ground_attachment_min_spread;
    s32 ground_attachment_max_tries;
    float ground_attachment_max_angle;
    float ground_attachment_ray_length_coeff;
    float leg_velocity_coeff;
    bool affect_flying;
    padding[3];
    s32 ray_skip_material;
    Vec2 mTarget;
    Vec2 mPrevTarget;
    Vec2 mPrevCenterPosition;
    s32 mState;
    // total 0x80
} [[color("AAAAAA")]];

// =============================================================================
// IngestionComponent (0xAC bytes)
// =============================================================================
struct IngestionComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    u32 unknown_48;
    u32 unknown_4c;
    u32 unknown_50;
    u32 unknown_54;
    s32 ingestion_cooldown_delay_frames;
    s32 ingestion_reduce_every_n_frame;
    float overingestion_damage;
    float blood_healing_speed;
    MsvcString ingestion_satiation_material_tag;
    s32 m_ingestion_cooldown_frames;
    s32 m_next_overeating_msg_frame;
    MsvcString m_ingestion_satiation_material_tag_cached;
    u32 unknown_a0;
    u32 unknown_a4;
    s32 m_damage_effect_lifetime;
    // total 0xAC
} [[color("FF4444")]];

// =============================================================================
// InheritTransformComponent (0x90 bytes)
// =============================================================================
struct InheritTransformComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    bool use_root_parent;
    bool only_position;
    padding[2];
    MsvcString parent_hotspot_tag;
    s32 parent_sprite_id;
    bool always_use_immediate_parent_rotation;
    bool rotate_based_on_x_scale;
    padding[2];
    float transform_pos_x;
    float transform_pos_y;
    float transform_rot;
    float transform_scale_x;
    float transform_scale_y;
    padding[12];
    s32 mUpdateFrame;
    // total 0x90
} [[color("AAAAAA")]];

// -----------------------------------------------------------------------------
// Components (Interactable–MagicConvert)
// -----------------------------------------------------------------------------
// =============================================================================
// InteractableComponent (0x80 / 128 bytes)
// =============================================================================
struct InteractableComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    float radius;
    MsvcString ui_text;
    MsvcString name;
    s32 exclusivity_group;
} [[color("AAAAAA")]];

// =============================================================================
// InventoryComponent (0x9C / 156 bytes)
// =============================================================================
struct InventoryComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    s32 ui_container_type;
    float ui_container_size_x;
    float ui_container_size_y;
    float ui_element_size_x;
    float ui_element_size_y;
    float ui_position_on_screen_x;
    float ui_position_on_screen_y;
    u32 ui_element_sprite [[comment("Sprite* pointer")]];
    padding[124 - 104];
    u32 actions [[comment("pointer")]];
    padding[148 - 128];
    u32 update_listener [[comment("void* pointer")]];
    u32 items_begin [[comment("pointer")]];
} [[color("AAAAAA")]];

// =============================================================================
// InventoryGuiComponent (0x7D / 125 bytes)
// =============================================================================
struct InventoryGuiComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    bool has_opened_inventory_edit;
    padding[3];
    s32 wallet_money_target [[comment("money")]];
    u32 unknown_50;
    s32 mLastFrameInteracted;
    s32 mLastFrameActionsVisible;
    u32 unknown_5c;
    bool mActive;
    padding[3];
    float mAlpha;
    float mBackgroundOverlayAlpha;
    s32 mFrameShake_ReloadBar;
    s32 mFrameShake_ManaBar;
    s32 mFrameShake_FlyBar;
    s32 mFrameShake_FireRateWaitBar;
    bool mDisplayFireRateWaitBar;
} [[color("FFD700")]];

// =============================================================================
// ItemAIKnowledgeComponent (0x53 / 83 bytes)
// =============================================================================
struct ItemAIKnowledgeComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47 (incl field_0xc)
    // +0x48
    bool is_ranged_weapon;
    bool is_throwable_weapon;
    bool is_melee_weapon;
    bool is_self_healing;
    bool is_other_healing;
    bool is_self_buffing;
    bool is_other_buffing;
    bool is_weapon;
    bool is_known;
    bool is_safe;
    bool is_consumed;
} [[color("AAAAAA")]];

// =============================================================================
// ItemChestComponent (0xA4 / 164 bytes)
// =============================================================================
struct ItemChestComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    u32 unknown_48;
    u32 unknown_4c;
    u32 unknown_50;
    u8  unknown_54;
    padding[3];
    u8  unknown_58;
    padding[15];
    u32 unknown_68;
    u32 unknown_6c;
    u8  unknown_70;
    padding[15];
    u32 unknown_80;
    u32 unknown_84;
    u8  unknown_88;
    padding[15];
    u32 unknown_98;
    u32 unknown_9c;
    u32 unknown_a0;
} [[color("AAAAAA")]];

// =============================================================================
// ItemPickUpperComponent (0x5C / 92 bytes)
// =============================================================================
struct ItemPickUpperComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    bool is_in_npc;
    bool pick_up_any_item_buggy;
    bool is_immune_to_kicks;
    padding[1];
    s32 only_pick_this_entity;
    bool drop_items_on_death;
    padding[3];
    Vec2 mLatestItemOverlapInfoBoxPosition;
} [[color("AAAAAA")]];

// =============================================================================
// ItemStashComponent (0x58 / 88 bytes)
// =============================================================================
struct ItemStashComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    s32 throw_openable_cooldown_frames;
    bool init_children;
    padding[3];
    s32 mNextFrameOpenable;
    s32 mFrameOpened;
} [[color("AAAAAA")]];

// =============================================================================
// KickComponent (0x9C / 156 bytes)
// =============================================================================
struct KickComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    u8  unknown_48;
    padding[3];
    u32 unknown_4c;
    u32 unknown_50;
    u32 unknown_54;
    u32 unknown_58;
    u32 unknown_5c;
    u32 unknown_60;
    padding[4];
    u32 unknown_68;
    u32 unknown_6c;
    u32 unknown_70;
    u32 unknown_74;
    u32 unknown_78;
    u32 unknown_7c;
    u32 unknown_80;
    u8  unknown_84;
    padding[15];
    u32 unknown_94;
    u32 unknown_98;
} [[color("AAAAAA")]];

// =============================================================================
// LaserEmitterComponent (0x88 / 136 bytes)
// =============================================================================
struct LaserEmitterComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    padding[124 - 72];
    bool is_emitting;
    padding[3];
    s32 emit_until_frame;
    float laser_angle_add_rad;
} [[color("AAAAAA")]];

// =============================================================================
// LifetimeComponent (0x6C / 108 bytes)
// =============================================================================
struct LifetimeComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    s32 creation_frame;
    s32 lifetime;
    ValueRange randomize_lifetime;
    bool fade_sprites;
    bool kill_parent;
    bool kill_all_parents;
    padding[1];
    s32 kill_frame;
    bool serialize_duration;
    padding[3];
    s32 kill_frame_serialized;
    s32 creation_frame_serialized;
} [[color("AAAAAA")]];

// =============================================================================
// LightningComponent (0x1EC / 492 bytes)
// =============================================================================
struct LightningComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    padding[444 - 72];
    MsvcString sprite_lightning_file;
    bool is_projectile;
    padding[3];
    s32 explosion_type;
    Vec2 mExPosition;
    s32 mArcTarget;
    s32 arc_lifetime;
} [[color("AAAAAA")]];

// =============================================================================
// LimbBossComponent (0x58 / 88 bytes)
// =============================================================================
struct LimbBossComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    s32 state;
    s32 mStatePrev;
    float mMoveToPositionX;
    float mMoveToPositionY;
} [[color("AAAAAA")]];

// =============================================================================
// LooseGroundComponent (0x8C / 140 bytes)
// =============================================================================
struct LooseGroundComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    float probability;
    s32 max_durability;
    float max_distance;
    float max_angle;
    s32 min_radius;
    s32 max_radius;
    float chunk_probability;
    padding[4];
    s32 chunk_count;
    s32 chunk_material;
    s32 mChunkCount;
    MsvcString collapse_images;
} [[color("AAAAAA")]];

// =============================================================================
// MagicConvertMaterialComponent (0xE4 / 228 bytes)
// =============================================================================
struct MagicConvertMaterialComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    u32 unknown_48;
    u32 unknown_4c;
    u8  unknown_50;
    padding[3];
    u32 unknown_54;
    u32 unknown_58;
    u8  unknown_5c;
    padding[15];
    u32 unknown_6c;
    u32 unknown_70;
    padding[36];
    u8  unknown_98;
    padding[15];
    u32 unknown_a8;
    u32 unknown_ac;
    u8  unknown_b0;
    padding[15];
    u32 unknown_c0;
    u32 unknown_c4;
    padding[4];
    u32 unknown_cc;
    u32 unknown_d0;
    u32 unknown_d4;
    u32 unknown_d8;
    u32 unknown_dc;
    u32 unknown_e0;
} [[color("AAAAAA")]];

// -----------------------------------------------------------------------------
// Components (MaterialAreaChecker–PhysicsKeepInWorld)
// -----------------------------------------------------------------------------
// =============================================================================
// MaterialAreaCheckerComponent (0x78 / 120 bytes)
// =============================================================================
struct MaterialAreaCheckerComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48: all fields undefined in Ghidra
    padding[120 - 72];
} [[color("AAAAAA")]];

// =============================================================================
// MaterialComponent (0x47 / 71 bytes)
// NB: 71 bytes total is smaller than the 72-byte ComponentHeader, so it cannot
// embed one. Only the vtable and a trailing byte are known in ghidra.
// =============================================================================
struct MaterialComponent {
    u32 vftable;
    padding[66];                 // 0x04..0x45 (undefined in ghidra)
    u8 field_0x46;               // 0x46
} [[color("AAAAAA")]];

// =============================================================================
// MaterialSeaSpawnerComponent (0x80 / 128 bytes)
// =============================================================================
struct MaterialSeaSpawnerComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    s32 material;
    u32 unknown_4c;
    u32 unknown_50;
    u32 unknown_54;
    u32 unknown_58;
    s32 speed;
    float sine_wavelength;
    float sine_amplitude;
    u32 unknown_68;
    u32 unknown_6c;
    u32 unknown_70;
    u32 unknown_74;
    s32 m_position;
    s32 frames_run;
} [[color("AAAAAA")]];

// =============================================================================
// MaterialSuckerComponent (0x94 / 148 bytes)
// =============================================================================
struct MaterialSuckerComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    s32 material_type;
    s32 barrel_size;
    s32 num_cells_sucked_per_frame;
    bool set_projectile_to_liquid;
    padding[3];
    s32 last_material_id;
    bool suck_gold;
    bool suck_health;
    bool suck_static_materials;
    padding[1];
    MsvcString suck_tag;                         // std::string, 24B
    IAABB randomized_position;                   // 16B
    s32 mAmountUsed;
    s32 mGoldAccumulator;
    s32 mLastFramePickedGold;
} [[color("FFD700")]];

// =============================================================================
// MoveToSurfaceOnCreateComponent (0x5c / 92 bytes)
// =============================================================================
struct MoveToSurfaceOnCreateComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    padding[4];                                  // 0x48..0x4b
    float lookup_radius;
    float offset_from_surface;
    s32 ray_count;
    float verlet_min_joint_distance;
} [[color("AAAAAA")]];

// =============================================================================
// NinjaRopeComponent (0x5c / 92 bytes)
// =============================================================================
struct NinjaRopeComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    float max_length;
    float mLength;
    u32 unknown_50;
    u32 unknown_54;
    u32 unknown_58;
} [[color("AAAAAA")]];

// =============================================================================
// OrbComponent (0x4c / 76 bytes)
// =============================================================================
struct OrbComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    s32 orb_id;
} [[color("AAAAAA")]];

// =============================================================================
// ParticleEmitterComponent (0x198 / 408 bytes)
// =============================================================================
struct ParticleEmitterComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    MsvcString emitted_material_name;            // std::string, 24B
    bool create_real_particles;
    bool emit_real_particles;
    bool emit_cosmetic_particles;
    bool cosmetic_force_create;
    bool render_back;
    bool render_ultrabright;
    bool collide_with_grid;
    bool collide_with_gas_and_fire;
    bool particle_single_width;
    bool emit_only_if_there_is_space;
    padding[2];
    s32 emitter_lifetime_frames;
    bool fire_cells_dont_ignite_damagemodel;
    bool color_is_based_on_pos;
    padding[2];
    u32 color;
    float custom_alpha;
    Vec2 offset;                                 // 8B
    float x_pos_offset_min;
    float y_pos_offset_min;
    float x_pos_offset_max;
    float y_pos_offset_max;
    ValueRange area_circle_radius;               // 8B
    float area_circle_sector_degrees;
    float x_vel_min;
    float x_vel_max;
    float y_vel_min;
    float y_vel_max;
    float direction_random_deg;
    Vec2 gravity;                                // 8B
    bool velocity_always_away_from_center;
    padding[3];
    float lifetime_min;
    float lifetime_max;
    float airflow_force;
    float airflow_time;
    float airflow_scale;
    float friction;
    float attractor_force;
    float count_min;
    padding[8];                                  // 0xe0..0xe7
    LensValueInt count_max;                       // 12B
    s32 emission_interval_min_frames;
    s32 emission_interval_max_frames;
    float emission_chance;
    s32 custom_style;
    s32 delay_frames;
    bool is_emitting;
    bool use_material_inventory;
    bool is_trail;
    padding[1];
    float trail_gap;
    bool render_on_grid;
    bool fade_based_on_lifetime;
    bool draw_as_long;
    padding[1];
    float b2_force;
    bool set_magic_creation;
    padding[3];
    u32 image_animation_file [[comment("pointer")]];
    padding[20];                                 // 0x120..0x133
    MsvcString image_animation_colors_file;      // std::string, 24B
    float image_animation_speed;
    bool image_animation_loop;
    padding[3];
    float image_animation_phase;
    float image_animation_emission_probability;
    bool image_animation_raytrace_from_center;
    bool image_animation_use_entity_rotation;
    bool ignore_transform_updated_msg;
    padding[1];
    Vec2 mExPosition;                            // 8B
    s32 mMaterialInventoryMax;
    LensValueInt m_material_id;                   // 12B
    s32 m_next_emit_frame;
    bool m_has_emitted;
    padding[3];
    Vec2i m_last_emit_position;                  // 8B
    u32 m_cached_image_animation [[comment("pointer")]];
    float m_image_based_animation_time;
    u32 m_collision_angles [[comment("pointer")]];
    s16 m_particle_attractor_id;
    padding[2];
} [[color("AAAAAA")]];

// =============================================================================
// PathFindingComponent (0x1c4 / 452 bytes)
// =============================================================================
struct PathFindingComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48: all fields undefined in Ghidra
    padding[452 - 72];
} [[color("AAAAAA")]];

// =============================================================================
// PathFindingGridMarkerComponent (0x60 / 96 bytes)
// =============================================================================
struct PathFindingGridMarkerComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48: all fields undefined in Ghidra
    padding[96 - 72];
} [[color("AAAAAA")]];

// =============================================================================
// PhysicsBody2Component (0xa0 / 160 bytes)
// =============================================================================
struct PhysicsBody2Component {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    padding[16];                                 // 0x48..0x57
    float linear_damping;
    float angular_damping;
    bool allow_sleep;
    bool fixed_rotation;
    bool is_bullet;
    padding[5];
    bool hax_fix_going_through_ground;
    bool hax_fix_going_through_sand;
    bool hax_wait_till_pixel_scenes_loaded;
    bool go_through_sand;
    bool auto_clean;
    bool force_add_update_areas;
    bool update_entity_transform;
    bool kill_entity_if_body_destroyed;
    bool kill_entity_after_initialized;
    bool manual_init;
    bool destroy_body_if_entity_destroyed;
    padding[1];
    float root_offset_x;
    float root_offset_y;
    float init_offset_x;
    float init_offset_y;
    bool mActiveState;
    padding[3];
    Vec2 mLocalPosition;                         // 8B
    s32 mPixelCountOrig;
    bool mInitialized;
    padding[3];
    s32 mPixelCount;
    bool mRefreshed;
    padding[3];
} [[color("AAAAAA")]];

// =============================================================================
// PhysicsBodyComponent (0xac / 172 bytes)
// =============================================================================
struct PhysicsBodyComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    u32 mBody [[comment("b2Body* pointer")]];
    padding[4];
    s32 mBodyId;
    padding[4];
    bool is_external;
    bool hax_fix_going_through_ground;
    bool hax_fix_going_through_sand;
    bool hax_wait_till_pixel_scenes_loaded;
    padding[4];
    bool is_enabled;
    padding[3];
    float linear_damping;
    float angular_damping;
    bool allow_sleep;
    bool fixed_rotation;
    padding[2];
    float buoyancy;
    float gravity_scale_if_has_no_image_shapes;
    bool is_bullet;
    bool is_static;
    bool is_kinematic;
    bool is_character;
    bool go_through_sand;
    bool gridworld_box2d;
    bool auto_clean;
    bool on_death_leave_physics_body;
    bool on_death_really_leave_body;
    bool update_entity_transform;
    bool force_add_update_areas;
    bool kills_entity;
    bool projectiles_rotate_toward_velocity;
    padding[3];
    Vec2 initial_velocity;                       // 8B
    bool randomize_init_velocity;
    bool mActiveState;
    padding[2];
    s32 mPixelCount;
    float mLocalPosition_x;
    float mLocalPosition_y;
    padding[8];                                  // 0xa0..0xa7
    bool mRefreshed;
    padding[3];
} [[color("AAAAAA")]];

// =============================================================================
// PhysicsImageShapeComponent (0x53 / 83 bytes)
// =============================================================================
// NB: Ghidra also lists fields at offsets 0x05,0x09,0x0d,0x25 which fall inside
// the 0x48-byte component header region; per spec those are header-covered and
// not modeled individually here.
struct PhysicsImageShapeComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    bool is_root;
    padding[3];
    s32 body_id;
    bool use_sprite;
    bool is_circle;
    bool centered;
} [[color("AAAAAA")]];

// =============================================================================
// PhysicsJoint2Component (0x80 / 128 bytes)
// =============================================================================
struct PhysicsJoint2Component {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    padding[4];                                  // 0x48..0x4b
    float break_force;
    float break_distance;
    bool break_on_body_modified;
    padding[3];
    float break_on_shear_angle_deg;
    padding[4];
    s32 body1_id;
    padding[4];
    float offset_x;
    float offset_y;
    float ray_x;
    float ray_y;
    float surface_attachment_offset_x;
    float surface_attachment_offset_y;
} [[color("AAAAAA")]];

// =============================================================================
// PhysicsJoint2MutatorComponent (0x4b / 75 bytes)
// =============================================================================
// NB: Ghidra also lists fields at offsets 0x05,0x15,0x19 inside the header
// region; per spec those are header-covered and not modeled individually.
struct PhysicsJoint2MutatorComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    padding[2];                                  // 0x48..0x49
    bool destroy;
} [[color("AAAAAA")]];

// =============================================================================
// PhysicsJointComponent (0x4b / 75 bytes)
// =============================================================================
struct PhysicsJointComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    padding[2];                                  // 0x48..0x49
    u8 unknown_4a;
} [[color("AAAAAA")]];

// =============================================================================
// PhysicsKeepInWorldComponent (0x4b / 75 bytes)
// =============================================================================
// NB: Ghidra also lists a float at offset 0x09 (mExRotation) inside the header
// region; per spec that is header-covered and not modeled individually.
struct PhysicsKeepInWorldComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    bool check_whole_aabb;
    bool predict_aabb;
    bool keep_at_last_valid_pos;
} [[color("AAAAAA")]];

// -----------------------------------------------------------------------------
// Components (PhysicsPickUp–SpriteParticleEmitter)
// -----------------------------------------------------------------------------
// =============================================================================
// PhysicsPickUpComponent (0x90 / 144 bytes)
// =============================================================================
struct PhysicsPickUpComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    padding[104 - 72];                          // unmodeled header tail / base
    Vec2 original_left_joint_pos;               // +0x68
    Vec2 original_right_joint_pos;              // +0x70
    float pick_up_strength;                     // +0x78
    bool isBroken;                              // +0x7c
    padding[128 - 125];
    Vec2 leftJointPos;                          // +0x80
    Vec2 rightJointPos;                         // +0x88
} [[color("AAAAAA")]];

// =============================================================================
// PhysicsRagdollComponent (0x80 / 128 bytes)
// =============================================================================
struct PhysicsRagdollComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    MsvcString filename;                        // +0x48
    MsvcString filenames;                       // +0x60
    float offset_x;                             // +0x78
    float offset_y;                             // +0x7c
} [[color("AAAAAA")]];

// =============================================================================
// PhysicsThrowableComponent (0x74 / 116 bytes)
// =============================================================================
struct PhysicsThrowableComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    float throw_force_coeff;                    // +0x48
    float max_throw_speed;                      // +0x4c
    float min_torque;                           // +0x50
    float max_torque;                           // +0x54
    float tip_check_offset_min;                 // +0x58
    float tip_check_offset_max;                 // +0x5c
    float tip_check_random_rotation_deg;        // +0x60
    padding[104 - 100];
    bool attach_to_surfaces_knife_style;        // +0x68
    padding[108 - 105];
    s32 hp;                                     // +0x6c
    bool mHasJoint;                             // +0x70
    padding[116 - 113];
} [[color("FF4444")]];

// =============================================================================
// PixelSceneComponent (0x9e / 158 bytes)
// =============================================================================
struct PixelSceneComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    padding[158 - 72];                          // all fields undefined (field_0x48..0x9c)
} [[color("AAAAAA")]];

// =============================================================================
// PixelSpriteComponent (0x88 / 136 bytes)
// =============================================================================
struct PixelSpriteComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    MsvcString image_file;                      // +0x48
    float anchor_x;                             // +0x60
    s32 anchor_y;                               // +0x64
    MsvcString material;                        // +0x68
    bool diggable;                              // +0x80
    bool clean_overlapping_pixels;              // +0x81
    bool kill_when_sprite_dies;                 // +0x82
    bool create_box2d_bodies;                   // +0x83
    u32 mPixelSprite [[comment("PixelSprite*")]]; // +0x84
} [[color("AAAAAA")]];

// =============================================================================
// PlatformShooterPlayerComponent (0x5f0 / 1520 bytes)
// =============================================================================
struct PlatformShooterPlayerComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    float aiming_reticle_distance_from_character; // +0x48
    float camera_max_distance_from_character;   // +0x4c
    float alcohol_drunken_speed;                // +0x50
    float blood_fungi_drunken_speed;            // +0x54
    float blood_worm_drunken_speed;             // +0x58
    float eating_area_min_x;                    // +0x5c
    float eating_area_min_y;                    // +0x60
    Vec2i eating_area_max;                      // +0x64
    s32 eating_cells_per_frame;                 // +0x6c
    float eating_probability;                   // +0x70
    s32 eating_delay_frames;                    // +0x74
    float stoned_speed;                         // +0x78
    bool center_camera_on_this_entity;          // +0x7c
    bool move_camera_with_aim;                  // +0x7d
    padding[128 - 126];
    float mSmoothedCameraPosition_x;            // +0x80
    float mSmoothedCameraPosition_y;            // +0x84
    Vec2 mSmoothedAimingVector;                 // +0x88
    float mCameraRecoil;                        // +0x90
    float mCameraRecoilTarget;                  // +0x94
    bool mCrouching;                            // +0x98
    padding[156 - 153];
    float mCameraDistanceLerped;                // +0x9c
    bool mRequireTriggerPull;                   // +0xa0
    padding[164 - 161];
    s32 mWarpDelay;                             // +0xa4
    s32 mItemTemporarilyHidden;                 // +0xa8
    Vec2 mDesiredCameraPos;                     // +0xac
    bool mHasGamepadControlsPrev;               // +0xb4
    bool mForceFireOnNextUpdate;                // +0xb5
    padding[184 - 182];
    float mFastMovementParticlesAlphaSmoothed;  // +0xb8
    padding[192 - 188];
    u32 mTeleBoltFramesDuringLastSecond_begin [[comment("vector begin")]]; // +0xc0
    u32 mTeleBoltFramesDuringLastSecond_end   [[comment("vector end")]];   // +0xc4
    float mCamCorrectionTeleSmoothed;           // +0xc8
    Vec2 mCamCorrectionGainSmoothed;            // +0xcc
    float mCameraErrorPrev_x;                   // +0xd4
    float mCameraErrorPrev_y;                   // +0xd8
    padding[1492 - 220];                        // large unmodeled region (+0xdc..+0x5d4)
    Vec2 mCamErrorAveraged;                     // +0x5d4
    bool mCamMovingFastPrev;                    // +0x5dc
    padding[1504 - 1501];
    s32 mCamFrameStartedMovingFast;             // +0x5e0
    s32 mCamFrameLastMovingFastExplosion;       // +0x5e4
    bool mCessationDo;                          // +0x5e8
    padding[1516 - 1513];
    s32 mCessationLifetime;                     // +0x5ec
} [[color("AAAAAA")]];

// =============================================================================
// PlayerCollisionComponent (0x60 / 96 bytes)
// =============================================================================
struct PlayerCollisionComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    s32 getting_crushed_threshold;              // +0x48
    s32 moving_up_before_getting_crushed_threshold; // +0x4c
    s32 getting_crushed_counter;                // +0x50
    s32 stuck_in_ground_counter;                // +0x54
    s32 DEBUG_stuck_in_static_ground;           // +0x58
    bool mCollidedHorizontally;                 // +0x5c
    padding[96 - 93];
} [[color("AAAAAA")]];

// =============================================================================
// PotionComponent (0x5b / 91 bytes)
// =============================================================================
struct PotionComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    padding[91 - 72];                           // only undefined field_0x5a present
} [[color("AAAAAA")]];

// =============================================================================
// PressurePlateComponent (0x68 / 104 bytes)
// =============================================================================
struct PressurePlateComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    s32 mNextFrame;                             // +0x48
    s32 check_every_x_frames;                   // +0x4c
    s32 state;                                  // +0x50
    Vec2 aabb_min;                              // +0x54
    Vec2 aabb_max;                              // +0x5c
    float material_percent;                     // +0x64
} [[color("AAAAAA")]];

// =============================================================================
// ProjectileComponent (0x650 / 1616 bytes)
// =============================================================================
struct ProjectileComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    s32 projectile_type;                        // +0x48
    s32 lifetime;                               // +0x4c
    s32 lifetime_randomness;                    // +0x50
    bool on_lifetime_out_explode;               // +0x54
    bool collide_with_world;                    // +0x55
    padding[88 - 86];
    padding[572] [[name("config"), comment("ConfigGunActionData (-BAD-), 572 bytes")]]; // +0x58
    float speed_min;                            // +0x294
    float speed_max;                            // +0x298
    float friction;                             // +0x29c
    float direction_random_rad;                 // +0x2a0
    float direction_nonrandom_rad;              // +0x2a4
    float lob_min;                              // +0x2a8
    float lob_max;                              // +0x2ac
    float camera_shake_when_shot;               // +0x2b0
    float shoot_light_flash_radius;             // +0x2b4
    float shoot_light_flash_r;                  // +0x2b8
    float shoot_light_flash_g;                  // +0x2bc
    u32 shoot_light_flash_b;                    // +0x2c0
    bool create_shell_casing;                   // +0x2c4
    padding[712 - 709];
    u32 shell_casing_material [[comment("char* material name")]]; // +0x2c8
    padding[736 - 716];
    Vec2 shell_casing_offset;                   // +0x2e0
    u32 muzzle_flash_file [[comment("char* file")]]; // +0x2e8
    padding[768 - 748];
    s32 bounces_left;                           // +0x300
    float bounce_energy;                        // +0x304
    bool bounce_always;                         // +0x308
    bool bounce_at_any_angle;                   // +0x309
    bool attach_to_parent_trigger;              // +0x30a
    padding[780 - 779];
    MsvcString bounce_fx_file;                  // +0x30c
    float angular_velocity;                     // +0x324
    bool velocity_sets_rotation;                // +0x328
    bool velocity_sets_scale;                   // +0x329
    padding[812 - 810];
    float velocity_sets_scale_coeff;            // +0x32c
    bool velocity_sets_y_flip;                  // +0x330
    padding[820 - 817];
    float velocity_updates_animation;           // +0x334
    float ground_penetration_coeff;             // +0x338
    s32 ground_penetration_max_durability_to_destroy; // +0x33c
    MsvcString go_through_this_material;        // +0x340
    bool do_moveto_update;                      // +0x358
    padding[860 - 857];
    s32 on_death_duplicate_remaining;           // +0x35c
    bool on_death_gfx_leave_sprite;             // +0x360
    bool on_death_explode;                      // +0x361
    bool on_death_emit_particle;                // +0x362
    padding[868 - 867];
    s32 on_death_emit_particle_count;           // +0x364
    bool die_on_liquid_collision;               // +0x368
    bool die_on_low_velocity;                   // +0x369
    padding[876 - 874];
    float die_on_low_velocity_limit;            // +0x36c
    MsvcString on_death_emit_particle_type;     // +0x370
    bool on_death_particle_check_concrete;      // +0x388
    bool ground_collision_fx;                   // +0x389
    bool explosion_dont_damage_shooter;         // +0x38a
    padding[908 - 907];
    padding[372] [[name("config_explosion"), comment("ConfigExplosion, 372 bytes")]]; // +0x38c
    float on_death_item_pickable_radius;        // +0x500
    bool penetrate_world;                       // +0x504
    padding[1288 - 1285];
    float penetrate_world_velocity_coeff;       // +0x508
    bool penetrate_entities;                    // +0x50c
    bool on_collision_die;                      // +0x50d
    bool on_collision_remove_projectile;        // +0x50e
    bool on_collision_spawn_entity;             // +0x50f
    MsvcString spawn_entity;                    // +0x510
    bool spawn_entity_is_projectile;            // +0x528
    padding[1324 - 1321];
    float physics_impulse_coeff;                // +0x52c
    s32 damage_every_x_frames;                  // +0x530
    bool damage_scaled_by_speed;                // +0x534
    padding[1336 - 1333];
    float damage_scale_max_speed;               // +0x538
    s32 ragdoll_fx_on_collision;                // +0x53c
    bool collide_with_entities;                 // +0x540
    padding[1348 - 1345];
    MsvcString collide_with_tag;                // +0x544
    MsvcString dont_collide_with_tag;           // +0x55c
    s32 collide_with_shooter_frames;            // +0x574
    bool friendly_fire;                         // +0x578
    padding[1404 - 1401];
    float damage;                               // +0x57c
    padding[64] [[name("damage_by_type"), comment("ConfigDamagesByType, 64 bytes")]]; // +0x580
    padding[16] [[name("damage_critical"), comment("ConfigDamageCritical, 16 bytes")]]; // +0x5c0
    float knockback_force;                      // +0x5d0
    float ragdoll_force_multiplier;             // +0x5d4
    float hit_particle_force_multiplier;        // +0x5d8
    float blood_count_multiplier;               // +0x5dc
    MsvcString damage_game_effect_entities;     // +0x5e0
    bool never_hit_player;                      // +0x5f8
    bool collect_materials_to_shooter;          // +0x5f9
    padding[1532 - 1530];
    s32 mWhoShot;                               // +0x5fc
    s32 mWhoShotEntityTypeID;                   // +0x600
    s32 mShooterHerdId;                         // +0x604
    s32 mStartingLifetime;                      // +0x608
    s32 mEntityThatShot;                        // +0x60c
    u32 mTriggers [[comment("vector ptr")]];    // +0x610
    padding[1588 - 1556];
    bool play_damage_sounds;                    // +0x634
    padding[1592 - 1589];
    s32 mLastFrameDamaged;                      // +0x638
    u32 mDamagedEntities_begin;                 // +0x63c
    u32 mDamagedEntities_end;                   // +0x640
    u32 mDamagedEntities_cap;                   // +0x644
    float mInitialSpeed;                        // +0x648
    padding[1616 - 1612];
} [[color("FF4444")]];

// =============================================================================
// ShotEffectComponent (0x68 / 104 bytes)
// =============================================================================
struct ShotEffectComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    u32 unknown_48;                             // +0x48 (undefined4)
    s32 condition_status;                       // +0x4c
    MsvcString extra_modifier;                  // +0x50
} [[color("AAAAAA")]];

// =============================================================================
// SineWaveComponent (0x54 / 84 bytes)
// =============================================================================
struct SineWaveComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    float sinewave_freq;                        // +0x48
    float sinewave_m;                           // +0x4c
    s32 lifetime;                               // +0x50
} [[color("AAAAAA")]];

// =============================================================================
// SpriteAnimatorComponent (0xa8 / 168 bytes)
// =============================================================================
struct SpriteAnimatorComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    MsvcString target_sprite_comp_name;         // +0x48
    bool rotate_to_surface_normal;              // +0x60
    padding[144 - 97];                          // field_0x64..0x8c (undefined)
    MsvcString mSendOnFinishedMessageName;      // +0x90
} [[color("AAAAAA")]];

// =============================================================================
// SpriteParticleEmitterComponent (0x170 / 368 bytes)
// =============================================================================
struct SpriteParticleEmitterComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    MsvcString sprite_file;                     // +0x48
    bool sprite_centered;                       // +0x60
    bool sprite_random_rotation;                // +0x61
    bool render_back;                           // +0x62
    padding[100 - 99];
    float delay;                                // +0x64
    float lifetime;                             // +0x68
    padding[148 - 108];                         // field_0x6c..0x90 (undefined)
    bool additive;                              // +0x94
    bool emissive;                              // +0x95
    padding[152 - 150];
    Vec2 velocity;                              // +0x98
    Vec2 gravity;                               // +0xa0
    float velocity_slowdown;                    // +0xa8
    float rotation;                             // +0xac
    float angular_velocity;                     // +0xb0
    bool use_velocity_as_rotation;              // +0xb4
    bool use_rotation_from_velocity_component;  // +0xb5
    bool use_rotation_from_entity;              // +0xb6
    padding[184 - 183];
    float entity_velocity_multiplier;           // +0xb8
    Vec2 scale;                                 // +0xbc
    Vec2 scale_velocity;                        // +0xc4
    float z_index;                              // +0xcc
    padding[232 - 208];                         // field_0xd0..0xe4 (undefined)
    bool randomize_position_inside_hitbox;      // +0xe8
    padding[236 - 233];
    padding[300 - 236];                         // field_0xec..0x128 (undefined)
    bool velocity_always_away_from_center;      // +0x12c
    padding[304 - 301];
    Vec2 expand_randomize_position;             // +0x130
    bool camera_bound;                          // +0x138
    padding[316 - 313];
    float camera_distance;                      // +0x13c
    bool is_emitting;                           // +0x140
    padding[324 - 321];
    s32 count_min;                              // +0x144
    s32 count_max;                              // +0x148
    s32 emission_interval_min_frames;           // +0x14c
    s32 emission_interval_max_frames;           // +0x150
    MsvcString entity_file;                     // +0x154
    s32 mNextEmitFrame;                         // +0x16c
} [[color("AAAAAA")]];

// -----------------------------------------------------------------------------
// Components (SpriteStains–WormPlayer)
// -----------------------------------------------------------------------------
// =============================================================================
// SpriteStainsComponent (0x68 / 104 bytes)
// =============================================================================
struct SpriteStainsComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    s32  sprite_id;                             // 0x48
    bool fade_stains_towards_srite_top;         // 0x4c
    padding[7];                                 // 0x4d..0x53
    u32  field_0x54;                            // 0x54
    u32  field_0x58;                            // 0x58
    u32  field_0x5c;                            // 0x5c
    u32  field_0x60;                            // 0x60
    u32  field_0x64;                            // 0x64
} [[color("AAAAAA")]];

// =============================================================================
// StatusEffectDataComponent (0xA1 / 161 bytes)
// =============================================================================
struct StatusEffectDataComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    padding[88]     [[name("fields_0x48"), comment("22x undefined4, 0x48..0x9f")]];
    u8   field_0xa0;                            // 0xa0
} [[color("AAAAAA")]];

// =============================================================================
// TelekinesisComponent (0x89 / 137 bytes)
// =============================================================================
struct TelekinesisComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    u32  field_0x48;                            // 0x48
    u32  field_0x4c;                            // 0x4c
    u32  field_0x50;                            // 0x50
    u32  field_0x54;                            // 0x54
    u32  field_0x58;                            // 0x58
    u8   field_0x5c;                            // 0x5c
    padding[3];                                 // 0x5d..0x5f
    u32  field_0x60;                            // 0x60
    padding[4];                                 // 0x64..0x67
    u32  field_0x68;                            // 0x68
    u32  field_0x6c;                            // 0x6c
    u32  field_0x70;                            // 0x70
    u32  field_0x74;                            // 0x74
    u32  field_0x78;                            // 0x78
    u32  field_0x7c;                            // 0x7c
    u32  field_0x80;                            // 0x80
    u32  field_0x84;                            // 0x84
    u8   field_0x88;                            // 0x88
} [[color("AAAAAA")]];

// =============================================================================
// TeleportComponent (0xAC / 172 bytes)
// =============================================================================
struct TeleportComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    u32  field_0x48;                            // 0x48
    u32  field_0x4c;                            // 0x4c
    u16  field_0x50;                            // 0x50
    padding[2];                                 // 0x52..0x53
    u8   field_0x54;                            // 0x54
    padding[15];                                // 0x55..0x63
    u32  field_0x64;                            // 0x64
    u32  field_0x68;                            // 0x68
    u8   field_0x6c;                            // 0x6c
    padding[15];                                // 0x6d..0x7b
    u32  field_0x7c;                            // 0x7c
    u32  field_0x80;                            // 0x80
    padding[12];                                // 0x84..0x8f
    u32  field_0x90;                            // 0x90
    u32  field_0x94;                            // 0x94
    u32  field_0x98;                            // 0x98
    u32  field_0x9c;                            // 0x9c
    u32  field_0xa0;                            // 0xa0
    u32  field_0xa4;                            // 0xa4
    u32  field_0xa8;                            // 0xa8
} [[color("AAAAAA")]];

// =============================================================================
// TextLogComponent (0x90 / 144 bytes)
// =============================================================================
struct TextLogComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    u8   field_0x48;                            // 0x48
    padding[15];                                // 0x49..0x57
    u32  field_0x58;                            // 0x58
    u32  field_0x5c;                            // 0x5c
    u8   field_0x60;                            // 0x60
    padding[15];                                // 0x61..0x6f
    u32  field_0x70;                            // 0x70
    u32  field_0x74;                            // 0x74
    u8   field_0x78;                            // 0x78
    padding[15];                                // 0x79..0x87
    u32  field_0x88;                            // 0x88
    u32  field_0x8c;                            // 0x8c
} [[color("AAAAAA")]];

// =============================================================================
// TorchComponent (0x68 / 104 bytes)
// =============================================================================
struct TorchComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    s32   probability_of_ignition_attempt;      // 0x48
    float suffocation_check_offset_y;           // 0x4c
    s32   frames_suffocated_to_extinguish;      // 0x50
    bool  extinguishable;                       // 0x54
    padding[3];                                 // 0x55..0x57
    float fire_audio_weight;                    // 0x58
    float mFlickerOffset;                       // 0x5c
    s32   mFramesSuffocated;                    // 0x60
    padding[1];                                 // 0x64
    bool  mFireIsBurningPrev;                   // 0x65
    padding[2];                                 // 0x66..0x67
} [[color("AAAAAA")]];

// =============================================================================
// UIIconComponent (0x93 / 147 bytes)
// =============================================================================
struct UIIconComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    MsvcString icon_sprite_file;                // 0x48 (std::string)
    MsvcString name;                            // 0x60 (std::string)
    MsvcString description;                      // 0x78 (std::string)
    bool  display_above_head;                   // 0x90
    bool  display_in_hud;                       // 0x91
    bool  is_perk;                              // 0x92
} [[color("AAAAAA")]];

// =============================================================================
// VariableStorageComponent (0x84 / 132 bytes)
// =============================================================================
struct VariableStorageComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    padding[24];                                // 0x48..0x5f
    u32   value_string [[comment("char* / std::string ptr")]]; // 0x60
    padding[20];                                // 0x64..0x77
    s32   value_int;                            // 0x78
    bool  value_bool;                           // 0x7c
    padding[3];                                 // 0x7d..0x7f
    float value_float;                          // 0x80
} [[color("AAAAAA")]];

// =============================================================================
// VegetationComponent (0xE4 / 228 bytes)
// NOTE: Ghidra layout is atypical - named fields begin at 0x04 (inside the
// 0x48 component header region). Per batch convention those sub-0x48 fields are
// absorbed into ComponentHeader; only fields fully at >= 0x48 are modeled.
// =============================================================================
struct VegetationComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    padding[16];                                // 0x48..0x57 (tail of tree_image_visual[])
    char  tree_material[24];                    // 0x58
    char  load_this_xml_instead[24];            // 0x70
    float visual_offset_x;                      // 0x88
    float visual_offset_y;                      // 0x8c
    char  visual_color[24];                     // 0x90
    bool  grass_requires_neighbors;             // 0xa8
    padding[3];                                 // 0xa9..0xab
    char  material_on_top_of[24];               // 0xac
    s32   height_check;                         // 0xc4
    s32   max_y;                                // 0xc8
    bool  random_flip_x_scale;                  // 0xcc
    padding[3];                                 // 0xcd..0xcf
    u32   mVisualColor;                         // 0xd0
    s32   tree_material_id;                     // 0xd4
    s32   material_id_on_top_of;                // 0xd8
    u32   mTreeData [[comment("TreeData*")]];   // 0xdc
    padding[4];                                 // 0xe0..0xe3
} [[color("AAAAAA")]];

// =============================================================================
// VerletPhysicsComponent (0x2B34 / 11060 bytes)
// Mostly an internal buffer; named fields modeled near the start, one big
// padding to reach the exact total size.
// =============================================================================
struct VerletPhysicsComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    s32   num_points;                           // 0x48
    s32   num_links;                            // 0x4c
    s32   width;                                // 0x50
    u32   field_0x54;                           // 0x54
    float resting_distance;                     // 0x58
    float mass_min;                             // 0x5c
    float mass_max;                             // 0x60
    float stiffness;                            // 0x64
    float velocity_dampening;                   // 0x68
    float liquid_damping;                       // 0x6c
    float gets_entity_velocity_coeff;           // 0x70
    bool  collide_with_cells;                   // 0x74
    bool  simulate_gravity;                     // 0x75
    bool  simulate_wind;                        // 0x76
    padding[1];                                 // 0x77
    float wind_change_speed;                    // 0x78
    bool  constrain_stretching;                 // 0x7c
    bool  pixelate_sprite_transforms;           // 0x7d
    bool  scale_sprite_x;                       // 0x7e
    bool  follow_entity_transform;              // 0x7f
    Vec2  animation_target_offset;              // 0x80
    float animation_amount;                     // 0x88
    float animation_speed;                      // 0x8c
    float animation_energy;                     // 0x90
    float cloth_sprite_z_index;                 // 0x94
    s32   stain_cells_probability;              // 0x98
    s32   cloth_color_edge;                     // 0x9c
    s32   cloth_color;                          // 0xa0
    Vec2  m_position_previous;                  // 0xa4
    bool  m_is_culled_previous;                 // 0xac
    padding[10883];                             // 0xad..0x2b2f (internal point/link buffer)
    u32   field_0x2b30;                         // 0x2b30
} [[color("AAAAAA")]];

// =============================================================================
// VerletWorldJointComponent (0x5C / 92 bytes)
// =============================================================================
struct VerletWorldJointComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    Vec2  world_position;                       // 0x48
    s32   verlet_point_index;                   // 0x50
    bool  mUpdated;                             // 0x54
    padding[3];                                 // 0x55..0x57
    u32   field_0x58;                           // 0x58
} [[color("AAAAAA")]];

// =============================================================================
// WormComponent (0xD4 / 212 bytes)
// =============================================================================
struct WormComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    float speed;                                // 0x48
    float acceleration;                         // 0x4c
    float ground_decceleration;                 // 0x50
    u32   field_0x54;                           // 0x54
    u32   field_0x58;                           // 0x58
    float gravity;                              // 0x5c
    float tail_gravity;                         // 0x60
    float part_distance;                        // 0x64
    s32   ground_check_offset;                  // 0x68
    float hitbox_radius;                        // 0x6c
    float bite_damage;                          // 0x70
    float target_kill_radius;                   // 0x74
    float target_kill_ragdoll_force;            // 0x78
    float jump_cam_shake;                       // 0x7c
    float jump_cam_shake_distance;              // 0x80
    float eat_anim_wait_mult;                   // 0x84
    MsvcString ragdoll_filename;                // 0x88 (std::string)
    bool  is_water_worm;                        // 0xa0
    padding[3];                                 // 0xa1..0xa3
    float max_speed;                            // 0xa4
    Vec2  mTargetVec;                           // 0xa8
    float mGravVelocity;                        // 0xb0
    float mSpeed;                               // 0xb4
    Vec2  mTargetPosition;                      // 0xb8
    float mTargetSpeed;                         // 0xc0
    bool  mOnGroundPrev;                        // 0xc4
    padding[3];                                 // 0xc5..0xc7
    s32   mMaterialIdPrev;                      // 0xc8
    s32   mFrameNextDamage;                     // 0xcc
    float mDirectionAdjustSpeed;                // 0xd0
} [[color("FF4444")]];

// =============================================================================
// WormPlayerComponent (0x58 / 88 bytes)
// =============================================================================
struct WormPlayerComponent {
    ComponentHeader header      [[inline]];     // 0x00..0x47
    // +0x48
    u32   field_0x48;                           // 0x48
    u32   field_0x4c;                           // 0x4c
    u32   field_0x50;                           // 0x50
    u32   field_0x54;                           // 0x54
} [[color("AAAAAA")]];

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
s32 g_newGamePlusCount       @ 0x01205024 [[color("AAFFAA"), comment("NG+ replay count, 0 on fresh run")]];
s32 g_deathCount             @ 0x01208AF8 [[color("AAFFAA"), comment("Night sky stars")]];
s32 g_numOrbsTotal           @ 0x01152544 [[color("AAFFAA"), comment("Total orb count (default 10)")]];
EntityManager *g_entityManager : u32 @ 0x01204B98 [[color("44FF44")]];
DeathMatchApp *g_deathMatchApp : u32 @ 0x01204BC0 [[color("44FF44")]];
GameGlobals *g_gameGlobals : u32     @ 0x0122374C [[color("AAAAAA")]];
WorldStateComponent *g_worldStateComponent : u32 @ 0x01205010 [[color("FF44FF")]];
u32 g_orbPersistence_ptr     @ 0x01207404 [[color("AAFFAA"), comment("-> OrbPersistenceData")]];

// --- Manager singletons & globals (names per ghidra labels) ---
// Pointers — hold a u32 that points at the heap object (deref once):
u32 g_entityTagManager        @ 0x01206FAC [[color("44FF44"), comment("ptr -> entity tag manager")]];
u32 g_componentTagManager     @ 0x01204B30 [[color("44FF44"), comment("ptr -> component tag manager")]];
u32 g_eventManager            @ 0x01204B34 [[color("44FF44"), comment("ptr -> event manager")]];
u32 g_worldStateEntity        @ 0x01204BD0 [[color("FF44FF"), comment("ptr -> world-state Entity")]];
u32 g_componentSystemManager  @ 0x012236E8 [[color("44FF44"), comment("ptr -> component system manager")]];
// Inline objects — struct data begins at this address (layout not modeled here):
u32 g_poroPlatformInstance    @ 0x01221BC0 [[color("AAAAAA"), comment("PoroPlatform, 176B inline")]];
u32 g_statsStruct             @ 0x01208940 [[color("AAFFAA"), comment("GlobalStats, 832B inline")]];
u32 g_gameMode                @ 0x0120761C [[color("AAFFAA"), comment("game mode (inline)")]];
u32 g_componentTypeManager    @ 0x01223C88 [[color("AAAAAA"), comment("component type manager (inline)")]];
u32 g_componentIdMap          @ 0x01208018 [[color("AAAAAA"), comment("component id map (inline)")]];
u32 g_translationManager      @ 0x01207C28 [[color("AAAAAA"), comment("translation manager (inline)")]];
u32 g_fileNames               @ 0x01207BD4 [[color("AAAAAA"), comment("file-names vector (inline)")]];
u32 g_inventory               @ 0x012224F0 [[color("AAAAAA"), comment("inventory (inline)")]];
u32 g_mods                    @ 0x01207E90 [[color("AAAAAA"), comment("mods (inline)")]];
// Scalars (typed by ghidra):
u8   g_logFlush               @ 0x01204FE0 [[color("AAAAAA"), comment("log flush flag (bool; u8 to avoid hexpatgen bool-global codegen bug)")]];
s32  g_logLevel               @ 0x01152518 [[color("AAAAAA"), comment("log level")]];
s32  g_maxComponent           @ 0x01152FF0 [[color("AAAAAA"), comment("max component count")]];
u8   g_debugSettings          @ 0x01207E20 [[color("AAAAAA"), comment("debug settings (struct base, first byte)")]];
