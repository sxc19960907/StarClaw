#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_DIR="$(mktemp -d)"
VERSION="${LOCAL_RELEASE_VERSION:-v0.0.0-local}"
BASE_URL="${NPM_SMOKE_BASE_URL:-http://127.0.0.1:7533}"

cleanup() {
  curl -fsS -X POST "$BASE_URL/shutdown" >/dev/null 2>&1 || true
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

fail() {
  echo "smoke_npm_install: $*" >&2
  exit 1
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "missing required command: $1"
}

expect_contains() {
  local name="$1"
  local needle="$2"
  if ! grep -Fq -- "$needle" "$TMP_DIR/$name.out"; then
    echo "---- $name output ----" >&2
    cat "$TMP_DIR/$name.out" >&2
    fail "$name output did not contain: $needle"
  fi
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
require_cmd npm

os="$(archive_os)"
arch="$(archive_arch)"
goos="$(echo "$os" | tr '[:upper:]' '[:lower:]')"
goarch="$arch"
if [[ "$goarch" == "x86_64" ]]; then
  goarch="amd64"
fi

binary="starclaw"
archive="$TMP_DIR/starclaw_${os}_${arch}.tar.gz"
if [[ "$os" == "Windows" ]]; then
  binary="starclaw.exe"
  archive="$TMP_DIR/starclaw_${os}_${arch}.zip"
  require_cmd zip
else
  require_cmd tar
fi

build_dir="$TMP_DIR/build"
project_dir="$TMP_DIR/project"
smoke_home="$TMP_DIR/home"
mkdir -p "$build_dir" "$project_dir" "$smoke_home/.starclaw"

cat > "$smoke_home/.starclaw/config.yaml" <<'YAML'
provider: ollama
ollama_endpoint: http://127.0.0.1:1
ollama_model: smoke-test
api_key: dummy
audit:
  enabled: false
YAML

ldflags=(
  "-X" "main.Version=$VERSION"
  "-X" "github.com/starclaw/starclaw/cmd.Version=$VERSION"
)

echo "==> building local release binary ($os/$arch, version $VERSION)"
(
  cd "$ROOT_DIR"
  CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" go build -ldflags "${ldflags[*]}" -o "$build_dir/$binary" .
)

if [[ "$os" == "Windows" ]]; then
  (cd "$build_dir" && zip -q "$archive" "$binary")
else
  tar -C "$build_dir" -czf "$archive" "$binary"
fi

echo "==> packing npm package"
package_tgz="$TMP_DIR/starclaw-cli.tgz"
(
  cd "$ROOT_DIR/npm"
  npm pack --silent --pack-destination "$TMP_DIR" >/dev/null
)
packed="$(find "$TMP_DIR" -maxdepth 1 -name 'starclaw-cli-*.tgz' | sort | head -n 1)"
[[ -n "$packed" ]] || fail "npm pack did not produce package"
mv "$packed" "$package_tgz"

echo "==> installing npm package with local archive"
(
  cd "$project_dir"
  npm init -y --silent >/dev/null
  STARCLAW_NPM_ARCHIVE="$archive" npm install --silent "$package_tgz"
)

echo "==> checking npm-installed starclaw"
env HOME="$smoke_home" "$project_dir/node_modules/.bin/starclaw" version > "$TMP_DIR/version.out"
expect_contains version "starclaw version $VERSION"

env HOME="$smoke_home" "$project_dir/node_modules/.bin/starclaw" app --check > "$TMP_DIR/app-check.out"
expect_contains app-check "StarClaw app launch readiness"
expect_contains app-check "Web UI:        $BASE_URL/app/"
expect_contains app-check "Config:        $smoke_home/.starclaw/config.yaml"

echo "smoke_npm_install: ok"
