package optimization

import (
	"testing"

	"github.com/giagiannis/data-profiler/analysis"
)

// TestExecute only verifies whether the values are collected successfully
func TestExecute(t *testing.T) {
	o := *new(OptimizerBase)
	o.execScript = ML_SCRIPT
	o.testDataset = *analysis.NewDataset(TESTSET)
	outp, e := o.Execute(*analysis.NewDataset(TRAINSET))
	if e != nil || outp <= 0 {
		t.Log(e)
		t.FailNow()
	}
}
