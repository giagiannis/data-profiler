package core

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestManagerAnalyze(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	rows, cols, d := 1000, 3, 20
	datasets := createPoolBasedDatasets(rows, d, cols)
	m := NewManager(datasets, 8, ANALYSIS_SCRIPT)
	m.Analyze()

	for _, f := range datasets {
		if _, ok := m.results[f.Id()]; !ok {
			t.Log(f.Id(),
				" not analyzed:",
				m.results)
			t.Fail()
		} else if len(m.results[f.Id()]) != cols*cols {
			t.Log("Serialized results missing from ",
				f.Id(),
				": ",
				m.results[f.Id()])
			t.Fail()
		}
	}

	for _, f := range datasets {
		os.Remove(f.Path())
	}
}

func TestManagerOptimizationResultsPruning(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	partitioner := NewDatasetPartitioner(TRAINSET, TRAINSET+"-splits", 100, UNIFORM)
	partitioner.Partition()
	datasets := DiscoverDatasets(TRAINSET + "-splits")

	manager := NewManager(datasets, 8, ANALYSIS_SCRIPT)
	manager.Analyze()

	manager.OptimizationResultsPruning()

	partitioner.Delete()

}
