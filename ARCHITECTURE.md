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

The world coordinate system is independent of the viewport. Resizing only changes the transform used for drawing particles.

Editors are split by responsibility: matrix controls, matrix input, matrix rendering, particle sections, world settings, and preset/file dialogs. Shared widgets define typography, colors, focus handling, and numeric validation. A `numericSetting` keeps each control's label, range, and edit callback together.

Dialog content replaces the active widget tree, preventing input from reaching controls behind it. Rebuilds suppress stale focus-loss commits and discard old bindings before rendering the next tree.

## Maintenance conventions

- Use descriptive domain names such as `particleType`, `relationship`, `population`, and `controller`. Short `x`/`y` coordinates and conventional Go test receivers are appropriate where their meaning is local and obvious.
- Keep constructors and callbacks on multiple lines. Separate validation, calculation, state mutation, and return values with meaningful blank lines; run `gofmt` after edits.
- Extract functions around a distinct operation or editor section. Avoid placing unrelated behavior in the frame loop or adding a general framework for one-off UI needs.
- Keep physics and persistence tests independent of graphics. Use the opt-in desktop workflow test for widget wiring and manual screenshots for visual layout.

The physics benchmark compares the original loop with runtime settings on the same seeded population and repeated evolution window. It measures physics cost only, not rendering or UI throughput.
