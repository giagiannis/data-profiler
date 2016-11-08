package main

import (
	"fmt"
	"os"
)

var utilsDescription = map[string]string{
	"help":         "prints this help message",
	"partition":    "split a dataset file into more smaller files",
	"smap":         "creates a heatmap of the datasets according to their accuracy",
	"similarities": "calculates and stores the similarity matrix of the specified datasets",
	"train":        "exhaustively trains the specified ML job with the datasets and creates a scores matrix",
}

func main() {
	// consume the first command
	if len(os.Args) < 2 || os.Args[1] == "help" {
		fmt.Fprintf(os.Stderr, "Usage: %s [command]\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "List of commands:")
		for name, description := range utilsDescription {
			fmt.Fprintf(os.Stderr, "\t%s - %s\n", name, description)
		}
		os.Exit(1)
	}
	command := os.Args[1]
	os.Args = os.Args[1:]

	if command == "partition" {
		partitionerRun()
	} else if command == "smap" {
		smapRun()
	} else {
		fmt.Fprintln(os.Stderr, "Command not identified")
	}

}
