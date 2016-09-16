package optimization

import (
	"testing"

	"github.com/giagiannis/data-profiler/analysis"
)

const (
	ML_SCRIPT     = "../ml_scripts/cart.sh"
	DATASET_TRAIN = "../datasets/shuttle-train.csv"
	DATASET_TEST  = "../datasets/shuttle-test.csv"
)

// TestExecute only verifies whether the values are collected successfully
func TestExecute(t *testing.T) {
	o := *new(OptimizerBase)
	o.execScript = ML_SCRIPT
	o.testDataset = *analysis.NewDataset(DATASET_TEST)
	outp, e := o.Execute(*analysis.NewDataset(DATASET_TRAIN))
	if e != nil || outp <= 0 {
		t.Log(e)
		t.FailNow()
	}
}
