package gpuphysics

import (
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"slices"
	"sort"
	"testing"
	"time"

	"github.com/ArashPoorazam/Genesis/internal/simulation"
	"github.com/hajimehoshi/ebiten/v2"
)

func TestShaderCompiles(t *testing.T) {
	shader, err := ebiten.NewShader(shaderSource)
	if err != nil {
		t.Fatal(err)
	}
	shader.Deallocate()
}

func TestGPUPhysics(t *testing.T) {
	if os.Getenv("GENESIS_GPU_TEST") != "1" {
		t.Skip("set GENESIS_GPU_TEST=1 to run actual GPU correctness and performance checks")
	}
	t.Setenv("EGL_PLATFORM", "x11")
	driver := &gpuTestGame{test: t}
	ebiten.SetWindowSize(320, 200)
	ebiten.SetVsyncEnabled(false)
	if err := ebiten.RunGame(driver); err != nil {
		t.Fatal(err)
	}
	if !driver.completed {
		t.Fatal("GPU checks did not finish")
	}
}

type gpuTestGame struct {
	test      *testing.T
	completed bool
}

func (*gpuTestGame) Layout(int, int) (int, int) { return 320, 200 }
func (*gpuTestGame) Draw(*ebiten.Image)         {}
func (game *gpuTestGame) Update() error {
	if game.completed {
		return ebiten.Termination
	}
	engine, err := New()
	if err != nil {
		game.test.Error(err)
		game.completed = true
		return ebiten.Termination
	}
	defer engine.Close()
	for _, population := range []int{0, 1, 2, 512, 1500, 5000} {
		for _, clustered := range []bool{false, true} {
			game.test.Run(fmt.Sprintf("Particles%d/Clustered%t", population, clustered), func(t *testing.T) {
				world := testWorld(t, population, clustered)
				initial := slices.Clone(world.Particles)
				for step := range 4 {
					copy(world.Particles, initial)
					world.Step()
					expected := slices.Clone(world.Particles)
					copy(world.Particles, initial)
					if err := world.StepWithAccelerator(engine); err != nil {
						t.Fatal(err)
					}
					maximum := 0.0
					for index, particle := range world.Particles {
						maximum = max(maximum, math.Abs(particle.VX-expected[index].VX), math.Abs(particle.VY-expected[index].VY))
						if particle.TypeIndex != expected[index].TypeIndex || particle.X < 0 || particle.X >= world.Config.WorldWidth || particle.Y < 0 || particle.Y >= world.Config.WorldHeight {
							t.Fatal("invalid particle state")
						}
					}
					if maximum > .02 {
						t.Logf("first particle: actual %+v; expected %+v", world.Particles[0], expected[0])
						t.Fatalf("step %d velocity error %.6f exceeds 0.02 world units/step", step, maximum)
					}
					t.Logf("step %d maximum velocity error %.6f", step, maximum)
					initial = slices.Clone(world.Particles)
				}
				if population >= 1500 {
					timings := make(map[string]float64)
					for _, gpu := range []bool{false, true} {
						var samples []float64
						for sample := range 13 {
							copy(world.Particles, initial)
							started := time.Now()
							if gpu {
								if err := world.StepWithAccelerator(engine); err != nil {
									t.Fatal(err)
								}
							} else {
								world.Step()
							}
							elapsed := float64(time.Since(started).Nanoseconds()) / 1e6
							if sample >= 3 {
								samples = append(samples, elapsed)
							}
						}
						sort.Float64s(samples)
						name := "CPU"
						if gpu {
							name = "GPU"
						}
						timings[name] = samples[len(samples)/2]
					}
					t.Logf("complete step incl. upload/readback: CPU %.3f ms, GPU %.3f ms (%.2fx)", timings["CPU"], timings["GPU"], timings["CPU"]/timings["GPU"])
				}
			})
		}
	}
	game.test.Run("EdgeCases", func(t *testing.T) { testEdgeCases(t, engine) })
	game.completed = true
	return nil
}

func testWorld(t *testing.T, population int, clustered bool) *simulation.Simulation {
	t.Helper()
	config := simulation.Default(42)
	for index := range config.ParticleTypes {
		config.ParticleTypes[index].Population = 0
	}
	world, _, err := simulation.New(config)
	if err != nil {
		t.Fatal(err)
	}
	random := rand.New(rand.NewPCG(42, 7))
	for index := range population {
		x, y := random.Float64()*config.WorldWidth, random.Float64()*config.WorldHeight
		if clustered {
			x = math.Mod(random.Float64()*200-100+config.WorldWidth, config.WorldWidth)
			y = math.Mod(random.Float64()*200-100+config.WorldHeight, config.WorldHeight)
		}
		kind := index % len(config.ParticleTypes)
		world.Config.ParticleTypes[kind].Population++
		world.Particles = append(world.Particles, simulation.Particle{TypeIndex: kind, X: x, Y: y, VX: random.Float64()*2 - 1, VY: random.Float64()*2 - 1})
	}
	return world
}
