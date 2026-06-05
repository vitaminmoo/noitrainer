package noitadata

import (
	"slices"
	"testing"
)

func TestParseZombie(t *testing.T) {
	n := mustFS(t)
	e, err := ParseEntity(n, "data/entities/animals/zombie.xml")
	if err != nil {
		t.Fatal(err)
	}
	if e.Name != "$animal_zombie" {
		t.Errorf("Name = %q, want $animal_zombie", e.Name)
	}
	if !slices.Contains(e.Tags, "zombie") {
		t.Errorf("Tags missing 'zombie': %v", e.Tags)
	}
	// zombie.xml has one <Base> block and two direct components.
	if len(e.rawBaseFiles) != 1 {
		t.Fatalf("want 1 base block, got %d", len(e.rawBaseFiles))
	}
	if e.rawBaseFiles[0].path != "data/entities/base_enemy_basic.xml" {
		t.Errorf("unexpected base path %q", e.rawBaseFiles[0].path)
	}
	if len(e.directComponents) != 2 {
		t.Errorf("want 2 direct components, got %d", len(e.directComponents))
	}
}

func TestResolveZombieWeakMergesChain(t *testing.T) {
	n := mustFS(t)
	e, err := ResolveEntity(n, "data/entities/animals/zombie_weak.xml")
	if err != nil {
		t.Fatal(err)
	}
	if !e.Resolved {
		t.Fatal("entity should be resolved")
	}
	// Chain must at minimum include zombie.xml (the immediate <Base>) and
	// walk beyond into base_enemy_basic.xml.
	for _, want := range []string{
		"data/entities/animals/zombie.xml",
		"data/entities/base_enemy_basic.xml",
	} {
		if !slices.Contains(e.BaseChain, want) {
			t.Errorf("BaseChain missing %q; got %v", want, e.BaseChain)
		}
	}
	// zombie_weak.xml overrides DamageModelComponent.hp = 0.2.
	var dm *Component
	for i := range e.Components {
		if e.Components[i].Type == "DamageModelComponent" {
			dm = &e.Components[i]
			break
		}
	}
	if dm == nil {
		t.Fatal("DamageModelComponent missing after resolve")
	}
	if dm.Attrs["hp"] != "0.2" {
		t.Errorf("hp = %q, want 0.2 (override should have won)", dm.Attrs["hp"])
	}
}

func TestEntityRefsIncludeBaseAndAttrs(t *testing.T) {
	n := mustFS(t)
	e, err := ResolveEntity(n, "data/entities/animals/zombie.xml")
	if err != nil {
		t.Fatal(err)
	}
	refs := e.Refs()
	if len(refs) == 0 {
		t.Fatal("expected some refs")
	}
	// base_enemy_basic is in the chain.
	if !slices.Contains(refs, "data/entities/base_enemy_basic.xml") {
		t.Error("refs should include base_enemy_basic.xml")
	}
	// sprite image_file is referenced.
	if !slices.Contains(refs, "data/enemies_gfx/zombie.xml") {
		t.Error("refs should include data/enemies_gfx/zombie.xml")
	}
	// audio bank file referenced (from direct AudioComponent).
	if !slices.Contains(refs, "data/audio/Desktop/animals.bank") {
		t.Error("refs should include animals.bank")
	}
}

func TestLooksLikePath(t *testing.T) {
	for _, c := range []struct {
		in   string
		want bool
	}{
		{"data/entities/foo.xml", true},
		{"foo/bar.lua", true},
		{"foo.xml", false}, // bare filename; needs slash
		{"scale_y", false},
		{"1.0", false},
		{"", false},
	} {
		if got := looksLikePath(c.in); got != c.want {
			t.Errorf("looksLikePath(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}
