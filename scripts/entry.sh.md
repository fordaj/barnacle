# The script entry point (`scripts/entry.sh`)

```mermaid
flowchart LR
    Script("<img src='../tools/lucide/icons/terminal.svg' width='20' alt=''/> Any scripts/*.sh") --> Entry("<img src='../tools/lucide/icons/terminal.svg' width='20' alt=''/> entry.sh")
    Entry --> Env("<img src='../tools/lucide/icons/terminal.svg' width='20' alt=''/> env.sh — cascade<br/>VERSIONER_BIN · LOGAMUFFIN_BIN · … <img src='../tools/lucide/icons/link.svg' width='16' alt=''/>")
    Entry --> Log("<img src='../tools/lucide/icons/terminal.svg' width='20' alt=''/> log LEVEL MESSAGE")
    Log -->|binary present| Logamuffin("<img src='../tools/lucide/icons/file-code.svg' width='20' alt=''/> logamuffin")
    Log -->|missing / fresh clone| Fallback("<img src='../tools/lucide/icons/terminal.svg' width='20' alt=''/> printf fallback")
    click Env href "env.sh.md"
    class Script,Entry,Env,Log,Fallback shFile
    class Logamuffin goFile
    classDef shFile fill:#abe07a,stroke:#3f6212,color:#16280a
    classDef goFile fill:#8fd8ef,stroke:#0277a8,color:#08303d
```

Every repository script sources `scripts/entry.sh` instead of the individual cascade files, because it does two jobs at once:

1. **It loads the environment** — it sources [env.sh.md](env.sh.md), which sources the hidden cascade (`.common.env` plus the platform file selected from `uname`).
2. **It defines `log()`** — the shared logging function used by the scripts and by the `tools/*/update.sh` refresh scripts.

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

- [env.sh.md](env.sh.md) — the cascade `entry.sh` loads
- [version.sh.md](version.sh.md) — the version script that sources `entry.sh`
- [build.sh.md](build.sh.md) · [release.sh.md](release.sh.md) — consumers of `log()` and the cascade
- [README.md](README.md) — the scripts hub
