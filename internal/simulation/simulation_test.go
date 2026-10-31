package simulation

import (
	"reflect"
	"strings"
	"testing"
)

func TestInfluencingTypeRadiusAndWrappedEdges(t *testing.T) {
	config := Default(1)
	for i := range config.ParticleTypes {
		config.ParticleTypes[i].Population = 0
		for j := range config.ParticleTypes {
			config.Relationships[i][j] = 0
		}
	}
	config.ParticleTypes[0].OuterRadius = 25
	config.ParticleTypes[1].OuterRadius = 100
	config.Relationships[0][1] = 1
	config.Motion.VelocityRetention = 1
	world, _, err := New(config)
	if err != nil {
		t.Fatal(err)
	}
	world.Particles = []Particle{{TypeIndex: 0, X: 5, Y: 100}, {TypeIndex: 1, X: config.WorldWidth - 45, Y: 100}}
	world.Step()
	if world.Particles[0].VX >= 0 {
		t.Fatal("wrapped neighbor should attract left using its 100 radius")
	}
	if world.Particles[1].VX != 0 {
		t.Fatal("reverse relationship should remain neutral")
	}
	if _, err = world.Apply(func() Config {
		updated := config.Clone()
		updated.ParticleTypes[1].OuterRadius = 30
		return updated
	}()); err != nil {
		t.Fatal(err)
	}
	if world.outerSquared[1] != 900 {
		t.Fatal("radius cache stale")
	}
}

func TestForceLaw(t *testing.T) {
	if Force(5, 10, 50, 0, 10) != -5 {
		t.Fatal("neutral relationship removed inner repulsion")
	}
	if Force(30, 10, 50, 2, 10) != 2 {
		t.Fatal("triangle peak changed")
	}
	if Force(50, 10, 50, 2, 10) != 0 {
		t.Fatal("force outside outer radius")
	}
}

func TestGridHandlesOneAndTwoCells(t *testing.T) {
	for _, size := range []float64{100, 200, 300} {
		grid := newGrid(size, size, 100)
		for _, neighbors := range grid.neighbors {
			seen := map[int]bool{}
			for _, neighbor := range neighbors[:grid.neighborCount] {
				if seen[neighbor] {
					t.Fatal("duplicate wrapped neighbor")
				}
				seen[neighbor] = true
			}
		}
	}
}

func TestCrowdedSpawnIsBoundedAndUnrelatedEditsDoNotRetry(t *testing.T) {
	config := Default(1)
	config.WorldWidth, config.WorldHeight = 100, 100
	for i := range config.ParticleTypes {
		config.ParticleTypes[i].Population = 0
	}
	config.ParticleTypes[0].Population = 5000
	config.Generation.SpawnSpacing = 50
	world, message, err := New(config)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(message, "exhausted") {
		t.Fatal("expected spawn failure feedback")
	}
	before := append([]Particle(nil), world.Particles...)
	config.Relationships[0][1] = 0
	if _, err = world.Apply(config); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, world.Particles) {
		t.Fatal("unrelated edit retried exhausted spawn")
	}
}
