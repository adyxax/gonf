package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
)

var (
	hostFlag string
)

func cmdBuild(ctx context.Context,
	f *flag.FlagSet,
	args []string,
	getenv func(string) string,
	stdout, stderr io.Writer,
) error {
	f.Init(`gonf build [-FLAG]
where FLAG can be one or more of`, flag.ContinueOnError)
	f.StringVar(&hostFlag, "host", "", "(REQUIRED) a valid $GONF_CONFIG/hosts/ subdirectory inside your gonf configurations repository")
	f.SetOutput(stderr)
	f.Parse(args)
	if helpMode {
		f.SetOutput(stdout)
		f.Usage()
	}
	if hostFlag == "" {
		f.Usage()
		return errors.New("Required -host FLAG is missing")
	}
	hostDir := configDir + "/hosts/" + hostFlag
	if info, err := os.Stat(hostDir); err != nil {
		f.Usage()
		return fmt.Errorf("Invalid host name %s, the %s directory returned error %+v", hostFlag, hostDir, err)
	} else if !info.IsDir() {
		f.Usage()
		return fmt.Errorf("Invalid host name %s, %s is not a directory", hostFlag, hostDir)
	}
	return runBuild(ctx, hostDir)
}

func runBuild(ctx context.Context, hostDir string) error {
	wd, err := os.Getwd()
	defer os.Chdir(wd)
	os.Chdir(hostDir)
	if err != nil {
		slog.Error("build", "hostDir", hostDir, "error", err)
		return err
	}
	cmd := exec.CommandContext(ctx, "go", "build", "-ldflags", "-s -w -extldflags \"-static\"", hostDir)
	cmd.Env = append(cmd.Environ(), "CGO_ENABLED=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		slog.Error("build", "hostDir", hostDir, "error", err, "combinedOutput", out)
		return err
	}
	slog.Debug("build", "hostDir", hostDir, "combinedOutput", out)
	return nil
}
