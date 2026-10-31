package desktop

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ArashPoorazam/Genesis/internal/experiment"
	"github.com/ArashPoorazam/Genesis/internal/simulation"
	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

// This integration test runs the actual Ebitengine/Ebiten UI lifecycle on a display.
// Unit tests remain usable without graphics: GENESIS_UI_TEST=1 enables this test.
func TestDesktopWorkflow(t *testing.T) {
	if os.Getenv("GENESIS_UI_TEST") != "1" {
		t.Skip("set GENESIS_UI_TEST=1 to run desktop integration checks")
	}
	t.Setenv("EGL_PLATFORM", "x11")

	view, err := newWindow(Options{Width: 1024, Height: 768, TypeCount: 5, Seed: 42, UIScale: 1})
	if err != nil {
		t.Fatal(err)
	}
	view.store = experiment.Store{Directory: t.TempDir()}
	view.controller.Paused = true

	driver := &workflowDriver{view: view, width: 1024, height: 768}
	driver.steps = driver.workflowSteps(t.TempDir())
	ebiten.SetWindowSize(1024, 768)
	ebiten.SetWindowTitle("Genesis · UI verification")

	if err := ebiten.RunGame(driver); err != nil {
		t.Fatal(err)
	}
	if driver.failure != nil {
		t.Fatal(driver.failure)
	}
	if driver.stepIndex != len(driver.steps) {
		t.Fatal("desktop workflow ended before all checks ran")
	}
}

type workflowDriver struct {
	view          *Window
	width, height int
	frames        int
	stepIndex     int
	steps         []func() error
	failure       error
}

func (driver *workflowDriver) Layout(_, _ int) (int, int) {
	return driver.view.Layout(driver.width, driver.height)
}

func (driver *workflowDriver) Draw(screen *ebiten.Image) {
	driver.view.Draw(screen)
}

func (driver *workflowDriver) Update() error {
	if err := driver.view.Update(); err != nil {
		return err
	}
	driver.frames++
	if driver.frames%4 != 0 {
		return nil
	}
	if driver.stepIndex == len(driver.steps) {
		return ebiten.Termination
	}

	if err := driver.steps[driver.stepIndex](); err != nil {
		driver.failure = fmt.Errorf("workflow step %d: %w", driver.stepIndex+1, err)
		return ebiten.Termination
	}
	driver.stepIndex++
	return nil
}

func (driver *workflowDriver) click(caption string) error {
	for _, focusable := range driver.view.rootContainer.GetFocusers() {
		button, ok := focusable.(*widget.Button)
		if ok && button.Text().Label == caption {
			button.Click()
			event.ExecuteDeferred()
			return nil
		}
	}
	return fmt.Errorf("button %q not found", caption)
}

func (driver *workflowDriver) submit(index int, value string) error {
	if index >= len(driver.view.textFields) {
		return fmt.Errorf("input %d not found", index)
	}
	input := driver.view.textFields[index]
	input.SetText(value)
	input.SubmitEvent.Fire(&widget.TextInputChangedEventArgs{InputText: value})
	event.ExecuteDeferred()
	return nil
}

func (driver *workflowDriver) workflowSteps(exchangeDirectory string) []func() error {
	view := driver.view
	var originalParticles []simulation.Particle
	var savedConfig simulation.Config

	return []func() error{
		func() error {
			originalParticles = append([]simulation.Particle(nil), view.controller.Simulation.Particles...)
			original := view.controller.Simulation.Config.Relationships[0][0]
			view.ui.SetFocusedWidget(view.textFields[0])
			view.ui.ClearFocus()
			event.ExecuteDeferred()
			if view.controller.Simulation.Config.Relationships[0][0] != original {
				return fmt.Errorf("focusing an untouched input rounded the relationship")
			}
			view.selectedRow, view.selectedColumn = 0, 1
			view.requestRebuild()
			return nil
		},
		func() error { return driver.submit(0, "-1.25") },
		func() error {
			if view.controller.Simulation.Config.Relationships[0][1] != -1.25 {
				return fmt.Errorf("relationship input did not update the selected pair")
			}
			if !reflect.DeepEqual(originalParticles, view.controller.Simulation.Particles) {
				return fmt.Errorf("paused live edit moved particles")
			}
			return driver.submit(0, "NaN")
		},
		func() error {
			if view.controller.Simulation.Config.Relationships[0][1] != -1.25 || view.status == "" {
				return fmt.Errorf("invalid input did not show an error while preserving the rule")
			}
			return driver.click("Reciprocal edits: off")
		},
		func() error { return driver.submit(0, "0.75") },
		func() error {
			matrix := view.controller.Simulation.Config.Relationships
			if matrix[0][1] != 0.75 || matrix[1][0] != 0.75 {
				return fmt.Errorf("reciprocal editing failed")
			}
			return driver.click("Zero all")
		},
		func() error { return driver.click("Undo") },
		func() error {
			if view.controller.Simulation.Config.Relationships[0][1] != 0.75 {
				return fmt.Errorf("undo did not restore matrix")
			}
			return driver.click("Randomize relationships")
		},
		func() error {
			if !reflect.DeepEqual(originalParticles, view.controller.Simulation.Particles) {
				return fmt.Errorf("randomizing relationships moved particles")
			}
			return driver.click("Particles")
		},
		func() error { return driver.click("+ Add particle type") },
		func() error {
			if len(view.controller.Simulation.Config.ParticleTypes) != 6 {
				return fmt.Errorf("add type failed")
			}
			return driver.submit(0, "Violet")
		},
		func() error {
			view.textFields[2].SetText("12")
			return driver.click("Update population")
		},
		func() error {
			if len(view.controller.Simulation.Particles) != len(originalParticles)+12 {
				return fmt.Errorf("population update failed")
			}
			view.textFields[2].SetText("6000")
			return driver.click("Update population")
		},
		func() error {
			if len(view.controller.Simulation.Particles) != len(originalParticles)+12 || view.status == "" {
				return fmt.Errorf("population limit was not enforced")
			}
			view.scrollContainers[0].ScrollTop = 1
			return nil
		},
		func() error { return driver.click("Remove this type…") },
		func() error { return driver.click("Remove type") },
		func() error {
			if len(view.controller.Simulation.Config.ParticleTypes) != 5 {
				return fmt.Errorf("confirmed removal failed")
			}
			return driver.click("Save")
		},
		func() error {
			view.textFields[0].SetText("Integration")
			return driver.click("Save preset")
		},
		func() error {
			preset, err := view.store.Load("Integration")
			if err != nil {
				return err
			}
			savedConfig = preset.Config
			if view.controller.Dirty() {
				return fmt.Errorf("successful save left the experiment dirty")
			}
			return driver.click("Relationships")
		},
		func() error { return driver.click("Randomize relationships") },
		func() error {
			view.presetDialog()
			return nil
		},
		func() error { return driver.click("Integration") },
		func() error { return driver.click("Cancel") },
		func() error {
			if !view.controller.Dirty() {
				return fmt.Errorf("cancel discarded unsaved changes")
			}
			view.presetDialog()
			return nil
		},
		func() error { return driver.click("Integration") },
		func() error { return driver.click("Discard changes") },
		func() error {
			if !reflect.DeepEqual(view.controller.Simulation.Config, savedConfig) {
				return fmt.Errorf("loading did not restore the saved setup")
			}
			view.exportPreset(filepath.Join(exchangeDirectory, "export.json"))
			return nil
		},
		func() error {
			_, err := experiment.ReadPreset(filepath.Join(exchangeDirectory, "export.json"))
			if err != nil {
				return err
			}
			view.fileDialog(false, exchangeDirectory)
			return nil
		},
		func() error { return driver.click("export.json") },
		func() error { return driver.click("Import") },
		func() error {
			if !reflect.DeepEqual(view.controller.Simulation.Config, savedConfig) {
				return fmt.Errorf("file chooser import changed saved rules")
			}
			return driver.click("Settings")
		},
		func() error {
			lastInput := view.textFields[len(view.textFields)-1]
			view.ui.SetFocusedWidget(lastInput)
			event.ExecuteDeferred()
			if view.scrollContainers[0].ScrollTop <= 0 {
				return fmt.Errorf("keyboard focus did not reveal an offscreen input")
			}
			view.ui.ClearFocus()
			view.scrollContainers[0].ScrollTop = 1
			return nil
		},
		func() error { return driver.click("125%") },
		func() error {
			if view.scale != 1.25 {
				return fmt.Errorf("scale control failed")
			}
			driver.width, driver.height = 1440, 900
			return nil
		},
		func() error {
			view.inspectorVisible = false
			view.requestRebuild()
			return nil
		},
		func() error {
			if view.controller.Simulation.Config.WorldWidth != savedConfig.WorldWidth {
				return fmt.Errorf("viewport changes modified world dimensions")
			}
			return nil
		},
	}
}
