package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type particleSprite struct {
	image  *ebiten.Image
	center float64
}

func makeParticleSprite(particleType ParticleType) particleSprite {
	center := math.Ceil(particleType.Mass+10) + 1
	sprite := ebiten.NewImage(int(center*2), int(center*2))
	red, green, blue, _ := particleType.Clr.RGBA()
	halo := color.NRGBA{R: uint8(red >> 8), G: uint8(green >> 8), B: uint8(blue >> 8), A: 20}
	vector.FillCircle(sprite, float32(center), float32(center), float32(particleType.Mass+10), halo, true)
	vector.FillCircle(sprite, float32(center), float32(center), float32(particleType.Mass), particleType.Clr, true)
	return particleSprite{image: sprite, center: center}
}

func (g *Game) Draw(screen *ebiten.Image) {
	options := ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	for _, particle := range g.Prts {
		sprite := g.sprites[particle.TypeID]
		options.GeoM.Reset()
		options.GeoM.Translate(particle.X-sprite.center, particle.Y-sprite.center)
		screen.DrawImage(sprite.image, &options)
	}
}
