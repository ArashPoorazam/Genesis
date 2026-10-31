package main

import (
	"errors"
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand/v2"
)

type Particle struct {
	TypeID int
	X      float64
	Y      float64
	VX     float64
	VY     float64
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

func createType(name string, clr color.Color, mass, outer, inner float64) (ParticleType, error) {
	if mass >= inner || mass >= outer {
		return ParticleType{}, errors.New("mass can't be bigger than the inner or outer influence radius")
	}
	if inner >= outer {
		return ParticleType{}, errors.New("the inner zone cannot be bigger than the outer zone")
	}
	return ParticleType{
		name,
		clr,
		mass,
		inner,
		outer,
	}, nil
}

func prtSpawn(TypeID int) Particle {
	Speed := 0
	if prtMaxSpawnSpeed != 0 {
		Speed = rand.IntN(prtMaxSpawnSpeed + 1)
	}
	direction := float64(rand.IntN(360)) * math.Pi / 180
	return Particle{
		TypeID,
		float64(rand.IntN(screenWidth)),
		float64(rand.IntN(screenHeight)),
		float64(Speed) * math.Cos(direction),
		float64(Speed) * math.Sin(direction),
	}
}

func (p *Particle) move() {
	p.X += p.VX
	p.Y += p.VY
}

func (p physics) maxRadius() float64 {
	radius := float64(maxOuterRadius)
	for _, particleType := range p.PT {
		radius = max(radius, particleType.OuterInfluenceRadius)
	}
	return radius
}

// spawn initialization
func Genesis(maxPopulation int) (physics, []Particle) {
	if minOuterRadius <= 10 || maxOuterRadius < minOuterRadius {
		log.Fatal("outer radius bounds must satisfy 10 < minOuterRadius <= maxOuterRadius to accommodate the inner radius")
	}
	var physics physics
	var RN []float64 // Random Relation numbers
	var outers []float64
	var inners []float64

	for range 5 {
		var o float64
		var i float64
		for {
			o = minOuterRadius + rand.Float64()*(maxOuterRadius-minOuterRadius)
			i = minInnerRadius + rand.Float64()*(maxInnerRadius-minInnerRadius)
			if o > i {
				break
			}
		}
		outers = append(outers, o)
		inners = append(inners, i)
	}

	redParticle, err := createType(
		"Red",
		color.RGBA{255, 0, 0, 255},
		2,
		outers[0],
		inners[0],
	)
	if err != nil {
		log.Fatal(err)
	}
	greenParticle, err := createType(
		"Green",
		color.RGBA{0, 255, 0, 255},
		2,
		outers[1],
		inners[1],
	)
	if err != nil {
		log.Fatal(err)
	}

	blueParticle, err := createType(
		"Blue",
		color.RGBA{0, 0, 255, 255},
		2,
		outers[2],
		inners[2],
	)
	if err != nil {
		log.Fatal(err)
	}

	pinkParticle, err := createType(
		"Pink",
		color.RGBA{255, 105, 180, 255},
		2,
		outers[3],
		inners[3],
	)
	if err != nil {
		log.Fatal(err)
	}

	orangeParticle, err := createType(
		"Orange",
		color.RGBA{255, 213, 128, 255},
		2,
		outers[4],
		inners[4],
	)
	if err != nil {
		log.Fatal(err)
	}

	physics.PT = append(physics.PT, redParticle, greenParticle, blueParticle, pinkParticle, orangeParticle)

	for range len(physics.PT) * len(physics.PT) {
		n := rand.Float64()*(maxRelationStrength*2) - maxRelationStrength
		RN = append(RN, n)
	}

	physics.Relations = [][]float64{
		//        Red    Green  Blue   pink   orange
		{RN[0], RN[1], RN[2], RN[3], RN[4]},      // Red responds to...
		{RN[5], RN[6], RN[7], RN[8], RN[9]},      // Green responds to...
		{RN[10], RN[11], RN[12], RN[13], RN[14]}, // Blue responds to...
		{RN[15], RN[16], RN[17], RN[18], RN[19]}, // pink responds to...
		{RN[20], RN[21], RN[22], RN[23], RN[24]}, // orange responds to...
	}

	printMatrix(physics.Relations)

	allPrts := []Particle{}
	spawnGrid := newSpatialGrid(physics.maxRadius())

	for i := range physics.PT {
		pop := rand.IntN(maxPopulation)
		fmt.Printf(
			"%d Created %d of %s with (o: %.2f, i: %0.2f)\n",
			i, pop, physics.PT[i].Name, physics.PT[i].OuterInfluenceRadius, physics.PT[i].InnerInfluenceRadius,
		)
		for range pop {
			for attempt := 0; ; attempt++ {
				if attempt == 1000 {
					log.Printf("Stopped spawning after 1000 unsuccessful attempts; created %d particles", len(allPrts))
					return physics, allPrts
				}
				prt := prtSpawn(i)
				if !spawnCollisionCheck(prt, allPrts, physics.PT[i].Mass, spawnGrid) {
					allPrts = append(allPrts, prt)
					spawnGrid.insert(prt)
					break
				}
			}
		}
	}

	return physics, allPrts
}

func spawnCollisionCheck(p Particle, prts []Particle, radius float64, grid *spatialGrid) bool {
	cell := grid.cell(p.X, p.Y)
	for _, neighbor := range grid.neighbors[cell][:grid.neighborCount] {
		for otherIndex := grid.heads[neighbor]; otherIndex != -1; otherIndex = grid.next[otherIndex] {
			other := prts[otherIndex]
			dx, dy := p.X-other.X, p.Y-other.Y
			if dx*dx+dy*dy < radius*radius {
				return true
			}
		}
	}
	return false
}

func EuclideanDistance(A, B Particle) float64 {
	dx := A.X - B.X
	dy := A.Y - B.Y

	return math.Sqrt(float64(dx*dx + dy*dy))
}

func printMatrix(matrix [][]float64) {
	fmt.Println("{")
	for _, row := range matrix {
		fmt.Print("   { ")
		for column, value := range row {
			if column > 0 {
				fmt.Print(", ")
			}
			fmt.Printf("%.2f", value)
		}
		fmt.Println(" }")
	}
	fmt.Println("}")
}
