package noitasave

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// realChunks returns every .png_petri under $NOITA_SAVE_WORLD, or skips.
// Point it at e.g. .../Nolla_Games_Noita/save00/world.
func realChunks(t *testing.T) []string {
	t.Helper()
	dir := os.Getenv("NOITA_SAVE_WORLD")
	if dir == "" {
		t.Skip("set NOITA_SAVE_WORLD to a save00/world directory to run this test")
	}
	paths, err := filepath.Glob(filepath.Join(dir, "*.png_petri"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Skipf("no .png_petri files in %s", dir)
	}
	return paths
}

// TestRealChunksRoundTrip is the load-bearing test: for every chunk the game
// wrote, decoding and re-encoding must reproduce the payload byte for byte.
// Any field we mis-sized or mis-ordered shows up as a length or content
// mismatch, and any field we failed to consume shows up as trailing bytes.
func TestRealChunksRoundTrip(t *testing.T) {
	paths := realChunks(t)

	var objects, joints, withColors int
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		payload, err := DecodeContainer(raw)
		if err != nil {
			t.Errorf("%s: container: %v", filepath.Base(path), err)
			continue
		}

		c, err := DecodeChunk(payload)
		if err != nil {
			t.Errorf("%s: decode: %v", filepath.Base(path), err)
			continue
		}
		if c.Version != CurrentVersion {
			t.Logf("%s: version %d (expected %d)", filepath.Base(path), c.Version, CurrentVersion)
		}
		if c.Width != ChunkSize || c.Height != ChunkSize {
			t.Errorf("%s: got %dx%d, want %dx%d", filepath.Base(path), c.Width, c.Height, ChunkSize, ChunkSize)
		}

		got, err := c.Encode()
		if err != nil {
			t.Errorf("%s: encode: %v", filepath.Base(path), err)
			continue
		}
		if !bytes.Equal(got, payload) {
			t.Errorf("%s: re-encoded payload differs (got %d bytes, want %d); first diff at %d",
				filepath.Base(path), len(got), len(payload), firstDiff(got, payload))
			continue
		}

		objects += len(c.PhysicsObjects)
		joints += len(c.Joints)
		if len(c.CustomColors) > 0 {
			withColors++
		}
	}
	t.Logf("round-tripped %d chunks: %d physics objects, %d joints, %d chunks with custom colors",
		len(paths), objects, joints, withColors)

	if objects == 0 {
		t.Error("no physics objects across any chunk — the parse is probably wrong")
	}
}

// TestRealChunksReproduceFileBytes is the end-to-end check: decoding a chunk
// and writing it straight back out must reproduce the original file byte for
// byte, container and compression included. That makes a re-written save
// indistinguishable from one the game wrote.
func TestRealChunksReproduceFileBytes(t *testing.T) {
	for _, path := range realChunks(t) {
		want, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		c, err := LoadChunk(path)
		if err != nil {
			t.Errorf("%s: %v", filepath.Base(path), err)
			continue
		}
		payload, err := c.Encode()
		if err != nil {
			t.Errorf("%s: encode: %v", filepath.Base(path), err)
			continue
		}
		got, err := EncodeContainer(payload)
		if err != nil {
			t.Errorf("%s: container: %v", filepath.Base(path), err)
			continue
		}
		if !bytes.Equal(got, want) {
			t.Errorf("%s: rewritten file differs (got %d bytes, want %d, first diff at %d)",
				filepath.Base(path), len(got), len(want), firstDiff(got, want))
		}
	}
}

// TestRealChunksConsistency sanity-checks decoded values against invariants the
// format implies, so a parse that happens to round-trip but is misaligned still
// gets caught.
func TestRealChunksConsistency(t *testing.T) {
	for _, path := range realChunks(t) {
		c, err := LoadChunk(path)
		if err != nil {
			t.Errorf("%s: %v", filepath.Base(path), err)
			continue
		}
		name := filepath.Base(path)

		// Every non-empty cell must name a real material.
		colorsNeeded := 0
		for i, cell := range c.Cells {
			if cell == 0 {
				continue
			}
			idx, custom := CellMaterial(cell)
			if int(idx) >= len(c.MaterialNames) {
				t.Fatalf("%s: cell %d material index %d >= %d names", name, i, idx, len(c.MaterialNames))
			}
			if custom {
				colorsNeeded++
			}
		}
		// Cells flagged with a custom color consume exactly the color vector.
		if colorsNeeded != len(c.CustomColors) {
			t.Errorf("%s: %d cells flagged custom-color but %d colors stored",
				name, colorsNeeded, len(c.CustomColors))
		}

		for i, o := range c.PhysicsObjects {
			if int(o.MaterialIndex) >= len(c.MaterialNames) {
				t.Errorf("%s: object %d material index %d >= %d names",
					name, i, o.MaterialIndex, len(c.MaterialNames))
			}
			// A body is either a circle or a pixel image, never neither.
			if o.CircleRadius <= 0 && len(o.Colors) == 0 {
				t.Errorf("%s: object %d has neither circle radius nor pixels", name, i)
			}
			// Damping values are small non-negative reals in practice; a
			// misaligned parse tends to produce absurd magnitudes.
			if o.LinearDamping < 0 || o.LinearDamping > 1e6 {
				t.Errorf("%s: object %d implausible linear damping %g", name, i, o.LinearDamping)
			}
		}

		bodies := make(map[uint64]bool, len(c.PhysicsObjects))
		for _, o := range c.PhysicsObjects {
			bodies[o.ID] = true
		}
		for i, j := range c.Joints {
			switch j.Kind {
			case JointRevolute, JointWeld, JointRagdollA, JointRagdollB,
				JointPlainWeld, JointPlainRevolute,
				JointRevoluteAttachToNearby, JointWeldAttachToNearbySurface:
			default:
				t.Errorf("%s: joint %d unknown kind 0x%x", name, i, j.Kind)
			}
			// Not an error: the partner body legitimately lives in a
			// neighbouring chunk. Worth surfacing because the game drops every
			// later joint in the chunk when it cannot resolve one.
			if !bodies[j.BodyAID] || !bodies[j.BodyBID] {
				t.Logf("%s: joint %d references a body outside this chunk", name, i)
			}
		}
	}
}

// TestCustomColorIndex checks the row-major colour cursor.
func TestCustomColorIndex(t *testing.T) {
	c := &Chunk{
		Width: 2, Height: 2,
		Cells:         []byte{0x81, 0x01, 0x00, 0x82},
		MaterialNames: []string{"air", "wood", "meat"},
		CustomColors:  []uint32{0xAABBCCDD, 0x11223344},
	}
	idx := c.CustomColorIndex()
	want := []int32{0, -1, -1, 1}
	for i := range want {
		if idx[i] != want[i] {
			t.Errorf("cell %d: got color index %d, want %d", i, idx[i], want[i])
		}
	}
}

// TestSyntheticRoundTrip exercises encode/decode without needing a real save,
// including the container and both physics record types.
func TestSyntheticRoundTrip(t *testing.T) {
	in := &Chunk{
		Version: CurrentVersion,
		Width:   2, Height: 2,
		Cells:         []byte{0x00, 0x01, 0x82, 0x03},
		MaterialNames: []string{"air", "wood", "meat", "glass"},
		CustomColors:  []uint32{0xDEADBEEF},
		PhysicsObjects: []PhysicsObject{{
			ID: 0x1122334455667788, MaterialIndex: 1,
			X: 12.5, Y: -3.25, Rot: 1.5,
			CircleRadius: 0, CircleCX: 0, CircleCY: 0,
			LinearDamping: 0.5, AngularDamping: 0.25,
			AllowSleep: true, IsStatic: false, IsBullet: true,
			AutoClean: true, Z: -0.5,
			Width: 2, Height: 1,
			Colors: []uint32{0xFF0000FF, 0x00FF00FF},
		}},
		Joints: []Joint{{
			ID: 42, Kind: JointRevolute,
			BodyAID: 0x1122334455667788, BodyBID: 9,
			CollideConnected: true, BreakForce: 2.5,
			AnchorAX: 1.5, AnchorBY: -2.5,
			CellAPX: 3, CellBPY: -4,
			SurfaceAttachX: 7.5,
			EnableMotor:    true, MotorSpeed: 3.5, MaxMotorTorque: 100,
		}},
	}

	payload, err := in.Encode()
	if err != nil {
		t.Fatal(err)
	}
	blob, err := EncodeContainer(payload)
	if err != nil {
		t.Fatal(err)
	}
	back, err := DecodeContainer(blob)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(back, payload) {
		t.Fatal("container round-trip corrupted the payload")
	}

	out, err := DecodeChunk(back)
	if err != nil {
		t.Fatal(err)
	}
	got, err := out.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("chunk round-trip differs at byte %d", firstDiff(got, payload))
	}
	if out.PhysicsObjects[0].ID != in.PhysicsObjects[0].ID ||
		out.Joints[0].MaxMotorTorque != in.Joints[0].MaxMotorTorque {
		t.Error("decoded values do not match the originals")
	}
}

// TestSaveChunkAtomic checks SaveChunk writes a loadable file and leaves no
// temporary files behind.
func TestSaveChunkAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "world_0_0.png_petri")

	c := &Chunk{
		Version: CurrentVersion,
		Width:   2, Height: 2,
		Cells:         []byte{1, 1, 1, 1},
		MaterialNames: []string{"air", "wood"},
	}
	if err := SaveChunk(path, c); err != nil {
		t.Fatal(err)
	}
	got, err := LoadChunk(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.MaterialNames) != 2 || got.MaterialNames[1] != "wood" {
		t.Errorf("reloaded chunk lost its material names: %v", got.MaterialNames)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("expected only the saved file, found %d entries", len(entries))
	}
}

func TestDecodeContainerRejectsBadHeader(t *testing.T) {
	if _, err := DecodeContainer([]byte{1, 2, 3}); err == nil {
		t.Error("expected an error for a short file")
	}
	// compressedSize disagrees with the actual body length.
	bad := []byte{0xFF, 0, 0, 0, 0x10, 0, 0, 0, 1, 2, 3}
	if _, err := DecodeContainer(bad); err == nil {
		t.Error("expected an error for a mismatched compressedSize")
	}
}

func firstDiff(a, b []byte) int {
	n := min(len(a), len(b))
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return n
}
