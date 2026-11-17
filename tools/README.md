# Pinned helper tools (`tools/`)

```mermaid
flowchart LR
    Gh("<img src='barnacle:tag.svg' width='20' alt=''/><br/>GitHub release<br/>fordaj/*") --> Update("<img src='barnacle:terminal.svg' width='20' alt=''/><br/>update.sh")
    Update --> Verify("<img src='barnacle:check.svg' width='20' alt=''/><br/>checksums<br/>when present")
    Verify --> Stage("<img src='barnacle:folder.svg' width='20' alt=''/><br/>stage in<br/>temp dir")
    Stage --> Swap("<img src='barnacle:folder.svg' width='20' alt=''/><br/>atomic swap<br/>delete superseded v*")
    class Gh tagNode
    class Update shFile
    class Verify checkNode
    class Stage,Swap folder
    classDef tagNode fill:#f6b8d4,stroke:#be185d,color:#500724
    classDef shFile fill:#abe07a,stroke:#3f6212,color:#16280a
    classDef checkNode fill:#9fe8c9,stroke:#0f766e,color:#052e22
    classDef folder fill:#f0c558,stroke:#8a6d1c,color:#332a05
```

`tools/` carries the repository's pinned helper binaries — formatting, logging and CI-runner tooling — plus one `update.sh` per tool that refreshes them. These are the tools used to build, log, format, release and run CI for this repository; they are not Go dependencies. Barnacle itself is not pinned here: its binary is built from this tree into `dist/` by `scripts/build.sh`.

These pages are **not part of the generated site** (`.barnacle.json` excludes `tools`); they are maintained as ordinary Markdown for people reading the repository.

## Bundling convention

Every bundle follows the same layout:

```text
tools/<name>/
  update.sh           # refresh from the upstream GitHub release
  v<version>/<platform>/<binary>[.exe]
```

- Binaries are pinned by version directory; the pinned version is named by `*_VERSION_DIR` in `scripts/.common.env`.
- `<platform>` is the hyphenated `<os>-<arch>` name (`linux-x86_64`, `macos-arm64`, `windows-x86_64`, …), matching Go and this repository's build layout.
- `update.sh` downloads the latest (or `--tag`ged) upstream release via `gh`, verifies checksums (mandatory for runner-spawner; warn-and-continue for the others when the release ships none), stages into a temporary directory, swaps in atomically, and deletes superseded version folders only after the new one is staged — a failed refresh leaves the previous bundle intact. Each script has `--dry-run`, `--tag`, `--repo`, and `--help`, and resolves its own directory, so it can run from anywhere.
- Archives are platform-native (`.tar.gz` on macOS/Linux, `.zip` on Windows); legacy `.7z` assets are still accepted as a fallback.

## The bundles

| Tool | Used for | Page |
| --- | --- | --- |
| [formattington-lite](formattington-lite/update.sh.md) | Pinned `formattington-lite` formatter binaries (not wired into `./scripts/format.sh` or CI — those run formattington via `uv`) | [update.sh](formattington-lite/update.sh.md) |
| [logamuffin](logamuffin/update.sh.md) | Structured logging backend for `log()` in the scripts | [update.sh](logamuffin/update.sh.md) |
| [runner-spawner](runner-spawner/update.sh.md) | Manages the self-hosted CI runners (see `.runner-spawner.json` and `.github/workflows/ci.yml`) | [update.sh](runner-spawner/update.sh.md) |
| [versioner](versioner/update.sh.md) | Computes the repository version for builds and releases | [update.sh](versioner/update.sh.md) |
| [lucide](lucide/update.sh.md) | Vendored [Lucide](https://lucide.dev) icons (ISC) used by the diagram conventions; inlined into the generated site by barnacle, embedded into every binary (`lucide.go`) and referenced via `barnacle:<name>.svg` / `barnacle list-icons` | [update.sh](lucide/update.sh.md) · [lucide.go](lucide/lucide.go.md) |

`scripts/.common.env` and `scripts/.<platform>.env` select these via `*_VERSION_DIR`/`*_BIN`; see [scripts/env.sh.md](../scripts/env.sh.md). (Those env files are barnacle-hidden; `scripts/env.sh.md` documents them.)

## Refreshing a pinned tool

```bash
tools/versioner/update.sh            # download the latest release, verify, swap
tools/versioner/update.sh --dry-run  # show what would change
```

(Same flags for `logamuffin`, `runner-spawner`, `formattington-lite`. `tools/lucide/update.sh` refreshes the vendored icon set the same way.)

## Related pages

- [../scripts/README.md](../scripts/README.md) — the scripts that source the cascade
- [../AGENTS.md](../AGENTS.md) — the documentation rules these pages follow
- [../README.md](../README.md) — the repository README
