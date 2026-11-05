package simulation

import (
	"math"
	"reflect"
	"testing"
)

func TestStartupDefaults(t *testing.T) {
	for seed := uint64(0); seed < 100; seed++ {
		config := Default(seed)
		if err := config.Validate(); err != nil {
			t.Fatal(err)
		}
		if config.Motion.RepulsionStrength != 5 || config.Generation.RelationshipAmplitude != 1 || config.Generation.MinInnerRadius != 18 || config.Generation.MaxInnerRadius != 25 || config.Generation.MinOuterRadius != 40 || config.Generation.MaxOuterRadius != 120 {
			t.Fatal("incorrect startup defaults")
		}
		for _, particleType := range config.ParticleTypes {
			if particleType.ParticleRadius != 1 || particleType.InnerRadius < 18 || particleType.InnerRadius >= 25 || particleType.OuterRadius < 40 || particleType.OuterRadius >= 120 {
				t.Fatalf("incorrect particle defaults: %+v", particleType)
			}
		}
		for _, row := range config.Relationships {
			for _, strength := range row {
				if math.Abs(strength) > 1 {
					t.Fatal("relationship outside new defaults")
				}
			}
		}
	}
}

func TestRelationshipLocksSurviveEditsShuffleAndNewWorld(t *testing.T) {
	config := Default(42)
	config.SetRelationship(1, 2, .73)
	config.ToggleRelationshipLock(1, 2)
	config.SetRelationship(1, 2, -.5)
	if config.Relationships[1][2] != .73 {
		t.Fatal("manual edit changed locked cell")
	}
	for seed := uint64(0); seed < 100; seed++ {
		config.RandomizeRelations(seed)
		if config.Relationships[1][2] != .73 {
			t.Fatal("shuffle changed locked cell")
		}
		config.RandomizeWorld(seed)
		if config.Relationships[1][2] != .73 {
			t.Fatal("new world changed locked cell")
		}
	}
	config.ToggleRelationshipLock(1, 2)
	config.SetRelationship(1, 2, -.5)
	if config.Relationships[1][2] != -.5 {
		t.Fatal("unlock did not restore editing")
	}
	for row := range config.Relationships {
		for column := range config.Relationships[row] {
			config.ToggleRelationshipLock(row, column)
		}
	}
	before := config.Clone()
	for seed := uint64(0); seed < 100; seed++ {
		config.RandomizeRelations(seed)
	}
	if !reflect.DeepEqual(before, config) {
		t.Fatal("all-locked shuffle changed configuration")
	}
}

func TestNeutralShuffleIsOccasionalDeterministicAndRespectsLocks(t *testing.T) {
	original := Default(42)
	original.SetRelationship(0, 1, .6)
	original.ToggleRelationshipLock(0, 1)
	// A locked zero does not prevent this type from becoming neutral.
	original.SetRelationship(4, 4, 0)
	original.ToggleRelationshipLock(4, 4)
	neutralCount := 0
	choseZeroLocked := false
	for seed := uint64(0); seed < 500; seed++ {
		config, repeat := original.Clone(), original.Clone()
		neutral := config.RandomizeRelations(seed)
		if repeat.RandomizeRelations(seed) != neutral || !reflect.DeepEqual(config, repeat) {
			t.Fatal("shuffle is not reproducible")
		}
		if config.Relationships[0][1] != .6 || config.Relationships[4][4] != 0 {
			t.Fatal("neutral generation broke a lock")
		}
		if neutral < 0 {
			continue
		}
		neutralCount++
		if neutral == 0 || neutral == 1 {
			t.Fatal("selected a type with a locked nonzero connection")
		}
		choseZeroLocked = choseZeroLocked || neutral == 4
		for other := range config.ParticleTypes {
			if config.Relationships[neutral][other] != 0 || config.Relationships[other][neutral] != 0 {
				t.Fatal("neutral type is not zero in both directions")
			}
		}
	}
	if neutralCount < 50 || neutralCount > 150 || !choseZeroLocked {
		t.Fatalf("unexpected neutral selection over 500 seeds: %d; zero lock selected=%t", neutralCount, choseZeroLocked)
	}
	for index := range original.ParticleTypes {
		if original.RelationshipLocked(index, index) {
			original.ToggleRelationshipLock(index, index)
		}
		original.SetRelationship(index, index, .8)
		original.ToggleRelationshipLock(index, index)
	}
	for seed := uint64(0); seed < 100; seed++ {
		if original.RandomizeRelations(seed) >= 0 {
			t.Fatal("neutralized a type despite conflicting locks on every type")
		}
	}
}

func TestRelationshipLockCloneAndTypeChanges(t *testing.T) {
	config := Default(42)
	config.ToggleRelationshipLock(3, 4)
	clone := config.Clone()
	clone.ToggleRelationshipLock(3, 4)
	if !config.RelationshipLocked(3, 4) {
		t.Fatal("cloned lock matrix aliases original")
	}
	if err := config.RemoveType(1); err != nil {
		t.Fatal(err)
	}
	if !config.RelationshipLocked(2, 3) {
		t.Fatal("removing type did not remap locks")
	}
	if err := config.AddType("new"); err != nil {
		t.Fatal(err)
	}
	if err := config.Validate(); err != nil {
		t.Fatal(err)
	}
	for index := range config.ParticleTypes {
		if config.RelationshipLocked(index, 4) || config.RelationshipLocked(4, index) {
			t.Fatal("new type inherited a lock")
		}
	}
	if config.ParticleTypes[4].ParticleRadius != 1 {
		t.Fatal("new types did not use the new radius")
	}
	for _, locks := range [][][]bool{{{true}}, make([][]bool, 5)} {
		invalid := config.Clone()
		invalid.RelationshipLocks = locks
		if invalid.Validate() == nil {
			t.Fatal("accepted malformed lock matrix")
		}
	}
}
