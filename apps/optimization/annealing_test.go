package optimization

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/giagiannis/data-profiler/analysis"
)

func TestRun(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	partitioner := analysis.NewDatasetPartitioner(TESTSET, TESTSET+"-splits", 100, analysis.UNIFORM)
	partitioner.Partition()
	datasets := analysis.DiscoverDatasets(TESTSET + "-splits")
	manager := analysis.NewManager(datasets, 8, ANALYSIS_SCRIPT)
	manager.Analyze()
	o := NewSimulatedAnnealingOptimizer(
		ML_SCRIPT,
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
