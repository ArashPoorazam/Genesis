#!/usr/bin/env python3
"""Render the handpan edition of Genesis's score and cues. Requires NumPy and ffmpeg.

Run with /usr/bin/python3 audio/generate_audio.py. No downloaded samples.
"""

from pathlib import Path
import json
import subprocess
import wave

import numpy as np


ROOT = Path(__file__).resolve().parent / 'handpan'
RATE = 44100
TEMPO = 72
BEAT = 60 / TEMPO
LENGTH = 32 * 4 * BEAT
RNG = np.random.default_rng(4207)


def hz(note):
    return 440 * 2 ** ((note - 69) / 12)


def stereo(signal, pan):
    angle = (pan + 1) * np.pi / 4
    return np.column_stack((signal * np.cos(angle), signal * np.sin(angle)))


def handpan(note, duration=4.4, softness=1):
    """Synthesize a rounded hand strike and a resonant, tuned metal body."""
    t = np.arange(int(duration * RATE)) / RATE
    frequency = hz(note)
    result = np.zeros_like(t)
    for partial, amplitude, decay in [(1, 1, 1.65), (2, .58, .95), (3, .34, .60),
                                      (4.76, .095, .085), (6.31, .045, .045)]:
        phase = 2 * np.pi * frequency * partial * t
        # A brief pitch relaxation and gentle beating give the metal some motion.
        phase += .20 * (1 - np.exp(-t / .018))
        beating = .88 + .12 * np.cos(2 * np.pi * 1.7 * t)
        result += amplitude * np.sin(phase) * np.exp(-t / (decay * softness)) * beating
    # A soft, low palm tap and short filtered contact noise distinguish the attack.
    result += .24 * np.sin(2 * np.pi * 145 * t) * np.exp(-t / .022)
    noise = np.random.default_rng(1000 + int(note)).standard_normal(len(t))
    noise = np.convolve(noise, np.ones(23) / 23, mode='same')
    result += .13 * noise * np.exp(-t / .012)
    result *= (1 - np.exp(-t / .0018))
    result *= np.minimum(1, (duration - t) / .15)
    return result * .31


def pad(note, duration):
    t = np.arange(int(duration * RATE)) / RATE
    f = hz(note)
    attack = np.minimum(t / 1.1, 1)
    release = np.minimum((duration - t) / 2.1, 1)
    envelope = np.sin(np.pi / 2 * np.minimum(attack, release)) ** 2
    result = (np.sin(2 * np.pi * f * t + .07 * np.sin(2 * np.pi * .21 * t))
              + .23 * np.sin(2 * np.pi * f * 2.001 * t)
              + .08 * np.sin(2 * np.pi * f * 3 * t))
    return result * envelope * .28


def place(target, signal, start, level=1, pan=0, wrap=False):
    signal = stereo(signal, pan) * level if signal.ndim == 1 else signal * level
    offset = int(round(start * RATE))
    if wrap:
        offset %= len(target)
        first = min(len(signal), len(target) - offset)
        target[offset:offset + first] += signal[:first]
        if first < len(signal):
            target[:len(signal) - first] += signal[first:]
    else:
        count = min(len(signal), len(target) - offset)
        if count > 0:
            target[offset:offset + count] += signal[:count]


def space(dry, circular=False, amount=.22):
    # A diffuse stereo echo field; smoothing keeps the tail behind the hand strikes.
    filtered = dry.copy()
    for shift in (1, 3, 7, 13, 29):
        filtered = (filtered + np.roll(filtered, shift, axis=0)) * .5
    wet = np.zeros_like(dry)
    for index, delay in enumerate((.071, .113, .173, .229, .307, .389, .467,
                                   .571, .683, .809, .947, 1.103, 1.307, 1.541, 1.809, 2.119)):
        count = int(delay * RATE)
        source = filtered[:, ::-1] if index % 2 else filtered
        gain = np.exp(-delay / .65) * amount
        if circular:
            wet += np.roll(source, count, axis=0) * gain
        elif count < len(dry):
            wet[count:] += source[:-count] * gain
    return dry + wet


def export(name, data, peak=None):
    path = ROOT / name
    path.parent.mkdir(parents=True, exist_ok=True)
    if peak is not None:
        data = data * (peak / max(np.max(np.abs(data)), 1e-9))
    assert np.isfinite(data).all() and np.max(np.abs(data)) < 1
    with wave.open(str(path), 'wb') as output:
        output.setparams((2, 2, RATE, 0, 'NONE', 'not compressed'))
        output.writeframes((data * 32767).round().astype('<i2').tobytes())
    return data


def encode(source, destination, codec):
    subprocess.run(['ffmpeg', '-v', 'error', '-y', '-i', str(ROOT / source),
                    '-c:a', codec, '-q:a', '5' if codec == 'libvorbis' else '2',
                    str(ROOT / destination)], check=True)


def score():
    dry = np.zeros((round(LENGTH * RATE), 2), dtype=np.float64)
    # Two-bar voicings: Dmaj9, Aadd9/C#, Bm9, Gmaj9, Em9, D/F#, Gmaj9, Asus.
    progression = [
        (38, [57, 61, 64, 66, 69]), (37, [57, 59, 64, 69, 73]),
        (35, [54, 57, 61, 62, 66]), (31, [54, 57, 59, 62, 66]),
        (40, [55, 59, 62, 66, 69]), (42, [57, 62, 64, 66, 69]),
        (31, [54, 57, 59, 62, 66]), (33, [55, 59, 62, 64, 69]),
    ]
    motifs = [
        [(1, 78), (2.5, 76), (5, 73), (6.5, 69)],
        [(1.5, 71), (3, 73), (5.5, 76)],
        [(0.5, 73), (2, 74), (4.5, 78), (6.5, 76)],
        [(1, 74), (3.5, 71), (6, 69)],
        [(1.5, 71), (3, 74), (5.5, 78)],
        [(1, 76), (2.5, 74), (5, 73), (7, 69)],
        [(1.5, 71), (4, 74), (6, 73)],
        [(1, 71), (3.5, 69), (6.5, 76)],
    ]
    pattern = [0, 2, 1, 3, 2, 4]
    positions = [0, .75, 1.5, 2, 2.75, 3.5]
    for phrase in range(16):
        bass, chord = progression[phrase % 8]
        start = phrase * 8 * BEAT
        for voice, note in enumerate(chord[:4]):
            place(dry, pad(note - 12, 8 * BEAT + 2.3), start,
                  .065, (voice - 1.5) * .36, True)
        for beat in (0, 4):
            place(dry, handpan(bass + 12, 4.8, 1.35), start + beat * BEAT,
                  .28, -.06, True)
        for bar in range(2):
            for step, beat in enumerate(positions):
                note = chord[pattern[(step + (bar if phrase >= 8 else 0)) % 6]]
                velocity = .23 * (1 + RNG.uniform(-.12, .12))
                place(dry, handpan(note, 3.7), start + (bar * 4 + beat) * BEAT + RNG.uniform(.002, .018),
                      velocity, -.30 + .11 * (step % 5), True)
        # Keep the melody in a warmer register for the handpan edition.
        if phrase not in (0, 7):
            for beat, note in motifs[phrase % 8]:
                place(dry, handpan(note - 12, 5, 1.2), start + beat * BEAT + .025,
                      .26 if phrase < 8 else .29, .19, True)
    music = space(dry, circular=True, amount=.13)
    # One consistent master gain: preserve the musical phrasing and dynamics.
    music *= .63 / np.max(np.abs(music))
    return music


def cue(notes, spacing, duration, level=.38):
    result = np.zeros((round(duration * RATE), 2))
    for index, note in enumerate(notes):
        note = note - 12 if note >= 72 else note
        place(result, handpan(note, min(1.8, duration), .48), index * spacing,
              level, (index - (len(notes) - 1) / 2) * .18)
    result = space(result, amount=.12)
    fade = min(round(.18 * RATE), len(result))
    result[-fade:] *= np.linspace(1, 0, fade)[:, None]
    return result


def runtime_cues():
    # Use tones shared with the next chord, so a cue's tail can cross a change.
    chords = [{2, 6, 9, 1, 4}, {9, 11, 4, 1}, {11, 2, 6, 9, 1},
              {7, 11, 2, 6, 9}, {4, 7, 11, 2, 6}, {2, 6, 9, 4},
              {7, 11, 2, 6, 9}, {9, 7, 11, 2, 4}]
    definitions = {'select': ([66], 0, .55, .22),
                   'play': ([59], 0, .70, .18),
                   'pause': ([55], 0, .65, .16),
                   'save': ([59], 0, .90, .19),
                   'new-world': ([50, 57, 62, 66, 69, 64], .11, 2.6, .27)}
    for index, chord in enumerate(chords):
        shared = chord & chords[(index + 1) % len(chords)]
        candidates = [note for note in range(48, 72) if note % 12 in shared]
        for name, (notes, spacing, duration, level) in definitions.items():
            tuned = [min(candidates, key=lambda candidate: (abs(candidate - note), candidate)) for note in notes]
            export(f'runtime/{index}/{name}.wav', cue(tuned, spacing, duration, level))


def main():
    runtime_cues()
    music = export('music/quiet-orbits-loop.wav', score())
    encode('music/quiet-orbits-loop.wav', 'music/quiet-orbits-loop.ogg', 'libvorbis')
    listening = music.copy()
    fade = int(2.5 * RATE)
    listening[:fade] *= np.linspace(0, 1, fade)[:, None]
    listening[-fade:] *= np.linspace(1, 0, fade)[:, None]
    export('preview/quiet-orbits.wav', listening)
    encode('preview/quiet-orbits.wav', 'preview/quiet-orbits.mp3', 'libmp3lame')
    sounds = {
        'select': cue([78], 0, .55, .22),
        # Single, quieter strikes avoid stacking rapid notes over the melody.
        # B fits the Aadd9/Bm9 transition at 13s and Em9 at 29s; G fits 21s.
        'play': cue([59], 0, .70, .18),
        'pause': cue([55], 0, .65, .16),
        'save': cue([59], 0, .90, .19),
        'new-world': cue([50, 57, 62, 66, 69, 76], .11, 2.6, .27),
    }
    reel = np.zeros((15 * RATE, 2))
    combined = listening[:round(45 * RATE)].copy()
    timeline = []
    for index, (name, sound) in enumerate(sounds.items()):
        export(f'sfx/{name}.wav', sound)
        start = .6 + index * 3
        place(reel, sound, start)
        mix_start = 5 + index * 8
        if name in ('play', 'pause', 'save'):
            # Give these cues space in the audition mix without editing the score.
            attack, release = round(.06 * RATE), round(.20 * RATE)
            offset = round(mix_start * RATE) - attack
            envelope = np.concatenate((np.linspace(1, .72, attack, endpoint=False),
                                       np.full(len(sound), .72),
                                       np.linspace(.72, 1, release)))
            combined[offset:offset + len(envelope)] *= envelope[:, None]
        place(combined, sound, mix_start, .8)
        timeline.append({'sound': name, 'reel_seconds': start, 'mix_seconds': 5 + index * 8})
    combined[-2 * RATE:] *= np.linspace(1, 0, 2 * RATE)[:, None]
    export('preview/sound-effects.wav', reel)
    encode('preview/sound-effects.wav', 'preview/sound-effects.mp3', 'libmp3lame')
    export('preview/genesis-audio-preview.wav', combined)
    encode('preview/genesis-audio-preview.wav', 'preview/genesis-audio-preview.mp3', 'libmp3lame')
    metrics = {'title': 'Quiet Orbits — Handpan', 'instrument': 'synthesized handpan',
               'bpm': TEMPO, 'duration_seconds': LENGTH,
               'sample_rate': RATE, 'channels': 2, 'peak_dbfs': float(20 * np.log10(np.max(np.abs(music)))),
               'rms_dbfs': float(20 * np.log10(np.sqrt(np.mean(music ** 2)))),
               'loop_boundary_delta': float(np.max(np.abs(music[0] - music[-1]))),
               'cue_timeline': timeline,
               'preview_revision': 2,
               'preview_changes': 'Play, pause, and save use quieter single strikes with a brief music dip; full score unchanged.'}
    (ROOT / 'preview' / 'manifest.json').write_text(json.dumps(metrics, indent=2) + '\n')
    print(json.dumps(metrics, indent=2))


if __name__ == '__main__':
    main()
