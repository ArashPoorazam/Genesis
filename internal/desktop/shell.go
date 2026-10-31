package desktop

import (
	"fmt"
	"image"
	"math/rand/v2"

	"github.com/ArashPoorazam/Genesis/internal/simulation"
	"github.com/hajimehoshi/ebiten/v2"
)

const inspectorWidth = 440

func (view *Window) inspectorRect() image.Rectangle {
	if view.matrixExpanded && view.activeTab == relationshipsTab {
		return rectangle(20, 76, view.width-40, view.height-110)
	}
	return rectangle(view.width-inspectorWidth, 64, inspectorWidth, view.height-96)
}

func (view *Window) buildUI() {
	view.rebuilding = true
	defer func() {
		view.rebuilding = false
	}()

	view.controller.EndGesture()
	view.ui.ClearFocus()
	view.refreshBindings = nil
	view.textFields = nil
	view.scrollContainers = nil
	view.rootContainer = newContainer(view.width, view.height)
	view.ui.Container = view.rootContainer

	lastTypeIndex := len(view.controller.Simulation.Config.ParticleTypes) - 1
	view.selectedType = min(view.selectedType, lastTypeIndex)
	view.selectedRow = min(view.selectedRow, lastTypeIndex)
	view.selectedColumn = min(view.selectedColumn, lastTypeIndex)

	if view.activeDialog != nil {
		view.buildDialog()
		return
	}

	canvasRight := view.width
	if view.inspectorVisible {
		canvasRight -= inspectorWidth
	}

	view.viewport = rectangle(24, 92, canvasRight-48, view.height-214)

	view.buildToolbar()
	view.buildPlayback(canvasRight)
	view.buildStatusBar()

	if !view.matrixExpanded || view.activeTab != relationshipsTab || !view.inspectorVisible {
		view.addLabel(view.rootContainer, 26, 72, 300, "WORLD / LIVE EXPERIMENT", 10, muted)
	}
	if view.inspectorVisible {
		view.buildInspector()
	}
}

func (view *Window) buildToolbar() {
	view.addLabel(view.rootContainer, 24, 12, 142, "GENESIS", 22, ink)
	view.addLabel(view.rootContainer, 25, 39, 140, "PARTICLE LABORATORY", 9, muted)

	presetButton := view.addButton(view.rootContainer, 188, 15, 230, "", false, view.presetDialog)
	view.refreshBindings = append(view.refreshBindings, func() {
		name := view.controller.Name
		if name == "" {
			name = "Untitled experiment"
		}
		if view.controller.Dirty() {
			name += " *"
		}
		presetButton.SetText(shorten(name, 27) + "  v")
	})

	view.addButton(view.rootContainer, 428, 15, 64, "Save", true, func() {
		view.save(false, nil)
	})
	undoButton := view.addButton(view.rootContainer, 504, 15, 64, "Undo", false, func() {
		view.controller.Undo()
		view.requestRebuild()
	})
	redoButton := view.addButton(view.rootContainer, 576, 15, 64, "Redo", false, func() {
		view.controller.Redo()
		view.requestRebuild()
	})
	view.refreshBindings = append(view.refreshBindings, func() {
		undoButton.GetWidget().Disabled = !view.controller.CanUndo()
		redoButton.GetWidget().Disabled = !view.controller.CanRedo()
	})

	caption := "Hide inspector"
	if !view.inspectorVisible {
		caption = "Show inspector"
	}

	view.addButton(view.rootContainer, view.width-168, 15, 144, caption, false, func() {
		view.inspectorVisible = !view.inspectorVisible
		view.requestRebuild()
	})
}

func (view *Window) buildInspector() {
	bounds := view.inspectorRect()
	x, y := bounds.Min.X+20, bounds.Min.Y+18
	width := bounds.Dx() - 40

	for index, title := range []string{"Relationships", "Particles", "Settings"} {
		view.addButton(view.rootContainer, x+index*(width/3), y, width/3-5, title, view.activeTab == index, func() {
			view.activeTab = index
			view.requestRebuild()
		})
	}

	contentHeight := bounds.Max.Y - y - 72
	switch view.activeTab {
	case relationshipsTab:
		view.buildMatrix(x, y+56, width)
	case particlesTab:
		view.buildParticles(x, y+56, width, contentHeight)
	case settingsTab:
		view.buildSettings(x, y+56, width, contentHeight)
	}
}

func (view *Window) buildPlayback(right int) {
	x, y := 24, view.height-94
	playButton := view.addButton(view.rootContainer, x, y, 94, "Pause", true, func() {
		view.controller.Paused = !view.controller.Paused
	})
	stepButton := view.addButton(view.rootContainer, x+102, y, 60, "Step", false, func() {
		if view.controller.Paused {
			view.controller.Simulation.Step()
		}
	})

	view.refreshBindings = append(view.refreshBindings, func() {
		caption := "Pause"
		if view.controller.Paused {
			caption = "Play"
		}
		playButton.SetText(caption)
		stepButton.GetWidget().Disabled = !view.controller.Paused
	})

	speedButton := view.addButton(view.rootContainer, x+170, y, 66, "1×", false, view.cyclePlaybackSpeed)
	view.refreshBindings = append(view.refreshBindings, func() {
		speedButton.SetText(fmt.Sprintf("%g×", view.controller.Speed))
	})

	view.addButton(view.rootContainer, x+244, y, 126, "Restart particles", false, func() {
		view.controller.Restart()
		view.status = "Restarted from the current seed"
	})
	view.addButton(view.rootContainer, max(right-132, x+382), y, 108, "New world", false, view.newWorld)
	view.addLabel(view.rootContainer, x, y+44, right-48, "SPACE pause   ·   . single step   ·   F11 fullscreen", 11, muted)
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

func (view *Window) buildStatusBar() {
	view.addLiveLabel(view.rootContainer, 24, view.height-24, 270, 11, muted, func() string {
		state := "RUNNING"
		if view.controller.Paused {
			state = "PAUSED"
		}

		caption := fmt.Sprintf("%s   ·   %d particles", state, len(view.controller.Simulation.Particles))
		if frameRate := ebiten.ActualFPS(); view.showPerformance && frameRate > 0 {
			caption += fmt.Sprintf("   ·   %.0f FPS", frameRate)
		}
		return caption
	})

	view.addLiveLabel(view.rootContainer, 310, view.height-24, view.width-334, 11, accent, func() string {
		message := view.status
		if view.controller.Message != "" {
			message = view.controller.Message
		}
		return shorten(message, (view.width-334)/6)
	})
}

func shorten(value string, maximumRunes int) string {
	characters := []rune(value)
	if len(characters) > maximumRunes && maximumRunes > 1 {
		return string(characters[:maximumRunes-1]) + "…"
	}
	return value
}
