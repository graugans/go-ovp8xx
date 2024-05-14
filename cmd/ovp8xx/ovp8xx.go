package main

import (
	"github.com/graugans/go-ovp8xx/v2/cmd/ovp8xx/cmd"
	"github.com/graugans/go-ovp8xx/v2/internal/versioninfo"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// If the version is "dev", it means that the binary is built using "go install",
	// "go build" or "go run".
	// However, if the binary is build by Goreleaser we use that version.
	if version == "dev" {
		version = versioninfo.Version
		commit = versioninfo.Revision
		date = versioninfo.LastCommit.String()
	}

	cmd.SetVersionInfo(
		version,
		commit,
		date,
	)
	cmd.Execute()
}
