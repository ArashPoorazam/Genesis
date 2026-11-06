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
