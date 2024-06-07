package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func addHostFlag(f *flag.FlagSet) *string {
	return f.String("host", "", "(REQUIRED) a valid $GONF_CONFIG/hosts/ subdirectory (overrides the GONF_HOST environment variable)")
}

func hostFlagToHostDir(hostFlag *string,
	getenv func(string) string,
) (string, error) {
	if *hostFlag == "" {
		*hostFlag = getenv("GONF_HOST")
		if *hostFlag == "" {
			return "", fmt.Errorf("the GONF_HOST environment variable is unset and the -host FLAG is missing. Please use one or the other")
		}
	}
	hostDir := filepath.Join(configDir, "hosts", *hostFlag)
	if info, err := os.Stat(hostDir); err != nil {
		return "", fmt.Errorf("invalid host name %s: %w", *hostFlag, err)
	} else if !info.IsDir() {
		return "", fmt.Errorf("invalid host name %s: %s is not a directory", *hostFlag, hostDir)
	}
	return hostDir, nil
}
