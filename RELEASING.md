# Releasing Genesis

The [Desktop releases workflow](.github/workflows/release.yml) builds Windows x64, Linux x64, macOS Intel, and macOS Apple Silicon packages. Each target runs on a matching GitHub-hosted runner, using the Go version declared in `go.mod` and Python 3.13 for packaging. No external Python packages are required.

## Try the workflow before publishing

Commit and push the workflow, `scripts/`, `packaging/`, and documentation to GitHub. Pull requests and branch pushes build and check all four packages without publishing a release. You can also choose **Actions → Desktop releases → Run workflow** after the workflow reaches the default branch.

Download the `release-*` artifacts from a successful workflow run. These contain the actual distribution archives and individual `.sha256` files. Development archives use `v0.0.0-dev.COMMIT` in their names. Actions artifacts expire after 14 days; published GitHub Release assets are the permanent download location.

Every target runs ordinary Go tests, vet, race checks for simulation and experiment, and a smoke test of the extracted package. The smoke test launches the real executable with audio muted, renders a 1024 × 768 screenshot, checks its PNG header and dimensions, then exits. Linux uses Xvfb and software OpenGL. This does not replace testing real graphics hardware, audio, or the experimental GPU mode. macOS packaging also verifies the ad-hoc code signature.

Before the first public release, try the downloads on real Windows, Intel Mac, Apple Silicon Mac, and Linux desktops. Check opening the downloaded app, sound, presets, and window resizing. GitHub runners cannot verify the download/quarantine experience on a user's machine.

## Publish a version

Once the intended source changes and release workflow are committed and pushed, create a new version tag on that commit:

```sh
git tag -a v0.1.0 -m "Genesis v0.1.0"
git push origin v0.1.0
```

Use `vMAJOR.MINOR.PATCH`, or a prerelease such as `v0.1.0-rc.1`. Do not reuse or move published version tags. A tag push triggers the same build checks. Only after all four targets pass does the publish job collect and verify the archives, create a draft GitHub Release, upload all four packages and `SHA256SUMS`, and publish the release. Tags with a suffix are marked as prereleases; stable versions are marked latest. The release notes are generated from GitHub's change history and can be edited afterward.

The workflow uses the repository's built-in `GITHUB_TOKEN`; no personal token is needed. Build jobs have read-only access. Only the tag-triggered publish job receives `contents: write`. If repository or organization policy disables Actions or release writes, an administrator must enable them.

If a build fails, no release is published. Fix the issue and use a new tag. If uploading fails after a draft is created, rerun the failed publish job: it can complete the draft. It deliberately refuses to overwrite an already published release. A manual workflow run, including one on a tag, only creates test artifacts.

The public download link is:

https://github.com/ArashPoorazam/Genesis/releases

## Build packages locally

Use Go from `go.mod` and Python 3.11 or later. From the repository root:

```sh
python3 scripts/package_release.py --version v0.1.0 --os linux --arch amd64
python3 scripts/package_release.py --version v0.1.0 --os windows --arch amd64
```

On macOS, build the desired architecture with:

```sh
python3 scripts/package_release.py --version v0.1.0 --os darwin --arch arm64
python3 scripts/package_release.py --version v0.1.0 --os darwin --arch amd64
```

On Windows use `python` instead of `python3`. Linux and Windows executables can be cross-compiled. macOS packaging requires macOS because it uses `sips`, `iconutil`, and `codesign` to create and verify the app bundle. Packages and individual checksums go into the ignored `dist/` directory; use `--output PATH` to choose another directory. No generated executable is committed.

On the matching operating system, check an archive with:

```sh
python3 scripts/smoke_release.py dist/Genesis-v0.1.0-linux-x64.tar.gz
```

Use the corresponding `.zip` path on macOS or Windows. The adjacent `.sha256` file must be present. A desktop display is required; on Linux CI, prefix the command with `LIBGL_ALWAYS_SOFTWARE=1 xvfb-run -a`.

## Signing and platform coverage

Windows ZIPs contain a GUI executable, so launching normally does not open a console window. They are not Authenticode-signed and may trigger Windows reputation warnings.

macOS ZIPs contain a `Genesis.app` bundle with version metadata, a generated icon, embedded licenses, and an ad-hoc signature. Ad-hoc signing does **not** establish a verified publisher or satisfy Gatekeeper notarization. Developer ID signing and notarization require the maintainer's Apple credentials and are not configured. To add them later, replace the ad-hoc signing stage with Developer ID signing, notarize and staple the app, and only then archive it and generate checksums.

Linux packages contain a portable executable, instructions, and licenses; they are not AppImages or distribution packages. They still use system graphics and audio services. ARM Linux, ARM Windows, and 32-bit platforms are not part of the release matrix.

Builds embed graphics, fonts, music, and effects. Packages include Genesis's license, the font license, Go's license, and top-level license/notice files for the Go modules used by that target.
