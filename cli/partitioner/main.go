package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/giagiannis/data-profiler/analysis"
)

var (
	input         *string
	output        *string
	splits        *int
	partitionType *analysis.PartitionerType
)

func parseParams() {
	input = flag.String("i", "", "Input file to partition")
	output = flag.String("o", "", "Input file to partition")
	splits = flag.Int("c", 0, "Number of splits to create")
	part := flag.String("t", "UNIFORM", "Type of partitioning")
	if *part == "UNIFORM" {
		partitionType = new(analysis.PartitionerType)
		*partitionType = analysis.UNIFORM
	}
	flag.Parse()
	if *input == "" || *output == "" || *splits == 0 || *part == "" {
		fmt.Println("Please type -h to see usage")
		os.Exit(1)
	}
}
func main() {
	parseParams()
	partitioner := analysis.NewDatasetPartitioner(
		*input,
		*output,
		*splits,
		*partitionType)

	partitioner.Partition()
	fmt.Println("Partitioning finished")
}
