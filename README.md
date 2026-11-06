# Genesis

This project is inspired by particle life.

Genesis is a desktop particle laboratory for exploring emergence: how small entities following simple rules can form complex structures. Change the rules while the world runs, inspect how each particle type responds to others, and save experiments worth revisiting.

![Genesis Logo](images/logo.png)

## Run

Use the Go version declared in `go.mod` and a desktop session with graphics support:

```sh
go run .
```

To build a standalone executable:

```sh
go build -o genesis .
./genesis
```

The application retains Ebitengine 2.10.5 and uses Ebiten UI 0.7.3. The welcome artwork (`images/logo.png`) and Adwaita Sans font (an Inter-derived interface typeface) are embedded in the executable. The font's SIL Open Font License is included in `internal/desktop/assets/FONT-LICENSE`. No web server or external assets are required. On Linux, an inherited `EGL_PLATFORM=wayland` is adjusted to `x11` inside this process to match Ebitengine's backend; desktop settings are not modified.

## Explore

The logo welcome screen opens with **press anything to continue...**; a key, mouse button, or touch enters the laboratory. The simulation starts with its inspector closed. Playback is centered in one top toolbar, with presets on the left and history, fullscreen, and help on the right. The simulation fills the remaining screen, including the sides and bottom. Interface text uses medium-weight Adwaita Sans, with 15 px buttons and inputs, a 13 px minimum for supporting text, and heavier headings.

- **Inspector:** use the integrated atom-sticker edge handle to open or close it; the three section tabs appear only inside. Opening is immediate, without a sliding animation or rebuilding controls. `1` / `2` / `3` still open rules, particles, or settings; the same shortcut closes the section. Scroll positions are remembered when content overflows; sections that fit stay anchored at the top. Shortcuts live in tooltips and the keyboard guide, not button captions.
- **Relationships:** select a matrix cell, then edit its slider or number. Right-click a cell to lock or unlock it; locked cells show a small padlock and cannot be edited, cleared, or shuffled. Locks are directional, saved with presets, and included in undo/redo. Reciprocal edits leave any locked reverse cell unchanged. Rows are the responding type; columns are the influencing type. Positive values attract, negative values repel, and zero still retains inner-zone repulsion.
- **Particles:** add up to 12 types, edit names and colors, adjust radii, and explicitly commit population changes. Existing particles retain their motion when possible. The total population limit is 5,000.
- **Settings:** compact Motion, Spawn, Ranges, Display, and Sound sections organize live physics, generation, randomization, appearance, and audio. Numeric text fields commit on Enter or focus loss; invalid values show inline errors. Sliders and number fields stay synchronized.
- **Linked radii:** changing a particle's size or influence zones adjusts dependent radii to preserve valid spacing. Generation bounds move neighboring bounds when necessary. Physical particle size sets the lower limit, and one undo restores all changes from a slider drag.
- **Shuffle relationships:** randomizes unlocked cells independently and leaves particles in place. No particle type is automatically made neutral. Locked values are preserved; manually zeroed relationships and a zero shuffle amplitude remain supported.
- **Restart particles:** regenerates positions and initial velocities from the current setup and seed.
- **New world:** chooses a new seed and generates relationships, influence radii, populations, and positions for the current particle types.

Startup defaults use relationship values from −1 to +1, inner repulsion strength 5, particle radius 1.5, fixed inner influence radii at interval midpoints between 18 and 25 (18.7, 20.1, 21.5, 22.9, 24.3 for five types), and fixed outer influence radii evenly spaced at interval midpoints between 40 and 120 (48, 64, 80, 96, 112 for the default five types). The random population ceiling is 500 per type (generated populations range from 1 to 500, subject to the 5,000 total limit). Particle glow defaults to off. World generation uses the same lock-aware relationship shuffle. Both inner and outer radii use the midpoint of each equal interval in type order: four types with bounds 80–120 receive 85, 95, 105, and 115. Populations remain random, with at least one reserved for every type, including newly added colors. Spawning alternates between types so earlier colors do not fill all available space first; insufficient space is still reported. Existing presets retain their saved values.

The world uses wrapped boundaries. Window size, UI scale, and inspector visibility affect only presentation, not world dimensions or physics. Positions map across the full available width and height without letterboxing or cropping. Particle markers remain circular; influence previews follow the horizontal and vertical world scales.

Physics automatically uses multiple CPU cores for larger populations, with serial execution for small worlds or a single available core. This requires no additional drivers or dependencies and preserves the serial simulation's force accumulation order. Rendering continues to use Ebitengine's graphics backend.

### GPU boost

Enable **Settings → Motion → GPU boost (experimental)** for optional GPU physics, or start with `go run . -gpu`. It defaults to off and applies only to the current session, independently of presets and undo history. Playback and both single-step controls use the selected backend. GPU boost uses Ebitengine's portable Kage shaders, so CUDA and NVIDIA-specific libraries are not required. It uses the graphics device selected by Ebitengine; a hybrid laptop may use its integrated GPU by default.

Activation runs a small accuracy check on the actual device. Errors returned by initialization, that check, or a velocity update fall back to CPU physics with a status message. A failed update completes once on the CPU without committing a partial GPU result, and the toggle returns to Off. Ebitengine shares the graphics backend with rendering: fatal backend errors, including driver or backend shader-compilation failures, can still terminate the application. Restart without `-gpu` in that case.

GPU arithmetic and packed particle coordinates have lower precision than CPU physics. The force law and spatial-neighbor search are retained, but trajectories are not bit-for-bit reproducible across backends and can diverge over time. Small or sparse worlds can be faster on the CPU; GPU boost is most useful for large, dense populations.

Local Linux checks passed on an RTX 3050 Ti Laptop GPU and Intel integrated graphics. In the 5,000-particle dense-cluster test, complete physics steps (including uploads and readback) measured approximately 31.0 → 6.5 ms on NVIDIA and 28.9 → 19.5 ms on Intel, compared with multicore CPU execution. These are median timings from synthetic snapshots, not desktop frame-rate guarantees. AMD, Windows, and macOS have not been hardware-tested.

On a Linux NVIDIA hybrid system with PRIME offload configured, select the discrete graphics path with:

```sh
__NV_PRIME_RENDER_OFFLOAD=1 __GLX_VENDOR_LIBRARY_NAME=nvidia go run . -gpu
```

### Sound

Genesis includes **Quiet Orbits**, the original handpan soundtrack, as a seamless background loop. Background music defaults to off. Soft water-drop cues accompany only play/pause, successful saves and loads, restarting particles, creating a world, shuffling relationships, and updating population. Routine controls are silent. Cues briefly lower enabled music and never stack over one another. When enabled, music continues while the simulation is paused.

Open **Settings → Sound** for separate music and effects switches and volume sliders. **M** mutes or unmutes everything when you are not typing. Listening preferences are remembered in `genesis/audio.json` beneath the user configuration directory, separately from experiment presets and undo history. Audio is embedded in the executable and works offline. Use `-mute` to run without opening an audio device; screenshot captures are automatically silent.

### Keyboard

| Key                                                  | Action                                                                  |
| ---------------------------------------------------- | ----------------------------------------------------------------------- |
| Space                                                | Play / pause                                                            |
| `M`                                                  | Mute / unmute music and effects                                          |
| `.`                                                  | One physics step while paused                                           |
| `1` / `2` / `3`                                      | Toggle rules / particles / settings drawer                              |
| `I`                                                  | Toggle the current inspector                                            |
| `R`                                                  | Restart particles from the current seed                                 |
| Ctrl+S / Ctrl+Shift+S                                | Save / save as                                                          |
| Ctrl+O                                               | Open preset library                                                     |
| Ctrl+Z / Ctrl+Shift+Z                                | Undo / redo configuration edits                                         |
| Tab / Shift+Tab                                      | Move focus between controls                                             |
| Enter                                                | Activate the focused button or commit a field                           |
| Arrow keys                                           | Navigate the selected matrix cell or adjust a focused slider            |
| Right-click a matrix cell / `L` while matrix focused | Toggle that cell’s lock                                                 |
| Shift+wheel over matrix                              | Scroll matrix columns                                                   |
| F11                                                  | Toggle fullscreen                                                       |
| Escape                                               | Leave a focused field or control, then close the drawer; close a dialog |
| `?` / F1                                             | Open the keyboard guide                                                 |

Global shortcuts are suppressed while typing, except Escape to leave the field. On macOS, Command also works for Ctrl shortcuts. Opening Rules with `1` enables matrix arrow navigation; Enter focuses the selected rule's number. A slider drag is one undo entry. Undo changes configuration; it does not rewind simulation time. Dialogs pause simulation updates until dismissed.

## Presets

The toolbar preset menu supports Save As, Load, Rename, Delete, Import, and Export. Presets are versioned JSON containing types, relationships, motion, generation settings, seed, and world dimensions. Loading starts a fresh seeded experiment, rather than restoring the current particle positions.

Named presets live in `genesis/presets` beneath the operating system's user configuration directory (`$XDG_CONFIG_HOME` or `~/.config` on Linux). Appearance and playback preferences are session settings. Writes use a synced temporary file before replacing the destination; invalid imports leave the running experiment intact.

Save and exit dialogs use compact layouts with a live preset-name preview and clear action footers. Enter in the name field saves; Escape cancels or goes back. Closing with unsaved changes offers Save & exit, Exit without saving, or Cancel. A canceled overwrite keeps the entered name, and a failed save keeps the dialog open for another attempt.

## Code organization

```text
main.go                 Command-line options and application startup
internal/
  simulation/           Configuration, validation, generation, spatial grid, physics
  experiment/           Live edits, history, playback, preset persistence
  desktop/              Ebitengine lifecycle, editor panels, shared controls, rendering
```

See [ARCHITECTURE.md](ARCHITECTURE.md) for the dependency direction, editing lifecycle, and maintenance conventions. [CONTRIBUTING.md](CONTRIBUTING.md) explains how to extend the model and editors, preserve invariants, and run the appropriate checks.

## Verification

```sh
go test ./...
go vet ./...
go test ./internal/simulation -run '^$' -bench BenchmarkPhysics -benchmem
go test ./internal/simulation -run '^$' -bench BenchmarkRelationshipEdit -benchmem
go test ./internal/simulation -run '^$' -bench '^BenchmarkPhysicsScaling$' -benchmem -count=3
```

An opt-in integration test runs the actual desktop widget lifecycle, using temporary preset files:

```sh
GENESIS_UI_TEST=1 go test ./internal/desktop -run TestDesktopWorkflow -count=1 -v
```

Verify that adding a color and choosing New world populates every type and spaces both influence radii correctly:

```sh
GENESIS_WORLD_UI_TEST=1 go test ./internal/desktop -run '^TestDesktopNewWorldIncludesAddedColor$' -count=1 -v
```

Verify save, overwrite, and exit flows separately (including failed saves):

```sh
GENESIS_DIALOG_TEST=1 go test ./internal/desktop -run TestDesktopDialogs -count=1 -v
```

Set `GENESIS_DIALOG_SCREENSHOTS` to an existing directory to capture dialog previews during that test.

Verify the audio device, looping, cue suppression, mute, and sound controls in separate processes:

```sh
GENESIS_AUDIO_TEST=1 go test ./internal/soundtrack -run '^TestAudioPlayback$' -count=1 -v
GENESIS_SOUND_UI_TEST=1 go test ./internal/desktop -run '^TestDesktopSoundSettings$' -count=1 -v
```

The playback check briefly plays audio. Set `GENESIS_SOUND_SCREENSHOTS` to an existing directory to capture the Sound panel. See `audio/README.md` for the score, previews, and reproducible synthesis source.

Measure paused UI rendering separately from physics (1440 × 900, 12 types, uncapped frames):

```sh
GENESIS_PERF_TEST=1 go test ./internal/desktop -run TestDesktopPerformance -count=1 -v
```

Set `GENESIS_PERF_SCREENSHOTS` to an existing directory to also save each layout and a comparison of font weights at 11, 13, 15, and 17 px. Run each opt-in display test separately; Ebitengine permits only one game lifecycle per process.

Verify GPU execution, numerical agreement, boundaries, live edits, and complete-step timings on the active graphics device:

```sh
GENESIS_GPU_TEST=1 go test ./internal/gpuphysics -run '^TestGPUPhysics$' -count=1 -v
GENESIS_GPU_UI_TEST=1 go test ./internal/desktop -run '^TestDesktopGPU$' -count=1 -v
```

Run these separately on each graphics device you intend to support. The UI check covers enable/disable, startup selection, load/restart, single-step playback, injected failure, CPU fallback, and retry. Set `GENESIS_GPU_SCREENSHOT` to a PNG path to capture the Motion settings. Ordinary tests compile the shader and check transactional fallback without requiring GPU physics execution.

Reproducible rendering checks can capture application screenshots and exit:

```sh
go run . -seed 42 -screenshot /tmp/genesis-welcome.png
go run . -seed 42 -skip-intro -width 1024 -height 768 -screenshot /tmp/genesis.png
go run . -seed 42 -skip-intro -inspector -tab 2 -screenshot /tmp/genesis-settings.png
go run . -seed 42 -skip-intro -types 12 -expanded -screenshot /tmp/genesis-matrix.png
```

Screenshot dimensions are fixed by these flags, even under a tiling window manager. For normal sessions, the app follows the window's actual dimensions. Use `-skip-intro` to enter the laboratory directly and `-inspector` to start with a drawer open; `-expanded` opens the enlarged rule matrix.
