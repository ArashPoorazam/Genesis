package desktop

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/ArashPoorazam/Genesis/internal/experiment"
	"github.com/ArashPoorazam/Genesis/internal/soundtrack"
)

func (view *Window) startAudio(options Options) {
	view.audioDisabled = options.Mute || options.ScreenshotPath != ""
	if view.audioDisabled {
		return
	}
	view.audioPreferencesPath = filepath.Join(filepath.Dir(view.store.Directory), "audio.json")
	preferences, err := experiment.ReadAudioPreferences(view.audioPreferencesPath)
	if err != nil {
		log.Printf("Audio preferences: %v; using defaults", err)
	}
	view.audioPreferences = preferences
	view.sound.SetPreferences(preferences)
	if err := view.sound.Start(); err != nil {
		view.status = "Audio unavailable: " + err.Error()
		view.audioDisabled = true
	}
}

func (view *Window) changeAudio(change func(*experiment.AudioPreferences)) {
	if view.audioDisabled {
		return
	}
	change(&view.audioPreferences)
	view.sound.SetPreferences(view.audioPreferences)
	view.audioSaveAt = time.Now().Add(600 * time.Millisecond)
}

func (view *Window) toggleMute() {
	view.changeAudio(func(preferences *experiment.AudioPreferences) { preferences.Muted = !preferences.Muted })
	view.requestRebuild()
}

func (view *Window) updateAudio() {
	now := time.Now()
	view.sound.Update(now)
	if !view.audioSaveAt.IsZero() && !now.Before(view.audioSaveAt) {
		view.saveAudioPreferences()
	}
}

func (view *Window) saveAudioPreferences() {
	view.audioSaveAt = time.Time{}
	if view.audioPreferencesPath == "" {
		return
	}
	if err := experiment.WriteAudioPreferences(view.audioPreferencesPath, view.audioPreferences); err != nil {
		view.reportError(fmt.Errorf("could not save audio preferences: %w", err))
		log.Print(view.status)
	}
}

func (view *Window) closeAudio() {
	if !view.audioSaveAt.IsZero() {
		view.saveAudioPreferences()
	}
	view.sound.Close()
}

func (view *Window) togglePlayback() {
	view.controller.Paused = !view.controller.Paused
	if view.controller.Paused {
		view.sound.Queue(soundtrack.Pause)
	} else {
		view.sound.Queue(soundtrack.Play)
	}
}

func (view *Window) restartParticles() {
	view.controller.Restart()
	view.status = "Restarted from the current seed"
	view.sound.Queue(soundtrack.NewWorld)
}
