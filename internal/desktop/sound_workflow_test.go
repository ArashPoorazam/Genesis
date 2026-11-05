package desktop

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/ArashPoorazam/Genesis/internal/experiment"
	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

func TestDesktopSoundSettings(t *testing.T) {
	if os.Getenv("GENESIS_SOUND_UI_TEST") != "1" {
		t.Skip("set GENESIS_SOUND_UI_TEST=1 to run sound UI verification")
	}
	t.Setenv("EGL_PLATFORM", "x11")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	view, err := newWindow(Options{Width: 1024, Height: 768, TypeCount: 5, Seed: 42, UIScale: 1, SkipIntro: true, OpenInspector: true, InitialTab: settingsTab})
	if err != nil {
		t.Fatal(err)
	}
	view.audioPreferencesPath = filepath.Join(t.TempDir(), "audio.json")
	view.controller.Paused = true
	driver := &workflowDriver{view: view, width: 1024, height: 768}
	dirty := view.controller.Dirty()
	driver.steps = []func() error{
		func() error { return driver.click("Sound") },
		func() error {
			if view.settingsSection != 4 {
				return fmt.Errorf("sound tab did not open")
			}
			sliders := view.rootContainer.GetFocusers()
			count := 0
			for _, focuser := range sliders {
				if slider, ok := focuser.(*widget.Slider); ok {
					slider.Current = 23 + count*48
					count++
				}
			}
			event.ExecuteDeferred()
			if count != 2 {
				return fmt.Errorf("expected 2 volume sliders, got %d", count)
			}
			return nil
		},
		func() error {
			if view.audioPreferences.MusicVolume != .23 || view.audioPreferences.EffectsVolume != .71 {
				return fmt.Errorf("volume sliders did not update preferences: %+v", view.audioPreferences)
			}
			if view.controller.Dirty() != dirty || view.controller.CanUndo() {
				return fmt.Errorf("audio settings entered experiment history")
			}
			view.toggleMute()
			view.saveAudioPreferences()
			loaded, err := experiment.ReadAudioPreferences(view.audioPreferencesPath)
			if err != nil || !loaded.Muted || loaded.MusicVolume != .23 {
				return fmt.Errorf("sound preferences were not saved: %v", err)
			}
			return nil
		},
		func() error { return driver.click("Reset sound settings") },
		func() error {
			if view.audioPreferences != experiment.DefaultAudioPreferences() {
				return fmt.Errorf("reset did not restore audio defaults")
			}
			if directory := os.Getenv("GENESIS_SOUND_SCREENSHOTS"); directory != "" {
				driver.capturePath = filepath.Join(directory, "sound-settings.png")
			}
			return nil
		},
		func() error { return nil },
	}
	ebiten.SetWindowSize(1024, 768)
	ebiten.SetWindowTitle("Genesis · Sound settings verification")
	if err := ebiten.RunGame(driver); err != nil {
		t.Fatal(err)
	}
	if driver.failure != nil {
		t.Fatal(driver.failure)
	}
}
