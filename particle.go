package main

import (
	"image/color"
	"math"
	"math/rand/v2"
)

const (
	Red int = iota
	Green
	Blue
)

type ParticleType struct {
	Clr                  color.Color
	Mass                 float64 // for simplicity the mass also represent the radius and the influence energy of a particle type
	InnerInfluenceRadius float64
	OuterInfluenceRadius float64
}

type Particle struct {
	TypeID    int
	X         float64
	Y         float64
	Velocity  float64
	Direction float64
}

type physics struct {
	PT        []ParticleType
	Relations [][]float64
}

func prtSpawn(TypeID int) Particle {
	return Particle{
		TypeID,
		float64(rand.IntN(screenWidth)),
		float64(rand.IntN(screenHeight)),
		float64(rand.IntN(prtMaxMomentum)),
		float64(rand.IntN(360)) * math.Pi / 180,
	}
}

func (p *Particle) move() {
	p.X += math.Cos(p.Direction) * p.Velocity
	p.Y += math.Sin(p.Direction) * p.Velocity
}

func (p *Particle) boost(force float64) {
	// it can be a negetive force or positive
	p.Velocity += force
}

func (p *Particle) influence() {
}

// spawn initialization
func Genesis(population int, radius float64) (physics, []Particle) {
	var physics physics

	redParticle := ParticleType{
		color.RGBA{255, 0, 0, 255},
		2,
		5,
		radius * 3,
	}
	greenParticle := ParticleType{
		color.RGBA{0, 255, 0, 255},
		3,
		6,
		radius * 6,
	}
	blueParticle := ParticleType{
		color.RGBA{0, 0, 255, 255},
		4,
		7,
		radius * 9,
	}
	physics.PT = append(physics.PT, redParticle, greenParticle, blueParticle)

	physics.Relations = [][]float64{
		// Red   Green  Blue
		{0.0, 0.8, -0.5}, // Red responds to...
		{-0.3, 0.0, 0.6}, // Green responds to...
		{0.2, -0.7, 0.0}, // Blue responds to...
	}

	allPrts := []Particle{}

	for i := range physics.PT {
		for range population {
			for {
				prt := prtSpawn(i)
				if !prtCollisionCheck(prt, allPrts, radius) {
					allPrts = append(allPrts, prt)
					break
				}
			}
		}
	}

	return physics, allPrts
}

func prtCollisionCheck(prt Particle, prts []Particle, radius float64) bool {
	for _, p := range prts {
		if distance := EuclideanDistance(prt, p); distance < radius {
			return true
		}
	}
	return false
}

func EuclideanDistance(A, B Particle) float64 {
	dx := A.X - B.X
	dy := A.Y - B.Y

	return math.Sqrt(float64(dx*dx + dy*dy))
}
