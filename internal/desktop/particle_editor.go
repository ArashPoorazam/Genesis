package desktop

import (
	"fmt"
	"image/color"
	"math/rand/v2"
	"strconv"
	"strings"

	"github.com/ArashPoorazam/Genesis/internal/simulation"
	uiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func (view *Window) buildParticles(x, y, width, height int) {
	config := view.controller.Simulation.Config
	listHeight := ((len(config.ParticleTypes) + 1) / 2) * 42
	content := view.addScrollArea(view.rootContainer, x, y, width, height, listHeight+980)
	width -= 24

	view.addLabel(content, 0, 0, width, "Particle types", 23, ink)
	view.addLabel(content, 0, 36, width, "Shape the inhabitants of your world.", 12, muted)
	view.buildTypeList(content, width, listHeight)

	detailsY := listHeight + 134
	typeIndex := view.selectedType
	view.buildTypeIdentity(content, width, detailsY, typeIndex)
	view.buildPopulationControl(content, width, detailsY, typeIndex)
	view.buildRadiusControls(content, width, detailsY, typeIndex)
	view.buildInfluenceDiagram(content, width, detailsY, typeIndex)
	view.buildRemoveTypeButton(content, width, detailsY, typeIndex)
}

func (view *Window) buildTypeList(content *widget.Container, width, listHeight int) {
	config := view.controller.Simulation.Config
	for typeIndex, particleType := range config.ParticleTypes {
		caption := shorten(particleType.Name, 12) + fmt.Sprintf(" · %d", particleType.Population)
		buttonX, buttonY := (typeIndex%2)*(width/2), 68+(typeIndex/2)*42
		button := view.addButton(content, buttonX, buttonY, width/2-6, caption, view.selectedType == typeIndex, func() {
			view.selectedType = typeIndex
			view.requestRebuild()
		})
		images := buttonImages(view.selectedType == typeIndex)
		images.Idle = uiimage.NewBorderedNineSliceColor(surface, particleType.Color, 1)
		if view.selectedType == typeIndex {
			images.Idle = uiimage.NewBorderedNineSliceColor(mix(surface, particleType.Color, .18), particleType.Color, 2)
		}
		button.SetImage(images)
	}
	add := view.addButton(content, 0, listHeight+76, width, "+ Add particle type", false, func() {
		view.editConfig(func(config *simulation.Config) error {
			return config.AddType(fmt.Sprintf("type-%x", rand.Uint64()))
		})
		view.selectedType = len(view.controller.Simulation.Config.ParticleTypes) - 1
		view.requestRebuild()
	})
	add.GetWidget().Disabled = len(config.ParticleTypes) >= simulation.MaxTypes

}

func (view *Window) buildTypeIdentity(content *widget.Container, width, detailsY, typeIndex int) {
	particleType := view.controller.Simulation.Config.ParticleTypes[typeIndex]

	view.addLabel(content, 0, detailsY, width, "TYPE DETAILS", 10, muted)
	view.addLabel(content, 0, detailsY+27, width, "Name", 13, ink)
	errLabel := view.addLabel(content, 0, detailsY+114, width, "", 11, danger)
	view.addTextInput(content, 0, detailsY+49, width, particleType.Name, func(caption string) error {
		err := view.editConfig(func(config *simulation.Config) error {
			config.ParticleTypes[typeIndex].Name = strings.TrimSpace(caption)
			return nil
		})
		if err != nil {
			errLabel.Label = err.Error()
		} else {
			errLabel.Label = ""
		}
		return err
	})
	view.addLabel(content, 0, detailsY+96, 80, "Color", 13, ink)
	view.addTextInput(content, 96, detailsY+88, width-96, fmt.Sprintf("#%02X%02X%02X", particleType.Color.R, particleType.Color.G, particleType.Color.B), func(caption string) error {
		caption = strings.TrimPrefix(strings.TrimSpace(caption), "#")
		if len(caption) != 6 {
			errLabel.Label = "Use a six-digit color, e.g. #FF4655"
			return fmt.Errorf("use a six-digit hex color")
		}
		value, err := strconv.ParseUint(caption, 16, 32)
		if err != nil {
			errLabel.Label = "Use a six-digit hex color"
			return err
		}
		errLabel.Label = ""
		return view.editConfig(func(config *simulation.Config) error {
			config.ParticleTypes[typeIndex].Color = color.NRGBA{uint8(value >> 16), uint8(value >> 8), uint8(value), 255}
			return nil
		})
	})
	palette := []color.NRGBA{{255, 70, 85, 255}, {109, 222, 145, 255}, {93, 153, 255, 255}, {244, 133, 190, 255}, {255, 201, 121, 255}, {181, 161, 242, 255}}
	for paletteIndex, swatchColor := range palette {
		button := view.addButton(content, paletteIndex*(width/6), detailsY+144, width/6-6, "", false, func() {
			view.editConfig(func(config *simulation.Config) error {
				config.ParticleTypes[typeIndex].Color = swatchColor
				return nil
			})
			view.requestRebuild()
		})
		images := buttonImages(false)
		images.Idle = uiimage.NewNineSliceColor(swatchColor)
		button.SetImage(images)
	}

}

func (view *Window) buildPopulationControl(content *widget.Container, width, detailsY, typeIndex int) {
	particleType := view.controller.Simulation.Config.ParticleTypes[typeIndex]

	view.addLabel(content, 0, detailsY+198, width, "Population", 14, ink)
	populationInput := view.addTextInput(content, 0, detailsY+225, 100, strconv.Itoa(particleType.Population), nil)
	popError := view.addLabel(content, 0, detailsY+266, width, "", 11, danger)
	view.addButton(content, 110, detailsY+225, width-110, "Update population", true, func() {
		population, err := strconv.Atoi(populationInput.GetText())
		if err == nil {
			err = view.editConfig(func(config *simulation.Config) error {
				config.ParticleTypes[typeIndex].Population = population
				return nil
			})
		}
		if err != nil {
			popError.Label = err.Error()
			view.reportError(err)
		} else {
			popError.Label = ""
			view.status = "Population updated; surviving particles retained"
		}
	})
	view.addLiveLabel(content, 0, detailsY+288, width, 11, muted, func() string {
		count := 0
		for _, particle := range view.controller.Simulation.Particles {
			if particle.TypeIndex == typeIndex {
				count++
			}
		}
		return fmt.Sprintf("%d placed · 5,000 total limit · new particles use spawn settings", count)
	})

}

func (view *Window) buildRadiusControls(content *widget.Container, width, detailsY, typeIndex int) {
	particleType := view.controller.Simulation.Config.ParticleTypes[typeIndex]

	view.addNumberControl(content, 0, detailsY+330, width, numericSetting{
		Title:   "Particle radius",
		Hint:    "Visual size and spawn clearance; not inertial mass.",
		Value:   particleType.ParticleRadius,
		Minimum: .5,
		Maximum: 20,
		Step:    .1,
		Apply: func(config *simulation.Config, value float64) {
			config.ParticleTypes[typeIndex].ParticleRadius = value
		},
	})
	view.addNumberControl(content, 0, detailsY+442, width, numericSetting{
		Title:   "Inner influence radius",
		Hint:    "Neighbors inside this zone are always repelled.",
		Value:   particleType.InnerRadius,
		Minimum: 1,
		Maximum: 499,
		Step:    .1,
		Apply: func(config *simulation.Config, value float64) {
			config.ParticleTypes[typeIndex].InnerRadius = value
		},
	})
	view.addNumberControl(content, 0, detailsY+554, width, numericSetting{
		Title:   "Outer influence radius",
		Hint:    "This type exerts no force beyond this distance.",
		Value:   particleType.OuterRadius,
		Minimum: 2,
		Maximum: 500,
		Step:    .1,
		Apply: func(config *simulation.Config, value float64) {
			config.ParticleTypes[typeIndex].OuterRadius = value
		},
	})

}

func (view *Window) buildInfluenceDiagram(content *widget.Container, width, detailsY, typeIndex int) {
	diagram := ebiten.NewImage(width, 106)
	graphic := widget.NewGraphic(widget.GraphicOpts.Image(diagram))
	placeWidget(content, graphic, rectangle(0, detailsY+669, width, 106))
	var prior simulation.ParticleType
	view.refreshBindings = append(view.refreshBindings, func() {
		particleType := view.controller.Simulation.Config.ParticleTypes[typeIndex]
		if particleType == prior {
			return
		}
		prior = particleType
		diagram.Fill(canvas)
		vector.StrokeCircle(diagram, 55, 53, 42, 1, muted, true)
		vector.StrokeCircle(diagram, 55, 53, float32(particleType.InnerRadius/particleType.OuterRadius*42), 1, danger, true)
		vector.FillCircle(diagram, 55, 53, float32(max(2, particleType.ParticleRadius/particleType.OuterRadius*42)), particleType.Color, true)
		view.drawText(diagram, "Outer · relationship force", 113, 23, 12, muted)
		view.drawText(diagram, "Inner · repulsion", 113, 46, 12, danger)
		view.drawText(diagram, "Center · particle radius", 113, 69, 12, particleType.Color)
	})

}

func (view *Window) buildRemoveTypeButton(content *widget.Container, width, detailsY, typeIndex int) {
	remove := view.addButton(content, 0, detailsY+800, width, "Remove this type…", false, func() {
		count := 0
		for _, particle := range view.controller.Simulation.Particles {
			if particle.TypeIndex == typeIndex {
				count++
			}
		}
		title := "Remove " + view.controller.Simulation.Config.ParticleTypes[typeIndex].Name + "?"
		message := fmt.Sprintf("This removes %d particles and its matrix row and column. Configuration can be undone; elapsed motion cannot.", count)
		view.confirm(title, message, "Remove type", func() {
			view.editConfig(func(config *simulation.Config) error {
				return config.RemoveType(typeIndex)
			})
			view.selectedType = 0
			view.requestRebuild()
		})
	})
	remove.GetWidget().Disabled = len(view.controller.Simulation.Config.ParticleTypes) == 1

}
