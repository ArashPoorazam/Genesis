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
		haloColor.A = 18
		vector.FillCircle(spriteImage, float32(center), float32(center), float32(particleType.ParticleRadius+10), haloColor, true)

		haloColor.A = 28
		vector.FillCircle(spriteImage, float32(center), float32(center), float32(particleType.ParticleRadius+4), haloColor, true)
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
	if view.rootContainer == nil {
		return
	}

	view.drawWorld(screen)
	view.drawPanelBackgrounds(screen)
	view.ui.Draw(screen)

	if focused := view.ui.GetFocusedWidget(); focused != nil {
		bounds := focused.GetWidget().Rect
		vector.StrokeRect(screen, float32(bounds.Min.X), float32(bounds.Min.Y), float32(bounds.Dx()), float32(bounds.Dy()), 1, accent, false)
	}

	view.captureFrame(screen)
}

func (view *Window) drawPanelBackgrounds(screen *ebiten.Image) {
	vector.FillRect(screen, 0, 0, float32(view.width), 64, panelBG, false)
	vector.StrokeLine(screen, 0, 63, float32(view.width), 63, 1, border, false)
	vector.FillRect(screen, 0, float32(view.height-32), float32(view.width), 32, panelBG, false)
	vector.StrokeLine(screen, 0, float32(view.height-32), float32(view.width), float32(view.height-32), 1, border, false)

	if view.inspectorVisible {
		bounds := view.inspectorRect()
		vector.FillRect(screen, float32(bounds.Min.X), float32(bounds.Min.Y), float32(bounds.Dx()), float32(bounds.Dy()), panelBG, false)
		vector.StrokeRect(screen, float32(bounds.Min.X), float32(bounds.Min.Y), float32(bounds.Dx()), float32(bounds.Dy()), 1, border, false)
		if view.activeDialog == nil {
			view.drawMatrix(screen)
		}
	}

	if view.activeDialog != nil {
		vector.FillRect(screen, 0, 0, float32(view.width), float32(view.height), color.NRGBA{A: 180}, false)
		bounds := view.dialogRect()
		vector.FillRect(screen, float32(bounds.Min.X), float32(bounds.Min.Y), float32(bounds.Dx()), float32(bounds.Dy()), panelBG, false)
		vector.StrokeRect(screen, float32(bounds.Min.X), float32(bounds.Min.Y), float32(bounds.Dx()), float32(bounds.Dy()), 1, border, false)
	}
}

func (view *Window) drawWorld(screen *ebiten.Image) {
	if view.viewport.Empty() {
		return
	}

	config := view.controller.Simulation.Config
	scale := math.Min(float64(view.viewport.Dx())/config.WorldWidth, float64(view.viewport.Dy())/config.WorldHeight)
	originX := float64(view.viewport.Min.X) + (float64(view.viewport.Dx())-config.WorldWidth*scale)/2
	originY := float64(view.viewport.Min.Y) + (float64(view.viewport.Dy())-config.WorldHeight*scale)/2

	worldBorder := color.NRGBA{R: 27, G: 35, B: 42, A: 255}
	vector.StrokeRect(screen, float32(originX), float32(originY), float32(config.WorldWidth*scale), float32(config.WorldHeight*scale), 1, worldBorder, false)

	for _, x := range []float64{originX, originX + config.WorldWidth*scale - 12} {
		for _, y := range []float64{originY, originY + config.WorldHeight*scale - 12} {
			vector.StrokeRect(screen, float32(x), float32(y), 12, 12, 1, border, false)
		}
	}

	sprites := view.currentSprites()
	options := ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	for _, particle := range view.controller.Simulation.Particles {
		sprite := sprites[particle.TypeIndex]
		options.GeoM.Reset()
		options.GeoM.Translate(-sprite.center, -sprite.center)
		options.GeoM.Scale(scale, scale)
		options.GeoM.Translate(originX+particle.X*scale, originY+particle.Y*scale)
		screen.DrawImage(sprite.image, &options)
	}

	if view.influencePreviewEnabled {
		view.drawInfluencePreview(screen, originX, originY, scale)
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

func (view *Window) drawInfluencePreview(screen *ebiten.Image, originX, originY, scale float64) {
	for _, particle := range view.controller.Simulation.Particles {
		if particle.TypeIndex != view.selectedType {
			continue
		}

		particleType := view.controller.Simulation.Config.ParticleTypes[particle.TypeIndex]
		x, y := float32(originX+particle.X*scale), float32(originY+particle.Y*scale)
		vector.StrokeCircle(screen, x, y, float32(particleType.OuterRadius*scale), 1, muted, true)
		vector.StrokeCircle(screen, x, y, float32(particleType.InnerRadius*scale), 1, danger, true)
		return
	}
}
