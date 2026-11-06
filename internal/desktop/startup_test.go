package desktop

import "testing"

func TestStartupUsesRequestedTypeCountForOuterRadii(t *testing.T) {
	config, err := initialConfig(Options{TypeCount: 4, Seed: 42})
	if err != nil {
		t.Fatal(err)
	}
	for index, expected := range []float64{50, 70, 90, 110} {
		if config.ParticleTypes[index].OuterRadius != expected {
			t.Fatalf("type %d outer radius = %g, want %g", index, config.ParticleTypes[index].OuterRadius, expected)
		}
	}
	if config.Generation.PopulationCeiling != 500 {
		t.Fatal("startup population ceiling must be 500")
	}
}
