// Package audioassets contains the handpan score and water-drop action cue.
package audioassets

import "embed"

//go:embed handpan/music/quiet-orbits-loop.ogg sfx/water-drop.wav
var Files embed.FS
