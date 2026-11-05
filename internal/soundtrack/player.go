package soundtrack

import (
	"bytes"
	"time"

	"github.com/ArashPoorazam/Genesis/internal/experiment"
	"github.com/hajimehoshi/ebiten/v2/audio"
)

// Player is controlled on the desktop update goroutine; audio reads immutable PCM.
type Player struct {
	context     *audio.Context
	music       *audio.Player
	effect      *audio.Player
	assets      assets
	preferences experiment.AudioPreferences
	lastCue     time.Time
	duckUntil   time.Time
	lastUpdate  time.Time
	gain        float64
	pending     Cue
}

func New(preferences experiment.AudioPreferences) *Player {
	return &Player{preferences: preferences}
}

// Start is explicit so tests and screenshot captures can avoid opening a device.
func (player *Player) Start() error {
	if player.music != nil {
		return nil
	}
	data, err := loadAssets()
	if err != nil {
		return err
	}
	context := audio.CurrentContext()
	if context == nil {
		context = audio.NewContext(sampleRate)
	}
	music, err := context.NewPlayer(audio.NewInfiniteLoop(bytes.NewReader(data.music), int64(len(data.music))))
	if err != nil {
		return err
	}
	player.context, player.music, player.assets = context, music, data
	music.SetVolume(0)
	return nil
}

func (player *Player) SetPreferences(preferences experiment.AudioPreferences) {
	player.preferences = preferences
	if preferences.Muted || !preferences.EffectsEnabled || preferences.EffectsVolume == 0 {
		player.pending = ""
		if player.effect != nil {
			player.effect.Pause()
		}
	}
	if player.effect != nil {
		player.effect.SetVolume(preferences.EffectsVolume)
	}
	if preferences.Muted && player.music != nil {
		player.music.Pause()
		player.music.SetVolume(0)
		player.gain = 0
	}
}

// Queue coalesces important actions in one frame.
func (player *Player) Queue(cue Cue) {
	if player == nil {
		return
	}
	player.pending = cue
}

func (player *Player) Update(now time.Time) {
	if player == nil || player.music == nil {
		return
	}
	if !player.context.IsReady() {
		player.pending = ""
		return
	}
	if player.effect != nil && !player.effect.IsPlaying() {
		player.effect.Close()
		player.effect = nil
	}
	if player.pending != "" {
		player.playCue(player.pending, now)
		player.pending = ""
	}
	target := player.preferences.MusicVolume
	if player.preferences.Muted || !player.preferences.MusicEnabled {
		target = 0
	}
	if now.Before(player.duckUntil) {
		target *= .72
	}
	delta := now.Sub(player.lastUpdate).Seconds()
	if player.lastUpdate.IsZero() {
		delta = 1.0 / 60
	}
	player.lastUpdate = now
	// Smooth volume, ducking, and startup without tying fades to simulation speed.
	rate := 2.0
	if target < player.gain {
		rate = 6
	}
	step := min(.1, delta) * rate
	if target > player.gain {
		player.gain = min(target, player.gain+step)
	} else {
		player.gain = max(target, player.gain-step)
	}
	player.music.SetVolume(player.gain)
	if player.gain > 0 {
		if !player.music.IsPlaying() {
			player.music.Play()
		}
	} else {
		player.music.Pause()
	}
}

func (player *Player) playCue(cue Cue, now time.Time) {
	if player.preferences.Muted || !player.preferences.EffectsEnabled || player.preferences.EffectsVolume == 0 {
		return
	}
	if now.Sub(player.lastCue) < 140*time.Millisecond {
		return
	}
	// One audible cue at a time: no double strikes from rapid actions or dialogs.
	if player.effect != nil && player.effect.IsPlaying() {
		return
	}
	pcm := player.assets.cues[cue]
	if len(pcm) == 0 {
		return
	}
	player.effect = player.context.NewPlayerFromBytes(pcm)
	player.effect.SetVolume(player.preferences.EffectsVolume)
	player.effect.Play()
	player.lastCue = now
	player.duckUntil = now.Add(pcmDuration(pcm))
}

func (player *Player) Close() {
	if player == nil {
		return
	}
	if player.effect != nil {
		player.effect.Close()
		player.effect = nil
	}
	if player.music != nil {
		player.music.Close()
		player.music = nil
	}
}
