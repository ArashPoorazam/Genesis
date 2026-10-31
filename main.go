package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

var screenWidth, screenHeight int

// variables
const (
	prtPop              = 500 // for each type of particle
	prtInitialGap       = 0
	prtMaxSpawnSpeed    = 0
	minOuterRadius      = 50.0
	maxOuterRadius      = 120.0
	minInnerRadius      = 15.0
	maxInnerRadius      = 10.0
	SpeedLimit          = 5
	damping             = 0.90
	repulsionStrength   = 8
	maxRelationStrength = 2
)

type Game struct {
	Physics      physics
	Prts         []Particle
	grid         *spatialGrid
	outerSquared []float64
	sprites      []particleSprite
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	monitor := ebiten.Monitor()
	if monitor == nil {
		log.Fatal("No monitor available")
	}

	screenWidth, screenHeight = monitor.Size()
	if screenWidth <= 0 || screenHeight <= 0 {
		log.Fatal("Monitor dimensions must be positive")
	}
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Genova")

	physics, prts := Genesis(prtPop)
	game := &Game{
		Physics:      physics,
		Prts:         prts,
		grid:         newSpatialGrid(physics.maxRadius()),
		outerSquared: make([]float64, len(physics.PT)),
		sprites:      make([]particleSprite, len(physics.PT)),
	}
	for typeID, particleType := range physics.PT {
		game.outerSquared[typeID] = particleType.OuterInfluenceRadius * particleType.OuterInfluenceRadius
		game.sprites[typeID] = makeParticleSprite(particleType)
	}
	game.grid.rebuild(game.Prts)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
