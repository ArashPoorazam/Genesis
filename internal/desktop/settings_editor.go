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
	view.addLabel(view.inspectorContainer, x, y, width, "Laboratory settings", 23, ink)
	for index, title := range []string{"Motion", "Spawn", "Ranges", "Display", "Sound"} {
		view.addButton(view.inspectorContainer, x+index*(width/5), y+40, width/5-5, title, view.settingsSection == index, func() {
			view.settingsSection = index
			view.requestRebuild()
		})
	}
	heights := []int{470, 480, 518, 410, 420}
	content := view.addScrollArea(view.inspectorContainer, x, y+92, width, height-92, heights[view.settingsSection])
	width -= 24
	switch view.settingsSection {
	case 0:
		view.buildMotionSettings(content, width)
	case 1:
		view.buildGenerationSettings(content, width)
	case 2:
		view.buildRandomizationSettings(content, width)
	case 3:
		view.buildAppearanceSettings(content, width)
	case 4:
		view.buildSoundSettings(content, width)
	}
}

func (view *Window) buildMotionSettings(content *widget.Container, width int) {
	config := view.controller.Simulation.Config

	view.addLabel(content, 0, 0, width, "MOTION / APPLIED LIVE", 10, muted)
	view.addNumberControl(content, 0, 33, width, numericSetting{
		Title:   "Velocity retention",
		Hint:    "Fraction retained per step. Lower values slow motion.",
		Value:   config.Motion.VelocityRetention,
		Read:    func(config simulation.Config) float64 { return config.Motion.VelocityRetention },
		Minimum: 0,
		Maximum: 1,
		Step:    .01,
		Apply: func(config *simulation.Config, value float64) {
			config.Motion.VelocityRetention = value
		},
	})
	view.addNumberControl(content, 0, 145, width, numericSetting{
		Title:   "Speed limit",
		Hint:    "Maximum distance traveled per physics step.",
		Value:   config.Motion.SpeedLimit,
		Read:    func(config simulation.Config) float64 { return config.Motion.SpeedLimit },
		Minimum: 0,
		Maximum: 50,
		Step:    .1,
		Apply: func(config *simulation.Config, value float64) {
			config.Motion.SpeedLimit = value
		},
	})
	view.addNumberControl(content, 0, 257, width, numericSetting{
		Title:   "Inner-zone repulsion",
		Hint:    "Strength of close-range separation.",
		Value:   config.Motion.RepulsionStrength,
		Read:    func(config simulation.Config) float64 { return config.Motion.RepulsionStrength },
		Minimum: 0,
		Maximum: 50,
		Step:    .1,
		Apply: func(config *simulation.Config, value float64) {
			config.Motion.RepulsionStrength = value
		},
	})
	view.toggle(content, 0, 382, width, "GPU boost (experimental)", view.controller.Accelerator != nil, view.setGPUBoost)
	view.addLabel(content, 0, 428, width, "May help large or dense worlds. Lower precision\ncan change trajectories. Falls back to CPU on errors.", 11, muted)
}

func (view *Window) buildGenerationSettings(content *widget.Container, width int) {
	config := view.controller.Simulation.Config

	view.addLabel(content, 0, 0, width, "GENERATION / NEXT SPAWN OR NEW WORLD", 10, muted)
	view.addLabel(content, 0, 31, width, "Seed", 14, ink)
	seedErr := view.addLabel(content, 0, 94, width, "", 11, danger)
	view.addTextInput(content, 0, 57, width, strconv.FormatUint(config.Generation.Seed, 10), func(caption string) error {
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
	view.addNumberControl(content, 0, 130, width, numericSetting{
		Title:   "Random population ceiling",
		Hint:    "New world chooses 0 up to this ceiling, exclusive.",
		Value:   float64(config.Generation.PopulationCeiling),
		Read:    func(config simulation.Config) float64 { return float64(config.Generation.PopulationCeiling) },
		Minimum: 1,
		Maximum: 5000,
		Step:    1,
		Apply: func(config *simulation.Config, value float64) {
			config.Generation.PopulationCeiling = int(value)
		},
	})
	view.addNumberControl(content, 0, 242, width, numericSetting{
		Title:   "Initial speed",
		Hint:    "Random speed between zero and this maximum.",
		Value:   config.Generation.InitialSpeed,
		Read:    func(config simulation.Config) float64 { return config.Generation.InitialSpeed },
		Minimum: 0,
		Maximum: 50,
		Step:    .1,
		Apply: func(config *simulation.Config, value float64) {
			config.Generation.InitialSpeed = value
		},
	})
	view.addNumberControl(content, 0, 354, width, numericSetting{
		Title:   "Spawn spacing",
		Hint:    "Extra clearance between particle surfaces.",
		Value:   config.Generation.SpawnSpacing,
		Read:    func(config simulation.Config) float64 { return config.Generation.SpawnSpacing },
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

	view.addLabel(content, 0, 0, width, "LINKED RANGES / NEIGHBORS ADJUST TOGETHER", 10, muted)
	view.addNumberControl(content, 0, 32, width, numericSetting{
		Title:   "Relationship amplitude",
		Hint:    "Random relationship range: −amplitude to +amplitude.",
		Value:   config.Generation.RelationshipAmplitude,
		Read:    func(config simulation.Config) float64 { return config.Generation.RelationshipAmplitude },
		Minimum: 0,
		Maximum: 10,
		Step:    .1,
		Apply: func(config *simulation.Config, value float64) {
			config.Generation.RelationshipAmplitude = value
		},
	})
	view.addNumberControl(content, 0, 144, width, numericSetting{
		Title:   "Minimum inner radius",
		Hint:    "",
		Value:   config.Generation.MinInnerRadius,
		Read:    func(config simulation.Config) float64 { return config.Generation.MinInnerRadius },
		Minimum: 1,
		Maximum: 499,
		Step:    .1,
		Apply: func(config *simulation.Config, value float64) {
			setGenerationRadius(config, minimumInnerRadius, value)
		},
	})
	view.addNumberControl(content, 0, 232, width, numericSetting{
		Title:   "Maximum inner radius",
		Hint:    "",
		Value:   config.Generation.MaxInnerRadius,
		Read:    func(config simulation.Config) float64 { return config.Generation.MaxInnerRadius },
		Minimum: 1,
		Maximum: 499,
		Step:    .1,
		Apply: func(config *simulation.Config, value float64) {
			setGenerationRadius(config, maximumInnerRadius, value)
		},
	})
	view.addNumberControl(content, 0, 320, width, numericSetting{
		Title:   "Minimum outer radius",
		Hint:    "",
		Value:   config.Generation.MinOuterRadius,
		Read:    func(config simulation.Config) float64 { return config.Generation.MinOuterRadius },
		Minimum: 2,
		Maximum: 500,
		Step:    .1,
		Apply: func(config *simulation.Config, value float64) {
			setGenerationRadius(config, minimumOuterRadius, value)
		},
	})
	view.addNumberControl(content, 0, 408, width, numericSetting{
		Title:   "Maximum outer radius",
		Hint:    "",
		Value:   config.Generation.MaxOuterRadius,
		Read:    func(config simulation.Config) float64 { return config.Generation.MaxOuterRadius },
		Minimum: 2,
		Maximum: 500,
		Step:    .1,
		Apply: func(config *simulation.Config, value float64) {
			setGenerationRadius(config, maximumOuterRadius, value)
		},
	})
}

func (view *Window) buildAppearanceSettings(content *widget.Container, width int) {
	view.addLabel(content, 0, 0, width, "APPEARANCE & WORKSPACE", 10, muted)
	view.toggle(content, 0, 32, width, "Particle glow", view.glowEnabled, func(value bool) {
		view.glowEnabled = value
	})
	view.toggle(content, 0, 78, width, "Selected-type influence preview", view.influencePreviewEnabled, func(value bool) {
		view.influencePreviewEnabled = value
	})
	view.toggle(content, 0, 124, width, "Performance display", view.showPerformance, func(value bool) {
		view.showPerformance = value
	})
	view.toggle(content, 0, 170, width, "Fullscreen", ebiten.IsFullscreen(), func(value bool) {
		ebiten.SetFullscreen(value)
	})
	view.addLabel(content, 0, 225, width, "Interface scale", 14, ink)
	for index, value := range []float64{1, 1.25, 1.5} {
		view.addButton(content, index*(width/3), 254, width/3-6, fmt.Sprintf("%.0f%%", value*100), view.scale == value, func() {
			view.scale = value
			view.requestRebuild()
		})
	}

	view.addLabel(content, 0, 302, width, "Scale adapts to keep controls usable in small windows.\nResizing does not change world size or physics.", 11, muted)
	view.addButton(content, 0, 354, width, "Hide inspector", false, func() {
		view.setInspector(false)
	})
}
