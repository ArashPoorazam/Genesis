# Genesis

Genesis is a desktop particle laboratory for exploring emergence: how small entities following simple rules can form complex structures. Change the rules while the world runs, inspect how each particle type responds to others, and save experiments worth revisiting.

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

The application retains Ebitengine 2.10.5 and uses Ebiten UI 0.7.3. It embeds its font and does not require a web server or external assets. On Linux, an inherited `EGL_PLATFORM=wayland` is adjusted to `x11` inside this process to match Ebitengine's backend; desktop settings are not modified.

## Explore

- **Relationships:** select a matrix cell, then edit its slider or number. Rows are the responding type; columns are the influencing type. Positive values attract, negative values repel, and zero still retains inner-zone repulsion. Reciprocal editing is optional.
- **Particles:** add up to 12 types, edit names and colors, adjust radii, and explicitly commit population changes. Existing particles retain their motion when possible. The total population limit is 5,000.
- **Settings:** adjust motion live and configure generation, randomization, appearance, and UI scale. Numeric text fields commit on Enter or focus loss; invalid values show inline errors.
- **Randomize relationships:** changes only the interaction matrix.
- **Restart particles:** regenerates positions and initial velocities from the current setup and seed.
- **New world:** chooses a new seed and generates relationships, influence radii, populations, and positions for the current particle types.

The world uses wrapped boundaries. Window size, UI scale, and inspector visibility affect only presentation, not world dimensions or physics.

### Keyboard

| Key | Action |
| --- | --- |
| Space | Play / pause |
| `.` | One physics step while paused |
| Ctrl+Z / Ctrl+Shift+Z | Undo / redo configuration edits |
| Tab / Shift+Tab | Move focus between controls |
| Enter | Activate the focused button or commit a field |
| Arrow keys | Navigate the selected matrix cell or adjust a focused slider |
| Shift+wheel over matrix | Scroll matrix columns |
| F11 | Toggle fullscreen |
| Escape | Close a dialog |

Global playback/history shortcuts are suppressed while typing. A slider drag is one undo entry. Undo changes configuration; it does not rewind simulation time. Dialogs pause simulation updates until dismissed.

## Presets

The toolbar preset menu supports Save As, Load, Rename, Delete, Import, and Export. Presets are versioned JSON containing types, relationships, motion, generation settings, seed, and world dimensions. Loading starts a fresh seeded experiment, rather than restoring the current particle positions.

Named presets live in `genesis/presets` beneath the operating system's user configuration directory (`$XDG_CONFIG_HOME` or `~/.config` on Linux). Appearance and playback preferences are session settings. Writes use a synced temporary file before replacing the destination; invalid imports leave the running experiment intact.

## Code organization

```text
main.go                 Command-line options and application startup
internal/
  simulation/           Configuration, validation, generation, spatial grid, physics
  experiment/           Live edits, history, playback, preset persistence
  desktop/              Ebitengine lifecycle, editor panels, shared controls, rendering
```

See [ARCHITECTURE.md](ARCHITECTURE.md) for the dependency direction, editing lifecycle, and maintenance conventions.

## Verification

```sh
go test ./...
go vet ./...
go test ./internal/simulation -run '^$' -bench BenchmarkPhysics -benchmem
```

An opt-in integration test runs the actual desktop widget lifecycle, using temporary preset files:

```sh
GENESIS_UI_TEST=1 go test ./internal/desktop -run TestDesktopWorkflow -count=1 -v
```

Reproducible rendering checks can capture application screenshots and exit:

```sh
go run . -seed 42 -width 1024 -height 768 -screenshot /tmp/genesis.png
go run . -seed 42 -types 12 -expanded -screenshot /tmp/genesis-matrix.png
```

Screenshot dimensions are fixed by these flags, even under a tiling window manager. For normal sessions, the app follows the window's actual dimensions.
