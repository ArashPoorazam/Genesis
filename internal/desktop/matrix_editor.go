package desktop

import (
	"fmt"
	"image"
	"math"
	"math/rand/v2"

	"github.com/ArashPoorazam/Genesis/internal/simulation"
)

// matrixLayout reserves space for the selected rule and global actions before
// deciding how many rows fit. Large matrices keep their own navigation.
func matrixLayout(panelHeight, y, width, typeCount int, expanded bool) (image.Rectangle, int, int, int, int) {
	cellHeight := 36
	columns := min(typeCount, max(1, (width-100)/56))
	cellWidth := (width - 100) / columns
	top := y + 108
	available := panelHeight - top - 310
	if expanded {
		cellHeight = 42
		available = panelHeight - top - 52
	}
	rows := min(typeCount, max(1, available/cellHeight))
	if rows < typeCount || columns < typeCount {
		rows = min(typeCount, max(1, (available-44)/cellHeight))
	}
	return rectangle(20+100, top, columns*cellWidth, rows*cellHeight), rows, columns, cellWidth, cellHeight
}

func (view *Window) buildMatrix(x, y, width int) {
	typeCount := len(view.controller.Simulation.Config.ParticleTypes)
	view.addLabel(view.inspectorContainer, x, y, width-100, "Small rules. Big possibilities.", 20, ink)
	expandCaption := "Expand"
	if view.matrixExpanded {
		expandCaption = "Collapse"
	}
	view.addButton(view.inspectorContainer, x+width-84, y-2, 84, expandCaption, false, func() {
		view.matrixExpanded = !view.matrixExpanded
		view.requestRebuild()
	})
	view.addLabel(view.inspectorContainer, x, y+34, width, "Shape a connection · Right-click a cell to lock it.", 13, muted)
	fullWidth := width
	if view.matrixExpanded {
		width -= 340
	}
	view.addLabel(view.inspectorContainer, x+100, y+57, width-100, "INFLUENCED BY", 13, muted)
	view.addLabel(view.inspectorContainer, x, y+83, 96, "RESPONDS", 13, muted)
	view.matrixBounds, view.matrixRows, view.matrixCols, view.cellWidth, view.cellHeight = matrixLayout(view.inspectorRect().Dy(), y, width, typeCount, view.matrixExpanded)
	view.firstVisibleRow = max(0, min(view.firstVisibleRow, typeCount-view.matrixRows))
	view.firstVisibleColumn = max(0, min(view.firstVisibleColumn, typeCount-view.matrixCols))
	bottom := view.matrixBounds.Max.Y
	if typeCount > view.matrixRows || typeCount > view.matrixCols {
		rowRange := fmt.Sprintf("Rows %d–%d of %d", view.firstVisibleRow+1, view.firstVisibleRow+view.matrixRows, typeCount)
		view.addLabel(view.inspectorContainer, x, bottom+18, width-176, rowRange, 13, muted)
		view.addButton(view.inspectorContainer, x+width-168, bottom+10, 36, "←", false, func() {
			view.firstVisibleColumn = max(0, view.firstVisibleColumn-1)
			view.requestRebuild()
		})
		view.addButton(view.inspectorContainer, x+width-124, bottom+10, 36, "→", false, func() {
			view.firstVisibleColumn = min(typeCount-view.matrixCols, view.firstVisibleColumn+1)
			view.requestRebuild()
		})
		view.addButton(view.inspectorContainer, x+width-80, bottom+10, 36, "↑", false, func() {
			view.firstVisibleRow = max(0, view.firstVisibleRow-1)
			view.requestRebuild()
		})
		view.addButton(view.inspectorContainer, x+width-36, bottom+10, 36, "↓", false, func() {
			view.firstVisibleRow = min(typeCount-view.matrixRows, view.firstVisibleRow+1)
			view.requestRebuild()
		})
		bottom += 44
	}
	view.addLabel(view.inspectorContainer, x, bottom+14, width/3, "− Repel", 13, pink)
	view.addLabel(view.inspectorContainer, x+width/3, bottom+14, width/3, "0 Neutral", 13, muted)
	view.addLabel(view.inspectorContainer, x+width*2/3, bottom+14, width/3, "+ Attract", 13, accent)
	editorY := bottom + 48
	if view.matrixExpanded {
		x += fullWidth - 300
		width = 300
		editorY = y + 65
	}
	view.relationshipCard = rectangle(x, editorY, width, 200)
	view.buildRelationshipControls(x, editorY, width)
}

func (view *Window) buildRelationshipControls(x, editorY, width int) {
	selectedRow, selectedColumn := view.selectedRow, view.selectedColumn
	config := view.controller.Simulation.Config
	view.addLabel(view.inspectorContainer, x+16, editorY+12, width-130, "SELECTED CONNECTION", 13, muted)
	state := view.addLabel(view.inspectorContainer, x+width-86, editorY+12, 76, "", 13, muted)
	view.refreshBindings = append(view.refreshBindings, func() {
		strength := view.controller.Simulation.Config.Relationships[selectedRow][selectedColumn]
		if view.controller.Simulation.Config.RelationshipLocked(selectedRow, selectedColumn) {
			state.Label = "Locked"
			state.SetColor(danger)
			return
		}
		state.Label = "Neutral"
		state.SetColor(muted)
		if strength < 0 {
			state.Label = "Repels"
			state.SetColor(pink)
		} else if strength > 0 {
			state.Label = "Attracts"
			state.SetColor(accent)
		}
	})
	title := fmt.Sprintf("%s feels %s", shorten(config.ParticleTypes[selectedRow].Name, 10), shorten(config.ParticleTypes[selectedColumn].Name, 10))
	view.addLabel(view.inspectorContainer, x+44, editorY+38, width-58, title, 15, ink)
	cardX, cardWidth := x, width
	x += 16
	width -= 32
	limit := math.Max(2, config.Generation.RelationshipAmplitude)
	for _, relationships := range config.Relationships {
		for _, strength := range relationships {
			limit = math.Max(limit, math.Abs(strength))
		}
	}

	view.addNumberControl(view.inspectorContainer, x, editorY+66, width, numericSetting{
		Title:    "Strength",
		Disabled: config.RelationshipLocked(selectedRow, selectedColumn),
		Read:     func(config simulation.Config) float64 { return config.Relationships[selectedRow][selectedColumn] },
		Hint:     "",
		Value:    config.Relationships[selectedRow][selectedColumn],
		Minimum:  -limit,
		Maximum:  limit,
		Step:     .01,
		Apply: func(config *simulation.Config, strength float64) {
			if config.RelationshipLocked(selectedRow, selectedColumn) {
				return
			}
			config.SetRelationship(selectedRow, selectedColumn, strength)
			if view.linked {
				config.SetRelationship(selectedColumn, selectedRow, strength)
			}
		},
	})
	zero := view.addButton(view.inspectorContainer, x, editorY+150, 100, "Zero cell", false, func() {
		view.editConfig(func(config *simulation.Config) error {
			if config.RelationshipLocked(selectedRow, selectedColumn) {
				return nil
			}
			config.SetRelationship(selectedRow, selectedColumn, 0)
			if view.linked {
				config.SetRelationship(selectedColumn, selectedRow, 0)
			}
			return nil
		})
		view.requestRebuild()
	})
	zeroImages := buttonImages(false)
	zeroImages.Idle = roundedImage(panelBG, panelBG, 8)
	zero.SetImage(zeroImages)
	zero.GetWidget().Disabled = config.RelationshipLocked(selectedRow, selectedColumn)
	link := "Link pair: off"
	if view.linked {
		link = "Link pair: on"
	}

	reciprocal := view.addButton(view.inspectorContainer, x+110, editorY+150, width-110, link, view.linked, func() {
		view.linked = !view.linked
		view.requestRebuild()
	})
	if !view.linked {
		reciprocal.SetImage(zeroImages)
	}
	x, width = cardX, cardWidth
	view.addButton(view.inspectorContainer, x, editorY+212, width-108, "Shuffle relationships", true, func() {
		neutral := -1
		if err := view.editConfig(func(config *simulation.Config) error {
			neutral = config.RandomizeRelations(rand.Uint64())
			return nil
		}); err != nil {
			return
		}
		view.status = "Relationships shuffled · Locked cells kept"
		if neutral >= 0 {
			view.status += " · " + view.controller.Simulation.Config.ParticleTypes[neutral].Name + " is neutral"
		}
		view.requestRebuild()
	})
	view.addButton(view.inspectorContainer, x+width-100, editorY+212, 100, "Zero all", false, func() {
		view.editConfig(func(config *simulation.Config) error {
			for i := range config.Relationships {
				for j := range config.Relationships[i] {
					config.SetRelationship(i, j, 0)
				}
			}
			return nil
		})
		view.requestRebuild()
	})
	if view.matrixExpanded {
		view.addLabel(view.inspectorContainer, x, editorY+264, width, "Every row feels the column you choose.\nInner repulsion stays active at zero.\n\nRight-click a cell to lock or unlock.\nArrow keys explore · L toggles lock.", 13, muted)
	}
}
