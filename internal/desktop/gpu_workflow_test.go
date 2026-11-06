package desktop

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/ArashPoorazam/Genesis/internal/simulation"
	"github.com/hajimehoshi/ebiten/v2"
)

type failedGPU struct{}

func (failedGPU) ComputeVelocities(simulation.VelocityInput, []simulation.Velocity) error {
	return errors.New("simulated GPU loss")
}

func TestDesktopGPU(t *testing.T) {
	if os.Getenv("GENESIS_GPU_UI_TEST") != "1" {
		t.Skip("set GENESIS_GPU_UI_TEST=1 to verify GPU controls on a display")
	}
	t.Setenv("EGL_PLATFORM", "x11")
	view, err := newWindow(Options{Width: 1440, Height: 900, TypeCount: 5, Seed: 42, UIScale: 1, SkipIntro: true, OpenInspector: true, InitialTab: settingsTab, GPUBoost: true})
	if err != nil {
		t.Fatal(err)
	}
	defer view.closeGPU()
	view.controller.Paused = true
	driver := &workflowDriver{view: view, width: 1440, height: 900}
	driver.steps = []func() error{
		func() error {
			if view.controller.Accelerator == nil || view.gpuEngine == nil {
				return fmt.Errorf("startup did not enable GPU: %s", view.status)
			}
			driver.capturePath = os.Getenv("GENESIS_GPU_SCREENSHOT")
			return nil
		},
		func() error {
			if err := driver.click("On"); err != nil {
				return err
			}
			if view.controller.Accelerator != nil || view.gpuEngine != nil {
				return errors.New("Off did not release GPU")
			}
			return nil
		},
		func() error {
			if err := driver.click("Off"); err != nil {
				return err
			}
			if view.controller.Accelerator == nil {
				return fmt.Errorf("On did not enable GPU: %s", view.status)
			}
			view.controller.Restart()
			view.controller.Step()
			if view.controller.Accelerator == nil {
				return errors.New("GPU single step failed")
			}
			if err := view.controller.Load(view.controller.Simulation.Config, "GPU check"); err != nil {
				return err
			}
			view.controller.Step()
			if view.controller.Accelerator == nil {
				return errors.New("load lost GPU")
			}
			view.controller.Accelerator = failedGPU{}
			return driver.click("Step")
		},
		func() error {
			view.controller.Paused = true
			if view.controller.Accelerator != nil || view.gpuEngine != nil || !strings.Contains(view.controller.Message, "using CPU") {
				return errors.New("GPU failure did not disable acceleration and report CPU fallback")
			}
			// The failed On state must refresh to Off, allowing one-click retry.
			if err := driver.click("Off"); err != nil {
				return err
			}
			if view.controller.Accelerator == nil {
				return errors.New("retry failed")
			}
			return nil
		},
	}
	ebiten.SetWindowSize(1440, 900)
	if err := ebiten.RunGame(driver); err != nil {
		t.Fatal(err)
	}
	if driver.failure != nil {
		t.Fatal(driver.failure)
	}
	if driver.stepIndex != len(driver.steps) {
		t.Fatal("GPU workflow did not finish")
	}
}
