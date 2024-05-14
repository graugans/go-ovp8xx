// Package versioninfo uses runtime.ReadBuildInfo() to set global executable revision information if possible.
package versioninfo

// SPDX-License-Identifier: MIT
// Copyright (c) 2021 Carl Johnson
// Based on https://github.com/earthboundkid/versioninfo

import (
	"runtime/debug"
	"time"
)

var (
	// Version will be the version tag if the binary is built with "go install url/tool@version".
	// If the binary is built some other way, it will be "(devel)".
	Version = "unknown"
	// Revision is taken from the vcs.revision tag in Go 1.18+.
	Revision = "unknown"
	// LastCommit is taken from the vcs.time tag in Go 1.18+.
	LastCommit time.Time
	// DirtyBuild is taken from the vcs.modified tag in Go 1.18+.
	DirtyBuild = true
)

func init() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	for _, kv := range info.Settings {
		if kv.Value == "" {
			continue
		}
		switch kv.Key {
		case "vcs.revision":
			Revision = kv.Value
		case "vcs.time":
			LastCommit, _ = time.Parse(time.RFC3339, kv.Value)
		case "vcs.modified":
			DirtyBuild = kv.Value == "true"
		}
	}
	if info.Main.Version != "" {
		Version = info.Main.Version
	}
}
