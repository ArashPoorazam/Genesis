package simulation

import (
	"math"
	"testing"
)

// legacyStep preserves the original pre-interface loop for regression comparison.
// Both benchmarks use the same seeded particles, matrix, grid, and default constants.
func legacyStep(world *Simulation) {
	world.grid.rebuild(world.Particles)

	for index := range world.Particles {
		particle := &world.Particles[index]
		velocityX, velocityY := particle.VX, particle.VY
		cell := world.grid.cell(particle.X, particle.Y)

		for _, neighborCell := range world.grid.neighbors[cell][:world.grid.neighborCount] {
			for otherIndex := world.grid.heads[neighborCell]; otherIndex != -1; otherIndex = world.grid.next[otherIndex] {
				other := &world.Particles[otherIndex]
				innerRadius := world.Config.ParticleTypes[other.TypeIndex].InnerRadius
				outerRadius := world.Config.ParticleTypes[other.TypeIndex].OuterRadius
				deltaX := wrappedDelta(other.X-particle.X, world.Config.WorldWidth)
				deltaY := wrappedDelta(other.Y-particle.Y, world.Config.WorldHeight)
				distanceSquared := deltaX*deltaX + deltaY*deltaY

				if distanceSquared == 0 || distanceSquared >= world.outerSquared[other.TypeIndex] {
					continue
				}

				distance := math.Sqrt(distanceSquared)
				relationship := world.Config.Relationships[particle.TypeIndex][other.TypeIndex]
				force := Force(distance, innerRadius, outerRadius, relationship, 10)
				velocityX += deltaX / distance * force
				velocityY += deltaY / distance * force
			}
		}

		velocityX *= 0.90
		velocityY *= 0.90
		if speedSquared := velocityX*velocityX + velocityY*velocityY; speedSquared > 25 {
			scale := 5 / math.Sqrt(speedSquared)
			velocityX *= scale
			velocityY *= scale
		}
		particle.VX, particle.VY = velocityX, velocityY
	}

	for index := range world.Particles {
		particle := &world.Particles[index]
		particle.X = wrap(particle.X+particle.VX, world.Config.WorldWidth)
		particle.Y = wrap(particle.Y+particle.VY, world.Config.WorldHeight)
	}
}

func TestDefaultPhysicsMatchesOriginalLoop(t *testing.T) {
	current, _, err := New(Default(42))
	if err != nil {
		t.Fatal(err)
	}
	legacy, _, err := New(Default(42))
	if err != nil {
		t.Fatal(err)
	}

	for range 20 {
		current.Step()
		legacyStep(legacy)
	}
	for index, particle := range current.Particles {
		if particle != legacy.Particles[index] {
			t.Fatalf("particle %d differs from the original physics loop", index)
		}
	}
}

func BenchmarkPhysics(b *testing.B) {
	implementations := []struct {
		name string
		step func(*Simulation)
	}{
		{name: "OriginalLoop", step: legacyStep},
		{name: "RuntimeSettings", step: (*Simulation).Step},
	}

	for _, implementation := range implementations {
		b.Run(implementation.name, func(b *testing.B) {
			world, _, err := New(Default(42))
			if err != nil {
				b.Fatal(err)
			}
			initial := append([]Particle(nil), world.Particles...)
			b.ReportAllocs()
			b.ResetTimer()

			for iteration := range b.N {
				// Repeat the same evolution window to avoid cluster density bias.
				if iteration%120 == 0 {
					copy(world.Particles, initial)
				}
				implementation.step(world)
			}
		})
	}
}
