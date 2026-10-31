# Genesis

This is a simulation that I use to Exercise and understand the "Emergence" concept

Run with `go run .` or build with `go build .`.

At startup, `main` reads the current monitor size into the global `screenWidth` and `screenHeight` variables before creating particles and the grid. The window, logical layout, spawning, grid, and physics wrapping all use these dimensions, in device-independent pixels.

## Particle performance

- Physics uses a reusable spatial grid to check nearby particles, including neighbors across the wrapped screen edges. Cell sizes cover the configured maximum outer radius, or any larger radius present in the particle types.
- Distance checks reject particles outside the influence radius before computing a square root. Particles store velocity components directly.
- Each particle type's circle and halo are rendered into a cached sprite and reused each frame.
- Spawning uses the grid for collision checks and stops with a log message after 1,000 failed attempts to place a particle.
