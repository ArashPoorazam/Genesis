package gpuphysics

import (
	"math"
	"slices"
	"testing"

	"github.com/ArashPoorazam/Genesis/internal/simulation"
)

func testEdgeCases(t *testing.T, engine *Engine) {
	for _, name := range []string{"one-cell", "two-cell", "half-world", "coincident", "zero-speed", "zero-retention", "tiny-speed", "maximum-grid", "twelve-types", "live-edits"} {
		t.Run(name, func(t *testing.T) {
			config := simulation.Default(42)
			if name == "twelve-types" {
				for len(config.ParticleTypes) < 12 {
					if err := config.AddType(string(rune('a' + len(config.ParticleTypes)))); err != nil {
						t.Fatal(err)
					}
				}
			}
			config.WorldWidth, config.WorldHeight = 100, 100
			for index := range config.ParticleTypes {
				config.ParticleTypes[index].Population = 0
				config.ParticleTypes[index].InnerRadius = 20
				config.ParticleTypes[index].OuterRadius = 80
				for other := range config.ParticleTypes {
					config.Relationships[index][other] = float64((index+other)%3 - 1)
				}
			}
			config.Relationships[0][1] = 0 // Neutral relationships still repel nearby particles.
			config.Relationships[1][0] = .7
			particles := []simulation.Particle{
				{TypeIndex: 0, X: .25, Y: .25, VX: .3, VY: -.4},
				{TypeIndex: 1, X: 99.75, Y: 99.75, VX: -.2, VY: .1},
				{TypeIndex: 2, X: 30, Y: 40, VX: .5},
			}
			switch name {
			case "two-cell":
				config.WorldWidth, config.WorldHeight = 200, 300
			case "half-world":
				particles[0].X = 25
				particles[1].X = 75
				particles[0].Y = 50
				particles[1].Y = 50
			case "coincident":
				particles[1].X = particles[0].X
				particles[1].Y = particles[0].Y
			case "zero-speed":
				config.Motion.SpeedLimit = 0
			case "zero-retention":
				config.Motion.VelocityRetention = 0
			case "tiny-speed":
				config.Motion.SpeedLimit = .000001
			case "maximum-grid":
				config.WorldWidth, config.WorldHeight = 10000, 10000
				for index := range config.ParticleTypes {
					config.ParticleTypes[index].InnerRadius = 1.6
					config.ParticleTypes[index].OuterRadius = 2
				}
				particles[1].X = 9999.75
				particles[1].Y = 9999.75
			case "twelve-types":
				particles[0].TypeIndex = 11
				particles[1].TypeIndex = 10
			}
			world, _, err := simulation.New(config)
			if err != nil {
				t.Fatal(err)
			}
			world.Particles = particles
			for _, particle := range particles {
				world.Config.ParticleTypes[particle.TypeIndex].Population++
			}
			for step := range 4 {
				if name == "live-edits" && step > 0 {
					next := world.Config.Clone()
					next.WorldWidth = 200
					next.ParticleTypes[0].OuterRadius += 10
					next.Relationships[0][1] = -.75
					next.ParticleTypes[0].Population = step * 10
					// Removing types must remap both relationships and GPU type indices.
					if step == 3 {
						if err := next.RemoveType(1); err != nil {
							t.Fatal(err)
						}
					}
					if message, err := world.Apply(next); err != nil || message != "" {
						t.Fatalf("edit: %v %s", err, message)
					}
				}
				initial := slices.Clone(world.Particles)
				world.Step()
				expected := slices.Clone(world.Particles)
				copy(world.Particles, initial)
				if err := world.StepWithAccelerator(engine); err != nil {
					t.Fatal(err)
				}
				for index, particle := range world.Particles {
					if math.Abs(particle.VX-expected[index].VX) > .002 || math.Abs(particle.VY-expected[index].VY) > .002 {
						t.Fatalf("step %d particle %d: GPU (%g,%g), CPU (%g,%g)", step, index, particle.VX, particle.VY, expected[index].VX, expected[index].VY)
					}
					if math.Hypot(particle.VX, particle.VY) > world.Config.Motion.SpeedLimit+1e-12 {
						t.Fatal("speed cap exceeded")
					}
				}
			}
		})
	}
}
