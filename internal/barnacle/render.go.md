# Renderer

```mermaid
flowchart TD
    MD["Markdown source"] --> GM[goldmark + GFM]
    GM --> Links["RewriteDocLink pass<br/>href rewriting"]
    Links --> Mermaid["Mermaid block pass<br/>fences become pre.mermaid"]
    Mermaid --> HTML[HTML fragment]
    click Links href "#link-rewriting"
    click Mermaid href "#mermaid"
    class MD mdFile
    class GM,Links,Mermaid goFile
    class HTML htmlFile
    classDef mdFile fill:#dcd6ce,stroke:#6b6255,color:#292524
    classDef goFile fill:#8fd8ef,stroke:#0277a8,color:#08303d
    classDef htmlFile fill:#ffc48a,stroke:#c2410c,color:#431407
```

Markdown is rendered to HTML with [goldmark](https://github.com/yuin/goldmark) (GFM: tables, strikethrough, task lists, autolinked URLs), plus two post-processing passes.

## Link rewriting

`RewriteDocLink` rewrites only _relative_ link destinations whose path ends in `.md`, `.markdown`, or `.mmd` (case-insensitive), replacing the extension with `.html`. Fragments and query strings are preserved. Absolute URLs (`https://…`, `mailto:…`), protocol-relative URLs, and pure `#fragment` links are left untouched. Because the generated site mirrors the source directory layout, no path arithmetic is needed — only the extension changes. `MarkdownToHTML` applies this to every `href="…"` in the rendered output via `rewriteDocLinks`.

## Mermaid

Fenced ` ```mermaid ` code blocks become `<pre class="mermaid">` elements, which mermaid.js (bundled at `assets/mermaid.min.js`) renders in the browser. Standalone `.mmd` files are placed in a `<pre class="mermaid">` block directly — see [the site generator](site.go.md). The block contents stay HTML-escaped; mermaid reads them via `textContent`, which unescapes the entities.

`HasMermaid` reports whether rendered HTML contains a `<pre class="mermaid">`; `Generate` uses it to decide whether to ship `mermaid.min.js` and to flag the page in the template.

### `click` directives (clickable nodes)

Mermaid flowcharts natively support `click` directives, which make nodes navigable. `rewriteMermaidLinks` rewrites documentation targets (`.md`/`.markdown`/`.mmd`) inside `click` directives to the generated `.html` pages, the same invariant as `RewriteDocLink` — relative targets are resolved against the page URL, and the site mirrors the source tree, so swapping the extension is enough. Tooltips, external URLs and JavaScript callbacks are left alone; blocks are unescaped and re-escaped around the rewrite (`escapeMermaidText`), so entity round-trips are safe. Works in fenced blocks, in hand-authored `<pre class="mermaid">` blocks, and in standalone `.mmd` files (every block is rewritten exactly once). Note: JavaScript callbacks (`click A callback`, `click A call fn()`) require Mermaid's `securityLevel: 'loose'`, which barnacle does not enable — node navigation via `href` works under `strict`, since relative URLs pass Mermaid's sanitizer. For transitions, see "HTML in diagram labels" below.

### HTML in diagram labels (clickable transitions)

Mermaid has no edge-`click` grammar, but edge labels render as real HTML inside the SVG (mermaid's `htmlLabels` mode, under `securityLevel: 'strict'`, through DOMPurify — relative URLs pass, `javascript:` is stripped). So an inline anchor in a label renders as a clickable HTML link, and barnacle rewrites its target like everything else:

```text
Idle --> Waiting : <a href="dispatch.md">send(cmd)</a>
Aim   --> Launch : <a href="launch.md" title="Launch sequence">go</a>
```

Mechanics:

- Every `href="..."` token in the block is rewritten to `.html` when it ends in `.md`/`.markdown`/`.mmd` (case-insensitive), fragments and query strings preserved — the same invariant as Markdown links and node `click` directives. In the rendered output, `&quot;`-escaping is applied/unapplied around the pass so attribute quotes round-trip cleanly.
- Attributes the author puts inside the anchor (`title`, `target="_blank"`, fragments, …) pass through DOMPurify untouched in `strict`.
- Mermaid's own quirks apply to the source syntax: state diagrams cannot contain `:` in transition text; flowchart labels with parentheses must be quoted (`-->|"<a href='x.md'>send(cmd)</a>"|`), using single quotes inside the double-quoted label.
- Node `click` directives keep working; inline anchors are for transitions and inline marks.

## Related pages

- [The scanner](scan.go.md) — finds the `.md`/`.mmd` files this renderer turns into pages
- [The site generator](site.go.md) — calls `MarkdownToHTML` / `MermaidPageHTML` per page
- [The page template](template.go.md) — injects the fragment and loads mermaid.js when needed
- [README.md](README.md) — package hub
