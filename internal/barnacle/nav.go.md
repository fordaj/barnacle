# Navigation and landing page (`nav.go`)

```mermaid
flowchart LR
    Docs["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M4.5 5h15v14h-15z'/><path d='M8 14.5v-4l2.2 2.8 2.2-2.8v4'/></svg>Sorted documents</span>"] --> Tree["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M8.5 7 4 12l4.5 5'/><path d='M15.5 7 20 12l-4.5 5'/></svg><svg class='label-icon' viewBox='0 0 24 24' width='12' height='12' style='position:absolute;right:-31px;top:-14px' fill='none' stroke='currentColor' stroke-width='2.4' stroke-linecap='round' stroke-linejoin='round'><path d='M10 14a5 5 0 0 0 7.07 0L20.6 10.4a5 5 0 0 0-7.07-7.07l-2 2'/><path d='M14 10a5 5 0 0 0-7.07 0L3.4 13.6a5 5 0 0 0 7.07 7.07l2-2'/></svg>buildTree</span>"]
    Tree --> Nav["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M8.5 7 4 12l4.5 5'/><path d='M15.5 7 20 12l-4.5 5'/></svg>renderNav<br/>sidebar per page</span>"]
    Tree --> Index["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M8.5 7 4 12l4.5 5'/><path d='M15.5 7 20 12l-4.5 5'/></svg>renderIndex<br/>landing page body</span>"]
    Nav --> Page["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M8.5 7 4 12l4.5 5'/><path d='M15.5 7 20 12l-4.5 5'/></svg><svg class='label-icon' viewBox='0 0 24 24' width='12' height='12' style='position:absolute;right:-31px;top:-14px' fill='none' stroke='currentColor' stroke-width='2.4' stroke-linecap='round' stroke-linejoin='round'><path d='M10 14a5 5 0 0 0 7.07 0L20.6 10.4a5 5 0 0 0-7.07-7.07l-2 2'/><path d='M14 10a5 5 0 0 0-7.07 0L3.4 13.6a5 5 0 0 0 7.07 7.07l2-2'/></svg>pageTemplate</span>"]
    Index --> Landing["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><rect x='3.5' y='5' width='17' height='14' rx='2'/><path d='M3.5 9.5h17'/></svg>index.html</span>"]
    click Tree href "site.go.md"
    click Page href "template.go.md"
    class Docs mdFile
    class Tree,Nav,Index,Page goFile
    class Landing htmlFile
    classDef mdFile fill:#dcd6ce,stroke:#6b6255,color:#292524
    classDef goFile fill:#8fd8ef,stroke:#0277a8,color:#08303d
    classDef htmlFile fill:#ffc48a,stroke:#c2410c,color:#431407
```

`nav.go` renders the sidebar navigation and the body of the landing page `index.html`. Both are built from the same directory tree derived from the scanned documents.

## `buildTree`

`buildTree(docs []Document) *treeNode` splits each document's `RelPath` on `/` and inserts it into a tree of `treeNode` values (`name`, `isDir`, `doc`, `children`). Directory and file names that collide on one segment are separate nodes (`isDir` disambiguates). Documents are already sorted by `RelPath` when they reach `buildTree`, so the tree — and everything rendered from it — is deterministic.

## `renderNav`

Renders the sidebar as nested `<ul class="nav-tree">` lists. Links are relative to the page being rendered (`currentOutRel`, slash-separated from the site root), computed with `relativeURL` from [site.go.md](site.go.md). The link for the current page gets `aria-current="page"` and the `current` class; directory entries are plain `<span class="nav-dir-name">` labels, not links.

## `renderIndex`

Renders the landing page body: the site title as `<h1>`, a short explainer sentence, and a `<ul class="doc-index">` of every page — each with its source path as a `doc-path` span. With no documents it emits an `empty-state` paragraph telling the user to add `.md`/`.mmd` files and re-run barnacle.

## Where it is used

`Generate` in [site.go.md](site.go.md) calls `buildTree` once, then `renderNav` per page (and once for `index.html`) and `renderIndex` for the landing page body. The output HTML is injected into the page shell as `template.HTML` — see [the page template](template.go.md).

## Related pages

- [The scanner](scan.go.md) — the `Document` list `buildTree` consumes
- [The site generator](site.go.md) — calls these renderers and supplies `relativeURL`
- [The page template](template.go.md) — the shell the nav and index are injected into
- [README.md](README.md) — package hub
