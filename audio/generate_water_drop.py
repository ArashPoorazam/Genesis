"""Synthesize a short lake-drop cue with a rising bubble and soft splash.

Uses only the Python standard library. Run from any directory to regenerate.
"""

import math
from pathlib import Path
import random
import struct
import wave


def main():
    rate = 44100
    rng = random.Random(42)
    samples = []
    phase = 0.0
    noise = 0.0
    for index in range(int(rate * 0.7)):
        t = index / rate
        phase += 2 * math.pi * (470 + 1350 * (1 - math.exp(-t / 0.045))) / rate
        envelope = (1 - math.exp(-t / 0.002)) * math.exp(-t / 0.065)
        bubble = math.sin(phase) * envelope
        noise = 0.65 * noise + 0.35 * rng.uniform(-1, 1)
        splash = noise * (1 - math.exp(-t / 0.001)) * math.exp(-t / 0.018)
        samples.append(0.48 * bubble + 0.11 * splash)
    stereo = bytearray()
    for index, sample in enumerate(samples):
        for delay in (1, 1.13):
            value = sample
            for seconds, gain in ((0.043, 0.16), (0.079, 0.09), (0.131, 0.04)):
                offset = index - int(rate * seconds * delay)
                if offset >= 0:
                    value += samples[offset] * gain
            stereo.extend(struct.pack('<h', round(max(-1, min(1, value)) * 32767)))
    path = Path(__file__).parent / 'sfx' / 'water-drop.wav'
    path.parent.mkdir(parents=True, exist_ok=True)
    with wave.open(str(path), 'wb') as output:
        output.setnchannels(2)
        output.setsampwidth(2)
        output.setframerate(rate)
        output.writeframes(stereo)


if __name__ == '__main__':
    main()
