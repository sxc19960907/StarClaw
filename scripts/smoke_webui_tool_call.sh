#!/usr/bin/env bash
set -euo pipefail

WEBUI_SMOKE_MODE=tool_call "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib/webui_smoke_common.sh"
