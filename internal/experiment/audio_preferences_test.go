package experiment

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestAudioPreferences(t *testing.T) {
	path := filepath.Join(t.TempDir(), "genesis", "audio.json")
	defaults, err := ReadAudioPreferences(path)
	if err != nil || defaults != DefaultAudioPreferences() {
		t.Fatalf("missing preferences: %v, %v", defaults, err)
	}
	preferences := AudioPreferences{Muted: true, MusicEnabled: true, EffectsEnabled: false, MusicVolume: .23, EffectsVolume: .78}
	if err := WriteAudioPreferences(path, preferences); err != nil {
		t.Fatal(err)
	}
	loaded, err := ReadAudioPreferences(path)
	if err != nil || loaded != preferences {
		t.Fatalf("round trip: %v, %v", loaded, err)
	}
	for _, invalid := range []float64{-.1, 1.1, math.NaN(), math.Inf(1)} {
		bad := preferences
		bad.MusicVolume = invalid
		if err := WriteAudioPreferences(path, bad); err == nil {
			t.Fatal("invalid volume accepted")
		}
		loaded, err = ReadAudioPreferences(path)
		if err != nil || loaded != preferences {
			t.Fatal("invalid write changed saved preferences")
		}
	}
	if err := os.WriteFile(path, []byte(`{"music_volume":2}`), 0600); err != nil {
		t.Fatal(err)
	}
	loaded, err = ReadAudioPreferences(path)
	if err == nil || loaded != DefaultAudioPreferences() {
		t.Fatal("invalid file did not fall back to safe defaults")
	}
}
