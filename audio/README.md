# Genesis audio

Genesis uses **Quiet Orbits**, an original synthesized handpan score, and a short synthesized lake water-drop interaction cue. No downloaded recordings or samples are required.

Background music defaults to off. The water-drop cue accompanies play/pause, restart/new world, successful preset saves and loads, shuffling relationships, and successful population updates. Routine buttons, tabs, and edits are silent. Settings → Sound controls music and effects separately; M toggles global mute. Cues briefly lower enabled music and never overlap.

## Embedded assets

`audio/assets.go` embeds only these two files:

- `handpan/music/quiet-orbits-loop.ogg`: the seamless stereo background loop, 72 BPM, 32 bars, approximately 106.667 seconds.
- `sfx/water-drop.wav`: the short stereo interaction cue, 44.1 kHz, 16-bit PCM.

The old harmony-matched handpan interaction banks in `handpan/runtime/` are unused audition material. They are ignored by Git and are not embedded in the app. Generated previews, lossless music masters, and download bundles also stay local.

## Reproduce

Run `python3 audio/generate_water_drop.py` to regenerate the water-drop WAV. It uses only the Python standard library and a fixed random seed.

Run `/usr/bin/python3 audio/generate_audio.py` with NumPy and ffmpeg available to regenerate the handpan music. It also creates local audition material: the old cue banks, a lossless music master, and listening previews under `handpan/preview/`. The preview combines music with legacy handpan cues; it does not demonstrate the app's current water-drop effects. `handpan/preview/manifest.json` records the score's duration, levels, loop boundary measurement, and preview cue timings.

The loop includes the previous cycle's note and reverb tails at its start. Listening previews have fades and must not replace the looping asset. Python, NumPy, and ffmpeg are development tools and are not required to run Genesis.
