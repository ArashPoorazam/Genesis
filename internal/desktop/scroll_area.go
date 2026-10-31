package desktop

import (
	uiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
)

func (view *Window) addScrollArea(parent *widget.Container, x, y, width, height, contentHeight int) *widget.Container {
	content := newContainer(width-18, contentHeight)
	scroll := widget.NewScrollContainer(
		widget.ScrollContainerOpts.Content(content),
		widget.ScrollContainerOpts.StretchContentWidth(),
		widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
			Idle: uiimage.NewNineSliceColor(panelBG),
			Mask: uiimage.NewNineSliceColor(ink),
		}),
	)
	placeWidget(parent, scroll, rectangle(x, y, width-14, height))
	view.scrollContainers = append(view.scrollContainers, scroll)

	scrollbar := widget.NewSlider(
		widget.SliderOpts.Orientation(widget.DirectionVertical),
		widget.SliderOpts.MinMax(0, 1000),
		widget.SliderOpts.FixedHandleSize(28),
		widget.SliderOpts.InitialCurrent(0),
		widget.SliderOpts.PageSizeFunc(func() int {
			return 100
		}),
		widget.SliderOpts.Images(
			&widget.SliderTrackImage{
				Idle: uiimage.NewNineSliceColor(panelBG), Hover: uiimage.NewNineSliceColor(surface),
			},
			&widget.ButtonImage{
				Idle: uiimage.NewNineSliceColor(border), Hover: uiimage.NewNineSliceColor(muted), Pressed: uiimage.NewNineSliceColor(accent),
			},
		),
		widget.SliderOpts.ChangedHandler(func(change *widget.SliderChangedEventArgs) {
			scroll.ScrollTop = float64(change.Current) / 1000
		}),
	)
	placeWidget(parent, scrollbar, rectangle(x+width-9, y, 6, height))
	scrollbar.GetWidget().Disabled = contentHeight <= height

	scroll.GetWidget().ScrolledEvent.AddHandler(func(arguments interface{}) {
		wheel := arguments.(*widget.WidgetScrolledEventArgs)
		distance := wheel.Y * 48 / float64(max(1, contentHeight-height))
		scroll.ScrollTop = max(0, min(1, scroll.ScrollTop-distance))
		scrollbar.Current = int(scroll.ScrollTop * 1000)
	})

	content.GetWidget().FocusEvent.AddHandler(func(arguments interface{}) {
		focus := arguments.(*widget.WidgetFocusEventArgs)
		if !focus.Focused || contentHeight <= height {
			return
		}

		visibleBounds := scroll.ViewRect()
		focusedBounds := focus.Widget.GetWidget().Rect
		offset := 0
		if focusedBounds.Min.Y < visibleBounds.Min.Y {
			offset = focusedBounds.Min.Y - visibleBounds.Min.Y - 8
		} else if focusedBounds.Max.Y > visibleBounds.Max.Y {
			offset = focusedBounds.Max.Y - visibleBounds.Max.Y + 8
		}

		scroll.ScrollTop = max(0, min(1, scroll.ScrollTop+float64(offset)/float64(contentHeight-height)))
		scrollbar.Current = int(scroll.ScrollTop * 1000)
	})

	return content
}
