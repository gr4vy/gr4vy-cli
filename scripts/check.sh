#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

go mod download
unformatted=$(gofmt -l .)
if [ -n "$unformatted" ]; then echo "gofmt needed on:"; echo "$unformatted"; exit 1; fi
go vet ./...
go build ./...
go test $(go list ./... | grep -v '/test\(/\|$\)') -count=1
go generate ./...
git diff --exit-code --stat -- internal/commands/generated docs
