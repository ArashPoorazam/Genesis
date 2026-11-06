package gpuphysics

import (
	"fmt"
	"math"

	"github.com/ArashPoorazam/Genesis/internal/simulation"
)

// verify exercises actual shader execution on the active device before enabling
// acceleration. It catches incorrect shader results as well as missing support.
func (engine *Engine) verify() error {
	for _, population := range []int{1, 16} {
		config := simulation.Default(42)
		config.WorldWidth, config.WorldHeight = 100, 100
		config.Generation.InitialSpeed = 1
		for index := range config.ParticleTypes {
			config.ParticleTypes[index].Population = 0
		}
		config.ParticleTypes[0].Population = population
		config.ParticleTypes[0].InnerRadius = 20
		config.ParticleTypes[0].OuterRadius = 80
		world, message, err := simulation.New(config)
		if err != nil || message != "" {
			return fmt.Errorf("GPU self-check setup: %v %s", err, message)
		}
		// Both worlds are independently generated from the same seed.
		reference, _, err := simulation.New(config)
		if err != nil {
			return err
		}
		reference.Step()
		if err := world.StepWithAccelerator(engine); err != nil {
			return err
		}
		for index, particle := range world.Particles {
			expected := reference.Particles[index]
			if math.Abs(particle.VX-expected.VX) > .002 || math.Abs(particle.VY-expected.VY) > .002 {
				return fmt.Errorf("GPU self-check failed; device produced inaccurate physics")
			}
		}
	}
	return nil
}
