package core

import (
	"os"
	"testing"
)

func TestJacobbiEstimatorCompute(t *testing.T) {
	datasets := createPoolBasedDatasets(1000, 20, 4)
	est := NewDatasetSimilarityEstimator("jacobbi", datasets)
	err := est.Compute()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	for _, d1 := range datasets {
		for _, d2 := range datasets {
			if est.GetSimilarity(d1, d2) != est.GetSimilarity(d2, d1) {
				t.Log("Similarity matrix not symmetrical")
				t.Fail()
			}
		}
	}

	for _, f := range datasets {
		os.Remove(f.Path())
	}
}
