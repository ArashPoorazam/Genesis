package desktop

import (
	"fmt"
	"image/color"
	"math"
	"strconv"

	"github.com/ArashPoorazam/Genesis/internal/simulation"
	uiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

// numericSetting keeps a control's label, range, and config edit together.
type numericSetting struct {
	Disabled bool
	Title    string
	Hint     string
	Value    float64
	Minimum  float64
	Maximum  float64
	Step     float64
	Apply    func(*simulation.Config, float64)
	Read     func(simulation.Config) float64
}

func (setting numericSetting) format(value float64) string {
	if setting.Step >= 1 {
		return fmt.Sprintf("%.0f", value)
	}
	return fmt.Sprintf("%.2f", value)
}

func (view *Window) addSlider(parent *widget.Container, x, y, width int, setting numericSetting, change func(float64) error) *widget.Slider {
	factor := 1 / setting.Step
	track := ebiten.NewImage(width, 24)
	view.transientImages = append(view.transientImages, track)
	graphic := widget.NewGraphic(widget.GraphicOpts.Image(track))
	placeWidget(parent, graphic, rectangle(x, y, width, 24))
	slider := widget.NewSlider(
		widget.SliderOpts.MinMax(int(math.Round(setting.Minimum*factor)), int(math.Round(setting.Maximum*factor))),
		widget.SliderOpts.InitialCurrent(int(math.Round(setting.Value*factor))),
		widget.SliderOpts.FixedHandleSize(16),
		widget.SliderOpts.TrackPadding(&widget.Insets{Top: 4, Bottom: 4}),
		widget.SliderOpts.PageSizeFunc(func() int {
			return 1
		}),
		widget.SliderOpts.Images(
			&widget.SliderTrackImage{
				Idle: uiimage.NewNineSliceColor(color.Transparent),
			},
			&widget.ButtonImage{
				Idle: roundedImage(accent, accent, 7), Hover: roundedImage(ink, accent, 7), Pressed: roundedImage(ink, ink, 7),
			},
		),
		widget.SliderOpts.ChangedHandler(func(changeEvent *widget.SliderChangedEventArgs) {
			if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
				view.controller.BeginGesture()
			}
			view.reportError(change(float64(changeEvent.Current) / factor))
		}),
	)
	lastDrawn := slider.Min - 1
	view.refreshBindings = append(view.refreshBindings, func() {
		if lastDrawn == slider.Current {
			return
		}
		lastDrawn = slider.Current
		track.Clear()
		fillRounded(track, rectangle(8, 10, width-16, 4), 2, border)
		fraction := float64(slider.Current-slider.Min) / float64(max(1, slider.Max-slider.Min))
		end := 8 + int(fraction*float64(width-16))
		start := 8
		tint := accent
		if setting.Minimum < 0 {
			start = width / 2
			if end < start {
				tint = pink
			}
		}
		fillRounded(track, rectangle(min(start, end), 10, max(1, abs(end-start)), 4), 2, tint)
	})

	placeWidget(parent, slider, rectangle(x, y, width, 24))
	return slider
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func (view *Window) addNumberControl(parent *widget.Container, x, y, width int, setting numericSetting) {
	view.addLabel(parent, x, y, width, setting.Title, 14, ink)
	errorLabel := view.addLabel(parent, x, y+66, width, "", 11, danger)

	applyValue := func(value float64) error {
		err := view.applyNumericSetting(setting, value)
		if err != nil {
			errorLabel.Label = err.Error()
			return err
		}
		errorLabel.Label = ""
		return nil
	}

	lastValue := setting.Value
	// The widget emits change events for programmatic Current updates too.
	// Record the expected position before assigning Current so those events do
	// not round the committed value or create an extra history entry.
	lastSlider := int(math.Round(setting.Value / setting.Step))
	var input *widget.TextInput
	var slider *widget.Slider
	slider = view.addSlider(parent, x, y+30, width-100, setting, func(value float64) error {
		position := int(math.Round(value / setting.Step))
		if position == lastSlider {
			return nil
		}
		lastSlider = position
		if err := applyValue(value); err != nil {
			return err
		}
		if setting.Read != nil {
			value = setting.Read(view.controller.Simulation.Config)
		}
		lastValue = value
		lastSlider = int(math.Round(value / setting.Step))
		slider.Current = lastSlider
		input.SetText(setting.format(lastValue))
		return nil
	})

	input = view.addTextInput(parent, x+width-86, y+24, 86, setting.format(setting.Value), func(raw string) error {
		value, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			errorLabel.Label = "Enter a number"
			return fmt.Errorf("%s: enter a number", setting.Title)
		}
		if setting.Step >= 1 && math.Trunc(value) != value {
			errorLabel.Label = "Enter a whole number"
			return fmt.Errorf("%s: enter a whole number", setting.Title)
		}
		if err := applyValue(value); err != nil {
			return err
		}

		if setting.Read != nil {
			value = setting.Read(view.controller.Simulation.Config)
		}
		lastValue = value
		lastSlider = int(math.Round(value / setting.Step))
		slider.Current = lastSlider
		input.SetText(setting.format(value))
		return nil
	})
	input.GetWidget().Disabled = setting.Disabled
	slider.GetWidget().Disabled = setting.Disabled
	view.refreshBindings = append(view.refreshBindings, func() {
		if setting.Read == nil {
			return
		}
		value := setting.Read(view.controller.Simulation.Config)
		if value != lastValue {
			lastValue = value
			if view.ui.GetFocusedWidget() != input {
				input.SetText(setting.format(value))
			}
			errorLabel.Label = ""
			lastSlider = int(math.Round(value / setting.Step))
			slider.Current = lastSlider
		}
	})

	if setting.Hint != "" {
		view.addLabel(parent, x, y+83, width, setting.Hint, 11, muted)
	}
}

func (view *Window) applyNumericSetting(setting numericSetting, value float64) error {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < setting.Minimum || value > setting.Maximum {
		return fmt.Errorf("enter a value between %g and %g", setting.Minimum, setting.Maximum)
	}
	return view.editConfig(func(config *simulation.Config) error {
		setting.Apply(config, value)
		return nil
	})
}
