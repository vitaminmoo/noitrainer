package noitadata

import (
	"slices"
	"testing"
)

func TestIndexHasZombie(t *testing.T) {
	n := mustFS(t)
	idx, err := n.Index()
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(idx.Entities, "data/entities/animals/zombie.xml") {
		t.Error("zombie.xml missing from index.Entities")
	}
}

func TestIndexOutboundAndInbound(t *testing.T) {
	n := mustFS(t)
	idx, err := n.Index()
	if err != nil {
		t.Fatal(err)
	}
	// zombie.xml should reference base_enemy_basic.xml.
	if !slices.Contains(idx.Outbound["data/entities/animals/zombie.xml"],
		"data/entities/base_enemy_basic.xml") {
		t.Error("zombie.xml outbound missing base_enemy_basic")
	}
	// base_enemy_basic.xml should have at least one inbound (zombie).
	if !slices.Contains(idx.Inbound["data/entities/base_enemy_basic.xml"],
		"data/entities/animals/zombie.xml") {
		t.Error("base_enemy_basic.xml inbound missing zombie")
	}
}

func TestIndexEntitiesWithComponent(t *testing.T) {
	n := mustFS(t)
	idx, err := n.Index()
	if err != nil {
		t.Fatal(err)
	}
	// DamageModelComponent should be present on many entities including
	// zombie.xml via its <Base> override.
	files := idx.Components["DamageModelComponent"]
	if len(files) < 10 {
		t.Errorf("expected many DamageModelComponent entities, got %d", len(files))
	}
	if !slices.Contains(files, "data/entities/animals/zombie.xml") {
		t.Error("zombie.xml missing from DamageModelComponent index")
	}
}

func TestFindEntitiesWithAttr(t *testing.T) {
	n := mustFS(t)
	// Find entities whose <Base> override of DamageModelComponent sets
	// blood_material="blood_fading".
	hits, err := n.FindEntitiesWith("DamageModelComponent", map[string]string{
		"blood_material": "blood_fading",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) == 0 {
		t.Fatal("expected at least one match")
	}
	if !slices.Contains(hits, "data/entities/animals/zombie.xml") {
		t.Error("zombie.xml should match blood_material=blood_fading")
	}
}
