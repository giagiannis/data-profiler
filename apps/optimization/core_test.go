package optimization

import (
	"testing"

	"github.com/giagiannis/data-profiler/core"
)

const (
	ANALYSIS_SCRIPT = "../../r_scripts/analysis/pca.R"
	ML_SCRIPT       = "../../r_scripts/classification/cart.R"
	TESTSET         = "../../_testdata/datatest.csv"
	TRAINSET        = "../../_testdata/datatraining.csv"
)

// TestExecute only verifies whether the values are collected successfully
func TestExecute(t *testing.T) {
	o := *new(OptimizerBase)
	o.execScript = ML_SCRIPT
	o.testDataset = *core.NewDataset(TESTSET)
	outp, e := o.Execute(*core.NewDataset(TRAINSET))
	if e != nil || outp <= 0 {
		t.Log(e)
		t.FailNow()
	}
}
