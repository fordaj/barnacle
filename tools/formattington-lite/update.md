# The formattington-lite bundle (`tools/formattington-lite/`)

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
- [../../scripts/format.md](../../scripts/format.md) — how formattington is actually run in this repository
- [../../scripts/env.md](../../scripts/env.md) — the cascade (which does not select this bundle)
