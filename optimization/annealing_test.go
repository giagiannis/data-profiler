package optimization

import (
	"os"
	"testing"

	"github.com/giagiannis/data-profiler/analysis"
)

const (
	ANALYSIS_SCRIPT       = "../r_scripts/pca.R"
	CLASSIFICATION_SCRIPT = "../ml_scripts/classification.sh"
	TESTSET               = "../datasets/shuttle-test.csv"
	TRAINSET              = "../datasets/shuttle-train.csv"
)

func TestRun(t *testing.T) {
	datasets := analysis.DatasetPartition(*analysis.NewDataset(TRAINSET), 100)
	manager := analysis.NewManager(datasets, 8, ANALYSIS_SCRIPT)
	manager.Analyze()
	o := NewSimulatedAnnealingOptimizer(
		CLASSIFICATION_SCRIPT,
		*analysis.NewDataset(TESTSET),
		50,
		0.95,
		10,
		manager.Results())
	o.Run()

	// cleanup actions
	for _, d := range datasets {
		os.Remove(d.Path())
	}
	if o.Result().Dataset.Id() == "" {
		t.Log("Result is not very good")
		t.FailNow()
	}
}
