package desktop

import (
	"image"

	"github.com/ebitenui/ebitenui/widget"
)

type dialog struct {
	title string
	body  string
	build func(parent *widget.Container, x, y, width, height int)
	// Compact dialogs share a header, summary, and anchored action footer.
	height  int
	eyebrow string
	warning bool
	cancel  func()
}

func (view *Window) dialogRect() image.Rectangle {
	width := min(760, view.width-64)
	height := min(620, view.height-96)
	if view.activeDialog != nil && view.activeDialog.height > 0 {
		width = min(576, view.width-64)
		height = min(view.activeDialog.height, view.height-64)
	}
	return rectangle((view.width-width)/2, (view.height-height)/2, width, height)
}

func (view *Window) openDialog(next *dialog) {
	view.controller.EndGesture()
	view.activeDialog = next
	view.status = ""
	view.requestRebuild()
}

func (view *Window) closeDialog() {
	view.activeDialog = nil
	view.status = ""
	view.requestRebuild()
}

func (view *Window) cancelDialog() {
	if view.activeDialog != nil && view.activeDialog.cancel != nil {
		view.activeDialog.cancel()
		return
	}
	view.closeDialog()
}

func (view *Window) buildDialog() {
	if view.activeDialog.height > 0 {
		view.buildCompactDialog()
		return
	}
	bounds := view.dialogRect()
	x, y := bounds.Min.X+28, bounds.Min.Y+24
	width, height := bounds.Dx()-56, bounds.Dy()-48

	view.addLabel(view.rootContainer, x, y, width-80, view.activeDialog.title, 24, ink)
	view.addButton(view.rootContainer, x+width-66, y-3, 66, "Close", false, view.closeDialog)
	view.addLabel(view.rootContainer, x, y+47, width, view.activeDialog.body, 13, muted)
	view.activeDialog.build(view.rootContainer, x, y+112, width, height-142)

	view.addLiveLabel(view.rootContainer, x, bounds.Max.Y-46, width, 11, danger, func() string {
		return shorten(view.status, width/6)
	})
}

func (view *Window) confirm(title, body, actionCaption string, action func()) {
	view.openDialog(&dialog{
		title: title,
		body:  body,
		build: func(parent *widget.Container, x, y, width, height int) {
			view.addButton(parent, x, y+80, 160, actionCaption, true, func() {
				view.closeDialog()
				action()
			})
			view.addButton(parent, x+172, y+80, 120, "Cancel", false, view.closeDialog)
		},
	})
}

// guard protects unsaved configuration before replacing the experiment or closing.
func (view *Window) guard(action func()) {
	view.guardUnsaved(action, false)
}

func (view *Window) requestExit() {
	view.guardUnsaved(func() { view.quit = true }, true)
}

func (view *Window) guardUnsaved(action func(), exiting bool) {
	if !view.controller.Dirty() {
		action()
		return
	}

	title, saveCaption, discardCaption := "Keep your changes?", "Save & continue", "Discard changes"
	body := "Save your latest changes before starting something new."
	if exiting {
		title, saveCaption, discardCaption = "Save before you leave?", "Save & exit", "Exit without saving"
		body = "Your experiment has changes that haven’t been saved."
	}
	view.openDialog(&dialog{
		title:   title,
		body:    body,
		height:  420,
		eyebrow: "UNSAVED EXPERIMENT",
		warning: true,
		build: func(parent *widget.Container, x, y, width, height int) {
			view.addExperimentSummary(parent, x, y, width, "Rules, particle types, world & seed", true)
			view.addLabel(parent, x, y+90, width, "Save a preset to pick up this idea again later.", 13, muted)
			footer := view.dialogRect().Max.Y - 58
			view.addButton(parent, x, footer, 88, "Cancel", false, view.cancelDialog)
			discard := view.addButton(parent, x+width-336, footer, 176, discardCaption, false, func() {
				view.closeDialog()
				action()
			})
			view.styleDestructiveButton(discard)
			view.addButton(parent, x+width-148, footer, 148, saveCaption, true, func() {
				view.save(false, func() {
					view.closeDialog()
					action()
				})
			})
		},
	})
}
