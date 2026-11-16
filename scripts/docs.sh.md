# The docs generator (`scripts/docs.sh`)

```mermaid
flowchart LR
    Src["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M3.5 5.7A2.2 2.2 0 0 1 5.7 3.5h4L12 6h5.3a2.2 2.2 0 0 1 2.2 2.2v8.6a2.2 2.2 0 0 1-2.2 2.2H5.7a2.2 2.2 0 0 1-2.2-2.2z'/></svg>Source tree<br/>.md / .mmd next to code</span>"] --> Barnacle["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M8.5 7 4 12l4.5 5'/><path d='M15.5 7 20 12l-4.5 5'/></svg>BARNACLE_BIN</span>"]
    Cascade["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M5 8l4.5 4L5 16'/><path d='M12 16.5h7'/></svg><svg class='label-icon' viewBox='0 0 24 24' width='12' height='12' style='position:absolute;right:-31px;top:-14px' fill='none' stroke='currentColor' stroke-width='2.4' stroke-linecap='round' stroke-linejoin='round'><path d='M10 14a5 5 0 0 0 7.07 0L20.6 10.4a5 5 0 0 0-7.07-7.07l-2 2'/><path d='M14 10a5 5 0 0 0-7.07 0L3.4 13.6a5 5 0 0 0 7.07 7.07l2-2'/></svg>env.sh cascade<br/>dist/&lt;platform&gt;/barnacle</span>"] --> Barnacle
    Barnacle --> Site["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M3.5 5.7A2.2 2.2 0 0 1 5.7 3.5h4L12 6h5.3a2.2 2.2 0 0 1 2.2 2.2v8.6a2.2 2.2 0 0 1-2.2 2.2H5.7a2.2 2.2 0 0 1-2.2-2.2z'/></svg>./site<br/>static HTML</span>"]
    click Cascade href "env.sh.md"
    class Src,Site folder
    class Cascade shFile
    class Barnacle goFile
    classDef folder fill:#f0c558,stroke:#8a6d1c,color:#332a05
    classDef shFile fill:#abe07a,stroke:#3f6212,color:#16280a
    classDef goFile fill:#8fd8ef,stroke:#0277a8,color:#08303d
```

Generates the documentation site you are reading: it runs the barnacle binary built from this source tree — `dist/<platform>/barnacle[.exe]`, produced by [`scripts/build.sh`](build.sh.md) — against this repository. This repository is the barnacle project itself, so the site you get always reflects the tree at hand; no release is needed to render your changes. On a fresh clone, run `scripts/build.sh` once before the first `scripts/docs.sh`.

Barnacle scans the working tree for Markdown (`.md`) and Mermaid (`.mmd`) files, renders each as a page, and keeps relative links between them. Generated output is static HTML; it is written to `./site/` by default and is never committed (it is in `.gitignore`; `site/` is always excluded from the scan).

## Usage

```bash
scripts/docs.sh                       # generate into ./site/
scripts/docs.sh --site ./public       # forward --site to barnacle
scripts/docs.sh --config ./docs/barnacle.json
scripts/docs.sh --help                # usage for this script
```

Any arguments other than `--help`/`-h` are forwarded to barnacle verbatim; see `barnacle --help` for `--config`/`--site` semantics. Barnacle's exit code is barnacle's own; a missing or non-executable binary makes this script fail with instructions to run `scripts/build.sh` (or set `BARNACLE_BIN` explicitly).

## How it locates barnacle

Sources [entry.sh.md](entry.sh.md) and resolves `BARNACLE_BIN` through the [cascade](env.sh.md) (`dist/<platform>/barnacle[.exe]` by default). It then runs barnacle with the repository root as the working directory, so barnacle's auto-discovered `./.barnacle.json` and its default `./site/` output land at the repository root no matter where the script is invoked from.

This repository's `.barnacle.json` sets `title` to `barnacle` and excludes `tools` and `dist` (the pinned helper bundles carry their own upstream documentation); the output directory is always excluded, and hidden files and directories (including the cascade's `.common.env` et al.) are skipped unless `.barnacle.json` sets `includeHidden`.

## Staying in sync with source

The binary under `dist/` is a build artifact of this tree, not a downloaded pin — so the site always reflects whatever the tree renders today. To refresh it after changing the renderer or the template, rerun `scripts/build.sh` (incremental; the whole cross-compile takes a few seconds) and then `scripts/docs.sh`.

Setting `BARNACLE_BIN` overrides the binary used, exactly as before:

```bash
BARNACLE_BIN=/path/to/barnacle scripts/docs.sh
```

`BARNACLE_BIN` is honoured verbatim — repository-relative paths are anchored to the repository root by the cascade; absolute paths are left untouched.

## What the site contains

Every `.md`/`.mmd` in the tree — including this one — except the excluded `tools/` and `dist/` trees and hidden files. See the [root README](../README.md) and [AGENTS.md](../AGENTS.md). Because `tools/` is excluded, the pages under `tools/` are not part of the site and are not linked from site pages.

## Related pages

- [entry.sh.md](entry.sh.md) · [env.sh.md](env.sh.md) — how `BARNACLE_BIN` is resolved
- [README.md](README.md) — the scripts hub
- [../AGENTS.md](../AGENTS.md) — the documentation rules this site follows
