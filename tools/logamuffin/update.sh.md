# The logamuffin bundle (`tools/logamuffin/`)

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

[Logamuffin](https://github.com/fordaj/logamuffin) is the structured logger used by this repository's scripts: timestamped, leveled `log <level> <message>` records. The scripts' `log()` wrapper ([scripts/entry.sh.md](../../scripts/entry.sh.md)) forwards to the pinned binary when it is present and falls back to an equivalent `printf` implementation when it is not, so a fresh clone works without it.

The scripts themselves document the contract (`log INFO "..."`), the `LOG_LEVEL`/`LOGAMUFFIN_LEVEL` filter, and the `LOG_SOURCE` caller-context field that `entry.sh` injects.

## Bundled platforms

`v0.1.3.0.35/` — `macos-arm64`, `macos-x86_64`, `linux-arm64`, `linux-x86_64`, `windows-arm64`, `windows-x86_64` (selected via `LOGAMUFFIN_BIN` from the cascade).

## `update.sh`

Refreshes from `fordaj/logamuffin` releases: archives whose stems carry the `<os>-<arch>` suffix (any single separator; `darwin-*`/`amd64`/`x64` are normalized onto the bundle layout), verified against the release's `checksums.txt`/`SHA256SUMS.txt` when present (mismatch is fatal; missing checksums warn and install anyway). Only the binary is kept; the swap is staged atomically and superseded `v*/` directories are deleted afterwards. `--dry-run`, `--tag`, `--repo`; env `LOGAMUFFIN_REPO`, `LOGAMUFFIN_TAG`. Logging via `scripts/entry.sh`. See [../README.md](../README.md) for the shared bundle contract.

## Related pages

- [../README.md](../README.md) — the bundle index
- [../../scripts/entry.sh.md](../../scripts/entry.sh.md) — the `log()` wrapper that calls it
- [../../scripts/env.sh.md](../../scripts/env.sh.md) — the cascade selecting `LOGAMUFFIN_BIN`
