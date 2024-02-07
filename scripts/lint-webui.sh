#!/usr/bin/env bash

set -euo pipefail

cd webui
bun install
bun run check-format
bun run lint
