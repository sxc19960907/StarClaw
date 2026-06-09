#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="${RELEASE_DIST_DIR:-$ROOT_DIR/dist}"
RUN_SNAPSHOT=false
NPM_ONLY=false
ASTRIA_LOCAL=false
RUN_UPDATER_BOUNDARY_SMOKE=false
RUN_UPDATER_DRY_RUN_SMOKE=false
RUN_ASTRIA_COMPATIBILITY_MANIFEST_SMOKE=false
RUN_UPDATER_TRANSACTION_PLAN_SMOKE=false

fail() {
  echo "validate_release_artifacts: $*" >&2
  exit 1
}

for arg in "$@"; do
  case "$arg" in
    --snapshot) RUN_SNAPSHOT=true ;;
    --npm-only) NPM_ONLY=true ;;
    --astria-local) ASTRIA_LOCAL=true ;;
    --updater-boundary-smoke) RUN_UPDATER_BOUNDARY_SMOKE=true ;;
    --updater-dry-run-smoke) RUN_UPDATER_DRY_RUN_SMOKE=true ;;
    --astria-compatibility-manifest-smoke) RUN_ASTRIA_COMPATIBILITY_MANIFEST_SMOKE=true ;;
    --updater-transaction-plan-smoke) RUN_UPDATER_TRANSACTION_PLAN_SMOKE=true ;;
    *) fail "unknown argument: $arg" ;;
  esac
done

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

assert_no_private_release_material() {
  echo "==> checking private release material boundary"
  local found
  found="$(
    find "$ROOT_DIR" \
      \( -path "$ROOT_DIR/.git" -o -path "$ROOT_DIR/build" -o -path "$ROOT_DIR/dist" -o -path "$ROOT_DIR/node_modules" -o -path "$ROOT_DIR/npm/node_modules" \) -prune \
      -o -type f \( \
        -name "*.p8" -o \
        -name "*.p12" -o \
        -name "*.cer" -o \
        -name "*.mobileprovision" -o \
        -name "*.provisionprofile" -o \
        -name "*.key" -o \
        -name "*notary*profile*" -o \
        -name "*keychain*profile*" \
      \) -print
  )"
  [[ -z "$found" ]] || fail "private signing/notarization material must not be committed: $found"
}

astria_updater_metadata_decision() {
  local search_dir="$1"
  local metadata
  metadata="$(
    find "$search_dir" -type f \( \
      -iname "*appcast*.xml" -o \
      -iname "*sparkle*.xml" -o \
      -iname "*update*.json" -o \
      -iname "*updater*.json" \
    \) -print
  )"
  if [[ -z "$metadata" ]]; then
    node - <<'NODE'
console.log(JSON.stringify({
  status: "unavailable_safe",
  replacement: "disabled",
  reason: "Astria updater metadata absent",
}, null, 2));
NODE
    return 0
  fi
  require_cmd node
  METADATA_FILES="$metadata" node - <<'NODE'
const fs = require("fs");
const files = (process.env.METADATA_FILES || "").split(/\n/).filter(Boolean);
const required = [
  "version",
  "artifact_url",
  "checksum_sha256",
  "signature",
  "signature_algorithm",
  "public_key_id",
  "min_app_version",
  "min_daemon_version",
];
const allowedAlgorithms = new Set(["ed25519", "rsa-pss-sha256"]);

function fail(message) {
  console.error(message);
  process.exit(1);
}

function walk(value, path = []) {
  if (!value || typeof value !== "object") return;
  for (const [key, child] of Object.entries(value)) {
    const lower = key.toLowerCase();
    if (
      lower.includes("private") ||
      lower.includes("secret") ||
      lower.includes("notary") ||
      lower.includes("keychain") ||
      lower === "p8" ||
      lower === "p12"
    ) {
      fail(`Astria updater metadata contains private field ${[...path, key].join(".")}`);
    }
    walk(child, [...path, key]);
  }
}

for (const file of files) {
  if (!file.toLowerCase().endsWith(".json")) {
    fail(`Astria updater metadata format is unsupported before signed JSON validation exists: ${file}`);
  }
  let data;
  try {
    data = JSON.parse(fs.readFileSync(file, "utf8"));
  } catch (error) {
    fail(`Astria updater metadata is not valid JSON: ${file}: ${error.message}`);
  }
  walk(data);
  for (const field of required) {
    if (typeof data[field] !== "string" || data[field].trim() === "") {
      fail(`Astria updater metadata ${file} missing required field ${field}`);
    }
  }
  if (!/^https:\/\//.test(data.artifact_url)) {
    fail(`Astria updater metadata ${file} artifact_url must be https`);
  }
  if (!/^[a-fA-F0-9]{64}$/.test(data.checksum_sha256)) {
    fail(`Astria updater metadata ${file} checksum_sha256 must be a 64-character SHA-256 hex digest`);
  }
  if (!allowedAlgorithms.has(data.signature_algorithm)) {
    fail(`Astria updater metadata ${file} signature_algorithm must be one of ${Array.from(allowedAlgorithms).join(", ")}`);
  }
  if (data.auto_install === true || data.replace_app === true || data.replacement_allowed === true) {
    fail(`Astria updater metadata ${file} must not enable app replacement before verified updater implementation exists`);
  }
  if (data.unavailable_safe !== true) {
    fail(`Astria updater metadata ${file} must declare unavailable_safe=true until replacement is implemented`);
  }
  const decision = {
    status: "verified_dry_run",
    replacement: "disabled",
    reason: "Astria updater metadata verified; app replacement is not implemented",
    version: data.version,
    artifact_url: data.artifact_url,
    checksum_sha256: data.checksum_sha256,
    signature_algorithm: data.signature_algorithm,
    public_key_id: data.public_key_id,
    min_app_version: data.min_app_version,
    min_daemon_version: data.min_daemon_version,
    metadata_file: file,
  };
  console.log(JSON.stringify(decision, null, 2));
}
NODE
}

assert_astria_updater_boundary() {
  echo "==> checking Astria updater boundary"
  local search_dir="${1:-${ASTRIA_UPDATER_METADATA_DIR:-$ROOT_DIR/desktop/macos/Astria}}"
  [[ -d "$search_dir" ]] || fail "missing Astria updater metadata search dir: $search_dir"
  local decision
  decision="$(astria_updater_metadata_decision "$search_dir")" || return 1
  if grep -Fq '"status": "unavailable_safe"' <<<"$decision"; then
    echo "==> Astria updater metadata absent; unavailable-safe"
  fi
}

run_astria_updater_dry_run() {
  local search_dir="${1:-${ASTRIA_UPDATER_METADATA_DIR:-$ROOT_DIR/desktop/macos/Astria}}"
  [[ -d "$search_dir" ]] || fail "missing Astria updater metadata search dir: $search_dir"
  astria_updater_metadata_decision "$search_dir"
}

run_astria_updater_boundary_smoke() {
  echo "==> checking Astria updater boundary smoke"
  local tmp
  tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' RETURN

  mkdir -p "$tmp/missing"
  assert_astria_updater_boundary "$tmp/missing"

  mkdir -p "$tmp/unsafe"
  cat > "$tmp/unsafe/update.json" <<'JSON'
{"version":"1.0.0","artifact_url":"https://example.com/Astria.zip"}
JSON
  if (assert_astria_updater_boundary "$tmp/unsafe") >/dev/null 2>"$tmp/unsafe.err"; then
    fail "unsafe Astria updater metadata unexpectedly passed validation"
  fi
  grep -Fq "missing required field checksum_sha256" "$tmp/unsafe.err" || fail "unsafe updater metadata failed for the wrong reason"

  mkdir -p "$tmp/private"
  cat > "$tmp/private/update.json" <<'JSON'
{
  "version": "1.0.0",
  "artifact_url": "https://example.com/Astria.zip",
  "checksum_sha256": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
  "signature": "signed-by-release-key",
  "signature_algorithm": "ed25519",
  "public_key_id": "astria-release-1",
  "min_app_version": "1.0.0",
  "min_daemon_version": "1.0.0",
  "unavailable_safe": true,
  "private_key": "do-not-commit"
}
JSON
  if (assert_astria_updater_boundary "$tmp/private") >/dev/null 2>"$tmp/private.err"; then
    fail "private Astria updater metadata unexpectedly passed validation"
  fi
  grep -Fq "private field private_key" "$tmp/private.err" || fail "private updater metadata failed for the wrong reason"

  mkdir -p "$tmp/safe"
  cat > "$tmp/safe/update.json" <<'JSON'
{
  "version": "1.0.0",
  "artifact_url": "https://example.com/Astria.zip",
  "checksum_sha256": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
  "signature": "signed-by-release-key",
  "signature_algorithm": "ed25519",
  "public_key_id": "astria-release-1",
  "min_app_version": "1.0.0",
  "min_daemon_version": "1.0.0",
  "unavailable_safe": true
}
JSON
  assert_astria_updater_boundary "$tmp/safe"
}

run_astria_updater_dry_run_smoke() {
  echo "==> checking Astria updater dry-run smoke"
  local tmp
  tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' RETURN

  mkdir -p "$tmp/missing"
  local missing
  missing="$(run_astria_updater_dry_run "$tmp/missing")"
  grep -Fq '"status": "unavailable_safe"' <<<"$missing" || fail "missing metadata did not produce unavailable-safe dry-run"
  grep -Fq '"replacement": "disabled"' <<<"$missing" || fail "missing metadata dry-run did not keep replacement disabled"

  mkdir -p "$tmp/valid"
  cat > "$tmp/valid/update.json" <<'JSON'
{
  "version": "1.2.3",
  "artifact_url": "https://example.com/Astria-1.2.3.zip",
  "checksum_sha256": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
  "signature": "signed-by-release-key",
  "signature_algorithm": "ed25519",
  "public_key_id": "astria-release-1",
  "min_app_version": "1.0.0",
  "min_daemon_version": "1.0.0",
  "unavailable_safe": true
}
JSON
  local valid
  valid="$(run_astria_updater_dry_run "$tmp/valid")"
  for required in \
    '"status": "verified_dry_run"' \
    '"replacement": "disabled"' \
    '"version": "1.2.3"' \
    '"min_app_version": "1.0.0"' \
    '"min_daemon_version": "1.0.0"' \
    '"signature_algorithm": "ed25519"'
  do
    grep -Fq "$required" <<<"$valid" || fail "valid updater metadata dry-run missing $required"
  done

  mkdir -p "$tmp/replacement"
  cat > "$tmp/replacement/update.json" <<'JSON'
{
  "version": "1.2.3",
  "artifact_url": "https://example.com/Astria-1.2.3.zip",
  "checksum_sha256": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
  "signature": "signed-by-release-key",
  "signature_algorithm": "ed25519",
  "public_key_id": "astria-release-1",
  "min_app_version": "1.0.0",
  "min_daemon_version": "1.0.0",
  "unavailable_safe": true,
  "replace_app": true
}
JSON
  if (run_astria_updater_dry_run "$tmp/replacement") >/dev/null 2>"$tmp/replacement.err"; then
    fail "replacement-enabled updater metadata unexpectedly produced dry-run success"
  fi
  grep -Fq "must not enable app replacement" "$tmp/replacement.err" || fail "replacement metadata failed for the wrong reason"
}

write_astria_compatibility_manifest() {
  local output="$1"
  local app_version="$2"
  local app_build="$3"
  local daemon_version="$4"
  local source_tag="$5"
  require_cmd node
  node - <<'NODE' "$output" "$app_version" "$app_build" "$daemon_version" "$source_tag"
const fs = require("fs");
const [output, appVersion, appBuild, daemonVersion, sourceTag] = process.argv.slice(2);
const manifest = {
  schema_version: "1",
  product: "Astria",
  local_only: true,
  replacement: "disabled",
  app: {
    version: appVersion,
    build: appBuild,
  },
  daemon: {
    version: daemonVersion,
  },
  compatibility: {
    source_tag: sourceTag,
    app_daemon_match: appVersion === daemonVersion,
    min_app_version: appVersion,
    min_daemon_version: daemonVersion,
  },
};
fs.writeFileSync(output, JSON.stringify(manifest, null, 2) + "\n");
NODE
}

assert_astria_compatibility_manifest() {
  local manifest="$1"
  require_cmd node
  node - <<'NODE' "$manifest"
const fs = require("fs");
const manifestPath = process.argv[2];

function fail(message) {
  console.error(message);
  process.exit(1);
}

let manifest;
try {
  manifest = JSON.parse(fs.readFileSync(manifestPath, "utf8"));
} catch (error) {
  fail(`Astria compatibility manifest is not valid JSON: ${error.message}`);
}

for (const field of ["schema_version", "product", "local_only", "replacement", "app", "daemon", "compatibility"]) {
  if (manifest[field] === undefined || manifest[field] === null) {
    fail(`Astria compatibility manifest missing ${field}`);
  }
}
if (manifest.product !== "Astria") fail("Astria compatibility manifest product must be Astria");
if (manifest.local_only !== true) fail("Astria compatibility manifest must declare local_only=true");
if (manifest.replacement !== "disabled") fail("Astria compatibility manifest must keep replacement disabled");
for (const field of ["version", "build"]) {
  if (typeof manifest.app[field] !== "string" || manifest.app[field].trim() === "") {
    fail(`Astria compatibility manifest missing app.${field}`);
  }
}
if (typeof manifest.daemon.version !== "string" || manifest.daemon.version.trim() === "") {
  fail("Astria compatibility manifest missing daemon.version");
}
for (const field of ["source_tag", "min_app_version", "min_daemon_version"]) {
  if (typeof manifest.compatibility[field] !== "string" || manifest.compatibility[field].trim() === "") {
    fail(`Astria compatibility manifest missing compatibility.${field}`);
  }
}
if (manifest.compatibility.app_daemon_match !== true) {
  fail("Astria compatibility manifest requires matching app and daemon release versions");
}
if (manifest.app.version !== manifest.daemon.version) {
  fail("Astria compatibility manifest app.version must match daemon.version");
}
if (manifest.compatibility.min_app_version !== manifest.app.version) {
  fail("Astria compatibility manifest compatibility.min_app_version must match app.version");
}
if (manifest.compatibility.min_daemon_version !== manifest.daemon.version) {
  fail("Astria compatibility manifest compatibility.min_daemon_version must match daemon.version");
}
NODE
}

run_astria_compatibility_manifest_smoke() {
  echo "==> checking Astria compatibility manifest smoke"
  local tmp
  tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' RETURN

  local matching="$tmp/astria-compatibility.json"
  write_astria_compatibility_manifest "$matching" "1.2.3" "456" "1.2.3" "v1.2.3"
  assert_astria_compatibility_manifest "$matching"

  local mismatched="$tmp/astria-compatibility-mismatch.json"
  write_astria_compatibility_manifest "$mismatched" "1.2.3" "456" "1.2.4" "v1.2.3"
  if (assert_astria_compatibility_manifest "$mismatched") >/dev/null 2>"$tmp/mismatch.err"; then
    fail "mismatched Astria compatibility manifest unexpectedly passed validation"
  fi
  grep -Fq "matching app and daemon release versions" "$tmp/mismatch.err" || fail "mismatched compatibility manifest failed for the wrong reason"

  local missing="$tmp/astria-compatibility-missing.json"
  cat > "$missing" <<'JSON'
{"schema_version":"1","product":"Astria","local_only":true,"replacement":"disabled","app":{"version":"1.2.3"},"daemon":{"version":"1.2.3"},"compatibility":{"source_tag":"v1.2.3","app_daemon_match":true,"min_app_version":"1.2.3","min_daemon_version":"1.2.3"}}
JSON
  if (assert_astria_compatibility_manifest "$missing") >/dev/null 2>"$tmp/missing.err"; then
    fail "incomplete Astria compatibility manifest unexpectedly passed validation"
  fi
  grep -Fq "missing app.build" "$tmp/missing.err" || fail "incomplete compatibility manifest failed for the wrong reason"
}

astria_updater_transaction_plan() {
  local metadata_file="$1"
  local manifest_file="$2"
  require_cmd node
  node - <<'NODE' "$metadata_file" "$manifest_file"
const fs = require("fs");
const [metadataPath, manifestPath] = process.argv.slice(2);

function fail(message) {
  console.error(message);
  process.exit(1);
}

function readJSON(path, label) {
  try {
    return JSON.parse(fs.readFileSync(path, "utf8"));
  } catch (error) {
    fail(`Astria updater transaction ${label} is not valid JSON: ${error.message}`);
  }
}

const metadata = readJSON(metadataPath, "metadata");
const manifest = readJSON(manifestPath, "compatibility manifest");
const blockingReasons = [];
const requiredChecks = [
  "metadata_checksum",
  "metadata_signature",
  "public_key_identity",
  "app_daemon_compatibility",
  "rollback_gate",
  "post_update_health_gate",
  "replacement_disabled",
];

function requireString(value, name) {
  if (typeof value !== "string" || value.trim() === "") {
    blockingReasons.push(`missing ${name}`);
  }
}

for (const field of ["version", "artifact_url", "checksum_sha256", "signature", "signature_algorithm", "public_key_id", "min_app_version", "min_daemon_version"]) {
  requireString(metadata[field], `metadata.${field}`);
}
if (!/^[a-fA-F0-9]{64}$/.test(metadata.checksum_sha256 || "")) {
  blockingReasons.push("metadata.checksum_sha256 must be a 64-character SHA-256 hex digest");
}
if (metadata.auto_install === true || metadata.replace_app === true || metadata.replacement_allowed === true) {
  blockingReasons.push("replacement-enabled metadata is blocked before transactional replacement exists");
}
if (metadata.unavailable_safe !== true) {
  blockingReasons.push("metadata.unavailable_safe must remain true");
}
if (manifest.product !== "Astria") {
  blockingReasons.push("compatibility manifest product must be Astria");
}
if (manifest.local_only !== true) {
  blockingReasons.push("compatibility manifest must remain local_only=true");
}
if (manifest.replacement !== "disabled") {
  blockingReasons.push("compatibility manifest replacement must be disabled");
}
if (manifest.compatibility?.app_daemon_match !== true) {
  blockingReasons.push("compatibility manifest must require matching app and daemon versions");
}
if (manifest.app?.version !== metadata.version) {
  blockingReasons.push("metadata.version must match manifest.app.version");
}
if (manifest.daemon?.version !== metadata.version) {
  blockingReasons.push("metadata.version must match manifest.daemon.version");
}
if (manifest.compatibility?.min_app_version !== metadata.min_app_version) {
  blockingReasons.push("metadata.min_app_version must match manifest.compatibility.min_app_version");
}
if (manifest.compatibility?.min_daemon_version !== metadata.min_daemon_version) {
  blockingReasons.push("metadata.min_daemon_version must match manifest.compatibility.min_daemon_version");
}

const transaction = metadata.transaction || {};
const rollbackGate = transaction.rollback_gate || {};
const postUpdateHealthGate = transaction.post_update_health_gate || {};
requireString(transaction.staging_mode, "metadata.transaction.staging_mode");
requireString(transaction.replacement_mode, "metadata.transaction.replacement_mode");
requireString(rollbackGate.id, "metadata.transaction.rollback_gate.id");
requireString(postUpdateHealthGate.id, "metadata.transaction.post_update_health_gate.id");
if (transaction.staging_mode && transaction.staging_mode !== "plan_only") {
  blockingReasons.push("metadata.transaction.staging_mode must be plan_only");
}
if (transaction.replacement_mode && transaction.replacement_mode !== "disabled") {
  blockingReasons.push("metadata.transaction.replacement_mode must be disabled");
}
if (rollbackGate.required !== true) {
  blockingReasons.push("metadata.transaction.rollback_gate.required must be true");
}
if (postUpdateHealthGate.required !== true) {
  blockingReasons.push("metadata.transaction.post_update_health_gate.required must be true");
}

const plan = {
  status: blockingReasons.length === 0 ? "planned_no_replacement" : "blocked",
  planReady: blockingReasons.length === 0,
  localOnly: true,
  replacementEnabled: false,
  version: metadata.version,
  stagingMode: transaction.staging_mode || "missing",
  replacementMode: transaction.replacement_mode || "missing",
  rollbackGate: rollbackGate.id || "missing",
  postUpdateHealthGate: postUpdateHealthGate.id || "missing",
  requiredChecks,
  blockingReasons,
  metadata_file: metadataPath,
  compatibility_manifest_file: manifestPath,
};
if (!plan.planReady) {
  console.error(JSON.stringify(plan, null, 2));
  process.exit(1);
}
console.log(JSON.stringify(plan, null, 2));
NODE
}

run_astria_updater_transaction_plan_smoke() {
  echo "==> checking Astria updater transaction plan smoke"
  local tmp
  tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' RETURN

  local manifest="$tmp/astria-compatibility.json"
  write_astria_compatibility_manifest "$manifest" "1.2.3" "456" "1.2.3" "v1.2.3"

  local valid="$tmp/update-valid.json"
  cat > "$valid" <<'JSON'
{
  "version": "1.2.3",
  "artifact_url": "https://example.com/Astria-1.2.3.zip",
  "checksum_sha256": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
  "signature": "signed-by-release-key",
  "signature_algorithm": "ed25519",
  "public_key_id": "astria-release-1",
  "min_app_version": "1.2.3",
  "min_daemon_version": "1.2.3",
  "unavailable_safe": true,
  "transaction": {
    "staging_mode": "plan_only",
    "replacement_mode": "disabled",
    "rollback_gate": {
      "id": "rollback-manifest-v1",
      "required": true
    },
    "post_update_health_gate": {
      "id": "astria-local-health-v1",
      "required": true
    }
  }
}
JSON
  local plan
  plan="$(astria_updater_transaction_plan "$valid" "$manifest")"
  for required in \
    '"status": "planned_no_replacement"' \
    '"planReady": true' \
    '"localOnly": true' \
    '"replacementEnabled": false' \
    '"stagingMode": "plan_only"' \
    '"rollbackGate": "rollback-manifest-v1"' \
    '"postUpdateHealthGate": "astria-local-health-v1"'
  do
    grep -Fq "$required" <<<"$plan" || fail "valid transaction plan missing $required"
  done

  local replacement="$tmp/update-replacement.json"
  cp "$valid" "$replacement"
  node - <<'NODE' "$replacement"
const fs = require("fs");
const path = process.argv[2];
const data = JSON.parse(fs.readFileSync(path, "utf8"));
data.replace_app = true;
fs.writeFileSync(path, JSON.stringify(data, null, 2) + "\n");
NODE
  if (astria_updater_transaction_plan "$replacement" "$manifest") >/dev/null 2>"$tmp/replacement.err"; then
    fail "replacement-enabled metadata unexpectedly produced a transaction plan"
  fi
  grep -Fq "replacement-enabled metadata is blocked" "$tmp/replacement.err" || fail "replacement transaction failed for the wrong reason"

  local missing_gate="$tmp/update-missing-gate.json"
  cp "$valid" "$missing_gate"
  node - <<'NODE' "$missing_gate"
const fs = require("fs");
const path = process.argv[2];
const data = JSON.parse(fs.readFileSync(path, "utf8"));
delete data.transaction.rollback_gate;
fs.writeFileSync(path, JSON.stringify(data, null, 2) + "\n");
NODE
  if (astria_updater_transaction_plan "$missing_gate" "$manifest") >/dev/null 2>"$tmp/missing-gate.err"; then
    fail "metadata missing rollback gate unexpectedly produced a transaction plan"
  fi
  grep -Fq "missing metadata.transaction.rollback_gate.id" "$tmp/missing-gate.err" || fail "missing rollback gate failed for the wrong reason"
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

if "$RUN_UPDATER_BOUNDARY_SMOKE"; then
  run_astria_updater_boundary_smoke
  echo "validate_release_artifacts: ok"
  exit 0
fi

if "$RUN_UPDATER_DRY_RUN_SMOKE"; then
  run_astria_updater_dry_run_smoke
  echo "validate_release_artifacts: ok"
  exit 0
fi

if "$RUN_ASTRIA_COMPATIBILITY_MANIFEST_SMOKE"; then
  run_astria_compatibility_manifest_smoke
  echo "validate_release_artifacts: ok"
  exit 0
fi

if "$RUN_UPDATER_TRANSACTION_PLAN_SMOKE"; then
  run_astria_updater_transaction_plan_smoke
  echo "validate_release_artifacts: ok"
  exit 0
fi

if "$NPM_ONLY"; then
  validate_npm_package
  if "$ASTRIA_LOCAL"; then
    assert_no_private_release_material
    assert_astria_updater_boundary
    run_astria_compatibility_manifest_smoke
    run_astria_updater_transaction_plan_smoke
    echo "==> checking Astria local shell"
    "$ROOT_DIR/scripts/smoke_macos_astria_shell.sh"
  fi
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

if "$ASTRIA_LOCAL"; then
  assert_no_private_release_material
  assert_astria_updater_boundary
  run_astria_compatibility_manifest_smoke
  run_astria_updater_transaction_plan_smoke
  echo "==> checking Astria local shell"
  "$ROOT_DIR/scripts/smoke_macos_astria_shell.sh"
fi

echo "validate_release_artifacts: ok"
