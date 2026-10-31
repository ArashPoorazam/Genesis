// Package simulation contains the deterministic particle model and physics.
// It has no dependency on the desktop interface or filesystem.
package simulation

import "image/color"

const (
	MaxTypes     = 12
	MaxParticles = 5000
	Version      = 1
)

// ParticleType.ID remains stable when the editor adds or removes other types.
// Particle.TypeIndex is the dense runtime index into Config.ParticleTypes.
type ParticleType struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	Color          color.NRGBA `json:"color"`
	Population     int         `json:"population"`
	ParticleRadius float64     `json:"radius"`
	InnerRadius    float64     `json:"inner"`
	OuterRadius    float64     `json:"outer"`
}

type Motion struct {
	VelocityRetention float64 `json:"retention"`
	SpeedLimit        float64 `json:"speed_limit"`
	RepulsionStrength float64 `json:"repulsion"`
}

type Generation struct {
	Seed                  uint64  `json:"seed"`
	PopulationCeiling     int     `json:"max_population"`
	InitialSpeed          float64 `json:"initial_speed"`
	SpawnSpacing          float64 `json:"spacing"`
	RelationshipAmplitude float64 `json:"amplitude"`
	MinInnerRadius        float64 `json:"min_inner"`
	MaxInnerRadius        float64 `json:"max_inner"`
	MinOuterRadius        float64 `json:"min_outer"`
	MaxOuterRadius        float64 `json:"max_outer"`
}

// Relationships[row][column] describes how the row type responds to the column type.
// The two directions are independent, and influence radii belong to the column type.
type Config struct {
	Version       int            `json:"version"`
	WorldWidth    float64        `json:"world_width"`
	WorldHeight   float64        `json:"world_height"`
	ParticleTypes []ParticleType `json:"types"`
	Relationships [][]float64    `json:"relations"`
	Motion        Motion         `json:"motion"`
	Generation    Generation     `json:"generation"`
}

func (config Config) Clone() Config {
	config.ParticleTypes = append([]ParticleType(nil), config.ParticleTypes...)

	relationships := make([][]float64, len(config.Relationships))
	for row := range relationships {
		relationships[row] = append([]float64(nil), config.Relationships[row]...)
	}
	config.Relationships = relationships

	return config
}
