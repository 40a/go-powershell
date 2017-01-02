// Copyright (c) 2017 Gorillalabs. All rights reserved.

package powershell

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/juju/errors"
)

const newline = "\r\n"

type Session struct {
	handle *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
}

func EnterSession(config *Config) (*Session, error) {
	cmd := exec.Command("powershell.exe", "-NoExit", "-Command", "-")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, errors.Annotate(err, "Could not get hold of the PowerShell's stdin stream")
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errors.Annotate(err, "Could not get hold of the PowerShell's stdout stream")
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, errors.Annotate(err, "Could not get hold of the PowerShell's stderr stream")
	}

	err = cmd.Start()
	if err != nil {
		return nil, errors.Annotate(err, "Could not spawn PowerShell process")
	}

	session := &Session{cmd, stdin, stdout, stderr}

	// setup the remote session
	err = session.prepare(config)
	if err != nil {
		return nil, errors.Annotate(err, "Could not prepapre remote session")
	}

	return session, nil
}

func (s *Session) prepare(config *Config) error {
	asserted, ok := config.Credential.(credential)
	if ok {
		credentialParamValue, err := asserted.prepare(s)
		if err != nil {
			return errors.Annotate(err, "Could not setup credentials")
		}

		config.Credential = credentialParamValue
	}

	args := strings.Join(config.ToArgs(), " ")
	_, _, err := s.executeDirectly("$gorillaSession = New-PSSession " + args)

	return err
}

func (s *Session) Execute(cmd string) (string, string, error) {
	return s.executeDirectly(fmt.Sprintf("Invoke-Command -Session $gorillaSession -Script {%s}", cmd))
}

func (s *Session) executeDirectly(cmd string) (string, string, error) {
	if s.handle == nil {
		return "", "", errors.Annotate(errors.New(cmd), "Cannot execute commands on closed sessions.")
	}

	outBoundary := createBoundary()
	errBoundary := createBoundary()

	// wrap the command in special markers so we know when to stop reading from the pipes
	full := fmt.Sprintf("%s; echo '%s'; [Console]::Error.WriteLine('%s')%s", cmd, outBoundary, errBoundary, newline)

	_, err := s.stdin.Write([]byte(full))
	if err != nil {
		return "", "", errors.Annotate(errors.Annotate(err, cmd), "Could not send PowerShell command")
	}

	// read stdout and stderr
	sout := ""
	serr := ""

	waiter := &sync.WaitGroup{}
	waiter.Add(2)

	go streamReader(s.stdout, outBoundary, &sout, waiter)
	go streamReader(s.stderr, errBoundary, &serr, waiter)

	waiter.Wait()

	if len(serr) > 0 {
		return sout, serr, errors.Annotate(errors.New(cmd), serr)
	}

	return sout, serr, nil
}

func (s *Session) Exit() {
	s.executeDirectly("Disconnect-PSSession -Session $gorillaSession")

	s.stdin.Write([]byte("exit" + newline))
	s.stdin.Close()
	s.handle.Wait()

	s.handle = nil
	s.stdin = nil
	s.stdout = nil
	s.stderr = nil
}

func streamReader(stream io.Reader, boundary string, buffer *string, signal *sync.WaitGroup) error {
	// read all output until we have found our boundary token
	output := ""
	bufsize := 64
	marker := boundary + newline

	for {
		buf := make([]byte, bufsize)
		read, err := stream.Read(buf)
		if err != nil {
			return err
		}

		output = output + string(buf[:read])

		if strings.HasSuffix(output, marker) {
			break
		}
	}

	*buffer = strings.TrimSuffix(output, marker)
	signal.Done()

	return nil
}

func createBoundary() string {
	c := 16
	b := make([]byte, c)

	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	return "gorilla" + hex.EncodeToString(b)
}
