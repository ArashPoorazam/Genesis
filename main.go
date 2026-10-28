package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// variables
const (
	screenWidth    = 1900
	screenHeight   = 1060
	prtPop         = 100 // for each type of particle
	prtRadius      = 2
	prtInitialGap  = 0
	prtMaxMomentum = 5
)

// structs
type Game struct {
	Physics physics
	Prts    []Particle
}

// loop logic
func (g *Game) Update() error {
	for i := range g.Prts {
		g.Prts[i].move()
		borderCollisionAdjustment(&g.Prts[i])
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	for _, p := range g.Prts {
		vector.FillCircle(
			screen,
			float32(p.X),
			float32(p.Y),
			float32(g.Physics.PT[p.TypeID].Mass),
			g.Physics.PT[p.TypeID].Clr,
			true,
		)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// main gate
func main() {
	physics, prts := Genesis(prtPop, prtRadius)
	game := &Game{physics, prts}

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Genova")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

// helper functions
func borderCollisionAdjustment(prt *Particle) {
	if prt.X < 0 {
		prt.X = screenWidth
	}
	if prt.X > screenWidth {
		prt.X = 0
	}
	if prt.Y < 0 {
		prt.Y = screenHeight
	}
	if prt.Y > screenHeight {
		prt.Y = 0
	}
}
