#!/usr/bin/env bash
# scripts/release.sh — compute the next version with versioner (dogfooding),
# cross-compile the binaries, tag HEAD, push the tag, and create a GitHub
# release with the per-platform archives attached.
#
# One archive is produced per platform:
#
#   barnacle-<platform>.tar.gz   (.zip on Windows)
#   containing barnacle (barnacle.exe), LICENSE, README.md and LUCIDE-LICENSE
#
# LUCIDE-LICENSE is the Lucide icon set's license (tools/lucide/LICENSE);
# the icons are embedded in the binary itself, and their license ships in
# every archive alongside it.
#
# The binaries are produced with CGO disabled and are statically linked and
# portable. The program version is dynamic via versioner (.versioner.json);
# the version is stamped into every cross-compiled binary with -ldflags -X,
# so the tag and `barnacle --version` agree by construction.
#
# Usage:
#   scripts/release.sh                # version, build, tag, push, release + upload
#   scripts/release.sh --build-only   # cross-compile into dist/ only
#   scripts/release.sh --dry-run      # compute + verify + build, touch nothing remote
#
# Prerequisites: git, a Go toolchain, tar, and zip (for the Windows
# archives); a real release additionally needs gh authenticated via
# `gh auth login`. Bundled versioner/logamuffin binaries are selected via
# scripts/entry.sh (refresh with tools/versioner/update.sh and
# tools/logamuffin/update.sh). Logging uses log <level> <message> from
# scripts/entry.sh (logamuffin, with printf fallback).
# Working tree must be clean for a real release; HEAD should be pushed.
#
# The script is idempotent: an existing tag is reused, and an existing
# GitHub release receives the archives through "gh release upload", so
# re-running it attaches the assets instead of failing.

set -euo pipefail

# Logging via logamuffin (log <level> <message> from scripts/entry.sh,
# with printf fallback until tools/logamuffin is populated). Source the
# entry point first: it pulls in the env cascade and defines log().
SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
# shellcheck disable=SC1091
source "$SCRIPT_DIR/entry.sh"
REPO_ROOT="$REPO_ROOT_DIR"

DRY_RUN=0
BUILD_ONLY=0

usage() {
  cat <<'EOF'
Usage: scripts/release.sh [--build-only | --dry-run]

Compute the next VERSION/TAG with versioner itself (dogfooding, bundled
binary selected via scripts/entry.sh), cross-compile barnacle for every
supported platform into dist/, verify the program's version matches,
create an annotated tag, push it, and create a GitHub release with the
per-platform archives attached, plus a checksums.txt manifest.

Flags:
  --build-only    Cross-compile into dist/ and stop (no tag/push/release).
  --dry-run, -n   Compute + verify + build, but do not tag/push/release.
  --help, -h      Show this help.

Env:
  VERSIONER_BIN   Path to versioner (default: via scripts/entry.sh, then PATH).
  LOGAMUFFIN_BIN  Path to logamuffin for log() (default: via scripts/entry.sh, then PATH).
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
  --dry-run | -n)
    DRY_RUN=1
    shift
    ;;
  --build-only)
    BUILD_ONLY=1
    shift
    ;;
  --help | -h)
    usage
    exit 0
    ;;
  *)
    log ERROR "unknown argument: $1"
    usage >&2
    exit 2
    ;;
  esac
done

die() {
  log ERROR "$*" || true
  exit 1
}
need_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "required command not found: $1"
}

need_cmd git
need_cmd go
need_cmd tar
need_cmd zip

if [[ "$DRY_RUN" -ne 1 && "$BUILD_ONLY" -ne 1 ]]; then
  need_cmd gh
  if ! gh auth status >/dev/null 2>&1; then
    die "gh is not authenticated (run: gh auth login)"
  fi
fi

cd "$REPO_ROOT"
log INFO "repo: $REPO_ROOT"

BRANCH="$(git rev-parse --abbrev-ref HEAD)"
HEAD_SHORT="$(git rev-parse --short HEAD)"
log INFO "branch: $BRANCH  HEAD: $HEAD_SHORT"
if [[ "$DRY_RUN" -eq 1 ]]; then
  log INFO "DRY RUN: no tag/push/release will happen"
fi

# --- Precondition: clean working tree ---------------------------------
# Only a real release needs a pristine tree; --build-only and --dry-run
# tolerate a dirty one.
if [[ -n "$(git status --porcelain)" ]]; then
  if [[ "$DRY_RUN" -eq 1 || "$BUILD_ONLY" -eq 1 ]]; then
    log WARNING "working tree is not clean (allowed here)"
  else
    die "working tree is not clean (commit or stash first):"$'\n'"$(git status --short)"
  fi
fi

# Best-effort fetch so remote tag-existence checks are current.
if ! git fetch origin --tags 2>/dev/null; then
  if [[ "$DRY_RUN" -eq 1 || "$BUILD_ONLY" -eq 1 ]]; then
    log WARNING "git fetch origin --tags failed; remote tag check may be stale"
  else
    die "git fetch origin --tags failed"
  fi
fi

# --- Dogfooding: versioner/logamuffin via scripts/entry.sh (sourced above)
# shellcheck disable=SC2154
if [[ -z "${VERSIONER_BIN:-}" ]]; then
  die "no versioner binary found (run tools/versioner/update.sh, or set VERSIONER_BIN)"
fi
[[ -x "$VERSIONER_BIN" ]] || die "VERSIONER_BIN=$VERSIONER_BIN is not executable"
log INFO "using versioner: $VERSIONER_BIN"
if [[ -n "${LOGAMUFFIN_BIN:-}" && -x "${LOGAMUFFIN_BIN}" ]]; then
  log INFO "using logamuffin: $LOGAMUFFIN_BIN"
else
  log WARNING "logamuffin not found; falling back to printf logging (run tools/logamuffin/update.sh)"
fi

# --- Compute VERSION/TAG with versioner (dogfooding) ------------------
log INFO "computing version with versioner (dogfooding)..."
VERSION_OUT="$("$VERSIONER_BIN" --repo-dir "$REPO_ROOT" --quiet)"
log INFO "versioner output:"
printf '%s\n' "$VERSION_OUT" | sed 's/^/release.sh:   /'

if ! parse_versioner_output "$VERSION_OUT"; then
  die "could not parse VERSION/TAG from versioner output"
fi
if ! [[ "$TAG" =~ ^v[0-9]+(\.[0-9]+)+$ ]]; then
  die "computed tag looks invalid: '$TAG' (expected vX.Y.Z...)"
fi
log INFO "VERSION=$VERSION  TAG=$TAG"

# --- The binary's reported version must agree with the tag ------------
# The binary reports the version stamped via -ldflags -X, so a mismatch
# means the stamping broke or the binary was built from elsewhere.
log INFO "verifying program version..."
STAGE_DIR="$(mktemp -d)"
cleanup_stage() {
  if [[ -n "${STAGE_DIR:-}" && -d "$STAGE_DIR" ]]; then
    rm -rf -- "$STAGE_DIR"
  fi
}
trap cleanup_stage EXIT

VERIFY_BIN="$STAGE_DIR/barnacle-verify"
go build -trimpath -ldflags "-s -w -X main.version=$VERSION" \
  -o "$VERIFY_BIN" ./cmd/barnacle
PKG_VERSION="$("$VERIFY_BIN" --version)"
log INFO "barnacle --version reports: $PKG_VERSION"
[[ "$PKG_VERSION" == *"$VERSION"* ]] ||
  die "program version '$PKG_VERSION' does not match versioner VERSION=$VERSION"

# --- Cross-compile the release binaries -------------------------------
# dist/<platform>/barnacle{,.exe} is produced by scripts/build.sh; the
# archives are built from a staging tree and land in dist/.
log INFO "cross-compiling into dist/..."
VERSION="$VERSION" "$SCRIPT_DIR/build.sh"

# --- Package the archives ----------------------------------------------
DIST_DIR="$REPO_ROOT/dist"
RELEASE_ASSETS=()

package_tree() {
  local directory="$1"
  local archive="$2"
  local parent base
  parent="$(dirname "$directory")"
  base="$(basename "$directory")"

  if [[ "$archive" == *.zip ]]; then
    (
      cd "$parent" && zip -q -r "$archive" "$base"
    ) || die "could not package $archive"
  else
    tar -czf "$archive" -C "$parent" "$base" ||
      die "could not package $archive"
  fi
}

for platform_dir in "$DIST_DIR"/*/; do
  platform="$(basename "$platform_dir")"
  binary="barnacle"
  extension="tar.gz"
  if [[ -f "$platform_dir/barnacle.exe" ]]; then
    binary="barnacle.exe"
    extension="zip"
  fi
  [[ -f "$platform_dir/$binary" ]] || die "missing binary in $platform_dir"

  tree="$STAGE_DIR/pkg/barnacle-$platform"
  rm -rf "$tree"
  mkdir -p "$tree"
  cp "$platform_dir/$binary" "$tree/$binary"
  cp "$REPO_ROOT/LICENSE" "$tree/LICENSE"
  cp "$REPO_ROOT/README.md" "$tree/README.md"
  cp "$REPO_ROOT/tools/lucide/LICENSE" "$tree/LUCIDE-LICENSE"

  archive="$DIST_DIR/barnacle-$platform.$extension"
  package_tree "$tree" "$archive"
  log INFO "  packaged $(basename "$archive")"
  RELEASE_ASSETS+=("$archive")
done

# Checksums manifest covering the archives (not the per-platform dirs).
(
  cd "$DIST_DIR" || exit 1
  assets=()
  for f in barnacle-*.tar.gz barnacle-*.zip; do
    [[ -f "$f" ]] && assets+=("$f")
  done
  ((${#assets[@]} > 0)) || exit 1
  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "${assets[@]}" >checksums.txt
  else
    sha256sum "${assets[@]}" >checksums.txt
  fi
) || die "could not generate checksums.txt"
RELEASE_ASSETS+=("$DIST_DIR/checksums.txt")
log INFO "  packaged checksums.txt"
log INFO "built ${#RELEASE_ASSETS[@]} release assets in dist/"

if [[ "$BUILD_ONLY" -eq 1 ]]; then
  log INFO "build only: stopping before tag/push/release"
  exit 0
fi

# --- Tag ----------------------------------------------------------------
# The script is idempotent: a tag that already exists is reused instead of
# being recreated, so a release can be re-run to (re)attach the binaries.
TAG_EXISTS_LOCALLY=0
TAG_EXISTS_REMOTELY=0
if git rev-parse -q --verify "refs/tags/$TAG" >/dev/null 2>&1; then
  TAG_EXISTS_LOCALLY=1
fi
if git ls-remote --exit-code --tags origin "refs/tags/$TAG" >/dev/null 2>&1; then
  TAG_EXISTS_REMOTELY=1
fi

if [[ "$DRY_RUN" -eq 1 ]]; then
  log INFO "DRY RUN: would run:"
  if [[ "$TAG_EXISTS_LOCALLY" -eq 1 ]]; then
    log INFO "  git tag -a $TAG -m \"$TAG\"   (skipped: the tag already exists)"
  else
    log INFO "  git tag -a $TAG -m \"$TAG\""
  fi
  if [[ "$TAG_EXISTS_REMOTELY" -eq 1 ]]; then
    log INFO "  git push origin $TAG   (skipped: the tag is already on origin)"
  else
    log INFO "  git push origin $TAG"
  fi
  log INFO "  gh release create $TAG --title \"$TAG\" --generate-notes ${RELEASE_ASSETS[*]}"
  log INFO "  gh release upload $TAG ${RELEASE_ASSETS[*]} --clobber   (only if the release already exists)"
  log INFO "dry run complete: VERSION=$VERSION TAG=$TAG"
  exit 0
fi

if [[ "$TAG_EXISTS_LOCALLY" -eq 0 ]]; then
  log INFO "creating annotated tag $TAG..."
  git tag -a "$TAG" -m "$TAG"
else
  log WARNING "tag $TAG already exists locally; reusing it"
fi

if [[ "$TAG_EXISTS_REMOTELY" -eq 0 ]]; then
  log INFO "pushing tag $TAG..."
  git push origin "$TAG"
else
  log WARNING "tag $TAG is already on origin; not pushing it again"
fi

# --- GitHub release (archives attached) ---------------------------------
if gh release view "$TAG" >/dev/null 2>&1; then
  log INFO "GitHub release $TAG already exists; attaching ${#RELEASE_ASSETS[@]} assets..."
  gh release upload "$TAG" "${RELEASE_ASSETS[@]}" --clobber
  log INFO "attached ${#RELEASE_ASSETS[@]} assets to release $TAG"
else
  log INFO "creating GitHub release $TAG with ${#RELEASE_ASSETS[@]} assets..."
  gh release create "$TAG" \
    --title "$TAG" \
    --generate-notes \
    "${RELEASE_ASSETS[@]}"
fi

if REPO_SLUG="$(gh repo view --json nameWithOwner -q .nameWithOwner 2>/dev/null)" && [[ -n "$REPO_SLUG" ]]; then
  log INFO "release: https://github.com/$REPO_SLUG/releases/tag/$TAG"
else
  log INFO "release $TAG created (see GitHub Releases)"
fi
