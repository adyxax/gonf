package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
)

var (
	batchMode bool
	configDir string
	helpMode  bool
)

func main() {
	if err := run(context.Background(),
		os.Args,
		os.Getenv,
		//os.Getwd,
		//os.Stdin,
		os.Stdout,
		os.Stderr,
	); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context,
	args []string,
	getenv func(string) string,
	//getwd func() (string, error),
	//stdin io.Reader,
	stdout, stderr io.Writer,
) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()
	f := flag.NewFlagSet(`gonf COMMAND [-FLAG]
where COMMAND is one of:
  * build: build configurations for one or more hosts
  * help: show contextual help
  * version: show build version and time
where FLAG can be one or more of`, flag.ContinueOnError)
	f.BoolVar(&batchMode, "batch", false, "skips all questions and confirmations, using the default (safe) choices each time")
	f.BoolVar(&helpMode, "help", false, "show contextual help")
	f.StringVar(&configDir, "config", "", "(REQUIRED for most commands) path to a gonf configurations repository (overrides the GONF_CONFIG environment variable)")
	f.SetOutput(stderr)
	f.Parse(args[1:])

	if f.NArg() < 1 {
		f.Usage()
		return errors.New("No command given")
	}
	cmd := f.Arg(0)
	switch cmd {
	case "help":
		f.SetOutput(stdout)
		f.Usage()
	case "version":
		cmdVersion()
	default:
		if configDir == "" {
			configDir = getenv("GONF_CONFIG")
			if configDir == "" {
				f.Usage()
				return errors.New("The GONF_CONFIG environment variable is unset and the -config FLAG is missing. Please use one or the other.")
			}
		}
		switch cmd {
		default:
			f.Usage()
			return fmt.Errorf("Invalid command: %s", cmd)
		}
	}
	return nil
}
