package desktop

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func mix(fromColor, toColor color.NRGBA, amount float64) color.NRGBA {
	return color.NRGBA{
		R: uint8(float64(fromColor.R) + (float64(toColor.R)-float64(fromColor.R))*amount),
		G: uint8(float64(fromColor.G) + (float64(toColor.G)-float64(fromColor.G))*amount),
		B: uint8(float64(fromColor.B) + (float64(toColor.B)-float64(fromColor.B))*amount),
		A: 255,
	}

}

func (view *Window) drawMatrix(screen *ebiten.Image) {
	if !view.inspectorVisible || view.activeTab != relationshipsTab {
		return
	}
	config := view.controller.Simulation.Config
	bounds := view.matrixBounds
	for visibleColumn := 0; visibleColumn < view.matrixCols; visibleColumn++ {
		typeIndex := visibleColumn + view.firstVisibleColumn
		if typeIndex >= len(config.ParticleTypes) {
			continue
		}
		particleType := config.ParticleTypes[typeIndex]
		x := bounds.Min.X + visibleColumn*view.cellWidth
		vector.FillCircle(screen, float32(x+view.cellWidth/2), float32(bounds.Min.Y-25), 3, particleType.Color, true)
		view.drawText(screen, shorten(particleType.Name, max(3, view.cellWidth/7)), float64(x+5), float64(bounds.Min.Y-18), 11, muted)
	}
	for visibleRow := 0; visibleRow < view.matrixRows; visibleRow++ {
		rowIndex := visibleRow + view.firstVisibleRow
		if rowIndex >= len(config.ParticleTypes) {
			continue
		}
		particleType := config.ParticleTypes[rowIndex]
		y := bounds.Min.Y + visibleRow*view.cellHeight
		vector.FillCircle(screen, float32(bounds.Min.X-90), float32(y+view.cellHeight/2), 4, particleType.Color, true)
		view.drawText(screen, shorten(particleType.Name, 10), float64(bounds.Min.X-79), float64(y+view.cellHeight/2-8), 12, muted)
		for visibleColumn := 0; visibleColumn < view.matrixCols; visibleColumn++ {
			columnIndex := visibleColumn + view.firstVisibleColumn
			if columnIndex >= len(config.ParticleTypes) {
				continue
			}
			strength := config.Relationships[rowIndex][columnIndex]
			x := bounds.Min.X + visibleColumn*view.cellWidth
			target := color.NRGBA{39, 100, 86, 255}
			if strength < 0 {
				target = color.NRGBA{137, 74, 47, 255}
			}
			fill := mix(surface, target, math.Min(1, math.Abs(strength)/math.Max(2, config.Generation.RelationshipAmplitude)))
			vector.FillRect(screen, float32(x+2), float32(y+2), float32(view.cellWidth-4), float32(view.cellHeight-4), fill, false)
			if rowIndex == view.selectedRow || columnIndex == view.selectedColumn {
				vector.StrokeRect(screen, float32(x+2), float32(y+2), float32(view.cellWidth-4), float32(view.cellHeight-4), 1, border, false)
			}
			if rowIndex == view.selectedRow && columnIndex == view.selectedColumn {
				vector.StrokeRect(screen, float32(x+1), float32(y+1), float32(view.cellWidth-2), float32(view.cellHeight-2), 2, accent, false)
			}
			view.drawText(screen, fmt.Sprintf("%+.2f", strength), float64(x+view.cellWidth/2-18), float64(y+view.cellHeight/2-8), 12, ink)
		}
	}
}
