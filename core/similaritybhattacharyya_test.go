package core

import (
	"os"
	"testing"
)

func TestBhattacharyyaCompute(t *testing.T) {
	datasets := createPoolBasedDatasets(20000, 20, 4)
	est := NewDatasetSimilarityEstimator(BHATTACHARYYA, datasets)
	err := est.Compute()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	s := est.GetSimilarities()
	if s == nil {
		t.Log("Nil similarities returned")
		t.FailNow()
	}
	for _, d1 := range datasets {
		for _, d2 := range datasets {
			//t.Log(s.Get(d1.Path(), d2.Path()))
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
