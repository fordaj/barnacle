# The environment cascade (`scripts/env.sh` + the hidden `.*.env` files)

`scripts/env.sh` is the plumbing that makes the pinned helper tools resolvable from any script, in any working directory, on every supported platform. It is sourced — never executed — and it works from both bash and zsh.

## What it does

1. Locates the cascade files relative to itself (bash via `$BASH_SOURCE`, zsh via `%x`, `eval`-wrapped so bash never parses the zsh-only syntax) so sourcing works from any cwd.
2. Brings in `.common.env`, then the platform file selected by `uname -s`/`uname -m` (`Darwin:arm64` → `.macos-arm64.env`, `Linux:x86_64` → `.linux-x86_64.env`, `Linux:aarch64` → `.linux-arm64.env`, `MINGW*`/`MSYS*`/`CYGWIN*` on x86_64/arm64 → `.windows-x86_64.env` / `.windows-arm64.env`). Unsupported platforms print an error and return `1`.
3. **Anchors** every repository-relative `*_BIN` to the repository root (`REPO_ROOT_DIR`), leaving absolute paths untouched. After sourcing, `$VERSIONER_BIN`, `$LOGAMUFFIN_BIN`, `$RUNNER_SPAWNER_BIN`, and `$BARNACLE_BIN` are absolute.
4. Sets `REPO_ROOT_DIR` for callers that need the repository root (for example [version.md](version.md)).

Everything is exported with `set -a` while sourcing and un-exported at the end, so the cascade does not leak into the caller's environment beyond the intended variables.

## The hidden files

Hidden files are excluded from barnacle's default scan (this repository's `.barnacle.json` excludes `tools` and `dist`, and hidden files are skipped everywhere by default), so their contract is documented here:

- **`.common.env`** — `TOOLS_DIR=tools`; the upstream `fordaj/*` repository slugs (with the versioner's `VERSIONER_TAG_PATTERN` and `VERSIONER_CONFIG_FILE`); the pinned `*_VERSION_DIR`s:
  - `VERSIONER_VERSION_DIR` → `tools/versioner/v0.0.3.0.5`
  - `LOGAMUFFIN_VERSION_DIR` → `tools/logamuffin/v0.1.3.0.35`
  - `RUNNER_SPAWNER_VERSION_DIR` → `tools/runner-spawner/v0.0.1.2.6`
  - `BARNACLE_VERSION_DIR` → `tools/barnacle/v0.0.1.0.5`

  When bumping a pinned tool, this file names the version directory the platform files point into.

- **`.<platform>.env`** — one per supported platform (`linux-arm64`, `linux-x86_64`, `macos-arm64`, `macos-x86_64`, `windows-arm64`, `windows-x86_64`). Each sets `PROJECT_PLATFORM` and the `*_BIN` paths, including `BARNACLE_BIN`. Two platform-file caveats are recorded inline: runner-spawner reports unsupported on Linux (macOS + Windows only; the Linux binaries ship for uniform bundle layout), and there is no windows-arm64 versioner bundle upstream — `.windows-arm64.env` points `VERSIONER_BIN` at the `windows-x86_64` path, which does not exist as a versioner bundle; set `VERSIONER_BIN` explicitly there if needed.

## Related pages

- [entry.md](entry.md) — the entry point most scripts source (it loads this file and defines `log()`)
- [docs.md](docs.md) — the docs generator that resolves `BARNACLE_BIN` through this cascade
- [README.md](README.md) — the scripts hub
