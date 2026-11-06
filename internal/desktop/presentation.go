package desktop

import (
	"image"
	"image/color"
	"math"

	uiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type surfaceKey struct {
	fill, edge color.NRGBA
	radius     int
}

var surfaceImages = map[surfaceKey]*uiimage.NineSlice{}
var shapeImages = map[int]*uiimage.NineSlice{}

func fillRounded(screen *ebiten.Image, bounds image.Rectangle, radius float32, fill color.Color) {
	if bounds.Empty() {
		return
	}
	corner := max(0, min(int(radius), (bounds.Dx()-1)/2, (bounds.Dy()-1)/2))
	// Reuse a white corner texture and tint at draw time. Caching by radius
	// avoids rasterizing each matrix cell or storing every possible cell color.
	shape := shapeImages[corner]
	if shape == nil {
		size := corner*2 + 3
		texture := ebiten.NewImage(size, size)
		rasterizeRounded(texture, texture.Bounds(), float32(corner), color.White)
		shape = uiimage.NewNineSlice(texture, [3]int{corner + 1, 1, corner + 1}, [3]int{corner + 1, 1, corner + 1})
		shapeImages[corner] = shape
	}
	shape.Draw(screen, bounds.Dx(), bounds.Dy(), func(options *ebiten.DrawImageOptions) {
		options.GeoM.Translate(float64(bounds.Min.X), float64(bounds.Min.Y))
		options.ColorScale.ScaleWithColor(fill)
	})
}

func rasterizeRounded(screen *ebiten.Image, bounds image.Rectangle, radius float32, fill color.Color) {
	left, top := float32(bounds.Min.X), float32(bounds.Min.Y)
	width, height := float32(bounds.Dx()), float32(bounds.Dy())
	radius = min(radius, width/2, height/2)
	vector.FillRect(screen, left+radius, top, width-radius*2, height, fill, true)
	vector.FillRect(screen, left, top+radius, radius, height-radius*2, fill, true)
	vector.FillRect(screen, left+width-radius, top+radius, radius, height-radius*2, fill, true)
	for _, centerX := range []float32{left + radius, left + width - radius} {
		for _, centerY := range []float32{top + radius, top + height - radius} {
			vector.FillCircle(screen, centerX, centerY, radius, fill, true)
		}
	}
}

func roundedImage(fill, edge color.NRGBA, radius int) *uiimage.NineSlice {
	key := surfaceKey{fill, edge, radius}
	if cached := surfaceImages[key]; cached != nil {
		return cached
	}
	size := radius*2 + 3
	texture := ebiten.NewImage(size, size)
	rasterizeRounded(texture, texture.Bounds(), float32(radius), edge)
	rasterizeRounded(texture, texture.Bounds().Inset(1), float32(max(0, radius-1)), fill)
	result := uiimage.NewNineSlice(texture, [3]int{radius + 1, 1, radius + 1}, [3]int{radius + 1, 1, radius + 1})
	surfaceImages[key] = result
	return result
}

func (view *Window) updateIntro() {
	if ebiten.IsWindowBeingClosed() {
		view.quit = true
		return
	}
	pressed := len(inpututil.AppendJustPressedKeys(nil)) > 0 || len(inpututil.AppendJustPressedTouchIDs(nil)) > 0
	for _, button := range []ebiten.MouseButton{ebiten.MouseButtonLeft, ebiten.MouseButtonMiddle, ebiten.MouseButtonRight} {
		pressed = pressed || inpututil.IsMouseButtonJustPressed(button)
	}
	if view.animationTicks > 10 && pressed {
		view.introLeaving = true
	}
	if view.introLeaving {
		view.introFade = min(1, view.introFade+.045)
		if view.introFade >= 1 {
			view.introVisible = false
			view.requestRebuild()
		}
	}
}

func (view *Window) centeredText(screen *ebiten.Image, caption string, center, top, size float64, tint color.Color) {
	width, _ := text.Measure(caption, view.fontFace(size), size*1.4)
	view.drawText(screen, caption, center-width/2, top, size, tint)
}

func (view *Window) fitText(caption string, width, size float64) string {
	for length := len([]rune(caption)); length > 1; length-- {
		candidate := shorten(caption, length)
		measured, _ := text.Measure(candidate, view.fontFace(size), size*1.4)
		if measured <= width {
			return candidate
		}
	}
	return "…"
}

func (view *Window) drawIntro(screen *ebiten.Image) {
	screen.Fill(color.NRGBA{3, 10, 26, 255})
	view.drawStars(screen)
	size := math.Min(float64(view.height)*.82, float64(view.width)*.68)
	options := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	options.GeoM.Scale(size/float64(view.logo.Bounds().Dx()), size/float64(view.logo.Bounds().Dy()))
	options.GeoM.Translate((float64(view.width)-size)/2, (float64(view.height)-size)/2-30)
	screen.DrawImage(view.logo, options)
	center := float64(view.width) / 2
	view.centeredText(screen, "A  P A R T I C L E  U N I V E R S E", center, float64(view.height)-132, 11, muted)
	pulse := .65 + .25*math.Sin(float64(view.animationTicks)/48)
	view.centeredText(screen, "press anything to continue...", center, float64(view.height)-91, 15, mix(muted, ink, pulse))
	view.centeredText(screen, "S M A L L  R U L E S .   I N F I N I T E  P O S S I B I L I T I E S .", center, float64(view.height)-35, 9, muted)
	if view.introFade > 0 {
		vector.FillRect(screen, 0, 0, float32(view.width), float32(view.height), color.NRGBA{R: canvas.R, G: canvas.G, B: canvas.B, A: uint8(255 * view.introFade)}, false)
	}
}

func (view *Window) drawStars(screen *ebiten.Image) {
	if view.star == nil {
		view.star = ebiten.NewImage(4, 4)
		vector.FillCircle(view.star, 2, 2, .6, color.NRGBA{73, 100, 142, 85}, true)
	}
	for index := 0; index < 80; index++ {
		positionX := float64((index*313 + 97) % max(1, view.width))
		positionY := float64((index*197 + 53) % max(1, view.height))
		options := ebiten.DrawImageOptions{}
		options.GeoM.Translate(positionX-2, positionY-2)
		screen.DrawImage(view.star, &options)
	}
}

func (view *Window) drawBrand(screen *ebiten.Image) {
	vector.StrokeCircle(screen, 35, 40, 12, 1, border, true)
	for index, tint := range []color.NRGBA{accent, pink, {255, 191, 55, 255}} {
		angle := float64(index)*math.Pi*2/3 - math.Pi/2
		vector.FillCircle(screen, 35+float32(math.Cos(angle)*12), 40+float32(math.Sin(angle)*12), 3, tint, true)
	}
	vector.FillCircle(screen, 35, 40, 3, ink, true)
}

func (view *Window) drawDrawerHandle(screen *ebiten.Image) {
	bounds := view.drawerHandleRect()
	// Extend past the seam and erase the shared edge, so the pull tab is part
	// of the panel silhouette rather than a button laid over its border.
	fillRounded(screen, bounds, 14, border)
	fillRounded(screen, bounds.Inset(1), 13, panelBG)
	vector.FillRect(screen, float32(bounds.Max.X-14), float32(bounds.Min.Y+1), 14, float32(bounds.Dy()-2), panelBG, false)
	if view.inspectorVisible {
		x := float32(view.inspectorRect().Min.X)
		vector.FillRect(screen, x-12, float32(bounds.Min.Y), 14, 1, border, false)
		vector.FillRect(screen, x-12, float32(bounds.Max.Y-1), 14, 1, border, false)
	}
	cursorX, cursorY := ebiten.CursorPosition()
	if image.Pt(cursorX, cursorY).In(bounds) || view.ui.GetFocusedWidget() == view.drawerHandle {
		fillRounded(screen, bounds.Inset(5), 10, mix(panelBG, accent, .12))
	}
	options := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	size := 36.0
	options.GeoM.Scale(size/float64(view.atomSticker.Bounds().Dx()), size/float64(view.atomSticker.Bounds().Dy()))
	options.GeoM.Translate(float64(bounds.Min.X)+(float64(bounds.Dx())-size)/2-1, float64(bounds.Min.Y)+(float64(bounds.Dy())-size)/2)
	screen.DrawImage(view.atomSticker, options)
}

func (view *Window) shortcutDialog() {
	view.openDialog(&dialog{
		title: "Stay in the flow",
		body:  "A few shortcuts, a whole universe. Available whenever you are not typing in a field.",
		build: func(parent *widget.Container, x, y, width, height int) {
			rows := [][2]string{
				{"Space / .", "Play or pause / advance one paused step"},
				{"1 / 2 / 3", "Rules / particles / settings drawer"},
				{"I / Escape", "Toggle inspector / leave field, then close drawer"},
				{"R / F11", "Restart particles / fullscreen"},
				{"Ctrl+S / Ctrl+Shift+S", "Save / save as"},
				{"Ctrl+O", "Open preset library"},
				{"Ctrl+Z / Ctrl+Shift+Z", "Undo / redo a configuration edit"},
				{"Tab / Shift+Tab", "Move forward / backward through controls"},
				{"Arrows / Enter", "Adjust slider or matrix / activate or commit"},
				{"Right-click / L", "Lock or unlock a matrix cell (L when matrix focused)"},
				{"? / F1", "Show this guide"},
				{"M", "Mute or unmute music and effects"},
			}
			for index, row := range rows {
				view.addLabel(parent, x, y+index*32, 202, row[0], 12, accent)
				view.addLabel(parent, x+212, y+index*32, width-212, row[1], 12, ink)
			}
		},
	})
}

func (view *Window) updateTooltip() {
	positionX, positionY := ebiten.CursorPosition()
	caption := ""
	for _, focusable := range view.rootContainer.GetFocusers() {
		button, ok := focusable.(*widget.Button)
		if !ok || button.Text() == nil || button.GetWidget().Disabled || !image.Pt(positionX, positionY).In(button.GetWidget().Rect) {
			continue
		}
		if button == view.drawerHandle {
			caption = "Open inspector · I"
			if view.inspectorVisible {
				caption = "Close inspector · I"
			}
			continue
		}
		switch button.Text().Label {
		case "Link pair: off", "Link pair: on":
			caption = "Edit both directions together · Locked cells stay unchanged"
		case "Zero cell", "Zero all":
			caption = "Clear attraction / repulsion; inner repulsion stays active"
		case "Shuffle relationships":
			caption = "Randomize unlocked relationships · Keep locked cells"
		case "Save":
			caption = "Save preset · Ctrl+S"
		case "Undo":
			caption = "Undo edit · Ctrl+Z"
		case "Redo":
			caption = "Redo edit · Ctrl+Shift+Z"
		case "Restart":
			caption = "Restart from this seed · R"
		case "Pause", "Play":
			caption = "Toggle playback · Space"
		case "Step":
			caption = "Advance one paused step · ."
		case "Fullscreen":
			caption = "Toggle fullscreen · F11"
		case "New world":
			caption = "Generate new rules, populations and positions"
		case "?":
			caption = "Keyboard guide · ? or F1"
		case "×":
			caption = "Cancel · Escape"
		}
	}
	if caption == view.tooltipText {
		view.tooltipTicks++
	} else {
		view.tooltipText, view.tooltipTicks = caption, 0
	}
}

func (view *Window) drawTooltip(screen *ebiten.Image) {
	if view.tooltipText == "" || view.tooltipTicks < 35 {
		return
	}
	positionX, positionY := ebiten.CursorPosition()
	width, _ := text.Measure(view.tooltipText, view.fontFace(12), 16)
	left := min(positionX+12, view.width-int(width)-32)
	top := positionY + 26
	if top+36 > view.height-20 {
		top = positionY - 42
	}
	fillRounded(screen, rectangle(left, top, int(width)+24, 34), 8, border)
	view.drawText(screen, view.tooltipText, float64(left+12), float64(top+8), 12, ink)
}
