// Copyright (c) 2017 Gorillalabs. All rights reserved.

package powershell

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"

	"github.com/juju/errors"
)

const newline = "\r\n"

type Shell struct {
	handle *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
}

// Start starts a powershell process and then waits for input commands.
func Start() (*Shell, error) {
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

	return &Shell{cmd, stdin, stdout, stderr}, nil
}

// Execute runs a single command and returns its stdout, stderr and possibly
// an error. Apart from general failures, the function will also return an
// error if the stderr is not empty.
func (s *Shell) Execute(cmd string) (string, string, error) {
	if s.handle == nil {
		return "", "", errors.Annotate(errors.New(cmd), "Cannot execute commands on closed shells.")
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

// Exit closes the powershell process and leaves the shell struct in a state
// where it cannot be used anymore. You need to create a new shell struct by
// calling Start() again.
func (s *Shell) Exit() {
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
	return "gorilla" + createRandomString(16)
}
