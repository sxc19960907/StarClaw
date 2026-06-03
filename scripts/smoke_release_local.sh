#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_DIR="$(mktemp -d)"
VERSION="${LOCAL_RELEASE_VERSION:-v0.0.0-local}"

cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

fail() {
  echo "smoke_release_local: $*" >&2
  exit 1
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "missing required command: $1"
}

archive_os() {
  case "$(uname -s)" in
    Darwin) echo "Darwin" ;;
    Linux) echo "Linux" ;;
    MINGW*|MSYS*|CYGWIN*) echo "Windows" ;;
    *) fail "unsupported OS: $(uname -s)" ;;
  esac
}

archive_arch() {
  case "$(uname -m)" in
    arm64|aarch64) echo "arm64" ;;
    x86_64|amd64) echo "x86_64" ;;
    *) fail "unsupported arch: $(uname -m)" ;;
  esac
}

require_cmd go
os="$(archive_os)"
arch="$(archive_arch)"

build_dir="$TMP_DIR/build"
archive_dir="$TMP_DIR/archive"
mkdir -p "$build_dir" "$archive_dir"

binary="starclaw"
goos="$(echo "$os" | tr '[:upper:]' '[:lower:]')"
goarch="$arch"
if [[ "$goarch" == "x86_64" ]]; then
  goarch="amd64"
fi
if [[ "$os" == "Windows" ]]; then
  binary="starclaw.exe"
  require_cmd zip
else
  require_cmd tar
fi

ldflags=(
  "-X" "main.Version=$VERSION"
  "-X" "github.com/starclaw/starclaw/cmd.Version=$VERSION"
)

echo "==> building local release binary ($os/$arch, version $VERSION)"
(
  cd "$ROOT_DIR"
  CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" go build -ldflags "${ldflags[*]}" -o "$build_dir/$binary" .
)

archive="$archive_dir/starclaw_${os}_${arch}.tar.gz"
if [[ "$os" == "Windows" ]]; then
  archive="$archive_dir/starclaw_${os}_${arch}.zip"
  (cd "$build_dir" && zip -q "$archive" "$binary")
else
  tar -C "$build_dir" -czf "$archive" "$binary"
fi

echo "==> running release install smoke against $archive"
RELEASE_ARCHIVE="$archive" "$ROOT_DIR/scripts/smoke_release_install.sh"

echo "smoke_release_local: ok"
