package desktop

import (
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestDesktopNewWorldIncludesAddedColor(t *testing.T) {
	if os.Getenv("GENESIS_WORLD_UI_TEST") != "1" {
		t.Skip("set GENESIS_WORLD_UI_TEST=1 to check adding colors and generating a world")
	}
	t.Setenv("EGL_PLATFORM", "x11")
	view, err := newWindow(Options{Width: 1024, Height: 768, TypeCount: 5, Seed: 42, UIScale: 1, SkipIntro: true, OpenInspector: true, InitialTab: particlesTab})
	if err != nil {
		t.Fatal(err)
	}
	view.controller.Paused = true
	driver := &workflowDriver{view: view, width: 1024, height: 768}
	addedID := ""
	driver.steps = []func() error{
		func() error { return driver.click("+ Add particle type") },
		func() error {
			config := view.controller.Simulation.Config
			if len(config.ParticleTypes) != 6 {
				return fmt.Errorf("type was not added")
			}
			addedID = config.ParticleTypes[5].ID
			return driver.click("New world")
		},
		func() error { return driver.click("Discard changes") },
		func() error {
			world := view.controller.Simulation
			config := world.Config
			if len(config.ParticleTypes) != 6 || config.ParticleTypes[5].ID != addedID {
				return fmt.Errorf("new world lost added color")
			}
			counts := make([]int, 6)
			for _, particle := range world.Particles {
				counts[particle.TypeIndex]++
			}
			for index, kind := range config.ParticleTypes {
				if kind.Population < 1 || counts[index] != kind.Population {
					return fmt.Errorf("type %s: requested %d, placed %d", kind.ID, kind.Population, counts[index])
				}
				inner := config.Generation.MinInnerRadius + (float64(index)+.5)*(config.Generation.MaxInnerRadius-config.Generation.MinInnerRadius)/6
				outer := config.Generation.MinOuterRadius + (float64(index)+.5)*(config.Generation.MaxOuterRadius-config.Generation.MinOuterRadius)/6
				if math.Abs(kind.InnerRadius-inner) > 1e-12 || math.Abs(kind.OuterRadius-outer) > 1e-12 {
					return fmt.Errorf("type %s has incorrect midpoint radii", kind.ID)
				}
			}
			return nil
		},
	}
	ebiten.SetWindowSize(1024, 768)
	if err := ebiten.RunGame(driver); err != nil {
		t.Fatal(err)
	}
	if driver.failure != nil {
		t.Fatal(driver.failure)
	}
	if driver.stepIndex != len(driver.steps) {
		t.Fatal("new world workflow did not finish")
	}
}
