package core

import (
	"os"
	"testing"
)

func TestJacobbiCompute(t *testing.T) {
	datasets := createPoolBasedDatasets(1000, 20, 4)
	newDatasets := make([]*Dataset, len(datasets))
	for i := 0; i < len(datasets); i++ {
		newDatasets[i] = &datasets[i]
	}
	est := NewDatasetSimilarityEstimator(JACOBBI, newDatasets)
	err := est.Compute()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	s := est.GetSimilarities()
	for _, d1 := range datasets {
		for _, d2 := range datasets {
			if s.Get(d1, d2) != s.Get(d2, d1) {
				t.Log("Similarity matrix not symmetrical")
				t.Fail()
			}
		}
	}

	for _, f := range datasets {
		os.Remove(f.Path())
	}
}
