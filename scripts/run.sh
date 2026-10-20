#!/bin/sh
# Run this checkout's terminal app; .env credentials are read by the app itself.
set -eu
SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
REPO_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)
APP="$REPO_DIR/novel-translate-cli"
if [ ! -x "$APP" ]; then
    command -v go >/dev/null 2>&1 || { echo "Install Go 1.25.5+ or place the Linux/macOS binary at $APP" >&2; exit 1; }
    (cd "$REPO_DIR" && GOWORK=off go build -trimpath -o "$APP" ./cmd/novel-translate-cli)
fi
if [ "$#" -eq 0 ]; then
    set -- tui
fi
exec "$APP" "$@"
