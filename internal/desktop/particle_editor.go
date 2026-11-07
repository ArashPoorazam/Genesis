package desktop

import (
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"

	"github.com/ArashPoorazam/Genesis/internal/simulation"
	"github.com/ArashPoorazam/Genesis/internal/soundtrack"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func (view *Window) buildParticles(x, y, width, height int) {
	config := view.controller.Simulation.Config
	listHeight := ((len(config.ParticleTypes) + 1) / 2) * 42
	content := view.addScrollArea(view.inspectorContainer, x, y, width, height, listHeight+1050)
	width -= 24

	view.addLabel(content, 0, 0, width, "Particle types", 23, ink)
	view.addLabel(content, 0, 36, width, "Shape the inhabitants of your world.", 12, muted)
	view.buildTypeList(content, width, listHeight)

	detailsY := listHeight + 134
	typeIndex := view.selectedType
	view.buildTypeIdentity(content, width, detailsY, typeIndex)
	view.buildPopulationControl(content, width, detailsY+176, typeIndex)
	view.buildRadiusControls(content, width, detailsY+344, typeIndex)
	view.buildInfluenceDiagram(content, width, detailsY+728, typeIndex)
	view.buildRemoveTypeButton(content, width, detailsY+858, typeIndex)
}

func (view *Window) buildTypeList(content *widget.Container, width, listHeight int) {
	config := view.controller.Simulation.Config
	for typeIndex, particleType := range config.ParticleTypes {
		caption := shorten(particleType.Name, 12) + fmt.Sprintf(" · %d", particleType.Population)
		buttonX, buttonY := (typeIndex%2)*((width+8)/2), 68+(typeIndex/2)*42
		button := view.addButton(content, buttonX, buttonY, (width-8)/2, caption, view.selectedType == typeIndex, func() {
			view.selectedType = typeIndex
			view.requestRebuild()
		})
		images := buttonImages(view.selectedType == typeIndex)
		images.Idle = roundedImage(surface, mix(surface, particleType.Color, .55), 8)
		if view.selectedType == typeIndex {
			images.Idle = roundedImage(mix(surface, particleType.Color, .18), particleType.Color, 8)
		}
		button.SetImage(images)
		previousColor := particleType.Color
		view.refreshBindings = append(view.refreshBindings, func() {
			current := view.controller.Simulation.Config.ParticleTypes[typeIndex]
			button.SetText(shorten(current.Name, 12) + fmt.Sprintf(" · %d", current.Population))
			if current.Color != previousColor {
				previousColor = current.Color
				fill := surface
				if view.selectedType == typeIndex {
					fill = mix(surface, current.Color, .18)
				}
				images.Idle = roundedImage(fill, current.Color, 8)
			}
		})
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
	errLabel := view.addLabel(content, 0, detailsY+88, width, "", 11, danger)
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
	view.addButton(content, 0, detailsY+120, width, "Custom color…", false, func() {
		view.openColorPicker(typeIndex)
	})
}

func (view *Window) buildPopulationControl(content *widget.Container, width, detailsY, typeIndex int) {
	particleType := view.controller.Simulation.Config.ParticleTypes[typeIndex]

	view.addLabel(content, 0, detailsY, width, "Population", 14, ink)
	populationInput := view.addTextInput(content, 0, detailsY+28, 100, strconv.Itoa(particleType.Population), nil)
	popError := view.addLabel(content, 0, detailsY+70, width, "", 11, danger)
	view.addButton(content, 110, detailsY+28, width-110, "Update population", true, func() {
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
			view.sound.Queue(soundtrack.Population)
		}
	})
	view.addLiveLabel(content, 0, detailsY+96, width, 11, muted, func() string {
		count := 0
		for _, particle := range view.controller.Simulation.Particles {
			if particle.TypeIndex == typeIndex {
				count++
			}
		}
		return fmt.Sprintf("%d placed · 5,000 total limit. New particles use spawn settings.", count)
	})
}

func (view *Window) buildRadiusControls(content *widget.Container, width, detailsY, typeIndex int) {
	particleType := view.controller.Simulation.Config.ParticleTypes[typeIndex]

	view.addNumberControl(content, 0, detailsY, width, numericSetting{
		Title:   "Particle radius",
		Hint:    "Visual size and spawn clearance; not inertial mass.",
		Value:   particleType.ParticleRadius,
		Read:    func(config simulation.Config) float64 { return config.ParticleTypes[typeIndex].ParticleRadius },
		Minimum: .5,
		Maximum: 20,
		Step:    .1,
		Apply: func(config *simulation.Config, value float64) {
			setParticleRadius(config, typeIndex, physicalRadius, value)
		},
	})
	view.addNumberControl(content, 0, detailsY+128, width, numericSetting{
		Title:   "Inner influence radius",
		Hint:    "Linked to outer radius; always larger than the particle.",
		Value:   particleType.InnerRadius,
		Read:    func(config simulation.Config) float64 { return config.ParticleTypes[typeIndex].InnerRadius },
		Minimum: 1,
		Maximum: 499,
		Step:    .1,
		Apply: func(config *simulation.Config, value float64) {
			setParticleRadius(config, typeIndex, innerInfluenceRadius, value)
		},
	})
	view.addNumberControl(content, 0, detailsY+256, width, numericSetting{
		Title:   "Outer influence radius",
		Hint:    "Linked to inner radius; changes keep both zones valid.",
		Value:   particleType.OuterRadius,
		Read:    func(config simulation.Config) float64 { return config.ParticleTypes[typeIndex].OuterRadius },
		Minimum: 2,
		Maximum: 500,
		Step:    .1,
		Apply: func(config *simulation.Config, value float64) {
			setParticleRadius(config, typeIndex, outerInfluenceRadius, value)
		},
	})
}

func (view *Window) buildInfluenceDiagram(content *widget.Container, width, detailsY, typeIndex int) {
	diagram := ebiten.NewImage(width, 106)
	view.transientImages = append(view.transientImages, diagram)
	graphic := widget.NewGraphic(widget.GraphicOpts.Image(diagram))
	placeWidget(content, graphic, rectangle(0, detailsY, width, 106))
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
	remove := view.addButton(content, 0, detailsY, width, "Remove this type…", false, func() {
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
