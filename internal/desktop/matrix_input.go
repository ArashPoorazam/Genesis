package desktop

import (
	"image"
	"math"

	"github.com/ArashPoorazam/Genesis/internal/simulation"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (view *Window) matrixPointerPress(point image.Point, button ebiten.MouseButton) {
	if !view.inspectorVisible || view.activeTab != relationshipsTab || view.activeDialog != nil {
		return
	}
	bounds := view.matrixBounds.Add(view.inspectorRect().Min)
	if !point.In(bounds) {
		view.matrixFocused = false
		return
	}
	view.ui.ClearFocus()
	view.selectedRow = view.firstVisibleRow + (point.Y-bounds.Min.Y)/view.cellHeight
	view.selectedColumn = view.firstVisibleColumn + (point.X-bounds.Min.X)/view.cellWidth
	view.matrixFocused = true
	if button == ebiten.MouseButtonRight {
		view.toggleSelectedRelationshipLock()
	}
	view.requestRebuild()
}

func (view *Window) toggleSelectedRelationshipLock() {
	view.controller.EndGesture()
	view.editConfig(func(config *simulation.Config) error {
		config.ToggleRelationshipLock(view.selectedRow, view.selectedColumn)
		return nil
	})
	view.requestRebuild()
}

func (view *Window) updateMatrix() {
	if !view.inspectorVisible || view.activeTab != relationshipsTab || view.activeDialog != nil {
		return
	}
	x, y := ebiten.CursorPosition()
	bounds := view.matrixBounds.Add(view.inspectorRect().Min)
	point := image.Pt(x, y)
	for _, button := range []ebiten.MouseButton{ebiten.MouseButtonLeft, ebiten.MouseButtonRight} {
		if inpututil.IsMouseButtonJustPressed(button) {
			view.matrixPointerPress(point, button)
		}
	}
	if point.In(bounds) {
		wheelX, wheelY := ebiten.Wheel()
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			wheelX, wheelY = wheelY, 0
		}
		if wheelX != 0 || wheelY != 0 {
			view.firstVisibleRow -= int(math.Copysign(math.Ceil(math.Abs(wheelY)), wheelY))
			view.firstVisibleColumn -= int(math.Copysign(math.Ceil(math.Abs(wheelX)), wheelX))
			view.requestRebuild()
		}
	}
	if !view.matrixFocused || view.ui.GetFocusedWidget() != nil {
		return
	}
	previousRow, previousColumn := view.selectedRow, view.selectedColumn
	typeCount := len(view.controller.Simulation.Config.ParticleTypes)
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		view.selectedRow = min(typeCount-1, view.selectedRow+1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		view.selectedRow = max(0, view.selectedRow-1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
		view.selectedColumn = min(typeCount-1, view.selectedColumn+1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
		view.selectedColumn = max(0, view.selectedColumn-1)
	}
	if previousRow != view.selectedRow || previousColumn != view.selectedColumn {
		view.firstVisibleRow = max(view.selectedRow-view.matrixRows+1, min(view.firstVisibleRow, view.selectedRow))
		view.firstVisibleColumn = max(view.selectedColumn-view.matrixCols+1, min(view.firstVisibleColumn, view.selectedColumn))
		view.requestRebuild()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyL) {
		view.toggleSelectedRelationshipLock()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) && len(view.textFields) > 0 && !view.textFields[0].GetWidget().Disabled {
		view.ui.SetFocusedWidget(view.textFields[0])
		view.matrixFocused = false
	}
}
