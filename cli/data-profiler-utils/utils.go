package main

import (
	"fmt"
	"log"
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

func setLogger(logfile string) {
	if logfile != "" {
		f, er := os.OpenFile(logfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if er != nil {
			fmt.Println(er)
			os.Exit(1)
		} else {
			log.SetOutput(f)
		}
	}
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)

}
