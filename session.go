// Copyright (c) 2017 Gorillalabs. All rights reserved.

package powershell

import (
	"fmt"
	"strings"

	"github.com/juju/errors"
)

// Session represents a remote session.
type Session struct {
	shell *Shell
	name  string
}

// EnterSession uses a shell to establish a remote session to a different
// computer.
func EnterSession(s *Shell, config *Config) (*Session, error) {
	asserted, ok := config.Credential.(credential)
	if ok {
		credentialParamValue, err := asserted.prepare(s)
		if err != nil {
			return nil, errors.Annotate(err, "Could not setup credentials")
		}

		config.Credential = credentialParamValue
	}

	name := "goSess" + createRandomString(8)
	args := strings.Join(config.ToArgs(), " ")

	_, _, err := s.Execute(fmt.Sprintf("$%s = New-PSSession %s", name, args))
	if err != nil {
		return nil, errors.Annotate(err, "Could not create new PSSession")
	}

	return &Session{s, name}, nil
}

// Execute will run the given command through the remote session.
func (s *Session) Execute(cmd string) (string, string, error) {
	if s.shell == nil {
		return "", "", errors.Annotate(errors.New(cmd), "Cannot execute commands on closed sessions.")
	}

	return s.shell.Execute(fmt.Sprintf("Invoke-Command -Session $%s -Script {%s}", s.name, cmd))
}

// Exit will close the remote session, but leave the underlying shell running.
// After calling this, the session struct is not usable anymore and you need to
// create a new one by calling EnterSession() again.
func (s *Session) Exit() {
	s.shell.Execute(fmt.Sprintf("Disconnect-PSSession -Session $%s", s.name))
	s.shell = nil
}
