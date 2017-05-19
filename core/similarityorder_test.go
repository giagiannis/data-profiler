package core

import (
	"os"
	"testing"
)

func TestOrderCompute(t *testing.T) {
	datasets := createPoolBasedDatasets(100, 20, 2)
	est := NewDatasetSimilarityEstimator(SIMILARITY_TYPE_ORDER, datasets)
	err := est.Compute()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	s := est.SimilarityMatrix()
	for i := range datasets {
		for j := range datasets {
			if s.Get(i, j) != s.Get(j, i) {
				t.Log("Similarity matrix not symmetrical")
				t.Fail()
			}
		}
	}

	for _, f := range datasets {
		os.Remove(f.Path())
	}
}

func TestOrderComputeAppxCnt(t *testing.T) {
	datasets := createPoolBasedDatasets(1000, 20, 4)
	est := NewDatasetSimilarityEstimator(SIMILARITY_TYPE_ORDER, datasets)
	pol := DatasetSimilarityPopulationPolicy{
		PolicyType: POPULATION_POL_APRX,
		Parameters: map[string]float64{
			"count": 20,
		},
	}
	est.SetPopulationPolicy(pol)
	err := est.Compute()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	s := est.SimilarityMatrix()
	for i := range datasets {
		for j := range datasets {
			if s.Get(i, j) != s.Get(j, i) {
				t.Log("Similarity matrix not symmetrical")
				t.Fail()
			}
		}
	}

	for _, f := range datasets {
		os.Remove(f.Path())
	}
}

func TestOrderComputeAppxThres(t *testing.T) {
	datasets := createPoolBasedDatasets(1000, 20, 4)
	est := NewDatasetSimilarityEstimator(SIMILARITY_TYPE_ORDER, datasets)
	pol := DatasetSimilarityPopulationPolicy{
		PolicyType: POPULATION_POL_APRX,
		Parameters: map[string]float64{
			"threshold": 0.95,
		},
	}
	est.SetPopulationPolicy(pol)
	err := est.Compute()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	s := est.SimilarityMatrix()
	for i := range datasets {
		for j := range datasets {
			if s.Get(i, j) != s.Get(j, i) {
				t.Log("Similarity matrix not symmetrical")
				t.Fail()
			}
		}
	}

	for _, f := range datasets {
		os.Remove(f.Path())
	}
}

func TestOrderSerialization(t *testing.T) {
	datasets := createPoolBasedDatasets(1000, 20, 4)
	est := *new(OrderEstimator)
	est.datasets = datasets
	pol := DatasetSimilarityPopulationPolicy{
		PolicyType: POPULATION_POL_APRX,
		Parameters: map[string]float64{
			"count": 5.0,
		},
	}
	est.SetPopulationPolicy(pol)
	err := est.Compute()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	bytes := est.Serialize()

	newEst := *new(OrderEstimator)
	newEst.Deserialize(bytes)
	if est.concurrency != newEst.concurrency {
		t.Log("Concurrency differs")
		t.Fail()
	}

	for i := range est.datasets {
		if est.datasets[i].Path() != newEst.datasets[i].Path() {
			t.Log("Dataset names are different", est.datasets[i], newEst.datasets[i])
			t.Fail()
		}
	}

	for i := 0; i < est.similarities.Capacity(); i++ {
		for j := 0; j < est.similarities.Capacity(); j++ {
			if est.similarities.Get(i, j) != newEst.similarities.Get(i, j) {
				t.Log("SM differs", i, j)
				t.Fail()
			}
		}
	}
}
