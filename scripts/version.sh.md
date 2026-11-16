# The version script (`scripts/version.sh`)

```mermaid
flowchart LR
    Git[Git history] --> Versioner["versioner --quiet"]
    Rules[".versioner.json"] --> Versioner
    Versioner --> Out["VERSION / TAG<br/>text · shell · json"]
    Out --> Build["build.sh — stamp binary"]
    Out --> Release["release.sh — next tag"]
    click Build href "build.sh.md"
    click Release href "release.sh.md"
    class Versioner tagNode
    class Rules jsonFile
    class Out mdFile
    class Build,Release shFile
    classDef tagNode fill:#f6b8d4,stroke:#be185d,color:#500724
    classDef jsonFile fill:#cdb9f7,stroke:#6d28d9,color:#2e1065
    classDef mdFile fill:#dcd6ce,stroke:#6b6255,color:#292524
    classDef shFile fill:#abe07a,stroke:#3f6212,color:#16280a
```

Prints the version of this repository as computed by the pinned versioner — the same computation the release pipeline uses, so a local build and the published tag can never drift apart.

## Usage

```bash
scripts/version.sh                 # VERSION=<v> and TAG=<t> lines (default)
scripts/version.sh --format text   # same
scripts/version.sh --format shell  # export VERSION=... and export TAG=...
scripts/version.sh --format json   # the raw versioner JSON document
```

`--format` requires a value; `--help`/`-h` prints usage. Unknown arguments and an unknown format are usage errors (exit `2`).

## How it works

Sources [entry.sh.md](entry.sh.md) for the cascade and `log()`. If `VERSIONER_BIN` is unset or not executable, it fails with `no versioner binary found` and exits `1`. Otherwise it runs `versioner --repo-dir <root> --quiet` (plus `--format json` for the `json` format); the repo directory defaults to `VERSIONER_REPO_DIR` or the repository root implied by `REPO_ROOT_DIR` (the parent of `scripts/`).

The computation itself follows `.versioner.json` at the repository root: `defaultBranchVersion: 254`, `exemptBranches: { main: 0 }`, `tagPattern: v[0-9]*`.

## Consumers

- [build.sh.md](build.sh.md) — stamps the version into the cross-compiled binaries (`--format shell`).
- [release.sh.md](release.sh.md) — computes the next tag, verifies the binary reports it, tags and publishes.

## Related pages

- [entry.sh.md](entry.sh.md) · [env.sh.md](env.sh.md)
- [build.sh.md](build.sh.md) · [release.sh.md](release.sh.md)
- [README.md](README.md) — the scripts hub
