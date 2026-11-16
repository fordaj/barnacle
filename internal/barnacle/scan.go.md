# Scanner

```mermaid
flowchart TD
    Root["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M3.5 5.7A2.2 2.2 0 0 1 5.7 3.5h4L12 6h5.3a2.2 2.2 0 0 1 2.2 2.2v8.6a2.2 2.2 0 0 1-2.2 2.2H5.7a2.2 2.2 0 0 1-2.2-2.2z'/></svg>Repository root</span>"] --> Walk["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M8.5 7 4 12l4.5 5'/><path d='M15.5 7 20 12l-4.5 5'/></svg>WalkDir</span>"]
    Walk --> Filter{{"<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M12 3.2 19.6 7.6v8.8L12 20.8l-7.6-4.4V7.6z'/></svg>Regular .md/.mmd?</span>"}}
    Filter -- no --> Skip[Skip]
    Filter -- yes --> Excludes{{"<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M12 3.2 19.6 7.6v8.8L12 20.8l-7.6-4.4V7.6z'/></svg><svg class='label-icon' viewBox='0 0 24 24' width='12' height='12' style='position:absolute;right:-31px;top:-14px' fill='none' stroke='currentColor' stroke-width='2.4' stroke-linecap='round' stroke-linejoin='round'><path d='M10 14a5 5 0 0 0 7.07 0L20.6 10.4a5 5 0 0 0-7.07-7.07l-2 2'/><path d='M14 10a5 5 0 0 0-7.07 0L3.4 13.6a5 5 0 0 0 7.07 7.07l2-2'/></svg>Excluded by config/defaults?</span>"}}
    Excludes -- yes --> Skip
    Excludes -- no --> Hidden{{"<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M12 3.2 19.6 7.6v8.8L12 20.8l-7.6-4.4V7.6z'/></svg>Hidden segment?</span>"}}
    Hidden -- yes --> Skip
    Hidden -- no --> Doc["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M4.5 5h15v14h-15z'/><path d='M8 14.5v-4l2.2 2.8 2.2-2.8v4'/></svg>Document</span>"]
    Doc --> Sorted["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M4.5 5h15v14h-15z'/><path d='M8 14.5v-4l2.2 2.8 2.2-2.8v4'/></svg><svg class='label-icon' viewBox='0 0 24 24' width='12' height='12' style='position:absolute;right:-31px;top:-14px' fill='none' stroke='currentColor' stroke-width='2.4' stroke-linecap='round' stroke-linejoin='round'><path d='M10 14a5 5 0 0 0 7.07 0L20.6 10.4a5 5 0 0 0-7.07-7.07l-2 2'/><path d='M14 10a5 5 0 0 0-7.07 0L3.4 13.6a5 5 0 0 0 7.07 7.07l2-2'/></svg>Sorted Document list</span>"]
    click Excludes href "config.go.md"
    click Sorted href "site.go.md"
    class Root folder
    class Walk goFile
    class Filter,Excludes,Hidden hexNode
    class Doc,Sorted mdFile
    classDef folder fill:#f0c558,stroke:#8a6d1c,color:#332a05
    classDef goFile fill:#8fd8ef,stroke:#0277a8,color:#08303d
    classDef hexNode fill:#ffe08a,stroke:#a16207,color:#422006
    classDef mdFile fill:#dcd6ce,stroke:#6b6255,color:#292524
```

`Scan` walks the repository root and returns every `.md` and `.mmd` file, sorted by slash-separated relative path. The output directory is never scanned.

## Rules

- `.git`, `node_modules`, `vendor`, and `dist` are always excluded.
- Dot-files and dot-directories are skipped unless `includeHidden` is set in the config.
- Additional exclusions come from the `exclude` config list: a bare name matches any path segment; an entry with a slash is matched as a glob against the slash-separated relative path and skips everything beneath it (`**` is not supported — `docs/archive` skips that directory and all of it).
- The output directory (`./site` by default) is skipped during the walk, even though it usually lives inside the root.
- Extensions are matched case-insensitively; only regular files are considered.

## Documents

Each discovered file becomes a `Document`:

- `RelPath` — source path relative to the root (`parser/parser.cpp.md`).
- `OutputRel` — generated page path; the extension is replaced with `.html` (`parser/parser.cpp.md` → `parser/parser.cpp.html`).
- `SourcePath` — absolute path of the file, used to read it later.
- `Title` — the first level-1 ATX heading for `.md` files (with closing hashes trimmed), otherwise the humanized file name (`my-page.md` → `My page`); `.mmd` files always get the humanized name. Falling back to the humanized name when there is no level-1 heading.
- `MermaidOnly` — true for `.mmd` files, which contain nothing but a diagram.

## Where it is used

`Generate` in [the site generator](site.go.md) calls `Scan` first; the returned list drives rendering, the navigation tree, and the collision check (a `.md` and a `.mmd` with the same base name both produce the same `.html`, and `Generate` fails instead of overwriting one).

## Related pages

- [The site generator](site.go.md) — calls `Scan` and consumes `Document`
- [Configuration](config.go.md) — supplies `Exclude` and `IncludeHidden`
- [The renderer](render.go.md) — turns the scanned files into HTML
- [Navigation and landing page](nav.go.md) — builds the tree from the scanned list
- [README.md](README.md) — package hub
