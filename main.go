package main

import (
	"flag"
	"log"
	"math/rand/v2"

	"github.com/ArashPoorazam/Genesis/internal/desktop"
)

func main() {
	options := desktop.Options{}
	flag.IntVar(&options.Width, "width", 1440, "initial window width")
	flag.IntVar(&options.Height, "height", 900, "initial window height")
	flag.Uint64Var(&options.Seed, "seed", rand.Uint64(), "initial experiment seed")
	flag.IntVar(&options.TypeCount, "types", 5, "initial particle types (1–12)")
	flag.Float64Var(&options.UIScale, "ui-scale", 1, "UI scale, 1–1.5")
	flag.IntVar(&options.InitialTab, "tab", 0, "inspector: 0 relationships, 1 particles, 2 settings")
	flag.BoolVar(&options.ExpandMatrix, "expanded", false, "expand relationship matrix")
	flag.StringVar(&options.ScreenshotPath, "screenshot", "", "save a PNG and exit after rendering")
	flag.IntVar(&options.CaptureFrames, "capture-frames", 90, "frames before screenshot and exit")
	flag.Parse()

	if err := desktop.Run(options); err != nil {
		log.Fatal(err)
	}
}
