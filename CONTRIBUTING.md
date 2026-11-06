# Working on Genesis

Use the Go version in `go.mod`. The application is a native desktop executable; there is no frontend build, web server, or runtime asset download. Run `go build -o /tmp/genesis .` to check compilation without replacing an existing local executable.

Read [ARCHITECTURE.md](ARCHITECTURE.md) before changing package boundaries. Start with `simulation` for model rules, `experiment` for edits/history/persistence, and the relevant editor file in `desktop` for controls. `main.go` only wires command-line options to startup.

## Readability

- Use names that describe the domain and the operation. Prefer `row`/`column` for relationships and typed radius selectors over unexplained numeric arguments.
- Group validation, calculation, mutation, and results into readable blocks. Use `gofmt` for spacing and indentation; avoid extra blank lines at function ends.
- Comment invariants and reasons: directed forces, wrapped-neighbor uniqueness, cache invalidation, deferred UI rebuilds, and gesture history deserve explanation. Avoid comments that merely restate an assignment.
- Keep callbacks focused on one operation. Extract a named helper when a callback coordinates several dependent values or an editor section has its own responsibility.
- Preserve existing user changes in the working tree. Do not commit generated executables or benchmark profiles.

## Changing behavior

For a new saved setting, update the model and validation together. Ensure `Config.Clone` owns any new slices or maps, and decide whether the edit requires geometry rebuilding, population reconciliation, or neither. Add the control through `numericSetting` when appropriate, including a `Read` binding so dependent values refresh.

For structural editor changes, call `requestRebuild` rather than replacing widgets inside a callback. Old bindings must not run after their type indices become invalid. Rebuild-time focus loss must not commit stale text. Close slider gestures before actions that replace configuration.

Keep seeded generation streams independent. Changes to random draw order or force accumulation order can change trajectories; compare like-for-like seeds and settings when evaluating regressions. Undo restores configuration while retaining surviving motion, not historical particle positions.

For preset changes, preserve compatibility deliberately. Validate imports before replacing controller state, keep reads and writes within the same size budget, and retain the temporary-file write path. Add round-trip and failed-save preservation checks when changing the format.

## Verification

Run the ordinary checks after code changes:

```sh
gofmt -w main.go internal images
go test ./...
go vet ./...
go test -race ./internal/simulation ./internal/experiment
go build -o /tmp/genesis .
```

For audio changes, also run `go test -race ./internal/soundtrack` and the opt-in device and Sound-panel checks documented in README. Generate the water-drop cue with `python3 audio/generate_water_drop.py`; keep audition renders out of version control via `audio/.gitignore`.

If the default Go build cache is read-only in a restricted workspace, prefix commands with `GOCACHE=/tmp/genesis-go-cache`.

The desktop lifecycle tests are opt-in and require a working graphics display. Run them in separate processes because Ebitengine permits only one game lifecycle per process:

```sh
GENESIS_UI_TEST=1 go test ./internal/desktop -run '^TestDesktopWorkflow$' -count=1 -v
GENESIS_DIALOG_TEST=1 go test ./internal/desktop -run '^TestDesktopDialogs$' -count=1 -v
```

The workflow covers relationship locks, live edits, history, population changes, linked controls, scrolling, import/export, and window layout. The dialog test covers save, overwrite, cancellation, and exit flows. A passing ordinary `go test ./...` does not mean these opt-in workflows ran. Screenshot commands and capture options are in [README.md](README.md#verification).

For physics or live-edit performance changes, compare repeated samples on the same machine without competing tests or builds:

```sh
go test ./internal/simulation -run '^$' -bench 'Benchmark(Physics|PhysicsScaling|RelationshipEdit)$' -benchmem -count=3
```

Physics measures a fixed evolution window. Relationship edits compare zero and 5,000 particles. Neither benchmark measures end-to-end desktop frame time. Use the separate paused-renderer measurement for rendering changes:

```sh
GENESIS_PERF_TEST=1 go test ./internal/desktop -run '^TestDesktopPerformance$' -count=1 -v
```

For GPU changes, run `TestGPUPhysics` and `TestDesktopGPU` using README's opt-in commands, separately on each available graphics device. Include transfers and readback in timings. Cover self/coincident particles, wrapped grids, half-world ties, maximum grid size, speed/retention boundaries, all 12 types, and live edits. GPU comparisons use explicit tolerances; CPU parallelism must still match serial results exactly. Failed, incomplete, or invalid acceleration results must never partially update live particles.

Prefer regression tests for observable behavior: invalid edits preserve state, world resizing preserves velocity, type reordering preserves identity, wrapped searches include every nearby pair exactly once, and failed saves preserve the readable preset. Avoid tests that only duplicate implementation details or formatting.
