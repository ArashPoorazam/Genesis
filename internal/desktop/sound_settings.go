package desktop

import (
	"fmt"

	"github.com/ArashPoorazam/Genesis/internal/experiment"
	"github.com/ebitenui/ebitenui/widget"
)

func (view *Window) buildSoundSettings(content *widget.Container, width int) {
	view.addLabel(content, 0, 0, width, "QUIET ORBITS · HANDPAN", 14, ink)
	if view.audioDisabled {
		view.addLabel(content, 0, 36, width, "Audio is disabled for this run.", 14, muted)
		return
	}
	view.toggle(content, 0, 36, width, "Mute all audio", view.audioPreferences.Muted, func(value bool) {
		view.changeAudio(func(preferences *experiment.AudioPreferences) { preferences.Muted = value })
	})
	view.toggle(content, 0, 88, width, "Background music", view.audioPreferences.MusicEnabled, func(value bool) {
		view.changeAudio(func(preferences *experiment.AudioPreferences) { preferences.MusicEnabled = value })
	})
	view.addAudioVolume(content, 134, width, "Music volume", true)
	view.toggle(content, 0, 210, width, "Interaction sounds", view.audioPreferences.EffectsEnabled, func(value bool) {
		view.changeAudio(func(preferences *experiment.AudioPreferences) { preferences.EffectsEnabled = value })
	})
	view.addAudioVolume(content, 256, width, "Effects volume", false)
	view.addLabel(content, 0, 334, width, "M mutes all audio. Your levels are remembered.\nMusic continues while the simulation is paused.", 13, muted)
	view.addButton(content, 0, 384, width, "Reset sound settings", false, func() {
		view.changeAudio(func(preferences *experiment.AudioPreferences) { *preferences = experiment.DefaultAudioPreferences() })
		view.requestRebuild()
	})
}

func (view *Window) addAudioVolume(content *widget.Container, y, width int, title string, music bool) {
	read := func() float64 {
		if music {
			return view.audioPreferences.MusicVolume
		}
		return view.audioPreferences.EffectsVolume
	}
	view.addLiveLabel(content, 0, y, width, 14, ink, func() string {
		return fmt.Sprintf("%s · %.0f%%", title, read()*100)
	})
	view.addSlider(content, 0, y+28, width, numericSetting{Session: true, Minimum: 0, Maximum: 100, Step: 1, Value: read() * 100}, func(value float64) error {
		view.changeAudio(func(preferences *experiment.AudioPreferences) {
			if music {
				preferences.MusicVolume = value / 100
			} else {
				preferences.EffectsVolume = value / 100
			}
		})
		return nil
	})
}
