#!/usr/bin/env python3
"""Extract a release archive and verify that its executable renders a frame."""

import argparse
import hashlib
from pathlib import Path
import shutil
import struct
import subprocess
import tempfile


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("archive", type=Path)
    args = parser.parse_args()
    archive = args.archive.resolve()
    expected = archive.with_name(archive.name + ".sha256").read_text().split()[0]
    if hashlib.sha256(archive.read_bytes()).hexdigest() != expected:
        raise RuntimeError("Archive checksum does not match")
    with tempfile.TemporaryDirectory(prefix="genesis-smoke-") as temporary:
        directory = Path(temporary)
        shutil.unpack_archive(archive, directory)
        binaries = [path for path in directory.rglob("*") if path.is_file()
                    and path.name in ("Genesis", "Genesis.exe", "genesis")]
        if len(binaries) != 1:
            raise RuntimeError(f"Expected one executable, found {len(binaries)}")
        binary = binaries[0]
        binary.chmod(0o755)  # Python's ZIP extractor does not restore Unix modes.
        screenshot = directory / "smoke.png"
        subprocess.run([
            str(binary), "-mute", "-seed", "42", "-skip-intro",
            "-width", "1024", "-height", "768", "-capture-frames", "3",
            "-screenshot", str(screenshot),
        ], cwd=directory, check=True, timeout=60)
        # The application currently logs screenshot-write errors without a failing
        # exit status, so successful exit alone is insufficient.
        png = screenshot.read_bytes()
        if png[:8] != b"\x89PNG\r\n\x1a\n" or struct.unpack(">II", png[16:24]) != (1024, 768):
            raise RuntimeError("Application did not produce the expected screenshot")
        print(f"Verified checksum, extraction, startup, and rendering: {archive.name}")


if __name__ == "__main__":
    main()
