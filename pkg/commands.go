package gonf

import (
	"bytes"
	"log/slog"
	"os/exec"
)

// ----- Globals ---------------------------------------------------------------
var commands []*CommandPromise

// ----- Init ------------------------------------------------------------------
func init() {
	commands = make([]*CommandPromise, 0)
}

// ----- Public ----------------------------------------------------------------
func Command(cmd string, args ...string) *CommandPromise {
	return CommandWithEnv([]string{}, cmd, args...)
}

func CommandWithEnv(env []string, cmd string, args ...string) *CommandPromise {
	return &CommandPromise{
		args:   args,
		chain:  nil,
		cmd:    cmd,
		env:    env,
		err:    nil,
		Status: PROMISED,
	}
}

type CommandPromise struct {
	args   []string
	chain  []Promise
	cmd    string
	env    []string
	err    error
	Status Status
	Stdout bytes.Buffer
	Stderr bytes.Buffer
}

func (c *CommandPromise) IfRepaired(p ...Promise) Promise {
	c.chain = p
	return c
}

func (c *CommandPromise) Promise() Promise {
	commands = append(commands, c)
	return c
}

func (c *CommandPromise) Resolve() {
	cmd := exec.Command(c.cmd, c.args...)
	for _, e := range c.env {
		cmd.Env = append(cmd.Environ(), e)
	}
	cmd.Stdout = &c.Stdout
	cmd.Stderr = &c.Stderr

	if c.err = cmd.Run(); c.err != nil {
		c.Status = BROKEN
		slog.Error("command", "args", c.args, "cmd", c.cmd, "env", c.env, "err", c.err, "stdout", c.Stdout.String(), "stderr", c.Stderr.String(), "status", c.Status)
		return
	}
	if c.Stdout.Len() == 0 && c.Stderr.Len() > 0 {
		c.Status = BROKEN
		slog.Error("command", "args", c.args, "cmd", c.cmd, "env", c.env, "stdout", c.Stdout.String(), "stderr", c.Stderr.String(), "status", c.Status)
		return
	}
	c.Status = REPAIRED
	slog.Info("command", "args", c.args, "cmd", c.cmd, "env", c.env, "stderr", c.Stderr.String(), "status", c.Status)
	// TODO add a notion of repaired?
	for _, p := range c.chain {
		p.Resolve()
	}
}

// ----- Internal --------------------------------------------------------------
func resolveCommands() (status Status) {
	status = KEPT
	for _, c := range commands {
		if c.Status == PROMISED {
			c.Resolve()
			switch c.Status {
			case BROKEN:
				return BROKEN
			case REPAIRED:
				status = REPAIRED
			}
		}
	}
	return
}
