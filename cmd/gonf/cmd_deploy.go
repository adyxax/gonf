package main

import (
	"flag"
	"log/slog"
	"path/filepath"
)

func (env *Env) cmdDeploy() error {
	env.flagSet.Init(`gonf deploy [-FLAG]
where FLAG can be one or more of`, flag.ContinueOnError)
	hostFlag := env.addHostFlag()
	env.flagSet.SetOutput(env.stderr)
	_ = env.flagSet.Parse(env.args)
	if env.helpMode {
		env.flagSet.SetOutput(env.stdout)
		env.flagSet.Usage()
	}
	hostDir, err := env.hostFlagToHostDir(hostFlag)
	if err != nil {
		env.flagSet.Usage()
		return err
	}
	return env.runDeploy(*hostFlag, hostDir)
}

func (env *Env) runDeploy(hostFlag string, hostDir string) error {
	sshc, err := env.newSSHClient(hostFlag + ":22")
	if err != nil {
		slog.Error("deploy", "action", "newSshClient", "error", err)
		return err
	}
	defer func() {
		if e := sshc.Close(); err == nil {
			err = e
		}
	}()

	if err = sshc.SendFile(filepath.Join(hostDir, hostFlag)); err != nil {
		slog.Error("deploy", "action", "SendFile", "error", err)
	}

	return err
}
