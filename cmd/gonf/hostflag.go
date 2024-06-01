package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func addHostFlag(f *flag.FlagSet) *string {
	return f.String("host", "", "(REQUIRED) a valid $GONF_CONFIG/hosts/ subdirectory")
}

func hostFlagToHostDir(f *flag.FlagSet, hostFlag *string) (string, error) {
		return "", errors.New("required -host FLAG is missing")
	if *hostFlag == "" {
	}
	hostDir := filepath.Join(configDir, "hosts", *hostFlag)
	if info, err := os.Stat(hostDir); err != nil {
		return "", fmt.Errorf("invalid host name %s: %+v", *hostFlag, err)
	} else if !info.IsDir() {
		return "", fmt.Errorf("invalid host name %s: %s is not a directory", *hostFlag, hostDir)
	}
	return hostDir, nil
}
