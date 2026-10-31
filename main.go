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
	screenWidth       = 1900
	screenHeight      = 1060
	prtPop            = 250 // for each type of particle
	prtInitialGap     = 0
	prtMaxSpawnSpeed  = 0
	SpeedLimit        = 5
	damping           = 0.90
	repulsionStrength = 4
)

// structs
type Game struct {
	Physics physics
	Prts    []Particle
}

// loop logic
func (g *Game) Update() error {
	for i := range g.Prts {
		g.applyInfluence(&g.Prts[i])
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
			A: 20,
		}

		vector.FillCircle(screen, x, y, float32(g.Physics.PT[p.TypeID].Mass+10), halo, true)
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
func (g *Game) applyInfluence(p *Particle) {
	vx := p.Speed * math.Cos(p.Direction)
	vy := p.Speed * math.Sin(p.Direction)

	for _, other := range g.Prts {
		inner := g.Physics.PT[other.TypeID].InnerInfluenceRadius
		outer := g.Physics.PT[other.TypeID].OuterInfluenceRadius

		dx := wrappedDelta(other.X-p.X, screenWidth)
		dy := wrappedDelta(other.Y-p.Y, screenHeight)
		distance := math.Hypot(dx, dy)

		if distance == 0 || distance > outer {
			continue
		}

		relation := g.Physics.Relations[p.TypeID][other.TypeID]

		force := forceCalculation(distance, inner, outer, relation)

		vx += (dx / distance) * force
		vy += (dy / distance) * force
	}

	vx *= damping
	vy *= damping

	p.Speed = min(math.Hypot(vx, vy), SpeedLimit)
	p.Direction = math.Atan2(vy, vx)
}

func forceCalculation(distance, inner, outer, relation float64) float64 {
	if distance >= outer {
		return 0
	}
	if distance < inner {
		return -repulsionStrength * (1 - distance/inner)
	}

	t := (distance - inner) / (outer - inner)
	weight := 1 - math.Abs(2*t-1)

	return relation * weight
}

func wrappedDelta(delta, size float64) float64 {
	return delta - size*math.Round(delta/size)
}

func wrapPosition(value, size float64) float64 {
	value = math.Mod(value, size)
	if value < 0 {
		value += size
	}
	return value
}

func borderCollisionAdjustment(p *Particle) {
	p.X = wrapPosition(p.X, screenWidth)
	p.Y = wrapPosition(p.Y, screenHeight)
}
