# The script entry point (`scripts/entry.sh`)

Every repository script sources `scripts/entry.sh` instead of the individual cascade files, because it does two jobs at once:

1. **It loads the environment** — it sources [env.md](env.md), which sources the hidden cascade (`.common.env` plus the platform file selected from `uname`).
2. **It defines `log()`** — the shared logging function used by the scripts and by the `tools/*/update.bash` refresh scripts.

## `log()` semantics

`log LEVEL MESSAGE...` forwards to the pinned logamuffin binary referenced by `LOGAMUFFIN_BIN`. The caller context is passed through in `LOG_SOURCE` (`file[:function]:line`), and logamuffin's own exit code (`0` for success/filtered, `2` for usage/config errors) is preserved even under `set -e`.

When the binary is missing or not executable (a fresh clone before tools are populated), `log()` falls back to a `printf` implementation with the same call shape: word-to-severity mapping (TRACE…ALWAYS → numeric thresholds), `LOG_LEVEL`/`LOGAMUFFIN_LEVEL` filtering, and stream routing (ERROR and above to stderr, everything else to stdout). The fallback exists so the scripts work on a brand-new checkout without network access. With no arguments, the fallback prints a usage message to stderr and returns `2`.

## Usage

```bash
#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
. "$SCRIPT_DIR/env.sh"      # or entry.sh when you also need log()
```

## Related pages

- [env.md](env.md) — the cascade `entry.sh` loads
- [version.md](version.md) — the version script that sources `entry.sh`
- [build.md](build.md) · [release.md](release.md) — consumers of `log()` and the cascade
- [README.md](README.md) — the scripts hub
