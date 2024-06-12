package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
)

type sshClient struct {
	agentConn net.Conn
	client    *ssh.Client
	ctx       context.Context
	getenv    func(string) string
	stdout    io.Writer
	stderr    io.Writer
}

func (env *Env) newSSHClient(destination string) (*sshClient, error) {
	sshc := sshClient{
		agentConn: nil,
		client:    nil,
		ctx:       env.ctx,
		getenv:    env.getenv,
		stdout:    env.stdout,
		stderr:    env.stderr,
	}
	var err error

	socket := sshc.getenv("SSH_AUTH_SOCK")
	if sshc.agentConn, err = net.Dial("unix", socket); err != nil {
		return nil, fmt.Errorf("failed to open SSH_AUTH_SOCK: %w", err)
	}
	agentClient := agent.NewClient(sshc.agentConn)

	hostKeyCallback, err := knownhosts.New(filepath.Join(sshc.getenv("HOME"), ".ssh/known_hosts"))
	if err != nil {
		return nil, fmt.Errorf("failed to create hostkeycallback function: %w", err)
	}

	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeysCallback(agentClient.Signers),
		},
		HostKeyCallback: hostKeyCallback,
	}
	sshc.client, err = ssh.Dial("tcp", destination, config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}
	return &sshc, nil
}

func (sshc *sshClient) Close() error {
	if err := sshc.client.Close(); err != nil {
		return err
	}
	if err := sshc.agentConn.Close(); err != nil {
		return err
	}
	return nil
}

func (sshc *sshClient) SendFile(filename string) error {
	session, err := sshc.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create ssh client session: %w", err)
	}
	defer session.Close()

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	w, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("sshClient failed to open session stdin pipe: %w", err)
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	errCh := make(chan error, 2)

	session.Stdout = sshc.stdout
	session.Stderr = sshc.stderr
	if err := session.Start("scp -t /usr/local/bin/gonf-run"); err != nil {
		return fmt.Errorf("failed to run scp: %w", err)
	}
	go func() {
		defer wg.Done()
		if e := session.Wait(); e != nil {
			errCh <- e
		}
	}()

	go func() {
		defer wg.Done()
		defer w.Close()
		// Write "C{mode} {size} {filename}\n"
		if _, e := fmt.Fprintf(w, "C%#o %d %s\n", 0700, fi.Size(), "gonf-run"); e != nil {
			errCh <- e
			return
		}
		// Write the file's contents.
		if _, e := io.Copy(w, file); e != nil {
			errCh <- e
			return
		}
		// End with a null byte.
		if _, e := fmt.Fprint(w, "\x00"); e != nil {
			errCh <- e
		}
	}()

	ctx, cancel := context.WithTimeout(sshc.ctx, 60*time.Second)
	defer cancel()

	// wait for all waitgroup.Done() or the timeout
	c := make(chan struct{})
	go func() {
		wg.Wait()
		close(c)
	}()
	select {
	case <-c:
	case <-ctx.Done():
		return ctx.Err()
	}

	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}
