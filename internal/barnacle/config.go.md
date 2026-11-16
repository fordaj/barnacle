# Configuration (`config.go`)

```mermaid
flowchart TD
    Start("<img src='../../tools/lucide/icons/file-code.svg' width='20' alt=''/> LoadConfig") --> Exists{{"<img src='../../tools/lucide/icons/hexagon.svg' width='20' alt=''/> File readable?"}}
    Exists -- "no, default path" --> Defaults("<img src='../../tools/lucide/icons/file-code.svg' width='20' alt=''/> Defaults")
    Exists -- "no, --config" --> Error("<img src='../../tools/lucide/icons/circle-exclamation-point.svg' width='20' alt=''/> Error → exit 1 <img src='../../tools/lucide/icons/link.svg' width='16' alt=''/>")
    Exists -- yes --> Parse{{"<img src='../../tools/lucide/icons/hexagon.svg' width='20' alt=''/> Valid JSON?"}}
    Parse -- no --> Error
    Parse -- yes --> Check{{"<img src='../../tools/lucide/icons/hexagon.svg' width='20' alt=''/> Blank exclude entry?"}}
    Check -- yes --> Error
    Check -- no --> Cfg("<img src='../../tools/lucide/icons/braces.svg' width='20' alt=''/> Config<br/>.barnacle.json <img src='../../tools/lucide/icons/link.svg' width='16' alt=''/>")
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
