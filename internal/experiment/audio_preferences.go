package experiment

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
)

// AudioPreferences are local listening preferences, never part of an experiment.
type AudioPreferences struct {
	Muted          bool    `json:"muted"`
	MusicEnabled   bool    `json:"music_enabled"`
	EffectsEnabled bool    `json:"effects_enabled"`
	MusicVolume    float64 `json:"music_volume"`
	EffectsVolume  float64 `json:"effects_volume"`
}

func DefaultAudioPreferences() AudioPreferences {
	return AudioPreferences{MusicEnabled: false, EffectsEnabled: true, MusicVolume: .45, EffectsVolume: .55}
}

func (preferences AudioPreferences) Validate() error {
	for _, volume := range []float64{preferences.MusicVolume, preferences.EffectsVolume} {
		if math.IsNaN(volume) || math.IsInf(volume, 0) || volume < 0 || volume > 1 {
			return fmt.Errorf("audio volume must be between 0 and 1")
		}
	}
	return nil
}

func ReadAudioPreferences(path string) (AudioPreferences, error) {
	defaults := DefaultAudioPreferences()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return defaults, nil
	}
	if err != nil {
		return defaults, err
	}
	preferences := defaults
	if err := json.Unmarshal(data, &preferences); err != nil {
		return defaults, err
	}
	if err := preferences.Validate(); err != nil {
		return defaults, err
	}
	return preferences, nil
}

func WriteAudioPreferences(path string, preferences AudioPreferences) error {
	if err := preferences.Validate(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(preferences, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".audio-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	defer file.Close()
	if _, err := file.Write(append(data, '\n')); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}
