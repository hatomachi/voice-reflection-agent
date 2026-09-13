#!/usr/bin/env bash
set -euo pipefail

# Set up PATH for agy, python3, and git
export PATH="/Users/s-ikari/.local/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin:$PATH"

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_DIR"

echo "=== Voice Reflection Agent Daemon Starting ==="
echo "Working directory: $PROJECT_DIR"
echo "Python: $(which python3)"
echo "AGY: $(which agy || echo 'not in PATH')"

exec /opt/homebrew/bin/python3 -m src.daemon "$@"
