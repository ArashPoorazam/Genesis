package main

type spatialGrid struct {
	columns       int
	cellWidth     float64
	cellHeight    float64
	heads         []int
	next          []int
	neighbors     [][9]int
	neighborCount int
}

func newSpatialGrid(radius float64) *spatialGrid {
	columns := max(1, int(float64(screenWidth)/max(radius, 1)))
	rows := max(1, int(float64(screenHeight)/max(radius, 1)))
	grid := &spatialGrid{
		columns:       columns,
		cellWidth:     float64(screenWidth) / float64(columns),
		cellHeight:    float64(screenHeight) / float64(rows),
		heads:         make([]int, columns*rows),
		neighbors:     make([][9]int, columns*rows),
		neighborCount: min(3, columns) * min(3, rows),
	}
	for row := range rows {
		for column := range columns {
			cell := row*columns + column
			neighborIndex := 0
			for rowOffset := range min(3, rows) {
				neighborRow := (row + rowOffset - 1 + rows) % rows
				for columnOffset := range min(3, columns) {
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
	column := min(int(x/grid.cellWidth), grid.columns-1)
	row := min(int(y/grid.cellHeight), len(grid.heads)/grid.columns-1)
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
