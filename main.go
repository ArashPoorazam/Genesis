package main

import (
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// variables
const (
	screenWidth      = 1900
	screenHeight     = 1060
	prtPop           = 50 // for each type of particle
	prtInitialGap    = 0
	prtMaxSpawnSpeed = 0
	SpeedLimit       = 3
	turnRate         = 1
)

// structs
type Game struct {
	Physics physics
	Prts    []Particle
}

// loop logic
func (g *Game) Update() error {
	for i := range g.Prts {
		g.prtInfluence(&g.Prts[i])
	}

	for i := range g.Prts {
		g.Prts[i].move()
		borderCollisionAdjustment(&g.Prts[i])
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	for _, p := range g.Prts {
		particleType := g.Physics.PT[p.TypeID]
		clr := particleType.Clr
		red, green, blue, _ := clr.RGBA()

		x, y := float32(p.X), float32(p.Y)
		radius := float32(particleType.Mass)

		halo := color.NRGBA{
			R: uint8(red >> 8),
			G: uint8(green >> 8),
			B: uint8(blue >> 8),
			A: 40,
		}

		vector.FillCircle(screen, x, y, float32(g.Physics.PT[p.TypeID].OuterInfluenceRadius), halo, true)
		vector.FillCircle(screen, x, y, radius, clr, true)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// main gate
func main() {
	physics, prts := Genesis(prtPop)
	game := &Game{physics, prts}

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Genova")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

// Update p's velocity based on nearby particles.
func (g *Game) prtInfluence(p *Particle) {
	vx := p.Speed * math.Cos(p.Direction)
	vy := p.Speed * math.Sin(p.Direction)

	radius := g.Physics.PT[p.TypeID].OuterInfluenceRadius

	for _, other := range g.Prts {
		dx := other.X - p.X
		dy := other.Y - p.Y
		distance := math.Hypot(dx, dy)

		if distance == 0 || distance > radius {
			continue
		}

		force := g.Physics.Relations[p.TypeID][other.TypeID]

		vx += (dx / distance) * force * turnRate
		vy += (dy / distance) * force * turnRate
	}

	p.Speed = min(math.Hypot(vx, vy), SpeedLimit)
	p.Direction = math.Atan2(vy, vx)
}

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
