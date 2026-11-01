package desktop

import (
	"reflect"
	"testing"

	"github.com/ArashPoorazam/Genesis/internal/experiment"
	"github.com/ArashPoorazam/Genesis/internal/simulation"
)

func TestLinkedGenerationRadiiStayValid(t *testing.T) {
	config := simulation.Default(42)
	setParticleRadius(&config, 0, 0, 20)
	for pass := 0; pass < 3; pass++ {
		for index := 0; index < 4; index++ {
			for _, value := range []float64{2, 20.1, 75, 499, 20, 1} {
				setGenerationRadius(&config, index, value)
				if err := config.Validate(); err != nil {
					t.Fatalf("range %d, value %g: %v", index, value, err)
				}
			}
		}
	}
}

func TestLinkedParticleRadiiStayValid(t *testing.T) {
	config := simulation.Default(42)
	for _, value := range []float64{20, .5, 12, 20} {
		setParticleRadius(&config, 0, 0, value)
		for _, inner := range []float64{499, 1, 20, 300} {
			setParticleRadius(&config, 0, 1, inner)
			for _, outer := range []float64{2, 500, 50, 3} {
				setParticleRadius(&config, 0, 2, outer)
				if err := config.Validate(); err != nil {
					t.Fatalf("particle %g, inner %g, outer %g: %v", value, inner, outer, err)
				}
			}
		}
	}
}

func TestLinkedGestureUndoesAllDependentChanges(t *testing.T) {
	controller, err := experiment.NewController(simulation.Default(42))
	if err != nil {
		t.Fatal(err)
	}
	original := controller.Simulation.Config.Clone()
	controller.BeginGesture()
	for _, value := range []float64{100, 200, 300} {
		if err := controller.Edit(func(config *simulation.Config) error {
			setGenerationRadius(config, 0, value)
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	controller.EndGesture()
	controller.Undo()
	if !reflect.DeepEqual(original, controller.Simulation.Config) || controller.CanUndo() {
		t.Fatal("one undo must restore the whole linked gesture")
	}
	controller.Redo()
	if controller.Simulation.Config.Generation.MinInnerRadius != 300 {
		t.Fatal("redo did not restore the linked edit")
	}
}
