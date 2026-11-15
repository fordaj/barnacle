# AGENTS.md

## Project Overview

barnacle is a documentation-site generator written in Go. It turns Markdown (`.md`) and Mermaid (`.mmd`) files scattered through a repository into a static, navigable HTML website — documentation lives next to the code it describes.

It ships as a single dependency-light binary: the only direct dependency is `github.com/yuin/goldmark` (GFM Markdown). Mermaid diagrams are rendered in the browser from a bundled `mermaid.min.js`.

## Core Principles

- **Documentation lives with the code.** Sibling `.md`/`.mmd` files are the contract; the tool reads them in place, never a separate docs tree.
- **Static, portable output.** No server, no database; the generated site works from any base URL (`file://`, GitHub Pages, any static host).
- **Fail loudly.** A missing documented file is an error; a `.md`/`.mmd` stem collision fails the whole build; an explicitly requested config that is missing or invalid is an error.
- **No magic.** The HTML shell and stylesheet are embedded with `go:embed`; the scan excludes `.git`, `node_modules`, `vendor`, `dist`, and the output directory — nothing else is hidden.
- **Dogfood the toolchain.** The repository formats itself with formattington via `uv`, releases with versioner, logs with logamuffin, and renders its own documentation with the pinned barnacle release.

## Build and Test

Requires the Go toolchain from `go.mod` (Go 1.27) and, for formatting, `uv`.

```bash
go build ./cmd/barnacle   # build the binary
go test ./...             # run the test suite
go vet ./...              # static analysis
./scripts/format.sh             # format in place (prettier, black, shfmt -i 2, taplo)
./ci/format_check.sh    # check formatting without modifying files (CI)
scripts/build.sh          # cross-compile into dist/<platform>/barnacle{,.exe}
scripts/docs.sh           # regenerate the documentation site into ./site/
```

CI (`.github/workflows/ci.yml`) runs on self-hosted macOS and Windows runners: `./ci/format_check.sh`, `go vet`, `gofmt -l .`, `go test ./...`, and `go build ./...`. The runners are managed with [runner-spawner](https://github.com/fordaj/runner-spawner) (see `.runner-spawner.json` and `tools/runner-spawner/`).

Releases go through `scripts/release.sh` (`--build-only` and `--dry-run` are the safe entry points); it computes the version with versioner, verifies `barnacle --version` agrees, and publishes per-platform archives. Do not tag or release without an explicit request.

## Repository Layout

```text
cmd/barnacle/     the process boundary (main.go — flag parsing, exit codes)
internal/barnacle/ the whole implementation (config, scan, render, nav, site, template)
scripts/          build/release/docs tooling (+ hidden .common.env / .<platform>.env cascade)
tools/            pinned helper binaries, each with an update.sh (not Go deps)
ci/               format check for CI
scripts/format.sh  formatter (formattington via uv)
site/             generated documentation (git-ignored)
```

Important files include `cmd/barnacle/main.go`, `internal/barnacle/{config,scan,render,nav,site,template}.go`, `internal/barnacle/assets/{barnacle.css,mermaid.min.js}` (embedded), and the scripts under `scripts/`.

## Documentation files (barnacle)

The repository's documentation is a set of Markdown (`.md`) and Mermaid (`.mmd`) source files rendered into a static site by the pinned [barnacle](https://github.com/fordaj/barnacle) binary (`tools/barnacle/`, wired into the environment cascade as `BARNACLE_BIN`). Generate the site with `scripts/docs.sh`; the output (`./site/`) is git-ignored and never committed.

Layout — documentation lives next to the code it describes:

- One sibling `.md` per Go file in `cmd/barnacle/` and `internal/barnacle/` (for example `scan.go` → `scan.md`), plus `internal/barnacle/README.md` as the package hub.
- One sibling `.md` per bash script (`scripts/entry.sh` → `entry.md`, `scripts/env.sh` → `env.md`, `scripts/version.sh` → `version.md`, `scripts/docs.sh` → `docs.md`, `scripts/build.sh` → `build.md`, `scripts/release.sh` → `release.md`, `scripts/format.sh` → `scripts/format.md`), plus `scripts/README.md` as the hub.
- One sibling `.md` per bundle in `tools/` (`tools/<name>/update.md`), plus `tools/README.md` as the hub. `tools/` is excluded from the generated site via `.barnacle.json` (the pinned binaries carry their own upstream docs), so these pages exist for the repository, not the site.

Rules for keeping these files current:

- When you add, rename, remove, or repurpose a documented file — a Go file, a script, or a bundle — update its sibling `.md` **in the same change**. A new file without its page is incomplete work; a stale page is worse than none.
- Source every fact: the `.go`/`.sh`/`.env` file itself, this file (AGENTS.md), and README.md are the accepted sources. Never invent flags, exit codes, extensions, or bundle-platform matrices.
- Use plain relative links (`scan.md`, `../../README.md`) with the linking file's `.md` path — barnacle rewrites them to `.html`. Do not link to files barnacle does not render (binaries, `.sh`, `LICENSE`); link the related `.md` page (for external references, a full `https://` URL is fine). Pages that are not rendered (`tools/`) may still be linked from other unrendered pages; rendered pages must not hyperlink into `tools/` — refer to those pages as code (for example `tools/barnacle/update.md`), because `.barnacle.json` excludes them from the site.
- A `foo.md` and a `foo.mmd` may not coexist for the same stem — barnacle fails the whole site rather than overwriting.
- Hub pages (`README.md` in each documented directory) list their siblings; adding a page means linking it from the hub too, in the same order as the rest of the directory's list.
- Barnacle skips hidden files by default, so the `.common.env` and `.<platform>.env` cascade is documented inside `scripts/env.md` — keep it accurate there instead of adding hidden doc files.
- Regenerate and self-review the site before publishing: `scripts/docs.sh`, then click through the changed pages. The generator output must remain free of broken local links.

Definition of Done addition: a change that touches any documented file updates (or creates) its sibling page in the same commit.

## Definition of Done

A change is complete when it:

- Implements the requested behavior.
- Preserves the CLI contract (`--config`, `--site`, `--version`, exit codes `0`/`1`/`2`) unless intentionally changed.
- Handles missing files, invalid configuration, and IO errors explicitly, without panics.
- Remains cross-platform (path handling via `path/filepath`, no shelling out unless explicitly required).
- Includes or updates tests where appropriate (`go test ./...` passes).
- Leaves the formatting clean (`./ci/format_check.sh`).
- Updates relevant documentation — including the sibling `.md` pages above for every Go, bash, or bundle file touched, and `scripts/docs.sh` re-run with no broken local links.
