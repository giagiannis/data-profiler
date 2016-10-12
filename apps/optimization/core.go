package optimization

import (
	"os/exec"
	"strconv"
	"strings"

	"github.com/giagiannis/data-profiler/analysis"
)

// Optimizer dictates the API of the implementing types
type Optimizer interface {
	// method used to execute the optimization
	Run() bool
	// executes the workflow for a given dataset
	Execute(d analysis.Dataset) float64
}

// Struct that holds the base fields/methods of the Optimizer object
type OptimizerBase struct {
	execScript  string           // script used to test the ML job
	testDataset analysis.Dataset // test dataset
	result      OptimizerResult  // result of the optimization process
}

// OptimizerResult holds the result of the optimization process
type OptimizerResult struct {
	Dataset analysis.Dataset // best dataset found
	Score   float64          // score/fitness of the best dataset
}

// Execute method runs the target ML job and returns an estimation of the error
func (b *OptimizerBase) Execute(d analysis.Dataset) (float64, error) {
	cmd := exec.Command(b.execScript, d.Path(), b.testDataset.Path())
	out, er := cmd.Output()
	if er != nil {
		return -1.0, er
	}
	result, er := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if er != nil {
		return -1.0, er
	}
	return result, nil
}

// Result is a getter function for the optimization result
func (b *OptimizerBase) Result() OptimizerResult {
	return b.result
}
