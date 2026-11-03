package simulation

import (
	"fmt"
	"math/rand/v2"
	"testing"
)

func TestGridBoundsMemoryAndFindsEveryNeighbor(t *testing.T) {
	for _, dimensions := range [][3]float64{
		{10000, 10000, 1}, // Both axes reach the cell limit.
		{10000, 100, 2},   // Only the wide axis reaches the limit.
		{100, 10000, 2},
		{100, 100, 500}, // One cell, radius larger than the world.
		{200, 300, 100}, // Two and three wrapped cells.
	} {
		width, height, radius := dimensions[0], dimensions[1], dimensions[2]
		t.Run(fmt.Sprintf("%gx%g/radius%g", width, height, radius), func(t *testing.T) {
			grid := newGrid(width, height, radius)
			if len(grid.heads) > maxGridAxisCells*maxGridAxisCells {
				t.Fatalf("grid allocated %d cells", len(grid.heads))
			}

			// Pairs straddle the world seam, a cell boundary, and an ordinary
			// interior point. Random particles also exercise unrelated candidates.
			particles := []Particle{
				{X: .1, Y: .1}, {X: width - .1, Y: height - .1},
				{X: grid.cellWidth - .1, Y: height / 2},
				{X: wrap(grid.cellWidth+.1, width), Y: height / 2},
				{X: width / 2, Y: height / 2},
				{X: width/2 + .1, Y: height/2 + .1},
			}
			random := rand.New(rand.NewPCG(42, 1))
			for range 100 {
				particles = append(particles, Particle{X: random.Float64() * width, Y: random.Float64() * height})
			}
			grid.rebuild(particles)

			for index, particle := range particles {
				candidates := make(map[int]bool)
				cell := grid.cell(particle.X, particle.Y)
				for _, neighbor := range grid.neighbors[cell][:grid.neighborCount] {
					for other := grid.heads[neighbor]; other != -1; other = grid.next[other] {
						if candidates[other] {
							t.Fatalf("particle %d visited neighbor %d twice", index, other)
						}
						candidates[other] = true
					}
				}

				// An all-pairs search is independent of the grid's cell partitioning.
				for other, candidate := range particles {
					deltaX := wrappedDelta(candidate.X-particle.X, width)
					deltaY := wrappedDelta(candidate.Y-particle.Y, height)
					if deltaX*deltaX+deltaY*deltaY < radius*radius && !candidates[other] {
						t.Fatalf("particle %d missed neighbor %d", index, other)
					}
				}
			}
		})
	}
}
