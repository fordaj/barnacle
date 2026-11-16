# The build script (`scripts/build.sh`)

```mermaid
flowchart LR
    Version["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M5 8l4.5 4L5 16'/><path d='M12 16.5h7'/></svg><svg class='label-icon' viewBox='0 0 24 24' width='12' height='12' style='position:absolute;right:-31px;top:-14px' fill='none' stroke='currentColor' stroke-width='2.4' stroke-linecap='round' stroke-linejoin='round'><path d='M10 14a5 5 0 0 0 7.07 0L20.6 10.4a5 5 0 0 0-7.07-7.07l-2 2'/><path d='M14 10a5 5 0 0 0-7.07 0L3.4 13.6a5 5 0 0 0 7.07 7.07l2-2'/></svg>version.sh</span>"] --> Stamp["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M8.5 7 4 12l4.5 5'/><path d='M15.5 7 20 12l-4.5 5'/></svg>go build -ldflags -X main.version</span>"]
    Platforms[6 targets<br/>darwin · linux · windows<br/>arm64 · amd64 — CGO off] --> Stamp
    Stamp --> Dist["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M3.5 5.7A2.2 2.2 0 0 1 5.7 3.5h4L12 6h5.3a2.2 2.2 0 0 1 2.2 2.2v8.6a2.2 2.2 0 0 1-2.2 2.2H5.7a2.2 2.2 0 0 1-2.2-2.2z'/></svg>dist/&lt;platform&gt;/<br/>barnacle{,.exe}</span>"]
    Dist --> Release["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M5 8l4.5 4L5 16'/><path d='M12 16.5h7'/></svg><svg class='label-icon' viewBox='0 0 24 24' width='12' height='12' style='position:absolute;right:-31px;top:-14px' fill='none' stroke='currentColor' stroke-width='2.4' stroke-linecap='round' stroke-linejoin='round'><path d='M10 14a5 5 0 0 0 7.07 0L20.6 10.4a5 5 0 0 0-7.07-7.07l-2 2'/><path d='M14 10a5 5 0 0 0-7.07 0L3.4 13.6a5 5 0 0 0 7.07 7.07l2-2'/></svg>release.sh — archives</span>"]
    click Release href "release.sh.md"
    click Version href "version.sh.md"
    class Version,Release shFile
    class Stamp goFile
    class Dist folder
    classDef shFile fill:#abe07a,stroke:#3f6212,color:#16280a
    classDef goFile fill:#8fd8ef,stroke:#0277a8,color:#08303d
    classDef folder fill:#f0c558,stroke:#8a6d1c,color:#332a05
```

Cross-compiles barnacle for every supported platform into `dist/<platform>/barnacle` (`barnacle.exe` on Windows), stamping the version computed by versioner into each binary.

## What it does

1. Sources [entry.sh.md](entry.sh.md) and changes into the repository root (`REPO_ROOT_DIR`).
2. Resolves the version: `$VERSION` if set, otherwise `eval "$(scripts/version.sh --format shell)"`.
3. Fails early if the `go` toolchain is not on `PATH`.
4. Builds six targets with `CGO_ENABLED=0`, `-trimpath`, and `-ldflags "-s -w -X main.version=$VERSION"`: `macos-arm64`, `macos-x86_64`, `linux-arm64`, `linux-x86_64`, `windows-arm64`, `windows-x86_64` (GOOS/GOARCH `darwin`/`linux`/`windows` × `arm64`/`amd64`).
5. Writes each binary to `dist/<platform>/` and logs progress through `log()`.

`dist/` is git-ignored and always excluded from the scan.

## Usage

```bash
scripts/build.sh            # build all six platforms into dist/
VERSION=1.2.3.4 scripts/build.sh   # override the stamped version
```

## Consumers

[release.sh.md](release.sh.md) calls `build.sh` with `VERSION` set, then packages the resulting per-platform directories into the release archives. The README's "Building from source" section documents it for one-off local builds.

## Related pages

- [version.sh.md](version.sh.md) — where the stamped version comes from
- [release.sh.md](release.sh.md) — the release pipeline built on this script
- [entry.sh.md](entry.sh.md) · [env.sh.md](env.sh.md)
- [README.md](README.md) — the scripts hub
