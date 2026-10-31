package desktop

import (
	"fmt"
	"math"
	"math/rand/v2"

	"github.com/ArashPoorazam/Genesis/internal/simulation"
)

func (view *Window) buildMatrix(x, y, width int) {
	typeCount := len(view.controller.Simulation.Config.ParticleTypes)
	view.addLabel(view.rootContainer, x, y, width-100, "Interaction rules", 23, ink)
	expandCaption := "Expand"
	if view.matrixExpanded {
		expandCaption = "Collapse"
	}

	view.addButton(view.rootContainer, x+width-94, y-2, 94, expandCaption, false, func() {
		view.matrixExpanded = !view.matrixExpanded
		view.requestRebuild()
	})
	view.addLabel(view.rootContainer, x, y+34, width, "One rule for every ordered pair of particle types.", 12, muted)
	fullWidth := width
	if view.matrixExpanded {
		width -= 340
	}

	view.addLabel(view.rootContainer, x+100, y+63, width-100, "INFLUENCING PARTICLE →", 10, muted)
	view.addLabel(view.rootContainer, x, y+88, 96, "RESPONDING\nPARTICLE ↓", 9, muted)
	view.cellHeight = 42
	maxRows := min(5, max(2, (view.height-570)/42))
	if typeCount <= 5 {
		// A five-type matrix does not need the extra navigation row.
		maxRows = 5
	}
	if view.matrixExpanded {
		view.cellHeight = 32
		maxRows = max(2, (view.height-y-210)/32)
	}

	view.matrixRows = min(typeCount, maxRows)
	view.matrixCols = min(typeCount, max(1, (width-100)/54))
	view.cellWidth = (width - 100) / view.matrixCols
	view.firstVisibleRow = max(0, min(view.firstVisibleRow, typeCount-view.matrixRows))
	view.firstVisibleColumn = max(0, min(view.firstVisibleColumn, typeCount-view.matrixCols))
	view.matrixBounds = rectangle(x+100, y+112, view.matrixCols*view.cellWidth, view.matrixRows*view.cellHeight)
	bottom := view.matrixBounds.Max.Y
	if typeCount > view.matrixRows || typeCount > view.matrixCols {
		rowRange := fmt.Sprintf("Rows %d–%d / %d · wheel to scroll", view.firstVisibleRow+1, view.firstVisibleRow+view.matrixRows, typeCount)
		view.addLabel(view.rootContainer, x, bottom+6, width-176, rowRange, 10, muted)
		view.addButton(view.rootContainer, x+width-168, bottom+4, 38, "←", false, func() {
			view.firstVisibleColumn = max(0, view.firstVisibleColumn-1)
			view.requestRebuild()
		})
		view.addButton(view.rootContainer, x+width-126, bottom+4, 38, "→", false, func() {
			view.firstVisibleColumn = min(typeCount-view.matrixCols, view.firstVisibleColumn+1)
			view.requestRebuild()
		})
		view.addButton(view.rootContainer, x+width-84, bottom+4, 38, "↑", false, func() {
			view.firstVisibleRow = max(0, view.firstVisibleRow-1)
			view.requestRebuild()
		})
		view.addButton(view.rootContainer, x+width-42, bottom+4, 38, "↓", false, func() {
			view.firstVisibleRow = min(typeCount-view.matrixRows, view.firstVisibleRow+1)
			view.requestRebuild()
		})
		bottom += 40
	}

	view.addLabel(view.rootContainer, x, bottom+16, width, "− Repel                 0 Neutral                 + Attract", 11, muted)
	editorY := bottom + 53
	if view.matrixExpanded {
		x += fullWidth - 300
		width = 300
		editorY = y + 100
	}

	view.buildRelationshipControls(x, editorY, width)
}

func (view *Window) buildRelationshipControls(x, editorY, width int) {
	selectedRow, selectedColumn := view.selectedRow, view.selectedColumn
	config := view.controller.Simulation.Config
	title := fmt.Sprintf("%s responds to %s", config.ParticleTypes[selectedRow].Name, config.ParticleTypes[selectedColumn].Name)
	limit := math.Max(2, config.Generation.RelationshipAmplitude)
	for _, relationships := range config.Relationships {
		for _, strength := range relationships {
			limit = math.Max(limit, math.Abs(strength))
		}
	}

	view.addNumberControl(view.rootContainer, x, editorY, width, numericSetting{
		Title:   title,
		Hint:    "",
		Value:   config.Relationships[selectedRow][selectedColumn],
		Minimum: -limit,
		Maximum: limit,
		Step:    .01,
		Apply: func(config *simulation.Config, strength float64) {
			config.Relationships[selectedRow][selectedColumn] = strength
			if view.linked {
				config.Relationships[selectedColumn][selectedRow] = strength
			}
		},
	})
	view.addButton(view.rootContainer, x, editorY+88, 100, "Zero cell", false, func() {
		view.editConfig(func(config *simulation.Config) error {
			config.Relationships[selectedRow][selectedColumn] = 0
			if view.linked {
				config.Relationships[selectedColumn][selectedRow] = 0
			}
			return nil
		})
		view.requestRebuild()
	})
	link := "Reciprocal edits: off"
	if view.linked {
		link = "Reciprocal edits: on"
	}

	view.addButton(view.rootContainer, x+110, editorY+88, width-110, link, view.linked, func() {
		view.linked = !view.linked
		view.requestRebuild()
	})
	view.addButton(view.rootContainer, x, editorY+134, width-108, "Randomize relationships", true, func() {
		view.editConfig(func(config *simulation.Config) error {
			config.RandomizeRelations(rand.Uint64())
			return nil
		})
		view.status = "Relationships randomized; particles preserved"
		view.requestRebuild()
	})
	view.addButton(view.rootContainer, x+width-100, editorY+134, 100, "Zero all", false, func() {
		view.editConfig(func(config *simulation.Config) error {
			for i := range config.Relationships {
				for j := range config.Relationships[i] {
					config.Relationships[i][j] = 0
				}
			}
			return nil
		})
		view.requestRebuild()
	})
	view.addLabel(view.rootContainer, x, editorY+183, width, "Inner-zone repulsion remains active at zero.\nSelect a cell, then use arrow keys to explore.", 11, muted)
}
