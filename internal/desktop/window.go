// Package desktop renders the particle laboratory and translates UI events into experiment commands.
package desktop

import (
	"image"

	"github.com/ArashPoorazam/Genesis/internal/experiment"
	"github.com/ArashPoorazam/Genesis/internal/simulation"
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

const (
	relationshipsTab = iota
	particlesTab
	settingsTab
)

// Window owns presentation state. Simulation configuration lives in the controller.
type Window struct {
	controller *experiment.Controller
	store      experiment.Store
	ui         *ebitenui.UI
	font       *text.GoTextFaceSource

	width                   int
	height                  int
	scale                   float64
	viewport                image.Rectangle
	inspectorVisible        bool
	activeTab               int
	selectedType            int
	glowEnabled             bool
	influencePreviewEnabled bool
	showPerformance         bool

	matrixExpanded     bool
	matrixFocused      bool
	linked             bool
	selectedRow        int
	selectedColumn     int
	firstVisibleRow    int
	firstVisibleColumn int
	matrixBounds       image.Rectangle
	matrixRows         int
	matrixCols         int
	cellWidth          int
	cellHeight         int

	rootContainer    *widget.Container
	activeDialog     *dialog
	textFields       []*widget.TextInput
	scrollContainers []*widget.ScrollContainer
	refreshBindings  []func()
	pendingActions   []func()
	needsRebuild     bool
	rebuilding       bool

	sprites map[string]particleSprite
	status  string
	quit    bool
	capture captureOptions
}

func (view *Window) Layout(width, height int) (int, int) {
	if view.capture.path != "" {
		width, height = view.capture.width, view.capture.height
	}

	availableScale := min(float64(width)/1024, float64(height)/768)
	effectiveScale := min(view.scale, max(1, availableScale))
	width = int(float64(width) / effectiveScale)
	height = int(float64(height) / effectiveScale)

	if width != view.width || height != view.height {
		view.width, view.height = width, height
		view.requestRebuild()
	}

	return width, height
}

func (view *Window) Update() error {
	if view.quit {
		return ebiten.Termination
	}
	if ebiten.IsWindowBeingClosed() && view.activeDialog == nil {
		view.guard(func() {
			view.quit = true
		})
	}

	if view.needsRebuild {
		view.buildUI()
		view.needsRebuild = false
	}

	view.ui.Update()
	view.runPendingActions()
	view.activateFocusedButton()

	if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		view.controller.EndGesture()
	}

	if view.activeDialog == nil {
		if !view.isTyping() {
			view.handleShortcuts()
			view.updateMatrix()
		}
		view.controller.Tick()
	} else if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		view.closeDialog()
	}

	// A structural edit invalidates old bindings until the next rebuild.
	if !view.needsRebuild {
		for _, refresh := range view.refreshBindings {
			refresh()
		}
	}

	return nil
}

func (view *Window) runPendingActions() {
	actions := view.pendingActions
	view.pendingActions = nil
	for _, action := range actions {
		action()
	}
}

func (view *Window) activateFocusedButton() {
	if view.isTyping() || !inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		return
	}
	if button, ok := view.ui.GetFocusedWidget().(*widget.Button); ok {
		button.Click()
	}
}

func (view *Window) handleShortcuts() {
	controlPressed := ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta)
	if controlPressed && inpututil.IsKeyJustPressed(ebiten.KeyZ) {
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			view.controller.Redo()
		} else {
			view.controller.Undo()
		}
		view.requestRebuild()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		view.controller.Paused = !view.controller.Paused
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyPeriod) && view.controller.Paused {
		view.controller.Simulation.Step()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF11) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}
}

func (view *Window) isTyping() bool {
	_, focused := view.ui.GetFocusedWidget().(*widget.TextInput)
	return focused
}

func (view *Window) requestRebuild() {
	view.needsRebuild = true
}

func (view *Window) reportError(err error) {
	if err != nil {
		view.status = err.Error()
	}
}

func (view *Window) editConfig(change func(*simulation.Config) error) error {
	err := view.controller.Edit(change)
	if err == nil {
		view.status = ""
	} else {
		view.reportError(err)
	}
	return err
}
