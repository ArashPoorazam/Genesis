package desktop

import "github.com/ArashPoorazam/Genesis/internal/simulation"

// These selectors keep callbacks explicit about which dependent value they edit.
type particleRadiusZone int

const (
	physicalRadius particleRadiusZone = iota
	innerInfluenceRadius
	outerInfluenceRadius
)

type generationRadiusBound int

const (
	minimumInnerRadius generationRadiusBound = iota
	maximumInnerRadius
	minimumOuterRadius
	maximumOuterRadius
)

// One slider increment separates radii that must be strictly ordered.
const radiusGap = 0.1

func minimumGeneratedRadius(config *simulation.Config) float64 {
	minimum := 1.0
	for _, particleType := range config.ParticleTypes {
		minimum = max(minimum, particleType.ParticleRadius+radiusGap)
	}
	return minimum
}

// setGenerationRadius moves neighboring bounds outward from the edited bound.
// Inner and outer intervals may each collapse to one value, but cannot overlap.
func setGenerationRadius(config *simulation.Config, bound generationRadiusBound, value float64) {
	generation := &config.Generation
	values := []float64{generation.MinInnerRadius, generation.MaxInnerRadius, generation.MinOuterRadius, generation.MaxOuterRadius}
	minimum := minimumGeneratedRadius(config)
	if bound >= minimumOuterRadius {
		minimum += radiusGap
	}
	values[bound] = max(minimum, value)
	for previous := int(bound) - 1; previous >= 0; previous-- {
		gap := 0.0
		if previous == int(maximumInnerRadius) {
			gap = radiusGap
		}
		values[previous] = min(values[previous], values[previous+1]-gap)
	}
	for next := int(bound) + 1; next < len(values); next++ {
		gap := 0.0
		if next == int(minimumOuterRadius) {
			gap = radiusGap
		}
		values[next] = max(values[next], values[next-1]+gap)
	}
	generation.MinInnerRadius, generation.MaxInnerRadius = values[0], values[1]
	generation.MinOuterRadius, generation.MaxOuterRadius = values[2], values[3]
}

// setParticleRadius preserves physical < inner < outer within one undoable edit.
func setParticleRadius(config *simulation.Config, typeIndex int, zone particleRadiusZone, value float64) {
	particleType := &config.ParticleTypes[typeIndex]
	switch zone {
	case physicalRadius:
		particleType.ParticleRadius = value
		particleType.InnerRadius = max(particleType.InnerRadius, value+radiusGap)
		particleType.OuterRadius = max(particleType.OuterRadius, particleType.InnerRadius+radiusGap)
		if config.Generation.MinInnerRadius <= value {
			setGenerationRadius(config, minimumInnerRadius, value+radiusGap)
		}
	case innerInfluenceRadius:
		particleType.InnerRadius = max(value, particleType.ParticleRadius+radiusGap)
		particleType.OuterRadius = max(particleType.OuterRadius, particleType.InnerRadius+radiusGap)
	case outerInfluenceRadius:
		particleType.OuterRadius = max(value, particleType.ParticleRadius+2*radiusGap)
		particleType.InnerRadius = min(particleType.InnerRadius, particleType.OuterRadius-radiusGap)
	}
}
