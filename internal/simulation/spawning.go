package simulation

import (
	"fmt"
	"math"
	"math/rand/v2"
)

const maxPlacementAttempts = 1000

func (world *Simulation) Restart() string {
	world.Particles = world.Particles[:0]
	counts := make([]int, len(world.Config.ParticleTypes))
	random := newRandomSource(world.Config.Generation.Seed, spawnStream)

	return world.populate(counts, random)
}

func (world *Simulation) populate(counts []int, random *rand.Rand) string {
	maximumRadius := 0.0
	for _, particleType := range world.Config.ParticleTypes {
		maximumRadius = max(maximumRadius, particleType.ParticleRadius)
	}

	// A minimum cell size avoids huge empty grids for tiny particles.
	cellRadius := max(16, 2*maximumRadius+world.Config.Generation.SpawnSpacing)
	spawnGrid := newGrid(world.Config.WorldWidth, world.Config.WorldHeight, cellRadius)
	spawnGrid.rebuild(world.Particles)

	for typeIndex, particleType := range world.Config.ParticleTypes {
		for counts[typeIndex] < particleType.Population {
			if !world.placeParticle(typeIndex, random, spawnGrid) {
				return fmt.Sprintf("Spawn space exhausted: %d particles placed. Reduce population or spacing.", len(world.Particles))
			}
			counts[typeIndex]++
		}
	}

	return ""
}

func (world *Simulation) placeParticle(typeIndex int, random *rand.Rand, grid *spatialGrid) bool {
	for range maxPlacementAttempts {
		particle := world.spawnCandidate(typeIndex, random)
		if world.spawnCollision(particle, grid) {
			continue
		}

		world.Particles = append(world.Particles, particle)
		grid.insert(particle)
		return true
	}

	return false
}

func (world *Simulation) spawnCandidate(typeIndex int, random *rand.Rand) Particle {
	direction := random.Float64() * 2 * math.Pi
	speed := random.Float64() * world.Config.Generation.InitialSpeed

	return Particle{
		TypeIndex: typeIndex,
		X:         random.Float64() * world.Config.WorldWidth,
		Y:         random.Float64() * world.Config.WorldHeight,
		VX:        math.Cos(direction) * speed,
		VY:        math.Sin(direction) * speed,
	}
}

func (world *Simulation) spawnCollision(particle Particle, grid *spatialGrid) bool {
	cell := grid.cell(particle.X, particle.Y)
	particleType := world.Config.ParticleTypes[particle.TypeIndex]

	for _, neighborCell := range grid.neighbors[cell][:grid.neighborCount] {
		for index := grid.heads[neighborCell]; index != -1; index = grid.next[index] {
			other := world.Particles[index]
			otherType := world.Config.ParticleTypes[other.TypeIndex]
			deltaX := wrappedDelta(particle.X-other.X, world.Config.WorldWidth)
			deltaY := wrappedDelta(particle.Y-other.Y, world.Config.WorldHeight)
			clearance := particleType.ParticleRadius + otherType.ParticleRadius + world.Config.Generation.SpawnSpacing

			if deltaX*deltaX+deltaY*deltaY < clearance*clearance {
				return true
			}
		}
	}

	return false
}
