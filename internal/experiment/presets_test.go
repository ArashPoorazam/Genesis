package experiment

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/ArashPoorazam/Genesis/internal/simulation"
)

func TestPresetLifecycleAndImportValidation(t *testing.T) {
	store := Store{t.TempDir()}
	config := simulation.Default(42)
	if err := store.Save("Orbit", config); err != nil {
		t.Fatal(err)
	}
	preset, err := store.Load("Orbit")
	if err != nil || !reflect.DeepEqual(config, preset.Config) {
		t.Fatalf("preset round trip: %v", err)
	}
	if err = store.Rename("Orbit", "Nebula"); err != nil {
		t.Fatal(err)
	}
	if store.Exists("Orbit") || !store.Exists("Nebula") {
		t.Fatal("rename failed")
	}
	if err = store.Save("Other", config); err != nil {
		t.Fatal(err)
	}
	if err = store.Rename("Nebula", "Other"); err == nil {
		t.Fatal("rename overwrote a preset")
	}
	for _, bad := range []string{"../escape", "", "a/b", "a\\b", ".."} {
		if err = store.Save(bad, config); err == nil {
			t.Fatalf("unsafe name accepted: %q", bad)
		}
	}
	path := filepath.Join(t.TempDir(), "export.json")
	if err = WritePreset(path, preset); err != nil {
		t.Fatal(err)
	}
	if _, err = ReadPreset(path); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, []byte(`{"name":"Broken","config":{"version":99}}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = ReadPreset(path); err == nil {
		t.Fatal("invalid imported preset accepted")
	}
	if err = store.Delete("Nebula"); err != nil {
		t.Fatal(err)
	}
	if store.Exists("Nebula") {
		t.Fatal("delete failed")
	}
}

func TestOversizedSavePreservesReadablePreset(t *testing.T) {
	store := Store{Directory: t.TempDir()}
	original := simulation.Default(42)
	if err := store.Save("Keep", original); err != nil {
		t.Fatal(err)
	}

	oversized := original.Clone()
	// IDs are opaque strings without a schema length limit, so a valid config
	// can still exceed the persistence format's byte budget.
	oversized.ParticleTypes[0].ID = strings.Repeat("x", maxPresetBytes)
	if err := oversized.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := store.Save("Keep", oversized); err == nil {
		t.Fatal("saved a preset that the reader would reject")
	}
	preset, err := store.Load("Keep")
	if err != nil || !reflect.DeepEqual(original, preset.Config) {
		t.Fatalf("oversized save replaced the readable preset: %v", err)
	}
}

func TestFailedSaveAndLoadPreserveCurrentState(t *testing.T) {
	controller := newTestController(t)
	before := controller.Simulation.Config.Clone()
	particles := append([]simulation.Particle(nil), controller.Simulation.Particles...)
	bad := before.Clone()
	bad.Relationships = nil
	if err := controller.Load(bad, "Invalid"); err == nil {
		t.Fatal("invalid load succeeded")
	}
	if !reflect.DeepEqual(before, controller.Simulation.Config) || !reflect.DeepEqual(particles, controller.Simulation.Particles) {
		t.Fatal("failed load changed running world")
	}
	directory := t.TempDir()
	block := filepath.Join(directory, "file")
	if err := os.WriteFile(block, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := (Store{block}).Save("Fail", before); err == nil {
		t.Fatal("expected write failure")
	}
	data, _ := os.ReadFile(block)
	if string(data) != "keep" {
		t.Fatal("failed save damaged existing file")
	}
	store := Store{directory}
	if err := store.Save("Good", before); err != nil {
		t.Fatal(err)
	}
	if err := store.Save("Good", bad); err == nil {
		t.Fatal("invalid save succeeded")
	}
	preset, err := store.Load("Good")
	if err != nil || !reflect.DeepEqual(preset.Config, before) {
		t.Fatal("failed save replaced valid preset")
	}
}

func TestPresetRejectsTrailingUnknownAndOversizedData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "preset.json")
	preset := Preset{"Test", simulation.Default(1)}
	if err := WritePreset(path, preset); err != nil {
		t.Fatal(err)
	}
	valid, _ := os.ReadFile(path)
	for _, data := range [][]byte{append(append([]byte(nil), valid...), []byte("{}")...), []byte(`{"unexpected":true}`), make([]byte, (1<<20)+1)} {
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := ReadPreset(path); err == nil {
			t.Fatal("malformed preset accepted")
		}
	}
}

func TestPresetPreservesLocksAndAcceptsOlderPresets(t *testing.T) {
	store := Store{t.TempDir()}
	config := simulation.Default(42)
	// No lock field is serialized for older-style, fully editable presets.
	if err := store.Save("Older", config); err != nil {
		t.Fatal(err)
	}
	older, err := store.Load("Older")
	if err != nil || older.Config.RelationshipLocked(0, 1) {
		t.Fatalf("older preset: %v", err)
	}
	config.ToggleRelationshipLock(0, 1)
	if err := store.Save("Locked", config); err != nil {
		t.Fatal(err)
	}
	loaded, err := store.Load("Locked")
	if err != nil || !reflect.DeepEqual(config, loaded.Config) {
		t.Fatalf("locked preset round trip: %v", err)
	}
}
