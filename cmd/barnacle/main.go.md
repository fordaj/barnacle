# Command barnacle

```mermaid
flowchart LR
    CLI("<img src='barnacle:folder.svg' width='20' alt=''/><br/>cmd/barnacle") --> Config("<img src='barnacle:file-code.svg' width='20' alt=''/><br/><u>LoadConfig</u>")
    CLI --> ListIcons("<img src='barnacle:list.svg' width='20' alt=''/><br/><u>list-icons</u>")
    Config --> Generate("<img src='barnacle:file-code.svg' width='20' alt=''/><br/><u>Generate</u>")
    Generate --> Scan("<img src='barnacle:file-code.svg' width='20' alt=''/><br/><u>Scan</u>")
    Generate --> Render("<img src='barnacle:file-code.svg' width='20' alt=''/><br/><u>MarkdownToHTML</u>")
    Scan --> Site("<img src='barnacle:folder.svg' width='20' alt=''/><br/>site/")
    Render --> Site
    click Config href "../../internal/barnacle/config.go.md"
    click Generate href "../../internal/barnacle/site.go.md"
    click Scan href "../../internal/barnacle/scan.go.md"
    click Render href "../../internal/barnacle/render.go.md"
    click ListIcons href "../../internal/barnacle/icons.go.md"
    class CLI,Site folder
    class ListIcons shFile
    class Config,Generate,Scan,Render goFile
    classDef folder fill:#f0c558,stroke:#8a6d1c,color:#332a05
    classDef shFile fill:#abe07a,stroke:#3f6212,color:#16280a
    classDef goFile fill:#8fd8ef,stroke:#0277a8,color:#08303d
```

The `barnacle` command turns the Markdown and Mermaid files that live alongside source code into a static, interconnected website. It takes one subcommand: `barnacle list-icons` (@tbl:subcommands).

## Subcommands

| Subcommand   | Description                                                                                                                                                 |
| ------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `list-icons` | Print the Lucide icons embedded in the binary (one per line, with the Lucide attribution) and exit. See [icons.go.md](../../internal/barnacle/icons.go.md). |

: The only subcommand: `barnacle list-icons` {#tbl:subcommands}

`list-icons` is dispatched before flag parsing, takes no arguments (anything after it is exit code `2`), and reads nothing from the filesystem or the configuration.

## Flags

| Flag              | Description                                                   |
| ----------------- | ------------------------------------------------------------- |
| `--config <path>` | Configuration file. Defaults to `./.barnacle.json`.           |
| `--site <path>`   | Output directory. Defaults to `./site`.                       |
| `--version`       | Print the version (stamped at build time) and exit.           |
| `--help`          | Show usage and exit (exit code `2`, like other usage errors). |

: The flags of the `barnacle` command {#tbl:flags}

`--config` changes only where the configuration is read from; the repository root is always the current working directory, and the scan starts there. Relative `--config` and `--site` paths are resolved against it.

## Exit codes

- `0` — success.
- `1` — runtime failure (unreadable or invalid config, scan/render error, unwritable output).
- `2` — usage error (unknown flag, unexpected positional argument, `--help`, arguments after `list-icons`).

`version` is `dev` unless stamped by the release scripts (`-ldflags "-X main.version=<v>"`, see `scripts/version.sh.md` and `scripts/release.sh.md`).

## The command lifecycle (edge links)

Mermaid can't click edges, but labels render as HTML — an inline anchor just works. Try it:

```mermaid
stateDiagram-v2
    [*] --> Idle
    Idle --> Scanning : <a href="#flags">barnacle(args)</a>
    Scanning --> Done : <a href="../../internal/barnacle/site.go.md">success, exit 0</a>
    Scanning --> Failed : error, exit 1
    Failed --> Idle : <a href="#best-effort-diagram-note">fix and retry</a>
```

`barnacle(args)` jumps to the [flags](#flags) table on this page; `success, exit 0` opens the site generator; `fix and retry` jumps to the diagram note below. The links rewrite exactly like Markdown links (see the README's "Clickable edges" section).

## Best-effort diagram note

After a successful generation, barnacle prints a non-fatal note to stderr listing every page that contains no Mermaid diagram (`barnacle: note: N page(s) have no Mermaid diagram: …`), plus one line of best-practice reminder. Pages are never rejected for lacking a diagram; set `"warnMissingMermaid": false` in `.barnacle.json` to silence the note. Exit code `0` is unaffected either way.

## Related pages

- [internal/barnacle/README.md](../../internal/barnacle/README.md) — the package that does the work
- [The icon lister](../../internal/barnacle/icons.go.md) — the `list-icons` subcommand and its attribution output
- [Configuration](../../internal/barnacle/config.go.md) — how `.barnacle.json` is loaded
- [The scanner](../../internal/barnacle/scan.go.md) — how documentation files are discovered
- [The site generator](../../internal/barnacle/site.go.md) — how the output is assembled
