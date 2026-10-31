package simulation

import "math"

func Force(distance, innerRadius, outerRadius, relationship, repulsion float64) float64 {
	if distance >= outerRadius {
		return 0
	}
	if distance < innerRadius {
		return -repulsion * (1 - distance/innerRadius)
	}

	progress := (distance - innerRadius) / (outerRadius - innerRadius)
	weight := 1 - math.Abs(2*progress-1)
	return relationship * weight
}

// Step advances one fixed physics interval. Playback speed changes how often it runs.
func (world *Simulation) Step() {
	world.grid.rebuild(world.Particles)

	// Compute all velocities before moving any particles.
	for index := range world.Particles {
		world.applyInfluence(&world.Particles[index])
	}

	for index := range world.Particles {
		particle := &world.Particles[index]
		particle.X = wrap(particle.X+particle.VX, world.Config.WorldWidth)
		particle.Y = wrap(particle.Y+particle.VY, world.Config.WorldHeight)
	}
}

func (world *Simulation) applyInfluence(particle *Particle) {
	velocityX, velocityY := particle.VX, particle.VY
	cell := world.grid.cell(particle.X, particle.Y)

	for _, neighborCell := range world.grid.neighbors[cell][:world.grid.neighborCount] {
		for otherIndex := world.grid.heads[neighborCell]; otherIndex != -1; otherIndex = world.grid.next[otherIndex] {
			other := &world.Particles[otherIndex]
			deltaX := wrappedDelta(other.X-particle.X, world.Config.WorldWidth)
			deltaY := wrappedDelta(other.Y-particle.Y, world.Config.WorldHeight)
			distanceSquared := deltaX*deltaX + deltaY*deltaY

			if distanceSquared == 0 || distanceSquared >= world.outerSquared[other.TypeIndex] {
				continue
			}

			influencingType := world.Config.ParticleTypes[other.TypeIndex]
			relationship := world.Config.Relationships[particle.TypeIndex][other.TypeIndex]
			distance := math.Sqrt(distanceSquared)
			force := Force(distance, influencingType.InnerRadius, influencingType.OuterRadius, relationship, world.Config.Motion.RepulsionStrength)

			velocityX += deltaX / distance * force
			velocityY += deltaY / distance * force
		}
	}

	velocityX *= world.Config.Motion.VelocityRetention
	velocityY *= world.Config.Motion.VelocityRetention

	speedSquared := velocityX*velocityX + velocityY*velocityY
	speedLimit := world.Config.Motion.SpeedLimit
	if speedSquared > speedLimit*speedLimit {
		scale := speedLimit / math.Sqrt(speedSquared)
		velocityX *= scale
		velocityY *= scale
	}

	particle.VX, particle.VY = velocityX, velocityY
}
