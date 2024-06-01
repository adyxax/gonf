package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func addHostFlag(f *flag.FlagSet) *string {
	return f.String("host", "", "(REQUIRED) a valid $GONF_CONFIG/hosts/ subdirectory")
}

func hostFlagToHostDir(hostFlag *string) (string, error) {
	if *hostFlag == "" {
		return "", fmt.Errorf("required -host FLAG is missing")
	}
	hostDir := filepath.Join(configDir, "hosts", *hostFlag)
	if info, err := os.Stat(hostDir); err != nil {
		return "", fmt.Errorf("invalid host name %s: %w", *hostFlag, err)
	} else if !info.IsDir() {
		return "", fmt.Errorf("invalid host name %s: %s is not a directory", *hostFlag, hostDir)
	}
	return hostDir, nil
}
