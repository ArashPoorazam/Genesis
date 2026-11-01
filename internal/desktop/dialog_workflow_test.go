package desktop

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ArashPoorazam/Genesis/internal/experiment"
	"github.com/ArashPoorazam/Genesis/internal/simulation"
	"github.com/hajimehoshi/ebiten/v2"
)

// Run separately from other display tests: Ebitengine allows one lifecycle.
func TestDesktopDialogs(t *testing.T) {
	if os.Getenv("GENESIS_DIALOG_TEST") != "1" {
		t.Skip("set GENESIS_DIALOG_TEST=1 to verify save and exit dialogs")
	}
	t.Setenv("EGL_PLATFORM", "x11")
	view, err := newWindow(Options{Width: 1024, Height: 768, TypeCount: 5, Seed: 42, UIScale: 1, SkipIntro: true})
	if err != nil {
		t.Fatal(err)
	}
	view.controller.Paused = true
	view.store = experiment.Store{Directory: t.TempDir()}
	driver := &dialogWorkflowDriver{workflowDriver: &workflowDriver{view: view, width: 1024, height: 768}}
	driver.steps = driver.dialogSteps(t.TempDir())
	ebiten.SetWindowSize(1024, 768)
	ebiten.SetWindowTitle("Genesis · Dialog verification")
	if err := ebiten.RunGame(driver); err != nil {
		t.Fatal(err)
	}
	if driver.failure != nil {
		t.Fatal(driver.failure)
	}
	if driver.stepIndex != len(driver.steps) {
		t.Fatal("dialog workflow ended before all checks ran")
	}
}

type dialogWorkflowDriver struct {
	*workflowDriver
	captureName   string
	captureFrames int
}

func (driver *dialogWorkflowDriver) Draw(screen *ebiten.Image) {
	driver.view.Draw(screen)
	if driver.captureName == "" || driver.view.needsRebuild {
		return
	}
	// Let live labels refresh after a scripted input change before capture.
	driver.captureFrames++
	if driver.captureFrames < 2 {
		return
	}
	if directory := os.Getenv("GENESIS_DIALOG_SCREENSHOTS"); directory != "" {
		if err := savePNG(screen, filepath.Join(directory, driver.captureName+".png")); err != nil {
			driver.failure = err
			driver.view.quit = true
		}
	}
	driver.captureName = ""
	driver.captureFrames = 0
}

func (driver *dialogWorkflowDriver) dialogSteps(failureDirectory string) []func() error {
	view := driver.view
	store := view.store
	change := func() error {
		return view.editConfig(func(config *simulation.Config) error {
			config.Relationships[0][1] += .1
			return nil
		})
	}
	var beforeDiscard simulation.Config
	return []func() error{
		func() error { view.save(true, nil); return nil },
		func() error {
			if view.dialogRect().Dx() != 576 || view.ui.GetFocusedWidget() != view.textFields[0] {
				return fmt.Errorf("save dialog is not compact or its name field lacks focus")
			}
			view.textFields[0].SetText("Orbital garden")
			driver.captureName = "save"
			return nil
		},
		func() error { return driver.submit(0, "") },
		func() error {
			if view.activeDialog == nil || view.status == "" {
				return fmt.Errorf("empty name did not leave a visible validation error")
			}
			driver.captureName = "save-validation"
			return nil
		},
		func() error { return driver.submit(0, "Orbital garden") },
		func() error {
			if view.activeDialog != nil || view.controller.Dirty() || !store.Exists("Orbital garden") {
				return fmt.Errorf("Enter did not save the named preset")
			}
			if err := change(); err != nil {
				return err
			}
			view.requestExit()
			driver.captureName = "exit"
			return nil
		},
		func() error { return driver.click("Cancel") },
		func() error {
			if view.quit || view.activeDialog != nil || !view.controller.Dirty() {
				return fmt.Errorf("canceling exit lost the running experiment")
			}
			view.requestExit()
			return nil
		},
		func() error {
			if err := driver.click("Save & exit"); err != nil {
				return err
			}
			if !view.quit || view.controller.Dirty() {
				return fmt.Errorf("save and exit did not finish")
			}
			view.quit = false
			preset, err := store.Load("Orbital garden")
			if err != nil {
				return err
			}
			beforeDiscard = preset.Config
			if !reflect.DeepEqual(preset.Config, view.controller.Simulation.Config) {
				return fmt.Errorf("exit saved the wrong configuration")
			}
			if err := change(); err != nil {
				return err
			}
			view.requestExit()
			return nil
		},
		func() error {
			if err := driver.click("Exit without saving"); err != nil {
				return err
			}
			if !view.quit {
				return fmt.Errorf("exit without saving did not close")
			}
			view.quit = false
			preset, err := store.Load("Orbital garden")
			if err != nil {
				return err
			}
			if !reflect.DeepEqual(preset.Config, beforeDiscard) {
				return fmt.Errorf("discard overwrote the saved preset")
			}
			view.controller.Name = ""
			view.requestExit()
			return nil
		},
		func() error { return driver.click("Save & exit") },
		func() error { return driver.click("Cancel") },
		func() error {
			if view.quit || view.activeDialog == nil || view.activeDialog.title != "Save before you leave?" {
				return fmt.Errorf("cancel save did not return to the exit decision")
			}
			return driver.click("Save & exit")
		},
		func() error { return driver.submit(0, "Orbital garden") },
		func() error {
			driver.captureName = "replace"
			return nil
		},
		func() error { return driver.click("Go back") },
		func() error {
			if view.textFields[0].GetText() != "Orbital garden" {
				return fmt.Errorf("overwrite cancellation lost the entered name")
			}
			return driver.submit(0, "Orbital garden")
		},
		func() error {
			blocked := filepath.Join(failureDirectory, "not-a-directory")
			if err := os.WriteFile(blocked, []byte("file"), 0600); err != nil {
				return err
			}
			view.store.Directory = blocked
			return driver.click("Replace")
		},
		func() error {
			if view.quit || view.activeDialog == nil || view.status == "" || !view.controller.Dirty() {
				return fmt.Errorf("failed save exited or hid the error")
			}
			driver.captureName = "save-failure"
			return nil
		},
		func() error {
			view.store = store
			if err := driver.click("Replace"); err != nil {
				return err
			}
			if !view.quit || view.controller.Dirty() {
				return fmt.Errorf("successful retry did not resume the original exit")
			}
			view.quit = false
			view.requestExit()
			if !view.quit || view.activeDialog != nil {
				return fmt.Errorf("clean exit unexpectedly asked to save")
			}
			view.quit = false
			return nil
		},
	}
}
