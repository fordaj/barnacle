#!/usr/bin/env bash
# update.sh — refresh the bundled Lucide icons from lucide-icons/lucide releases.
#
# Downloads just the icons archive (lucide-icons-<tag>.zip, the SVG files) of
# the latest release — or the one named with --tag — verifies its SHA256
# against the digest GitHub reports, and swaps the bundle in atomically:
# stage everything first, then move the old icons aside and the new ones in,
# then delete the old ones. A failed download or verification leaves the
# previous bundle intact.
#
# These icons are not executed by any script; they are a content bundle used
# by this repository's documentation (inlined into diagram labels by
# barnacle's local-SVG inlining). Layout after a successful run:
#   tools/lucide/
#     update.sh
#     VERSION           # the release tag, without a leading v
#     LICENSE           # the upstream slash-delimited license text
#     icons/<name>.svg  # stroke="currentColor", 24x24 viewBox
#
# Usage:
#   tools/lucide/update.sh              # download latest, replace the icons
#   tools/lucide/update.sh --dry-run    # resolve tag + plan, change nothing
#   tools/lucide/update.sh --help       # usage
#
# Prerequisites: gh (https://cli.github.com/), unzip, curl, sha256sum or shasum.
# Logging uses log <level> <message> from scripts/entry.sh (logamuffin, with a
# printf fallback); the script resolves its own directory, so it runs from anywhere.
#
# Env:
#   LUCIDE_REPO  Repository slug (default: lucide-icons/lucide)
#   LUCIDE_TAG   Pinned tag (default: latest release)

set -euo pipefail

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
REPO_ROOT_FOR_ENV=$(cd -- "$SCRIPT_DIR/../.." &>/dev/null && pwd)
if [[ ! -f "$REPO_ROOT_FOR_ENV/scripts/entry.sh" ]]; then
    printf 'update.sh: error: entry point not found: %s\n' "$REPO_ROOT_FOR_ENV/scripts/entry.sh" >&2
    exit 1
fi
# shellcheck disable=SC1091
. "$REPO_ROOT_FOR_ENV/scripts/entry.sh"

DRY_RUN=0
TAG="${LUCIDE_TAG:-}"
REPO="${LUCIDE_REPO:-lucide-icons/lucide}"

usage() {
    cat <<'EOF'
Usage: tools/lucide/update.sh [--dry-run] [--tag TAG] [--repo REPO]

Download the latest Lucide release from GitHub via gh, extract the SVG icons
into tools/lucide/icons/ (replacing the current set atomically), refresh the
bundled LICENSE and VERSION file, and leave any failure before the swap with
the previous bundle intact. The JSON twins of the icons (for JS consumers)
are deliberately not vendored.

Flags:
  --dry-run, -n   Resolve tag + plan the swap, change nothing.
  --tag TAG       Pin a specific release tag (default: latest release).
  --repo REPO     Repository slug (default: lucide-icons/lucide).
  --help, -h      Show this help.

Env:
  LUCIDE_REPO  Repository slug (overridden by --repo).
  LUCIDE_TAG   Pinned tag (overridden by --tag).
EOF
}

while [[ $# -gt 0 ]]; do
    case "$1" in
    --dry-run | -n)
        DRY_RUN=1
        shift
        ;;
    --help | -h)
        usage
        exit 0
        ;;
    --tag)
        [[ $# -ge 2 ]] || {
            log ERROR "--tag requires a value"
            usage >&2
            exit 2
        }
        TAG="$2"
        shift 2
        ;;
    --tag=*)
        TAG="${1#--tag=}"
        shift
        ;;
    --repo)
        [[ $# -ge 2 ]] || {
            log ERROR "--repo requires a value"
            usage >&2
            exit 2
        }
        REPO="$2"
        shift 2
        ;;
    --repo=*)
        REPO="${1#--repo=}"
        shift
        ;;
    *)
        log ERROR "unknown argument: $1"
        usage >&2
        exit 2
        ;;
    esac
done

[[ -n "$REPO" ]] || {
    log ERROR "repository must not be empty"
    exit 2
}

die() {
    log ERROR "$*" || true
    exit 1
}
need_cmd() {
    command -v "$1" >/dev/null 2>&1 || die "required command not found: $1"
}

need_cmd gh
need_cmd unzip
need_cmd curl

TOOL_ROOT="$SCRIPT_DIR"
[[ -d "$TOOL_ROOT" ]] || die "lucide bundle directory not found: $TOOL_ROOT"
log INFO "lucide bundle directory: $TOOL_ROOT"
log INFO "repository: $REPO"

if [[ -z "$TAG" ]]; then
    # The latest release overall may be an @lucide/lab or prerelease tag;
    # look through the newest few and pick the first plain semver-like one.
    TAG="$(gh api "repos/$REPO/releases" --jq '.[].tag_name' | head -20 | grep -E '^[0-9]+(\.[0-9]+)+$' | head -1)" ||
        die "could not determine latest icons release from $REPO (is gh authenticated and the repo visible?)"
fi
[[ -n "$TAG" ]] || die "could not determine release tag from $REPO"
[[ "$TAG" =~ ^[0-9]+(\.[0-9]+)+$ ]] || die "invalid tag '$TAG' (expected X.Y.Z...)"
log INFO "release tag: $TAG"
ICONS_ASSET="lucide-icons-${TAG}.zip"

digest="$(gh api "repos/$REPO/releases/tags/$TAG" --jq ".assets[] | select(.name == \"lucide-icons-$TAG.zip\") | .digest // \"\"")" ||
    die "could not look up assets for $TAG in $REPO"
[[ -n "$digest" ]] || die "asset $ICONS_ASSET not found in release $TAG (got: $(gh api "repos/$REPO/releases/tags/$TAG" --jq '.assets[].name'))"
want_sha="${digest#sha256:}"
log INFO "asset digest: $digest"

if [[ "$DRY_RUN" -eq 1 ]]; then
    log INFO "DRY RUN: no download, extraction, or deletion will happen"
    log INFO "would download $ICONS_ASSET into $TOOL_ROOT (swap in icons/), and update VERSION and LICENSE"
    icons_count="$(gh api "repos/$REPO/git/trees/$TAG?recursive=1" --jq '[.tree[].path | select(startswith("icons/") and endswith(".svg"))] | length')" ||
        die "could not list icons at $TAG"
    ((icons_count > 100)) || die "unexpected icon count at $TAG ($icons_count); refusing to guess"
    log INFO "would vendor $icons_count SVG icons"
    exit 0
fi

# --- Real update: download to a temp dir, stage, then swap atomically. ---
TMP_DIR=""
STAGE_DIR=""
cleanup() {
    [[ -n "${TMP_DIR:-}" && -d "$TMP_DIR" ]] && rm -rf -- "$TMP_DIR"
    [[ -n "${STAGE_DIR:-}" && -d "$STAGE_DIR" ]] && rm -rf -- "$STAGE_DIR"
    return 0
}
trap cleanup EXIT

TMP_DIR="$(mktemp -d)"
STAGE_DIR="$(mktemp -d "$TOOL_ROOT/.tmp.XXXXXX")"

log INFO "downloading $ICONS_ASSET from $REPO $TAG..."
gh release download "$TAG" --repo "$REPO" --dir "$TMP_DIR" --pattern "$ICONS_ASSET" --clobber ||
    die "download of $ICONS_ASSET from $REPO $TAG failed"

if command -v shasum >/dev/null 2>&1; then
    got_sha="$(shasum -a 256 "$TMP_DIR/$ICONS_ASSET" | awk '{print $1}')"
elif command -v sha256sum >/dev/null 2>&1; then
    got_sha="$(sha256sum "$TMP_DIR/$ICONS_ASSET" | awk '{print $1}')"
else
    die "no sha256 tool found (shasum or sha256sum)"
fi
[[ "$got_sha" == "$want_sha" ]] || die "sha256 mismatch for $ICONS_ASSET: got $got_sha, want $want_sha"
log INFO "sha256 verified"

log INFO "extracting icons (svg only) ..."
( cd -- "$TMP_DIR" && unzip -q "$ICONS_ASSET" 'icons/*.svg' )
mv "$TMP_DIR/icons" "$STAGE_DIR/icons"
(($(find "$STAGE_DIR/icons" -type f -name '*.svg' | wc -l) > 100)) || die "icon set looks truncated; refusing to swap"

log INFO "fetching LICENSE at tag $TAG..."
license_url="$(gh api "repos/$REPO/contents/LICENSE?ref=$TAG" --jq .download_url)" ||
    die "could not resolve LICENSE at $TAG"
curl -fsSL "$license_url" -o "$STAGE_DIR/LICENSE" || die "LICENSE download failed"
[[ -s "$STAGE_DIR/LICENSE" ]] || die "LICENSE download empty"

printf '%s\n' "$TAG" > "$STAGE_DIR/VERSION"

# Stage complete; swap in. Move the current bundle aside first so a
# mid-swap failure only needs the temp cleanup to leave things usable.
OLD_DIR=""
swap_old=""
if [[ -d "$TOOL_ROOT/icons" ]]; then
    OLD_DIR="$TMP_DIR/icons-old"
    mv "$TOOL_ROOT/icons" "$OLD_DIR"
fi
[[ -f "$TOOL_ROOT/VERSION" ]] && mv "$TOOL_ROOT/VERSION" "$TMP_DIR/VERSION-old"
[[ -f "$TOOL_ROOT/LICENSE" ]] && mv "$TOOL_ROOT/LICENSE" "$TMP_DIR/LICENSE-old"

if ! mv "$STAGE_DIR/icons" "$TOOL_ROOT/icons" ||
   ! mv "$STAGE_DIR/LICENSE" "$TOOL_ROOT/LICENSE" ||
   ! mv "$STAGE_DIR/VERSION" "$TOOL_ROOT/VERSION"; then
    # Roll everything back to the previous bundle.
    rm -rf -- "$TOOL_ROOT/icons"
    [[ -d "$OLD_DIR" ]] && mv "$OLD_DIR" "$TOOL_ROOT/icons"
    [[ -f "$TMP_DIR/VERSION-old" ]] && mv "$TMP_DIR/VERSION-old" "$TOOL_ROOT/VERSION"
    [[ -f "$TMP_DIR/LICENSE-old" ]] && mv "$TMP_DIR/LICENSE-old" "$TOOL_ROOT/LICENSE"
    die "swap failed; rolled back to the previous bundle"
fi

# Success: drop the old bundle entirely.
rm -rf -- "$OLD_DIR"
log INFO "lucide icons updated to $TAG ($(find "$TOOL_ROOT/icons" -type f -name '*.svg' | wc -l) icons)"
