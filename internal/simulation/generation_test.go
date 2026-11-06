package simulation

import (
	"fmt"
	"math"
	"testing"
)

func TestOuterRadiiUseFixedIntervalMidpoints(t *testing.T) {
	for _, test := range []struct {
		count            int
		minimum, maximum float64
		expected         []float64
	}{
		{1, 80, 120, []float64{100}},
		{4, 80, 120, []float64{85, 95, 105, 115}},
		{4, 80, 80, []float64{80, 80, 80, 80}},
		{5, 40, 120, []float64{48, 64, 80, 96, 112}},
		{12, 80, 200, []float64{85, 95, 105, 115, 125, 135, 145, 155, 165, 175, 185, 195}},
	} {
		t.Run(fmt.Sprintf("Types%d/%g-%g", test.count, test.minimum, test.maximum), func(t *testing.T) {
			config := Default(42)
			for len(config.ParticleTypes) > test.count {
				if err := config.RemoveType(len(config.ParticleTypes) - 1); err != nil {
					t.Fatal(err)
				}
			}
			for len(config.ParticleTypes) < test.count {
				if err := config.AddType(fmt.Sprintf("extra-%d", len(config.ParticleTypes))); err != nil {
					t.Fatal(err)
				}
			}
			config.Generation.MinOuterRadius, config.Generation.MaxOuterRadius = test.minimum, test.maximum
			for _, seed := range []uint64{0, 42, 2026} {
				config.RandomizeWorld(seed)
				if err := config.Validate(); err != nil {
					t.Fatal(err)
				}
				for index, kind := range config.ParticleTypes {
					if math.Abs(kind.OuterRadius-test.expected[index]) > 1e-12 {
						t.Fatalf("seed %d type %d: radius %g, want %g", seed, index, kind.OuterRadius, test.expected[index])
					}
				}
			}
		})
	}
}

func TestInnerRadiiUseFixedIntervalMidpoints(t *testing.T) {
	for _, count := range []int{1, 4, 12} {
		config := Default(42)
		for len(config.ParticleTypes) > count {
			if err := config.RemoveType(len(config.ParticleTypes) - 1); err != nil {
				t.Fatal(err)
			}
		}
		for len(config.ParticleTypes) < count {
			if err := config.AddType(fmt.Sprintf("extra-%d", len(config.ParticleTypes))); err != nil {
				t.Fatal(err)
			}
		}
		config.Generation.MinOuterRadius = 150
		config.Generation.MaxOuterRadius = 200
		for _, bounds := range [][2]float64{{80, 120}, {80, 80}} {
			config.Generation.MinInnerRadius, config.Generation.MaxInnerRadius = bounds[0], bounds[1]
			for _, seed := range []uint64{0, 42, 2026} {
				config.RandomizeWorld(seed)
				if err := config.Validate(); err != nil {
					t.Fatal(err)
				}
				for index, kind := range config.ParticleTypes {
					expected := bounds[0] + (float64(index)+.5)*(bounds[1]-bounds[0])/float64(count)
					if math.Abs(kind.InnerRadius-expected) > 1e-12 {
						t.Fatalf("%d types, seed %d: inner radius %g, want %g", count, seed, kind.InnerRadius, expected)
					}
				}
			}
		}
	}
}

func TestNewWorldIncludesAddedColorsWithinPopulationLimit(t *testing.T) {
	config := Default(42)
	for len(config.ParticleTypes) < MaxTypes {
		if err := config.AddType(fmt.Sprintf("new-%d", len(config.ParticleTypes))); err != nil {
			t.Fatal(err)
		}
	}
	for _, ceiling := range []int{1, 500, MaxParticles} {
		config.Generation.PopulationCeiling = ceiling
		for seed := uint64(0); seed < 100; seed++ {
			config.RandomizeWorld(seed)
			if err := config.Validate(); err != nil {
				t.Fatal(err)
			}
			for _, kind := range config.ParticleTypes {
				if kind.Population < 1 || kind.Population > ceiling {
					t.Fatalf("%s missing or over ceiling: %d", kind.ID, kind.Population)
				}
			}
		}
		world, message, err := New(config)
		if err != nil || message != "" {
			t.Fatalf("new world: %v %s", err, message)
		}
		counts := make([]int, len(config.ParticleTypes))
		for _, particle := range world.Particles {
			counts[particle.TypeIndex]++
		}
		for index, kind := range config.ParticleTypes {
			if counts[index] != kind.Population {
				t.Fatalf("%s placed %d, want %d", kind.ID, counts[index], kind.Population)
			}
		}
	}
}

func TestCrowdedSpawnAttemptsEveryColor(t *testing.T) {
	config := Default(42)
	config.WorldWidth, config.WorldHeight = 100, 100
	config.Generation.SpawnSpacing = 10
	for index := range config.ParticleTypes {
		config.ParticleTypes[index].Population = 500
	}
	world, message, err := New(config)
	if err != nil {
		t.Fatal(err)
	}
	if message == "" {
		t.Fatal("expected crowded-world warning")
	}
	counts := make([]int, len(config.ParticleTypes))
	for _, particle := range world.Particles {
		counts[particle.TypeIndex]++
	}
	for index, count := range counts {
		if count == 0 {
			t.Fatalf("type %d was starved by earlier populations", index)
		}
	}
}
