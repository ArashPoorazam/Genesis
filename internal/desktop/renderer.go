package desktop

import (
	"image/color"
	"math"

	"github.com/ArashPoorazam/Genesis/internal/simulation"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type particleSprite struct {
	image  *ebiten.Image
	center float64
	radius float64
	color  color.NRGBA
	glow   bool
}

func makeParticleSprite(particleType simulation.ParticleType, glow bool) particleSprite {
	center := math.Ceil(particleType.ParticleRadius+10) + 1
	spriteImage := ebiten.NewImage(int(center*2), int(center*2))

	if glow {
		haloColor := particleType.Color
		for layer := 10; layer > 0; layer-- {
			haloColor.A = uint8(2 + (10-layer)*2)
			vector.FillCircle(spriteImage, float32(center), float32(center), float32(particleType.ParticleRadius+float64(layer)), haloColor, true)
		}
	}
	vector.FillCircle(spriteImage, float32(center), float32(center), float32(particleType.ParticleRadius), particleType.Color, true)

	return particleSprite{
		image:  spriteImage,
		center: center,
		radius: particleType.ParticleRadius,
		color:  particleType.Color,
		glow:   glow,
	}
}

func (view *Window) drawText(destination *ebiten.Image, caption string, x, y, size float64, textColor color.Color) {
	options := &text.DrawOptions{}
	options.GeoM.Translate(x, y)
	options.ColorScale.ScaleWithColor(textColor)
	text.Draw(destination, caption, view.fontFace(size), options)
}

func (view *Window) Draw(screen *ebiten.Image) {
	screen.Fill(canvas)
	if view.introVisible {
		view.drawIntro(screen)
		view.captureFrame(screen)
		return
	}
	if view.rootContainer == nil {
		return
	}

	view.drawStars(screen)
	view.drawWorld(screen)
	view.drawStatus(screen)
	view.drawPanelBackgrounds(screen)
	view.ui.Draw(screen)
	view.drawTooltip(screen)

	if focused := view.ui.GetFocusedWidget(); focused != nil {
		bounds := focused.GetWidget().Rect
		vector.StrokeRect(screen, float32(bounds.Min.X), float32(bounds.Min.Y), float32(bounds.Dx()), float32(bounds.Dy()), 1, accent, false)
	}

	if view.introFade > 0 {
		vector.FillRect(screen, 0, 0, float32(view.width), float32(view.height), color.NRGBA{R: canvas.R, G: canvas.G, B: canvas.B, A: uint8(255 * view.introFade)}, false)
	}
	view.captureFrame(screen)
}

func (view *Window) drawPanelBackgrounds(screen *ebiten.Image) {
	fillRounded(screen, rectangle(12, 12, view.width-24, 56), 10, panelBG)
	view.drawBrand(screen)

	if view.inspectorVisible {
		bounds := view.inspectorRect()
		fillRounded(screen, bounds, 10, border)
		fillRounded(screen, bounds.Inset(1), 9, panelBG)
		if view.activeDialog == nil {
			view.drawMatrix(screen)
		}
	}
	if view.activeDialog == nil {
		view.drawDrawerHandle(screen)
	}

	if view.activeDialog != nil {
		view.drawDialogBackground(screen)
	}
}

func (view *Window) drawWorld(screen *ebiten.Image) {
	if view.viewport.Empty() {
		return
	}

	config := view.controller.Simulation.Config
	scaleX := float64(view.viewport.Dx()) / config.WorldWidth
	scaleY := float64(view.viewport.Dy()) / config.WorldHeight
	scale := min(scaleX, scaleY)
	originX, originY := float64(view.viewport.Min.X), float64(view.viewport.Min.Y)
	destination := screen.SubImage(view.viewport).(*ebiten.Image)

	sprites := view.currentSprites()
	options := ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	for _, particle := range view.controller.Simulation.Particles {
		sprite := sprites[particle.TypeIndex]
		options.GeoM.Reset()
		options.GeoM.Translate(-sprite.center, -sprite.center)
		options.GeoM.Scale(scale, scale)
		options.GeoM.Translate(originX+particle.X*scaleX, originY+particle.Y*scaleY)
		destination.DrawImage(sprite.image, &options)
	}

	if view.influencePreviewEnabled {
		view.drawInfluencePreview(destination, originX, originY, scaleX, scaleY)
	}
	if len(view.controller.Simulation.Particles) == 0 {
		view.drawText(screen, "A world waiting to happen", originX+24, originY+28, 22, ink)
		view.drawText(screen, "Add a population in the Particles tab, or create a new world.", originX+24, originY+66, 13, muted)
	}
}

func (view *Window) currentSprites() []particleSprite {
	particleTypes := view.controller.Simulation.Config.ParticleTypes
	sprites := make([]particleSprite, len(particleTypes))
	activeIDs := make(map[string]bool, len(particleTypes))

	for index, particleType := range particleTypes {
		activeIDs[particleType.ID] = true
		sprite, exists := view.sprites[particleType.ID]
		appearanceChanged := sprite.radius != particleType.ParticleRadius || sprite.color != particleType.Color || sprite.glow != view.glowEnabled
		if !exists || appearanceChanged {
			if exists {
				sprite.image.Deallocate()
			}
			sprite = makeParticleSprite(particleType, view.glowEnabled)
			view.sprites[particleType.ID] = sprite
		}
		sprites[index] = sprite
	}

	for id, sprite := range view.sprites {
		if !activeIDs[id] {
			sprite.image.Deallocate()
			delete(view.sprites, id)
		}
	}

	return sprites
}

func (view *Window) drawInfluencePreview(screen *ebiten.Image, originX, originY, scaleX, scaleY float64) {
	for _, particle := range view.controller.Simulation.Particles {
		if particle.TypeIndex != view.selectedType {
			continue
		}

		particleType := view.controller.Simulation.Config.ParticleTypes[particle.TypeIndex]
		x, y := float32(originX+particle.X*scaleX), float32(originY+particle.Y*scaleY)
		for index, radius := range []float64{particleType.OuterRadius, particleType.InnerRadius} {
			var path vector.Path
			for segment := 0; segment <= 96; segment++ {
				angle := float64(segment) * 2 * math.Pi / 96
				pointX := x + float32(math.Cos(angle)*radius*scaleX)
				pointY := y + float32(math.Sin(angle)*radius*scaleY)
				if segment == 0 {
					path.MoveTo(pointX, pointY)
				} else {
					path.LineTo(pointX, pointY)
				}
			}
			path.Close()
			tint := muted
			if index == 1 {
				tint = danger
			}
			options := &vector.DrawPathOptions{AntiAlias: true}
			options.ColorScale.ScaleWithColor(tint)
			vector.StrokePath(screen, &path, &vector.StrokeOptions{Width: 1}, options)
		}
		return
	}
}
