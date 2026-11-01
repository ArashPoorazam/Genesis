package desktop

import (
	"fmt"
	"image"
	"image/color"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func (view *Window) buildCompactDialog() {
	dialog := view.activeDialog
	bounds := view.dialogRect()
	x, y, width := bounds.Min.X+32, bounds.Min.Y, bounds.Dx()-64
	tint := accent
	if dialog.warning {
		tint = danger
	}
	view.addLabel(view.rootContainer, x+58, y+39, width-100, dialog.eyebrow, 13, tint)
	view.addButton(view.rootContainer, x+width-34, y+28, 34, "×", false, view.cancelDialog)
	view.addLabel(view.rootContainer, x, y+88, width, dialog.title, 26, ink)
	view.addLabel(view.rootContainer, x, y+130, width, dialog.body, 14, muted)
	dialog.build(view.rootContainer, x, y+180, width, bounds.Dy()-260)
	view.addLiveLabel(view.rootContainer, x, bounds.Max.Y-118, width, 13, danger, func() string {
		return view.status
	})
}

func (view *Window) addExperimentSummary(parent *widget.Container, x, y, width int, detail string, unsaved bool) *widget.Text {
	card := newContainer(width, 72)
	card.SetBackgroundImage(roundedImage(surface, border, 10))
	placeWidget(parent, card, rectangle(x, y, width, 72))
	name := view.controller.Name
	if name == "" {
		name = "Untitled experiment"
	}
	title := view.addLabel(card, 16, 12, width-130, view.fitText(name, float64(width-146), 15), 15, ink)
	view.addLabel(card, 16, 40, width-32, detail, 13, muted)
	if unsaved {
		view.addLabel(card, width-100, 13, 84, "● Unsaved", 13, danger)
	} else {
		config := view.controller.Simulation.Config
		view.addLabel(card, width-100, 13, 84, fmt.Sprintf("%d types", len(config.ParticleTypes)), 13, accent)
	}
	return title
}

func (view *Window) styleDestructiveButton(button *widget.Button) {
	button.SetImage(&widget.ButtonImage{
		Idle:    roundedImage(panelBG, border, 8),
		Hover:   roundedImage(mix(panelBG, danger, .12), mix(border, danger, .45), 8),
		Pressed: roundedImage(mix(panelBG, danger, .2), danger, 8),
	})
}

func (view *Window) drawDialogBackground(screen *ebiten.Image) {
	vector.FillRect(screen, 0, 0, float32(view.width), float32(view.height), color.NRGBA{R: 2, G: 7, B: 18, A: 196}, false)
	bounds := view.dialogRect()
	fillRounded(screen, bounds.Inset(-10).Add(image.Pt(0, 6)), 25, color.NRGBA{A: 38})
	fillRounded(screen, bounds.Inset(-4), 21, color.NRGBA{A: 70})
	fillRounded(screen, bounds, 18, border)
	fillRounded(screen, bounds.Inset(1), 17, panelBG)
	if view.activeDialog.height == 0 {
		return
	}
	x, y := bounds.Min.X+32, bounds.Min.Y+28
	tint := accent
	if view.activeDialog.warning {
		tint = danger
	}
	fillRounded(screen, rectangle(x, y, 42, 42), 12, mix(panelBG, tint, .1))
	options := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	options.GeoM.Scale(34/float64(view.atomSticker.Bounds().Dx()), 34/float64(view.atomSticker.Bounds().Dy()))
	options.GeoM.Translate(float64(x+4), float64(y+4))
	screen.DrawImage(view.atomSticker, options)
	vector.FillRect(screen, float32(bounds.Min.X+32), float32(bounds.Max.Y-82), float32(bounds.Dx()-64), 1, border, false)
}
