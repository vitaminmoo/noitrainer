#pragma description "Noita world chunk (save00/world/world_<x>_<y>.png_petri) -- inline fastlz decompress"
#pragma endian big

// =============================================================================
// Apply DIRECTLY to a raw .png_petri -- no pre-decompression needed.
//
// Requires the FastLZ ImHex plugin (fastlz::*), same as world_tree_compressed.hexpat:
//     ~/opt/imhex/lib64/imhex/plugins/fastlz_plugin.hexplug
// plus the pattern-language standard library (std.mem). If the plugin is
// missing, run the payload through wt_decompress and delete the two
// decompression lines at the bottom, changing `in payload` to a plain @ 0x00.
//
// Container on disk (LITTLE-endian, consumed by the plugin):
//     le u32 compressedSize (== filesize-8); le u32 uncompressedSize; u8 data[]
// Payload (BIG-endian) parsed below.
//
// Layout reversed from PngPetri_Serialize (noita.exe @ 0x0073e590); the save
// side is Grid_IMPL_SaveWorldChunk @ 0x00740e70 and the restore side is
// PngPetri_RestorePhysicsObjectsAndJoints @ 0x0073eed0. Cross-checked by
// round-tripping every chunk of a real save byte-for-byte (noitasave package).
//
// Serialization is uniform: vector<T> = be u32 count + elements;
// std::string = be u32 length + bytes (no NUL). This is the ISerializer wire
// form, NOT the in-memory MSVC layout in noita.hexpat.
// =============================================================================

import std.mem;

// ---- container header (LITTLE-endian; overrides the big-endian pragma) ------
le u32 container_compressed_size   @ 0x00 [[color("888888"), comment("== filesize - 8")]];
le u32 container_uncompressed_size @ 0x04 [[color("888888"), comment("decompressed payload size")]];

std::mem::Section payload = std::mem::create_section("png_petri payload (decompressed)");
fastlz::decompress_to_section(fastlz::main_section(), 0x00, payload);

// Version gates the record layouts, so read it up front.
//   >= 18  joints section; PhysicsObject gains id + damping/shape doubles
//   >= 19  Joint.id
//   >= 20  Joint motor fields
//   >= 21  PhysicsObject.auto_clean and .z
be u32 gVersion @ 0x00 in payload;

// -----------------------------------------------------------------------------

struct NoitaString {
    be u32 length [[color("888888")]];
    char value[length];
} [[sealed, format("fmt_noita_string")]];

fn fmt_noita_string(NoitaString s) {
    return s.value;
};

// One rigid body: a Box2D body plus the pixel image its collision shape and
// appearance are rebuilt from. Velocities are deliberately NOT stored, so
// bodies always restore at rest.
struct PhysicsObject {
    if (gVersion >= 18) {
        be u64 id [[color("FFD23C"), comment("b2Body::m_userData; joints reference bodies by this")]];
    }
    be u32 material_index [[color("3CA0FF"), comment("index into material_names")]];
    be float x [[color("8FE36B")]];
    be float y [[color("8FE36B")]];
    be float rot_radians [[color("8FE36B")]];

    if (gVersion >= 18) {
        // Non-zero radius marks a single-b2CircleShape body; zero means the
        // fixtures are re-derived from the pixel image below.
        be double circle_radius [[color("FF9F3C"), comment("0 for ordinary pixel bodies")]];
        be double circle_center_x [[color("FF9F3C")]];
        be double circle_center_y [[color("FF9F3C")]];
        be double linear_damping;
        be double angular_damping;
        bool allow_sleep    [[comment("Box2D e_autoSleepFlag")]];
        bool fixed_rotation;
    }
    bool is_static [[comment("b2_staticBody")]];
    if (gVersion >= 18) {
        bool is_bullet [[comment("continuous collision detection")]];
    }
    if (gVersion >= 21) {
        bool auto_clean [[comment("simulation may delete this body when buried in sand")]];
        be float z      [[comment("render depth; no physics effect")]];
    }

    be u32 width;
    be u32 height;
    // RGBA, one word per pixel. Alpha 0 = empty, so damage taken before the
    // save persists as transparent pixels.
    be u32 colors[width * height] [[color("C86BE3")]];
};

// Joint kinds (the tag Noita keeps at b2Joint+0xa, widened to u32 on disk).
enum JointKind : u32 {
    REVOLUTE                          = 1,
    WELD                              = 8,
    RAGDOLL_A                         = 0x2b67,
    RAGDOLL_B                         = 0x2b68,
    PLAIN_WELD                        = 0x2b69,
    PLAIN_REVOLUTE                    = 0x2b6a,
    REVOLUTE_ATTACH_TO_NEARBY_SURFACE = 0x2b6b,
    WELD_ATTACH_TO_NEARBY_SURFACE     = 0x2b6c,
};

// Connects two PhysicsObjects by id. On load, a joint whose bodies are not both
// resolvable makes the game stop restoring joints for the WHOLE chunk -- every
// later joint is silently dropped.
struct Joint {
    if (gVersion >= 19) {
        be u64 id [[color("FFD23C")]];
    }
    JointKind kind [[color("FF5555")]];
    be u64 body_a_id [[color("FFD23C"), comment("matches PhysicsObject.id")]];
    be u64 body_b_id [[color("FFD23C"), comment("matches PhysicsObject.id")]];

    bool collide_connected;
    // Multiplier, not an absolute: the game computes
    // (massA+massB) * PHYSICS_JOINT_MAX_FORCE_MULTIPLIER * break_force.
    be float break_force;
    be float break_distance;
    be float break_on_shear_angle_deg;
    bool break_on_body_modified;

    be double anchor_a_x [[color("8FE36B")]];
    be double anchor_a_y [[color("8FE36B")]];
    be double anchor_b_x [[color("8FE36B")]];
    be double anchor_b_y [[color("8FE36B")]];

    // Pixel coords of the attachment cells within each body, used to re-find
    // the right CSolidCell on load.
    be s32 cell_a_px;
    be s32 cell_a_py;
    be s32 cell_b_px;
    be s32 cell_b_py;

    // Raycast hit for the ATTACH_TO_NEARBY_SURFACE kinds. Written through the
    // u32 serializer but semantically float.
    be float surface_attach_x;
    be float surface_attach_y;

    if (gVersion >= 20) {
        bool enable_motor;
        be double motor_speed;
        be double max_motor_torque;
    }
};

struct PngPetri {
    be u32 version [[color("FF5555"), comment("24 in current builds")]];
    be u32 width   [[color("FFD23C")]];
    be u32 height  [[color("FFD23C")]];

    // Each byte: bit 7 = has custom color, bits 0-6 = index into
    // material_names. 0 = empty cell. Cells with bit 7 set consume
    // custom_colors entries in row-major order.
    u8 cells[width * height] [[color("3CA0FF")]];

    be u32 material_name_count;
    NoitaString material_names[material_name_count];

    be u32 custom_color_count;
    be u32 custom_colors[custom_color_count] [[color("C86BE3")]];

    be u32 physics_object_count [[color("FFD23C")]];
    PhysicsObject physics_objects[physics_object_count];

    if (version >= 18) {
        be u32 joint_count [[color("FFD23C")]];
        Joint joints[joint_count];
    }

    // Sequential member: zero-length when the file ends exactly where the
    // known structure does. Any MAGENTA bytes mean an undocumented trailing
    // section -- reverse-engineer them.
    u8 unexplained_trailing[while($ < container_uncompressed_size)]
        [[color("FF00FF"),
          comment("UNEXPECTED: bytes past the known structure")]];
};

PngPetri chunk @ 0x00 in payload;
