package simulation

import (
	"fmt"
	"math"
)

// VelocityInput is a borrowed, read-only view valid only during ComputeVelocities.
// Grid lists retain the CPU neighbor traversal order. No slice may be retained.
type VelocityInput struct {
	Config                Config
	Particles             []Particle
	Heads, Next           []int
	Columns, Rows         int
	CellWidth, CellHeight float64
}

type Velocity struct{ X, Y float64 }

// VelocityAccelerator computes a complete velocity snapshot without mutating input.
// Implementations run synchronously on the caller's goroutine.
type VelocityAccelerator interface {
	ComputeVelocities(VelocityInput, []Velocity) error
}

// StepWithAccelerator falls back to CPU physics on any acceleration error. A step
// always completes exactly once; a returned error reports that fallback occurred.
func (world *Simulation) StepWithAccelerator(accelerator VelocityAccelerator) error {
	world.grid.rebuild(world.Particles)
	var err error
	if accelerator != nil && len(world.Particles) > 0 {
		err = world.acceleratedVelocities(accelerator)
	}
	if accelerator == nil || err != nil {
		world.updateVelocities()
	}
	for index := range world.Particles {
		particle := &world.Particles[index]
		particle.X = wrap(particle.X+particle.VX, world.Config.WorldWidth)
		particle.Y = wrap(particle.Y+particle.VY, world.Config.WorldHeight)
	}
	return err
}

func (world *Simulation) acceleratedVelocities(accelerator VelocityAccelerator) error {
	if cap(world.accelerated) < len(world.Particles) {
		world.accelerated = make([]Velocity, len(world.Particles))
	}
	velocities := world.accelerated[:len(world.Particles)]
	// Incomplete results must fail validation instead of reusing an earlier step.
	for index := range velocities {
		velocities[index] = Velocity{math.NaN(), math.NaN()}
	}
	grid := world.grid
	input := VelocityInput{
		Config: world.Config, Particles: world.Particles,
		Heads: grid.heads, Next: grid.next, Columns: grid.columns, Rows: grid.rows,
		CellWidth: grid.cellWidth, CellHeight: grid.cellHeight,
	}
	if err := accelerator.ComputeVelocities(input, velocities); err != nil {
		return err
	}
	for _, velocity := range velocities {
		if math.IsNaN(velocity.X) || math.IsNaN(velocity.Y) || math.IsInf(velocity.X, 0) || math.IsInf(velocity.Y, 0) || math.Hypot(velocity.X, velocity.Y) > world.Config.Motion.SpeedLimit+0.001 {
			return fmt.Errorf("accelerator returned invalid velocity")
		}
	}
	for index, velocity := range velocities {
		world.Particles[index].VX, world.Particles[index].VY = velocity.X, velocity.Y
	}
	return nil
}
