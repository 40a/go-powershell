// Copyright (c) 2017 Gorillalabs. All rights reserved.

package powershell

import (
	"fmt"

	"github.com/juju/errors"
)

type credential interface {
	prepare(s *Session) (interface{}, error)
}

type UserPasswordCredential struct {
	Username string
	Password string
}

func (c *UserPasswordCredential) prepare(s *Session) (interface{}, error) {
	_, _, err := s.executeDirectly(fmt.Sprintf("$gorillaPassword = ConvertTo-SecureString -String %s -AsPlainText -Force", QuoteArg(c.Password)))
	if err != nil {
		return nil, errors.Annotate(err, "Could not convert password to secure string")
	}

	_, _, err = s.executeDirectly(fmt.Sprintf("$gorillaCredential = New-Object -TypeName 'System.Management.Automation.PSCredential' -ArgumentList %s, $gorillaPassword", QuoteArg(c.Username)))
	if err != nil {
		return nil, errors.Annotate(err, "Could not create PSCredential object")
	}

	return "$gorillaCredential", nil
}
