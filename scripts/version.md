# The version script (`scripts/version.sh`)

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

Sources [entry.md](entry.md) for the cascade and `log()`. If `VERSIONER_BIN` is unset or not executable, it fails with `no versioner binary found` and exits `1`. Otherwise it runs `versioner --repo-dir <root> --quiet` (plus `--format json` for the `json` format); the repo directory defaults to `VERSIONER_REPO_DIR` or the repository root implied by `REPO_ROOT_DIR` (the parent of `scripts/`).

The computation itself follows `.versioner.json` at the repository root: `defaultBranchVersion: 254`, `exemptBranches: { main: 0 }`, `tagPattern: v[0-9]*`.

## Consumers

- [build.md](build.md) — stamps the version into the cross-compiled binaries (`--format shell`).
- [release.md](release.md) — computes the next tag, verifies the binary reports it, tags and publishes.

## Related pages

- [entry.md](entry.md) · [env.md](env.md)
- [build.md](build.md) · [release.md](release.md)
- [README.md](README.md) — the scripts hub
