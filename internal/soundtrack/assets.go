// Package soundtrack plays the local handpan score and water-drop action cues.
package soundtrack

import (
	"bytes"
	"io"
	"time"

	audioassets "github.com/ArashPoorazam/Genesis/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

const sampleRate = 44100

type Cue string

const (
	Play       Cue = "play"
	Pause      Cue = "pause"
	Save       Cue = "save"
	NewWorld   Cue = "new-world"
	Shuffle    Cue = "shuffle"
	Population Cue = "population"
)

var cueNames = []Cue{Play, Pause, Save, NewWorld, Shuffle, Population}

type assets struct {
	music []byte
	cues  map[Cue][]byte
}

func loadAssets() (assets, error) {
	var result assets
	encoded, err := audioassets.Files.ReadFile("handpan/music/quiet-orbits-loop.ogg")
	if err != nil {
		return result, err
	}
	stream, err := vorbis.DecodeWithSampleRate(sampleRate, bytes.NewReader(encoded))
	if err != nil {
		return result, err
	}
	result.music, err = io.ReadAll(stream)
	if err != nil {
		return result, err
	}
	encoded, err = audioassets.Files.ReadFile("sfx/water-drop.wav")
	if err != nil {
		return result, err
	}
	drop, err := wav.DecodeWithSampleRate(sampleRate, bytes.NewReader(encoded))
	if err != nil {
		return result, err
	}
	pcm, err := io.ReadAll(drop)
	if err != nil {
		return result, err
	}
	result.cues = make(map[Cue][]byte)
	for _, name := range cueNames {
		result.cues[name] = pcm
	}
	return result, nil
}

func pcmDuration(pcm []byte) time.Duration {
	return time.Duration(len(pcm)) * time.Second / (sampleRate * 4)
}
