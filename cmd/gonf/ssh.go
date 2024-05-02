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
}

func newSSHClient(context context.Context,
	getenv func(string) string,
	destination string,
) (*sshClient, error) {
	var sshc sshClient
	var err error

	socket := getenv("SSH_AUTH_SOCK")
	if sshc.agentConn, err = net.Dial("unix", socket); err != nil {
		return nil, fmt.Errorf("failed to open SSH_AUTH_SOCK: %+v", err)
	}
	agentClient := agent.NewClient(sshc.agentConn)

	hostKeyCallback, err := knownhosts.New(filepath.Join(getenv("HOME"), ".ssh/known_hosts"))
	if err != nil {
		return nil, fmt.Errorf("could not create hostkeycallback function: %+v", err)
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
		return nil, fmt.Errorf("failed to dial: %+v", err)
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

func (sshc *sshClient) SendFile(ctx context.Context,
	stdout, stderr io.Writer,
	filename string,
) (err error) {
	session, err := sshc.client.NewSession()
	if err != nil {
		return fmt.Errorf("sshClient failed to create session: %+v", err)
	}
	defer func() {
		if e := session.Close(); err == nil && e != io.EOF {
			err = e
		}
	}()

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("sshClient failed to open file: %+v", err)
	}
	defer func() {
		if e := file.Close(); err == nil {
			err = e
		}
	}()

	fi, err := file.Stat()
	if err != nil {
		return fmt.Errorf("sshClient failed to stat file: %+v", err)
	}

	w, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("sshClient failed to open session stdin pipe: %+v", err)
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	errCh := make(chan error, 2)

	session.Stdout = stdout
	session.Stderr = stderr
	if err = session.Start("scp -t /usr/local/bin/gonf-run"); err != nil {
		return fmt.Errorf("sshClient failed to run scp: %+v", err)
	}
	go func() {
		defer wg.Done()
		if e := session.Wait(); e != nil {
			errCh <- e
		}
	}()

	go func() {
		defer wg.Done()
		defer func() {
			if e := w.Close(); e != nil {
				errCh <- e
			}
		}()
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

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
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
