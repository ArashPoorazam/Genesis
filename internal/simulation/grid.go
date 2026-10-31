package simulation

import "math"

// spatialGrid stores linked particle indices in cells and precomputes wrapped neighbors.
// Cells are at least as wide as the maximum interaction radius.
type spatialGrid struct {
	columns, rows         int
	cellWidth, cellHeight float64
	heads, next           []int
	neighbors             [][9]int
	neighborCount         int
}

func newGrid(width, height, radius float64) *spatialGrid {
	columns := max(1, int(width/max(radius, 1)))
	rows := max(1, int(height/max(radius, 1)))
	grid := &spatialGrid{
		columns:       columns,
		rows:          rows,
		cellWidth:     width / float64(columns),
		cellHeight:    height / float64(rows),
		heads:         make([]int, columns*rows),
		neighbors:     make([][9]int, columns*rows),
		neighborCount: min(3, columns) * min(3, rows),
	}

	for row := range rows {
		for column := range columns {
			cell := row*columns + column
			neighborIndex := 0

			for rowOffset := range min(3, rows) {
				for columnOffset := range min(3, columns) {
					neighborRow := (row + rowOffset - 1 + rows) % rows
					neighborColumn := (column + columnOffset - 1 + columns) % columns
					grid.neighbors[cell][neighborIndex] = neighborRow*columns + neighborColumn
					neighborIndex++
				}
			}
		}
	}

	grid.rebuild(nil)
	return grid
}

func (grid *spatialGrid) cell(x, y float64) int {
	column := min(max(int(x/grid.cellWidth), 0), grid.columns-1)
	row := min(max(int(y/grid.cellHeight), 0), grid.rows-1)
	return row*grid.columns + column
}

func (grid *spatialGrid) insert(particle Particle) {
	cell := grid.cell(particle.X, particle.Y)
	index := len(grid.next)
	grid.next = append(grid.next, grid.heads[cell])
	grid.heads[cell] = index
}

func (grid *spatialGrid) rebuild(particles []Particle) {
	for cell := range grid.heads {
		grid.heads[cell] = -1
	}

	grid.next = grid.next[:0]
	for _, particle := range particles {
		grid.insert(particle)
	}
}

func wrappedDelta(delta, size float64) float64 {
	return delta - size*math.Round(delta/size)
}

func wrap(value, size float64) float64 {
	value = math.Mod(value, size)
	if value < 0 {
		value += size
	}
	return value
}
