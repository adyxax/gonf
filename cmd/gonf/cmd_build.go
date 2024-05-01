package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
)

func cmdBuild(ctx context.Context,
	f *flag.FlagSet,
	args []string,
	getenv func(string) string,
	stdout, stderr io.Writer,
) error {
	f.Init(`gonf build [-FLAG]
where FLAG can be one or more of`, flag.ContinueOnError)
	hostFlag := addHostFlag(f)
	f.SetOutput(stderr)
	_ = f.Parse(args)
	if helpMode {
		f.SetOutput(stdout)
		f.Usage()
	}
	hostDir, err := hostFlagToHostDir(f, hostFlag)
	if err != nil {
		f.Usage()
		return err
	}
	return runBuild(ctx, stderr, hostDir)
}

func runBuild(ctx context.Context, stderr io.Writer, hostDir string) (err error) {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	defer func() {
		if e := os.Chdir(wd); err == nil {
			err = e
		}
	}()
	if err = os.Chdir(hostDir); err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "go", "build", "-ldflags", "-s -w -extldflags \"-static\"", hostDir)
	cmd.Env = append(cmd.Environ(), "CGO_ENABLED=0")
	if out, err := cmd.CombinedOutput(); err != nil {
		_, _ = fmt.Fprint(stderr, string(out))
		return err
	}
	return nil
}
