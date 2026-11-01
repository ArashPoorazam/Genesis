# Architecture

Genesis separates simulation rules, experiment operations, and desktop presentation. The dependency direction is:

```text
desktop → experiment → simulation
   └────────────────→ simulation configuration types
```

`main.go` parses flags and calls `desktop.Run`. No physics or widget construction belongs in the entry point.

## Simulation

`internal/simulation` has no windowing or filesystem dependency.

- `config.go` defines the saved model; `validation.go` owns its invariants.
- `generation.go` constructs types and random rules with independent seeded random streams.
- `simulation.go` applies validated configurations, retaining particles by stable type ID and rebuilding derived geometry only when necessary.
- `spawning.go` handles population changes and bounded placement attempts.
- `physics.go` calculates directed forces and advances a fixed step. `grid.go` limits neighbor searches and handles wrapped coordinates.

A type's `ID` is stable; a particle's `TypeIndex` is a dense runtime array index. The matrix row responds to the column. Influence radii belong to the influencing column type. Keep these conventions consistent when adding features.

Relationship locks are an optional square matrix in the configuration. Missing locks in older presets mean all cells are editable. Cloning, type addition/removal, presets, and history preserve the lock matrix. Manual, reciprocal, clear, and random edits use `SetRelationship` to honor locks. Neutral-type selection uses its own seeded random stream and excludes types with any locked nonzero incoming or outgoing connection. Loading another preset and explicit undo/redo restore their complete configurations, including locks.

## Experiment operations

`internal/experiment.Controller` owns the running simulation, configuration history, playback state, and saved baseline. Edits clone and validate the complete configuration before committing it. A gesture captures the starting configuration once and commits one undo entry on release.

Undo/redo stores configurations, not particle snapshots. Surviving particles keep their current position and velocity. Restoring a removed type creates replacement particles; it cannot restore that type's elapsed trajectory.

Preset storage validates names and complete files before replacing state. JSON writes occur through a temporary file in the destination directory. The schema is versioned independently of the desktop presentation. World generation and restart use the saved seed; appearance controls are session preferences.

## Desktop presentation

`internal/desktop.Window` implements Ebitengine's `Layout`, `Update`, and `Draw` methods.

1. Input and widget callbacks request controller operations during `Update`.
2. Structural UI changes schedule a rebuild at the next update boundary.
3. Valid physics edits take effect before the next fixed simulation step.
4. `Draw` reads simulation state and refreshes rendering caches when appearance changes.

The world coordinate system is independent of the viewport. Resizing only changes the transform used for drawing particles. The canvas reaches both side edges and the bottom below the top toolbar. Positions use separate horizontal and vertical scales to fill it without cropping; particle markers use the smaller scale to remain circular, and influence previews follow both axes.

The welcome screen decodes the embedded logo once and consumes its dismissal input before entering the laboratory. The inspector is a retained widget container toggled by a single edge handle over a fixed viewport. Visibility changes neither rebuild controls nor animate a relayout of the widget tree. Its three navigation tabs exist only inside the inspector. Matrix drawing and hit testing share its origin. Settings use small sections with independent remembered scroll positions; scrolling applies wheel input directly without an easing backlog. Content shorter than the viewport fills its height and resets wheel and restored offsets to zero. The drawer pull tab shares the panel outline and renders an embedded atom sticker. Relationship layout reserves space for a selected-connection card before calculating visible matrix rows. Temporary control textures are released on structural rebuilds.

Rounded shapes are rasterized once per corner radius, then drawn as tinted nine-slices. Stars also reuse a sprite. This avoids repeated antialiased vector tessellation and graphics submissions for every matrix cell and background on every frame. Buttons use immediate hover states instead of generating transitional surfaces. Font faces are cached by size, use a 13 px minimum and variable weight 550 (650 for headings); buttons and inputs use 15 px. Compare weights visually with the optional performance-test typography preview.

The opt-in `TestDesktopPerformance` measures uncapped frame intervals with physics paused, including closed, rules, particles, and settings states. In the local 1440 × 900 / 12-type comparison, median intervals changed from 92.87 to 1.67 ms for rules, 25.90 to 6.38 ms for particles, and 21.14 to 4.43 ms for settings. These are local measurements, not frame-rate guarantees. CPU profiling identified repeated vector fill submission as a major cost. Screenshot readback is deferred until all measurements finish.

Editors are split by responsibility: matrix controls, matrix input, matrix rendering, particle sections, world settings, and preset/file dialogs. Shared widgets define typography, colors, focus handling, and numeric validation. A `numericSetting` keeps each control's label, range, and edit callback together.

Numeric controls read committed values through a binding so dependent controls refresh without rebuilding the editor or interrupting a drag. Programmatic slider updates are distinguished from user edits to avoid rounding unchanged values or introducing history entries. `linked_settings.go` applies the UI's dependent-radius adjustments within one controller edit; simulation validation remains the final authority for imports and live edits.

Dialog content replaces the active widget tree, preventing input from reaching controls behind it. Rebuilds suppress stale focus-loss commits and discard old bindings before rendering the next tree.

## Maintenance conventions

- Use descriptive domain names such as `particleType`, `relationship`, `population`, and `controller`. Short `x`/`y` coordinates and conventional Go test receivers are appropriate where their meaning is local and obvious.
- Keep constructors and callbacks on multiple lines. Separate validation, calculation, state mutation, and return values with meaningful blank lines; run `gofmt` after edits.
- Extract functions around a distinct operation or editor section. Avoid placing unrelated behavior in the frame loop or adding a general framework for one-off UI needs.
- Keep physics and persistence tests independent of graphics. Use the opt-in desktop workflow test for widget wiring and manual screenshots for visual layout.

The physics benchmark compares the original loop with runtime settings on the same seeded population and repeated evolution window. It measures physics cost only, not rendering or UI throughput.
