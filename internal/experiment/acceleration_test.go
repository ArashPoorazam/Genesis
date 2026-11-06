package experiment

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/ArashPoorazam/Genesis/internal/simulation"
)

type testAccelerator struct {
	calls int
	fail  bool
}

func (accelerator *testAccelerator) ComputeVelocities(_ simulation.VelocityInput, output []simulation.Velocity) error {
	accelerator.calls++
	if accelerator.fail {
		return errors.New("test failure")
	}
	for index := range output {
		output[index] = simulation.Velocity{}
	}
	return nil
}

func TestControllerAccelerationLifecycle(t *testing.T) {
	controller := newTestController(t)
	accelerator := &testAccelerator{}
	controller.Accelerator = accelerator
	controller.Paused = true
	controller.Tick()
	if accelerator.calls != 0 {
		t.Fatal("paused tick ran physics")
	}
	controller.Step()
	if accelerator.calls != 1 {
		t.Fatal("single step bypassed acceleration")
	}
	controller.Restart()
	config := controller.Simulation.Config.Clone()
	if err := controller.Load(config, "test"); err != nil {
		t.Fatal(err)
	}
	controller.Paused = false
	controller.Speed = 2
	controller.Tick()
	if accelerator.calls != 3 {
		t.Fatal("load or playback speed lost acceleration")
	}
	reference, _, err := simulation.New(config)
	if err != nil {
		t.Fatal(err)
	}
	reference.Particles = slices.Clone(controller.Simulation.Particles)
	reference.Step()
	accelerator.fail = true
	controller.Step()
	if !slices.Equal(controller.Simulation.Particles, reference.Particles) {
		t.Fatal("failed step was skipped or advanced twice")
	}
	if controller.Accelerator != nil || !strings.Contains(controller.Message, "using CPU") {
		t.Fatal("acceleration failure was not surfaced and disabled")
	}
	controller.Step()
	if accelerator.calls != 4 {
		t.Fatal("failed accelerator was retried")
	}
}
