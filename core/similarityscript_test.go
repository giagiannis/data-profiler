package core

import (
	"os"
	"testing"
)

func TestScriptSimilarityDatasetAnalysis(t *testing.T) {
	datasets := createPoolBasedDatasets(1000, 20, 4)
	est := new(ScriptSimilarityEstimator)
	est.datasets = datasets
	conf := map[string]string{
		"concurrency": "10",
		"script":      ANALYSIS_SCRIPT,
		"norm":        "1",
	}
	est.Configure(conf)
	results := est.analyzeDatasets()
	if len(results) != len(datasets) || len(results) == 0 {
		t.Log("Not all datasets analyzed")
		t.FailNow()
	}

	size := len(results[0])
	for _, val := range results {
		if len(val) != size {
			t.Log("Incorrect number of elements")
			t.FailNow()
		}
	}

	for _, f := range datasets {
		os.Remove(f.Path())
	}
}
func TestScriptSimilarityCompute(t *testing.T) {
	datasets := createPoolBasedDatasets(1000, 50, 4)
	est := NewDatasetSimilarityEstimator(SIMILARITY_TYPE_SCRIPT, datasets)
	conf := map[string]string{
		"concurrency": "10",
		"script":      ANALYSIS_SCRIPT,
		"norm":        "1",
	}
	est.Configure(conf)
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

func TestScriptSimilarityComputeAppxThres(t *testing.T) {
	datasets := createPoolBasedDatasets(1000, 50, 4)
	est := NewDatasetSimilarityEstimator(SIMILARITY_TYPE_SCRIPT, datasets)
	conf := map[string]string{
		"concurrency": "10",
		"script":      ANALYSIS_SCRIPT,
		"norm":        "1",
	}
	est.Configure(conf)
	pol := DatasetSimilarityPopulationPolicy{
		PolicyType: POPULATION_POL_APRX,
		Parameters: map[string]float64{
			"threshold": 0.95,
		},
	}
	est.PopulationPolicy(pol)
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

func TestScriptSimilarityComputeAppxCnt(t *testing.T) {
	datasets := createPoolBasedDatasets(1000, 50, 4)
	est := NewDatasetSimilarityEstimator(SIMILARITY_TYPE_SCRIPT, datasets)
	conf := map[string]string{
		"concurrency": "10",
		"script":      ANALYSIS_SCRIPT,
		"norm":        "1",
	}
	est.Configure(conf)
	pol := DatasetSimilarityPopulationPolicy{
		PolicyType: POPULATION_POL_APRX,
		Parameters: map[string]float64{
			"count": 10,
		},
	}
	est.PopulationPolicy(pol)
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
