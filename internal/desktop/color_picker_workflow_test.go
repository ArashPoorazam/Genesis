package desktop

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

func TestDesktopColorPicker(t *testing.T) {
	if os.Getenv("GENESIS_COLOR_UI_TEST") != "1" {
		t.Skip("set GENESIS_COLOR_UI_TEST=1 to check the custom color picker")
	}
	t.Setenv("EGL_PLATFORM", "x11")
	view, err := newWindow(Options{Width: 1024, Height: 768, TypeCount: 5, Seed: 42, UIScale: 1, SkipIntro: true, OpenInspector: true, InitialTab: particlesTab})
	if err != nil {
		t.Fatal(err)
	}
	view.controller.Paused = true
	driver := &workflowDriver{view: view, width: 1024, height: 768}
	original := view.controller.Simulation.Config.ParticleTypes[0].Color
	unchanged := func() error {
		if view.controller.Simulation.Config.ParticleTypes[0].Color != original {
			return fmt.Errorf("draft changed particle color before Apply")
		}
		return nil
	}
	capture := func(name string) {
		if directory := os.Getenv("GENESIS_COLOR_CAPTURE_DIR"); directory != "" {
			driver.capturePath = filepath.Join(directory, name+".png")
		}
	}
	driver.steps = []func() error{
		func() error { capture("particles-top"); return nil },
		func() error { view.scrollContainers[0].ScrollTop = .55; return nil },
		func() error { capture("particles-middle"); return nil },
		func() error { view.scrollContainers[0].ScrollTop = 1; return nil },
		func() error { capture("particles-bottom"); return nil },
		func() error { view.scrollContainers[0].ScrollTop = 0; return nil },
		func() error { return driver.click("Custom color…") },
		func() error { capture("picker-minimum"); return nil },
		func() error { return driver.submit(0, "#123456") },
		func() error {
			if err := unchanged(); err != nil {
				return err
			}
			driver.width = 1200
			return nil
		},
		func() error {
			if view.textFields[0].GetText() != "#123456" {
				return fmt.Errorf("resize lost draft")
			}
			return driver.click("Cancel")
		},
		func() error {
			if err := unchanged(); err != nil {
				return err
			}
			return driver.click("Custom color…")
		},
		func() error {
			view.textFields[0].SetText("invalid")
			return driver.click("Apply color")
		},
		func() error {
			if view.activeDialog == nil || view.status == "" {
				return fmt.Errorf("invalid hex was not rejected")
			}
			if err := unchanged(); err != nil {
				return err
			}
			capture("picker-error")
			return nil
		},
		func() error { return driver.submit(0, "#123456") },
		func() error {
			for _, focusable := range view.rootContainer.GetFocusers() {
				if slider, ok := focusable.(*widget.Slider); ok {
					slider.Current = 200
					return nil
				}
			}
			return fmt.Errorf("RGB slider missing")
		},
		func() error {
			if view.textFields[0].GetText() != "#C83456" {
				return fmt.Errorf("slider did not update hex: %s", view.textFields[0].GetText())
			}
			if err := unchanged(); err != nil {
				return err
			}
			driver.capturePath = os.Getenv("GENESIS_COLOR_SCREENSHOT")
			return nil
		},
		func() error { return driver.click("Apply color") },
		func() error {
			if view.activeDialog != nil || view.controller.Simulation.Config.ParticleTypes[0].Color != (color.NRGBA{200, 52, 86, 255}) {
				return fmt.Errorf("Apply did not save color")
			}
			return driver.click("Undo")
		},
		unchanged,
	}
	ebiten.SetWindowSize(1024, 768)
	if err := ebiten.RunGame(driver); err != nil {
		t.Fatal(err)
	}
	if driver.failure != nil {
		t.Fatal(driver.failure)
	}
	if driver.stepIndex != len(driver.steps) {
		t.Fatal("color picker workflow did not finish")
	}
}
