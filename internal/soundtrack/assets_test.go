package soundtrack

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"
	"time"

	"github.com/ArashPoorazam/Genesis/internal/experiment"
	"github.com/hajimehoshi/ebiten/v2/audio"
)

func TestEmbeddedAudioAndLoop(t *testing.T) {
	data, err := loadAssets()
	if err != nil {
		t.Fatal(err)
	}
	duration := pcmDuration(data.music)
	if duration < 106*time.Second || duration > 107*time.Second {
		t.Fatalf("wrong score duration: %v", duration)
	}
	checkPCM := func(pcm []byte) {
		t.Helper()
		if len(pcm) == 0 || len(pcm)%4 != 0 {
			t.Fatal("invalid stereo PCM")
		}
		peak := 0
		for i := 0; i < len(pcm); i += 2 {
			value := int(int16(binary.LittleEndian.Uint16(pcm[i:])))
			if value < 0 {
				value = -value
			}
			peak = max(peak, value)
		}
		if peak < 100 || peak >= 32767 {
			t.Fatalf("silent or clipped audio: peak %d", peak)
		}
	}
	checkPCM(data.music)
	for _, name := range cueNames {
		pcm := data.cues[name]
		checkPCM(pcm)
		if pcmDuration(pcm) > time.Second {
			t.Fatal("water drop cue is too long")
		}
	}
	loop := audio.NewInfiniteLoop(bytes.NewReader(data.music), int64(len(data.music)))
	if _, err := loop.Seek(int64(len(data.music)-16), io.SeekStart); err != nil {
		t.Fatal(err)
	}
	boundary := make([]byte, 32)
	if _, err := io.ReadFull(loop, boundary); err != nil {
		t.Fatal(err)
	}
	want := append(append([]byte{}, data.music[len(data.music)-16:]...), data.music[:16]...)
	if !bytes.Equal(boundary, want) {
		t.Fatal("loop inserts or drops samples at the boundary")
	}
}

func TestCueCoalescing(t *testing.T) {
	player := New(experiment.DefaultAudioPreferences())
	player.Queue(Play)
	player.Queue(Save)
	if player.pending != Save {
		t.Fatal("last action did not replace the pending cue")
	}
	prefs := experiment.DefaultAudioPreferences()
	prefs.Muted = true
	player.SetPreferences(prefs)
	if player.pending != "" {
		t.Fatal("muting retained queued audio")
	}
}
