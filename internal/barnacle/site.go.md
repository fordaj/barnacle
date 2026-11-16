# Site generator

```mermaid
flowchart TD
    Config["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M9.5 4.5c-2.4 0-3.4 1-3.4 3.5 0 1.8-1 2.8-2 3.5 1 .7 2 1.7 2 3.5 0 2.5 1 3.5 3.4 3.5'/><path d='M14.5 4.5c2.4 0 3.4 1 3.4 3.5 0 1.8 1 2.8 2 3.5-1 .7-2 1.7-2 3.5 0 2.5-1 3.5-3.4 3.5'/></svg><svg class='label-icon' viewBox='0 0 24 24' width='12' height='12' style='position:absolute;right:-31px;top:-14px' fill='none' stroke='currentColor' stroke-width='2.4' stroke-linecap='round' stroke-linejoin='round'><path d='M10 14a5 5 0 0 0 7.07 0L20.6 10.4a5 5 0 0 0-7.07-7.07l-2 2'/><path d='M14 10a5 5 0 0 0-7.07 0L3.4 13.6a5 5 0 0 0 7.07 7.07l2-2'/></svg>.barnacle.json</span>"] --> Generate["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M8.5 7 4 12l4.5 5'/><path d='M15.5 7 20 12l-4.5 5'/></svg>Generate</span>"]
    Generate --> Scan["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M8.5 7 4 12l4.5 5'/><path d='M15.5 7 20 12l-4.5 5'/></svg><svg class='label-icon' viewBox='0 0 24 24' width='12' height='12' style='position:absolute;right:-31px;top:-14px' fill='none' stroke='currentColor' stroke-width='2.4' stroke-linecap='round' stroke-linejoin='round'><path d='M10 14a5 5 0 0 0 7.07 0L20.6 10.4a5 5 0 0 0-7.07-7.07l-2 2'/><path d='M14 10a5 5 0 0 0-7.07 0L3.4 13.6a5 5 0 0 0 7.07 7.07l2-2'/></svg>Scan (walk .md/.mmd)</span>"]
    Scan --> Render["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M8.5 7 4 12l4.5 5'/><path d='M15.5 7 20 12l-4.5 5'/></svg><svg class='label-icon' viewBox='0 0 24 24' width='12' height='12' style='position:absolute;right:-31px;top:-14px' fill='none' stroke='currentColor' stroke-width='2.4' stroke-linecap='round' stroke-linejoin='round'><path d='M10 14a5 5 0 0 0 7.07 0L20.6 10.4a5 5 0 0 0-7.07-7.07l-2 2'/><path d='M14 10a5 5 0 0 0-7.07 0L3.4 13.6a5 5 0 0 0 7.07 7.07l2-2'/></svg>MarkdownToHTML</span>"]
    Render --> Pages["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><rect x='3.5' y='5' width='17' height='14' rx='2'/><path d='M3.5 9.5h17'/></svg>Per-page HTML<br/>+ list of pages without a diagram</span>"]
    Generate --> Index["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><rect x='3.5' y='5' width='17' height='14' rx='2'/><path d='M3.5 9.5h17'/></svg><svg class='label-icon' viewBox='0 0 24 24' width='12' height='12' style='position:absolute;right:-31px;top:-14px' fill='none' stroke='currentColor' stroke-width='2.4' stroke-linecap='round' stroke-linejoin='round'><path d='M10 14a5 5 0 0 0 7.07 0L20.6 10.4a5 5 0 0 0-7.07-7.07l-2 2'/><path d='M14 10a5 5 0 0 0-7.07 0L3.4 13.6a5 5 0 0 0 7.07 7.07l2-2'/></svg>index.html<br/>landing page</span>"]
    Pages --> Assets["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M3.5 5.7A2.2 2.2 0 0 1 5.7 3.5h4L12 6h5.3a2.2 2.2 0 0 1 2.2 2.2v8.6a2.2 2.2 0 0 1-2.2 2.2H5.7a2.2 2.2 0 0 1-2.2-2.2z'/></svg>assets/<br/>barnacle.css · mermaid.min.js</span>"]
    Pages --> Out["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M3.5 5.7A2.2 2.2 0 0 1 5.7 3.5h4L12 6h5.3a2.2 2.2 0 0 1 2.2 2.2v8.6a2.2 2.2 0 0 1-2.2 2.2H5.7a2.2 2.2 0 0 1-2.2-2.2z'/></svg>site/</span>"]
    Index --> Out
    Assets --> Out
    click Scan href "scan.go.md"
    click Render href "render.go.md"
    click Index href "nav.go.md"
    click Config href "config.go.md"
    class Config jsonFile
    class Generate,Scan,Render goFile
    class Pages,Index htmlFile
    class Assets,Out folder
    classDef jsonFile fill:#cdb9f7,stroke:#6d28d9,color:#2e1065
    classDef goFile fill:#8fd8ef,stroke:#0277a8,color:#08303d
    classDef htmlFile fill:#ffc48a,stroke:#c2410c,color:#431407
    classDef folder fill:#f0c558,stroke:#8a6d1c,color:#332a05
```

`Generate` builds the static site from the root directory into the output directory.

## Steps

1. **Scan** the root for documentation files (see [the scanner](scan.go.md)).
2. **Render** each file — Markdown to HTML, or `.mmd` into a Mermaid block (see [the renderer](render.go.md)). Pages that contain a Mermaid diagram pull in `assets/mermaid.min.js`. It also collects the relative paths of pages with **no** Mermaid diagram and returns them — the caller (`cmd/barnacle/main.go.md`) prints them as a non-fatal stderr note, never an error.
3. **Prepare** the output directory. If it holds barnacle's `.barnacle-site` marker, it is wiped first so deleted pages disappear; otherwise it is created and barnacle's files are overwritten one by one.
4. **Write** `assets/barnacle.css` (and `mermaid.min.js` when needed), every page, a landing `index.html`, and the marker file.

## Return value

`Generate` returns `(pages int, missingMermaid []string, err error)`. `pages` counts every written page including `index.html`; `missingMermaid` lists the scanned pages whose rendered HTML contains no `<pre class="mermaid">` (standalone `.mmd` pages never qualify — they always contain one). Callers use it for guidance only; the slice never affects the build outcome.

## Pages

Every page is the same shell: a sidebar with a tree of all documented pages (links relative to the page; the current page is marked `aria-current="page"`) plus the page's content. The landing page lists every page with its source path.

Relative links between pages are just relative links between the mirrored directories, so the generated site works from any base URL — open it locally, publish it with GitHub Pages, or host it anywhere static files are served.

`.md`/`.mmd` pairs with the same base name would both produce the same `.html` file, so `Generate` fails with an explicit collision error instead of silently overwriting one of them.

## Where it is used

Called exactly once per run by `main` in [cmd/barnacle/main.go.md](../../cmd/barnacle/main.go.md) — nowhere else. It is the assembly point: everything else in the package is a helper it wires together.

## Related pages

- [The scanner](scan.go.md) — produces the document list
- [The renderer](render.go.md) — renders each document's body
- [Navigation and landing page](nav.go.md) — sidebar and `index.html` body
- [The page template](template.go.md) — the HTML shell every page shares
- [Configuration](config.go.md) — what `Generate` receives as `cfg`
- [README.md](README.md) — package hub
