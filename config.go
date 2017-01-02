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

// Config represents the possible options for establishing remote
// connects (i.e. they wrap the options for the New-PSSession cmdlet).
type Config struct {
	ComputerName          string
	AllowRedirection      bool
	Authentication        string
	CertificateThumbprint string
	Credential            interface{}
	Port                  int
	UseSSL                bool
}

// NewDefaultConfig returns an empty configuration. In the future, there
// might be additional, pre-selected values, but right now the struct is just
// empty.
func NewDefaultConfig() *Config {
	return &Config{}
}

// ToArgs turns the configuration into a string slice containing all the
// configured options.
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

// QuoteArg can be used to quote a PowerShell argument. Note that the resulting
// string will reproduce the input verbatim, so you cannot use this to quote
// variables.
func QuoteArg(s string) string {
	return "'" + strings.Replace(s, "'", "\"", -1) + "'"
}
