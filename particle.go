package main

import (
	"fmt"
	"image/color"
	"math"
	"math/rand/v2"
)

type Particle struct {
	TypeID    int
	X         float64
	Y         float64
	Speed     float64
	Direction float64
}

type ParticleType struct {
	Name                 string
	Clr                  color.Color
	Mass                 float64 // mass also represent the radius of the particle
	InnerInfluenceRadius float64
	OuterInfluenceRadius float64
}

type physics struct {
	PT        []ParticleType
	Relations [][]float64
}

func createType(name string, clr color.Color, mass, outer, inner float64) ParticleType {
	return ParticleType{
		name,
		clr,
		mass,
		inner,
		outer,
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

// spawn initialization
func Genesis(maxPopulation int) (physics, []Particle) {
	var physics physics

	redParticle := createType(
		"Red",
		color.RGBA{255, 0, 0, 255},
		2,
		80,
		10,
	)
	greenParticle := createType(
		"Green",
		color.RGBA{0, 255, 0, 255},
		3,
		60,
		20,
	)
	blueParticle := createType(
		"Blue",
		color.RGBA{0, 0, 255, 255},
		4,
		100,
		10,
	)
	pinkParticle := createType(
		"Pink",
		color.RGBA{255, 105, 180, 255},
		2,
		30,
		10,
	)
	orangeParticle := createType(
		"Orange",
		color.RGBA{255, 213, 128, 255},
		3,
		120,
		40,
	)

	physics.PT = append(physics.PT, redParticle, greenParticle, blueParticle, pinkParticle, orangeParticle)

	var RN []float64
	for range len(physics.PT) * len(physics.PT) {
		n := rand.Float64()*2 - 1
		RN = append(RN, n)
	}
	fmt.Println(RN)

	physics.Relations = [][]float64{
		//        Red    Green  Blue   pink   orange
		{RN[0], RN[1], RN[2], RN[3], RN[4]},      // Red follows green; avoids blue.
		{RN[5], RN[6], RN[7], RN[8], RN[9]},      // Green follows blue; avoids red.
		{RN[10], RN[11], RN[12], RN[13], RN[14]}, // Blue follows red; avoids gree
		{RN[15], RN[16], RN[17], RN[18], RN[19]}, // pink follows red; avoids green.
		{RN[20], RN[21], RN[22], RN[23], RN[24]}, // orange follows red; avoids green.
	}
	allPrts := []Particle{}

	for i := range physics.PT {
		pop := rand.IntN(maxPopulation)
		fmt.Printf("Create %d of %s\n", pop, physics.PT[i].Name)
		for range pop {
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

func spawnCollisionCheck(p Particle, prts []Particle, radius float64) bool {
	for _, tp := range prts {
		if distance := EuclideanDistance(p, tp); distance < radius {
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
