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
	bounds := view.matrixBounds.Add(view.inspectorRect().Min)
	card := view.relationshipCard.Add(view.inspectorRect().Min)
	fillRounded(screen, card, 12, border)
	fillRounded(screen, card.Inset(1), 11, surface)
	rowColor := config.ParticleTypes[view.selectedRow].Color
	columnColor := config.ParticleTypes[view.selectedColumn].Color
	vector.FillCircle(screen, float32(card.Min.X+24), float32(card.Min.Y+48), 9, mix(surface, rowColor, .18), true)
	vector.FillCircle(screen, float32(card.Min.X+21), float32(card.Min.Y+46), 5, rowColor, true)
	vector.FillCircle(screen, float32(card.Min.X+29), float32(card.Min.Y+51), 5, columnColor, true)
	for visibleColumn := 0; visibleColumn < view.matrixCols; visibleColumn++ {
		typeIndex := visibleColumn + view.firstVisibleColumn
		if typeIndex >= len(config.ParticleTypes) {
			continue
		}
		particleType := config.ParticleTypes[typeIndex]
		x := bounds.Min.X + visibleColumn*view.cellWidth
		center := float64(x + view.cellWidth/2)
		vector.FillCircle(screen, float32(center), float32(bounds.Min.Y-28), 3, particleType.Color, true)
		view.centeredText(screen, view.fitText(particleType.Name, float64(view.cellWidth-8), 13), center, float64(bounds.Min.Y-21), 13, muted)
	}
	for visibleRow := 0; visibleRow < view.matrixRows; visibleRow++ {
		rowIndex := visibleRow + view.firstVisibleRow
		if rowIndex >= len(config.ParticleTypes) {
			continue
		}
		particleType := config.ParticleTypes[rowIndex]
		y := bounds.Min.Y + visibleRow*view.cellHeight
		vector.FillCircle(screen, float32(bounds.Min.X-90), float32(y+view.cellHeight/2), 4, particleType.Color, true)
		view.drawText(screen, shorten(particleType.Name, 9), float64(bounds.Min.X-79), float64(y+view.cellHeight/2-9), 13, muted)
		for visibleColumn := 0; visibleColumn < view.matrixCols; visibleColumn++ {
			columnIndex := visibleColumn + view.firstVisibleColumn
			if columnIndex >= len(config.ParticleTypes) {
				continue
			}
			strength := config.Relationships[rowIndex][columnIndex]
			x := bounds.Min.X + visibleColumn*view.cellWidth
			target := color.NRGBA{17, 111, 132, 255}
			if strength < 0 {
				target = color.NRGBA{132, 42, 100, 255}
			}
			fill := mix(surface, target, math.Min(1, math.Abs(strength)/math.Max(2, config.Generation.RelationshipAmplitude)))
			cell := rectangle(x+3, y+3, view.cellWidth-6, view.cellHeight-6)
			if rowIndex == view.selectedRow || columnIndex == view.selectedColumn {
				fill = mix(fill, ink, .06)
			}
			if rowIndex == view.selectedRow && columnIndex == view.selectedColumn {
				fillRounded(screen, cell.Inset(-2), 9, accent)
			}
			fillRounded(screen, cell, 7, fill)
			view.centeredText(screen, fmt.Sprintf("%+.2f", strength), float64(x+view.cellWidth/2), float64(y+view.cellHeight/2-9), 13, ink)
			if config.RelationshipLocked(rowIndex, columnIndex) {
				drawMatrixLock(screen, cell.Max.X-7, cell.Min.Y+1)
			}
		}
	}
}

var matrixLockIcon *ebiten.Image

func drawMatrixLock(screen *ebiten.Image, x, y int) {
	if matrixLockIcon == nil {
		matrixLockIcon = ebiten.NewImage(12, 14)
		fillRounded(matrixLockIcon, matrixLockIcon.Bounds(), 4, panelBG)
		vector.StrokeCircle(matrixLockIcon, 6, 5, 2.5, 1.3, danger, true)
		fillRounded(matrixLockIcon, rectangle(2, 6, 8, 6), 2, danger)
	}
	options := &ebiten.DrawImageOptions{}
	options.GeoM.Translate(float64(x-2), float64(y-2))
	screen.DrawImage(matrixLockIcon, options)
}
