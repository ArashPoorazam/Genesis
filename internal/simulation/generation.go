package simulation

import (
	"errors"
	"fmt"
	"image/color"
	"math/rand/v2"
	"strings"
)

const (
	relationshipStream = 1
	worldStream        = 2
	spawnStream        = 3
)

func Default(seed uint64) Config {
	config := Config{
		Version:     Version,
		WorldWidth:  1600,
		WorldHeight: 1000,
		Motion: Motion{
			VelocityRetention: 0.90,
			SpeedLimit:        5,
			RepulsionStrength: 10,
		},
		Generation: Generation{
			Seed:                  seed,
			PopulationCeiling:     300,
			RelationshipAmplitude: 2,
			MinInnerRadius:        10,
			MaxInnerRadius:        15,
			MinOuterRadius:        50,
			MaxOuterRadius:        120,
		},
	}

	names := []string{"Red", "Green", "Blue", "Pink", "Orange"}
	colors := []color.NRGBA{
		{R: 255, G: 70, B: 85, A: 255},
		{R: 109, G: 222, B: 145, A: 255},
		{R: 93, G: 153, B: 255, A: 255},
		{R: 244, G: 133, B: 190, A: 255},
		{R: 255, G: 201, B: 121, A: 255},
	}

	for index, name := range names {
		config.ParticleTypes = append(config.ParticleTypes, ParticleType{
			ID:             strings.ToLower(name),
			Name:           name,
			Color:          colors[index],
			ParticleRadius: 2,
			InnerRadius:    12,
			OuterRadius:    80,
		})
		config.Relationships = append(config.Relationships, make([]float64, len(names)))
	}

	config.RandomizeWorld(seed)
	return config
}

// Independent streams keep changes in one generator from changing the others.
func newRandomSource(seed, stream uint64) *rand.Rand {
	return rand.New(rand.NewPCG(seed, stream))
}

func (config *Config) RandomizeRelations(seed uint64) {
	random := newRandomSource(seed, relationshipStream)

	for row := range config.Relationships {
		for column := range config.Relationships[row] {
			config.Relationships[row][column] = (random.Float64()*2 - 1) * config.Generation.RelationshipAmplitude
		}
	}
}

func (config *Config) RandomizeWorld(seed uint64) {
	config.Generation.Seed = seed
	config.RandomizeRelations(seed)

	random := newRandomSource(seed, worldStream)
	settings := config.Generation
	remainingPopulation := MaxParticles

	for index := range config.ParticleTypes {
		particleType := &config.ParticleTypes[index]
		particleType.InnerRadius = settings.MinInnerRadius + random.Float64()*(settings.MaxInnerRadius-settings.MinInnerRadius)
		particleType.OuterRadius = settings.MinOuterRadius + random.Float64()*(settings.MaxOuterRadius-settings.MinOuterRadius)
		particleType.Population = min(random.IntN(settings.PopulationCeiling), remainingPopulation)
		remainingPopulation -= particleType.Population
	}
}

func (config *Config) AddType(id string) error {
	if len(config.ParticleTypes) >= MaxTypes {
		return fmt.Errorf("maximum %d types reached", MaxTypes)
	}

	config.ParticleTypes = append(config.ParticleTypes, ParticleType{
		ID:             id,
		Name:           fmt.Sprintf("Type %d", len(config.ParticleTypes)+1),
		Color:          color.NRGBA{R: 181, G: 161, B: 242, A: 255},
		ParticleRadius: 2,
		InnerRadius:    config.Generation.MinInnerRadius,
		OuterRadius:    config.Generation.MinOuterRadius,
	})

	for row := range config.Relationships {
		config.Relationships[row] = append(config.Relationships[row], 0)
	}
	config.Relationships = append(config.Relationships, make([]float64, len(config.ParticleTypes)))

	return nil
}

func (config *Config) RemoveType(index int) error {
	if len(config.ParticleTypes) == 1 {
		return errors.New("keep at least one particle type")
	}
	if index < 0 || index >= len(config.ParticleTypes) {
		return errors.New("unknown particle type")
	}

	config.ParticleTypes = append(config.ParticleTypes[:index], config.ParticleTypes[index+1:]...)
	config.Relationships = append(config.Relationships[:index], config.Relationships[index+1:]...)
	for row := range config.Relationships {
		config.Relationships[row] = append(config.Relationships[row][:index], config.Relationships[row][index+1:]...)
	}

	return nil
}
