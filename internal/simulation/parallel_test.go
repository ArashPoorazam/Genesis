package simulation

import (
	"fmt"
	"math/rand/v2"
	"runtime"
	"slices"
	"testing"
)

// serialStep retains the original scheduling and force accumulation order.
func serialStep(world *Simulation) {
	world.grid.rebuild(world.Particles)
	for index := range world.Particles {
		world.applyInfluence(&world.Particles[index])
	}
	for index := range world.Particles {
		particle := &world.Particles[index]
		particle.X = wrap(particle.X+particle.VX, world.Config.WorldWidth)
		particle.Y = wrap(particle.Y+particle.VY, world.Config.WorldHeight)
	}
}

func physicsWorkload(t testing.TB, population int, clustered bool) *Simulation {
	t.Helper()
	config := Default(42)
	for index := range config.ParticleTypes {
		config.ParticleTypes[index].Population = 0
	}
	world, _, err := New(config)
	if err != nil {
		t.Fatal(err)
	}
	random := rand.New(rand.NewPCG(42, 7))
	for index := range population {
		x, y := random.Float64()*config.WorldWidth, random.Float64()*config.WorldHeight
		if clustered {
			// A tight cluster across both world seams exercises wrapped interactions.
			x = wrap(random.Float64()*200-100, config.WorldWidth)
			y = wrap(random.Float64()*200-100, config.WorldHeight)
		}
		kind := index % len(config.ParticleTypes)
		world.Config.ParticleTypes[kind].Population++
		world.Particles = append(world.Particles, Particle{TypeIndex: kind, X: x, Y: y, VX: random.Float64()*2 - 1, VY: random.Float64()*2 - 1})
	}
	return world
}

func TestParallelPhysicsMatchesSerial(t *testing.T) {
	previous := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(previous)
	for _, cpus := range []int{1, 2, 4, 8} {
		runtime.GOMAXPROCS(cpus)
		for _, population := range []int{0, 1, 511, 512, 1537, MaxParticles} {
			t.Run(fmt.Sprintf("CPUs%d/Particles%d", cpus, population), func(t *testing.T) {
				current := physicsWorkload(t, population, population == 1537)
				reference := physicsWorkload(t, population, population == 1537)
				for step := range 6 {
					if step == 3 {
						// Live edits and a geometry rebuild must be visible to every worker.
						config := current.Config.Clone()
						config.Relationships[0][1] = -.75
						config.ParticleTypes[1].OuterRadius += 25
						config.Motion.VelocityRetention = .8
						config.WorldWidth = 900
						for _, world := range []*Simulation{current, reference} {
							if _, err := world.Apply(config); err != nil {
								t.Fatal(err)
							}
						}
					}
					if step >= 4 {
						// Cross the serial/parallel threshold after population edits.
						config := current.Config.Clone()
						for index := range config.ParticleTypes {
							config.ParticleTypes[index].Population = 0
						}
						config.ParticleTypes[0].Population = 100
						if step == 5 {
							config.ParticleTypes[0].Population = 600
						}
						for _, world := range []*Simulation{current, reference} {
							if message, err := world.Apply(config); err != nil || message != "" {
								t.Fatalf("population edit: %v %s", err, message)
							}
						}
					}
					current.Step()
					serialStep(reference)
					if !slices.Equal(current.Particles, reference.Particles) {
						t.Fatalf("physics differs from serial execution after step %d", step)
					}
				}
			})
		}
	}
}

func BenchmarkPhysicsScaling(b *testing.B) {
	for _, population := range []int{128, 512, 1500, MaxParticles} {
		for _, clustered := range []bool{false, true} {
			b.Run(fmt.Sprintf("Particles%d/Clustered%t", population, clustered), func(b *testing.B) {
				for _, implementation := range []struct {
					name string
					step func(*Simulation)
				}{
					{"Serial", serialStep}, {"Automatic", (*Simulation).Step},
				} {
					b.Run(implementation.name, func(b *testing.B) {
						world := physicsWorkload(b, population, clustered)
						initial := slices.Clone(world.Particles)
						world.grid.rebuild(world.Particles)
						b.ReportAllocs()
						b.ResetTimer()
						for range b.N {
							// Compare identical density even when the slower backend
							// completes fewer iterations. Both timings include this copy.
							copy(world.Particles, initial)
							implementation.step(world)
						}
					})
				}
			})
		}
	}
}
