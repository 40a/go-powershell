// Copyright (c) 2017 Gorillalabs. All rights reserved.

package powershell

import (
	"strconv"
	"strings"
)

const (
	HTTPPort  = 5985
	HTTPSPort = 5986
)

type Config struct {
	ComputerName          string
	AllowRedirection      bool
	Authentication        string
	CertificateThumbprint string
	Credential            interface{}
	Port                  int
	UseSSL                bool
}

func NewDefaultConfig() *Config {
	return &Config{}
}

func (c *Config) ToArgs() []string {
	args := make([]string, 0)

	if c.ComputerName != "" {
		args = append(args, "-ComputerName")
		args = append(args, QuoteArg(c.ComputerName))
	}

	if c.AllowRedirection {
		args = append(args, "-AllowRedirection")
	}

	if c.Authentication != "" {
		args = append(args, "-Authentication")
		args = append(args, QuoteArg(c.Authentication))
	}

	if c.CertificateThumbprint != "" {
		args = append(args, "-CertificateThumbprint")
		args = append(args, QuoteArg(c.CertificateThumbprint))
	}

	if c.Port > 0 {
		args = append(args, "-Port")
		args = append(args, strconv.Itoa(c.Port))
	}

	if asserted, ok := c.Credential.(string); ok {
		args = append(args, "-Credential")
		args = append(args, asserted) // do not quote, as it contains a variable name when using password auth
	}

	if c.UseSSL {
		args = append(args, "-UseSSL")
	}

	return args
}

func QuoteArg(s string) string {
	return "'" + strings.Replace(s, "'", "\"", -1) + "'"
}
