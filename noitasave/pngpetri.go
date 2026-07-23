package noitasave

// .png_petri — a 512x512 world chunk, saved as save00/world/world_<x>_<y>.png_petri.
//
// Layout of the decompressed payload (all BIG-endian). Reversed from
// PngPetri_Serialize (noita.exe @ 0x0073e590) and its helpers; the save side is
// Grid_IMPL_SaveWorldChunk @ 0x00740e70, the restore side is
// PngPetri_RestorePhysicsObjectsAndJoints @ 0x0073eed0.
//
//	u32 version                  // 24 in current builds
//	u32 width, height            // always 512 x 512
//	u8  cells[width*height]      // bit 7 = has custom color, bits 0-6 = material index
//	vector<string> materialNames // cell material index -> name
//	vector<u32>    customColors  // consumed in row-major order by cells with bit 7
//	vector<PhysicsObject>
//	vector<Joint>                // version >= 18 only
//
// Version gates, from the serializer's branches:
//
//	>= 18  joints section exists; PhysicsObject gains id + damping/shape doubles
//	>= 19  Joint.ID
//	>= 20  Joint motor fields
//	>= 21  PhysicsObject.AutoClean and .Z
//
// Only version 24 has been exercised against real saves; the older branches are
// implemented from the decompiled serializer but are untested.

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
)

// CurrentVersion is the format version current Noita builds write.
const CurrentVersion = 24

// ChunkSize is the fixed pixel width and height of a world chunk.
const ChunkSize = 512

// PhysicsObject is one rigid body stored in a chunk: a Box2D body plus the
// pixel image its collision shape and appearance are rebuilt from.
type PhysicsObject struct {
	// ID is the body's unique id (b2Body::m_userData). Joints reference
	// bodies by this. Absent before version 18.
	ID uint64

	// MaterialIndex indexes Chunk.MaterialNames — the material every pixel of
	// this body is made of (wood, glass, meat, ...).
	MaterialIndex uint32

	// X, Y and Rot are the body pose in world pixel coordinates / radians.
	// Velocities are deliberately not stored: bodies restore at rest.
	X, Y float32
	Rot  float32

	// CircleRadius > 0 marks a body whose collider is a single b2CircleShape,
	// with CircleCX/CY as its body-local center. Zero for ordinary pixel
	// bodies, whose fixtures are re-derived from Colors on load.
	CircleRadius       float64
	CircleCX, CircleCY float64

	LinearDamping  float64
	AngularDamping float64

	AllowSleep    bool // Box2D e_autoSleepFlag
	FixedRotation bool
	IsStatic      bool // b2_staticBody
	IsBullet      bool // continuous collision detection

	// AutoClean lets the simulation delete this body when it is buried in
	// sand. Z is the render depth. Both version >= 21.
	AutoClean bool
	Z         float32

	// Width, Height and Colors are the body's RGBA pixel image. Alpha 0 means
	// empty, so damage taken before the save persists as transparent pixels.
	Width, Height uint32
	Colors        []uint32
}

// Joint kinds, as stored in Joint.Kind (the tag Noita keeps at b2Joint+0xa).
const (
	JointRevolute                  = 1
	JointWeld                      = 8
	JointRagdollA                  = 0x2b67
	JointRagdollB                  = 0x2b68
	JointPlainWeld                 = 0x2b69
	JointPlainRevolute             = 0x2b6a
	JointRevoluteAttachToNearby    = 0x2b6b
	JointWeldAttachToNearbySurface = 0x2b6c
)

// Joint connects two PhysicsObjects by their IDs.
//
// On load, a joint whose bodies are not both present causes the game to stop
// restoring joints for the whole chunk — every later joint is silently
// dropped. Keep BodyAID/BodyBID pointing at live bodies when editing.
type Joint struct {
	ID   uint64 // version >= 19
	Kind uint32 // widened from a u16 in memory; see the Joint* constants

	BodyAID, BodyBID uint64

	CollideConnected bool

	// BreakForce is a multiplier, not an absolute: the game computes
	// (massA+massB) * PHYSICS_JOINT_MAX_FORCE_MULTIPLIER * BreakForce.
	BreakForce           float32
	BreakDistance        float32
	BreakOnShearAngleDeg float32
	BreakOnBodyModified  bool

	// Box2D local anchors on each body.
	AnchorAX, AnchorAY float64
	AnchorBX, AnchorBY float64

	// Pixel coordinates of the attachment cells within each body, used to
	// re-find the right CSolidCell on load.
	CellAPX, CellAPY int32
	CellBPX, CellBPY int32

	// Raycast hit point for the ATTACH_TO_NEARBY_SURFACE kinds. Stored through
	// the u32 serializer but semantically float.
	SurfaceAttachX, SurfaceAttachY float32

	// Motor fields, version >= 20.
	EnableMotor    bool
	MotorSpeed     float64
	MaxMotorTorque float64
}

// Chunk is a decoded .png_petri world chunk.
type Chunk struct {
	Version       uint32
	Width, Height uint32

	// Cells is Width*Height material bytes in row-major order. Use CellMaterial
	// to decode one.
	Cells []byte

	MaterialNames []string
	CustomColors  []uint32

	PhysicsObjects []PhysicsObject
	Joints         []Joint
}

// CellMaterial decodes a cell byte into its material index and whether the
// cell carries a custom color. A zero byte means empty (no material).
func CellMaterial(cell byte) (materialIndex uint8, hasCustomColor bool) {
	return cell & 0x7f, cell&0x80 != 0
}

// CellAt returns the raw cell byte at (x, y).
func (c *Chunk) CellAt(x, y int) (byte, error) {
	if x < 0 || y < 0 || x >= int(c.Width) || y >= int(c.Height) {
		return 0, fmt.Errorf("noitasave: cell (%d,%d) out of bounds %dx%d", x, y, c.Width, c.Height)
	}
	return c.Cells[y*int(c.Width)+x], nil
}

// MaterialNameAt returns the material name of the cell at (x, y), or "" for an
// empty cell.
func (c *Chunk) MaterialNameAt(x, y int) (string, error) {
	cell, err := c.CellAt(x, y)
	if err != nil {
		return "", err
	}
	if cell == 0 {
		return "", nil
	}
	idx, _ := CellMaterial(cell)
	if int(idx) >= len(c.MaterialNames) {
		return "", fmt.Errorf("noitasave: material index %d exceeds %d names", idx, len(c.MaterialNames))
	}
	return c.MaterialNames[idx], nil
}

// CustomColorIndex maps each cell to its index into CustomColors, or -1 when
// the cell has no custom color. Colors are consumed in row-major order, so
// resolving one cell means counting every flagged cell before it; this walks
// the grid once and hands back the whole mapping.
func (c *Chunk) CustomColorIndex() []int32 {
	idx := make([]int32, len(c.Cells))
	next := int32(0)
	for i, cell := range c.Cells {
		if cell != 0 && cell&0x80 != 0 {
			idx[i] = next
			next++
		} else {
			idx[i] = -1
		}
	}
	return idx
}

// --- decoding -------------------------------------------------------------

type reader struct {
	b   []byte
	pos int
	err error
}

func (r *reader) fail(format string, args ...any) {
	if r.err == nil {
		r.err = fmt.Errorf("noitasave: "+format, args...)
	}
}

func (r *reader) take(n int) []byte {
	if r.err != nil {
		return nil
	}
	if n < 0 || r.pos+n > len(r.b) {
		r.fail("payload truncated: need %d bytes at offset %d, have %d", n, r.pos, len(r.b)-r.pos)
		return nil
	}
	b := r.b[r.pos : r.pos+n]
	r.pos += n
	return b
}

func (r *reader) u8() uint8 {
	b := r.take(1)
	if b == nil {
		return 0
	}
	return b[0]
}

func (r *reader) bool() bool { return r.u8() != 0 }

func (r *reader) u32() uint32 {
	b := r.take(4)
	if b == nil {
		return 0
	}
	return binary.BigEndian.Uint32(b)
}

func (r *reader) i32() int32 { return int32(r.u32()) }

func (r *reader) u64() uint64 {
	b := r.take(8)
	if b == nil {
		return 0
	}
	return binary.BigEndian.Uint64(b)
}

func (r *reader) f32() float32 { return math.Float32frombits(r.u32()) }

func (r *reader) f64() float64 { return math.Float64frombits(r.u64()) }

func (r *reader) str() string {
	n := r.u32()
	if r.err != nil {
		return ""
	}
	b := r.take(int(n))
	if b == nil {
		return ""
	}
	return string(b)
}

// count reads a vector length and sanity-checks it against the bytes left, so
// a corrupt file fails fast instead of trying to allocate gigabytes.
func (r *reader) count(minElemSize int) int {
	n := r.u32()
	if r.err != nil {
		return 0
	}
	if minElemSize > 0 && int64(n)*int64(minElemSize) > int64(len(r.b)-r.pos) {
		r.fail("vector of %d elements exceeds remaining %d bytes", n, len(r.b)-r.pos)
		return 0
	}
	return int(n)
}

// DecodeChunk parses a decompressed .png_petri payload.
func DecodeChunk(payload []byte) (*Chunk, error) {
	r := &reader{b: payload}
	c := &Chunk{}

	c.Version = r.u32()
	c.Width = r.u32()
	c.Height = r.u32()
	if r.err != nil {
		return nil, r.err
	}
	if c.Width == 0 || c.Height == 0 || int64(c.Width)*int64(c.Height) > int64(len(payload)) {
		return nil, fmt.Errorf("noitasave: implausible chunk dimensions %dx%d", c.Width, c.Height)
	}

	cells := r.take(int(c.Width) * int(c.Height))
	if r.err != nil {
		return nil, r.err
	}
	c.Cells = make([]byte, len(cells))
	copy(c.Cells, cells)

	n := r.count(4) // each string is at least a 4-byte length
	c.MaterialNames = make([]string, 0, max(n, 0))
	for i := 0; i < n && r.err == nil; i++ {
		c.MaterialNames = append(c.MaterialNames, r.str())
	}

	n = r.count(4)
	c.CustomColors = make([]uint32, 0, max(n, 0))
	for i := 0; i < n && r.err == nil; i++ {
		c.CustomColors = append(c.CustomColors, r.u32())
	}

	n = r.count(physicsObjectMinSize(c.Version))
	c.PhysicsObjects = make([]PhysicsObject, 0, max(n, 0))
	for i := 0; i < n && r.err == nil; i++ {
		c.PhysicsObjects = append(c.PhysicsObjects, r.physicsObject(c.Version))
	}

	if c.Version >= 18 {
		n = r.count(jointSize(c.Version))
		c.Joints = make([]Joint, 0, max(n, 0))
		for i := 0; i < n && r.err == nil; i++ {
			c.Joints = append(c.Joints, r.joint(c.Version))
		}
	}

	if r.err != nil {
		return nil, r.err
	}
	if r.pos != len(payload) {
		return nil, fmt.Errorf("noitasave: %d trailing bytes after chunk (parsed %d of %d)",
			len(payload)-r.pos, r.pos, len(payload))
	}
	return c, nil
}

// physicsObjectMinSize is the fixed-field size of a serialized PhysicsObject,
// excluding its pixel image.
func physicsObjectMinSize(version uint32) int {
	switch {
	case version >= 21:
		return 81
	case version >= 18:
		return 76
	default:
		return 25
	}
}

// jointSize is the exact serialized size of a Joint: 115 bytes at version 24.
func jointSize(version uint32) int {
	// kind(4) + bodies(16) + collide(1) + break floats(12) + break bool(1)
	// + anchors(32) + cell coords(16) + surface xy(8)
	n := 90
	if version >= 19 {
		n += 8 // ID
	}
	if version >= 20 {
		n += 17 // motor: bool + 2 doubles
	}
	return n
}

func (r *reader) physicsObject(version uint32) PhysicsObject {
	var o PhysicsObject
	if version >= 18 {
		o.ID = r.u64()
	}
	o.MaterialIndex = r.u32()
	o.X = r.f32()
	o.Y = r.f32()
	o.Rot = r.f32()
	if version >= 18 {
		o.CircleRadius = r.f64()
		o.CircleCX = r.f64()
		o.CircleCY = r.f64()
		o.LinearDamping = r.f64()
		o.AngularDamping = r.f64()
		o.AllowSleep = r.bool()
		o.FixedRotation = r.bool()
	}
	o.IsStatic = r.bool()
	if version >= 18 {
		o.IsBullet = r.bool()
	}
	if version >= 21 {
		o.AutoClean = r.bool()
		o.Z = r.f32()
	}

	o.Width = r.u32()
	o.Height = r.u32()
	if r.err != nil {
		return o
	}
	npx := int64(o.Width) * int64(o.Height)
	if npx*4 > int64(len(r.b)-r.pos) {
		r.fail("physics object image %dx%d exceeds remaining %d bytes", o.Width, o.Height, len(r.b)-r.pos)
		return o
	}
	o.Colors = make([]uint32, npx)
	for i := range o.Colors {
		o.Colors[i] = r.u32()
	}
	return o
}

func (r *reader) joint(version uint32) Joint {
	var j Joint
	if version >= 19 {
		j.ID = r.u64()
	}
	j.Kind = r.u32()
	j.BodyAID = r.u64()
	j.BodyBID = r.u64()
	j.CollideConnected = r.bool()
	j.BreakForce = r.f32()
	j.BreakDistance = r.f32()
	j.BreakOnShearAngleDeg = r.f32()
	j.BreakOnBodyModified = r.bool()
	j.AnchorAX = r.f64()
	j.AnchorAY = r.f64()
	j.AnchorBX = r.f64()
	j.AnchorBY = r.f64()
	j.CellAPX = r.i32()
	j.CellAPY = r.i32()
	j.CellBPX = r.i32()
	j.CellBPY = r.i32()
	j.SurfaceAttachX = r.f32()
	j.SurfaceAttachY = r.f32()
	if version >= 20 {
		j.EnableMotor = r.bool()
		j.MotorSpeed = r.f64()
		j.MaxMotorTorque = r.f64()
	}
	return j
}

// --- encoding -------------------------------------------------------------

type writer struct{ b []byte }

func (w *writer) u8(v uint8) { w.b = append(w.b, v) }
func (w *writer) bool(v bool) {
	if v {
		w.u8(1)
	} else {
		w.u8(0)
	}
}
func (w *writer) u32(v uint32)  { w.b = binary.BigEndian.AppendUint32(w.b, v) }
func (w *writer) i32(v int32)   { w.u32(uint32(v)) }
func (w *writer) u64(v uint64)  { w.b = binary.BigEndian.AppendUint64(w.b, v) }
func (w *writer) f32(v float32) { w.u32(math.Float32bits(v)) }
func (w *writer) f64(v float64) { w.u64(math.Float64bits(v)) }

func (w *writer) str(s string) {
	w.u32(uint32(len(s)))
	w.b = append(w.b, s...)
}

// Encode serializes the chunk back to an uncompressed payload.
func (c *Chunk) Encode() ([]byte, error) {
	if int64(c.Width)*int64(c.Height) != int64(len(c.Cells)) {
		return nil, fmt.Errorf("noitasave: %d cells does not match %dx%d", len(c.Cells), c.Width, c.Height)
	}
	for i, o := range c.PhysicsObjects {
		if int64(o.Width)*int64(o.Height) != int64(len(o.Colors)) {
			return nil, fmt.Errorf("noitasave: physics object %d has %d colors for %dx%d",
				i, len(o.Colors), o.Width, o.Height)
		}
	}
	if c.Version < 18 && len(c.Joints) > 0 {
		return nil, fmt.Errorf("noitasave: version %d cannot store joints", c.Version)
	}

	w := &writer{b: make([]byte, 0, 12+len(c.Cells)+4096)}
	w.u32(c.Version)
	w.u32(c.Width)
	w.u32(c.Height)
	w.b = append(w.b, c.Cells...)

	w.u32(uint32(len(c.MaterialNames)))
	for _, s := range c.MaterialNames {
		w.str(s)
	}

	w.u32(uint32(len(c.CustomColors)))
	for _, v := range c.CustomColors {
		w.u32(v)
	}

	w.u32(uint32(len(c.PhysicsObjects)))
	for i := range c.PhysicsObjects {
		w.physicsObject(&c.PhysicsObjects[i], c.Version)
	}

	if c.Version >= 18 {
		w.u32(uint32(len(c.Joints)))
		for i := range c.Joints {
			w.joint(&c.Joints[i], c.Version)
		}
	}
	return w.b, nil
}

func (w *writer) physicsObject(o *PhysicsObject, version uint32) {
	if version >= 18 {
		w.u64(o.ID)
	}
	w.u32(o.MaterialIndex)
	w.f32(o.X)
	w.f32(o.Y)
	w.f32(o.Rot)
	if version >= 18 {
		w.f64(o.CircleRadius)
		w.f64(o.CircleCX)
		w.f64(o.CircleCY)
		w.f64(o.LinearDamping)
		w.f64(o.AngularDamping)
		w.bool(o.AllowSleep)
		w.bool(o.FixedRotation)
	}
	w.bool(o.IsStatic)
	if version >= 18 {
		w.bool(o.IsBullet)
	}
	if version >= 21 {
		w.bool(o.AutoClean)
		w.f32(o.Z)
	}
	w.u32(o.Width)
	w.u32(o.Height)
	for _, px := range o.Colors {
		w.u32(px)
	}
}

func (w *writer) joint(j *Joint, version uint32) {
	if version >= 19 {
		w.u64(j.ID)
	}
	w.u32(j.Kind)
	w.u64(j.BodyAID)
	w.u64(j.BodyBID)
	w.bool(j.CollideConnected)
	w.f32(j.BreakForce)
	w.f32(j.BreakDistance)
	w.f32(j.BreakOnShearAngleDeg)
	w.bool(j.BreakOnBodyModified)
	w.f64(j.AnchorAX)
	w.f64(j.AnchorAY)
	w.f64(j.AnchorBX)
	w.f64(j.AnchorBY)
	w.i32(j.CellAPX)
	w.i32(j.CellAPY)
	w.i32(j.CellBPX)
	w.i32(j.CellBPY)
	w.f32(j.SurfaceAttachX)
	w.f32(j.SurfaceAttachY)
	if version >= 20 {
		w.bool(j.EnableMotor)
		w.f64(j.MotorSpeed)
		w.f64(j.MaxMotorTorque)
	}
}

// --- files ----------------------------------------------------------------

// LoadChunk reads and decodes a .png_petri file.
func LoadChunk(path string) (*Chunk, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	payload, err := DecodeContainer(raw)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	c, err := DecodeChunk(payload)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return c, nil
}

// SaveChunk encodes and writes a .png_petri file.
//
// It writes to a temporary file in the same directory and renames it into
// place, so an interrupted write cannot leave a half-written chunk that the
// game would fail to load.
func SaveChunk(path string, c *Chunk) error {
	payload, err := c.Encode()
	if err != nil {
		return err
	}
	blob, err := EncodeContainer(payload)
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dirOf(path), ".png_petri-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := tmp.Write(blob); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

func dirOf(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[:i]
		}
	}
	return "."
}
