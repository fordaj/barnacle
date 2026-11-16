# Repository script tooling (`scripts/`)

```mermaid
flowchart TD
    Entry["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M5 8l4.5 4L5 16'/><path d='M12 16.5h7'/></svg><svg class='label-icon' viewBox='0 0 24 24' width='12' height='12' style='position:absolute;right:-31px;top:-14px' fill='none' stroke='currentColor' stroke-width='2.4' stroke-linecap='round' stroke-linejoin='round'><path d='M10 14a5 5 0 0 0 7.07 0L20.6 10.4a5 5 0 0 0-7.07-7.07l-2 2'/><path d='M14 10a5 5 0 0 0-7.07 0L3.4 13.6a5 5 0 0 0 7.07 7.07l2-2'/></svg>entry.sh<br/>sources env.sh + log()</span>"] --> Env["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M5 8l4.5 4L5 16'/><path d='M12 16.5h7'/></svg><svg class='label-icon' viewBox='0 0 24 24' width='12' height='12' style='position:absolute;right:-31px;top:-14px' fill='none' stroke='currentColor' stroke-width='2.4' stroke-linecap='round' stroke-linejoin='round'><path d='M10 14a5 5 0 0 0 7.07 0L20.6 10.4a5 5 0 0 0-7.07-7.07l-2 2'/><path d='M14 10a5 5 0 0 0-7.07 0L3.4 13.6a5 5 0 0 0 7.07 7.07l2-2'/></svg>env.sh<br/>cascade locator/anchor</span>"]
    Env --> Common["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M5 8l4.5 4L5 16'/><path d='M12 16.5h7'/></svg>.common.env</span>"]
    Env --> Plat["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M5 8l4.5 4L5 16'/><path d='M12 16.5h7'/></svg>.os-arch.env<br/>one per platform</span>"]
    Common --> Tools["<span style='position:relative;display:inline-block'><svg class='label-icon' viewBox='0 0 24 24' width='15' height='15' style='position:absolute;left:-32px;top:-15px' fill='none' stroke='currentColor' stroke-width='1.9' stroke-linecap='round' stroke-linejoin='round'><path d='M3.5 5.7A2.2 2.2 0 0 1 5.7 3.5h4L12 6h5.3a2.2 2.2 0 0 1 2.2 2.2v8.6a2.2 2.2 0 0 1-2.2 2.2H5.7a2.2 2.2 0 0 1-2.2-2.2z'/></svg>tools/&lt;name&gt;/v*<br/>pinned binaries</span>"]
    Plat --> Tools
    click Entry href "entry.sh.md"
    click Env href "env.sh.md"
    class Entry,Env,Common,Plat shFile
    class Tools folder
    classDef shFile fill:#abe07a,stroke:#3f6212,color:#16280a
    classDef folder fill:#f0c558,stroke:#8a6d1c,color:#332a05
```

The `scripts/` directory carries the repository's build/release/docs tooling: an environment cascade the other scripts source, plus one entry point per task. Everything here is bash; nothing requires a global install beyond what each script documents.

Vernacular used in the headers (`log INFO`, `$REPO_ROOT_DIR`, `VERSIONER_BIN`): read [entry.sh.md](entry.sh.md) first, then [env.sh.md](env.sh.md).

## The cascade (why the dotfiles exist)

- `scripts/entry.sh` — sources `scripts/env.sh` and defines `log()` (logamuffin, with a printf fallback). Sourced by everything else. See [entry.sh.md](entry.sh.md).
- `scripts/env.sh` — locates and sources the cascade below, then anchors the repository-relative `*_BIN` tool paths to absolute paths. See [env.sh.md](env.sh.md).
- `scripts/.common.env`, `scripts/.<os>-<arch>.env` — the cascade proper. Hidden files, so barnacle does not render them as pages; [env.sh.md](env.sh.md) is their documentation. `.common.env` holds the platform-independent settings (`TOOLS_DIR=tools`, the `fordaj/*` upstream repos, and the pinned `*_VERSION_DIR`s: versioner v0.0.3.0.5, logamuffin v0.1.3.0.35, runner-spawner v0.0.1.2.6) — barnacle itself is not pinned, `dist/` is built from this tree; each platform file sets `PROJECT_PLATFORM` and the `*_BIN` paths for its `<os>-<arch>`, with `BARNACLE_BIN` defaulting to `dist/<platform>/barnacle[.exe]`.

## Task scripts

- [build.sh](build.sh.md) — cross-compile barnacle for the six supported platforms into `dist/`, stamping the versioner-computed version into each binary.
- [version.sh](version.sh.md) — print the repository version computed by versioner (`--format text|shell|json`).
- [release.sh](release.sh.md) — version, verify, cross-compile, package, tag, push, and publish the GitHub release.
- [docs.sh](docs.sh.md) — generate the documentation site you are reading, with the barnacle built from this tree (`dist/`).
- [format.sh](format.sh.md) — format the repository in place with formattington.

## Refreshing the pinned tools

The `tools/<name>/update.sh` scripts refresh the pinned helper bundles (versioner, logamuffin, runner-spawner, formattington-lite). They are not rendered as part of this site (`tools/` is excluded in `.barnacle.json`); their pages live under `tools/` and are listed in `tools/README.md`. Barnacle is not one of them — its binary comes from `scripts/build.sh`.

## Related pages

- [../README.md](../README.md) — the repository README
- [../AGENTS.md](../AGENTS.md) — the documentation rules these pages follow
