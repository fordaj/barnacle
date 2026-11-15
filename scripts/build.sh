#!/usr/bin/env bash
# scripts/build.sh — cross-compile barnacle for the major operating
# systems and processor architectures.
#
#   scripts/build.sh
#
# Writes dist/<platform>/barnacle (barnacle.exe on Windows) for each
# target platform, using the version computed by versioner (see
# scripts/version.bash) stamped into the binary via -ldflags.
#
# Environment:
#   VERSION   Override the version string embedded in the binaries.

set -euo pipefail

ENTRY_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$ENTRY_DIR/entry.sh"

REPO_ROOT="$REPO_ROOT_DIR"
cd "$REPO_ROOT"

if [[ -z "${VERSION:-}" ]]; then
  eval "$("$ENTRY_DIR/version.bash" --format shell)"
fi
log INFO "building barnacle $VERSION"

if ! command -v go >/dev/null 2>&1; then
  log ERROR "go toolchain not found on PATH"
  exit 1
fi

PLATFORMS=(
  macos-arm64:darwin:arm64
  macos-x86_64:darwin:amd64
  linux-arm64:linux:arm64
  linux-x86_64:linux:amd64
  windows-arm64:windows:arm64
  windows-x86_64:windows:amd64
)

for entry in "${PLATFORMS[@]}"; do
  platform="${entry%%:*}"
  rest="${entry#*:}"
  goos="${rest%%:*}"
  goarch="${rest##*:}"

  out_dir="dist/$platform"
  bin="barnacle"
  if [[ "$goos" == "windows" ]]; then
    bin="barnacle.exe"
  fi
  mkdir -p "$out_dir"

  log INFO "building $out_dir/$bin ($goos/$goarch)"
  GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 \
    go build -trimpath -ldflags "-s -w -X main.version=$VERSION" \
    -o "$out_dir/$bin" ./cmd/barnacle
done

log INFO "done: binaries in dist/"
