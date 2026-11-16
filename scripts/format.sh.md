# The formatter (`scripts/format.sh`)

```mermaid
flowchart LR
    uv["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M5 8l4.5 4L5 16'/><path d='M12 16.5h7'/></svg>uv run</span>"] --> FFT["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M5 8l4.5 4L5 16'/><path d='M12 16.5h7'/></svg>formattington</span>"]
    Config["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M9.5 4.5c-2.4 0-3.4 1-3.4 3.5 0 1.8-1 2.8-2 3.5 1 .7 2 1.7 2 3.5 0 2.5 1 3.5 3.4 3.5'/><path d='M14.5 4.5c2.4 0 3.4 1 3.4 3.5 0 1.8 1 2.8 2 3.5-1 .7-2 1.7-2 3.5 0 2.5-1 3.5-3.4 3.5'/></svg>.formattington.json</span>"] --> FFT
    FFT --> Prettier["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M4.5 5h15v14h-15z'/><path d='M8 14.5v-4l2.2 2.8 2.2-2.8v4'/></svg>prettier — md/json/yml</span>"]
    FFT --> Black[black — py]
    FFT --> Shfmt["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M5 8l4.5 4L5 16'/><path d='M12 16.5h7'/></svg>shfmt -i 2 — sh</span>"]
    FFT --> Taplo["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M9.5 4.5c-2.4 0-3.4 1-3.4 3.5 0 1.8-1 2.8-2 3.5 1 .7 2 1.7 2 3.5 0 2.5 1 3.5 3.4 3.5'/><path d='M14.5 4.5c2.4 0 3.4 1 3.4 3.5 0 1.8 1 2.8 2 3.5-1 .7-2 1.7-2 3.5 0 2.5-1 3.5-3.4 3.5'/></svg>taplo — toml</span>"]
    Check["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M5 8l4.5 4L5 16'/><path d='M12 16.5h7'/></svg>ci/format_check.sh --check</span>"] -.->|same config, exit 1 on diff| FFT
    class uv,FFT shFile
    class Config jsonFile
    class Prettier mdFile
    class Shfmt shFile
    class Taplo jsonFile
    class Check shFile
    classDef shFile fill:#abe07a,stroke:#3f6212,color:#16280a
    classDef jsonFile fill:#cdb9f7,stroke:#6d28d9,color:#2e1065
    classDef mdFile fill:#dcd6ce,stroke:#6b6255,color:#292524
```

Formats the barnacle repository in place with [formattington](https://github.com/fordaj/formattington). This repository dogfoods it for its own Markdown, Python, shell, and TOML files.

## What it does

- Resolves the repository root (the parent of `scripts/`) and `cd`s there.
- Requires `uv` on `PATH` (https://docs.astral.sh/uv/); fails with a clear error otherwise.
- Runs `uv run formattington --config "$REPO_ROOT/.formattington.json" all` with any extra arguments you pass.

Formattington is pinned in `pyproject.toml` to a git commit (no PyPI release exists yet) and locked in `uv.lock`; `uv run` creates `.venv/` automatically on first use. This project uses Python 3.14 (see `.python-version`), the latest version formattington supports.

## Scope (`.formattington.json`)

| Files                    | Tool         |
| ------------------------ | ------------ |
| Markdown (`*.md`)        | prettier     |
| JSON (`*.json`)          | prettier     |
| YAML (`*.yml`)           | prettier     |
| Python (`*.py`)          | black        |
| Shell (`*.sh`, `*.bash`) | shfmt `-i 2` |
| TOML (`*.toml`)          | taplo        |

Style for the JavaScript/Markdown/YAML/JSON formatter is configured by formattington's own defaults plus any `.prettierrc` present at the repository root (the `scripts/format.sh` header notes this; no `.prettierrc` is currently tracked). The bundled binaries under `tools/` and generated locations such as `dist/` are skipped.

## Prerequisites

Formatter executables must be on `PATH` (formattington exits `127` if one is missing): `shfmt`, `prettier`, `taplo` (e.g. via your OS package manager or npm for prettier); `black` is provided by the uv environment. `uv` is required at run time.

## Usage

```bash
./scripts/format.sh              # format all matching files in place
./scripts/format.sh --series     # format sequentially (ordered output)
./scripts/format.sh --verbose    # show which formatter executables were resolved
```

Extra flags are forwarded to `formattington` before `all` (e.g. `--series`, `--verbose`). `--config` is managed by the script; do not pass it.

To check formatting without modifying files (CI), use `./ci/format_check.sh`, which runs the same configuration with `--check` — see the README's "Formatting and CI" section.

## Related pages

- [README.md](../README.md) — the repository README (formatting and CI are documented there)
- [AGENTS.md](../AGENTS.md) — the same rules restated for contributors and agents
- [README.md](README.md) — the scripts that consume the same cascade
