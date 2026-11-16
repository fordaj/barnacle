# Configuration (`config.go`)

```mermaid
flowchart TD
    Start["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M8.5 7 4 12l4.5 5'/><path d='M15.5 7 20 12l-4.5 5'/></svg>LoadConfig</span>"] --> Exists{{"<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M12 3.2 19.6 7.6v8.8L12 20.8l-7.6-4.4V7.6z'/></svg>File readable?</span>"}}
    Exists -- "no, default path" --> Defaults["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M8.5 7 4 12l4.5 5'/><path d='M15.5 7 20 12l-4.5 5'/></svg>Defaults</span>"]
    Exists -- "no, --config" --> Error["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><circle cx='12' cy='12' r='8.3'/><path d='M12 8.2v5'/><path d='M12 16.3h.01'/></svg><svg class='label-icon' viewBox='0 0 24 24' width='12' height='12' style='position:absolute;right:-31px;top:-14px' fill='none' stroke='currentColor' stroke-width='2.4' stroke-linecap='round' stroke-linejoin='round'><path d='M10 14a5 5 0 0 0 7.07 0L20.6 10.4a5 5 0 0 0-7.07-7.07l-2 2'/><path d='M14 10a5 5 0 0 0-7.07 0L3.4 13.6a5 5 0 0 0 7.07 7.07l2-2'/></svg>Error → exit 1</span>"]
    Exists -- yes --> Parse{{"<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M12 3.2 19.6 7.6v8.8L12 20.8l-7.6-4.4V7.6z'/></svg>Valid JSON?</span>"}}
    Parse -- no --> Error
    Parse -- yes --> Check{{"<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M12 3.2 19.6 7.6v8.8L12 20.8l-7.6-4.4V7.6z'/></svg>Blank exclude entry?</span>"}}
    Check -- yes --> Error
    Check -- no --> Cfg["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M9.5 4.5c-2.4 0-3.4 1-3.4 3.5 0 1.8-1 2.8-2 3.5 1 .7 2 1.7 2 3.5 0 2.5 1 3.5 3.4 3.5'/><path d='M14.5 4.5c2.4 0 3.4 1 3.4 3.5 0 1.8 1 2.8 2 3.5-1 .7-2 1.7-2 3.5 0 2.5-1 3.5-3.4 3.5'/></svg><svg class='label-icon' viewBox='0 0 24 24' width='12' height='12' style='position:absolute;right:-31px;top:-14px' fill='none' stroke='currentColor' stroke-width='2.4' stroke-linecap='round' stroke-linejoin='round'><path d='M10 14a5 5 0 0 0 7.07 0L20.6 10.4a5 5 0 0 0-7.07-7.07l-2 2'/><path d='M14 10a5 5 0 0 0-7.07 0L3.4 13.6a5 5 0 0 0 7.07 7.07l2-2'/></svg>Config<br/>.barnacle.json</span>"]
    click Error href "../../cmd/barnacle/main.go.md"
    click Cfg href "site.go.md"
    class Start,Defaults goFile
    class Exists,Parse,Check hexNode
    class Error errNode
    class Cfg jsonFile
    classDef goFile fill:#8fd8ef,stroke:#0277a8,color:#08303d
    classDef hexNode fill:#ffe08a,stroke:#a16207,color:#422006
    classDef errNode fill:#f6b8b8,stroke:#b91c1c,color:#450a0a
    classDef jsonFile fill:#cdb9f7,stroke:#6d28d9,color:#2e1065
```

`LoadConfig` reads and validates the repository-level configuration file (`.barnacle.json`) and produces the `Config` value the rest of the pipeline uses.

## The `Config` struct

- `Title` — shown in the sidebar, the landing page, and page titles. Defaults to the base name of the repository root; `.`, a bare separator, or empty falls back to `Documentation`.
- `Exclude` — directories or files to skip. An entry without a slash matches the base name of every path on the walk; an entry with a slash is matched as a glob against the slash-separated path relative to the root, and everything beneath it is skipped too. An empty exclusion entry is a config error. (`**` is not supported; use `docs/archive` to skip a directory.)
- `IncludeHidden` — allow scanning dot-directories and dot-files other than `.git`, which is always excluded.
- `WarnMissingMermaid` — `*bool`; `warnMissingMermaid` in JSON. Controls the non-fatal note barnacle prints listing pages that contain no Mermaid diagram. Unset or `true` enables it; `false` silences it (see `MermaidWarningEnabled`). Missing diagrams are never an error.

`defaultExcludes` always skips `.git`, `node_modules`, `vendor`, and `dist`, on top of anything configured here.

## Loading semantics

`LoadConfig(path, explicit, root)` has two failure modes on purpose. A missing file at the _default_ location is not an error — defaults are returned (`Config{Title: <root base name>}`). A missing or unreadable file at an _explicitly requested_ path (`--config`) is an error, as is invalid JSON or a blank exclusion entry. A blank `title` in the file falls back to the default title.

The configuration is read once at startup; `root` is always the current working directory — `--config` only changes where the file is read from.

## Where it is used

`cmd/barnacle/main.go.md` calls `LoadConfig`; the resulting `Config` flows into `Generate` (page titles) and `Scan` (exclusions and hidden-file policy). See [the scanner](scan.go.md) and [the site generator](site.go.md).

## Related pages

- [The scanner](scan.go.md) — consumes `Exclude` and `IncludeHidden`
- [The site generator](site.go.md) — consumes `Title` for the landing page
- [The page template](template.go.md) — what ends up on screen
- [README.md](README.md) — package hub
- [cmd/barnacle/main.go.md](../../cmd/barnacle/main.go.md) — calls `LoadConfig` at startup
