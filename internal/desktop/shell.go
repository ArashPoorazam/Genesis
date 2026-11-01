package desktop

import (
	"fmt"
	"image"
	"image/color"
	"math/rand/v2"

	"github.com/ArashPoorazam/Genesis/internal/simulation"
	uiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

const inspectorWidth = 440

func (view *Window) inspectorRect() image.Rectangle {
	width := inspectorWidth
	if view.matrixExpanded && view.activeTab == relationshipsTab {
		width = view.width - 58
	}
	return rectangle(view.width-width-12, 80, width, view.height-92)
}

func (view *Window) setInspector(visible bool) {
	view.ui.ClearFocus()
	view.matrixFocused = false
	view.inspectorVisible = visible
	view.syncInspectorVisibility()
}

func (view *Window) openInspectorTab(index int) {
	if view.inspectorVisible && view.activeTab == index {
		view.setInspector(false)
		return
	}
	view.activeTab = index
	view.setInspector(true)
	view.requestRebuild()
	view.matrixFocused = index == relationshipsTab
}

func (view *Window) syncInspectorVisibility() {
	if view.inspectorContainer != nil {
		visibility := widget.Visibility_Hide
		if view.inspectorVisible {
			visibility = widget.Visibility_Show
		}
		view.inspectorContainer.GetWidget().SetVisibility(visibility)
	}
	if view.drawerHandle != nil {
		view.drawerHandle.GetWidget().LayoutData = view.drawerHandleRect()
		view.drawerHandle.SetLocation(view.drawerHandleRect())
	}
}

func (view *Window) drawerHandleRect() image.Rectangle {
	left := view.width - 44
	if view.inspectorVisible {
		left = view.inspectorRect().Min.X - 42
	}
	return rectangle(left, 102, 44, 64)
}

func (view *Window) buildUI() {
	view.rebuilding = true
	defer func() { view.rebuilding = false }()
	if len(view.scrollContainers) > 0 && view.scrollKey != "" {
		view.scrollPositions[view.scrollKey] = view.scrollContainers[0].ScrollTop
	}
	view.scrollKey = ""
	view.controller.EndGesture()
	view.ui.ClearFocus()
	for _, texture := range view.transientImages {
		texture.Deallocate()
	}
	view.transientImages = nil
	view.refreshBindings = nil
	view.textFields = nil
	view.scrollContainers = nil
	view.inspectorContainer = nil
	view.drawerHandle = nil
	view.rootContainer = newContainer(view.width, view.height)
	view.ui.Container = view.rootContainer
	view.viewport = rectangle(0, 76, view.width, view.height-76)

	lastTypeIndex := len(view.controller.Simulation.Config.ParticleTypes) - 1
	view.selectedType = min(view.selectedType, lastTypeIndex)
	view.selectedRow = min(view.selectedRow, lastTypeIndex)
	view.selectedColumn = min(view.selectedColumn, lastTypeIndex)
	if view.activeDialog != nil {
		view.buildDialog()
		return
	}
	view.buildToolbar()
	view.buildPlayback()
	view.buildInspector()
	bounds := view.drawerHandleRect()
	view.drawerHandle = view.addButton(view.rootContainer, bounds.Min.X, bounds.Min.Y, bounds.Dx(), "", false, func() {
		view.setInspector(!view.inspectorVisible)
	})
	// The shell paints one continuous drawer silhouette behind this hit target.
	clear := uiimage.NewNineSliceColor(color.Transparent)
	view.drawerHandle.SetImage(&widget.ButtonImage{Idle: clear, Hover: clear, Pressed: clear})
	view.syncInspectorVisibility()
}

func (view *Window) buildToolbar() {
	view.addLabel(view.rootContainer, 58, 22, 92, "Genesis", 21, ink)
	presetButton := view.addButton(view.rootContainer, 154, 23, 100, "", false, view.presetDialog)
	view.refreshBindings = append(view.refreshBindings, func() {
		name := view.controller.Name
		if name == "" {
			name = "Presets"
		}
		if view.controller.Dirty() {
			name += " *"
		}
		presetButton.SetText(shorten(name, 10))
	})
	view.addButton(view.rootContainer, 262, 23, 52, "Save", false, func() { view.save(false, nil) })
	right := view.width - 284
	undo := view.addButton(view.rootContainer, right, 23, 58, "Undo", false, func() {
		view.controller.Undo()
		view.requestRebuild()
	})
	redo := view.addButton(view.rootContainer, right+66, 23, 58, "Redo", false, func() {
		view.controller.Redo()
		view.requestRebuild()
	})
	view.refreshBindings = append(view.refreshBindings, func() {
		undo.GetWidget().Disabled = !view.controller.CanUndo()
		redo.GetWidget().Disabled = !view.controller.CanRedo()
	})
	view.addButton(view.rootContainer, right+132, 23, 98, "Fullscreen", false, func() { ebiten.SetFullscreen(!ebiten.IsFullscreen()) })
	view.addButton(view.rootContainer, view.width-50, 23, 30, "?", false, view.shortcutDialog)
}

func (view *Window) buildInspector() {
	bounds := view.inspectorRect()
	view.inspectorContainer = newContainer(bounds.Dx(), bounds.Dy())
	placeWidget(view.rootContainer, view.inspectorContainer, bounds)
	parent := view.inspectorContainer
	width := bounds.Dx() - 40
	for index, title := range []string{"Relationships", "Particles", "Settings"} {
		view.addButton(parent, 20+index*(width/3), 18, width/3-5, title, view.activeTab == index, func() {
			view.activeTab = index
			view.requestRebuild()
		})
	}
	switch view.activeTab {
	case relationshipsTab:
		view.buildMatrix(20, 78, width)
	case particlesTab:
		view.scrollKey = fmt.Sprintf("particles/%d", view.selectedType)
		view.buildParticles(20, 78, width, bounds.Dy()-98)
	case settingsTab:
		view.scrollKey = fmt.Sprintf("settings/%d", view.settingsSection)
		view.buildSettings(20, 78, width, bounds.Dy()-98)
	}
}

func (view *Window) playbackRect() image.Rectangle {
	return rectangle((view.width-380)/2, 23, 380, 34)
}

func (view *Window) buildPlayback() {
	bounds := view.playbackRect()
	x, y := bounds.Min.X, bounds.Min.Y
	play := view.addButton(view.rootContainer, x, y, 68, "", true, func() {
		view.controller.Paused = !view.controller.Paused
	})
	step := view.addButton(view.rootContainer, x+76, y, 56, "Step", false, func() {
		if view.controller.Paused {
			view.controller.Simulation.Step()
		}
	})
	view.refreshBindings = append(view.refreshBindings, func() {
		caption := "Pause"
		if view.controller.Paused {
			caption = "Play"
		}
		play.SetText(caption)
		step.GetWidget().Disabled = !view.controller.Paused
	})
	speed := view.addButton(view.rootContainer, x+140, y, 48, "1×", false, view.cyclePlaybackSpeed)
	view.refreshBindings = append(view.refreshBindings, func() { speed.SetText(fmt.Sprintf("%g×", view.controller.Speed)) })
	view.addButton(view.rootContainer, x+204, y, 74, "Restart", false, func() {
		view.controller.Restart()
		view.status = "Restarted from the current seed"
	})
	view.addButton(view.rootContainer, x+286, y, 94, "New world", false, view.newWorld)
}

func (view *Window) cyclePlaybackSpeed() {
	speeds := []float64{0.25, 0.5, 1, 2, 4}
	for index, speed := range speeds {
		if speed == view.controller.Speed {
			view.controller.Speed = speeds[(index+1)%len(speeds)]
			return
		}
	}
}

func (view *Window) newWorld() {
	view.guard(func() {
		err := view.editConfig(func(config *simulation.Config) error {
			config.RandomizeWorld(rand.Uint64())
			return nil
		})
		if err != nil {
			return
		}
		view.controller.Restart()
		view.status = "New world generated"
		view.requestRebuild()
	})
}

func (view *Window) drawStatus(screen *ebiten.Image) {
	if view.inspectorVisible && view.matrixExpanded && view.activeTab == relationshipsTab {
		return
	}
	state := "LIVE"
	if view.controller.Paused {
		state = "PAUSED"
	}
	caption := fmt.Sprintf("%s  /  %d particles", state, len(view.controller.Simulation.Particles))
	if frameRate := ebiten.ActualFPS(); view.showPerformance && frameRate > 0 {
		caption += fmt.Sprintf("  /  %.0f fps", frameRate)
	}
	view.drawText(screen, caption, 16, float64(view.height-28), 13, muted)
	message := view.status
	if view.controller.Message != "" {
		message = view.controller.Message
	}
	if message != "" {
		view.drawText(screen, shorten(message, max(40, (view.width-inspectorWidth-100)/8)), 16, float64(view.height-52), 13, accent)
	}
}

func shorten(value string, maximumRunes int) string {
	characters := []rune(value)
	if len(characters) > maximumRunes && maximumRunes > 1 {
		return string(characters[:maximumRunes-1]) + "…"
	}
	return value
}
