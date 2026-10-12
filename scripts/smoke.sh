#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

bin=$(mktemp -d)
trap 'rm -rf "$bin"' EXIT
go build -o "$bin/gr4vy" .
"$bin/gr4vy" version
"$bin/gr4vy" --help >/dev/null
"$bin/gr4vy" transactions list --help >/dev/null
