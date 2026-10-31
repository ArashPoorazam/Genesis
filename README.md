# Genesis

The name is just another word for beginning of something big like the big bang.
This is a simulation that I use to Exercise and understand the "Emergence" concept, emergence is about how small entity like a particle with a gene(rules) would create complex structure and systems like life.

For now it uses 5 hard coded particle types with their own colors and the computer chooses randomly interactions(Relations) between each particle and the range of their influence.

From what I have seen each reset feels like a whole new world with new physics rules, it is so fascinating to watch.

## Install

For now just clone the repo and run with `go run .` or build with `go build .`
Every time you start, a set of unique rules will be created for each particle and it gets printed on terminal so every new start feels like a new world.

## Fun set of particle types

```
 redParticle = createType(
  "Red",
  color.RGBA{255, 0, 0, 255},
  2,
  80,
  10,
 )
 greenParticle = createType(
  "Green",
  color.RGBA{0, 255, 0, 255},
  2,
  60,
  20,
 )
 blueParticle= createType(
  "Blue",
  color.RGBA{0, 0, 255, 255},
  2,
  100,
  10,
 )
 pinkParticle = createType(
  "Pink",
  color.RGBA{255, 105, 180, 255},
  2,
  30,
  10,
 )
 orangeParticle = createType(
  "Orange",
  color.RGBA{255, 213, 128, 255},
  2,
  120,
  40,
 )
```

## Particle performance

- Physics uses a reusable spatial grid to check nearby particles, including neighbors across the wrapped screen edges. Cell sizes cover the configured maximum outer radius, or any larger radius present in the particle types.
- Distance checks reject particles outside the influence radius before computing a square root. Particles store velocity components directly.
- Each particle type's circle and halo are rendered into a cached sprite and reused each frame.
- Spawning uses the grid for collision checks and stops with a log message after 1,000 failed attempts to place a particle.
