# Scanner

`Scan` walks the repository root and returns every `.md` and `.mmd` file, sorted by slash-separated relative path.

## Rules

- `.git`, `node_modules`, `vendor`, and `dist` are always excluded.
- Dot-files and dot-directories are skipped unless `includeHidden` is set in the config.
- Additional exclusions come from the `exclude` config list: a bare name matches any path segment; an entry with a slash is matched as a glob against the relative path and skips everything beneath it.
- The output directory is never scanned (it is usually `./site` inside the root).
- Extensions are matched case-insensitively.

## Documents

Each discovered file becomes a `Document`:

- `RelPath` — source path relative to the root (`parser/parser.md`).
- `OutputRel` — generated page path; the extension is replaced with `.html`.
- `Title` — the first level-1 ATX heading for `.md` files, the humanized file name otherwise.

See [the site generator](site.md) for how documents are rendered, and [the renderer](render.md) for how page bodies are produced.
