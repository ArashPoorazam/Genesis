package desktop

import (
	"image"

	"github.com/ebitenui/ebitenui/widget"
)

type dialog struct {
	title string
	body  string
	build func(parent *widget.Container, x, y, width, height int)
}

func (view *Window) dialogRect() image.Rectangle {
	width := min(760, view.width-64)
	height := min(620, view.height-96)
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

func (view *Window) buildDialog() {
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
	if !view.controller.Dirty() {
		action()
		return
	}

	view.openDialog(&dialog{
		title: "Keep your experiment?",
		body:  "This setup has unsaved changes. Save it as a preset before continuing, or discard the changes.",
		build: func(parent *widget.Container, x, y, width, height int) {
			view.addButton(parent, x, y+72, 150, "Save & continue", true, func() {
				view.save(false, func() {
					view.closeDialog()
					action()
				})
			})
			view.addButton(parent, x+162, y+72, 150, "Discard changes", false, func() {
				view.closeDialog()
				action()
			})
			view.addButton(parent, x+324, y+72, 100, "Cancel", false, view.closeDialog)
		},
	})
}
