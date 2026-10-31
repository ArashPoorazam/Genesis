// Package experiment coordinates simulation edits, playback, history, and presets.
package experiment

import (
	"reflect"

	"github.com/ArashPoorazam/Genesis/internal/simulation"
)

const historyLimit = 100

// Controller owns configuration history. History never snapshots simulation time.
type Controller struct {
	Simulation *simulation.Simulation
	Paused     bool
	Speed      float64
	Message    string
	Name       string

	baseline    simulation.Config
	undo        []simulation.Config
	redo        []simulation.Config
	gesture     *simulation.Config
	accumulator float64
}

func NewController(config simulation.Config) (*Controller, error) {
	world, message, err := simulation.New(config)
	if err != nil {
		return nil, err
	}

	return &Controller{
		Simulation: world,
		Speed:      1,
		Message:    message,
		baseline:   config.Clone(),
	}, nil
}

func (controller *Controller) Dirty() bool {
	return !reflect.DeepEqual(controller.baseline, controller.Simulation.Config)
}

func (controller *Controller) MarkSaved(name string) {
	controller.Name = name
	controller.baseline = controller.Simulation.Config.Clone()
}

func (controller *Controller) CanUndo() bool {
	return len(controller.undo) > 0
}

func (controller *Controller) CanRedo() bool {
	return len(controller.redo) > 0
}

// BeginGesture groups successive slider changes into one undo entry.
func (controller *Controller) BeginGesture() {
	if controller.gesture != nil {
		return
	}

	config := controller.Simulation.Config.Clone()
	controller.gesture = &config
}

func (controller *Controller) EndGesture() {
	if controller.gesture == nil {
		return
	}

	if !reflect.DeepEqual(*controller.gesture, controller.Simulation.Config) {
		controller.pushUndo(*controller.gesture)
		controller.redo = nil
	}
	controller.gesture = nil
}

func (controller *Controller) pushUndo(config simulation.Config) {
	controller.undo = append(controller.undo, config)
	if len(controller.undo) > historyLimit {
		controller.undo = controller.undo[1:]
	}
}

func (controller *Controller) Edit(change func(*simulation.Config) error) error {
	next := controller.Simulation.Config.Clone()
	if err := change(&next); err != nil {
		return err
	}
	if err := next.Validate(); err != nil {
		return err
	}
	if reflect.DeepEqual(next, controller.Simulation.Config) {
		return nil
	}

	previous := controller.Simulation.Config.Clone()
	message, err := controller.Simulation.Apply(next)
	if err != nil {
		return err
	}

	if controller.gesture == nil {
		controller.pushUndo(previous)
		controller.redo = nil
	}
	controller.Message = message

	return nil
}

func (controller *Controller) Undo() {
	controller.EndGesture()
	if !controller.CanUndo() {
		return
	}

	lastIndex := len(controller.undo) - 1
	previous := controller.undo[lastIndex]
	controller.undo = controller.undo[:lastIndex]
	controller.redo = append(controller.redo, controller.Simulation.Config.Clone())

	// History contains only configurations that previously passed validation.
	controller.Message, _ = controller.Simulation.Apply(previous)
}

func (controller *Controller) Redo() {
	controller.EndGesture()
	if !controller.CanRedo() {
		return
	}

	lastIndex := len(controller.redo) - 1
	next := controller.redo[lastIndex]
	controller.redo = controller.redo[:lastIndex]
	controller.pushUndo(controller.Simulation.Config.Clone())
	controller.Message, _ = controller.Simulation.Apply(next)
}

func (controller *Controller) Load(config simulation.Config, name string) error {
	world, message, err := simulation.New(config)
	if err != nil {
		return err
	}

	controller.Simulation = world
	controller.Message = message
	controller.undo = nil
	controller.redo = nil
	controller.gesture = nil
	controller.accumulator = 0
	controller.MarkSaved(name)

	return nil
}

func (controller *Controller) Restart() {
	controller.Message = controller.Simulation.Restart()
	controller.accumulator = 0
}

func (controller *Controller) Tick() {
	if controller.Paused {
		controller.accumulator = 0
		return
	}

	controller.accumulator += controller.Speed
	for controller.accumulator >= 1 {
		controller.Simulation.Step()
		controller.accumulator--
	}
}
