package main

import (
	"fmt"
	"os"
	"strings"
)

const (
	optsDelimiter = ","
	kvDelimiter   = "="
)

func parseOptions(optString string) map[string]string {
	opts := make(map[string]string)
	if optString == "" {
		return opts
	}
	for _, option := range strings.Split(optString, optsDelimiter) {
		sp := strings.Split(option, kvDelimiter)
		if len(sp) < 2 {
			fmt.Fprintln(os.Stderr, "Malformed options")
			os.Exit(1)
		}
		opts[sp[0]] = sp[1]
	}
	fmt.Println(opts)
	return opts
}
