package desktop

import (
	"fmt"
	"strconv"

	"github.com/ArashPoorazam/Genesis/internal/simulation"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

func (view *Window) toggle(content *widget.Container, x, y, width int, title string, value bool, set func(bool)) {
	caption := "Off"
	if value {
		caption = "On"
	}

	view.addLabel(content, x, y+7, width-88, title, 14, ink)
	view.addButton(content, x+width-80, y, 80, caption, value, func() {
		set(!value)
		view.requestRebuild()
	})
}

func (view *Window) buildSettings(x, y, width, height int) {
	content := view.addScrollArea(view.rootContainer, x, y, width, height, 1830)
	width -= 24

	view.addLabel(content, 0, 0, width, "World settings", 23, ink)
	view.buildMotionSettings(content, width)
	view.buildGenerationSettings(content, width)
	view.buildRandomizationSettings(content, width)
	view.buildAppearanceSettings(content, width)
}

func (view *Window) buildMotionSettings(content *widget.Container, width int) {
	config := view.controller.Simulation.Config

	view.addLabel(content, 0, 37, width, "MOTION / APPLIED LIVE", 10, muted)
	view.addNumberControl(content, 0, 70, width, numericSetting{
		Title:   "Velocity retention",
		Hint:    "Fraction retained per step. Lower values slow motion.",
		Value:   config.Motion.VelocityRetention,
		Minimum: 0,
		Maximum: 1,
		Step:    .01,
		Apply: func(config *simulation.Config, value float64) {
			config.Motion.VelocityRetention = value
		},
	})
	view.addNumberControl(content, 0, 182, width, numericSetting{
		Title:   "Speed limit",
		Hint:    "Maximum distance traveled per physics step.",
		Value:   config.Motion.SpeedLimit,
		Minimum: 0,
		Maximum: 50,
		Step:    .1,
		Apply: func(config *simulation.Config, value float64) {
			config.Motion.SpeedLimit = value
		},
	})
	view.addNumberControl(content, 0, 294, width, numericSetting{
		Title:   "Inner-zone repulsion",
		Hint:    "Strength of close-range separation.",
		Value:   config.Motion.RepulsionStrength,
		Minimum: 0,
		Maximum: 50,
		Step:    .1,
		Apply: func(config *simulation.Config, value float64) {
			config.Motion.RepulsionStrength = value
		},
	})

}

func (view *Window) buildGenerationSettings(content *widget.Container, width int) {
	config := view.controller.Simulation.Config

	view.addLabel(content, 0, 418, width, "GENERATION / NEXT SPAWN OR NEW WORLD", 10, muted)
	view.addLabel(content, 0, 449, width, "Seed", 14, ink)
	seedErr := view.addLabel(content, 0, 512, width, "", 11, danger)
	view.addTextInput(content, 0, 475, width, strconv.FormatUint(config.Generation.Seed, 10), func(caption string) error {
		seed, err := strconv.ParseUint(caption, 10, 64)
		if err != nil {
			seedErr.Label = "Use a nonnegative whole number (up to 64 bits)"
			return fmt.Errorf("invalid seed")
		}
		seedErr.Label = ""
		return view.editConfig(func(config *simulation.Config) error {
			config.Generation.Seed = seed
			return nil
		})
	})
	view.addNumberControl(content, 0, 548, width, numericSetting{
		Title:   "Random population ceiling",
		Hint:    "New world chooses 0 up to this ceiling, exclusive.",
		Value:   float64(config.Generation.PopulationCeiling),
		Minimum: 1,
		Maximum: 5000,
		Step:    1,
		Apply: func(config *simulation.Config, value float64) {
			config.Generation.PopulationCeiling = int(value)
		},
	})
	view.addNumberControl(content, 0, 660, width, numericSetting{
		Title:   "Initial speed",
		Hint:    "Random speed between zero and this maximum.",
		Value:   config.Generation.InitialSpeed,
		Minimum: 0,
		Maximum: 50,
		Step:    .1,
		Apply: func(config *simulation.Config, value float64) {
			config.Generation.InitialSpeed = value
		},
	})
	view.addNumberControl(content, 0, 772, width, numericSetting{
		Title:   "Spawn spacing",
		Hint:    "Extra clearance between particle surfaces.",
		Value:   config.Generation.SpawnSpacing,
		Minimum: 0,
		Maximum: 50,
		Step:    .1,
		Apply: func(config *simulation.Config, value float64) {
			config.Generation.SpawnSpacing = value
		},
	})

}

func (view *Window) buildRandomizationSettings(content *widget.Container, width int) {
	config := view.controller.Simulation.Config

	view.addLabel(content, 0, 898, width, "RANDOMIZATION / NEXT RANDOMIZE ACTION", 10, muted)
	view.addNumberControl(content, 0, 930, width, numericSetting{
		Title:   "Relationship amplitude",
		Hint:    "Random relationship range: −amplitude to +amplitude.",
		Value:   config.Generation.RelationshipAmplitude,
		Minimum: 0,
		Maximum: 10,
		Step:    .1,
		Apply: func(config *simulation.Config, value float64) {
			config.Generation.RelationshipAmplitude = value
		},
	})
	view.addNumberControl(content, 0, 1042, width, numericSetting{
		Title:   "Minimum inner radius",
		Hint:    "",
		Value:   config.Generation.MinInnerRadius,
		Minimum: 1,
		Maximum: 499,
		Step:    .1,
		Apply: func(config *simulation.Config, value float64) {
			config.Generation.MinInnerRadius = value
		},
	})
	view.addNumberControl(content, 0, 1130, width, numericSetting{
		Title:   "Maximum inner radius",
		Hint:    "",
		Value:   config.Generation.MaxInnerRadius,
		Minimum: 1,
		Maximum: 499,
		Step:    .1,
		Apply: func(config *simulation.Config, value float64) {
			config.Generation.MaxInnerRadius = value
		},
	})
	view.addNumberControl(content, 0, 1218, width, numericSetting{
		Title:   "Minimum outer radius",
		Hint:    "",
		Value:   config.Generation.MinOuterRadius,
		Minimum: 2,
		Maximum: 500,
		Step:    .1,
		Apply: func(config *simulation.Config, value float64) {
			config.Generation.MinOuterRadius = value
		},
	})
	view.addNumberControl(content, 0, 1306, width, numericSetting{
		Title:   "Maximum outer radius",
		Hint:    "",
		Value:   config.Generation.MaxOuterRadius,
		Minimum: 2,
		Maximum: 500,
		Step:    .1,
		Apply: func(config *simulation.Config, value float64) {
			config.Generation.MaxOuterRadius = value
		},
	})

}

func (view *Window) buildAppearanceSettings(content *widget.Container, width int) {
	view.addLabel(content, 0, 1412, width, "APPEARANCE & WORKSPACE", 10, muted)
	view.toggle(content, 0, 1444, width, "Particle glow", view.glowEnabled, func(value bool) {
		view.glowEnabled = value
	})
	view.toggle(content, 0, 1490, width, "Selected-type influence preview", view.influencePreviewEnabled, func(value bool) {
		view.influencePreviewEnabled = value
	})
	view.toggle(content, 0, 1536, width, "Performance display", view.showPerformance, func(value bool) {
		view.showPerformance = value
	})
	view.toggle(content, 0, 1582, width, "Fullscreen", ebiten.IsFullscreen(), func(value bool) {
		ebiten.SetFullscreen(value)
	})
	view.addLabel(content, 0, 1637, width, "Interface scale", 14, ink)
	for index, value := range []float64{1, 1.25, 1.5} {
		view.addButton(content, index*(width/3), 1666, width/3-6, fmt.Sprintf("%.0f%%", value*100), view.scale == value, func() {
			view.scale = value
			view.requestRebuild()
		})
	}

	view.addLabel(content, 0, 1714, width, "Scale adapts to keep controls usable in small windows.\nResizing does not change world size or physics.", 11, muted)
	view.addButton(content, 0, 1766, width, "Hide inspector", false, func() {
		view.inspectorVisible = false
		view.requestRebuild()
	})

}
