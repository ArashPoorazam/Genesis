package desktop

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ArashPoorazam/Genesis/internal/experiment"
	"github.com/ebitenui/ebitenui/widget"
)

// fileDialog uses in-app controls so preset exchange does not require a desktop portal.
func (view *Window) fileDialog(export bool, directory string) {
	if directory == "" {
		directory = view.store.Directory
		if _, err := os.Stat(directory); err != nil {
			directory, _ = os.UserHomeDir()
		}
	}

	absolutePath, err := filepath.Abs(directory)
	if err != nil {
		view.reportError(err)
		return
	}
	entries, err := os.ReadDir(absolutePath)
	if err != nil {
		view.reportError(err)
		return
	}

	title := "Import experiment"
	if export {
		title = "Export experiment"
	}

	view.openDialog(&dialog{
		title: title,
		body:  "Choose a folder and a JSON file. The current experiment remains intact if a file cannot be read or validated.",
		build: func(parent *widget.Container, x, y, width, height int) {
			view.buildFilePicker(parent, x, y, width, height, absolutePath, entries, export)
		},
	})
}

func (view *Window) buildFilePicker(parent *widget.Container, x, y, width, height int, directory string, entries []os.DirEntry, export bool) {
	view.addButton(parent, x, y, 50, "Up", false, func() {
		view.fileDialog(export, filepath.Dir(directory))
	})
	view.addTextInput(parent, x+60, y, width-60, directory, func(path string) error {
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return fmt.Errorf("choose a folder")
		}
		view.pendingActions = append(view.pendingActions, func() {
			view.fileDialog(export, path)
		})
		return nil
	})

	initialName := "experiment.json"
	if view.controller.Name != "" {
		initialName = view.controller.Name + ".json"
	}
	filename := view.addTextInput(parent, x, y+height-50, width-130, initialName, nil)

	files := visiblePresetFiles(entries)
	content := view.addScrollArea(parent, x, y+46, width, height-110, max(1, len(files))*40)
	for index, entry := range files {
		caption := entry.Name()
		if entry.IsDir() {
			caption = ">  " + caption
		}
		view.addButton(content, 0, index*40, width-24, shorten(caption, 80), false, func() {
			if entry.IsDir() {
				view.fileDialog(export, filepath.Join(directory, entry.Name()))
			} else {
				filename.SetText(entry.Name())
			}
		})
	}

	caption := "Import"
	if export {
		caption = "Export"
	}

	view.addButton(parent, x+width-120, y+height-50, 120, caption, true, func() {
		name := strings.TrimSpace(filename.GetText())
		if filepath.Base(name) != name || !strings.EqualFold(filepath.Ext(name), ".json") {
			view.reportError(fmt.Errorf("choose a JSON filename without a folder path"))
			return
		}

		path := filepath.Join(directory, name)
		if export {
			view.exportPreset(path)
		} else {
			view.importPreset(path)
		}
	})
}

func visiblePresetFiles(entries []os.DirEntry) []os.DirEntry {
	visible := make([]os.DirEntry, 0, len(entries))
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		if entry.IsDir() || strings.EqualFold(filepath.Ext(entry.Name()), ".json") {
			visible = append(visible, entry)
		}
	}
	return visible
}

func (view *Window) exportPreset(path string) {
	write := func() {
		name := view.controller.Name
		if name == "" {
			name = "Untitled experiment"
		}
		preset := experiment.Preset{Name: name, Config: view.controller.Simulation.Config}
		if err := experiment.WritePreset(path, preset); err != nil {
			view.reportError(err)
			return
		}
		view.closeDialog()
		view.status = "Exported " + path
	}

	if _, err := os.Stat(path); err == nil {
		view.confirm("Replace file?", path+" already exists.", "Replace file", write)
	} else {
		write()
	}
}

func (view *Window) importPreset(path string) {
	preset, err := experiment.ReadPreset(path)
	if err != nil {
		view.reportError(err)
		return
	}

	view.guard(func() {
		if err := view.controller.Load(preset.Config, ""); err != nil {
			view.reportError(err)
			return
		}
		view.closeDialog()
		view.status = "Imported " + preset.Name + " · Save to keep a local preset"
	})
}
