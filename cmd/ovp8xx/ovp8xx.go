package main

import (
	"github.com/graugans/go-ovp8xx/cmd/ovp8xx/cmd"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.SetVersionInfo(
		version,
		commit,
		date,
	)
	cmd.Execute()
}
