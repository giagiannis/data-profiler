package core

import (
	"os"
	"testing"
)

const MDSCALING_SCRIPT = "../r_scripts/mdscaling/mdscaling.R"

func TestMDScalingScript(t *testing.T) {
	datasets := createPoolBasedDatasets(2000, 10, 4)
	est := NewDatasetSimilarityEstimator(SIMILARITY_TYPE_JACOBBI, datasets)
	est.Configure(map[string]string{
		"concurrency": "10",
	})
	err := est.Compute()
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	md := NewMDScaling(est.GetSimilarities(), 2, MDSCALING_SCRIPT+"-bad")
	err = md.Compute()
	if err == nil {
		t.Log(err)
		t.Fail()
	}
	if md.Coordinates() != nil {
		t.Log("Script should not have succeeded")
		t.FailNow()
	}

	for _, d := range datasets {
		os.Remove(d.Path())
	}
}

func TestMDScalingCompute(t *testing.T) {
	datasets := createPoolBasedDatasets(100, 50, 2)
	est := NewDatasetSimilarityEstimator(SIMILARITY_TYPE_JACOBBI, datasets)
	est.Configure(map[string]string{
		"concurrency": "10",
	})
	err := est.Compute()
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	for _, k := range []int{1, 2, 3, 4, 5, 6} {
		md := NewMDScaling(est.GetSimilarities(), k, MDSCALING_SCRIPT)
		err = md.Compute()
		if err != nil {
			t.Log(err)
			t.Fail()
		}
		if md.Coordinates() == nil {
			t.Log("Coordinates are nil")
			t.FailNow()
		}
		_, err := md.Variances()

		if err != nil {
			t.Log(err)
			t.Fail()
		}
	}

}
