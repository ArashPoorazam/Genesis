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

type Particle struct {
	TypeID    int
	X         float64
	Y         float64
	Speed     float64
	Direction float64
}

type ParticleType struct {
	Clr                  color.Color
	Mass                 float64 // mass also represent the radius of the particle
	InnerInfluenceRadius float64
	OuterInfluenceRadius float64
}

type physics struct {
	PT        []ParticleType
	Relations [][]float64
}

func createType(clr color.Color, mass float64) ParticleType {
	return ParticleType{
		clr,
		mass,
		mass + 3,
		(mass * 5) + mass,
	}
}

func prtSpawn(TypeID int) Particle {
	Speed := 0
	if prtMaxSpawnSpeed != 0 {
		Speed = rand.IntN(prtMaxSpawnSpeed + 1)
	}
	return Particle{
		TypeID,
		float64(rand.IntN(screenWidth)),
		float64(rand.IntN(screenHeight)),
		float64(Speed),
		float64(rand.IntN(360)) * math.Pi / 180,
	}
}

func (p *Particle) move() {
	p.X += math.Cos(p.Direction) * p.Speed
	p.Y += math.Sin(p.Direction) * p.Speed
}

func (p *Particle) acceleration(force float64) {
	// it can be a negetive force or positive
	p.Speed += force
}

// spawn initialization
func Genesis(population int) (physics, []Particle) {
	var physics physics

	redParticle := createType(color.RGBA{255, 0, 0, 255}, 2)
	greenParticle := createType(color.RGBA{0, 255, 0, 255}, 3)
	blueParticle := createType(color.RGBA{0, 0, 255, 255}, 4)

	physics.PT = append(physics.PT, redParticle, greenParticle, blueParticle)

	physics.Relations = [][]float64{
		// Red  Green  Blue
		{0, -1, -2}, // Red responds to...
		{2, 1, -1},  // Green responds to...
		{3, 0, -1},  // Blue responds to...
	}

	allPrts := []Particle{}

	for i := range physics.PT {
		for range population {
			for {
				prt := prtSpawn(i)
				if !spawnCollisionCheck(prt, allPrts, physics.PT[i].Mass) {
					allPrts = append(allPrts, prt)
					break
				}
			}
		}
	}

	return physics, allPrts
}

func spawnCollisionCheck(prt Particle, prts []Particle, radius float64) bool {
	for _, tp := range prts {
		if distance := EuclideanDistance(prt, tp); distance < radius {
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
