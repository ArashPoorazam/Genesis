Genesis — Particle laboratory
https://github.com/ArashPoorazam/Genesis

No Go installation or internet connection is needed to run Genesis.
Extract the entire archive before starting the application.

Windows (x64)
-------------
Open Genesis.exe. This portable build does not need an installer.
The executable is not publisher-signed; Windows may show a reputation warning.
Only run downloads from a source you trust.

macOS (Apple Silicon or Intel)
-----------------------------
Choose arm64 for Apple Silicon, or x64 for Intel Macs.
Drag Genesis.app to Applications, then open it.
The bundle is ad-hoc signed, but is not Developer ID signed or notarized.
If macOS blocks opening it, follow Apple's instructions for an app you trust:
https://support.apple.com/guide/mac-help/open-a-mac-app-from-an-unknown-developer-mh40616/mac
Do not disable Gatekeeper system-wide.

Linux (x64)
-----------
Open a terminal in the extracted folder and run ./genesis.
If your archive manager lost the executable permission, run chmod +x genesis.
A working graphics driver and X11 or XWayland desktop session are required.
Core runtime libraries: libX11 and OpenGL (libGL/libGLX).
Recommended X11 libraries: libXcursor, libXi, libXinerama, libXrandr,
libXrender, and libXext.
Audio uses PulseAudio or PipeWire's PulseAudio service, with an ALSA fallback.
Run ./genesis -mute to start without opening an audio device.
Distribution-specific library instructions:
https://ebitengine.org/en/documents/install.html

Getting started
---------------
Press any key or click to enter the laboratory. Space plays or pauses.
Use the edge handle to open the inspector. F1 opens the keyboard guide.
GPU boost is experimental and off by default.
Presets and audio preferences are saved in your user configuration directory,
not beside the executable. Replacing this application preserves those files.

Checksums and licenses
---------------------
SHA256SUMS is published beside the downloads on GitHub Releases.
On Linux: sha256sum --ignore-missing -c SHA256SUMS
On macOS: shasum -a 256 Genesis-*.zip (compare with SHA256SUMS)
On Windows PowerShell: Get-FileHash .\Genesis-*.zip -Algorithm SHA256
Genesis's license is in LICENSE. Dependency and font licenses are in licenses/
(inside Genesis.app/Contents/Resources/ on macOS).
