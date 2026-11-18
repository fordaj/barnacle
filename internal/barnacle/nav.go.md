# Navigation and landing page (`nav.go`)

```mermaid
flowchart LR
    Docs("<img src='barnacle:file-text.svg' width='20' alt=''/><br/>Sorted documents") --> Tree("<img src='barnacle:file-code.svg' width='20' alt=''/><br/><u>buildTree</u>")
    Tree --> Nav("<img src='barnacle:file-code.svg' width='20' alt=''/><br/>renderNav<br/>sidebar per page")
    Tree --> Index("<img src='barnacle:file-code.svg' width='20' alt=''/><br/>renderIndex<br/>landing page body")
    Nav --> Page("<img src='barnacle:file-code.svg' width='20' alt=''/><br/><u>pageTemplate</u>")
    Index --> Landing("<img src='barnacle:app-window.svg' width='20' alt=''/><br/>index.html")
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

Page titles carry raw Markdown (they come from the `#` heading, backticks and all — see [scan.go.md](scan.go.md)), so in HTML contexts they render single-backtick spans as `<code>` via `renderInlineCode`: everything is HTML-escaped first, then the spans are wrapped. Anything else (links, emphasis) stays literal text. The landing page links do the same (`renderIndexNodes`), and the site title as the page `<h1>` as well; the browser tab (`<title>`) keeps raw text.

## `renderIndex`

Renders the landing page body: the site title as `<h1>`, a short explainer sentence, and a `<ul class="doc-index">` of every page — each with its source path as a `doc-path` span. With no documents it emits an `empty-state` paragraph telling the user to add `.md`/`.mmd` files and re-run barnacle.

## Where it is used

`Generate` in [site.go.md](site.go.md) calls `buildTree` once, then `renderNav` per page (and once for `index.html`) and `renderIndex` for the landing page body. The output HTML is injected into the page shell as `template.HTML` — see [the page template](template.go.md).

## Related pages

- [The scanner](scan.go.md) — the `Document` list `buildTree` consumes
- [The site generator](site.go.md) — calls these renderers and supplies `relativeURL`
- [The page template](template.go.md) — the shell the nav and index are injected into
- [README.md](README.md) — package hub
