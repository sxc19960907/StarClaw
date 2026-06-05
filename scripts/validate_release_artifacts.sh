#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="${RELEASE_DIST_DIR:-$ROOT_DIR/dist}"
RUN_SNAPSHOT=false
NPM_ONLY=false

for arg in "$@"; do
  case "$arg" in
    --snapshot) RUN_SNAPSHOT=true ;;
    --npm-only) NPM_ONLY=true ;;
    *) fail "unknown argument: $arg" ;;
  esac
done

fail() {
  echo "validate_release_artifacts: $*" >&2
  exit 1
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "missing required command: $1"
}

assert_file() {
  local path="$1"
  [[ -f "$path" ]] || fail "missing artifact: $path"
}

assert_archive_contains() {
  local archive="$1"
  local binary="$2"
  case "$archive" in
    *.tar.gz)
      tar -tzf "$archive" | grep -Eq "(^|/)$binary$" || fail "$archive does not contain $binary"
      ;;
    *.zip)
      require_cmd unzip
      unzip -Z1 "$archive" | grep -Eq "(^|/)$binary$" || fail "$archive does not contain $binary"
      ;;
    *)
      fail "unsupported archive format: $archive"
      ;;
  esac
}

find_one() {
  local pattern="$1"
  find "$DIST_DIR" -maxdepth 2 -type f -name "$pattern" | sort | head -n 1
}

validate_npm_package() {
  require_cmd npm
  require_cmd node
  echo "==> checking npm package"
  local output
  output="$(cd "$ROOT_DIR/npm" && npm pack --dry-run --json)"
  node - <<'NODE' "$output"
const packages = JSON.parse(process.argv[2]);
const files = new Set((packages[0]?.files || []).map((file) => file.path));
for (const file of ["bin/starclaw", "scripts/install.js", "scripts/uninstall.js", "package.json"]) {
  if (!files.has(file)) {
    console.error(`npm package missing ${file}`);
    process.exit(1);
  }
}
NODE
}

if "$NPM_ONLY"; then
  validate_npm_package
  echo "validate_release_artifacts: ok"
  exit 0
fi

if "$RUN_SNAPSHOT"; then
  echo "==> building GoReleaser snapshot"
  require_cmd goreleaser
  (cd "$ROOT_DIR" && goreleaser release --snapshot --clean --skip=publish)
fi

echo "==> checking release archives in $DIST_DIR"
for artifact in \
  "starclaw_Darwin_arm64.tar.gz" \
  "starclaw_Darwin_x86_64.tar.gz" \
  "starclaw_Linux_arm64.tar.gz" \
  "starclaw_Linux_x86_64.tar.gz" \
  "starclaw_Windows_arm64.zip" \
  "starclaw_Windows_x86_64.zip"
do
  assert_file "$DIST_DIR/$artifact"
done

assert_archive_contains "$DIST_DIR/starclaw_Darwin_arm64.tar.gz" "starclaw"
assert_archive_contains "$DIST_DIR/starclaw_Darwin_x86_64.tar.gz" "starclaw"
assert_archive_contains "$DIST_DIR/starclaw_Linux_arm64.tar.gz" "starclaw"
assert_archive_contains "$DIST_DIR/starclaw_Linux_x86_64.tar.gz" "starclaw"
assert_archive_contains "$DIST_DIR/starclaw_Windows_arm64.zip" "starclaw.exe"
assert_archive_contains "$DIST_DIR/starclaw_Windows_x86_64.zip" "starclaw.exe"

echo "==> checking package artifacts"
deb="$(find_one "starclaw_*_linux_*.deb")"
rpm="$(find_one "starclaw_*_linux_*.rpm")"
apk="$(find_one "starclaw_*_linux_*.apk")"
[[ -n "$deb" ]] || fail "missing deb package artifact"
[[ -n "$rpm" ]] || fail "missing rpm package artifact"
[[ -n "$apk" ]] || fail "missing apk package artifact"

validate_npm_package

echo "validate_release_artifacts: ok"
