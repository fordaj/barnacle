# The formattington-lite bundle (`tools/formattington-lite/`)

```mermaid
flowchart LR
    Gh("<img src='../../tools/lucide/icons/tag.svg' width='20' alt=''/> fordaj/* release") --> Resolve("<img src='../../tools/lucide/icons/terminal.svg' width='20' alt=''/> resolve tag")
    Resolve --> Download("<img src='../../tools/lucide/icons/download.svg' width='20' alt=''/> gh release download")
    Download --> Check{{"<img src='../../tools/lucide/icons/hexagon.svg' width='20' alt=''/> checksums present?"}}
    Check -->|yes| Verify("<img src='../../tools/lucide/icons/check.svg' width='20' alt=''/> verify — mismatch is fatal")
    Check -->|no| Warn(warn and continue)
    Verify --> Stage("<img src='../../tools/lucide/icons/folder.svg' width='20' alt=''/> stage in temp dir")
    Warn --> Stage
    Stage --> Swap("<img src='../../tools/lucide/icons/folder.svg' width='20' alt=''/> atomic swap · delete superseded v*")
    class Gh tagNode
    class Resolve shFile
    class Download dlNode
    class Check hexNode
    class Verify checkNode
    class Stage,Swap folder
    classDef tagNode fill:#f6b8d4,stroke:#be185d,color:#500724
    classDef shFile fill:#abe07a,stroke:#3f6212,color:#16280a
    classDef dlNode fill:#93e2e2,stroke:#0f766e,color:#042f2e
    classDef hexNode fill:#ffe08a,stroke:#a16207,color:#422006
    classDef checkNode fill:#9fe8c9,stroke:#0f766e,color:#052e22
    classDef folder fill:#f0c558,stroke:#8a6d1c,color:#332a05
```

Pinned [formattington](https://github.com/fordaj/formattington) `formattington-lite` binaries, refreshed from `fordaj/formattington` releases. Note that nothing in this repository's scripts consumes this bundle — `./scripts/format.sh` and CI (`./ci/format_check.sh`) run formattington through `uv` (pinned commit in `pyproject.toml`/`uv.lock`), not through this directory. The bundle is kept ready for direct use.

## Layout

```text
tools/formattington-lite/
  update.sh                     # refresh from fordaj/formattington releases
  v0.0.2.0.53/
    <platform>/formattington{,.exe}
    <platform>/NOTICE.md
    <platform>/LICENSES/*.txt
```

`<platform>` is one of `macos-arm64`, `macos-x86_64`, `linux-arm64`, `linux-x86_64`, `windows-arm64`, `windows-x86_64`. Unlike the other bundles, this one keeps `NOTICE.md` and the `LICENSES/` directory alongside the binary (attribution for the bundled formatters).

## `update.sh`

Refreshes from `fordaj/formattington`: consumes only assets whose stem ends in `-lite` (e.g. `formattington-<platform>-lite.tar.gz`/`.zip`; the `-full` variant is rejected), mapped onto the bundle layout (`darwin-amd64` → `macos-x86_64`, etc., normalizing `amd64`/`x64` → `x86_64`). Verified against the release's `checksums.txt`/`SHA256SUMS.txt` when present (mismatch is fatal; missing checksums warn and install anyway). Legacy `.7z` assets are accepted with `7z`/`7zz`. Only the binary, `NOTICE.md`, and `LICENSES/` are kept; the swap is staged atomically and superseded `v*/` directories are deleted only afterwards. `--dry-run`, `--tag`, `--repo`; env `FORMATTINGTON_REPO`, `FORMATTINGTON_TAG`. Logging via `scripts/entry.sh`.

## Related pages

- [../README.md](../README.md) — the bundle index
- [../../scripts/format.sh.md](../../scripts/format.sh.md) — how formattington is actually run in this repository
- [../../scripts/env.sh.md](../../scripts/env.sh.md) — the cascade (which does not select this bundle)
