package desktop

import (
	"fmt"
	"image/png"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

type captureOptions struct {
	path          string
	width, height int
	frameLimit    int
	frames        int
}

// captureFrame is enabled only by the explicit screenshot command-line option.
func (view *Window) captureFrame(screen *ebiten.Image) {
	view.capture.frames++
	if view.capture.path == "" || view.capture.frames < view.capture.frameLimit {
		return
	}

	if err := savePNG(screen, view.capture.path); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	view.quit = true
	view.capture.path = ""
}

func savePNG(screen *ebiten.Image, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := png.Encode(file, screen); err != nil {
		file.Close()
		return err
	}
	return file.Close()
}
