package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/giagiannis/data-profiler/core"
)

var (
	input         *string
	output        *string
	splits        *int
	partitionType *core.PartitionerType
)

func partitionerParseParams() {
	input = flag.String("i", "", "Input file to partition")
	output = flag.String("o", "", "Input file to partition")
	splits = flag.Int("c", 0, "Number of splits to create")
	part := flag.String("t", "UNIFORM", "Type of partitioning")
	if *part == "UNIFORM" {
		partitionType = new(core.PartitionerType)
		*partitionType = core.UNIFORM
	}
	flag.Parse()
	if *input == "" || *output == "" || *splits == 0 || *part == "" {
		fmt.Println("Please type -h to see usage")
		os.Exit(1)
	}
}

func partitionerRun() {
	partitionerParseParams()
	partitioner := core.NewDatasetPartitioner(
		*input,
		*output,
		*splits,
		*partitionType)

	partitioner.Partition()
	fmt.Println("Partitioning finished")
}
