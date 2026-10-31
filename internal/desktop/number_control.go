package desktop

import (
	"fmt"
	"math"
	"strconv"

	"github.com/ArashPoorazam/Genesis/internal/simulation"
	uiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

// numericSetting keeps a control's label, range, and config edit together.
type numericSetting struct {
	Title   string
	Hint    string
	Value   float64
	Minimum float64
	Maximum float64
	Step    float64
	Apply   func(*simulation.Config, float64)
}

func (setting numericSetting) format(value float64) string {
	if setting.Step >= 1 {
		return fmt.Sprintf("%.0f", value)
	}
	return fmt.Sprintf("%.2f", value)
}

func (view *Window) addSlider(parent *widget.Container, x, y, width int, setting numericSetting, change func(float64) error) *widget.Slider {
	factor := 1 / setting.Step
	slider := widget.NewSlider(
		widget.SliderOpts.MinMax(int(math.Round(setting.Minimum*factor)), int(math.Round(setting.Maximum*factor))),
		widget.SliderOpts.InitialCurrent(int(math.Round(setting.Value*factor))),
		widget.SliderOpts.FixedHandleSize(12),
		widget.SliderOpts.TrackPadding(&widget.Insets{Top: 4, Bottom: 4}),
		widget.SliderOpts.PageSizeFunc(func() int {
			return 1
		}),
		widget.SliderOpts.Images(
			&widget.SliderTrackImage{
				Idle: uiimage.NewNineSliceColor(border), Hover: uiimage.NewNineSliceColor(muted),
			},
			&widget.ButtonImage{
				Idle: uiimage.NewNineSliceColor(accent), Hover: uiimage.NewNineSliceColor(ink), Pressed: uiimage.NewNineSliceColor(ink),
			},
		),
		widget.SliderOpts.ChangedHandler(func(changeEvent *widget.SliderChangedEventArgs) {
			if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
				view.controller.BeginGesture()
			}
			view.reportError(change(float64(changeEvent.Current) / factor))
		}),
	)

	placeWidget(parent, slider, rectangle(x, y, width, 24))
	return slider
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

	var input *widget.TextInput
	slider := view.addSlider(parent, x, y+30, width-100, setting, func(value float64) error {
		if err := applyValue(value); err != nil {
			return err
		}
		input.SetText(setting.format(value))
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

		slider.Current = int(math.Round(value / setting.Step))
		return nil
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
