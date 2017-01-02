// Copyright (c) 2017 Gorillalabs. All rights reserved.

package powershell

import (
	"crypto/rand"
	"encoding/hex"
)

func createRandomString(bytes int) string {
	c := bytes
	b := make([]byte, c)

	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	return hex.EncodeToString(b)
}
