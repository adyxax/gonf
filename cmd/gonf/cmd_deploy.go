package main

import (
	"context"
	"flag"
	"io"
	"log/slog"
	"path/filepath"
)

func cmdDeploy(ctx context.Context,
	f *flag.FlagSet,
	args []string,
	getenv func(string) string,
	stdout, stderr io.Writer,
) error {
	f.Init(`gonf deploy [-FLAG]
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
	_ = hostDir
	return runDeploy(ctx, getenv, stdout, stderr, *hostFlag, hostDir)
}

func runDeploy(ctx context.Context,
	getenv func(string) string,
	stdout, stderr io.Writer,
	hostFlag string,
	hostDir string,
) error {
	sshc, err := newSSHClient(ctx, getenv, hostFlag+":22")
	if err != nil {
		slog.Error("deploy", "action", "newSshClient", "error", err)
		return err
	}
	defer func() {
		if e := sshc.Close(); err == nil {
			err = e
		}
	}()

	if err = sshc.SendFile(ctx, stdout, stderr, filepath.Join(hostDir, hostFlag)); err != nil {
		slog.Error("deploy", "action", "SendFile", "error", err)
	}

	return err
}
