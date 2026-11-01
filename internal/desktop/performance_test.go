package desktop

import (
	"fmt"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

func TestDesktopPerformance(t *testing.T) {
	if os.Getenv("GENESIS_PERF_TEST") != "1" {
		t.Skip("set GENESIS_PERF_TEST=1 to measure the desktop renderer")
	}
	t.Setenv("EGL_PLATFORM", "x11")
	view, err := newWindow(Options{TypeCount: 12, Seed: 42, UIScale: 1, SkipIntro: true})
	if err != nil {
		t.Fatal(err)
	}
	view.controller.Paused = true
	ebiten.SetWindowSize(1440, 900)
	ebiten.SetVsyncEnabled(false)
	ebiten.SetTPS(ebiten.SyncWithFPS)
	driver := &performanceDriver{view: view, test: t}
	if err := ebiten.RunGame(driver); err != nil {
		t.Fatal(err)
	}
}

type performanceDriver struct {
	view          *Window
	test          *testing.T
	stage, frames int
	previous      time.Time
	samples       []float64
	previews      []*ebiten.Image
}

func (driver *performanceDriver) Layout(_, _ int) (int, int) {
	return driver.view.Layout(1440, 900)
}

func (driver *performanceDriver) Update() error {
	return driver.view.Update()
}

func (driver *performanceDriver) Draw(screen *ebiten.Image) {
	now := time.Now()
	if driver.frames > 60 {
		driver.samples = append(driver.samples, float64(now.Sub(driver.previous).Microseconds())/1000)
	}
	driver.previous = now
	driver.view.Draw(screen)
	driver.frames++
	if driver.frames < 300 {
		return
	}
	sort.Float64s(driver.samples)
	names := []string{"closed", "rules", "particles", "settings"}
	driver.test.Logf("%s frame interval: median %.2f ms, p95 %.2f ms", names[driver.stage], driver.samples[len(driver.samples)/2], driver.samples[len(driver.samples)*95/100])
	if directory := os.Getenv("GENESIS_PERF_SCREENSHOTS"); directory != "" {
		preview := ebiten.NewImage(screen.Bounds().Dx(), screen.Bounds().Dy())
		preview.DrawImage(screen, nil)
		driver.previews = append(driver.previews, preview)
	}
	driver.stage++
	if driver.stage == len(names) {
		for index, preview := range driver.previews {
			if err := savePNG(preview, fmt.Sprintf("%s/%s.png", os.Getenv("GENESIS_PERF_SCREENSHOTS"), names[index])); err != nil {
				driver.test.Error(err)
			}
			preview.Deallocate()
		}
		if directory := os.Getenv("GENESIS_PERF_SCREENSHOTS"); directory != "" {
			if err := driver.view.saveTypographyPreview(directory + "/typography.png"); err != nil {
				driver.test.Error(err)
			}
		}
		driver.view.quit = true
		return
	}
	driver.frames = 0
	driver.samples = nil
	driver.view.activeTab = driver.stage - 1
	driver.view.setInspector(true)
	driver.view.requestRebuild()
}

func (view *Window) saveTypographyPreview(path string) error {
	preview := ebiten.NewImage(1200, 440)
	defer preview.Deallocate()
	preview.Fill(panelBG)
	for column, weight := range []float32{400, 550, 700} {
		left := float64(24 + column*400)
		view.drawText(preview, fmt.Sprintf("Adwaita Sans · weight %.0f", weight), left, 20, 20, ink)
		for row, size := range []float64{11, 13, 15, 17} {
			face := &text.GoTextFace{Source: view.font, Size: size}
			face.SetVariation(text.MustParseTag("wght"), weight)
			options := &text.DrawOptions{}
			options.GeoM.Translate(left, float64(70+row*88))
			options.ColorScale.ScaleWithColor(ink)
			text.Draw(preview, fmt.Sprintf("%.0f px · Velocity retention / Pause", size), face, options)
			options.GeoM.Translate(0, 29)
			text.Draw(preview, "0.90  100.10  −1.25  Inner influence radius", face, options)
		}
	}
	return savePNG(preview, path)
}
