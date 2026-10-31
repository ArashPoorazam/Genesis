package desktop

import (
	"image"
	"image/color"

	uiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

var (
	canvas  = color.NRGBA{R: 10, G: 14, B: 18, A: 255}
	panelBG = color.NRGBA{R: 20, G: 26, B: 32, A: 255}
	surface = color.NRGBA{R: 29, G: 37, B: 45, A: 255}
	border  = color.NRGBA{R: 47, G: 58, B: 68, A: 255}
	ink     = color.NRGBA{R: 232, G: 233, B: 226, A: 255}
	muted   = color.NRGBA{R: 146, G: 162, B: 174, A: 255}
	accent  = color.NRGBA{R: 133, G: 218, B: 195, A: 255}
	danger  = color.NRGBA{R: 239, G: 163, B: 119, A: 255}
)

// fixedLayout positions controls in local coordinates, including scroll content.
type fixedLayout struct {
	width  int
	height int
}

func (layout fixedLayout) PreferredSize(_ []widget.PreferredSizeLocateableWidget) (int, int) {
	return layout.width, layout.height
}

func (layout fixedLayout) Layout(children []widget.PreferredSizeLocateableWidget, bounds image.Rectangle) {
	for _, child := range children {
		if childBounds, ok := child.GetWidget().LayoutData.(image.Rectangle); ok {
			child.SetLocation(childBounds.Add(bounds.Min))
		}
	}
}

func rectangle(x, y, width, height int) image.Rectangle {
	return image.Rect(x, y, x+width, y+height)
}

func newContainer(width, height int) *widget.Container {
	return widget.NewContainer(
		widget.ContainerOpts.Layout(fixedLayout{width: width, height: height}),
	)
}

func placeWidget(parent *widget.Container, child widget.PreferredSizeLocateableWidget, bounds image.Rectangle) {
	child.GetWidget().LayoutData = bounds
	parent.AddChild(child)
}

func (view *Window) fontFace(size float64) text.Face {
	return &text.GoTextFace{Source: view.font, Size: size}
}

func (view *Window) addLabel(parent *widget.Container, x, y, width int, caption string, size float64, textColor color.Color) *widget.Text {
	face := view.fontFace(size)
	label := widget.NewText(
		widget.TextOpts.Text(caption, &face, textColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
		widget.TextOpts.MaxWidth(float64(width)),
	)

	placeWidget(parent, label, rectangle(x, y, width, int(size*2.6)))
	return label
}

func (view *Window) addLiveLabel(parent *widget.Container, x, y, width int, size float64, textColor color.Color, caption func() string) {
	label := view.addLabel(parent, x, y, width, caption(), size, textColor)
	view.refreshBindings = append(view.refreshBindings, func() {
		label.Label = caption()
	})
}

func buttonImages(primary bool) *widget.ButtonImage {
	idle := surface
	hover := color.NRGBA{R: 43, G: 55, B: 64, A: 255}
	pressed := color.NRGBA{R: 53, G: 71, B: 78, A: 255}

	if primary {
		idle = color.NRGBA{R: 39, G: 70, B: 64, A: 255}
		hover = color.NRGBA{R: 48, G: 88, B: 77, A: 255}
		pressed = color.NRGBA{R: 60, G: 107, B: 94, A: 255}
	}

	return &widget.ButtonImage{
		Idle:     uiimage.NewBorderedNineSliceColor(idle, border, 1),
		Hover:    uiimage.NewBorderedNineSliceColor(hover, accent, 1),
		Pressed:  uiimage.NewBorderedNineSliceColor(pressed, accent, 1),
		Disabled: uiimage.NewBorderedNineSliceColor(panelBG, border, 1),
	}
}

func (view *Window) addButton(parent *widget.Container, x, y, width int, caption string, primary bool, action func()) *widget.Button {
	face := view.fontFace(14)
	button := widget.NewButton(
		widget.ButtonOpts.Image(buttonImages(primary)),
		widget.ButtonOpts.Text(caption, &face, &widget.ButtonTextColor{
			Idle: ink, Disabled: muted, Hover: ink,
		}),
		widget.ButtonOpts.TextPadding(&widget.Insets{
			Left: 8, Right: 8, Top: 6, Bottom: 6,
		}),
		widget.ButtonOpts.ClickedHandler(func(_ *widget.ButtonClickedEventArgs) {
			action()
		}),
	)

	// Space controls playback. Window.activateFocusedButton handles Enter separately.
	button.DisableDefaultKeys = true
	placeWidget(parent, button, rectangle(x, y, width, 34))
	return button
}

func (view *Window) addTextInput(parent *widget.Container, x, y, width int, initial string, submit func(string) error) *widget.TextInput {
	face := view.fontFace(14)
	input := widget.NewTextInput(
		widget.TextInputOpts.Face(&face),
		widget.TextInputOpts.Image(&widget.TextInputImage{
			Idle:     uiimage.NewBorderedNineSliceColor(canvas, border, 1),
			Disabled: uiimage.NewNineSliceColor(panelBG),
		}),
		widget.TextInputOpts.Color(&widget.TextInputColor{
			Idle: ink, Disabled: muted, Caret: accent, DisabledCaret: muted,
		}),
		widget.TextInputOpts.Padding(&widget.Insets{
			Left: 10, Right: 8, Top: 8, Bottom: 8,
		}),
	)
	input.SetText(initial)

	lastCommitted := initial
	commit := func() {
		value := input.GetText()
		if submit == nil || view.rebuilding || value == lastCommitted {
			return
		}
		if err := submit(value); err != nil {
			view.reportError(err)
			return
		}
		lastCommitted = value
	}

	input.SubmitEvent.AddHandler(func(_ interface{}) {
		commit()
	})
	input.GetWidget().FocusEvent.AddHandler(func(arguments interface{}) {
		focus := arguments.(*widget.WidgetFocusEventArgs)
		if !focus.Focused {
			commit()
		}
	})

	placeWidget(parent, input, rectangle(x, y, width, 34))
	view.textFields = append(view.textFields, input)
	return input
}
