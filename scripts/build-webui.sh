#!/usr/bin/env bash

set -euo pipefail

[[ -z "${1:-}" ]] && echo "Usage: $0 <version>" && exit 1

export VERSION="${1}"
echo "building webui ($VERSION)..."

cd webui
bun install
rm -rf out
bun run build
