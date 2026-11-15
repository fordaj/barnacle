# Command barnacle

```mermaid
flowchart LR
    CLI[cmd/barnacle] --> Config[LoadConfig]
    Config --> Generate[Generate]
    Generate --> Scan[Scan]
    Generate --> Render[MarkdownToHTML]
    Scan --> Site[site/]
    Render --> Site
    click Config href "../../internal/barnacle/config.go.md"
    click Generate href "../../internal/barnacle/site.go.md"
    click Scan href "../../internal/barnacle/scan.go.md"
    click Render href "../../internal/barnacle/render.go.md"
```

The `barnacle` command turns the Markdown and Mermaid files that live alongside source code into a static, interconnected website.

## Flags

| Flag              | Description                                                   |
| ----------------- | ------------------------------------------------------------- |
| `--config <path>` | Configuration file. Defaults to `./.barnacle.json`.           |
| `--site <path>`   | Output directory. Defaults to `./site`.                       |
| `--version`       | Print the version (stamped at build time) and exit.           |
| `--help`          | Show usage and exit (exit code `2`, like other usage errors). |

`--config` changes only where the configuration is read from; the repository root is always the current working directory, and the scan starts there. Relative `--config` and `--site` paths are resolved against it.

## Exit codes

- `0` — success.
- `1` — runtime failure (unreadable or invalid config, scan/render error, unwritable output).
- `2` — usage error (unknown flag, unexpected positional argument, `--help`).

`version` is `dev` unless stamped by the release scripts (`-ldflags "-X main.version=<v>"`, see `scripts/version.sh.md` and `scripts/release.sh.md`).

## The command lifecycle (edge links)

Edge links (`%% barnacle: link`) make transitions clickable — try it:

```mermaid
stateDiagram-v2
    [*] --> Idle
    Idle --> Scanning : barnacle(args)
    Scanning --> Done : success, exit 0
    Scanning --> Failed : error, exit 1
    Failed --> Idle : fix and retry
    %% barnacle: link "barnacle(args)" href "#flags" "Jump to the flags table"
    %% barnacle: link "success, exit 0" href "../../internal/barnacle/site.go.md" "How the output is assembled"
    %% barnacle: link "fix and retry" href "#best-effort-diagram-note" "The note that lists pages without a diagram"
```

`barnacle(args)` jumps to the [flags](#flags) table on this page; `success, exit 0` opens the site generator; `fix and retry` jumps to the diagram note below. These are the same `%% barnacle: link` directives documented in the README's "Clickable edges" section.

## Best-effort diagram note

After a successful generation, barnacle prints a non-fatal note to stderr listing every page that contains no Mermaid diagram (`barnacle: note: N page(s) have no Mermaid diagram: …`), plus one line of best-practice reminder. Pages are never rejected for lacking a diagram; set `"warnMissingMermaid": false` in `.barnacle.json` to silence the note. Exit code `0` is unaffected either way.

## Related pages

- [internal/barnacle/README.md](../../internal/barnacle/README.md) — the package that does the work
- [Configuration](../../internal/barnacle/config.go.md) — how `.barnacle.json` is loaded
- [The scanner](../../internal/barnacle/scan.go.md) — how documentation files are discovered
- [The site generator](../../internal/barnacle/site.go.md) — how the output is assembled
