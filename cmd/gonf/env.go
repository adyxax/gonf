package main

import (
	"context"
	"flag"
	"io"
)

type Env struct {
	batchMode bool
	ctx       context.Context
	configDir string
	flagSet   *flag.FlagSet
	helpMode  bool
	args      []string
	getenv    func(string) string
	stdout    io.Writer
	stderr    io.Writer
}
