package simulation

type Particle struct {
	TypeIndex int
	X, Y      float64
	VX, VY    float64
}

type Simulation struct {
	Config    Config
	Particles []Particle

	grid         *spatialGrid
	outerSquared []float64
}

func New(config Config) (*Simulation, string, error) {
	if err := config.Validate(); err != nil {
		return nil, "", err
	}

	world := &Simulation{Config: config.Clone()}
	world.rebuildGeometry()
	message := world.Restart()

	return world, message, nil
}

func (world *Simulation) rebuildGeometry() {
	maximumRadius := 1.0
	world.outerSquared = make([]float64, len(world.Config.ParticleTypes))

	for index, particleType := range world.Config.ParticleTypes {
		maximumRadius = max(maximumRadius, particleType.OuterRadius)
		world.outerSquared[index] = particleType.OuterRadius * particleType.OuterRadius
	}

	world.grid = newGrid(world.Config.WorldWidth, world.Config.WorldHeight, maximumRadius)
}

// Apply preserves surviving particles, using stable IDs to remap their dense indices.
// It validates before mutating any simulation state.
func (world *Simulation) Apply(config Config) (string, error) {
	if err := config.Validate(); err != nil {
		return "", err
	}

	previous := world.Config
	counts := world.retainParticles(config)
	world.Config = config.Clone()

	if geometryChanged(previous, config) {
		world.rebuildGeometry()
	}
	if populationsChanged(previous, config) {
		return world.populate(counts, newRandomSource(config.Generation.Seed, spawnStream)), nil
	}

	return "", nil
}

func (world *Simulation) retainParticles(config Config) []int {
	indicesByID := make(map[string]int, len(config.ParticleTypes))
	for index, particleType := range config.ParticleTypes {
		indicesByID[particleType.ID] = index
	}

	counts := make([]int, len(config.ParticleTypes))
	retained := world.Particles[:0]

	for _, particle := range world.Particles {
		previousType := world.Config.ParticleTypes[particle.TypeIndex]
		newIndex, exists := indicesByID[previousType.ID]
		if !exists || counts[newIndex] >= config.ParticleTypes[newIndex].Population {
			continue
		}

		particle.TypeIndex = newIndex
		particle.X = wrap(particle.X, config.WorldWidth)
		particle.Y = wrap(particle.Y, config.WorldHeight)
		retained = append(retained, particle)
		counts[newIndex]++
	}

	world.Particles = retained
	return counts
}

func geometryChanged(previous, next Config) bool {
	if previous.WorldWidth != next.WorldWidth || previous.WorldHeight != next.WorldHeight || len(previous.ParticleTypes) != len(next.ParticleTypes) {
		return true
	}

	for index, particleType := range next.ParticleTypes {
		oldType := previous.ParticleTypes[index]
		if particleType.ID != oldType.ID || particleType.OuterRadius != oldType.OuterRadius {
			return true
		}
	}

	return false
}

func populationsChanged(previous, next Config) bool {
	if len(previous.ParticleTypes) != len(next.ParticleTypes) {
		return true
	}

	for index, particleType := range next.ParticleTypes {
		oldType := previous.ParticleTypes[index]
		if particleType.ID != oldType.ID || particleType.Population != oldType.Population {
			return true
		}
	}

	return false
}
