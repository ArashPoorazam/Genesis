package desktop

import (
	"image"
	"image/color"

	uiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

var (
	canvas  = color.NRGBA{R: 3, G: 10, B: 26, A: 255}
	panelBG = color.NRGBA{R: 12, G: 22, B: 42, A: 255}
	surface = color.NRGBA{R: 22, G: 35, B: 57, A: 255}
	border  = color.NRGBA{R: 39, G: 57, B: 82, A: 255}
	ink     = color.NRGBA{R: 233, G: 243, B: 255, A: 255}
	muted   = color.NRGBA{R: 177, G: 193, B: 215, A: 255}
	accent  = color.NRGBA{R: 55, G: 201, B: 249, A: 255}
	pink    = color.NRGBA{R: 248, G: 91, B: 189, A: 255}
	danger  = color.NRGBA{R: 255, G: 177, B: 112, A: 255}
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
	size = max(13, size)
	if face := view.fontFaces[size]; face != nil {
		return face
	}
	face := &text.GoTextFace{Source: view.font, Size: size}
	weight := float32(550)
	if size >= 20 {
		weight = 650
	}
	face.SetVariation(text.MustParseTag("wght"), weight)
	view.fontFaces[size] = face
	return face
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
	hover := mix(surface, accent, .16)
	pressed := mix(surface, accent, .26)

	if primary {
		idle = color.NRGBA{R: 17, G: 75, B: 103, A: 255}
		hover = mix(idle, accent, .25)
		pressed = mix(idle, accent, .4)
	}

	return &widget.ButtonImage{
		Idle:     roundedImage(idle, idle, 8),
		Hover:    roundedImage(hover, mix(border, accent, .4), 8),
		Pressed:  roundedImage(pressed, accent, 8),
		Disabled: roundedImage(panelBG, surface, 8),
	}
}

func (view *Window) addButton(parent *widget.Container, x, y, width int, caption string, primary bool, action func()) *widget.Button {
	face := view.fontFace(15)
	images := buttonImages(primary)
	button := widget.NewButton(
		widget.ButtonOpts.Image(images),
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
	button.SetImage(images)
	placeWidget(parent, button, rectangle(x, y, width, 34))
	return button
}

func (view *Window) addTextInput(parent *widget.Container, x, y, width int, initial string, submit func(string) error) *widget.TextInput {
	face := view.fontFace(15)
	input := widget.NewTextInput(
		widget.TextInputOpts.Face(&face),
		widget.TextInputOpts.Image(&widget.TextInputImage{
			Idle:     roundedImage(canvas, border, 7),
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
