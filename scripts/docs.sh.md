# The docs generator (`scripts/docs.sh`)

```mermaid
flowchart LR
    Src["Source tree<br/>.md / .mmd next to code"] --> Barnacle[BARNACLE_BIN]
    Cascade["env.sh cascade<br/>tools/barnacle/v0.0.1.0.16"] --> Barnacle
    Barnacle --> Site["./site<br/>static HTML"]
    click Cascade href "env.sh.md"
    class Src,Site folder
    class Cascade shFile
    class Barnacle goFile
    classDef folder fill:#f0c558,stroke:#8a6d1c,color:#332a05
    classDef shFile fill:#abe07a,stroke:#3f6212,color:#16280a
    classDef goFile fill:#8fd8ef,stroke:#0277a8,color:#08303d
```

Generates the documentation site you are reading: it runs the pinned barnacle binary against this repository. This repository is the barnacle project itself, so the pinned release (`tools/barnacle/v0.0.1.0.16`) — not a binary built from this source tree — is what renders the site by default.

Barnacle scans the working tree for Markdown (`.md`) and Mermaid (`.mmd`) files, renders each as a page, and keeps relative links between them. Generated output is static HTML; it is written to `./site/` by default and is never committed (it is in `.gitignore`; `site/` is always excluded from the scan).

## Usage

```bash
scripts/docs.sh                       # generate into ./site/
scripts/docs.sh --site ./public       # forward --site to barnacle
scripts/docs.sh --config ./docs/barnacle.json
scripts/docs.sh --help                # usage for this script
```

Any arguments other than `--help`/`-h` are forwarded to barnacle verbatim; see `barnacle --help` for `--config`/`--site` semantics. Barnacle's exit code is barnacle's own; a missing or non-executable bundle makes this script fail with instructions to run `tools/barnacle/update.sh` (or set `BARNACLE_BIN` explicitly).

## How it locates barnacle

Sources [entry.sh.md](entry.sh.md) and resolves `BARNACLE_BIN` through the [cascade](env.sh.md) (`tools/barnacle/v<version>/<platform>/barnacle[.exe]`). It then runs barnacle with the repository root as the working directory, so barnacle's auto-discovered `./.barnacle.json` and its default `./site/` output land at the repository root no matter where the script is invoked from.

This repository's `.barnacle.json` sets `title` to `barnacle` and excludes `tools` and `dist` (the pinned binaries carry their own documentation); the output directory is always excluded, and hidden files and directories (including the cascade's `.common.env` et al.) are skipped unless `.barnacle.json` sets `includeHidden`.

## Dogfooding unreleased changes

The pinned binary lags `main`. To render the site with barnacle built from this source tree — for example to verify a renderer change end-to-end before a release — build and point the script at the result:

```bash
scripts/build.sh    # stamps the versioner-computed version into dist/<platform>/barnacle
BARNACLE_BIN=dist/macos-arm64/barnacle scripts/docs.sh
```

`BARNACLE_BIN` is honoured verbatim — repository-relative paths are anchored to the repository root by the cascade; absolute paths are left untouched. The pin bump to a release happens through `tools/barnacle/update.sh` after the release is cut.

## What the site contains

Every `.md`/`.mmd` in the tree — including this one — except the excluded `tools/` and `dist/` trees and hidden files. See the [root README](../README.md) and [AGENTS.md](../AGENTS.md). Because `tools/` is excluded, the pinned-bundle pages under `tools/` are not part of the site and are not linked from site pages.

## Related pages

- [entry.sh.md](entry.sh.md) · [env.sh.md](env.sh.md) — how `BARNACLE_BIN` is resolved
- [README.md](README.md) — the scripts hub
- [../AGENTS.md](../AGENTS.md) — the documentation rules this site follows
