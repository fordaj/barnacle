# The page template (`template.go`)

`template.go` holds the single HTML page shell every generated page shares, rendered with `html/template`.

## `pageData`

The template's data:

- `SiteTitle` — from the `.barnacle.json` `title` (see [config.md](config.md)); shown as the sidebar site title and in `<title>`.
- `PageTitle` — the page's own title (`<title>` and heading content come from the page body itself).
- `Content` — `template.HTML` page body (already-rendered HTML from [render.md](render.md) or [MermaidPageHTML](render.md)), injected verbatim.
- `Nav` — `template.HTML` sidebar navigation from `renderNav` (see [nav.md](nav.md)).
- `RelPrefix` — `"../"` repeated per level of nesting (`""` for top-level pages, `"../"` for `parser/parser.html`); prefixes the `assets/barnacle.css` and `assets/mermaid.min.js` URLs so relative links work from any depth. Computed by `relPrefix` in [site.md](site.md).
- `HasMermaid` — when true, the shell adds `<script src="{{.RelPrefix}}assets/mermaid.min.js">` and calls `mermaid.initialize({ startOnLoad: true, securityLevel: 'strict' })`.

## The shell

A fixed `<!DOCTYPE html>` document: `<meta name="generator" content="barnacle">`, a sidebar `<aside>` with the site title linked to `index.html` and the nav, a `<main class="content">`, and a footer linking to the barnacle repository. `<title>` is `PageTitle · SiteTitle`, collapsing to just `PageTitle` when the two are equal (the landing page).

Page content is inserted as `template.HTML` — it is trusted HTML produced by the renderer (goldmark with `WithUnsafe`), not user input from the browser. Everything else (titles, nav) is escaped by `html/template` or `html.EscapeString` upstream.

## Where it is used

`renderPage` in [site.go](site.md) executes `pageTemplate` once per page, including `index.html`.

## Related pages

- [The site generator](site.md) — executes the template via `renderPage`
- [Navigation and landing page](nav.md) — produces `Nav` and the index content
- [The renderer](render.md) — produces `Content` and reports `HasMermaid`
- [Configuration](config.md) — supplies `SiteTitle`
- [README.md](README.md) — package hub
