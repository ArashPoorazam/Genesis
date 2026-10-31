package simulation

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

func inRange(value, minimum, maximum float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= minimum && value <= maximum
}

// Validate is shared by live edits and preset imports. Invalid changes never reach physics.
func (config Config) Validate() error {
	if config.Version != Version {
		return fmt.Errorf("unsupported preset version %d", config.Version)
	}
	if !inRange(config.WorldWidth, 100, 10000) || !inRange(config.WorldHeight, 100, 10000) {
		return errors.New("world dimensions must be between 100 and 10,000")
	}

	if err := config.validateParticleTypes(); err != nil {
		return err
	}
	if err := config.validateRelationships(); err != nil {
		return err
	}
	if err := config.Motion.validate(); err != nil {
		return err
	}
	if err := config.Generation.validate(); err != nil {
		return err
	}

	for _, particleType := range config.ParticleTypes {
		if particleType.ParticleRadius >= config.Generation.MinInnerRadius {
			return errors.New("minimum generated inner radius must exceed every particle radius")
		}
	}

	return nil
}

func (config Config) validateParticleTypes() error {
	if len(config.ParticleTypes) < 1 || len(config.ParticleTypes) > MaxTypes {
		return fmt.Errorf("choose between 1 and %d particle types", MaxTypes)
	}

	seenIDs := make(map[string]bool, len(config.ParticleTypes))
	totalPopulation := 0

	for _, particleType := range config.ParticleTypes {
		if strings.TrimSpace(particleType.ID) == "" || seenIDs[particleType.ID] {
			return errors.New("particle IDs must be unique and nonempty")
		}
		seenIDs[particleType.ID] = true

		if err := particleType.validate(); err != nil {
			return err
		}
		totalPopulation += particleType.Population
	}

	if totalPopulation > MaxParticles {
		return fmt.Errorf("total population cannot exceed %d", MaxParticles)
	}

	return nil
}

func (particleType ParticleType) validate() error {
	if strings.TrimSpace(particleType.Name) == "" || len([]rune(particleType.Name)) > 24 {
		return errors.New("type names need 1–24 characters")
	}
	if particleType.Color.A != 255 {
		return errors.New("particle colors must be opaque")
	}
	if particleType.Population < 0 || particleType.Population > MaxParticles {
		return fmt.Errorf("%s: invalid population", particleType.Name)
	}

	radiusValid := inRange(particleType.ParticleRadius, 0.5, 20)
	innerValid := inRange(particleType.InnerRadius, 0.5, 500)
	outerValid := inRange(particleType.OuterRadius, 0.5, 500)
	radiiOrdered := particleType.ParticleRadius < particleType.InnerRadius && particleType.InnerRadius < particleType.OuterRadius

	if !radiusValid || !innerValid || !outerValid || !radiiOrdered {
		return fmt.Errorf("%s: radius < inner < outer (maximum 500)", particleType.Name)
	}

	return nil
}

func (config Config) validateRelationships() error {
	if len(config.Relationships) != len(config.ParticleTypes) {
		return errors.New("relationship matrix does not match particle types")
	}

	for _, row := range config.Relationships {
		if len(row) != len(config.ParticleTypes) {
			return errors.New("relationship matrix must be square")
		}
		for _, strength := range row {
			if !inRange(strength, -10, 10) {
				return errors.New("relationships must be between -10 and +10")
			}
		}
	}

	return nil
}

func (motion Motion) validate() error {
	if !inRange(motion.VelocityRetention, 0, 1) || !inRange(motion.SpeedLimit, 0, 50) || !inRange(motion.RepulsionStrength, 0, 50) {
		return errors.New("invalid motion settings")
	}
	return nil
}

func (generation Generation) validate() error {
	if generation.PopulationCeiling < 1 || generation.PopulationCeiling > MaxParticles {
		return errors.New("population ceiling must be between 1 and 5,000")
	}
	if !inRange(generation.InitialSpeed, 0, 50) || !inRange(generation.SpawnSpacing, 0, 50) || !inRange(generation.RelationshipAmplitude, 0, 10) {
		return errors.New("invalid generation settings")
	}

	innerBoundsValid := inRange(generation.MinInnerRadius, 1, 499) && inRange(generation.MaxInnerRadius, generation.MinInnerRadius, 499)
	outerBoundsValid := inRange(generation.MinOuterRadius, 1, 500) && inRange(generation.MaxOuterRadius, generation.MinOuterRadius, 500)
	zonesSeparated := generation.MaxInnerRadius < generation.MinOuterRadius

	if !innerBoundsValid || !outerBoundsValid || !zonesSeparated {
		return errors.New("bounds require min inner ≤ max inner < min outer ≤ max outer")
	}
	return nil
}
