package main

import (
	"fmt"
	"runtime/debug"
)

var buildRevision, buildTime, buildModified string

func init() {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				buildRevision = setting.Value
			case "vcs.modified":
				buildModified = setting.Value
			case "vcs.time":
				buildTime = setting.Value
			}
		}
	}
}

func cmdVersion() {
	modified := "clean"
	if buildModified == "true" {
		modified = "dirty"
	}
	fmt.Printf("gonf - %s %s %s\n", buildRevision, buildTime, modified)
}
