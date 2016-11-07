package main

import (
	"fmt"
	"os"
)

func main() {
	// consume the first command
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Please provide command argument")
		os.Exit(1)
	}
	command := os.Args[1]
	os.Args = os.Args[1:]

	if command == "partition" {
		partitionerRun()
	} else if command == "help" || command == "list" {
		fmt.Fprintln(os.Stderr, "List of commands:")
		fmt.Fprintln(os.Stderr, "\thelp")
		fmt.Fprintln(os.Stderr, "\tpartition")
	} else {
		fmt.Fprintln(os.Stderr, "Command not identified")
	}

}
