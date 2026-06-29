#!/usr/bin/env bash
# CloudOS — Development startup script.
# Starts the kernel with hot-reload via Air.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

echo "==> CloudOS Development Environment"
echo "    Kernel: http://localhost:${CLOUDOS_PORT:-8080}"
echo ""

# Ensure the air binary is available.
if ! command -v air &>/dev/null; then
	echo "Installing air (hot-reload)..."
	go install github.com/air-verse/air@latest
fi

exec air
