package desktop

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"image"
	"math"
)

func (view *Window) updateMatrix() {
	if !view.inspectorVisible || view.activeTab != relationshipsTab {
		return
	}
	x, y := ebiten.CursorPosition()
	point := image.Pt(x, y)
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if point.In(view.matrixBounds) {
			view.selectedRow = view.firstVisibleRow + (y-view.matrixBounds.Min.Y)/view.cellHeight
			view.selectedColumn = view.firstVisibleColumn + (x-view.matrixBounds.Min.X)/view.cellWidth
			view.matrixFocused = true
			view.requestRebuild()
		} else {
			view.matrixFocused = false
		}
	}
	if point.In(view.matrixBounds) {
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
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) && len(view.textFields) > 0 {
		view.ui.SetFocusedWidget(view.textFields[0])
		view.matrixFocused = false
	}
}
