package desktop

import (
	"errors"
	"os"
	"strings"

	"github.com/ebitenui/ebitenui/widget"
)

func (view *Window) save(saveAs bool, afterSave func()) {
	if !saveAs && view.controller.Name != "" {
		if err := view.store.Save(view.controller.Name, view.controller.Simulation.Config); err != nil {
			view.reportPresetSaveError(err)
			return
		}
		view.controller.MarkSaved(view.controller.Name)
		view.status = "Preset saved"
		if afterSave != nil {
			afterSave()
		}
		return
	}

	view.savePresetDialog(view.controller.Name, afterSave, view.activeDialog)
}

func (view *Window) savePresetDialog(draft string, afterSave func(), previous *dialog) {
	next := &dialog{
		title:   "Save your experiment",
		body:    "Keep a setup worth coming back to.",
		height:  492,
		eyebrow: "YOUR PRESET LIBRARY",
	}
	if previous != nil {
		next.cancel = func() { view.openDialog(previous) }
	}
	next.build = func(parent *widget.Container, x, y, width, height int) {
		view.addLabel(parent, x, y, width, "Preset name", 14, ink)
		nameInput := view.addTextInput(parent, x, y+28, width, draft, nil)
		nameInput.GetWidget().LayoutData = rectangle(x, y+28, width, 42)
		nameInput.ChangedEvent.AddHandler(func(arguments interface{}) {
			draft = arguments.(*widget.TextInputChangedEventArgs).InputText
			if view.activeDialog == next {
				view.status = ""
			}
		})
		view.addLabel(parent, x, y+80, width, "1–64 characters · Saved on this computer", 13, muted)
		summary := view.addExperimentSummary(parent, x, y+112, width, "Rules, types, world & seed · Loading starts fresh", false)
		view.refreshBindings = append(view.refreshBindings, func() {
			name := strings.TrimSpace(nameInput.GetText())
			if name == "" {
				name = "Untitled experiment"
			}
			summary.Label = view.fitText(name, float64(width-146), 15)
		})
		submit := func() {
			name := strings.TrimSpace(nameInput.GetText())
			draft = name
			if name == "" {
				view.status = "Give your experiment a name before saving."
				return
			}
			if view.store.Exists(name) {
				view.replacePresetDialog(name, afterSave, next)
			} else {
				view.saveNamedPreset(name, afterSave)
			}
		}
		nameInput.SubmitEvent.AddHandler(func(_ interface{}) { submit() })
		footer := view.dialogRect().Max.Y - 58
		view.addButton(parent, x, footer, 88, "Cancel", false, view.cancelDialog)
		view.addButton(parent, x+width-148, footer, 148, "Save preset", true, submit)
		view.pendingActions = append(view.pendingActions, func() {
			if view.activeDialog == next {
				view.ui.SetFocusedWidget(nameInput)
				nameInput.SelectAll()
			}
		})
	}
	view.openDialog(next)
}

func (view *Window) replacePresetDialog(name string, afterSave func(), previous *dialog) {
	view.openDialog(&dialog{
		title:   "Replace this preset?",
		body:    "The saved setup will be replaced with your current experiment.",
		height:  388,
		eyebrow: "PRESET ALREADY EXISTS",
		warning: true,
		cancel:  func() { view.openDialog(previous) },
		build: func(parent *widget.Container, x, y, width, height int) {
			view.addLabel(parent, x, y, width, view.fitText(name, float64(width), 20), 20, ink)
			view.addLabel(parent, x, y+36, width, "Choose another name to keep both versions.", 13, muted)
			footer := view.dialogRect().Max.Y - 58
			view.addButton(parent, x, footer, 100, "Go back", false, view.cancelDialog)
			view.addButton(parent, x+width-148, footer, 148, "Replace", true, func() {
				view.saveNamedPreset(name, afterSave)
			})
		},
	})
}

func (view *Window) saveNamedPreset(name string, afterSave func()) {
	if err := view.store.Save(name, view.controller.Simulation.Config); err != nil {
		view.reportPresetSaveError(err)
		return
	}

	view.controller.MarkSaved(name)
	view.closeDialog()
	view.status = "Saved " + name
	if afterSave != nil {
		afterSave()
	}
}

func (view *Window) reportPresetSaveError(err error) {
	var pathError *os.PathError
	if errors.As(err, &pathError) {
		view.status = "Couldn’t save the preset. Check your preset folder’s permissions and available disk space, then try again."
		return
	}
	view.reportError(err)
}

func (view *Window) presetDialog() {
	names, err := view.store.List()
	if err != nil {
		view.reportError(err)
		return
	}

	view.openDialog(&dialog{
		title: "Your experiments",
		body:  "Named setups stored locally. Load a preset to restart its world from the saved seed.",
		build: func(parent *widget.Container, x, y, width, height int) {
			view.addButton(parent, x, y, 116, "Save as…", true, func() {
				view.save(true, nil)
			})
			view.addButton(parent, x+126, y, 116, "Import JSON", false, func() {
				view.fileDialog(false, "")
			})
			view.addButton(parent, x+252, y, 116, "Export JSON", false, func() {
				view.fileDialog(true, "")
			})

			if len(names) == 0 {
				view.addLabel(parent, x, y+76, width, "No saved presets yet. Save a setup worth returning to.", 14, muted)
				return
			}

			content := view.addScrollArea(parent, x, y+58, width, height-64, len(names)*48)
			for index, name := range names {
				view.buildPresetRow(content, width, index*48, name)
			}
		},
	})
}

func (view *Window) buildPresetRow(parent *widget.Container, width, y int, name string) {
	view.addButton(parent, 0, y, width-202, shorten(name, 45), false, func() {
		preset, err := view.store.Load(name)
		if err != nil {
			view.reportError(err)
			return
		}
		view.guard(func() {
			if err := view.controller.Load(preset.Config, name); err != nil {
				view.reportError(err)
				return
			}
			view.closeDialog()
			view.status = "Loaded " + name
		})
	})

	view.addButton(parent, width-194, y, 82, "Rename", false, func() {
		view.renameDialog(name)
	})
	view.addButton(parent, width-104, y, 80, "Delete", false, func() {
		view.confirm("Delete preset?", "Remove “"+name+"” from this computer? The running experiment is kept.", "Delete preset", func() {
			if err := view.store.Delete(name); err != nil {
				view.reportError(err)
				return
			}
			if view.controller.Name == name {
				view.controller.Name = ""
			}
			view.presetDialog()
		})
	})
}

func (view *Window) renameDialog(previousName string) {
	view.openDialog(&dialog{
		title: "Rename preset",
		body:  "Choose a new name for “" + previousName + "”.",
		build: func(parent *widget.Container, x, y, width, height int) {
			nameInput := view.addTextInput(parent, x, y, width, previousName, nil)
			view.addButton(parent, x, y+58, 130, "Rename", true, func() {
				name := strings.TrimSpace(nameInput.GetText())
				if err := view.store.Rename(previousName, name); err != nil {
					view.reportError(err)
					return
				}
				if view.controller.Name == previousName {
					view.controller.Name = name
				}
				view.presetDialog()
			})
		},
	})
}
