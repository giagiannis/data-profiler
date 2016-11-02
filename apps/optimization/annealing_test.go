package optimization

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/giagiannis/data-profiler/core"
)

func TestAnnealingRun(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	partitioner := core.NewDatasetPartitioner(TRAINSET, TRAINSET+"-splits", 100, core.UNIFORM)
	partitioner.Partition()
	datasets := core.DiscoverDatasets(TRAINSET + "-splits")
	manager := core.NewManager(datasets, 8, ANALYSIS_SCRIPT)
	manager.Analyze()
	o := NewSimulatedAnnealingOptimizer(
		ML_SCRIPT,
		*core.NewDataset(TESTSET),
		50,
		0.95,
		10,
		manager.Results(),
		core.EUCLIDEAN)
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
