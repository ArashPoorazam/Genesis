package experiment

import (
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/ArashPoorazam/Genesis/internal/simulation"
)

func newTestController(t *testing.T) *Controller {
	t.Helper()
	controller, err := NewController(simulation.Default(42))
	if err != nil {
		t.Fatal(err)
	}
	return controller
}

func TestDirectedEditPreservesMotion(t *testing.T) {
	controller := newTestController(t)
	controller.Simulation.Step()
	before := append([]simulation.Particle(nil), controller.Simulation.Particles...)
	reverse := controller.Simulation.Config.Relationships[1][0]
	if err := controller.Edit(func(config *simulation.Config) error {
		config.Relationships[0][1] = 1.25
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if controller.Simulation.Config.Relationships[1][0] != reverse {
		t.Fatal("reverse relationship changed")
	}
	if !reflect.DeepEqual(before, controller.Simulation.Particles) {
		t.Fatal("editing forces moved particles")
	}
}

func TestHistoryGroupsGestureAndDoesNotRewind(t *testing.T) {
	controller := newTestController(t)
	original := controller.Simulation.Config.Clone()
	controller.BeginGesture()
	for _, strength := range []float64{.2, .4, .8} {
		if err := controller.Edit(func(config *simulation.Config) error {
			config.Relationships[0][1] = strength
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	controller.EndGesture()
	if len(controller.undo) != 1 {
		t.Fatalf("drag produced %d history entries", len(controller.undo))
	}
	controller.Simulation.Step()
	positions := append([]simulation.Particle(nil), controller.Simulation.Particles...)
	controller.Undo()
	if !reflect.DeepEqual(original, controller.Simulation.Config) || !reflect.DeepEqual(positions, controller.Simulation.Particles) {
		t.Fatal("undo did not restore rules while retaining motion")
	}
	controller.Redo()
	if controller.Simulation.Config.Relationships[0][1] != .8 {
		t.Fatal("redo failed")
	}
	controller.MarkSaved("test")
	if controller.Dirty() {
		t.Fatal("saved setup is dirty")
	}
	controller.Undo()
	if !controller.Dirty() {
		t.Fatal("undo should differ from saved baseline")
	}
}

func TestTypeRemovalRemapsParticlesAndRelations(t *testing.T) {
	controller := newTestController(t)
	before := controller.Simulation.Config.Clone()
	particles := append([]simulation.Particle(nil), controller.Simulation.Particles...)
	if err := controller.Edit(func(config *simulation.Config) error {
		return config.RemoveType(1)
	}); err != nil {
		t.Fatal(err)
	}
	want := []simulation.Particle{}
	for _, particle := range particles {
		if particle.TypeIndex == 1 {
			continue
		}
		if particle.TypeIndex > 1 {
			particle.TypeIndex--
		}
		want = append(want, particle)
	}
	if !reflect.DeepEqual(want, controller.Simulation.Particles) {
		t.Fatal("surviving particles were not preserved")
	}
	for rowIndex, previousRow := range []int{0, 2, 3, 4} {
		for columnIndex, previousColumn := range []int{0, 2, 3, 4} {
			if controller.Simulation.Config.Relationships[rowIndex][columnIndex] != before.Relationships[previousRow][previousColumn] {
				t.Fatal("surviving relation changed")
			}
		}
	}
	if err := controller.Edit(func(config *simulation.Config) error {
		return config.AddType("new")
	}); err != nil {
		t.Fatal(err)
	}
	lastTypeIndex := len(controller.Simulation.Config.ParticleTypes) - 1
	if controller.Simulation.Config.ParticleTypes[lastTypeIndex].Population != 0 {
		t.Fatal("new type populated")
	}
	for rowIndex := range controller.Simulation.Config.ParticleTypes {
		if controller.Simulation.Config.Relationships[rowIndex][lastTypeIndex] != 0 || controller.Simulation.Config.Relationships[lastTypeIndex][rowIndex] != 0 {
			t.Fatal("new relationships are not neutral")
		}
	}
}

func TestSeededRestartAndPopulationEdits(t *testing.T) {
	controller := newTestController(t)
	initial := append([]simulation.Particle(nil), controller.Simulation.Particles...)
	for range 10 {
		controller.Simulation.Step()
	}
	controller.Restart()
	if !reflect.DeepEqual(initial, controller.Simulation.Particles) {
		t.Fatal("seeded restart differs")
	}
	if err := controller.Edit(func(config *simulation.Config) error {
		config.ParticleTypes[0].Population += 10
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(initial, controller.Simulation.Particles[:len(initial)]) {
		t.Fatal("population increase moved existing particles")
	}
	if len(controller.Simulation.Particles) != len(initial)+10 {
		t.Fatal("population was not increased")
	}
	for rowIndex := range controller.Simulation.Config.ParticleTypes {
		if err := controller.Edit(func(config *simulation.Config) error {
			config.ParticleTypes[rowIndex].Population = 0
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	if len(controller.Simulation.Particles) != 0 {
		t.Fatal("zero population not applied")
	}
}

func TestInvalidEditsAreAtomic(t *testing.T) {
	for name, edit := range map[string]func(*simulation.Config){
		"nan": func(config *simulation.Config) {
			config.Relationships[0][0] = math.NaN()
		},
		"inner radius": func(config *simulation.Config) {
			config.ParticleTypes[0].InnerRadius = config.ParticleTypes[0].OuterRadius
		},
		"population limit": func(config *simulation.Config) {
			config.ParticleTypes[0].Population = 5000
		},
		"matrix dimensions": func(config *simulation.Config) {
			config.Relationships = config.Relationships[:2]
		},
		"duplicate IDs": func(config *simulation.Config) {
			config.ParticleTypes[1].ID = config.ParticleTypes[0].ID
		},
		"inverted generation bounds": func(config *simulation.Config) {
			config.Generation.MinInnerRadius = 100
		},
		"version": func(config *simulation.Config) {
			config.Version = 99
		},
	} {
		t.Run(name, func(t *testing.T) {
			controller := newTestController(t)
			before := controller.Simulation.Config.Clone()
			ps := append([]simulation.Particle(nil), controller.Simulation.Particles...)
			if err := controller.Edit(func(config *simulation.Config) error {
				edit(config)
				return nil
			}); err == nil {
				t.Fatal("invalid edit accepted")
			}
			if !reflect.DeepEqual(before, controller.Simulation.Config) || !reflect.DeepEqual(ps, controller.Simulation.Particles) || controller.CanUndo() {
				t.Fatal("failed edit changed state")
			}
		})
	}
}

func TestPlaybackFixedSteps(t *testing.T) {
	controller := newTestController(t)
	reference := newTestController(t)
	controller.Speed = .25
	for range 4 {
		controller.Tick()
	}
	reference.Simulation.Step()
	if !reflect.DeepEqual(controller.Simulation.Particles, reference.Simulation.Particles) {
		t.Fatal("quarter speed did not run exactly one fixed step")
	}
	controller.Paused = true
	before := append([]simulation.Particle(nil), controller.Simulation.Particles...)
	controller.Tick()
	if !reflect.DeepEqual(before, controller.Simulation.Particles) {
		t.Fatal("pause changed motion")
	}
	controller.Paused = false
	controller.Speed = 4
	controller.Tick()
	for range 4 {
		reference.Simulation.Step()
	}
	if !reflect.DeepEqual(controller.Simulation.Particles, reference.Simulation.Particles) {
		t.Fatal("4x speed changed step size")
	}
}

func TestRandomizationRespectsBoundsAndSurvivingMotion(t *testing.T) {
	controller := newTestController(t)
	before := append([]simulation.Particle(nil), controller.Simulation.Particles...)
	types := append([]simulation.ParticleType(nil), controller.Simulation.Config.ParticleTypes...)
	if err := controller.Edit(func(config *simulation.Config) error {
		config.RandomizeRelations(7)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(types, controller.Simulation.Config.ParticleTypes) || !reflect.DeepEqual(before, controller.Simulation.Particles) {
		t.Fatal("relationship randomization modified world")
	}
	config := simulation.Default(1)
	for len(config.ParticleTypes) < simulation.MaxTypes {
		if err := config.AddType(strings.Repeat("x", len(config.ParticleTypes))); err != nil {
			t.Fatal(err)
		}
	}
	if err := config.AddType("overflow"); err == nil {
		t.Fatal("accepted more than 12 types")
	}
	config.Generation.PopulationCeiling = 5000
	config.RandomizeWorld(8)
	if err := config.Validate(); err != nil {
		t.Fatal(err)
	}
	regenerated := config.Clone()
	regenerated.RandomizeWorld(8)
	if !reflect.DeepEqual(config, regenerated) {
		t.Fatal("world generation not reproducible")
	}
}
