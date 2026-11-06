package simulation

import (
	"math"
	"reflect"
	"testing"
)

func TestStartupDefaults(t *testing.T) {
	for seed := uint64(0); seed < 100; seed++ {
		config := Default(seed)
		if config.Generation.PopulationCeiling != 500 {
			t.Fatal("incorrect random population ceiling")
		}
		if err := config.Validate(); err != nil {
			t.Fatal(err)
		}
		if config.Motion.RepulsionStrength != 5 || config.Generation.RelationshipAmplitude != 1 || config.Generation.MinInnerRadius != 18 || config.Generation.MaxInnerRadius != 25 || config.Generation.MinOuterRadius != 40 || config.Generation.MaxOuterRadius != 120 {
			t.Fatal("incorrect startup defaults")
		}
		for _, particleType := range config.ParticleTypes {
			if particleType.Population < 1 || particleType.Population > 500 {
				t.Fatal("population outside default ceiling")
			}
			if particleType.ParticleRadius != 1.5 || particleType.InnerRadius < 18 || particleType.InnerRadius >= 25 || particleType.OuterRadius < 40 || particleType.OuterRadius >= 120 {
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

func TestShuffleIsDeterministicWithoutForcedNeutralTypes(t *testing.T) {
	original := Default(42)
	original.SetRelationship(0, 1, .6)
	original.ToggleRelationshipLock(0, 1)
	original.SetRelationship(4, 4, 0)
	original.ToggleRelationshipLock(4, 4)
	for seed := uint64(0); seed < 500; seed++ {
		config, repeat := original.Clone(), original.Clone()
		config.RandomizeRelations(seed)
		repeat.RandomizeRelations(seed)
		if !reflect.DeepEqual(config, repeat) {
			t.Fatal("shuffle is not reproducible")
		}
		if config.Relationships[0][1] != .6 || config.Relationships[4][4] != 0 {
			t.Fatal("shuffle broke a lock")
		}
		for row, strengths := range config.Relationships {
			for column, strength := range strengths {
				if !config.RelationshipLocked(row, column) && strength == 0 {
					t.Fatalf("seed %d forced unlocked cell (%d,%d) to zero", seed, row, column)
				}
			}
		}
	}
	original.Generation.RelationshipAmplitude = 0
	original.RandomizeRelations(42)
	for row, strengths := range original.Relationships {
		for column, strength := range strengths {
			if !original.RelationshipLocked(row, column) && strength != 0 {
				t.Fatal("zero amplitude should still allow intentionally neutral relationships")
			}
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
	if config.ParticleTypes[4].ParticleRadius != 1.5 {
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
