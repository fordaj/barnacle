# The `internal/barnacle` package

All of barnacle's implementation lives in this package: configuration, scanning, rendering, HTML assembly, navigation, and the page template. `cmd/barnacle/main.go` is a two-line shim over `Generate` here.

If you want the big picture first, open [cmd/barnacle/main.md](../../cmd/barnacle/main.md), then [site.md](site.md). To refresh the pinned helper binaries, see `tools/README.md` (not part of this site); to render this documentation, see [scripts/docs.md](../../scripts/docs.md).

## Pages

- [Configuration — config.go](config.md) — `.barnacle.json` loading and validation
- [The scanner — scan.go](scan.md) — walk the tree, apply excludes, build `Document`s
- [The renderer — render.go](render.md) — goldmark → HTML, link rewriting, Mermaid blocks
- [Navigation and landing page — nav.go](nav.md) — the sidebar tree and the `index.html` body
- [The page template — template.go](template.md) — the HTML shell shared by every page
- [The site generator — site.go](site.md) — orchestration: scan → render → write `site/`
- [cmd/barnacle/main.go](../../cmd/barnacle/main.md) — the process boundary that calls `Generate`

## How the pieces fit

`Generate` ([site.md](site.md)) is the single entry point: it `Scan`s the tree ([scan.md](scan.md)), renders each document ([render.md](render.md)), builds the navigation tree ([nav.md](nav.md)), and executes the shared page template ([template.md](template.md)) once per page plus the landing page. `LoadConfig` ([config.md](config.md)) is called once by `main` before `Generate`.

## Assets

`assets/barnacle.css` (the site stylesheet) and `assets/mermaid.min.js` (the Mermaid renderer, copied into `assets/` only on pages that contain a diagram) are embedded with `go:embed` (`assetsFS` in [site.md](site.md)) and written into the output on every generation.

## Testing

Tests live next to the code as `*_test.go` (`render_test.go`, `site_test.go`) and use only the standard `testing` package. Run them with `go test ./...` (see the README's "Building from source").
