package desktop

import "github.com/ArashPoorazam/Genesis/internal/simulation"

func minimumGeneratedRadius(config *simulation.Config) float64 {
	minimum := 1.0
	for _, particleType := range config.ParticleTypes {
		minimum = max(minimum, particleType.ParticleRadius+.1)
	}
	return minimum
}

func setGenerationRadius(config *simulation.Config, index int, value float64) {
	generation := &config.Generation
	values := []float64{generation.MinInnerRadius, generation.MaxInnerRadius, generation.MinOuterRadius, generation.MaxOuterRadius}
	minimum := minimumGeneratedRadius(config)
	if index >= 2 {
		minimum += .1
	}
	values[index] = max(minimum, value)
	for previous := index - 1; previous >= 0; previous-- {
		gap := 0.0
		if previous == 1 {
			gap = .1
		}
		values[previous] = min(values[previous], values[previous+1]-gap)
	}
	for next := index + 1; next < len(values); next++ {
		gap := 0.0
		if next == 2 {
			gap = .1
		}
		values[next] = max(values[next], values[next-1]+gap)
	}
	generation.MinInnerRadius, generation.MaxInnerRadius = values[0], values[1]
	generation.MinOuterRadius, generation.MaxOuterRadius = values[2], values[3]
}

func setParticleRadius(config *simulation.Config, typeIndex, zone int, value float64) {
	particleType := &config.ParticleTypes[typeIndex]
	switch zone {
	case 0:
		particleType.ParticleRadius = value
		particleType.InnerRadius = max(particleType.InnerRadius, value+.1)
		particleType.OuterRadius = max(particleType.OuterRadius, particleType.InnerRadius+.1)
		if config.Generation.MinInnerRadius <= value {
			setGenerationRadius(config, 0, value+.1)
		}
	case 1:
		particleType.InnerRadius = max(value, particleType.ParticleRadius+.1)
		particleType.OuterRadius = max(particleType.OuterRadius, particleType.InnerRadius+.1)
	case 2:
		particleType.OuterRadius = max(value, particleType.ParticleRadius+.2)
		particleType.InnerRadius = min(particleType.InnerRadius, particleType.OuterRadius-.1)
	}
}
