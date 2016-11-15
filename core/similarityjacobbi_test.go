package core

import (
	"os"
	"testing"
)

func TestJacobbiCompute(t *testing.T) {
	datasets := createPoolBasedDatasets(1000, 20, 4)
	est := NewDatasetSimilarityEstimator(JACOBBI, datasets)
	err := est.Compute()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	s := est.GetSimilarities()
	for _, d1 := range datasets {
		for _, d2 := range datasets {
			if s.Get(d1.Path(), d2.Path()) != s.Get(d2.Path(), d1.Path()) {
				t.Log("Similarity matrix not symmetrical")
				t.Fail()
			}
		}
	}

	for _, f := range datasets {
		os.Remove(f.Path())
	}
}
