package simulation

import (
	"errors"
	"math"
	"slices"
	"testing"
)

type acceleratorFunc func(VelocityInput, []Velocity) error

func (compute acceleratorFunc) ComputeVelocities(input VelocityInput, output []Velocity) error {
	return compute(input, output)
}

func TestAccelerationFailureCompletesOneCPUStep(t *testing.T) {
	for _, failure := range []string{"error", "incomplete", "nan", "infinite", "speed"} {
		t.Run(failure, func(t *testing.T) {
			world := physicsWorkload(t, 600, false)
			expected := physicsWorkload(t, 600, false)
			expected.Step()
			accelerator := acceleratorFunc(func(input VelocityInput, output []Velocity) error {
				if failure == "incomplete" {
					output[0] = Velocity{}
					return nil
				}
				for index := range output {
					output[index] = Velocity{}
				}
				switch failure {
				case "error":
					return errors.New("device unavailable")
				case "nan":
					output[1].X = math.NaN()
				case "infinite":
					output[1].Y = math.Inf(1)
				case "speed":
					output[1].X = input.Config.Motion.SpeedLimit + 1
				}
				return nil
			})
			if err := world.StepWithAccelerator(accelerator); err == nil {
				t.Fatal("failed accelerator was accepted")
			}
			if !slices.Equal(world.Particles, expected.Particles) {
				t.Fatal("fallback did not reproduce exactly one CPU step")
			}
		})
	}
}

func TestAccelerationCommitsSnapshotAndWraps(t *testing.T) {
	world := physicsWorkload(t, 2, false)
	world.Particles[0].X = world.Config.WorldWidth - .25
	world.Particles[1].Y = .25
	before := slices.Clone(world.Particles)
	calls := 0
	accelerator := acceleratorFunc(func(input VelocityInput, output []Velocity) error {
		calls++
		if !slices.Equal(input.Particles, before) {
			t.Fatal("accelerator did not receive the old snapshot")
		}
		output[0] = Velocity{X: 1}
		output[1] = Velocity{Y: -1}
		return nil
	})
	if err := world.StepWithAccelerator(accelerator); err != nil {
		t.Fatal(err)
	}
	if calls != 1 || world.Particles[0].X != .75 || world.Particles[1].Y != world.Config.WorldHeight-.75 {
		t.Fatal("accelerated velocities were not committed and wrapped exactly once")
	}
}
