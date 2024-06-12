package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func (env *Env) addHostFlag() *string {
	return env.flagSet.String("host", "", "(REQUIRED) a valid $GONF_CONFIG/hosts/ subdirectory (overrides the GONF_HOST environment variable)")
}

func (env *Env) hostFlagToHostDir(hostFlag *string) (string, error) {
	if *hostFlag == "" {
		*hostFlag = env.getenv("GONF_HOST")
		if *hostFlag == "" {
			return "", fmt.Errorf("the GONF_HOST environment variable is unset and the -host FLAG is missing. Please use one or the other")
		}
	}
	hostDir := filepath.Join(env.configDir, "hosts", *hostFlag)
	if info, err := os.Stat(hostDir); err != nil {
		return "", fmt.Errorf("invalid host name %s: %w", *hostFlag, err)
	} else if !info.IsDir() {
		return "", fmt.Errorf("invalid host name %s: %s is not a directory", *hostFlag, hostDir)
	}
	return hostDir, nil
}
