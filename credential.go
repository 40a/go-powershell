// Copyright (c) 2017 Gorillalabs. All rights reserved.

package powershell

import (
	"fmt"

	"github.com/juju/errors"
)

type executor interface {
	Execute(string) (string, string, error)
}

type credential interface {
	prepare(executor) (interface{}, error)
}

type UserPasswordCredential struct {
	Username string
	Password string
}

func (c *UserPasswordCredential) prepare(s executor) (interface{}, error) {
	name := "goCred" + createRandomString(8)
	pwname := "goPass" + createRandomString(8)

	_, _, err := s.Execute(fmt.Sprintf("$%s = ConvertTo-SecureString -String %s -AsPlainText -Force", pwname, QuoteArg(c.Password)))
	if err != nil {
		return nil, errors.Annotate(err, "Could not convert password to secure string")
	}

	_, _, err = s.Execute(fmt.Sprintf("$%s = New-Object -TypeName 'System.Management.Automation.PSCredential' -ArgumentList %s, $%s", name, QuoteArg(c.Username), pwname))
	if err != nil {
		return nil, errors.Annotate(err, "Could not create PSCredential object")
	}

	return fmt.Sprintf("$%s", name), nil
}
