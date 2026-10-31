package desktop

import (
	"strings"

	"github.com/ebitenui/ebitenui/widget"
)

func (view *Window) save(saveAs bool, afterSave func()) {
	if !saveAs && view.controller.Name != "" {
		if err := view.store.Save(view.controller.Name, view.controller.Simulation.Config); err != nil {
			view.reportError(err)
			return
		}
		view.controller.MarkSaved(view.controller.Name)
		view.status = "Preset saved"
		if afterSave != nil {
			afterSave()
		}
		return
	}

	view.openDialog(&dialog{
		title: "Save experiment",
		body:  "Saves rules, types, world dimensions and seed. Loading starts a fresh experiment, not a snapshot of the current motion.",
		build: func(parent *widget.Container, x, y, width, height int) {
			view.addLabel(parent, x, y, width, "Preset name", 14, ink)
			nameInput := view.addTextInput(parent, x, y+30, width, view.controller.Name, nil)

			view.addButton(parent, x, y+90, 150, "Save preset", true, func() {
				name := strings.TrimSpace(nameInput.GetText())
				write := func() {
					view.saveNamedPreset(name, afterSave)
				}
				if view.store.Exists(name) {
					view.confirm("Replace existing preset?", "A preset named “"+name+"” already exists.", "Replace", write)
				} else {
					write()
				}
			})
			view.addButton(parent, x+162, y+90, 110, "Cancel", false, view.closeDialog)
		},
	})
}

func (view *Window) saveNamedPreset(name string, afterSave func()) {
	if err := view.store.Save(name, view.controller.Simulation.Config); err != nil {
		view.reportError(err)
		return
	}

	view.controller.MarkSaved(name)
	view.closeDialog()
	view.status = "Saved " + name
	if afterSave != nil {
		afterSave()
	}
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
