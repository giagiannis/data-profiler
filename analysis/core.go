package analysis

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
)

// Dataset struct represents a dataset object.
type Dataset struct {
	id   string
	path string
}

// NewDataset is the constructor for the Dataset struct. A random ID is assigned
// to a new dataset
func NewDataset(path string) *Dataset {
	buffer := make([]byte, 4)
	rand.Read(buffer)
	id := fmt.Sprintf("%x", buffer)
	d := Dataset{id, path}
	return &d
}

// Id getter for dataset
func (d Dataset) Id() string {
	return d.id
}

// Path getter for dataset
func (d Dataset) Path() string {
	return d.path
}

// String method for dataset object - only returns its id
func (d Dataset) String() string {
	return d.Id()
}

// Analyzer interface that expresses the Analyzer's functionality
type Analyzer interface {
	// executes the analysis - blocking process
	Analyze() bool
	// returns the status of the Analyzer
	Status() Status
	// returns the result of the analyzer - a serialized version of it
	Result() Result
	//returns the duration of the analysis
	Duration() float64
	// returns the dataset of the analysis
	Dataset() Dataset
}

// Status is the type representing the domain of the Analyzer's status
type Status uint8

// Values of the AnalyzerStatus type
const (
	PENDING Status = iota + 1
	ANALYZING
	ANALYZED
)

// String method is used to print the Status enum in a pretty manner
func (s Status) String() string {
	switch s {
	case PENDING:
		return "PENDING"
	case ANALYZING:
		return "ANALYZING"
	case ANALYZED:
		return "ANALYZED"
	}
	return "UNKNOWN"
}

// Result struct holds the results of the analysis
type Result []float64

// Utility function used to partition a dataset into multiple disjoint subsets
// in a uniform manner
func DatasetPartition(original Dataset, copies int) []Dataset {
	// init new files
	newFiles := make([]*os.File, copies)
	results := make([]Dataset, copies)
	for i := 0; i < copies; i++ {
		fileName := fmt.Sprintf("%s-split-%d", original.Path(), i)
		newFiles[i], _ = os.Create(fileName)
		results[i] = *NewDataset(fileName)
	}
	file, _ := os.Open(original.Path())
	scanner := bufio.NewScanner(file)
	scanner.Scan()
	header := scanner.Text()
	for _, f := range newFiles {
		f.WriteString(header + "\n")
	}

	for scanner.Scan() {
		fileChosen := rand.Int() % copies
		newFiles[fileChosen].WriteString(scanner.Text() + "\n")
	}
	for _, f := range newFiles {
		f.Close()
	}
	file.Close()

	return results
}

// Function used to cleanup the file partitions
func DatasetPartitionCleanup(datasets []Dataset) {
	for _, d := range datasets {
		os.Remove(d.Path())
	}
}
