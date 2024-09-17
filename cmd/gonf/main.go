package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
)

func main() {
	env := Env{
		batchMode: false,
		ctx:       context.Background(),
		configDir: "",
		flagSet:   nil,
		helpMode:  false,
		args:      os.Args,
		getenv:    os.Getenv,
		stdout:    os.Stdout,
		stderr:    os.Stderr,
	}
	if err := env.run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func (env *Env) run() error {
	ctx, cancel := signal.NotifyContext(env.ctx, os.Interrupt)
	defer cancel()
	env.ctx = ctx
	env.flagSet = flag.NewFlagSet(`gonf COMMAND [-FLAG]
where COMMAND is one of:
  * build: build configuration for a host
  * deploy: deploy configuration for a host
  * help: show contextual help
  * version: show build version and time
where FLAG can be one or more of`, flag.ContinueOnError)
	env.flagSet.BoolVar(&env.batchMode, "batch", false, "skips all questions and confirmations, using the default (safe) choices each time")
	env.flagSet.BoolVar(&env.helpMode, "help", false, "show contextual help")
	env.flagSet.StringVar(&env.configDir, "config", "", "(REQUIRED for most commands) path to a gonf configurations repository (overrides the GONF_CONFIG environment variable)")
	env.flagSet.SetOutput(env.stderr)
	_ = env.flagSet.Parse(env.args[1:])

	if env.flagSet.NArg() < 1 {
		if env.helpMode {
			env.flagSet.SetOutput(env.stdout)
			env.flagSet.Usage()
			return nil
		}
		env.flagSet.Usage()
		return fmt.Errorf("no command given")
	}
	cmd := env.flagSet.Arg(0)
	env.args = env.flagSet.Args()[1:]
	switch cmd {
	case "help":
		env.flagSet.SetOutput(env.stdout)
		env.flagSet.Usage()
	case "version":
		cmdVersion()
	default:
		if env.configDir == "" {
			env.configDir = env.getenv("GONF_CONFIG")
			if env.configDir == "" {
				env.flagSet.Usage()
				return fmt.Errorf("the GONF_CONFIG environment variable is unset and the -config FLAG is missing. Please use one or the other")
			}
		}
		switch cmd {
		case "build":
			return env.cmdBuild()
		case "deploy":
			return env.cmdDeploy()
		default:
			env.flagSet.Usage()
			return fmt.Errorf("invalid command: %s", cmd)
		}
	}
	return nil
}
