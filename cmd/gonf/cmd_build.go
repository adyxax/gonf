package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
)

func (env *Env) cmdBuild() error {
	env.flagSet.Init(`gonf build [-FLAG]
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
	return env.runBuild(hostDir)
}

func (env *Env) runBuild(hostDir string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(wd)
	if err = os.Chdir(hostDir); err != nil {
		return err
	}
	cmd := exec.CommandContext(env.ctx, "go", "build", "-ldflags", "-s -w -extldflags \"-static\"", hostDir)
	cmd.Env = append(cmd.Environ(), "CGO_ENABLED=0")
	if out, err := cmd.CombinedOutput(); err != nil {
		_, _ = fmt.Fprint(env.stderr, string(out))
		return err
	}
	return nil
}
