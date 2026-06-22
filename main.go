// Command gr4vy is the command-line interface for the Gr4vy payment
// orchestration platform.
package main

//go:generate go run ./internal/gen
//go:generate go run ./internal/gendocs

import "github.com/gr4vy/gr4vy-cli/internal/cli"

// Build metadata, populated via -ldflags at release time:
//
//	-X main.version=... -X main.commit=... -X main.date=...
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cli.Execute(cli.BuildInfo{Version: version, Commit: commit, Date: date})
}
