package soundtrack

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ArashPoorazam/Genesis/internal/experiment"
	"github.com/hajimehoshi/ebiten/v2"
)

func TestAudioPlayback(t *testing.T) {
	if os.Getenv("GENESIS_AUDIO_TEST") != "1" {
		t.Skip("set GENESIS_AUDIO_TEST=1 with a working display and audio device")
	}
	t.Setenv("EGL_PLATFORM", "x11")
	preferences := experiment.DefaultAudioPreferences()
	preferences.MusicEnabled = true
	player := New(preferences)
	if err := player.Start(); err != nil {
		t.Fatal(err)
	}
	defer player.Close()
	driver := &playbackDriver{player: player, started: time.Now()}
	ebiten.SetWindowSize(320, 160)
	ebiten.SetWindowTitle("Genesis · Audio verification")
	ebiten.SetRunnableOnUnfocused(true)
	if err := ebiten.RunGame(driver); err != nil {
		t.Fatal(err)
	}
	if driver.frames < 180 {
		t.Fatal("audio verification ended early")
	}
}

type playbackDriver struct {
	player  *Player
	frames  int
	started time.Time
}

func (driver *playbackDriver) Layout(_, _ int) (int, int) { return 320, 160 }
func (driver *playbackDriver) Draw(_ *ebiten.Image)       {}
func (driver *playbackDriver) Update() error {
	player := driver.player
	player.Update(time.Now())
	if !player.context.IsReady() {
		if time.Since(driver.started) > 10*time.Second {
			return fmt.Errorf("audio device did not become ready")
		}
		return nil
	}
	driver.frames++
	switch driver.frames {
	case 40:
		if !player.music.IsPlaying() || player.music.Position() <= 0 {
			return fmt.Errorf("music did not start")
		}
		player.Queue(Save)
	case 45:
		if player.effect == nil || !player.effect.IsPlaying() {
			return fmt.Errorf("save cue did not start")
		}
		if player.gain >= player.preferences.MusicVolume {
			return fmt.Errorf("cue did not lower the music")
		}
		existing := player.effect
		player.playCue(Play, time.Now())
		if player.effect != existing {
			return fmt.Errorf("rapid cue overlapped existing sound")
		}
		prefs := player.preferences
		prefs.Muted = true
		player.SetPreferences(prefs)
		if player.music.IsPlaying() || player.effect.IsPlaying() {
			return fmt.Errorf("mute did not stop audio")
		}
	case 60:
		prefs := player.preferences
		prefs.Muted, prefs.EffectsEnabled = false, false
		player.SetPreferences(prefs)
		player.Queue(Play)
	case 80:
		if !player.music.IsPlaying() || player.effect != nil {
			return fmt.Errorf("independent effect mute failed")
		}
		prefs := player.preferences
		prefs.MusicEnabled = false
		player.SetPreferences(prefs)
	case 100:
		if player.music.IsPlaying() {
			return fmt.Errorf("music switch failed")
		}
		prefs := player.preferences
		prefs.MusicEnabled, prefs.EffectsEnabled = true, true
		player.SetPreferences(prefs)
		if err := player.music.SetPosition(pcmDuration(player.assets.music) - 100*time.Millisecond); err != nil {
			return err
		}
	case 180:
		if !player.music.IsPlaying() || player.music.Position() <= pcmDuration(player.assets.music) {
			return fmt.Errorf("score did not continue across loop boundary")
		}
		return ebiten.Termination
	}
	return nil
}
