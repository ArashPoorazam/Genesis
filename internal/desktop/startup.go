package desktop

import (
	"bytes"
	"fmt"
	"os"
	"runtime"

	"github.com/ArashPoorazam/Genesis/internal/experiment"
	"github.com/ArashPoorazam/Genesis/internal/simulation"
	"github.com/ebitenui/ebitenui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/goregular"
)

type Options struct {
	Width          int
	Height         int
	Seed           uint64
	TypeCount      int
	UIScale        float64
	InitialTab     int
	ExpandMatrix   bool
	ScreenshotPath string
	CaptureFrames  int
}

func Run(options Options) error {
	// Ebitengine 2.10 uses X11 on Linux, including inside Wayland compositors.
	// Adjust this process only, never the user's desktop configuration.
	if runtime.GOOS == "linux" && os.Getenv("EGL_PLATFORM") == "wayland" {
		if err := os.Setenv("EGL_PLATFORM", "x11"); err != nil {
			return err
		}
	}

	view, err := newWindow(options)
	if err != nil {
		return err
	}

	ebiten.SetWindowSize(options.Width, options.Height)
	ebiten.SetWindowSizeLimits(1024, 768, -1, -1)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("Genesis · Particle laboratory")
	ebiten.SetWindowClosingHandled(true)

	return ebiten.RunGame(view)
}

func newWindow(options Options) (*Window, error) {
	config, err := initialConfig(options)
	if err != nil {
		return nil, err
	}

	controller, err := experiment.NewController(config)
	if err != nil {
		return nil, err
	}
	font, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		return nil, err
	}
	store, err := experiment.DefaultStore()
	if err != nil {
		return nil, err
	}

	return &Window{
		controller:       controller,
		store:            store,
		font:             font,
		ui:               &ebitenui.UI{},
		scale:            max(1, min(1.5, options.UIScale)),
		inspectorVisible: true,
		glowEnabled:      true,
		showPerformance:  true,
		needsRebuild:     true,
		activeTab:        max(relationshipsTab, min(settingsTab, options.InitialTab)),
		matrixExpanded:   options.ExpandMatrix,
		sprites:          make(map[string]particleSprite),
		capture: captureOptions{
			path:       options.ScreenshotPath,
			width:      options.Width,
			height:     options.Height,
			frameLimit: options.CaptureFrames,
		},
	}, nil
}

func initialConfig(options Options) (simulation.Config, error) {
	if options.TypeCount < 1 || options.TypeCount > simulation.MaxTypes {
		return simulation.Config{}, fmt.Errorf("types must be between 1 and %d", simulation.MaxTypes)
	}

	config := simulation.Default(options.Seed)
	for len(config.ParticleTypes) < options.TypeCount {
		id := fmt.Sprintf("type-%d", len(config.ParticleTypes))
		if err := config.AddType(id); err != nil {
			return simulation.Config{}, err
		}
	}
	for len(config.ParticleTypes) > options.TypeCount {
		if err := config.RemoveType(len(config.ParticleTypes) - 1); err != nil {
			return simulation.Config{}, err
		}
	}

	return config, nil
}
