package desktop

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"github.com/ArashPoorazam/Genesis/internal/simulation"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

// Keep the draft outside the dialog builder so resizing preserves the selection.
func (view *Window) openColorPicker(typeIndex int) {
	draft := view.controller.Simulation.Config.ParticleTypes[typeIndex].Color
	view.openDialog(&dialog{
		title:   "Custom particle color",
		height:  588,
		eyebrow: "PARTICLE APPEARANCE",
		body:    "Mix red, green and blue, or enter a hex color.",
		build: func(parent *widget.Container, x, y, width, height int) {
			preview := ebiten.NewImage(width, 48)
			view.transientImages = append(view.transientImages, preview)
			placeWidget(parent, widget.NewGraphic(widget.GraphicOpts.Image(preview)), rectangle(x, y, width, 48))
			view.refreshBindings = append(view.refreshBindings, func() {
				preview.Clear()
				fillRounded(preview, preview.Bounds(), 8, draft)
			})

			var hexInput *widget.TextInput
			var sliders [3]*widget.Slider
			channels := []*uint8{&draft.R, &draft.G, &draft.B}
			hexCaption := func() string { return fmt.Sprintf("#%02X%02X%02X", draft.R, draft.G, draft.B) }
			for index, name := range []string{"Red", "Green", "Blue"} {
				channel := channels[index]
				rowY := y + 68 + index*48
				view.addLabel(parent, x, rowY+3, 66, name, 14, ink)
				view.addLiveLabel(parent, x+width-42, rowY+3, 42, 14, ink, func() string {
					return strconv.Itoa(int(*channel))
				})
				sliders[index] = view.addSlider(parent, x+72, rowY, width-130, numericSetting{
					Session: true, Value: float64(*channel), Minimum: 0, Maximum: 255, Step: 1,
				}, func(value float64) error {
					if *channel != uint8(value) {
						*channel = uint8(value)
						hexInput.SetText(hexCaption())
						view.status = ""
					}
					return nil
				})
			}
			parseHex := func(raw string) error {
				raw = strings.TrimPrefix(strings.TrimSpace(raw), "#")
				value, err := strconv.ParseUint(raw, 16, 24)
				if len(raw) != 6 || err != nil {
					return fmt.Errorf("use a six-digit hex color, e.g. #FF4655")
				}
				draft = color.NRGBA{R: uint8(value >> 16), G: uint8(value >> 8), B: uint8(value), A: 255}
				for index, channel := range channels {
					sliders[index].Current = int(*channel)
				}
				view.status = ""
				return nil
			}
			view.addLabel(parent, x, y+220, 66, "Hex", 14, ink)
			hexInput = view.addTextInput(parent, x+72, y+212, width-72, hexCaption(), parseHex)
			view.addButton(parent, x+width-252, view.dialogRect().Max.Y-58, 110, "Cancel", false, view.cancelDialog)
			view.addButton(parent, x+width-130, view.dialogRect().Max.Y-58, 130, "Apply color", true, func() {
				// Validate even when Apply is clicked before the field loses focus.
				if err := parseHex(hexInput.GetText()); err != nil {
					view.reportError(err)
					return
				}
				if err := view.editConfig(func(config *simulation.Config) error {
					config.ParticleTypes[typeIndex].Color = draft
					return nil
				}); err == nil {
					view.closeDialog()
				}
			})
		},
	})
}
