package simulation

import (
	"fmt"
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

func TestPhysicsMatchesOriginalLoopWithOriginalMotion(t *testing.T) {
	config := originalMotionConfig()
	current, _, err := New(config)
	if err != nil {
		t.Fatal(err)
	}
	legacy, _, err := New(config)
	if err != nil {
		t.Fatal(err)
	}

	for step := range 20 {
		current.Step()
		legacyStep(legacy)
		for index, particle := range current.Particles {
			reference := legacy.Particles[index]
			if particle.TypeIndex != reference.TypeIndex {
				t.Fatalf("step %d particle %d changed type", step, index)
			}
			// Go permits fused floating-point operations. The historical loop's
			// constants and the runtime settings can round differently on ARM64,
			// especially with race instrumentation. Compare each step from the same
			// snapshot so tiny rounding differences cannot amplify across steps.
			for _, component := range []struct {
				name             string
				actual, expected float64
			}{
				{"X", particle.X, reference.X}, {"Y", particle.Y, reference.Y},
				{"VX", particle.VX, reference.VX}, {"VY", particle.VY, reference.VY},
			} {
				tolerance := 1e-12 * max(1, math.Abs(component.expected))
				if !(math.Abs(component.actual-component.expected) <= tolerance) {
					t.Fatalf("step %d particle %d %s: got %.17g, want %.17g (tolerance %g)",
						step, index, component.name, component.actual, component.expected, tolerance)
				}
			}
		}
		copy(legacy.Particles, current.Particles)
	}
}

func originalMotionConfig() Config {
	config := Default(42)
	// Keep the historical force-law regression independent of startup tuning.
	config.Motion = Motion{VelocityRetention: .90, SpeedLimit: 5, RepulsionStrength: 10}
	return config
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
			world, _, err := New(originalMotionConfig())
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

// Relationship drags should cost roughly the same regardless of population.
func BenchmarkRelationshipEdit(b *testing.B) {
	for _, population := range []int{0, MaxParticles} {
		b.Run(fmt.Sprintf("Particles%d", population), func(b *testing.B) {
			config := Default(42)
			config.WorldWidth, config.WorldHeight = 10000, 10000
			for index := range config.ParticleTypes {
				config.ParticleTypes[index].Population = 0
			}
			config.ParticleTypes[0].Population = population
			world, message, err := New(config)
			if err != nil || message != "" {
				b.Fatalf("create benchmark world: %v %s", err, message)
			}
			b.ReportAllocs()
			b.ResetTimer()

			for iteration := range b.N {
				config.Relationships[0][1] = float64(iteration % 2)
				if _, err := world.Apply(config); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
