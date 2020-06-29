// +build !arm !arm64

package main

import (
	"io/ioutil"
	"log"
	"strings"
)

func readToken(fpath string) string {
	b, err := ioutil.ReadFile(fpath)
	if err != nil {
		log.Fatalf("Could not read token file at path %s", fpath)
	}
	return strings.ReplaceAll(string(b), "\n", "")
}

