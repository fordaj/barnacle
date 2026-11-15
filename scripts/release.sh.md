# The release script (`scripts/release.sh`)

```mermaid
flowchart LR
    Versioner[versioner] --> Verify[build throwaway<br/>--version matches?]
    Verify --> Build[build.sh — 6 platforms]
    Build --> Package[archives + checksums.txt]
    Package --> Tag[tag HEAD + push]
    Tag --> GH[gh release create<br/>or refresh assets]
    click Build href "build.sh.md"
    click Versioner href "version.sh.md"
```

Computes the next version with versioner (dogfooding), cross-compiles the binaries, packages the release archives, tags HEAD, and creates the GitHub release with the archives attached.

One archive is produced per platform:

```text
barnacle-<platform>.tar.gz    (.zip on Windows)
barnacle-<platform>/          (top-level directory in the archive)
  barnacle{,.exe}  LICENSE  README.md
```

The binaries are built with CGO disabled and are statically linked and portable. The program version is dynamic via versioner (`.versioner.json`); it is stamped into every cross-compiled binary with `-ldflags -X`, so the tag and `barnacle --version` agree by construction.

## Usage

```bash
scripts/release.sh                # version, build, tag, push, release + upload
scripts/release.sh --build-only   # cross-compile into dist/ only (stops before tag/push/release)
scripts/release.sh --dry-run      # compute + verify + build, touch nothing remote
```

`--build-only` tolerates a dirty working tree; a real release refuses one (commit or stash first). `--dry-run` logs the exact tag/push/release commands it would run.

## Prerequisites

`git`, a Go toolchain, `tar`, and `zip`; a real release additionally needs `gh` authenticated for the target repository. Bundled versioner/logamuffin binaries are selected via [entry.sh.md](entry.sh.md) (refresh with `tools/versioner/update.sh` and `tools/logamuffin/update.sh`). Logging uses `log <level> <message>` from `entry.sh` (logamuffin, with printf fallback). The working tree must be clean and `HEAD` should be pushed for a real release.

## What it does

1. Sources [entry.sh.md](entry.sh.md); checks required commands (`gh` only for a real release; `gh auth status` must pass).
2. Refuses to continue on a dirty tree (except `--build-only`/`--dry-run`, which warn); best-effort `git fetch origin --tags`.
3. Computes `VERSION`/`TAG` with `versioner --repo-dir "$REPO_ROOT" --quiet`; the tag must match `vX.Y.Z...`.
4. **Verifies**: builds a throwaway binary stamped with `$VERSION` and requires `barnacle --version` to report it.
5. Cross-compiles into `dist/` via [build.sh.md](build.sh.md) (`VERSION="$VERSION"`), then packages each `dist/<platform>/` into `dist/barnacle-<platform>.tar.gz`/`.zip` staging in a temp directory (binary + `LICENSE` + `README.md`), and writes `dist/checksums.txt` (`shasum -a 256` or `sha256sum`).
6. `--build-only`: stops here. Otherwise tags HEAD (an existing tag is reused), pushes it, and either creates the GitHub release (`gh release create --generate-notes`) or, if it already exists, re-uploads the assets with `--clobber` — the script is idempotent and safe to re-run to (re)attach binaries.

## Related pages

- [build.sh.md](build.sh.md) — the cross-compile step this script wraps
- [version.sh.md](version.sh.md) — how the version is computed
- [entry.sh.md](entry.sh.md) · [env.sh.md](env.sh.md) — the cascade that provides `VERSIONER_BIN`/`LOGAMUFFIN_BIN`
- [README.md](README.md) — the scripts hub
