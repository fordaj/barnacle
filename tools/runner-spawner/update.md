# The runner-spawner bundle (`tools/runner-spawner/`)

[Runner-spawner](https://github.com/fordaj/runner-spawner) manages self-hosted GitHub Actions runners. In this repository it backs the self-hosted CI described in the README ("Formatting and CI"): `.github/workflows/ci.yml` runs `check-formatting` and `test` jobs on `self-hosted` macOS and Windows runners, registered and lifecycle-managed with the bundled `runner-spawner` binary against the committed `.runner-spawner.json`:

```json
{
  "owner": "fordaj",
  "repo": "barnacle",
  "runnerNames": ["runner01", "runner02"],
  "runnersDir": ".runners",
  "runnerGroup": "Default",
  "runnerVersion": "2.338.2"
}
```

The runner tree itself (`.runners/`) is machine-local state and is not committed.

## Bundled platforms

`v0.0.1.2.6` — `macos-arm64`, `macos-x86_64`, `linux-arm64`, `linux-x86_64`, `windows-arm64`, `windows-x86_64`. Note: runner-spawner reports unsupported on Linux (macOS + Windows only); the Linux binaries are shipped for uniform bundle layout, and the Linux platform env files carry this caveat.

The pinned binary selection goes through the [env cascade](../../scripts/env.md) (`RUNNER_SPAWNER_BIN`). `RUNNER_SPAWNER_TOOL_REPO` is the bundle-refresh slug (`fordaj/runner-spawner`), not a script-selection variable.

## `update.sh`

Refreshes from `fordaj/runner-spawner` releases: archives whose stems carry the `<os>-<arch>` suffix (any single separator; `darwin-*`/`amd64`/`x64` are normalized onto the bundle layout), verified against the release's `checksums.txt`/`SHA256SUMS.txt`. Verification is mandatory here: a release with no checksums file aborts the refresh (`refusing to install unverified binaries`) — unlike the other bundles, which warn and install anyway. Only the binary is kept; the swap is staged atomically and superseded `v*/` directories are deleted afterwards. `--dry-run`, `--tag`, `--repo`; env `RUNNER_SPAWNER_TOOL_REPO`, `RUNNER_SPAWNER_TAG`. Logging via `scripts/entry.sh`. See [../README.md](../README.md) for the shared bundle contract.

## Related pages

- [../README.md](../README.md) — the bundle index
- [../../scripts/env.md](../../scripts/env.md) — the cascade selecting `RUNNER_SPAWNER_BIN`
- [../../README.md](../../README.md) — the CI section of the root README
