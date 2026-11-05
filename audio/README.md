# Current app audio

The app uses a synthesized lake water-drop cue for play/pause, restart/new world, successful preset saves and loads, shuffling relationships, and successful population updates. Routine buttons, tabs, and edits are silent. Background music defaults to off and can be enabled in Settings → Sound. Regenerate the cue with `python3 audio/generate_water_drop.py`; it requires only the Python standard library. The handpan music remains unchanged. Older handpan cue banks below are retained as audition material and are no longer embedded.

# Genesis audio draft — Quiet Orbits

The current version is the **handpan edition** in `handpan/`. It replaces the keyboard with synthesized hand strikes, a rounded metal resonance, and a softer ambient backing. The melody is voiced an octave lower and the bass an octave higher. Tempo, harmony, and the musical phrases are preserved. All five interaction cues now use the same handpan sound.

Listen to `handpan/preview/genesis-audio-preview.mp3` for the combined preview or `handpan/preview/quiet-orbits.mp3` for the complete track. Individual cues are in `handpan/sfx/`, and the looping masters are in `handpan/music/`. This is a synthesized approximation, not a sampled acoustic instrument.

Handpan preview revision 2 replaces the layered play, pause, and save chimes at 13, 21, and 29 seconds with quieter single strikes that fit the harmony at those moments. The preview briefly lowers the backing under these three cues. The complete soundtrack and looping masters are unchanged. These cue pitches are arranged for the preview; future runtime integration should adapt pitched cues to the active harmony.

The original keyboard edition remains in the top-level `music/`, `preview/`, and `sfx/` folders for comparison; the descriptions below refer to that original edition.

An original instrumental composition and synthesized interaction cues created for Genesis. All tones are synthesized; no recordings, samples, or melodies were taken from 3Blue1Brown or another soundtrack.

The musical direction is quiet, curious, and spacious: mellow electric-piano tones, a repeating arpeggio, sparse melody, and warm sustained harmony. The score is 72 BPM, 32 bars, and 106.667 seconds long.

## Listen

- `preview/genesis-audio-preview.mp3`: 45-second music-and-effects demonstration. Cues occur at 5, 13, 21, 29, and 37 seconds.
- `preview/quiet-orbits.mp3`: the complete soundtrack with listening fades.
- `preview/sound-effects.mp3`: isolated cues in this order: select, play, pause, save, new world. They start at 0.6, 3.6, 6.6, 9.6, and 12.6 seconds.

## App assets

- `music/quiet-orbits-loop.ogg`: compact stereo loop for background playback.
- `music/quiet-orbits-loop.wav`: lossless stereo master with the same loop boundaries.
- `sfx/*.wav`: five individual stereo cues.

The loop includes the previous cycle's note and reverb tails at its start. The MP3 preview has fades and should not be used as the looping asset. All assets are 44.1 kHz; WAV files are 16-bit PCM.

Genesis now embeds the approved handpan loop and the harmony-matched cue banks in `handpan/runtime/`. Music and effects have separate controls in Settings → Sound; M toggles global mute. The runtime fades volume changes, lowers music under cues, and prevents overlapping cues. Cues use tones shared by the current and next chord so their tails remain compatible with chord changes. Continuous collision sounds are deliberately omitted to keep the mix calm. The original preview files remain available for comparison.

## Reproduce

Run `/usr/bin/python3 audio/generate_audio.py` from the repository root with NumPy and ffmpeg available to regenerate the current handpan edition. The fixed random seed preserves the small timing and velocity variations. `handpan/preview/manifest.json` records duration, levels, the loop boundary measurement, and cue timings. The original edition's generator is retained inside `genesis-audio-pack.zip`.
