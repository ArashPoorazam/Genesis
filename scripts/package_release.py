#!/usr/bin/env python3
"""Build and package Genesis using Go and the Python standard library."""

import argparse
import hashlib
import os
from pathlib import Path
import plistlib
import re
import shutil
import subprocess
import sys
import tarfile
import tempfile
import zipfile


ROOT = Path(__file__).resolve().parents[1]
TARGETS = {
    ("windows", "amd64"): "windows-x64",
    ("linux", "amd64"): "linux-x64",
    ("darwin", "amd64"): "macos-x64",
    ("darwin", "arm64"): "macos-arm64",
}


def run(*args, **kwargs):
    return subprocess.run(args, cwd=ROOT, check=True, **kwargs)


def copy_licenses(destination, env):
    destination.mkdir(parents=True, exist_ok=True)
    shutil.copy2(ROOT / "internal/desktop/assets/FONT-LICENSE", destination)
    goroot = run("go", "env", "GOROOT", capture_output=True, text=True).stdout.strip()
    shutil.copy2(Path(goroot) / "LICENSE", destination / "GO-LICENSE")
    # Include the notices for modules actually used by this target's executable.
    modules = run(
        "go", "list", "-deps", "-f",
        "{{if .Module}}{{if not .Module.Main}}{{.Module.Path}}\t{{.Module.Dir}}{{end}}{{end}}",
        ".", env=env, capture_output=True, text=True,
    ).stdout.splitlines()
    for module in sorted(set(modules) - {""}):
        name, directory = module.split("\t")
        notices = [path for path in Path(directory).iterdir() if path.is_file()
                   and path.name.upper().startswith(("LICENSE", "COPYING", "NOTICE", "AUTHORS"))]
        if not notices:
            raise RuntimeError(f"No license found for {name}")
        target = destination / name
        target.mkdir(parents=True, exist_ok=True)
        for notice in notices:
            shutil.copy2(notice, target)


def macos_bundle(folder, version):
    app = folder / "Genesis.app"
    contents = app / "Contents"
    (contents / "MacOS").mkdir(parents=True)
    resources = contents / "Resources"
    resources.mkdir()
    # Generate standard Dock/Finder icon sizes from the existing project artwork.
    with tempfile.TemporaryDirectory(prefix="genesis-icon-") as temporary:
        iconset = Path(temporary) / "Genesis.iconset"
        iconset.mkdir()
        for size in (16, 32, 128, 256, 512):
            for scale in (1, 2):
                suffix = "@2x" if scale == 2 else ""
                run("sips", "-z", str(size * scale), str(size * scale),
                    str(ROOT / "images/logo.png"), "--out",
                    str(iconset / f"icon_{size}x{size}{suffix}.png"),
                    stdout=subprocess.DEVNULL)
        run("iconutil", "-c", "icns", str(iconset), "-o", str(resources / "Genesis.icns"))
    with (contents / "Info.plist").open("wb") as plist:
        plistlib.dump({
            "CFBundleName": "Genesis",
            "CFBundleDisplayName": "Genesis",
            "CFBundleIdentifier": "io.github.arashpoorazam.genesis",
            "CFBundleExecutable": "Genesis",
            "CFBundlePackageType": "APPL",
            "CFBundleIconFile": "Genesis.icns",
            "CFBundleShortVersionString": version.removeprefix("v").split("-")[0],
            "CFBundleVersion": version.removeprefix("v").split("-")[0],
            "NSHighResolutionCapable": True,
        }, plist)
    return contents / "MacOS/Genesis"


def package(version, goos, arch, output):
    env = {**os.environ, "GOOS": goos, "GOARCH": arch, "CGO_ENABLED": "0"}
    name = f"Genesis-{version}-{TARGETS[goos, arch]}"
    output.mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory(prefix="genesis-package-") as temporary:
        folder = Path(temporary) / name
        folder.mkdir()
        binary = folder / ("Genesis.exe" if goos == "windows" else "genesis")
        if goos == "darwin":
            binary = macos_bundle(folder, version)
        flags = "-s -w"
        if goos == "windows":
            flags += " -H=windowsgui"
        run("go", "build", "-trimpath", "-ldflags", flags, "-o", str(binary), ".", env=env)
        binary.chmod(0o755)
        shutil.copy2(ROOT / "LICENSE", folder)
        shutil.copy2(ROOT / "packaging/README.txt", folder)
        licenses = folder / "licenses"
        if goos == "darwin":
            licenses = folder / "Genesis.app/Contents/Resources/licenses"
            # Keep legal notices with the app when it is dragged to Applications.
            licenses.mkdir(parents=True)
            shutil.copy2(ROOT / "LICENSE", licenses / "GENESIS-LICENSE")
        copy_licenses(licenses, env)
        if goos == "darwin":
            app = folder / "Genesis.app"
            # Ad-hoc signing makes the bundle structurally valid on Apple Silicon.
            # Developer ID signing and notarization require the maintainer's identity.
            run("codesign", "--force", "--sign", "-", str(app))
            run("codesign", "--verify", "--deep", "--strict", str(app))
        if goos == "linux":
            archive = output / f"{name}.tar.gz"
            with tarfile.open(archive, "w:gz") as tar:
                tar.add(folder, arcname=name)
        else:
            archive = output / f"{name}.zip"
            with zipfile.ZipFile(archive, "w", zipfile.ZIP_DEFLATED) as zipped:
                for path in sorted(folder.rglob("*")):
                    if path.is_file():
                        zipped.write(path, path.relative_to(folder.parent))
    digest = hashlib.sha256(archive.read_bytes()).hexdigest()
    archive.with_name(archive.name + ".sha256").write_text(
        f"{digest}  {archive.name}\n", encoding="utf-8")
    print(archive)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--version", required=True, help="vMAJOR.MINOR.PATCH[-prerelease]")
    parser.add_argument("--os", choices=("windows", "linux", "darwin"), required=True)
    parser.add_argument("--arch", choices=("amd64", "arm64"), required=True)
    parser.add_argument("--output", type=Path, default=ROOT / "dist")
    args = parser.parse_args()
    if not re.fullmatch(r"v(?:0|[1-9]\d*)\.(?:0|[1-9]\d*)\.(?:0|[1-9]\d*)(?:-[0-9A-Za-z]+(?:[.-][0-9A-Za-z]+)*)?", args.version):
        parser.error("version must look like v0.1.0 or v0.1.0-rc.1")
    if (args.os, args.arch) not in TARGETS:
        parser.error("unsupported OS/architecture combination")
    if args.os == "darwin" and sys.platform != "darwin":
        parser.error("macOS packaging requires a Mac for iconutil and codesign")
    package(args.version, args.os, args.arch, args.output.resolve())


if __name__ == "__main__":
    main()
